package crossconformance

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/protocoljson"
)

// External-evidence semantic rows (§11): each recorded field of an
// external build record is mutated in turn; the mismatch against the
// protected verified receipt must be detected, and repair must
// re-derive the record from verified state instead of adopting the
// recorded bytes.
//
// Detection compares two production artifacts: the recorded marker arm
// (untrusted claim) against the protected receipt's driver input
// (verified at publish by the pipeline and re-verified by exact lookup
// on every install). This is the comparison the CLI status comparator
// performs; the CLI verdict itself needs the exact external source
// reachable (status re-acquires rather than trusting recorded bytes),
// which the headless lane cannot provide — see
// TestDraftExternalStatusFailsClosedWithoutSource, referenced by every
// row. Repair runs install.Project: the marker must come back
// byte-equal to the honest baseline, and protected tampering must
// rebuild (the builder runs) rather than adopt.

func init() {
	for _, field := range []string{"repository", "declared_identity", "declared_locked_commit", "declared_tag", "effective_identity", "object_format", "commit", "substituted", "substitution", "build_source", "descriptor_target", "execution_policy", "cache_key", "receipt_sha256", "artifact_sha256", "artifact_path", "input.package"} {
		field := field
		registerDraftSemantic("external-evidence-mismatch-"+field, func(t *testing.T, c draftSemanticCase) {
			driveExternalMismatch(t, c, field)
		})
	}
	registerDraftSemantic("external-only-current", driveExternalOnlyCurrent)
}

