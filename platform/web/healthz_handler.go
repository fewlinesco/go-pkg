package web

import (
	"context"
	"net/http"

	"go.opencensus.io/trace"
)

// HealthzState represents all the possible state an application can be in
type HealthzState string

const (
	// HealthzStateHealthy everything is working as expected
	HealthzStateHealthy HealthzState = "healthy"
	// HealthzStateDegraded some of its behaviours might not work but the core business is working as expected
	HealthzStateDegraded HealthzState = "degraded"
	// HealthzStateUnhealthy some important part of the core business is not working and thus it shouldn't handle any request to make sure we don't make the data in an inconsistent state
	HealthzStateUnhealthy HealthzState = "unhealthy"
)

// HealthzChecker is the type application handlers needs to conform when they define health check
type HealthzChecker func(context.Context) HealthzStatus

// HealthzStatus represents the state that will be returned by the healthz endpoint for each check
type HealthzStatus struct {
	Type        string            `json:"type"`
	Description string            `json:"description"`
	Error       string            `json:"error,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	State       HealthzState      `json:"state"`
}

// HealthzResponse is the top level JSON structure sent by the healthz endpoint
type HealthzResponse struct {
	State    string          `json:"state"`
	Services []HealthzStatus `json:"services"`
}

// HealthzHandler reprensents the generic healthz HTTP handler in charge of building / displaying the full overview of an application health. The global health state is equal to the worse value return by the checkers.
func HealthzHandler(serviceCheckers []HealthzChecker) Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
		ctx, span := trace.StartSpan(ctx, "platform.web.HealthzHandler")
		defer span.End()

		var response HealthzResponse

		state := HealthzStateHealthy

		for _, serviceChecker := range serviceCheckers {
			service := serviceChecker(ctx)
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
