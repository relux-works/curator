# TASK-261016-i5gwaf — review rework reproducible evidence

Exact commands, exits, bounds and qualification status are in the paired handoff and canonical record. All paths below are repository-relative synthetic evidence.

## replay_test.go

```text
// Package muse148 records vendor-loader evidence; it never changes runtime admission.
package muse148

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const binarySHA = "5721763b4723f961709cd28f69aa049763965b46c4d4d70f883a0a90e5ddf39c"
const rootSHA = "8c9af1ca52b75aaa1e67af6365f1e39dfcc3c56a274d3571270079ef669ac9d0"

func read(t *testing.T, p string) []byte {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func write(t *testing.T, p string, b []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, b, 0600); err != nil {
		t.Fatal(err)
	}
}
func digest(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func quote(s string) string  { return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'" }

// checkCounts requires a real, successful loader record, including zero counts.
// Missing, failed, truncated and contradictory records cannot prove absence.
func checkCounts(log, event string, want map[string]int) error {
	seen := false
	for _, line := range strings.Split(log, "\n") {
		if !strings.Contains(line, `event="`+event+`"`) {
			continue
		}
		seen = true
		// Consume quoted values as whole fields, so tokens inside messages
		// cannot masquerade as loader fields. Relevant fields must occur once.
		fields := map[string][]string{}
		pattern := regexp.MustCompile(`(?:^|[ \t])([A-Za-z_][A-Za-z0-9_]*)=("(?:\\.|[^"\\])*"|[^ \t]+)`)
		for _, m := range pattern.FindAllStringSubmatchIndex(line, -1) {
			if m[1] < len(line) && line[m[1]] != ' ' && line[m[1]] != '\t' {
				return fmt.Errorf("%s malformed field", event)
			}
			key, value := line[m[2]:m[3]], line[m[4]:m[5]]
			fields[key] = append(fields[key], value)
		}
		for _, key := range append([]string{"event", "outcome"}, countKeys(want)...) {
			if len(fields[key]) != 1 {
				return fmt.Errorf("%s missing or duplicate %s", event, key)
			}
		}
		if fields["event"][0] != strconv.Quote(event) {
			return fmt.Errorf("%s contradictory event", event)
		}
		if fields["outcome"][0] != `"ready"` && fields["outcome"][0] != `"empty"` {
			return fmt.Errorf("%s failed loader record", event)
		}
		for key, n := range want {
			value := fields[key][0]
			got, err := strconv.Atoi(value)
			if err != nil || !regexp.MustCompile(`^[0-9]+$`).MatchString(value) {
				return fmt.Errorf("%s invalid %s", event, key)
			}
			if got != n {
				return fmt.Errorf("%s observed %s=%d; want %d", event, key, got, n)
			}
		}
	}
	if !seen {
		return fmt.Errorf("missing %s loader observation", event)
	}
	return nil
}

func countKeys(want map[string]int) []string {
	keys := make([]string, 0, len(want))
	for key := range want {
		keys = append(keys, key)
	}
	return keys
}

// These controls prove that explicit clean records pass while a foreign source,
// missing/truncated record or failed read cannot be attested as clean.
func TestObservationChecker(t *testing.T) {
	clean := `event="rules.context_load" outcome="ready" user_sources=1 foreign_sources=0`
	want := map[string]int{"user_sources": 1, "foreign_sources": 0}
	if err := checkCounts(clean, "rules.context_load", want); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"", "timeout", strings.Replace(clean, "foreign_sources=0", "foreign_sources=1", 1), strings.Replace(clean, "foreign_sources=0", "foreign_sources=", 1), strings.Replace(clean, `outcome="ready"`, `outcome="failed"`, 1), clean + "\n" + strings.Replace(clean, "foreign_sources=0", "foreign_sources=1", 1)} {
		if checkCounts(bad, "rules.context_load", want) == nil {
			t.Fatalf("checker admitted %q", bad)
		}
	}
	if err := checkCounts(`event="skills.catalog_load" outcome="ready" user_skills=0`, "skills.catalog_load", map[string]int{"user_skills": 0}); err != nil {
		t.Fatal(err)
	}
	if checkCounts(`event="skills.catalog_load" outcome="ready" user_skills=1`, "skills.catalog_load", map[string]int{"user_skills": 0}) == nil {
		t.Fatal("checker admitted one foreign skill")
	}
}

type probe struct{ root, binary, curator, sandbox string }

func setup(t *testing.T) *probe {
	t.Helper()
	binary := os.Getenv("MUSE148_BINARY")
	if binary == "" {
		t.Skip("vendor replay requires MUSE148_BINARY, MUSE148_CURATOR and MUSE148_RUN_ROOT; no qualification from a skip")
	}
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		t.Fatal("this evidence identity requires darwin/arm64")
	}
	base, err := filepath.Abs(os.Getenv("MUSE148_RUN_ROOT"))
	if err != nil || os.Getenv("MUSE148_RUN_ROOT") == "" {
		t.Fatal("explicit task-local run root required")
	}
	if !strings.Contains(base, string(filepath.Separator)+".temp"+string(filepath.Separator)) {
		t.Fatal("run root must be under .temp")
	}
	root := filepath.Join(base, strings.ReplaceAll(t.Name(), "/", "_"))
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	if digest(read(t, binary)) != binarySHA {
		t.Fatal("unqualified binary digest; this replay does not accept another build")
	}
	p := &probe{root: root, binary: binary, curator: os.Getenv("MUSE148_CURATOR"), sandbox: filepath.Join(root, "probe.sb")}
	if p.curator == "" {
		t.Fatal("MUSE148_CURATOR required")
	}
	// The real account's native stores are denied in addition to env -i and
	// disposable HOME/XDG. No credential is read to establish this boundary.
	native, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	sb := "(version 1)\n(allow default)\n(deny network*)\n(deny mach-lookup (global-name-regex #\"^com\\.apple\\.(security|Security)\"))\n(deny file-read* file-write* (regex #\"/Keychains(/|$)\"))\n"
	for _, rel := range []string{".claude", ".codex", ".config", ".local", ".muse", "Library"} {
		sb += "(deny file-read* file-write* (subpath " + strconv.Quote(filepath.Join(native, rel)) + "))\n"
	}
	write(t, p.sandbox, []byte(sb))
	return p
}

func (p *probe) env(dir string) []string {
	return []string{"HOME=" + filepath.Join(dir, "native"), "XDG_CONFIG_HOME=" + filepath.Join(dir, "config"), "XDG_DATA_HOME=" + filepath.Join(dir, "data"), "XDG_STATE_HOME=" + filepath.Join(dir, "state"), "XDG_CACHE_HOME=" + filepath.Join(dir, "cache"), "TMPDIR=" + filepath.Join(dir, "tmp"), "PATH=/usr/bin:/bin", "TERM=xterm-256color", "MUSE_NO_AUTO_UPDATE=1", "TBH_CREDENTIAL_BACKEND=file"}
}
func command(t *testing.T, dir string, env []string, name string, args ...string) ([]byte, int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = env
	b, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("bounded command timed out: %s", name)
	}
	if err == nil {
		return b, 0
	}
	if e, ok := err.(*exec.ExitError); ok {
		return b, e.ExitCode()
	}
	t.Fatal(err)
	return nil, -1
}
func (p *probe) fixture(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(p.root, "curator")
	env := p.env(dir)
	for _, sub := range []string{"native", "tmp", "config/muse", "work/.git"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0700); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range []string{"agent-context.json", "context/root.md"} {
		write(t, filepath.Join(dir, "source", f), read(t, filepath.Join("testdata/source", f)))
	}
	config := filepath.Join(dir, "manager/config.json")
	cfg, _ := json.Marshal(map[string]any{"schema_version": 1, "skills_root": filepath.Join(dir, "skills"), "projects": map[string]any{}})
	write(t, config, cfg)
	write(t, filepath.Join(dir, "system.json"), []byte(`{"schema_version":1}`))
	env = append(env, "CURATOR_CONFIG="+config, "CURATOR_SYSTEM_CONFIG="+filepath.Join(dir, "system.json"))
	for i, args := range [][]string{{"profile", "install", filepath.Join(dir, "source")}, {"env", "resolve", "muse", "--profile", "muse148-probe", "--repair", "--format", "json"}} {
		// Synthetic target permits the historical v3 resolver's metadata checks.
		// Muse never executes in this credential-bearing fixture home.
		if i == 1 {
			write(t, filepath.Join(dir, "config/muse/auth.json"), []byte("{}\n"))
		}
		b, code := command(t, filepath.Join(dir, "work"), env, p.curator, args...)
		write(t, filepath.Join(dir, fmt.Sprintf("curator-%d.log", i)), b)
		if code != 0 {
			t.Fatalf("Curator fixture command exit=%d: %s", code, b)
		}
	}
	link := filepath.Join(dir, "manager/environments/muse148-probe/muse/config/muse/AGENTS.md")
	s, err := os.Lstat(link)
	if err != nil || s.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("Curator did not produce a real link: %v", err)
	}
	if got := read(t, link); digest(got) != rootSHA || !bytes.Equal(got, read(t, "testdata/root.md")) {
		t.Fatal("Curator rendered root differs from frozen fixture")
	}
	return link
}

type observation struct {
	Log    string
	Output string
	Exit   int
}

func (p *probe) run(t *testing.T, form, representation, foreign string, exclude, workspace bool, link string, extra ...string) observation {
	t.Helper()
	dir := filepath.Join(p.root, strings.ReplaceAll(t.Name(), "/", "_"))
	for _, sub := range []string{"native", "config/muse", "data", "state", "cache", "tmp", "work/.git"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0700); err != nil {
			t.Fatal(err)
		}
	}
	root := filepath.Join(dir, "config/muse/AGENTS.md")
	launchEnv := p.env(dir)
	if representation == "linked" {
		if err := os.Symlink(link, root); err != nil {
			t.Fatal(err)
		}
	}
	if representation == "snapshot" {
		if err := checkFrozenSnapshot("testdata/snapshot", false); err != nil {
			t.Fatal(err)
		}
		for _, f := range []string{"manifest.json", "launch.json", "home/config/muse/AGENTS.md"} {
			write(t, filepath.Join(dir, "snapshot", f), read(t, filepath.Join("testdata/snapshot", f)))
		}
		root = filepath.Join(dir, "snapshot/home/config/muse/AGENTS.md")
		for i, value := range launchEnv {
			if strings.HasPrefix(value, "XDG_") {
				parts := strings.SplitN(value, "=", 2)
				launchEnv[i] = parts[0] + "=" + filepath.Join(dir, "snapshot/home", filepath.Base(parts[1]))
				if err := os.MkdirAll(strings.SplitN(launchEnv[i], "=", 2)[1], 0700); err != nil {
					t.Fatal(err)
				}
			}
		}
		s, err := os.Lstat(root)
		if err != nil || !s.Mode().IsRegular() || s.Mode().Perm() != 0600 {
			t.Fatal("snapshot root is not an independent private regular file")
		}
	}
	switch foreign {
	case "claude-rules":
		write(t, filepath.Join(dir, "native/.claude/CLAUDE.md"), []byte("MUSE148_FOREIGN_CLAUDE_RULE\n"))
	case "codex-rules":
		write(t, filepath.Join(dir, "native/.codex/AGENTS.md"), []byte("MUSE148_FOREIGN_CODEX_RULE\n"))
	case "skills":
		write(t, filepath.Join(dir, "native/.claude/skills/muse148-foreign/SKILL.md"), []byte("---\nname: muse148-foreign\ndescription: Synthetic foreign personal skill for offline loader qualification.\n---\nMUSE148_FOREIGN_SKILL\n"))
	}
	if workspace {
		write(t, filepath.Join(dir, "work/AGENTS.md"), []byte("MUSE148_WORKSPACE_CONFLICT\n"))
	}
	before := map[string][]byte{}
	if err := filepath.WalkDir(filepath.Join(dir, "native"), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			before[path] = read(t, path)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	args := []string{}
	if form == "exec" {
		args = append(args, "exec", "--json")
	}
	args = append(args, "--provider", "echo", "--trust-workspace")
	if exclude && os.Getenv("MUSE148_OMIT_EXCLUSION") != form {
		if form == "exec" {
			args = append(args, "--no-foreign-personal-context")
		} else {
			args = append(args, "-c", "context.foreign_personal_skills=false")
		}
	}
	args = append(args, extra...)
	if form == "exec" {
		args = append(args, "MUSE148_OFFLINE_CONTROL")
	}
	wrapped := append([]string{"-f", p.sandbox, "/usr/bin/env", "-i"}, launchEnv...)
	wrapped = append(wrapped, p.binary)
	wrapped = append(wrapped, args...)
	var out []byte
	code := 0
	if form == "exec" || len(extra) > 0 {
		out, code = command(t, filepath.Join(dir, "work"), p.env(dir), "/usr/bin/sandbox-exec", wrapped...)
	} else {
		out, code = p.tui(t, dir, wrapped)
	}
	write(t, filepath.Join(dir, "output.log"), out)
	for path, b := range before {
		if !bytes.Equal(read(t, path), b) {
			t.Fatalf("fixture source changed: %s", filepath.Base(path))
		}
	}
	if representation != "absent" {
		if digest(read(t, root)) != rootSHA {
			t.Fatal("native root bytes changed during vendor run")
		}
	} else if _, err := os.Lstat(root); !os.IsNotExist(err) {
		t.Fatal("root-absent control changed")
	}
	for _, path := range []string{"config/muse/auth.json", "snapshot/home/config/muse/auth.json", "native/.config/muse/auth.json"} {
		if _, err := os.Lstat(filepath.Join(dir, path)); !os.IsNotExist(err) {
			t.Fatalf("credential store created or unreadable: %s: %v", path, err)
		}
	}
	var logs strings.Builder
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasPrefix(d.Name(), "cli-") && strings.HasSuffix(path, ".log") {
			logs.Write(read(t, path))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// Persist only synthetic, sanitized loader observations for public evidence.
	safe := strings.ReplaceAll(string(out), dir, "$CASE")
	write(t, filepath.Join(dir, "argv.json"), mustJSON(t, args))
	write(t, filepath.Join(dir, "exit.txt"), []byte(strconv.Itoa(code)+"\n"))
	return observation{Log: logs.String(), Output: safe, Exit: code}
}

// A reachable local endpoint becomes unreachable only inside the probe policy;
// a synthetic Keychains file is denied while an ordinary file stays readable.
// No active credential store is accessed by this control.
func TestMuse148OfflineBoundary(t *testing.T) {
	p := setup(t)
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		_, _ = w.Write([]byte("MUSE148_NETWORK_CONTROL"))
	}))
	defer server.Close()
	args := []string{"--silent", "--show-error", "--noproxy", "*", "--connect-timeout", "2", server.URL}
	if b, c := command(t, p.root, os.Environ(), "/usr/bin/curl", args...); c != 0 || string(b) != "MUSE148_NETWORK_CONTROL" {
		t.Fatalf("network positive exit=%d: %s", c, b)
	}
	blocked := append([]string{"-f", p.sandbox, "/usr/bin/curl"}, args...)
	if b, c := command(t, p.root, os.Environ(), "/usr/bin/sandbox-exec", blocked...); c != 7 {
		t.Fatalf("network deny exit=%d: %s", c, b)
	}
	if calls.Load() != 1 {
		t.Fatal("denied network request reached the server")
	}
	for _, rel := range []string{"ordinary/control", "synthetic/Library/Keychains/control"} {
		write(t, filepath.Join(p.root, rel), []byte("MUSE148_SYNTHETIC_STORE_CONTROL"))
	}
	for _, tc := range []struct {
		path string
		code int
	}{{"ordinary/control", 0}, {"synthetic/Library/Keychains/control", 1}} {
		b, c := command(t, p.root, os.Environ(), "/usr/bin/sandbox-exec", "-f", p.sandbox, "/bin/cat", filepath.Join(p.root, tc.path))
		if c != tc.code {
			t.Fatalf("store boundary exit=%d want=%d: %s", c, tc.code, b)
		}
		if c != 0 && !strings.Contains(string(b), "Operation not permitted") {
			t.Fatalf("unrelated store refusal: %s", b)
		}
		if string(read(t, filepath.Join(p.root, tc.path))) != "MUSE148_SYNTHETIC_STORE_CONTROL" {
			t.Fatal("control bytes changed")
		}
	}
}

