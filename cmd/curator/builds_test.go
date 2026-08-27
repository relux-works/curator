package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildsource"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/godriver"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/marker"
)

// testDigest builds a distinct, well-formed sha256 identity per seed so a
// classification test can assert on identity mismatches rather than on shape.
func testDigest(seed int) string {
	return "sha256:" + strings.Repeat(string("0123456789abcdef"[seed%16]), 64)
}

func testBuildSource() buildsource.Identity {
	return buildsource.Identity{Algorithm: buildsource.Algorithm, ContentSHA256: testDigest(1)}
}

// testFacts mirrors what the planner actually produces: receipt and artifact
// values exist only for a hit, because only a hit read a protected entry.
func testFacts(outcome string) buildFacts {
	facts := buildFacts{
		Skill: "build-skill", Command: "build-tool", Driver: buildmeta.DriverGoV1,
		BuildRoot: "assets/build-tool", SourceDir: "assets/build-tool/cmd/tool",
		Source:   testBuildSource(),
		Target:   buildmeta.Target{GOOS: "linux", GOARCH: "amd64", Tuning: map[string]string{"GOAMD64": "v1"}},
		CacheKey: buildmeta.CacheKey(testDigest(2)),
		Outcome:  outcome,
	}
	if outcome == string(install.BuildCacheHit) {
		facts.ReceiptHash = buildmeta.ReceiptHash(testDigest(3))
		facts.Artifact = buildmeta.Artifact{Path: "bin/build-tool", SHA256: testDigest(4), Size: 4096}
	}
	return facts
}

func testRecordedBuild() marker.Build {
	return marker.Build{
		Driver: buildmeta.DriverGoV1, CacheKey: buildmeta.CacheKey(testDigest(2)),
		ReceiptSHA256: buildmeta.ReceiptHash(testDigest(3)), ArtifactSHA256: testDigest(4),
		ArtifactPath: "bin/build-tool",
	}
}

func testRecordedMarker() *marker.Marker {
	source := testBuildSource()
	return &marker.Marker{
		SchemaVersion: marker.SchemaVersion, Name: "build-skill",
		Commands:    []string{"build-tool"},
		BuildRoots:  []string{"assets/build-tool"},
		BuildSource: &source,
		Builds:      map[string]marker.Build{"build-tool": testRecordedBuild()},
		Files:       []string{"SKILL.md", "assets/prompt.md"},
	}
}

// markerRecording returns the marker of testRecordedMarker with one build
// entry replaced, so a classification case varies exactly one recorded value.
func markerRecording(build marker.Build) *marker.Marker {
	recorded := testRecordedMarker()
	recorded.Builds["build-tool"] = build
	return recorded
}

// TestClassifyBuildCommandMapsEveryCacheOutcomeToADistinctCode proves the
// planner vocabulary and the currentness vocabulary stay in exact
// correspondence, and that a hit is the only code status accepts.
func TestClassifyBuildCommandMapsEveryCacheOutcomeToADistinctCode(t *testing.T) {
	cases := map[string]string{
		string(install.BuildCacheHit):                   buildCurrent,
		string(install.BuildWouldPreflightAndBuild):     buildMissingArtifact,
		string(install.BuildWouldRebuildUntrustedCache): buildUntrustedCache,
		string(install.BuildCorrupt):                    buildCorruptCache,
		string(install.BuildUnsupported):                buildUnsupportedPlatform,
		string(install.BuildToolchainUnavailable):       buildUnusableToolchain,
		"a-future-outcome":                              buildUnknownState,
	}
	seen := map[string]string{}
	for outcome, want := range cases {
		state, _, detail := classifyBuildCommand(testRecordedBuild(), testRecordedMarker(), testFacts(outcome))
		if state != want {
			t.Fatalf("outcome %q classified as %q, want %q (detail %q)", outcome, state, want, detail)
		}
		if previous, duplicate := seen[state]; duplicate {
			t.Fatalf("outcome %q and %q share the code %q", outcome, previous, state)
		}
		seen[state] = outcome
		if state != buildCurrent && detail == "" {
			t.Fatalf("outcome %q produced code %q without an operator detail", outcome, state)
		}
		if state == buildCurrent && detail != "" {
			t.Fatalf("a cache hit produced a detail: %q", detail)
		}
	}
}

