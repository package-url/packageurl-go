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

// Package packageurl implements the package-url spec
package packageurl

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"
)

var (
	// QualifierKeyPattern describes a valid qualifier key:
	//
	// - The key must be composed only of ASCII letters and numbers, '.',
	//   '-' and '_' (period, dash and underscore).
	// - A key cannot start with a number.
	QualifierKeyPattern = regexp.MustCompile(`^[A-Za-z\.\-_][0-9A-Za-z\.\-_]*$`)
)

// These are the known purl types as defined in the spec. Some of these require
// special treatment during parsing.
// https://github.com/package-url/purl-spec#known-purl-types
var (
	// TypeAlpm is a pkg:alpm purl.
	TypeAlpm = "alpm"
	// TypeApk is a pkg:apk purl.
	TypeApk = "apk"
	// TypeBitbucket is a pkg:bitbucket purl.
	TypeBitbucket = "bitbucket"
	// TypeCocoapods is a pkg:cocoapods purl.
	TypeCocoapods = "cocoapods"
	// TypeCargo is a pkg:cargo purl.
	TypeCargo = "cargo"
	// TypeComposer is a pkg:composer purl.
	TypeComposer = "composer"
	// TypeConan is a pkg:conan purl.
	TypeConan = "conan"
	// TypeConda is a pkg:conda purl.
	TypeConda = "conda"
	// TypeCran is a pkg:cran purl.
	TypeCran = "cran"
	// TypeDebian is a pkg:deb purl.
	TypeDebian = "deb"
	// TypeDocker is a pkg:docker purl.
	TypeDocker = "docker"
	// TypeGem is a pkg:gem purl.
	TypeGem = "gem"
	// TypeGeneric is a pkg:generic purl.
	TypeGeneric = "generic"
	// TypeGithub is a pkg:github purl.
	TypeGithub = "github"
	// TypeGolang is a pkg:golang purl.
	TypeGolang = "golang"
	// TypeHackage is a pkg:hackage purl.
	TypeHackage = "hackage"
	// TypeHex is a pkg:hex purl.
	TypeHex = "hex"
	// TypeMaven is a pkg:maven purl.
	TypeMaven = "maven"
	// TypeNPM is a pkg:npm purl.
	TypeNPM = "npm"
	// TypeNuget is a pkg:nuget purl.
	TypeNuget = "nuget"
	// TypeQPKG is a pkg:qpkg purl.
	TypeQpkg = "qpkg"
	// TypeOCI is a pkg:oci purl
	TypeOCI = "oci"
	// TypePyPi is a pkg:pypi purl.
	TypePyPi = "pypi"
	// TypeRPM is a pkg:rpm purl.
	TypeRPM = "rpm"
	// TypeSWID is pkg:swid purl
	TypeSWID = "swid"
	// TypeSwift is pkg:swift purl
	TypeSwift = "swift"
	// TypeHuggingface is pkg:huggingface purl.
	TypeHuggingface = "huggingface"
	// TypeMLflow is pkg:mlflow purl.
	TypeMLFlow = "mlflow"
	// TypeJulia is a pkg:julia purl
	TypeJulia = "julia"
)

// Qualifier represents a single key=value qualifier in the package url
type Qualifier struct {
	Key   string
	Value string
}

func (q Qualifier) String() string {
	// A value must be a percent-encoded string
	return fmt.Sprintf("%s=%s", q.Key, url.PathEscape(q.Value))
}

// Qualifiers is a slice of key=value pairs, with order preserved as it appears
// in the package URL.
type Qualifiers []Qualifier

// urlQuery returns a raw URL query with all the qualifiers as keys + values.
func (q Qualifiers) urlQuery() (rawQuery string) {
	v := make(url.Values)
	for _, qq := range q {
		v.Add(qq.Key, qq.Value)
	}
	return v.Encode()
}

// QualifiersFromMap constructs a Qualifiers slice from a string map. To get a
// deterministic qualifier order (despite maps not providing any iteration order
// guarantees) the returned Qualifiers are sorted in increasing order of key.
func QualifiersFromMap(mm map[string]string) Qualifiers {
	q := Qualifiers{}

	for k, v := range mm {
		q = append(q, Qualifier{Key: k, Value: v})
	}

	// sort for deterministic qualifier order
	sort.Slice(q, func(i int, j int) bool { return q[i].Key < q[j].Key })

	return q
}

