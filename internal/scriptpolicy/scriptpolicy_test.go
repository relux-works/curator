package scriptpolicy

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/skillspec"
)

func TestAdmitAcceptsDeclaredOnlyCommands(t *testing.T) {
	commands := map[string]skillspec.Command{
		"tool":  {Name: "tool", Type: "script", UnixPath: "scripts/tool"},
		"build": {Name: "build", Type: "build", Driver: "go-v1", SourceDir: "cmd/build"},
		"sys":   {Name: "sys", Type: "system", Command: "git"},
	}
	if err := Admit(commands); err != nil {
		t.Fatalf("declared-only commands were refused: %v", err)
	}
	for name, command := range commands {
		if Enforced(command) {
			t.Fatalf("%s reported as enforced", name)
		}
	}
}

func TestAdmitAdmitsEnforcedCommandWithCompleteTable(t *testing.T) {
	command := skillspec.Command{
		Name: "enforced-tool", Type: "script", UnixPath: "scripts/enforced",
		ExecutionPolicy: skillspec.ScriptExecutionPolicy, Interpreter: "python3-v1",
	}
	if !Enforced(command) {
		t.Fatal("an execution policy did not mark the command enforced")
	}
	// The implemented policy enters the preflight with a complete table,
	// so admission proceeds: only a host that cannot provide a mandatory
	// control refuses, at the install and invoke entries.
	if err := Admit(map[string]skillspec.Command{"enforced-tool": command}); err != nil {
		t.Fatalf("an enforced command was refused at admission: %v", err)
	}
}

func TestHostRefusalCarriesClosedDiagnostic(t *testing.T) {
	err := HostControlError("commands.enforced-tool.execution_policy",
		[]string{"descendant-domain-termination"}, "native control is unavailable on this host")
	if Code(err) != ControlUnavailable {
		t.Fatalf("Code = %q, want %q", Code(err), ControlUnavailable)
	}
	var refusal *Error
	if !errors.As(error(err), &refusal) {
		t.Fatalf("the host refusal did not produce a *scriptpolicy.Error: %T", err)
	}
	if refusal.State != StateUnsupported || refusal.Severity != SeverityError {
		t.Fatalf("state/severity = %q/%q, want %q/%q",
			refusal.State, refusal.Severity, StateUnsupported, SeverityError)
	}
	if refusal.Path != "commands.enforced-tool.execution_policy" {
		t.Fatalf("Path = %q", refusal.Path)
	}
	// The rendered message leads with the closed code so an operator reading a
	// bare install failure sees the diagnostic without a JSON surface.
	if !strings.Contains(err.Error(), ControlUnavailable) {
		t.Fatalf("Error() = %q, want it to name %q", err.Error(), ControlUnavailable)
	}
	// The refusal names every unavailable control, so the operator reads what
	// the host cannot provide rather than a bare verdict.
	if !strings.Contains(refusal.Detail, "descendant-domain-termination") {
		t.Fatalf("Detail = %q, want it to name the unavailable control", refusal.Detail)
	}
}

// An unknown execution policy keeps the original unconditional refusal: the
// manager does not implement it at all, so the preflight never starts and
// the message states the three things the policy forbids.
func TestAdmitRefusesUnknownPolicyUnchanged(t *testing.T) {
	command := skillspec.Command{
		Name: "future-tool", Type: "script", UnixPath: "scripts/future",
		ExecutionPolicy: "script-worker-v9", Interpreter: "python3-v1",
	}
	err := Admit(map[string]skillspec.Command{"future-tool": command})
	if err == nil {
		t.Fatal("an unknown execution policy was admitted")
	}
	if Code(err) != PolicyUnsupported {
		t.Fatalf("Code = %q, want %q", Code(err), PolicyUnsupported)
	}
	var refusal *Error
	if !errors.As(err, &refusal) {
		t.Fatalf("Admit did not produce a *scriptpolicy.Error: %T", err)
	}
	if refusal.State != StateUnsupported || refusal.Severity != SeverityError {
		t.Fatalf("state/severity = %q/%q, want %q/%q",
			refusal.State, refusal.Severity, StateUnsupported, SeverityError)
	}
	if refusal.Path != "commands.future-tool.execution_policy" {
		t.Fatalf("Path = %q", refusal.Path)
	}
	for _, forbidden := range []string{"declared-only", "downgrading", "ignoring"} {
		if !strings.Contains(refusal.Detail, forbidden) {
			t.Fatalf("Detail = %q, want it to name %q", refusal.Detail, forbidden)
		}
	}
}

// The refused command must not depend on Go map iteration order, or the same
// manifest reports a different command on different runs.
func TestAdmitRefusesFirstEnforcedCommandLexically(t *testing.T) {
	commands := map[string]skillspec.Command{
		"zulu":  {Name: "zulu", Type: "script", ExecutionPolicy: "script-worker-v9", Interpreter: "node-v1"},
		"alpha": {Name: "alpha", Type: "script", ExecutionPolicy: "script-worker-v9", Interpreter: "node-v1"},
		"mike":  {Name: "mike", Type: "script", UnixPath: "scripts/mike"},
	}
	for attempt := 0; attempt < 32; attempt++ {
		var refusal *Error
		if !errors.As(Admit(commands), &refusal) {
			t.Fatal("an enforced command was admitted")
		}
		if refusal.Path != "commands.alpha.execution_policy" {
			t.Fatalf("attempt %d: Path = %q", attempt, refusal.Path)
		}
	}
}

func TestCodeIgnoresForeignErrors(t *testing.T) {
	if code := Code(errors.New("unrelated")); code != "" {
		t.Fatalf("Code = %q, want empty", code)
	}
	if code := Code(nil); code != "" {
		t.Fatalf("Code(nil) = %q, want empty", code)
	}
	// A refusal stays recognisable after the wrapping install applies.
	wrapped := fmt.Errorf("skill.command: %w", HostControlError("",
		[]string{"inventory-controls-applied"}, "native control is unavailable on this host"))
	if Code(wrapped) != ControlUnavailable {
		t.Fatalf("Code(wrapped) = %q", Code(wrapped))
	}
	unimplemented := fmt.Errorf("skill.command: %w", Admit(map[string]skillspec.Command{
		"t": {Name: "t", Type: "script", ExecutionPolicy: "script-worker-v9"},
	}))
	if Code(unimplemented) != PolicyUnsupported {
		t.Fatalf("Code(unimplemented) = %q", Code(unimplemented))
	}
}

func TestErrorRendersWithoutPath(t *testing.T) {
	err := &Error{DiagnosticCode: PolicyUnsupported, Detail: "detail"}
	if err.Error() != PolicyUnsupported+": detail" {
		t.Fatalf("Error() = %q", err.Error())
	}
	bare := &Error{DiagnosticCode: PolicyUnsupported}
	if bare.Error() != PolicyUnsupported {
		t.Fatalf("Error() = %q", bare.Error())
	}
}
