package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// ErrorBody is the JSON envelope returned by WriteError. The RequestID lets
// clients correlate failures with server logs.
type ErrorBody struct {
	Error     string `json:"error"`
	RequestID string `json:"requestID,omitempty"`
}

// WriteJSON serialises body as JSON and writes it with the given status. On
// marshal failure it logs and falls back to a 500 plaintext response.
func WriteJSON(w http.ResponseWriter, _ *http.Request, status int, body any) {
	data, err := json.Marshal(body)
	if err != nil {
		slog.Error("encode response", slog.Any("err", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(data)
}

// WriteError emits the standard JSON error envelope.
func WriteError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	WriteJSON(w, r, status, ErrorBody{
		Error:     msg,
		RequestID: RequestIDFromContext(r.Context()),
	})
}

// NoContent writes a 204 with no body.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
