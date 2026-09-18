package contextmaterialize

import (
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextpkg"
)

// admissionLock builds root -> mid -> leaf with an unrelated overlay, so
// leaf is transitive, mid is direct through the root, and the overlay is
// direct by flag.
func admissionLock() *contextlock.Lock {
	return &contextlock.Lock{
		Root: "sysroot",
		Members: []contextlock.Member{
			{Kind: contextlock.KindContext, Name: "sysleaf", Version: "1.0.0", Commit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Source: "github.com/example/sysleaf", RequiredBy: []string{"sysmid"}},
			{Kind: contextlock.KindContext, Name: "sysmid", Version: "1.0.0", Commit: "9999999999999999999999999999999999999999", Source: "github.com/example/sysmid", RequiredBy: []string{"sysroot"}},
			{Kind: contextlock.KindContext, Name: "sysovl", Version: "1.0.0", Commit: "7777777777777777777777777777777777777777", Source: "github.com/example/sysovl", RequiredBy: []string{}, Overlay: true},
			{Kind: contextlock.KindContext, Name: "sysroot", Version: "1.0.0", Commit: "8888888888888888888888888888888888888888", Source: "github.com/example/sysroot", RequiredBy: []string{}},
		},
	}
}

func admissionPackages() map[string]Package {
	system := func(text string) Module {
		return Module{Module: contextpkg.Module{Path: "90-system.md", Class: "system"}, Bytes: []byte(text)}
	}
	return map[string]Package{
		"sysroot": {HasContext: true, Modules: []Module{system("Root system prompt.\n")}},
		"sysmid":  {HasContext: true, Modules: []Module{system("Mid system prompt.\n")}},
		"sysleaf": {HasContext: true, Modules: []Module{system("Leaf system prompt.\n")}},
		"sysovl":  {HasContext: true},
	}
}

// TestDirectSet proves the §3 partition: the root, every overlay, and
// every package the root or an overlay requires are direct; the rest is
// transitive. A package required by both the root and a transitive
// member stays direct, and a package an overlay requires is direct even
// when a transitive member requires it too.
func TestDirectSet(t *testing.T) {
	direct := DirectSet(admissionLock())
	for _, name := range []string{"sysroot", "sysmid", "sysovl"} {
		if !direct[name] {
			t.Fatalf("%s is not direct", name)
		}
	}
	if direct["sysleaf"] {
		t.Fatalf("sysleaf is direct")
	}
	overlay := &contextlock.Lock{
		Root: "sysroot",
		Members: []contextlock.Member{
			{Kind: contextlock.KindContext, Name: "sysleaf", RequiredBy: []string{"sysmid", "sysovl"}},
			{Kind: contextlock.KindContext, Name: "sysmid", RequiredBy: []string{"sysroot"}},
			{Kind: contextlock.KindContext, Name: "sysovl", RequiredBy: []string{}, Overlay: true},
			{Kind: contextlock.KindContext, Name: "sysroot", RequiredBy: []string{}},
		},
	}
	if !DirectSet(overlay)["sysleaf"] {
		t.Fatalf("a package named by an active overlay's requires is not direct")
	}
	shared := &contextlock.Lock{
		Root: "sysroot",
		Members: []contextlock.Member{
			{Kind: contextlock.KindContext, Name: "shared", RequiredBy: []string{"sysmid", "sysroot"}},
			{Kind: contextlock.KindContext, Name: "sysmid", RequiredBy: []string{"sysroot"}},
			{Kind: contextlock.KindContext, Name: "sysroot", RequiredBy: []string{}},
		},
	}
	if !DirectSet(shared)["shared"] {
		t.Fatalf("a package named by the root's requires is not direct")
	}
	if got := DirectSet(nil); len(got) != 0 {
		t.Fatalf("nil lock direct set %+v", got)
	}
}

// TestSystemPromptDropSkipsTransitive proves the drop default: the
// document holds exactly the admitted modules' bytes and dropped names
// the skipped package and module.
func TestSystemPromptDropSkipsTransitive(t *testing.T) {
	lock := admissionLock()
	document, written, dropped, err := SystemPrompt(lock, DefaultPrecedence, "claude_code", admissionPackages(), Admission{})
	if err != nil {
		t.Fatal(err)
	}
	if !written {
		t.Fatalf("written is false")
	}
	if want := "Leaf system prompt.\n"; strings.Contains(string(document), want) {
		t.Fatalf("document carries the transitive module:\n%s", document)
	}
	if want := "Mid system prompt.\n\nRoot system prompt.\n"; string(document) != want {
		t.Fatalf("document %q, want %q", document, want)
	}
	if len(dropped) != 1 || dropped[0] != (DroppedModule{Package: "sysleaf", Path: "90-system.md"}) {
		t.Fatalf("dropped %+v", dropped)
	}
}

