# Review note for TASK-260922-1s0zja revision 1 (orchestrator, binding) — F-M1a permission-mode member

Control root: skill-agents-management (shared with the operator's own session — never write outside
the assigned Story worktree/disposable copies). Read the brief `1s0zja-brief.md` (precondition), the
producer's `TASK-260922-1s0zja_results.md`, and the owning decision:
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/decisions/0018-curator-run-permission-interface.md`
(choices 1, 3, 6 and the Compatibility section — agents-management owns the flag spelling; the
launcher only resolves and passes the mode, Decision 0013 D5). The board validation command
(`go build ./... && go test ./... -count=1`) ran at handoff with exit 0 — reuse that evidence, do not
rerun the whole suite; run the narrow packages yourself.

Judge, each on its own:
1. **Member + validation.** `PermissionMode` (`native`/`yolo`, zero = native), `LaunchRequest.
   PermissionMode`, `Resolve()` as the single value reader, `BuildPlan` refusing unknown values and
   any non-zero value outside `LaunchModeInteractive` BEFORE any plugin surface is touched. Verify
   the pre-existing `native`/zero-value goldens are byte-identical (no argv drift anywhere).
2. **Mapping, spelled once.** claude `--dangerously-skip-permissions`, codex
   `--dangerously-bypass-approvals-and-sandbox`, emitted exactly once, after model+effort and before
   any prompt text; `pi`/`pinative` refuse with `ErrPermissionModeUnsupported`. The producer's pi
   evidence is `pi --help` at 0.84.2 showing only `--approve, -a` — check that claim yourself if the
   binary is available, and say so if it is not. Judge the claimed release pins (claude 2.1.261,
   codex 0.153.2) against what the repository actually pins.
3. **The argvguard bound the producer scoped.** Module-wide single-site is claimed IMPOSSIBLE for
   `--dangerously-skip-permissions` because agy's golden-pinned exec spelling shares it, so the
   proof is per-plugin plus `TestTheAgySharingIsStillTheReasonForScoping`. Decide whether that is an
   acceptable bound or a finding: does anything now prevent a THIRD spelling site appearing outside
   the claude plugin and agy? The new `internal/argvguard.LiteralSites` counter is the tool — attack
   it (plant a second literal in a third package and see whether any committed test fails).
4. **Mutants.** Re-run at least two of the producer's six (one value/scope bound, one duplicate or
   unsupported bound) in a disposable copy and confirm the named test fails; then attack with one of
   your own — e.g. emit the flag before the model flags, or emit it twice.
5. **Recorded decisions to judge, not to assume.** Explicit `native` outside interactive is refused
   (`ErrPermissionModeNotInteractive`); the duplicate check reads only `Composition.Prefix` with
   exact-element match (`=`-forms deferred to F-M1b/F-L1); no `NativeArgs` member. Each is either
   defensible under 0018 or a finding — say which and why.

Bounds to state explicitly: the (environment, tool release) capability table, the drift rule and the
parsing rule are F-M1b (TASK-260922-1wvwc3) and must NOT be treated as missing here; the
launcher-side end-to-end "yolo + raw bypass after `--`" is F-L1. No release tag in this leaf.

Record exactly one verdict: `accept_cr(TASK-260922-1s0zja, revision=1, evidence=<your outcome
resource>)` on ACCEPT, or a changes-requested verdict routed with `set_status` naming file:line and
an executable reproduction. Do not write into any control root's LOGBOOK.md.
