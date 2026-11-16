package cors

import (
	"net/http"
	"slices"
	"sort"
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

// convert converts a list of string using the passed converter function
func convert(s []string, f func(string) string) []string {
	out := make([]string, len(s))
	for i := range s {
		out[i] = f(s[i])
	}
	return out
}

// normalizeMethods normalizes a list of configured HTTP methods:
// every method is upper-cased (per the method-normalization rules of the
// Fetch standard) and duplicates are collapsed while preserving the order of
// first appearance. The special "*" value, once encountered, takes
// precedence over the rest of the list and is returned on its own.
func normalizeMethods(methods []string) []string {
	normalized := make([]string, 0, len(methods))
	seen := make(map[string]bool, len(methods))
	for _, method := range methods {
		method = strings.ToUpper(method)
		if method == "" {
			continue
		}
		if method == "*" {
			return []string{"*"}
		}
		if seen[method] {
			continue
		}
		seen[method] = true
		normalized = append(normalized, method)
	}
	return normalized
}

// normalizeHeaderNames normalizes a list of configured header names:
// every name is canonicalized with http.CanonicalHeaderKey and duplicates
// are collapsed while preserving the order of first appearance. The special
// "*" value, once encountered, takes precedence over the rest of the list
// and is returned on its own.
func normalizeHeaderNames(names []string) []string {
	normalized := make([]string, 0, len(names))
	seen := make(map[string]bool, len(names))
	for _, name := range names {
		name = http.CanonicalHeaderKey(name)
		if name == "" {
			continue
		}
		if name == "*" {
			return []string{"*"}
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		normalized = append(normalized, name)
	}
	return normalized
}

// normalizeRequestHeaderValues turns the raw values of a list-based request
// header field (such as Access-Control-Request-Headers) into a single
// normalized list of header names. It splits field values on commas, trims
// optional whitespace around each element, lower-cases the names, drops
// empty elements and collapses duplicates. The result is sorted
// lexicographically, as guaranteed by the Fetch standard for CORS-unsafe
// request-header names.
//
// Policy acceptance runs first via SortedSet.Accepts; this function only
// reformats values that the policy already accepted, so its inputs are
// expected to be short. It therefore avoids map allocations and relies on
// linear scans instead.
func normalizeRequestHeaderValues(values []string) []string {
	var names []string
	for _, value := range values {
		for _, element := range strings.Split(value, ",") {
			name := strings.ToLower(strings.Trim(element, " \t"))
			if name == "" || slices.Contains(names, name) {
				continue
			}
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}
