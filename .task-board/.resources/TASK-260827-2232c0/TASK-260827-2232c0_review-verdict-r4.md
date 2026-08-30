# TASK-260827-2232c0 review verdict (round 4): ACCEPTED

Reviewer run against `CR-TASK-260827-2232c0-4` revision 4.
Delta: `git diff 41ab53cd9d6fa00a2abb05aa8a63b3096f1c4681 541ce4338ce8ff25c57868d654e0c02d8d96035a`
(4 paths, +88 / -597). Base verified as an ancestor of worktree HEAD (`git merge-base
--is-ancestor` exit 0; HEAD *is* the base). Candidate tree matches the working tree
exactly (`git diff 541ce433 --stat` empty). `2bb54a25` (the adapter landing) is an
ancestor, so the base is the adapter-landed 659-line README.

The round-4 delta is narrow: one rewritten paragraph in `README.md`, seven added rows
plus a new section in `docs/ci-gates.md`, the `rc.8` header in
`docs/implementation-plan.md`, and one `LOGBOOK.md` entry. `docs/compiled-commands.md`
is unchanged since the base and carries the round-3-verified content.

## Round-3 blocking items: all six discharged

### R1. Suite-pin block, discharged

`docs/ci-gates.md:27-33` (new section "Protocol-suite pin and verification") carries both
paragraphs from base README 578-591, substantively verbatim. Every fact verified against
the source, not against the old prose:

| Claim | Source | Result |
| --- | --- | --- |
| module release pin is `v1.0.0-rc.8` | `internal/buildrepo/release_pin.go:18` `SpecReleaseTag` | matches |
| release commit `f8c405aa…6359` | `release_pin.go:20` `SpecReleaseCommit` | matches |
| suite manifest SHA-256 `d14e3a16…95a1` | `internal/buildrepo/buildrepo.go:32` `ConformanceManifestSHA256` | matches |
| release metadata SHA-256 `293f101d…1ede` | `release_pin.go:26` `ReleaseMetadataSHA256` | matches |
| empty implementation/platform/conformance claim sets | `release_pin.go:166` (three `len(...) != 0` refusals) | matches |
| CI pin promoted past rc.8, owned by `SPEC_PIN` | `.github/workflows/ci.yml:44` `SPEC_PIN: 0ed5c691…` (≠ the rc.8 commit) | consistent |
| `make verify-spec-pin` is real | `Makefile:16`, `Makefile:80-81` (`go run ./cmd/curator-spec-pin`) | exists |

The cycle-1/round-3 failure mode (a reference deleted with no destination) does not
recur: `grep -rn "curator-spec-pin\|verify-spec-pin" --include="*.md"` outside
`.task-board/` now hits `docs/ci-gates.md`.

### R2. Six harness rows, discharged

The gate table is 17 rows (`grep -c '^| ' docs/ci-gates.md` = 19, minus header and
separator): the 10 core gates plus `python_protocol_golden.py`, npm, pnpm 10.33.0,
Yarn Classic 1.22.22, Modern Yarn 4.9.2, Swift/SwiftPM, and cross-adapter conformance.
The "What it gates" and invocation columns are carried across intact, including both
bootstrap commands and the `CURATOR_TEST_YARN_MODERN_JS` export.

Every referent exists in this tree: `internal/nodesource/testdata/python_protocol_golden.py`,
`internal/npmsource`, `internal/pnpmsource`, `internal/yarnclassicsource`,
`internal/yarnmodernsource`, `internal/swiftpmsource`, `internal/crossconformance`,
`internal/crossconformance/testdata/cross-adapter-protocol-export.json`. The two named
npm tests exist verbatim: `internal/npmsource/conformance_test.go:902`
(`TestN01RealNPMCIUsesOnlyDerivedPrivateCache`) and `:649`
(`TestVerifiedProviderObservesRealNodeLaunchedNPMBoundary`), so the `-run` mask matches
something. `CURATOR_TEST_YARN_MODERN_JS` is read at
`internal/yarnmodernsource/conformance_test.go:434`.

### R3. Verified-mode fail-closed guarantee, discharged

`README.md:98` now states that `execution.mode: verified` is an explicit non-fallback
selection, that this release ships no platform provider, that a missing, unhealthy,
incompatible, or drifted provider fails closed rather than falling back to portable
execution, that portable/verified/legacy/cross-provider/capability-drifted entries occupy
disjoint cache key identities, and that every adopted or published artifact carries the
exact build-session receipt used at dispatch. That is the security-relevant half of base
README 100-119 restored.

### R4. Rust in the adapter list, discharged

`README.md:98` names Rust, SwiftPM, npm, pnpm, Yarn Classic, and Modern Yarn.
`docs/source-closure-adapter-conformance.md:38` lists `rust-source-v1` / `internal/rustsource`
as one of the six adapter paths; `internal/rustsource` exists. The README no longer
contradicts the document it links.

### R5. LOGBOOK claim, discharged

`grep -n 'TASK-260827-2232c0' LOGBOOK.md` returns exactly one hit (`:4093`); the round-3
entry with the false "complete gate catalog" claim is gone. The new entry's claims check
out: the 17-row count is right, the caller correction (`suite-plan.sh`,
`ledger-consistency.sh`) is right, the relative links are right, the README line count
(122) is right, and the quoted implementation-plan header matches the file byte for byte.

### R6. rc.2 / rc.8 contradiction, resolved

