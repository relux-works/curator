# Bounded repository transport resolution

This documents bounded repository endpoint resolution under
`repository-transport` revisions 1 and 2. Project Skillfile source
resolution uses the policy planner by default. External build-repository
resolution remains behind the separate `CURATOR_DRAFT_TRANSPORT_RESOLUTION`
operator gate and is outside the Skillfile source default.

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

Revision 2 ports, mirrors, and aliases (repository-transport §§5-7)
resolve through the same executor. `config.ResolveRepositoryEndpoints`
plans the §6 attempt order over the revision-2 reader's model
(`mirror_of`, `alias`, resolved connection host and port per attempt);
the executor revalidates the single resolved-host predicate over the
carried values — `repository_mirror_undeclared` for a differing
resolved host without attestation, `repository_policy_invalid` for
every misuse — with zero attempts, no fallback, and no cache shortcut.
The two-endpoint cap, one attempt per endpoint, no implicit retry, the
shared total deadline, and the revision-1 fallback table are unchanged:
ports, mirrors, and aliases are properties of an endpoint, not new
attempts, so a declared mirror or aliased endpoint that fails with an
availability/authentication error follows the ordinary
`availability-auth` rule. `buildsource` was inspected and needs no
change: locked-content verification for network acquisition is the
lane's raw-object proof, reused per attempt, identically for mirrors.

The strict external-build lane admits only the port-free, alias-free
subset (§7): a selected endpoint with an explicit port (URL port or
alias port) or an `alias` field fails
`build_repository_identity_invalid` at plan validation, before any
network I/O — the lane never strips a port or ignores an alias to
force admission, and the SSH wrapper keeps its closed fixed argv. A
mirror endpoint without ports or aliases is an ordinary lane URL and
fetches with verification identical to revision 1, proving the same
canonical identity and locked content.

Sanitized endpoint provenance rides each machine-private attempt
record: the canonical plan identity (never an alias or mirror host),
the listed URL with its port, the resolved connection host and port,
and the alias and `mirror_of` properties when used. Records never
carry fetch output or secrets. Exhaustion still reports only the
closed class vocabulary with the canonical identity; full URLs reach
only the trace callback. Provenance never enters portable artifacts:
no port field is added to receipt inputs (the existing
declared/effective `https`/`ssh` transport enum is unchanged), and
locks, markers, and manifests are untouched. User Git/SSH
configuration (`~/.ssh/config` Host aliases, `insteadOf`,
`ProxyCommand`, helpers, environment overrides) is never consulted:
the lane pins its own configuration paths, broker, and wrapper, and
fetches the declared URL literally when no policy entry applies.

## External build-repository caller

The external-repository lane (repository-transport §3: an existing lane
using its URL declaration with an admitted machine policy) selects this
executor behind the separate opt-in switch
`CURATOR_DRAFT_TRANSPORT_RESOLUTION=1`, read once at the CLI boundary into
`install.ExternalDeps`. The logical `repository` spelling stays parser-only:
no caller here mints it, so resolution always starts from the declared URL.

`install.acquireDraftNetwork` loads the machine policy beside the loaded
manager configuration, resolves the declared URL with
`config.ResolveRepositoryEndpoints`, converts the `config.Resolution` to a
`TransportPlan` field-for-field — including the revision-2 `mirror_of`,
`alias`, and resolved connection address — and calls
`AcquireNetworkResolved`. A present policy selects the resolved lane; an
absent policy file, or the gate disabled, runs `AcquireNetwork` with the
exact legacy request — the disabled-gate path never opens the policy file.
A present-but-invalid policy fails `repository_policy_invalid` before
any fetch; an unattested mirror fails `repository_mirror_undeclared`
and a dangling alias fails `repository_alias_unknown`, both with zero
fetches. A selected port or alias endpoint fails
`build_repository_identity_invalid` in this strict lane before any
fetch. On Windows the executor refuses with
`transport_resolution_unsupported_platform` before any process
creation; the CLI pipeline reports every acquisition failure as the
lane diagnostic, so the typed code is visible only below the pipeline.
The trace callback is `ExternalDeps.DraftTransportTrace`, assigned by
the production CLI to `install.DraftTransportProvenanceTrace(cfg.Home())`:
sanitized attempt records append as one JSON line per endpoint to the
manager-home operation diagnostics log
`draft-transport-provenance.jsonl` (mode 0600, fixed allowlisted fields
only: canonical identity, listed URL, resolved host and port, alias and
`mirror_of` when used, lane transport, provider identifier, outcome) —
never portable artifacts. The legacy lane never invokes it, so
disabled-gate runs create no file.

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
fetches it. The disabled-gate regression test pins the legacy bytes.
