package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Pulse mirrors the JSON record from pulse_receiver.py
type Pulse struct {
	Seq            int                    `json:"seq"`
	OriginNode     string                 `json:"origin_node"`
	HexAnchorState string                 `json:"hex_anchor_state"`
	Pulse          map[string]interface{} `json:"pulse"`
	BioTokenReady  bool                   `json:"bio_token_ready"`
	Timestamp      float64                `json:"timestamp"`
}

// Processed result from any processor
type Processed struct {
	Anchor   string
	PType    string
	Result   string
	Verified bool
	Details  map[string]interface{}
}

// nftHarmonicFarseerProcessor verifies NFT meaning via harmonic + farseer layer
// "Farseer" = predictive/lore-aligned semantic resonance check (mocked as trait + keyword match)
func nftHarmonicFarseerProcessor(p Pulse, out chan<- Processed, wg *sync.WaitGroup) {
	defer wg.Done()

	traitsIface, ok := p.Pulse["traits"]
	if !ok {
		return
	}
	traits, ok := traitsIface.([]interface{})
	if !ok {
		return
	}

	hasHarmonic := false
	hasFarseer := false
	hasLore := false
	for _, t := range traits {
		if s, ok := t.(string); ok {
			ls := strings.ToLower(s)
			if ls == "harmonic" {
				hasHarmonic = true
			}
			if ls == "farseer" {
				hasFarseer = true
			}
			if strings.Contains(ls, "lore") || strings.Contains(ls, "hearth") || strings.Contains(ls, "care") {
				hasLore = true
			}
		}
	}

	verified := hasHarmonic && hasFarseer && hasLore
	meaning := "NFT traits lack full harmonic/farseer/lore resonance"
	if verified {
		meaning = "✅ NFT Harmonic + Farseer layer VERIFIED: meaning aligned to Aetheria / Care Weave / Hearth Guardian mythos"
	}

	details := map[string]interface{}{
		"has_harmonic": hasHarmonic,
		"has_farseer":  hasFarseer,
		"has_lore":     hasLore,
		"nft_id":       p.Pulse["nft_id"],
	}

	out <- Processed{
		Anchor:   p.HexAnchorState,
		PType:    "nft_harmonic_farseer",
		Result:   meaning,
		Verified: verified,
		Details:  details,
	}
}

// bioTokenProcessor mints/anchors BioMe identity concepts into the stream
func bioTokenProcessor(p Pulse, out chan<- Processed, wg *sync.WaitGroup) {
	defer wg.Done()
	if !p.BioTokenReady {
		return
	}

	ptype := ""
	if pt, ok := p.Pulse["type"].(string); ok {
		ptype = pt
	}

	result := "BioMe identity pulse received but not ready for tokenization"
	verified := false

	if ptype == "bio_identity" {
		careAction := ""
		if ca, ok := p.Pulse["care_action"].(string); ok {
			careAction = ca
		}
		result = fmt.Sprintf("🧬 BioMe Token ANCHORED :: Proof-of-Care committed | Action: %s | Biome hash ready for DID/TEE soulbound", careAction)
		verified = true
	} else if ptype == "vqe_quantum_rl" {
		result = "🧬 BioMe Token ANCHORED via VQE-RL pulse :: Quantum-optimized care/gaming state folded into identity lattice"
		verified = true
	}

	out <- Processed{
		Anchor:   p.HexAnchorState,
		PType:    "bio_token_biome",
		Result:   result,
		Verified: verified,
		Details:  map[string]interface{}{"origin": p.OriginNode},
	}
}

// vqeGamingProcessor handles the gaming architecture side (BG3 Honor Mode example)
func vqeGamingProcessor(p Pulse, out chan<- Processed, wg *sync.WaitGroup) {
	defer wg.Done()

	ptype := ""
	if pt, ok := p.Pulse["type"].(string); ok {
		ptype = pt
	}
	if ptype != "vqe_quantum_rl" {
		return
	}

	energy := 0.0
	if e, ok := p.Pulse["energy"].(float64); ok {
		energy = e
	}
	gameCtx := ""
	if gc, ok := p.Pulse["game_context"].(map[string]interface{}); ok {
		if g, ok := gc["game"].(string); ok {
			gameCtx = g
		}
	}

	result := fmt.Sprintf("⚡ VQE+QuantumRL GAMING :: Energy=%.6f | Context=%s | Policy updated for persistent Honor Mode / dragon energy / ded burst tactics", energy, gameCtx)
	out <- Processed{
		Anchor:   p.HexAnchorState,
		PType:    "vqe_quantum_rl_gaming",
		Result:   result,
		Verified: true,
		Details:  map[string]interface{}{"energy": energy},
	}
}

