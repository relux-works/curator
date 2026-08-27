# TASK-260729-osjeay — gate status, rework cycle 6 (revision 7)

**Date:** 2026-07-29 · **Scope:** read-only execution-map revision. No product, spec, CI,
`Makefile`, pin, or `TASK-260720-1pvfj5` field was modified.

## Evidence-honesty ledger for this cycle

Every command below was run as a standalone process — no pipe, no `tee`, no `pipefail`.

| Command | Real exit | Status |
|---|---|---|
| `sh .temp/TASK-260729-osjeay/verify-recipes-cycle6.sh` (run 1) | **0** | **green** — `ALL 55 EXPECTATIONS MET` |
| `sh .temp/TASK-260729-osjeay/verify-recipes-cycle6.sh` (run 2) | **0** | **green** — `ALL 55 EXPECTATIONS MET`, reproduced |
| `zsh .temp/TASK-260729-osjeay/f1-repro/probe.txt` | **0** | green — prints `A=<1 B=2> B=<>`, the F1 defect, outside the harness |
| `/bin/sh .temp/TASK-260729-osjeay/f1-repro/probe.txt` | **0** | green — prints `A=<1> B=<2>` |
| `zsh .temp/TASK-260729-osjeay/f1-repro/mkprobe.txt` | **127** | **expected-red** — `no such file or directory: make GOROOT_EXPECTED=…`; this is the finding, reproduced independently |
| `/bin/sh .temp/TASK-260729-osjeay/f1-repro/mkprobe.txt` | **0** | green — `make ran GOROOT_EXPECTED=/tmp/approved-go` |
| `zsh` + `/bin/sh` on the corrected forms (`fixprobe.txt`) | **0** / **0** | green — `A=<1> B=<2>` and `make ran …` in both shells |
| `git status --short -- .github/ Makefile go.mod internal/ cmd/ conformance/ .scripts/` | **0** | green — **empty output** |
| `task-board q 'get(TASK-260720-1pvfj5){id status}'` | **0** | green — `backlog`, untouched |
| `grep -n '1\.25\.x' .research/260729_final-curator-ci-execution-map.md` | **0** | green — 10 hits, classified by hand, **zero** requirement-shaped |

**No Go command of any kind was executed** — no `go`, `go test`, `go vet`, `go build`, `go list`,
`go version`, `gofmt`, `golangci-lint`. **No Windows or Linux host was contacted** — `ssh`/`scp`
remain `/bin/sh` stubs; the new cycle-6 cases contact nothing at all. **No network read of any
kind this cycle.** No install, download, or dependency fetch.

## Reviewer findings — disposition

| # | Finding | Status | Executable evidence |
|---|---|---|---|
| F1 | the §5.0 T-P1 and §9.2 command contracts are not executable under `zsh` | **fixed** | T-P1 now spells `env GOROOT="$GO_ROOT" GOTOOLCHAIN=local GOENV=off` as separate words on every probe; §9.2 uses `mk() { make GOROOT_EXPECTED="$GO_ROOT" "$@"; }`. **The same class was also present, unflagged, in the §9.1 step-6 `ssh lev` body** (ssh runs that string through *lev's* login shell) and is corrected identically. Harness group 10: AP **1**, AQ **0**, AR **0**, AS **0**, AT **0**. Group 11: AU **127**, AV **0**, AW **0**, AX **0**, AY **0**, AZ **0**. Group 12: BA **1**, BB **0**, BC **0**. Independently reproduced outside the harness (§10 row 59) |
| F2 | revision 6 pointed the reviewer at the superseded 21-case harness | **fixed** | §7.4, ledger row 53 and row 41 now name the **exact** board resource with its own SHA-256 and case count. §7.4 gives an executable `task-board resource get` + `shasum` + `sh` sequence against `TASK-260729-osjeay_verify-recipes-cycle6.sh` (`c2391ab755af5c0cb4163012eed0f690e7800fcc1228dc1d7fd71f85612e2a41`, 55 cases). A three-row table records all three harness generations and their real hashes. **The 21-case and 41-case resources are retained unchanged**; nothing was overwritten. A stated naming invariant now forbids the recurrence |
| F3 | §5.3 still weakened the Linux prerequisite to Go 1.25.x | **fixed** | §5.3 now requires an approved absolute `GOROOT` **whose launcher reports exactly `go1.25.5`**, cross-referenced to §5.0 T-P1 / I10 / I11 / §9.1 step 6, and names the `go1.25.1` case as the one it excludes. Post-edit `grep` re-run and every hit classified: 10 total, **0 requirement-shaped** (3 disqualified-case, 3 explicit exclusions, 4 meta) — §10 row 61, line numbers listed |

**Correction of a prior claim in the document itself.** The §1.2d F4 row asserted that "the two
remaining `1.25.x` mentions describe the disqualified older-patch case". That was wrong — §5.3 held
a third, and it was a *requirement*. The row now says so.

## Why the harness needed `zsh`, not just more cases

Cycles 3–5 grew the harness from 7 → 21 → 41 cases and it was green each time. It was also
`/bin/sh`-only, and both F1 defects are invisible to `/bin/sh` by construction: `sh` word-splits an
unquoted scalar, `zsh` does not. A harness that only ever runs the non-declared shell cannot fail on
a shell-portability defect no matter how many cases it has. The cycle-6 harness therefore **requires**
`zsh` and exits **2** if it is absent, rather than silently reporting a smaller green count.

Ranked by consequence, AU is the one that mattered: `MK="make GOROOT_EXPECTED=…"; $MK <target>` is
**exit 127** in the project's own shell, so every §9.2 local Make gate was dead. AP/BA fail *closed*
at T-P1 assertion 5 — bad, but only because assertions 5 and 6 exist to catch exactly that.

## Checklist reconciliation

- Item 13 `Tests green` — remains **unchecked** (read-only, no-Go scope; the append-only CLI has no
  `remove_checklist_item`, so no agent can reword it). Unchanged.
- Item 15 (independent reviewer verification) — remains **unchecked**. Only the reviewer may check it.
- Items 16/17/18 record 7/7, 21/21 and 41/41 for cycles 3/4/5. The same append-only constraint
  prevents rewording them, so a new item records this cycle's **55/55**.

## Artifact digests

| Artifact | SHA-256 |
|---|---|
| `.research/260729_final-curator-ci-execution-map.md` (revision 7, 3096 lines) | `d6e2c6a92f8c1a7da62ed0a79ddf0959541e3a0e5296650907ebbd2f838ba1f3` |
| `verify-recipes-cycle6.sh` (55 cases) | `c2391ab755af5c0cb4163012eed0f690e7800fcc1228dc1d7fd71f85612e2a41` |
| `verify-recipes-cycle6.log` (run 1) | `1a3dd11b17f99f19a56dea0f0c69df5cca56d8846cd28ce763d4e2aebf636128` |

## Preserved invariants (unchanged this cycle)

- rc.5 manifest `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`
- rc.5 tree `e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae` over **448** files
- conformance root **3 modified, 354 untracked**
- committed pin `00b1688a9b2457ca397a0bb550acf47cad8ee967`, unmoved, at `ci.yml:28` and `ci.yml:81`
- Linux native validation **non-gating** pending its named prerequisite — now stated as exact
  `go1.25.5`, which *narrows* the prerequisite, it does not relax the boundary
- every command in §5–§9 stated as a **future producer gate**; no green CI result claimed anywhere