// TestUnusableToolchainRowsCarryTheDriverBoundaryCode proves the toolchain
// diagnostic keeps the stable go-v1 code as its machine-readable cause instead
// of leaving an operator to parse prose.
func TestUnusableToolchainRowsCarryTheDriverBoundaryCode(t *testing.T) {
	facts := testFacts(string(install.BuildToolchainUnavailable))
	facts.Diagnostic = "go_toolchain_missing"
	facts.Reason = "go-v1 go_toolchain_missing: no trusted Go installation was selected"

	state, cause, detail := classifyBuildCommand(testRecordedBuild(), testRecordedMarker(), facts)
	if state != buildUnusableToolchain || cause != "go_toolchain_missing" {
		t.Fatalf("state = %q cause = %q", state, cause)
	}
	row := report(facts, state, cause, detail)
	if row.Cause != "go_toolchain_missing" || !strings.Contains(row.Describe(), "cause=go_toolchain_missing") {
		t.Fatalf("row does not publish the boundary code: %s", row.Describe())
	}
}

func TestClassifyBuildCommandDetectsRecordedIdentityDrift(t *testing.T) {
	other := buildsource.Identity{Algorithm: buildsource.Algorithm, ContentSHA256: testDigest(9)}
	sourcelessMarker := func() *marker.Marker {
		recorded := testRecordedMarker()
		recorded.BuildSource = nil
		return recorded
	}()

	for name, testCase := range map[string]struct {
		recorded marker.Build
		marker   *marker.Marker
		facts    buildFacts
		want     string
	}{
		"unsupported driver in the marker": {
			recorded: func() marker.Build { b := testRecordedBuild(); b.Driver = "kotlin-v1"; return b }(),
			marker:   testRecordedMarker(), facts: testFacts(string(install.BuildCacheHit)),
			want: buildUnsupportedDriver,
		},
		"unsupported driver in the plan": {
			recorded: testRecordedBuild(), marker: testRecordedMarker(),
			facts: func() buildFacts { f := testFacts(string(install.BuildCacheHit)); f.Driver = "kotlin-v1"; return f }(),
			want:  buildUnsupportedDriver,
		},
		"marker without a build source": {
			recorded: testRecordedBuild(), marker: sourcelessMarker,
			facts: testFacts(string(install.BuildCacheHit)), want: stateInvalidMarker,
		},
		"snapshot identity moved": {
			recorded: testRecordedBuild(),
			marker: func() *marker.Marker {
				recorded := testRecordedMarker()
				recorded.BuildSource = &other
				return recorded
			}(),
			facts: testFacts(string(install.BuildCacheHit)), want: buildSourceDrift,
		},
		"logical key moved": {
			recorded: func() marker.Build {
				b := testRecordedBuild()
				b.CacheKey = buildmeta.CacheKey(testDigest(7))
				return b
			}(),
			marker: testRecordedMarker(), facts: testFacts(string(install.BuildCacheHit)),
			want: buildInputDrift,
		},
		"receipt identity moved": {
			recorded: func() marker.Build {
				b := testRecordedBuild()
				b.ReceiptSHA256 = buildmeta.ReceiptHash(testDigest(8))
				return b
			}(),
			marker: testRecordedMarker(), facts: testFacts(string(install.BuildCacheHit)),
			want: buildCorruptReceipt,
		},
		"artifact hash moved": {
			recorded: func() marker.Build { b := testRecordedBuild(); b.ArtifactSHA256 = testDigest(6); return b }(),
			marker:   testRecordedMarker(), facts: testFacts(string(install.BuildCacheHit)),
			want: buildArtifactDrift,
		},
		"artifact path moved": {
			recorded: func() marker.Build { b := testRecordedBuild(); b.ArtifactPath = "bin/other"; return b }(),
			marker:   testRecordedMarker(), facts: testFacts(string(install.BuildCacheHit)),
			want: buildArtifactDrift,
		},
	} {
		t.Run(name, func(t *testing.T) {
			state, _, detail := classifyBuildCommand(testCase.recorded, testCase.marker, testCase.facts)
			if state != testCase.want {
				t.Fatalf("state = %q, want %q (detail %q)", state, testCase.want, detail)
			}
		})
	}
}

// testInput is the logical build input every adversarial component case
// derives from. It is a real, valid go-v1 input, so every key below is the key
// the manager would actually derive for that input.
func testInput() buildmeta.Input {
	return buildmeta.Input{
		SchemaVersion: buildmeta.SchemaVersion,
		Driver:        buildmeta.DriverGoV1,
		BuildSource:   testBuildSource(),
		BuildRoot:     "assets/build-tool",
		Command:       "build-tool",
		SourceDir:     "assets/build-tool/cmd/tool",
		Target: buildmeta.Target{
			GOOS: "linux", GOARCH: "amd64", Tuning: map[string]string{"GOAMD64": "v1"},
		},
		Toolchain: buildmeta.Toolchain{
			Algorithm: "curator-go-toolchain-v1", GoRelpath: "bin/go",
			GoVersion: "go1.25.5", ContentSHA256: testDigest(11),
		},
		Policy: buildmeta.FixedPolicy(),
	}
}

