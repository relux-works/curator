package install

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/conformancecoverage"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/staging"
	"github.com/relux-works/curator/internal/transaction"
)

// writeLegacyBuildSkill creates a normal registry skill for the production
// install.Project path. Schema 6 exercises the unchanged local receipt-v1
// lane; schema 7 adds the independently acquired external command.
func writeLegacyBuildSkill(t *testing.T, e *env, schema int, drivers []string) {
	writeLegacyBuildSkillNamed(t, e, "golden-skill", schema, drivers)
}

func writeLegacyBuildSkillNamed(t *testing.T, e *env, name string, schema int, drivers []string) {
	t.Helper()
	dir := filepath.Join(e.skillsRoot, name)
	for _, rel := range []string{"src/cmd/local-helper", "references"} {
		if err := os.MkdirAll(filepath.Join(dir, filepath.FromSlash(rel)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	e.write(dir, "SKILL.md", "---\nname: "+name+"\ndescription: Test\n---\n# Golden skill\n")
	e.write(dir, "references/notes.md", "notes\n")
	e.write(dir, "src/go.mod", "module example.test/golden-skill\n")
	e.write(dir, "src/cmd/local-helper/main.go", "package main\n\nfunc main() {}\n")
	commands := map[string]any{}
	external := false
	for _, driver := range drivers {
		switch driver {
		case "go-v1":
			commands["local-helper"] = map[string]any{"type": "build", "driver": "go-v1", "source_dir": "src/cmd/local-helper"}
		case "go-repository-v1":
			external = true
			commands["golden-tool"] = map[string]any{"type": "build", "driver": "go-repository-v1", "repository": "golden-tools", "target": "golden-tool"}
		default:
			t.Fatalf("unsupported published build driver %q", driver)
		}
	}
	spec := map[string]any{"schema_version": schema, "capabilities": map[string]any{}, "commands": commands}
	if _, hasLocalBuild := commands["local-helper"]; hasLocalBuild {
		spec["build_roots"] = []string{"src"}
	}
	if external {
		spec["build_repositories"] = map[string]any{"golden-tools": map[string]any{
			"git":           "https://github.com/example/golden-tools.git",
			"locked_commit": map[string]any{"object_format": "sha1", "hex": externalLockedCommit},
			"tag":           "v1.4.0",
		}}
	}
	payload, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	e.write(dir, "csk-skill.json", string(payload))
	e.git(dir, "init", "-q", "-b", "main")
	e.git(dir, "add", ".")
	e.git(dir, "commit", "-qm", "init")
	e.git(dir, "tag", "v1")
	e.declare(name)
}

func lifecycleSnapshot(target string) *buildrepo.Snapshot {
	return lifecycleSnapshotWithField(target, "")
}

func lifecycleSnapshotWithField(target, forbiddenField string) *buildrepo.Snapshot {
	descriptor := `{"schema_version":1,"targets":{"` + target + `":{"driver":"go-repository-v1","build_root":"tools","source_dir":"tools/cmd/` + target + `"}}}`
	if forbiddenField != "" {
		descriptor = strings.TrimSuffix(descriptor, "}}}") + `,"` + forbiddenField + `":"bin/forged"}}}`
	}
	files := []buildrepo.File{
		{Path: "skill-build.json", Content: []byte(descriptor)},
		{Path: "tools/cmd/" + target + "/main.go", Content: []byte("package main\n\nfunc main() {}\n")},
		{Path: "tools/go.mod", Content: []byte("module example.test/golden-tools\n")},
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	canonical := []byte("curator-build-source-v1\x00")
	for _, file := range files {
		canonical = append(canonical, 'F')
		canonical = binary.BigEndian.AppendUint64(canonical, uint64(len(file.Path)))
		canonical = append(canonical, file.Path...)
		canonical = binary.BigEndian.AppendUint64(canonical, uint64(len(file.Content)))
		canonical = append(canonical, file.Content...)
	}
	sum := sha256.Sum256(canonical)
	return &buildrepo.Snapshot{ObjectFormat: "sha1", Commit: externalLockedCommit, Files: files, CanonicalBytes: canonical, Digest: "sha256:" + hex.EncodeToString(sum[:]), TagVerified: true}
}

func lifecycleBuildDeps(t *testing.T) (BuildDeps, *fakeBuilder) {
	t.Helper()
	toolchain := &fakeToolchain{t: t, target: testTarget(), toolchain: testToolchain()}
	builder := &fakeBuilder{t: t, failOn: map[string]error{}}
	return BuildDeps{Toolchain: toolchain, Builder: builder}, builder
}

func lifecycleExternalDeps() ExternalDeps {
	return lifecycleExternalDepsWith(func() *buildrepo.Snapshot { return lifecycleSnapshot("golden-tool") })
}

func lifecycleExternalDepsWith(snapshot func() *buildrepo.Snapshot) ExternalDeps {
	return ExternalDeps{
		Acquire: func(context.Context, ExternalSource) (*buildrepo.Snapshot, error) {
			return snapshot(), nil
		},
		Audit: func(context.Context, buildrepo.AuditSubject) error { return nil },
	}
}

func TestLegacyMixedBuildProjectInstallProducesMarkerV3(t *testing.T) {
	e := newEnv(t)
	writeLegacyBuildSkill(t, e, 7, []string{"go-repository-v1", "go-v1"})
	deps, _ := lifecycleBuildDeps(t)
	result := e.install(Options{Build: deps, External: lifecycleExternalDeps()})
	if result.Status != "ok" {
		t.Fatalf("install result = %+v", result)
	}
	installed := filepath.Join(e.project, ".agents", "skills", "golden-skill")
	recorded := marker.Read(installed)
	if recorded == nil || recorded.SchemaVersion != marker.ExternalSchemaVersion {
		t.Fatalf("marker = %+v", recorded)
	}
	if len(recorded.Builds) != 2 {
		t.Fatalf("marker builds = %+v, want local and external commands", recorded.Builds)
	}
	local, localOK := recorded.Builds["local-helper"]
	external, externalOK := recorded.Builds["golden-tool"]
	if !localOK || local.Driver != buildmeta.DriverGoV1 || local.ReceiptSchemaVersion != 1 {
		t.Fatalf("local receipt = %+v, present=%v", local, localOK)
	}
	if !externalOK || external.Driver != "go-repository-v1" || external.ReceiptSchemaVersion != 2 || external.DescriptorTarget != "golden-tool" {
		t.Fatalf("external receipt = %+v, present=%v", external, externalOK)
	}
}

func TestExternalCommandNameCollisionFailsBeforeMutation(t *testing.T) {
	e := newEnv(t)
	writeLegacyBuildSkillNamed(t, e, "golden-skill", 7, []string{"go-repository-v1"})
	writeLegacyBuildSkillNamed(t, e, "second-skill", 7, []string{"go-repository-v1"})
	e.declare("golden-skill", "second-skill")
	deps, builder := lifecycleBuildDeps(t)
	acquisitions := 0
	external := lifecycleExternalDepsWith(func() *buildrepo.Snapshot {
		acquisitions++
		return lifecycleSnapshot("golden-tool")
	})
	before := snapshotState(t, e)
	result := e.install(Options{Build: deps, External: external})
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, "\n"), "command collision for \"golden-tool\"") {
		t.Fatalf("collision install result=%+v", result)
	}
	if acquisitions != 0 || len(builder.calls) != 0 {
		t.Fatalf("collision crossed build boundary: acquisitions=%d builds=%v", acquisitions, builder.calls)
	}
	if err := stateUnchanged(t, before, snapshotState(t, e), "command collision"); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"golden-skill", "second-skill"} {
		if marker.Read(filepath.Join(e.project, ".agents", "skills", name)) != nil {
			t.Fatalf("command collision installed marker for %s", name)
		}
	}
	if _, err := os.Lstat(filepath.Join(e.project, ".agents", "bin", shimName("golden-tool"))); !os.IsNotExist(err) {
		t.Fatalf("command collision published a shim: %v", err)
	}
}

func TestGlobalMixedBuildStagesExternalBeforeLocal(t *testing.T) {
	e := newEnv(t)
	writeLegacyBuildSkill(t, e, 7, []string{"go-repository-v1", "go-v1"})
	e.globalDeclare("golden-skill")
	deps, builder := lifecycleBuildDeps(t)
	result := Global(e.cfg, t.TempDir(), Options{Platform: installPlatform(), Build: deps, External: lifecycleExternalDeps()})
	if result.Status != "ok" {
		t.Fatalf("global mixed install result=%+v", result)
	}
	if !reflect.DeepEqual(builder.calls, []string{"golden-tool", "local-helper"}) {
		t.Fatalf("global mixed stage order=%v", builder.calls)
	}
	recorded := marker.Read(filepath.Join(GlobalRoot(e.home), "skills", "golden-skill"))
	if recorded == nil || recorded.SchemaVersion != marker.ExternalSchemaVersion || len(recorded.Builds) != 2 {
		t.Fatalf("global mixed marker=%+v", recorded)
	}
	if recorded.Builds["golden-tool"].ReceiptSchemaVersion != 2 || recorded.Builds["local-helper"].ReceiptSchemaVersion != 1 {
		t.Fatalf("global mixed receipt versions=%+v", recorded.Builds)
	}
}

type lifecycleMixedCase struct {
	Name                        string   `json:"name"`
	ManifestSchema              int      `json:"manifest_schema"`
	MarkerVersion               int      `json:"marker_version"`
	Drivers                     []string `json:"drivers"`
	ReceiptVersions             []int    `json:"receipt_versions"`
	ExpectedMarker              string   `json:"expected_marker"`
	DeclaredAndEffectiveSources bool     `json:"declared_and_effective_sources"`
}

type externalLifecycleVectors struct {
	MixedBuildCases  []lifecycleMixedCase       `json:"mixed_build_cases"`
	PathShimCases    []lifecyclePathShimCase    `json:"path_shim_cases"`
	SigningCases     []lifecycleSigningCase     `json:"signing_cases"`
	TransactionCases []lifecycleTransactionCase `json:"transaction_cases"`
}

type lifecyclePathShimCase struct {
	Name                          string `json:"name"`
	ArtifactExecutedDuringInstall bool   `json:"artifact_executed_during_install"`
	ForwardArguments              bool   `json:"forward_arguments"`
	PathEntryDerivedByManager     bool   `json:"path_entry_derived_by_manager"`
	PreserveExitStatus            bool   `json:"preserve_exit_status"`
	PreserveInheritedPath         bool   `json:"preserve_inherited_path"`
	ExpectedError                 string `json:"expected_error"`
	ShimPublished                 bool   `json:"shim_published"`
	PriorShimRestored             bool   `json:"prior_shim_restored"`
}

type lifecycleSigningCase struct {
	Name                          string `json:"name"`
	ArtifactExecutedDuringInstall bool   `json:"artifact_executed_during_install"`
	ManagerPostSigning            bool   `json:"manager_post_signing"`
	ExpectedError                 string `json:"expected_error"`
	SignerStarted                 bool   `json:"signer_started"`
	Result                        string `json:"result"`
	Owner                         string `json:"owner"`
}

type lifecycleTransactionCase struct {
	Name                        string `json:"name"`
	FailureAt                   string `json:"failure_at"`
	LiveStateUnchanged          bool   `json:"live_state_unchanged"`
	ConsumerMarkerCommitted     bool   `json:"consumer_marker_committed"`
	Rollback                    string `json:"rollback"`
	PartialCurrentnessForbidden bool   `json:"partial_currentness_forbidden"`
	Recovery                    string `json:"recovery"`
	JournalRetainedIfUncertain  bool   `json:"journal_retained_if_uncertain"`
	GCRetainsJournalRoots       bool   `json:"gc_retains_journal_roots"`
}

func readExternalLifecycleVectors(t *testing.T) (string, externalLifecycleVectors) {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "external-repository-lifecycle.json")) // #nosec G304 -- explicit conformance root
	if err != nil {
		t.Fatal(err)
	}
	var vectors externalLifecycleVectors
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	if len(vectors.MixedBuildCases) == 0 {
		t.Fatal("the authoritative suite publishes no mixed build cases")
	}
	if len(vectors.PathShimCases) == 0 || len(vectors.SigningCases) == 0 || len(vectors.TransactionCases) == 0 {
		t.Fatal("the authoritative suite is missing a lifecycle vector table")
	}
	return root, vectors
}

