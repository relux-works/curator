# Repository endpoint policy for Skillfile schema 2

Skillfile schema-2 project resolution uses the accepted `repository-transport`
revisions 1 and 2 policy reader and endpoint planner by default. Schema-1
Skillfiles keep their existing meaning and are not migrated.

`config.LoadSourcePolicy` reads the operator-owned `source-policy.json`
beside the manager configuration; `config.ParseSourcePolicy` validates
one document. Schema 1 loads with revision-1 semantics byte-identically
to the pre-revision-2 loader (ports, `mirror_of`, `alias`, and the
`aliases` table are rejected there as unknown fields or lane-grammar
violations). Schema 2 (`schema_version: 2`) loads as the additive
superset of §4: port-bearing endpoint and pin URLs, per-endpoint
`mirror_of` attestation, and the operator host-alias table.

Entries are keyed by exact canonical identity without normalization; each lists one or two
distinct endpoints, an opaque operator provider reference per endpoint
(`gitcred.ValidProvider`, never a command or path), and a closed
fallback mode. `pin` selects exactly one listed URL by exact string
equality — including any port — and forces no fallback.

Canonical host/path is the only portable identity (§5). Ports, mirrors,
and aliases are machine-policy endpoint properties carried per attempt
(`config.Attempt`: `MirrorOf`, `Alias`, `ResolvedHost`,
`ResolvedPort`, `HasExplicitPort`) and never enter identity, the lock,
receipts, markers, or allowlist matching. A port never changes receipt
or cache identity beyond the existing `https`/`ssh` transport enum.
Endpoint URLs admit an explicit port only in URI form
(`https://host[:port]/path`, `ssh://[user@]host[:port]/path`, decimal
`1`–`65535`, no leading zeros); scp-like spellings carry no port. The
port-stripped path of every endpoint must equal the key path; only the
host may differ, and only with `mirror_of` equal to the key exactly. A
differing resolved host without attestation fails
`repository_mirror_undeclared`; an `alias` naming no table entry fails
`repository_alias_unknown`; every other misuse (mismatch or spurious
attestation, mirror URL combined with an alias, embedded alias host,
chained alias, double port, authentication mismatch, pin mismatch, bad
port or alias grammar) fails `repository_policy_invalid` — all before
any network I/O, with zero attempts and no fallback.

`config.ResolveRepositoryEndpoints` plans attempts for one declaration
without network I/O: a URL without an entry attempts the declared URL
once with no policy provider; a logical identity without an entry fails
`repository_endpoint_unavailable`; an invalid or unreadable policy fails
`repository_policy_invalid` before any acquisition. Package declarations
stay canonical and logical and never name aliases.

This layer plans only. Project source acquisition applies the bounded
attempt rules and failure classification from the transport contract. The
strict external-build lane has a separate opt-in gate; its execution and
provenance details are in `docs/draft-transport-resolution.md`.
