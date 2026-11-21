package envprofile

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/transaction"
)

func claudeMemoryPublicationFixture(t *testing.T) (*operation, *ResolveRequest, string) {
	t.Helper()
	req := &ResolveRequest{Home: t.TempDir(), Profile: "example", EnvID: envregistry.ClaudeCode, LaunchDir: t.TempDir()}
	root := ManagedHomeDir(req.Home, req.Profile, req.EnvID)
	if err := privatedir.MakeAll(root); err != nil {
		t.Fatalf("prepare private Claude memory root: %v", err)
	}
	if err := privatedir.Validate(root); err != nil {
		t.Fatalf("Claude memory fixture root is not private: %v", err)
	}
	op, err := beginOperation(req.Home)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := op.close(); err != nil {
			t.Error(err)
		}
	})
	return op, req, root
}

// Managed state uses slash-separated paths on every host; host-native
// backslashes and traversal remain refused without creating any state.
func TestClaudeMemoryStatePathCompatibilityAndRefusal(t *testing.T) {
	req := &ResolveRequest{Profile: "fixture", EnvID: envregistry.ClaudeCode}
	rel := claudeMemoryStateRel(req)
	root := t.TempDir()
	const want = "state/claude-memory/v1/16343367258ed8116b776c3781910cfca9f631f4b473a0a57051080e93fd22cf.json"
	if rel != want {
		t.Fatalf("managed state path changed: got %q, want %q", rel, want)
	}
	valid := []struct {
		name string
		path string
	}{
		{"producer", rel},
		{"managed-join", managedJoin("", rel)},
	}
	for _, tc := range valid {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := managedPath(root, tc.path, false); err != nil {
				t.Fatalf("legitimate managed state path refused: %q: %v", tc.path, err)
			}
		})
	}
	if !strings.HasPrefix(rel, "state/claude-memory/v1/") || !strings.HasSuffix(rel, ".json") || strings.Contains(rel, "\\") {
		t.Fatalf("state producer emitted a nonportable managed path: %q", rel)
	}
	invalid := []struct {
		name string
		path string
	}{
		{"host-native-separators", strings.ReplaceAll(rel, "/", "\\")},
		{"parent-traversal", strings.Replace(rel, "/claude-memory/", "/../", 1)},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := managedPath(root, tc.path, false); err == nil {
				t.Fatalf("forbidden managed state path admitted: %q", tc.path)
			}
			if _, err := os.Lstat(filepath.Join(root, "state")); !os.IsNotExist(err) {
				t.Fatalf("path refusal created managed state: %v", err)
			}
		})
	}
}

// Omitted LaunchDir uses the real CWD for memory policy, without granting CWD
// project trust. Explicit CWD has the same policy; a relative directory refuses
// without replacing settings, ownership, seed, marker or composed policy.
func TestClaudeAncestorSettingsOmittedLaunchDirectory(t *testing.T) {
	for _, form := range []string{envregistry.FormReferenced, envregistry.FormMonolithic} {
		t.Run(form, func(t *testing.T) {
			_, req, seed := claudeRepairFixture(t, form)
			t.Chdir(t.TempDir())
			cwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			req.LaunchDir = ""
			if _, err := Resolve(req); err != nil {
				t.Fatalf("omitted directory: %v", err)
			}
			home := filepath.Dir(seed)
			settings := filepath.Join(home, "settings.json")
			implicit := readClaudeState(t, settings)
			var document struct {
				Projects map[string]json.RawMessage `json:"projects"`
			}
			if err := json.Unmarshal(readClaudeState(t, seed), &document); err != nil {
				t.Fatal(err)
			}
			if _, exists := document.Projects[cwd]; exists {
				t.Fatal("memory-policy fallback granted CWD project trust")
			}
			req.LaunchDir = cwd
			if _, err := Resolve(req); err != nil {
				t.Fatalf("explicit directory: %v", err)
			}
			if !bytes.Equal(implicit, readClaudeState(t, settings)) {
				t.Fatal("omitted and explicit CWD have different memory policies")
			}
			type snapshot struct {
				data []byte
				info os.FileInfo
			}
			before := map[string]snapshot{}
			for _, path := range []string{settings, filepath.Join(req.Home, claudeMemoryStateRel(&req)), seed, filepath.Join(home, ".agent-environment.json"), filepath.Join(home, "CLAUDE.md")} {
				info, err := os.Stat(path)
				if err != nil {
					t.Fatal(err)
				}
				before[path] = snapshot{readClaudeState(t, path), info}
			}
			req.LaunchDir = "relative"
			if _, err := Resolve(req); err == nil || !strings.Contains(err.Error(), envregistry.DiagSeedUnreadable+": claude launch directory must be absolute") {
				t.Fatalf("relative directory did not specifically refuse: %v", err)
			}
			for path, old := range before {
				info, err := os.Stat(path)
				if err != nil || !os.SameFile(old.info, info) || old.info.Mode() != info.Mode() || !bytes.Equal(old.data, readClaudeState(t, path)) {
					t.Fatalf("relative-directory refusal changed %s: %v", filepath.Base(path), err)
				}
			}
		})
	}
}