type expectedMixedMarker struct {
	SchemaVersion      int      `json:"schema_version"`
	SkillSchemaVersion int      `json:"skill_schema_version"`
	Commands           []string `json:"commands"`
	Builds             map[string]struct {
		Driver               string `json:"driver"`
		ReceiptSchemaVersion int    `json:"receipt_schema_version"`
		Substituted          bool   `json:"substituted"`
		DescriptorTarget     string `json:"descriptor_target"`
		Repository           string `json:"repository"`
		DeclaredIdentity     struct {
			Kind  string `json:"kind"`
			Value string `json:"value"`
		} `json:"declared_identity"`
		DeclaredLockedCommit struct {
			ObjectFormat string `json:"object_format"`
			Hex          string `json:"hex"`
		} `json:"declared_locked_commit"`
		DeclaredTag       string `json:"declared_tag"`
		EffectiveIdentity struct {
			Kind  string `json:"kind"`
			Value string `json:"value"`
		} `json:"effective_identity"`
		ObjectFormat    string `json:"object_format"`
		Commit          string `json:"commit"`
		ExecutionPolicy string `json:"execution_policy"`
		ArtifactPath    string `json:"artifact_path"`
	} `json:"builds"`
}

type expectedMixedPlan struct {
	SchemaVersion       int `json:"schema_version"`
	MarkerSchemaVersion int `json:"marker_schema_version"`
	Commands            []struct {
		Name                 string `json:"name"`
		Driver               string `json:"driver"`
		ReceiptSchemaVersion int    `json:"receipt_schema_version"`
	} `json:"commands"`
	PublicationOrder []string `json:"publication_order"`
}

