# TASK-260909-3d1589 results: complete owner/form registry gate + executable adoption

Status: **ready for review** (developer handoff; independent review is the next role).

Closes the two repeated gate classes from `TASK-260909-xtvqf3_review-verdict-rev2.md`:
R2 (fixed-code owners lack claimed positive forms) and R1 (manual adoption
deletes its own destination). Single exhaustive registry, executable
instructions, artifact-only: repository delta is empty, all prior
`TASK-260909-xtvqf3` artifacts preserved untouched.

## Delivered package (board outcome resources, `TASK-260909-3d1589_*`)

| File | sha256 (short) | Role |
| --- | --- | --- |
| `TASK-260909-3d1589_gate-conformance_test.go` | `8495946d…` | Registry-driven owner/form gate (overlay) |
| `TASK-260909-3d1589_gate-framing_test.go` | `5031d27f…` | Preserved framing probe (overlay) |
| `TASK-260909-3d1589_vectors.json` | `a404a9b7…` | Registry, forms, mutants, provenance |
| `TASK-260909-3d1589_run-gate.sh` | `9c2fb8bb…` | Fail-closed runner, 9 mutants + 5 negatives |
| `TASK-260909-3d1589_adoption.md` | `085364fd…` | Fixed manual + automatic adoption |
| `TASK-260909-3d1589_gate-evidence.log` | `b1ec84ce…` | Complete run evidence (this result's basis) |
| `TASK-260909-3d1589_results.md` | this file | Handoff summary |

Provenance: base `3ff66a9`, candidate tree
`fbe90d5e60593a3a069721b2ad9e53cd071d8c02`; all 8 pinned blob OIDs and
SPEC sha256 `5a7ccf0b…` independently re-verified (see evidence log).
Darwin arm64, Go 1.25.5. No production edits, commits, installs, tags,
hosted CI, real ax/model calls, or private-record edits.

## What changed vs rev2

- **One owner registry** (`gateOwnerRegistry`: usage, resolve, layer,
  refusal, axconfig with kind, owned codes, owner-constant pins, call
  site, constructor, typed nil) crossed with **one form registry**
  (`gateFormRegistry`: direct/wrapped/joined). All positives
  (`TestGateOwnerFormPositives`, 36 cells), all nils
  (`TestGateOwnerFormNil`, 23 cases) and the mutable foreign matrices
  derive from these tables; `TestGateCoverageCounts` asserts the full
  arithmetic (5 owners, 3 forms, 18 normative, 44 pairs, 168 rejection,
  36 positive = 30 mutable + 6 fixed, 23 nil) plus the normative code
  mapping (fixed constant pins, owner-constant transcription).
- **Fixed owners truly covered**: `TestGateOwnerFormPositives` pins
  usage and axconfig in all three forms; `TestGateOwnFamilyAndNilPreserved`
  gains the 6 fixed direct/wrapped/joined positives (its broader comment is
  now accurate). `fixed_code_note` in vectors matches measured coverage.
- **Runner proves 9 narrowings**: 3 foreign admissions, joined-mutable
  rejection (rev1 survivor), **4 fixed-owner narrowings** (usage-joined =
  exact rev2 survivor, usage-wrapped, axconfig-joined, axconfig-wrapped),
  1 framing exemption — each exit 1 with its named assertion; baseline 8/8
  + framing 2/2 + full suites exit 0; 5/5 self-check negatives trip.
- **Manual adoption fixed**: explicit `GITROOT`/`GATEDIR` resolution,
  fresh task-local `WORK` that must be absent-or-empty (refuses non-empty,
  prior data preserved), `git -C "$GITROOT" archive`, no `rm -rf`. The
  exact published block was extracted from the doc and run verbatim.

## Measured outcomes (all observed, see evidence log)

| Check | Exit | Outcome |
| --- | ---: | --- |
| Exact published manual block | 0 | 8/8 diagnostics + 2/2 framing PASS |
| Manual block rerun (populated dest) | 1 | Refuses non-empty, prior data intact |
| Published full runner | 0 | Baseline 0; all 9 mutants exit 1 named |
| `--self-check` | 0 | 5/5 negative guards trip |
| Gap probe: 4 fixed mutants vs exact rev2 gate | 0 | All 4 SURVIVE rev2 (incl. exact prior survivor) |
| Gap probe: 4 fixed mutants vs new registry test | 1 each | Killed with named `<owner> <form> lost`; nil matrix stays green; no panics |
| Worktree `git status` / `git diff` | — | Empty (artifact-only) |

## Findings for the logbook

1. **Owner `Unwrap` methods panic on typed-nil receivers**
   (`axconfig.Error`, `fragment.ResolveError`, `composition.LayerError`,
   `systemprompt.Refusal` dereference `e.Err`/`e.Cause`). Production
   `chain()` is safe only because it skips `nilValue` links first. Any
   future `CodeOf` pre-check that calls `Unwrap` directly must guard the
   same way — my first wrapped-mutant draft panicked on the rev2 nil loop
   and was fixed with the existing `nilValue` helper before delivery.
2. **Fix verified stronger than required**: each fixed mutant was proven to
   survive the *exact published rev2 gate* (exit 0), not just my new tests —
   so this closes the demonstrated R2 gap rather than a remembered example.
3. No external blocker; no Stop-The-Line. Remaining: independent review of
   this package, then production adoption in a diagnostics revision.

## Checklist mapping

- Registry positive/nil + foreign coverage asserted: done (36/23/168).
- Runner detects both fixed-owner mutants + prior mutants; baseline and
  five negatives pass: done (9 mutants, 5/5 self-check).
- Exact manual block + automatic flow in fresh task-local destinations,
  fail-closed, prior data preserved; package + provenance + review evidence
  attached: done (evidence log + this file).
- Code per description/AC; task-scoped outcome artifact; findings recorded:
  done (this file; findings above).
