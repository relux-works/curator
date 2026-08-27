package godriver

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
	"time"
)

// driverRejectionVector is the authoritative rejection record. Only the fields
// the driver has to honour are decoded; every portable expected value stays in
// the suite.
type driverRejectionVector struct {
	Name     string `json:"name"`
	Boundary string `json:"boundary"`
	Expected struct {
		Result           string `json:"result"`
		Error            string `json:"error"`
		Reuse            bool   `json:"reuse"`
		ArtifactExecuted bool   `json:"artifact_executed"`
	} `json:"expected"`
}

func loadDriverRejections(t *testing.T) map[string]driverRejectionVector {
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
		RejectionCases []driverRejectionVector `json:"rejection_cases"`
	}
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	indexed := make(map[string]driverRejectionVector, len(vectors.RejectionCases))
	for _, testCase := range vectors.RejectionCases {
		indexed[testCase.Name] = testCase
	}
	return indexed
}

// driverRejection is the Curator half of one mapping. code is the stable
// Curator diagnostic code the seam reports; when it differs from the published
// protocol code, note records why the Curator seam is the equivalent boundary.
type driverRejection struct {
	boundary string
	code     string
	note     string
	// run reproduces the condition and returns the Curator error.
	run func(t *testing.T) error
}

// graphRejection builds a case that edits the go list package stream. The
// compiler must never start, so exactly one source-aware call is permitted.
func graphRejection(edit func(*workerFixture, *[]packageJSON)) func(t *testing.T) error {
	return func(t *testing.T) error {
		fixture := newSnapshotFixture(t)
		items := []packageJSON{fixture.rootPackage()}
		if edit != nil {
			edit(fixture, &items)
		}
		fixture.start(stubScript{ListStdout: string(encodePackages(t, items...)), Artifact: "artifact"})
		_, err := Build(context.Background(), fixture.request(ResourceLimits{Timeout: 30 * time.Second}))
		calls := fixture.sourceAwareCalls()
		if len(calls) != 1 || calls[0].Argv[0] != "list" {
			t.Fatalf("source-aware calls = %+v, the compiler must not start after a graph rejection", calls)
		}
		return err
	}
}

// snapshotRejection builds a case that only edits the frozen snapshot, so no
// source-aware Go child may start at all.
func snapshotRejection(edit func(*workerFixture)) func(t *testing.T) error {
	return func(t *testing.T) error {
		fixture := newSnapshotFixture(t)
		edit(fixture)
		fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact"})
		_, err := Build(context.Background(), fixture.request(ResourceLimits{Timeout: 30 * time.Second}))
		if calls := fixture.sourceAwareCalls(); len(calls) != 0 {
			t.Fatalf("source-aware calls = %+v, no Go child may start", calls)
		}
		return err
	}
}

// listFailure classifies a failed fixed go list from its bounded stderr.
func listFailure(stderr string) func(t *testing.T) error {
	return func(t *testing.T) error {
		fixture := newSnapshotFixture(t)
		fixture.start(stubScript{ListExit: 1, ListStderr: stderr})
		_, err := Build(context.Background(), fixture.request(ResourceLimits{Timeout: 30 * time.Second}))
		if calls := fixture.sourceAwareCalls(); len(calls) != 1 {
			t.Fatalf("source-aware calls = %d, want exactly the fixed go list", len(calls))
		}
		return err
	}
}

// buildFailure classifies a failed fixed go build from its bounded stderr. The
// package artifact is never produced and never started.
func buildFailure(stderr string) func(t *testing.T) error {
	return func(t *testing.T) error {
		fixture := newSnapshotFixture(t)
		fixture.start(stubScript{
			ListStdout: string(encodePackages(t, fixture.rootPackage())),
			BuildExit:  1, BuildStderr: stderr,
		})
		_, err := Build(context.Background(), fixture.request(ResourceLimits{Timeout: 30 * time.Second}))
		calls := fixture.sourceAwareCalls()
		if len(calls) != 2 || calls[1].Argv[0] != "build" {
			t.Fatalf("source-aware calls = %+v, want the fixed list then build", calls)
		}
		return err
	}
}

