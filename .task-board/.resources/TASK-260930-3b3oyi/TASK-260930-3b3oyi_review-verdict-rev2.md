# TASK-260930-3b3oyi — review verdict, CR rev2 (tree a540fdd6): CHANGES_REQUESTED

Probed with binaries built from the worktree (which equals the candidate delta) and from base 5ed5c4e1, on a throwaway HOME.

## Verified OK
1. Clean HOME, \`curator global init\` -> exit 1, stderr:
   \`curator: global config not found: <HOME>/.curator/config.json\` then
   \`curator: run \`curator bootstrap --skills-root <dir>\` first (see docs/cli.md#bootstrap)\`.
   \`global --help\`, \`env --help\` -> exit 0 without config. \`env status\` on a clean HOME gets the same hint.
   Malformed config gets no hint (ErrConfigNotFound is an absence-only sentinel), which is correct.
2. The documented example \`curator bootstrap --if-missing --non-interactive --skills-root "$HOME/skills"\` -> exit 0 and \`wrote .../config.json\`.
   The old form -> exit 2, \`curator: bootstrap requires --skills-root\`. \`global init\` after bootstrap -> exit 0.
3. env status --check (internal/envprofile/status.go): only profiles current in some scope decide NonCurrent; other rows are still printed. The row asserts both exit codes.
4. Resolve diagnostic is \`environment_home_stale: home unprovisioned; rerun with --repair\`. The exit code is unchanged (1), and the --repair --dry-run path carries no hint.
5. Warning prints once on a single command.
6. GOOS=windows go vet ./cmd/curator ./internal/... exit 0. No CHANGELOG/LOGBOOK change, no stray files.

## Findings (changes requested)
F1 (blocking, review-note item 5): the posture-suppression marker is honoured from ANY environment, so it is not "only honoured when set by curator itself".
  cmd/curator/main.go:354 \`os.Getenv(postureWarnedEnv) == "1"\`.
  Production probe: \`CURATOR_INTERNAL_POSTURE_WARNED=1 curator env status\` -> exit 0, stderr EMPTY (unmarked: the warning prints). Any user, wrapper, CI env file or inherited env hides the warning for a user-started command.
  cmd/curator/firstrun_ux_test.go:228 pins this abuse as intended behaviour (\`t.Setenv(postureWarnedEnv, "1")\` then asserts 0 warnings for a directly started command). The row proves the suppression is reachable. It does not prove the marker is unforgeable.
  Fix shape: bind the marker to the dispatching parent, e.g. export \`postureWarnedEnv=<parent pid>\` from childEnviron and honour it only when it equals os.Getppid() (a pid-less "1" must not suppress). Alternatively a per-tree random nonce is not verifiable, so prefer the ppid binding. Then replace the test at :228 with negative rows:
  (a) marker "1" set by the user -> warning still prints;
  (b) marker with a pid that is not the parent -> prints;
  (c) real \`curator run ...\` nested -> exactly one warning (keep the existing row).
  Mutant to kill: dropping the Getppid comparison (honouring any non-empty value) must fail (a)/(b). Also say in a comment that the ancestor-chain grandchild case (curator run -> shell -> curator) either warns again or is a stated bound.

F2 (regression, medium): wantsHelp at cmd/curator/main.go:376 hijacks per-subcommand help for every group, even WITH a config.
  Base: \`curator env resolve --help\` (config present) prints \`Usage of env resolve:\` plus the flag list (-dry-run, -format, -profile, -repair, -takeover) and exits 2. Same for \`global add --help\` (-branch, -git, -revision...).
  Candidate: both print only the group usage line (\`usage: curator env <resolve|status|...> [flags]\`) and exit 0. The flag documentation is lost for all env/global/profile subcommands, and the exit code changes silently.
  Fix: answer the group-level help (\`global --help\`, \`help\`) without config, but let \`<group> <sub> --help\` reach the subcommand FlagSet. If the subcommand help must also work without a config, the flag sets must be buildable before loadConfig; otherwise keep the bootstrap hint on the refusal and keep flag help behind config. Add rows: \`env resolve --help\` with a config prints \`-repair\`; without a config it either prints the same or the hint.
  The rows at firstrun_ux_test.go:61-62 (\`global init --help\`, \`env status -h\`) pin only the generic usage line, so they would accept the regression.

## Evidence limits
\`go test ./cmd/curator -run 'Bootstrap|GlobalInit|EnvStatus|EnvResolve|Posture'\` was started locally and failed after 301 s with \`acquire package host GOROOT test lock: context deadline exceeded\` (known host lock contention; see local-gate host limits). I did not get a local test result and make no claim from it. Hosted gate green is accepted from the orchestrator. The probes above ran the real binary.