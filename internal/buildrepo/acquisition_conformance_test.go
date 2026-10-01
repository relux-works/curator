package buildrepo

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
	"github.com/relux-works/curator/internal/stateread"
)

const (
	acquisitionVectorPath = "vectors/external-repository-acquisition.json"
	acquisitionManifestID = "be11bb1e4c46f21fb5684d586f9c2a8b0d59f3b437bc7ea7aa5aa530fe4d47ca"
)

type externalRepositoryAcquisitionVector struct {
	ProtocolVersion        string                              `json:"protocol_version"`
	Cases                  []externalRepositoryAcquisitionCase `json:"cases"`
	CommonFetchArgv        []string                            `json:"common_fetch_argv"`
	CleanEnvironment       []string                            `json:"clean_environment"`
	ForbiddenFetchFeatures []string                            `json:"forbidden_fetch_features"`
}

type externalRepositoryAcquisitionCase struct {
	Name                    string `json:"name"`
	ObjectFormat            string `json:"object_format"`
	Transport               string `json:"transport"`
	FetchRefspec            string `json:"fetch_refspec"`
	Result                  string `json:"result"`
	ExpectedError           string `json:"expected_error"`
	Ref                     string `json:"ref"`
	DirectOIDFetchAttempted bool   `json:"direct_oid_fetch_attempted"`
	GitStarted              *bool  `json:"git_started"`
	AuditBeforeCache        bool   `json:"audit_before_cache"`
	AuditBeforeCompiler     bool   `json:"audit_before_compiler"`
	AuditStarted            bool   `json:"audit_started"`
	ArtifactCacheLookup     bool   `json:"artifact_cache_lookup"`
	CompilerStarted         bool   `json:"compiler_started"`
}

func TestExternalRepositoryAcquisitionConformance(t *testing.T) {
	vector := loadExternalRepositoryAcquisitionVector(t)
	if vector.ProtocolVersion != "1.0.0-rc.13" {
		t.Fatalf("acquisition vector protocol_version = %q, want 1.0.0-rc.13", vector.ProtocolVersion)
	}
	manifestID, err := conformancecoverage.SelectedSuiteManifestSHA256()
	if err != nil {
		t.Fatal(err)
	}
	if manifestID != acquisitionManifestID {
		t.Fatalf("selected conformance manifest = %s, want pinned rc.13 %s", manifestID, acquisitionManifestID)
	}

	t.Run("cases", func(t *testing.T) {
		testExternalRepositoryAcquisitionCases(t, vector)
	})
	t.Run("common-fetch-argv", func(t *testing.T) {
		testExternalRepositoryAcquisitionArgv(t, vector.CommonFetchArgv)
	})
	t.Run("clean-environment", func(t *testing.T) {
		testExternalRepositoryCleanEnvironment(t, vector.CleanEnvironment)
	})
	t.Run("forbidden-fetch-features", func(t *testing.T) {
		testExternalRepositoryForbiddenFeatures(t, vector.ForbiddenFetchFeatures)
	})
}

func loadExternalRepositoryAcquisitionVector(t *testing.T) externalRepositoryAcquisitionVector {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set; rc.13 acquisition vectors were not selected")
	}
	path := filepath.Join(root, filepath.FromSlash(acquisitionVectorPath))
	file, err := stateread.ReadRegularFile(path)
	if err != nil {
		t.Fatalf("read pinned acquisition vector: %v", err)
	}
	if file.Kind != stateread.KindPresent {
		t.Fatalf("pinned acquisition vector state = %q, want present", file.Kind)
	}
	var vector externalRepositoryAcquisitionVector
	if err := json.Unmarshal(file.Bytes, &vector); err != nil {
		t.Fatalf("decode pinned acquisition vector: %v", err)
	}
	return vector
}

type acquisitionArgRow struct {
	Index int
	Value string
}

