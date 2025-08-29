package main

import (
	"fmt"
	"log"
	"time"
	"trading-engine/internal/orderbook"
	"trading-engine/internal/types"
)

func main() {
	fmt.Println("=== ORDER BOOK MATHEMATICAL VALIDATION ===")

	// Test 1: Basic sorting validation
	testBasicSorting()

	// Test 2: Market order execution validation
	testMarketOrderExecution()

	// Test 3: Limit order validation
	testLimitOrderValidation()

	// Test 4: Edge cases validation
	testEdgeCases()

	fmt.Println("\n=== ALL MATHEMATICAL VALIDATIONS PASSED ===")
}

func testBasicSorting() {
	fmt.Println("1. Testing Basic Order Book Sorting...")

	ob := orderbook.New()

	// Create unsorted snapshot
	snapshot := types.OrderBookSnapshot{
		Symbol:    "TESTCOIN",
		Timestamp: time.Now(),
		Bids: []types.OrderBookEntry{
			{Price: 100.00, Quantity: 1.0}, // Should be 3rd
			{Price: 105.00, Quantity: 2.0}, // Should be 1st (highest bid)
			{Price: 102.50, Quantity: 1.5}, // Should be 2nd
			{Price: 95.00, Quantity: 3.0},  // Should be 4th (lowest bid)
		},
		Asks: []types.OrderBookEntry{
			{Price: 110.00, Quantity: 1.0}, // Should be 2nd
			{Price: 108.00, Quantity: 2.0}, // Should be 1st (lowest ask)
			{Price: 115.00, Quantity: 1.5}, // Should be 3rd
			{Price: 120.00, Quantity: 3.0}, // Should be 4th (highest ask)
		},
	}

	ob.Update(snapshot)

	// Validate bid sorting (descending)
	bidPrice, bidQty, exists := ob.GetBestBid()
	if !exists || bidPrice != 105.00 || bidQty != 2.0 {
		log.Fatalf("❌ Best bid incorrect: got %.2f@%.2f, expected 105.00@2.0", bidPrice, bidQty)
	}

	// Validate ask sorting (ascending)
	askPrice, askQty, exists := ob.GetBestAsk()
	if !exists || askPrice != 108.00 || askQty != 2.0 {
		log.Fatalf("❌ Best ask incorrect: got %.2f@%.2f, expected 108.00@2.0", askPrice, askQty)
	}

	// Validate spread
	spread, exists := ob.GetSpread()
	if !exists || spread != 3.00 {
		log.Fatalf("❌ Spread incorrect: got %.2f, expected 3.00", spread)
	}

	// Validate mid price
	midPrice, exists := ob.GetMidPrice()
	expectedMid := (105.00 + 108.00) / 2.0
	if !exists || midPrice != expectedMid {
		log.Fatalf("❌ Mid price incorrect: got %.2f, expected %.2f", midPrice, expectedMid)
	}

	fmt.Printf("   ✅ Best Bid: %.2f @ %.2f\n", bidPrice, bidQty)
	fmt.Printf("   ✅ Best Ask: %.2f @ %.2f\n", askPrice, askQty)
	fmt.Printf("   ✅ Spread: %.2f\n", spread)
	fmt.Printf("   ✅ Mid Price: %.2f\n", midPrice)
}

func testMarketOrderExecution() {
	fmt.Println("\n2. Testing Market Order Execution Math...")

	ob := orderbook.New()

	snapshot := types.OrderBookSnapshot{
		Symbol:    "TESTCOIN",
		Timestamp: time.Now(),
		Bids: []types.OrderBookEntry{
			{Price: 100.00, Quantity: 1.0},
			{Price: 99.50, Quantity: 2.0},
			{Price: 99.00, Quantity: 1.5},
		},
		Asks: []types.OrderBookEntry{
			{Price: 101.00, Quantity: 1.0},
			{Price: 101.50, Quantity: 2.0},
			{Price: 102.00, Quantity: 1.5},
		},
	}

	ob.Update(snapshot)

	// Test market buy order (buying 2.5 units against asks)
	// Should fill: 1.0 @ 101.00 + 1.5 @ 101.50 = total 2.5 units
	// Expected average price: (1.0*101.00 + 1.5*101.50) / 2.5 = 253.25 / 2.5 = 101.30
	fillPrice, canFill := ob.GetFillPrice(types.SideBuy, 2.5)
	expectedPrice := (1.0*101.00 + 1.5*101.50) / 2.5

	if !canFill {
		log.Fatalf("❌ Market buy should be fillable")
	}

	if fmt.Sprintf("%.2f", fillPrice) != fmt.Sprintf("%.2f", expectedPrice) {
		log.Fatalf("❌ Market buy price incorrect: got %.2f, expected %.2f", fillPrice, expectedPrice)
	}

	// Test market sell order (selling 2.5 units against bids)
	// Should fill: 1.0 @ 100.00 + 1.5 @ 99.50 = total 2.5 units
	// Expected average price: (1.0*100.00 + 1.5*99.50) / 2.5 = 249.25 / 2.5 = 99.70
	fillPrice, canFill = ob.GetFillPrice(types.SideSell, 2.5)
	expectedPrice = (1.0*100.00 + 1.5*99.50) / 2.5

	if !canFill {
		log.Fatalf("❌ Market sell should be fillable")
	}

	if fmt.Sprintf("%.2f", fillPrice) != fmt.Sprintf("%.2f", expectedPrice) {
		log.Fatalf("❌ Market sell price incorrect: got %.2f, expected %.2f", fillPrice, expectedPrice)
	}

	fmt.Printf("   ✅ Market Buy 2.5 units: %.2f (expected %.2f)\n", fillPrice, expectedPrice)

	// Test re-calculate for sell
	fillPrice, _ = ob.GetFillPrice(types.SideSell, 2.5)
	expectedPrice = (1.0*100.00 + 1.5*99.50) / 2.5
	fmt.Printf("   ✅ Market Sell 2.5 units: %.2f (expected %.2f)\n", fillPrice, expectedPrice)
}

