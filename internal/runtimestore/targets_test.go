package runtimestore

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildsource"
	"github.com/relux-works/curator/internal/skillspec"
)

func TestPrepareScriptRuntimeStagesIncompleteReplacementWithoutBuildRoots(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "manager")
	snapshot := filepath.Join(root, "snapshot")
	stage := filepath.Join(root, "private stage")
	lay(t, snapshot, map[string]string{
		"runtime scripts/tool":          "#!/bin/sh\nexit 0\n",
		"runtime scripts/lib/data":      "runtime",
		"build source/cmd/tool/main.go": "package main",
	})
	live := Dir(home, "skill-a", "commit-a")
	lay(t, live, map[string]string{
		"runtime scripts/lib/data": "incomplete",
		"build source/sentinel":    "must remain live until transaction",
	})

	plan, err := PrepareScriptRuntime(stage, ScriptRuntimeSpec{
		Home: home, SkillName: "skill-a", Commit: "commit-a", Snapshot: snapshot,
		RuntimeRoots: []string{"runtime scripts"},
		Commands:     []skillspec.Command{{Name: "tool", Type: "script", UnixPath: "runtime scripts/tool"}},
		Platform:     "unix",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !plan.Replacement || plan.Desired == nil {
		t.Fatalf("incomplete runtime plan = %+v", plan)
	}
	if plan.Desired.LivePath != live || plan.Commands["tool"].ExecutablePath() != filepath.Join(live, "runtime scripts", "tool") {
		t.Fatalf("runtime targets do not select eventual live tree: %+v", plan)
	}
	if _, err := os.Stat(filepath.Join(live, "build source", "sentinel")); err != nil {
		t.Fatalf("planning mutated the live runtime: %v", err)
	}
	if _, err := os.Stat(filepath.Join(plan.Desired.StagedPath, "runtime scripts", "tool")); err != nil {
		t.Fatalf("staged replacement lacks active script: %v", err)
	}
	if _, err := os.Stat(filepath.Join(plan.Desired.StagedPath, "build source")); !os.IsNotExist(err) {
		t.Fatalf("build root leaked into staged runtime: %v", err)
	}
}

func TestPrepareScriptRuntimeReusesOnlyCompleteManagedTree(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "manager")
	snapshot := filepath.Join(root, "snapshot")
	stage := filepath.Join(root, "stage")
	lay(t, snapshot, map[string]string{"scripts/tool": "#!/bin/sh\n"})
	command := skillspec.Command{Name: "tool", Type: "script", UnixPath: "scripts/tool"}
	live := Dir(home, "skill-a", "commit-a")
	lay(t, live, map[string]string{"scripts/tool": "#!/bin/sh\n"})
	if err := os.Chmod(filepath.Join(live, "scripts", "tool"), 0o755); err != nil {
		t.Fatal(err)
	}

	plan, err := PrepareScriptRuntime(stage, ScriptRuntimeSpec{
		Home: home, SkillName: "skill-a", Commit: "commit-a", Snapshot: snapshot,
		RuntimeRoots: []string{"scripts"}, Commands: []skillspec.Command{command}, Platform: "unix",
	})
	if err != nil || plan.Desired != nil || plan.Replacement {
		t.Fatalf("complete runtime not reused: %+v, %v", plan, err)
	}

	lay(t, live, map[string]string{"build/tool": "unexpected"})
	plan, err = PrepareScriptRuntime(stage, ScriptRuntimeSpec{
		Home: home, SkillName: "skill-a", Commit: "commit-a", Snapshot: snapshot,
		RuntimeRoots: []string{"scripts"}, Commands: []skillspec.Command{command}, Platform: "unix",
	})
	if err != nil || plan.Desired == nil || !plan.Replacement {
		t.Fatalf("runtime containing unmanaged build content was reused: %+v, %v", plan, err)
	}
}

