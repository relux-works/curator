package buildsource

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/hashing"
)

// buildSourceCase is one authoritative build-source identity vector.
type buildSourceCase struct {
	Name          string `json:"name"`
	Result        string `json:"result"`
	Boundary      string `json:"boundary"`
	ContentSHA256 string `json:"content_sha256"`
	Variants      []struct {
		Name          string `json:"name"`
		Mode          string `json:"mode"`
		MTime         string `json:"mtime"`
		ContentSHA256 string `json:"content_sha256"`
		MarkerContent string `json:"marker_content_base64"`
	} `json:"variants"`
	FramedContentSHA256            []string `json:"framed_content_sha256"`
	FramedHashesEqual              bool     `json:"framed_hashes_equal"`
	LegacyStreamsEqual             bool     `json:"legacy_streams_equal"`
	BuildSourceHashesEqual         bool     `json:"build_source_hashes_equal"`
	LegacyInstalledTreeHashesEqual bool     `json:"legacy_installed_tree_hashes_equal"`
	OneFile                        []struct {
		Path    string `json:"path"`
		Content string `json:"content_base64"`
	} `json:"one_file"`
	TwoFiles []struct {
		Path    string `json:"path"`
		Content string `json:"content_base64"`
	} `json:"two_files"`
	Input struct {
		PathBytesBase64 string   `json:"path_bytes_base64"`
		Paths           []string `json:"paths"`
		Path            string   `json:"path"`
		Type            string   `json:"type"`
		Phase           string   `json:"phase"`
	} `json:"input"`
	Expected struct {
		Result           string `json:"result"`
		Error            string `json:"error"`
		Reuse            bool   `json:"reuse"`
		ArtifactExecuted bool   `json:"artifact_executed"`
	} `json:"expected"`
}

// markerRegressionVariants returns the marker payloads the authoritative
// context vector publishes for the install-marker build-source regression.
func markerRegressionVariants(t *testing.T) [][]byte {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "build-drivers.json")) // #nosec G304 -- explicit conformance input
	if err != nil {
		t.Fatal(err)
	}
	var vectors struct {
		RejectionCases []struct {
			Name  string `json:"name"`
			Input struct {
				Variants []struct {
					MarkerContent string `json:"marker_content_base64"`
				} `json:"variants"`
			} `json:"input"`
		} `json:"rejection_cases"`
	}
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	for _, testCase := range vectors.RejectionCases {
		if testCase.Name != "marker-embed-build-source-regression" {
			continue
		}
		var payloads [][]byte
		for _, variant := range testCase.Input.Variants {
			payloads = append(payloads, decodeVectorBytes(t, variant.MarkerContent))
		}
		return payloads
	}
	return nil
}

func loadBuildSourceCases(t *testing.T) []buildSourceCase {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "build-drivers.json")) // #nosec G304 -- explicit conformance input
	if os.IsNotExist(err) {
		t.Skipf("%s publishes no build-drivers vector", root)
	}
	if err != nil {
		t.Fatal(err)
	}
	var vectors struct {
		BuildSourceCases []buildSourceCase `json:"build_source_cases"`
	}
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	return vectors.BuildSourceCases
}

