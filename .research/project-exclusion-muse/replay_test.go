// Package museexclusion records native discovery observations, not adapter admission.
package museexclusion

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
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

const binarySHA = "6c04ce1719844685153ef6f04b98113cb61b942d97f0a8b0bfb3932c343984d9"

func read(t *testing.T, p string) []byte {
	t.Helper()
	b, e := os.ReadFile(p)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func write(t *testing.T, p string, b []byte) {
	t.Helper()
	if e := os.MkdirAll(filepath.Dir(p), 0700); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(p, b, 0600); e != nil {
		t.Fatal(e)
	}
}
func jsonBytes(t *testing.T, v any) []byte {
	t.Helper()
	b, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	return append(b, '\n')
}
func quote(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'" }

type harness struct{ binary, base, home string }
type probe struct {
	h                   *harness
	root, work, sandbox string
	seq                 int
}

func setup(t *testing.T) *harness {
	t.Helper()
	binary := os.Getenv("MUSE_EXCL_BINARY")
	if binary == "" {
		t.Skip("requires digest-pinned MUSE_EXCL_BINARY and task-local MUSE_EXCL_RUN_ROOT; skipped is unmeasured")
	}
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		t.Fatal("qualification requires darwin/arm64")
	}
	sum := sha256.Sum256(read(t, binary))
	if hex.EncodeToString(sum[:]) != binarySHA {
		t.Fatal("binary digest mismatch")
	}
	base := os.Getenv("MUSE_EXCL_RUN_ROOT")
	if !filepath.IsAbs(base) || !strings.Contains(filepath.Clean(base), "/.temp/") {
		t.Fatal("explicit absolute .temp run root required")
	}
	home, e := os.UserHomeDir()
	if e != nil {
		t.Fatal(e)
	}
	return &harness{binary: binary, base: base, home: home}
}
func (h *harness) fixture(t *testing.T) *probe {
	t.Helper()
	root := filepath.Join(h.base, strings.ReplaceAll(t.Name(), "/", "_"))
	if _, e := os.Stat(root); !os.IsNotExist(e) {
		t.Fatal("fresh case directory required")
	}
	p := &probe{h: h, root: root, work: filepath.Join(root, "ancestor/project"), sandbox: filepath.Join(root, "offline.sb")}
	for _, d := range []string{"home", "config/muse", "data", "state", "cache", "tmp", "ancestor/project"} {
		if e := os.MkdirAll(filepath.Join(root, d), 0700); e != nil {
			t.Fatal(e)
		}
	}
	sb := "(version 1)\n(allow default)\n(deny network*)\n(deny mach-lookup (global-name-regex #\"^com\\.apple\\.(security|Security)\"))\n(deny file-read* file-write* (regex #\"/Keychains(/|$)\"))\n"
	for _, d := range []string{".claude", ".codex", ".config", ".local", ".muse", ".curator", ".ssh", ".gnupg", ".aws", ".azure", ".netrc", ".npmrc", "Library"} {
		sb += "(deny file-read* file-write* (subpath " + strconv.Quote(filepath.Join(h.home, d)) + "))\n"
	}
	if os.Getenv("MUSE_EXCL_MUTANT") == "network" {
		sb = strings.Replace(sb, "(deny network*)\n", "", 1)
	}
	if os.Getenv("MUSE_EXCL_MUTANT") == "keychains" {
		sb = strings.Replace(sb, "(deny file-read* file-write* (regex #\"/Keychains(/|$)\"))\n", "", 1)
	}
	write(t, p.sandbox, []byte(sb))
	// A real checkout with a .git FILE; metadata stays outside the checkout.
	out, code := p.command(t, "/usr/bin/git", "init", "--separate-git-dir", filepath.Join(root, "metadata"), p.work)
	if code != 0 {
		t.Fatalf("fixture git init exit=%d: %s", code, out)
	}
	write(t, filepath.Join(root, "config/muse/AGENTS.md"), read(t, "testdata/AGENTS.md"))
	return p
}
func (p *probe) env() []string {
	return []string{"HOME=" + filepath.Join(p.root, "home"), "XDG_CONFIG_HOME=" + filepath.Join(p.root, "config"), "XDG_DATA_HOME=" + filepath.Join(p.root, "data"), "XDG_STATE_HOME=" + filepath.Join(p.root, "state"), "XDG_CACHE_HOME=" + filepath.Join(p.root, "cache"), "TMPDIR=" + filepath.Join(p.root, "tmp"), "PATH=/usr/bin:/bin", "TERM=xterm-256color", "MUSE_NO_AUTO_UPDATE=1", "TBH_CREDENTIAL_BACKEND=file"}
}
func (p *probe) command(t *testing.T, name string, args ...string) ([]byte, int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = p.work
	cmd.Env = p.env()
	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b
	if e := cmd.Start(); e != nil {
		t.Fatal(e)
	}
	pid := cmd.Process.Pid
	err := cmd.Wait()
	code := 0
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			code = e.ExitCode()
		} else {
			t.Fatal(err)
		}
	}
	p.seq++
	write(t, filepath.Join(p.root, fmt.Sprintf("process-%03d.json", p.seq)), jsonBytes(t, map[string]any{"pid": pid, "command": filepath.Base(name), "exit": code, "timeout": ctx.Err() != nil}))
	if ctx.Err() != nil {
		t.Fatalf("bounded command timed out: %s", filepath.Base(name))
	}
	return b.Bytes(), code
}
func (p *probe) args(args ...string) []string {
	a := []string{"-f", p.sandbox, "/usr/bin/env", "-i"}
	a = append(a, p.env()...)
	a = append(a, p.h.binary)
	return append(a, args...)
}
func (p *probe) vendor(t *testing.T, args ...string) ([]byte, int) {
	t.Helper()
	out, code := p.command(t, "/usr/bin/sandbox-exec", p.args(args...)...)
	write(t, filepath.Join(p.root, fmt.Sprintf("vendor-%03d.log", p.seq)), out)
	return out, code
}
func (p *probe) exec(t *testing.T, trusted bool, extra ...string) string {
	t.Helper()
	args := []string{"exec", "--json", "--provider", "echo", "--disable-reminders", "--approval-judge", "off"}
	if trusted || os.Getenv("MUSE_EXCL_MUTANT") == "trust" {
		args = append(args, "--trust-workspace")
	}
	args = append(args, extra...)
	args = append(args, "MUSE_EXCL_OFFLINE")
	out, code := p.vendor(t, args...)
	if code != 0 || !bytes.Contains(out, []byte(`"text":"echo: MUSE_EXCL_OFFLINE"`)) {
		t.Fatalf("echo exit=%d: %s", code, out)
	}
	return p.startup(t)
}
func (p *probe) startup(t *testing.T) string {
	t.Helper()
	var logs []string
	err := filepath.WalkDir(filepath.Join(p.root, "data/muse/sessions"), func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.IsDir() && strings.HasPrefix(d.Name(), "cli-") && strings.HasSuffix(d.Name(), ".log") {
			logs = append(logs, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// Each case executes once unless it explicitly tests resume; select the latest
	// process log there. Bootstrap duplicates and post-security-mode runtime reloads are excluded
	// from this explicitly startup-only observation, not counted as exclusion proof.
	var latest string
	var stamp time.Time
	for _, f := range logs {
		s, e := os.Stat(f)
		if e != nil {
			t.Fatal(e)
		}
		if latest == "" || s.ModTime().After(stamp) {
			latest = f
			stamp = s.ModTime()
		}
	}
	if latest == "" {
		t.Fatal("no native session log")
	}
	raw := string(read(t, latest))
	boundary := strings.Index(raw, `event="security_mode.resolve"`)
	if boundary < 0 {
		t.Fatal("missing startup boundary")
	}
	raw = raw[:boundary]
	var selected []string
	for _, line := range strings.Split(raw, "\n") {
		if strings.Contains(line, `event="`) {
			i := strings.Index(line, `event="`)
			selected = append(selected, line[i:])
		}
	}
	summary := strings.Join(selected, "\n") + "\n"
	write(t, filepath.Join(p.root, fmt.Sprintf("startup-%03d.txt", p.seq)), []byte(summary))
	return summary
}

var fieldRE = regexp.MustCompile(`(?:^|[ \t])([A-Za-z_][A-Za-z0-9_]*)=("(?:\\.|[^"\\])*"|[^ \t]+)`)

func check(log, event, outcome string, want map[string]int) error {
	var lines []string
	for _, l := range strings.Split(log, "\n") {
		if strings.Contains(l, `event="`+event+`"`) {
			lines = append(lines, l)
		}
	}
	if len(lines) != 1 {
		return fmt.Errorf("%s: expected one startup observation, got %d", event, len(lines))
	}
	fields := map[string][]string{}
	for _, m := range fieldRE.FindAllStringSubmatchIndex(lines[0], -1) {
		if m[1] < len(lines[0]) && lines[0][m[1]] != ' ' && lines[0][m[1]] != '\t' {
			return fmt.Errorf("malformed field")
		}
		k, v := lines[0][m[2]:m[3]], lines[0][m[4]:m[5]]
		fields[k] = append(fields[k], v)
	}
	for _, k := range []string{"event", "outcome"} {
		if len(fields[k]) != 1 {
			return fmt.Errorf("missing/duplicate %s", k)
		}
	}
	if fields["event"][0] != strconv.Quote(event) {
		return fmt.Errorf("contradictory event")
	}
	if fields["outcome"][0] != strconv.Quote(outcome) {
		return fmt.Errorf("%s outcome: %v; want %s", event, fields["outcome"], outcome)
	}
	for k, n := range want {
		if len(fields[k]) != 1 || fields[k][0] != strconv.Itoa(n) {
			return fmt.Errorf("%s %s: %v; want %d", event, k, fields[k], n)
		}
	}
	return nil
}
func require(t *testing.T, log, event, outcome string, want map[string]int) {
	t.Helper()
	if e := check(log, event, outcome, want); e != nil {
		t.Fatal(e)
	}
	t.Logf("%s outcome=%s %v", event, outcome, want)
}
func rules(t *testing.T, log string, project int) {
	t.Helper()
	require(t, log, "rules.context_load", "ready", map[string]int{"project_sources": project, "user_sources": 1, "foreign_sources": 0})
}
func (p *probe) settings(t *testing.T, cfg map[string]any) {
	t.Helper()
	cfg["schema_version"] = 1
	write(t, filepath.Join(p.root, "config/muse/settings.json"), jsonBytes(t, cfg))
}
func (p *probe) trust(t *testing.T, on bool) {
	t.Helper()
	decision := "untrusted"
	if on {
		decision = "trusted"
	}
	write(t, filepath.Join(p.root, "config/muse/trust.json"), jsonBytes(t, map[string]any{"schema_version": 1, "projects": map[string]any{p.work: map[string]any{"decision": decision}}}))
}
func (p *probe) skill(t *testing.T, path string) {
	t.Helper()
	write(t, path, []byte(strings.Replace(string(read(t, "testdata/SKILL.md")), "name: exclusion-probe", "name: "+filepath.Base(filepath.Dir(path)), 1)))
}
func (p *probe) hooks(t *testing.T, marker string) map[string]any {
	t.Helper()
	script := filepath.Join(p.root, "hook.sh")
	write(t, script, read(t, "testdata/hook.sh"))
	return map[string]any{"UserPromptSubmit": []any{map[string]any{"hooks": []any{map[string]any{"type": "command", "command": "/bin/sh " + quote(script) + " " + quote(filepath.Join(p.root, marker))}}}}}
}

// The checker admits an explicit zero and its positive one; missing, failed,
// duplicate and contradictory records must never become evidence of absence.
func TestChecker(t *testing.T) {
	setup(t)
	clean := `event="rules.context_load" outcome="ready" project_sources=0`
	if e := check(clean, "rules.context_load", "ready", map[string]int{"project_sources": 0}); e != nil {
		t.Fatal(e)
	}
	for _, bad := range []string{"", clean + "\n" + clean, strings.Replace(clean, "ready", "failed", 1), strings.Replace(clean, "=0", "=1", 1), clean + " project_sources=0", clean + ` outcome="ready"`, clean + "junk", strings.Replace(clean, `event="rules.context_load"`, `event="other" extraevent="rules.context_load"`, 1)} {
		if check(bad, "rules.context_load", "ready", map[string]int{"project_sources": 0}) == nil {
			t.Fatalf("checker accepted %q", bad)
		}
	}
	if e := check(strings.Replace(clean, "=0", "=1", 1), "rules.context_load", "ready", map[string]int{"project_sources": 1}); e != nil {
		t.Fatal(e)
	}
}

// P1 measures each filename separately at checkout root and an ancestor inside the checkout. Fresh
// untrusted homes exclude it while the approved home root still loads.
func TestP1RootAndAncestors(t *testing.T) {
	h := setup(t)
	for _, name := range []string{"AGENTS.md", "CLAUDE.md", ".agents/AGENTS.md", ".claude/CLAUDE.md"} {
		for _, where := range []string{"root", "ancestor"} {
			for _, trust := range []bool{true, false} {
				t.Run(fmt.Sprintf("%s/%s/trusted-%t", strings.ReplaceAll(name, "/", "_"), where, trust), func(t *testing.T) {
					p := h.fixture(t)
					base := p.work
					if where == "ancestor" {
						p.work = filepath.Join(p.work, "sub")
						if e := os.MkdirAll(p.work, 0700); e != nil {
							t.Fatal(e)
						}
					}
					write(t, filepath.Join(base, name), []byte("MUSE_EXCL_PROJECT_ROOT\n"))
					log := p.exec(t, trust)
					n := 0
					if trust {
						n = 1
					}
					rules(t, log, n)
				})
			}
		}
	}
}

// P2 measures a separate launch in a nested directory, not a lazy tool read in
// an existing session. The latter remains explicitly unmeasured in the record.
func TestP2NestedStartup(t *testing.T) {
	h := setup(t)
	for _, trust := range []bool{true, false} {
		t.Run(fmt.Sprint(trust), func(t *testing.T) {
			p := h.fixture(t)
			write(t, filepath.Join(p.work, "AGENTS.md"), []byte("ROOT\n"))
			p.work = filepath.Join(p.work, "sub")
			write(t, filepath.Join(p.work, "CLAUDE.md"), []byte("NESTED\n"))
			log := p.exec(t, trust)
			n := 0
			if trust {
				n = 2
			}
			rules(t, log, n)
		})
	}
}

// P3 pairs native/compat project roots with a home skill. The counts establish
// catalog discovery only; they do not claim that an LLM invoked a skill.
func TestP3Skills(t *testing.T) {
	h := setup(t)
	for _, source := range []string{".agents", ".claude", ".codex"} {
		for _, trust := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/%t", source, trust), func(t *testing.T) {
				p := h.fixture(t)
				p.skill(t, filepath.Join(p.work, source, "skills/project/SKILL.md"))
				p.skill(t, filepath.Join(p.root, "config/muse/skills/approved/SKILL.md"))
				log := p.exec(t, trust)
				rules(t, log, 0)
				n := 0
				if trust {
					n = 1
				}
				require(t, log, "skills.catalog_load", "ready", map[string]int{"project_skills": n, "user_skills": 1, "invalid": 0})
			})
		}
	}
}

