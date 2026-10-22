package projectmcpstate

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/projectmcp"
	"github.com/relux-works/curator/internal/protocoljson"
)

// DecodePlan checks carrier structure, total unique pair ordering, closed
// stage-A codes and its independent digest. It grants no installation authority.
func DecodePlan(payload []byte) (*Plan, error) {
	var plan Plan
	raw, err := decodeRecord(payload, &plan, "schema_version", "capability", "project_id", "manifest_sha256", "lock_sha256", "dry_run", "rows", "install_options", "plan_sha256")
	if err != nil {
		return nil, err
	}
	rows, ok := raw["rows"].([]any)
	if !ok {
		return nil, invalid("MCP plan rows")
	}
	for _, value := range rows {
		row, ok := value.(map[string]any)
		if !ok || !required(row, "name", "environment", "path", "outcome", "code") {
			return nil, invalid("MCP plan row fields")
		}
	}
	options, ok := raw["install_options"].([]any)
	if !ok {
		return nil, invalid("MCP plan options")
	}
	for _, value := range options {
		option, ok := value.(map[string]any)
		if !ok || !required(option, "name", "option", "handling") {
			return nil, invalid("MCP option fields")
		}
		if _, ok := option["option"].(string); !ok {
			return nil, invalid("MCP option name")
		}
	}
	if err := plan.Validate(); err != nil {
		return nil, err
	}
	return &plan, nil
}

// Validate checks observations, not native preimages or protected approvals.
func (plan *Plan) Validate() error {
	if plan.SchemaVersion != 1 || plan.Capability != projectmcp.Capability || !validProject(plan.ProjectID) || !digestRE.MatchString(plan.ManifestSHA256) || plan.Rows == nil || plan.InstallOptions == nil {
		return invalid("MCP plan header")
	}
	if plan.LockSHA256 != nil && !digestRE.MatchString(*plan.LockSHA256) {
		return invalid("MCP plan lock identity")
	}
	refused, success := false, false
	previous := ""
	for _, row := range plan.Rows {
		key := row.Name + "\x00" + row.Environment
		if !projectmcp.ValidName(row.Name) || !projectmcp.ValidEnvironment(row.Environment) || key <= previous {
			return invalid("MCP pair order")
		}
		previous = key
		if row.Path != nil && (!identifiers.PortablePath(*row.Path) || *row.Path == ".") {
			return invalid("MCP target path")
		}
		if row.Outcome == "refused" {
			if row.Code == nil || !refusalCodes[*row.Code] {
				return invalid("MCP refusal code")
			}
			refused = true
		} else if row.Code != nil {
			return invalid("MCP nonrefusal code")
		}
		switch row.Outcome {
		case "added", "updated", "unchanged", "restored", "removed", "released":
			success = true
			if row.Path == nil {
				return invalid("MCP successful path")
			}
		case "skipped: unsupported":
			if row.Path != nil {
				return invalid("MCP unsupported path")
			}
		case "skipped: drift":
			if row.Path == nil {
				return invalid("MCP drift path")
			}
		case "conflict: foreign entry":
			refused = true
			if row.Path == nil {
				return invalid("MCP foreign path")
			}
		case "refused":
		default:
			return invalid("MCP outcome")
		}
	}
	if refused && success {
		return invalid("MCP partial success on refusal")
	}
	if plan.LockSHA256 == nil && len(plan.Rows) > 0 && !refused {
		return invalid("MCP successful nonempty plan requires lock identity")
	}
	previous = ""
	for _, option := range plan.InstallOptions {
		key := option.Name + "\x00" + option.Option
		if !projectmcp.ValidName(option.Name) || option.Handling != "ignored" || key <= previous {
			return invalid("MCP ignored-option order")
		}
		previous = key
	}
	hash, err := carrierHash(plan, "plan_sha256")
	if err != nil {
		return err
	}
	if hash != plan.PlanSHA256 {
		return invalid("MCP plan digest")
	}
	return nil
}

// Approval is only the closed ownership-action record shape. Parsing it never
// establishes protected provenance, activates an action, or admits inline input.
type Approval struct {
	SchemaVersion int    `json:"schema_version"`
	Capability    string `json:"capability"`
	Kind          string `json:"kind"`
	ProjectID     string `json:"project_id"`
	Name          string `json:"name"`
	ServerSHA256  string `json:"server_sha256"`
	ContentSHA256 string `json:"content_sha256"`
	Operation     string `json:"operation"`
	Approver      string `json:"approver"`
	PolicySHA256  string `json:"policy_sha256"`
	State         string `json:"state"`
}

// DecodeApproval validates data only; no stage-A install path consumes it.
func DecodeApproval(payload []byte) (*Approval, error) {
	var record Approval
	if _, err := decodeRecord(payload, &record, "schema_version", "capability", "kind", "project_id", "name", "server_sha256", "content_sha256", "operation", "approver", "policy_sha256", "state"); err != nil {
		return nil, err
	}
	if record.SchemaVersion != 1 || record.Capability != projectmcp.Capability || record.Kind != "package-ownership-action" || !validProject(record.ProjectID) || !validProject(record.Approver) || !projectmcp.ValidName(record.Name) || !digestRE.MatchString(record.ServerSHA256) || !digestRE.MatchString(record.ContentSHA256) || !digestRE.MatchString(record.PolicySHA256) {
		return nil, invalid("MCP action record shape")
	}
	if record.State != "active" && record.State != "revoked" {
		return nil, invalid("MCP action record state")
	}
	switch record.Operation {
	case "replace", "restore", "remove", "release":
	default:
		return nil, invalid("MCP action operation")
	}
	return &record, nil
}

func decodeRecord(payload []byte, value any, fields ...string) (map[string]any, error) {
	if err := protocoljson.RequireCanonical(payload); err != nil {
		return nil, invalid("MCP carrier JSON")
	}
	var raw map[string]any
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(&raw); err != nil || !required(raw, fields...) {
		return nil, invalid("MCP carrier fields")
	}
	decoder = json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return nil, invalid("MCP carrier field types")
	}
	// Explicit null must never become a zero value for a required scalar.
	nullable := map[string]bool{"path": true, "code": true, "lock_sha256": true}
	for key, item := range raw {
		if item == nil && !nullable[key] {
			return nil, invalid("MCP carrier null " + strings.ReplaceAll(key, "\n", ""))
		}
	}
	return raw, nil
}
