# Review verdict — BUG-260802-3ibgu1 (csk-go-e2e-readiness-audit)

**Reviewer run:** `RUN-260802-bc3127` (`[reviewer] reviewer (claude)`), not goal-bound (`task-board spawn goal` → `Active Goal: none`).
**Reviewed producer run:** `RUN-260802-7aa31c` (`[analyst] researcher (codex)`, exit 0).
**Reviewed artifact:** `BUG-260802-3ibgu1_csk-go-e2e-readiness-audit.md`, SHA-256 `fc8a2e8d80e8321ecdd628b04fad8956e908fbc9cbd16d425ce5a03675b9dd07`.
**Verdict:** **ACCEPTED** → `done`.
**Mode:** read-only. No code, repository, worktree, branch, pin, tag, release, or delivery-task status was modified by this review.

## 1. Verdict summary

The audit satisfies every clause of the `BUG-260802-3ibgu1` acceptance criterion. It is independently
fact-checked: 16 distinct verification commands were re-run against the live repositories, and **every
substantive claim reproduced exactly** — no SHA, digest, reason code, env-var name, line citation, or
policy statement was found wrong.

The document is correctly scoped as a readiness *plan*, not a delivery. It does not claim gates it did
not run, does not fabricate the post-merge base SHA, and correctly refuses to treat the existing native
fixture as install/launch E2E evidence. Four cosmetic imprecisions are recorded in §4; none changes a
conclusion, a command's semantics on the primary path, or the handoff order, so none blocks acceptance.

## 2. Independent verification log (this review)

All commands re-run by the reviewer; results compared against the audit's assertions.

| # | Verification | Result | Audit claim |
|---|---|---|---|
| 1 | `git status --porcelain=v2 --branch` in cocoaskills | clean `main`, `dacccaaf3ed18740a4d501fe8a3bfec64644c03e`, `+0/-0` | matches |
| 2 | `git merge-base --is-ancestor 53b4eb0a… HEAD` | exit 0 | matches |
| 3 | `git rev-parse HEAD` in curator-spec | `432eb2ee1fe2d6b271e37269f867c8851c325539` | matches |
| 4 | `shasum -a 256 conformance/v1/manifest.json` | `12e58b82579645ba1ccafba49d3e2dd3216005ddf37ae63c68a9fafd46773071` | matches |
| 5 | `git diff --quiet 0c81c1f8… 432eb2ee… -- conformance release/1.0.0-rc.6.json` | exit 0 | byte-equivalence confirmed |
| 6 | `git ls-remote origin refs/heads/task/TASK-260720-12r55p-shared-v6-vectors` | `6e7742f0d28ad95ddd7d8e92364b84062571ad0b` | matches |
| 7 | `gh pr view 19 --json state,headRefOid,mergeable` | `OPEN`, `MERGEABLE`, head `6e7742f0…` | matches |
| 8 | `.github/workflows/ci.yml` matrix + pin | 3 OSes × Python 3.11–3.14; `ref: 0c81c1f8…` at line 32 | matches |
| 9 | `src/csk/builds/toolchain.py` | `TESTED_GO_FAMILIES: Final[tuple[str, ...]] = ("1.25",)` | matches |
| 10 | `protocol/core.md` | "toolchain MUST be Go 1.23 or newer" | floor confirmed |
| 11 | `conformance/v1/vectors/go-host-execution-policy.json` | `"exhaustive": true`, `"platforms": ["macos","windows"]` | matches |
| 12 | `git ls-remote --tags` curator-spec | `v1.0.0-rc.5` present, **no** `v1.0.0-rc.6` | matches |
| 13 | `release/1.0.0-rc.6.json` | `committed_release_pin_advanced=false`, `claims_emitted=[]`, macOS/Windows `pending-downstream-native-evidence`, `linux_excluded_until_task` | matches |
| 14 | `grep build_execution_control_unavailable` | real code at `src/csk/builds/go_v1.py:56`; expected_error in vectors; `protocol/core.md:378,451` "before starting the worker" | matches |
| 15 | Env-var names in handoff commands | `CSK_GO_V1_MANAGER_EXECUTABLE`, `CSK_GO_V1_GO_EXECUTABLE`, `CURATOR_CONFORMANCE_ROOT` all real and read exactly as documented | matches |
| 16 | `shasum -a 256` of the outcome artifact | `fc8a2e8d…` | matches the hash recorded in task notes |

### Cited-file existence check

All cited paths exist at the stated locations. `tests/protocol_lifecycle_observations.py` is **absent from
`main`** but **present at PR 19 head `6e7742f0…`** (`git cat-file -e` exit 0) — consistent with the audit,
which attributes it to PR 19, not to the current base. `tests/test_go_build_e2e.py` and
`tests/fixtures/skill_go_e2e/` correctly do **not** exist; they are the deliverable.

### Load-bearing characterizations spot-checked in source

