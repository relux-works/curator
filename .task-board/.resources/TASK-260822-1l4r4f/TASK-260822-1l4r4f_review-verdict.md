# TASK-260822-1l4r4f — review verdict: ACCEPTED

**Reviewer run:** RUN-260822-378a01 (claude, read-only)
**Date:** 2026-08-22
**Artifact under review:** `TASK-260822-1l4r4f_analysis.md` revision 2 (796 lines), board outcome resource; byte-identical copy at `.research/260822_decision-0008-open-questions.md` (`diff -q` clean)
**Spec baseline verified against:** `/Users/iv/Developer/ReluxWorks/curator-spec` `main` = `a2d44eb` (rc.8), plus `/Users/iv/Developer/ReluxWorks/curator`

## Verdict

**Accepted.** The acceptance criterion — "analysis.md outcome resource answers all five questions
with a single recommendation each and enough rationale for the prose tasks to cite without further
input" — is met. Four refinements are recorded below for the downstream prose tasks; none of them
is rework of this deliverable.

## AC check

| Requirement | Result |
|---|---|
| Five questions, one recommendation each | met — Q1 §2, Q2 §3, Q3 §4, Q4 §5, Q5 §6; each has exactly one "Recommendation" block |
| Rationale per recommendation | met — 5–8 numbered arguments each, every one anchored to a `path:line` or a vendor citation |
| At least one rejected option per recommendation | met — 4/5/5/5/8 rejected options respectively, each with the reason for rejection |
| Consistent with the decision 0006 doctrine (mandatory portable / native inventory / deferred hardened) | met — no new control enters the mandatory set, no deferred guarantee is named as an inventory entry, `network-isolation-domain` is deliberately not spelled `total-network-denial` (`core.md:443-450` freezes those six names against exactly that) |
| Enough for prose tasks to cite without further input | met — every question carries a "Prose hooks" block naming the target file and the sentence to write; the four consumer tasks (`1f533i`, `1mwy10`, `3fkfmf`, `f4qv7w`) are named correctly |
| Findings written to file, linked as a task-scoped outcome resource | met |
| Logbook | met — `LOGBOOK.md` entry `2026-08-22 — TASK-260822-1l4r4f`, including the revision-1 correction and the silent-prior-artifact anomaly |
| Tests green | n/a — analysis-only task, no product code touched. The `internal/config`, `internal/skillspec` working-tree changes predate this run (cocoaskills parity) and are not attributable here |

## Independent fact-check

I re-derived the ledger rather than trusting it. **20 of 24 in-repo claims re-verified at first hand**;
the 3 load-bearing external claims re-fetched from the primary sources.

Spot-verified in `curator-spec@a2d44eb`:

