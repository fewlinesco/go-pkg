package tests

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/GuiaBolso/darwin"
	"github.com/fewlinesco/go-pkg/platform/database"
)

func TestProdDatabase(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Fatalf("should not have panicked but got panic with %#v", err)
		}
	}()
	cfgfile, err := os.Open("./testdata/databaseConfig.json")
	if err != nil {
		t.Fatalf("can't open databaseConfig file: %#v", err)
	}

	cfg := database.DefaultConfig

	if err := json.NewDecoder(cfgfile).Decode(&cfg); err != nil {
		t.Fatalf("can't parse file: %#v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("could not connect to DB: %#v, with config: %#v", err, cfg)
	}
	defer func() { db.Close() }()

	err = database.Migrate(db, []darwin.Migration{
		{
			Version:     1,
			Description: "Create test data table",
			Script: `CREATE TABLE test_data(
				id UUID PRIMARY KEY,
				value VARCHAR(63)
			)`,
		},
	})

	defer func() {
		db, err := database.Connect(cfg)
		if err != nil {
			t.Fatalf("could not connect to DB: %#v", err)
		}
		defer db.Close()

		_, err = db.ExecContext(context.Background(), `DROP TABLE IF EXISTS test_data; DROP TABLE IF EXISTS darwin_migrations;`)
		if err != nil {
			t.Fatalf("could not clean the database: %#v", err)
		}
	}()

	if err != nil {
		t.Fatalf("could not migrate the database: %#v", err)
	}

	// recreating a database each time to prove it can persist across different connections
	db, err = database.Connect(cfg)
	if err != nil {
		t.Fatalf("could not connect to the database: %#v", err)
	}

	firstUUID := "ef79f1d4-4150-45ff-b94d-9e4691cc05aa"
	firstValue := "first_value"

	_, err = db.ExecContext(
		context.Background(),
		`INSERT INTO test_data (id, value) VALUES ($1, $2);`,
		firstUUID,
		firstValue,
	)

	if err != nil {
		t.Fatalf("could not ExecContext %#v", err)
	}

	_, err = db.ExecContext(
		context.Background(),
		`INSERT INTO test_data (id, value) VALUES ($1, $2);`,
		firstUUID,
		firstValue,
	)

	if err == nil {
		t.Fatalf("erroring exec don't return an error (most likely panics)")
	}

	type testData struct {
		ID    string `database:"id"`
		Value string `database:"value"`
	}

	secondUUID := "bdc90138-ee0f-456c-8e31-e92514fac45e"
	secondValue := "second_value"

	_, err = db.NamedExecContext(
		context.Background(),
		`INSERT INTO test_data (id, value) VALUES (:id, :value)`,
		testData{ID: secondUUID, Value: secondValue},
	)

	if err != nil {
		t.Fatalf("could not NamedExecContext: %#v", err)
	}
	db.Close()

	// recreating a database each time to prove it can persist across different connections
	db, err = database.Connect(cfg)
	if err != nil {
		t.Fatalf("could not connect to the database: %#v", err)
	}

	var getTestData testData
	err = db.GetContext(
		context.Background(),
		&getTestData,
		`SELECT * from test_data WHERE ID = $1`,
		firstUUID,
	)

	if err != nil {
		t.Fatalf("could not GetContext first test data: %#v", err)
	}

	if getTestData.ID != firstUUID || getTestData.Value != firstValue {
		t.Fatalf("expected test data with ID : %s and Value: %s, but got %#v", firstUUID, firstValue, getTestData)
	}

	db.Close()

	// recreating a database each time to prove it can persist across different connections
	db, err = database.Connect(cfg)
	if err != nil {
		t.Fatalf("could not connect to the database: %#v", err)
	}

	var selectTestData []testData
	err = db.SelectContext(
		context.Background(),
		&selectTestData,
		"SELECT * from test_data",
	)

	if err != nil {
		t.Fatalf("could not SelectContext: %#v", err)
	}

	checkTestData := map[string]bool{
		"first_row":  false,
		"second_row": false,
	}

	for _, row := range selectTestData {
		if row.ID == firstUUID && row.Value == firstValue && checkTestData["first_row"] == false {
			checkTestData["first_row"] = true
		} else if row.ID == secondUUID && row.Value == secondValue && checkTestData["second_row"] == false {
			checkTestData["second_row"] = true
		} else {
			t.Fatalf("Expected selectTestData to be first test data and second test data but got %#v", selectTestData)
		}
	}

	if !checkTestData["first_row"] {
		t.Fatalf("expect selectTestData to carry firstValue but got %#v", selectTestData)
	}

	if !checkTestData["second_row"] {
		t.Fatalf("expect selectTestData to carry secondValue but got %#v", selectTestData)
	}

	db.Close()

	// recreating a database each time to prove it can persist across different connections
	db, err = database.Connect(cfg)
	if err != nil {
		t.Fatalf("could not connect to the database: %#v", err)
	}

	tx, err := db.Begin()

	if err != nil {
		t.Fatalf("could not start a transaction: %#v", err)
	}
	defer func() { tx.Rollback() }()

	_, err = tx.ExecContext(context.Background(), "UPDATE test_data SET value = $1 WHERE id = $2", "updated_value", firstUUID)
	if err != nil {
		t.Fatalf("could not ExecContext in a transaction: %#v", err)
	}

	// This is inserting a primary key already existing in order to corrupt the transaction
	_, err = tx.ExecContext(
		context.Background(),
		`INSERT INTO test_data (id, value) VALUES ($1, $2);`,
		firstUUID,
		firstValue,
	)

	if err == nil {
		t.Fatalf("ExecContext should return an error")
	}

	err = tx.Commit()
	if err == nil {
		t.Fatalf("should not be able to commit faulty transaction")
	}

	err = db.Close()
	if err != nil {
		t.Fatalf("COULD NOT CLOSE THE DATABASE AAAAAAAAAAAH: %v", err)
	}

	db, err = database.Connect(cfg)
	if err != nil {
		t.Fatalf("could not connect to database: %v", err)
	}

	tx, err = db.Begin()

	err = db.GetContext(
		context.Background(),
		&getTestData,
		`SELECT * from test_data WHERE ID = $1`,
		firstUUID,
	)
	if err != nil {
		t.Fatalf("could not use getContext %#v", err)
	}

	if getTestData.Value != firstValue {
		t.Fatalf("A part of the rollbacked transaction has been committed")
	}

	if err != nil {
		t.Fatalf("could not start a transaction: %#v", err)
	}

	thirdUUID := "5afb4fef-4ecb-4ca3-81fa-10306a9cbb60"
	thirdValue := "third_value"
	_, err = db.NamedExecContext(
		context.Background(),
		`INSERT INTO test_data (id, value) VALUES (:id, :value)`,
		testData{ID: thirdUUID, Value: thirdValue},
	)

	if err != nil {
		t.Fatalf("could not shoot named exec context to the database: %#v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("could not commit transaction")
	}

	db.Close()

	// recreating a database each time to prove it can persist across different connections
	db, err = database.Connect(cfg)
	if err != nil {
		t.Fatalf("could not connect to the database: %#v", err)
	}

	err = db.GetContext(
		context.Background(),
		&getTestData,
		`SELECT * from test_data WHERE ID = $1`,
		thirdUUID,
	)
	if err != nil {
		t.Fatalf("could not get data inserted in a transaction %#v", err)
	}
	t.Fatal("muhahaha")

}
