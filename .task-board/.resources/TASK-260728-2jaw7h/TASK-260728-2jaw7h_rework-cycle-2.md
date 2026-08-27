# TASK-260728-2jaw7h — rework cycle 2

Developer handoff evidence for the two defects of review cycle 1. Status: ready
for review.

Execution base: `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-2jaw7h/curator-spec-worktree`,
unchanged from cycle 1. Nothing was staged, committed, published, pinned, or
advanced. The accepted `TASK-260728-2kp3tv` predecessor worktree stayed read
only and was used only as the comparison reference.

## Defect 1 — the frozen rc.5 boundary

### What was wrong

The corpus was generated into `conformance/v1`. The suite manifest is a walk of
that directory, and `release/1.0.0-rc.5.json` pins the SHA-256 of the manifest,
so eleven new corpus entries moved a released identity. Regeneration rewrote
both the suite and the document that pins it in one pass, so the pair agreed
afterwards and the release gate — which compares them against each other —
passed. A self-consistent rewrite proved nothing.

### The fix

**A second suite root, `conformance/next`.** The five toolchain vector files and
the six schema-case directories for `agent-skill-v8`, `csk-skill-v8`,
`skill-build-v2`, `toolchain-registry-v1`, `toolchain-guidance-catalog-v1` and
`toolchain-diagnostic-v1` are generated there, with their own
`schema-cases/index.json` and `manifest.json`. Nothing about their content
changed: all five vector files and four of the six case directories are
byte-identical to cycle 1 (the other two are discussed under "One latent defect"
below). `conformance/v1` is byte-identical to the accepted predecessor.

The candidate manifest carries no `protocol_version`, because naming one would
mint one. It records `released: false`, `candidate_against: 1.0.0-rc.5`, and
`release_pin_owner: TASK-260728-251p01`, and `validate.py` fails if any of those
three drift or if a version appears.

**An authored frozen-identity record, `release/frozen.json`.** It holds, per
released version, the SHA-256 of the release document, of the suite manifest it
pins, and of that suite's schema-case index, plus the suite root the document
must name. It is authored and never generated, which is the whole point:
regeneration cannot move the expectation along with the artifact.

Three gates enforce it, at three different moments:

| Gate | Where | What it stops |
|---|---|---|
| `assertFrozenReleaseIdentity` | end of `tools/generate-vectors` | a regeneration that would leave a rewritten release behind — it fails instead of writing one |
| `validate_frozen_releases` | `tools/validate.py`, every run | a rewritten release reaching a green local validation |
| `validate_frozen_releases` | `tools/release_gate.py`, before a tag | a rewritten release reaching a release |

Each also cross-checks the release document's own `suite_root` and both pins
against the frozen record, so a document that kept its accepted bytes while
pointing somewhere else fails too. A candidate that publishes a release document
must itself be recorded, so the version being cut is never the one that escapes
comparison.

`Makefile`, `.github/workflows/ci.yml` and `.github/workflows/release.yml`
extend the determinism diff to `conformance/next`.

### Proof it stops the exact cycle-1 defect

In an isolated scratch git repository, the manifest was extended and the release
document repinned to the new digest — exactly what regeneration does, so the two
agree with each other:

```
release gate exit=1
release gate failed: frozen release 1.0.0-rc.5 was rewritten:
release/1.0.0-rc.5.json is sha256:b52fce81…,
release/frozen.json requires sha256:b32ee9d3…
```

Reverting the probe returns the gate to exit 0. The generator refuses the same
rewrite at source: planting one file under `conformance/v1/vectors` and running
`go run ./tools/generate-vectors -root .` exits **1** with
`generation rewrote frozen release 1.0.0-rc.5`.

### Frozen bytes, measured against the accepted predecessor

| Artifact | SHA-256 | Result |
|---|---|---|
| `release/1.0.0-rc.5.json` | `b32ee9d35fc2fce0539d0b1c4b15f9c5239115c22fe54a72401f5da6d6646441` | **match** |
| `conformance/v1/manifest.json` | `9ba9b8ecf6f06cafda1425aed3539dee8d12af43dabd36f803f4f420dbb1dacf` | **match** |
| `conformance/v1/schema-cases/index.json` | `2faa2baaadca30b3eebe3b9248260efcfac30e92cef4fa209bf37f3f23efd4f0` | **match** |
| whole `conformance/v1` tree | — | `diff -rq` exit **0**, byte-identical |

The schema-case index entry set is unchanged from cycle 1 and only partitioned:
540 entries before, 376 released plus 164 candidate after.

## Defect 2 — the stale reference contract

`docs/compiled-build-toolchain-requirements.md` announced two completions in its
header and contradicted both in its body.

