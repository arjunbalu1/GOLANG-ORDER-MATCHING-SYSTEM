package database

import (
	"fmt"

	"order-matching-system/internal/models"
)

type TradeRepository struct {
	db DBTX
}

func NewTradeRepository(db DBTX) *TradeRepository {
	return &TradeRepository{db: db}
}

func (r *TradeRepository) CreateTrade(trade *models.Trade) error {
	query := `
		INSERT INTO trades (symbol, buy_order_id, sell_order_id, price, quantity)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(
		query,
		trade.Symbol,
		trade.BuyOrderID,
		trade.SellOrderID,
		trade.Price,
		trade.Quantity,
	)
	if err != nil {
		return fmt.Errorf("failed to create trade: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	trade.ID = int(id)
	return nil
}

func (r *TradeRepository) GetTradesBySymbol(symbol string) ([]*models.Trade, error) {
	query := `
		SELECT id, symbol, buy_order_id, sell_order_id, price, quantity, created_at
		FROM trades
		WHERE symbol = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get trades: %w", err)
	}
	defer rows.Close()

	var trades []*models.Trade
	for rows.Next() {
		trade := &models.Trade{}
		err := rows.Scan(
			&trade.ID,
			&trade.Symbol,
			&trade.BuyOrderID,
			&trade.SellOrderID,
			&trade.Price,
			&trade.Quantity,
			&trade.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trade: %w", err)
		}
		trades = append(trades, trade)
	}

	return trades, nil
}

func (r *TradeRepository) GetAllTrades() ([]*models.Trade, error) {
	query := `
		SELECT id, symbol, buy_order_id, sell_order_id, price, quantity, created_at
		FROM trades
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all trades: %w", err)
	}
	defer rows.Close()

	var trades []*models.Trade
	for rows.Next() {
		trade := &models.Trade{}
		err := rows.Scan(
			&trade.ID,
			&trade.Symbol,
			&trade.BuyOrderID,
			&trade.SellOrderID,
			&trade.Price,
			&trade.Quantity,
			&trade.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trade: %w", err)
		}
		trades = append(trades, trade)
	}

	return trades, nil
}
