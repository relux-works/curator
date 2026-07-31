package atomicity

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/scopes"
	"github.com/relux-works/curator/internal/transaction"
)

// env is one manager home, skill source root, and checkout. It mirrors the
// fixture internal/install uses; this suite keeps its own copy because it drives
// the installer from outside the package.
type env struct {
	t          *testing.T
	skillsRoot string
	home       string
	project    string
	// userHome is the machine-wide scope's mirror namespace.
	userHome string
	cfg      *config.Config
}

// newEnv builds a scope in the production-default adapter mode, so the mirror
// entries every case below commits and restores are the symbolic links a real
// machine gets rather than copies.
func newEnv(t *testing.T) *env {
	t.Helper()
	e := &env{t: t, skillsRoot: t.TempDir(), home: t.TempDir(), project: t.TempDir()}
	e.git(e.project, "init", "-q")
	e.cfg = &config.Config{
		Path:          filepath.Join(e.home, "config.json"),
		SkillsRoot:    e.skillsRoot,
		DefaultAgents: []string{"claude_code"},
		AdapterMode:   "auto",
	}
	return e
}

func (e *env) git(dir string, args ...string) {
	e.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		e.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func (e *env) write(root, rel, content string) {
	e.t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		e.t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		e.t.Fatal(err)
	}
}

// skill creates a tagged skill repository with one exported script command.
func (e *env) skill(name string) {
	e.t.Helper()
	dir := filepath.Join(e.skillsRoot, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		e.t.Fatal(err)
	}
	e.git(dir, "init", "-q", "-b", "main")
	e.write(dir, "SKILL.md", "---\nname: "+name+"\ndescription: d\n---\n# "+name+"\n")
	e.write(dir, "references/info.md", "ref")
	e.write(dir, "scripts/"+name+"-tool", "#!/bin/sh\necho "+name+"\n")
	spec := map[string]any{
		"schema_version": 4,
		"capabilities":   map[string]any{},
		"runtime_roots":  []string{"scripts"},
		"commands": map[string]any{
			name + "-tool": map[string]any{
				"type": "script", "unix_path": "scripts/" + name + "-tool", "win_path": "scripts/" + name + "-tool",
			},
		},
	}
	payload, _ := json.MarshalIndent(spec, "", "  ")
	e.write(dir, "csk-skill.json", string(payload))
	e.git(dir, "add", ".")
	e.git(dir, "commit", "-qm", "init")
	e.git(dir, "tag", "v1")
}

func (e *env) declare(names ...string) {
	e.t.Helper()
	e.declareWithAgents([]string{"claude_code"}, names...)
}

// declareWithAgents writes a project Skillfile for one explicit agent set. The
// selected agents reach both the install marker and the adapter roots, so a
// change to them is the cheapest way to force a genuine context replacement.
func (e *env) declareWithAgents(agents []string, names ...string) {
	e.t.Helper()
	skills := []map[string]any{}
	for _, name := range names {
		skills = append(skills, map[string]any{"name": name, "tag": "v1"})
	}
	payload, _ := json.MarshalIndent(map[string]any{
		"schema_version": 1,
		"agents":         agents,
		"skills":         skills,
	}, "", "  ")
	e.write(e.project, "Skillfile.json", string(payload))
	ignored := ".agents/\nSkillfile.dev.json\n"
	for _, agent := range agents {
		ignored += adapters.AgentPaths[agent] + "/\n"
	}
	e.write(e.project, ".gitignore", ignored)
}

// hybridDeclare declares skills in the machine-level hybrid scope, targeting
// this project by alias, so their context materializes in the hybrid store.
func (e *env) hybridDeclare(names ...string) {
	e.t.Helper()
	e.hybridDeclareTargeting([]string{"test"}, names...)
}

// hybridDeclareTargeting writes the machine-level hybrid manifest with an
// explicit activation target list. Passing a target this project does not match
// retargets the declarations away; passing no names empties the manifest.
func (e *env) hybridDeclareTargeting(targets []string, names ...string) {
	e.t.Helper()
	skills := []map[string]any{}
	for _, name := range names {
		skills = append(skills, map[string]any{"name": name, "tag": "v1", "targets": targets})
	}
	payload, _ := json.MarshalIndent(map[string]any{
		"schema_version": 1,
		"skills":         skills,
	}, "", "  ")
	e.write(e.home, "hybrid/Skillfile.json", string(payload))
}

// globalDeclareAll declares several skills in the machine-wide scope.
func (e *env) globalDeclareAll(agents []string, names ...string) {
	e.t.Helper()
	if _, err := install.GlobalInit(e.home); err != nil {
		e.t.Fatal(err)
	}
	skills := []map[string]any{}
	for _, name := range names {
		skills = append(skills, map[string]any{"name": name, "tag": "v1"})
	}
	payload, _ := json.MarshalIndent(map[string]any{
		"schema_version": 1,
		"agents":         agents,
		"skills":         skills,
	}, "", "  ")
	e.write(install.GlobalRoot(e.home), "Skillfile.json", string(payload))
}

