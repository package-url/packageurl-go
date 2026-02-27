/*
Copyright (c) the purl authors

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package packageurl_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/package-url/packageurl-go"
)

// OrderedMap is used to store the TestFixture.QualifierMap, to ensure that the
// declaration order of qualifiers is preserved.
type OrderedMap struct {
	OrderedKeys []string
	Map         map[string]string
}

// qualifiersMapPattern is used to parse the TestFixture "qualifiers" field to
// ensure that it's a json object.
var qualifiersMapPattern = regexp.MustCompile(`(?ms)^\{.*\}$`)

// UnmarshalJSON unmarshals the qualifiers field for a TestFixture. The
// qualifiers field is given as a json object such as:
//
//	"qualifiers": {"arch": "i386", "distro": "fedora-25"}
//
// This function performs in-order parsing of these values into an OrderedMap to
// preserve items in order of declaration. Note that parsing as a
// map[string]string won't preserve element order.
func (m *OrderedMap) UnmarshalJSON(bytes []byte) error {
	data := string(bytes)
	switch data {
	case "null":
		m.OrderedKeys = []string{}
		m.Map = make(map[string]string)
		return nil
	default:
		// ensure that the data is a json object "{...}"
		if !qualifiersMapPattern.MatchString(data) {
			return fmt.Errorf("qualifiers parse error: not a json object: %s", data)
		}

		// find out the order in which map keys occur
		dec := json.NewDecoder(strings.NewReader(data))
		// consume opening '{'
		_, _ = dec.Token()
		for dec.More() {
			t, _ := dec.Token()
			switch token := t.(type) {
			case json.Delim:
				if token != '}' {
					return fmt.Errorf("qualifiers parse error: expected delimiter '}', got: %v", token)
				}
				// closed json object -> we're done
			case string:
				// this token is a dictionary key
				m.OrderedKeys = append(m.OrderedKeys, token)
				// consume the value (the token following the colon after the key)
				_, _ = dec.Token()
			}
		}

		// now that we know the key order, just fill the OrderedMap.Map field
		if err := json.Unmarshal(bytes, &m.Map); err != nil {
			return err
		}
		return nil
	}
}

type ComponentData struct {
	PackageType  string     `json:"type"`
	Namespace    string     `json:"namespace"`
	Name         string     `json:"name"`
	Version      string     `json:"version"`
	QualifierMap OrderedMap `json:"qualifiers"`
	Subpath      string     `json:"subpath"`
}

// Qualifiers converts the ComponentData.QualifierMap field to an object of type
// packageurl.Qualifiers.
func (t ComponentData) Qualifiers() packageurl.Qualifiers {
	q := packageurl.Qualifiers{}

	for _, key := range t.QualifierMap.OrderedKeys {
		q = append(q, packageurl.Qualifier{Key: key, Value: t.QualifierMap.Map[key]})
	}

	return q
}

type ComponentsOrPurl struct {
	Purl          *string
	PurlComponent *ComponentData
}

func (cop *ComponentsOrPurl) UnmarshalJSON(data []byte) error {
	// Try string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		cop.Purl = &s
		return nil
	}

	var comp ComponentData
	if err := json.Unmarshal(data, &comp); err == nil {
		cop.PurlComponent = &comp
		return nil
	}

	return fmt.Errorf("ComponentsOrPurl: data is neither a string nor PURL component")
}

type TestFixture struct {
	Description           string           `json:"description"`
	TestGroup             string           `json:"test_group"`
	TestType              string           `json:"test_type"`
	Input                 ComponentsOrPurl `json:"input"`
	ExpectedFailure       bool             `json:"expected_failure"`
	ExpectedOutput        ComponentsOrPurl `json:"expected_output"`
	ExpectedFailureReason *string          `json:"expected_failure_reason"`
}

type TestSuite struct {
	Schema string        `json:"$schema"`
	Tests  []TestFixture `json:"tests"`
}

type jsonFile struct {
	name    string
	content []byte
}

func readJSONFilesFromDir(dirPath string) ([]jsonFile, error) {
	var result []jsonFile

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("reading dir %s: %w", dirPath, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		fullPath := filepath.Join(dirPath, entry.Name())
		data, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("reading file %s: %w", fullPath, err)
		}

		result = append(result, jsonFile{name: entry.Name(), content: data})
	}

	return result, nil
}

func roundTripTest(tc TestFixture, t *testing.T) {
	p, err := packageurl.FromString(*tc.Input.Purl)
	if tc.ExpectedFailure == false {
		if err != nil {
			t.Logf("%s failed: %s", tc.Description, err)
			t.Fail()
		}

		if tc.ExpectedOutput.Purl != nil {
			if *tc.ExpectedOutput.Purl != p.String() {
				t.Logf("%s: '%s' test failed: wanted: '%s', got '%s'", tc.Description, tc.TestType, *tc.ExpectedOutput.Purl, p.String())
				t.Fail()
			}
		} else {
			t.Logf("%s: expected output nil: '%s'", tc.Description, *tc.ExpectedOutput.Purl)
			t.Fail()
		}

	} else {
		if err == nil {
			t.Logf("%s did not fail and returned %#v", tc.Description, p)
			t.Fail()
		}

	}
}

func parseTest(tc TestFixture, t *testing.T) {
	p, err := packageurl.FromString(*tc.Input.Purl)
	if tc.ExpectedFailure == false {
		if err != nil {
			t.Logf("%s failed: %s", tc.Description, err)
			t.Fail()
		}
		// verify parsing
		expected := tc.ExpectedOutput.PurlComponent
		if p.Type != expected.PackageType {
			t.Logf("%s: incorrect package type: wanted: '%s', got '%s'", tc.Description, expected.PackageType, p.Type)
			t.Fail()
		}
		if p.Namespace != expected.Namespace {
			t.Logf("%s: incorrect namespace: wanted: '%s', got '%s'", tc.Description, expected.Namespace, p.Namespace)
			t.Fail()
		}
		if p.Name != expected.Name {
			t.Logf("%s: incorrect name: wanted: '%s', got '%s'", tc.Description, expected.Name, p.Name)
			t.Fail()
		}
		if p.Version != expected.Version {
			t.Logf("%s: incorrect version: wanted: '%s', got '%s'", tc.Description, expected.Version, p.Version)
			t.Fail()
		}
		want := expected.Qualifiers()
		sort.Slice(want, func(i, j int) bool {
			return want[i].Key < want[j].Key
		})
		got := p.Qualifiers
		sort.Slice(got, func(i, j int) bool {
			return got[i].Key < got[j].Key
		})
		if !reflect.DeepEqual(want, got) {
			t.Logf("%s: incorrect qualifiers: wanted: '%#v', got '%#v'", tc.Description, want, p.Qualifiers)
			t.Fail()
		}

		if p.Subpath != expected.Subpath {
			t.Logf("%s: incorrect subpath: wanted: '%s', got '%s'", tc.Description, expected.Subpath, p.Subpath)
			t.Fail()
		}
	} else {
		// Invalid cases
		if err == nil {
			t.Logf("%s did not fail and returned %#v", tc.Description, p)
			t.Fail()
		}
	}

}

func buildTest(tc TestFixture, t *testing.T) {
	input := tc.Input.PurlComponent
	instance := packageurl.NewPackageURL(
		input.PackageType, input.Namespace, input.Name, input.Version,
		// Use QualifiersFromMap so that the qualifiers have a defined order, which is needed for string comparisons
		packageurl.QualifiersFromMap(input.Qualifiers().Map()), input.Subpath)
	result := instance.ToString()
	canonicalExpectedPurl := tc.ExpectedOutput.Purl

	if tc.ExpectedFailure {
		// String()/ToString() signature won't error so the only reasonable thing is to check this here.
		err := instance.Normalize()
		if err == nil {
			t.Logf("'%s' did not fail for %#v", tc.Description, instance)
			t.Fail()
		}
	} else {
		if result != *canonicalExpectedPurl {
			t.Logf("%s: '%s' test failed: wanted: '%s', got '%s'", tc.Description, tc.TestType, *canonicalExpectedPurl, result)
			t.Fail()
		}
	}

}

func TestPurlSpecFixtures(t *testing.T) {
	testFiles, err := readJSONFilesFromDir("testdata/purl-spec/tests/types/")
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range testFiles {
		var suite TestSuite
		err := json.Unmarshal(file.content, &suite)
		if err != nil {
			t.Fatal(err)
		}

		for idx, tc := range suite.Tests {
			testName := fmt.Sprintf("%s[%d]%s", file.name, (idx + 1), tc.TestType)
			t.Run(testName, func(t *testing.T) {
				testType := tc.TestType

				switch testType {
				case "roundtrip":
					roundTripTest(tc, t)
				case "parse":
					parseTest(tc, t)
				case "build":
					buildTest(tc, t)
				default:
					t.Fatalf("Unsupported test type: %s", testType)
				}

			})

		}
	}
}

// Verify correct conversion of Qualifiers to a string map and vice versa.
func TestQualifiersMapConversion(t *testing.T) {
	tests := []struct {
		kvMap      map[string]string
		qualifiers packageurl.Qualifiers
	}{
		{
			kvMap:      map[string]string{},
			qualifiers: packageurl.Qualifiers{},
		},
		{
			kvMap: map[string]string{"arch": "amd64"},
			qualifiers: packageurl.Qualifiers{
				packageurl.Qualifier{Key: "arch", Value: "amd64"},
			},
		},
		{
			kvMap: map[string]string{"arch": "amd64", "os": "linux"},
			qualifiers: packageurl.Qualifiers{
				packageurl.Qualifier{Key: "arch", Value: "amd64"},
				packageurl.Qualifier{Key: "os", Value: "linux"},
			},
		},
	}

	for _, test := range tests {
		// map -> Qualifiers
		got := packageurl.QualifiersFromMap(test.kvMap)
		if !reflect.DeepEqual(got, test.qualifiers) {
			t.Logf("map -> qualifiers conversion failed: got: %#v, wanted: %#v", got, test.qualifiers)
			t.Fail()
		}

		// Qualifiers -> map
		mp := test.qualifiers.Map()
		if !reflect.DeepEqual(mp, test.kvMap) {
			t.Logf("qualifiers -> map conversion failed: got: %#v, wanted: %#v", mp, test.kvMap)
			t.Fail()
		}
	}
}

func TestNameEscaping(t *testing.T) {
	testCases := map[string]string{
		"abc":  "pkg:deb/abc",
		"ab/c": "pkg:deb/ab%2Fc",
	}
	for name, output := range testCases {
		t.Run(name, func(t *testing.T) {
			p := &packageurl.PackageURL{Type: "deb", Name: name}
			if s := p.ToString(); s != output {
				t.Fatalf("wrong escape. expected=%q, got=%q", output, s)
			}
		})
	}

}

func TestQualifierMissingEqual(t *testing.T) {
	input := "pkg:npm/test-pkg?key"
	want := packageurl.PackageURL{
		Type:       "npm",
		Name:       "test-pkg",
		Qualifiers: packageurl.Qualifiers{},
	}
	got, err := packageurl.FromString(input)
	if err != nil {
		t.Fatalf("FromString(%s): unexpected error: %v", input, err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("FromString(%s): want %q got %q", input, want, got)
	}
}

func TestNormalize(t *testing.T) {
	testCases := []struct {
		name    string
		input   packageurl.PackageURL
		want    packageurl.PackageURL
		wantErr bool
	}{{
		name: "type is case insensitive",
		input: packageurl.PackageURL{
			Type: "NpM",
			Name: "pkg",
		},
		want: packageurl.PackageURL{
			Type:       "npm",
			Name:       "pkg",
			Qualifiers: packageurl.Qualifiers{},
		},
	}, {
		name: "type is manditory",
		input: packageurl.PackageURL{
			Name: "pkg",
		},
		wantErr: true,
	}, {
		name: "leading and traling / on namespace are trimmed",
		input: packageurl.PackageURL{
			Type:      "npm",
			Namespace: "/namespace/org/",
			Name:      "pkg",
		},
		want: packageurl.PackageURL{
			Type:       "npm",
			Namespace:  "namespace/org",
			Name:       "pkg",
			Qualifiers: packageurl.Qualifiers{},
		},
	}, {
		name: "qualifiers with empty values are removed",
		input: packageurl.PackageURL{
			Type: "npm",
			Name: "pkg",
			Qualifiers: packageurl.Qualifiers{{
				Key: "k1", Value: "v1",
			}, {
				Key: "k2", Value: "",
			}, {
				Key: "k3", Value: "v3",
			}},
		},
		want: packageurl.PackageURL{
			Type: "npm",
			Name: "pkg",
			Qualifiers: packageurl.Qualifiers{{
				Key: "k1", Value: "v1",
			}, {
				Key: "k3", Value: "v3",
			}},
		},
	}, {
		name: "qualifiers are sorted by key",
		input: packageurl.PackageURL{
			Type: "npm",
			Name: "pkg",
			Qualifiers: packageurl.Qualifiers{{
				Key: "k3", Value: "v3",
			}, {
				Key: "k2", Value: "v2",
			}, {
				Key: "k1", Value: "v1",
			}},
		},
		want: packageurl.PackageURL{
			Type: "npm",
			Name: "pkg",
			Qualifiers: packageurl.Qualifiers{{
				Key: "k1", Value: "v1",
			}, {
				Key: "k2", Value: "v2",
			}, {
				Key: "k3", Value: "v3",
			}},
		},
	}, {
		name: "duplicate keys are invalid",
		input: packageurl.PackageURL{
			Type: "npm",
			Name: "pkg",
			Qualifiers: packageurl.Qualifiers{{
				Key: "k1", Value: "v1",
			}, {
				Key: "k1", Value: "v2",
			}},
		},
		wantErr: true,
	}, {
		name: "keys are made lower case",
		input: packageurl.PackageURL{
			Type: "npm",
			Name: "pkg",
			Qualifiers: packageurl.Qualifiers{{
				Key: "KeY", Value: "v1",
			}},
		},
		want: packageurl.PackageURL{
			Type: "npm",
			Name: "pkg",
			Qualifiers: packageurl.Qualifiers{{
				Key: "key", Value: "v1",
			}},
		},
	}, {
		name: "name is required",
		input: packageurl.PackageURL{
			Type: "npm",
		},
		wantErr: true,
	}, {
		name: "leading and traling / on subpath are trimmed",
		input: packageurl.PackageURL{
			Type:    "npm",
			Name:    "pkg",
			Subpath: "/sub/path/",
		},
		want: packageurl.PackageURL{
			Type:       "npm",
			Name:       "pkg",
			Qualifiers: packageurl.Qualifiers{},
			Subpath:    "sub/path",
		},
	}, {
		name: "'.' is an invalid subpath segment",
		input: packageurl.PackageURL{
			Type:    "npm",
			Name:    "pkg",
			Subpath: "/sub/./path/",
		},
		wantErr: true,
	}, {
		name: "'..' is an invalid subpath segment",
		input: packageurl.PackageURL{
			Type:    "npm",
			Name:    "pkg",
			Subpath: "/sub/../path/",
		},
		wantErr: true,
	}, {
		name: "'./' is a valid subpath prefix",
		input: packageurl.PackageURL{
			Type:    "npm",
			Name:    "pkg",
			Subpath: "./sub/path",
		},
		want: packageurl.PackageURL{
			Type:       "npm",
			Name:       "pkg",
			Qualifiers: packageurl.Qualifiers{},
			Subpath:    "./sub/path",
		},
	}, {
		name: "'../' is a valid subpath prefix",
		input: packageurl.PackageURL{
			Type:    "npm",
			Name:    "pkg",
			Subpath: "../sub/path",
		},
		want: packageurl.PackageURL{
			Type:       "npm",
			Name:       "pkg",
			Qualifiers: packageurl.Qualifiers{},
			Subpath:    "../sub/path",
		},
	}, {
		name: "known type namespace adjustments",
		input: packageurl.PackageURL{
			Type:      "apk",
			Namespace: "NaMeSpAcE",
			Name:      "pkg",
		},
		want: packageurl.PackageURL{
			Type:       "apk",
			Namespace:  "namespace",
			Name:       "pkg",
			Qualifiers: packageurl.Qualifiers{},
		},
	}, {
		name: "known type name adjustments",
		input: packageurl.PackageURL{
			Type: "alpm",
			Name: "nAmE",
		},
		want: packageurl.PackageURL{
			Type:       "alpm",
			Name:       "name",
			Qualifiers: packageurl.Qualifiers{},
		},
	}, {
		name: "known type version adjustments",
		input: packageurl.PackageURL{
			Type:    "huggingface",
			Name:    "name",
			Version: "VeRsIoN",
		},
		want: packageurl.PackageURL{
			Type:       "huggingface",
			Name:       "name",
			Version:    "version",
			Qualifiers: packageurl.Qualifiers{},
		},
	}}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.input
			err := got.Normalize()
			if err != nil && testCase.wantErr {
				return
			}
			if err != nil && !testCase.wantErr {
				t.Fatalf("Normalize(%s): unexpected error: %v", testCase.name, err)
			}
			if testCase.wantErr {
				t.Fatalf("Normalize(%s): want error, got none", testCase.name)
			}
			if !reflect.DeepEqual(testCase.want, got) {
				t.Fatalf("Normalize(%s):\nwant %#v\ngot %#v", testCase.name, testCase.want, got)
			}
		})
	}
}

// purlExpectation represents an expected canonical purl and destructured [packageurl.PackageURL]
// for a given input string.
type purlExpectation struct {
	// input represents a user-supplied purl input string to be parsed.
	input string

	// canonical is the expected canonical string represenation for input.
	canonical string
	// purl is the expected [packageurl.PackageURL] for input.
	purl packageurl.PackageURL
}

// TestRoundtrip is intended to cover some tricky purl parsing/canonicalization that is not covered
// by the purl-spec tests (yet).
func TestRoundtrip(t *testing.T) {
	tests := []struct {
		name        string
		expectation purlExpectation
	}{
		{
			name: "input version with unescaped slashes",
			expectation: purlExpectation{
				input:     "pkg:github/golang/mod@refs/tags/v0.30.0",
				canonical: "pkg:github/golang/mod@refs%2Ftags%2Fv0.30.0",
				purl: packageurl.PackageURL{
					Type:       packageurl.TypeGithub,
					Namespace:  "golang",
					Name:       "mod",
					Version:    "refs/tags/v0.30.0",
					Qualifiers: packageurl.Qualifiers{}}},
		},

		{
			name: "go modules can have vanity urls without namespace",
			expectation: purlExpectation{
				input:     "pkg:golang/go.opencensus.io@v0.20.1",
				canonical: "pkg:golang/go.opencensus.io@v0.20.1",
				purl: packageurl.PackageURL{
					Type:       packageurl.TypeGolang,
					Name:       "go.opencensus.io",
					Version:    "v0.20.1",
					Qualifiers: packageurl.Qualifiers{}}},
		},

		{
			name: "version with unescaped plus characters, qualifiers and subpath",
			expectation: purlExpectation{
				input:     "pkg:deb/debian/nuget@2.8.7+md510+dhx1-1.1?distro=stretch&repository_url=http://deb.debian.org&arch=amd64#subpath/to/file",
				canonical: "pkg:deb/debian/nuget@2.8.7%2Bmd510%2Bdhx1-1.1?arch=amd64&distro=stretch&repository_url=http:%2F%2Fdeb.debian.org#subpath/to/file",
				purl: packageurl.PackageURL{
					Type:      packageurl.TypeDebian,
					Namespace: "debian",
					Name:      "nuget",
					Version:   "2.8.7+md510+dhx1-1.1",
					Qualifiers: packageurl.Qualifiers{
						{Key: "arch", Value: "amd64"},
						{Key: "distro", Value: "stretch"},
						{Key: "repository_url", Value: "http://deb.debian.org"},
					},
					Subpath: "subpath/to/file"}},
		},

		{
			name: "npm package with unescaped @ in scope and no version",
			expectation: purlExpectation{
				input:     "pkg:npm/@opentelemetry/sdk-trace-node",
				canonical: "pkg:npm/%40opentelemetry/sdk-trace-node",
				purl: packageurl.PackageURL{
					Type:       packageurl.TypeNPM,
					Namespace:  "@opentelemetry",
					Name:       "sdk-trace-node",
					Qualifiers: packageurl.Qualifiers{}}},
		},

		{
			name: "npm package with unescaped @ in scope and version",
			expectation: purlExpectation{
				input:     "pkg:npm/@opentelemetry/sdk-trace-node@2.2.0",
				canonical: "pkg:npm/%40opentelemetry/sdk-trace-node@2.2.0",
				purl: packageurl.PackageURL{
					Type:       packageurl.TypeNPM,
					Namespace:  "@opentelemetry",
					Name:       "sdk-trace-node",
					Version:    "2.2.0",
					Qualifiers: packageurl.Qualifiers{}}},
		},

		// See https://github.com/package-url/purl-spec/discussions/814#discussioncomment-15837007
		{
			name: "interpret + character in qualifier as literal plus (not space)",
			expectation: purlExpectation{
				input:     "pkg:generic/grafana@12.0.1?checksum=sha256:18a348109d3f92772bee72a55eabb9d318596add6a70b92adb6ff8e789d587a8&download_url=https://dl.grafana.com/enterprise/release/grafana-enterprise-12.0.1+security-01.linux-amd64.tar.gz",
				canonical: "pkg:generic/grafana@12.0.1?checksum=sha256:18a348109d3f92772bee72a55eabb9d318596add6a70b92adb6ff8e789d587a8&download_url=https:%2F%2Fdl.grafana.com%2Fenterprise%2Frelease%2Fgrafana-enterprise-12.0.1%2Bsecurity-01.linux-amd64.tar.gz",
				purl: packageurl.PackageURL{
					Type:    packageurl.TypeGeneric,
					Name:    "grafana",
					Version: "12.0.1",
					Qualifiers: packageurl.Qualifiers{
						{Key: "checksum", Value: "sha256:18a348109d3f92772bee72a55eabb9d318596add6a70b92adb6ff8e789d587a8"},
						{Key: "download_url", Value: "https://dl.grafana.com/enterprise/release/grafana-enterprise-12.0.1+security-01.linux-amd64.tar.gz"},
					}}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := packageurl.FromString(tc.expectation.input)
			if err != nil {
				t.Fatalf("FromString(%s) unexpectedly failed: %v", tc.expectation.input, err)
			}
			if !reflect.DeepEqual(tc.expectation.purl, got) {
				t.Fatalf("FromString(%s):\nwanted: %#v\ngot: %#v", tc.expectation.input, tc.expectation.purl, got)
			}

			if got.String() != tc.expectation.canonical {
				t.Fatalf("String(%s):\nwanted: %s\ngot: %s", tc.expectation.input, tc.expectation.canonical, got.String())
			}
		})
	}
}