// P4 hooks prove project execution when trusted, inactivity when untrusted,
// and actual approved user-hook execution beside the negative case.
func TestP4Hooks(t *testing.T) {
	h := setup(t)
	for _, trust := range []bool{true, false} {
		t.Run(fmt.Sprint(trust), func(t *testing.T) {
			p := h.fixture(t)
			write(t, filepath.Join(p.work, ".muse/hooks.json"), jsonBytes(t, map[string]any{"hooks": p.hooks(t, "project-hook")}))
			p.settings(t, map[string]any{"hooks": p.hooks(t, "approved-hook")})
			log := p.exec(t, trust)
			rules(t, log, 0)
			outcome := "partial"
			runnable, inactive := 1, 1
			if trust {
				outcome = "ready"
				runnable, inactive = 2, 0
			}
			require(t, log, "hook.admission", outcome, map[string]int{"handlers": 2, "runnable": runnable, "inactive": inactive, "diagnostics": 0})
			if string(read(t, filepath.Join(p.root, "approved-hook"))) != "MUSE_EXCL_HOOK\n" {
				t.Fatal("approved hook did not execute")
			}
			_, e := os.Stat(filepath.Join(p.root, "project-hook"))
			if trust && e != nil {
				t.Fatal(e)
			}
			if !trust && !os.IsNotExist(e) {
				t.Fatal("unapproved project hook executed")
			}
		})
	}
}
func plugin(t *testing.T, root string) {
	t.Helper()
	write(t, filepath.Join(root, "commands/probe.md"), []byte("---\ndescription: Synthetic command control.\n---\nMUSE_EXCL_PLUGIN\n"))
	write(t, filepath.Join(root, ".muse-plugin/plugin.json"), jsonBytes(t, map[string]any{"schemaVersion": 1, "name": "exclusion-probe", "displayName": "Exclusion Probe", "version": "0.1.0", "description": "Synthetic offline control.", "compat": map[string]any{"source": "native", "manifestDir": ".muse-plugin"}, "capabilities": map[string]any{"skills": []any{}, "hooks": []any{}, "mcpServers": []any{}, "reminders": []any{}, "commands": []any{map[string]any{"id": "probe", "path": "commands/probe.md", "enabledDefault": true}}}}))
}

