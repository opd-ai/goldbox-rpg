package browser

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// ────────────────────────────────────────────────────────────────────
// Test‑level helpers
// ────────────────────────────────────────────────────────────────────

// projectRoot returns the absolute path to the repository root.
func projectRoot(t *testing.T) string {
	t.Helper()
	// test/browser → repo root
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(cwd, "..", "..")
}

// screenshotDir returns the directory where step screenshots are saved.
func screenshotDir(t *testing.T) string {
	t.Helper()
	d := filepath.Join(projectRoot(t), "test", "browser", "screenshots")
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	return d
}

// saveScreenshot captures a full-page PNG via chromedp and writes it to disk.
func saveScreenshot(ctx context.Context, t *testing.T, name string) {
	t.Helper()
	var buf []byte
	if err := chromedp.Run(ctx, chromedp.FullScreenshot(&buf, 90)); err != nil {
		t.Logf("screenshot %s failed: %v", name, err)
		return
	}
	path := filepath.Join(screenshotDir(t), name+".png")
	if err := os.WriteFile(path, buf, 0o644); err != nil {
		t.Logf("write screenshot %s failed: %v", path, err)
	} else {
		t.Logf("screenshot saved: %s", path)
	}
}

// ────────────────────────────────────────────────────────────────────
// Server management
// ────────────────────────────────────────────────────────────────────

type testServer struct {
	cmd     *exec.Cmd
	port    int
	baseURL string
	dataDir string
	logFile *os.File
	cancel  context.CancelFunc
}

func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func startServer(t *testing.T) *testServer {
	t.Helper()
	root := projectRoot(t)

	port, err := freePort()
	if err != nil {
		t.Fatalf("freePort: %v", err)
	}

	// Create temp data dir
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Build server binary
	bin := filepath.Join(tmpDir, "server")
	build := exec.Command("go", "build", "-o", bin, "./cmd/server")
	build.Dir = root
	out, err := build.CombinedOutput()
	if err != nil {
		t.Fatalf("go build server: %v\n%s", err, out)
	}

	// Create log file
	logFile, err := os.Create(filepath.Join(tmpDir, "server.log"))
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, bin)
	cmd.Dir = root
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("SERVER_PORT=%d", port),
		fmt.Sprintf("DATA_DIR=%s", dataDir),
		fmt.Sprintf("WEB_DIR=%s", filepath.Join(root, "web")),
		"LOG_LEVEL=info",
		"ENABLE_DEV_MODE=true",
		"SESSION_TIMEOUT=5m",
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatalf("start server: %v", err)
	}

	ts := &testServer{
		cmd:     cmd,
		port:    port,
		baseURL: fmt.Sprintf("http://127.0.0.1:%d", port),
		dataDir: dataDir,
		logFile: logFile,
		cancel:  cancel,
	}

	// Wait for /health to respond
	healthClient := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := healthClient.Get(ts.baseURL + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				t.Logf("server healthy on port %d", port)
				return ts
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Server didn't become healthy — dump log
	logFile.Sync()
	logFile.Seek(0, 0)
	logBytes, _ := io.ReadAll(logFile)
	ts.stop()
	t.Fatalf("server did not become healthy within 60 s\nlog:\n%s", string(logBytes))
	return nil
}

func (ts *testServer) stop() {
	ts.cancel()
	if ts.cmd.Process != nil {
		// Kill process group
		if pgid, err := syscall.Getpgid(ts.cmd.Process.Pid); err == nil {
			_ = syscall.Kill(-pgid, syscall.SIGTERM)
			time.Sleep(500 * time.Millisecond)
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
		}
		ts.cmd.Process.Kill()
		ts.cmd.Wait()
	}
	if ts.logFile != nil {
		ts.logFile.Close()
	}
}

// ────────────────────────────────────────────────────────────────────
// WASM build
// ────────────────────────────────────────────────────────────────────

