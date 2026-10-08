package envprofile

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

type dotfileManagerVector struct {
	ProtocolVersion    string                     `json:"protocol_version"`
	SchemaVersion      int                        `json:"schema_version"`
	Capability         string                     `json:"capability"`
	CapabilityRevision int                        `json:"capability_revision"`
	Rule               string                     `json:"rule"`
	Diagnostic         string                     `json:"diagnostic"`
	Managers           []string                   `json:"managers"`
	Platforms          []string                   `json:"platforms"`
	Table              []dotfileManagerVectorRow  `json:"table"`
	Cases              []dotfileManagerVectorCase `json:"cases"`
}

type dotfileManagerVectorRow struct {
	Manager string                    `json:"manager"`
	MacOS   *dotfileManagerVectorCell `json:"macos"`
	Linux   *dotfileManagerVectorCell `json:"linux"`
	Windows *dotfileManagerVectorCell `json:"windows"`
}

type dotfileManagerVectorCell struct {
	Base string `json:"base"`
	Leaf string `json:"leaf"`
}

type dotfileManagerVectorCase struct {
	Name     string             `json:"name"`
	Platform string             `json:"platform"`
	Home     string             `json:"home"`
	Env      map[string]string  `json:"env"`
	States   map[string]*string `json:"states"`
	Expected struct {
		Blocks       bool    `json:"blocks"`
		NamesManager *string `json:"names_manager"`
		Notice       *string `json:"notice"`
		ResolvedPath *string `json:"resolved_path"`
	} `json:"expected"`
}

func TestDotfileManagerVectors(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	vectorPath := filepath.Join(root, "vectors", "environments-dotfile-managers.json")
	payload, err := os.ReadFile(vectorPath) // #nosec G304 -- explicit conformance input
	if errors.Is(err, os.ErrNotExist) {
		t.Skip("the supplied conformance root is a pre-revision root")
	}
	if err != nil {
		t.Fatal(err)
	}

	var vectors dotfileManagerVector
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatalf("decode %s: %v", vectorPath, err)
	}
	if vectors.SchemaVersion != 1 ||
		vectors.Capability != "agent-environments" ||
		vectors.CapabilityRevision != 1 ||
		vectors.Diagnostic != DiagForeignSuspect {
		t.Fatalf("unexpected vector identity: schema=%d capability=%q revision=%d diagnostic=%q",
			vectors.SchemaVersion, vectors.Capability, vectors.CapabilityRevision, vectors.Diagnostic)
	}
	if vectors.Rule != "environments section 9.5 dotfile-manager state table" {
		t.Fatalf("unexpected vector rule %q", vectors.Rule)
	}
	assertDotfileManagerVectorTable(t, vectors)
	if len(vectors.Cases) == 0 {
		t.Fatal("vectors/environments-dotfile-managers.json publishes no cases")
	}

	platform, ok := dotfilePlatformForGOOS(runtime.GOOS)
	if !ok {
		t.Fatalf("unsupported test platform %q", runtime.GOOS)
	}
	ran := 0
	for _, tc := range vectors.Cases {
		if tc.Platform != platform {
			continue
		}
		tc := tc
		ran++
		t.Run(tc.Name, func(t *testing.T) {
			runDotfileManagerVectorCase(t, platform, tc)
		})
	}
	if ran == 0 {
		t.Fatalf("the vector publishes no cases for native platform %s", platform)
	}
	t.Logf("executed %d of %d vector cases on %s; other platform cases run on their native CI lanes", ran, len(vectors.Cases), platform)
}

