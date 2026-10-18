# Review B brief: BUG-261018-dkgv1p — a help flag must never run the command

Deciding review. Check against the candidate:
- the guard sits at dispatch and covers every registered subcommand;
- the table test really enumerates the registry, not a hand-written list;
- `--` passthrough is unchanged;
- no subcommand writes anything on `-h`, `--help` or `help <sub>`;
- the mutant control fails for the intended reason (rerun it yourself);
- the full suite is green and the commit is signed.

Record a swept surface table. Verdict via `accept_cr` or changes requested with P0/P1/P2 findings; attach `BUG-261018-dkgv1p_review-verdict-rev<N>.md`; end your turn.
