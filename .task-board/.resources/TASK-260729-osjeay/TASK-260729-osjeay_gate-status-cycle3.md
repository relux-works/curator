# Gate status — TASK-260729-osjeay rework cycle 3

Supersedes the cycle-2 gate status. Same evidence-honesty posture; the list below is what **this**
cycle actually ran.

## Commands actually executed this cycle

Each ran as a standalone foreground process. None was piped through `tee`; no gate's status was
taken from a pipeline.

| # | Command | Exit | Result |
|---|---|---|---|
| 1 | `cat .temp/TASK-260720-2284br/gates-rework1/gate-race.log`; `cat …/gate-race-exit.txt` | 0 | 6 `ok` lines — `internal/install` **609.117 s**, `internal/install/atomicity` **1422.407 s**, `transaction` 35.800 s, `managerlock` 13.875 s, `staging` 1.737 s, `adapters` 3.193 s; exit file reads **`race exit=0`**. **This is F3: the race lane has a measured pass, not only timeouts.** |
| 2 | `grep -n '1732' LOGBOOK.md`; entry 1732 read in full | 0 | race factors are **not** shared — install **×2.67**, atomicity **×4.02**; corrected projections **890–1000 s** / **1284–1494 s**; atomicity is absent from `candidate-source-delta-post.txt`, so its cost is pre-existing debt |
| 3 | `command -v go`; `which -a go` | 0 | **`/opt/homebrew/bin/go`**; three launchers total (`/opt/homebrew/bin/go`, `/usr/local/go/bin/go`, `/Users/iv/.goenv/shims/go`) |
| 4 | `head -1 <root>/VERSION` for each candidate GOROOT — **file reads, not `go version`** | 0 | Homebrew Cellar `go1.25.5`; **`/usr/local/go` `go1.25.1`**; goenv 1.25.5 `go1.25.5`; goenv 1.25.1 `go1.25.1` |
| 5 | `grep -nE '^(go\|toolchain) ' go.mod` | 0 | `go 1.25.5`, **no `toolchain` line** → a `go1.25.1` launcher would **download** a toolchain under the default `GOTOOLCHAIN=auto` |
| 6 | `echo "$GOROOT"`, `$GOTOOLCHAIN`, `$GOENV`, `$GOFLAGS`; `ls` both Go env-file paths; `cat /Users/iv/.goenv/version`; `ls /Users/iv/.goenv/versions/`; `cat .go-version` | 0 / 1 for the absent files | `GOROOT=/Users/iv/.goenv/versions/1.25.5` **exported while bare `go` resolved to the Homebrew tree**; `GOTOOLCHAIN`/`GOENV`/`GOFLAGS` unset; **neither Go env config file exists**; goenv global `1.25.5` with both 1.25.1 and 1.25.5 installed; **no `.go-version`** in the repo (the `cat` non-zero is the expected "no such file") |
| 7 | `file /Users/iv/.goenv/shims/go`; `head -5` | 0 | `Bourne-Again shell script text executable` — the shim is a script, not a launcher |
| 8 | `git show origin/main:.github/workflows/ci.yml`; `git show origin/main:Makefile` | 0 | re-read in full; `setup-go@v5` + `go-version-file: go.mod` on all three job groups; `Makefile` still `GO ?= go`, no race, no timeout, no root plumbing |
| 9 | `make` against `/bin/sh` stubs — revision-3 discovery form | **0** | **expected-red as a defect demonstration**: `BAD_PKGS` silently became a **2-package** list from a `go list` that exited 1, and the recipe body ran. This is F2 reproduced. |
| 10 | `make` against the same stub — revision-3 `grep -q` guard | **0** | **expected-red as a defect demonstration**: printed `bad-guard: reported ok` against a failing `go list` |
| 11 | `sh .temp/TASK-260729-osjeay/verify-recipes.sh` (7 cases, corrected recipes copied verbatim from §7) | **0** | `ALL 7 EXPECTATIONS MET`. Healthy **0** (3 of 4 packages, godriver excluded); partial `go list` **2**; relative `GO` **2**; `go1.25.1` toolchain **2**; hosted-runner exception **0** |
| 12 | `task-board m 'schema(checklist)'`; `task-board q 'get(TASK-260729-osjeay){checklist}'` | 0 | mutation set is `add_checklist_item` / `check_item` / `uncheck_item` — **no `remove_checklist_item`, no edit mutation**. Item 13 `Tests green` = `done:false` |

Rows 9–11 ran under `.temp/TASK-260729-osjeay/`. No product path was written.

## An honest note about row 11