func assertDotfileManagerVectorTable(t *testing.T, vectors dotfileManagerVector) {
	t.Helper()
	wantManagers := make([]string, 0, len(dotfileStateTable))
	wantTable := make([]dotfileManagerVectorRow, 0, len(dotfileStateTable))
	for _, row := range dotfileStateTable {
		wantManagers = append(wantManagers, row.manager)
		wantTable = append(wantTable, dotfileManagerVectorRow{
			Manager: row.manager,
			MacOS:   vectorCell(row.macOS),
			Linux:   vectorCell(row.linux),
			Windows: vectorCell(row.windows),
		})
	}
	if !equalStringSlices(vectors.Managers, wantManagers) {
		t.Fatalf("vector managers %v, want spec order %v", vectors.Managers, wantManagers)
	}
	if !equalStringSlices(vectors.Platforms, []string{"macos", "linux", "windows"}) {
		t.Fatalf("vector platforms %v, want [macos linux windows]", vectors.Platforms)
	}
	if len(vectors.Table) != len(wantTable) {
		t.Fatalf("vector table has %d rows, want %d", len(vectors.Table), len(wantTable))
	}
	for index := range wantTable {
		if !reflect.DeepEqual(vectors.Table[index], wantTable[index]) {
			t.Fatalf("vector table row %d = %+v, want %+v", index+1, vectors.Table[index], wantTable[index])
		}
	}
}

func vectorCell(cell dotfileStateCell) *dotfileManagerVectorCell {
	if cell.baseEnv == "" {
		return nil
	}
	return &dotfileManagerVectorCell{Base: cell.baseEnv, Leaf: cell.leaf}
}