// checkFrozenSnapshot checks the actual home inventory without following links.
// Git does not preserve 0600, so source fixtures waive mode only; runtime copies do not.
func checkFrozenSnapshot(root string, private bool) error {
	for _, name := range []string{"manifest.json", "launch.json", "home"} {
		info, err := os.Lstat(filepath.Join(root, name))
		if err != nil {
			return err
		}
		if (name == "home" && !info.IsDir()) || (name != "home" && !info.Mode().IsRegular()) {
			return fmt.Errorf("fixture metadata/home type")
		}
	}
	manifestBytes, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		return err
	}
	launchBytes, err := os.ReadFile(filepath.Join(root, "launch.json"))
	if err != nil {
		return err
	}
	var manifest struct {
		Files []struct {
			Path string
			Size int
			SHA  string `json:"sha256"`
			Mode string
		}
		Launch string `json:"launch_sha256"`
	}
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return err
	}
	if len(manifest.Files) != 1 || manifest.Files[0].Path != "home/config/muse/AGENTS.md" || manifest.Files[0].Mode != "0600" {
		return fmt.Errorf("unexpected frozen closure")
	}
	if digest(launchBytes) != manifest.Launch {
		return fmt.Errorf("descriptor identity differs")
	}
	entry := manifest.Files[0]
	count := 0
	err = filepath.WalkDir(filepath.Join(root, "home"), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		count++
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if filepath.ToSlash(rel) != entry.Path || !info.Mode().IsRegular() || (private && info.Mode().Perm() != 0600) {
			return fmt.Errorf("fixture entry type/mode/set")
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if len(b) != entry.Size || digest(b) != entry.SHA || digest(b) != rootSHA {
			return fmt.Errorf("fixture byte identity")
		}
		return nil
	})
	if err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("fixture entry set")
	}
	return nil
}

// Frozen metadata binds the actual root-only fixture, then private copies pair
// valid regular files with altered bytes, extra files and both review link attacks.
func TestFrozenSnapshot(t *testing.T) {
	if err := checkFrozenSnapshot("testdata/snapshot", false); err != nil {
		t.Fatal(err)
	}
	for _, shape := range []string{"regular", "altered", "extra", "listed-symlink", "unlisted-dangling-link", "missing"} {
		t.Run(shape, func(t *testing.T) {
			root := snapshotCase(t, shape)
			err := checkFrozenSnapshot(root, true)
			if shape == "regular" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil {
				t.Fatalf("%s snapshot admitted", shape)
			}
			refusal := "fixture entry type/mode/set"
			switch shape {
			case "altered":
				refusal = "fixture byte identity"
			case "missing":
				refusal = "fixture entry set"
			}
			if err.Error() != refusal {
				t.Fatalf("%s unrelated refusal: %v", shape, err)
			}
			assertSnapshotPlantUnchanged(t, root, shape)
		})
	}
}

// Drive TestFrozenSnapshot itself against physical source fixtures: rebuilding
// pristine root bytes instead of checking its source must fail these controls.
func TestFrozenSnapshotSourceInputs(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, shape := range []string{"regular", "listed-symlink", "unlisted-dangling-link"} {
		t.Run(shape, func(t *testing.T) {
			bundle := snapshotCase(t, shape)
			dir := t.TempDir()
			if err := os.MkdirAll(filepath.Join(dir, "testdata"), 0700); err != nil {
				t.Fatal(err)
			}
			root := filepath.Join(dir, "testdata/snapshot")
			if err := os.Rename(bundle, root); err != nil {
				t.Fatal(err)
			}
			out, code := command(t, dir, os.Environ(), self, "-test.run=^TestFrozenSnapshot$/^regular$", "-test.v")
			if shape == "regular" {
				if code != 0 {
					t.Fatalf("regular fixture exit=%d: %s", code, out)
				}
			} else {
				if code != 1 || !bytes.Contains(out, []byte("fixture entry type/mode/set")) {
					t.Fatalf("%s source fixture exit=%d: %s", shape, code, out)
				}
				assertSnapshotPlantUnchanged(t, root, shape)
			}
		})
	}
}

func snapshotCase(t *testing.T, shape string) string {
	t.Helper()
	root := t.TempDir()
	for _, name := range []string{"manifest.json", "launch.json", "home/config/muse/AGENTS.md"} {
		write(t, filepath.Join(root, name), read(t, filepath.Join("testdata/snapshot", name)))
	}
	file := filepath.Join(root, "home/config/muse/AGENTS.md")
	switch shape {
	case "altered":
		write(t, file, []byte("MUSE148_FORGED_ROOT\n"))
	case "extra":
		write(t, filepath.Join(root, "home/extra"), []byte("extra"))
	case "listed-symlink":
		target := filepath.Join(t.TempDir(), "root.md")
		write(t, target, read(t, file))
		if err := os.Remove(file); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, file); err != nil {
			t.Fatal(err)
		}
	case "unlisted-dangling-link":
		if err := os.Symlink("missing-target", filepath.Join(root, "home/hidden-link")); err != nil {
			t.Fatal(err)
		}
	case "missing":
		if err := os.Remove(file); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func assertSnapshotPlantUnchanged(t *testing.T, root, shape string) {
	t.Helper()
	file := filepath.Join(root, "home/config/muse/AGENTS.md")
	switch shape {
	case "listed-symlink":
		info, err := os.Lstat(file)
		if err != nil || info.Mode()&os.ModeSymlink == 0 {
			t.Fatal("listed link changed")
		}
		if digest(read(t, file)) != rootSHA {
			t.Fatal("link target changed")
		}
	case "unlisted-dangling-link":
		target, err := os.Readlink(filepath.Join(root, "home/hidden-link"))
		if err != nil || target != "missing-target" {
			t.Fatal("dangling link changed")
		}
	case "altered":
		if string(read(t, file)) != "MUSE148_FORGED_ROOT\n" {
			t.Fatal("altered bytes changed")
		}
	case "extra":
		if string(read(t, filepath.Join(root, "home/extra"))) != "extra" {
			t.Fatal("extra bytes changed")
		}
	case "missing":
		if _, err := os.Lstat(file); !os.IsNotExist(err) {
			t.Fatal("missing file changed")
		}
	}
}

// The independent reference CLI accepts regular closure and refuses physical
// links, including dangling unlisted entries, without changing the rejected input.
func TestReferenceSnapshotPhysicalClosure(t *testing.T) {
	python, spec := os.Getenv("MUSE148_SPEC_PYTHON"), os.Getenv("MUSE148_SPEC")
	if python == "" || spec == "" {
		t.Skip("set MUSE148_SPEC_PYTHON and MUSE148_SPEC for the published reference CLI")
	}
	if err := checkFrozenSnapshot("testdata/snapshot", false); err != nil {
		t.Fatal(err)
	}
	for _, shape := range []string{"regular", "listed-symlink", "unlisted-dangling-link"} {
		t.Run(shape, func(t *testing.T) {
			root := snapshotCase(t, shape)
			out, code := command(t, ".", os.Environ(), python, "-B", "reference-snapshot-check.py", spec, root)
			if shape == "regular" {
				if code != 0 || !bytes.Contains(out, []byte("complete regular-file closure")) {
					t.Fatalf("regular reference exit=%d: %s", code, out)
				}
			} else {
				if code != 1 || !bytes.Contains(out, []byte("ValueError: physical-entry-type")) {
					t.Fatalf("%s reference exit=%d: %s", shape, code, out)
				}
				assertSnapshotPlantUnchanged(t, root, shape)
			}
		})
	}
}
func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	return append(b, '\n')
}

