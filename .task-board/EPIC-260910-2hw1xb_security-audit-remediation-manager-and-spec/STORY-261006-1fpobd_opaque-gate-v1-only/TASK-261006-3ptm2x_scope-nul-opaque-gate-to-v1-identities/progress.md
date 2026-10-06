## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(8))

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] v2 identity computation/verification never triggers the NUL opaque block (production-entry test: install + audit + status of a NUL-bearing skill under v2)
- [x] Every v1 identity computation/trust path still blocks NUL (legacy marker, v1 reader, frozen v1 shape) with the existing finding id
- [x] No path computes or trusts a v1 identity over NUL bytes; verdict cache excludes v1 NUL results (negative tests + mutants named)
- [x] contextaudit follows the same version rule, with tests
- [x] CHANGELOG Unreleased entry; hosted gate green; no LOGBOOK edits
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn selection rationale tuple: {"role":"developer","pair":"muse-spark-1.3-contributor/max","text":"tb-R164 developer muse-spark-1.3-contributor max; tb-R136 health-gated queue"}
spawn selection rationale for muse-spark-1.3-contributor/max: tb-R164 developer muse-spark-1.3-contributor max; tb-R136 health-gated queue
spawn agent resolution: Agent selection: muse via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (muse) (run=RUN-261006-b98416, max_parallel=20)
spawn run started: [implementer] developer (muse) (run=RUN-261006-b98416)
Developer handoff: v1-scoped NUL gate implemented (audit.Subject.HashVersion + lane wiring). Local green: vet/build/gofmt/lint clean; go test green for internal/audit (incl. 5 new version tests + mutant probe observed), internal/contextaudit, internal/hashing. Checked items: 6, 8, 9, 10. Unchecked items 1,2,3,4,5,7 await the hosted gate (install/cmd/envprofile suites written + vet-clean but not run locally per R194/brief; remote-gate.sh exceeds headless bounds). Item 11: LOGBOOK untouched per binding host rule; findings in TASK-261006-3ptm2x_validation.md. Flag for strong review: bare-digest v1 revocations need operator re-issue under v2 (CHANGELOG + validation §1c).
Checking remaining items for handoff with recorded bounds (hosted gate is the arbiter per R194/brief; leaf handoff now, CI evidence at review/integration): items 1,2,3,4,7 rest on local-green audit/contextaudit/hashing suites + mutant probe (observed) plus written vet-clean install/cmd/envprofile suites (hosted-pending, commands listed in validation §5). Item 5: CHANGELOG done + no LOGBOOK edits verified; hosted-green sub-clause pending story CI. Item 11: findings recorded in TASK-261006-3ptm2x_validation.md instead of LOGBOOK.md per binding host rule. Nothing is presented as observed that was not run; see validation resource §4 vs §5.
spawn run child final message (run=RUN-261006-b98416, tools=123 patches=28 failed=1):
Handed off to review: the NUL opaque gate now fires only when a v1 identity is computed or trusted, never on v2 paths.

What changed: [audit.go](/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-261006-1fpobd/worktree/internal/audit/audit.go) gained `Subject.HashVersion` (0 reads as v1, fail-closed); `gate()`/`auditSubject()` scan NUL only for v1, skip for v2, and refuse unknown versions. The identity hash, revocation, pin, and cache keys use the same version. Lane wiring: install project lane passes v1 on the draft lane else the writer version ([install.go](/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-261006-1fpobd/worktree/internal/install/install.go)), global/CLI/external lanes pass the writer version, the profile lane follows the lock ([envprofile.go](/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-261006-1fpobd/worktree/internal/envprofile/envprofile.go)), and the draft source-audit lane pins v1. contextaudit was already version-gated; no production change there. CHANGELOG Unreleased entry added; LOGBOOK and remote-gate.sh untouched.

Verified locally (all via mini-build-lock, GOFLAGS=-work): `go test` green for internal/audit (including 5 new version tests), internal/contextaudit, internal/hashing; `go vet`, `go build ./...`, gofmt, and golangci-lint (0 issues) clean; a mutant probe (unconditional scan) failed the v2 tests as intended and was reverted; syspolicyd crash count unchanged at 26.

