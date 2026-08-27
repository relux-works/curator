# TASK-260728-wy3dsw manager-profile outcome

## Scope

The task worktree is:

`/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-wy3dsw/curator-spec-worktree`

It remains at pinned HEAD
`57c1f56846d221ecc55786bd3c2467ec32f11730`. The accepted predecessor state
from `TASK-260728-17sclp` was seeded byte-for-byte before documentation work.
Task-owned changes are limited to:

- `profiles/manager.md`
- `cli/curator.md`

An `rsync -anic --delete` comparison against the accepted predecessor,
excluding `.git`, `.temp`, caches, and build artifacts, exited 0 and reported
only those two files.

## Architecture-v6 mapping

Binding architecture:
`TASK-260720-1nvomm_external-build-repositories-architecture-v6.md`,
SHA-256
`2abae77d80eba6789f9911db7e9722595b4f21ba47391ca9eafd0064af03d67e`.

| Architecture-v6 sections | Resulting documentation |
|---|---|
| 1-5: version boundary, ownership, canonical identity, lock/tag, declared/effective state | Manager profile 11 introduction and 11.2; CLI rc.5 qualification |
| 6.1-6.4: trusted Git, clean environment, init/fetch/ref flows, SSH wrapper | Manager profile 11.1-11.3 with exact argv/environment and exact-tag-only path |
| 6.5: local ordinary files-ref admission, config, refs, pack/index | Manager profile 11.4 |
| 6.6: common cat-file grammar, object recomputation, commit/tag/tree/blob graph, LFS | Manager profile 11.5 |
| 7-8: snapshot/audit order, receipt-v2 keys, mixed receipt planning | Manager profile 11.6 |
| 9: marker-v3, protected identities, status, repair, deduplication, GC | Manager profile 11.6, 11.7, and 11.9 |
| 10: trusted go-v1 session reuse, publication, shims, PATH, rollback | Manager profile 11.8 |
| 11-13: signing/credential boundary, closed behavior, threat controls | Manager profile 11.1, 11.8, and 11.10 |
| 17: exact examples and behavioral checks | Curator CLI external-repository lifecycle examples and focused presence gate |

The profile explicitly freezes schemas 1-6, `go-v1`, receipt-v1, and
marker-v1/v2 meanings. The CLI guide identifies the added command surface as
the rc.5 contract and states that an rc.4-only Curator binary does not yet
accept schema-7 objects or expose `repair`.

## Stable behavior

The manager profile keeps these outcomes distinct:

- `build_repository_source_unavailable`
- `build_repository_ref_moved`
- `build_repository_unverified_offline`
- `build_repository_incomplete_source`
- `build_repository_git_object_semantics_invalid`
- `build_repository_git_lfs_unsupported`
- `build_repository_local_gitfile_unsupported`
- `build_repository_local_bare_unsupported`
- `build_repository_local_linked_worktree_unsupported`
- `build_repository_local_layout_unsafe`
- `build_repository_local_format_unsupported`
- `build_repository_local_object_format_unsupported`

It also rejects package/repository control of Git, SSH, HTTPS, helpers, hooks,
filters, alternates, replacement/graft/promisor/lazy reads, credentials,
compiler argv/environment/output, PATH entries, signing, and diagnostic
rendering.

## Validation evidence

- Initial host `python3 tools/validate.py`: exit 1. Expected environment
  failure because host Python lacked `jsonschema`; not counted as passing.
- Task-local venv creation from `requirements-dev.txt`: exit 0.
- Pinned-environment `python3 tools/validate.py`: exit 0; 42 schemas and 400
  vector files validated.
- Pinned-environment Python unit tests: exit 0; 15 tests passed.
- `go test ./tools/...`: exit 0.
- Post-edit `make validate` with task venv prepended to the normal toolchain
  PATH: exit 0; 42 schemas, 400 vectors, 15 Python tests, and Go tool tests
  passed.
- Focused Git/offline/audit/cache/transaction/status/repair/GC presence gate:
  first invocation exit 1 because one assertion crossed Markdown line
  wrapping; corrected literal gate exit 0 and is the counted result.
- `git diff --check`: exit 0.
- `git diff --cached --exit-code`: exit 0.
- `git rev-parse HEAD`: exit 0 and returned the pinned HEAD above.

Document hashes at handoff:

- `profiles/manager.md`:
  `44db1200a5b22f63785acf2ea304c20b0c26ec5a981b65818dbd097eb56b5839`
- `cli/curator.md`:
  `6f2525612ba3b7e62cf94e8b859798d72e7e332b7d0f4a2406fa1329b69c059d`

No file was staged or committed.
