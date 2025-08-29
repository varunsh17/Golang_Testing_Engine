package orderbook

import (
	"testing"
	"time"
	"trading-engine/internal/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOrderBook(t *testing.T) {
	ob := New()
	assert.NotNil(t, ob)
	assert.Empty(t, ob.bids)
	assert.Empty(t, ob.asks)
}

func TestOrderBookUpdate(t *testing.T) {
	ob := New()

	snapshot := types.OrderBookSnapshot{
		Symbol:    "BTCUSD",
		Timestamp: time.Now(),
		Bids: []types.OrderBookEntry{
			{Price: 50000, Quantity: 1.0},
			{Price: 49950, Quantity: 2.0},
			{Price: 49900, Quantity: 1.5},
		},
		Asks: []types.OrderBookEntry{
			{Price: 50100, Quantity: 1.0},
			{Price: 50150, Quantity: 2.0},
			{Price: 50200, Quantity: 1.5},
		},
	}

	ob.Update(snapshot)

	assert.Equal(t, "BTCUSD", ob.symbol)
	assert.Len(t, ob.bids, 3)
	assert.Len(t, ob.asks, 3)

	// Check bid sorting (descending)
	assert.Equal(t, 50000.0, ob.bids[0].Price)
	assert.Equal(t, 49950.0, ob.bids[1].Price)
	assert.Equal(t, 49900.0, ob.bids[2].Price)

	// Check ask sorting (ascending)
	assert.Equal(t, 50100.0, ob.asks[0].Price)
	assert.Equal(t, 50150.0, ob.asks[1].Price)
	assert.Equal(t, 50200.0, ob.asks[2].Price)
}

func TestGetBestBidAsk(t *testing.T) {
	ob := New()

	// Test empty order book
	_, _, exists := ob.GetBestBid()
	assert.False(t, exists)

	_, _, exists = ob.GetBestAsk()
	assert.False(t, exists)

	// Add data
	snapshot := types.OrderBookSnapshot{
		Symbol:    "BTCUSD",
		Timestamp: time.Now(),
		Bids: []types.OrderBookEntry{
			{Price: 50000, Quantity: 1.0},
			{Price: 49950, Quantity: 2.0},
		},
		Asks: []types.OrderBookEntry{
			{Price: 50100, Quantity: 1.0},
			{Price: 50150, Quantity: 2.0},
		},
	}

	ob.Update(snapshot)

	// Test best bid
	bidPrice, bidQty, exists := ob.GetBestBid()
	assert.True(t, exists)
	assert.Equal(t, 50000.0, bidPrice)
	assert.Equal(t, 1.0, bidQty)

	// Test best ask
	askPrice, askQty, exists := ob.GetBestAsk()
	assert.True(t, exists)
	assert.Equal(t, 50100.0, askPrice)
	assert.Equal(t, 1.0, askQty)
}

func TestGetSpread(t *testing.T) {
	ob := New()

	snapshot := types.OrderBookSnapshot{
		Symbol:    "BTCUSD",
		Timestamp: time.Now(),
		Bids: []types.OrderBookEntry{
			{Price: 50000, Quantity: 1.0},
		},
		Asks: []types.OrderBookEntry{
			{Price: 50100, Quantity: 1.0},
		},
	}

	ob.Update(snapshot)

	spread, exists := ob.GetSpread()
	assert.True(t, exists)
	assert.Equal(t, 100.0, spread)
}

func TestGetMidPrice(t *testing.T) {
	ob := New()

	snapshot := types.OrderBookSnapshot{
		Symbol:    "BTCUSD",
		Timestamp: time.Now(),
		Bids: []types.OrderBookEntry{
			{Price: 50000, Quantity: 1.0},
		},
		Asks: []types.OrderBookEntry{
			{Price: 50100, Quantity: 1.0},
		},
	}

	ob.Update(snapshot)

	midPrice, exists := ob.GetMidPrice()
	assert.True(t, exists)
	assert.Equal(t, 50050.0, midPrice)
}

func TestGetCumulativeDepth(t *testing.T) {
	ob := New()

	snapshot := types.OrderBookSnapshot{
		Symbol:    "BTCUSD",
		Timestamp: time.Now(),
		Bids: []types.OrderBookEntry{
			{Price: 50000, Quantity: 1.0},
			{Price: 49950, Quantity: 2.0},
			{Price: 49900, Quantity: 1.5},
		},
		Asks: []types.OrderBookEntry{
			{Price: 50100, Quantity: 1.0},
			{Price: 50150, Quantity: 2.0},
			{Price: 50200, Quantity: 1.5},
		},
	}

	ob.Update(snapshot)

	// Test bid depth
	bidDepth := ob.GetCumulativeDepth(types.SideBuy, 49950)
	assert.Equal(t, 3.0, bidDepth) // 1.0 + 2.0 from levels >= 49950

	// Test ask depth
	askDepth := ob.GetCumulativeDepth(types.SideSell, 50150)
	assert.Equal(t, 3.0, askDepth) // 1.0 + 2.0 from levels <= 50150
}

