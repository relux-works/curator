# BUG-261017-47xdyt validation gaps

## Refresh onto 402fdbc9

The converged README needs no edit: removing its single ownership paragraph
reproduces trunk README bytes exactly. CLIHELP and SNAPPUB text is retained,
without duplicating either addition. No repository bytes changed in this round.

Checkpoint: `402fdbc9b77803cb2d6b11f4f1cfd83961e6897b`.
Signed CI candidate: `db030e1faca9dfc0506876e14e724414ac563030` (signature `G`),
tree `191172290510ad94239590f3d4a695c80bfd738b`.
Story HEAD and real index remain at the checkpoint; candidate is uncommitted.
No `refresh-candidate` was required.

| Command / verification | Actual exit / verdict |
| --- | --- |
| README byte comparison; signed tree/index/remote-head verification | 0 each |
| `go build -o .temp/BUG-261017-47xdyt/r3-refresh/curator ./cmd/curator` | 0 |
| `golangci-lint run ./internal/install/... ./internal/marker/... ./cmd/curator/...` | 0 |
| Full local suite via bounded `go test -count=1 -json` package/name masks | 98 selected commands, all 0; 85/85 packages, 3504/3504 top-level tests driven, including all 425 CLI tests without skips |
| Seven behavioral narrowing overlays | 1 each, expected red; 7/7 killed at named assertions |
| `go test -tags snapshotcontrols -count=1 -timeout=5m -run TestDestinationNarrowingControls ./internal/envsnapshot` | 0 |
| `git diff --check` | 0 |
| Evidence-checker control | Intact fixture 0; missing final package verdict 1, expected red |
| Full hosted CI, exact signed candidate | Success: 20 executed jobs green, 2 conditional skips; all five test/race artifacts verified |

[Hosted gate](https://github.com/relux-works/curator/actions/runs/39090264021).
Exact argv, exits, complete logs, platform skips, identities and artifact hashes
are preserved in `BUG-261017-47xdyt_refresh_evidence.tar.gz`.
The board refuses binary resources (attach exit 1); the archive stays outside Git.
`BUG-261017-47xdyt_refresh_evidence.md` is the attached locator, size, digest
and structured-summary record, following that explicit CLI instruction.

One superseded batch (`suite-048`) exited **1** at the 360s process limit after
host-GOROOT lock contention; no assertion failed. It is not passing evidence.
Its four neighbours and all **36/36** activation-matrix combinations were rerun
green in bounded replacements. Initial lint PATH isolation prevented launch
(wrapper exit 1); the actual lint replacement above exited 0. No code or test
was changed for either diagnostic. Older cancelled CI on `5d8e6b07` is excluded.

**6 of 6 AC rows driven and green** within the unchanged bounds. The implementation
outcome retains the production control map, current mutant table, caller analysis,
21 ownership / 10 writer mappings and explicit out-of-contract rows. No new bound
is introduced; no catalog surface table was supplied. All 3676 prior test names
remain. This task-scoped logbook records the refresh; ready for review.
