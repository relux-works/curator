// Package projectadapter prepares native-layout data from protected project
// compositions. Offline output does not certify a provider or authorize launch.
package projectadapter

// Integration candidate: reconcile this file with the Claude/Codex adapter
// owners before integrating. Surface mirrors the I-LEASE receipt exactly;
// this package does not own lease state or acquire a second manager lock.
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
	Surface SurfaceName
	Reason  string
}

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
func (e *Unsupported) Error() string { return e.Code() + ": " + string(e.Surface) + ": " + e.Reason }
