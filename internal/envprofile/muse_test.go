package envprofile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/stateread"
)

const museSuiteDigest = "bd03456b92a7368d90ea74fe6953db10bc188588020683024a6a6b8735a40783"

type museVector struct {
	Name     string `json:"name"`
	State    string `json:"state"`
	Repair   bool   `json:"repair"`
	Expected struct {
		Action       string `json:"action"`
		Diagnostic   string `json:"diagnostic"`
		Emit         bool   `json:"emit_fragment"`
		Status       string `json:"status"`
		BytesChanged bool   `json:"credential_bytes_changed"`
	} `json:"expected"`
}

func museFixture(t *testing.T) (*managedFixture, string, string) {
	t.Helper()
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "acme")
	fx.native[envregistry.Muse] = t.TempDir()
	native := filepath.Join(fx.native[envregistry.Muse], "auth.json")
	if err := os.WriteFile(native, []byte("synthetic-native\n"), 0600); err != nil {
		t.Fatal(err)
	}
	req := fx.request(envregistry.Muse)
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	return fx, filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.Muse), "config", "muse", "auth.json"), native
}

// All 16 source-derived states drive Resolve; StatusOf shares verifyHome.
func TestMusePublishedLinkStates(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	digest, err := conformancecoverage.SelectedSuiteManifestSHA256()
	if err != nil {
		t.Fatal(err)
	}
	if digest != museSuiteDigest {
		t.Log("selected suite has no Muse family; rc.13 remains unchanged")
		return
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "environments-muse.json"))
	if err != nil {
		t.Fatal(err)
	}
	var vectors struct {
		Cases []museVector `json:"cases"`
	}
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	conformancecoverage.Run(t, "environments-muse/cases", vectors.Cases, func(tc museVector) string { return tc.Name }, driveMuseLinkState)
}

