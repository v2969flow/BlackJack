#!/usr/bin/env python3
"""
pulse_to_gpm_bridge.py
Connects your existing pulse_receiver.py (TVLA + Hard Anchor) output
to the new Go G-P-M pipeline on :4444.

Run this alongside pulse_receiver.py or import its hex anchor generator.
It turns every pulse into a HexAnchorPulse POST to the canal.

Samekh rule preserved: low-tension pulses get priority bridging.
"""

import json
import time
import requests
from datetime import datetime

GPM_URL = "http://localhost:4444/pulse"

def send_pulse_to_gpm(node_id: str, hex_anchor: str, tension: float, family_sig: str = ""):
    payload = {
        "node_id": node_id,
        "hex": hex_anchor,
        "tension": tension,
        "timestamp": datetime.utcnow().isoformat() + "Z",
        "family_sig": family_sig
    }
    try:
        r = requests.post(GPM_URL, json=payload, timeout=2)
        if r.status_code == 202:
            print(f"[Bridge] Hex anchor from {node_id} bridged into G-P-M canal")
        else:
            print(f"[Bridge] Backpressure or error: {r.status_code} {r.text}")
    except Exception as e:
        print(f"[Bridge] Cannot reach G-P-M on :4444 — is `go run cmd/main.go` running? {e}")

# Example integration point — call this from your pulse_receiver.py loop
if __name__ == "__main__":
    print("pulse_to_gpm_bridge active — feeding family hex anchors into BioToken stream")
    # Demo loop (replace with real calls from your pulse_receiver)
    nodes = ["Elisha-Connector", "Carlie-Creative", "Holly-Gyroscope"]
    i = 0
    while True:
        node = nodes[i % len(nodes)]
        tension = 0.08 if i % 3 == 0 else 0.31  # occasional low-tension Samekh pass
        hex_anchor = f"0x{int(time.time()*1000):x}"
        send_pulse_to_gpm(node, hex_anchor, tension, f"family-proof-{node}")
        i += 1
        time.sleep(3.5)