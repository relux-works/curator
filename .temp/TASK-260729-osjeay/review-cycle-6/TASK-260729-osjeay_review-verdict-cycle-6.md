# TASK-260729-osjeay — review verdict, cycle 6

**Role:** reviewer  
**Artifact reviewed:** `TASK-260729-osjeay_final-ci-execution-map-rev6.md`, revision 6,
SHA-256 `04639cac006c41bdc87d38db0d64458a90c3c653efc59c078a87ab080ccc29ec`  
**Verdict:** **changes requested → `analysis`**

Revision 6 closes the four cycle-5 findings in the paths covered by its extended harness. The
41-case no-Go/no-Windows harness independently exits 0. The execution map still has two
operator-facing command contracts that are not executable in this project's required `zsh`
environment, points reviewers at the superseded 21-case harness while claiming 41 cases, and retains
one contradictory Linux prerequisite. These are bounded research/document corrections, not product
implementation work, so the correct route is `analysis`.

## Independently verified

- The current 41-case board resource
  `TASK-260729-osjeay_verify-recipes-cycle5.sh` has SHA-256
  `fcb11c565a04218222c75573fe59147e9a200dfb6d0f26bbaf1f677e69baf9f9`.
  I inspected it before execution and ran it from the materialized board resource. It printed
  `ALL 41 EXPECTATIONS MET` and the real process exit was `0`.
- `main` and `origin/main` are both
  `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`; current `HEAD` is
  `c06aa1a15e4093410a686ff0ce4f579fba59dec1`.
- `origin/main:.github/workflows/ci.yml` still contains pin
  `00b1688a9b2457ca397a0bb550acf47cad8ee967` twice, `checkout@v4`,
  `setup-go@v5`, `golangci-lint-action@v7`, and mutable `version: latest`.
  The current `Makefile` remains the six-target baseline and `go.mod` requires `go 1.25.5`.
- The candidate identity remeasures exactly: manifest
  `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`,
  tree `e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae`,
  448 files, and repository-level conformance status 3 modified + 354 untracked paths.
- The accepted-to-candidate product delta remains 23 `_test.go` files (20 new, 3 modified);
  no production file appears in the delta. The six D3 hard-fail sites remain present.
- `TASK-260720-1pvfj5` remains `backlog`, blocked by completed
  `TASK-260720-2qqq0w` and in-development `TASK-260720-jrrgw9`.
