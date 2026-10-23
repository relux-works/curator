package projectadapter

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/identifiers"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

// CodexRequest comes only from manager-held admitted composition/seed inputs.
// There is deliberately no repository config, permission file or native-home
// reader here. Generated must be the immutable I-COMP receipt for Files.
type CodexRequest struct {
	ProfileDocuments     map[string][]byte
	ProfileGenerated     map[string]Surface
	NativeProjectEnabled bool
	Files                map[string][]byte
	Generated            map[string]Surface
	SeedConfig           []byte
	SeedRevision         string
	ProviderVersion      string
	Platform             string
	Mode                 string
	Model                string
	Effort               string
	Approval             string
	Sandbox              string
	Home                 string
	Worktree             string
}

// Plan owns its rendered bytes and receipt; accessors return copies so a caller
// cannot bless drift by mutating the materialization baseline.
type Plan struct {
	files     map[string][]byte
	generated map[string]Surface
	launch    LaunchContribution
	report    Certification
	notes     []Note
	home      string
	worktree  string
}

func (p *Plan) Files() map[string][]byte {
	out := map[string][]byte{}
	for k, v := range p.files {
		out[k] = bytes.Clone(v)
	}
	return out
}
func (p *Plan) Receipt() map[string]Surface {
	out := map[string]Surface{}
	for k, v := range p.generated {
		out[k] = v
	}
	return out
}
func (p *Plan) Launch() LaunchContribution {
	env := map[string]string{}
	for k, v := range p.launch.Env {
		env[k] = v
	}
	return LaunchContribution{Mode: p.launch.Mode, Notice: p.launch.Notice, Env: env, Args: append([]string{}, p.launch.Args...)}
}
func (p *Plan) Report() Certification {
	r := p.report
	r.Certified = append([]SurfaceName{}, r.Certified...)
	r.Refused = append([]Refusal{}, r.Refused...)
	return r
}
func (p *Plan) Notes() []Note { return append([]Note{}, p.notes...) }

// RequireStrict always refuses this pinned adapter. Native-project delivery does
// not weaken I-COMP's native-source-exclusion gate or manufacture certification.
func (p *Plan) RequireStrict() error {
	return &Unsupported{NativeSourceExclusion, codexExclusionReason}
}

// RequireLaunch refuses native-project permission policy at this pin because
// repository MCP permission keys survive manager argv. Rendered plans remain
// inspection data; neither this gate nor strict mode may be bypassed by a
// delegated runner.
func (p *Plan) RequireLaunch() error {
	if p.report.Mode != NativeProject {
		return p.RequireStrict()
	}
	return &Unsupported{Permissions, codexNativeProjectReason}
}

const codexNativeProjectReason = "Codex 0.162.1 repository mcp_servers.*.disabled_tools survives manager argv and changes exposed MCP tools; launch-time dotted overrides do not cover late servers at native thread entry, so native-project permissions are unsupported"

const codexExclusionReason = "Codex 0.162.1 retains repository skills, including late-created paths; lazy, resume and child exclusion remain unqualified"

