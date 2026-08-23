# Final Curator compiled-build CI execution map — revision 6

**Task:** TASK-260729-osjeay (read-only audit, rework cycle 5)
**Target task:** TASK-260720-1pvfj5 `enforce-cross-platform-ci-gates`
**Date:** 2026-07-29
**Supersedes:** revision 5, whose four blocking defects are corrected in §1.2d. The revision-2,
revision-3 and revision-4 defects remain corrected as recorded in §1.2, §1.2b and §1.2c.

**Classification.** Read-only. No product, spec, CI, `Makefile`, pin, or `TASK-260720-1pvfj5` field
was modified. **No Go command of any kind was executed** — no `go`, `go test`, `go vet`, `go build`,
`go list`, `go version`, `gofmt`, `golangci-lint`. No install, download, or package fetch.
**Every command in §5, §6, §7, §8 and §9 is a future producer gate, not evidence.** Nothing in this
document is a green CI result. The commands this audit actually ran are enumerated in §10, and only
those; the scratch-directory `tar` reproductions in §10 rows 24–25 ran under
`.temp/TASK-260729-osjeay/tarcheck/`, and the `make` stub harness (§7.4, §10 rows 36–39, 41) runs
under `$TMPDIR` against `/bin/sh` stubs — **not** against `go`, and (new this cycle) against `/bin/sh`
`ssh`/`scp` stubs, **not** against a Windows host. Neither touched a product path.

> **One honesty correction to revisions 1–4, carried forward.** Those revisions each stated "no
> network fetch to GitHub". **Cycle 4 made six read-only network reads** — five `gh api` calls against
> `api.github.com` (release lists, release bodies, major-version tag refs) and one `curl` of
> `raw.githubusercontent.com`. They were necessary to source-verify that cycle's F4, which cannot be
> answered from anything on disk. They are recorded as §10 rows 44–49. Nothing was pulled, installed,
> downloaded as a dependency, or written outside `.temp/` and `/tmp`; no repository, host, or board
> state was mutated by them. **Cycle 5 (this revision) made no network read of any kind** — every new
> measurement is the local stub harness in §7.4 (§10 rows 53–55).

---

## 1. Executive summary

Implementation cannot start from the target task as currently worded. Six of its clauses are
unsatisfiable against the accepted rc.5 candidate, and four need a board-owner decision before a
producer writes a line of YAML. §3 is the decision packet with exact proposed wording — **D1, D2,
D3 and D5** gate the target task; **D7** gates *this* task's own handoff (a stored checklist item
this read-only scope cannot satisfy and the CLI cannot remove).

The four structural findings, in descending order of blast radius:

1. **The composite is red on every platform under the committed pin — before Linux even enters the
   picture.** Six pre-existing conformance tests hard-`t.Fatal` when `CURATOR_CONFORMANCE_ROOT` is
   set but the artifact is absent, and the committed pin publishes none of the rc.5 artifacts they
   read. The delta's *own new* tests already solved this with `t.Skipf`; the older six were never
   updated. §3 D3, §4.4.
2. **The race gate is red at Go's default 10-minute timeout, the AC demands it, and the exact gate
   has never been run.** `go test -count=1 -race ./...` exits 1 with `internal/install` at 603.306 s
   and `internal/install/atomicity` at 603.701 s — cumulative duration, no `DATA RACE`. A **focused**
   race gate at 45 m passed those same packages at **609.117 s** and **1422.407 s** (exit 0), so
   `-timeout 30m` is load-bearing and probably sufficient — but atomicity's 1422.407 s leaves only
   **1.27×** headroom against the 1800 s alarm, and it was measured with five packages in flight
   rather than forty. §3 D6, §6.3.
3. **Linux cannot run `internal/godriver` at all.** `rc5-native-control-inventory-v1` covers macOS
   and Windows only; the driver fails closed with `build_execution_control_unavailable` before the
   worker starts. §4.3.
4. **The rc.5 candidate has no revision, so no hosted runner can consume it.** It exists only as an
   uncommitted `curator-spec` working tree. §5 selects one executable delivery mechanism end to end
   for every host that can consume it, and names the fail-closed prerequisite for every host that
   cannot.

The producer's edit surface stays two files (`.github/workflows/ci.yml`, `Makefile`) **only if**
decision D3 resolves to the fallback. The recommended resolution of D3 adds six test files, and the
Windows lane adds two new script files. §4.5, §5.4.

### 1.1 Corrections carried forward from revision 1

| # | Revision 1 said | Verified truth | Why it matters |
|---|---|---|---|
| C1 | Repo HEAD `c06aa1a` is main; committed pin is `e72defe…`; worktree base `17804ce` is "stale" | **`main` = `origin/main` = `17804ce`**, pin **`00b1688a…`**. `c06aa1a` is the tip of the *divergent* branch `agent/link-curator-skill-registry`; neither is an ancestor of the other (merge base `ecb6c1a`). | **I6 was inverted.** Composing on `c06aa1a` is what would revert the pin — from `00b1688a` (14 vectors, incl. `manager-lifecycle.json`) down to `e72defe` (10 vectors), breaking `internal/closure`. The worktree base is correct. |
| C2 | rc.5 root is "3 modified + 18 untracked files" | `git status --short --untracked-files=all -- conformance/` → **3 modified, 354 untracked** (357 lines), re-measured this cycle. | The provenance claim was understated by 20×. Digest `b6f56aac…04c` was and remains correct. |
| C3 | Linux `go test ./...` is the load-bearing blocker | It is the *third* blocker. The pin-artifact `t.Fatal` set and the 10-minute timeout are both larger and hit all three platforms. | Revision 1's YAML shape would still have been red on macOS and Windows. |

### 1.2 Corrections to revision 2 — the four blocking review findings

| # | Revision 2 defect | Correction in this revision | Where |
|---|---|---|---|
| **F1** | The selected archive command `tar -cf … -C "$(dirname "$DST")" conformance` resolves its operand to `…/candidate/conformance/conformance`, which does not exist. The one selected transport broke before it began. | `-C "$(dirname "$(dirname "$DST")")"`. **Reproduced in a scratch directory: the revision-2 form exits 1 with `tar: conformance: Cannot stat`; the corrected form exits 0 and lists `conformance/v1/manifest.json`.** A fail-closed preflight and a mandatory archive-listing assertion are added. | §5.2 C3, §10 rows 24–25 |
| **F2** | "If the frozen digests differ, adopt the new values as the candidate identity" contradicted both the task's immutable rc.5 input and invariant I2. `candidate-digest` checked only `manifest.json`; the Windows check was manifest-only. | **Adoption language deleted.** The three accepted identities are fixed constants; any mismatch aborts the run and escalates to the board owner as a *different candidate*, never a silent re-baseline. `candidate-digest` now verifies manifest digest, whole-tree digest, file count, and every file, and emits all three identities. Windows gets an equivalent, executable whole-tree verification (§5.4 W5) built on the same inventory file. | §5.2 C2, §5.4 W5, §7, I2 |
| **F3** | `check-ci` was labelled an exact CI mirror but omitted the conformance root, ran `go test ./...` with the root unset, and added `linux-package-guard`, which the macOS/Windows `test` job does not run. `make race` / `make race-full` likewise omitted the root the CI race job exports. | A `require-pin-root` guard target makes the root **mandatory** for `test`, `test-linux`, `race`, `race-full`, `check-ci`, `check-ci-linux`. `check-ci` mirrors the `test` job only; `check-ci-linux` is a separate target mirroring `test-linux`. The target→job table is rewritten so every claimed correspondence is executable and labelled *exact* / *equivalent* / *intentionally different*. | §7, §7.2, §9.2 |
| **F4** | Windows transport was still an option set — `scp` "or" an unspecified chunked base64 fallback — with no whole-tree verification and no exact test command. | §5.4 specifies **one** Windows path end to end: preflight → two archives → board-recorded digest handoff → in-box `tar.exe` extraction → PowerShell whole-tree verification → fixed native root → transferred `.cmd` runner → exit capture → evidence retrieval → cleanup. The base64 fallback is **deleted**. If any preflight step fails, the lane fails closed against a named prerequisite. | §5.4 |

### 1.2b Corrections to revision 3 — the four blocking review findings

| # | Revision 3 defect | Correction in this revision | Where |
|---|---|---|---|
| **F1** | Every native gate resolved Go through ambient `PATH` — `where go`, bare `go`/`gofmt`/`make`, `GO ?= go` — while the map's own §5.3 diagnosed exactly that hazard. | §5.0 defines **one absolute, identity-verified Go executable and `GOROOT` per host**, with `GOTOOLCHAIN=local` and `GOENV=off`. `require-toolchain` makes it executable and fail-closed; `GO`, `GOFMT` and the Windows `WIN_GO_EXE`/`WIN_GOROOT` are absolute everywhere. **The hazard is not theoretical and is now measured: the reviewer's shell resolved bare `go` to `/Users/iv/.goenv/shims/go`, this cycle's shell resolved it to `/opt/homebrew/bin/go` — same host, same repo, two different toolchains.** A third launcher on `PATH`, `/usr/local/go`, is **go1.25.1** against `go.mod`'s `go 1.25.5`. | §5.0, §5.2 C3, §5.4 W1/W6, §7, §9.2, I11 |
| **F2** | `LINUX_PKGS = $(shell $(GO) list ./… \| grep …)` and both importer guards discarded `go list`'s exit status, so a partial listing could become a green partial test lane. | Discovery moves **into the recipe** behind a status-checked assignment (`rows="$$(…)" \|\| exit 1`), then filters only materialized output; the safe set is asserted non-empty and the exclusion asserted to be exactly `internal/godriver`. **Both the defect and the fix were reproduced executably this cycle with shell stubs** (no Go): the revision-3 form exits **0** with a silently truncated 2-package list, the corrected form exits **2**. | §7, §7.4, §10 rows 36–39 |
| **F3** | The map repeatedly said the only executed race numbers are timeouts, and modelled atomicity at 2.75× / 1121 s. | **Withdrawn.** `.temp/TASK-260720-2284br/gates-rework1/gate-race.log` records **passes** under `-timeout 45m`: `internal/install` **609.117 s**, `internal/install/atomicity` **1422.407 s**, gate exit **0**. LOGBOOK 1732 supersedes the 2.75× model with **4.02×** and 1284–1494 s. Read first-hand this cycle (§10 rows 34–35). The full `-race ./... -timeout 30m` gate stays unproven — and the measured 1422.407 s is now shown against the 1800 s alarm as **1.27× headroom**, which is the real risk. | D6, §6.3, §8, §9.1, I7 |
| **F4** | Checklist item 13 `Tests green` stayed unchecked, and a verdict cannot silently treat an unchecked DoD item as satisfied. | Escalated as **decision D7** with exact proposed wording. **New source-verified constraint:** the board mutation set is `add_checklist_item` / `check_item` / `uncheck_item` only — there is **no** `remove_checklist_item` or edit mutation (§10 row 40), so the reviewer's "replace it" branch is **not executable by any agent through the CLI**. Item 13 stays honestly unchecked; a new executable task-scoped validation item covering §7.4's harness was added and checked against real exit codes. **The board owner had already recorded the D7-c reconciliation in task notes and reached the same append-only conclusion independently**, so F4's contract is resolved for this task; only the class fix (D7-a) stays open. | §3 D7 |

### 1.2c Corrections to revision 4 — the four blocking review findings

| # | Revision 4 defect | Correction in this revision | Where |
|---|---|---|---|
| **F1** | `require-toolchain` claimed the §5.0 matching-root contract but did not enforce it: no approved-root input, no comparison of `go env GOROOT` against an approved root, only "*some* `bin/gofmt` exists under the reported root", and it neither supplied nor checked `GOTOOLCHAIN=local` / `GOENV=off`. | One operator input, `GOROOT_EXPECTED`, from which `GO` and `GOFMT` are **derived**. The recipe compares the reported root byte-for-byte against it, requires the launcher to be that root's `bin/go` and the formatter to be that root's `bin/gofmt`, **exports** `GOROOT` / `GOTOOLCHAIN=local` / `GOENV=off` around every Go invocation via `$(GOENVPREFIX)`, and **reads both back** and fails closed if either drifted. A `go.mod`-vs-`Makefile` version self-check is added. **The revision-4 recipe was executed and accepted all three defect shapes (harness cases C, D, E, exit 0); the corrected recipe rejects all three plus five more (cases I–Q, exit 2).** | §5.0, §7, §7.4, I11 |
| **F2** | W1 and W4 printed `go version`, `go env GOROOT` and both archive digests with a "must equal" comment and no comparison. The batch runner captured `RC`, printed it, then ran `endlocal` without `exit /b %RC%`, so the SSH process status was not the Go test status, and W8 never retrieved the printed code. | Every W1/W4 value is captured on the **control host**, CR-stripped, and compared with `[ … ] \|\| exit 1`. `SRC_TAR_SHA256` is verified **before** the source is extracted. The runner ends with `endlocal & exit /b %RC%` and also persists `EXITCODE=` to a file; W8 retrieves the log **and** the exit file and asserts the SSH status, the file and the printed line agree. Four producer-time negative checks (W-N1…W-N4) must be run before the lane's exit 0 is trusted. | §5.4 W1, W4, W6, W8, W-N1–W-N4, I14 |
| **F3** | Source staging used `tar -cf - … \| tar -xf - -C "$SRCSTAGE"`. Under `set -e` the pipeline's status is the extractor's, so a failed producer was masked; the follow-up checks proved only that three paths existed, not that the whole set arrived. | The pipe is gone. An intermediate archive is created, listed and extracted as three separately status-checked steps, and a **complete-set assertion** — a per-file inventory enumerated *at the origin* plus a destination file count — proves nothing is missing, changed or extra. **Executed: the revision-4 pipeline exits 0 with 1 of 3 files staged (case R); the corrected form exits 1 on the same producer (S), on a missing file (T) and on an extra file (U).** | §5.2 C3, §7.4, I13 |
| **F4** | The repository's `checkout@v4` / `setup-go@v5` / `golangci-lint-action@v7` and the moving `*-latest` labels were inventoried but never compared against current upstream, and no retain-or-upgrade decision was recorded. | §2.3 is a dated, source-verified action and runner ledger with an explicit disposition per row. **It also corrects the verdict's own figures:** current majors are **checkout v7** and **setup-go v7** (both cut in the last two weeks), not v6; `golangci-lint-action v9` is right. It surfaces one *load-bearing* upstream fact — **`setup-go` v6 forces `GOTOOLCHAIN=local`, v5 does not** — which turns a version bump into an I11 correctness matter, and it corrects revision 4's proposed `golangci-lint` pin from `v2.4.0` (2025-08-14) to `v2.12.2` (2026-05-06), the version `latest` resolves to today. | §2.3, §6.0, §6.1, §6.4, I15 |

### 1.2d Corrections to revision 5 — the four blocking review findings

| # | Revision 5 defect | Correction in this revision | Where |
|---|---|---|---|
| **F1** | I11 claimed "every identity assertion still runs" on hosted runners, but the exact YAML implemented none of it: `test` ran bare `gofmt`/`go vet`/`go test` after `setup-go@v5` with no comparison step, `interop` was a bare `go test`, `lint` never checked the toolchain the action inherits, and §6.0 asked a producer only to *print* `go env GOTOOLCHAIN` once. Only `test-linux` and `race` reached `require-toolchain`, and even `test-linux`'s final godriver-rejection command bypassed it. | **§6.0a is a new, exact, fail-closed `Verify Go toolchain identity` step** — seven comparisons, added to **every** Go-consuming job (`test`, `test-linux`, `race`, `lint`, `interop`; `naming-gate` consumes no Go and gets none). Given in a `shell: bash` form that runs unchanged on all three runner OSes and a `shell: pwsh` alternate. §6.0 gains `GOENV: off` alongside `GOTOOLCHAIN: local`, without which the new step's assertion 7 cannot pass. **Executed against `/bin/sh` stubs: the step passes a healthy POSIX runner (AB) and a Windows-shape root (AG), and rejects `go1.25.1` (AC), a forced `GOTOOLCHAIN=auto` (AD), a user `GOENV` file (AE), a cross-root formatter on both shell shapes (AF, AH), a shim launcher outside `GOROOT` (AI) and `go.mod` drift (AJ).** | §6.0, §6.0a, §6.1–§6.5, §7.2, §7.4e, I11 |
| **F2** | I13 said "no producer→consumer pipe anywhere in the transport path", but §5.2 C3 step 1 and §5.2 C2 both enumerated the origin with `find … -print0 \| sort -z \| xargs -0 shasum`. `/bin/sh` reports only the last command's status, so a `find` that emitted a valid partial stream and then died produced a **short but internally consistent** inventory — and in C3 that inventory *is* the completeness authority, so nothing downstream could catch it. | Enumeration becomes **three materialized, separately status-checked stages**: paths → sorted paths → digests, each written to its own file and each `rc` compared before the next runs. No `pipefail` (POSIX `/bin/sh` has none). A `want > 0` assertion rejects an empty enumeration. **Executed: the revision-5 pipeline exits 0 with 2 of 3 files against a failing `find` (V); the corrected form exits 1 on a partial-then-failing `find` (X), a failing `sort` (Y) and a failing digest producer (Z), and produces no archive at all (AA).** | §5.2 C2, §5.2 C3, §7.4f, I13 |
| **F3** | W2 was an unguarded `ssh win "mkdir …"` under `set -u` (not `set -e`). The prose noted that an existing directory exits 1, but nothing stopped the lane; W4 then extracted both archives into that same base. Candidate extras are caught by W5, but the source tree has no post-extraction equivalent — so a stale extra source file could survive and enter the suite input with a correct archive digest. | W2 becomes an **executable empty-root precondition**: the base must be proved absent, then created with its status checked, then proved to contain **zero** entries, before W3 transports anything. W9 cleanup is status-checked **and** absence-confirmed before any retry. Two new producer-time negatives, **W-N5** (stale pre-existing base) and **W-N6** (base non-empty after creation), join W-N1…W-N4. **The control-host block was executed verbatim against `ssh`/`scp` stubs — no Windows host was contacted:** a healthy base transports (AK), a stale base stops the lane with nothing crossing the wire (AL), a base that is non-empty after creation does the same (AM), W9 confirms absence (AN) and rejects a cleanup that reported success but left the base standing (AO). | §5.4 W2, W3, W9, W-N5, W-N6, §7.4g, I14 |
| **F4** | §9.1 step 6 required an approved GOROOT and then ran `CURATOR_CONFORMANCE_ROOT=<root> go test ./internal/godriver/ …` — ambient `go`, no `GOROOT`, no `GOTOOLCHAIN=local`, no `GOENV=off` — and called the prerequisite "Go 1.25.x" while every other gate binds to exact `go1.25.5`. It contradicted I10/I11 and reinstated the defect revision 5 otherwise removed. | Step 6 is rewritten around **one operator-approved absolute Linux GOROOT** (`LEV_GO_ROOT`), a derived `LEV_GO_EXE`, the full six-assertion preflight from §5.0 run on `lev` first, and then the suite command under `GOROOT` / `GOTOOLCHAIN=local` / `GOENV=off`. Every "1.25.x" that describes a **gate requirement** is now exact `go1.25.5` (§9.1 step 6, I10; the two remaining `1.25.x` mentions describe the *disqualified* older-patch case and are correct as written). | §9.1, I10 |

### 1.3 New measurements taken

**Cycle 5 (this revision)** added exactly one kind of measurement and no network read: the recipe
harness grew from 21 to **41** cases, extending it to the §6.0a hosted identity step, the §5.2
origin-enumeration contract and the §5.4 W2/W3/W9 empty-root precondition (§10 rows 53–55). One
finding came out of building it, and it is recorded because it is the same defect class the harness
exists to catch: `/bin/sh`'s `echo` **expands `\r`** inside a Windows-shape `GOROOT`
(`…\rootW` → `…ootW`), which silently corrupted both the stub's reply and the step's own evidence
line. Both now use `printf '%s\n'`. On a real `windows-latest` runner under `shell: bash` this would
have mangled the recorded GOROOT in every job's log.

**Cycle 4** added four kinds of measurement, all in §10 rows 41–52: the 21-case
recipe/staging harness (row 41), a third re-verification of the rc.5 identities and the pin/dependency
state (rows 42–43), the upstream action-major and runner-label ledger behind §2.3 (rows 44–49, the
six disclosed network reads), and the repository's own action/label inventory read from
`origin/main` (rows 50–52). The single most consequential new fact is row 45: **`actions/setup-go`
only began forcing `GOTOOLCHAIN=local` in v6.0.0, and this repository pins `@v5`** — so until §6.0's
one-line `env:` addition lands, the hosted lane does not satisfy I11 by inheritance.

**Cycle 3** re-measured three host facts, which are unchanged and are carried forward:

| Host | Revision 2 said | Measured 2026-07-29, this audit | Consequence |
|---|---|---|---|
| `ssh win` | "blocked on an operator-installed Go" | **unreachable** — `ssh -o ConnectTimeout=10 win 'echo ok'` → `Operation timed out`, **exit 255**, twice | Windows candidate evidence has **two** named prerequisites now, not one: reachability (W-P1) and an approved Go root (W-P2). |
| `ssh lev` | "no Go; non-gating regardless" | **unreachable** — same command, **exit 255**, twice | Linux native stays non-gating; the prerequisite list grows a reachability item. Unchanged in effect. |
| `ssh relux` | "Go presence unverified since 2026-07-28" | **reachable, exit 0**; `command -v go` in the non-interactive shell returns nothing, but **`/usr/local/bin/go` exists and is executable**; `uname -sm` → `Darwin x86_64`; `/usr/bin/tar` present | relux is a usable macOS amd64 candidate host, but every gate must invoke `/usr/local/bin/go` by absolute path — a bare `go` is not on the non-interactive PATH. Its **version is unmeasured**: this audit runs no Go, so `go version` is a producer preflight (§5.3 R-P1), not a fact. |

Two structural facts about the module were also measured for the first time this cycle, and both
change what the Windows lane has to ship (§5.2 C3): the module has **4 direct + 17 indirect external
dependencies and no `vendor/` directory**, and its `replace` directive points at a **git submodule**
that must be present before `go mod vendor` can resolve it.

**One self-caught correction the reviewer did not raise.** Revision 2 asserted that
`go test ./...` on the composite is red at Go's 10-minute default and that "without `-timeout 30m`
the composite CI is red regardless of D1–D5". That inherited a pre-rework `2kaopg` measurement and
never re-checked it. The most recent verifier-3 run records the exact command exiting **0 in 444 s**
with `cmd/curator` at 384.270 s. The gate that *is* red at the default alarm is the **race** lane —
`internal/install` 603.306 s, `internal/install/atomicity` 603.701 s — which is precisely the clause
the AC names. D6, §6.3, §8 row 4, §9.1 step 4, and invariant I7 are all corrected; §10 row 32 is the
source. The recipe does not change; what the producer is allowed to *claim* does.

---

## 2. Verified state of the world

| Fact | Value | Source (§10 row) |
|---|---|---|
| `main` / `origin/main` | `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8` "Pin landed rc.3 protocol", 2026-07-14 05:20 | 1 |
| Current checkout | `c06aa1a` on `agent/link-curator-skill-registry` — **divergent from main** | 1, 2 |
| **Committed suite pin (main)** | **`00b1688a9b2457ca397a0bb550acf47cad8ee967`** (`ci.yml:28`, `ci.yml:81` at `17804ce`) | 3 |
| Pin's tag position | `v1.0.0-rc.2-1-g00b1688` — **one commit past rc.2, not a release tag** | 4 |
| Branch-only pin | `e72defe…` at `c06aa1a` — untagged, ancestor of `00b1688a`, 10 vectors. Not main's pin. | 4, 5 |
| curator-spec tags | `v1.0.0-rc.1`, `-rc.2`, `-rc.3` only. **No rc.4, no rc.5 tag.** | 6 |
| Nearest real release | `v1.0.0-rc.3` = `57c1f56…`, same 14 vectors as `00b1688a` | 4, 7 |
| rc.5 candidate root | `.temp/TASK-260729-3nx97g/worktree/conformance/v1` — dirty tree at `57c1f56` (`v1.0.0-rc.3`) | 8 |
| rc.5 dirt | **3 modified, 354 untracked paths** (357 status lines) — re-measured this cycle | 8, 23 |
| **rc.5 manifest digest** | **`b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`** — re-verified, exit 0 | 9, 23 |
| **rc.5 tree digest** | **`e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae`** over **448** files — re-verified, exit 0 | 9, 23 |
| rc.5 tree shape | 0 symlinks, 0 non-regular files, 68 directories, 0 empty directories, **0 non-ASCII path bytes, 0 paths containing whitespace** | 26 |
| Go version | `go 1.25.5` (`go.mod:3`); CI resolves via `actions/setup-go@v5` + `go-version-file: go.mod` | 10 |
| Module dependencies | **4 direct + 17 indirect external modules**, `go.sum` 45 lines, **no `vendor/`** | 28 |
| `replace` directive | `github.com/relux-works/skill-go-testing-tools/tuitestkit` → `./agents/skills/skill-go-testing-tools/tuitestkit`, a **git submodule** pinned at `21585d0e937cae47e54a788d8ae36b1780eae47f` (`v1.0.1-4-g21585d0`) | 28, 29 |
| Package count | 40 buildable package dirs under `cmd` + `internal` in the candidate | 11 |
| Candidate delta | 23 product files, **all `_test.go`**, zero production drift | 12 |
| Source-tree symlinks | exactly **2**, both outside the Go build: `.claude/skills/skill-go-testing-tools`, `.codex/skills/skill-go-testing-tools` | 30 |
| golangci-lint | action `@v7`, `version: latest` (**mutable**); `.golangci.yml` `version: "2"`, byte-identical main↔candidate | 13 |
| Race gate today | **none** in `ci.yml`, **none** in `Makefile` | 14, 31, 32 |
| Free disk | 25 GiB | 15 |

### 2.1 Current CI job inventory (`.github/workflows/ci.yml` at `origin/main`)

| Job | Runner | Steps | `CURATOR_CONFORMANCE_ROOT`? |
|---|---|---|---|
| `test` | matrix `ubuntu-latest`, `macos-latest`, `windows-latest` | gofmt `cmd internal` (`if: runner.os != 'Windows'`), `go vet ./...`, `go test ./...` | yes, on the `go test` step only |
| `lint` | `ubuntu-latest` | `golangci/golangci-lint-action@v7`, `version: latest` | no |
| `interop` | `ubuntu-latest` | `go test ./internal/interop/ -v` | yes |
| `naming-gate` | `ubuntu-latest` | inline `grep` gate | n/a |

All four jobs use `actions/checkout@v4` with `submodules: true`. No job passes `-timeout`. No job
runs `-race`. `.github/workflows/release.yml` is out of 1pvfj5 scope.

### 2.2 Current `Makefile` targets

`build`, `test` (`go test ./...`), `fmt`, `vet`, `lint`, `check` (`= vet test` + `gofmt -l .`).
No `race`, no timeout, no conformance-root plumbing, no platform targets.

**Drift:** CI checks `gofmt -l cmd internal`; `make check` checks `gofmt -l .`, which additionally
walks the `agents/skills/skill-go-testing-tools` submodule. Not equivalent gates. §7 does not change
`check`; it adds `check-ci` / `check-ci-linux` as the truthful mirrors and leaves `check` alone.

### 2.3 Action-major and runner-label ledger — measured 2026-07-29 (F4)

Everything in this section is source-verified against upstream on the date shown, by read-only
`gh api` / `curl` reads recorded in §10 rows 44–49. **Re-measure before acting on it**; upstream
majors here moved twice in the nine days before this audit.

**Action majors.** Repository values are from `origin/main:.github/workflows/ci.yml` (§10 row 50).

| Action | Repo pins | Current major | Latest release (date) | Disposition |
|---|---|---|---|---|
| `actions/checkout` | **`@v4`** (4 uses: lines 20, 25, 57, 73, 78, 97) | **`v7`** (`refs/tags/v7` → `3d3c42e5…`) | `v7.0.1`, **2026-07-20** | **Retain `@v4` in 1pvfj5.** v5 and v6 are Node-24 / credential-persistence changes with a runner floor of `v2.327.1`; v7 adds fork-PR checkout blocking for `pull_request_target`/`workflow_run`. None of it is required by any gate this task adds, and a checkout-major bump is a supply-chain change with its own blast radius. **Record it as deliberate, and hand the upgrade to a separate task.** |
| `actions/setup-go` | **`@v5`** (3 uses: lines 31, 61, 84), all `go-version-file: go.mod` | **`v7`** (`refs/tags/v7` → `b7ad1dad…`) | `v7.0.0`, **2026-07-16** | **Decision D8 — this one is not cosmetic.** `setup-go` **v6.0.0** landed a *breaking* change, PR #460 "Improve toolchain handling", whose first bullet is: *"Configure environment to avoid toolchain installs — force `go` to always use the local toolchain … via setting the `GOTOOLCHAIN` environment variable to `local`."* **Under the repo's current `@v5`, hosted jobs run with the default `GOTOOLCHAIN=auto`**, so a hosted runner is free to download a toolchain — the exact behaviour I11 forbids on native hosts. Two ways to close it; §6.0 takes (b). |
| `golangci/golangci-lint-action` | **`@v7`**, `version: latest` (lines 65–67) | **`v9`** (`refs/tags/v9` → `db9de0fc…`) | `v9.3.0`, **2026-06-29** | **Retain `@v7`.** v8 requires `golangci-lint >= v2.1.0`; v9 is the Node-20 → Node-24 runtime move plus `install-only` and the module plugin system. `.golangci.yml` is schema `version: "2"`, which v7 already supports, and nothing this task adds needs v8/v9. **The load-bearing fix here is the mutable `version: latest`, not the action major** — §6.4. |

