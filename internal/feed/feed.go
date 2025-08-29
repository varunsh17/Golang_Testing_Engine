package feed

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"
	"trading-engine/internal/types"
)

// Feed reads order book data from a JSON file and publishes updates
type Feed struct {
	filename string
	updates  chan<- types.OrderBookSnapshot
	data     []types.OrderBookSnapshot
}

// New creates a new feed instance
func New(filename string, updates chan<- types.OrderBookSnapshot) *Feed {
	return &Feed{
		filename: filename,
		updates:  updates,
	}
}

// Start begins the feed simulation
func (f *Feed) Start() {
	defer close(f.updates)

	// Load data from file
	if err := f.loadData(); err != nil {
		log.Printf("Error loading feed data: %v", err)
		return
	}

	log.Printf("Feed loaded %d snapshots from %s", len(f.data), f.filename)

	// Publish snapshots with timing to simulate real-time feed
	baseTime := time.Now()

	for i, snapshot := range f.data {
		// Adjust timestamp to simulate real-time progression
		snapshot.Timestamp = baseTime.Add(time.Duration(i) * 100 * time.Millisecond)

		select {
		case f.updates <- snapshot:
			log.Printf("Published snapshot %d: %s @ %v", i+1, snapshot.Symbol, snapshot.Timestamp.Format("15:04:05.000"))
		default:
			log.Printf("Channel full, dropping snapshot %d", i+1)
		}

		// Simulate real-time delay
		time.Sleep(100 * time.Millisecond)
	}

	log.Println("Feed completed")
}

// loadData loads order book snapshots from JSON file
func (f *Feed) loadData() error {
	data, err := ioutil.ReadFile(f.filename)
	if err != nil {
		return err
	}

	// Try to parse as array first
	var snapshots []types.OrderBookSnapshot
	if err := json.Unmarshal(data, &snapshots); err != nil {
		// If that fails, try as single snapshot
		var snapshot types.OrderBookSnapshot
		if err := json.Unmarshal(data, &snapshot); err != nil {
			return err
		}
		snapshots = []types.OrderBookSnapshot{snapshot}
	}

	f.data = snapshots
	return nil
}
