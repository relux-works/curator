# TASK-260825-1d0eo5 — review verdict: ACCEPTED

Reviewer run RUN-260825-231105, 2026-08-25. Verdict: **accepted → done**.
The implementer run (RUN-260825-69de9d) was killed at its 55-minute timeout
*after* the merge landed, so every claim below was re-verified independently
rather than taken from the run's own report.

## What was verified, and how

### Landing mechanics

- **PR #43 is MERGED** at 2026-08-25T04:38:33Z as `9bba77de355c380cf818a14980e1d9a66588e234`;
  `origin/main` moved `e027667 → 9bba77d` (verified via `gh pr view` and
  `git fetch`, not the implementer's report — the run died 40 seconds later).
- **Branch cut from origin/main**: root commit `1d34f71`'s parent is `e027667`,
  the pre-merge main head. Merge commit parents are `e027667` + `1c493e2`.
- **Merged tree is byte-identical to the PR head tree**
  (`git diff 1c493e2 9bba77d` is empty), so PR-head CI covers the merged
  content exactly.
- **CI on the PR head, all SUCCESS**: Test ×3 platforms, Race ×2, Lint,
  Gate self-test ×3, **Interop conformance gate**, Naming gate.
  `Candidate suite` is SKIPPED **by design** — its `if:` gates it to
  `workflow_dispatch` with a candidate input; it never runs on pull requests.
- **Post-merge CI on main** (`9bba77d`): at review time 9/12 jobs completed,
  all successful (incl. Interop conformance gate); the three remaining jobs are
  the known-slow macOS test/race and Windows test lanes, zero failures, and the
  identical content already passed all three on the PR head.

### Scope and fidelity of the composite

- **27 files, +4473/−45**, every file epic-owned (source, tests, docs, one
  ledger row in `.github/ci/skip-classes.tsv` from the Windows skip
  classification). No board state, no `.temp/`, and the primary checkout's
  unrelated `.github/ci/*.sh` edits were confirmed absent.
- **Source fidelity**: every source/test file byte-identical to the
  STORY-260825-32bopo delivery worktree (the newest superset), except the two
  deltas earned on the branch by the Windows CI lane (`e8a16e2` wrapper
  copy-not-hardlink, `1c493e2` skip classification) — both reviewed, both
  correct and narrowly scoped.
- **Docs fidelity**: `docs/build-https.md` byte-identical to the
  STORY-260825-39h6vz docs-of-record; README credential bullet identical;
  CHANGELOG kept **both** sides (feature entry added above the docs task's
  entry); `internal/config/config.go` / README / CHANGELOG extra content
  verified as origin/main-side drift correctly preserved by the three-way patch
  (e.g. the ratified-lockability comment exists at `e027667`).
- **LOGBOOK reconciliation**: all nine epic entries present under 2026-08-25 in
  descending order plus the landing entry `0650`; `0052`/`0057` preserved
  verbatim as the report claimed.

### Policy gates, re-run by the reviewer

- **Naming gate re-run verbatim on an archive of the merged tree**: zero
  matches outside README.md, exactly one README line.
- **No absolute local paths** anywhere in the full PR diff (`grep /Users/`
  over `e027667..9bba77d`: clean).
- **Commit messages** (all 11): reference the Curator Specification
  (Spec core §12.2, §6.1, §6.3, §7.2) and this repository only. The scrubbed
  logbook entry `0130` states explicitly that no code from any other
  implementation was copied.

### Local gate evidence (attached tarball, corroborating CI)

gate-selftest 81 passed / 0 failed; ledger-consistency 80 rows ok across
linux/darwin/windows; lint 0 issues; no-broad-suppression ok; test-gate
`go test exit=0, platform-case gate exit=0` with 11 recorded skips —
matching the landing report exactly. fmt/build/vet local logs were not in the
tarball but all three are proven by the CI Test and Lint jobs on 3 platforms.

### Code review (merged content)

- `internal/gitcred` — single mechanism (`git credential fill|approve|reject`);
  prompting closed on every surface (askpass vars dropped, not emptied, because
  git reads empty as unset; `core.askPass=` + `credential.interactive=false`
  on the command line; case-insensitive env suppression for Windows); writes
  proven by read-back; `WaitDelay` bounds the orphaned-helper case; helper
  output bounded at 64KiB; protocol injection refused.
- `internal/buildrepo` broker — answers only the two exact Git prompts for the
  pinned host; state file secret-free, strictly validated (absolute path,
  regular file, ≤4KiB, `DisallowUnknownFields`, single JSON value); secret
  redacted under `%s` and `%#v`; wrapper materialized `O_EXCL` in a 0700 dir.
- `internal/install` resolver — precedence exactly as ratified: host-pinned
  override → longest configured scope → interactive prompt → anonymous;
  named-but-unreadable sources fail closed with source-specific remedies while
  only unmatched repositories degrade to anonymous (the absence-vs-read-failure
  boundary is drawn at the right layer); nil resolver is the headless path.
- CLI — three mutually exclusive source flags, no flag accepts a literal
  token; `login` prompts hidden; `remove` deletes stored material only for a
  keyring scope.
- Architecture fit: mirrors the established `build_ssh` surface, shares the
  scope grammar via one generic `longestScope` helper instead of duplicating,
  stays out of `LockableKeys` per the ratified decision.

### Negative-evidence check

Every gating behavior ships a test that fails if the gate admits what it must
reject, several pinned against a real git binary:
`TestReadHostRefusesAManagerNamespacedAnswer`,
`TestReadScopedRefusesAnswerForAnotherUsername`,
`TestStoreRejectsAHelperThatPersistsNothing`,
`TestStoreRefusesValuesTheProtocolCannotCarry`,
`TestACallIsBoundedWhenTheHelperOutlivesGit`,
`TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts`,
`TestHTTPSBrokerStateContainsHostAndUsernameOnly`,
`TestSelectedHTTPSFetchEnvironmentIsScopedAndOverridesBothAskPassSurfaces`,
`TestBuildHTTPSSelectedSourcesFailClosedWithExactRemedies`,
`TestBuildHTTPSHostPinMakesOtherHostsResolveWithoutTheOverride`,
`TestPromptedBuildHTTPSAbortStopsTheProductionPlanBeforeAnyFetch`,
`TestHTTPSPromptThisRunOnlyNeverReachesPersistenceOnEitherCredentialSurface`,
`TestHTTPSThisRunOnlyPromptNeverReachesConfigOrCredentialStore`,
`TestConfigBuildHTTPSRemoveNeverTouchesTheOperatorsOwnGitCredential`,
literal-secret-in-token rejection at parse and validate level. The broker and
gitcred tests drive the production entry points (`main()` broker dispatch,
real git, real fetch).

## Discrepancies found (none material)

- `landing-final.md` says "26 files, +4457/−45"; the true merge-range diffstat
  is 27 files, +4473/−45 (it misses the `skip-classes.tsv` row and the
  Windows-fix delta). The landing-report's own account is consistent.
- The board notes' "Eight signed commits" predates the ninth (logbook) commit;
  the report correctly says nine (eleven after the two CI fixes).
- The 32bopo worktree's broker files were synced to the fixed state at merge
  time (mtime 04:38Z), so "the worktrees hold pre-landing copies" is one
  reworded skip-reason line away from exact — immaterial; the supersession
  note stands.

## AC → verdict

| AC clause | Result |
| --- | --- |
| Branch applies clean on origin/main | ✅ merged; root parent = pre-merge main head |
| fmt, build, vet, lint, gate self-tests, ledger, full suite green | ✅ local logs + CI ×3 platforms |
| PR opened and merged with green CI incl. interop conformance gate | ✅ PR #43, all checks SUCCESS |

Accepted. The commit-owning scope was executed by the implementer (merge
`9bba77d` on origin/main); nothing remains uncommitted for this task's scope.
The delivery worktrees under `.temp/STORY-260825-*` and the primary checkout's
stale copies remain to be cleaned up by the orchestrator, per the landing
note.
