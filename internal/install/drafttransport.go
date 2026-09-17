package install

// Draft bounded transport resolution wiring (repository-transport-v1 §2-3).
//
// The external-repository lane declares URLs only. Behind the draft/opt-in
// switch, one fetch with a present machine policy runs the resolved executor
// over the policy's endpoint plan; with the switch off, or with no policy
// file, the legacy lane runs unchanged. The logical `repository` spelling is
// admitted only in new Skillfile source objects (§3) and no caller here mints
// it, so resolution always starts from the declared URL.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/config"
)

// EnvDraftTransportResolution names the draft/opt-in switch for bounded
// transport resolution. Exactly "1" enables the resolved lane; every other
// value, including unset and empty, keeps the legacy lane. The switch is
// operator-owned: package data can neither set it nor observe it.
const EnvDraftTransportResolution = "CURATOR_DRAFT_TRANSPORT_RESOLUTION"

// draftTransportSSHConnectTimeout is the single connection timeout, in
// seconds, pinned into every per-attempt SSH wrapper policy. The lane's
// total deadline still bounds both attempts.
const draftTransportSSHConnectTimeout = 15

// DraftTransportEnabled reports whether the switch selects the resolved lane.
func DraftTransportEnabled(getenv func(string) string) bool {
	return getenv != nil && getenv(EnvDraftTransportResolution) == "1"
}

// DraftTransportProvenanceFileName names the machine-private operation
// diagnostics file holding sanitized resolved-lane endpoint provenance
// (repository-transport §7). It lives directly under the manager home
// beside the machine policy: the home directory already exists, so no
// new directory is created, and portable artifacts (locks, receipts,
// markers, manifests) never carry these properties.
const DraftTransportProvenanceFileName = "draft-transport-provenance.jsonl"

// DraftTransportProvenancePath resolves the operation-diagnostics sink
// for one manager home.
func DraftTransportProvenancePath(home string) string {
	return filepath.Join(home, DraftTransportProvenanceFileName)
}

// DraftTransportProvenanceTrace returns the production TransportTrace
// for one manager home: every sanitized attempt record is appended as
// one JSON line carrying only the fixed allowlisted provenance fields
// (canonical identity, listed URL, resolved host and port, alias and
// mirror_of when used, lane transport, provider identifier, outcome).
// The record type carries no secrets, and the allowlist keeps it that
// way structurally: a future record field is never serialized unless
// named here. Writes are best-effort and never fail an acquisition; an
// empty home records nothing.
func DraftTransportProvenanceTrace(home string) buildrepo.TransportTrace {
	if home == "" {
		return nil
	}
	return func(record buildrepo.AttemptRecord) {
		appendTransportProvenance(DraftTransportProvenancePath(home), record)
	}
}

// transportProvenanceLine is the exact serialized sink shape: fixed
// sanitized fields only, never secrets, never fetch output.
type transportProvenanceLine struct {
	Index            int    `json:"index"`
	Identity         string `json:"identity"`
	URL              string `json:"url"`
	Transport        string `json:"transport"`
	Provider         string `json:"provider"`
	Class            string `json:"class"`
	NetworkAttempted bool   `json:"network_attempted"`
	Succeeded        bool   `json:"succeeded"`
	ResolvedHost     string `json:"resolved_host"`
	ResolvedPort     int    `json:"resolved_port"`
	Alias            string `json:"alias,omitempty"`
	MirrorOf         string `json:"mirror_of,omitempty"`
}

func appendTransportProvenance(path string, record buildrepo.AttemptRecord) {
	line, err := json.Marshal(transportProvenanceLine{
		Index: record.Index, Identity: record.Identity, URL: record.URL,
		Transport: record.Transport, Provider: record.Provider, Class: string(record.Class),
		NetworkAttempted: record.NetworkAttempted, Succeeded: record.Succeeded,
		ResolvedHost: record.ResolvedHost, ResolvedPort: record.ResolvedPort,
		Alias: record.Alias, MirrorOf: record.MirrorOf,
	})
	if err != nil {
		return
	}
	line = append(line, '\n')
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600) // #nosec G304 -- manager-home operation-diagnostics sink resolved above.
	if err != nil {
		return
	}
	_, _ = file.Write(line)
	_ = file.Close()
}

