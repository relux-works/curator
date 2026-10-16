# TASK-261016-i5gwaf — rework 1 probe evidence

Exact-build offline loader observations and controls. No hosted-gate verdict
is claimed by this probe packet. See the separate role handoff for full-gate
results. Exec loader rows pass; interactive skill exclusion FAILED; section
7.9 interactive release gate pending. Model-visible behavior/refresh unmeasured.

## docs/qualification/muse-1.4.8-R6256.1.md

SHA-256: `95864e6719afb6ef6b3ac2b55f435e037d7de5d62f9004cfa950ed3b14e8f1d3`

````text
# Muse 1.4.8-R6256.1 loader qualification record

Task: TASK-261016-i5gwaf, M2, rc.7. This is a vendor observation record,
not runtime admission authority. **Exec loader rows pass; interactive P1–P4
skill exclusion FAILS on this build. The environments §7.9 release gate
remains pending for the interactive form.** Curator qualifies no runtime
itself; M1/M3 own registry/status and consumer changes. No runtime constants,
qualification gates or credential policy change in this leaf.

## Exact identity and observation boundary

| Identity | Value |
|---|---|
| Vendor version | `Muse Code 1.4.8 (1.4.8-R6256.1)` |
| Binary SHA-256 | `5721763b4723f961709cd28f69aa049763965b46c4d4d70f883a0a90e5ddf39c` |
| Binary locator | `$MUSE148_BINARY`; exact absolute path recorded privately in ignored task-local `identity.json` (public-board path hygiene) |
| Platform | macOS 27.0 arm64 |
| Curator checkpoint | `0d00a109ad1da77a5506747c89f24af9442eb341` |
| Normative curator-spec main | `d2a942cf17e4eb017c45475442da1456d8dfbe6c` |
| Harness | Go 1.25.5, tmux 3.6a, Apple sandbox-exec |
| Native root SHA-256 | `8c9af1ca52b75aaa1e67af6365f1e39dfcc3c56a274d3571270079ef669ac9d0` |
| Snapshot manifest SHA-256 | `924378006ba6e6e7db9d0277375456a958c14cbaff03d1ebd63dfa6e4b7ceac8` |
| Snapshot launch SHA-256 | `2991bc9c97904d77aef70bb5c3ea7a7044ff462382310ae613bbd31f461fd5f1` |

The binary locator is the separately scoped vendor executable, never the
updating launcher. The harness refuses a different binary digest. Its
`MUSE148_BINARY` environment gate is opt-in; ordinary `go test ./...` excludes
this hidden evidence package. An unset binary skips vendor tests and proves
no qualification.

Normative sources read: environments §§7.1 P1–P4, 7.9, 10.2; CIP-0017
implementation leaf 2 and minimal probe items 1–4; and
`conformance/environments-muse-v4/README.md`. Current main identity was
rechecked during rework. The 1.4.7 exclusion claim is not promoted to 1.4.8.
CIP credential probes are separate evidence under M2's explicit boundary.

Every vendor invocation runs with `MUSE_NO_AUTO_UPDATE=1`, `env -i`, disposable
HOME and XDG config/data/state/cache directories under task-local `.temp/`,
`TBH_CREDENTIAL_BACKEND=file`, and `--provider echo`. The tested OS sandbox
denies all network, security-service lookups and Keychain file access, plus
real native configuration/library directories. The offline boundary test
pairs an accessible synthetic loopback endpoint with a denied connection
(exit 7 and no second server hit), and readable ordinary bytes with a denied
synthetic Keychains file (exit 1, Operation not permitted, unchanged bytes).
No real Keychain item is inspected. No login, logout or auth command runs.

Echo reaches the real context loader without remote inference. Interactive
startup shows the skill before the echo turn; a local sentinel echo flushes
loader diagnostics, followed by `/rules` and `/exit`. This is pre-remote-
inference loader evidence, not proof of termination before a local echo turn.
No quota/auth failure is used as a network boundary. Credential files remain
absent in each vendor home and fixture source/root bytes remain unchanged.
These observations do not establish credential backend/login/refresh behavior.

## Fixtures, commands and production entrypoints

[Executable probe and fixtures](../../.research/muse148/README.md) contain all
synthetic fixture bytes. `testdata/source` is installed with the real Curator
`profile install SOURCE`, then `env resolve muse --profile muse148-probe
--repair --format json`. The resulting symlink must be a real link and match
the frozen root digest above. A synthetic `{}` auth target is used only for
Curator's v3 metadata check; Muse never runs in that fixture home.

The snapshot is a regular `0600` file at `home/config/muse/AGENTS.md` plus
CCJ-1 manifest/launch metadata declaring all four XDG suffixes. Its complete
single-file closure is independently checked with the published spec
`managed_home_snapshots.check_bundle`. Git does not retain private mode bits;
the harness writes private copies and the independent check sets mode 0600.
This frozen fixture proves representation/loader behavior, not export
provenance or the future `env materialize -> Muse` integration.

Each case invokes the exact real executable through sandbox-exec:

```text
exec:        $MUSE148_BINARY exec --json --provider echo --trust-workspace [--no-foreign-personal-context] MUSE148_OFFLINE_CONTROL
interactive: $MUSE148_BINARY --provider echo --trust-workspace [-c context.foreign_personal_skills=false]
```

