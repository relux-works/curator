# BUG-260921-30ycv0 review verdict — revision 1 (reviewer, claude-opus-5)

Change Request `CR-BUG-260921-30ycv0-1`, base `98f8e633`, candidate tree
`d1d86ec2076d5dffcda95e98cb2d3a2be6e2d056`, patch sha256
`ad9f4bd1d7bbbc984302215cd9486f84300eaea9baaeb1a3a443f864d7ece011` (verified).

## Verdict: ACCEPT

## 0. Tree identity

- Story worktree temp-index `write-tree` = `d1d86ec2…` (untracked test files included).
- Disposable clone `/tmp/30ycv0-review/cand` (`--no-checkout` clone of the control root,
  `checkout --detach 98f8e633`, `git apply --index <rev1 patch>`, commit) → `HEAD^{tree}` =
  `d1d86ec2…`; still `d1d86ec2…` with 0 dirty paths after every rerun.
- Hosted gate run 35545383544: `headSha 3e74823a` → `tree d1d86ec2…`, parent `98f8e633`
  (exact candidate). Jobs: Lint, Naming gate, Interop conformance gate, Gate self-test ×3,
  Test (ubuntu/macos/windows), Race (ubuntu/macos) all `success`; Candidate suite and
  rose-air skipped by configuration.
- Delta: 4 paths — `CHANGELOG.md` (+5), `internal/gitops/gitops.go` (1 line),
  `internal/gitops/writeblobs_spawn_test.go` (new), `internal/install/writeblobs_spawn_test.go`
  (new). Non-test product change is exactly `gitops.go:402` (`return err` →
  `fmt.Errorf("git cat-file --batch failed in %s: %w", repo, err)`).

## 1. Production-entry before/after (my own probe, not the producer's)

Throwaway `TestZZProbeResultShape` (untracked, in the mutant clone only) reproduces the
committed install row's scenario (`install.Project` over a v1 git source; PATH shim `git`
0755 whose shebang interpreter is swapped for a 0644 file after serving `ls-tree`) and dumps
the whole `install.Result` as JSON. Diff candidate vs mutant M1 (bare return), temp paths
normalized:

```
7c7
<     "git cat-file --batch failed in <T>/001/skill-a: fork/exec <T>/004/git: permission denied"
---
>     "fork/exec <T>/004/git: permission denied"
```

Every other field (`Alias`, `Path`, `Status: failed`, `Messages: null`, `Builds`,
`BuildsComplete`, `Staged`, `BuildDiagnostic`, `BuildCacheRetained`, `Attestations`) is
identical → nothing else in `Result` changed.

## 2. Sanitization and error chain (`gitops.go` read in full around the change)

- New text = `git cat-file --batch failed in <repo>: <cmd.Start() error>`. The `<repo>`
  operand is already reported by the same function's `Wait` path (`gitops.go:471`), its
  abort path (`:426`) and by `listTree` (`:243`, `git ls-tree failed in %s`); the
  `fork/exec <git path>: permission denied` suffix is exactly the text that previously
  surfaced bare. No environment, no new path class; `gitops.run` (`:51`,
  `git %s failed: %s`) reports argv + stderr and nothing more sensitive.
- `%w` keeps the chain: the in-package row asserts `errors.Is(err, syscall.EACCES)`; the
  previous bare `*os.PathError` matched the same predicates, so caller-visible classification
  is unchanged (and no caller inspects it: `closure.go:446`, `contextstore.go:76`,
  `snapshot.go:105` return the error as-is; `snapshot.go:65` replaces it with a static
  `source_snapshot_unavailable` text; `envprofile/gitsource.go:244,380` wrap with `%v`;
  `install.go:412` renders `closure.Build` errors with `failf("%v")`, no redaction, no
  `errors.Is/As` on this path).
- `Wait` errors were already wrapped in the same shape (`:471`); the batch-protocol paths
  (short header read `:426`, malformed header/size, missing object, mid-stream I/O) are
  byte-identical (diff shows the single line).
- `StdinPipe`/`StdoutPipe` errors (`:395`, `:399`) stay bare — pre-spawn fd setup, not the
  fork/exec spawn this bug covers; producer records it as a bound, I concur.
- No retry anywhere in the product delta (one `fmt.Errorf`).
- R4: `internal/gitops/gitops.go` has no logging/diagnostic seam (`grep` for log/Logf/
  Printf/slog/Diag/Hook/func vars: none) — recorded bound, no product diagnostics added.

## 3. Independent reruns (candidate clone, `#!/bin/bash` driver, `set -o pipefail`, real rc)

| Step | Result |
|---|---|
| `go build ./...` | rc=0 |
| `go vet ./internal/gitops/ ./internal/install/` | rc=0 |
| `gofmt -l internal/gitops internal/install` | 0 lines |
| `go test ./internal/gitops/ -count=1 -v` | ok 11.9 s, 26/26 PASS (all `TestExtract*` success/refusal rows, `TestExtractReproducesByteExactVector`, deadlock/termination rows, new `TestWriteBlobsSpawnFailureIsWrapped` 0.02 s) |
| `go test ./internal/install/ -run '^TestProjectWriteBlobsSpawnFailureIsWrapped$' -count=1 -v` | PASS 1.21 s, ok |
| `go test ./internal/install/ -run '^(TestLegacyInstallUntouchedWhenDraftOff\|TestLegacyDeclaredTagBumpIsNotAMovedTag\|TestEndToEndInstall\|TestGitFixture.*\|TestDryRunEffectBindingsSeeWhatARealOperationWrites)$'` with `CURATOR_CONFORMANCE_ROOT` set | ok 51.3 s, 15/15 PASS (incl. the gate-flake test from 3vfwch, 23.3 s) |
| `go test ./internal/closure/ ./internal/snapshot/ ./internal/contextstore/ ./internal/envprofile/ -count=1` | ok 124 s / 20.6 s / 3.3 s / 459 s |
| `golangci-lint run ./internal/gitops/... ./internal/install/...` (local) | 0 issues, rc=0 (first attempt killed by the host stall watchdog, rc=137; rerun after the window cleared) |

