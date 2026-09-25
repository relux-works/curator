# TASK-260924-2v4v2m results

## Implementation

- `validBuildState` now checks each v5 local `go-v1` build record against its
  closed raw seven-field shape before decoded values can treat external-only
  `repository`, `substituted`, or `substitution` members set to `null` as
  omitted zero values.
- The v5 external build reader validates a present `declared_tag` with
  `identity.DraftSourceRefName`, the existing draft-sources Git ref-name
  validator. Empty and malformed tag names are refused; valid tags remain
  readable.
- The new guards run only for marker schema v5. The v1-v4 reader paths were not
  changed. `CHANGELOG.md` records the behavior change.

## Production-entry coverage

`TestMarkerV5LocalBuildClosedShape` and
`TestMarkerV5DeclaredTagRequiresValidGitRef` write a valid marker, rewrite its
record bytes to represent a foreign writer, and assert the result of
`marker.Read` directly.

- Local v5 `go-v1`: `repository:null`, `substituted:null`, and
  `substitution:null` are each refused.
- External v5 `go-repository-v1`: present empty and malformed (`release..next`)
  `declared_tag` values are refused. A valid `release/v1.4.0` tag remains
  readable.
- Existing controls continue to cover valid local and external records,
  including external records without a tag and with substitution. The complete
  marker package suite covers v1-v4 controls.

## Validation

All commands below were run directly, without a pipeline.

| Command | Exit | Result |
|---|---:|---|
| `gofmt -w internal/marker/marker.go internal/marker/marker_v5_builds_test.go` | 0 | Files formatted |
| `go test ./internal/marker` | 0 | Package passed (`9.356s`) |
| `golangci-lint run ./internal/marker/...` | 0 | `0 issues` |
| `go vet ./internal/marker` | 0 | Passed |
| `go build ./cmd/curator` | 0 | CLI package compiled; generated local binary was removed |
| `git diff --check` | 0 | Passed |

## Narrowing mutants

Mutants were applied only in a disposable source copy under `.temp/` and then
removed. Each test run below is expected to fail with exit 1 because the
matching `marker.Read` row admits its malformed record under the mutant.

| Mutant | Matching command | Exit | Result |
|---|---|---:|---|
| Allow only `repository` through the local raw-shape check | `go test ./internal/marker -run 'TestMarkerV5LocalBuildClosedShape/repository-null' -count=1` | 1 | Killed: `Read` admitted `repository:null` |
| Allow only `substituted` through the local raw-shape check | `go test ./internal/marker -run 'TestMarkerV5LocalBuildClosedShape/substituted-null' -count=1` | 1 | Killed: `Read` admitted `substituted:null` |
| Allow only `substitution` through the local raw-shape check | `go test ./internal/marker -run 'TestMarkerV5LocalBuildClosedShape/substitution-null' -count=1` | 1 | Killed: `Read` admitted `substitution:null` |
| Exempt empty tags from ref validation | `go test ./internal/marker -run 'TestMarkerV5DeclaredTagRequiresValidGitRef/empty' -count=1` | 1 | Killed: `Read` admitted `declared_tag:""` |
| Check only non-emptiness and skip Git ref grammar | `go test ./internal/marker -run 'TestMarkerV5DeclaredTagRequiresValidGitRef/malformed' -count=1` | 1 | Killed: `Read` admitted `declared_tag:"release..next"` |

The first malformed-tag mutant attempt also exited 1, but failed to compile
because the disposable edit left an unused import. That attempt is not counted
as a killed mutant; the corrected mutant was rerun and the test failed at the
assertion as shown above.

Hosted CI is left to the configured handoff landing gate; no full landing suite
was run manually.

---

## Revision 2 (carry-forward republish)

Trunk moved to `a48f584c`; the accepted revision-1 delta was carried by
three-way converge and is republished here with content unchanged (except the
CHANGELOG policy revert below). Per-path verification against
`TASK-260924-2v4v2m_change-request_rev1.patch`:

- `internal/marker/marker_v5_builds_test.go` — trunk did NOT touch: trunk HEAD
  blob is byte-equal to the rev1 base blob `4c2e2eb7` (`cmp` clean), and the
  worktree append (lines 347-399, +53/-0) is byte-equal in order to rev1's
  added block (`cmp` clean). Byte-identical to revision 1.
- `internal/marker/marker.go` — INTERSECTING path: trunk changed one line
  (SchemaV5 SkillSchemaVersion bound `> 8` to `> 9`). Both sides are present
  in the worktree: trunk's `> 9` bound at line 338 plus every rev1 hunk
  (`identity` import, `v5LocalBuildShape`, `declared_tag` ref-name check,
  `validV5LocalBuildShape`, `validBuildState` go-v1/external dispatch; +39/-3
  added/removed line multisets identical to rev1). No conflict markers.
- `CHANGELOG.md` — INTERSECTING path: trunk added entries (hunk anchor
  193 -> 298). Converge kept both sides with no markers (added lines identical
  to rev1, +4/-0); per orchestrator CHANGELOG policy the hunk was then reverted
  entirely — `CHANGELOG.md` is now byte-equal to trunk (`cmp` against
  `git show HEAD:CHANGELOG.md` clean). Entry text preserved verbatim in
  "## CHANGELOG entry (for release prep)" below for the release-prep leaf.
- Stray paths: no root `TASK-*.md` / `BUG-*.md`, no `test/`, no `ledger/`.
  No conflict markers in any touched file (`grep` found no matches).

Focused bounded validation (each run directly, no pipeline):

| Command | Exit | Result |
|---|---:|---|
| `go test ./internal/marker/... -count=1` | 0 | `ok ... 10.184s` |
| `go vet ./internal/marker/...` | 0 | clean |
| `gofmt -l internal/marker/` | 0 | no output (clean) |
| `git diff --check` | 0 | clean |

Mutant evidence is unchanged from revision 1 (accepted verdict
`TASK-260924-2v4v2m_review-verdict-rev1.md`): five narrowing mutants killed
at the `marker.Read` production entry point; this republish changes no
product or test bytes except the CHANGELOG revert, so that coverage stands.
Worktree diff is now exactly the two code paths (`marker.go`,
`marker_v5_builds_test.go`).

## CHANGELOG entry (for release prep)

```
- Draft marker-v5 build records now validate the closed raw shape of local
  `go-v1` builds, refusing external-only fields even when their JSON value is
  `null`. External `declared_tag` values must be non-empty valid Git ref names;
  marker schemas 1 through 4 keep their existing rules.
```

---

## Revision 2 re-verification (carry-forward republish, independent run)

Re-verified the carried delta against
`TASK-260924-2v4v2m_change-request_rev1.patch` on trunk `a48f584c`,
changing no source bytes:

- `internal/marker/marker_v5_builds_test.go` — trunk did NOT touch it
  (HEAD blob `4c2e2eb7` equals the rev1 base); the rev1 hunk re-applied to
  the HEAD blob in /tmp is `cmp`-clean against the worktree file.
  Byte-identical to revision 1.
- `internal/marker/marker.go` — INTERSECTING path: trunk's one-line change
  (`SkillSchemaVersion` bound `> 8` to `> 9`, worktree line 338) is present
  alongside all 39 rev1 added lines; all 3 rev1 removed lines are absent;
  distinctive blocks are single (`DraftSourceRefName` x1). No conflict
  markers in any touched file.
- `CHANGELOG.md` — byte-equal to trunk (`cmp` against
  `git show HEAD:CHANGELOG.md` clean), per the orchestrator CHANGELOG
  policy; entry text stays verbatim in "## CHANGELOG entry (for release
  prep)" above. No stray root `TASK-*`/`BUG-*`, `test/`, or `ledger/`
  paths.
- Focused bounded validation, run directly:
  `go test ./internal/marker/... -count=1` exited 0
  (`ok github.com/relux-works/curator/internal/marker 59.995s`).

Mutant evidence stands from accepted revision 1 (five narrowing mutants
killed at the `marker.Read` production entry point); this run changed no
product or test bytes. Worktree diff remains exactly the two code paths
(`marker.go`, `marker_v5_builds_test.go`).