// Map converts a Qualifiers struct to a string map.
func (qq Qualifiers) Map() map[string]string {
	m := make(map[string]string)

	for i := 0; i < len(qq); i++ {
		k := qq[i].Key
		v := qq[i].Value
		m[k] = v
	}

	return m
}

func (qq Qualifiers) String() string {
	var kvPairs []string
	for _, q := range qq {
		kvPairs = append(kvPairs, q.String())
	}
	return strings.Join(kvPairs, "&")
}

// PackageURL is the struct representation of the parts that make a package url
type PackageURL struct {
	Type       string
	Namespace  string
	Name       string
	Version    string
	Qualifiers Qualifiers
	Subpath    string
}

// NewPackageURL creates a new PackageURL struct instance based on input
func NewPackageURL(purlType, namespace, name, version string,
	qualifiers Qualifiers, subpath string) *PackageURL {

	return &PackageURL{
		Type:       purlType,
		Namespace:  namespace,
		Name:       name,
		Version:    version,
		Qualifiers: qualifiers,
		Subpath:    subpath,
	}
}

// ToString returns the human-readable instance of the PackageURL structure.
// This is the literal purl as defined by the spec.
func (p *PackageURL) ToString() string {
	u := &url.URL{
		Scheme:   "pkg",
		RawQuery: p.Qualifiers.urlQuery(),
		Fragment: p.Subpath,
	}

	// we use JoinPath and EscapePath as the behavior for "/" is only correct with that.
	// We don't want to escape "/", but want to escape all other characters that are necessary.
	u = u.JoinPath(p.Type, p.Namespace, strings.Join([]string{p.Name, p.Version}, "@"))
	// write the actual path into the "Opaque" block, so that the generated string at the end is
	// pkg:<path> and not pkg://<path>.
	u.Opaque, u.Path = u.EscapedPath(), ""

	return u.String()
}

func (p PackageURL) String() string {
	return p.ToString()
}

// FromString parses a valid package url string into a PackageURL structure
func FromString(purl string) (PackageURL, error) {
	u, err := url.Parse(purl)
	if err != nil {
		return PackageURL{}, fmt.Errorf("failed to parse as URL: %w", err)
	}

	if u.Scheme != "pkg" {
		return PackageURL{}, fmt.Errorf("purl scheme is not \"pkg\": %q", u.Scheme)
	}

	p := u.Opaque
	// if a purl starts with pkg:/ or even pkg://, we need to fall back to host + path.
	if p == "" {
		p = strings.TrimPrefix(path.Join(u.Host, u.Path), "/")
	}

	typ, p, ok := strings.Cut(p, "/")
	if !ok {
		return PackageURL{}, fmt.Errorf("purl is missing type or name")
	}
	typ = strings.ToLower(typ)

	qualifiers, err := parseQualifiers(u.RawQuery)
	if err != nil {
		return PackageURL{}, fmt.Errorf("invalid qualifiers: %w", err)
	}
	namespace, name, version, err := separateNamespaceNameVersion(p)
	if err != nil {
		return PackageURL{}, err
	}

	pURL := PackageURL{
		Qualifiers: qualifiers,
		Type:       typ,
		Namespace:  typeAdjustNamespace(typ, namespace),
		Name:       typeAdjustName(typ, name, qualifiers),
		Version:    typeAdjustVersion(typ, version),
		Subpath:    strings.Trim(u.Fragment, "/"),
	}

	return pURL, validCustomRules(pURL)
}

func separateNamespaceNameVersion(path string) (ns, name, version string, err error) {
	namespaceSep := strings.LastIndex(path, "/")
	if namespaceSep != -1 {
		ns, err = url.PathUnescape(path[:namespaceSep])
		if err != nil {
			return "", "", "", fmt.Errorf("error unescaping namespace: %w", err)
		}

		path = path[namespaceSep+1:]
	}

	v := strings.Split(path, "@")
	name, err = url.PathUnescape(v[0])
	if err != nil {
		return "", "", "", fmt.Errorf("error unescaping name: %w", err)
	}

	if name == "" {
		return "", "", "", fmt.Errorf("purl is missing name")
	}

	if len(v) > 1 {
		version, err = url.PathUnescape(v[1])
		if err != nil {
			return "", "", "", fmt.Errorf("error unescaping version: %w", err)
		}
	}

	return ns, name, version, nil
}

