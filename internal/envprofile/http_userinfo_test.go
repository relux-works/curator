package envprofile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/contextstore"
)

const httpUserCanary = "synthetic-user-canary"

// Public Install witnesses are compatible with the original PR parent. Routing
// both raw and safe endpoints keeps safety independent of fixture availability;
// insteadOf is offline routing, not wire authentication or rewrite isolation.
func TestInstallHTTPUserinfoSinks(t *testing.T) {
	for _, sink := range []string{"cache-origin", "clone-error", "range-conflict"} {
		for _, user := range []string{"", httpUserCanary} {
			name := "valid-sibling"
			if user != "" {
				name = "no-canary"
			}
			t.Run(sink+"/"+name, func(t *testing.T) {
				pinHomes(t)
				pinOperatorHome(t, t.TempDir())
				ids := newGitIdentities(t)
				const safe = "https://example.com/review/dep.git"
				declared := safe
				if user != "" {
					declared = "https://" + user + "@example.com/review/dep.git"
				}
				const payload = "# Independent F1 dependency bytes\n"
				repo := gitRepo(t, map[string]string{
					"agent-context.json": `{"schema_version":1,"name":"dep","version":"1.0.0","context":{"modules":[{"path":"a.md"}]}}`,
					"context/a.md":       payload,
				}, "v1.0.0")
				pin := fixtureGitOutput(t, repo, "rev-parse", "HEAD")
				target := repo
				if sink == "clone-error" {
					target = filepath.Join(t.TempDir(), "missing-repo")
				}
				for _, endpoint := range []string{safe, "https://example.com/review/dep", "https://" + httpUserCanary + "@example.com/review/dep.git"} {
					ids.serve(target, endpoint)
				}
				constraint := `"tag":"v1.0.0"`
				if sink == "range-conflict" {
					constraint = `"range":"^2.0.0"`
				}
				root, home := t.TempDir(), t.TempDir()
				writeManifestPackage(t, root, fmt.Sprintf(`{"schema_version":1,"name":"root","version":"1.0.0","requires":{"contexts":{"dep":{"git":%q,%s}}}}`, declared, constraint), nil)
				info, _, _, err := Install(home, InstallOptions{Operand: root})
				var observed string
				switch sink {
				case "cache-origin":
					if err != nil {
						t.Fatal(err)
					}
					member, ok := info.Lock.Find(contextlock.KindContext, "dep")
					if !ok || member.Commit != pin || member.Source != "example.com/review/dep" {
						t.Fatalf("independent pin/canonical lock: %+v", member)
					}
					canonical, canonicalErr := info.Lock.Canonical()
					persisted, lockErr := os.ReadFile(lockPath(home, "root"))
					if canonicalErr != nil || lockErr != nil || string(persisted) != string(canonical) || strings.Contains(string(persisted), httpUserCanary) {
						t.Fatalf("portable lock bytes: %v %v", canonicalErr, lockErr)
					}
					data, readErr := os.ReadFile(filepath.Join(contextstore.EntryDir(home, member.Kind, member.Name, member.PinKey()), "context", "a.md"))
					if readErr != nil || string(data) != payload {
						t.Fatalf("independent content: %q %v", data, readErr)
					}
					observed = fixtureGitOutput(t, newGitManager(home).repoDir(member.Source), "config", "--get", "remote.origin.url")
					if observed != safe && observed != "https://example.com/review/dep" && !strings.Contains(observed, httpUserCanary) {
						t.Fatalf("unexpected cache origin: %q", observed)
					}
				case "clone-error":
					if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) || !strings.Contains(err.Error(), "clone") || !strings.Contains(err.Error(), "does not appear to be a git repository") {
						t.Fatalf("expected missing-repository acquisition failure: %v", err)
					}
					observed = err.Error()
				case "range-conflict":
					var conflict *contextresolve.Error
					if !errors.As(err, &conflict) || conflict.Diagnostic != contextresolve.DiagRangeConflict || len(conflict.Requirers) == 0 {
						t.Fatalf("expected structured range conflict: %v", err)
					}
					observed = fmt.Sprintf("%v %+v", err, conflict.Requirers)
				}
				if strings.Contains(observed, httpUserCanary) {
					t.Fatalf("%s leaked synthetic userinfo: %s", sink, observed)
				}
			})
		}
	}
}

