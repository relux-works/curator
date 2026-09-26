# TASK-260924-20o9dk — conform curator to the released skillfile-sources corpus (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md`, the task description (full scope), and `pr96-review-for-20o9dk.md` (section 6: curator gaps C1/C2 and
the failing-case table by owner). curator-spec main 5746367 now carries the ACCEPTED skillfile-sources revision 1 (PR #96, Decision 0022):
conformance/skillfile-sources-v1 + schemas/skillfile-sources-v1 (moved from draft-sources-v1; includes #90 clarifications, global-scope
vectors and five §3 lock-replay vectors). v9 manifest material is NOT in it (stays a curator draft capability; do not remove curator's v9
support here).
1. Re-vendor internal/crossconformance testdata from conformance/skillfile-sources-v1 at 5746367 (pin file, MANIFEST, README; rename the
   draft-sources-v1 vendoring; counts updated from the corpus, not hand-set).
2. Production-entry rows for EVERY case of the released corpus; no gap rows for it. Implement what fails: #90 rules (refresh uses the
   current resolved endpoint plan; SCP-like endpoint + alias port → repository_policy_invalid before I/O; SSH URI + alias port rendering;
   repository+commit revocation deny-wins under advisory), global/project schema-2 profile-lock rules, replay rows (path identical/drifted,
   git missing snapshot fetches locked commit, moved tag replays locked commit incl. through a mirror, unreachable source).
3. C1: replay verifies the locked OBJECT FORMAT (member.Package.Commit.ObjectFormat vs the fetched repository) — production fix + killing
   row. C2: driveV2DeclaredMirror must compare c.Expected (no vacuous pass).
4. Split SemanticCases so no single `go test` call exceeds ~5 minutes; keep each call bounded. Host memory is tight: never run the whole
   module test suite locally; the hosted gate is the arbiter.
5. Mutants: one per implemented rule (fetch stored origin; render SCP alias port; advisory ignores repository+commit revocation; skip
   object-format check) — survive→killed with real exit codes.
No CHANGELOG edit (entry text in results under "## CHANGELOG entry (for release prep)"); no LOGBOOK edit; artifacts only in $TMPDIR.
Attach results (case table by family: driven/passing), check DoD, `task-board handoff TASK-260924-20o9dk --role developer`. A write-boundary
`policy warn` block is a warning.
