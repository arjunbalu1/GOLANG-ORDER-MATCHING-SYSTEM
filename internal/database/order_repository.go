package database

import (
	"database/sql"
	"fmt"

	"order-matching-system/internal/models"
)

type OrderRepository struct {
	db DBTX
}

func NewOrderRepository(db DBTX) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) CreateOrder(order *models.Order) error {
	query := `
		INSERT INTO orders (symbol, side, type, price, initial_quantity, remaining_quantity, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(
		query,
		order.Symbol,
		order.Side,
		order.Type,
		order.Price,
		order.InitialQuantity,
		order.RemainingQuantity,
		order.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	order.ID = int(id)
	return nil
}

func (r *OrderRepository) GetOrderByID(id int) (*models.Order, error) {
	query := `
		SELECT id, symbol, side, type, price, initial_quantity, remaining_quantity, status, created_at, updated_at
		FROM orders
		WHERE id = ?
	`

	order := &models.Order{}
	var price sql.NullFloat64

	err := r.db.QueryRow(query, id).Scan(
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
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if price.Valid {
		order.Price = price.Float64
	}

	return order, nil
}

func (r *OrderRepository) GetOpenOrdersBySymbol(symbol string) ([]*models.Order, error) {
	query := `
		SELECT id, symbol, side, type, price, initial_quantity, remaining_quantity, status, created_at, updated_at
		FROM orders
		WHERE symbol = ? AND status = 'open'
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get open orders: %w", err)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		order := &models.Order{}
		var price sql.NullFloat64

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

func (r *OrderRepository) UpdateOrderStatus(id int, status models.OrderStatus, remainingQuantity float64) error {
	query := `
		UPDATE orders
		SET status = ?, remaining_quantity = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query, status, remainingQuantity, id)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

func (r *OrderRepository) CancelOrder(id int) error {
	query := `
		UPDATE orders
		SET status = 'canceled'
		WHERE id = ? AND status = 'open'
	`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found or already filled/canceled")
	}

	return nil
}