// TestSystemPromptDropWarnsPerModule proves every dropped module is
// reported in emitted order: two transitive packages yield two dropped
// entries, dependencies first.
func TestSystemPromptDropWarnsPerModule(t *testing.T) {
	lock := admissionLock()
	packages := admissionPackages()
	packages["sysovl"] = Package{HasContext: true, Modules: []Module{
		{Module: contextpkg.Module{Path: "90-system.md", Class: "system"}, Bytes: []byte("Overlay system prompt.\n")},
	}}
	// sysovl is direct by flag, so only sysleaf drops.
	_, _, dropped, err := SystemPrompt(lock, DefaultPrecedence, "claude_code", packages, Admission{Transitive: TransitiveDrop})
	if err != nil {
		t.Fatal(err)
	}
	if len(dropped) != 1 || dropped[0].Package != "sysleaf" {
		t.Fatalf("dropped %+v, want only the transitive package", dropped)
	}
}

// TestSystemPromptErrorRefusesFirst proves the error refusal names the
// first non-admitted module in emitted order and returns no document.
func TestSystemPromptErrorRefusesFirst(t *testing.T) {
	lock := admissionLock()
	document, written, dropped, err := SystemPrompt(lock, DefaultPrecedence, "claude_code", admissionPackages(), Admission{Transitive: TransitiveError})
	if err == nil {
		t.Fatalf("error policy admitted a transitive system module")
	}
	transitive, ok := err.(*TransitiveSystemModuleError)
	if !ok {
		t.Fatalf("error %T (%v) is not the refusal", err, err)
	}
	if transitive.Package != "sysleaf" || transitive.Module != "90-system.md" {
		t.Fatalf("refusal %+v names neither the package nor the module", transitive)
	}
	if !strings.Contains(err.Error(), DiagSystemModuleTransitive) {
		t.Fatalf("error %q carries no %s", err, DiagSystemModuleTransitive)
	}
	if document != nil || written {
		t.Fatalf("refusal returned a document")
	}
	if len(dropped) != 1 || dropped[0].Package != "sysleaf" {
		t.Fatalf("dropped %+v", dropped)
	}
}

// TestSystemPromptWaiverAdmits proves a waiver admits the transitive
// package's system modules under both policies, while an entry naming no
// member changes nothing.
func TestSystemPromptWaiverAdmits(t *testing.T) {
	lock := admissionLock()
	packages := admissionPackages()
	for _, policy := range []string{TransitiveDrop, TransitiveError} {
		document, written, dropped, err := SystemPrompt(lock, DefaultPrecedence, "claude_code", packages,
			Admission{Transitive: policy, Waivers: map[string]bool{"sysleaf": true}})
		if err != nil {
			t.Fatal(err)
		}
		if !written || len(dropped) != 0 {
			t.Fatalf("policy %s: written=%v dropped=%+v", policy, written, dropped)
		}
		if want := "Leaf system prompt.\n\nMid system prompt.\n\nRoot system prompt.\n"; string(document) != want {
			t.Fatalf("policy %s: document %q, want %q", policy, document, want)
		}
	}
	_, _, dropped, err := SystemPrompt(lock, DefaultPrecedence, "claude_code", packages,
		Admission{Waivers: map[string]bool{"absent": true}})
	if err != nil {
		t.Fatal(err)
	}
	if len(dropped) != 1 {
		t.Fatalf("a waiver naming no member admitted a module: %+v", dropped)
	}
}

// TestSystemPromptSelectorExcludes proves a system module whose selector
// excludes the environment is not applicable: it is neither admitted nor
// dropped, and it never refuses under error.
func TestSystemPromptSelectorExcludes(t *testing.T) {
	lock := admissionLock()
	packages := admissionPackages()
	packages["sysleaf"] = Package{HasContext: true, Modules: []Module{
		{Module: contextpkg.Module{Path: "90-system.md", Class: "system", Environments: []string{"pi"}}, Bytes: []byte("Leaf system prompt.\n")},
	}}
	document, written, dropped, err := SystemPrompt(lock, DefaultPrecedence, "claude_code", packages, Admission{Transitive: TransitiveError})
	if err != nil {
		t.Fatal(err)
	}
	if !written || len(dropped) != 0 {
		t.Fatalf("written=%v dropped=%+v", written, dropped)
	}
	if strings.Contains(string(document), "Leaf") {
		t.Fatalf("document carries a module inapplicable to claude_code:\n%s", document)
	}
}

