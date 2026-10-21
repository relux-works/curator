# TASK-261021-vb3z6f — revision 3

## Delivery and scope

Reapplied the two accepted commits onto the fetched current `origin/main` on a
new branch, using signed cherry-picks. The r3 request was the orchestrator's
addressed request in topic `exact-plan-final`; assignment and all five board
preconditions were confirmed and read before the ack and implementation. The
accepted round-2 review was also read.

- New branch: `exact-plan-final-form-r3`
- New base: `9702382be175e2297e943f1ea97cc6e6bd8dda8f`
- New first commit: `f3726868695c6d4091b45bda52094d4b88a81398`
- New head: `76784377893ab578489f8ba098763b5937ed04fa`
- Both new signatures: `G`, verified before publication and in the exact-head
  normal validation clone using the source repository's configured public
  allowed-signers file.
- Original branch `exact-plan-final-form` remains unchanged at
  `cf1241cbb828180e5308a8b02f2671ebedabcedd`, locally and remotely.
- New branch pushed normally; remote head read back equal to the handed-over SHA.

The first cherry-pick had exactly four mechanical conflicts: `CHANGELOG.md`,
`README.md`, `UNRESOLVED_QUESTIONS.md`, and `cips/README.md`. Each was resolved
by keeping both sides' entries, including landed CIP-0022 and CIP-0002 entries.
No existing lines were deleted in the final diff against the new base.
The second cherry-pick applied without conflict.

The optional link replacements were unnecessary: existing commit-pinned
references remain valid and were preserved. The final CIP-0023 Git blob is
identical to accepted round 2. Both cherry-picks' per-file added/deleted line
sequences are identical to their respective originals after excluding hunk
positions and unchanged context. The first range-diff `!` reflects changed
insertion context only; the second patch is `=`. Landed CIP-0002 and CIP-0022
blobs remain identical to the new base.

No design, schema, vector, executable tool or runtime behavior changed. The Draft
owner questions are unchanged. No board status change, force-push, PR or
default-branch push was performed.

## Files changed against the new base

Five original files, 652 insertions and no deletions:

| File | Added | Removed | Purpose |
| --- | ---: | ---: | --- |
| `cips/CIP-0023-exact-plan-hosted-launches.md` | 631 | 0 | Exact accepted Draft, including the two distinct leases. |
| `CHANGELOG.md` | 6 | 0 | Original Draft filing entry after the existing CIP-0022 entry. |
| `README.md` | 5 | 0 | Original proposal discovery paragraph, keeping CIP-0022. |
| `UNRESOLVED_QUESTIONS.md` | 9 | 0 | Original owner-question pointer, keeping CIP-0022 and CIP-0002. |
| `cips/README.md` | 1 | 0 | Original CIP-0023 index row after CIP-0022. |

## Requested range-diff

Command, exit 0 (`g` is defined below):

```sh
g range-diff --no-color   0c1f830fc4006299f463f731af2310ec5d9f8eb1^..cf1241cbb828180e5308a8b02f2671ebedabcedd   9702382be175e2297e943f1ea97cc6e6bd8dda8f..76784377893ab578489f8ba098763b5937ed04fa
```

Complete output (`range-diff.log`):

