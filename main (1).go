package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"sovereign-l7/gpm/internal/filters"
	"sovereign-l7/gpm/internal/manager"
	"sovereign-l7/gpm/internal/processors"
)

// G-P-M Pipeline: Goroutines / Processors / Manager
// Canal and Bridge filtering metaphor implemented with Go channels.
// Hex anchor states from pulse_receiver.py flow in → washed → BioToken stream out.
// NFT Harmonic / Farseer layer verifies meaning before BioMe identity mint/anchor.

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("[GPM] Thurisaz boundary received — gracefully draining canals...")
		cancel()
	}()

	log.Println("╔════════════════════════════════════════════════════════════╗")
	log.Println("║  SOVEREIGN L7 — G-P-M PIPELINE (Go)                        ║")
	log.Println("║  Goroutines • Processors • Manager                         ║")
	log.Println("║  Canal & Bridge Filtering + Wash Cycle + BioToken Stream   ║")
	log.Println("╚════════════════════════════════════════════════════════════╝")

	// === CANALS (buffered channels) ===
	pulseCanal := make(chan processors.HexAnchorPulse, 64)      // raw from pulse_receiver
	washCanal := make(chan processors.WashedState, 32)          // after wash cycle
	biotokenCanal := make(chan manager.BioTokenEvent, 16)       // verified BioMe identity events
	nftVerifyCanal := make(chan manager.NFTMeaningEvent, 8)     // farseer harmonic verification

	var wg sync.WaitGroup

	// === BRIDGES (fan-in / fan-out + select patterns) ===
	// Bridge 1: Pulse Receiver → multiple processors
	wg.Add(1)
	go filters.CanalBridge(ctx, &wg, pulseCanal, []chan processors.HexAnchorPulse{
		makeProcessorFan(pulseCanal, "pulse-validator"),
		makeProcessorFan(pulseCanal, "hex-anchor-normalizer"),
	})

	// Bridge 2: Wash Cycle processor (the "typing mechanic as meter")
	wg.Add(1)
	go processors.WashCycleProcessor(ctx, &wg, pulseCanal, washCanal)

	// Bridge 3: NFT Harmonic Farseer (meaning verification layer)
	wg.Add(1)
	go processors.NFTHarmonicFarseer(ctx, &wg, washCanal, nftVerifyCanal)

	// Bridge 4: BioToken / BioMe Identity Stream (final verified output)
	wg.Add(1)
	go manager.BioTokenManager(ctx, &wg, washCanal, nftVerifyCanal, biotokenCanal)

	// === HTTP API for Node persona_engine.js and Godot to interact ===
	http.HandleFunc("/pulse", handlePulseIngest(pulseCanal))           // called by extended pulse_receiver.py
	http.HandleFunc("/state", handleState(biotokenCanal))              // live BioToken stream state
	http.HandleFunc("/wash", handleWashTrigger(washCanal))             // manual family wash cycle trigger
	http.HandleFunc("/nft/verify", handleNFTHarmonic(nftVerifyCanal))  // farseer verification hook

	server := &http.Server{Addr: ":4444", Handler: nil}

	go func() {
		log.Println("[GPM] Canal API listening on :4444 — Node persona engine & Godot connect here")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[GPM] HTTP server error: %v", err)
		}
	}()

	// === Demo / Heartbeat (remove in prod) ===
	go func() {
		ticker := time.NewTicker(8 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Simulate a family node pulse (Real-Time Family Connection)
				select {
				case pulseCanal <- processors.HexAnchorPulse{
					NodeID:    "Elisha-Connector",
					Hex:       "0x" + fmt.Sprintf("%x", time.Now().UnixNano()),
					Tension:   0.12, // low tension = Samekh passes
					Timestamp: time.Now(),
				}:
					log.Println("[GPM] Family pulse bridged into canal (Elisha)")
				default:
				}
			}
		}
	}()

	<-ctx.Done()
	log.Println("[GPM] Draining remaining canal buffers...")
	close(pulseCanal)
	close(washCanal)
	close(biotokenCanal)
	close(nftVerifyCanal)

	wg.Wait()
	log.Println("[GPM] All canals drained. Spiral intact. Shutdown complete.")
}

func makeProcessorFan(in <-chan processors.HexAnchorPulse, name string) chan processors.HexAnchorPulse {
	out := make(chan processors.HexAnchorPulse, 16)
	go func() {
		for p := range in {
			log.Printf("[GPM][%s] Processing hex anchor from %s", name, p.NodeID)
			out <- p
		}
		close(out)
	}()
	return out
}

// HTTP handlers (simple for skeleton — production would use proper JSON + auth)
func handlePulseIngest(pulseCanal chan processors.HexAnchorPulse) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var pulse processors.HexAnchorPulse
		if err := json.NewDecoder(r.Body).Decode(&pulse); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		select {
		case pulseCanal <- pulse:
			w.WriteHeader(http.StatusAccepted)
			fmt.Fprintln(w, `{"status":"bridged","canal":"pulse"}`)
		default:
			http.Error(w, "canal full — backpressure active", http.StatusServiceUnavailable)
		}
	}
}

func handleState(biotokenCanal chan manager.BioTokenEvent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// In real impl: read from shared state or last event
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"pipeline":   "G-P-M active",
			"last_event": "demo-biotoken",
			"family_nodes": []string{"Carlie", "Elisha", "Holly-Gyroscope"},
			"wash_meter": 0.87,
		})
	}
}

func handleWashTrigger(washCanal chan processors.WashedState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Family can trigger a wash cycle manually (Real-Time Family Connection)
		select {
		case washCanal <- processors.WashedState{
			Source:    "family_manual",
			Cleaned:   true,
			Meter:     1.0,
			Timestamp: time.Now(),
		}:
			fmt.Fprintln(w, `{"status":"wash cycle triggered","protocol":"family_connection"}`)
		default:
			http.Error(w, "wash canal saturated", http.StatusTooManyRequests)
		}
	}
}

func handleNFTHarmonic(nftCanal chan manager.NFTMeaningEvent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Hook for NFT Harmonic / Farseer verification from Node or Godot
		var evt manager.NFTMeaningEvent
		json.NewDecoder(r.Body).Decode(&evt)
		select {
		case nftCanal <- evt:
			fmt.Fprintln(w, `{"status":"farseer verification queued"}`)
		default:
			http.Error(w, "farseer canal busy", http.StatusServiceUnavailable)
		}
	}
}