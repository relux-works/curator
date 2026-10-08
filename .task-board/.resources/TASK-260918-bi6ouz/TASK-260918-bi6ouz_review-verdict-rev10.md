# TASK-260918-bi6ouz — manager-dotfile-manager-table: review revision 10

Verdict: ACCEPTED. No blocking findings. Code was not modified; Go tests were not run, per instruction. Goal query reports not goal-bound; no directives pending.

Identity: base 1720a0fd8f7b12c5eede04c3e237329e612de855; candidate ac371bdc72265bc74369b9353da7bb00326b4683.
Downloaded rev10 patch SHA-256 matches f6d4e85572dc6b76f1558478caca8183ce6cbd0b2130bb8fb06aba29faaa3ef9. Stable patch ID matches exact base-to-candidate diff: b49398ea05b8382ca50dc9436b75d9e0756524cd. Hosted gate commit f75fde4de7cde92b62e1f409f296c182d97bc4b1 has exactly the candidate tree. Working files equal candidate files.

Read TASK-260918-bi6ouz_review-verdict-rev7.md: it is ACCEPTED, despite the brief describing a rejection.
Per-file stable patch-ID comparison with rev7: 5/8 paths identical: platform-cases.tsv, troubleshooting.md, managed_dotfile_conformance_test.go, takeover_test.go, fixture.
Three differences reviewed on content:
1. managed.go preserves trunk stateread inventory and errors, adds accepted boolean return, updates newer applyPlan caller, and replaces injected raw syscall/IsNotExist with typed seam Metadata and production stateread.Lstat.
2. global.go adds one-line discard of inventory boolean at newer preflight caller, preserving errors.
3. managed_dotfile_test.go adds stateread.LstatWith wrapper around five existing fault-injection subtests; assertions unchanged.
All differences are sound, scoped and tested. The seam-steer suggested dropping injection; the implementation instead injects typed seam Metadata with stateread.Lstat in production and the existing LstatWith seam in tests. This preserves useful error injection, removes the unguarded absence route and needs no allowlist.

## Surface sweep
| Surface | Evidence / result | Repeat-of mechanism |
|---|---|---|
| Table order/cells/none | Fixture patch unchanged; SHA-256 6575912115cf5b87facc0a6981e2d17bbcc1716f6ca7db3e0da5ddbf98853670; byte equality and vector table tests retained | rev7 accepted table |
| Home/XDG/platform resolution | Native injected home/env tests retained; empty/relative defaults, absolute override, USERPROFILE via production UserHomeDir; native Windows gate green | rev7 accepted resolution |
| Absence versus inspection failure | Production seam distinguishes absent/unreadable, helper retains error and scans later rows; fault assertions retained | rev8/rev9 raw function-value read guard failure fixed |
| Symlink/regular-file exclusion | Lstat and directory mode check; real UseWithPolicy symlink regression catches Stat mutant | earlier symlink-following finding resolved |
| Inventory/production notices | repairUnderLock gates on plainUnmanaged; switch retains len(taken)>0; global/applyPlan discard only new boolean; real takeover entry exercised | rev7 accepted behavior |
| Vectors | Unchanged harness patch, real UseWithPolicy, manager-name/quiet assertions; reported 6/22 macOS cases, 13 Linux and 3 Windows assigned native lanes | rev7 R1/R2 retained below |
| Docs/release | Docs patch unchanged; CHANGELOG absent; release text present under CHANGELOG entry (for release prep) in results | intended policy carry |
| Architecture/hygiene | Exact 8 paths, no stray root TASK-/BUG-, test/ledger paths; whitespace check clean; seed/writer trunk code preserved; no LOGBOOK edit | carry requirements |

## Validation and limits
Accepted attached rev10 validation log: hosted run 37757436666 success, exit 0, with Test macOS/Linux/Windows, Race macOS/Linux, Lint, Naming, Interop conformance, gate self-tests and Go driver lanes green. Exact command-shard coverage 1/1 green, failed=0, missing=0; test-case coverage unknown.
Reapply-results records successor guard test and guard mutant companion PASS, default mask 12 PASS + 1 SKIP (root unset), supplied-root mask PASS, 6/6 macOS vector subtests and 5/5 fault-injection subtests PASS, vet/format clean.
Reviewer independently ran diff/patch-ID, hash, candidate/gate tree identity, working-file identity and whitespace checks only. No tests or build rerun. Summary logs do not expose the measured reader ratio; no measured 490/490 claim is made. The exact-tree green gate and dedicated guard PASS establish the failure is resolved.

Inherited nonblocking bounds: rev7 R1 live XDG_CONFIG_HOME isolation means home-manager unreadable vector rows become absent at the production hint; quiet expectations cannot distinguish those. Fault-injection tests cover error preservation/continuation through the shared seam. Rev7 R2 isolation regression proves harness behavior rather than a production mutant; separate real takeover symlink regression provides production negative evidence.
Missing-vector skip uses the approved pre-revision-root class in live checklist/ledger, despite original unset-only AC wording. Path parsing is exercised on native OS lanes, not emulated cross-OS on POSIX.
All 12 live checklist items remain checked. Acceptance routes to integrating; work is not yet landed.
