package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/GuiaBolso/darwin"
	"github.com/fewlinesco/go-pkg/platform/metrics"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// Config represents the database configuration that can be defined / overridden by the application.
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

type WriteDB interface {
	NewGenericDriver(dialect darwin.Dialect) *darwin.GenericDriver
	Begin() (Tx, error)
	Close() error
	ExecContext(ctx context.Context, statement string, arg ...interface{}) (sql.Result, error)
	NamedExecContext(ctx context.Context, statement string, arg interface{}) (sql.Result, error)
	PingContext(ctx context.Context) error
}

type ReadDB interface {
	NewGenericDriver(dialect darwin.Dialect) *darwin.GenericDriver
	Begin() (Tx, error)
	Close() error
	SelectContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error
	SelectMultipleContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error
	PingContext(ctx context.Context) error
}

// DB is a generic interface for database interaction
type DB interface {
	NewGenericDriver(dialect darwin.Dialect) *darwin.GenericDriver
	Begin() (Tx, error)
	Close() error
	ExecContext(ctx context.Context, statement string, arg ...interface{}) (sql.Result, error)
	NamedExecContext(ctx context.Context, statement string, arg interface{}) (sql.Result, error)
	SelectContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error
	SelectMultipleContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error
	PingContext(ctx context.Context) error
}

// Tx is a generic interface for database transactions
type Tx interface {
	SelectContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error
	ExecContext(ctx context.Context, statement string, arg ...interface{}) (sql.Result, error)
	NamedExecContext(ctx context.Context, statement string, arg interface{}) (sql.Result, error)
	Commit() error
	Rollback() error
}

// DB represents the database connection
type prodDB struct {
	db *sqlx.DB
}

// Tx represents a database transaction
type prodTx struct {
	tx *sqlx.Tx
}

func (db *prodDB) NewGenericDriver(dialect darwin.Dialect) *darwin.GenericDriver {
	return darwin.NewGenericDriver(db.db.DB, darwin.PostgresDialect{})
}

func connect(config Config) (*sqlx.DB, error) {
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
	return sqlx.Connect(config.Driver, config.URL)
}

// Connect configures the driver and opens a database connection
func Connect(config Config) (DB, error) {
	db, err := connect(config)
	if err != nil {
		return nil, fmt.Errorf("can't connect to database: %v", err)
	}

	return &prodDB{db: db}, nil
}

// ConnectWriteDatabase creates a new database meant for write operations
func ConnectWriteDatabase(config Config) (WriteDB, error) {
	return Connect(config)
}

