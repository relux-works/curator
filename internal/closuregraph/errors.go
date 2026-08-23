package closuregraph

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// DiagnosticCode is a stable machine-readable closure diagnostic.
type DiagnosticCode string

const (
	// CodeGraphSchemaUnsupported and the related constants are stable common
	// graph and checkpoint diagnostics.
	CodeGraphSchemaUnsupported DiagnosticCode = "closure_graph_schema_unsupported"
	// CodeGraphIncomplete reports an incomplete graph or binding.
	CodeGraphIncomplete DiagnosticCode = "closure_graph_incomplete"
	// CodeGraphReferenceInvalid reports a dangling, wrong-kind, or replaced record.
	CodeGraphReferenceInvalid DiagnosticCode = "closure_graph_reference_invalid"
	// CodeDerivationUnauthorized reports an absent pre-C5 derivation authority.
	CodeDerivationUnauthorized DiagnosticCode = "closure_derivation_unauthorized"
	// CodeDerivationDrift reports drift from committed derivation evidence.
	CodeDerivationDrift DiagnosticCode = "closure_derivation_drift"
	// CodeBuildCycle reports a cycle in the execution projection.
	CodeBuildCycle DiagnosticCode = "closure_build_cycle"
	// CodeInteropUndeclared reports a missing explicit interop declaration.
	CodeInteropUndeclared DiagnosticCode = "closure_interop_undeclared"
	// CodeCheckpointInvalid reports an invalid checkpoint schema or chain.
	CodeCheckpointInvalid DiagnosticCode = "closure_checkpoint_invalid"
	// CodeTargetIdentityChanged reports target-platform drift before use.
	CodeTargetIdentityChanged DiagnosticCode = "closure_target_identity_changed"
	// CodeGeneratedInputUndeclared reports selected local artifacts without
	// exact producer lineage before any consuming action or C5 plan exists.
	CodeGeneratedInputUndeclared DiagnosticCode = "artifact_generated_input_undeclared"
)

// Issue is one canonical validation finding. Key is the stable graph key or
// ID used to order otherwise equivalent findings independently of input order.
type Issue struct {
	Code    DiagnosticCode
	Key     string
	Table   string
	Path    string
	Message string
}

// ValidationError contains one or more deterministically ordered findings.
// The first issue is the canonical primary diagnostic.
type ValidationError struct {
	Issues []Issue
}

func (e *ValidationError) Error() string {
	if len(e.Issues) == 0 {
		return "closure graph validation failed"
	}
	issue := e.Issues[0]
	prefix := string(issue.Code)
	if issue.Path != "" {
		prefix += " at " + issue.Path
	}
	if issue.Message == "" {
		return prefix
	}
	return prefix + ": " + issue.Message
}

// Code returns the canonical primary diagnostic code.
func (e *ValidationError) Code() DiagnosticCode {
	if len(e.Issues) == 0 {
		return ""
	}
	return e.Issues[0].Code
}

// ErrorCode extracts a stable closure diagnostic code when one is present.
func ErrorCode(err error) DiagnosticCode {
	var validation *ValidationError
	if errors.As(err, &validation) {
		return validation.Code()
	}
	var cycle *BuildCycleError
	if errors.As(err, &cycle) {
		return CodeBuildCycle
	}
	return ""
}

type issueCollector struct {
	issues []Issue
}

func (c *issueCollector) add(code DiagnosticCode, key, table, path, format string, args ...any) {
	c.issues = append(c.issues, Issue{
		Code: code, Key: key, Table: table, Path: path,
		Message: fmt.Sprintf(format, args...),
	})
}

func (c *issueCollector) err() error {
	if len(c.issues) == 0 {
		return nil
	}
	sort.Slice(c.issues, func(i, j int) bool {
		left, right := c.issues[i], c.issues[j]
		lk := strings.Join([]string{left.Key, left.Table, left.Path, string(left.Code), left.Message}, "\x00")
		rk := strings.Join([]string{right.Key, right.Table, right.Path, string(right.Code), right.Message}, "\x00")
		return lk < rk
	})
	return &ValidationError{Issues: append([]Issue(nil), c.issues...)}
}
