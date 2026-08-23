# TASK-260729-osjeay — gate status, rework cycle 5 (revision 6)

**Date:** 2026-07-29 · **Scope:** read-only execution-map revision. No product, spec, CI,
`Makefile`, pin, or `TASK-260720-1pvfj5` field was modified (`git status --short` over
`.github/ Makefile go.mod internal/ cmd/ conformance/ .scripts/` → empty; `1pvfj5` still `backlog`).

## Evidence-honesty ledger for this cycle

| Command | Standalone? | Real exit | Status |
|---|---|---|---|
| `sh .temp/TASK-260729-osjeay/verify-recipes.sh` | yes, no pipe, no `tee` | **0** | **green** — `ALL 41 EXPECTATIONS MET`, reproduced twice |
| `command -v pwsh` / `command -v powershell` | yes | **1** / **1** | **expected-red** — no PowerShell on this host, so §6.0a's `shell: pwsh` alternate is specified, not executed |
| `git status --short -- <product paths>` | yes | **0** | green — empty output |
| `task-board q 'get(TASK-260720-1pvfj5){id status}'` | yes | **0** | green — `backlog`, untouched |

**No Go command of any kind was executed.** No `go`, `go test`, `go vet`, `go build`, `go list`,
`go version`, `gofmt`, `golangci-lint`. **No Windows host was contacted** — `ssh`/`scp` are `/bin/sh`
stubs. **No network read of any kind this cycle** (cycle 4's six reads stay disclosed in the header
and as §10 rows 44–49).

## Reviewer findings — disposition

| # | Finding | Status | Executable evidence |
|---|---|---|---|
| F1 | hosted Go jobs did not execute the toolchain identity contract | **fixed** | New §6.0a `Verify Go toolchain identity` step, added to `test`, `test-linux`, `race`, `lint`, `interop` (not `naming-gate`). §6.0 gains `GOENV: off`. Harness cases AB **0**, AG **0**, AC/AD/AE/AF/AH/AI/AJ each **1**. `shell: pwsh` alternate given but **not executed** (no PowerShell here) |
| F2 | source-origin enumeration masked a producer failure | **fixed** | Three materialized, separately status-checked stages in **both** §5.2 C2 and C3, plus a `>0` count assertion, no `pipefail`. Case V shows the rev-5 pipeline exiting **0** with 2 of 3 files; X/Y/Z each **1** (find, sort, digest); AA proves **no archive** is produced |
| F3 | W2 permitted a stale Windows source tree | **fixed in specification; control-host half executed** | W2 is now absent→create→prove-empty (`Get-ChildItem -Force` = 0) before W3; W9 is status-checked **and** absence-confirmed; new negatives **W-N5/W-N6**. Harness AK **0**, AL/AM wrappers **0** (guard fired, nothing crossed the wire), AN **0**, AO **1**. `cmd.exe`/PowerShell syntax stays an unexecuted producer gate |
| F4 | the Linux confirmation command violated the map's own toolchain rule | **fixed** | §9.1 step 6 now takes one operator-approved absolute `LEV_GO_ROOT`, derives `LEV_GO_EXE`, runs the full six-assertion preflight on `lev`, then invokes that absolute launcher under `GOROOT` / `GOTOOLCHAIN=local` / `GOENV=off`. Exact `go1.25.5` everywhere a gate is described; I10 tightened |

**One self-caught defect, recorded because it is the class the harness exists for.** `/bin/sh`'s
`echo` expands `\r` inside a Windows-shape `GOROOT` (`…\rootW` → `…ootW`). It turned the AG case red
on first run for a reason unrelated to the contract. Both the stub and §6.0a's four diagnostic lines
now use `printf '%s\n'`; on a real `windows-latest` runner the same expansion would have corrupted
the recorded `GOROOT` in every job's evidence.

## Checklist reconciliation

- Item 13 `Tests green` — remains **unchecked** (read-only, no-Go scope; the append-only CLI has no
  `remove_checklist_item`, so it cannot be reworded by any agent). Unchanged.
- Item 15 (independent reviewer verification) — remains **unchecked**. Only the reviewer may check it.
- Item 17 (cycle-4 21/21 wording) — superseded by this cycle's **41/41**; the same append-only
  constraint prevents rewording it, so a new item records the current figure.

## Artifact digests

| Artifact | SHA-256 |
|---|---|
| `.research/260729_final-curator-ci-execution-map.md` (revision 6, 2962 lines) | `04639cac006c41bdc87d38db0d64458a90c3c653efc59c078a87ab080ccc29ec` |
| `verify-recipes.sh` (41 cases) | `fcb11c565a04218222c75573fe59147e9a200dfb6d0f26bbaf1f677e69baf9f9` |
| `verify-recipes-cycle5.log` | `27fded11438df3b372fb8bfa02cdde771e1ba9bab47a22863663d68c8b9f8503` |

## Preserved invariants (unchanged this cycle)

- rc.5 manifest `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`
- rc.5 tree `e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae` over **448** files
- conformance root **3 modified, 354 untracked**
- committed pin `00b1688a9b2457ca397a0bb550acf47cad8ee967`, unmoved, at `ci.yml:28` and `ci.yml:81`
- Linux native validation **non-gating** pending its named prerequisite
- every command in §5–§9 stated as a **future producer gate**; no green CI result claimed anywhere
