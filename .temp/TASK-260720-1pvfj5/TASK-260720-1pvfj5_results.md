# TASK-260720-1pvfj5 — Enforce cross-platform compiled-build CI gates

**Role:** developer · **Date:** 2026-07-29 · **Board status at handoff:** `to-review`

**Working tree:** `/Users/iv/Developer/ReluxWorks/curator`, branch `agent/link-curator-skill-registry` @ `c06aa1a`
**Verification worktree:** `.temp/TASK-260720-1pvfj5/worktree` @ `origin/main` `17804ce` (detached)
**Host:** Darwin arm64, `go1.25.5` (`GOROOT=/Users/iv/.goenv/versions/1.25.5`), `golangci-lint v2.12.2`

Nothing here is a release, a release claim or a conformance claim. No commit, stage,
tag, publish or pin promotion was performed.

---

## 1. What changed

| File | State | Purpose |
| --- | --- | --- |
| `.github/workflows/ci.yml` | modified | one committed pin in `env:`, three-OS test matrix, race matrix, gate self-test job, pinned lint, non-default candidate job |
| `Makefile` | modified | `ci-test`, `race`, `race-full`, `check-ci`, `gate-selftest`, `candidate-verify-ref`, `candidate-record`, `require-pin-root` |
| `README.md` | modified | tools-and-gates table (per repo docs policy); no release wording added |
| `.github/ci/platform-cases.tsv` | new | declarative per-GOOS platform case ledger |
| `.github/ci/platform-case-gate.sh` | new | enforces the ledger against a `go test -json` stream |
| `.github/ci/test-gate.sh` | new | runs `go test -json`, then the ledger gate; both statuses fatal |
| `.github/ci/candidate-suite.sh` | new | candidate revision shape validation, identity digests, non-release evidence |
| `.github/ci/toolchain-identity.sh` | new | asserts the resolved Go toolchain is exactly `go.mod`'s |
| `.github/ci/gate-selftest.sh` | new | 31-case self-test proving the gates reject what they claim to |

No product Go file was touched.

---

## 2. The committed pin — a correction, not a promotion

`origin/main` (`17804ce`, "Pin landed rc.3 protocol") commits
`SPEC_PIN = 00b1688a9b2457ca397a0bb550acf47cad8ee967`.
This branch's `ci.yml` still carried `e72defe8510fbf5fb6e17c6027ad4da6a41e02a0` —
an **untagged ancestor** of that pin with a smaller vector set, left behind when the
branch diverged at `ecb6c1a`.

The workflow now references `00b1688a` from a single `env: SPEC_PIN`, used by every
suite checkout. Measured properties of that pin (local checkout
`.temp/TASK-260729-2kaopg/protocol-spec`, `git rev-parse HEAD` = `00b1688a…`):

```
protocol_version   1.0.0-rc.3
files              94
```

* This is **not a promotion.** It restores the branch to the revision `main` already
  commits, and prevents a silent pin regression when the branch merges.
* Pin **promotion** past this revision remains owned by `TASK-260720-38l1sy`, after
  `TASK-260720-25d05o` qualifies the release. Nothing here moves it.
* The pin publishes `1.0.0-rc.3`. No committed file claims rc.4 or rc.5 — verified
  by `grep -niE 'rc\.[0-9]|1\.0\.0-|qualif|conformance claim' README.md` → no match.
* Recorded for `38l1sy`: `00b1688a` is described by **no release tag**
  (`v1.0.0-rc.2-1-g00b1688`). It is immutable but untagged.

---

## 3. Candidate lane — explicit, non-default, never a release claim

The `candidate-conformance` job runs **only** on `workflow_dispatch` with a candidate
supplied, never on `push` or `pull_request`, and never touches `SPEC_PIN`:

```yaml
if: github.event_name == 'workflow_dispatch' &&
    (inputs.candidate_ref != '' || inputs.candidate_root != '')
```

Inputs: `candidate_ref` (full 40-hex revision), `candidate_root` (pre-materialised
root), `candidate_manifest_sha256` (expected digest). Matrix: ubuntu, macos, windows.

`candidate-suite.sh verify-ref` rejects, by shape and before any fetch: branches,
tags, `HEAD`, short hashes, uppercase hex, placeholders, the empty string, the null
commit, and any revision equal to `SPEC_PIN`. `candidate-suite.sh record` additionally
refuses a candidate whose `manifest.json` digest equals the committed pin's.

Enumeration is three separately status-checked stages (paths → sorted → per-file
digests) with a `want > 0` assertion, so a partially-failing `find` cannot produce a
short-but-consistent digest.

