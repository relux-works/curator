# Review note for TASK-260907-187z6x revision 1 (orchestrator, binding) — git same-source reinstall

Read the brief `187z6x-brief.md` (precondition), the producer's `TASK-260907-187z6x_results.md`, and
the source finding (stage (c) review cycle 6, C6-m1 = repeat of C5-M1; the related open bug is
BUG-260916-3aco9f / issue #73, which is blocked on this leaf). Reuse the hosted gate evidence for the
exact candidate tree; do not rerun the full landing suite.

The fix gives the git same-source reinstall branch of `installLocked` the SAME activation seams the
path reinstall already uses (`reinstallActivation` → `activateReinstall` → `useLocked` + scoped
resync), reported as an update exactly as before. The producer cites environments §9.1 (activation
applies to every install-row invocation; the reinstall sentence governs resolution/reporting only),
cli/curator.md's install row, and §9.5 (takeover carried by `profile install`) as the reason parity —
not a refusal — is correct. **Check those citations against the pinned spec yourself**
(`SPEC_PIN` in `.github/workflows/ci.yml`, spec checkout at
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec`): if any of them actually carves the
reinstall out, the chosen behaviour is a finding.

Judge:
1. **The eight rows** (four flag combinations × current / not-current) drive `run()` through the
   insteadOf fixture (no `file://`). Verify they are real production-entry rows, that the asserted
   operator lines and activation outcomes match what the code does, and that row 2
   (`--use` without `--takeover` over an unmanaged blocker ⇒ exit 1,
   `environment_surface_unmanaged_conflict` + `profile_use_partial`, nothing written, no backup) is
   the right semantics rather than a regression: a plain `--use` on a reinstall that previously
   exited 0 now exits 1 when the surface is blocked. Decide whether that is the §8.3/§9.5 contract
   (it should be — the same as a first install) and whether any caller relied on the old silence.
2. **No path-root behaviour change** — prove it (the path sibling test plus a diff read of
   `reinstallPathLocked`/`reinstallActivation`; only the doc comment was to be generalized).
3. **The mutant**: re-run the producer's (drop the activation call) and confirm both named rows fail
   while `takeover-without-use` still passes; add one of your own (e.g. activate unconditionally,
   ignoring `--use`, or pass the takeover flag but skip the resync).
4. **Takeover safety**: row 4 claims backup generation 1 holds the operator bytes. Verify the backup
   is a §8.3.1 write (directory entry replaced after backup; the link target never opened for
   writing) and that the notice is printed.
5. Ledger rows if the new test needs them; CHANGELOG `Fixed`; issue #73 stays open (the orchestrator
   closes it with the landed commit through BUG-260916-3aco9f).

Record exactly one verdict: `accept_cr(TASK-260907-187z6x, revision=1, evidence=<your outcome
resource>)` on ACCEPT, or a changes-requested verdict routed with `set_status` naming file:line and
an executable reproduction. Do not write into the control root's LOGBOOK.md.