func TestPrepareSingleScriptStagesManagedBinWithoutCopyingSnapshotTree(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "manager")
	snapshot := filepath.Join(root, "snapshot")
	lay(t, snapshot, map[string]string{
		"entrypoint.sh":       "#!/bin/sh\nexit 0\n",
		"build/cmd/tool/main": "build source",
	})
	plan, err := PrepareScriptRuntime(filepath.Join(root, "stage"), ScriptRuntimeSpec{
		Home: home, SkillName: "skill-a", Commit: "commit-a", Snapshot: snapshot,
		Commands: []skillspec.Command{{Name: "tool", Type: "script", UnixPath: "entrypoint.sh"}},
		Platform: "unix",
	})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Desired == nil || plan.Replacement || plan.Commands["tool"].RuntimeDir() != Dir(home, "skill-a", "commit-a") ||
		plan.Commands["tool"].ExecutablePath() != filepath.Join(Dir(home, "skill-a", "commit-a"), "bin", "tool") {
		t.Fatalf("single-script runtime plan = %+v", plan)
	}
	stagedCommand := filepath.Join(plan.Desired.StagedPath, "bin", "tool")
	if info, err := os.Stat(stagedCommand); err != nil || info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("staged single command is not executable: %v, %v", info, err)
	}
	if _, err := os.Stat(filepath.Join(plan.Desired.StagedPath, "entrypoint.sh")); !os.IsNotExist(err) {
		t.Fatalf("source path was overloaded as a runtime target: %v", err)
	}
	if _, err := os.Stat(filepath.Join(plan.Desired.StagedPath, "build")); !os.IsNotExist(err) {
		t.Fatalf("snapshot build tree was copied into runtime: %v", err)
	}
}

func TestPrepareScriptRuntimeRejectsInvalidTypedInputs(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "manager")
	snapshot := filepath.Join(root, "snapshot")
	lay(t, snapshot, map[string]string{"scripts/tool": "#!/bin/sh\n"})
	valid := ScriptRuntimeSpec{
		Home: home, SkillName: "skill-a", Commit: "commit-a", Snapshot: snapshot,
		RuntimeRoots: []string{"scripts"},
		Commands:     []skillspec.Command{{Name: "tool", Type: "script", UnixPath: "scripts/tool"}},
		Platform:     "unix",
	}
	for _, testCase := range []struct {
		name   string
		mutate func(*ScriptRuntimeSpec)
	}{
		{name: "build-command", mutate: func(spec *ScriptRuntimeSpec) { spec.Commands[0].Type = "build" }},
		{name: "outside-root", mutate: func(spec *ScriptRuntimeSpec) { spec.Commands[0].UnixPath = "other/tool" }},
		{name: "duplicate-root", mutate: func(spec *ScriptRuntimeSpec) { spec.RuntimeRoots = []string{"scripts", "scripts"} }},
		{name: "unsupported-platform", mutate: func(spec *ScriptRuntimeSpec) { spec.Platform = "plan9" }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			spec := valid
			spec.RuntimeRoots = append([]string(nil), valid.RuntimeRoots...)
			spec.Commands = append([]skillspec.Command(nil), valid.Commands...)
			testCase.mutate(&spec)
			if _, err := PrepareScriptRuntime(filepath.Join(root, "stage", testCase.name), spec); err == nil {
				t.Fatal("invalid script runtime input was accepted")
			}
		})
	}
}

