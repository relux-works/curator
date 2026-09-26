# TASK-260922-cww1ov — integration landing preconditions (integration run RUN-260926-febb25)

Task: TASK-260922-cww1ov — envprofile-0017-production-entry-tests
Story: STORY-260922-1cenbr — credential-modes-0017-implementation
Accepted change request: CR-TASK-260922-cww1ov-11, revision 11, state `accepted`.

## Landing preconditions (all confirmed read-only; no file changed)

- Board status of TASK-260922-cww1ov is `integrating` (already; no status write made).
- `task-board worktree status STORY-260922-1cenbr` (exit 0): workspace active,
  path and branch present, tip `b120247d`, tree `dirty` (uncommitted candidate),
  lease held by this run RUN-260926-febb25.
- Change-request row: `TASK-260922-cww1ov rev 11 accepted
  (repository_delta=present, 33 changed path(s))`.
- `task-board worktree integrating --json` (exit 0): protected ref
  `refs/heads/main` at `60498052a1833f7511bd16086d953c8099fe7eed`;
  TASK-260922-cww1ov classification `awaiting_landing`, revision 11,
  `accepted`, `repository_delta=present`, `candidate_tree_on_trunk=no`
  (landed_tree_not_on_trunk). The landing act is still owed and belongs to the
  orchestrator/runner integration transaction — this run executed no
  `worktree integrate`, `worktree checkpoint`, `handoff`, or `set_status`.
- Working tree: uncommitted candidate only, nothing staged
  (`git diff --cached --name-only` empty, exit 0), no commit of our own on the
  story branch (tip is `b120247d`, another task's record).
- Scope note: `git status --porcelain` shows the accepted candidate's 16
  modified tracked paths plus 2 untracked additions
  (`internal/envprofile/credential_production_test.go`, `internal/stateread/`);
  CHANGELOG.md carries a deletion-heavy diff (62 lines removed) as part of the
  accepted revision — recorded here, not re-judged.

## Validation gates (standalone processes, real exit codes)

- `go vet ./internal/envprofile ./internal/stateread ./cmd/curator` → VET_EXIT=0
- `GOOS=windows go vet ./internal/envprofile ./internal/install` → WINVET_EXIT=0
- `go test ./internal/stateread -count=1` → ok (0.543s), STATEREAD_EXIT=0
- `go test ./internal/envprofile -count=1` → ok (364.162s), ENVPROFILE_EXIT=0
- `go test ./cmd/curator -count=1 -v` → CURATOR_EXIT=1 (FAIL in 600.668s:
  `panic: test timed out after 10m0s`). Full log attached as
  TASK-260922-cww1ov_cmd-curator-suite.log (42 KB). Measured facts:
  63 top-level tests finished, ALL passed; 0 `--- FAIL` lines anywhere;
  28 top-level tests are `t.Parallel` (printed `=== RUN`/`=== PAUSE`, resume
  only after the sequential chain) and never resumed because the timeout fired
  first. At panic time the single running test was
  TestReviewProbeLegacyLaneIdentityInvalid (11s in), blocked inside
  `godriver.Session.Close → VerifyToolchain → fingerprintToolchain →
  digestToolchainRecords → digestCopyDiagnostic` (internal/godriver/
  fingerprint.go:322, goroutine runnable) via cmdInstall → install.Project →
  planExternalBuilds → Probe — i.e. file-by-file digest of the Go toolchain
  tree. Pre-timeout slowness is concentrated in draft-transport/provider
  subtests taking 9–60s each (fetch-refusal backoffs, e.g.
  TestDraftTransportProviderAdmission 59.06s). A concurrent unrelated
  `go test ./internal/install -count=1` process (not started by this run, left
  untouched) was active on the same host during the run. No assertion failed;
  the suite did not complete inside the 10m default test timeout in this
  worktree/host environment. Reported as a finding — no file changed, no
  re-diagnosis beyond this evidence, landing decision left to the
  orchestrator/runner integration transaction.

No test was weakened; no product file was modified by this run.