func keyOf(t *testing.T, input buildmeta.Input) buildmeta.CacheKey {
	t.Helper()
	if err := input.Validate(); err != nil {
		t.Fatalf("adversarial input is not a valid go-v1 input: %v", err)
	}
	key, err := input.CacheKey()
	if err != nil {
		t.Fatal(err)
	}
	return key
}

// TestBuildInputDriftIsAttributedOnlyAsFarAsTheMarkerProves is the adversarial
// pass over every component of the logical build input.
//
// Each case moves exactly one component, derives the real key for the moved
// input, and asserts that the classification neither misses the drift nor
// claims a cause the install marker cannot support. In particular, a toolchain,
// source-directory, architecture, tuning, or policy change must never be
// reported as a target change: the marker records no prior input, so those are
// indistinguishable and are reported as unattributed.
func TestBuildInputDriftIsAttributedOnlyAsFarAsTheMarkerProves(t *testing.T) {
	baseline := testInput()
	baselineKey := keyOf(t, baseline)

	for name, testCase := range map[string]struct {
		moved func(buildmeta.Input) buildmeta.Input
		// recordedRoots is the build-root set the install marker recorded.
		recordedRoots []string
		want          string
	}{
		"build root": {
			moved: func(input buildmeta.Input) buildmeta.Input {
				input.BuildRoot = "assets/other-tool"
				input.SourceDir = "assets/other-tool/cmd/tool"
				return input
			},
			recordedRoots: []string{"assets/build-tool"}, want: causeBuildRoot,
		},
		"source directory": {
			moved: func(input buildmeta.Input) buildmeta.Input {
				input.SourceDir = "assets/build-tool/cmd/other"
				return input
			},
			recordedRoots: []string{"assets/build-tool"}, want: causeUnattributed,
		},
		"target operating system": {
			moved: func(input buildmeta.Input) buildmeta.Input {
				input.Target.GOOS = "windows"
				return input
			},
			recordedRoots: []string{"assets/build-tool"}, want: causeTarget,
		},
		"target architecture": {
			moved: func(input buildmeta.Input) buildmeta.Input {
				input.Target.GOARCH = "arm64"
				input.Target.Tuning = map[string]string{"GOARM64": "v8.0"}
				return input
			},
			recordedRoots: []string{"assets/build-tool"}, want: causeUnattributed,
		},
		"target tuning": {
			moved: func(input buildmeta.Input) buildmeta.Input {
				input.Target.Tuning = map[string]string{"GOAMD64": "v3"}
				return input
			},
			recordedRoots: []string{"assets/build-tool"}, want: causeUnattributed,
		},
		"toolchain release": {
			moved: func(input buildmeta.Input) buildmeta.Input {
				input.Toolchain.GoVersion = "go1.24.9"
				return input
			},
			recordedRoots: []string{"assets/build-tool"}, want: causeUnattributed,
		},
		"toolchain content": {
			moved: func(input buildmeta.Input) buildmeta.Input {
				input.Toolchain.ContentSHA256 = testDigest(12)
				return input
			},
			recordedRoots: []string{"assets/build-tool"}, want: causeUnattributed,
		},
	} {
		t.Run(name, func(t *testing.T) {
			moved := testCase.moved(baseline)
			movedKey := keyOf(t, moved)
			if movedKey == baselineKey {
				t.Fatalf("moving %s did not change the logical key, so the case proves nothing", name)
			}

			artifact, err := buildmeta.ArtifactPath(baseline.Command, baseline.Target.GOOS)
			if err != nil {
				t.Fatal(err)
			}
			recorded := testRecordedBuild()
			recorded.CacheKey = baselineKey
			recorded.ArtifactPath = artifact
			recordedMarker := markerRecording(recorded)
			recordedMarker.BuildRoots = testCase.recordedRoots

			facts := testFacts(string(install.BuildCacheHit))
			facts.BuildRoot = moved.BuildRoot
			facts.SourceDir = moved.SourceDir
			facts.Target = moved.Target
			facts.CacheKey = movedKey

			state, cause, detail := classifyBuildCommand(recorded, recordedMarker, facts)
			if state != buildInputDrift {
				t.Fatalf("state = %q, want %q (detail %q)", state, buildInputDrift, detail)
			}
			if cause != testCase.want {
				t.Fatalf("cause = %q, want %q (detail %q)", cause, testCase.want, detail)
			}
			if detail == "" {
				t.Fatal("input drift produced no operator detail")
			}
		})
	}

	// The build schema version and the manager build policy are the two
	// components no valid input can vary: this manager refuses to derive a key
	// for either. A drift in them can therefore only arrive as a key some other
	// manager version recorded, which is exactly an opaque digest this manager
	// cannot attribute — and it must not be blamed on the target.
	t.Run("another manager version recorded the key", func(t *testing.T) {
		moved := baseline
		moved.Policy.Telemetry = "off-by-another-manager"
		if _, err := moved.CacheKey(); err == nil {
			t.Fatal("this manager derived a key for a policy outside the fixed go-v1 policy")
		}
		recorded := testRecordedBuild()
		recorded.CacheKey = buildmeta.CacheKey(testDigest(13))
		facts := testFacts(string(install.BuildCacheHit))
		facts.CacheKey = baselineKey

		state, cause, _ := classifyBuildCommand(recorded, markerRecording(recorded), facts)
		if state != buildInputDrift || cause != causeUnattributed {
			t.Fatalf("state = %q cause = %q", state, cause)
		}
	})
}

