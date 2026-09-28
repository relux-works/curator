package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envprofile"
)

func requireOnePermissivePostureWarning(t *testing.T, stderr string) {
	t.Helper()
	if got := strings.Count(stderr, "security_posture_permissive"); got != 1 {
		t.Fatalf("permissive security-posture warnings = %d, want exactly one in stderr:\n%s", got, stderr)
	}
}

func TestProfileUpdateSystemDeltaConfirmationGolden(t *testing.T) {
	requireGit(t)
	source, home := profileHome(t)
	const rawSource = "https://example.com/delta"
	repo := t.TempDir()
	writeGitRepoFile(t, repo, "agent-context.json", `{"schema_version":1,"name":"delta","version":"1.0.0"}`+"\n")
	runGitRepo(t, repo, "init")
	commitGitRepo(t, repo, "v1.0.0")
	oldCommit := strings.TrimSpace(string(mustGitOutput(t, repo, "rev-parse", "HEAD")))
	serveGitRepos(t, map[string]string{rawSource: repo})
	code, _, stderr := runProfile(t, source, "profile", "install", rawSource)
	if code != exitOK {
		t.Fatalf("initial install = %d: %s", code, stderr)
	}
	lockPath := filepath.Join(envprofile.ProfileDir(home, "delta"), "lock.json")
	oldLock, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	writeGitRepoFile(t, repo, "agent-context.json", `{"schema_version":1,"name":"delta","version":"1.0.1","context":{"modules":[{"path":"system.md","class":"system","environments":["claude_code"]}]}}`+"\n")
	writeGitRepoFile(t, repo, "context/system.md", "system prompt v1\n")
	commitGitRepo(t, repo, "v1.0.1")
	newCommit := strings.TrimSpace(string(mustGitOutput(t, repo, "rev-parse", "HEAD")))
	refusedCode, refusedOut, refusedErr := runProfile(t, source, "profile", "update", "delta")
	if refusedCode != exitFail {
		t.Fatalf("unconfirmed update = %d\nstdout:\n%s\nstderr:\n%s", refusedCode, refusedOut, refusedErr)
	}
	requireOnePermissivePostureWarning(t, refusedErr)
	afterRefusal, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(afterRefusal) != string(oldLock) {
		t.Fatal("confirmation refusal changed the installed lock")
	}
	candidateEntry := contextstore.EntryDir(home, contextlock.KindContext, "delta", newCommit)
	if _, err := os.Lstat(candidateEntry); !os.IsNotExist(err) {
		t.Fatalf("confirmation refusal installed a candidate store entry: %v", err)
	}
	confirmedCode, confirmedOut, confirmedErr := runProfile(t, source, "profile", "update", "delta", "--confirm-system-delta")
	if confirmedCode != exitOK {
		t.Fatalf("confirmed update = %d\nstdout:\n%s\nstderr:\n%s", confirmedCode, confirmedOut, confirmedErr)
	}
	requireOnePermissivePostureWarning(t, confirmedErr)
	if bytesEqual, err := os.ReadFile(lockPath); err != nil || string(bytesEqual) == string(oldLock) {
		t.Fatalf("confirmed update did not publish the new lock: err=%v", err)
	}
	if _, err := os.Stat(candidateEntry); err != nil {
		t.Fatalf("confirmed update did not install the candidate store entry: %v", err)
	}
	transcript := "refused: exit=1\nstdout:\n" + refusedOut + "stderr:\n" + refusedErr +
		"confirmed: exit=0\nstdout:\n" + confirmedOut + "stderr:\n" + confirmedErr
	transcript = normalizeDeltaTranscript(transcript, oldCommit, newCommit)
	want, err := os.ReadFile(filepath.Join("testdata", "profile-update-system-delta.golden"))
	if err != nil {
		t.Fatalf("missing golden: %v\nactual transcript:\n%s", err, transcript)
	}
	if transcript != string(want) {
		t.Fatalf("system-delta output differs from the pinned golden:\n%s", transcript)
	}
}

