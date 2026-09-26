# TASK-260924-1fnikm — review of relux-works/curator-spec#96 (Decision 0022)

Reviewed head 5746367 on base 7b06bb3 (origin/main). I worked in a disposable clone ($TMPDIR/pr96) and pushed nothing to the PR branch.

## Verdict: APPROVE (spec side), with non-blocking notes N1–N5

Binding criteria 1–5 and 7 pass. Criterion 6 is curator-owned: it finds one conformance gap (object-format check) and 12 undriven rows. Neither needs a change to this PR.

## 1. Separate namespace, frozen v1 — PASS
- `git diff --stat origin/main HEAD -- 'schemas/v1/*.json'` produces no output. `conformance/v1` is untouched, and `make regenerate-check` exits 0.
- The only schemas/v1 change is prose in `schemas/v1/README.md:239-245`.
- The corpus stays in `conformance/skillfile-sources-v1` and `schemas/skillfile-sources-v1`, and is not merged into `conformance/v1/manifest.json` (`conformance/README.md` §7).

## 2. rc.10 independence — PASS
Script `/tmp/rc10check.py`: it collects every `$ref` into `../v1/` from the `schemas/skillfile-sources-v1/*.json` files, resolves each JSON pointer at HEAD and at tag v1.0.0-rc.10, and compares them canonically, including one level of inner local `$ref`s. Output:
`refs=38 bad=0`, exit 0. All 38 are OK: 17 `common.schema.json#/$defs/*`, 18 `install-marker-v4#/properties/*` and 3 `skillfile-v1#/properties/*`.
A stronger check shows the three referenced files are byte-identical between rc.10 and HEAD: `git diff --quiet v1.0.0-rc.10 HEAD -- schemas/v1/{common,install-marker-v4,skillfile-v1}.schema.json` succeeds for all three. That makes the check transitively complete.
Prose: see N1. The only post-rc.10 reference is environments §9.4, and `protocol/environments.md` does not exist at rc.10. The clause is conditional, though: "when the manager implements that capability"; otherwise the global scope stays on schema 1 and schema 2 is rejected. An rc.10-only client therefore gets a complete, self-contained rule.

## 3. Rename identity — PASS
Name-status summary: 123 files are R100. All 9 JSON schemas are R<100 only because of the `$id` path. A canonical-JSON comparison with `draft-sources-v1`→`skillfile-sources-v1` substituted gives IDENTICAL for all 9. Complete list of non-rename changes:
- `conformance/skillfile-sources-v1/README.md`: wording, and the recipe paths changed to `skillfile-sources-v1`.
- `schemas/skillfile-sources-v1/README.md`: wording.
- `semantic-cases.json`, compared case by case:
  - Removed: `missing-snapshot`.
  - Added: `path-missing-snapshot-identical-bytes`, `path-missing-snapshot-drifted-bytes`, `git-missing-snapshot-fetches-locked-commit`, `git-moved-tag-replays-locked-commit`, `git-moved-tag-replays-locked-commit-through-mirror`, `missing-snapshot-unreachable-source`, `global-schema2-without-profile-lock-refused`, `project-schema2-without-profile-lock-accepted`.
  - Changed: `v2-declared-mirror`. Verification is now scoped to resolution, and the fields `operation`, `declared_tag` and `resolved_commit` were added.
- Prose files: `protocol/{skillfile-sources,repository-transport,core}.md`, `profiles/manager.md`, `docs/skillfile-sources.md`, README, CHANGELOG, COMPATIBILITY, `conformance/README.md`, `schemas/v1/README.md`, and the new `decisions/0022-*.md`.
- #90 (2cp9w9, 7b06bb3) is present and unchanged. Its four semantic ids are not in the changed set, and its schema cases are in the R100 set.
- v9 manifest material is absent: `git ls-files | grep -i v9` finds nothing. Decision 0022 explicitly excludes v9.

## 4. Global scope — PASS
The rule is in `protocol/skillfile-sources.md:17-22` and Decision 0022 item 3. The positive vector is `project-schema2-without-profile-lock-accepted`; the negative is `global-schema2-without-profile-lock-refused` (upgrade-error).

## 5. §3 lock replay — PASS
- `protocol/skillfile-sources.md:175-187`: exact commit by object ID; object format and commit verified; path members read from current bytes; accepted only on identity + content_sha256, otherwise `source_snapshot_changed`; `unavailable` only when the source is unreachable; the lock is never rewritten; no ref is resolved.
- There are 6 replay vectors (≥5 required).
- Tag consistency: `repository-transport.md:83-96` scopes exact-tag verification to resolution and refresh/upgrade. `:211-219` does the same for mirrors, the failure table at `:70` agrees, and `v2-declared-mirror` was rescoped to match. `git grep` finds no remaining sentence that requires tag verification when installing from a lock.
- The core §6.3 tagged flow (`core.md:1265-1285`) belongs to the external build-repository lane, not to schema-2 member replay. That is out of scope (N5).

