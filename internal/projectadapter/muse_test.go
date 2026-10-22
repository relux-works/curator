package projectadapter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func offlineFixture() MuseOfflineInput {
	files := map[string][]byte{"config/muse/AGENTS.md": []byte("Profile.\n\nProject.\n"), "data/muse/skills/helper/SKILL.md": []byte("# Helper\n"), "data/muse/skills/helper/scripts/tool": []byte("inert artifact\n")}
	policy := []byte(`{"policy":{"memory":"off"},"permissions":"manager-default-deny"}`)
	in := MuseOfflineInput{Files: files, Generated: map[string]Surface{}, PolicyRecord: policy, PolicySHA256: sum(policy), Memory: "off"}
	for p, b := range files {
		in.Generated[p] = Surface{sum(b), int64(len(b)), 0600}
	}
	s := in.Generated["data/muse/skills/helper/scripts/tool"]
	s.Mode = 0700
	in.Generated["data/muse/skills/helper/scripts/tool"] = s
	return in
}
func offlineHome(t *testing.T) string {
	t.Helper()
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("first-slice filesystem qualification is macOS/Linux only")
	}
	parent := t.TempDir()
	if err := os.Chmod(parent, 0700); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(parent, "offline-muse")
}
func requireError(t *testing.T, err error, text string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), text) {
		t.Fatalf("want %s; got %v", text, err)
	}
}

// Claim: offline copying preserves approved bytes and full receipt, creates
// four XDG parents, and retains manager policy without inventing native settings.
// This proves no native loader or permission enforcement behavior.
func TestMuseOfflineMaterialization(t *testing.T) {
	in := offlineFixture()
	home := offlineHome(t)
	got, err := MaterializeMuseOffline(context.Background(), home, in)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Parents) != 4 || got.Memory != "off" || got.Mode != ModeStrict || !bytes.Equal(got.PolicyRecord, in.PolicyRecord) {
		t.Fatal("offline plan differs")
	}
	for key, suffix := range map[string]string{"XDG_CONFIG_HOME": "config", "XDG_DATA_HOME": "data", "XDG_STATE_HOME": "state", "XDG_CACHE_HOME": "cache"} {
		if got.Parents[key] != filepath.Join(home, suffix) {
			t.Fatal("XDG parent differs")
		}
		if info, e := os.Stat(got.Parents[key]); e != nil || !info.IsDir() || info.Mode().Perm() != 0700 {
			t.Fatal("XDG parent not private")
		}
	}
	for p, want := range in.Files {
		b, e := os.ReadFile(filepath.Join(home, filepath.FromSlash(p)))
		if e != nil || !bytes.Equal(b, want) || got.Generated[p] != in.Generated[p] {
			t.Fatal("offline output differs")
		}
	}
	in.PolicyRecord[0] = 'x'
	in.Files["config/muse/AGENTS.md"][0] = 'x'
	delete(in.Generated, "config/muse/AGENTS.md")
	if err := VerifyMuseOffline(home, got.Generated); err != nil {
		t.Fatal("caller mutation changed retained receipt", err)
	}
	if !bytes.Equal(got.PolicyRecord, offlineFixture().PolicyRecord) {
		t.Fatal("caller mutation changed policy")
	}
	if _, e := os.Stat(filepath.Join(home, "config/muse/settings.json")); !os.IsNotExist(e) {
		t.Fatal("native settings invented")
	}
}

// Claim: required unqualified MCP refuses before allocation, while approved
// optional MCP is omitted by item name with no fabricated native writer path.
func TestMuseMCPDisposition(t *testing.T) {
	in := offlineFixture()
	in.MCPItems = []MuseMCPItem{{"optional-docs", false}}
	got, err := MaterializeMuseOffline(context.Background(), offlineHome(t), in)
	if err != nil || len(got.Omitted) != 1 || got.Omitted[0] != (Omission{"optional-docs", "unsupported"}) {
		t.Fatalf("optional disposition: %+v %v", got, err)
	}
	in.MCPItems[0].Required = true
	home := offlineHome(t)
	_, err = MaterializeMuseOffline(context.Background(), home, in)
	var unsupported *Unsupported
	if !errors.As(err, &unsupported) || unsupported.Surface != MCP || unsupported.Code() != "mcp_contribution_policy_refused" || unsupported.Reason != "native_writer_unqualified" {
		t.Fatalf("required MCP was not refused: %v", err)
	}
	if _, e := os.Stat(home); !os.IsNotExist(e) {
		t.Fatal("required MCP published partial controls")
	}
}

