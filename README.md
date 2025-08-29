# Go Trading Engine

A minimalist Go trading engine that simulates order execution against Level-2 order book data. The engine reads order book snapshots from JSON files and executes trades based on configurable strategy parameters.

## Features

- **L2 Order Book**: Full Level-2 order book implementation with best bid/ask and cumulative depth queries
- **Multi-Component Architecture**: Separate goroutines for feed, engine, strategy, and broker components
- **Comprehensive Strategy**: Combines liquidity-based entry, profit targets, stop-loss, order book imbalance, and time-based exits
- **Deterministic Simulation**: File-based order book simulation ensures reproducible results
- **Trade Logging**: Exports detailed trade logs to CSV format
- **Performance Analytics**: Real-time P&L calculation and summary statistics

## Architecture

The system consists of four main components communicating via channels:

1. **Feed**: Reads order book snapshots from JSON files and publishes updates
2. **Engine**: Processes order book updates and maintains current market state
3. **Strategy**: Analyzes market conditions and generates trade signals
4. **Broker**: Executes trade signals against the order book and reports fills

```
Feed -> Engine -> Strategy -> Broker
 |        |         |         |
 v        v         v         v
JSON -> OrderBook -> Signals -> Executions
```

## Installation

```bash
git clone <repository>
cd trading-engine
go mod tidy
```

## Usage

### Basic Usage

```bash
# Run with default parameters using sample1.json
go run main.go

# Specify custom orderbook file
go run main.go -orderbook data/sample2.json

# Set specific entry price
go run main.go -entry 50000 -size 1.5

# Configure risk parameters
go run main.go -stop 0.03 -profit 0.08 -hold 45s
```

### CLI Parameters

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-orderbook` | string | `data/sample1.json` | Path to orderbook JSON file |
| `-entry` | float64 | `0` | Entry price (0 for auto/market) |
| `-size` | float64 | `100` | Order size |
| `-stop` | float64 | `0.02` | Stop loss percentage (0.02 = 2%) |
| `-profit` | float64 | `0.05` | Take profit percentage (0.05 = 5%) |
| `-liquidity` | float64 | `1000` | Minimum liquidity threshold |
| `-hold` | duration | `30s` | Maximum hold time |
| `-output` | string | `trades.csv` | Output CSV file for trades |

### Example Commands

```bash
# Conservative strategy with tight stops
go run main.go -stop 0.01 -profit 0.02 -hold 15s -size 50

# Aggressive strategy with wide targets
go run main.go -stop 0.05 -profit 0.15 -hold 2m -size 200

# High-frequency strategy with auto-entry
go run main.go -entry 0 -hold 5s -liquidity 500 -output hf_trades.csv

# Test with different assets
go run main.go -orderbook data/sample2.json -entry 3000 -size 5
go run main.go -orderbook data/sample3.json -entry 0.45 -size 10000
```

## Order Book Data Format

The engine expects JSON files containing order book snapshots:

```json
[
  {
    "symbol": "BTCUSD",
    "timestamp": "2025-08-30T10:00:00Z",
    "bids": [
      {"price": 50000.00, "quantity": 1.5},
      {"price": 49950.00, "quantity": 2.0}
    ],
    "asks": [
      {"price": 50100.00, "quantity": 1.2},
      {"price": 50150.00, "quantity": 1.8}
    ]
  }
]
```

### Sample Data Files

- `data/sample1.json`: Bitcoin (BTCUSD) order book with ~$100 spread
- `data/sample2.json`: Ethereum (ETHUSD) order book with tighter spread
- `data/sample3.json`: Cardano (ADAUSD) order book with high liquidity

## Strategy Logic

The strategy implements a multi-factor approach:

1. **Liquidity-Based Entry**: Only enters when liquidity exceeds threshold
2. **Auto-Entry**: Market orders when entry price is 0
3. **Stop-Loss**: Exits at percentage loss from entry
4. **Take-Profit**: Exits at percentage gain from entry
5. **Order Book Imbalance**: Considers bid/ask ratio for timing
6. **Time-Based Exit**: Maximum holding period to limit exposure

## Output

### Terminal Output

```
Starting trading engine with:
  Orderbook file: data/sample1.json
  Entry price: 0.00 (0 = auto)
  Order size: 100.00
  Stop loss: 2.00%
  Take profit: 5.00%
  Liquidity threshold: 1000.00
  Max hold time: 30s
  Output file: trades.csv

