package envprofile

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/testhomeguard"
	"github.com/relux-works/curator/internal/whitelist"
)

const renderedGuardOrigin = "CURATOR_TEST_RENDERED_GUARD_ORIGIN"
const renderedGuardShape = "CURATOR_TEST_RENDERED_GUARD_SHAPE"
const renderedGuardPlant = "CURATOR_TEST_RENDERED_GUARD_PLANT"

func renderedGuardPayload(t *testing.T, shape, text string) (string, []byte) {
	t.Helper()
	if shape == "referenced" {
		lock := &contextlock.Lock{Root: "fixture", Members: []contextlock.Member{{Name: "fixture", Kind: contextlock.KindContext, Version: "1.0.0", Commit: "0123456789abcdef0123456789abcdef01234567", Weight: 100}}}
		files, written, err := contextmaterialize.Referenced(lock, "sha256:aaaa", contextmaterialize.DefaultPrecedence, "claude_code", map[string]contextmaterialize.Package{
			"fixture": {HasContext: true, Modules: []contextmaterialize.Module{{Module: contextpkg.Module{Path: "instruction.md"}, Bytes: []byte("# " + text + "\n")}}},
		})
		rel := ".agent-context/modules/fixture/instruction.md"
		payload, found := files[rel]
		if err != nil || !written || !found {
			t.Fatalf("real referenced renderer: written=%v found=%v err=%v", written, found, err)
		}
		return rel, payload
	}
	rel := "skills/synced/references/manifest.json"
	if shape == "manifest-directory" {
		rel += "/SKILL.md"
	}
	return rel, []byte(text)
}