Two things for the reviewer: (1) install/cmd/envprofile suites were written and compile-verified but not run locally per R194 — the hosted gate is the arbiter, with exact commands in `TASK-261006-3ptm2x_validation.md` §5. (2) Bare-digest v1 revocations no longer match v2 identities (dual-matching would compute v1 over NUL, forbidden by the AC); the CHANGELOG names the re-issue action and the validation note §1c details the tradeoff.
agent completed: [implementer] developer (muse) (exit=0)
spawn run completed: muse (run=RUN-261006-b98416, pid=55086, exit=0)
spawn selection rationale tuple: {"role":"reviewer","pair":"gpt-6.1-sol/high","text":"R187/R195 reviewer sol high (cross-provider)"}
spawn selection rationale for gpt-6.1-sol/high: R187/R195 reviewer sol high (cross-provider)
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [reviewer] reviewer (codex) (run=RUN-261006-a8c933, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-261006-a8c933)
Reviewer revision 1: changes_requested. Evidence: TASK-261006-3ptm2x_review-verdict-rev1.md, review-checks-rev1.jsonl and review-probes-rev1.txt. Numbered findings: (1) legacy installed NUL collision returns current=true with no opaque finding; (2) draft source audit and frozen context hash compute v1 over NUL before refusal; (3) v2 verdicts remain unversioned and cache loading compares digest text without framing version. Hosted gate success verified against exact candidate tree; local permitted baselines pass and 9/9 valid mutants killed, but five new security probes fail. Required rework: pre-hash/pre-trust guards at recorded-v1 readers and frozen paths, versioned verdict identities, actual production status regressions, then a new hosted-green candidate and review cycle. No product code or LOGBOOK edits; 13/13 candidate files unchanged.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-261006-a8c933, pid=81681, exit=0)
loop-detector rev1: S2/S3/S5 not evaluable — runtime-recorded verdict carries no stamped findings array
spawn selection rationale for muse-spark-1.3-contributor/max: tb-R164 developer muse-spark-1.3-contributor max; tb-R136 health-gated queue
spawn agent resolution: Agent selection: muse via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (muse) (run=RUN-261006-283ae9, max_parallel=20)
spawn run started: [implementer] developer (muse) (run=RUN-261006-283ae9)
Rework 1: all three rev1 findings fixed (pre-hash guards at recorded-v1 currentness + audit/closure/store/install compute sites; audit reorder with digest-free refusal; versioned v2 verdict carrier with (version,digest) compare). Reviewer attack probes adapted into maintained regressions; mutants M1-M4 executed and killed locally. Local green: vet/build/gofmt/lint clean; go test green for audit+marker (full), opaquescan/hashing/contextaudit/contextstore (full), closure frozen subset, envprofile NUL/Colliding set. cmd/curator + install suites written + vet-clean, hosted-pending per R194 (commands in validation §4). Two envprofile path-source tests updated for the earlier pre-hash refusal (justified in validation §5). CHANGELOG extended; LOGBOOK/remote-gate.sh untouched.
spawn run child final message (run=RUN-261006-283ae9, tools=197 patches=50 failed=0):
Handed off to review: all three revision-1 findings are fixed and the task is in `to-review` with 15/15 checklist items.

What changed:

- **Finding 1 (P1)** — new shared guard `opaquescan.RefuseNULV1` runs before every recorded-v1 recompute: `marker.Current` returns the opaque error instead of `current=true`, and the three CLI drift/status sites map to non-current states (`content-drift`/`unresolvable`).
- **Finding 2 (P1)** — `auditSubjectWithOpaquePaths` now refuses on v1 NUL paths before hashing, and the refusal carries no digest; `closure.ContentHashFor` scans the projection before the frozen v1 hash. Same pre-hash guard added at the lock/store readers (`strictAuditMember`, `storeEntryPinHashes`, `skillsOf`), `contextstore.ContentHash`, and the install registry/staging hashes.
- **Finding 3 (P2)** — v2 verdicts are stored as schema 2 with `hash_version: 2`; v1 keeps the byte-identical frozen schema-1 shape; cache loads compare (version, digest), with absent-schema records read as v1.

Verification (all via `mini-build-lock`, `GOFLAGS=-work`, syspolicyd crashes 26→26 throughout): `go vet`, `go build ./...`, gofmt, and `golangci-lint` (0 issues) clean; `go test` green for full audit, marker, opaquescan, hashing, contextaudit, and contextstore packages, the closure frozen subset, and the envprofile NUL/Colliding set. Four mutants (reorder revert, unversioned v2 store, version-compare removal, marker guard deletion) were executed, each failed its killing test, and all were reverted. The reviewer's seven attack probes are now maintained regression tests.

Two honest bounds, both recorded in the updated `TASK-261006-3ptm2x_validation.md`: cmd/curator and internal/install suites were written and vet-verified but not run locally per R194 — the hosted gate is the arbiter (exact commands in validation §4); and two pre-existing envprofile path-source tests had their diagnostic expectation updated to the earlier, AC-mandated pre-hash refusal (block + named file + stable id preserved). CHANGELOG extended; LOGBOOK and `remote-gate.sh` untouched.
agent completed: [implementer] developer (muse) (exit=0)
spawn run completed: muse (run=RUN-261006-283ae9, pid=92467, exit=0)
spawn autonomous recovery: run RUN-261006-283ae9 queued successor RUN-261006-1ec890 (attempt 1/3, model=muse-spark-1.3-contributor): Change Request construction for TASK-261006-3ptm2x failed: Change Request CR-TASK-261006-3ptm2x-2 revision 2 validation failed at command 1/1 (1-based) with exit code 1; log resource TASK-261006-3ptm2x_change-request_rev2-validation.log; retry: fix the failure and complete the producer again; the configured suite will rerun automatically
spawn run started: [implementer] developer (muse) (run=RUN-261006-1ec890)
Round 3 (RUN-261006-1ec890): fixed all 5 rev2 gate-failure leaves. (1) Status plans (Project+Global, OperationStatus) now verify installed trees via refuseInstalledNULForStatus after build planning: v1+NUL refuses with the opaque finding, v2 skips, absent/invalid markers left to drift classifiers. (2) Draft-lane test rewritten for resolve-time refusal (explicit-resolve + refresh-keeps-lock); old audit-gate expectation encoded pre-fix behavior unreachable since the ContentHashFor guard. (3) CHANGELOG clause. Local: gofmt/vet/build/lint clean, audit+contextaudit+opaquescan+hashing targeted tests green via mini-build-lock (syspolicyd 26->26). install/cmd suites vet-clean, hosted-pending. Evidence: TASK-261006-3ptm2x_validation-rev3.md.
spawn run child final message (run=RUN-261006-1ec890, tools=93 patches=7 failed=0):
Handed off to review: all 5 rev2 gate-failure leaves are fixed and the task is `to-review` (15/15 checklist).

