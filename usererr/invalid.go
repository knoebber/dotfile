package usererr

import "fmt"

type invalid struct {
	reason string
}

func (i *invalid) Error() string {
	return fmt.Sprintf("validation error: %s", i.reason)
}

func (i *invalid) Message() string {
	return i.reason
}

// Invalid returns a new invalid error.
func Invalid(reason string) *invalid {
	return &invalid{reason: reason}
}