// Claim: additional repository instructions, settings and approval files cannot
// be imported through the offline copier even with matching self-claimed hashes.
// The valid approved inventory succeeds beside each refusal; launch is separate.
func TestMuseRepositoryControls(t *testing.T) {
	for _, p := range []string{"config/muse/approval.json", "config/muse/settings.json", "config/muse/trust.json", "AGENTS.md", "config/muse/CLAUDE.md", ".muse/settings.json"} {
		t.Run(p, func(t *testing.T) {
			in := offlineFixture()
			if _, err := MaterializeMuseOffline(context.Background(), offlineHome(t), in); err != nil {
				t.Fatal(err)
			}
			in.Files[p] = []byte(`{"approved":true,"permissions":"allow-all"}`)
			in.Generated[p] = Surface{sum(in.Files[p]), int64(len(in.Files[p])), 0600}
			home := offlineHome(t)
			_, err := MaterializeMuseOffline(context.Background(), home, in)
			requireError(t, err, "unapproved_control_path")
			if _, e := os.Stat(home); !os.IsNotExist(e) {
				t.Fatal("unapproved controls partially written")
			}
		})
	}
}

// Claim: bad policy, memory-on, incomplete/changed receipts, forbidden modes,
// path aliases/collisions and escaping paths refuse without an allocation.
func TestMuseOfflineInputGates(t *testing.T) {
	cases := []struct {
		name   string
		change func(*MuseOfflineInput)
	}{
		{"memory", func(in *MuseOfflineInput) { in.Memory = "on" }},
		{"policy", func(in *MuseOfflineInput) { in.PolicyRecord = []byte(`{"changed":true}`) }},
		{"empty", func(in *MuseOfflineInput) { in.Files = nil; in.Generated = nil }},
		{"missing", func(in *MuseOfflineInput) { delete(in.Generated, "config/muse/AGENTS.md") }},
		{"hash", func(in *MuseOfflineInput) {
			s := in.Generated["config/muse/AGENTS.md"]
			s.SHA256 = strings.Repeat("0", 64)
			in.Generated["config/muse/AGENTS.md"] = s
		}},
		{"size", func(in *MuseOfflineInput) {
			s := in.Generated["config/muse/AGENTS.md"]
			s.Size++
			in.Generated["config/muse/AGENTS.md"] = s
		}},
		{"mode", func(in *MuseOfflineInput) {
			s := in.Generated["config/muse/AGENTS.md"]
			s.Mode = 0644
			in.Generated["config/muse/AGENTS.md"] = s
		}},
		{"escape", func(in *MuseOfflineInput) { addOffline(in, "data/muse/skills/helper/../../escape", []byte("bad")) }},
		{"alias", func(in *MuseOfflineInput) { addOffline(in, "data/muse/skills/helper/skill.md", []byte("bad")) }},
		{"unicode-alias", func(in *MuseOfflineInput) {
			addOffline(in, "data/muse/skills/helper/caf\u00e9.md", []byte("one"))
			addOffline(in, "data/muse/skills/helper/cafe\u0301.md", []byte("two"))
		}},
		{"ancestor", func(in *MuseOfflineInput) { addOffline(in, "data/muse/skills/helper/scripts", []byte("bad")) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := offlineFixture()
			if _, err := MaterializeMuseOffline(context.Background(), offlineHome(t), in); err != nil {
				t.Fatal(err)
			}
			tc.change(&in)
			want := "generated bytes or metadata differ"
			switch tc.name {
			case "memory", "policy":
				want = "manager policy or memory differs"
			case "empty", "missing":
				want = "complete receipt required"
			case "escape":
				want = "unapproved_control_path"
			case "alias", "unicode-alias":
				want = "generated path alias"
			case "ancestor":
				want = "generated file ancestor collision"
			}
			_, _, preflightErr := museInput(in)
			requireError(t, preflightErr, want)
			home := offlineHome(t)
			_, err := MaterializeMuseOffline(context.Background(), home, in)
			requireError(t, err, want)
			if _, e := os.Stat(home); !os.IsNotExist(e) {
				t.Fatal("invalid input allocated home")
			}
		})
	}
}

