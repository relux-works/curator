// Package verr defines the validation error type shared by protocol parsers.
//
// Every format violation carries the offending field path so callers and
// tests can assert on structure instead of message strings.
package verr

import "fmt"

// Error is a validation error bound to a field path such as
// "dependencies.skills.skill-wiki.ref.kind".
type Error struct {
	Path    string
	Message string
}

func (e *Error) Error() string {
	if e.Path == "" {
		return e.Message
	}
	return e.Path + ": " + e.Message
}

// New builds a validation error for a field path.
func New(path, format string, args ...any) *Error {
	return &Error{Path: path, Message: fmt.Sprintf(format, args...)}
}
