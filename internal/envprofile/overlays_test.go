package envprofile

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextresolve"
)

func itoa(n int) string { return strconv.Itoa(n) }

// chapterOrder reads the emitted root-context document and returns the
// chapter member names in order.
func chapterOrder(t *testing.T, path string) []string {
	t.Helper()
	payload, err := os.ReadFile(path) // #nosec G304 -- test home file
	if err != nil {
		t.Fatal(err)
	}
	var order []string
	for _, line := range strings.Split(string(payload), "\n") {
		if name, ok := strings.CutPrefix(line, "## Context: "); ok {
			fields := strings.Fields(name)
			if len(fields) == 0 {
				t.Fatalf("chapter line %q carries no member", line)
			}
			order = append(order, fields[0])
		}
	}
	return order
}

// installComposedPair installs a path root with a path overlay through the
// production Install and returns the manager home and profile name.
func installComposedPair(t *testing.T) (string, string) {
	t.Helper()
	home := t.TempDir()
	pinHomes(t)
	root := filepath.Join(t.TempDir(), "root")
	writeManifestPackage(t, root,
		`{"schema_version": 1, "name": "acme", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "root\n"})
	overlay := filepath.Join(t.TempDir(), "overlay")
	writeManifestPackage(t, overlay,
		`{"schema_version": 1, "name": "personal", "version": "0.3.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "overlay\n"})
	base := Policy{
		OverlaysAllowed:      true,
		OverlayDefaultWeight: 1000,
		Overlays:             map[string][]OverlaySpec{"acme": {{Source: overlay}}},
	}
	if _, _, _, err := Install(home, InstallOptions{Operand: root, Policy: base}); err != nil {
		t.Fatal(err)
	}
	return home, "acme"
}

// TestPrecedencePrimitivesDriveEmission drives the production switch under
// all four precedence pairs: winner and placement each flip the emitted
// chapter order independently of the other.
func TestPrecedencePrimitivesDriveEmission(t *testing.T) {
	cases := []struct {
		winner    string
		placement string
		want      []string
	}{
		{"higher-weight", "winner-last", []string{"acme", "personal"}},
		{"higher-weight", "winner-first", []string{"personal", "acme"}},
		{"lower-weight", "winner-last", []string{"personal", "acme"}},
		{"lower-weight", "winner-first", []string{"acme", "personal"}},
	}
	for _, tc := range cases {
		t.Run(tc.winner+"/"+tc.placement, func(t *testing.T) {
			home, profile := installComposedPair(t)
			// The helper already resolved the weighed lock; the switch
			// reads it from the store and emits under the candidate
			// precedence pair.
			policy := Policy{
				OverlaysAllowed:      true,
				OverlayDefaultWeight: 1000,
				PrecedenceWinner:     tc.winner,
				PrecedencePlacement:  tc.placement,
			}
			results, err := UseWithPolicy(home, profile, "", "", false, policy)
			if err != nil {
				t.Fatal(err)
			}
			for _, result := range results {
				if !result.OK {
					t.Fatalf("%s: %s", result.Adapter, result.Detail)
				}
			}
			got := chapterOrder(t, filepath.Join(os.Getenv("CLAUDE_CONFIG_DIR"), "CLAUDE.md"))
			if strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Fatalf("chapters %v, want %v", got, tc.want)
			}
		})
	}
}

// TestPrecedenceTieKeepsTopologicalOrder drives the production switch with
// equal weights under both placements: placement never inverts a tie.
func TestPrecedenceTieKeepsTopologicalOrder(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	root := filepath.Join(t.TempDir(), "root")
	writeManifestPackage(t, root,
		`{"schema_version": 1, "name": "acme", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "root\n"})
	overlay := filepath.Join(t.TempDir(), "overlay")
	writeManifestPackage(t, overlay,
		`{"schema_version": 1, "name": "personal", "version": "0.3.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "overlay\n"})
	zero := int64(0)
	base := Policy{
		OverlaysAllowed:      true,
		OverlayDefaultWeight: 1000,
		Overlays:             map[string][]OverlaySpec{"acme": {{Source: overlay, Weight: &zero}}},
	}
	if _, _, _, err := Install(home, InstallOptions{Operand: root, Policy: base}); err != nil {
		t.Fatal(err)
	}
	var orders []string
	for _, placement := range []string{"winner-last", "winner-first"} {
		policy := base
		policy.PrecedenceWinner = "higher-weight"
		policy.PrecedencePlacement = placement
		results, err := UseWithPolicy(home, "acme", "", "", false, policy)
		if err != nil {
			t.Fatal(err)
		}
		for _, result := range results {
			if !result.OK {
				t.Fatalf("%s: %s", result.Adapter, result.Detail)
			}
		}
		orders = append(orders, strings.Join(chapterOrder(t,
			filepath.Join(os.Getenv("CLAUDE_CONFIG_DIR"), "CLAUDE.md")), ","))
	}
	if orders[0] != orders[1] || orders[0] != "acme,personal" {
		t.Fatalf("tied orders %v, want [acme,personal] under both placements", orders)
	}
}

// Production entry points under test: Install, UpdateWithPolicy,
// UseWithPolicy, Remove.

// writeManifestPackage writes a context package with an explicit manifest
// and one module per entry of modules. Every path joins through
// filepath.Join over the temporary base, so no fixture carries a
// platform-absolute literal.
func writeManifestPackage(t *testing.T, root, manifest string, modules map[string]string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "context"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "agent-context.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	for name, content := range modules {
		full := filepath.Join(root, "context", filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func lockMember(lock *contextlock.Lock, name string) (contextlock.Member, bool) {
	for _, member := range lock.Members {
		if member.Name == name {
			return member, true
		}
	}
	return contextlock.Member{}, false
}

// TestPathOverlayJoinsClosure drives the production Install with a
// machine-declared path overlay: the overlay joins the closure beside the
// root, the lock flags it overlay, and its weight is the machine default.
func TestPathOverlayJoinsClosure(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	root := filepath.Join(t.TempDir(), "root")
	writeManifestPackage(t, root,
		`{"schema_version": 1, "name": "acme", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "root\n"})
	overlay := filepath.Join(t.TempDir(), "overlay")
	writeManifestPackage(t, overlay,
		`{"schema_version": 1, "name": "personal", "version": "0.3.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "overlay\n"})
	policy := Policy{
		OverlaysAllowed:      true,
		OverlayDefaultWeight: 1000,
		Overlays:             map[string][]OverlaySpec{"acme": {{Source: overlay}}},
	}
	info, _, _, err := Install(home, InstallOptions{Operand: root, Policy: policy})
	if err != nil {
		t.Fatal(err)
	}
	if info.Name != "acme" {
		t.Fatalf("profile %q", info.Name)
	}
	member, ok := lockMember(info.Lock, "personal")
	if !ok {
		t.Fatalf("lock members %+v carry no overlay", info.Lock.Members)
	}
	if !member.Overlay {
		t.Fatalf("overlay member %+v is not flagged overlay", member)
	}
	if member.Weight != 1000 {
		t.Fatalf("overlay weight %d, want the machine default 1000", member.Weight)
	}
	if member.StateHash == "" || member.Commit != "" {
		t.Fatalf("path overlay member %+v carries no state pin", member)
	}
	if len(member.RequiredBy) != 0 {
		t.Fatalf("overlay required_by %+v, want empty", member.RequiredBy)
	}
	rootMember, ok := lockMember(info.Lock, "acme")
	if !ok || rootMember.Overlay {
		t.Fatalf("root member %+v", rootMember)
	}
	// The overlay is joint resolution, not a union: update keeps it.
	updated, moved, err := UpdateWithPolicy(home, "acme", policy)
	if err != nil {
		t.Fatal(err)
	}
	if moved {
		t.Fatal("unchanged closure must not move the lock")
	}
	if _, ok := lockMember(updated.Lock, "personal"); !ok {
		t.Fatalf("update dropped the overlay: %+v", updated.Lock.Members)
	}
}

// TestOverlayExplicitWeightOverridesDefault drives Install with a declared
// weight: rule 4 outranks the manifest weight.
func TestOverlayExplicitWeightOverridesDefault(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	root := filepath.Join(t.TempDir(), "root")
	writeManifestPackage(t, root,
		`{"schema_version": 1, "name": "acme", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "root\n"})
	overlay := filepath.Join(t.TempDir(), "overlay")
	writeManifestPackage(t, overlay,
		`{"schema_version": 1, "name": "personal", "version": "0.3.0", "weight": 5,`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "overlay\n"})
	weight := int64(700)
	policy := Policy{
		OverlaysAllowed:      true,
		OverlayDefaultWeight: 1000,
		Overlays:             map[string][]OverlaySpec{"acme": {{Source: overlay, Weight: &weight}}},
	}
	info, _, _, err := Install(home, InstallOptions{Operand: root, Policy: policy})
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lockMember(info.Lock, "personal")
	if !ok {
		t.Fatalf("lock members %+v carry no overlay", info.Lock.Members)
	}
	if member.Weight != 700 {
		t.Fatalf("overlay weight %d, want the declared 700 over manifest 5 and default 1000", member.Weight)
	}
}

// TestGitOverlayJoinsClosure drives the production Install with a
// machine-declared git overlay: the overlay resolves jointly with the
// root and the lock flags it overlay with the declared weight.
func TestGitOverlayJoinsClosure(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	ids := newGitIdentities(t)
	overlayOperand := gitContextRepo(t, ids, "https://example.com/personal",
		`{"schema_version": 1, "name": "personal", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`,
		map[string]string{"a.md": "overlay\n"})
	rootOperand := gitContextRepo(t, ids, "https://example.com/root",
		`{"schema_version": 1, "name": "root", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`,
		map[string]string{"a.md": "root\n"})
	weight := int64(250)
	policy := Policy{
		OverlaysAllowed:      true,
		OverlayDefaultWeight: 1000,
		Overlays: map[string][]OverlaySpec{
			"root": {{Source: overlayOperand, Range: "*", Weight: &weight}},
		},
	}
	info, _, _, err := Install(home, InstallOptions{Operand: rootOperand, Policy: policy})
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lockMember(info.Lock, "personal")
	if !ok || !member.Overlay {
		t.Fatalf("lock members %+v carry no flagged overlay", info.Lock.Members)
	}
	if member.Weight != 250 {
		t.Fatalf("overlay weight %d, want the declared 250", member.Weight)
	}
	if member.Commit == "" || member.Source != "example.com/personal" {
		t.Fatalf("git overlay member %+v carries no commit pin", member)
	}
}

// TestPathOverlayFormIsSourceInvalid drives Install with a path overlay
// carrying an exact form: reader grammar accepts it (§12.1), but the
// section 1 declaration rule refuses it at resolution with
// profile_source_invalid.
func TestPathOverlayFormIsSourceInvalid(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	root := filepath.Join(t.TempDir(), "root")
	writeManifestPackage(t, root,
		`{"schema_version": 1, "name": "acme", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "root\n"})
	overlay := filepath.Join(t.TempDir(), "overlay")
	writeManifestPackage(t, overlay,
		`{"schema_version": 1, "name": "personal", "version": "0.3.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "overlay\n"})
	policy := Policy{
		OverlaysAllowed:      true,
		OverlayDefaultWeight: 1000,
		Overlays: map[string][]OverlaySpec{
			"acme": {{Source: overlay, Revision: strings.Repeat("ab", 20)}},
		},
	}
	_, _, _, err := Install(home, InstallOptions{Operand: root, Policy: policy})
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("err = %v, want %s", err, DiagSourceInvalid)
	}
}

// TestOverlayRepeatedNameIsCompositionInvalid drives Install with two
// overlays carrying one package name: the repeated declaration is
// environment_composition_invalid.
func TestOverlayRepeatedNameIsCompositionInvalid(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	root := filepath.Join(t.TempDir(), "root")
	writeManifestPackage(t, root,
		`{"schema_version": 1, "name": "acme", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "root\n"})
	first := filepath.Join(t.TempDir(), "first")
	writeManifestPackage(t, first,
		`{"schema_version": 1, "name": "personal", "version": "0.3.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "first\n"})
	second := filepath.Join(t.TempDir(), "second")
	writeManifestPackage(t, second,
		`{"schema_version": 1, "name": "personal", "version": "0.4.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "second\n"})
	policy := Policy{
		OverlaysAllowed:      true,
		OverlayDefaultWeight: 1000,
		Overlays:             map[string][]OverlaySpec{"acme": {{Source: first}, {Source: second}}},
	}
	_, _, _, err := Install(home, InstallOptions{Operand: root, Policy: policy})
	if err == nil || !strings.Contains(err.Error(), contextresolve.DiagCompositionInvalid) {
		t.Fatalf("err = %v, want %s", err, contextresolve.DiagCompositionInvalid)
	}
}

// TestOverlayDuplicateNameIsCompositionInvalid drives Install with an
// overlay whose package name repeats the root: the declaration is
// environment_composition_invalid.
func TestOverlayDuplicateNameIsCompositionInvalid(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	root := filepath.Join(t.TempDir(), "root")
	writeManifestPackage(t, root,
		`{"schema_version": 1, "name": "acme", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "root\n"})
	overlay := filepath.Join(t.TempDir(), "overlay")
	writeManifestPackage(t, overlay,
		`{"schema_version": 1, "name": "acme", "version": "9.9.9",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "overlay\n"})
	policy := Policy{
		OverlaysAllowed:      true,
		OverlayDefaultWeight: 1000,
		Overlays:             map[string][]OverlaySpec{"acme": {{Source: overlay}}},
	}
	_, _, _, err := Install(home, InstallOptions{Operand: root, Policy: policy})
	if err == nil || !strings.Contains(err.Error(), contextresolve.DiagCompositionInvalid) {
		t.Fatalf("err = %v, want %s", err, contextresolve.DiagCompositionInvalid)
	}
}

// TestForbiddenOverlaysResolveAlone drives Install with a forbidding
// composition policy: the declared overlay list empties and the root
// resolves alone.
func TestForbiddenOverlaysResolveAlone(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	root := filepath.Join(t.TempDir(), "root")
	writeManifestPackage(t, root,
		`{"schema_version": 1, "name": "acme", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "root\n"})
	overlay := filepath.Join(t.TempDir(), "overlay")
	writeManifestPackage(t, overlay,
		`{"schema_version": 1, "name": "personal", "version": "0.3.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "overlay\n"})
	policy := Policy{
		OverlaysAllowed:      false,
		OverlayDefaultWeight: 1000,
		Overlays:             map[string][]OverlaySpec{"acme": {{Source: overlay}}},
	}
	if !policy.ForbidsOverlays() {
		t.Fatal("forbidding policy must report ForbidsOverlays")
	}
	if len(policy.EffectiveOverlays("acme")) != 0 {
		t.Fatal("forbidding policy must empty every overlay list")
	}
	info, _, _, err := Install(home, InstallOptions{Operand: root, Policy: policy})
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Lock.Members) != 1 {
		t.Fatalf("lock members %+v, want the root alone", info.Lock.Members)
	}
}

// TestRemoveRefusesOverlayOfAnotherProfile drives Remove for a profile
// whose package is an overlay member of another installed profile's lock:
// the removal fails with profile_in_use.
func TestRemoveRefusesOverlayOfAnotherProfile(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	overlay := filepath.Join(t.TempDir(), "overlay")
	writeManifestPackage(t, overlay,
		`{"schema_version": 1, "name": "personal", "version": "0.3.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "overlay\n"})
	if _, _, _, err := Install(home, InstallOptions{Operand: overlay, As: "personal-profile"}); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(t.TempDir(), "root")
	writeManifestPackage(t, root,
		`{"schema_version": 1, "name": "acme", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "root\n"})
	policy := Policy{
		OverlaysAllowed:      true,
		OverlayDefaultWeight: 1000,
		Overlays:             map[string][]OverlaySpec{"acme": {{Source: overlay}}},
	}
	if _, _, _, err := Install(home, InstallOptions{Operand: root, Policy: policy}); err != nil {
		t.Fatal(err)
	}
	// Move the machine current off the overlay profile so the removal
	// refusal can only come from the overlay guard, not the scope guard.
	if _, err := UseWithPolicy(home, "acme", "", "", false, policy); err != nil {
		t.Fatal(err)
	}
	if err := Remove(home, "personal-profile", false); err == nil ||
		!strings.Contains(err.Error(), DiagInUse) {
		t.Fatalf("err = %v, want %s", err, DiagInUse)
	} else if !strings.Contains(err.Error(), "overlay") {
		t.Fatalf("err = %v, want the overlay reason, not the scope reason", err)
	}
}

// gitContextRepo builds a tagged git fixture repository carrying one
// context package and serves it under a fake canonical network identity,
// returning the install operand. Every path joins through filepath.Join
// over the temporary base.
func gitContextRepo(t *testing.T, ids *gitIdentities, identity, manifest string, modules map[string]string) string {
	t.Helper()
	files := map[string]string{"agent-context.json": manifest + "\n"}
	for name, content := range modules {
		files["context/"+name] = content
	}
	repo := gitRepo(t, files, "v1.0.0")
	return ids.serve(repo, identity)
}

func leafManifest(weightClause string) string {
	return `{"schema_version": 1, "name": "leaf", "version": "1.0.0",` + weightClause +
		`"context": {"modules": [{"path": "a.md"}]}}`
}

// TestWeightRulesApplyInOrder drives Install over git fixtures where the
// root requires one leaf: the manifest weight (rule 1), the agreeing edge
// weight (rule 2), and the root weights map (rule 3) each win in turn.
func TestWeightRulesApplyInOrder(t *testing.T) {
	cases := []struct {
		name       string
		leafWeight string
		edgeWeight string
		rootMap    string
		want       int64
	}{
		{"manifest", `"weight": 5,`, "", "", 5},
		{"edge", `"weight": 5,`, `, "weight": 60`, "", 60},
		{"rootMap", `"weight": 5,`, "", `"weights": {"leaf": 10},`, 10},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			pinHomes(t)
			ids := newGitIdentities(t)
			leafOperand := gitContextRepo(t, ids, "https://example.com/leaf",
				leafManifest(tc.leafWeight), map[string]string{"a.md": "leaf\n"})
			rootManifest := `{"schema_version": 1, "name": "root", "version": "1.0.0",` + tc.rootMap +
				`"requires": {"contexts": {"leaf": {"git": "` + leafOperand + `", "range": "*"` + tc.edgeWeight + `}}},` +
				`"context": {"modules": [{"path": "a.md"}]}}`
			rootOperand := gitContextRepo(t, ids, "https://example.com/root",
				rootManifest, map[string]string{"a.md": "root\n"})
			info, _, _, err := Install(home, InstallOptions{Operand: rootOperand})
			if err != nil {
				t.Fatal(err)
			}
			member, ok := lockMember(info.Lock, "leaf")
			if !ok {
				t.Fatalf("lock members %+v carry no leaf", info.Lock.Members)
			}
			if member.Weight != tc.want {
				t.Fatalf("leaf weight %d, want %d", member.Weight, tc.want)
			}
		})
	}
}

// TestWeightDuplicateRefused drives Install where the root names a package
// both on a requirement edge with a weight and in the weights map: the
// lock fails with context_weights_duplicate.
func TestWeightDuplicateRefused(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	ids := newGitIdentities(t)
	leafOperand := gitContextRepo(t, ids, "https://example.com/leaf",
		leafManifest(""), map[string]string{"a.md": "leaf\n"})
	rootManifest := `{"schema_version": 1, "name": "root", "version": "1.0.0",` +
		`"weights": {"leaf": 10},` +
		`"requires": {"contexts": {"leaf": {"git": "` + leafOperand + `", "range": "*", "weight": 60}}},` +
		`"context": {"modules": [{"path": "a.md"}]}}`
	rootOperand := gitContextRepo(t, ids, "https://example.com/root",
		rootManifest, map[string]string{"a.md": "root\n"})
	_, _, _, err := Install(home, InstallOptions{Operand: rootOperand})
	if err == nil || !strings.Contains(err.Error(), contextresolve.DiagWeightsDuplicate) {
		t.Fatalf("err = %v, want %s", err, contextresolve.DiagWeightsDuplicate)
	}
}

// TestWeightUnknownRefused drives Install where the root weights map names
// a package outside the closure: context_weight_unknown.
func TestWeightUnknownRefused(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	ids := newGitIdentities(t)
	leafOperand := gitContextRepo(t, ids, "https://example.com/leaf",
		leafManifest(""), map[string]string{"a.md": "leaf\n"})
	rootManifest := `{"schema_version": 1, "name": "root", "version": "1.0.0",` +
		`"weights": {"ghost": 10},` +
		`"requires": {"contexts": {"leaf": {"git": "` + leafOperand + `", "range": "*"}}},` +
		`"context": {"modules": [{"path": "a.md"}]}}`
	rootOperand := gitContextRepo(t, ids, "https://example.com/root",
		rootManifest, map[string]string{"a.md": "root\n"})
	_, _, _, err := Install(home, InstallOptions{Operand: rootOperand})
	if err == nil || !strings.Contains(err.Error(), contextresolve.DiagWeightUnknown) {
		t.Fatalf("err = %v, want %s", err, contextresolve.DiagWeightUnknown)
	}
}

// TestNonRootWeightsMapRefused drives Install where a non-root member
// carries a non-empty weights map: context_weights_not_root at resolution
// time, even though the snapshot itself validates.
func TestNonRootWeightsMapRefused(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	ids := newGitIdentities(t)
	leafManifest := `{"schema_version": 1, "name": "leaf", "version": "1.0.0",` +
		`"weights": {"leaf": 1},` +
		`"context": {"modules": [{"path": "a.md"}]}}`
	leafOperand := gitContextRepo(t, ids, "https://example.com/leaf",
		leafManifest, map[string]string{"a.md": "leaf\n"})
	rootManifest := `{"schema_version": 1, "name": "root", "version": "1.0.0",` +
		`"requires": {"contexts": {"leaf": {"git": "` + leafOperand + `", "range": "*"}}},` +
		`"context": {"modules": [{"path": "a.md"}]}}`
	rootOperand := gitContextRepo(t, ids, "https://example.com/root",
		rootManifest, map[string]string{"a.md": "root\n"})
	_, _, _, err := Install(home, InstallOptions{Operand: rootOperand})
	if err == nil || !strings.Contains(err.Error(), contextresolve.DiagWeightsNotRoot) {
		t.Fatalf("err = %v, want %s", err, contextresolve.DiagWeightsNotRoot)
	}
}

// TestWeightConflictErrorAndWarning drives Install where two direct
// requirers disagree on one member's edge weight: without a root map entry
// the disagreement is context_weight_conflict; with one the root has the
// final word and the same diagnostic is a warning.
func TestWeightConflictErrorAndWarning(t *testing.T) {
	// Two non-root requirers disagree on the leaf's edge weight; the root
	// carries no edge weight on the leaf, so no duplicate fires.
	build := func(t *testing.T, rootMap string) (Policy, string, string) {
		ids := newGitIdentities(t)
		leafOperand := gitContextRepo(t, ids, "https://example.com/leaf",
			leafManifest(""), map[string]string{"a.md": "leaf\n"})
		midOperand := func(name, identity string, weight int) string {
			manifest := `{"schema_version": 1, "name": "` + name + `", "version": "1.0.0",` +
				`"requires": {"contexts": {"leaf": {"git": "` + leafOperand +
				`", "range": "*", "weight": ` + itoa(weight) + `}}},` +
				`"context": {"modules": [{"path": "a.md"}]}}`
			return gitContextRepo(t, ids, identity, manifest, map[string]string{"a.md": name + "\n"})
		}
		mid := midOperand("mid", "https://example.com/mid", 2)
		mid2 := midOperand("mid2", "https://example.com/mid2", 3)
		rootManifest := `{"schema_version": 1, "name": "root", "version": "1.0.0",` + rootMap +
			`"requires": {"contexts": {` +
			`"mid": {"git": "` + mid + `", "range": "*"}, ` +
			`"mid2": {"git": "` + mid2 + `", "range": "*"}}},` +
			`"context": {"modules": [{"path": "a.md"}]}}`
		rootOperand := gitContextRepo(t, ids, "https://example.com/root",
			rootManifest, map[string]string{"a.md": "root\n"})
		return Policy{}, rootOperand, leafOperand
	}
	t.Run("error", func(t *testing.T) {
		home := t.TempDir()
		pinHomes(t)
		_, rootOperand, _ := build(t, "")
		_, _, _, err := Install(home, InstallOptions{Operand: rootOperand})
		if err == nil || !strings.Contains(err.Error(), contextresolve.DiagWeightConflict) {
			t.Fatalf("err = %v, want %s", err, contextresolve.DiagWeightConflict)
		}
	})
	t.Run("warning", func(t *testing.T) {
		home := t.TempDir()
		pinHomes(t)
		_, rootOperand, _ := build(t, ` "weights": {"leaf": 9},`)
		// The root manifest above declares the weights map entry, so the
		// disagreement resolves to a warning under the same diagnostic
		// and the leaf takes the root map weight.
		info, _, _, err := Install(home, InstallOptions{Operand: rootOperand})
		if err != nil {
			t.Fatal(err)
		}
		member, ok := lockMember(info.Lock, "leaf")
		if !ok || member.Weight != 9 {
			t.Fatalf("leaf member %+v, want weight 9 by the root's final word", member)
		}
		found := false
		for _, warning := range info.Warnings {
			if strings.Contains(warning, contextresolve.DiagWeightConflict) {
				found = true
			}
		}
		if !found {
			t.Fatalf("warnings %+v carry no %s", info.Warnings, contextresolve.DiagWeightConflict)
		}
	})
}
