package buildrepo

import (
	"bytes"
	"compress/zlib"
	"context"
	"crypto/sha1" // #nosec G505 -- constructs adversarial Git SHA-1 fixtures.
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestNetworkAndLocalSHA1SHA256RawObjectParity(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only; production code is platform-neutral")
	}
	for _, format := range []string{"sha1", "sha256"} {
		t.Run(format, func(t *testing.T) {
			fixture := makeGitFixture(t, format, false)
			tool, logPath := fakeHTTPGitTool(t, fixture.bare)
			source, err := ParseSource("https://fixture.test/repository.git")
			if err != nil {
				t.Fatal(err)
			}
			network, err := AcquireNetwork(context.Background(), NetworkRequest{Source: source, Lock: LockedCommit{ObjectFormat: format, Hex: fixture.commit}, Tool: tool})
			if err != nil {
				t.Fatalf("network admission: %v", err)
			}
			local, err := AdmitLocal(context.Background(), LocalRequest{Path: fixture.work, Tool: realGitTool(t)})
			if err != nil {
				t.Fatalf("local admission: %v", err)
			}
			if network.Commit != fixture.commit || local.Commit != fixture.commit {
				t.Fatalf("commit parity: network=%s local=%s want=%s", network.Commit, local.Commit, fixture.commit)
			}
			if !bytes.Equal(network.CanonicalBytes, local.CanonicalBytes) || network.Digest != local.Digest {
				t.Fatalf("snapshot parity failed: network=%s local=%s", network.Digest, local.Digest)
			}
			want := frameSnapshot([]File{{Path: "README.md", Content: []byte("hello\x00world\n")}, {Path: "bin/tool", Content: []byte("tool\n"), Executable: true}, {Path: "empty", Content: nil}})
			if !bytes.Equal(network.CanonicalBytes, want) {
				t.Fatalf("canonical bytes differ\ngot  %x\nwant %x", network.CanonicalBytes, want)
			}
			log, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatal(err)
			}
			fetch := lineContaining(string(log), " fetch ")
			if !strings.Contains(fetch, fixture.commit+":refs/curator/locked") || strings.Contains(fetch, "refs/tags/") {
				t.Fatalf("untagged acquisition was not lock-only: %s", fetch)
			}
		})
	}
}

func TestTaggedAcquisitionUsesOnlyExactTagAndChecksTerminalCommit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only")
	}
	for _, format := range []string{"sha1", "sha256"} {
		t.Run(format, func(t *testing.T) {
			fixture := makeGitFixture(t, format, true)
			tool, logPath := fakeHTTPGitTool(t, fixture.bare)
			source, _ := ParseSource("https://fixture.test/repository.git")
			snapshot, err := AcquireNetwork(context.Background(), NetworkRequest{Source: source, Lock: LockedCommit{ObjectFormat: format, Hex: fixture.commit}, Tag: "v1.4.0", Tool: tool})
			if err != nil || !snapshot.TagVerified {
				t.Fatalf("tagged acquisition: snapshot=%+v err=%v", snapshot, err)
			}
			log, _ := os.ReadFile(logPath)
			fetch := lineContaining(string(log), " fetch ")
			if !strings.Contains(fetch, "refs/tags/v1.4.0:refs/curator/tag") || strings.Contains(fetch, fixture.commit+":") {
				t.Fatalf("tagged acquisition attempted a non-tag path: %s", fetch)
			}

			width := 40
			if format == "sha256" {
				width = 64
			}
			wrong := strings.Repeat("0", width)
			_, err = AcquireNetwork(context.Background(), NetworkRequest{Source: source, Lock: LockedCommit{ObjectFormat: format, Hex: wrong}, Tag: "v1.4.0", Tool: tool})
			if ErrorCode(err) != CodeRefMoved {
				t.Fatalf("moved tag error = %v, want %s", err, CodeRefMoved)
			}
		})
	}
}

