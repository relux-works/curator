package gitops

import (
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/identity"
)

// FetchURL refuses an empty or option-shaped endpoint before running git, and
// the arguments it builds name the endpoint as a positional with the same
// head and tag refspecs as FetchURLIsolated: no configuration override, and
// so no credential helper, is added to the operator's transport.
func TestFetchURLArgs(t *testing.T) {
	repo := t.TempDir()
	for _, url := range []string{"", "   ", "--upload-pack=touch /tmp/x", "-c"} {
		if err := FetchURL(repo, url); err == nil || !strings.HasPrefix(err.Error(), "refusing to fetch") {
			t.Fatalf("FetchURL(%q) err = %v, want a refusal", url, err)
		}
	}
	if _, err := fetchURLArgs(repo, "git@example.com:team/profiles.git"); err == nil {
		t.Fatal("a directory that is not a repository was accepted")
	}
	if _, err := run(repo, "init", "-q"); err != nil {
		t.Fatal(err)
	}
	args, err := fetchURLArgs(repo, " git@example.com:team/profiles.git ")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"fetch", "--force", "--prune", "--no-write-fetch-head", "--", "git@example.com:team/profiles.git",
		"+refs/heads/*:refs/remotes/origin/*", "+refs/tags/*:refs/tags/*"}
	if !slices.Equal(args, want) {
		t.Fatalf("args = %q, want %q", args, want)
	}
	for _, arg := range args {
		if arg == "-c" || strings.Contains(arg, "credential") {
			t.Fatalf("args carry a configuration override: %q", args)
		}
	}
}

