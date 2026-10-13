# THE ONLY CURRENT INSTRUCTION — TASK-261012-13lcu6 revision 2: re-apply the revision-1 change onto fresh main (W-LRM1r2, developer)

**What happened.**
- Revision 1 of this task, your launcher v0.3.0 change (25 files), was never reviewed.
- While it waited, main advanced to `f08f566` (LR6, TASK-261013-zaqbpr: the wave-5 owner decisions in the launcher spec). That landing changed README.md, SPEC.md, `diagrams/README.md` and four `.puml` figures.
- Both reviewer spawns refused with `revision_base_superseded`, and `worktree converge` conflicted on SPEC.md.
- The orchestrator then:
  - withdrew revision 1 (`aborted`, with a reason);
  - discarded the Story workspace;
  - returned the task to `to-dev`.
- The revision-1 delta is the board resource `TASK-261012-13lcu6_change-request_rev1.patch` on this task. Fetch it with `task-board resource get`.

**Do this:**
1. **Re-apply.** In your fresh Story workspace, based on current main (`f08f566` or later), apply the revision-1 delta with `git apply --3way` from that patch. The orchestrator checked this in a scratch clone: every file applies cleanly except SPEC.md.
2. **Resolve SPEC.md.** The one conflict is in the revision notes at the top. Keep BOTH entries:
   - first your `0.5.1-draft` entry (launcher v0.3.0, pending tag);
   - then LR6's `0.5.0-draft` documentation-revision entry.

   The specification version header stays `0.5.1-draft`.
3. **Read the merged README.md and SPEC.md once, as a whole.** Wherever your v0.3.0 text touches a passage that LR6 changed, the result must state both without contradiction. Your v0.3.0 text covers:
   - pin v0.5.62 with attested Muse builds;
   - `muse-wonder`;
   - fragment v3;
   - hosted launch unavailable and failing closed with `session_host_missing`.

   Keep LR6's text, and reword it only where consistency needs it.

   Check that LR6's figures do not present hosted launch as available: `diagrams/components.puml`, `launch-paths.puml`, `launch-states.puml` and `artifact-layout.puml`. If one does, record that in the gaps file. Do not redraw figures in this revision.
4. **Make no other change.** For the 23 files that main did not touch, the delta against the new base must equal revision 1's delta exactly.
5. **Validation.**
   - Run no local `go test` and no local `go build` (host rules). The code files are unchanged from revision 1, and the base change touched only docs and figures.
   - Run the hosted gate on the new candidate as for revision 1, using the same lanes as its outcomes `TASK-261012-13lcu6_hosted_gate.json`, `_gate1_linux.log` and `_gate2_linux.log`. Attach the results.
   - If a gate fails because of the base change, fix only that, and record it.
6. **Publish revision 2** through the normal producer handoff to review. The CR validation log must give:
   - the new base OID;
   - the `git apply --3way` result per file;
   - the SPEC.md resolution;
   - the delta-identity check for the 23 untouched files;
   - the hosted gate result.

**Preconditions still binding:** `gw-LRM1-brief.md` (the original task), `gw5-decisions.md`, `gw4b-decisions.md` and `gw4-decisions.md`.