// P4 plugin controls distinguish an inert repository bundle, a rejected explicit
// project install, and an approved home install. Installation is not auto-discovery.
func TestP4Plugins(t *testing.T) {
	h := setup(t)
	for _, mode := range []string{"raw", "project-install", "trusted-project-install", "approved-install"} {
		t.Run(mode, func(t *testing.T) {
			p := h.fixture(t)
			source := filepath.Join(p.work, "plugin")
			if mode == "approved-install" {
				source = filepath.Join(p.root, "home/approved-plugin")
			}
			plugin(t, source)
			if mode == "trusted-project-install" {
				p.trust(t, true)
			}
			if mode != "raw" {
				scope := "project"
				if mode == "approved-install" {
					scope = "user"
				}
				out, code := p.vendor(t, "plugins", "install", source, "--scope", scope, "--json")
				if mode == "project-install" {
					if code != 1 || !bytes.Contains(out, []byte("workspace is untrusted")) {
						t.Fatalf("project install rejection exit=%d: %s", code, out)
					}
				} else if code != 0 {
					t.Fatalf("approved install exit=%d: %s", code, out)
				}
			}
			log := p.exec(t, mode == "trusted-project-install")
			rules(t, log, 0)
			n := 0
			if mode == "approved-install" || mode == "trusted-project-install" {
				n = 1
			}
			require(t, log, "plugin_capability_snapshot.compose", "ready", map[string]int{"installed_plugins": n, "commands": 1 + n, "diagnostics": 0})
		})
	}
}