func compareMixedMarker(t *testing.T, path string, got *marker.Marker) {
	t.Helper()
	payload, err := os.ReadFile(path) // #nosec G304 -- path is a published expected fixture
	if err != nil {
		t.Fatal(err)
	}
	var want expectedMixedMarker
	if err := json.Unmarshal(payload, &want); err != nil {
		t.Fatal(err)
	}
	if got.SchemaVersion != want.SchemaVersion || got.SkillSchemaVersion != want.SkillSchemaVersion || !reflect.DeepEqual(got.Commands, want.Commands) {
		t.Fatalf("marker header does not match expected fixture: got version=%d skill=%d commands=%v, want version=%d skill=%d commands=%v", got.SchemaVersion, got.SkillSchemaVersion, got.Commands, want.SchemaVersion, want.SkillSchemaVersion, want.Commands)
	}
	if len(got.Builds) != len(want.Builds) {
		t.Fatalf("marker builds=%v, expected fixture builds=%v", got.Builds, want.Builds)
	}
	for name, expected := range want.Builds {
		actual, ok := got.Builds[name]
		if !ok || actual.Driver != expected.Driver || actual.ReceiptSchemaVersion != expected.ReceiptSchemaVersion || actual.Substituted != expected.Substituted || actual.DescriptorTarget != expected.DescriptorTarget ||
			(expected.Repository != "" && actual.Repository != expected.Repository) ||
			(expected.DeclaredIdentity.Value != "" && (actual.DeclaredIdentity == nil || actual.DeclaredIdentity.Kind != expected.DeclaredIdentity.Kind || actual.DeclaredIdentity.Value != expected.DeclaredIdentity.Value)) ||
			(expected.DeclaredLockedCommit.Hex != "" && (actual.DeclaredLockedCommit == nil || actual.DeclaredLockedCommit.ObjectFormat != expected.DeclaredLockedCommit.ObjectFormat || actual.DeclaredLockedCommit.Hex != expected.DeclaredLockedCommit.Hex)) ||
			(expected.DeclaredTag != "" && actual.DeclaredTag != expected.DeclaredTag) ||
			(expected.EffectiveIdentity.Value != "" && (actual.EffectiveIdentity == nil || actual.EffectiveIdentity.Kind != expected.EffectiveIdentity.Kind || actual.EffectiveIdentity.Value != expected.EffectiveIdentity.Value)) ||
			(expected.ObjectFormat != "" && actual.ObjectFormat != expected.ObjectFormat) ||
			(expected.Commit != "" && actual.Commit != expected.Commit) ||
			(expected.ExecutionPolicy != "" && actual.ExecutionPolicy != expected.ExecutionPolicy) ||
			(expected.ArtifactPath != "" && strings.TrimSuffix(actual.ArtifactPath, ".exe") != expected.ArtifactPath) {
			t.Errorf("marker build %q=%+v, expected fixture=%+v, present=%v", name, actual, expected, ok)
		}
	}
}