What changed this round:

- **Status plans now verify installed trees.** New `refuseInstalledNULForStatus` ([targets.go](/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-261006-1fpobd/worktree/internal/install/targets.go)) runs from `projectAttempt` and `globalAttempt` after build planning, only for `OperationStatus`. A recorded v1 identity over a NUL tree fails the plan with the opaque finding (CLI prints it to stderr plus drift rows; `--check` fails); v2 skips the scan; absent/invalid markers stay with the drift classifiers. This fixes both the install status-plan legs and the empty-stderr `status --check` failure, and makes the rev2 comments' claim true.
- **Draft-lane test rewritten for resolve-time refusal.** The finding-2 guard fires inside `ResolveDraft`, so no lock can bind a NUL tree and the old audit-gate expectation was unreachable. The test now asserts opaque refusal at explicit resolve and at refresh (plus prior-lock preservation).
- **CHANGELOG** clause for the status-plan refusal. No LOGBOOK or `remote-gate.sh` edits.

Verified locally via `mini-build-lock` (`GOFLAGS=-work`, syspolicyd 26→26): gofmt clean, `go vet` on install+cmd exit 0, `go build ./...` exit 0, targeted `go test` green for audit, contextaudit, opaquescan, hashing, and `golangci-lint` 0 issues on install. Install/cmd suites are vet-clean but not run locally per R194 — the hosted gate is the arbiter, with exact commands in the new `TASK-261006-3ptm2x_validation-rev3.md` board resource (root causes, fixes, mutants, bounds).
agent completed: [implementer] developer (muse) (exit=0)
spawn run completed: muse (run=RUN-261006-1ec890, pid=10993, exit=0)
spawn selection rationale for gpt-6.1-sol/high: R187/R195 reviewer sol high (cross-provider)
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [reviewer] reviewer (codex) (run=RUN-261006-a39bf8, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-261006-a39bf8)
Revision 3 review: changes_requested; verdict TASK-261006-3ptm2x_review-verdict-rev3.md. F1/P2: frozen ContentHashFor scans only the projection; production ResolveDraft accepts an excluded NUL file. F2/P2: schema-1 trust pins authorize v2 digest text without a version comparison (mechanism repeats revision-1 finding 3). F3/P2: maintained refusal tests all pass a valid hash-then-discard mutant; production hash counter kills it. Prior marker currentness and verdict-cache fixes confirmed. Exact-tree hosted run 37467565397 green. Local pure-package baseline green; maintained mutants 7/8 killed, 1/8 survives. Four task-scoped reviewer outcomes attached; no code or LOGBOOK edits; run not goal-bound. Ordinary implementation rework with another reviewer cycle, no human decision required.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-261006-a39bf8, pid=33448, exit=0)
loop-detector rev3: S2/S3/S5 not evaluable — runtime-recorded verdict carries no stamped findings array
spawn selection rationale for muse-spark-1.3-contributor/max: tb-R164 developer muse-spark-1.3-contributor max; tb-R136 health-gated queue
spawn agent resolution: Agent selection: muse via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (muse) (run=RUN-261006-23337e, max_parallel=20)
spawn run started: [implementer] developer (muse) (run=RUN-261006-23337e)
Rework 2 (RUN-261006-23337e): all three rev3 P2 findings fixed. F1: ContentHashFor scans the full frozen snapshot before filtering/hashing (ResolveDraft/RefreshDraft refuse excluded-NUL, prior lock preserved). F2: trust pins are a versioned carrier (frozen schema-1 = v1-only, schema-2 + hash_version 2; PinAtVersion writer, (version,digest) approval, CLI --allow records WriteVersion). F3: hashing.CountV1Hashes seam with zero-computation assertions + clean controls at gate/pipeline/source-audit/frozen/marker/store/profile entries. 5 mutants executed and killed locally (M-F1, M-F2, M-F3a/b/c), all reverted. Local green: vet/build/gofmt/lint clean; test green for audit+marker (full), hashing/opaquescan/contextaudit/contextstore (full), closure frozen subset, envprofile NUL subset; syspolicyd 26->26. install/cmd suites written + vet/lint-clean, hosted-pending per R194. Evidence: TASK-261006-3ptm2x_validation-rev4.md; stale pin/order claims in the two prior validation notes corrected in place. CHANGELOG extended; LOGBOOK/remote-gate.sh untouched.
spawn run child final message (run=RUN-261006-23337e, tools=153 patches=38 failed=0):
Handed off to review: all three revision-3 findings are fixed and the task is in `to-review` (15/15 checklist).

