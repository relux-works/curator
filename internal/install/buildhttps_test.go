package install

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/gitcred"
)

type fakeBuildHTTPSReader struct {
	hosts  map[string]gitcred.HostCredential
	scopes map[string]string
	reads  []string
}

func (r *fakeBuildHTTPSReader) ReadHost(_ context.Context, host string) (gitcred.HostCredential, bool) {
	r.reads = append(r.reads, "host:"+host)
	credential, ok := r.hosts[host]
	return credential, ok
}

func (r *fakeBuildHTTPSReader) ReadScoped(_ context.Context, scope, host string) (string, bool) {
	r.reads = append(r.reads, "scope:"+scope+"@"+host)
	secret, ok := r.scopes[scope]
	return secret, ok
}

func (r *fakeBuildHTTPSReader) Discover(_ context.Context, host string, scopes []string) gitcred.HostMaterial {
	r.reads = append(r.reads, "discover:"+host)
	material := gitcred.HostMaterial{}
	if credential, ok := r.hosts[host]; ok {
		material.HostCredential = true
		material.HostUsername = credential.Username
	}
	for _, scope := range scopes {
		if _, ok := r.scopes[scope]; ok {
			material.Scopes = append(material.Scopes, scope)
		}
	}
	return material
}

func httpsRow(skill, command, identity string) plannedExternal {
	return externalRow(skill, command, "network-git", identity, "https")
}

func resolvedHTTPSFor(t *testing.T, credentials map[buildHTTPSKey]BuildHTTPSCredentials, skill, command string) BuildHTTPSCredentials {
	t.Helper()
	selected, ok := credentials[buildHTTPSKey{skill: skill, command: command}]
	if !ok {
		t.Fatalf("%s.%s selected no HTTPS credentials", skill, command)
	}
	return selected
}

func TestBuildHTTPSPrecedenceLongestScopeAndAnonymousFallback(t *testing.T) {
	selection := BuildHTTPSSelection{
		Override: CapturedBuildHTTPSOverride{secret: "run-wide-secret"},
		Scopes: map[string]config.BuildHTTPSCredential{
			"git.example.test":         {Scope: "git.example.test", TokenEnv: "HOST_TOKEN"},
			"git.example.test/portals": {Scope: "git.example.test/portals", TokenEnv: "PORTALS_TOKEN", Username: "oauth2"},
		},
		environment: map[string]string{"HOST_TOKEN": "host-secret", "PORTALS_TOKEN": "scope-secret"},
	}
	rows := []plannedExternal{
		httpsRow("portals", "build", "git.example.test/portals/app"),
		httpsRow("public", "build", "public.example.test/open/tool"),
	}
	credentials, provenance, err := resolveBuildHTTPS(context.Background(), selection, rows)
	if err != nil {
		t.Fatal(err)
	}
	if got := resolvedHTTPSFor(t, credentials, "portals", "build"); !got.Selected() || got.Secret() != "run-wide-secret" {
		t.Fatalf("precedence selected credential %+v, want run-wide override", got)
	}
	if got := resolvedHTTPSFor(t, credentials, "public", "build"); got.Secret() != "run-wide-secret" {
		t.Fatalf("unpinned override did not cover the unmatched host: %q", got.Secret())
	}
	joined := strings.Join(provenance, "\n")
	for _, want := range []string{"operator environment override", "public.example.test/open/tool"} {
		if !strings.Contains(joined, want) {
			t.Errorf("provenance missing %q:\n%s", want, joined)
		}
	}

	selection.Override = CapturedBuildHTTPSOverride{}
	credentials, provenance, err = resolveBuildHTTPS(context.Background(), selection, rows)
	if err != nil {
		t.Fatal(err)
	}
	selected := resolvedHTTPSFor(t, credentials, "portals", "build")
	if selected.Secret() != "scope-secret" || selected.Username != "oauth2" {
		t.Fatalf("longest scope selected %+v, want portals token and username", selected)
	}
	if _, selected := credentials[buildHTTPSKey{skill: "public", command: "build"}]; selected {
		t.Fatal("an unmatched public repository did not remain anonymous")
	}
	if joined := strings.Join(provenance, "\n"); !strings.Contains(joined, "public.example.test/open/tool") || !strings.Contains(joined, "anonymous") {
		t.Fatalf("anonymous provenance missing: %s", joined)
	}
}

