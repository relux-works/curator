package godriver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"testing"
	"time"
	"unicode/utf8"
)

// toolchainCase is one authoritative toolchain identity vector.
type toolchainCase struct {
	Name                string `json:"name"`
	Result              string `json:"result"`
	Boundary            string `json:"boundary"`
	ContentSHA256       string `json:"content_sha256"`
	NormalizedGoVersion string `json:"normalized_go_version"`
	GoVersionStdout     string `json:"go_version_stdout_base64"`
	Entries             []struct {
		Path    string `json:"path"`
		Type    string `json:"type"`
		Target  string `json:"target"`
		Content string `json:"content_base64"`
	} `json:"entries"`
	Variants []struct {
		Mode  string `json:"mode"`
		MTime string `json:"mtime"`
	} `json:"variants"`
	Input struct {
		StdoutBase64    string   `json:"stdout_base64"`
		PathBytesBase64 string   `json:"path_bytes_base64"`
		Paths           []string `json:"paths"`
		Path            string   `json:"path"`
		Target          string   `json:"target"`
		Required        string   `json:"required"`
		Selected        string   `json:"selected"`
		Phase           string   `json:"phase"`
	} `json:"input"`
	Expected struct {
		Result           string `json:"result"`
		Error            string `json:"error"`
		Reuse            bool   `json:"reuse"`
		ArtifactExecuted bool   `json:"artifact_executed"`
	} `json:"expected"`
}

// argvCase is one authoritative fixed Go argument vector.
type argvCase struct {
	Name        string   `json:"name"`
	Argv        []string `json:"argv"`
	CWD         string   `json:"cwd"`
	SourceAware bool     `json:"source_aware"`
}

type positiveVectors struct {
	FixedEnvironment      map[string]string      `json:"fixed_environment"`
	FixedEnvironmentCases []fixedEnvironmentCase `json:"fixed_environment_cases"`
	Argv                  []argvCase             `json:"argv"`
	ToolchainCases        []toolchainCase        `json:"toolchain_cases"`
	PositiveCases         []struct {
		Name    string `json:"name"`
		Result  string `json:"result"`
		Package struct {
			Name              string   `json:"name"`
			MainPackages      int      `json:"main_packages"`
			Dependencies      []string `json:"dependencies"`
			VendorRequired    bool     `json:"vendor_required"`
			VendorMode        string   `json:"vendor_mode"`
			ModuleRoot        string   `json:"module_root"`
			TransitivePackage string   `json:"transitive_package"`
			VendoredPackage   string   `json:"vendored_package"`
			EmbeddedInputs    []string `json:"embedded_inputs"`
		} `json:"package"`
	} `json:"positive_cases"`
}

type fixedEnvironmentCase struct {
	Name              string            `json:"name"`
	GOOS              string            `json:"goos"`
	GOARCH            string            `json:"goarch"`
	Environment       map[string]string `json:"environment"`
	OptionalVariables []string          `json:"optional_variables"`
}

func loadPositiveVectors(t *testing.T) positiveVectors {
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
	var vectors positiveVectors
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	return vectors
}

