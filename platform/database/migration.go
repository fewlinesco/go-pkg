package database

import (
	"fmt"

	"github.com/GuiaBolso/darwin"
	"github.com/jmoiron/sqlx"
)

func Migrate(db *sqlx.DB, migrations []darwin.Migration) error {
	driver := darwin.NewGenericDriver(db.DB, darwin.PostgresDialect{})

	d := darwin.New(driver, migrations, nil)

	if err := d.Migrate(); err != nil {
		return fmt.Errorf("can't migrate: %v", err)
	}

	return nil
}
