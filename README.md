# packageurl-go

[![build](https://github.com/package-url/packageurl-go/workflows/test/badge.svg)](https://github.com/package-url/packageurl-go/actions?query=workflow%3Atest) [![Coverage Status](https://coveralls.io/repos/github/package-url/packageurl-go/badge.svg)](https://coveralls.io/github/package-url/packageurl-go) [![PkgGoDev](https://pkg.go.dev/badge/github.com/package-url/packageurl-go)](https://pkg.go.dev/github.com/package-url/packageurl-go) [![Go Report Card](https://goreportcard.com/badge/github.com/package-url/packageurl-go)](https://goreportcard.com/report/github.com/package-url/packageurl-go)

Go implementation of the package url spec.


## Install
```
go get -u github.com/package-url/packageurl-go
```

## Versioning

The versions will follow the spec. So if the spec is released at ``1.0``. Then all versions in the ``1.x.y`` will follow the ``1.x`` spec.


## Usage

### Create from parts
```go
package main

import (
	"fmt"

	"github.com/package-url/packageurl-go"
)

func main() {
	instance := packageurl.NewPackageURL("test", "ok", "name", "version", nil, "")
	fmt.Printf("%s", instance.ToString())
}
```

### Parse from string
```go
package main

import (
	"fmt"

	"github.com/package-url/packageurl-go"
)

func main() {
	instance, err := packageurl.FromString("test:ok/name@version")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v", instance)
}

```


## Test
Testing uses the normal ``go test`` command. 

There are two main categories of tests. Both are run with `make test`.

1. Tests that exercise the code against the [purl-spec testsuite](https://github.com/package-url/purl-spec/tree/main/tests): `TestCoreSpec` and `TestPurlTypes`.
   The testsuite is included as a git submodule.
   The testsuite is under construction and sees frequent changes. 
   Such changes can be pulled in with `make testsuite-update`, which updates the submodule.
2. Tests that try to verify behavior not necessarily covered by the purl-spec testsuite (any other `Test*` function).


## Fuzzing

Fuzzing is done with standard [Go fuzzing](https://go.dev/doc/fuzz/), introduced in Go 1.18.

Fuzz tests check for inputs that cause `FromString` to panic.

Using `make fuzz` will run fuzz tests for one minute.

To run fuzz tests longer:

```
go test -fuzztime=60m -fuzz .
```

Or omit `-fuzztime` entirely to run indefinitely.
