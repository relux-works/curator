package buildrepo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/gitcred"
)

// TestSemanticFallbackGateDecidesAllPublishedCases drives the §2 fallback
// gate through every published fallback/pin semantic case: the two
// fallback-eligible first failures advance, and every fail-closed row —
// TLS, host-key, integrity, identity, ref, audit, unknown, unreadable
// policy, ambiguous 404 — stops with no second attempt.
func TestSemanticFallbackGateDecidesAllPublishedCases(t *testing.T) {
	published := []struct {
		id       string
		expected string
	}{
		{"fallback-dns", "attempt-second"},
		{"fallback-auth-rejected", "attempt-second"},
		{"fallback-tls", "stop-no-second-attempt"},
		{"fallback-host-key", "stop-no-second-attempt"},
		{"fallback-integrity", "stop-no-second-attempt"},
		{"fallback-identity", "stop-no-second-attempt"},
		{"fallback-ref-moved", "stop-no-second-attempt"},
		{"fallback-audit", "stop-no-second-attempt"},
		{"fallback-unknown", "stop-no-second-attempt"},
		{"fallback-policy-unreadable", "stop-no-second-attempt"},
		{"fallback-http-404", "stop-no-second-attempt"},
	}
	for _, publishedCase := range published {
		t.Run(publishedCase.id, func(t *testing.T) {
			token := strings.TrimPrefix(publishedCase.id, "fallback-")
			class, ok := SemanticFailureClass(token)
			if !ok {
				t.Fatalf("semantic token %q has no failure class", token)
			}
			advance := AllowSecondAttempt(TransportFallbackAvailabilityAuth, class)
			want := publishedCase.expected == "attempt-second"
			if advance != want {
				t.Fatalf("gate(%q) advance = %v, want %v", token, advance, want)
			}
		})
	}
	// pinned-auth: pin forces FallbackNone at plan time, so even an
	// auth rejection stops. fallback:none never advances by class.
	if AllowSecondAttempt(TransportFallbackNone, FailureAuth) {
		t.Fatal("fallback none advanced past an auth rejection")
	}
	if AllowSecondAttempt(TransportFallbackNone, FailureAvailability) {
		t.Fatal("fallback none advanced past an availability failure")
	}
	if AllowSecondAttempt("sometimes", FailureAuth) {
		t.Fatal("unknown fallback mode advanced past an auth rejection")
	}
	if _, ok := SemanticFailureClass("no-such-token"); ok {
		t.Fatal("unknown semantic token mapped to a class")
	}
}