**D8 — how to stop a hosted toolchain download.** Both options are executable; pick one and record it.

- **(a)** bump `actions/setup-go` to `@v6` or `@v7` and inherit its `GOTOOLCHAIN=local`;
- **(b)** keep `@v5` and set `GOTOOLCHAIN: local` in the workflow-level `env:` block.

**Recommendation: (b).** It is one line, it is visible at the top of `ci.yml` next to `SPEC_PIN`
instead of buried in an action's behaviour, it keeps 1pvfj5's action-version surface at zero, and it
survives a later setup-go bump unchanged. It also makes the hosted lane state I11's invariant
*explicitly* rather than depending on a third party continuing to imply it. Take (a) instead only if
the board owner wants the Node-24 runtime move in this task — then it is an action-version task, not
a CI-gates task.

**Runner labels.** From `actions/runner-images` `README.md`, read 2026-07-29 (§10 row 48). Rows are
quoted with the badge images stripped.

| Label | Current image | Arch |
|---|---|---|
| `ubuntu-latest` | Ubuntu 24.04 | **x64** |
| `macos-latest` | **macOS 26** | **arm64** |
| `windows-latest` | Windows Server 2025 | x64 |

The three rows as they appear in the README (badge images stripped, otherwise verbatim):

```
| Ubuntu 24.04<br> | x64 | `ubuntu-latest` or `ubuntu-24.04` | [ubuntu-24.04] |
| macOS 26 Arm64<br> | arm64 | `macos-latest`, `macos-26` or `macos-26-xlarge` | [macOS-26-arm64] |
| Windows Server 2025<br> | x64 | `windows-latest`, `windows-2025`, or `windows-2025-vs2026` | [windows-2025-vs2026] |
```

**Disposition: keep the moving `*-latest` labels, deliberately, and record why.** 1pvfj5's AC is
about *platform* coverage — Linux, macOS, Windows — not about a frozen image. Pinning
`macos-26` / `ubuntu-24.04` / `windows-2025` would freeze this task's gates to images that GitHub
deprecates on its own schedule and would create a silent-staleness maintenance debt the AC never
asked for. Two consequences must be stated in the evidence rather than discovered later:

1. **`macos-latest` is arm64.** The hosted macOS lane matches the local candidate host (macOS arm64)
   and does **not** match `ssh relux` (`Darwin x86_64`, §5.3). Any evidence line that says "macOS"
   must say which architecture, because they are two different lanes.
2. **The label can move under the task.** `runner-images` documents the migration as *"gradual and
   happens over 1-2 months … any workflow using the `-latest` label may see changes in the OS
   version"*, and there are open announcements for **Ubuntu 26.04 in public preview** (2026-06-11)
   and **Ubuntu 22 deprecation** (2026-06-16). The revalidation rule for the producer: **record the
   concrete image version in every CI evidence line** (`ubuntu-24.04`, `macOS 26 arm64`,
   `windows-2025` as measured in the run, not as copied from this table), so a later red job can be
   attributed to an image migration instead of to a code change.

---

## 3. Target-task contract drift — board-owner decision packet

Six clauses of `TASK-260720-1pvfj5` conflict with the accepted rc.5 candidate. **D1, D3 and D5 need
a board-owner decision**; D2 needs a clarification only; D4 and D6 are mechanical. Exact proposed
wording is given for each. All six are stated as *proposals*; this audit writes nothing to 1pvfj5.

**Two further decisions live outside this section** and are cross-referenced here so the packet is
complete: **D7** (§3, below — this task's own checklist item 13, already reconciled by the board
owner) and **D8** (§2.3 — how the hosted lane stops an implicit toolchain download, with a
recommendation of the one-line `env:` option over an action-major bump).

### D1 — Scope demands the full candidate suite on Linux; Linux cannot run `internal/godriver`

**Current scope text:**
> "Run the full supplied candidate suite on Linux, macOS, and Windows using the repository Go version…"

**Current AC text:**
> "…runs every compiled-build case on ubuntu, macos, and windows with no case silently skipped
> except protocol-defined unsupported platform controls."

**Conflict.** `internal/godriver` fails closed on Linux *before the worker starts* (§4.3). This is
not a "protocol-defined unsupported platform control" — those are individual `t.Skip`ped cases
(`boundary_test.go:268` Windows-only, `build_test.go:433` macOS-only). It is a whole-package
platform exclusion declared normatively in
`conformance/v1/vectors/conformance-claim-v3-qualification.json` as
`{"name":"linux","status":"excluded","until_task":"TASK-260728-1skseh"}`.

**Proposed scope wording (replace that sentence):**
> "Run the full supplied candidate suite on macOS and Windows using the repository Go version. On
> Linux run every package outside `rc5-native-control-inventory-v1` plus the inventory's own
> fail-closed rejection case; `internal/godriver` execution on Linux is deferred to
> TASK-260728-1skseh, which the rc.5 qualification vector names as the `until_task` for the linux
> exclusion."

**Proposed AC wording (replace that clause):**
> "…runs every compiled-build case on macos and windows with no case silently skipped except
> protocol-defined unsupported platform controls, and on ubuntu runs every package the rc.5
> qualification vector does not exclude, with the exclusion asserted by the inventory rejection
> test rather than by omission."

### D2 — AC demands `go test -race ./...`; scope demands scoped race on Linux

**Current AC text:**
> "`go test -race ./...` passes on the selected supported runner"

**Current scope text:**
> "add a supported race job on at least Linux covering transaction, cache, install, and conformance
> packages"

**Conflict.** These two clauses cannot both be met on one runner. `-race ./...` on Linux includes
`internal/godriver` → red (D1). The scope's Linux package list is by construction *not* `./...`.

**Resolution — no AC change needed if two race jobs are run.** `macos-latest` is an inventory
platform and the Go race detector supports `darwin/arm64`, so `go test -race ./...` is executable
there verbatim and satisfies the AC's "selected supported runner". A second, scoped
`race (ubuntu-latest)` satisfies the scope's "at least Linux" clause. Both are in §6.3.

**Proposed scope clarification (append):**
> "The AC's `go test -race ./...` gate is satisfied on macos-latest, the selected supported runner.
> The Linux race job is additionally required and is scoped to the packages named above."

D2 is the only one of the six a producer can implement without a wording change. The clarification
exists so a later reviewer does not read the two clauses as contradictory.

### D3 — The committed pin cannot serve six hard-`t.Fatal` conformance reads (**largest**)

**Current scope text:**
> "Keep the committed curator-spec checkout at the currently qualified released revision during
> candidate development."

**Conflict.** Six test sites read an rc.5-only artifact and call `t.Fatal` when it is missing. They
skip only when `CURATOR_CONFORMANCE_ROOT` is *unset*, not when the artifact is absent. The committed
pin `00b1688a` publishes none of those artifacts (§4.4). CI exports the root for `go test ./...`.
The composite's default `test` job is therefore **statically predicted red on ubuntu, macOS and
Windows alike**, in six packages.

The candidate's own new tests already solved this. Five of the 20 new files guard with
`t.Skipf("%s publishes no build-drivers vector", root)` before touching the artifact
(`internal/skillcheck/builddriver_context_conformance_test.go:24`,
`internal/whitelist/builddriver_context_conformance_test.go:25`,
`internal/skillspec/builddriver_conformance_test.go:39` and `:301`,
`internal/buildcache/builddriver_positive_conformance_test.go:31`). The six older sites were never
updated to that pattern. This is an internal inconsistency in the accepted composite, not a
1pvfj5 defect.

**The six sites, exactly:**

| # | File:line | Missing input | Provenance |
|---|---|---|---|
| 1 | `internal/buildsource/conformance_test.go:16-19` | `vectors/build-drivers.json` | accepted composite |
| 2 | `internal/buildcache/conformance_test.go:15` → `readJSONObject` `:63-66` | `vectors/build-drivers.json` | modified by the jrrgw9 delta |
| 3 | `internal/scopes/gc_conformance_test.go:38-41` | `vectors/external-repository-lifecycle.json` | **new in the jrrgw9 delta** |
| 4 | `internal/marker/marker_v2_test.go:37-41` | `schema-cases/install-marker-v2/` | accepted composite |
| 5 | `internal/whitelist/conformance_test.go:20-24` and `:25-28` | `fixtures/go-build-skill`, `expected/build-driver/context_files.json` | accepted composite |
| 6 | `internal/skillspec/conformance_test.go:106-109` | `fixtures/go-build-skill` | accepted composite |

**Recommended resolution (P1).** Convert those six reads to the delta's own `t.Skipf` pattern. It
removes the maintained list entirely, makes any pin/root combination safe permanently, and is six
mechanically identical edits. It requires a board-owner decision because it widens 1pvfj5's file
surface beyond `ci.yml` + `Makefile`.

**Proposed scope wording (append):**
> "1pvfj5 additionally owns the six conformance test sites that hard-fail on an artifact the
> committed pin does not publish (`internal/buildsource/conformance_test.go`,
> `internal/buildcache/conformance_test.go`, `internal/scopes/gc_conformance_test.go`,
> `internal/marker/marker_v2_test.go`, `internal/whitelist/conformance_test.go`,
> `internal/skillspec/conformance_test.go`). Each is changed to `t.Skipf` on a missing artifact,
> matching the pattern the accepted candidate already uses in its own build-driver conformance
> tests. No assertion is weakened: with the candidate root supplied, every one of them still runs."

Site 3 is a file `TASK-260720-jrrgw9` has not yet handed to acceptance, so that one edit can instead
be routed to jrrgw9 before it lands. Flagged, not actioned — this audit does not write to jrrgw9.

**Fallback (P3), no test-file edits, implementable inside `ci.yml` + `Makefile` only.** Split the
pin lane: export `CURATOR_CONFORMANCE_ROOT` only for a maintained "pin-servable" package list and
run the remaining packages with the variable unset. Cost: a second maintained list that drifts every
time a package gains a conformance read, with no guard that can detect the drift before it goes red.
**Not recommended**, but it is the only path that needs no board decision.

### D4 — Stale rc.4 wording (mechanical)

| Where | Stale text | Correct text |
|---|---|---|
| `checklist[0]` | "CI pins the reviewed immutable **rc.4** protocol commit" | "CI keeps the committed suite pin on the qualified revision; the rc.5 candidate enters only through an explicit non-default input" |
| `ac` | "No README release wording or committed suite pin claims **rc.4** before TASK-260720-25d05o" | `rc.5` / `1.0.0-rc.5` |
| `scope` | "Provide an explicit caller-supplied full candidate revision **or** `CURATOR_CONFORMANCE_ROOT` path" | The "full candidate revision" branch is unavailable — no rc.5 revision exists (§5.1). Only the path branch is executable today. |
| `notes` ¶1 | "candidate **rc.4** tests may use an explicitly supplied `CURATOR_CONFORMANCE_ROOT`" | `rc.5` |

`description` and the two attached `…candidate-release-ci-gates.puml/.svg` resources carry no rc.4
claim and need no change.

### D5 — "one immutable currently released curator-spec pin" is factually false

**Current AC text:**
> "Normal Curator CI keeps one immutable currently released curator-spec pin…"

**Conflict.** The committed pin `00b1688a` is `v1.0.0-rc.2-1-g00b1688` — one commit *past* the rc.2
tag, described by no release tag. It is immutable; it is not a release. The nearest actual release
with identical vector coverage is `v1.0.0-rc.3` = `57c1f56…`.

**Two options, board owner's call:**

- **D5-a (no code change):** soften the AC to "one immutable committed pin at the currently
  qualified revision", and record for `TASK-260720-38l1sy` that the pin it will audit is untagged.
- **D5-b (pin promotion):** move the pin to `57c1f56846d221ecc55786bd3c2467ec32f11730`
  (`v1.0.0-rc.3`), an actual published release with a byte-identical vector set. This would make the
  AC true for the first time. **But pin promotion is explicitly owned by `TASK-260720-38l1sy` after
  `TASK-260720-25d05o`**, so 1pvfj5 must not do it unilaterally.

**Recommendation: D5-a.** It is a wording fix inside the task, changes no gate, and leaves promotion
where the story put it. D5-b is a real improvement but belongs to `38l1sy`.

### D7 — Checklist item 13 `Tests green` cannot be satisfied or removed by this task

**The conflict.** `TASK-260729-osjeay` is a read-only audit whose scope brief forbids running Go or
heavy tests. Its stored definition-of-done nonetheless carries the generic item 13 `Tests green`,
which is tied to a command this task may not run. Under the evidence-honesty contract a command-tied
item may be checked only after that exact command has exited 0. It did not run. So item 13 stays
unchecked, `task-board handoff` fails closed, and the role end status must be applied with an
explicit `set_status`.

**The new constraint, source-verified this cycle (§10 row 40).** The cycle-3 verdict asked the board
owner to *"reconcile item 13 as not-applicable … or replace it with an executable task-scoped
validation item."* **The replace branch is not executable through the CLI by anyone.** The complete
mutation set is `add_checklist_item`, `check_item`, `uncheck_item` — there is **no**
`remove_checklist_item`, no `set_checklist_item`, no edit. Checklists are **append-only**. Combined
with the standing instruction never to edit board files directly, the only CLI-reachable states for
item 13 are *unchecked* (today) and *checked* — and checking it without a green run is precisely the
manufactured claim the verdict forbids.

**What this task did instead.** A new executable, task-scoped validation item was appended and
checked against real exit codes: the §7.4 stub harness, which runs the proposed `require-toolchain`,
`test-linux` and `linux-package-guard` recipes under `make` and confirms 0/0/2/2/2/0 across the
healthy, masked-failure, relative-toolchain, wrong-version and hosted-runner cases. That is a real
gate with real exit codes and no Go. It does not retire item 13; it gives the reviewer something
checkable that is actually true.

**The decision, with exact proposed wording.** Board owner picks one:

| Option | Action | Exact wording | Consequence |
|---|---|---|---|
| **D7-a** *(recommended)* | Add a role-level exemption so read-only/no-Go audit roles do not inherit item 13 | In the researcher role definition, qualify item 13 as: **"Tests green — where the task's scope permits executing tests; for read-only or preflight-limited tasks, satisfied by an executable task-scoped validation with recorded exit codes."** | Fixes the class, not the instance. `TASK-260729-2sxx7k` hit the identical conflict on 2026-07-29 (LOGBOOK 0510), so this is the second occurrence, not a one-off. |
| **D7-b** | Add a `remove_checklist_item` mutation, then remove item 13 here | — | Correct long-term, but it is a `task-board` CLI change and outside this board's scope. |
| **D7-c** | Board owner records item 13 as not-applicable in task notes and accepts the handoff with it unchecked | "Item 13 `Tests green` is **not applicable** to `TASK-260729-osjeay`: the task scope forbids running Go. Superseded by item 16 (§7.4 stub-harness validation, exit codes recorded). Handoff accepted with item 13 unchecked." | Unblocks this task today without a CLI change. Leaves the class defect for the next read-only task. |

**Recommendation: D7-c now, D7-a next.** D7-c is the only option that resolves *this* handoff without
either a tool change or a false claim; D7-a is what stops the third occurrence. **Do not resolve it
by checking item 13.**

> **Status: D7-c has already been taken.** The board owner recorded the reconciliation in this
> task's notes on 2026-07-29, before this cycle began, and reached the append-only constraint
> independently: *"inherited item 13 Tests green is not applicable to this read-only no-Go audit and
> remains deliberately unchecked. The supported board DSL has no remove or reword checklist
> mutation. Added a task-scoped executable review item instead; the next independent reviewer must
> check that item only after source/command-contract verification and persist an explicit verdict.
> If accepted, the orchestrator will use explicit `set_status(done)`, preserving item 13 as N/A
> rather than manufacturing test evidence."*
>
> That is item **15** — the reviewer's own gate, correctly unchecked until the reviewer checks it.
> So F4's contract is **resolved for this task**, and what remains open is only the class fix,
> **D7-a**. Item 13 must still never be checked here.

### D6 — Every `go test` gate needs `-timeout 30m` (mechanical, no decision) — **premise corrected**

Revision 2 justified this with a `TASK-260729-2kaopg` measurement and asserted "without it the
composite CI is red regardless of D1–D5". **That assertion is stale and is withdrawn.** The
measurement chain, in order:

| When | Command | Result | Source |
|---|---|---|---|
| `2kaopg`, pre-rework | `go test -count=1 ./...` | **exit 1**, `cmd/curator` reached 602.193 s at Go's fixed 10-minute default; measured twice with a different late victim each time (both victims passed in isolation at 40.0 s and 9.3 s) | LOGBOOK 0607 / the `2kaopg` timeout diagnostic |
| `jrrgw9` verifier 3, **most recent** | `CURATOR_CONFORMANCE_ROOT=<immutable rc.5 root> go test -count=1 ./...` | **exit 0 in 444 s**, `cmd/curator` 384.270 s — "the shared compiled-fixture rework recovered sufficient unchanged-timeout margin" | LOGBOOK 1637 |
| `jrrgw9` verifier 3, **most recent** | `go test -count=1 -race ./...` | **exit 1 in 610 s**: `internal/install` FAIL 603.306 s, `internal/install/atomicity` FAIL 603.701 s, both at the default alarm | LOGBOOK 1637 |
| `2284br` rework 1, **read first-hand this cycle** | focused `go test -race … -timeout 45m` over 6 packages | **exit 0**: `internal/install` **ok 609.117 s**, `internal/install/atomicity` **ok 1422.407 s**, `transaction` 35.800 s, `managerlock` 13.875 s, `staging` 1.737 s, `adapters` 3.193 s | `.temp/TASK-260720-2284br/gates-rework1/gate-race.log` + `gate-race-exit.txt` (`race exit=0`), §10 rows 34–35 |

**Correction (F3).** Revision 3 said the race lane rested only on the 918 s / 1121 s projections and
that "the only executed numbers are timeouts". **Both halves are withdrawn.** The row above is a
measured *pass* for the two expensive packages, and LOGBOOK 1732 supersedes the shared-factor model:
the packages do **not** share a race factor — `internal/install` is **×2.67** (609.117 s vs
228.344 s) and `internal/install/atomicity` is **×4.02** (1422.407 s vs 353.629 s), because
atomicity's journals carry 19–20 targets against install's 8 through an O(P²) leaf. Corrected
projections: install **890–1000 s**, atomicity **1284–1494 s**. The 1121 s figure omitted the three
post-sweep tests and is dead.

**What this changes about the risk — it sharpens it rather than relieving it.** Against a 30-minute
alarm:

| Package | Measured focused race wall | Alarm at `-timeout 30m` | Headroom |
|---|---|---|---|
| `internal/install` | 609.117 s | 1800 s | 2.96× |
| `internal/install/atomicity` | **1422.407 s** | 1800 s | **1.27×** |