**`source_ref.surface`.** The body now lists `manifest`, `descriptor`,
`registry`, `source_metadata`, matching `protocol/core.md` and the
`toolchainSourceRefV1` enum, and explains why the fourth token exists: the
baseline is a contributing source of the intersection, so an unsatisfiable
conflict has to be able to name it as a `fragments` element. It also states what
`registry` is *not* — the `conflict` bound lists use the literal
`registry_baseline`, because a bound achieved by the baseline is achieved by the
registry as a whole rather than at a location inside it. That distinction was
checked against the schema and the corpus rather than asserted.

**Value classifiers.** Section 3.1.1 now carries the `matches` legend
(`absence` | `value`), the rule that at most one class matches absence and it is
declared first, the requirement that the catch-all matches a value, and the
statement that the forbidden-before-compared precedence is quantified over the
`value` classes only. Both classifier tables gained a `matches` column with the
shipped class names, and the `toolchain` table says plainly why an absence class
at position 1 ahead of two `forbidden` classes is not a precedence inversion.

**The drift guard.** `validate_reference_document` in `tools/validate.py` makes
the reference's own "a disagreement with `protocol/core.md` is a defect in this
file" executable. It:

- pulls the surface token list out of both the reference's and `core.md`'s own
  sentences and requires each to equal `common.schema.json`'s enum;
- requires the section 3.1.1 legend to equal
  `toolchain-registry-v1.schema.json`'s `valueClass.matches` enum, and that
  schema to still require the member;
- parses both classifier tables and requires each row's (class name, `matches`,
  disposition) triple, in order, to equal the shipped registry entry's classes
  for the `go.mod` `go` and `toolchain` fields, with contiguous numbering so the
  prose's class references still resolve.

## One latent defect found while fixing the first

`agent-skill-v8/valid-non-primary-identifier-is-a-runtime-code.json` and its
`csk-skill-v8` twin were shipping in cycle 1's `conformance/v1` and in its
manifest, but were named by no index entry. The generator writes cases and never
prunes, so a renamed case leaves its file behind; the manifest walk then hashes
an orphan nothing validates into a pinned digest. Removing the directories and
regenerating cleared them.

`validate_schemas` now requires the case files present under each schema-case
root to be exactly the set the index names, in both directions. Planting an
orphan under either root fails validation with the file named.

## Gate evidence — each command run standalone, real exit codes

In the task worktree:

| Command | Exit | Result |
|---|---|---|
| `python tools/validate.py` | **0** | 48 schemas, 592 vector files (422 released, 170 candidate) |
| `python -B -m unittest discover -s tools -p 'test_*.py'` | **0** | Ran 169 tests, OK (141 before this cycle) |
| `go test ./tools/...` | **0** | ok generate-vectors |
| `go vet ./tools/...` | **0** | clean |
| `gofmt -l tools` | **0** | no output |
| `python -m compileall -q tools` | **0** | clean, external `PYTHONPYCACHEPREFIX`, no `tools/__pycache__` left |
| `git diff --check` | **0** | clean |
| `go run ./tools/generate-vectors -root .` | **0** | |

Determinism, two passes against a pre-run snapshot:

| Probe | Exit |
|---|---|
| generate run 1, `diff -r conformance` vs snapshot | **0** |
| generate run 1, `diff release/1.0.0-rc.5.json` and `release/frozen.json` | **0** |
| generate run 2, `diff -r conformance` vs snapshot | **0** |
| generate run 2, `diff release/1.0.0-rc.5.json` and `release/frozen.json` | **0** |

Clean-checkout probe, an isolated scratch git repository under
`.temp/TASK-260728-2jaw7h/clean-probe-cycle2`, never the real repository:

| Command | Exit |
|---|---|
| `make validate` | **0** |
| `make regenerate-check` (run 1) | **0** |
| `make regenerate-check` (run 2) | **0** |
| `make release-check VERSION=1.0.0-rc.5` | **0** — `release gate passed for 1.0.0-rc.5 at 0b14f202…` |

Boundary probe, against Go 1.25.1 (`/usr/local/go/bin/go`) and Go 1.25.5
(`~/.goenv/versions/1.25.5/bin/go`):

| Run | Exit |
|---|---|
| `make boundary-probe` | **0** — 2 toolchains, 16 go-directive and 13 toolchain-directive cases, 0 failures |
| `make boundary-probe-controls` | **0** — all five controls failed as required |

