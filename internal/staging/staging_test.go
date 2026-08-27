package staging

import (
	"strings"
	"testing"
)

func TestSortedOrdersClassesThenIdentifiersWithConsumerLast(t *testing.T) {
	var plan Plan
	plan.Replace(ClassConsumer, "consumers", "/home/consumers.json", "/stage/consumers.json")
	plan.RemoveEntry("adapter/root/skill-gone", "/project/.claude/skills/skill-gone")
	plan.ReplaceEntry(ClassAdapterLedger, "root/entry/skill-a", "/project/.claude/skills/skill-a", "/stage/link")
	plan.Replace(ClassAdapterLedger, "root/ledger", "/project/.claude/skills/.csk-managed.json", "/stage/ledger")
	plan.Replace(ClassContext, "project/skill-a", "/project/.agents/skills/skill-a", "/stage/skill-a")

	var order []string
	for _, target := range plan.Sorted() {
		order = append(order, target.Class+"/"+target.Identifier)
	}
	want := []string{
		ClassContext + "/project/skill-a",
		// "entry/..." sorts before "ledger", so a ledger never claims an entry
		// that is not durable yet.
		ClassAdapterLedger + "/root/entry/skill-a",
		ClassAdapterLedger + "/root/ledger",
		ClassRemoval + "/adapter/root/skill-gone",
		ClassConsumer + "/consumers",
	}
	if strings.Join(order, " ") != strings.Join(want, " ") {
		t.Fatalf("commit order = %v, want %v", order, want)
	}
	if last := plan.Sorted()[len(want)-1]; last.Class != ClassConsumer {
		t.Fatalf("last class = %q, want the consumer ledger", last.Class)
	}
}

func TestRemovalAndEntryKindsAreRecorded(t *testing.T) {
	var plan Plan
	plan.ReplaceEntry(ClassAdapterLedger, "root/entry/skill-a", "/live/skill-a", "/stage/skill-a")
	plan.RemoveEntry("adapter/root/skill-gone", "/live/skill-gone")
	plan.Remove("skill/project/old", "/live/old")
	plan.Replace(ClassEnvFile, "env.sh", "/live/env.sh", "/stage/env.sh")

	byIdentifier := map[string]Target{}
	for _, target := range plan.Targets {
		byIdentifier[target.Identifier] = target
	}
	if got := byIdentifier["root/entry/skill-a"]; got.Kind != KindEntry || got.Removal() {
		t.Fatalf("staged mirror entry = %+v, want the entry kind and a desired replacement", got)
	}
	if got := byIdentifier["adapter/root/skill-gone"]; got.Kind != KindEntry || !got.Removal() || got.Class != ClassRemoval {
		t.Fatalf("stale mirror removal = %+v, want an absent entry-kind target in the removal class", got)
	}
	if got := byIdentifier["skill/project/old"]; got.Kind != KindBytes || !got.Removal() {
		t.Fatalf("stale byte removal = %+v, want an absent byte target", got)
	}
	if got := byIdentifier["env.sh"]; got.Kind != KindBytes || got.Removal() {
		t.Fatalf("env file = %+v, want a byte replacement", got)
	}
}

func TestValidateRejectsProducerDefects(t *testing.T) {
	for name, plan := range map[string]Plan{
		"no class": {Targets: []Target{{Identifier: "x", LivePath: "/live/x"}}},
		"no identifier": {Targets: []Target{
			{Class: ClassContext, LivePath: "/live/x"},
		}},
		"relative live path": {Targets: []Target{
			{Class: ClassContext, Identifier: "x", LivePath: "live/x"},
		}},
		"unclean live path": {Targets: []Target{
			{Class: ClassContext, Identifier: "x", LivePath: "/live/../live/x"},
		}},
		"relative staged path": {Targets: []Target{
			{Class: ClassContext, Identifier: "x", LivePath: "/live/x", StagedPath: "stage/x"},
		}},
		"unknown kind": {Targets: []Target{
			{Class: ClassContext, Identifier: "x", Kind: "bogus", LivePath: "/live/x"},
		}},
		"duplicate identifier": {Targets: []Target{
			{Class: ClassContext, Identifier: "x", LivePath: "/live/one"},
			{Class: ClassContext, Identifier: "x", LivePath: "/live/two"},
		}},
		"two producers on one live path": {Targets: []Target{
			{Class: ClassContext, Identifier: "one", LivePath: "/live/x"},
			{Class: ClassAdapterLedger, Identifier: "two", LivePath: "/live/x"},
		}},
	} {
		if err := plan.Validate(); err == nil {
			t.Fatalf("%s was accepted", name)
		}
	}

	var valid Plan
	valid.Replace(ClassContext, "project/skill-a", "/live/skill-a", "/stage/skill-a")
	valid.ReplaceEntry(ClassAdapterLedger, "root/entry/skill-a", "/mirror/skill-a", "/stage/link")
	valid.RemoveEntry("adapter/root/skill-gone", "/mirror/skill-gone")
	if err := valid.Validate(); err != nil {
		t.Fatalf("a well-formed plan was rejected: %v", err)
	}
	if valid.Empty() {
		t.Fatal("a plan with targets reports itself empty")
	}
	if !(Plan{}).Empty() {
		t.Fatal("an empty plan does not report itself empty")
	}
}

func TestMergeKeepsEveryTarget(t *testing.T) {
	var left, right Plan
	left.Replace(ClassContext, "a", "/live/a", "/stage/a")
	right.ReplaceEntry(ClassAdapterLedger, "b", "/live/b", "/stage/b")
	left.Merge(right)
	if len(left.Targets) != 2 {
		t.Fatalf("merged plan has %d targets, want 2", len(left.Targets))
	}
	if left.Targets[1].Kind != KindEntry {
		t.Fatalf("merge lost the target kind: %+v", left.Targets[1])
	}
}
