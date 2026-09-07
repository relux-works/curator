package probe

import (
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/skill-agents-management/pkg/agentic"
	"github.com/relux-works/skill-agents-management/pkg/agentic/systems/claude"
	"github.com/relux-works/skill-agents-management/pkg/agentic/systems/codex"
)

func keyOf(e string) string { k, _, _ := strings.Cut(e, "="); return k }

func toMap(env []string) map[string]string {
	m := map[string]string{}
	for _, e := range env {
		k, v, _ := strings.Cut(e, "=")
		m[k] = v
	}
	return m
}

// diff classifies Plan.Env against the parent the plan was requested with.
func diff(parent, child []string) (removed, changed, added []string) {
	p, c := toMap(parent), toMap(child)
	for k := range p {
		if cv, ok := c[k]; !ok {
			removed = append(removed, k)
		} else if cv != p[k] {
			changed = append(changed, k)
		}
	}
	for k := range c {
		if _, ok := p[k]; !ok {
			added = append(added, k)
		}
	}
	sort.Strings(removed); sort.Strings(changed); sort.Strings(added)
	return
}

var parentSeed = []string{
	"PATH=/usr/bin:/tmp/parent/codex-path:/bin",
	"HOME=/Users/op",
	"FAKE_SECRET=s3cr3t-inherited",
	"CLAUDECODE=1",
	"CLAUDE_CONFIG_DIR=/op/claude-home",
	"CODEX_THREAD_ID=t-parent",
	"TASK_BOARD_CODEX_APP_SERVER_AUTH_TOKEN_ENV=MY_TOK",
	"MY_TOK=tok-inherited",
	"TASK_BOARD_RUN_ID=RUN-parent",
	"TASK_BOARD_DIR=/op/board",
}

// Untracked mode: Plan.Env over the real parent is a FULL filtered environment.
func TestClaudePlanEnvIsAFullFilteredEnvironment(t *testing.T) {
	env, err := claude.New().ChildEnv(parentSeed, agentic.LaunchRequest{})
	if err != nil { t.Fatal(err) }
	removed, changed, added := diff(parentSeed, env)
	t.Logf("claude removed=%v changed=%v added=%v", removed, changed, added)
	want := []string{"CLAUDECODE", "TASK_BOARD_DIR", "TASK_BOARD_RUN_ID"}
	if strings.Join(removed, ",") != strings.Join(want, ",") { t.Fatalf("removed %v want %v", removed, want) }
	if len(changed)+len(added) != 0 { t.Fatalf("expected pure removal, got changed=%v added=%v", changed, added) }
	if toMap(env)["FAKE_SECRET"] != "s3cr3t-inherited" { t.Fatal("inherited value not retained in Plan.Env") }
}

func TestCodexPlanEnvRemovesRuntimeFamilyAndRewritesPATH(t *testing.T) {
	env, err := codex.New().ChildEnv(parentSeed, agentic.LaunchRequest{})
	if err != nil { t.Fatal(err) }
	removed, changed, added := diff(parentSeed, env)
	t.Logf("codex removed=%v changed=%v added=%v PATH=%q", removed, changed, added, toMap(env)["PATH"])
	want := []string{"CODEX_THREAD_ID", "MY_TOK", "TASK_BOARD_CODEX_APP_SERVER_AUTH_TOKEN_ENV", "TASK_BOARD_DIR", "TASK_BOARD_RUN_ID"}
	if strings.Join(removed, ",") != strings.Join(want, ",") { t.Fatalf("removed %v want %v", removed, want) }
	if strings.Join(changed, ",") != "PATH" { t.Fatalf("changed %v want [PATH]", changed) }
	if len(added) != 0 { t.Fatalf("added %v", added) }
}

// Negative: the naive "every Plan.Env entry is an env_literal" rule serializes an inherited secret.
func TestNaiveLiteralsLeakInheritedValues(t *testing.T) {
	env, _ := claude.New().ChildEnv(parentSeed, agentic.LaunchRequest{})
	if _, leaked := toMap(env)["FAKE_SECRET"]; !leaked {
		t.Fatal("probe premise broken: the naive rule would not have leaked here")
	}
}

// Tracked mode: request the plan with Env == nil; Plan.Env is then the plugin's OWN names only.
func TestNilParentYieldsOwnNamesOnly(t *testing.T) {
	for name, sys := range map[string]agentic.System{"claude": claude.New(), "codex": codex.New()} {
		env, err := sys.ChildEnv(nil, agentic.LaunchRequest{})
		if err != nil { t.Fatal(err) }
		if len(env) != 0 { t.Fatalf("%s: nil parent, zero run context: want empty own-name set, got %v", name, env) }
		req := agentic.LaunchRequest{Run: agentic.RunContext{RunID: "RUN-x", TaskID: "TASK-y", BoardDir: "/abs/board", ContextID: "CTX-z"}, ServiceTier: "flex"}
		env, err = sys.ChildEnv(nil, req)
		if err != nil { t.Fatal(err) }
		var keys []string
		for _, e := range env { keys = append(keys, keyOf(e)) }
		sort.Strings(keys)
		t.Logf("%s own names with run context: %v", name, keys)
		for _, k := range keys {
			if !strings.HasPrefix(k, "TASK_BOARD_") { t.Fatalf("%s: non-own name %q in nil-parent plan", name, k) }
		}
	}
}

// The strip list is invisible under a nil parent: a destination-side CLAUDECODE cannot be expressed.
func TestStripIsUnexpressibleWithoutTheParent(t *testing.T) {
	env, _ := claude.New().ChildEnv(nil, agentic.LaunchRequest{})
	for _, e := range env { if keyOf(e) == "CLAUDECODE" { t.Fatal("unexpected") } }
	// Nothing in env says "remove CLAUDECODE"; a launch-plan document has no member for it.
}