func TestCanFill(t *testing.T) {
	ob := New()

	snapshot := types.OrderBookSnapshot{
		Symbol:    "BTCUSD",
		Timestamp: time.Now(),
		Bids: []types.OrderBookEntry{
			{Price: 50000, Quantity: 1.0},
			{Price: 49950, Quantity: 2.0},
		},
		Asks: []types.OrderBookEntry{
			{Price: 50100, Quantity: 1.0},
			{Price: 50150, Quantity: 2.0},
		},
	}

	ob.Update(snapshot)

	// Test buy order (against asks)
	canFill := ob.CanFill(types.SideBuy, 50150, 2.5)
	assert.True(t, canFill) // Can fill 1.0 + 2.0 = 3.0 >= 2.5

	canFill = ob.CanFill(types.SideBuy, 50150, 4.0)
	assert.False(t, canFill) // Only 3.0 available

	// Test sell order (against bids)
	canFill = ob.CanFill(types.SideSell, 49950, 2.5)
	assert.True(t, canFill) // Can fill 1.0 + 2.0 = 3.0 >= 2.5

	canFill = ob.CanFill(types.SideSell, 49950, 4.0)
	assert.False(t, canFill) // Only 3.0 available
}

func TestGetFillPrice(t *testing.T) {
	ob := New()

	snapshot := types.OrderBookSnapshot{
		Symbol:    "BTCUSD",
		Timestamp: time.Now(),
		Bids: []types.OrderBookEntry{
			{Price: 50000, Quantity: 1.0},
			{Price: 49950, Quantity: 2.0},
		},
		Asks: []types.OrderBookEntry{
			{Price: 50100, Quantity: 1.0},
			{Price: 50150, Quantity: 2.0},
		},
	}

	ob.Update(snapshot)

	// Test buy order (market buy against asks)
	fillPrice, canFill := ob.GetFillPrice(types.SideBuy, 2.0)
	require.True(t, canFill)
	expectedPrice := (1.0*50100 + 1.0*50150) / 2.0 // 50125
	assert.Equal(t, expectedPrice, fillPrice)

	// Test sell order (market sell against bids)
	fillPrice, canFill = ob.GetFillPrice(types.SideSell, 2.0)
	require.True(t, canFill)
	expectedPrice = (1.0*50000 + 1.0*49950) / 2.0 // 49975
	assert.Equal(t, expectedPrice, fillPrice)

	// Test insufficient liquidity
	_, canFill = ob.GetFillPrice(types.SideBuy, 5.0)
	assert.False(t, canFill)
}

func TestGetLiquidity(t *testing.T) {
	ob := New()

	snapshot := types.OrderBookSnapshot{
		Symbol:    "BTCUSD",
		Timestamp: time.Now(),
		Bids: []types.OrderBookEntry{
			{Price: 50000, Quantity: 1.0}, // within 1% from mid (50050)
			{Price: 49500, Quantity: 2.0}, // outside 1% from mid
		},
		Asks: []types.OrderBookEntry{
			{Price: 50100, Quantity: 1.0}, // within 1% from mid (50050)
			{Price: 50600, Quantity: 2.0}, // outside 1% from mid
		},
	}

	ob.Update(snapshot)

	// Mid price = (50000 + 50100) / 2 = 50050
	// 1% range: 49549.5 to 50550.5
	// Bids within range: 50000 (1.0)
	// Asks within range: 50100 (1.0)
	bidLiq, askLiq := ob.GetLiquidity(0, 0.01) // 1%
	assert.Equal(t, 1.0, bidLiq)               // Only the 50000 bid is within 1%
	assert.Equal(t, 1.0, askLiq)               // Only the 50100 ask is within 1%
}

func TestGetOrderBookImbalance(t *testing.T) {
	ob := New()

	snapshot := types.OrderBookSnapshot{
		Symbol:    "BTCUSD",
		Timestamp: time.Now(),
		Bids: []types.OrderBookEntry{
			{Price: 50000, Quantity: 3.0},
		},
		Asks: []types.OrderBookEntry{
			{Price: 50100, Quantity: 1.0},
		},
	}

	ob.Update(snapshot)

	imbalance := ob.GetOrderBookImbalance()
	// (3.0 - 1.0) / (3.0 + 1.0) = 0.5
	assert.Equal(t, 0.5, imbalance)
}
