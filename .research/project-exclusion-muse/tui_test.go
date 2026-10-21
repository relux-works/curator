package museexclusion

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

type terminal struct {
	p    *probe
	t    *testing.T
	tmux string
}

func (p *probe) terminal(t *testing.T, args ...string) *terminal {
	t.Helper()
	tmux, e := exec.LookPath("tmux")
	if e != nil {
		t.Fatal(e)
	}
	v := &terminal{p: p, t: t, tmux: tmux}
	// Relative private socket avoids macOS path limits. No user tmux config loads.
	quoted := []string{quote("/usr/bin/sandbox-exec")}
	for _, a := range p.args(args...) {
		quoted = append(quoted, quote(a))
	}
	cmd := "cd " + quote(p.work) + " && " + strings.Join(quoted, " ") + "; printf '%s\\n' \"$?\" > " + quote(filepath.Join(p.root, "tui-exit.pending")) + "; /bin/mv " + quote(filepath.Join(p.root, "tui-exit.pending")) + " " + quote(filepath.Join(p.root, "tui-exit.txt"))
	v.call("new-session", "-d", "-x", "220", "-y", "70", "-s", "probe", cmd)
	ids := strings.Fields(v.call("display-message", "-p", "#{pid} #{pane_pid}"))
	if len(ids) != 2 {
		t.Fatal("missing private process identities")
	}
	write(t, filepath.Join(p.root, "terminal-pids.txt"), []byte(strings.Join(ids, " ")+"\n"))
	pid, e := strconv.Atoi(ids[0])
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		out, code := p.command(t, tmux, "-S", filepath.Join("..", "..", "tui.sock"), "display-message", "-p", "#{pid}")
		if code == 0 && strings.TrimSpace(string(out)) == strconv.Itoa(pid) {
			proc, e := os.FindProcess(pid)
			if e == nil {
				_ = proc.Signal(syscall.SIGTERM)
			}
		}
	})
	return v
}
func (v *terminal) call(args ...string) string {
	v.t.Helper()
	out, code := v.p.command(v.t, v.tmux, append([]string{"-S", filepath.Join("..", "..", "tui.sock"), "-f", "/dev/null"}, args...)...)
	if code != 0 {
		v.t.Fatalf("private tmux exit=%d: %s", code, out)
	}
	return string(out)
}
func (v *terminal) pane() string { return v.call("capture-pane", "-p", "-S", "-2000") }
func (v *terminal) wait(needle string) string {
	v.t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	var pane string
	for time.Now().Before(deadline) {
		pane = v.pane()
		if strings.Contains(pane, needle) {
			return pane
		}
		time.Sleep(100 * time.Millisecond)
	}
	v.t.Fatalf("TUI missing %q: %s", needle, pane)
	return ""
}
func (v *terminal) send(s string) {
	v.t.Helper()
	v.call("send-keys", "-l", s)
	time.Sleep(150 * time.Millisecond)
	v.call("send-keys", "Enter")
}
func (v *terminal) exit() {
	v.t.Helper()
	v.send("/exit")
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if b, e := os.ReadFile(filepath.Join(v.p.root, "tui-exit.txt")); e == nil {
			if strings.TrimSpace(string(b)) != "0" {
				v.t.Fatalf("TUI exit %s", b)
			}
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	v.t.Fatal("no normal TUI exit")
}

var tuiBase = []string{"--provider", "echo", "--approval-judge", "off", "-c", "run.reminder_roster=[]"}

// P7 observes retained root-rule snapshots separately from live catalogs after
// granting the resume trust gate. Approved home sources remain usable; any
// revived project catalog defeats a claim of complete resume exclusion.
func TestP7ResumeTrustGate(t *testing.T) {
	h := setup(t)
	p := h.fixture(t)
	write(t, filepath.Join(p.work, "AGENTS.md"), []byte("MUSE_EXCL_RESUME_PROJECT\n"))
	p.skill(t, filepath.Join(p.work, ".agents/skills/project/SKILL.md"))
	p.skill(t, filepath.Join(p.root, "config/muse/skills/approved/SKILL.md"))
	write(t, filepath.Join(p.work, ".muse/hooks.json"), jsonBytes(t, map[string]any{"hooks": p.hooks(t, "project-hook")}))
	write(t, filepath.Join(p.work, ".mcp.json"), []byte(`{"mcpServers":{"project-probe":{"command":"/usr/bin/true"}}}`))
	p.settings(t, map[string]any{"hooks": p.hooks(t, "approved-hook"), "mcpServers": map[string]any{"approved-probe": map[string]any{"command": "/usr/bin/true"}}})
	rules(t, p.exec(t, false), 0)
	args := append([]string{"resume", "--last"}, tuiBase...)
	v := p.terminal(t, args...)
	screen := v.wait("Do you trust this workspace?")
	write(t, filepath.Join(p.root, "resume-trust.txt"), []byte(screen))
	if !strings.Contains(screen, "2  Quit") || strings.Contains(screen, "Continue without trusting") {
		t.Fatal("resume trust choices changed")
	}
	v.call("send-keys", "1", "Enter")
	v.wait("Voice input")
	v.send("MUSE_EXCL_RESUME")
	v.wait("echo: MUSE_EXCL_RESUME")
	v.send("/rules")
	screen = v.wait("Rules loaded this session")
	write(t, filepath.Join(p.root, "resume-rules.txt"), []byte(screen))
	if !strings.Contains(screen, "user: $CONFIG_DIR/AGENTS.md") || !strings.Contains(screen, "project: none") {
		t.Fatal("resumed rule snapshot changed")
	}
	v.exit()
	log := p.startup(t)
	require(t, log, "skills.catalog_load", "ready", map[string]int{"project_skills": 1, "user_skills": 1})
	require(t, log, "hook.admission", "ready", map[string]int{"handlers": 2, "runnable": 2, "inactive": 0})
	require(t, log, "mcp.config.resolve", "ready", map[string]int{"trusted_project_source_count": 1, "user_settings_source_count": 1, "server_count": 2})
	if string(read(t, filepath.Join(p.root, "project-hook"))) != "MUSE_EXCL_HOOK\n" {
		t.Fatal("resumed project hook did not execute")
	}
	t.Log("resume retains root snapshot but trusting revives project skill/hook/MCP catalogs; child execution unmeasured")
}

// P8 observes /rules before and after a late-created root in one trusted TUI,
// and compares a fresh trusted startup. Echo cannot drive a lazy read_file or
// child tool, so this test deliberately cannot certify late-discovery exclusion.
func TestP8LateRoot(t *testing.T) {
	h := setup(t)
	p := h.fixture(t)
	args := append(append([]string{}, tuiBase...), "--trust-workspace")
	v := p.terminal(t, args...)
	v.wait("Voice input")
	v.send("MUSE_EXCL_BEFORE")
	v.wait("echo: MUSE_EXCL_BEFORE")
	v.send("/rules")
	before := v.wait("Rules loaded this session")
	write(t, filepath.Join(p.root, "late-before.txt"), []byte(before))
	write(t, filepath.Join(p.work, "AGENTS.md"), []byte("MUSE_EXCL_LATE_PROJECT\n"))
	v.send("MUSE_EXCL_AFTER")
	v.wait("echo: MUSE_EXCL_AFTER")
	v.send("/rules")
	deadline := time.Now().Add(5 * time.Second)
	var after string
	for time.Now().Before(deadline) {
		after = v.pane()
		if strings.Count(after, "Rules loaded this session") == 2 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if strings.Count(after, "Rules loaded this session") != 2 || strings.Count(after, "project: none") != 2 || strings.Count(after, "user: $CONFIG_DIR/AGENTS.md") != 2 {
		t.Fatal("missing distinct before/after native rule observations")
	}
	write(t, filepath.Join(p.root, "late-after.txt"), []byte(after))
	v.exit()
	log := p.startup(t)
	rules(t, log, 0)
	// A separate native process supplies the nearby positive for the same file.
	rules(t, p.exec(t, true), 1)
	t.Logf("late source created between echo turns; /rules snapshots retained (%d/%d bytes); no lazy-tool claim", len(before), len(after))
}

// Untrusted exec removes delegation availability, but absence of a real child
// tool invocation is explicitly a coverage gap rather than proof of inheritance.
func TestP7ChildAvailability(t *testing.T) {
	h := setup(t)
	p := h.fixture(t)
	write(t, filepath.Join(p.work, "AGENTS.md"), []byte("PROJECT\n"))
	out, code := p.vendor(t, "exec", "--json", "--provider", "echo", "--disable-reminders", "--approval-judge", "off", "CONTROL")
	if code != 0 || !strings.Contains(string(out), "Agent delegation: auto unavailable: workspace is untrusted.") {
		t.Fatalf("child availability exit=%d: %s", code, out)
	}
	rules(t, p.startup(t), 0)
	t.Log("native child invocation: unmeasured; offline echo emits text only")
}
