package projectcontext

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/envfragment"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/protocoljson"
)

// ClaudeProjectVersion is the exact native evidence release for this adapter.
const ClaudeProjectVersion = "2.1.305"
const claudeOwnershipPath = ".agent-context/claude-managed.json"

// ClaudeControls is manager-programmed policy, never native repository settings.
// Empty grants mean default deny. Native permission patterns and settings keys
// have no input arm. A manager may grant or deny an entire named native tool.
type ClaudeControls struct {
	Version    int
	AllowTools []string
	DenyTools  []string
}

// ClaudeDiagnostic is a typed project refusal with no filesystem locators.
type ClaudeDiagnostic struct {
	Code   string
	Detail string
}

func (d *ClaudeDiagnostic) Error() string    { return d.Code + ": " + d.Detail }
func claudeRefuse(code, detail string) error { return &ClaudeDiagnostic{code, detail} }

// ClaudeSurface describes the bounded native evidence, not a capability
// advertisement. Uncertified reloads keep the complete strict tuple closed.
type ClaudeSurface string

const (
	ClaudeStartup      ClaudeSurface = "startup-sources"
	ClaudeNestedRead   ClaudeSurface = "nested-read-instructions"
	ClaudeSkills       ClaudeSurface = "skills-and-explicit-reload"
	ClaudeMCP          ClaudeSurface = "explicit-stdio-mcp"
	ClaudePermissions  ClaudeSurface = "startup-settings-permissions"
	ClaudeSafeResume   ClaudeSurface = "excluded-origin-resume"
	ClaudeChild        ClaudeSurface = "background-task-child"
	ClaudeLateHooks    ClaudeSurface = "late-hooks"
	ClaudeLatePlugins  ClaudeSurface = "late-plugins"
	ClaudeLateMCP      ClaudeSurface = "late-mcp"
	ClaudeLateSettings ClaudeSurface = "late-settings"
	ClaudeImports      ClaudeSurface = "imports-and-additional-directories"
	ClaudeMemory       ClaudeSurface = "automatic-memory-off"
	ClaudeToolPolicy   ClaudeSurface = "manager-default-deny-tools"
)

type ClaudeCertification struct {
	ProviderVersion string
	Platform        string
	Architecture    string
	Certified       []ClaudeSurface
	Uncertified     []ClaudeSurface
}

func ClaudeCertificationReport(version, platform, architecture string) ClaudeCertification {
	r := ClaudeCertification{version, platform, architecture, []ClaudeSurface{}, []ClaudeSurface{}}
	measured := []ClaudeSurface{ClaudeStartup, ClaudeNestedRead, ClaudeSkills, ClaudeMCP, ClaudePermissions, ClaudeSafeResume, ClaudeChild}
	r.Uncertified = []ClaudeSurface{ClaudeLateHooks, ClaudeLatePlugins, ClaudeLateMCP, ClaudeLateSettings, ClaudeImports, ClaudeMemory, ClaudeToolPolicy}
	if version == ClaudeProjectVersion && platform == "macos" && architecture == "arm64" {
		r.Certified = measured
	} else {
		r.Uncertified = append(measured, r.Uncertified...)
	}
	return r
}
func (r ClaudeCertification) RequireStrict() error {
	// No complete tuple has native qualification in this release.
	return claudeRefuse("project_context_unsupported", "Claude complete source-exclusion tuple is not certified")
}

