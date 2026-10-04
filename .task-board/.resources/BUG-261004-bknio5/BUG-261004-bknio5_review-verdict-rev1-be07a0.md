# BUG-261004-bknio5 — gc-sweeps-live-runtime-on-uncertain-marks

Review of CR-BUG-261004-bknio5-1 revision 1, run RUN-261004-be07a0.
Verdict: accepted; no blocking findings. This independent follow-up preserves the earlier review artifacts.

Base 934952a45953587a1d4184b692b3fb4ee401e732; candidate tree 31d96981112f3a682357b8064b28ed8ba5c1d42b.

| Reviewed surface | Result |
| --- | --- |
| Production path | Read run -> cmdGC -> collectUnderLock -> scopes.Collect, real manager locks, WriteBinShim launcher execution. |
| AC1 conservative collection | Guard precedes runtime, protected build cache and external build cache sweeps. Registry pruning remains conservative. Internal tests exercise repeated uncertainty and runtime-only API. |
| AC2 CLI regression | 3/3 uncertainty rows, each with two GC invocations and working shim before/after: truncated consumers registry, invalid marker, unreadable marker. Directory-at-marker is a real read failure; internal tests additionally cover permission failures. |
| AC3 complete-reference control | 1/1 row passes on base, candidate, and every mutant. Removes orphan runtime while retaining live executable across two passes. |
| AC4 warnings | Each uncertainty checks its source diagnostic, runtime sweep skipped, incomplete live reference set, and build cache sweep skipped. |
| Broader GC and rc.14 | Restored candidate runs scoped GC/consumer tests with exact rc.14 43bf0a2506d5c354a73bbc3ea4623d4653db10c7. All 5/5 published GC retention roots pass. Status/repair rows retain existing explicit bounds; no wider status/repair conformance claim. |
| Scope and architecture | Exactly three assigned files. Existing mark/prune/sweep structure retained. No unrelated changes or CHANGELOG.md/LOGBOOK.md edits. No reviewer changes to worktree source, index, branch or commits. |

All Go commands used GOFLAGS=-work, -count=1, and bounded timeouts for the broader/mutant checks. Base and mutant replays used a private archive of the exact candidate; only its gc.go was replaced. As gc.go is the only changed production file, the base replay is base production plus the new regressions. The archive was restored byte-for-byte before the final green run.

| Independently rerun check | Exit |
| --- | ---: |
| Base: go test ./cmd/curator -run '^TestGCPreservesRuntimeOnUncertainMarks$' -count=1 -v | 1 |
| Candidate: same CLI regression | 0 |
| Mutant 1: move Collect uncertainty guard after runtime sweep; same CLI regression, -timeout=3m | 1 |
| Mutant 2: omit registry-read uncertainty append while preserving registryUnknown; same regression | 1 |
| Additional reviewer mutant: narrow Collect guard from len > 0 to len > 1; same regression | 1 |
| Restored candidate: go test -p 2 ./internal/scopes ./cmd/curator -run '^(TestGC\|TestCollect\|TestAuthoritativeGarbageCollection\|TestNonDirectory\|TestRecording\|TestParseConsumers\|TestARepeatedRegistry\|TestWritersRefuse\|TestStructDecoding)' -count=1 -timeout=5m -v; exact rc.14 CURATOR_CONFORMANCE_ROOT | 0 |
| git diff --check BASE CANDIDATE | 0 |
| Working tree changed paths compared to candidate via git diff --exit-code | 0 |

Base red reproduces actual launcher failure (no such file or directory) after GC, not merely missing warning text. Mutants killed: 3/3. Reordering and the len > 1 narrowing break all three uncertainty rows; omitting registry uncertainty breaks the registry row only. All mutants retain a passing complete-reference control.

Hosted CR validation: https://github.com/relux-works/curator/actions/runs/37173583916 . Independently queried with gh run view --exit-status: exit 0, success, 11/11 required jobs green, two optional lanes skipped. Head 7ac0b9f3f79eb54f05e6c6374ad6bd8c977e2af3 resolves to the exact candidate tree. The attached change-request_rev1-validation.log also records exit 0. Full multi-OS tests, race, lint and build coverage are accepted from this hosted evidence; they were not all rerun locally.

Fresh origin main advertisement and exact-ref fetch both identify 7347499843dc7f0d2bc80f025d5eba4342d8060f. Upstream changes since base affect board records, CHANGELOG.md, repository admission documentation, and internal/buildrepo admission/budget tests. No candidate-path overlap. Combined-tree validation and integration remain producer responsibilities.

Local normal GC emits an existing process-inventory warning on this host; the normal runtime control still asserts deletion and live retention. Internal cache-spy tests independently verify no build sweep on uncertainty. Syspolicyd successive crashes: 402 before checks and 402 at closing observation; no increase observed. Initial incorrect service-name probe was corrected before tests. No overall host-health claim.

The run is not goal-bound (spawn goal reports none). Findings are recorded here instead of LOGBOOK.md per the binding task instruction. Acceptance must use accept_cr, routing to integrating; this reviewer does not set done or supply commit_ack.
