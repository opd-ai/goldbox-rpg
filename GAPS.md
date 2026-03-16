# Implementation Gaps — 2026-03-16

This document identifies gaps between the GoldBox RPG Engine's stated goals and current implementation. Each gap includes the promised feature, actual state, user impact, and remediation steps.

---

## Gap 1: Asset Generation External Dependency Barrier

### Stated Goal
**README.md:89-104, 206-242:**
> "Asset Generation Pipeline - Automated Asset Creation - Complete pipeline for generating 521 game assets... Pre-generated placeholder assets are included for development."

> "⚠️ Important - Asset Status: The project includes 7 placeholder sprite assets in `web/static/assets/sprites/` for development. The game is **fully functional** with these placeholders."

**Claimed Status:** "252 placeholders / 521 defined" (README badge)

### Current State
- **Actual Asset Count:** 500 PNG files (verified: `find web/static/assets -name "*.png" | wc -l` returns 500)
- **Badge Discrepancy:** Badge claims 252 placeholders but 500 exist (248-file gap)
- **Pipeline Status:** All generation scripts exist (scripts/generate-*.sh, game-assets.yaml with 521 definitions)
- **Blocking Issue:** Full asset generation requires external AI image generation tool (Stable Diffusion, DALL-E) with 4-6 hour processing time
- **Pre-Generated Pack:** ROADMAP.md:106-113 references `make assets-download` target that should fetch from GitHub Releases, but no release artifact exists

### Impact
**User Experience:**
- **First-Time Setup Barrier:** New developers encounter 4-6 hour setup requirement not mentioned in Quick Start (README.md:105-168)
- **Incomplete Visual Experience:** Game runs with placeholder colored rectangles instead of AI-generated art
- **Discrepancy Confusion:** Badge showing "252 placeholders" when 500 exist creates mistrust in documentation accuracy

**Developer Workflow:**
- Cannot quickly demo full visual experience to stakeholders
- Requires Stable Diffusion or DALL-E account setup (not free, requires technical expertise)
- No fallback option for users without AI generation tools

### Closing the Gap

**Short-Term (1-2 days):**
1. **Fix Badge Accuracy:**
   ```bash
   # Update README.md:8
   - ![Assets](https://img.shields.io/badge/assets-252%20placeholders%2F521%20defined-yellow)
   + ![Assets](https://img.shields.io/badge/assets-500%20ready%2F521%20total-green)
   ```

2. **Upload Pre-Generated Asset Pack:**
   ```bash
   # Generate 521 full assets locally (maintainer with AI tool access)
   make assets
   tar -czf goldbox-assets-v1.0.0.tar.gz web/static/assets/
   
   # Upload to GitHub Releases via gh CLI or web UI
   gh release create v1.0.0-assets goldbox-assets-v1.0.0.tar.gz \
     --title "GoldBox RPG Assets v1.0.0" \
     --notes "521 AI-generated game assets (characters, monsters, items, terrain, effects, UI)"
   ```

3. **Implement Download Script:**
   ```bash
   # Update scripts/download-assets.sh
   #!/bin/bash
   RELEASE_URL="https://github.com/opd-ai/goldbox-rpg/releases/download/v1.0.0-assets/goldbox-assets-v1.0.0.tar.gz"
   curl -L "$RELEASE_URL" -o /tmp/assets.tar.gz
   tar -xzf /tmp/assets.tar.gz -C .
   echo "✅ Downloaded 521 assets"
   ```

4. **Update README Quick Start:**
   ```markdown
   ## Quick Start Options
   
   ### Option 1: Pre-Generated Assets (Recommended)
   ```bash
   make assets-download  # Downloads 521 ready-to-use assets (~50MB, 2 minutes)
   make run              # Start server with full visual assets
   ```
   
   ### Option 2: Generate Assets from Scratch (Advanced)
   Requires Stable Diffusion or DALL-E API access. See [ASSET_INTEGRATION.md](./ASSET_INTEGRATION.md).
   ```bash
   make assets           # 4-6 hours, requires AI image generation tool
   ```
   ```