func testLimitOrderValidation() {
	fmt.Println("\n3. Testing Limit Order Validation...")

	ob := orderbook.New()

	snapshot := types.OrderBookSnapshot{
		Symbol:    "TESTCOIN",
		Timestamp: time.Now(),
		Bids: []types.OrderBookEntry{
			{Price: 100.00, Quantity: 1.0},
			{Price: 99.50, Quantity: 2.0},
		},
		Asks: []types.OrderBookEntry{
			{Price: 101.00, Quantity: 1.0},
			{Price: 101.50, Quantity: 2.0},
		},
	}

	ob.Update(snapshot)

	// Test limit buy at 101.00 (should fill against ask)
	canFill := ob.CanFill(types.SideBuy, 101.00, 1.0)
	if !canFill {
		log.Fatalf("❌ Limit buy at 101.00 should be fillable")
	}

	// Test limit buy at 100.50 (should NOT fill - price too low)
	canFill = ob.CanFill(types.SideBuy, 100.50, 1.0)
	if canFill {
		log.Fatalf("❌ Limit buy at 100.50 should NOT be fillable")
	}

	// Test limit sell at 100.00 (should fill against bid)
	canFill = ob.CanFill(types.SideSell, 100.00, 1.0)
	if !canFill {
		log.Fatalf("❌ Limit sell at 100.00 should be fillable")
	}

	// Test limit sell at 101.50 (should NOT fill - price too high)
	canFill = ob.CanFill(types.SideSell, 101.50, 1.0)
	if canFill {
		log.Fatalf("❌ Limit sell at 101.50 should NOT be fillable")
	}

	fmt.Printf("   ✅ Limit buy at 101.00: fillable\n")
	fmt.Printf("   ✅ Limit buy at 100.50: not fillable\n")
	fmt.Printf("   ✅ Limit sell at 100.00: fillable\n")
	fmt.Printf("   ✅ Limit sell at 101.50: not fillable\n")
}

func testEdgeCases() {
	fmt.Println("\n4. Testing Edge Cases...")

	ob := orderbook.New()

	// Empty order book
	_, _, exists := ob.GetBestBid()
	if exists {
		log.Fatalf("❌ Empty order book should not have best bid")
	}

	_, _, exists = ob.GetBestAsk()
	if exists {
		log.Fatalf("❌ Empty order book should not have best ask")
	}

	_, exists = ob.GetSpread()
	if exists {
		log.Fatalf("❌ Empty order book should not have spread")
	}

	// Insufficient liquidity
	snapshot := types.OrderBookSnapshot{
		Symbol:    "TESTCOIN",
		Timestamp: time.Now(),
		Bids: []types.OrderBookEntry{
			{Price: 100.00, Quantity: 1.0},
		},
		Asks: []types.OrderBookEntry{
			{Price: 101.00, Quantity: 1.0},
		},
	}

	ob.Update(snapshot)

	// Try to fill more than available
	_, canFill := ob.GetFillPrice(types.SideBuy, 2.0)
	if canFill {
		log.Fatalf("❌ Should not be able to fill 2.0 when only 1.0 available")
	}

	_, canFill = ob.GetFillPrice(types.SideSell, 2.0)
	if canFill {
		log.Fatalf("❌ Should not be able to fill 2.0 when only 1.0 available")
	}

	fmt.Printf("   ✅ Empty order book handled correctly\n")
	fmt.Printf("   ✅ Insufficient liquidity detected correctly\n")
}
