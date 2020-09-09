package cqrs_application_test

import (
	"context"
	"testing"

	"github.com/GuiaBolso/darwin"
	"github.com/fewlinesco/go-pkg/platform"
)

type cqrsTest struct {
	ID    int    `db:"id"`
	Value string `db:"value"`
}

func TestCQRSApplication(t *testing.T) {
	// SETUP
	cleanMigrationQuery := `DROP TABLE IF EXISTS darwin_migrations`

	migrations := []darwin.Migration{
		{
			Version:     1,
			Description: "Create the cqrs test table",
			Script: `CREATE TABLE cqrs_test (
				id SERIAL,
				value VARCHAR NOT NULL
			);`,
		},
	}

	cqrsAppConfig := platform.DefaultCQRSApplicationConfig

	err := platform.ReadConfiguration("../../../configs/cqrs_config.json", &cqrsAppConfig)
	if err != nil {
		t.Fatalf("Could not read the configuration: %v", err)
	}

	cqrsApplication, err := platform.NewCQRSApplication(cqrsAppConfig)
	if err != nil {
		t.Fatalf("Could not create the application: %v", err)
	}

	_, err = cqrsApplication.WriteDatabase.ExecContext(context.Background(), cleanMigrationQuery)
	if err != nil {
		t.Fatalf("Unable to clean migration table: %v", err)
	}

	err = cqrsApplication.StartMigrations(migrations)
	if err != nil {
		t.Fatalf("could not migrate the Database: %v", err)
	}

	// TESTING
	t.Run("testCQRS", func(t *testing.T) {
		writeQuery := `INSERT INTO cqrs_test (value) VALUES(:value)`
		_, err = cqrsApplication.WriteDatabase.NamedExecContext(context.Background(), writeQuery, cqrsTest{Value: "test"})
		if err != nil {
			t.Fatalf("Unable to insert with the Write database: %v", err)
		}

		var testRead cqrsTest
		readQuery := `SELECT * FROM cqrs_test WHERE id = 1`

		err = cqrsApplication.ReadDatabase.GetContext(context.Background(), &testRead, readQuery)
		if err != nil {
			t.Fatalf("Unable to read with the Read database: %v", err)
		}

		_, err = cqrsApplication.ReadDatabase.NamedExecContext(context.Background(), writeQuery, cqrsTest{Value: "test 2"})
		if err == nil {
			t.Fatalf("Should not be able to write with the Read database")
		}
	})

	// CLEANING
	migrations = append(migrations, darwin.Migration{
		Version:     2,
		Description: "Delete the cqrs test table",
		Script:      `DROP TABLE cqrs_test`,
	})

	err = cqrsApplication.StartMigrations(migrations)
	if err != nil {
		t.Fatalf("could not migrate the Database: %v", err)
	}

	_, err = cqrsApplication.WriteDatabase.ExecContext(context.Background(), cleanMigrationQuery)
	if err != nil {
		t.Fatalf("Unable to clean migration table: %v", err)
	}

}