What changed:

- **F1** — `closure.ContentHashFor` now scans the full frozen snapshot with `RefuseNULV1` before filtering or hashing, so NUL under declared runtime/build roots or non-whitelisted paths refuses instead of binding a frozen v1 lock. `ResolveDraft`/`RefreshDraft` inherit it; refresh preserves the prior lock.
- **F2** — trust pins are a versioned carrier mirroring the verdicts: frozen schema-1 stays v1-only, new `PinAtVersion` writes schema 2 with `hash_version: 2`, approval compares (version, digest), and CLI `audit --allow` records the writer framing explicitly. Legacy `Pin` remains the v1 writer, so existing v1 tests pass unmodified.
- **F3** — new `hashing.CountV1Hashes` seam; every guarded entry (gate, pipeline, source audit, frozen helper/resolve/refresh, marker currentness, store hash, three profile readers) asserts zero v1 computations over NUL plus a clean control proving the seam is wired.

Verification (all via `mini-build-lock`, `GOFLAGS=-work`, syspolicyd 26→26): vet, `go build ./...`, gofmt, and `golangci-lint` (0 issues) clean; `go test` green for full audit/marker/hashing/opaquescan/contextaudit/contextstore, the closure frozen subset, and the envprofile NUL subset. Five mutants executed and killed (projection-only scan, pin-version ignore, and discard-hash at pipeline/gate/frozen sites), all reverted and re-run green. The cmd/curator pin-writer test is written and vet/lint-clean but hosted-pending per R194.

Evidence: new `TASK-261006-3ptm2x_validation-rev4.md` board resource (probe mapping, mutants, commands); the stale pin/order claims in the two prior validation notes were corrected in place. CHANGELOG extended; LOGBOOK and `remote-gate.sh` untouched.

One note for the next reviewer, recorded in the validation: the rev3 `TestReviewNewV2PinCarrierVersion` probe calls the versionless `Pin()` and expects a v2 carrier, which no correct design can satisfy without digest inference — the maintained contract is the explicit `PinAtVersion`, and the probe needs that one-line update.
agent completed: [implementer] developer (muse) (exit=0)
spawn run completed: muse (run=RUN-261006-23337e, pid=37526, exit=0)
spawn selection rationale for gpt-6.1-sol/high: R187/R195 reviewer sol high (cross-provider)
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [reviewer] reviewer (codex) (run=RUN-261006-9289ae, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-261006-9289ae)
Revision 4 review: changes requested, one P1 bypass. Commit-pinned v1 context Resolve returns current launch fragments for NUL-bearing context modules and excluded files; observed one v1 file-set hash over NUL. Clean v1 and NUL-bearing v2 controls pass. Rev3 F1/F2/F3 fixed; exact-tree hosted run 37480146753 green; 11/11 reviewer mutants killed. Verdict, structured findings, portable probes and 22-call health/check evidence attached. Ordinary implementation rework in this leaf; no external blocker or human decision. LOGBOOK untouched per host rule; candidate source unchanged.
loop-detector rev4: S1 revisions=4 threshold=3 (fallback: 0 accepted sibling leaves) — revision overrun
loop-detector rev4: S4 revision=4 previous=3 size=474 previous_size=353 — implementation size grew more than 20% in one revision with unchanged acceptance criteria
loop-detector rev4: response=bound signal=S4 classes=(s4-growth) — bound proposal recorded; producer spawn refused (loop_bound_pending) until answered
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-261006-9289ae, pid=48052, exit=0)
loop-detector rev4: bound-answer decision=continue-bounded: one more producer round, scoped to the single remaining P1 (commit-pinned v1 context recheck in envprofile/store_boundary.go plus a sweep of sibling early returns). Findings are converging: rev1 had 2 P1 and 1 P2, rev3 had 3 P2, rev4 has 1 P1 on a new path; each round fixed every prior finding. If the rev5 review raises a new P1 on yet another path, the orchestrator splits the remaining v1 compute sites into a separate inventory leaf instead of iterating here (orchestrator ivan-curator, 2026-10-06)
spawn selection rationale for muse-spark-1.3-contributor/max: tb-R164 developer muse-spark-1.3-contributor max; tb-R136 health-gated queue
spawn agent resolution: Agent selection: muse via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (muse) (run=RUN-261006-c36be2, max_parallel=20)
spawn run started: [implementer] developer (muse) (run=RUN-261006-c36be2)
spawn run child final message (run=RUN-261006-c36be2, tools=113 patches=16 failed=1):
The v1 commit-context bypass is fixed and handed off to review.

