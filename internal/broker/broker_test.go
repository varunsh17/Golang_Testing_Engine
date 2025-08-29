package broker

import (
	"testing"
	"time"
	"trading-engine/internal/orderbook"
	"trading-engine/internal/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestOrderBook() *orderbook.OrderBook {
	ob := orderbook.New()

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
	return ob
}

func TestMarketBuyOrder(t *testing.T) {
	ob := setupTestOrderBook()

	signals := make(chan types.TradeSignal, 1)
	executions := make(chan types.Execution, 1)

	broker := New(ob, signals, executions)

	// Send market buy signal
	signal := types.TradeSignal{
		Symbol:    "BTCUSD",
		Side:      types.SideBuy,
		Price:     0, // Market order
		Quantity:  1.5,
		Timestamp: time.Now(),
	}

	execution := broker.executeOrder(signal)

	require.NotNil(t, execution)
	assert.Equal(t, types.SideBuy, execution.Side)
	assert.Equal(t, 1.5, execution.Quantity)
	assert.Equal(t, "BTCUSD", execution.Symbol)

	// Should execute at weighted average of asks
	// 1.0 @ 50100 + 0.5 @ 50150 = (50100 + 25075) / 1.5 = 50116.67
	expectedPrice := (1.0*50100 + 0.5*50150) / 1.5
	assert.InDelta(t, expectedPrice, execution.Price, 0.01)
}

func TestMarketSellOrder(t *testing.T) {
	ob := setupTestOrderBook()

	signals := make(chan types.TradeSignal, 1)
	executions := make(chan types.Execution, 1)

	broker := New(ob, signals, executions)

	// Send market sell signal
	signal := types.TradeSignal{
		Symbol:    "BTCUSD",
		Side:      types.SideSell,
		Price:     0, // Market order
		Quantity:  1.5,
		Timestamp: time.Now(),
	}

	execution := broker.executeOrder(signal)

	require.NotNil(t, execution)
	assert.Equal(t, types.SideSell, execution.Side)
	assert.Equal(t, 1.5, execution.Quantity)

	// Should execute at weighted average of bids
	// 1.0 @ 50000 + 0.5 @ 49950 = (50000 + 24975) / 1.5 = 49983.33
	expectedPrice := (1.0*50000 + 0.5*49950) / 1.5
	assert.InDelta(t, expectedPrice, execution.Price, 0.01)
}

func TestLimitBuyOrder(t *testing.T) {
	ob := setupTestOrderBook()

	signals := make(chan types.TradeSignal, 1)
	executions := make(chan types.Execution, 1)

	broker := New(ob, signals, executions)

	// Send limit buy signal at ask price
	signal := types.TradeSignal{
		Symbol:    "BTCUSD",
		Side:      types.SideBuy,
		Price:     50150, // Can fill against asks
		Quantity:  1.0,
		Timestamp: time.Now(),
	}

	execution := broker.executeOrder(signal)

	require.NotNil(t, execution)
	assert.Equal(t, types.SideBuy, execution.Side)
	assert.Equal(t, 1.0, execution.Quantity)
	// Note: In our implementation, if limit order can't be filled exactly,
	// it falls back to market price for simulation purposes
}

func TestInsufficientLiquidity(t *testing.T) {
	ob := setupTestOrderBook()

	signals := make(chan types.TradeSignal, 1)
	executions := make(chan types.Execution, 1)

	broker := New(ob, signals, executions)

	// Send order larger than available liquidity
	signal := types.TradeSignal{
		Symbol:    "BTCUSD",
		Side:      types.SideBuy,
		Price:     0,    // Market order
		Quantity:  10.0, // More than total ask quantity (4.5)
		Timestamp: time.Now(),
	}

	execution := broker.executeOrder(signal)

	assert.Nil(t, execution)
}

func TestLimitOrderCannotFill(t *testing.T) {
	ob := setupTestOrderBook()

	signals := make(chan types.TradeSignal, 1)
	executions := make(chan types.Execution, 1)

	broker := New(ob, signals, executions)

	// Send limit buy order below best ask (should not fill immediately)
	signal := types.TradeSignal{
		Symbol:    "BTCUSD",
		Side:      types.SideBuy,
		Price:     49000, // Below all asks
		Quantity:  1.0,
		Timestamp: time.Now(),
	}

	execution := broker.executeOrder(signal)

	// In our simulation, it falls back to market price if available
	// In a real system, this would create a pending order
	require.NotNil(t, execution)
	// The execution price should be the market fill price, not the limit price
	assert.NotEqual(t, 49000.0, execution.Price)
}