// P5 observes resolved server configuration, including an explicitly empty home
// set. It does not assert a server handshake or tool invocation under echo.
func TestP5MCP(t *testing.T) {
	h := setup(t)
	for _, empty := range []bool{true, false} {
		for _, trust := range []bool{true, false} {
			t.Run(fmt.Sprintf("empty-%t/trusted-%t", empty, trust), func(t *testing.T) {
				p := h.fixture(t)
				project := map[string]any{}
				if !empty {
					project["project-probe"] = map[string]any{"command": "/usr/bin/true", "args": []string{}}
				}
				write(t, filepath.Join(p.work, ".mcp.json"), jsonBytes(t, map[string]any{"mcpServers": project}))
				p.settings(t, map[string]any{"mcpServers": map[string]any{"approved-probe": map[string]any{"command": "/usr/bin/true", "args": []string{}}}})
				log := p.exec(t, trust)
				rules(t, log, 0)
				sources, servers := 0, 1
				if trust {
					sources = 1
					if !empty {
						servers++
					}
				}
				require(t, log, "mcp.config.resolve", "ready", map[string]int{"trusted_project_source_count": sources, "user_settings_source_count": 1, "server_count": servers, "server_failure_count": 0})
			})
		}
	}
	t.Run("empty-home-with-nonempty-project", func(t *testing.T) {
		p := h.fixture(t)
		write(t, filepath.Join(p.work, ".mcp.json"), []byte(`{"mcpServers":{"project-probe":{"command":"/usr/bin/true"}}}`))
		p.settings(t, map[string]any{"mcpServers": map[string]any{}})
		log := p.exec(t, true)
		rules(t, log, 0)
		require(t, log, "mcp.config.resolve", "ready", map[string]int{"trusted_project_source_count": 1, "server_count": 1})
		t.Log("empty home MCP is NOT an exclusive set in a trusted checkout")
	})
}