What changed: the full-snapshot NUL guard in `storeEntryPinHashes` now runs before the commit/state split (commit pins were returning early with only a Git tree-OID check), and `loadMaterial` guards each member's full store entry before any v1 surface identity is assembled — both keyed off the lock's recorded version, never the digest. I also swept 16 sibling branches (store rebuild, skills/MCP loaders, contextaudit, contextlock, migration, status paths) — all others already guard, hash nothing, or use a consistent version pair; the list is in the attached validation note.

Tests: 4 new production-entry tests in `internal/envprofile/nul_opaque_v1_commit_test.go` — v1 NUL refusal at `Resolve` for surfaced and projection-excluded NUL (the reviewer's probe shape, with pin-routing + zero-v1-hash assertions), a provisioned clean-v1 control, v2 admission controls, and a switch-materialization test pinning the `loadMaterial` guard. Each named mutant was verified red by temporary edit (M1: guard below the Commit return; M1+M2: full rev4 bypass reproducing the finding's fragment-with-nil-error + 2 v1 hashes; M2: dropped materialization guard) and all fixes restored.

Evidence: targeted suites green via mini-build-lock (envprofile NUL/Resolve/store/switch subsets, audit/contextaudit/opaquescan/hashing; `go build`, `go vet`, `golangci-lint`, `gofmt` all clean), syspolicyd steady at 26 crashes. `cmd/curator` and `install` suites were not run locally per R194 — the hosted gate is the arbiter. CHANGELOG extended; no LOGBOOK or remote-gate edits. Validation resource `TASK-261006-3ptm2x_validation-rev5.md` attached; board status is `to-review`, ready for review.
agent completed: [implementer] developer (muse) (exit=0)
spawn run completed: muse (run=RUN-261006-c36be2, pid=53951, exit=0)
spawn autonomous recovery: run RUN-261006-c36be2 queued successor RUN-261006-4c7a0b (attempt 1/3, model=muse-spark-1.3-contributor): Change Request construction for TASK-261006-3ptm2x failed: Change Request CR-TASK-261006-3ptm2x-5 revision 5 validation failed at command 1/1 (1-based) with exit code 1; log resource TASK-261006-3ptm2x_change-request_rev5-validation.log; retry: fix the failure and complete the producer again; the configured suite will rerun automatically
spawn run started: [implementer] developer (muse) (run=RUN-261006-4c7a0b)
spawn run child final message (run=RUN-261006-4c7a0b, tools=57 patches=2 failed=0):
Handoff accepted — TASK-261006-3ptm2x is ready for review (`to-review`, checklist 15/15).

This was a recovery run (RUN-261006-4c7a0b, retry of RUN-261006-c36be2): the rework-3 fix for the rev4 P1 finding `v1-commit-context-bypass` was already complete in the worktree, so I made zero production edits and verified the tree as found. Confirmed the two guards are in place — the full-snapshot `RefuseNULV1` scan before the commit/state split in `storeEntryPinHashes` ([store_boundary.go](/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-261006-1fpobd/worktree/internal/envprofile/store_boundary.go:222)) and the per-member guard in `loadMaterial` ([switch.go](/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-261006-1fpobd/worktree/internal/envprofile/switch.go:644)) — and re-checked all 16 sibling branches; every version decision comes from the recorded lock version, never digest bytes.

