# Draft bounded transport resolution boundary

This is an internal implementation of unreleased `repository-transport-v1`
revision 1 §2 (bounded authenticated endpoint resolution) on the strict
external-repository lane, not an enabled fetch path. Frozen v1 and release
qualification are unchanged.

`buildrepo.AcquireNetworkResolved` applies a `TransportPlan` — one or two
closed-grammar endpoints for one canonical identity, an opaque operator
provider reference per endpoint, and an effective fallback mode — to the
existing strict acquisition: trusted Git, clean configuration/environment,
closed process graph, per-attempt SSH wrapper policy, per-fetch HTTPS
credential broker, exact-ref fetch, and raw-object proof of the locked
content. Pin is already applied when the plan is built (one attempt, no
fallback), so the executor needs no pin logic.

At most one fetch per listed endpoint, no implicit retry, never a
synthesized URL, and one total deadline shared by both attempts. Every
fetching attempt runs the same admission checks as a direct lane call
under the shared deadline, so resolved acquisition refuses exactly what
the lane refuses. Every SSH attempt additionally binds its endpoint and
its selected credentials into a real wrapper policy through the existing
`SSHPolicyFor` builder; the fetch then runs behind a per-attempt wrapper
copy pinned to that policy, and a selection that cannot form a policy
(an identity without pinned host keys, unresolvable paths, no manager
base) refuses before any traffic. The second endpoint is attempted only
under `availability-auth` after a positively classified availability or
authentication failure (DNS, refused/timeout, HTTP 502/503/504, missing
auth method, explicit rejection). Classification is a closed line table
over the captured fetch stderr: the output is fallback-eligible only
when every non-empty line matches exactly one full-line entry — fixed
text plus bounded tokens (dotted hostname, numeric port, three-digit
status, single-quoted endpoint span, enumerated reason phrase) — and at
least one matched line is an availability or authentication diagnostic.
Patterns that carry the endpoint consume it inside the quoted span, so
endpoint text can never read as a signal. Only git's fixed ssh/rpc
failure framing is skipped as non-evidence, and only on real newline
boundaries — the classifier never invents line breaks, so a helper
that glues framing onto its own final line without a newline fails
closed like any other unmatched line. Any other
unmatched line — an unknown facility, a future-git sentence, a secret
echo, an informational warning, a progress line, any unconsumed tail —
poisons the whole output to unknown. TLS, host-key, ref, identity, integrity, audit, unknown,
truncated, and ambiguous-404 failures fail closed: the lane's own
diagnostic is returned unchanged, with no alternate traffic and no
cache shortcut. SSH authentication rejection requires the daemon's
parenthesized method list; a bare "Permission denied" is a local
failure and stays unknown, as is a local-filesystem failure quoting a
signal as a filename. Exhaustion returns
`repository_endpoint_unavailable` with sanitized classifications built
from a closed vocabulary; fetch output, secrets, and full endpoint URLs
never enter errors. Sanitized per-endpoint records reach only the
machine-private trace callback. A legacy-shape plan (one attempt, no
provider, no fallback) behaves exactly like `AcquireNetwork`.

Provider references resolve only through operator configuration:
`gitcred.Access.ReadProvider` reads HTTPS secrets from the operator's own
credential machinery under a provider namespace disjoint from scope and
host entries, and SSH selections admit absolute operator paths through
the existing validator. Identifiers are never commands or paths. A
provider with no usable material records an auth failure with no fetch
traffic.

The total deadline bounds the whole fetch process graph: each lane Git
child leads its own process group where the platform allows,
cancellation kills the group, and pipe draining past the deadline is
bounded. Windows has no process-group equivalent in this lane, so the
bounded resolved lane refuses there with
`transport_resolution_unsupported_platform` before any process
creation (after plan validation, with no attempt recorded) rather than
run with an unbounded tree; the legacy lane keeps its WaitDelay-bound
behavior there unchanged.

Revision 2 ports, mirrors, and aliases need a revision-2 reader and are
a separate leaf. `buildsource` was inspected and needs no change:
locked-content verification for network acquisition is the lane's
raw-object proof, reused per attempt.

## Production caller (draft)

The external-repository lane (repository-transport §3: an existing lane
using its URL declaration with an admitted machine policy) selects this
executor behind the draft/opt-in switch
`CURATOR_DRAFT_TRANSPORT_RESOLUTION=1`, read once at the CLI boundary into
`install.ExternalDeps`. The logical `repository` spelling stays parser-only:
no caller here mints it, so resolution always starts from the declared URL.

`install.acquireDraftNetwork` loads the machine policy beside the loaded
manager configuration, resolves the declared URL with
`config.ResolveRepositoryEndpoints`, converts the `config.Resolution` to a
`TransportPlan` field-for-field, and calls `AcquireNetworkResolved`. A
present policy selects the resolved lane; an absent policy file, or the
switch off, runs `AcquireNetwork` with the exact legacy request — the
switch-off path never opens the policy file. A present-but-invalid policy
fails `repository_policy_invalid` before any fetch. On Windows the executor
refuses with `transport_resolution_unsupported_platform` before any
process creation; the CLI pipeline reports every acquisition failure as
the lane diagnostic, so the typed code is visible only below the
pipeline. The trace callback is nil:
sanitized attempt records have no machine-private CLI sink yet.

Policy-named providers resolve only through the operator's provider table,
`source-providers.json` beside the manager configuration, read through
`buildrepo.CredentialProviders`: per-provider HTTPS usernames and explicit
anonymity, per-provider SSH paths, and HTTPS secrets from the operator's
own credential machinery under the provider namespace. An unknown provider
name, or a configured provider with no usable material, is unavailable:
its attempt records an auth failure with no fetch traffic. Only an
explicitly anonymous provider fetches without credentials; a missing
selection is never anonymity. Attempts without a policy provider keep the
tool's bound lane credentials, and the providers file is consulted only
when a planned attempt names a provider — never by the legacy lane. A
present-but-invalid providers file fails `repository_policy_invalid`
before any fetch. Selections never cross providers or repositories.

SSH attempts need two manager-owned inputs beside the per-repository
credential selection, both populated per fetch: the wrapper base (the
platform `ssh` executable plus fresh empty configuration files, removed
after acquisition) and the manager binary itself as the per-attempt wrapper
copy source. A missing SSH executable yields a zero base and the executor
refuses each SSH attempt; an unadmittable manager binary refuses SSH
resolution before any traffic rather than copying a file that would bypass
the bound policy. The manager binary dispatches the wrapper basename before
its public CLI, next to the HTTPS broker dispatch; the wrapper tuple, fixed
argv, and state-file shape are unchanged.

The wiring also repairs the declared-URL threading the lane needs to fetch
at all: the default closure used to pass the repository's configured name
where the lane requires the URL, so every unfetched acquisition refused in
admission. `ExternalSource` now carries the declared URL and the legacy lane
fetches it. The switch-off golden pins the repaired legacy bytes.
