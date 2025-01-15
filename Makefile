


fmt:
	find . -name '*.go' -exec gofumpt -w -s -extra {} \;

doc:
	find ./ -type d -exec bash -c "godocdown {} | tee {}/doc.md" \;
	rm -f data/doc.md data/*/doc.md cmd/server/doc.md web/doc.md web/*/doc.md
	find ./.git -name 'doc.md' -exec rm -vf {} \;
	find ./ -name 'doc.md' -exec git add -v {} \;

yaml:
	find . -name '*.go' -exec code2prompt --template ~/code2prompt/templates/yaml.hbs --output {}.md {} \;

clean:
	find . -name '*.go.md' -exec rm -v {} \;