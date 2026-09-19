package install

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/devsub"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/sourcelock"
)

// This suite proves the install-marker migration at the production entry
// (install.Project): every schema-2 installation records marker schema 5
// with its frozen package and binding lock, Git members keep the
// attestation and substitution semantics the migration table preserves,
// local snapshots admit neither, and the currentness readers compare the
// full identity against the effective plan without ever authorizing from
// a recorded summary.

// setupGitScriptInstall resolves one Git-selected script package with a
// declared runtime root through the production closure and writes its
// lock plus machine bindings, returning the project, home and repository.
func setupGitScriptInstall(t *testing.T) (project, home, repo string) {
	t.Helper()
	payload := `{"schema_version":2,"sources":{"s":{"git":"https://example.org/kit.git","tag":"v1"}},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`
	project, home, _ = draftProject(t, payload, nil)
	repo = t.TempDir()
	writeDraftScriptSkill(t, filepath.Join(repo, "skills", "review"), "review", "rtool", "#!/bin/sh\necho review-ok\n", nil)
	testGit(t, repo, "init", "-q", "-b", "main")
	testGit(t, repo, "add", ".")
	testGit(t, repo, "commit", "-qm", "fixture")
	testGit(t, repo, "tag", "v1")
	m, err := manifest.ParseBytesWithOptions([]byte(payload), filepath.Join(project, "Skillfile.json"), manifest.ParseOptions{DraftSourcesV1: true})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := closure.ResolveDraft(closure.DraftResolveConfig{
		ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: []byte(payload),
		Expansion: manifest.ExpansionOptions{GitRoots: map[string]string{"s": repo}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := sourcelock.Write(sourcelock.PathIn(project), plan.Lock); err != nil {
		t.Fatal(err)
	}
	bindings, err := sourcelock.NewBindings(plan.Lock.LockSHA256, map[string]sourcelock.SourceBinding{"s": {Location: repo}})
	if err != nil {
		t.Fatal(err)
	}
	if err := sourcelock.WriteBindings(DraftBindingsPath(home, project), bindings); err != nil {
		t.Fatal(err)
	}
	return project, home, repo
}

// setupGitBuildInstall resolves one Git-selected build package declaring
// a local go-v1 command and an external go-repository-v1 command.
func setupGitBuildInstall(t *testing.T) (project, home string) {
	t.Helper()
	payload := `{"schema_version":2,"sources":{"s":{"git":"https://example.org/kit.git","tag":"v1"}},"skills":[{"name":"tooling","from":"s","directory":"skills/tooling"}]}`
	project, home, _ = draftProject(t, payload, nil)
	repo := t.TempDir()
	writeDraftBuildSkill(t, filepath.Join(repo, "skills", "tooling"), "tooling", true)
	testGit(t, repo, "init", "-q", "-b", "main")
	testGit(t, repo, "add", ".")
	testGit(t, repo, "commit", "-qm", "fixture")
	testGit(t, repo, "tag", "v1")
	m, err := manifest.ParseBytesWithOptions([]byte(payload), filepath.Join(project, "Skillfile.json"), manifest.ParseOptions{DraftSourcesV1: true})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := closure.ResolveDraft(closure.DraftResolveConfig{
		ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: []byte(payload),
		Expansion: manifest.ExpansionOptions{GitRoots: map[string]string{"s": repo}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := sourcelock.Write(sourcelock.PathIn(project), plan.Lock); err != nil {
		t.Fatal(err)
	}
	bindings, err := sourcelock.NewBindings(plan.Lock.LockSHA256, map[string]sourcelock.SourceBinding{"s": {Location: repo}})
	if err != nil {
		t.Fatal(err)
	}
	if err := sourcelock.WriteBindings(DraftBindingsPath(home, project), bindings); err != nil {
		t.Fatal(err)
	}
	return project, home
}

func draftRealInstall(t *testing.T, project, home string, opts Options) Result {
	t.Helper()
	cfg := draftTestConfig(home, t.TempDir())
	opts.DraftSourcesV1 = true
	opts.Platform = installPlatform()
	return Project(cfg, project, "test", opts)
}

// TestDraftGitInstallWritesMarkerV5 is the migration core: a network-git
// draft member installs for real under marker schema 5 with the exact
// frozen package and binding lock, the replaced legacy identity is
// absent, the declared ref stays bound through the lock and manifest,
// and a reinstall is current through that marker.
func TestDraftGitInstallWritesMarkerV5(t *testing.T) {
	project, home, _, _ := setupGitInstall(t)
	result := draftRealInstall(t, project, home, Options{})
	if result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}
	lock, err := sourcelock.Read(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lock.Find("review")
	if !ok {
		t.Fatal("lock misses review")
	}
	installed := filepath.Join(project, ".agents", "skills", "review")
	recorded := marker.Read(installed)
	if recorded == nil || recorded.SchemaVersion != marker.SchemaV5 || recorded.Package == nil {
		t.Fatalf("marker = %+v, want schema 5 with a package", recorded)
	}
	if recorded.Package.Kind != sourcelock.KindNetworkGit ||
		recorded.Package.Repository != member.Package.Repository ||
		recorded.Package.Commit == nil ||
		recorded.Package.Commit.ObjectFormat != member.Package.Commit.ObjectFormat ||
		recorded.Package.Commit.Hex != member.Package.Commit.Hex ||
		recorded.Package.Directory != member.Package.Directory ||
		recorded.LockSHA256 != lock.LockSHA256 {
		t.Fatalf("marker package %+v lock %s does not bind locked %+v %s",
			recorded.Package, recorded.LockSHA256, member.Package, lock.LockSHA256)
	}
	if recorded.Source != "" || recorded.Git != "" ||
		recorded.RefKind != "" || recorded.Ref != "" || recorded.Commit != "" {
		t.Fatalf("marker identity = %q %q %s %s %s, want empty",
			recorded.Source, recorded.Git, recorded.RefKind, recorded.Ref, recorded.Commit)
	}
	// The declared ref (tag v1) lives in the manifest and bound lock, not
	// the marker: the lock binds the declaring Skillfile bytes and the
	// marker binds the lock.
	manifestPayload, err := os.ReadFile(filepath.Join(project, "Skillfile.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(manifestPayload), `"tag":"v1"`) {
		t.Fatalf("manifest does not declare tag v1:\n%s", manifestPayload)
	}
	if err := lock.CheckStale(manifestPayload); err != nil {
		t.Fatalf("lock does not bind the declaring manifest: %v", err)
	}
	payload, err := os.ReadFile(filepath.Join(installed, marker.Name))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"source", "git", "ref_kind", "ref", "commit"} {
		if _, present := raw[field]; present {
			t.Fatalf("v5 Git marker carries replaced field %q", field)
		}
	}
	for _, field := range []string{"package", "lock_sha256"} {
		if _, present := raw[field]; !present {
			t.Fatalf("v5 Git marker misses field %q", field)
		}
	}
	again := draftRealInstall(t, project, home, Options{})
	if again.Status != "ok" {
		t.Fatalf("reinstall = %+v", again)
	}
	if !strings.Contains(strings.Join(again.Messages, "\n"), "up-to-date") {
		t.Fatalf("reinstall did not report up-to-date installations: %q", again.Messages)
	}
}

// TestDraftGitMarkerLockBindingRoundTrip is the Git-arm currency
// regression row: the marker a real install.Project records binds the
// locked package and lock generation, the currentness reader holds it
// current for an identical restaging, and a declared-ref move —
// re-resolved to tag v2 through the production closure — yields a new
// lock the installed marker is non-current against, so the reinstall
// re-stages instead of trusting the recorded summary.
func TestDraftGitMarkerLockBindingRoundTrip(t *testing.T) {
	project, home, repo, _ := setupGitInstall(t)
	result := draftRealInstall(t, project, home, Options{})
	if result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}
	lock, err := sourcelock.Read(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lock.Find("review")
	if !ok {
		t.Fatal("lock misses review")
	}
	installed := filepath.Join(project, ".agents", "skills", "review")
	recorded := marker.Read(installed)
	if recorded == nil {
		t.Fatalf("installed marker unreadable in %s", installed)
	}
	if recorded.SchemaVersion != marker.SchemaV5 {
		t.Fatalf("schema = %d, want %d", recorded.SchemaVersion, marker.SchemaV5)
	}
	if recorded.Package == nil || recorded.Package.Commit == nil ||
		recorded.Package.Commit.Hex != member.Package.Commit.Hex ||
		recorded.LockSHA256 != lock.LockSHA256 {
		t.Fatalf("marker binds %+v %s, want the locked %+v %s",
			recorded.Package, recorded.LockSHA256, member.Package, lock.LockSHA256)
	}
	if recorded.Source != "" || recorded.Git != "" ||
		recorded.RefKind != "" || recorded.Ref != "" || recorded.Commit != "" {
		t.Fatalf("marker identity = %q %q %s %s %s, want empty",
			recorded.Source, recorded.Git, recorded.RefKind, recorded.Ref, recorded.Commit)
	}
	identical := *recorded
	if current, err := marker.Current(installed, &identical); err != nil || !current {
		t.Fatalf("Current = %v, %v, want true for an identical restaging", current, err)
	}
	// Move the declared ref: a new upstream commit tagged v2, declared in
	// the manifest and re-resolved like an explicit refresh.
	skillMD := filepath.Join(repo, "skills", "review", "SKILL.md")
	skillPayload, err := os.ReadFile(skillMD)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skillMD, append(skillPayload, []byte("\n<!-- v2 -->\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	testGit(t, repo, "add", ".")
	testGit(t, repo, "commit", "-qm", "second")
	testGit(t, repo, "tag", "v2")
	movedPayload := `{"schema_version":2,"sources":{"s":{"git":"https://example.org/kit.git","tag":"v2"}},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(movedPayload), 0o644); err != nil {
		t.Fatal(err)
	}
	moved, err := manifest.ParseBytesWithOptions([]byte(movedPayload), filepath.Join(project, "Skillfile.json"), manifest.ParseOptions{DraftSourcesV1: true})
	if err != nil {
		t.Fatal(err)
	}
	movedPlan, err := closure.ResolveDraft(closure.DraftResolveConfig{
		ProjectRoot: project, Home: home, Manifest: moved, ManifestPayload: []byte(movedPayload),
		Expansion: manifest.ExpansionOptions{GitRoots: map[string]string{"s": repo}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := sourcelock.Write(sourcelock.PathIn(project), movedPlan.Lock); err != nil {
		t.Fatal(err)
	}
	movedBindings, err := sourcelock.NewBindings(movedPlan.Lock.LockSHA256, map[string]sourcelock.SourceBinding{"s": {Location: repo}})
	if err != nil {
		t.Fatal(err)
	}
	if err := sourcelock.WriteBindings(DraftBindingsPath(home, project), movedBindings); err != nil {
		t.Fatal(err)
	}
	movedMember, ok := movedPlan.Lock.Find("review")
	if !ok {
		t.Fatal("moved lock misses review")
	}
	if movedPlan.Lock.LockSHA256 == lock.LockSHA256 {
		t.Fatalf("re-resolve kept lock %s, want a new generation", lock.LockSHA256)
	}
	if movedMember.Package.Commit.Hex == member.Package.Commit.Hex {
		t.Fatalf("re-resolve kept commit %s, want the v2 commit", member.Package.Commit.Hex)
	}
	// The installed marker is non-current against the moved generation:
	// the lock comparison observes the declared-ref move.
	movedExpectation := *recorded
	movedExpectation.Package = draftMarkerPackage(movedMember.Package)
	movedExpectation.LockSHA256 = movedPlan.Lock.LockSHA256
	if current, err := marker.Current(installed, &movedExpectation); err != nil || current {
		t.Fatalf("Current = %v, %v across a declared-ref move, want false", current, err)
	}
	// The reinstall re-stages and binds the moved generation; a further
	// reinstall is up-to-date through the new marker.
	again := draftRealInstall(t, project, home, Options{})
	if again.Status != "ok" {
		t.Fatalf("reinstall = %+v", again)
	}
	restaged := marker.Read(installed)
	if restaged == nil || restaged.SchemaVersion != marker.SchemaV5 ||
		restaged.LockSHA256 != movedPlan.Lock.LockSHA256 ||
		restaged.Package == nil || restaged.Package.Commit == nil ||
		restaged.Package.Commit.Hex != movedMember.Package.Commit.Hex {
		t.Fatalf("restaged marker = %+v, want the moved package and lock", restaged)
	}
	third := draftRealInstall(t, project, home, Options{})
	if third.Status != "ok" {
		t.Fatalf("second reinstall = %+v", third)
	}
	if !strings.Contains(strings.Join(third.Messages, "\n"), "up-to-date") {
		t.Fatalf("second reinstall did not report up-to-date installations: %q", third.Messages)
	}
}

// TestDraftGitRuntimeMaterializesUnderSourceV1Key proves the migrated
// runtime boundary for Git: the declared runtime root materializes from
// the frozen commit tree under the namespaced package key — the same key
// GC derives from the recorded v5 marker — as copied bytes, never links.
func TestDraftGitRuntimeMaterializesUnderSourceV1Key(t *testing.T) {
	project, home, _ := setupGitScriptInstall(t)
	result := draftRealInstall(t, project, home, Options{})
	if result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}
	locked, key := draftLockedRuntimeKey(t, project, "review")
	runtimeScript := filepath.Join(home, "runtime", "review", key, "scripts", "rtool.sh")
	scriptPayload, err := os.ReadFile(runtimeScript)
	if err != nil {
		t.Fatalf("runtime script missing under source-v1 key %s: %v", key, err)
	}
	if !strings.Contains(string(scriptPayload), "review-ok") {
		t.Fatalf("runtime script has unexpected bytes: %q", scriptPayload)
	}
	assertNoLiveLinks(t, filepath.Join(home, "runtime", "review", key))
	installed := filepath.Join(project, ".agents", "skills", "review")
	if _, err := os.Lstat(filepath.Join(installed, "scripts")); !os.IsNotExist(err) {
		t.Fatalf("runtime root leaked into installed Git context")
	}
	recorded := marker.Read(installed)
	if recorded == nil || recorded.SchemaVersion != marker.SchemaV5 || recorded.Package == nil ||
		recorded.Package.Kind != sourcelock.KindNetworkGit {
		t.Fatalf("marker = %+v, want schema 5 network-git", recorded)
	}
	if recorded.Source != "" || recorded.Git != "" ||
		recorded.RefKind != "" || recorded.Ref != "" || recorded.Commit != "" {
		t.Fatalf("marker identity = %q %q %s %s %s, want empty",
			recorded.Source, recorded.Git, recorded.RefKind, recorded.Ref, recorded.Commit)
	}
	if recorded.LockSHA256 != locked.LockSHA256 {
		t.Fatalf("marker lock = %s, want the bound %s", recorded.LockSHA256, locked.LockSHA256)
	}
	shimPayload, err := os.ReadFile(filepath.Join(project, ".agents", "bin", shimName("rtool")))
	if err != nil {
		t.Fatalf("shim missing: %v", err)
	}
	if !strings.Contains(string(shimPayload), filepath.Join("runtime", "review")) {
		t.Fatalf("shim does not reach the protected runtime store:\n%s", shimPayload)
	}
	again := draftRealInstall(t, project, home, Options{})
	if again.Status != "ok" {
		t.Fatalf("reinstall = %+v", again)
	}
	if !strings.Contains(strings.Join(again.Messages, "\n"), "up-to-date") {
		t.Fatalf("reinstall did not report up-to-date installations: %q", again.Messages)
	}
}

// TestDraftGitInstallCarriesRegistryAttestation proves the recorded
// attestation equals exactly the registry evidence the effective plan
// selected — registry, status and key presence — and that changed
// evidence makes the installation non-current so the reinstall records
// the new summary instead of trusting the old one.
func TestDraftGitInstallCarriesRegistryAttestation(t *testing.T) {
	project, home, _, _ := setupGitInstall(t)
	attested := &marker.Attestation{Registry: "example.org/kit", Status: "audited", KeyID: "0123456789abcdef"}
	attest := func(attestation *marker.Attestation) func([]*closure.Node) (map[string]*marker.Attestation, []string, error) {
		return func([]*closure.Node) (map[string]*marker.Attestation, []string, error) {
			return map[string]*marker.Attestation{"review": attestation}, nil, nil
		}
	}
	installed := filepath.Join(project, ".agents", "skills", "review")
	if result := draftRealInstall(t, project, home, Options{ResolveAttest: attest(attested)}); result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}
	recorded := marker.Read(installed)
	if recorded == nil || recorded.SchemaVersion != marker.SchemaV5 {
		t.Fatalf("marker = %+v, want schema 5", recorded)
	}
	if recorded.Attestation == nil || *recorded.Attestation != *attested {
		t.Fatalf("attestation = %+v, want exactly %+v", recorded.Attestation, attested)
	}
	// Key absence is significant: evidence without a key records no key.
	unkeyed := &marker.Attestation{Registry: "example.org/kit", Status: "deprecated"}
	if result := draftRealInstall(t, project, home, Options{ResolveAttest: attest(unkeyed)}); result.Status != "ok" {
		t.Fatalf("reinstall = %+v", result)
	}
	reread := marker.Read(installed)
	if reread == nil || reread.Attestation == nil || *reread.Attestation != *unkeyed {
		t.Fatalf("attestation after evidence change = %+v, want exactly %+v", reread.Attestation, unkeyed)
	}
	payload, err := os.ReadFile(filepath.Join(installed, marker.Name))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	var attestationRaw map[string]json.RawMessage
	if err := json.Unmarshal(raw["attestation"], &attestationRaw); err != nil {
		t.Fatal(err)
	}
	if _, present := attestationRaw["key_id"]; present {
		t.Fatalf("marker records a key the evidence does not carry: %s", raw["attestation"])
	}
}

// TestDraftInstallRevokedEvidenceRefusesDespiteAttestedMarker proves a
// recorded attestation never authorizes: when the registry evidence is
// revoked the reinstall is refused and the prior marker is preserved
// byte-identically instead of becoming an unattested success.
func TestDraftInstallRevokedEvidenceRefusesDespiteAttestedMarker(t *testing.T) {
	project, home, _, _ := setupGitInstall(t)
	attested := &marker.Attestation{Registry: "example.org/kit", Status: "audited"}
	installed := filepath.Join(project, ".agents", "skills", "review")
	first := draftRealInstall(t, project, home, Options{ResolveAttest: func([]*closure.Node) (map[string]*marker.Attestation, []string, error) {
		return map[string]*marker.Attestation{"review": attested}, nil, nil
	}})
	if first.Status != "ok" {
		t.Fatalf("install = %+v", first)
	}
	before, err := os.ReadFile(filepath.Join(installed, marker.Name))
	if err != nil {
		t.Fatal(err)
	}
	revoked := errors.New("review is revoked by example.org/kit")
	again := draftRealInstall(t, project, home, Options{ResolveAttest: func([]*closure.Node) (map[string]*marker.Attestation, []string, error) {
		return nil, nil, revoked
	}})
	if again.Status != "failed" || !strings.Contains(strings.Join(again.Errors, ";"), "revoked") {
		t.Fatalf("reinstall = %+v, want the revocation refusal", again)
	}
	after, err := os.ReadFile(filepath.Join(installed, marker.Name))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("refused reinstall rewrote the attested marker")
	}
}

// TestDraftLocalRejectsConfusedAttestation proves the install lane fails
// closed when the registry backend selects evidence for a local snapshot:
// local content has no network attestation identity, so the run is
// refused with the arm diagnostic and stages nothing live.
func TestDraftLocalRejectsConfusedAttestation(t *testing.T) {
	project, home, _ := draftProject(t, draftLocalCollectionPayload, nil)
	writeDraftPackage(t, filepath.Join(project, "skills", "review"), "review")
	resolveDraftForInstall(t, project, home, draftLocalCollectionPayload)
	result := draftRealInstall(t, project, home, Options{ResolveAttest: func([]*closure.Node) (map[string]*marker.Attestation, []string, error) {
		return map[string]*marker.Attestation{"review": {Registry: "example.org/kit", Status: "audited"}}, nil, nil
	}})
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "source_member_invalid") {
		t.Fatalf("install = %+v, want the source_member_invalid refusal", result)
	}
	if !strings.Contains(strings.Join(result.Errors, ";"), "cannot carry a registry attestation") {
		t.Fatalf("install errors = %q, want the arm diagnostic", result.Errors)
	}
	if _, err := os.Lstat(filepath.Join(project, ".agents", "skills", "review")); !os.IsNotExist(err) {
		t.Fatalf("refused install staged a live installation")
	}
}

// TestDraftMarkerArmGuards proves buildDraftMarker carries the Git
// evidence summaries on the package binding and rejects both summaries
// for the local arm. The local rows are unreachable through
// install.Project — the closure and install gates refuse substitutions
// first and local nodes resolve no registry identity — so the builder is
// the narrowest honest seam.
func TestDraftMarkerArmGuards(t *testing.T) {
	node := func(substituted string) *closure.Node {
		return &closure.Node{
			Name:        "review",
			Spec:        &skillspec.Spec{SchemaVersion: 4, Commands: map[string]skillspec.Command{}, Dependencies: map[string]skillspec.CommandDependency{}, Requirements: map[string]skillspec.Requirement{}},
			Edges:       []closure.Edge{{Consumer: closure.ProjectEdge, Mode: "full"}},
			Substituted: substituted,
		}
	}
	gitMember := func() sourcelock.Member {
		pkg, err := sourcelock.NetworkGitPackage("example.org/kit",
			sourcelock.Commit{ObjectFormat: "sha1", Hex: strings.Repeat("cd", 20)}, "skills/review")
		if err != nil {
			t.Fatal(err)
		}
		return sourcelock.Member{Name: "review", Package: pkg}
	}
	localMember := func() sourcelock.Member {
		pkg, err := sourcelock.LocalPackage("sha256:" + strings.Repeat("ab", 32))
		if err != nil {
			t.Fatal(err)
		}
		return sourcelock.Member{Name: "review", Package: pkg}
	}
	attestation := &marker.Attestation{Registry: "example.org/kit", Status: "audited"}
	lockSHA := "sha256:" + strings.Repeat("ef", 32)
	assertNoLegacyIdentity := func(t *testing.T, m *marker.Marker) {
		t.Helper()
		if m.Source != "" || m.Git != "" || m.RefKind != "" || m.Ref != "" || m.Commit != "" {
			t.Fatalf("Git marker carries legacy identity %q %q %s %s %s",
				m.Source, m.Git, m.RefKind, m.Ref, m.Commit)
		}
	}

	carried, err := buildDraftMarker(node(""), gitMember(), lockSHA, "en", []string{"claude_code"}, nil, nil, attestation, nil, nil)
	if err != nil {
		t.Fatalf("Git marker with attestation = %v", err)
	}
	if carried.Attestation == nil || *carried.Attestation != *attestation {
		t.Fatalf("Git attestation = %+v, want %+v", carried.Attestation, attestation)
	}
	assertNoLegacyIdentity(t, carried)
	if carried.Package == nil || carried.Package.Kind != sourcelock.KindNetworkGit {
		t.Fatalf("Git marker = %+v, want the package binding", carried.Package)
	}
	substituted, err := buildDraftMarker(node("dev:review"), gitMember(), lockSHA, "en", []string{"claude_code"}, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("Git marker with substitution = %v", err)
	}
	if substituted.Substituted != "dev:review" {
		t.Fatalf("Git substituted = %q, want the operator identifier", substituted.Substituted)
	}
	assertNoLegacyIdentity(t, substituted)
	if _, err := buildDraftMarker(node(""), localMember(), lockSHA, "en", []string{"claude_code"}, nil, nil, attestation, nil, nil); err == nil {
		t.Fatalf("local marker admitted a registry attestation")
	} else if !strings.Contains(err.Error(), "source_member_invalid") {
		t.Fatalf("local attestation error = %v, want source_member_invalid", err)
	}
	if _, err := buildDraftMarker(node("dev:review"), localMember(), lockSHA, "en", []string{"claude_code"}, nil, nil, nil, nil, nil); err == nil {
		t.Fatalf("local marker admitted a development substitution")
	} else if !strings.Contains(err.Error(), "source_member_invalid") {
		t.Fatalf("local substitution error = %v, want source_member_invalid", err)
	}
}

// TestDraftSelectorSubstitutionForbidden proves a new from selector can
// never enable a legacy development substitution: an operator
// substitution manifest alongside a schema-2 project refuses the install
// with nothing published, and strict audit refuses even earlier.
func TestDraftSelectorSubstitutionForbidden(t *testing.T) {
	writeSubstitutions := func(t *testing.T, project string) {
		t.Helper()
		content := `{"substitutions": {"review": {"path": ` + quoteJSONPath(t.TempDir()) + "}}}\n"
		if err := os.WriteFile(devsub.PathIn(project), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		// The install gitignore gate runs before substitution planning:
		// the operator file must be ignored like any generated path.
		ignore, err := os.ReadFile(filepath.Join(project, ".gitignore"))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(project, ".gitignore"), append(ignore, []byte(devsub.Name+"\n")...), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Run("draft-selectors-refuse", func(t *testing.T) {
		project, home, _ := draftProject(t, draftLocalCollectionPayload, nil)
		writeDraftPackage(t, filepath.Join(project, "skills", "review"), "review")
		resolveDraftForInstall(t, project, home, draftLocalCollectionPayload)
		writeSubstitutions(t, project)
		result := draftRealInstall(t, project, home, Options{})
		if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "development substitutions are not admitted with draft selectors") {
			t.Fatalf("install = %+v, want the draft substitution refusal", result)
		}
		if _, err := os.Lstat(filepath.Join(project, ".agents", "skills")); !os.IsNotExist(err) {
			t.Fatalf("refused install published skills state")
		}
	})
	t.Run("strict-refuses-first", func(t *testing.T) {
		project, home, _ := draftProject(t, draftLocalCollectionPayload, nil)
		writeDraftPackage(t, filepath.Join(project, "skills", "review"), "review")
		resolveDraftForInstall(t, project, home, draftLocalCollectionPayload)
		writeSubstitutions(t, project)
		cfg := draftTestConfig(home, t.TempDir())
		cfg.Audit.Enabled = true
		cfg.Audit.Mode = "strict"
		result := Project(cfg, project, "test", Options{DraftSourcesV1: true, Platform: installPlatform()})
		if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "strict audit refuses substituted installs") {
			t.Fatalf("install = %+v, want the strict refusal", result)
		}
		if _, err := os.Lstat(filepath.Join(project, ".agents", "skills")); !os.IsNotExist(err) {
			t.Fatalf("refused install published skills state")
		}
	})
}

// TestDraftGitBuildsPublishReceipt3 proves the receipt binding migrated
// with the Git marker: a network-git member with a local go-v1 command
// and an external go-repository-v1 command publishes both under the
// distinct receipt-3 namespaces, the marker records bind receipt version
// 3 with every retained field, and a reinstall takes exact cache hits
// without compiling again.
func TestDraftGitBuildsPublishReceipt3(t *testing.T) {
	project, home := setupGitBuildInstall(t)
	deps, _, builder := draftBuildDeps(t)
	result := draftBuildInstall(t, home, project, deps, draftExternalDeps(t))
	if result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}
	lock, err := sourcelock.Read(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lock.Find("tooling")
	if !ok {
		t.Fatal("lock misses tooling")
	}
	wantPackage, err := member.Package.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	recorded := marker.Read(filepath.Join(project, ".agents", "skills", "tooling"))
	if recorded == nil || recorded.SchemaVersion != marker.SchemaV5 || recorded.Package == nil ||
		recorded.Package.Kind != sourcelock.KindNetworkGit {
		t.Fatalf("marker = %+v, want schema 5 network-git", recorded)
	}
	if len(recorded.Builds) != 2 {
		t.Fatalf("marker builds = %+v, want both arms", recorded.Builds)
	}
	local, external := recorded.Builds["ltool"], recorded.Builds["etool"]
	if local.Driver != buildmeta.DriverGoV1 || local.ReceiptSchemaVersion != 3 || local.ExecutionPolicy != buildmeta.ExecutionPolicy {
		t.Fatalf("local marker record = %+v", local)
	}
	if external.Driver != "go-repository-v1" || external.ReceiptSchemaVersion != 3 || external.Repository != "tools" ||
		external.DeclaredLockedCommit == nil || external.DeclaredLockedCommit.Hex != externalLockedCommit || external.Substituted {
		t.Fatalf("external marker record lost retained fields: %+v", external)
	}
	if len(dirNames(t, filepath.Join(home, "cache", "build", buildcache.Namespace))) != 0 {
		t.Fatalf("Git local build published outside the receipt-3 namespace")
	}
	sourceAwareLocal := dirNames(t, filepath.Join(home, "cache", "build", buildcache.SourceAwareNamespace))
	if len(sourceAwareLocal) != 1 || "sha256:"+sourceAwareLocal[0] != string(local.CacheKey) {
		t.Fatalf("local receipt-3 entries = %v, want the marker key %s", sourceAwareLocal, local.CacheKey)
	}
	localReceipt := readReceiptObject(t, filepath.Join(home, "cache", "build", buildcache.SourceAwareNamespace, sourceAwareLocal[0], buildcache.ReceiptFilename))
	localInput := localReceipt["input"].(map[string]any)
	if !jsonEqualCanonical(t, localInput["package"], wantPackage) {
		t.Fatalf("local receipt package does not equal the Git lock package")
	}
	externalRoot := filepath.Join(home, "external-build-cache")
	if len(dirNames(t, filepath.Join(externalRoot, buildrepo.ArtifactsDir(buildrepo.LegacyReceiptSchemaVersion)))) != 0 {
		t.Fatalf("Git external build published outside the receipt-3 namespace")
	}
	sourceAwareExternal := dirNames(t, filepath.Join(externalRoot, buildrepo.ArtifactsDir(buildrepo.SourceAwareReceiptSchemaVersion)))
	if len(sourceAwareExternal) != 1 || "sha256:"+sourceAwareExternal[0] != string(external.CacheKey) {
		t.Fatalf("external receipt-3 entries = %v, want the marker key %s", sourceAwareExternal, external.CacheKey)
	}
	if strings.Join(builder.calls, ",") != "ltool,etool" && strings.Join(builder.calls, ",") != "etool,ltool" {
		t.Fatalf("builder calls = %v", builder.calls)
	}
	again := draftBuildInstall(t, home, project, deps, draftExternalDeps(t))
	if again.Status != "ok" {
		t.Fatalf("reinstall = %+v", again)
	}
	messages := strings.Join(again.Messages, "\n")
	if !strings.Contains(messages, "tooling.ltool build source=") || !strings.Contains(messages, "key="+string(local.CacheKey)+" outcome=cache-hit") ||
		!strings.Contains(messages, "tooling.etool external build key="+string(external.CacheKey)+" outcome=cache-hit") {
		t.Fatalf("reinstall did not take exact receipt-3 cache hits: %q", again.Messages)
	}
}

func jsonEqualCanonical(t *testing.T, value any, want []byte) bool {
	t.Helper()
	return string(canonicalOf(t, value)) == string(want)
}