func parseQualifiers(rawQuery string) (Qualifiers, error) {
	// we need to parse the qualifiers ourselves and cannot rely on the `url.Query` type because
	// that uses a map, meaning it's unordered. We want to keep the order of the qualifiers, so this
	// function re-implements the `url.parseQuery` function based on our `Qualifier` type. Most of
	// the code here is taken from `url.parseQuery`.
	q := Qualifiers{}
	for rawQuery != "" {
		var key string
		key, rawQuery, _ = strings.Cut(rawQuery, "&")
		if strings.Contains(key, ";") {
			return nil, fmt.Errorf("invalid semicolon separator in query")
		}
		if key == "" {
			continue
		}
		key, value, _ := strings.Cut(key, "=")
		key, err := url.QueryUnescape(key)
		if err != nil {
			return nil, fmt.Errorf("error unescaping qualifier key %q", key)
		}

		if !validQualifierKey(key) {
			return nil, fmt.Errorf("invalid qualifier key: '%s'", key)
		}

		value, err = url.QueryUnescape(value)
		if err != nil {
			return nil, fmt.Errorf("error unescaping qualifier value %q", value)
		}

		q = append(q, Qualifier{
			Key: strings.ToLower(key),
			// only the first character needs  to be lowercase. Note that pURL is always UTF8, so we
			// don't need to care about unicode here.
			Value: strings.ToLower(value[:1]) + value[1:],
		})
	}
	return q, nil
}

// Make any purl type-specific adjustments to the parsed namespace.
// See https://github.com/package-url/purl-spec#known-purl-types
func typeAdjustNamespace(purlType, ns string) string {
	switch purlType {
	case TypeAlpm,
		TypeApk,
		TypeBitbucket,
		TypeComposer,
		TypeDebian,
		TypeGithub,
		TypeGolang,
		TypeNPM,
		TypeRPM,
		TypeQpkg:
		return strings.ToLower(ns)
	}
	return ns
}

// Make any purl type-specific adjustments to the parsed name.
// See https://github.com/package-url/purl-spec#known-purl-types
func typeAdjustName(purlType, name string, qualifiers Qualifiers) string {
	quals := qualifiers.Map()
	switch purlType {
	case TypeAlpm,
		TypeApk,
		TypeBitbucket,
		TypeComposer,
		TypeDebian,
		TypeGithub,
		TypeGolang,
		TypeNPM:
		return strings.ToLower(name)
	case TypePyPi:
		return strings.ToLower(strings.ReplaceAll(name, "_", "-"))
	case TypeMLFlow:
		return adjustMlflowName(name, quals)
	}
	return name
}

// Make any purl type-specific adjustments to the parsed version.
// See https://github.com/package-url/purl-spec#known-purl-types
func typeAdjustVersion(purlType, version string) string {
	switch purlType {
	case TypeHuggingface:
		return strings.ToLower(version)
	}
	return version
}

// https://github.com/package-url/purl-spec/blob/master/PURL-TYPES.rst#mlflow
func adjustMlflowName(name string, qualifiers map[string]string) string {
	if repo, ok := qualifiers["repository_url"]; ok {
		if strings.Contains(repo, "azureml") {
			// Azure ML is case-sensitive and must be kept as-is
			return name
		} else if strings.Contains(repo, "databricks") {
			// Databricks is case-insensitive and must be lowercased
			return strings.ToLower(name)
		} else {
			// Unknown repository type, keep as-is
			return name
		}
	} else {
		// No repository qualifier given, keep as-is
		return name
	}
}

// validQualifierKey validates a qualifierKey against our QualifierKeyPattern.
func validQualifierKey(key string) bool {
	return QualifierKeyPattern.MatchString(key)
}

// validCustomRules evaluates additional rules for each package url type, as specified in the package-url specification.
// On success, it returns nil. On failure, a descriptive error will be returned.
func validCustomRules(p PackageURL) error {
	q := p.Qualifiers.Map()
	switch p.Type {
	case TypeConan:
		if p.Namespace != "" {
			if val, ok := q["channel"]; ok {
				if val == "" {
					return errors.New("the qualifier channel must be not empty if namespace is present")
				}
			} else {
				return errors.New("channel qualifier does not exist")
			}
		} else {
			if val, ok := q["channel"]; ok {
				if val != "" {
					return errors.New("namespace is required if channel is non empty")
				}
			}
		}
	case TypeSwift:
		if p.Namespace == "" {
			return errors.New("namespace is required")
		}
		if p.Version == "" {
			return errors.New("version is required")
		}
	case TypeCran:
		if p.Version == "" {
			return errors.New("version is required")
		}
	}
	return nil
}