func TestInputCausesAreDistinctAndDocumented(t *testing.T) {
	seen := map[string]bool{}
	for _, cause := range inputCauses() {
		if cause == "" || seen[cause] {
			t.Fatalf("cause subcode %q is empty or duplicated", cause)
		}
		seen[cause] = true
	}
	documentation, ok := readDocumentation(t)
	if !ok {
		return
	}
	for _, cause := range inputCauses() {
		if !strings.Contains(documentation, "`"+cause+"`") {
			t.Fatalf("README.md does not document the cause subcode %q", cause)
		}
	}
}

func TestClassifySkillBuildsAcceptsOnlyAnExactlyCurrentInstallation(t *testing.T) {
	installed := t.TempDir()
	state, rows := classifySkillBuilds(installed, testRecordedMarker(),
		[]buildFacts{testFacts(string(install.BuildCacheHit))})
	if state != buildCurrent {
		t.Fatalf("state = %q, want %q (rows %+v)", state, buildCurrent, rows)
	}
	if len(rows) != 1 || rows[0].State != buildCurrent || rows[0].CacheOutcome != string(install.BuildCacheHit) {
		t.Fatalf("rows = %+v", rows)
	}
	if rows[0].BuildRoot != "assets/build-tool" || rows[0].SourceDir != "assets/build-tool/cmd/tool" ||
		rows[0].Driver != buildmeta.DriverGoV1 || rows[0].Target != "linux/amd64+GOAMD64=v1" ||
		rows[0].CacheKey != testDigest(2) || rows[0].ArtifactPath != "bin/build-tool" ||
		rows[0].BuildSource != testBuildSource() {
		t.Fatalf("row does not report the full active build command: %+v", rows[0])
	}
	if rows[0].Cause != "" {
		t.Fatalf("a current row carries the cause %q", rows[0].Cause)
	}
}

func TestClassifySkillBuildsWithoutCompiledStateStaysSilent(t *testing.T) {
	recorded := testRecordedMarker()
	recorded.Builds = map[string]marker.Build{}
	recorded.BuildSource = nil
	recorded.BuildRoots = nil
	state, rows := classifySkillBuilds(t.TempDir(), recorded, nil)
	if state != "" || rows != nil {
		t.Fatalf("a runtime-only installation produced compiled diagnostics: %q %+v", state, rows)
	}
}

func TestClassifySkillBuildsDetectsContextExposure(t *testing.T) {
	t.Run("recorded context file inside a build root", func(t *testing.T) {
		recorded := testRecordedMarker()
		recorded.Files = append(recorded.Files, "assets/build-tool/go.mod")
		state, rows := classifySkillBuilds(t.TempDir(), recorded,
			[]buildFacts{testFacts(string(install.BuildCacheHit))})
		if state != buildContextExposed || len(rows) != 1 || rows[0].State != buildContextExposed {
			t.Fatalf("state = %q rows = %+v", state, rows)
		}
	})
	t.Run("build root materialized in agent-facing context", func(t *testing.T) {
		installed := t.TempDir()
		if err := os.MkdirAll(filepath.Join(installed, "assets", "build-tool"), 0o755); err != nil {
			t.Fatal(err)
		}
		state, _ := classifySkillBuilds(installed, testRecordedMarker(),
			[]buildFacts{testFacts(string(install.BuildCacheHit))})
		if state != buildContextExposed {
			t.Fatalf("state = %q, want %q", state, buildContextExposed)
		}
	})
}