func TestProfileUpdateMCPDeltaConfirmationGolden(t *testing.T) {
	requireGit(t)
	source, home := profileHome(t)
	const rootSource = "https://example.com/mcp-root"
	const toolSource = "https://example.com/mcp-tool"
	rootRepo := t.TempDir()
	writeGitRepoFile(t, rootRepo, "agent-context.json", `{"schema_version":1,"name":"mcp-root","version":"1.0.0"}`+"\n")
	runGitRepo(t, rootRepo, "init")
	commitGitRepo(t, rootRepo, "v1.0.0")
	oldRootCommit := strings.TrimSpace(string(mustGitOutput(t, rootRepo, "rev-parse", "HEAD")))
	toolRepo := t.TempDir()
	writeGitRepoFile(t, toolRepo, "agent-mcp.json", `{"schema_version":1,"name":"tool","version":"1.0.0","server":{"transport":"stdio","command":"npx","args":["tool"],"env_names":[]}}`+"\n")
	runGitRepo(t, toolRepo, "init")
	commitGitRepo(t, toolRepo, "v1.0.0")
	toolCommit := strings.TrimSpace(string(mustGitOutput(t, toolRepo, "rev-parse", "HEAD")))
	serveGitRepos(t, map[string]string{rootSource: rootRepo, toolSource: toolRepo})
	code, _, stderr := runProfile(t, source, "profile", "install", rootSource)
	if code != exitOK {
		t.Fatalf("initial install = %d: %s", code, stderr)
	}
	lockPath := filepath.Join(envprofile.ProfileDir(home, "mcp-root"), "lock.json")
	oldLock, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	writeGitRepoFile(t, rootRepo, "agent-context.json", `{"schema_version":1,"name":"mcp-root","version":"1.1.0","requires":{"mcp":{"tool":{"git":"`+toolSource+`","range":"*"}}}}`+"\n")
	commitGitRepo(t, rootRepo, "v1.1.0")
	newRootCommit := strings.TrimSpace(string(mustGitOutput(t, rootRepo, "rev-parse", "HEAD")))
	refusedCode, refusedOut, refusedErr := runProfile(t, source, "profile", "update", "mcp-root")
	if refusedCode != exitFail {
		t.Fatalf("unconfirmed update = %d\nstdout:\n%s\nstderr:\n%s", refusedCode, refusedOut, refusedErr)
	}
	requireOnePermissivePostureWarning(t, refusedErr)
	afterRefusal, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(afterRefusal) != string(oldLock) {
		t.Fatal("confirmation refusal changed the installed lock")
	}
	candidateEntries := []string{
		contextstore.EntryDir(home, contextlock.KindContext, "mcp-root", newRootCommit),
		contextstore.EntryDir(home, contextlock.KindMCP, "tool", toolCommit),
	}
	for _, entry := range candidateEntries {
		if _, err := os.Lstat(entry); !os.IsNotExist(err) {
			t.Fatalf("confirmation refusal installed candidate store entry %s: %v", entry, err)
		}
	}
	confirmedCode, confirmedOut, confirmedErr := runProfile(t, source, "profile", "update", "mcp-root", "--confirm-system-delta")
	if confirmedCode != exitOK {
		t.Fatalf("confirmed update = %d\nstdout:\n%s\nstderr:\n%s", confirmedCode, confirmedOut, confirmedErr)
	}
	requireOnePermissivePostureWarning(t, confirmedErr)
	if bytesEqual, err := os.ReadFile(lockPath); err != nil || string(bytesEqual) == string(oldLock) {
		t.Fatalf("confirmed update did not publish the new lock: err=%v", err)
	}
	for _, entry := range candidateEntries {
		if _, err := os.Stat(entry); err != nil {
			t.Fatalf("confirmed update did not install candidate store entry %s: %v", entry, err)
		}
	}
	transcript := "refused: exit=1\nstdout:\n" + refusedOut + "stderr:\n" + refusedErr +
		"confirmed: exit=0\nstdout:\n" + confirmedOut + "stderr:\n" + confirmedErr
	transcript = strings.ReplaceAll(transcript, oldRootCommit, "<old-root-commit>")
	transcript = strings.ReplaceAll(transcript, newRootCommit, "<new-root-commit>")
	transcript = strings.ReplaceAll(transcript, toolCommit, "<tool-commit>")
	transcript = regexp.MustCompile(`sha256:[0-9a-f]{64}`).ReplaceAllString(transcript, "sha256:<lock-hash>")
	want, err := os.ReadFile(filepath.Join("testdata", "profile-update-mcp-delta.golden"))
	if err != nil {
		t.Fatalf("missing golden: %v\nactual transcript:\n%s", err, transcript)
	}
	if transcript != string(want) {
		t.Fatalf("MCP-delta output differs from the pinned golden:\n%s", transcript)
	}
}