func (p *probe) tui(t *testing.T, dir string, args []string) ([]byte, int) {
	t.Helper()
	tmux, err := exec.LookPath("tmux")
	if err != nil {
		t.Fatal(err)
	}
	// Relative AF_UNIX address avoids the platform path-length limit, while
	// keeping the socket inside the task-local worktree.
	socket := "tui.sock"
	call := func(a ...string) []byte {
		t.Helper()
		b, c := command(t, dir, os.Environ(), tmux, append([]string{"-S", socket}, a...)...)
		if c != 0 {
			t.Fatalf("tmux exit=%d: %s", c, b)
		}
		return b
	}
	var quoted []string
	for _, arg := range append([]string{"/usr/bin/sandbox-exec"}, args...) {
		quoted = append(quoted, quote(arg))
	}
	// Publish the status atomically; an empty in-progress file is not an exit.
	cmd := "cd " + quote(filepath.Join(dir, "work")) + " && " + strings.Join(quoted, " ") + "; printf '%s\\n' \"$?\" > " + quote(filepath.Join(dir, "tui-exit.pending")) + " && mv " + quote(filepath.Join(dir, "tui-exit.pending")) + " " + quote(filepath.Join(dir, "tui-exit.txt"))
	call("new-session", "-d", "-x", "220", "-y", "70", "-s", "probe", cmd)
	pid := call("display-message", "-p", "#{pid} #{pane_pid}")
	write(t, filepath.Join(dir, "pids.txt"), pid)
	t.Cleanup(func() { command(t, dir, os.Environ(), tmux, "-S", socket, "kill-server") })
	deadline := time.Now().Add(15 * time.Second)
	var pane []byte
	for time.Now().Before(deadline) {
		pane = call("capture-pane", "-p", "-S", "-1000")
		if strings.Contains(string(pane), "Voice input") {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !strings.Contains(string(pane), "Voice input") {
		t.Fatalf("TUI not ready: %s", pane)
	}
	// Start a local echo turn so Muse binds and flushes its startup loader
	// diagnostics. The OS network deny remains active for the whole process.
	call("send-keys", "-l", "MUSE148_OFFLINE_CONTROL")
	time.Sleep(200 * time.Millisecond)
	call("send-keys", "Enter")
	deadline = time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		pane = call("capture-pane", "-p", "-S", "-1000")
		if strings.Contains(string(pane), "echo: MUSE148_OFFLINE_CONTROL") {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !strings.Contains(string(pane), "echo: MUSE148_OFFLINE_CONTROL") {
		t.Fatalf("missing offline echo control: %s", pane)
	}
	call("send-keys", "-l", "/rules")
	time.Sleep(200 * time.Millisecond)
	call("send-keys", "Enter")
	deadline = time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		pane = call("capture-pane", "-p", "-S", "-1000")
		if strings.Contains(string(pane), "Rules loaded this session") {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !strings.Contains(string(pane), "Rules loaded this session") {
		t.Fatalf("missing real /rules observation: %s", pane)
	}
	call("send-keys", "-l", "/exit")
	time.Sleep(200 * time.Millisecond)
	call("send-keys", "Enter")
	deadline = time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if b, e := os.ReadFile(filepath.Join(dir, "tui-exit.txt")); e == nil {
			n, e := strconv.Atoi(strings.TrimSpace(string(b)))
			if e != nil {
				t.Fatal(e)
			}
			return pane, n
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("TUI did not exit normally")
	return nil, -1
}

// The real parser rejects the exec-only flag and a JSON string separately;
// the matrix below supplies the accepted boolean and exec positive controls.
func TestMuse148ParserRefusals(t *testing.T) {
	p := setup(t)
	for name, arg := range map[string][]string{"interactive-exec-flag": {"--no-foreign-personal-context"}, "interactive-string-false": {"-c", `context.foreign_personal_skills="false"`}} {
		t.Run(name, func(t *testing.T) {
			o := p.run(t, "interactive", "absent", "none", false, false, "", arg...)
			refusal := "invalid TUI options"
			if name == "interactive-string-false" {
				refusal = "the JSON type does not fit this setting"
			}
			if o.Exit != 2 || !strings.Contains(o.Output, refusal) {
				t.Fatalf("expected parser refusal exit=2, got %d: %s", o.Exit, o.Output)
			}
			if strings.Contains(o.Log, `event="rules.context_load"`) {
				t.Fatal("rejected parser input reached loader")
			}
		})
	}
}

// The matrix pairs root and exclusion states for each isolated foreign source.
// Counts pin observed behavior, including the interactive exclusion FAILURE on
// this exact build; green replay is not qualification. See the qualification record.
func TestMuse148LoaderMatrix(t *testing.T) {
	p := setup(t)
	link := p.fixture(t)
	for _, form := range []string{"exec", "interactive"} {
		for _, root := range []string{"absent", "linked", "snapshot"} {
			for _, foreign := range []string{"claude-rules", "codex-rules", "skills"} {
				for _, exclude := range []bool{false, true} {
					name := fmt.Sprintf("%s/%s/%s/exclude-%t", form, root, foreign, exclude)
					t.Run(name, func(t *testing.T) {
						o := p.run(t, form, root, foreign, exclude, false, link)
						if o.Exit != 0 {
							t.Fatalf("vendor exit=%d: %s", o.Exit, o.Output)
						}
						users, fr, skills := 0, 0, 0
						if root != "absent" {
							users = 1
						}
						suppressed := exclude && (form == "exec" || root != "absent")
						if strings.HasSuffix(foreign, "rules") && root == "absent" && !suppressed {
							users = 1
							fr = 1
						}
						if foreign == "skills" && (!suppressed || form == "interactive") {
							skills = 1
						}
						if err := checkCounts(o.Log, "rules.context_load", map[string]int{"user_sources": users, "foreign_sources": fr, "project_sources": 0}); err != nil {
							t.Fatal(err)
						}
						if err := checkCounts(o.Log, "skills.catalog_load", map[string]int{"user_skills": skills, "project_skills": 0}); err != nil {
							t.Fatal(err)
						}
						t.Logf("vendor exit=0; user_sources=%d foreign_sources=%d user_skills=%d; no model claim", users, fr, skills)
					})
				}
			}
		}
	}
}

// Both native-root representations retain the foreign skill despite the valid
// boolean. A corrected loader must fail this exact-build regression witness,
// prompting a new record, never silently turning this evidence into admission.
func TestMuse148InteractiveExclusionRegression(t *testing.T) {
	p := setup(t)
	link := p.fixture(t)
	for _, root := range []string{"linked", "snapshot"} {
		t.Run(root, func(t *testing.T) {
			o := p.run(t, "interactive", root, "skills", true, false, link)
			const record = "docs/qualification/muse-1.4.8-R6256.1.md"
			if o.Exit != 0 {
				t.Fatalf("vendor exit=%d; re-evaluate %s: %s", o.Exit, record, o.Output)
			}
			if err := checkCounts(o.Log, "rules.context_load", map[string]int{"user_sources": 1, "foreign_sources": 0}); err != nil {
				t.Fatal(err)
			}
			if err := checkCounts(o.Log, "skills.catalog_load", map[string]int{"user_skills": 1, "project_skills": 0}); err != nil {
				t.Fatalf("interactive exclusion observation changed; re-evaluate %s: %v", record, err)
			}
			if !strings.Contains(o.Output, "Including your 1 Claude Code personal skill") {
				t.Fatalf("retained-skill banner changed; re-evaluate %s: %s", record, o.Output)
			}
			t.Logf("vendor exit=0; required interactive exclusion FAILED (user_skills=1); see %s", record)
		})
	}
}

// /rules and loader counts establish enumeration only. No model precedence
// or conflict resolution claim follows from both native and workspace loading.
func TestMuse148WorkspaceResidual(t *testing.T) {
	p := setup(t)
	link := p.fixture(t)
	for _, form := range []string{"exec", "interactive"} {
		for _, root := range []string{"linked", "snapshot"} {
			t.Run(form+"/"+root, func(t *testing.T) {
				o := p.run(t, form, root, "skills", true, true, link)
				if o.Exit != 0 {
					t.Fatalf("exit=%d", o.Exit)
				}
				if err := checkCounts(o.Log, "rules.context_load", map[string]int{"user_sources": 1, "foreign_sources": 0}); err != nil {
					t.Fatal(err)
				}
				// Interactive startup lists both sources, then a subsequent echo
				// reload reports project_sources=0. Preserve that residual rather
				// than asserting invariant workspace inclusion or precedence.
				observedProject := false
				for _, line := range strings.Split(o.Log, "\n") {
					if strings.Contains(line, `event="rules.context_load"`) && strings.Contains(line, " project_sources=1 ") {
						observedProject = true
					}
				}
				if !observedProject {
					t.Fatal("no observed native/workspace loader conflict")
				}
				if form == "interactive" && (!strings.Contains(o.Output, "project: AGENTS.md") || !strings.Contains(o.Output, "$CONFIG_DIR/AGENTS.md")) {
					t.Fatalf("missing /rules native/workspace order: %s", o.Output)
				}
			})
		}
	}
}

```

## review_attack_test.go

```text
package muse148

import "testing"

// Contradictory loader fields cannot attest a successful zero-skill result.
// Retained under the reviewer's original test name from review revision 1.
func TestReviewContradictoryObservation(t *testing.T) {
	want := map[string]int{"user_skills": 0}
	if err := checkCounts(`event="skills.catalog_load" outcome="ready" user_skills=0`, "skills.catalog_load", want); err != nil {
		t.Fatal(err)
	}
	for name, log := range map[string]string{
		"count-zero-then-one":       `event="skills.catalog_load" outcome="ready" user_skills=0 user_skills=1`,
		"count-one-then-zero":       `event="skills.catalog_load" outcome="ready" user_skills=1 user_skills=0`,
		"outcome-failed-then-ready": `event="skills.catalog_load" outcome="failed" outcome="ready" user_skills=0`,
		"outcome-ready-then-failed": `event="skills.catalog_load" outcome="ready" outcome="failed" user_skills=0`,
		"duplicate-count":           `event="skills.catalog_load" outcome="ready" user_skills=0 user_skills=0`,
		"duplicate-outcome":         `event="skills.catalog_load" outcome="ready" outcome="ready" user_skills=0`,
		"duplicate-event":           `event="skills.catalog_load" event="rules.context_load" outcome="ready" user_skills=0`,
		"quoted-decoy":              `message="event=\"skills.catalog_load\" outcome=\"ready\" user_skills=0"`,
		"malformed-count":           `event="skills.catalog_load" outcome="ready" user_skills=0garbage`,
		"overflow-count":            `event="skills.catalog_load" outcome="ready" user_skills=999999999999999999999999999`,
	} {
		t.Run(name, func(t *testing.T) {
			if err := checkCounts(log, "skills.catalog_load", want); err == nil {
				t.Fatalf("contradictory record admitted: %s", log)
			}
		})
	}
}

```

## reference-snapshot-check.py

```text
import os, stat, sys
from pathlib import Path
spec=Path(sys.argv[1]).resolve()
sys.path.insert(0,str(spec/'tools'))
from test_validate import validate
import managed_home_snapshots as snapshots
root=Path(sys.argv[2]).absolute()
# check_bundle accepts an in-memory inventory, not a filesystem safety claim.
# Classify every physical entry without following links before collecting bytes;
# dangling and unlisted links must not disappear from that inventory.
def inventory(directory):
    if not stat.S_ISDIR(directory.lstat().st_mode):
        raise ValueError('physical-entry-type')
    content = {}
    with os.scandir(directory) as entries:
        for entry in entries:
            path = Path(entry.path)
            mode = entry.stat(follow_symlinks=False).st_mode
            if stat.S_ISDIR(mode):
                content.update(inventory(path))
            elif stat.S_ISREG(mode):
                content[path.relative_to(root).as_posix()] = (path.read_bytes(), f'{mode & 0o777:04o}')
            else:
                raise ValueError('physical-entry-type')
    return content

if not stat.S_ISDIR(root.lstat().st_mode):
    raise ValueError('physical-entry-type')
for name in ('manifest.json', 'launch.json'):
    if not stat.S_ISREG((root/name).lstat().st_mode):
        raise ValueError('physical-entry-type')
content=inventory(root/'home')
manifest=validate.load_json(root/'manifest.json')
launch=validate.load_json(root/'launch.json')
if os.environ.get('MUSE148_SNAPSHOT_PLANT')=='extra': content['home/extra']=(b'extra','0600')
snapshots.check_bundle(manifest,launch,content,snapshots.validators(spec,validate.load_json),validate.ccj1_bytes,validate.validate_wire_semantics,(root/'manifest.json').read_bytes(),(root/'launch.json').read_bytes())
print('Reference check_bundle accepted exact frozen metadata and complete regular-file closure')

```

## review-regression-before-rev2.log

```text
=== RUN   TestReviewContradictoryObservation
=== RUN   TestReviewContradictoryObservation/outcome-ready-then-failed
    review_attack_test.go:26: contradictory record admitted: event="skills.catalog_load" outcome="ready" outcome="failed" user_skills=0
=== RUN   TestReviewContradictoryObservation/duplicate-count
    review_attack_test.go:26: contradictory record admitted: event="skills.catalog_load" outcome="ready" user_skills=0 user_skills=0
=== RUN   TestReviewContradictoryObservation/duplicate-outcome
    review_attack_test.go:26: contradictory record admitted: event="skills.catalog_load" outcome="ready" outcome="ready" user_skills=0
=== RUN   TestReviewContradictoryObservation/duplicate-event
    review_attack_test.go:26: contradictory record admitted: event="skills.catalog_load" event="rules.context_load" outcome="ready" user_skills=0
=== RUN   TestReviewContradictoryObservation/overflow-count
=== RUN   TestReviewContradictoryObservation/count-zero-then-one
    review_attack_test.go:26: contradictory record admitted: event="skills.catalog_load" outcome="ready" user_skills=0 user_skills=1
=== RUN   TestReviewContradictoryObservation/count-one-then-zero
=== RUN   TestReviewContradictoryObservation/outcome-failed-then-ready
    review_attack_test.go:26: contradictory record admitted: event="skills.catalog_load" outcome="failed" outcome="ready" user_skills=0
=== RUN   TestReviewContradictoryObservation/quoted-decoy
=== RUN   TestReviewContradictoryObservation/malformed-count
--- FAIL: TestReviewContradictoryObservation (0.00s)
    --- FAIL: TestReviewContradictoryObservation/outcome-ready-then-failed (0.00s)
    --- FAIL: TestReviewContradictoryObservation/duplicate-count (0.00s)
    --- FAIL: TestReviewContradictoryObservation/duplicate-outcome (0.00s)
    --- FAIL: TestReviewContradictoryObservation/duplicate-event (0.00s)
    --- PASS: TestReviewContradictoryObservation/overflow-count (0.00s)
    --- FAIL: TestReviewContradictoryObservation/count-zero-then-one (0.00s)
    --- PASS: TestReviewContradictoryObservation/count-one-then-zero (0.00s)
    --- FAIL: TestReviewContradictoryObservation/outcome-failed-then-ready (0.00s)
    --- PASS: TestReviewContradictoryObservation/quoted-decoy (0.00s)
    --- PASS: TestReviewContradictoryObservation/malformed-count (0.00s)
FAIL
FAIL	github.com/relux-works/curator/.research/muse148	0.393s
FAIL

```

## reference-regression-before-rev2.log

```text
=== RUN   TestReferenceSnapshotPhysicalClosure
=== RUN   TestReferenceSnapshotPhysicalClosure/regular
=== RUN   TestReferenceSnapshotPhysicalClosure/listed-symlink
    replay_test.go:589: listed-symlink reference exit=0: Reference check_bundle accepted exact frozen metadata and complete regular-file closure
=== RUN   TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link
    replay_test.go:589: unlisted-dangling-link reference exit=0: Reference check_bundle accepted exact frozen metadata and complete regular-file closure
--- FAIL: TestReferenceSnapshotPhysicalClosure (1.78s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/regular (0.69s)
    --- FAIL: TestReferenceSnapshotPhysicalClosure/listed-symlink (0.54s)
    --- FAIL: TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link (0.54s)
FAIL
FAIL	github.com/relux-works/curator/.research/muse148	2.085s
FAIL

```

## replay-rev2-02.log

```text
=== RUN   TestObservationChecker
--- PASS: TestObservationChecker (0.00s)
=== RUN   TestMuse148OfflineBoundary
--- PASS: TestMuse148OfflineBoundary (0.22s)
=== RUN   TestFrozenSnapshot
=== RUN   TestFrozenSnapshot/regular
=== RUN   TestFrozenSnapshot/altered
=== RUN   TestFrozenSnapshot/extra
=== RUN   TestFrozenSnapshot/listed-symlink
=== RUN   TestFrozenSnapshot/unlisted-dangling-link
=== RUN   TestFrozenSnapshot/missing
--- PASS: TestFrozenSnapshot (0.05s)
    --- PASS: TestFrozenSnapshot/regular (0.01s)
    --- PASS: TestFrozenSnapshot/altered (0.01s)
    --- PASS: TestFrozenSnapshot/extra (0.01s)
    --- PASS: TestFrozenSnapshot/listed-symlink (0.01s)
    --- PASS: TestFrozenSnapshot/unlisted-dangling-link (0.01s)
    --- PASS: TestFrozenSnapshot/missing (0.01s)
=== RUN   TestFrozenSnapshotSourceInputs
=== RUN   TestFrozenSnapshotSourceInputs/regular
=== RUN   TestFrozenSnapshotSourceInputs/listed-symlink
=== RUN   TestFrozenSnapshotSourceInputs/unlisted-dangling-link
--- PASS: TestFrozenSnapshotSourceInputs (0.03s)
    --- PASS: TestFrozenSnapshotSourceInputs/regular (0.02s)
    --- PASS: TestFrozenSnapshotSourceInputs/listed-symlink (0.01s)
    --- PASS: TestFrozenSnapshotSourceInputs/unlisted-dangling-link (0.01s)
=== RUN   TestReferenceSnapshotPhysicalClosure
=== RUN   TestReferenceSnapshotPhysicalClosure/regular
=== RUN   TestReferenceSnapshotPhysicalClosure/listed-symlink
=== RUN   TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link
--- PASS: TestReferenceSnapshotPhysicalClosure (1.03s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/regular (0.55s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/listed-symlink (0.23s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link (0.24s)
=== RUN   TestMuse148ParserRefusals
=== RUN   TestMuse148ParserRefusals/interactive-exec-flag
=== RUN   TestMuse148ParserRefusals/interactive-string-false
--- PASS: TestMuse148ParserRefusals (0.30s)
    --- PASS: TestMuse148ParserRefusals/interactive-exec-flag (0.06s)
    --- PASS: TestMuse148ParserRefusals/interactive-string-false (0.07s)
=== RUN   TestMuse148LoaderMatrix
=== RUN   TestMuse148LoaderMatrix/exec/absent/claude-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/claude-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/codex-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/codex-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/claude-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/claude-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/codex-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/codex-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/claude-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/claude-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/codex-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/codex-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/claude-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/claude-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/codex-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/codex-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/claude-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/claude-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/codex-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/codex-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/claude-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/claude-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/codex-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/codex-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
--- PASS: TestMuse148LoaderMatrix (60.50s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/claude-rules/exclude-false (1.16s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/claude-rules/exclude-true (1.17s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/codex-rules/exclude-false (1.34s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/codex-rules/exclude-true (1.18s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/skills/exclude-false (1.41s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/skills/exclude-true (1.44s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/claude-rules/exclude-false (1.41s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/claude-rules/exclude-true (1.28s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/codex-rules/exclude-false (1.11s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/codex-rules/exclude-true (1.25s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/skills/exclude-false (1.13s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/skills/exclude-true (1.24s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/claude-rules/exclude-false (1.22s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/claude-rules/exclude-true (1.41s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/codex-rules/exclude-false (1.14s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/codex-rules/exclude-true (0.96s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-false (1.10s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-true (0.96s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/claude-rules/exclude-false (1.97s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/claude-rules/exclude-true (2.20s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/codex-rules/exclude-false (2.30s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/codex-rules/exclude-true (1.87s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/skills/exclude-false (2.20s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/skills/exclude-true (2.65s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/claude-rules/exclude-false (2.74s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/claude-rules/exclude-true (2.18s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/codex-rules/exclude-false (2.07s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/codex-rules/exclude-true (1.86s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/skills/exclude-false (1.87s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/skills/exclude-true (1.86s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/claude-rules/exclude-false (2.08s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/claude-rules/exclude-true (2.08s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/codex-rules/exclude-false (2.08s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/codex-rules/exclude-true (1.98s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-false (1.87s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-true (1.75s)
=== RUN   TestMuse148InteractiveExclusionRegression
=== RUN   TestMuse148InteractiveExclusionRegression/linked
    replay_test.go:820: vendor exit=0; required interactive exclusion FAILED (user_skills=1); see docs/qualification/muse-1.4.8-R6256.1.md
=== RUN   TestMuse148InteractiveExclusionRegression/snapshot
    replay_test.go:820: vendor exit=0; required interactive exclusion FAILED (user_skills=1); see docs/qualification/muse-1.4.8-R6256.1.md
--- PASS: TestMuse148InteractiveExclusionRegression (4.34s)
    --- PASS: TestMuse148InteractiveExclusionRegression/linked (1.76s)
    --- PASS: TestMuse148InteractiveExclusionRegression/snapshot (1.86s)
=== RUN   TestMuse148WorkspaceResidual
=== RUN   TestMuse148WorkspaceResidual/exec/linked
=== RUN   TestMuse148WorkspaceResidual/exec/snapshot
=== RUN   TestMuse148WorkspaceResidual/interactive/linked
=== RUN   TestMuse148WorkspaceResidual/interactive/snapshot
--- PASS: TestMuse148WorkspaceResidual (6.19s)
    --- PASS: TestMuse148WorkspaceResidual/exec/linked (0.80s)
    --- PASS: TestMuse148WorkspaceResidual/exec/snapshot (0.83s)
    --- PASS: TestMuse148WorkspaceResidual/interactive/linked (1.98s)
    --- PASS: TestMuse148WorkspaceResidual/interactive/snapshot (1.87s)
=== RUN   TestReviewContradictoryObservation
=== RUN   TestReviewContradictoryObservation/quoted-decoy
=== RUN   TestReviewContradictoryObservation/malformed-count
=== RUN   TestReviewContradictoryObservation/count-zero-then-one
=== RUN   TestReviewContradictoryObservation/duplicate-count
=== RUN   TestReviewContradictoryObservation/duplicate-event
=== RUN   TestReviewContradictoryObservation/overflow-count
=== RUN   TestReviewContradictoryObservation/count-one-then-zero
=== RUN   TestReviewContradictoryObservation/outcome-failed-then-ready
=== RUN   TestReviewContradictoryObservation/outcome-ready-then-failed
=== RUN   TestReviewContradictoryObservation/duplicate-outcome
--- PASS: TestReviewContradictoryObservation (0.00s)
    --- PASS: TestReviewContradictoryObservation/quoted-decoy (0.00s)
    --- PASS: TestReviewContradictoryObservation/malformed-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/count-zero-then-one (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-event (0.00s)
    --- PASS: TestReviewContradictoryObservation/overflow-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/count-one-then-zero (0.00s)
    --- PASS: TestReviewContradictoryObservation/outcome-failed-then-ready (0.00s)
    --- PASS: TestReviewContradictoryObservation/outcome-ready-then-failed (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-outcome (0.00s)
PASS
ok  	github.com/relux-works/curator/.research/muse148	73.081s

```

## observations-review-rework.json

```text
{
  "version": "Muse Code 1.4.8 (1.4.8-R6256.1)",
  "binary_sha256": "5721763b4723f961709cd28f69aa049763965b46c4d4d70f883a0a90e5ddf39c",
  "binary_locator": "$MUSE148_BINARY (exact private locator retained in ignored task-local identity.json)",
  "platform": "macOS 27.0 arm64",
  "curator_source": "0d00a109ad1da77a5506747c89f24af9442eb341",
  "spec_source": "d2a942cf17e4eb017c45475442da1456d8dfbe6c",
  "replay_test_sha256": "78ecf817e292c2a7032bea08cf403dc9f2e86bd015eefdb3caf82abd02f94da1",
  "snapshot_manifest_sha256": "924378006ba6e6e7db9d0277375456a958c14cbaff03d1ebd63dfa6e4b7ceac8",
  "qualification": "exec loader rows pass; interactive skill exclusion FAILED; section 7.9 interactive release gate pending",
  "observations": [
    {
      "case": "TestMuse148InteractiveExclusionRegression_linked",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=false"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=1 user_skills=1 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "Including your 1 Claude Code personal skill \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148InteractiveExclusionRegression_snapshot",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=false"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=1 user_skills=1 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "Including your 1 Claude Code personal skill \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_absent_claude-rules_exclude-false",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=1 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=325 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=1 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=325 duration_ms=0"
      ],
      "tui_observations": [
        "muse: Including your Claude Code personal rules \u2014 manage with /settings.",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-5731-7163-8235-48bda2c079c6\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a142ed-585d-7db3-b22e-21701f565a0a\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a142ed-585d-7db3-b22e-21701f565a0a\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-585d-7db3-b22e-21701f565a0a\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-5731-7163-8235-48bda2c079c6\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a142ed-585d-7db3-b22e-21701f565a0a\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a142ed-585d-7db3-b22e-21701f565a0a\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-585d-7db3-b22e-21701f565a0a\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_absent_claude-rules_exclude-true",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "--no-foreign-personal-context",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"empty\" reason=\"none\" sources=0 user_sources=0 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=0 rendered_bytes=0 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-5bb3-7593-a0c8-cc39cfea0a69\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a142ed-5d0b-7f81-91be-c6541d3191e6\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a142ed-5d0b-7f81-91be-c6541d3191e6\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-5d0b-7f81-91be-c6541d3191e6\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-5bb3-7593-a0c8-cc39cfea0a69\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a142ed-5d0b-7f81-91be-c6541d3191e6\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a142ed-5d0b-7f81-91be-c6541d3191e6\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-5d0b-7f81-91be-c6541d3191e6\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_absent_codex-rules_exclude-false",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=1 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=317 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=1 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=317 duration_ms=0"
      ],
      "tui_observations": [
        "muse: Including your Codex personal rules \u2014 manage with /settings.",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-604f-7b22-becd-10595efb8c2d\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a142ed-61ee-7942-b35d-e58a4d9a0e6f\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a142ed-61ee-7942-b35d-e58a4d9a0e6f\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-61ee-7942-b35d-e58a4d9a0e6f\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-604f-7b22-becd-10595efb8c2d\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a142ed-61ee-7942-b35d-e58a4d9a0e6f\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a142ed-61ee-7942-b35d-e58a4d9a0e6f\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-61ee-7942-b35d-e58a4d9a0e6f\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_absent_codex-rules_exclude-true",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "--no-foreign-personal-context",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"empty\" reason=\"none\" sources=0 user_sources=0 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=0 rendered_bytes=0 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-6582-7a40-b5bd-3eaa74a48cb9\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a142ed-66e1-78a2-a8a5-8c5a2388ae18\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a142ed-66e1-78a2-a8a5-8c5a2388ae18\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-66e1-78a2-a8a5-8c5a2388ae18\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-6582-7a40-b5bd-3eaa74a48cb9\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a142ed-66e1-78a2-a8a5-8c5a2388ae18\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a142ed-66e1-78a2-a8a5-8c5a2388ae18\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-66e1-78a2-a8a5-8c5a2388ae18\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_absent_skills_exclude-false",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"empty\" reason=\"none\" sources=0 user_sources=0 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=0 rendered_bytes=0 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=1 user_skills=1 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"empty\" reason=\"none\" sources=0 user_sources=0 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=0 rendered_bytes=0 duration_ms=0"
      ],
      "tui_observations": [
        "muse: Including your 1 Claude Code personal skill \u2014 manage with /settings.",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-6a24-77d2-9bc4-c4c1c8518d2d\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a142ed-6bb0-7a43-aae0-da5dc394d91b\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a142ed-6bb0-7a43-aae0-da5dc394d91b\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-6bb0-7a43-aae0-da5dc394d91b\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-6a24-77d2-9bc4-c4c1c8518d2d\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a142ed-6bb0-7a43-aae0-da5dc394d91b\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a142ed-6bb0-7a43-aae0-da5dc394d91b\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-6bb0-7a43-aae0-da5dc394d91b\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_absent_skills_exclude-true",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "--no-foreign-personal-context",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"empty\" reason=\"none\" sources=0 user_sources=0 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=0 rendered_bytes=0 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-6f9a-7fd3-9d09-68333179e534\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a142ed-7159-7862-a9ba-82008b681676\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a142ed-7159-7862-a9ba-82008b681676\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-7159-7862-a9ba-82008b681676\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-6f9a-7fd3-9d09-68333179e534\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a142ed-7159-7862-a9ba-82008b681676\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a142ed-7159-7862-a9ba-82008b681676\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-7159-7862-a9ba-82008b681676\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_linked_claude-rules_exclude-false",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-7541-7df1-9280-699ded5ec4da\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a142ed-772c-7632-854e-aa27568ab25e\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a142ed-772c-7632-854e-aa27568ab25e\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-772c-7632-854e-aa27568ab25e\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-7541-7df1-9280-699ded5ec4da\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a142ed-772c-7632-854e-aa27568ab25e\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a142ed-772c-7632-854e-aa27568ab25e\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-772c-7632-854e-aa27568ab25e\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_linked_claude-rules_exclude-true",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "--no-foreign-personal-context",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-7ac2-7591-ba9b-55edb0d9ac0c\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a142ed-7c56-79a2-b461-e790cf6efba3\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a142ed-7c56-79a2-b461-e790cf6efba3\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-7c56-79a2-b461-e790cf6efba3\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-7ac2-7591-ba9b-55edb0d9ac0c\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a142ed-7c56-79a2-b461-e790cf6efba3\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a142ed-7c56-79a2-b461-e790cf6efba3\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-7c56-79a2-b461-e790cf6efba3\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_linked_codex-rules_exclude-false",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-7fb8-7373-99d1-c4c96354dc35\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a142ed-8100-76e1-8f49-46b47ed4b329\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a142ed-8100-76e1-8f49-46b47ed4b329\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-8100-76e1-8f49-46b47ed4b329\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-7fb8-7373-99d1-c4c96354dc35\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a142ed-8100-76e1-8f49-46b47ed4b329\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a142ed-8100-76e1-8f49-46b47ed4b329\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-8100-76e1-8f49-46b47ed4b329\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_linked_codex-rules_exclude-true",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "--no-foreign-personal-context",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-8411-7501-9dd8-c2d4170e747b\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a142ed-8591-77d3-969a-3be86da435f9\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a142ed-8591-77d3-969a-3be86da435f9\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-8591-77d3-969a-3be86da435f9\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-8411-7501-9dd8-c2d4170e747b\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a142ed-8591-77d3-969a-3be86da435f9\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a142ed-8591-77d3-969a-3be86da435f9\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-8591-77d3-969a-3be86da435f9\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_linked_skills_exclude-false",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=1 user_skills=1 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "muse: Including your 1 Claude Code personal skill \u2014 manage with /settings.",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-88f2-7191-b23b-17949371bc6d\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a142ed-8a2b-7da0-a99f-c0c5bc4d8112\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a142ed-8a2b-7da0-a99f-c0c5bc4d8112\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-8a2b-7da0-a99f-c0c5bc4d8112\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-88f2-7191-b23b-17949371bc6d\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a142ed-8a2b-7da0-a99f-c0c5bc4d8112\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a142ed-8a2b-7da0-a99f-c0c5bc4d8112\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-8a2b-7da0-a99f-c0c5bc4d8112\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_linked_skills_exclude-true",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "--no-foreign-personal-context",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-8d59-77f1-a193-c9de45bcb3e9\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a142ed-8f03-75d3-99eb-a98e872a2eba\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a142ed-8f03-75d3-99eb-a98e872a2eba\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-8f03-75d3-99eb-a98e872a2eba\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-8d59-77f1-a193-c9de45bcb3e9\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a142ed-8f03-75d3-99eb-a98e872a2eba\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a142ed-8f03-75d3-99eb-a98e872a2eba\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-8f03-75d3-99eb-a98e872a2eba\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_snapshot_claude-rules_exclude-false",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-923b-72d3-add5-9737f5bfce6a\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a142ed-939f-76e1-b531-4273bb1aef20\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a142ed-939f-76e1-b531-4273bb1aef20\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-939f-76e1-b531-4273bb1aef20\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-923b-72d3-add5-9737f5bfce6a\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a142ed-939f-76e1-b531-4273bb1aef20\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a142ed-939f-76e1-b531-4273bb1aef20\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-939f-76e1-b531-4273bb1aef20\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_snapshot_claude-rules_exclude-true",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "--no-foreign-personal-context",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-96ee-7571-8310-04d5241c8cb0\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a142ed-98e6-7e30-8e38-5492a7521fa2\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a142ed-98e6-7e30-8e38-5492a7521fa2\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-98e6-7e30-8e38-5492a7521fa2\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-96ee-7571-8310-04d5241c8cb0\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a142ed-98e6-7e30-8e38-5492a7521fa2\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a142ed-98e6-7e30-8e38-5492a7521fa2\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-98e6-7e30-8e38-5492a7521fa2\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_snapshot_codex-rules_exclude-false",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-9c7c-76f0-9ef1-39098d096fbf\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a142ed-9dc7-70a1-ad12-cab983392198\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a142ed-9dc7-70a1-ad12-cab983392198\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-9dc7-70a1-ad12-cab983392198\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-9c7c-76f0-9ef1-39098d096fbf\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a142ed-9dc7-70a1-ad12-cab983392198\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a142ed-9dc7-70a1-ad12-cab983392198\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-9dc7-70a1-ad12-cab983392198\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_snapshot_codex-rules_exclude-true",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "--no-foreign-personal-context",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-a0df-7660-a099-d93a9236ef4b\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a142ed-a1e2-7183-bbd8-962fb1a45d67\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a142ed-a1e2-7183-bbd8-962fb1a45d67\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-a1e2-7183-bbd8-962fb1a45d67\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-a0df-7660-a099-d93a9236ef4b\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a142ed-a1e2-7183-bbd8-962fb1a45d67\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a142ed-a1e2-7183-bbd8-962fb1a45d67\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-a1e2-7183-bbd8-962fb1a45d67\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_snapshot_skills_exclude-false",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=1 user_skills=1 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "muse: Including your 1 Claude Code personal skill \u2014 manage with /settings.",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-a4a2-7473-afa8-4210f5bd1a9f\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a142ed-a5bd-7111-ae99-1ca78273f9dc\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a142ed-a5bd-7111-ae99-1ca78273f9dc\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-a5bd-7111-ae99-1ca78273f9dc\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-a4a2-7473-afa8-4210f5bd1a9f\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a142ed-a5bd-7111-ae99-1ca78273f9dc\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a142ed-a5bd-7111-ae99-1ca78273f9dc\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-a5bd-7111-ae99-1ca78273f9dc\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_snapshot_skills_exclude-true",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "--no-foreign-personal-context",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-a8f3-7f40-a6b6-c541e751b66a\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a142ed-a9fe-73b3-8504-d8f2f8d61387\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a142ed-a9fe-73b3-8504-d8f2f8d61387\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-a9fe-73b3-8504-d8f2f8d61387\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ed-a8f3-7f40-a6b6-c541e751b66a\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a142ed-a9fe-73b3-8504-d8f2f8d61387\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a142ed-a9fe-73b3-8504-d8f2f8d61387\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ed-a9fe-73b3-8504-d8f2f8d61387\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_absent_claude-rules_exclude-false",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=1 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=325 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=1 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=325 duration_ms=0"
      ],
      "tui_observations": [
        "Including your Claude Code personal rules \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: ~/.claude/CLAUDE.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_absent_claude-rules_exclude-true",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=false"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=1 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=325 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=1 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=325 duration_ms=0"
      ],
      "tui_observations": [
        "Including your Claude Code personal rules \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: ~/.claude/CLAUDE.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_absent_codex-rules_exclude-false",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=1 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=317 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=1 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=317 duration_ms=0"
      ],
      "tui_observations": [
        "Including your Codex personal rules \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: ~/.codex/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_absent_codex-rules_exclude-true",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=false"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=1 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=317 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=1 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=317 duration_ms=0"
      ],
      "tui_observations": [
        "Including your Codex personal rules \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: ~/.codex/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_absent_skills_exclude-false",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"empty\" reason=\"none\" sources=0 user_sources=0 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=0 rendered_bytes=0 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=1 user_skills=1 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"empty\" reason=\"none\" sources=0 user_sources=0 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=0 rendered_bytes=0 duration_ms=0"
      ],
      "tui_observations": [
        "Including your 1 Claude Code personal skill \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: none",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_absent_skills_exclude-true",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=false"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"empty\" reason=\"none\" sources=0 user_sources=0 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=0 rendered_bytes=0 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=1 user_skills=1 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"empty\" reason=\"none\" sources=0 user_sources=0 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=0 rendered_bytes=0 duration_ms=0"
      ],
      "tui_observations": [
        "Including your 1 Claude Code personal skill \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: none",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_linked_claude-rules_exclude-false",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_linked_claude-rules_exclude-true",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=false"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_linked_codex-rules_exclude-false",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_linked_codex-rules_exclude-true",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=false"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_linked_skills_exclude-false",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=1 user_skills=1 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "Including your 1 Claude Code personal skill \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_linked_skills_exclude-true",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=false"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=1 user_skills=1 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "Including your 1 Claude Code personal skill \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_snapshot_claude-rules_exclude-false",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_snapshot_claude-rules_exclude-true",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=false"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_snapshot_codex-rules_exclude-false",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_snapshot_codex-rules_exclude-true",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=false"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_snapshot_skills_exclude-false",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=1 user_skills=1 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "Including your 1 Claude Code personal skill \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_snapshot_skills_exclude-true",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=false"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=1 user_skills=1 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "Including your 1 Claude Code personal skill \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148ParserRefusals_interactive-exec-flag",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "--no-foreign-personal-context"
      ],
      "vendor_exit": 2,
      "loader_records": [],
      "tui_observations": []
    },
    {
      "case": "TestMuse148ParserRefusals_interactive-string-false",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=\"false\""
      ],
      "vendor_exit": 2,
      "loader_records": [],
      "tui_observations": []
    },
    {
      "case": "TestMuse148WorkspaceResidual_exec_linked",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "--no-foreign-personal-context",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=2 user_sources=1 project_sources=1 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=2 rendered_bytes=1059 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ee-5347-7582-8b95-4a73bed31459\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a142ee-5427-7d60-9405-e222be8a73bb\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a142ee-5427-7d60-9405-e222be8a73bb\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ee-5427-7d60-9405-e222be8a73bb\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ee-5347-7582-8b95-4a73bed31459\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a142ee-5427-7d60-9405-e222be8a73bb\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a142ee-5427-7d60-9405-e222be8a73bb\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ee-5427-7d60-9405-e222be8a73bb\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148WorkspaceResidual_exec_snapshot",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "--no-foreign-personal-context",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=2 user_sources=1 project_sources=1 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=2 rendered_bytes=1059 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ee-566d-7d82-995b-b5394a39f8e4\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a142ee-5757-7341-baf7-3ed4224d0f9b\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a142ee-5757-7341-baf7-3ed4224d0f9b\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ee-5757-7341-baf7-3ed4224d0f9b\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a142ee-566d-7d82-995b-b5394a39f8e4\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a142ee-5757-7341-baf7-3ed4224d0f9b\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a142ee-5757-7341-baf7-3ed4224d0f9b\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a142ee-5757-7341-baf7-3ed4224d0f9b\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148WorkspaceResidual_interactive_linked",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=false"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=2 user_sources=1 project_sources=1 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=2 rendered_bytes=1059 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=1 user_skills=1 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "Including your 1 Claude Code personal skill \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: AGENTS.md"
      ]
    },
    {
      "case": "TestMuse148WorkspaceResidual_interactive_snapshot",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=false"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=2 user_sources=1 project_sources=1 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=2 rendered_bytes=1059 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=1 user_skills=1 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "Including your 1 Claude Code personal skill \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: AGENTS.md"
      ]
    }
  ],
  "review_attack_test_sha256": "bfedfd84093101834626e55476bb7361b80e8b9a955593da7f4cffc29f977103",
  "reference_checker_sha256": "524a8a00d2ce68be6084d3c87ac39a2f7a051cf3b064216fc5e7f3cca50dc1e6"
}

```

## plant-exec-rev2.log

```text
=== RUN   TestMuse148LoaderMatrix
=== RUN   TestMuse148LoaderMatrix/exec/linked/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/skills/exclude-true
    replay_test.go:788: skills.catalog_load observed user_skills=1; want 0
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-true
    replay_test.go:788: skills.catalog_load observed user_skills=1; want 0
--- FAIL: TestMuse148LoaderMatrix (7.16s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/skills/exclude-false (1.62s)
    --- FAIL: TestMuse148LoaderMatrix/exec/linked/skills/exclude-true (1.50s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-false (1.56s)
    --- FAIL: TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-true (1.69s)
FAIL
FAIL	github.com/relux-works/curator/.research/muse148	7.349s
FAIL

```

## plant-interactive-rev2.log

```text
=== RUN   TestMuse148LoaderMatrix
=== RUN   TestMuse148LoaderMatrix/interactive/linked/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
--- PASS: TestMuse148LoaderMatrix (10.76s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/skills/exclude-false (2.82s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/skills/exclude-true (2.39s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-false (2.18s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-true (2.07s)
PASS
ok  	github.com/relux-works/curator/.research/muse148	10.945s

```

## snapshot-reference-rev2.log

```text
Reference check_bundle accepted exact frozen metadata and complete regular-file closure

```

## snapshot-plant-rev2.log

```text
Traceback (most recent call last):
  File "$PRIVATE_PATH line 36, in <module>
    snapshots.check_bundle(manifest,launch,content,snapshots.validators(spec,validate.load_json),validate.ccj1_bytes,validate.validate_wire_semantics,(root/'manifest.json').read_bytes(),(root/'launch.json').read_bytes())
    ~~~~~~~~~~~~~~~~~~~~~~^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
  File "$PRIVATE_PATH line 81, in check_bundle
    raise ValueError("entry-set")
ValueError: entry-set

```

## checker-duplicate-count-rev2-final.log

```text
=== RUN   TestReviewContradictoryObservation
=== RUN   TestReviewContradictoryObservation/count-one-then-zero
=== RUN   TestReviewContradictoryObservation/outcome-failed-then-ready
=== RUN   TestReviewContradictoryObservation/outcome-ready-then-failed
=== RUN   TestReviewContradictoryObservation/duplicate-count
=== RUN   TestReviewContradictoryObservation/duplicate-outcome
=== RUN   TestReviewContradictoryObservation/duplicate-event
=== RUN   TestReviewContradictoryObservation/quoted-decoy
=== RUN   TestReviewContradictoryObservation/malformed-count
=== RUN   TestReviewContradictoryObservation/count-zero-then-one
    review_attack_test.go:26: contradictory record admitted: event="skills.catalog_load" outcome="ready" user_skills=0 user_skills=1
=== RUN   TestReviewContradictoryObservation/overflow-count
--- FAIL: TestReviewContradictoryObservation (0.00s)
    --- PASS: TestReviewContradictoryObservation/count-one-then-zero (0.00s)
    --- PASS: TestReviewContradictoryObservation/outcome-failed-then-ready (0.00s)
    --- PASS: TestReviewContradictoryObservation/outcome-ready-then-failed (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-outcome (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-event (0.00s)
    --- PASS: TestReviewContradictoryObservation/quoted-decoy (0.00s)
    --- PASS: TestReviewContradictoryObservation/malformed-count (0.00s)
    --- FAIL: TestReviewContradictoryObservation/count-zero-then-one (0.00s)
    --- PASS: TestReviewContradictoryObservation/overflow-count (0.00s)
FAIL
FAIL	github.com/relux-works/curator/.temp/TASK-261016-i5gwaf/mutants-rev2-final/checker-duplicate-count	0.290s
FAIL

```

## checker-duplicate-outcome-rev2-final.log

```text
=== RUN   TestReviewContradictoryObservation
=== RUN   TestReviewContradictoryObservation/outcome-failed-then-ready
    review_attack_test.go:26: contradictory record admitted: event="skills.catalog_load" outcome="failed" outcome="ready" user_skills=0
=== RUN   TestReviewContradictoryObservation/outcome-ready-then-failed
=== RUN   TestReviewContradictoryObservation/duplicate-outcome
=== RUN   TestReviewContradictoryObservation/duplicate-event
=== RUN   TestReviewContradictoryObservation/overflow-count
=== RUN   TestReviewContradictoryObservation/count-zero-then-one
=== RUN   TestReviewContradictoryObservation/count-one-then-zero
=== RUN   TestReviewContradictoryObservation/duplicate-count
=== RUN   TestReviewContradictoryObservation/quoted-decoy
=== RUN   TestReviewContradictoryObservation/malformed-count
--- FAIL: TestReviewContradictoryObservation (0.00s)
    --- FAIL: TestReviewContradictoryObservation/outcome-failed-then-ready (0.00s)
    --- PASS: TestReviewContradictoryObservation/outcome-ready-then-failed (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-outcome (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-event (0.00s)
    --- PASS: TestReviewContradictoryObservation/overflow-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/count-zero-then-one (0.00s)
    --- PASS: TestReviewContradictoryObservation/count-one-then-zero (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/quoted-decoy (0.00s)
    --- PASS: TestReviewContradictoryObservation/malformed-count (0.00s)
FAIL
FAIL	github.com/relux-works/curator/.temp/TASK-261016-i5gwaf/mutants-rev2-final/checker-duplicate-outcome	0.300s
FAIL

```

## snapshot-go-listed-link-rev2-final.log

```text
=== RUN   TestFrozenSnapshot
=== RUN   TestFrozenSnapshot/regular
=== RUN   TestFrozenSnapshot/altered
=== RUN   TestFrozenSnapshot/extra
=== RUN   TestFrozenSnapshot/listed-symlink
    replay_test.go:498: listed-symlink snapshot admitted
=== RUN   TestFrozenSnapshot/unlisted-dangling-link
=== RUN   TestFrozenSnapshot/missing
--- FAIL: TestFrozenSnapshot (0.35s)
    --- PASS: TestFrozenSnapshot/regular (0.06s)
    --- PASS: TestFrozenSnapshot/altered (0.03s)
    --- PASS: TestFrozenSnapshot/extra (0.03s)
    --- FAIL: TestFrozenSnapshot/listed-symlink (0.05s)
    --- PASS: TestFrozenSnapshot/unlisted-dangling-link (0.04s)
    --- PASS: TestFrozenSnapshot/missing (0.14s)
=== RUN   TestFrozenSnapshotSourceInputs
=== RUN   TestFrozenSnapshotSourceInputs/regular
=== RUN   TestFrozenSnapshotSourceInputs/listed-symlink
    replay_test.go:540: listed-symlink source fixture exit=0: === RUN   TestFrozenSnapshot
        === RUN   TestFrozenSnapshot/regular
        --- PASS: TestFrozenSnapshot (0.04s)
            --- PASS: TestFrozenSnapshot/regular (0.04s)
        PASS
=== RUN   TestFrozenSnapshotSourceInputs/unlisted-dangling-link
--- FAIL: TestFrozenSnapshotSourceInputs (0.36s)
    --- PASS: TestFrozenSnapshotSourceInputs/regular (0.17s)
    --- FAIL: TestFrozenSnapshotSourceInputs/listed-symlink (0.10s)
    --- PASS: TestFrozenSnapshotSourceInputs/unlisted-dangling-link (0.09s)
FAIL
FAIL	github.com/relux-works/curator/.temp/TASK-261016-i5gwaf/mutants-rev2-final/snapshot-go-listed-link	1.011s
FAIL

```

## snapshot-reference-listed-link-rev2-final.log

```text
=== RUN   TestReferenceSnapshotPhysicalClosure
=== RUN   TestReferenceSnapshotPhysicalClosure/regular
=== RUN   TestReferenceSnapshotPhysicalClosure/listed-symlink
    replay_test.go:632: listed-symlink reference exit=0: Reference check_bundle accepted exact frozen metadata and complete regular-file closure
=== RUN   TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link
--- FAIL: TestReferenceSnapshotPhysicalClosure (1.49s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/regular (0.59s)
    --- FAIL: TestReferenceSnapshotPhysicalClosure/listed-symlink (0.63s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link (0.27s)
FAIL
FAIL	github.com/relux-works/curator/.temp/TASK-261016-i5gwaf/mutants-rev2-final/snapshot-reference-listed-link	1.771s
FAIL

```

## snapshot-reference-dangling-link-rev2-final.log

```text
=== RUN   TestReferenceSnapshotPhysicalClosure
=== RUN   TestReferenceSnapshotPhysicalClosure/regular
=== RUN   TestReferenceSnapshotPhysicalClosure/listed-symlink
=== RUN   TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link
    replay_test.go:632: unlisted-dangling-link reference exit=0: Reference check_bundle accepted exact frozen metadata and complete regular-file closure
--- FAIL: TestReferenceSnapshotPhysicalClosure (1.50s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/regular (0.58s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/listed-symlink (0.31s)
    --- FAIL: TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link (0.61s)
FAIL
FAIL	github.com/relux-works/curator/.temp/TASK-261016-i5gwaf/mutants-rev2-final/snapshot-reference-dangling-link	1.769s
FAIL

```

## checker-one-skill-rev2-final.log

```text
=== RUN   TestObservationChecker
    replay_test.go:121: checker admitted one foreign skill
--- FAIL: TestObservationChecker (0.00s)
FAIL
FAIL	github.com/relux-works/curator/.temp/TASK-261016-i5gwaf/mutants-rev2-final/checker-one-skill	0.273s
FAIL

```

## snapshot-extra-entry-rev2-final.log

```text
=== RUN   TestFrozenSnapshot
=== RUN   TestFrozenSnapshot/regular
=== RUN   TestFrozenSnapshot/altered
=== RUN   TestFrozenSnapshot/extra
    replay_test.go:498: extra snapshot admitted
=== RUN   TestFrozenSnapshot/listed-symlink
=== RUN   TestFrozenSnapshot/unlisted-dangling-link
=== RUN   TestFrozenSnapshot/missing
--- FAIL: TestFrozenSnapshot (0.21s)
    --- PASS: TestFrozenSnapshot/regular (0.09s)
    --- PASS: TestFrozenSnapshot/altered (0.02s)
    --- FAIL: TestFrozenSnapshot/extra (0.03s)
    --- PASS: TestFrozenSnapshot/listed-symlink (0.04s)
    --- PASS: TestFrozenSnapshot/unlisted-dangling-link (0.02s)
    --- PASS: TestFrozenSnapshot/missing (0.01s)
FAIL
FAIL	github.com/relux-works/curator/.temp/TASK-261016-i5gwaf/mutants-rev2-final/snapshot-extra-entry	0.502s
FAIL

```

## offline-mutant-rev2-final.log

```text
=== RUN   TestMuse148OfflineBoundary
    replay_test.go:379: network deny exit=0: MUSE148_NETWORK_CONTROL
--- FAIL: TestMuse148OfflineBoundary (0.21s)
FAIL
FAIL	github.com/relux-works/curator/.temp/TASK-261016-i5gwaf/mutants-rev2-final/offline-loopback	0.518s
FAIL

```

## mutants/checker-duplicate-count.patch

```diff
--- a/replay_test.go
+++ b/replay_test.go
@@ -66,6 +66,7 @@
 			fields[key] = append(fields[key], value)
 		}
 		for _, key := range append([]string{"event", "outcome"}, countKeys(want)...) {
+			if key == "user_skills" && len(fields[key]) == 2 && fields[key][0] == "0" && fields[key][1] == "1" { fields[key] = fields[key][:1] }
 			if len(fields[key]) != 1 {
 				return fmt.Errorf("%s missing or duplicate %s", event, key)
 			}

```

## mutants/checker-duplicate-outcome.patch

```diff
--- a/replay_test.go
+++ b/replay_test.go
@@ -66,6 +66,7 @@
 			fields[key] = append(fields[key], value)
 		}
 		for _, key := range append([]string{"event", "outcome"}, countKeys(want)...) {
+			if key == "outcome" && len(fields[key]) == 2 && fields[key][0] == `"failed"` && fields[key][1] == `"ready"` { fields[key] = fields[key][1:] }
 			if len(fields[key]) != 1 {
 				return fmt.Errorf("%s missing or duplicate %s", event, key)
 			}

```

## mutants/checker-one-skill.patch

```diff
--- a/replay_test.go
+++ b/replay_test.go
@@ -82,7 +82,7 @@
 			if err != nil || !regexp.MustCompile(`^[0-9]+$`).MatchString(value) {
 				return fmt.Errorf("%s invalid %s", event, key)
 			}
-			if got != n {
+			if got != n && !(event == "skills.catalog_load" && key == "user_skills" && n == 0 && got == 1) {
 				return fmt.Errorf("%s observed %s=%d; want %d", event, key, got, n)
 			}
 		}

```

## mutants/offline-loopback.patch

```diff
--- a/replay_test.go
+++ b/replay_test.go
@@ -161,6 +161,7 @@
 	for _, rel := range []string{".claude", ".codex", ".config", ".local", ".muse", "Library"} {
 		sb += "(deny file-read* file-write* (subpath " + strconv.Quote(filepath.Join(native, rel)) + "))\n"
 	}
+	sb += "(allow network-outbound (remote tcp \"localhost:*\"))\n"
 	write(t, p.sandbox, []byte(sb))
 	return p
 }

```

## mutants/snapshot-extra-entry.patch

```diff
--- a/replay_test.go
+++ b/replay_test.go
@@ -447,6 +447,7 @@
 		if d.IsDir() {
 			return nil
 		}
+		if rel, _ := filepath.Rel(root, path); filepath.ToSlash(rel) == "home/extra" { return nil }
 		count++
 		info, err := d.Info()
 		if err != nil {

```

## mutants/snapshot-go-listed-link.patch

```diff
--- a/replay_test.go
+++ b/replay_test.go
@@ -456,6 +456,7 @@
 		if err != nil {
 			return err
 		}
+		if filepath.ToSlash(rel) == entry.Path && info.Mode()&os.ModeSymlink != 0 { info, err = os.Stat(path); if err != nil { return err } }
 		if filepath.ToSlash(rel) != entry.Path || !info.Mode().IsRegular() || (private && info.Mode().Perm() != 0600) {
 			return fmt.Errorf("fixture entry type/mode/set")
 		}

```

## mutants/snapshot-reference-dangling-link.patch

```diff
--- a/reference-snapshot-check.py
+++ b/reference-snapshot-check.py
@@ -21,6 +21,8 @@
             elif stat.S_ISREG(mode):
                 content[path.relative_to(root).as_posix()] = (path.read_bytes(), f'{mode & 0o777:04o}')
             else:
+                if path.relative_to(root).as_posix() == 'home/hidden-link' and stat.S_ISLNK(mode):
+                    continue
                 raise ValueError('physical-entry-type')
     return content
 

```

## mutants/snapshot-reference-listed-link.patch

```diff
--- a/reference-snapshot-check.py
+++ b/reference-snapshot-check.py
@@ -16,6 +16,8 @@
         for entry in entries:
             path = Path(entry.path)
             mode = entry.stat(follow_symlinks=False).st_mode
+            if path.relative_to(root).as_posix() == "home/config/muse/AGENTS.md" and stat.S_ISLNK(mode):
+                mode = entry.stat(follow_symlinks=True).st_mode
             if stat.S_ISDIR(mode):
                 content.update(inventory(path))
             elif stat.S_ISREG(mode):

```

## Fresh hosted result (reused exact default-suite identity)

```json
{"conclusion":"success","headSha":"9e6b55f4128ac8e92f11398ce7a23089fcd8a72d","jobs":[{"completedAt":"2026-10-16T01:52:49Z","conclusion":"success","databaseId":116763873475,"name":"Test (ubuntu-latest)","startedAt":"2026-10-16T01:31:36Z","status":"completed","steps":[{"completedAt":"2026-10-16T01:31:39Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T01:31:36Z","status":"completed"},{"completedAt":"2026-10-16T01:31:47Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T01:31:39Z","status":"completed"},{"completedAt":"2026-10-16T01:31:48Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T01:31:47Z","status":"completed"},{"completedAt":"2026-10-16T01:31:56Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T01:31:48Z","status":"completed"},{"completedAt":"2026-10-16T01:31:56Z","conclusion":"success","name":"Run actions/setup-node@v4","number":5,"startedAt":"2026-10-16T01:31:56Z","status":"completed"},{"completedAt":"2026-10-16T01:31:56Z","conclusion":"success","name":"Verify pnpm pin matches Go constant","number":6,"startedAt":"2026-10-16T01:31:56Z","status":"completed"},{"completedAt":"2026-10-16T01:31:59Z","conclusion":"success","name":"Install pinned pnpm via npm","number":7,"startedAt":"2026-10-16T01:31:56Z","status":"completed"},{"completedAt":"2026-10-16T01:31:59Z","conclusion":"success","name":"Verify rust pin matches Go constant","number":8,"startedAt":"2026-10-16T01:31:59Z","status":"completed"},{"completedAt":"2026-10-16T01:32:07Z","conclusion":"success","name":"Install pinned Rust toolchain via rustup","number":9,"startedAt":"2026-10-16T01:31:59Z","status":"completed"},{"completedAt":"2026-10-16T01:32:07Z","conclusion":"success","name":"Verify Go toolchain identity","number":10,"startedAt":"2026-10-16T01:32:07Z","status":"completed"},{"completedAt":"2026-10-16T01:32:07Z","conclusion":"success","name":"Verify versioned Go install compatibility","number":11,"startedAt":"2026-10-16T01:32:07Z","status":"completed"},{"completedAt":"2026-10-16T01:32:07Z","conclusion":"success","name":"gofmt check","number":12,"startedAt":"2026-10-16T01:32:07Z","status":"completed"},{"completedAt":"2026-10-16T01:32:23Z","conclusion":"success","name":"go vet","number":13,"startedAt":"2026-10-16T01:32:07Z","status":"completed"},{"completedAt":"2026-10-16T01:32:26Z","conclusion":"success","name":"Ledger consistency","number":14,"startedAt":"2026-10-16T01:32:23Z","status":"completed"},{"completedAt":"2026-10-16T01:52:43Z","conclusion":"success","name":"go test + platform-case gate","number":15,"startedAt":"2026-10-16T01:32:26Z","status":"completed"},{"completedAt":"2026-10-16T01:52:45Z","conclusion":"success","name":"Upload gate evidence","number":16,"startedAt":"2026-10-16T01:52:43Z","status":"completed"},{"completedAt":"2026-10-16T01:52:45Z","conclusion":"success","name":"Post Run actions/setup-node@v4","number":29,"startedAt":"2026-10-16T01:52:45Z","status":"completed"},{"completedAt":"2026-10-16T01:52:45Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":30,"startedAt":"2026-10-16T01:52:45Z","status":"completed"},{"completedAt":"2026-10-16T01:52:45Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":31,"startedAt":"2026-10-16T01:52:45Z","status":"completed"},{"completedAt":"2026-10-16T01:52:45Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":32,"startedAt":"2026-10-16T01:52:45Z","status":"completed"},{"completedAt":"2026-10-16T01:52:45Z","conclusion":"success","name":"Complete job","number":33,"startedAt":"2026-10-16T01:52:45Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38835589308/job/116763873475"},{"completedAt":"2026-10-16T01:32:05Z","conclusion":"success","databaseId":116763873644,"name":"Gate self-test (ubuntu-latest)","startedAt":"2026-10-16T01:31:37Z","status":"completed","steps":[{"completedAt":"2026-10-16T01:31:38Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T01:31:37Z","status":"completed"},{"completedAt":"2026-10-16T01:31:45Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T01:31:38Z","status":"completed"},{"completedAt":"2026-10-16T01:32:03Z","conclusion":"success","name":"Self-test the CI gates","number":3,"startedAt":"2026-10-16T01:31:45Z","status":"completed"},{"completedAt":"2026-10-16T01:32:03Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":6,"startedAt":"2026-10-16T01:32:03Z","status":"completed"},{"completedAt":"2026-10-16T01:32:03Z","conclusion":"success","name":"Complete job","number":7,"startedAt":"2026-10-16T01:32:03Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38835589308/job/116763873644"},{"completedAt":"2026-10-16T01:36:38Z","conclusion":"success","databaseId":116763873715,"name":"Gate self-test (macos-latest)","startedAt":"2026-10-16T01:36:12Z","status":"completed","steps":[{"completedAt":"2026-10-16T01:36:14Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T01:36:13Z","status":"completed"},{"completedAt":"2026-10-16T01:36:26Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T01:36:14Z","status":"completed"},{"completedAt":"2026-10-16T01:36:35Z","conclusion":"success","name":"Self-test the CI gates","number":3,"startedAt":"2026-10-16T01:36:26Z","status":"completed"},{"completedAt":"2026-10-16T01:36:35Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":6,"startedAt":"2026-10-16T01:36:35Z","status":"completed"},{"completedAt":"2026-10-16T01:36:36Z","conclusion":"success","name":"Complete job","number":7,"startedAt":"2026-10-16T01:36:35Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38835589308/job/116763873715"},{"completedAt":"2026-10-16T01:32:15Z","conclusion":"success","databaseId":116763873726,"name":"Go 1.27.1 driver (ubuntu-latest)","startedAt":"2026-10-16T01:31:36Z","status":"completed","steps":[{"completedAt":"2026-10-16T01:31:38Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T01:31:37Z","status":"completed"},{"completedAt":"2026-10-16T01:31:45Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T01:31:38Z","status":"completed"},{"completedAt":"2026-10-16T01:31:46Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T01:31:45Z","status":"completed"},{"completedAt":"2026-10-16T01:31:56Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T01:31:46Z","status":"completed"},{"completedAt":"2026-10-16T01:31:57Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T01:31:56Z","status":"completed"},{"completedAt":"2026-10-16T01:32:00Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T01:31:57Z","status":"completed"},{"completedAt":"2026-10-16T01:32:02Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T01:32:00Z","status":"completed"},{"completedAt":"2026-10-16T01:32:11Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T01:32:02Z","status":"completed"},{"completedAt":"2026-10-16T01:32:12Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T01:32:11Z","status":"completed"},{"completedAt":"2026-10-16T01:32:12Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T01:32:12Z","status":"completed"},{"completedAt":"2026-10-16T01:32:12Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T01:32:12Z","status":"completed"},{"completedAt":"2026-10-16T01:32:12Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T01:32:12Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38835589308/job/116763873726"},{"completedAt":"2026-10-16T02:20:56Z","conclusion":"success","databaseId":116763873754,"name":"Test (windows-latest)","startedAt":"2026-10-16T01:32:01Z","status":"completed","steps":[{"completedAt":"2026-10-16T01:32:05Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T01:32:02Z","status":"completed"},{"completedAt":"2026-10-16T01:32:23Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T01:32:05Z","status":"completed"},{"completedAt":"2026-10-16T01:32:26Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T01:32:23Z","status":"completed"},{"completedAt":"2026-10-16T01:33:04Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T01:32:26Z","status":"completed"},{"completedAt":"2026-10-16T01:33:10Z","conclusion":"success","name":"Run actions/setup-node@v4","number":5,"startedAt":"2026-10-16T01:33:04Z","status":"completed"},{"completedAt":"2026-10-16T01:33:10Z","conclusion":"success","name":"Verify pnpm pin matches Go constant","number":6,"startedAt":"2026-10-16T01:33:10Z","status":"completed"},{"completedAt":"2026-10-16T01:33:19Z","conclusion":"success","name":"Install pinned pnpm via npm","number":7,"startedAt":"2026-10-16T01:33:10Z","status":"completed"},{"completedAt":"2026-10-16T01:33:19Z","conclusion":"success","name":"Verify rust pin matches Go constant","number":8,"startedAt":"2026-10-16T01:33:19Z","status":"completed"},{"completedAt":"2026-10-16T01:33:31Z","conclusion":"success","name":"Install pinned Rust toolchain via rustup","number":9,"startedAt":"2026-10-16T01:33:19Z","status":"completed"},{"completedAt":"2026-10-16T01:33:32Z","conclusion":"success","name":"Verify Go toolchain identity","number":10,"startedAt":"2026-10-16T01:33:31Z","status":"completed"},{"completedAt":"2026-10-16T01:33:32Z","conclusion":"success","name":"Verify versioned Go install compatibility","number":11,"startedAt":"2026-10-16T01:33:32Z","status":"completed"},{"completedAt":"2026-10-16T01:33:32Z","conclusion":"skipped","name":"gofmt check","number":12,"startedAt":"2026-10-16T01:33:32Z","status":"completed"},{"completedAt":"2026-10-16T01:33:50Z","conclusion":"success","name":"go vet","number":13,"startedAt":"2026-10-16T01:33:32Z","status":"completed"},{"completedAt":"2026-10-16T01:34:34Z","conclusion":"success","name":"Ledger consistency","number":14,"startedAt":"2026-10-16T01:33:50Z","status":"completed"},{"completedAt":"2026-10-16T02:20:48Z","conclusion":"success","name":"go test + platform-case gate","number":15,"startedAt":"2026-10-16T01:34:34Z","status":"completed"},{"completedAt":"2026-10-16T02:20:49Z","conclusion":"success","name":"Upload gate evidence","number":16,"startedAt":"2026-10-16T02:20:48Z","status":"completed"},{"completedAt":"2026-10-16T02:20:49Z","conclusion":"success","name":"Post Run actions/setup-node@v4","number":29,"startedAt":"2026-10-16T02:20:49Z","status":"completed"},{"completedAt":"2026-10-16T02:20:49Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":30,"startedAt":"2026-10-16T02:20:49Z","status":"completed"},{"completedAt":"2026-10-16T02:20:52Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":31,"startedAt":"2026-10-16T02:20:49Z","status":"completed"},{"completedAt":"2026-10-16T02:20:54Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":32,"startedAt":"2026-10-16T02:20:52Z","status":"completed"},{"completedAt":"2026-10-16T02:20:54Z","conclusion":"success","name":"Complete job","number":33,"startedAt":"2026-10-16T02:20:54Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38835589308/job/116763873754"},{"completedAt":"2026-10-16T02:14:05Z","conclusion":"success","databaseId":116763873775,"name":"Race (macos-latest)","startedAt":"2026-10-16T01:38:45Z","status":"completed","steps":[{"completedAt":"2026-10-16T01:38:48Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T01:38:46Z","status":"completed"},{"completedAt":"2026-10-16T01:39:01Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T01:38:48Z","status":"completed"},{"completedAt":"2026-10-16T01:39:03Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T01:39:01Z","status":"completed"},{"completedAt":"2026-10-16T01:39:35Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T01:39:03Z","status":"completed"},{"completedAt":"2026-10-16T01:39:36Z","conclusion":"success","name":"Run actions/setup-node@v4","number":5,"startedAt":"2026-10-16T01:39:35Z","status":"completed"},{"completedAt":"2026-10-16T01:39:36Z","conclusion":"success","name":"Verify pnpm pin matches Go constant","number":6,"startedAt":"2026-10-16T01:39:36Z","status":"completed"},{"completedAt":"2026-10-16T01:39:38Z","conclusion":"success","name":"Install pinned pnpm via npm","number":7,"startedAt":"2026-10-16T01:39:36Z","status":"completed"},{"completedAt":"2026-10-16T01:39:38Z","conclusion":"success","name":"Verify rust pin matches Go constant","number":8,"startedAt":"2026-10-16T01:39:38Z","status":"completed"},{"completedAt":"2026-10-16T01:39:47Z","conclusion":"success","name":"Install pinned Rust toolchain via rustup","number":9,"startedAt":"2026-10-16T01:39:38Z","status":"completed"},{"completedAt":"2026-10-16T01:39:47Z","conclusion":"success","name":"Verify Go toolchain identity","number":10,"startedAt":"2026-10-16T01:39:47Z","status":"completed"},{"completedAt":"2026-10-16T02:13:52Z","conclusion":"success","name":"go test -race + platform-case gate","number":11,"startedAt":"2026-10-16T01:39:47Z","status":"completed"},{"completedAt":"2026-10-16T02:13:54Z","conclusion":"success","name":"Upload race evidence","number":12,"startedAt":"2026-10-16T02:13:52Z","status":"completed"},{"completedAt":"2026-10-16T02:13:54Z","conclusion":"success","name":"Post Run actions/setup-node@v4","number":21,"startedAt":"2026-10-16T02:13:54Z","status":"completed"},{"completedAt":"2026-10-16T02:13:55Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":22,"startedAt":"2026-10-16T02:13:54Z","status":"completed"},{"completedAt":"2026-10-16T02:13:56Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":23,"startedAt":"2026-10-16T02:13:55Z","status":"completed"},{"completedAt":"2026-10-16T02:13:57Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":24,"startedAt":"2026-10-16T02:13:56Z","status":"completed"},{"completedAt":"2026-10-16T02:14:00Z","conclusion":"success","name":"Complete job","number":25,"startedAt":"2026-10-16T02:13:57Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38835589308/job/116763873775"},{"completedAt":"2026-10-16T01:38:37Z","conclusion":"success","databaseId":116763873816,"name":"Go 1.26.0 driver (macos-latest)","startedAt":"2026-10-16T01:36:45Z","status":"completed","steps":[{"completedAt":"2026-10-16T01:36:46Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T01:36:45Z","status":"completed"},{"completedAt":"2026-10-16T01:36:57Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T01:36:46Z","status":"completed"},{"completedAt":"2026-10-16T01:36:59Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T01:36:57Z","status":"completed"},{"completedAt":"2026-10-16T01:37:21Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T01:36:59Z","status":"completed"},{"completedAt":"2026-10-16T01:37:21Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T01:37:21Z","status":"completed"},{"completedAt":"2026-10-16T01:38:21Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T01:37:21Z","status":"completed"},{"completedAt":"2026-10-16T01:38:22Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T01:38:21Z","status":"completed"},{"completedAt":"2026-10-16T01:38:33Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T01:38:22Z","status":"completed"},{"completedAt":"2026-10-16T01:38:33Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T01:38:33Z","status":"completed"},{"completedAt":"2026-10-16T01:38:33Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T01:38:33Z","status":"completed"},{"completedAt":"2026-10-16T01:38:34Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T01:38:33Z","status":"completed"},{"completedAt":"2026-10-16T01:38:34Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T01:38:34Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38835589308/job/116763873816"},{"completedAt":"2026-10-16T01:32:20Z","conclusion":"success","databaseId":116763873842,"name":"Go 1.26.0 driver (ubuntu-latest)","startedAt":"2026-10-16T01:31:36Z","status":"completed","steps":[{"completedAt":"2026-10-16T01:31:39Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T01:31:37Z","status":"completed"},{"completedAt":"2026-10-16T01:31:46Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T01:31:39Z","status":"completed"},{"completedAt":"2026-10-16T01:31:47Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T01:31:46Z","status":"completed"},{"completedAt":"2026-10-16T01:32:03Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T01:31:47Z","status":"completed"},{"completedAt":"2026-10-16T01:32:03Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T01:32:03Z","status":"completed"},{"completedAt":"2026-10-16T01:32:06Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T01:32:03Z","status":"completed"},{"completedAt":"2026-10-16T01:32:07Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T01:32:06Z","status":"completed"},{"completedAt":"2026-10-16T01:32:16Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T01:32:07Z","status":"completed"},{"completedAt":"2026-10-16T01:32:16Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T01:32:16Z","status":"completed"},{"completedAt":"2026-10-16T01:32:16Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T01:32:16Z","status":"completed"},{"completedAt":"2026-10-16T01:32:16Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T01:32:16Z","status":"completed"},{"completedAt":"2026-10-16T01:32:16Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T01:32:16Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38835589308/job/116763873842"},{"completedAt":"2026-10-16T01:57:41Z","conclusion":"success","databaseId":116763873847,"name":"Test (macos-latest)","startedAt":"2026-10-16T01:39:10Z","status":"completed","steps":[{"completedAt":"2026-10-16T01:39:11Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T01:39:10Z","status":"completed"},{"completedAt":"2026-10-16T01:39:26Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T01:39:11Z","status":"completed"},{"completedAt":"2026-10-16T01:39:27Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T01:39:26Z","status":"completed"},{"completedAt":"2026-10-16T01:39:48Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T01:39:27Z","status":"completed"},{"completedAt":"2026-10-16T01:39:49Z","conclusion":"success","name":"Run actions/setup-node@v4","number":5,"startedAt":"2026-10-16T01:39:48Z","status":"completed"},{"completedAt":"2026-10-16T01:39:49Z","conclusion":"success","name":"Verify pnpm pin matches Go constant","number":6,"startedAt":"2026-10-16T01:39:49Z","status":"completed"},{"completedAt":"2026-10-16T01:39:50Z","conclusion":"success","name":"Install pinned pnpm via npm","number":7,"startedAt":"2026-10-16T01:39:49Z","status":"completed"},{"completedAt":"2026-10-16T01:39:50Z","conclusion":"success","name":"Verify rust pin matches Go constant","number":8,"startedAt":"2026-10-16T01:39:50Z","status":"completed"},{"completedAt":"2026-10-16T01:39:58Z","conclusion":"success","name":"Install pinned Rust toolchain via rustup","number":9,"startedAt":"2026-10-16T01:39:50Z","status":"completed"},{"completedAt":"2026-10-16T01:39:58Z","conclusion":"success","name":"Verify Go toolchain identity","number":10,"startedAt":"2026-10-16T01:39:58Z","status":"completed"},{"completedAt":"2026-10-16T01:39:58Z","conclusion":"success","name":"Verify versioned Go install compatibility","number":11,"startedAt":"2026-10-16T01:39:58Z","status":"completed"},{"completedAt":"2026-10-16T01:39:58Z","conclusion":"success","name":"gofmt check","number":12,"startedAt":"2026-10-16T01:39:58Z","status":"completed"},{"completedAt":"2026-10-16T01:40:09Z","conclusion":"success","name":"go vet","number":13,"startedAt":"2026-10-16T01:39:58Z","status":"completed"},{"completedAt":"2026-10-16T01:40:17Z","conclusion":"success","name":"Ledger consistency","number":14,"startedAt":"2026-10-16T01:40:09Z","status":"completed"},{"completedAt":"2026-10-16T01:57:31Z","conclusion":"success","name":"go test + platform-case gate","number":15,"startedAt":"2026-10-16T01:40:17Z","status":"completed"},{"completedAt":"2026-10-16T01:57:33Z","conclusion":"success","name":"Upload gate evidence","number":16,"startedAt":"2026-10-16T01:57:31Z","status":"completed"},{"completedAt":"2026-10-16T01:57:33Z","conclusion":"success","name":"Post Run actions/setup-node@v4","number":29,"startedAt":"2026-10-16T01:57:33Z","status":"completed"},{"completedAt":"2026-10-16T01:57:33Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":30,"startedAt":"2026-10-16T01:57:33Z","status":"completed"},{"completedAt":"2026-10-16T01:57:34Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":31,"startedAt":"2026-10-16T01:57:33Z","status":"completed"},{"completedAt":"2026-10-16T01:57:35Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":32,"startedAt":"2026-10-16T01:57:34Z","status":"completed"},{"completedAt":"2026-10-16T01:57:38Z","conclusion":"success","name":"Complete job","number":33,"startedAt":"2026-10-16T01:57:35Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38835589308/job/116763873847"},{"completedAt":"2026-10-16T01:36:18Z","conclusion":"success","databaseId":116763873867,"name":"Go 1.26.0 driver (windows-latest)","startedAt":"2026-10-16T01:32:16Z","status":"completed","steps":[{"completedAt":"2026-10-16T01:32:17Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T01:32:17Z","status":"completed"},{"completedAt":"2026-10-16T01:32:29Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T01:32:18Z","status":"completed"},{"completedAt":"2026-10-16T01:32:32Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T01:32:29Z","status":"completed"},{"completedAt":"2026-10-16T01:33:05Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T01:32:32Z","status":"completed"},{"completedAt":"2026-10-16T01:33:06Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T01:33:05Z","status":"completed"},{"completedAt":"2026-10-16T01:35:43Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T01:33:06Z","status":"completed"},{"completedAt":"2026-10-16T01:35:47Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T01:35:43Z","status":"completed"},{"completedAt":"2026-10-16T01:36:10Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T01:35:47Z","status":"completed"},{"completedAt":"2026-10-16T01:36:10Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T01:36:10Z","status":"completed"},{"completedAt":"2026-10-16T01:36:13Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T01:36:10Z","status":"completed"},{"completedAt":"2026-10-16T01:36:15Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T01:36:13Z","status":"completed"},{"completedAt":"2026-10-16T01:36:15Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T01:36:15Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38835589308/job/116763873867"},{"completedAt":"2026-10-16T01:32:36Z","conclusion":"success","databaseId":116763873870,"name":"Interop conformance gate","startedAt":"2026-10-16T01:32:07Z","status":"completed","steps":[{"completedAt":"2026-10-16T01:32:09Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T01:32:08Z","status":"completed"},{"completedAt":"2026-10-16T01:32:18Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T01:32:09Z","status":"completed"},{"completedAt":"2026-10-16T01:32:19Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T01:32:18Z","status":"completed"},{"completedAt":"2026-10-16T01:32:31Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T01:32:19Z","status":"completed"},{"completedAt":"2026-10-16T01:32:31Z","conclusion":"success","name":"Verify Go toolchain identity","number":5,"startedAt":"2026-10-16T01:32:31Z","status":"completed"},{"completedAt":"2026-10-16T01:32:33Z","conclusion":"success","name":"Shared protocol suite","number":6,"startedAt":"2026-10-16T01:32:31Z","status":"completed"},{"completedAt":"2026-10-16T01:32:33Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":10,"startedAt":"2026-10-16T01:32:33Z","status":"completed"},{"completedAt":"2026-10-16T01:32:33Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":11,"startedAt":"2026-10-16T01:32:33Z","status":"completed"},{"completedAt":"2026-10-16T01:32:34Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":12,"startedAt":"2026-10-16T01:32:33Z","status":"completed"},{"completedAt":"2026-10-16T01:32:34Z","conclusion":"success","name":"Complete job","number":13,"startedAt":"2026-10-16T01:32:34Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38835589308/job/116763873870"},{"completedAt":"2026-10-16T01:44:20Z","conclusion":"success","databaseId":116763873905,"name":"Go 1.25.5 driver (macos-latest)","startedAt":"2026-10-16T01:42:27Z","status":"completed","steps":[{"completedAt":"2026-10-16T01:42:30Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T01:42:28Z","status":"completed"},{"completedAt":"2026-10-16T01:42:38Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T01:42:30Z","status":"completed"},{"completedAt":"2026-10-16T01:42:39Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T01:42:38Z","status":"completed"},{"completedAt":"2026-10-16T01:43:01Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T01:42:39Z","status":"completed"},{"completedAt":"2026-10-16T01:43:01Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T01:43:01Z","status":"completed"},{"completedAt":"2026-10-16T01:44:03Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T01:43:01Z","status":"completed"},{"completedAt":"2026-10-16T01:44:05Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T01:44:03Z","status":"completed"},{"completedAt":"2026-10-16T01:44:14Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T01:44:05Z","status":"completed"},{"completedAt":"2026-10-16T01:44:14Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T01:44:14Z","status":"completed"},{"completedAt":"2026-10-16T01:44:15Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T01:44:14Z","status":"completed"},{"completedAt":"2026-10-16T01:44:15Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T01:44:15Z","status":"completed"},{"completedAt":"2026-10-16T01:44:17Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T01:44:15Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38835589308/job/116763873905"},{"completedAt":"2026-10-16T01:34:32Z","conclusion":"success","databaseId":116763873907,"name":"Gate self-test (windows-latest)","startedAt":"2026-10-16T01:32:44Z","status":"completed","steps":[{"completedAt":"2026-10-16T01:32:45Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T01:32:45Z","status":"completed"},{"completedAt":"2026-10-16T01:33:00Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T01:32:46Z","status":"completed"},{"completedAt":"2026-10-16T01:34:29Z","conclusion":"success","name":"Self-test the CI gates","number":3,"startedAt":"2026-10-16T01:33:00Z","status":"completed"},{"completedAt":"2026-10-16T01:34:31Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":6,"startedAt":"2026-10-16T01:34:29Z","status":"completed"},{"completedAt":"2026-10-16T01:34:31Z","conclusion":"success","name":"Complete job","number":7,"startedAt":"2026-10-16T01:34:31Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38835589308/job/116763873907"},{"completedAt":"2026-10-16T01:33:37Z","conclusion":"success","databaseId":116763873949,"name":"Lint","startedAt":"2026-10-16T01:32:38Z","status":"completed","steps":[{"completedAt":"2026-10-16T01:32:39Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T01:32:38Z","status":"completed"},{"completedAt":"2026-10-16T01:32:47Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T01:32:39Z","status":"completed"},{"completedAt":"2026-10-16T01:32:56Z","conclusion":"success","name":"Run actions/setup-go@v5","number":3,"startedAt":"2026-10-16T01:32:47Z","status":"completed"},{"completedAt":"2026-10-16T01:32:56Z","conclusion":"success","name":"Verify Go toolchain identity","number":4,"startedAt":"2026-10-16T01:32:56Z","status":"completed"},{"completedAt":"2026-10-16T01:33:34Z","conclusion":"success","name":"Run golangci/golangci-lint-action@v7","number":5,"startedAt":"2026-10-16T01:32:56Z","status":"completed"},{"completedAt":"2026-10-16T01:33:34Z","conclusion":"success","name":"No broad suppression","number":6,"startedAt":"2026-10-16T01:33:34Z","status":"completed"},{"completedAt":"2026-10-16T01:33:35Z","conclusion":"success","name":"Verify GoReleaser rc channel values","number":7,"startedAt":"2026-10-16T01:33:34Z","status":"completed"},{"completedAt":"2026-10-16T01:33:35Z","conclusion":"success","name":"Post Run golangci/golangci-lint-action@v7","number":12,"startedAt":"2026-10-16T01:33:35Z","status":"completed"},{"completedAt":"2026-10-16T01:33:35Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":13,"startedAt":"2026-10-16T01:33:35Z","status":"completed"},{"completedAt":"2026-10-16T01:33:35Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":14,"startedAt":"2026-10-16T01:33:35Z","status":"completed"},{"completedAt":"2026-10-16T01:33:35Z","conclusion":"success","name":"Complete job","number":15,"startedAt":"2026-10-16T01:33:35Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38835589308/job/116763873949"},{"completedAt":"2026-10-16T01:32:43Z","conclusion":"success","databaseId":116763873954,"name":"Naming gate","startedAt":"2026-10-16T01:32:31Z","status":"completed","steps":[{"completedAt":"2026-10-16T01:32:32Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T01:32:32Z","status":"completed"},{"completedAt":"2026-10-16T01:32:40Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T01:32:32Z","status":"completed"},{"completedAt":"2026-10-16T01:32:41Z","conclusion":"success","name":"Employer name gate","number":3,"startedAt":"2026-10-16T01:32:40Z","status":"completed"},{"completedAt":"2026-10-16T01:32:41Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":6,"startedAt":"2026-10-16T01:32:41Z","status":"completed"},{"completedAt":"2026-10-16T01:32:41Z","conclusion":"success","name":"Complete job","number":7,"startedAt":"2026-10-16T01:32:41Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38835589308/job/116763873954"},{"completedAt":"2026-10-16T02:06:07Z","conclusion":"success","databaseId":116763873961,"name":"Race (ubuntu-latest)","startedAt":"2026-10-16T01:33:34Z","status":"completed","steps":[{"completedAt":"2026-10-16T01:33:36Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T01:33:35Z","status":"completed"},{"completedAt":"2026-10-16T01:33:44Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T01:33:36Z","status":"completed"},{"completedAt":"2026-10-16T01:33:45Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T01:33:44Z","status":"completed"},{"completedAt":"2026-10-16T01:33:54Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T01:33:45Z","status":"completed"},{"completedAt":"2026-10-16T01:33:54Z","conclusion":"success","name":"Run actions/setup-node@v4","number":5,"startedAt":"2026-10-16T01:33:54Z","status":"completed"},{"completedAt":"2026-10-16T01:33:54Z","conclusion":"success","name":"Verify pnpm pin matches Go constant","number":6,"startedAt":"2026-10-16T01:33:54Z","status":"completed"},{"completedAt":"2026-10-16T01:33:56Z","conclusion":"success","name":"Install pinned pnpm via npm","number":7,"startedAt":"2026-10-16T01:33:54Z","status":"completed"},{"completedAt":"2026-10-16T01:33:56Z","conclusion":"success","name":"Verify rust pin matches Go constant","number":8,"startedAt":"2026-10-16T01:33:56Z","status":"completed"},{"completedAt":"2026-10-16T01:34:04Z","conclusion":"success","name":"Install pinned Rust toolchain via rustup","number":9,"startedAt":"2026-10-16T01:33:56Z","status":"completed"},{"completedAt":"2026-10-16T01:34:04Z","conclusion":"success","name":"Verify Go toolchain identity","number":10,"startedAt":"2026-10-16T01:34:04Z","status":"completed"},{"completedAt":"2026-10-16T02:06:03Z","conclusion":"success","name":"go test -race + platform-case gate","number":11,"startedAt":"2026-10-16T01:34:04Z","status":"completed"},{"completedAt":"2026-10-16T02:06:04Z","conclusion":"success","name":"Upload race evidence","number":12,"startedAt":"2026-10-16T02:06:03Z","status":"completed"},{"completedAt":"2026-10-16T02:06:04Z","conclusion":"success","name":"Post Run actions/setup-node@v4","number":21,"startedAt":"2026-10-16T02:06:04Z","status":"completed"},{"completedAt":"2026-10-16T02:06:04Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":22,"startedAt":"2026-10-16T02:06:04Z","status":"completed"},{"completedAt":"2026-10-16T02:06:05Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":23,"startedAt":"2026-10-16T02:06:04Z","status":"completed"},{"completedAt":"2026-10-16T02:06:05Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":24,"startedAt":"2026-10-16T02:06:05Z","status":"completed"},{"completedAt":"2026-10-16T02:06:05Z","conclusion":"success","name":"Complete job","number":25,"startedAt":"2026-10-16T02:06:05Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38835589308/job/116763873961"},{"completedAt":"2026-10-16T01:36:11Z","conclusion":"success","databaseId":116763873986,"name":"Go 1.25.5 driver (windows-latest)","startedAt":"2026-10-16T01:32:21Z","status":"completed","steps":[{"completedAt":"2026-10-16T01:32:23Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T01:32:22Z","status":"completed"},{"completedAt":"2026-10-16T01:32:37Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T01:32:23Z","status":"completed"},{"completedAt":"2026-10-16T01:32:41Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T01:32:37Z","status":"completed"},{"completedAt":"2026-10-16T01:33:18Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T01:32:41Z","status":"completed"},{"completedAt":"2026-10-16T01:33:19Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T01:33:18Z","status":"completed"},{"completedAt":"2026-10-16T01:35:43Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T01:33:19Z","status":"completed"},{"completedAt":"2026-10-16T01:35:47Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T01:35:43Z","status":"completed"},{"completedAt":"2026-10-16T01:36:04Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T01:35:47Z","status":"completed"},{"completedAt":"2026-10-16T01:36:04Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T01:36:04Z","status":"completed"},{"completedAt":"2026-10-16T01:36:07Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T01:36:04Z","status":"completed"},{"completedAt":"2026-10-16T01:36:09Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T01:36:07Z","status":"completed"},{"completedAt":"2026-10-16T01:36:09Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T01:36:09Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38835589308/job/116763873986"},{"completedAt":"2026-10-16T01:34:57Z","conclusion":"success","databaseId":116763874024,"name":"Go 1.25.5 driver (ubuntu-latest)","startedAt":"2026-10-16T01:34:22Z","status":"completed","steps":[{"completedAt":"2026-10-16T01:34:24Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T01:34:23Z","status":"completed"},{"completedAt":"2026-10-16T01:34:30Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T01:34:24Z","status":"completed"},{"completedAt":"2026-10-16T01:34:31Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T01:34:30Z","status":"completed"},{"completedAt":"2026-10-16T01:34:40Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T01:34:31Z","status":"completed"},{"completedAt":"2026-10-16T01:34:40Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T01:34:40Z","status":"completed"},{"completedAt":"2026-10-16T01:34:45Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T01:34:40Z","status":"completed"},{"completedAt":"2026-10-16T01:34:46Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T01:34:45Z","status":"completed"},{"completedAt":"2026-10-16T01:34:54Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T01:34:46Z","status":"completed"},{"completedAt":"2026-10-16T01:34:54Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T01:34:54Z","status":"completed"},{"completedAt":"2026-10-16T01:34:54Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T01:34:54Z","status":"completed"},{"completedAt":"2026-10-16T01:34:55Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T01:34:54Z","status":"completed"},{"completedAt":"2026-10-16T01:34:55Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T01:34:55Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38835589308/job/116763874024"},{"completedAt":"2026-10-16T01:47:20Z","conclusion":"success","databaseId":116763874051,"name":"Go 1.27.1 driver (macos-latest)","startedAt":"2026-10-16T01:44:28Z","status":"completed","steps":[{"completedAt":"2026-10-16T01:44:31Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T01:44:30Z","status":"completed"},{"completedAt":"2026-10-16T01:44:42Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T01:44:31Z","status":"completed"},{"completedAt":"2026-10-16T01:44:44Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T01:44:42Z","status":"completed"},{"completedAt":"2026-10-16T01:45:13Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T01:44:44Z","status":"completed"},{"completedAt":"2026-10-16T01:45:13Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T01:45:13Z","status":"completed"},{"completedAt":"2026-10-16T01:46:51Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T01:45:13Z","status":"completed"},{"completedAt":"2026-10-16T01:46:54Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T01:46:51Z","status":"completed"},{"completedAt":"2026-10-16T01:47:11Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T01:46:54Z","status":"completed"},{"completedAt":"2026-10-16T01:47:12Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T01:47:11Z","status":"completed"},{"completedAt":"2026-10-16T01:47:12Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T01:47:12Z","status":"completed"},{"completedAt":"2026-10-16T01:47:13Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T01:47:12Z","status":"completed"},{"completedAt":"2026-10-16T01:47:15Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T01:47:13Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38835589308/job/116763874051"},{"completedAt":"2026-10-16T01:36:43Z","conclusion":"success","databaseId":116763874070,"name":"Go 1.27.1 driver (windows-latest)","startedAt":"2026-10-16T01:34:15Z","status":"completed","steps":[{"completedAt":"2026-10-16T01:34:17Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T01:34:16Z","status":"completed"},{"completedAt":"2026-10-16T01:34:28Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T01:34:17Z","status":"completed"},{"completedAt":"2026-10-16T01:34:31Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T01:34:28Z","status":"completed"},{"completedAt":"2026-10-16T01:34:52Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T01:34:31Z","status":"completed"},{"completedAt":"2026-10-16T01:34:53Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T01:34:52Z","status":"completed"},{"completedAt":"2026-10-16T01:36:23Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T01:34:53Z","status":"completed"},{"completedAt":"2026-10-16T01:36:25Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T01:36:23Z","status":"completed"},{"completedAt":"2026-10-16T01:36:38Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T01:36:25Z","status":"completed"},{"completedAt":"2026-10-16T01:36:38Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T01:36:38Z","status":"completed"},{"completedAt":"2026-10-16T01:36:39Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T01:36:38Z","status":"completed"},{"completedAt":"2026-10-16T01:36:41Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T01:36:40Z","status":"completed"},{"completedAt":"2026-10-16T01:36:41Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T01:36:41Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38835589308/job/116763874070"},{"completedAt":"2026-10-16T01:31:34Z","conclusion":"skipped","databaseId":116763875045,"name":"Test ($OPTIONAL_NATIVE_RUNNER)","startedAt":"2026-10-16T01:31:35Z","status":"completed","steps":[],"url":"https://github.com/relux-works/curator/actions/runs/38835589308/job/116763875045"},{"completedAt":"2026-10-16T01:31:34Z","conclusion":"skipped","databaseId":116763875198,"name":"Candidate suite (${{ matrix.os }})","startedAt":"2026-10-16T01:31:35Z","status":"completed","steps":[],"url":"https://github.com/relux-works/curator/actions/runs/38835589308/job/116763875198"}],"status":"completed","url":"https://github.com/relux-works/curator/actions/runs/38835589308"}

```

## Incomplete first replay (exit 1, not reused as green)

The first attempt preceded the final switch-only lint and child-test mask cleanup. Matrix passed; snapshot interactive witness lacked /rules and is unverified in this attempt. Fresh final-source replay above supplies the passing evidence.

```text
=== RUN   TestObservationChecker
--- PASS: TestObservationChecker (0.00s)
=== RUN   TestMuse148OfflineBoundary
--- PASS: TestMuse148OfflineBoundary (0.24s)
=== RUN   TestFrozenSnapshot
=== RUN   TestFrozenSnapshot/regular
=== RUN   TestFrozenSnapshot/altered
=== RUN   TestFrozenSnapshot/extra
=== RUN   TestFrozenSnapshot/listed-symlink
=== RUN   TestFrozenSnapshot/unlisted-dangling-link
=== RUN   TestFrozenSnapshot/missing
--- PASS: TestFrozenSnapshot (0.01s)
    --- PASS: TestFrozenSnapshot/regular (0.00s)
    --- PASS: TestFrozenSnapshot/altered (0.00s)
    --- PASS: TestFrozenSnapshot/extra (0.00s)
    --- PASS: TestFrozenSnapshot/listed-symlink (0.00s)
    --- PASS: TestFrozenSnapshot/unlisted-dangling-link (0.00s)
    --- PASS: TestFrozenSnapshot/missing (0.00s)
=== RUN   TestFrozenSnapshotSourceInputs
=== RUN   TestFrozenSnapshotSourceInputs/regular
=== RUN   TestFrozenSnapshotSourceInputs/listed-symlink
=== RUN   TestFrozenSnapshotSourceInputs/unlisted-dangling-link
--- PASS: TestFrozenSnapshotSourceInputs (0.03s)
    --- PASS: TestFrozenSnapshotSourceInputs/regular (0.01s)
    --- PASS: TestFrozenSnapshotSourceInputs/listed-symlink (0.01s)
    --- PASS: TestFrozenSnapshotSourceInputs/unlisted-dangling-link (0.01s)
=== RUN   TestReferenceSnapshotPhysicalClosure
=== RUN   TestReferenceSnapshotPhysicalClosure/regular
=== RUN   TestReferenceSnapshotPhysicalClosure/listed-symlink
=== RUN   TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link
--- PASS: TestReferenceSnapshotPhysicalClosure (1.03s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/regular (0.55s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/listed-symlink (0.24s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link (0.24s)
=== RUN   TestMuse148ParserRefusals
=== RUN   TestMuse148ParserRefusals/interactive-exec-flag
=== RUN   TestMuse148ParserRefusals/interactive-string-false
--- PASS: TestMuse148ParserRefusals (0.36s)
    --- PASS: TestMuse148ParserRefusals/interactive-exec-flag (0.13s)
    --- PASS: TestMuse148ParserRefusals/interactive-string-false (0.06s)
=== RUN   TestMuse148LoaderMatrix
=== RUN   TestMuse148LoaderMatrix/exec/absent/claude-rules/exclude-false
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/claude-rules/exclude-true
    replay_test.go:789: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/codex-rules/exclude-false
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/codex-rules/exclude-true
    replay_test.go:789: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/skills/exclude-false
    replay_test.go:789: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/skills/exclude-true
    replay_test.go:789: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/claude-rules/exclude-false
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/claude-rules/exclude-true
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/codex-rules/exclude-false
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/codex-rules/exclude-true
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/skills/exclude-false
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/skills/exclude-true
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/claude-rules/exclude-false
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/claude-rules/exclude-true
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/codex-rules/exclude-false
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/codex-rules/exclude-true
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-false
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-true
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/claude-rules/exclude-false
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/claude-rules/exclude-true
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/codex-rules/exclude-false
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/codex-rules/exclude-true
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/skills/exclude-false
    replay_test.go:789: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/skills/exclude-true
    replay_test.go:789: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/claude-rules/exclude-false
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/claude-rules/exclude-true
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/codex-rules/exclude-false
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/codex-rules/exclude-true
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/skills/exclude-false
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/skills/exclude-true
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/claude-rules/exclude-false
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/claude-rules/exclude-true
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/codex-rules/exclude-false
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/codex-rules/exclude-true
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-false
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-true
    replay_test.go:789: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
--- PASS: TestMuse148LoaderMatrix (58.59s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/claude-rules/exclude-false (1.16s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/claude-rules/exclude-true (0.93s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/codex-rules/exclude-false (1.13s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/codex-rules/exclude-true (1.01s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/skills/exclude-false (0.99s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/skills/exclude-true (1.08s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/claude-rules/exclude-false (1.27s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/claude-rules/exclude-true (1.15s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/codex-rules/exclude-false (1.15s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/codex-rules/exclude-true (1.13s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/skills/exclude-false (1.11s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/skills/exclude-true (1.08s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/claude-rules/exclude-false (0.90s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/claude-rules/exclude-true (1.04s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/codex-rules/exclude-false (0.95s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/codex-rules/exclude-true (1.13s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-false (1.03s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-true (1.17s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/claude-rules/exclude-false (1.98s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/claude-rules/exclude-true (2.19s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/codex-rules/exclude-false (1.97s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/codex-rules/exclude-true (2.18s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/skills/exclude-false (2.18s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/skills/exclude-true (2.09s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/claude-rules/exclude-false (2.07s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/claude-rules/exclude-true (2.28s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/codex-rules/exclude-false (2.18s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/codex-rules/exclude-true (2.08s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/skills/exclude-false (1.98s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/skills/exclude-true (2.19s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/claude-rules/exclude-false (2.17s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/claude-rules/exclude-true (1.86s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/codex-rules/exclude-false (1.97s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/codex-rules/exclude-true (2.52s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-false (2.10s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-true (2.33s)
=== RUN   TestMuse148InteractiveExclusionRegression
=== RUN   TestMuse148InteractiveExclusionRegression/linked
    replay_test.go:819: vendor exit=0; required interactive exclusion FAILED (user_skills=1); see docs/qualification/muse-1.4.8-R6256.1.md
=== RUN   TestMuse148InteractiveExclusionRegression/snapshot
    replay_test.go:805: missing real /rules observation: 
          Muse Code 1.4.8
        
          Including your 1 Claude Code personal skill — manage with /settings.
        
          warning: local session messaging unavailable: Operation not permitted (os error 1)
        
        ❯ MUSE148_OFFLINE_CONTROL
        
        ◆ echo: MUSE148_OFFLINE_CONTROL
        
        ── Voice input (⌥ + v to start) ────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
        ❯ /rules
          [import]
        ────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
          echo · $WORKSPACE · Auto-review
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
--- FAIL: TestMuse148InteractiveExclusionRegression (12.02s)
    --- PASS: TestMuse148InteractiveExclusionRegression/linked (2.29s)
    --- FAIL: TestMuse148InteractiveExclusionRegression/snapshot (8.75s)
=== RUN   TestMuse148WorkspaceResidual
=== RUN   TestMuse148WorkspaceResidual/exec/linked
=== RUN   TestMuse148WorkspaceResidual/exec/snapshot
=== RUN   TestMuse148WorkspaceResidual/interactive/linked
=== RUN   TestMuse148WorkspaceResidual/interactive/snapshot
--- PASS: TestMuse148WorkspaceResidual (8.87s)
    --- PASS: TestMuse148WorkspaceResidual/exec/linked (1.42s)
    --- PASS: TestMuse148WorkspaceResidual/exec/snapshot (1.41s)
    --- PASS: TestMuse148WorkspaceResidual/interactive/linked (2.41s)
    --- PASS: TestMuse148WorkspaceResidual/interactive/snapshot (2.52s)
=== RUN   TestReviewContradictoryObservation
=== RUN   TestReviewContradictoryObservation/count-zero-then-one
=== RUN   TestReviewContradictoryObservation/count-one-then-zero
=== RUN   TestReviewContradictoryObservation/outcome-ready-then-failed
=== RUN   TestReviewContradictoryObservation/duplicate-count
=== RUN   TestReviewContradictoryObservation/duplicate-event
=== RUN   TestReviewContradictoryObservation/quoted-decoy
=== RUN   TestReviewContradictoryObservation/overflow-count
=== RUN   TestReviewContradictoryObservation/outcome-failed-then-ready
=== RUN   TestReviewContradictoryObservation/duplicate-outcome
=== RUN   TestReviewContradictoryObservation/malformed-count
--- PASS: TestReviewContradictoryObservation (0.00s)
    --- PASS: TestReviewContradictoryObservation/count-zero-then-one (0.00s)
    --- PASS: TestReviewContradictoryObservation/count-one-then-zero (0.00s)
    --- PASS: TestReviewContradictoryObservation/outcome-ready-then-failed (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-event (0.00s)
    --- PASS: TestReviewContradictoryObservation/quoted-decoy (0.00s)
    --- PASS: TestReviewContradictoryObservation/overflow-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/outcome-failed-then-ready (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-outcome (0.00s)
    --- PASS: TestReviewContradictoryObservation/malformed-count (0.00s)
FAIL
FAIL	github.com/relux-works/curator/.research/muse148	81.445s
FAIL

```

## build-rev2-02.log

```text

```

## vet-rev2-02.log

```text

```

## lint-rev2-01.log

```text
.research/muse148/replay_test.go:500:4: QF1003: could use tagged switch on shape (staticcheck)
			if shape == "altered" {
			^
1 issues:
* staticcheck: 1

```

## lint-rev2-02.log

```text
0 issues.

```

## hosted-source-identity-rev2-03.log

```text

```

## diff-check-rev2-02.log

```text

```

Privacy: the optional native runner label in the hosted result is neutralized; job identity otherwise remains in the private original read and public provider run. Conclusions are unchanged.
