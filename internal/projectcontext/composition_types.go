package projectcontext

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"
	"unicode"

	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/identifiers"
)

// CompositionRecord is the closed project-composition/1 wire record. The
// selected profile lock remains immutable; derived details are supporting data.
type CompositionRecord struct {
	Kind                  string                    `json:"kind"`
	RegistrationRef       string                    `json:"registration_ref"`
	RegistrationSHA256    string                    `json:"registration_sha256"`
	GenerationSHA256      string                    `json:"generation_sha256"`
	Profile               Profile                   `json:"profile"`
	Environment           string                    `json:"environment"`
	Mode                  string                    `json:"mode"`
	AdapterRef            string                    `json:"adapter_ref"`
	AdapterSHA256         string                    `json:"adapter_sha256"`
	PolicyRef             string                    `json:"policy_ref"`
	PolicySHA256          string                    `json:"policy_sha256"`
	Memory                string                    `json:"memory"`
	HomeRef               string                    `json:"home_ref"`
	WorktreeRef           string                    `json:"worktree_ref"`
	MaterializationSHA256 string                    `json:"materialization_sha256"`
	CommandOwners         []CompositionCommandOwner `json:"command_owners"`
	Omitted               []CompositionOmission     `json:"omitted"`
}

// CompositionCommandOwner binds a command to its complete audited artifact.
type CompositionCommandOwner struct {
	Name           string `json:"name"`
	ItemID         string `json:"item_id"`
	ArtifactSHA256 string `json:"artifact_sha256"`
}

// CompositionOmission records a disabled or unsupported selected item.
type CompositionOmission struct {
	ItemID string `json:"item_id"`
	Reason string `json:"reason"`
}

// GenerationItem binds an admitted item to its approval and content.
type GenerationItem struct {
	ItemID      string `json:"item_id"`
	ItemSHA256  string `json:"item_sha256"`
	ApprovalRef string `json:"approval_ref"`
}

// Generation identifies the exact admitted project item sequence.
type Generation struct {
	Kind               string           `json:"kind"`
	RegistrationRef    string           `json:"registration_ref"`
	RegistrationSHA256 string           `json:"registration_sha256"`
	Items              []GenerationItem `json:"items"`
}

// AdapterContract describes the explicit provider release and supported surfaces.
type AdapterContract struct {
	Kind            string   `json:"kind"`
	Environment     string   `json:"environment"`
	ProviderVersion string   `json:"provider_version"`
	Platform        string   `json:"platform"`
	Surfaces        []string `json:"surfaces"`
	Modes           []string `json:"modes"`
}

// MaterializationReceipt binds generated file identities to adapter and policy inputs.
type MaterializationReceipt struct {
	Version       int    `json:"version"`
	Files         []File `json:"files"`
	AdapterSHA256 string `json:"adapter_sha256"`
	PolicySHA256  string `json:"policy_sha256"`
	WorktreeRef   string `json:"worktree_ref"`
}

// CompositionPolicy is retained as protected data, never loaded from source.
// Its hash also binds the exact ordered module and effective artifact plan.
type CompositionPolicy struct {
	Version int                          `json:"version"`
	Memory  string                       `json:"memory"`
	System  contextmaterialize.Admission `json:"system"`
}

// ComposeRequest selects protected inputs for one derived composition.
type ComposeRequest struct {
	CheckoutID      string
	Mode            string
	Selections      []Selection
	Adapter         AdapterContract
	Policy          CompositionPolicy
	ConcurrentState bool
}

// Surface and Composition are the resolver contract shared with I-LEASE.
type Surface struct {
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
	Mode   uint32 `json:"mode"`
}

// Composition is the protected projection consumed by the lease authority.
type Composition struct {
	Ref             string             `json:"ref"`
	SHA256          string             `json:"sha256"`
	CheckoutID      string             `json:"checkout_id"`
	Profile         string             `json:"profile"`
	Environment     string             `json:"environment"`
	HomeRef         string             `json:"home_ref"`
	ConcurrentState bool               `json:"concurrent_state"`
	BuildKeys       []string           `json:"build_keys"`
	RuntimeEntries  []string           `json:"runtime_entries"`
	Generated       map[string]Surface `json:"generated"`
}

// CompositionPreview retains the derived records and in-memory output bytes.
type CompositionPreview struct {
	Record       CompositionRecord
	Generation   Generation
	PolicyRecord json.RawMessage
	Receipt      MaterializationReceipt
	Files        map[string][]byte `json:"-"`
	// Profile-only documents are recomputed from protected profile inputs.
	ProfileDocuments       map[string][]byte `json:"-"`
	ProfileDocumentReceipt []File            `json:"-"`
	Modules                []OrderedModule
	Skills                 []EffectiveSkill
	Ref                    string
	SHA256                 string
}

// OrderedModule records one module in effective rendering order.
type OrderedModule struct {
	ItemID  string
	Package string
	Path    string
	Class   string
	SHA256  string
}

// EffectiveSkill identifies the audited skill selected from either layer.
type EffectiveSkill struct {
	Layer          string
	Name           string
	Pin            string
	ItemID         string
	ArtifactSHA256 string
}

