# Review note for TASK-260922-1s0zja revision 2 (orchestrator, binding) — F-M1a rework

Revision 2 = rework 1 for your revision-1 verdict (`TASK-260922-1s0zja_review-verdict-rev1.md`;
brief `1s0zja-rework-1.md`). Scope of the delta ONLY:
1. **F1 (your P1)**: the plugin-scoped count plus the "agy is non-empty" assertion is replaced by a
   module-wide closed allowlist — `TestTheBypassFlagIsSpelledAtExactlyTwoKnownSites` requires
   `argvguard.LiteralSites` over all non-test module sources to report EXACTLY the claude const
   (`pkg/agentic/systems/claude/args.go` / `bypassPermissionsFlag`) and agy's construction site
   (`pkg/agentic/systems/agy/args.go` / `Args`), with a committed narrowing
   (`TestTheTwoSiteProofBitesOnAThirdSpelling`) carrying your verbatim reproduction plus a
   second-site-inside-agy case.
2. **F2 (your P2)**: CHANGELOG/README corrected so `ErrPermissionModeDuplicate` is attributed to the
   direct plugin `Argv` path and `ErrCompositionNotInteractive` to `BuildPlan`; the README
   interactive paragraph now acknowledges the optional typed member.
3. The overclaim about "no plugin surface" is narrowed to what the tests measure.

Everything you judged correct at revision 1 stands as committed — do not re-open the member,
`Resolve`, sentinels, mapping order, pi/pinative refusal, the `Composition.Prefix` duplicate scope
or the F-M1b/F-L1 boundaries.

Verify:
- run your OWN rev1 third-site reproduction again (plant
  `pkg/agentic/review_third_spelling.go`): the guard must now FAIL, and removing it must restore
  green;
- plant a SECOND site inside the claude plugin and inside agy: both must fail;
- make a known site go silent (rename the const's literal): that must fail too (the allowlist is
  closed in both directions);
- confirm the codex flag's proof shape and say whether it is module-wide or plugin-scoped;
- confirm the documents now match production by reading the code paths, not the prose;
- re-run the narrow packages with real exit codes; the board validation command ran once at handoff
  (reuse that evidence).

Record exactly one verdict: `accept_cr(TASK-260922-1s0zja, revision=2, evidence=<your outcome
resource>)` on ACCEPT, or a changes-requested verdict routed with `set_status` naming file:line and
an executable reproduction. Do not write into any control root's LOGBOOK.md.