// A launch-root change refreshes generated exclusions through public Resolve;
// an ancestor becoming the project root is no longer excluded. User choices
// survive, and verification diagnoses an existing missing ancestor by name.
func TestClaudeAncestorSettingsAcrossLaunchRoots(t *testing.T) {
	for _, form := range []string{envregistry.FormReferenced, envregistry.FormMonolithic} {
		t.Run(form, func(t *testing.T) {
			_, req, seed := claudeRepairFixture(t, form)
			ancestor := filepath.Join(t.TempDir(), "ancestor")
			project := filepath.Join(ancestor, "project")
			if err := os.MkdirAll(project, 0o700); err != nil {
				t.Fatal(err)
			}
			loader := filepath.Join(ancestor, "CLAUDE.md")
			if err := os.WriteFile(loader, []byte("fixture ancestor"), 0o600); err != nil {
				t.Fatal(err)
			}
			req.LaunchDir, req.Repair = project, false
			if _, err := Resolve(req); err == nil || !strings.Contains(err.Error(), loader) {
				t.Fatalf("missing ancestor was not diagnosed by path: %v", err)
			}
			req.Repair = true
			if _, err := Resolve(req); err != nil {
				t.Fatal(err)
			}
			settings := filepath.Join(filepath.Dir(seed), "settings.json")
			hasLoader := func() bool {
				t.Helper()
				var document struct {
					Excludes []string `json:"claudeMdExcludes"`
				}
				if err := json.Unmarshal(readClaudeState(t, settings), &document); err != nil {
					t.Fatal(err)
				}
				for _, pattern := range document.Excludes {
					if pattern == filepath.ToSlash(loader) {
						return true
					}
				}
				return false
			}
			if !hasLoader() {
				t.Fatal("ancestor loader is not excluded")
			}
			req.LaunchDir = ancestor
			if _, err := Resolve(req); err != nil {
				t.Fatal(err)
			}
			if hasLoader() {
				t.Fatal("old ancestor exclusion now hides project policy")
			}
			req.Repair = false
			if _, err := Resolve(req); err != nil {
				t.Fatalf("satisfied policy was not stable: %v", err)
			}
		})
	}
}

