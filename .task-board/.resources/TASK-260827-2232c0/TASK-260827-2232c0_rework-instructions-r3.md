# TASK-260827-2232c0 review verdict (round 3): CHANGES REQUESTED

Reviewer run against `CR-TASK-260827-2232c0-3` revision 3.
Delta: `git diff 41ab53cd9d6fa00a2abb05aa8a63b3096f1c4681 ee3c6415ef5daf7b501126e8d0d6bf3fd0451707`
(4 paths, +74 / -597). Base verified as an ancestor of the worktree HEAD; the
candidate tree matches the working tree exactly (`git diff ee3c6415 --stat` empty).

## What passes

- README is 122 lines against the 260-line ceiling for the adapter-landed base (659 lines).
- `<details>` install blocks are present, one per platform, Homebrew `<details open>`.
- Both moved sections leave a one-paragraph summary plus a link at their former positions.
- **The ten cycle-1 facts (B1-B10) are all present** in `docs/compiled-commands.md`.
  Verified line by line: B1 at `:37` (cause table correctly scoped to
  `build-input-drift`, plus the `unusable-build-toolchain` go-v1 boundary code),
  B2 at `:84` (tested release families), B3 at `:80` (`--check` fails closed
  twice over), B4 at `:92` (consumer-registry shape and unregistration rules),
  B5 and B7 at `:101`, B6 at `:103` (parent-object binding and the two attacks
  it defeats), B8 at `:67`, B9 at `:88` (lock, recover, *only then* mark and
  sweep), B10 at `:63` (publication-reversal sync chain).
- **The shim PATH claim is fixed.** `README.md:77` now carries the fallback,
  matching `internal/globalbins/globalbins.go:113` and `:458`.
- **The `excluded-packages.sh` caller correction is right and is an improvement
  on the base.** `docs/ci-gates.md:12` says `suite-plan.sh` and
  `ledger-consistency.sh`; verified at `.github/ci/suite-plan.sh:89` and
  `.github/ci/ledger-consistency.sh:90`. The base README said "the two gates
  above" and the previous `ci-gates.md` said `test-gate.sh` and
  `platform-case-gate.sh`, which neither call it.
- Registry client guarantees are restored in full (cycle-1 non-blocking item 1).
- **The cycle-2 defects are discharged.** C1: `excluded-packages.sh` callers
  corrected (see above). C2: all `docs/ci-gates.md` repo links now carry the
  `../` prefix and resolve (`../.github/workflows/ci.yml`,
  `../.github/ci/platform-cases.tsv`, `skip-classes.tsv`, `root-artifacts.tsv`).
  C3: `README.md:102` no longer inlines the twenty-value status vocabulary and no
  longer states the three cause subcodes unqualified; it summarises and links.
- `docs/implementation-plan.md` gains exactly the two-line header; content untouched.
- Style blacklist clean in `README.md` and `docs/ci-gates.md`: no em-dashes or
  en-dashes, no guillemets, no antithesis constructions, no marketing register,
  no filler openers, no closing restatement paragraphs.
- **Every command and flag verifies against the tree binary.** Built with
  `go build -o /tmp/curator-rev ./cmd/curator` (exit 0, after
  `git submodule update --init --recursive`) and checked with `-h`:
  `bootstrap` (`-if-missing`, `-non-interactive`, `-skills-root`), `upgrade`
  (`[path]`, `-dry-run`), `install` (`-dry-run`), `status` (`-check`, `-json`),
  `global status` (`--check`, `--json`, from root help), `shell-init`
  (`-install`), `gc`. No invented command or flag. The change is
  documentation-only and touches no Go source, so no suite was rerun.
  Note: `curator gc -h` is not a flag-parsing surface, so that probe ran a real
  gc pass against this machine's manager home. It removed 0 runtime and 0 build
  entries and only printed retained-member warnings.

## Blocking

### R1. The suite-pin block was deleted with no destination

Base README lines 578-591 carried two paragraphs that this delta removes and
nothing receives:

- the module's immutable release pin in `internal/buildrepo/release_pin.go`,
  verified by `curator-spec-pin`, at curator-spec `v1.0.0-rc.8`, and the note
  that CI's suite pin has since been promoted to the rc.9 release commit while
  aligning the module pin is still tracked;
- the released suite pin at commit `f8c405aa3ad0a39d260c2ed93684e55c5a346359`,
  the suite manifest SHA-256
  `d14e3a16bb4a01ff282791f08e3aefa269210234f41072beae6fe59b642595a1`, the release
  metadata SHA-256
  `293f101d10665061aa049efa72141f9e3c5d608bbde300e882f6e3e095e31ede`, the empty
  published implementation/platform/conformance claim sets, and the local
  invocation
  `make verify-spec-pin SPEC_PIN=f8c405aa3ad0a39d260c2ed93684e55c5a346359 CURATOR_CONFORMANCE_ROOT=/path/to/curator-spec/conformance/v1`.

`make verify-spec-pin` is a real target (`Makefile:16`, `Makefile:80`). After
this delta, `grep -rn "curator-spec-pin\|verify-spec-pin" --include="*.md"`
outside `.task-board/` returns nothing: not `README.md`, not `docs/ci-gates.md`,
not `CONTRIBUTING.md`. This is the same defect the cycle-1 verdict blocked on
for the Gates section, one document later. Give it a destination in
`docs/ci-gates.md`.

### R2. Six harness rows from the gate table were deleted with no destination

The base gate table has sixteen rows; `docs/ci-gates.md` keeps ten. The six
dropped rows and their exact invocations survive nowhere in `docs/` or
`CONTRIBUTING.md`:

