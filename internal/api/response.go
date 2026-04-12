package api

import (
	"net/http"

	"github.com/acdgbrasil/svc-analysis-bi/internal/api/handlers"
)

// Response types are re-exported from handlers where they are canonically
// defined. This prevents duplication while avoiding import cycles
// (api imports handlers, but handlers never imports api).

// Response is the standard envelope for all API responses.
type Response = handlers.Response

// ResponseMeta carries metadata about the response.
type ResponseMeta = handlers.ResponseMeta

// ErrorBody is the payload for error responses.
type ErrorBody = handlers.ErrorBody

// NewResponseMeta creates a ResponseMeta with defaults.
var NewResponseMeta = handlers.NewResponseMeta

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, resp Response) {
	handlers.WriteJSON(w, status, resp)
}

// WriteError writes a standardized error response.
func WriteError(w http.ResponseWriter, status int, msg string) {
	handlers.WriteError(w, status, msg)
}