func TestClassifySkillBuildsDetectsCommandDriftInBothDirections(t *testing.T) {
	t.Run("closure activates a command the marker does not record", func(t *testing.T) {
		recorded := testRecordedMarker()
		facts := testFacts(string(install.BuildCacheHit))
		extra := facts
		extra.Command = "second-tool"
		state, rows := classifySkillBuilds(t.TempDir(), recorded, []buildFacts{facts, extra})
		if state != buildCommandDrift || len(rows) != 2 {
			t.Fatalf("state = %q rows = %+v", state, rows)
		}
	})
	t.Run("marker records a command the closure no longer activates", func(t *testing.T) {
		recorded := testRecordedMarker()
		recorded.Builds["second-tool"] = testRecordedBuild()
		state, rows := classifySkillBuilds(t.TempDir(), recorded,
			[]buildFacts{testFacts(string(install.BuildCacheHit))})
		if state != buildCommandDrift || len(rows) != 2 {
			t.Fatalf("state = %q rows = %+v", state, rows)
		}
		if rows[1].Command != "second-tool" || rows[1].State != buildCommandDrift {
			t.Fatalf("row for the abandoned command = %+v", rows[1])
		}
	})
}

func TestClassifySkillBuildsRefusesAMarkerThatCannotDescribeABuild(t *testing.T) {
	facts := []buildFacts{testFacts(string(install.BuildCacheHit))}

	state, rows := classifySkillBuilds(t.TempDir(), nil, facts)
	if state != stateInvalidMarker || len(rows) != 1 || rows[0].State != stateInvalidMarker {
		t.Fatalf("unreadable marker: state = %q rows = %+v", state, rows)
	}

	legacy := &marker.Marker{SchemaVersion: marker.LegacySchemaVersion, Name: "build-skill"}
	state, rows = classifySkillBuilds(t.TempDir(), legacy, facts)
	if state != stateNeedsInstall || len(rows) != 1 || rows[0].State != stateNeedsInstall {
		t.Fatalf("legacy marker: state = %q rows = %+v", state, rows)
	}
}

