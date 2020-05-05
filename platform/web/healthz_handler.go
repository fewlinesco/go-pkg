package web

import (
	"context"
	"net/http"
)

type HealthzState string

const (
	HealthzStateHealthy   HealthzState = "healthy"
	HealthzStateDegraded  HealthzState = "degraded"
	HealthzStateUnhealthy HealthzState = "unhealthy"
)

type HealthzChecker func() HealthzStatus

type HealthzStatus struct {
	Type        string            `json:"type"`
	Description string            `json:"description"`
	Error       string            `json:"error,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	State       HealthzState      `json:"state"`
}

type HealthzResponse struct {
	State    string          `json:"state"`
	Services []HealthzStatus `json:"services"`
}

func HealthzHandler(serviceCheckers []HealthzChecker) Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
		var response HealthzResponse

		state := HealthzStateHealthy

		for _, serviceChecker := range serviceCheckers {
			service := serviceChecker()
			if state != HealthzStateUnhealthy {
				switch service.State {
				case HealthzStateUnhealthy:
					state = HealthzStateUnhealthy

				case HealthzStateDegraded:
					state = HealthzStateDegraded
				}
			}

			response.Services = append(response.Services, service)
		}

		response.State = string(state)

		Respond(ctx, w, response, state.httpStatus())
		return nil
	}
}

func (state HealthzState) httpStatus() int {
	if state == HealthzStateUnhealthy {
		return http.StatusServiceUnavailable
	}

	return http.StatusOK
}
