package db

import (
	"database/sql"
	"encoding/base64"
	"time"

	"github.com/pkg/errors"
)

const sessionLength = 24

// Session is the model for the sessions table.
// It tracks a users active sessions.
type Session struct {
	ID        int64
	Session   string `validate:"required"`
	UserID    int64  `validate:"required"`
	CreatedAt time.Time
}

func (*Session) createStmt() string {
	return `
CREATE TABLE IF NOT EXISTS sessions(
id         INTEGER PRIMARY KEY,
session    TEXT NOT NULL UNIQUE,
user_id    INTEGER NOT NULL REFERENCES users,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS sessions_user_index ON sessions(user_id);`
}

func (s *Session) insertStmt() (sql.Result, error) {
	return connection.Exec("INSERT INTO sessions(session, user_id) VALUES(?, ?)", s.Session, s.UserID)
}

func createSession(userID int64) (*Session, error) {
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
	buff, err := randomBytes(sessionLength)
	if err != nil {
		return "", errors.Wrap(err, "generating session ID")
	}

	return base64.URLEncoding.EncodeToString(buff), nil
}
