# TASK-260824-3ojxne — reviewer verdict: ACCEPTED

Reviewed independently against `relux-works/curator-spec` main, not against the
implementer's summary. Every claim below was re-derived locally from git objects,
the GitHub API, and a fresh worktree at the landed commit.

## Verdict

**accepted** — the atomic landing satisfies the AC and the DoD. One
carry-forward finding is recorded in §5; it is a follow-up scope gap, not a
defect in the delivered change, and it does not block acceptance.

## 1. Identity and atomicity

| fact | verified value | how |
| --- | --- | --- |
| `curator-spec` main HEAD | `0ed5c691e9208eea52f21db2fc05e226ce3516fd` | `git rev-parse origin/main` |
| squash parent | `09f0423a` (single parent — true squash) | `git log -1 --format=%P` |
| squash tree | `ee1ac91de4c3472630c07d4d9def4042a89ec9b6` | `git rev-parse 0ed5c69^{tree}` |
| PR 29 branch head tree | `ee1ac91de4c3472630c07d4d9def4042a89ec9b6` — **identical** | `git rev-parse daa1cf3^{tree}` |
| candidate ancestry | `6001dc3` and `09f0423` are both ancestors of `daa1cf3` | `git merge-base --is-ancestor` |
| commit signature | `verified: true`, reason `valid` | GitHub commits API |