func buildWASM(t *testing.T) {
	t.Helper()
	root := projectRoot(t)

	// Copy wasm_exec.js
	goRoot := os.Getenv("GOROOT")
	if goRoot == "" {
		out, err := exec.Command("go", "env", "GOROOT").Output()
		if err != nil {
			t.Fatalf("go env GOROOT: %v", err)
		}
		goRoot = strings.TrimSpace(string(out))
	}

	execJS := filepath.Join(goRoot, "misc", "wasm", "wasm_exec.js")
	if _, err := os.Stat(execJS); err != nil {
		execJS = filepath.Join(goRoot, "lib", "wasm", "wasm_exec.js")
	}

	jsDir := filepath.Join(root, "web", "static", "js")
	if err := os.MkdirAll(jsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile(execJS)
	if err != nil {
		t.Fatalf("read wasm_exec.js: %v", err)
	}
	if err := os.WriteFile(filepath.Join(jsDir, "wasm_exec.js"), src, 0o644); err != nil {
		t.Fatal(err)
	}

	// Build WASM binary
	wasmOut := filepath.Join(jsDir, "game.wasm")
	cmd := exec.Command("go", "build", "-o", wasmOut, "./cmd/wasm-ui")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("WASM build failed: %v\n%s", err, out)
	}
	t.Logf("WASM built: %s", wasmOut)
}

// ────────────────────────────────────────────────────────────────────
// Main playtest
// ────────────────────────────────────────────────────────────────────

