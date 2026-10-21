# TASK-261021-vb3z6f — revision 2

## Delta and delivery

F1 is addressed in a new signed commit on the existing branch. CIP-0023 now
requires two distinct protected bindings: the CIP-0002 original-home session
lease, retained through confirmed termination and authority/state reconciliation,
and the Curator capture lease, released only after verified capture and durable
root handoff or an authoritative safe abort. Capture release never releases the
session lease. Busy updates, preparation/transfer ordering, independent unknown
liveness recovery, cold resume and both GC roots are explicit.

- Branch: `exact-plan-final-form`
- Parent: `0c1f830fc4006299f463f731af2310ec5d9f8eb1`
- Head: `cf1241cbb828180e5308a8b02f2671ebedabcedd`
- Signature: `G`, verified before push and again on the normal validation clone.
- Published by a normal fast-forward push; remote feature head verified equal.
- CIP-0002 candidate: `f202f41ba4718b82cdfefac28c8dfaa9af3a7de5` on
  `cip02-ready-for-acceptance`, fetched for this revision and confirmed unchanged
  by remote readback after validation.

One tracked file changed: `cips/CIP-0023-exact-plan-hosted-launches.md`,
112 insertions and 38 deletions. No executable behavior, accepted contract,
schema, vector or tool was changed. The proposal remains Draft. Board status
was not changed; no PR or default-branch push was made.

## Validation and its limits

All three `make validate` recipe stages passed at the exact handover head in a
disposable normal-`.git` clone. Per the standing shell-call limit, the Python
unittest discovery was partitioned into 31 bounded shell executions, with five
release-gate tests or at most 100 other tests per group. `make -n validate`
recorded the canonical recipes. **No single combined `make validate` process
or combined exit 0 is claimed.** The unchanged stages completed as follows:

| Check | Exit | Claim supported |
| --- | ---: | --- |
| `python3 tools/validate.py` | 0 | 74 schemas and 1294 vector files satisfy repository schema/vector/link checks. Runtime qualification remains unverified. |
| Exact Makefile unittest discovery, all 31 groups | 0 each | All 879 existing tests passed; original bodies and claim comments unchanged, no skipped-test summary. Per-group claims and exact test-ID manifest are below. |
| `check-groups.py` | 0 | Partition matches every discovered ID exactly once in discovery order; omission and duplication control plants fail for the intended coverage defect. |
| `go test ./tools/...` | 0 | The generate-vectors package's existing assertions pass (20.527s). |
| Draft checker | 0 | Required template metadata/sections, unique Draft index, no private path/seat text, and actual repository local Markdown targets pass. Valid link control passes; broken-link plant fails with the specific intended diagnostic. Remote links/anchors are not covered by this checker. |
| Lease crosswalk | Review completed | Every lease-bearing sentence, including summary, references and planned tests, agrees with the pinned CIP-0002 clauses or explicitly proposes a separate gated Curator companion. Candidate section headings for the two pinned remote anchors were checked locally. |

Platform/toolchain: macOS arm64, Python 3.14.7, jsonschema 4.25.1,
Go 1.25.5, GNU Make 3.81. Every validation-clone test command unset `GIT_DIR`
and `GIT_WORK_TREE`, used `TMPDIR=/tmp/cw-dev1`, and disabled Python bytecode.
Go used the workspace-local build cache. The full suite ran on exactly
`cf1241cbb828180e5308a8b02f2671ebedabcedd`. The clone was verified normal,
clean and at that head, then deleted with guarded non-force removal.

No native runtime launch, adapter qualification or implementation of the new
lease contract is claimed. As requested for this documentation-only delta,
four new test-plan rows specify executable follow-up acceptance requirements:

- Distinct leases: completed capture releases only capture; detached session
  retains the original-home lease; reconciled terminal state allows release.
  Mutants alias both refs, make session leasing optional or release both together.