// externalBaseline installs one external etool skill with the fake
// build seam and returns the project, home, recorded arm, and raw
// protected receipt.
func externalBaseline(t *testing.T, substituted bool) (project, home string, arm map[string]any, receiptRaw []byte) {
	t.Helper()
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	project = t.TempDir()
	home = t.TempDir()
	writeExternalSkill(t, filepath.Join(project, "skills", "review"), "review")
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	runDraftGit(t, project, "init", "-q")
	if err := os.WriteFile(filepath.Join(project, ".gitignore"), []byte(".agents/\n.claude/skills/\nSkillfile.dev.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	resolveDraftPlan(t, project, home, payload)
	if substituted {
		writeDevSubstitution(t, project, `{"schema_version":2,"substitutions":{},"build_repository_substitutions":{"review":{"tools":{"git":"https://git.example.com/skills/tools-mirror.git","ref":{"kind":"tag","value":"v9"}}}}}`)
	}
	deps, _ := xbBuildDeps(t)
	result := install.Project(draftInstallConfig(home), project, "test", install.Options{Platform: draftPlatform(), Build: deps, External: xbExternalDeps()})
	if result.Status != "ok" {
		t.Fatalf("baseline install = %+v", result)
	}
	return project, home, readExternalArm(t, project), readProtectedReceiptRaw(t, home)
}

func deepCopyArm(t *testing.T, arm map[string]any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(arm)
	if err != nil {
		t.Fatal(err)
	}
	var copied map[string]any
	if err := json.Unmarshal(raw, &copied); err != nil {
		t.Fatal(err)
	}
	return copied
}

func readExternalArm(t *testing.T, project string) map[string]any {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join(project, ".agents", "skills", "review", marker.Name))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	builds, ok := doc["builds"].(map[string]any)
	if !ok {
		t.Fatalf("marker has no builds: %.200s", payload)
	}
	arm, ok := builds["etool"].(map[string]any)
	if !ok {
		t.Fatalf("marker has no etool arm: %.200s", payload)
	}
	return arm
}

func writeExternalArm(t *testing.T, project string, arm map[string]any) {
	t.Helper()
	markerPath := filepath.Join(project, ".agents", "skills", "review", marker.Name)
	payload, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	doc["builds"].(map[string]any)["etool"] = arm
	mutated, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(markerPath, mutated, 0o644); err != nil {
		t.Fatal(err)
	}
}

func readProtectedReceiptRaw(t *testing.T, home string) []byte {
	t.Helper()
	var found string
	if err := filepath.Walk(filepath.Join(home, "external-build-cache"), func(path string, info os.FileInfo, err error) error {
		if err == nil && info.Name() == "receipt.json" {
			found = path
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if found == "" {
		t.Fatal("no protected receipt published")
	}
	payload, err := os.ReadFile(found)
	if err != nil {
		t.Fatal(err)
	}
	// The external driver input is verified by the pipeline's own exact
	// lookup; here the row pins the canonical bytes plus the closed
	// top-level shape.
	if err := protocoljson.Validate(payload); err != nil {
		t.Fatalf("protected receipt is not canonical: %v", err)
	}
	var shape map[string]any
	if err := json.Unmarshal(payload, &shape); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"schema_version", "cache_key", "input", "artifact"} {
		if _, ok := shape[key]; !ok {
			t.Fatalf("protected receipt misses %q", key)
		}
	}
	return payload
}

func protectedReceiptPath(t *testing.T, home string) string {
	t.Helper()
	var found string
	if err := filepath.Walk(filepath.Join(home, "external-build-cache"), func(path string, info os.FileInfo, err error) error {
		if err == nil && info.Name() == "receipt.json" {
			found = path
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if found == "" {
		t.Fatal("no protected receipt published")
	}
	return found
}

func parseReceiptJSON(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var doc map[string]any
	if err := decoder.Decode(&doc); err != nil {
		t.Fatal(err)
	}
	return doc
}

func receiptBuild(t *testing.T, receipt map[string]any) map[string]any {
	t.Helper()
	input, ok := receipt["input"].(map[string]any)
	if !ok {
		t.Fatalf("receipt has no input: %v", receipt)
	}
	build, ok := input["build"].(map[string]any)
	if !ok {
		t.Fatalf("receipt input has no build wrapper: %v", input)
	}
	return build
}

// externalFieldBinding asserts the recorded arm and the verified
// receipt agree on every compared field except the mutated one, which
// must disagree. The receipt side is the verified planned evidence;
// the arm side is the recorded claim.
func externalFieldBinding(t *testing.T, arm map[string]any, receipt map[string]any, mutated string) {
	t.Helper()
	build := receiptBuild(t, receipt)
	source, _ := build["source"].(map[string]any)
	declared, _ := source["declared"].(map[string]any)
	declaredIdentity, _ := declared["identity"].(map[string]any)
	declaredCommit, _ := declared["locked_commit"].(map[string]any)
	effective, _ := source["effective"].(map[string]any)
	effectiveIdentity, _ := effective["identity"].(map[string]any)
	descriptor, _ := source["descriptor"].(map[string]any)
	effectiveBuildSource, _ := effective["build_source"].(map[string]any)
	policy, _ := build["policy"].(map[string]any)
	artifact, _ := receipt["artifact"].(map[string]any)
	canonical, err := protocoljson.MarshalCanonical(receipt)
	if err != nil {
		t.Fatalf("canonical receipt: %v", err)
	}
	receiptSum := sha256.Sum256(canonical)
	armDeclaredIdentity, _ := arm["declared_identity"].(map[string]any)
	armDeclaredCommit, _ := arm["declared_locked_commit"].(map[string]any)
	armEffectiveIdentity, _ := arm["effective_identity"].(map[string]any)
	armBuildSource, _ := arm["build_source"].(map[string]any)

	compare := map[string][2]any{
		"repository":             {arm["repository"], source["repository"]},
		"declared_identity":      {armDeclaredIdentity["value"], declaredIdentity["value"]},
		"declared_locked_commit": {armDeclaredCommit["hex"], declaredCommit["hex"]},
		"declared_tag":           {arm["declared_tag"], declared["tag"]},
		"effective_identity":     {armEffectiveIdentity["value"], effectiveIdentity["value"]},
		"object_format":          {arm["object_format"], effective["object_format"]},
		"commit":                 {arm["commit"], effective["commit"]},
		"substituted":            {arm["substituted"], effective["substituted"]},
		"substitution":           {arm["substitution"], effective["substitution"]},
		"build_source":           {armBuildSource["content_sha256"], effectiveBuildSource["content_sha256"]},
		"descriptor_target":      {arm["descriptor_target"], descriptor["target"]},
		"execution_policy":       {arm["execution_policy"], policy["execution_policy"]},
		"cache_key":              {arm["cache_key"], receipt["cache_key"]},
		"receipt_sha256":         {arm["receipt_sha256"], "sha256:" + hex.EncodeToString(receiptSum[:])},
		"artifact_sha256":        {arm["artifact_sha256"], artifact["sha256"]},
		"artifact_path":          {arm["artifact_path"], artifact["path"]},
	}
	for field, pair := range compare {
		match := reflect.DeepEqual(pair[0], pair[1])
		if field == mutated || (mutated == "object_format" && (field == "commit" || field == "declared_locked_commit")) ||
			(mutated == "substituted" && field == "substitution") {
			continue
		}
		if !match {
			t.Fatalf("field %s disagrees outside the mutation: recorded=%v verified=%v", field, pair[0], pair[1])
		}
	}
	if mutated == "\x00none" {
		return
	}
	pair, ok := compare[mutated]
	if !ok {
		t.Fatalf("no binding for field %q", mutated)
	}
	if reflect.DeepEqual(pair[0], pair[1]) {
		t.Fatalf("mutated field %s still agrees: %v", mutated, pair[0])
	}
}

func mutateExternalField(t *testing.T, project, home string, field string, arm map[string]any) {
	t.Helper()
	otherHex := func() string { return strings.Repeat("b", 40) }
	otherSHA := func() string { return "sha256:" + strings.Repeat("f", 64) }
	switch field {
	case "repository":
		arm["repository"] = "tools2"
	case "declared_identity":
		arm["declared_identity"].(map[string]any)["value"] = "git.example.com/skills/other"
	case "declared_locked_commit":
		arm["declared_locked_commit"].(map[string]any)["hex"] = otherHex()
	case "declared_tag":
		arm["declared_tag"] = "v9"
	case "effective_identity":
		arm["effective_identity"].(map[string]any)["value"] = "git.example.com/skills/other"
	case "object_format":
		// object_format is length-coupled to both commit fields: sha256
		// requires 64 hex digits, and the declared commit repeats the
		// format, so the row mutates the coupled triple.
		arm["object_format"] = "sha256"
		arm["commit"] = strings.Repeat("ab", 32)
		arm["declared_locked_commit"].(map[string]any)["hex"] = strings.Repeat("ab", 32)
		arm["declared_locked_commit"].(map[string]any)["object_format"] = "sha256"
	case "commit":
		arm["commit"] = otherHex()
	case "substituted":
		arm["substituted"] = true
		arm["substitution"] = map[string]any{"type": "local-path"}
	case "substitution":
		sub, ok := arm["substitution"].(map[string]any)
		if !ok {
			t.Fatal("substitution baseline recorded no substitution")
		}
		ref, ok := sub["ref"].(map[string]any)
		if !ok {
			t.Fatal("substitution baseline recorded no ref")
		}
		ref["value"] = "v10"
	case "build_source":
		arm["build_source"].(map[string]any)["content_sha256"] = otherSHA()
	case "descriptor_target":
		arm["descriptor_target"] = "other"
	case "execution_policy":
		arm["execution_policy"] = "arbitrary"
	case "cache_key":
		arm["cache_key"] = otherSHA()
	case "receipt_sha256":
		arm["receipt_sha256"] = otherSHA()
	case "artifact_sha256":
		arm["artifact_sha256"] = otherSHA()
	case "artifact_path":
		// The closed shape admits the unix or the windows derivation;
		// the row records the foreign one. The honest baseline is the
		// native derivation, so the mutation reads the recorded value
		// instead of assuming unix.
		if arm["artifact_path"] == "bin/etool.exe" {
			arm["artifact_path"] = "bin/etool"
		} else {
			arm["artifact_path"] = "bin/etool.exe"
		}
	case "input.package":
		raw, err := os.ReadFile(protectedReceiptPath(t, home))
		if err != nil {
			t.Fatal(err)
		}
		receipt := parseReceiptJSON(t, raw)
		receipt["input"].(map[string]any)["package"].(map[string]any)["snapshot"] = otherSHA()
		canonical, err := protocoljson.MarshalCanonical(receipt)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(protectedReceiptPath(t, home), canonical, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	default:
		t.Fatalf("no mutation for field %q", field)
	}
	writeExternalArm(t, project, arm)
}

func driveExternalMismatch(t *testing.T, _ draftSemanticCase, field string) {
	project, home, baseline, receiptRaw := externalBaseline(t, field == "substitution")
	receipt := parseReceiptJSON(t, receiptRaw)
	// The honest baseline binds every compared field; declared_tag is
	// absent on both sides of an untagged declaration.
	externalFieldBinding(t, baseline, receipt, "\x00none")
	mutatedArm := deepCopyArm(t, baseline)
	mutateExternalField(t, project, home, field, mutatedArm)
	if field == "input.package" {
		// The protected receipt is the tampered evidence: the pipeline
		// input no longer matches the verified entry, so repair must
		// rebuild rather than adopt.
		tampered, err := os.ReadFile(protectedReceiptPath(t, home))
		if err != nil {
			t.Fatal(err)
		}
		if err := protocoljson.Validate(tampered); err != nil {
			t.Fatalf("tampered receipt is not canonical: %v", err)
		}
		if parseReceiptJSON(t, tampered)["input"].(map[string]any)["package"].(map[string]any)["snapshot"] == receipt["input"].(map[string]any)["package"].(map[string]any)["snapshot"] {
			t.Fatal("package tamper did not land")
		}
		deps, builder := xbBuildDeps(t)
		result := install.Project(draftInstallConfig(home), project, "test", install.Options{Platform: draftPlatform(), Build: deps, External: xbExternalDeps()})
		if result.Status != "ok" {
			t.Fatalf("repair = %+v, want rebuild success", result)
		}
		if len(builder.calls) == 0 {
			t.Fatal("repair adopted the tampered entry instead of rebuilding")
		}
		restored := readExternalArm(t, project)
		if !reflect.DeepEqual(restored, baseline) {
			t.Fatal("repair did not re-derive the honest record")
		}
		assertProtectedExternalBytes(t, home, readExternalArm(t, project), "etool")
		return
	}
	mutated := readExternalArm(t, project)
	if field == "execution_policy" {
		// execution_policy is a fixed constant in the closed arm shape,
		// so the mutation voids the marker itself: status refuses the
		// document and repair records a fresh honest one.
		if marker.Read(filepath.Join(project, ".agents", "skills", "review")) != nil {
			t.Fatal("voided marker still reads")
		}
	} else {
		if marker.Read(filepath.Join(project, ".agents", "skills", "review")) == nil {
			t.Fatalf("mutation of %s voided the marker; want a valid mismatch", field)
		}
		externalFieldBinding(t, mutated, receipt, field)
	}
	deps, _ := xbBuildDeps(t)
	result := install.Project(draftInstallConfig(home), project, "test", install.Options{Platform: draftPlatform(), Build: deps, External: xbExternalDeps()})
	if result.Status != "ok" {
		t.Fatalf("repair = %+v, want re-derivation success", result)
	}
	restored := readExternalArm(t, project)
	if !reflect.DeepEqual(restored, baseline) {
		t.Fatalf("repair kept recorded bytes: %+v", restored)
	}
	assertProtectedExternalBytes(t, home, readExternalArm(t, project), "etool")
}

// assertProtectedExternalBytes proves the protected artifact is the
// verified staged output and hashes to the recorded digest, never
// recorded bytes.
func assertProtectedExternalBytes(t *testing.T, home string, arm map[string]any, command string) {
	t.Helper()
	artifactPath := filepath.Join(filepath.Dir(protectedReceiptPath(t, home)), "artifact")
	payload, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(payload) != "artifact:"+command {
		t.Fatalf("protected %s artifact is not the verified staged output", command)
	}
	sum := sha256.Sum256(payload)
	if got := "sha256:" + hex.EncodeToString(sum[:]); arm["artifact_sha256"] != got {
		t.Fatalf("artifact_sha256 = %v, want %s", arm["artifact_sha256"], got)
	}
}

func driveExternalOnlyCurrent(t *testing.T, _ draftSemanticCase) {
	_, home, arm, receiptRaw := externalBaseline(t, false)
	receipt := parseReceiptJSON(t, receiptRaw)
	// Completeness: the verified receipt binds every recorded field,
	// the receipt bytes hash to the recorded digest, and the protected
	// artifact hashes to the recorded digest.
	externalFieldBinding(t, arm, receipt, "\x00none")
	sum := sha256.Sum256(receiptRaw)
	if got := "sha256:" + hex.EncodeToString(sum[:]); arm["receipt_sha256"] != got {
		t.Fatalf("receipt_sha256 = %v, want %s", arm["receipt_sha256"], got)
	}
	assertProtectedExternalBytes(t, home, arm, "etool")
}

// TestDraftExternalStatusFailsClosedWithoutSource proves the CLI status
// verdict never trusts recorded external bytes without the exact
// source: without network it fails closed with
// build_repository_source_unavailable and mutates nothing. Every
// external-evidence row references this bound for its status half.
func TestDraftExternalStatusFailsClosedWithoutSource(t *testing.T) {
	root := t.TempDir()
	configPath, project, home := setupCLIProject(t, root)
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	writeExternalSkill(t, filepath.Join(project, "skills", "review"), "review")
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	resolveDraftPlan(t, project, home, payload)
	cfg := draftCLIConfig(t, root, configPath)
	deps, _ := xbBuildDeps(t)
	result := install.Project(cfg, project, "test", install.Options{Platform: draftPlatform(), Build: deps, External: xbExternalDeps()})
	if result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}
	before := treeDigestFiltered(t, project, nil) + treeDigestFiltered(t, home, func(rel string) bool {
		return rel == "cache" || strings.HasPrefix(rel, "cache"+string(filepath.Separator)) ||
			rel == "state" || strings.HasPrefix(rel, "state"+string(filepath.Separator))
	})
	code, stdout, stderr := runCurator(t, home, configPath, nil, "status", "app")
	if code == 0 {
		t.Fatalf("status without the exact source succeeded:\n%s\n%s", stdout, stderr)
	}
	if !strings.Contains(stdout+stderr, "build_repository_source_unavailable") {
		t.Fatalf("status misses the source-unavailable refusal:\n%s\n%s", stdout, stderr)
	}
	after := treeDigestFiltered(t, project, nil) + treeDigestFiltered(t, home, func(rel string) bool {
		return rel == "cache" || strings.HasPrefix(rel, "cache"+string(filepath.Separator)) ||
			rel == "state" || strings.HasPrefix(rel, "state"+string(filepath.Separator))
	})
	if after != before {
		t.Fatal("status mutated state")
	}
}
