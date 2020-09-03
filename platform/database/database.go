package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/fewlinesco/go-pkg/platform/metrics"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// Config represents the database configuration that can be defined / override by the application
type Config struct {
	URL      string            `json:"url"`
	Driver   string            `json:"driver"`
	Scheme   string            `json:"scheme"`
	Host     string            `json:"host"`
	Port     int               `json:"port"`
	Username string            `json:"username"`
	Password string            `json:"password"`
	Database string            `json:"database"`
	Options  map[string]string `json:"options"`
}

// DefaultConfig are the default values for any application
var DefaultConfig = Config{
	URL:      "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
	Driver:   "postgres",
	Scheme:   "postgresql",
	Host:     "localhost",
	Port:     5432,
	Username: "postgres",
	Password: "postgres",
	Database: "postgres",
}

var (
	metricQueryLatencyMs  = metrics.Float64("sql_query_latency_ms", "The query latency in milliseconds", metrics.UnitMilliseconds)
	metricQueryErrorTotal = metrics.Float64("sql_query_error_total", "The query error total", metrics.UnitDimensionless)
)

// MetricViews are the generic metrics generated for any datbase based applications
var MetricViews = []*metrics.View{
	{
		Name:        "sql/query_latency",
		Measure:     metricQueryLatencyMs,
		Description: "The distribution of the latencies",
		Aggregation: metrics.ViewDistribution(0, 25, 100, 200, 400, 800, 10000),
	},
	{
		Name:        "sql/query_error",
		Measure:     metricQueryErrorTotal,
		Description: "The number of errors",
		Aggregation: metrics.ViewCount(),
	},
}

// DB represents the database connection
type DB struct {
	db *sqlx.DB
}

// Tx represents a database transaction
type Tx struct {
	tx *sqlx.Tx
}

// Connect configures the driver and opens a database connection
func Connect(config Config) (*DB, error) {
	if config.URL == "" {
		options := make(url.Values)
		for key, value := range config.Options {
			options.Add(key, value)
		}

		connectionURL := url.URL{
			Scheme:   config.Scheme,
			User:     url.UserPassword(config.Username, config.Password),
			Host:     fmt.Sprintf("%s:%d", config.Host, config.Port),
			Path:     config.Database,
			RawQuery: options.Encode(),
		}
		config.URL = connectionURL.String()
	}
	db, err := sqlx.Connect(config.Driver, config.URL)
	if err != nil {
		return nil, fmt.Errorf("can't connect to database: %v", err)
	}

	return &DB{db: db}, nil
}

// IsUniqueConstraintError is a helper checking the current database error and returnning true if it's a PG unique index
// error for a specific constraint name
func IsUniqueConstraintError(err error, constraintName string) bool {
	e, ok := err.(*pq.Error)
	if !ok {
		return false
	}

	return e.Code == "23505" && e.Constraint == constraintName
}

// GetCurrentTimestamp is a helper function that generates a new UTC timestamp truncated at the millisecond
// because PG is not able to handle nanoseconds
func GetCurrentTimestamp() time.Time {
	return time.Now().UTC().Truncate(time.Millisecond)
}

// Begin starts a new transaction
func (db *DB) Begin() (*Tx, error) {
	tx, err := db.db.Beginx()
	if err != nil {
		return nil, err
	}

	return &Tx{tx: tx}, nil
}

// SelectContext fetches a slice of elements from database.
func (db *DB) SelectContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error {
	var err error

	metrics.RecordElapsedTimeInMilliseconds(ctx, metricQueryLatencyMs, func() {
		err = db.db.SelectContext(ctx, dest, statement, args...)
	})

	metrics.RecordError(ctx, metricQueryErrorTotal, err)

	return err
}

// GetContext fetches one elements from database.
func (db *DB) GetContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error {
	var err error

	metrics.RecordElapsedTimeInMilliseconds(ctx, metricQueryLatencyMs, func() {
		err = db.db.GetContext(ctx, dest, statement, args...)
	})

	metrics.RecordError(ctx, metricQueryErrorTotal, err)

	return err
}

// ExecContext executes any SQL query to the server. It's mostly use for insert/update commands
func (db *DB) ExecContext(ctx context.Context, statement string, arg ...interface{}) (sql.Result, error) {
	var (
		response sql.Result
		err      error
	)

	metrics.RecordElapsedTimeInMilliseconds(ctx, metricQueryLatencyMs, func() {
		response, err = db.db.ExecContext(ctx, statement, arg...)
	})

	metrics.RecordError(ctx, metricQueryErrorTotal, err)

	return response, err
}

// NamedExecContext same as ExecContext but use name arguments in the statement and a struct as parameter
func (db *DB) NamedExecContext(ctx context.Context, statement string, arg interface{}) (sql.Result, error) {
	var (
		response sql.Result
		err      error
	)

	metrics.RecordElapsedTimeInMilliseconds(ctx, metricQueryLatencyMs, func() {
		response, err = db.db.NamedExecContext(ctx, statement, arg)
	})

	metrics.RecordError(ctx, metricQueryErrorTotal, err)

	return response, err
}

// PingContext pings the database to make sure the connection is still open and working
func (db *DB) PingContext(ctx context.Context) error {
	var err error

	metrics.RecordElapsedTimeInMilliseconds(ctx, metricQueryLatencyMs, func() {
		err = db.db.PingContext(ctx)
	})

	metrics.RecordError(ctx, metricQueryErrorTotal, err)

	return err
}

// Commit persists the transaction
func (tx *Tx) Commit() error {
	return tx.tx.Commit()
}

// Rollback aborts the transaction
func (tx *Tx) Rollback() error {
	return tx.tx.Rollback()
}

// GetContext same as db.GetContext but for the current transction
func (tx *Tx) GetContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error {
	var err error

	metrics.RecordElapsedTimeInMilliseconds(ctx, metricQueryLatencyMs, func() {
		err = tx.tx.GetContext(ctx, dest, statement, args)
	})

	metrics.RecordError(ctx, metricQueryErrorTotal, err)

	return err
}

// NamedExecContext same as db.NamedExecContext but for the current transction
func (tx *Tx) NamedExecContext(ctx context.Context, statement string, arg interface{}) (sql.Result, error) {
	var (
		response sql.Result
		err      error
	)

	metrics.RecordElapsedTimeInMilliseconds(ctx, metricQueryLatencyMs, func() {
		response, err = tx.tx.NamedExecContext(ctx, statement, arg)
	})

	metrics.RecordError(ctx, metricQueryErrorTotal, err)

	return response, err
}