// canalBridgePostProcessor applies additional Go-side filtering / verification if needed
func canalBridgePostProcessor(p Pulse, out chan<- Processed, wg *sync.WaitGroup) {
	defer wg.Done()

	mode := "canal"
	if m, ok := p.Pulse["_filter_mode"].(string); ok {
		mode = m
	}

	status := "UNKNOWN"
	if s, ok := p.Pulse["_canal_status"].(string); ok {
		status = s
	} else if s, ok := p.Pulse["_bridge_status"].(string); ok {
		status = s
	}

	result := fmt.Sprintf("🌉 Canal/Bridge Filter :: mode=%s | status=%s | anchor integrity preserved for Syncroni City IoC", mode, status)
	out <- Processed{
		Anchor:   p.HexAnchorState,
		PType:    "canal_bridge_filter",
		Result:   result,
		Verified: true,
		Details:  map[string]interface{}{"mode": mode, "status": status},
	}
}

// manager orchestrates the G-P-M pipeline using goroutines + channels
func manager(input <-chan Pulse, done chan<- bool) {
	var wg sync.WaitGroup

	nftOut := make(chan Processed, 32)
	bioOut := make(chan Processed, 32)
	vqeOut := make(chan Processed, 32)
	filterOut := make(chan Processed, 32)

	// Fan-out to processors for each incoming pulse
	pulseCount := 0
	for p := range input {
		pulseCount++
		wg.Add(4)
		go nftHarmonicFarseerProcessor(p, nftOut, &wg)
		go bioTokenProcessor(p, bioOut, &wg)
		go vqeGamingProcessor(p, vqeOut, &wg)
		go canalBridgePostProcessor(p, filterOut, &wg)
	}

	// Close outs when all processors done
	go func() {
		wg.Wait()
		close(nftOut)
		close(bioOut)
		close(vqeOut)
		close(filterOut)
	}()

	// Collect and print results (in real system: route to Care Treasury, NFT mint, DID registry, etc.)
	fmt.Println("\n// === G-P-M PIPELINE RESULTS (Trappist TEE9 Array) ===")
	fmt.Printf("// Processed %d pulses through Goroutines/Processors/Manager\n\n", pulseCount)

	collect := func(name string, ch <-chan Processed) {
		for res := range ch {
			verifiedIcon := "❌"
			if res.Verified {
				verifiedIcon = "✅"
			}
			fmt.Printf("[%s] %s | %s | anchor=%s...\n", name, verifiedIcon, res.Result, res.Anchor[:12])
			if len(res.Details) > 0 {
				dj, _ := json.Marshal(res.Details)
				fmt.Printf("         details: %s\n", string(dj))
			}
		}
	}

	var collectWg sync.WaitGroup
	collectWg.Add(4)
	go func() { defer collectWg.Done(); collect("NFT-HARMONIC", nftOut) }()
	go func() { defer collectWg.Done(); collect("BIO-TOKEN", bioOut) }()
	go func() { defer collectWg.Done(); collect("VQE-RL", vqeOut) }()
	go func() { defer collectWg.Done(); collect("CANAL-BRIDGE", filterOut) }()

	collectWg.Wait()
	fmt.Println("\n// Pipeline complete. Hex anchors ready for BioToken stream / NFT portal indexing / Care Treasury commit.")
	fmt.Println("// Next: integrate with KENNA orchestrator, IOTA DePIN, or Cosmos agent blocks.")

	done <- true
}

