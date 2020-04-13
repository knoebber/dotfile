package db

import (
	"crypto/rand"
	"database/sql"
	"io"

	"github.com/pkg/errors"
)

type inserter interface {
	insertStmt() (sql.Result, error)
}

func insert(i inserter) (id int64, err error) {
	if err = validate.Struct(i); err != nil {
		return 0, err
	}

	res, err := i.insertStmt()
	if err != nil {
		return 0, err
	}

	id, err = res.LastInsertId()

	if err != nil {
		return 0, err
	}

	return id, nil
}

// Rolls back a database transaction.
// It always returns an error.
func rollback(tx *sql.Tx, err error) error {
	if rbError := tx.Rollback(); rbError != nil {
		return errors.Wrap(rbError, "rolling back database transaction")
	}

	return errors.Wrap(err, "rolled back from error")
}

// Used for generating random IDs.
func randomBytes(n int) ([]byte, error) {
	buff := make([]byte, n)

	if _, err := io.ReadFull(rand.Reader, buff); err != nil {
		return nil, err
	}

	return buff, nil
}
