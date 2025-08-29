package orderbook

import (
	"fmt"
	"sort"
	"sync"
	"time"
	"trading-engine/internal/types"
)

// OrderBook represents an L2 order book
type OrderBook struct {
	mu          sync.RWMutex
	symbol      string
	bids        []types.OrderBookEntry // sorted descending by price
	asks        []types.OrderBookEntry // sorted ascending by price
	lastUpdated time.Time
}

// New creates a new order book
func New() *OrderBook {
	return &OrderBook{
		bids: make([]types.OrderBookEntry, 0),
		asks: make([]types.OrderBookEntry, 0),
	}
}

// Update updates the order book with new snapshot
func (ob *OrderBook) Update(snapshot types.OrderBookSnapshot) {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	ob.symbol = snapshot.Symbol
	ob.lastUpdated = snapshot.Timestamp

	// Copy and sort bids (descending by price)
	ob.bids = make([]types.OrderBookEntry, len(snapshot.Bids))
	copy(ob.bids, snapshot.Bids)
	sort.Slice(ob.bids, func(i, j int) bool {
		return ob.bids[i].Price > ob.bids[j].Price
	})

	// Copy and sort asks (ascending by price)
	ob.asks = make([]types.OrderBookEntry, len(snapshot.Asks))
	copy(ob.asks, snapshot.Asks)
	sort.Slice(ob.asks, func(i, j int) bool {
		return ob.asks[i].Price < ob.asks[j].Price
	})
}

// GetBestBid returns the highest bid price and quantity
func (ob *OrderBook) GetBestBid() (float64, float64, bool) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	if len(ob.bids) == 0 {
		return 0, 0, false
	}
	return ob.bids[0].Price, ob.bids[0].Quantity, true
}

// GetBestAsk returns the lowest ask price and quantity
func (ob *OrderBook) GetBestAsk() (float64, float64, bool) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	if len(ob.asks) == 0 {
		return 0, 0, false
	}
	return ob.asks[0].Price, ob.asks[0].Quantity, true
}

// GetSpread returns the bid-ask spread
func (ob *OrderBook) GetSpread() (float64, bool) {
	bidPrice, _, bidExists := ob.GetBestBid()
	askPrice, _, askExists := ob.GetBestAsk()

	if !bidExists || !askExists {
		return 0, false
	}

	return askPrice - bidPrice, true
}

// GetMidPrice returns the mid price
func (ob *OrderBook) GetMidPrice() (float64, bool) {
	bidPrice, _, bidExists := ob.GetBestBid()
	askPrice, _, askExists := ob.GetBestAsk()

	if !bidExists || !askExists {
		return 0, false
	}

	return (bidPrice + askPrice) / 2, true
}

// GetCumulativeDepth returns cumulative quantity up to a price level
func (ob *OrderBook) GetCumulativeDepth(side types.Side, priceLevel float64) float64 {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	var depth float64

	if side == types.SideBuy {
		// For bids, sum all quantities at prices >= priceLevel
		for _, bid := range ob.bids {
			if bid.Price >= priceLevel {
				depth += bid.Quantity
			}
		}
	} else {
		// For asks, sum all quantities at prices <= priceLevel
		for _, ask := range ob.asks {
			if ask.Price <= priceLevel {
				depth += ask.Quantity
			}
		}
	}

	return depth
}

// GetLiquidity returns total liquidity within a price range
func (ob *OrderBook) GetLiquidity(fromMid, percentage float64) (float64, float64) {
	midPrice, exists := ob.GetMidPrice()
	if !exists {
		return 0, 0
	}

	ob.mu.RLock()
	defer ob.mu.RUnlock()

	var bidLiquidity, askLiquidity float64

	// Calculate bid liquidity within percentage range
	minBidPrice := midPrice * (1 - percentage)
	for _, bid := range ob.bids {
		if bid.Price >= minBidPrice {
			bidLiquidity += bid.Quantity
		}
	}

	// Calculate ask liquidity within percentage range
	maxAskPrice := midPrice * (1 + percentage)
	for _, ask := range ob.asks {
		if ask.Price <= maxAskPrice {
			askLiquidity += ask.Quantity
		}
	}

	return bidLiquidity, askLiquidity
}

// GetOrderBookImbalance calculates the order book imbalance
func (ob *OrderBook) GetOrderBookImbalance() float64 {
	bidLiq, askLiq := ob.GetLiquidity(0, 0.01) // 1% from mid

	totalLiq := bidLiq + askLiq
	if totalLiq == 0 {
		return 0
	}

	// Returns positive for bid-heavy, negative for ask-heavy
	return (bidLiq - askLiq) / totalLiq
}

// CanFill checks if an order can be filled at the given price and quantity
func (ob *OrderBook) CanFill(side types.Side, price, quantity float64) bool {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	var availableQty float64

	if side == types.SideBuy {
		// Buying against asks
		for _, ask := range ob.asks {
			if ask.Price <= price {
				availableQty += ask.Quantity
				if availableQty >= quantity {
					return true
				}
			}
		}
	} else {
		// Selling against bids
		for _, bid := range ob.bids {
			if bid.Price >= price {
				availableQty += bid.Quantity
				if availableQty >= quantity {
					return true
				}
			}
		}
	}

	return false
}

// GetFillPrice calculates the average fill price for a market order
func (ob *OrderBook) GetFillPrice(side types.Side, quantity float64) (float64, bool) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	var remainingQty = quantity
	var totalCost float64

	if side == types.SideBuy {
		// Buying against asks
		for _, ask := range ob.asks {
			if remainingQty <= 0 {
				break
			}

			fillQty := ask.Quantity
			if fillQty > remainingQty {
				fillQty = remainingQty
			}

			totalCost += fillQty * ask.Price
			remainingQty -= fillQty
		}
	} else {
		// Selling against bids
		for _, bid := range ob.bids {
			if remainingQty <= 0 {
				break
			}

			fillQty := bid.Quantity
			if fillQty > remainingQty {
				fillQty = remainingQty
			}

			totalCost += fillQty * bid.Price
			remainingQty -= fillQty
		}
	}

	if remainingQty > 0 {
		// Could not fill completely
		return 0, false
	}

	return totalCost / quantity, true
}

// String returns a string representation of the order book
func (ob *OrderBook) String() string {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	result := fmt.Sprintf("OrderBook[%s] - %v\n", ob.symbol, ob.lastUpdated.Format("15:04:05"))
	result += "ASKS:\n"

	// Show top 5 asks (reversed for display)
	for i := len(ob.asks) - 1; i >= 0 && i >= len(ob.asks)-5; i-- {
		result += fmt.Sprintf("  %.2f @ %.2f\n", ob.asks[i].Quantity, ob.asks[i].Price)
	}

	if spread, exists := ob.GetSpread(); exists {
		result += fmt.Sprintf("--- SPREAD: %.2f ---\n", spread)
	}

	result += "BIDS:\n"
	// Show top 5 bids
	for i := 0; i < len(ob.bids) && i < 5; i++ {
		result += fmt.Sprintf("  %.2f @ %.2f\n", ob.bids[i].Quantity, ob.bids[i].Price)
	}

	return result
}