func compareMixedPlan(t *testing.T, path string, got *marker.Marker, stagedOrder []string) {
	t.Helper()
	payload, err := os.ReadFile(path) // #nosec G304 -- path is a published expected fixture
	if err != nil {
		t.Fatal(err)
	}
	var want expectedMixedPlan
	if err := json.Unmarshal(payload, &want); err != nil {
		t.Fatal(err)
	}
	if want.SchemaVersion != got.SkillSchemaVersion || want.MarkerSchemaVersion != got.SchemaVersion {
		t.Fatalf("mixed plan schema mismatch: plan=(%d,%d), marker=(%d,%d)", want.SchemaVersion, want.MarkerSchemaVersion, got.SkillSchemaVersion, got.SchemaVersion)
	}
	if len(want.Commands) != len(got.Commands) || len(want.Commands) != len(got.Builds) {
		t.Fatalf("mixed plan commands=%+v, marker commands=%v builds=%v", want.Commands, got.Commands, got.Builds)
	}
	for index, expected := range want.Commands {
		if got.Commands[index] != expected.Name {
			t.Fatalf("mixed plan command order=%v, want command %d=%q", got.Commands, index, expected.Name)
		}
		actual, ok := got.Builds[expected.Name]
		if !ok || actual.Driver != expected.Driver || actual.ReceiptSchemaVersion != expected.ReceiptSchemaVersion {
			t.Fatalf("mixed plan command %q=%+v, expected driver=%q receipt=%d present=%v", expected.Name, actual, expected.Driver, expected.ReceiptSchemaVersion, ok)
		}
	}
	if !reflect.DeepEqual(stagedOrder, func() []string {
		order := make([]string, len(want.Commands))
		for index, command := range want.Commands {
			order[index] = command.Name
		}
		return order
	}()) {
		t.Fatalf("mixed plan stage order=%v, expected commands=%+v", stagedOrder, want.Commands)
	}
	if len(want.PublicationOrder) != len(stagedOrder)+1 || want.PublicationOrder[len(want.PublicationOrder)-1] != "marker-consumer-last" || !reflect.DeepEqual(want.PublicationOrder[:len(stagedOrder)], stagedOrder) {
		t.Fatalf("published mixed plan does not put the marker consumer last: %v", want.PublicationOrder)
	}
}

func TestAuthoritativeMixedBuildCasesUseProjectInstallEntry(t *testing.T) {
	root, vectors := readExternalLifecycleVectors(t)
	conformancecoverage.RunOutcomes(t, "external-repository-lifecycle/mixed_build_cases", vectors.MixedBuildCases,
		func(testCase lifecycleMixedCase) string { return testCase.Name },
		func(t *testing.T, testCase lifecycleMixedCase) conformancecoverage.Observation {
			e := newEnv(t)
			writeLegacyBuildSkill(t, e, testCase.ManifestSchema, testCase.Drivers)
			deps, builder := lifecycleBuildDeps(t)
			var external ExternalDeps
			if containsString(testCase.Drivers, "go-repository-v1") {
				external = lifecycleExternalDeps()
			}
			if strings.Contains(testCase.Name, "substituted") {
				e.write(e.project, "Skillfile.dev.json", `{"schema_version":2,"substitutions":{},"build_repository_substitutions":{"golden-skill":{"golden-tools":{"git":"https://mirror.example.test/skills/golden-tools.git","ref":{"kind":"branch","value":"release/v2"}}}}}`)
			}
			result := e.install(Options{Build: deps, External: external})
			if result.Status != "ok" {
				return conformancecoverage.Observation{FailureReason: "install.Project: " + strings.Join(result.Errors, "; ")}
			}
			installed := filepath.Join(e.project, ".agents", "skills", "golden-skill")
			recorded := marker.Read(installed)
			if recorded == nil || recorded.SchemaVersion != testCase.MarkerVersion || len(recorded.Builds) != len(testCase.Drivers) {
				return conformancecoverage.Observation{FailureReason: "install.Project produced marker/build counts that differ from the published vector"}
			}
			if len(testCase.ReceiptVersions) != len(testCase.Drivers) {
				return conformancecoverage.Observation{FailureReason: "published drivers and receipt versions have different lengths"}
			}
			actualVersions := map[string]int{}
			for _, build := range recorded.Builds {
				version := build.ReceiptSchemaVersion
				if version == 0 && build.Driver == buildmeta.DriverGoV1 && recorded.SchemaVersion < marker.ExternalSchemaVersion {
					version = 1 // legacy markers imply the local receipt-v1 namespace
				}
				actualVersions[build.Driver] = version
			}
			for index, driver := range testCase.Drivers {
				if actualVersions[driver] != testCase.ReceiptVersions[index] {
					return conformancecoverage.Observation{FailureReason: "install.Project receipt version does not match published driver order"}
				}
			}
			if testCase.DeclaredAndEffectiveSources {
				build := recorded.Builds["golden-tool"]
				if build.DeclaredIdentity == nil || build.EffectiveIdentity == nil || build.DeclaredIdentity.Value == build.EffectiveIdentity.Value || !build.Substituted {
					return conformancecoverage.Observation{FailureReason: "install.Project marker does not retain distinct declared and effective substituted sources"}
				}
			}
			if testCase.ExpectedMarker != "" {
				compareMixedMarker(t, filepath.Join(root, filepath.FromSlash(testCase.ExpectedMarker)), recorded)
				if testCase.Name == "schema7-mixed" {
					compareMixedPlan(t, filepath.Join(root, "expected", "external-repository", "mixed-build-plan.json"), recorded, builder.calls)
				}
			}
			return conformancecoverage.Observation{}
		})
}

