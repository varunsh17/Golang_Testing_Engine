package strategy

import (
	"log"
	"time"
	"trading-engine/internal/types"
)

// Config holds strategy configuration
type Config struct {
	EntryPrice      float64
	OrderSize       float64
	StopLoss        float64
	TakeProfit      float64
	LiquidityThresh float64
	MaxHoldTime     time.Duration
}

// Strategy implements a multi-factor trading strategy
type Strategy struct {
	config     Config
	signals    chan<- types.TradeSignal
	executions <-chan types.Execution
	position   *types.Position
}

// New creates a new strategy instance
func New(config Config, signals chan<- types.TradeSignal, executions <-chan types.Execution) *Strategy {
	return &Strategy{
		config:     config,
		signals:    signals,
		executions: executions,
	}
}

// Start begins the strategy execution
func (s *Strategy) Start() {
	log.Println("Strategy started")

	// Listen for executions to track position
	go s.handleExecutions()

	// For this simulation, we'll generate a simple buy signal after a delay
	// Wait for the feed to start publishing data
	time.Sleep(500 * time.Millisecond)

	entryPrice := s.config.EntryPrice
	if entryPrice == 0 {
		// Auto-entry mode - use a market order
		log.Println("Generating market buy signal (auto-entry)")
		signal := types.TradeSignal{
			Symbol:    "BTCUSD", // Default symbol
			Side:      types.SideBuy,
			Price:     0, // Market order
			Quantity:  s.config.OrderSize,
			Timestamp: time.Now(),
		}
		select {
		case s.signals <- signal:
			log.Println("Buy signal sent")
		default:
			log.Println("Failed to send buy signal - channel full")
		}
	} else {
		// Use specified entry price
		log.Printf("Generating limit buy signal at %.2f", entryPrice)
		signal := types.TradeSignal{
			Symbol:    "BTCUSD",
			Side:      types.SideBuy,
			Price:     entryPrice,
			Quantity:  s.config.OrderSize,
			Timestamp: time.Now(),
		}
		select {
		case s.signals <- signal:
			log.Println("Limit buy signal sent")
		default:
			log.Println("Failed to send limit buy signal - channel full")
		}
	}
}

// handleExecutions processes trade executions and manages positions
func (s *Strategy) handleExecutions() {
	log.Println("Strategy execution handler started")
	for execution := range s.executions {
		log.Printf("Strategy received execution: %+v", execution)

		if s.position == nil && execution.Side == types.SideBuy {
			// Opening position
			s.position = &types.Position{
				Symbol:     execution.Symbol,
				Quantity:   execution.Quantity,
				EntryPrice: execution.Price,
				EntryTime:  execution.Timestamp,
			}

			log.Printf("Position opened: %.2f @ %.2f", s.position.Quantity, s.position.EntryPrice)

			// Schedule exit signals
			go s.scheduleExitSignals()

		} else if s.position != nil && execution.Side == types.SideSell {
			// Closing position
			log.Printf("Position closed: %.2f @ %.2f", execution.Quantity, execution.Price)

			// Calculate PnL
			pnl := (execution.Price - s.position.EntryPrice) * execution.Quantity
			holdTime := execution.Timestamp.Sub(s.position.EntryTime)

			log.Printf("Trade PnL: %.2f (held for %v)", pnl, holdTime)

			s.position = nil
		}
	}
	log.Println("Strategy execution handler finished")
}

// scheduleExitSignals generates exit signals based on strategy rules
func (s *Strategy) scheduleExitSignals() {
	if s.position == nil {
		return
	}

	log.Println("Scheduling exit signals...")

	// Time-based exit
	go func() {
		time.Sleep(s.config.MaxHoldTime)
		if s.position != nil {
			log.Println("Generating time-based exit signal")
			signal := types.TradeSignal{
				Symbol:    s.position.Symbol,
				Side:      types.SideSell,
				Price:     0, // Market order
				Quantity:  s.position.Quantity,
				Timestamp: time.Now(),
			}
			select {
			case s.signals <- signal:
				log.Println("Exit signal sent")
			default:
				log.Println("Failed to send exit signal - channel full")
			}
		}
	}()

	// Take profit exit (trigger first for demo)
	if s.config.TakeProfit > 0 {
		go func() {
			time.Sleep(2 * time.Second) // Trigger before time-based exit

			if s.position != nil {
				profitPrice := s.position.EntryPrice * (1 + s.config.TakeProfit)
				log.Printf("Generating take-profit exit signal at %.2f", profitPrice)
				signal := types.TradeSignal{
					Symbol:    s.position.Symbol,
					Side:      types.SideSell,
					Price:     0, // Use market order for demo
					Quantity:  s.position.Quantity,
					Timestamp: time.Now(),
				}
				select {
				case s.signals <- signal:
					log.Println("Take-profit signal sent")
				default:
					log.Println("Failed to send take-profit signal - channel full")
				}
			}
		}()
	}

	// Stop loss exit
	if s.config.StopLoss > 0 {
		go func() {
			time.Sleep(3 * time.Second) // Simulate some time for price movement

			if s.position != nil {
				stopPrice := s.position.EntryPrice * (1 - s.config.StopLoss)
				log.Printf("Generating stop-loss exit signal at %.2f", stopPrice)
				signal := types.TradeSignal{
					Symbol:    s.position.Symbol,
					Side:      types.SideSell,
					Price:     0, // Use market order for demo
					Quantity:  s.position.Quantity,
					Timestamp: time.Now(),
				}
				select {
				case s.signals <- signal:
					log.Println("Stop-loss signal sent")
				default:
					log.Println("Failed to send stop-loss signal - channel full")
				}
			}
		}()
	}
}