func equalStringSlices(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func runDotfileManagerVectorCase(t *testing.T, platform string, tc dotfileManagerVectorCase) {
	t.Helper()
	if tc.Home == "" {
		t.Fatal("vector home is empty")
	}
	if tc.Expected.Blocks {
		t.Fatal("current dotfile-manager vectors must remain warning-only")
	}

	home := t.TempDir()
	pinHomes(t)
	installIdleProfile(t, home, "dotfile-manager-vector")
	native := claudeHome(t)
	if err := os.MkdirAll(native, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(native, "CLAUDE.md"), []byte("operator-owned\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	operatorHome := t.TempDir()
	pinOperatorHome(t, operatorHome)
	runtimeEnv := setDotfileVectorEnvironment(t, tc.Env, operatorHome)
	for _, row := range dotfileStateTable {
		state, exists := tc.States[row.manager]
		if !exists {
			t.Fatalf("vector case omits state for %s", row.manager)
		}
		prepareDotfileVectorState(t, platform, operatorHome, runtimeEnv, row, state)
	}
	isolateLiveXDGFromUnreadableFixture(t, platform, tc)

	assertDotfileVectorResolvedPath(t, platform, operatorHome, runtimeEnv, tc)
	results, err := UseWithPolicy(home, "dotfile-manager-vector", "", "", false, Policy{Takeover: true})
	if err != nil {
		t.Fatalf("takeover blocked: %v", err)
	}
	var warningForManager bool
	var anyWarning bool
	for _, result := range results {
		if !result.OK {
			t.Fatalf("%s: %s", result.Adapter, result.Detail)
		}
		for _, warning := range result.Warnings {
			if strings.HasPrefix(warning, DiagForeignSuspect+":") {
				anyWarning = true
				if tc.Expected.NamesManager != nil &&
					strings.HasPrefix(warning, DiagForeignSuspect+": "+*tc.Expected.NamesManager+" ") {
					warningForManager = true
				}
			}
		}
	}
	if tc.Expected.Notice == nil {
		if anyWarning {
			t.Fatalf("unexpected dotfile-manager warning in %s: %+v", tc.Name, results)
		}
	} else {
		if *tc.Expected.Notice != DiagForeignSuspect {
			t.Fatalf("notice %q, want %q", *tc.Expected.Notice, DiagForeignSuspect)
		}
		if tc.Expected.NamesManager == nil || !warningForManager {
			t.Fatalf("warning did not name expected manager %v: %+v", tc.Expected.NamesManager, results)
		}
	}
}

func setDotfileVectorEnvironment(t *testing.T, vectorEnv map[string]string, operatorHome string) map[string]string {
	t.Helper()
	values := make(map[string]string, 2)
	for _, name := range []string{"XDG_DATA_HOME", "XDG_CONFIG_HOME"} {
		value := vectorEnv[name]
		if value != "" && filepath.IsAbs(value) {
			value = filepath.Join(operatorHome, strings.ToLower(name)+"-override")
		}
		t.Setenv(name, value)
		values[name] = value
	}
	return values
}

func prepareDotfileVectorState(t *testing.T, platform, home string, env map[string]string, row dotfileStateRow, state *string) {
	t.Helper()
	path, resolved := resolveDotfileStatePath(row.cell(platform), home, func(name string) string { return env[name] })
	if !resolved {
		if state != nil {
			t.Fatalf("none cell for %s has vector state %q", row.manager, *state)
		}
		return
	}
	if state == nil {
		t.Fatalf("path cell for %s has no vector state", row.manager)
	}
	if *state == "absent" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	switch *state {
	case "directory":
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
	case "symlink":
		target := path + "-target"
		if err := os.MkdirAll(target, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, path); err != nil {
			t.Skip("symlink unavailable")
		}
	case "file":
		if err := os.WriteFile(path, []byte("state"), 0o600); err != nil {
			t.Fatal(err)
		}
	case "unreadable":
		parent := filepath.Dir(path)
		if err := os.Chmod(parent, 0); err != nil {
			t.Skip("this host cannot create an unreadable dotfile state path")
		}
		t.Cleanup(func() { _ = os.Chmod(parent, 0o700) })
		_, err := os.Lstat(path)
		if err == nil {
			t.Skip("root ignores directory permissions")
		}
		if !os.IsPermission(err) {
			t.Fatalf("unreadable fixture inspection error %v, want permission denied", err)
		}
	default:
		t.Fatalf("unknown vector location state %q for %s", *state, row.manager)
	}
}

func assertDotfileVectorResolvedPath(t *testing.T, platform, home string, env map[string]string, tc dotfileManagerVectorCase) {
	t.Helper()
	if tc.Expected.NamesManager == nil {
		if tc.Expected.ResolvedPath != nil {
			t.Fatalf("quiet case unexpectedly declares resolved path %q", *tc.Expected.ResolvedPath)
		}
		return
	}
	if tc.Expected.ResolvedPath == nil {
		t.Fatal("present manager case has no resolved_path")
	}
	var row *dotfileStateRow
	for index := range dotfileStateTable {
		if dotfileStateTable[index].manager == *tc.Expected.NamesManager {
			row = &dotfileStateTable[index]
			break
		}
	}
	if row == nil {
		t.Fatalf("expected unknown manager %q", *tc.Expected.NamesManager)
	}
	cell := row.cell(platform)
	path, resolved := resolveDotfileStatePath(cell, home, func(name string) string { return env[name] })
	if !resolved {
		t.Fatalf("%s has no resolved path on %s", row.manager, platform)
	}

	vectorBase := filepath.Join(tc.Home, filepath.FromSlash(cell.defaultBase))
	if value := tc.Env[cell.baseEnv]; value != "" && filepath.IsAbs(value) {
		vectorBase = value
	}
	relative, err := filepath.Rel(vectorBase, *tc.Expected.ResolvedPath)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		t.Fatalf("vector resolved path %q is outside base %q: %v", *tc.Expected.ResolvedPath, vectorBase, err)
	}
	localBase := env[cell.baseEnv]
	if localBase == "" || !filepath.IsAbs(localBase) {
		localBase = filepath.Join(home, filepath.FromSlash(cell.defaultBase))
	}
	want := filepath.Join(localBase, relative)
	if path != want {
		t.Fatalf("resolved path %q, want remapped vector path %q", path, want)
	}
}

// isolateLiveXDGFromUnreadableFixture keeps adapter homes operable when a
// vector "unreadable" fixture sabotages a directory the machine switch also
// needs. The opencode adapter derives its native home from live
// XDG_CONFIG_HOME (falling back to $HOME/.config when it is empty), the
// same root the home-manager state resolves under. Chmodding that root to
// build the unreadable fixture would otherwise fail the opencode entry with
// permission denied and report profile_use_partial for the whole switch —
// collateral unrelated to the dotfile heuristic under test (rev6 gate run
// 36205109346 failed TestDotfileManagerVectors on
// home-manager-linux-unreadable-quiet exactly this way).
//
// The vector-mapped table stays exactly as published: fixtures and
// resolved-path assertions keep using the vector environment, and the live
// hint still probes every row through the real inventory path. Only the
// live process variable moves to a writable sandbox so adapter entries
// remain operable. The quiet expectation is identical whether the row reads
// absent or failed-unknown — neither names a manager nor blocks — and the
// failed-inspection-continues rule stays covered by
// TestForeignManagerHintLstatDiscipline with injected errors. Rows under
// XDG_DATA_HOME need no isolation: no adapter derives its home from it.
func isolateLiveXDGFromUnreadableFixture(t *testing.T, platform string, tc dotfileManagerVectorCase) {
	t.Helper()
	for _, row := range dotfileStateTable {
		if row.cell(platform).baseEnv != "XDG_CONFIG_HOME" {
			continue
		}
		state, exists := tc.States[row.manager]
		if !exists || state == nil || *state != "unreadable" {
			continue
		}
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "xdg-config"))
		return
	}
}

