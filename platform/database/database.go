package database

import (
	"fmt"
	"net/url"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

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

func Connect(config Config) (*sqlx.DB, error) {
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

	return db, nil
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
