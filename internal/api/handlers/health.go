// Package handlers contains HTTP handler functions for the analysis-bi API.
package handlers

import (
	"context"
	"net/http"
)

// HealthChecker verifies database connectivity.
type HealthChecker interface {
	Ping(ctx context.Context) error
}

// NATSChecker verifies NATS JetStream connectivity.
type NATSChecker interface {
	IsConnected() bool
}

// HealthHandler handles the /health liveness endpoint.
// It always returns 200 OK to indicate the process is running.
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := Response{
			Data: map[string]string{"status": "ok"},
			Meta: NewResponseMeta(),
		}
		WriteJSON(w, http.StatusOK, resp)
	}
}

// ReadyHandler handles the /ready readiness endpoint.
// It checks database and NATS connectivity before returning 200 OK.
// If any dependency is unhealthy, it returns 503 Service Unavailable.
func ReadyHandler(db HealthChecker, nats NATSChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbOK := true
		natsOK := true

		if db != nil {
			if err := db.Ping(r.Context()); err != nil {
				dbOK = false
			}
		}

		if nats != nil {
			natsOK = nats.IsConnected()
		}

		if !dbOK || !natsOK {
			resp := Response{
				Data: map[string]any{
					"status":   "unavailable",
					"database": dbOK,
					"nats":     natsOK,
				},
				Meta: NewResponseMeta(),
			}
			WriteJSON(w, http.StatusServiceUnavailable, resp)
			return
		}

		resp := Response{
			Data: map[string]any{
				"status":   "ready",
				"database": dbOK,
				"nats":     natsOK,
			},
			Meta: NewResponseMeta(),
		}
		WriteJSON(w, http.StatusOK, resp)
	}
}
