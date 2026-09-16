# TASK-260910-1o9x1f results — machine-owned repository endpoint policy

Producer: developer. Shell for all commands below: bash; exit codes captured
without pipes (`> file; echo exit=$?`), so each is the gate's real status.

## Revisions

- Spec: curator-spec main `871d11bcdfd240a6260d0722503bdd1642a8fce8`
  (verified-at). Policy contract files unchanged since `a4fcaf02`:
  `protocol/repository-transport.md` (§1–§3 = revision 1),
  `schemas/draft-sources-v1/source-policy-v1.schema.json`,
  `conformance/draft-sources-v1/schema-cases/source-policy-v1/*` (5 files),
  `conformance/draft-sources-v1/semantic-cases.json`
  (`fallback-*`, `pinned-auth`, `endpoint-identity-mismatch`),
  `docs/skillfile-sources.md` (machine policy example).
- Code: Story worktree `task-board/story/STORY-260910-1bhj0g`, uncommitted
  handoff snapshot. Builds on the landed Skillfile v2 parser
  (`internal/manifest`: schema-2 sources); parser untouched.

## What was implemented (revision-1 only; no v2 behaviour)

- `internal/config/sourcepolicy.go` (new): `LoadSourcePolicy` /
  `ParseSourcePolicy` (source-policy schema 1, closed), `SourcePolicyPath`
  (beside manager config), `ResolveRepositoryEndpoints` (declaration to
  ordered `Resolution`, zero network I/O). Diagnostics carry the exact
  classes `repository_policy_invalid` / `repository_endpoint_unavailable`
  (transport §2) and `source_selection_invalid` (sources §5) for malformed
  declarations.
- `internal/identity/draft_sources.go` (+20 lines): `DraftCanonicalKey`
  (already-exact canonical host/path via the core §6.1 canonicalizer).
- `internal/gitcred/gitcred.go` (+12 lines): `ValidProvider` (opaque
  operator provider refs are closed identifiers; never commands/paths).
- `docs/draft-source-policy.md` (new): boundary note in the style of
  `docs/draft-source-expansion.md`.
- `internal/config/testdata/draft-sources-v1/source-policy-v1/`: all 5
  published schema cases copied verbatim + pin README.

## Acceptance rows → committed tests (all reach production entry points)

- Source-policy schema (closed v1, schema_version const, fallback enum,
  root_inputs shape): `TestDraftPolicyPublishedSchemaCases` (5 vendored
  cases through `ParseSourcePolicy`), `TestParseSourcePolicyDocumentShape`,
  `TestParseSourcePolicyRootInputs`, `TestParseSourcePolicyRejectsRevision2Closed`
  (schema 2, `aliases`, `alias`, `mirror_of` all rejected).
- Exact canonical identity: `TestDraftCanonicalKey` (identity unit),
  `TestParseSourcePolicyEntryKeysAreExactCanonical` (load),
  `TestPolicyKeyAndEndpointIdentityAgree` (identity/lane fixed-point pact).
- One/two distinct endpoints + provider refs: `TestParseSourcePolicyEndpoints`
  (mismatch incl. the published `endpoint-identity-mismatch` refusal,
  repeated URL with distinct auth, 0/3 endpoints, foreign grammar,
  command/path-shaped auth).
- Order and pin: `TestParseSourcePolicyPinAndFallback` (pin must equal a
  listed URL exactly; same-identity-but-unlisted pins rejected),
  `TestResolveRepositoryEndpoints` (list order kept when the hint is the
  second endpoint; pin yields one attempt with fallback forced to none).
- URL without policy attempts declared URL once: `TestResolveRepositoryEndpoints`
  (nil policy and entry-less policy; single attempt, empty provider =
  existing lane policy, fallback none).
- Logical identity without entry fails: same test (nil + present policy →
  `repository_endpoint_unavailable` naming the identity).
- Invalid/unreadable policy fails before network: `TestLoadSourcePolicy`
  (malformed file, directory-as-path → `repository_policy_invalid`;
  missing file → absent `(nil, nil)`, never an error); this layer performs
  no network I/O by construction (no net/exec use in the new code path).
- Package inputs cannot introduce providers or commands: `Resolve` takes
  only `(policy, gitURL, repository)` — no provider/command parameter exists;
  auth values pass through byte-exact from machine policy only
  (`TestResolveRepositoryEndpoints`); command/path-shaped refs are rejected
  at load (`TestParseSourcePolicyEndpoints`, `TestValidProviderAdmitsOpaqueIdentifiersOnly`).

## Validation (real exit codes)

- `go test -count=1 ./internal/config/ ./internal/identity/ ./internal/gitcred/` → exit 0
  (`ok config`, `ok identity`, `ok gitcred`; legacy suites included, no skips).
- `go vet ./internal/config/ ./internal/identity/ ./internal/gitcred/` → exit 0.
- `gofmt -l internal/config internal/identity internal/gitcred` → no output.
- `golangci-lint run ./internal/config/ ./internal/identity/ ./internal/gitcred/` → exit 0, `0 issues.`
  (First run was exit 1 with 5 findings — 3 errcheck, 1 revive, 1 staticcheck —
  all fixed in the diff, none suppressed.)