```text
1:  0c1f830 ! 1:  f372686 docs: draft exact-plan hosted successor with detached sessions
    @@ Commit message
     
      ## CHANGELOG.md ##
     @@ CHANGELOG.md: Versioning for the complete specification set.
    - 
    - ### Added
    +   neutral writer-local `install` options and total per-environment reporting.
    +   No normative schema, conformance corpus or implementation behavior changes.
      
     +- File Draft [CIP-0023](cips/CIP-0023-exact-plan-hosted-launches.md):
     +  snapshot-bound exact-plan hosting with detached/doorbell/session-record
    @@ CHANGELOG.md: Versioning for the complete specification set.
        using the context/MCP path discipline and existing diagnostics. Bind the
     
      ## README.md ##
    -@@ README.md: Protocol and implementation designs are recorded as
    - [CIP-0001](cips/CIP-0001-curator-improvement-proposals.md); adoption follows
    - [GOVERNANCE.md](GOVERNANCE.md).
    +@@ README.md: proposes project schema 3, package-first resolution and digest-approved inline
    + contributions. It remains Draft; no normative schema or implementation support
    + is changed.
      
     +[CIP-0023: Exact-plan hosted launches](cips/CIP-0023-exact-plan-hosted-launches.md)
     +proposes a snapshot-bound successor with detached sessions, doorbells, durable
    @@ README.md: Protocol and implementation designs are recorded as
      [Source index and phase/authority notes](cips/diagrams/README.md) map the B4 module views to their defining CIP sections. These views describe current contracts and proposed targets; they do not assert implementation or qualification.
     
      ## UNRESOLVED_QUESTIONS.md ##
    -@@ UNRESOLVED_QUESTIONS.md: interactive hosted rows remain distinct. `tb-sessiond` lookup/restart is
    - refused under existing hosted §§4/6; enabling it requires an explicit later
    - amendment, not a callback or hidden wrapper. Unqualified vendor child-leak or
    - settings-reload behavior keeps capability unavailable.
    +@@ UNRESOLVED_QUESTIONS.md: The closed leaf-1 candidate and its vectors still need coordinated publication,
    + an unused project-capable outer carrier and per-runtime source-exclusion/route
    + qualification. Published fragment v4/v5 remain unchanged; no implementation
    + support follows from filing this candidate.
     +
     +## Exact-plan successor (CIP-0023)
     +
    @@ cips/CIP-0023-exact-plan-hosted-launches.md (new)
     
      ## cips/README.md ##
     @@ cips/README.md: Keep operator questions numbered and decision-ready, with a recommended answer.
    - | [CIP-0019](CIP-0019-carrier-admission-by-public-key.md) | Carrier admission by public key (swarma-registrar) | Draft |
      | [CIP-0020](CIP-0020-workspace-trust-ledger.md) | Workspace trust ledger for managed homes | Adopted in part (A1) |
      | [CIP-0021](CIP-0021-profile-skill-directories.md) | Package directories in profile skill requirements | Accepted |
    + | [CIP-0022](CIP-0022-mcp-server-declarations-in-project-skillfile.md) | MCP server declarations in the project Skillfile | Draft |
     +| [CIP-0023](CIP-0023-exact-plan-hosted-launches.md) | Exact-plan hosted launches with detached sessions and doorbells | Draft |
2:  cf1241c = 2:  7678437 Clarify distinct project session and snapshot capture leases
```

## Full validation and limits

All three `make validate` recipes passed at exactly
`76784377893ab578489f8ba098763b5937ed04fa` in a disposable normal-`.git` clone.
To follow the standing per-shell runtime limit, the unchanged unittest discovery
was run in 31 bounded groups: five release-gate tests or at most 100 other tests
per group. `make -n validate` recorded the canonical recipes. No single combined
`make validate` process or combined exit 0 is claimed.

| Check | Exit | Claim proved and limit |
| --- | ---: | --- |
| `python3 tools/validate.py` | 0 | 74 schemas / 1294 vector files satisfy the repository's schema, semantic, corpus and local-link checks; runtime qualification remains unverified. |
| Complete discovered Python suite | 0 each group | All 879 existing tests passed, no skipped-test summary. Existing assertions and claim comments were unchanged; exact IDs and group claims are recorded below. |
| Partition checker and control plants | 0 | Actual partition matches complete ordered discovery exactly once; omission and duplication plants are rejected for the intended coverage defect. |
| `go test ./tools/...` | 0 | Existing generate-vectors package assertions pass (49.369s). |
| Draft checker and link controls | 0 | Required template metadata/sections, unique Draft index, neutral text, and actual repository local targets pass. Valid link control passes; broken-link plant produces its specific intended rejection. Remote URLs and anchors are outside this checker. |
| Cherry-pick equivalence and scope assertions | 0 | Both per-file edit sequences equal originals; final Draft blob and original branch are unchanged; new base is exact; five-file final diff is additive; landed CIP-0002/CIP-0022 blobs are preserved. |
| Signature/head/clean-checkout checks | 0 | Both introduced objects verify G; clone and source at exact expected head; normal clone clean before deletion. |

This mechanical reapplication adds no product gate or behavior. New product
negative tests or narrowing mutants are therefore inapplicable; all existing
negative assertions remain in the full suite. The actual partition and link
checker controls were rerun to show the reused validation instruments still
reject their known violations. No native adapter, provider launch or new
lease runtime implementation was qualified by these checks.

Platform/toolchain: macOS arm64; Python 3.14.7, jsonschema 4.25.1,
Go 1.25.5, GNU Make 3.81. Validation commands unset inherited `GIT_DIR` and
`GIT_WORK_TREE`, used `TMPDIR=/tmp/cw-dev1`, and disabled Python bytecode.
Go used the assigned workspace's build cache. After tests, both signatures,
normal `.git`, exact head and clean checkout were verified; the clone was
removed with a guarded non-force command and absence was confirmed.

## Commands and exits

