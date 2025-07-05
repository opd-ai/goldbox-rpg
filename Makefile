
build :
	go build -o bin/server cmd/server/main.go

run: build
	./bin/server


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
	find . -name '*.html' -exec rm -v {} \;
	make doc