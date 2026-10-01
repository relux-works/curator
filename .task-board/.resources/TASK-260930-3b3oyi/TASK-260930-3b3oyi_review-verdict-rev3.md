# TASK-260930-3b3oyi rev3 review verdict: ACCEPTED

Candidate tree 45c81c5a (worktree tree verified identical via temp-index write-tree), base 5ed5c4e1. Probes ran on a throwaway HOME with binaries built from the candidate and from base.

## F1 posture marker
- Marker honoured only when \`os.Getenv(CURATOR_INTERNAL_POSTURE_WARNED) == strconv.Itoa(os.Getppid())\` and ppid>1 (main.go loadConfig). The parent exports its own pid via \`childEnviron()\` at the single dispatch site (umbrella.go:723), and any inherited marker is stripped first.
- Production-binary probes, all exit 0:
  - \`env status\`: 1 warning.
  - \`CURATOR_INTERNAL_POSTURE_WARNED=1 env status\`: 1 warning.
  - \`...=99999 env status\`: 1 warning.
  - \`curator run env status\`: exactly 1 warning (base binary: 2).
- Rows (a) user-set "1", (b) not-parent pid and (c) real nested \`curator run\` all exist and pass.
- Mutant: replace the Getppid comparison with \`!= ""\`. TestPostureWarningOncePerProcessTree/user-set-one and /not-parent-pid FAIL, exit 1. Killed.
- The grandchild case (run -> shell -> curator) is commented at main.go:206: it warns again, a stated bound.
- Residual, accepted: a user who deliberately exports the pid of their own shell (\`CURATOR_INTERNAL_POSTURE_WARNED=$$\`) suppresses the warning for that shell's direct children. This is a deliberate self-suppression, and the same user can edit the config. A default or accidental value ("1", a stale pid) never suppresses.

## F2 subcommand help
- With a config present, \`env resolve --help\` and \`global add --help\` give exit 2. Stderr is byte-identical to base 5ed5c4e1 (md5 equal for both).
- Group help (\`global|env|profile\` x \`--help|-h|help\`) prints the usage text and exits 0 without config.
- Subcommand help without config keeps the refusal plus the hint, exit 1.
- Rows TestGlobalInitSubcommandHelpWithoutConfigRefuses and TestGlobalInitEnvResolveSubcommandHelpPreservesFlagSet assert the exact FlagSet usage bytes, so the hijack would be caught.

## Items 1-4
1. Clean HOME, \`global init\`: stderr \`curator: global config not found: <path>\` then \`curator: run \`curator bootstrap --skills-root <dir>\` first (see docs/cli.md#bootstrap)\`, exit 1. A malformed config gets no hint (test row). docs/cli.md carries the \`bootstrap\` anchor and the prerequisite text.
2. The docs example \`curator bootstrap --if-missing --non-interactive --skills-root "$HOME/skills"\` exits 0 and writes config. The old form exits 2 \`bootstrap requires --skills-root\` (TestBootstrapDocsExampleRuns executes the doc line).
3. \`env status --check\`: only profiles current in some scope decide NonCurrent. The test shows exit 1 with one current-scope unprovisioned home, exit 0 once repaired, and the non-current \`default codex_cli\` row still printed.
4. \`env resolve\` without --repair: \`environment_home_stale: home unprovisioned; rerun with --repair\`, exit 1 unchanged. \`--repair --dry-run\` carries no hint.

## Windows fix / hygiene
The posture test builds a native curator-run (.exe on windows) with no shell, and nothing is skipped. \`GOOS=windows go vet ./cmd/curator ./internal/envprofile ./internal/config\` exit 0. Changed paths are 9: no CHANGELOG/LOGBOOK, no stray files, no employer names.

## Tests (own runs, -p 1)
- Targeted rows (TestGlobalInit*, TestBootstrapDocsExampleRuns, TestEnvResolveStaleHintsRepair, TestEnvStatusCheckCurrentScopeOnly, TestPostureWarningOncePerProcessTree): PASS, exit 0.
- \`go test ./internal/envprofile -run 'Status|Resolve|Stale|Unprovisioned'\`: ok, exit 0. \`./internal/config\`: ok.
- NOT confirmed locally: the broad \`-run 'Bootstrap|GlobalInit|EnvStatus|EnvResolve|Posture'\` on ./cmd/curator and the full ./internal/envprofile package both hit the default 10-minute go test timeout under host load (exit 1, ~601 s). I read that as load, not a test failure, but the output did not show the timeout panic header, so it is not proven. The hosted gate (green on this CR) is the arbiter for those.