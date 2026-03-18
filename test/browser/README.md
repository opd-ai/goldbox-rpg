# Browser Playtest

Headless Chromium playtest that exercises the actual WASM binary served by the
real game server. This catches bugs that only manifest in a real browser: WASM
instantiation, WebSocket origin checks, JS↔WASM bridge errors, Ebitengine
canvas panics, etc.

## Implementation

Uses `github.com/chromedp/chromedp` (Go, Chrome DevTools Protocol) to drive a
headless Chromium instance against the live server.

### What the test exercises

1. Builds the WASM binary and starts the server on a random free port.
2. Launches headless Chromium and navigates to the game URL.
3. Waits for the splash screen to finish and the Ebitengine canvas to appear.
4. Captures browser console messages — `console.error` and uncaught exceptions
   are logged as warnings. The test validates canvas presence and server health
   but does not hard-fail on all console errors since some are benign during
   WASM boot.
5. Polls the server `/health` endpoint to confirm the WebSocket upgrade succeeded.
6. Exercises the core gameplay loop:
   - WebSocket connects and session confirmation is received
   - Main menu is visible
   - Character creation completes
   - Exploration screen is visible
   - At least one move action succeeds
   - Combat round completes
7. On any step timeout or JS error, captures a screenshot and fails with a
   descriptive message.
8. Tears down browser and server cleanly.

### Running

```bash
make test-browser
```

Or directly:

```bash
go test ./test/browser/... -v -timeout 5m
```

### Screenshots

On failure (or success), step screenshots are saved to
`test/browser/screenshots/`.

### Requirements

- Google Chrome or Chromium must be installed
- `make wasm` must succeed (builds `web/static/js/game.wasm`)
- `go build ./cmd/server` must succeed

### Why chromedp?

`chromedp` is a pure Go package that speaks the Chrome DevTools Protocol
directly, works with any installed Chrome/Chromium, and requires no external
binary beyond the browser itself. It is the most widely used Go headless-browser
library.