func driveMuseLinkState(t *testing.T, tc museVector) {
	t.Helper()
	fx, link, native := museFixture(t)
	req := fx.request(envregistry.Muse)
	req.Repair = tc.Repair
	switch tc.State {
	case "live":
	case "write-through-refresh":
		if err := os.WriteFile(link, []byte("synthetic-refreshed\n"), 0600); err != nil {
			t.Fatal(err)
		}
	case "missing-link":
		if err := os.Remove(link); err != nil {
			t.Fatal(err)
		}
	case "temp-rename-fork":
		fork := filepath.Join(filepath.Dir(link), "fork.tmp")
		if err := os.WriteFile(fork, []byte("synthetic-fork\n"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(fork, link); err != nil {
			t.Fatal(err)
		}
	case "retargeted-link":
		other := filepath.Join(t.TempDir(), "other.json")
		if err := os.WriteFile(other, []byte("synthetic-other\n"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(link); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(other, link); err != nil {
			t.Fatal(err)
		}
	case "dangling-target":
		if err := os.Remove(native); err != nil {
			t.Fatal(err)
		}
	case "metadata-unreadable":
		req.passthroughLstat = func(string) (os.FileInfo, error) { return nil, os.ErrPermission }
	case "target-unreadable":
		req.museTargetStat = func(string) (os.FileInfo, error) { return nil, os.ErrPermission }
	default:
		t.Fatalf("undriven state %q", tc.State)
	}
	beforeNative, nativeErr := os.ReadFile(native)
	beforeLink, linkErr := os.ReadFile(link)
	beforeText, _ := os.Readlink(link)
	statusReq := statusRequest(fx)
	statusReq.passthroughLstat = req.passthroughLstat
	statusReq.passthroughReadlink = req.passthroughReadlink
	statusReq.museTargetStat = req.museTargetStat
	status, err := StatusOf(statusReq)
	if err != nil {
		t.Fatal(err)
	}
	row := findHome(status, fx.profile, envregistry.Muse)
	if row == nil {
		t.Fatal("Muse status row missing")
	}
	reasons := strings.Join(row.Findings, "; ")
	if tc.Expected.Status == "current" {
		if !row.Current {
			t.Fatalf("status non-current: %s", reasons)
		}
	} else if row.Current || !strings.Contains(reasons, tc.Expected.Status) {
		t.Fatalf("status %q, want %s", reasons, tc.Expected.Status)
	}
	if tc.Expected.Status == envregistry.DiagPassthroughUnreadable && strings.Contains(reasons, envregistry.DiagPassthroughDetached) {
		t.Fatalf("unreadable classified detached: %s", reasons)
	}
	result, err := Resolve(req)
	if tc.Expected.Diagnostic == "" {
		if err != nil {
			t.Fatal(err)
		}
	} else if err == nil || !strings.HasPrefix(err.Error(), tc.Expected.Diagnostic+":") {
		t.Fatalf("error %v, want %s", err, tc.Expected.Diagnostic)
	}
	emitted := result != nil && len(result.Document) > 0
	if emitted != tc.Expected.Emit {
		t.Fatalf("fragment emitted=%v, want %v", emitted, tc.Expected.Emit)
	}
	afterNative, afterNativeErr := os.ReadFile(native)
	afterLink, afterLinkErr := os.ReadFile(link)
	if string(beforeNative) != string(afterNative) || (nativeErr == nil) != (afterNativeErr == nil) {
		t.Fatal("native bytes changed")
	}
	if tc.Expected.Action == "relink" {
		if target, err := os.Readlink(link); err != nil || target != native {
			t.Fatalf("not relinked: %s %v", target, err)
		}
	} else {
		afterText, _ := os.Readlink(link)
		if beforeText != afterText || string(beforeLink) != string(afterLink) || (linkErr == nil) != (afterLinkErr == nil) {
			t.Fatal("managed credential changed")
		}
	}
	if tc.Expected.BytesChanged {
		t.Fatal("vector unexpectedly permits credential changes")
	}
}

func TestMuseForkRefusedRegression(t *testing.T) {
	for _, repair := range []bool{false, true} {
		tc := museVector{State: "temp-rename-fork", Repair: repair}
		tc.Expected.Status = envregistry.DiagPassthroughDetached
		tc.Expected.Diagnostic = DiagHomeStale
		if repair {
			tc.Expected.Diagnostic = envregistry.DiagCredentialConflict
		}
		driveMuseLinkState(t, tc)
	}
}

// Keep a live final auth link below a replaced parent: missing auth would
// exercise a different refusal and leave the current-home bypass untested.
func TestMuseAuthParentBoundaryAtResolve(t *testing.T) {
	for _, parent := range []string{"config", "config/muse"} {
		for _, destination := range []string{"outside", "inside"} {
			for _, repair := range []bool{false, true} {
				name := parent + "/" + destination + "/resolve"
				if repair {
					name += "-repair"
				}
				t.Run(name, func(t *testing.T) {
					fx, link, native := museFixture(t)
					home := ManagedHomeDir(fx.home, fx.profile, envregistry.Muse)
					original := filepath.Join(home, filepath.FromSlash(parent))
					moved := filepath.Join(t.TempDir(), "preserved-parent")
					if destination == "inside" {
						moved = filepath.Join(home, "preserved-parent")
					}
					if err := os.Rename(original, moved); err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink(moved, original); err != nil {
						t.Fatal(err)
					}
					if got, err := os.Readlink(link); err != nil || got != native {
						t.Fatalf("fixture must retain live auth link: %s %v", got, err)
					}
					before, err := os.ReadFile(native)
					if err != nil {
						t.Fatal(err)
					}
					preserved := filepath.Join(moved, "preserve.txt")
					if err := os.WriteFile(preserved, []byte("outside-owned\n"), 0600); err != nil {
						t.Fatal(err)
					}
					forbiddenRead := false
					req := fx.request(envregistry.Muse)
					req.Repair = repair
					req.passthroughLstat = func(path string) (os.FileInfo, error) {
						forbiddenRead = true
						return os.Lstat(path)
					}
					result, err := Resolve(req)
					if err == nil || !strings.Contains(err.Error(), DiagWriteWouldFollowLink) {
						t.Fatalf("symlinked auth parent accepted: result=%+v err=%v", result, err)
					}
					if result != nil && len(result.Document) != 0 {
						t.Fatal("refusal emitted fragment")
					}
					if forbiddenRead {
						t.Fatal("auth inspection crossed parent before boundary refusal")
					}
					if got, err := os.ReadFile(native); err != nil || string(got) != string(before) {
						t.Fatalf("native bytes changed: %v", err)
					}
					if got, err := os.ReadFile(preserved); err != nil || string(got) != "outside-owned\n" {
						t.Fatalf("parent bytes changed: %v", err)
					}
					if got, err := os.Readlink(link); err != nil || got != native {
						t.Fatalf("preserved auth link changed: %s %v", got, err)
					}
				})
			}
		}
	}
}

func TestMuseDirectoriesAndIsolation(t *testing.T) {
	fx, _, _ := museFixture(t)
	home := ManagedHomeDir(fx.home, fx.profile, envregistry.Muse)
	for _, dir := range []string{"config/muse", "data/muse", "state", "cache"} {
		info, err := os.Stat(filepath.Join(home, filepath.FromSlash(dir)))
		if err != nil || !info.IsDir() {
			t.Fatalf("missing directory %s: %v", dir, err)
		}
	}
	marker := readManagedMarker(t, fx, envregistry.Muse)
	for _, surface := range []string{"root-context", "system-prompt", "mcp"} {
		if _, ok := marker.Surfaces[surface]; ok {
			t.Fatalf("unverified surface %s emitted", surface)
		}
	}
	req := fx.request(envregistry.Muse)
	req.Machine.Isolation = map[string]map[string]string{fx.profile: {envregistry.Muse: "isolated"}}
	if _, err := Resolve(req); err == nil || !strings.Contains(err.Error(), envregistry.DiagIsolatedUnsupported) {
		t.Fatalf("isolated: %v", err)
	}
}

func TestMuseAuthRepairAdditionalStates(t *testing.T) {
	for _, state := range []string{"empty-directory", "nonempty-directory", "relative-link", "target-directory", "readlink-error", "missing-link-missing-target", "config-parent-link"} {
		t.Run(state, func(t *testing.T) {
			fx, link, native := museFixture(t)
			req := fx.request(envregistry.Muse)
			req.Repair = true
			switch state {
			case "empty-directory", "nonempty-directory":
				if err := os.Remove(link); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(link, 0700); err != nil {
					t.Fatal(err)
				}
				if state == "nonempty-directory" {
					if err := os.WriteFile(filepath.Join(link, "keep"), []byte("preserve"), 0600); err != nil {
						t.Fatal(err)
					}
				}
			case "relative-link":
				if err := os.Remove(link); err != nil {
					t.Fatal(err)
				}
				relative, err := filepath.Rel(filepath.Dir(link), native)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(relative, link); err != nil {
					t.Fatal(err)
				}
			case "target-directory":
				if err := os.Remove(native); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(native, 0700); err != nil {
					t.Fatal(err)
				}
			case "readlink-error":
				req.passthroughReadlink = func(string) (string, error) { return "", os.ErrPermission }
			case "missing-link-missing-target":
				if err := os.Remove(link); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(native); err != nil {
					t.Fatal(err)
				}
			case "config-parent-link":
				config := filepath.Dir(filepath.Dir(link))
				if err := os.RemoveAll(config); err != nil {
					t.Fatal(err)
				}
				outside := t.TempDir()
				if err := os.Mkdir(filepath.Join(outside, "muse"), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, config); err != nil {
					t.Fatal(err)
				}
			}
			result, err := Resolve(req)
			wantSuccess := state == "empty-directory" || state == "relative-link"
			if (err == nil) != wantSuccess {
				t.Fatalf("repair: %v", err)
			}
			if !wantSuccess && result != nil && len(result.Document) > 0 {
				t.Fatal("refusal emitted fragment")
			}
			if wantSuccess {
				if _, err := os.Readlink(link); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestMuseGlobalReconciliationUsesExplicitProvisioning(t *testing.T) {
	adapter, err := envregistry.ByID(envregistry.Muse)
	if err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	participates, err := managedAdapterParticipates(home, "acme", adapter)
	if err != nil || participates {
		t.Fatalf("unprovisioned Muse participated: %v %v", participates, err)
	}
	fx, _, _ := museFixture(t)
	participates, err = managedAdapterParticipates(fx.home, fx.profile, adapter)
	if err != nil || !participates {
		t.Fatalf("provisioned Muse missing: %v %v", participates, err)
	}
}

// Production StatusOf must distinguish an unused optional home from unknown
// state; otherwise adding Muse makes existing healthy profiles non-current.
func TestMuseStatusOptionalProvisioning(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	for _, adapter := range envregistry.Registry {
		if adapter.ID != envregistry.Muse {
			provision(t, fx, adapter.ID, envregistry.DefaultMachineConfig())
		}
	}
	home := ManagedHomeDir(fx.home, fx.profile, envregistry.Muse)
	markerPath := filepath.Join(home, envmarker.Name)
	for _, name := range []string{"absent", "marker-unreadable", "backup-unreadable", "marker-malformed"} {
		t.Run(name, func(t *testing.T) {
			req := statusRequest(fx)
			switch name {
			case "marker-unreadable":
				req.readStateFile = func(path string) (stateread.File, error) {
					if path == markerPath {
						return stateread.File{}, os.ErrPermission
					}
					return stateread.ReadFile(path)
				}
			case "backup-unreadable":
				req.readStateDirectory = func(path string) (stateread.Directory, error) {
					if strings.HasPrefix(path, home+string(filepath.Separator)) {
						return stateread.Directory{}, os.ErrPermission
					}
					return stateread.ReadDir(path)
				}
			case "marker-malformed":
				if err := os.MkdirAll(home, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(markerPath, []byte("{"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			status, err := StatusOf(req)
			if err != nil {
				t.Fatal(err)
			}
			if status.NonCurrent != (name != "absent") {
				t.Fatalf("%s: NonCurrent=%v, homes=%+v", name, status.NonCurrent, status.Homes)
			}
			row := findHome(status, fx.profile, envregistry.Muse)
			if row == nil || row.Provisioned || row.Current {
				t.Fatalf("optional row lost its unprovisioned state: %+v", row)
			}
		})
	}
}

func TestMusePublishedRegistryLayout(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	digest, err := conformancecoverage.SelectedSuiteManifestSHA256()
	if err != nil {
		t.Fatal(err)
	}
	if digest != museSuiteDigest {
		t.Log("selected suite has no Muse registry/layout; rc.13 unchanged")
		return
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "environments-muse.json"))
	if err != nil {
		t.Fatal(err)
	}
	var vector struct {
		Registry struct {
			Environment    string            `json:"environment"`
			Variables      map[string]string `json:"home_variables"`
			ReplaceHOME    bool              `json:"replace_HOME"`
			RootTarget     *string           `json:"root_context_target"`
			RootVerified   bool              `json:"root_context_verified"`
			Skills         string            `json:"skills_target"`
			SkillsVerified bool              `json:"skills_discovery_verified"`
			Version        string            `json:"tool_version"`
			Refresh        string            `json:"refresh_semantics"`
			Gap            string            `json:"isolation_gap"`
		} `json:"registry"`
		Fixtures struct {
			Layout string `json:"layout"`
		} `json:"fixtures"`
	}
	if err := json.Unmarshal(payload, &vector); err != nil {
		t.Fatal(err)
	}
	adapter, err := envregistry.ByID(vector.Registry.Environment)
	if err != nil {
		t.Fatal(err)
	}
	if vector.Registry.ReplaceHOME || vector.Registry.RootTarget != nil || vector.Registry.RootVerified || vector.Registry.SkillsVerified || adapter.RootTarget != "" || adapter.SkillsDir != vector.Registry.Skills || adapter.VerifiedRelease != vector.Registry.Version || vector.Registry.Refresh != "unverified" || vector.Registry.Gap != "foreign-personal-context" {
		t.Fatalf("registry evidence bounds do not match: %+v", vector.Registry)
	}
	fx, _, native := museFixture(t)
	req := fx.request(envregistry.Muse)
	result, err := Resolve(req)
	if err != nil {
		t.Fatal(err)
	}
	var fragment struct {
		Env map[string]string `json:"env"`
	}
	if err := json.Unmarshal(result.Document, &fragment); err != nil {
		t.Fatal(err)
	}
	parent := ManagedHomeDir(fx.home, fx.profile, envregistry.Muse)
	if len(fragment.Env) != len(vector.Registry.Variables) {
		t.Fatal("XDG variable set changed")
	}
	for name, expected := range vector.Registry.Variables {
		if fragment.Env[name] != filepath.Join(parent, filepath.Base(expected)) {
			t.Fatalf("XDG variable %s not at source-defined parent", name)
		}
	}
	raw, err := os.ReadFile(filepath.Join(root, vector.Fixtures.Layout))
	if err != nil {
		t.Fatal(err)
	}
	var layout struct {
		Directories  []string `json:"tool_directories"`
		ManagedAuth  string   `json:"managed_auth"`
		LinkKind     string   `json:"link_kind"`
		Seeds        []string `json:"seeds"`
		HOMEReplaced bool     `json:"HOME_replaced"`
	}
	if err := json.Unmarshal(raw, &layout); err != nil {
		t.Fatal(err)
	}
	if layout.HOMEReplaced || layout.LinkKind != envregistry.StrategyFileLink || len(layout.Seeds) != len(adapter.Seeds) {
		t.Fatalf("layout contract %+v", layout)
	}
	for i, seed := range layout.Seeds {
		if adapter.Seeds[i] != seed {
			t.Fatalf("seed contract %s", seed)
		}
	}
	for _, directory := range layout.Directories {
		info, err := os.Stat(filepath.Join(parent, filepath.FromSlash(directory)))
		if err != nil || !info.IsDir() {
			t.Fatalf("layout directory %s: %v", directory, err)
		}
	}
	target, err := os.Readlink(filepath.Join(parent, filepath.FromSlash(layout.ManagedAuth)))
	if err != nil || target != native {
		t.Fatalf("layout auth link %s: %v", target, err)
	}
	marker := readManagedMarker(t, fx, envregistry.Muse)
	surface := marker.Surfaces["skills"]
	if len(surface.Paths) != 1 || surface.Paths[0] != vector.Registry.Skills+"/myskill" {
		t.Fatalf("skills surface: %+v", surface)
	}
	status, err := StatusOf(statusRequest(fx))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(status.Notes, ";"), vector.Registry.Gap) {
		t.Fatal("standing isolation gap missing")
	}
	t.Log("registry/layout production checks exercised; exec/serve permission mappings remain the separately owned launcher integration gap")
}
