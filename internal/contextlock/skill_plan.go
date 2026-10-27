package contextlock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unicode"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/protocoljson"
)

// SkillPlan is the closed profile-skill-plan v1 portable carrier (PSP01–05).
// Platform artifacts and authority records deliberately have no fields here.
type SkillPlan struct {
	SchemaVersion     int                `json:"schema_version"`
	Nodes             []SkillPin         `json:"nodes"`
	Edges             []SkillEdge        `json:"edges"`
	DirectConstraints []DirectConstraint `json:"direct_constraints"`
}

// SkillPin binds one selected skill package to exact source and content pins.
type SkillPin struct {
	Kind          string `json:"kind"`
	Name          string `json:"name"`
	Source        string `json:"source"`
	Commit        string `json:"commit"`
	Directory     string `json:"directory"`
	HashVersion   int    `json:"hash_version"`
	ContentSHA256 string `json:"content_sha256"`
}

// SkillConsumer identifies the context, skill or machine owning an edge.
type SkillConsumer struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

// SkillRef records the exact declared tag or revision for an incoming edge.
type SkillRef struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// SkillEdge carries one dependency edge and its requested activation.
type SkillEdge struct {
	Consumer SkillConsumer `json:"consumer"`
	Provider string        `json:"provider"`
	Ref      SkillRef      `json:"ref"`
	Mode     string        `json:"mode"`
	Commands []string      `json:"commands,omitempty"`
}

// DirectConstraint binds an overlapping direct declaration to its resolved pin.
type DirectConstraint struct {
	Name              string   `json:"name"`
	Source            string   `json:"source"`
	Commit            string   `json:"commit"`
	Directory         string   `json:"directory"`
	Ref               SkillRef `json:"ref"`
	DeclarationSHA256 string   `json:"declaration_sha256"`
}

// Object returns the portable plan representation for canonical encoding.
func (plan *SkillPlan) Object() map[string]any {
	payload, _ := json.Marshal(plan)
	var object map[string]any
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	_ = decoder.Decode(&object)
	return object
}

// Hash returns the SHA-256 binding over the plan's CCJ-1 representation.
func (plan *SkillPlan) Hash() (string, error) {
	payload, err := protocoljson.MarshalCanonical(plan.Object())
	if err != nil {
		return "", err
	}
	return HashBytes(payload), nil
}

// ParseSkillPlan strictly reads and validates the closed portable carrier.
func ParseSkillPlan(payload []byte) (*SkillPlan, error) {
	if err := protocoljson.Validate(payload); err != nil {
		return nil, fmt.Errorf("profile_skill_plan_invalid: %w", err)
	}
	var plan SkillPlan
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&plan); err != nil {
		return nil, fmt.Errorf("profile_skill_plan_invalid: %w", err)
	}
	// omitempty must not erase a present empty/null selection on a read.
	var raw struct {
		Edges []map[string]json.RawMessage `json:"edges"`
	}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, err
	}
	for _, edge := range raw.Edges {
		if value, ok := edge["commands"]; ok {
			var commands []string
			if json.Unmarshal(value, &commands) != nil || len(commands) == 0 {
				return nil, fmt.Errorf("profile_skill_plan_invalid: empty command selection")
			}
		}
	}
	if err := plan.Validate(); err != nil {
		return nil, err
	}
	return &plan, nil
}

func validSkillSource(source string) bool {
	canonical, err := identity.Parse("https://" + source)
	return err == nil && identity.ValidCanonical(source) && canonical == source
}

func validSkillDirectory(directory string) bool {
	return directory == "." || identifiers.PortablePath(directory)
}
func validSkillRef(ref SkillRef, commit string) bool {
	switch ref.Kind {
	case "revision":
		return commitRE.MatchString(ref.Value) && ref.Value == commit
	case "tag":
		return identity.DraftSourceRefName(ref.Value) && !strings.ContainsAny(ref.Value, "<>") && strings.IndexFunc(ref.Value, unicode.IsSpace) < 0
	}
	return false
}

// SkillEdgeKey orders edges by consumer kind, consumer name and provider.
func SkillEdgeKey(edge SkillEdge) string {
	return edge.Consumer.Kind + "\x00" + edge.Consumer.Name + "\x00" + edge.Provider
}

