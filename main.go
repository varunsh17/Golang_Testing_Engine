package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
	"trading-engine/internal/broker"
	"trading-engine/internal/engine"
	"trading-engine/internal/feed"
	"trading-engine/internal/orderbook"
	"trading-engine/internal/strategy"
	"trading-engine/internal/types"
)

type TradingSession struct {
	ID            string
	OrderbookFile string
	Config        SessionConfig
	Results       SessionResults
}

type SessionConfig struct {
	EntryPrice      float64
	OrderSize       float64
	StopLoss        float64
	TakeProfit      float64
	LiquidityThresh float64
	MaxHoldTime     time.Duration
	OutputFile      string
}

type SessionResults struct {
	TradeLog    []types.Execution
	TotalPnL    float64
	TotalTrades int
	Duration    time.Duration
	Success     bool
	Error       error
}

func main() {
	// CLI flags
	var (
		concurrent      = flag.Bool("concurrent", false, "Run all 3 samples concurrently")
		sessionID       = flag.String("session", "", "Specific session ID to run (btc, eth, ada)")
		orderbookFile   = flag.String("orderbook", "data/sample1.json", "Path to orderbook JSON file")
		entryPrice      = flag.Float64("entry", 0, "Entry price (0 for auto)")
		orderSize       = flag.Float64("size", 100, "Order size")
		stopLoss        = flag.Float64("stop", 0.02, "Stop loss percentage (0.02 = 2%)")
		takeProfit      = flag.Float64("profit", 0.05, "Take profit percentage (0.05 = 5%)")
		liquidityThresh = flag.Float64("liquidity", 1000, "Minimum liquidity threshold")
		maxHoldTime     = flag.Duration("hold", 30*time.Second, "Maximum hold time")
		outputFile      = flag.String("output", "trades.csv", "Output CSV file for trades")
	)
	flag.Parse()

	fmt.Println("🔥 GO TRADING ENGINE - Goroutines & Channels Demo")
	fmt.Println("================================================")

	if *concurrent {
		runConcurrentSessions()
		return
	} else if *sessionID != "" {
		runSpecificSession(*sessionID)
		return
	}

	// Single session mode (original functionality)
	runSingleSession(*orderbookFile, *entryPrice, *orderSize, *stopLoss,
		*takeProfit, *liquidityThresh, *maxHoldTime, *outputFile)
}

