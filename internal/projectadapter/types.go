// Package projectadapter prepares native-layout data from protected project
// compositions. Offline output does not certify a provider or authorize launch.
package projectadapter

// Surface mirrors the I-LEASE receipt exactly. Integration must reconcile
// this contract with the Claude/Codex adapter owners;
// this package does not own lease state or acquire a second manager lock.
type Surface struct {
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
	Mode   uint32 `json:"mode"`
}

// SurfaceName identifies a provider surface in a certification report.
type SurfaceName string

// Surface names cover the manager-controlled outputs and launch constraints.
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

// Refusal records an unqualified surface and its reason.
type Refusal struct {
	Surface SurfaceName
	Reason  string
}

// Certification reports qualified and refused surfaces for an exact provider tuple.
type Certification struct {
	Environment, ProviderVersion, Platform, Mode string
	Form, Notice                                 string
	Certified                                    []SurfaceName
	Refused                                      []Refusal
}

// LaunchContribution is returned only after mandatory native qualification.
// An unsuccessful contribution must leave both members nil.
type LaunchContribution struct {
	Env  map[string]string
	Args []string
}

// Unsupported is a typed refusal for an unsupported project contribution.
type Unsupported struct {
	Surface SurfaceName
	Reason  string
}

// Code returns the diagnostic family for the refused surface.
func (e *Unsupported) Code() string {
	if e.Surface == MCP {
		return "mcp_contribution_policy_refused"
	}
	return "project_context_unsupported"
}
func (e *Unsupported) Error() string { return e.Code() + ": " + string(e.Surface) + ": " + e.Reason }