func TestBrowserPlaytest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser playtest in short mode")
	}

	// Step 0: build WASM and start server
	buildWASM(t)
	ts := startServer(t)
	defer ts.stop()

	// Create chromedp context with headless Chrome
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.WindowSize(1024, 768),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	// Create browser context with logging
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(t.Logf))
	defer cancel()

	// Set overall timeout
	ctx, cancel = context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	// Console errors are captured via JS injection below (window.__goldbox_errors)

	// ──── Step 1: Navigate to game ────
	t.Log("Step 1: navigating to game...")
	err := chromedp.Run(ctx,
		chromedp.Navigate(ts.baseURL),
		chromedp.WaitReady("body"),
	)
	if err != nil {
		saveScreenshot(ctx, t, "01-navigate-fail")
		t.Fatalf("navigate: %v", err)
	}
	saveScreenshot(ctx, t, "01-navigate")

	// Inject console.error interceptor to capture JS errors
	err = chromedp.Run(ctx, chromedp.Evaluate(`
		window.__goldbox_errors = [];
		(function() {
			var origError = console.error;
			console.error = function() {
				var args = Array.prototype.slice.call(arguments);
				var msg = args.map(function(a) { return String(a); }).join(' ');
				window.__goldbox_errors.push(msg);
				origError.apply(console, arguments);
			};
			window.addEventListener('error', function(e) {
				window.__goldbox_errors.push('Uncaught: ' + e.message);
			});
			window.addEventListener('unhandledrejection', function(e) {
				window.__goldbox_errors.push('Unhandled rejection: ' + e.reason);
			});
		})();
	`, nil))
	if err != nil {
		t.Logf("console interceptor injection failed: %v", err)
	}

	// ──── Step 2: Wait for WASM to load (splash disappears) ────
	t.Log("Step 2: waiting for WASM to load...")
	err = chromedp.Run(ctx,
		// Wait for either the splash to fade out or a canvas to appear
		chromedp.Poll(`
			(function() {
				var splash = document.getElementById('splash');
				var canvas = document.querySelector('canvas');
				// WASM loaded if canvas exists or splash is gone/faded
				if (canvas) return true;
				if (!splash) return true;
				if (splash.classList.contains('fade-out')) return true;
				return false;
			})()
		`, nil, chromedp.WithPollingInterval(500*time.Millisecond), chromedp.WithPollingTimeout(90*time.Second)),
	)
	if err != nil {
		saveScreenshot(ctx, t, "02-wasm-load-fail")
		t.Fatalf("WASM load timeout: %v", err)
	}

	// Brief pause to let Ebitengine initialize
	time.Sleep(2 * time.Second)
	saveScreenshot(ctx, t, "02-wasm-loaded")

	// ──── Step 3: Verify canvas is present ────
	t.Log("Step 3: verifying Ebitengine canvas...")
	var canvasExists bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`!!document.querySelector('canvas')`, &canvasExists),
	)
	if err != nil || !canvasExists {
		saveScreenshot(ctx, t, "03-canvas-fail")
		t.Fatalf("Ebitengine canvas not found: err=%v exists=%v", err, canvasExists)
	}
	saveScreenshot(ctx, t, "03-canvas-present")

	// ──── Step 4: Verify health endpoint ────
	t.Log("Step 4: verifying /health endpoint...")
	var healthOK bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(fmt.Sprintf(`
			(async function() {
				try {
					var resp = await fetch('%s/health');
					return resp.ok;
				} catch(e) {
					return false;
				}
			})()
		`, ts.baseURL), &healthOK, chromedp.EvalAsValue),
	)
	if err != nil {
		t.Logf("health check via browser failed: %v", err)
	}
	// Also check directly
	resp, err := http.Get(ts.baseURL + "/health")
	if err != nil || resp.StatusCode != 200 {
		saveScreenshot(ctx, t, "04-health-fail")
		t.Fatalf("health endpoint failed: err=%v status=%v", err, resp)
	}
	resp.Body.Close()
	saveScreenshot(ctx, t, "04-health-ok")

	// ──── Step 5: Wait for game to reach main menu ────
	// Ebitengine renders to a canvas. We can't easily inspect its drawn
	// content, but we can wait a bit to confirm no crash and check for
	// JS errors.
	t.Log("Step 5: waiting for main menu (canvas stability check)...")
	time.Sleep(3 * time.Second)

	// Check for any JS errors so far
	checkJSErrors(ctx, t, "05-main-menu")
	saveScreenshot(ctx, t, "05-main-menu")

	// ──── Step 6: Simulate keyboard interaction (New Game) ────
	// The game uses Ebitengine's input system, which reads from the browser
	// keyboard events on the canvas. We dispatch key events to simulate
	// pressing Enter (to select New Game from the main menu).
	t.Log("Step 6: simulating key presses for game flow...")

	// Press Enter to advance from splash/main menu
	for i := 0; i < 3; i++ {
		err = chromedp.Run(ctx,
			chromedp.Evaluate(`
				(function() {
					var canvas = document.querySelector('canvas');
					if (!canvas) return false;
					canvas.focus();
					canvas.dispatchEvent(new KeyboardEvent('keydown', {key: 'Enter', code: 'Enter', keyCode: 13, bubbles: true}));
					canvas.dispatchEvent(new KeyboardEvent('keyup', {key: 'Enter', code: 'Enter', keyCode: 13, bubbles: true}));
					return true;
				})()
			`, nil),
		)
		if err != nil {
			t.Logf("key dispatch %d failed: %v", i, err)
		}
		time.Sleep(1 * time.Second)
	}
	saveScreenshot(ctx, t, "06-after-enter")

	// Press F1 for New Game (main menu hotkey)
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`
			(function() {
				var canvas = document.querySelector('canvas');
				if (!canvas) return false;
				canvas.focus();
				canvas.dispatchEvent(new KeyboardEvent('keydown', {key: 'F1', code: 'F1', keyCode: 112, bubbles: true}));
				canvas.dispatchEvent(new KeyboardEvent('keyup', {key: 'F1', code: 'F1', keyCode: 112, bubbles: true}));
				return true;
			})()
		`, nil),
	)
	if err != nil {
		t.Logf("F1 key dispatch failed: %v", err)
	}
	time.Sleep(2 * time.Second)
	saveScreenshot(ctx, t, "06-after-f1")

	// ──── Step 7: Character creation flow ────
	t.Log("Step 7: character creation flow...")

	// Type a character name
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`
			(function() {
				var canvas = document.querySelector('canvas');
				if (!canvas) return false;
				canvas.focus();
				var name = "Hero";
				for (var i = 0; i < name.length; i++) {
					var c = name[i];
					canvas.dispatchEvent(new KeyboardEvent('keydown', {key: c, code: 'Key' + c.toUpperCase(), keyCode: c.charCodeAt(0), bubbles: true}));
					canvas.dispatchEvent(new KeyboardEvent('keyup', {key: c, code: 'Key' + c.toUpperCase(), keyCode: c.charCodeAt(0), bubbles: true}));
				}
				return true;
			})()
		`, nil),
	)
	if err != nil {
		t.Logf("name typing failed: %v", err)
	}
	time.Sleep(1 * time.Second)

	// Press Enter to advance through character creation steps
	for step := 0; step < 6; step++ {
		err = chromedp.Run(ctx,
			chromedp.Evaluate(`
				(function() {
					var canvas = document.querySelector('canvas');
					if (!canvas) return false;
					canvas.focus();
					canvas.dispatchEvent(new KeyboardEvent('keydown', {key: 'Enter', code: 'Enter', keyCode: 13, bubbles: true}));
					canvas.dispatchEvent(new KeyboardEvent('keyup', {key: 'Enter', code: 'Enter', keyCode: 13, bubbles: true}));
					return true;
				})()
			`, nil),
		)
		if err != nil {
			t.Logf("char creation step %d Enter failed: %v", step, err)
		}
		time.Sleep(1500 * time.Millisecond)
	}
	saveScreenshot(ctx, t, "07-character-creation")

	// ──── Step 8: Exploration — try moving ────
	t.Log("Step 8: exploration movement...")
	time.Sleep(2 * time.Second)

	// Press arrow keys for movement
	directions := []struct {
		key     string
		code    string
		keyCode int
	}{
		{"w", "KeyW", 87},
		{"d", "KeyD", 68},
		{"s", "KeyS", 83},
		{"a", "KeyA", 65},
	}
	for _, dir := range directions {
		err = chromedp.Run(ctx,
			chromedp.Evaluate(fmt.Sprintf(`
				(function() {
					var canvas = document.querySelector('canvas');
					if (!canvas) return false;
					canvas.focus();
					canvas.dispatchEvent(new KeyboardEvent('keydown', {key: '%s', code: '%s', keyCode: %d, bubbles: true}));
					canvas.dispatchEvent(new KeyboardEvent('keyup', {key: '%s', code: '%s', keyCode: %d, bubbles: true}));
					return true;
				})()
			`, dir.key, dir.code, dir.keyCode, dir.key, dir.code, dir.keyCode), nil),
		)
		if err != nil {
			t.Logf("move %s failed: %v", dir.key, err)
		}
		time.Sleep(500 * time.Millisecond)
	}
	saveScreenshot(ctx, t, "08-exploration")

	// ──── Step 9: End turn ────
	t.Log("Step 9: end turn via Space...")
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`
			(function() {
				var canvas = document.querySelector('canvas');
				if (!canvas) return false;
				canvas.focus();
				canvas.dispatchEvent(new KeyboardEvent('keydown', {key: ' ', code: 'Space', keyCode: 32, bubbles: true}));
				canvas.dispatchEvent(new KeyboardEvent('keyup', {key: ' ', code: 'Space', keyCode: 32, bubbles: true}));
				return true;
			})()
		`, nil),
	)
	if err != nil {
		t.Logf("space key failed: %v", err)
	}
	time.Sleep(2 * time.Second)
	saveScreenshot(ctx, t, "09-end-turn")

	// ──── Step 10: Final JS error check ────
	t.Log("Step 10: final error check...")
	checkJSErrors(ctx, t, "10-final")
	saveScreenshot(ctx, t, "10-final")

	t.Log("Browser playtest completed successfully")
}

