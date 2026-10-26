// Package muse1413 records vendor-loader evidence; it never changes runtime admission.
package muse1413

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

const binarySHA = "9fd40ae9c6dfb847e3ef58b61405c004d0ebbf19ceb0297be966f750b3b84aa0"
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
	binary := os.Getenv("MUSE1413_BINARY")
	if binary == "" {
		t.Skip("vendor replay requires MUSE1413_BINARY, MUSE1413_CURATOR and MUSE1413_RUN_ROOT; no qualification from a skip")
	}
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		t.Fatal("this evidence identity requires darwin/arm64")
	}
	base, err := filepath.Abs(os.Getenv("MUSE1413_RUN_ROOT"))
	if err != nil || os.Getenv("MUSE1413_RUN_ROOT") == "" {
		t.Fatal("explicit task-local run root required")
	}
	if !strings.Contains(base, string(filepath.Separator)+".temp"+string(filepath.Separator)) {
		t.Fatal("run root must be under .temp")
	}
	if digest(read(t, binary)) != binarySHA {
		t.Fatal("unqualified binary digest; this replay does not accept another build")
	}
	root := filepath.Join(base, strings.ReplaceAll(t.Name(), "/", "_"))
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	p := &probe{root: root, binary: binary, curator: os.Getenv("MUSE1413_CURATOR"), sandbox: filepath.Join(root, "probe.sb")}
	if p.curator == "" {
		t.Fatal("MUSE1413_CURATOR required")
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
		write(t, filepath.Join(dir, "native/.claude/CLAUDE.md"), []byte("MUSE1413_FOREIGN_CLAUDE_RULE\n"))
	case "codex-rules":
		write(t, filepath.Join(dir, "native/.codex/AGENTS.md"), []byte("MUSE1413_FOREIGN_CODEX_RULE\n"))
	case "skills":
		write(t, filepath.Join(dir, "native/.claude/skills/muse1413-foreign/SKILL.md"), []byte("---\nname: muse1413-foreign\ndescription: Synthetic foreign personal skill for offline loader qualification.\n---\nMUSE1413_FOREIGN_SKILL\n"))
	}
	if workspace {
		write(t, filepath.Join(dir, "work/AGENTS.md"), []byte("MUSE1413_WORKSPACE_CONFLICT\n"))
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
	if exclude && os.Getenv("MUSE1413_OMIT_EXCLUSION") != form {
		if form == "exec" {
			args = append(args, "--no-foreign-personal-context")
		} else {
			args = append(args, "-c", "context.foreign_personal_skills=false")
		}
	}
	args = append(args, extra...)
	if form == "exec" {
		args = append(args, "MUSE1413_OFFLINE_CONTROL")
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
func TestMuse1413OfflineBoundary(t *testing.T) {
	p := setup(t)
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		_, _ = w.Write([]byte("MUSE1413_NETWORK_CONTROL"))
	}))
	defer server.Close()
	args := []string{"--silent", "--show-error", "--noproxy", "*", "--connect-timeout", "2", server.URL}
	if b, c := command(t, p.root, os.Environ(), "/usr/bin/curl", args...); c != 0 || string(b) != "MUSE1413_NETWORK_CONTROL" {
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
		write(t, filepath.Join(p.root, rel), []byte("MUSE1413_SYNTHETIC_STORE_CONTROL"))
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
		if string(read(t, filepath.Join(p.root, tc.path))) != "MUSE1413_SYNTHETIC_STORE_CONTROL" {
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
		write(t, file, []byte("MUSE1413_FORGED_ROOT\n"))
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
		if string(read(t, file)) != "MUSE1413_FORGED_ROOT\n" {
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
	python, spec := os.Getenv("MUSE1413_SPEC_PYTHON"), os.Getenv("MUSE1413_SPEC")
	if python == "" || spec == "" {
		t.Skip("set MUSE1413_SPEC_PYTHON and MUSE1413_SPEC for the published reference CLI")
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
		b, c := command(t, dir, os.Environ(), tmux, append([]string{"-S", socket, "-f", "/dev/null"}, a...)...)
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
	call("send-keys", "-l", "MUSE1413_OFFLINE_CONTROL")
	time.Sleep(200 * time.Millisecond)
	call("send-keys", "Enter")
	deadline = time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		pane = call("capture-pane", "-p", "-S", "-1000")
		if strings.Contains(string(pane), "echo: MUSE1413_OFFLINE_CONTROL") {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !strings.Contains(string(pane), "echo: MUSE1413_OFFLINE_CONTROL") {
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
func TestMuse1413ParserRefusals(t *testing.T) {
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
func TestMuse1413LoaderMatrix(t *testing.T) {
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
func TestMuse1413InteractiveExclusionRegression(t *testing.T) {
	p := setup(t)
	link := p.fixture(t)
	for _, root := range []string{"linked", "snapshot"} {
		t.Run(root, func(t *testing.T) {
			o := p.run(t, "interactive", root, "skills", true, false, link)
			const record = "docs/qualification/muse-1.4.13-R6996.1.md"
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
func TestMuse1413WorkspaceResidual(t *testing.T) {
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

// Unlike the brief's expected refusal, this exact build accepts the interactive
// config spelling on exec. This records parser acceptance only, not exclusion.
func TestMuse1413ExecInteractiveKeyObservation(t *testing.T) {
	p := setup(t)
	o := p.run(t, "exec", "absent", "none", false, false, "", "-c", "context.foreign_personal_skills=false")
	if o.Exit != 0 {
		t.Fatalf("exec config-key observation changed, exit=%d: %s", o.Exit, o.Output)
	}
	if err := checkCounts(o.Log, "rules.context_load", map[string]int{"user_sources": 0, "foreign_sources": 0}); err != nil {
		t.Fatal(err)
	}
	t.Log("vendor exit=0; exec accepts -c context.foreign_personal_skills=false; exclusion semantics not measured by this parser case")
}
