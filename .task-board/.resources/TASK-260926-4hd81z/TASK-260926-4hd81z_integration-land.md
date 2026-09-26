# TASK-260926-4hd81z integration-land (bound developer run)

Revision 1 is ACCEPTED (reviewer verdict `TASK-260926-4hd81z_review-verdict-rev1.md`).
Board status observed `integrating`; left untouched. No `worktree integrate`,
no `handoff`, no status write was performed by this run — the synchronous
landing transaction is the runner's step.

## Landing preconditions confirmed

- Worktree delta is exactly one uncommitted file:
  `internal/scriptpolicy/conformance_test.go` (40 insertions, 6 deletions).
  Blob indices `a858071b..ec1727e5` match
  `TASK-260926-4hd81z_change-request_rev1.patch`.
- No CHANGELOG.md or LOGBOOK.md edit (`git diff --name-only` shows neither).
- No commit on the Story branch; change is uncommitted, ready for the
  handoff snapshot.
- Suite roots materialized fresh in $TMPDIR (never in the worktree) from
  full 40-hex revisions fetched from curator-spec origin:
  pinned `dcc7f015e2d97edf2d52928afb6fd79ec8129e8b` (vector label
  `1.0.0-rc.9`), candidate `f6bd748c59e015125b475428269ccdd367420d30`
  (vector label `1.0.0-rc.13`).

## Evidence rerun by this run (zsh, standalone processes, real exit codes)

- `env CURATOR_CONFORMANCE_ROOT=<pinned>/conformance/v1 go test
  ./internal/scriptpolicy -count=1` — exit 0.
- `env CURATOR_CONFORMANCE_ROOT=<candidate>/conformance/v1 go test
  ./internal/scriptpolicy -count=1` — exit 0.
- `TestScriptExecutionPolicyIdentityMatchesTheSuite` with `-v` on each root —
  PASS on both (driven, not skipped; the unset-root skip path was not taken).
- rc.14 mutant root (candidate tree with `protocol_version` rewritten to
  `1.0.0-rc.14`, valid JSON):
  `TestScriptExecutionPolicyIdentityMatchesTheSuite` — FAIL, real exit 1,
  with `want one of the unchanged script-worker-v1 labels
  [1.0.0-rc.9, 1.0.0-rc.13]`. Any-version mutant killed.
- `go vet ./internal/scriptpolicy` — exit 0.
- `go build ./internal/scriptpolicy` — exit 0.
- `gofmt -l internal/scriptpolicy` — empty, exit 0; `git diff --check` —
  exit 0.
- Bare `go test ./internal/scriptpolicy -count=1` (no root) also exit 0, but
  the suite identity test SKIPs there by design; the two rooted runs above
  are the binding evidence.

## Scope note

One file: the protocol-version check accepts exactly
`{1.0.0-rc.9, 1.0.0-rc.13}` with a comment recording that rc.13 is a
label-only change over the unchanged script-worker-v1 identity;
`schema_version`, `execution_policy`, and interpreter-set bindings unchanged.
New closed-set table test covers rc.9 / rc.13 accept and rc.14 reject.

## CHANGELOG entry (for release prep)

Accept the rc.13 suite label for the unchanged script-worker-v1
execution-policy identity while retaining a closed protocol-version set.
