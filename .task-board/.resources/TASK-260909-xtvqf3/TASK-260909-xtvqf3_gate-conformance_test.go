// Gate: source-derived owner-family rejection matrix for diagnostics CR2.
// Tree: fbe90d5e60593a3a069721b2ad9e53cd071d8c02 (exact candidate).
// This file is an overlay: copy to internal/diagnostics/gate_conformance_test.go
// in a disposable copy of the exact tree; do not commit into production.
// Vectors are derived at test runtime from SPEC section 6 and owner constants,
// never hand-enumerated. Each owner gate has its own named test so a
// single-foreign-code weakening fails exactly one name.
package diagnostics_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/relux-works/curator-agent-launcher/internal/axconfig"
	"github.com/relux-works/curator-agent-launcher/internal/cli"
	"github.com/relux-works/curator-agent-launcher/internal/composition"
	"github.com/relux-works/curator-agent-launcher/internal/diagnostics"
	"github.com/relux-works/curator-agent-launcher/internal/fragment"
	sp "github.com/relux-works/curator-agent-launcher/internal/systemprompt"
)

var gateSpecCodeToken = regexp.MustCompile("^[a-z][a-z0-9_]*$")

func gateSpecPath(t *testing.T) string {
	t.Helper()
	if raw, err := os.ReadFile(filepath.Join("..", "..", "SPEC.md")); err == nil && len(raw) > 0 {
		return filepath.Join("..", "..", "SPEC.md")
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller unavailable")
	}
	dir := filepath.Dir(file)
	for i := 0; i < 6; i++ {
		cand := filepath.Join(dir, "SPEC.md")
		if raw, err := os.ReadFile(cand); err == nil && len(raw) > 0 {
			return cand
		}
		dir = filepath.Dir(dir)
	}
	t.Fatal("SPEC.md not found")
	return ""
}

func gateSpecSection6Codes(t *testing.T) []string {
	t.Helper()
	raw, err := os.ReadFile(gateSpecPath(t))
	if err != nil {
		t.Fatalf("read SPEC.md: %v", err)
	}
	lines := strings.Split(string(raw), "\n")
	start := -1
	for i, l := range lines {
		if l == "## 6. Errors and diagnostics" {
			start = i
			break
		}
	}
	if start < 0 {
		t.Fatal("SPEC.md has no '## 6. Errors and diagnostics' section")
	}
	var codes []string
	inTable := false
	for _, l := range lines[start+1:] {
		if !strings.HasPrefix(l, "|") {
			if inTable {
				break
			}
			continue
		}
		inTable = true
		fields := strings.Split(l, "|")
		if len(fields) < 4 {
			t.Fatalf("malformed SPEC section 6 table row: %q", l)
		}
		if strings.Contains(fields[1], "Family") || strings.Contains(fields[1], "---") {
			continue
		}
		for _, tok := range strings.Split(fields[2], "`") {
			tok = strings.TrimSpace(tok)
			if tok == "" || !gateSpecCodeToken.MatchString(tok) {
				continue
			}
			codes = append(codes, tok)
		}
	}
	if len(codes) == 0 {
		t.Fatal("derived no codes from the SPEC section 6 table")
	}
	return codes
}

func gateResolveFamily() []string {
	return []string{
		fragment.CodeInvocationFailed,
		fragment.CodeEnvironmentUnknown,
		fragment.CodeProfileUnknown,
		fragment.CodeRepairFailed,
		fragment.CodeLockUnavailable,
		fragment.CodeFragmentInvalid,
	}
}

func gateLayerFamily() []string {
	return []string{
		composition.CodeMCPLayerMissing,
		composition.CodeMCPLayerUnreadable,
	}
}

func gateRefusalFamily() []string {
	return []string{
		sp.CodeUnavailable,
		sp.CodeUnreadable,
	}
}

func gateExtraStrangers() []string {
	return []string{
		"environment_home_stale",
		"environment_unknown",
		"invented_code",
		"",
	}
}

func gateForeign(normative, own []string) []string {
	var out []string
	for _, c := range normative {
		if !slices.Contains(own, c) {
			out = append(out, c)
		}
	}
	for _, c := range gateExtraStrangers() {
		if !slices.Contains(own, c) && !slices.Contains(out, c) {
			out = append(out, c)
		}
	}
	return out
}

