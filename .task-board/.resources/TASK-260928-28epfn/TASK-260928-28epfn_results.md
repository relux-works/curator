# TASK-260928-28epfn Results

## Summary

Strengthened the rc.13 no-follow production-entry rows in `internal/envprofile/write_nofollow_conformance_test.go`. Parent-link cases now check the redirected destination entry itself before checking the returned diagnostic: an absent target remains absent, and a planted target symlink retains its original link text. The backup-parent case checks that the symlink is retained and its destination stays empty before checking the diagnostic. `internal/envprofile/nofollow.go` is unchanged; no production behavior changed.

The starting tree already had the relevant vector rows and both mutants failed against them. This change strengthens the evidence that no out-of-root target entry was published. Therefore the brief's requested “survive before” result was not reproducible on this starting tree; I report the baseline kill results below rather than claiming otherwise.

## Rules → production rows → pinned vectors

Normative source: curator-spec v1.0.0-rc.13, commit `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`, environments §8.3.1. It requires lstat-class inspection for components below the managed root, refuses unowned parent symlinks with `environment_write_would_follow_link`, and requires writes to replace entries without following links.

| Rule / mutant | Production-entry row | rc.13 vector(s) | Evidence |
| --- | --- | --- | --- |
| M1: parent walk changes `Lstat` to `Stat` in `managedPath` | `Resolve` via `TestEnvironmentWriteNofollowVectors` | `materialize-symlinked-parent-refused`; `takeover-symlinked-parent-authorized-still-refused` | Exact target assertion catches creation of `system-prompt.md` outside the managed root or replacement of the planted symlink. M1 failed with exit 1 on both rows. |
| M2: symlink refusal disabled in `managedPath` and `managedDirectory` | `Resolve` parent-link rows; `UseWithPolicy` via `runSwitchBackupParentCase` | Both parent-link vectors; `backup-symlinked-destination-refused` | Parent/backup links and outside state remain unchanged before the required diagnostic is checked. M2 failed with exit 1: parent rows returned “not a directory” and the backup row did not report `environment_write_would_follow_link`. |

Story/task-owned rows in `.github/ci/conformance-gaps.tsv`: **0 before → 0 after**. There were no owned passing gap rows to remove; the ledger was not changed.

## Mutant evidence

Each mutation was applied separately to `nofollow.go`; the source was restored after each run and verified unchanged. Both mutations also exited 1 against the starting-tree tests using the exact pinned vectors, so these are not reported as baseline survivors.

- M1 (`stateread.Lstat(current)` → `stateread.Stat(current)` at the parent walk): exit **1** before and after the assertion change. The updated rows identify an out-of-root target entry created/replaced.
- M2 (remove the symlink-refusal branches in `managedPath` and `managedDirectory`): exit **1** before and after the assertion change. The rows reject the weakened diagnostic; the backup-parent row also detects the refusal lost at the backup directory boundary.

## Validation

The vector root below was archived from commit `23435129` into `$TMPDIR/curator-spec-rc13`; it is not the local curator-spec checkout at `17d887950623db8cacaeae9d985841398c6dfa02`.

| Command | Exit | Result |
| --- | ---: | --- |
| `env CURATOR_CONFORMANCE_ROOT="$TMPDIR/curator-spec-rc13/conformance/v1" go test ./internal/envprofile -run '^TestEnvironmentWriteNofollowVectors$' -count=1` | 0 | Complete pinned vector family passed. |
| `go test ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$' -count=1` | 0 | Required manager-state read guard passed. |
| `go vet ./internal/envprofile` | 0 | Passed. |
| `golangci-lint run ./internal/envprofile` | 0 | 0 issues. |
| `go build -o "$TMPDIR/curator" ./cmd/curator` | 0 | CLI build passed; artifact stayed in `$TMPDIR`. |
| `gofmt -d internal/envprofile/write_nofollow_conformance_test.go` | 0 | No formatting diff. |
| `git diff --check` | 0 | Passed. |
| Starting-tree M1 vector-family mutant probe | 1 | Expected failure; mutant killed by existing parent-link rows. |
| Starting-tree M2 vector-family mutant probe | 1 | Expected failure; mutant killed by existing refusal assertions. |
| Updated M1 pinned vector-family mutant probe | 1 | Expected failure; out-of-root target entry detected. |
| Updated M2 pinned vector-family mutant probe | 1 | Expected failure; weakened refusal diagnostic detected. |
| `env CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envprofile -count=1` | 1 | Interrupted with Ctrl-C after over six minutes without output; this used the unpinned local checkout and is not counted as passing or pinned evidence. |

Earlier iterations superseded by the exact-pin run: a subtest-only filter exited 1 because the coverage harness requires complete family classification; one intermediate assertion-reordering run exited 1 because the test helper shadowed the operation error with a directory-read error, then was fixed and the complete family passed. Runs initially aimed at the local spec checkout were also superseded by the exact-pin archive above.

The Story workspace's recorded base is `main` at `213a53e5c701ef16b961f3398c7f2f0b82c98094`; the recorded protected authority reports equal advertised/fetched/selected OIDs, and a fresh `git ls-remote --symref origin HEAD refs/heads/main` returned the same OID.

Only `internal/envprofile/write_nofollow_conformance_test.go` is changed. No CHANGELOG or LOGBOOK file was edited.

## CHANGELOG entry (for release prep)

- Add regression coverage for managed environment writes blocked by symlinked parent and backup paths.
