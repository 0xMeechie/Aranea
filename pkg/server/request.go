package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// DefaultMaxRequestBytes caps inbound JSON bodies to protect against runaway
// clients. Override per-call with DecodeJSONLimit.
const DefaultMaxRequestBytes = 1 << 20 // 1 MiB

// DecodeJSON reads a JSON body into dst. It enforces DefaultMaxRequestBytes,
// rejects unknown fields, and rejects bodies containing more than one value.
func DecodeJSON(r *http.Request, dst any) error {
	return DecodeJSONLimit(r, dst, DefaultMaxRequestBytes)
}

// DecodeJSONLimit is DecodeJSON with a caller-supplied byte limit.
func DecodeJSONLimit(r *http.Request, dst any, max int64) error {
	if r.Body == nil {
		return errors.New("missing request body")
	}
	r.Body = http.MaxBytesReader(nil, r.Body, max)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return errors.New("empty request body")
		}
		return fmt.Errorf("decode request body: %w", err)
	}
	if dec.More() {
		return errors.New("request body must contain a single JSON value")
	}
	return nil
}

// QueryString returns the trimmed value of the named query param.
func QueryString(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

// PathValue is a thin wrapper over r.PathValue for symmetry with QueryString.
// Patterns registered with Go 1.22+ syntax (e.g. "/widgets/{id}") populate it.
func PathValue(r *http.Request, name string) string {
	return r.PathValue(name)
}