- Busy updates: prepared/active/unknown session leasing still blocks repair,
  profile switching and skill/CLI updates after capture release, with no writes;
  after all leases end, a serialized/rechecked idle publication is allowed.
  Mutants equate snapshot completion with idle or omit the publication recheck.
- Crash/handoff: safe reconciled pre-start failure and proven capture handoff
  permit their respective releases; uncertain liveness retains corresponding
  roots and forbids duplicate start. Mutants use caller exit/TTL/PID or drop roots.
- GC: capture pins produced roots before admission and session pins them after
  capture release; collection requires absence of both and other lifetime roots.
  Mutants root only capture, root only ready sessions, or collect at capture release.

These are planned positive/negative controls and narrowing mutants, not tests
that were executed against a product implementation. The existing lifetime/GC
row also retains the valid explicit-forget/unreferenced-terminal collection control.

## Lease consistency crosswalk

CIP-0002 candidate: `f202f41ba4718b82cdfefac28c8dfaa9af3a7de5`, fetched from `origin/cip02-ready-for-acceptance` for this revision. Candidate source retained in `cip02-candidate.txt` beside this file. These are proposed contracts, not adopted runtime behavior.

Every lease-bearing sentence in CIP-0023 is covered below, including summaries, references, implementation rows and planned tests. Section references remain stable across wrapping changes.

| CIP-0023 sentences | CIP-0002 candidate counterpart | Finding |
| --- | --- | --- |
| Summary; current-state pinned dependency; specification changes; implementation plan rows 1, 2, 5; owner question 4 | Stable homes, leases and login; Lease state machine; Launch-route binding, exact-plan paragraph | Original-home session lease remains mandatory through confirmed termination and authority/state reconciliation. No capture-only replacement. |
| Section 6, first paragraph and closed inputs.project, protected bindings/digests | Closed project-lease/1 binding near line 316; exact-plan binding near line 572 | Session ref/hash identifies the unchanged closed candidate binding. Its lifecycle is separate from the immutable digest. Curator capture ref/hash identifies a distinct proposed closed companion; it does not add members to project-lease/1. |
| Section 6, preparation/capture acquisition | State machine lines 454-457; acquisition before collection near line 247 | Prepared session lease precedes fragment publication, view collection and capture; capture lease precedes materialization. Both use lifecycle serialization. |
| Section 6, final admission, durable transfer and roots | Active before provider start, transfer before launching caller exits; launch-route exact-process identity | Both bindings rechecked. Session owner/root handoff is durable before start, and uncertainty forbids start/duplicate start. New capture roots are a Curator companion requirement, not an existing candidate record field. |
| Section 6, capture completion, abort, unknown capture and release | Candidate has no separate capture lease; retained original-home session rule remains binding | Curator may release only its capture pin after verified completion plus durable root handoff, or authoritative safe abort. Interrupted/unknown capture retains its own resources until recovery. Capture release never releases session lease. |
| Section 6, session termination, failed preparation, unknown recovery, independent facts | State machine lines 457-464; descendants/unknown liveness near line 247; exact-plan reconciliation near line 579 | Manager/session lifecycle owner releases only after authoritative terminal/no-child facts and reconciled authority/state. Unknown session retains original home/artifacts. No PID/TTL/caller-exit shortcut. Independent capture reconciliation cannot establish session termination. |
| Section 6, busy updates and stable-home role | One composition while any lease exists, repair/profile/skill/CLI changes wait/refuse project_view_busy near line 245; transactional recheck lines 463-464 | Prepared/active/unknown session remains busy after capture release. All leases must end; serialized publication recheck prevents acquisition races. Session lease protects ownership/state even though content reads use the snapshot. |
| Section 7, resume | Recorded composition/current authority and no fresh-launch fallback, near lines 247 and 462-463 | Live/unknown generation retains original binding. Next-generation preparation/transfer must follow owning candidate contract; release is not silently reversed and current-home recapture is not allowed. |
| Section 7, GC roots and unconsumed preparations | No collection of leased homes/compositions/artifacts near line 247; unknown retains state lines 459-461 | Both distinct lease roots and other session/resume/outbox roots are independently honored. Capture pins before admission; session pins after capture release. Termination alone cannot collect independently rooted resumable content. Expiry alone releases neither lease. |
| Test-plan project capture, distinct leases, busy updates, crash/handoff and GC rows | Same acquisition, retention, busy/refusal, unknown and collection clauses above | Positive terminal/idle/safe collection controls accompany forbidden early release, busy writes and unknown-liveness collection; each row names narrowing mutants. These are future executable requirements, not runtime test results. |
| Existing CIP-0016 candidate home-lease/exclusive-profile reference in Current state | Candidate exact-plan/delegated route distinction near lines 563-568 | Remains explicitly unaccepted and separate from CIP-0002's mandatory selected-home lease. |

