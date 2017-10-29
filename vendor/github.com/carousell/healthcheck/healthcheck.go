package healthcheck

import (
	"net/http"
)

var (
	healthCheck = Health{
		healthy: false,
	}
)

// Health is the default implementation of Healthcheck
type Health struct {
	healthy bool
}

// SetHealth sets the health of the system
func (h *Health) SetHealth(health bool) {
	h.healthy = health
}

// IsHealthy checks the current health
func (h *Health) IsHealthy() bool {
	return h.healthy
}

// HealthCheck is the default http.HandlerFunc
// It returns HTTP 200 when healthy and HTTP 500 when unhealthy
func (h *Health) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	if h.healthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// SetHealthy is an implementation of http.HandlerFunc that sets the health to healthy
func (h *Health) SetHealthy(w http.ResponseWriter, r *http.Request) {
	h.SetHealth(true)
}

// SetUnhealthy is an implementation of http.HandlerFunc that sets the health to unhealthy
func (h *Health) SetUnhealthy(w http.ResponseWriter, r *http.Request) {
	h.SetHealth(false)
}

// GetHealthCheck returns the global HealthCheck object
func GetHealthCheck() *Health {
	return &healthCheck
}