// acquireDraftNetwork routes one external-repository fetch behind the draft
// switch. git/transport/identity carry the effective fetch endpoint: the
// declared URL, or the network substitution's. The legacy call keeps the
// exact legacy request shape. The resolved call passes the policy's endpoint
// plan; attempts without a policy provider keep the tool's bound lane
// credentials, while named providers resolve only through the operator's
// provider table and trusted broker — never through the repository's own
// lane selection.
func (deps ExternalDeps) acquireDraftNetwork(ctx context.Context, tool buildrepo.GitTool, git, transport, identity string, lock buildrepo.LockedCommit, tag, refKind, refValue string) (*buildrepo.Snapshot, error) {
	base := buildrepo.NetworkRequest{Source: buildrepo.Source{Git: git, Transport: transport, Identity: identity}, Lock: lock, Tag: tag, RefKind: refKind, RefValue: refValue, Tool: tool, Limits: deps.Limits}
	if !deps.DraftTransportResolution {
		return buildrepo.AcquireNetwork(ctx, base)
	}
	path := deps.DraftPolicyPath
	if path == "" {
		path = config.SourcePolicyPath()
	}
	policy, err := config.LoadSourcePolicy(path)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return buildrepo.AcquireNetwork(ctx, base)
	}
	resolution, err := config.ResolveRepositoryEndpoints(policy, git, "")
	if err != nil {
		return nil, err
	}
	plan, err := draftTransportPlan(resolution)
	if err != nil {
		return nil, err
	}
	var providers buildrepo.AuthProvider
	if draftPlanNamesProvider(plan) {
		providersPath := deps.DraftProvidersPath
		if providersPath == "" {
			providersPath = config.SourceProvidersPath()
		}
		set, err := config.LoadSourceProviders(providersPath)
		if err != nil {
			return nil, err
		}
		providers = buildrepo.CredentialProviders{Set: set, Reader: deps.DraftProviderReader}
	}
	if draftPlanNeedsSSH(plan) {
		var cleanup func()
		base.Tool.SSHBase, cleanup, err = draftSSHWrapperBase()
		if err != nil {
			return nil, err
		}
		defer cleanup()
		// The per-attempt wrapper copy must dispatch the manager binary: only
		// it answers the wrapper tuple against the bound state file. An
		// unadmittable manager refuses SSH resolution before any traffic
		// rather than copying a file that would silently bypass the policy.
		manager, err := draftManagerExecutable()
		if err != nil {
			return nil, err
		}
		base.Tool.SSHWrapper = manager
	}
	return buildrepo.AcquireNetworkResolved(ctx, base, plan, providers, deps.DraftTransportTrace)
}

// draftPlanNamesProvider reports whether any planned attempt names a policy
// provider. The operator provider table is consulted only then: a
// providerless resolved fetch keeps the tool's bound lane credentials and
// never opens the providers file.
func draftPlanNamesProvider(plan buildrepo.TransportPlan) bool {
	for _, attempt := range plan.Attempts {
		if attempt.Authentication != "" {
			return true
		}
	}
	return false
}