Abbreviations: `WORK` is the assigned Werkstatt root, `SPEC=$WORK/curator-spec`,
`VAL=$WORK/.val/TASK-261021-vb3z6f-r3`, `LOG=$SPEC/.temp/TASK-261021-vb3z6f/r3`,
`PY=$SPEC/.temp/TASK-261021-yiyrus/venv/bin/python3`.
`g` expands to `git --git-dir="$WORK/curator-spec.git" --work-tree="$SPEC"`.
The prescribed `agent-git-env` helper was sourced before fetch, signed
cherry-picks/continuation and push. All other-repository Git commands used that
explicit git directory/work tree; validation-clone Git unset inherited bindings.

| Command/check | Exit | Evidence/result |
| --- | ---: | --- |
| `a2a inbox`; task assignment query; five precondition gets/reads; accepted r2 review get/read; ack | 0 each | Correct addressed r3 request and assignee verified before implementation. |
| `g status --short`; `g fetch origin` | 0 each | Clean source checkout; current main fetched at `9702382`. |
| `g switch main`; `g merge --ff-only origin/main`; equality assertion; `g switch -c exact-plan-final-form-r3 origin/main` | 0 each | New branch from exact current main, not from the old feature branch. |
| `g cherry-pick -S 0c1f830fc4006299f463f731af2310ec5d9f8eb1` | 1 | Expected four mechanical conflicts only, listed above; all entries retained. |
| Scoped conflict resolution, `g diff --check`, scoped `g add`, `g -c commit.gpgsign=true cherry-pick --continue` | 0 each | Signed first reapplication, original edit sequence preserved. |
| `g cherry-pick -S cf1241cbb828180e5308a8b02f2671ebedabcedd` | 0 | Second signed reapplication, no conflict. |
| `g log -2 --format="%H %G? %s"` | 0 | Both new signatures G; `head-scope.log` records full ancestry. |
| Requested `g range-diff --no-color` above | 0 | Full output above and in `range-diff.log`. |
| Inline Python Git edit-sequence/blob/base/scope/original-branch comparisons | 0 | `patch-equivalence.log`; exact assertions summarized above. |
| `env -u GIT_DIR -u GIT_WORK_TREE git clone --no-hardlinks --no-checkout "$WORK/curator-spec.git" "$VAL"`; detached exact-head checkout; normal/clean guards | 0 each | Normal checkout of handover head. |
| `make -n validate` | 0 | `make-recipes.log`, canonical recipe inventory. |
| `python3 tools/validate.py` with validation environment | 0 | `validate.log`, 74 schemas / 1294 vectors. |
| `"$PY" -B "$LOG/run-groups.py" list "$LOG/groups.json"` with validation environment | 0 | `groups-list.log`, exact 879-case discovery / 31 groups. |
| `"$PY" -B "$LOG/run-groups.py" N`, N=1..31, with validation environment | 0 each | `python-group-NN.log`, `group-exits.json`; all group claims below. |
| `"$PY" -B "$LOG/check-groups.py"` with validation environment | 0 | `coverage-check.log`, omission/duplication controls included. |
| `go test ./tools/...` with validation environment | 0 | `go-tools.log`, package passed. |
| `"$PY" -B "$SPEC/.temp/TASK-261021-vb3z6f/check-draft-01.py"` with validation environment | 0 | `draft-check.log`, positive structure/links and intended broken-link rejection. |
| Clone exact-head/normal/clean guards and two-signature assertions using configured public allowed-signers | 0 | `clone-final.log`, both full new hashes followed by G. |
| Source exact branch/head/clean/two-signature guards; `g diff --check 9702382..HEAD` | 0 | Correct signed scope before publication. |
| `g push -u origin exact-plan-final-form-r3` | 0 | New remote branch, no force. |
| `g ls-remote --heads origin refs/heads/exact-plan-final-form refs/heads/exact-plan-final-form-r3`; exact SHA assertions | 0 | `remote-heads.log`, new head verified and original branch unchanged. |
| Guarded `rm -r "$VAL"`; absence assertion | 0 | `cleanup.log`, clone removed. |
| Tool versions, file counts, scope logging | 0 | `tool-versions.log`, `files.log`, `head-scope.log`. |

Validation environment prefix, from `VAL`:

```sh
env -u GIT_DIR -u GIT_WORK_TREE TMPDIR=/tmp/cw-dev1   GOCACHE="$WORK/.cache/go-build" PYTHONDONTWRITEBYTECODE=1   PATH="$SPEC/.temp/TASK-261021-yiyrus/venv/bin:$PATH"
```

The explicit-interpreter Draft and coverage commands used the same git
unsetting/TMPDIR/bytecode settings; no Go cache setting was needed for those
helpers. The Draft checker fixture directory was first created under the clone's
ignored `.temp/TASK-261021-vb3z6f`. No incomplete or interrupted test run is used
as passing evidence in this revision. Logs and helper scripts remain under
`LOG`, outside the deleted validation clone.

## Python groups and claims