func TestBuildHTTPSDiscoveryOnlyListsAndNeverSelectsAHostCredential(t *testing.T) {
	reader := &fakeBuildHTTPSReader{hosts: map[string]gitcred.HostCredential{
		"git.example.test": {Username: "operator", Secret: "must-not-be-read"},
	}}
	selection := BuildHTTPSSelection{
		Reader: reader,
		Resolve: func(_ context.Context, requests []BuildHTTPSRequest, candidates map[string]gitcred.HostMaterial, _ BuildHTTPSCredentialReader) (map[string]BuildHTTPSCredentials, error) {
			if len(requests) != 1 || !candidates["git.example.test"].HostCredential {
				t.Fatalf("prompt input = requests %v candidates %v", requests, candidates)
			}
			return nil, nil
		},
	}
	credentials, provenance, err := resolveBuildHTTPS(context.Background(), selection, []plannedExternal{
		httpsRow("public", "build", "git.example.test/team/app"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(credentials) != 0 || len(provenance) != 1 || !strings.Contains(provenance[0], "anonymous") {
		t.Fatalf("discovery selected material: credentials=%v provenance=%v", credentials, provenance)
	}
	if got := strings.Join(reader.reads, ","); got != "discover:git.example.test" {
		t.Fatalf("discovery read credential material instead of presence only: %s", got)
	}
}

func TestBuildHTTPSHostPinMakesOtherHostsResolveWithoutTheOverride(t *testing.T) {
	selection := BuildHTTPSSelection{
		Override: CapturedBuildHTTPSOverride{Host: "git.example.test", secret: "pinned-secret"},
		Scopes: map[string]config.BuildHTTPSCredential{
			"other.example.test/team": {Scope: "other.example.test/team", TokenEnv: "OTHER_TOKEN"},
		},
		environment: map[string]string{"OTHER_TOKEN": "scoped-secret"},
	}
	rows := []plannedExternal{
		httpsRow("pinned", "build", "git.example.test/team/app"),
		httpsRow("scoped", "build", "other.example.test/team/app"),
		httpsRow("anonymous", "build", "third.example.test/open/app"),
	}
	credentials, _, err := resolveBuildHTTPS(context.Background(), selection, rows)
	if err != nil {
		t.Fatal(err)
	}
	if got := resolvedHTTPSFor(t, credentials, "pinned", "build").Secret(); got != "pinned-secret" {
		t.Fatalf("pinned host secret = %q", got)
	}
	if got := resolvedHTTPSFor(t, credentials, "scoped", "build").Secret(); got != "scoped-secret" {
		t.Fatalf("non-covered host did not resolve as if the override were absent: %q", got)
	}
	if _, selected := credentials[buildHTTPSKey{skill: "anonymous", command: "build"}]; selected {
		t.Fatal("a non-covered unmatched host received the pinned override")
	}
}

func TestCaptureBuildHTTPSSelectionFreezesAndRedactsSecrets(t *testing.T) {
	environment := map[string]string{
		EnvBuildHTTPSToken: "captured-override-secret",
		EnvBuildHTTPSHost:  "git.example.test",
		"SCOPED_TOKEN":     "captured-scope-secret",
	}
	cfg := &config.Config{BuildHTTPS: map[string]config.BuildHTTPSCredential{
		"other.example.test": {Scope: "other.example.test", TokenEnv: "SCOPED_TOKEN"},
	}}
	selection := CaptureBuildHTTPSSelection(cfg, func(name string) string { return environment[name] })
	if selection.Override.Secret() != "captured-override-secret" {
		t.Fatal("capture did not retain the override")
	}
	environment[EnvBuildHTTPSToken] = "later-override-secret"
	environment["SCOPED_TOKEN"] = "later-scope-secret"

	credentials, _, err := resolveBuildHTTPS(context.Background(), selection, []plannedExternal{
		httpsRow("pinned", "build", "git.example.test/team/app"),
		httpsRow("scoped", "build", "other.example.test/team/app"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := resolvedHTTPSFor(t, credentials, "pinned", "build").Secret(); got != "captured-override-secret" {
		t.Fatalf("override was reread after capture: %q", got)
	}
	if got := resolvedHTTPSFor(t, credentials, "scoped", "build").Secret(); got != "captured-scope-secret" {
		t.Fatalf("scope environment was reread after capture: %q", got)
	}
	for _, diagnostic := range []string{
		fmt.Sprintf("%v", selection.Override), fmt.Sprintf("%+v", selection.Override), fmt.Sprintf("%#v", selection.Override),
		fmt.Sprintf("%v", selection), fmt.Sprintf("%+v", selection), fmt.Sprintf("%#v", selection),
		fmt.Sprintf("%v", resolvedHTTPSFor(t, credentials, "pinned", "build")),
		fmt.Sprintf("%#v", resolvedHTTPSFor(t, credentials, "pinned", "build")),
	} {
		for _, secret := range []string{"captured-override-secret", "captured-scope-secret"} {
			if strings.Contains(diagnostic, secret) {
				t.Fatalf("diagnostic rendered secret %q: %s", secret, diagnostic)
			}
		}
		if !strings.Contains(diagnostic, "<redacted>") {
			t.Fatalf("diagnostic does not state redaction: %s", diagnostic)
		}
	}
}

func TestBuildHTTPSResolutionSkipsOtherEffectiveTransports(t *testing.T) {
	reader := &fakeBuildHTTPSReader{}
	selection := BuildHTTPSSelection{
		Override: CapturedBuildHTTPSOverride{secret: "must-not-be-used"}, Reader: reader,
	}
	rows := []plannedExternal{
		sshRow("ssh", "build", "git.example.test/team/app"),
		externalRow("local", "build", "operator-local-git", "local/team/app", ""),
	}
	credentials, provenance, err := resolveBuildHTTPS(context.Background(), selection, rows)
	if err != nil {
		t.Fatal(err)
	}
	if len(credentials) != 0 || len(provenance) != 0 || len(reader.reads) != 0 {
		t.Fatalf("transport skip produced credentials=%v provenance=%v reads=%v", credentials, provenance, reader.reads)
	}
}

func TestBuildHTTPSSelectedSourcesFailClosedWithExactRemedies(t *testing.T) {
	tests := []struct {
		name       string
		credential config.BuildHTTPSCredential
		want       []string
	}{
		{
			name: "token env", credential: config.BuildHTTPSCredential{Scope: "git.example.test/team", TokenEnv: "TEAM_TOKEN"},
			want: []string{"TEAM_TOKEN is empty", "set TEAM_TOKEN"},
		},
		{
			name: "git credentials", credential: config.BuildHTTPSCredential{Scope: "git.example.test/team", Token: config.TokenSourceGitCredentials},
			want: []string{"no Git HTTPS credential", "'git credential approve'", "protocol=https", "host=git.example.test"},
		},
		{
			name: "keyring", credential: config.BuildHTTPSCredential{Scope: "git.example.test/team", Token: config.TokenSourceKeyring},
			want: []string{"no manager credential", "'curator config build-https login git.example.test/team'"},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			selection := BuildHTTPSSelection{
				Scopes: map[string]config.BuildHTTPSCredential{testCase.credential.Scope: testCase.credential},
				Reader: &fakeBuildHTTPSReader{}, environment: map[string]string{},
			}
			_, _, err := resolveBuildHTTPS(context.Background(), selection, []plannedExternal{
				httpsRow("team", "build", "git.example.test/team/app"),
			})
			if err == nil {
				t.Fatal("selected source with absent material was admitted")
			}
			for _, want := range append([]string{`build_https scope "git.example.test/team"`, "git.example.test/team/app"}, testCase.want...) {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("diagnostic missing %q:\n%s", want, err)
				}
			}
		})
	}
}

func TestBuildHTTPSCredentialReadersMaterializeConfiguredSources(t *testing.T) {
	reader := &fakeBuildHTTPSReader{
		hosts: map[string]gitcred.HostCredential{
			"git.example.test": {Username: "operator", Secret: "operator-secret"},
		},
		scopes: map[string]string{"key.example.test/team": "keyring-secret"},
	}
	selection := BuildHTTPSSelection{
		Scopes: map[string]config.BuildHTTPSCredential{
			"git.example.test/team": {Scope: "git.example.test/team", Token: config.TokenSourceGitCredentials},
			"key.example.test/team": {Scope: "key.example.test/team", Token: config.TokenSourceKeyring, Username: "oauth2"},
		},
		Reader: reader,
	}
	credentials, _, err := resolveBuildHTTPS(context.Background(), selection, []plannedExternal{
		httpsRow("operator", "build", "git.example.test/team/app"),
		httpsRow("keyring", "build", "key.example.test/team/app"),
	})
	if err != nil {
		t.Fatal(err)
	}
	operator := resolvedHTTPSFor(t, credentials, "operator", "build")
	if operator.Username != "operator" || operator.Secret() != "operator-secret" {
		t.Fatalf("operator credential = %+v", operator)
	}
	keyring := resolvedHTTPSFor(t, credentials, "keyring", "build")
	if keyring.Username != "oauth2" || keyring.Secret() != "keyring-secret" {
		t.Fatalf("keyring credential = %+v", keyring)
	}
}

func TestBuildHTTPSResolutionIsCarriedByTheExternalInstallPlan(t *testing.T) {
	node := sshBuildNode("team", "build", "git.example.test/team/app")
	repository := node.Spec.BuildRepositories["tools"]
	repository.Git = "https://git.example.test/team/app.git"
	repository.Transport = "https"
	node.Spec.BuildRepositories["tools"] = repository
	deps := ExternalDeps{
		StoreRoot: t.TempDir(),
		BuildHTTPS: BuildHTTPSSelection{
			Scopes: map[string]config.BuildHTTPSCredential{
				"git.example.test/team": {Scope: "git.example.test/team", TokenEnv: "TEAM_TOKEN"},
			},
			environment: map[string]string{"TEAM_TOKEN": "plan-secret"},
		},
		Audit: func(context.Context, buildrepo.AuditSubject) error { return nil },
		Acquire: func(context.Context, ExternalSource) (*buildrepo.Snapshot, error) {
			return nil, fmt.Errorf("stop after resolution")
		},
	}
	plan, err := planExternalBuilds(context.Background(), "project", "project/id", t.TempDir(),
		[]*closure.Node{node}, nil, probeOnlyToolchain{}, deps, NewPortableBuildAuthority(), true)
	if err == nil || !strings.Contains(err.Error(), buildrepo.CodeSourceUnavailable) {
		t.Fatalf("plan error = %v, want the post-resolution acquisition stop", err)
	}
	if len(plan.rows) != 1 {
		t.Fatalf("plan rows = %d, want one", len(plan.rows))
	}
	selected := plan.httpsCredentialsFor(plan.rows[0])
	if selected.Secret() != "plan-secret" {
		t.Fatalf("install plan did not carry the resolved credential: %+v", selected)
	}
	if len(plan.messages) != 1 || !strings.Contains(plan.messages[0], `config scope "git.example.test/team"`) {
		t.Fatalf("install plan provenance = %v", plan.messages)
	}
}

func TestPromptedBuildHTTPSAbortStopsTheProductionPlanBeforeAnyFetch(t *testing.T) {
	node := sshBuildNode("team", "build", "git.example.test/team/app")
	repository := node.Spec.BuildRepositories["tools"]
	repository.Git = "https://git.example.test/team/app.git"
	repository.Transport = "https"
	node.Spec.BuildRepositories["tools"] = repository
	fetched := 0
	reader := &fakeBuildHTTPSReader{hosts: map[string]gitcred.HostCredential{
		"git.example.test": {Username: "operator", Secret: "candidate-secret"},
	}}
	deps := ExternalDeps{
		StoreRoot: t.TempDir(),
		BuildHTTPS: BuildHTTPSSelection{
			Reader: reader,
			Resolve: func(_ context.Context, requests []BuildHTTPSRequest, candidates map[string]gitcred.HostMaterial, _ BuildHTTPSCredentialReader) (map[string]BuildHTTPSCredentials, error) {
				if fetched != 0 {
					t.Fatal("fetch ran before the HTTPS credential precheck")
				}
				if len(requests) != 1 || !candidates["git.example.test"].HostCredential {
					t.Fatalf("precheck input = requests %v candidates %v", requests, candidates)
				}
				return nil, ErrBuildHTTPSAborted
			},
		},
		Audit: func(context.Context, buildrepo.AuditSubject) error { return nil },
		Acquire: func(context.Context, ExternalSource) (*buildrepo.Snapshot, error) {
			fetched++
			return nil, errors.New("fetch should not run")
		},
	}
	_, err := planExternalBuilds(context.Background(), "project", "project/id", t.TempDir(),
		[]*closure.Node{node}, nil, probeOnlyToolchain{}, deps, NewPortableBuildAuthority(), false)
	if !errors.Is(err, ErrBuildHTTPSAborted) {
		t.Fatalf("plan error = %v, want prompt abort", err)
	}
	if fetched != 0 {
		t.Fatalf("%d fetches ran after prompt abort", fetched)
	}
}

func TestResolvedHTTPSCredentialIsBoundToOnlyItsEffectiveFetch(t *testing.T) {
	source := ExternalSource{Effective: buildrepo.EffectiveState{
		IdentityKind: "network-git", Identity: "git.example.test/team/app", Transport: "https",
	}}
	selected := NewBuildHTTPSCredentials("oauth2", "fetch-secret")
	tool := externalGitTool(buildrepo.GitTool{}, source, buildrepo.OperatorSSHCredentials{}, selected)
	if !tool.HTTPSCredentials.Selected() || tool.HTTPSCredentials.Host != "git.example.test" || tool.HTTPSCredentials.Username != "oauth2" {
		t.Fatalf("bound HTTPS credential = %+v", tool.HTTPSCredentials)
	}
	if strings.Contains(fmt.Sprintf("%#v", tool), "fetch-secret") {
		t.Fatal("Git tool diagnostic disclosed the bound secret")
	}

	other := externalGitTool(buildrepo.GitTool{}, ExternalSource{Effective: buildrepo.EffectiveState{
		IdentityKind: "network-git", Identity: "other.example.test/team/app", Transport: "https",
	}}, buildrepo.OperatorSSHCredentials{}, BuildHTTPSCredentials{})
	if other.HTTPSCredentials.Selected() {
		t.Fatal("an anonymous fetch inherited another repository's credential")
	}
}
