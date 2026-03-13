build: wasm
	go build -o bin/server cmd/server/main.go

run: build
	./bin/server

test:
	go test ./... -v

###################
# WASM UI Build
###################

.PHONY: wasm wasm-deps wasm-clean

# Build WASM UI (Ebitengine-based game client)
wasm: wasm-deps
	@echo "Building WASM UI..."
	GOOS=js GOARCH=wasm go build -o web/static/js/game.wasm ./cmd/wasm-ui
	@echo "WASM build complete: web/static/js/game.wasm"

# Install WASM dependencies (wasm_exec.js)
wasm-deps:
	@echo "Copying wasm_exec.js..."
	@if [ -f "$$(go env GOROOT)/misc/wasm/wasm_exec.js" ]; then \
		cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" web/static/js/; \
	elif [ -f "$$(go env GOROOT)/lib/wasm/wasm_exec.js" ]; then \
		cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" web/static/js/; \
	else \
		echo "Error: wasm_exec.js not found in Go installation"; \
		exit 1; \
	fi

# Clean WASM build artifacts
wasm-clean:
	rm -f web/static/js/game.wasm
	rm -f web/static/js/editor.wasm
	rm -f web/static/js/wasm_exec.js

# Build WASM Map Editor
wasm-editor: wasm-deps
	@echo "Building WASM Map Editor..."
	GOOS=js GOARCH=wasm go build -o web/static/js/editor.wasm ./cmd/wasm-editor
	@echo "WASM editor build complete: web/static/js/editor.wasm"

# Build both server and WASM UI
build-all: build wasm wasm-editor
	@echo "All builds complete!"

# Run E2E integration tests
test-e2e: build
	go test ./test/e2e/... -v -timeout 5m

# Run E2E tests with race detector
test-e2e-race: build
	go test ./test/e2e/... -v -race -timeout 5m

# Test coverage analysis
test-coverage:
	./scripts/analyze_test_coverage.sh

test-coverage-verbose:
	./scripts/analyze_test_coverage.sh -v

test-coverage-json:
	./scripts/analyze_test_coverage.sh -j

find-untested:
	./scripts/find_untested_files.sh

fmt:
	find ./pkg -name '*.go' -exec gofumpt -w -s -extra {} \;
	find ./cmd -name '*.go' -exec gofumpt -w -s -extra {} \;

doc:
	find ./pkg -type d -exec bash -c "godocdown {} | tee {}/doc.md" \;
	rm -f data/doc.md data/*/doc.md cmd/server/doc.md web/doc.md web/*/doc.md game/doc.md game/index.html pkg/doc.md
	find ./.git -name 'doc.md' -exec rm -vf {} \;
	find ./web -name 'doc.md' -exec rm -v {} \;
	find ./pkg -name 'doc.md' -exec git add -v {} \;
	find ./pkg -name 'doc.md' -exec projects -index -mdoverride {} \;
	find ./pkg -name 'index.html' -exec git add -v {} \;
	projects -index -mdoverride ./pkg/README-RPC.md

yaml:
	find . -name '*.go' -exec code2prompt --template ~/code2prompt/templates/yaml.hbs --output {}.md {} \;

godoc:
	find . -name '*.go' -exec code2prompt --template ~/code2prompt/templates/document-the-code.hbs --output {}.md {} \;r

###################
# OpenAPI Generation
###################

.PHONY: openapi-gen openapi-validate

# Generate OpenAPI spec from Go code
openapi-gen:
	@echo "Generating OpenAPI spec from Go source code..."
	go run cmd/openapi-gen/main.go
	@echo "OpenAPI spec updated: api/openapi.yaml"

# Validate OpenAPI spec (requires npx/openapi-generator-cli)
openapi-validate:
	@echo "Validating OpenAPI spec..."
	@if command -v npx >/dev/null 2>&1; then \
		cd api && npx @redocly/cli lint openapi.yaml; \
	else \
		echo "Warning: npx not found. Skipping validation."; \
	fi

clean:
	find . -name '*.go.md' -exec rm -v {} \;
	find . -name '*.out' -exec rm -v {} \;
	find . -name '*.test' -exec rm -v {} \;
	find . -name '*.test' -exec rm -v {} \;
	make doc

###################
# Asset Generation
###################

.PHONY: assets assets-preview assets-clean assets-optimize assets-verify assets-priority

# Generate all game assets using the pipeline
assets:
	@echo "Generating all game assets..."
	./scripts/generate-all.sh --seed 42

# Preview asset generation without creating files (dry-run)
assets-preview:
	@echo "Previewing asset generation..."
	./scripts/generate-all.sh --dry-run

# Generate only Priority 1 (critical) assets for quick testing
assets-priority:
	@echo "Generating priority assets..."
	./scripts/generate-priority1.sh

# Optimize generated assets for production
assets-optimize:
	@echo "Optimizing assets..."
	./scripts/post-process.sh

# Verify that all required assets have been generated
assets-verify:
	@echo "Verifying assets..."
	./scripts/verify-assets.sh

# Generate placeholder assets (no AI required)
assets-placeholders:
	@echo "Generating placeholder assets..."
	./scripts/generate-placeholders.sh
	@echo "Placeholder assets generated"

# Generate adventure-specific placeholder assets
adventures-placeholders:
	@echo "Generating adventure placeholder assets..."
	./scripts/generate-adventure-placeholders.sh
	@echo "Adventure placeholder assets generated"

# Verify all adventures load and pass validation
adventures-verify:
	@echo "Verifying all adventures..."
	@go run scripts/verify_adventures.go
	@echo "Adventure verification complete"

# Clean all generated assets
assets-clean:
	@echo "Cleaning generated assets..."
	rm -rf ./web/static/assets/sprites/characters/
	rm -rf ./web/static/assets/sprites/monsters/
	rm -rf ./web/static/assets/sprites/items/
	rm -rf ./web/static/assets/sprites/terrain/
	rm -rf ./web/static/assets/sprites/effects/
	rm -rf ./web/static/assets/sprites/ui/
	@echo "Generated assets cleaned"

###################
# Docker Commands
###################

# Build Docker image
docker-build:
	docker build -t goldbox-rpg .

# Run Docker container
docker-run:
	docker run -p 8080:8080 goldbox-rpg

# Build and run in one command
docker:
	docker run -p 8080:8080 $$(docker build -q .)

# Run in development mode (shows logs)
docker-dev:
	docker run --rm -p 8080:8080 goldbox-rpg

# Check if container is healthy
docker-health:
	curl -f http://localhost:8080/health