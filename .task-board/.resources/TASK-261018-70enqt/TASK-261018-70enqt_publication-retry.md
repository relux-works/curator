# TASK-261018-70enqt: publication retry evidence


## Publication retry

2026-10-19, producer RUN-261018-77aa5f. No source, test or documentation
file was edited in the managed worktree in this retry. CHANGELOG.md and the
root README were not edited. The existing changelog delta belongs to the
previous candidate; this run neither removes nor republishes it as a new edit.

Last signed candidate located from the authoritative rev2 validation resource:
`da16449862e45250ab62168c0a1085109d768650`, tree
`5ea9873c866ef558a0f0bc07f67887b69ae06071`. Git signature verification reports
`G`. All 118 candidate-changed paths were compared directly against that
snapshot: **117 of 118 are byte-identical**. The only difference is the existing
`.research/codex162/TASK-261018-70enqt_gaps.md` r3 refresh-blocker appendix.
All six tracked candidate paths, including CHANGELOG.md, match the signed
snapshot (diff exit 0). There is no source/test/configuration refresh delta.
This resource adds only the publication-retry record; it does not alter the
worktree's research record or qualification artifacts.

Fresh authority resolved by `git ls-remote --symref origin HEAD` is `main` at
`b9caa20901ee859eec20e649940693033e6297d4`; the subsequent exact main fetch
agrees. CHANGELOG.md is the sole overlapping candidate path. Accordingly r4
requires refresh before publication. The production refresh entry point
refuses the current CR state before preparation: revision 2 remains `ready`.
Task status `development` does not grant a rework CR lifecycle transition.

| Command/check actually run | Real exit | Result |
| --- | ---: | --- |
| `task-board m 'set_status(TASK-261018-70enqt, status=development)'` | 0 | Task entered development; CR lifecycle unchanged. |
| `git ls-remote --symref origin HEAD` | 0 | Fresh main authority recorded above. |
| `git fetch origin main` | 0 | Remote-tracking main equals advertised authority. |
| `git show -s --format='%H %T %G? %s' da16449862e45250ab62168c0a1085109d768650` | 0 | Exact signed snapshot identified; signature G. |
| Python direct byte comparison of all 118 snapshot-changed paths | 0 | 117 identical; only prior r3 gaps appendix differs. |
| `git diff --exit-code da16449862e45250ab62168c0a1085109d768650 -- CHANGELOG.md cmd/curator/env_credential_marker_test.go internal/envprofile/credential_link_test.go internal/envprofile/credential_record_test.go internal/envregistry/envregistry.go internal/envregistry/envregistry_test.go` | 0 | All tracked candidate files match. |
| `task-board worktree refresh-candidate TASK-261018-70enqt` | **1** | `candidate refresh requires a rework revision; TASK-261018-70enqt is ready` |
| `git diff --exit-code -- README.md` | 0 | Root README unchanged. |
| `git diff --check` | 0 | Existing candidate whitespace valid. |

Evidence discrepancy: the attached rev2 validation log reports hosted run
39108873281 successful with command exit 0. It does not contain the curator-4
cleanup failure described by r4. This run does not infer another candidate or
claim a new hosted result from that mismatch. No full local suite, vendor
replay, build, lint or hosted gate ran in this publication retry: refresh was
refused and source/test/configuration identity is unchanged. Historical green
qualification evidence and all original negative/control bounds remain as
recorded above. No post-turn validation checklist items were added.

### Publication blocker and recovery

Constraint: a stale ready CR cannot be refreshed by its tracked producer.
The same production refusal was reproduced in r3 and now r4. No branch switch,
rebase, merge, commit, ledger edit, operator-identity bypass or control-root
file edit was attempted. The signed candidate and uncommitted work are preserved.

Required external input: the orchestrator must use the authorized CR lifecycle
to move revision 2 to a supported rework-admissible state, then resume the
producer for `refresh-candidate`. Recommended option: repair that admission,
resolve only the overlapping combination under the rc.7 ownership rule, and
publish through normal handoff with automatic exact-candidate validation.
Alternative: an operator-owned supported convergence before retry; any conflict
must respect the rc.7 prohibition on producer CHANGELOG/root README edits.
Manual Git or ledger rewriting would bypass ownership and is not a viable
producer workaround. If the supported lifecycle cannot express this transition,
its owning tool needs repair before publication. This is an ownership/lifecycle
blocker, not a qualification or runner failure.

Task-scoped logbook, 2026-10-19: publication retry preserved all candidate files,
verified the signed snapshot and fresh trunk overlap, and reproduced the
ready-CR refresh refusal (exit 1). No refreshed candidate or hosted-green claim
was made. Evidence was attached before blocking. Coverage remains **11 of 12 AC
rows driven**, with the existing preservation bound, brief surface-table gap,
out-of-contract rows, reviewer-test evidence and mutant table unchanged.

### authority-01.log

```text
ref: refs/heads/main	HEAD
b9caa20901ee859eec20e649940693033e6297d4	HEAD
```

### refresh-retry-01.log

```text
candidate refresh requires a rework revision; TASK-261018-70enqt is ready
```

### candidate-comparison-01.log

```text
candidate_commit=da16449862e45250ab62168c0a1085109d768650
base_commit=402fdbc9b77803cb2d6b11f4f1cfd83961e6897b
paths_compared=118
candidate_files_match=117
difference=.research/codex162/TASK-261018-70enqt_gaps.md
working_sha256=81d80d113d601a81bf90c3d9b2ac9cb7b9597cf8ed0997478ff43bc17f213f0d
```

### root-readme-02.log

```text
```

### diff-check-02.log

```text
```