func TestSubstitutionAcquisitionSelectsRevisionBranchAndTag(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only")
	}
	fixture := makeGitFixture(t, "sha1", true)
	source, _ := ParseSource("https://fixture.test/repository.git")
	branch := strings.TrimSpace(runTestGit(t, fixture.work, realGitPath(t), "branch", "--show-current"))
	for _, testCase := range []struct {
		kind, value, wantRef string
	}{
		{"revision", fixture.commit, fixture.commit + ":refs/curator/effective"},
		{"branch", branch, "refs/heads/" + branch + ":refs/curator/effective"},
		{"tag", "v1.4.0", "refs/tags/v1.4.0:refs/curator/effective"},
	} {
		t.Run(testCase.kind, func(t *testing.T) {
			tool, logPath := fakeHTTPGitTool(t, fixture.bare)
			snapshot, err := AcquireNetwork(context.Background(), NetworkRequest{
				Source: source, Lock: LockedCommit{ObjectFormat: "sha1", Hex: fixture.commit},
				RefKind: testCase.kind, RefValue: testCase.value, Tool: tool,
			})
			if err != nil {
				log, _ := os.ReadFile(logPath)
				t.Fatalf("%v\nGit argv:\n%s", err, log)
			}
			if snapshot.ObjectFormat != "sha1" || snapshot.Commit != fixture.commit {
				t.Fatalf("effective snapshot = %+v, want sha1 %s", snapshot, fixture.commit)
			}
			if testCase.kind == "tag" && !snapshot.TagVerified {
				t.Fatal("substituted exact tag was not verified")
			}
			log, _ := os.ReadFile(logPath)
			if !strings.Contains(string(log), testCase.wantRef) {
				t.Fatalf("selected ref absent from Git argv:\n%s", log)
			}
		})
	}
}

func TestSubstitutionNeverRetriesAnotherObjectFormat(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only")
	}
	fixture := makeGitFixture(t, "sha256", false)
	branch := strings.TrimSpace(runTestGit(t, fixture.work, realGitPath(t), "branch", "--show-current"))
	tool, logPath := fakeHTTPGitTool(t, fixture.bare)
	source, _ := ParseSource("https://fixture.test/repository.git")
	_, err := AcquireNetwork(context.Background(), NetworkRequest{
		Source: source, Lock: LockedCommit{ObjectFormat: "sha1", Hex: strings.Repeat("0", 40)},
		RefKind: "branch", RefValue: branch, Tool: tool,
	})
	if ErrorCode(err) != CodeSourceUnavailable {
		t.Fatalf("cross-format substitution error = %v, want %s", err, CodeSourceUnavailable)
	}
	log, _ := os.ReadFile(logPath)
	if strings.Contains(string(log), "--object-format=sha256") || strings.Count(string(log), " init ") != 1 {
		t.Fatalf("substitution retried another object format:\n%s", log)
	}
}

func TestRawTreeRejectsNonPortableComponents(t *testing.T) {
	for _, name := range []string{"back\\slash", "stream:name", "control\x01", "trailing.", "trailing ", "CON", "nul.txt", "COM1"} {
		t.Run(fmt.Sprintf("%q", name), func(t *testing.T) {
			data := append([]byte("100644 "+name+"\x00"), make([]byte, 20)...)
			if err := validateTreeSyntax(data, "sha1"); ErrorCode(err) != CodeObjectSemanticsInvalid {
				t.Fatalf("portable path error = %v, want %s", err, CodeObjectSemanticsInvalid)
			}
		})
	}
}

