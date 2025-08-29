package main

import (
	"fmt"
	"log"
	"time"
	"trading-engine/internal/orderbook"
	"trading-engine/internal/types"
)

func main() {
	fmt.Println("=== DETAILED MARKET ORDER VALIDATION ===")

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

	fmt.Println("Order Book State:")
	fmt.Println("ASKS:")
	fmt.Println("  102.00 @ 1.5")
	fmt.Println("  101.50 @ 2.0")
	fmt.Println("  101.00 @ 1.0")
	fmt.Println("--- SPREAD ---")
	fmt.Println("BIDS:")
	fmt.Println("  100.00 @ 1.0")
	fmt.Println("   99.50 @ 2.0")
	fmt.Println("   99.00 @ 1.5")

	fmt.Println("\nTesting Market BUY order for 2.5 units:")
	fmt.Println("Should fill against asks from best price upward:")
	fmt.Println("  1.0 units @ 101.00 = 101.00")
	fmt.Println("  1.5 units @ 101.50 = 152.25")
	fmt.Println("  Total cost: 253.25")
	fmt.Println("  Average price: 253.25 / 2.5 = 101.30")

	fillPrice, canFill := ob.GetFillPrice(types.SideBuy, 2.5)
	fmt.Printf("  Actual result: %.2f, canFill: %v\n", fillPrice, canFill)

	expectedPrice := (1.0*101.00 + 1.5*101.50) / 2.5
	fmt.Printf("  Expected: %.2f\n", expectedPrice)

	if !canFill {
		log.Fatalf("❌ Should be able to fill")
	}

	if fmt.Sprintf("%.2f", fillPrice) != fmt.Sprintf("%.2f", expectedPrice) {
		fmt.Printf("❌ Price mismatch! Got %.6f, expected %.6f\n", fillPrice, expectedPrice)
	} else {
		fmt.Printf("✅ Market BUY calculation correct\n")
	}

	fmt.Println("\nTesting Market SELL order for 2.5 units:")
	fmt.Println("Should fill against bids from best price downward:")
	fmt.Println("  1.0 units @ 100.00 = 100.00")
	fmt.Println("  1.5 units @ 99.50 = 149.25")
	fmt.Println("  Total proceeds: 249.25")
	fmt.Println("  Average price: 249.25 / 2.5 = 99.70")

	fillPrice, canFill = ob.GetFillPrice(types.SideSell, 2.5)
	fmt.Printf("  Actual result: %.2f, canFill: %v\n", fillPrice, canFill)

	expectedPrice = (1.0*100.00 + 1.5*99.50) / 2.5
	fmt.Printf("  Expected: %.2f\n", expectedPrice)

	if !canFill {
		log.Fatalf("❌ Should be able to fill")
	}

	if fmt.Sprintf("%.2f", fillPrice) != fmt.Sprintf("%.2f", expectedPrice) {
		fmt.Printf("❌ Price mismatch! Got %.6f, expected %.6f\n", fillPrice, expectedPrice)
	} else {
		fmt.Printf("✅ Market SELL calculation correct\n")
	}
}
