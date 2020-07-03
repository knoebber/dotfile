// Package db creates tables when they don't exist and starts a connection to a sqlite database.
package db

import (
	"database/sql"

	// Driver for sql
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"gopkg.in/go-playground/validator.v9"
)

type tableCreator interface {
	createStmt() string
}

var (
	connection *sql.DB
	validate   *validator.Validate
)

// Creates the required tables if they doesn't exist
func createTables() error {
	for _, model := range []tableCreator{
		new(User),
		new(ReservedUsername),
		new(Session),
		new(SessionLocation),
		new(File),
		new(TempFile),
		new(Commit),
	} {
		_, err := connection.Exec(model.createStmt())
		if err != nil {
			return errors.Wrap(err, "creating tables")
		}
	}
	return nil
}

// Start opens a connection a sqlite3 database.
// It will create a new sqlite database file if not found.
func Start(dbPath string) (err error) {
	dsn := "?_foreign_keys=true"
	connection, err = sql.Open("sqlite3", dbPath+dsn)
	if err != nil {
		return err
	}

	validate = validator.New()
	return createTables()
}

// Close closes the connection.
func Close() {
	connection.Close()
}
