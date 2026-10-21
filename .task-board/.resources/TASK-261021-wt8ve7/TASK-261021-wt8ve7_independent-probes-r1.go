package claudeexclusion

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Distinguishes old injected instruction history from newly loaded sources.
// Old project bytes are replaced and approved home bytes changed before resume.
func TestReviewerResumeProvenance(t *testing.T) {
	for _, fork := range []bool{false, true} {
		t.Run(fmt.Sprintf("fork-%t", fork), func(t *testing.T) {
			p := setup(t)
			write(t, filepath.Join(p.work, "CLAUDE.md"), "EXCL_OLD_PROJECT_INSTRUCTION\n")
			first := p.run("first", false)
			contains(t, payload(t, first), "EXCL_OLD_PROJECT_INSTRUCTION", true)
			write(t, filepath.Join(p.work, "CLAUDE.md"), "EXCL_NEW_PROJECT_INSTRUCTION\n")
			write(t, filepath.Join(p.home, "CLAUDE.md"), approved+"\nEXCL_CHANGED_APPROVED_HOME\n")
			args := []string{"--resume", first.Init["session_id"].(string)}
			if fork {
				args = append(args, "--fork-session")
			}
			resumed := p.run("resume", true, args...)
			p.assertRoot(resumed)
			contains(t, payload(t, resumed), "EXCL_OLD_PROJECT_INSTRUCTION", true)
			contains(t, payload(t, resumed), "EXCL_NEW_PROJECT_INSTRUCTION", false)
			// Record exact field locations and surrounding text for independent inspection.
			p.save("provenance.json", mustJSON(t, map[string]any{"requests": resumed.Requests}))
			fresh := p.run("fresh", true)
			contains(t, payload(t, fresh), "EXCL_OLD_PROJECT_INSTRUCTION", false)
			contains(t, payload(t, fresh), "EXCL_NEW_PROJECT_INSTRUCTION", false)
			contains(t, payload(t, fresh), "EXCL_CHANGED_APPROVED_HOME", true)
		})
	}
}

// Sparse materialization is measured through real startup and SDK skill reload.
// A late file is a control showing whether sparse index alone prevents discovery.
func TestReviewerSparse(t *testing.T) {
	for _, exclude := range []bool{false, true} {
		t.Run(fmt.Sprintf("exclude-%t", exclude), func(t *testing.T) {
			p := setup(t)
			original := p.work
			write(t, filepath.Join(original, "CLAUDE.md"), foreign+"\n")
			write(t, filepath.Join(original, "src", "ordinary.txt"), "EXCL_SPARSE_WORK_DATA\n")
			p.skill(filepath.Join(original, ".claude"), "foreign-probe", "EXCL_FOREIGN_SKILL")
			p.skill(p.home, "approved-probe", "EXCL_APPROVED_SKILL")
			p.settings("user", map[string]any{"enableAllProjectMcpServers": true, "hooks": p.hook("approved-hook")})
			p.settings("project", map[string]any{"hooks": p.hook("foreign-hook")})
			write(t, filepath.Join(original, ".mcp.json"), mustJSON(t, map[string]any{"mcpServers": map[string]any{"foreign-probe": p.mcp("foreign-mcp")}}))
			accepted := filepath.Join(p.home, "approved-mcp.json")
			write(t, accepted, mustJSON(t, map[string]any{"mcpServers": map[string]any{"approved-probe": p.mcp("approved-mcp")}}))
			git := func(dir string, args ...string) string {
				cmd := exec.Command("/usr/bin/git", args...)
				cmd.Dir = dir
				cmd.Env = p.env
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("synthetic git %v: %v %s", args, err, out)
				}
				return string(out)
			}
			git(original, "add", ".")
			git(original, "-c", "user.name=Synthetic Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "fixture")
			baseline := p.run("original", false, "--mcp-config", accepted)
			for _, m := range []string{foreign, "EXCL_FOREIGN_SKILL", approved, "EXCL_APPROVED_SKILL"} {
				contains(t, payload(t, baseline), m, true)
			}
			for _, name := range []string{"approved-hook", "foreign-hook", "approved-mcp", "foreign-mcp"} {
				if _, err := os.Stat(filepath.Join(p.root, name+".witness")); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(filepath.Join(p.root, name+".witness")); err != nil {
					t.Fatal(err)
				}
			}
			view := filepath.Join(p.root, "external", "view")
			git(original, "worktree", "add", "--detach", "--no-checkout", view, "HEAD")
			git(view, "sparse-checkout", "set", "--no-cone", "/src/")
			git(view, "read-tree", "-mu", "HEAD")
			if read(t, filepath.Join(view, "src", "ordinary.txt")) != "EXCL_SPARSE_WORK_DATA\n" {
				t.Fatal("work data missing")
			}
			index := git(view, "ls-files", "-t")
			contains(t, index, "S CLAUDE.md", true)
			for _, name := range []string{"CLAUDE.md", ".claude", ".mcp.json"} {
				if _, err := os.Stat(filepath.Join(view, name)); !os.IsNotExist(err) {
					t.Fatalf("materialized %s: %v", name, err)
				}
			}
			contains(t, read(t, filepath.Join(original, "CLAUDE.md")), foreign, true)
			p.work = view
			sparse := p.run("sparse-start", exclude, "--mcp-config", accepted)
			p.assertRoot(sparse)
			contains(t, payload(t, sparse), foreign, false)
			contains(t, payload(t, sparse), "EXCL_FOREIGN_SKILL", false)
			contains(t, payload(t, sparse), "EXCL_APPROVED_SKILL", true)
			contains(t, mustJSON(t, sparse.Init["mcp_servers"]), "approved-probe", true)
			contains(t, mustJSON(t, sparse.Init["mcp_servers"]), "foreign-probe", false)
			for _, name := range []string{"foreign-hook", "foreign-mcp"} {
				if _, err := os.Stat(filepath.Join(p.root, name+".witness")); !os.IsNotExist(err) {
					t.Fatalf("forbidden effect %s", name)
				}
			}
			for _, name := range []string{"approved-hook", "approved-mcp"} {
				if _, err := os.Stat(filepath.Join(p.root, name+".witness")); err != nil {
					t.Fatal(err)
				}
			}
			for _, text := range []string{"EXCLUSION_PROBE_FIRST", "/reload-skills", "EXCLUSION_PROBE_SECOND"} {
				b, err := json.Marshal(map[string]any{"type": "user", "message": map[string]any{"role": "user", "content": text}})
				if err != nil {
					t.Fatal(err)
				}
				p.input += string(b) + "\n"
			}
			p.before = func(n int) {
				if n == len(baseline.Requests)+len(sparse.Requests) {
					p.skill(filepath.Join(view, ".claude"), "foreign-late", "EXCL_FOREIGN_LATE_SKILL")
					p.skill(p.home, "approved-late", "EXCL_APPROVED_LATE_SKILL")
				}
			}
			late := p.run("sparse-late", exclude, "--mcp-config", accepted)
			if len(late.Requests) < 2 {
				t.Fatal("missing second SDK turn")
			}
			contains(t, mustJSON(t, late.Requests[0]), "EXCL_FOREIGN_LATE_SKILL", false)
			contains(t, mustJSON(t, late.Requests[1:]), "EXCL_FOREIGN_LATE_SKILL", !exclude)
			contains(t, mustJSON(t, late.Requests[1:]), "EXCL_APPROVED_LATE_SKILL", true)
			p.save("layout.json", mustJSON(t, map[string]any{"index": index, "original_sources_preserved": true, "late_discovery_without_source_flag": !exclude}))
		})
	}
}