// Validate checks portable graph relations. Agreement with admitted package
// manifests and protected retained members is checked by closure.BuildProfile.
func (plan *SkillPlan) Validate() error {
	invalid := func(detail string) error { return fmt.Errorf("profile_skill_plan_invalid: %s", detail) }
	if plan == nil || plan.SchemaVersion != 1 || plan.Nodes == nil || plan.Edges == nil || plan.DirectConstraints == nil {
		return invalid("required plan fields")
	}
	nodes := map[string]SkillPin{}
	for i, pin := range plan.Nodes {
		if pin.Kind != KindSkill || !identifiers.Valid(pin.Name) || !validSkillSource(pin.Source) || !commitRE.MatchString(pin.Commit) || !validSkillDirectory(pin.Directory) || pin.HashVersion != 2 || !hex256RE.MatchString(pin.ContentSHA256) {
			return invalid("invalid skill pin")
		}
		if i > 0 && plan.Nodes[i-1].Name >= pin.Name {
			return invalid("nodes are not sorted and unique")
		}
		nodes[pin.Name] = pin
	}
	direct := map[string]DirectConstraint{}
	for i, constraint := range plan.DirectConstraints {
		pin, ok := nodes[constraint.Name]
		if !ok || constraint.Source != pin.Source || constraint.Commit != pin.Commit || constraint.Directory != pin.Directory || !validSkillRef(constraint.Ref, pin.Commit) || !validDigest(constraint.DeclarationSHA256) {
			return invalid("invalid direct constraint")
		}
		if i > 0 && plan.DirectConstraints[i-1].Name >= constraint.Name {
			return invalid("constraints are not sorted and unique")
		}
		direct[constraint.Name] = constraint
	}
	incoming := map[string]bool{}
	machine := map[string]bool{}
	graph := map[string][]string{}
	roots := []string{}
	for i, edge := range plan.Edges {
		pin, ok := nodes[edge.Provider]
		if !ok || !identifiers.Valid(edge.Consumer.Name) || !validSkillRef(edge.Ref, pin.Commit) {
			return invalid("invalid edge pin or consumer")
		}
		if i > 0 && SkillEdgeKey(plan.Edges[i-1]) >= SkillEdgeKey(edge) {
			return invalid("edges are not sorted and unique")
		}
		if edge.Mode != "full" && edge.Mode != "runtime" && edge.Mode != "context" {
			return invalid("invalid edge mode")
		}
		if edge.Commands != nil {
			if edge.Mode != "runtime" || len(edge.Commands) == 0 {
				return invalid("invalid selection mode")
			}
			for j, command := range edge.Commands {
				if !identifiers.Valid(command) || (j > 0 && edge.Commands[j-1] >= command) {
					return invalid("invalid command selection")
				}
			}
		}
		incoming[edge.Provider] = true
		switch edge.Consumer.Kind {
		case "context":
			roots = append(roots, edge.Provider)
		case "skill":
			if _, ok := nodes[edge.Consumer.Name]; !ok {
				return invalid("unknown skill consumer")
			}
			graph[edge.Consumer.Name] = append(graph[edge.Consumer.Name], edge.Provider)
		case "machine":
			constraint, ok := direct[edge.Provider]
			if !ok || edge.Consumer.Name != edge.Provider || edge.Mode != "full" || edge.Commands != nil || constraint.Ref != edge.Ref {
				return invalid("invented machine edge")
			}
			machine[edge.Provider] = true
		default:
			return invalid("invalid consumer kind")
		}
	}
	for name := range direct {
		if !machine[name] {
			return invalid("missing machine edge")
		}
	}
	state := map[string]int{}
	var visit func(string) error
	visit = func(name string) error {
		if state[name] == 1 {
			return invalid("cyclic graph")
		}
		if state[name] == 2 {
			return nil
		}
		state[name] = 1
		for _, provider := range graph[name] {
			if err := visit(provider); err != nil {
				return err
			}
		}
		state[name] = 2
		return nil
	}
	for _, name := range roots {
		if err := visit(name); err != nil {
			return err
		}
	}
	for name := range nodes {
		if !incoming[name] || state[name] != 2 {
			return invalid("unreachable node")
		}
	}
	return nil
}

func validDigest(value string) bool {
	return strings.HasPrefix(value, "sha256:") && hex256RE.MatchString(strings.TrimPrefix(value, "sha256:"))
}

func (lock *Lock) validateSkillPlan() error {
	if err := lock.SkillPlan.Validate(); err != nil {
		return err
	}
	digest, err := lock.SkillPlan.Hash()
	if err != nil {
		return err
	}
	if digest != lock.SkillPlanSHA256 {
		return fmt.Errorf("profile_skill_plan_binding_invalid: plan digest differs")
	}
	for _, pin := range lock.SkillPlan.Nodes {
		member, ok := lock.Find(KindSkill, pin.Name)
		if !ok || member.Source != pin.Source || member.Commit != pin.Commit || member.Directory != pin.Directory {
			return fmt.Errorf("profile_skill_plan_invalid: plan/member disagreement for %s", pin.Name)
		}
		required := []string{}
		seen := map[string]bool{}
		for _, edge := range lock.SkillPlan.Edges {
			if edge.Consumer.Kind == "context" {
				if _, ok := lock.Find(KindContext, edge.Consumer.Name); !ok {
					return fmt.Errorf("profile_skill_plan_invalid: unknown context consumer")
				}
			}
			if edge.Provider == pin.Name && edge.Consumer.Kind != "machine" && !seen[edge.Consumer.Name] {
				required = append(required, edge.Consumer.Name)
				seen[edge.Consumer.Name] = true
			}
		}
		sort.Strings(required)
		if !reflect.DeepEqual(required, member.RequiredBy) {
			return fmt.Errorf("profile_skill_plan_invalid: required_by differs")
		}
	}
	return nil
}

// VerifySkillBinding rejects legacy runtime enablement and an externally
// supplied stale/tampered lock binding. It never resolves or upgrades a ref.
func (lock *Lock) VerifySkillBinding(expected string) error {
	if lock.SchemaVersion != SchemaVersion5 {
		return fmt.Errorf("profile_skill_plan_required: explicit update/reinstall required")
	}
	actual, err := lock.Hash()
	if err != nil {
		return err
	}
	if actual != expected {
		return fmt.Errorf("profile_skill_plan_binding_invalid: lock digest differs")
	}
	return nil
}
