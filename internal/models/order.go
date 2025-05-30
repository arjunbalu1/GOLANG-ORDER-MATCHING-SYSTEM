package models

import (
	"time"
)

type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

type OrderType string

const (
	OrderTypeLimit  OrderType = "limit"
	OrderTypeMarket OrderType = "market"
)

type OrderStatus string

const (
	OrderStatusOpen     OrderStatus = "open"
	OrderStatusFilled   OrderStatus = "filled"
	OrderStatusCanceled OrderStatus = "canceled"
)

type Order struct {
	ID                int         `json:"id"`
	Symbol            string      `json:"symbol"`
	Side              OrderSide   `json:"side"`
	Type              OrderType   `json:"type"`
	Price             float64     `json:"price,omitempty"` // Only for limit orders
	InitialQuantity   float64     `json:"initial_quantity"`
	RemainingQuantity float64     `json:"remaining_quantity"`
	Status            OrderStatus `json:"status"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
}

type PlaceOrderRequest struct {
	Symbol   string    `json:"symbol" binding:"required"`
	Side     OrderSide `json:"side" binding:"required,oneof=buy sell"`
	Type     OrderType `json:"type" binding:"required,oneof=limit market"`
	Price    float64   `json:"price" binding:"omitempty,min=0"`
	Quantity float64   `json:"quantity" binding:"required,min=0"`
}

type OrderBookEntry struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
	Orders   int     `json:"orders"` // Number of orders at this price level
}

type OrderBook struct {
	Symbol string           `json:"symbol"`
	Bids   []OrderBookEntry `json:"bids"` // Buy orders (sorted highest to lowest)
	Asks   []OrderBookEntry `json:"asks"` // Sell orders (sorted lowest to highest)
}
