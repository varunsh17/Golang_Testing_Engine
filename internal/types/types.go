package types

import "time"

// Side represents order side
type Side string

const (
	SideBuy  Side = "BUY"
	SideSell Side = "SELL"
)

// OrderBookEntry represents a single order book entry
type OrderBookEntry struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
}

// OrderBookSnapshot represents a complete L2 order book snapshot
type OrderBookSnapshot struct {
	Symbol    string           `json:"symbol"`
	Timestamp time.Time        `json:"timestamp"`
	Bids      []OrderBookEntry `json:"bids"`
	Asks      []OrderBookEntry `json:"asks"`
}

// TradeSignal represents a trading signal from strategy to broker
type TradeSignal struct {
	Symbol    string
	Side      Side
	Price     float64
	Quantity  float64
	Timestamp time.Time
}

// Execution represents a completed trade
type Execution struct {
	Symbol    string
	Side      Side
	Price     float64
	Quantity  float64
	Timestamp time.Time
}

// Position represents a current position
type Position struct {
	Symbol        string
	Quantity      float64
	EntryPrice    float64
	EntryTime     time.Time
	CurrentPrice  float64
	UnrealizedPnL float64
}