// TestUnreadableDotfileStateKeepsTakeoverQuiet is the regression test for
// the rev6 gate failure (run 36205109346): with XDG variables defaulted,
// the unreadable home-manager fixture chmods an ancestor of the opencode
// adapter fallback home, and the machine takeover must still succeed
// quietly. Without isolateLiveXDGFromUnreadableFixture the switch reports
// profile_use_partial from the opencode entry.
func TestUnreadableDotfileStateKeepsTakeoverQuiet(t *testing.T) {
	platform, ok := dotfilePlatformForGOOS(runtime.GOOS)
	if !ok {
		t.Skip("unsupported test platform")
	}
	home := t.TempDir()
	pinHomes(t)
	installIdleProfile(t, home, "unreadable-keeps-quiet")
	native := claudeHome(t)
	if err := os.MkdirAll(native, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(native, "CLAUDE.md"), []byte("operator-owned\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	operatorHome := t.TempDir()
	pinOperatorHome(t, operatorHome)
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	runtimeEnv := map[string]string{"XDG_DATA_HOME": "", "XDG_CONFIG_HOME": ""}
	var managerRow *dotfileStateRow
	for index := range dotfileStateTable {
		if dotfileStateTable[index].manager == "home-manager" {
			managerRow = &dotfileStateTable[index]
			break
		}
	}
	if managerRow == nil {
		t.Fatal("dotfile table has no home-manager row")
	}
	if _, resolved := resolveDotfileStatePath(managerRow.cell(platform), operatorHome, func(name string) string {
		return runtimeEnv[name]
	}); !resolved {
		t.Skip("this host cannot create the unreadable fixture: home-manager has no resolvable state cell on this platform")
	}
	unreadable := "unreadable"
	absent := "absent"
	prepareDotfileVectorState(t, platform, operatorHome, runtimeEnv, *managerRow, &unreadable)
	isolateLiveXDGFromUnreadableFixture(t, platform, dotfileManagerVectorCase{
		Name:     "unreadable-keeps-takeover-quiet",
		Platform: platform,
		States: map[string]*string{
			"chezmoi":      &absent,
			"home-manager": &unreadable,
			"yadm":         &absent,
			"stow":         nil,
			"dotbot":       nil,
		},
	})
	results, err := UseWithPolicy(home, "unreadable-keeps-quiet", "", "", false, Policy{Takeover: true})
	if err != nil {
		t.Fatalf("takeover over an unreadable dotfile state must stay quiet: %v", err)
	}
	for _, result := range results {
		if !result.OK {
			t.Fatalf("%s: %s", result.Adapter, result.Detail)
		}
		for _, warning := range result.Warnings {
			if strings.HasPrefix(warning, DiagForeignSuspect+":") {
				t.Fatalf("unexpected dotfile-manager warning: %q", warning)
			}
		}
	}
}