// TestMarkerRefusalSeparatesUnsupportedFromInvalid proves the refusal reasons
// the marker reader collapses into "no marker" stay distinguishable to an
// operator: a document from a newer manager, a build driver outside the closed
// set, and a document that is simply not a marker are three different codes.
func TestMarkerRefusalSeparatesUnsupportedFromInvalid(t *testing.T) {
	for name, testCase := range map[string]struct {
		payload string
		absent  bool
		want    string
	}{
		"no marker at all": {absent: true, want: stateInvalidMarker},
		"not a JSON document": {
			payload: "not a marker", want: stateInvalidMarker,
		},
		"no schema version": {
			payload: `{"name":"build-skill"}`, want: stateInvalidMarker,
		},
		"schema from a newer manager": {
			payload: `{"schema_version":3,"name":"build-skill"}`, want: stateUnsupportedMarker,
		},
		"schema below the oldest readable": {
			payload: `{"schema_version":0,"name":"build-skill"}`, want: stateUnsupportedMarker,
		},
		"build driver outside the closed set": {
			payload: `{"schema_version":2,"name":"build-skill",` +
				`"builds":{"build-tool":{"driver":"go-v2"}}}`,
			want: buildUnsupportedDriver,
		},
		"readable schema that is still not a valid marker": {
			payload: `{"schema_version":2,"name":"build-skill"}`, want: stateInvalidMarker,
		},
	} {
		t.Run(name, func(t *testing.T) {
			installed := t.TempDir()
			if !testCase.absent {
				if err := os.WriteFile(filepath.Join(installed, marker.Name),
					[]byte(testCase.payload), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			// Every payload here must still be refused by the reader itself: this
			// classification only explains a refusal, it never admits one.
			if marker.Read(installed) != nil {
				t.Fatalf("the marker reader accepted %q, so the case proves nothing", testCase.payload)
			}
			state, detail := markerRefusal(installed)
			if state != testCase.want {
				t.Fatalf("state = %q, want %q (detail %q)", state, testCase.want, detail)
			}
			if detail == "" {
				t.Fatal("a refusal produced no operator detail")
			}
		})
	}
}

func TestCurrentnessCodesAreDistinctAndOnlyExactStatesPass(t *testing.T) {
	seen := map[string]bool{}
	for _, code := range currentnessCodes() {
		if code == "" {
			t.Fatal("an empty currentness code is not machine-readable")
		}
		if seen[code] {
			t.Fatalf("duplicate currentness code %q", code)
		}
		seen[code] = true
	}
	for _, code := range currentnessCodes() {
		want := code == stateUpToDate || code == buildCurrent
		if currentCode(code) != want {
			t.Fatalf("currentCode(%q) = %v, want %v", code, currentCode(code), want)
		}
	}
}

// readDocumentation loads README.md, or reports that it is out of reach.
//
// A cross-compiled test binary is run outside the checkout on hosts without a
// Go toolchain. A guard that belongs to the repository states plainly that it
// did not run rather than failing for the absence of a tree it never had.
func readDocumentation(t *testing.T) (string, bool) {
	t.Helper()
	documentationPath := filepath.Join("..", "..", "README.md")
	payload, err := os.ReadFile(documentationPath) // #nosec G304 -- fixed repository path
	if os.IsNotExist(err) {
		t.Skipf("%s is not reachable from this test binary", documentationPath)
		return "", false
	}
	if err != nil {
		t.Fatal(err)
	}
	return string(payload), true
}

// TestEveryCurrentnessCodeIsDocumented keeps the operator documentation and
// the reachable vocabulary in sync: a code nobody can look up is not a stable
// interface.
func TestEveryCurrentnessCodeIsDocumented(t *testing.T) {
	documentation, ok := readDocumentation(t)
	if !ok {
		return
	}
	for _, code := range currentnessCodes() {
		if !strings.Contains(documentation, "`"+code+"`") {
			t.Fatalf("README.md does not document the currentness code %q", code)
		}
	}
}

// TestCheckFailsForEveryNonCurrentCode proves `status --check` exits zero only
// for exactly current state, on both surfaces it consults: the declared-skill
// map and the compiled-command rows a transitively resolved node produces.
func TestCheckFailsForEveryNonCurrentCode(t *testing.T) {
	for _, code := range currentnessCodes() {
		want := !currentCode(code)
		if failed := checkFailed(map[string]string{"skill-a": code}, nil); failed != want {
			t.Fatalf("checkFailed for skill code %q = %v, want %v", code, failed, want)
		}
		rows := []buildReport{{Skill: "build-skill", Command: "build-tool", State: code}}
		if failed := checkFailed(map[string]string{}, rows); failed != want {
			t.Fatalf("checkFailed for build code %q = %v, want %v", code, failed, want)
		}
	}
	if checkFailed(map[string]string{"skill-a": stateUpToDate},
		[]buildReport{{State: buildCurrent}}) {
		t.Fatal("an exactly current scope failed --check")
	}
	if checkFailed(nil, nil) {
		t.Fatal("an empty scope failed --check")
	}
}

// TestBuildReportsNeverPublishAnAbsolutePath proves an untrusted cache reason
// cannot publish a manager-private location or drive a terminal through a
// machine-readable diagnostic field, in free-standing, embedded, and URI form.
func TestBuildReportsNeverPublishAnAbsolutePath(t *testing.T) {
	for _, reason := range []string{
		"inspect cache boundary: lstat /Users/operator/.curator/cache/build/go-v1/abcd: permission denied",
		`open C:\Users\operator\AppData\curator\cache: access denied`,
		`entry \\fileserver\share\curator is not protected`,
		"artifact hash mismatch\r\n\x1b[31mnot really an error\x1b[0m",
		"source=/Users/operator/.curator/cache/build/go-v1/abcd is unreadable",
		"receipt at file:///Users/operator/.curator/cache/build/go-v1/abcd failed",
		`error=C:\Users\operator\cache and error=\\fileserver\share\curator both failed`,
	} {
		facts := testFacts(string(install.BuildWouldRebuildUntrustedCache))
		facts.Reason = reason
		state, cause, detail := classifyBuildCommand(testRecordedBuild(), testRecordedMarker(), facts)
		if state != buildUntrustedCache {
			t.Fatalf("state = %q", state)
		}
		row := report(facts, state, cause, detail)
		for _, forbidden := range []string{"/Users/", `C:\`, `\\fileserver`, "operator", "\n", "\r", "\x1b"} {
			if strings.Contains(row.Detail, forbidden) {
				t.Fatalf("detail %q leaked %q", row.Detail, forbidden)
			}
		}
		if !strings.Contains(row.Detail, "<path>") && strings.Contains(reason, "/") {
			t.Fatalf("detail %q did not redact the location in %q", row.Detail, reason)
		}
	}
}

func TestBuildReportDetailIsBounded(t *testing.T) {
	facts := testFacts(string(install.BuildCorrupt))
	facts.Reason = strings.Repeat("compiler noise ", 500)
	state, cause, detail := classifyBuildCommand(testRecordedBuild(), testRecordedMarker(), facts)
	row := report(facts, state, cause, detail)
	if len([]rune(row.Detail)) > 240 {
		t.Fatalf("detail is %d runes, want at most 240", len([]rune(row.Detail)))
	}
}

// TestGoToolchainGuidanceNamesTheAcceptedSelectionAndTestedFamilies pins the
// operator guidance to the accepted selection mechanisms and the tested Go
// release families, and proves it never suggests a PATH lookup or a download.
func TestGoToolchainGuidanceNamesTheAcceptedSelectionAndTestedFamilies(t *testing.T) {
	if guidance := goToolchainGuidance(""); guidance != "" {
		t.Fatalf("a run without a driver diagnostic produced guidance %q", guidance)
	}
	families := godriver.TestedFamilies()
	if len(families) == 0 {
		t.Fatal("the driver reports no tested Go release family")
	}
	for _, code := range []string{
		"go_toolchain_missing", "untrusted_go_executable", "toolchain_executable_mismatch",
		"toolchain_mutated", "unsupported_go_family", "malformed_go_version", "go_build_failed",
	} {
		guidance := goToolchainGuidance(code)
		for _, want := range append([]string{
			code, godriver.SelectionCuratorGo, godriver.SelectionGOROOT,
		}, families...) {
			if !strings.Contains(guidance, want) {
				t.Fatalf("guidance for %q does not name %q: %s", code, want, guidance)
			}
		}
		const closedRule = "Curator never searches PATH and never downloads a toolchain"
		if !strings.Contains(guidance, closedRule) {
			t.Fatalf("guidance for %q does not state the closed selection rule: %s", code, guidance)
		}
		// Outside the one sentence that denies them, neither a PATH lookup nor
		// a download may be mentioned at all, so no remedy can be read as an
		// accepted selection mechanism.
		remainder := strings.ToLower(strings.ReplaceAll(guidance, closedRule, ""))
		for _, forbidden := range []string{
			"path", "download", "fetch", "http", "brew", "curl", "apt-get", "winget", "choco", "asdf", "gvm",
		} {
			if strings.Contains(remainder, forbidden) {
				t.Fatalf("guidance for %q suggests %q: %s", code, forbidden, guidance)
			}
		}
	}
}

func TestRepairNoticesDistinguishRepairFromAPreservedInstallation(t *testing.T) {
	hit := testFacts(string(install.BuildCacheHit))

	for name, testCase := range map[string]struct {
		facts    buildFacts
		status   string
		retained bool
		want     string
		absent   string
		noNotic  bool
	}{
		"untrusted state was rebuilt": {
			facts: testFacts(string(install.BuildWouldRebuildUntrustedCache)), status: "ok",
			want: "rebuilt untrusted build cache state",
		},
		"untrusted state was not repaired": {
			facts: testFacts(string(install.BuildWouldRebuildUntrustedCache)), status: "failed",
			want: "did not repair untrusted build cache state",
		},
		"corrupt state was rebuilt": {
			facts: testFacts(string(install.BuildCorrupt)), status: "ok",
			want: "rebuilt corrupt build cache state", absent: "before any mutation",
		},
		"corrupt state was not repaired": {
			facts: testFacts(string(install.BuildCorrupt)), status: "failed",
			want: "did not repair corrupt build cache state",
		},
		// A failed run that could not withdraw what it rebuilt must never repeat
		// the ordinary "the live build cache is unchanged" claim.
		"a rebuilt entry could not be withdrawn": {
			facts: testFacts(string(install.BuildCorrupt)), status: "failed", retained: true,
			want:   "did not finish repairing corrupt build cache state",
			absent: "the live build cache are unchanged",
		},
		"an unprotectable host is refused": {
			facts: testFacts(string(install.BuildUnsupported)), status: "failed",
			want: "before any mutation",
		},
		"an unusable toolchain is refused": {
			facts: testFacts(string(install.BuildToolchainUnavailable)), status: "failed",
			want: "before any mutation",
		},
	} {
		t.Run(name, func(t *testing.T) {
			notices := repairNotices("app", []buildFacts{testCase.facts}, testCase.status, testCase.retained)
			if len(notices) != 1 || !strings.Contains(notices[0], testCase.want) {
				t.Fatalf("notices = %v, want one containing %q", notices, testCase.want)
			}
			if testCase.absent != "" && strings.Contains(notices[0], testCase.absent) {
				t.Fatalf("notice %q still claims %q", notices[0], testCase.absent)
			}
		})
	}
	if notices := repairNotices("app", []buildFacts{hit}, "ok", false); len(notices) != 0 {
		t.Fatalf("a cache hit produced repair notices: %v", notices)
	}
}

func TestMarkerDigestsDetectAConcurrentInstallMarkerChange(t *testing.T) {
	store := t.TempDir()
	installed := filepath.Join(store, "build-skill")
	if err := os.MkdirAll(installed, 0o755); err != nil {
		t.Fatal(err)
	}
	markerPath := filepath.Join(installed, marker.Name)
	if err := os.WriteFile(markerPath, []byte(`{"schema_version":2}`), 0o644); err != nil {
		t.Fatal(err)
	}
	before := markerDigests(store)
	if markerMoved(before, markerDigests(store), installed) {
		t.Fatal("an unchanged marker was reported as moved")
	}
	if err := os.WriteFile(markerPath, []byte(`{"schema_version":2,"name":"x"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if !markerMoved(before, markerDigests(store), installed) {
		t.Fatal("a rewritten marker was not reported as moved")
	}
	if err := os.Remove(markerPath); err != nil {
		t.Fatal(err)
	}
	if !markerMoved(before, markerDigests(store), installed) {
		t.Fatal("a removed marker was not reported as moved")
	}
}

// TestStatusReportMarksCompiledStateThatMovedDuringTheCheck proves the whole
// classification window is guarded: a marker that differs from the fingerprint
// the run started with makes every verdict for that skill stale rather than
// authoritative, and the scope fails --check.
func TestStatusReportMarksCompiledStateThatMovedDuringTheCheck(t *testing.T) {
	project := t.TempDir()
	cfg := &config.Config{Path: filepath.Join(t.TempDir(), "home", "config.json"), SkillsRoot: t.TempDir()}
	installed := filepath.Join(project, ".agents", "skills", "build-skill")
	writeCompiledMarker(t, installed)
	facts := []buildFacts{testFacts(string(install.BuildCacheHit))}

	settled := markerDigests(statusStores(cfg, project)...)
	_, rows := statusReport(cfg, project, facts, settled)
	if len(rows) != 1 || rows[0].State != buildCurrent {
		t.Fatalf("settled compiled state = %+v", rows)
	}

	// The marker really does move between the two fingerprints, exactly as a
	// concurrent install would move it.
	rewriteMarker(t, installed, func(object map[string]any) { object["ref"] = "v2" })
	drift, rows := statusReport(cfg, project, facts, settled)
	if len(rows) != 1 || rows[0].State != buildStateChanged {
		t.Fatalf("moved compiled state = %+v", rows)
	}
	if !checkFailed(drift, rows) {
		t.Fatal("compiled state that moved during the check passed --check")
	}
}

// TestStatusReportReportsCompiledCommandsOfAnUninstalledSkill proves a build
// planned for a node with no installed context is reported instead of silently
// dropped, which is what a transitively resolved provider looks like before its
// first installation.
func TestStatusReportReportsCompiledCommandsOfAnUninstalledSkill(t *testing.T) {
	project := t.TempDir()
	cfg := &config.Config{Path: filepath.Join(t.TempDir(), "home", "config.json"), SkillsRoot: t.TempDir()}
	facts := []buildFacts{testFacts(string(install.BuildCacheHit))}

	drift, rows := statusReport(cfg, project, facts, map[string]string{})
	if len(rows) != 1 || rows[0].State != stateNotInstalled {
		t.Fatalf("rows = %+v", rows)
	}
	if !checkFailed(drift, rows) {
		t.Fatal("a compiled command of an uninstalled skill passed --check")
	}
}

// writeCompiledMarker installs a valid schema 2 marker that records exactly the
// compiled state testFacts derives.
func writeCompiledMarker(t *testing.T, installed string) {
	t.Helper()
	if err := os.MkdirAll(installed, 0o755); err != nil {
		t.Fatal(err)
	}
	contentHash, err := hashing.ContentSHA256(installed, nil)
	if err != nil {
		t.Fatal(err)
	}
	source := testBuildSource()
	if err := marker.Write(installed, &marker.Marker{
		Name: "build-skill", Source: "build-skill", RefKind: "tag", Ref: "v1",
		Commit: strings.Repeat("a", 40), ContentSHA256: contentHash,
		Agents: []string{}, Commands: []string{"build-tool"}, Dependencies: []string{},
		SkillSchemaVersion: 6, RuntimeRoots: []string{}, BuildRoots: []string{"assets/build-tool"},
		BuildSource: &source, Builds: map[string]marker.Build{"build-tool": testRecordedBuild()},
		Files: []string{}, InstalledAt: "2026-07-20T00:00:00Z",
	}); err != nil {
		t.Fatal(err)
	}
}
