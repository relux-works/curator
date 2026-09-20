# TASK-260910-stbg4d — review verdict, revision 2: ACCEPT

Reviewer run RUN-260919-ccc2b1 (claude-opus-5, independent; the same
reviewer as revision 1). Read-only: no repository file was modified;
mutants ran in a disposable clone under /tmp at the gate commit, never
in the Story worktree. Shell: zsh; exit codes are the binary's own.
Run goal queried: not goal-bound; no directives recorded. Full command
evidence: TASK-260910-stbg4d_review-rev2-evidence.log.

## Candidate identity

- CR-TASK-260910-stbg4d-2, base `97a161da…`, candidate tree
  `727d33465bf20f51d42ed8a0739e0e53bcebe186`; patch resource sha256
  `60f90e66…` matches the CR record.
- Working tree of `.temp/STORY-260910-1cnwwp/worktree` re-hashed through
  a temporary index → `727d3346…` before and after the review: exact
  candidate, unchanged.
- Hosted gate: run 35446654018 conclusion `success`, headSha
  `d18977aa…`; `git cat-file -p d18977aa` → `tree 727d3346…`,
  `parent 97a161da…`. The green gate (ubuntu/macos/windows tests, race
  lanes, lint, naming, interop, gate self-tests; rose-air and candidate
  suite skipped, not passing) ran on the exact revision-2 tree.
- Revision 1 → 2 delta is confined to the rework: cmd/curator/main.go
  (gated intercept), project_resolve.go (+`stripCloneFraming`),
  draft_help.go (comment), draft_diagnostics_test.go, docs/cli.md.
  README, docs/troubleshooting.md, internal/buildrepo/transport.go and
  draft_diagnostics.go are byte-identical to revision 1 (`git diff
  --stat 071b60a5 727d3346` on those paths is empty).

## Rev1 findings — verified closed at the production entry

### F1 (clone framing → `unknown`, dead fallback) — CLOSED
- `stripCloneFraming` drops only lines matching
  `^Cloning into '[^']+'\.\.\.$` (after TrimSpace) from clone stderr
  before `ClassifyFetchOutput`, in the clone loop only; the fetch
  branch, `gitFailureDetail`, the retry structure and the buildrepo
  closed table are untouched. `gitops.Clone` runs a plain
  `git clone -- <url> <dest>` (not `--bare`), so the real framing line is
  exactly the one stripped.
- Real git, candidate binary, `invalid.invalid` fixture (my rev1 probe
  rerun): single endpoint → `warning: kit: endpoint 1 (https):
  availability: endpoint unavailable (DNS, connection, or HTTP
  502/503/504 failure)` + exhaustion line with remediation, exit 1, no
  lock; two-endpoint `availability-auth` policy → endpoint 2 (ssh,
  provider "team-ssh") is attempted and classified availability;
  `fallback: none` → `endpoint 2: not attempted`. No URL, scheme, `.git`,
  `Could not resolve host` or raw git text on stderr. The base binary on
  the same fixture prints the raw URL-bearing git output and `unknown`.
- Production rows rerun: `TestDraftFetchFailureSanitizedThroughCLI`
  (fake-git arm with the production two-line stderr, asserts the
  availability clause and the absence of `unknown: unclassified failure`
  / `Could not resolve host`; no network), `TestDraftAvailabilityFallbackThroughCLI`
  (arm 1 DNS-failing, arm 2 scp-like URL rewritten to the bare fixture →
  exit 0, lock bound to the alternate commit with identity
  `fixture.test/kit`, no `endpoint 2: not attempted`),
  `TestDraftStripCloneFramingTable` (framing+DNS → availability, DNS alone
  → availability, framing alone → unknown, near-match kept → unknown,
  empty → unknown, diagnostic-only input byte-identical) — all PASS.
- Mutants (each attributed by its failure line; exit 1 = killed):
  c1 re-poison (strip removed at the call site) → sanitized test
  `:875` shows the exact rev1 symptom `unknown: unclassified failure`,
  fallback test `:916` exit 1 → KILLED; c3 pattern widened to the prefix
  `^Cloning into` → strip table `near-match-kept: class = availability,
  want unknown` → KILLED (fail-closed bound holds); c4 pattern also
  swallowing `fatal:` lines → sanitized test + `framing-plus-dns`/`dns-alone`
  rows → KILLED.

### F2 (`-h|--help|help` intercept changed switch-off output) — CLOSED
- cmd/curator/main.go:1179 now intercepts only `-h`/`--help` and only
  when `install.DraftSourcesEnabled(os.Getenv)` — the same predicate as
  `appendDraftUsage`, `draft_status.go:41` and `project_resolve.go:67`
  (exactly `"1"`). The bare word `help` is never intercepted.
- Base binary (97a161da, disposable clone) vs candidate binary, switch
  unset, v1 fixture with aliases `app` and `help`: 18/18 invocations
  byte-identical on exit+stdout+stderr, including the three rev1
  differers `project resolve -h|--help|help` and their `refresh`
  counterparts (rev1 measured 9/12).