`-timeout` is per test binary, so 1422.407 s does fit inside 1800 s — **but it was measured on a
focused 6-package run.** Under `./...` Go builds and runs up to `-p` (default `GOMAXPROCS`) package
binaries concurrently, so per-package wall time inflates relative to a focused run; LOGBOOK 1732
records exactly this caveat ("focused per-package numbers are measured with far fewer packages in
flight than `./...`, so they read optimistic against the real gate"). A 27 % margin is not a
comfortable place to meet that inflation. **The exact gate — `go test -race ./... -timeout 30m` —
has still never been run, and it is the one the AC names.**

One further fact from LOGBOOK 1732 that reassigns ownership: `internal/install/atomicity` does
**not** appear in `candidate-source-delta-post.txt` — it is byte-identical to the accepted worktree.
Its 1422 s is **pre-existing debt masked by `-timeout 45m` on the focused gates, not a `jrrgw9`
regression.**

So the honest statement is narrower and sharper than revision 2's:

- **Non-race:** the default-timeout run is *no longer* red on the current candidate. `-timeout 30m`
  is retained as margin, not as a fix. `cmd/curator` at 384.270 s leaves only 216 s of headroom
  against a 600 s alarm, and a single slow test regression re-crosses it — which is exactly what
  happened once already.
- **Race:** the default-timeout run **is** red, and this is the clause the AC names. LOGBOOK 1659
  diagnosed it as cumulative duration, not deadlock — no `DATA RACE` marker anywhere in the 320-line
  log; all three CPU-runnable goroutines sit in the same `saveJournal` → `namespacePathsOverlap`
  leaf. Under a **45-minute** alarm the same two packages **passed**, at 609.117 s and 1422.407 s.
  Neither fits inside 600 s; both fit inside 1800 s *when run focused*, atomicity with only 27 %
  to spare. The corrected race factors are **2.67×** and **4.02×** (LOGBOOK 1732), superseding the
  uniform 2.69× / 2.75× model and its 918 s / 1121 s projections.

**This changes nothing about the recipe** — every gate in §6, §7 and §9 still carries
`-count=1 -timeout 30m` — but it changes what the producer may claim. Do not report "CI was red
without the timeout" as a current fact; report the race lane as the one gate the timeout actually
rescues, and cite the 603.306 s / 603.701 s measurements.

**Live dependency.** `TASK-260720-jrrgw9` was routed back to `development` on exactly this race
result, for bounded test-only mitigation (`t.Parallel()` on 88 allowlisted `internal/install` tests
plus partitioning the atomicity sweep), and the production fix — hoisting or memoizing the namespace
validation out of `saveJournal` — was split out as `TASK-260729-3dr6hw`. Until that mitigation lands
and is re-measured, §6.3's race job is specified but **unproven** — not because the numbers are
unmeasured, but because the *exact* gate is: the measured pass is a focused 6-package run under a
45-minute alarm, and the AC's gate is `./...` under a 30-minute one. §6.3 states this as a named risk
rather than assuming the 30-minute alarm is sufficient.

---

## 4. Source evidence for the structural constraints

### 4.1 Dependency state

```
TASK-260720-1pvfj5   status=backlog   isBlocked=true
  blockedBy: TASK-260720-2qqq0w   status=done          ✔ satisfied
             TASK-260720-jrrgw9   status=development   ✘ NOT accepted
  blocks:    TASK-260720-38l1sy, -3pvihp, -z2z795, -z9j4c9
```

`TASK-260720-jrrgw9`'s own blockers (`-2284br`, `-1ljev5`, `-1nlmvv`) are all `done`; jrrgw9 itself
is not. `TASK-260728-1skseh` (`run-linux-native-external-repository-qualification`) is `backlog`.
The NEXT-STEP directive requires both 1pvfj5 blockers independently accepted. **1pvfj5 is not
startable today.** This audit is the pre-work; it does not unblock it.

### 4.2 rc.5 candidate identity — re-verified this run

```bash
$ cd .temp/TASK-260729-3nx97g/worktree/conformance/v1
$ shasum -a 256 manifest.json
b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c  manifest.json     # exit 0
$ find . -type f -print0 | LC_ALL=C sort -z | xargs -0 shasum -a 256 | shasum -a 256
e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae  -                 # exit 0
$ find . -type f | wc -l
448
```

The manifest digest matches the value named in this task's scope brief. The tree digest is a
**snapshot of a live, mutable worktree** — `TASK-260729-1y7okj` recorded a candidate worktree file
moving three times in one afternoon under another producer. The producer must freeze before
measuring (§5.2 C1), and §5.2 C2 states exactly what happens if the frozen values differ.

**The inventory-file identity, which the whole cross-platform verification rests on.** Verified this
run: writing the pipeline's *intermediate* stream to a file and hashing that file reproduces the
tree digest exactly, because the final `shasum` in the pipeline hashes the same bytes from stdin.

```bash
$ cd <candidate-root>
$ find . -type f -print0 | LC_ALL=C sort -z | xargs -0 shasum -a 256 > candidate-inventory.sha256
$ wc -l < candidate-inventory.sha256
448
$ shasum -a 256 candidate-inventory.sha256
e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae  candidate-inventory.sha256
```

Line format is `<64 lowercase hex><two spaces>./<relative/path>`; first line
`f116910f…  ./expected/adapter-ledger.json`, last line `5a11e968…  ./vectors/source-identities.json`.

So **`CANDIDATE_TREE_SHA256` is simultaneously the tree identity and the authenticity seal of the
inventory file**. One constant, three uses: recompute-and-compare on POSIX, authenticate the shipped
inventory on any host, and drive per-file verification on Windows (§5.4 W5). No second digest is
introduced and the documented value is unchanged.

The tree is safe for a PowerShell reimplementation: **0 symlinks, 0 non-regular files, 0 non-ASCII
path bytes, 0 paths containing whitespace** (§10 row 26). Therefore `find . -type f` and
`Get-ChildItem -Recurse -File -Force` enumerate the same set, and ordinal string comparison equals
`LC_ALL=C` byte order. Those four properties are preconditions of §5.4 W5 and the producer must
re-assert them after the freeze; if any becomes non-zero, W5's set comparison still holds but the
enumeration equivalence must be re-derived.

### 4.3 Linux is outside `rc5-native-control-inventory-v1`

```go
// internal/godriver/controls.go:75
var nativeControlPlatforms = map[string]map[string]inventoryRecord{
    PlatformMacOS:   { … },   // "macos"
    PlatformWindows: { … },   // "windows"
}
// internal/godriver/controls.go:200
func InventoryPlatform(goos string) string {
    switch goos {
    case "darwin":  return PlatformMacOS
    case "windows": return PlatformWindows
    default:        return ""            // ← linux
    }
}
// internal/godriver/controls.go:241-245
if platform == "" || nativeControlPlatforms[platform] == nil {
    return "", nil, diagnostic(CodeControlUnavailable, …)
}
```

`internal/godriver/controls_other.go` (`//go:build !darwin && !windows`) is a fail-closed stub whose
own comment says its entry points "must never be reached". `Build` probes at
`internal/godriver/build.go:161`. The protocol side agrees:
`conformance/v1/vectors/conformance-claim-v3-qualification.json` declares
`linux: excluded, until_task TASK-260728-1skseh`.

The godriver tests that drive `Build` / `prepareControlDomain` / `probeNativeControls`
(`build_test.go`, `worker_test.go`, `boundary_test.go`, `executor_test.go`,
`fingerprint_equivalence_test.go`, `graph_test.go`) carry **no build constraint and no Linux skip**;
`newSnapshotFixture` (`main_test.go:134`) has no platform guard. Only five godriver test files carry
constraints, all `_unix`/`_windows` helper files.

**Blast radius is exactly one package.** The only importers of `internal/godriver` are `cmd/curator`
(`builds.go`, `main.go`; tests use `TestedFamilies`, `SelectionCuratorGo`, `SelectionGOROOT`,
`WorkerMode`, `RunWorker`) and `internal/install` (`builddeps.go`, `global.go`, `install.go`,
`plan.go`; `stage_test.go:1288` uses `&godriver.Diagnostic{}` as a fixture value). Neither test set
calls `godriver.Build`. So the scope's named Linux race packages — transaction, cache, install,
conformance — are all Linux-safe, and `internal/interop` imports no godriver either.

**The rejection test named in §6.2 exists.** `TestProbeRejectsAnUncoveredPlatformBeforeTheWorker` is
at `internal/godriver/worker_test.go:434` in the candidate worktree (§10 row 27). Invariant I8 still
requires the producer to re-confirm it at edit time — a `-run` pattern matching nothing exits 0.

### 4.4 Pin capability matrix — which conformance inputs each revision publishes

`✔` present, `✘` absent, each verified with `git cat-file -e <rev>:<path>`.

| Conformance input | `00b1688a` (**committed pin**) | `57c1f56` (`v1.0.0-rc.3`) | `e72defe` (branch-only) | rc.5 candidate | Consumer behaviour when absent |
|---|:--:|:--:|:--:|:--:|---|
| `vectors/closures.json` | ✔ | ✔ | ✔ | ✔ | — |
| `vectors/portable-paths.json` | ✔ | ✔ | ✔ | ✔ | — |
| `vectors/manager-lifecycle.json` | ✔ | ✔ | **✘** | ✔ | `internal/closure` **FATAL** |
| `schema-cases/index.json` | ✔ | ✔ | ✔ | ✔ | — |
| `vectors/build-drivers.json` | ✘ | ✘ | ✘ | ✔ | `buildsource`, `buildcache` **FATAL**; 5 new tests skip |
| `vectors/external-repository-lifecycle.json` | ✘ | ✘ | ✘ | ✔ | `internal/scopes` **FATAL** |
| `schema-cases/install-marker-v2/` | ✘ | ✘ | ✘ | ✔ | `internal/marker` **FATAL** |
| `fixtures/go-build-skill` | ✘ | ✘ | ✘ | ✔ | `whitelist`, `skillspec` **FATAL** |
| `expected/build-driver/` | ✘ | ✘ | ✘ | ✔ | `whitelist` **FATAL**; `buildmeta`, `godriver` skip |
| `schema-cases/agent-skill-v6/valid.json` | ✘ | ✘ | ✘ | ✔ | `skillspec` builddriver test skips earlier |
| `vectors/go-host-execution-policy.json` | ✘ | ✘ | ✘ | ✔ | `godriver` skips |
| `schema-cases/build-receipt-v1/valid.json` | ✘ | ✘ | ✘ | ✔ | `buildmeta` skips |

**`build-drivers.json` exists at no published curator-spec revision** — it is untracked even in the
rc.5 worktree. The build-driver conformance suite can therefore run only under the candidate root.
This restates, with the exact per-revision matrix, what `TASK-260720-2qqq0w` recorded and explicitly
routed to 1pvfj5: "pin/gate reconciliation belongs to TASK-260720-1pvfj5".

Main's own CI is green today because main's tree predates all of this: `internal/godriver` does not
exist at `17804ce`, and main's `internal/closure/conformance_test.go` and
`internal/skillspec/conformance_test.go` read only `closures.json` and `portable-paths.json`, both
of which the pin publishes. The six FATAL sites arrive with the accepted `2kaopg` composite and the
`jrrgw9` delta.

### 4.5 Candidate delta and edit-ownership conflict check

`diff -rq` accepted (`TASK-260729-2kaopg/worktree`) vs candidate (`TASK-260720-jrrgw9/worktree`),
excluding `.git`/`.task-board`: **24 entries — 21 `Only in`, 3 `Files … differ`.** One `Only in`
is the non-product `.temp` directory. Product delta = **20 new + 3 modified = 23 files, all
`_test.go`, zero production drift.**

| Package | New files |
|---|---|
| `cmd/curator` | `lifecycle_conformance_test.go` |
| `internal/buildcache` | `builddriver_positive_conformance_test.go`, `builddriver_rejection_conformance_test.go`, `builddriver_conformance_unix_test.go`, `builddriver_conformance_other_test.go` |
| `internal/buildmeta` | `builddriver_policy_conformance_test.go` |
| `internal/buildsource` | `builddriver_conformance_test.go`, `builddriver_conformance_unix_test.go`, `builddriver_conformance_other_test.go` |
| `internal/godriver` | `builddriver_positive_conformance_test.go`, `builddriver_rejection_conformance_test.go` |
| `internal/install` | `cache_conformance_test.go`, `dryrun_conformance_test.go` |
| `internal/runtimestore` | `launcher_conformance_test.go` |
| `internal/scopes` | `gc_conformance_test.go` |
| `internal/skillcheck` | `builddriver_context_conformance_test.go` |
| `internal/skillspec` | `builddriver_conformance_test.go`, `builddriver_conformance_unix_test.go`, `builddriver_conformance_other_test.go` |
| `internal/whitelist` | `builddriver_context_conformance_test.go` |

Modified: `cmd/curator/status_test.go`, `internal/buildcache/conformance_test.go`,
`internal/closure/conformance_test.go`.

The three `_unix_test.go` / `_other_test.go` pairs use complementary constraints
(`aix||darwin||dragonfly||freebsd||linux||netbsd||openbsd||solaris` and its negation), so every host
compiles exactly one of each pair — no host is silently uncovered.

**Conflict check.** Under the recommended D3-P1 resolution 1pvfj5 owns `ci.yml`, `Makefile`, the two
new Windows script files (§5.4), and six `_test.go` files. Five of the six test files are in the
accepted composite and are **not** in the candidate delta — no conflict. The sixth,
`internal/scopes/gc_conformance_test.go`, **is** in the delta and is still owned by in-flight
`jrrgw9`: route that one edit to jrrgw9, or apply it after jrrgw9 is accepted. Under the D3-P3
fallback the test-file intersection is empty.

**Explicitly not owned by 1pvfj5:** `README.md` (owned by `2qqq0w`, done), `.golangci.yml`
(byte-identical main↔candidate), `go.mod`/`go.sum`, any `conformance/` byte, 1pvfj5's own dependency
links, the committed pin value (owned by `38l1sy`), the submodule pin
`21585d0e937cae47e54a788d8ae36b1780eae47f`.

---

## 5. Candidate delivery — one selected mechanism, end to end

### 5.0 Toolchain selection — one absolute Go per host

**This section is a precondition of every command in §5–§9.** Revision 3 diagnosed the ambient-`PATH`
hazard in §5.3 and then reintroduced it in every primary gate (`where go`, bare `go`, bare `gofmt`,
bare `make`, `GO ?= go`). That is F1, and it is corrected here.

**The hazard is measured, not theoretical.** On this one host, two agent processes in the same repo
resolved bare `go` to two different toolchains:

| Observer | `command -v go` | Effective root |
|---|---|---|
| cycle-3 reviewer | `/Users/iv/.goenv/shims/go` (a 411-byte **bash script**, §10 row 33) | `/Users/iv/.goenv/versions/1.25.5` |
| this cycle | `/opt/homebrew/bin/go` → `../Cellar/go/1.25.5/bin/go` → `../libexec/bin/go` | `/opt/homebrew/Cellar/go/1.25.5/libexec` |

`which -a go` returns **three** launchers, and their `VERSION` files were read directly this cycle
(§10 row 31) — no Go executed:

| # | Launcher on `PATH` | Root | `VERSION` file | Verdict against `go.mod:3` = `go 1.25.5` |
|---|---|---|---|---|
| 1 | `/opt/homebrew/bin/go` | `/opt/homebrew/Cellar/go/1.25.5/libexec` | `go1.25.5` | **acceptable** |
| 2 | `/usr/local/go/bin/go` | `/usr/local/go` | **`go1.25.1`** | **DISQUALIFIED — older than the module requires** |
| 3 | `/Users/iv/.goenv/shims/go` | `/Users/iv/.goenv/versions/1.25.5` | `go1.25.5` | acceptable, but reached through a shim whose target is repointable |

Launcher 2 is the sharp edge. `go.mod` declares `go 1.25.5` and carries **no `toolchain` line**. Under
Go's default `GOTOOLCHAIN=auto`, a `go1.25.1` binary asked to build this module **downloads
go1.25.5** from the toolchain proxy — a network fetch this task forbids and a gate host must never
perform implicitly. `GOTOOLCHAIN=local` converts that silent download into an immediate fail-closed
error.

**The selection rule.** **One** operator-approved absolute root per host is the *only* input. The
executable and the formatter are **derived** from it, never supplied independently — that is what
makes a cross-root pairing unrepresentable rather than merely discouraged:

```bash
# Local macOS arm64 — the primary candidate host.
GO_ROOT=/opt/homebrew/Cellar/go/1.25.5/libexec   # the ONE operator-approved constant
GO_EXE="$GO_ROOT/bin/go"                          # derived
GOFMT_EXE="$GO_ROOT/bin/gofmt"                    # derived
```

> **This is F1's fix, and revision 4 got it wrong in a way that was executably demonstrable.**
> Revision 4 accepted `GO` and `GOFMT` as two independent absolute paths and then checked only that
> *some* `bin/gofmt` existed under whatever root `go env GOROOT` happened to report. Run against
> shell stubs this cycle (§7.4, harness cases C, D, E), that recipe **exits 0** when the compiler
> comes from root A and the formatter from root B; **exits 0** when the launcher's real root is not
> the operator-approved one at all; and **exits 0** with `GOTOOLCHAIN=auto` in force. All three are
> the failure this section exists to prevent. The `Makefile` in §7 now takes `GOROOT_EXPECTED` as its
> single input, derives `GO`/`GOFMT` from it, and rejects all three (cases L, M, N).

The Homebrew Cellar root is the recorded default because its **path encodes the exact version**, so a
version change cannot happen without the constant changing, and it is not indirected through a shim
that a version-manager operation can repoint. `/Users/iv/.goenv/versions/1.25.5` is an equally valid
substitution — goenv currently has **both 1.25.1 and 1.25.5 installed** and its global version file
says `1.25.5` (§10 row 32) — but if the operator substitutes it, the constant above is what changes,
and everything downstream follows from the constant. **What is not acceptable is resolving it at run
time.**

**Environment, and exactly what each setting does.**

```bash
export GOROOT="$GO_ROOT"      # match the selected executable; never inherit
export GOTOOLCHAIN=local      # never download a toolchain; fail closed instead
export GOENV=off              # ignore the per-user `go env -w` config file
```

- `GOROOT` **must be exported to the root matching `$GO_EXE`**, not inherited. This shell exports
  `GOROOT=/Users/iv/.goenv/versions/1.25.5` (§10 row 32) while resolving bare `go` to the Homebrew
  binary — a cross-tree pairing that happens to work only because both are `go1.25.5`. It is latent,
  and one `brew upgrade` or `goenv install` separates them.
- `GOTOOLCHAIN=local` is load-bearing, per launcher 2 above.
- `GOENV=off` disables Go's per-user env config (`~/Library/Application Support/go/env` on macOS,
  `~/.config/go/env` elsewhere) — the file `go env -w` writes, which can silently inject `GOFLAGS`
  or `GOTOOLCHAIN` into every later invocation. **Measured this cycle: neither path exists today, and
  `GOFLAGS`/`GOTOOLCHAIN` are both unset** (§10 row 32). So this is defence-in-depth against a future
  `go env -w`, **not** a fix for a present condition — state it that way in the evidence. Note the
  name collision: Go's `GOENV` is unrelated to the `goenv` version manager, which uses
  `GOENV_VERSION` / `GOENV_ROOT`.

**T-P1 — the toolchain preflight, per host, before any gate.** Producer gate; not run by this audit.
Six assertions, every one comparing rather than printing. This is the shell form of the
`require-toolchain` recipe in §7; running `make require-toolchain GOROOT_EXPECTED="$GO_ROOT"` is the
same contract and is the form the gates actually use.

```bash
set -u
test -x "$GO_EXE"    || { echo "toolchain: $GO_EXE is not executable"; exit 1; }
test -x "$GOFMT_EXE" || { echo "toolchain: $GOFMT_EXE is not executable"; exit 1; }

# every probe runs under the same environment the gates will run under
E="GOROOT=$GO_ROOT GOTOOLCHAIN=local GOENV=off"

# 1. the Makefile's pinned version still matches go.mod
want="go$(awk '/^go[ \t]/{print $2; exit}' go.mod)"
[ "$want" = go1.25.5 ] || { echo "toolchain: go.mod requires $want, not go1.25.5"; exit 1; }

# 2. exact version — not "the 1.25 family"
v="$(env $E "$GO_EXE" version)" || { echo 'toolchain: `go version` failed'; exit 1; }
echo "toolchain: $v"
case "$v" in *' go1.25.5 '*) ;; *) echo "toolchain: expected go1.25.5, got: $v"; exit 1;; esac

# 3. reported root == approved root, byte for byte
r="$(env $E "$GO_EXE" env GOROOT)" || { echo 'toolchain: `go env GOROOT` failed'; exit 1; }
echo "toolchain: GOROOT=$r"
[ "$r" = "$GO_ROOT" ] || { echo "toolchain: GOROOT drift $r != $GO_ROOT"; exit 1; }

# 4. the launcher and the formatter are THIS root's, not merely absolute
[ "$GO_EXE"    = "$r/bin/go" ]    || { echo "toolchain: launcher is not $r/bin/go"; exit 1; }
[ "$GOFMT_EXE" = "$r/bin/gofmt" ] || { echo "toolchain: formatter is not $r/bin/gofmt"; exit 1; }

# 5. no implicit toolchain download
tc="$(env $E "$GO_EXE" env GOTOOLCHAIN)" || { echo 'toolchain: `go env GOTOOLCHAIN` failed'; exit 1; }
echo "toolchain: GOTOOLCHAIN=$tc"
[ "$tc" = local ] || { echo "toolchain: GOTOOLCHAIN=$tc, not local"; exit 1; }

# 6. no per-user go env file in the loop
ge="$(env $E "$GO_EXE" env GOENV)" || { echo 'toolchain: `go env GOENV` failed'; exit 1; }
echo "toolchain: GOENV=$ge"
[ "$ge" = off ] || { echo "toolchain: GOENV=$ge, not off"; exit 1; }
```

Assertion 2 is `go1.25.5` exactly, not "the 1.25 family": `go.mod` names `1.25.5`, and a
`1.25.x < 1.25.5` launcher is the disqualified case above. Assertion 1 is what stops the pinned
constant from silently outliving `go.mod`. **Record the real output of every `echo` above in the
evidence** — this is the one place where "which toolchain produced this result" is answerable.

**Why 5 and 6 are read-backs and not just exports.** The environment is exported *and* read back
because a launcher can override its caller: a `goenv` shim, a corporate wrapper, or a container
entrypoint is free to set `GOTOOLCHAIN` itself, and this host already runs bare `go` through such a
shim (§10 row 37). Exporting alone is unverifiable; reading back is what fails closed. **One
producer confirmation is owed here:** this audit executed no Go, so the exact strings `go env
GOTOOLCHAIN` and `go env GOENV` print under an explicit export are asserted from Go's documented
behaviour, not measured. Run assertions 5 and 6 once by hand on the approved toolchain before
trusting the gate; if either prints something else, fix the assertion — do **not** delete it. The
control flow itself *is* executably verified against stubs (§7.4 cases N, O).

**Every downstream command uses `$GO_EXE` / `$GOFMT_EXE`, never a bare name**: the vendoring step
(§5.2 C3), the Windows batch runner (§5.4 W6), the relux lane (§5.4 W10), every `make` invocation
(`GO="$GO_EXE" GOFMT="$GOFMT_EXE"`, §7), and the local gate set (§9.2). `make` itself may be invoked
by name — it is not part of the Go toolchain and no shim shadows it — but the `GO`/`GOFMT` it passes
down must be absolute, which `require-toolchain` (§7) enforces.

**Hosted GitHub runners get a narrow exception, and only for path discovery.** `actions/setup-go@v5`
with `go-version-file: go.mod` installs and `PATH`-fronts exactly the requested toolchain in a fresh
image with no competing launcher and no version manager, so on a hosted runner bare `go` **is** the
identity-verified toolchain and no operator can supply an approved root in advance. Those steps
therefore set `TOOLCHAIN_ALLOW_PATH=1`, which relaxes **the absolute-path shape only**: the root is
taken from `go env GOROOT` instead of from an operator constant, and every other assertion still
runs — exact version, launcher `== $r/bin/go`, formatter `== $r/bin/gofmt`, `GOTOOLCHAIN=local`,
`GOENV=off`. **Executably demonstrated: the exception passes a healthy setup-go shape (case G) and
still rejects a formatter from a different root (case Q).** §6's YAML keeps bare `go`.

**One hosted-runner caveat that F4 turned up.** `setup-go` only started forcing `GOTOOLCHAIN=local`
in **v6.0.0**; the repository pins **`@v5`**, so hosted jobs currently run with the default
`GOTOOLCHAIN=auto`. The hosted lane therefore does *not* satisfy this section by inheritance today —
§6.0 closes it with a workflow-level `GOTOOLCHAIN: local`. See §2.3 D8.

### 5.1 Why hosted candidate jobs cannot exist yet

The rc.5 candidate is an uncommitted `curator-spec` working tree (3 modified, 354 untracked) on top
of `v1.0.0-rc.3`. `vectors/build-drivers.json` — the file 15 test sites read — is itself untracked.
There is no ref for `actions/checkout` to fetch, and a Unix absolute path is meaningless to a hosted
runner. `TASK-260728-1jafds` independently recorded that `make regenerate-check` exits 2 and
`release_gate.py` exits 1 with "release gate requires a clean candidate checkout" for the same
reason.

Writing a hosted candidate job today means either fabricating a hash (forbidden by 1pvfj5 scope) or
shipping a job that can never fire. **Decision: 1pvfj5 adds no hosted candidate job.** The CI-side
candidate lane is deferred to `TASK-260720-38l1sy`, which already owns pin promotion and will have a
real revision to consume once rc.5 is committed to `relux-works/curator-spec`.

No self-hosted runner is assumed. This audit did not query GitHub, so it makes no claim about
registered runner labels; the plan requires none.

### 5.2 The selected mechanism — freeze, seal, archive, hand off

Four steps, in order. Each fails closed.

**C1 — freeze.** The authoritative root is live and has been observed moving under another producer.
Copy it once into a task-owned, read-only snapshot and never read the live tree again.

```bash
set -eu
SRC=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1
STAGE=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1pvfj5
DST="$STAGE/candidate/conformance/v1"

test -f "$SRC/manifest.json" || { echo 'preflight: source root is not a conformance root'; exit 1; }
test ! -e "$DST" || { echo 'preflight: destination exists; refusing to overwrite a freeze'; exit 1; }

mkdir -p "$(dirname "$DST")"
COPYFILE_DISABLE=1 cp -R "$SRC" "$DST"
chmod -R a-w "$DST"
```

`$(dirname "$DST")` is `$STAGE/candidate/conformance`, which does not yet contain `v1`, so
`cp -R "$SRC" "$DST"` lands the tree *at* `$DST`. The `test ! -e` guard is what makes a re-run fail
instead of silently merging a second copy into the first.

**C2 — seal the identity. Three fixed constants; a mismatch aborts.**

```bash
ACCEPTED_MANIFEST_SHA256=b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c
ACCEPTED_TREE_SHA256=e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae
ACCEPTED_FILE_COUNT=448

INV="$STAGE/candidate-inventory.sha256"
cd "$DST"

got_manifest="$(shasum -a 256 manifest.json | cut -d' ' -f1)"

# Enumeration is THREE materialized, separately status-checked stages. A single
# `find … | sort | xargs` reports only xargs's status, so a find that emits a
# valid partial stream and then dies yields a SHORT BUT CONSISTENT inventory
# (F2, I13; executably reproduced in §7.4f case V).
find . -type f -print0 > "$STAGE/candidate.paths0"
rc=$?; [ "$rc" = 0 ] || { echo "C2: candidate path enumeration failed rc=$rc"; exit 1; }
LC_ALL=C sort -z < "$STAGE/candidate.paths0" > "$STAGE/candidate.sorted0"
rc=$?; [ "$rc" = 0 ] || { echo "C2: candidate path sort failed rc=$rc"; exit 1; }
xargs -0 shasum -a 256 < "$STAGE/candidate.sorted0" > "$INV"
rc=$?; [ "$rc" = 0 ] || { echo "C2: candidate digest generation failed rc=$rc"; exit 1; }

got_tree="$(shasum -a 256 "$INV" | cut -d' ' -f1)"
got_count="$(wc -l < "$INV" | tr -d ' ')"

echo "candidate manifest sha256 $got_manifest"
echo "candidate tree     sha256 $got_tree"
echo "candidate files           $got_count"

[ "$got_manifest" = "$ACCEPTED_MANIFEST_SHA256" ] || { echo 'IDENTITY MISMATCH: manifest'; exit 1; }
[ "$got_tree"     = "$ACCEPTED_TREE_SHA256" ]     || { echo 'IDENTITY MISMATCH: tree';     exit 1; }
[ "$got_count"    = "$ACCEPTED_FILE_COUNT" ]      || { echo 'IDENTITY MISMATCH: count';    exit 1; }
```

**What a mismatch means, and what it does not.** The manifest digest `b6f56aac…04c` is the
**immutable rc.5 input named by this task's scope brief**; the tree digest and file count are the
identity of the exact tree that carries it. If any of the three differs at freeze time, the live
worktree has moved and **the producer is holding a different candidate**. That is a stop-the-line:
record the measured values, do **not** re-baseline, do **not** proceed to transport, and escalate
for a refreshed accepted candidate input and a board-owner decision. Revision 2's instruction to
"adopt the new values as the candidate identity" is **withdrawn** — silently adopting a moved tree
would make every downstream identity check tautological and would contradict invariant I2.

The digest is never written as a `ref:` (I2). No log line calls this a release (I3).

**C3 — archive.** One mechanism for every remote host, two payloads. Use the flags
`TASK-260729-2sxx7k` proved necessary: BSD `tar` embeds AppleDouble/xattr records that GNU `tar`
materializes as `._*` files, which would break the file count and the tree digest on arrival.

```bash
CAND_TAR="$STAGE/candidate.tar"

# Preflight: the archive operand must exist relative to the -C directory.
CAND_C="$(dirname "$(dirname "$DST")")"          # -> $STAGE/candidate
test -d "$CAND_C/conformance/v1" || { echo "preflight: $CAND_C/conformance/v1 missing"; exit 1; }

COPYFILE_DISABLE=1 tar --no-mac-metadata --no-xattrs --no-acls --no-fflags \
  -cf "$CAND_TAR" -C "$CAND_C" conformance
rc=$?; [ "$rc" = 0 ] || { echo "archive: candidate tar creation failed rc=$rc"; exit 1; }

# Fail-closed listing assertions: the archive really carries the tree.
# The listing is MATERIALIZED to a file and its status checked before anything
# filters it -- `tar -tf … | grep -q` can exit 0 on a tar that failed (F3, I13).
tar -tf "$CAND_TAR" > "$STAGE/candidate.list" || { echo 'archive: listing failed'; exit 1; }
grep -qx 'conformance/v1/manifest.json' "$STAGE/candidate.list" \
  || { echo 'archive does not contain conformance/v1/manifest.json'; exit 1; }
n="$(grep -c '^conformance/v1/.*[^/]$' "$STAGE/candidate.list")"
[ "$n" = "$ACCEPTED_FILE_COUNT" ] || { echo "archive holds $n files, expected $ACCEPTED_FILE_COUNT"; exit 1; }

CAND_TAR_SHA256="$(shasum -a 256 "$CAND_TAR" | cut -d' ' -f1)"
echo "CANDIDATE_TAR_SHA256=$CAND_TAR_SHA256"
```

> **This is F1's fix.** Revision 2 archived with `-C "$(dirname "$DST")"`, i.e.
> `$STAGE/candidate/conformance`, and then named the operand `conformance` — resolving to
> `$STAGE/candidate/conformance/conformance`, which does not exist. Reproduced in a scratch
> directory this run: the revision-2 form exits **1** with `tar: conformance: Cannot stat: No such
> file or directory`; the two-level `dirname` form exits **0** and `tar -tf` lists `conformance/`,
> `conformance/v1/`, `conformance/v1/manifest.json` (§10 rows 24–25). The extraction root in §5.4 W4
> and the Windows-visible root in §5.4 W6 are consistent with this archive shape: extracting into
> `<base>` yields `<base>\conformance\v1`.

**Second payload — the source tree.** `go test` needs the composite module, and neither remote host
has `git`. Ship it the same way. Five exclusions are load-bearing.

The staging is written as **five separately status-checked steps with no pipe anywhere**, and it ends
in a complete-set assertion. Revision 4 used `tar -cf - … | tar -xf - -C "$SRCSTAGE"`, and `set -e`
observes only a pipeline's *last* command: a producer that dies after emitting a valid partial stream
is extracted successfully and the script continues. That is exactly the partial-input class F2
removed from `go list` in the previous cycle, reintroduced on the transport side.

```bash
set -u
SRCSTAGE="$STAGE/src"
STAGE_TAR="$STAGE/src-stage.tar"
SRC_ORIGIN=<composite-worktree>
rm -rf "$SRCSTAGE" "$STAGE_TAR"; mkdir -p "$SRCSTAGE"

# The five exclusions, stated ONCE as anchored './x' patterns so the tar set and
# the find set below cannot drift apart:
#   .git / .temp / .task-board  -- not build inputs, and large.
#   .claude / .codex            -- they hold the only two symlinks in the tree
#     (.claude/skills/skill-go-testing-tools, .codex/skills/skill-go-testing-tools).
#     Extracting a symlink on Windows needs Developer Mode or admin; excluding
#     them removes the failure mode entirely and changes no Go input.

# --- step 1: enumerate the INTENDED set at the origin, before any archiving.
#     This inventory, not the archive, is the authority on completeness -- which
#     is exactly why it may not be produced by a pipeline. `find | sort | xargs`
#     reports only xargs's status, so a find that emits a valid partial stream
#     and then dies yields a SHORT BUT INTERNALLY CONSISTENT inventory that every
#     later check agrees with (F2, I13). Three stages, three materialized files,
#     three separately compared statuses. No `pipefail`: POSIX /bin/sh has none.
( cd "$SRC_ORIGIN" && find . -type f \
    -not -path './.git/*'  -not -path './.temp/*' -not -path './.task-board/*' \
    -not -path './.claude/*' -not -path './.codex/*' \
    -print0 ) > "$STAGE/src-origin.paths0"
rc=$?; [ "$rc" = 0 ] || { echo "stage: origin path enumeration failed rc=$rc"; exit 1; }

LC_ALL=C sort -z < "$STAGE/src-origin.paths0" > "$STAGE/src-origin.sorted0"
rc=$?; [ "$rc" = 0 ] || { echo "stage: origin path sort failed rc=$rc"; exit 1; }

( cd "$SRC_ORIGIN" && xargs -0 shasum -a 256 < "$STAGE/src-origin.sorted0" ) \
  > "$STAGE/src-origin.sha256"
rc=$?; [ "$rc" = 0 ] || { echo "stage: origin digest generation failed rc=$rc"; exit 1; }

SRC_ORIGIN_COUNT="$(grep -c . < "$STAGE/src-origin.sha256")"
[ "$SRC_ORIGIN_COUNT" -gt 0 ] || { echo 'stage: origin enumeration produced no files'; exit 1; }
echo "source origin files $SRC_ORIGIN_COUNT"

# --- step 2: create the archive as its own process. Check ITS status.
COPYFILE_DISABLE=1 tar --no-mac-metadata --no-xattrs --no-acls --no-fflags \
  -cf "$STAGE_TAR" -C "$SRC_ORIGIN" \
  --exclude './.git' --exclude './.temp' --exclude './.task-board' \
  --exclude './.claude' --exclude './.codex' .
rc=$?; [ "$rc" = 0 ] || { echo "stage: source archive creation failed rc=$rc"; exit 1; }

# --- step 3 and 4: list, then extract. Each status-checked, neither piped.
tar -tf "$STAGE_TAR" > "$STAGE/src-stage.list" || { echo 'stage: listing failed'; exit 1; }
tar -xf "$STAGE_TAR" -C "$SRCSTAGE"            || { echo 'stage: extraction failed'; exit 1; }

# --- step 5: the complete-set assertion. `-c` catches changed and MISSING files;
#     the count catches EXTRA ones, which `-c` structurally cannot see.
( cd "$SRCSTAGE" && shasum -a 256 -c --status "$STAGE/src-origin.sha256" ) \
  || { echo 'stage: per-file verification failed (changed or missing)'; exit 1; }
n="$(cd "$SRCSTAGE" && find . -type f | wc -l | tr -d ' ')"
[ "$n" = "$SRC_ORIGIN_COUNT" ] \
  || { echo "stage: staged $n files, origin enumerated $SRC_ORIGIN_COUNT"; exit 1; }
echo "stage: ok, $n of $SRC_ORIGIN_COUNT source files verified"
```

> **This is F3's fix, and both halves were executed this cycle against `/bin/sh` fixtures — no Go,
> no product path** (§7.4c, §10 row 41). Harness case **R**: the revision-4 pipeline, given a
> producer that emits a valid one-file stream and then exits 1, **exits 0** with **1 of 3 files**
> staged. Cases **S/T/U**: the form above **exits 1** on that same producer, **exits 1** when a file
> is deleted after extraction, and **exits 1** when an unexpected file is added. **Step 1's own
> three-stage form was executed this cycle** (§7.4f): the revision-5 pipeline exits **0** with **2 of
> 3** files against a failing `find` (case V), while the form above exits **1** on a
> partial-then-failing `find` (X), a failing `sort` (Y) and a failing digest producer (Z) — and case
> **AA** proves no archive is created when enumeration fails.

The exclusion patterns are anchored `./x` in both the `find` and the `tar` invocation on purpose. If
they ever drift apart the run does not silently ship a different tree: excluding *more* in `tar` than
`find` enumerated fails step 5's `-c` as *missing*, and excluding *less* fails the count as *extra*.
The assertion is the guard on its own exclusion list.

Only then does vendoring run, and the vendor tree is deliberately **outside** the verified set —
`vendor/` is created after step 5, so it never has to be reconciled against the origin inventory.

