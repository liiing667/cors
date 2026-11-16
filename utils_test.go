package cors

import (
	"net/http"
	"slices"
	"strings"
	"testing"
)

func TestWildcard(t *testing.T) {
	w := wildcard{"foo", "bar"}
	if !w.match("foobar") {
		t.Error("foo*bar should match foobar")
	}
	if !w.match("foobazbar") {
		t.Error("foo*bar should match foobazbar")
	}
	if w.match("foobaz") {
		t.Error("foo*bar should not match foobaz")
	}

	w = wildcard{"foo", "oof"}
	if w.match("foof") {
		t.Error("foo*oof should not match foof")
	}
}

func TestConvert(t *testing.T) {
	s := convert([]string{"A", "b", "C"}, strings.ToLower)
	e := []string{"a", "b", "c"}
	if s[0] != e[0] || s[1] != e[1] || s[2] != e[2] {
		t.Errorf("%v != %v", s, e)
	}
}

func TestNormalizeOrigin(t *testing.T) {
	cases := []struct {
		origin string
		want   string
	}{
		{"http://example.com", "http://example.com"},
		{"http://example.com:80", "http://example.com"},
		{"https://example.com:443", "https://example.com"},
		{"https://example.com:8443", "https://example.com:8443"},
		{"HTTP://EXAMPLE.COM", "http://example.com"},
		{"http://Example.COM:80", "http://example.com"},
		{"http://example.com:8080", "http://example.com:8080"},
		{"http://[::1]:8080", "http://[::1]:8080"},
		{"http://[::1]:80", "http://[::1]"},
		{"null", "null"},
		{"http://example.com/", "http://example.com/"},
		{"http://example.com/path", "http://example.com/path"},
		{"file://", "file://"},
	}
	for _, tc := range cases {
		if got := normalizeOrigin(tc.origin); got != tc.want {
			t.Errorf("normalizeOrigin(%q) = %q, want %q", tc.origin, got, tc.want)
		}
	}
}

func TestAddVary(t *testing.T) {
	cases := []struct {
		name     string
		existing []string
		values   []string
		want     []string
	}{
		{"empty", nil, []string{"Origin"}, []string{"Origin"}},
		{"merge", []string{"Accept-Encoding"}, []string{"Origin"}, []string{"Accept-Encoding, Origin"}},
		{"dedup", []string{"Origin"}, []string{"origin"}, []string{"Origin"}},
		{"dedup_values", nil, []string{"Origin", "Origin"}, []string{"Origin"}},
		{"combined", nil, []string{"Origin, Access-Control-Request-Method"}, []string{"Origin, Access-Control-Request-Method"}},
		{"star", []string{"*"}, []string{"Origin"}, []string{"*"}},
		{"multi_line", []string{"Accept-Encoding", "origin"}, []string{"Authorization"}, []string{"Accept-Encoding, origin, Authorization"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := http.Header{}
			if tc.existing != nil {
				h["Vary"] = tc.existing
			}
			addVary(h, tc.values...)
			if got := h["Vary"]; !slices.Equal(got, tc.want) {
				t.Errorf("Vary = %q, want %q", got, tc.want)
			}
		})
	}
}

func BenchmarkWildcard(b *testing.B) {
	w := wildcard{"foo", "bar"}
	b.Run("match", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			w.match("foobazbar")
		}
	})
	b.Run("too short", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			w.match("fobar")
		}
	})
}