No remaining lease disagreement was found against this pinned candidate. The candidate does not yet specify the separate Curator capture record or its serialized root accounting; publication of that owning contract remains an explicit activation gate, not an assumed implementation.

Regression witness: the prior revision's sentence allowing early release of a capture-only stable-home lease failed the candidate's exact-plan requirement: “The original Werkstatt remains leased until the session's authority/state ownership is reconciled on termination.” The revised section 6 explicitly keeps that session lease and releases only the separate capture pin.

## Commands and evidence

Abbreviations: `WORK` is the assigned Werkstatt root, `SPEC=$WORK/curator-spec`,
`VAL=$WORK/.val/TASK-261021-vb3z6f-r2`, `LOG=$SPEC/.temp/TASK-261021-vb3z6f/r2`,
`PY=$SPEC/.temp/TASK-261021-yiyrus/venv/bin/python3`.
`g` means `git --git-dir="$WORK/curator-spec.git" --work-tree="$SPEC"`.
The authorized `agent-git-env` helper was sourced before fetch, commit and push.

| Command/check | Exit | Evidence/result |
| --- | ---: | --- |
| `a2a inbox`; assignment query; all four precondition resource gets/reads; review outcome read; ack | 0 each | New orchestrator request, correct assignee and briefs confirmed before editing. |
| `g fetch origin` | 0 | Current candidate head fetched; materialized source retained as `cip02-candidate.txt`. |
| `g switch exact-plan-final-form` | 0 | Continued the explicitly requested existing branch at the previous reviewed head. |
| `g diff --check`; scoped `g add`; `g commit -S -m "Clarify distinct project session and snapshot capture leases"` | 0 each | One-file reviewed scope; signed child of the previous head. |
| `g log -1 --format="%H%n%P%n%G?"` | 0 | Exact head, parent and `G` recorded. |
| `env -u GIT_DIR -u GIT_WORK_TREE git clone --no-hardlinks --no-checkout "$WORK/curator-spec.git" "$VAL"`; detached checkout of handover head | 0 each | Normal `.git`, exact head, clean source. |
| `make -n validate` | 0 | `make-recipes.log`; records unchanged recipes without claiming execution of one combined process. |
| `python3 tools/validate.py` with validation environment below | 0 | `validate.log`; 74 schemas / 1294 vector files. |
| `"$PY" -B "$LOG/run-groups.py" list "$LOG/groups.json"` with validation environment | 0 | `groups-list.log`; 879 tests, 31 complete ordered groups. |
| `"$PY" -B "$LOG/run-groups.py" N` with validation environment, N=1..31 | 0 each | `python-group-NN.log`, `group-exits.json`; counts/times/claims below. |
| `"$PY" -B "$LOG/check-groups.py"` with validation environment | 0 | `coverage-check.log`; complete coverage and intended omission/duplication rejections. |
| `go test ./tools/...` with validation environment | 0 | `go-tools.log`; generate-vectors passed. |
| `"$PY" -B "$SPEC/.temp/TASK-261021-vb3z6f/check-draft-01.py"` with validation environment | 0 | `draft-check.log`; positive structure/links and specific broken-link rejection. |
| Final clone head, clean status, normal `.git`, signature using source repository's configured public allowed-signers file | 0 | `clone-final.log`: handover head and `G`. |
| Source checkout head/branch/clean/signature assertions; `g diff --check HEAD^ HEAD`; `g merge-base --is-ancestor 0c1f830fc4006299f463f731af2310ec5d9f8eb1 HEAD` | 0 each | Correct scope and non-rewritten ancestry before publication. |
| `g push -u origin exact-plan-final-form` | 0 | Fast-forward publication from `0c1f830` to `cf1241c`, no force. |
| `g ls-remote --heads origin refs/heads/exact-plan-final-form refs/heads/cip02-ready-for-acceptance`; SHA assertions | 0 | `remote-heads.log`; exact feature head and unchanged pinned candidate. |
| Guarded `rm -r "$VAL"`; absence assertion | 0 | `cleanup.log`; exact-head validation clone removed. |
| Tool version commands | 0 each | `tool-versions.log`. |

