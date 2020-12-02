package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/GuiaBolso/darwin"
	"github.com/fewlinesco/go-pkg/platform/metrics"
	"github.com/jmoiron/sqlx"
)

type sandboxDB struct {
	db *sqlx.DB
	tx *sqlx.Tx
}

type sandboxTx struct {
	tx            *sqlx.Tx
	savepointName string
}

func TestConnect(config Config) (DB, error) {
	db, err := connect(config)
	if err != nil {
		return nil, fmt.Errorf("can't connect to database: %v", err)
	}

	tx, err := db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("could not create the test database transaction: %w", err)
	}

	return &sandboxDB{
		db: db,
		tx: tx,
	}, nil
}

func (db *sandboxDB) Close() error {
	err := db.tx.Rollback()
	closeErr := db.db.Close()
	if err != nil {
		return fmt.Errorf("unable to rollback test transaction: %w", err)
	}
	if closeErr != nil {
		return fmt.Errorf("unable to close database connection: %w", closeErr)
	}
	return nil
}

func (db *sandboxDB) NewGenericDriver(dialect darwin.Dialect) *darwin.GenericDriver {
	panic("test database cannot produce a darwin driver")
}

func (db *sandboxDB) Begin() (Tx, error) {
	fmt.Println("BEGIN")
	db.tx.MustExec("SAVEPOINT postgres_savepoint;")

	return &sandboxTx{tx: db.tx}, nil
}

func (db *sandboxDB) SelectContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error {
	var err error

	metrics.RecordElapsedTimeInMilliseconds(ctx, metricQueryLatencyMs, func() {
		err = db.tx.SelectContext(ctx, dest, statement, args...)
	})

	metrics.RecordError(ctx, metricQueryErrorTotal, err)

	return err
}

func (db *sandboxDB) GetContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error {
	var err error

	metrics.RecordElapsedTimeInMilliseconds(ctx, metricQueryLatencyMs, func() {
		err = db.tx.GetContext(ctx, dest, statement, args...)
	})

	metrics.RecordError(ctx, metricQueryErrorTotal, err)

	return err
}

func (db *sandboxDB) ExecContext(ctx context.Context, statement string, arg ...interface{}) (sql.Result, error) {
	var (
		response sql.Result
		err      error
	)

	metrics.RecordElapsedTimeInMilliseconds(ctx, metricQueryLatencyMs, func() {
		response, err = db.tx.ExecContext(ctx, statement, arg...)
	})

	metrics.RecordError(ctx, metricQueryErrorTotal, err)

	return response, err
}

func (db *sandboxDB) NamedExecContext(ctx context.Context, statement string, arg interface{}) (sql.Result, error) {
	var (
		response sql.Result
		err      error
	)

	metrics.RecordElapsedTimeInMilliseconds(ctx, metricQueryLatencyMs, func() {
		response, err = db.tx.NamedExecContext(ctx, statement, arg)
	})

	metrics.RecordError(ctx, metricQueryErrorTotal, err)

	return response, err
}

func (db *sandboxDB) PingContext(ctx context.Context) error {
	var err error

	metrics.RecordElapsedTimeInMilliseconds(ctx, metricQueryLatencyMs, func() {
		err = db.db.PingContext(ctx)
	})

	metrics.RecordError(ctx, metricQueryErrorTotal, err)

	return err
}

func (tx *sandboxTx) Commit() error {
	fmt.Println("COMMIT")
	tx.tx.MustExec("RELEASE SAVEPOINT postgres_savepoint;")

	return nil
}

func (tx *sandboxTx) Rollback() error {
	fmt.Println("ROLLBACK")
	_, err := tx.tx.Exec("ROLLBACK TO SAVEPOINT postgres_savepoint; RELEASE SAVEPOINT postgres_savepoint;")
	if err != nil {
		return fmt.Errorf("could not rollback to savepoint (test Database transaction rollback emulation) %w", err)
	}

	return nil
}

func (tx *sandboxTx) SelectContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error {
	var err error

	metrics.RecordElapsedTimeInMilliseconds(ctx, metricQueryLatencyMs, func() {
		err = tx.tx.SelectContext(ctx, dest, statement, args...)
	})

	metrics.RecordError(ctx, metricQueryErrorTotal, err)

	return err
}

func (tx *sandboxTx) GetContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error {
	var err error

	metrics.RecordElapsedTimeInMilliseconds(ctx, metricQueryLatencyMs, func() {
		err = tx.tx.GetContext(ctx, dest, statement, args...)
	})

	metrics.RecordError(ctx, metricQueryErrorTotal, err)

	return err
}

func (tx *sandboxTx) ExecContext(ctx context.Context, statement string, arg ...interface{}) (sql.Result, error) {
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

func (tx *sandboxTx) NamedExecContext(ctx context.Context, statement string, arg interface{}) (sql.Result, error) {
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
