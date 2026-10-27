package operatorpolicy

import (
	"bytes"
	"crypto/ed25519"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// eventSchema is copied from curator-spec e1051a7, workspace-trust-policy-grant-v1.
//
//go:embed event.schema.json
var eventSchema []byte
var compiledEvent = func() *jsonschema.Schema {
	c := jsonschema.NewCompiler()
	var doc any
	if e := json.Unmarshal(eventSchema, &doc); e != nil {
		panic(e)
	}
	if e := c.AddResource("event.json", doc); e != nil {
		panic(e)
	}
	s, e := c.Compile("event.json")
	if e != nil {
		panic(e)
	}
	return s
}()

func validEvent(e Event) error {
	b, _ := canonical(e)
	var doc any
	if json.Unmarshal(b, &doc) != nil || compiledEvent.Validate(doc) != nil {
		return refusal("workspace_trust_ledger_unreadable")
	}
	if e.Subject.Path != e.Subject.Repository.Root || !exactPath(e.Subject.Path) || e.Subject.Repository.Remote != e.Subject.Destination.remote() {
		return refusal("workspace_trust_ledger_unreadable")
	}
	if !validHooks(e.Surfaces.Hooks) {
		return refusal("workspace_trust_ledger_unreadable")
	}
	states := append(append([]State{}, e.History...), e.State)
	chain := ""
	var last time.Time
	for _, s := range states {
		t, err := timestamp(s.Time)
		if err != nil || t.Before(last) || (s.By == nil) != (s.Status == "lapsed") {
			return refusal("workspace_trust_ledger_unreadable")
		}
		chain += s.Status + "/"
		last = t
	}
	if chain != "active/" && chain != "active/revoked/" && chain != "active/lapsed/" && chain != "active/lapsed/revoked/" {
		return refusal("workspace_trust_ledger_unreadable")
	}
	start := states[0]
	deadline, err := timestamp(e.ReviewBy)
	began, _ := timestamp(e.Time)
	if err != nil || !deadline.After(began) || start.Time != e.Time || !reflect.DeepEqual(start.By, &e.Issuer) {
		return refusal("workspace_trust_ledger_unreadable")
	}
	return nil
}
func exactPath(p string) bool {
	if len(p) < 2 || len(p) > 4096 || !strings.HasPrefix(p, "/") || strings.ContainsAny(p, "\\*?[]\x00\n\r") {
		return false
	}
	for _, c := range strings.Split(p[1:], "/") {
		if c == "" || c == "." || c == ".." {
			return false
		}
	}
	return true
}
func (d Destination) remote() string {
	if !regexp.MustCompile(`^[a-z0-9]+(?:[.-][a-z0-9]+)*$`).MatchString(d.Host) || len(d.Host) > 253 || d.Port < 1 || d.Port > 65535 || !exactPath(d.Path) || strings.ContainsAny(d.Path, "%?#") {
		return ""
	}
	authority := d.Host
	defaultPort := 443
	switch d.Transport {
	case "https":
		if d.User != "" || len(d.Pins) != 0 {
			return ""
		}
	case "ssh":
		if !regexp.MustCompile(`^[A-Za-z0-9_.-]+$`).MatchString(d.User) || len(d.Pins) == 0 {
			return ""
		}
		seen := map[string]bool{}
		for _, p := range d.Pins {
			if !regexp.MustCompile(`^SHA256:[A-Za-z0-9+/]{43}$`).MatchString(p) || seen[p] {
				return ""
			}
			seen[p] = true
		}
		authority = d.User + "@" + authority
		defaultPort = 22
	default:
		return ""
	}
	if d.Port != defaultPort {
		authority += fmt.Sprintf(":%d", d.Port)
	}
	return d.Transport + "://" + authority + d.Path
}
func validHooks(h []Hook) bool {
	if h == nil || len(h) > 256 {
		return false
	}
	seen := map[string]bool{}
	for _, v := range h {
		if !idPattern.MatchString(v.ID) || !digestPattern.MatchString(v.SHA256) || seen[v.ID] {
			return false
		}
		seen[v.ID] = true
	}
	return true
}
func unique(ss []string) bool {
	if len(ss) == 0 || len(ss) > 256 {
		return false
	}
	seen := map[string]bool{}
	for _, s := range ss {
		if !idPattern.MatchString(s) || seen[s] {
			return false
		}
		seen[s] = true
	}
	return true
}
func has(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}
func validPolicy(p Policy) bool {
	t, e := timestamp(p.Time)
	until, u := timestamp(p.ReviewBy)
	if e != nil || u != nil || !until.After(t) || !idPattern.MatchString(p.ID) || strings.TrimSpace(p.Reason) == "" || len(p.Reason) > 4096 || !validHooks(p.Hooks) || p.Destination.remote() == "" || !unique(p.Environments) || !unique(p.Lanes) {
		return false
	}
	if !regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`).MatchString(p.RootCommit) {
		return false
	}
	for _, s := range p.Environments {
		if s != "claude_code" && s != "codex_cli" && s != "muse" {
			return false
		}
	}
	return len(p.Spellings) == 1 && p.Spellings[0] == p.Destination.remote()
}
func authenticate(keys map[string]string, s SignedCommand, now time.Time) error {
	c := s.Command
	i := c.Issuer
	t, e := timestamp(c.Time)
	if c.Kind != "operator-policy-command/1" || i.Role != "operator" || !idPattern.MatchString(i.Principal) || !digestPattern.MatchString(i.Approval) || !idPattern.MatchString(i.Actor.Principal) || (i.Actor.Kind != "operator" && i.Actor.Kind != "keeper") || strings.TrimSpace(c.Reason) == "" || len(c.Reason) > 4096 || e != nil || t.After(now) {
		return refusal("workspace_trust_policy_authority")
	}
	b, e := canonical(c)
	if e != nil {
		return refusal("workspace_trust_policy_authority")
	}
	pub, e := base64.StdEncoding.DecodeString(keys[i.Principal])
	sig, se := base64.StdEncoding.DecodeString(s.ApprovalSignature)
	actor, ae := base64.StdEncoding.DecodeString(c.ActorPublicKey)
	actorSig, ase := base64.StdEncoding.DecodeString(s.ActorSignature)
	if e != nil || se != nil || ae != nil || ase != nil || len(pub) != ed25519.PublicKeySize || len(actor) != ed25519.PublicKeySize || !ed25519.Verify(pub, b, sig) || !ed25519.Verify(actor, b, actorSig) {
		return refusal("workspace_trust_policy_authority")
	}
	if i.Actor.Kind == "operator" && (i.Actor.Principal != i.Principal || !bytes.Equal(pub, actor)) {
		return refusal("workspace_trust_policy_authority")
	}
	switch c.Operation {
	case "grant":
		if c.Policy == nil || c.Selection == nil || c.Policy.ID != c.EntryID || !validPolicy(*c.Policy) || c.Policy.Issuer != i || c.Policy.Time != c.Time || c.Policy.Reason != c.Reason {
			return refusal("workspace_trust_policy_authority")
		}
	case "revoke":
		if c.Policy != nil || c.Selection != nil || !idPattern.MatchString(c.EntryID) {
			return refusal("workspace_trust_policy_authority")
		}
	default:
		return refusal("workspace_trust_policy_authority")
	}
	return nil
}
