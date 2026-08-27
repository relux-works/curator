# TASK-260729-osjeay review verdict — cycle 4

**Verdict:** changes requested  
**Route:** `analysis` (the deliverable is a read-only execution-map/research artifact)

Revision 4 closes cycle-3 findings F1–F4 at the narrative level and materially improves the plan.
Independent checks reproduced:

- attached map SHA-256 `73a2da0ea21b71a0da5723e13436767850c8aff58d419257fba06776b113c36e`;
- manifest SHA-256 `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`;
- tree SHA-256 `e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae`
  over 448 files;
- conformance status 3 modified plus 354 untracked paths;
- main/origin-main `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`, with the documented
  `00b1688a…` pin on that base;
- target task rc.4 wording and dependency state (`2qqq0w=done`, `jrrgw9=development`);
- 23 product delta files between accepted comparison and candidate, all `_test.go`;
- the attached no-Go harness: **ALL 7 EXPECTATIONS MET**, real process exit 0.

No Go command, build, test, vet, lint, fetch, install, product/spec/CI/Make/pin/target-task edit, or
heavy test was performed by this review.

The producer note cites a stale revision-4 SHA `74414d…`; the attached board payload and its current
resource description agree on `73a2da…c36e`. Treat the board payload as authoritative and correct
the note on the next handoff.

## Blocking findings

### F1 — `require-toolchain` does not enforce the absolute matching-toolchain invariant it claims

Revision 4 correctly requires one operator-approved absolute `GO`, matching `GOROOT`, matching
`GOROOT/bin/gofmt`, `GOTOOLCHAIN=local`, and `GOENV=off` on native hosts. The proposed Make target
does not enforce that contract:

- it has no expected-root input and never compares `$(GO) env GOROOT` with an approved root;
- it only checks that the reported root contains *a* `bin/gofmt`; it never verifies
  `$(GOFMT) == <reported-root>/bin/gofmt`;
- it never checks or supplies `GOTOOLCHAIN=local` or `GOENV=off`;
- therefore an absolute executable paired with a different absolute formatter/root, or an older
  launcher allowed to auto-switch toolchains, can satisfy the target.

The 7-case harness verifies relative-path and wrong-version failures, but has no mismatched-root,
mismatched-formatter, `GOTOOLCHAIN=auto`, or user-GOENV case. Its green result is valid but narrower
than invariant I11.

Required correction: make the native contract executable, not caller discipline. Add an explicit
approved-root input, compare the reported root byte-for-byte, require the formatter path to be that
root's `bin/gofmt`, and invoke native Go commands with `GOROOT`, `GOTOOLCHAIN=local`, and `GOENV=off`
inside the recipe (or fail closed when any differs). Preserve the hosted-runner exception only for
path discovery; it must not bypass version/root identity. Extend the no-Go stub harness to prove the
cross-root and auto-toolchain cases fail while preserving the task's executable harness evidence.

### F2 — the Windows lane prints critical values instead of enforcing them

W1 and W4 are described as fail-closed, but their commands only print:

- `go version`;
- `go env GOROOT`;
- candidate archive SHA-256;
- source archive SHA-256.

The comments say “must equal”; no command compares the output with the recorded constants or exits
non-zero on a mismatch. W5 eventually authenticates the extracted candidate, but nothing
fail-closed verifies `SRC_TAR_SHA256` before the source under test is used.

The transferred batch runner also captures `RC`, prints `EXITCODE=%RC%`, then ends with `endlocal`.
It never executes `exit /b %RC%`, so the `ssh ... cmd /c ...` process status is not contractually
the Go test status. W8 retrieves only the test log, not the stdout line that carries `EXITCODE=`.
Microsoft documents `exit /b <exitcode>` as the operation that sets the batch/process status:
https://learn.microsoft.com/en-us/windows-server/administration/windows-commands/exit

Required correction: compare every W1/W4 value in an executable PowerShell or batch assertion and
exit non-zero on mismatch. End the runner with a parse-safe preservation pattern such as
`endlocal & exit /b %RC%`, make the SSH exit code authoritative, and retain the printed code in a
persisted evidence file. Add producer-time negative checks for wrong archive digest and non-zero
stub test status before accepting the Windows command contract.

### F3 — source materialization reintroduces an unchecked-pipeline failure

The source staging command is:

```sh
tar -cf - ... | tar -xf - -C "$SRCSTAGE"
```

`set -e` observes the pipeline's final command, not the first producer command, unless a shell with
`pipefail` is explicitly selected. A source-side `tar` failure can therefore be masked by a
successful extraction of a partial stream. The subsequent checks prove only that `go.mod`, the
submodule path, and `vendor/` exist; they do not prove that every intended source/package entered
the bundle. This is the same partial-input class revision 4 correctly removed from `go list`.

Required correction: avoid the pipe or specify and verify a pipefail-capable shell. Prefer an
intermediate source archive with separately checked creation/listing/extraction statuses and a
source inventory or other complete-set assertion. Then make the W4 `SRC_TAR_SHA256` comparison
executable as required by F2.

### F4 — current runner/action dependency drift is not dispositioned

The task DoD requires current CI/toolchain/dependency drift and runner constraints to be
source-verified. Revision 4 inventories the repository's current `checkout@v4`, `setup-go@v5`, and
`golangci-lint-action@v7`, but does not record that the current upstream majors are checkout v6,
setup-go v6, and golangci-lint-action v9, or decide whether 1pvfj5 intentionally retains the older
majors. It also uses moving `*-latest` runner labels without recording their current
architecture/image mapping or the migration risk.

Primary sources:

- GitHub-hosted runner labels and `-latest` behavior:
  https://docs.github.com/en/actions/reference/runners/github-hosted-runners
- runner-image label/migration policy:
  https://github.com/actions/runner-images
- setup-go v6 current behavior and runner floor:
  https://github.com/actions/setup-go
- checkout v6 release:
  https://github.com/actions/checkout/releases
- golangci-lint-action compatibility table:
  https://github.com/golangci/golangci-lint-action

Required correction: add a dated dependency/runner ledger. For each action major, explicitly retain
with a compatibility rationale or include an upgrade in the exact producer plan. For runner labels,
record the current OS/architecture mapping and either pin versioned labels or state why moving
`latest` labels are intentional and what revalidation occurs at producer time. Do not silently
expand implementation scope; make the choice explicit.

## Re-review gate

Revise only task-scoped execution-map, gate-status, and harness evidence. Preserve the verified pin,
candidate manifest/tree/count identities, 3/354 status, Linux native non-gating prerequisite,
race-evidence corrections, D7 checklist reconciliation, and future-gate posture. Do not edit
product, CI, Makefile, target-task fields, spec, or pins.

Re-review can accept when:

1. native Make recipes enforce the full matching-root/no-auto-toolchain invariant;
2. Windows digest/toolchain/test exit contracts fail closed and persist authoritative results;
3. source staging cannot accept a partial producer stream; and
4. current action-major and runner-label drift is source-verified and explicitly dispositioned.