// RenderCodex preserves profile-only root bytes and admitted skill bytes, then programs only
// manager-owned config keys. It never discovers or approves source files.
func RenderCodex(req CodexRequest) (*Plan, error) {
	if req.Mode != NativeProject {
		return nil, &Unsupported{NativeSourceExclusion, "only native-project delivery is admitted; " + codexExclusionReason}
	}
	if !req.NativeProjectEnabled {
		return nil, &Unsupported{NativeSourceExclusion, "native-project is not enabled by the operator"}
	}
	if req.ProviderVersion != "0.162.1" || req.Platform != "macos" {
		return nil, &Unsupported{NativeSourceExclusion, "release/platform lacks the pinned Codex qualification"}
	}
	if !filepath.IsAbs(req.Home) || filepath.Clean(req.Home) != req.Home || !filepath.IsAbs(req.Worktree) || filepath.Clean(req.Worktree) != req.Worktree || req.Home == req.Worktree {
		return nil, fmt.Errorf("project_context_invalid: manager home/worktree allocations required")
	}
	if req.SeedRevision != envregistry.CodexSeedRevisionA && req.SeedRevision != envregistry.CodexSeedRevisionB {
		return nil, fmt.Errorf("project_context_invalid: unknown Codex seed revision")
	}
	if (req.Approval != "never" && req.Approval != "on-request") || (req.Sandbox != "read-only" && req.Sandbox != "workspace-write") {
		return nil, &Unsupported{Permissions, "unsupported manager approval or sandbox policy"}
	}
	if req.Model == "" || len(req.Model) > 256 || strings.ContainsAny(req.Model, "\n\r\x00") {
		return nil, fmt.Errorf("project_context_invalid: explicit manager model required")
	}
	if !map[string]bool{"minimal": true, "low": true, "medium": true, "high": true, "xhigh": true}[req.Effort] {
		return nil, fmt.Errorf("project_context_invalid: unsupported reasoning effort")
	}
	if err := validateInput(req.Files, req.Generated); err != nil {
		return nil, err
	}
	if err := validateInput(req.ProfileDocuments, req.ProfileGenerated); err != nil {
		return nil, fmt.Errorf("profile documents: %w", err)
	}
	adapter, _ := envregistry.ByID(envregistry.CodexCLI)
	bundle, err := codexSystemSkills(codexSystemBundle)
	if err != nil {
		return nil, err
	}
	p := &Plan{files: bundle, generated: map[string]Surface{}, home: req.Home, worktree: req.Worktree}

	for name, b := range req.ProfileDocuments {
		if name != adapter.RootTarget && name != ".agent-context/system-prompt.md" {
			return nil, &Unsupported{Context, "profile document output is not a registered instruction surface"}
		}
		p.files[name] = bytes.Clone(b)
	}
	var mcpBytes []byte
	for name, b := range req.Files {
		folded := cases.Fold().String(norm.NFC.String(name))
		if folded == "skills/.system" || strings.HasPrefix(folded, "skills/.system/") {
			return nil, &Unsupported{Skills, "reserved native skill namespace"}
		}
		switch {
		case name == adapter.RootTarget || name == ".agent-context/system-prompt.md":
			// Native-project never installs concatenated project instructions.
		case strings.HasPrefix(name, adapter.SkillsDir+"/"):
			p.files[name] = bytes.Clone(b)
		case name == contextmaterialize.MCPCodexPath:
			mcpBytes = b
		default:
			return nil, &Unsupported{Context, "composition contains an unsupported native output"}
		}
	}

	config := map[string]any{}
	if _, err := toml.Decode(string(req.SeedConfig), &config); err != nil {
		return nil, fmt.Errorf("project_context_invalid: seed config is not valid TOML")
	}
	// These seed inputs select alternate control sources. They are outside this
	// plan's ownership, so refuse explicitly instead of silently pruning them.
	for _, key := range []string{"instructions", "base_instructions", "developer_instructions", "permissions", "default_permissions", "notify", "experimental_instructions_file", "hooks", "plugins", "agents", "rules", "policy"} {
		if _, ok := config[key]; ok {
			return nil, fmt.Errorf("project_context_unsupported: seed control key %s requires separate admission", key)
		}
	}
	replace := func(key string, value any) {
		if _, ok := config[key]; ok {
			p.notes = append(p.notes, Note{key, "replaced-by-composition"})
		}
		config[key] = value
	}
	drop := func(key string) {
		if _, ok := config[key]; ok {
			p.notes = append(p.notes, Note{key, "dropped-by-composition"})
			delete(config, key)
		}
	}
	mcp := map[string]any{}
	if _, err := toml.Decode(string(mcpBytes), &mcp); err != nil {
		return nil, fmt.Errorf("project_context_drift: invalid composed MCP TOML")
	}
	for key := range mcp {
		if key != "mcp_servers" {
			return nil, fmt.Errorf("project_context_invalid: non-MCP composed config key")
		}
	}
	servers := map[string]any{}
	if v, ok := mcp["mcp_servers"]; ok {
		var valid bool
		servers, valid = v.(map[string]any)
		if !valid {
			return nil, fmt.Errorf("project_context_invalid: MCP table required")
		}
	}
	for name, raw := range servers {
		table, ok := raw.(map[string]any)
		if !identifiers.Valid(name) || !ok {
			return nil, fmt.Errorf("project_context_invalid: composed MCP server")
		}
		if _, http := table["url"]; http {
			return nil, &Unsupported{MCP, "HTTP native entry loading is unqualified at this release/platform"}
		}
		command, ok := table["command"].(string)
		if !ok || command == "" {
			return nil, fmt.Errorf("project_context_invalid: composed stdio command required")
		}

		for k := range table {
			if k != "command" && k != "args" && k != "url" {
				return nil, fmt.Errorf("project_context_invalid: unowned MCP server setting")
			}
		}
	}
	if raw, ok := config["mcp_servers"]; ok {
		seedServers, valid := raw.(map[string]any)
		if !valid {
			return nil, fmt.Errorf("project_context_invalid: seed MCP table required")
		}
		for name := range seedServers {
			p.notes = append(p.notes, Note{"mcp_servers." + name, "dropped-by-composition"})
		}
	}
	replace("mcp_servers", servers)
	replace("model", req.Model)
	replace("model_reasoning_effort", req.Effort)
	replace("approval_policy", req.Approval)
	replace("approvals_reviewer", "user")
	replace("sandbox_mode", req.Sandbox)
	replace("sandbox_workspace_write", map[string]any{"network_access": false, "writable_roots": []string{}, "exclude_tmpdir_env_var": true, "exclude_slash_tmp": true})
	// Native-project reads repository prose and config under native rules.
	// Trust is not Curator admission. Manager argv owns policy/memory settings,
	// but surviving repository MCP permission effects still make launch refuse.
	replace("projects", map[string]any{req.Worktree: map[string]any{"trust_level": "trusted"}})
	replace("project_doc_max_bytes", int64(32768))
	replace("project_doc_fallback_filenames", []string{})
	drop("skills")
	replace("project_root_markers", []string{".git"})
	drop("profile")
	drop("profiles")
	drop("model_instructions_file")
	setTable := func(table, key string, value any) error {
		v, exists := config[table]
		if !exists {
			v = map[string]any{}
			config[table] = v
		}
		obj, ok := v.(map[string]any)
		if !ok {
			return fmt.Errorf("project_context_invalid: seed %s table required", table)
		}
		if _, ok = obj[key]; ok {
			p.notes = append(p.notes, Note{table + "." + key, "replaced-by-composition"})
		}
		obj[key] = value
		return nil
	}
	for _, row := range []struct{ table, key string }{{"features", "memories"}, {"features", "memory_tool"}, {"memories", "generate_memories"}, {"memories", "use_memories"}, {"memories", "dedicated_tools"}} {
		if err := setTable(row.table, row.key, false); err != nil {
			return nil, err
		}
	}
	if _, ok := p.files[".agent-context/system-prompt.md"]; ok {
		config["model_instructions_file"] = filepath.Join(req.Home, ".agent-context", "system-prompt.md")
	}
	var rendered bytes.Buffer
	rendered.WriteString("# Curator-owned composition: mcp_servers, policy, model and memory keys.\n")
	if err := toml.NewEncoder(&rendered).Encode(config); err != nil {
		return nil, fmt.Errorf("project_context_invalid: cannot render composed config")
	}
	p.files["config.toml"] = rendered.Bytes()
	for name, b := range p.files {
		p.generated[name] = surface(b, 0600)
	}
	overrides := []string{"mcp_servers", "model", "model_reasoning_effort", "approval_policy", "approvals_reviewer", "sandbox_mode", "sandbox_workspace_write", "projects", "project_doc_max_bytes", "project_doc_fallback_filenames", "project_root_markers", "features.memories", "features.memory_tool", "memories.generate_memories", "memories.use_memories", "memories.dedicated_tools"}
	if _, ok := config["model_instructions_file"]; ok {
		overrides = append(overrides, "model_instructions_file")
	}
	args := []string{}
	for _, key := range overrides {
		value := configValue(config, key)
		literal, err := tomlLiteral(value)
		if err != nil {
			return nil, err
		}
		args = append(args, "-c", key+"="+literal)
	}
	p.launch = LaunchContribution{Mode: NativeProject, Notice: NativeProjectNotice, Env: map[string]string{adapter.EnvVar: req.Home}, Args: args}
	p.report = Certification{Environment: adapter.ID, ProviderVersion: req.ProviderVersion, Platform: req.Platform, Mode: req.Mode, Notice: NativeProjectNotice, Certified: []SurfaceName{Context, Skills, MCP, MemoryOff}, Refused: []Refusal{{NativeSourceExclusion, codexExclusionReason}, {Permissions, codexNativeProjectReason}}}

	sort.Slice(p.notes, func(i, j int) bool { return p.notes[i].Key < p.notes[j].Key })
	return p, nil
}
func configValue(config map[string]any, key string) any {
	var v any = config
	for _, part := range strings.Split(key, ".") {
		v = v.(map[string]any)[part]
	}
	return v
}