**Measured against the accepted rc.5 candidate root**
(`.temp/TASK-260729-3nx97g/worktree/conformance/v1`), exit **0**:

```
candidate_revision      <pre-materialised root, no revision supplied>
protocol_version        1.0.0-rc.5
manifest_sha256         sha256:b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c
tree_sha256             sha256:e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae
file_count              448
committed_released_pin  00b1688a9b2457ca397a0bb550acf47cad8ee967
evidence_class          candidate-only
release_claim           none
conformance_claim       none
```

The manifest digest, tree digest and file count reproduce the independently accepted
`TASK-260729-3nx97g` values exactly — an independent cross-check of the digest code.

---

## 4. Platform coverage — how "no case silently skipped" is enforced

`.github/ci/platform-cases.tsv` declares, per case, `must_run_on` and
`skip_allowed_on`. `platform-case-gate.sh` enforces two rules against the real
`go test -json` stream:

1. a skip is rejected unless a row tolerates it on this GOOS — **an undeclared skip
   fails the gate**;
2. every row required on this GOOS must be observed **passing** — a rename, a
   deletion, a `-run` filter matching nothing, a missing interpreter or an unset
   conformance root all fail by name instead of shrinking the run.

| AC behaviour | Enforcing case | Required on |
| --- | --- | --- |
| Windows `.cmd` launcher | `internal/runtimestore::TestInstallSingleCommandAndShims` (writes `run.cmd`, asserts `@echo off`) | linux, darwin, windows |
| Windows `.cmd` launcher | `internal/runtimestore::TestRuntimeCommandPath` (`tool.cmd` → command name) | linux, darwin, windows |
| Windows reparse point / link | `internal/adapters::TestSymlinkModeUsesRelativeLinks` | linux, darwin, windows |
| Unix no-follow | `internal/gitops::TestArchiveRejectsLinks` | linux, darwin (skip tolerated on windows: no `ln` on the job PATH) |
| Unix ownership / permission | `internal/config::TestBootstrapAndAddProject` (asserts 0o600 on unix) | linux, darwin, windows |
| Unix executable behaviour | `internal/shell::TestPosixHookEntersNestedSwitchesAndLeavesProjects` | linux, darwin |
| Windows executable behaviour | `internal/shell::TestPowerShellHookRunsOnEveryPrompt` | windows |
| Unix executable behaviour | `internal/envfiles::TestWriteProjectShapesAndSourcing` | linux, darwin, windows |
| Executable shim materialisation | `internal/install::TestEndToEndInstall` | linux, darwin, windows |
| Read-only source | `internal/install::TestDryRunTouchesNothing` | linux, darwin, windows |
| Read-only source boundary | `internal/whitelist::TestRuntimeRootsExcludedFromContext` | linux, darwin, windows |
| Opt-in developer probe | `internal/hashing::TestCrossCheckTree` | none; skip tolerated everywhere |

**Stated gap, not papered over.** Two AC terms have **no consumer package in this
repository revision** and are therefore not claimed by the ledger:

* **resource-policy** — the rc.5 native control inventory
  (`vectors/go-host-execution-policy.json#native_control_inventory`:
  `per-file-size-limit`, `aggregate-memory-limit`, `active-process-count-limit`,
  `inherited-handle-restriction`, `descendant-domain-termination`);
* **DACL specifically**, and the **read-only build-source** cases
  (`vectors/build-drivers.json#build_source_cases`).

`internal/godriver`, `internal/buildcache` and `internal/buildsource` — which the
prior audit assumed — **do not exist** at `c06aa1a`, at `origin/main` `17804ce`, or in
any worktree on this machine (`git grep -lni 'dacl|reparse' origin/main -- '*.go'` →
no match). Windows reparse coverage is provided by the link case above; DACL has no
implementation to gate. This is recorded in the ledger header and here rather than
satisfied with an invented test name.

---

## 5. Race gate

`race` runs on **`ubuntu-latest` and `macos-latest`**, both with
`go test -race -count=1 -timeout 30m ./...` plus the platform-case gate.

