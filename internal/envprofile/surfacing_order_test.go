package envprofile

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// This file observes the §2.3 emission order through the production entry
// points: the surfacing rows must reach the sink after the audit gate
// passes and before the lock is published or any surface is
// (re-)materialized. A test that only inspects Info.Surfacing after the
// operation returns cannot see the order; these tests watch the
// filesystem while the rows arrive.

const orderToolManifest = `{"schema_version": 1, "name": "tool", "version": "1.0.0",` +
	`"server": {"transport": "stdio", "command": "npx", "args": ["-y", "tool"], "env_names": ["TOOL_TOKEN"]}}` + "\n"

const orderToolRow = `mcp-declaration tool 1.0.0 stdio command=npx args=["-y","tool"] env_names=["TOOL_TOKEN"]`

// lockObservingSink records every §2.3 row it receives and snapshots the
// profile lock's filesystem state when the first row arrives: whether
// lock.json exists and its exact bytes at that instant.
type lockObservingSink struct {
	lock     string
	rows     []string
	pending  string
	observed bool
	existed  bool
	bytes    []byte
}

func (w *lockObservingSink) Write(p []byte) (int, error) {
	w.pending += string(p)
	for {
		line, rest, ok := strings.Cut(w.pending, "\n")
		if !ok {
			break
		}
		w.pending = rest
		if !strings.HasPrefix(line, "mcp-declaration ") {
			continue
		}
		if !w.observed {
			w.observed = true
			payload, err := os.ReadFile(w.lock) // #nosec G304 -- test lock path
			w.existed = err == nil
			w.bytes = payload
		}
		w.rows = append(w.rows, line)
	}
	return len(p), nil
}

// serveOrderTool serves one stdio MCP declaration under a fake identity
// and returns the requirement operand for it.
func serveOrderTool(t *testing.T, ids *gitIdentities) string {
	t.Helper()
	tool := gitRepo(t, map[string]string{"agent-mcp.json": orderToolManifest}, "v1.0.0")
	return ids.serve(tool, "https://example.com/mcp-tool")
}

// serveOrderRoot serves a root requiring the tool under a fake identity
// and returns the install operand for it.
func serveOrderRoot(t *testing.T, ids *gitIdentities, name, toolOperand string) string {
	t.Helper()
	root := gitRepo(t, map[string]string{
		"agent-context.json": `{"schema_version": 1, "name": "` + name + `", "version": "1.0.0",` +
			`"requires": {"mcp": {"tool": {"git": "` + toolOperand + `", "range": "*"}}}}` + "\n",
	}, "v1.0.0")
	return ids.serve(root, "https://example.com/"+name)
}

// requireSingleEmission pins the exactly-once contract: the sink carries
// the want rows in order, and Info carries none with a sink set.
func requireSingleEmission(t *testing.T, sink *lockObservingSink, info Info, want []string) {
	t.Helper()
	if !sink.observed {
		t.Fatal("no §2.3 row reached the sink")
	}
	if !equalOrdered(sink.rows, want) {
		t.Fatalf("emitted rows = %v, want %v", sink.rows, want)
	}
	if len(info.Surfacing) != 0 {
		t.Fatalf("Info.Surfacing = %v with a sink set: the rows would print twice", info.Surfacing)
	}
}

// observeLiveInstallEmission installs an MCP-carrying root with an
// observing sink and asserts the §2.3 emission order: the exact row
// arrives while lock.json is still absent, and Info carries no rows to
// reprint. Shared by the order tests and the surfacing_order vector
// driver, which must observe the runtime order rather than the shape of
// a helper's output.
func observeLiveInstallEmission(t *testing.T) (home, operand string) {
	t.Helper()
	home = t.TempDir()
	pinHomes(t)
	ids := newGitIdentities(t)
	operand = serveOrderRoot(t, ids, "withmcp", serveOrderTool(t, ids))
	sink := &lockObservingSink{lock: lockPath(home, "withmcp")}
	info, _, _, err := Install(home, InstallOptions{Operand: operand, SurfacingSink: sink})
	if err != nil {
		t.Fatal(err)
	}
	requireSingleEmission(t, sink, info, []string{orderToolRow})
	if sink.existed {
		t.Fatalf("lock.json already exists when the first mcp-declaration row is written: %s", sink.lock)
	}
	assertWarningPresent(t, info.Warnings, DiagAllowlistEmpty)
	return home, operand
}