func TestClassifyFetchOutput(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		stderr string
		want   FailureClass
	}{
		{"dns", "fatal: unable to access 'https://git.example.org/kit.git/': Could not resolve host: git.example.org", FailureAvailability},
		{"dns-temporary", "ssh: Could not resolve hostname git.example.org: Temporary failure in name resolution", FailureAvailability},
		{"dns-unknown-host", "ssh: Could not resolve hostname git.example.org: Name or service not known", FailureAvailability},
		{"dns-nodename", "ssh: Could not resolve hostname git.example.org: nodename nor servname provided, or not known", FailureAvailability},
		{"connection-refused", "ssh: connect to host git.example.org port 22: Connection refused", FailureAvailability},
		{"ssh-connect-timeout", "ssh: connect to host git.example.org port 22: Connection timed out", FailureAvailability},
		{"ssh-connect-unreachable", "ssh: connect to host git.example.org port 22: Network is unreachable", FailureAvailability},
		{"connection-timeout", "fatal: unable to access 'https://git.example.org/kit.git/': Operation timed out after 30000 milliseconds with 0 bytes received", FailureAvailability},
		{"network-unreachable", "fatal: unable to access 'https://git.example.org/kit.git/': Failed to connect to git.example.org: Network is unreachable", FailureAvailability},
		{"connect-no-route", "fatal: unable to access 'https://git.example.org/kit.git/': Failed to connect to git.example.org: No route to host", FailureAvailability},
		{"connect-refused", "fatal: unable to access 'https://git.example.org/kit.git/': Failed to connect to git.example.org: Connection refused", FailureAvailability},
		{"http-502", "fatal: unable to access 'https://git.example.org/kit.git/': The requested URL returned error: 502", FailureAvailability},
		{"http-503", "error: RPC failed; HTTP 503 curl 22 The requested URL returned error: 503", FailureAvailability},
		{"http-504", "fatal: unable to access 'https://git.example.org/kit.git/': The requested URL returned error: 504", FailureAvailability},
		{"http-401", "fatal: unable to access 'https://git.example.org/kit.git/': The requested URL returned error: 401", FailureAuth},
		{"http-403", "fatal: unable to access 'https://git.example.org/kit.git/': The requested URL returned error: 403", FailureAuth},
		{"https-auth-failed", "fatal: Authentication failed for 'https://git.example.org/kit.git/'", FailureAuth},
		{"ssh-permission-denied", "git@git.example.org: Permission denied (publickey).", FailureAuth},
		{"remote-invalid-password", "remote: Invalid username or password.", FailureAuth},
		{"remote-permission-denied", "remote: Permission to acme/kit.git denied to deploy-key.", FailureAuth},
		{"tls-issuer", "fatal: unable to access 'https://git.example.org/kit.git/': SSL certificate problem: unable to get local issuer certificate", FailureTLS},
		{"tls-bare", "fatal: SSL certificate problem: unable to get local issuer certificate", FailureTLS},
		{"tls-self-signed", "fatal: unable to access 'https://git.example.org/kit.git/': SSL certificate problem: self signed certificate", FailureTLS},
		{"tls-schannel", "fatal: unable to access 'https://git.example.org/kit.git/': schannel: failed to receive handshake, SSL/TLS connection failed", FailureTLS},
		{"tls-beats-timeout", "fatal: unable to access 'https://git.example.org/kit.git/': OpenSSL SSL_connect: Connection timed out", FailureTLS},
		{"host-key", "Host key verification failed.", FailureHostKey},
		{"host-key-changed", "WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!", FailureHostKey},
		{"host-key-beats-auth", "Host key verification failed.\nPermission denied (publickey).", FailureHostKey},
		{"redirect-forbidden", "fatal: unable to access 'https://git.example.org/kit.git/': Redirection from https://git.example.org/kit.git to https://evil.example.org/kit.git is forbidden", FailureIdentity},
		{"remote-ref-missing", "fatal: couldn't find remote ref refs/heads/no-such-branch", FailureRef},
		{"http-404", "fatal: unable to access 'https://git.example.org/kit.git/': The requested URL returned error: 404", FailureHTTP404},
		{"github-repository-not-found", "remote: Repository not found.", FailureHTTP404},
		{"empty", "", FailureUnknown},
		{"unfamiliar", "fatal: something entirely new from a future git", FailureUnknown},
		{"bare-numbers-are-not-statuses", "remote: Enumerating objects: 1502, done.\nremote: Total 502 (delta 0), reused 0", FailureUnknown},
		{"reset-is-not-refused", "fatal: unable to access 'https://git.example.org/kit.git/': Recv failure: Connection reset by peer", FailureUnknown},
		{"ambiguous-view-grant", "remote: The project you were looking for could not be found or you don't have permission to view it.", FailureUnknown},
		{"access-denied-stays-unknown", "fatal: unable to access 'https://git.example.org/kit.git/': Access denied", FailureUnknown},
		{"secret-tail-poisons-to-unknown", "fatal: unable to access 'https://x/': Could not resolve host: x (saw SUPERSECRETPROVIDER9)", FailureUnknown},
		{"bare-permission-denied-is-local", "fatal: cannot create temporary file: Permission denied", FailureUnknown},
		{"os-error-paren-stays-unknown", "fatal: cannot create directory: Permission denied (os error 13)", FailureUnknown},
		{"ssh-method-list-still-auth", "Permission denied (publickey,keyboard-interactive).", FailureAuth},
		{"integrity-beats-timeout", "error: object hash mismatch\nfatal: connection timed out", FailureIntegrity},
		{"integrity-corrupt", "fatal: fsck error: object abc is corrupted", FailureIntegrity},
		{"integrity-checksum", "error: checksum mismatch for object 0123456789abcdef", FailureIntegrity},
		{"audit-beats-timeout", "remote: audit denied: operation timed out", FailureAudit},
		{"audit-revoked", "remote: key is revoked: access denied", FailureAudit},
		{"audit-canary", "remote: canary check failed for this repository", FailureAudit},
		{"audit-beats-auth", "remote: assurance failure: invalid credentials presented", FailureAudit},
		{"https-no-username", "fatal: could not read Username for 'https://git.example.org/kit.git/': terminal prompts disabled", FailureAuth},
		{"https-no-password", "fatal: could not read Password for 'https://git.example.org/kit.git/': terminal prompts disabled", FailureAuth},
		{"https-no-tty", "fatal: could not read Username for 'https://git.example.org/kit.git/': No such device or address", FailureAuth},
		{"ssh-no-route", "ssh: connect to host git.example.org port 22: No route to host", FailureAvailability},
		{"unexpected-protocol-response", "fatal: unexpected protocol response", FailureUnknown},
		{"unreadable-object-is-unknown", "fatal: unable to read object 0123456789012345678901234567890123456789", FailureUnknown},
		{"unknown-tail-poisons-dns", "fatal: unable to access 'https://git.example.org/kit.git/': Could not resolve host: git.example.org\nfatal: unexpected protocol response", FailureUnknown},
		{"remote-progress-tail-poisons-dns", "fatal: unable to access 'https://git.example.org/kit.git/': Could not resolve host: git.example.org\nremote: Enumerating objects: 3, done.", FailureUnknown},
		{"object-failure-poisons-timeout", "fatal: unable to read object 0123456789012345678901234567890123456789\nfatal: connection timed out", FailureUnknown},
		{"quoted-signal-filename-is-local", "fatal: cannot create temporary file 'connection timed out': Permission denied", FailureUnknown},
		{"forge-timeout-words-are-not-endpoint", "remote: Connection timed out", FailureUnknown},
		{"bare-503-is-not-status", "remote: Total 503 (delta 0), reused 0", FailureUnknown},
		{"unknown-facility-poisons-dns", "leaked-secret: SUPERSECRETPROVIDER9\nfatal: unable to access 'https://git.example.org/kit.git/': Could not resolve host: git.example.org", FailureUnknown},
		{"hint-line-poisons-dns", "fatal: unable to access 'https://git.example.org/kit.git/': Could not resolve host: git.example.org\nhint: check your network connection", FailureUnknown},
		{"ssh-warning-poisons-auth", "Warning: Permanently added 'git.example.org' (ED25519) to the list of known hosts.\ngit@git.example.org: Permission denied (publickey).", FailureUnknown},
		{"audit-facility-poisons-dns", "fatal: unable to access 'https://git.example.org/kit.git/': Could not resolve host: git.example.org\naudit: policy denied", FailureUnknown},
		{"http-503-tail-poisons", "fatal: unable to access 'https://git.example.org/kit.git/': The requested URL returned error: 503; object verification failed", FailureUnknown},
		{"ssh-resolver-tail-poisons", "ssh: Could not resolve hostname git.example.org: unexpected resolver protocol response", FailureUnknown},
		{"crlf-dns", "fatal: unable to access 'https://git.example.org/kit.git/': Could not resolve host: git.example.org\r", FailureAvailability},
		{"ssh-trailer-ignored", "ssh: connect to host git.example.org port 22: Connection refused\nfatal: Could not read from remote repository.\n\nPlease make sure you have the correct access rights\nand the repository exists.", FailureAvailability},
		{"rpc-hung-up-ignored", "error: RPC failed; HTTP 503 curl 22 The requested URL returned error: 503\nfatal: the remote end hung up unexpectedly", FailureAvailability},
		{"lone-trailer-is-unknown", "fatal: Could not read from remote repository.", FailureUnknown},
		{"lone-hung-up-is-unknown", "fatal: the remote end hung up unexpectedly", FailureUnknown},
		{"terminated-trailer-splits", "ssh: connect to host git.example.org port 22: Connection refused\nfatal: Could not read from remote repository.", FailureAvailability},
		{"glued-trailer-refuses", "ssh: connect to host git.example.org port 22: Connection refusedfatal: Could not read from remote repository.", FailureUnknown},
		{"glued-unknown-stays-unknown", "ssh: connect to host git.example.org port 22: Connection refusedfatal: unexpected protocol response", FailureUnknown},
		{"glued-503-hung-up-refuses", "fatal: unable to access 'https://git.example.org/kit.git/': The requested URL returned error: 503fatal: the remote end hung up unexpectedly", FailureUnknown},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if got := ClassifyFetchOutput(testCase.stderr); got != testCase.want {
				t.Fatalf("ClassifyFetchOutput = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestClassifyAdmissionCode(t *testing.T) {
	for code, want := range map[string]FailureClass{
		CodeRefMoved:                               FailureRef,
		CodeIncompleteSource:                       FailureIntegrity,
		CodeObjectSemanticsInvalid:                 FailureIntegrity,
		CodeLFSUnsupported:                         FailureIntegrity,
		CodeIdentityInvalid:                        FailureIdentity,
		CodeLocalLayoutUnsafe:                      FailureIdentity,
		CodeSSHCredentialMissing:                   FailureAuth,
		CodeAuditBlocked:                           FailureAudit,
		CodeDescriptorInvalid:                      FailureIntegrity,
		CodeTransportResolutionUnsupportedPlatform: FailureUnknown,
		CodeSourceUnavailable:                      FailureUnknown,
		CodeUnverifiedOffline:                      FailureIntegrity,
		"no-such-code":                             FailureUnknown,
		"":                                         FailureUnknown,
	} {
		if got := ClassifyAdmissionCode(code); got != want {
			t.Errorf("ClassifyAdmissionCode(%q) = %q, want %q", code, got, want)
		}
	}
}

func TestValidateTransportPlan(t *testing.T) {
	valid := []TransportPlan{
		{Identity: "fixture.test/repository", Attempts: []TransportAttempt{{URL: "https://fixture.test/repository.git"}}, Fallback: TransportFallbackNone},
		{Identity: "fixture.test/repository", Attempts: []TransportAttempt{
			{URL: "https://fixture.test/repository.git", Authentication: "https-prov"},
			{URL: "ssh://git@fixture.test/repository.git", Authentication: "ssh-prov"},
		}, Fallback: TransportFallbackAvailabilityAuth},
		{Identity: "fixture.test/repository", Attempts: []TransportAttempt{
			{URL: "git@fixture.test:repository.git", Authentication: "ssh-prov"},
			{URL: "https://fixture.test/repository", Authentication: "https-prov"},
		}, Fallback: TransportFallbackNone},
	}
	for index, plan := range valid {
		if err := ValidateTransportPlan(plan); err != nil {
			t.Errorf("valid plan %d: %v", index, err)
		}
	}
	invalid := []struct {
		name string
		plan TransportPlan
	}{
		{"no-attempts", TransportPlan{Identity: "fixture.test/repository", Fallback: TransportFallbackNone}},
		{"three-attempts", TransportPlan{Identity: "fixture.test/repository", Attempts: []TransportAttempt{
			{URL: "https://fixture.test/repository.git"},
			{URL: "ssh://git@fixture.test/repository.git"},
			{URL: "git@fixture.test:repository.git"},
		}, Fallback: TransportFallbackNone}},
		{"unknown-fallback", TransportPlan{Identity: "fixture.test/repository", Attempts: []TransportAttempt{{URL: "https://fixture.test/repository.git"}}, Fallback: "sometimes"}},
		{"empty-fallback", TransportPlan{Identity: "fixture.test/repository", Attempts: []TransportAttempt{{URL: "https://fixture.test/repository.git"}}}},
		{"empty-identity", TransportPlan{Attempts: []TransportAttempt{{URL: "https://fixture.test/repository.git"}}, Fallback: TransportFallbackNone}},
		{"bad-grammar", TransportPlan{Identity: "fixture.test/repository", Attempts: []TransportAttempt{{URL: "https://fixture.test/has space.git"}}, Fallback: TransportFallbackNone}},
		{"file-protocol", TransportPlan{Identity: "fixture.test/repository", Attempts: []TransportAttempt{{URL: "file:///tmp/repository.git"}}, Fallback: TransportFallbackNone}},
		{"identity-mismatch", TransportPlan{Identity: "fixture.test/repository", Attempts: []TransportAttempt{{URL: "https://evil.test/repository.git"}}, Fallback: TransportFallbackNone}},
		{"cross-identity-pair", TransportPlan{Identity: "fixture.test/repository", Attempts: []TransportAttempt{
			{URL: "https://fixture.test/repository.git"},
			{URL: "https://fixture.test/other.git"},
		}, Fallback: TransportFallbackAvailabilityAuth}},
		{"duplicate-urls", TransportPlan{Identity: "fixture.test/repository", Attempts: []TransportAttempt{
			{URL: "https://fixture.test/repository.git", Authentication: "one"},
			{URL: "https://fixture.test/repository.git", Authentication: "two"},
		}, Fallback: TransportFallbackAvailabilityAuth}},
		{"command-provider", TransportPlan{Identity: "fixture.test/repository", Attempts: []TransportAttempt{{URL: "https://fixture.test/repository.git", Authentication: "/bin/sh"}}, Fallback: TransportFallbackNone}},
		{"path-provider", TransportPlan{Identity: "fixture.test/repository", Attempts: []TransportAttempt{{URL: "https://fixture.test/repository.git", Authentication: "../escape"}}, Fallback: TransportFallbackNone}},
		{"space-provider", TransportPlan{Identity: "fixture.test/repository", Attempts: []TransportAttempt{{URL: "https://fixture.test/repository.git", Authentication: "team ssh"}}, Fallback: TransportFallbackNone}},
	}
	for _, testCase := range invalid {
		t.Run(testCase.name, func(t *testing.T) {
			err := ValidateTransportPlan(testCase.plan)
			if err == nil {
				t.Fatal("invalid plan validated")
			}
			if ErrorCode(err) != CodeRepositoryPolicyInvalid {
				t.Fatalf("error = %v, want %s", err, CodeRepositoryPolicyInvalid)
			}
		})
	}
}

// stubProviderReader is an in-memory trusted broker for provider tests.
type stubProviderReader struct {
	credentials map[string]gitcred.HostCredential
	calls       int
}

func (s *stubProviderReader) ReadProvider(_ context.Context, provider, host string) (gitcred.HostCredential, bool) {
	s.calls++
	credential, ok := s.credentials[provider+"\x00"+host]
	return credential, ok
}

func TestCredentialProvidersResolveHTTPS(t *testing.T) {
	ctx := context.Background()
	reader := &stubProviderReader{credentials: map[string]gitcred.HostCredential{
		"team-https\x00git.example.org": {Username: "broker-user", Secret: "broker-secret"},
	}}
	set := ProviderSet{
		"team-https": {HTTPS: &ProviderHTTPS{}},
		"named-user": {HTTPS: &ProviderHTTPS{Username: "entry-user"}},
		"anon":       {HTTPS: &ProviderHTTPS{Anonymous: true}},
		"ssh-only":   {SSH: &ProviderSSH{}},
	}
	providers := CredentialProviders{Set: set, Reader: reader}

	resolved, err := providers.ResolveHTTPS(ctx, "team-https", "git.example.org")
	if err != nil {
		t.Fatalf("ResolveHTTPS: %v", err)
	}
	if !resolved.Selected() || resolved.Host != "git.example.org" || resolved.Username != "broker-user" || resolved.secret != "broker-secret" {
		t.Fatalf("ResolveHTTPS = %+v", resolved)
	}
	// The broker holds no named-user entry: the entry username alone
	// authenticates nothing.
	if _, err := providers.ResolveHTTPS(ctx, "named-user", "git.example.org"); !errors.Is(err, ErrProviderUnavailable) {
		t.Fatalf("ResolveHTTPS without broker material = %v, want unavailable", err)
	}
	reader.credentials["named-user\x00git.example.org"] = gitcred.HostCredential{Username: "broker-user", Secret: "s"}
	resolved, err = providers.ResolveHTTPS(ctx, "named-user", "git.example.org")
	if err != nil || resolved.Username != "entry-user" {
		t.Fatalf("ResolveHTTPS named-user = %+v, %v", resolved, err)
	}

	before := reader.calls
	if _, err := providers.ResolveHTTPS(ctx, "anon", "git.example.org"); !errors.Is(err, ErrProviderAnonymous) {
		t.Fatalf("anonymous provider = %v, want anonymous", err)
	}
	if reader.calls != before {
		t.Fatal("anonymous resolution touched the credential broker")
	}
	for _, provider := range []string{"missing", "ssh-only", "../escape", ""} {
		if _, err := providers.ResolveHTTPS(ctx, provider, "git.example.org"); !errors.Is(err, ErrProviderUnavailable) {
			t.Fatalf("ResolveHTTPS(%q) = %v, want unavailable", provider, err)
		}
	}
	brokerless := CredentialProviders{Set: set}
	if _, err := brokerless.ResolveHTTPS(ctx, "team-https", "git.example.org"); !errors.Is(err, ErrProviderUnavailable) {
		t.Fatalf("brokerless ResolveHTTPS = %v, want unavailable", err)
	}
	var nilSet CredentialProviders
	if _, err := nilSet.ResolveHTTPS(ctx, "team-https", "git.example.org"); !errors.Is(err, ErrProviderUnavailable) {
		t.Fatalf("nil-set ResolveHTTPS = %v, want unavailable", err)
	}
}

func TestCredentialProvidersResolveSSH(t *testing.T) {
	identity := filepath.Join(t.TempDir(), "id_example")
	knownHosts := filepath.Join(t.TempDir(), "known_hosts")
	if err := os.WriteFile(identity, []byte("key"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(knownHosts, []byte("hosts"), 0o600); err != nil {
		t.Fatal(err)
	}
	set := ProviderSet{
		"team-ssh":   {SSH: &ProviderSSH{Identity: identity, KnownHosts: knownHosts}},
		"dangling":   {SSH: &ProviderSSH{Identity: filepath.Join(t.TempDir(), "absent"), KnownHosts: knownHosts}},
		"relative":   {SSH: &ProviderSSH{Identity: "relative/path", KnownHosts: knownHosts}},
		"hosts-only": {SSH: &ProviderSSH{KnownHosts: knownHosts}},
		"https-only": {HTTPS: &ProviderHTTPS{Anonymous: true}},
	}
	providers := CredentialProviders{Set: set}
	resolved, err := providers.ResolveSSH("team-ssh")
	if err != nil {
		t.Fatalf("ResolveSSH: %v", err)
	}
	if !resolved.Selected() || resolved.Identity == "" || resolved.KnownHosts == "" {
		t.Fatalf("ResolveSSH = %+v", resolved)
	}
	for _, provider := range []string{"missing", "dangling", "relative", "hosts-only", "https-only", "../escape", ""} {
		if _, err := providers.ResolveSSH(provider); !errors.Is(err, ErrProviderUnavailable) {
			t.Fatalf("ResolveSSH(%q) = %v, want unavailable", provider, err)
		}
	}
}

// stubAuthProvider is a scripted AuthProvider for executor binding tests.
type stubAuthProvider struct {
	https     map[string]HTTPSCredentials
	httpsErr  map[string]error
	ssh       map[string]OperatorSSHCredentials
	sshErr    map[string]error
	httpsCall []string
	sshCall   []string
}

func (s *stubAuthProvider) ResolveHTTPS(_ context.Context, provider, host string) (HTTPSCredentials, error) {
	s.httpsCall = append(s.httpsCall, provider+"\x00"+host)
	if err, ok := s.httpsErr[provider]; ok {
		return HTTPSCredentials{}, err
	}
	return s.https[provider], nil
}

func (s *stubAuthProvider) ResolveSSH(provider string) (OperatorSSHCredentials, error) {
	s.sshCall = append(s.sshCall, provider)
	if err, ok := s.sshErr[provider]; ok {
		return OperatorSSHCredentials{}, err
	}
	return s.ssh[provider], nil
}

func TestBindProviderAttempt(t *testing.T) {
	ctx := context.Background()
	httpsSource, err := ParseSource("https://fixture.test/repository.git")
	if err != nil {
		t.Fatal(err)
	}
	sshSource, err := ParseSource("ssh://git@fixture.test/repository.git")
	if err != nil {
		t.Fatal(err)
	}
	decoy := GitTool{HTTPSCredentials: NewHTTPSCredentials("fixture.test", "decoy-user", "DECOY-SECRET")}

	t.Run("https-binds-provider-material-over-lane-credentials", func(t *testing.T) {
		stub := &stubAuthProvider{https: map[string]HTTPSCredentials{
			"p": NewHTTPSCredentials("fixture.test", "prov-user", "PROVIDER-SECRET"),
		}}
		bound, skip := bindProviderAttempt(ctx, decoy, httpsSource, "p", stub)
		if skip != nil {
			t.Fatal("provider material skipped")
		}
		if bound.HTTPSCredentials.Username != "prov-user" {
			t.Fatalf("bound username = %q", bound.HTTPSCredentials.Username)
		}
		if len(stub.httpsCall) != 1 || stub.httpsCall[0] != "p\x00fixture.test" {
			t.Fatalf("resolve calls = %q", stub.httpsCall)
		}
	})

	t.Run("https-anonymous-clears-lane-credentials", func(t *testing.T) {
		stub := &stubAuthProvider{httpsErr: map[string]error{"anon": ErrProviderAnonymous}}
		bound, skip := bindProviderAttempt(ctx, decoy, httpsSource, "anon", stub)
		if skip != nil {
			t.Fatal("anonymous provider skipped")
		}
		if bound.HTTPSCredentials.Selected() {
			t.Fatalf("anonymous attempt kept credentials: %+v", bound.HTTPSCredentials)
		}
	})

	t.Run("https-unavailable-skips-without-traffic", func(t *testing.T) {
		stub := &stubAuthProvider{httpsErr: map[string]error{"p": ErrProviderUnavailable}}
		_, skip := bindProviderAttempt(ctx, decoy, httpsSource, "p", stub)
		if skip == nil || skip.class != FailureAuth {
			t.Fatalf("skip = %+v, want auth", skip)
		}
		_, skip = bindProviderAttempt(ctx, decoy, httpsSource, "p", nil)
		if skip == nil || skip.class != FailureAuth {
			t.Fatalf("nil-provider skip = %+v, want auth", skip)
		}
	})

	t.Run("https-unselected-or-foreign-host-skips", func(t *testing.T) {
		stub := &stubAuthProvider{https: map[string]HTTPSCredentials{
			"empty":   {},
			"foreign": NewHTTPSCredentials("other.test", "u", "s"),
		}}
		for _, provider := range []string{"empty", "foreign"} {
			if _, skip := bindProviderAttempt(ctx, decoy, httpsSource, provider, stub); skip == nil {
				t.Fatalf("provider %q bound without usable material", provider)
			}
		}
	})

	t.Run("ssh-binds-selection", func(t *testing.T) {
		stub := &stubAuthProvider{ssh: map[string]OperatorSSHCredentials{
			"p": {Identity: "/abs/id", KnownHosts: "/abs/kh"},
		}}
		bound, skip := bindProviderAttempt(ctx, GitTool{}, sshSource, "p", stub)
		if skip != nil {
			t.Fatal("SSH selection skipped")
		}
		if !bound.SSHCredentials.Selected() || bound.SSHCredentials.Identity != "/abs/id" {
			t.Fatalf("bound SSH = %+v", bound.SSHCredentials)
		}
	})

	t.Run("ssh-unavailable-or-unselected-skips", func(t *testing.T) {
		stub := &stubAuthProvider{
			ssh:    map[string]OperatorSSHCredentials{"empty": {}},
			sshErr: map[string]error{"missing": ErrProviderUnavailable},
		}
		for _, provider := range []string{"empty", "missing"} {
			if _, skip := bindProviderAttempt(ctx, GitTool{}, sshSource, provider, stub); skip == nil || skip.class != FailureAuth {
				t.Fatalf("provider %q skip = %+v, want auth", provider, skip)
			}
		}
	})

	t.Run("unknown-transport-fails-closed", func(t *testing.T) {
		stub := &stubAuthProvider{}
		_, skip := bindProviderAttempt(ctx, GitTool{}, Source{Git: "x", Identity: "fixture.test/repository", Transport: "file"}, "p", stub)
		if skip == nil || skip.class != FailureIdentity {
			t.Fatalf("skip = %+v, want identity", skip)
		}
	})
}

// blockingProvider never answers: resolution returns only when the
// caller's context expires. It proves the shared deadline gates the
// second attempt without any fetch traffic.
type blockingProvider struct{}

func (blockingProvider) ResolveHTTPS(ctx context.Context, _ string, _ string) (HTTPSCredentials, error) {
	<-ctx.Done()
	return HTTPSCredentials{}, ErrProviderUnavailable
}

func (blockingProvider) ResolveSSH(_ string) (OperatorSSHCredentials, error) {
	return OperatorSSHCredentials{}, ErrProviderUnavailable
}

func TestResolvedTransportSharedDeadlineStopsSecondAttempt(t *testing.T) {
	requireResolvedLane(t)
	source, err := ParseSource("https://fixture.test/repository.git")
	if err != nil {
		t.Fatal(err)
	}
	plan := TransportPlan{
		Identity: source.Identity,
		Attempts: []TransportAttempt{
			{URL: "https://fixture.test/repository.git", Authentication: "slow"},
			{URL: "ssh://git@fixture.test/repository.git", Authentication: "slow"},
		},
		Fallback: TransportFallbackAvailabilityAuth,
	}
	base := NetworkRequest{
		Source: source,
		Lock:   LockedCommit{ObjectFormat: "sha1", Hex: strings.Repeat("1", 40)},
		Limits: Limits{Timeout: 200 * time.Millisecond},
	}
	var records []AttemptRecord
	start := time.Now()
	_, err = AcquireNetworkResolved(context.Background(), base, plan, blockingProvider{}, func(record AttemptRecord) {
		records = append(records, record)
	})
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("expired plan succeeded")
	}
	if ErrorCode(err) != CodeRepositoryEndpointUnavailable {
		t.Fatalf("error = %v, want %s", err, CodeRepositoryEndpointUnavailable)
	}
	if !strings.Contains(err.Error(), "not attempted (total deadline expired)") {
		t.Fatalf("error names no expired deadline: %v", err)
	}
	if len(records) != 1 || records[0].NetworkAttempted || records[0].Class != FailureAuth {
		t.Fatalf("records = %+v, want one traffic-free auth record", records)
	}
	if elapsed > 10*time.Second {
		t.Fatalf("elapsed %v exceeds any shared deadline", elapsed)
	}
}

func TestExhaustionErrorIsClosedVocabulary(t *testing.T) {
	plan := TransportPlan{
		Identity: "fixture.test/repository",
		Attempts: []TransportAttempt{
			{URL: "https://fixture.test/repository.git", Authentication: "https-prov"},
			{URL: "ssh://git@fixture.test/repository.git", Authentication: "ssh-prov"},
		},
		Fallback: TransportFallbackAvailabilityAuth,
	}
	records := []AttemptRecord{
		{Index: 1, URL: plan.Attempts[0].URL, Transport: "https", Provider: "https-prov", Class: FailureAvailability, NetworkAttempted: true},
		{Index: 2, URL: plan.Attempts[1].URL, Transport: "ssh", Provider: "ssh-prov", Class: FailureAuth, NetworkAttempted: true},
	}
	message := exhaustionError(plan, records, false).Error()
	for _, want := range []string{
		CodeRepositoryEndpointUnavailable, "fixture.test/repository",
		"endpoint 1 (https", "availability", "endpoint 2 (ssh",
		"provider \"ssh-prov\"", "auth", "source-policy.json",
	} {
		if !strings.Contains(message, want) {
			t.Errorf("exhaustion %q lacks %q", message, want)
		}
	}
	for _, forbidden := range []string{"https://", "ssh://", ".git", "\n"} {
		if strings.Contains(message, forbidden) {
			t.Errorf("exhaustion %q contains %q", message, forbidden)
		}
	}
}

// --- bounded-resolution executor tests ---
//
// The executor is driven through its production entry point,
// AcquireNetworkResolved, against a stand-in git: a POSIX shell wrapper
// around the real git that fails a fetch with fixture stderr or rewrites
// it to a local fixture repository. Everything else — tool validation,
// private-state initialization, strict fetch argv, clean environment,
// broker materialization, raw-object proof — is the real lane.

const (
	transportHTTPS = "https://fixture.test/repository.git"
	transportSSH   = "ssh://git@fixture.test/repository.git"
)

const transportDNSStderr = "fatal: unable to access 'https://fixture.test/repository.git/': Could not resolve host: fixture.test"

// transportBehavior is what the fake transport git does when a fetch
// names url.
type transportBehavior struct {
	// succeed rewrites the fetch to fileRepo; otherwise the fetch
	// fails with stderr after sleep.
	succeed  bool
	fileRepo string
	stderr   string
	sleep    time.Duration
	// hang replaces the wrapper with the sleep itself, so the lane's
	// context kill lands on the sleeper directly. Without exec, the
	// killed shell would orphan a grandchild sleep that holds Go's
	// exec pipes open until it exits on its own.
	hang bool
	// leakSecret prints the live broker secret into the failing
	// fetch's stderr, simulating a git that echoes its environment.
	leakSecret bool
	// flood prints more than the lane's 64KiB stderr capture bound
	// before the configured stderr, overflowing the bound. The filler
	// repeats one closed-table availability line, so only the
	// truncation flag — never the retained content — may close the
	// fallback.
	flood bool
}

func failTransport(stderr string) transportBehavior { return transportBehavior{stderr: stderr} }

func fakeTransportGitTool(t *testing.T, routes map[string]transportBehavior) (GitTool, string) {
	t.Helper()
	realTool := realGitTool(t)
	root := t.TempDir()
	wrapper := filepath.Join(root, "git-wrapper")
	logPath := filepath.Join(root, "argv.log")
	urls := make([]string, 0, len(routes))
	for url := range routes {
		urls = append(urls, url)
	}
	// Sorted, so one route set always generates one script.
	for i := 1; i < len(urls); i++ {
		for j := i; j > 0 && urls[j] < urls[j-1]; j-- {
			urls[j], urls[j-1] = urls[j-1], urls[j]
		}
	}
	var failArms, rewrites strings.Builder
	for _, url := range urls {
		route := routes[url]
		if route.succeed {
			fmt.Fprintf(&rewrites, "if [ \"$arg\" = %s ]; then arg=%s; fi\n", shellQuote(url), shellQuote("file://"+route.fileRepo))
			continue
		}
		fmt.Fprintf(&failArms, "if [ \"$arg\" = %s ]; then\n", shellQuote(url))
		if route.flood {
			// Builtins only: the fetch environment carries an empty
			// PATH by design. 1700 closed-table availability lines
			// to stderr overflow the 64KiB capture bound.
			failArms.WriteString("i=0; while [ \"$i\" -lt 1700 ]; do printf 'ssh: connect to host fixture.test port 22: Connection refused\\n' >&2; i=$((i+1)); done\n")
		}
		if route.sleep > 0 {
			// Absolute path: the fetch environment carries an empty
			// PATH by design, so only builtins and absolute paths run.
			if route.hang {
				fmt.Fprintf(&failArms, "exec /bin/sleep %d\n", int(route.sleep/time.Second))
			} else {
				fmt.Fprintf(&failArms, "/bin/sleep %d\n", int(route.sleep/time.Second))
			}
		}
		if route.leakSecret {
			failArms.WriteString("printf 'leaked-secret: %s\\n' \"${" + EnvHTTPSBrokerSecret + "-}\" >&2\n")
		}
		fmt.Fprintf(&failArms, "printf '%%s' %s >&2; exit 128\nfi\n", shellQuote(route.stderr))
	}
	script := fmt.Sprintf(`#!/bin/sh
{
printf 'argv:'
for arg in "$@"; do printf ' <%%s>' "$arg"; done
printf ' | env: HOME=%%s CFGGLOBAL=%%s CFGSYSTEM=%%s GIT_SSH=%%s askpass=%%s' "${HOME-}" "${GIT_CONFIG_GLOBAL-}" "${GIT_CONFIG_SYSTEM-}" "${GIT_SSH-}" "${GIT_ASKPASS-}"
secret_present=0
state_present=0
[ -n "${%s-}" ] && secret_present=1
[ -n "${%s-}" ] && state_present=1
printf ' secret=%%s state=%%s\n' "$secret_present" "$state_present"
if [ "$state_present" = 1 ] && [ -f "${%s-}" ]; then
broker_state=""
read -r broker_state < "${%s-}"
printf 'broker-state: %%s\n' "$broker_state"
fi
} >> %s
is_fetch=0
for arg in "$@"; do
if [ "$arg" = "fetch" ]; then is_fetch=1; fi
done
if [ "$is_fetch" = 1 ]; then
for arg in "$@"; do
:
%s
done
fi
args=""
for arg in "$@"; do
%s
case "$arg" in
protocol.https.allow=always|protocol.ssh.allow=always) arg=protocol.file.allow=always ;;
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
`, EnvHTTPSBrokerSecret, EnvHTTPSBrokerState, EnvHTTPSBrokerState, EnvHTTPSBrokerState,
		shellQuote(logPath), failArms.String(), rewrites.String(), shellQuote(realTool.Executable))
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	realTool.Executable = wrapper
	realTool.AskPass = "/usr/bin/false"
	realTool.SSHWrapper = "/usr/bin/false"
	return realTool, logPath
}

// transportFetchLines returns the logged fetch invocations in order.
func transportFetchLines(t *testing.T, logPath string) []string {
	t.Helper()
	payload, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("reading git log: %v", err)
	}
	var fetches []string
	for _, line := range strings.Split(string(payload), "\n") {
		if strings.Contains(line, "<fetch>") {
			fetches = append(fetches, line)
		}
	}
	return fetches
}

func transportBase(t *testing.T, tool GitTool, fixture gitFixture) NetworkRequest {
	t.Helper()
	source, err := ParseSource(transportHTTPS)
	if err != nil {
		t.Fatal(err)
	}
	return NetworkRequest{Source: source, Lock: LockedCommit{ObjectFormat: "sha1", Hex: fixture.commit}, Tool: tool}
}

func transportSSHMaterial(t *testing.T) (identity, knownHosts string) {
	t.Helper()
	identity = filepath.Join(t.TempDir(), "id_example")
	knownHosts = filepath.Join(t.TempDir(), "known_hosts")
	if err := os.WriteFile(identity, []byte("key"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(knownHosts, []byte("hosts"), 0o600); err != nil {
		t.Fatal(err)
	}
	return identity, knownHosts
}

// transportSSHBase builds the manager-owned base of per-attempt SSH
// wrapper policies. The SSH executable never runs in fixture-rewritten
// fetches; it must only be an admitted absolute regular file.
func transportSSHBase(t *testing.T) SSHWrapperBase {
	t.Helper()
	dir := t.TempDir()
	ssh := filepath.Join(dir, "ssh")
	if err := os.WriteFile(ssh, []byte("#!/bin/sh\nexit 1\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	emptyConfig := filepath.Join(dir, "ssh.config")
	if err := os.WriteFile(emptyConfig, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	emptyKnownHosts := filepath.Join(dir, "empty_known_hosts")
	if err := os.WriteFile(emptyKnownHosts, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	return SSHWrapperBase{SSH: ssh, EmptyConfig: emptyConfig, EmptyKnownHosts: emptyKnownHosts, ConnectTimeout: 15}
}

// transportSSHBaseTool equips a fake-transport tool for SSH attempts:
// base lane credentials plus the manager wrapper-policy base.
func transportSSHBaseTool(t *testing.T, tool GitTool) GitTool {
	t.Helper()
	identity, knownHosts := transportSSHMaterial(t)
	tool.SSHCredentials = OperatorSSHCredentials{Identity: identity, KnownHosts: knownHosts}
	tool.SSHBase = transportSSHBase(t)
	return tool
}

func requirePOSIXTransport(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only; production code is platform-neutral")
	}
}

func TestResolvedTransportFallsBackOnAvailabilityThenVerifies(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", false)
	tool, logPath := fakeTransportGitTool(t, map[string]transportBehavior{
		transportHTTPS: failTransport(transportDNSStderr),
		transportSSH:   {succeed: true, fileRepo: fixture.bare},
	})
	identity, knownHosts := transportSSHMaterial(t)
	reader := &stubProviderReader{credentials: map[string]gitcred.HostCredential{
		"https-prov\x00fixture.test": {Username: "prov-user", Secret: "PROVIDER-SECRET"},
	}}
	providers := CredentialProviders{Set: ProviderSet{
		"https-prov": {HTTPS: &ProviderHTTPS{}},
		"ssh-prov":   {SSH: &ProviderSSH{Identity: identity, KnownHosts: knownHosts}},
	}, Reader: reader}
	plan := TransportPlan{
		Identity: "fixture.test/repository",
		Attempts: []TransportAttempt{
			{URL: transportHTTPS, Authentication: "https-prov"},
			{URL: transportSSH, Authentication: "ssh-prov"},
		},
		Fallback: TransportFallbackAvailabilityAuth,
	}
	var records []AttemptRecord
	base := transportBase(t, tool, fixture)
	base.Tool.SSHBase = transportSSHBase(t)
	snapshot, err := AcquireNetworkResolved(context.Background(), base, plan, providers, func(record AttemptRecord) {
		records = append(records, record)
	})
	if err != nil {
		t.Fatalf("resolved acquisition: %v", err)
	}
	if snapshot.Commit != fixture.commit || snapshot.ObjectFormat != "sha1" {
		t.Fatalf("snapshot = %+v, want sha1 %s", snapshot, fixture.commit)
	}
	want := frameSnapshot([]File{{Path: "README.md", Content: []byte("hello\x00world\n")}, {Path: "bin/tool", Content: []byte("tool\n"), Executable: true}, {Path: "empty", Content: nil}})
	if !bytes.Equal(snapshot.CanonicalBytes, want) {
		t.Fatalf("canonical bytes differ\ngot  %x\nwant %x", snapshot.CanonicalBytes, want)
	}
	fetches := transportFetchLines(t, logPath)
	if len(fetches) != 2 {
		t.Fatalf("fetch count = %d, want 2", len(fetches))
	}
	if !strings.Contains(fetches[0], transportHTTPS) || !strings.Contains(fetches[1], transportSSH) {
		t.Fatalf("fetch order wrong:\n%s", strings.Join(fetches, "\n"))
	}
	if len(records) != 2 {
		t.Fatalf("trace records = %+v, want 2", records)
	}
	if records[0].Class != FailureAvailability || !records[0].NetworkAttempted || records[0].Succeeded {
		t.Fatalf("first record = %+v, want attempted availability failure", records[0])
	}
	if records[1].Class != "" || !records[1].NetworkAttempted || !records[1].Succeeded {
		t.Fatalf("second record = %+v, want success", records[1])
	}
	payload, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(payload), `"username":"prov-user"`) {
		t.Fatalf("first fetch was not offered the provider credential:\n%s", payload)
	}
}

func TestResolvedTransportFailClosedClassesStopAfterOneFetch(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", false)
	for _, testCase := range []struct {
		name   string
		stderr string
		class  FailureClass
	}{
		{"tls", "fatal: unable to access 'https://fixture.test/repository.git/': SSL certificate problem: unable to get local issuer certificate", FailureTLS},
		{"host-key", "Host key verification failed.", FailureHostKey},
		{"http-404", "fatal: unable to access 'https://fixture.test/repository.git/': The requested URL returned error: 404", FailureHTTP404},
		{"repository-not-found", "remote: Repository not found.", FailureHTTP404},
		{"unknown", "fatal: remote error: something a future git invented", FailureUnknown},
		{"empty", "", FailureUnknown},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			tool, logPath := fakeTransportGitTool(t, map[string]transportBehavior{
				transportHTTPS: failTransport(testCase.stderr),
				transportSSH:   {succeed: true, fileRepo: fixture.bare},
			})
			plan := TransportPlan{
				Identity: "fixture.test/repository",
				Attempts: []TransportAttempt{{URL: transportHTTPS}, {URL: transportSSH}},
				Fallback: TransportFallbackAvailabilityAuth,
			}
			var records []AttemptRecord
			_, err := AcquireNetworkResolved(context.Background(), transportBase(t, tool, fixture), plan, nil, func(record AttemptRecord) {
				records = append(records, record)
			})
			if err == nil {
				t.Fatal("fail-closed failure succeeded")
			}
			// Fail-closed classes retain the lane's specific diagnostic.
			if err.Error() != CodeSourceUnavailable+": exact source fetch failed" {
				t.Fatalf("error = %q, want the lane diagnostic", err)
			}
			if fetches := transportFetchLines(t, logPath); len(fetches) != 1 {
				t.Fatalf("fetch count = %d, want 1 (no alternate traffic)", len(fetches))
			}
			if len(records) != 1 || records[0].Class != testCase.class || !records[0].NetworkAttempted {
				t.Fatalf("records = %+v, want one %s record", records, testCase.class)
			}
		})
	}
}

func TestResolvedTransportExhaustionListsSanitizedClasses(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", false)
	const secret = "SUPERSECRETPROVIDER9"
	tool, logPath := fakeTransportGitTool(t, map[string]transportBehavior{
		transportHTTPS: failTransport(transportDNSStderr),
		transportSSH:   failTransport("git@fixture.test: Permission denied (publickey)."),
	})
	reader := &stubProviderReader{credentials: map[string]gitcred.HostCredential{
		"https-prov\x00fixture.test": {Username: "prov-user", Secret: secret},
	}}
	identity, knownHosts := transportSSHMaterial(t)
	providers := CredentialProviders{Set: ProviderSet{
		"https-prov": {HTTPS: &ProviderHTTPS{}},
		"ssh-prov":   {SSH: &ProviderSSH{Identity: identity, KnownHosts: knownHosts}},
	}, Reader: reader}
	plan := TransportPlan{
		Identity: "fixture.test/repository",
		Attempts: []TransportAttempt{
			{URL: transportHTTPS, Authentication: "https-prov"},
			{URL: transportSSH, Authentication: "ssh-prov"},
		},
		Fallback: TransportFallbackAvailabilityAuth,
	}
	var records []AttemptRecord
	base := transportBase(t, tool, fixture)
	base.Tool.SSHBase = transportSSHBase(t)
	_, err := AcquireNetworkResolved(context.Background(), base, plan, providers, func(record AttemptRecord) {
		records = append(records, record)
	})
	if err == nil {
		t.Fatal("double failure succeeded")
	}
	if ErrorCode(err) != CodeRepositoryEndpointUnavailable {
		t.Fatalf("error = %v, want %s", err, CodeRepositoryEndpointUnavailable)
	}
	message := err.Error()
	for _, want := range []string{"fixture.test/repository", "availability", "auth", "provider \"https-prov\"", "provider \"ssh-prov\""} {
		if !strings.Contains(message, want) {
			t.Errorf("exhaustion %q lacks %q", message, want)
		}
	}
	for _, forbidden := range []string{secret, "Could not resolve", "Permission denied", "https://", "ssh://"} {
		if strings.Contains(message, forbidden) {
			t.Errorf("exhaustion %q contains %q", message, forbidden)
		}
	}
	if fetches := transportFetchLines(t, logPath); len(fetches) != 2 {
		t.Fatalf("fetch count = %d, want exactly 2 (one per endpoint)", len(fetches))
	}
	if len(records) != 2 || records[0].Class != FailureAvailability || records[1].Class != FailureAuth {
		t.Fatalf("records = %+v", records)
	}
}

func TestResolvedTransportLegacyShapeMatchesLane(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", false)
	legacy := TransportPlan{
		Identity: "fixture.test/repository",
		Attempts: []TransportAttempt{{URL: transportHTTPS}},
		Fallback: TransportFallbackNone,
	}
	t.Run("failure", func(t *testing.T) {
		resolvedTool, _ := fakeTransportGitTool(t, map[string]transportBehavior{transportHTTPS: failTransport(transportDNSStderr)})
		directTool, _ := fakeTransportGitTool(t, map[string]transportBehavior{transportHTTPS: failTransport(transportDNSStderr)})
		_, resolvedErr := AcquireNetworkResolved(context.Background(), transportBase(t, resolvedTool, fixture), legacy, nil, nil)
		_, directErr := AcquireNetwork(context.Background(), transportBase(t, directTool, fixture))
		if resolvedErr == nil || directErr == nil {
			t.Fatalf("resolved=%v direct=%v, want both failing", resolvedErr, directErr)
		}
		if resolvedErr.Error() != directErr.Error() {
			t.Fatalf("resolved %q != lane %q", resolvedErr, directErr)
		}
	})
	t.Run("success", func(t *testing.T) {
		resolvedTool, _ := fakeTransportGitTool(t, map[string]transportBehavior{transportHTTPS: {succeed: true, fileRepo: fixture.bare}})
		directTool, _ := fakeTransportGitTool(t, map[string]transportBehavior{transportHTTPS: {succeed: true, fileRepo: fixture.bare}})
		resolved, resolvedErr := AcquireNetworkResolved(context.Background(), transportBase(t, resolvedTool, fixture), legacy, nil, nil)
		direct, directErr := AcquireNetwork(context.Background(), transportBase(t, directTool, fixture))
		if resolvedErr != nil || directErr != nil {
			t.Fatalf("resolved=%v direct=%v", resolvedErr, directErr)
		}
		if resolved.Digest != direct.Digest || resolved.Commit != direct.Commit {
			t.Fatalf("resolved %+v != lane %+v", resolved, direct)
		}
	})
}

func TestResolvedTransportSkipsUnconfiguredProviderWithoutTraffic(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", false)
	tool, logPath := fakeTransportGitTool(t, map[string]transportBehavior{
		transportHTTPS: failTransport(transportDNSStderr),
		transportSSH:   {succeed: true, fileRepo: fixture.bare},
	})
	identity, knownHosts := transportSSHMaterial(t)
	providers := CredentialProviders{Set: ProviderSet{
		"ssh-prov": {SSH: &ProviderSSH{Identity: identity, KnownHosts: knownHosts}},
	}, Reader: &stubProviderReader{}}
	plan := TransportPlan{
		Identity: "fixture.test/repository",
		Attempts: []TransportAttempt{
			{URL: transportHTTPS, Authentication: "missing-prov"},
			{URL: transportSSH, Authentication: "ssh-prov"},
		},
		Fallback: TransportFallbackAvailabilityAuth,
	}
	var records []AttemptRecord
	base := transportBase(t, tool, fixture)
	base.Tool.SSHBase = transportSSHBase(t)
	snapshot, err := AcquireNetworkResolved(context.Background(), base, plan, providers, func(record AttemptRecord) {
		records = append(records, record)
	})
	if err != nil {
		t.Fatalf("resolved acquisition: %v", err)
	}
	if snapshot.Commit != fixture.commit {
		t.Fatalf("snapshot commit = %s, want %s", snapshot.Commit, fixture.commit)
	}
	fetches := transportFetchLines(t, logPath)
	if len(fetches) != 1 || !strings.Contains(fetches[0], transportSSH) {
		t.Fatalf("fetches = %q, want only the SSH endpoint", fetches)
	}
	if len(records) != 2 || records[0].Class != FailureAuth || records[0].NetworkAttempted || !records[1].Succeeded {
		t.Fatalf("records = %+v", records)
	}
}

func TestResolvedTransportNilProvidersSkipWithoutTraffic(t *testing.T) {
	requireResolvedLane(t)
	source, err := ParseSource(transportHTTPS)
	if err != nil {
		t.Fatal(err)
	}
	plan := TransportPlan{
		Identity: source.Identity,
		Attempts: []TransportAttempt{{URL: transportHTTPS, Authentication: "p"}},
		Fallback: TransportFallbackNone,
	}
	base := NetworkRequest{Source: source, Lock: LockedCommit{ObjectFormat: "sha1", Hex: strings.Repeat("1", 40)}}
	var records []AttemptRecord
	_, err = AcquireNetworkResolved(context.Background(), base, plan, nil, func(record AttemptRecord) {
		records = append(records, record)
	})
	if ErrorCode(err) != CodeRepositoryEndpointUnavailable {
		t.Fatalf("error = %v, want %s", err, CodeRepositoryEndpointUnavailable)
	}
	if len(records) != 1 || records[0].NetworkAttempted {
		t.Fatalf("records = %+v, want one traffic-free record", records)
	}
}

func TestResolvedTransportPinnedAuthStopsWithoutSecondAttempt(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", false)
	tool, logPath := fakeTransportGitTool(t, map[string]transportBehavior{
		transportHTTPS: failTransport("fatal: unable to access 'https://fixture.test/repository.git/': The requested URL returned error: 403"),
	})
	providers := CredentialProviders{Set: ProviderSet{
		"https-prov": {HTTPS: &ProviderHTTPS{}},
	}, Reader: &stubProviderReader{credentials: map[string]gitcred.HostCredential{
		"https-prov\x00fixture.test": {Username: "u", Secret: "s"},
	}}}
	// pinned-auth: one policy attempt with fallback none. The auth
	// rejection is fallback-eligible but no second endpoint exists.
	plan := TransportPlan{
		Identity: "fixture.test/repository",
		Attempts: []TransportAttempt{{URL: transportHTTPS, Authentication: "https-prov"}},
		Fallback: TransportFallbackNone,
	}
	_, err := AcquireNetworkResolved(context.Background(), transportBase(t, tool, fixture), plan, providers, nil)
	if ErrorCode(err) != CodeRepositoryEndpointUnavailable {
		t.Fatalf("error = %v, want %s", err, CodeRepositoryEndpointUnavailable)
	}
	if fetches := transportFetchLines(t, logPath); len(fetches) != 1 {
		t.Fatalf("fetch count = %d, want 1", len(fetches))
	}
}

func TestResolvedTransportSecondEndpointMustProveLockedContent(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", true)
	moved := makeGitFixture(t, "sha1", true)
	git := realGitPath(t)
	runTestGit(t, moved.work, git, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.test", "commit", "--quiet", "--allow-empty", "-m", "second")
	runTestGit(t, moved.work, git, "tag", "-d", "v1.4.0")
	runTestGit(t, moved.work, git, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.test", "tag", "-a", "v1.4.0", "-m", "moved")
	_ = os.RemoveAll(moved.bare)
	runTestGit(t, "", git, "clone", "--quiet", "--bare", "--", moved.work, moved.bare)

	tool, logPath := fakeTransportGitTool(t, map[string]transportBehavior{
		transportHTTPS: failTransport(transportDNSStderr),
		transportSSH:   {succeed: true, fileRepo: moved.bare},
	})
	plan := TransportPlan{
		Identity: "fixture.test/repository",
		Attempts: []TransportAttempt{{URL: transportHTTPS}, {URL: transportSSH}},
		Fallback: TransportFallbackAvailabilityAuth,
	}
	base := transportBase(t, tool, fixture)
	base.Tag = "v1.4.0"
	base.Tool = transportSSHBaseTool(t, base.Tool)
	var records []AttemptRecord
	_, err := AcquireNetworkResolved(context.Background(), base, plan, nil, func(record AttemptRecord) {
		records = append(records, record)
	})
	// The alternate served bytes, but not the locked content: the exact
	// tag disagrees with the lock, which fails closed.
	if ErrorCode(err) != CodeRefMoved {
		t.Fatalf("error = %v, want %s", err, CodeRefMoved)
	}
	if len(records) != 2 || records[0].Class != FailureAvailability || records[1].Class != FailureRef {
		t.Fatalf("records = %+v", records)
	}
	_ = logPath
}

func TestResolvedTransportAnonymousProviderOffersNothing(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", false)
	tool, logPath := fakeTransportGitTool(t, map[string]transportBehavior{
		transportHTTPS: {succeed: true, fileRepo: fixture.bare},
	})
	base := transportBase(t, tool, fixture)
	base.Tool.HTTPSCredentials = NewHTTPSCredentials("fixture.test", "decoy-user", "DECOY-SECRET")
	providers := CredentialProviders{Set: ProviderSet{"anon": {HTTPS: &ProviderHTTPS{Anonymous: true}}}}
	plan := TransportPlan{
		Identity: "fixture.test/repository",
		Attempts: []TransportAttempt{{URL: transportHTTPS, Authentication: "anon"}},
		Fallback: TransportFallbackNone,
	}
	snapshot, err := AcquireNetworkResolved(context.Background(), base, plan, providers, nil)
	if err != nil {
		t.Fatalf("anonymous acquisition: %v", err)
	}
	if snapshot.Commit != fixture.commit {
		t.Fatalf("snapshot commit = %s", snapshot.Commit)
	}
	fetches := transportFetchLines(t, logPath)
	if len(fetches) != 1 {
		t.Fatalf("fetch count = %d", len(fetches))
	}
	if !strings.Contains(fetches[0], "secret=0 state=0") {
		t.Fatalf("anonymous fetch offered credentials: %s", fetches[0])
	}
	payload, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), "broker-state:") {
		t.Fatalf("anonymous fetch materialized a broker:\n%s", payload)
	}
}

func TestResolvedTransportRejectsPlanBeforeAnyGitCall(t *testing.T) {
	source, err := ParseSource(transportHTTPS)
	if err != nil {
		t.Fatal(err)
	}
	base := NetworkRequest{
		Source: source,
		Lock:   LockedCommit{ObjectFormat: "sha1", Hex: strings.Repeat("1", 40)},
		// A tool that cannot run: any process spawn fails loudly,
		// proving plan validation precedes every Git call.
		Tool: GitTool{Executable: filepath.Join(t.TempDir(), "no-such-git")},
	}
	invalid := []TransportPlan{
		{Identity: source.Identity, Fallback: TransportFallbackNone},
		{Identity: source.Identity, Attempts: []TransportAttempt{{URL: transportHTTPS}, {URL: transportHTTPS}}, Fallback: TransportFallbackAvailabilityAuth},
		{Identity: "other.test/x", Attempts: []TransportAttempt{{URL: transportHTTPS}}, Fallback: TransportFallbackNone},
		{Identity: source.Identity, Attempts: []TransportAttempt{{URL: "file:///tmp/x.git"}}, Fallback: TransportFallbackNone},
		{Identity: source.Identity, Attempts: []TransportAttempt{{URL: transportHTTPS, Authentication: "/bin/sh"}}, Fallback: TransportFallbackNone},
	}
	for index, plan := range invalid {
		if _, err := AcquireNetworkResolved(context.Background(), base, plan, nil, nil); ErrorCode(err) != CodeRepositoryPolicyInvalid {
			t.Errorf("plan %d: error = %v, want %s", index, err, CodeRepositoryPolicyInvalid)
		}
	}
	// A cross-identity plan (valid shape, wrong acquisition) fails the
	// same way, before any Git call.
	cross := TransportPlan{
		Identity: "fixture.test/other",
		Attempts: []TransportAttempt{{URL: "https://fixture.test/other.git"}},
		Fallback: TransportFallbackNone,
	}
	if _, err := AcquireNetworkResolved(context.Background(), base, cross, nil, nil); ErrorCode(err) != CodeRepositoryPolicyInvalid {
		t.Fatalf("cross-identity error = %v, want %s", err, CodeRepositoryPolicyInvalid)
	}
}

func TestResolvedTransportTotalDeadlineBoundsSlowFetch(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", false)
	tool, logPath := fakeTransportGitTool(t, map[string]transportBehavior{
		transportHTTPS: {stderr: transportDNSStderr, sleep: 3 * time.Second},
		transportSSH:   {stderr: transportDNSStderr, sleep: 30 * time.Second, hang: true},
	})
	plan := TransportPlan{
		Identity: "fixture.test/repository",
		Attempts: []TransportAttempt{{URL: transportHTTPS}, {URL: transportSSH}},
		Fallback: TransportFallbackAvailabilityAuth,
	}
	base := transportBase(t, tool, fixture)
	base.Tool = transportSSHBaseTool(t, base.Tool)
	base.Limits = Limits{Timeout: 6 * time.Second}
	var records []AttemptRecord
	start := time.Now()
	_, err := AcquireNetworkResolved(context.Background(), base, plan, nil, func(record AttemptRecord) {
		records = append(records, record)
	})
	elapsed := time.Since(start)
	// A per-attempt clock would spend 3s + 6s = 9s minimum (both sleeps
	// are lower bounds); the shared clock bounds the total near 6s.
	if elapsed > 7500*time.Millisecond {
		t.Fatalf("elapsed %v exceeds the shared total deadline", elapsed)
	}
	if err == nil || ErrorCode(err) != CodeSourceUnavailable {
		t.Fatalf("error = %v, want the fail-closed lane diagnostic", err)
	}
	if fetches := transportFetchLines(t, logPath); len(fetches) != 2 {
		t.Fatalf("fetch count = %d, want 2 (the second starts, then the shared clock stops it)", len(fetches))
	}
	if len(records) != 2 || records[0].Class != FailureAvailability || records[1].Class != FailureUnknown {
		t.Fatalf("records = %+v, want availability then deadline-unknown", records)
	}
}

func TestResolvedTransportLeaksNothingIntoErrors(t *testing.T) {
	requirePOSIXTransport(t)
	const secret = "SUPERSECRETPROVIDER9"
	t.Run("exhaustion", func(t *testing.T) {
		fixture := makeGitFixture(t, "sha1", false)
		tool, _ := fakeTransportGitTool(t, map[string]transportBehavior{
			transportHTTPS: failTransport(transportDNSStderr),
			transportSSH:   failTransport("git@fixture.test: Permission denied (publickey)."),
		})
		identity, knownHosts := transportSSHMaterial(t)
		reader := &stubProviderReader{credentials: map[string]gitcred.HostCredential{
			"https-prov\x00fixture.test": {Username: "u", Secret: secret},
		}}
		providers := CredentialProviders{Set: ProviderSet{
			"https-prov": {HTTPS: &ProviderHTTPS{}},
			"ssh-prov":   {SSH: &ProviderSSH{Identity: identity, KnownHosts: knownHosts}},
		}, Reader: reader}
		plan := TransportPlan{
			Identity: "fixture.test/repository",
			Attempts: []TransportAttempt{
				{URL: transportHTTPS, Authentication: "https-prov"},
				{URL: transportSSH, Authentication: "ssh-prov"},
			},
			Fallback: TransportFallbackAvailabilityAuth,
		}
		var records []AttemptRecord
		base := transportBase(t, tool, fixture)
		base.Tool.SSHBase = transportSSHBase(t)
		_, err := AcquireNetworkResolved(context.Background(), base, plan, providers, func(record AttemptRecord) {
			records = append(records, record)
		})
		// The live broker secret authenticates both attempts through
		// the reader, and the closed-vocabulary exhaustion plus the
		// trace records must not contain it. A fetch that echoes the
		// secret into stderr poisons the output to unknown under the
		// closed table (see the fail-closed subtest): no echo can
		// reach an exhaustion message.
		if ErrorCode(err) != CodeRepositoryEndpointUnavailable {
			t.Fatalf("error = %v, want %s", err, CodeRepositoryEndpointUnavailable)
		}
		if strings.Contains(err.Error(), secret) {
			t.Fatalf("exhaustion leaked the broker secret: %v", err)
		}
		if dump := fmt.Sprintf("%+v", records); strings.Contains(dump, secret) {
			t.Fatalf("trace leaked the broker secret: %s", dump)
		}
	})
	t.Run("fail-closed", func(t *testing.T) {
		fixture := makeGitFixture(t, "sha1", false)
		tool, _ := fakeTransportGitTool(t, map[string]transportBehavior{
			transportHTTPS: {stderr: "fatal: SSL certificate problem: unable to get local issuer certificate", leakSecret: true},
		})
		providers := CredentialProviders{Set: ProviderSet{
			"https-prov": {HTTPS: &ProviderHTTPS{}},
		}, Reader: &stubProviderReader{credentials: map[string]gitcred.HostCredential{
			"https-prov\x00fixture.test": {Username: "u", Secret: secret},
		}}}
		plan := TransportPlan{
			Identity: "fixture.test/repository",
			Attempts: []TransportAttempt{{URL: transportHTTPS, Authentication: "https-prov"}},
			Fallback: TransportFallbackAvailabilityAuth,
		}
		_, err := AcquireNetworkResolved(context.Background(), transportBase(t, tool, fixture), plan, providers, nil)
		if err == nil || strings.Contains(err.Error(), secret) {
			t.Fatalf("fail-closed error leaked the broker secret: %v", err)
		}
	})
}

func TestResolvedTransportKeepsStrictLanePerAttempt(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", false)
	tool, logPath := fakeTransportGitTool(t, map[string]transportBehavior{
		transportHTTPS: failTransport(transportDNSStderr),
		transportSSH:   {succeed: true, fileRepo: fixture.bare},
	})
	plan := TransportPlan{
		Identity: "fixture.test/repository",
		Attempts: []TransportAttempt{{URL: transportHTTPS}, {URL: transportSSH}},
		Fallback: TransportFallbackAvailabilityAuth,
	}
	base := transportBase(t, tool, fixture)
	base.Tool = transportSSHBaseTool(t, base.Tool)
	if _, err := AcquireNetworkResolved(context.Background(), base, plan, nil, nil); err != nil {
		t.Fatalf("resolved acquisition: %v", err)
	}
	fetches := transportFetchLines(t, logPath)
	if len(fetches) != 2 {
		t.Fatalf("fetch count = %d", len(fetches))
	}
	for index, fetch := range fetches {
		for _, marker := range []string{
			"--git-dir=", "protocol.allow=never", "http.followRedirects=false",
			"http.sslVerify=true", "credential.helper=", "core.askPass=", "core.hooksPath=",
		} {
			if !strings.Contains(fetch, marker) {
				t.Errorf("fetch %d lacks %q:\n%s", index+1, marker, fetch)
			}
		}
		if !strings.Contains(fetch, "HOME=") || !strings.Contains(fetch, "curator-buildrepo-") {
			t.Errorf("fetch %d ran outside a private HOME:\n%s", index+1, fetch)
		}
		if !strings.Contains(fetch, "CFGGLOBAL=") || !strings.Contains(fetch, "global.gitconfig") {
			t.Errorf("fetch %d ran without a pinned global config:\n%s", index+1, fetch)
		}
	}
	if !strings.Contains(fetches[0], "protocol.https.allow=always") || !strings.Contains(fetches[1], "protocol.ssh.allow=always") {
		t.Fatalf("per-attempt protocol allowlist wrong:\n%s", strings.Join(fetches, "\n"))
	}
	// The SSH attempt runs behind its own bound wrapper copy, not the
	// tool's static wrapper; the HTTPS attempt carries no wrapper.
	if !strings.Contains(fetches[1], "GIT_SSH=") || !strings.Contains(fetches[1], SSHWrapperName) {
		t.Fatalf("SSH attempt did not use its bound per-attempt wrapper:\n%s", fetches[1])
	}
	if strings.Contains(fetches[0], SSHWrapperName) {
		t.Fatalf("HTTPS attempt carried an SSH wrapper:\n%s", fetches[0])
	}
	payload, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), "insteadOf") {
		t.Fatalf("ambient remapping reached a fetch:\n%s", payload)
	}
}

func TestAcquireNetworkFetchFailureKeepsLaneDiagnostic(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", false)
	for _, stderr := range []string{
		transportDNSStderr,
		"fatal: unable to access 'https://fixture.test/repository.git/': SSL certificate problem: unable to get local issuer certificate",
		"Host key verification failed.",
	} {
		tool, _ := fakeTransportGitTool(t, map[string]transportBehavior{transportHTTPS: failTransport(stderr)})
		_, err := AcquireNetwork(context.Background(), transportBase(t, tool, fixture))
		if err == nil {
			t.Fatal("failing fetch succeeded")
		}
		// The fetchError refactor must not change the legacy lane:
		// fetch stderr never surfaces, whatever it says.
		if err.Error() != CodeSourceUnavailable+": exact source fetch failed" {
			t.Fatalf("error = %q", err)
		}
	}
}

func TestReviewResolvedMustRetainAdmission(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", false)
	for _, mode := range []string{"missing-ssh-wrapper", "missing-ssh-credentials", "missing-https-broker", "invalid-ref-kind"} {
		t.Run(mode, func(t *testing.T) {
			tool, logPath := fakeTransportGitTool(t, map[string]transportBehavior{transportHTTPS: {succeed: true, fileRepo: fixture.bare}, transportSSH: {succeed: true, fileRepo: fixture.bare}})
			base := transportBase(t, tool, fixture)
			url := transportHTTPS
			switch mode {
			case "missing-ssh-wrapper":
				url = transportSSH
				base.Tool.SSHWrapper = ""
			case "missing-ssh-credentials":
				url = transportSSH
				base.Tool.SSHCredentials = OperatorSSHCredentials{}
			case "missing-https-broker":
				base.Tool.AskPass = ""
			case "invalid-ref-kind":
				base.RefKind = "not-supported"
				base.RefValue = fixture.commit
			}
			base.Source, _ = ParseSource(url)
			_, direct := AcquireNetwork(context.Background(), base)
			if direct == nil {
				t.Fatal("control did not refuse")
			}
			plan := TransportPlan{Identity: base.Source.Identity, Attempts: []TransportAttempt{{URL: url}}, Fallback: TransportFallbackNone}
			_, got := AcquireNetworkResolved(context.Background(), base, plan, nil, nil)
			if got == nil {
				t.Fatalf("resolved accepted %s; legacy refused %v; fetches=%d", mode, direct, len(transportFetchLines(t, logPath)))
			}
		})
	}
}

func TestReviewAmbiguousFailureMustNotFallback(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", false)
	for _, msg := range []string{"fatal: cannot create temporary file: Permission denied", "remote: audit denied: operation timed out", "error: object hash mismatch\nfatal: connection timed out"} {
		t.Run(msg, func(t *testing.T) {
			tool, logPath := fakeTransportGitTool(t, map[string]transportBehavior{transportHTTPS: failTransport(msg), transportSSH: {succeed: true, fileRepo: fixture.bare}})
			plan := TransportPlan{Identity: "fixture.test/repository", Attempts: []TransportAttempt{{URL: transportHTTPS}, {URL: transportSSH}}, Fallback: TransportFallbackAvailabilityAuth}
			_, err := AcquireNetworkResolved(context.Background(), transportBase(t, tool, fixture), plan, nil, nil)
			if n := len(transportFetchLines(t, logPath)); n != 1 {
				t.Fatalf("ambiguous/local/integrity error caused %d fetches (err=%v)", n, err)
			}
		})
	}
}

func TestReviewDeadlineIncludesChildPipes(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", false)
	tool, _ := fakeTransportGitTool(t, map[string]transportBehavior{transportHTTPS: {stderr: transportDNSStderr, sleep: 5 * time.Second}})
	base := transportBase(t, tool, fixture)
	base.Limits.Timeout = 2 * time.Second
	plan := TransportPlan{Identity: base.Source.Identity, Attempts: []TransportAttempt{{URL: transportHTTPS}}, Fallback: TransportFallbackNone}
	start := time.Now()
	_, err := AcquireNetworkResolved(context.Background(), base, plan, nil, nil)
	if elapsed := time.Since(start); elapsed > 4*time.Second {
		t.Fatalf("2s total deadline returned after %s: %v", elapsed, err)
	}
}

func TestResolvedTransportAmbiguousFailureMustNotFallbackWhenAlternateReady(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", false)
	for _, testCase := range []struct {
		name   string
		stderr string
		class  FailureClass
	}{
		{"local-permission-denied", "fatal: cannot create temporary file: Permission denied", FailureUnknown},
		{"audit-timeout", "remote: audit denied: operation timed out", FailureAudit},
		{"integrity-timeout", "error: object hash mismatch\nfatal: connection timed out", FailureIntegrity},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			tool, logPath := fakeTransportGitTool(t, map[string]transportBehavior{
				transportHTTPS: failTransport(testCase.stderr),
				transportSSH:   {succeed: true, fileRepo: fixture.bare},
			})
			// The alternate is fully fetchable: any fallback would
			// reach it and succeed, so one fetch proves the
			// classifier itself closed the plan.
			tool = transportSSHBaseTool(t, tool)
			plan := TransportPlan{
				Identity: "fixture.test/repository",
				Attempts: []TransportAttempt{{URL: transportHTTPS}, {URL: transportSSH}},
				Fallback: TransportFallbackAvailabilityAuth,
			}
			var records []AttemptRecord
			_, err := AcquireNetworkResolved(context.Background(), transportBase(t, tool, fixture), plan, nil, func(record AttemptRecord) {
				records = append(records, record)
			})
			if err == nil {
				t.Fatal("ambiguous failure fell back to a ready alternate and succeeded")
			}
			if err.Error() != CodeSourceUnavailable+": exact source fetch failed" {
				t.Fatalf("error = %v, want the fail-closed lane diagnostic", err)
			}
			if fetches := transportFetchLines(t, logPath); len(fetches) != 1 {
				t.Fatalf("fetch count = %d, want 1", len(fetches))
			}
			if len(records) != 1 || records[0].Class != testCase.class || !records[0].NetworkAttempted {
				t.Fatalf("records = %+v, want one attempted %s record", records, testCase.class)
			}
		})
	}
}