// JoinRootDocuments preserves all document bytes and skips empty documents.
// One additional LF separates adjacent non-empty documents (CIP-0002 §3).
func JoinRootDocuments(profile []byte, project ...[]byte) []byte {
	parts := [][]byte{}
	if len(profile) != 0 {
		parts = append(parts, profile)
	}
	for _, document := range project {
		if len(document) != 0 {
			parts = append(parts, document)
		}
	}
	return bytes.Join(parts, []byte{'\n'})
}
func tagged(ref, tag string) bool {
	return strings.HasPrefix(ref, tag+":") && idRE.MatchString(strings.TrimPrefix(ref, tag+":"))
}

// ParseCompositionRecord enforces the closed grammar, including presence/null
// distinctions which Go zero values do not retain. Authority is checked later.
func ParseCompositionRecord(b []byte) (CompositionRecord, error) {
	var out CompositionRecord
	if err := decode(b, &out); err != nil {
		return out, err
	}
	var raw map[string]json.RawMessage
	_ = json.Unmarshal(b, &raw)
	need := []string{"kind", "mode", "registration_ref", "registration_sha256", "generation_sha256", "profile", "environment", "adapter_ref", "adapter_sha256", "policy_ref", "policy_sha256", "memory", "home_ref", "worktree_ref", "materialization_sha256", "command_owners", "omitted"}
	for _, k := range need {
		if len(raw[k]) == 0 || bytes.Equal(raw[k], []byte("null")) {
			return out, fail("project_context_invalid", "missing or null composition member")
		}
	}
	valid := out.Kind == "project-composition/1" && validSourceMode(out.Mode) && out.Memory == "off" && environment(out.Environment) && identifiers.Valid(out.Profile.Name) && hexRE.MatchString(out.Profile.LockSHA256)
	for _, h := range []string{out.RegistrationSHA256, out.GenerationSHA256, out.AdapterSHA256, out.PolicySHA256, out.MaterializationSHA256} {
		valid = valid && hexRE.MatchString(h)
	}
	for tag, ref := range map[string]string{"registration": out.RegistrationRef, "adapter": out.AdapterRef, "policy": out.PolicyRef, "home": out.HomeRef, "worktree": out.WorktreeRef} {
		valid = valid && tagged(ref, tag)
	}
	if !valid || len(out.CommandOwners) > 1 {
		return out, fail("project_context_invalid", "invalid composition binding")
	}
	for _, o := range out.CommandOwners {
		if !identifiers.Valid(o.Name) || !idRE.MatchString(o.ItemID) || !hexRE.MatchString(o.ArtifactSHA256) {
			return out, fail("project_context_invalid", "invalid command owner")
		}
	}
	for _, o := range out.Omitted {
		if !idRE.MatchString(o.ItemID) || (o.Reason != "disabled" && o.Reason != "unsupported") {
			return out, fail("project_context_invalid", "invalid omission")
		}
	}
	// Re-encoding catches missing nested fields and null arrays/objects without
	// accepting unknown members or normalizing their values into authority.
	original := map[string]any{}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	_ = d.Decode(&original)
	encoded, err := canonical(out)
	if err != nil {
		return out, err
	}
	normalized, err := canonical(original)
	if err != nil || !bytes.Equal(encoded, normalized) {
		return out, fail("project_context_invalid", "missing or null nested composition member")
	}
	return out, nil
}
func validSourceMode(mode string) bool { return mode == "strict" || mode == "native-project" }

func (a AdapterContract) validate(env, mode string) error {
	if !validSourceMode(mode) {
		return fail("project_context_invalid", "unknown source mode")
	}
	if a.Kind != "project-adapter/1" || a.Environment != env || (a.Platform != "macos" && a.Platform != "linux") || invalidProviderVersion(a.ProviderVersion) || !sort.StringsAreSorted(a.Surfaces) {
		return fail("project_context_invalid", "invalid adapter release contract")
	}
	advertised := false
	if a.Modes == nil || !sort.StringsAreSorted(a.Modes) {
		return fail("project_context_invalid", "invalid adapter modes")
	}
	for i, m := range a.Modes {
		if !validSourceMode(m) || (i > 0 && m == a.Modes[i-1]) {
			return fail("project_context_invalid", "invalid adapter modes")
		}
		advertised = advertised || m == mode
	}
	if !advertised {
		return fail("project_context_unsupported", "adapter does not advertise requested mode")
	}
	seen := map[string]bool{}
	for _, s := range a.Surfaces {
		if seen[s] || !map[string]bool{"context": true, "skills": true, "mcp": true, "commands": true, "permissions": true, "memory-off": true, "native-source-exclusion": true}[s] {
			return fail("project_context_invalid", "invalid adapter surfaces")
		}
		seen[s] = true
	}
	required := []string{"context", "permissions", "memory-off"}
	if mode == "strict" {
		required = append(required, "native-source-exclusion")
	}
	for _, s := range required {
		if !seen[s] {
			return fail("project_context_unsupported", "adapter cannot reproduce required surface: "+s)
		}
	}
	return nil
}

func invalidProviderVersion(v string) bool {
	return v == "" || len(v) > 256 || strings.ContainsAny(v, "*<>=^~|,\ufeff") || strings.IndexFunc(v, unicode.IsSpace) >= 0
}
