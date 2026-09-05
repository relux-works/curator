package contextmaterialize

import (
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextpkg"
)

// Production entry points under test: EmittedOrder, Header, Join,
// ChapterPart, Applicable, Monolithic, SystemPrompt, SurfaceHash, FileHash.

func weightLock() *contextlock.Lock {
	return &contextlock.Lock{Root: "root", Members: []contextlock.Member{
		{Kind: contextlock.KindContext, Name: "heavy", Commit: strings.Repeat("a", 40), Source: "https://example.com/heavy", Version: "1.0.0", Weight: 9, RequiredBy: []string{"root"}},
		{Kind: contextlock.KindContext, Name: "light", Commit: strings.Repeat("b", 40), Source: "https://example.com/light", Version: "1.0.0", Weight: 1, RequiredBy: []string{"root"}},
		{Kind: contextlock.KindContext, Name: "root", Commit: strings.Repeat("c", 40), Source: "https://example.com/root", Version: "1.0.0", Weight: 5},
	}}
}

func modulePackages() map[string]Package {
	return map[string]Package{
		"root":  {HasContext: true, Modules: []Module{{Module: contextpkg.Module{Path: "r.md"}, Bytes: []byte("root module\n")}}},
		"heavy": {HasContext: true, Modules: []Module{{Module: contextpkg.Module{Path: "h.md"}, Bytes: []byte("heavy module\n")}}},
		"light": {HasContext: true, Modules: []Module{{Module: contextpkg.Module{Path: "l.md"}, Bytes: []byte("light module\n")}}},
	}
}

func orderNames(t *testing.T, precedence Precedence) string {
	t.Helper()
	order, err := EmittedOrder(weightLock(), precedence)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, member := range order {
		names = append(names, member.Name)
	}
	return strings.Join(names, ",")
}

// TestEmittedOrderFollowsPrecedence pins both primitives: higher-weight
// winner-last emits ascending (light, root, heavy); higher-weight
// winner-first emits descending.
func TestEmittedOrderFollowsPrecedence(t *testing.T) {
	if got := orderNames(t, Precedence{Winner: WinnerHigher, Placement: PlacementLast}); got != "light,root,heavy" {
		t.Fatalf("winner-last order %s", got)
	}
	if got := orderNames(t, Precedence{Winner: WinnerLower, Placement: PlacementFirst}); got != "light,root,heavy" {
		t.Fatalf("lower-first order %s", got)
	}
	if got := orderNames(t, Precedence{Winner: WinnerHigher, Placement: PlacementFirst}); got != "heavy,root,light" {
		t.Fatalf("higher-first order %s", got)
	}
	if got := orderNames(t, Precedence{Winner: WinnerLower, Placement: PlacementLast}); got != "heavy,root,light" {
		t.Fatalf("lower-last order %s", got)
	}
}

// TestBadPrecedenceIsRejected narrows the precedence gate.
func TestBadPrecedenceIsRejected(t *testing.T) {
	if _, err := EmittedOrder(weightLock(), Precedence{Winner: "heaviest", Placement: PlacementLast}); err == nil {
		t.Fatal("bad winner must fail")
	}
	if _, err := EmittedOrder(weightLock(), Precedence{Winner: WinnerHigher, Placement: "middle"}); err == nil {
		t.Fatal("bad placement must fail")
	}
}

// TestMonolithicShape pins the document shape through the production
// Monolithic: the curator-root-context-v2 header, one chapter part per
// member with applicable modules, in emitted order.
func TestMonolithicShape(t *testing.T) {
	lock := weightLock()
	hash, err := lock.Hash()
	if err != nil {
		t.Fatal(err)
	}
	document, written, err := Monolithic(lock, hash, Precedence{Winner: WinnerHigher, Placement: PlacementLast}, "claude_code", modulePackages())
	if err != nil {
		t.Fatal(err)
	}
	if !written {
		t.Fatal("monolithic must be written")
	}
	text := string(document)
	if !strings.Contains(text, "<!--\n"+HeaderTypeLine+"\n") {
		t.Fatalf("document carries no v2 header:\n%q", text[:120])
	}
	light := strings.Index(text, "## Context: light 1.0.0")
	root := strings.Index(text, "## Context: root 1.0.0")
	heavy := strings.Index(text, "## Context: heavy 1.0.0")
	if light < 0 || root < 0 || heavy < 0 || light >= root || root >= heavy {
		t.Fatalf("chapter parts out of order:\n%s", text)
	}
	if got := FileHash(document); len(got) != len("sha256:")+64 {
		t.Fatalf("file hash %q", got)
	}
}

// TestNoChapterCase checks the root without context: nothing is written and
// no file exists.
func TestNoChapterCase(t *testing.T) {
	lock := weightLock()
	hash, err := lock.Hash()
	if err != nil {
		t.Fatal(err)
	}
	packages := modulePackages()
	packages["root"] = Package{HasContext: false}
	_, written, err := Monolithic(lock, hash, DefaultPrecedence, "claude_code", packages)
	if err != nil {
		t.Fatal(err)
	}
	if written {
		t.Fatal("a root without context must write nothing")
	}
}

// TestJoinSeparatesPartsWithOneBlankLine pins part joining.
func TestJoinSeparatesPartsWithOneBlankLine(t *testing.T) {
	joined := Join([][]byte{[]byte("a\n"), []byte("b\n")})
	if string(joined) != "a\n\nb\n" {
		t.Fatalf("joined %q", joined)
	}
}

// TestApplicableHonoursSelectors checks module selection by class and
// environment.
func TestApplicableHonoursSelectors(t *testing.T) {
	pkg := Package{HasContext: true, Modules: []Module{
		{Module: contextpkg.Module{Path: "all.md"}, Bytes: []byte("all\n")},
		{Module: contextpkg.Module{Path: "sys.md", Class: "system"}, Bytes: []byte("sys\n")},
		{Module: contextpkg.Module{Path: "scoped.md", Environments: []string{"codex_cli"}}, Bytes: []byte("scoped\n")},
	}}
	modules := Applicable(pkg, "root", "claude_code")
	if len(modules) != 1 || modules[0].Path != "all.md" {
		t.Fatalf("applicable %+v", modules)
	}
	if got := Applicable(pkg, "system", "claude_code"); len(got) != 1 {
		t.Fatalf("system applicable %+v", got)
	}
}
