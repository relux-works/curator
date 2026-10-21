# TASK-261018-ujp9nh — muse-1-4-10-exclusion-replay review verdict

Verdict: **accepted**. CR-TASK-261018-ujp9nh-1 revision 1.
Candidate tree: `e6cdeb8e7632d5ff7874e58207d38c173d2785e9`.
Base and fresh origin/main: `e88fdf2dbec7f94c48b0232a350a0a9f9484fdeb`.
Repository delta is present: only the new opt-in harness and qualification record.

```json
{
  "verdict": "accepted",
  "findings": [],
  "notes": [],
  "free_hunt_findings": []
}
```

# Review sweep — TASK-261018-ujp9nh
Candidate e6cdeb8e7632d5ff7874e58207d38c173d2785e9, base e88fdf2dbec7f94c48b0232a350a0a9f9484fdeb.
All 75 candidate files match the reviewed blobs. Fresh origin/main equals base.
AC coverage inspected before implementation: 6 of 6 rows mapped to named tests/commands and entry points; default-suite reuse reviewed separately.

| Surface | Result | Evidence |
|---|---|---|
| Binary identity | held | Ungated package exit 0; TestVendorIntakeBoundary unset and wrong-digest pass, no fixture; exact digest and sandboxed identity pass in full.log. |
| Isolation | held | TestMuse1410OfflineBoundary positive/negative network and synthetic Keychains controls pass in full.log; env -i, echo, security mach lookup and real native directories denied. |
| P1–P4 per form | held | Full gated package exit 0, 36 matrix cases plus 9 regression/parser/workspace cases. All observations reproduce: exec exclusion zero skills; interactive linked/snapshot boolean retains one; parser refusals exit 2 and exec config key accepted. |
| §7.9 status line | held | Compared reproduced rows to record and normative sections: exec measured rows qualify, interactive refused; green witness suite does not imply interactive qualification. |
| Narrowing control | held | Plant exit 1 on user_skills=1 versus required 0, neighboring no-exclusion control passes. One-skill allowance mutant exit 1: checker admitted one foreign skill. |
| Spec citations | held | Fresh remote main e41f96387287f5e76604627f9e63662f7d6d86d4 matches record; environments §§7.1, 7.9, 10.2, CIP-0017 and Muse v4 README inspected. |
| Scope and hygiene | held | Exact 75-file delta limited to hidden harness and qualification record; protected root files equal fresh trunk. Personal-path screening clean (relative golden/home fixtures are not personal paths). Source digest bindings match. |

## Validation and evidence boundaries

Coverage: **6 of 6 AC rows driven**. The candidate coverage map names vendor
exec/TUI parsers and loaders, Curator profile install and env resolve, the
published snapshot checker, child vendor-test intake, and sandbox-exec
curl/cat controls. Every listed surface was swept; held describes these
attacks only, not proof of absence.

Reviewer reran `env -u MUSE1410_BINARY go test ./.research/muse1410 -count=1 -v`
(exit 0), the full explicitly gated package with exact pinned binary,
reviewer-built Curator, fresh private run root, pinned spec and prepared
reference Python (exit 0, 74.380s), and `go vet ./.research/muse1410` (exit 0).
The full run includes 36 matrix cases, 2 parser refusals, 1 exec config-key
observation, 2 interactive regression witnesses and 4 workspace cases.
Vendor exits: 43 zero and 2 parser exit 2; sandboxed identity also exit 0.
The recorded per-form facts reproduce. Interactive retains the foreign skill
with boolean exclusion on both linked and regular-file roots; this is a
successful failure witness, not interactive qualification.

Attacks independently rerun:
- Missing opt-in and synthetic wrong digest through child vendor-test entry;
  both leave the vendor fixture absent, with skip/refusal respectively.
- Offline curl positive/negative and synthetic Keychains/ordinary-file controls.
- Real vendor plant: `MUSE1410_OMIT_EXCLUSION=exec go test
  ./.research/muse1410 -run '^TestMuse1410LoaderMatrix$/exec/linked/skills'
  -count=1 -v`, with private binary/Curator/root environment: exit 1 for the
  intended user_skills=1 versus 0 assertion; adjacent control passes.
- Isolated copy with exact checker-one-skill narrowing applied: `go test
  ./.temp/review-ujp9nh/mutant -run '^TestObservationChecker$' -count=1 -v`,
  exit 1, intended assertion `checker admitted one foreign skill`.
- Free hunt beyond the seven surfaces: contradictory count/outcome/event,
  quoted decoy, overflow/malformed count, listed symlink, unlisted dangling
  symlink and extra snapshot entry controls through the candidate behavioral
  suite, including the published reference CLI. No additional finding.

Reused rather than rerun: producer build/vet/lint and other narrowing mutants;
unchanged default-suite hosted evidence at exact checkpoint. Fresh hosting
API read confirms run 39193852508 completed/success at the checkpoint.
All 1,671 input-manifest entries match baseline blobs, source/test/config
paths are unchanged by this CR, and the original hosted environment remains
the boundary of that evidence. The attached producer summary covers 85 of 85
default packages (80 pass, 5 without tests). Local full default tests are
explicitly failed/incomplete, not relabeled green or treated as equivalent
to hosted execution. New hidden opt-in code was independently rerun here.

Evidence: `TASK-261018-ujp9nh_review-evidence-rev1.md` contains reviewer logs.
No product code or candidate bytes changed; the narrowing copy and all raw
work remain ignored. No commit or landing performed by this reviewer.

## Review logbook

Exact binary intake, offline controls and record assertions held. The exec
parser accepts the interactive configuration spelling contrary to the brief's
anticipation; the candidate correctly records the observation and bounds its
unmeasured semantics. Interactive P4 still fails, so its §7.9 gate remains
pending. Accepting the record does not authorize runtime admission changes.
The run is not goal-bound, confirmed with `task-board spawn goal`.