func runConcurrentSessions() {
	fmt.Println("🚀 STARTING CONCURRENT TRADING SESSIONS")
	fmt.Println("======================================")

	// Define 3 concurrent trading sessions
	sessions := []TradingSession{
		{
			ID:            "BTC-Aggressive",
			OrderbookFile: "data/sample1.json",
			Config: SessionConfig{
				EntryPrice:      0, // Auto entry
				OrderSize:       2.5,
				StopLoss:        0.015, // 1.5%
				TakeProfit:      0.04,  // 4%
				LiquidityThresh: 800,
				MaxHoldTime:     8 * time.Second,
				OutputFile:      "concurrent_btc_trades.csv",
			},
		},
		{
			ID:            "ETH-Conservative",
			OrderbookFile: "data/sample2.json",
			Config: SessionConfig{
				EntryPrice:      3000, // Specific entry
				OrderSize:       5.0,
				StopLoss:        0.01,  // 1%
				TakeProfit:      0.025, // 2.5%
				LiquidityThresh: 1200,
				MaxHoldTime:     12 * time.Second,
				OutputFile:      "concurrent_eth_trades.csv",
			},
		},
		{
			ID:            "ADA-HighFreq",
			OrderbookFile: "data/sample3.json",
			Config: SessionConfig{
				EntryPrice:      0, // Auto entry
				OrderSize:       8000,
				StopLoss:        0.005, // 0.5%
				TakeProfit:      0.015, // 1.5%
				LiquidityThresh: 2000,
				MaxHoldTime:     6 * time.Second,
				OutputFile:      "concurrent_ada_trades.csv",
			},
		},
	}

	fmt.Printf("📋 Launching %d concurrent trading sessions:\n", len(sessions))
	for i, session := range sessions {
		fmt.Printf("   %d. %s (File: %s)\n", i+1, session.ID, session.OrderbookFile)
	}
	fmt.Println()

	// GOROUTINES & CHANNELS DEMONSTRATION:
	// 1. Results channel to collect outputs from all goroutines
	resultsChan := make(chan TradingSession, len(sessions))

	// 2. Progress channel to show real-time updates
	progressChan := make(chan string, 50)

	// 3. WaitGroup to coordinate goroutine completion
	var wg sync.WaitGroup

	startTime := time.Now()

	// GOROUTINE 1: Progress reporter
	go func() {
		fmt.Println("📊 Real-time session updates:")
		for update := range progressChan {
			fmt.Printf("   %s %s\n", time.Now().Format("15:04:05"), update)
		}
	}()

	// GOROUTINES 2-4: Trading sessions (one per sample file)
	for i := range sessions {
		wg.Add(1)
		go func(session TradingSession, sessionNum int) {
			defer wg.Done()

			sessionStart := time.Now()
			progressChan <- fmt.Sprintf("🟢 [%s] Starting trading session", session.ID)

			// Run the trading session (contains more goroutines internally)
			result := runTradingSession(session, progressChan)
			result.Results.Duration = time.Since(sessionStart)

			progressChan <- fmt.Sprintf("✅ [%s] Completed in %v - %d trades",
				session.ID, result.Results.Duration, result.Results.TotalTrades)

			// Send result through channel
			resultsChan <- result
		}(sessions[i], i+1)
	}

	// GOROUTINE 5: Wait for all sessions and close channels
	go func() {
		wg.Wait()
		close(resultsChan)
		close(progressChan)
	}()

	// Collect results from all concurrent sessions
	var allResults []TradingSession
	for result := range resultsChan {
		allResults = append(allResults, result)
	}

	totalDuration := time.Since(startTime)

	// Display comprehensive summary
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("🏁 CONCURRENT EXECUTION RESULTS")
	fmt.Println(strings.Repeat("=", 70))

	totalTrades := 0
	totalPnL := 0.0
	successfulSessions := 0

	for _, result := range allResults {
		if result.Results.Success {
			successfulSessions++
			totalTrades += result.Results.TotalTrades
			totalPnL += result.Results.TotalPnL

			fmt.Printf("\n📈 %s:\n", result.ID)
			fmt.Printf("   📁 Data Source: %s\n", result.OrderbookFile)
			fmt.Printf("   💹 Executed Trades: %d\n", result.Results.TotalTrades)
			fmt.Printf("   💰 Session P&L: $%.2f\n", result.Results.TotalPnL)
			fmt.Printf("   ⏱️  Execution Time: %v\n", result.Results.Duration)
			fmt.Printf("   📊 Strategy: Entry=%.0f, Size=%.1f, Stop=%.1f%%, Profit=%.1f%%\n",
				result.Config.EntryPrice, result.Config.OrderSize,
				result.Config.StopLoss*100, result.Config.TakeProfit*100)
			fmt.Printf("   📄 Trade Log: %s\n", result.Config.OutputFile)
		} else {
			fmt.Printf("\n❌ %s: Failed with error - %v\n", result.ID, result.Results.Error)
		}
	}

	// Concurrency analysis
	sequentialTime := float64(len(sessions)) * 10.0 // Estimated sequential time
	actualTime := totalDuration.Seconds()
	speedup := sequentialTime / actualTime

	fmt.Printf("\n🎯 CONCURRENCY PERFORMANCE:\n")
	fmt.Printf("   ✅ Successful Sessions: %d/%d\n", successfulSessions, len(sessions))
	fmt.Printf("   📈 Total Trades Executed: %d\n", totalTrades)
	fmt.Printf("   💵 Combined Portfolio P&L: $%.2f\n", totalPnL)
	fmt.Printf("   🚄 Total Wall-Clock Time: %v\n", totalDuration)
	fmt.Printf("   ⚡ Estimated Speedup: %.1fx faster than sequential\n", speedup)
	fmt.Printf("   🔧 Goroutines Used: %d main sessions + internal goroutines per session\n", len(sessions))
	fmt.Printf("   📡 Channels Used: Results, Progress, + 4 channels per session\n")

	fmt.Println("\n🧠 GOROUTINES & CHANNELS ARCHITECTURE:")
	fmt.Println("   • Main goroutine: Orchestrates and collects results")
	fmt.Println("   • Progress goroutine: Real-time status updates via channel")
	fmt.Println("   • Session goroutines: One per trading session (3 total)")
	fmt.Println("   • Per-session goroutines: Feed, Engine, Strategy, Broker (4 each)")
	fmt.Println("   • Channel communication: orderbook updates, trade signals, executions")
	fmt.Println("   • Total concurrent goroutines: ~17 running simultaneously!")
}

