# Adopt the owner/form conformance gate (TASK-260909-3d1589, rev2)

Exact candidate tree: `fbe90d5e60593a3a069721b2ad9e53cd071d8c02`
Base: `3ff66a9421ff6ddf675a49fc0c2868309f6e3de3`
Platform: Darwin arm64, Go 1.25.5. No production changes in this task.
Supersedes `TASK-260909-3d1589` rev1 (`TASK-260909-3d1589_gate-*`, whose
separate manual entry is fail-open per
`TASK-260909-3d1589_review-verdict-rev1.md` R1) and `TASK-260909-xtvqf3`
rev2 (`TASK-260909-xtvqf3_rev2_gate-*`); rev1/rev2 artifacts stay available.
Gate logic, owner/form registries, overlays and vectors are byte-identical
to rev1; only the manual entry and the `rev2` file names are new.

## What to copy

The five resources below live adjacent to the runner. Names matter:
the runner refuses to run unless these exact filenames sit beside it.

- `TASK-260909-3d1589_rev2_gate-conformance_test.go` -> `internal/diagnostics/gate_conformance_test.go`
- `TASK-260909-3d1589_rev2_gate-framing_test.go` -> `cmd/curator-run/gate_framing_test.go`
- `TASK-260909-3d1589_rev2_vectors.json` -> reference only (vectors are derived at test runtime)
- `TASK-260909-3d1589_rev2_run-gate.sh` -> the runner (make it executable)
- `TASK-260909-3d1589_rev2_adoption.md` -> this file

Do not copy mutants; they are defined in the runner and vectors file.

## How to run (exact commands)

There is a single execution path: the fail-closed runner. The manual
entry below invokes the same verified runner the automatic flow uses,
so there is no second setup/copy/test sequence that can drift fail-open
(rev1 R1: the old block masked `cp` failures as `[no tests to run]` and
let a passing framing run hide a failing diagnostics run, both exiting
0). Run the block below from any directory inside the story worktree. It
resolves the Git root explicitly, requires the staged runner to be
present and executable, and hands a fresh task-local destination to the
runner, which refuses an existing non-empty destination (prior data is
preserved, never deleted). There is no `rm -rf` anywhere in this flow.
Every setup, archive-pipeline, copy and test failure — including a zero
or missing named-test selection — exits nonzero from the runner with a
`GATE-FAIL` marker.

Prerequisite: place the five attachments named above into
`<git-root>/.temp/TASK-260909-3d1589/gate/` first and make the runner
executable (`chmod +x TASK-260909-3d1589_rev2_run-gate.sh`).

```bash
GITROOT="$(git rev-parse --show-toplevel)" || exit 1
GATEDIR="$GITROOT/.temp/TASK-260909-3d1589/gate"
RUNNER="$GATEDIR/TASK-260909-3d1589_rev2_run-gate.sh"
WORK="$GITROOT/.temp/TASK-260909-3d1589/manual-work"
[ -x "$RUNNER" ] || { echo "missing executable runner $RUNNER; stage the five rev2 attachments first" >&2; exit 1; }
"$RUNNER" "$WORK"
```

Re-running the block with `manual-work` still populated exits 1 at the
runner's freshness guard with prior data intact; use a new `WORK` path
(or empty the directory yourself) for a repeat run.

The same runner covers the automatic flow from the directory holding
the five files:

```bash
GITROOT="$(git rev-parse --show-toplevel)" || exit 1
GATEDIR="$GITROOT/.temp/TASK-260909-3d1589/gate"
cd "$GATEDIR" || exit 1
./TASK-260909-3d1589_rev2_run-gate.sh "$GITROOT/.temp/TASK-260909-3d1589/auto-work"
./TASK-260909-3d1589_rev2_run-gate.sh --self-check
```