// TestSystemPromptAllDroppedUnwritten proves a system-prompt output with
// no admitted module is absent: written is false and the caller writes no
// file and emits no fragment section.
func TestSystemPromptAllDroppedUnwritten(t *testing.T) {
	lock := &contextlock.Lock{
		Root: "sysroot",
		Members: []contextlock.Member{
			{Kind: contextlock.KindContext, Name: "sysleaf", RequiredBy: []string{"sysmid"}},
			{Kind: contextlock.KindContext, Name: "sysmid", RequiredBy: []string{"sysroot"}},
			{Kind: contextlock.KindContext, Name: "sysroot", RequiredBy: []string{}},
		},
	}
	packages := map[string]Package{
		"sysroot": {HasContext: true},
		"sysmid":  {HasContext: true},
		"sysleaf": {HasContext: true, Modules: []Module{
			{Module: contextpkg.Module{Path: "90-system.md", Class: "system"}, Bytes: []byte("Leaf system prompt.\n")},
		}},
	}
	document, written, dropped, err := SystemPrompt(lock, DefaultPrecedence, "claude_code", packages, Admission{})
	if err != nil {
		t.Fatal(err)
	}
	if document != nil || written {
		t.Fatalf("all-dropped output is written")
	}
	if len(dropped) != 1 {
		t.Fatalf("dropped %+v", dropped)
	}
}

// TestSystemPromptRejectsPolicy proves a policy outside the closed enum
// fails closed instead of silently selecting a behavior.
func TestSystemPromptRejectsPolicy(t *testing.T) {
	_, _, _, err := SystemPrompt(admissionLock(), DefaultPrecedence, "claude_code", admissionPackages(), Admission{Transitive: "quarantine"})
	if err == nil || !strings.Contains(err.Error(), "quarantine") {
		t.Fatalf("err = %v, want the policy refusal", err)
	}
}

// TestFirstTransitiveSystemModuleOrder proves the resolution refusal
// scans emitted order with manifest order within a package, and only
// modules that apply to a registered environment.
func TestFirstTransitiveSystemModuleOrder(t *testing.T) {
	lock := admissionLock()
	modules := map[string][]contextpkg.Module{
		"sysroot": {{Path: "00-root.md"}, {Path: "90-system.md", Class: "system"}},
		"sysmid":  {{Path: "90-system.md", Class: "system"}},
		"sysleaf": {
			{Path: "00-leaf.md"},
			{Path: "80-other.md", Class: "system", Environments: []string{"unknown_env"}},
			{Path: "90-system.md", Class: "system"},
			{Path: "91-second.md", Class: "system"},
		},
	}
	first, err := FirstTransitiveSystemModule(lock, DefaultPrecedence, modules,
		Admission{Transitive: TransitiveError}, []string{"claude_code", "codex_cli", "opencode", "pi"})
	if err != nil {
		t.Fatal(err)
	}
	if first == nil || *first != (DroppedModule{Package: "sysleaf", Path: "90-system.md"}) {
		t.Fatalf("first %+v, want sysleaf/90-system.md in manifest order past the root modules and the unregistered selector", first)
	}
	// A module selecting only an unregistered environment never refuses.
	onlyUnknown := map[string][]contextpkg.Module{
		"sysleaf": {{Path: "90-system.md", Class: "system", Environments: []string{"unknown_env"}}},
	}
	first, err = FirstTransitiveSystemModule(lock, DefaultPrecedence, onlyUnknown,
		Admission{Transitive: TransitiveError}, []string{"claude_code", "codex_cli", "opencode", "pi"})
	if err != nil {
		t.Fatal(err)
	}
	if first != nil {
		t.Fatalf("first %+v, want nil for a module that selects nothing", first)
	}
	// Drop never refuses.
	first, err = FirstTransitiveSystemModule(lock, DefaultPrecedence, modules, Admission{}, []string{"claude_code"})
	if err != nil || first != nil {
		t.Fatalf("drop refused: %+v %v", first, err)
	}
	// A waiver admits.
	first, err = FirstTransitiveSystemModule(lock, DefaultPrecedence, modules,
		Admission{Transitive: TransitiveError, Waivers: map[string]bool{"sysleaf": true}}, []string{"claude_code"})
	if err != nil || first != nil {
		t.Fatalf("waived refusal: %+v %v", first, err)
	}
}