func decodeBase64(t *testing.T, value string) []byte {
	t.Helper()
	payload, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

// TestFixedEnvironmentAndFiveDirectArgvFormsVector proves Curator's five fixed
// Go argument vectors and its complete closed environment are exactly the
// published ones, including which forms are source-aware.
func TestFixedEnvironmentAndFiveDirectArgvFormsVector(t *testing.T) {
	vectors := loadPositiveVectors(t)
	if len(vectors.Argv) == 0 {
		t.Skip("this conformance root publishes no fixed argument vectors")
	}
	if len(vectors.Argv) != 5 {
		t.Fatalf("suite publishes %d argument-vector forms, want the exact five", len(vectors.Argv))
	}

	// Curator's own inventory of package-independent probe forms plus the two
	// source-aware forms.
	probes := map[string][]string{
		"telemetry-off": {"telemetry", "off"},
		"version":       {"version"},
		"env":           append([]string{"env", "-json"}, probeEnvNames...),
	}
	for _, testCase := range vectors.Argv {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			if len(testCase.Argv) == 0 {
				t.Fatal("vector publishes an empty argument vector")
			}
			arguments := testCase.Argv[1:]
			switch testCase.Name {
			case "list":
				if !testCase.SourceAware {
					t.Fatal("go list must be source-aware")
				}
				if !reflect.DeepEqual(arguments, listArguments) {
					t.Fatalf("list argv = %q, Curator uses %q", arguments, listArguments)
				}
			case "build":
				if !testCase.SourceAware {
					t.Fatal("go build must be source-aware")
				}
				if len(arguments) != len(buildArgumentPrefix)+2 {
					t.Fatalf("build argv = %q", arguments)
				}
				if !reflect.DeepEqual(arguments[:len(buildArgumentPrefix)], buildArgumentPrefix) {
					t.Fatalf("build prefix = %q, Curator uses %q", arguments[:len(buildArgumentPrefix)], buildArgumentPrefix)
				}
				if arguments[len(arguments)-1] != "." {
					t.Fatalf("build argv does not end at the validated source directory: %q", arguments)
				}
			default:
				if testCase.SourceAware {
					t.Fatalf("%s must stay package-independent", testCase.Name)
				}
				want, ok := probes[testCase.Name]
				if !ok {
					t.Fatalf("Curator has no package-independent form named %q", testCase.Name)
				}
				if !reflect.DeepEqual(arguments, want) {
					t.Fatalf("%s argv = %q, Curator uses %q", testCase.Name, arguments, want)
				}
			}
		})
	}

	t.Run("fixed environment", func(t *testing.T) {
		expected, optional := fixedEnvironmentForHost(t, vectors, runtime.GOOS, runtime.GOARCH)
		if len(expected) == 0 {
			t.Skip("this conformance root publishes no fixed environment")
		}
		fixture := newSnapshotFixture(t)
		fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact"})
		values := environmentMap(fixture.session.Environment())

		published := make([]string, 0, len(expected))
		for key := range expected {
			published = append(published, key)
		}
		sort.Strings(published)
		for _, key := range published {
			want := expected[key]
			got, present := values[key]
			if !present {
				t.Fatalf("Curator's closed environment omits %s", key)
			}
			// Manager-derived locations are host-specific by construction; the
			// suite marks them with a placeholder rather than a literal value.
			if want == "" || want[0] == '<' {
				continue
			}
			if got != want {
				t.Fatalf("closed environment %s = %q, the suite publishes %q", key, got, want)
			}
		}
		for key := range values {
			if _, ok := expected[key]; !ok && !optional[key] {
				t.Fatalf("Curator's closed environment carries %s, which the suite does not publish", key)
			}
		}
	})
}

func fixedEnvironmentForHost(t *testing.T, vectors positiveVectors, goos, goarch string) (map[string]string, map[string]bool) {
	t.Helper()
	if len(vectors.FixedEnvironmentCases) == 0 {
		return vectors.FixedEnvironment, nil
	}
	for _, testCase := range vectors.FixedEnvironmentCases {
		if testCase.GOOS != goos || testCase.GOARCH != goarch {
			continue
		}
		optional := make(map[string]bool, len(testCase.OptionalVariables))
		for _, key := range testCase.OptionalVariables {
			optional[key] = true
		}
		return testCase.Environment, optional
	}
	t.Fatalf("the suite publishes no fixed environment for native host %s/%s", goos, goarch)
	return nil, nil
}

