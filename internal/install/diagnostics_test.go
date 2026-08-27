package install

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/godriver"
)

func TestRedactDiagnosticReplacesEveryAbsoluteLocation(t *testing.T) {
	t.Parallel()
	for name, testCase := range map[string]struct{ detail, want string }{
		"unix path": {
			detail: "inspect cache boundary: lstat /Users/operator/.curator/cache/build/go-v1/abcd: denied",
			want:   "inspect cache boundary: lstat <path>: denied",
		},
		"windows path": {
			detail: `open C:\Users\operator\AppData\curator: access denied`,
			want:   "open <path>: access denied",
		},
		"unc path": {
			detail: `entry \\fileserver\share\curator is not protected`,
			want:   "entry <path> is not protected",
		},
		"quoted path": {
			detail: `staged output "/private/var/folders/x/staging" has multiple links`,
			want:   `staged output "<path>" has multiple links`,
		},
		"relative declaration paths survive": {
			detail: `source_dir "assets/build-tool/cmd/tool" is not below build root assets/build-tool`,
			want:   `source_dir "assets/build-tool/cmd/tool" is not below build root assets/build-tool`,
		},
		"a bare separator is not a path": {
			detail: "artifact path / is not a direct bin child",
			want:   "artifact path / is not a direct bin child",
		},
		"embedded after a key": {
			detail: "source=/private/var/folders/x/staging is unreadable",
			want:   "source=<path> is unreadable",
		},
		"embedded windows path after a key": {
			detail: `error=C:\Users\operator\cache could not be opened`,
			want:   "error=<path> could not be opened",
		},
		"embedded unc path after a key": {
			detail: `entry=\\fileserver\share\curator is not protected`,
			want:   "entry=<path> is not protected",
		},
		"uri form": {
			detail: "receipt at file:///private/var/folders/x/staging failed",
			want:   "receipt at file:<path> failed",
		},
		"several embedded locations in one detail": {
			detail: `moved=/private/a/b to=C:\Users\operator\b and kept=\\host\share\c`,
			want:   "moved=<path> to=<path> and kept=<path>",
		},
		"a bracketed path": {
			detail: "staging (/private/var/folders/x/staging) is not empty",
			want:   "staging (<path>) is not empty",
		},
		"a trailing sentence stop is not part of the path": {
			detail: "could not remove /private/var/folders/x/staging.",
			want:   "could not remove <path>.",
		},
		"protocol-relative values embedded after a key survive": {
			detail: "root=assets/build-tool dir=assets/build-tool/cmd/tool target=linux/amd64",
			want:   "root=assets/build-tool dir=assets/build-tool/cmd/tool target=linux/amd64",
		},
	} {
		t.Run(name, func(t *testing.T) {
			if got := RedactDiagnostic(testCase.detail); got != testCase.want {
				t.Fatalf("RedactDiagnostic(%q) = %q, want %q", testCase.detail, got, testCase.want)
			}
		})
	}
}

// TestRedactDiagnosticCannotDriveATerminal proves untrusted compiler, receipt,
// and cache bytes cannot forge a second report line or emit an escape sequence.
func TestRedactDiagnosticCannotDriveATerminal(t *testing.T) {
	t.Parallel()
	detail := "artifact hash mismatch\r\n\x1b[2Jerror: everything is fine\x00\u200bnot really"
	got := RedactDiagnostic(detail)
	for _, forbidden := range []string{"\n", "\r", "\x1b", "\x00", "\u200b"} {
		if strings.ContainsAny(got, forbidden) {
			t.Fatalf("redacted detail %q still carries %q", got, forbidden)
		}
	}
	if !strings.HasPrefix(got, "artifact hash mismatch ") {
		t.Fatalf("redacted detail lost its leading diagnostic: %q", got)
	}
}

func TestRedactDiagnosticIsBounded(t *testing.T) {
	t.Parallel()
	got := RedactDiagnostic(strings.Repeat("unbounded compiler output ", 400))
	if count := utf8.RuneCountInString(got); count != maxDiagnosticRunes {
		t.Fatalf("bounded detail is %d runes, want exactly %d", count, maxDiagnosticRunes)
	}
	if !strings.HasSuffix(got, truncationMarker) {
		t.Fatalf("a truncated detail does not say so: %q", got)
	}
	short := "artifact size mismatch"
	if RedactDiagnostic(short) != short {
		t.Fatalf("a short detail was altered: %q", RedactDiagnostic(short))
	}
}

func TestRedactDiagnosticKeepsInvalidUTF8Out(t *testing.T) {
	t.Parallel()
	got := RedactDiagnostic("receipt \xff\xfe bytes")
	if !utf8.ValidString(got) {
		t.Fatalf("redacted detail is not valid UTF-8: %q", got)
	}
	if got != "receipt bytes" {
		t.Fatalf("RedactDiagnostic dropped more or less than the invalid bytes: %q", got)
	}
}

