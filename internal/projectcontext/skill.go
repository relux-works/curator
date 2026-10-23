package projectcontext

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/audit"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/hashing"
	skillmanifest "github.com/relux-works/curator/internal/manifest"
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
		return Candidate{}, fail("project_context_invalid", "whole-skill admission requires an explicit portable manifest")
	}
	ref, h, err := op.put(capt)
	if err != nil {
		return Candidate{}, err
	}
	snapshot := filepath.Join(op.root, "snapshots", h, "files")
	pin, err := auditSkill(filepath.Dir(op.root), snapshot)
	if err != nil {
		if isCode(err, "project_context_unsupported") {
			// Bind an optional omission to the exact validated retained input,
			// never to a second live read after validation.
			return Candidate{SourceSHA256: identity(capt.Manifest)}, err
		}
		return Candidate{}, err
	}
	item := Item{"project-item/1", req.ID, "skill", req.Environment, req.Required, Payload{SnapshotRef: ref, SnapshotSHA256: h}}
	return Candidate{item, identity(item), req, identity(capt.Manifest), []Member{{req.ID, ref, h, "state sha256:" + pin}}, AppendAfterProfile}, nil
}
func auditSkill(home, snapshot string) (string, error) {
	_, spec, err := skillmanifest.ReadSkillPackage(snapshot)
	if err != nil {
		return "", fail("project_context_invalid", "whole-skill package document or manifest is invalid")
	}
	if err := closedSkill(snapshot); err != nil {
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

// The owning parser checks modern closed manifests. Apply exact member
// closure to the legacy schema as well at this project-admission boundary.
func closedSkill(snapshot string) error {
	b, err := os.ReadFile(filepath.Join(snapshot, skillspec.CanonicalManifestName))
	if err != nil {
		return err
	}
	var obj map[string]json.RawMessage
	if err := decode(b, &obj); err != nil {
		return err
	}
	var schema int
	if err := json.Unmarshal(obj["schema_version"], &schema); err != nil {
		return fail("project_context_invalid", "invalid skill schema")
	}
	if schema != 1 {
		return nil
	} // Modern owning readers already enforce exact closure.
	allowed := map[string]bool{"schema_version": true, "commands": true}
	for key := range obj {
		if !allowed[key] {
			return fail("project_context_invalid", "unknown skill manifest member")
		}
	}
	var commands map[string]map[string]json.RawMessage
	if raw, ok := obj["commands"]; ok {
		if err := json.Unmarshal(raw, &commands); err != nil {
			return fail("project_context_invalid", "invalid skill commands")
		}
	}
	for _, command := range commands {
		var kind string
		_ = json.Unmarshal(command["type"], &kind)
		fields := map[string]bool{"type": true}
		switch kind {
		case "script":
			fields["unix_path"], fields["win_path"] = true, true
		case "system":
			fields["command"], fields["hint"] = true, true
		default:
			return fail("project_context_invalid", "unsupported legacy skill command")
		}
		for key := range command {
			if !fields[key] {
				return fail("project_context_invalid", "unknown skill command member")
			}
		}
	}
	return nil
}

func (op *operation) skillSemantics(c Candidate) (*SkillSemantics, error) {
	if err := op.verify(c); err != nil {
		return nil, err
	}
	snapshot := filepath.Join(op.root, "snapshots", c.Item.Payload.SnapshotSHA256, "files")
	spec, err := skillspec.Load(snapshot)
	if err != nil {
		return nil, err
	}
	capt, err := op.get(c.Item.Payload.SnapshotRef, c.Item.Payload.SnapshotSHA256)
	if err != nil {
		return nil, err
	}
	var raw struct {
		Commands map[string]map[string]any `json:"commands"`
	}
	if err := json.Unmarshal(capt.Bytes[skillspec.CanonicalManifestName], &raw); err != nil {
		return nil, err
	}
	// Defaults are made visible; requested capabilities never program policy.
	caps := map[string]any{"network": spec.Capabilities.Network, "filesystem": map[string]any{"keyword": spec.Capabilities.Filesystem.Keyword, "paths": spec.Capabilities.Filesystem.Paths}, "exec": spec.Capabilities.Exec, "secrets": spec.Capabilities.Secrets, "env_read": spec.Capabilities.EnvRead, "prompt_scope": spec.Capabilities.PromptScope}
	out := &SkillSemantics{CommandOwners: []CommandOwner{}, RequestedCapabilities: caps, MCPEndpoints: []string{}, PermissionEffects: "none; manager policy only; candidate capabilities are inert requests"}
	names := []string{}
	for name := range spec.Commands {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		out.CommandOwners = append(out.CommandOwners, CommandOwner{Name: name, ItemID: c.Item.ID, Declaration: raw.Commands[name]})
	}
	return out, nil
}