Interactive uses a private background tmux terminal, never the active UI.
Parser refusals independently invoke interactive with the exec-only flag and
with `-c context.foreign_personal_skills="false"`. Both must exit **2**, with
`invalid TUI options` / `the JSON type does not fit this setting`, respectively,
and no rules-loader record. Accepted calls exit **0** and expose real records.

Observation channels: `rules.context_load`, `skills.catalog_load`, interactive
startup banner and `/rules`. [Sanitized observations](../../.research/muse148/observations.json)
preserve argv, vendor exit and loader records per case; raw paths stay private.
[Rework observations](../../.research/muse148/observations-rework.json) bind all
44 calls to replay source SHA-256 `f4110c8c614f791d89def4fc04232f411f920ab0cda37fe63d3dd3ae9bc85cdf`.
The first replay's [red test log](../../.research/muse148/replay-03.log)
remains historical evidence. The rework's green regression suite pins that
failure as observed behavior; it does not turn the qualification result green.

## Coverage map

**6 of 6 in-scope AC observation rows driven through production entrypoints;
5 of 6 satisfy the loader contract.** The supplied seventh observation row
expressly requires separate evidence. No catalog-keyed attack-surface table
was supplied: this is a brief gap. The declared surface is vendor-build
evidence authorizing form-specific exclusion; every supplied observation row
is mapped below. Green witness tests and vendor qualification are separate.

| AC observation | Named tests / production call site | Exec | Interactive | Narrowing control or bound |
|---|---|---|---|---|
| Parser acceptance/refusal | `TestMuse148ParserRefusals`, `TestMuse148LoaderMatrix`; real exec/TUI parser | Exact flag accepted | Boolean accepted; exec flag and string false rejected independently, exit 2, before loader | Independent malformed-type and wrong-form calls; no help-only inference |
| Root fallback and persistent skills | `TestMuse148LoaderMatrix`; real rules/skills loader | Observed | Observed | Each isolated foreign source with root absent/present and exclusion absent/present |
| Required exclusion with native root | Matrix plus `TestMuse148InteractiveExclusionRegression`; real skills loader/banner | PASS: 1 → 0 skill | **FAIL: 1 → 1 skill**, linked and regular-file roots | Exec omission killed; interactive omission cannot be distinguished from broken vendor exclusion |
| Boolean without native root | Matrix `interactive/absent`; real TUI loader | Exec's explicit exclusion removes fallback | Fallback remains; skill remains | Paired absent/present native root, separate Claude/Codex rules |
| Curator link and frozen snapshot identity | `TestFrozenSnapshot`, matrix; real Curator install/resolve and Muse loader | Both load one native user source | Both load one native user source | Altered bytes, extra entry, link rejected; independent published snapshot checker |
| Workspace conflict | `TestMuse148WorkspaceResidual`; real loader and TUI `/rules` | Native and workspace sources counted | User then project listed; project count 1 → 0 on subsequent reload | Enumeration only; model-visible precedence remains unknown |
| Credentials/model removal | Not driven | Not measured | Not measured | Out of contract: M2 “Separate evidence” |

The matrix runs 36 calls (2 forms × 3 root states × 3 separate foreign sources
× 2 exclusion states), parser 2, workspace 4, explicit regression witness 2:
**44 vendor entry cases** in the rework replay. Without exclusion a foreign
skill visibly loads, so zero counts cannot pass vacuously. All checker reads
require explicit successful records; missing, failed, truncated or
contradictory observations cannot attest absence.

Out-of-contract rows: credentials/login/refresh and model-visible removal
(M2 “Separate evidence”); model-visible workspace precedence (M2 workspace
unknown); capture locks, export provenance, publication, destination admission,
runtime-B and materialize integration (M2 snapshot-fixture clause); runtime
consumer refusals/constants (M1/M3 ownership). No consumer qualification is
inferred from vendor tests or borrowed credential measurements.

## Validation and release effect

Reproduction commands and per-command exit codes are in the
[task handoff](../../.research/muse148/TASK-261016-i5gwaf_outcome.md).
The original interactive omission is a non-discriminating survivor: it cannot
prove exclusion where the real build already retains the skill. The checker
still rejects that skill when asked to attest zero. No exclusion or consumer
gate is weakened. Keep the §7.9 interactive release gate pending on this build;
D10 credential-mode admission remains independent. A corrected vendor build
requires a new exact-build record and replay, coordinated with M1/M3.

````

## .research/muse148/replay_test.go

SHA-256: `f4110c8c614f791d89def4fc04232f411f920ab0cda37fe63d3dd3ae9bc85cdf`

