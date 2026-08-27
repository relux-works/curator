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
	// Remedy is the operator action that resolves this boundary, and is empty
	// for every boundary an operator cannot act on from their own host.
	//
	// It is kept out of Detail on purpose. Detail is the protocol string a
	// reader matches on — a conformance vector, an operator grepping a log —
	// and advice folded into it would move that string every time the advice
	// was reworded. Error renders the remedy behind Detail instead, so the
	// protocol string stays byte for byte what it was while the operator still
	// reads what to do about it.
	Remedy string
	Err    error
}

func (err *Diagnostic) Error() string {
	message := "go-v1 " + err.Code
	if err.Detail != "" {
		message += ": " + err.Detail
	}
	if err.Remedy != "" {
		message += "; " + err.Remedy
	}
	return message
}

func (err *Diagnostic) Unwrap() error { return err.Err }

func diagnostic(code, format string, args ...any) error {
	return &Diagnostic{Code: code, Detail: fmt.Sprintf(format, args...)}
}

func diagnosticErr(code string, err error, format string, args ...any) error {
	return &Diagnostic{Code: code, Detail: fmt.Sprintf(format, args...), Err: err}
}

// diagnosticRemedy is diagnostic for a boundary that carries an operator
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
