package install

import (
	"errors"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
)

// promptCandidates is the material every prompt test is offered: a live agent
// and two public keys, so the default entry ("agent, pinned to the first key")
// exists and every other shape is reachable.
func promptCandidates() BuildSSHCandidates {
	return BuildSSHCandidates{
		AgentSocket: "/run/agent.sock", AgentKeys: 2, AgentKeysKnown: true,
		Identities: []string{"~/.ssh/id_ed25519.pub", "~/.ssh/work.pub"},
	}
}

func portalsRequest() BuildSSHRequest {
	return BuildSSHRequest{
		Skill: "portals", Command: "build-tool",
		Identity:     "git.example.test/portals/app",
		DefaultScope: "git.example.test/portals",
	}
}

// runPrompt drives the resolver with scripted stdin and reports what it
// persisted, in order, alongside the transcript the operator would have seen.
func runPrompt(t *testing.T, script string, requests []BuildSSHRequest, candidates BuildSSHCandidates) ([]config.BuildSSHCredential, map[string]config.BuildSSHCredential, string, error) {
	t.Helper()
	var persisted []config.BuildSSHCredential
	transcript := &strings.Builder{}
	resolver := InteractiveBuildSSHResolver(strings.NewReader(script), transcript,
		func(credential config.BuildSSHCredential) error {
			persisted = append(persisted, credential)
			return nil
		})
	added, err := resolver(requests, candidates)
	return persisted, added, transcript.String(), err
}

func TestPromptDefaultSelectionPinsTheAgentToTheFirstDiscoveredKey(t *testing.T) {
	// Two bare newlines: accept the default credential, accept the default
	// scope. Accepting is still an explicit answer — nothing is applied for an
	// operator who never answers.
	persisted, added, transcript, err := runPrompt(t, "\n\n", []BuildSSHRequest{portalsRequest()}, promptCandidates())
	if err != nil {
		t.Fatal(err)
	}
	want := config.BuildSSHCredential{
		Scope: "git.example.test/portals", Agent: true, Identity: "~/.ssh/id_ed25519.pub",
	}
	if len(persisted) != 1 || persisted[0] != want {
		t.Fatalf("persisted = %+v, want exactly %+v", persisted, want)
	}
	if got := added["git.example.test/portals"]; got != want {
		t.Fatalf("returned scope = %+v, want %+v", got, want)
	}
	for _, required := range []string{
		"git.example.test/portals/app needs SSH credentials",
		"1) agent, pinned to ~/.ssh/id_ed25519.pub  [default]",
		"2) SSH agent at /run/agent.sock holding 2 keys, any key it holds",
		"3) identity ~/.ssh/id_ed25519.pub",
		"4) identity ~/.ssh/work.pub",
		"m) enter an identity path",
		"q) abort",
		"scope [git.example.test/portals]",
		"recorded build_ssh scope git.example.test/portals",
	} {
		if !strings.Contains(transcript, required) {
			t.Errorf("transcript missing %q:\n%s", required, transcript)
		}
	}
}

func TestPromptNumberedSelectionAndAnExplicitScope(t *testing.T) {
	persisted, _, _, err := runPrompt(t, "2\ngit.example.test\n", []BuildSSHRequest{portalsRequest()}, promptCandidates())
	if err != nil {
		t.Fatal(err)
	}
	want := config.BuildSSHCredential{Scope: "git.example.test", Agent: true}
	if len(persisted) != 1 || persisted[0] != want {
		t.Fatalf("persisted = %+v, want %+v", persisted, want)
	}
	identityOnly, _, _, err := runPrompt(t, "4\n\n", []BuildSSHRequest{portalsRequest()}, promptCandidates())
	if err != nil {
		t.Fatal(err)
	}
	wantIdentity := config.BuildSSHCredential{
		Scope: "git.example.test/portals", Identity: "~/.ssh/work.pub",
	}
	if len(identityOnly) != 1 || identityOnly[0] != wantIdentity {
		t.Fatalf("persisted = %+v, want %+v", identityOnly, wantIdentity)
	}
}

func TestPromptManualIdentityPathCoversAKeyDiscoveryDidNotList(t *testing.T) {
	persisted, _, transcript, err := runPrompt(t,
		"m\n/operator/keys/portals.pub\ngit.example.test/portals\n",
		[]BuildSSHRequest{portalsRequest()}, promptCandidates())
	if err != nil {
		t.Fatal(err)
	}
	want := config.BuildSSHCredential{
		Scope: "git.example.test/portals", Identity: "/operator/keys/portals.pub",
	}
	if len(persisted) != 1 || persisted[0] != want {
		t.Fatalf("persisted = %+v, want %+v", persisted, want)
	}
	if !strings.Contains(transcript, "identity path (") {
		t.Fatalf("transcript never asked for a path:\n%s", transcript)
	}
}

func TestPromptRejectsAMalformedManualPathAndAsksAgain(t *testing.T) {
	persisted, _, transcript, err := runPrompt(t,
		"m\nkeys/portals.pub\n/operator/keys/portals.pub\n\n",
		[]BuildSSHRequest{portalsRequest()}, promptCandidates())
	if err != nil {
		t.Fatal(err)
	}
	if len(persisted) != 1 || persisted[0].Identity != "/operator/keys/portals.pub" {
		t.Fatalf("persisted = %+v", persisted)
	}
	if !strings.Contains(transcript, "curator: identity path "+config.BuildSSHPathRule) {
		t.Fatalf("a relative path was accepted without a complaint:\n%s", transcript)
	}
}