func TestAuthoritativePathShimCasesUseProjectInstallEntry(t *testing.T) {
	_, vectors := readExternalLifecycleVectors(t)
	conformancecoverage.RunOutcomes(t, "external-repository-lifecycle/path_shim_cases", vectors.PathShimCases,
		func(testCase lifecyclePathShimCase) string { return testCase.Name },
		func(t *testing.T, testCase lifecyclePathShimCase) conformancecoverage.Observation {
			switch testCase.Name {
			case "external-command-shim":
				return exerciseExternalCommandShim(t, testCase)
			case "package-path-entry-rejected":
				return exercisePackagePathEntryRejection(t, testCase)
			case "shim-collision-rolls-back":
				return exerciseShimRollback(t, testCase)
			default:
				return conformancecoverage.Observation{FailureReason: "published path/shim case has no production install binding"}
			}
		})
}

func shellQuoteTest(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func exerciseExternalCommandShim(t *testing.T, testCase lifecyclePathShimCase) conformancecoverage.Observation {
	t.Helper()
	e := newEnv(t)
	writeLegacyBuildSkill(t, e, 7, []string{"go-repository-v1"})
	deps, builder := lifecycleBuildDeps(t)
	ran := filepath.Join(t.TempDir(), "artifact-ran")
	builder.payloads = map[string][]byte{"golden-tool": []byte("#!/bin/sh\nprintf x >> " + shellQuoteTest(ran) + "\nprintf 'path=<%s>' \"$PATH\"\nfor arg in \"$@\"; do printf 'arg=<%s>' \"$arg\"; done\nexit 37\n")}
	result := e.install(Options{Build: deps, External: lifecycleExternalDeps()})
	if result.Status != "ok" {
		return conformancecoverage.Observation{FailureReason: "install.Project: " + strings.Join(result.Errors, "; ")}
	}
	installed := filepath.Join(e.project, ".agents", "skills", "golden-skill")
	recorded := marker.Read(installed)
	if recorded == nil {
		return conformancecoverage.Observation{FailureReason: "install.Project published no marker for the external command"}
	}
	build := recorded.Builds["golden-tool"]
	bin := filepath.Join(e.project, ".agents", "bin")
	shim := filepath.Join(bin, shimName("golden-tool"))
	info, err := os.Stat(shim)
	if err != nil {
		return conformancecoverage.Observation{FailureReason: "manager-derived external command shim is absent: " + err.Error()}
	}
	if installPlatform() == "unix" && info.Mode().Perm()&0o111 == 0 {
		return conformancecoverage.Observation{FailureReason: "manager-derived external command shim is not executable"}
	}
	shimBytes, err := os.ReadFile(shim)
	if err != nil {
		return conformancecoverage.Observation{FailureReason: "read external command shim: " + err.Error()}
	}
	artifact := filepath.Join(e.home, "external-build-cache", buildrepo.ArtifactsDir(build.ReceiptSchemaVersion), strings.TrimPrefix(string(build.CacheKey), "sha256:"), "artifact")
	if !strings.Contains(string(shimBytes), artifact) || !strings.Contains(string(shimBytes), bin) {
		return conformancecoverage.Observation{FailureReason: "external shim target or PATH entry is not manager-derived from marker/cache state"}
	}
	if _, err := os.Lstat(ran); !os.IsNotExist(err) {
		return conformancecoverage.Observation{FailureReason: "external artifact ran during install.Project"}
	}
	if testCase.ArtifactExecutedDuringInstall || !testCase.PathEntryDerivedByManager {
		return conformancecoverage.Observation{FailureReason: "published path/shim vector contradicts the manager contract"}
	}
	if installPlatform() == "unix" {
		command := exec.Command(shim, "space value", "two") // #nosec G204 -- fixture launches the manager-derived test shim
		command.Env = []string{"PATH=/usr/bin"}
		output, runErr := command.CombinedOutput()
		var exitErr *exec.ExitError
		if !errors.As(runErr, &exitErr) || exitErr.ExitCode() != 37 || !strings.Contains(string(output), "arg=<space value>") || !strings.Contains(string(output), "arg=<two>") || !strings.Contains(string(output), ":/usr/bin") {
			return conformancecoverage.Observation{FailureReason: fmt.Sprintf("shim did not preserve arguments, inherited PATH and exit status: code=%v output=%q", runErr, output)}
		}
	} else if !strings.Contains(string(shimBytes), "%*") || !strings.Contains(string(shimBytes), "%ERRORLEVEL%") || !strings.Contains(string(shimBytes), ";%PATH%") {
		return conformancecoverage.Observation{FailureReason: "Windows shim does not preserve arguments, inherited PATH and exit status"}
	}
	return conformancecoverage.Observation{}
}

func exercisePackagePathEntryRejection(t *testing.T, testCase lifecyclePathShimCase) conformancecoverage.Observation {
	t.Helper()
	e := newEnv(t)
	writeLegacyBuildSkill(t, e, 7, []string{"go-repository-v1"})
	deps, _ := lifecycleBuildDeps(t)
	external := lifecycleExternalDepsWith(func() *buildrepo.Snapshot { return lifecycleSnapshotWithField("golden-tool", "path_entries") })
	result := e.install(Options{Build: deps, External: external})
	joined := strings.Join(result.Errors, "\n")
	if result.Status != "failed" || !strings.Contains(joined, testCase.ExpectedError) {
		return conformancecoverage.Observation{FailureReason: fmt.Sprintf("install.Project status=%s error=%q, want %s", result.Status, joined, testCase.ExpectedError)}
	}
	shim := filepath.Join(e.project, ".agents", "bin", shimName("golden-tool"))
	if _, err := os.Lstat(shim); !os.IsNotExist(err) || marker.Read(filepath.Join(e.project, ".agents", "skills", "golden-skill")) != nil {
		return conformancecoverage.Observation{FailureReason: "forbidden package path entry left a shim or marker after refusal"}
	}
	if testCase.ShimPublished {
		return conformancecoverage.Observation{FailureReason: "published path/shim vector allows a shim after output-path refusal"}
	}
	return conformancecoverage.Observation{}
}

func exerciseShimRollback(t *testing.T, testCase lifecyclePathShimCase) conformancecoverage.Observation {
	t.Helper()
	e := newEnv(t)
	writeLegacyBuildSkill(t, e, 7, []string{"go-repository-v1", "go-v1"})
	deps, _ := lifecycleBuildDeps(t)
	if first := e.install(Options{Build: deps, External: lifecycleExternalDeps()}); first.Status != "ok" {
		return conformancecoverage.Observation{FailureReason: "baseline install.Project: " + strings.Join(first.Errors, "; ")}
	}
	installed := filepath.Join(e.project, ".agents", "skills", "golden-skill")
	shimPath := filepath.Join(e.project, ".agents", "bin", shimName("golden-tool"))
	localShimPath := filepath.Join(e.project, ".agents", "bin", shimName("local-helper"))
	priorMarker, err := os.ReadFile(filepath.Join(installed, marker.Name))
	if err != nil {
		t.Fatal(err)
	}
	priorShim, err := os.ReadFile(shimPath)
	if err != nil {
		t.Fatal(err)
	}
	priorLocalShim, err := os.ReadFile(localShimPath)
	if err != nil {
		t.Fatal(err)
	}
	injected := false
	commit := CommitDeps{Hooks: transaction.Hooks{Fault: func(event transaction.Event) error {
		if event.Point == transaction.PointAfterBackup && event.Class == staging.ClassCanonicalShim && event.Identifier == "golden-tool" {
			injected = true
			return errors.New("injected shim collision after backup")
		}
		return nil
	}}}
	result := e.install(Options{Build: deps, External: lifecycleExternalDeps(), Commit: commit})
	if result.Status != "failed" || !injected {
		return conformancecoverage.Observation{FailureReason: fmt.Sprintf("install.Project status=%s; shim transaction fault fired=%v", result.Status, injected)}
	}
	markerAfter, markerErr := os.ReadFile(filepath.Join(installed, marker.Name))
	shimAfter, shimErr := os.ReadFile(shimPath)
	localShimAfter, localShimErr := os.ReadFile(localShimPath)
	if markerErr != nil || shimErr != nil || localShimErr != nil || !reflect.DeepEqual(markerAfter, priorMarker) || !reflect.DeepEqual(shimAfter, priorShim) || !reflect.DeepEqual(localShimAfter, priorLocalShim) {
		return conformancecoverage.Observation{FailureReason: fmt.Sprintf("transaction rollback did not restore prior mixed marker/shims: markerErr=%v shimErr=%v localShimErr=%v", markerErr, shimErr, localShimErr)}
	}
	joined := strings.Join(result.Errors, "\n")
	if testCase.ExpectedError != "" && !strings.Contains(joined, testCase.ExpectedError) {
		return conformancecoverage.Observation{FailureReason: "install.Project restored prior state but did not report expected stable error code " + testCase.ExpectedError + ": " + joined}
	}
	if !testCase.PriorShimRestored {
		return conformancecoverage.Observation{FailureReason: "published shim rollback vector does not expect restoration"}
	}
	return conformancecoverage.Observation{}
}

func TestAuthoritativeSigningCasesUseProjectInstallEntry(t *testing.T) {
	_, vectors := readExternalLifecycleVectors(t)
	conformancecoverage.RunOutcomes(t, "external-repository-lifecycle/signing_cases", vectors.SigningCases,
		func(testCase lifecycleSigningCase) string { return testCase.Name },
		func(t *testing.T, testCase lifecycleSigningCase) conformancecoverage.Observation {
			switch testCase.Name {
			case "unsigned-local-build":
				e := newEnv(t)
				writeLegacyBuildSkill(t, e, 6, []string{"go-v1"})
				deps, builder := lifecycleBuildDeps(t)
				result := e.install(Options{Build: deps})
				if result.Status != "ok" || len(builder.calls) != 1 || builder.calls[0] != "local-helper" || testCase.Result != "supported" || testCase.ArtifactExecutedDuringInstall || testCase.ManagerPostSigning {
					return conformancecoverage.Observation{FailureReason: fmt.Sprintf("unsigned local install status=%s builds=%v vector=%+v", result.Status, builder.calls, testCase)}
				}
				return conformancecoverage.Observation{}
			case "package-signing-request":
				e := newEnv(t)
				writeLegacyBuildSkill(t, e, 7, []string{"go-repository-v1"})
				dir := filepath.Join(e.skillsRoot, "golden-skill")
				payload, err := os.ReadFile(filepath.Join(dir, "csk-skill.json"))
				if err != nil {
					t.Fatal(err)
				}
				var spec map[string]any
				if err := json.Unmarshal(payload, &spec); err != nil {
					t.Fatal(err)
				}
				commands := spec["commands"].(map[string]any)
				commands["golden-tool"].(map[string]any)["signing"] = "required"
				payload, err = json.Marshal(spec)
				if err != nil {
					t.Fatal(err)
				}
				e.write(dir, "csk-skill.json", string(payload))
				e.git(dir, "add", ".")
				e.git(dir, "commit", "-qm", "package signing request")
				e.git(dir, "tag", "-f", "v1")
				acquisitions := 0
				external := lifecycleExternalDepsWith(func() *buildrepo.Snapshot { acquisitions++; return lifecycleSnapshot("golden-tool") })
				deps, _ := lifecycleBuildDeps(t)
				result := e.install(Options{Build: deps, External: external})
				if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, "\n"), testCase.ExpectedError) || acquisitions != 0 || testCase.SignerStarted {
					return conformancecoverage.Observation{FailureReason: fmt.Sprintf("package signing gate status=%s acquisitions=%d errors=%v signerStarted=%v", result.Status, acquisitions, result.Errors, testCase.SignerStarted)}
				}
				return conformancecoverage.Observation{}
			case "platform-requires-local-signing":
				e := newEnv(t)
				writeLegacyBuildSkill(t, e, 7, []string{"go-repository-v1"})
				acquisitions := 0
				external := lifecycleExternalDepsWith(func() *buildrepo.Snapshot { acquisitions++; return lifecycleSnapshot("golden-tool") })
				external.SigningPolicy = "platform-required"
				deps, builder := lifecycleBuildDeps(t)
				result := e.install(Options{Build: deps, External: external})
				if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, "\n"), testCase.ExpectedError) || acquisitions != 0 || len(builder.calls) != 0 || testCase.SignerStarted {
					return conformancecoverage.Observation{FailureReason: fmt.Sprintf("signer policy gate status=%s acquisitions=%d builds=%v errors=%v", result.Status, acquisitions, builder.calls, result.Errors)}
				}
				return conformancecoverage.Observation{}
			case "release-pipeline-signing":
				return conformancecoverage.Observation{BoundReason: "release signing belongs to the external release pipeline; install.Project exposes no post-signing operation, and the vector names its release-pipeline owner"}
			default:
				return conformancecoverage.Observation{FailureReason: "published signing case has no production install binding"}
			}
		})
}

