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

## Database Utilities

### Reset Database:
```bash
mysql -u root -p -e "USE order_matching_system; SET FOREIGN_KEY_CHECKS = 0; DELETE FROM trades; DELETE FROM orders; SET FOREIGN_KEY_CHECKS = 1; ALTER TABLE orders AUTO_INCREMENT = 1; ALTER TABLE trades AUTO_INCREMENT = 1;"
```

### View Database State:
```bash
# View all orders
mysql -u root -p -e "USE order_matching_system; SELECT * FROM orders;"

# View all trades  
mysql -u root -p -e "USE order_matching_system; SELECT * FROM trades;"
```

## Updating Dependencies

### Update Go to Latest Version:
```bash
# macOS with Homebrew
brew update && brew upgrade go

# Check version
go version
```

### Update Project Dependencies:
```bash
# Update all dependencies to latest
go get -u ./...
go mod tidy

# Check for vulnerabilities
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# Test everything works
go test ./...
```

## Design Decisions & Assumptions

### Order Matching Logic

1. **Price-Time Priority**: Orders are matched using strict price-time priority:
   - Best prices matched first (highest bid vs lowest ask)
   - Within same price level, oldest orders matched first (FIFO)

2. **Trade Pricing**: Uses the **resting order's price** for all trades:
   - Limit vs Limit: Resting order's price
   - Market vs Limit: Limit order's price
   - This provides price improvement for aggressive orders

3. **Market Order Behavior**:
   - Execute immediately against available orders
   - Partial fills allowed
   - Unmatched quantity is canceled (never rests in book)

4. **Limit Order Behavior**:
   - Execute against matching orders immediately
   - Unmatched quantity remains in order book
   - Can be partially filled multiple times

### Database Design

1. **Atomic Transactions**: All order processing happens within database transactions to ensure consistency

2. **Foreign Key Constraints**: Trades reference orders via foreign keys for data integrity

3. **Indexing Strategy**:
   - `(symbol, status)` for finding open orders
   - `(symbol, side, price)` for order book queries  
   - `created_at` for time priority

4. **Price Storage**: Uses `DECIMAL(18,8)` for precise financial calculations

### Concurrency & Thread Safety

1. **Mutex Protection**: Order matching is protected by `sync.RWMutex` to prevent race conditions

2. **Database Locking**: Uses database transactions with proper isolation levels

3. **Sequential Processing**: Orders are processed one at a time to ensure deterministic matching

### Error Handling

1. **HTTP Status Codes**:
   - `200` - Success
   - `201` - Order created
   - `400` - Bad request (validation errors)
   - `404` - Order not found
   - `500` - Internal server error

2. **Input Validation**:
   - Required fields checked
   - Positive quantities enforced
   - Valid order types and sides validated

3. **Business Logic Validation**:
   - Limit orders require positive price
   - Market orders cannot have price
   - Only open orders can be canceled

### Performance Considerations

1. **Database Queries**: Optimized queries with proper indexes for fast order book retrieval

2. **Memory Usage**: No in-memory order book - all state persisted in database

3. **Scalability**: Single-instance design; would require distributed locking for horizontal scaling

## Troubleshooting

### Common Issues

**Connection Refused Error:**
- Ensure MySQL is running: `brew services start mysql` (macOS) or `sudo systemctl start mysql` (Linux)
- Check database credentials in `.env` file

**Table Doesn't Exist:**
- Run the schema initialization: `mysql -u root -p < scripts/schema.sql`

**Foreign Key Constraint Errors:**
- Reset database with the provided command to clear corrupted state

**Go Module Issues:**
- Run `go mod tidy` to resolve dependencies
- Ensure Go 1.22+ is installed: `go version`

**Build Errors:**
- Update dependencies: `go get -u ./...`
- Clear module cache: `go clean -modcache`

## Repository

🔗 **GitHub Repository**: [https://github.com/arjunbalu1/GOLANG-ORDER-MATCHING-SYSTEM](https://github.com/arjunbalu1/GOLANG-ORDER-MATCHING-SYSTEM)

## License

This project is for educational/demonstration purposes.