- `core.md:437` `| an execution_policy other than manager-worker-v1 | build_execution_hardened_claim_forbidden |` — **exact**. This is F2's whole load, and it holds: a `script-worker-v1` record under `capability-evidence-v1` is rejected by the current normative text, so decision 0008 §4 is genuinely not implementable as written and the new record version is the minimal correction. `core.md:1306-1311` independently forbids fixing it by widening the closed constant.
- `core.md:412-414` "a cached, inherited, or configured result is not a probe" — **exact**; Q4's per-install-generation rejection is textual, not stylistic.
- `core.md:325-327`, `369-370`, `421-422`, `439-442`, `443-450`, `461-468`, `557-558`, `942-968`, `1288-1315` — all **exact** at the quoted lines.
- `manager.md:211-219`, `246-257`, `268-273`, `503-509` — all **exact**. The rc5 inventory table reproduced in Q5 matches `manager.md:246-257` cell for cell.
- `common.schema.json`: `capabilities.filesystem.default = "repo"` **confirmed**, against 0008 §3's "no writes outside the private runtime area" — the F1 conflict is real. Host-glob pattern `^[^\s/\\]+$` **confirmed** (no scheme, port, or path). `scriptCommand` is `additionalProperties: false`, so Q1's "new `$defs`, schema 7 bytes frozen" is the only legal shape.
- `agent-skill-v7.schema.json` top-level properties **confirmed** — `capabilities` is manifest-level, `commandV7` carries none. F6 stands.
- `go-host-execution-policy.json`: `platforms = [macos, windows]`, `unavailable_reasons = [no-private-aggregate-domain]`, and the `unknown-native-control-is-rejected` case with control `host-firewall-profile` — **all exact**.
- `provider-capability-receipt-v1.schema.json`: exactly six `established` ids, `minItems`/`maxItems` 6 — **confirmed**; F4's 0007 naming (`verified-provider-policy-v1`, `verified-provider-execution-v1`, `host-execution-provider-v1`) verified at `decisions/0007:26-28,49-51`.
- Zero `linux` hits in `core.md` and `manager.md` — **confirmed**; `linux` already in the `conformance-claim-v1..v4` / `verified-provider-v1` enums, so F5 introduces no new vocabulary.
- `curator/internal/runtimestore/runtimestore.go` `WriteBinShim` — **confirmed** symlink on unix, `@echo off\r\n"<path>" %*` on Windows. F3's "the current launcher structurally cannot host enforcement" is a fact about shipped code, not a reading.
- `curator/.github/ci/platform-cases.tsv` rows 101–102 — **confirmed** (`TestPerFileSizeLimitIsReallyApplied` darwin/windows, `TestBuildFailsClosedWhenTheGoChildCannotStart` windows/darwin, class `platform-control`).
- `STORY-260822-2h0v9j` AC reads verbatim "Conformance suite for script-worker-v1 green on **ubuntu**/macos/windows lanes per the platform-case ledger" — **confirmed**. This is the strongest single support for the Linux column, and it is board fact rather than inference.

External, re-fetched from primary sources by this reviewer:

- **Microsoft Learn, `UpdateProcThreadAttribute`** — the quoted sentence is verbatim: `PROCESS_CREATION_CHILD_PROCESS_RESTRICTED` "is only effective in sandboxed applications (such as AppContainer) which ensure privileged process handles are not accessible to the process… if a process restricting child process creation is able to access another process handle with PROCESS_CREATE_PROCESS or PROCESS_VM_WRITE access rights, then it may be possible to bypass the child process restriction." Revision 2's reversal of revision 1 is **correct and vendor-sourced**. `PROC_THREAD_ATTRIBUTE_CHILD_PROCESS_POLICY` = Windows 10 / Server 2016+ confirmed; `PROC_THREAD_ATTRIBUTE_HANDLE_LIST` carries no version note of its own and inherits the function's Vista minimum — the "documented Vista+" reading is fair, and the doc's own note "if you use this attribute, pass in a value of TRUE for `bInheritHandles`" matches mandatory-set item 7 exactly.
- **kernel.org Landlock** — 5.13+, `CONFIG_SECURITY_LANDLOCK=y`, unprivileged under `no_new_privs`, `LANDLOCK_ACCESS_FS_EXECUTE` and `..._WRITE_FILE`, inherited across `clone(2)` and never removable, network = TCP bind/connect **ports** (ABI 4) and UDP (ABI 10). **All confirmed.** Claim 20 is the sharpest argument in Q3 and it holds: the strongest kernel mechanism available anywhere cannot express a hostname, so host filtering fails the inventory test on every platform, not merely portably.

## Findings for the prose tasks (non-blocking)

1. **Ledger claim 21 overstates the version range.** The analysis asserts "Ubuntu 23.10+/24.04 default to `kernel.apparmor_restrict_unprivileged_userns=1`" and marks it "verified for the sysctl and the default". Its own cited blog says the opposite for 23.10 — the feature shipped **opt-in** on release day, with default-on planned via SRU. Canonical's 24.04 material is where the default lands, and it confirms the substantive half precisely: in 24.04 LTS "the use of unprivileged user namespaces is allowed for all applications but access to any additional permissions within the namespace are denied". **No recommendation changes** — `ubuntu-latest` is 24.04, so the `host-conditional` mapping for `network-isolation-domain` is right, and the claim was already hedged. Narrow the range to 24.04 before `3fkfmf` or `f4qv7w` cites it.

