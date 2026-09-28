# Review verdict — TASK-260928-q5100t CR rev2: ACCEPTED

Base 17d88795, candidate tree d4880af0 (worktree `git write-tree` matches exactly), 9 paths, no stray files.

## Checks (reran myself, disposable copy /tmp/q5copy, venv with jsonschema)
- `python tools/validate.py` → exit 0 ("validated 64 schemas and 1170 vector files").
- `go run ./tools/generate-vectors` then `git diff --exit-code` → 0 (regenerate-check clean, manifest + rc.13 candidate pin are generator output).
- `unittest test_validate.GlobalLockPublicationVectorTests TakeoverClosedSetTextTests` → 37 OK; `test_release_gate`, `test_skillfile_sources_independence` → OK; `go test ./tools/...` → ok.
- Full test_validate suite (~20 min) NOT rerun in full; only the classes above.

## Mutants on real files
- Move §9.5 takeover sentence into §9.4 (whitespace-tolerant) → validate exit 1 "section 9.5 does not contain exactly one pinned takeover sentence". The adapted TakeoverClosedSetTextTests test still asserts exact-one match and the same error; carrier widen/drop/reorder tests unchanged.
- Vector conflict case profile_lock_after=old → exit 1 (digest); semantic gate also proven by unit mutants (lock rollback, partial surfaces, other state, overwritten conflict surface, wrong diagnostic surface, sync without takeover/old lock, missing install conflict case, carrier set + "global add").

## Spec
- environments.md:2462-2476: global add/install publish store entries + extended lock before in-place materialization; cite §9.2 steps 4–5; lock KEPT on environment_surface_unmanaged_conflict; diagnostic names surface; other failures ordinary rollback. §9.4 recovery sentence (2455-2460) and §9.5 closed set (2655) untouched → no new flag.
- manager.md:568-592 (§2.5 handled-disposition exception, journaled), 622-627 (§2.6 recovery keeps lock), 2716-2727 (§12.3). Consistent; grep of global add/install/rollback shows no contradiction.
- CHANGELOG Unreleased/Added entry present. release/1.0.0-rc.13.json change is only the generator-owned candidate manifest pin (tag v1.0.0-rc.13 bytes untouched; release_gate tests pass).

## Non-blocking observation
The new §9.4 ordering prose is not text-pinned by a validator (a MUST→MAY edit would pass validate.py); the normative behavior is carried by the exact-match vector gate. Stated bound, not a defect under the brief.