// Each pair uses the real projector/copier or referenced renderer, then the
// production managed writer and actual package TestMain. Only sentinel-checked
// temporary homes are planted. An isolated write passes; the same payload in
// the synthetic original home must fail and name its exact changed path.
func TestNativeHomeRenderedGuard(t *testing.T) {
	for _, shape := range []string{"synced-manifest", "manifest-directory", "referenced"} {
		for _, plant := range []bool{false, true} {
			name := "isolated"
			if plant {
				name = "plant"
			}
			t.Run(shape+"/"+name, func(t *testing.T) {
				origin, report := t.TempDir(), t.TempDir()
				must := func(err error) {
					t.Helper()
					if err != nil {
						t.Fatal(err)
					}
				}
				must(os.WriteFile(filepath.Join(origin, "rendered-guard-fixture"), []byte("synthetic original home"), 0600))
				native := filepath.Join(origin, ".claude")
				rel, before := renderedGuardPayload(t, shape, "before")
				if shape != "referenced" {
					source := t.TempDir()
					must(os.WriteFile(filepath.Join(source, "SKILL.md"), []byte("---\nname: synced\ndescription: Synthetic guard fixture\n---\n# Fixture\n"), 0600))
					payloadRel := strings.TrimPrefix(rel, "skills/synced/")
					must(os.MkdirAll(filepath.Dir(filepath.Join(source, filepath.FromSlash(payloadRel))), 0700))
					must(os.WriteFile(filepath.Join(source, filepath.FromSlash(payloadRel)), before, 0600))
					if _, err := projectSkill(t.TempDir(), source, hashing.VersionV2); err != nil {
						t.Fatalf("real skill projector: %v", err)
					}
					copied, err := whitelist.CopyContext(source, filepath.Join(native, "skills", "synced"), true, nil)
					must(err)
					found := false
					for _, path := range copied {
						if path == payloadRel {
							found = true
						}
					}
					if !found {
						t.Fatal("real copier did not retain collision payload")
					}
				} else {
					must(os.MkdirAll(native, 0700))
					must(atomicManagedFile(native, rel, before, 0600))
				}
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
				defer cancel()
				cmd := exec.CommandContext(ctx, os.Args[0], "-test.v", "-test.run=^TestNativeHomeRenderedGuardChild$") // #nosec G204,G702 -- This test binary, fixed child test and no shell.
				for _, entry := range os.Environ() {
					key, _, _ := strings.Cut(entry, "=")
					if key == "HOME" || key == "USERPROFILE" || strings.HasPrefix(key, "XDG_") || strings.HasPrefix(key, "CURATOR_TEST_HOME_GUARD_") || key == "CODEX_HOME" || key == "CLAUDE_CONFIG_DIR" || key == "PI_CODING_AGENT_DIR" || key == renderedGuardOrigin || key == renderedGuardShape || key == renderedGuardPlant {
						continue
					}
					cmd.Env = append(cmd.Env, entry)
				}
				cmd.Env = append(cmd.Env, "HOME="+origin, "USERPROFILE="+origin, renderedGuardOrigin+"="+origin, renderedGuardShape+"="+shape, "CURATOR_TEST_HOME_GUARD_REPORT="+report)
				if plant {
					cmd.Env = append(cmd.Env, renderedGuardPlant+"=1")
				}
				out, err := cmd.CombinedOutput()
				if !strings.Contains(string(out), "synthetic rendered write completed") {
					t.Fatalf("plant/control did not execute: %v\n%s", err, out)
				}
				if !plant {
					if err != nil || !strings.Contains(string(out), "protected native homes unchanged") {
						t.Fatalf("isolated positive control: %v\n%s", err, out)
					}
				} else {
					exit, ok := err.(*exec.ExitError)
					if !ok || exit.ExitCode() != 1 || !strings.Contains(string(out), testhomeguard.Changed) {
						t.Fatalf("rendered payload plant admitted: %v\n%s", err, out)
					}
				}
				// Assert exact-path evidence, independently of the aggregate refusal.
				target := filepath.Join(native, filepath.FromSlash(rel))
				found := false
				entries, err := os.ReadDir(report)
				must(err)
				for _, entry := range entries {
					f, err := os.Open(filepath.Join(report, entry.Name()))
					must(err)
					scanner := bufio.NewScanner(f)
					for scanner.Scan() {
						var row struct {
							Path string `json:"path"`
						}
						must(json.Unmarshal(scanner.Bytes(), &row))
						if row.Path == target {
							found = true
						}
						inside, err := filepath.Rel(origin, row.Path)
						if err != nil || inside == ".." || strings.HasPrefix(inside, ".."+string(filepath.Separator)) {
							t.Fatal("changed path outside the synthetic boundary")
						}
					}
					must(scanner.Err())
					must(f.Close())
				}
				if found != plant {
					t.Fatal("exact synthetic changed path missing or isolated control changed original home")
				}
				postimage, err := os.ReadFile(target)
				must(err) // Synthetic fixture only.
				_, want := renderedGuardPayload(t, shape, "after!")
				if !plant {
					want = before
				}
				if string(postimage) != string(want) {
					t.Fatal("unexpected synthetic postimage")
				}
			})
		}
	}
}

func TestNativeHomeRenderedGuardChild(t *testing.T) {
	origin := os.Getenv(renderedGuardOrigin)
	if origin == "" {
		return
	}
	marker, err := os.ReadFile(filepath.Join(origin, "rendered-guard-fixture")) // #nosec G304 -- Explicit synthetic fixture sentinel.
	if err != nil || string(marker) != "synthetic original home" || !filepath.IsAbs(origin) {
		t.Fatal("missing synthetic rendered fixture")
	}
	plant := os.Getenv(renderedGuardPlant) == "1"
	if plant {
		t.Setenv("HOME", origin)
		t.Setenv("USERPROFILE", origin)
	}
	home, err := os.UserHomeDir()
	if err != nil || (!plant && home == origin) {
		t.Fatal("neutral helper isolation missing")
	}
	adapter, _ := adapterByID("claude_code")
	native, err := NativeHome(adapter)
	if err != nil || native != filepath.Join(home, ".claude") {
		t.Fatal("native resolver escaped synthetic boundary")
	}
	rel, payload := renderedGuardPayload(t, os.Getenv(renderedGuardShape), "after!")
	if err := os.MkdirAll(native, 0700); err != nil {
		t.Fatal(err)
	}
	if err := atomicManagedFile(native, rel, payload, 0600); err != nil {
		t.Fatal(err)
	}
	t.Log("synthetic rendered write completed")
}