func gateCheckRejects(t *testing.T, owner string, makeErr func(string) error, own []string) {
	t.Helper()
	normative := gateSpecSection6Codes(t)
	if got := diagnostics.Codes(); !slices.Equal(got, normative) {
		t.Fatalf("%s: Codes() = %q, SPEC section 6 = %q", owner, got, normative)
	}
	for _, want := range own {
		if !slices.Contains(normative, want) {
			t.Fatalf("%s: owned code %q not in SPEC section 6 table", owner, want)
		}
	}
	foreign := gateForeign(normative, own)
	if len(foreign) == 0 {
		t.Fatalf("%s: no foreign codes derived", owner)
	}
	for _, code := range foreign {
		err := makeErr(code)
		cands := map[string]error{
			"direct":  err,
			"wrapped": fmt.Errorf("outer: %w", err),
			"joined":  errors.Join(errors.New("outer"), err),
		}
		for name, cand := range cands {
			if got, ok := diagnostics.CodeOf(cand); ok || got != "" {
				t.Errorf("%s accepted foreign %q as %q (%s form)", owner, code, got, name)
			}
		}
	}
	for _, code := range own {
		if got, ok := diagnostics.CodeOf(fmt.Errorf("outer: %w", makeErr(code))); !ok || got != code {
			t.Errorf("%s own code %q: CodeOf = %q %v, want acceptance", owner, code, got, ok)
		}
	}
}

// TestGateResolveRejectsForeignCodes gates the resolve owner: every
// normative non-resolve code plus unknown/empty must yield no code in all
// three wrap forms; every resolve-family member stays accepted.
func TestGateResolveRejectsForeignCodes(t *testing.T) {
	gateCheckRejects(t, "resolve", func(c string) error {
		return &fragment.ResolveError{Code: c, Detail: "x"}
	}, gateResolveFamily())
}

// TestGateLayerRejectsForeignCodes gates the layer owner, including the
// five resolve codes the prior hand-selected matrix omitted.
func TestGateLayerRejectsForeignCodes(t *testing.T) {
	gateCheckRejects(t, "layer", func(c string) error {
		return &composition.LayerError{Code: c, Path: "p"}
	}, gateLayerFamily())
}

// TestGateRefusalRejectsForeignCodes gates the refusal owner on the same
// complete derived set.
func TestGateRefusalRejectsForeignCodes(t *testing.T) {
	gateCheckRejects(t, "refusal", func(c string) error {
		return &sp.Refusal{Code: c, Path: "p"}
	}, gateRefusalFamily())
}

// TestGateOwnFamilyAndNilPreserved pins the acceptance side the rejection
// matrix must not break: production-typed values, wrapped/joined chains,
// typed-nil fails closed without panic, and unclaimed errors invent nothing.
func TestGateOwnFamilyAndNilPreserved(t *testing.T) {
	if code, ok := diagnostics.CodeOf(&cli.UsageError{Detail: "missing <env-id>"}); !ok || code != "usage" {
		t.Fatalf("usage error: %q %v", code, ok)
	}
	re := &fragment.ResolveError{Code: fragment.CodeRepairFailed, Detail: "x"}
	if code, ok := diagnostics.CodeOf(re); !ok || code != "resolve_repair_failed" {
		t.Fatalf("resolve error: %q %v", code, ok)
	}
	missing, unreadable := codexBoundary(t)
	if code, ok := diagnostics.CodeOf(missing); !ok || code != "mcp_layer_missing" {
		t.Fatalf("mcp missing: %q %v (%v)", code, ok, missing)
	}
	if code, ok := diagnostics.CodeOf(unreadable); !ok || code != "mcp_layer_unreadable" {
		t.Fatalf("mcp unreadable: %q %v (%v)", code, ok, unreadable)
	}
	if code, ok := diagnostics.CodeOf(syspromptSelection(t)); !ok || code != "sysprompt_channel_unavailable" {
		t.Fatalf("sysprompt selection: %q %v", code, ok)
	}
	if code, ok := diagnostics.CodeOf(syspromptProbe(t)); !ok || code != "sysprompt_file_unreadable" {
		t.Fatalf("sysprompt probe: %q %v", code, ok)
	}
	if code, ok := diagnostics.CodeOf(axconfigFailure(t)); !ok || code != "defaults_config_invalid" {
		t.Fatalf("axconfig failure: %q %v", code, ok)
	}
	var ue *cli.UsageError
	var reNil *fragment.ResolveError
	var leNil *composition.LayerError
	var srNil *sp.Refusal
	var aeNil *axconfig.Error
	for name, err := range map[string]error{
		"usage": ue, "resolve": reNil, "layer": leNil, "refusal": srNil, "ax": aeNil,
		"nil": nil, "plain": errors.New("boom"),
	} {
		if got, ok := diagnostics.CodeOf(err); ok || got != "" {
			t.Errorf("%s: CodeOf = %q %v, want no code", name, got, ok)
		}
		wrapped := fmt.Errorf("outer: %w", err)
		if got, ok := diagnostics.CodeOf(wrapped); ok || got != "" {
			t.Errorf("%s wrapped: CodeOf = %q %v, want no code", name, got, ok)
		}
	}
}