// generateMockPulses creates internal equivalent of pulse_receiver output for standalone demo
func generateMockPulses(count int) chan Pulse {
	out := make(chan Pulse, count)
	go func() {
		defer close(out)
		mockData := []struct {
			ptype       string
			extra       interface{}
			bioReady    bool
			origin      string
		}{
			{"vqe_quantum_rl", map[string]interface{}{"game": "Baldurs_Gate_3_Honor", "event": "dragon_energy_surge"}, true, "KORE"},
			{"bio_identity", "care_provision_daughter_fashion_support", true, "KENNA"},
			{"nft_harmonic", map[string]interface{}{"nft_id": "0xAETHERIA00A1", "traits": []interface{}{"harmonic", "farseer", "lore_aligned", "hearth_guardian"}}, false, "SAGE"},
			{"vqe_quantum_rl", map[string]interface{}{"game": "Baldurs_Gate_3_Honor", "event": "dedicated_burst_tactics"}, true, "VERA"},
			{"bio_identity", "family_hearth_stabilization", true, "CORA"},
			{"nft_harmonic", map[string]interface{}{"nft_id": "0xAETHERIA00A2", "traits": []interface{}{"harmonic", "void_gothagal", "care_weave"}}, false, "FEMINI"},
		}
		for i := 0; i < count && i < len(mockData); i++ {
			md := mockData[i]
			pulseMap := map[string]interface{}{"type": md.ptype}
			if md.ptype == "vqe_quantum_rl" {
				pulseMap["energy"] = 0.187 + float64(i)*0.03
				pulseMap["game_context"] = md.extra
			} else if md.ptype == "bio_identity" {
				pulseMap["care_action"] = md.extra
				pulseMap["intensity"] = 0.8 + float64(i%3)*0.05
			} else if md.ptype == "nft_harmonic" {
				pulseMap = md.extra.(map[string]interface{})
				pulseMap["type"] = md.ptype
			}
			// simulate filter fields
			pulseMap["_filter_mode"] = "canal"
			if md.ptype == "vqe_quantum_rl" {
				pulseMap["_canal_status"] = "PASSED_STABLE"
			} else {
				pulseMap["_canal_status"] = "GENERIC_CANAL_PASS"
			}

			// compute a fake anchor (in real would come from py)
			anchorSeed := fmt.Sprintf("%d-%s-%v", i, md.ptype, pulseMap)
			anchor := fmt.Sprintf("%x", []byte(anchorSeed))[:64] // fake but fixed length-ish

			out <- Pulse{
				Seq:            i,
				OriginNode:     md.origin,
				HexAnchorState: anchor,
				Pulse:          pulseMap,
				BioTokenReady:  md.bioReady,
				Timestamp:      float64(time.Now().Unix()),
			}
			time.Sleep(120 * time.Millisecond)
		}
	}()
	return out
}

func main() {
	live := flag.Bool("live", false, "Attempt to exec python pulse_receiver.py and pipe output (requires pulse_receiver.py in same dir)")
	stdin := flag.Bool("stdin", false, "Read JSON pulse records from stdin instead of generating mocks")
	count := flag.Int("count", 6, "Number of pulses for mock mode")
	flag.Parse()

	fmt.Println("// G-P-M PIPELINE :: Goroutines / Processors / Manager")
	fmt.Println("// PRO_GARDEN Trappist TEE9 Array | KENNA Interface Orchestrator")
	fmt.Println("// VQE+QuantumRL gaming | NFT harmonic/farseer verification | BioMe identity | canal+bridge filter")
	fmt.Println("// Hex anchors -> BioToken stream ready")

	var input chan Pulse
	done := make(chan bool, 1)

	if *stdin {
		fmt.Println("// Mode: STDIN (connect via: python3 pulse_receiver.py | ./gpm_pipeline --stdin)")
		input = make(chan Pulse, *count)
		go func() {
			defer close(input)
			scanner := bufio.NewScanner(os.Stdin)
			seq := 0
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					continue
				}
				var p Pulse
				if err := json.Unmarshal([]byte(line), &p); err != nil {
					fmt.Fprintf(os.Stderr, "// stdin parse error: %v\n", err)
					continue
				}
				p.Seq = seq
				input <- p
				seq++
			}
		}()
	} else if *live {
		fmt.Println("// Mode: LIVE pipe from pulse_receiver.py")
		cmd := exec.Command("python3", "./pulse_receiver.py", "--count", fmt.Sprintf("%d", *count), "--filter-mode", "canal", "--node", "KENNA")
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			fmt.Printf("// exec error, falling back to mock: %v\n", err)
			input = generateMockPulses(*count)
		} else {
			if err := cmd.Start(); err != nil {
				fmt.Printf("// python start error, fallback to mock: %v\n", err)
				input = generateMockPulses(*count)
			} else {
				input = make(chan Pulse, *count)
				go func() {
					defer close(input)
					defer cmd.Wait()
					scanner := bufio.NewScanner(stdout)
					seq := 0
					for scanner.Scan() {
						line := strings.TrimSpace(scanner.Text())
						if line == "" || strings.HasPrefix(line, "//") {
							continue
						}
						var p Pulse
						if err := json.Unmarshal([]byte(line), &p); err != nil {
							continue
						}
						p.Seq = seq
						input <- p
						seq++
					}
				}()
			}
		}
	} else {
		fmt.Println("// Mode: MOCK (standalone demo). Use --live or pipe for real pulse_receiver.py integration")
		input = generateMockPulses(*count)
	}

	go manager(input, done)
	<-done

	fmt.Println("\n// SHRED complete. All hex anchors processed into BioToken / NFT verified streams.")
	fmt.Println("// Ready for next layer: IOTA DePIN micro-tx, Cosmos SDK agent blocks, or Syncroni City IoC commit.")
}