// observeLiveUpdateEmission installs a declaration-free root, moves it to
// a revision carrying a declaration, and asserts the §2.3 emission
// order: the update moves, the row arrives while lock.json still holds
// the old lock bytes, and Info carries no rows to reprint. An unchanged
// update then still surfaces the candidate set exactly once. Shared by
// the order tests and the surfacing_order vector driver.
func observeLiveUpdateEmission(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	pinHomes(t)
	ids := newGitIdentities(t)
	toolOperand := serveOrderTool(t, ids)
	root := t.TempDir()
	writeGitFile(t, root, "agent-context.json", `{"schema_version": 1, "name": "grows", "version": "1.0.0"}`+"\n")
	gitRun(t, root, "init")
	gitRun(t, root, "add", ".")
	gitRun(t, root, "commit", "-m", "one")
	gitRun(t, root, "tag", "v1.0.0")
	if _, _, _, err := Install(home, InstallOptions{Operand: ids.serve(root, "https://example.com/grows")}); err != nil {
		t.Fatal(err)
	}
	oldBytes, err := os.ReadFile(lockPath(home, "grows")) // #nosec G304 -- test lock path
	if err != nil {
		t.Fatal(err)
	}
	writeGitFile(t, root, "agent-context.json", `{"schema_version": 1, "name": "grows", "version": "1.0.1",`+
		`"requires": {"mcp": {"tool": {"git": "`+toolOperand+`", "range": "*"}}}}`+"\n")
	gitRun(t, root, "add", ".")
	gitRun(t, root, "commit", "-m", "two")
	gitRun(t, root, "tag", "v1.0.1")
	sink := &lockObservingSink{lock: lockPath(home, "grows")}
	info, moved, err := UpdateWithOptions(home, "grows", UpdateOptions{SurfacingSink: sink})
	if err != nil {
		t.Fatal(err)
	}
	if !moved {
		t.Fatal("the update must move to the revision carrying the declaration")
	}
	requireSingleEmission(t, sink, info, []string{orderToolRow})
	if !sink.existed {
		t.Fatal("lock.json is absent when the update emits: the old lock must still stand")
	}
	if !bytes.Equal(sink.bytes, oldBytes) {
		t.Fatalf("lock.json at first row = %q, want the old lock bytes", sink.bytes)
	}
	assertWarningPresent(t, info.Warnings, DiagAllowlistEmpty)
	// An unchanged update publishes nothing but still surfaces the
	// candidate set exactly once.
	restill := &lockObservingSink{lock: lockPath(home, "grows")}
	restillInfo, moved, err := UpdateWithOptions(home, "grows", UpdateOptions{SurfacingSink: restill})
	if err != nil {
		t.Fatal(err)
	}
	if moved {
		t.Fatal("an update with no new revision must not move")
	}
	requireSingleEmission(t, restill, restillInfo, []string{orderToolRow})
}

// TestInstallEmitsSurfacingBeforePublication observes the live install
// emission order, then covers the same-source git reinstall: it
// delegates to the update path with the same sink, so the rows still
// emit exactly once, describing the lock that stands.
func TestInstallEmitsSurfacingBeforePublication(t *testing.T) {
	home, operand := observeLiveInstallEmission(t)
	resink := &lockObservingSink{lock: lockPath(home, "withmcp")}
	reinfo, _, updated, err := Install(home, InstallOptions{Operand: operand, SurfacingSink: resink})
	if err != nil {
		t.Fatal(err)
	}
	if !updated {
		t.Fatal("a same-source reinstall must report updated")
	}
	requireSingleEmission(t, resink, reinfo, []string{orderToolRow})
}

// TestUpdateEmitsSurfacingBeforePublication observes the live update
// emission order with old-lock identity.
func TestUpdateEmitsSurfacingBeforePublication(t *testing.T) {
	observeLiveUpdateEmission(t)
}

