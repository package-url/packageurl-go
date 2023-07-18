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
	"os"
	"reflect"
	"testing"

	"github.com/package-url/packageurl-go"
)

type TestFixture struct {
	Description   string            `json:"description"`
	Purl          string            `json:"purl"`
	CanonicalPurl string            `json:"canonical_purl"`
	PackageType   string            `json:"type"`
	Namespace     string            `json:"namespace"`
	Name          string            `json:"name"`
	Version       string            `json:"version"`
	QualifierMap  map[string]string `json:"qualifiers"`
	Subpath       string            `json:"subpath"`
	IsInvalid     bool              `json:"is_invalid"`
}

func (t TestFixture) Qualifiers() packageurl.Qualifiers {
	q := packageurl.Qualifiers{}
	for k, v := range t.QualifierMap {
		q.Add(k, v)
	}
	return q
}

// TestFromStringExamples verifies that parsing example strings produce expected
// results.
func TestFromStringExamples(t *testing.T) {
	// Read the json file
	data, err := os.ReadFile("testdata/test-suite-data.json")
	if err != nil {
		t.Fatal(err)
	}
	// Load the json file contents into a structure
	testData := []TestFixture{}
	err = json.Unmarshal(data, &testData)
	if err != nil {
		t.Fatal(err)
	}

	// Use FromString on each item in the test set
	for _, tc := range testData {
		// Should parse without issue
		p, err := packageurl.FromString(tc.Purl)
		if tc.IsInvalid == false {
			if err != nil {
				t.Logf("%s failed: %s", tc.Description, err)
				t.Fail()
			}
			// verify parsing
			if p.Type != tc.PackageType {
				t.Logf("%s: incorrect package type: wanted: '%s', got '%s'", tc.Description, tc.PackageType, p.Type)
				t.Fail()
			}
			if p.Namespace != tc.Namespace {
				t.Logf("%s: incorrect namespace: wanted: '%s', got '%s'", tc.Description, tc.Namespace, p.Namespace)
				t.Fail()
			}
			if p.Name != tc.Name {
				t.Logf("%s: incorrect name: wanted: '%s', got '%s'", tc.Description, tc.Name, p.Name)
				t.Fail()
			}
			if p.Version != tc.Version {
				t.Logf("%s: incorrect version: wanted: '%s', got '%s'", tc.Description, tc.Version, p.Version)
				t.Fail()
			}
			if !reflect.DeepEqual(p.Qualifiers, tc.Qualifiers()) {
				t.Logf("%s: incorrect qualifiers: wanted: '%#v', got '%#v'", tc.Description, tc.Qualifiers(), p.Qualifiers)
				t.Fail()
			}

			if p.Subpath != tc.Subpath {
				t.Logf("%s: incorrect subpath: wanted: '%s', got '%s'", tc.Description, tc.Subpath, p.Subpath)
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
}

// TestToStringExamples verifies that the resulting package urls created match
// the expected format.
func TestToStringExamples(t *testing.T) {
	// Read the json file
	data, err := os.ReadFile("testdata/test-suite-data.json")
	if err != nil {
		t.Fatal(err)
	}
	// Load the json file contents into a structure
	var testData []TestFixture
	err = json.Unmarshal(data, &testData)
	if err != nil {
		t.Fatal(err)
	}
	// Use ToString on each item
	for _, tc := range testData {
		// Skip invalid items
		if tc.IsInvalid == true {
			continue
		}
		instance := packageurl.NewPackageURL(
			tc.PackageType, tc.Namespace, tc.Name, tc.Version,
			tc.Qualifiers(), tc.Subpath)
		result := instance.ToString()

		// NOTE: We create a purl with ToString and then load into a PackageURL
		//       because qualifiers may not be in any order. By reparsing back
		//       we can ensure the data transfers between string and instance form.
		canonical, _ := packageurl.FromString(tc.CanonicalPurl)
		toTest, _ := packageurl.FromString(result)
		// If the two results don't equal then the ToString failed
		if !reflect.DeepEqual(toTest, canonical) {
			t.Logf("%s failed: %s != %s", tc.Description, result, tc.CanonicalPurl)
			t.Fail()
		}
	}
}

// TestStringer verifies that the Stringer implementation produces results
// equivalent with the ToString method.
func TestStringer(t *testing.T) {
	// Read the json file
	data, err := os.ReadFile("testdata/test-suite-data.json")
	if err != nil {
		t.Fatal(err)
	}
	// Load the json file contents into a structure
	var testData []TestFixture
	err = json.Unmarshal(data, &testData)
	if err != nil {
		t.Fatal(err)
	}
	// Use ToString on each item
	for _, tc := range testData {
		// Skip invalid items
		if tc.IsInvalid == true {
			continue
		}
		purlPtr := packageurl.NewPackageURL(
			tc.PackageType, tc.Namespace, tc.Name,
			tc.Version, tc.Qualifiers(), tc.Subpath)
		purlValue := *purlPtr

		// Verify that the Stringer implementation returns a result
		// equivalent to ToString().
		if purlPtr.ToString() != purlPtr.String() {
			t.Logf("%s failed: Stringer implementation differs from ToString: %s != %s", tc.Description, purlPtr.String(), purlPtr.ToString())
			t.Fail()
		}

		// Verify that the %s format modifier works for values.
		fmtStr := purlValue.String()
		if fmtStr != purlPtr.String() {
			t.Logf("%s failed: %%s format modifier does not work on values: %s != %s", tc.Description, fmtStr, purlPtr.ToString())
			t.Fail()
		}
	}
}

func TestNameEscaping(t *testing.T) {
	testCases := map[string]string{
		"abc":  "pkg:abc",
		"ab/c": "pkg:ab%2Fc",
	}
	for name, output := range testCases {
		t.Run(name, func(t *testing.T) {
			p := &packageurl.PackageURL{Name: name}
			if s := p.ToString(); s != output {
				t.Fatalf("wrong escape. expected=%q, got=%q", output, s)
			}
		})
	}

}
