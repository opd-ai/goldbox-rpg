# Gold Box RPG — Static WASM Demo

This directory contains a **static, serverless demo** of the Gold Box RPG
Ebitengine client. It runs entirely in the browser via WebAssembly — no
backend server is required.

## What You See

The demo loads the full game UI (viewport grid, character panel, combat log,
action buttons) in **offline mode**. Because there is no JSON-RPC server, the
WebSocket connection will gracefully fail and the status indicator will show
**"Disconnected"**. The Ebitengine rendering loop runs normally, so you can
see the game interface and interact with the UI controls.

## How It Works

| File | Purpose |
|---|---|
| `index.html` | Splash screen that loads the WASM binary |
| `game.wasm` | Ebitengine client compiled from `cmd/wasm-demo/` (generated at build time) |
| `wasm_exec.js` | Go WASM glue copied from `$(go env GOROOT)/misc/wasm/` (generated at build time) |
| `assets/` | Placeholder sprites copied from `web/static/assets/` (generated at build time) |

The entry point (`cmd/wasm-demo/main.go`) calls `wasmui.NewGame()` — the same
initializer as the full client — and lets Ebitengine take over. The RPC
client's `connectAndJoin()` goroutine runs in the background, fails to reach
a server, and the game continues in disconnected mode.

## Building Locally

```bash
# Build the WASM binary
GOOS=js GOARCH=wasm go build -o docs/game.wasm ./cmd/wasm-demo

# Copy the Go WASM glue file
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" docs/

# Copy placeholder assets (remove first to avoid nested dirs on repeat runs)
rm -rf docs/assets
cp -r web/static/assets docs/assets

# Serve locally (any static file server works)
cd docs && python3 -m http.server 8080
# Then open http://localhost:8080
```

## Deployment

The GitHub Actions workflow (`.github/workflows/pages.yml`) automatically
builds and deploys this directory to GitHub Pages on every push to `main`.

## Full Experience

For the complete server-backed experience with real-time multiplayer, combat,
and character progression, see the main project:
**[Gold Box RPG Engine](https://github.com/opd-ai/goldbox-rpg)**
