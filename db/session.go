package db

import (
	"database/sql"
	"encoding/base64"
	"time"

	"github.com/pkg/errors"
)

const sessionIDLength = 32

// Session is the model for the sessions table.
// It tracks a users active sessions.
type Session struct {
	ID        int64
	Session   string `validate:"required"`
	UserID    int    `validate:"required"`
	CreatedAt time.Time
}

func (*Session) createStmt() string {
	return `
CREATE TABLE IF NOT EXISTS sessions(
id                   INTEGER PRIMARY KEY,
session              TEXT NOT NULL UNIQUE,
user_id              INTEGER NOT NULL,
created_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY(user_id) REFERENCES users(id)
);`
}

func (s *Session) insertStmt() (sql.Result, error) {
	return connection.Exec("INSERT INTO sessions(session, user_id) VALUES(?, ?)", s.Session, s.UserID)
}

func CreateSession(userID int) (*Session, error) {
	session, err := session()
	if err != nil {
		return nil, err
	}

	s := &Session{
		Session: session,
		UserID:  userID,
	}

	id, err := insert(s)
	s.ID = id

	return s, nil
}

func session() (string, error) {
	buff, err := randomBytes(cliTokenLength)
	if err != nil {
		return "", errors.Wrap(err, "generating session ID")
	}

	return base64.URLEncoding.EncodeToString(buff), nil
}
