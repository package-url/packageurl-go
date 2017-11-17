.PHONY: test clean lint

test:
	curl -L https://raw.githubusercontent.com/package-url/purl-test-suite/master/test-suite-data.json -o testdata/test-suite-data.json
	go test -v -cover ./...

clean:
	find . -name "test-suite-data.json" | xargs rm -f

lint:
	go get -u github.com/golang/lint/golint
	golint -set_exit_status