Feed loaded 5 snapshots from data/sample1.json
Engine started
Strategy started
Broker started
Strategy received execution: {Symbol:BTCUSD Side:BUY Price:50116.67 Quantity:100 Timestamp:2025-08-30 ...}
Position opened: 100.00 @ 50116.67

=== TRADING SUMMARY ===
Total trades: 2
Total P&L: 150.25
Return: 0.30%
Trade log written to: trades.csv
```

### CSV Trade Log

The output CSV contains detailed trade records:

```csv
Timestamp,Side,Price,Quantity,Symbol
2025-08-30T10:00:02.123Z,BUY,50116.67,100.00,BTCUSD
2025-08-30T10:00:32.456Z,SELL,50267.92,100.00,BTCUSD
```

## Testing

Run the test suite:

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/orderbook
go test ./internal/broker

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...
```

### Test Coverage

- **Order Book Tests**: L2 book operations, depth calculations, liquidity analysis
- **Matching Engine Tests**: Market orders, limit orders, partial fills, insufficient liquidity

## API Reference

### OrderBook Methods

```go
// Core market data
func (ob *OrderBook) GetBestBid() (float64, float64, bool)
func (ob *OrderBook) GetBestAsk() (float64, float64, bool)
func (ob *OrderBook) GetSpread() (float64, bool)
func (ob *OrderBook) GetMidPrice() (float64, bool)

// Liquidity analysis
func (ob *OrderBook) GetCumulativeDepth(side Side, priceLevel float64) float64
func (ob *OrderBook) GetLiquidity(fromMid, percentage float64) (float64, float64)
func (ob *OrderBook) GetOrderBookImbalance() float64

// Order execution
func (ob *OrderBook) CanFill(side Side, price, quantity float64) bool
func (ob *OrderBook) GetFillPrice(side Side, quantity float64) (float64, bool)
```

## Performance Considerations

- **Memory**: Order book snapshots are loaded entirely into memory
- **Concurrency**: All components run in separate goroutines with channel communication
- **Determinism**: File-based simulation ensures reproducible results
- **Scalability**: Channel buffer sizes can be adjusted for high-frequency data

## Limitations

- **Simulation Only**: Does not connect to real exchanges
- **Single Asset**: Processes one symbol at a time
- **Simplified Matching**: Basic order matching without partial fills tracking
- **No Persistence**: Order book state is not persisted between runs

## Future Enhancements

- Real-time exchange connectivity (WebSocket feeds)
- Multi-asset portfolio management
- Advanced order types (iceberg, TWAP, etc.)
- Risk management and position sizing
- Machine learning signal generation
- Real-time performance dashboards

## Dependencies

- `github.com/stretchr/testify`: Testing framework

## License

MIT License - see LICENSE file for details


SAMPLE COMMMAND AND OUTPUT

