# TASK-260923-em42lw — emit launch-env-fragment-v2 with the permissions member (curator)

Control root: curator; work only in your Story worktree (STORY-260923-1lu2o3). Read `campaign-producer-rules.md`
and this task's description/AC (exact). Landing gate = hosted CI, once, at handoff.

Normative source (read it, do not paraphrase from memory): curator-spec main (checkout at
/Users/administrator/Developer/ReluxWorks/curator/curator-spec, ≥ ec8dc656): environments §10.1 "Permission mode",
§10.2 (the fragment's two revisions), §12.1 (`permissions.<profile>` knob), §12.2 (force-native lock), §12.5,
manager §1 (locked list + warning), `schemas/v1/launch-env-fragment-v2.schema.json` and its cases under
`conformance/v1/schema-cases/launch-env-fragment-v2/`, Decision 0018 choice 7.

Conformance root caveat: the CI pin is rc.12 (`dced9b8`), which predates these files. Validate your output
against the schema from the local spec-main checkout in a test that reads it via an explicit path fixture you
copy byte-for-byte (cite path + hash); vector consumers of the pinned root must not regress. The forward-pin
Story (2goxjs) will make the published cases driven later.

Rules: keep v1 emission reachable only where the spec still requires it (read §10.2: until the transport exists
managers emitted v1 — now curator emits v2); never let a v1 fragment carry the member. The launcher is a
separate leaf (F-L1) — change nothing there. Mutants: one per lattice rule (locked⇔global; native when
global/default; yolo only with source=profile), each killed by a named production-entry row. CHANGELOG + docs.
Attach `TASK-260923-em42lw_results.md` and hand off with `task-board handoff TASK-260923-em42lw --role developer`.
