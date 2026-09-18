package install

import (
	"fmt"
	"time"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/audit"
	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/registry"
	"github.com/relux-works/curator/internal/scriptpolicy"
	"github.com/relux-works/curator/internal/sourcelock"
)

// checkDraftSourceAudit binds draft members to the existing assurance gates
// (skillfile-sources §4) before any cache or compiler work. It supplements
// the shared audit gate, registry resolution, and build assurance preflight
// that run later in the same order for every lane; nothing those gates
// refused can pass here.
//
// Per member, in order:
//
//  1. registry: a local snapshot under a strict registry policy fails —
//     local content has no network attestation identity and the registry
//     never forges one (source-audit objects are not attestations);
//  2. source audit: the machine-local source-audit-v1 binding is validated
//     (or established on the mutating path) with the authorized pins,
//     revocations, and effective script/assurance policy labels in force.
//
// The frozen v1 lane never reaches this function: it runs only for a
// schema-2 lock consumed on the draft lane.
func checkDraftSourceAudit(cfg *config.Config, nodes []*closure.Node, lock *sourcelock.Lock, dryRun bool, now time.Time) ([]string, error) {
	scriptLabels := scriptpolicy.EffectiveLabels()
	assuranceLabels := artifactpolicy.EffectiveLabels()
	var warnings []string
	for _, node := range nodes {
		member, ok := lock.Find(node.Name)
		if !ok {
			return warnings, fmt.Errorf("source_member_missing: %s is not a locked member", node.Name)
		}
		if err := registry.CheckLocalPackage(node.Name, cfg.Audit.RegistryPolicy, member.Package.Kind); err != nil {
			return warnings, err
		}
		subject := audit.SourceSubject{
			Name: node.Name, Source: node.Decl.Source, Git: node.Decl.Git,
			Commit: node.Resolved.Commit, Snapshot: node.Snapshot,
			SchemaVersion: node.Spec.SchemaVersion, Capabilities: node.Spec.Capabilities,
			Package: toSourceAuditPackage(member.Package), ContentSHA256: member.ContentSHA256,
			ScriptLabels: scriptLabels, AssuranceLabels: assuranceLabels,
		}
		nodeWarnings, err := audit.CheckSourceAudit(cfg, subject, !dryRun, now)
		warnings = append(warnings, nodeWarnings...)
		if err != nil {
			return warnings, err
		}
	}
	return warnings, nil
}

// toSourceAuditPackage translates a locked package identity into the
// source-audit arm. The translation is exact: the local arm carries no
// commit, so its marshalled shape matches the source-types-v1 arms.
func toSourceAuditPackage(pkg sourcelock.Package) audit.SourcePackage {
	translated := audit.SourcePackage{
		Kind:       pkg.Kind,
		Snapshot:   pkg.Snapshot,
		Repository: pkg.Repository,
		Source:     pkg.Source,
		Directory:  pkg.Directory,
	}
	if pkg.IsGit() {
		translated.Commit = &audit.SourceCommit{ObjectFormat: pkg.Commit.ObjectFormat, Hex: pkg.Commit.Hex}
	}
	return translated
}