```
go run main.go -concurrent
PS C:\Users\varun\Desktop\Code\Vs_code\trading_project> go run main.go -concurrent
🔥 GO TRADING ENGINE - Goroutines & Channels Demo
================================================
🚀 STARTING CONCURRENT TRADING SESSIONS
======================================
📋 Launching 3 concurrent trading sessions:
   1. BTC-Aggressive (File: data/sample1.json)
   2. ETH-Conservative (File: data/sample2.json)
   3. ADA-HighFreq (File: data/sample3.json)

📊 Real-time session updates:
   01:25:43 🟢 [BTC-Aggressive] Starting trading session
   01:25:43 🔧 [BTC-Aggressive] Initialized channels and orderbook
   01:25:43 ⚙️  [BTC-Aggressive] Starting 4 component goroutines
   01:25:43 🟢 [ETH-Conservative] Starting trading session
   01:25:43 🔧 [ETH-Conservative] Initialized channels and orderbook
2025/08/30 01:25:43 Strategy started
2025/08/30 01:25:43 Engine started
   01:25:43 ⚙️  [ETH-Conservative] Starting 4 component goroutines
2025/08/30 01:25:43 Strategy started
   01:25:43 🟢 [ADA-HighFreq] Starting trading session
2025/08/30 01:25:43 Strategy execution handler started
   01:25:43 🔧 [ADA-HighFreq] Initialized channels and orderbook
2025/08/30 01:25:43 Engine started
   01:25:43 ⚙️  [ADA-HighFreq] Starting 4 component goroutines
2025/08/30 01:25:43 Broker started
2025/08/30 01:25:43 Strategy started
2025/08/30 01:25:43 Engine started
2025/08/30 01:25:43 Broker started
2025/08/30 01:25:43 Broker started
2025/08/30 01:25:43 Feed loaded 5 snapshots from data/sample1.json
2025/08/30 01:25:43 Feed loaded 5 snapshots from data/sample3.json
2025/08/30 01:25:43 Feed loaded 5 snapshots from data/sample2.json
2025/08/30 01:25:43 Strategy execution handler started
2025/08/30 01:25:43 Strategy execution handler started
2025/08/30 01:25:43 Published snapshot 1: BTCUSD @ 01:25:43.895
2025/08/30 01:25:43 Published snapshot 1: ADAUSD @ 01:25:43.895
2025/08/30 01:25:43 Published snapshot 1: ETHUSD @ 01:25:43.895
2025/08/30 01:25:44 Published snapshot 2: ADAUSD @ 01:25:43.995
2025/08/30 01:25:44 Published snapshot 2: ETHUSD @ 01:25:43.995
2025/08/30 01:25:44 Published snapshot 2: BTCUSD @ 01:25:43.995
2025/08/30 01:25:44 Published snapshot 3: BTCUSD @ 01:25:44.095
2025/08/30 01:25:44 Published snapshot 3: ADAUSD @ 01:25:44.095
2025/08/30 01:25:44 Published snapshot 3: ETHUSD @ 01:25:44.095
2025/08/30 01:25:44 Published snapshot 4: ETHUSD @ 01:25:44.195
2025/08/30 01:25:44 Published snapshot 4: BTCUSD @ 01:25:44.195
2025/08/30 01:25:44 Published snapshot 4: ADAUSD @ 01:25:44.195
2025/08/30 01:25:44 Published snapshot 5: BTCUSD @ 01:25:44.295
2025/08/30 01:25:44 Published snapshot 5: ETHUSD @ 01:25:44.295
2025/08/30 01:25:44 Published snapshot 5: ADAUSD @ 01:25:44.295
2025/08/30 01:25:44 Generating market buy signal (auto-entry)
2025/08/30 01:25:44 Buy signal sent
2025/08/30 01:25:44 Generating limit buy signal at 3000.00
2025/08/30 01:25:44 Limit buy signal sent
2025/08/30 01:25:44 Generating market buy signal (auto-entry)
2025/08/30 01:25:44 Broker received signal: {Symbol:BTCUSD Side:BUY Price:0 Quantity:
2.5 Timestamp:2025-08-30 01:25:44.3978823 +0530 IST m=+0.514300201}                  2025/08/30 01:25:44 Broker received signal: {Symbol:BTCUSD Side:BUY Price:3000 Quanti
ty:5 Timestamp:2025-08-30 01:25:44.3989209 +0530 IST m=+0.515338801}                 2025/08/30 01:25:44 Buy signal sent
2025/08/30 01:25:44 Broker received signal: {Symbol:BTCUSD Side:BUY Price:0 Quantity:
8000 Timestamp:2025-08-30 01:25:44.40058 +0530 IST m=+0.516997901}                   2025/08/30 01:25:44 Market order cannot be filled: insufficient liquidity
2025/08/30 01:25:44 Order executed: BUY 2.50 @ 50150.00
2025/08/30 01:25:44 Limit order cannot be filled at 3000.00
2025/08/30 01:25:44 Executing at best available price: 3008.80
2025/08/30 01:25:44 Order executed: BUY 5.00 @ 3008.80
   01:25:44 💱 [BTC-Aggressive] Trade #1: BUY 2.50 @ $50150.00
2025/08/30 01:25:44 Strategy received execution: {Symbol:BTCUSD Side:BUY Price:50150 
Quantity:2.5 Timestamp:2025-08-30 01:25:44.4012191 +0530 IST m=+0.517637001}            01:25:44 💱 [ETH-Conservative] Trade #1: BUY 5.00 @ $3008.80
2025/08/30 01:25:44 Execution sent: {Symbol:BTCUSD Side:BUY Price:50150 Quantity:2.5 
Timestamp:2025-08-30 01:25:44.4012191 +0530 IST m=+0.517637001}                      2025/08/30 01:25:44 Execution sent: {Symbol:BTCUSD Side:BUY Price:3008.8 Quantity:5 T
imestamp:2025-08-30 01:25:44.4062552 +0530 IST m=+0.522673101}                       2025/08/30 01:25:44 Strategy received execution: {Symbol:BTCUSD Side:BUY Price:3008.8
 Quantity:5 Timestamp:2025-08-30 01:25:44.4062552 +0530 IST m=+0.522673101}          2025/08/30 01:25:44 Position opened: 2.50 @ 50150.00
2025/08/30 01:25:44 Scheduling exit signals...
2025/08/30 01:25:44 Position opened: 5.00 @ 3008.80
2025/08/30 01:25:44 Scheduling exit signals...
2025/08/30 01:25:44 Feed completed
2025/08/30 01:25:44 Engine finished processing 5 total updates
   01:25:44 📡 [ETH-Conservative] Data feed completed
2025/08/30 01:25:44 Feed completed
2025/08/30 01:25:44 Engine finished processing 5 total updates
2025/08/30 01:25:44 Feed completed
   01:25:44 📡 [BTC-Aggressive] Data feed completed
2025/08/30 01:25:44 Engine finished processing 5 total updates
   01:25:44 📡 [ADA-HighFreq] Data feed completed
2025/08/30 01:25:46 Generating take-profit exit signal at 3084.02
2025/08/30 01:25:46 Take-profit signal sent
2025/08/30 01:25:46 Generating take-profit exit signal at 52156.00
2025/08/30 01:25:46 Broker received signal: {Symbol:BTCUSD Side:SELL Price:0 Quantity
:5 Timestamp:2025-08-30 01:25:46.4196312 +0530 IST m=+2.536049101}                   2025/08/30 01:25:46 Order executed: SELL 5.00 @ 3003.50
2025/08/30 01:25:46 Take-profit signal sent
   01:25:46 💱 [ETH-Conservative] Trade #2: SELL 5.00 @ $3003.50
2025/08/30 01:25:46 Broker received signal: {Symbol:BTCUSD Side:SELL Price:0 Quantity
:2.5 Timestamp:2025-08-30 01:25:46.4208886 +0530 IST m=+2.537306501}                 2025/08/30 01:25:46 Order executed: SELL 2.50 @ 50018.00
2025/08/30 01:25:46 Execution sent: {Symbol:BTCUSD Side:SELL Price:50018 Quantity:2.5
 Timestamp:2025-08-30 01:25:46.4225119 +0530 IST m=+2.538929801}                        01:25:46 💱 [BTC-Aggressive] Trade #2: SELL 2.50 @ $50018.00
2025/08/30 01:25:46 Execution sent: {Symbol:BTCUSD Side:SELL Price:3003.5 Quantity:5 
Timestamp:2025-08-30 01:25:46.4212066 +0530 IST m=+2.537624501}                      2025/08/30 01:25:46 Strategy received execution: {Symbol:BTCUSD Side:SELL Price:3003.
5 Quantity:5 Timestamp:2025-08-30 01:25:46.4212066 +0530 IST m=+2.537624501}         2025/08/30 01:25:46 Strategy received execution: {Symbol:BTCUSD Side:SELL Price:50018
 Quantity:2.5 Timestamp:2025-08-30 01:25:46.4225119 +0530 IST m=+2.538929801}        2025/08/30 01:25:46 Position closed: 2.50 @ 50018.00
2025/08/30 01:25:46 Trade PnL: -330.00 (held for 2.0212928s)
2025/08/30 01:25:46 Position closed: 5.00 @ 3003.50
2025/08/30 01:25:46 Trade PnL: -26.50 (held for 2.0149514s)
2025/08/30 01:25:52 Broker finished
2025/08/30 01:25:52 Strategy execution handler finished
   01:25:52 🎯 [ADA-HighFreq] All executions completed
   01:25:52 ✅ [ADA-HighFreq] Completed in 8.5637968s - 0 trades
2025/08/30 01:25:54 Broker finished
2025/08/30 01:25:54 Strategy execution handler finished
   01:25:54 🎯 [BTC-Aggressive] All executions completed
   01:25:54 📝 [BTC-Aggressive] Trade log written to concurrent_btc_trades.csv       
   01:25:54 ✅ [BTC-Aggressive] Completed in 10.5763033s - 2 trades
2025/08/30 01:25:58 Broker finished
2025/08/30 01:25:58 Strategy execution handler finished
   01:25:58 🎯 [ETH-Conservative] All executions completed
   01:25:58 📝 [ETH-Conservative] Trade log written to concurrent_eth_trades.csv     
   01:25:58 ✅ [ETH-Conservative] Completed in 14.5681875s - 2 trades

======================================================================
🏁 CONCURRENT EXECUTION RESULTS
======================================================================

📈 ADA-HighFreq:
   📁 Data Source: data/sample3.json
   💹 Executed Trades: 0
   💰 Session P&L: $0.00
   ⏱️  Execution Time: 8.5637968s
   📊 Strategy: Entry=0, Size=8000.0, Stop=0.5%, Profit=1.5%
   📄 Trade Log: concurrent_ada_trades.csv

📈 BTC-Aggressive:
   📁 Data Source: data/sample1.json
   💹 Executed Trades: 2
   💰 Session P&L: $-330.00
   ⏱️  Execution Time: 10.5763033s
   📊 Strategy: Entry=0, Size=2.5, Stop=1.5%, Profit=4.0%
   📄 Trade Log: concurrent_btc_trades.csv

📈 ETH-Conservative:
   📁 Data Source: data/sample2.json
   💹 Executed Trades: 2
   💰 Session P&L: $-26.50
   ⏱️  Execution Time: 14.5681875s
   📊 Strategy: Entry=3000, Size=5.0, Stop=1.0%, Profit=2.5%
   📄 Trade Log: concurrent_eth_trades.csv

🎯 CONCURRENCY PERFORMANCE:
   ✅ Successful Sessions: 3/3
   📈 Total Trades Executed: 4
   💵 Combined Portfolio P&L: $-356.50
   🚄 Total Wall-Clock Time: 14.5688458s
   ⚡ Estimated Speedup: 2.1x faster than sequential
   🔧 Goroutines Used: 3 main sessions + internal goroutines per session
   📡 Channels Used: Results, Progress, + 4 channels per session

🧠 GOROUTINES & CHANNELS ARCHITECTURE:
   • Main goroutine: Orchestrates and collects results
   • Progress goroutine: Real-time status updates via channel
   • Session goroutines: One per trading session (3 total)
   • Per-session goroutines: Feed, Engine, Strategy, Broker (4 each)
   • Channel communication: orderbook updates, trade signals, executions
   • Total concurrent goroutines: ~17 running simultaneously!

cd "c:\Users\varun\Desktop\Code\Vs_code\trading_project\tools"; go run validate_orders.go
PS C:\Users\varun\Desktop\Code\Vs_code\trading_project> cd "c:\Users\varun\Desktop\Co
de\Vs_code\trading_project\tools"; go run validate_orders.go                         === ORDER BOOK MATHEMATICAL VALIDATION ===

1. Testing Basic Order Book Sorting...
   ✅ Best Bid: 105.00 @ 2.00
   ✅ Best Ask: 108.00 @ 2.00
   ✅ Spread: 3.00
   ✅ Mid Price: 106.50

2. Testing Market Order Execution Math...
   ✅ Market Buy 2.5 units: 99.70 (expected 99.70)
   ✅ Market Sell 2.5 units: 99.70 (expected 99.70)

3. Testing Limit Order Validation...
   ✅ Limit buy at 101.00: fillable
   ✅ Limit buy at 100.50: not fillable
   ✅ Limit sell at 100.00: fillable
   ✅ Limit sell at 101.50: not fillable

4. Testing Edge Cases...
   ✅ Empty order book handled correctly
   ✅ Insufficient liquidity detected correctly

=== ALL MATHEMATICAL VALIDATIONS PASSED ===

cd "c:\Users\varun\Desktop\Code\Vs_code\trading_project"; go run tools/validate_orders.go
top\Code\Vs_code\trading_project"; go run tools/validate_orders.go                   === ORDER BOOK MATHEMATICAL VALIDATION ===

1. Testing Basic Order Book Sorting...
   ✅ Best Bid: 105.00 @ 2.00
   ✅ Best Ask: 108.00 @ 2.00
   ✅ Spread: 3.00
   ✅ Mid Price: 106.50

2. Testing Market Order Execution Math...
   ✅ Market Buy 2.5 units: 99.70 (expected 99.70)
   ✅ Market Sell 2.5 units: 99.70 (expected 99.70)

3. Testing Limit Order Validation...
   ✅ Limit buy at 101.00: fillable
   ✅ Limit buy at 100.50: not fillable
   ✅ Limit sell at 100.00: fillable
   ✅ Limit sell at 101.50: not fillable

4. Testing Edge Cases...
   ✅ Empty order book handled correctly
   ✅ Insufficient liquidity detected correctly

=== ALL MATHEMATICAL VALIDATIONS PASSED ===

go run tools/comprehensive_validation.go
=== COMPREHENSIVE ORDER MATHEMATICS VALIDATION ===

1. Testing Sample1 Order Book Mathematics...
   ✅ Best Bid: 50000.00 @ 1.50
   ✅ Best Ask: 50100.00 @ 1.20
   ✅ Spread: 100.00
   ✅ Market Buy 1.0: 50100.00
   ✅ Market Buy 2.0: 50120.00 (expected 50120.00)

2. Testing Edge Cases Mathematics...
   ✅ Small quantities handled correctly
   ✅ Exact fills work correctly
   ✅ Overfill detection works

3. Testing Cumulative Depth Calculations...
   ✅ Bid depth at 99.0: 3.00
   ✅ Ask depth at 102.0: 3.00
   ✅ Liquidity within 1%: bid=1.00 ask=1.00

🎉 ALL MATHEMATICAL VALIDATIONS PASSED! 🎉

✅ Order book sorting is mathematically correct
✅ Market order execution follows FIFO price-time priority
✅ Limit order validation respects price constraints
✅ Cumulative depth calculations are accurate
✅ P&L calculations are precise
✅ Edge cases are handled properly


go run main.go -session=btc 
================================================

🎯 Running specific session: btc
2025/08/30 01:28:01 Engine started
2025/08/30 01:28:01 Strategy started
2025/08/30 01:28:01 Strategy execution handler started
2025/08/30 01:28:01 Broker started
2025/08/30 01:28:01 Feed loaded 5 snapshots from data/sample1.json
2025/08/30 01:28:01 Published snapshot 1: BTCUSD @ 01:28:01.402
2025/08/30 01:28:01 Published snapshot 2: BTCUSD @ 01:28:01.502
2025/08/30 01:28:01 Published snapshot 3: BTCUSD @ 01:28:01.602
2025/08/30 01:28:01 Published snapshot 4: BTCUSD @ 01:28:01.702
2025/08/30 01:28:01 Published snapshot 5: BTCUSD @ 01:28:01.802
2025/08/30 01:28:01 Generating market buy signal (auto-entry)
2025/08/30 01:28:01 Buy signal sent
2025/08/30 01:28:01 Broker received signal: {Symbol:BTCUSD Side:BUY Price:0 Quantity:
1.5 Timestamp:2025-08-30 01:28:01.9047813 +0530 IST m=+0.508338301}                  2025/08/30 01:28:01 Order executed: BUY 1.50 @ 50130.00
2025/08/30 01:28:01 Execution sent: {Symbol:BTCUSD Side:BUY Price:50130 Quantity:1.5 
Timestamp:2025-08-30 01:28:01.9059094 +0530 IST m=+0.509466401}                      2025/08/30 01:28:01 Strategy received execution: {Symbol:BTCUSD Side:BUY Price:50130 
Quantity:1.5 Timestamp:2025-08-30 01:28:01.9059094 +0530 IST m=+0.509466401}         2025/08/30 01:28:01 Position opened: 1.50 @ 50130.00
2025/08/30 01:28:01 Scheduling exit signals...
2025/08/30 01:28:01 Feed completed
2025/08/30 01:28:01 Engine finished processing 5 total updates
2025/08/30 01:28:03 Generating take-profit exit signal at 52636.50
2025/08/30 01:28:03 Take-profit signal sent
2025/08/30 01:28:03 Broker received signal: {Symbol:BTCUSD Side:SELL Price:0 Quantity
:1.5 Timestamp:2025-08-30 01:28:03.9198131 +0530 IST m=+2.523370101}                 2025/08/30 01:28:03 Order executed: SELL 1.50 @ 50030.00
2025/08/30 01:28:03 Execution sent: {Symbol:BTCUSD Side:SELL Price:50030 Quantity:1.5
 Timestamp:2025-08-30 01:28:03.921096 +0530 IST m=+2.524653001}                      2025/08/30 01:28:03 Strategy received execution: {Symbol:BTCUSD Side:SELL Price:50030
 Quantity:1.5 Timestamp:2025-08-30 01:28:03.921096 +0530 IST m=+2.524653001}         2025/08/30 01:28:03 Position closed: 1.50 @ 50030.00
2025/08/30 01:28:03 Trade PnL: -150.00 (held for 2.0151866s)
2025/08/30 01:28:09 Broker finished
2025/08/30 01:28:09 Strategy execution handler finished
✅ Session completed successfully!
   💹 Trades: 2
   💰 P&L: -150.00
   📄 Output: btc_test_trades.csv
```



