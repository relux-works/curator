package contextaudit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Production entry points under test: Detect, DetectFiles, InScope, PinKey,
// SystemModules, Report.Blocking.
//
// Note: example credentials below are assembled at runtime ("AK" + ...)
// so no secret-shaped literal appears in the source; the detector still
// sees the joined bytes.

const testPin = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

// exampleKey is a key-shaped token the detector must flag: after the class
// prefix the remainder is neither one repeated character nor EXAMPLE-suffixed,
// so the placeholder rule does not apply.
func exampleKey() string { return "AK" + "IA1234567890ABCDEF" }

// placeholderKey is a placeholder-shaped body the detector must skip: the
// remainder after the class prefix is one repeated character.
func placeholderKey() string { return "AK" + "IA" + strings.Repeat("X", 16) }

// TestSecretIsBlocking checks the happy-path gate: an AWS key id in a
// context module blocks the install.
func TestSecretIsBlocking(t *testing.T) {
	report := DetectFiles(map[string][]byte{
		"context/notes.md": []byte("key " + exampleKey() + " here\n"),
	}, testPin, nil)
	if !report.Blocking() {
		t.Fatalf("findings %+v must block", report.Findings)
	}
	if len(report.Findings) != 1 || report.Findings[0].Class != ClassSecretMaterial {
		t.Fatalf("findings %+v", report.Findings)
	}
}

// TestPlaceholderDoesNotBlock narrows the placeholder rule: an example key
// with the documented placeholder shape must not block. A mutant that
// reports every key-shaped token must fail this test.
func TestPlaceholderDoesNotBlock(t *testing.T) {
	report := DetectFiles(map[string][]byte{
		"context/notes.md": []byte("example " + placeholderKey() + " not real\n"),
	}, testPin, nil)
	if report.Blocking() {
		t.Fatalf("placeholder findings %+v must not block", report.Findings)
	}
}

// TestOutOfScopeIsIgnored checks files outside the scope never report, even
// with secret-shaped content.
func TestOutOfScopeIsIgnored(t *testing.T) {
	report := DetectFiles(map[string][]byte{
		"notes.md":         []byte("key " + exampleKey() + " here\n"),
		"README.md":        []byte("-----BEGIN RSA PRIVATE KEY-----\n"),
		"context/notes.md": []byte("clean\n"),
	}, testPin, nil)
	if report.Blocking() || len(report.Findings) != 0 {
		t.Fatalf("out-of-scope findings %+v", report.Findings)
	}
	if InScope("notes.md") || !InScope("context/deep/file.md") || !InScope("agent-mcp.json") {
		t.Fatal("scope classification is wrong")
	}
}

// TestMCPArgsAndURLAreInScope checks the section 2.2 scope: secrets inside
// agent-mcp.json (args, url) are reported.
func TestMCPArgsAndURLAreInScope(t *testing.T) {
	report := DetectFiles(map[string][]byte{
		"agent-mcp.json": []byte(`{"server": {"args": ["--token", "Bearer abcdefghijklmnopqrst"]}}` + "\n"),
	}, testPin, nil)
	if !report.Blocking() {
		t.Fatalf("mcp findings %+v must block", report.Findings)
	}
}

// TestScopedWaiverClearsOnlyItsSpan narrows the waiver gate: a waiver at the
// member pin clears exactly its file and span; the same span at another pin
// still blocks. A mutant that clears by file alone must fail this test.
func TestScopedWaiverClearsOnlyItsSpan(t *testing.T) {
	files := map[string][]byte{
		"context/a.md": []byte("key " + exampleKey() + " here\n"),
	}
	finding := DetectFiles(files, testPin, nil)
	if len(finding.Findings) != 1 {
		t.Fatalf("findings %+v", finding.Findings)
	}
	span := finding.Findings[0].Span
	waived := DetectFiles(files, testPin, []Waiver{{Pin: testPin, File: "context/a.md", Span: span, Reason: "rotated"}})
	if waived.Blocking() {
		t.Fatalf("waived findings %+v must not block", waived.Findings)
	}
	otherPin := DetectFiles(files, strings.Repeat("f", 64), []Waiver{{Pin: testPin, File: "context/a.md", Span: span, Reason: "rotated"}})
	if !otherPin.Blocking() {
		t.Fatal("a waiver at another pin must not apply")
	}
	// A content-hash pin spelling of the same pin applies: only the pin
	// identity matters, never the spelling.
	hashed := DetectFiles(files, testPin, []Waiver{{Pin: "state sha256:" + testPin, File: "context/a.md", Span: span, Reason: "rotated"}})
	if hashed.Blocking() {
		t.Fatal("an equivalent pin spelling must apply")
	}
	if PinKey("commit "+testPin) != testPin || PinKey("state sha256:"+testPin) != testPin {
		t.Fatal("PinKey normalization is wrong")
	}
}

