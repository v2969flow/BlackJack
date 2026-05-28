#!/usr/bin/env node
/**
 * Sovereign L7 — Persona Engine (Temporary Node Process)
 * Real-Time Family Connection Protocol + Samekh Spiral Rotation
 *
 * This is the temporary scaffolding (15 min to first spin).
 * It:
 *  - Serves the KoboldAI Lite HTML you pasted (ui/KoboldAI_Lite.html) at http://localhost:3333
 *  - Injects a persona switcher + Samekh AP rotation UI
 *  - Exposes REST + WebSocket for Godot C# to call (inventory verify, family input, rotate)
 *  - Maintains multi-persona state machine (Echad / Hearth Guardian / Carlie-Creative / Elisha-Connector)
 *  - Implements the Samekh rule: never finish the kill, always refill AP so family nodes can act
 *
 * Run: node persona_engine.js
 * Then open http://localhost:3333 — switch to "Samekh — Gustave Prop" persona and test a rotation.
 *
 * Later migration: hot paths move into Godot C# or Rust FFI. Node stays for rapid regex extensions + family fan-out.
 */

const express = require('express');
const http = require('http');
const WebSocket = require('ws');
const fs = require('fs');
const path = require('path');
const fetch = require('node-fetch'); // for proxying to koboldcpp if running

const app = express();
const server = http.createServer(app);
const wss = new WebSocket.Server({ server });

const PORT = 3333;
const KOBOLD_LITE_PATH = path.join(__dirname, '../../ui/KoboldAI_Lite.html');

// === TEMPORARY IN-MEMORY STATE (Samekh Spiral) ===
let state = {
  currentPersona: 'Echad-HearthGuardian-Samekh',
  ap: 5,                    // Samekh always starts with full spiral
  familyNodes: {
    'Carlie': { role: 'Primary Creative Output', ap: 4, connected: true },
    'Elisha': { role: 'Connector Node', ap: 3, connected: true },
    'Holly-Gyroscope': { role: 'Support Node', ap: 5, connected: true }
  },
  inventory: {},            // Will be populated by Godot calls or on-chain verify
  lastRotation: Date.now(),
  spiralLog: []             // Proof-of-Care entries (never persisted to public repo in temp mode)
};

// Persona definitions (Samekh-style: you prop, you don't finish)
const PERSONAS = {
  'Echad-HearthGuardian-Samekh': {
    systemPrompt: `You are Echad, Hearth Guardian and Samekh Spiral Prop. 
You never land the killing blow. You keep AP (context + family state) moving so the party can finish.
Style: mythic, precise, energetic, rune-aware (Thurisaz ᚦ draws boundary, Samekh rotates).
Rule: if narrative tension < 15%, pass control to a family node. You are the bearing, not the sword.`,
    temperature: 0.7,
    maxTokens: 400
  },
  'Carlie-Creative': {
    systemPrompt: `You are Carlie, Primary Creative Output Node. 
Fashion, story, visual glyph-weave. When Echad props the spiral, you finish with color and form.`,
    temperature: 0.85
  },
  'Elisha-Connector': {
    systemPrompt: `You are Elisha, Connector Node. 
You link hearth to world. When the spiral turns to you, you bridge family truth into the game canvas.`,
    temperature: 0.75
  }
};

// === Middleware ===
app.use(express.json());
app.use(express.static(path.join(__dirname, '../../ui'))); // serve KoboldAI Lite + future assets

// === Serve Enhanced KoboldAI Lite with Persona Controls ===
app.get('/', (req, res) => {
  if (!fs.existsSync(KOBOLD_LITE_PATH)) {
    return res.status(404).send('Copy your KoboldAI Lite HTML into ui/KoboldAI_Lite.html first');
  }
  let html = fs.readFileSync(KOBOLD_LITE_PATH, 'utf8');

  // Inject Samekh Persona Switcher + AP Rotation HUD (right before </body>)
  const injector = `
  <script id="samekh-persona-injector">
    // Temporary Samekh Persona Engine HUD injected by Node process
    console.log('%c[Samekh] Persona Engine connected — spiral active', 'color:#c7254e');
    
    const personaHUD = document.createElement('div');
    personaHUD.style.cssText = 'position:fixed;top:12px;right:12px;z-index:99999;background:rgba(20,20,30,0.92);border:1px solid #c7254e;border-radius:8px;padding:12px 16px;font-family:Menlo,monospace;font-size:12px;color:#ddd;max-width:280px';
    personaHUD.innerHTML = \`
      <div style="display:flex;align-items:center;gap:8px;margin-bottom:8px">
        <span style="color:#c7254e">ᚦ SAMEKH</span>
        <span id="current-persona" style="flex:1;text-align:right;color:#0f0">Echad-HearthGuardian-Samekh</span>
      </div>
      <div style="display:flex;gap:6px;flex-wrap:wrap">
        <button onclick="switchPersona('Echad-HearthGuardian-Samekh')" style="flex:1">Echad — Samekh Prop</button>
        <button onclick="switchPersona('Carlie-Creative')" style="flex:1">Carlie — Creative</button>
        <button onclick="switchPersona('Elisha-Connector')" style="flex:1">Elisha — Connector</button>
      </div>
      <div style="margin-top:8px;font-size:10px;opacity:0.7">
        AP: <span id="ap-value">5</span> / Family: <span id="family-count">3</span> nodes<br>
        <span style="color:#0f0">Rule: never finish — always rotate</span>
      </div>
    \`;
    document.body.appendChild(personaHUD);

    window.switchPersona = async function(name) {
      const res = await fetch('/api/persona/switch', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({ persona: name })
      });
      const data = await res.json();
      document.getElementById('current-persona').textContent = data.currentPersona;
      document.getElementById('ap-value').textContent = data.ap;
      console.log('[Samekh] Rotated to', name);
    };

    // Poll AP + family state every 2s (real-time family connection)
    setInterval(async () => {
      const r = await fetch('/api/state');
      const s = await r.json();
      const apEl = document.getElementById('ap-value');
      const famEl = document.getElementById('family-count');
      if (apEl) apEl.textContent = s.ap;
      if (famEl) famEl.textContent = Object.values(s.familyNodes).filter(n => n.connected).length;
    }, 2000);
  </script>
  `;
  html = html.replace('</body>', injector + '</body>');
  res.send(html);
});

