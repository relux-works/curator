package staging

import (
	"path/filepath"
	"strings"
	"testing"
)

// absoluteTestRoot is the host's own filesystem root: "/" on unix, and the
// volume of the working directory on Windows. Plan.Validate demands an
// OS-absolute, already-clean live and staged path, and a POSIX literal is
// neither on Windows -- filepath.IsAbs rejects "/live/x" for having no volume,
// and filepath.Clean rewrites its separators. Building the fixtures from the
// host root keeps the case about path shape rather than POSIX syntax.
var absoluteTestRoot = func() string {
	root, err := filepath.Abs(string(filepath.Separator))
	if err != nil {
		panic("resolve the host filesystem root: " + err.Error())
	}
	return root
}()

// absoluteTestPath joins an absolute, clean host path from POSIX-shaped parts.
func absoluteTestPath(elements ...string) string {
	return filepath.Join(absoluteTestRoot, filepath.Join(elements...))
}

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
	live := absoluteTestPath("live", "x")
	// Absolute but not yet clean, so the defect under test is the unresolved
	// element rather than a path that is merely relative on this host.
	unclean := absoluteTestRoot + strings.Join([]string{"live", "..", "live", "x"}, string(filepath.Separator))
	for name, plan := range map[string]Plan{
		"no class": {Targets: []Target{{Identifier: "x", LivePath: live}}},
		"no identifier": {Targets: []Target{
			{Class: ClassContext, LivePath: live},
		}},
		"relative live path": {Targets: []Target{
			{Class: ClassContext, Identifier: "x", LivePath: filepath.Join("live", "x")},
		}},
		"unclean live path": {Targets: []Target{
			{Class: ClassContext, Identifier: "x", LivePath: unclean},
		}},
		"relative staged path": {Targets: []Target{
			{Class: ClassContext, Identifier: "x", LivePath: live, StagedPath: filepath.Join("stage", "x")},
		}},
		"unknown kind": {Targets: []Target{
			{Class: ClassContext, Identifier: "x", Kind: "bogus", LivePath: live},
		}},
		"duplicate identifier": {Targets: []Target{
			{Class: ClassContext, Identifier: "x", LivePath: absoluteTestPath("live", "one")},
			{Class: ClassContext, Identifier: "x", LivePath: absoluteTestPath("live", "two")},
		}},
		"two producers on one live path": {Targets: []Target{
			{Class: ClassContext, Identifier: "one", LivePath: live},
			{Class: ClassAdapterLedger, Identifier: "two", LivePath: live},
		}},
	} {
		if err := plan.Validate(); err == nil {
			t.Fatalf("%s was accepted", name)
		}
	}

	var valid Plan
	valid.Replace(ClassContext, "project/skill-a", absoluteTestPath("live", "skill-a"), absoluteTestPath("stage", "skill-a"))
	valid.ReplaceEntry(ClassAdapterLedger, "root/entry/skill-a", absoluteTestPath("mirror", "skill-a"), absoluteTestPath("stage", "link"))
	valid.RemoveEntry("adapter/root/skill-gone", absoluteTestPath("mirror", "skill-gone"))
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
