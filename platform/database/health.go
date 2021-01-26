package database

import (
	"context"
	"fmt"

	"go.opencensus.io/trace"

	"github.com/fewlinesco/go-pkg/platform/web"
)

// HealthCheck is a generic health checker in charge of checking the database availability
func (db *prodDB) HealthCheck(dbName string) web.HealthzChecker {
	return genericHealthCheck(dbName)(db)
}

// HealthCheck is a generic health checker in charge of checking the database availability
func (db *sandboxDB) HealthCheck(dbName string) web.HealthzChecker {
	return genericHealthCheck(dbName)(db)
}

func genericHealthCheck(databaseName string) func(db DB) web.HealthzChecker {
	spanName := fmt.Sprintf("%s.HealthChecker", databaseName)
	description := fmt.Sprintf("Check the availability of the service's %s", databaseName)

	return func(db DB) web.HealthzChecker {
		return func(ctx context.Context) web.HealthzStatus {
			ctx, span := trace.StartSpan(ctx, spanName)
			span.End()

			service := web.HealthzStatus{
				Type:        "Database",
				Description: description,
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
}
