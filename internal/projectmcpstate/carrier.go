package projectmcpstate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/pkgversion"
	"github.com/relux-works/curator/internal/projectmcp"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/sourcelock"
)

// Resolved binds a package declaration to exact acquired content and server.
type Resolved struct {
	Source        string         `json:"source"`
	Directory     *string        `json:"directory"`
	Commit        string         `json:"commit"`
	Version       string         `json:"version"`
	ContentSHA256 string         `json:"content_sha256"`
	Server        map[string]any `json:"server"`
	ServerSHA256  string         `json:"server_sha256"`
}

// Entry is the stage-A package lock record, without any inline admission arm.
type Entry struct {
	Name              string         `json:"name"`
	Declaration       map[string]any `json:"declaration"`
	DeclarationSHA256 string         `json:"declaration_sha256"`
	Resolved          Resolved       `json:"resolved"`
}

// Lock is the new outer carrier; SkillsLock remains the unchanged source model.
type Lock struct {
	SchemaVersion  int            `json:"schema_version"`
	Capability     string         `json:"capability"`
	ProjectID      string         `json:"project_id"`
	ManifestSHA256 string         `json:"manifest_sha256"`
	SkillsLock     map[string]any `json:"skills_lock"`
	Entries        []Entry        `json:"entries"`
	LockSHA256     string         `json:"lock_sha256"`
}

// Row is one full server/environment observation. A nil Path means unavailable.
type Row struct {
	Name        string  `json:"name"`
	Environment string  `json:"environment"`
	Path        *string `json:"path"`
	Outcome     string  `json:"outcome"`
	Code        *string `json:"code"`
}

// Option records only an ignored option's name, never its value.
type Option struct {
	Name     string `json:"name"`
	Option   string `json:"option"`
	Handling string `json:"handling"`
}

// Plan reports a declaration set without granting ownership or approval.
type Plan struct {
	SchemaVersion  int      `json:"schema_version"`
	Capability     string   `json:"capability"`
	ProjectID      string   `json:"project_id"`
	ManifestSHA256 string   `json:"manifest_sha256"`
	LockSHA256     *string  `json:"lock_sha256"`
	DryRun         bool     `json:"dry_run"`
	Rows           []Row    `json:"rows"`
	InstallOptions []Option `json:"install_options"`
	PlanSHA256     string   `json:"plan_sha256"`
}

