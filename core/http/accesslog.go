package http

import (
	"net/url"
	"strings"
)

// localHandlerError is the Locals key holding a handler's error for the access log.
const localHandlerError = "goe.http.handler_error"

// defaultRedactedQuery lists the query parameters whose values the access log
// redacts unless configured otherwise.
var defaultRedactedQuery = []string{
	"access_token", "api_key", "apikey", "auth", "code", "id_token", "key",
	"password", "refresh_token", "secret", "sig", "signature", "token",
}

// queryRedactor replaces the values of sensitive query parameters.
type queryRedactor map[string]struct{}

func newQueryRedactor(names []string) queryRedactor {
	r := make(queryRedactor, len(names))
	for _, name := range names {
		r[strings.ToLower(name)] = struct{}{}
	}
	return r
}

// redact returns rawURL with the values of the redactor's parameters replaced
// by REDACTED, keeping everything else as it was sent.
func (r queryRedactor) redact(rawURL string) string {
	path, query, hasQuery := strings.Cut(rawURL, "?")
	if !hasQuery || len(r) == 0 {
		return rawURL
	}
	pairs := strings.Split(query, "&")
	changed := false
	for i, pair := range pairs {
		key, _, hasValue := strings.Cut(pair, "=")
		if !hasValue {
			continue
		}
		name, err := url.QueryUnescape(key)
		if err != nil {
			name = key
		}
		if _, sensitive := r[strings.ToLower(name)]; sensitive {
			pairs[i] = key + "=REDACTED"
			changed = true
		}
	}
	if !changed {
		return rawURL
	}
	return path + "?" + strings.Join(pairs, "&")
}
