# BUGFIX REPORT — WASM ↔ WebSocket Communication Audit

Audit date: 2026-03-18
Scope: `pkg/wasmui/rpc_client_wasm.go`, `pkg/server/websocket.go`, `pkg/server/handlers.go`

---

## Bug 1 — WebSocket error responses always return code -32000

| | |
|---|---|
| **Found in** | `pkg/server/websocket.go` — `NewErrorResponse()` |
| **Root cause** | `NewErrorResponse` hardcoded error code `-32000` for every error, regardless of type. When `handleMethod` returned a `*JSONRPCError` (e.g. `-32601` for method-not-found, `-32602` for invalid params), the specific code was discarded. The HTTP path (`writeJSONRPCError`) already performed a type assertion to extract the real code; the WebSocket path did not. |
| **Impact** | WASM clients receiving error responses over WebSocket always saw code `-32000`, making it impossible to distinguish method-not-found from invalid-params or parse errors. |
| **Fix applied** | Added a `*JSONRPCError` type assertion in `NewErrorResponse` to preserve `Code`, `Message`, and `Data` fields. Plain `error` values still fall back to `-32000`. |
| **Regression test** | `pkg/server/websocket_comms_regression_test.go` — `TestNewErrorResponse_PreservesJSONRPCErrorCode`, `TestNewErrorResponse_JSONRoundTrip`, `TestNewErrorResponse_JSONRPCErrorWithData` |

---

## Bug 2 — Data race on `RPCClient.sessionID`

| | |
|---|---|
| **Found in** | `pkg/wasmui/rpc_client_wasm.go` |
| **Root cause** | A package-level `sessionIDMu sync.RWMutex` existed and was used in `drainPendingRequests`, but `captureSessionID`, `Call`, `GetSessionID`, `SetSessionID`, and `JoinGame` all accessed `c.sessionID` without holding the lock. This created a data race between the `handleMessage` goroutine (which calls `captureSessionID` on the server's session confirmation) and the calling goroutine (which reads `sessionID` in `Call`). |
| **Impact** | In native Go test/race-detector builds, this is a flagged data race. In WASM (single-threaded goroutine scheduling) the race rarely manifests, but it can cause the first RPC call after connect to omit `session_id` from params if the session confirmation hasn't been processed yet. The server's `enrichRequestParams` mitigates the immediate failure, but the inconsistency is a latent correctness bug. |
| **Fix applied** | All reads and writes to `c.sessionID` now use `sessionIDMu` — `RLock` for reads in `GetSessionID` and `Call`, `Lock` for writes in `SetSessionID`, `captureSessionID`, `JoinGame`, and `drainPendingRequests`. |
| **Regression test** | Existing native tests exercise these paths; the consistent locking prevents `-race` detector failures. |

---

## Bug 3 — `handleJoinGame` response missing `player_id`

| | |
|---|---|
| **Found in** | `pkg/server/handlers.go` — `handleJoinGame()` |
| **Root cause** | Both response paths (WebSocket session-attach and HTTP new-session) returned only `success` and `session_id`, omitting `player_id`. The WASM client's `JoinGameResult` struct expects a `player_id` field. |
| **Impact** | `JoinGameResult.PlayerID` was always empty on the client, preventing any client-side logic that needs to reference the player by ID. |
| **Fix applied** | Both response paths now include `"player_id": creationResult.PlayerData.GetID()`. |
| **Regression test** | `pkg/server/websocket_comms_regression_test.go` — `TestHandleJoinGame_ReturnsPlayerID` (tests both WebSocket and HTTP code paths). |

---

## Verification

| Check | Result |
|---|---|
| `go test ./pkg/server/...` | ✅ PASS |
| `go test ./pkg/wasmui/...` | ✅ PASS |
| `go test ./test/e2e/... -timeout 5m` | ✅ PASS |
| `GOOS=js GOARCH=wasm go build ./cmd/wasm-ui` | ✅ OK |
| `go build ./pkg/server/...` | ✅ OK |
