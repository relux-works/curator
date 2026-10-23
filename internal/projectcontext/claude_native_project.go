package projectcontext

import (
	"bytes"
	"context"
	"errors"
	"strings"

	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/projectadapter"
)

// ClaudeProfileDocumentSource is the I-COMP profile-only handoff. Its owner
// must resolve ProfileDocuments/ProfileDocumentReceipt from the same protected
// composition reference/digest. Never recover them by splitting composed prose.
// The stub below is used until that owner publishes its r3 implementation.
type ClaudeProfileDocumentSource interface {
	ProfileDocuments(context.Context, string, string) (map[string][]byte, []File, error)
}

// ClaudeProfileDocumentsUnavailable refuses native preparation until the profile-only producer is integrated.
type ClaudeProfileDocumentsUnavailable struct{}

// ProfileDocuments refuses without inventing a profile-only document from composed prose.
func (ClaudeProfileDocumentsUnavailable) ProfileDocuments(context.Context, string, string) (map[string][]byte, []File, error) {
	return nil, nil, claudeRefuse("project_context_unsupported", "dedicated I-COMP profile documents are not available")
}

// ClaudePreparation is an offline candidate, not a protected Composition. Its
// receipt binds this mode's new bytes but cannot authorize a lease publication.
// There is intentionally no writer, publisher, executable argv or environment.
type ClaudePreparation struct {
	SourceContextRef    string
	SourceContextSHA256 string
	Files               map[string][]byte
	Receipt             MaterializationReceipt
	Certification       projectadapter.Certification
}

func cloneClaudeFiles(files map[string][]byte) map[string][]byte {
	out := map[string][]byte{}
	for p, b := range files {
		out[p] = append([]byte{}, b...)
	}
	return out
}

// PrepareClaudeMode requires an explicit mode and dedicated profile-only input.
// Native discovery is intentionally not isolated in native-project. Only
// approved skills and MCP survive from project outputs; both instruction
// channels are replaced by exact profile documents. Strict remains refused.
func PrepareClaudeMode(ctx context.Context, mode string, preview CompositionPreview, controls ClaudeControls, source ClaudeProfileDocumentSource) (ClaudePreparation, error) {
	var out ClaudePreparation
	if err := ctx.Err(); err != nil {
		return out, err
	}
	if mode == projectadapter.Strict {
		return out, ClaudeCertificationReport(ClaudeProjectVersion, "macos", "arm64").RequireStrict()
	}
	if err := requireClaudeNativeMode(mode); err != nil {
		return out, err
	}
	if source == nil {
		source = ClaudeProfileDocumentsUnavailable{}
	}
	if err := controls.validate(); err != nil {
		return out, err
	}
	if err := validateClaudeFileSet(preview.Files, preview.Receipt.Files); err != nil {
		return out, claudeRefuse("project_context_invalid", "base composition receipt differs")
	}
	if preview.Record.Environment != "claude_code" || identity(preview.Receipt) != preview.Record.MaterializationSHA256 || identity(preview.Record) != preview.SHA256 || preview.Ref != "composition:"+hashID(preview.SHA256) {
		return out, claudeRefuse("project_context_invalid", "base composition binding differs")
	}
	documents, receipt, err := source.ProfileDocuments(ctx, preview.Ref, preview.SHA256)
	if err != nil {
		var refusal *ClaudeDiagnostic
		if errors.As(err, &refusal) {
			return out, err
		}
		return out, claudeRefuse("project_context_unreadable", "profile document source unavailable")
	}
	if err = validateClaudeFileSet(documents, receipt); err != nil {
		return out, claudeRefuse("project_context_invalid", "profile document receipt differs")
	}
	for _, f := range receipt {
		if (f.Path != "CLAUDE.md" && f.Path != ".agent-context/system-prompt.md") || f.Mode != "0600" {
			return out, claudeRefuse("project_context_invalid", "profile document surface is not allowed")
		}
	}
	files := cloneClaudeFiles(documents)
	executable := map[string]bool{}
	for _, f := range preview.Receipt.Files {
		if strings.HasPrefix(f.Path, "skills/") || f.Path == contextmaterialize.MCPClaudePath {
			files[f.Path] = append([]byte{}, preview.Files[f.Path]...)
			executable[f.Path] = f.Mode == "0700"
		}
	}
	if err = claudeOutputs(files, controls); err != nil {
		return out, err
	}
	if !bytes.Equal(files["settings.json"], preview.Files["settings.json"]) || !bytes.Equal(files[claudeOwnershipPath], preview.Files[claudeOwnershipPath]) {
		return out, claudeRefuse("project_context_invalid", "manager policy differs from composition")
	}
	policy, err := canonical(struct {
		BaseComposition, BasePolicy, Mode, Form string
		ProfileDocuments                        []File
	}{preview.SHA256, preview.Record.PolicySHA256, mode, "managed-home", receipt})
	if err != nil {
		return out, err
	}
	candidate, err := compositionReceipt(files, executable, preview.Receipt.AdapterSHA256, digest(policy), preview.Receipt.WorktreeRef)
	if err != nil {
		return out, err
	}
	if err = ctx.Err(); err != nil {
		return out, err
	}
	return ClaudePreparation{preview.Ref, preview.SHA256, files, candidate, ClaudeModeCertificationReport(ClaudeProjectVersion, "macos", "arm64", mode).Certification}, nil
}