// installPlatform is the shim platform an installation on this host actually
// produces. This suite used to pin "unix" on every runner, which made a Windows
// run assert a shape the platform cannot even build: a unix script runtime is
// validated for a POSIX execute bit no file on Windows carries, so every
// baseline install failed before a single rollback was exercised.
func installPlatform() string { return runtimestore.Platform() }

// shimName is the launcher filename one command gets on this host: Windows
// launchers are `.cmd` files, so a bare command name names nothing there.
func shimName(command string) string {
	if installPlatform() == "windows" {
		return command + ".cmd"
	}
	return command
}

func (e *env) install(opts install.Options) install.Result {
	e.t.Helper()
	opts.Platform = installPlatform()
	return install.Project(e.cfg, e.project, "test", opts)
}

// installGlobal runs the machine-wide scope against a stable user home, so a
// baseline and the installation that follows it share one adapter and user-bin
// mirror namespace.
func (e *env) installGlobal(opts install.Options) install.Result {
	e.t.Helper()
	if e.userHome == "" {
		e.userHome = e.t.TempDir()
	}
	opts.Platform = installPlatform()
	return install.Global(e.cfg, e.userHome, opts)
}

// sharedState is a digest of every class of state one installation commits.
// A rollback is only correct if the complete map returns to its prior value,
// so the cases below compare whole snapshots rather than individual paths.
type sharedState map[string]string

func snapshotState(t *testing.T, e *env) sharedState {
	t.Helper()
	paths := map[string]string{
		"project/skills":   filepath.Join(e.project, ".agents", "skills"),
		"project/bin":      filepath.Join(e.project, ".agents", "bin"),
		"project/env.sh":   filepath.Join(e.project, ".agents", "env.sh"),
		"project/env.ps1":  filepath.Join(e.project, ".agents", "env.ps1"),
		"project/adapters": filepath.Join(e.project, filepath.FromSlash(adapters.AgentPaths["claude_code"])),
		"project/adapters-codex": filepath.Join(e.project,
			filepath.FromSlash(adapters.AgentPaths["codex_cli"])),
		"hybrid/store":   scopes.HybridSkillsRoot(e.home),
		"home/runtime":   filepath.Join(e.home, "runtime"),
		"home/consumers": filepath.Join(e.home, scopes.ConsumersName),
		"global/skills":  filepath.Join(e.home, "global", "skills"),
		"global/bin":     filepath.Join(e.home, "global", "bin"),
		"global/env.sh":  filepath.Join(e.home, "global", "env.sh"),
	}
	if e.userHome != "" {
		paths["user/adapters"] = filepath.Join(e.userHome, filepath.FromSlash(adapters.AgentPaths["claude_code"]))
		paths["user/adapters-codex"] = filepath.Join(e.userHome, filepath.FromSlash(adapters.AgentPaths["codex_cli"]))
		paths["user/bin"] = filepath.Join(e.userHome, ".local", "bin")
	}
	state := sharedState{}
	for key, path := range paths {
		state[key] = entryDigest(path)
	}
	return state
}

// entryDigest reads one snapshot path exactly as it is. A symbolic link is
// digested by its destination rather than dereferenced, so a mirror that was
// replaced, re-pointed, or removed shows up as a change instead of resolving to
// the same canonical bytes.
func entryDigest(path string) string {
	info, err := os.Lstat(path)
	switch {
	case os.IsNotExist(err):
		return transaction.DigestAbsent
	case err != nil:
		return "unreadable:" + err.Error()
	case info.Mode()&os.ModeSymlink != 0:
		digest, err := transaction.DigestTarget(transaction.KindEntry, path)
		if err != nil {
			return "unreadable:" + err.Error()
		}
		return digest
	case !info.IsDir():
		digest, err := transaction.DigestPath(path)
		if err != nil {
			return "unreadable:" + err.Error()
		}
		return digest
	}
	// A directory may legitimately hold mirror links, which DigestPath refuses
	// inside a tree, so the tree is summarized entry by entry.
	entries, err := os.ReadDir(path)
	if err != nil {
		return "unreadable:" + err.Error()
	}
	parts := make([]string, 0, len(entries))
	for _, entry := range entries {
		parts = append(parts, entry.Name()+"="+entryDigest(filepath.Join(path, entry.Name())))
	}
	sort.Strings(parts)
	return "dir[" + strings.Join(parts, ",") + "]"
}

