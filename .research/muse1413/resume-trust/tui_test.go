package resumetrust

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
	write(t, filepath.Join(p.work, "AGENTS.md"), []byte("MUSE1413_RESUME_RESUME_PROJECT\n"))
	p.skill(t, filepath.Join(p.work, ".agents/skills/project/SKILL.md"))
	p.skill(t, filepath.Join(p.root, "config/muse/skills/approved/SKILL.md"))
	write(t, filepath.Join(p.work, ".muse/hooks.json"), jsonBytes(t, map[string]any{"hooks": p.hooks(t, "project-hook")}))
	write(t, filepath.Join(p.work, ".mcp.json"), []byte(`{"mcpServers":{"project-probe":{"command":"/usr/bin/true"}}}`))
	p.settings(t, map[string]any{"hooks": p.hooks(t, "approved-hook"), "mcpServers": map[string]any{"approved-probe": map[string]any{"command": "/usr/bin/true"}}})
	log := p.exec(t, false)
	rules(t, log, 0)
	require(t, log, "skills.catalog_load", "ready", map[string]int{"project_skills": 0, "user_skills": 1})
	require(t, log, "hook.admission", "partial", map[string]int{"handlers": 2, "runnable": 1, "inactive": 1})
	require(t, log, "mcp.config.resolve", "ready", map[string]int{"trusted_project_source_count": 0, "user_settings_source_count": 1, "server_count": 1})
	if _, err := os.Lstat(filepath.Join(p.root, "project-hook")); !os.IsNotExist(err) {
		t.Fatal("untrusted startup executed project hook or marker unreadable")
	}
	if string(read(t, filepath.Join(p.root, "approved-hook"))) != "MUSE_EXCL_HOOK\n" {
		t.Fatal("selected home hook did not execute")
	}
	args := append([]string{"resume", "--last"}, tuiBase...)
	v := p.terminal(t, args...)
	screen := v.wait("Do you trust this workspace?")
	write(t, filepath.Join(p.root, "resume-trust.txt"), []byte(screen))
	if !strings.Contains(screen, "2  Quit") || strings.Contains(screen, "Continue without trusting") {
		t.Fatal("resume trust choices changed")
	}
	v.call("send-keys", "1", "Enter")
	v.wait("Voice input")
	v.send("MUSE1413_RESUME_RESUME")
	v.wait("echo: MUSE1413_RESUME_RESUME")
	v.send("/rules")
	screen = v.wait("Rules loaded this session")
	write(t, filepath.Join(p.root, "resume-rules.txt"), []byte(screen))
	if !strings.Contains(screen, "user: $CONFIG_DIR/AGENTS.md") || !strings.Contains(screen, "project: none") {
		t.Fatal("resumed rule snapshot changed")
	}
	v.exit()
	log = p.startup(t)
	require(t, log, "skills.catalog_load", "ready", map[string]int{"project_skills": 1, "user_skills": 1})
	require(t, log, "hook.admission", "ready", map[string]int{"handlers": 2, "runnable": 2, "inactive": 0})
	require(t, log, "mcp.config.resolve", "ready", map[string]int{"trusted_project_source_count": 1, "user_settings_source_count": 1, "server_count": 2})
	if string(read(t, filepath.Join(p.root, "project-hook"))) != "MUSE_EXCL_HOOK\n" {
		t.Fatal("resumed project hook did not execute")
	}
	t.Log("resume retains root snapshot but trusting revives project skill/hook/MCP catalogs; child execution unmeasured")
}