func TestShimTransitionMatrixIsDeterministicAndManagerScoped(t *testing.T) {
	root := t.TempDir()
	artifact := filepath.Join(root, "cache", "artifact")
	lay(t, root, map[string]string{"cache/artifact": "artifact", "live/project/manual": "user-owned"})
	if err := os.Chmod(artifact, 0o755); err != nil {
		t.Fatal(err)
	}
	compiled := compiledTargetFixture(t, artifact, runtime.GOOS)
	script := ScriptTarget{runtimeDir: filepath.Join(root, "runtime"), executable: artifact}
	project := mustManagedShim(t, ProjectShim, filepath.Join(root, "live", "project"), "tool", "unix")
	canonical := mustManagedShim(t, GlobalCanonicalShim, filepath.Join(root, "live", "global"), "tool", "unix")
	forward := mustManagedShim(t, SafeForwardingShim, filepath.Join(root, "live", "user-bin"), "tool", "unix")
	current := []ManagedShim{forward, project, canonical}

	for _, testCase := range []struct {
		name         string
		target       RuntimeTarget
		destinations []ManagedShim
		wantDesired  int
		wantRemoved  []ShimRole
	}{
		{name: "script-to-build", target: compiled, destinations: current, wantDesired: 3},
		{name: "build-to-script", target: script, destinations: current, wantDesired: 3},
		{name: "remove-global-publication", target: script, destinations: []ManagedShim{project}, wantDesired: 1, wantRemoved: []ShimRole{GlobalCanonicalShim, SafeForwardingShim}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			desired := make([]ShimSpec, 0, len(testCase.destinations))
			for _, destination := range testCase.destinations {
				desired = append(desired, ShimSpec{Destination: destination, Target: testCase.target})
			}
			plan, err := StageShimTransition(filepath.Join(root, "stage", testCase.name), desired, current)
			if err != nil {
				t.Fatal(err)
			}
			if len(plan.Desired) != testCase.wantDesired || len(plan.Removals) != len(testCase.wantRemoved) {
				t.Fatalf("transition plan = %+v", plan)
			}
			for index := 1; index < len(plan.Desired); index++ {
				if string(plan.Desired[index-1].Role) > string(plan.Desired[index].Role) {
					t.Fatalf("desired targets are not deterministic: %+v", plan.Desired)
				}
			}
			for index, role := range testCase.wantRemoved {
				if plan.Removals[index].Role != role {
					t.Fatalf("removal order = %+v", plan.Removals)
				}
			}
		})
	}

	removed, err := StageShimTransition(filepath.Join(root, "stage", "removal"), nil, current)
	if err != nil || len(removed.Desired) != 0 || len(removed.Removals) != 3 {
		t.Fatalf("full removal plan = %+v, %v", removed, err)
	}
	if payload, err := os.ReadFile(filepath.Join(root, "live", "project", "manual")); err != nil || string(payload) != "user-owned" {
		t.Fatalf("unmanaged live target changed: %q, %v", payload, err)
	}
}

func TestCompiledShimsStageWithoutLaunchThenPostInstallForwardExactly(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX post-install launch fixture")
	}
	root := t.TempDir()
	marker := filepath.Join(root, "artifact-launched")
	artifactDir := filepath.Join(root, "immutable cache's % Юникод")
	artifact := filepath.Join(artifactDir, "compiled tool")
	if err := os.MkdirAll(artifactDir, 0o755); err != nil {
		t.Fatal(err)
	}
	payload := "#!/bin/sh\n" +
		"printf launched > " + shellQuote(marker) + "\n" +
		"printf 'PATH=<%s>\\n' \"$PATH\"\n" +
		"printf 'ARGS='\n" +
		"for arg do printf '<%s>' \"$arg\"; done\n" +
		"printf '\\n'\n" +
		"exit 37\n"
	if err := os.WriteFile(artifact, []byte(payload), 0o755); err != nil {
		t.Fatal(err)
	}
	compiled := compiledTargetFixture(t, artifact, runtime.GOOS)
	helper := filepath.Join(root, "dependency's % Юникод")
	inherited := filepath.Join(root, "inherited path")
	for _, dir := range []string{helper, inherited} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	destinations := []ManagedShim{
		mustManagedShim(t, SafeForwardingShim, filepath.Join(root, "live", "forward"), "tool", "unix"),
		mustManagedShim(t, ProjectShim, filepath.Join(root, "live", "project"), "tool", "unix"),
		mustManagedShim(t, GlobalCanonicalShim, filepath.Join(root, "live", "global"), "tool", "unix"),
	}
	var desired []ShimSpec
	for _, destination := range destinations {
		desired = append(desired, ShimSpec{Destination: destination, Target: compiled, PathEntries: []string{helper}})
	}
	plan, err := StageShimTransition(filepath.Join(root, "private-stage"), desired, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("artifact launched during staging: %v", err)
	}
	for _, target := range plan.Desired {
		content, err := os.ReadFile(target.StagedPath)
		if err != nil || !strings.Contains(string(content), "exec "+shellQuote(artifact)+" \"$@\"") {
			t.Fatalf("%s shim does not point directly to selected artifact: %v\n%s", target.Role, err, content)
		}
		installFixtureTarget(t, target)
	}

	arguments := []string{"space value", `quote"value`, "percent%value", "Юникод", ""}
	for index, target := range plan.Desired {
		command := exec.Command(target.LivePath, arguments...)
		if index == 0 {
			command.Env = []string{"PATH="}
		} else {
			command.Env = []string{"PATH=" + inherited}
		}
		output, runErr := command.CombinedOutput()
		var exitErr *exec.ExitError
		if !errors.As(runErr, &exitErr) || exitErr.ExitCode() != 37 {
			t.Fatalf("%s exit = %v; output:\n%s", target.Role, runErr, output)
		}
		wantPath := helper
		if index != 0 {
			wantPath += ":" + inherited
		}
		if !strings.Contains(string(output), "PATH=<"+wantPath+">") ||
			!strings.Contains(string(output), `ARGS=<space value><quote"value><percent%value><Юникод><>`) {
			t.Fatalf("%s forwarding output:\n%s", target.Role, output)
		}
	}
}