// TestReinstallEmitsSurfacingBeforePublication drives a changed-path
// reinstall of a root requiring a stdio declaration: the reinstall
// moves, the row arrives while lock.json still holds the old lock
// bytes, and Info carries no rows to reprint.
func TestReinstallEmitsSurfacingBeforePublication(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	ids := newGitIdentities(t)
	toolOperand := serveOrderTool(t, ids)
	source := t.TempDir()
	writeManifestPackage(t, source,
		`{"schema_version": 1, "name": "localmcp", "version": "1.0.0",`+
			`"requires": {"mcp": {"tool": {"git": "`+toolOperand+`", "range": "*"}}},`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "hello\n"})
	first := &lockObservingSink{lock: lockPath(home, "localmcp")}
	firstInfo, _, _, err := Install(home, InstallOptions{Operand: source, SurfacingSink: first})
	if err != nil {
		t.Fatal(err)
	}
	requireSingleEmission(t, first, firstInfo, []string{orderToolRow})
	if first.existed {
		t.Fatalf("lock.json already exists when the first mcp-declaration row is written: %s", first.lock)
	}
	oldBytes, err := os.ReadFile(lockPath(home, "localmcp")) // #nosec G304 -- test lock path
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "context", "a.md"), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sink := &lockObservingSink{lock: lockPath(home, "localmcp")}
	info, activated, updated, err := Install(home, InstallOptions{Operand: source, SurfacingSink: sink})
	if err != nil {
		t.Fatal(err)
	}
	if !updated || activated {
		t.Fatalf("reinstall activated=%v updated=%v, want an update without activation", activated, updated)
	}
	requireSingleEmission(t, sink, info, []string{orderToolRow})
	if !sink.existed {
		t.Fatal("lock.json is absent when the reinstall emits: the old lock must still stand")
	}
	if !bytes.Equal(sink.bytes, oldBytes) {
		t.Fatalf("lock.json at first row = %q, want the old lock bytes", sink.bytes)
	}
}

// TestSurfacingEmittedDespitePublicationFailure obstructs the candidate
// profile directory with a regular file, so the journal's directory
// creation fails and publication fails on every platform. The install
// fails — but the §2.3 rows were already printed: a failure after the
// emission point must not discard them.
func TestSurfacingEmittedDespitePublicationFailure(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	ids := newGitIdentities(t)
	seed := t.TempDir()
	writePackage(t, seed, "acme", "1.0.0", "hello\n")
	if _, _, _, err := Install(home, InstallOptions{Operand: seed}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ProfileDir(home, "withmcp"), []byte("obstruction"), 0o644); err != nil {
		t.Fatal(err)
	}
	operand := serveOrderRoot(t, ids, "withmcp", serveOrderTool(t, ids))
	sink := &lockObservingSink{lock: lockPath(home, "withmcp")}
	_, _, _, err := Install(home, InstallOptions{Operand: operand, SurfacingSink: sink})
	if err == nil {
		t.Fatal("an install onto an obstructed profile directory must fail")
	}
	if !sink.observed {
		t.Fatalf("no §2.3 row reached the sink before the failure: %v", err)
	}
	if !equalOrdered(sink.rows, []string{orderToolRow}) {
		t.Fatalf("emitted rows = %v, want %v", sink.rows, []string{orderToolRow})
	}
	if _, statErr := os.Stat(lockPath(home, "withmcp")); statErr == nil {
		t.Fatal("the failed install published a lock: nothing must be published")
	}
}

// failingSink refuses every write. Emission to it proves surfacing never
// fails the operation.
type failingSink struct{}

func (failingSink) Write([]byte) (int, error) { return 0, errors.New("sink refused") }

// TestSurfacingSinkWriteFailureIsNonFatal drives Install with a sink that
// refuses every write: the install still succeeds, because §2.3
// surfacing is informative and never fails the operation.
func TestSurfacingSinkWriteFailureIsNonFatal(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	ids := newGitIdentities(t)
	operand := serveOrderRoot(t, ids, "withmcp", serveOrderTool(t, ids))
	info, _, _, err := Install(home, InstallOptions{Operand: operand, SurfacingSink: failingSink{}})
	if err != nil {
		t.Fatalf("a refusing sink must not fail the install: %v", err)
	}
	if len(info.Surfacing) != 0 {
		t.Fatalf("Info.Surfacing = %v with a sink set: the rows would print twice", info.Surfacing)
	}
}