- `tests/test_builds_go_v1_fixture.py:65-68` — `skipif` reason is literally "the portable source-aware
  policy covers exactly macOS and Windows"; `:89-95` skips without `CSK_GO_V1_MANAGER_EXECUTABLE`;
  `:409` asserts `not marker.exists(), "the verified output must never be launched"`. The audit's
  central claim — that this file is **direct-driver** evidence and explicitly asserts the artifact is
  *never launched*, therefore cannot satisfy install/launch E2E — is exactly right and is the single
  most important judgment in the document.
- `tests/test_build_activation.py:470-500` — artifacts are synthetic bytes
  (`b'#!/bin/sh\necho "args:$*"\nexit 9\n'`) injected via `_publish(...)`, not produced by a CocoaSkills
  install. The audit's "synthetic cached artifact bytes" characterization is precise.

## 3. Acceptance-criterion mapping (this task)

| Clause of BUG-260802-3ibgu1 AC | Where satisfied | Status |
|---|---|---|
| Publish a task-scoped outcome resource | Registered in `outcomeResources` with task-ID-prefixed name | met |
| Map **every** 3pemm6 AC to evidence or a concrete missing step | §3 AC 1–5, each with "Existing evidence" + "Concrete gap" | met |
| Exact worktree/base/fixture/CI commands for macOS, Windows, Ubuntu | §4.1 bash, §4.2 PowerShell, §4.3 fixture provenance, §5 three platform blocks + gates | met |
| Record release-boundary constraints | §2 provenance/ownership table, §3 AC 4, §7 stop conditions | met |
| Ordered handoff plan startable immediately after 12r55p lands | §6, ten ordered steps with the dependency gate first | met |

Coverage of the target task's own AC sentences was checked clause by clause — vendored networkless
build/explicit launch/argv+exit/shims (→AC 1); the eight named lifecycle behaviours, cache hits through
two-project preservation (→AC 2, as a 9-item list that adds repair/status); native matrix + Ubuntu
unavailable-control (→AC 3); rc.6 root/digest with no pin/tag/release/claim change (→AC 4); pytest,
mypy, build, twine, diff gates (→AC 5). **No clause of TASK-260720-3pemm6 is left unmapped.**

## 4. Non-blocking findings for the implementer

1. **`sha256sum` is not stock on macOS.** §5's macOS and Ubuntu blocks and the candidate-authentication
   gate use `sha256sum`. It resolved here only via Homebrew (`/opt/homebrew/bin/sha256sum`); a stock
   `macos-latest` runner does not have it. The document notes a Windows substitution but not a macOS
   one. Use `shasum -a 256` on macOS. *(This reviewer hit the same issue and had to substitute.)*
2. **`rg` in the pin guard.** The no-pin-change guard pipes through `rg`, which is not a stock runner
   tool. `grep -o` is equivalent and portable.
3. **`TASK-260720-3pemm6` is `backlog`, not `blocked`.** §6 step 1 says "while it remains blocked";
   the board reports `status:backlog`, `isBlocked:false`, `blockedBy:[]`. Wording only — the
   instruction not to touch the task is correct and was honoured.
4. **Linux-exclusion owner unnamed.** §2 says "Linux excluded" without citing
   `TASK-260728-1skseh` (`run-linux-native-external-repository-qualification`), which
   `release/1.0.0-rc.6.json` records as `linux_excluded_until_task`. Worth citing alongside the other
   ownership boundaries.

## 5. Scope and boundary compliance

- Read-only mandate honoured: cocoaskills is still clean `main` at `dacccaaf…` with `+0/-0`; curator-spec
  still at `432eb2ee…`; PR 19's worktree untouched; no new worktree created.
- No status change to any delivery task: `TASK-260720-3pemm6` remains `backlog`,
  `TASK-260720-12r55p` remains `reviewing`.
- No release pin, tag, GitHub Release, or conformance claim was altered.
- Ownership boundaries in §2 verified against the board: `TASK-260720-25d05o` =
  `qualify-protocol-release-evidence`, `TASK-260720-1utsx8` = `audit-csk-released-suite-pin`. Both exist
  and match their described roles.

## 6. Gate status

No product validation suite applies: this task changed no code, so "tests green" has no executable
target. The audit is explicit about this and cites existing CI results as evidence rather than
reporting them as gates it executed — the honest framing, and the reason this item is not treated as an
unmet DoD requirement. Full pytest/mypy/build/twine/diff gates are correctly deferred to
`TASK-260720-3pemm6`, where §5 and §6 step 9 require them at the exact task head.

## 7. Handoff to the commit-owning mover

Per the reviewer constraint, this run supplies **no `commit_ack`**. Acceptance evidence is recorded here
for the commit-owning mover, which commits the scope and performs the final `done` transition with
`commit_ack=scope_committed`.

The accepted artifact is ready to drive `TASK-260720-3pemm6` as soon as `TASK-260720-12r55p` is accepted
and PR 19 lands. The implementer should start at §6 step 2 (freeze base and resolve `BASE_SHA` by
`6e7742f0…` ancestry) and apply the four §4 corrections above in passing.