func TestPromptRejectsAnUnofferedChoiceAndAMalformedScope(t *testing.T) {
	persisted, _, transcript, err := runPrompt(t,
		"9\nzzz\n1\nGit.Example.Test\ngit.example.test/portals\n",
		[]BuildSSHRequest{portalsRequest()}, promptCandidates())
	if err != nil {
		t.Fatal(err)
	}
	if len(persisted) != 1 || persisted[0].Scope != "git.example.test/portals" {
		t.Fatalf("persisted = %+v", persisted)
	}
	for _, required := range []string{
		`curator: "9" is not one of the offered choices`,
		`curator: "zzz" is not one of the offered choices`,
		"curator: scope " + config.BuildSSHScopeRule,
	} {
		if !strings.Contains(transcript, required) {
			t.Errorf("transcript missing %q:\n%s", required, transcript)
		}
	}
}

func TestPromptAbortFailsClosedWithoutPersistingAnything(t *testing.T) {
	for name, script := range map[string]string{
		"at the credential question": "q\n",
		"at the scope question":      "1\nq\n",
		"at the manual path":         "m\nq\n",
		// End of input is an abort too: a question nobody is there to answer
		// must not resolve to the default.
		"end of input before any answer": "",
		"end of input at the scope":      "1\n",
	} {
		persisted, added, _, err := runPrompt(t, script, []BuildSSHRequest{portalsRequest()}, promptCandidates())
		if !errors.Is(err, ErrBuildSSHAborted) {
			t.Errorf("%s: err = %v, want %v", name, err, ErrBuildSSHAborted)
		}
		if len(persisted) != 0 || len(added) != 0 {
			t.Errorf("%s: an aborted prompt persisted %+v / returned %+v", name, persisted, added)
		}
	}
}

func TestAbortCarriesTheProtocolCodeSoTheRunStillFailsClosed(t *testing.T) {
	if !strings.HasPrefix(ErrBuildSSHAborted.Error(), "build_repository_ssh_credential_missing") {
		t.Fatalf("abort error = %q", ErrBuildSSHAborted)
	}
}

func TestOneScopeChoiceCoversEverySiblingRepository(t *testing.T) {
	requests := []BuildSSHRequest{
		portalsRequest(),
		{
			Skill: "portals", Command: "build-agent",
			Identity:     "git.example.test/portals/agent",
			DefaultScope: "git.example.test/portals",
		},
		{
			Skill: "infra", Command: "build-kit",
			Identity:     "other.example.test/tools/kit",
			DefaultScope: "other.example.test/tools",
		},
	}
	// Answers: default for the first repository, then default for the third.
	// The second is never asked about, because the scope just chosen already
	// covers it.
	persisted, added, transcript, err := runPrompt(t, "\n\n\n\n", requests, promptCandidates())
	if err != nil {
		t.Fatal(err)
	}
	if len(persisted) != 2 || len(added) != 2 {
		t.Fatalf("persisted = %+v", persisted)
	}
	if strings.Contains(transcript, "git.example.test/portals/agent") {
		t.Fatalf("a repository already covered by the chosen scope was asked about:\n%s", transcript)
	}
	if _, ok := added["other.example.test/tools"]; !ok {
		t.Fatalf("the second namespace was not covered: %+v", added)
	}
}

func TestPromptWithoutCandidatesStillOffersTheManualPath(t *testing.T) {
	persisted, _, transcript, err := runPrompt(t,
		"m\n/operator/keys/portals.pub\n\n", []BuildSSHRequest{portalsRequest()}, BuildSSHCandidates{})
	if err != nil {
		t.Fatal(err)
	}
	if len(persisted) != 1 || persisted[0].Identity != "/operator/keys/portals.pub" {
		t.Fatalf("persisted = %+v", persisted)
	}
	if !strings.Contains(transcript, "no SSH agent and no ~/.ssh/*.pub identity were detected") {
		t.Fatalf("transcript did not say the host offered nothing:\n%s", transcript)
	}
	// With no entries there is no default to accept, so a bare newline must be
	// refused rather than silently selecting something.
	_, _, bareTranscript, err := runPrompt(t, "\n\n", []BuildSSHRequest{portalsRequest()}, BuildSSHCandidates{})
	if !errors.Is(err, ErrBuildSSHAborted) {
		t.Fatalf("err = %v, want the prompt to run out of input rather than default", err)
	}
	if !strings.Contains(bareTranscript, "is not one of the offered choices") {
		t.Fatalf("an empty answer with no options was accepted:\n%s", bareTranscript)
	}
}

func TestPromptListsTheUnlistedIdentityCount(t *testing.T) {
	candidates := promptCandidates()
	candidates.MoreIdentities = 4
	_, _, transcript, err := runPrompt(t, "\n\n", []BuildSSHRequest{portalsRequest()}, candidates)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(transcript, "4 further ~/.ssh/*.pub file(s) are not listed") {
		t.Fatalf("the capped listing did not report its remainder:\n%s", transcript)
	}
}

func TestAPersistFailureStopsTheRunRatherThanContinuing(t *testing.T) {
	failure := errors.New("config is read-only")
	resolver := InteractiveBuildSSHResolver(strings.NewReader("\n\n"), &strings.Builder{},
		func(config.BuildSSHCredential) error { return failure })
	added, err := resolver([]BuildSSHRequest{portalsRequest()}, promptCandidates())
	if !errors.Is(err, failure) {
		t.Fatalf("err = %v, want %v", err, failure)
	}
	if len(added) != 0 {
		t.Fatalf("a credential that was not persisted was returned as selected: %+v", added)
	}
}
