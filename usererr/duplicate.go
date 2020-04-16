package usererr

import "fmt"

type duplicate struct {
	field string
	value string
}

func (d *duplicate) Error() string {
	return fmt.Sprintf("duplication error: %s", d.Message())
}

func (d *duplicate) Message() string {
	return fmt.Sprintf("%s %#v already exists", d.field, d.value)
}

// Duplicate returns a new duplicat error.
// This should be used when a unique field already has value.
func Duplicate(field, value string) *duplicate {
	return &duplicate{field: field, value: value}
}
