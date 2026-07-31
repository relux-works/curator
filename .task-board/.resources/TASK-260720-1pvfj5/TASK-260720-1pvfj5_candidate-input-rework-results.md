# TASK-260720-1pvfj5 candidate-input provenance rework

Date: 2026-07-30
Role: developer
Composite: `.temp/TASK-260720-1pvfj5/rework/composite`
Disposition: ready for independent review

## Outcome

The candidate CI lane now rejects a dispatch that supplies both
`candidate_ref` and `candidate_root`. The validation runs before candidate
revision checkout and before candidate evidence recording. A caller must supply
exactly one candidate source.

Ref-only and root-only dispatch semantics remain valid. The released
`SPEC_PIN`, candidate evidence wording, accepted seven-path product delta, and
every unrelated composite byte remain unchanged.

## Exact source delta

Only these three CI paths changed relative to the accepted 374-entry final
integration manifest:

1. `.github/ci/candidate-suite.sh`
2. `.github/ci/gate-selftest.sh`
3. `.github/workflows/ci.yml`

The manifest comparison checked all 374 entries and exited 0. It found exactly
those three changed paths, no additions, no removals, and no other byte change.
Therefore all accepted product paths, including the seven blocker-owned paths,
remain byte-identical.

| Path | Accepted SHA-256 | Rework SHA-256 |
| --- | --- | --- |
| `.github/ci/candidate-suite.sh` | `e2e874a699db5d45abd81d9c0a74f2c53981de9dac6cfdcf08a560a9949e149f` | `25abd5c01ba1b35170491e460941297b6069e0eccebe30a285d1d600e1719ced` |
| `.github/ci/gate-selftest.sh` | `bd7eafe9b8a51d8b4292ec3330762fa1e2af3879009debb050edab1af9bd7a6a` | `bc2bd91e4389db1e22e5978df32116a2ee7ec4220ff1b4b830a85710ef4ac169` |
| `.github/workflows/ci.yml` | `0626efe3818add42fc5cb9b8ee4c24829755d7b2baa5eb3b87e701feee794630` | `bad1befc76d84ca3bf58c8e3369473ccc0a81d85a0867d8607290993912b0c5e` |

The workflow contains one released pin with the unchanged value
`00b1688a9b2457ca397a0bb550acf47cad8ee967`.

## Direct validation ledger

Every listed validation was run as a standalone process and its real exit code
is reported.

| Validation | Exit | Result |
| --- | ---: | --- |
| `bash -n .github/ci/candidate-suite.sh .github/ci/gate-selftest.sh` | 0 | shell syntax valid |
| `candidate-suite.sh verify-inputs <full-ref> /candidate/root` | 1 | expected-red: ambiguous inputs rejected |
| `candidate-suite.sh verify-inputs <full-ref> ""` | 0 | ref-only path accepted |
| `candidate-suite.sh verify-inputs "" /candidate/root` | 0 | root-only path accepted |
| `bash .github/ci/gate-selftest.sh` | 0 | 74 passed, 0 failed |
| `actionlint .github/workflows/ci.yml` | 0 | workflow valid |
| `shellcheck .github/ci/candidate-suite.sh` | 0 | changed implementation clean |
| `shellcheck --severity=warning .github/ci/gate-selftest.sh` | 0 | changed regression clean at warning/error severity |
| `git diff --check` | 0 | no whitespace errors |
| `git diff --cached --quiet` | 0 | nothing staged |
| current composite manifest generation | 0 | 374 entries |
| exact accepted-to-rework manifest comparison | 0 | only the three allowed CI paths changed |
| exact `SPEC_PIN` count/value assertion | 0 | one unchanged released pin |
| focused patch reverse-apply check against the live rework | 0 | patch exactly describes the three live file changes |

A default-severity `shellcheck` over both changed scripts exited 1 on two
pre-existing informational findings in untouched regions of
`gate-selftest.sh`: SC2329 for the indirectly invoked `run_gate` helper and
SC2016 for literal backticks in a single-quoted diagnostic. The focused
candidate script check and the warning/error gate both exited 0. No suppression
or unrelated cleanup was added.

The first manifest-generation attempt exited 1 because the evidence directory
was mistakenly created relative to the composite while the output path was
project-root absolute. The corrected absolute task directory was created, and
the rerun exited 0 with 374 entries. This diagnostic did not change source.

The first reverse-apply check of the hand-authored focused patch exited 128
because its last unified-diff hunk declared one extra output line. The hunk
count was corrected without touching the composite, and the exact reverse
check then exited 0.

## Reused heavy evidence

Per the mandatory focused-rework instruction, default-pin, explicit rc.5
candidate, full Go, lint, vet, build, and race suites were not rerun solely for
this CI-input validation change. Their accepted evidence remains applicable
because no product byte, pin, vector, timeout, fixture, suppression policy,
Makefile, or README byte changed:

- default-pin test gate: exit 0;
- explicit rc.5 candidate gate: exit 0;
- single serialized race gate: exit 0;
- pinned golangci-lint v2.12.2: exit 0, 0 issues;
- gofmt, vet, build, deterministic cancellation: exit 0;
- ledger consistency: exit 0, 49 rows.

The affected gate self-test was rerun because its bytes changed; its new result
is 74 passed and 0 failed.

## Scope controls

No file was staged, committed, published, pinned, or promoted. No release or
conformance claim was added. The candidate remains an explicit, non-default
qualification input.