func TestAuthoritativeTransactionCasesUseProjectInstallEntry(t *testing.T) {
	_, vectors := readExternalLifecycleVectors(t)
	conformancecoverage.RunOutcomes(t, "external-repository-lifecycle/transaction_cases", vectors.TransactionCases,
		func(testCase lifecycleTransactionCase) string { return testCase.Name },
		func(t *testing.T, testCase lifecycleTransactionCase) conformancecoverage.Observation {
			switch testCase.Name {
			case "failure-before-publication":
				return exerciseBuildFailure(t, testCase)
			case "failure-after-private-stage":
				return exercisePublicationFailure(t, testCase)
			case "marker-consumer-last":
				return exerciseMarkerCommitFailure(t, testCase)
			case "recovery-uncertain-journal":
				return exerciseUncertainJournalRecovery(t, testCase)
			default:
				return conformancecoverage.Observation{FailureReason: "published transaction case has no production install binding"}
			}
		})
}

func transactionFixture(t *testing.T) (*env, BuildDeps, *fakeBuilder) {
	t.Helper()
	e := newEnv(t)
	writeLegacyBuildSkill(t, e, 7, []string{"go-repository-v1", "go-v1"})
	deps, builder := lifecycleBuildDeps(t)
	return e, deps, builder
}

func stateUnchanged(t *testing.T, before, after sharedState, label string) error {
	t.Helper()
	diff := before.diff(after)
	if len(diff) != 0 {
		return fmt.Errorf("%s changed live manager state: %v", label, diff)
	}
	return nil
}