## 6. Curator conformance (curator STORY-260924-3eywt2 @66bc92aa + worktree diff, not landed; origin/main ab34556e)
- `internal/install/draftsources.go:489/520` (`fetchLockedCommitFromDeclaredSource`) fetches by object ID through `gitops.FetchCommitFromURLIsolated`; the in-place path at `:422` does the same via `gitops.go:376-389`, with `fetch --no-tags <oid>`.
- Repository identity is checked at `:507-508`. `content_sha256` and the package are checked in `verifyLockedGitMember` at `:453-484`. There is no tag check.
- **GAP C1 — object format not verified.** `gitops.validateLockedCommit` (`gitops.go:431-441`) only checks for 40 or 64 hex characters, and nothing reads `member.Package.Commit.ObjectFormat` to compare it with the fetched repository's format. The new text requires "verify its object format and commit". Owner: curator STORY-260924-3eywt2 (or a follow-up leaf).
- Path replay: `replayLocalDraftSnapshot` at `:342-401` exists and maps both `changed` and `unavailable` correctly.
- Crossconformance suite (`internal/crossconformance`) run against the PR corpus in a disposable curator clone:
  - It hard-codes `testdata/draft-sources-v1/...` (`draftsources_corpus_test.go:33,37`) with counts 115/94. I added a shim: corpus copied under the old path, MANIFEST regenerated, counts set to 121/105.
  - Pin and SnapshotVectors pass. CorpusCounts and SchemaCases fail without the counts bump (exit 1). SemanticCoverage exits 1.
  - SemanticCases, split into 6 batches: 105 cases run, 93 pass, 12 fail. The unsplit run timed out at 8m50s. Every failure is "has no production-entry row"; none is a behavioural failure.

| Owner | Failing cases |
|---|---|
| TASK-260924-20o9dk (#90 rows) | attestation-evidence-revoked-identity-commit-advisory, v2-refresh-current-endpoint-existing-checkout, v2-scp-alias-port-refused, v2-ssh-uri-alias-port |
| NEW curator row: replay drivers | path-missing-snapshot-identical-bytes, path-missing-snapshot-drifted-bytes, git-missing-snapshot-fetches-locked-commit, git-moved-tag-replays-locked-commit, git-moved-tag-replays-locked-commit-through-mirror, missing-snapshot-unreachable-source |
| NEW curator row: global/project drivers | global-schema2-without-profile-lock-refused, project-schema2-without-profile-lock-accepted |
| NEW curator row: pin move (handoff step 4) | hard-coded draft-sources-v1 path, MANIFEST, counts 115/94→121/105, stale `missing-snapshot` driver id |

- **C2:** `v2-declared-mirror` passes vacuously. `driveV2DeclaredMirror` (`draftsources_semantic_v2_test.go:66`, about lines 185-208) never compares against `c.Expected`, so the pass says nothing about the rescoped expectation. It needs re-review in the pin-move row.

## 7. Recipes (run in the clone; exit codes are real)
- The corpus README command (in a venv with jsonschema): exit 0. It reports 121/121 schema cases, 96/96 negatives, 8/8 wire schemas, 18/18 refusal mutants, 3/3 snapshot vectors, and "Manager semantic execution: 0 cases".
- `python3 tools/validate.py`: exit 0 (64 schemas, 1169 vector files).
- `make regenerate-check`: exit 0.
- `go test ./tools/...`: exit 0.
- **Not run:** `python3 -m unittest discover -s tools` (about 20 minutes, which exceeds the bounded call budget). The PR does not touch tools/.

## Non-blocking notes
- N1 `protocol/skillfile-sources.md:19`, `decisions/0022-accept-skillfile-sources-v1.md:22,45`: the link targets a clause that does not exist at rc.10. It is acceptable because it is conditional; consider marking it informative for rc.10-pinned clients.
- N2 `CHANGELOG.md:466`: "Machine-global scope follows environments §9.4 profile locks" drops the "when implemented; otherwise schema 1" qualifier.
- N3 `CHANGELOG.md:466-469`: the wording attaches the identity + content-hash gate only to path replay. Per §3, Git/repository replay is gated the same way.
- N4 `conformance/skillfile-sources-v1/semantic-cases.json:251`: `missing-snapshot-unreachable-source` has no `"operation": "install-with-existing-lock"`, unlike its sibling replay vectors.
- N5 `schemas/v1/README.md:243`: "The Curator orchestrator verified…" is a process attribution inside normative-adjacent prose. The rc.10 check above now reproduces it independently.