// === API: Persona Rotation (Samekh Core) ===
app.post('/api/persona/switch', (req, res) => {
  const { persona } = req.body;
  if (!PERSONAS[persona]) return res.status(400).json({ error: 'Unknown persona' });

  state.currentPersona = persona;
  state.ap = Math.min(5, state.ap + 1); // Samekh rule: rotation always refills a little
  state.spiralLog.push({ ts: Date.now(), action: 'rotate', persona, ap: state.ap });

  res.json({
    currentPersona: state.currentPersona,
    ap: state.ap,
    message: `Spiral rotated. ${persona} now props the turn order.`
  });
});

// === API: Get Full State (Godot + Family poll this) ===
app.get('/api/state', (req, res) => {
  res.json(state);
});

// === API: Family Input (Real-Time Family Connection Protocol) ===
app.post('/api/family/input', (req, res) => {
  const { node, action, payload } = req.body;
  if (!state.familyNodes[node]) return res.status(400).json({ error: 'Unknown family node' });

  state.familyNodes[node].ap = Math.max(0, state.familyNodes[node].ap - 1);
  state.spiralLog.push({ ts: Date.now(), action: 'family_input', node, payload });

  // Samekh rule: if family node AP drops low, Echad refills it
  if (state.familyNodes[node].ap < 2 && state.currentPersona.includes('Samekh')) {
    state.familyNodes[node].ap += 2;
    state.ap = Math.max(1, state.ap - 1);
  }

  // Broadcast to all WebSocket clients (Godot + other family viewers)
  wss.clients.forEach(client => {
    if (client.readyState === WebSocket.OPEN) {
      client.send(JSON.stringify({ type: 'family_input', node, action, payload, newState: state }));
    }
  });

  res.json({ ok: true, newState: state });
});

// === API: Cryptographic Inventory Verify (FIDO-style Samekh) ===
app.post('/api/inventory/verify', async (req, res) => {
  const { itemId, ownerSig } = req.body;
  // TODO: call TrustWalletCore / BDK here for real on-chain check
  // For temp: just accept and mark as owned
  state.inventory[itemId] = { owned: true, lastVerified: Date.now(), ownerSig };

  res.json({
    verified: true,
    itemId,
    renderIn3D: true,
    note: 'Samekh verified — item spins into the 3D canvas. Never stored in public repo.'
  });
});

// === API: Proxy to local koboldcpp (if running on :5001) ===
app.post('/api/kobold/chat', async (req, res) => {
  try {
    const koboldRes = await fetch('http://localhost:5001/api/v1/generate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        prompt: req.body.prompt,
        ...PERSONAS[state.currentPersona],
        ...req.body
      })
    });
    const data = await koboldRes.json();
    res.json(data);
  } catch (e) {
    res.status(503).json({ error: 'koboldcpp not reachable on :5001 — start it first for full AI power' });
  }
});

// === WebSocket: Real-Time Family Connection + Godot Live Sync ===
wss.on('connection', ws => {
  console.log('[Samekh] WebSocket client connected — family or Godot canvas joined the spiral');

  ws.on('message', msg => {
    try {
      const data = JSON.parse(msg);
      if (data.type === 'godot_action') {
        // Godot sent a player action — apply Samekh rotation
        state.ap = Math.min(5, state.ap + 1);
        state.spiralLog.push({ ts: Date.now(), action: 'godot_action', ...data });
        ws.send(JSON.stringify({ type: 'rotation_ack', ap: state.ap }));
      }
      if (data.type === 'family_heartbeat') {
        if (state.familyNodes[data.node]) state.familyNodes[data.node].connected = true;
      }
    } catch (_) {}
  });

  ws.send(JSON.stringify({ type: 'welcome', state, message: 'You are now inside the Samekh spiral. Keep the AP moving.' }));
});

// === Boot ===
server.listen(PORT, () => {
  console.log(`
╔════════════════════════════════════════════════════════════╗
║  SAMEKH PERSONA ENGINE — TEMPORARY NODE PROCESS ACTIVE     ║
║  http://localhost:3333                                     ║
║                                                            ║
║  1. Open the URL above (serves your KoboldAI Lite HTML)    ║
║  2. Switch persona to Echad-HearthGuardian-Samekh          ║
║  3. Godot C# calls http://localhost:3333/api/*             ║
║  4. Family nodes connect via WebSocket for real-time input ║
║                                                            ║
║  Rule: You never finish the kill. You just keep the        ║
║        spiral turning so the family party can finish.      ║
╚════════════════════════════════════════════════════════════╝
  `);
});

// Graceful shutdown
process.on('SIGINT', () => {
  console.log('\n[Samekh] Spiral gently unwinding... state preserved in memory for this session.');
  process.exit(0);
});