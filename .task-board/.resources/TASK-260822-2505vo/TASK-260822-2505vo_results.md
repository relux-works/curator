# per-repository-credential-resolution — results

Tree: `.temp/TASK-260822-2505vo/worktree`, branch off `origin/main` @ `6a9b201`.
Patch: `TASK-260822-2505vo_final.patch` — 14 files, +1937/−9, applies clean to
`6a9b201` (`git apply --check` exit 0; regenerated twice byte-identical across
independent runs).

## Delivered

- Per-repository SSH credential resolution in the external install path:
  explicit flags/env selection covers every repository; otherwise the longest
  matching `build_ssh` scope for the repository's canonical identity.
- https and local-substitution repositories skip selection entirely.
- Empty selection for an ssh repository fails closed with
  `build_repository_ssh_credential_missing`.
- Pinned-agent, agent-only, and identity-only selections all reach `SSHPolicy`;
  `~` expansion for config-sourced identity/known_hosts paths.
- CI skip-class dictionary fix: five new `t.Skip` reasons classified in
  `.github/ci/skip-classes.tsv` (three reworded onto the existing vocabulary,
  two genuinely new). Without this both Windows CI lanes would fail
  `FATAL-unclassified` while local darwin gates stay green — reproduced and
  verified via the gate's native `CI_GATE_GOOS=windows` override (before: 4
  unrecognised; after: 0).

## Gates (each standalone, real exit codes)

| gate | exit |
| --- | ---: |
| gofmt -l cmd internal | 0 |
| go build ./... | 0 |
| go vet ./... | 0 |
| golangci-lint run (CI pin) | 0 |
| .github/ci/gate-selftest.sh | 0 (75 passed) |
| .github/ci/ledger-consistency.sh | 0 |
| .github/ci/no-broad-suppression.sh | 0 |
| go test ./... -count=1 (final tree) | 0 — 41 packages ok (`TASK-260822-2505vo_full-suite-final.log`) |

Finalization (suite run + artifact attachment) completed by the orchestrator
after three worker runs each ended mid-final-gate; implementation and all
evidence above are the workers' own.