* `./...` rather than a maintained package list: it covers `internal/install` and
  `internal/interop` (the scope's "install" and "conformance" packages) with no list
  that can drift out of coverage.
* The scope also names "transaction" and "cache" packages. **Neither exists in this
  repository revision** (`go list ./...` → 31 packages, no `transaction`, no `cache`);
  they belong to the unlanded schema-v6 build-driver line. `./...` will cover them the
  moment they land.
* **`windows-latest` is deliberately absent**, recorded in the workflow comment: the
  Go race detector needs a C toolchain there and this task has no measurement that the
  hosted image satisfies it. Absence stated, not silently dropped.

---

## 6. Gates measured — real commands, real exit codes

Full transcript: `.temp/TASK-260720-1pvfj5/final-verification.log`. Each gate ran as
its own unpiped process; the status shown is that process's own.

### 6.1 On `origin/main` `17804ce` (verification worktree, submodules initialised)

| # | Command | Exit |
| --- | --- | ---: |
| A | `CURATOR_CONFORMANCE_ROOT=<pin 00b1688a> bash .github/ci/test-gate.sh … ./...` | **0** (`go test` 0, platform gate 0, 10/10 darwin cases ok) |
| B | `CURATOR_CONFORMANCE_ROOT=<rc.5 candidate> bash .github/ci/test-gate.sh … ./...` | **0** (`go test` 0, platform gate 0, 10/10 darwin cases ok) |
| C | `GO_TEST_FLAGS=-race … bash .github/ci/test-gate.sh … ./...` | **0** (`go test -race ./...` 0, platform gate 0) |

Logs: `gate-main-03.log` (A, B), `gate-race-main-01.log` (C).
Race wall time 2026-07-29T17:54:11Z → 17:54:31Z; no package approached the 30 m alarm.

### 6.2 On this branch `c06aa1a`

| # | Command | Exit |
| --- | --- | ---: |
| 1 | `make gate-selftest` | **0** — 31 passed, 0 failed |
| 2 | `gofmt -l cmd internal` | **0**, empty listing |
| 3 | `go vet ./...` | **0** |
| 4 | `golangci-lint run` (v2.12.2, the version CI pins) | **0** — `0 issues.` |
| 5 | no-broad-suppression guard (the `ci.yml` lint step body, run verbatim) | **0** |
| 6 | naming gate over tracked files | **0** — one README line, nothing elsewhere |
| 7 | `make build` | **0** |
| 8 | `bash .github/ci/test-gate.sh` vs pin `00b1688a` | **1** — see §7 |

### 6.3 Gate self-test coverage (all 31 assert a real exit code)

Platform gate: required case skipped (windows `.cmd`, windows reparse, linux
no-follow), required case never ran, required case failed, undeclared skip,
tolerated skip, subtest→parent resolution, `output` text that mimics a result event,
and "the shipped ledger is satisfiable" on each of linux/darwin/windows.
Candidate suite: branch / tag / `HEAD` / short hash / uppercase hex / placeholder /
empty / null commit / equal-to-pin all rejected, full 40-hex accepted; missing
manifest, nonexistent root, digest mismatch and pin-collision rejected; and the
emitted evidence asserted to carry `NOT A RELEASE`, `release_claim none`,
`conformance_claim none`, both digests, the file count and the protocol version.

---

## 7. Pre-existing branch failure — diagnosed, outside this task's ownership

Gate 8 above is **red**, and the cause is not this task's change:

```
internal/interop::TestGoldenContextCopy
    context file list diverges:
     got [".skill_triggers/en.md","SKILL.md","references/notes.md","scripts/golden-tool"]
    want [".skill_triggers/en.md","SKILL.md","references/notes.md"]
```

`agent/link-curator-skill-registry` diverged from `main` at `ecb6c1a` and is missing
41 files' worth of `main`, including the `internal/skillspec` manifest-name support
(`parse.go` +102, `types.go` +19) and the whole `internal/globalbins` package. Without
it the golden fixture's spec loads empty, so `scripts/golden-tool` is pulled into the
agent context.

Evidence that this is staleness and not a CI-gate defect: the **same command on
`origin/main` exits 0** (§6.1 A), and the same failure reproduces on this branch
*before* any of my changes (`.temp/TASK-260720-1pvfj5/pin-test-01.json`, captured
first thing).

**Not fixed here**, deliberately: the fix is "merge `main` into the branch", which is
a VCS decision outside this task's scope and outside the "never commit or stage
automatically" rule. Flagged for review as the one thing standing between this
workflow and a green run on this branch.

---

## 8. Deviations from the prior audit (`.research/260729_final-curator-ci-execution-map.md`, revision 7)

That document was written against a candidate composite that is **not present on this
machine**. Deviations, each with its reason:

| Audit item | Deviation | Reason |
| --- | --- | --- |
| §6.2 `test-linux` job, `make linux-package-guard`, godriver rejection `-run` | **not implemented** | `internal/godriver` does not exist at `c06aa1a`, at `17804ce`, or in any worktree found. A guard excluding a nonexistent package, and a `-run` pattern matching nothing, are both vacuous gates. |
| D1 / D3 (linux godriver exclusion; six hard-`t.Fatal` conformance sites) | **do not arise** | Those packages and test sites do not exist here. Measured: `go test ./...` against pin `00b1688a` on `main` is exit 0 with 2 tolerated skips. |
| §6.3 race timing risk (`internal/install` 603 s, `atomicity` 1422 s) | **does not reproduce** | `internal/install/atomicity` does not exist here; the full `-race ./...` gate completes in ~20 s wall on this host. `-timeout 30m` is kept as headroom for slower runners. |
| §6.0a fixed `REQUIRED=go1.25.5` constant | **derived from `go.mod` instead** | `setup-go` resolves from `go-version-file: go.mod`, so go.mod is the single source of truth; a duplicated constant makes every bump a two-file edit that reddens CI in between. The exact-version assertion is unchanged. |
| §6.4 `golangci-lint v2.12.2` | **adopted** | Installed and run locally: `0 issues.`, exit 0. Replaces the mutable `version: latest`. |
| §2.3 D8 | **option (b) adopted** | `GOTOOLCHAIN: "local"` + `GOENV: "off"` in workflow `env:`; no action major moves in this task. |
| CI calling gates through `make` | **calls the scripts directly** | `make` is not a guaranteed tool on `windows-latest`; the Makefile targets remain the local mirror of the same script invocations. |

**One real defect caught while validating:** `GOENV: off` unquoted is a YAML 1.1
boolean and reaches the runner as `"false"` — a genuine per-user go env file path,
not "off", which would have failed the identity assertion for a reason nobody would
have guessed. Both values are now quoted, with the reason in a comment.

---

## 9. Handoff to `TASK-260720-38l1sy` (released-pin audit)

Exact state to audit later:

* **Committed pin under audit:** `SPEC_PIN = 00b1688a9b2457ca397a0bb550acf47cad8ee967`,
  declared once at `.github/workflows/ci.yml` `env:`, referenced by `test`, `race`,
  `interop` and (for comparison only) `candidate-conformance`. Protocol `1.0.0-rc.3`,
  94 files, **no release tag** (`v1.0.0-rc.2-1-g00b1688`).
* **This task did not promote it.** It restored the branch to the value `main` already
  commits; the branch previously carried the untagged ancestor `e72defe…`.
* **Candidate evidence format** the promotion audit should expect from a
  `candidate-conformance` run: `candidate-suite-identity.txt` (revision, root,
  protocol version, `manifest_sha256`, `tree_sha256`, `file_count`, the committed pin
  it was compared against, `evidence_class candidate-only`, `release_claim none`,
  `conformance_claim none`), plus `go-test.json`, `observed-cases.tsv` and
  `platform-cases.txt` per runner, uploaded as `candidate-evidence-<os>`.
* **Reference candidate identity already measured** (rc.5 root, exit 0):
  `manifest sha256:b6f56aac…204c`, `tree sha256:e6a13215…2fae`, 448 files,
  `protocol_version 1.0.0-rc.5`.
* **Promotion pre-flight the gates now give you for free:** point `SPEC_PIN` at the
  proposed revision and the `test`/`race` jobs must stay exit 0 on all three runners
  with the platform-case gate green; `candidate-suite.sh verify-ref <rev>` must accept
  the revision shape; and `verify-ref` will refuse a revision equal to the current pin,
  so a no-op "promotion" cannot be recorded as one.

---

## 10. Producer confirmations owed on the first real CI run

These cannot be measured from macOS and are stated rather than assumed:

1. **`windows-latest` reparse points.** `internal/adapters::TestSymlinkModeUsesRelativeLinks`
   is required on windows. If the hosted image cannot create a reparse point, the gate
   fails **loudly** with `required case skipped on windows` — which is the correct
   outcome, and a real finding, not something to relax by editing the ledger.
2. **`internal/shell::TestPowerShellHookRunsOnEveryPrompt` on windows** needs `pwsh` on
   PATH; the runner image ships it, unverified here.
3. **`go env GOTOOLCHAIN` / `GOENV` exact strings** under the workflow `env:` block are
   asserted from documented Go behaviour; confirm once, and if either prints something
   else, fix the assertion — do not drop `GOENV` from the `env:` block to make it green.
4. **`golangci-lint v2.12.2` under `golangci-lint-action@v7`** ran clean locally as a
   standalone binary; the action's own install path is unverified here.
5. **Race on `ubuntu-latest`** is unmeasured on Linux (macOS-only measurement above).
