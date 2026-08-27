# TASK-260824-1n98b3 review — advance-consumer-pins

Verdict: **ACCEPTED**. Every AC and DoD line is met, and each claim below was
re-derived from the repos and the runs, not read off the producer notes.

## Release identity — both pins name the same immutable commit

`git ls-remote --tags relux-works/curator-spec v1.0.0-rc.9`:

- tag object `b6796644…`, peeled `v1.0.0-rc.9^{}` = `0ed5c691e9208eea52f21db2fc05e226ce3516fd`

Both consumers pin that peeled commit, not the tag and not a branch, so neither
pin can be re-pointed under the consumer.

`conformance/v1/manifest.json` digest at `0ed5c691` (released) and at
`6001dc33` (the qualified schema-8 candidate declaration) both hash to
`803918bf8672f76cf990985e51db213b826674cd5bb54fbf47731b8404b44403` — the
release publishes the qualified candidate's bytes, which is what makes the
"digest deliberately not repeated in ci.yml" note in both repos correct rather
than a weakened gate.

## curator half

| Item | Evidence |
| --- | --- |
| Pin | `ci.yml:44 SPEC_PIN: 0ed5c691…`, the only 40-hex pin literal in the file; `00b1688a` (rc.3) appears nowhere on main |
| Landing | PR 38 → merge commit `272b203` (two parents — merge-commit convention) |
| Post-merge main | run `32770177743`, conclusion **success**, 12 jobs: Test/Race/Gate self-test on ubuntu+macos+windows, Lint, Naming gate, Interop conformance gate; `Candidate suite` correctly **skipped** (non-default by construction) |

Pulled the uploaded `test-evidence-*` artifacts instead of trusting the green:

