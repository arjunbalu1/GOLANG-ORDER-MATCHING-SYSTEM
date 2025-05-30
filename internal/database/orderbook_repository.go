package database

import (
	"fmt"

	"order-matching-system/internal/models"
)

type OrderBookRepository struct {
	db DBTX
}

func NewOrderBookRepository(db DBTX) *OrderBookRepository {
	return &OrderBookRepository{db: db}
}

func (r *OrderBookRepository) GetOrderBook(symbol string) (*models.OrderBook, error) {
	orderBook := &models.OrderBook{
		Symbol: symbol,
		Bids:   []models.OrderBookEntry{},
		Asks:   []models.OrderBookEntry{},
	}

	bidsQuery := `
		SELECT price, SUM(remaining_quantity) as total_quantity, COUNT(*) as order_count
		FROM orders
		WHERE symbol = ? AND side = 'buy' AND status = 'open' AND type = 'limit'
		GROUP BY price
		ORDER BY price DESC
		LIMIT 50
	`

	bidRows, err := r.db.Query(bidsQuery, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get bids: %w", err)
	}
	defer bidRows.Close()

	for bidRows.Next() {
		var entry models.OrderBookEntry
		err := bidRows.Scan(&entry.Price, &entry.Quantity, &entry.Orders)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bid: %w", err)
		}
		orderBook.Bids = append(orderBook.Bids, entry)
	}

	// Get sell orders (asks) - sorted by price ASC
	asksQuery := `
		SELECT price, SUM(remaining_quantity) as total_quantity, COUNT(*) as order_count
		FROM orders
		WHERE symbol = ? AND side = 'sell' AND status = 'open' AND type = 'limit'
		GROUP BY price
		ORDER BY price ASC
		LIMIT 50
	`

	askRows, err := r.db.Query(asksQuery, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get asks: %w", err)
	}
	defer askRows.Close()

	for askRows.Next() {
		var entry models.OrderBookEntry
		err := askRows.Scan(&entry.Price, &entry.Quantity, &entry.Orders)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ask: %w", err)
		}
		orderBook.Asks = append(orderBook.Asks, entry)
	}

	return orderBook, nil
}
