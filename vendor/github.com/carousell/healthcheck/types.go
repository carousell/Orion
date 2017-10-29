package healthcheck

import (
	"net/http"
)

// HealthCheck is the default health checker implementation
type HealthCheck interface {
	SetHealth(health bool)
	IsHealthy() bool
	HealthCheck(http.ResponseWriter, *http.Request)
	SetHealthy(http.ResponseWriter, *http.Request)
	SetUnhealthy(http.ResponseWriter, *http.Request)
}