// Claim: receipt sizes are checked independently of hashes, and malformed MCP
// identities cannot become optional omissions beside a valid optional item.
func TestMuseIndependentMetadata(t *testing.T) {
	home := offlineHome(t)
	out, err := MaterializeMuseOffline(context.Background(), home, offlineFixture())
	if err != nil {
		t.Fatal(err)
	}
	s := out.Generated["config/muse/AGENTS.md"]
	s.Size++
	out.Generated["config/muse/AGENTS.md"] = s
	requireError(t, VerifyMuseOffline(home, out.Generated), "offline metadata differs")
	for _, items := range [][]MuseMCPItem{{{"../escape", false}}, {{"same", false}, {"same", false}}} {
		in := offlineFixture()
		in.MCPItems = items
		_, _, err := museInput(in)
		requireError(t, err, "invalid MCP item identity")
	}
}

// Claim: parent privacy and allocation grammar are preconditions, not repaired
// by copying. Symlink allocations are preserved instead of followed/replaced.
func TestMuseAllocationBoundaries(t *testing.T) {
	home := offlineHome(t)
	parent := filepath.Dir(home)
	if err := os.Chmod(parent, 0755); err != nil {
		t.Fatal(err)
	}
	_, err := MaterializeMuseOffline(context.Background(), home, offlineFixture())
	requireError(t, err, "offline parent unavailable")
	if err := os.Chmod(parent, 0700); err != nil {
		t.Fatal(err)
	}
	_, err = MaterializeMuseOffline(context.Background(), "relative", offlineFixture())
	requireError(t, err, "invalid offline allocation")
	if err := os.Symlink(parent, home); err != nil {
		t.Fatal(err)
	}
	_, err = MaterializeMuseOffline(context.Background(), home, offlineFixture())
	requireError(t, err, "already exists")
	if info, e := os.Lstat(home); e != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("existing link changed")
	}
}
func addOffline(in *MuseOfflineInput, p string, b []byte) {
	in.Files[p] = b
	in.Generated[p] = Surface{sum(b), int64(len(b)), 0600}
}

// Claim: the offline preflight enforces the same file-count, per-file and total
// byte caps as I-COMP. A bounded valid inventory succeeds before each violation.
func TestMuseOfflineBounds(t *testing.T) {
	for _, kind := range []string{"count", "file", "total", "policy-json"} {
		t.Run(kind, func(t *testing.T) {
			in := offlineFixture()
			if _, _, err := museInput(in); err != nil {
				t.Fatal(err)
			}
			want := "generated bytes or metadata differ"
			switch kind {
			case "count":
				for i := 0; i < 10001; i++ {
					addOffline(&in, fmt.Sprintf("data/muse/skills/helper/file-%d", i), []byte("x"))
				}
				want = "complete receipt required"
			case "file":
				addOffline(&in, "data/muse/skills/helper/large", make([]byte, (16<<20)+1))
			case "total":
				for i := 0; i < 4; i++ {
					addOffline(&in, fmt.Sprintf("data/muse/skills/helper/large-%d", i), make([]byte, 16<<20))
				}
				want = "generated byte bound exceeded"
			case "policy-json":
				in.PolicyRecord = []byte("invalid")
				in.PolicySHA256 = sum(in.PolicyRecord)
				want = "manager policy or memory differs"
			}
			_, _, err := museInput(in)
			requireError(t, err, want)
		})
	}
}

