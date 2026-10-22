package projectcontext

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/audit"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/skillspec"
)

func (op *operation) skill(r registered, req Request) (Candidate, error) {
	root, err := os.OpenRoot(r.Record.ProjectRoot)
	if err != nil {
		return Candidate{}, err
	}
	defer root.Close()
	capt, err := capture(root, req.Source)
	if err != nil {
		return Candidate{}, err
	}
	if err = validateInventory(capt); err != nil {
		return Candidate{}, err
	}
	if _, ok := capt.Bytes[skillspec.CanonicalManifestName]; !ok {
		return Candidate{}, fail("project_context_unsupported", "whole-skill admission requires an explicit portable manifest")
	}
	ref, h, err := op.put(capt)
	if err != nil {
		return Candidate{}, err
	}
	snapshot := filepath.Join(op.root, "snapshots", h, "files")
	pin, err := auditSkill(filepath.Dir(op.root), snapshot)
	if err != nil {
		return Candidate{}, err
	}
	item := Item{"project-item/1", req.ID, "skill", req.Environment, req.Required, Payload{SnapshotRef: ref, SnapshotSHA256: h}}
	return Candidate{item, identity(item), req, identity(capt.Manifest), []Member{{req.ID, ref, h, "state sha256:" + pin}}, AppendAfterProfile}, nil
}
func auditSkill(home, snapshot string) (string, error) {
	spec, err := skillspec.Load(snapshot)
	if err != nil {
		return "", err
	}
	// Executable admission never treats an arbitrary directory as a package.
	// Cross-package requirements need an owning pinned activation closure; this
	// first admission entry point refuses them instead of silently dropping it.
	if len(spec.Requirements) != 0 || len(spec.McpServers) != 0 || len(spec.BuildRepositories) != 0 {
		return "", fail("project_context_unsupported", "whole-skill transitive/external activation closure is required")
	}
	pin, err := hashing.ContentSHA256WithVersion(snapshot, nil, hashing.VersionV2)
	if err != nil {
		return "", err
	}
	if !audit.CanaryPasses() {
		return "", fail("project_context_invalid", "skill audit canary failed")
	}
	cfg := &config.Config{Path: filepath.Join(home, "config.json"), Audit: config.Audit{Enabled: true, Mode: "strict", FailOn: "high", Backend: "null"}}
	_, errs := audit.GateReadOnly(cfg, []audit.Subject{{Kind: "skill", Name: "project-item", Source: pin, Snapshot: snapshot, SchemaVersion: spec.SchemaVersion, Capabilities: spec.Capabilities, HashVersion: hashing.VersionV2}})
	if len(errs) != 0 {
		return "", fail("project_context_invalid", "whole-skill audit refused: "+strings.Join(errs, "; "))
	}
	return strings.TrimPrefix(pin, "sha256:"), nil
}
