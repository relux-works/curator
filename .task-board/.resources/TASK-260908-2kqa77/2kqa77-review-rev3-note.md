# Review note for TASK-260908-2kqa77 revision 3 (orchestrator, binding) — goreleaser config value gate

Revision 1 was NOT reviewed: its gate failed on Linux because the gate's awk carried
`[+-0-9]`, a mid-class range gawk rejects (BSD awk tolerates it, which is why the producer's macOS
run was green). Revision 2 fixed the class to hyphen-last `[+0-9-]` and went green on all three
runners (run 35725881034); revision 3 adds the portability PIN that rework item 4 asked for and is
green again (run 35729614942). So this is the first review of the gate's content, and the delta to
judge is the whole leaf.

Read the brief `2kqa77-brief.md`, the rework brief `2kqa77-rework-1.md`, the producer's
`TASK-260908-2kqa77_results.md` (its rev3 section carries the bracket-expression audit and the
divergence scan), and the source finding:
`.task-board/.resources/TASK-260908-1jv1h3/TASK-260908-1jv1h3_review-verdict-3.md` §5 (goreleaser
check discriminates 0 of 6 wrong-value mutants; no in-repo gate reads `.goreleaser.yml`; rows C and
E are the named negatives). Reuse the hosted gate evidence for the exact candidate tree; do not
rerun the full landing suite.

Specific to this revision, judge also:
- the two new self-test rows are SOURCE-level greps (`grep -qF '[+-0-9]'` must fail the gate;
  `assert_contains '[+0-9-]'`). A grep pin is weaker than a behavioural one — decide whether it is
  the right instrument here (the producer's argument: BSD awk accepts the bad class, so behaviour
  cannot pin it on macOS). Attack it: does the pin survive an equivalent regression spelled
  differently (`[+-9]`, `[a-+]`, a range introduced in a DIFFERENT bracket expression)? If a
  realistic regression escapes both rows, say so and propose the bound.
- the producer executed only BSD awk locally (gawk and Git-Bash awk unavailable on the host) and
  leans on the hosted lanes. That is an acceptable bound only if the hosted self-test really would
  fail on a bad class — confirm the gate is invoked on all three self-test runners.

The producer wrote a POSIX `sh`+`awk` structural gate (`.github/ci/goreleaser-config-gate.sh`) wired
into the `lint` job and pinned by `gate-selftest.sh`. Judge:

1. **Scope mapping.** The AC said `brews[*].skip_upload`; the config has no `brews` stanza, so the
   producer checks `homebrew_casks[*]`, `scoops[*]` and `release.prerelease`. Verify against the
   committed `.goreleaser.yml` that those ARE the publishing surfaces that decide whether an rc
   reaches an install channel, and that nothing publishing was left unchecked (read the file, not
   the summary). A missed stanza is a finding.
2. **Parse, not grep.** Attack the parser yourself — at minimum: `"Auto"` (row E), the key absent
   (row C), `true`, `ato`, a key nested one level deeper under `repository:`, a second entry added
   with a bad value, tab indentation, a duplicate key, a block scalar containing the word
   `skip_upload`, and a value with a trailing comment. Each must fail or pass as claimed; find one
   shape the walk mis-parses and it is a finding. Also check it is genuinely POSIX awk (it must run
   under BSD awk and Git Bash — the self-test runs on all three runners).
3. **Wiring.** The gate must actually run on every push/PR (not only in a job that can be skipped),
   and `gate-selftest.sh` must pin the exact invocation so the wiring cannot silently disappear —
   the Windows-budget leaf (TASK-260919-3ux95w) previously left a line unpinned and a mutant
   survived. Attack the pin: delete the ci.yml step and the self-test must fail; change the gate's
   path and the self-test must fail.
4. **Fail-closed behaviour**: unreadable file, missing stanza, empty stanza — all must exit non-zero
   with a message naming field, entry and observed value.
5. CHANGELOG entry present; `.goreleaser.yml`, `release.yml` and the ancestry gate unchanged.

Bounds you may accept: no `goreleaser` binary runs here (the cycle-3 review already drove the real
publish path); the gate proves values, not upstream semantics.

Record exactly one verdict: `accept_cr(TASK-260908-2kqa77, revision=3, evidence=<your outcome
resource>)` on ACCEPT, or a changes-requested verdict routed with `set_status` naming file:line and
an executable reproduction. Do not write into the control root's LOGBOOK.md.
