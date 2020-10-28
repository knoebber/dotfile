package db

import (
	"database/sql"
	"encoding/base64"
	"time"

	"github.com/pkg/errors"
)

const sessionLength = 24

// SessionRecord models the sessions table.
// It tracks a user's active sessions.
type SessionRecord struct {
	ID        int64
	Session   string `validate:"required"`
	UserID    int64  `validate:"required"`
	IP        string
	CreatedAt time.Time
	DeletedAt *time.Time
}

func (*SessionRecord) createStmt() string {
	return `
CREATE TABLE IF NOT EXISTS sessions(
id         INTEGER PRIMARY KEY,
session    TEXT NOT NULL UNIQUE,
user_id    INTEGER NOT NULL REFERENCES users,
ip         TEXT NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
deleted_at DATETIME
);

CREATE INDEX IF NOT EXISTS sessions_user_index ON sessions(user_id);`
}

func (s *SessionRecord) insertStmt(e Executor) (sql.Result, error) {
	return e.Exec("INSERT INTO sessions(session, user_id, ip) VALUES(?, ?, ?)", s.Session, s.UserID, s.IP)
}

func createSession(e Executor, username, ip string) (*SessionRecord, error) {
	var userID int64

	err := e.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		return nil, errors.Wrapf(err, "querying for userID from %q", username)
	}

	session, err := session()
	if err != nil {
		return nil, err
	}

	s := &SessionRecord{
		Session: session,
		UserID:  userID,
		IP:      ip,
	}

	id, err := insert(e, s)
	if err != nil {
		return nil, err
	}

	s.ID = id

	return s, nil
}

func session() (string, error) {
	buff, err := randomBytes(sessionLength)
	if err != nil {
		return "", errors.Wrap(err, "generating session")
	}

	return base64.URLEncoding.EncodeToString(buff), nil
}

// Logout sets the session to deleted.
func Logout(e Executor, session string) error {
	_, err := e.Exec("UPDATE sessions SET deleted_at = ? WHERE session = ?", time.Now(), session)
	if err != nil {
		return errors.Wrapf(err, "setting session %q to deleted", session)
	}

	return nil
}