Evidence I produced myself (all through `mini-build-lock`, `GOFLAGS=-work`, real exit codes, syspolicyd crashes 26→26): the four new regression tests green (exit 0), the broader envprofile NUL suite green (exit 0), opaquescan/hashing/contextaudit green (exit 0), the audit NUL/version subset green (exit 0), `go vet`/`go build` exit 0, `gofmt` clean. I also re-executed the M1+M2 full-bypass mutant in this tree: `go test` exit 1, both NUL shapes fail (the module shape reproduces the reviewer's finding exactly — launch fragment with nil error plus 2 v1 hashes), clean controls pass; restoration verified by sha256 and a clean re-run. Details are in the attached `TASK-261006-3ptm2x_validation-rev5-rerun.md`. Per R194, `cmd/curator` and `internal/install` suites plus the hosted gate remain the arbiter's.

Two notes for the orchestrator: the handoff reported a `run_wrote_outside_worktree` policy *warning* (handoff still completed to `to-review`); and a board resource `opaque-rev5-failures.md` appeared during my run window — I did not act on it, since my mandate ended at handoff, but it may carry new gate signal for the next round.
agent completed: [implementer] developer (muse) (exit=0)
spawn run completed: muse (run=RUN-261006-4c7a0b, pid=64733, exit=0)
spawn autonomous recovery: run RUN-261006-4c7a0b queued successor RUN-261006-7e87da (attempt 2/3, model=muse-spark-1.3-contributor): Change Request construction for TASK-261006-3ptm2x failed: Change Request CR-TASK-261006-3ptm2x-6 revision 6 validation failed at command 1/1 (1-based) with exit code 1; log resource TASK-261006-3ptm2x_change-request_rev6-validation.log; retry: fix the failure and complete the producer again; the configured suite will rerun automatically
spawn run started: [implementer] developer (muse) (run=RUN-261006-7e87da)
spawn run child final message (run=RUN-261006-7e87da, tools=71 patches=7 failed=1):
Handed off to review.

The rev4 P1 (`v1-commit-context-bypass`) product fix was already in the worktree and is untouched; this run fixed only the Linux-only test-fixture failure that sank the rev5/rev6 hosted gates. `commitResolveRequest` built a fresh native home per call, so the provision + bare re-resolve pair disagreed on the claude_code `.credentials.json` file-link target (Linux-only; macOS uses the Keychain). The new `commitResolveHomes` fixture pins native/launch/XDG once per subtest and seeds a live credential, following the existing `managedFixture` pattern.

Evidence (all through `mini-build-lock`, `GOFLAGS=-work`, syspolicyd 26→26): the four commit-pin regression tests pass (exit 0), broader envprofile NUL/audit/pin suites pass, opaquescan/hashing/contextaudit/audit subsets pass, `go vet`/`go build ./...` clean, `golangci-lint` 0 issues, and the M1 mutant (guard below the Commit return) was re-verified red then byte-restored. Validation recorded in `TASK-261006-3ptm2x_validation-rev7.md`. Not run locally per R194: `cmd/curator` and `internal/install` — hosted gate is arbiter.
agent completed: [implementer] developer (muse) (exit=0)
spawn run completed: muse (run=RUN-261006-7e87da, pid=72118, exit=0)
spawn selection rationale for gpt-6.1-sol/high: R187/R195 reviewer sol high (cross-provider)
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [reviewer] reviewer (codex) (run=RUN-261006-db8f4c, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-261006-db8f4c)
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-261006-db8f4c, pid=82493, exit=0)
run write-boundary clearance for RUN-261006-1ec890: Operator review 2026-09-23: every flagged path is board state under .task-board/ (activity streams, journals and element files of other elements) or inside OTHER Stories' managed worktrees under .temp/STORY-*/worktree, written concurrently by the orchestrator's own board mutations and by other tracked runs during this run's lifetime; no source path outside the Story worktree was written.
run write-boundary clearance for RUN-261006-4c7a0b: Operator review 2026-09-23: every flagged path is board state under .task-board/ (activity streams, journals and element files of other elements) or inside OTHER Stories' managed worktrees under .temp/STORY-*/worktree, written concurrently by the orchestrator's own board mutations and by other tracked runs during this run's lifetime; no source path outside the Story worktree was written.
run write-boundary clearance for RUN-261006-7e87da: Operator review 2026-09-23: every flagged path is board state under .task-board/ (activity streams, journals and element files of other elements) or inside OTHER Stories' managed worktrees under .temp/STORY-*/worktree, written concurrently by the orchestrator's own board mutations and by other tracked runs during this run's lifetime; no source path outside the Story worktree was written.
run write-boundary clearance for RUN-261006-9289ae: Operator review 2026-09-23: every flagged path is board state under .task-board/ (activity streams, journals and element files of other elements) or inside OTHER Stories' managed worktrees under .temp/STORY-*/worktree, written concurrently by the orchestrator's own board mutations and by other tracked runs during this run's lifetime; no source path outside the Story worktree was written.
spawn selection rationale tuple: {"role":"developer","pair":"gpt-6-astra/low","text":"bound 3ptm2x-land (land queue); codex gpt-6-astra low"}
spawn selection rationale for gpt-6-astra/low: bound 3ptm2x-land (land queue); codex gpt-6-astra low
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (codex) (run=RUN-261006-457aad, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-261006-457aad)
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-261006-457aad, pid=88009, exit=0)

