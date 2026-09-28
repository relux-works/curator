# Review note for TASK-260922-1hla8q revision 1 (orchestrator, binding) — F-S2 fragment permission members

Control root: curator-spec (separate board owner). Read the brief `1hla8q-brief.md` (precondition),
the producer's `TASK-260922-1hla8q_results.md`, and the source: Decision 0018 adoption choice 7 and
the Compatibility section, plus environments §10.1/§10.2/§12.1/§12.2/§13 and manager §12.5. The
board validation command (`spec-gate.sh` = `make validate`: validate.py, unittest, `go test
./tools/...`) ran at handoff with exit 0 — reuse that evidence, do not run the full gate again; run
the narrow checks you need (`python3 tools/validate.py`, targeted unittest rows, the generator's
regenerate-check in a disposable copy).

The producer fixed what choice 7 left open:
- token = the fragment revision `launch-env-fragment-v2` (REQUIRED `fragment` member); transport
  support is established iff the fragment is v2 or later;
- member `permissions { mode: native|yolo, locked: bool, source: profile|global|default }`, REQUIRED
  in every v2 fragment, closed object;
- schema-enforced consistency: `locked` true iff `source == global`; `mode` is `native` whenever
  `source` is `global` or `default` — so `yolo` occurs only as `{yolo, locked:false, source:profile}`;
- v1 fragments get a new `invalid-permissions-member` case (a v1 carrying the member is invalid);
- headless marker set {CI, GITHUB_ACTIONS} stated once in §10.1, mirrored in the launcher SPEC.

Judge, independently:
1. **Is the consistency lattice right?** `source:global` is defined as "the fleet-global system-file
   lock fixed it", never the launcher-global `defaults.json`. Check that against environments §12.2
   and manager §1: is a `global` source without an engaged lock really impossible, and does the
   launcher have everything it needs to distinguish a SILENT profile level (`source:default`) from
   an explicit `native` (`source:profile`)? A lattice that cannot express a real state is a finding.
2. **Schema fidelity**: read `schemas/v1/launch-env-fragment-v2.schema.json` and verify it is v1 plus
   the member with `additionalProperties:false` everywhere, that v1's file is byte-untouched, and
   that the three if/then rules actually reject the four contradictory shapes. Attack it: craft at
   least two shapes the producer did not enumerate (e.g. `permissions` present but empty; `mode`
   valid with an unknown extra key inside `permissions`; `source:global` with `mode:yolo`) and run
   them through `tools/validate.py`.
3. **Generator/regenerate-check**: the 14 v2 case files and the index/manifest deltas must be
   reproducible — regenerate in a disposable copy and diff against the committed tree (the campaign
   recipe: staged candidate, not `git archive`). A hand-edited case file is a finding.
4. **Normative text**: §10.1/§10.2/§12.5 and the appended 0018 choice-7 row must agree with the
   schema and with each other, the fail-closed rule must cite the token, and the marker set must be
   stated in exactly one normative place with mirrors named. Check that the adoption history of
   0018 was not rewritten (only the fixed values appended).
5. **Follow-ups named**: the curator emission leaf (manager emits v2 with the member) and F-L1 must
   be named with the exact grammar the orchestrator will paste into a brief.

Bounds you may accept: no launcher/curator code in this leaf; the launcher SPEC §4.1 mirror lands
with F-L1; frozen v1 protocol schemas untouched (the fragment is an environments-spec object).

Record exactly one verdict: `accept_cr(TASK-260922-1hla8q, revision=1, evidence=<your outcome
resource>)` on ACCEPT, or a changes-requested verdict routed with `set_status` naming file:line and
an executable reproduction. Do not write into any control root's LOGBOOK.md.