// tomlLiteral uses the vendor's native -c TOML value grammar, never shell text.
func tomlLiteral(value any) (string, error) {
	switch v := value.(type) {
	case string:
		// Encode strings with the same TOML encoder as config.toml.
		var b bytes.Buffer
		if err := toml.NewEncoder(&b).Encode(map[string]any{"value": v}); err != nil {
			return "", err
		}
		return strings.TrimPrefix(strings.TrimSpace(b.String()), "value = "), nil
	case bool:
		return strconv.FormatBool(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case []any:
		parts := []string{}
		for _, item := range v {
			s, err := tomlLiteral(item)
			if err != nil {
				return "", err
			}
			parts = append(parts, s)
		}
		return "[" + strings.Join(parts, ", ") + "]", nil
	case []string:
		parts := []string{}
		for _, item := range v {
			s, err := tomlLiteral(item)
			if err != nil {
				return "", err
			}
			parts = append(parts, s)
		}
		return "[" + strings.Join(parts, ", ") + "]", nil
	case map[string]any:
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		parts := []string{}
		for _, key := range keys {
			k, err := tomlLiteral(key)
			if err != nil {
				return "", err
			}
			item, err := tomlLiteral(v[key])
			if err != nil {
				return "", err
			}
			parts = append(parts, k+" = "+item)
		}
		return "{" + strings.Join(parts, ", ") + "}", nil
	default:
		return "", fmt.Errorf("project_context_unsupported: unsupported override value")
	}
}
func surface(b []byte, mode uint32) Surface {
	sum := sha256.Sum256(b)
	return Surface{hex.EncodeToString(sum[:]), int64(len(b)), mode}
}
func validateInput(files map[string][]byte, receipt map[string]Surface) error {
	if files == nil || receipt == nil || len(files) != len(receipt) || len(files) > 10000 {
		return fmt.Errorf("project_context_drift: composition receipt inventory mismatch")
	}
	aliases := map[string]bool{}
	total := 0
	for name, b := range files {
		if !safePath(name) {
			return fmt.Errorf("project_context_invalid: unsafe output path")
		}
		alias := cases.Fold().String(norm.NFC.String(name))
		if aliases[alias] {
			return fmt.Errorf("project_context_invalid: output path alias")
		}
		aliases[alias] = true
		expected, ok := receipt[name]
		if !ok || expected.Mode != 0600 || expected != surface(b, expected.Mode) {
			return fmt.Errorf("project_context_drift: composition bytes differ from receipt")
		}
		if len(b) > 16<<20 {
			return fmt.Errorf("project_context_invalid: output too large")
		}
		total += len(b)
	}
	if total > 64<<20 {
		return fmt.Errorf("project_context_invalid: composition too large")
	}
	for name := range aliases {
		parts := strings.Split(name, "/")
		for i := 1; i < len(parts); i++ {
			if aliases[strings.Join(parts[:i], "/")] {
				return fmt.Errorf("project_context_invalid: output file ancestor collision")
			}
		}
	}
	return nil
}
func safePath(name string) bool {
	if name == "" || path.Clean(name) != name || name == "." || strings.HasPrefix(name, "/") || strings.ContainsAny(name, "\\:\x00") {
		return false
	}
	for _, part := range strings.Split(name, "/") {
		if part == ".." || strings.EqualFold(part, ".git") {
			return false
		}
	}
	return true
}