func TestResolvedTransportFallbackIntoAdmissionRefusalExhausts(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", false)
	tool, logPath := fakeTransportGitTool(t, map[string]transportBehavior{
		transportHTTPS: failTransport(transportDNSStderr),
		transportSSH:   {succeed: true, fileRepo: fixture.bare},
	})
	plan := TransportPlan{
		Identity: "fixture.test/repository",
		Attempts: []TransportAttempt{{URL: transportHTTPS}, {URL: transportSSH}},
		Fallback: TransportFallbackAvailabilityAuth,
	}
	// The base tool carries no SSH credentials and no manager base: the
	// SSH attempt refuses in admission with no traffic, and the policy
	// plan exhausts on auth without waiting out the deadline.
	start := time.Now()
	var records []AttemptRecord
	_, err := AcquireNetworkResolved(context.Background(), transportBase(t, tool, fixture), plan, nil, func(record AttemptRecord) {
		records = append(records, record)
	})
	elapsed := time.Since(start)
	if ErrorCode(err) != CodeRepositoryEndpointUnavailable {
		t.Fatalf("error = %v, want %s", err, CodeRepositoryEndpointUnavailable)
	}
	if fetches := transportFetchLines(t, logPath); len(fetches) != 1 {
		t.Fatalf("fetch count = %d, want 1", len(fetches))
	}
	if len(records) != 2 || records[0].Class != FailureAvailability || !records[0].NetworkAttempted ||
		records[1].Class != FailureAuth || records[1].NetworkAttempted {
		t.Fatalf("records = %+v, want attempted availability then traffic-free auth", records)
	}
	if elapsed > 30*time.Second {
		t.Fatalf("elapsed %v: a refused attempt must not wait out the deadline", elapsed)
	}
}

