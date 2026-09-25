# TASK-260924-1aa9wb review verdict — CR rev1: ACCEPTED

Reviewer: claude-opus-5-5 low. Worktree == candidate tree 23fcffb3 (only diff: untracked path_capture_test.go, which is the candidate's added file).

Checks
1. Switch gone: grep DRAFT_SOURCES|DraftSourcesV1|DraftSourcesEnabled over cmd/internal/docs/README hits only cmd/curator/draft_sources_test.go:182 (TestProjectResolveIgnoresRemovedSwitch sets =0, schema-2 resolve succeeds + lock written) and the unrelated DRAFT_SOURCES_PIN corpus pin. ParseOptions/ParseWithOptions/LoadWithOptions removed; manifest.go:104 admits 1 and 2.
2. Schema-1 never enters schema-2 path: project_resolve.go:72 gate; install via readProjectManifestDocument. Remaining CURATOR_DRAFT_TRANSPORT_RESOLUTION references are the separate transport switch (not in scope).
3. Global scope: explicit refusal added global.go:92 and main.go:1250; pinned by install_test.go:781 and global_status_test.go:72.
4. Mutants re-applied by me (disposable rsync copy):
   - M-refuse (manifest.go:104 -> `number != SchemaVersion`): KILLED by TestProjectResolveIgnoresRemovedSwitch, TestGlobalStatusRejectsSchema2SkillfileThroughCLI.
   - M-route (project_resolve.go:72 gate `&& false`, schema-1 routed into schema-2 plan): KILLED by TestProjectResolveSchema1KeepsReadOnlyMeaning, TestDraftProjectResolveHelp.
5. Validation log: hosted Test ubuntu/macos/windows, Race, Lint, Interop, Gate self-test all success; rose-air skipped (unverified).
6. Local reruns: go test internal/manifest envprofile sourcelock closure — ran; internal/crossconformance hit the 10-min go test timeout on this loaded host (known exec-stall host limit), hosted gate green is the arbiter.

Bounds / residuals (non-blocking)
- Base-vs-candidate binary byte-diff on v1 corpus not rerun by me; schema-1 identity rests on committed resolve row + hosted suite.
- Minor stale wording: cmd/curator/draft_status.go:4,131 "draft lane" comments (comments only).
- Local crossconformance timeout = host load, not a finding.