2. **Q3 contradicts decision 0008's own text and is not listed as a conflict.** 0008 §3 says kernel-grade host filtering "**enters the native inventory per platform** and is never silently claimed." Q3 recommends the reverse — no host-filtering entry in the script inventory at all. I agree with Q3 on the merits (claim 20 settles it), but the analysis flags exactly three 0008 conflicts (F1 schema default, F2 evidence record, F5 Linux) and this fourth divergence belongs in that list. `1f533i` and `f4qv7w` will otherwise inherit it silently, and a maintainer diffing the prose against the accepted decision will find it unannounced.

3. **`host-conditional` → `host-capability` does not fit the existing ledger class.** Q5's platform-case table routes every host-conditional control to class `host-capability`. `.github/ci/platform-cases.tsv` defines that class as "the runner **filesystem** cannot create the artefact the vector needs (symlink, hard link, reparse point, FIFO, non-UTF-8 name)" — a missing Landlock/cgroup/netns kernel feature is not that, and claim 24 paraphrases the definition with "filesystem" dropped. `f4qv7w` needs either a widened class definition or a new class (`host-kernel-feature`); better decided in the prose than discovered when `platform-case-gate.sh` matches the class against the reason text the test actually printed.

4. **Q3 rationale point 1 claims more than its evidence.** "The spec already rules on this, in the vectors" — the `unknown-native-control-is-rejected` vector proves only that a control outside the **rc5 build** inventory is rejected; it does not constrain what a *new* inventory may contain. The vector exists exactly as described; the inference is the overreach. Points 2, 5, and 6 carry the recommendation on their own, so this costs nothing except a citation a reviewer can push back on. Soften it when `1f533i` lifts the argument.

## What I checked and did not fault

- **Q1 vs `core.md:325-327`** ("the policy identity … is never a package-visible option"). The apparent contradiction is real on a first read, and the analysis meets it head-on rather than routing around it: a one-value enum is an opt-in to enforcement, not a choice of policy, and the influence is monotonically restrictive. That reasoning is sound and it already schedules the disambiguating sentence in the prose hooks.
- **The Linux column (F5/Q5)** is the largest call in the document and it is honestly labelled as such — maintainer item 4 states the reversal, the cost, and a cheap undo path. It also survives the strongest objection I could raise: with the column, the three new controls are `host-conditional` on Linux and `unavailable` elsewhere, so they are never *guaranteed* on any platform, which is uncomfortably near the "promise in a machine-readable file" that Q3 rejects. The distinction holds — `host-conditional` licenses applying and reporting the control on hosts that have it, which `unavailable` forbids outright — and `STORY-260822-2h0v9j`'s ubuntu lane makes the column load-bearing rather than speculative.
- **Divergences from the build policy (open stdin, streamed output, no wall-clock deadline)** are correctly identified as deliberate, justified against `manager.md:508-509`'s transparency requirement, and restated as a positive control (explicit manager-controlled stream binding) rather than a dropped one. This is the kind of thing that silently breaks a CLI shim if inherited by reflex; catching it here is worth more than most of the document.
- **Process discipline.** The prior spawn's orphaned artifact was reconciled via `resource update` rather than overwritten, and the anomaly is recorded in the logbook with the operational lesson.

## Handoff

No `commit_ack` supplied — reviewer archetype. Uncommitted deliverables at review time:
`.research/260822_decision-0008-open-questions.md` (untracked) and the `LOGBOOK.md` entry. The
commit-owning mover should commit that scope. The board artifact itself is already persisted.
