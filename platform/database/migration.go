package database

import (
	"fmt"

	"github.com/GuiaBolso/darwin"
)

// Migrate is a helper function in charge of running pending migrations
func Migrate(db DB, migrations []darwin.Migration) error {
	driver := db.NewGenericDriver(darwin.PostgresDialect{})

	d := darwin.New(driver, migrations, nil)

	if err := d.Migrate(); err != nil {
		return fmt.Errorf("can't migrate: %v", err)
	}

	return nil
}
