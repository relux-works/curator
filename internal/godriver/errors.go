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
	// Remedy is the operator remedy for a boundary the operator can resolve
	// from their own host, and is empty for every boundary that has none.
	//
	// It is deliberately not part of Detail. Detail is the protocol string a
	// reader — a conformance vector, a troubleshooting heading, an operator
	// grepping their logs — matches on, and advice that grew into it would
	// move that string every time the advice was reworded. Error renders the
	// remedy behind Detail instead, so the protocol string stays byte for byte
	// what it was and an operator still reads what to do about it.
	Remedy string
	Err    error
}

func (err *Diagnostic) Error() string {
	message := "go-v1 " + err.Code
	if err.Detail != "" {
		message = fmt.Sprintf("go-v1 %s: %s", err.Code, err.Detail)
	}
	if err.Remedy == "" {
		return message
	}
	return message + "; " + err.Remedy
}

func (err *Diagnostic) Unwrap() error { return err.Err }

func diagnostic(code, format string, args ...any) error {
	return &Diagnostic{Code: code, Detail: fmt.Sprintf(format, args...)}
}

func diagnosticErr(code string, err error, format string, args ...any) error {
	return &Diagnostic{Code: code, Detail: fmt.Sprintf(format, args...), Err: err}
}

// diagnosticRemedy builds a boundary diagnostic that carries an operator
// remedy alongside its protocol detail.
func diagnosticRemedy(code, remedy, format string, args ...any) error {
	return &Diagnostic{Code: code, Detail: fmt.Sprintf(format, args...), Remedy: remedy}
}

// diagnosticErrRemedy is diagnosticRemedy for a boundary that also has a cause
// to unwrap.
func diagnosticErrRemedy(code, remedy string, err error, format string, args ...any) error {
	return &Diagnostic{Code: code, Detail: fmt.Sprintf(format, args...), Remedy: remedy, Err: err}
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
