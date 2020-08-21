package healthcheck

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthyStatus represents the current status of the service
type HealthyStatus int

// Healthy indicates that the service is in a healthy state
// Unhealthy indicates that the service is in an unhealthy state
const (
	Healthy HealthyStatus = iota
	Unhealthy
)

func (h HealthyStatus) String() string {
	return [...]string{"Healthy", "Unhealthy"}[h]
}

// NewOptions represent the intialization options for the healthcheck controller
type NewOptions struct {
	Function func() bool
}

// Response is the shape of the healthcheck response
type Response struct {
	Health             HealthyStatus `json:"health"`
	CurrentTime        time.Time     `json:"currentTime"`
	Uptime             time.Duration `json:"uptime"`
	TimeInCurrentState time.Duration `json:"timeInCurrentState"`
}

// New instantiates a new healtcheck handler with the passed in healthcheck function
func New(o *NewOptions) http.HandlerFunc {
	startTime := time.Now()
	st := map[string]interface{}{
		"health":            Healthy,
		"startTime":         startTime,
		"lastStateChangeAt": startTime,
	}
	fn := o.Function
	if fn == nil {
		fn = func() bool {
			return true
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var currentHealth HealthyStatus
		now := time.Now()
		isHealthy := fn()
		if isHealthy {
			currentHealth = Healthy
		} else {
			currentHealth = Unhealthy
		}

		lastHealthValue, ok := st["health"].(HealthyStatus)
		if !ok {
			lastHealthValue = Unhealthy
		}

		lastStatusChangeAt, ok := st["lastStateChangeAt"].(time.Time)
		if !ok || lastHealthValue != currentHealth {
			lastStatusChangeAt = time.Now()
		}
		stateTime := now.Sub(lastStatusChangeAt)

		startTime, ok := st["startTime"].(time.Time)
		if !ok {
			startTime = time.Now()
			st["startTime"] = time.Now()
		}
		uptime := now.Sub(startTime)
		res := Response{
			Health:             currentHealth,
			CurrentTime:        now,
			TimeInCurrentState: stateTime,
			Uptime:             uptime,
		}

		j, err := json.Marshal(res)

		if err != nil {
			currentHealth = Unhealthy
		}

		statusCode := 200
		if currentHealth != Healthy {
			statusCode = 500
		}
		st["health"] = currentHealth
		st["lastStateChangeAt"] = lastStatusChangeAt
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		w.Write(j)
	}
}
