package database

import (
	"context"

	"github.com/jmoiron/sqlx"
	"go.opencensus.io/trace"

	"github.com/fewlinesco/go-pkg/platform/web"
)

func HealthCheck(db *sqlx.DB) web.HealthzChecker {
	return func(ctx context.Context) web.HealthzStatus {
		ctx, span := trace.StartSpan(ctx, "database.HealthChecker")
		span.End()

		service := web.HealthzStatus{
			Type:        "Database",
			Description: "Check the availability of the service's database",
			State: web.HealthzStateHealthy,
		}

		err := db.PingContext(ctx)
		if err != nil {
			service.Error = err.Error()
			service.State = web.HealthzStateUnhealthy
		}

		return service
	}
}
