# TASK-260928-2iu83q — gate fix (THE ONLY CURRENT INSTRUCTION)

Rev1 (tree 3e740af5) reproduces the accepted 31gaka content exactly, but the gate fails (run 36406667044) in internal/config:
TestManagerConfigV2SchemaCases ("known-gap case manager-config-v2/schema-cases/valid-overlay-path-source.json now passes; remove its ledger
row") and TestOverlayConformancePassesWithPermissionsAndSourceSigners (valid-overlay-git-http(s)-uppercase.json "still has gap row owned by
STORY-260916-ioemse"). Cause: your .github/ci/conformance-gaps.tsv re-added rows that trunk REMOVED after 31gaka's old base (E1 1zgucp and
E6 yvxbs1 drove those cases). Fix: .github/ci/conformance-gaps.tsv must equal trunk's file (origin/main) with ONLY the 31gaka changes applied
(the Codex seed revision-B rows owned by TASK-260927-1e5qqm added, the revision-A rows removed). `git diff origin/main --
.github/ci/conformance-gaps.tsv` must show only 31gaka rows. Change nothing else. `task-board m 'set_status(TASK-260928-2iu83q,
status=development)'` first; run `go test ./internal/config ./internal/envprofile -run 'Conformance|Schema|Seed|Codex'` with real exit codes;
update results; handoff; END YOUR TURN.
