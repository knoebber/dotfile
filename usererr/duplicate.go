package usererr

import "fmt"

// DuplicateErr should be thrown when a unique field already has value.
type DuplicateErr struct {
	field string
	value string
}

func (d *DuplicateErr) Error() string {
	return fmt.Sprintf("duplication error: %s", d.Message())
}

// Message implements the Messager interface.
func (d *DuplicateErr) Message() string {
	return fmt.Sprintf("%s %#v already exists", d.field, d.value)
}

// Duplicate creates a new duplicate error.
func Duplicate(field, value string) *DuplicateErr {
	return &DuplicateErr{field: field, value: value}
}
