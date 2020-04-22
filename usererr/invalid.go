package usererr

import "fmt"

// InvalidErr should be thrown when the user attempts an invalid action.
type InvalidErr struct {
	reason string
}

func (i *InvalidErr) Error() string {
	return fmt.Sprintf("validation error: %s", i.reason)
}

// Message implements the Messager interface.
func (i *InvalidErr) Message() string {
	return i.reason
}

// Invalid returns a new invalid error.
func Invalid(reason string) *InvalidErr {
	return &InvalidErr{reason: reason}
}
