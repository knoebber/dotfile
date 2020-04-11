// Package db creates tables when they don't exist and starts a connection to a sqlite database.
package db

import (
	"database/sql"
	"log"
	// Driver for sql
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

// Connection is a sql connection.
var Connection *sql.DB

// Creates the required tables if they doesn't exist
func createTables() error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS users(
                        id         INTEGER PRIMARY KEY,
                        email      TEXT NOT NULL UNIQUE,
                        username   TEXT NOT NULL UNIQUE,
	                created_at DATETIME DEFAULT CURRENT_TIMESTAMP
                 );`,
		`CREATE TABLE IF NOT EXISTS files(
                        id      INTEGER PRIMARY KEY,
                        user_id INTEGER NOT NULL,
	                alias   TEXT NOT NULL,
                        path    TEXT NOT NULL,
                        FOREIGN KEY(user_id) REFERENCES users(id)
                 );`,
		`CREATE TABLE IF NOT EXISTS commits(
                        id                   INTEGER PRIMARY KEY,
                        hash                 TEXT NOT NULL UNIQUE,
                        message              TEXT,
                        file_id              INTEGER NOT NULL,
                        timestamp            DATETIME NOT NULL,
                        FOREIGN KEY(file_id) REFERENCES files(id)
                 );`,
		`CREATE TABLE IF NOT EXISTS revisions(
                        id                     INTEGER PRIMARY KEY,
                        commit_id              INTEGER NOT NULL,
                        contents               BLOB NOT NULL,
                        FOREIGN KEY(commit_id) REFERENCES commits(id)
                 );`,
		`CREATE TABLE IF NOT EXISTS temps(
                        id                   INTEGER PRIMARY KEY,
                        user_id              INTEGER NOT NULL UNIQUE,
                        contents             BLOB NOT NULL,
                        FOREIGN KEY(user_id) REFERENCES users(id)
                 );`,
	}

	for _, table := range tables {
		_, err := Connection.Exec(table)
		if err != nil {
			return errors.Wrap(err, "creating tables")
		}
	}
	return nil
}

// Start opens a connection a sqlite3 database.
// It will create a new sqlite database file if not found.
func Start(dbPath string) (err error) {
	Connection, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	log.Printf("using sqlite3 database %s", dbPath)

	// Foreign keys must be enabled in sqlite.
	// https://sqlite.org/foreignkeys.html
	_, err = Connection.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return err
	}

	return createTables()
}

// Close closes the connection.
func Close() {
	Connection.Close()
}

// Rollback rolls back a database transaction.
// It always returns an error.
func Rollback(tx *sql.Tx, err error) error {
	if rbError := tx.Rollback(); rbError != nil {
		return errors.Wrap(rbError, "rolling back database transaction")
	}

	return errors.Wrap(err, "rolled back from error")
}