// FetchURL admits the descriptor contract's network endpoints and refuses,
// before git runs and without echoing the input, the ones that carry
// credentials or a malformed authority. A fake git on PATH records every
// invocation: each refused case leaves the marker empty and keeps the canary
// out of the error, and each admitted case reaches git. Refused: remote-helper
// transports of every spelling, passwords and user information outside ssh
// and scp, percent-encoding, whitespace and newlines, empty and extra-slash
// authorities, queries, fragments, opaque and option-shaped spellings, file
// and local forms, a scheme-named scp host without a user, out-of-range or
// empty ports, bracketed hosts that are not plain IPv6 addresses (IPv4-mapped
// included), scp addresses where git's split and the identity parser disagree
// on the user or host (`user:secret@h:path`, where git contacts host `user`;
// a bracketed scp host the identity parser does not read), and "@[" in a scp
// path, which git itself cannot use ("no path specified"). Admitted: https, http,
// git:// and ssh in any scheme case, '@', Unicode and inner '-' in a path,
// single-character and explicit-user scheme-named scp hosts, trailing-dot and
// IDN hosts, ports 1-65535 and IPv6 literals.
func TestFetchURLAdmitsTheDescriptorContract(t *testing.T) {
	const c = "canary-secret-7f3a"
	bin := t.TempDir()
	marker := filepath.Join(t.TempDir(), "git-ran")
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte("#!/bin/sh\necho ran >> \"$GIT_MARKER\"\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("GIT_MARKER", marker)
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	refused := []string{
		"", " ", "-" + c, "--upload-pack=" + c,
		"https::https://user:" + c + "@example.invalid/team/repo.git",
		"http::https://example.invalid/" + c + ".git",
		"ext::sh -c " + c, "fd::7", "file::/tmp/" + c, "x::" + c, "1::" + c, "a_b::" + c,
		"https://user:" + c + "@example.invalid/team/repo.git",
		"https://" + c + "@example.invalid/team/repo.git",
		"https://user%3A" + c + "@example.invalid/team/repo.git",
		"http://user:" + c + "@example.invalid/team/repo.git",
		"git://user:" + c + "@example.invalid/team/repo.git",
		"git://" + c + "@example.invalid/team/repo.git",
		"file:///tmp/" + c + ".git",
		"ftp://example.invalid/" + c + ".git",
		"ssh://git:" + c + "@example.invalid/team/repo.git",
		"ssh://git:" + c + "@[::1]/team/repo.git",
		"git:" + c + "@example.invalid:team/repo.git",
		"user:" + c + "@[::1]:team/repo.git",
		"https:opaque-" + c,
		"https:///user:" + c + "@example.invalid/team/repo.git",
		"https:////user:" + c + "@example.invalid/team/repo.git",
		"https://example.invalid/team/repo.git?password=" + c,
		"https://example.invalid/team/repo.git?access_token=" + c,
		"https://example.invalid/team/repo.git?next=https://user:" + c + "@example.invalid/",
		"https://example.invalid/team/repo.git#" + c,
		"https://example.invalid/team/%2F" + c,
		"https://example.invalid/team\\" + c,
		" https://example.invalid/" + c + ".git",
		"https://example.invalid/" + c + ".git ",
		"https://example.invalid/team/repo.git\n" + c,
		"https://exa_mple.invalid/" + c + ".git",
		"https://example.invalid",
		"https://example.invalid/",
		"https://example.invalid:0/" + c + ".git",
		"https://example.invalid:65536/" + c + ".git",
		"https://example.invalid:99999/" + c + ".git",
		"https://example.invalid:/" + c + ".git",
		"https://[.]/" + c + ".git",
		"https://[:::]/" + c + ".git",
		"https://[abcd]/" + c + ".git",
		"https://[127.0.0.1]/" + c + ".git",
		"ssh://git@[abcd]/" + c + ".git",
		"git@[127.0.0.1]:" + c + ".git",
		"[:::]:" + c + ".git",
		"user:" + c + "@example.invalid:org/repo.git",
		"x:" + c + "@example.invalid:22:org/repo.git",
		"git:" + c + "@example.invalid:22:org/repo.git",
		"git@example.invalid:org/r@[" + c + "].git",
		"example.invalid:org/r@[" + c + "].git",
		"git@[::1]:team/repo.git",
		"https://[::1]:/" + c + ".git",
		"ssh://git@[::1]:/" + c + ".git",
		"https://[::ffff:127.0.0.1]/" + c + ".git",
		"https://[::ffff:7f00:1]/" + c + ".git",
		"ssh:" + c,
		"https:" + c,
		"C:/" + c,
		"C:\\" + c,
		"/tmp/" + c + ".git",
		"./" + c,
	}
	for _, url := range refused {
		err := FetchURL(repo, url)
		if err == nil || !strings.HasPrefix(err.Error(), "refusing to fetch") || strings.Contains(err.Error(), c) {
			t.Errorf("FetchURL(%q) err = %v, want an input-free refusal", url, err)
		}
		if _, statErr := os.Stat(marker); statErr == nil {
			t.Fatalf("FetchURL(%q) ran git before refusing", url)
		}
	}
	for _, url := range []string{
		"https://example.invalid/team/repo.git",
		"HTTPS://example.invalid/team/repo.git",
		"http://example.invalid/team/repo.git",
		"git://example.invalid/team/repo.git",
		"SSH://git@example.invalid/team/repo.git",
		"https://example.invalid/team/r@name.git",
		"https://example.invalid/team/репо.git",
		"https://example.invalid/team/repo!name.git",
		"https://example.invalid/-team/repo.git",
		"https://example.invalid//team/repo.git",
		"https://example.invalid/team/repo.git/",
		"https://example.invalid./team/repo.git",
		"https://xn--bcher-kva.example/team/repo.git",
		"https://example.invalid:8443/team/repo.git",
		"https://example.invalid:1/team/repo.git",
		"https://example.invalid:65535/team/repo.git",
		"https://[::1]/team/repo.git",
		"https://[2001:db8::1]:8443/team/repo.git",
		"ssh://git@example.invalid/team/repo.git",
		"ssh://git@example.invalid:2222/team/repo.git",
		"ssh://example.invalid/team/repo.git",
		"ssh://git@[::1]/team/repo.git",
		"ssh://foo;bar@example.invalid/team/repo.git",
		"ssh://foo@bar@example.invalid/team/repo.git",
		"git@example.invalid:team/repo.git",
		"git@ssh:team/repo.git",
		"foo[bar]@example.invalid:org/repo.git",
		"git@example.invalid:org/r@name.git",
		"github-work:team/repo.git",
		"git@github-work:Org/nested/repo.git",
		"c:team/repo.git",
		"git@c:team/repo.git",
		"example.invalid:/srv/team/repo.git",
		"example.invalid:-team/repo.git",
	} {
		_ = os.Remove(marker)
		if err := FetchURL(repo, url); err != nil && strings.HasPrefix(err.Error(), "refusing to fetch") {
			t.Fatalf("FetchURL(%q) refused an admitted endpoint: %v", url, err)
		}
		if _, err := os.Stat(marker); err != nil {
			t.Fatalf("FetchURL(%q) did not reach git", url)
		}
	}
}

// What git actually contacts over ssh is what the identity binding names. A
// recording ssh on GIT_SSH captures the destination argument of a real git
// fetch for every admitted ssh:// and scp-like descriptor family; its user
// and host equal the identity parser's, and no argument carries the planted
// secret. The scp spellings where git and the identity parser disagree refuse
// before git, and the recording ssh never runs for them.
func TestFetchURLSSHDestinationMatchesIdentity(t *testing.T) {
	const secret = "planted-secret-91c2"
	bin := t.TempDir()
	log := filepath.Join(t.TempDir(), "ssh.log")
	script := "#!/bin/sh\nfor a in \"$@\"; do printf '%s\\n' \"$a\"; done > \"$SSH_LOG\"\nexit 1\n"
	if err := os.WriteFile(filepath.Join(bin, "ssh"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_SSH", filepath.Join(bin, "ssh"))
	t.Setenv("SSH_LOG", log)
	repo := t.TempDir()
	if _, err := run(repo, "init", "-q"); err != nil {
		t.Fatal(err)
	}
	for _, endpoint := range []string{
		"ssh://git@github.com/Org/re.po_name-1.git",
		"ssh://git@github-work/Org/nested/re.po_name-1.git",
		"SSH://git@example.invalid/org/repo.git",
		"ssh://foo;bar@example.invalid/org/repo.git",
		"ssh://foo@bar@example.invalid/org/repo.git",
		"ssh://git@example.invalid:2222/org/repo.git",
		"git@github.com:Org/re.po_name-1.git",
		"git@github-work:Org/nested/re.po_name-1.git",
		"github-work:Org/repo.git",
		"c:org/repo.git",
		"git@c:org/repo.git",
		"git@ssh:org/repo.git",
		"git@example.invalid:/org/repo.git",
		"foo[bar]@example.invalid:org/repo.git",
		"git@example.invalid:org/r@name.git",
	} {
		_ = os.Remove(log)
		if err := FetchURL(repo, endpoint); err != nil && strings.HasPrefix(err.Error(), "refusing to fetch") {
			t.Fatalf("%q refused: %v", endpoint, err)
		}
		recorded, err := os.ReadFile(log)
		if err != nil {
			t.Fatalf("%q: git did not run ssh: %v", endpoint, err)
		}
		args := strings.Split(strings.TrimSpace(string(recorded)), "\n")
		destination := ""
		for i, arg := range args {
			if strings.HasPrefix(arg, "git-upload-pack ") && i > 0 {
				destination = args[i-1]
			}
		}
		at := strings.LastIndex(destination, "@")
		gotUser, gotHost := "", destination
		if at >= 0 {
			gotUser, gotHost = destination[:at], destination[at+1:]
		}
		wantUser, wantHost := identityUserHost(t, endpoint)
		if gotUser != wantUser || !strings.EqualFold(gotHost, wantHost) {
			t.Errorf("%q: git contacted user %q host %q, the identity binding names user %q host %q", endpoint, gotUser, gotHost, wantUser, wantHost)
		}
		if strings.Contains(string(recorded), secret) {
			t.Errorf("%q: the ssh argv carries the secret", endpoint)
		}
	}
	for _, endpoint := range []string{
		"user:" + secret + "@example.invalid:org/repo.git",
		"x:" + secret + "@example.invalid:22:org/repo.git",
		"git:" + secret + "@example.invalid:22:org/repo.git",
	} {
		_ = os.Remove(log)
		err := FetchURL(repo, endpoint)
		if err == nil || !strings.HasPrefix(err.Error(), "refusing to fetch") || strings.Contains(err.Error(), secret) {
			t.Fatalf("%q: err = %v, want an input-free refusal", endpoint, err)
		}
		if _, err := os.Stat(log); err == nil {
			t.Fatalf("%q: ssh ran for a refused endpoint", endpoint)
		}
	}
}

// identityUserHost is the user and host the identity binding reads.
func identityUserHost(t *testing.T, endpoint string) (string, string) {
	t.Helper()
	if strings.Contains(endpoint, "://") {
		parsed, err := url.Parse(endpoint)
		if err != nil {
			t.Fatal(err)
		}
		return parsed.User.Username(), parsed.Hostname()
	}
	user, host, ok := identity.SCPUserHost(endpoint)
	if !ok {
		t.Fatalf("%q: not a scp-like address the identity parser reads", endpoint)
	}
	return user, host
}
