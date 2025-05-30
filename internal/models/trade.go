package models

import (
	"time"
)

type Trade struct {
	ID          int       `json:"id"`
	Symbol      string    `json:"symbol"`
	BuyOrderID  int       `json:"buy_order_id"`
	SellOrderID int       `json:"sell_order_id"`
	Price       float64   `json:"price"`
	Quantity    float64   `json:"quantity"`
	CreatedAt   time.Time `json:"created_at"`
}
