# BUG-260921-1fpaij — review verdict, revision 3 (CR-BUG-260921-1fpaij-3)

Reviewer: claude-opus-5 (RUN-260921-18a4d1), 2026-09-21 09:19–09:52Z, host e11-1.
Verdict: **ACCEPT** (`accept_cr(BUG-260921-1fpaij, revision=3)`), with three recorded
residuals for the orchestrator (R1–R3 below; none is inside the rework-1 ruling's scope).

Reviewed revision 3 only (rev2 with the `main_test.go` fixture repair is superseded and
was not considered). Everything below was rerun by me on disposable clones; nothing was
taken from the producer's evidence alone.

## 1. Exact tree, gate, patch

| Check | Result |
|---|---|
| Story worktree content (temp-index `git add -A` + `write-tree`) | `afe568ab2d694e7f9085a8576b0ad52a417d3238` = candidate tree OID |
| Patch `BUG-260921-1fpaij_change-request_rev3.patch` sha256 | `7f08482a…70ed5c` (matches CR record); 5 paths, +248/−2 |
| Disposable clone `/tmp/1fpaij-rev/cand` (base `d4fe8347` + rev3 patch, committed) | `HEAD^{tree}` = `afe568ab…` |
| Mutant clone `/tmp/1fpaij-rev/mut`, same recipe; after all mutants `git checkout --` | `HEAD^{tree}` = `afe568ab…`, tracked tree clean (only my untracked probe) |
| Hosted gate run 35575924994 (`gate/STORY-260915-3w11un/260921-080331-80578-1`) | conclusion `success`; head `1a95eb2c` → `tree afe568ab…`, `parent d4fe8347` (exact candidate) |
| Gate jobs | Test ubuntu/macos/windows, Race ubuntu/macos, Lint, Naming, Interop conformance, Gate self-test ×3: all success; rose-air + candidate suite skipped (conditional lanes) |
| `cmd/curator/main_test.go` | untouched in rev3 (not among the 5 paths) |

## 2. Classifier sits on the rework-1 boundary (rows rerun on the candidate clone)

`internal/gitignore` `go test -count=1 -v` (6/6, 1.0 s): `TestEnsureFailsWithoutIgnoreThenFixes`,
`TestEnsurePassesWhenAlreadyIgnored`, `TestMissingExitOneIsNotIgnored` (c),
`TestMissingSpawnFailureIsToolError` (a), `TestMissingNonRepositoryIsNotIgnored` (b),
`TestMissingGitAbsentIsToolError` — all PASS. `go vet`, `gofmt -l`, `golangci-lint run
./internal/gitignore/...` (0 issues) clean.

Precompiled `internal/install` binary (candidate clone, package cwd):
`TestProjectRefusesOnGitignoreSpawnFailure` (d) PASS 0.19 s, `TestGitignoreGateSkips` PASS,
`TestAdapterLedgerCommitsAfterTheMirrorsItClaims` (the PR #83 victim) PASS,
`TestMissingSystemCommandFails` PASS.

Precompiled `cmd/curator` binary: the two tests the rev1 gate broke —
`TestCLIExecutionAssuranceSelectionIsPortableDefaultAndVerifiedFailClosed` (3 subtests) and
`TestUpgradeDryRunDoesNotCreateOrFetchSkillsRoot` — PASS on rev3 with the unmodified non-repo
fixtures; the rework mask `Gitignore|DryRun|ExecutionAssurance` (4 cmd/curator tests) PASS.

### Before / after at the production entry `install.Project` (same fixture as row d)

| Tree | `Result.Status` | Surface |
|---|---|---|
| base `d4fe8347` (row d copied in: FAIL `status = "skipped", want failed`) | `skipped` | `Messages: ["test: generated paths are not ignored by git; missing entries: .agents/, .claude/skills/; skipped"]` while git could not be executed |
| candidate `afe568ab` | `failed` | `Errors: ["git check-ignore failed for \".agents/\": fork/exec <shim>/git: permission denied"]`, `Messages: []` |

CLI mapping (`cmd/curator/main.go` `cmdInstallMode`): `failed` → `error: …` on stderr, exit 1;
`skipped` stays exit 0. Non-repository root at the entry: unchanged `skipped` + policy message
(that is what the two CLI dry-run tests pin).

### Sanitization

Wrapped shape `git check-ignore failed for "<entry>": [<stderr>: ]<errno chain>` — entry
name only; the probe path (`<entry>/.curator-probe`), the project root and the environment
are absent (verified on the printed candidate message above and by mutant M5). The errno text
names the resolved git binary path, exactly as the sibling `gitops` spawn wrap (30ycv0,
`git cat-file --batch failed in %s: %w`) does. Error chain preserved: `errors.Is(err,
syscall.EACCES)` and `errors.Is(err, exec.ErrNotFound)` are pinned by the rows.

### Caller table (every `Missing`/`Ensure` call site in the module)

| Call site | Spawn/exec failure (non-`*exec.ExitError`) | Any git exit status (`*NotIgnoredError`) |
|---|---|---|
| `internal/install/install.go:262` managed gate | `failf("%v")` → `failed` (same class as the manifest-read `failf` right above) | `skipped` + `alias: <policy>; skipped` (unchanged) |
| `internal/install/install.go:291` devsub gate | same (`failed`) — driven by my probe, see §3 | `skipped` (unchanged) |
| `cmd/curator/main.go:453,1167` | only `gitignore.Append` — no classification involved | n/a |

No retry anywhere in `Missing`/`Ensure` (single `cmd.Run()` per entry). The policy string
is byte-identical (`NotIgnoredError.Error()` renders the old `fmt.Errorf` text; M7 shows the
rows and `TestGitignoreGateSkips` pin the type, not only the text).

## 3. Mutants (mutant clone; per mutant: gitignore package, install rows incl. my devsub probe, the two CLI tests)

| # | Mutant (`internal/gitignore/gitignore.go` unless noted) | Killed by | Survives |
|---|---|---|---|
| M1 | conflation restored (`if _ = exitErr; true {` → every error = not ignored) | (a) `TestMissingSpawnFailureIsToolError`, `TestMissingGitAbsentIsToolError`, (d) `TestProjectRefusesOnGitignoreSpawnFailure`, probe `TestZZDevsubGate…` | (b), (c), CLI pass (as they must) |
| M2 | rev1 strictness (`errors.As && ExitCode() == 1`) | (b) `TestMissingNonRepositoryIsNotIgnored`; CLI `…/portable` (`main_test.go:345`) and `TestUpgradeDryRunDoesNotCreateOrFetchSkillsRoot` (`main_test.go:703`) — the exact rev1 gate failure reproduced | (a), (d) |
| M3 | `install.go:263-266` managed-gate spawn branch dropped | (d) (`status = "skipped"`) | — |
| M4 | `install.go:292-295` devsub-gate spawn branch dropped | **only my probe** `TestZZDevsubGateRefusesOnGitignoreSpawnFailure` | **every committed row** → stated bound (R3) |
| M5 | probe path in the message (`%q, probe`) | (a) (`.curator-probe` pin); also my devsub probe | (d) (no probe-path pin at the entry) |
| M6 | project root in the message (`for %q in %s`) | — | **survivor** → stated bound (R3): root absence is verified by inspection/printed message, not pinned |
| M7 | `Ensure` returns the old plain `fmt.Errorf` (no `NotIgnoredError`) | (b), (c), `TestGitignoreGateSkips` (`failed` instead of `skipped`), both CLI tests | (a), (d) |
| M9 | only `exec.ErrNotFound` is a tool error (EACCES → not ignored) | (a), (d), my probe | `TestMissingGitAbsentIsToolError` (as expected) |

Reviewer probe `TestZZDevsubGateRefusesOnGitignoreSpawnFailure` (untracked, mutant clone only,
never committed): dev substitution active and ignored; PATH git = 0755 script whose shebang
interpreter is a tiny Go binary that counts calls, execs the real git and chmods itself 0644
after the 2nd call. Result on the candidate: interpreter ran exactly twice
(`check-ignore -q .agents/.curator-probe`, `check-ignore -q .claude/skills/.curator-probe`),
the 3rd spawn (devsub gate) failed at execve and `install.Project` returned `failed` with
`git check-ignore failed for "Skillfile.dev.json": fork/exec <shim>/git: permission denied`,
no policy message. Under M4 the same probe fails (`skipped`).

## 4. Hosted lanes (artifacts of run 35575924994, per-test rows extracted from `go-test.json`)

| Lane | (a) | (b) | (c) | GitAbsent | (d) | GateSkips | AdapterLedger… | CLI portable / UpgradeDryRun |
|---|---|---|---|---|---|---|---|---|
| Test ubuntu | pass | pass | pass | pass | pass | pass | pass | pass / pass |
| Test macos | pass | pass | pass | pass | pass | pass | pass | pass / pass |
| Race ubuntu | pass | pass | pass | pass | pass | pass | pass | pass / pass |
| Race macos | pass | pass | pass | pass | pass | pass | pass | pass / pass |
| Test windows | skip¹ | pass | pass | pass | skip¹ | pass | pass | pass / pass |

¹ `executable-bit spawn refusal is exercised on the unix runners` — `skips-observed.tsv`
classifies both as `platform-control / allowed-platform-control` via the existing ledger
regex `(is|are) exercised (on|by) ` (the same text as the 3vfwch/30ycv0 sibling rows). No
new ledger vocabulary.

Ratio line (`internal/crossconformance` `draftsources_semantic_test.go:54`): ubuntu
`semantic cases: 88 driven, 4 known-gap, 1 bound, 1 skipped, 94 total` — byte-identical to
the previous Story gate on the same lane (run 35545383544, 30ycv0); macos 89/4/1/0, windows
42/2/1/49 as before. No corpus change, as claimed.

## 5. Legacy goldens, CHANGELOG, docs

- Legacy goldens / full suites: I did not rerun the full `internal/install` (295 tests) or
  `internal/crossconformance` locally (host exec-stall window 09:26–09:44Z during this
  review: fresh binaries stranded at 12 KB RSS and SIGKILLed; `syspolicyd` was absent and
  came back at 09:43Z). The hosted gate on the exact candidate tree is the arbiter for those
  (all lanes green, Interop conformance gate green); the producer's "supplementary-run
  blocker" left nothing unrun that the hosted evidence does not cover.
- CHANGELOG: entry under `## Unreleased` → `### Fixed` (line 107–114), wording matches the
  narrowed rule ("A git that cannot be executed … Git's own verdicts, including 'not a
  repository', are policy outcomes as before … success path unchanged").
- Docs: `docs/*.md` only lists the `--fix-gitignore` flag; no prose describes the gate's
  skip/refuse semantics, so no doc drift. The narrowed rule lives in the `Missing` and
  `NotIgnoredError` comments.
- Windows: declared skips only through the ledger vocabulary (see §4).

## 6. Residuals (recorded, not blocking — outside the rework-1 ruling's letter)

- **R1 — signal-killed git is classified as a policy outcome.** Probe (candidate,
  `TestZZSignalKilledGitClassification`, git shim `kill -9 $$`): `cmd.Run()` → `signal:
  killed`, `*exec.ExitError` with `Exited()=false ExitCode()=-1` → `Missing` returns
  `[".agents/"], nil` and `Ensure` returns the `NotIgnoredError` policy message. This
  follows the ruling exactly ("error iff non-`*exec.ExitError`") but a killed git produced
  no verdict; on this host's stall windows (SIGKILL of stranded children) an install would
  still `skipped` with "not ignored". Cheapest follow-up if wanted: treat
  `!exitErr.Exited()` as a tool error (`git check-ignore failed for %q: %w`). Orchestrator's
  call; not part of this bug's AC.
- **R2 — `--fix-gitignore` write failure now `failed` (was `skipped`).** Probe on both trees
  (`.gitignore` removed, project dir 0555, `Options{FixGitignore: true}`): base →
  `skipped`, `Messages: ["test: open …/.gitignore: permission denied; skipped"]`; candidate
  → `failed`, `Errors: ["open …/.gitignore: permission denied"]`. Consequence of `Append`'s
  `*fs.PathError` not being a `NotIgnoredError`; fail-closed and consistent with the fix's
  principle (a write failure is not a policy outcome), CLI exit changes 0 → 1 for that
  case. Not in the CHANGELOG entry; a one-clause mention at landing would make it declared.
- **R3 — bounds.** M4 (devsub-gate caller branch) survives every committed row — killed only
  by my uncommitted probe; M6 (project root in the message) is unpinned — root absence is
  verified by the printed message, not by a row.

## 7. Reruns I did myself vs accepted

Rerun myself (clones, host e11-1, shell bash drivers with `set -o pipefail`, logs under
`/tmp/1fpaij-rev/logs`): everything in §2–§3 and the base/candidate before-after pair.
Accepted from evidence: hosted lanes in §4 (artifact rows, gate commit tree verified).
Not run: full local `internal/install` / `internal/crossconformance` (see §5).