func TestResolvedTransportTruncatedStderrFailsClosed(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", false)
	tool, logPath := fakeTransportGitTool(t, map[string]transportBehavior{
		transportHTTPS: {stderr: "fatal: trailing tail", flood: true},
		transportSSH:   {succeed: true, fileRepo: fixture.bare},
	})
	// The alternate is fully fetchable: any fallback would reach it and
	// succeed, so one fetch proves truncation itself closed the plan.
	tool = transportSSHBaseTool(t, tool)
	plan := TransportPlan{
		Identity: "fixture.test/repository",
		Attempts: []TransportAttempt{{URL: transportHTTPS}, {URL: transportSSH}},
		Fallback: TransportFallbackAvailabilityAuth,
	}
	var records []AttemptRecord
	_, err := AcquireNetworkResolved(context.Background(), transportBase(t, tool, fixture), plan, nil, func(record AttemptRecord) {
		records = append(records, record)
	})
	if err == nil {
		t.Fatal("truncated evidence fell back to a ready alternate and succeeded")
	}
	if err.Error() != CodeSourceUnavailable+": exact source fetch failed" {
		t.Fatalf("error = %v, want the fail-closed lane diagnostic", err)
	}
	if fetches := transportFetchLines(t, logPath); len(fetches) != 1 {
		t.Fatalf("fetch count = %d, want 1 (truncated evidence must not authorize fallback)", len(fetches))
	}
	if len(records) != 1 || records[0].Class != FailureUnknown || !records[0].NetworkAttempted {
		t.Fatalf("records = %+v, want one attempted unknown record", records)
	}
	// The retained filler alone classifies availability: only the
	// truncation flag closed the fallback, never the content.
	if got := ClassifyFetchOutput("ssh: connect to host fixture.test port 22: Connection refused"); got != FailureAvailability {
		t.Fatalf("flood filler classifies %q, want availability", got)
	}
}

