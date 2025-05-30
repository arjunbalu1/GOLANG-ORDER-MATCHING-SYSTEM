package api

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

func SetupRouter(db *sql.DB) *gin.Engine {
	router := gin.Default()

	handler := NewHandler(db)

	router.POST("/orders", handler.PlaceOrder)
	router.DELETE("/orders/:orderId", handler.CancelOrder)
	router.GET("/orders/:orderId", handler.GetOrderStatus)
	router.GET("/orderbook", handler.GetOrderBook)
	router.GET("/trades", handler.ListTrades)

	return router
}