// packageOf feeds resolver diagnostic constraints; real Identity/recordRaw feeds
// acquisition. These tests claim those boundaries, not external HTTP/SSH auth.
func TestHTTPUserinfoDeclarationBoundaries(t *testing.T) {
	for _, tc := range []struct{ name, declared, endpoint string }{
		{"https-user", "https://" + httpUserCanary + "@example.com/team/dep.git", "https://example.com/team/dep.git"},
		{"http-user", "http://" + httpUserCanary + "@example.com/team/dep.git", "http://example.com/team/dep.git"},
		{"mixed-case", "HTTPS://" + httpUserCanary + "@EXAMPLE.com/team/dep.git", "https://EXAMPLE.com/team/dep.git"},
		{"empty-user", "https://@example.com/team/dep.git", "https://example.com/team/dep.git"},
		{"punctuation-user", "https://token+name@example.com/team/dep.git", "https://example.com/team/dep.git"},
		{"https-hint", "https://example.com/team/dep.git", "https://example.com/team/dep.git"},
		{"ssh-url", "ssh://git@example.com/team/dep.git", "ssh://git@example.com/team/dep.git"},
		{"ssh-scp", "git@example.com:team/dep.git", "git@example.com:team/dep.git"},
		{"canonical", "example.com/team/dep", ""},
	} {
		for _, kind := range []string{contextlock.KindContext, contextlock.KindSkill, contextlock.KindMCP} {
			t.Run(tc.name+"/"+kind, func(t *testing.T) {
				req := contextpkg.Requirement{Git: tc.declared, Tag: "v1.0.0", Directory: "nested"}
				m := &contextpkg.Manifest{Contexts: map[string]contextpkg.Requirement{}, Skills: map[string]contextpkg.Requirement{}, MCP: map[string]contextpkg.Requirement{}}
				switch kind {
				case contextlock.KindContext:
					m.Contexts["dep"] = req
				case contextlock.KindSkill:
					m.Skills["dep"] = req
				case contextlock.KindMCP:
					m.MCP["dep"] = req
				}
				got := packageOf(m).Requires[0]
				want := tc.endpoint
				if want == "" {
					want = tc.declared
				}
				if got.Source != want || got.Tag != req.Tag || got.Directory != req.Directory {
					t.Fatalf("safe declaration/metadata: %+v; want %s", got, want)
				}
				for _, source := range []string{got.Source, tc.declared} {
					manager := newGitManager(t.TempDir())
					key, err := manager.Identity(kind, "dep", source)
					if err != nil || key != "example.com/team/dep" {
						t.Fatalf("identity: %q %v", key, err)
					}
					if endpoint, exists := manager.fetchRaw[key]; endpoint != tc.endpoint || exists != (tc.endpoint != "") {
						t.Fatalf("acquisition endpoint: %q present=%t; want %q", endpoint, exists, tc.endpoint)
					}
				}
			})
		}
	}
}

// Refused grammar stays refused before acquisition, including empty passwords
// and percent-encoded usernames. Do not sanitize a password into valid input.
func TestHTTPUserinfoRefusals(t *testing.T) {
	for _, declaration := range []string{
		"https://user:synthetic-password@example.com/team/dep.git",
		"http://user:@example.com/team/dep.git",
		"https://synthetic%2Duser@example.com/team/dep.git",
		"https://user@example.com:443/team/dep.git",
		"https://user@example.com/team/dep.git?query",
		"https://user@example.com/team/dep.git#fragment",
	} {
		t.Run(declaration, func(t *testing.T) {
			manager := newGitManager(t.TempDir())
			if _, err := manager.Identity("context", "dep", declaration); err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
				t.Fatalf("direct identity must refuse: %v", err)
			}
			pkg := packageOf(&contextpkg.Manifest{Contexts: map[string]contextpkg.Requirement{"dep": {Git: declaration}}})
			if _, err := manager.Identity("context", "dep", pkg.Requires[0].Source); err == nil {
				t.Fatal("declaration sanitizing bypassed source refusal")
			}
			if len(manager.fetchRaw) != 0 {
				t.Fatalf("refused source recorded: %+v", manager.fetchRaw)
			}
		})
	}
}
