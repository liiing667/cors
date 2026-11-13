package cors

import (
	"net/http"
	"net/url"
	"strings"
)

type wildcard struct {
	prefix string
	suffix string
}

func (w wildcard) match(s string) bool {
	return len(s) >= len(w.prefix)+len(w.suffix) &&
		strings.HasPrefix(s, w.prefix) &&
		strings.HasSuffix(s, w.suffix)
}

// originFromRequest extracts and validates the Origin header of the request,
// returning the normalized origin. A missing header yields ("", true). ok is
// false when the header is present but malformed, or when it carries multiple
// values: Origin is a singleton header and duplicates (as produced by some
// reverse proxies) cannot be reconciled safely, so they are rejected.
func originFromRequest(r *http.Request) (origin string, ok bool) {
	values := r.Header["Origin"]
	if len(values) == 0 {
		return "", true
	}
	if len(values) > 1 {
		return "", false
	}
	return normalizeOrigin(values[0])
}

// normalizeOrigin parses a serialized origin
// (scheme "://" host [ ":" port ], or the literal "null") and returns its
// normalized form with a lowercased scheme and host. ok is false when the
// value is not a well-formed origin.
func normalizeOrigin(origin string) (normalized string, ok bool) {
	if origin == "null" {
		return origin, true
	}
	u, err := url.Parse(origin)
	if err != nil || u.Scheme == "" || u.Host == "" || u.User != nil ||
		u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return "", false
	}
	return u.Scheme + "://" + strings.ToLower(u.Host), true
}

// convert converts a list of string using the passed converter function
func convert(s []string, f func(string) string) []string {
	out := make([]string, len(s))
	for i := range s {
		out[i] = f(s[i])
	}
	return out
}
