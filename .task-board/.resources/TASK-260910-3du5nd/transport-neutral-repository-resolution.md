# Transport-neutral repository identity and machine-local acquisition

Date: 2026-09-10.
Status: separate future specification amendment requested during local-source
discussion; no normative schema or implementation change accepted yet.

## User requirement

A Skillfile or other repository declaration should identify the repository
using a supported URL or an explicit logical identity. Each machine should
resolve that declaration to a permitted, usable SSH or HTTPS endpoint using
the operator's authentication configuration. A repository should not require
different checked-in declarations solely because a workstation uses SSH and
CI uses HTTPS.

## Existing foundation and boundary

`protocol/core.md` section 6.1 already canonicalizes supported network Git
addresses to `host/path`, removing the transport, SSH username and trailing
`.git`. A canonical repository identity is therefore distinct from its
connection address. This is not yet a complete machine-local transport
selection contract.

The external-build transport in `profiles/manager.md` section 11 intentionally
clears inherited Git/SSH environment and configuration. Its closed executor,
SSH wrapper and credential broker cannot be replaced by unrestricted user
configuration inheritance. An operator-selected credential helper or platform
store can supply a host-bound secret through the existing broker contract.
Repository data must not select a credential helper, account or executable.

## Proposed model

Keep these concepts separate:

1. Repository identity: a stable logical identifier for audit, trust and locks.
2. Revision selection: the tag, branch or revision resolved to pinned content.
3. Machine policy: permitted endpoints, preferred transport, SSH connection
   details and references to operator-owned authentication providers.
4. Acquisition result: effective endpoint, verified identity and resolved
   commit, recorded without credential material.

Illustrative resolution, not a proposed final JSON schema:

```text
Input:
  https://github.com/acme/agent-kit.git
  git@github.com:acme/agent-kit.git
  an explicitly marked logical identity for github.com/acme/agent-kit

Stable identity:
  github.com/acme/agent-kit

Workstation policy:
  git@github.com:acme/agent-kit.git, operator-approved SSH authentication

CI policy:
  https://github.com/acme/agent-kit.git, operator-approved credential provider

Both acquire the same locked commit and run the same content checks.
```

An explicit logical-identity form must be syntactically distinguishable from
relative local paths. Endpoint aliases, custom SSH host aliases and mirrors
require an explicit operator-owned mapping to repository identity; textual
similarity or possession of an expected commit is not enough to infer trust
equivalence. Preserve case-sensitive repository path rules.

## Proposed resolution and failure contract

- Normalize supported input URLs before selecting transport. Define whether
  URL transport is a default hint and how a caller explicitly requires it.
- Use a deterministic ordered endpoint policy with bounded attempts. Do not
  discover credentials by broadly probing accounts or exporting Git/SSH state.
- Permit an alternate transport only when machine policy authorizes it and
  the failure class allows it, such as an unavailable endpoint or unavailable
  authentication method. Missing access remains an actionable failure.
- Fail closed on TLS or SSH host-key validation failure, untrusted identity
  mapping, audit rejection, revision mismatch or content integrity failure;
  these must not trigger a weaker fallback.
- Keep credentials, tokens and private key material out of manifests, locks,
  build environments, receipts and diagnostics. Shared lock identity and
  content pins must remain stable across machines. Keep machine connection
  policy and non-secret acquisition diagnostics outside portable identity.
- Apply the shared identity/endpoint model to each Git-backed package or
  context acquisition lane where admitted. Preserve any stricter lane-specific
  rules, particularly external build repositories and their credential broker.
- Local filesystem paths remain filesystem acquisition, even when the
  selected directory is a Git working tree.

## Decisions for the separate amendment

- Exact spelling of an explicit logical repository identity and reusable
  aliases, and its relationship to proposed Skillfile `sources`.
- Machine configuration location, transport preference and explicit pinning.
- Supported endpoint forms, enterprise hosts, ports, mirrors and SSH aliases;
  existing rejected forms must not become silently valid.
- Narrow supported translation from user Git/SSH configuration, including
  which `insteadOf`, credential-helper, host-alias and identity selections can
  be represented safely. No blanket inheritance of executable configuration.
- Failure classification, non-interactive authentication behavior and concise
  remediation when no authorized endpoint succeeds.
- Compatibility, lock/audit identity, endpoint provenance and tests for all
  acquisition lanes, including malicious remapping and failed host validation.

## Acceptance outline

One checked-in repository declaration resolves over SSH on one configured
machine and HTTPS on another without changing its logical identity or pinned
content. Existing URL declarations remain readable. Private credentials stay
machine-local. All attempts remain within explicit policy, existing trust
checks and the closed external-build execution boundary. No available access
produces an actionable diagnostic rather than silent identity substitution.

Evidence: `protocol/core.md` sections 6.1-6.3 and `profiles/manager.md`
section 11 at curator-spec commit
`d019f0e7179520b5c8dcde321c4fe51e04552f58`.
