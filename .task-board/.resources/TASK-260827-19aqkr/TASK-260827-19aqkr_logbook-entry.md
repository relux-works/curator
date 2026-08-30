Logbook entry for `LOGBOOK.md`, pending placement.

Not appended by this review. The story worktree was torn down mid-review, and
the only remaining `LOGBOOK.md` is the local `main` checkout's, which is stale
against `origin/main` by 2997 deleted and 104 inserted lines (the logbook was
pruned upstream). Appending there would fight that rewrite. The orchestrator
should land this block after rebasing the story branch onto `origin/main`.

---

## 2026-08-27

### 0640 — `TASK-260827-19aqkr`: a docs task wrote 285 lines about code that was not in its worktree

- ROOT CAUSE: the story worktree for `STORY-260827-3a5efk` was forked from `1f55f1b`, 55 commits behind `origin/main`. That tree contains no HTTPS broker at all: `internal/install/buildhttps.go`, `buildhttpsprompt.go`, and both test files exist only upstream. The doc-writer was told to verify every claim against `internal/` and had nothing to verify against, so it reconstructed the contract from the CocoaSkills sibling document supplied as a precondition. Nine blocking divergences followed, and the three most misleading ones (`--token` instead of `--git-credentials|--keyring|--token-env`, the `oauth2` username default, the multi-option credential menu) are verbatim CocoaSkills artifacts.
- FINDING: `docs/build-https.md` already exists on `origin/main`, 180 lines, accurate. The task premise (create the file) was stale before the run started. `docs/build-ssh.md` had also moved upstream, so the one-line cross-link the task scoped lands on a superseded copy.
- FINDING, THE ONE THAT WOULD HAVE HURT OPERATORS: the document states three times that an uncovered private HTTPS repository fails closed before the first fetch, and prints a fabricated `build_repository_identity_invalid` precheck diagnostic with ready-to-run commands. The code does the opposite. `cmd/curator/main.go:1338-1343` keeps the resolver nil off a terminal and `internal/install/buildhttps.go:182-187` then records provenance `anonymous` and continues. `internal/config/buildhttps.go:26-28` states the rule outright: unlike `build_ssh`, an unmatched HTTPS scope is not an error, because anonymous HTTPS is a real transport. A CI operator following the document would expect a named Curator refusal and instead get an unauthenticated fetch that fails inside Git.
- FINDING: `build_repository_identity_invalid` is an admission code for non-absolute paths and non-SSH transports (`internal/buildrepo/admission.go:31`, `credentials.go:59,87`, `httpsbroker.go:135`). The identifiers this surface actually emits are `build_repository_https_credential_missing` (a prompt header, not a failure) and `build_repository_https_credential_selection_aborted`. Neither appears in the document.
- DECISION: routed `to-dev` rather than accepted-with-fixes. The correct half (scope grammar, longest-prefix semantics, the `curator-build-https:` namespace, the `build_https` shape, the full transport-policy flag list, both cross-links, the prose style) is worth keeping, but every operator-facing string needs re-derivation against a rebased tree, and the create-versus-extend question against the existing upstream document is an orchestrator call.
- NOTE: worktree `.temp/STORY-260827-3a5efk/worktree` was removed by a concurrent session at 06:27 during this review. Work is intact on `task-board/story/STORY-260827-3a5efk` at `81e56a13` and mirrored at `.temp/docs-backup-260827/`; reviewed content md5 `c031cd47a3b6a0090e8db4f65802467f` matches the committed blob.
- STATUS: pending rework. Verdict evidence in `TASK-260827-19aqkr_review-verdict.md`.