`docs/implementation-plan.md:1-2` reads "Historical plan of record for Curator v0.1
against protocol 1.0.0-rc.8." against line 6's "immutable 1.0.0-rc.8 Curator
Specification". Exactly two lines, content untouched. The producer took R6's first
option rather than shipping the self-contradiction the task text would have mandated.

## Acceptance criteria

| AC | Result |
| --- | --- |
| README under 260 lines | 122 |
| No reference dumps in README | none; every moved section leaves a one-paragraph summary plus a link |
| `compiled-commands.md` preserves every fact and command | B1-B10 all present, spot-verified at `:37`, `:59`, `:63`, `:80`, `:84`, `:88`, `:92`, `:103`; file unchanged since the round-3 verification |
| `<details>` blocks render | four, one per platform, Homebrew `<details open>` |
| Historical header present | yes, rc.8 |
| Style guide holds | one violation, see below |

### Command surface verified against the tree binary

Built from this worktree with `go build -o /tmp/curator-rev4 ./cmd/curator` (exit 0;
submodule already initialised at `agents/skills/skill-go-testing-tools`
`21585d0e937c`). Checked with `--help` / `-h`:

- `bootstrap`: `-if-missing`, `-non-interactive`, `-skills-root` — all present
- `install`: `-dry-run` present
- `upgrade`: `[path]` positional and `-dry-run` present (shares the install flag set)
- `status`: `-check`, `-json` present (also `-attest`, `-all`)
- `global status (--check, --json)`: present in root help
- `shell-init`: `-install` present; shells `auto, zsh, bash, powershell`, which matches
  README:94's zsh/bash-from-`SHELL`, Git Bash, PowerShell claim
- `gc`: present in root help ("remove unreferenced runtime entries"). Not probed with
  `-h` this round: `gc` is not a flag-parsing surface and the round-3 probe ran a real
  gc pass against this machine's manager home.

No invented command or flag in the delta. The change is documentation-only and touches
no Go source, so no test suite was rerun; the tree compiles.

### Links

Every relative link in both changed docs resolves from its own file. The `../` prefixes
in `docs/ci-gates.md` are correct: `../.github/workflows/ci.yml`,
`../.github/ci/platform-cases.tsv`, `../.github/ci/skip-classes.tsv`,
`../.github/ci/root-artifacts.tsv`, `../internal/buildrepo/release_pin.go`. All eleven
relative README targets exist.

## Must fix in the style sweep (not blocking here)

`docs/ci-gates.md:29` carries an em-dash used as rhetorical glue: "promoted to the rc.9
release commit — see `SPEC_PIN` in …". It is carried verbatim from the base README, but
the blacklist in `docs/prose-style.md` rejects it, and the round-3 verdict recorded both
changed docs as blacklist-clean. A semicolon or a sentence break fixes it.

Not blocking because `TASK-260827-xdbobc: style-sweep-and-delivery-prep` exists in
backlog with exactly this scope ("sweep README.md and every docs/*.md against
docs/prose-style.md and its blacklist; fix violations in place"), scheduled to run after
the content tasks are accepted. Bouncing a fifth cycle to duplicate a planned task would
cost more than it buys. Recorded here so the sweep has the `file:line`.

`README.md` and `docs/implementation-plan.md` are blacklist-clean: no dashes, no
guillemets, no antithesis constructions, no marketing register, no filler openers, no
closing restatement paragraphs.

## For the orchestrator, not this task's defect

The base README's six per-profile source-closure sections (lines 120-304, ~185 lines)
are dropped under the round-3 directive's "link, do not duplicate" instruction. The two
linked documents carry the contract level well (`Package.resolved`, `pnpm-lock`,
`npm-shrinkwrap`, `binding.gyp`, CGP05/CGP10, the 53 records, the assurance model), but
not the profile-specific operational sentences. Absent from all of `docs/`:
`npm ci --offline --ignore-scripts` as the replay command, the derived cacache content
store and its pruned canonical receipt, SHA-512 SRI, the direct staged-Node launch with
the fingerprinted `npm-cli.js` first argument, pnpm `importers`, and the per-profile
"does not claim lossless host network/read/write/process observation" disclaimers.
The producer followed an explicit orchestrator direction, so this is a scope decision to
confirm, not a rework item: either accept the loss, or route the operational detail into
`docs/authoring-language-adapters.md` under a separate task.

## Still open, non-blocking, carried from cycles 1 and 3

1. The Development section drops the pointer to the agent tooling under `agents/` and to
   `skill-go-testing-tools` / `tuitestkit`. Nothing else names it.
2. "An open protocol" drops that cocoaskills is a Python implementation, that conformance
   is enforced directly from the versioned protocol suite in CI, that this repository
   carries no private copy of the expected protocol values, and that the registry serves
   revocation records and a transparency log. The no-private-copy claim is the one worth
   keeping.
3. The Status section drops lint and the naming gate from the CI description; the Install
   section drops the link to the releases page.
4. `README.md:102` names `--check` and `--json` for `curator status`; the binary also
   accepts `--attest` and `--all`. An omission, not an error.
5. `docs/ci-gates.md`'s "Suite consumption, not suite presence" heading is a mild
   antithesis construction, preserved from upstream. Leave it unless the sweep says
   otherwise.

## Verdict

Accepted. Every round-3 blocking item is discharged and independently verified against
the Go source, the Makefile, the workflow, and the tree binary rather than against the
prose it replaced. The one remaining style-guide violation has a scoped destination task
already on the board.

No `commit_ack` supplied. Acceptance evidence is handed to the commit-owning mover.
