# Review note for TASK-260922-1wvwc3 revision 1 (orchestrator, binding) — F-M1b capability table

Control root: skill-agents-management (shared with the operator's own session). Read the brief
`1wvwc3-brief.md`, the producer's `TASK-260922-1wvwc3_results.md`, and Decision 0018 choices 3 and 6
(`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/decisions/0018-curator-run-permission-interface.md`).
F-M1a (the member + mapping + two-site guard) is already on main (`fcbaa74f`) — do not re-review it.
The board validation command ran at handoff with exit 0; reuse it, run narrow packages yourself.

What was built: token `permission-grammar-v1`; per-plugin `verifiedReleases` rows read only through
`agentic.LookupReleaseCapability` (no central table, by the single-source guard); new
`LaunchRequest.ToolRelease` and `LaunchRequest.NativeArgs`; sentinels
`ErrPermissionModeUnverifiedRelease`, `ErrNativePolicyUnknown`, `ErrToolReleaseUndetected`,
`ErrNativeArgsNotInteractive`; the parsing rule in `internal/nativeargs`; per-plugin `--version`
probes in `internal/toolprobe`; goldens and narrowings; proposed release v0.5.17.

Judge:
1. **Drift rule** exactly as choice 6: yolo on an unverified/newer/empty release is refused FIRST
   with a named diagnostic that cites the token; `native` forwards verbatim at any release with no
   claim. The producer notes installed claude `2.1.274` and codex `0.153.4` are NEWER than the pins
   (`2.1.261`, `0.153.2`), so yolo is refused on this machine today — confirm that is correct
   behaviour, and say whether the table should gain verified rows for the installed releases (that
   needs real `--help` evidence per release, not an assumption) — a follow-up, not a blocker, unless
   the pins themselves are wrong.
2. **Choice 3**: unknown native policy forms (new codex `-c` keys incl. `=`/attached/dangling forms,
   unknown claude `--permission-mode` values) refuse as usage and are never resolved into a claim.
   Attack the parsing rule: prompt text that looks like a flag, `--` placement, a lone `-`, an
   `=`-value containing `=`, and a `-c` key that differs only by case.
3. **Probes**: `--version` runs with an explicit argv, a context timeout, an output cap and a handed
   (never ambient) environment; probe failure yields `ErrToolReleaseUndetected` and the caller passes
   `""` (so yolo refuses, native passes). No real provider session is launched in tests.
4. **Scope**: `NativeArgs` is refused outside interactive; bypass lands BEFORE the verbatim suffix
   (0018 item 1); the item-4 conflict table and the item-5 headless detector are explicitly left to
   later leaves (F-L1). Check nothing in the launcher/curator/ax/task-board was touched.
5. Narrowing mutants executed and killed (re-run at least two yourself, add one of your own), README
   and CHANGELOG release-ready for v0.5.17 with the token named. No tag in this leaf.

Record exactly one verdict: `accept_cr(TASK-260922-1wvwc3, revision=1, evidence=<your outcome
resource>)` on ACCEPT, or changes-requested with file:line and reproduction. No LOGBOOK.md writes.