// P6 tests named candidate settings/permission locations. A poison user copy
// refuses startup; identical project bytes do not, even with workspace trust.
// This proves non-loading of these candidates, not absence of every dialect.
func TestP6SettingsCandidates(t *testing.T) {
	h := setup(t)
	for _, name := range []string{".muse/settings.json", ".claude/settings.json", ".claude/settings.local.json"} {
		t.Run(strings.ReplaceAll(name, "/", "_"), func(t *testing.T) {
			p := h.fixture(t)
			write(t, filepath.Join(p.work, name), []byte("INVALID_SETTINGS_CONTROL\n"))
			p.settings(t, map[string]any{"model": "approved-settings-marker"})
			log := p.exec(t, true)
			rules(t, log, 0)
			require(t, log, "run_preset.resolve", "ready", map[string]int{})
			if !strings.Contains(log, `model_source="settings"`) {
				t.Fatal("approved setting source missing")
			}
		})
	}
	t.Run("poison-user-control", func(t *testing.T) {
		p := h.fixture(t)
		write(t, filepath.Join(p.root, "config/muse/settings.json"), []byte("INVALID_SETTINGS_CONTROL\n"))
		out, code := p.vendor(t, "exec", "--provider", "echo", "--disable-reminders", "--approval-judge", "off", "CONTROL")
		if code != 1 || !bytes.Contains(out, []byte("malformed settings file")) {
			t.Fatalf("poison user control exit=%d: %s", code, out)
		}
	})
}

