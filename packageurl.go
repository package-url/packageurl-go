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
	"strings"
)

// Qualifiers houses each key=value pair in the package url
type Qualifiers map[string]string

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

// ToString returns the human readable instance of the PackageURL structure.
// This is the literal purl as defined by the spec.
func (p *PackageURL) ToString() string {
	// Start with the type and a colon
	purl := fmt.Sprintf("pkg:%s/", p.Type)
	// Add namespaces if provided
	if p.Namespace != "" {
		ns := []string{}
		for _, item := range strings.Split(p.Namespace, "/") {
			ns = append(ns, url.QueryEscape(item))
		}
		purl = purl + strings.Join(ns, "/") + "/"
	}
	// The name is always required
	purl = purl + p.Name
	// If a version is provided, add it after the at symbol
	if p.Version != "" {
		purl = purl + "@" + p.Version
	}

	// Iterate over qualifiers and make groups of key=value
	qualifiers := []string{}
	for k, v := range p.Qualifiers {
		qualifiers = append(qualifiers, fmt.Sprintf("%s=%s", k, v))
	}
	// If there one or more key=value pairs then append on the package url
	if len(qualifiers) != 0 {
		purl = purl + "?" + strings.Join(qualifiers, "&")
	}
	// Add a subpath if available
	if p.Subpath != "" {
		purl = purl + "#" + p.Subpath
	}
	return purl
}

// FromString parses a valid package url string into a PackageURL structure
func FromString(purl string) (PackageURL, error) {
	initialIndex := strings.Index(purl, "#")
	// Start with purl being stored in the remainder
	remainder := purl
	substring := ""
	if initialIndex != -1 {
		initialSplit := strings.SplitN(purl, "#", 2)
		remainder = initialSplit[0]
		rightSide := initialSplit[1]
		rightSide = strings.TrimLeft(rightSide, "/")
		rightSide = strings.TrimRight(rightSide, "/")
		var rightSides []string

		for _, item := range strings.Split(rightSide, "/") {
			item = strings.Replace(item, ".", "", -1)
			item = strings.Replace(item, "..", "", -1)
			if item != "" {
				i, err := url.PathUnescape(item)
				if err != nil {
					return PackageURL{}, fmt.Errorf("failed to unescape path: %s", err)
				}
				rightSides = append(rightSides, i)
			}
		}
		substring = strings.Join(rightSides, "")
	}
	qualifiers := Qualifiers{}
	index := strings.LastIndex(remainder, "?")
	// If we don't have anything to split then return an empty result
	if index != -1 {
		qualifier := remainder[index+1:]
		for _, item := range strings.Split(qualifier, "&") {
			kv := strings.Split(item, "=")
			key := strings.ToLower(kv[0])
			// TODO
			//  - If the `key` is `checksums`, split the `value` on ',' to create
			//    a list of `checksums`
			if kv[1] == "" {
				continue
			}
			value, err := url.PathUnescape(kv[1])
			if err != nil {
				return PackageURL{}, fmt.Errorf("failed to unescape path: %s", err)
			}
			qualifiers[key] = value
		}
		remainder = remainder[:index]
	}

	nextSplit := strings.SplitN(remainder, ":", 2)
	if len(nextSplit) != 2 || nextSplit[0] != "pkg" {
		return PackageURL{}, errors.New("scheme is missing")
	}
	remainder = nextSplit[1]

	nextSplit = strings.SplitN(remainder, "/", 2)
	if len(nextSplit) != 2 {
		return PackageURL{}, errors.New("type is missing")
	}
	purlType := nextSplit[0]
	remainder = nextSplit[1]

	index = strings.LastIndex(remainder, "/")
	name := remainder[index+1:]
	version := ""

	atIndex := strings.Index(name, "@")
	if atIndex != -1 {
		version = name[atIndex+1:]
		name = name[:atIndex]
	}
	namespaces := []string{}

	if index != -1 {
		remainder = remainder[:index]

		for _, item := range strings.Split(remainder, "/") {
			if item != "" {
				unescaped, err := url.PathUnescape(item)
				if err != nil {
					return PackageURL{}, fmt.Errorf("failed to unescape path: %s", err)
				}
				namespaces = append(namespaces, unescaped)
			}
		}
	}

	// Fail if name is empty at this point
	if name == "" {
		return PackageURL{}, errors.New("name is required")
	}

	return PackageURL{
		Type:       purlType,
		Namespace:  strings.Join(namespaces, "/"),
		Name:       name,
		Version:    version,
		Qualifiers: qualifiers,
		Subpath:    substring,
	}, nil
}
