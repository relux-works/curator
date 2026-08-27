package godriver

import (
	"errors"
	"fmt"
)

// Diagnostic is a stable, machine-testable failure at a go-v1 trust boundary.
// Detail is intended for operators; callers should branch on Code.
type Diagnostic struct {
	Code   string
	Detail string
	Err    error
}

func (err *Diagnostic) Error() string {
	if err.Detail == "" {
		return "go-v1 " + err.Code
	}
	return fmt.Sprintf("go-v1 %s: %s", err.Code, err.Detail)
}

func (err *Diagnostic) Unwrap() error { return err.Err }

func diagnostic(code, format string, args ...any) error {
	return &Diagnostic{Code: code, Detail: fmt.Sprintf(format, args...)}
}

func diagnosticErr(code string, err error, format string, args ...any) error {
	return &Diagnostic{Code: code, Detail: fmt.Sprintf(format, args...), Err: err}
}

// DiagnosticCode returns the stable go-v1 code carried by err, or an empty
// string when err did not originate at a driver trust boundary.
func DiagnosticCode(err error) string {
	var failure *Diagnostic
	if errors.As(err, &failure) {
		return failure.Code
	}
	return ""
}