- Switch on: `project resolve -h` = `--help` = `refresh --help` (one
  golden, md5 47d2ebd8…, labelled "Draft Skillfile sources (opt-in,
  unreleased)", "Frozen v1" first); `project resolve help` and
  `project refresh help` print the v1 report for alias `help` (exit 0)
  with the switch on and off; `CURATOR_DRAFT_SOURCES_V1=0 project
  resolve -h` → v1 report.
- `TestDraftProjectResolveHelp` now covers both halves at `run()`
  (switch off: flag spellings byte-identical to the bare verb, no draft
  markers, `help` alias resolves; switch on: golden, `help` alias still
  resolves) and ran on the hosted Windows lane (no skip). Mutants:
  f2a ungated (rev1 behaviour) → `:216` KILLED; f2b bare `help`
  intercepted → `:245` KILLED; f2c gate inverted → `:216` KILLED.
- docs/cli.md documents the rule at both `resolve -h` and `refresh -h`
  plus the bare-`help` sentence; `install -h`/`status -h` sentences
  already said "with the switch set".

### F3 (providers pointed at a lane that never reads them) — CLOSED
- docs/cli.md:497-504 states that `source-providers.json` applies to the
  external build lane under `CURATOR_DRAFT_TRANSPORT_RESOLUTION`, not to
  Skillfile sources, which clone/fetch with the operator's ambient Git
  credentials and validate-but-ignore the `authentication` identifier.
  Verified against the code: the switch name is
  `install.EnvDraftTransportResolution`, `DraftProvidersPath` is consulted
  only under `DraftTransportResolution`, and `cmd/curator/project_resolve.go`
  references no provider. Pinned by `TestDraftDocsPinExamples`
  (`ambient Git credentials`, `CURATOR_DRAFT_TRANSPORT_RESOLUTION`);
  mutant f3 (phrase rewritten) → `:1145` KILLED.

### Nits — done
- Duplicate row renamed `repository-alias-unknown` (15/15 rows PASS, no
  `#01`). README untouched; "unreleased" pin present in heading and anchor.
- The sanitized-failure test no longer performs a live DNS lookup.

## Accepted-as-is surfaces re-verified only where the rework could touch them
- Leaf tests rerun on the exact candidate from a precompiled binary
  (`-count=1`): 8 fast (tables, seam, strip table, both help-gating
  tests, docs pins) + remediation 15 rows + sanitized + fallback +
  documented shapes 5 rows + machine policy + git revision + frozen
  launch + repair → 18 top-level tests, all PASS, exit 0.
- Legacy/sibling regression with the switch unset: `TestUsage|TestRun|
  TestStatus|TestInstall` → 27 PASS, exit 0; `TestProjectResolve|
  TestProjectRefresh|TestDraftStatus|TestClassifyDraftMember` → 41 PASS,
  exit 0. buildrepo `TestClassifyFetchOutput|TestClassifyAdmissionCode|
  TestExhaustionErrorIsClosedVocabulary|TestResolvedTransportExhaustionListsSanitizedClasses|
  TestResolvedTransportLeaksNothingIntoErrors` → 5 PASS, exit 0
  (transport.go unchanged since rev1; the rest of that package is
  accepted from the hosted gate).
- Rev1 mutants rerun on the rev2 tree: a (install re-resolves live
  inputs before materializing) → `:1019` KILLED; b (URL appended to the
  attempt clause) → 3 tests KILLED incl. the new fallback row; c
  (remediation dropped at `printFailures`) → `:640` KILLED; d (docs
  heading "released") → `:1145` KILLED; new e (raw git stderr appended to
  the per-attempt warning) → KILLED by both fetch rows.
- `go vet ./cmd/curator/ ./internal/buildrepo/` exit 0; `gofmt -l cmd
  internal` empty; `git diff --check base..candidate` exit 0. Lint
  accepted from the hosted Lint job on the exact tree.
- Windows skips use declared reasons verbatim (`test transport wrapper is
  POSIX-only` line 60, `executes POSIX skill commands` matches line 57;
  `git is not available` follows base precedent and never fires on
  hosted runners — the platform-case gate passed).

## Bounds (unchanged from revision 1, not findings)
- After a fetch failure on an existing checkout the clone loop restarts
  at endpoint 1 (sibling logic, wording only is this leaf's).
- The `-h` intercept looks at `args[1]` only (`project resolve app -h`
  resolves `app`, as at base); README "unreleased" pin is heading + anchor.
- `TestDraftProjectResolveHelp`'s switch-off half compares candidate `-h`
  against the candidate bare verb; the base-binary identity above is the
  independent measurement.

## Verdict

ACCEPT — every acceptance clause has positive and negative rows at the
production entry, sanitization holds under two leak mutants, the launch
path performs no live-input rescan (mutant a killed), draft support is
labelled accurately, and released v1 behaviour is byte-identical with
the switch off against the base binary. Recording
`accept_cr(TASK-260910-stbg4d, revision=2, evidence=TASK-260910-stbg4d_review-verdict-rev2.md)`.
