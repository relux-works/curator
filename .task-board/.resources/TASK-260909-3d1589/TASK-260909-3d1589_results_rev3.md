# TASK-260909-3d1589 results rev3: automatic-block short-circuit (rev2 R1 rework)

Status: **ready for review** (developer handoff; independent review is the next role).

Addresses `TASK-260909-3d1589_review-verdict-rev2.md` R1 (medium): the rev2
published automatic block ran the primary runner and `--self-check` as
consecutive commands, so a primary refusal (`GATE-FAIL ... exit 1` on a
stale destination) was masked by the passing self-check and the whole
block exited 0. Only the automatic entry is changed; the repaired manual
block, runner, overlays and vectors are preserved byte-identical to rev2.

## Fix (one line)

```bash
./TASK-260909-3d1589_rev2_run-gate.sh "$GITROOT/.temp/TASK-260909-3d1589/auto-work" || exit $?
./TASK-260909-3d1589_rev2_run-gate.sh --self-check
```

The primary invocation now short-circuits with its own status before
`--self-check` can run. Full doc diff rev2→rev3 is in the evidence log;
the only behavioral delta is this line plus the supersession prose.

## Delivered package (new board revision resources, `TASK-260909-3d1589_rev3_*`)

| File | sha256 (short) | Role |
| --- | --- | --- |
| `TASK-260909-3d1589_rev3_adoption.md` | `861431c6…` | Corrected adoption; references the unchanged rev2 runner/overlays/vectors by their existing `rev2` names — no new overlay, vector or runner |
| `TASK-260909-3d1589_rev3_gate-evidence.log` | this run | Precise doc diff, exact published blocks, replay commands/exits/hashes, preservation maps, limitation statement |
| `TASK-260909-3d1589_results_rev3.md` | this file | Handoff summary |

Unchanged rev2 package (re-verified `cmp` clean against board bytes, staged
hashes `8495946d…` / `5031d27f…` / `a404a9b7…` / `6b278b35…`):
`TASK-260909-3d1589_rev2_gate-conformance_test.go`,
`TASK-260909-3d1589_rev2_gate-framing_test.go`,
`TASK-260909-3d1589_rev2_vectors.json`,
`TASK-260909-3d1589_rev2_run-gate.sh`.
The published manual block is byte-identical to the rev2 manual block.

Provenance: base `3ff66a9`, candidate tree
`fbe90d5e60593a3a069721b2ad9e53cd071d8c02` (present in the replay clone's
object db); Darwin arm64, Go 1.25.5. No production edits, commits,
installs, tags, hosted CI, real ax/model calls, or private-record edits.
Worktree `git status` empty.

## Measured outcomes (all observed, see evidence log)

Replay env: fresh task-local `git clone` of the worktree
(`.temp/TASK-260909-3d1589-rev3/replay`, HEAD `3ff66a9`); both ```bash
blocks were extracted from the staged rev3 adoption and run verbatim
with cwd at the clone root.

| Check | Exit | Outcome |
| --- | ---: | --- |
| Exact automatic block, fresh destination | 0 | GATE RESULT: PASS; 8/8 diagnostics + 2/2 framing named PASS; full-package suites 0; all 9 mutants exit 1 with named assertions; `--self-check` 5/5 |
| Exact automatic block, populated dest (106 files) | 1 | GATE-FAIL refuses non-empty; `--self-check` never runs; sha256 map of all 106 files identical before/after |
| Exact manual block, fresh destination | 0 | GATE RESULT: PASS (same runner, same 9 mutants killed) |
| Exact manual block, populated dest (106 files) | 1 | GATE-FAIL refuses non-empty; sha256 map of all 106 files identical before/after |
| Unguarded control (rev2-exact automatic block, same populated dest) | 0 | GATE-FAIL refusal followed by SELF-CHECK 5/5 — the reviewed masking defect reproduced; deleting `|| exit $?` flips stale from 1 to 0, so this regression kills removal of the short-circuit |

The 9 killed narrowings in the fresh run: resolve/layer/refusal
foreign-code admissions, joined-`resolve_invocation_failed` rejection,
usage-joined (exact rev2 survivor), usage-wrapped, axconfig-joined,
axconfig-wrapped, and the exact framing exemption (companions stay 0,
diagnostics matrix stays 0 under the mutant).

## Reran vs inherited (explicit)

- Reran in rev3: manual fresh positive + manual stale with
  byte-preservation, automatic fresh positive, automatic stale with
  byte-preservation, unguarded-masking control, staged-vs-board
  byte-identity of all four rev2 files, manual-block text identity.
- Inherited from attached rev1/rev2 independent evidence (not rerun):
  manual-block negatives (missing overlay, zero selection, injected
  diagnostics failure — attacked by the reviewer on the identical block),
  full normative-source transcription analysis, and the registry count
  formulas as executed (not independently re-measured) assertions.
