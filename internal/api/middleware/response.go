package middleware

import (
	"encoding/json"
	"net/http"
)

// writeJSONError writes a JSON error response matching the standard API envelope.
// Used by middleware that cannot import the handlers package (to avoid cycles).
func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error":   http.StatusText(status),
		"status":  status,
		"message": msg,
	})
}
