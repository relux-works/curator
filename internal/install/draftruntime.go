package install

import (
	"fmt"
	"sort"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildsource"
	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/sourcelock"
)

// draftRuntimeKeys derives the protected runtime-store leaf of every locked
// draft member from its frozen package identity (skillfile-sources §4).
// The key is SHA-256 over the CCJ-1 bytes of the locked package in the
// distinct source-v1 namespace — never the bare snapshot digest and never
// a Git commit substituted into the other arm. A runtime-only refresh
// changes the package, so it changes the key and requires new runtime
// publication instead of reusing the previous tree. The marker and the
// runtime key migrated together: every draft member records marker schema
// 5, so every draft runtime leaf is namespaced and GC marks them from the
// recorded package.
func draftRuntimeKeys(lock *sourcelock.Lock) (map[string]string, error) {
	keys := make(map[string]string, len(lock.Members))
	for _, member := range lock.Members {
		digest, err := member.Package.Digest()
		if err != nil {
			return nil, err
		}
		key, err := runtimestore.SourceV1Key(digest)
		if err != nil {
			return nil, err
		}
		keys[member.Name] = key
	}
	return keys, nil
}

// draftBuildPackages selects the frozen package identity of every draft
// member: every draft installation is recorded under marker schema 5,
// exactly the members buildDraftMarker stages. Their compiled commands
// are keyed and receipted under the receipt-3 wrapper on both the local
// go-v1 and the external go-repository-v1 arm; the marker and the receipt
// binding migrated together, never one without the other. nil on the
// frozen v1 lane.
func draftBuildPackages(lock *sourcelock.Lock) map[string]*buildmeta.Package {
	if lock == nil {
		return nil
	}
	packages := make(map[string]*buildmeta.Package, len(lock.Members))
	for _, member := range lock.Members {
		packages[member.Name] = receiptPackage(member.Package)
	}
	return packages
}

// receiptPackage translates a locked package identity into the receipt
// model arm for arm. The translation is exact and lossless: the local arm
// carries no commit, so a snapshot digest can never enter the Git arm, and
// a commit can never enter the snapshot arm.
func receiptPackage(pkg sourcelock.Package) *buildmeta.Package {
	translated := &buildmeta.Package{
		Kind:       pkg.Kind,
		Snapshot:   pkg.Snapshot,
		Repository: pkg.Repository,
		Source:     pkg.Source,
		Directory:  pkg.Directory,
	}
	if pkg.IsGit() {
		translated.Commit = &buildmeta.PackageCommit{ObjectFormat: pkg.Commit.ObjectFormat, Hex: pkg.Commit.Hex}
	}
	return translated
}

// draftRuntimeKey selects the runtime-store leaf of one node: the frozen
// source-v1 key on the draft lane, the resolved commit on the frozen v1
// lane. An empty key map keeps the legacy behavior byte-identically.
func draftRuntimeKey(node *closure.Node, runtimeKeys map[string]string) string {
	if key, ok := runtimeKeys[node.Name]; ok && key != "" {
		return key
	}
	return node.Resolved.Commit
}

// draftMarkerPackage translates a locked package identity into the draft
// marker-v5 arm. The translation is exact: the local arm carries no commit,
// so a snapshot digest can never enter the Git arm, and a commit can never
// enter the snapshot arm.
func draftMarkerPackage(pkg sourcelock.Package) *marker.Package {
	translated := &marker.Package{
		Kind:       pkg.Kind,
		Snapshot:   pkg.Snapshot,
		Repository: pkg.Repository,
		Source:     pkg.Source,
		Directory:  pkg.Directory,
	}
	if pkg.IsGit() {
		translated.Commit = &marker.Commit{ObjectFormat: pkg.Commit.ObjectFormat, Hex: pkg.Commit.Hex}
	}
	return translated
}

// buildDraftMarker stages the install marker of one draft member: marker
// schema 5 with the frozen package and the binding lock, every retained
// field carried exactly like the legacy marker (skillfile-sources §4
// migration table). The package replaces the legacy source identity on
// every arm; the declared ref selection lives in the manifest and bound
// lock, never in the marker. A Git member carries the registry
// attestation the effective plan selected and the legacy
// development-substitution identifier the node carries. A local snapshot
// admits neither — local content has no network identity and no selector
// to substitute — so staging an attestation or substitution for one
// fails closed instead of writing a marker the reader must refuse. The
// summaries recorded here never authorize: every gate re-resolves its
// own evidence.
func buildDraftMarker(
	node *closure.Node,
	member sourcelock.Member,
	lockSHA256 string,
	effectiveLocale string,
	agents []string,
	activeCommands []string,
	mcp map[string][]string,
	attestation *marker.Attestation,
	builds map[string]marker.Build,
	source *buildsource.Identity,
) (*marker.Marker, error) {
	if member.Package.Kind == sourcelock.KindLocalSnapshot {
		if attestation != nil {
			return nil, fmt.Errorf("source_member_invalid: local-snapshot member %s cannot carry a registry attestation", node.Name)
		}
		if node.Substituted != "" {
			return nil, fmt.Errorf("source_member_invalid: local-snapshot member %s cannot carry a development substitution", node.Name)
		}
	}
	var commands []string
	for name, command := range node.Spec.Commands {
		if command.Type == "script" || command.Type == "build" {
			commands = append(commands, name)
		}
	}
	var dependencies []string
	for name := range node.Spec.Dependencies {
		dependencies = append(dependencies, name)
	}
	var requirements []string
	for name := range node.Spec.Requirements {
		requirements = append(requirements, name)
	}
	if builds == nil {
		builds = map[string]marker.Build{}
	}
	// Build state is all-or-nothing, exactly like the legacy marker: a
	// node with no published build carries no build source at all.
	// Marker-5 build entries bind receipt version 3 and the execution policy
	// for every driver, whatever the skill schema: the receipt the cache
	// holds for this member is the package wrapper, so no legacy record
	// shape can describe it (skillfile-sources §4).
	upgraded := make(map[string]marker.Build, len(builds))
	for command, build := range builds {
		build.ReceiptSchemaVersion = buildmeta.SourceAwareSchemaVersion
		build.ExecutionPolicy = buildmeta.ExecutionPolicy
		upgraded[command] = build
	}
	builds = upgraded
	if len(builds) == 0 {
		source = nil
	}
	buildRoots := append([]string(nil), node.Spec.BuildRoots...)
	sort.Strings(buildRoots)
	expected := &marker.Marker{
		Name:               node.Name,
		Package:            draftMarkerPackage(member.Package),
		LockSHA256:         lockSHA256,
		Locale:             effectiveLocale,
		Agents:             agents,
		Commands:           commands,
		Dependencies:       dependencies,
		SkillSchemaVersion: node.Spec.SchemaVersion,
		RuntimeRoots:       node.Spec.RuntimeRoots,
		BuildRoots:         buildRoots,
		BuildSource:        source,
		Builds:             builds,
		Requirements:       requirements,
		McpServers:         mcp,
		Attestation:        attestation,
		Activation:         &marker.Activation{Context: node.ContextActive(), Commands: activeCommands},
		Requirers:          node.Consumers(),
		Substituted:        node.Substituted,
	}
	if !node.ContextActive() {
		expected.Locale = ""
		expected.Agents = []string{}
	}
	return expected, nil
}
