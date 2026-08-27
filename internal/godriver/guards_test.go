package godriver

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// workerGuards returns a live session's closed environment plus the derived
// roots, so the worker-side guards can be exercised directly against real
// manager-owned values.
func workerGuards(t *testing.T) (environment []string, goroot string, privateRoots []string, artifact string) {
	t.Helper()
	fixture := newSnapshotFixture(t)
	fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact"})
	stage, err := os.MkdirTemp(fixture.session.operation, ".curator-go-build-")
	if err != nil {
		t.Fatal(err)
	}
	artifact = filepath.Join(stage, "bin", "golden-tool")
	if err := os.MkdirAll(filepath.Dir(artifact), 0o700); err != nil {
		t.Fatal(err)
	}
	environment = fixture.session.Environment()
	return environment, fixture.session.GOROOT(), privateRoots2(environment, stage), artifact
}

func privateRoots2(environment []string, stage string) []string {
	return privateRoots(environment, stage)
}

func TestWorkerEnvironmentGuardAcceptsOnlyTheClosedOfflineEnvironment(t *testing.T) {
	environment, goroot, roots, _ := workerGuards(t)
	if err := validateWorkerEnvironment(environment, goroot, roots); err != nil {
		t.Fatalf("closed environment rejected: %v", err)
	}
	for _, testCase := range []struct {
		name   string
		mutate func(map[string]string)
	}{
		{name: "network proxy", mutate: func(values map[string]string) { values["GOPROXY"] = "https://proxy.example" }},
		{name: "checksum database", mutate: func(values map[string]string) { values["GOSUMDB"] = "sum.golang.org" }},
		{name: "version control", mutate: func(values map[string]string) { values["GOVCS"] = "*:all" }},
		{name: "workspace", mutate: func(values map[string]string) { values["GOWORK"] = "/tmp/go.work" }},
		{name: "cgo", mutate: func(values map[string]string) { values["CGO_ENABLED"] = "1" }},
		{name: "external link", mutate: func(values map[string]string) { values["GO_EXTLINK_ENABLED"] = "1" }},
		{name: "toolchain switch", mutate: func(values map[string]string) { values["GOTOOLCHAIN"] = "auto" }},
		{name: "go flags", mutate: func(values map[string]string) { values["GOFLAGS"] = "-toolexec=/bin/sh" }},
		{name: "go env file", mutate: func(values map[string]string) { values["GOENV"] = "/tmp/env" }},
		{name: "host compiler", mutate: func(values map[string]string) { values["CC"] = "/usr/bin/clang" }},
		{name: "credential helper", mutate: func(values map[string]string) { values["GOAUTH"] = "netrc" }},
		{name: "proxy variable", mutate: func(values map[string]string) { values["HTTPS_PROXY"] = "http://proxy" }},
		{name: "tool directory", mutate: func(values map[string]string) { values["GOTOOLDIR"] = "/tmp/tools" }},
		{name: "output binary directory", mutate: func(values map[string]string) { values["GOBIN"] = "/tmp/bin" }},
		{name: "non protocol variable", mutate: func(values map[string]string) { values["CURATOR_EXTRA"] = "1" }},
		{name: "shared cache", mutate: func(values map[string]string) { values["GOCACHE"] = os.TempDir() }},
		{name: "shared home", mutate: func(values map[string]string) { values["HOME"] = os.TempDir() }},
		{name: "missing target", mutate: func(values map[string]string) { delete(values, "GOARCH") }},
		{name: "foreign GOROOT", mutate: func(values map[string]string) { values["GOROOT"] = os.TempDir() }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			values := environmentMap(environment)
			testCase.mutate(values)
			err := validateWorkerEnvironment(environmentSlice(values), goroot, roots)
			if DiagnosticCode(err) != CodeWorkerProtocolInvalid {
				t.Fatalf("error = %v, want %s", err, CodeWorkerProtocolInvalid)
			}
		})
	}
	t.Run("malformed entry", func(t *testing.T) {
		err := validateWorkerEnvironment(append(environment, "=novalue"), goroot, roots)
		if DiagnosticCode(err) != CodeWorkerProtocolInvalid {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("duplicate entry", func(t *testing.T) {
		err := validateWorkerEnvironment(append(environment, "GOROOT="+goroot), goroot, roots)
		if DiagnosticCode(err) != CodeWorkerProtocolInvalid {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestFixedBuildArgvGuardAcceptsOnlyTheProtocolVector(t *testing.T) {
	_, _, roots, artifact := workerGuards(t)
	valid := append(append([]string(nil), buildArgumentPrefix...), artifact, ".")
	if err := validateFixedBuildArgv(valid, artifact, roots); err != nil {
		t.Fatalf("protocol vector rejected: %v", err)
	}
	for _, testCase := range []struct {
		name string
		argv []string
	}{
		{name: "package pattern", argv: append(append([]string(nil), buildArgumentPrefix...), artifact, "./...")},
		{name: "extra flag", argv: append(append(append([]string(nil), buildArgumentPrefix...), artifact, "-race"), ".")},
		{name: "short vector", argv: []string{"build", "."}},
		{name: "relative output", argv: append(append([]string(nil), buildArgumentPrefix...), "bin/tool", ".")},
		{name: "other command", argv: append(append([]string(nil), "run"), valid[1:]...)},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if code := DiagnosticCode(validateFixedBuildArgv(testCase.argv, artifact, roots)); code != CodeWorkerProtocolInvalid {
				t.Fatalf("code = %q, want %s", code, CodeWorkerProtocolInvalid)
			}
		})
	}
	t.Run("output outside private staging", func(t *testing.T) {
		escaped := filepath.Join(t.TempDir(), "escaped")
		argv := append(append([]string(nil), buildArgumentPrefix...), escaped, ".")
		if code := DiagnosticCode(validateFixedBuildArgv(argv, escaped, roots)); code != CodeWorkerProtocolInvalid {
			t.Fatalf("code = %q", code)
		}
	})
}

func TestProbeSetGuardMatchesTheInventoryExactly(t *testing.T) {
	platform := InventoryPlatform(runtime.GOOS)
	probes := syntheticProbes(platform)
	if err := validateProbeSet(platform, probes); err != nil {
		t.Fatalf("inventory probe set rejected: %v", err)
	}
	// Inside the worker every probe fault is an evidence fault. The
	// mandatory-control rejection belongs exclusively to the parent's
	// pre-worker boundary.
	if code := DiagnosticCode(validateProbeSet("linux", probes)); code != CodeCapabilityEvidenceInvalid {
		t.Fatalf("uncovered platform code = %q", code)
	}
	short := append([]ControlProbe(nil), probes[:len(probes)-1]...)
	if code := DiagnosticCode(validateProbeSet(platform, short)); code != CodeCapabilityEvidenceInvalid {
		t.Fatalf("short probe set code = %q", code)
	}
	reordered := append([]ControlProbe(nil), probes...)
	reordered[0], reordered[1] = reordered[1], reordered[0]
	if code := DiagnosticCode(validateProbeSet(platform, reordered)); code != CodeCapabilityEvidenceInvalid {
		t.Fatalf("reordered probe set code = %q", code)
	}
	unknownTiming := append([]ControlProbe(nil), probes...)
	unknownTiming[0].ProbedAt = "install-time"
	if code := DiagnosticCode(validateProbeSet(platform, unknownTiming)); code != CodeCapabilityEvidenceInvalid {
		t.Fatalf("unknown probe timing code = %q", code)
	}
	unknownAvailability := append([]ControlProbe(nil), probes...)
	unknownAvailability[0].Availability = "partial"
	if code := DiagnosticCode(validateProbeSet(platform, unknownAvailability)); code != CodeCapabilityEvidenceInvalid {
		t.Fatalf("unknown availability code = %q", code)
	}
}

func TestAppliedControlReportMustAgreeWithItsEvidence(t *testing.T) {
	platform := PlatformWindows
	probes := syntheticProbes(platform)
	var applied []string
	for _, probe := range probes {
		if probe.Availability == AvailabilityAvailable {
			applied = append(applied, probe.Name)
		}
	}
	evidence := evidenceFromApplied(platform, probes, applied)
	if err := matchAppliedControls(applied, evidence); err != nil {
		t.Fatalf("consistent report rejected: %v", err)
	}
	if code := DiagnosticCode(matchAppliedControls(applied[:len(applied)-1], evidence)); code != CodeCapabilityEvidenceInvalid {
		t.Fatalf("under-report code = %q", code)
	}
	if code := DiagnosticCode(matchAppliedControls(append(applied, ControlPerFileSizeLimit), evidence)); code != CodeCapabilityEvidenceInvalid {
		t.Fatalf("over-report code = %q", code)
	}
	if code := DiagnosticCode(matchAppliedControls(append(applied, applied[0]), evidence)); code != CodeCapabilityEvidenceInvalid {
		t.Fatalf("duplicate report code = %q", code)
	}
	// An available control the worker did not apply is recorded faithfully and
	// then rejected, rather than silently downgraded.
	partial := evidenceFromApplied(platform, probes, applied[:len(applied)-1])
	if code := DiagnosticCode(validateCapabilityEvidence(partial, platform, probes)); code != CodeCapabilityEvidenceInvalid {
		t.Fatalf("unapplied available control code = %q", code)
	}
}

func TestResourceLimitsStayInsideManagerBounds(t *testing.T) {
	normalized, err := normalizeBuildLimits(ResourceLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if normalized.Timeout != defaultBuildTimeout || normalized.OutputBytes != defaultBuildOutput ||
		normalized.ArtifactBytes != defaultArtifactLimit || normalized.FileBytes != defaultFileLimit ||
		normalized.DiskBytes != defaultDiskLimit || normalized.MemoryBytes != defaultMemoryLimit ||
		normalized.Processes != defaultProcessLimit {
		t.Fatalf("defaults = %s", normalized)
	}
	for _, limits := range []ResourceLimits{
		{Timeout: defaultBuildTimeout + time.Second},
		{OutputBytes: defaultBuildOutput + 1},
		{ArtifactBytes: defaultArtifactLimit + 1},
		{ArtifactBytes: 1024, FileBytes: 512},
		{ArtifactBytes: 1024, DiskBytes: 512},
		{MemoryBytes: defaultMemoryLimit + 1},
		{Processes: defaultProcessLimit + 1},
		{Processes: -1},
	} {
		if _, err := normalizeBuildLimits(limits); DiagnosticCode(err) != "invalid_resource_limits" {
			t.Fatalf("limits %+v were accepted: %v", limits, err)
		}
	}
}

func TestExecutableIdentityRejectsSubstitutionAndMutation(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "manager")
	writeTestFile(t, path, []byte("manager bytes"), 0o755)
	identity, err := resolveExecutableIdentity(path)
	if err != nil {
		t.Fatal(err)
	}
	if identity.Path != mustPhysical(t, path) || identity.Size != int64(len("manager bytes")) || !strings.HasPrefix(identity.SHA256, "sha256:") {
		t.Fatalf("identity = %+v", identity)
	}
	if err := identity.Verify(); err != nil {
		t.Fatalf("unchanged executable rejected: %v", err)
	}
	if code := DiagnosticCode(identity.matches(identity.Path, identity.SHA256, identity.Size+1)); code != CodeWorkerIdentityInvalid {
		t.Fatalf("size mismatch code = %q", code)
	}
	writeTestFile(t, path, []byte("replaced bytes"), 0o755)
	if code := DiagnosticCode(identity.Verify()); code != CodeWorkerIdentityInvalid {
		t.Fatalf("replaced executable code = %q", code)
	}
	if code := DiagnosticCode(mustFail(resolveExecutableIdentity(filepath.Join(directory, "missing")))); code != CodeWorkerIdentityInvalid {
		t.Fatalf("missing executable code = %q", code)
	}
	if code := DiagnosticCode(mustFail(resolveExecutableIdentity(directory))); code != CodeWorkerIdentityInvalid {
		t.Fatalf("directory code = %q", code)
	}
	if code := DiagnosticCode(mustFail(resolveExecutableIdentity(""))); code != CodeWorkerIdentityInvalid {
		t.Fatalf("empty path code = %q", code)
	}
	empty := filepath.Join(directory, "empty")
	writeTestFile(t, empty, nil, 0o755)
	if code := DiagnosticCode(mustFail(resolveExecutableIdentity(empty))); code != CodeWorkerIdentityInvalid {
		t.Fatalf("empty executable code = %q", code)
	}
}

func TestSessionFramingRejectsMalformedAndOversizeFrames(t *testing.T) {
	var buffer bytes.Buffer
	if err := writeMessage(&buffer, workerMessage{Kind: kindShutdown, Nonce: "abc"}); err != nil {
		t.Fatal(err)
	}
	message, err := readMessage(bytes.NewReader(buffer.Bytes()))
	if err != nil || message.Kind != kindShutdown || message.Nonce != "abc" {
		t.Fatalf("round trip = %+v, %v", message, err)
	}
	for _, testCase := range []struct {
		name    string
		payload []byte
	}{
		{name: "zero length", payload: []byte{0, 0, 0, 0}},
		{name: "oversize length", payload: []byte{0xff, 0xff, 0xff, 0xff}},
		{name: "truncated payload", payload: []byte{0, 0, 0, 8, '{'}},
		{name: "unknown field", payload: frame(`{"kind":"list","nonce":"a","extra":1}`)},
		{name: "unknown kind", payload: frame(`{"kind":"run","nonce":"a"}`)},
		{name: "trailing data", payload: frame(`{"kind":"list","nonce":"a"} {}`)},
		{name: "not an object", payload: frame(`["list"]`)},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := readMessage(bytes.NewReader(testCase.payload))
			if DiagnosticCode(err) != CodeWorkerProtocolInvalid {
				t.Fatalf("error = %v, want %s", err, CodeWorkerProtocolInvalid)
			}
		})
	}
}

func TestBuildPermitIsBoundToTheSessionAndVector(t *testing.T) {
	argv := append(append([]string(nil), buildArgumentPrefix...), "/private/bin/tool", ".")
	permit := buildPermit("00ff", "nonce", argv)
	if !validPermit(permit, "00ff", "nonce", argv) {
		t.Fatal("a permit is not valid for its own session and vector")
	}
	if validPermit(permit, "00ff", "other-nonce", argv) {
		t.Fatal("a permit replayed into another session was accepted")
	}
	if validPermit(permit, "1100", "nonce", argv) {
		t.Fatal("a permit minted under another secret was accepted")
	}
	other := append(append([]string(nil), buildArgumentPrefix...), "/private/bin/other", ".")
	if validPermit(permit, "00ff", "nonce", other) {
		t.Fatal("a permit covering another vector was accepted")
	}
	if validPermit("", "00ff", "nonce", argv) {
		t.Fatal("an empty permit was accepted")
	}
}

func TestPackageCommandSurfaceIsClosed(t *testing.T) {
	base := BuildRequest{SourceDir: "build/cmd/tool"}
	base.CommandObject = BuildCommand{"type": "build", "driver": "go-v1", "source_dir": "build/cmd/tool"}
	if err := validatePackageCommandSurface(base); err != nil {
		t.Fatalf("closed surface rejected: %v", err)
	}
	for _, testCase := range []struct {
		name   string
		mutate func(BuildCommand)
	}{
		{name: "wrong type", mutate: func(object BuildCommand) { object["type"] = "script" }},
		{name: "wrong driver", mutate: func(object BuildCommand) { object["driver"] = "go-repository-v1" }},
		{name: "mismatched source dir", mutate: func(object BuildCommand) { object["source_dir"] = "build/other" }},
		{name: "unknown field", mutate: func(object BuildCommand) { object["curator_worker"] = "yes" }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			request := BuildRequest{SourceDir: base.SourceDir, CommandObject: BuildCommand{
				"type": "build", "driver": "go-v1", "source_dir": "build/cmd/tool",
			}}
			testCase.mutate(request.CommandObject)
			if code := DiagnosticCode(validatePackageCommandSurface(request)); code != CodePackageInfluenceForbidden {
				t.Fatalf("code = %q, want %s", code, CodePackageInfluenceForbidden)
			}
		})
	}
}

func frame(payload string) []byte {
	header := []byte{byte(len(payload) >> 24), byte(len(payload) >> 16), byte(len(payload) >> 8), byte(len(payload))}
	return append(header, []byte(payload)...)
}

func mustFail(_ ExecutableIdentity, err error) error { return err }
