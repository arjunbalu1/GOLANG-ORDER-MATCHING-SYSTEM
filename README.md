# Order Matching System

## Architecture

```
├── cmd/server/          # Application entry point
├── internal/
│   ├── api/            # HTTP handlers and routing (Gin)
│   ├── database/       # Data access layer (raw SQL)
│   ├── models/         # Data structures and types
│   └── service/        # Business logic (matching engine)
├── scripts/            # Database schema
└── .env               # Configuration
```

## Installation & Setup

### 1. Install Go Dependencies

```bash
# Download dependencies
go mod download

# Update to latest versions (optional)
go get -u ./...
go mod tidy
```

### 2. MySQL Setup

#### Set MySQL Root Password:
```bash
mysql -u root -p
ALTER USER 'root'@'localhost' IDENTIFIED BY 'password';
FLUSH PRIVILEGES;
EXIT;
```

*Note: This sets the root password to match your .env configuration.*

### 3. Environment Configuration

Create a `.env` file in the project root:

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=order_matching_system

# Server Configuration
SERVER_PORT=8080
```

### 4. Database Initialization

Initialize the database schema:

```bash
mysql -u root -p < scripts/schema.sql
```

This creates:
- `order_matching_system` database
- `orders` table (with proper indexes)
- `trades` table (with foreign key constraints)

## Running the Application

### Start the Server

```bash
go run cmd/server/main.go
```

Expected output:
```
2025/05/30 15:36:02 Connecting to database...
2025/05/30 15:36:02 Database connected successfully
[GIN-debug] POST   /orders                   --> order-matching-system/internal/api.(*Handler).PlaceOrder-fm (3 handlers)
[GIN-debug] DELETE /orders/:orderId          --> order-matching-system/internal/api.(*Handler).CancelOrder-fm (3 handlers)
[GIN-debug] GET    /orders/:orderId          --> order-matching-system/internal/api.(*Handler).GetOrderStatus-fm (3 handlers)
[GIN-debug] GET    /orderbook                --> order-matching-system/internal/api.(*Handler).GetOrderBook-fm (3 handlers)
[GIN-debug] GET    /trades                   --> order-matching-system/internal/api.(*Handler).ListTrades-fm (3 handlers)
2025/05/30 15:36:02 Starting server on port 8080...
```

### Build for Production

```bash
# Build binary
go build -o order-matching-server cmd/server/main.go

# Run binary
./order-matching-server
```

## API Documentation

### Base URL
```
http://localhost:8080
```

### 1. Place Order

**Endpoint:** `POST /orders`

#### Limit Order Examples:

**Buy Limit Order:**
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "AAPL",
    "side": "buy",
    "type": "limit",
    "price": 150.50,
    "quantity": 100
  }'
```

**Sell Limit Order:**
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "AAPL",
    "side": "sell", 
    "type": "limit",
    "price": 149.75,
    "quantity": 50
  }'
```

#### Market Order Examples:

**Market Buy Order:**
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "AAPL",
    "side": "buy",
    "type": "market",
    "quantity": 25
  }'
```

**Market Sell Order:**
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "AAPL",
    "side": "sell",
    "type": "market",
    "quantity": 75
  }'
```

#### Response:
```json
{
  "id": 1,
  "symbol": "AAPL",
  "side": "buy",
  "type": "limit",
  "price": 150.50,
  "initial_quantity": 100,
  "remaining_quantity": 100,
  "status": "open",
  "created_at": "2025-05-30T15:36:07Z",
  "updated_at": "2025-05-30T15:36:07Z"
}
```

### 2. Get Order Status

**Endpoint:** `GET /orders/{orderId}`

```bash
curl -X GET http://localhost:8080/orders/1
```

### 3. Cancel Order

**Endpoint:** `DELETE /orders/{orderId}`

```bash
curl -X DELETE http://localhost:8080/orders/1
```

### 4. Get Order Book

**Endpoint:** `GET /orderbook?symbol={symbol}`

```bash
curl -X GET "http://localhost:8080/orderbook?symbol=AAPL"
```

#### Response:
```json
{
  "symbol": "AAPL",
  "bids": [
    {"price": 150.00, "quantity": 100, "orders": 2},
    {"price": 149.50, "quantity": 50, "orders": 1}
  ],
  "asks": [
    {"price": 151.00, "quantity": 75, "orders": 1},
    {"price": 151.50, "quantity": 200, "orders": 3}
  ]
}
```

### 5. List Trades

**Endpoint:** `GET /trades?symbol={symbol}` (optional symbol filter)

```bash
# All trades
curl -X GET http://localhost:8080/trades

# Trades for specific symbol
curl -X GET "http://localhost:8080/trades?symbol=AAPL"
```

#### Response:
```json
[
  {
    "id": 1,
    "symbol": "AAPL",
    "buy_order_id": 1,
    "sell_order_id": 2,
    "price": 150.00,
    "quantity": 50,
    "created_at": "2025-05-30T15:39:25Z"
  }
]
```

## Testing Scenarios

### Complete Matching Example:

```bash
# 1. Reset database
mysql -u root -p -e "USE order_matching_system; SET FOREIGN_KEY_CHECKS = 0; DELETE FROM trades; DELETE FROM orders; SET FOREIGN_KEY_CHECKS = 1; ALTER TABLE orders AUTO_INCREMENT = 1; ALTER TABLE trades AUTO_INCREMENT = 1;"

# 2. Place a sell limit order
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"symbol": "AAPL", "side": "sell", "type": "limit", "price": 150.00, "quantity": 100}'

# 3. Place a buy limit order that matches
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"symbol": "AAPL", "side": "buy", "type": "limit", "price": 150.50, "quantity": 50}'

# 4. Check the order book (should show remaining sell order)
curl -X GET "http://localhost:8080/orderbook?symbol=AAPL"

# 5. Check trades (should show the executed trade)
curl -X GET "http://localhost:8080/trades?symbol=AAPL"
```

### Partial Market Order Example:

```bash
# Place a buy limit order
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"symbol": "TSLA", "side": "buy", "type": "limit", "price": 100.00, "quantity": 50}'

# Place a market sell order for more than available
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"symbol": "TSLA", "side": "sell", "type": "market", "quantity": 100}'

# Check results - market order should be partially filled then canceled
curl -X GET "http://localhost:8080/trades?symbol=TSLA"
```