// ClaudeModeCertificationReport is status data for an explicit operator mode.
// Strict measurements do not certify native-project's different permission
// configuration. P6 covers Read atoms; no memory-on/off probe exists.
func ClaudeModeCertificationReport(version, platform, architecture, mode string) ClaudeCertification {
	if mode == projectadapter.Strict {
		return ClaudeCertificationReport(version, platform, architecture)
	}
	r := ClaudeCertification{projectadapter.Certification{Environment: "claude_code", ProviderVersion: version, Platform: platform + "/" + architecture, Mode: mode, Form: "managed-home", Certified: []projectadapter.SurfaceName{}, Refused: []projectadapter.Refusal{}}}
	if mode != projectadapter.NativeProject {
		r.Refused = append(r.Refused, projectadapter.Refusal{Surface: "mode", Reason: "explicit supported operator mode required"})
		return r
	}
	r.Notice = projectadapter.NativeProjectNotice
	r.Refused = append(r.Refused,
		projectadapter.Refusal{Surface: projectadapter.Permissions, Reason: "P6 does not qualify native-project manager default deny"},
		projectadapter.Refusal{Surface: projectadapter.MemoryOff, Reason: "no discriminating native automatic-memory-off probe"})
	if version != ClaudeProjectVersion || platform != "macos" || architecture != "arm64" {
		r.Refused = append(r.Refused, projectadapter.Refusal{Surface: "native-tuple", Reason: "exact native tuple not qualified"})
	}
	return r
}

// ClaudeLaunchMode refuses before authority lookup while mandatory native
// permissions/memory qualifications are missing. It cannot start a provider or
// publish a home. The legacy strict construction seam remains internal only.
func (a Authority) ClaudeLaunchMode(_ context.Context, _ managerlock.Held, _, _, _ string, _ ClaudeEntry, mode string) (projectadapter.LaunchContribution, error) {
	var out projectadapter.LaunchContribution
	if mode == projectadapter.Strict {
		return out, ClaudeCertificationReport(ClaudeProjectVersion, "macos", "arm64").RequireStrict()
	}
	if err := requireClaudeNativeMode(mode); err != nil {
		return out, err
	}
	r := ClaudeModeCertificationReport(ClaudeProjectVersion, "macos", "arm64", mode)
	return out, &ClaudeDiagnostic{Code: "project_context_unsupported", Detail: "Claude native-project requires qualified manager permissions and memory-off", Mode: mode, Notice: r.Notice, Refused: r.Refused}
}

// Inventory grammar alone does not check the byte map. Require exact set,
// sizes and digests as well, before projecting any candidate native outputs.
func validateClaudeFileSet(files map[string][]byte, receipt []File) error {
	if err := validateInventory(captured{Manifest: Snapshot{1, receipt}}); err != nil {
		return err
	}
	if len(files) != len(receipt) {
		return claudeRefuse("project_context_invalid", "file receipt set differs")
	}
	for _, f := range receipt {
		b, ok := files[f.Path]
		if !ok || int64(len(b)) != f.Size || digest(b) != f.SHA256 {
			return claudeRefuse("project_context_invalid", "file receipt bytes differ")
		}
	}
	return nil
}

func requireClaudeNativeMode(mode string) error {
	if mode != projectadapter.NativeProject {
		return claudeRefuse("project_context_invalid", "explicit Claude mode required")
	}
	return nil
}