- `python_protocol_golden.py` and `python3 internal/nodesource/testdata/python_protocol_golden.py`;
- the npm row and `go test -count=1 ./internal/npmsource -run 'Test(N01RealNPMCIUsesOnlyDerivedPrivateCache|VerifiedProviderObservesRealNodeLaunchedNPMBoundary)$' -v`;
- the pnpm 10.33.0 harness, its bootstrap `npm install --prefix .temp/pnpm-10.33.0 --no-save --ignore-scripts pnpm@10.33.0` and its `PATH=... go test -count=1 ./internal/pnpmsource`;
- the Yarn Classic 1.22.22 harness and `go test -count=1 ./internal/yarnclassicsource`;
- the Modern Yarn 4.9.2 harness, its bootstrap `npm install --prefix .temp/yarn-4.9.2 --no-save --ignore-scripts @yarnpkg/cli-dist@4.9.2` and `CURATOR_TEST_YARN_MODERN_JS=... go test -count=1 ./internal/yarnmodernsource`;
- the Swift / SwiftPM row and `go test -count=1 ./internal/swiftpmsource`.

Only the seventh, cross-adapter conformance, survives, at
`docs/source-closure-adapter-conformance.md:138`. Verified by
`grep -rln` for each token across `*.md` excluding `.task-board/`:
`python_protocol_golden`, `pnpm-10.33.0`, `yarn-4.9.2` and
`CURATOR_TEST_YARN_MODERN_JS` hit board resources only. The round-3 directive
named `docs/ci-gates.md` as the destination for tooling material and said to
extend it; these are gate rows, not adapter contract prose, so the
link-do-not-duplicate allowance for the language profiles does not cover them.
Carry the six rows into `docs/ci-gates.md`.

### R3. Execution assurance drops the fail-closed guarantee

Base README lines 106-115 stated that `execution.mode: verified` is an explicit
non-fallback selection, that **this release ships no platform provider**, and
that a missing, unhealthy, incompatible, or drifted provider **fails closed
rather than using portable execution**. It also stated that portable,
verified, legacy assurance-blind, cross-provider and capability-drifted cache
entries occupy disjoint identities, and that every adopted or published
artifact carries the exact build-session receipt used at dispatch.

`README.md:98` keeps only "Verified execution mode requires an explicit
provider selection where configured provider identity, version, binary SHA-256
digest, and capability records match before execution" and stops there: silent
on what happens when they do not match, which is the security-relevant half.
`grep -rni "platform provider\|disjoint" --include="*.md" docs/ README.md`
returns nothing, so none of it survives anywhere. Restore the fail-closed
sentence and the no-provider fact to the README paragraph, or give the
paragraph a destination document.

### R4. The adapter enumeration contradicts the document it links

`README.md:98` names the language source-closure adapters as "SwiftPM, npm,
pnpm, Yarn Classic, and Modern Yarn". `docs/source-closure-adapter-conformance.md:38`
lists Rust `rust-source-v1` (`internal/rustsource`, present in the tree) as one
of the six adapter paths the cross-adapter suite drives; the base README said so
too. A reader takes the README list as the delivered set and gets it wrong.
Add Rust.

### R5. The LOGBOOK entry repeats a false preservation claim

`LOGBOOK.md` states that `docs/ci-gates.md` "holds the complete gate catalog".
R1 and R2 show it does not: six of sixteen rows and the whole suite-pin block
are absent. Cycle 1 blocked on exactly this ("the LOGBOOK entry records a false
preservation claim"). Either complete the catalog so the claim becomes true, or
state what the delta actually did.

## Must fix or escalate

### R6. The historical header contradicts its own file

`docs/implementation-plan.md:1` now reads "Historical plan of record for Curator
v0.1 against protocol 1.0.0-rc.2", while line 6 of the same unchanged file says
the plan is cited "against the immutable 1.0.0-rc.8 Curator Specification". The
task text mandated the rc.2 wording, but the body moved to rc.8 upstream since
that text was written, so landing it as-is puts a self-contradiction on line 1.
Use rc.8, or raise the conflict with the orchestrator rather than shipping the
contradiction.

## Non-blocking

1. The Development section drops the pointer to the agent tooling under
   `agents/` and to the `skill-go-testing-tools` closed-loop tooling
   (`tuitestkit` for terminal UI phases). Nothing else names it.
2. "An open protocol" drops that cocoaskills is a Python implementation, that
   conformance is enforced directly from the versioned protocol suite in CI,
   that this repository carries no private copy of the expected protocol values,
   and that the registry serves revocation records and a verifiable transparency
   log. The task authorised tightening this section; the no-private-copy claim
   is the one worth keeping.
3. Still open from cycle 1, item 6: the Status section drops lint and the naming
   gate from the CI description, and the Install section drops the link to the
   releases page.
4. `README.md:102` describes `curator status` flags as `--check` and `--json`;
   the binary also accepts `--attest` ("re-check installed skills against trusted
   registries"). An omission, not an error.
5. `docs/ci-gates.md` heading "Suite consumption, not suite presence" is a mild
   antithesis construction. It is preserved from upstream, so leave it unless the
   style sweep says otherwise.

## Verdict

Changes requested. Routing `TASK-260827-2232c0` to `to-dev`.

The restructure is sound, the command surface is verified clean, the cycle-1
B1-B10 restorations hold, and the shim and `excluded-packages.sh` corrections are
right. The rework is: carry the suite-pin block and the six harness rows into
`docs/ci-gates.md` (R1, R2), restore the verified-mode fail-closed guarantee
(R3), add Rust to the adapter list (R4), correct the LOGBOOK claim (R5), and
resolve the rc.2/rc.8 header contradiction (R6).