The harness was **red on its first execution** — case I reported exit 2 instead of 0. The recipe was
correct; the harness was not. A prefix assignment before a shell **function** call
(`STUB_VER=go1.25.1 run …`) persists in the caller's shell under POSIX `sh`, so case H's downgraded
version leaked into case I. Fixed by assigning the stub variables explicitly per case, then re-run
green. Recorded rather than quietly repaired, because "a test harness that lies about the thing under
test" is the same defect class as F2.

## Commands NOT executed, and why

**No Go command of any kind ran this cycle** — no `go`, `go test`, `go vet`, `go build`, `go list`,
`go version`, `gofmt`, `golangci-lint`. The toolchain versions in row 4 come from reading `VERSION`
**files**. The stub in rows 9–11 is a 28-line `/bin/sh` script; `make` drove it, never a compiler.
No install, download, GitHub fetch, or host mutation. The scope brief forbids it: *"do not
pull/install/download, and do not run Go or heavy tests while verifier3 is active."*

Consequences worth stating plainly:

- The §7.4 exit codes describe the **stub harness**. They prove the recipes' control flow and
  fail-closed behaviour. They say nothing about whether Curator's tests pass.
- **`go test -race ./... -timeout 30m` — the gate the AC names — has still never been run.** The
  closest measurement is a focused 6-package pass at `-timeout 45m` with atomicity at 1422.407 s,
  i.e. **1.27× headroom** against the 1800 s alarm, taken with five packages in flight rather than
  forty.
- No native host was contacted this cycle; the cycle-2 reachability results stand (`ssh win` and
  `ssh lev` exit 255, `ssh relux` reachable, Go version still unmeasured).
- The PowerShell script in §5.4 W5 has still **never been run**.

Every command in §5–§9 of the execution map remains a **future producer gate**. **No green CI result
is claimed anywhere in this task output.**

## Reviewer findings F1–F4 (cycle 3) — disposition

| Finding | Disposition | Evidence |
|---|---|---|
| F1 native gates accept ambient Go | **fixed, and the hazard measured** | Rows 3–7. New §5.0 fixes one absolute executable + matching `GOROOT` per host with `GOTOOLCHAIN=local` / `GOENV=off`; `require-toolchain` enforces it (rows 11 G/H); §5.2 C3, §5.4 W1/W6/W10, §7, §9.2 all use absolute paths; new invariant I11. **Bare `go` resolved to two different toolchains in two agent shells on this host, and a third launcher on `PATH` is `go1.25.1` against `go.mod`'s `1.25.5`.** |
| F2 `go list` failure masked | **fixed, and both the defect and the fix executed** | Rows 9–11. Discovery moved into the recipe behind `rows="$$(…)" \|\| exit 1`; exclusion asserted to be exactly one package; safe set asserted non-empty; both guard probes materialize complete output first. New invariant I12. |
| F3 race evidence stale | **fixed** | Rows 1–2. `gate-race.log` and LOGBOOK 1732 incorporated into D6, §6.3, §8 rows 4–5, §9.1 step 4, I7 and §10 rows 34–35. The 918 s / 1121 s / 2.75× model is withdrawn; the risk is restated as atomicity's **1.27×** headroom on an unproven full-suite gate. |
| F4 stored review contract incomplete | **escalated with a new blocking constraint** | Row 12. Item 13 stays unchecked. **The verdict's "replace it" branch is not executable: the board has no `remove_checklist_item` mutation** — checklists are append-only. §3 D7 gives three options with exact wording; recommendation is D7-c now (owner records not-applicable and accepts with it unchecked), D7-a next (role-level qualification, since this is the second occurrence after `TASK-260729-2sxx7k` / LOGBOOK 0510). A new executable task-scoped validation item covering §7.4 was appended **and checked against real exit codes**. **The board owner had already recorded the D7-c reconciliation in task notes on 2026-07-29 and reached the same append-only conclusion independently** — item 15 is the reviewer gate that owner added — so F4's contract is resolved for this task and only the class fix (D7-a) stays open. |

## Preserved from prior cycles, unchanged

- Candidate manifest `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`,
  tree `e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae`, **448** files.
- Conformance-root status: **3 modified, 354 untracked**.
- Committed suite pin `00b1688a9b2457ca397a0bb550acf47cad8ee967`, unmoved.
- Linux native validation **non-gating**, blocked on its named external prerequisite plus
  `TASK-260728-1skseh`.
- Every command stated as a future producer gate, not green evidence.

## Artifact identity

`TASK-260729-osjeay_final-ci-execution-map.md` **revision 4**, SHA-256
`73a2da0ea21b71a0da5723e13436767850c8aff58d419257fba06776b113c36e` (revision 3 was
`d93e155edf2a4ddf2b23b353f5c411bb40223a9c4be8295973a03bc2990c7d93`).

## Mutation scope

No product, spec, CI, `Makefile`, pin, or `TASK-260720-1pvfj5` field was modified. Board writes were
limited to this task: its own status, notes, checklist and outcome resources.
