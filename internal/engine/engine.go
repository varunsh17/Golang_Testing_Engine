package engine

import (
	"log"
	"trading-engine/internal/orderbook"
	"trading-engine/internal/types"
)

// Engine processes order book updates and maintains the current state
type Engine struct {
	orderbook *orderbook.OrderBook
	updates   <-chan types.OrderBookSnapshot
	done      chan<- bool
}

// New creates a new engine instance
func New(ob *orderbook.OrderBook, updates <-chan types.OrderBookSnapshot, done chan<- bool) *Engine {
	return &Engine{
		orderbook: ob,
		updates:   updates,
		done:      done,
	}
}

// Start begins processing order book updates
func (e *Engine) Start() {
	log.Println("Engine started")

	updateCount := 0

	for snapshot := range e.updates {
		updateCount++

		// Update the order book
		e.orderbook.Update(snapshot)

		// Log periodic updates
		if updateCount%10 == 0 {
			log.Printf("Processed %d order book updates", updateCount)

			if bid, bidQty, bidExists := e.orderbook.GetBestBid(); bidExists {
				if ask, askQty, askExists := e.orderbook.GetBestAsk(); askExists {
					spread, _ := e.orderbook.GetSpread()
					log.Printf("Best: %.2f(%.2f) - %.2f(%.2f), Spread: %.2f",
						bid, bidQty, ask, askQty, spread)
				}
			}
		}
	}

	log.Printf("Engine finished processing %d total updates", updateCount)
	e.done <- true
}