// Canonical encodes typed carriers through JSON into the CCJ-1 domain. It does
// not project out sig or other arbitrary option names.
func Canonical(value any) ([]byte, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var obj any
	if err := decoder.Decode(&obj); err != nil {
		return nil, err
	}
	return protocoljson.MarshalCanonical(obj)
}
func carrierHash(value any, field string) (string, error) {
	payload, err := Canonical(value)
	if err != nil {
		return "", err
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var obj map[string]any
	if err := decoder.Decode(&obj); err != nil {
		return "", err
	}
	delete(obj, field)
	return projectmcp.Digest(obj)
}

// Seal computes the complete lock identity after validating all package binds.
func (lock *Lock) Seal() error {
	hash, err := carrierHash(lock, "lock_sha256")
	if err != nil {
		return err
	}
	lock.LockSHA256 = hash
	return lock.Validate()
}

// Validate checks both independent digests, exact declaration membership and
// the enclosed source-lock manifest binding before any replay may use records.
func (lock *Lock) Validate() error {
	if lock.SchemaVersion != 1 || lock.Capability != projectmcp.Capability || !validProject(lock.ProjectID) || !digestRE.MatchString(lock.ManifestSHA256) || lock.Entries == nil {
		return invalid("MCP lock header")
	}
	sourceBytes, err := protocoljson.MarshalCanonical(lock.SkillsLock)
	if err != nil {
		return invalid("enclosed skills lock")
	}
	skills, err := sourcelock.Parse(sourceBytes)
	if err != nil {
		return err
	}
	if skills.ManifestSHA256 != lock.ManifestSHA256 {
		return fmt.Errorf("source_lock_stale: enclosed manifest identity differs")
	}
	previous := ""
	for _, entry := range lock.Entries {
		if !projectmcp.ValidName(entry.Name) || entry.Name <= previous {
			return invalid("MCP lock NAME order")
		}
		previous = entry.Name
		decl, err := projectmcp.ParseDeclaration(entry.Declaration)
		if err != nil {
			return err
		}
		if err := decl.RequirePackage(); err != nil {
			return err
		}
		modelBytes, err := Canonical(decl.Model())
		if err != nil {
			return err
		}
		declaredBytes, err := Canonical(entry.Declaration)
		if err != nil || !bytes.Equal(modelBytes, declaredBytes) {
			return invalid("MCP declaration normalization")
		}
		if _, err := pkgversion.ParseVersion(entry.Resolved.Version); err != nil {
			return invalid("MCP resolved version")
		}
		if revision, ok := decl.Package["revision"].(string); ok && revision != entry.Resolved.Commit {
			return invalid("MCP resolved revision")
		}
		declarationHash, err := projectmcp.Digest(decl.Model())
		if err != nil {
			return err
		}
		if declarationHash != entry.DeclarationSHA256 {
			return invalid("MCP declaration digest")
		}
		normalized, err := projectmcp.NormalizeServer(entry.Resolved.Server)
		if err != nil {
			return err
		}
		hash, err := projectmcp.Digest(normalized)
		if err != nil {
			return err
		}
		if hash != entry.Resolved.ServerSHA256 {
			return invalid("MCP server digest")
		}
		normalizedBytes, _ := Canonical(normalized)
		serverBytes, _ := Canonical(entry.Resolved.Server)
		if !bytes.Equal(normalizedBytes, serverBytes) {
			return invalid("MCP server normalization")
		}
		if !contextpkg.ValidCommit(entry.Resolved.Commit) || !identity.ValidCanonical(entry.Resolved.Source) || !digestRE.MatchString(entry.Resolved.ContentSHA256) {
			return invalid("MCP resolved identity")
		}
		source, _ := identity.Parse(decl.Package["git"].(string))
		if source != entry.Resolved.Source {
			return invalid("MCP resolved source")
		}
		directory, _ := decl.Package["directory"].(string)
		if (directory == "") != (entry.Resolved.Directory == nil) {
			return invalid("MCP resolved directory")
		}
		if entry.Resolved.Directory != nil && directory != *entry.Resolved.Directory {
			return invalid("MCP resolved directory")
		}
	}
	hash, err := carrierHash(lock, "lock_sha256")
	if err != nil {
		return err
	}
	if lock.LockSHA256 != hash {
		return invalid("MCP lock digest")
	}
	return nil
}

var digestRE = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

// DecodeLock rejects extra carrier fields and noncanonical encodings before
// interpreting a record. Canonical shape alone never supplies local authority.
func DecodeLock(payload []byte) (*Lock, error) {
	if err := protocoljson.RequireCanonical(payload); err != nil {
		return nil, invalid("MCP lock JSON")
	}
	var shape map[string]any
	rawDecoder := json.NewDecoder(bytes.NewReader(payload))
	rawDecoder.UseNumber()
	if err := rawDecoder.Decode(&shape); err != nil {
		return nil, invalid("MCP lock JSON")
	}
	if !required(shape, "schema_version", "capability", "project_id", "manifest_sha256", "skills_lock", "entries", "lock_sha256") {
		return nil, invalid("MCP lock fields")
	}
	entries, ok := shape["entries"].([]any)
	if !ok {
		return nil, invalid("MCP lock entries")
	}
	for _, raw := range entries {
		entry, ok := raw.(map[string]any)
		if !ok || !required(entry, "name", "declaration", "declaration_sha256", "resolved") {
			return nil, invalid("MCP lock entry")
		}
		resolved, ok := entry["resolved"].(map[string]any)
		if !ok || !required(resolved, "source", "directory", "commit", "version", "content_sha256", "server", "server_sha256") {
			return nil, invalid("MCP lock resolved fields")
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
	var lock Lock
	if err := decoder.Decode(&lock); err != nil {
		return nil, invalid("MCP lock fields")
	}
	if err := lock.Validate(); err != nil {
		return nil, err
	}
	return &lock, nil
}

// NewPlan reports no writers in this parser/resolver-only stage. Explicit known
// targets remain visible as unsupported; absent selectors mean writable targets
// (currently none), independently of Skillfile skill-agent selection.
func NewPlan(project, manifest string, decls map[string]projectmcp.Declaration, entries []Entry, lockHash *string, dry bool, refusals map[string]string) (*Plan, error) {
	plan := &Plan{SchemaVersion: 1, Capability: projectmcp.Capability, ProjectID: project, ManifestSHA256: manifest, LockSHA256: lockHash, DryRun: dry, Rows: []Row{}, InstallOptions: []Option{}}
	if !validProject(project) || !digestRE.MatchString(manifest) || (lockHash != nil && !digestRE.MatchString(*lockHash)) {
		return nil, invalid("MCP plan identity")
	}
	for name, code := range refusals {
		_, present := decls[name]
		if !present || !refusalCodes[code] {
			return nil, invalid("MCP plan refusal code")
		}
	}
	resolved := map[string]map[string]any{}
	for _, entry := range entries {
		resolved[entry.Name] = entry.Resolved.Server
	}
	refused := len(refusals) > 0
	for _, name := range projectmcp.Names(decls) {
		d := decls[name]
		if d.Server != nil {
			refused = true
		}
		server := d.Server
		if server == nil {
			server = resolved[name]
		}
		for _, option := range d.OptionNames() {
			plan.InstallOptions = append(plan.InstallOptions, Option{name, option, "ignored"})
		}
		var environments []string
		switch list := server["environments"].(type) {
		case []string:
			environments = append([]string{}, list...)
		case []any:
			for _, raw := range list {
				if text, ok := raw.(string); ok {
					environments = append(environments, text)
				}
			}
		}
		sort.Strings(environments)
		for _, environment := range environments {
			row := Row{Name: name, Environment: environment, Outcome: "skipped: unsupported"}
			code := refusals[name]
			if d.Server != nil {
				code = projectmcp.InlineUnsupported
			}
			if code != "" {
				row.Outcome = "refused"
				row.Code = &code
			}
			plan.Rows = append(plan.Rows, row)
		}
	}
	if lockHash == nil && len(decls) > 0 && !refused {
		return nil, invalid("successful nonempty plan requires lock identity")
	}
	hash, err := carrierHash(plan, "plan_sha256")
	if err != nil {
		return nil, err
	}
	plan.PlanSHA256 = hash
	if err := plan.Validate(); err != nil {
		return nil, err
	}
	return plan, nil
}

// Lines renders total pair observations and ignored option names, never values.
func (plan *Plan) Lines() []string {
	lines := []string{}
	prefix := ""
	if plan.DryRun {
		prefix = "would: "
	}
	for _, row := range plan.Rows {
		path := "unavailable"
		if row.Path != nil {
			path = *row.Path
		}
		outcome := row.Outcome
		if row.Code != nil {
			outcome += ": " + *row.Code
		}
		lines = append(lines, fmt.Sprintf("%smcp %s -> %s path: %s (%s)", prefix, row.Name, row.Environment, path, outcome))
	}
	for _, option := range plan.InstallOptions {
		lines = append(lines, fmt.Sprintf("%smcp %s install option %q: ignored", prefix, option.Name, option.Option))
	}
	return lines
}

// CheckManifest binds local replay to the exact raw source identity and project.
func (lock *Lock) CheckManifest(project, manifest string) error {
	if lock.ProjectID != project || lock.ManifestSHA256 != manifest {
		return fmt.Errorf("source_lock_stale: MCP project/manifest identity changed")
	}
	return nil
}

func validProject(project string) bool {
	return project != "" && utf8.ValidString(project) && utf8.RuneCountInString(project) <= 4096
}

// Refusal codes are closed for this carrier. Existing source/audit diagnostics
// keep their original spelling; reserved stage-B option/approval codes cannot
// be smuggled into a stage-A row.
var refusalCodes = map[string]bool{
	projectmcp.Invalid: true, projectmcp.InlineUnsupported: true,
	"mcp_package_not_allowed": true, "mcp_package_allowlist_empty": true,
	"mcp_contribution_policy_refused": true, "mcp_native_entry_conflict": true,
	"context-secret-material": true, "source_snapshot_unavailable": true,
	"source_snapshot_changed": true, "source_lock_stale": true,
	"context_source_signer_rejected": true, "environment_surface_unreadable": true,
	"profile_source_invalid": true,
}

// Refusals applies one transaction refusal to every independently known pair.
// Unknown I/O classes remain declaration-level diagnostics, never invented codes.
func Refusals(decls map[string]projectmcp.Declaration, err error) map[string]string {
	out := map[string]string{}
	code := strings.SplitN(err.Error(), ":", 2)[0]
	if !refusalCodes[code] {
		return out
	}
	for name := range decls {
		out[name] = code
	}
	return out
}

func invalid(field string) error { return &projectmcp.Error{Code: projectmcp.Invalid, Field: field} }

func required(obj map[string]any, names ...string) bool {
	for _, name := range names {
		if _, ok := obj[name]; !ok {
			return false
		}
	}
	return true
}

// FailurePlan preserves independently known pairs without authorizing partial
// admission. Malformed declarations themselves supply no rows or option names.
func FailurePlan(project string, payload []byte, decls map[string]projectmcp.Declaration, dry bool, cause error) *Plan {
	var partial *projectmcp.SetError
	if errors.As(cause, &partial) {
		decls = partial.Valid
	}
	if decls == nil {
		return nil
	}
	hash, err := sourcelock.ManifestDigest(payload)
	if err != nil {
		return nil
	}
	plan, err := NewPlan(project, hash, decls, nil, nil, dry, Refusals(decls, cause))
	if err != nil {
		return nil
	}
	return plan
}