```bash
# The module has 4 direct + 17 indirect external dependencies and no vendor/ dir,
# so a fresh host would otherwise need proxy.golang.org. Vendor once here: the
# remote lane is then hermetic and deterministic. `go mod vendor` resolves the
# `replace … => ./agents/skills/skill-go-testing-tools/tuitestkit` submodule path,
# which is why the submodule content must already be present in $SRCSTAGE --
# step 5 above has already proved it arrived.
test -f "$SRCSTAGE/agents/skills/skill-go-testing-tools/tuitestkit/go.mod" \
  || { echo 'preflight: submodule content missing from source stage'; exit 1; }

# §5.0: approved root, derived launcher, no toolchain download, no env file.
# A bare `go` here would vendor with whichever launcher PATH happened to win.
( cd "$SRCSTAGE" \
  && GOROOT="$GO_ROOT" GOTOOLCHAIN=local GOENV=off "$GO_EXE" mod vendor ) \
  || { echo 'go mod vendor failed'; exit 1; }
test -d "$SRCSTAGE/vendor" || { echo 'go mod vendor produced no vendor/'; exit 1; }

SRC_TAR="$STAGE/curator-src.tar"
COPYFILE_DISABLE=1 tar --no-mac-metadata --no-xattrs --no-acls --no-fflags \
  -cf "$SRC_TAR" -C "$(dirname "$SRCSTAGE")" src
rc=$?; [ "$rc" = 0 ] || { echo "stage: transport archive creation failed rc=$rc"; exit 1; }
tar -tf "$SRC_TAR" > "$STAGE/curator-src.list" || { echo 'source archive is malformed'; exit 1; }
grep -qx 'src/go.mod'    "$STAGE/curator-src.list" || { echo 'source archive lacks go.mod'; exit 1; }
grep -qx 'src/vendor/'   "$STAGE/curator-src.list" || { echo 'source archive lacks vendor/'; exit 1; }
SRC_TAR_SHA256="$(shasum -a 256 "$SRC_TAR" | cut -d' ' -f1)"
echo "SRC_TAR_SHA256=$SRC_TAR_SHA256"
```

`go mod vendor` runs in `$SRCSTAGE`, a copy — **never in the repository worktree**. `vendor/` must
never be committed (I9). Both `tar -tf` results are written to a file and grepped from the file, not
piped, for the same reason step 3 is: `tar -tf … | grep -q` can exit 0 on a `tar` that failed.

**C4 — the digest handoff.** Write one file and attach it to `TASK-260720-1pvfj5` as a task-scoped
outcome resource **before any transport**. The board copy, not the transported copy, is the
authority a remote host's archive digest is compared against.

```
# $STAGE/bundle-digests.txt   → board resource TASK-260720-1pvfj5_candidate-bundle-digests.txt
CANDIDATE_MANIFEST_SHA256=b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c
CANDIDATE_TREE_SHA256=e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae
CANDIDATE_FILE_COUNT=448
CANDIDATE_TAR_SHA256=<measured at C3, producer fills in>
SRC_TAR_SHA256=<measured at C3, producer fills in>
```

The archive digest is a **transport-integrity check only**. It is verified before extraction and it
never substitutes for the post-extraction tree verification, which is what actually proves the files
on the remote host are the accepted candidate.

### 5.3 Host availability — measured 2026-07-29, this audit

| Host | Platform | Reachable? | Go | Usable for 1pvfj5 candidate evidence? |
|---|---|---|---|---|
| local | macOS arm64 | n/a | **three launchers on `PATH`**, two at `go1.25.5` and one at **`go1.25.1`** — read from `VERSION` files, §5.0 | **yes — the primary candidate host**, *only* through the §5.0 absolute constants. Bare `go` resolved differently in two agent shells this week (§5.0 table); `/usr/local/go` is too old for `go.mod` and would trigger a toolchain **download** under the default `GOTOOLCHAIN=auto`. |
| `ssh relux` | macOS amd64 (`Darwin x86_64`) | **yes, exit 0** | **binary present at `/usr/local/bin/go`**; not on the non-interactive PATH; **version unmeasured** | **conditionally yes** — see preflight R-P1 |
| `ssh win` | Windows 10 Pro 19045 amd64 | **no — `Operation timed out`, exit 255, twice** | unknown; last inventory (2026-07-29 04:40, sibling task) found **no `go.exe`** on process/user/machine PATH, no MSI/uninstall registry entry, no tree at conventional roots | **blocked** on W-P1 and W-P2 |
| `ssh lev` | Ubuntu 26.04 x86_64 | **no — `Operation timed out`, exit 255, twice** | previously absent from PATH and from `/usr/local/go`, `/opt/curator/toolchains/go1.25.12`, `/usr/lib/go`, `/snap/go/current`; distro offers 1.26, outside the accepted `1.25` family | **blocked**, and non-gating regardless |

**R-P1 — the relux preflight, exactly.** `/usr/local/bin/go` exists and is executable, but this
audit ran no Go, so its version is not a fact. Before relux produces any evidence:

```bash
RELUX_GO_EXE=/usr/local/bin/go                       # measured present + executable, §10 row 21

ssh relux "GOTOOLCHAIN=local GOENV=off $RELUX_GO_EXE version"
# must report exactly `go1.25.5 darwin/amd64` — go.mod:3 names 1.25.5, and an older
# 1.25.x would download a toolchain under the default GOTOOLCHAIN=auto (§5.0).

ssh relux "GOTOOLCHAIN=local GOENV=off $RELUX_GO_EXE env GOROOT"
# record the absolute root as RELUX_GO_ROOT; every later relux command exports
# GOROOT="$RELUX_GO_ROOT" and uses "$RELUX_GO_EXE" / "$RELUX_GO_ROOT/bin/gofmt".
```

If the version is not `go1.25.5`, relux is **blocked** on the same operator-installed-toolchain
prerequisite as the other hosts — do not let `auto` fetch the difference. `/usr/local/bin/go` may
itself be a symlink; `env GOROOT` is what resolves the real tree, which is why it is recorded rather
than assumed. Every relux gate command invokes the absolute executable under the §5.0 environment; a
bare `go` is not on the non-interactive PATH at all (§10 row 20), so an unqualified command there
fails as "command not found" rather than silently picking a wrong toolchain — the opposite of the
local host's failure mode, and the reason both need the same discipline for different reasons.

**The named Linux prerequisite, precisely:** an operator-installed, manager-approved absolute
Go 1.25.x `GOROOT` with a trusted `curator-go-toolchain-v1` identity on `ssh lev`, recorded as a
stop-the-line blocker on `TASK-260729-2sxx7k`, **plus** host reachability, which is currently absent.
Linux *platform qualification* is separately owned by `TASK-260728-1skseh`, the `until_task` named
in the rc.5 exclusion vector. Until all three land, **native Linux validation stays non-gating and
must not be attempted** — no auto-install, no download, no ambient PATH. This is distinct from the
hosted `ubuntu-latest` GitHub runner, which does have Go and remains the right host for vet, lint,
and the scoped Linux test and race jobs.

**State this plainly at handoff:** as of 2026-07-29 the candidate-root evidence 1pvfj5 can produce
is **macOS arm64 locally**, plus **macOS amd64 on relux if R-P1 passes**. Windows candidate evidence
is blocked on an operator action and on host reachability. That limitation belongs in the handoff to
`TASK-260720-38l1sy`, not hidden behind a green macOS run.

### 5.4 The Windows lane — one exact path, W1 → W9

This replaces revision 2's `scp`-or-base64 option set. **There is no fallback:** if a preflight step
fails, the lane stops and records the named prerequisite. All commands assume the `ssh win` config
already present (`HostName 100.120.84.42`, `User admin`), so the SSH home directory is
`C:\Users\admin` and the shell is `cmd.exe`.

Fixed paths, chosen once:

```
base (Windows)   C:\Users\admin\AppData\Local\Temp\TASK-260720-1pvfj5
base (scp, rel.) AppData/Local/Temp/TASK-260720-1pvfj5
candidate root   C:\Users\admin\AppData\Local\Temp\TASK-260720-1pvfj5\conformance\v1
source root      C:\Users\admin\AppData\Local\Temp\TASK-260720-1pvfj5\src
```

The lane is one control-host script and shares these constants throughout; `$WBASE` below is the
Windows base path above, and `$CANDIDATE_TAR_SHA256` / `$SRC_TAR_SHA256` are read from the C4 board
resource (§5.2), not re-measured on the remote host:

```bash
WBASE='C:\Users\admin\AppData\Local\Temp\TASK-260720-1pvfj5'
```

`scp` targets are written **relative to the SSH home directory**, never as `win:C:\…` — a
drive-letter target is parsed as a host by `scp`.

**W1 — preflight, fail closed.** Every assertion **compares**; none merely prints. Remote output
arrives with CRLF line endings, so every captured value is passed through `tr -d '\r'` before
comparison — without that strip, every `[ "$got" = "$want" ]` below fails on a healthy host and the
lane is uselessly red. The comparisons run on the **control host**, in `sh`, not inside `cmd.exe`:
that keeps the quoting sane and makes the assertion visible in the local evidence log.

```bash
set -u
ssh -o BatchMode=yes -o ConnectTimeout=10 win "echo ok" > /tmp/w-p1.txt 2>&1 \
  || { echo 'W-P1: ssh win unreachable'; exit 1; }
[ "$(tr -d '\r' < /tmp/w-p1.txt)" = "ok" ] || { echo 'W-P1: unexpected reply'; exit 1; }

# In-box since Win10 1803 / PS 5.1 has Get-FileHash. Presence is asserted, not eyeballed.
got="$(ssh win "where tar" | tr -d '\r')" || { echo 'W1: `where tar` failed'; exit 1; }
case "$got" in *'\tar.exe') ;; *) echo "W1: unexpected tar location: $got"; exit 1;; esac
got="$(ssh win "where powershell" | tr -d '\r')" || { echo 'W1: `where powershell` failed'; exit 1; }
case "$got" in *'\powershell.exe') ;; *) echo "W1: unexpected powershell location: $got"; exit 1;; esac
```

**W-P2 — the approved Windows toolchain, by fixed absolute path.** The operator supplies
`WIN_GOROOT` as a constant; the agent installs nothing and **does not discover it**. Revision 3 ran
`where go` here, which is the same ambient-`PATH` defect §5.0 corrects — worse on Windows, where
`where` also searches the current directory and a stray `go.exe` beside the working directory would
win.

Exactly as on POSIX hosts (§5.0), the operator supplies **one** constant — the root — and the
executable is derived from it. Every probe is compared, not printed.

```bash
# Operator-supplied constant, recorded on the board before the lane runs. Example shape only —
# the real value comes from the operator, and a mismatch fails the lane closed.
WIN_GOROOT='C:\Program Files\Go'
WIN_GO_EXE="$WIN_GOROOT\\bin\\go.exe"          # derived, never supplied separately
WIN_ENV='set "GOTOOLCHAIN=local" & set "GOENV=off" & '

ssh win "if exist \"$WIN_GO_EXE\" (echo GO_EXE_OK) else (echo GO_EXE_MISSING & exit /b 1)" \
  > /tmp/w-p2-exe.txt 2>&1 || { echo "W-P2: $WIN_GO_EXE missing"; exit 1; }
[ "$(tr -d '\r' < /tmp/w-p2-exe.txt)" = "GO_EXE_OK" ] || { echo 'W-P2: unexpected reply'; exit 1; }

got="$(ssh win "$WIN_ENV \"$WIN_GO_EXE\" version" | tr -d '\r')" \
  || { echo 'W-P2: `go version` failed'; exit 1; }
echo "W-P2 toolchain: $got"
[ "$got" = "go version go1.25.5 windows/amd64" ] \
  || { echo "W-P2: expected 'go version go1.25.5 windows/amd64', got '$got'"; exit 1; }

got="$(ssh win "$WIN_ENV \"$WIN_GO_EXE\" env GOROOT" | tr -d '\r')" \
  || { echo 'W-P2: `go env GOROOT` failed'; exit 1; }
echo "W-P2 GOROOT: $got"
[ "$got" = "$WIN_GOROOT" ] \
  || { echo "W-P2: GOROOT drift: reported '$got' != approved '$WIN_GOROOT'"; exit 1; }

got="$(ssh win "$WIN_ENV \"$WIN_GO_EXE\" env GOTOOLCHAIN" | tr -d '\r')" \
  || { echo 'W-P2: `go env GOTOOLCHAIN` failed'; exit 1; }
echo "W-P2 GOTOOLCHAIN: $got"
[ "$got" = "local" ] || { echo "W-P2: GOTOOLCHAIN=$got, not local"; exit 1; }

got="$(ssh win "$WIN_ENV \"$WIN_GO_EXE\" env GOENV" | tr -d '\r')" \
  || { echo 'W-P2: `go env GOENV` failed'; exit 1; }
echo "W-P2 GOENV: $got"
[ "$got" = "off" ] || { echo "W-P2: GOENV=$got, not off"; exit 1; }
```

If the derived path is absent, the version is not `go1.25.5 windows/amd64`, or `env GOROOT` /
`GOTOOLCHAIN` / `GOENV` differ from the approved values, the lane **stops** and records W-P2 as
unsatisfied. There is no discovery fallback and no `where go`. These are the same six T-P1
assertions, expressed for `cmd.exe`; the `go.mod` version self-check is not repeated here because
`$WIN_GOROOT` is checked against the same `go1.25.5` constant that §7's `require-toolchain` reconciles
against `go.mod` on the control host.

**Current status: W-P1 fails.** `ssh win 'echo ok'` timed out with exit 255 twice this run (§10
row 22), so `where tar` / `where powershell` and the W-P2 assertions were never reachable to measure.
W-P2 was already unsatisfied at the last inventory — no `go.exe` on any PATH, no MSI/uninstall
registry entry, no tree at a conventional root — so `WIN_GOROOT` has **no known value today** and the
example above is a shape, not a fact. Both prerequisites are operator actions; the agent installs
nothing.

**W2 — the base directory is an executable empty-root precondition, not a `mkdir`.**

Revision 5 ran a bare `ssh win "mkdir …"` under `set -u` and only *described* what a non-zero exit
means. That is the F3 defect: W4 extracts both archives into this base, and while candidate extras
are caught by W5, **the source tree has no post-extraction equivalent** — so one stale `.go` file
left by an earlier run survives extraction, joins `go test ./...`'s input, and every archive digest
still matches. The base must therefore be proved *absent*, then created, then proved *empty*, and
each of those three steps compared, before anything crosses the wire.

```bash
set -u
WBASE='C:\Users\admin\AppData\Local\Temp\TASK-260720-1pvfj5'
W2LOG=/tmp/w2.txt

# 1. the base must NOT pre-exist. `cmd`'s mkdir would fail on it anyway, but the
#    lane must stop HERE, with a named remedy, not drift into W3.
ssh win "if exist \"$WBASE\" (echo BASE_EXISTS & exit /b 1) else (echo BASE_ABSENT)" > "$W2LOG" 2>&1
rc=$?
[ "$rc" = 0 ] \
  || { echo 'W2: base already exists on the remote host; run W9, confirm absence, then re-run W2'; exit 1; }
[ "$(tr -d '\r' < "$W2LOG")" = 'BASE_ABSENT' ] \
  || { echo "W2: unexpected precondition reply: $(tr -d '\r' < "$W2LOG")"; exit 1; }

# 2. create it, and check THIS command's status
ssh win "mkdir \"$WBASE\"" > "$W2LOG" 2>&1
rc=$?
[ "$rc" = 0 ] || { echo "W2: mkdir failed rc=$rc: $(tr -d '\r' < "$W2LOG")"; exit 1; }

# 3. prove it is EMPTY. `dir /a /b` writes "File Not Found" to stderr and exits 1
#    on an empty directory, so its status is unusable here; Get-FileHash's own
#    PowerShell is already a W1 prerequisite and returns a bare integer.
ssh win "powershell -NoProfile -Command \"(Get-ChildItem -LiteralPath '$WBASE' -Force | Measure-Object).Count\"" > "$W2LOG" 2>&1
rc=$?
[ "$rc" = 0 ] || { echo 'W2: post-create listing failed; the base is not usable'; exit 1; }
got="$(tr -d '\r' < "$W2LOG")"
[ "$got" = '0' ] \
  || { echo "W2: base is not empty after creation ($got entries); a stale file would join the suite input"; exit 1; }
echo 'W2: base created and proved empty'
```

`-Force` is load-bearing: without it `Get-ChildItem` hides hidden and system entries, and a hidden
leftover would count as zero.

**W3 — transport.** Four files, one command, relative destination. It runs **only** after W2 printed
`base created and proved empty`, and its own status is checked.

```bash
scp -O "$STAGE/candidate.tar" "$STAGE/curator-src.tar" \
       "$STAGE/candidate-inventory.sha256" .scripts/verify-candidate.ps1 \
    win:AppData/Local/Temp/TASK-260720-1pvfj5/ \
  || { echo 'W3: transport failed'; exit 1; }
echo 'W3: transport complete'
```

> **Executed this cycle against `/bin/sh` `ssh`/`scp` stubs — no Windows host was contacted** (§7.4g,
> §10 row 55). The W2+W3 control-host block above is what the harness runs, verbatim; only `ssh` and
> `scp` are replaced, and they return the documented replies and statuses against a simulated base.
> A healthy absent base is created, proved empty and transported (case **AK**); a stale pre-existing
> base stops the lane with **nothing crossing the wire** (**AL**); a base that is non-empty after
> creation does the same (**AM**). The `cmd.exe` and PowerShell *syntax* remains unexecuted and is
> covered by producer negatives W-N5 and W-N6.

`-O` forces the legacy SCP protocol, which uses the remote `scp.exe` that ships with OpenSSH for
Windows rather than an SFTP subsystem that may not be registered. `verify-candidate.ps1` is
**transferred, never inlined**: `ssh win "powershell -EncodedCommand <b64>"` dies at exit 1 past the
~8191-char `cmd.exe` limit — a recorded measurement puts the boundary between 6,372 chars (worked)
and 8,660 (failed).

**W4 — archive integrity before extraction, then extract.** **Both digests are compared and both
comparisons gate their own extraction.** Revision 4 printed the two hashes under a "must equal"
comment and extracted unconditionally, so a corrupted or truncated transfer of *either* archive
reached the suite; in particular nothing fail-closed verified `SRC_TAR_SHA256` before the source
under test was used. The constants come from the board resource written at C4 (§5.2), which is the
authority — never from a value re-measured on the remote host.

```bash
set -u
PSHASH() {  # $1 = remote path -> lowercase sha256 on stdout
  ssh win "powershell -NoProfile -Command \"(Get-FileHash -LiteralPath '$1' -Algorithm SHA256).Hash.ToLower()\"" \
    | tr -d '\r'
}

got="$(PSHASH "$WBASE\\candidate.tar")" || { echo 'W4: candidate hash failed'; exit 1; }
echo "W4 candidate.tar sha256 $got"
[ "$got" = "$CANDIDATE_TAR_SHA256" ] \
  || { echo "W4: candidate archive digest mismatch: $got != $CANDIDATE_TAR_SHA256"; exit 1; }

got="$(PSHASH "$WBASE\\curator-src.tar")" || { echo 'W4: source hash failed'; exit 1; }
echo "W4 curator-src.tar sha256 $got"
[ "$got" = "$SRC_TAR_SHA256" ] \
  || { echo "W4: source archive digest mismatch: $got != $SRC_TAR_SHA256"; exit 1; }

# Only now, and only after BOTH comparisons passed, does anything get extracted.
ssh win "tar -xf $WBASE\\candidate.tar -C $WBASE"   || { echo 'W4: candidate extraction failed'; exit 1; }
ssh win "tar -xf $WBASE\\curator-src.tar -C $WBASE" || { echo 'W4: source extraction failed'; exit 1; }
```

`Get-FileHash` is used instead of `certutil -hashfile` because certutil's output formatting varies
across Windows builds (some insert spaces into the hash); `Get-FileHash` returns a bare hex string.
`.Hash` is upper-case hex, so `.ToLower()` is what makes it comparable with `shasum` output — an
unlowered comparison is red on a healthy transfer. Extraction with `-C <base>` yields
`<base>\conformance\v1` and `<base>\src` — matching the archive shape C3 produces. The candidate
archive digest is a transport check only; W5 is what proves the extracted tree is the accepted
candidate. **The source archive has no W5 equivalent on the remote host — its digest comparison here
is the *only* thing standing between a partial transfer and the suite**, which is why it is a hard
gate rather than a printed value.

**W5 — post-extraction whole-tree verification.** This is the Windows equivalent of the POSIX tree
digest, and it is what F2 requires. It is a *set* comparison in both directions, so it catches
changed, missing, **and extra** files without depending on reproducing the POSIX byte stream. Its
seal is that the inventory file's own SHA-256 **is** `CANDIDATE_TREE_SHA256` (§4.2), so a tampered
inventory fails before it is used.

```powershell
# .scripts/verify-candidate.ps1  — new file, owned by TASK-260720-1pvfj5
[CmdletBinding()]
param(
  [Parameter(Mandatory=$true)][string]$Root,
  [Parameter(Mandatory=$true)][string]$Inventory,
  [Parameter(Mandatory=$true)][string]$ExpectedManifest,
  [Parameter(Mandatory=$true)][string]$ExpectedTree,
  [Parameter(Mandatory=$true)][int]$ExpectedCount
)
$ErrorActionPreference = 'Stop'
function Fail([string]$m) { Write-Output "candidate-identity: FAIL $m"; exit 1 }

# 1. the inventory is authentic: its own digest IS the accepted tree identity
$invHash = (Get-FileHash -LiteralPath $Inventory -Algorithm SHA256).Hash.ToLower()
Write-Output "candidate tree     sha256 $invHash"
if ($invHash -ne $ExpectedTree.ToLower()) { Fail "tree $invHash != $ExpectedTree" }

# 2. the manifest identity named by the task brief
$root = (Resolve-Path -LiteralPath $Root).Path.TrimEnd('\')
$manHash = (Get-FileHash -LiteralPath (Join-Path $root 'manifest.json') -Algorithm SHA256).Hash.ToLower()
Write-Output "candidate manifest sha256 $manHash"
if ($manHash -ne $ExpectedManifest.ToLower()) { Fail "manifest $manHash != $ExpectedManifest" }

# 3. whole-tree verification, both directions
$expected = @{}
foreach ($line in [System.IO.File]::ReadAllLines($Inventory)) {
  if ($line -match '^([0-9a-f]{64})\s\s\./(.+)$') { $expected[$Matches[2]] = $Matches[1] }
  elseif ($line.Trim().Length -gt 0)              { Fail "unparsable inventory line: $line" }
}
if ($expected.Count -ne $ExpectedCount) { Fail "inventory lists $($expected.Count), expected $ExpectedCount" }

$actual = @{}
foreach ($f in Get-ChildItem -LiteralPath $root -Recurse -File -Force) {
  $rel = $f.FullName.Substring($root.Length + 1).Replace('\','/')
  $actual[$rel] = (Get-FileHash -LiteralPath $f.FullName -Algorithm SHA256).Hash.ToLower()
}
Write-Output "candidate files           $($actual.Count)"
if ($actual.Count -ne $ExpectedCount) { Fail "extracted tree has $($actual.Count) files, expected $ExpectedCount" }

foreach ($k in $expected.Keys) {
  if (-not $actual.ContainsKey($k))   { Fail "missing file $k" }
  if ($actual[$k] -ne $expected[$k])  { Fail "content mismatch $k" }
}
foreach ($k in $actual.Keys) {
  if (-not $expected.ContainsKey($k)) { Fail "unexpected extra file $k" }
}
Write-Output "candidate-identity: OK $ExpectedCount files"
exit 0
```

```bash
ssh win "powershell -NoProfile -ExecutionPolicy Bypass -File C:\Users\admin\AppData\Local\Temp\TASK-260720-1pvfj5\verify-candidate.ps1 -Root C:\Users\admin\AppData\Local\Temp\TASK-260720-1pvfj5\conformance\v1 -Inventory C:\Users\admin\AppData\Local\Temp\TASK-260720-1pvfj5\candidate-inventory.sha256 -ExpectedManifest b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c -ExpectedTree e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae -ExpectedCount 448"
```

That command is roughly 640 characters — well inside the measured 6,372-char working ceiling. The
script emits **both identities plus the file count** before any gate runs, satisfying I2's "emit
before every candidate gate" clause. Its correctness rests on the four tree properties verified in
§4.2 (no symlinks, no non-regular files, ASCII-only paths, no whitespace in paths), which make
`Get-ChildItem -Recurse -File -Force` enumerate exactly the set `find . -type f` enumerates.

**This script has not been executed.** `ssh win` is unreachable and this audit runs no PowerShell.
It is a producer gate. The producer must run it once against a deliberately corrupted copy — delete
one file, add one file, flip one byte — and confirm **exit 1** in each of the three cases before
trusting an exit 0.

**W6 — the runner script, transferred not inlined.** Two recorded `cmd.exe` traps make an inline
command wrong: `set VAR=1 && …` keeps the trailing space so the value becomes `"1 "` and an
env-gated test silently skips; and `cmd … & echo %ERRORLEVEL%` expands `%ERRORLEVEL%` at *parse*
time and therefore always prints 0 — a trap that once invalidated an entire Windows evidence set.
A batch file avoids both: `set "VAR=value"` quotes the assignment, and a batch file is parsed line
by line as it executes, so `%ERRORLEVEL%` on its own line expands after the previous line ran.

```bat
@echo off
REM .scripts/win/run-candidate.cmd — new file, owned by TASK-260720-1pvfj5
REM The toolchain is TRANSFERRED IN as two operator-approved absolute constants
REM (W-P2). Nothing here resolves `go` from PATH.
setlocal
set "BASE=C:\Users\admin\AppData\Local\Temp\TASK-260720-1pvfj5"
set "GOROOT=%~1"
set "GO_EXE=%~2"

if "%GOROOT%"=="" ( echo USAGE: run-candidate.cmd ^<GOROOT^> ^<GO_EXE^> & exit /b 2 )
if "%GO_EXE%"=="" ( echo USAGE: run-candidate.cmd ^<GOROOT^> ^<GO_EXE^> & exit /b 2 )
if not exist "%GO_EXE%" ( echo TOOLCHAIN_MISSING %GO_EXE% & exit /b 2 )

set "GOTOOLCHAIN=local"
set "GOENV=off"
set "CURATOR_CONFORMANCE_ROOT=%BASE%\conformance\v1"
set "GOFLAGS=-mod=vendor"
cd /d "%BASE%\src"

REM Re-assert identity on the host that will actually run the suite, so the
REM evidence names its own toolchain.
"%GO_EXE%" version
"%GO_EXE%" env GOROOT

"%GO_EXE%" test -count=1 -timeout 30m ./... > "%BASE%\gate-win-candidate.log" 2>&1
set "RC=%ERRORLEVEL%"
echo EXITCODE=%RC%
> "%BASE%\gate-win-exit.txt" echo EXITCODE=%RC%
endlocal & exit /b %RC%
```

`%~1` / `%~2` strip surrounding quotes, so a `GOROOT` containing spaces (`C:\Program Files\Go`) is
handled. The `if "%GOROOT%"==""` guards use the quoted-comparison form, which is the one that
survives an empty or space-bearing value. Exit code **2** is reserved for a toolchain preflight
failure so it is distinguishable from a test failure in the captured `EXITCODE=`.

