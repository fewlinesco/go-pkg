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

type Metrics struct {
	QueryLatencyMs  *metrics.Float64Measure
	QueryErrorTotal *metrics.Float64Measure
}

type Config struct {
	Driver   string            `json:"driver"`
	Scheme   string            `json:"scheme"`
	Host     string            `json:"host"`
	Port     int               `json:"port"`
	Username string            `json:"username"`
	Password string            `json:"password"`
	Database string            `json:"database"`
	Options  map[string]string `json:"options"`
}

var DefaultConfig = Config{
	Driver:   "postgres",
	Scheme:   "postgresql",
	Host:     "localhost",
	Port:     5432,
	Username: "postgres",
	Password: "postgres",
	Database: "postgres",
}

var DefaultMetrics = Metrics{
	QueryLatencyMs:  metrics.Float64("sql_query_latency_ms", "The query latency in milliseconds", metrics.UnitMilliseconds),
	QueryErrorTotal: metrics.Float64("sql_query_error_total", "The query error total", metrics.UnitDimensionless),
}

var MetricViews = []*metrics.View{
	&metrics.View{
		Name:        "sql/query_latency",
		Measure:     DefaultMetrics.QueryLatencyMs,
		Description: "The distribution of the latencies",
		Aggregation: metrics.ViewDistribution(0, 25, 100, 200, 400, 800, 10000),
	},
	&metrics.View{
		Name:        "sql/query_error",
		Measure:     DefaultMetrics.QueryErrorTotal,
		Description: "The number of errors",
		Aggregation: metrics.ViewCount(),
	},
}

type DB struct {
	db *sqlx.DB
}

type Tx struct {
	tx *sqlx.Tx
}

func Connect(config Config) (*DB, error) {
	options := make(url.Values)
	for key, value := range config.Options {
		options.Add(key, value)
	}

	connectionstring := url.URL{
		Scheme:   config.Scheme,
		User:     url.UserPassword(config.Username, config.Password),
		Host:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Path:     config.Database,
		RawQuery: options.Encode(),
	}

	db, err := sqlx.Connect(config.Driver, connectionstring.String())
	if err != nil {
		return nil, fmt.Errorf("can't connect to database: %v", err)
	}

	return &DB{db: db}, nil
}

func IsUniqueConstraintError(err error, constraintName string) bool {
	e, ok := err.(*pq.Error)
	if !ok {
		return false
	}

	return e.Code == "23505" && e.Constraint == constraintName
}

func GetCurrentTimestamp() time.Time {
	return time.Now().UTC().Truncate(time.Millisecond)
}

func (db *DB) Begin() (*Tx, error) {
	tx, err := db.db.Beginx()
	if err != nil {
		return nil, err
	}

	return &Tx{tx: tx}, nil
}

func (db *DB) SelectContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error {
	var err error

	metrics.RecordElapsedTimeInMilliseconds(ctx, DefaultMetrics.QueryLatencyMs, func() {
		err = db.db.SelectContext(ctx, dest, statement, args...)
	})

	metrics.RecordError(ctx, DefaultMetrics.QueryErrorTotal, err)

	return err
}

func (db *DB) GetContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error {
	var err error

	metrics.RecordElapsedTimeInMilliseconds(ctx, DefaultMetrics.QueryLatencyMs, func() {
		err = db.db.GetContext(ctx, dest, statement, args...)
	})

	metrics.RecordError(ctx, DefaultMetrics.QueryErrorTotal, err)

	return err
}

func (db *DB) ExecContext(ctx context.Context, statement string, arg ...interface{}) (sql.Result, error) {
	var (
		response sql.Result
		err      error
	)

	metrics.RecordElapsedTimeInMilliseconds(ctx, DefaultMetrics.QueryLatencyMs, func() {
		response, err = db.db.ExecContext(ctx, statement, arg...)
	})

	metrics.RecordError(ctx, DefaultMetrics.QueryErrorTotal, err)

	return response, err
}
func (db *DB) NamedExecContext(ctx context.Context, statement string, arg interface{}) (sql.Result, error) {
	var (
		response sql.Result
		err      error
	)

	metrics.RecordElapsedTimeInMilliseconds(ctx, DefaultMetrics.QueryLatencyMs, func() {
		response, err = db.db.NamedExecContext(ctx, statement, arg)
	})

	metrics.RecordError(ctx, DefaultMetrics.QueryErrorTotal, err)

	return response, err
}

func (db *DB) PingContext(ctx context.Context) error {
	var err error

	metrics.RecordElapsedTimeInMilliseconds(ctx, DefaultMetrics.QueryLatencyMs, func() {
		err = db.db.PingContext(ctx)
	})

	metrics.RecordError(ctx, DefaultMetrics.QueryErrorTotal, err)

	return err
}

func (tx *Tx) Commit() error {
	return tx.tx.Commit()
}

func (tx *Tx) Rollback() error {
	return tx.tx.Rollback()
}

func (tx *Tx) GetContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error {
	var err error

	metrics.RecordElapsedTimeInMilliseconds(ctx, DefaultMetrics.QueryLatencyMs, func() {
		err = tx.tx.GetContext(ctx, dest, statement, args)
	})

	metrics.RecordError(ctx, DefaultMetrics.QueryErrorTotal, err)

	return err
}

func (tx *Tx) NamedExecContext(ctx context.Context, statement string, arg interface{}) (sql.Result, error) {
	var (
		response sql.Result
		err      error
	)

	metrics.RecordElapsedTimeInMilliseconds(ctx, DefaultMetrics.QueryLatencyMs, func() {
		response, err = tx.tx.NamedExecContext(ctx, statement, arg)
	})

	metrics.RecordError(ctx, DefaultMetrics.QueryErrorTotal, err)

	return response, err
}