The bytes and the pins are in the **same** squash commit: `daa1cf3` ("Advance
the implementation pins with the schema-8 bytes") sits inside PR 29, and PR 29
landed as one commit. There is no interval of `main` carrying schema 8 with
non-consuming pins. DoD item 1 holds.

## 2. Candidate content merged intact

`git diff --stat 6001dc3 0ed5c69` touches 9 files. Seven are the landing's own
additions (coverage ledger, tool, tool tests, workflow, CHANGELOG,
COMPATIBILITY, README); the other two — `profiles/manager.md` and
`protocol/core.md` — carry exactly PR 28's prose, which the candidate predates.

- Not one `schemas/`, `conformance/`, or `release/` file from the candidate was
  altered during landing.
- Every added line of PR 28 (`517a130..09f0423`, 20 insertions across the two
  files) was located verbatim in the blobs at `0ed5c69`. Zero missing.
- Suite manifest at `0ed5c69` hashes to
  `sha256:803918bf8672f76cf990985e51db213b826674cd5bb54fbf47731b8404b44403`,
  **byte-identical** to the candidate at `6001dc3`. Both implementations were
  therefore qualified against exactly the bytes that landed.

## 3. Pins point at the two qualified commits

| impl | pinned ref on main | qualification | independently checked |
| --- | --- | --- | --- |
| Go (`relux-works/curator`) | `a3abcf3468b4854904313295672eef6f7d8826fd` | dispatch `32689488293` | run `head_sha` **is** `a3abcf34`, conclusion `success`; commit is `relux-works/curator` origin/main HEAD and is the merge of PR 37 |
| Python (`ivanopcode/cocoaskills`) | `3ecca1dba9f8831e1617b7466c17ecc8a2957d3f` | run `32756144649` | run `head_sha` is `ae1b6f2a` (PR 43 branch head), conclusion `success`. `ae1b6f2a` and `3ecca1db` have the **same tree** `fae69431` and the **same single parent** `66068a1b`, so the qualified bytes and the pinned bytes are the same bytes. `3ecca1db` is PR 43's merge commit on main (main is 1 commit ahead, a docs-only change). |
| Registry | `d690bea6fab1c8e6392e05d3a3cdfcf1168bc914` | unchanged | schema 8 adds no registry surface — confirmed, `schemas/` delta is v8 skill schemas, marker v4, claim v5, and additive `$defs` only |

The pin-run mismatch on the cocoaskills side (run recorded against the PR head
rather than the merge commit) is **not** a gap: tree and parent equality make
the two commits byte-equivalent. Worth stating explicitly because the evidence
artifact reported the run as if it had executed on `3ecca1db`.

## 4. AC clauses, one by one

**"every lane green"** — all 8 required contexts on branch protection
(`Formatting`, `Links`, `Specification` ×3, `Implementations` ×3) are `success`
on the PR head `daa1cf3`, last one completing `18:33:48Z`, merge at
`18:34:11Z`. Green **pre**-merge, verified from the check-runs API rather than
from the merge button. Post-merge `main` is green on all 9 including
`Release target provenance` (which is `skipped` on the PR head and runs on
push — the evidence artifact's "runs only off pull requests" is backwards, but
the substantive claim that it passed is correct).

**"rc.8 untouched"** — `release/1.0.0-rc.8.json` blob OID is `e05e4e92` at both
`09f0423` and `0ed5c69`; the file hashes to
`sha256:293f101d10665061aa049efa72141f9e3c5d608bbde300e882f6e3e095e31ede`.
`release/1.0.0-rc.9.json` records that exact digest under `historical_release`
with `immutable: true`, and records the landed suite manifest `803918bf...`
under both `candidate_protocol_pin` and `downstream_consumption`.

**"double regeneration proven"** — re-run from scratch in a clean worktree at
`0ed5c69`:

| pass | `go run ./tools/generate-vectors -root .` | `git diff --exit-code` | tree after |
| --- | ---: | ---: | --- |
| 1 | 0 | 0 | `ee1ac91d` |
| 2 | 0 | 0 | `ee1ac91d` |

The post-generation tree equals the landed main tree exactly — a stronger
result than the artifact's reported digest `effb543a`, which does not
correspond to any tree in this landing and appears to be a stale figure. The
determinism claim itself is confirmed.

**Other gates re-run locally at `0ed5c69`:** `python tools/validate.py` → 53
schemas / 691 vector files, exit 0. `python -B -m unittest discover -s tools -p
'test_*.py'` → 134 tests, OK. `go test ./tools/...` → ok. `gofmt -l tools` →
empty. `git diff --check` → clean. All match the reported figures.

## 5. The coverage contract, and the one thing it does not cover

The landing adds `.github/ci/implementation-coverage.tsv` (18 rows: 7 Go, 11
manager), `tools/implementation_coverage.py`, and 36 unit tests. This is wider
than the task text asked for, but it is what makes DoD item 1 *checkable* —
without it "the pins consume the bytes" is an assertion, not a gate. It fits
the repository's architecture: the ledger is owned by the specification, so an
implementation that renames or drops a schema-8 consumer fails *this* repo's
gate, not only its own.

Reproduced independently:

- `families --root conformance/v1` → `18 declared claim(s) upheld`, exit 0.
- `families` against the pre-schema-8 rc.8 root from `09f0423`
  (`sha256:d14e3a16...`) → exit 1, naming **17** rows that read an unpublished
  family. Matches the reported negative proof exactly.
- Fail-closed behaviour is unit-tested at the parser level: renamed case,
  failed case, skipped case, skipped *sub*test, module-level skip, empty run,
  missing stream, absent declared row. All green.
- Real CI evidence in `TASK-260824-3ojxne_ci-coverage-gate.log` shows
  `18 / 7 / 11 declared claim(s) upheld` on all three runners.

### Finding — two pre-schema-8 vector families lost their only consumer

The landing removes `tests/test_protocol_conformance.py` from the
Implementations job. **The stated justification is factually correct**: at pin
`3ecca1db` that module asserts
`digest == "sha256:12e58b82..."` (protocol `1.0.0-rc.6`) and fails collection
against any other root, and the previous pin `6fc2fd97` performed no root
authentication at all (verified: no manifest digest assertion anywhere in the
file at that pin). It genuinely could not be pointed at this repository's
moving root.

But that module ran 17 vector-driven test functions against this repo's root,
and the replacement does not cover all of them. Cross-checking every family it
read against what still runs in the job:

| family it read | still consumed in this job? |
| --- | --- |
| canonical-valid / canonical-invalid / identifiers / locale-selectors / manager-config / source-identities | yes — Go `internal/interop`, and the registry suite |
| portable-paths | yes — Go `internal/skillspec`, and the registry suite |
| closures | yes — Go `internal/closure` |
| **registry-client** | **no** — no Go source or test file at pin `a3abcf34` references it; the registry pin's conformance module reads `registry-service.json`, not `registry-client.json` |
| **skill-manifest-resolution** | **no** — no `.go` file at pin `a3abcf34` references it; not read by the registry module either |

So `conformance/v1/vectors/registry-client.json` and
`conformance/v1/vectors/skill-manifest-resolution.json` are now published by
`main` with no pinned implementation observed reading them — the exact shape of
false green this landing's own README and CHANGELOG exist to condemn. Both are
still exercised in cocoaskills' own CI against its rc.6 released-suite pin, so
the vectors are not unexercised in absolute terms; what is missing is any proof
in *this* repository that a pin consumes them against *this* moving root.

Why this does not block acceptance: it is bounded to two pre-schema-8 families,
it is disclosed in three places (workflow block comment, board notes, evidence
artifact), the removal was forced by the pin advance rather than chosen, and it
is scheduled to end at landing-order step 9. The AC and DoD for *this* task
concern the atomic landing, which is correct.

**Carry-forward for the orchestrator:** the restoration is a `curator-spec`
workflow change (re-adding the step), and no board item currently owns it.
`TASK-260824-1n98b3` scopes only advancing cocoaskills' `RELEASED_SUITE_PIN`
and curator's `SPEC_PIN`. Add the curator-spec-side restoration to that task's
scope, or enrol it separately, or those two families stay unguarded here
indefinitely. Consider also adding ledger rows for them once the module is
back, so the gate — not a comment — keeps them honest.

## 6. Documentation

`CHANGELOG.md` records the coverage contract under Added and the pin advance
under Changed, naming both commits and the suite manifest. `COMPATIBILITY.md`
adds the two rules the prose left implicit; both were checked against the
schema bytes rather than taken on faith:
`schemas/v1/common.schema.json` `$defs.scriptCommandV8` carries
`dependentRequired: {execution_policy: [interpreter], interpreter:
[execution_policy]}` with neither in `required` — so absence means declared-only
and a half-declaration is invalid, exactly as written. Marker v1–v3 byte
stability holds: `git diff 09f0423 0ed5c69` over the v1/v2/v3 marker schemas and
schema-cases is empty, and `common.schema.json` is pure addition at one
insertion point. `README.md` points at the ledger.

## Checks run for this review

| command | result |
| --- | --- |
| `git rev-parse` / `merge-base` / `diff` over `curator-spec` | as tabulated above |
| GitHub API: PR 29, branch protection, check-runs on `daa1cf3` and `0ed5c69` | 8/8 required green pre-merge |
| GitHub API: curator run 32689488293, cocoaskills run 32756144649, PR 43 trees | both SUCCESS, pin bytes equal qualified bytes |
| `python tools/validate.py` @0ed5c69 | 0 — 53 schemas, 691 vector files |
| `python -m unittest discover -s tools` @0ed5c69 | 0 — 134 tests |
| `go test ./tools/...`, `gofmt -l tools`, `git diff --check` @0ed5c69 | 0 / empty / 0 |
| double `generate-vectors` + `diff --exit-code` @0ed5c69 | 0/0 both passes, tree stable at `ee1ac91d` |
| `implementation_coverage.py families` positive + rc.8 negative | exit 0 (18 upheld) / exit 1 (17 named) |

Verification worktree: `curator-spec/.temp/TASK-260824-3ojxne/verify` (removed
after the run).