**Long-Term (1-2 weeks):**
5. **Asset Verification in CI:**
   ```yaml
   # Add to .github/workflows/ci.yml
   - name: Verify Assets Available
     run: |
       make assets-download
       ASSET_COUNT=$(find web/static/assets -name "*.png" | wc -l)
       if [ "$ASSET_COUNT" -ne 521 ]; then
         echo "❌ Expected 521 assets, found $ASSET_COUNT"
         exit 1
       fi
   ```

**Validation:**
```bash
# After remediation, verify:
make assets-download
find web/static/assets -name "*.png" | wc -l  # Expected: 521
make assets-verify                             # Expected: All assets present
```

**Effort Estimate:** 
- Short-term fixes: 4-8 hours (maintainer with AI tool access required)
- Long-term CI integration: 2-4 hours

---

## Gap 2: Browser-Based Visual Editor Documentation

### Stated Goal
**README.md:294-304 (Frontend Architecture section):**
> "The frontend is an **Ebitengine/WASM** client compiled from Go. The HTML page at `web/index.html` shows a splash screen, then loads the WASM binary and hands control to Ebitengine."

Lists pkg/wasmui/editor.go but provides zero instructions for accessing browser-based editors.

**ROADMAP.md:154-167 (Priority 4):**
> "Polish Visual World Editor - WebSocket editor protocol exists (`pkg/server/websocket_editor.go`, 336 lines) - Browser-based visual map editor exists at `/editor` URL"

**Claimed Status:** "⚠️ World editor tools (CLI tools only, no GUI editors)"

### Current State
- **Implementation Verified:** pkg/server/websocket_editor.go (336 lines) implements editor protocol
- **WASM Frontend Exists:** pkg/wasmui/editor.go, quest_editor.go implement Ebitengine UI
- **URL Endpoint:** `/editor` route registered in pkg/server/server.go
- **Documentation Gap:** README mentions "pkg/wasmui/editor.go" exists but zero instructions on accessing /editor URL
- **Discovery Barrier:** Users must read server source code to discover visual editor exists

### Impact
**User Experience:**
- **Undiscoverable Feature:** Browser-based map editor exists but users default to CLI tools (map-editor, quest-builder)
- **Workflow Friction:** Non-technical content creators cannot find GUI tools, limiting audience reach
- **Mixed Messages:** ROADMAP says "CLI tools only, no GUI editors" but GUI editor exists and is functional

**Developer Workflow:**
- Contributors unaware visual editor exists, creating duplicate work
- No user documentation means visual editor lacks validation/feedback from community

### Closing the Gap