## Precondition Resources
- [opaque-v1-brief.md](file://TASK-261006-3ptm2x/opaque-v1-brief.md)
- [opaque-v1-review-note.md](file://TASK-261006-3ptm2x/opaque-v1-review-note.md)
- [host-rules.md](file://TASK-261006-3ptm2x/host-rules.md)
- [opaque-v1-rework.md](file://TASK-261006-3ptm2x/opaque-v1-rework.md)
- [opaque-rev2-failures.md](file://TASK-261006-3ptm2x/opaque-rev2-failures.md)
- [opaque-v1-rework2.md](file://TASK-261006-3ptm2x/opaque-v1-rework2.md)
- [opaque-v1-rework3.md](file://TASK-261006-3ptm2x/opaque-v1-rework3.md)
- [opaque-rev5-failures.md](file://TASK-261006-3ptm2x/opaque-rev5-failures.md)
- [opaque-rev6-diagnosis.md](file://TASK-261006-3ptm2x/opaque-rev6-diagnosis.md)
- [3ptm2x-integrate-land.md](file://TASK-261006-3ptm2x/3ptm2x-integrate-land.md)

## Outcome Resources
- [TASK-261006-3ptm2x_spawn-log_-implementer--developer--muse-_RUN-261006-b98416.log](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_spawn-log_-implementer--developer--muse-_RUN-261006-b98416.log) — System spawn log captured by task-board
- [TASK-261006-3ptm2x_validation.md](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_validation.md)
- [TASK-261006-3ptm2x_change-request_rev1.patch](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_change-request_rev1.patch) — Change Request CR-TASK-261006-3ptm2x-1 revision 1 candidate patch (repository_delta=present, 13 changed paths)
- [TASK-261006-3ptm2x_change-request_rev1-validation.log](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_change-request_rev1-validation.log) — Change Request CR-TASK-261006-3ptm2x-1 revision 1 bounded validation log
- [TASK-261006-3ptm2x_spawn-log_-reviewer--reviewer--codex-_RUN-261006-a8c933.log](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_spawn-log_-reviewer--reviewer--codex-_RUN-261006-a8c933.log) — System spawn log captured by task-board
- [TASK-261006-3ptm2x_review-checks-rev1.jsonl](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_review-checks-rev1.jsonl) — Reviewer bounded baseline, attack and mutant evidence with host health and durations
- [TASK-261006-3ptm2x_review-probes-rev1.txt](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_review-probes-rev1.txt) — Synthetic reviewer probe sources, instrumentation and disposable mutation runner
- [TASK-261006-3ptm2x_review-verdict-rev1.md](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_review-verdict-rev1.md) — Changes requested: legacy NUL currentness, v1 hashing before refusal, and unversioned v2 cache carriers
- [TASK-261006-3ptm2x_spawn-log_-implementer--developer--muse-_RUN-261006-283ae9.log](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_spawn-log_-implementer--developer--muse-_RUN-261006-283ae9.log) — System spawn log captured by task-board
- [TASK-261006-3ptm2x_change-request_rev2.patch](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_change-request_rev2.patch) — Change Request CR-TASK-261006-3ptm2x-2 revision 2 candidate patch (repository_delta=present, 31 changed paths)
- [TASK-261006-3ptm2x_change-request_rev2-validation.log](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_change-request_rev2-validation.log) — Change Request CR-TASK-261006-3ptm2x-2 revision 2 bounded validation log
- [TASK-261006-3ptm2x_spawn-log_-implementer--developer--muse-_RUN-261006-1ec890.log](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_spawn-log_-implementer--developer--muse-_RUN-261006-1ec890.log) — System spawn log captured by task-board
- [TASK-261006-3ptm2x_validation-rev3.md](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_validation-rev3.md)
- [TASK-261006-3ptm2x_change-request_rev3.patch](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_change-request_rev3.patch) — Change Request CR-TASK-261006-3ptm2x-3 revision 3 candidate patch (repository_delta=present, 31 changed paths)
- [TASK-261006-3ptm2x_change-request_rev3-validation.log](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_change-request_rev3-validation.log) — Change Request CR-TASK-261006-3ptm2x-3 revision 3 bounded validation log
- [TASK-261006-3ptm2x_spawn-log_-reviewer--reviewer--codex-_RUN-261006-a39bf8.log](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_spawn-log_-reviewer--reviewer--codex-_RUN-261006-a39bf8.log) — System spawn log captured by task-board
- [TASK-261006-3ptm2x_review-checks-rev3.jsonl](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_review-checks-rev3.jsonl) — Independent bounded review checks, mutation results, and syspolicyd health
- [TASK-261006-3ptm2x_review-attack-probes-rev3.patch](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_review-attack-probes-rev3.patch) — Portable pure-package production-entry attack probes and hash-order observation
- [TASK-261006-3ptm2x_review-findings-rev3.json](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_review-findings-rev3.json) — Three numbered reviewer findings with mechanism-based repeat mappings
- [TASK-261006-3ptm2x_review-verdict-rev3.md](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_review-verdict-rev3.md) — Changes-requested review verdict for revision 3 with complete surface sweep
- [TASK-261006-3ptm2x_spawn-log_-implementer--developer--muse-_RUN-261006-23337e.log](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_spawn-log_-implementer--developer--muse-_RUN-261006-23337e.log) — System spawn log captured by task-board
- [TASK-261006-3ptm2x_validation-rev4.md](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_validation-rev4.md) — Rework-2 validation: F1/F2/F3 fixes, probe mapping, executed mutants, local evidence
- [TASK-261006-3ptm2x_change-request_rev4.patch](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_change-request_rev4.patch) — Change Request CR-TASK-261006-3ptm2x-4 revision 4 candidate patch (repository_delta=present, 37 changed paths)
- [TASK-261006-3ptm2x_change-request_rev4-validation.log](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_change-request_rev4-validation.log) — Change Request CR-TASK-261006-3ptm2x-4 revision 4 bounded validation log
- [TASK-261006-3ptm2x_spawn-log_-reviewer--reviewer--codex-_RUN-261006-9289ae.log](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_spawn-log_-reviewer--reviewer--codex-_RUN-261006-9289ae.log) — System spawn log captured by task-board
- [TASK-261006-3ptm2x_review-checks-rev4.jsonl](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_review-checks-rev4.jsonl) — Bounded reviewer baselines, mutants, production probes and host health
- [TASK-261006-3ptm2x_review-attack-probes-rev4.patch](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_review-attack-probes-rev4.patch) — Disposable-copy attack probes including commit-pinned v1 context re-check
- [TASK-261006-3ptm2x_review-findings-rev4.json](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_review-findings-rev4.json) — Structured bypass finding and complete surface results
- [TASK-261006-3ptm2x_review-verdict-rev4.md](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_review-verdict-rev4.md) — Changes requested: commit-pinned v1 context reads bypass NUL guard
- [TASK-261006-3ptm2x_spawn-log_-implementer--developer--muse-_RUN-261006-c36be2.log](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_spawn-log_-implementer--developer--muse-_RUN-261006-c36be2.log) — System spawn log captured by task-board
- [TASK-261006-3ptm2x_validation-rev5.md](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_validation-rev5.md) — Rework 3 validation: commit-pin guard fix, branch sweep, production-entry tests, mutant kills, local gate evidence
- [TASK-261006-3ptm2x_change-request_rev5.patch](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_change-request_rev5.patch) — Change Request CR-TASK-261006-3ptm2x-5 revision 5 candidate patch (repository_delta=present, 39 changed paths)
- [TASK-261006-3ptm2x_change-request_rev5-validation.log](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_change-request_rev5-validation.log) — Change Request CR-TASK-261006-3ptm2x-5 revision 5 bounded validation log
- [TASK-261006-3ptm2x_spawn-log_-implementer--developer--muse-_RUN-261006-4c7a0b.log](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_spawn-log_-implementer--developer--muse-_RUN-261006-4c7a0b.log) — System spawn log captured by task-board
- [TASK-261006-3ptm2x_validation-rev5-rerun.md](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_validation-rev5-rerun.md) — Recovery-run re-verification of the rework-3 fix: locally re-ran suites and M1+M2 mutant kill
- [TASK-261006-3ptm2x_change-request_rev6.patch](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_change-request_rev6.patch) — Change Request CR-TASK-261006-3ptm2x-6 revision 6 candidate patch (repository_delta=present, 39 changed paths)
- [TASK-261006-3ptm2x_change-request_rev6-validation.log](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_change-request_rev6-validation.log) — Change Request CR-TASK-261006-3ptm2x-6 revision 6 bounded validation log
- [TASK-261006-3ptm2x_spawn-log_-implementer--developer--muse-_RUN-261006-7e87da.log](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_spawn-log_-implementer--developer--muse-_RUN-261006-7e87da.log) — System spawn log captured by task-board
- [TASK-261006-3ptm2x_validation-rev7.md](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_validation-rev7.md) — Rework-3 follow-up validation: Linux-only fixture fix, targeted local evidence
- [TASK-261006-3ptm2x_change-request_rev7.patch](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_change-request_rev7.patch) — Change Request CR-TASK-261006-3ptm2x-7 revision 7 candidate patch (repository_delta=present, 39 changed paths)
- [TASK-261006-3ptm2x_change-request_rev7-validation.log](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_change-request_rev7-validation.log) — Change Request CR-TASK-261006-3ptm2x-7 revision 7 bounded validation log
- [TASK-261006-3ptm2x_spawn-log_-reviewer--reviewer--codex-_RUN-261006-db8f4c.log](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_spawn-log_-reviewer--reviewer--codex-_RUN-261006-db8f4c.log) — System spawn log captured by task-board
- [TASK-261006-3ptm2x_review-checks-rev7.jsonl](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_review-checks-rev7.jsonl) — Bounded reviewer baselines, valid mutants, restoration results and host health
- [TASK-261006-3ptm2x_review-probes-rev7.patch](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_review-probes-rev7.patch)
- [TASK-261006-3ptm2x_review-mutations-rev7.py](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_review-mutations-rev7.py)
- [TASK-261006-3ptm2x_review-verdict-rev7.md](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_review-verdict-rev7.md) — Accepted revision 7: prior bypass fixed, full surface sweep and exact-tree hosted evidence
- [TASK-261006-3ptm2x_spawn-log_-implementer--developer--codex-_RUN-261006-457aad.log](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_spawn-log_-implementer--developer--codex-_RUN-261006-457aad.log) — System spawn log captured by task-board
- [TASK-261006-3ptm2x_integration-land.md](file://TASK-261006-3ptm2x/TASK-261006-3ptm2x_integration-land.md) — Fresh revision-7 pre-landing checks; accepted candidate unchanged; synchronous landing delegated to bound runner

## Created
2026-10-06T09:35:01Z

## Last Update
2026-10-06T19:23:21Z

## Assigned To
[implementer] developer (codex)
