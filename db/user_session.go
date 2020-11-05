package db

import (
	"time"

	"github.com/pkg/errors"
)

// UserSession is a model for joining the users and sessions tables.
type UserSession struct {
	Session       string
	IP            string
	UserID        int64
	Username      string
	Email         *string
	CLIToken      string
	Timezone      *string
	Theme         UserTheme
	UserCreatedAt string
}

// Session finds a user with session.
func Session(e Executor, session string) (*UserSession, error) {
	var createdAt time.Time

	res := new(UserSession)

	err := e.QueryRow(`
SELECT session,
       ip,
       user_id,
       username,
       email,
       cli_token,
       timezone,
       theme,
       users.created_at
FROM users
JOIN sessions ON sessions.user_id = users.id
WHERE session = ?
`, session).Scan(
		&res.Session,
		&res.IP,
		&res.UserID,
		&res.Username,
		&res.Email,
		&res.CLIToken,
		&res.Timezone,
		&res.Theme,
		&createdAt,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "querying for user session %q", session)
	}

	res.UserCreatedAt = formatTime(createdAt, res.Timezone)

	return res, nil
}
