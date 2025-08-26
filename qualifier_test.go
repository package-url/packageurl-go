package packageurl

import "testing"

// Verifies that qualifier values are properly percent-encoded.
func TestQualifierValueEncoding(t *testing.T) {
	tests := []struct {
		name  string
		given PackageURL
		want  string
	}{
		{
			name: "must percent-encode a qualifier value",
			given: PackageURL{
				Type:    "generic",
				Name:    "openssl",
				Version: "1.2.3",
				Qualifiers: Qualifiers{
					{Key: "download_url", Value: "dl.site.org/openssl"},
				},
			},
			want: "pkg:generic/openssl@1.2.3?download_url=dl.site.org%2Fopenssl",
		},

		{
			name: "must not percent-encode colons in qualifier values",
			given: PackageURL{
				Type:    "generic",
				Name:    "openssl",
				Version: "1.2.3",
				Qualifiers: Qualifiers{
					{Key: "download_url", Value: "dl.site.org:443/openssl"},
				},
			},
			want: "pkg:generic/openssl@1.2.3?download_url=dl.site.org:443%2Fopenssl",
		},

		// Note: checks that the use of [url.QueryEscape] does not lead to a space being
		// encoded as "+".
		{
			name: "must properly percent-encode a plus",
			given: PackageURL{
				Type:    "generic",
				Name:    "openssl",
				Version: "1.2.3",
				Qualifiers: Qualifiers{
					{Key: "download_url", Value: "dl.site.org:443/openssl secure"},
				},
			},
			want: "pkg:generic/openssl@1.2.3?download_url=dl.site.org:443%2Fopenssl%20secure",
		},

		{
			name: "must order qualifier values lexicographically by key",
			given: PackageURL{
				Type:    "generic",
				Name:    "openssl",
				Version: "1.2.3",
				Qualifiers: Qualifiers{
					{Key: "download_url", Value: "dl.site.org"},
					{Key: "checksum", Value: "sha256:abc123"},
				},
			},
			want: "pkg:generic/openssl@1.2.3?checksum=sha256:abc123&download_url=dl.site.org",
		},

		{
			name: "must percent-encode separators in qualifier value",
			given: PackageURL{
				Type:    "generic",
				Name:    "openssl",
				Version: "1.2.3",
				Qualifiers: Qualifiers{
					{Key: "download_url", Value: "https://dl.site.org:443/openssl+secure?type=zip&fast=yes"},
					{Key: "checksum", Value: "sha256:abc123"},
				},
			},
			want: "pkg:generic/openssl@1.2.3?checksum=sha256:abc123&download_url=https:%2F%2Fdl.site.org:443%2Fopenssl%2Bsecure%3Ftype%3Dzip%26fast%3Dyes",
		},

		{
			name: "must escape a qualifier value",
			given: PackageURL{
				Type:    "generic",
				Name:    "openssl",
				Version: "1.2.3",
				Qualifiers: Qualifiers{
					{Key: "download_url", Value: "dl.site.org/openssl"},
				},
			},
			want: "pkg:generic/openssl@1.2.3?download_url=dl.site.org%2Fopenssl",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.given.String()
			if tc.want != got {
				t.Logf("'%s' test failed: wanted: '%s', got '%s'", tc.name, tc.want, got)
				t.Fail()
			}
		})
	}
}

// Exercise the [encodeQualifierValue] function.
func TestEncodeQualifierValue(t *testing.T) {
	tests := []struct {
		name       string
		givenValue string
		want       string
	}{
		{
			name:       "must not percent-encode alphanumeric characters",
			givenValue: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
			want:       "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
		},

		{
			name:       "must not percent-encode punctuation characters",
			givenValue: ".-_~",
			want:       ".-_~",
		},

		{
			name:       "must not percent-encode colon",
			givenValue: ":",
			want:       ":",
		},

		{
			name:       "must percent-encode separator characters other than colon",
			givenValue: "/@?=&#",
			// slash: %2F, at: %40, question mark: %3F, equal: %3D, ampersand: %26,
			// pound-sign: %23
			want: "%2F%40%3F%3D%26%23",
		},

		{
			name:       "must percent-encode a plus",
			givenValue: "+",
			want:       "%2B",
		},

		{
			name:       "must percent-encode a whitespace",
			givenValue: " ",
			want:       "%20",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := encodeQualifierValue(tc.givenValue)
			if tc.want != got {
				t.Logf("'%s' test failed: wanted: '%s', got '%s'", tc.name, tc.want, got)
				t.Fail()
			}
		})
	}
}
