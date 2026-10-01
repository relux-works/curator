# TASK-260930-3b3oyi — rework 1 (THE ONLY CURRENT INSTRUCTION, with secondfix-brief.md)

The rev2 review (`TASK-260930-3b3oyi_review-verdict-rev2.md`) verified items 1–4 and the Windows fix. Two findings remain:

F1 (blocking): the posture-warning suppression marker is honoured from ANY environment (cmd/curator/main.go:354). With
`CURATOR_INTERNAL_POSTURE_WARNED=1 curator env status`, the warning disappears.
- Bind the marker to the dispatching parent: the parent exports `CURATOR_INTERNAL_POSTURE_WARNED=<own pid>`, and the child honours it
  only when the value equals os.Getppid(). A pid-less "1" never suppresses.
- Replace the test at firstrun_ux_test.go:228 with three rows:
  - (a) a user-set "1" → the warning prints;
  - (b) a pid that is not the parent → prints;
  - (c) a real nested `curator run` → exactly one warning.
- State the grandchild case (curator run → shell → curator) in a comment: it either warns again or is a stated bound.
- Mutant: drop the Getppid comparison → (a)/(b) must fail. Give real exit codes.

F2 (regression): `wantsHelp` (main.go:376) hijacks per-subcommand help. `curator env resolve --help` and `global add --help`, with a
config present, used to print the FlagSet usage and exit 2. Now they print only the group usage and exit 0.
- Answer group-level help (`<group> --help`, `help`) without config.
- `<group> <sub> --help` must reach the subcommand FlagSet unchanged when a config exists. Without a config, keep the bootstrap-hint
  refusal, unless the flag sets can be built before loadConfig.
- Add rows asserting the exact base behaviour for `env resolve --help` and `global add --help` with a config.
- Fix the rows at firstrun_ux_test.go:61-62 so that they would catch this regression.

Use `go test -p 1` locally; the host is loaded. The hosted gate is the arbiter. Set status development, update the results, run
`task-board handoff TASK-260930-3b3oyi --role developer`, then END YOUR TURN. No CHANGELOG/LOGBOOK edit.
