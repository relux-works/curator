# TASK-260909-3d1589 results rev2: fail-closed manual entry (R1 rework)

Status: **ready for review** (developer handoff; independent review is the next role).

Addresses `TASK-260909-3d1589_review-verdict-rev1.md` R1 (medium): the rev1
published manual block exited 0 on a missing conformance overlay (`cp`
failure followed by `[no tests to run]`) and on a failing diagnostics
gate (a passing framing run masked it). Only the manual entry is changed;
the review-accepted owner registry, full matrices, framing probes,
automatic runner and all prior artifacts are preserved byte-identical.

## Fix

The rev2 manual entry is the verified automatic runner as the single
execution path — there is no second setup/copy/test sequence:

```bash
GITROOT="$(git rev-parse --show-toplevel)" || exit 1
GATEDIR="$GITROOT/.temp/TASK-260909-3d1589/gate"
RUNNER="$GATEDIR/TASK-260909-3d1589_rev2_run-gate.sh"
WORK="$GITROOT/.temp/TASK-260909-3d1589/manual-work"
[ -x "$RUNNER" ] || { echo "missing executable runner $RUNNER; stage the five rev2 attachments first" >&2; exit 1; }
"$RUNNER" "$WORK"
```

No `mkdir -p "$GATEDIR"` (which masked a missing gate dir), no
subshell-masked `go test` lines, no bare `cd`. The runner's explicit
checks (`PIPESTATUS` on the archive pipeline, `cp` failures, `GATE-FAIL`
on `no tests to run`/compile errors/baseline exits) now protect the
manual route. The rev2 automatic block additionally sets `GITROOT`
itself and guards its `cd` (rev1 relied on shell carry-over).

## Delivered package (board outcome resources, `TASK-260909-3d1589_rev2_*`)

| File | sha256 (short) | Role |
| --- | --- | --- |
| `TASK-260909-3d1589_rev2_gate-conformance_test.go` | `8495946d…` | Unchanged rev1 overlay (byte-identical) |
| `TASK-260909-3d1589_rev2_gate-framing_test.go` | `5031d27f…` | Unchanged rev1 overlay (byte-identical) |
| `TASK-260909-3d1589_rev2_vectors.json` | `a404a9b7…` | Unchanged rev1 vectors (byte-identical) |
| `TASK-260909-3d1589_rev2_run-gate.sh` | `6b278b35…` | Rev1 runner, only overlay/self names → `rev2` |
| `TASK-260909-3d1589_rev2_adoption.md` | `7819a5d0…` | Single-path manual + automatic adoption |
| `TASK-260909-3d1589_rev2_gate-evidence.log` | this run | Complete replay evidence (basis below) |
| `TASK-260909-3d1589_results_rev2.md` | this file | Handoff summary |

Provenance: base `3ff66a9`, candidate tree
`fbe90d5e60593a3a069721b2ad9e53cd071d8c02`; all 8 pinned blob OIDs and
SPEC sha256 `5a7ccf0b…` re-verified inside the replay clone (see evidence
log). Darwin arm64, Go 1.25.5. No production edits, commits, installs,
tags, hosted CI, real ax/model calls, or private-record edits.
Worktree `git status` empty.

## Measured outcomes (all observed, see evidence log)

Replay env: fresh task-local `git clone` of the worktree
(`.temp/TASK-260909-3d1589/rev2/replay`, HEAD `3ff66a9`, tree present);
the exact first ```bash block was extracted from the staged rev2 adoption
and run verbatim with cwd at the clone root.

| Check | Exit | Outcome |
| --- | ---: | --- |
| Exact manual block, fresh destination | 0 | GATE RESULT: PASS; 8/8 diagnostics + 2/2 framing named PASS |
| Exact manual block, populated dest + SENTINEL | 1 | GATE-FAIL refuses non-empty; SENTINEL bytes intact |
| Exact manual block, conformance overlay absent | 1 | GATE-FAIL missing overlay source (was exit 0 in rev1) |
| Exact manual block, zero named tests selected | 1 | GATE-FAIL zero tests selected (shell-only fix would miss this) |
| Exact manual block, injected diagnostics failure | 1 | Named `--- FAIL: TestGateCoverageCounts` + GATE-FAIL baseline exit=1 (was exit 0 in rev1) |
| Automatic rev2 runner, fresh auto-work | 0 | Baseline 0; all 9 mutants exit 1 with named assertions |
| `--self-check` | 0 | 5/5 negative guards trip |

Staged overlays restored byte-identical (`cmp` clean) after each
negative probe. Prior `TASK-260909-3d1589_*` rev1 resources untouched.

## Checklist mapping

- R1 manual fail-open closed for setup/archive-pipeline/copy/test AND
  zero/missing named tests: done (5/5 verbatim replays above).
- Accepted registry/matrices/probes/runner preserved: done (3 files
  byte-identical; runner diff is names only; automatic + self-check green
  under `rev2` names).
- Exact evidence + corrected package as new revision resources: done
  (evidence log + this file).
- Artifact-only, no production changes: done (`git status` empty).