func TestResolvedTransportSSHWithoutKnownHostsRefusesBeforeTraffic(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", false)
	identity := filepath.Join(t.TempDir(), "id_example")
	if err := os.WriteFile(identity, []byte("key"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Run("provider", func(t *testing.T) {
		tool, logPath := fakeTransportGitTool(t, map[string]transportBehavior{
			transportSSH: {succeed: true, fileRepo: fixture.bare},
		})
		providers := CredentialProviders{Set: ProviderSet{
			"ssh-nohosts": {SSH: &ProviderSSH{Identity: identity}},
		}}
		base := transportBase(t, tool, fixture)
		base.Tool.SSHBase = transportSSHBase(t)
		base.Source, _ = ParseSource(transportSSH)
		plan := TransportPlan{
			Identity: base.Source.Identity,
			Attempts: []TransportAttempt{{URL: transportSSH, Authentication: "ssh-nohosts"}},
			Fallback: TransportFallbackNone,
		}
		var records []AttemptRecord
		_, err := AcquireNetworkResolved(context.Background(), base, plan, providers, func(record AttemptRecord) {
			records = append(records, record)
		})
		// An identity without pinned host keys cannot form a wrapper
		// policy: the attempt records an unavailable auth method with
		// no traffic and the policy plan exhausts.
		if ErrorCode(err) != CodeRepositoryEndpointUnavailable {
			t.Fatalf("error = %v, want %s", err, CodeRepositoryEndpointUnavailable)
		}
		if !strings.Contains(err.Error(), "no fetch attempted") {
			t.Fatalf("exhaustion names traffic it never sent: %v", err)
		}
		if fetches := transportFetchLines(t, logPath); len(fetches) != 0 {
			t.Fatalf("fetch count = %d, want 0", len(fetches))
		}
		if len(records) != 1 || records[0].Class != FailureAuth || records[0].NetworkAttempted {
			t.Fatalf("records = %+v, want one traffic-free auth record", records)
		}
	})
	t.Run("base-credentials", func(t *testing.T) {
		tool, logPath := fakeTransportGitTool(t, map[string]transportBehavior{
			transportSSH: {succeed: true, fileRepo: fixture.bare},
		})
		base := transportBase(t, tool, fixture)
		base.Tool.SSHBase = transportSSHBase(t)
		base.Tool.SSHCredentials = OperatorSSHCredentials{Identity: identity}
		base.Source, _ = ParseSource(transportSSH)
		plan := TransportPlan{
			Identity: base.Source.Identity,
			Attempts: []TransportAttempt{{URL: transportSSH}},
			Fallback: TransportFallbackNone,
		}
		var records []AttemptRecord
		_, err := AcquireNetworkResolved(context.Background(), base, plan, nil, func(record AttemptRecord) {
			records = append(records, record)
		})
		// A legacy-shape plan keeps the lane's own diagnostic: the
		// SSH refusal surfaces unchanged, with no fetch traffic.
		if ErrorCode(err) != CodeSSHCredentialMissing {
			t.Fatalf("error = %v, want %s", err, CodeSSHCredentialMissing)
		}
		if fetches := transportFetchLines(t, logPath); len(fetches) != 0 {
			t.Fatalf("fetch count = %d, want 0", len(fetches))
		}
		if len(records) != 1 || records[0].Class != FailureAuth || records[0].NetworkAttempted {
			t.Fatalf("records = %+v, want one traffic-free auth record", records)
		}
	})
}

func TestResolvedTransportCancelledContextFetchesNothing(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", false)
	tool, logPath := fakeTransportGitTool(t, map[string]transportBehavior{
		transportHTTPS: {succeed: true, fileRepo: fixture.bare},
	})
	plan := TransportPlan{
		Identity: "fixture.test/repository",
		Attempts: []TransportAttempt{{URL: transportHTTPS}},
		Fallback: TransportFallbackNone,
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := AcquireNetworkResolved(ctx, transportBase(t, tool, fixture), plan, nil, nil); err == nil {
		t.Fatal("cancelled acquisition succeeded")
	}
	// A missing log means the wrapper never ran at all: either way, no
	// fetch went out.
	if _, err := os.Stat(logPath); err == nil {
		if fetches := transportFetchLines(t, logPath); len(fetches) != 0 {
			t.Fatalf("fetch count = %d, want 0", len(fetches))
		}
	} else if !os.IsNotExist(err) {
		t.Fatal(err)
	}
}

// fakeSSHScript writes a logging stand-in for the SSH executable: every
// invocation appends its argv to a sibling log file, then behaves per the
// attempt it serves. The first attempt's identity path carries "first" and
// fails refused (availability); the second attempt's carries "second" and
// fails host-key verification (fail closed). It returns the script path
// and the log path.
func fakeSSHScript(t *testing.T) (script, logPath string) {
	t.Helper()
	dir := t.TempDir()
	logPath = filepath.Join(dir, "ssh-argv.log")
	script = filepath.Join(dir, "fake-ssh")
	content := fmt.Sprintf(`#!/bin/sh
{
printf 'ssh-argv:'
for arg in "$@"; do printf ' <%%s>' "$arg"; done
printf '\n'
} >> %s
case "$*" in
*first*) printf '%%s\n' 'ssh: connect to host git@fixture.test port 22: Connection refused' >&2; exit 128;;
*) printf '%%s\n' 'Host key verification failed.' >&2; exit 128;;
esac
`, shellQuote(logPath))
	if err := os.WriteFile(script, []byte(content), 0o700); err != nil {
		t.Fatal(err)
	}
	return script, logPath
}

func TestResolvedTransportSSHAttemptBindsPerAttemptCredentials(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", false)
	first := t.TempDir()
	second := t.TempDir()
	writeSSHFile := func(path, content string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	identityFirst := filepath.Join(first, "first", "id")
	knownHostsFirst := filepath.Join(first, "first", "known_hosts")
	identitySecond := filepath.Join(second, "second", "id")
	knownHostsSecond := filepath.Join(second, "second", "known_hosts")
	for _, path := range []string{identityFirst, knownHostsFirst, identitySecond, knownHostsSecond} {
		writeSSHFile(path, "material\n")
	}
	// The wrapper policy carries resolved paths; resolve the markers
	// the same way so the argv assertions compare equal spellings.
	resolve := func(path string) string {
		t.Helper()
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			t.Fatal(err)
		}
		return resolved
	}
	identityFirst, knownHostsFirst = resolve(identityFirst), resolve(knownHostsFirst)
	identitySecond, knownHostsSecond = resolve(identitySecond), resolve(knownHostsSecond)
	fakeSSH, sshLog := fakeSSHScript(t)
	manager := t.TempDir()
	emptyConfig := filepath.Join(manager, "ssh.config")
	emptyKnownHosts := filepath.Join(manager, "empty_known_hosts")
	for _, path := range []string{emptyConfig, emptyKnownHosts} {
		writeSSHFile(path, "")
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	tool := realGitTool(t)
	tool.SSHWrapper = executable
	tool.SSHBase = SSHWrapperBase{SSH: fakeSSH, EmptyConfig: emptyConfig, EmptyKnownHosts: emptyKnownHosts, ConnectTimeout: 15}
	providers := CredentialProviders{Set: ProviderSet{
		"ssh-first":  {SSH: &ProviderSSH{Identity: identityFirst, KnownHosts: knownHostsFirst}},
		"ssh-second": {SSH: &ProviderSSH{Identity: identitySecond, KnownHosts: knownHostsSecond}},
	}}
	plan := TransportPlan{
		Identity: "fixture.test/repository",
		Attempts: []TransportAttempt{
			{URL: "ssh://git@fixture.test/repository.git", Authentication: "ssh-first"},
			{URL: "git@fixture.test:repository.git", Authentication: "ssh-second"},
		},
		Fallback: TransportFallbackAvailabilityAuth,
	}
	var records []AttemptRecord
	_, err = AcquireNetworkResolved(context.Background(), transportBase(t, tool, fixture), plan, providers, func(record AttemptRecord) {
		records = append(records, record)
	})
	// The first SSH attempt fails refused and the plan advances; the
	// second fails host-key verification and closes with the lane's own
	// diagnostic. Real git drove a real per-attempt wrapper on both.
	if err == nil || err.Error() != CodeSourceUnavailable+": exact source fetch failed" {
		t.Fatalf("error = %v, want the fail-closed lane diagnostic", err)
	}
	if len(records) != 2 || records[0].Class != FailureAvailability || !records[0].NetworkAttempted ||
		records[1].Class != FailureHostKey || !records[1].NetworkAttempted {
		t.Fatalf("records = %+v, want attempted availability then host-key", records)
	}
	payload, err := os.ReadFile(sshLog)
	if err != nil {
		t.Fatal(err)
	}
	invocations := strings.Split(strings.TrimSpace(string(payload)), "\n")
	if len(invocations) != 2 {
		t.Fatalf("ssh invocations = %d, want 2:\n%s", len(invocations), payload)
	}
	// Each attempt offered exactly its provider's material and endpoint:
	// the pinned identity and host keys, the fixed argv, and the exact
	// host/path tuple git hands to SSH — slash-prefixed for ssh://, bare
	// for the scp-like form.
	for _, marker := range []string{"<" + identityFirst + ">", "<UserKnownHostsFile=" + knownHostsFirst + ">",
		"<git@fixture.test>", "<git-upload-pack '/repository.git'>", "<BatchMode=yes>", "<StrictHostKeyChecking=yes>"} {
		if !strings.Contains(invocations[0], marker) {
			t.Errorf("first ssh invocation lacks %s:\n%s", marker, invocations[0])
		}
	}
	for _, marker := range []string{"<" + identitySecond + ">", "<UserKnownHostsFile=" + knownHostsSecond + ">",
		"<git@fixture.test>", "<git-upload-pack 'repository.git'>", "<BatchMode=yes>", "<StrictHostKeyChecking=yes>"} {
		if !strings.Contains(invocations[1], marker) {
			t.Errorf("second ssh invocation lacks %s:\n%s", marker, invocations[1])
		}
	}
	if strings.Contains(invocations[0], identitySecond) || strings.Contains(invocations[0], knownHostsSecond) {
		t.Errorf("first attempt was offered the second attempt's material:\n%s", invocations[0])
	}
	if strings.Contains(invocations[1], identityFirst) || strings.Contains(invocations[1], knownHostsFirst) {
		t.Errorf("second attempt was offered the first attempt's material:\n%s", invocations[1])
	}
}

func TestReviewRev2MixedUnknownMustNotFallback(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", false)
	for _, msg := range []string{
		"fatal: unable to access 'https://fixture.test/repository.git/': Could not resolve host: fixture.test\nfatal: unexpected protocol response",
		"fatal: unable to read object 0123456789012345678901234567890123456789\nfatal: connection timed out",
		"fatal: cannot create temporary file 'connection timed out': Permission denied",
	} {
		t.Run(msg, func(t *testing.T) {
			tool, logPath := fakeTransportGitTool(t, map[string]transportBehavior{transportHTTPS: failTransport(msg), transportSSH: {succeed: true, fileRepo: fixture.bare}})
			tool = transportSSHBaseTool(t, tool)
			base := transportBase(t, tool, fixture)
			plan := TransportPlan{Identity: base.Source.Identity, Attempts: []TransportAttempt{{URL: transportHTTPS}, {URL: transportSSH}}, Fallback: TransportFallbackAvailabilityAuth}
			_, err := AcquireNetworkResolved(context.Background(), base, plan, nil, nil)
			n := len(transportFetchLines(t, logPath))
			if err == nil || n != 1 {
				t.Fatalf("fail-closed diagnostic produced %d fetches, err=%v", n, err)
			}
		})
	}
}

// TestResolvedTransportUnknownTailMustNotFallbackWhenAlternateReady
// proves the poison rule at the production entry: a positively
// identified DNS diagnostic followed by an unparsed forge progress line
// fails closed with the lane's own diagnostic and no alternate traffic,
// even though the alternate is fully fetchable.
func TestResolvedTransportUnknownTailMustNotFallbackWhenAlternateReady(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", false)
	tool, logPath := fakeTransportGitTool(t, map[string]transportBehavior{
		transportHTTPS: failTransport(transportDNSStderr + "\nremote: Enumerating objects: 3, done."),
		transportSSH:   {succeed: true, fileRepo: fixture.bare},
	})
	tool = transportSSHBaseTool(t, tool)
	plan := TransportPlan{
		Identity: "fixture.test/repository",
		Attempts: []TransportAttempt{{URL: transportHTTPS}, {URL: transportSSH}},
		Fallback: TransportFallbackAvailabilityAuth,
	}
	var records []AttemptRecord
	_, err := AcquireNetworkResolved(context.Background(), transportBase(t, tool, fixture), plan, nil, func(record AttemptRecord) {
		records = append(records, record)
	})
	if err == nil {
		t.Fatal("unknown-tailed failure fell back to a ready alternate and succeeded")
	}
	if err.Error() != CodeSourceUnavailable+": exact source fetch failed" {
		t.Fatalf("error = %v, want the fail-closed lane diagnostic", err)
	}
	if fetches := transportFetchLines(t, logPath); len(fetches) != 1 {
		t.Fatalf("fetch count = %d, want 1", len(fetches))
	}
	if len(records) != 1 || records[0].Class != FailureUnknown || !records[0].NetworkAttempted {
		t.Fatalf("records = %+v, want one attempted unknown record", records)
	}
}

// TestResolvedTransportClosedTableFramingStillFallsBack proves the
// framing rule at the production entry: git's fixed ssh/rpc trailers
// beside a positively identified availability diagnostic do not close
// the plan, so the ready alternate is attempted and proves the locked
// content. Any other extra line — wrapper chatter, warnings, unknown
// facilities — poisons the output instead (see the poison unit cases
// and the closed-grammar table test below).
func TestResolvedTransportClosedTableFramingStillFallsBack(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", false)
	tool, logPath := fakeTransportGitTool(t, map[string]transportBehavior{
		transportHTTPS: failTransport("error: RPC failed; HTTP 503 curl 22 The requested URL returned error: 503\nfatal: the remote end hung up unexpectedly"),
		transportSSH:   {succeed: true, fileRepo: fixture.bare},
	})
	tool = transportSSHBaseTool(t, tool)
	plan := TransportPlan{
		Identity: "fixture.test/repository",
		Attempts: []TransportAttempt{{URL: transportHTTPS}, {URL: transportSSH}},
		Fallback: TransportFallbackAvailabilityAuth,
	}
	var records []AttemptRecord
	snapshot, err := AcquireNetworkResolved(context.Background(), transportBase(t, tool, fixture), plan, nil, func(record AttemptRecord) {
		records = append(records, record)
	})
	if err != nil {
		t.Fatalf("framing beside RPC 503 closed the plan: %v (records = %+v)", err, records)
	}
	if snapshot.Commit != fixture.commit {
		t.Fatalf("commit = %q, want locked %q", snapshot.Commit, fixture.commit)
	}
	if fetches := transportFetchLines(t, logPath); len(fetches) != 2 {
		t.Fatalf("fetch count = %d, want 2", len(fetches))
	}
	if len(records) != 2 || records[0].Class != FailureAvailability || !records[1].Succeeded {
		t.Fatalf("records = %+v, want availability then success", records)
	}
}

// requireResolvedLane skips behavior tests of the bounded resolved lane
// where the lane itself is refused: Windows has no fetch process-graph
// lifetime control, so AcquireNetworkResolved returns
// transport_resolution_unsupported_platform there before any traffic.
// (The POSIX shell stand-in additionally needs requirePOSIXTransport;
// tests using it keep both gates.)
func requireResolvedLane(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("bounded resolved lane refuses on windows; behavior is exercised on unix runners")
	}
}

// TestTransportResolutionPlatformGate unit-covers the platform gate on
// every host: Windows refuses with the typed diagnostic, every other
// GOOS passes.
func TestTransportResolutionPlatformGate(t *testing.T) {
	if err := transportResolutionPlatformError("windows"); ErrorCode(err) != CodeTransportResolutionUnsupportedPlatform {
		t.Fatalf("windows gate error = %v, want %s", err, CodeTransportResolutionUnsupportedPlatform)
	}
	for _, goos := range []string{"linux", "darwin", "freebsd", ""} {
		if err := transportResolutionPlatformError(goos); err != nil {
			t.Fatalf("gate(%q) = %v, want nil", goos, err)
		}
	}
}

// TestResolvedTransportWindowsRefusesBeforeAnyProcess runs on Windows
// only and proves the refusal precedes every lane action: the plan is
// valid, yet the tool cannot run and no attempt is recorded. A refusal
// that came after validation, binding, or a spawn would surface a
// different code or a record.
func TestResolvedTransportWindowsRefusesBeforeAnyProcess(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows refusal probe is exercised on Windows")
	}
	source, err := ParseSource(transportHTTPS)
	if err != nil {
		t.Fatal(err)
	}
	base := NetworkRequest{
		Source: source,
		Lock:   LockedCommit{ObjectFormat: "sha1", Hex: strings.Repeat("1", 40)},
		// A tool that cannot run: any process spawn fails loudly,
		// proving the refusal precedes every Git call.
		Tool: GitTool{Executable: filepath.Join(t.TempDir(), "no-such-git.exe")},
	}
	plan := TransportPlan{
		Identity: source.Identity,
		Attempts: []TransportAttempt{{URL: transportHTTPS}},
		Fallback: TransportFallbackNone,
	}
	var records []AttemptRecord
	_, err = AcquireNetworkResolved(context.Background(), base, plan, nil, func(record AttemptRecord) {
		records = append(records, record)
	})
	if ErrorCode(err) != CodeTransportResolutionUnsupportedPlatform {
		t.Fatalf("error = %v, want %s", err, CodeTransportResolutionUnsupportedPlatform)
	}
	if len(records) != 0 {
		t.Fatalf("records = %+v, want no attempt recorded", records)
	}
}

func TestReviewRev4ClosedGrammar(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", false)
	for _, msg := range []string{
		"fatal: unable to access 'https://fixture.test/repository.git/': Could not resolve host: fixture.test\naudit: policy denied",
		"fatal: unable to access 'https://fixture.test/repository.git/': The requested URL returned error: 503; object verification failed",
		"ssh: Could not resolve hostname fixture.test: unexpected resolver protocol response",
	} {
		t.Run(msg, func(t *testing.T) {
			tool, logPath := fakeTransportGitTool(t, map[string]transportBehavior{transportHTTPS: failTransport(msg), transportSSH: {succeed: true, fileRepo: fixture.bare}})
			tool = transportSSHBaseTool(t, tool)
			base := transportBase(t, tool, fixture)
			plan := TransportPlan{Identity: base.Source.Identity, Attempts: []TransportAttempt{{URL: transportHTTPS}, {URL: transportSSH}}, Fallback: TransportFallbackAvailabilityAuth}
			_, err := AcquireNetworkResolved(context.Background(), base, plan, nil, nil)
			n := len(transportFetchLines(t, logPath))
			if err == nil || n != 1 {
				t.Fatalf("fail-closed diagnostic produced %d fetches, err=%v", n, err)
			}
		})
	}
}

// TestResolvedTransportClosedGrammarTable enumerates every
// fallback-eligible accepted line shape at the production entry: each
// shape as a first-attempt failure must permit the ready alternate (two
// fetches, locked content proved, first record carrying the expected
// class), while three one-character mutations of the same shape — an
// appended tail byte, a removed quote or a broken facility separator,
// and one extra unknown-facility line — must refuse after exactly one
// fetch with the lane diagnostic.
func TestResolvedTransportClosedGrammarTable(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", false)
	httpsShapes := []struct {
		name   string
		stderr string
		class  FailureClass
	}{
		{"https-dns", "fatal: unable to access 'https://fixture.test/repository.git/': Could not resolve host: fixture.test", FailureAvailability},
		{"https-op-timeout", "fatal: unable to access 'https://fixture.test/repository.git/': Operation timed out after 30000 milliseconds with 0 bytes received", FailureAvailability},
		{"https-connect-unreachable", "fatal: unable to access 'https://fixture.test/repository.git/': Failed to connect to fixture.test: Network is unreachable", FailureAvailability},
		{"https-503", "fatal: unable to access 'https://fixture.test/repository.git/': The requested URL returned error: 503", FailureAvailability},
		{"https-rpc-503", "error: RPC failed; HTTP 503 curl 22 The requested URL returned error: 503", FailureAvailability},
		{"https-401", "fatal: unable to access 'https://fixture.test/repository.git/': The requested URL returned error: 401", FailureAuth},
		{"https-auth-failed", "fatal: Authentication failed for 'https://fixture.test/repository.git/'", FailureAuth},
		{"https-no-username", "fatal: could not read Username for 'https://fixture.test/repository.git/': terminal prompts disabled", FailureAuth},
		{"remote-invalid-password", "remote: Invalid username or password.", FailureAuth},
		{"remote-permission", "remote: Permission to fixture/repository.git denied to deploy-key.", FailureAuth},
		{"bare-conn-timed-out", "fatal: connection timed out", FailureAvailability},
	}
	sshShapes := []struct {
		name   string
		stderr string
		class  FailureClass
	}{
		{"ssh-resolve", "ssh: Could not resolve hostname fixture.test: Temporary failure in name resolution", FailureAvailability},
		{"ssh-refused", "ssh: connect to host fixture.test port 22: Connection refused", FailureAvailability},
		{"ssh-permission", "git@fixture.test: Permission denied (publickey).", FailureAuth},
	}
	runShape := func(t *testing.T, first, second string, shape struct {
		name   string
		stderr string
		class  FailureClass
	}) {
		t.Helper()
		mutations := []struct {
			name   string
			stderr string
		}{
			{"tail", shape.stderr + ";"},
			{"break", breakClosedGrammarLine(shape.stderr)},
			{"facility", shape.stderr + "\naudit: policy denied"},
		}
		t.Run(shape.name+"/positive", func(t *testing.T) {
			tool, logPath := fakeTransportGitTool(t, map[string]transportBehavior{
				first:  failTransport(shape.stderr),
				second: {succeed: true, fileRepo: fixture.bare},
			})
			tool = transportSSHBaseTool(t, tool)
			plan := TransportPlan{
				Identity: "fixture.test/repository",
				Attempts: []TransportAttempt{{URL: first}, {URL: second}},
				Fallback: TransportFallbackAvailabilityAuth,
			}
			var records []AttemptRecord
			snapshot, err := AcquireNetworkResolved(context.Background(), transportBase(t, tool, fixture), plan, nil, func(record AttemptRecord) {
				records = append(records, record)
			})
			if err != nil {
				t.Fatalf("accepted shape refused: %v (records = %+v)", err, records)
			}
			if snapshot.Commit != fixture.commit {
				t.Fatalf("commit = %q, want locked %q", snapshot.Commit, fixture.commit)
			}
			if fetches := transportFetchLines(t, logPath); len(fetches) != 2 {
				t.Fatalf("fetch count = %d, want 2", len(fetches))
			}
			if len(records) != 2 || records[0].Class != shape.class || !records[0].NetworkAttempted || !records[1].Succeeded {
				t.Fatalf("records = %+v, want attempted %s then success", records, shape.class)
			}
		})
		for _, mutation := range mutations {
			t.Run(shape.name+"/"+mutation.name, func(t *testing.T) {
				tool, logPath := fakeTransportGitTool(t, map[string]transportBehavior{
					first:  failTransport(mutation.stderr),
					second: {succeed: true, fileRepo: fixture.bare},
				})
				tool = transportSSHBaseTool(t, tool)
				plan := TransportPlan{
					Identity: "fixture.test/repository",
					Attempts: []TransportAttempt{{URL: first}, {URL: second}},
					Fallback: TransportFallbackAvailabilityAuth,
				}
				var records []AttemptRecord
				_, err := AcquireNetworkResolved(context.Background(), transportBase(t, tool, fixture), plan, nil, func(record AttemptRecord) {
					records = append(records, record)
				})
				if err == nil {
					t.Fatal("mutated shape fell back to a ready alternate and succeeded")
				}
				if err.Error() != CodeSourceUnavailable+": exact source fetch failed" {
					t.Fatalf("error = %v, want the fail-closed lane diagnostic", err)
				}
				if fetches := transportFetchLines(t, logPath); len(fetches) != 1 {
					t.Fatalf("fetch count = %d, want 1", len(fetches))
				}
				if len(records) != 1 || records[0].Class != FailureUnknown || !records[0].NetworkAttempted {
					t.Fatalf("records = %+v, want one attempted unknown record", records)
				}
			})
		}
	}
	for _, shape := range httpsShapes {
		t.Run(shape.name, func(t *testing.T) {
			runShape(t, transportHTTPS, transportSSH, shape)
		})
	}
	for _, shape := range sshShapes {
		t.Run(shape.name, func(t *testing.T) {
			runShape(t, transportSSH, transportHTTPS, shape)
		})
	}
}

// breakClosedGrammarLine applies the one-character structural mutation
// for the closed-grammar table test: it removes the first single quote
// from quoted lines, or breaks the facility separator of unquoted
// lines. Either change leaves no table entry matching.
func breakClosedGrammarLine(line string) string {
	if i := strings.Index(line, "'"); i >= 0 {
		return line[:i] + line[i+1:]
	}
	if i := strings.Index(line, ":"); i >= 0 {
		return line[:i] + ";" + line[i+1:]
	}
	return line[:len(line)-1]
}