Hosted per-lane evidence (`test-evidence-<os>` artifacts, `test/go-test.json` +
`observed-cases.tsv` + `skips-observed.tsv`):

| Lane | `internal/gitops TestWriteBlobsSpawnFailureIsWrapped` | `internal/install TestProjectWriteBlobsSpawnFailureIsWrapped` |
|---|---|---|
| ubuntu-latest | pass | pass (0.03 s) |
| macos-latest | pass | pass (0.21 s) |
| windows-latest | skip, class `platform-control` / `allowed-platform-control`, reason `executable-bit spawn refusal is exercised on the unix runners` | same |

The Windows skip reason is byte-identical to the 3vfwch sibling rows
(`internal/install/gitfixture_retry_test.go:228,305`, `internal/install/atomicity/…`) and
matches ledger row `platform-control (is|are) exercised (on|by)` in `.github/ci/skip-classes.tsv:55`.

## 4. Mutant table (mutant clone `/tmp/30ycv0-review/mut`, one line each, restored with
`git checkout --` between mutants; the two committed rows run per mutant with `-count=1`)

| # | Mutation (`gitops.go`) | gitops row | install row | Reading |
|---|---|---|---|---|
| M0 | none (candidate) | PASS | PASS | baseline |
| M1 | `:402` → `return err` (bare, the gate-observed shape) | FAIL `err = "fork/exec …/git: permission denied", want the wrapped operation context` | FAIL `errors = "fork/exec …/git: permission denied", want the wrapped cat-file operation context` | killed; proves the install row reads the `Start` error through `install.Project` |
| M2 | `%w` → `%v` (chain broken) | FAIL (`errors.Is(err, syscall.EACCES)`) | PASS | killed by the in-package row; install row is text-only (expected) |
| M3 | repository operand dropped (`git cat-file --batch failed: %w`) | PASS | PASS | **survivor** — residual R-A |
| M4 | operation renamed (`git ls-tree failed in %s: %w`) | FAIL | FAIL | killed |
| M5 | cause dropped (`git cat-file --batch failed in %s`) | FAIL `err = "git cat-file --batch failed in …/repo", want the kernel refusal preserved` | FAIL (same assertion, `…/skill-a`) | killed |
| M6 | call site `Extract` `:217` → `return fmt.Errorf("snapshot extraction failed")` | PASS (helper-direct, expected) | FAIL `errors = "snapshot extraction failed", want the wrapped cat-file operation context` | killed at the production entry: the install row observes `Extract`'s return through `install.Project`, not the helper |

Mutant clone ended at tree `d1d86ec2…` (only the untracked throwaway probe present).

## 5. Rulings check

- R1 wrap/classify like the package's own shape, sanitized, caller mapping unchanged — met (§1, §2).
- R2 no retry, success path unchanged (26 gitops rows incl. byte-exact vector; legacy install
  rows; closure/snapshot/contextstore/envprofile packages green), CHANGELOG `## Unreleased` →
  `### Fixed` entry present (`CHANGELOG.md:102-106`, inside the section that starts at `:67`)
  — met.
- R3 production-entry row with the prescribed injected-git shape (0755 script, 0644
  interpreter), bare-return mutant fails both rows, Windows skip in ledger vocabulary shared
  with a sibling — met (§3, §4). The install row's trip is event-driven (the `ls-tree` shim
  process rewrites its interpreter before exiting; the next spawn is `cat-file --batch`, the
  only `ls-tree` spawn in product code is `listTree`), so no wall-clock race; 1.2 s locally.
- R4 no seam; bound recorded — met.

## 6. Residuals (non-blocking, for the ledger)

- R-A: neither row pins the repository operand (mutant M3 survives). The operation name and
  the kernel cause are pinned; the operand is the same one the sibling `Wait`/`ls-tree`
  messages already carry. A `strings.Contains(err.Error(), "failed in "+repo)` assertion in
  the gitops row would close it.
- R-B: `StdinPipe`/`StdoutPipe` failures remain bare (pre-spawn fd exhaustion; no
  deterministic driver) — carried over from the producer's bound.
- R-C: Windows lanes cannot exercise the executable-bit refusal (declared skip); the
  Windows message shape for a spawn failure is therefore unverified — same bound as 3vfwch.
- Host note: an exec-stall window (fresh test binaries stranded at `_dyld_start`, SIGKILL at
  ~90 s) ran ≈01:03Z–01:20Z (M5 needed 12 attempts); the mutant driver's 40×15 s retry rode it out; local
  golangci-lint was killed by the same watchdog (rc=137) — hosted Lint job is the arbiter.

## 7. Checklist mapping

AC items 1–4 and DoD items verified as above; reviewer items (Implementation matches AC /
fits architecture / Tests green / verdict routing) checked before `accept_cr`.