func TestFixedEnvironmentForHostSelectsNativeCase(t *testing.T) {
	vectors := positiveVectors{
		FixedEnvironment: map[string]string{"GOARCH": "arm64"},
		FixedEnvironmentCases: []fixedEnvironmentCase{
			{GOOS: "darwin", GOARCH: "arm64", Environment: map[string]string{"GOARCH": "arm64"}},
			{GOOS: "windows", GOARCH: "amd64", Environment: map[string]string{"GOARCH": "amd64"}, OptionalVariables: []string{"SYSTEMROOT", "WINDIR"}},
		},
	}
	environment, optional := fixedEnvironmentForHost(t, vectors, "windows", "amd64")
	if environment["GOARCH"] != "amd64" {
		t.Fatalf("selected GOARCH = %q", environment["GOARCH"])
	}
	if !optional["SYSTEMROOT"] || !optional["WINDIR"] {
		t.Fatalf("optional variables = %#v", optional)
	}
}

// TestToolchainIdentityVectors proves every authoritative toolchain identity
// case, positive and negative, is executable against Curator's fingerprint.
func TestToolchainIdentityVectors(t *testing.T) {
	vectors := loadPositiveVectors(t)
	if len(vectors.ToolchainCases) == 0 {
		t.Skip("this conformance root publishes no toolchain cases")
	}
	for _, testCase := range vectors.ToolchainCases {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			switch testCase.Name {
			case "unsorted-directories-files-and-internal-link":
				root := materializeToolchain(t, testCase)
				digest, _, err := fingerprintToolchain(context.Background(), root, testCase.NormalizedGoVersion)
				if err != nil {
					t.Fatal(err)
				}
				if digest != testCase.ContentSHA256 {
					t.Fatalf("digest = %s, the suite publishes %s", digest, testCase.ContentSHA256)
				}
			case "crlf-version-normalizes-to-lf-identity":
				normalized, _, _, _, err := parseGoVersion(decodeBase64(t, testCase.GoVersionStdout))
				if err != nil {
					t.Fatal(err)
				}
				if normalized != testCase.NormalizedGoVersion {
					t.Fatalf("normalized version = %q, the suite publishes %q", normalized, testCase.NormalizedGoVersion)
				}
			case "toolchain-mode-and-timestamp-are-non-inputs":
				if len(testCase.Variants) < 2 {
					t.Fatal("vector publishes fewer than two variants")
				}
				var digests []string
				for _, variant := range testCase.Variants {
					root := t.TempDir()
					writeTestFile(t, filepath.Join(root, "bin", "go"), []byte("GO"), toolchainVariantMode(t, variant.Mode))
					stamp := toolchainVariantTime(t, variant.MTime)
					if err := os.Chtimes(filepath.Join(root, "bin", "go"), stamp, stamp); err != nil {
						t.Fatal(err)
					}
					digest, _, err := fingerprintToolchain(context.Background(), root, "go version go1.25.5 darwin/arm64")
					if err != nil {
						t.Fatal(err)
					}
					digests = append(digests, digest)
				}
				for _, digest := range digests[1:] {
					if digest != digests[0] {
						t.Fatalf("mode and timestamp changed the toolchain identity: %v", digests)
					}
				}
			case "toolchain-version-missing-terminal-lf", "toolchain-version-multiple-terminal-newlines":
				_, _, _, _, err := parseGoVersion(decodeBase64(t, testCase.Input.StdoutBase64))
				requireDiagnostic(t, testCase, err)
			case "invalid-unicode-toolchain-path":
				root := t.TempDir()
				writeTestFile(t, filepath.Join(root, "bin", "go"), []byte("GO"), 0o755)
				if !writeInvalidUnicodeMember(t, filepath.Join(root, "bin"), decodeBase64(t, testCase.Input.PathBytesBase64)) {
					t.Skip("this host cannot create a member whose name is not valid Unicode")
				}
				_, _, err := fingerprintToolchain(context.Background(), root, "go version go1.25.5 darwin/arm64")
				requireDiagnostic(t, testCase, err)
			case "duplicate-toolchain-path":
				if len(testCase.Input.Paths) != 2 || testCase.Input.Paths[0] != testCase.Input.Paths[1] {
					t.Fatalf("vector paths = %q, want one repeated path", testCase.Input.Paths)
				}
				seen := map[string]struct{}{testCase.Input.Paths[0]: {}}
				requireDiagnostic(t, testCase, claimEncodedPath(seen, testCase.Input.Paths[1]))
			case "escaping-toolchain-link", "absolute-toolchain-link", "dangling-toolchain-link":
				root := t.TempDir()
				if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(filepath.Dir(testCase.Input.Path))), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.FromSlash(testCase.Input.Target), filepath.Join(root, filepath.FromSlash(testCase.Input.Path))); err != nil {
					t.Skipf("this host cannot create the symbolic link the vector needs: %v", err)
				}
				_, _, err := fingerprintToolchain(context.Background(), root, "go version go1.25.5 darwin/arm64")
				requireDiagnostic(t, testCase, err)
			case "selected-go-outside-goroot":
				outside := testToolchain(t)
				inside := t.TempDir()
				if err := os.MkdirAll(filepath.Join(inside, "bin"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(outside, "bin", platformGoName), filepath.Join(inside, "bin", platformGoName)); err != nil {
					t.Skipf("this host cannot create the symbolic link the vector needs: %v", err)
				}
				_, _, _, err := selectToolchain(Config{CuratorGo: filepath.Join(inside, "bin", platformGoName)}, outside)
				requireDiagnostic(t, testCase, err)
			case "toolchain-tree-mutation-during-use":
				fixture := newSnapshotFixture(t)
				fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact"})
				writeTestFile(t, filepath.Join(fixture.goroot, "VERSION"), []byte("go1.25.5 mutated\n"), 0o644)
				requireDiagnostic(t, testCase, fixture.session.VerifyToolchain(context.Background()))
			default:
				t.Fatalf("authoritative toolchain case %q has no Curator assertion", testCase.Name)
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

func requireDiagnostic(t *testing.T, testCase toolchainCase, err error) {
	t.Helper()
	if testCase.Expected.Result != "reject" || testCase.Expected.Reuse || testCase.Expected.ArtifactExecuted {
		t.Fatalf("vector %q no longer fails closed: %+v", testCase.Name, testCase.Expected)
	}
	if err == nil {
		t.Fatalf("%s was accepted, want the %s rejection", testCase.Name, testCase.Expected.Error)
	}
	if got := DiagnosticCode(err); got != testCase.Expected.Error {
		t.Fatalf("%s Curator code = %q (%v), the suite publishes %q", testCase.Name, got, err, testCase.Expected.Error)
	}
}

// materializeToolchain writes the exact published toolchain tree.
func materializeToolchain(t *testing.T, testCase toolchainCase) string {
	t.Helper()
	root := t.TempDir()
	for _, entry := range testCase.Entries {
		target := filepath.Join(root, filepath.FromSlash(entry.Path))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatal(err)
		}
		switch entry.Type {
		case "directory":
			if err := os.MkdirAll(target, 0o755); err != nil {
				t.Fatal(err)
			}
		case "symlink":
			// The link payload is a protocol byte input. filepath.FromSlash would
			// rewrite its separators on Windows and silently change the digest.
			if err := os.Symlink(entry.Target, target); err != nil {
				t.Skipf("this host cannot create the symbolic link the vector needs: %v", err)
			}
			if observed, err := os.Readlink(target); err != nil || observed != entry.Target {
				t.Fatalf("materialized link target = %q, %v; want exact protocol target %q", observed, err, entry.Target)
			}
		default:
			writeTestFile(t, target, decodeBase64(t, entry.Content), 0o755)
		}
	}
	return root
}

func toolchainVariantMode(t *testing.T, value string) os.FileMode {
	t.Helper()
	mode, err := strconv.ParseUint(value, 8, 32)
	if err != nil {
		t.Fatalf("vector mode %q: %v", value, err)
	}
	return os.FileMode(uint32(mode))
}

func toolchainVariantTime(t *testing.T, value string) time.Time {
	t.Helper()
	stamp, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("vector mtime %q: %v", value, err)
	}
	return stamp
}

