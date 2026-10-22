// Package projectadapter renders protected composition bytes for a native home.
// It starts no provider and grants no admission or lease authority.
package projectadapter

// Surface is the candidate shared receipt entry, temporarily identical to
// projectlease.Surface. Replace this definition with that import at integration;
// the adapter consumes the lease contract, never the lease implementation.
type Surface struct {
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
	Mode   uint32 `json:"mode"`
}

type SurfaceName string

const (
	Context               SurfaceName = "context"
	Skills                SurfaceName = "skills"
	MCP                   SurfaceName = "mcp"
	Commands              SurfaceName = "commands"
	Permissions           SurfaceName = "permissions"
	MemoryOff             SurfaceName = "memory-off"
	NativeSourceExclusion SurfaceName = "native-source-exclusion"
	LaunchForm            SurfaceName = "launch-form"
)

type Refusal struct {
	Surface SurfaceName `json:"surface"`
	Reason  string      `json:"reason"`
}

// NativeProjectNotice must accompany status and launch in this mode.
const NativeProjectNotice = "project sources not isolated"

// Mode names are the owner-selected wave-1 modes, pending wire integration.
const (
	NativeProject = "native-project"
	Strict        = "strict"
)

// Certification is a candidate shared definition for native adapters.
// It includes the published Muse candidate fields; JSON spelling and the
// optional launch notice fields must be reconciled when the branches integrate.
// Certified concerns the named generated surfaces at this exact release/platform.
// It is not an admission record. In particular a refused exclusion surface MUST
// prevent a caller from advertising a strict project-adapter/1 contract.
type Certification struct {
	Notice          string        `json:"notice"`
	Environment     string        `json:"environment"`
	ProviderVersion string        `json:"provider_version"`
	Platform        string        `json:"platform"`
	Mode            string        `json:"mode"`
	Form            string        `json:"form,omitempty"`
	Certified       []SurfaceName `json:"certified"`
	Refused         []Refusal     `json:"refused"`
}

// Unsupported is a typed capability refusal, never an optional omission.
type Unsupported struct {
	Surface SurfaceName
	Reason  string
}

func (e *Unsupported) Code() string {
	if e.Surface == MCP {
		return "mcp_contribution_policy_refused"
	}
	return "project_context_unsupported"
}
func (e *Unsupported) Error() string {
	return e.Code() + ": " + string(e.Surface) + ": " + e.Reason
}

// LaunchContribution uses the existing adapter's home variable and native argv
// channels. It is inert data for a manager-side delegated consumer. Consumers
// still need current authority, an active lease and certification before start.
type LaunchContribution struct {
	Mode   string
	Notice string
	Env    map[string]string
	Args   []string
}

// Note records the exact plan-owned seed key that was replaced or dropped.
// It carries names and disposition, never values or credential bytes.
type Note struct {
	Key         string `json:"key"`
	Disposition string `json:"disposition"`
}