// TestUnmatchedWaiverIsReported checks waivers that clear nothing are
// surfaced as unmatched rather than silently accepted.
func TestUnmatchedWaiverIsReported(t *testing.T) {
	report := DetectFiles(map[string][]byte{
		"context/a.md": []byte("clean\n"),
	}, testPin, []Waiver{{Pin: testPin, File: "context/a.md", Span: [2]int{0, 5}, Reason: "nothing there"}})
	if report.Blocking() {
		t.Fatal("no findings must not block")
	}
	if len(report.Waivers) != 1 || report.Waivers[0].Diagnostic != DiagWaiverUnmatched {
		t.Fatalf("waivers %+v", report.Waivers)
	}
}

// TestDetectorIsUnpinnable documents the unpinnable bound: Detect takes no
// allowlist or pin exemption — only scoped waivers clear findings, and a
// content-hash pin never does. There is no flag that disables the detector.
func TestDetectorIsUnpinnable(t *testing.T) {
	files := map[string][]byte{"context/a.md": []byte("key " + exampleKey() + "\n")}
	if !DetectFiles(files, testPin, nil).Blocking() {
		t.Fatal("detector must block without waivers")
	}
}

func TestOpaqueNULBlocksOutsideSecretScopeAndIgnoresWaivers(t *testing.T) {
	files := map[string][]byte{"assets/deep/image.bin": []byte("prefix\x00suffix")}
	report := DetectFiles(files, testPin, []Waiver{{
		Pin: testPin, File: "assets/deep/image.bin", Span: [2]int{6, 7}, Reason: "must not waive opaque content",
	}})
	if !report.Blocking() {
		t.Fatalf("opaque findings %+v must block", report.Findings)
	}
	if len(report.Findings) != 1 {
		t.Fatalf("opaque findings %+v, want one file finding", report.Findings)
	}
	finding := report.Findings[0]
	if finding.Class != ClassOpaqueFile || finding.Pattern != PatternNULByte || finding.Severity != SeverityBlocking || finding.File != "assets/deep/image.bin" || finding.Waived {
		t.Fatalf("opaque finding %+v", finding)
	}
	if len(report.Waivers) != 1 || report.Waivers[0].Diagnostic != DiagWaiverUnmatched {
		t.Fatalf("opaque waiver result %+v, want unmatched waiver", report.Waivers)
	}
}

func TestDetectScansDeepRegularFiles(t *testing.T) {
	root := t.TempDir()
	regular := filepath.Join(root, "assets", "deep", "image.bin")
	if err := os.MkdirAll(filepath.Dir(regular), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(regular, []byte{1, 0, 2}, 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := Detect(root, testPin, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Blocking() || len(report.Findings) != 1 || report.Findings[0].File != "assets/deep/image.bin" {
		t.Fatalf("findings %+v, want only the deep regular file", report.Findings)
	}
}

func TestDetectDoesNotFollowSymlinks(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.bin")
	if err := os.WriteFile(outside, []byte{1, 0, 2}, 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "docs", "linked.bin")
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	report, err := Detect(root, testPin, nil)
	if err != nil {
		t.Fatal(err)
	}
	if report.Blocking() || len(report.Findings) != 0 {
		t.Fatalf("findings %+v, want no findings from a symlink target", report.Findings)
	}
}