func testExternalRepositoryAcquisitionArgv(t *testing.T, published []string) {
	t.Helper()
	tool := realGitTool(t)
	paths, err := makePrivatePaths(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	tool.AskPass = filepath.Join(paths.root, "manager-askpass")
	transport := "https"
	source := "https://fixture.test/repository.git"
	refspec := "0123456789abcdef0123456789abcdef01234567:refs/curator/locked"
	want := make([]string, len(published))
	rows := make([]acquisitionArgRow, len(published))
	for index, value := range published {
		rows[index] = acquisitionArgRow{Index: index, Value: value}
		want[index] = resolveAcquisitionArg(value, paths, tool, transport, source, refspec)
	}
	got := append([]string{tool.Executable}, strictFetchArgs(paths.repo, paths.hooks, tool.AskPass, transport, source, refspec)...)
	conformancecoverage.RunOutcomes(t, "external-repository/acquisition/common-fetch-argv", rows,
		func(row acquisitionArgRow) string { return fmt.Sprintf("arg-%02d", row.Index) },
		func(_ *testing.T, row acquisitionArgRow) conformancecoverage.Observation {
			if row.Index >= len(got) {
				return conformancecoverage.Observation{FailureReason: fmt.Sprintf("argv row %d is absent", row.Index)}
			}
			if got[row.Index] != want[row.Index] {
				return conformancecoverage.Observation{FailureReason: fmt.Sprintf("argv[%d] = %q, vector resolves to %q", row.Index, got[row.Index], want[row.Index])}
			}
			if row.Index == len(rows)-1 && len(got) != len(want) {
				return conformancecoverage.Observation{FailureReason: fmt.Sprintf("fetch argv has %d entries, vector pins %d", len(got), len(want))}
			}
			return conformancecoverage.Observation{}
		})
}

func resolveAcquisitionArg(value string, paths privatePaths, tool GitTool, transport, source, refspec string) string {
	switch value {
	case "<absolute-trusted-git>":
		return tool.Executable
	case "--git-dir=<operation-private>/repo.git":
		return "--git-dir=" + paths.repo
	case "protocol.<selected>.allow=always":
		return "protocol." + transport + ".allow=always"
	case "core.askPass=<manager-broker>":
		return "core.askPass=" + tool.AskPass
	case "core.hooksPath=<operation-private>/empty-hooks":
		return "core.hooksPath=" + paths.hooks
	case "<validated-url>":
		return source
	case "<one-manager-refspec>":
		return refspec
	default:
		return value
	}
}

func testExternalRepositoryCleanEnvironment(t *testing.T, published []string) {
	t.Helper()
	tool := realGitTool(t)
	paths, err := makePrivatePaths(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	// The vector pins the common environment. The same production builder adds
	// only the exact transport broker variables for an HTTPS or SSH fetch.
	tool.AskPass = ""
	tool.SSHWrapper = ""
	gotEntries := cleanGitEnvironment(paths, tool, "")
	got, parseErr := environmentByName(gotEntries)
	want := expectedAcquisitionEnvironment(t, published, paths, tool, "")
	rows := make([]environmentRow, len(published))
	for index, value := range published {
		name, _, ok := strings.Cut(value, "=")
		if !ok {
			t.Fatalf("published clean-environment row %q has no equals sign", value)
		}
		rows[index] = environmentRow{Index: index, Name: name}
	}
	var unexpected []string
	for name := range got {
		if _, expected := want[name]; !expected {
			unexpected = append(unexpected, name)
		}
	}
	sort.Strings(unexpected)
	conformancecoverage.RunOutcomes(t, "external-repository/acquisition/clean-environment", rows,
		func(row environmentRow) string { return row.Name },
		func(_ *testing.T, row environmentRow) conformancecoverage.Observation {
			if parseErr != nil {
				return conformancecoverage.Observation{FailureReason: parseErr.Error()}
			}
			wantValue, exists := want[row.Name]
			gotValue, present := got[row.Name]
			if !exists || !present || gotValue != wantValue {
				return conformancecoverage.Observation{FailureReason: fmt.Sprintf("environment %s = %q, vector resolves to %q", row.Name, gotValue, wantValue)}
			}
			if row.Index == len(rows)-1 && len(unexpected) != 0 {
				return conformancecoverage.Observation{FailureReason: "unexpected environment entries: " + strings.Join(unexpected, ", ")}
			}
			if row.Index == len(rows)-1 && len(got) != len(want) {
				return conformancecoverage.Observation{FailureReason: fmt.Sprintf("clean environment has %d entries, vector plus platform allowlist pins %d", len(got), len(want))}
			}
			return conformancecoverage.Observation{}
		})
	assertTransportEnvironment := func(transport string, additions map[string]string) {
		t.Helper()
		transportTool := tool
		transportTool.AskPass = filepath.Join(paths.root, "manager-askpass")
		transportTool.SSHWrapper = filepath.Join(paths.root, "manager-ssh-wrapper")
		transportEntries, err := environmentByName(cleanGitEnvironment(paths, transportTool, transport))
		if err != nil {
			t.Fatal(err)
		}
		wantTransport := make(map[string]string, len(got)+len(additions))
		for name, value := range got {
			wantTransport[name] = value
		}
		for name, value := range additions {
			wantTransport[name] = value
		}
		if !reflect.DeepEqual(transportEntries, wantTransport) {
			t.Fatalf("%s clean environment = %#v, want vector base plus manager variables %#v", transport, transportEntries, wantTransport)
		}
	}
	assertTransportEnvironment("https", map[string]string{"GIT_ASKPASS": filepath.Join(paths.root, "manager-askpass")})
	assertTransportEnvironment("ssh", map[string]string{
		"GIT_SSH":         filepath.Join(paths.root, "manager-ssh-wrapper"),
		"GIT_SSH_VARIANT": "ssh",
	})
}

func expectedAcquisitionEnvironment(t *testing.T, published []string, paths privatePaths, tool GitTool, transport string) map[string]string {
	t.Helper()
	entries := make([]string, len(published))
	for index, value := range published {
		entries[index] = resolveAcquisitionEnvironment(value, paths, tool)
	}
	want, err := environmentByName(entries)
	if err != nil {
		t.Fatalf("invalid published clean environment: %v", err)
	}
	for _, name := range []string{"SYSTEMROOT", "WINDIR", "COMSPEC", "PATHEXT"} {
		if value := os.Getenv(name); value != "" {
			want[name] = value
		}
	}
	switch transport {
	case "https":
		want["GIT_ASKPASS"] = tool.AskPass
	case "ssh":
		want["GIT_SSH"] = tool.SSHWrapper
		want["GIT_SSH_VARIANT"] = "ssh"
	}
	return want
}

type environmentRow struct {
	Index int
	Name  string
}

func resolveAcquisitionEnvironment(value string, paths privatePaths, tool GitTool) string {
	name, setting, ok := strings.Cut(value, "=")
	if !ok {
		return value
	}
	switch name {
	case "GIT_CONFIG_GLOBAL":
		setting = filepath.Join(paths.root, "global.gitconfig")
	case "GIT_CONFIG_SYSTEM":
		setting = filepath.Join(paths.root, "system.gitconfig")
	case "GIT_EXEC_PATH":
		setting = tool.ExecPath
	case "HOME":
		setting = paths.home
	case "PATH":
		setting = paths.path
	case "XDG_CONFIG_HOME":
		setting = paths.config
	}
	return name + "=" + setting
}

func environmentByName(entries []string) (map[string]string, error) {
	values := make(map[string]string, len(entries))
	for _, entry := range entries {
		name, value, ok := strings.Cut(entry, "=")
		if !ok || name == "" {
			return nil, fmt.Errorf("malformed clean-environment entry %q", entry)
		}
		if _, exists := values[name]; exists {
			return nil, fmt.Errorf("duplicate clean-environment variable %s", name)
		}
		values[name] = value
	}
	return values, nil
}

func testExternalRepositoryForbiddenFeatures(t *testing.T, published []string) {
	t.Helper()
	tool := realGitTool(t)
	paths, err := makePrivatePaths(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	tool.AskPass = filepath.Join(paths.root, "manager-askpass")
	transport := "https"
	source := "https://fixture.test/repository.git"
	refspec := "0123456789abcdef0123456789abcdef01234567:refs/curator/locked"
	args := strictFetchArgs(paths.repo, paths.hooks, tool.AskPass, transport, source, refspec)
	conformancecoverage.RunOutcomes(t, "external-repository/acquisition/forbidden-fetch-features", published,
		func(feature string) string { return feature },
		func(_ *testing.T, feature string) conformancecoverage.Observation {
			if feature == "helper-selected-transport" {
				if _, err := ParseSource("ext::git-upload-pack fixture.test/repository.git"); err == nil {
					return conformancecoverage.Observation{FailureReason: "source parser accepts helper-selected transport"}
				}
			}
			accepted, recognized := forbiddenFetchFeatureAccepted(feature, args, source, transport)
			if !recognized {
				return conformancecoverage.Observation{FailureReason: "published forbidden feature has no refusal assertion"}
			}
			if accepted {
				return conformancecoverage.Observation{FailureReason: "fetch argv admits forbidden feature " + feature}
			}
			return conformancecoverage.Observation{}
		})
}

func forbiddenFetchFeatureAccepted(feature string, args []string, source, transport string) (bool, bool) {
	contains := func(want string) bool {
		for _, arg := range args {
			if arg == want {
				return true
			}
		}
		return false
	}
	containsPrefix := func(prefix string) bool {
		for _, arg := range args {
			if strings.HasPrefix(arg, prefix) {
				return true
			}
		}
		return false
	}
	containsConfig := func(prefix string) bool {
		for index := 0; index+1 < len(args); index++ {
			if args[index] == "-c" && strings.HasPrefix(args[index+1], prefix) {
				return true
			}
		}
		return false
	}
	selectedSource := func() string {
		separator := -1
		for index, arg := range args {
			if arg == "--" {
				separator = index
				break
			}
		}
		if separator < 0 || separator+1 >= len(args) {
			return ""
		}
		return args[separator+1]
	}

	switch feature {
	case "configured-refspec":
		return !contains("--refmap=") || !contains("--"), true
	case "depth":
		return containsPrefix("--depth") || containsConfig("fetch.depth="), true
	case "filter":
		return containsPrefix("--filter") || containsConfig("fetch.filter=") || containsConfig("extensions.partialclone="), true
	case "helper-selected-transport":
		return !contains("protocol.allow=never") || !contains("protocol."+transport+".allow=always"), true
	case "mirror":
		return contains("--mirror") || containsConfig("remote.") && containsConfig(".mirror=true"), true
	case "prune":
		return contains("--prune") || contains("--prune-tags") || containsConfig("fetch.prune=") || containsConfig("remote.") && containsConfig(".prune=true"), true
	case "remote-name":
		return selectedSource() != source, true
	case "server-option":
		return containsPrefix("--server-option") || containsConfig("protocol.") && containsConfig(".serveroption="), true
	case "source-upload-pack":
		count := 0
		for _, arg := range args {
			if strings.HasPrefix(arg, "--upload-pack=") {
				count++
				if arg != "--upload-pack=git-upload-pack" {
					return true, true
				}
			}
		}
		return count != 1, true
	case "stdin-refspec":
		return contains("--stdin"), true
	case "tag-auto-follow":
		return !contains("--no-tags"), true
	default:
		return false, false
	}
}

type acquisitionGitShim struct {
	Executable string
	RealGit    string
	Tool       GitTool
}

type acquisitionGitShimConfig struct {
	Git        string `json:"git"`
	Transport  string `json:"transport"`
	Source     string `json:"source"`
	Repository string `json:"repository"`
	LogPath    string `json:"log_path"`
}

type acquisitionGitShimInvocation struct {
	Args        []string `json:"args"`
	Environment []string `json:"environment"`
}

func newAcquisitionGitShim(t *testing.T) acquisitionGitShim {
	t.Helper()
	tool := realGitTool(t)
	realGit := tool.Executable
	name := "git-shim"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	executable := filepath.Join(t.TempDir(), name)
	command := exec.Command("go", "build", "-o", executable, "./testdata/acquisitiongitshim")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("build acquisition Git shim: %v\n%s", err, output)
	}
	tool.Executable = executable
	tool.AskPass = executable
	tool.SSHWrapper = executable
	tool.SSHCredentials = OperatorSSHCredentials{Identity: realGit, KnownHosts: realGit}
	return acquisitionGitShim{Executable: executable, RealGit: realGit, Tool: tool}
}

func (shim acquisitionGitShim) configure(t *testing.T, transport, source, repository string) string {
	t.Helper()
	logPath := filepath.Join(t.TempDir(), "git-invocations.jsonl")
	config := acquisitionGitShimConfig{Git: shim.RealGit, Transport: transport, Source: source, Repository: repository, LogPath: logPath}
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(shim.Executable+".json", data, 0o600); err != nil {
		t.Fatal(err)
	}
	return logPath
}

func testExternalRepositoryAcquisitionCases(t *testing.T, vector externalRepositoryAcquisitionVector) {
	t.Helper()
	shim := newAcquisitionGitShim(t)
	conformancecoverage.RunOutcomes(t, "external-repository/acquisition/cases", vector.Cases,
		func(testCase externalRepositoryAcquisitionCase) string { return testCase.Name },
		func(caseT *testing.T, testCase externalRepositoryAcquisitionCase) conformancecoverage.Observation {
			request, fixture, source, expectedRefspec := acquisitionRequestForCase(caseT, testCase, shim)
			logPath := shim.configure(caseT, request.Source.Transport, source, func() string {
				if fixture == nil {
					return ""
				}
				return fixture.bare
			}())

			var snapshot *Snapshot
			var acquisitionErr error
			var pipelineErr error
			var phases []string
			auditStarted := false
			if testCase.ExpectedError != "" {
				_, pipelineErr = RunPipeline(context.Background(), PipelineRequest{
					Operation: OperationInstall,
					Declared:  DeclaredState{Tag: request.Tag},
					Acquire: func(ctx context.Context) (*Snapshot, error) {
						snapshot, acquisitionErr = AcquireNetwork(ctx, request)
						return snapshot, acquisitionErr
					},
					Audit: func(context.Context, AuditSubject) error {
						auditStarted = true
						return nil
					},
					Trace: func(phase string) { phases = append(phases, phase) },
				})
			} else {
				snapshot, acquisitionErr = AcquireNetwork(context.Background(), request)
			}

			if testCase.ExpectedError != "" {
				if acquisitionErr == nil || ErrorCode(acquisitionErr) != testCase.ExpectedError {
					return conformancecoverage.Observation{FailureReason: fmt.Sprintf("acquisition error = %v (%s), vector requires %s", acquisitionErr, ErrorCode(acquisitionErr), testCase.ExpectedError)}
				}
				if pipelineErr == nil {
					return conformancecoverage.Observation{FailureReason: "pipeline did not stop after acquisition failure"}
				}
				if auditStarted || testCase.AuditStarted || testCase.ArtifactCacheLookup || testCase.CompilerStarted || !reflect.DeepEqual(phases, []string{"exact-source-acquisition"}) {
					return conformancecoverage.Observation{FailureReason: fmt.Sprintf("downstream pipeline work started: audit=%v phases=%v", auditStarted, phases)}
				}
			} else {
				if acquisitionErr != nil || snapshot == nil {
					return conformancecoverage.Observation{FailureReason: fmt.Sprintf("source resolution failed: %v", acquisitionErr)}
				}
				if snapshot.ObjectFormat != request.Lock.ObjectFormat || fixture == nil || snapshot.Commit != fixture.commit {
					return conformancecoverage.Observation{FailureReason: fmt.Sprintf("resolved snapshot identity = %s/%s, fixture requires %s/%s", snapshot.ObjectFormat, snapshot.Commit, request.Lock.ObjectFormat, fixture.commit)}
				}
				// Bound: these raw-object fixtures have no build descriptor, so
				// successful pipeline audit/cache/compiler order is not observed.
				// The case is driven for acquisition only; see task results R-b.
			}

			invocations, err := readAcquisitionGitShimInvocations(logPath)
			if err != nil {
				return conformancecoverage.Observation{FailureReason: err.Error()}
			}
			if testCase.GitStarted != nil && *testCase.GitStarted != (len(invocations) > 0) {
				return conformancecoverage.Observation{FailureReason: fmt.Sprintf("git_started = %v, vector requires %v", len(invocations) > 0, *testCase.GitStarted)}
			}
			if testCase.Name == "malformed-ref-rejected-before-git" {
				if len(invocations) != 0 {
					return conformancecoverage.Observation{FailureReason: fmt.Sprintf("malformed ref started Git: %v", invocations)}
				}
				return conformancecoverage.Observation{}
			}
			fetches := fetchInvocations(invocations)
			if len(fetches) != 1 {
				return conformancecoverage.Observation{FailureReason: fmt.Sprintf("observed %d fetch invocations, want one exact-ref fetch", len(fetches))}
			}
			if err := verifyAcquisitionFetchCallSite(caseT, fetches[0], vector, request, expectedRefspec); err != nil {
				return conformancecoverage.Observation{FailureReason: err.Error()}
			}
			actualSource, actualRefspec, ok := sourceAndRefspec(fetches[0].Args)
			if !ok || actualSource != request.Source.Git || actualRefspec != expectedRefspec {
				return conformancecoverage.Observation{FailureReason: fmt.Sprintf("fetch source/refspec = %q:%q, vector resolves to %q:%q", actualSource, actualRefspec, request.Source.Git, expectedRefspec)}
			}
			exactTagPath := request.Tag != "" || request.RefKind == "tag"
			if exactTagPath && !testCase.DirectOIDFetchAttempted && strings.HasPrefix(actualRefspec, request.Lock.Hex+":") {
				return conformancecoverage.Observation{FailureReason: "an exact lock OID fetch was attempted despite direct_oid_fetch_attempted=false"}
			}
			return conformancecoverage.Observation{}
		})
}

// verifyAcquisitionFetchCallSite checks the invocation recorded at the shim's
// entry, before it rewrites the test transport. Expected flags and environment
// come from the vector, never from the production builders under test.
func verifyAcquisitionFetchCallSite(t *testing.T, invocation acquisitionGitShimInvocation, vector externalRepositoryAcquisitionVector, request NetworkRequest, refspec string) error {
	t.Helper()
	var repo, hooks string
	for index, arg := range invocation.Args {
		if strings.HasPrefix(arg, "--git-dir=") {
			if repo != "" {
				return fmt.Errorf("fetch has multiple private repository paths")
			}
			repo = strings.TrimPrefix(arg, "--git-dir=")
		}
		if arg == "-c" && index+1 < len(invocation.Args) && strings.HasPrefix(invocation.Args[index+1], "core.hooksPath=") {
			if hooks != "" {
				return fmt.Errorf("fetch has multiple private hooks paths")
			}
			hooks = strings.TrimPrefix(invocation.Args[index+1], "core.hooksPath=")
		}
	}
	root := filepath.Dir(repo)
	if !filepath.IsAbs(root) || !strings.HasPrefix(filepath.Base(root), "curator-buildrepo-") ||
		repo != filepath.Join(root, "repo.git") || hooks != filepath.Join(root, "empty-hooks") {
		return fmt.Errorf("fetch repo %q and hooks %q must share one operation-private root", repo, hooks)
	}
	// Acquisition has already removed this root. Resolve names without creating
	// or reading any new state, and do not infer expected values from logged env.
	paths := privatePaths{root: root, repo: repo, hooks: hooks,
		home: filepath.Join(root, "home"), config: filepath.Join(root, "config"), path: filepath.Join(root, "empty-path")}
	wantArgs := make([]string, len(vector.CommonFetchArgv))
	for index, value := range vector.CommonFetchArgv {
		wantArgs[index] = resolveAcquisitionArg(value, paths, request.Tool, request.Source.Transport, request.Source.Git, refspec)
	}
	gotArgs := append([]string{request.Tool.Executable}, invocation.Args...)
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		return fmt.Errorf("production fetch argv = %q, vector requires %q", gotArgs, wantArgs)
	}
	gotEnvironment, err := environmentByName(invocation.Environment)
	if err != nil {
		return fmt.Errorf("production fetch environment: %w", err)
	}
	wantEnvironment := expectedAcquisitionEnvironment(t, vector.CleanEnvironment, paths, request.Tool, request.Source.Transport)
	if !reflect.DeepEqual(gotEnvironment, wantEnvironment) {
		return fmt.Errorf("production fetch environment = %#v, vector requires %#v", gotEnvironment, wantEnvironment)
	}
	return nil
}

func acquisitionRequestForCase(t *testing.T, testCase externalRepositoryAcquisitionCase, shim acquisitionGitShim) (NetworkRequest, *gitFixture, string, string) {
	t.Helper()
	transport := testCase.Transport
	if transport == "" {
		transport = "https"
	}
	sourceURL := "https://fixture.test/repository.git"
	if transport == "ssh" {
		sourceURL = "ssh://fixture.test/repository.git"
	}
	source, err := ParseSource(sourceURL)
	if err != nil {
		t.Fatalf("parse vector source: %v", err)
	}
	format := testCase.ObjectFormat
	if format == "" {
		format = "sha1"
	}
	if testCase.Name == "malformed-ref-rejected-before-git" {
		return NetworkRequest{
			Source:  source,
			Lock:    LockedCommit{ObjectFormat: format, Hex: strings.Repeat("0", objectIDWidth(format))},
			RefKind: "tag", RefValue: testCase.Ref,
			Tool: shim.Tool,
		}, nil, sourceURL, ""
	}

	fixture := makeGitFixture(t, format, false)
	tagName := acquisitionCaseTag(testCase.Name)
	switch testCase.Name {
	case "sha1-tagged-https", "sha256-tagged-ssh", "tag-moved", "network-substitution-tag":
		createAcquisitionTag(t, fixture, tagName)
	case "tag-malformed-object":
		createMalformedAcquisitionTag(t, fixture, tagName)
	case "tag-missing":
		// The empty tag namespace models a source that cannot resolve the exact tag.
	case "network-substitution-branch":
		git := realGitPath(t)
		runTestGit(t, "", git, "--git-dir="+fixture.bare, "update-ref", "refs/heads/release/v2", fixture.commit)
	}

	lock := fixture.commit
	if testCase.Name == "tag-moved" {
		lock = strings.Repeat("0", objectIDWidth(format))
	}
	if testCase.Name == "untagged-missing-object" {
		prefix, _, ok := strings.Cut(testCase.FetchRefspec, ":")
		if !ok {
			t.Fatalf("missing-object vector case has malformed fetch_refspec %q", testCase.FetchRefspec)
		}
		lock = prefix
	}
	request := NetworkRequest{Source: source, Lock: LockedCommit{ObjectFormat: format, Hex: lock}, Tool: shim.Tool}
	switch testCase.Name {
	case "sha1-tagged-https", "sha256-tagged-ssh", "tag-moved", "tag-missing", "tag-malformed-object":
		request.Tag = tagName
	case "network-substitution-revision":
		request.RefKind, request.RefValue = "revision", fixture.commit
	case "network-substitution-tag":
		request.RefKind, request.RefValue = "tag", tagName
	case "network-substitution-branch":
		request.RefKind, request.RefValue = "branch", "release/v2"
	}
	expectedRefspec := testCase.FetchRefspec
	if testCase.Result == "source-resolved" {
		prefix, suffix, ok := strings.Cut(expectedRefspec, ":")
		if ok && isFullObjectID(prefix) {
			expectedRefspec = fixture.commit + ":" + suffix
		}
	}
	return request, &fixture, sourceURL, expectedRefspec
}

func acquisitionCaseTag(name string) string {
	if name == "sha256-tagged-ssh" {
		return "v2.0.0"
	}
	return "v1.4.0"
}

func objectIDWidth(format string) int {
	if format == "sha256" {
		return 64
	}
	return 40
}

func isFullObjectID(value string) bool {
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	for _, character := range value {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	}
	return true
}

func createAcquisitionTag(t *testing.T, fixture gitFixture, tag string) {
	t.Helper()
	git := realGitPath(t)
	runTestGit(t, "", git, "--git-dir="+fixture.bare, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.test", "tag", "-a", tag, fixture.commit, "-m", "fixture tag")
}

func createMalformedAcquisitionTag(t *testing.T, fixture gitFixture, tag string) {
	t.Helper()
	git := realGitPath(t)
	tree := strings.TrimSpace(runTestGit(t, fixture.work, git, "rev-parse", "HEAD^{tree}"))
	body := []byte(fmt.Sprintf("object %s\ntype tree\ntag %s\ntagger Fixture <fixture@example.test> 1 +0000\n\nmalformed terminal target\n", tree, tag))
	oid := writeLooseObject(t, fixture.bare, "tag", body)
	ref := filepath.Join(fixture.bare, "refs", "tags", filepath.FromSlash(tag))
	if err := os.MkdirAll(filepath.Dir(ref), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ref, []byte(oid+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}

func readAcquisitionGitShimInvocations(path string) ([]acquisitionGitShimInvocation, error) {
	file, err := stateread.ReadRegularFile(path)
	if err != nil {
		return nil, err
	}
	switch file.Kind {
	case stateread.KindAbsent:
		return nil, nil
	case stateread.KindPresent:
	default:
		return nil, stateread.UnusableError(path, fmt.Errorf("Git shim log state is %q", file.Kind))
	}
	scanner := bufio.NewScanner(bytes.NewReader(file.Bytes))
	scanner.Buffer(make([]byte, 4096), 1<<20)
	var invocations []acquisitionGitShimInvocation
	for scanner.Scan() {
		var invocation acquisitionGitShimInvocation
		if err := json.Unmarshal(scanner.Bytes(), &invocation); err != nil {
			return nil, fmt.Errorf("decode Git shim invocation: %w", err)
		}
		invocations = append(invocations, invocation)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return invocations, nil
}

func fetchInvocations(invocations []acquisitionGitShimInvocation) []acquisitionGitShimInvocation {
	var fetches []acquisitionGitShimInvocation
	for _, invocation := range invocations {
		for _, arg := range invocation.Args {
			if arg == "fetch" {
				fetches = append(fetches, invocation)
				break
			}
		}
	}
	return fetches
}

func sourceAndRefspec(args []string) (string, string, bool) {
	separator := -1
	for index, arg := range args {
		if arg == "--" {
			separator = index
			break
		}
	}
	if separator < 0 || len(args) != separator+3 {
		return "", "", false
	}
	return args[separator+1], args[separator+2], true
}