// draftTransportPlan converts one machine-policy resolution to the executor's
// attempt plan field-for-field, including the revision-2 endpoint
// properties (mirror_of, alias, resolved connection address) the §6
// predicate revalidates. The executor revalidates before any network
// I/O; the validation here fails a mistranslation before SSH base discovery.
// A selected port or alias endpoint is refused by the executor in this
// strict external-build lane (§7); the conversion preserves it so the
// refusal names the planned endpoint rather than a mistranslation.
func draftTransportPlan(resolution config.Resolution) (buildrepo.TransportPlan, error) {
	var fallback string
	switch resolution.Fallback {
	case config.FallbackNone:
		fallback = buildrepo.TransportFallbackNone
	case config.FallbackAvailabilityAuth:
		fallback = buildrepo.TransportFallbackAvailabilityAuth
	default:
		return buildrepo.TransportPlan{}, fmt.Errorf("%s: unknown effective fallback %q", config.CodeRepositoryPolicyInvalid, resolution.Fallback)
	}
	attempts := make([]buildrepo.TransportAttempt, 0, len(resolution.Attempts))
	for _, attempt := range resolution.Attempts {
		attempts = append(attempts, buildrepo.TransportAttempt{URL: attempt.URL, Authentication: attempt.Authentication,
			MirrorOf: attempt.MirrorOf, Alias: attempt.Alias, ResolvedHost: attempt.ResolvedHost,
			ResolvedPort: attempt.ResolvedPort, HasExplicitPort: attempt.HasExplicitPort})
	}
	plan := buildrepo.TransportPlan{Identity: resolution.Identity, Attempts: attempts, Fallback: fallback}
	if err := buildrepo.ValidateTransportPlan(plan); err != nil {
		return buildrepo.TransportPlan{}, err
	}
	return plan, nil
}

// draftPlanNeedsSSH reports whether any planned attempt fetches over SSH.
// Port-bearing ssh:// URLs never parse under the lane's closed grammar,
// so the scheme prefix is the fallback signal for revision-2 endpoints.
func draftPlanNeedsSSH(plan buildrepo.TransportPlan) bool {
	for _, attempt := range plan.Attempts {
		if parsed, err := buildrepo.ParseSource(attempt.URL); err == nil && parsed.Transport == "ssh" {
			return true
		} else if strings.HasPrefix(attempt.URL, "ssh://") {
			return true
		}
	}
	return false
}

// draftSSHWrapperBase discovers the manager-owned base of per-attempt SSH
// wrapper policies: the platform SSH executable and fresh empty configuration
// files owned by this fetch. cleanup removes the empty files; the caller runs
// it after acquisition returns. A missing or unadmitted SSH executable yields
// a zero base with no error: the executor then refuses every SSH attempt
// before any traffic. Only empty-file creation failures are errors.
func draftSSHWrapperBase() (buildrepo.SSHWrapperBase, func(), error) {
	none := func() {}
	ssh, err := exec.LookPath("ssh")
	if err != nil {
		return buildrepo.SSHWrapperBase{}, none, nil
	}
	ssh, err = filepath.EvalSymlinks(ssh)
	if err != nil {
		return buildrepo.SSHWrapperBase{}, none, nil
	}
	if info, err := os.Lstat(ssh); err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return buildrepo.SSHWrapperBase{}, none, nil
	}
	dir, err := os.MkdirTemp("", "curator-draft-ssh-")
	if err != nil {
		return buildrepo.SSHWrapperBase{}, none, err
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	emptyConfig := filepath.Join(dir, "ssh_config")
	if err := os.WriteFile(emptyConfig, nil, 0o600); err != nil {
		cleanup()
		return buildrepo.SSHWrapperBase{}, none, err
	}
	emptyKnownHosts := filepath.Join(dir, "known_hosts")
	if err := os.WriteFile(emptyKnownHosts, nil, 0o600); err != nil {
		cleanup()
		return buildrepo.SSHWrapperBase{}, none, err
	}
	return buildrepo.SSHWrapperBase{SSH: ssh, EmptyConfig: emptyConfig, EmptyKnownHosts: emptyKnownHosts, ConnectTimeout: draftTransportSSHConnectTimeout}, cleanup, nil
}

// draftManagerExecutable admits the running manager binary as the SSH wrapper
// copy source for resolved SSH attempts.
func draftManagerExecutable() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("build_repository_identity_invalid: SSH transport resolution requires the manager executable: %v", err)
	}
	resolved, err := filepath.EvalSymlinks(executable)
	if err != nil || !filepath.IsAbs(resolved) {
		return "", fmt.Errorf("build_repository_identity_invalid: SSH transport resolution requires an absolute manager executable")
	}
	if info, err := os.Lstat(resolved); err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("build_repository_identity_invalid: SSH transport resolution requires an admitted manager executable")
	}
	return resolved, nil
}