// Malformed ownership refuses before home repair changes settings or composed
// policy. Restoring the valid record is the positive recovery control.
func TestClaudeAncestorOwnershipRefusalPreservesSettings(t *testing.T) {
	_, req, seed := claudeRepairFixture(t, envregistry.FormReferenced)
	statePath := filepath.Join(req.Home, claudeMemoryStateRel(&req))
	valid := readClaudeState(t, statePath)
	settingsPath := filepath.Join(filepath.Dir(seed), "settings.json")
	settings := readClaudeState(t, settingsPath)
	rootPath := filepath.Join(filepath.Dir(seed), "CLAUDE.md")
	root := readClaudeState(t, rootPath)
	if err := os.WriteFile(statePath, []byte(`{"version":77,"owned_patterns":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Resolve(req); err == nil || !strings.Contains(err.Error(), "ownership schema") {
		t.Fatalf("invalid ownership was not specifically refused: %v", err)
	}
	if !bytes.Equal(settings, readClaudeState(t, settingsPath)) || !bytes.Equal(root, readClaudeState(t, rootPath)) {
		t.Fatal("invalid ownership changed managed policy")
	}
	if err := os.WriteFile(statePath, valid, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Resolve(req); err != nil {
		t.Fatalf("valid ownership restoration failed: %v", err)
	}
}

// A fault after settings installation but before ownership installation rolls
// both targets back. The same policy then publishes without the fault. This
// exercises real journal ordering, not a mock pair of independent file writes.
func TestClaudeAncestorSettingsTransactionRollback(t *testing.T) {
	op, req, root := claudeMemoryPublicationFixture(t)
	settingsPath := filepath.Join(root, "settings.json")
	before := []byte(`{"operator":true}`)
	if err := os.WriteFile(settingsPath, before, 0o600); err != nil {
		t.Fatal(err)
	}
	statePath := filepath.Join(req.Home, claudeMemoryStateRel(req))
	injected := errors.New("fixture ownership install interruption")
	settingsInstalled, faultReached := false, false
	var err error
	op.engine, err = newOperationEngine(req.Home, transaction.WithHooks(transaction.Hooks{
		Observe: func(event transaction.Event) {
			if event.Point == transaction.PointAfterInstall && event.LivePath == settingsPath {
				settingsInstalled = true
			}
		},
		Fault: func(event transaction.Event) error {
			if event.Point == transaction.PointBeforeInstall && event.LivePath == statePath {
				faultReached = true
				return injected
			}
			return nil
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if err := publishClaudeAncestorSettings(op, req); !errors.Is(err, injected) || !settingsInstalled || !faultReached {
		t.Fatalf("intended ordered interruption not reached: %v; installed=%v, fault=%v", err, settingsInstalled, faultReached)
	}
	if !bytes.Equal(readClaudeState(t, settingsPath), before) {
		t.Fatal("interrupted publication left replacement settings")
	}
	if _, err := os.Lstat(statePath); !os.IsNotExist(err) {
		t.Fatalf("interrupted publication left ownership state: %v", err)
	}
	op.engine, err = newOperationEngine(req.Home)
	if err != nil {
		t.Fatal(err)
	}
	if err := publishClaudeAncestorSettings(op, req); err != nil {
		t.Fatalf("uninterrupted positive control failed: %v", err)
	}
	plan, err := prepareClaudeAncestorSettings(req)
	if err != nil || plan.settings.payload != nil || plan.statePayload != nil {
		t.Fatalf("successful pair is not satisfied: %v", err)
	}
}

// An unchanged ownership record is a frozen transaction input even when only
// settings need replacement. A real external edit after preparation must refuse
// before settings change, preserving the external record; the no-edit control
// publishes normally. This covers the engine boundary, not filesystem CAS.
func TestClaudeAncestorSettingsFrozenOwnership(t *testing.T) {
	for _, changed := range []bool{false, true} {
		name := "unchanged"
		if changed {
			name = "changed"
		}
		t.Run(name, func(t *testing.T) {
			op, req, root := claudeMemoryPublicationFixture(t)
			if err := publishClaudeAncestorSettings(op, req); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(root, "settings.json")
			before := []byte(`{"custom":true}`)
			if err := os.WriteFile(path, before, 0o600); err != nil {
				t.Fatal(err)
			}
			plan, err := prepareClaudeAncestorSettings(req)
			if err != nil || plan.settings.payload == nil || plan.statePayload != nil {
				t.Fatalf("fixture does not isolate a frozen record: %v", err)
			}
			statePath := filepath.Join(req.Home, claudeMemoryStateRel(req))
			external := []byte("{\"version\":1,\"owned_patterns\":[]}\n")
			var editErr error
			reached := false
			op.engine, err = newOperationEngine(req.Home, transaction.WithHooks(transaction.Hooks{Observe: func(event transaction.Event) {
				if event.Point == transaction.PointPrepared {
					reached = true
					if changed {
						editErr = os.WriteFile(statePath, external, 0o600)
					}
				}
			}}))
			if err != nil {
				t.Fatal(err)
			}
			err = publishClaudeAncestorSettings(op, req)
			if !reached || editErr != nil {
				t.Fatalf("intended edit boundary not reached: %v", editErr)
			}
			if changed {
				if err == nil || !strings.Contains(err.Error(), "transaction_input_changed") {
					t.Fatalf("changed frozen ownership was not refused: %v", err)
				}
				if !bytes.Equal(readClaudeState(t, path), before) || !bytes.Equal(readClaudeState(t, statePath), external) {
					t.Fatal("frozen-input refusal overwrote settings or external record")
				}
			} else if err != nil {
				t.Fatalf("unchanged frozen input refused: %v", err)
			}
		})
	}
}

// Omitting the frozen-input pin must fail for the changed record while leaving
// the unchanged positive control green; an overlay/setup failure is not proof.
func TestClaudeAncestorFrozenOwnershipControl(t *testing.T) {
	output, code := claudePublicationOverlay(t, "claude_ancestor_settings.go", "tx.ReadInputs[file.path] = file.digest", "// frozen input deliberately omitted", "^TestClaudeAncestorSettingsFrozenOwnership$", false)
	if code != 1 || !strings.Contains(output, "--- FAIL: TestClaudeAncestorSettingsFrozenOwnership/changed ") || !strings.Contains(output, "--- PASS: TestClaudeAncestorSettingsFrozenOwnership/unchanged ") || !strings.Contains(output, "changed frozen ownership was not refused") {
		t.Fatalf("frozen-input plant failed outside its intended witness: exit %d", code)
	}
}

// Successful memory publication participates in the enclosing global rollback:
// unchanged surfaces restore exact settings bytes and remove newly owned state.
// Foreign settings/state bytes or a narrower parent mode refuse before either
// target changes; this exercises compensation, not filesystem CAS.
func TestClaudeAncestorSettingsGlobalCompensation(t *testing.T) {
	for _, change := range []string{"none", "settings", "ownership", "parent-mode"} {
		t.Run(change, func(t *testing.T) {
			op, req, root := claudeMemoryPublicationFixture(t)
			settingsPath := filepath.Join(root, "settings.json")
			before := []byte(`{"operator":true}`)
			if err := os.WriteFile(settingsPath, before, 0600); err != nil {
				t.Fatal(err)
			}
			op.undo = &globalUndo{entries: map[string]*globalImage{}, managerHome: req.Home}
			if err := publishClaudeAncestorSettings(op, req); err != nil {
				t.Fatal(err)
			}
			statePath := filepath.Join(req.Home, claudeMemoryStateRel(req))
			parent := filepath.Dir(statePath)
			wantConflict := "global profile compensation postimage changed"
			var wantParentMode os.FileMode
			switch change {
			case "settings", "ownership":
				path := settingsPath
				if change == "ownership" {
					path = statePath
				}
				if err := os.WriteFile(path, []byte("operator replacement\n"), 0600); err != nil {
					t.Fatal(err)
				}
			case "parent-mode":
				beforeMode, err := os.Stat(parent)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.Chmod(parent, 0500); err != nil {
					t.Fatal(err)
				}
				narrowedMode, err := os.Stat(parent)
				if err != nil {
					t.Fatal(err)
				}
				wantParentMode = narrowedMode.Mode().Perm()
				if wantParentMode == beforeMode.Mode().Perm() || wantParentMode&0222 != 0 {
					t.Fatalf("parent-mode fixture did not create a distinct non-writable mode: before=%v after=%v", beforeMode.Mode(), narrowedMode.Mode())
				}
				t.Cleanup(func() { _ = os.Chmod(parent, 0700) })
				wantConflict = "compensation parent postimage changed"
			}
			settings, state := readClaudeState(t, settingsPath), readClaudeState(t, statePath)
			settingsInfo, err := os.Stat(settingsPath)
			if err != nil {
				t.Fatal(err)
			}
			stateInfo, err := os.Stat(statePath)
			if err != nil {
				t.Fatal(err)
			}
			err = op.undo.restore()
			if change == "none" {
				if err != nil {
					t.Fatalf("unchanged memory publication blocked global compensation: %v", err)
				}
				if !bytes.Equal(readClaudeState(t, settingsPath), before) {
					t.Fatal("global compensation lost original settings")
				}
				if _, err := os.Lstat(statePath); !os.IsNotExist(err) {
					t.Fatalf("global compensation retained newly owned state: %v", err)
				}
				if _, err := os.Lstat(filepath.Join(req.Home, "state", "claude-memory")); !os.IsNotExist(err) {
					t.Fatalf("global compensation retained newly created memory parents: %v", err)
				}
				return
			}
			if !errors.Is(err, transaction.ErrImplementationCorruption) || !strings.Contains(err.Error(), wantConflict) {
				t.Fatalf("foreign memory postimage did not specifically refuse: %v", err)
			}
			if !bytes.Equal(readClaudeState(t, settingsPath), settings) || !bytes.Equal(readClaudeState(t, statePath), state) {
				t.Fatal("compensation refusal changed settings or ownership state")
			}
			for path, before := range map[string]os.FileInfo{settingsPath: settingsInfo, statePath: stateInfo} {
				after, err := os.Stat(path)
				if err != nil || !os.SameFile(before, after) || before.Mode() != after.Mode() || !before.ModTime().Equal(after.ModTime()) {
					t.Fatalf("compensation refusal rewrote a target: %s: %v", filepath.Base(path), err)
				}
			}
			if change == "parent-mode" {
				info, err := os.Stat(parent)
				if err != nil {
					t.Fatalf("compensation refusal removed the operator parent mode: %v", err)
				}
				if got := info.Mode().Perm(); got != wantParentMode || got&0222 != 0 {
					t.Fatalf("compensation refusal overwrote the operator parent mode: got %v, want %v", got, wantParentMode)
				}
			}
		})
	}
}

// Omitting this publisher's compensation checkpoint must fail at the unchanged
// rollback assertion. A setup error or arbitrary child failure is insufficient;
// the unmodified paired test above supplies the passing control.
func TestClaudeAncestorSettingsCompensationCheckpointControl(t *testing.T) {
	output, code := claudePublicationOverlay(t, "claude_ancestor_settings.go",
		"return op.recordPostimages(plan.settingsPath, plan.statePath)", "return nil",
		"^TestClaudeAncestorSettingsGlobalCompensation$/^none$", false)
	if code != 1 || !strings.Contains(output, "--- FAIL: TestClaudeAncestorSettingsGlobalCompensation/none ") || !strings.Contains(output, "unchanged memory publication blocked global compensation") || !strings.Contains(output, "compensation parent postimage changed") {
		t.Fatalf("compensation checkpoint plant failed outside its intended witness: exit %d", code)
	}
}