Each row is one command with N equal to its group number. Full output is in
`python-group-NN.log`; `groups.json` lists every exact test ID and
`group-summary.json` records checked counts and times.

| Group | Module | Tests | Time | Exit | Claim exercised |
| ---: | --- | ---: | ---: | ---: | --- |
| 1 | `test_allowed_signers` | 7 | 0.091s | 0 | Valid signer lines and malformed/unknown key rejection. |
| 2 | `test_implementation_coverage` | 39 | 1.177s | 0 | Inventory mappings and accepted/rejected stream parsing. |
| 3 | `test_local_hosted_launches` | 10 | 0.040s | 0 | Launch selection, conflicting flags and forbidden receiver probes. |
| 4 | `test_local_setup_token` | 17 | 19.086s | 0 | Local-source refusals, preservation and secret/branch control plants. |
| 5 | `test_login_readiness` | 9 | 0.568s | 0 | Metadata controls and malformed/unindexed-case rejection. |
| 6 | `test_managed_home_snapshots` | 16 | 1.438s | 0 | Snapshot schema, adapter/profile binding and dispatch. |
| 7 | `test_muse` | 11 | 1.011s | 0 | Fragment/vector controls and forbidden channels. |
| 8 | `test_muse_launch_argv` | 6 | 0.014s | 0 | Argv exclusions and typed overrides. |
| 9 | `test_muse_v4` | 7 | 1.644s | 0 | Accepted v4 cases and backend/schema rejection controls. |
| 10 | `test_network_contacts` | 8 | 0.050s | 0 | Blocking canary and narrowing-mutant controls. |
| 11 | `test_profile_codex_settings` | 8 | 2.757s | 0 | Closed settings enums and qualification assertions. |
| 12 | `test_rc7_carriers` | 18 | 13.673s | 0 | Invalid carrier plants and ambient/status-store invariants. |
| 13 | `test_release_gate` | 5 | 108.681s | 0 | Exact manifest subset of artifact, pin, review and malformed-release assertions. |
| 14 | `test_release_gate` | 5 | 22.678s | 0 | Exact manifest subset of artifact, pin, review and malformed-release assertions. |
| 15 | `test_release_gate` | 5 | 14.867s | 0 | Exact manifest subset of artifact, pin, review and malformed-release assertions. |
| 16 | `test_release_gate` | 5 | 25.116s | 0 | Exact manifest subset of artifact, pin, review and malformed-release assertions. |
| 17 | `test_release_gate` | 5 | 15.253s | 0 | Exact manifest subset of artifact, pin, review and malformed-release assertions. |
| 18 | `test_release_gate` | 5 | 13.376s | 0 | Exact manifest subset of artifact, pin, review and malformed-release assertions. |
| 19 | `test_release_gate` | 5 | 24.820s | 0 | Exact manifest subset of artifact, pin, review and malformed-release assertions. |
| 20 | `test_release_gate` | 5 | 2.299s | 0 | Exact manifest subset of artifact, pin, review and malformed-release assertions. |
| 21 | `test_skillfile_sources_independence` | 8 | 5.509s | 0 | Accepted candidate and citation violation controls. |
| 22 | `test_source_marker` | 14 | 17.013s | 0 | Ownership/refusal positive and negative fixtures. |
| 23 | `test_validate` | 100 | 122.630s | 0 | Exact manifest subset of schema, vector, semantic and text-contract assertions. |
| 24 | `test_validate` | 100 | 4.323s | 0 | Exact manifest subset of schema, vector, semantic and text-contract assertions. |
| 25 | `test_validate` | 100 | 205.483s | 0 | Exact manifest subset of schema, vector, semantic and text-contract assertions. |
| 26 | `test_validate` | 100 | 17.963s | 0 | Exact manifest subset of schema, vector, semantic and text-contract assertions. |
| 27 | `test_validate` | 100 | 1.782s | 0 | Exact manifest subset of schema, vector, semantic and text-contract assertions. |
| 28 | `test_validate` | 74 | 49.065s | 0 | Exact manifest subset of schema, vector, semantic and text-contract assertions. |
| 29 | `test_verify_release_commit` | 5 | 0.007s | 0 | Accepted verification and rejected signatures. |
| 30 | `test_verify_release_merge_policy` | 5 | 0.005s | 0 | Accepted and forbidden merge policies. |
| 31 | `test_workspace_trust` | 77 | 0.299s | 0 | Carrier and trust-transition valid/invalid fixtures. |

## Open decisions

No new decision or design change was introduced by r3. Existing Draft owner
questions and qualification/publication gates remain unchanged. There is no
remaining rebase conflict or delivery blocker; review/landing belongs to the
orchestrator. Acceptance of the original r2 is recorded in the board review;
this producer outcome does not claim a new review verdict for r3.