// Claim: offline drift is detected without rebaselining, including changes in
// bytes/size/mode, missing or extra entries, and a linked control file.
func TestMuseOfflineDrift(t *testing.T) {
	for _, kind := range []string{"bytes", "size", "mode", "missing", "extra", "link"} {
		t.Run(kind, func(t *testing.T) {
			home := offlineHome(t)
			got, err := MaterializeMuseOffline(context.Background(), home, offlineFixture())
			if err != nil {
				t.Fatal(err)
			}
			if err = VerifyMuseOffline(home, got.Generated); err != nil {
				t.Fatal(err)
			}
			p := filepath.Join(home, "config/muse/AGENTS.md")
			switch kind {
			case "bytes":
				err = os.WriteFile(p, []byte("Changed.\n\nProject.\n"), 0600)
			case "size":
				err = os.WriteFile(p, []byte("x"), 0600)
			case "mode":
				err = os.Chmod(p, 0700)
			case "missing":
				err = os.Remove(p)
			case "extra":
				err = os.WriteFile(filepath.Join(home, "config/muse/settings.json"), []byte("{}"), 0600)
			case "link":
				if err = os.Remove(p); err == nil {
					err = os.Symlink("../../data/muse/skills/helper/SKILL.md", p)
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			requireError(t, VerifyMuseOffline(home, got.Generated), "project_context_drift")
			if got.Generated["config/muse/AGENTS.md"].SHA256 != sum(offlineFixture().Files["config/muse/AGENTS.md"]) {
				t.Fatal("receipt was rebaselined")
			}
		})
	}
}

// Claim: existing allocations and cancelled work are preserved; a fresh
// offline allocation remains usable. These checks grant no live lease authority.
func TestMuseOfflinePreservesExisting(t *testing.T) {
	home := offlineHome(t)
	got, err := MaterializeMuseOffline(context.Background(), home, offlineFixture())
	if err != nil {
		t.Fatal(err)
	}
	_, err = MaterializeMuseOffline(context.Background(), home, offlineFixture())
	requireError(t, err, "already exists")
	if err = VerifyMuseOffline(home, got.Generated); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	other := offlineHome(t)
	_, err = MaterializeMuseOffline(ctx, other, offlineFixture())
	if err != context.Canceled {
		t.Fatalf("cancelled preparation: %v", err)
	}
	if _, err = os.Stat(other); !os.IsNotExist(err) {
		t.Fatal("cancelled allocation published")
	}
}

// Claim: the exact measured tuple and all unknown tuples report no full strict
// certified surface, and exec/interactive/resume never contribute env or argv.
// Offline materialization is the positive control, not a successful launch.
func TestMuseLaunchRefused(t *testing.T) {
	for _, mode := range []string{ModeStrict, ModeNativeProject} {
		for _, version := range []string{MuseVersion, "1.4.10-R6643.1", "future"} {
			for _, platform := range []string{"macos", "linux"} {
				for _, form := range []string{"exec", "interactive", "resume"} {
					t.Run(mode+"/"+version+"/"+platform+"/"+form, func(t *testing.T) {
						report := MuseCertification(version, platform, mode, form)
						if len(report.Certified) != 0 || len(report.Refused) < 7 || report.Environment != "muse" || report.ProviderVersion != version || report.Platform != platform || report.Mode != mode || report.Form != form {
							t.Fatal("native certification invented")
						}
						surfaces := []SurfaceName{}
						for _, refusal := range report.Refused {
							if refusal.Reason == "" {
								t.Fatal("missing refusal reason")
							}
							surfaces = append(surfaces, refusal.Surface)
						}
						want := []SurfaceName{Commands, Context, MCP, MemoryOff, NativeSourceExclusion, Permissions, Skills}
						if form != "exec" {
							want = append(want, LaunchForm)
						}
						if !reflect.DeepEqual(surfaces, want) {
							t.Fatal("refused surface set differs")
						}
						contribution, err := MuseLaunch(version, platform, mode, form)
						var unsupported *Unsupported
						if !errors.As(err, &unsupported) || unsupported.Code() != "project_context_unsupported" {
							t.Fatalf("unsafe launch accepted: %v", err)
						}
						if contribution.Env != nil || contribution.Args != nil {
							t.Fatal("refused launch contributed executable data")
						}
						if form != "exec" && unsupported.Surface != LaunchForm {
							t.Fatal("interactive/resume gate not applied")
						}
						if form == "exec" && mode == ModeStrict && unsupported.Surface != NativeSourceExclusion {
							t.Fatal("mandatory exclusion gate not applied")
						}
						if mode == ModeNativeProject {
							if report.Notice != "project sources not isolated" {
								t.Fatal("native source warning missing")
							}
							if !strings.Contains(err.Error(), "project sources not isolated") {
								t.Fatal("native launch warning missing")
							}
							if form == "exec" && (unsupported.Surface != Permissions || !strings.Contains(err.Error(), "project sources not isolated")) {
								t.Fatal("mandatory permissions gate not applied")
							}
						}
					})
				}
			}
		}
	}
	_, err := MuseLaunch(MuseVersion, "macos", "unknown", "exec")
	requireError(t, err, "unknown_composition_mode")
}
