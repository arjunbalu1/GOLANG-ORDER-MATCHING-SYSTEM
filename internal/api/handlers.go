package api

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"order-matching-system/internal/database"
	"order-matching-system/internal/models"
	"order-matching-system/internal/service"
)

type Handler struct {
	orderRepo      *database.OrderRepository
	tradeRepo      *database.TradeRepository
	orderBookRepo  *database.OrderBookRepository
	matchingEngine *service.MatchingEngine
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		orderRepo:      database.NewOrderRepository(db),
		tradeRepo:      database.NewTradeRepository(db),
		orderBookRepo:  database.NewOrderBookRepository(db),
		matchingEngine: service.NewMatchingEngine(db),
	}
}

func (h *Handler) PlaceOrder(c *gin.Context) {
	var req models.PlaceOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate price for limit orders
	if req.Type == models.OrderTypeLimit && req.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "price must be greater than 0 for limit orders"})
		return
	}

	// Validate quantity
	if req.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "quantity must be greater than 0"})
		return
	}

	// Create order
	order := &models.Order{
		Symbol:          req.Symbol,
		Side:            req.Side,
		Type:            req.Type,
		Price:           req.Price,
		InitialQuantity: req.Quantity,
	}

	// Process order through matching engine
	if err := h.matchingEngine.ProcessOrder(order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *Handler) CancelOrder(c *gin.Context) {
	orderIDStr := c.Param("orderId")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	// Cancel order
	if err := h.orderRepo.CancelOrder(orderID); err != nil {
		if err.Error() == "order not found or already filled/canceled" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order canceled successfully"})
}

func (h *Handler) GetOrderBook(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol parameter is required"})
		return
	}

	orderBook, err := h.orderBookRepo.GetOrderBook(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orderBook)
}


func (h *Handler) ListTrades(c *gin.Context) {
	symbol := c.Query("symbol")

	var trades []*models.Trade
	var err error

	if symbol != "" {
		trades, err = h.tradeRepo.GetTradesBySymbol(symbol)
	} else {
		trades, err = h.tradeRepo.GetAllTrades()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trades)
}

func (h *Handler) GetOrderStatus(c *gin.Context) {
	orderIDStr := c.Param("orderId")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	order, err := h.orderRepo.GetOrderByID(orderID)
	if err != nil {
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}
