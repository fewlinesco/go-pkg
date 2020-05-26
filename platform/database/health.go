package database

import (
	"context"

	"go.opencensus.io/trace"

	"github.com/fewlinesco/go-pkg/platform/web"
)

func HealthCheck(db *DB) web.HealthzChecker {
	return func(ctx context.Context) web.HealthzStatus {
		ctx, span := trace.StartSpan(ctx, "database.HealthChecker")
		span.End()

		service := web.HealthzStatus{
			Type:        "Database",
			Description: "Check the availability of the service's database",
			State:       web.HealthzStateHealthy,
		}

		err := db.PingContext(ctx)
		if err != nil {
			errorMessage := err.Error()

			service.Error = errorMessage
			span.AddAttributes(trace.StringAttribute("database-health-error", errorMessage))

			service.State = web.HealthzStateUnhealthy
		}

		return service
	}
}