// TestValidPackageGraphVectors proves both authoritative accepted package
// shapes pass Curator's complete graph validation. The validation is the exact
// seam the vectors describe and runs no Go child at all, so an accepted shape
// is proved without compiling or starting any package code.
func TestValidPackageGraphVectors(t *testing.T) {
	vectors := loadPositiveVectors(t)
	for _, testCase := range vectors.PositiveCases {
		if testCase.Name != "valid-standard-library-only-main" && testCase.Name != "valid-vendor-only-main-with-transitive-embed" {
			continue
		}
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			if testCase.Result != "accepted" {
				t.Fatalf("vector result = %q, want accepted", testCase.Result)
			}
			if testCase.Package.MainPackages != 1 {
				t.Fatalf("vector declares %d main packages", testCase.Package.MainPackages)
			}
			fixture := newSnapshotFixture(t)
			goroot := mustPhysical(t, t.TempDir())
			validation := graphValidation{BuildRoot: fixture.buildRoot, SourceDir: fixture.sourceDir, GOROOT: goroot}
			root := fixture.rootPackage()
			items := []packageJSON{root}

			if testCase.Name == "valid-standard-library-only-main" {
				if len(testCase.Package.Dependencies) != 1 || testCase.Package.Dependencies[0] != "standard-library" {
					t.Fatalf("vector dependencies = %q", testCase.Package.Dependencies)
				}
				if testCase.Package.VendorRequired {
					t.Fatal("the standard-library shape must not require a vendor tree")
				}
				standardDir := filepath.Join(goroot, "src", "fmt")
				writeTestFile(t, filepath.Join(standardDir, "print.go"), []byte("package fmt\n"), 0o644)
				items = append(items, packageJSON{
					Dir: standardDir, Root: goroot, ImportPath: "fmt", Name: "fmt",
					Standard: true, Goroot: true, DepOnly: true, GoFiles: []string{"print.go"},
				})
			} else {
				if testCase.Package.VendorMode != "vendor" || testCase.Package.VendoredPackage == "" ||
					testCase.Package.TransitivePackage == "" || len(testCase.Package.EmbeddedInputs) == 0 {
					t.Fatalf("vector package shape = %+v", testCase.Package)
				}
				for _, embedded := range testCase.Package.EmbeddedInputs {
					name := filepath.Base(filepath.FromSlash(embedded))
					writeTestFile(t, filepath.Join(fixture.sourceDir, name), []byte(name), 0o644)
					root.EmbedFiles = append(root.EmbedFiles, name)
				}
				items = []packageJSON{root}
				items = append(items, transitivePackage(fixture, func(*packageJSON) {}))
				items = append(items, vendoredPackage(t, fixture, testCase.Package.VendoredPackage))
			}

			if err := validatePackageGraph(encodePackages(t, items...), validation); err != nil {
				t.Fatalf("accepted package shape was rejected: %v", err)
			}
		})
	}
}

// vendoredPackage returns a checked-in vendored dependency of the root package,
// materialised at the import path the suite publishes.
func vendoredPackage(t *testing.T, fixture *workerFixture, importPath string) packageJSON {
	t.Helper()
	vendorRoot := filepath.Join(fixture.buildRoot, "vendor")
	directory := filepath.Join(vendorRoot, filepath.FromSlash(importPath))
	writeTestFile(t, filepath.Join(directory, "decorate.go"), []byte("package decorate\n"), 0o644)
	modulePath := filepath.ToSlash(filepath.Dir(filepath.FromSlash(importPath)))
	writeTestFile(t, filepath.Join(vendorRoot, "modules.txt"),
		[]byte("# "+modulePath+" v1.0.0\n## explicit; go 1.23\n"+importPath+"\n"), 0o644)
	return packageJSON{
		Dir:        directory,
		ImportPath: importPath,
		Name:       filepath.Base(filepath.FromSlash(importPath)), DepOnly: true,
		GoFiles: []string{"decorate.go"},
		Module: &moduleJSON{
			Path: modulePath, Version: "v1.0.0", Dir: directory, GoVersion: "1.23",
		},
	}
}