// probeFailure drives one deterministic fault into the package-independent
// probe sequence: 0 is go telemetry off, 1 is go version, 2 is go env.
func probeFailure(mutate func(t *testing.T, index int, output Output) (Output, error)) func(t *testing.T) error {
	return func(t *testing.T) error {
		root := testToolchain(t)
		host := hostFacts{runtimeGOROOT: root, goos: runtime.GOOS, goarch: runtime.GOARCH}
		valid := validProbeExecutor(t, root, host)
		executor := &recordingExecutor{run: func(index int, process Process) (Output, error) {
			output, err := valid.run(index, process)
			if err != nil {
				return output, err
			}
			return mutate(t, index, output)
		}}
		base := t.TempDir()
		_, err := establish(context.Background(), Config{
			PrivateBase: base, CuratorGo: filepath.Join(root, "bin", platformGoName), Executor: executor,
		}, host)
		assertDirectoryEmpty(t, base)
		return err
	}
}

func driverRejectionMappings() map[string]driverRejection {
	return map[string]driverRejection{
		// module
		"missing-root-go-mod": {
			boundary: "module", code: "build_module_missing",
			run: snapshotRejection(func(fixture *workerFixture) {
				if err := os.Remove(filepath.Join(fixture.buildRoot, "go.mod")); err != nil {
					fixture.t.Fatal(err)
				}
			}),
		},
		"nested-module": {
			boundary: "module", code: "nested_build_module",
			run: graphRejection(func(fixture *workerFixture, items *[]packageJSON) {
				(*items)[0].Module.Dir = fixture.sourceDir
				(*items)[0].Module.GoMod = filepath.Join(fixture.sourceDir, "go.mod")
			}),
		},

		// dependency-graph
		"non-main-package": {
			boundary: "dependency-graph", code: "build_package_not_main",
			run: graphRejection(func(_ *workerFixture, items *[]packageJSON) { (*items)[0].Name = "library" }),
		},
		"multiple-packages": {
			boundary: "dependency-graph", code: "build_package_ambiguous",
			run: graphRejection(func(_ *workerFixture, items *[]packageJSON) {
				second := (*items)[0]
				second.ImportPath += "/second"
				*items = append(*items, second)
			}),
		},
		"missing-vendored-dependency": {
			boundary: "dependency-graph", code: "vendor_dependency_missing",
			run: listFailure("cannot find module providing package example.test/missing: import lookup disabled by -mod=vendor\n"),
		},
		"inconsistent-vendor-modules": {
			boundary: "dependency-graph", code: "vendor_metadata_inconsistent",
			run: listFailure("go: inconsistent vendoring in /source\n"),
		},
		"workspace-only-dependency": {
			boundary: "dependency-graph", code: "workspace_dependency_forbidden",
			run: snapshotRejection(func(fixture *workerFixture) {
				writeTestFile(fixture.t, filepath.Join(fixture.buildRoot, "go.work"), []byte("go 1.23\nuse .\n"), 0o644)
			}),
		},
		"cgo-only-package": {
			boundary: "dependency-graph", code: "cgo_required",
			run: graphRejection(func(_ *workerFixture, items *[]packageJSON) {
				(*items)[0].CgoFiles = []string{"bridge.go"}
			}),
		},
		"native-c-input": {
			boundary: "dependency-graph", code: "go_native_input_forbidden",
			run: graphRejection(func(_ *workerFixture, items *[]packageJSON) {
				(*items)[0].CFiles = []string{"native.c"}
			}),
		},
		"native-cxx-input": {
			boundary: "dependency-graph", code: "go_native_input_forbidden",
			run: graphRejection(func(_ *workerFixture, items *[]packageJSON) {
				(*items)[0].CXXFiles = []string{"native.cc"}
			}),
		},
		"native-swig-input": {
			boundary: "dependency-graph", code: "go_native_input_forbidden",
			run: graphRejection(func(_ *workerFixture, items *[]packageJSON) {
				(*items)[0].SwigFiles = []string{"native.swig"}
			}),
		},
		"root-syso": {
			boundary: "dependency-graph", code: "go_syso_forbidden",
			run: graphRejection(func(_ *workerFixture, items *[]packageJSON) {
				(*items)[0].SysoFiles = []string{"root.syso"}
			}),
		},
		"transitive-syso": {
			boundary: "dependency-graph", code: "go_syso_forbidden",
			run: graphRejection(func(fixture *workerFixture, items *[]packageJSON) {
				*items = append(*items, transitivePackage(fixture, func(item *packageJSON) {
					item.SysoFiles = []string{"transitive.syso"}
				}))
			}),
		},
		"root-assembly-absolute-include": {
			boundary: "dependency-graph", code: "go_assembly_forbidden",
			run: graphRejection(func(_ *workerFixture, items *[]packageJSON) {
				(*items)[0].SFiles = []string{"root.s"}
			}),
		},
		"transitive-assembly-escaping-include": {
			boundary: "dependency-graph", code: "go_assembly_forbidden",
			run: graphRejection(func(fixture *workerFixture, items *[]packageJSON) {
				*items = append(*items, transitivePackage(fixture, func(item *packageJSON) {
					item.SFiles = []string{"transitive.s"}
				}))
			}),
		},
		"escaped-embed-input": {
			boundary: "dependency-graph", code: "go_embed_input_escape",
			run: graphRejection(func(_ *workerFixture, items *[]packageJSON) {
				(*items)[0].EmbedFiles = []string{"../../../../outside.txt"}
			}),
		},

		// compiler-directive
		"cgo-import-dynamic": {
			boundary: "compiler-directive", code: "go_forbidden_compiler_directive",
			run: directiveRejection("package main\n//go:cgo_import_dynamic x y z\nfunc main() {}\n", false),
		},
		"attempted-go-generate": {
			boundary: "compiler-directive", code: "go_generator_forbidden",
			run: directiveRejection("package main\n//go:generate forbidden\nfunc main() {}\n", false),
		},
		"default-pgo": {
			boundary: "compiler-directive", code: "go_pgo_forbidden",
			run: directiveRejection("package main\nfunc main() {}\n", true),
		},

		// toolchain
		"toolchain-switch-request": {
			boundary: "toolchain", code: "toolchain_switch_forbidden",
			run: snapshotRejection(func(fixture *workerFixture) {
				writeTestFile(fixture.t, filepath.Join(fixture.buildRoot, "go.mod"),
					[]byte("module example.test/build\n\ngo 1.23\ntoolchain\tgo1.99.0\n"), 0o644)
			}),
		},
		"unsupported-go-pre-1-23": {
			boundary: "toolchain", code: "unsupported_go_family",
			run: probeFailure(func(_ *testing.T, index int, output Output) (Output, error) {
				if index == 1 {
					output.Stdout = []byte("go version go1.22.9 " + runtime.GOOS + "/" + runtime.GOARCH + "\n")
				}
				return output, nil
			}),
		},
		"unsupported-go-future-family": {
			boundary: "toolchain", code: "unsupported_go_family",
			run: probeFailure(func(_ *testing.T, index int, output Output) (Output, error) {
				if index == 1 {
					output.Stdout = []byte("go version go1.99.0 " + runtime.GOOS + "/" + runtime.GOARCH + "\n")
				}
				return output, nil
			}),
		},
		"wrong-go-executable-path": {
			boundary: "toolchain", code: "toolchain_executable_mismatch",
			run: func(t *testing.T) error {
				inside, outside := t.TempDir(), testToolchain(t)
				if err := os.MkdirAll(filepath.Join(inside, "bin"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(outside, "bin", platformGoName), filepath.Join(inside, "bin", platformGoName)); err != nil {
					t.Skipf("this host cannot create the symbolic link the vector needs: %v", err)
				}
				_, _, _, err := selectToolchain(Config{CuratorGo: filepath.Join(inside, "bin", platformGoName)}, outside)
				return err
			},
		},
		"toolchain-digest-mismatch": {
			boundary: "toolchain", code: "toolchain_mutated",
			note: "Curator reports every fingerprinted-tree divergence found by the " +
				"post-exec revalidation as toolchain_mutated; the published digest " +
				"mismatch is exactly that revalidation failing before the child is trusted",
			run: func(t *testing.T) error {
				fixture := newSnapshotFixture(t)
				fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact"})
				writeTestFile(t, filepath.Join(fixture.goroot, "VERSION"), []byte("go1.25.5 mutated\n"), 0o644)
				return fixture.session.VerifyToolchain(context.Background())
			},
		},

		// process
		"poisoned-path": {
			boundary: "process", code: CodeWorkerProtocolInvalid,
			note: "Curator never inherits a host PATH: the worker environment guard " +
				"admits only the closed manager-owned environment",
			run: workerEnvironmentRejection(func(values map[string]string) { values["PATH"] = "/host/tools" }),
		},
		"inherited-goflags-toolexec": {
			boundary: "process", code: CodeWorkerProtocolInvalid,
			note: "the closed worker environment admits only GOFLAGS=\"\"",
			run:  workerEnvironmentRejection(func(values map[string]string) { values["GOFLAGS"] = "-toolexec=/bin/sh" }),
		},
		"inherited-goenv": {
			boundary: "process", code: CodeWorkerProtocolInvalid,
			note: "the closed worker environment admits only GOENV=off",
			run:  workerEnvironmentRejection(func(values map[string]string) { values["GOENV"] = "/host/env" }),
		},
		"inherited-gowork": {
			boundary: "process", code: CodeWorkerProtocolInvalid,
			note: "the closed worker environment admits only GOWORK=off",
			run:  workerEnvironmentRejection(func(values map[string]string) { values["GOWORK"] = "/host/go.work" }),
		},
		"vcs-metadata": {
			boundary: "process", code: CodeWorkerProtocolInvalid,
			note: "Curator forbids ambient VCS input structurally: -buildvcs=false is " +
				"part of both fixed argument vectors and the worker rejects any other vector",
			run: func(t *testing.T) error {
				_, _, roots, artifact := workerGuards(t)
				argv := make([]string, 0, len(buildArgumentPrefix)+3)
				for _, item := range buildArgumentPrefix {
					if item == "-buildvcs=false" {
						item = "-buildvcs=true"
					}
					argv = append(argv, item)
				}
				return validateFixedBuildArgv(append(argv, "-o", artifact, "."), artifact, roots)
			},
		},
		"repository-local-fake-go": {
			boundary: "process", code: "untrusted_go_executable",
			run: func(t *testing.T) error {
				repository := t.TempDir()
				root := filepath.Join(repository, "vendor-go")
				writeTestFile(t, filepath.Join(root, "bin", platformGoName), testExecutableBytes(), 0o755)
				writeTestFile(t, filepath.Join(root, "VERSION"), []byte("go1.25.5\n"), 0o644)
				_, _, _, err := selectToolchain(Config{
					CuratorGo:      filepath.Join(root, "bin", platformGoName),
					ForbiddenRoots: []string{repository},
				}, root)
				return err
			},
		},
		"telemetry-command-failure": {
			boundary: "process", code: "telemetry_initialization_failed",
			run: probeFailure(func(_ *testing.T, index int, output Output) (Output, error) {
				if index == 0 {
					return output, errors.New("exit status 2")
				}
				return output, nil
			}),
		},
		"telemetry-private-dir-escape": {
			boundary: "process", code: "telemetry_directory_untrusted",
			run: probeFailure(func(t *testing.T, index int, output Output) (Output, error) {
				if index == 2 {
					output.Stdout = mutateJSON(t, output.Stdout, "GOTELEMETRYDIR", t.TempDir())
				}
				return output, nil
			}),
		},
		"external-link-required": {
			boundary: "process", code: "external_link_forbidden",
			run: buildFailure("# example.test/build/cmd/tool\nruntime/cgo: requires external linking\n"),
		},
		"libgcc-fallback-attempt": {
			boundary: "process", code: "libgcc_fallback_forbidden",
			run: buildFailure("# example.test/build/cmd/tool\nlink: running gcc failed: exec: \"gcc\": libgcc lookup\n"),
		},
		"child-outside-goroot-tools": {
			boundary: "process", code: CodeWorkerIdentityInvalid,
			note: "the manager-selected process graph is closed: a program started " +
				"below the worker that is not the fingerprinted GOROOT tool child is " +
				"an identity failure before anything is published",
			run: func(t *testing.T) error {
				scenario := newWorkerScenario(t)
				scenario.request.ToolDirectory = filepath.Join(scenario.fixture.root, "pkg", "tool")
				code := scenario.failureCode()
				scenario.requireNoCompilerStarted()
				return diagnostic(code, "worker refused the request")
			},
		},
	}
}

// failureCode starts the real worker over this scenario's request and returns
// the stable diagnostic code it refused with.
func (scenario *workerScenario) failureCode() string {
	scenario.fixture.t.Helper()
	worker := scenario.start()
	message := worker.receive()
	if message.Kind != kindFailure || message.Failure == nil {
		scenario.fixture.t.Fatalf("worker sent %q, want a failure", message.Kind)
	}
	return message.Failure.Code
}

// transitivePackage returns a dependency of the root package inside the same
// module, so the transitive vectors exercise a real non-root graph member.
func transitivePackage(fixture *workerFixture, edit func(*packageJSON)) packageJSON {
	directory := filepath.Join(fixture.buildRoot, "internal", "render")
	writeTestFile(fixture.t, filepath.Join(directory, "render.go"), []byte("package render\n"), 0o644)
	item := packageJSON{
		Dir:        filepath.Join(fixture.buildRoot, "internal", "render"),
		ImportPath: "example.test/build/internal/render",
		Name:       "render", DepOnly: true, GoFiles: []string{"render.go"},
		Module: &moduleJSON{
			Path: "example.test/build", Main: true, Dir: fixture.buildRoot,
			GoMod: filepath.Join(fixture.buildRoot, "go.mod"), GoVersion: "1.23",
		},
	}
	edit(&item)
	return item
}

// directiveRejection writes a real source file so the directive scan runs
// against actual package bytes.
func directiveRejection(content string, pgo bool) func(t *testing.T) error {
	return func(t *testing.T) error {
		fixture := newSnapshotFixture(t)
		writeTestFile(t, filepath.Join(fixture.sourceDir, "main.go"), []byte(content), 0o644)
		if pgo {
			writeTestFile(t, filepath.Join(fixture.sourceDir, "default.pgo"), []byte("profile"), 0o644)
		}
		fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact"})
		_, err := Build(context.Background(), fixture.request(ResourceLimits{Timeout: 30 * time.Second}))
		if calls := fixture.sourceAwareCalls(); len(calls) != 1 {
			t.Fatalf("source-aware calls = %d, the compiler must not start", len(calls))
		}
		return err
	}
}

// workerEnvironmentRejection proves the closed worker environment refuses one
// inherited host value.
func workerEnvironmentRejection(mutate func(map[string]string)) func(t *testing.T) error {
	return func(t *testing.T) error {
		environment, goroot, roots, _ := workerGuards(t)
		values := environmentMap(environment)
		mutate(values)
		return validateWorkerEnvironment(environmentSlice(values), goroot, roots)
	}
}

// TestDriverRejectionClustersMapToStableCuratorErrors proves every authoritative
// module, dependency-graph, compiler-directive, toolchain, and process rejection
// reaches a stable Curator diagnostic without the package's own code ever being
// compiled, linked, or started.
func TestDriverRejectionClustersMapToStableCuratorErrors(t *testing.T) {
	published := loadDriverRejections(t)
	owned := driverRejectionMappings()
	boundaries := map[string]bool{
		"module": true, "dependency-graph": true, "compiler-directive": true,
		"toolchain": true, "process": true,
	}

	names := make([]string, 0, len(published))
	for name, vector := range published {
		if boundaries[vector.Boundary] {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	for _, name := range names {
		vector := published[name]
		mapping, ok := owned[name]
		if !ok {
			t.Errorf("authoritative rejection %q has no Curator mapping", name)
			continue
		}
		t.Run(name, func(t *testing.T) {
			if vector.Expected.Result != "reject" || vector.Expected.Reuse || vector.Expected.ArtifactExecuted {
				t.Fatalf("vector %q no longer fails closed: %+v", name, vector.Expected)
			}
			if mapping.boundary != vector.Boundary {
				t.Fatalf("Curator owns %q at %s, the suite publishes %s", name, mapping.boundary, vector.Boundary)
			}
			if mapping.code != vector.Expected.Error && mapping.note == "" {
				t.Fatalf("Curator code %q differs from the published %q without a recorded equivalence",
					mapping.code, vector.Expected.Error)
			}
			err := mapping.run(t)
			if err == nil {
				t.Fatalf("%s was accepted, want the %s rejection", name, vector.Expected.Error)
			}
			if got := DiagnosticCode(err); got != mapping.code {
				t.Fatalf("%s Curator code = %q (%v), want %q", name, got, err, mapping.code)
			}
		})
	}

	for name := range owned {
		if _, ok := published[name]; !ok {
			t.Errorf("Curator maps %q, which the authoritative suite no longer publishes", name)
		}
	}
}
