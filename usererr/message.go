package usererr

// Messager returns a message to the user.
type Messager interface {
	Message() string
}