**The last line is F2's fix, and it is load-bearing.** Revision 4 ended the file with a bare
`endlocal`, so the batch script's exit status was whatever the last command happened to leave, not
`%RC%` — meaning `ssh win "cmd /c …"` could return **0 while the Go suite failed**, and the only
truthful signal was a printed line W8 never retrieved. Microsoft documents `exit /b <exitcode>` as
the operation that sets the status
(<https://learn.microsoft.com/en-us/windows-server/administration/windows-commands/exit>). Two
details make the one-line form the correct one:

- `endlocal & exit /b %RC%` is parsed as a single line **before** either command runs, so `%RC%` is
  substituted while the `setlocal` environment still exists. Splitting it across two lines discards
  `RC` with `endlocal` and exits with an empty argument.
- the redirect-first form `> "file" echo EXITCODE=%RC%` is used instead of
  `echo EXITCODE=%RC%> "file"`, where `cmd` would try to parse the digit adjacent to `>` as a stream
  handle. The file is the durable copy; the printed line stays for the human reading the transcript.

**None of this batch file has been executed.** `ssh win` is unreachable (§5.3) and this audit runs no
`cmd.exe`. The `%ERRORLEVEL%`-at-parse-time and quoted-`set` traps below are recorded prior
measurements from sibling tasks; the `exit /b` contract is from the cited Microsoft documentation.
Both are producer gates, and W-N3/W-N4 are how the producer proves them before trusting an exit 0.

`GOFLAGS=-mod=vendor` is redundant with a present `vendor/` on go 1.25, but it is set explicitly so
the run fails loudly rather than silently reaching for the network if the vendor tree is missing.
`GOTOOLCHAIN=local` is what stops a version mismatch from becoming a download on a host that must
stay hermetic (§5.0).

```bash
scp -O .scripts/win/run-candidate.cmd win:AppData/Local/Temp/TASK-260720-1pvfj5/ \
  || { echo 'W6: runner transfer failed'; exit 1; }

ssh win "cmd /c $WBASE\\run-candidate.cmd \"$WIN_GOROOT\" \"$WIN_GO_EXE\"" \
  > .temp/TASK-260720-1pvfj5/gate-win-stdout.txt 2>&1
WIN_RC=$?
echo "ssh win run-candidate.cmd exit=$WIN_RC"
# EXITCODE=2 means the toolchain preflight failed, not the suite.
```

`$WIN_RC` — the SSH process status — is now **the** authoritative result, because the batch file ends
in `endlocal & exit /b %RC%`. The printed and persisted `EXITCODE=` lines are corroboration, and W8
asserts all three agree rather than choosing between them.

The approved root and executable are **passed in** from the W-P2 constants, never resolved on the
remote host. That is the Windows half of F1's fix: the same two values are asserted in W1, handed to
the runner here, and re-printed by the runner itself, so the evidence names the toolchain that
produced it.

**W7 — no `-race` on Windows.** The Go race detector on Windows requires a C toolchain and this
audit has no measurement that `ssh win` has one. The Windows candidate lane runs without `-race`,
and that absence is **recorded explicitly** in the evidence rather than omitted.

**W8 — retrieve the evidence, and reconcile the three exit signals.**

```bash
scp -O win:AppData/Local/Temp/TASK-260720-1pvfj5/gate-win-candidate.log \
       win:AppData/Local/Temp/TASK-260720-1pvfj5/gate-win-exit.txt \
       .temp/TASK-260720-1pvfj5/ \
  || { echo 'W8: evidence retrieval failed'; exit 1; }

file_rc="$(tr -d '\r' < .temp/TASK-260720-1pvfj5/gate-win-exit.txt | sed -n 's/^EXITCODE=//p')"
out_rc="$(tr -d '\r' < .temp/TASK-260720-1pvfj5/gate-win-stdout.txt | sed -n 's/^EXITCODE=//p')"
echo "W8: ssh=$WIN_RC persisted=$file_rc printed=$out_rc"
[ -n "$file_rc" ] || { echo 'W8: no EXITCODE= in the persisted file'; exit 1; }
[ "$file_rc" = "$WIN_RC" ] && [ "$out_rc" = "$WIN_RC" ] \
  || { echo 'W8: exit-code signals disagree -- the runner contract is broken, do not trust the run'; exit 1; }
```

Three independent signals for one number is not redundancy for its own sake: a disagreement is the
symptom of exactly the F2 defect class (an `exit /b` that was dropped, a `%ERRORLEVEL%` expanded at
parse time, an `scp` that fetched a stale file from an uncleaned previous run). If they disagree the
run is discarded, not interpreted.

**W-N1 … W-N6 — the negative checks that must pass before any Windows exit 0 is trusted.** Each is
run once, by hand, at producer time. An assertion that has never been seen to fail is not known to
work — LOGBOOK 1740's rule, applied to the lane the agent cannot execute from here.

| # | Inject | Expected | Proves |
|---|---|---|---|
| **W-N1** | flip one byte of `candidate.tar` after transport | W4 exits non-zero at the candidate comparison, **before** any extraction | the digest is a gate, not a printout |
| **W-N2** | flip one byte of `curator-src.tar` after transport | W4 exits non-zero at the source comparison, **before** the source is extracted | closes the specific hole F2 named |
| **W-N3** | point the runner at a stub that runs `exit /b 7` in place of `go test` | `$WIN_RC` = 7, persisted file = 7, printed = 7, W8 agrees | `endlocal & exit /b %RC%` really propagates |
| **W-N4** | run W5 against a copy with one file deleted, one added, one byte flipped | exit 1 in **each** of the three cases | the PowerShell verifier detects missing, extra and changed |
| **W-N5** (new, F3) | leave `$WBASE` in place from a previous run with one extra `.go` file in `$WBASE\src`, then start the lane | W2 exits non-zero at step 1 with `base already exists`; **no `scp` runs** — confirm by timestamp or by `Get-ChildItem` showing the base unchanged | a stale source tree cannot reach the suite behind a correct archive digest |
| **W-N6** (new, F3) | after W2 step 2 succeeds, create one file in `$WBASE` before step 3 | W2 exits non-zero at step 3 with `base is not empty after creation`; **no `scp` runs** | the emptiness proof is a gate, not a formality, and `-Force` really sees the entry |

W-N4 restates the check already required in W5; it is listed here so the negatives are one checklist.
Record the real exit code of every injection. **If a negative check passes when it should fail, the
lane's positive result is void** — report that, do not repair it silently. W-N5 and W-N6 are the
Windows-syntax half of what §7.4g already proved for the control-host half: the harness executed the
W2/W3 block verbatim against `ssh`/`scp` stubs and both stale-base shapes stopped before transport
(cases AL, AM), but no `cmd.exe` or PowerShell ran, so these two remain **producer gates** and are
**not claimed executed**.

**W9 — cleanup, status-checked and absence-confirmed.** Revision 5 issued two unchecked commands and
read neither. A `rmdir /s /q` that reports success while an open handle keeps the directory alive is
the exact input W2's precondition then trips over — or worse, a retry that reuses it.

```bash
set -u
W9LOG=/tmp/w9.txt

ssh win "rmdir /s /q \"$WBASE\"" > "$W9LOG" 2>&1
rc=$?
[ "$rc" = 0 ] || { echo "W9: cleanup command failed rc=$rc"; exit 1; }

# The command's own success is not proof. Confirm ABSENCE, with the same
# precondition probe W2 uses, so the two can never disagree.
ssh win "if exist \"$WBASE\" (echo BASE_EXISTS & exit /b 1) else (echo BASE_ABSENT)" > "$W9LOG" 2>&1
rc=$?
[ "$rc" = 0 ] || { echo 'W9: base still present after cleanup; do NOT retry the lane'; exit 1; }
[ "$(tr -d '\r' < "$W9LOG")" = 'BASE_ABSENT' ] || { echo 'W9: unexpected absence reply'; exit 1; }
echo 'W9: base removed and absence confirmed'
```

> **Executed against the same `ssh` stub** (§7.4g): a healthy cleanup confirms absence (case **AN**),
> and a cleanup that **reports success but leaves the base standing** is rejected (**AO**). A retry
> after an unconfirmed cleanup is what W2's precondition exists to refuse.

**W10 — relux uses the same W3–W9 shape** with POSIX equivalents: `scp -O` to `relux:.temp/…`, then
**both** archive digests captured and compared with `[ "$got" = "$CANDIDATE_TAR_SHA256" ] || exit 1`
and `[ "$got" = "$SRC_TAR_SHA256" ] || exit 1` **before** either archive is extracted — the same hard
gate as W4, not a printed value — then `tar -xf`, then from the extracted root both
`shasum -a 256 -c candidate-inventory.sha256` **and** a recomputed tree digest compared with
`ACCEPTED_TREE_SHA256`, **and** a file count compared with `ACCEPTED_FILE_COUNT`: `shasum -c` alone
cannot detect an *extra* file, so the recompute and the count are what close that hole. The remote
gate's exit code is captured directly from `ssh`, which on a POSIX host already propagates the remote
command's status — the `exit /b` problem is `cmd.exe`-specific. Then, under the §5.0 environment with
the R-P1 constants:

```bash
ssh relux "cd <extracted src root> \
  && GOROOT='$RELUX_GO_ROOT' GOTOOLCHAIN=local GOENV=off GOFLAGS=-mod=vendor \
     CURATOR_CONFORMANCE_ROOT='<extracted candidate root>' \
     '$RELUX_GO_EXE' test -count=1 -timeout 30m ./..."
echo "exit=$?"
```

and `rm -rf` for cleanup. No bare `go`, no inherited `GOROOT`.

---

## 6. Exact `.github/workflows/ci.yml` changes

Job by job. `test` splits, `race` is new, `lint`/`interop`/`naming-gate` change minimally. No hosted
candidate job is added (§5.1).

### 6.0 Workflow-level `env` — the pin lives in exactly one place, and so does the toolchain policy

```yaml
env:
  SPEC_PIN: 00b1688a9b2457ca397a0bb550acf47cad8ee967
  # I11 on hosted runners. actions/setup-go only began forcing this in v6.0.0
  # (PR #460); this workflow pins @v5, so without this line hosted jobs run with
  # the default GOTOOLCHAIN=auto and may download a toolchain. See §2.3 D8.
  GOTOOLCHAIN: local
  # I11 assertion 7. Without this the §6.0a identity step fails closed on every
  # job, which is the correct behaviour and the reason the two must land together.
  GOENV: off
```

Replace the literal at the current `ci.yml:28` and `ci.yml:81` with `${{ env.SPEC_PIN }}`. Every new
checkout of the suite references the same value (invariant I5).

`GOTOOLCHAIN: local` is the **D8(b)** resolution from §2.3 and is the whole of 1pvfj5's response to
the action-version drift: no action major moves in this task. It applies to every job in the file,
including `lint`, `interop` and `naming-gate`, which is correct — none of them should be downloading
a toolchain either. **Setting these two variables is not the assertion.** An `env:` block states an
intent; a shim, a wrapper, a container entrypoint or a later `setup-go` change can override it, which
is why §6.0a **reads both back** in every job that runs Go. If a job then fails with
`go.mod requires go >= 1.25.5 (running go1.x.y; GOTOOLCHAIN=local)`, that is `setup-go` having
installed the wrong version — a real defect this line surfaced, not one it caused. Do **not** revert
the line to make such a job green.

### 6.0a `Verify Go toolchain identity` — the step every Go-consuming job runs

**This is F1's fix.** Revision 5's I11 asserted that on hosted runners "every identity assertion still
runs" while relaxing path shape — but no job's YAML contained an assertion at all. `test` ran bare
`gofmt`/`go vet`/`go test` straight after `setup-go@v5`; `interop` was a bare `go test`; `lint` never
looked at the toolchain the action inherits; `test-linux`'s final godriver-rejection command bypassed
the `require-toolchain` its earlier steps reached through `make`. The invariant and the workflow were
two different contracts.

They become one contract by adding **the same step**, verbatim, to `test`, `test-linux`, `race`,
`lint` and `interop`, immediately after `actions/setup-go` and before **any** other Go command.
`naming-gate` consumes no Go and gets no such step.

```yaml
      - name: Verify Go toolchain identity
        shell: bash                    # windows-latest included: git-bash ships in the image
        run: |
          set -u
          REQUIRED=go1.25.5

          # 1. the pinned constant still matches go.mod
          want="go$(awk '/^go[ \t]/{print $2; exit}' go.mod)" \
            || { echo 'toolchain: cannot read go.mod'; exit 1; }
          [ "$want" = "$REQUIRED" ] \
            || { echo "toolchain: go.mod requires $want but this workflow pins $REQUIRED"; exit 1; }

          # 2. the launcher and the formatter, as PATH actually resolves them after setup-go
          goexe="$(command -v go)"     || { echo 'toolchain: no go on PATH after setup-go'; exit 1; }
          fmtexe="$(command -v gofmt)" || { echo 'toolchain: no gofmt on PATH after setup-go'; exit 1; }

          # 3. exact version -- not "the 1.25 family"
          v="$(go version)" || { echo 'toolchain: go version failed'; exit 1; }
          printf '%s\n' "toolchain: $v"
          case "$v" in *" $REQUIRED "*) ;; *) echo "toolchain: expected $REQUIRED, got: $v"; exit 1;; esac

          # 4. absolute reported root. git-bash on windows-latest reports
          #    C:\hostedtoolcache\... here and /c/hostedtoolcache/.../go.exe from
          #    `command -v`; MSYS accepts C:/..., so a separator swap makes both nameable.
          r="$(go env GOROOT)" || { echo 'toolchain: go env GOROOT failed'; exit 1; }
          printf '%s\n' "toolchain: GOROOT=$r"
          rp="$(printf '%s' "$r" | tr '\\' '/')"
          [ -n "$rp" ] || { echo 'toolchain: go env GOROOT is empty'; exit 1; }
          case "$rp" in [A-Za-z]:/*|/*) ;; *) echo "toolchain: reported GOROOT is not absolute: $r"; exit 1;; esac

          # 5. the launcher IS this root's go and the formatter IS this root's gofmt.
          #    Textual first, then -ef (device+inode), which is what makes the assertion
          #    survive /c/... vs C:/... and the .exe suffix without weakening it.
          same() { [ "$1" = "$2" ] || [ "$1" = "$2.exe" ] || [ "$1" -ef "$2" ] || [ "$1" -ef "$2.exe" ]; }
          same "$goexe"  "$rp/bin/go" \
            || { echo "toolchain: launcher $goexe is not $r/bin/go"; exit 1; }
          same "$fmtexe" "$rp/bin/gofmt" \
            || { echo "toolchain: formatter $fmtexe is not $r/bin/gofmt"; exit 1; }

          # 6. no implicit toolchain download -- READ BACK, never assume the env: block
          tc="$(go env GOTOOLCHAIN)" || { echo 'toolchain: go env GOTOOLCHAIN failed'; exit 1; }
          printf '%s\n' "toolchain: GOTOOLCHAIN=$tc"
          [ "$tc" = local ] || { echo "toolchain: GOTOOLCHAIN=$tc, not local"; exit 1; }

          # 7. no per-user go env file in the loop
          ge="$(go env GOENV)" || { echo 'toolchain: go env GOENV failed'; exit 1; }
          printf '%s\n' "toolchain: GOENV=$ge"
          [ "$ge" = off ] || { echo "toolchain: GOENV=$ge, not off"; exit 1; }

          echo 'ci-toolchain-identity: ok'
```

**Why `printf '%s\n'` and not `echo` for the four diagnostics.** A Windows `GOROOT` contains
backslashes, and `echo` in some shells expands `\r`, `\t` and `\b` inside them — measured this cycle:
`…\rootW` printed as `…ootW`. Under `shell: bash` `echo` would not expand them, but the evidence line
is the one place "which toolchain produced this result" is answerable (§5.0), so it is made
shell-independent. This was a real defect in the harness's own stub before it was a note here.

**The `shell: pwsh` alternate**, for a future runner image without git-bash. It is **specified, not
executed** — `pwsh` and `powershell` are both absent from this audit host (`command -v` → exit 1,
§10 row 56), so its syntax is a producer confirmation exactly like the §5.4 `cmd.exe` lane:

```yaml
      - name: Verify Go toolchain identity
        shell: pwsh
        run: |
          $ErrorActionPreference = 'Stop'
          $required = 'go1.25.5'
          function Fail($m) { Write-Output "toolchain: $m"; exit 1 }

          $want = 'go' + ((Select-String -Path go.mod -Pattern '^go\s+(\S+)').Matches[0].Groups[1].Value)
          if ($want -ne $required) { Fail "go.mod requires $want but this workflow pins $required" }

          $goexe  = (Get-Command go    -CommandType Application).Source
          $fmtexe = (Get-Command gofmt -CommandType Application).Source
          $v = (& go version)                 ; Write-Output "toolchain: $v"
          if ($v -notmatch [regex]::Escape(" $required ")) { Fail "expected $required, got: $v" }
          $r = (& go env GOROOT)              ; Write-Output "toolchain: GOROOT=$r"
          if (-not [IO.Path]::IsPathRooted($r)) { Fail "reported GOROOT is not absolute: $r" }

          function Same($actual, $expected) {
            foreach ($c in @($expected, "$expected.exe")) {
              if (Test-Path -LiteralPath $c) {
                if ((Get-Item -LiteralPath $c).FullName -ieq (Get-Item -LiteralPath $actual).FullName) { return $true }
              }
            }
            return $false
          }
          if (-not (Same $goexe  (Join-Path $r 'bin\go')))    { Fail "launcher $goexe is not $r\bin\go" }
          if (-not (Same $fmtexe (Join-Path $r 'bin\gofmt'))) { Fail "formatter $fmtexe is not $r\bin\gofmt" }

          $tc = (& go env GOTOOLCHAIN)        ; Write-Output "toolchain: GOTOOLCHAIN=$tc"
          if ($tc -ne 'local') { Fail "GOTOOLCHAIN=$tc, not local" }
          $ge = (& go env GOENV)              ; Write-Output "toolchain: GOENV=$ge"
          if ($ge -ne 'off')   { Fail "GOENV=$ge, not off" }
          Write-Output 'ci-toolchain-identity: ok'
```

**What this step does and does not cover in `lint`.** `golangci/golangci-lint-action` runs the
linter against the Go that `setup-go` put on `PATH`, so asserting identity *after* `setup-go` and
*before* the action is the only point where the toolchain the action will inherit is nameable. The
action's internal build of the linter binary is not otherwise controllable from this workflow; that
limit is stated rather than papered over.

**One producer confirmation is still owed**, unchanged from §5.0: this audit ran no Go, so the exact
strings `go env GOTOOLCHAIN` and `go env GOENV` print under the `env:` block are asserted from Go's
documented behaviour, not measured. Run the step once on a branch before relying on it; if either
prints something else, fix the assertion — do **not** delete it, and do **not** drop `GOENV: off` from
§6.0 to make the step green.

> **Executed against `/bin/sh` stubs — no Go ran** (§7.4e, §10 row 54). The bash form above is what
> the harness executes, verbatim. It **passes** a healthy POSIX runner (case **AB**) and a
> Windows-shape root where the reported root uses a backslash separator, the binaries carry `.exe`
> and the `PATH` entry is a hard link under another name (**AG** — only the `tr` / `.exe` / `-ef`
> arms can satisfy that). It **fails closed** on `go1.25.1` (**AC**), a wrapper-forced
> `GOTOOLCHAIN=auto` (**AD**), a user `GOENV` file (**AE**), a cross-root `gofmt` on both the POSIX
> and the Windows shape (**AF**, **AH**), a shim launcher outside the reported `GOROOT` (**AI**) and
> a `go.mod` that drifted to 1.25.4 (**AJ**). The one thing the harness cannot model is the
> drive-letter mapping itself (`/c/…` ↔ `C:\…`), because no such path exists on macOS; the separator,
> suffix and inode arms that make it work are all exercised, and the drive-letter arm is a named
> producer confirmation.

### 6.1 `test` — drop `ubuntu-latest`, add the timeout

```yaml
  test:
    name: Test (${{ matrix.os }})
    runs-on: ${{ matrix.os }}
    strategy:
      fail-fast: false
      matrix:
        os: [macos-latest, windows-latest]        # ubuntu moves to the test-linux job
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: true

      - name: Checkout authoritative protocol suite
        uses: actions/checkout@v4
        with:
          repository: relux-works/curator-spec
          ref: ${{ env.SPEC_PIN }}
          path: protocol-spec

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      # §6.0a, verbatim. First Go-touching step in the job, on BOTH runners.
      - name: Verify Go toolchain identity
        shell: bash
        run: |
          # ... the exact body from §6.0a ...

      - name: gofmt check
        if: runner.os != 'Windows'
        run: |
          unformatted="$(gofmt -l cmd internal)"
          if [ -n "$unformatted" ]; then
            echo "gofmt: files need formatting:"; echo "$unformatted"; exit 1
          fi

      - name: go vet
        run: go vet ./...

      - name: go test
        env:
          CURATOR_CONFORMANCE_ROOT: ${{ github.workspace }}/protocol-spec/conformance/v1
        run: go test -count=1 -timeout 30m ./...
```

### 6.2 `test-linux` — new job: scoped execution plus the inventory rejection assertion

```yaml
  test-linux:
    name: Test (ubuntu-latest, inventory-scoped)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: true

      - name: Checkout authoritative protocol suite
        uses: actions/checkout@v4
        with:
          repository: relux-works/curator-spec
          ref: ${{ env.SPEC_PIN }}
          path: protocol-spec

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      # §6.0a, verbatim. Runs before gofmt/vet, which are direct Go commands.
      - name: Verify Go toolchain identity
        shell: bash
        run: |
          # ... the exact body from §6.0a ...

      - name: gofmt check
        run: |
          unformatted="$(gofmt -l cmd internal)"
          if [ -n "$unformatted" ]; then
            echo "gofmt: files need formatting:"; echo "$unformatted"; exit 1
          fi

      # Full breadth on purpose: vet compiles every package, including
      # internal/godriver, without executing a single test. Linux still
      # type-checks the whole tree.
      - name: go vet
        run: go vet ./...

      # TOOLCHAIN_ALLOW_PATH=1: on a hosted runner actions/setup-go owns PATH in a
      # fresh image, so bare `go` IS the identity-verified toolchain (§5.0). On a
      # native host this variable is unset and require-toolchain demands absolute
      # paths. require-toolchain still asserts go1.25.5 and GOROOT here.
      - name: godriver exclusion guard
        env:
          TOOLCHAIN_ALLOW_PATH: "1"
        run: make linux-package-guard

      - name: go test (packages outside rc5-native-control-inventory-v1)
        env:
          TOOLCHAIN_ALLOW_PATH: "1"
          CURATOR_CONFORMANCE_ROOT: ${{ github.workspace }}/protocol-spec/conformance/v1
        run: make test-linux

      # This one is a DIRECT go command, so it does not reach require-toolchain
      # through make. §6.0a above is what covers it -- that is precisely why the
      # identity step is a job-level step and not a Make prerequisite.
      - name: inventory rejection is asserted, not assumed
        env:
          CURATOR_CONFORMANCE_ROOT: ${{ github.workspace }}/protocol-spec/conformance/v1
        run: |
          go test -count=1 -timeout 30m \
            -run 'TestProbeRejectsAnUncoveredPlatformBeforeTheWorker' \
            ./internal/godriver/
```

The last step is what keeps the exclusion honest: Linux does not skip `internal/godriver`, it proves
the package fails closed there. The test exists at `internal/godriver/worker_test.go:434` in the
candidate (§10 row 27); **the producer must re-confirm the name at edit time** — a Go `-run` pattern
matching nothing exits 0, which would make the gate vacuous. `TASK-260729-1y7okj` recorded exactly
that regression class after a test rename. (I8.)

### 6.3 `race` — new job, two runners

```yaml
  race:
    name: Race (${{ matrix.os }})
    runs-on: ${{ matrix.os }}
    strategy:
      fail-fast: false
      matrix:
        include:
          - os: macos-latest      # AC clause: go test -race ./... on the selected supported runner
            target: race-full
          - os: ubuntu-latest     # scope clause: race on at least Linux, named packages
            target: race
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: true

      - name: Checkout authoritative protocol suite
        uses: actions/checkout@v4
        with:
          repository: relux-works/curator-spec
          ref: ${{ env.SPEC_PIN }}
          path: protocol-spec

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      # §6.0a, verbatim.
      - name: Verify Go toolchain identity
        shell: bash
        run: |
          # ... the exact body from §6.0a ...

      - name: go test -race
        env:
          TOOLCHAIN_ALLOW_PATH: "1"          # hosted runner; see §6.2 note and §5.0
          CURATOR_CONFORMANCE_ROOT: ${{ github.workspace }}/protocol-spec/conformance/v1
        run: make ${{ matrix.target }}
```

The job exports `CURATOR_CONFORMANCE_ROOT`; `make` picks it up through `PIN_ROOT ?=` and
`require-pin-root` fails the target if it is empty (§7). That is what makes the local `make race`
in §9.2 and this CI step the same gate — F3's fix.

`windows-latest` is deliberately absent: the Go race detector on Windows requires a C toolchain and
this audit has no measurement that the hosted image satisfies it. If the producer measures that it
does, add the row; if not, **record the absence explicitly** rather than dropping it silently.

**Named risk — this job is specified but unproven.** Two measurements bound it, and they disagree
about comfort rather than about facts:

- `go test -count=1 -race ./...` at Go's **default** alarm: **exit 1**, `internal/install`
  FAIL 603.306 s, `internal/install/atomicity` FAIL 603.701 s (LOGBOOK 1637). This is what
  `-timeout 30m` exists to rescue.
- A **focused** 6-package race gate at `-timeout 45m`: **exit 0**, `internal/install`
  **609.117 s**, `internal/install/atomicity` **1422.407 s**
  (`.temp/TASK-260720-2284br/gates-rework1/gate-race.log`).

So the packages do complete under race — atomicity at **1422.407 s against a 1800 s alarm, 1.27×
headroom**, measured with only five other packages in flight. The AC's gate runs the whole module,
where Go schedules up to `GOMAXPROCS` package binaries concurrently and per-package wall time rises.
**`go test -race ./... -timeout 30m` has never been executed.** `TASK-260720-jrrgw9` is back in
`development` carrying the test-only mitigation; `TASK-260729-3dr6hw` carries the production fix; and
atomicity's cost is byte-identical to the accepted worktree, so it is pre-existing debt rather than a
`jrrgw9` regression (LOGBOOK 1732).

The producer must therefore treat step 3 of §9.1 as having a sibling: **measure `make race-full`
locally before writing this job**, and record per-package wall time, not just the exit code — the
number that matters is what atomicity does with the full module in flight. If it exceeds 1800 s, the
AC's `-race ./...` clause is blocked on `jrrgw9`/`3dr6hw`, not on 1pvfj5. Do not widen the timeout
past 30 m to make it pass — that was explicitly refused upstream ("no timeout override, assertion
deletion, source edit"). Report the real exit code either way.

### 6.4 `lint` — pin the linter version

```yaml
      # §6.0a, verbatim -- after setup-go, BEFORE the action, because the action
      # lints with the Go that setup-go put on PATH (§6.0a's lint note).
      - name: Verify Go toolchain identity
        shell: bash
        run: |
          # ... the exact body from §6.0a ...

      - uses: golangci/golangci-lint-action@v7     # major retained deliberately, §2.3
        with:
          version: v2.12.2       # was: latest
```

`version: latest` is a mutable supply-chain input: a new golangci-lint minor turns the `lint` job
red with no repository change. The action major stays at `@v7` — `.golangci.yml` is schema
`version: "2"`, which v7 supports, and v8/v9 add nothing this task needs (§2.3).

**Correction to revision 4.** It proposed pinning `v2.4.0` because a sibling task had run that
version. `v2.4.0` was released **2025-08-14**; the current release is **v2.12.2**, released
**2026-05-06** (§10 row 47). Pinning v2.4.0 would silently *downgrade* the linter by eight minor
versions against what `version: latest` resolves to today — the opposite of the supply-chain
stability the pin is for. **Pin `v2.12.2`, the version `latest` resolves to now**, so the change is a
freeze rather than a move, and confirm it green locally before committing (§9.2's caveat: this audit
ran no lint command and `golangci-lint` was absent from this machine at the previous cycle).
Recorded trap: a sudden `#nosec`-suppression failure on files byte-identical to the accepted base is
a **stale golangci-lint cache**, not a code defect — run `golangci-lint cache clean` before hunting
the code.

The config itself needs no change. The existing `gosec` excludes (`G306`, `G301`, `G122`, `G703`)
are narrow and documented, and the `_test\.go` gosec exclusion is scoped to tests — the AC's "no
broad suppression for new security code" is satisfied as-is.

### 6.5 `interop` and `naming-gate`

`interop`: replace the literal ref with `${{ env.SPEC_PIN }}`, add `-count=1 -timeout 30m`, **and add
the §6.0a identity step after `setup-go`** — its `go test` is a direct Go command that reaches no
Make target, so without the step it is the one remaining unasserted toolchain in the workflow.
`internal/interop` is a single test-only package (`golden_test.go`) that imports no godriver, so it
stays on `ubuntu-latest`, otherwise unchanged.

`naming-gate`: unchanged, and it gets **no** identity step — it runs `grep`/`rg` over the tree and
consumes no Go toolchain, so an assertion there would be noise, not a gate.

---

## 7. Exact `Makefile` recipes

Additions, plus one changed line in `test`. `check` keeps its current body (§2.2 documents why it
differs from CI; `check-ci` is the truthful mirror).

```make
VERSION ?= dev
LDFLAGS := -X github.com/relux-works/curator/internal/version.value=$(VERSION)

# --- TASK-260720-1pvfj5 additions -------------------------------------------
MODULE          := github.com/relux-works/curator
GODRIVER_PKG    := $(MODULE)/internal/godriver
GO_VERSION_REQUIRED := go1.25.5          # reconciled against go.mod in require-toolchain

# §5.0 / F1. ONE operator input on a native host: the approved absolute GOROOT.
# GO and GOFMT are DERIVED from it so a cross-root pairing cannot be expressed.
# The empty default is what hosted GitHub runners use -- there actions/setup-go
# owns PATH in a fresh image and the step sets TOOLCHAIN_ALLOW_PATH=1, which
# relaxes the path SHAPE only; every identity assertion still runs.
GOROOT_EXPECTED ?=
GO              ?= $(if $(strip $(GOROOT_EXPECTED)),$(strip $(GOROOT_EXPECTED))/bin/go,go)
GOFMT           ?= $(if $(strip $(GOROOT_EXPECTED)),$(strip $(GOROOT_EXPECTED))/bin/gofmt,gofmt)

# Every Go invocation in this file runs under this prefix -- the environment is
# SUPPLIED, not assumed, and require-toolchain reads both settings back.
GOENVPREFIX := GOTOOLCHAIN=local GOENV=off
ifneq ($(strip $(GOROOT_EXPECTED)),)
GOENVPREFIX := GOROOT='$(strip $(GOROOT_EXPECTED))' $(GOENVPREFIX)
endif

# macOS ships `shasum`; Linux ships `sha256sum`. Both print "<hash>  <path>"
# and both accept -c, so every recipe below is portable across the two.
SHA256          := $(shell command -v shasum >/dev/null 2>&1 && echo 'shasum -a 256' || echo 'sha256sum')

GODRIVER_IMPORTERS := $(MODULE)/cmd/curator $(MODULE)/internal/install
RACE_PKGS       := ./internal/transaction/... ./internal/buildcache/... \
                   ./internal/install/... ./internal/closure/... \
                   ./internal/skillspec/... ./internal/whitelist/... \
                   ./internal/scopes/... ./internal/runtimestore/...

# TASK-260729-2kaopg measured `go test ./...` exiting 1 at Go's 10-minute
# default in cmd/curator, twice. 30m is that task's recorded recommendation.
GOTESTFLAGS     := -count=1 -timeout 30m

# The committed-pin conformance root. CI exports CURATOR_CONFORMANCE_ROOT on the
# step, so PIN_ROOT inherits it there; locally the caller passes PIN_ROOT=...
# Either way require-pin-root fails closed when it is empty or not a root.
PIN_ROOT        ?= $(CURATOR_CONFORMANCE_ROOT)

.PHONY: build test fmt lint vet check \
        require-pin-root require-toolchain race race-full test-linux linux-package-guard \
        candidate-digest conformance-candidate check-ci check-ci-linux

require-pin-root:
	@test -n '$(PIN_ROOT)' \
	  || { echo 'PIN_ROOT (or CURATOR_CONFORMANCE_ROOT) is required: pass the committed-pin conformance root'; exit 1; }
	@test -f '$(PIN_ROOT)/manifest.json' \
	  || { echo 'PIN_ROOT=$(PIN_ROOT) is not a conformance root (no manifest.json)'; exit 1; }

# §5.0 / F1. Seven assertions, every one a comparison. The whole recipe is ONE
# shell so `goexe`, `gofmtexe` and `r` are shared across the checks that need
# them. Executably validated against /bin/sh stubs in §7.4 (cases F,G,I-Q).
#   1. the pinned version still matches go.mod
#   2. on a native host: GOROOT_EXPECTED present and absolute, GO/GOFMT absolute
#   3. exact `go version`
#   4. reported GOROOT == GOROOT_EXPECTED, byte for byte (native)
#   5. the launcher IS $GOROOT/bin/go and the formatter IS $GOROOT/bin/gofmt
#   6. GOTOOLCHAIN reads back as `local`  -- no implicit toolchain download
#   7. GOENV reads back as `off`          -- no per-user go env file in the loop
# 6 and 7 are read-BACKS, not just exports: a shim/wrapper launcher can override
# its caller's environment, and this host already runs bare `go` through one.
require-toolchain:
	@set -u; \
	 want="go$$(awk '/^go[ \t]/{print $$2; exit}' go.mod)"; \
	 [ "$$want" = '$(GO_VERSION_REQUIRED)' ] \
	   || { echo "toolchain: go.mod requires $$want but the Makefile pins $(GO_VERSION_REQUIRED)"; exit 1; }; \
	 if [ "$(TOOLCHAIN_ALLOW_PATH)" != "1" ]; then \
	   test -n '$(strip $(GOROOT_EXPECTED))' \
	     || { echo 'toolchain: GOROOT_EXPECTED is required on a native host (see §5.0)'; exit 1; }; \
	   case '$(strip $(GOROOT_EXPECTED))' in /*) ;; *) echo 'toolchain: GOROOT_EXPECTED must be an absolute path; got: $(GOROOT_EXPECTED)'; exit 1;; esac; \
	   case '$(GO)'    in /*) ;; *) echo 'toolchain: GO must be an absolute path; got: $(GO)'; exit 1;; esac; \
	   case '$(GOFMT)' in /*) ;; *) echo 'toolchain: GOFMT must be an absolute path; got: $(GOFMT)'; exit 1;; esac; \
	 fi; \
	 goexe="$$(command -v '$(GO)' 2>/dev/null)" || { echo 'toolchain: GO=$(GO) not found'; exit 1; }; \
	 gofmtexe="$$(command -v '$(GOFMT)' 2>/dev/null)" || { echo 'toolchain: GOFMT=$(GOFMT) not found'; exit 1; }; \
	 test -x "$$goexe"    || { echo "toolchain: $$goexe is not executable"; exit 1; }; \
	 test -x "$$gofmtexe" || { echo "toolchain: $$gofmtexe is not executable"; exit 1; }; \
	 v="$$($(GOENVPREFIX) "$$goexe" version)" || { echo 'toolchain: `go version` failed'; exit 1; }; \
	 echo "toolchain: $$v"; \
	 case "$$v" in *' $(GO_VERSION_REQUIRED) '*) ;; *) echo "toolchain: expected $(GO_VERSION_REQUIRED), got: $$v"; exit 1;; esac; \
	 r="$$($(GOENVPREFIX) "$$goexe" env GOROOT)" || { echo 'toolchain: `go env GOROOT` failed'; exit 1; }; \
	 echo "toolchain: GOROOT=$$r"; \
	 case "$$r" in /*) ;; *) echo "toolchain: reported GOROOT is not absolute: $$r"; exit 1;; esac; \
	 if [ "$(TOOLCHAIN_ALLOW_PATH)" != "1" ]; then \
	   [ "$$r" = '$(strip $(GOROOT_EXPECTED))' ] \
	     || { echo "toolchain: GOROOT drift: reported $$r != approved $(GOROOT_EXPECTED)"; exit 1; }; \
	 fi; \
	 { [ "$$goexe" = "$$r/bin/go" ] || [ "$$goexe" -ef "$$r/bin/go" ]; } \
	   || { echo "toolchain: launcher $$goexe is not $$r/bin/go"; exit 1; }; \
	 { [ "$$gofmtexe" = "$$r/bin/gofmt" ] || [ "$$gofmtexe" -ef "$$r/bin/gofmt" ]; } \
	   || { echo "toolchain: formatter $$gofmtexe is not $$r/bin/gofmt"; exit 1; }; \
	 tc="$$($(GOENVPREFIX) "$$goexe" env GOTOOLCHAIN)" || { echo 'toolchain: `go env GOTOOLCHAIN` failed'; exit 1; }; \
	 echo "toolchain: GOTOOLCHAIN=$$tc"; \
	 [ "$$tc" = "local" ] \
	   || { echo "toolchain: GOTOOLCHAIN=$$tc, not local -- an implicit toolchain download is possible"; exit 1; }; \
	 ge="$$($(GOENVPREFIX) "$$goexe" env GOENV)" || { echo 'toolchain: `go env GOENV` failed'; exit 1; }; \
	 echo "toolchain: GOENV=$$ge"; \
	 [ "$$ge" = "off" ] \
	   || { echo "toolchain: GOENV=$$ge, not off -- a per-user go env file can inject GOFLAGS/GOTOOLCHAIN"; exit 1; }; \
	 echo 'require-toolchain: ok'

test: require-toolchain require-pin-root
	$(GOENVPREFIX) CURATOR_CONFORMANCE_ROOT='$(PIN_ROOT)' $(GO) test $(GOTESTFLAGS) ./...

race: require-toolchain require-pin-root
	$(GOENVPREFIX) CURATOR_CONFORMANCE_ROOT='$(PIN_ROOT)' $(GO) test -race $(GOTESTFLAGS) $(RACE_PKGS)

race-full: require-toolchain require-pin-root
	$(GOENVPREFIX) CURATOR_CONFORMANCE_ROOT='$(PIN_ROOT)' $(GO) test -race $(GOTESTFLAGS) ./...

# rc5-native-control-inventory-v1 covers macOS and Windows only. On every other
# host internal/godriver fails closed before the worker starts, so Linux runs
# every package except that one. The list is DERIVED at run time and never
# hand-maintained: a new package is included automatically.
#
# F2: discovery lives INSIDE the recipe, behind a status-checked assignment.
# `LINUX_PKGS = $(shell $(GO) list ./... | grep -v …)` is exactly what this
# replaces -- Make's $(shell) discards the exit status, so a `go list` that
# printed 3 of 40 packages and then failed became a green 2-package test lane.
# Reproduced with shell stubs in §7.4: that form exits 0, this one exits 2.
test-linux: require-toolchain require-pin-root linux-package-guard
	@rows="$$($(GOENVPREFIX) $(GO) list ./...)" \
	   || { echo 'test-linux: `go list ./...` failed; refusing a partial package set'; exit 1; }; \
	 pkgs="$$(printf '%s\n' "$$rows" | grep -v -x '$(GODRIVER_PKG)')"; \
	 test -n "$$pkgs" || { echo 'test-linux: safe package set is empty'; exit 1; }; \
	 excluded="$$(printf '%s\n' "$$rows" | grep -c -x '$(GODRIVER_PKG)')"; \
	 [ "$$excluded" = "1" ] \
	   || { echo "test-linux: expected exactly 1 excluded package, found $$excluded"; exit 1; }; \
	 total="$$(printf '%s\n' "$$rows" | grep -c .)"; \
	 kept="$$(printf '%s\n' "$$pkgs" | grep -c .)"; \
	 [ "$$kept" = "$$((total - 1))" ] \
	   || { echo "test-linux: $$kept kept of $$total listed; exclusion is not exactly one package"; exit 1; }; \
	 echo "test-linux: $$kept of $$total packages (excluded: $(GODRIVER_PKG))"; \
	 $(GOENVPREFIX) CURATOR_CONFORMANCE_ROOT='$(PIN_ROOT)' $(GO) test $(GOTESTFLAGS) $$pkgs

# Two invariants, both executable, both fail closed.
#  1. the excluded package still exists under the path the exclusion names
#  2. the set of packages that import it has not grown a third member that
#     might drive godriver.Build from a Linux test
#
# F2: both probes materialize COMPLETE `go list` output in a status-checked
# assignment before any filtering. Revision 3 piped `go list` straight into
# `grep -q`, which can exit 0 the moment it sees the match -- before `go list`
# reaches a later package and fails. That guard reported ok on a failing list.
linux-package-guard: require-toolchain
	@rows="$$($(GOENVPREFIX) $(GO) list ./...)" \
	   || { echo 'guard: `go list ./...` failed; cannot validate the Linux exclusion'; exit 1; }; \
	 printf '%s\n' "$$rows" | grep -q -x '$(GODRIVER_PKG)' \
	   || { echo 'guard: $(GODRIVER_PKG) no longer exists; the Linux exclusion is stale'; exit 1; }
	@imports="$$($(GOENVPREFIX) $(GO) list -f '{{.ImportPath}} {{join .Imports " "}} {{join .TestImports " "}} {{join .XTestImports " "}}' ./...)" \
	   || { echo 'guard: `go list -f` failed; cannot validate the importer set'; exit 1; }; \
	 got="$$(printf '%s\n' "$$imports" \
	          | grep '$(GODRIVER_PKG)' | awk '{print $$1}' \
	          | grep -v -x '$(GODRIVER_PKG)' | LC_ALL=C sort | tr '\n' ' ')"; \
	 want="$$(printf '%s\n' $(GODRIVER_IMPORTERS) | LC_ALL=C sort | tr '\n' ' ')"; \
	 [ "$$got" = "$$want" ] \
	   || { echo 'guard: godriver importer set drifted'; echo "  got:  $$got"; echo "  want: $$want"; exit 1; }
	@echo 'linux-package-guard: ok'

# Full candidate identity: manifest digest, whole-tree digest, file count, and
# every file. All four inputs are required; any mismatch aborts. All three
# identities are emitted before conformance-candidate runs a single case.
# CANDIDATE_INVENTORY is the file written by §5.2 C2; its own SHA-256 IS the
# tree digest, which is why one constant seals both.
candidate-digest:
	@test -n '$(CANDIDATE_ROOT)'             || { echo 'CANDIDATE_ROOT is required'; exit 1; }
	@test -n '$(CANDIDATE_INVENTORY)'        || { echo 'CANDIDATE_INVENTORY is required'; exit 1; }
	@test -n '$(CANDIDATE_MANIFEST_SHA256)'  || { echo 'CANDIDATE_MANIFEST_SHA256 is required'; exit 1; }
	@test -n '$(CANDIDATE_TREE_SHA256)'      || { echo 'CANDIDATE_TREE_SHA256 is required'; exit 1; }
	@test -n '$(CANDIDATE_FILE_COUNT)'       || { echo 'CANDIDATE_FILE_COUNT is required'; exit 1; }
	@m="$$($(SHA256) '$(CANDIDATE_ROOT)/manifest.json' | cut -d' ' -f1)"; \
	  t="$$($(SHA256) '$(CANDIDATE_INVENTORY)' | cut -d' ' -f1)"; \
	  n="$$(cd '$(CANDIDATE_ROOT)' && find . -type f | wc -l | tr -d ' ')"; \
	  echo "candidate manifest sha256 $$m"; \
	  echo "candidate tree     sha256 $$t"; \
	  echo "candidate files           $$n"; \
	  [ "$$m" = '$(CANDIDATE_MANIFEST_SHA256)' ] || { echo "IDENTITY MISMATCH: manifest $$m != $(CANDIDATE_MANIFEST_SHA256)"; exit 1; }; \
	  [ "$$t" = '$(CANDIDATE_TREE_SHA256)' ]     || { echo "IDENTITY MISMATCH: tree $$t != $(CANDIDATE_TREE_SHA256)"; exit 1; }; \
	  [ "$$n" = '$(CANDIDATE_FILE_COUNT)' ]      || { echo "IDENTITY MISMATCH: count $$n != $(CANDIDATE_FILE_COUNT)"; exit 1; }
	@cd '$(CANDIDATE_ROOT)' && $(SHA256) -c --status '$(CANDIDATE_INVENTORY)' \
	  || { echo 'IDENTITY MISMATCH: per-file verification failed'; \
	       cd '$(CANDIDATE_ROOT)' && $(SHA256) -c '$(CANDIDATE_INVENTORY)' | grep -v ': OK$$'; exit 1; }
	@echo 'candidate-digest: ok'

# Candidate evidence. Never a release claim: the target name and every log line
# say "candidate".
conformance-candidate: require-toolchain candidate-digest
	$(GOENVPREFIX) CURATOR_CONFORMANCE_ROOT='$(CANDIDATE_ROOT)' $(GO) test $(GOTESTFLAGS) ./...

# Mirrors the CI `test` job as it runs on macOS. Windows omits the gofmt step;
# ubuntu does not run this job at all -- see check-ci-linux.
# $(GOFMT) is the gofmt from the SAME GOROOT as $(GO) (§5.0): a bare `gofmt`
# can come from a different install than the compiler being gated.
check-ci: require-toolchain require-pin-root
	@out="$$($(GOFMT) -l cmd internal)" || { echo 'gofmt: invocation failed'; exit 1; }; \
	 test -z "$$out" || { echo 'gofmt: files need formatting:'; printf '%s\n' "$$out"; exit 1; }
	$(GOENVPREFIX) $(GO) vet ./...
	$(GOENVPREFIX) CURATOR_CONFORMANCE_ROOT='$(PIN_ROOT)' $(GO) test $(GOTESTFLAGS) ./...

# Mirrors the CI `test-linux` job, step for step and in order.
# Sub-makes carry GOROOT_EXPECTED, not GO/GOFMT: one input, derived everywhere.
check-ci-linux: require-toolchain require-pin-root
	@out="$$($(GOFMT) -l cmd internal)" || { echo 'gofmt: invocation failed'; exit 1; }; \
	 test -z "$$out" || { echo 'gofmt: files need formatting:'; printf '%s\n' "$$out"; exit 1; }
	$(GOENVPREFIX) $(GO) vet ./...
	$(MAKE) linux-package-guard GOROOT_EXPECTED='$(GOROOT_EXPECTED)' TOOLCHAIN_ALLOW_PATH='$(TOOLCHAIN_ALLOW_PATH)'
	$(MAKE) test-linux GOROOT_EXPECTED='$(GOROOT_EXPECTED)' TOOLCHAIN_ALLOW_PATH='$(TOOLCHAIN_ALLOW_PATH)' PIN_ROOT='$(PIN_ROOT)'
	$(GOENVPREFIX) CURATOR_CONFORMANCE_ROOT='$(PIN_ROOT)' $(GO) test $(GOTESTFLAGS) \
	  -run 'TestProbeRejectsAnUncoveredPlatformBeforeTheWorker' ./internal/godriver/
```

> **This is F3's fix.** Revision 2's `check-ci` claimed to be an exact CI mirror while omitting the
> conformance root entirely (so every conformance test silently skipped), running `go test ./...`
> unset, and running `linux-package-guard`, which the macOS/Windows `test` job never runs.
> Revision 2's `make race` / `make race-full` likewise omitted the root that the CI race step
> exports. The `require-pin-root` guard makes the root a hard input for every suite target, and
> `check-ci` / `check-ci-linux` now mirror exactly one CI job each.

### 7.1 Changed and unchanged existing targets

| Target | Change |
|---|---|
| `test` | now `require-pin-root` + `-count=1 -timeout 30m` + explicit root. Without the timeout it is red (D6); without the root every conformance test silently skips. |
| `check` | **unchanged body.** It stays `vet test` + `gofmt -l .`. Because `test` now requires `PIN_ROOT`, `check` inherits that requirement — correct, not a regression: a `check` that silently skipped every conformance test was the defect. Its `gofmt -l .` still walks the submodule and still differs from CI; that is why `check-ci` exists. |
| `build`, `fmt`, `vet`, `lint` | unchanged. |

### 7.2 Target → CI job correspondence (every row executable)

| Make target | Mirrors | Exact correspondence |
|---|---|---|
| `test-linux` | `test-linux` job → step "go test (packages outside rc5-native-control-inventory-v1)" | **exact** — the job calls this target |
| `linux-package-guard` | `test-linux` job → step "godriver exclusion guard" | **exact** — the job calls this target |
| `race` | `race` job, matrix row `ubuntu-latest` | **exact** — the job calls this target with the root exported |
| `race-full` | `race` job, matrix row `macos-latest` | **exact** — the job calls this target with the root exported |
| `check-ci` | `test` job **on macOS** (gofmt + vet + go test) | **equivalent**, not called by CI. Windows runs the same job without the gofmt step. |
| `check-ci-linux` | `test-linux` job, all five steps in order | **equivalent**, not called by CI |
| `test` | nothing directly | used by `check`; same command shape as the `test` job's `go test` step |
| `check` | nothing | **intentionally different** — `gofmt -l .` walks the submodule; CI checks `cmd internal` |
| `candidate-digest`, `conformance-candidate` | nothing | **no CI job by design** (§5.1) — native candidate lane only |
| `require-toolchain` | the `Verify Go toolchain identity` step of §6.0a | **equivalent, and now genuinely paired** — the Make target guards native and Make-mediated gates; the §6.0a step guards the hosted jobs' **direct** Go commands, which reach no Make target at all (`test`'s gofmt/vet/test, `test-linux`'s godriver-rejection command, `interop`'s test, the `lint` action's inherited toolchain). Same seven assertions, same failure strings. Native hosts pass one input, `GOROOT_EXPECTED`; hosted Make steps set `TOOLCHAIN_ALLOW_PATH=1`, which relaxes **path shape only** (§7.4d cases G and Q; §7.4e cases AB–AJ) |

**Which artifact covers which job step — no Go command may be uncovered.**

| Job | Go-consuming steps | Covered by |
|---|---|---|
| `test` (macOS, Windows) | gofmt, vet, go test | §6.0a step |
| `test-linux` | gofmt, vet | §6.0a step |
| `test-linux` | `make linux-package-guard`, `make test-linux` | §6.0a step **and** `require-toolchain` |
| `test-linux` | godriver-rejection `go test -run …` | §6.0a step |
| `race` (macOS, Linux) | `make race-full` / `make race` | §6.0a step **and** `require-toolchain` |
| `lint` | `golangci-lint-action` (inherits PATH Go) | §6.0a step, before the action |
| `interop` | `go test ./internal/interop/` | §6.0a step |
| `naming-gate` | none | — (no Go; no step by design) |

**Platform note.** This `Makefile` is Unix-shell only (`$$(…)`, `grep`, `awk`, `find`). `$(SHA256)`
resolves `shasum` on macOS and `sha256sum` on Linux, so `candidate-digest` is portable across both.
The Windows lane never calls `make`: §5.4 W6 invokes `go test` from a `.cmd` file, and §5.4 W5 does
the identity verification in PowerShell.

### 7.3 Linux package inventory (39 of 40)

Listed for review. The `Makefile` derives this at run time and does **not** hard-code it, which is
why new-package drift is structurally impossible; `linux-package-guard` covers the drift that *can*
happen — the godriver importer set growing a third member.

`./cmd/curator`, `./internal/adapters`, `./internal/audit`, `./internal/buildcache`,
`./internal/buildmeta`, `./internal/buildsource`, `./internal/capabilities`, `./internal/closure`,
`./internal/config`, `./internal/devsub`, `./internal/envfiles`, `./internal/gitignore`,
`./internal/gitops`, `./internal/globalbins`, `./internal/hashing`, `./internal/identifiers`,
`./internal/identity`, `./internal/install`, `./internal/install/atomicity`, `./internal/interop`,
`./internal/locale`, `./internal/managerlock`, `./internal/manifest`, `./internal/marker`,
`./internal/mcp`, `./internal/protocoljson`, `./internal/registry`, `./internal/runtimestore`,
`./internal/scopes`, `./internal/shell`, `./internal/skillcheck`, `./internal/skillspec`,
`./internal/snapshot`, `./internal/staging`, `./internal/transaction`, `./internal/ui`,
`./internal/verr`, `./internal/version`, `./internal/whitelist`.

**Excluded: `./internal/godriver` only.** `./internal/interop` is test-only (`golden_test.go`); it is
still a package `go list` reports and `go test` runs.

### 7.4 The recipes and the staging contract were executed against stubs before being written down

LOGBOOK 1740 recorded the cost of the opposite habit: *a shell command written into a plan and never
run is not a plan.* The `require-toolchain`, `test-linux` and `linux-package-guard` recipes above,
**and** the §5.2 C3 source-staging contract, were therefore run — as literal copies, under `make`
and `/bin/sh` — against stubs standing in for `go`. **No Go was executed**: the stub is a `/bin/sh`
script that prints canned `version`, `env GOROOT`, `env GOTOOLCHAIN`, `env GOENV`, `list ./...` and
`list -f` output, and can be told to fail mid-listing or to override the caller's environment the
way a shim launcher does.

**The reviewer can re-run all 41 cases in one command.** The harness is attached to this task as the
outcome resource `TASK-260729-osjeay_verify-recipes.sh` (working copy
`.temp/TASK-260729-osjeay/verify-recipes.sh`, SHA-256
`fcb11c565a04218222c75573fe59147e9a200dfb6d0f26bbaf1f677e69baf9f9`); it is self-contained, writes
only under `$TMPDIR`, cleans up after itself, and prints `expected=` / `actual=` per case:

```
sh .temp/TASK-260729-osjeay/verify-recipes.sh
# last line: ALL 41 EXPECTATIONS MET      harness exit 0
```

The cycle-3 harness had 7 cases and the cycle-4 harness 21. The cycle-4 verdict correctly observed
that "21/21 is true but narrower than the stated staging invariant": it injected failure into the
archive producer but not into `find`, `sort` or the digest producer, and it exercised
`require-toolchain` but not the hosted YAML, which had no assertion at all. This cycle adds
**twenty cases** covering exactly those gaps — six for origin enumeration (§7.4f), nine for the
§6.0a hosted identity step (§7.4e) and five for the §5.4 W2/W3/W9 empty-root precondition (§7.4g).
The full run log is attached as `TASK-260729-osjeay_verify-recipes-cycle5.log`
(SHA-256 `27fded11438df3b372fb8bfa02cdde771e1ba9bab47a22863663d68c8b9f8503`); the cycle-4 log stays
attached unchanged.

> **A defect the harness caught in itself, kept because it is the class this section exists for.**
> The Windows-shape cases were red on first run: `/bin/sh`'s `echo` **expands `\r`** inside
> `…\rootW`, so the stub reported `…ootW` and the comparison failed for a reason that had nothing to
> do with the contract. Both the stub and §6.0a's four diagnostic lines now use `printf '%s\n'`. On a
> real `windows-latest` runner the same expansion would have silently corrupted the recorded `GOROOT`
> in every job's evidence — the one line §5.0 says must answer "which toolchain produced this".

> **Trap the cycle-3 harness itself hit, kept here because it is the same class of defect this
> section is about.** Its first run reported a case red. The recipe was fine; the *harness* was
> wrong — a prefix assignment before a **function** call (`STUB_VER=go1.25.1 run …`) persists in the
> caller's shell in POSIX `sh`, so one case's value leaked into the next. The current harness has an
> explicit `reset_stub` before every case for exactly that reason.

**7.4a — the revision-3 discovery forms, reproduced failing.** `MODE=partialfail` makes the stub
print three package lines, write a `load: parse … syntax error` to stderr, and exit 1 — exactly a
`go list ./...` that dies partway through a broken tree.

| # | Form under test | Real exit | What happened |
|---|---|---|---|
| A | `BAD_PKGS = $(shell $(GO) list ./… \| grep -v -x …)` | **0** | `BAD_PKGS` silently became a **2-package** list and the recipe body ran. A test lane over that variable would have tested 2 of 40 packages and reported green. |
| B | `$(GO) list ./… \| grep -q -x '…godriver'` | **0** | printed `bad-guard: reported ok` against a `go list` that exited 1 — `grep -q` returns as soon as it matches, before the producer reaches its error. |

**7.4b — the revision-4 `require-toolchain`, reproduced accepting all three F1 shapes.** The
verbatim revision-4 recipe, run against two *complete* stub roots (`rootA`, `rootB`) — both really
contain an executable `bin/gofmt`, which is precisely why the "some gofmt exists under the reported
root" check passed.

| # | Injected defect | Real exit | What revision 4 printed |
|---|---|---|---|
| C | `GO` from **rootA**, `GOFMT` from **rootB** | **0** | `rev4 require-toolchain: ACCEPTED` — a compiler gated by a formatter from a different install |
| D | launcher's real `GOROOT` is **rootB**, operator approved **rootA** | **0** | `toolchain: GOROOT=…/rootB` then `ACCEPTED` — it printed the drift and accepted it |
| E | a wrapper forces **`GOTOOLCHAIN=auto`** | **0** | `ACCEPTED` — revision 4 never read the value, so an implicit toolchain download stayed possible |

**7.4c — the revision-4 source staging, reproduced masking a failed producer.**

| # | Form under test | Real exit | What happened |
|---|---|---|---|
| R | `set -e; producer \| tar -xf - -C "$DEST"` where the producer emits a valid 1-file stream then exits 1 | **0** | `rev4 staging: completed, 1 of 3 files present` — `set -e` saw only the extractor's 0 |

**7.4d — the corrected forms, executed.** Same stubs, same `make`, recipes copied from §7 and
§5.2 C3. *Verbatim* is meant literally and was checked: `diff` of the 39-line `require-toolchain`
recipe as it appears in §7 against the copy the harness executed reports **exactly one differing
line** — the message string `(see §5.0)` in the map versus `(see 5.0)` in the harness, where the
section sign was dropped to keep the shell heredoc ASCII-clean. The `GOROOT_EXPECTED` / `GO` /
`GOFMT` / `GOENVPREFIX` definitions are byte-identical in both.

| # | Scenario | Real exit | Decisive output |
|---|---|---|---|
| F | healthy **native** host | **0** | `GOTOOLCHAIN=local`, `GOENV=off`, `require-toolchain: ok`, `test-linux: 3 of 4 packages (excluded: …/internal/godriver)`, then `go test` over exactly the 3 kept packages |
| G | healthy **hosted** runner (setup-go shape: `TOOLCHAIN_ALLOW_PATH=1`, root on `PATH`) | **0** | `require-toolchain: ok` — the documented exception works and is not a blanket bypass |
| H | `go list` fails mid-listing | **2** | `guard: go list ./... failed; cannot validate the Linux exclusion` |
| I | native host with **no `GOROOT_EXPECTED`** | **2** | `GOROOT_EXPECTED is required on a native host` |
| J | native host, **relative `GO=go`** | **2** | `GO must be an absolute path; got: go` |
| K | launcher reports **`go1.25.1`** | **2** | `expected go1.25.5, got: go version go1.25.1 darwin/arm64` |
| L | reported `GOROOT` = rootB, approved = rootA | **2** | `GOROOT drift: reported …/rootB != approved …/rootA` |
| M | `GOFMT` from a **different root** than the launcher | **2** | `formatter …/rootB/bin/gofmt is not …/rootA/bin/gofmt` |
| N | wrapper-forced **`GOTOOLCHAIN=auto`** | **2** | `GOTOOLCHAIN=auto, not local -- an implicit toolchain download is possible` |
| O | wrapper-forced **user `GOENV` file** | **2** | `GOENV=/…/go/env, not off -- a per-user go env file can inject GOFLAGS/GOTOOLCHAIN` |
| P | `go.mod` drifts to **1.25.4** while the `Makefile` pins `go1.25.5` | **2** | `go.mod requires go1.25.4 but the Makefile pins go1.25.5` |
| Q | **hosted** exception + `GOFMT` from a different root | **2** | same formatter error — **the exception does not bypass identity** |
| S | corrected staging, same failed producer as R | **1** | `stage: source archive creation failed rc=1` |
| T | corrected staging, one file **deleted** after extraction | **1** | `stage: per-file verification failed (changed or missing)` |
| U | corrected staging, one file **added** after extraction | **1** | `stage: destination has 4 files, origin enumerated 3` |

F and G are the only exit-0 rows in 7.4d, and both are cases that *should* pass. **Every row from H
to U is a way revision 3 or revision 4 could have produced a green gate over a wrong, partial or
mis-toolchained input.** Case Q is the one the cycle-4 verdict asked for specifically: it proves the
hosted-runner escape hatch relaxes *path shape only* and still fails closed on a cross-root pairing.

**7.4e — the §6.0a hosted toolchain-identity step (this cycle's F1).** The bash body from §6.0a,
executed verbatim as a file, driven by the same `/bin/sh` stubs. `rootW` is a second stub root whose
binaries carry `.exe` and whose `PATH` entry is a **hard link under a different directory and name**,
so the textual arm cannot succeed and only the separator/`.exe`/`-ef` arms can.

| # | Scenario | Real exit | Decisive output |
|---|---|---|---|
| AB | healthy POSIX hosted runner | **0** | `GOTOOLCHAIN=local`, `GOENV=off`, `ci-toolchain-identity: ok` |
| AC | launcher reports `go1.25.1` | **1** | `expected go1.25.5, got: go version go1.25.1 …` |
| AD | wrapper-forced `GOTOOLCHAIN=auto` | **1** | `GOTOOLCHAIN=auto, not local` |
| AE | wrapper-forced user `GOENV` file | **1** | `GOENV=/…/go/env, not off` |
| AF | `gofmt` from a different root, ahead on `PATH` | **1** | `formatter …/fmtonly/gofmt is not …/rootA/bin/gofmt` |
| AG | **Windows-shape root**: `\` separator, `.exe` binaries, hard-linked `PATH` entry | **0** | `GOROOT=…\rootW` printed intact, then `ci-toolchain-identity: ok` |
| AH | Windows-shape root **plus** a cross-root `gofmt` | **1** | `formatter …/winpath-badfmt/gofmt is not …\rootW/bin/gofmt` |
| AI | goenv-style shim launcher outside the reported `GOROOT` | **1** | `launcher …/shim/go is not …/rootA/bin/go` |
| AJ | `go.mod` drifts to 1.25.4 | **1** | `go.mod requires go1.25.4 but this workflow pins go1.25.5` |

The one arm the harness cannot model is the drive-letter mapping itself (`/c/…` ↔ `C:\…`), because
no such path exists on macOS; the separator, `.exe`-suffix and inode arms that make it work are each
exercised, and the drive-letter arm is a named producer confirmation. The `shell: pwsh` alternate is
**not executed at all** — neither `pwsh` nor `powershell` exists on this host (§10 row 56).

**7.4f — origin enumeration (this cycle's F2).** `find` is replaced by a stub that emits a valid
**partial** NUL stream and then exits 1 — the enumeration analogue of the archive producer in 7.4c.
`sort` and the digest command get their own failing stubs.

| # | Form under test | Real exit | What happened |
|---|---|---|---|
| V | rev-5 `find … -print0 \| sort -z \| xargs -0 shasum` | **0** | `rev5 enumeration: completed, 2 of 3 files` — a **short but internally consistent** inventory, which in §5.2 C3 *is* the completeness authority |
| W | corrected three-stage enumeration, healthy origin | **0** | `stage: origin enumerated 3 files` |
| X | corrected form, partial-then-failing `find` | **1** | `stage: origin path enumeration failed rc=1` |
| Y | corrected form, failing `sort` | **1** | `stage: origin path sort failed rc=1` |
| Z | corrected form, failing digest producer | **1** | `stage: origin digest generation failed rc=123` |
| AA | full staging with a failing `find` | **0** (assertion wrapper) | staging exited non-zero **and no archive file was produced** — the failure lands before anything is packaged |

`pipefail` is deliberately not used anywhere: POSIX `/bin/sh` does not provide it, and the harness
runs under `/bin/sh` for exactly that reason.

**7.4g — the §5.4 W2/W3/W9 empty-root precondition (this cycle's F3).** The control-host block is
**verbatim**; `ssh` and `scp` are `/bin/sh` stubs returning the documented replies and statuses
against a simulated base directory. **No Windows host was contacted and no `cmd.exe` or PowerShell
ran** — what is validated is the guard's ordering and status handling, not Windows syntax, which
stays covered by producer negatives W-N5/W-N6.

| # | Scenario | Real exit | Decisive output |
|---|---|---|---|
| AK | absent base | **0** | `W2: base created and proved empty`, then `W3: transport complete` |
| AL | base pre-exists with a stale `.go` file | **0** (assertion wrapper) | `W2: base already exists …` and **nothing crossed the wire** |
| AM | base non-empty immediately after creation | **0** (assertion wrapper) | `W2: base is not empty after creation (1 entries)` and **nothing crossed the wire** |
| AN | healthy W9 cleanup | **0** | `W9: base removed and absence confirmed` |
| AO | cleanup reports success but the base survives | **1** | `W9: base still present after cleanup; do NOT retry the lane` |

These exit codes describe the **stub harness**, not the Curator suite. They prove the recipes' and
steps' control flow and fail-closed behaviour, nothing about whether Curator's tests pass. The
producer still runs every §9 gate for real. Three contracts in this document remain **unproven by
execution** and are flagged as such where they appear: the exact strings `go env GOTOOLCHAIN` /
`go env GOENV` print under an explicit export (§5.0 T-P1 and §6.0a, asserted from Go's documented
behaviour); the `shell: pwsh` alternate in §6.0a (no PowerShell on this host); and the entire
`cmd.exe` lane (§5.4, no Windows host reachable) — for which W-N1…W-N6 are the producer's substitute.

---

## 8. Executable platform matrix

**G** = gating required check, **N** = non-gating. Every row is a *future* producer gate; none has
been run.

**Hosted runner images, measured 2026-07-29 (§2.3):** `ubuntu-latest` = Ubuntu 24.04 **x64**,
`macos-latest` = macOS 26 **arm64**, `windows-latest` = Windows Server 2025 **x64**. The labels are
moving on purpose; the producer records the concrete image version in every evidence line so a later
red job can be attributed to an image migration rather than to a code change.

| # | Lane | Runner label / host | Command | Root | Gate | Status |
|---|---|---|---|---|---|---|
| 1 | `test` | `macos-latest` (hosted) | `gofmt -l cmd internal`, `go vet ./...`, `go test -count=1 -timeout 30m ./...` | committed pin `00b1688a` | **G** | executable once D3 resolves |
| 2 | `test` | `windows-latest` (hosted) | `go vet ./...`, `go test -count=1 -timeout 30m ./...` | committed pin | **G** | executable once D3 resolves; see §8.2 risk |
| 3 | `test-linux` | `ubuntu-latest` (hosted) | `go vet ./...` (full breadth), `make linux-package-guard`, `make test-linux`, inventory rejection test | committed pin | **G** | executable once D1 + D3 resolve |
| 4 | `race` | `macos-latest` (hosted) | `make race-full` → `go test -race -count=1 -timeout 30m ./...` | committed pin | **G** | satisfies the AC verbatim, **never executed**. Bounding measurements: exit 1 at the default alarm; exit 0 focused at 45 m with atomicity at 1422.407 s — **1.27× headroom** against 1800 s (§6.3) |
| 5 | `race` | `ubuntu-latest` (hosted) | `make race` → 8 named package groups (includes `./internal/install/...`) | committed pin | **G** | satisfies the scope verbatim; carries the same atomicity headroom risk, and its narrower package set is where that risk is *smallest* |
| 6 | `lint` | `ubuntu-latest` (hosted) | `golangci-lint run` at a pinned version | — | **G** | vet-class only, Linux-safe |
| 7 | `interop` | `ubuntu-latest` (hosted) | `go test ./internal/interop/ -v -count=1 -timeout 30m` | committed pin | **G** | no godriver import |
| 8 | `naming-gate` | `ubuntu-latest` (hosted) | inline `grep` | — | **G** | unchanged |
| 9 | candidate | **local macOS arm64** | `make conformance-candidate CANDIDATE_ROOT=… CANDIDATE_INVENTORY=… CANDIDATE_MANIFEST_SHA256=b6f56aac…04c CANDIDATE_TREE_SHA256=e6a13215…2fae CANDIDATE_FILE_COUNT=448` | frozen rc.5 snapshot | **N** | **the primary candidate lane** |
| 10 | candidate | `ssh relux` (macOS amd64) | §5.4 W10, `/usr/local/bin/go` by absolute path | frozen rc.5 snapshot | **N** | reachable; **gated on R-P1** (Go version unmeasured) |
| 11 | candidate | `ssh win` (Windows amd64) | §5.4 W1–W9, no `-race` (W7) | frozen rc.5 snapshot | **N** | **blocked** — W-P1 host unreachable (exit 255, twice, 2026-07-29) and W-P2 no approved Go |
| 12 | native Linux | `ssh lev` (Ubuntu 26.04) | — | — | **deferred** | **blocked** — unreachable (exit 255, twice), no approved Go, and gated on TASK-260728-1skseh |

Rows 9–12 are **not** CI jobs. There is no hosted candidate job (§5.1).

### 8.1 Why macOS and Windows are both mandatory

Two protocol-defined platform-exclusive controls exist nowhere else:

- `TestBuildFailsClosedWhenTheGoChildCannotStart` — `boundary_test.go:267`, Windows only.
- `TestPerFileSizeLimitIsReallyApplied` — `build_test.go:433`, macOS only.

Windows additionally holds the only DACL / reparse / `.cmd` coverage:
`internal/buildcache/collect_windows_test.go`, `internal/buildcache/protection_windows_test.go`,
`internal/runtimestore/targets_windows_test.go`, `internal/scopes/gc_conservative_windows_test.go`,
`internal/scopes/gc_integration_windows_test.go`, `internal/managerlock/identity_windows_test.go`,
`internal/transaction/validation_windows_test.go`.

Unix holds the only ownership / no-follow / readonly-source / executable coverage:
`internal/buildsource/buildsource_special_unix_test.go`, `internal/buildcache/collect_unix_test.go`,
`internal/buildcache/protection_unix_test.go`, `internal/godriver/fingerprint_unix_test.go`,
`internal/transaction/root_metadata_unix_test.go`, `internal/adapters/fifo_unix_test.go`,
`internal/runtimestore/targets_unix_test.go`, `internal/scopes/gc_conservative_unix_test.go`.

Dropping either runner loses a class the AC explicitly requires.

### 8.2 Known Windows risk the producer must triage early

On 2026-07-28 a native `ssh win` full-suite run exited non-zero in **five packages this task does
not own** — `buildcache` (`owner does not match the effective user` on temp roots), `buildsource`
(Windows rename/path semantics), `globalbins` and `runtimestore` (`script command is not
executable`, POSIX exec-bit assumption), `shell` (PowerShell hook). All passed on both macOS hosts.
Recorded as "Reported, not fixed."

Three of those packages have since been reworked by tasks that are now `done` (`2284br`, `1ljev5`,
`1nlmvv`), but **no post-fix Windows measurement exists** — `ssh win` lost its Go before one could
be taken, and as of this audit the host is not even reachable. Row 2 (`windows-latest`, hosted) is
therefore at real risk of being red for reasons outside 1pvfj5's ownership. Measure it early
(§9.1 step 5). If it is still red, that is a separate stop-the-line, not something to absorb into
the CI task.

---

## 9. Producer gate commands

**None of these has been executed. Every one is a future gate.** Record the real exit code of each.

### 9.1 Order — cheapest disproof first

1. **Compose on the right base.** New worktree from `origin/main` = `17804ce`, **not** `c06aa1a`
   (§1.1 C1). Apply the accepted `2kaopg` composite, then the 23-file `jrrgw9` delta. Then:
   ```bash
   grep -n 'ref:' .github/workflows/ci.yml     # must show 00b1688a… twice, unchanged
   ```
2. **Freeze and seal the candidate** — §5.2 C1 → C4. A mismatch at C2 stops the task and escalates
   (§5.2); it is not re-baselined.
3. **Disprove or confirm D3 (largest, cheapest).** One command, macOS, seconds:
   ```bash
   GOROOT="$GO_ROOT" GOTOOLCHAIN=local GOENV=off \
   CURATOR_CONFORMANCE_ROOT=<pin-checkout>/conformance/v1 \
     "$GO_EXE" test -count=1 -timeout 30m \
       ./internal/buildsource/ ./internal/buildcache/ ./internal/scopes/ \
       ./internal/marker/ ./internal/whitelist/ ./internal/skillspec/
   echo "exit=$?"
   ```
   Predicted non-zero with `no such file or directory` in each. **A non-zero exit here is the
   expected-red confirmation of D3, not a defect** — report it with that rationale and do not
   present it as passing. If it exits 0, D3 is wrong and evaporates; say so plainly.
4. **Measure the race lane before writing the race job** (D6, §6.3). The AC clause rides on it, and
   the closest existing measurement is a *focused* pass with 27 % headroom, not this gate:
   ```bash
   # The exact AC gate. Not a focused subset -- the whole module, which is what
   # has never been run. Full paths per §5.0.
   GOROOT="$GO_ROOT" GOTOOLCHAIN=local GOENV=off \
   CURATOR_CONFORMANCE_ROOT=<pin-root> \
     "$GO_EXE" test -race -count=1 -timeout 30m ./... \
       > .temp/TASK-260720-1pvfj5/gate-race-full.log 2>&1
   echo "exit=$?"
   grep -E '^(ok|FAIL|---)' .temp/TASK-260720-1pvfj5/gate-race-full.log
   ```
   **Record per-package elapsed time, not just the exit code.** The number that decides this is
   `internal/install/atomicity`: it measured **1422.407 s focused** against a 1800 s alarm, and the
   full-module run puts more binaries in flight. Anything past 1800 s means the AC's `-race ./...`
   clause is blocked on `TASK-260720-jrrgw9` / `TASK-260729-3dr6hw`, not on 1pvfj5. **Do not raise
   the timeout past 30 m to make it pass** — a timeout override was explicitly refused upstream.
5. **Measure Windows early** (§8.2), if and only if W-P1 (reachability) and W-P2 (an operator-
   supplied approved Go root) are both satisfied. Neither is today.
6. **Disprove or confirm D1.** Requires reachability and **one operator-approved absolute Linux
   `GOROOT` whose launcher reports exactly `go1.25.5`** on `ssh lev` — the same contract as every
   other host (§5.0, I10, I11), not a looser "1.25.x" one. Revision 5 wrote a bare `go test` here,
   with no `GOROOT`, no `GOTOOLCHAIN=local` and no `GOENV=off`; that contradicted the section it
   sits in and is corrected:
   ```bash
   # Operator-supplied constant, recorded on the board before the run. The agent
   # installs nothing and does not discover it. LEV_GO_EXE is DERIVED (§5.0).
   LEV_GO_ROOT=<operator-approved absolute Linux GOROOT>
   LEV_GO_EXE="$LEV_GO_ROOT/bin/go"

   # Full §5.0 preflight FIRST, on lev, all six assertions comparing:
   ssh lev "set -u; \
     test -x '$LEV_GO_EXE' || { echo 'L-P1: launcher missing'; exit 1; }; \
     E=\"GOROOT=$LEV_GO_ROOT GOTOOLCHAIN=local GOENV=off\"; \
     v=\$(env \$E '$LEV_GO_EXE' version) || exit 1; echo \"toolchain: \$v\"; \
     case \"\$v\" in *' go1.25.5 '*) ;; *) echo \"expected go1.25.5, got: \$v\"; exit 1;; esac; \
     r=\$(env \$E '$LEV_GO_EXE' env GOROOT) || exit 1; echo \"toolchain: GOROOT=\$r\"; \
     [ \"\$r\" = '$LEV_GO_ROOT' ] || { echo \"GOROOT drift: \$r\"; exit 1; }; \
     [ \"$LEV_GO_EXE\" = \"\$r/bin/go\" ] || { echo 'launcher is not \$GOROOT/bin/go'; exit 1; }; \
     tc=\$(env \$E '$LEV_GO_EXE' env GOTOOLCHAIN) || exit 1; echo \"toolchain: GOTOOLCHAIN=\$tc\"; \
     [ \"\$tc\" = local ] || { echo 'GOTOOLCHAIN not local'; exit 1; }; \
     ge=\$(env \$E '$LEV_GO_EXE' env GOENV) || exit 1; echo \"toolchain: GOENV=\$ge\"; \
     [ \"\$ge\" = off ] || { echo 'GOENV not off'; exit 1; }" \
     || { echo 'L-P1 unsatisfied: no approved go1.25.5 root on lev'; exit 1; }

   # Only then the gate itself, through the SAME absolute launcher and environment.
   ssh lev "cd <extracted src root> && \
     GOROOT='$LEV_GO_ROOT' GOTOOLCHAIN=local GOENV=off GOFLAGS=-mod=vendor \
     CURATOR_CONFORMANCE_ROOT=<extracted candidate root> \
     '$LEV_GO_EXE' test ./internal/godriver/ -count=1 -timeout 30m"
   echo "exit=$?"
   ```
   Expected on macOS (the same command shape with `$GO_EXE`): 0. Expected on Linux: non-zero with
   `build_execution_control_unavailable` — an **expected-red** result, to be reported as such with
   its real exit code. Until the host is reachable and has an approved `go1.25.5` root this step
   **cannot run**; say that rather than inferring the outcome, and do not substitute an ambient `go`.
7. **Take the D1–D6 decisions** with the board owner. Record each choice and its rationale.
8. **Edit.** `ci.yml` (including the workflow-level `GOTOOLCHAIN: local` from D8(b) and the
   `golangci-lint` `version: v2.12.2` pin), `Makefile`, `.scripts/verify-candidate.ps1`,
   `.scripts/win/run-candidate.cmd` (+ the six test files if D3 → P1). **No action major moves**
   (§2.3). Apply I1–I15.
9. **Validate** §9.2 locally, §9.3 natively.
10. **Attach** exact commands, real exit codes, elapsed times, all three candidate identities,
   per-platform evidence, and cleanup evidence. No release claim.
11. **Hand off** to review with the candidate CI evidence `TASK-260720-38l1sy` needs, including an
    explicit statement of which platforms could not be measured and why.

### 9.2 Local macOS gate set

`PIN_ROOT` is the committed-pin conformance checkout; without it every suite target fails closed at
`require-pin-root`. That is what makes these commands the same gates CI runs — F3's fix.

```bash
# --- §5.0: ONE operator-approved constant. Everything else is derived. ---
GO_ROOT=/opt/homebrew/Cellar/go/1.25.5/libexec
GO_EXE="$GO_ROOT/bin/go"                            # derived
GOFMT_EXE="$GO_ROOT/bin/gofmt"                      # derived
export GOROOT="$GO_ROOT" GOTOOLCHAIN=local GOENV=off

PIN=<pin-checkout>/conformance/v1
CAND=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1pvfj5/candidate/conformance/v1
INV=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1pvfj5/candidate-inventory.sha256

# T-P1 first, in full (§5.0): all six assertions, each comparing. Nothing below
# is interpretable without knowing which toolchain ran it.
make require-toolchain GOROOT_EXPECTED="$GO_ROOT"   # exit 0 required

"$GO_EXE" vet ./...                                 # exit 0 required
"$GOFMT_EXE" -l cmd internal                        # empty output required
golangci-lint run                                   # exit 0 required — see caveat

MK="make GOROOT_EXPECTED=$GO_ROOT"                  # one input; GO/GOFMT derive from it

$MK linux-package-guard                             # exit 0 required
$MK check-ci   PIN_ROOT="$PIN"                      # mirrors the macOS `test` job
$MK race-full  PIN_ROOT="$PIN"                      # mirrors race (macos-latest)
$MK race       PIN_ROOT="$PIN"                      # mirrors race (ubuntu-latest), run locally

$MK conformance-candidate \
     CANDIDATE_ROOT="$CAND" \
     CANDIDATE_INVENTORY="$INV" \
     CANDIDATE_MANIFEST_SHA256=b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c \
     CANDIDATE_TREE_SHA256=e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae \
     CANDIDATE_FILE_COUNT=448
```

Omitting `GOROOT_EXPECTED=` is not a shortcut — `require-toolchain` rejects it with
`GOROOT_EXPECTED is required on a native host` and exit 2 (§7.4d row I), and passing a relative `GO`
is rejected too (row J). That is deliberate: on this host bare `go` resolved to two different
toolchains in two agent shells this week, and one of the three launchers on `PATH` is too old for
`go.mod`. Passing `GO=`/`GOFMT=` **separately** is no longer the interface at all — supplying them
independently is what let revision 4 accept a cross-root pairing (§7.4b case C).

**Note the architecture.** This host is macOS **arm64**, and `macos-latest` is also **arm64**
(§2.3), so the local gate set is a genuine proxy for the hosted macOS lane. `ssh relux` is macOS
**amd64** and is not — label its evidence by architecture, never just "macOS".

`make check-ci-linux PIN_ROOT="$PIN"` is the ubuntu mirror; running it on macOS exercises the recipe
wiring but **is not** Linux evidence — `LINUX_PKGS` excludes godriver on any host, so a green macOS
`check-ci-linux` proves the target works, not that Linux passes. State that distinction in the
evidence.

**`golangci-lint` caveat:** it was **not installed on this machine** at the previous audit cycle
(`golangci-lint --version` → command not found) and this cycle ran no lint command. The producer
must either install it at the pinned version or state plainly that the local lint gate did not run.
**Do not check the lint checklist item off a CI-only green.**

### 9.3 Native supplemental

Same commands, native, after §5.4 transport and identity verification.

- `ssh relux`: §5.4 W10, `$RELUX_GO_EXE` by absolute path under `GOROOT="$RELUX_GO_ROOT"
  GOTOOLCHAIN=local GOENV=off`, after R-P1 confirms `go1.25.5`.
- `ssh win`: §5.4 W1–W9, toolchain passed in as `WIN_GOROOT` / `WIN_GO_EXE`, no `-race` (W7).
  **Blocked** until W-P1 and W-P2 are satisfied — do not install Go from the agent, do not accept an
  ambient PATH, do not discover the toolchain with `where go`.
- `ssh lev`: not attempted. Non-gating; blocked on reachability, an approved toolchain, and
  `TASK-260728-1skseh`.

### 9.4 Gate hygiene (evidence-honesty contract)

Run each gate as a standalone foreground process. Never pipe it through `tee` or an unguarded pipe —
the pipe's exit status is not the gate's. To capture output use
`go test … > .temp/TASK-260720-1pvfj5/gate-NN.log 2>&1; echo "exit=$?"`, or enable `set -o pipefail`
and read `PIPESTATUS`.

**The same rule applies to *transport* pipes, not just gate pipes.** `producer | consumer` under
`set -e` reports the consumer's status, so a producer that dies after emitting a valid partial stream
is silently accepted. That is F3, executably reproduced at 1 of 3 files (§7.4c case R). Stage through
an intermediate file with separately checked statuses, and finish with a complete-set assertion
(§5.2 C3, I13).

On `ssh win`, three separate traps apply, all recorded:
- `cmd … & echo %ERRORLEVEL%` expands at parse time and **always prints 0** — it once invalidated a
  whole Windows evidence set. §5.4 W6's batch file avoids it by putting `set "RC=%ERRORLEVEL%"` on
  its own line, which a batch file parses only after the previous line has run.
- a batch file that ends in a bare `endlocal` does **not** propagate `%RC%`, so
  `ssh win "cmd /c …"` can exit 0 over a failed suite. The file must end
  `endlocal & exit /b %RC%`, as one line (F2, §5.4 W6).
- remote output is CRLF. Every captured value is `tr -d '\r'`-stripped before comparison, or every
  comparison is red on a healthy host.

---

## 10. Fact-check ledger

Every claim traces to one of these reads. Rows 1–52 are from earlier cycles of this audit; rows
53–57 were run in this cycle. All are read-only.

| # | Claim | Command | Result |
|---|---|---|---|
| 1 | `main` = `origin/main` = `17804ce`; `c06aa1a` is divergent | `git rev-parse origin/main`; `git merge-base --is-ancestor` both directions; `git merge-base` | `17804cea…`; **NO** in both directions; merge base `ecb6c1a` |
| 2 | `c06aa1a` is a branch tip, not main | `git log --oneline -1`; `git branch -a --contains 17804ce` | tip of `agent/link-curator-skill-registry`; `17804ce` on `main` and `origin/main` |
| 3 | Committed pin is `00b1688a…` | `git show origin/main:.github/workflows/ci.yml \| grep -n 'ref:'` | `00b1688a…` at lines 28 and 81, **exit 0** — re-run this cycle |
| 4 | Pin/tag positions | `git describe --tags` on four revisions | `00b1688a`→`v1.0.0-rc.2-1-g00b1688`; `57c1f56`→`v1.0.0-rc.3`; `e72defe` and `6c9b1cf`→"No tags can describe" |
| 5 | `e72defe` is the branch-only pin, ancestor of `00b1688a` | `git log -L 28,28:.github/workflows/ci.yml`; `git merge-base --is-ancestor e72defe 00b1688a` | `6c9b1cf`→`e72defe` at commit `0f63a8d`; ancestor **YES** |
| 6 | No rc.4 / rc.5 tag | `git for-each-ref refs/tags` in curator-spec | `v1.0.0-rc.1`, `-rc.2`, `-rc.3` only |
| 7 | Pin capability matrix (§4.4) | `git cat-file -e <rev>:<path>` over 12 paths × 3 revisions; `git ls-tree <rev> conformance/v1/vectors/` | as tabulated; `build-drivers.json` absent at all three |
| 8 | rc.5 root dirt | `git status --short --untracked-files=all -- conformance/` | 3 modified, 354 untracked (357 lines) |
| 9 | rc.5 identity | `shasum -a 256 manifest.json`; `find … \| sort -z \| xargs -0 shasum -a 256 \| shasum -a 256`; `find . -type f \| wc -l` | `b6f56aac…04c`; `e6a13215…2fae`; 448 |
| 10 | Go 1.25.5 | `head -3 go.mod`; `grep 'go-version-file' ci.yml` | `go 1.25.5`; `go-version-file: go.mod` |
| 11 | 40 package dirs; only `interop` is test-only | `find cmd internal -name '*.go' -not -path '*/testdata/*'` deduped to dirs; per-dir non-test `.go` count | 40 dirs; `internal/interop` has 0 non-test `.go` files |
| 12 | Candidate delta = 23 product files, all `_test.go` | `diff -rq` accepted vs candidate, excluding `.git`/`.task-board` | 24 entries: 21 `Only in` (one is `.temp`) + 3 `Files differ`; every product path ends `_test.go` |
| 13 | `.golangci.yml` byte-identical, schema v2 | `diff -q`; `head -3` | IDENTICAL; `version: "2"` |
| 14 | No race gate today | `grep -n race .github/workflows/ci.yml Makefile` | no matches |
| 15 | Disk headroom | `df -h /Users/iv` | 25 GiB free |
| 16 | Linux outside the inventory | `controls.go:75,200,241`; `controls_other.go`; `conformance-claim-v3-qualification.json` | exhaustive over macos+windows; linux `excluded`, `until_task TASK-260728-1skseh` |
| 17 | godriver importers are exactly two | `grep -rl 'internal/godriver' --include='*_test.go' cmd internal`; same without `_test.go` | tests: `cmd/curator` (4 files), `internal/install/stage_test.go`. prod: `cmd/curator` (2), `internal/install` (4) |
| 18 | Six hard-`t.Fatal` sites | direct read of each file at the cited lines | `buildsource:16-19`, `buildcache:15` + `readJSONObject:63-66`, `scopes:38-41`, `marker:37-41`, `whitelist:20-28`, `skillspec:106-109` — all `t.Fatal`, none guarded by an artifact-presence check |
| 19 | Five new delta tests DO guard | `grep -n 'Skipf' <5 files>` | `t.Skipf("%s publishes no build-drivers vector", root)` at `skillcheck:24`, `whitelist:25`, `skillspec:39` and `:301`, `buildcache/builddriver_positive:31` |
| 20 | Main is green because its tree predates this | `ls internal/godriver` at main (absent); `grep 'filepath.Join(root' internal/{closure,skillspec}/conformance_test.go` at main | `internal/godriver` does not exist at main; main reads only `closures.json` and `portable-paths.json`, both present at the pin |
| 21 | Dependency states | `task-board q 'get(<id>) { full }'` × 6 | `1pvfj5` backlog/blocked; `2qqq0w` done; `jrrgw9` development (its own three blockers all done); `1skseh` backlog |
| **22** | **`ssh win` and `ssh lev` are unreachable; `ssh relux` is reachable with Go present** | `ssh -o BatchMode=yes -o ConnectTimeout=10 <host> 'echo ok'`, twice per host; on relux `command -v go`, a probe of five conventional Go paths, `uname -sm`, `command -v tar` | `win`: `connect to host 100.120.84.42 port 22: Operation timed out`, **exit 255**, both attempts. `lev`: same for `100.67.190.45`, **exit 255**, both attempts. `relux`: **exit 0**; `command -v go` → empty; **`FOUND /usr/local/bin/go`**; `/usr/local/go` and `/opt/homebrew/opt/go` absent; `Darwin x86_64`; `/usr/bin/tar` |
| **23** | **rc.5 identity and dirt re-verified this cycle** | `shasum -a 256 manifest.json`; the tree pipeline; `find . -type f \| wc -l`; `git status --short --untracked-files=all -- conformance/ \| awk '{print $1}' \| sort \| uniq -c` | `b6f56aac…04c` **exit 0**; `e6a13215…2fae` **exit 0**; **448**; **354 `??` + 3 `M` = 357 lines** |
| **24** | **F1: the revision-2 archive command is not executable** | in `.temp/TASK-260729-osjeay/tarcheck/`, `tar -cf broken.tar -C "$(dirname "$DST")" conformance` with `DST=<scratch>/cand/conformance/v1` | `tar: conformance: Cannot stat: No such file or directory`, **exit 1** |
| **25** | **F1 fix: the two-level `dirname` form works, with metadata-free flags** | `COPYFILE_DISABLE=1 tar --no-mac-metadata --no-xattrs --no-acls --no-fflags -cf full.tar -C "$(dirname "$(dirname "$DST")")" conformance`; `tar -tf full.tar \| grep -x 'conformance/v1/manifest.json'` | archive **exit 0**; listing assertion **exit 0**; entries `conformance/`, `conformance/v1/`, `conformance/v1/manifest.json`. Host tar: `bsdtar 3.5.3 - libarchive 3.7.4` |
| **26** | **The candidate tree is safe for a PowerShell reimplementation** | `find . -type l \| wc -l`; `find . ! -type f ! -type d ! -type l \| wc -l`; `find . -type d \| wc -l`; `find . -type d -empty`; `find . -type f -print \| LC_ALL=C grep -n '[^ -~]'`; `find . -type f -print \| LC_ALL=C grep -c '[[:space:]]'` | **0** symlinks, **0** other non-regular, 68 dirs, **0** empty dirs, **no** non-ASCII match, **0** paths with whitespace |
| **27** | **The inventory file's own SHA-256 IS the tree digest** | `find . -type f -print0 \| LC_ALL=C sort -z \| xargs -0 shasum -a 256 > $INV`; `wc -l < $INV`; `shasum -a 256 $INV`; `cd <root> && shasum -a 256 -c --status $INV` | write **exit 0**; **448** lines; **`e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae`**; per-file check **exit 0**. First line `f116910f…  ./expected/adapter-ledger.json`, last `5a11e968…  ./vectors/source-identities.json` |
| **28** | **The godriver rejection test named in §6.2 exists** | `grep -rn 'func TestProbeRejects…' internal/godriver/*_test.go` in the candidate worktree | `internal/godriver/worker_test.go:434: func TestProbeRejectsAnUncoveredPlatformBeforeTheWorker(t *testing.T)`, **exit 0** |
| **29** | **The module has external deps, no vendor dir, and a submodule `replace`** | `cat go.mod`; `wc -l go.sum`; `ls -d vendor` | 4 direct + 17 indirect requires; `go.sum` 45 lines; **no vendor dir**; `replace github.com/relux-works/skill-go-testing-tools/tuitestkit => ./agents/skills/skill-go-testing-tools/tuitestkit` |
| **30** | **The replace target is a populated git submodule** | `cat .gitmodules`; `git submodule status`; `ls agents/skills/skill-go-testing-tools/tuitestkit` in repo and candidate worktree | submodule at `agents/skills/skill-go-testing-tools`, pinned `21585d0e937cae47e54a788d8ae36b1780eae47f` (`v1.0.1-4-g21585d0`); `go.mod` + sources present in **both** trees |
| **31** | **Exactly two symlinks in the source tree, both outside the Go build** | `find . -type l -not -path './.git/*' -not -path './.temp/*' -not -path './.task-board/*'` in the candidate worktree | `./.claude/skills/skill-go-testing-tools`, `./.codex/skills/skill-go-testing-tools`; **count = 2** |
| **32** | **D6's inherited premise is stale; the race gate is the one the timeout rescues** | `LOGBOOK.md` entries 1637, 1659, 1653, 0607 read in full | 1637: `CURATOR_CONFORMANCE_ROOT=<immutable rc.5 root> go test -count=1 ./...` **exit 0 in 444 s**, `cmd/curator` 384.270 s; `go test -count=1 -race ./...` **exit 1 in 610 s**, `internal/install` FAIL **603.306 s**, `internal/install/atomicity` FAIL **603.701 s**, no `DATA RACE`. 1659: cumulative duration in the `saveJournal` → `namespacePathsOverlap` leaf; race factors **2.69×** / **2.75×**; projections **918 s** / **1121 s**; `jrrgw9` routed back to `development`, production fix split to `TASK-260729-3dr6hw`. 0607: the pre-rework `2kaopg` measurement, `cmd/curator` 602.193 s. **Rows 1659's factors and projections are SUPERSEDED by LOGBOOK 1732 — see row 35.** |
| **33** | **Current CI and Makefile inventories (§2.1, §2.2)** | `git show origin/main:.github/workflows/ci.yml` read in full; `sed -n '1,60p' Makefile` | CI: `test` matrix `[ubuntu-latest, macos-latest, windows-latest]`; gofmt `if: runner.os != 'Windows'`; `go vet ./...`; `go test ./...` with the root; `lint` `version: latest`; `interop` `go test ./internal/interop/ -v`; `naming-gate` inline grep; all four `submodules: true`; **no `-timeout`, no `-race`**. Makefile: `build`, `test` (`go test ./...`), `fmt`, `vet`, `lint`, `check` (`vet test` + `gofmt -l .`); **no race, no timeout, no root plumbing** |

Rows 34–40 were run in **this** cycle (rework cycle 3). All read-only; **still no Go command**.

| # | Claim | Command | Result |
|---|---|---|---|
| **34** | **F3: the race lane has a measured PASS, not only timeouts** | `cat .temp/TASK-260720-2284br/gates-rework1/gate-race.log`; `cat …/gate-race-exit.txt` | 6 `ok` lines: `internal/install` **609.117 s**, `internal/install/atomicity` **1422.407 s**, `transaction` 35.800 s, `managerlock` 13.875 s, `staging` 1.737 s, `adapters` 3.193 s; exit file reads **`race exit=0`**. Revision 3's "the only executed numbers are timeouts" is false |
| **35** | **The 2.75× / 1121 s atomicity model is superseded** | `grep -n '1732' LOGBOOK.md`; entry 1732 read in full | Factors are **not** shared: install **×2.67** (609.117 / 228.344), atomicity **×4.02** (1422.407 / 353.629). Corrected projections **890–1000 s** and **1284–1494 s**. Atomicity is **absent from `candidate-source-delta-post.txt`** → pre-existing debt, not a `jrrgw9` regression |
| **36** | **F1: bare `go` is ambiguous on this host — three launchers, two versions** | `which -a go`; then `head -1 <root>/VERSION` for each root (**file reads, not `go version`**) | `/opt/homebrew/bin/go` → `…/Cellar/go/1.25.5/libexec` = **`go1.25.5`**; `/usr/local/go/bin/go` → `/usr/local/go` = **`go1.25.1`**; `/Users/iv/.goenv/shims/go` → `/Users/iv/.goenv/versions/1.25.5` = **`go1.25.5`**. `go.mod:3` = `go 1.25.5`, **no `toolchain` line** → under default `GOTOOLCHAIN=auto` launcher 2 would **download** a toolchain |
| **37** | **The same bare `go` resolved differently in two agent shells** | `command -v go` this cycle, compared with the cycle-3 verdict's recorded value | This cycle: **`/opt/homebrew/bin/go`**. Cycle-3 reviewer: **`/Users/iv/.goenv/shims/go`**. Same host, same repo. The shim is a `Bourne-Again shell script`, per `file` |
| **38** | **`GOROOT` is exported and does not match the resolved launcher; no Go env file exists** | `echo "$GOROOT"`, `$GOTOOLCHAIN`, `$GOENV`, `$GOFLAGS`; `ls` of `~/Library/Application Support/go/env` and `~/.config/go/env`; `cat /Users/iv/.goenv/version`; `ls /Users/iv/.goenv/versions/` | `GOROOT=/Users/iv/.goenv/versions/1.25.5` (goenv tree) while bare `go` resolved to the **Homebrew** tree; `GOTOOLCHAIN`, `GOENV`, `GOFLAGS` all **unset**; **neither Go env config file exists**; goenv global version `1.25.5`, with **both 1.25.1 and 1.25.5** installed; repo has **no `.go-version`** |
| **39** | **F2: the revision-3 discovery/guard forms mask failure; the corrected forms fail closed** | `make` against a `/bin/sh` stub `go` (`.temp/TASK-260729-osjeay/f2-repro/`, `.temp/TASK-260729-osjeay/f2-recipe/`). **No Go executed** | Revision-3 `$(shell …\|grep)` → **exit 0** with a silently truncated 2-package list; revision-3 `go list \| grep -q` guard → **exit 0** on a `go list` that exited 1. Corrected recipes, copied verbatim from §7: healthy **0** (3 of 4 pkgs, godriver excluded), partial-`go list` failure **2**, relative `GO` **2**, `go1.25.1` toolchain **2**, hosted-runner escape hatch **0**. Full table in §7.4 |
| **40** | **F4: board checklists are append-only — "replace item 13" is not executable** | `task-board m 'schema(checklist)'`; `task-board q 'get(TASK-260729-osjeay){checklist}'` | Mutation set is `add_checklist_item`, `check_item`, `uncheck_item` — **no `remove_checklist_item`, no edit mutation**. Item 13 `Tests green` = `done:false`; item 15 (reviewer's own) = `done:false`. See D7 |

Rows 41–52 were run in **this** cycle (rework cycle 4). All read-only; **still no Go command**.
Rows 44–49 are the six network reads disclosed in the header — read-only documentation and release
metadata, no dependency pull, install or download.

| # | Claim | Command | Result |
|---|---|---|---|
| **41** | **F1/F3: the revision-4 `require-toolchain` accepts three defect shapes; the corrected recipe and staging contract reject eight; the harness is re-runnable** | `sh .temp/TASK-260729-osjeay/verify-recipes.sh` — 21 cases, `make` + `/bin/sh` against stub `go`. **No Go executed** | **ALL 21 EXPECTATIONS MET, real process exit 0**, twice. rev4 accepts cross-root `GOFMT` (C), unapproved `GOROOT` (D) and `GOTOOLCHAIN=auto` (E), all **exit 0**. rev4 `tar\|tar` staging **exit 0** with **1 of 3 files** (R). Corrected: healthy native **0** (F), healthy hosted **0** (G), and **exit 2** for partial `go list` (H), missing `GOROOT_EXPECTED` (I), relative `GO` (J), `go1.25.1` (K), root drift (L), cross-root formatter (M), `GOTOOLCHAIN=auto` (N), user `GOENV` (O), `go.mod` drift (P), hosted+cross-root formatter (Q); corrected staging **exit 1** on failed producer (S), missing file (T), extra file (U). Log: `TASK-260729-osjeay_verify-recipes-cycle4.log` |
| **42** | **rc.5 identity and dirt re-verified again this cycle — unchanged** | `shasum -a 256 manifest.json`; `find . -type f -print0 \| LC_ALL=C sort -z \| xargs -0 shasum -a 256` → inventory → `shasum -a 256`; `wc -l`; `git status --short --untracked-files=all -- conformance/ \| awk '{print $1}' \| sort \| uniq -c` | `b6f56aac…04c` **exit 0**; `e6a13215…2fae`; **448**; **354 `??` + 3 `M`**. Identical to rows 9 and 23 |
| **43** | **Pin and dependency state unchanged** | `git show origin/main:.github/workflows/ci.yml \| grep -n 'ref:'`; `task-board q 'get(…){id status}'` × 4 | `00b1688a…` at lines **28 and 81**, exit 0; `1pvfj5` **backlog**, `jrrgw9` **development**, `2qqq0w` **done**, `1skseh` **backlog** |
| **44** | **F4: `actions/checkout` current major is v7, not v6** | `gh api repos/actions/checkout/releases`; `gh api repos/actions/checkout/git/refs/tags` | `v7.0.1` **2026-07-20**, `v6.1.0` 2026-07-20, `v5.1.0` 2026-07-20; floating majors `v1`…**`v7`** (`v7` → `3d3c42e5…`). Repo pins `@v4`. **The cycle-4 verdict's "checkout v6" is one major stale** |
| **45** | **F4: `actions/setup-go` current major is v7, not v6 — and v6 forces `GOTOOLCHAIN=local`** | `gh api repos/actions/setup-go/releases`; `…/releases/tags/v6.0.0`; `…/releases/tags/v7.0.0`; `gh api repos/actions/setup-go/pulls/460`; `…/git/refs/tags` | `v7.0.0` **2026-07-16**, `v6.5.0` 2026-06-24; floating majors up to **`v7`** (`b7ad1dad…`). v6.0.0 **Breaking Changes**: PR #460 *"Improve toolchain handling"* + node20→node24 ("runner v2.327.1 or later"). PR #460 body: *"force `go` to always use the local toolchain … via setting the `GOTOOLCHAIN` environment variable to `local`"*, explicitly *"a **breaking change** for some workflows"*. Repo pins `@v5` → hosted jobs run `GOTOOLCHAIN=auto` today |
| **46** | **F4: `golangci-lint-action` current major is v9 (verdict correct here); v8 raised the linter floor** | `gh api repos/golangci/golangci-lint-action/releases`; `…/releases/tags/v8.0.0`; `…/tags/v9.0.0`; `…/git/refs/tags` | `v9.3.0` **2026-06-29**; majors up to **`v9`** (`db9de0fc…`). v8.0.0: *"Requires `golangci-lint` version >= `v2.1.0`"*. v9.0.0: node20→node24 runtime, `install-only`, module plugin system. Repo pins `@v7` |
| **47** | **F4: revision 4's proposed `v2.4.0` linter pin is an 8-minor downgrade** | `gh api repos/golangci/golangci-lint/releases`; `…/releases/tags/v2.4.0` | current `v2.12.2` **2026-05-06**; `v2.4.0` **2025-08-14**. `.golangci.yml` line 1 is `version: "2"` (local read) |
| **48** | **F4: current `*-latest` runner label mapping and the migration policy** | `curl -sS https://raw.githubusercontent.com/actions/runner-images/main/README.md` (http 200) then `grep -n` for the three labels and `sed -n '90,110p'` | `ubuntu-latest` → **Ubuntu 24.04 x64**; `macos-latest` → **macOS 26 arm64**; `windows-latest` → **Windows Server 2025 x64**. Policy: *"The `-latest` migration process is gradual and happens over 1-2 months … any workflow using the `-latest` label may see changes in the OS version"* |
| **49** | **F4: named upcoming runner-image changes** | `gh api 'repos/actions/runner-images/issues?labels=Announcement&state=all&per_page=8'` | open: *"[Ubuntu] Ubuntu 26.04 and Ubuntu 26.04 Arm is now available as a public preview"* (2026-06-11); *"[Ubuntu] The Ubuntu 22 based runner images will begin deprecation on September 17th…"* (2026-06-16); *"[macOS] Default Xcode on macOS 26 Tahoe will be set to Xcode 26.6 on 2026.07.21"* (2026-07-07) |
| **50** | **The repository's action pins and runner labels, from source** | `git show origin/main:.github/workflows/ci.yml \| grep -n -E 'uses:\|runs-on:\|os:\|version:\|ref:'` | `actions/checkout@v4` × 6 (lines 20, 25, 57, 73, 78, 97); `actions/setup-go@v5` × 3 (31, 61, 84), each `go-version-file: go.mod`; `golangci/golangci-lint-action@v7` (65) with `version: latest` (67); `os: [ubuntu-latest, macos-latest, windows-latest]` (18); three `runs-on: ubuntu-latest` (55, 71, 95) |
| **51** | **`release.yml` is a separate action surface, out of 1pvfj5 scope** | `git show origin/main:.github/workflows/release.yml \| grep -n -E 'uses:\|runs-on:'` | `macos-latest`; `checkout@v4`, `setup-go@v5`, `cosign-installer@v3`, `sbom-action/download-syft@v0`, `goreleaser-action@v6`, `attest-build-provenance@v2`. **Inventoried, not touched** |
| **52** | **`go.mod` still requires exactly 1.25.5, which is what `GO_VERSION_REQUIRED` reconciles against** | `awk '/^go /{print}' go.mod` | `go 1.25.5` |
| **53** | **Cycle 5: the extended recipe/step/staging harness is green, 41 cases** | `sh .temp/TASK-260729-osjeay/verify-recipes.sh` — standalone, no pipe, run **twice** | `ALL 41 EXPECTATIONS MET`, real exit **0** both runs. Log attached as `TASK-260729-osjeay_verify-recipes-cycle5.log` |
| **54** | **The §6.0a hosted identity step passes healthy and Windows-shape runners and fails closed on six defect shapes** | §7.4e cases AB–AJ inside row 53 | AB **0**, AG **0**; AC/AD/AE/AF/AH/AI/AJ each **1** |
| **55** | **The §5.4 W2/W3/W9 block stops before transport on a stale base, and W9 rejects an unconfirmed cleanup** | §7.4g cases AK–AO inside row 53, `ssh`/`scp` stubbed — **no Windows host contacted** | AK **0**; AL/AM wrappers **0** (guard failed, no transport marker); AN **0**; AO **1** |
| **56** | **No PowerShell on this audit host, so §6.0a's `shell: pwsh` alternate is specified, not executed** | `command -v pwsh`; `command -v powershell` | both **exit 1**, no output |
| **57** | **This cycle made no network read of any kind** | no `gh`, `curl`, `git fetch`, `go`, install or download command was issued | the only new measurement is row 53 |

**Not verified, by design:** **no Go command of any kind was executed** in this cycle — no `go`,
`go test`, `go vet`, `go build`, `go list`, `go version`, `gofmt`, `golangci-lint`. Toolchain
versions in row 36 were read from `VERSION` **files**, never by running a compiler. The §7.4 harness
runs `make` against `/bin/sh` stubs, so its exit codes describe the harness, **not** the Curator
suite. No install, no download, no dependency fetch, no host mutation, no board mutation outside this
task's own status, notes, checklist and resources. **Six read-only network reads were made** (rows
44–49) and are disclosed in the header, correcting revisions 1–4's blanket "no network fetch to
GitHub"; they read release metadata and documentation only. `/usr/local/bin/go` on relux was proved
to **exist and be executable**; its version is unmeasured. The PowerShell script in §5.4 W5 has
**never been run**, and **no part of the `cmd.exe` lane has been executed** — `ssh win` is
unreachable, so W1/W4/W6/W8 and W-N1…W-N6 are producer gates whose control flow rests on the cited
Microsoft `exit /b` documentation and on prior recorded traps, not on a measurement taken here. The
exact output strings of `go env GOTOOLCHAIN` / `go env GOENV` under an explicit export are asserted
from Go's documented behaviour and owed one producer confirmation (§5.0). `go test -race ./...
-timeout 30m` — the AC's gate — has **never been run**. Every red-prediction in §3, §4.3 and §8.2 is
static analysis or a cited prior measurement, each carrying a named confirmation command in §9.1.
**No claim in this document is a green CI result.**

---

## 11. Invariants the producer must preserve

- **I1.** The candidate enters only through a non-default input, defaulting to empty. With it empty,
  CI uses the committed pin and nothing else.
- **I2 (strengthened).** The candidate's accepted identity is three fixed constants —
  `CANDIDATE_MANIFEST_SHA256=b6f56aac…04c`, `CANDIDATE_TREE_SHA256=e6a13215…2fae`,
  `CANDIDATE_FILE_COUNT=448`. Every candidate run verifies **all three** before executing a case and
  **emits all three** into the log. Any mismatch aborts the run. A mismatch at freeze time means the
  producer is holding a **different candidate**: record the measured values, stop, and escalate for
  a refreshed accepted input and a board-owner decision — **never re-baseline silently**. The
  archive SHA-256 is a transport-integrity check handed off through the board resource in §5.2 C4;
  it is verified before extraction and never substitutes for post-extraction tree verification. The
  digest is never written as a `ref:`.
- **I3.** No job, target, or log line labels candidate output a release, a published conformance
  claim, or "rc.5 qualified". They say *candidate*.
- **I4.** 1pvfj5 does not move the committed pin. `TASK-260720-38l1sy` owns promotion after
  `TASK-260720-25d05o`.
- **I5.** The pin appears in exactly one place (`env.SPEC_PIN`). Every suite checkout references it.
- **I6.** The composite is built on `origin/main` = `17804ce`, **not** on `c06aa1a`. Building on
  `c06aa1a` would revert the pin from `00b1688a` to `e72defe` and break `internal/closure`, which
  needs `manager-lifecycle.json`.
- **I7 (premise corrected twice).** Every `go test` invocation carries `-count=1 -timeout 30m`. What
  that flag rescues is the **race** lane, which failed at the default alarm (`internal/install`
  603.306 s, `internal/install/atomicity` 603.701 s) and **passed** focused at 45 m (609.117 s and
  **1422.407 s**, gate exit 0). On the *non*-race lane the flag is margin, not a fix — the current
  candidate passes the default-timeout run at 444 s. Do not claim "CI is red without the timeout" as
  a present-tense fact, and do not claim the race lane is only projected either (D6): it is measured
  focused, and **unmeasured for `./...`**, where atomicity's 1.27× headroom is the exposure. Never
  raise the alarm past 30 m to make a gate pass.
- **I8.** No `-run` selector is added without confirming it matches a test that exists in the
  composite. A Go `-run` matching nothing exits 0 — a vacuous gate reads exactly like a passing one.
- **I9 (new).** `vendor/` is a transport artifact only. `go mod vendor` runs in the staging copy
  (§5.2 C3), never in the repository worktree, and `vendor/` is never committed. The submodule pin
  `21585d0e937cae47e54a788d8ae36b1780eae47f` is not 1pvfj5's to move.
- **I10 (new).** Native-host prerequisites are named, not worked around. `ssh win` needs
  reachability (W-P1) and an operator-installed, manager-approved **`go1.25.5`** root (W-P2), not a
  looser "1.25.x" one. `ssh relux`
  needs R-P1. `ssh lev` needs reachability, an approved absolute `go1.25.5` root (L-P1, §9.1 step 6),
  and `TASK-260728-1skseh`. The
  agent installs nothing, downloads nothing, and does not accept an ambient PATH. A blocked lane is
  reported as blocked, never inferred green and never silently dropped from the evidence.
- **I11 (strengthened).** On every **native** host the toolchain is **one** operator-approved
  absolute `GOROOT_EXPECTED`; `GO` and `GOFMT` are **derived** from it, never supplied separately, so
  a cross-root pairing cannot be expressed. Before any gate, `require-toolchain` asserts, by
  comparison and not by printing: `go.mod` still requires `go1.25.5`; `go version` is exactly that;
  `go env GOROOT` equals `GOROOT_EXPECTED` byte for byte; the launcher **is** `$GOROOT/bin/go` and
  the formatter **is** `$GOROOT/bin/gofmt`; `go env GOTOOLCHAIN` reads back `local`; `go env GOENV`
  reads back `off`. The environment is **supplied** by `$(GOENVPREFIX)` around every Go invocation
  **and read back**, because a shim or wrapper launcher can override its caller. No `where go`, no
  bare `go`/`gofmt`, no inherited `GOROOT`; `make` carries `GOROOT_EXPECTED=`. Exits 2 otherwise
  (§7.4d rows I–P). The one exception is a **hosted** GitHub runner, where `actions/setup-go` owns
  `PATH` in a fresh image: those steps set `TOOLCHAIN_ALLOW_PATH=1`, which relaxes **path shape
  only** for the Make-mediated gates — and, because the hosted jobs also issue **direct** Go commands
  that reach no Make target, every Go-consuming job additionally runs the exact `Verify Go toolchain
  identity` step of **§6.0a**, which re-asserts all seven properties in the job itself. `naming-gate`
  consumes no Go and runs no such step. Executed against stubs: the Make exception still rejects a
  cross-root formatter (§7.4d row Q), and the §6.0a step passes a healthy runner and a Windows-shape
  root while rejecting wrong version, forced `auto`, user `GOENV`, cross-root formatter on both shell
  shapes, a shim launcher and `go.mod` drift (§7.4e rows AB–AJ). Evidence lines use `printf '%s\n'`,
  never `echo`: `echo` expands `\r` inside a Windows `GOROOT` and corrupts the record (§1.3).
  Hosted `GOTOOLCHAIN=local` comes from the workflow `env:` block, not from `setup-go@v5`, which does
  not force it (§2.3 D8). Evidence records the toolchain that produced it — on this host bare `go`
  resolved to two different toolchains in two agent shells in one week, and a third launcher on
  `PATH` is older than `go.mod` requires.
- **I12 (new).** Package discovery never flows through `$(shell …)` or an unguarded pipe.
  `go list` output is materialized in a status-checked assignment and only then filtered; the safe
  set is asserted non-empty and the exclusion asserted to be exactly one package (§7, §7.4). A
  partial listing must abort, never become a smaller green test lane.
- **I13 (new).** **No producer→consumer pipe anywhere in the transport path.** Every archive is
  created, listed and extracted as separate processes whose statuses are each checked, because
  `set -e` observes only a pipeline's last command and a producer that emits a valid partial stream
  is otherwise accepted silently (executably reproduced at 1 of 3 files, §7.4c). **The same rule binds
  the origin enumeration that judges completeness**: `find … -print0 | sort -z | xargs -0 shasum`
  reports only `xargs`'s status, so a `find` that dies after a valid partial stream yields a short but
  internally consistent inventory that nothing downstream can contradict (reproduced at 2 of 3 files,
  §7.4f case V). Enumeration is therefore three materialized, separately status-checked stages —
  paths, sorted paths, digests — with a `> 0` count assertion, in **both** §5.2 C2 and §5.2 C3.
  `pipefail` is never relied on: POSIX `/bin/sh` does not provide it. Every staged tree
  ends in a **complete-set assertion**: a per-file digest inventory enumerated *at the origin*
  (catches changed and missing) **plus** a destination file count (catches extra, which `-c`
  structurally cannot see). `tar -tf … | grep -q` is likewise replaced by a materialized listing file
  (§5.2 C3).
- **I14 (new).** **Every remote assertion compares; none merely prints.** On `ssh win` that means:
  each captured value is `tr -d '\r'`-stripped and compared with `[ … ] || exit 1`; both archive
  digests are verified **before** their own extraction, the source digest included, because the
  source tree has no post-extraction equivalent of W5; the batch runner ends
  `endlocal & exit /b %RC%` **and** persists `EXITCODE=` to a file; and W8 asserts the SSH status,
  the persisted code and the printed code all agree, discarding the run if they do not. **The base
  directory is an executable empty-root precondition, not a `mkdir`**: proved absent, created with its
  status checked, proved to hold **zero** entries (`Get-ChildItem -Force`), all before W3 transports
  anything — because the source tree has no post-extraction equivalent of W5, so one stale file would
  otherwise enter the suite behind a correct archive digest. W9 cleanup is status-checked **and**
  absence-confirmed with the same probe W2 uses, so the two can never disagree. The six
  negative injections W-N1…W-N6 must each be observed failing before any Windows exit 0 is trusted —
  the lane cannot be executed from this audit, so a never-seen-to-fail assertion is not yet a gate.
- **I15 (new).** Action majors and runner labels are a **dated, source-verified ledger with an
  explicit disposition**, not an inheritance (§2.3). 1pvfj5 moves **no** action major: `checkout@v4`,
  `setup-go@v5` and `golangci-lint-action@v7` are retained deliberately, the mutable
  `version: latest` is frozen to the version `latest` resolves to today, and the one behavioural gap
  a retained major leaves — `setup-go@v5` not forcing `GOTOOLCHAIN=local` — is closed in the
  workflow `env:` block rather than by a version bump. The moving `*-latest` labels are kept on
  purpose; in exchange, every CI evidence line records the **concrete image version and
  architecture** it ran on, so a later red job can be attributed to an image migration rather than to
  a code change. `macos-latest` is **arm64**; never write "macOS" without the architecture.