func exerciseBuildFailure(t *testing.T, testCase lifecycleTransactionCase) conformancecoverage.Observation {
	t.Helper()
	e, deps, builder := transactionFixture(t)
	before := snapshotState(t, e)
	builder.failOn["golden-tool"] = errors.New("injected external build failure")
	result := e.install(Options{Build: deps, External: lifecycleExternalDeps()})
	if result.Status != "failed" {
		return conformancecoverage.Observation{FailureReason: "install.Project did not fail at the injected build boundary"}
	}
	if err := stateUnchanged(t, before, snapshotState(t, e), testCase.Name); err != nil {
		return conformancecoverage.Observation{FailureReason: err.Error()}
	}
	if marker.Read(filepath.Join(e.project, ".agents", "skills", "golden-skill")) != nil || len(dirNames(t, filepath.Join(e.home, "external-build-cache", buildrepo.ArtifactsDir(buildrepo.LegacyReceiptSchemaVersion)))) != 0 {
		return conformancecoverage.Observation{FailureReason: "failed external build published an install marker or protected receipt"}
	}
	return conformancecoverage.Observation{}
}

func exercisePublicationFailure(t *testing.T, testCase lifecycleTransactionCase) conformancecoverage.Observation {
	t.Helper()
	e, deps, _ := transactionFixture(t)
	before := snapshotState(t, e)
	injected := false
	commit := CommitDeps{Hooks: transaction.Hooks{Fault: func(event transaction.Event) error {
		if event.Point == transaction.PointAfterBackup && event.Class == "05-external-cache" && strings.Contains(event.Identifier, "golden-tool") {
			injected = true
			return errors.New("injected external publication failure")
		}
		return nil
	}}}
	result := e.install(Options{Build: deps, External: lifecycleExternalDeps(), Commit: commit})
	if result.Status != "failed" || !injected {
		return conformancecoverage.Observation{FailureReason: fmt.Sprintf("install.Project status=%s; private-stage publication fault fired=%v", result.Status, injected)}
	}
	if err := stateUnchanged(t, before, snapshotState(t, e), testCase.Name); err != nil {
		return conformancecoverage.Observation{FailureReason: err.Error()}
	}
	if marker.Read(filepath.Join(e.project, ".agents", "skills", "golden-skill")) != nil || len(dirNames(t, filepath.Join(e.home, "external-build-cache", buildrepo.ArtifactsDir(buildrepo.LegacyReceiptSchemaVersion)))) != 0 {
		return conformancecoverage.Observation{FailureReason: "failed private-stage publication left a marker or receipt"}
	}
	return conformancecoverage.Observation{}
}