// TestPlanLinesRedactAnUntrustedReason proves the installation report itself,
// not only the status surface, publishes bounded and redacted details.
func TestPlanLinesRedactAnUntrustedReason(t *testing.T) {
	t.Parallel()
	build := PlannedBuild{
		skill: "build-skill", command: "tool", outcome: BuildWouldRebuildUntrustedCache,
		reason: "inspect cache boundary: lstat /Users/operator/.curator/cache/build/go-v1/abcd: denied",
	}
	line := build.Describe()
	if strings.Contains(line, "/Users/") {
		t.Fatalf("plan line published a manager-private location: %s", line)
	}
	if !strings.Contains(line, "<path>") {
		t.Fatalf("plan line did not redact the location: %s", line)
	}
}

// TestBuildFailuresAreRedactedInTheResult proves the other half of the
// installation surface: the error text of a failed build phase is rendered
// through the same bounded, path-redacted rendering as a report line, so a
// blocking plan cannot publish a manager-private location through Result.Errors.
func TestBuildFailuresAreRedactedInTheResult(t *testing.T) {
	t.Parallel()
	var result Result
	result.failBuild(fmt.Errorf(
		"build cache refused reuse: build-skill.tool is unsupported: " +
			"inspect cache boundary: lstat /Users/operator/.curator/cache/build/go-v1/abcd: denied"))
	if len(result.Errors) != 1 || result.Status != "failed" {
		t.Fatalf("result = %+v", result)
	}
	if strings.Contains(result.Errors[0], "/Users/") || !strings.Contains(result.Errors[0], "<path>") {
		t.Fatalf("build failure published a manager-private location: %q", result.Errors[0])
	}

	var bounded Result
	bounded.failBuild(fmt.Errorf("%s", strings.Repeat("unbounded compiler output ", 400)))
	if utf8.RuneCountInString(bounded.Errors[0]) != maxDiagnosticRunes {
		t.Fatalf("build failure is %d runes, want exactly %d",
			utf8.RuneCountInString(bounded.Errors[0]), maxDiagnosticRunes)
	}
}

// TestGoToolchainRemedyReachesTheOperatorIntact proves the half of the
// version-manager remedy this package owns: the go-v1 boundary renders it
// behind the protocol detail of a toolchain_executable_mismatch, and the
// failure surface an operator reads goes through RedactDiagnostic, whose job is
// to remove locations. The remedy names a shell assignment rather than a
// location, so it must survive that rendering verbatim — a remedy that reaches
// the operator as `<path>` or as a truncated fragment is not a remedy.
//
// The diagnostic is taken from the driver rather than written out here, so the
// case fails if either side of the boundary changes without the other.
func TestGoToolchainRemedyReachesTheOperatorIntact(t *testing.T) {
	t.Parallel()
	const remedy = `; put the real GOROOT/bin first on PATH, e.g. PATH="$(go env GOROOT)/bin:$PATH"`

	launcher := "go"
	if runtime.GOOS == "windows" {
		launcher = "go.exe"
	}
	// The shape of a version-manager selection: the launcher named by the
	// trusted selection variable resolves to a real launcher under a different
	// root, so the boundary refuses it and says what to do about it.
	wrapperRoot, realRoot := t.TempDir(), t.TempDir()
	for _, dir := range []string{wrapperRoot, realRoot} {
		if err := os.MkdirAll(filepath.Join(dir, "bin"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(realRoot, "bin", launcher), []byte("binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(realRoot, "bin", launcher), filepath.Join(wrapperRoot, "bin", launcher)); err != nil {
		t.Skipf("this host cannot create the symbolic link the case needs: %v", err)
	}

	_, err := godriver.Probe(context.Background(), godriver.Config{
		PrivateBase: t.TempDir(),
		CuratorGo:   filepath.Join(wrapperRoot, "bin", launcher),
	})
	if code := godriver.DiagnosticCode(err); code != "toolchain_executable_mismatch" {
		t.Fatalf("error = %v, code = %q, want %q", err, code, "toolchain_executable_mismatch")
	}
	if !strings.Contains(err.Error(), remedy) {
		t.Fatalf("driver diagnostic carries no operator remedy: %q", err)
	}

	var result Result
	result.failBuild(err)
	if len(result.Errors) != 1 {
		t.Fatalf("result = %+v", result)
	}
	if !strings.Contains(result.Errors[0], remedy) {
		t.Fatalf("rendered failure lost the operator remedy: %q", result.Errors[0])
	}
}