func runTradingSession(session TradingSession, progressChan chan<- string) TradingSession {
	// Initialize orderbook for this session
	ob := orderbook.New()

	// CHANNELS for inter-component communication (core of the architecture)
	orderbookUpdates := make(chan types.OrderBookSnapshot, 100)
	tradeSignals := make(chan types.TradeSignal, 10)
	executions := make(chan types.Execution, 10)
	strategyExecutions := make(chan types.Execution, 10)
	done := make(chan bool)

	if progressChan != nil {
		progressChan <- fmt.Sprintf("🔧 [%s] Initialized channels and orderbook", session.ID)
	}

	// Initialize components
	feedInstance := feed.New(session.OrderbookFile, orderbookUpdates)

	strategyConfig := strategy.Config{
		EntryPrice:      session.Config.EntryPrice,
		OrderSize:       session.Config.OrderSize,
		StopLoss:        session.Config.StopLoss,
		TakeProfit:      session.Config.TakeProfit,
		LiquidityThresh: session.Config.LiquidityThresh,
		MaxHoldTime:     session.Config.MaxHoldTime,
	}
	strategyInstance := strategy.New(strategyConfig, tradeSignals, strategyExecutions)
	brokerInstance := broker.New(ob, tradeSignals, executions)
	engineInstance := engine.New(ob, orderbookUpdates, done)

	if progressChan != nil {
		progressChan <- fmt.Sprintf("⚙️  [%s] Starting 4 component goroutines", session.ID)
	}

	// Start all components in separate GOROUTINES
	go feedInstance.Start()     // GOROUTINE: Feed data from JSON
	go engineInstance.Start()   // GOROUTINE: Process orderbook updates
	go strategyInstance.Start() // GOROUTINE: Generate trade signals
	go brokerInstance.Start()   // GOROUTINE: Execute trades

	// Track results through CHANNEL communication
	var tradeLog []types.Execution
	executionsDone := make(chan bool)

	// GOROUTINE: Execution collector and broadcaster
	go func() {
		tradeCount := 0
		for execution := range executions {
			tradeCount++
			if progressChan != nil {
				progressChan <- fmt.Sprintf("💱 [%s] Trade #%d: %s %.2f @ $%.2f",
					session.ID, tradeCount, execution.Side, execution.Quantity, execution.Price)
			}

			// Send to strategy via CHANNEL
			select {
			case strategyExecutions <- execution:
			default:
				log.Printf("[%s] Strategy executions channel full", session.ID)
			}

			// Collect for results
			tradeLog = append(tradeLog, execution)
		}
		close(strategyExecutions)
		executionsDone <- true
	}()

	// Wait for feed completion via CHANNEL
	<-done
	if progressChan != nil {
		progressChan <- fmt.Sprintf("📡 [%s] Data feed completed", session.ID)
	}

	// Allow strategy to finish processing
	time.Sleep(session.Config.MaxHoldTime + 2*time.Second)
	close(tradeSignals)

	// Wait for executions to finish via CHANNEL
	<-executionsDone
	if progressChan != nil {
		progressChan <- fmt.Sprintf("🎯 [%s] All executions completed", session.ID)
	}

	// Calculate P&L
	var totalPnL float64
	buyTotal := 0.0
	sellTotal := 0.0

	for _, trade := range tradeLog {
		if trade.Side == types.SideBuy {
			buyTotal += trade.Price * trade.Quantity
		} else {
			sellTotal += trade.Price * trade.Quantity
		}
	}
	totalPnL = sellTotal - buyTotal

	// Write trade log to CSV
	var err error
	if len(tradeLog) > 0 {
		err = writeTradeLog(session.Config.OutputFile, tradeLog)
		if err == nil && progressChan != nil {
			progressChan <- fmt.Sprintf("📝 [%s] Trade log written to %s",
				session.ID, session.Config.OutputFile)
		}
	}

	// Return results
	session.Results = SessionResults{
		TradeLog:    tradeLog,
		TotalPnL:    totalPnL,
		TotalTrades: len(tradeLog),
		Success:     err == nil,
		Error:       err,
	}

	return session
}