// ConnectReadDatabase creates a new database meant for read operations
func ConnectReadDatabase(config Config) (ReadDB, error) {
	return Connect(config)
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

// IsInsuficientPrivilegeError is a helper checking the current database error and returning true if it's a PG insuficient privilege error
func IsInsuficientPrivilegeError(err error) bool {
	e, ok := err.(*pq.Error)
	if !ok {
		return false
	}

	return e.Code == "42501"
}

// IsCheckConstraintError is a helper checking the current database error and returning true if it's a PG check constraint error
func IsCheckConstraintError(err error, constraintName string) bool {
	e, ok := err.(*pq.Error)
	if !ok {
		return false
	}

	return e.Code == "23514" && e.Constraint == constraintName
}

// IsForeignKeyConstraintError is a helper checking the current database error and returning true if it's a PG foreign key constraint error
func IsForeignKeyConstraintError(err error, constraintName string) bool {
	e, ok := err.(*pq.Error)
	if !ok {
		return false
	}

	return e.Code == "23503" && e.Constraint == constraintName
}

// IsEnumInvalidValueError is a helper checking a database error and returns true if it's a invalid input value for a given enum type
func IsEnumInvalidValueError(err error, enumName string) bool {
	e, ok := err.(*pq.Error)
	if !ok {
		return false
	}
	return e.Code == "22P02" && strings.Contains(e.Message, fmt.Sprintf("invalid input value for enum %s", enumName))
}

// GetCurrentTimestamp is a helper function that generates a new UTC timestamp truncated at the millisecond
// because PG is not able to handle nanoseconds
func GetCurrentTimestamp() time.Time {
	return time.Now().UTC().Truncate(time.Millisecond)
}

// Begin starts a new transaction
func (db *prodDB) Begin() (Tx, error) {
	tx, err := db.db.Beginx()
	if err != nil {
		return nil, err
	}

	return &prodTx{tx: tx}, nil
}

// Close closes the connection to the database
func (db *prodDB) Close() error {
	return db.db.Close()
}

// SelectContext fetches a slice of elements from database.
func (db *prodDB) SelectContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error {
	var err error

	metrics.RecordElapsedTimeInMilliseconds(ctx, metricQueryLatencyMs, func() {
		err = db.db.SelectContext(ctx, dest, statement, args...)
	})

	metrics.RecordError(ctx, metricQueryErrorTotal, err)

	return err
}

// SelectMultipleContext fetches a slice of elements from database which match any value from a list.
// This method allows you to write a query with an `in` statement eg:
// SELECT * from table WHERE id IN (?)
func (db *prodDB) SelectMultipleContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error {
	var err error

	metrics.RecordElapsedTimeInMilliseconds(ctx, metricQueryLatencyMs, func() {
		query, queryArguments, statementErr := sqlx.In(statement, args...)
		if statementErr != nil {
			err = fmt.Errorf("an error occured whilst preparing the statement: %v", err)
			return
		}

		query = db.db.Rebind(query)
		err = db.db.SelectContext(ctx, dest, query, queryArguments...)
	})

	metrics.RecordError(ctx, metricQueryErrorTotal, err)

	return err
}

// GetContext fetches one elements from database.
func (db *prodDB) GetContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error {
	var err error

	metrics.RecordElapsedTimeInMilliseconds(ctx, metricQueryLatencyMs, func() {
		err = db.db.GetContext(ctx, dest, statement, args...)
	})

	metrics.RecordError(ctx, metricQueryErrorTotal, err)

	return err
}

// ExecContext executes any SQL query to the server. It's mostly use for insert/update commands
func (db *prodDB) ExecContext(ctx context.Context, statement string, arg ...interface{}) (sql.Result, error) {
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
func (db *prodDB) NamedExecContext(ctx context.Context, statement string, arg interface{}) (sql.Result, error) {
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
func (db *prodDB) PingContext(ctx context.Context) error {
	var err error

	metrics.RecordElapsedTimeInMilliseconds(ctx, metricQueryLatencyMs, func() {
		err = db.db.PingContext(ctx)
	})

	metrics.RecordError(ctx, metricQueryErrorTotal, err)

	return err
}

// Commit persists the transaction
func (tx *prodTx) Commit() error {
	return tx.tx.Commit()
}

// Rollback aborts the transaction
func (tx *prodTx) Rollback() error {
	return tx.tx.Rollback()
}

// GetContext same as db.GetContext but for the current transction
func (tx *prodTx) GetContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error {
	var err error

	metrics.RecordElapsedTimeInMilliseconds(ctx, metricQueryLatencyMs, func() {
		err = tx.tx.GetContext(ctx, dest, statement, args...)
	})

	metrics.RecordError(ctx, metricQueryErrorTotal, err)

	return err
}

// NamedExecContext same as db.NamedExecContext but for the current transction
func (tx *prodTx) NamedExecContext(ctx context.Context, statement string, arg interface{}) (sql.Result, error) {
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

// SelectContext fetches a slice of elements from database.
func (tx *prodTx) SelectContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error {
	var err error

	metrics.RecordElapsedTimeInMilliseconds(ctx, metricQueryLatencyMs, func() {
		err = tx.tx.SelectContext(ctx, dest, statement, args...)
	})

	metrics.RecordError(ctx, metricQueryErrorTotal, err)

	return err
}

// ExecContext executes any SQL query to the server. It's mostly use for insert/update commands
func (tx *prodTx) ExecContext(ctx context.Context, statement string, arg ...interface{}) (sql.Result, error) {
	var (
		response sql.Result
		err      error
	)

	metrics.RecordElapsedTimeInMilliseconds(ctx, metricQueryLatencyMs, func() {
		response, err = tx.tx.ExecContext(ctx, statement, arg...)
	})

	metrics.RecordError(ctx, metricQueryErrorTotal, err)

	return response, err
}
