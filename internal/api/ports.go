// Package api provides the HTTP API layer for svc-analysis-bi.
//
// It defines the chi router, standard response envelope, middleware chain,
// and health endpoints. All external dependencies are expressed as interfaces
// (ports) so that handlers remain testable in isolation.
//
// The canonical interface definitions live in handlers/ (consumer-side per DIP).
// This file re-exports them so callers can reference api.HealthChecker without
// importing the handlers sub-package.
package api

import "github.com/acdgbrasil/svc-analysis-bi/internal/api/handlers"

// HealthChecker verifies database connectivity.
// Canonical definition: handlers.HealthChecker.
type HealthChecker = handlers.HealthChecker

// NATSChecker verifies NATS JetStream connectivity.
// Canonical definition: handlers.NATSChecker.
type NATSChecker = handlers.NATSChecker
