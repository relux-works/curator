package install

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/gitcred"
)

func httpsPromptRequest() BuildHTTPSRequest {
	return BuildHTTPSRequest{
		Skill: "portals", Command: "build-tool", Identity: "git.example.test/portals/app",
		Host: "git.example.test", DefaultScope: "git.example.test/portals",
	}
}

func runHTTPSPrompt(t *testing.T, script, token string) ([]config.BuildHTTPSCredential, []string, map[string]BuildHTTPSCredentials, string, error) {
	t.Helper()
	reader := &fakeBuildHTTPSReader{hosts: map[string]gitcred.HostCredential{
		"git.example.test": {Username: "operator", Secret: "existing-secret"},
	}}
	var persisted []config.BuildHTTPSCredential
	var persistedSecrets []string
	transcript := &strings.Builder{}
	resolver := InteractiveBuildHTTPSResolver(strings.NewReader(script), transcript,
		func() (string, error) { return token, nil },
		func(_ context.Context, credential config.BuildHTTPSCredential, secret string) error {
			persisted = append(persisted, credential)
			persistedSecrets = append(persistedSecrets, secret)
			return nil
		})
	added, err := resolver(context.Background(), []BuildHTTPSRequest{httpsPromptRequest()}, map[string]gitcred.HostMaterial{
		"git.example.test": {HostCredential: true, HostUsername: "operator"},
	}, reader)
	return persisted, persistedSecrets, added, transcript.String(), err
}

func TestHTTPSPromptDefaultSelectsExistingCredentialAndNarrowestSavedScope(t *testing.T) {
	persisted, secrets, added, transcript, err := runHTTPSPrompt(t, "\n\n", "unused")
	if err != nil {
		t.Fatal(err)
	}
	want := config.BuildHTTPSCredential{Scope: "git.example.test/portals", Token: config.TokenSourceGitCredentials}
	if len(persisted) != 1 || persisted[0] != want || len(secrets) != 1 || secrets[0] != "" {
		t.Fatalf("persisted = %+v secrets=%v, want existing source at narrowest scope", persisted, secrets)
	}
	if got := added[want.Scope]; got.Username != "operator" || got.Secret() != "existing-secret" {
		t.Fatalf("selected = %+v", got)
	}
	for _, required := range []string{
		"existing Git HTTPS credential for this host (username operator)  [default]",
		"t) enter a token now",
		"scope [git.example.test/portals]",
		"recorded build_https scope git.example.test/portals",
	} {
		if !strings.Contains(transcript, required) {
			t.Errorf("transcript missing %q:\n%s", required, transcript)
		}
	}
}

func TestHTTPSPromptThisRunOnlyNeverReachesPersistenceOnEitherCredentialSurface(t *testing.T) {
	tests := []struct {
		name, script, token, wantSecret string
	}{
		{name: "existing host credential", script: "\nr\n", token: "unused", wantSecret: "existing-secret"},
		{name: "token entered now", script: "t\nr\n", token: "run-secret", wantSecret: "run-secret"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			persisted, secrets, added, transcript, err := runHTTPSPrompt(t, testCase.script, testCase.token)
			if err != nil {
				t.Fatal(err)
			}
			if len(persisted) != 0 || len(secrets) != 0 {
				t.Fatalf("this-run-only answer reached saved config/keyring: %+v / %v", persisted, secrets)
			}
			if got := added["git.example.test/portals"].Secret(); got != testCase.wantSecret {
				t.Fatalf("run-local secret = %q, want selected material", got)
			}
			if !strings.Contains(transcript, "using HTTPS credential for this run only") {
				t.Fatalf("transcript did not mark run-only choice:\n%s", transcript)
			}
		})
	}
}

func TestHTTPSPromptPersistsEnteredTokenOnlyAfterScopeChoice(t *testing.T) {
	persisted, secrets, added, _, err := runHTTPSPrompt(t, "t\n\n", "saved-secret")
	if err != nil {
		t.Fatal(err)
	}
	want := config.BuildHTTPSCredential{Scope: "git.example.test/portals", Token: config.TokenSourceKeyring}
	if len(persisted) != 1 || persisted[0] != want || len(secrets) != 1 || secrets[0] != "saved-secret" {
		t.Fatalf("persisted = %+v secrets=%v", persisted, secrets)
	}
	if got := added[want.Scope].Secret(); got != "saved-secret" {
		t.Fatalf("selected token = %q", got)
	}
}

func TestHTTPSPromptAbortNeverPersistsOrSelects(t *testing.T) {
	for name, script := range map[string]string{
		"credential choice":           "q\n",
		"scope after existing choice": "1\nq\n",
		"scope after token entry":     "t\nq\n",
		"end of input":                "",
	} {
		t.Run(name, func(t *testing.T) {
			persisted, secrets, added, _, err := runHTTPSPrompt(t, script, "entered-before-abort")
			if !errors.Is(err, ErrBuildHTTPSAborted) {
				t.Fatalf("err = %v, want %v", err, ErrBuildHTTPSAborted)
			}
			if len(persisted) != 0 || len(secrets) != 0 || len(added) != 0 {
				t.Fatalf("aborted prompt changed state: %+v / %v / %+v", persisted, secrets, added)
			}
		})
	}
}
