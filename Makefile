


fmt:
	find . -name '*.go' -exec gofumpt -w -s -extra {} \;

doc:
	find ./pkg -type d -exec bash -c "godocdown {} | tee {}/doc.md" \;
	rm -f data/doc.md data/*/doc.md cmd/server/doc.md web/doc.md web/*/doc.md game/doc.md game/index.html
	find ./.git -name 'doc.md' -exec rm -vf {} \;
	find ./web -name 'doc.md' -exec rm -v {} \;
	find ./pkg -name 'doc.md' -exec git add -v {} \;
	find ./pkg -name 'doc.md' -exec projects -index -mdoverride {} \;
	find ./web -name 'index.html' -exec rm -v {} \;
	find ./pkg -name 'index.html' -exec git add -v {} \;

yaml:
	find . -name '*.go' -exec code2prompt --template ~/code2prompt/templates/yaml.hbs --output {}.md {} \;

godoc:
	find . -name '*.go' -exec code2prompt --template ~/code2prompt/templates/document-the-code.hbs --output {}.md {} \;r

clean:
	find . -name '*.go.md' -exec rm -v {} \;