# TASK-260922-1t551d rework 2 (orchestrator, binding) — F-C1

Revision 2 gate FAILED on the two ubuntu lanes (run 35685549418), two pre-existing rows:
`TestResolveClaudeProjectEntry` (managed_test.go:481) and `TestClaudeSeedMergePreservesToolState`
(status_test.go:412): `environment_repair_failed: passthrough entry .credentials.json is detached:
link target …/.credentials.json does not exist (environment_credential_conflict)`. Your liveness
check (rework 1, R1) now makes REPAIR fail whenever the declared native credential file does not
exist yet — but on Linux `claude_code` is file-link (0017 O2) and the native
`~/.claude/.credentials.json` only appears after the user's first login: provisioning a
passthrough link to the DECLARED native path before that is the normal, expected state
("expected-to-detach until …"), not a conflict.

Ruling (refines brief R2 — keep the operator's case, separate the states):
- MIS-TARGETED (recorded target ≠ the strategy's declared native path — the operator's Pi case,
  `~/.pi/auth.json` vs `~/.pi/agent/auth.json`): `status` and `resolve` report DETACHED with
  `environment_credential_conflict` wording; fix-first repair relinks a recorded symlink to the
  declared path (never touching bytes) and refuses only when a regular file sits at the link path.
- DANGLING-TO-DECLARED (recorded target == declared native path, target absent): provisioning and
  repair SUCCEED (the link is correct); `status` reports the entry as detached-pending ("target
  does not exist yet — log in to <tool> to populate it") as a finding/warning, never
  `environment_repair_failed`; `resolve` surfaces the same finding without failing. Distinguish
  "absent" from "inspection failed" (EACCES ⇒ conflict).
- The reviewer's probe (`TestReviewerDanglingExpectedTarget`) expects the dangling expected-target
  case to be REPORTED (not silent) — satisfy it with the finding, not with a failure; adjust the
  committed rows accordingly and add the mis-targeted vs dangling-to-declared distinction as two
  rows with a narrowing mutant (collapse the two states → fails).
Keep everything else from rev2 (TOML parser, codex admission table). Continue from the revision-2
tree (no checkout/clean/stash); append "Revision 3" to results.md; republish only on a green gate.