Validation command environment prefix (executed from `VAL`):

```sh
env -u GIT_DIR -u GIT_WORK_TREE TMPDIR=/tmp/cw-dev1   GOCACHE="$WORK/.cache/go-build" PYTHONDONTWRITEBYTECODE=1   PATH="$SPEC/.temp/TASK-261021-yiyrus/venv/bin:$PATH"
```

The Draft/coverage/discovery helper invocations need no Go cache; those commands
used the same git unsetting, TMPDIR and bytecode settings with the explicit
virtualenv interpreter. The cloned fixture directory for the Draft checker was
created under its ignored `.temp/TASK-261021-vb3z6f` before invocation.

Procedural retries, not passing evidence: the initial working-tree Draft check
found that the newly materialized candidate `.md` under scratch had broken
relative links when scanned from that scratch location. The candidate was
renamed to `.txt`, then that check and the exact-head clone check passed.
The initial failure was followed by `cat`, so the enclosing shell returned 0;
its traceback is retained in `draft-precommit.log` and is not reported as a pass.
An initial helper-edit attempt used a clone-relative path that did not exist
and raised FileNotFoundError before writing; the following discovery made that
shell return 0. The edit was corrected to an absolute workspace path and the
final 31-group partition was regenerated and independently checked.
The first clone signature display returned 0 with an allowed-signers configuration
diagnostic; it was not accepted as verification. Reusing the source repository's
configured public allowed-signers file produced `G` with an explicit assertion.
The signed source head had already been verified before publication.

Logs, exact test-ID manifest, completed counts/durations, exit codes and helper
scripts are retained under `LOG`, outside the deleted validation clone.
`lease-crosswalk.md` retains the manual consistency review; its full content is
also included above. `commit-scope.log` records the signed commit scope.

## Python groups and claims

Each row is the group command with N equal to its group number; full output is
`python-group-NN.log`. The manifest records the exact test IDs in every subset.
These are existing fixture assertions, not runtime implementation evidence for
this Draft.