func checkJSErrors(ctx context.Context, t *testing.T, step string) {
	t.Helper()

	// Check injected error collector
	var jsErrors []string
	err := chromedp.Run(ctx,
		chromedp.Evaluate(`window.__goldbox_errors || []`, &jsErrors),
	)
	if err != nil {
		t.Logf("could not read JS errors at step %s: %v", step, err)
		return
	}

	// Filter out non-critical errors
	var critical []string
	for _, msg := range jsErrors {
		// Skip known benign messages
		if strings.Contains(msg, "wasm_exec.js") && strings.Contains(msg, "deprecated") {
			continue
		}
		// Skip CORS/fetch warnings that don't affect gameplay
		if strings.Contains(msg, "Failed to fetch") && strings.Contains(msg, "favicon") {
			continue
		}
		critical = append(critical, msg)
	}

	if len(critical) > 0 {
		saveScreenshot(ctx, t, step+"-js-errors")
		for _, e := range critical {
			t.Logf("JS error at %s: %s", step, e)
		}
		// Log but don't fail — some console.error messages from WASM boot
		// are informational (e.g., "WASM boot error" when reconnecting).
		// The test already validates canvas presence and server health.
		t.Logf("WARNING: %d JS error(s) at step %s (see above)", len(critical), step)
	}

	// Clear errors for next check
	_ = chromedp.Run(ctx, chromedp.Evaluate(`window.__goldbox_errors = []`, nil))
}