func runSpecificSession(sessionID string) {
	fmt.Printf("🎯 Running specific session: %s\n", sessionID)

	sessions := map[string]TradingSession{
		"btc": {
			ID:            "BTC-Test",
			OrderbookFile: "data/sample1.json",
			Config: SessionConfig{
				EntryPrice: 0, OrderSize: 1.5, StopLoss: 0.02, TakeProfit: 0.05,
				LiquidityThresh: 1000, MaxHoldTime: 6 * time.Second,
				OutputFile: "btc_test_trades.csv",
			},
		},
		"eth": {
			ID:            "ETH-Test",
			OrderbookFile: "data/sample2.json",
			Config: SessionConfig{
				EntryPrice: 3000, OrderSize: 3.0, StopLoss: 0.015, TakeProfit: 0.03,
				LiquidityThresh: 1000, MaxHoldTime: 8 * time.Second,
				OutputFile: "eth_test_trades.csv",
			},
		},
		"ada": {
			ID:            "ADA-Test",
			OrderbookFile: "data/sample3.json",
			Config: SessionConfig{
				EntryPrice: 0, OrderSize: 2000, StopLoss: 0.01, TakeProfit: 0.02,
				LiquidityThresh: 1000, MaxHoldTime: 5 * time.Second,
				OutputFile: "ada_test_trades.csv",
			},
		},
	}

	session, exists := sessions[sessionID]
	if !exists {
		fmt.Printf("❌ Unknown session ID: %s\n", sessionID)
		fmt.Printf("Available sessions: btc, eth, ada\n")
		return
	}

	result := runTradingSession(session, nil)

	if result.Results.Success {
		fmt.Printf("✅ Session completed successfully!\n")
		fmt.Printf("   💹 Trades: %d\n", result.Results.TotalTrades)
		fmt.Printf("   💰 P&L: %.2f\n", result.Results.TotalPnL)
		fmt.Printf("   📄 Output: %s\n", result.Config.OutputFile)
	} else {
		fmt.Printf("❌ Session failed: %v\n", result.Results.Error)
	}
}

func runSingleSession(orderbookFile string, entryPrice, orderSize, stopLoss,
	takeProfit, liquidityThresh float64, maxHoldTime time.Duration, outputFile string) {

	session := TradingSession{
		ID:            "Single",
		OrderbookFile: orderbookFile,
		Config: SessionConfig{
			EntryPrice:      entryPrice,
			OrderSize:       orderSize,
			StopLoss:        stopLoss,
			TakeProfit:      takeProfit,
			LiquidityThresh: liquidityThresh,
			MaxHoldTime:     maxHoldTime,
			OutputFile:      outputFile,
		},
	}

	fmt.Printf("🔧 Starting single trading session with:\n")
	fmt.Printf("  📁 Orderbook file: %s\n", session.OrderbookFile)
	fmt.Printf("  💰 Entry price: %.2f\n", session.Config.EntryPrice)
	fmt.Printf("  📊 Order size: %.2f\n", session.Config.OrderSize)
	fmt.Printf("  � Stop loss: %.1f%%\n", session.Config.StopLoss*100)
	fmt.Printf("  🎯 Take profit: %.1f%%\n", session.Config.TakeProfit*100)
	fmt.Printf("  💧 Liquidity threshold: %.0f\n", session.Config.LiquidityThresh)
	fmt.Printf("  ⏰ Max hold time: %v\n", session.Config.MaxHoldTime)
	fmt.Printf("  �📄 Output file: %s\n", session.Config.OutputFile)
	fmt.Println()

	result := runTradingSession(session, nil)

	// Print summary
	fmt.Printf("\n=== TRADING SUMMARY ===\n")
	fmt.Printf("Total trades: %d\n", result.Results.TotalTrades)
	fmt.Printf("Total P&L: %.2f\n", result.Results.TotalPnL)
	if result.Results.Success {
		fmt.Printf("Trade log written to: %s\n", session.Config.OutputFile)
	}
}

func writeTradeLog(filename string, trades []types.Execution) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Timestamp", "Side", "Price", "Quantity", "Symbol"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write trades
	for _, trade := range trades {
		record := []string{
			trade.Timestamp.Format(time.RFC3339),
			string(trade.Side),
			fmt.Sprintf("%.8f", trade.Price),
			fmt.Sprintf("%.8f", trade.Quantity),
			trade.Symbol,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}
