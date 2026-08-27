# TASK-260720-3ag6pi integrated protocol v6 verification

## Outcome

The accepted `curator-spec` protocol `1.0.0-rc.4` composite passes every
repository-supported validation, deterministic regeneration, lint, and release
gate required by the task. No product file, generated expectation, schema,
vector, implementation pin, review, claim, or release artifact was changed by
this verification task.

## Provenance and clean integration state

- Fetched `origin/main` and verified exact base
  `57c1f56846d221ecc55786bd3c2467ec32f11730`.
- Created task worktree
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3ag6pi/worktree`
  detached at that base.
- Imported the complete accepted product diff from the actual retained
  TASK-260720-q5oy3o worktree at
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-q5oy3o/curator-spec-worktree`.
  The board note's shorter `/worktree` path was stale; the accepted task result
  identifies the retained path above.
- A checksum-aware mirror comparison is empty apart from ignored metadata and
  timestamps. The normative binary fixtures were preserved byte-for-byte:
  `build-source.preimage.bin` is
  `27cdcac0734aa3e069e95a10341e89b118a07c60002516e7b401e95477f01332`;
  `toolchain.preimage.bin` is
  `baf7c5f3b9c3f1fae3da4c356381bf74442aa7f8f0b6fb2304c9c10833d6032e`.
- The real Git index has zero staged paths. No commit was created.

## Verification gates

- `make validate`: pass — 35 schemas, 189 manifest entries, 27 Python tests,
  and `go test ./tools/...`; no skips were reported.
- First clean `make regenerate`: pass — 190 files before and after, aggregate
  SHA-256
  `41aa37774478c26377455877ee79ef74f8cb5cf5562ea5b1501e5c94fe9c1fa0`,
  with no candidate-index diff.
- `make regenerate-check`: pass with no diff and the same aggregate digest.
- Second independent worktree regeneration: pass — the same 190-file digest,
  no candidate-index diff, and recursive byte comparison against the first
  regenerated tree is empty.
- `make release-check VERSION=1.0.0-rc.4`: pass at the exact base after all
  validation/tests and regeneration checks; release gate output is
  `release gate passed for 1.0.0-rc.4`.
- `gofmt` check, `go vet ./tools/...`, uncached `go test -count=1 ./tools/...`,
  and `git diff --check`: pass.

The accepted composite is intentionally uncommitted and the task forbids
staging. Disposable alternate indexes represented the candidate snapshot for
`regenerate-check` and `release-check`. A task-local Git wrapper isolated the
unit tests' temporary Git repositories from that outer index, and treated the
release gate's exact `git status --porcelain` query as clean only when the
worktree matched the candidate index and no unindexed file existed.
`PYTHONDONTWRITEBYTECODE=1` prevented `release_gate.py` from creating an
untracked `tools/__pycache__`. This changes no repository behavior and does not
invent a commit or release record. Earlier failed harness attempts are retained
as task-scoped diagnostics.

## Manifest and compatibility evidence

- Manifest SHA-256:
  `70cd274150d629f16ca04f2073a8eae5a0f3fff4584f905d807775040af21eae`.
- All 189 listed paths exactly equal the filesystem inventory and all 189
  listed SHA-256 values match their bytes.
- New schema-case filesystem/manifest/index counts agree: agent-skill-v6 17,
  csk-skill-v6 17, build-receipt-v1 16, install-marker-v2 14, and
  conformance-claim-v2 7.
- All 11 expected build-driver files and all 13 fixture files are inventoried.
  `build-drivers.json`, `manager-lifecycle.json`, and the schema-case index have
  correct manifest hashes.
- Agent-skill and csk-skill schemas 1 through 5, all of their cases,
  install-marker-v1 plus its cases, and conformance-claim-v1 plus its cases are
  byte-identical to `origin/main`. Their filtered schema-case index entries are
  identical. Claim v1 remains schema 1 / protocol `1.0.0-rc.3` with frozen
  schema SHA-256
  `c9f49460618ccc8b1d7d2dfaf760fc6ad3a53a870a6685a685ddc148d3c87b3f`.

## Coverage and safety

`TASK-260720-3ag6pi_coverage-matrix.md` maps every story acceptance criterion
and each structural, path, identity, toolchain, module/source, process, cache,
context, claim, transaction, concurrency, recovery, currentness, repair, GC,
and fail-closed rejection cluster to named executable evidence.

All 75 build-driver rejection cases are uniquely named and require
`artifact_executed: false` plus `reuse: false`. The accepted fixed driver uses
no shell and does not execute its output; cache-hit and dry-run cases run no
source-aware Go command. The generator and validator contain no process
execution API. Release-gate subprocesses are limited to Git and the
specification validator.

No implementation pin or review changed. `RELEASE.md` contains zero checked
release-evidence boxes and 25 explicitly unchecked prerequisites. No downstream
schema-6 manager release, claim-v2 result, cross-platform interoperability run,
candidate tag, signature, archive checksum, or provenance attestation is
claimed or created.

## Test-scope note

This task changes no product behavior, so adding a new repository test would
duplicate the accepted fail-closed inventory tests. The existing 27 Python
tests and Go generator suite are the executable tests for this verification
scope and passed both in `make validate` and the full release gate.

The board's structural validator still reports the same unrelated pre-existing
12 broken EPIC-260712 references and one orphan TASK-260713 resource recorded
by the story planning work. None belongs to this protocol story or was changed
by this task.