````text
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
		if !strings.Contains(line, `outcome="ready"`) && !strings.Contains(line, `outcome="empty"`) {
			return fmt.Errorf("%s failed loader record", event)
		}
		for key, n := range want {
			m := regexp.MustCompile(`(?:^| )` + key + `=([0-9]+)(?: |$)`).FindStringSubmatch(line)
			if m == nil {
				return fmt.Errorf("%s missing %s", event, key)
			}
			got, _ := strconv.Atoi(m[1])
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

// Frozen metadata binds the single root-only closure; altered bytes, extra
// files and links refuse independently of successful vendor parsing.
// Published-schema validation is separately recorded in the evidence packet.
func TestFrozenSnapshot(t *testing.T) {
	var manifest struct {
		Files []struct {
			Path string
			Size int
			SHA  string `json:"sha256"`
			Mode string
		}
		Launch string `json:"launch_sha256"`
	}
	if err := json.Unmarshal(read(t, "testdata/snapshot/manifest.json"), &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Files) != 1 || manifest.Files[0].Path != "home/config/muse/AGENTS.md" || manifest.Files[0].Mode != "0600" {
		t.Fatal("unexpected frozen closure")
	}
	if digest(read(t, "testdata/snapshot/launch.json")) != manifest.Launch {
		t.Fatal("descriptor identity differs")
	}
	entry := manifest.Files[0]
	check := func(root string) error {
		count := 0
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
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
			if filepath.ToSlash(rel) != entry.Path || !info.Mode().IsRegular() || info.Mode().Perm() != 0600 {
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
	root := t.TempDir()
	file := filepath.Join(root, entry.Path)
	frozen := read(t, "testdata/root.md")
	write(t, file, frozen)
	if err := check(root); err != nil {
		t.Fatal(err)
	}
	write(t, file, bytes.Replace(frozen, []byte("MUSE148_NATIVE_ROOT"), []byte("MUSE148_FORGED_ROOT"), 1))
	if check(root) == nil {
		t.Fatal("altered bytes admitted")
	}
	write(t, file, frozen)
	extra := filepath.Join(root, "home/extra")
	write(t, extra, []byte("extra"))
	if check(root) == nil {
		t.Fatal("extra entry admitted")
	}
	if err := os.Remove(extra); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}
	abs, err := filepath.Abs("testdata/root.md")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(abs, file); err != nil {
		t.Fatal(err)
	}
	if check(root) == nil {
		t.Fatal("linked snapshot admitted")
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

````

## .research/muse148/observations-rework.json

SHA-256: `96687427deb10dd10e9ed3a9195341a5e1c69f0c88fbc244e2fc1bfc61f33e84`

````text
{
  "version": "Muse Code 1.4.8 (1.4.8-R6256.1)",
  "binary_sha256": "5721763b4723f961709cd28f69aa049763965b46c4d4d70f883a0a90e5ddf39c",
  "binary_locator": "$MUSE148_BINARY (exact private locator retained in ignored task-local identity.json)",
  "platform": "macOS 27.0 arm64",
  "curator_source": "0d00a109ad1da77a5506747c89f24af9442eb341",
  "spec_source": "d2a942cf17e4eb017c45475442da1456d8dfbe6c",
  "replay_test_sha256": "f4110c8c614f791d89def4fc04232f411f920ab0cda37fe63d3dd3ae9bc85cdf",
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
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=1"
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
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-1fac-7e13-8bc8-a17ce55c996d\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1425d-21b0-7bc0-a2e1-6d520ae32bc2\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1425d-21b0-7bc0-a2e1-6d520ae32bc2\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-21b0-7bc0-a2e1-6d520ae32bc2\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-1fac-7e13-8bc8-a17ce55c996d\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1425d-21b0-7bc0-a2e1-6d520ae32bc2\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1425d-21b0-7bc0-a2e1-6d520ae32bc2\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-21b0-7bc0-a2e1-6d520ae32bc2\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
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
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-252c-7fb2-9330-6d62af58ec83\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1425d-26fd-7ba2-abcb-e0bb57b65c53\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1425d-26fd-7ba2-abcb-e0bb57b65c53\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-26fd-7ba2-abcb-e0bb57b65c53\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-252c-7fb2-9330-6d62af58ec83\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1425d-26fd-7ba2-abcb-e0bb57b65c53\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1425d-26fd-7ba2-abcb-e0bb57b65c53\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-26fd-7ba2-abcb-e0bb57b65c53\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
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
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-2a80-72d0-af1d-8c92c0b0a545\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1425d-2cc8-70e2-b8ff-b214aa200988\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1425d-2cc8-70e2-b8ff-b214aa200988\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-2cc8-70e2-b8ff-b214aa200988\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-2a80-72d0-af1d-8c92c0b0a545\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1425d-2cc8-70e2-b8ff-b214aa200988\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1425d-2cc8-70e2-b8ff-b214aa200988\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-2cc8-70e2-b8ff-b214aa200988\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
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
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-305f-7473-84f3-dc40dda1f18b\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1425d-3213-7af1-9dc0-ff521f13d538\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1425d-3213-7af1-9dc0-ff521f13d538\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-3213-7af1-9dc0-ff521f13d538\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-305f-7473-84f3-dc40dda1f18b\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1425d-3213-7af1-9dc0-ff521f13d538\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1425d-3213-7af1-9dc0-ff521f13d538\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-3213-7af1-9dc0-ff521f13d538\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
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
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-3632-74c2-aea1-6a1376e22c4f\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1425d-37f1-7c72-8968-f71b24b8d72f\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1425d-37f1-7c72-8968-f71b24b8d72f\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-37f1-7c72-8968-f71b24b8d72f\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-3632-74c2-aea1-6a1376e22c4f\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1425d-37f1-7c72-8968-f71b24b8d72f\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1425d-37f1-7c72-8968-f71b24b8d72f\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-37f1-7c72-8968-f71b24b8d72f\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
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
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-3b2f-7f23-b134-09f547b514f1\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1425d-3d64-7240-a2e0-c83c3cbf6c5f\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1425d-3d64-7240-a2e0-c83c3cbf6c5f\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-3d64-7240-a2e0-c83c3cbf6c5f\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-3b2f-7f23-b134-09f547b514f1\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1425d-3d64-7240-a2e0-c83c3cbf6c5f\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1425d-3d64-7240-a2e0-c83c3cbf6c5f\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-3d64-7240-a2e0-c83c3cbf6c5f\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
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
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=1",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=1"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-4157-7de0-87a7-37d85a20c0a5\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1425d-4315-70a2-a17a-eab376aef4f1\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1425d-4315-70a2-a17a-eab376aef4f1\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-4315-70a2-a17a-eab376aef4f1\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-4157-7de0-87a7-37d85a20c0a5\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1425d-4315-70a2-a17a-eab376aef4f1\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1425d-4315-70a2-a17a-eab376aef4f1\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-4315-70a2-a17a-eab376aef4f1\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
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
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=1",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-469b-79c0-91f1-d2c32369e991\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1425d-4939-7f83-afe2-2a7e7081d4ed\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1425d-4939-7f83-afe2-2a7e7081d4ed\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-4939-7f83-afe2-2a7e7081d4ed\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-469b-79c0-91f1-d2c32369e991\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1425d-4939-7f83-afe2-2a7e7081d4ed\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1425d-4939-7f83-afe2-2a7e7081d4ed\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-4939-7f83-afe2-2a7e7081d4ed\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
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
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=1",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-4c96-7480-be23-4425c5d87f22\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1425d-4e90-7e33-bbfb-501928dd57ff\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1425d-4e90-7e33-bbfb-501928dd57ff\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-4e90-7e33-bbfb-501928dd57ff\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-4c96-7480-be23-4425c5d87f22\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1425d-4e90-7e33-bbfb-501928dd57ff\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1425d-4e90-7e33-bbfb-501928dd57ff\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-4e90-7e33-bbfb-501928dd57ff\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
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
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-51fe-72f1-87db-623aa1b6e01f\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1425d-5424-70d2-94c8-f4cb3f2115ce\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1425d-5424-70d2-94c8-f4cb3f2115ce\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-5424-70d2-94c8-f4cb3f2115ce\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-51fe-72f1-87db-623aa1b6e01f\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1425d-5424-70d2-94c8-f4cb3f2115ce\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1425d-5424-70d2-94c8-f4cb3f2115ce\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-5424-70d2-94c8-f4cb3f2115ce\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
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
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=1",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=1 user_skills=1 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "muse: Including your 1 Claude Code personal skill \u2014 manage with /settings.",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-57d0-71e1-b29e-035a1505d148\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1425d-5a2e-7313-99d7-c212a0940b91\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1425d-5a2e-7313-99d7-c212a0940b91\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-5a2e-7313-99d7-c212a0940b91\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-57d0-71e1-b29e-035a1505d148\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1425d-5a2e-7313-99d7-c212a0940b91\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1425d-5a2e-7313-99d7-c212a0940b91\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-5a2e-7313-99d7-c212a0940b91\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
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
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=1",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-5de6-7581-8351-1b63242c4d1d\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1425d-5fbc-7d40-95c0-52d942588a0b\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1425d-5fbc-7d40-95c0-52d942588a0b\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-5fbc-7d40-95c0-52d942588a0b\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-5de6-7581-8351-1b63242c4d1d\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1425d-5fbc-7d40-95c0-52d942588a0b\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1425d-5fbc-7d40-95c0-52d942588a0b\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-5fbc-7d40-95c0-52d942588a0b\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
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
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-62af-7502-bf97-bde9da31da9e\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1425d-6564-7910-9a39-2115c1cb8716\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1425d-6564-7910-9a39-2115c1cb8716\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-6564-7910-9a39-2115c1cb8716\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-62af-7502-bf97-bde9da31da9e\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1425d-6564-7910-9a39-2115c1cb8716\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1425d-6564-7910-9a39-2115c1cb8716\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-6564-7910-9a39-2115c1cb8716\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
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
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-68f0-74d3-a188-10a387503d22\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1425d-6ad2-77c0-8198-0f1d6a6fe262\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1425d-6ad2-77c0-8198-0f1d6a6fe262\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-6ad2-77c0-8198-0f1d6a6fe262\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-68f0-74d3-a188-10a387503d22\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1425d-6ad2-77c0-8198-0f1d6a6fe262\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1425d-6ad2-77c0-8198-0f1d6a6fe262\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-6ad2-77c0-8198-0f1d6a6fe262\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
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
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-6e20-7461-b74e-de45bd0a73c2\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1425d-70ad-7623-994a-4274ce83a476\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1425d-70ad-7623-994a-4274ce83a476\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-70ad-7623-994a-4274ce83a476\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-6e20-7461-b74e-de45bd0a73c2\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1425d-70ad-7623-994a-4274ce83a476\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1425d-70ad-7623-994a-4274ce83a476\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-70ad-7623-994a-4274ce83a476\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
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
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-74a5-7571-b91a-534a633f3cec\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1425d-7674-74a3-9c0c-fb259200e866\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1425d-7674-74a3-9c0c-fb259200e866\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-7674-74a3-9c0c-fb259200e866\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-74a5-7571-b91a-534a633f3cec\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1425d-7674-74a3-9c0c-fb259200e866\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1425d-7674-74a3-9c0c-fb259200e866\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-7674-74a3-9c0c-fb259200e866\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
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
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-79e6-7f61-8377-a4a4fc845ebc\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1425d-7c81-7220-96a5-ed6fcbff245f\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1425d-7c81-7220-96a5-ed6fcbff245f\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-7c81-7220-96a5-ed6fcbff245f\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-79e6-7f61-8377-a4a4fc845ebc\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1425d-7c81-7220-96a5-ed6fcbff245f\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1425d-7c81-7220-96a5-ed6fcbff245f\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-7c81-7220-96a5-ed6fcbff245f\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
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
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-8000-7810-844d-c0a88cc72573\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1425d-81f1-7023-9b98-d851f8fe747b\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1425d-81f1-7023-9b98-d851f8fe747b\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-81f1-7023-9b98-d851f8fe747b\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425d-8000-7810-844d-c0a88cc72573\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1425d-81f1-7023-9b98-d851f8fe747b\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1425d-81f1-7023-9b98-d851f8fe747b\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425d-81f1-7023-9b98-d851f8fe747b\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
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
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=1",
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
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=1",
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
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=1",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=1"
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
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=1",
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
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=1"
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
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425e-523c-72a1-bbaf-f79173044a74\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1425e-555e-7bf2-83af-80868db0be3d\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1425e-555e-7bf2-83af-80868db0be3d\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425e-555e-7bf2-83af-80868db0be3d\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425e-523c-72a1-bbaf-f79173044a74\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1425e-555e-7bf2-83af-80868db0be3d\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1425e-555e-7bf2-83af-80868db0be3d\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425e-555e-7bf2-83af-80868db0be3d\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
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
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425e-58a4-7970-bf1f-20f3e67ff1b6\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1425e-5a95-7b71-9e9d-03049c215209\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1425e-5a95-7b71-9e9d-03049c215209\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425e-5a95-7b71-9e9d-03049c215209\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1425e-58a4-7970-bf1f-20f3e67ff1b6\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1425e-5a95-7b71-9e9d-03049c215209\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1425e-5a95-7b71-9e9d-03049c215209\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1425e-5a95-7b71-9e9d-03049c215209\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
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
  ]
}

````

## .research/muse148/reference-snapshot-check.py

SHA-256: `bb580313fa9ca31db92ac2a285074c671df3cd6d41e4c510d05501fb1fbc7342`

````text
import os, sys
from pathlib import Path
spec=Path(sys.argv[1]).resolve()
sys.path.insert(0,str(spec/'tools'))
from test_validate import validate
import managed_home_snapshots as snapshots
root=Path(sys.argv[2]).resolve()
manifest=validate.load_json(root/'manifest.json')
launch=validate.load_json(root/'launch.json')
content={p.relative_to(root).as_posix():(p.read_bytes(),f'{p.stat().st_mode & 0o777:04o}') for p in (root/'home').rglob('*') if p.is_file()}
if os.environ.get('MUSE148_SNAPSHOT_PLANT')=='extra': content['home/extra']=(b'extra','0600')
snapshots.check_bundle(manifest,launch,content,snapshots.validators(spec,validate.load_json),validate.ccj1_bytes,validate.validate_wire_semantics,(root/'manifest.json').read_bytes(),(root/'launch.json').read_bytes())
print('Reference check_bundle accepted exact frozen metadata and complete regular-file closure')

````

## .research/muse148/replay-rework-03.log

SHA-256: `bc56caa3b8dd5e76e500e8da9583ac1d6fa6d0deac3a147482afa3e4e184257d`

````text
=== RUN   TestObservationChecker
--- PASS: TestObservationChecker (0.00s)
=== RUN   TestMuse148OfflineBoundary
--- PASS: TestMuse148OfflineBoundary (0.28s)
=== RUN   TestFrozenSnapshot
--- PASS: TestFrozenSnapshot (0.01s)
=== RUN   TestMuse148ParserRefusals
=== RUN   TestMuse148ParserRefusals/interactive-exec-flag
=== RUN   TestMuse148ParserRefusals/interactive-string-false
--- PASS: TestMuse148ParserRefusals (0.38s)
    --- PASS: TestMuse148ParserRefusals/interactive-exec-flag (0.11s)
    --- PASS: TestMuse148ParserRefusals/interactive-string-false (0.08s)
=== RUN   TestMuse148LoaderMatrix
=== RUN   TestMuse148LoaderMatrix/exec/absent/claude-rules/exclude-false
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/claude-rules/exclude-true
    replay_test.go:619: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/codex-rules/exclude-false
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/codex-rules/exclude-true
    replay_test.go:619: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/skills/exclude-false
    replay_test.go:619: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/skills/exclude-true
    replay_test.go:619: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/claude-rules/exclude-false
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/claude-rules/exclude-true
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/codex-rules/exclude-false
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/codex-rules/exclude-true
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/skills/exclude-false
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/skills/exclude-true
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/claude-rules/exclude-false
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/claude-rules/exclude-true
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/codex-rules/exclude-false
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/codex-rules/exclude-true
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-false
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-true
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/claude-rules/exclude-false
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/claude-rules/exclude-true
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/codex-rules/exclude-false
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/codex-rules/exclude-true
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/skills/exclude-false
    replay_test.go:619: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/skills/exclude-true
    replay_test.go:619: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/claude-rules/exclude-false
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/claude-rules/exclude-true
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/codex-rules/exclude-false
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/codex-rules/exclude-true
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/skills/exclude-false
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/skills/exclude-true
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/claude-rules/exclude-false
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/claude-rules/exclude-true
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/codex-rules/exclude-false
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/codex-rules/exclude-true
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-false
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-true
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
--- PASS: TestMuse148LoaderMatrix (72.57s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/claude-rules/exclude-false (1.42s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/claude-rules/exclude-true (1.36s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/codex-rules/exclude-false (1.50s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/codex-rules/exclude-true (1.46s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/skills/exclude-false (1.31s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/skills/exclude-true (1.58s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/claude-rules/exclude-false (1.33s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/claude-rules/exclude-true (1.55s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/codex-rules/exclude-false (1.38s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/codex-rules/exclude-true (1.48s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/skills/exclude-false (1.58s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/skills/exclude-true (1.22s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/claude-rules/exclude-false (1.61s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/claude-rules/exclude-true (1.33s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/codex-rules/exclude-false (1.67s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/codex-rules/exclude-true (1.33s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-false (1.56s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-true (1.42s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/claude-rules/exclude-false (2.64s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/claude-rules/exclude-true (2.73s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/codex-rules/exclude-false (2.79s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/codex-rules/exclude-true (2.84s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/skills/exclude-false (2.57s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/skills/exclude-true (2.65s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/claude-rules/exclude-false (2.44s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/claude-rules/exclude-true (2.34s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/codex-rules/exclude-false (2.21s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/codex-rules/exclude-true (2.59s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/skills/exclude-false (2.22s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/skills/exclude-true (2.42s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/claude-rules/exclude-false (2.21s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/claude-rules/exclude-true (2.43s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/codex-rules/exclude-false (2.67s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/codex-rules/exclude-true (2.53s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-false (2.42s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-true (2.43s)
=== RUN   TestMuse148InteractiveExclusionRegression
=== RUN   TestMuse148InteractiveExclusionRegression/linked
    replay_test.go:649: vendor exit=0; required interactive exclusion FAILED (user_skills=1); see docs/qualification/muse-1.4.8-R6256.1.md
=== RUN   TestMuse148InteractiveExclusionRegression/snapshot
    replay_test.go:649: vendor exit=0; required interactive exclusion FAILED (user_skills=1); see docs/qualification/muse-1.4.8-R6256.1.md
--- PASS: TestMuse148InteractiveExclusionRegression (6.01s)
    --- PASS: TestMuse148InteractiveExclusionRegression/linked (2.29s)
    --- PASS: TestMuse148InteractiveExclusionRegression/snapshot (2.50s)
=== RUN   TestMuse148WorkspaceResidual
=== RUN   TestMuse148WorkspaceResidual/exec/linked
=== RUN   TestMuse148WorkspaceResidual/exec/snapshot
=== RUN   TestMuse148WorkspaceResidual/interactive/linked
=== RUN   TestMuse148WorkspaceResidual/interactive/snapshot
--- PASS: TestMuse148WorkspaceResidual (9.41s)
    --- PASS: TestMuse148WorkspaceResidual/exec/linked (1.66s)
    --- PASS: TestMuse148WorkspaceResidual/exec/snapshot (1.54s)
    --- PASS: TestMuse148WorkspaceResidual/interactive/linked (2.41s)
    --- PASS: TestMuse148WorkspaceResidual/interactive/snapshot (2.55s)
PASS
ok  	github.com/relux-works/curator/.research/muse148	88.874s

````

## .research/muse148/plant-exec-rework-03.log

SHA-256: `c04e167a029aedd6ddce7aa03d4f2e295647efbc2dde79d25fc51537e1224559`

````text
=== RUN   TestMuse148LoaderMatrix
=== RUN   TestMuse148LoaderMatrix/exec/linked/skills/exclude-false
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/skills/exclude-true
    replay_test.go:617: skills.catalog_load observed user_skills=1; want 0
--- FAIL: TestMuse148LoaderMatrix (2.99s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/skills/exclude-false (1.14s)
    --- FAIL: TestMuse148LoaderMatrix/exec/linked/skills/exclude-true (1.06s)
FAIL
FAIL	github.com/relux-works/curator/.research/muse148	3.279s
FAIL

````

## .research/muse148/plant-interactive-rework-03.log

SHA-256: `49026dd169703ee2d4c3a447495b9dc2a8631f7a40fb26a9423e785ed68b8707`

````text
=== RUN   TestMuse148LoaderMatrix
=== RUN   TestMuse148LoaderMatrix/interactive/linked/skills/exclude-false
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/skills/exclude-true
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
--- PASS: TestMuse148LoaderMatrix (5.65s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/skills/exclude-false (2.39s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/skills/exclude-true (2.17s)
PASS
ok  	github.com/relux-works/curator/.research/muse148	5.843s

````

## .research/muse148/checker-mutant-rework-03.log

SHA-256: `06fa1d97910355e0407341bc2bee678bcf6367f4d7a1badc0d5a1e7770af8217`

````text
=== RUN   TestObservationChecker
    replay_test.go:94: checker admitted one foreign skill
--- FAIL: TestObservationChecker (0.00s)
FAIL
FAIL	github.com/relux-works/curator/.temp/TASK-261016-i5gwaf/checker-mutant-rework-03	0.295s
FAIL

````

## .research/muse148/snapshot-mutant-rework-04.log

SHA-256: `afdf602ea943adaa23c58fffb54f847567db02d082bb7925b8402e7cc6e897bf`

````text
=== RUN   TestFrozenSnapshot
    replay_test.go:450: extra entry admitted
--- FAIL: TestFrozenSnapshot (0.01s)
FAIL
FAIL	github.com/relux-works/curator/.temp/TASK-261016-i5gwaf/snapshot-mutant-rework-03	0.178s
FAIL

````

## .research/muse148/offline-mutant-rework-05.log

SHA-256: `3f845271ef5ef557a527e9c490335f51bbb4d5e088dfcbeec7807e2bda32e01f`

````text
=== RUN   TestMuse148OfflineBoundary
    replay_test.go:349: network deny exit=0: MUSE148_NETWORK_CONTROL
--- FAIL: TestMuse148OfflineBoundary (0.22s)
FAIL
FAIL	github.com/relux-works/curator/.temp/TASK-261016-i5gwaf/offline-mutant-rework-03	0.509s
FAIL

````

## .research/muse148/snapshot-reference-rework-01.log

SHA-256: `c3aa82a13039aeb5fe05903e5be0e656aff72a33fba2c5015cbec05d3efe22a9`

````text
Reference check_bundle accepted exact frozen metadata and complete regular-file closure

````

## .research/muse148/snapshot-plant-rework-01.log

SHA-256: `3ec7492ef13d88dfee7a6e581c2dd02ef7e9c212487d4f1837550b5baeada140`

````text
Traceback (most recent call last):
  File "$PRIVATE_PATH", line 12, in <module>
    snapshots.check_bundle(manifest,launch,content,snapshots.validators(spec,validate.load_json),validate.ccj1_bytes,validate.validate_wire_semantics,(root/'manifest.json').read_bytes(),(root/'launch.json').read_bytes())
    ~~~~~~~~~~~~~~~~~~~~~~^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
  File "$TASK_ROOT/spec-main/tools/managed_home_snapshots.py", line 81, in check_bundle
    raise ValueError("entry-set")
ValueError: entry-set

````

## .research/muse148/default-status-isolated-01.log

SHA-256: `37822eeca5ee6ee72bd49d6ab13b46d0047ef5693e7dc5ce00bc52491e67a796`

````text
=== RUN   TestDraftStatusReportsUpToDateThroughCLI
--- PASS: TestDraftStatusReportsUpToDateThroughCLI (0.90s)
=== RUN   TestDraftStatusNeedsInstallAfterRefreshWithoutInstall
--- PASS: TestDraftStatusNeedsInstallAfterRefreshWithoutInstall (1.66s)
=== RUN   TestDraftStatusLiveMutationWithoutRefreshStaysCurrent
--- PASS: TestDraftStatusLiveMutationWithoutRefreshStaysCurrent (0.67s)
=== RUN   TestEnvStatusMatrix
--- PASS: TestEnvStatusMatrix (11.81s)
=== RUN   TestEnvStatusRegistryBoundaryPostureAndCheck
--- PASS: TestEnvStatusRegistryBoundaryPostureAndCheck (16.65s)
PASS
ok  	github.com/relux-works/curator/cmd/curator	32.188s

````

## .research/muse148/testdata/root.md

SHA-256: `8c9af1ca52b75aaa1e67af6365f1e39dfcc3c56a274d3571270079ef669ac9d0`

````text
<!--
curator-root-context-v2
root: muse148-probe 1.0.0 state sha256:730819a169c050ab2dbeabe75f0fea93395e67d14ca8d7d240cbebc1a326e968
member: muse148-probe 1.0.0 state sha256:730819a169c050ab2dbeabe75f0fea93395e67d14ca8d7d240cbebc1a326e968 weight 0
precedence: winner=higher-weight placement=winner-last
lock: sha256:458646149b0ef02eb037c74cb6fe38ca74b325c4be05bfbe5d9f8e302fb5f6fc
generated: Curator Protocol environments revision 1 (https://github.com/relux-works/curator-spec)
notice: generated file; direct edits are unsupported and are detected as drift; update the source profile repository or its composed profiles instead
-->

---

## Context: muse148-probe 1.0.0

MUSE148_NATIVE_ROOT

````

## .research/muse148/testdata/snapshot/home/config/muse/AGENTS.md

SHA-256: `8c9af1ca52b75aaa1e67af6365f1e39dfcc3c56a274d3571270079ef669ac9d0`

````text
<!--
curator-root-context-v2
root: muse148-probe 1.0.0 state sha256:730819a169c050ab2dbeabe75f0fea93395e67d14ca8d7d240cbebc1a326e968
member: muse148-probe 1.0.0 state sha256:730819a169c050ab2dbeabe75f0fea93395e67d14ca8d7d240cbebc1a326e968 weight 0
precedence: winner=higher-weight placement=winner-last
lock: sha256:458646149b0ef02eb037c74cb6fe38ca74b325c4be05bfbe5d9f8e302fb5f6fc
generated: Curator Protocol environments revision 1 (https://github.com/relux-works/curator-spec)
notice: generated file; direct edits are unsupported and are detected as drift; update the source profile repository or its composed profiles instead
-->

---

## Context: muse148-probe 1.0.0

MUSE148_NATIVE_ROOT

````

## .research/muse148/testdata/snapshot/launch.json

SHA-256: `2991bc9c97904d77aef70bb5c3ea7a7044ff462382310ae613bbd31f461fd5f1`

````text
{"descriptor":"snapshot-launch-v1","env":{"XDG_CACHE_HOME":"home/cache","XDG_CONFIG_HOME":"home/config","XDG_DATA_HOME":"home/data","XDG_STATE_HOME":"home/state"},"environment":"muse","permissions":{"locked":false,"mode":"native","source":"default"},"precedence":{"placement":"winner-last","winner":"higher-weight"},"profile":{"lock_sha256":"458646149b0ef02eb037c74cb6fe38ca74b325c4be05bfbe5d9f8e302fb5f6fc","name":"muse148-probe"},"root_context":{"exclusive":true}}

````

## .research/muse148/testdata/snapshot/manifest.json

SHA-256: `924378006ba6e6e7db9d0277375456a958c14cbaff03d1ebd63dfa6e4b7ceac8`

````text
{"adapter":"muse","curator_version":"0.15.0-rc.6.0.20261015232206-0d00a109ad1d+dirty","files":[{"mode":"0600","path":"home/config/muse/AGENTS.md","sha256":"8c9af1ca52b75aaa1e67af6365f1e39dfcc3c56a274d3571270079ef669ac9d0","size":692}],"fragment_sha256":"9eda0ac5bed34653262ee49b75b8e0bf7a74a29979ae9dfc272319b3333dfce9","launch_sha256":"2991bc9c97904d77aef70bb5c3ea7a7044ff462382310ae613bbd31f461fd5f1","profile":{"lock_sha256":"458646149b0ef02eb037c74cb6fe38ca74b325c4be05bfbe5d9f8e302fb5f6fc","name":"muse148-probe"},"snapshot":"managed-home-snapshot-v1"}

````

## .research/muse148/testdata/source/agent-context.json

SHA-256: `a89fab21a84ea4e8f1b599f6b46b89f5df3ea075c8416c9c92cc9ad7b376c1b7`

````text
{"schema_version":1,"name":"muse148-probe","version":"1.0.0","context":{"modules":[{"path":"root.md"}]}}

````

## .research/muse148/testdata/source/context/root.md

SHA-256: `ae4a2c75505cec507e6957217cea7f0511564b96ee6dddacd9f693dbc5074ab0`

````text
MUSE148_NATIVE_ROOT

````

## .research/muse148/mutants/README.md

SHA-256: `2c9cb58b6452c7babbb82f0a8e2bdf743acf7ff7526072609f36d0ea4767d1c5`

````text
# Narrowing controls

These patches apply to the recorded replay source, not runtime code. They
retain the gate and weaken one boundary. Apply only in disposable task-local
copies. Never patch the candidate or run the vendor matrix with the offline
safety policy mutant.

```sh
cp -R .research/muse148 .temp/TASK-261016-i5gwaf/mutant-new
git apply --directory=.temp/TASK-261016-i5gwaf/mutant-new .research/muse148/mutants/checker-one-skill.patch
go test ./.temp/TASK-261016-i5gwaf/mutant-new -run '^TestObservationChecker$' -count=1 -v
```

Use a new copy for each control. `checker-one-skill.patch` must fail in
`TestObservationChecker` with `checker admitted one foreign skill`.
`snapshot-extra-entry.patch` must fail in `TestFrozenSnapshot` with
`extra entry admitted`; copying `testdata` is necessary for this control.
`offline-loopback.patch` must fail in `TestMuse148OfflineBoundary` with
`network deny exit=0: MUSE148_NETWORK_CONTROL`. That test needs the documented
`MUSE148_BINARY`, `MUSE148_CURATOR` and a fresh `MUSE148_RUN_ROOT`; it starts
only the synthetic loopback endpoint and curl, never Muse/model inference.
Initial missing-fixture/unsupported-policy attempts do not count as kills.

All three corrected controls exited 1 for the intended assertion. The exec
omission plant uses `MUSE148_OMIT_EXCLUSION=exec` with the matrix test; the
interactive omission is an explicitly reported survivor.

````

## .research/muse148/mutants/checker-one-skill.patch

SHA-256: `a4f7a0886957d06e4b571a863604664a184c185e0f5254d1671811b42f1b0f07`

````text
--- a/replay_test.go
+++ b/replay_test.go
@@ -63,7 +63,7 @@
 				return fmt.Errorf("%s missing %s", event, key)
 			}
 			got, _ := strconv.Atoi(m[1])
-			if got != n {
+			if got != n && !(event == "skills.catalog_load" && key == "user_skills" && n == 0 && got == 1) {
 				return fmt.Errorf("%s observed %s=%d; want %d", event, key, got, n)
 			}
 		}

````

## .research/muse148/mutants/offline-loopback.patch

SHA-256: `7653706c7349f28b0cc0ee7daac3fe61daedce9c654f5cb73eaf49190aa515f4`

````text
--- a/replay_test.go
+++ b/replay_test.go
@@ -134,6 +134,7 @@
 	for _, rel := range []string{".claude", ".codex", ".config", ".local", ".muse", "Library"} {
 		sb += "(deny file-read* file-write* (subpath " + strconv.Quote(filepath.Join(native, rel)) + "))\n"
 	}
+	sb += "(allow network-outbound (remote tcp \"localhost:*\"))\n"
 	write(t, p.sandbox, []byte(sb))
 	return p
 }

````

## .research/muse148/mutants/snapshot-extra-entry.patch

SHA-256: `411bcafc9d9b306609c990846732e9838477a17f16f5bd9ab6989756320e720d`

````text
--- a/replay_test.go
+++ b/replay_test.go
@@ -402,6 +402,7 @@
 			if d.IsDir() {
 				return nil
 			}
+			if rel, _ := filepath.Rel(root, path); filepath.ToSlash(rel) == "home/extra" { return nil }
 			count++
 			info, err := d.Info()
 			if err != nil {

````
