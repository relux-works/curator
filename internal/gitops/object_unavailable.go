package gitops

import (
	"errors"
	"fmt"
	"os/exec"
)

// ObjectUnavailableError records a failed local object read, not a filesystem
// structure violation. ExitCode is -1 when Git could not be started.
type ObjectUnavailableError struct {
	ObjectType string
	ExitCode   int
	Cause      error
}

func (e *ObjectUnavailableError) Error() string {
	return fmt.Sprintf("pinned %s object unavailable: %v", e.ObjectType, e.Cause)
}
func (e *ObjectUnavailableError) Unwrap() error { return e.Cause }

func objectUnavailable(kind string, err error) error {
	code := -1
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		code = exit.ExitCode()
	}
	return &ObjectUnavailableError{ObjectType: kind, ExitCode: code, Cause: err}
}
