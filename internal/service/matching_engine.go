package service

import (
	"database/sql"
	"fmt"
	"sync"

	"order-matching-system/internal/database"
	"order-matching-system/internal/models"
)

type MatchingEngine struct {
	db          *sql.DB
	orderRepo   *database.OrderRepository
	tradeRepo   *database.TradeRepository
	orderBookMu sync.RWMutex // Protects concurrent access to order book
}

func NewMatchingEngine(db *sql.DB) *MatchingEngine {
	return &MatchingEngine{
		db:        db,
		orderRepo: database.NewOrderRepository(db),
		tradeRepo: database.NewTradeRepository(db),
	}
}

func (me *MatchingEngine) ProcessOrder(order *models.Order) error {
	me.orderBookMu.Lock()
	defer me.orderBookMu.Unlock()

	tx, err := me.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create repositories with transaction
	orderRepo := database.NewOrderRepository(tx)
	tradeRepo := database.NewTradeRepository(tx)

	// Save the order to database first
	order.Status = models.OrderStatusOpen
	order.RemainingQuantity = order.InitialQuantity
	if err := orderRepo.CreateOrder(order); err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	// Get matching orders from the opposite side
	matchingOrders, err := me.getMatchingOrders(tx, order.Symbol, order.Side)
	if err != nil {
		return fmt.Errorf("failed to get matching orders: %w", err)
	}

	// Process matches
	for _, matchOrder := range matchingOrders {
		if order.RemainingQuantity <= 0 {
			break
		}

		// Check if orders can match
		if !me.canMatch(order, matchOrder) {
			continue	//next iteration
		}

		// Determine trade price (use resting order's price)
		tradePrice := me.determineTradePrice(order, matchOrder)

		// Calculate trade quantity
		tradeQuantity := min(order.RemainingQuantity, matchOrder.RemainingQuantity)

		// Create trade record
		trade := &models.Trade{
			Symbol:   order.Symbol,
			Price:    tradePrice,
			Quantity: tradeQuantity,
		}

		// Set buy and sell order IDs
		if order.Side == models.OrderSideBuy {
			trade.BuyOrderID = order.ID
			trade.SellOrderID = matchOrder.ID
		} else {
			trade.BuyOrderID = matchOrder.ID
			trade.SellOrderID = order.ID
		}

		// Save trade
		if err := tradeRepo.CreateTrade(trade); err != nil {
			return fmt.Errorf("failed to create trade: %w", err)
		}

		// Update order quantities
		order.RemainingQuantity -= tradeQuantity
		matchOrder.RemainingQuantity -= tradeQuantity

		// Update matched order status
		matchStatus := models.OrderStatusOpen
		if matchOrder.RemainingQuantity == 0 {
			matchStatus = models.OrderStatusFilled
		}
		if err := orderRepo.UpdateOrderStatus(matchOrder.ID, matchStatus, matchOrder.RemainingQuantity); err != nil {
			return fmt.Errorf("failed to update matched order: %w", err)
		}
	}

	// Update the incoming order status
	finalStatus := models.OrderStatusOpen
	if order.RemainingQuantity == 0 {
		finalStatus = models.OrderStatusFilled
	} else if order.Type == models.OrderTypeMarket && order.RemainingQuantity < order.InitialQuantity {
		// Market order partially filled, cancel remaining
		finalStatus = models.OrderStatusCanceled
		order.RemainingQuantity = 0
	} else if order.Type == models.OrderTypeMarket {
		// Market order with no matches, cancel it
		finalStatus = models.OrderStatusCanceled
		order.RemainingQuantity = 0
	}

	if err := orderRepo.UpdateOrderStatus(order.ID, finalStatus, order.RemainingQuantity); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	order.Status = finalStatus
	return nil
}

// getMatchingOrders retrieves orders that can potentially match
func (me *MatchingEngine) getMatchingOrders(tx *sql.Tx, symbol string, side models.OrderSide) ([]*models.Order, error) {
	var query string

	if side == models.OrderSideBuy {
		// For buy orders, get sell orders sorted by price ASC, then by time
		query = `
			SELECT id, symbol, side, type, price, initial_quantity, remaining_quantity, status, created_at, updated_at
			FROM orders
			WHERE symbol = ? AND side = 'sell' AND status = 'open'
			ORDER BY 
				CASE WHEN type = 'market' THEN 0 ELSE price END ASC,
				created_at ASC
		`
	} else {
		// For sell orders, get buy orders sorted by price DESC, then by time
		query = `
			SELECT id, symbol, side, type, price, initial_quantity, remaining_quantity, status, created_at, updated_at
			FROM orders
			WHERE symbol = ? AND side = 'buy' AND status = 'open'
			ORDER BY 
				CASE WHEN type = 'market' THEN 999999999 ELSE price END DESC,
				created_at ASC
		`
	}

	rows, err := tx.Query(query, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to query matching orders: %w", err)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		order := &models.Order{}
		var price sql.NullFloat64	//market 

		err := rows.Scan(
			&order.ID,
			&order.Symbol,
			&order.Side,
			&order.Type,
			&price,
			&order.InitialQuantity,
			&order.RemainingQuantity,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		if price.Valid {
			order.Price = price.Float64
		}

		orders = append(orders, order)
	}

	return orders, nil
}

// canMatch determines if two orders can match
func (me *MatchingEngine) canMatch(incoming, resting *models.Order) bool {
	// Market orders always match
	if incoming.Type == models.OrderTypeMarket || resting.Type == models.OrderTypeMarket {	//resting.Type == models.OrderTypeMarket wont happen
		return true
	}

	// For limit orders, check price compatibility
	if incoming.Side == models.OrderSideBuy {
		// Buy order matches if its price >= sell order price
		return incoming.Price >= resting.Price
	} else {
		// Sell order matches if its price <= buy order price
		return incoming.Price <= resting.Price
	}
}

// determineTradePrice determines the execution price for a trade
func (me *MatchingEngine) determineTradePrice(incoming, resting *models.Order) float64 {
	// Always use the resting (existing) order's price for limit/limit matches
	// For market/limit matches, use the limit order's price
	if resting.Type == models.OrderTypeLimit {
		return resting.Price
	}
	// Both market orders shouldn't happen, but if it does, use incoming price
	return incoming.Price
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
