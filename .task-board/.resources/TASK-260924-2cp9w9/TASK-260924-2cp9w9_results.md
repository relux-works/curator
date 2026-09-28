# TASK-260924-2cp9w9 results

## Scope

Settled the three assigned residual clauses in `protocol/skillfile-sources.md`
and `protocol/repository-transport.md`, added distinguishing schema and semantic
vectors, and recorded the changes in the Unreleased changelog. No namespace,
status, frozen v1, or release artifact was changed. No Curator implementation
code was in scope.

## Clause changes

### Registry evidence, revocation, and hash

**Before (excerpt):**

> Both `network-git` and legacy `configured-git` packages may record it only
> when existing registry rules establish the exact canonical repository, name,
> commit and context hash. A configured source path alone is not a registry
> identity. This summary is not authorization: signed records and current trust,
> revocation, freshness and assurance policy MUST still be checked when required.

**After (new normative text):**

> For a schema-2 network-Git member, a non-revoked record MUST NOT provide
> positive evidence or an attestation unless all four values match: canonical
> package name, canonical repository identity, locked commit, and registry
> artifact content hash. This is a conjunctive grant; matching only content, or
> only repository and commit, does not grant positive evidence. The compared
> registry `content_sha256` is the SHA-256 of the raw frozen package tree under
> registry §3 hashing rules. It is not the lock member's projected context
> `content_sha256`.
>
> For schema-2 network-Git members, revocation retains registry §3's broader
> artifact match: a verified `revoked` record matches when either its artifact
> content hash equals the raw frozen package-tree hash, or its canonical
> repository identity and commit both match. Registry §4 is deny-wins, so such a
> matching revocation MUST block the member under advisory as well as strict
> registry policy, even when its name or other grant dimensions do not match.
> Exact positive matching MUST NOT narrow this revocation check.

The advisory choice preserves the existing fail-closed revocation boundary;
positive authorization remains exact.

**Vectors:** `attested-network-current` accepts the exact tuple while the lock
projection differs; `attestation-evidence-wrong-name`, `-wrong-repository`,
`-wrong-commit`, and `-wrong-context` each isolate one mismatching grant field
under strict policy; `attestation-evidence-revoked` covers content-only match
under advisory; `attestation-evidence-revoked-identity-commit-advisory` covers
repository-plus-commit match with name and hash differing.

### Marker v5 local arm and `declared_tag`

**Before (excerpt):**

> Only receipt_schema_version changes to 3. External substitution shape, commit
> length, exact-tag provenance and all cross-field equality rules remain intact.
> An external record MUST NOT omit these fields in favor of a receipt hash alone.

The previous text did not expressly close the local `go-v1` arm at the raw JSON
boundary or state the value grammar for `declared_tag`.

**After (new normative text):**

> The local `go-v1` record is a closed raw JSON arm: no member defined only by
> the `go-repository-v1` arm may be present, including with JSON `null` (for
> example, `repository`, `declared_tag`, `substituted`, or `substitution`).
> Presence is invalid even when decoding would otherwise collapse `null` to an
> absent zero value. In a `go-repository-v1` record, `declared_tag` is optional;
> when present it MUST be a non-empty Git ref name, and null, empty, or
> ref-name-invalid values MUST be rejected.

**Vectors:** added isolated schema negatives for local `repository: null`,
`substituted: null`, and `substitution: null`, plus external `declared_tag: ""`
and `declared_tag: "release..candidate"`. Existing valid external records
retain a valid tag control.

### Refresh endpoint and scp alias port

**Before (excerpt):**

> Explicit refresh reruns all gates and atomically replaces the lock and marker
> only after success. Failure preserves prior state.

The source-policy rules resolved aliases for new endpoint attempts but did not
say that refresh of an existing checkout must use the current plan. Revision 2
also left scp spelling with an alias port without a rendering rule.

**After (new normative text):**

> For an explicit refresh of an existing network-Git checkout, the manager MUST
> resolve the current source-policy endpoint plan before network I/O and MUST
> send every refresh fetch to an endpoint connection target resolved by that
> plan. A checkout's persisted `remote.origin.url` is machine-private state; it
> MUST NOT select, replace, or bypass the current endpoint plan, even when the
> checkout is reused. When no policy entry applies, use the declaration's
> endpoint once under repository-transport revision 1; a logical declaration
> without an entry fails with `repository_endpoint_unavailable`. An unreadable
> or invalid policy, or failure of all permitted current endpoint attempts,
> MUST fail without changing the prior lock or installed state.
>
> An endpoint written in SCP-like form (`[user@]host:path`) has no URI port
> position. If its named alias supplies a port, the manager MUST fail with
> `repository_policy_invalid` before network I/O. Converting it to an SSH URI
> would require a `~/` remote-path convention that some SSH front ends do not
> accept, so the manager MUST NOT guess that rendering or risk changing the
> target path. An SSH URI endpoint MAY use an alias port; its connection target
> is the SSH URI with the resolved alias host and port, while the repository
> identity remains the entry key.

The absent-policy-entry behavior preserves revision 1 URL resolution; the
fail-closed rule applies to invalid/unreadable policy and exhausted current
attempts.

**Vectors:** `v2-refresh-current-endpoint-existing-checkout` gives an existing
checkout a stale stored origin and requires refresh to fetch from the current
resolved alias target. `v2-scp-alias-port-refused` requires zero attempts;
`v2-ssh-uri-alias-port` proves the refusal is specific to scp spelling and
preserves SSH URI alias-port rendering.

## Validation

| Check | Result |
|---|---|
| `python3 tools/validate.py` with repository dev requirements in the worktree venv | exit 0; 64 schemas and 1,169 vector files |
| Draft schema/vector inventory check | exit 0; 121/121 schema cases across 8/8 families; 98 semantic records parsed with unique IDs. Shape only; no manager execution. |
| Full `python3 -B -m unittest discover -s tools -p 'test_*.py'` attempt | exit 130 when interrupted at the ten-minute command bound; no test failure had been reported. The suite was rerun in bounded groups below. |
| `tools/test_validate.py`, bounded class groups | exit 0 across 7 groups; 533 tests passed |
| Other four `tools/test_*.py` modules, bounded group | exit 0; 78 tests passed |
| `go test ./tools/...` | exit 0; generator package compiled and tests passed |
| `go run ./tools/generate-vectors -root .` | exit 0 |
| `make regenerate-check` | exit 0; generator and frozen-artifact diff check both passed |
| `git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json` | exit 0; frozen generated artifacts unchanged |
| `git diff --check` | exit 0 |
| Lint configuration | No repository lint target or configured linter found; `git diff --check` passed |

The first system-Python validation attempt exited 1 because `jsonschema` was
unavailable. The declared `requirements-dev.txt` dependency was installed in
an ignored `.temp` virtual environment, and the validation and test reruns above
used that environment. The full unittest discovery command exceeded the
per-command bound, so its test modules were run in bounded groups; all 611 tests
passed in those groups.
