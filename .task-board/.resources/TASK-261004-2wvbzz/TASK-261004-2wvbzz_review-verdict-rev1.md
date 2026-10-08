# Review verdict, TASK-261004-2wvbzz rev1: ACCEPTED

Scope: CR-TASK-261004-2wvbzz-1 rev1, one added file `.research/261004_inline-audit-wave-2.md` (147 lines), delta e5489b6b..e361af58. Read-only review; no builds (R193/R194); evidence archives not needed and not opened. The worktree copy of the record is byte-identical to the candidate tree blob.

## Checks (reading the record only)

| Check | Result |
|---|---|
| Each finding has severity, location, evidence | N6 medium (gitcred.go bounded buffer / Access.call), N7 medium (ownedTarget vs StageForwarding / publication recheck), N8 low (main.go pin flag / trustDir / Normalize), N9 low (spec manager.md pin contract / Pin). Each carries a named probe, a result table, a Go exit code and a command-ledger reference. PASS |
| Counts match | 4 findings in title, intro and closing; ledger rows 01..13 = 13 commands as stated; N6 "6/6: 4 pass, 2 red" matches the table (3 real-Git + 3 controlled rows, reds at 65,664 and 65,537); global ownership "4/4: 3 pass, 1 red" matches; npm "8/8" matches the eight listed cases; pins "10/10 rows = 20/20 decisions" consistent. Exit codes quoted in findings (N6 cmds 01/07 = 1, N7 cmd 07/04 = 1, N8/N9 cmds 02/07/12 = 1) agree with the ledger. Not independently re-run: the 11 probe digests and 9,641/153 byte-comparison figures (they live in the evidence bundle). PASS on internal consistency |
| Public-record rule (sensitive findings only as counts/classes) | The four findings are a truncated-answer integrity defect, a legacy-shim reconcile failure, an operator-supplied-flag validation gap and a missing provenance field. None discloses a remotely exploitable chain, a live secret, a bypass recipe or a vulnerable deployment; N8 is explicitly bounded as operator-supplied with fixed filename and no package-controlled route, N6 states no authentication with the truncated value was attempted. Same class and detail level as wave 1 (docs/security-audit-2026-10-inline.md N1-N3), already public. Reproduction fixtures are synthetic (length-only credential assertions, `wave2-outside-audit` placeholder). Unproven surfaces (broker pipe cap, pre-admission read bounds) are stated as unknown, not as exploit detail. PASS: no line names a sensitive class that needs withholding |
| No secrets | grep for key/token/PEM patterns: none. PASS |
| No personal paths, addresses, employer/client names | grep for home-directory, temp-dir, user-name, e-mail patterns: none; the text states logs redact temp/personal paths. The only hosts named are the project's own public repositories (curator, curator-spec). PASS |
| No product code changed | Delta is exactly one file under .research. PASS |
| Whitespace | `git diff --check` over the delta: exit 0. PASS |

## Non-blocking observations (no rework required)

- Report date reads 2026-10-04 while review is on 2026-10-08; harmless.
- N7 and N8/N9 fix recommendations are guidance only; follow-up implementation tasks should carry them, with the attached probes as the regression slice (expected-red tests must not be used as a green gate).
- Coverage bounds (2/10 npm limit fields touched, no Windows dynamic cases, broker pipe cap unknown) are stated honestly and should be carried into the follow-up planning.

Verdict: ACCEPTED, route to integrating. No `commit_ack` supplied.
