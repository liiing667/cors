package cors

import (
	"net"
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

// convert converts a list of string using the passed converter function
func convert(s []string, f func(string) string) []string {
	out := make([]string, len(s))
	for i := range s {
		out[i] = f(s[i])
	}
	return out
}

// defaultPorts maps URL schemes to their default port.
var defaultPorts = map[string]string{
	"http":  "80",
	"https": "443",
}

// normalizeOrigin normalizes an origin so that equivalent origins compare
// equal: scheme and host are lowercased and the default port (80 for http,
// 443 for https) is dropped. Origins that do not parse as a plain
// scheme://host[:port] URL are merely lowercased.
func normalizeOrigin(origin string) string {
	u, err := url.Parse(origin)
	if err != nil || u.Scheme == "" || u.Host == "" || u.User != nil ||
		u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return strings.ToLower(origin)
	}
	scheme := strings.ToLower(u.Scheme)
	host := strings.ToLower(u.Hostname())
	if port := u.Port(); port != "" && defaultPorts[scheme] != port {
		return scheme + "://" + net.JoinHostPort(host, port)
	}
	if strings.Contains(host, ":") {
		// IPv6 literal without a port
		host = "[" + host + "]"
	}
	return scheme + "://" + host
}

// normalizeWildcardOrigin lowercases a wildcard origin pattern and drops a
// trailing default port so that it matches normalized origins.
func normalizeWildcardOrigin(origin string) string {
	origin = strings.ToLower(origin)
	scheme, rest, ok := strings.Cut(origin, "://")
	if !ok {
		return origin
	}
	if def, ok := defaultPorts[scheme]; ok {
		rest = strings.TrimSuffix(rest, ":"+def)
	}
	return scheme + "://" + rest
}

// addVary merges the given header names into the Vary response header,
// preserving pre-existing entries and avoiding duplicates. If Vary already
// contains "*", it is left untouched.
func addVary(headers http.Header, values ...string) {
	var tokens []string
	seen := map[string]bool{}
	for _, v := range headers["Vary"] {
		for _, tok := range strings.Split(v, ",") {
			tok = strings.TrimSpace(tok)
			if tok == "" {
				continue
			}
			if tok == "*" {
				// Vary: * already covers everything
				return
			}
			if key := strings.ToLower(tok); !seen[key] {
				seen[key] = true
				tokens = append(tokens, tok)
			}
		}
	}
	for _, v := range values {
		for _, tok := range strings.Split(v, ",") {
			tok = strings.TrimSpace(tok)
			if tok == "" {
				continue
			}
			if key := strings.ToLower(tok); !seen[key] {
				seen[key] = true
				tokens = append(tokens, tok)
			}
		}
	}
	if len(tokens) > 0 {
		headers["Vary"] = []string{strings.Join(tokens, ", ")}
	}
}
