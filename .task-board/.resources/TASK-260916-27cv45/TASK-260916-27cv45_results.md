# TASK-260916-27cv45 results — policy schema 2 and endpoint identity

Role: developer (implementer). Worktree: `.temp/STORY-260916-v58b5y/worktree`, branch `task-board/story/STORY-260916-v58b5y`, uncommitted.
Spec: curator-spec `8ba9c23`, `protocol/repository-transport.md` §§4–7. Sibling leaf TASK-260916-hxr6qv (§6 attempt bounds, two new failure classes at resolution, §7 provenance/secrets) intentionally NOT implemented here.

## What changed

- `internal/config/sourcepolicy.go` — revision-2 reader:
  - Accepts `schema_version` 1 (new `parseSourcePolicyV1`, logic identical to the old loader) and 2 (`parseSourcePolicyV2`). Unknown versions fail `repository_policy_invalid` before I/O.
  - Schema 2: port-bearing endpoint/pin URLs (URI forms only, decimal 1–65535, no leading zeros; scp-like `:` stays path), `mirror_of` attestation, operator `aliases` table (`{host, port?, authentication}`, lowercase-only names/hosts, integer port, single substitution, auth equality).
  - §5 identity core (`checkEndpointSemantics`, shared by load and plan): port-stripped path must equal key path; URL host equal to an alias key refused (compared lowercased — DNS is case-insensitive — fail-closed); mirror-URL+alias refused; `mirror_of` present iff resolved host differs and always equal to the key; alias lookup failures → `repository_alias_unknown`, unattested mirrors → `repository_mirror_undeclared`, all other misuse → `repository_policy_invalid`.
  - `Resolution.Identity` is always the entry key; per-`Attempt` carries `MirrorOf`, `Alias`, `ResolvedHost`, `ResolvedPort` (0 = default), `HasExplicitPort` for the sibling leaf.
  - New codes `repository_mirror_undeclared`, `repository_alias_unknown`.
- `internal/config/sourcepolicy_test.go` — `TestParseSourcePolicyRejectsRevision2Closed` replaced by `TestParseSourcePolicyRevision2Boundary` (v2 accepted, v1 still rejects v2 members closed, incl. ports under v1).
- `internal/config/sourcepolicy_v2_test.go` (new) — corpus runner, schema-1 golden, 5 positives, 34 refusal rows with exact classes, hand-built planning violations.
- `internal/config/testdata/draft-sources-v1/source-policy-v2/` (new) — all 13 published schema cases verbatim from the spec commit.
- `docs/draft-source-policy.md` — revision-2 boundary doc.

## Verification (all observed this session, `set -o pipefail` shells excluded; plain `go` exit codes)

- `go test ./internal/config/ -count=1` → ok exit 0 (full package, incl. 13-file v2 corpus + pre-existing v1 corpus still 5 files).
- Narrow: `go test ./internal/config/ -run 'TestDraftPolicy|TestParseSourcePolicy|TestSchema1Golden|TestResolve|TestLoadSourcePolicy|TestPolicyKey'` → ok exit 0.
- Consumers: `go test ./internal/install/ -run 'TestDraftTransport|TestAcquireDraft|TestDraftPlan'` → ok exit 0; `go test ./internal/buildrepo/ -run 'TestParse|TestCanonical|TestTransportPlan|TestValidateTransport'` → ok exit 0.
- `go build ./...` → exit 0. `go vet ./internal/config/` → exit 0. `gofmt -l internal/config/` → clean. `golangci-lint run internal/config/...` → `0 issues` exit 0.
- Mutants (narrowed v2 suite, each reverted after; `diff -q` confirmed restore): M1 flipped mirror predicate → exit 1 (19 failing); M2 dropped embedded-alias refusal → exit 1 (2); M3 dropped alias-auth check → exit 1 (2); M4 allowed leading-zero ports → exit 1 (4); M5 collapsed `repository_mirror_undeclared` into `repository_policy_invalid` → exit 1 (4). All killed, no survivors.
- Remote gate (`scripts/remote-gate.sh`, pushes + polls GitHub CI, exceeds the single-shell bound and runs exactly once at handoff per campaign rules): NOT run manually; green is established by handoff/landing.

## Coverage notes (§5 refusal rows → loader entry)

`repository_mirror_undeclared`: undeclared mirror URL; alias-to-other-host without `mirror_of` (both at `ParseSourcePolicy` and at `ResolveRepositoryEndpoints` for hand-built policies). `repository_alias_unknown`: dangling alias (empty/absent/populated table). `repository_policy_invalid`: mirror_of mismatch, spurious mirror_of (plain + same-host alias), mirror URL + alias, embedded alias host, auth mismatch, chained alias, double port, pin-port mismatch, port 0/65536/70000/leading-zero/empty/non-numeric, scp colon-as-path, mirror path mismatch, HTTPS userinfo, uppercase alias name/target/field, string/zero alias port, alias missing auth/host/extra member, null mirror_of/aliases, non-canonical mirror_of, unknown top-level, extra endpoint member, duplicate port-bearing URLs. Schema-corpus negatives (9 files) additionally pin the grammar. Sibling-owned and therefore not loader-tested: `v2-external-build-port-refused`, `v2-external-build-alias-refused` (`build_repository_identity_invalid` at the strict lane), `v2-user-ssh-alias-ignored` / `v2-user-insteadof-ignored` (no user-config input exists at the loader; lane keeps clean config with no insteadOf import).

## Decisions / anomalies

- Embedded-alias comparison uses the lowercased URL host (fail-closed under DNS case-insensitivity); error attribution only, accept/reject identical either way since canonicalization lowercases before the identity check.
- `8443.0`-style alias ports decode as integral floats and are accepted, matching JSON Schema 2020-12 integer semantics.
- No LOGBOOK.md edit: worktree rules forbid control-root/LOGBOOK writes; findings above are the record.
- One self-authored expectation was wrong on first run (uppercase alias *field* is malformed → `repository_policy_invalid`, not `alias_unknown`); fixed the test, not the code. No production change from test feedback.