| Runner | served | deferred | excluded |
| --- | ---: | ---: | ---: |
| ubuntu-latest | 42 | 0 | 1 (`internal/godriver`, excluded on linux by the root's own `conformance-claim-v3-qualification.json`) |
| macos-latest | 43 | 0 | 0 |
| windows-latest | 43 | 0 | 0 |

All nine packages that were deferred under rc.3 — buildcache, buildsource,
install, marker, moduleroots, scopes, scriptpolicy, skillspec, whitelist — are
present in `plan-served.txt` on all three runners. `plan-deferred.txt` is empty
on all three. The pin advance did the substantive thing it was for.

## cocoaskills half

| Item | Evidence |
| --- | --- |
| Pin | `ci.yml:53 RELEASED_SUITE_PIN: 0ed5c691…` |
| Landing | PR 46 → `e8cc5f4`, single parent on `f94ad35` — rebase-merge as required |
| Post-merge main | run `32775205440`, conclusion **success**, 30 jobs |
| Decisive Windows shards | all six protocol shards green post-merge: p00-contract-and-registry, p01-lifecycle-cached-baseline, p02/p03/p04/p05-lifecycle-sabotage-a–d, plus `ubuntu-full` and `macos-full` |

Every fail-closed identity binding moved in the same commit, verified in the diff:

- `test_protocol_conformance.py` — manifest digest → `sha256:803918bf…`,
  `EXPECTED_CANDIDATE_PROTOCOL_VERSION` → `1.0.0-rc.9`, release record →
  `release/1.0.0-rc.9.json` with `claim_v3` → `claim_v5`, in-scope inventory
  24/24/102 → 28/28/110, qualification-vector version now derived from the
  constant instead of re-literalled
- `test_build_metadata.py` — same digest, and its comment advanced too
- `test_ci_workflow.py` — pin literal
- `test_protocol_shards.py` — 1045 → 1053, p00 565 → 573
- shard contract `.research/TASK-260803-2ol7ok_*` — `protocol_commit`,
  `baseline_count`, ordered baseline, `audited_files_sha256`, and the verifier's
  `EXPECTED_PROTOCOL` plus a newly named `EXPECTED_BASELINE_COUNT` replacing a
  bare 1045

Machine-checked rather than taken on faith: all four `audited_files_sha256`
entries recompute exactly against `origin/main`. The collection delta is exactly
the 8 new `invalid-v8-*` rows under agent-skill-v6 and csk-skill-v6, all joining
the existing `function::test_rc6_generated_schema_case_is_consumed` cluster in
shard p00 — no cluster added, removed or re-membered.

Keeping the `rc6_`-prefixed test names is a deliberate, documented call: renaming
would rewrite every row of the ordered node-id baseline and the per-node
isolation classification for no behavioural gain. Correct trade, and it is now
written down in the module header instead of being folklore.

## Finding — stale narration, follow-up not rework

`0ed5c691` publishes `schema-cases/agent-skill-v8`, `csk-skill-v8` and
`install-marker-v4` (verified by `git ls-tree` at that commit). Two places in
cocoaskills still say the opposite:

- `.github/ci/candidate-artifacts.tsv` header: "The default protocol lanes run
  against the released suite pin, which publishes no schema-8 family, so the
  schema-8 consumer defers there and says so."
- `tests/test_schema8_candidate_conformance.py` module docstring: "a root that
  publishes none of it -- the released suite pin, for instance -- defers this
  consumer"

Both were true under rc.6 and are false under rc.9. **No behavioural impact**:
the served/deferred partition is derived from the root at runtime, the schema-8
consumer is invoked only from the candidate lane, and no test asserts the
released root defers it (`test_candidate_consumption.py` uses synthetic roots),
which is why CI is green. It is comment-only drift — but it is the same class
the producer correctly fixed in `test_build_metadata.py`, and it is the sentence
the next person advancing this pin will read. Worth a follow-up edit.

Secondary nit, pre-existing and untouched: curator `ci.yml:35-37` still says
promoting SPEC_PIN "is owned by TASK-260720-38l1sy, after TASK-260720-25d05o
qualifies the release" — this task promoted it, so that ownership line is stale
inside the very hunk that was edited.

Carried forward from the producer, not acted on: cocoaskills repo variable
`CSK_E2E_CURATOR_SPEC_SHA=432eb2ee` is dead config from 2026-08-04. It is inert
(`test_ci_workflow.py` asserts the name appears nowhere in the workflow, and
`test_candidate_suite.py` proves a `vars` expression is rejected as non-immutable)
and deleting repo-level config is outward-facing and outside this AC. Owner call.

## Not accepted as evidence

The producer's local full `tests/test_protocol_conformance.py` run hit its own
1500s bound (exit 124) at ~90% with zero failures observed. It was honestly
declared as a gap rather than dressed up as a pass, and it is not needed: the
post-merge `ubuntu-full` and `macos-full` shards cover that module end to end
and are green on the merge commit.

## Reviewer scope note

The results artifact asked the reviewer to check "the two pin values and the two
post-merge conclusions; nothing else". Reviewed the full delta anyway — that is
how the schema-8 narration drift surfaced.

## Commit ownership

Nothing is left uncommitted in either consumer repo: both halves are merged and
green on main. No `commit_ack` supplied — reviewer archetype.

## Follow-up this task unblocks (out of its AC, but now triggered)

curator-spec's `.github/workflows/implementations.yml` still carries the block
comment stating that the Python manager's `tests/test_protocol_conformance.py`
step "returns to this job when its released suite pin advances to rc.9 -- the
step the landing order schedules after rc.9 is published". **That condition is
now met.** As of `origin/main` the step is still absent, the manager pin there
is still `3ecca1db`, and the two vector families LOGBOOK entry 2249 flagged as
having no consumer at all — `vectors/registry-client.json` and
`vectors/skill-manifest-resolution.json` — are still unenrolled.

This is correctly outside this task: its AC scopes only the two consumer pins,
and the prior reviewer already recorded the curator-spec side as unowned. But
completing this task is precisely the trigger, so it needs an owner now:
re-add the step, advance the manager pin, and add ledger rows for both families
so a gate keeps them honest instead of a comment.

## Working-tree state left by this review

`LOGBOOK.md` in curator is modified and uncommitted — the 2026-08-25 0119 entry
recording this verdict, the unblocked curator-spec follow-up and the schema-8
narration drift. Reviewer archetype does not commit; it belongs to the
commit-owning mover.