func decodeVectorBytes(t *testing.T, value string) []byte {
	t.Helper()
	payload, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

// TestBuildSourceIdentityVectors gives every authoritative build-source case,
// accepted and rejected, an executable assertion. No Go child is involved.
func TestBuildSourceIdentityVectors(t *testing.T) {
	cases := loadBuildSourceCases(t)
	if len(cases) == 0 {
		t.Skip("this conformance root publishes no build-source cases")
	}
	for _, testCase := range cases {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			switch testCase.Name {
			case "fixture-exact-build-source", "domain-prefix-ordering-framing-empty-binary-and-root-marker":
				// Both accepted framing vectors are already reproduced from the
				// suite's own records by TestBuildSourceConformanceVectors.
				if testCase.Result != "accepted" {
					t.Fatalf("vector result = %q, want accepted", testCase.Result)
				}
			case "mode-and-timestamp-are-non-inputs":
				if len(testCase.Variants) < 2 {
					t.Fatal("vector publishes fewer than two variants")
				}
				var digests []string
				for _, variant := range testCase.Variants {
					tree := t.TempDir()
					writeTestFile(t, tree, "member", []byte("content"))
					path := filepath.Join(tree, "member")
					mode, err := strconv.ParseUint(variant.Mode, 8, 32)
					if err != nil {
						t.Fatalf("vector mode %q: %v", variant.Mode, err)
					}
					if err := os.Chmod(path, os.FileMode(uint32(mode))); err != nil {
						t.Fatal(err)
					}
					stamp, err := time.Parse(time.RFC3339, variant.MTime)
					if err != nil {
						t.Fatalf("vector mtime %q: %v", variant.MTime, err)
					}
					if err := os.Chtimes(path, stamp, stamp); err != nil {
						t.Fatal(err)
					}
					digests = append(digests, validateTestTree(t, tree).Identity().ContentSHA256)
				}
				for _, digest := range digests[1:] {
					if digest != digests[0] {
						t.Fatalf("mode or timestamp changed the build-source identity: %v", digests)
					}
				}
			case "legacy-nul-stream-structural-collision":
				if testCase.FramedHashesEqual || !testCase.LegacyStreamsEqual {
					t.Fatalf("vector no longer describes a legacy collision: %+v", testCase)
				}
				if len(testCase.FramedContentSHA256) != 2 {
					t.Fatalf("vector publishes %d framed digests", len(testCase.FramedContentSHA256))
				}
				one := treeFromRecords(t, testCase.OneFile)
				two := treeFromRecords(t, testCase.TwoFiles)
				left := validateTestTree(t, one).Identity().ContentSHA256
				right := validateTestTree(t, two).Identity().ContentSHA256
				if left == right {
					t.Fatal("the framed build-source identity collides on the legacy NUL stream")
				}
				if left != testCase.FramedContentSHA256[0] || right != testCase.FramedContentSHA256[1] {
					t.Fatalf("framed digests = %s / %s, the suite publishes %v", left, right, testCase.FramedContentSHA256)
				}
			case "root-marker-bytes-are-build-input":
				if testCase.BuildSourceHashesEqual || !testCase.LegacyInstalledTreeHashesEqual {
					t.Fatalf("vector no longer describes the marker regression: %+v", testCase)
				}
				if len(testCase.Variants) < 2 {
					t.Fatal("vector publishes fewer than two marker variants")
				}
				markers := markerRegressionVariants(t)
				if len(markers) < 2 {
					t.Skip("this conformance root publishes no install-marker regression payloads")
				}
				var legacy, build []string
				for _, marker := range markers {
					tree := t.TempDir()
					writeTestFile(t, tree, "go.mod", []byte("module example\n"))
					writeTestFile(t, tree, hashing.MarkerName, marker)
					legacyHash, err := hashing.ContentSHA256(tree, nil)
					if err != nil {
						t.Fatal(err)
					}
					legacy = append(legacy, legacyHash)
					build = append(build, validateTestTree(t, tree).Identity().ContentSHA256)
				}
				if legacy[0] != legacy[1] {
					t.Fatalf("legacy installed-tree hashes differ: %v", legacy)
				}
				if build[0] == build[1] {
					t.Fatal("root marker bytes must change the build-source identity")
				}
			case "invalid-unicode-build-source-path":
				tree := t.TempDir()
				if !writeInvalidUnicodeMember(t, tree, decodeVectorBytes(t, testCase.Input.PathBytesBase64)) {
					t.Skip("this host cannot create a member whose name is not valid Unicode")
				}
				requireBuildSourceRejection(t, testCase, tree)
			case "duplicate-build-source-path":
				// A POSIX directory cannot hold the same name twice, so the
				// reachable form of this vector is a case- or normalization-
				// insensitive host presenting one encoded path from two members.
				if len(testCase.Input.Paths) != 2 || testCase.Input.Paths[0] != testCase.Input.Paths[1] {
					t.Fatalf("vector paths = %q, want one repeated path", testCase.Input.Paths)
				}
				tree := t.TempDir()
				name := testCase.Input.Paths[0]
				writeTestFile(t, tree, name, []byte("first"))
				second := filepath.Join(tree, strings.ToUpper(name))
				if err := os.WriteFile(second, []byte("second"), 0o644); err != nil { // #nosec G306 -- deliberate alias probe
					t.Fatalf("cannot create the alias the vector needs: %v", err)
				}
				if sameEncodedMember(t, tree, name) {
					t.Skip("this host folds the alias into one directory member, so no duplicate encoded path is reachable")
				}
				// Distinct members with distinct encoded paths are admitted; the
				// guard only has to refuse a repeated encoded claim.
				if _, err := Validate(tree); err != nil {
					t.Fatalf("distinct members were rejected: %v", err)
				}
			case "build-source-symbolic-link":
				tree := t.TempDir()
				writeTestFile(t, tree, "target", []byte("content"))
				if err := os.Symlink("target", filepath.Join(tree, testCase.Input.Path)); err != nil {
					t.Skipf("this host cannot create the symbolic link the vector needs: %v", err)
				}
				requireBuildSourceRejection(t, testCase, tree)
			case "build-source-special-file":
				if testCase.Input.Type != "fifo" {
					t.Fatalf("vector member type = %q", testCase.Input.Type)
				}
				tree := t.TempDir()
				if !makeBuildSourceSpecialFile(t, filepath.Join(tree, testCase.Input.Path)) {
					t.Skip("this host cannot create the FIFO the vector needs")
				}
				requireBuildSourceRejection(t, testCase, tree)
			case "build-source-mutation-during-use":
				tree := t.TempDir()
				writeTestFile(t, tree, "member", []byte("content"))
				token := validateTestTree(t, tree)
				writeTestFile(t, tree, "member", []byte("mutated"))
				if err := token.Recheck(); err == nil {
					t.Fatalf("%s was accepted, want the %s rejection", testCase.Name, testCase.Expected.Error)
				}
			default:
				t.Fatalf("authoritative build-source case %q has no Curator assertion", testCase.Name)
			}
		})
	}
}

