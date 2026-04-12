package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// Response is the standard envelope for all API responses.
// This is the single canonical definition — used by all handlers.
type Response struct {
	Data any          `json:"data"`
	Meta ResponseMeta `json:"meta"`
}

// ResponseMeta carries metadata about the response for observability
// and K-anonymity transparency.
type ResponseMeta struct {
	Timestamp        string `json:"timestamp"`
	Period           string `json:"period,omitempty"`
	KThreshold       int    `json:"k_threshold"`
	SuppressedGroups int    `json:"suppressed_groups"`
	TotalRecords     int    `json:"total_records"`
}

// ErrorBody is the payload for error responses.
type ErrorBody struct {
	Error   string `json:"error"`
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// NewResponseMeta creates a ResponseMeta with the current timestamp and
// default K-anonymity threshold of 5.
func NewResponseMeta() ResponseMeta {
	return ResponseMeta{
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		KThreshold: 5,
	}
}

// WriteJSON writes a JSON response with the given status code and body.
func WriteJSON(w http.ResponseWriter, status int, resp Response) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

// WriteError writes a standardized error response.
func WriteError(w http.ResponseWriter, status int, msg string) {
	body := ErrorBody{
		Error:   http.StatusText(status),
		Status:  status,
		Message: msg,
	}
	resp := Response{
		Data: body,
		Meta: NewResponseMeta(),
	}
	WriteJSON(w, status, resp)
}
