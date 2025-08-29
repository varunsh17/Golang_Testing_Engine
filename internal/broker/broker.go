package broker

import (
	"log"
	"time"
	"trading-engine/internal/orderbook"
	"trading-engine/internal/types"
)

// Broker handles order execution and matching
type Broker struct {
	orderbook  *orderbook.OrderBook
	signals    <-chan types.TradeSignal
	executions chan<- types.Execution
}

// New creates a new broker instance
func New(ob *orderbook.OrderBook, signals <-chan types.TradeSignal, executions chan<- types.Execution) *Broker {
	return &Broker{
		orderbook:  ob,
		signals:    signals,
		executions: executions,
	}
}

// Start begins processing trade signals
func (b *Broker) Start() {
	log.Println("Broker started")

	for signal := range b.signals {
		log.Printf("Broker received signal: %+v", signal)

		execution := b.executeOrder(signal)
		if execution != nil {
			select {
			case b.executions <- *execution:
				log.Printf("Execution sent: %+v", *execution)
			default:
				log.Printf("Failed to send execution - channel full")
			}
		}
	}

	log.Println("Broker finished")
	close(b.executions)
}

// executeOrder attempts to execute a trade signal
func (b *Broker) executeOrder(signal types.TradeSignal) *types.Execution {
	// Determine execution price
	var execPrice float64
	var canFill bool

	if signal.Price == 0 {
		// Market order
		execPrice, canFill = b.orderbook.GetFillPrice(signal.Side, signal.Quantity)
		if !canFill {
			log.Printf("Market order cannot be filled: insufficient liquidity")
			return nil
		}
	} else {
		// Limit order
		execPrice = signal.Price
		canFill = b.orderbook.CanFill(signal.Side, signal.Price, signal.Quantity)
		if !canFill {
			log.Printf("Limit order cannot be filled at %.2f", signal.Price)

			// For simulation purposes, we'll still execute at best available price
			if bestPrice, canFillAtBest := b.orderbook.GetFillPrice(signal.Side, signal.Quantity); canFillAtBest {
				log.Printf("Executing at best available price: %.2f", bestPrice)
				execPrice = bestPrice
				canFill = true
			} else {
				return nil
			}
		}
	}

	if !canFill {
		log.Printf("Order cannot be executed: no liquidity")
		return nil
	}

	// Create execution
	execution := &types.Execution{
		Symbol:    signal.Symbol,
		Side:      signal.Side,
		Price:     execPrice,
		Quantity:  signal.Quantity,
		Timestamp: time.Now(),
	}

	log.Printf("Order executed: %s %.2f @ %.2f",
		string(execution.Side), execution.Quantity, execution.Price)

	// In a real system, we would update the order book by removing the filled quantities
	// For this simulation, we'll leave the order book unchanged

	return execution
}
