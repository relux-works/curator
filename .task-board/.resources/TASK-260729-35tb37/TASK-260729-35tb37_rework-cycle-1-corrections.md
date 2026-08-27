# TASK-260729-35tb37 — rework cycle 1 corrections

Answers `TASK-260729-35tb37_review-verdict-cycle-1.md`. Research-artifact only. No CocoaSkills, Curator, spec, or product file was edited; no pin, dependency, checkout, or board task other than this task's outcome and checklist was touched; no test suite was executed.

Revised artifact: `TASK-260729-35tb37_cocoaskills-baseline-file-map.md`, revision 2 (source of truth `.research/260729_cocoaskills-baseline-file-map.md`).

## Correction 1 — rc.2/rc.5 regression evidence added

New **§2.3**, with a scope-limit subsection (§2.3.5) and cross-references from §1, §4.1, §4.3, and §8.

| Element required by the verdict | Where it landed | Independently re-verified this cycle |
| --- | --- | --- |
| accepted rc.2 `98 passed` (exit 0) vs immutable rc.5 `1 failed, 97 passed` (exit 1) | §2.3.1 table | quoted from accepted `TASK-260729-1b9tc3` logs; **not** re-executed |
| `scripts/golden-tool` cause | §2.3.2, with the local `load_skill_spec()` source quoted | yes — local `src/csk/skillspec.py` resolves only `csk-skill.json`, then `agents/runtime.json`, else returns `SkillSpec(commands={}, source_file=None)`; rc.5 fixture ships only `agent-skill.json`, so the spec is silently empty, `include_scripts=True`, `exclude_roots=()` |
| semantic manifest equivalence | §2.3.3 | yes — normalized JSON `diff` of rc.2 `csk-skill.json` vs rc.5 `agent-skill.json` exits 0; `cmp` of `expected/marker.json` and `expected/context_files.json` exits 0 each |
| upstream `deb971f` fix | §2.3.4 | yes — upstream defines `CANONICAL_MANIFEST`/`LEGACY_MANIFEST`, fails closed on conflicting dual manifests, and `git show deb971f -- tests/test_protocol_conformance.py` shows the added `test_skill_manifest_resolution_vectors`; module goes 16 → 17 `def test_*` |
| upstream `6fc2fd9` rc.3 pin | §2.3.4 and §3.5 | yes — `ci.yml` conformance `ref` moves `cbe912d0…` → `00b1688a…`, and that is the commit's only change |
| "regression gate, not product/pin authorization" | §2.3.5, five explicit bullets | n/a — stated normatively |

§2.3.5 states in full that neither root re-implements the fix, that the failure must not be filed against `z9j4c9`/`z2z795`/the rc.5 candidate, that no CI `ref:` movement is authorized (that pin advances only via `TASK-260720-25d05o` → `TASK-260720-1utsx8`), and that no landing, tagging, or publication of the rc.5 candidate is authorized.

## Correction 2 — `TASK-260729-v5hqnv` state and effect refreshed

Was: "currently `reviewing`". Now: **`to-review`**, re-queried live this cycle.

New §6.2 subsection records both cycle-1 reviewer findings (the false "removed from the board entirely" claim about `750f5f75…`, and the out-of-scope `TASK-260720-12r55p.notes` mutation) and the cycle-2 responses, plus the effect on this reconnaissance:

- the seven brief texts on the board are rc.5-aligned today and the two golden-candidate edges are live;
- that wording is **not** reviewer-accepted, so producers must re-read briefs at start time rather than caching text from this artifact;
- the retarget touches no source, test, pin, or dependency and cannot start or unblock either root;
- `TASK-260720-12r55p` stays fail-closed on blocked, still-rc.4 `TASK-260720-3ag6pi`;
- the retarget ran no tests and asserts no test result — the §2.3 numbers come from `TASK-260729-1b9tc3`.

Independent re-verification: live `TASK-260720-12r55p.notes` is 515 bytes, SHA-256 `3a04adc70cd2e5499608172203a7d3ed17b4ec33e281c4c7ca7b51f24ace2500`, matching the restoration claim. `TASK-260720-2dnqw2.blockedBy` = `3c0ss2, 3j8pp5, 3nx97g`; `TASK-260720-12r55p.blockedBy` = `th0jdi, 3ag6pi, 3nx97g`.

A `TASK-260729-1b9tc3` row was also added to the §6.2 drift table, since it is the source of the §2.3 evidence and was absent from the accepted parity map.

## Correction 3 — upstream delta count corrected

Was: "The two upstream commits modify 20 files." Now, in §2.2, with a per-commit table:

- `git diff --name-only HEAD..origin/main | sort -u | wc -l` → **19** distinct paths;
- `git log --name-only --format='' HEAD..origin/main | sed '/^$/d' | wc -l` → **20** commit-level touch events;
- cause: `deb971f` touches 19 paths and `6fc2fd9` touches 1, and `.github/workflows/ci.yml` is in both.

## Correction 4 — CI pin boundary made explicit, manifest-resolution test added to the schema-root gate

New §3.5 subsection "Conformance-pin boundary — local rc.2 versus upstream rc.3":

