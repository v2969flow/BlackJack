package processors

import (
	"context"
	"log"
	"sync"
	"time"
)

// WashCycleProcessor — the "typing-mechanic-as-meter" formalized.
// Repeated actions (pulses) from family nodes build a wash meter.
// When meter hits threshold, noise is cleansed and meaning is clarified (Samekh rotation continues).
// This is the COSMO wash cycle turned into concrete concurrent filter.

type WashedState struct {
	Source    string    `json:"source"`
	Cleaned   bool      `json:"cleaned"`
	Meter     float64   `json:"meter"`
	Timestamp time.Time `json:"timestamp"`
	Notes     string    `json:"notes"`
}

func WashCycleProcessor(ctx context.Context, wg *sync.WaitGroup, in <-chan HexAnchorPulse, out chan<- WashedState) {
	defer wg.Done()
	defer close(out)

	var washMeter float64 = 0.0
	const washThreshold = 0.85
	const decayPerSecond = 0.03

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[WashCycle] Canal draining — final wash performed")
			return

		case pulse, ok := <-in:
			if !ok {
				return
			}
			// Every incoming pulse (family action or hex anchor) adds to the meter
			// Low tension pulses (Samekh rule) give bigger wash contribution
			contribution := 0.12
			if pulse.Tension < 0.15 {
				contribution = 0.28 // Samekh passes — bigger cleansing
			}
			washMeter = min(1.0, washMeter+contribution)
			log.Printf("[WashCycle] Pulse from %s → meter=%.2f (tension=%.2f)", pulse.NodeID, washMeter, pulse.Tension)

			if washMeter >= washThreshold {
				out <- WashedState{
					Source:    pulse.NodeID,
					Cleaned:   true,
					Meter:     washMeter,
					Timestamp: time.Now(),
					Notes:     "Wash cycle complete — noise filtered, meaning clarified for BioToken stream",
				}
				washMeter = 0.15 // partial reset keeps the spiral gentle
			}

		case <-ticker.C:
			// Natural decay if no pulses — keeps the mechanic alive but not infinite
			washMeter = max(0.0, washMeter-decayPerSecond)
		}
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}