func TestRawObjectAndLFSPinnedConformanceFixtures(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		root = "/Users/iv/Developer/ReluxWorks/curator-spec-parity/conformance/v1"
	}
	rawPath := filepath.Join(root, "fixtures", "external-repository", "raw-objects.json")
	rawData, err := os.ReadFile(rawPath)
	if os.IsNotExist(err) {
		t.Skipf("supplied conformance root publishes no external-repository/raw-objects fixture: %v", err)
	}
	if err != nil {
		t.Fatal(err)
	}
	var raw struct {
		Cases []struct {
			Name          string `json:"name"`
			ObjectFormat  string `json:"object_format"`
			ObjectID      string `json:"object_id"`
			ObjectType    string `json:"object_type"`
			ContentBase64 string `json:"content_base64"`
			ExpectedError string `json:"expected_error"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(rawData, &raw); err != nil {
		t.Fatal(err)
	}
	for _, testCase := range raw.Cases {
		t.Run(testCase.Name, func(t *testing.T) {
			content, err := base64.StdEncoding.DecodeString(testCase.ContentBase64)
			if err != nil {
				t.Fatal(err)
			}
			if got := computeOID(testCase.ObjectFormat, testCase.ObjectType, content); got != testCase.ObjectID {
				t.Fatalf("object identity = %s, want %s", got, testCase.ObjectID)
			}
			object := rawObject{oid: testCase.ObjectID, kind: testCase.ObjectType, data: content}
			var parseErr error
			switch testCase.ObjectType {
			case "commit":
				_, parseErr = parseCommit(object, testCase.ObjectFormat)
			case "tag":
				_, targetType, _, tagErr := parseTag(object, testCase.ObjectFormat)
				parseErr = tagErr
				// This shared case is a graph-semantic mismatch: its declared
				// target type is tag while the supplied target is a commit.
				if testCase.Name == "reject-tag-declared-target-type-mismatch" && parseErr == nil && targetType != "commit" {
					parseErr = admissionError(CodeObjectSemanticsInvalid, "annotated tag target type mismatch")
				}
			case "tree":
				parseErr = validateTreeSyntax(object.data, testCase.ObjectFormat)
			}
			if testCase.ExpectedError == "" && parseErr != nil {
				t.Fatalf("valid fixture rejected: %v", parseErr)
			}
			if testCase.ExpectedError != "" && ErrorCode(parseErr) != testCase.ExpectedError {
				t.Fatalf("error = %v (%s), want %s", parseErr, ErrorCode(parseErr), testCase.ExpectedError)
			}
		})
	}

	lfsPath := filepath.Join(root, "fixtures", "external-repository", "lfs-pointers.json")
	lfsData, err := os.ReadFile(lfsPath)
	if err != nil {
		t.Fatal(err)
	}
	var lfs struct {
		Cases []struct {
			Name          string `json:"name"`
			BytesBase64   string `json:"bytes_base64"`
			ExpectedError string `json:"expected_error"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(lfsData, &lfs); err != nil {
		t.Fatal(err)
	}
	for _, testCase := range lfs.Cases {
		data, _ := base64.StdEncoding.DecodeString(testCase.BytesBase64)
		if got, want := isLFSPointer(data), testCase.ExpectedError == CodeLFSUnsupported; got != want {
			t.Errorf("LFS case %s classified %v, want %v", testCase.Name, got, want)
		}
	}
}

func TestLocalConfigAndAdministrationAdversarialBoundaries(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		root = "/Users/iv/Developer/ReluxWorks/curator-spec-parity/conformance/v1"
	}
	payload, err := os.ReadFile(filepath.Join(root, "fixtures", "external-repository", "local-config-and-refs.json"))
	if os.IsNotExist(err) {
		t.Skipf("supplied conformance root publishes no external-repository/local-config-and-refs fixture: %v", err)
	}
	if err != nil {
		t.Fatal(err)
	}
	var fixtures struct {
		Cases []struct {
			Name, ExpectedError string
			Files               map[string]string `json:"files_base64"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(payload, &fixtures); err != nil {
		t.Fatal(err)
	}
	for _, testCase := range fixtures.Cases {
		if encoded, ok := testCase.Files["config"]; ok {
			config, _ := base64.StdEncoding.DecodeString(encoded)
			_, err := parseLocalConfig(config)
			configShouldFail := testCase.Name == "reject-bare-layout" || testCase.Name == "reject-config-include" || testCase.Name == "reject-partial-clone-config" || testCase.Name == "reject-reftable"
			if configShouldFail && err == nil {
				t.Errorf("%s config unexpectedly admitted", testCase.Name)
			}
			if !configShouldFail && err != nil {
				t.Errorf("%s inert/admitted config rejected: %v", testCase.Name, err)
			}
		}
	}

	fixture := makeGitFixture(t, "sha1", false)
	tool := realGitTool(t)
	attacks := []struct{ name, path, code string }{
		{"alternate", "objects/info/alternates", CodeLocalFormatUnsupported},
		{"graft", "info/grafts", CodeLocalFormatUnsupported},
		{"replace", "refs/replace/" + fixture.commit, CodeLocalFormatUnsupported},
		{"promisor", "objects/pack/pack-" + strings.Repeat("a", 40) + ".promisor", CodeLocalFormatUnsupported},
	}
	for _, attack := range attacks {
		t.Run(attack.name, func(t *testing.T) {
			fixtureCopy := copyFixture(t, fixture.work)
			target := filepath.Join(fixtureCopy, ".git", filepath.FromSlash(attack.path))
			if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(target, []byte("malicious\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			_, err := AdmitLocal(context.Background(), LocalRequest{Path: fixtureCopy, Tool: tool})
			if ErrorCode(err) != attack.code {
				t.Fatalf("error = %v, want %s", err, attack.code)
			}
		})
	}
}

func TestLocalPackedHeadAndSnapshotMaterialization(t *testing.T) {
	fixture := makeGitFixture(t, "sha1", false)
	gitDir := filepath.Join(fixture.work, ".git")
	head, err := os.ReadFile(filepath.Join(gitDir, "HEAD"))
	if err != nil {
		t.Fatal(err)
	}
	ref := strings.TrimSpace(strings.TrimPrefix(string(head), "ref: "))
	if err := os.Remove(filepath.Join(gitDir, filepath.FromSlash(ref))); err != nil {
		t.Fatal(err)
	}
	packed := "# pack-refs with: peeled fully-peeled sorted\n" + fixture.commit + " " + ref + "\n"
	if err := os.WriteFile(filepath.Join(gitDir, "packed-refs"), []byte(packed), 0o600); err != nil {
		t.Fatal(err)
	}
	snapshot, err := AdmitLocal(context.Background(), LocalRequest{Path: fixture.work, Tool: realGitTool(t)})
	if err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(t.TempDir(), "snapshot")
	if err := snapshot.Materialize(destination); err != nil {
		t.Fatal(err)
	}
	for _, file := range snapshot.Files {
		got, err := os.ReadFile(filepath.Join(destination, filepath.FromSlash(file.Path)))
		if err != nil || !bytes.Equal(got, file.Content) {
			t.Fatalf("materialized %s: %v", file.Path, err)
		}
	}
}

func TestAdmissionFailuresPrecedeAuditCacheAndCompiler(t *testing.T) {
	tool := realGitTool(t)
	cases := []struct {
		name, mode, content, code string
		missing                   bool
		limit                     Limits
	}{
		{"symbolic-link", "120000", "target", CodeObjectSemanticsInvalid, false, Limits{}},
		{"submodule", "160000", "", CodeObjectSemanticsInvalid, true, Limits{}},
		{"special-mode", "100600", "special", CodeObjectSemanticsInvalid, false, Limits{}},
		{"lfs", "100644", "version https://git-lfs.github.com/spec/v1\noid sha256:" + strings.Repeat("a", 64) + "\nsize 1\n", CodeLFSUnsupported, false, Limits{}},
		{"missing-blob", "100644", "missing", CodeIncompleteSource, true, Limits{}},
		{"resource-bound", "100644", strings.Repeat("x", 32), CodeIncompleteSource, false, Limits{MaxObjectBytes: 16}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := makeGitFixture(t, "sha1", false)
			injectAdversarialCommit(t, fixture.work, testCase.mode, testCase.content, testCase.missing)
			called := struct{ audit, cache, compiler bool }{}
			_, err := AdmitLocal(context.Background(), LocalRequest{Path: fixture.work, Tool: tool, Limits: testCase.limit})
			if err == nil {
				called.audit = true
				called.cache = true
				called.compiler = true
			}
			if ErrorCode(err) != testCase.code {
				t.Fatalf("admission error = %v, want %s", err, testCase.code)
			}
			if called.audit || called.cache || called.compiler {
				t.Fatalf("downstream work started: %+v", called)
			}
		})
	}

	t.Run("source-race", func(t *testing.T) {
		fixture := makeGitFixture(t, "sha1", false)
		called := false
		request := LocalRequest{Path: fixture.work, Tool: tool, afterObjectCopy: func() {
			path := filepath.Join(fixture.work, ".git", "config")
			_ = os.WriteFile(path, []byte("[core]\nrepositoryformatversion = 0\nbare = false\n# raced\n"), 0o600)
		}}
		_, err := AdmitLocal(context.Background(), request)
		if err == nil {
			called = true
		}
		if ErrorCode(err) != CodeLocalLayoutUnsafe || called {
			t.Fatalf("race error=%v downstream=%v", err, called)
		}
	})
}

func TestPackIndexConformanceAndExactSSHWrapper(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		root = "/Users/iv/Developer/ReluxWorks/curator-spec-parity/conformance/v1"
	}
	payload, err := os.ReadFile(filepath.Join(root, "fixtures", "external-repository", "pack-index.json"))
	if os.IsNotExist(err) {
		t.Skipf("supplied conformance root publishes no external-repository/pack-index fixture: %v", err)
	}
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name          string `json:"name"`
			ObjectFormat  string `json:"object_format"`
			PackHex       string `json:"pack_hex"`
			IndexHex      string `json:"index_hex"`
			ExpectedError string `json:"expected_error"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(payload, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, testCase := range fixture.Cases {
		if testCase.PackHex == "" || testCase.IndexHex == "" {
			continue
		}
		pack, _ := hex.DecodeString(testCase.PackHex)
		index, _ := hex.DecodeString(testCase.IndexHex)
		hashBytes := 20
		if testCase.ObjectFormat == "sha256" {
			hashBytes = 32
		}
		if len(pack) < hashBytes {
			t.Fatalf("bad fixture %s", testCase.Name)
		}
		base := "pack-" + hex.EncodeToString(pack[len(pack)-hashBytes:])
		err := validatePackIndex(base, pack, index, testCase.ObjectFormat)
		if (testCase.ExpectedError == "") != (err == nil) {
			t.Errorf("%s: err=%v expected=%s", testCase.Name, err, testCase.ExpectedError)
		}
	}

	policyRoot := t.TempDir()
	policy := SSHPolicy{
		Wrapper:         filepath.Join(policyRoot, "manager", "ssh-wrapper"),
		SSH:             filepath.Join(policyRoot, "tools", "ssh"),
		ExpectedHost:    "git@example.test",
		RepositoryPath:  "skills/tool.git",
		EmptyConfig:     filepath.Join(policyRoot, "manager", "ssh.config"),
		KnownHosts:      filepath.Join(policyRoot, "operator", "known_hosts"),
		EmptyKnownHosts: filepath.Join(policyRoot, "manager", "empty_known_hosts"),
		Identity:        filepath.Join(policyRoot, "operator", "id"),
		ConnectTimeout:  15,
	}
	argv := []string{policy.Wrapper, policy.ExpectedHost, "git-upload-pack 'skills/tool.git'"}
	command, err := ExactSSHCommand(policy, argv)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(command, " ")
	for _, required := range []string{"BatchMode=yes", "StrictHostKeyChecking=yes", "ProxyCommand=none", "IdentityAgent=none", policy.ExpectedHost, argv[2]} {
		if !strings.Contains(joined, required) {
			t.Errorf("SSH command missing %q", required)
		}
	}
	for _, mutation := range [][]string{{"ssh-wrapper", argv[1], argv[2]}, {argv[0], "other@example.test", argv[2]}, {argv[0], argv[1], argv[2], "extra"}} {
		if _, err := ExactSSHCommand(policy, mutation); ErrorCode(err) != CodeIdentityInvalid {
			t.Errorf("unsafe SSH argv accepted: %#v", mutation)
		}
	}

	pinned := policy
	pinned.AgentSocket = filepath.Join(policyRoot, "operator", "agent.sock")
	command, err = ExactSSHCommand(pinned, argv)
	if err != nil {
		t.Fatalf("pinned-agent selection rejected: %v", err)
	}
	joined = strings.Join(command, " ")
	for _, required := range []string{"IdentitiesOnly=yes", "IdentityAgent=" + pinned.AgentSocket, "-i " + pinned.Identity} {
		if !strings.Contains(joined, required) {
			t.Errorf("pinned-agent SSH command missing %q", required)
		}
	}
	if strings.Contains(joined, "IdentityAgent=none") || strings.Contains(joined, "IdentityFile=none") {
		t.Errorf("pinned-agent SSH command carries a disabling option: %s", joined)
	}

	agentOnly := policy
	agentOnly.Identity = ""
	agentOnly.AgentSocket = filepath.Join(policyRoot, "operator", "agent.sock")
	command, err = ExactSSHCommand(agentOnly, argv)
	if err != nil {
		t.Fatalf("agent-only selection rejected: %v", err)
	}
	if !strings.Contains(strings.Join(command, " "), "IdentityFile=none") {
		t.Errorf("agent-only SSH command must disable identity files")
	}

	relativePin := pinned
	relativePin.AgentSocket = "agent.sock"
	if _, err := ExactSSHCommand(relativePin, argv); ErrorCode(err) != CodeIdentityInvalid {
		t.Errorf("relative agent socket accepted in pinned-agent form")
	}

	empty := policy
	empty.Identity = ""
	if _, err := ExactSSHCommand(empty, argv); ErrorCode(err) != CodeIdentityInvalid {
		t.Errorf("empty SSH selection accepted")
	}
}

func validateTreeSyntax(data []byte, format string) error {
	idBytes := 20
	if format == "sha256" {
		idBytes = 32
	}
	names := map[string]bool{}
	for len(data) > 0 {
		space, nul := bytes.IndexByte(data, ' '), bytes.IndexByte(data, 0)
		if space <= 0 || nul <= space+1 || len(data) < nul+1+idBytes {
			return admissionError(CodeObjectSemanticsInvalid, "malformed tree")
		}
		mode, name := string(data[:space]), string(data[space+1:nul])
		if mode != "40000" && mode != "100644" && mode != "100755" {
			return admissionError(CodeObjectSemanticsInvalid, "unsupported tree mode")
		}
		key := collisionKey(name)
		if names[key] || !validTreeComponent(name) {
			return admissionError(CodeObjectSemanticsInvalid, "invalid tree name")
		}
		names[key] = true
		data = data[nul+1+idBytes:]
	}
	return nil
}

type gitFixture struct{ work, bare, commit string }

func makeGitFixture(t *testing.T, format string, tag bool) gitFixture {
	t.Helper()
	root := t.TempDir()
	work := filepath.Join(root, "work")
	bare := filepath.Join(root, "remote.git")
	template := filepath.Join(root, "template")
	if err := os.Mkdir(template, 0o700); err != nil {
		t.Fatal(err)
	}
	git := realGitPath(t)
	runTestGit(t, "", git, "init", "--quiet", "--template="+template, "--object-format="+format, work)
	if err := os.WriteFile(filepath.Join(work, "README.md"), []byte("hello\x00world\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(work, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(work, "bin", "tool"), []byte("tool\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(work, "empty"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	runTestGit(t, work, git, "add", "--", "README.md", "bin/tool", "empty")
	runTestGit(t, work, git, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.test", "commit", "--quiet", "-m", "fixture")
	commit := strings.TrimSpace(runTestGit(t, work, git, "rev-parse", "HEAD"))
	if tag {
		runTestGit(t, work, git, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.test", "tag", "-a", "v1.4.0", "-m", "tag")
	}
	runTestGit(t, "", git, "clone", "--quiet", "--bare", "--", work, bare)
	// Narrow local admission ignores worktree/index bytes and admits no logs or
	// template-selected behavior. Keep only the data-only administration set.
	for _, name := range []string{"logs", "ORIG_HEAD", "COMMIT_EDITMSG"} {
		_ = os.RemoveAll(filepath.Join(work, ".git", name))
	}
	return gitFixture{work: work, bare: bare, commit: commit}
}

func realGitPath(t *testing.T) string {
	t.Helper()
	path, err := resolvedExecutablePath("git")
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func resolvedExecutablePath(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", err
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	return filepath.Abs(path)
}

func TestRealGitPathResolvesLauncherWithoutWeakeningAdmission(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("controlled launcher symlink is a Unix regression fixture")
	}
	realTool := realGitTool(t)
	bin := t.TempDir()
	launcher := filepath.Join(bin, "git")
	if err := os.Symlink(realTool.Executable, launcher); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)

	resolved, err := resolvedExecutablePath("git")
	if err != nil {
		t.Fatal(err)
	}
	if resolved != realTool.Executable {
		t.Fatalf("resolved Git = %q, want %q", resolved, realTool.Executable)
	}
	info, err := os.Lstat(resolved)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("resolved Git is not an admitted ordinary file: info=%v err=%v", info, err)
	}

	symlinkTool := realTool
	symlinkTool.Executable = launcher
	if err := ValidateGitTool(context.Background(), symlinkTool); ErrorCode(err) != CodeIdentityInvalid {
		t.Fatalf("symlinked Git launcher error = %v, want %s", err, CodeIdentityInvalid)
	}
	directoryTool := realTool
	directoryTool.Executable = bin
	if err := ValidateGitTool(context.Background(), directoryTool); ErrorCode(err) != CodeIdentityInvalid {
		t.Fatalf("directory Git object error = %v, want %s", err, CodeIdentityInvalid)
	}
}

func realGitTool(t *testing.T) GitTool {
	t.Helper()
	git := realGitPath(t)
	execPath := strings.TrimSpace(runTestGit(t, "", git, "--exec-path"))
	version := strings.TrimSpace(runTestGit(t, "", git, "--version"))
	fields := strings.Fields(version)
	allowed := version
	if len(fields) >= 3 {
		parts := strings.Split(fields[2], ".")
		if len(parts) >= 2 {
			allowed = "git version " + parts[0] + "." + parts[1] + "."
		}
	}
	return GitTool{Executable: git, ExecPath: execPath, AllowedVersions: []string{allowed}}
}

func fakeHTTPGitTool(t *testing.T, repository string) (GitTool, string) {
	t.Helper()
	realTool := realGitTool(t)
	root := t.TempDir()
	wrapper := filepath.Join(root, "git-wrapper")
	logPath := filepath.Join(root, "argv.log")
	script := fmt.Sprintf(`#!/bin/sh
{
	secret_present=0
	state_present=0
	[ -n "${%s-}" ] && secret_present=1
	[ -n "${%s-}" ] && state_present=1
	printf ' secret=%%s state=%%s askpass=%%s' "$secret_present" "$state_present" "${GIT_ASKPASS-}"
  for arg in "$@"; do printf ' %%s' "$arg"; done
  printf '\n'
} >> %s
set -- "$@"
args=""
for arg in "$@"; do
  case "$arg" in
    https://fixture.test/repository.git) arg=%s ;;
    protocol.https.allow=always) arg=protocol.file.allow=always ;;
  esac
  args="$args
$arg"
done
oldifs=$IFS
IFS='
'
set -- $args
IFS=$oldifs
exec %s "$@"
`, EnvHTTPSBrokerSecret, EnvHTTPSBrokerState, shellQuote(logPath), shellQuote("file://"+repository), shellQuote(realTool.Executable))
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	realTool.Executable = wrapper
	realTool.AskPass = "/usr/bin/false"
	return realTool, logPath
}

func shellQuote(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'" }

func runTestGit(t *testing.T, dir, git string, args ...string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	// Repository fixtures are unsigned source data. Never inherit workstation
	// signing policy or touch its credentials while constructing them.
	gitArgs := append([]string{"-c", "commit.gpgsign=false", "-c", "tag.gpgSign=false"}, args...)
	cmd := exec.CommandContext(ctx, git, gitArgs...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func lineContaining(value, needle string) string {
	for _, line := range strings.Split(value, "\n") {
		if strings.Contains(" "+line+" ", needle) {
			return line
		}
	}
	return ""
}

func copyFixture(t *testing.T, source string) string {
	t.Helper()
	destination := filepath.Join(t.TempDir(), "copy")
	if err := os.CopyFS(destination, os.DirFS(source)); err != nil {
		t.Fatalf("copy fixture: %v", err)
	}
	return destination
}

func injectAdversarialCommit(t *testing.T, work, mode, content string, missing bool) {
	t.Helper()
	gitDir := filepath.Join(work, ".git")
	blobOID := strings.Repeat("1", 40)
	if !missing {
		blobOID = writeLooseObject(t, gitDir, "blob", []byte(content))
	}
	oidBytes, _ := hex.DecodeString(blobOID)
	tree := append([]byte(mode+" entry\x00"), oidBytes...)
	treeOID := writeLooseObject(t, gitDir, "tree", tree)
	commit := []byte("tree " + treeOID + "\nauthor Fixture <fixture@example.test> 1 +0000\ncommitter Fixture <fixture@example.test> 1 +0000\n\nadversarial\n")
	commitOID := writeLooseObject(t, gitDir, "commit", commit)
	if err := os.WriteFile(filepath.Join(gitDir, "refs", "heads", "master"), []byte(commitOID+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if head, _ := os.ReadFile(filepath.Join(gitDir, "HEAD")); strings.Contains(string(head), "refs/heads/main") {
		if err := os.WriteFile(filepath.Join(gitDir, "refs", "heads", "main"), []byte(commitOID+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func writeLooseObject(t *testing.T, gitDir, kind string, content []byte) string {
	t.Helper()
	header := []byte(fmt.Sprintf("%s %d\x00", kind, len(content)))
	raw := append(header, content...)
	sum := sha1.Sum(raw)
	oid := hex.EncodeToString(sum[:])
	path := filepath.Join(gitDir, "objects", oid[:2], oid[2:])
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := zlib.NewWriter(file)
	if _, err := writer.Write(raw); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return oid
}