type confirmationVectors struct {
	AllCases []struct {
		Name     string `json:"name"`
		Flag     bool   `json:"flag"`
		Expected struct {
			Profiles []struct {
				Profile   string              `json:"profile"`
				Trigger   []string            `json:"trigger"`
				RevisionB profileDeltaOutcome `json:"revision_b"`
			} `json:"profiles"`
			Stopped struct {
				RevisionB *string `json:"revision_b"`
			} `json:"stopped"`
		} `json:"expected"`
	} `json:"all_cases"`
	ReinstallCases []struct {
		Name     string `json:"name"`
		Flag     bool   `json:"flag"`
		Expected struct {
			RevisionB profileDeltaOutcome `json:"revision_b"`
		} `json:"expected"`
	} `json:"reinstall_cases"`
}

type profileDeltaOutcome struct {
	Diagnostic string
	Proceeds   bool
	Untouched  bool
}

func (outcome *profileDeltaOutcome) UnmarshalJSON(payload []byte) error {
	var untouched string
	if err := json.Unmarshal(payload, &untouched); err == nil {
		outcome.Untouched = untouched == "untouched"
		return nil
	}
	var value struct {
		Diagnostic string `json:"diagnostic"`
		Proceeds   bool   `json:"proceeds"`
	}
	if err := json.Unmarshal(payload, &value); err != nil {
		return err
	}
	outcome.Diagnostic, outcome.Proceeds = value.Diagnostic, value.Proceeds
	return nil
}

func loadConfirmationVectors(t *testing.T) confirmationVectors {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "environments-source-signers.json")) // #nosec G304 -- explicit conformance root
	if err != nil {
		t.Fatal(err)
	}
	var vectors confirmationVectors
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	if len(vectors.AllCases) != 2 || len(vectors.ReinstallCases) != 2 {
		t.Fatalf("pinned all/reinstall case counts = %d/%d, want 2/2", len(vectors.AllCases), len(vectors.ReinstallCases))
	}
	return vectors
}

