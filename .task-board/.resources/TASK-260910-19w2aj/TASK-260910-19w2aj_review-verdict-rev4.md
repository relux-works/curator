# TASK-260910-19w2aj — revision 4 independent review

Verdict: CHANGES_REQUESTED. Route to `to-dev`; ordinary implementation rework, no human decision required.

Reviewed CR-TASK-260910-19w2aj-4, candidate tree `78ce30d6eac6178b0a69c99757308c5813a15187`, base `62ea2d2ced3fac7c3c5f7a2ff4e887e0f92f4540`. Independently compared all 5,539 candidate blobs against worktree bytes (including symlink targets): zero mismatches. No product code changed. `task-board spawn goal` reports this run is not goal-bound.

## Findings

1. **P1 — production resolve omits the configured dependency repository root.** `cmd/curator/project_resolve.go:113-121` constructs DraftResolveConfig without SkillsRoot. `internal/closure/resolve.go:150` passes its empty value to closure.Options; `internal/closure/closure.go:359` consequently looks up/clones `provider` relative to process CWD instead of the configured skills root. A local selected package requiring a valid tagged provider already present at `<configured skills root>/provider` fails `project resolve` with exit 1 and attempts a clone to relative `provider`. The independent fixture denies that unnecessary acquisition through fake Git. Exposing the same existing provider via a CWD symlink, without changing config, Skillfile, source bytes or code, makes resolve succeed with two locked members (exit 0). Thus resolution depends on incidental CWD and may create repositories outside the managed source root. Supply the configured root and necessary existing closure options in the CLI path. Add a production CLI regression for transitive dependencies with CWD outside the configured source root, and verify frozen consumption afterward. Helper test TestDraftTransitiveGitDependencyFrozen supplies SkillsRoot itself and cannot catch this omission.

2. **P2 — frozen identity recovery chooses a repository's first alias rather than the locked selection's alias.** `internal/install/draftsources.go:122-137` collapses aliases by canonical repository identity, then lines 147-153 recover bindings and declared ref from that first alias. Valid manifest: aliases `a` -> repo/tag v1 and `s` -> same repo/tag v2, with the sole selection `from: s`. Both tags point to the same commit, so all byte checks pass. CLI resolve, dry-run install and real install return 0, but the installed `.csk-install.json` records `ref: v1`; the actual declaration is v2. The dry-run also reports v1. Use the lock member's selection index to recover the exact manifest entry and its source alias; preserve separate handling for unchanged legacy entries. Add production entry coverage for individual and collection selections sharing a repository across aliases, including different declared refs.

These are contract regressions: skillfile-sources §1 retains configured-root semantics and admits alias-specific ref declarations; §3 binds each root member to its declaring selection index. Repository identity is not selection identity.

## Independent verification and bounds

All commands ran in zsh; fixture pipelines used `set -o pipefail`. Observed:

- `go build -o .temp/review-19w2aj-rev4/curator ./cmd/curator`: exit 0.
- `go test -p 1 ./cmd/curator -run 'TestProjectResolve' -count=1 -timeout=90s`: exit 0, 14.894s. Includes rev3 rework's runtime-tamper dry/real refusals, missing member, real install and pinned reinstall.
- `go test -p 1 ./internal/closure ./internal/install -run 'Test(ResolveDraft|OpenDraftFrozen|Refresh|Draft|LegacyInstall)' -count=1 -timeout=90s`: exit 0; closure 8.454s, install 7.235s.
- Independent alias CLI reproduction: resolve 0, dry-run 0, real install 0, declared v2 / recorded v1. Script asserts this observed mismatch and exits 0.
- Independent transitive CLI reproduction: resolve 1 with configured provider present; CWD-provider control resolve 0. Script asserts both exit codes and exits 0.
- Accepted attached hosted evidence, not rerun: rev4-validation.log records GitHub run 35191907971 success, remote-gate exit 0. Verified gate commit `56ea84e08ba6beb9f6df788211bb07e3552aa6cc` has exactly the candidate tree. Hosted Windows/Ubuntu/macOS and lint are reported green there; rose-air is skipped, not verified.
- Producer's narrowing mutant and ref-enrichment mutant evidence was read, not independently replayed. No independent mutants needed to establish these two actual production failures. Attack coverage is 2/2 deliberately probed new cases reproducing defects, not a claim of exhaustive clause coverage.
- A slow initial per-file git-diff scan was terminated (143) and replaced by the successful complete blob comparison; it is not counted as validation.
- Local-snapshot real marker support is explicitly deferred in producer evidence to STORY-260910-1s75e1; no passing local real-install/launch claim is made here. No full local suite or cross-platform run performed.

## Handoff

Preserve the rev4 full-inventory authentication and real Git marker fixes. Fix the two concrete findings above; add CLI tests that cannot supply omitted production configuration; rerun focused checks and hand off a new revision for independent review. Evidence scripts/logs are attached as TASK-260910-19w2aj_review-rev4-{alias,transitive}-repro.{py,log}. Scripts are review artifacts, not proposed production files.
