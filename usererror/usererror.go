// Package usererror creates errors that are expected from a user making a mistake.
package usererror

import "fmt"

// Reason is the reason that the user error occurred.
type Reason string

// Error reasons.
const (
	ReasonDuplicate = "duplicate"
	ReasonInvalid   = "invalid"
)

// Error returns a message to the user.
type Error struct {
	Message string
	Reason  Reason
}

func (e *Error) Error() string {
	return e.Message
}

// Duplicate creates a new duplicate error.
func Duplicate(field, value string) *Error {
	return &Error{
		Message: fmt.Sprintf("%s %#v already exists", field, value),
		Reason:  ReasonDuplicate,
	}
}

// Invalid returns a new invalid error.
func Invalid(message string) *Error {
	return &Error{
		Message: message,
		Reason:  ReasonInvalid,
	}
}
