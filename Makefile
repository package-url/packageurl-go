.PHONY: test clean lint

# The purl-spec testsuite is defined in a separate repo.
#
# See https://github.com/package-url/purl-spec/blob/main/docs/tests/test-suite.md
testdata/purl-spec/tests:
	git submodule update --init

# Bring in the latest version of the upstream testspec.
testsuite-update: testdata/purl-spec/tests
	git submodule update --remote

test: testdata/purl-spec/tests
	go test -v -cover ./...

fuzz:
	go test -fuzztime=1m -fuzz .

clean:
	find . -name "test-suite-data.json" | xargs rm -f

lint:
	go get -u golang.org/x/lint/golint
	golint -set_exit_status
