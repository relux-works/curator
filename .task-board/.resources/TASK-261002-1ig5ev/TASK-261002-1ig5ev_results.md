# TASK-261002-1ig5ev — restore-rc13-record-and-freeze-history

Ready for review. Changes remain uncommitted in the assigned Story worktree.

## Changes and decisions

- Restored `release/1.0.0-rc.13.json` to its exact tagged bytes. Published core digest remains `sha256:be11bb1e4c46f21fb5684d586f9c2a8b0d59f3b437bc7ea7aa5aa530fe4d47ca`.
- Removed the release-record writer from the generator. Minimal new bookkeeping is `conformance/candidate.json`: candidate status and independently computed core/source-suite manifest paths and digests. No existing candidate metadata file was found. It selects no next-version number and creates no rc.14 release record.
- Added `published_release_records()` and `validate_released_record_immutability()` to the real `validate.main()` path. All raw bytes, including whitespace, are frozen; missing records, failed Git reads, missing required tag/record evidence, malformed inventories, and symlink replacements fail.
- Coverage: tag guard inventories and checks **6/6** historical release records (rc.5, rc.6, rc.7, rc.8, rc.9, rc.13). Own-version tags are preferred. Rc.6 has no local version tag; its first tagged snapshot is v1.0.0-rc.7. The bound is the versioned records shipped by available local `v<semver>` tags, with v1.0.0-rc.13 evidence required; this is a byte comparison, not tag signature authentication.
- The real generator integration test preserves bytes and modification times for **7/7** records (six historical records plus an unknown-version sentinel), recreates candidate metadata from scratch, leaves missing published records absent, and creates no rc.14 record.
- Regeneration diff scopes in Makefile/CI/release workflows now cover generated candidate metadata rather than published records. `make regenerate-check` runs the tag-history guard first, so the intentional rc.13 restoration does not fail a comparison with the rewritten HEAD record.
- Existing main-entry vector mutation tests now update candidate pins. The release assurance-policy test uses the tagged rc.13 manifest and verifies a green baseline before mutating assurance fields; production release-publication checks remain intact. Both suites' digests in the rc.13 Go test come from that tag.
- Added one Unreleased CHANGELOG line and documented candidate bookkeeping in README. No LOGBOOK changes, as explicitly required by the task brief. The generic LOGBOOK checklist item was replaced with a task-specific board-outcomes/notes requirement under that instruction, and checked after both outcomes were attached. Findings are recorded here and on the board.

## Direct validation evidence

All commands below were run by this developer; no earlier attached evidence was substituted. Gates ran as standalone processes, without tee or pipe chains. Python dependency setup used `.temp/TASK-261002-1ig5ev-venv`, with the repository requirements installed successfully (exit 0).

| Command / scenario | Real exit code | Result |
| --- | --- | --- |
| `python3 tools/validate.py` before dependency setup | 1 | Failed because jsonschema was not installed; resolved with task-local virtual environment |
| Initial `make validate` in the virtual environment | 2 | Validator green; 667 Python tests ran, with four assurance-policy subtest failures caused by using today's manifest with frozen rc.13 metadata |
| Focused release assurance / release-history / candidate tests | 0 | 9 tests green after correcting the tagged manifest fixture |
| Standalone `python -B tools/validate.py` with an appended newline in rc.13 | 1 | Expected-red: published release record bytes differ from v1.0.0-rc.13; restored original bytes afterwards |
| Narrowed production tag guard checking only rc.13, running the all-record mutation test | 1 | Expected-red: five older-record diagnostic assertions fail; mutant restored from saved bytes |
| `go test -count=1 ./tools/...` during implementation | 0 | Uncached generator tools tests green; later Go test changes were exercised again by the current make validate run |
| `go build ./tools/...` after all code changes | 0 | Build green |
| CI-equivalent `gofmt -l tools` empty-output gate | 0 | Formatting green |
| `git diff --check` | 0 | Whitespace green |
| `make regenerate-check` after all code changes and mutant restoration | 0 | History guard, generator, and generated-output diff green |
| Current `make validate` after all code changes and mutant restoration | 0 | 73 schemas / 1294 vector files; 669 Python tests green in 416.931 seconds; Go tools tests green in 2.909 seconds |
| `git diff --quiet v1.0.0-rc.13 -- release/1.0.0-rc.13.json` after current make validate | 0 | Exact tagged bytes preserved |
| `test ! -e release/1.0.0-rc.14.json` | 0 | No rc.14 record |

The narrowed guard negative proves the new tag guard reaches each older record, through diagnostic assertions. The legacy hash guards still reject those five mutations; this experiment does **not** claim an acceptance bypass. The standalone rc.13 negative proves rejection at the real CLI entry. See attached `TASK-261002-1ig5ev_negative.log` for exact failure diagnostics and exit codes.

Both full Python runs were bounded below ten minutes; no gate was backgrounded or left running. Syspolicy was checked and reported running before gates. No required local validation was omitted.

## Snapshot

- Worktree HEAD: e41c561b3300a0a4d3425437fddc5a7048d7b11e
- Tagged rc.13 record SHA-256: 09b5bcfcb4e03df3bd10c6c328704fe3d387ac35beca3b832afb82801445f3dc
- Current core candidate manifest SHA-256: bd03456b92a7368d90ea74fe6953db10bc188588020683024a6a6b8735a40783
- Source-suite manifest SHA-256: 061ec05ddb1746b72168d157f81d2930047cd62371ffa46cf585844da61ee6be

## Handoff checklist correction

The first `task-board handoff TASK-261002-1ig5ev --role developer` exited 1 because the generic LOGBOOK item remained unchecked. The explicit task brief forbids LOGBOOK changes. Replaced that obsolete conditional item with the task-specific requirement to record findings in board outcomes and notes, and checked that requirement after attaching the evidence. No repository code changed after green gates.
