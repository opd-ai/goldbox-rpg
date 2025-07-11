build :
	go build -o bin/server cmd/server/main.go

run: build
	./bin/server

test:
	go test ./... -v

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
	find ./web -name '*.js' -exec ./node_modules/.bin/prettier --write {} \;

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

clean:
	find . -name '*.go.md' -exec rm -v {} \;
	find . -name '*.out' -exec rm -v {} \;
	find . -name '*.test' -exec rm -v {} \;
	find . -name '*.test' -exec rm -v {} \;
	make doc

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