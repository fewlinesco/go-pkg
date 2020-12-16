package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/GuiaBolso/darwin"
	"github.com/jmoiron/sqlx"
)

type sandboxDB struct {
	db *sqlx.DB
	tx *sqlx.Tx
}

type sandboxTx struct {
	tx                    *sqlx.Tx
	rollBackedOrCommitted bool
}

// SandboxConnect returns a sandboxed database connection from the given configuration
func SandboxConnect(config Config) (DB, error) {
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
	db.tx.MustExec("SAVEPOINT go_pkg_database_sandbox_savepoint;")

	return &sandboxTx{tx: db.tx}, nil
}

func (db *sandboxDB) SelectContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error {

	return db.tx.SelectContext(ctx, dest, statement, args...)

}

func (db *sandboxDB) SelectMultipleContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error {
	query, queryArguments, err := sqlx.In(statement, args...)
	if err != nil {
		return fmt.Errorf("an error occured whilst preparing the statement: %v", err)
	}

	query = db.db.Rebind(query)

	return db.db.SelectContext(ctx, dest, statement, queryArguments)
}

func (db *sandboxDB) GetContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error {

	return db.tx.GetContext(ctx, dest, statement, args...)

}

func (db *sandboxDB) ExecContext(ctx context.Context, statement string, arg ...interface{}) (sql.Result, error) {

	return db.tx.ExecContext(ctx, statement, arg...)

}

func (db *sandboxDB) NamedExecContext(ctx context.Context, statement string, arg interface{}) (sql.Result, error) {

	return db.tx.NamedExecContext(ctx, statement, arg)

}

func (db *sandboxDB) PingContext(ctx context.Context) error {

	return db.db.PingContext(ctx)

}

func (tx *sandboxTx) Commit() error {
	if tx.rollBackedOrCommitted {
		return fmt.Errorf("transaction has already been rollbacked or commited")
	}

	_, err := tx.tx.Exec("RELEASE SAVEPOINT go_pkg_database_sandbox_savepoint;")
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			panic(fmt.Sprintf("tried to commit but got: %#v; transaction had to be rollbacked but got: %#v", err, rollbackErr))
		}
		return fmt.Errorf("could not commit savepoint (database sandbox transaction commit emulation): %#w", err)
	}

	tx.rollBackedOrCommitted = true

	return nil
}

func (tx *sandboxTx) Rollback() error {
	if tx.rollBackedOrCommitted {
		return fmt.Errorf("transaction has already been rollbacked or commited")
	}

	_, err := tx.tx.Exec("ROLLBACK TO SAVEPOINT go_pkg_database_sandbox_savepoint; RELEASE SAVEPOINT go_pkg_database_sandbox_savepoint;")
	if err != nil {
		return fmt.Errorf("could not rollback to savepoint (database sandbox transaction rollback emulation) %w", err)
	}

	tx.rollBackedOrCommitted = true

	return nil
}

func (tx *sandboxTx) SelectContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error {

	return tx.tx.SelectContext(ctx, dest, statement, args...)

}

func (tx *sandboxTx) GetContext(ctx context.Context, dest interface{}, statement string, args ...interface{}) error {
	return tx.tx.GetContext(ctx, dest, statement, args...)
}

func (tx *sandboxTx) ExecContext(ctx context.Context, statement string, arg ...interface{}) (sql.Result, error) {
	return tx.tx.ExecContext(ctx, statement, arg...)
}

func (tx *sandboxTx) NamedExecContext(ctx context.Context, statement string, arg interface{}) (sql.Result, error) {
	return tx.tx.NamedExecContext(ctx, statement, arg)
}