- Full module suite deliberately NOT run (host rule: narrow packages only;
  remote gate runs once at handoff).

## Mutant attack (narrowing; bytes restored, proven by `git diff`)

- M1 pin exact-equality → same-identity acceptance: KILLED.
  `TestParseSourcePolicyPinAndFallback` FAIL (accepted unlisted
  `https://example.org/kit` pin).
- M2 distinct-URL check → whole-object uniqueness: KILLED, exit 1.
  `TestParseSourcePolicyEndpoints` FAIL (accepted repeated URL with
  distinct auth).
- M3 `ValidProvider` → non-empty-only: KILLED, exit 1 at both levels
  (`TestValidProviderAdmitsOpaqueIdentifiersOnly`: 20 failures;
  `TestParseSourcePolicyEndpoints` FAIL via the production loader).
  (First M3 shape at the call site broke the build via an unused import;
  recast at the predicate, which is the real gate.)
- M4 `DraftCanonicalKey` equality → case-folded: KILLED by
  `TestDraftCanonicalKey` (exit 1) and by
  `TestPolicyKeyAndEndpointIdentityAgree` (exit 1). Note:
  `TestParseSourcePolicyEntryKeysAreExactCanonical` SURVIVED M4 because
  the endpoint-identity-mismatch gate still rejects (`Example.org/kit`
  key vs lowercase endpoint identity) — defense in depth, with the
  boundary itself pinned by the unit + agreement tests.
- Post-attack `git diff` shows only the intended hunks (2 modified files,
  +32 lines) plus the 6 new files; full narrow suites re-run green after
  restore (exit 0).

## Bounds (not covered here, by scope split)

- Failure classification and second-attempt execution (`fallback-*`,
  `pinned-auth` semantic cases beyond plan shape): sibling backlog task
  TASK-260910-5nrmtt consumes `Resolution` (attempts + effective fallback).
- Provider-ref → credential-material resolution: same sibling; this layer
  validates refs as opaque identifiers and threads them through untouched.
- `root_inputs` existence/coverage and overlap-at-acquisition: loader
  validates shape + intra-list overlap; existence is acquisition-time.
- Revision 2 (ports/mirrors/aliases): rejected closed; hook is the
  `schema_version` discriminator only.
- No live credential export, no runtime-home writes, no network use.

## Checklist note

- Item 8 (logbook): no LOGBOOK.md edit per host rules (control-root and
  LOGBOOK writes are forbidden to producers); the M4 second-gate
  observation above is the only finding and is recorded here.

## Revision 2 — rework 1 (reviewer R1: explicit null for optional fields)

Reviewer verdict `TASK-260910-1o9x1f_review-verdict-rev1.md` (CHANGES_REQUESTED):
`ParseSourcePolicy` treated explicit JSON `null` for the optional fields
`root_inputs` and `pin` as absence, accepting two invalid documents. Fixed;
no other verdict items required work.

Fix (`internal/config/sourcepolicy.go`, 2 lines):
- `root_inputs` presence check no longer skips null: a present-but-null value
  now flows into `parseRootInputs`, which rejects the non-object with
  `repository_policy_invalid` before any network I/O.
- `pin` presence check no longer skips null: a present-but-null value now
  flows into the string assertion, which rejects it with
  `repository_policy_invalid` (a null pin no longer silently becomes
  unpinned list/fallback behavior).

Tests (committed, all through the production entry point):
- New `TestLoadSourcePolicyRejectsNullOptionalFields`: both exact verdict
  documents through `LoadSourcePolicy` with real temporary files, each
  expecting `repository_policy_invalid`; controls (omitted optionals and
  valid `root_inputs` + valid `pin`) still load clean via the same path.
- Parser tables extended with the null shapes: `pin:null` in
  `TestParseSourcePolicyPinAndFallback`, `root_inputs:null` in
  `TestParseSourcePolicyRootInputs`.

Validation, revision 2 (shell: bash; each exit code captured without pipes):
- `go test -count=1 ./internal/config/ -run TestLoadSourcePolicyRejectsNullOptionalFields -v` → exit 0 (4/4 subtests pass).
- `go test -count=1 ./internal/config/ ./internal/identity/ ./internal/gitcred/` → exit 0 (`ok` x3, legacy suites included).
- `go vet ./internal/config/ ./internal/identity/ ./internal/gitcred/` → exit 0.
- `gofmt -l internal/config internal/identity internal/gitcred` → exit 0, no output.
- `golangci-lint run ./internal/config/ ./internal/identity/ ./internal/gitcred/` → exit 0, `0 issues.`
- Full module suite deliberately NOT run (host rule: narrow packages only).

Mutant attack, revision 2 (narrowing; bytes restored, proven by `cmp` identical):
- M-A root_inputs guard reverted only (`present && rawInputs != nil`):
  new test exit 1, ONLY the `null_root_inputs` subtest FAILs. KILLED.
- M-B pin guard reverted only (`present && rawPin != nil`):
  new test exit 1, ONLY the `null_pin` subtest FAILs. KILLED.
- 2/2 killed; each subtest independently pins its own branch.