func TestProfileUpdateAllDeltaVectorsAtCLI(t *testing.T) {
	requireGit(t)
	vectors := loadConfirmationVectors(t)
	for _, vector := range vectors.AllCases {
		t.Run(vector.Name, func(t *testing.T) {
			source, home := profileHome(t)
			var repos map[string]string
			var profiles []string
			if vector.Flag {
				const contextSource = "https://example.com/sysleaf"
				const toolSource = "https://example.com/figma"
				companyRepo := newContextRepo(t, "companyA", "1.0.0", "")
				sysRepo := newContextRepo(t, "sysleaf", "1.0.0", "")
				personalRepo := newContextRepo(t, "personal", "1.0.0", `{"mcp":{"figma-devmode":{"git":"`+toolSource+`","range":"*"}}}`)
				toolRepo := newMCPRepo(t, "figma-devmode", "1.2.0", `["figma"]`)
				repos = map[string]string{contextSource: sysRepo, toolSource: toolRepo, "https://example.com/companyA": companyRepo, "https://example.com/personal": personalRepo}
				profiles = []string{"companyA", "personal"}
				serveGitRepos(t, repos)
				for _, profile := range profiles {
					code, _, stderr := runProfile(t, source, "profile", "install", "https://example.com/"+profile)
					if code != exitOK {
						t.Fatalf("install %s = %d: %s", profile, code, stderr)
					}
				}
				oldLocks := profileLocks(t, home, profiles)
				writeGitRepoFile(t, sysRepo, "agent-context.json", `{"schema_version":1,"name":"sysleaf","version":"2.0.0","context":{"modules":[{"path":"90-system.md","class":"system","environments":["claude_code"]}]}}`+"\n")
				writeGitRepoFile(t, sysRepo, "context/90-system.md", "system prompt v1\n")
				commitGitRepo(t, sysRepo, "v2.0.0")
				writeGitRepoFile(t, companyRepo, "agent-context.json", `{"schema_version":1,"name":"companyA","version":"1.1.0","context":{"modules":[{"path":"a.md"}]},"requires":{"contexts":{"sysleaf":{"git":"`+contextSource+`","range":"*"}}}}`+"\n")
				commitGitRepo(t, companyRepo, "v1.1.0")
				writeGitRepoFile(t, toolRepo, "agent-mcp.json", `{"schema_version":1,"name":"figma-devmode","version":"1.3.0","server":{"transport":"stdio","command":"npx","args":["figma","--new"],"env_names":[]}}`+"\n")
				commitGitRepo(t, toolRepo, "v1.3.0")
				code, stdout, stderr := runProfile(t, source, "profile", "update", "--all", "--confirm-system-delta")
				if code != exitOK {
					t.Fatalf("profile update --all --confirm-system-delta = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
				}
				for _, profile := range profiles {
					if string(readProfileLock(t, home, profile)) == string(oldLocks[profile]) {
						t.Errorf("--all did not publish updated lock for %s", profile)
					}
				}
				if !strings.Contains(stdout, "lock-delta added context sysleaf") || !strings.Contains(stdout, "lock-delta moved mcp figma-devmode") {
					t.Errorf("--all output omitted a pinned trigger delta:\n%s", stdout)
				}
				if len(vector.Expected.Profiles) != 2 || !vector.Expected.Profiles[0].RevisionB.Proceeds || !vector.Expected.Profiles[1].RevisionB.Proceeds {
					t.Fatalf("pinned confirmed all vector %q no longer describes two proceeding profiles", vector.Name)
				}
				if !vector.Flag {
					t.Fatal("pinned all vector requires --confirm-system-delta")
				}
			}
		})
	}
}

func TestProfileUpdateAllStopsAtFirstUnconfirmedDelta(t *testing.T) {
	requireGit(t)
	vectors := loadConfirmationVectors(t)
	var vector *confirmationVectorsEntry
	for i := range vectors.AllCases {
		if vectors.AllCases[i].Name == "all-without-flag-stops-at-first-refusal" {
			vector = &confirmationVectorsEntry{Flag: vectors.AllCases[i].Flag, Names: []string{}}
			for _, profile := range vectors.AllCases[i].Expected.Profiles {
				vector.Names = append(vector.Names, profile.Profile)
			}
		}
	}
	if vector == nil || vector.Flag || strings.Join(vector.Names, ",") != "first,second,third" {
		t.Fatalf("pinned stopping vector = %+v", vector)
	}
	source, home := profileHome(t)
	const toolSource = "https://example.com/stop-tool"
	firstRepo := newContextRepo(t, "first", "1.0.0", "")
	secondRepo := newContextRepo(t, "second", "1.0.0", `{"mcp":{"tool":{"git":"`+toolSource+`","range":"*"}}}`)
	thirdRepo := newContextRepo(t, "third", "1.0.0", "")
	toolRepo := newMCPRepo(t, "tool", "1.0.0", `["tool"]`)
	serveGitRepos(t, map[string]string{
		"https://example.com/first": firstRepo, "https://example.com/second": secondRepo,
		"https://example.com/third": thirdRepo, toolSource: toolRepo,
	})
	for _, profile := range vector.Names {
		code, _, stderr := runProfile(t, source, "profile", "install", "https://example.com/"+profile)
		if code != exitOK {
			t.Fatalf("install %s = %d: %s", profile, code, stderr)
		}
	}
	oldLocks := profileLocks(t, home, vector.Names)
	advanceContextRepo(t, firstRepo, "first", "1.1.0")
	writeGitRepoFile(t, toolRepo, "agent-mcp.json", `{"schema_version":1,"name":"tool","version":"1.1.0","server":{"transport":"stdio","command":"npx","args":["tool","--new"],"env_names":[]}}`+"\n")
	commitGitRepo(t, toolRepo, "v1.1.0")
	advanceContextRepo(t, thirdRepo, "third", "1.1.0")
	code, stdout, stderr := runProfile(t, source, "profile", "update", "--all")
	if code != exitFail || !strings.Contains(stderr, envprofile.DiagSystemDeltaConfirmationRequired) {
		t.Fatalf("profile update --all = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if string(readProfileLock(t, home, "first")) == string(oldLocks["first"]) {
		t.Fatal("safe first profile was not updated before the refusal")
	}
	if string(readProfileLock(t, home, "second")) != string(oldLocks["second"]) {
		t.Fatal("second profile lock changed despite the refusal")
	}
	if string(readProfileLock(t, home, "third")) != string(oldLocks["third"]) {
		t.Fatal("--all did not stop after the first confirmation refusal")
	}
}

type confirmationVectorsEntry = struct {
	Flag  bool
	Names []string
}

func TestProfileInstallReinstallDeltaVectorsAtCLI(t *testing.T) {
	requireGit(t)
	vectors := loadConfirmationVectors(t)
	for _, vector := range vectors.ReinstallCases {
		t.Run(vector.Name, func(t *testing.T) {
			source, home := profileHome(t)
			const systemSource = "https://example.com/reinstall-sysleaf"
			const toolSource = "https://example.com/reinstall-tool"
			rootName := "companyA"
			var rootRepo, systemRepo, toolRepo string
			var rootSource string
			if strings.Contains(vector.Name, "with-flag") {
				systemRepo = newContextRepo(t, "sysleaf", "1.0.0", "")
				rootRepo = newContextRepo(t, rootName, "1.0.0", `{"contexts":{"sysleaf":{"git":"`+systemSource+`","range":"*"}}}`)
				rootSource = "https://example.com/reinstall-company"
				serveGitRepos(t, map[string]string{systemSource: systemRepo, rootSource: rootRepo})
			} else {
				rootName = "mcp-root"
				rootRepo = newContextRepo(t, rootName, "1.0.0", `{"mcp":{"figma-devmode":{"git":"`+toolSource+`","range":"*"}}}`)
				toolRepo = newMCPRepo(t, "figma-devmode", "1.2.0", `["figma"]`)
				rootSource = "https://example.com/reinstall-mcp-root"
				serveGitRepos(t, map[string]string{toolSource: toolRepo, rootSource: rootRepo})
			}
			code, _, stderr := runProfile(t, source, "profile", "install", rootSource)
			if code != exitOK {
				t.Fatalf("initial install = %d: %s", code, stderr)
			}
			oldLock := readProfileLock(t, home, rootName)
			args := []string{"profile", "install", rootSource}
			if vector.Flag {
				args = append(args, "--confirm-system-delta")
			}
			if vector.Flag {
				writeGitRepoFile(t, systemRepo, "agent-context.json", `{"schema_version":1,"name":"sysleaf","version":"2.0.0","context":{"modules":[{"path":"90-system.md","class":"system","environments":["claude_code"]}]}}`+"\n")
				writeGitRepoFile(t, systemRepo, "context/90-system.md", "system prompt v1\n")
				commitGitRepo(t, systemRepo, "v2.0.0")
			} else {
				writeGitRepoFile(t, toolRepo, "agent-mcp.json", `{"schema_version":1,"name":"figma-devmode","version":"1.3.0","server":{"transport":"stdio","command":"npx","args":["figma","--new"],"env_names":[]}}`+"\n")
				commitGitRepo(t, toolRepo, "v1.3.0")
			}
			if vector.Flag {
				code, stdout, stderr := runProfile(t, source, args...)
				if code != exitOK {
					t.Fatalf("confirmed reinstall = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
				}
				if string(readProfileLock(t, home, rootName)) == string(oldLock) {
					t.Fatal("confirmed reinstall did not publish the resolved update")
				}
			} else {
				code, stdout, stderr := runProfile(t, source, args...)
				if code != exitFail || !strings.Contains(stderr, envprofile.DiagSystemDeltaConfirmationRequired) {
					t.Fatalf("unconfirmed reinstall = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
				}
				if string(readProfileLock(t, home, rootName)) != string(oldLock) {
					t.Fatal("refused reinstall changed the installed lock")
				}
			}
			if vector.Expected.RevisionB.Proceeds != vector.Flag {
				t.Fatalf("pinned reinstall vector %q has unexpected proceed=%t for flag=%t", vector.Name, vector.Expected.RevisionB.Proceeds, vector.Flag)
			}
		})
	}
}

func newContextRepo(t *testing.T, name, version, requires string) string {
	t.Helper()
	repo := t.TempDir()
	manifest := fmt.Sprintf(`{"schema_version":1,"name":%q,"version":%q,"context":{"modules":[{"path":"a.md"}]}`, name, version)
	if requires != "" {
		manifest += `,"requires":` + requires
	}
	manifest += "}\n"
	writeGitRepoFile(t, repo, "agent-context.json", manifest)
	writeGitRepoFile(t, repo, "context/a.md", "context\n")
	runGitRepo(t, repo, "init")
	commitGitRepo(t, repo, "v"+version)
	return repo
}

func newMCPRepo(t *testing.T, name, version, args string) string {
	t.Helper()
	repo := t.TempDir()
	manifest := fmt.Sprintf(`{"schema_version":1,"name":%q,"version":%q,"server":{"transport":"stdio","command":"npx","args":%s,"env_names":[]}}`, name, version, args)
	writeGitRepoFile(t, repo, "agent-mcp.json", manifest+"\n")
	runGitRepo(t, repo, "init")
	commitGitRepo(t, repo, "v"+version)
	return repo
}

func advanceContextRepo(t *testing.T, repo, name, version string) {
	t.Helper()
	writeGitRepoFile(t, repo, "agent-context.json", fmt.Sprintf(`{"schema_version":1,"name":%q,"version":%q,"context":{"modules":[{"path":"a.md"}]}}`+"\n", name, version))
	commitGitRepo(t, repo, "v"+version)
}

func profileLocks(t *testing.T, home string, names []string) map[string][]byte {
	t.Helper()
	locks := make(map[string][]byte, len(names))
	for _, name := range names {
		locks[name] = readProfileLock(t, home, name)
	}
	return locks
}

func readProfileLock(t *testing.T, home, name string) []byte {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join(envprofile.ProfileDir(home, name), "lock.json"))
	if err != nil {
		t.Fatalf("read %s lock: %v", name, err)
	}
	return payload
}

func normalizeDeltaTranscript(transcript, oldCommit, newCommit string) string {
	transcript = strings.ReplaceAll(transcript, oldCommit, "<old-commit>")
	transcript = strings.ReplaceAll(transcript, newCommit, "<new-commit>")
	return regexp.MustCompile(`sha256:[0-9a-f]{64}`).ReplaceAllString(transcript, "sha256:<lock-hash>")
}