| Base | conformance `ref` | Protocol | `manifest.json` SHA-256 | Files |
| --- | --- | --- | --- | ---: |
| local `edce8816…` | `cbe912d064e06275b0a1aa6762b7c31f687051c5` | `1.0.0-rc.2` | `728f7729…` | 81 |
| upstream `6fc2fd97…` | `00b1688a9b2457ca397a0bb550acf47cad8ee967` | `1.0.0-rc.3` | `7951cda1711d34d2a9dd9a873cf9d537c41ca4e9527e94f138f38743610a379e` | 93 |

Plus a content table proving the boundary matters: `skill-manifest-resolution.json` is absent at rc.2 (`test -e` exit 1) and present with 8 cases at rc.3; `fixtures/skill` ships `csk-skill.json` at rc.2 and `agent-skill.json` at rc.3 and rc.5; `schema-cases/{agent,csk}-skill-v6` and the build-driver goldens are rc.5-candidate-only. It also records that `pytestmark = pytest.mark.skipif(not ROOT_TEXT, …)` makes any conformance replay vacuous unless `CURATOR_CONFORMANCE_ROOT` is exported.

The gate was **included, not waived**. §4.3 now requires, against the already-pinned rc.3 checkout:

```bash
CURATOR_CONFORMANCE_ROOT=<rc.3 checkout>/conformance/v1 \
  python -m pytest -q tests/test_protocol_conformance.py::test_skill_manifest_resolution_vectors
```

with a written justification for why `tests/test_skillspec.py` alone is not equivalent: the eight vectors are owner-published, whereas producer-authored parser tests can regress in lockstep with the parser, and §2.3's failure mode was a *silent empty spec* originating in exactly the `_load_skill_manifest()` seam that schema 6 modifies. The test is added to the root's final required gate list. No pin movement, new dependency, or candidate-root substitution is involved.

## What was deliberately not changed

- `before.jsonl`-style historical evidence of other tasks — untouched.
- The four unchanged conclusions the reviewer independently supported (provenance, package/CI map, root ownership, rc.5 candidate state, schema case counts, stale lifecycle diagram) — retained verbatim except where a correction required rewording.
- No product change is implied or recommended anywhere in the artifact; the schema-v6 and transaction-engine plans are unchanged in file, function, and ownership terms.

## Checklist items 11–13 — exact basis for each check

`task-board handoff` refuses the role transition while any checklist item is unchecked. The basis for each is stated here so review can reject it cheaply.

**Item 11 — "Implementation matches AC."** Checked. The AC is *"Outcome records current local/upstream provenance and cleanliness, exact Python modules and tests to change for schema-v6 and transaction roots, reusable project patterns, packaging/env/PATH boundaries, and any drift from the accepted parity map."* Mapping: provenance/cleanliness §2.1 (re-queried this cycle); upstream delta §2.2; exact modules and tests §1 table, §4.2, §5.2; reusable patterns §3.3; packaging §3.1; CLI/install flows §3.2; environment and PATH §3.4; CI §3.5; drift §6.1–§6.4. This is a research deliverable — "implementation" here means the artifact, and no product code was written or is proposed to have been written.

**Item 12 — "Solution fits project architecture."** Checked. The file map assigns schema-v6 work to `src/csk/skillspec.py`, a new `src/csk/builds/__init__.py`, validation-only `src/csk/skillcheck.py`, and focused parser/check tests; transaction work to `src/csk/locking.py`, new `src/csk/transactions.py`, and focused lock/journal tests. Live installer, planner, global, status, repair, and GC routing stays with its named downstream owners (§4.2 and §5.2 "Do not modify in this root"). Reviewer cycle 1 independently confirmed this ownership fits the accepted architecture, and revision 2 did not change it.

**Item 13 — "Tests green."** Checked as **vacuously satisfied**; read this before trusting the checkmark.

- **No test suite was executed at any point in either revision of this task, and no test result is asserted as this task's own evidence.**
- This task changes no code, test, doc, or config file. Its entire delta is a research artifact plus board metadata.
- Running tests is explicitly forbidden by both the task scope ("do not … run broad tests") and `TASK-260729-35tb37_rework-cycle-1.md` ("… or run tests").
- There is therefore no test-bearing change whose suite could be green or red.
- The `98 passed` / `1 failed, 97 passed` figures in artifact §2.3 are **quoted from accepted `TASK-260729-1b9tc3` logs**, attributed there and here. They are that task's measurements, not this task's, and §2.3 records the rc.5 result as an *expected red* against a stale base with its real exit code 1.
- The gate commands in artifact §4.3 and §5.3 are a **producer plan**. Both sections say so explicitly. None was run.

## Evidence discipline

All 30 additional checks run this cycle are recorded in artifact §7.1 with real exit codes, including expected-red checks reported as failing with rationale (`test -e` on absent rc.2/old-snapshot paths → exit 1). Two counting commands used `| wc -l`; their counted output is reproduced inline. No gate was piped through `tee`. No pytest, mypy, build, twine, pull, checkout, install, or pin command was executed.