- Current upstream drift claims remain source-current:
  [checkout v7.0.1](https://github.com/actions/checkout/releases/tag/v7.0.1),
  [setup-go v7.0.0](https://github.com/actions/setup-go/releases/tag/v7.0.0),
  [golangci-lint-action v9.3.0](https://github.com/golangci/golangci-lint-action/releases/tag/v9.3.0),
  [golangci-lint v2.12.2](https://github.com/golangci/golangci-lint/releases/tag/v2.12.2), and the
  [runner-images mapping](https://github.com/actions/runner-images/blob/main/README.md) confirms
  `ubuntu-latest` = Ubuntu 24.04 x64, `macos-latest` = macOS 26 arm64, and
  `windows-latest` = Windows Server 2025 x64.

No Go command, test, build, vet, format, lint, install, dependency fetch, product/spec/CI/Make/pin
edit, or `TASK-260720-1pvfj5` mutation was performed. Official GitHub pages were read for current
source verification.

## Findings

### F1 — the local preflight and gate commands are not executable under zsh

The project instructions set the shell to `zsh`, but two operator-facing command blocks rely on
POSIX `sh` scalar word splitting without declaring or invoking a POSIX shell.

First, §5.0 T-P1 packs three environment assignments into one scalar:

```sh
E="GOROOT=$GO_ROOT GOTOOLCHAIN=local GOENV=off"
v="$(env $E "$GO_EXE" version)"
```

Under zsh, `$E` stays one argument. `env` therefore sets `GOROOT` to the literal value
`$GO_ROOT GOTOOLCHAIN=local GOENV=off`; it does not set `GOTOOLCHAIN` or `GOENV`. The independent
no-Go reproduction shows the semantic difference:

```text
zsh: A=<1 CHECK_B=2> B=<>
sh:  A=<1> B=<2>
```

Second, §9.2 packs the command and a Make variable assignment into one scalar:

```sh
MK="make GOROOT_EXPECTED=$GO_ROOT"
$MK linux-package-guard
```

The exact no-Go reproduction under zsh is:

```text
zsh:1: no such file or directory: make GOROOT_EXPECTED=/tmp/approved-go
zsh_exit=127
```

The same text exits 0 under `/bin/sh`, which explains why the `/bin/sh`-only harness missed it.
Every §9.2 Make gate using `$MK` is therefore dead in the documented project shell.

**Required rework:** remove packed command/environment scalars. Spell T-P1 probes as
`env GOROOT="$GO_ROOT" GOTOOLCHAIN=local GOENV=off "$GO_EXE" ...`. For §9.2, use a zsh-safe shell
function or repeat `make GOROOT_EXPECTED="$GO_ROOT"` on each command. Extend the no-Go harness with
literal zsh cases for T-P1 assignment propagation and the §9.2 Make wrapper, recording expected and
actual exits.

### F2 — revision 6 points the reviewer at the superseded 21-case harness

Section 7.4 says the reviewer can rerun all 41 cases from outcome resource
`TASK-260729-osjeay_verify-recipes.sh` and from
`.temp/TASK-260729-osjeay/verify-recipes.sh`. Fact-ledger row 53 and the cycle-5 gate-status
artifact repeat that path.

The board truth is different:

| Resource | SHA-256 | Declared expectation |
|---|---|---|
| `TASK-260729-osjeay_verify-recipes.sh` | `65a02fbee0bffe0f5dfefbe64f89f3537b5a32185c024ce0ed2a25e90c774e5a` | `ALL 21 EXPECTATIONS MET` |
| `TASK-260729-osjeay_verify-recipes-cycle5.sh` | `fcb11c565a04218222c75573fe59147e9a200dfb6d0f26bbaf1f677e69baf9f9` | `ALL 41 EXPECTATIONS MET` |

The hash printed in revision 6 is the 41-case script, but the named resource is the 21-case script.
A reviewer following the map literally runs the superseded coverage and cannot reproduce the
claimed 41/41 result.

**Required rework:** point §7.4, row 53, and the cycle gate-status artifact to the exact
`TASK-260729-osjeay_verify-recipes-cycle5.sh` board resource and give an executable
`task-board resource get` plus `sh` command. Preserve the historical 21-case resource rather than
silently replacing it.

### F3 — §5.3 still weakens the Linux prerequisite to Go 1.25.x

Revision 6 says every gate requirement now names exact `go1.25.5`, and I10 plus §9.1 step 6 do so.
Section 5.3 nevertheless calls the named Linux prerequisite, “precisely,” an approved
“Go 1.25.x `GOROOT`.” That is the same broader prerequisite cycle 5 rejected, and it would permit
an operator to supply `go1.25.1` even though the executable preflight later rejects it.

**Required rework:** replace the §5.3 prerequisite with exact `go1.25.5`, then re-run the document
search proving the only remaining `1.25.x` uses describe rejected older-patch cases. Update the
gate-status claim that exact `go1.25.5` appears everywhere a gate is described.

## Re-review gate

Re-review revision 7 only after:

1. F1–F3 are corrected without changing product, spec, CI, `Makefile`, pins, or target-task fields.
2. The extended harness includes the exact zsh operator-facing snippets and exits 0 with every
   expected/actual result recorded.
3. The execution map and gate-status artifact reference the actual current harness resource by
   name and hash.
4. Candidate manifest/tree/count, committed pin, no-Go honesty, non-gating native Linux boundary,
   and every previously corrected fail-closed contract remain unchanged.
