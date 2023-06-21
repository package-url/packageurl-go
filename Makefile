.PHONY: test clean lint

test:
	curl -Ls https://raw.githubusercontent.com/package-url/purl-spec/master/test-suite-data.json -o testdata/test-suite-data.json
	go test -v -cover ./...

lint:
	go get -u golang.org/x/lint/golint
	golint -set_exit_status