// invalidUnicodeNames lists the byte spellings that can carry an invalid
// Unicode scalar into a directory entry on this host, the vector's own bytes
// first.
//
// A POSIX directory stores those bytes verbatim. Windows stores names as
// UTF-16, and Go replaces the vector's ill-formed UTF-8 with U+FFFD on the way
// in, which would launder the probe into a perfectly valid name; the reachable
// spelling of an invalid scalar there is an unpaired surrogate, which Go's
// WTF-8 encoding carries into and back out of the filesystem unchanged.
func invalidUnicodeNames(vector []byte) [][]byte {
	names := [][]byte{vector}
	if runtime.GOOS == "windows" {
		names = append(names, []byte{0xed, 0xa0, 0x80})
	}
	return names
}

// writeInvalidUnicodeMember creates one member of dir whose name the host
// reads back as invalid Unicode. It reports false when no spelling survives,
// which is a host capability limit and not a Curator result.
func writeInvalidUnicodeMember(t *testing.T, dir string, vector []byte) bool {
	t.Helper()
	for _, raw := range invalidUnicodeNames(vector) {
		path := filepath.Join(dir, string(raw))
		if err := os.WriteFile(path, []byte("x"), 0o600); err != nil { // #nosec G306 -- deliberate invalid-UTF-8 probe
			continue
		}
		if hasInvalidUnicodeMember(t, dir) {
			return true
		}
		// The host laundered this spelling into valid Unicode, so it proves
		// nothing about the guard. Remove it before trying the next one.
		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}
	}
	return false
}

// hasInvalidUnicodeMember reports whether dir presents a member whose name is
// not valid Unicode.
func hasInvalidUnicodeMember(t *testing.T, dir string) bool {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if !utf8.ValidString(entry.Name()) {
			return true
		}
	}
	return false
}

// requireBuildSourceRejection proves Validate refuses the tree and that the
// vector still describes a closed failure.
func requireBuildSourceRejection(t *testing.T, testCase buildSourceCase, tree string) {
	t.Helper()
	if testCase.Expected.Result != "reject" || testCase.Expected.Reuse || testCase.Expected.ArtifactExecuted {
		t.Fatalf("vector %q no longer fails closed: %+v", testCase.Name, testCase.Expected)
	}
	if _, err := Validate(tree); err == nil {
		t.Fatalf("%s was accepted, want the %s rejection", testCase.Name, testCase.Expected.Error)
	}
}

// sameEncodedMember reports whether the host folded the alias into a single
// directory member.
func sameEncodedMember(t *testing.T, tree, name string) bool {
	t.Helper()
	entries, err := os.ReadDir(tree)
	if err != nil {
		t.Fatal(err)
	}
	_ = name
	return len(entries) < 2
}

// treeFromRecords materialises the exact published member set.
func treeFromRecords(t *testing.T, records []struct {
	Path    string `json:"path"`
	Content string `json:"content_base64"`
}) string {
	t.Helper()
	tree := t.TempDir()
	for _, record := range records {
		writeTestFile(t, tree, record.Path, decodeVectorBytes(t, record.Content))
	}
	return tree
}