func exerciseMarkerCommitFailure(t *testing.T, testCase lifecycleTransactionCase) conformancecoverage.Observation {
	t.Helper()
	e, deps, _ := transactionFixture(t)
	before := snapshotState(t, e)
	injected := false
	commit := CommitDeps{Hooks: transaction.Hooks{Fault: func(event transaction.Event) error {
		if event.Point == transaction.PointAfterBackup && event.Class == staging.ClassContext && event.Identifier == "project/golden-skill" {
			injected = true
			return errors.New("injected marker commit failure")
		}
		return nil
	}}}
	result := e.install(Options{Build: deps, External: lifecycleExternalDeps(), Commit: commit})
	if result.Status != "failed" || !injected {
		return conformancecoverage.Observation{FailureReason: fmt.Sprintf("install.Project status=%s; marker commit fault fired=%v", result.Status, injected)}
	}
	if err := stateUnchanged(t, before, snapshotState(t, e), testCase.Name); err != nil {
		return conformancecoverage.Observation{FailureReason: err.Error()}
	}
	if marker.Read(filepath.Join(e.project, ".agents", "skills", "golden-skill")) != nil || len(mustRegisteredConsumers(t, e.home)) != 0 {
		return conformancecoverage.Observation{FailureReason: "marker failure published partial currentness or consumer state"}
	}
	return conformancecoverage.Observation{}
}

func exerciseUncertainJournalRecovery(t *testing.T, testCase lifecycleTransactionCase) conformancecoverage.Observation {
	t.Helper()
	e, deps, _ := transactionFixture(t)
	if first := e.install(Options{Build: deps, External: lifecycleExternalDeps()}); first.Status != "ok" {
		return conformancecoverage.Observation{FailureReason: "baseline install.Project: " + strings.Join(first.Errors, "; ")}
	}
	before := snapshotState(t, e)
	afterBackup, afterRestore := false, false
	commit := CommitDeps{Hooks: transaction.Hooks{Fault: func(event transaction.Event) error {
		if event.Class != staging.ClassCanonicalShim || event.Identifier != "golden-tool" {
			return nil
		}
		if event.Point == transaction.PointAfterBackup && !afterBackup {
			afterBackup = true
			return errors.New("injected uncertain shim transaction")
		}
		if event.Point == transaction.PointAfterRestore && !afterRestore {
			afterRestore = true
			return errors.New("injected journal state-write interruption")
		}
		return nil
	}}}
	failed := e.install(Options{Build: deps, External: lifecycleExternalDeps(), Commit: commit})
	if failed.Status != "failed" || !afterBackup || !afterRestore {
		return conformancecoverage.Observation{FailureReason: fmt.Sprintf("uncertain transaction status=%s backup=%v restore=%v", failed.Status, afterBackup, afterRestore)}
	}
	journalRoot := filepath.Join(e.home, "state", "transactions", "v1")
	if len(dirNames(t, journalRoot)) == 0 {
		return conformancecoverage.Observation{FailureReason: "uncertain transaction did not retain a recovery journal"}
	}
	commit, err := (CommitDeps{}).resolve(e.home, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := recoverJournals(context.Background(), commit); err != nil {
		return conformancecoverage.Observation{FailureReason: "production recovery refused a repairable uncertain journal: " + err.Error()}
	}
	if err := stateUnchanged(t, before, snapshotState(t, e), testCase.Name); err != nil {
		return conformancecoverage.Observation{FailureReason: err.Error()}
	}
	if len(dirNames(t, journalRoot)) != 0 || testCase.Recovery != "fail-closed-or-rollback" {
		return conformancecoverage.Observation{FailureReason: "journal recovery did not reach a terminal rollback state"}
	}
	return conformancecoverage.Observation{}
}