func (state sharedState) diff(other sharedState) []string {
	var changed []string
	for key, digest := range state {
		if other[key] != digest {
			changed = append(changed, fmt.Sprintf("%s: %s -> %s", key, digest, other[key]))
		}
	}
	sort.Strings(changed)
	return changed
}

// commitProbe records the ordered target boundaries a commit crosses and can
// fail exactly one of them.
//
// The fault fires at PointAfterBackup, which the engine emits once for every
// target — including a removal, whose desired state is already reached once its
// preimage has been moved aside and which therefore never reaches the install
// boundary. Faulting there is what makes a sweep able to reach every class.
type commitProbe struct {
	mu        sync.Mutex
	committed []transaction.Event
	rolled    []transaction.Event
	// preimage records, per target index, whether the live path held anything
	// before the commit touched it. A target with nothing to restore is
	// correctly absent from the rollback sequence.
	preimage map[int]bool
	failed   *transaction.Event
	// failClass fails the first target of one class.
	failClass string
	failErr   error
}

func (probe *commitProbe) hooks() transaction.Hooks {
	return transaction.Hooks{
		Observe: func(event transaction.Event) {
			probe.mu.Lock()
			defer probe.mu.Unlock()
			switch event.Point {
			case transaction.PointBeforeBackup:
				if probe.preimage == nil {
					probe.preimage = map[int]bool{}
				}
				_, err := os.Lstat(event.LivePath)
				probe.preimage[event.TargetIndex] = err == nil
			case transaction.PointTargetCommitted:
				probe.committed = append(probe.committed, event)
			case transaction.PointTargetRolledBack:
				probe.rolled = append(probe.rolled, event)
			}
		},
		Fault: func(event transaction.Event) error {
			probe.mu.Lock()
			defer probe.mu.Unlock()
			if probe.failErr == nil || probe.failed != nil ||
				event.Point != transaction.PointAfterBackup || event.Class != probe.failClass {
				return nil
			}
			probe.failed = &event
			return probe.failErr
		},
	}
}

func (probe *commitProbe) committedClasses() []string {
	probe.mu.Lock()
	defer probe.mu.Unlock()
	classes := make([]string, 0, len(probe.committed))
	for _, event := range probe.committed {
		classes = append(classes, event.Class)
	}
	return classes
}

// assertReverseRollback proves the restore sequence is the failing target
// followed by every committed target in exact reverse commit order. A target
// that had no prior state is skipped: there is nothing of it to put back.
func (probe *commitProbe) assertReverseRollback(t *testing.T) {
	t.Helper()
	probe.mu.Lock()
	committed := append([]transaction.Event(nil), probe.committed...)
	rolled := append([]transaction.Event(nil), probe.rolled...)
	preimage := map[int]bool{}
	for index, existed := range probe.preimage {
		preimage[index] = existed
	}
	failed := probe.failed
	probe.mu.Unlock()

	if failed == nil {
		t.Fatal("the injected fault never fired")
	}
	var want []transaction.Event
	if preimage[failed.TargetIndex] {
		want = append(want, *failed)
	}
	for index := len(committed) - 1; index >= 0; index-- {
		want = append(want, committed[index])
	}
	if len(rolled) != len(want) {
		t.Fatalf("rolled back %d targets, want %d (the failing target plus %d committed)",
			len(rolled), len(want), len(committed))
	}
	for offset, event := range rolled {
		if event.Class != want[offset].Class || event.Identifier != want[offset].Identifier {
			t.Fatalf("rollback step %d restored %s/%s, want %s/%s (exact reverse order)",
				offset, event.Class, event.Identifier, want[offset].Class, want[offset].Identifier)
		}
	}
}

// assertNoJournalRemains proves a failed operation left no durable transaction
// behind. A prepared or committing journal is one recovery would finish, so an
// installation that reported failure must not leave one for the next run.
func assertNoJournalRemains(t *testing.T, home string) {
	t.Helper()
	root := filepath.Join(home, "state", "transactions", "v1")
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	var remaining []string
	for _, entry := range entries {
		remaining = append(remaining, entry.Name())
	}
	if len(remaining) > 0 {
		t.Fatalf("a failed installation left transaction journals behind: %v", remaining)
	}
}

// adapterEntryState renders one adapter mirror entry exactly as it is on disk:
// its link destination, or a marker that it is a copied tree or absent. It is
// deliberately Lstat/Readlink based, because a link and the tree it points at
// are indistinguishable through any dereferencing check.
func adapterEntryState(t *testing.T, path string) string {
	t.Helper()
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return "absent"
	}
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		destination, err := os.Readlink(path)
		if err != nil {
			t.Fatal(err)
		}
		return "link:" + destination
	}
	if info.IsDir() {
		return "tree:" + entryDigest(path)
	}
	return "file:" + entryDigest(path)
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
