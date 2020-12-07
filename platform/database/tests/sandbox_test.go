package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/fewlinesco/go-pkg/platform/database"
)

func TestSandboxDatabase(t *testing.T) {

	type testData struct {
		ID   string `database:"id"`
		Code string `database:"code"`
	}

	var firstData = testData{ID: "ef79f1d4-4150-45ff-b94d-9e4691cc05aa", Code: "first_value"}
	var secondData = testData{ID: "bdc90138-ee0f-456c-8e31-e92514fac45e", Code: "second_value"}

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

	t.Run("ExecContext", func(t *testing.T) {
		cleanup := migrate(cfg, t)
		defer cleanup()
		sqlxDB, err := connect(cfg)
		if err != nil {
			t.Fatalf("could not create sqlx connection: %#v", err)
		}
		defer sqlxDB.Close()

		_, err = sqlxDB.NamedExecContext(
			context.Background(),
			`INSERT INTO test_data (id, code) VALUES (:id, :code)`,
			firstData,
		)

		if err != nil {
			t.Fatalf("cannot setup test: %#v", err)
		}

		type testCase struct {
			name           string
			data           testData
			shouldErr      bool
			shouldFindData bool
		}

		tcs := []testCase{
			{
				name:           "when everything is fine",
				data:           secondData,
				shouldErr:      false,
				shouldFindData: true,
			},
			{
				name:           "when a constraint is not respected",
				data:           firstData,
				shouldErr:      true,
				shouldFindData: false,
			},
		}

		doTest := func(tc testCase, t *testing.T) {
			t.Run(tc.name, func(t *testing.T) {
				db, err := database.SandboxConnect(cfg)
				if err != nil {
					t.Fatalf("could not connect to the database: %#v", err)
				}
				defer db.Close()

				_, err = db.ExecContext(
					context.Background(),
					`INSERT INTO test_data (id, code) VALUES ($1, $2);`,
					tc.data.ID,
					tc.data.Code,
				)

				if tc.shouldErr {
					if err == nil {
						t.Fatalf("erroring exec don't return an error (most likely panics)")
					}
				} else {
					if err != nil {
						t.Fatalf("could not ExecContext %#v", err)
					}
				}

				if tc.shouldFindData {
					var getTestData testData
					err = db.GetContext(context.Background(), &getTestData, `SELECT * FROM test_data WHERE ID = $1`, tc.data.ID)
					if err != nil {
						t.Fatalf("could not get inserted data : %#v", err)
					}

					if getTestData.ID != tc.data.ID || getTestData.Code != tc.data.Code {
						t.Fatalf("expected test data with ID : %s and Code: %s, but got %#v", tc.data.ID, tc.data.Code, getTestData)
					}
				}

				sqlxDB, err := connect(cfg)
				if err != nil {
					t.Fatalf("could not connect sqlx DB: %#v", err)
				}
				defer sqlxDB.Close()
				var selectTestData []testData
				err = sqlxDB.SelectContext(context.Background(), &selectTestData, `SELECT * FROM test_data;`)
				if err != nil {
					t.Fatalf("could not select test_data: %#v", err)
				}

				if (len(selectTestData) != 1) || (selectTestData[0].ID != firstData.ID) || (selectTestData[0].Code != firstData.Code) {
					t.Fatalf("select all values from testdata should have returned [%#v] but got %#v", firstData, selectTestData)
				}
			})
		}

		for _, tc := range tcs {
			doTest(tc, t)
		}

	})

	t.Run("NamedExecContext", func(t *testing.T) {
		cleanup := migrate(cfg, t)
		defer cleanup()
		sqlxDB, err := connect(cfg)
		if err != nil {
			t.Fatalf("could not create sqlx connection: %#v", err)
		}
		defer sqlxDB.Close()

		_, err = sqlxDB.NamedExecContext(
			context.Background(),
			`INSERT INTO test_data (id, code) VALUES (:id, :code)`,
			testData{ID: firstData.ID, Code: firstData.Code},
		)

		if err != nil {
			t.Fatalf("cannot setup test: %#v", err)
		}

		type testCase struct {
			name           string
			data           testData
			shouldErr      bool
			shouldFindData bool
		}

		tcs := []testCase{
			{
				name:           "when everything is fine",
				data:           secondData,
				shouldErr:      false,
				shouldFindData: true,
			},
			{
				name:           "when a constraint is not respected",
				data:           firstData,
				shouldErr:      true,
				shouldFindData: false,
			},
		}

		doTest := func(tc testCase, t *testing.T) {
			t.Run(tc.name, func(t *testing.T) {
				db, err := database.SandboxConnect(cfg)
				if err != nil {
					t.Fatalf("could not connect to the database: %#v", err)
				}
				defer db.Close()

				_, err = db.NamedExecContext(
					context.Background(),
					`INSERT INTO test_data (id, code) VALUES (:id, :code);`,
					tc.data,
				)

				if tc.shouldErr {
					if err == nil {
						t.Fatalf("erroring exec don't return an error (most likely panics)")
					}
				} else {
					if err != nil {
						t.Fatalf("could not NamedExecContext %#v", err)
					}
				}

				if tc.shouldFindData {
					var getTestData testData
					err = db.GetContext(context.Background(), &getTestData, `SELECT * FROM test_data WHERE ID = $1`, tc.data.ID)
					if err != nil {
						t.Fatalf("could not get inserted data : %#v", err)
					}

					if getTestData.ID != tc.data.ID || getTestData.Code != tc.data.Code {
						t.Fatalf("expected test data with ID : %s and Code: %s, but got %#v", tc.data.ID, tc.data.Code, getTestData)
					}
				}

				sqlxDB, err := connect(cfg)
				if err != nil {
					t.Fatalf("could not connect sqlx DB: %#v", err)
				}

				defer sqlxDB.Close()
				var selectTestData []testData
				err = sqlxDB.SelectContext(context.Background(), &selectTestData, `SELECT * FROM test_data;`)
				if err != nil {
					t.Fatalf("could not select test_data: %#v", err)
				}

				if (len(selectTestData) != 1) || (selectTestData[0].ID != firstData.ID) || (selectTestData[0].Code != firstData.Code) {
					t.Fatalf("select all values from testdata should have returned [%#v] but got %#v", firstData, selectTestData)
				}
			})
		}

		for _, tc := range tcs {
			doTest(tc, t)
		}
	})

	t.Run("GetContext previously inserted data", func(t *testing.T) {
		cleanup := migrate(cfg, t)
		defer cleanup()
		sqlxDB, err := connect(cfg)
		if err != nil {
			t.Fatalf("could not create sqlx connection: %#v", err)
		}
		defer sqlxDB.Close()

		_, err = sqlxDB.NamedExecContext(
			context.Background(),
			`INSERT INTO test_data (id, code) VALUES (:id, :code)`,
			testData{ID: firstData.ID, Code: firstData.Code},
		)

		if err != nil {
			t.Fatalf("cannot setup test: %#v", err)
		}

		type testCase struct {
			name           string
			data           testData
			shouldErr      bool
			shouldFindData bool
		}

		tcs := []testCase{
			{
				name:           "when data has been inserted before",
				data:           firstData,
				shouldErr:      false,
				shouldFindData: true,
			},
			{
				name:           "when a constraint is not respected",
				data:           secondData,
				shouldErr:      true,
				shouldFindData: false,
			},
		}

		doTest := func(tc testCase, t *testing.T) {
			t.Run(tc.name, func(t *testing.T) {
				db, err := database.SandboxConnect(cfg)
				if err != nil {
					t.Fatalf("could not connect to the database: %#v", err)
				}
				defer db.Close()

				var getTestData testData
				err = db.GetContext(context.Background(), &getTestData, `SELECT * FROM test_data WHERE ID = $1`, tc.data.ID)
				if tc.shouldErr {
					if err == nil {
						t.Fatalf("erroring get don't return an error (most likely panics)")
					}
				} else {
					if err != nil {
						t.Fatalf("could not GetContext %#v", err)
					}
				}

				if tc.shouldFindData {
					if getTestData.ID != tc.data.ID || getTestData.Code != tc.data.Code {
						t.Fatalf("expected test data with ID : %s and Code: %s, but got %#v", tc.data.ID, tc.data.Code, getTestData)
					}
				} else {
					if !(reflect.Zero(reflect.TypeOf(getTestData)).Interface() == getTestData) {
						t.Fatalf("getTestData should not have been populated but got %#v", err)
					}
				}
			})
		}

		for _, tc := range tcs {
			doTest(tc, t)
		}
	})

	t.Run("SelectContext previously inserted data", func(t *testing.T) {
		cleanup := migrate(cfg, t)
		defer cleanup()
		sqlxDB, err := connect(cfg)
		if err != nil {
			t.Fatalf("could not create sqlx connection: %#v", err)
		}
		defer sqlxDB.Close()

		_, err = sqlxDB.NamedExecContext(
			context.Background(),
			`INSERT INTO test_data (id, code) VALUES (:id, :code)`,
			firstData,
		)

		if err != nil {
			t.Fatalf("cannot setup test: %#v", err)
		}

		_, err = sqlxDB.NamedExecContext(
			context.Background(),
			`INSERT INTO test_data (id, code) VALUES (:id, :code)`,
			secondData,
		)

		if err != nil {
			t.Fatalf("cannot setup test: %#v", err)
		}

		type testCase struct {
			name      string
			condition *struct {
				sql string
				arg string
			}
			shouldFindData []testData
			shouldErr      bool
		}

		tcs := []testCase{
			{
				name:           "when no condition is provided, it gets all the data",
				shouldFindData: []testData{firstData, secondData},
				shouldErr:      false,
			},
			{
				name: "when a condition is provided it only gets the requested data",
				condition: &struct {
					sql string
					arg string
				}{sql: "WHERE id = $1", arg: firstData.ID},
				shouldFindData: []testData{firstData},
				shouldErr:      false,
			},
			{
				name: "when no data is matching the condition, it returns an empty slice with no error",
				condition: &struct {
					sql string
					arg string
				}{sql: "WHERE id = $1", arg: "3237b466-b3c6-4521-96c5-61022c4a1796"},
				shouldFindData: []testData{},
				shouldErr:      false,
			},
			{
				name: "when the condition is faulty, it does not populate the slice and return an error",
				condition: &struct {
					sql string
					arg string
				}{sql: "WHERE non_exisiting_field = ", arg: firstData.ID},
				shouldFindData: []testData{},
				shouldErr:      true,
			},
		}

		doTest := func(tc testCase, t *testing.T) {
			t.Run(tc.name, func(t *testing.T) {
				db, err := database.SandboxConnect(cfg)
				if err != nil {
					t.Fatalf("could not connect to the database: %#v", err)
				}
				defer db.Close()

				var selectTestData []testData
				if tc.condition == nil {
					err = db.SelectContext(context.Background(), &selectTestData, `SELECT * FROM test_data;`)
				} else {
					err = db.SelectContext(context.Background(), &selectTestData, fmt.Sprintf("%s %s;", `SELECT * FROM test_data`, tc.condition.sql), tc.condition.arg)
				}

				if tc.shouldErr {
					if err == nil {
						t.Fatalf("erroring select don't return an error (most likely panics)")
					}
				} else {
					if err != nil {
						t.Fatalf("could not SelectContext %#v", err)
					}
				}

				if len(tc.shouldFindData) != len(selectTestData) {
					t.Fatalf("expected tc.shouldFindData and selectTestData to have the same length but got tc.shouldFindData: %#v, selectTestData: %#v", tc.shouldFindData, selectTestData)
				}

				for _, sfd := range tc.shouldFindData {
					found := false
					for _, std := range selectTestData {
						if std.ID == sfd.ID {
							found = true
						}
					}
					if !found {
						t.Fatalf("should find %#v in selectTestData but got %#v", sfd, selectTestData)
					}
				}
			})
		}

		for _, tc := range tcs {
			doTest(tc, t)
		}
	})

	t.Run("Transactions Commit", func(t *testing.T) {
		type testCase struct {
			name        string
			transaction func(tx database.Tx, t *testing.T)
			shouldErr   bool
			data        []testData
		}

		tcs := []testCase{
			{
				name: "when everything works in the transaction it can be commited",
				transaction: func(tx database.Tx, t *testing.T) {
					_, err = tx.NamedExecContext(
						context.Background(),
						`INSERT INTO test_data (id, code) VALUES (:id, :code)`,
						secondData,
					)
				},
				shouldErr: false,
				data:      []testData{firstData, secondData},
			},
			{
				name: "when something fails in the transaction it cannot be commited",
				transaction: func(tx database.Tx, t *testing.T) {
					_, err = tx.NamedExecContext(
						context.Background(),
						`INSERT INTO test_data (id, code) VALUES (:id, :code)`,
						firstData,
					)
				},
				shouldErr: true,
				data:      []testData{firstData},
			},
			{
				name: "when everything works in the transaction but it has been manually rollbacked it cannot be commited",
				transaction: func(tx database.Tx, t *testing.T) {
					_, err = tx.NamedExecContext(
						context.Background(),
						`INSERT INTO test_data (id, code) VALUES (:id, :code)`,
						secondData,
					)
					tx.Rollback()
				},
				shouldErr: true,
				data:      []testData{firstData},
			},
			{
				name: "when a transaction has already been commited it cannot be commited again",
				transaction: func(tx database.Tx, t *testing.T) {
					_, err = tx.NamedExecContext(
						context.Background(),
						`INSERT INTO test_data (id, code) VALUES (:id, :code)`,
						secondData,
					)
					tx.Commit()
				},
				shouldErr: true,
				data:      []testData{firstData, secondData},
			},
		}
		doTest := func(tc testCase, t *testing.T) {
			t.Run(tc.name, func(t *testing.T) {
				cleanup := migrate(cfg, t)
				defer cleanup()
				sqlxDB, err := connect(cfg)
				if err != nil {
					t.Fatalf("could not create sqlx connection: %#v", err)
				}
				defer sqlxDB.Close()

				_, err = sqlxDB.NamedExecContext(
					context.Background(),
					`INSERT INTO test_data (id, code) VALUES (:id, :code)`,
					firstData,
				)

				if err != nil {
					t.Fatalf("cannot setup test: %#v", err)
				}

				db, err := database.SandboxConnect(cfg)
				if err != nil {
					t.Fatalf("could not connect to the database: %#v", err)
				}

				tx, err := db.Begin()
				if err != nil {
					t.Fatalf("could not start a transaction: %#v", err)
				}

				tc.transaction(tx, t)
				err = tx.Commit()
				if tc.shouldErr {
					if err == nil {
						t.Fatalf("commit should throw an error but err was nil")
					}
				} else {
					if err != nil {
						t.Fatalf("commit shouldn't return an error but returned: %#v", err)
					}
				}

				var selectTestData []testData
				err = db.SelectContext(context.Background(), &selectTestData, `SELECT * FROM test_data;`)
				if err != nil {
					t.Fatalf("could not fetch data from the sandboxedDB: %#v", err)
				}

				if len(tc.data) != len(selectTestData) {
					t.Fatalf("expected tc.data and selectTestData to have the same length but got tc.data: %#v, selectTestData: %#v", tc.data, selectTestData)
				}

				for _, sfd := range tc.data {
					found := false
					for _, std := range selectTestData {
						if std.ID == sfd.ID {
							found = true
						}
					}
					if !found {
						t.Fatalf("should find %#v in selectTestData but got %#v", sfd, selectTestData)
					}
				}
				db.Close()

				sqlxDB, err = connect(cfg)
				if err != nil {
					t.Fatalf("could not connect sqlx DB: %#v", err)
				}
				defer sqlxDB.Close()
				var previouslyInsertedData []testData
				err = sqlxDB.SelectContext(context.Background(), &previouslyInsertedData, `SELECT * FROM test_data;`)
				if err != nil {
					t.Fatalf("could not select test_data: %#v", err)
				}

				if (len(previouslyInsertedData) != 1) || (previouslyInsertedData[0].ID != firstData.ID) || (previouslyInsertedData[0].Code != firstData.Code) {
					t.Fatalf("select all values from testdata should have returned [%#v] but got %#v", firstData, previouslyInsertedData)
				}
			})
		}
		for _, tc := range tcs {
			doTest(tc, t)
		}
	})

	t.Run("Transactions Rollback", func(t *testing.T) {
		type testCase struct {
			name        string
			transaction func(tx database.Tx, t *testing.T)
			shouldErr   bool
			data        []testData
		}

		tcs := []testCase{
			{
				name: "when a transaction is rollbacked data is not saved",
				transaction: func(tx database.Tx, t *testing.T) {
					_, err = tx.NamedExecContext(
						context.Background(),
						`INSERT INTO test_data (id, code) VALUES (:id, :code)`,
						secondData,
					)
				},
				shouldErr: false,
				data:      []testData{firstData},
			},
			{
				name: "when something fails in the transaction it can be rollbacked",
				transaction: func(tx database.Tx, t *testing.T) {
					_, err = tx.NamedExecContext(
						context.Background(),
						`INSERT INTO test_data (id, code) VALUES (:id, :code)`,
						firstData,
					)
				},
				shouldErr: false,
				data:      []testData{firstData},
			},
			{
				name: "when a transaction is already committed, rollbacking has no effect",
				transaction: func(tx database.Tx, t *testing.T) {
					_, err = tx.NamedExecContext(
						context.Background(),
						`INSERT INTO test_data (id, code) VALUES (:id, :code)`,
						secondData,
					)
					tx.Commit()
				},
				shouldErr: true,
				data:      []testData{firstData, secondData},
			},
			{
				name: "when a transaction has already been rollbacked it cannot be rollbacked again",
				transaction: func(tx database.Tx, t *testing.T) {
					_, err = tx.NamedExecContext(
						context.Background(),
						`INSERT INTO test_data (id, code) VALUES (:id, :code)`,
						secondData,
					)
					tx.Rollback()
				},
				shouldErr: true,
				data:      []testData{firstData},
			},
		}
		doTest := func(tc testCase, t *testing.T) {
			t.Run(tc.name, func(t *testing.T) {
				cleanup := migrate(cfg, t)
				defer cleanup()
				sqlxDB, err := connect(cfg)
				if err != nil {
					t.Fatalf("could not create sqlx connection: %#v", err)
				}
				defer sqlxDB.Close()

				_, err = sqlxDB.NamedExecContext(
					context.Background(),
					`INSERT INTO test_data (id, code) VALUES (:id, :code)`,
					firstData,
				)

				if err != nil {
					t.Fatalf("cannot setup test: %#v", err)
				}

				db, err := database.SandboxConnect(cfg)
				if err != nil {
					t.Fatalf("could not connect to the database: %#v", err)
				}

				tx, err := db.Begin()
				if err != nil {
					t.Fatalf("could not start a transaction: %#v", err)
				}

				tc.transaction(tx, t)
				err = tx.Rollback()
				if tc.shouldErr {
					if err == nil {
						t.Fatalf("rollback should throw an error but err was nil")
					}
				} else {
					if err != nil {
						t.Fatalf("rollback shouldn't return an error but returned: %#v", err)
					}
				}

				var selectTestData []testData
				err = db.SelectContext(context.Background(), &selectTestData, `SELECT * FROM test_data;`)
				if err != nil {
					t.Fatalf("could not fetch data from the sqlxDB: %#v", err)
				}

				if len(tc.data) != len(selectTestData) {
					t.Fatalf("expected tc.data and selectTestData to have the same length but got tc.data: %#v, selectTestData: %#v", tc.data, selectTestData)
				}

				for _, sfd := range tc.data {
					found := false
					for _, std := range selectTestData {
						if std.ID == sfd.ID {
							found = true
						}
					}
					if !found {
						t.Fatalf("should find %#v in selectTestData but got %#v", sfd, selectTestData)
					}
				}

				db.Close()

				sqlxDB, err = connect(cfg)
				if err != nil {
					t.Fatalf("could not connect sqlx DB: %#v", err)
				}

				defer sqlxDB.Close()
				var previouslyInsertedData []testData
				err = sqlxDB.SelectContext(context.Background(), &previouslyInsertedData, `SELECT * FROM test_data;`)
				if err != nil {
					t.Fatalf("could not select test_data: %#v", err)
				}

				if (len(previouslyInsertedData) != 1) || (previouslyInsertedData[0].ID != firstData.ID) || (previouslyInsertedData[0].Code != firstData.Code) {
					t.Fatalf("select all values from testdata should have returned [%#v] but got %#v", firstData, previouslyInsertedData)
				}

			})
		}
		for _, tc := range tcs {
			doTest(tc, t)
		}
	})
}