**Reported truthfully as failing, because it is:** `git diff --exit-code --
conformance/v1 conformance/next release/1.0.0-rc.5.json` **inside the task
worktree** exits **1**, and `make regenerate-check` therefore exits **2**
(make's own code). It diffs against `57c1f56` while the whole candidate is
uncommitted. Determinism is established instead by the pre-run snapshot
comparison and the clean git-init probe above, exactly as in every predecessor
candidate.

Python gates use the task-local `validation-venv`; ambient `python3` lacks
`jsonschema`.

## Negative probes — each fired, each reverted

| Probe | Exit | Diagnostic |
|---|---|---|
| a file planted under `conformance/v1/vectors`, then regenerated | 1 | `generation rewrote frozen release 1.0.0-rc.5` |
| suite manifest extended and the release document repinned to agree | 1 | `frozen release 1.0.0-rc.5 was rewritten` (release gate, clean scratch repo) |
| `registry` dropped from the reference's surface list | 1 | `states source_ref surfaces [...], but common.schema.json closes them at [...]` |
| `registry` dropped from `protocol/core.md` | 1 | same, naming `protocol/core.md` (unit test) |
| the absence class relabelled `value` | 1 | `documents the 'go' classifier as [('absent', 'value', ...)]` |
| a `forbidden` class reordered after a `compared` one | 1 | `documents the 'toolchain' classifier as [...]` |
| a `protocol_version` added to the candidate manifest | 1 | `names a protocol version; the surface it belongs to is unminted and reserved to TASK-260728-251p01` |
| an orphan case file planted under either schema-case root | 1 | `carries case files that no index entry names, so nothing validates them` |
| restore, re-run `validate.py` | 0 | 48 schemas, 592 vector files |

## Tests added this cycle

28 new tests, 141 → 169.

- `FrozenReleaseIdentityTests` (7): the shipped record matches; a rewritten
  manifest fails although the document agrees; a rewritten manifest fails when
  the document is left byte-stable, so the manifest is pinned in its own right;
  a rewritten index fails; a document naming another suite root fails; a missing
  artifact fails; an empty record list fails.
- `CandidateSuiteBoundaryTests` (8): the split holds in both directions, every
  candidate schema has cases under the candidate root and none under the
  released one, a candidate case indexed under the released root fails, a
  version or a release claim on the candidate manifest fails, a toolchain vector
  inside the pinned suite fails, an unindexed case file fails.
- `ReferenceDocumentDriftTests` (8): the shipped reference agrees; dropping
  `registry` from either document fails; relabelling the absence class fails;
  removing a classifier row fails; dropping a legend token fails; a registry
  schema that stops requiring `matches` fails.
- `FrozenReleaseGateTests` (5, `test_release_gate.py`): the shipped repository
  passes; a rewritten manifest fails the release; a regenerated pair that agrees
  with itself still fails; a release document without a frozen record fails; a
  stable version with no release document is not required to be recorded.
- `tools/generate-vectors/frozen_release_test.go` (3): candidate cases stay out
  of the pinned suite and every candidate schema has a home under it; the
  candidate manifest mints no version; the generator's own guard rejects a
  self-consistent rewrite.

## Files changed this cycle

Versus the cycle-1 candidate, and nothing else:

New: `release/frozen.json`, `conformance/next/` (11 corpus entries plus two
generated indexes), `tools/generate-vectors/frozen_release_test.go`.

Modified: `.github/workflows/ci.yml`, `.github/workflows/release.yml`,
`CHANGELOG.md`, `COMPATIBILITY.md`, `Makefile`, `README.md`,
`conformance/README.md`, `docs/compiled-build-toolchain-requirements.md`,
`schemas/v1/README.md`, `tools/generate-vectors/main.go`,
`tools/release_gate.py`, `tools/test_release_gate.py`, `tools/test_validate.py`,
`tools/validate.py`.

Restored to the accepted predecessor's exact bytes:
`conformance/v1/manifest.json`, `conformance/v1/schema-cases/index.json`,
`release/1.0.0-rc.5.json`, and the six schema-case directories and five vector
files removed from `conformance/v1`.

No normative rule was weakened, no control removed, no scope broadened.
`protocol/core.md`, `profiles/manager.md`, `SECURITY.md`, every schema, and
`tools/toolchain_gate.py` are untouched this cycle.

## Reviewer focus

1. Whether `conformance/next` is the right shape for an unreleased candidate
   root, or whether the corpus should have been deferred entirely to
   `TASK-260728-251p01` rather than shipped unpinned.
2. Whether the frozen record belongs at `release/frozen.json` covering all
   released versions, or per-release beside each document.
3. Whether requiring a candidate that publishes a release document to be
   recorded in `release/frozen.json` is the right coupling for a future mint —
   it makes authoring the record a deliberate, reviewable step, at the cost of a
   two-step generate-then-author flow.
4. The four cycle-1 reviewer-focus items that were not touched this cycle remain
   open as stated in `TASK-260728-2jaw7h_results.md`: the `id` closure over
   `{go, kotlin, rust, swift}`, the `prerelease-track` unclassifiable token,
   `expected_metadata_sources` on a reserved registry entry, and the candidate
   still saying `1.0.0-rc.5` while shipping schema 8.
