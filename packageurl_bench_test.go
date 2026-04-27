package packageurl

import (
	"testing"
)

// Sample purls of varying complexity
var (
	simplePurl     = "pkg:npm/lodash@4.17.21"
	namespacePurl  = "pkg:maven/org.apache.commons/commons-lang3@3.12.0"
	qualifiersPurl = "pkg:npm/%40angular/core@16.0.0?repository_url=https://registry.npmjs.org"
	complexPurl    = "pkg:rpm/fedora/curl@7.50.3-1.fc25?arch=i386&distro=fedora-25&repository_url=http://example.com"
	subpathPurl    = "pkg:github/package-url/purl-spec@244fd47e07d1004f0aed9c#src/main/java"
	fullPurl       = "pkg:deb/debian/dpkg@1.19.0.4?arch=amd64&distro=stretch&repository_url=http://deb.debian.org#subpath/to/file"
)

// Pre-parsed PackageURL structs for ToString benchmarks
var (
	simplePackageURL = PackageURL{
		Type:    "npm",
		Name:    "lodash",
		Version: "4.17.21",
	}
	namespacePackageURL = PackageURL{
		Type:      "maven",
		Namespace: "org.apache.commons",
		Name:      "commons-lang3",
		Version:   "3.12.0",
	}
	qualifiersPackageURL = PackageURL{
		Type:      "npm",
		Namespace: "@angular",
		Name:      "core",
		Version:   "16.0.0",
		Qualifiers: Qualifiers{
			{Key: "repository_url", Value: "https://registry.npmjs.org"},
		},
	}
	complexPackageURL = PackageURL{
		Type:      "rpm",
		Namespace: "fedora",
		Name:      "curl",
		Version:   "7.50.3-1.fc25",
		Qualifiers: Qualifiers{
			{Key: "arch", Value: "i386"},
			{Key: "distro", Value: "fedora-25"},
			{Key: "repository_url", Value: "http://example.com"},
		},
	}
	fullPackageURL = PackageURL{
		Type:      "deb",
		Namespace: "debian",
		Name:      "dpkg",
		Version:   "1.19.0.4",
		Qualifiers: Qualifiers{
			{Key: "arch", Value: "amd64"},
			{Key: "distro", Value: "stretch"},
			{Key: "repository_url", Value: "http://deb.debian.org"},
		},
		Subpath: "subpath/to/file",
	}
)

// FromString benchmarks - parsing
func BenchmarkFromString_Simple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = FromString(simplePurl)
	}
}

func BenchmarkFromString_Namespace(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = FromString(namespacePurl)
	}
}

func BenchmarkFromString_Qualifiers(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = FromString(qualifiersPurl)
	}
}

func BenchmarkFromString_Complex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = FromString(complexPurl)
	}
}

func BenchmarkFromString_Subpath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = FromString(subpathPurl)
	}
}

func BenchmarkFromString_Full(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = FromString(fullPurl)
	}
}

// ToString benchmarks - serialization
func BenchmarkToString_Simple(b *testing.B) {
	p := simplePackageURL
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.ToString()
	}
}

func BenchmarkToString_Namespace(b *testing.B) {
	p := namespacePackageURL
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.ToString()
	}
}

func BenchmarkToString_Qualifiers(b *testing.B) {
	p := qualifiersPackageURL
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.ToString()
	}
}

func BenchmarkToString_Complex(b *testing.B) {
	p := complexPackageURL
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.ToString()
	}
}

func BenchmarkToString_Full(b *testing.B) {
	p := fullPackageURL
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.ToString()
	}
}

// Normalize benchmarks
func BenchmarkNormalize_Simple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := simplePackageURL
		_ = p.Normalize()
	}
}

func BenchmarkNormalize_Complex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := complexPackageURL
		_ = p.Normalize()
	}
}

func BenchmarkNormalize_Full(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := fullPackageURL
		_ = p.Normalize()
	}
}

// Roundtrip benchmarks (parse + serialize)
func BenchmarkRoundtrip_Simple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p, _ := FromString(simplePurl)
		_ = p.ToString()
	}
}

func BenchmarkRoundtrip_Complex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p, _ := FromString(complexPurl)
		_ = p.ToString()
	}
}

func BenchmarkRoundtrip_Full(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p, _ := FromString(fullPurl)
		_ = p.ToString()
	}
}

// Qualifier operations
func BenchmarkQualifiersFromMap_Small(b *testing.B) {
	m := map[string]string{
		"arch": "amd64",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = QualifiersFromMap(m)
	}
}

func BenchmarkQualifiersFromMap_Medium(b *testing.B) {
	m := map[string]string{
		"arch":           "amd64",
		"distro":         "debian-10",
		"repository_url": "http://example.com",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = QualifiersFromMap(m)
	}
}

func BenchmarkQualifiersFromMap_Large(b *testing.B) {
	m := map[string]string{
		"arch":           "amd64",
		"distro":         "debian-10",
		"repository_url": "http://example.com",
		"checksum":       "sha256:abc123",
		"vcs_url":        "git+https://github.com/foo/bar",
		"download_url":   "https://example.com/download",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = QualifiersFromMap(m)
	}
}

func BenchmarkQualifiers_Map_Small(b *testing.B) {
	q := Qualifiers{{Key: "arch", Value: "amd64"}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = q.Map()
	}
}

func BenchmarkQualifiers_Map_Medium(b *testing.B) {
	q := Qualifiers{
		{Key: "arch", Value: "amd64"},
		{Key: "distro", Value: "debian-10"},
		{Key: "repository_url", Value: "http://example.com"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = q.Map()
	}
}

func BenchmarkQualifiers_Map_Large(b *testing.B) {
	q := Qualifiers{
		{Key: "arch", Value: "amd64"},
		{Key: "distro", Value: "debian-10"},
		{Key: "repository_url", Value: "http://example.com"},
		{Key: "checksum", Value: "sha256:abc123"},
		{Key: "vcs_url", Value: "git+https://github.com/foo/bar"},
		{Key: "download_url", Value: "https://example.com/download"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = q.Map()
	}
}

// Qualifier String/Normalize
func BenchmarkQualifiers_String(b *testing.B) {
	q := complexPackageURL.Qualifiers
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = q.String()
	}
}

func BenchmarkQualifiers_Normalize(b *testing.B) {
	for i := 0; i < b.N; i++ {
		q := Qualifiers{
			{Key: "Arch", Value: "amd64"},
			{Key: "DISTRO", Value: "debian-10"},
			{Key: "repository_url", Value: "http://example.com"},
		}
		_ = q.Normalize()
	}
}

// Validation benchmarks
func BenchmarkValidQualifierKey(b *testing.B) {
	keys := []string{"arch", "repository_url", "vcs_url", "checksum.sha256"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, k := range keys {
			_ = validQualifierKey(k)
		}
	}
}

func BenchmarkValidType(b *testing.B) {
	types := []string{"npm", "maven", "pypi", "golang", "deb", "rpm"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, t := range types {
			_ = validType(t)
		}
	}
}
