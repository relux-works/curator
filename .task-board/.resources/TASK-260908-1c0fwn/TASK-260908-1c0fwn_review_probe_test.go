package probe

import (
	"strings"
	"testing"

	"github.com/relux-works/skill-agents-management/pkg/agentic"
	"github.com/relux-works/skill-agents-management/pkg/agentic/systems/claude"
	"github.com/relux-works/skill-agents-management/pkg/agentic/systems/codex"
)

func toMap(env []string) map[string]string {
	m := map[string]string{}
	for _, e := range env { k, v, _ := strings.Cut(e, "="); m[k] = v }
	return m
}

var parent = []string{"PATH=/usr/bin:/bin", "HOME=/h", "SECRET=x", "CLAUDECODE=1", "TASK_BOARD_RUN_ID=old", "CLAUDE_CONFIG_DIR=/op"}

// T5 at v0.5.10: own names (nil parent) are a value-equal subset of Plan.Env over a real parent.
func TestOwnNamesSubsetOfFullEnv(t *testing.T) {
	req := agentic.LaunchRequest{Run: agentic.RunContext{RunID: "R", TaskID: "T", BoardDir: "/b", ContextID: "C"}, ServiceTier: "flex"}
	for name, sys := range map[string]agentic.System{"claude": claude.New(), "codex": codex.New()} {
		own, err := sys.ChildEnv(nil, req); if err != nil { t.Fatal(err) }
		full, err := sys.ChildEnv(parent, req); if err != nil { t.Fatal(err) }
		o, f := toMap(own), toMap(full)
		for k, v := range o { if f[k] != v { t.Fatalf("%s: own %s=%q but full has %q", name, k, v, f[k]) } }
		if len(o) != 5 { t.Fatalf("%s: own names %v", name, o) }
		if _, ok := o["SECRET"]; ok { t.Fatal("inherited leaked into own names") }
		if _, ok := f["CLAUDECODE"]; ok && name == "claude" { t.Fatal("CLAUDECODE not stripped") }
		if f["TASK_BOARD_RUN_ID"] != "R" { t.Fatalf("%s: run id not rewritten: %q", name, f["TASK_BOARD_RUN_ID"]) }
		t.Logf("%s own=%v", name, o)
	}
}

// Mutant: substituting Plan.Env for ChildEnv(nil) must be detectable (the narrowing the producer names in T2).
func TestMutantFullEnvAsLiteralsFails(t *testing.T) {
	full, _ := claude.New().ChildEnv(parent, agentic.LaunchRequest{})
	if _, ok := toMap(full)["SECRET"]; !ok { t.Fatal("mutant undetectable") }
}