var claudeToolName = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,127}$`)

func (c ClaudeControls) validate() error {
	if c.Version != 1 {
		return claudeRefuse("project_context_invalid", "invalid Claude manager policy")
	}
	for _, names := range [][]string{c.AllowTools, c.DenyTools} {
		if len(names) > 256 || !sort.StringsAreSorted(names) {
			return claudeRefuse("project_context_invalid", "tool policy must be a bounded sorted set")
		}
		for i, n := range names {
			if !claudeToolName.MatchString(n) || n == "default" || strings.HasPrefix(n, "mcp__") || (i > 0 && names[i-1] == n) {
				return claudeRefuse("project_context_invalid", "invalid tool policy name")
			}
		}
	}
	return nil
}

// claudeOutputs runs before composition hashing. The complete receipt therefore
// binds the explicit empty MCP set, manager settings and ownership marker too.
func claudeOutputs(files map[string][]byte, controls ClaudeControls) error {
	if err := controls.validate(); err != nil {
		return err
	}
	allow := append([]string{}, controls.AllowTools...)
	deny := append(append([]string{}, controls.DenyTools...), "mcp__*")
	sort.Strings(deny)
	settings, err := canonical(map[string]any{"permissions": map[string]any{"defaultMode": "dontAsk", "allow": allow, "deny": deny}})
	if err != nil {
		return err
	}
	files["settings.json"] = append(settings, '\n')
	if _, ok := files[contextmaterialize.MCPClaudePath]; !ok {
		files[contextmaterialize.MCPClaudePath] = []byte("{\"mcpServers\":{}}\n")
	}
	// HTTP has no native loading qualification in the cited adaptation sheet.
	var mcp struct {
		Servers map[string]struct {
			Type    string   `json:"type"`
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}
	if err = protocoljson.UnmarshalCanonical(bytes.TrimSuffix(files[contextmaterialize.MCPClaudePath], []byte{'\n'}), &mcp); err != nil {
		return claudeRefuse("project_context_unsupported", "only qualified Claude stdio MCP encoding is supported")
	}
	for _, s := range mcp.Servers {
		if s.Type != "stdio" {
			return claudeRefuse("project_context_unsupported", "Claude MCP transport is not qualified")
		}
	}
	marker, err := canonical(map[string]any{"version": 1, "owner": "curator", "environment": "claude_code", "mcp_path": contextmaterialize.MCPClaudePath, "mcp_sha256": digest(files[contextmaterialize.MCPClaudePath]), "settings_sha256": digest(files["settings.json"])})
	if err != nil {
		return err
	}
	files[claudeOwnershipPath] = marker
	return nil
}

// ClaudePublisher is exactly I-LEASE's publication method. The lease owner
// checks current authority/receipt and publishes transactionally. This adapter
// never writes around that owner or mutates a leased home itself.
type ClaudePublisher interface {
	PublishPrepared(context.Context, string, map[string][]byte, bool) error
}

// PublishClaude obtains fresh protected bytes and delegates their publication.
// The publisher rechecks them against its reserved lease after this read lock
// is released; revocation or a changed composition must refuse there as well.
func (a Authority) PublishClaude(ctx context.Context, ref, sha, owner, leaseID string, publisher ClaudePublisher) error {
	if publisher == nil || !idRE.MatchString(leaseID) {
		return claudeRefuse("project_lease_mismatch", "prepared lease publisher required")
	}
	op, err := a.Store.begin(false)
	if err != nil {
		return err
	}
	plan, err := a.claudePlan(ctx, op, ref, sha, owner)
	op.close()
	if err != nil {
		return err
	}
	return publisher.PublishPrepared(ctx, leaseID, plan.Preview.Files, false)
}
func (a Authority) claudePlan(ctx context.Context, op *operation, ref, sha, owner string) (compositionPlan, error) {
	var empty compositionPlan
	if owner != op.state.Principal {
		return empty, claudeRefuse("project_item_unapproved", "current manager principal required")
	}
	_, err := a.resolve(ctx, op, ref, sha)
	if err != nil {
		return empty, err
	}
	root, err := os.OpenRoot(filepath.Join(op.root, "compositions"))
	if err != nil {
		return empty, claudeRefuse("project_context_unreadable", "composition unavailable")
	}
	defer root.Close()
	b, _, err := read(root, sha+".json", MaxRecord)
	if err != nil {
		return empty, err
	}
	var saved compositionEnvelope
	if err = decode(b, &saved); err != nil {
		return empty, err
	}
	req := saved.Request
	if req.Claude == nil || req.Adapter.Environment != "claude_code" {
		return empty, claudeRefuse("project_context_unsupported", "composition has no Claude manager controls")
	}
	if req.Adapter.ProviderVersion != ClaudeProjectVersion || req.Adapter.Platform != "macos" {
		return empty, claudeRefuse("project_context_unsupported", "Claude adapter release or platform is unqualified")
	}
	return op.compose(ctx, req)
}

type ClaudeEntry string

const (
	ClaudeFresh  ClaudeEntry = "fresh"
	ClaudeResume ClaudeEntry = "resume"
	ClaudeFork   ClaudeEntry = "fork"
)

// ClaudeLaunchContribution is internal data for the delegated construction
// site. Fragment contains only existing home/channel fields; it MUST NOT be
// emitted as a standalone project-capable v2 outer carrier. The destination
// must also negotiate and bind project composition/lease authority. This type
// does not certify strict launch, perform login or accept caller argv/env.
type ClaudeLaunchContribution struct {
	ContextRef    string
	ContextSHA256 string
	HomeRef       string
	// Args is the complete channel/control argv; do not expand Fragment channels again.
	Fragment       envfragment.Fragment
	Args           []string
	MemoryDisabled bool
	Certification  ClaudeCertification
}

// Environment applies the fixed native memory control alongside the existing
// home fragment. This is an internal adapter operation, not a new wire env arm.
func (c ClaudeLaunchContribution) Environment() map[string]string {
	out := map[string]string{}
	for k, v := range c.Fragment.Env {
		out[k] = v
	}
	if c.MemoryDisabled {
		out["CLAUDE_CODE_DISABLE_AUTO_MEMORY"] = "1"
	}
	return out
}

// ClaudeLaunch borrows the same-home lock, recomputes admission, and verifies
// every materialized receipt entry before offering the internal contribution.
// All resume/fork entries refuse until protected origin provenance is wired.
func (a Authority) ClaudeLaunch(ctx context.Context, held managerlock.Held, ref, sha, owner string, entry ClaudeEntry) (ClaudeLaunchContribution, error) {
	var out ClaudeLaunchContribution
	if entry != ClaudeFresh {
		return out, claudeRefuse("project_context_unsupported", "Claude resume/fork requires protected excluded-origin provenance")
	}
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		return out, claudeRefuse("project_context_unsupported", "Claude native tuple is unqualified")
	}
	op, err := a.Store.readHeld(held)
	if err != nil {
		return out, err
	}
	plan, err := a.claudePlan(ctx, op, ref, sha, owner)
	if err != nil {
		return out, err
	}
	p := plan.Preview
	home := filepath.Join(filepath.Dir(op.root), "environments", "project-views", plan.Preview.Generation.RegistrationRef[len("registration:"):], p.Record.Profile.Name, "claude_code")
	if err = verifyClaudeReceipt(home, p.Receipt); err != nil {
		return out, err
	}
	out.ContextRef, out.ContextSHA256, out.HomeRef = p.Ref, p.SHA256, p.Record.HomeRef
	out.Fragment = envfragment.Fragment{Environment: "claude_code", Profile: p.Record.Profile.Name, LockSHA256: p.Record.Profile.LockSHA256, Winner: "higher-weight", Placement: "winner-last", Permissions: envfragment.Permissions{Mode: "native", Source: "default"}, Env: map[string]string{"CLAUDE_CONFIG_DIR": home}}
	adapter, _ := envregistry.ByID("claude_code")
	out.Fragment.MCP = &envfragment.MCP{Path: filepath.Join(home, contextmaterialize.MCPClaudePath), EnvNames: append([]string{}, p.MCPEnvNames...), Channels: []envregistry.Channel{*adapter.MCP}}
	out.Args = []string{"--setting-sources", "user", "--mcp-config", out.Fragment.MCP.Path, "--strict-mcp-config", "--permission-mode", "dontAsk", "--tools", strings.Join(plan.Claude.AllowTools, ",")}
	if _, ok := p.Files[".agent-context/system-prompt.md"]; ok {
		out.Fragment.SystemPrompt = &envfragment.SystemPrompt{Path: filepath.Join(home, ".agent-context/system-prompt.md"), Channels: []envregistry.Channel{adapter.SystemPrompt[0]}}
		out.Args = append(out.Args, "--append-system-prompt-file", out.Fragment.SystemPrompt.Path)
	}
	out.MemoryDisabled = true
	out.Certification = ClaudeCertificationReport(ClaudeProjectVersion, "macos", "arm64")
	if err = held.Validate(filepath.Dir(op.root)); err != nil {
		return ClaudeLaunchContribution{}, err
	}
	if err = ctx.Err(); err != nil {
		return ClaudeLaunchContribution{}, err
	}
	return out, nil
}
func verifyClaudeReceipt(home string, receipt MaterializationReceipt) error {
	if err := pathboundary.ValidateRoot(home); err != nil {
		return claudeRefuse("project_context_drift", "Claude home boundary changed")
	}
	if err := privatedir.Validate(home); err != nil {
		return claudeRefuse("project_context_drift", "Claude home protection changed")
	}
	root, err := os.OpenRoot(home)
	if err != nil {
		return claudeRefuse("project_context_unreadable", "Claude home unavailable")
	}
	defer root.Close()
	for _, f := range receipt.Files {
		b, _, e := read(root, f.Path, MaxFileBytes)
		if e != nil {
			return claudeRefuse("project_context_drift", "Claude generated surface unavailable")
		}
		st, e := root.Lstat(f.Path)
		mode, parseErr := strconv.ParseUint(f.Mode, 8, 32)
		if e != nil || parseErr != nil || uint32(st.Mode().Perm()) != uint32(mode) || privatedir.ValidateFile(filepath.Join(home, filepath.FromSlash(f.Path))) != nil || int64(len(b)) != f.Size || digest(b) != f.SHA256 {
			return claudeRefuse("project_context_drift", "Claude generated surface differs from protected receipt")
		}
	}
	return nil
}