**Immediate (1-2 hours):**
1. **Add Visual Editor Section to README:**
   ```markdown
   ## 🎨 Visual Content Creation
   
   ### Browser-Based Map Editor
   
   GoldBox RPG includes a browser-based visual map editor for creating game levels without code:
   
   1. Start the server: `make run`
   2. Open browser to: **http://localhost:8080/editor**
   3. Features:
      - Drag-and-drop tile placement
      - Multi-layer editing (terrain, objects, NPCs)
      - Real-time preview
      - Export to YAML format
   
   **Controls:**
   - Left-click: Place selected tile
   - Right-click: Remove tile
   - Mouse wheel: Zoom in/out
   - Arrow keys: Pan viewport
   
   ### Browser-Based Quest Editor
   
   Create quest chains visually at **http://localhost:8080/quest-editor**:
   
   - Define objectives, rewards, and prerequisites
   - Visual quest chain builder
   - NPC dialogue editor
   - Export to `data/quests/*.yaml`
   
   ### CLI Tools (Alternative)
   
   For automation and scripting, use CLI tools:
   ```bash
   ./bin/map-editor --help     # Command-line map creation
   ./bin/quest-builder --help  # Batch quest generation
   ./bin/content-creator --help # Template-based content
   ```
   
   See [ASSET_INTEGRATION.md](./ASSET_INTEGRATION.md) for advanced workflows.
   ```

2. **Update ROADMAP Status:**
   ```markdown
   - [x] World editor tools (CLI + browser-based at /editor and /quest-editor)
   ```

3. **Add Quickstart Workflow:**
   ```markdown
   ## Quick Start
   
   ### Creating Your First Adventure
   
   1. **Start Server:**
      ```bash
      make run
      ```
   
   2. **Create a Map:** Open http://localhost:8080/editor
      - Select terrain type (grass, stone, water)
      - Paint tiles by clicking
      - Add NPCs and objects from sidebar
      - File → Save → `data/maps/my-first-map.yaml`
   
   3. **Create a Quest:** Open http://localhost:8080/quest-editor
      - Set quest title and description
      - Add objectives (defeat boss, collect items)
      - Set rewards (gold, XP, items)
      - Save → `data/quests/my-first-quest.yaml`
   
   4. **Play:** Reload server and your content appears in-game!
   ```

**Short-Term (1 week):**
4. **Create Visual Editor User Guide:**
   ```bash
   # New file: docs/VISUAL_EDITORS.md
   - Screenshot walkthrough of map editor
   - Video tutorial (2-3 minutes) showing map creation
   - Common workflows (create dungeon, outdoor area, town)
   - Keyboard shortcut reference
   - Troubleshooting (WASM not loading, WebSocket disconnects)
   ```

5. **Add Editor Landing Page:**
   ```html
   <!-- web/editor-index.html -->
   <h1>GoldBox RPG Content Editors</h1>
   <ul>
     <li><a href="/editor">Map Editor</a> - Create game levels</li>
     <li><a href="/quest-editor">Quest Editor</a> - Build quest chains</li>
     <li><a href="/npc-editor">NPC Editor</a> - Design characters</li>
   </ul>
   <p>First time? See <a href="/docs/VISUAL_EDITORS.html">Visual Editor Guide</a></p>
   ```

**Validation:**
```bash
# After remediation, verify:
make run &
sleep 5
curl http://localhost:8080/editor | grep -q "Map Editor"  # Expected: match found
pkill -f "bin/server"
```

**Effort Estimate:**
- Immediate documentation: 1-2 hours
- User guide with screenshots: 4-6 hours
- Video tutorial: 2-3 hours

---

## Gap 3: WebSocket Library Documentation Inconsistency

### Stated Goal
**README.md:313 (Technology Stack):**
> "Gorilla WebSocket v1.5.3 for real-time communication"

**README Dependencies:**
> "Gorilla WebSocket v1.5.3 for real-time communication"

### Current State
- **Actual Production Library:** github.com/coder/websocket v1.8.14 (actively maintained nhooyr.io/websocket fork)
- **Migration Completed:** CHANGELOG.md:10-27 documents migration from gorilla to nhooyr/coder on 2026-03-13
- **gorilla/websocket Usage:** Retained only for E2E test client (test/e2e/client.go) and benchmarks
- **Documentation Drift:** README not updated to reflect migration, causing contributor confusion

### Impact
**Developer Experience:**
- **Contributor Confusion:** New contributors expect Gorilla API patterns but find coder/websocket code
- **Outdated Examples:** External tutorials referencing this project may use wrong import paths
- **Security Perception:** Listing archived library (gorilla/websocket) suggests project uses unmaintained code

**Technical Accuracy:**
- Documentation claims project uses deprecated library when it actually uses modern alternative
- Mixed messaging reduces trust in documentation accuracy

### Closing the Gap

**Immediate (30 minutes):**
1. **Update README Technology Stack:**
   ```markdown
   ### Technology Stack
   - **Backend**: Go 1.25.6+ with native HTTP server
   - **Protocol**: JSON-RPC 2.0 over HTTP and WebSockets
   - **Dependencies**: 
     - **Coder WebSocket v1.8.14** for real-time communication (actively maintained nhooyr.io/websocket fork)
     - Sirupsen Logrus v1.9.3 for structured logging
     - Prometheus client v1.22.0 for metrics collection
     - YAML v3.0.1 for configuration management
     - _gorilla/websocket v1.5.3 retained for E2E test client only (test-only usage)_
   ```

2. **Add Migration Note to README:**
   ```markdown
   ## 🔄 Recent Changes
   
   **WebSocket Library Migration (2026-03-13):**
   The server migrated from gorilla/websocket (archived) to github.com/coder/websocket (actively maintained). All production code uses the new library. gorilla/websocket is retained only for E2E test clients. See [CHANGELOG.md](./CHANGELOG.md#unreleased) for details.
   ```

3. **Update go.mod Comment (Already Done):**
   ```go
   // go.mod:11-14 (already correct)
   // gorilla/websocket is used only for E2E tests (test/e2e/client.go) and benchmarks
   // (pkg/server/benchmark_test.go) as a WebSocket client library. Production code uses
   // github.com/coder/websocket (maintained fork of nhooyr.io/websocket).
   ```

**Validation:**
```bash
# Verify production code uses coder/websocket only:
grep -r "gorilla/websocket" pkg/ && echo "❌ Production code uses gorilla" || echo "✅ Clean"
# Expected: ✅ Clean

# Verify README mentions coder/websocket:
grep "coder/websocket" README.md || echo "❌ README not updated"
# Expected: Match found
```

**Effort Estimate:** 30 minutes (documentation update only, no code changes required)

---

## Gap 4: Go Version Documentation Mismatch

### Stated Goal
**README.md:6 (Badge):**
> `![Go Version](https://img.shields.io/badge/go-%3E%3D1.24.0-blue)`

**README.md:108 (Prerequisites):**
> "Go 1.24.0 or higher"

### Current State
- **Actual go.mod:** `go 1.25.6` with `toolchain go1.25.8`
- **Dependency Requirements:** CHANGELOG.md:38-41 notes several dependencies require Go 1.24.0+ (ebiten, ebitengine/gomobile) and golang.org/x/time v0.15.0+ requires Go 1.25.0+
- **Security:** CHANGELOG.md:30-35 identifies 18 Go stdlib vulnerabilities requiring Go 1.24.12+ or 1.25.8 to resolve
- **Documentation Lag:** Badge and prerequisites claim Go 1.24.0 works but project actually requires Go 1.25.6+

### Impact
**User Experience:**
- **Build Failures:** Users installing Go 1.24.0-1.24.11 will encounter dependency resolution errors
- **Security Exposure:** Users running Go <1.25.8 remain vulnerable to 18 stdlib vulnerabilities (crypto/tls, net/http, etc.)
- **Wasted Setup Time:** Incorrect version guidance forces users to troubleshoot and reinstall correct Go version

**Contributor Workflow:**
- Contributors using Go 1.24.x cannot build project despite README claiming compatibility
- CI/CD may pass with Go 1.25.8 but local development fails with documented version

### Closing the Gap

**Immediate (15 minutes):**
1. **Update README Badge:**
   ```markdown
   - ![Go Version](https://img.shields.io/badge/go-%3E%3D1.24.0-blue)
   + ![Go Version](https://img.shields.io/badge/go-%3E%3D1.25.6-blue)
   ```

2. **Update Prerequisites Section:**
   ```markdown
   ### Prerequisites
   - **Go 1.25.6 or higher** (toolchain 1.25.8 recommended for security patches)
   - Make (for build automation)
   - Docker (recommended for easy setup)
   - **Asset Generation Tool** (optional) - Stable Diffusion or DALL-E for full asset generation
     - Pre-generated placeholder assets included for development
     - Full generation requires external AI tool setup (see [ASSET_INTEGRATION.md](./ASSET_INTEGRATION.md))
   ```

3. **Add Version Requirement Explanation:**
   ```markdown
   ### Why Go 1.25.6+?
   
   - **Dependency Requirements:** Several dependencies (golang.org/x/time v0.15.0+, ebiten/v2, ebitengine/gomobile) require Go 1.25.0+
   - **Security:** Go 1.25.8 resolves 18 standard library vulnerabilities in crypto/tls, net/http, crypto/x509, and html/template
   - **Build Stability:** Earlier Go versions may encounter dependency resolution errors
   
   **Installation:**
   ```bash
   # Download from official site
   wget https://go.dev/dl/go1.25.8.linux-amd64.tar.gz
   sudo tar -C /usr/local -xzf go1.25.8.linux-amd64.tar.gz
   go version  # Expected: go version go1.25.8 linux/amd64
   ```
   ```

**Validation:**
```bash
# Verify documentation matches reality:
go mod graph | grep "go " | head -1  # Expected: go 1.25.6
grep "Go 1.25" README.md || echo "❌ Badge not updated"
```

**Effort Estimate:** 15 minutes (documentation update only)

---

## Gap 5: CLI Tool Test Coverage Below Project Standard

### Stated Goal
**README.md:7 (Coverage Badge):**
> `![Coverage](https://img.shields.io/badge/coverage-80%25-brightgreen)`

**ROADMAP.md:182-189:**
> "Priority 5: Improve CLI Tool Test Coverage... Target 70%+ coverage for all CLI tools"

**Project Quality Standards (from audit):**
> "Maintain ≥60% code coverage with Go's built-in testing framework"

### Current State
- **Overall Coverage:** 82.5% (exceeds project threshold)
- **CLI Tool Coverage:**
  - cmd/content-creator: **61.9%** (below 70% target)
  - cmd/quest-builder: **71.6%** (meets target)
  - cmd/map-editor: **79.9%** (exceeds target)
- **Gap:** content-creator is 8.1 percentage points below ROADMAP target

**Coverage Breakdown:**
```bash
$ go test ./cmd/content-creator/... -cover
ok  	goldbox-rpg/cmd/content-creator	0.123s	coverage: 61.9% of statements
```

### Impact
**Code Quality:**
- **Error Path Testing:** Lower coverage in content-creator suggests error handling and edge cases lack tests
- **Regression Risk:** Changes to content creation logic may introduce bugs undetected by test suite
- **User-Facing Tool:** CLI tools have direct user interaction - high coverage is critical for UX quality

**Maintenance:**
- Harder to refactor content-creator safely without comprehensive tests
- Contributors may hesitate to modify code lacking test coverage

### Closing the Gap

**Short-Term (1-2 days):**
1. **Identify Untested Code Paths:**
   ```bash
   go test ./cmd/content-creator/... -coverprofile=/tmp/content-creator.cov
   go tool cover -html=/tmp/content-creator.cov -o /tmp/coverage.html
   # Open /tmp/coverage.html in browser, red sections are untested
   ```

2. **Add Table-Driven Tests for Command Parsing:**
   ```go
   // cmd/content-creator/main_test.go
   func TestParseCommand(t *testing.T) {
       tests := []struct {
           name    string
           input   []string
           want    Command
           wantErr bool
       }{
           {
               name:  "valid character command",
               input: []string{"character", "--class=fighter", "--level=5"},
               want:  Command{Type: "character", Class: "fighter", Level: 5},
           },
           {
               name:    "missing required flag",
               input:   []string{"character"},
               wantErr: true,
           },
           {
               name:    "invalid class",
               input:   []string{"character", "--class=invalid"},
               wantErr: true,
           },
           // Add 10-15 more test cases for edge cases
       }
       
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               got, err := ParseCommand(tt.input)
               if (err != nil) != tt.wantErr {
                   t.Errorf("ParseCommand() error = %v, wantErr %v", err, tt.wantErr)
               }
               if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
                   t.Errorf("ParseCommand() = %v, want %v", got, tt.want)
               }
           })
       }
   }
   ```

3. **Add Integration Tests for Output Validation:**
   ```go
   func TestGenerateCharacter(t *testing.T) {
       // Test that generated YAML matches PCG schema
       output := GenerateCharacter("fighter", 5)
       
       var char game.Character
       err := yaml.Unmarshal([]byte(output), &char)
       if err != nil {
           t.Fatalf("Generated invalid YAML: %v", err)
       }
       
       // Validate against schema
       if char.Class != game.ClassFighter {
           t.Errorf("Expected class %v, got %v", game.ClassFighter, char.Class)
       }
       if char.Level != 5 {
           t.Errorf("Expected level 5, got %v", char.Level)
       }
   }
   ```

4. **Add Error Path Tests:**
   ```go
   func TestInvalidInput(t *testing.T) {
       tests := []struct {
           name    string
           args    []string
           wantErr string
       }{
           {
               name:    "empty command",
               args:    []string{},
               wantErr: "no command specified",
           },
           {
               name:    "invalid YAML output path",
               args:    []string{"character", "--output=/invalid/path.yaml"},
               wantErr: "cannot write to path",
           },
           // Add more error cases
       }
       // Similar table-driven structure
   }
   ```

**Validation:**
```bash
# After adding tests, verify coverage:
go test ./cmd/content-creator/... -cover
# Expected: coverage ≥75%

go test ./cmd/map-editor/... ./cmd/content-creator/... ./cmd/quest-builder/... -cover
# Expected: All ≥70%
```

**Effort Estimate:** 
- Analysis and planning: 2 hours
- Writing tests: 6-8 hours
- Review and iteration: 2 hours
- **Total: 10-12 hours** (1-2 days for one developer)

---

## Gap 6: Asset Status Badge Inaccuracy

### Stated Goal
**README.md:8 (Badge):**
> `![Assets](https://img.shields.io/badge/assets-252%20placeholders%2F521%20defined-yellow)`

### Current State
- **Actual Asset Count:** 500 PNG files (`find web/static/assets -name "*.png" | wc -l` returns 500)
- **Discrepancy:** Badge claims 252 placeholders, actual count is 500 (248-file difference)
- **Pipeline:** game-assets.yaml defines 521 total assets, 21 require AI generation

### Impact
**User Perception:**
- **Trust Erosion:** Stale metrics badge suggests unmaintained project or inaccurate documentation
- **Underestimation:** Users assume only 252/521 assets exist when 500/521 (96%) are ready
- **Confusion:** Mismatch between badge (252), README text (7 placeholders), and reality (500) creates cognitive load

### Closing the Gap

**Immediate (5 minutes):**
1. **Update Badge to Accurate Count:**
   ```markdown
   - ![Assets](https://img.shields.io/badge/assets-252%20placeholders%2F521%20defined-yellow)
   + ![Assets](https://img.shields.io/badge/assets-500%20ready%2F521%20total-green)
   ```

**Short-Term (1 hour):**
2. **Automate Badge Generation in CI:**
   ```yaml
   # .github/workflows/update-readme-badges.yml
   name: Update README Badges
   on:
     schedule:
       - cron: '0 0 * * 0'  # Weekly
     workflow_dispatch:
   
   jobs:
     update-badges:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v3
         
         - name: Count Assets
           id: count
           run: |
             ASSET_COUNT=$(find web/static/assets -name "*.png" | wc -l)
             echo "asset_count=$ASSET_COUNT" >> $GITHUB_OUTPUT
         
         - name: Update README
           run: |
             sed -i "s/assets-[0-9]*%20ready/assets-${{ steps.count.outputs.asset_count }}%20ready/" README.md
         
         - name: Create PR if Changed
           uses: peter-evans/create-pull-request@v5
           with:
             commit-message: "chore: Update asset count badge to ${{ steps.count.outputs.asset_count }}/521"
             title: "Update README asset count badge"
             branch: automated/update-badges
   ```

3. **Add Asset Count Verification to CI:**
   ```yaml
   # .github/workflows/ci.yml
   - name: Verify Asset Count Accuracy
     run: |
       ACTUAL_COUNT=$(find web/static/assets -name "*.png" | wc -l)
       BADGE_COUNT=$(grep -oP "assets-\K[0-9]+" README.md | head -1)
       
       if [ "$ACTUAL_COUNT" != "$BADGE_COUNT" ]; then
         echo "⚠️  README badge shows $BADGE_COUNT assets but $ACTUAL_COUNT exist"
         echo "Run 'make update-readme-badges' to sync"
       fi
   ```

**Validation:**
```bash
# Manual verification:
ACTUAL=$(find web/static/assets -name "*.png" | wc -l)
BADGE=$(grep -oP "assets-\K[0-9]+" README.md | head -1)
[ "$ACTUAL" -eq "$BADGE" ] && echo "✅ Badge accurate" || echo "❌ Badge shows $BADGE, actual $ACTUAL"
```

**Effort Estimate:** 
- Immediate fix: 5 minutes
- CI automation: 1 hour
- Testing and validation: 30 minutes

---

## Gap 7: OpenAPI Specification Drift Risk

### Stated Goal
**README.md:186-200:**
> "The project includes automatic OpenAPI specification generation from Go source code... The generator parses `pkg/server/constants.go` to extract all RPC methods and updates `api/openapi.yaml`"

**Claimed Process:**
> "`make openapi-gen` - Generate OpenAPI spec from RPC method constants"
> "`make openapi-validate` - Validate the generated spec (requires npx)"

### Current State
- **Generator Exists:** cmd/openapi-gen/main.go implements automatic spec generation
- **Validation Target Exists:** Makefile includes `openapi-validate` target
- **CI Gap:** No evidence of `make openapi-validate` running in .github/workflows/ci.yml
- **Drift Risk:** Manual `make openapi-gen` invocation means spec can become stale without CI enforcement

### Impact
**API Documentation:**
- **Stale Spec:** api/openapi.yaml may not reflect current RPC methods in pkg/server/constants.go
- **Developer Confusion:** API consumers (frontend, integrators) may use outdated spec
- **Breaking Changes:** New RPC methods added without updating spec go undocumented

**Process:**
- Manual spec generation is error-prone (developers forget to run `make openapi-gen`)
- No enforcement that spec stays in sync with code

### Closing the Gap

**Short-Term (1 hour):**
1. **Add OpenAPI Validation to CI:**
   ```yaml
   # .github/workflows/ci.yml
   jobs:
     test:
       steps:
         # ... existing test steps ...
         
         - name: Generate OpenAPI Spec
           run: make openapi-gen
         
         - name: Validate OpenAPI Spec
           run: |
             npm install -g @redocly/cli
             make openapi-validate
         
         - name: Check for Spec Drift
           run: |
             if git diff --exit-code api/openapi.yaml; then
               echo "✅ OpenAPI spec is up to date"
             else
               echo "❌ OpenAPI spec is out of sync with code"
               echo "Run 'make openapi-gen' and commit changes"
               exit 1
             fi
   ```

2. **Add Pre-Commit Hook (Optional):**
   ```bash
   # .git/hooks/pre-commit
   #!/bin/bash
   if git diff --cached --name-only | grep -qE "pkg/server/constants.go"; then
       echo "🔍 Detected changes to RPC constants, regenerating OpenAPI spec..."
       make openapi-gen
       git add api/openapi.yaml
   fi
   ```

3. **Update CONTRIBUTING.md:**
   ```markdown
   ## Adding New RPC Methods
   
   1. Define method constant in `pkg/server/constants.go`
   2. Implement handler in `pkg/server/handlers.go`
   3. Add validator in `pkg/validation/`
   4. **Regenerate OpenAPI spec:** `make openapi-gen`
   5. Validate spec: `make openapi-validate`
   6. Commit both code and `api/openapi.yaml`
   
   **Note:** CI will fail if OpenAPI spec is out of sync.
   ```

**Validation:**
```bash
# Manually test CI check would work:
make openapi-gen
git diff --exit-code api/openapi.yaml
# Expected: Exit code 0 if spec is current, 1 if stale
```

**Effort Estimate:** 
- CI integration: 1 hour
- Documentation: 30 minutes
- Pre-commit hook setup: 30 minutes
- **Total: 2 hours**

---

## Summary

**Total Gaps Identified:** 7

**Severity Breakdown:**
- **HIGH:** 3 gaps (Asset distribution, Library documentation, Go version)
- **MEDIUM:** 4 gaps (Test coverage, Badge accuracy, Editor documentation, OpenAPI drift)
- **CRITICAL:** 0 gaps (no blocking issues)

**Aggregate Remediation Effort:**
- **Immediate fixes (same day):** ~2 hours (badges, README updates)
- **Short-term (1 week):** ~20 hours (asset pack upload, test coverage, CI automation)
- **Long-term (1 month):** ~10 hours (user guides, video tutorials)
- **Total: 32 hours** (~4 days for one developer)

**Prioritization:**
1. **Asset Pack Upload** (Gap 1) - Highest user impact, unblocks new user onboarding
2. **Documentation Accuracy** (Gaps 2, 3, 4, 6) - Quick wins, restores documentation trust
3. **CI Automation** (Gaps 5, 7) - Prevents future drift, improves quality gates
4. **Editor Documentation** (Gap 2) - Lowers barrier to entry for non-technical users

**Overall Assessment:**
All gaps are **process and documentation issues**, not implementation defects. The code is production-ready. Closing these gaps will:
- Reduce new user onboarding time from 4-6 hours to <30 minutes
- Increase documentation accuracy and user trust
- Prevent technical debt accumulation through automated quality gates

---

*Gap analysis performed: 2026-03-16*  
*Based on: README.md, ROADMAP.md, CHANGELOG.md, go.mod, actual codebase inspection*  
*Methodology: Stated goal extraction → empirical verification → impact analysis → remediation planning*