| Group | Module | Tests | Time | Exit | Claim exercised |
| ---: | --- | ---: | ---: | ---: | --- |
| 1 | `test_allowed_signers` | 7 | 0.021s | 0 | Valid signer lines and malformed/unknown key rejection. |
| 2 | `test_implementation_coverage` | 39 | 0.864s | 0 | Inventory mappings and accepted/rejected stream parsing. |
| 3 | `test_local_hosted_launches` | 10 | 0.001s | 0 | Launch selection, conflicting flags and forbidden receiver probes. |
| 4 | `test_local_setup_token` | 17 | 21.389s | 0 | Local-source refusals, preservation and secret/branch control plants. |
| 5 | `test_login_readiness` | 9 | 0.537s | 0 | Metadata controls and malformed/unindexed-case rejection. |
| 6 | `test_managed_home_snapshots` | 16 | 1.332s | 0 | Snapshot schema, adapter/profile binding and dispatch. |
| 7 | `test_muse` | 11 | 1.308s | 0 | Fragment/vector controls and forbidden channels. |
| 8 | `test_muse_launch_argv` | 6 | 0.014s | 0 | Argv exclusions and typed overrides. |
| 9 | `test_muse_v4` | 7 | 1.854s | 0 | Accepted v4 cases and backend/schema rejection controls. |
| 10 | `test_network_contacts` | 8 | 0.041s | 0 | Blocking canary and narrowing-mutant controls. |
| 11 | `test_profile_codex_settings` | 8 | 2.629s | 0 | Closed settings enums and qualification assertions. |
| 12 | `test_rc7_carriers` | 18 | 30.141s | 0 | Invalid carrier plants and ambient/status-store invariants. |
| 13 | `test_release_gate` | 5 | 63.070s | 0 | Exact manifest subset of artifact, pin, review and malformed-release assertions. |
| 14 | `test_release_gate` | 5 | 35.896s | 0 | Exact manifest subset of artifact, pin, review and malformed-release assertions. |
| 15 | `test_release_gate` | 5 | 31.888s | 0 | Exact manifest subset of artifact, pin, review and malformed-release assertions. |
| 16 | `test_release_gate` | 5 | 34.654s | 0 | Exact manifest subset of artifact, pin, review and malformed-release assertions. |
| 17 | `test_release_gate` | 5 | 36.975s | 0 | Exact manifest subset of artifact, pin, review and malformed-release assertions. |
| 18 | `test_release_gate` | 5 | 17.670s | 0 | Exact manifest subset of artifact, pin, review and malformed-release assertions. |
| 19 | `test_release_gate` | 5 | 10.842s | 0 | Exact manifest subset of artifact, pin, review and malformed-release assertions. |
| 20 | `test_release_gate` | 5 | 3.684s | 0 | Exact manifest subset of artifact, pin, review and malformed-release assertions. |
| 21 | `test_skillfile_sources_independence` | 8 | 4.634s | 0 | Accepted candidate and citation violation controls. |
| 22 | `test_source_marker` | 14 | 21.373s | 0 | Ownership/refusal positive and negative fixtures. |
| 23 | `test_validate` | 100 | 113.738s | 0 | Exact manifest subset of schema, vector, semantic and text-contract assertions. |
| 24 | `test_validate` | 100 | 1.532s | 0 | Exact manifest subset of schema, vector, semantic and text-contract assertions. |
| 25 | `test_validate` | 100 | 181.425s | 0 | Exact manifest subset of schema, vector, semantic and text-contract assertions. |
| 26 | `test_validate` | 100 | 17.328s | 0 | Exact manifest subset of schema, vector, semantic and text-contract assertions. |
| 27 | `test_validate` | 100 | 2.005s | 0 | Exact manifest subset of schema, vector, semantic and text-contract assertions. |
| 28 | `test_validate` | 74 | 58.663s | 0 | Exact manifest subset of schema, vector, semantic and text-contract assertions. |
| 29 | `test_verify_release_commit` | 5 | 0.002s | 0 | Accepted verification and rejected signatures. |
| 30 | `test_verify_release_merge_policy` | 5 | 0.002s | 0 | Accepted and forbidden merge policies. |
| 31 | `test_workspace_trust` | 77 | 0.409s | 0 | Carrier and trust-transition valid/invalid fixtures. |

## Open decisions

No remaining lease disagreement was found against the fetched and rechecked
CIP-0002 candidate. The separate Curator capture binding, serialized GC-root
accounting and successor project snapshot/receipt contract still require owning
publication before activation, as explicitly stated in the Draft. Existing Draft
owner questions and native qualification gates remain open; this handover does
not claim their adoption or a review verdict. No human decision blocks this
requested Draft rework.
