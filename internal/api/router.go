package api

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/acdgbrasil/svc-analysis-bi/internal/api/handlers"
	"github.com/acdgbrasil/svc-analysis-bi/internal/api/middleware"
	"github.com/acdgbrasil/svc-analysis-bi/internal/export"
)

// RouterDeps holds the dependencies required to build the API router.
type RouterDeps struct {
	Logger       *slog.Logger
	DB           HealthChecker
	NATS         NATSChecker
	JWTValidator middleware.JWTValidator

	// RateLimit configuration. Zero value disables rate limiting.
	RateLimitRPS   float64
	RateLimitBurst int

	// Indicators is the store used for indicator queries. When nil,
	// indicator and export endpoints return 501 Not Implemented.
	Indicators handlers.IndicatorQuerier

	// Encoders is the export format registry. When nil, export and
	// metadata/formats endpoints return empty results.
	Encoders map[string]export.Encoder
}

// NewRouter creates and configures the HTTP handler with the full
// middleware chain and route table using chi/v5.
//
// Middleware order: recovery -> security headers -> rate limit -> (routes with optional JWT)
//
// Routes:
//
//	GET /health                              - liveness probe (public)
//	GET /ready                               - readiness probe (public, checks DB + NATS)
//	GET /api/v1/indicators/{axis}            - indicator queries (auth required)
//	GET /api/v1/export/{format}              - data export (auth required)
//	GET /api/v1/metadata/{resource}          - metadata queries (auth required)
func NewRouter(deps RouterDeps) http.Handler {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}

	r := chi.NewRouter()

	// Global middleware chain (outermost first)
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.SecurityHeaders())

	if deps.RateLimitRPS > 0 && deps.RateLimitBurst > 0 {
		r.Use(middleware.RateLimit(middleware.RateLimitConfig{
			RequestsPerSecond: deps.RateLimitRPS,
			BurstSize:         deps.RateLimitBurst,
		}))
	}

	// Public health endpoints
	r.Get("/health", handlers.HealthHandler())
	r.Get("/ready", handlers.ReadyHandler(deps.DB, deps.NATS))

	// Protected API group
	r.Group(func(r chi.Router) {
		if deps.JWTValidator != nil {
			skipPaths := map[string]bool{}
			r.Use(middleware.JWTAuth(deps.JWTValidator, skipPaths))
		} else {
			logger.Warn("JWT authentication is DISABLED — no JWTValidator provided")
		}

		// Wire real handlers when stores are available, otherwise 501 placeholders.
		if deps.Indicators != nil {
			r.Get("/api/v1/indicators/{axis}", handlers.IndicatorsHandler(deps.Indicators))
		} else {
			r.Get("/api/v1/indicators/{axis}", placeholderHandler("indicators"))
		}

		encoders := deps.Encoders
		if encoders == nil {
			encoders = map[string]export.Encoder{}
		}

		if deps.Indicators != nil {
			r.Get("/api/v1/export/{format}", handlers.ExportHandler(deps.Indicators, encoders, logger))
		} else {
			r.Get("/api/v1/export/{format}", placeholderHandler("export"))
		}

		r.Get("/api/v1/metadata/{resource}", handlers.MetadataHandler(encoders))
	})

	return r
}

// placeholderHandler returns a 501 Not Implemented response for endpoints
// that require a store dependency that was not provided.
func placeholderHandler(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		WriteError(w, http.StatusNotImplemented, name+" endpoint not yet implemented")
	}
}