func TestStagingRejectsLiveOverlapAndUnsafePathEntries(t *testing.T) {
	root := t.TempDir()
	artifact := filepath.Join(root, "artifact")
	if err := os.WriteFile(artifact, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	target := ScriptTarget{runtimeDir: root, executable: artifact}
	shim := mustManagedShim(t, ProjectShim, filepath.Join(root, "live"), "tool", "unix")
	if _, err := StageShimTransition(filepath.Join(root, "live", "stage"), []ShimSpec{{Destination: shim, Target: target}}, nil); err == nil {
		t.Fatal("staging root overlapping a live bin was accepted")
	}
	if _, err := StageShimTransition(filepath.Join(root, "stage"), []ShimSpec{{Destination: shim, Target: target, PathEntries: []string{filepath.Join(root, "bad:path")}}}, nil); err == nil {
		t.Fatal("Unix PATH entry containing a separator was accepted")
	}
}

func TestParseHelperOutputRecoversPayloadFromConsoleNoise(t *testing.T) {
	payload := `{"args":["space value","quote\"value","percent%PATH%value","Юникод",""],"path":"C:\\dependency path %PATH% Юникод"}`
	for name, output := range map[string]string{
		"bare payload":      payload + "\r\n",
		"leading noise":     "The system cannot find the path specified.\r\n" + payload + "\r\n",
		"surrounding noise": "'tool' is not recognized.\r\n" + payload + "\r\nAccess is denied.\r\n",
		"unix line endings": payload + "\n",
	} {
		t.Run(name, func(t *testing.T) {
			decoded, err := parseHelperOutput([]byte(output))
			if err != nil {
				t.Fatalf("parseHelperOutput() error = %v", err)
			}
			args, ok := decoded["args"].([]any)
			wantArgs := []string{"space value", `quote"value`, "percent%PATH%value", "Юникод", ""}
			if !ok || len(args) != len(wantArgs) {
				t.Fatalf("args = %#v", decoded["args"])
			}
			for index, want := range wantArgs {
				if args[index] != want {
					t.Fatalf("arg %d = %#v, want %#v", index, args[index], want)
				}
			}
			if want := `C:\dependency path %PATH% Юникод`; decoded["path"] != want {
				t.Fatalf("path = %#v, want %#v", decoded["path"], want)
			}
		})
	}
}

func TestParseHelperOutputRejectsOutputWithoutPayload(t *testing.T) {
	for name, output := range map[string]string{
		"empty":          "",
		"console only":   "The system cannot find the path specified.\r\n",
		"truncated json": "{\"args\":[\"space value\"\r\n",
		"json array":     "[\"args\"]\r\n",
	} {
		t.Run(name, func(t *testing.T) {
			if decoded, err := parseHelperOutput([]byte(output)); err == nil {
				t.Fatalf("parseHelperOutput() = %#v, want error", decoded)
			}
		})
	}
}

// parseHelperOutput extracts the helper payload from combined wrapper output.
// Wrappers run through cmd.exe, so stderr from the console host can surround the
// single JSON object the helper prints; the payload is the last line that
// decodes into a JSON object.
//
// It lives here rather than beside its only caller in targets_windows_test.go so
// that the parsing contract is compiled and exercised on every platform. The
// `decodeHelperOutput` wrapper stays windows-only because the `unused` linter
// runs on linux, where the calling test is not part of the build.
func parseHelperOutput(output []byte) (map[string]any, error) {
	lines := strings.Split(strings.ReplaceAll(string(output), "\r\n", "\n"), "\n")
	for index := len(lines) - 1; index >= 0; index-- {
		line := strings.TrimSpace(lines[index])
		if !strings.HasPrefix(line, "{") {
			continue
		}
		var decoded map[string]any
		if err := json.Unmarshal([]byte(line), &decoded); err != nil {
			continue
		}
		return decoded, nil
	}
	return nil, fmt.Errorf("helper output carried no JSON payload")
}

func compiledTargetFixture(t *testing.T, artifactPath, goos string) CompiledTarget {
	t.Helper()
	input := buildmeta.Input{
		SchemaVersion: buildmeta.SchemaVersion,
		Driver:        buildmeta.DriverGoV1,
		BuildSource:   buildsource.Identity{Algorithm: buildsource.Algorithm, ContentSHA256: "sha256:" + strings.Repeat("b", 64)},
		BuildRoot:     "build", Command: "tool", SourceDir: "build/cmd/tool",
		Target:    buildmeta.Target{GOOS: goos, GOARCH: runtime.GOARCH, Tuning: nativeTuning(runtime.GOARCH)},
		Toolchain: buildmeta.Toolchain{Algorithm: buildmeta.ToolchainAlgorithm, GoRelpath: buildmeta.ToolchainGoRelpath, GoVersion: "go version fixture", ContentSHA256: "sha256:" + strings.Repeat("c", 64)},
		Policy:    buildmeta.FixedPolicy(),
	}
	info, err := os.Stat(artifactPath)
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := buildmeta.NewReceipt(input, buildmeta.Artifact{Path: "bin/tool" + map[bool]string{true: ".exe"}[goos == "windows"], SHA256: "sha256:" + strings.Repeat("d", 64), Size: info.Size()})
	if err != nil {
		t.Fatal(err)
	}
	receiptBytes, err := receipt.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	receiptHash, err := buildmeta.HashReceiptBytes(receiptBytes)
	if err != nil {
		t.Fatal(err)
	}
	platform := "unix"
	if goos == "windows" {
		platform = "windows"
	}
	target, err := CompiledTargetFromCache(buildcache.Result{
		Status: buildcache.Hit, Receipt: receipt, ReceiptBytes: receiptBytes,
		ReceiptHash: receiptHash, ArtifactPath: artifactPath,
	}, platform)
	if err != nil {
		t.Fatal(err)
	}
	return target
}

func nativeTuning(goarch string) map[string]string {
	switch goarch {
	case "386":
		return map[string]string{"GO386": "sse2"}
	case "amd64":
		return map[string]string{"GOAMD64": "v1"}
	case "arm":
		return map[string]string{"GOARM": "7"}
	case "arm64":
		return map[string]string{"GOARM64": "v8.0"}
	case "mips", "mipsle":
		return map[string]string{"GOMIPS": "hardfloat"}
	case "mips64", "mips64le":
		return map[string]string{"GOMIPS64": "hardfloat"}
	case "ppc64", "ppc64le":
		return map[string]string{"GOPPC64": "power8"}
	case "riscv64":
		return map[string]string{"GORISCV64": "rva20u64"}
	case "wasm":
		return map[string]string{"GOWASM": "satconv"}
	default:
		return map[string]string{}
	}
}

func mustManagedShim(t *testing.T, role ShimRole, binDir, command, platform string) ManagedShim {
	t.Helper()
	shim, err := NewManagedShim(role, binDir, command, platform)
	if err != nil {
		t.Fatal(err)
	}
	return shim
}

func installFixtureTarget(t *testing.T, target DesiredTarget) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(target.LivePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(target.StagedPath, target.LivePath); err != nil {
		t.Fatal(err)
	}
}