// The sandbox permits a harmless fixture read and a positive loopback request
// outside it, but rejects the same connection and synthetic Keychains read inside.
func TestOfflineBoundary(t *testing.T) {
	h := setup(t)
	p := h.fixture(t)
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hits.Add(1); _, _ = w.Write([]byte("offline-control")) }))
	defer srv.Close()
	args := []string{"--silent", "--show-error", "--noproxy", "*", "--connect-timeout", "2", "--max-time", "3", srv.URL}
	out, code := p.command(t, "/usr/bin/curl", args...)
	if code != 0 || string(out) != "offline-control" || hits.Load() != 1 {
		t.Fatalf("loopback positive exit=%d hits=%d", code, hits.Load())
	}
	sandboxArgs := []string{"-f", p.sandbox, "/usr/bin/curl"}
	sandboxArgs = append(sandboxArgs, args...)
	_, code = p.command(t, "/usr/bin/sandbox-exec", sandboxArgs...)
	if code != 7 || hits.Load() != 1 {
		t.Fatalf("network deny exit=%d hits=%d", code, hits.Load())
	}
	ordinary := filepath.Join(p.root, "ordinary")
	secret := filepath.Join(p.root, "Keychains/sentinel")
	write(t, ordinary, []byte("ordinary"))
	write(t, secret, []byte("synthetic"))
	out, code = p.command(t, "/usr/bin/sandbox-exec", "-f", p.sandbox, "/bin/cat", ordinary)
	if code != 0 || string(out) != "ordinary" {
		t.Fatal("ordinary positive failed")
	}
	out, code = p.command(t, "/usr/bin/sandbox-exec", "-f", p.sandbox, "/bin/cat", secret)
	if code != 1 || !bytes.Contains(out, []byte("Operation not permitted")) || string(read(t, secret)) != "synthetic" {
		t.Fatal("synthetic Keychains denial failed")
	}
}

// A source beyond the checkout's Git boundary is not discovered in either trust
// mode; P1's inside-checkout positive prevents mistaking this for general exclusion.
func TestP1OutsideGitBoundary(t *testing.T) {
	h := setup(t)
	for _, trust := range []bool{true, false} {
		t.Run(fmt.Sprint(trust), func(t *testing.T) {
			p := h.fixture(t)
			write(t, filepath.Join(filepath.Dir(p.work), "AGENTS.md"), []byte("OUTSIDE_GIT\n"))
			rules(t, p.exec(t, trust), 0)
		})
	}
}