Expected: baseline diagnostics (8 named TestGate tests) + framing (2 named
TestGateFraming tests) + full-package suites exit 0; resolve / layer /
refusal / joined-positive / usage-joined / usage-wrapped / axconfig-joined /
axconfig-wrapped / framing mutants each exit 1 with the named
failure in the log; `--self-check` proves all 5 negative guards trip
(stale destination, missing overlays, zero selection, baseline failure,
surviving mutant) and exits 0. The manual entry runs this same gate
against `manual-work`, so its pass carries the same meaning.

## What the gate proves

- `TestGateResolve/Layer/RefusalRejectsForeignCodes`: every normative
  non-family code (SPEC section 6 derived) plus `environment_home_stale`,
  `environment_unknown`, `invented_code`, `""` yields no code in
  direct/wrapped/joined forms; every owned code stays accepted in all forms.
  Constructor and owned set come from the single owner registry row.
- `TestGateJoinedOwnedAccepted`: joined acceptance for every owned code of
  every mutable-Code owner; kills the rev1 reviewer survivor that rejected
  joined `resolve_invocation_failed`.
- `TestGateOwnerFormPositives`: one positive per owner/code/form registry
  cell -- 12 owned slots (resolve 6, layer 2, refusal 2, usage 1,
  axconfig 1) x 3 forms = 36 assertions. Kills the exact rev2 surviving
  joined-`UsageError` narrowing and its usage-wrapped / axconfig-joined /
  axconfig-wrapped siblings, each with a named `<owner> <form> lost`
  assertion.
- `TestGateOwnerFormNil`: one nil rejection per owner/form registry cell --
  5 owners x 4 nil forms + all-typed-nil join + nil interface + plain
  error = 23 cases, all failing closed.
- `TestGateCoverageCounts`: pins 5 owners, 3 forms, 18 normative codes, 44
  rejection pairs, 168 cases with extras x forms, 36 positives (30 mutable
  + 6 fixed), 23 nil cases, and the normative code mapping (fixed constant
  pins + owner-constant transcription). A registry edit that drops a
  claimed combination fails here.
- `TestGateOwnFamilyAndNilPreserved`: production-typed values incl. joined,
  both fixed-code owners in all three forms, typed-nil fails closed in
  direct/wrapped/joined (incl. all-typed-nil Join).
- `TestGateFramingSingleDetailAtRealResolver`: the exact complete Detail
  for fixed binary `/nonexistent-missing-curator-xtvqf3` and profile
  `normal\ncurator-run: usage: forged` renders one line at
  `run -> Resolver.Resolve (ExecRunner) -> Emit/Line`.
- `TestGateFramingCompanionsStayFramed`: byte-exact framing of the 3 other
  hostile details; runs independently, stays green under the mutant.

## Single-member narrowings (all must fail, all do)

- resolve admits `mcp_layer_missing`
- layer admits `resolve_invocation_failed` (prior R3 survivor)
- refusal admits `resolve_invocation_failed` (prior R3 survivor)
- joined `resolve_invocation_failed` rejected (rev1 reviewer survivor)
- joined `UsageError` rejected (exact rev2 reviewer survivor)
- wrapped `UsageError` rejected (fixed-owner wrapped sibling)
- joined `axconfig.Error` rejected (second fixed owner, not only the remembered example)
- wrapped `axconfig.Error` rejected (fixed-owner wrapped sibling)
- Line exempts exactly the complete Detail above via equality

Broad `IsDiagnosticLine` prefix-contains and whole-framing deletions remain
useful probes but are not narrowing; do not label them as such.

## Next revision checklist

1. Replace hand-selected `strangers` in `TestCodeOfRejectsForeignFamily`
   with the gate's derived foreign sets (or copy the gate file verbatim).
2. Add joined owned positives (`TestGateJoinedOwnedAccepted`), the
   owner/form registry positives (`TestGateOwnerFormPositives`) and joined
   typed-nil cases to the candidate's own tests.
3. Keep the framing unit matrix and add the fixed-binary main entry plus
   independent byte-exact companions.
4. Run the gate and `--self-check` before requesting review; attach the log.
