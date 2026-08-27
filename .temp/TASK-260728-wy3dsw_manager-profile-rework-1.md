# TASK-260728-wy3dsw manager-profile rework 1

## Scope

Reworked the documentation in the pinned isolated worktree:

`/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-wy3dsw/curator-spec-worktree`

HEAD remains
`57c1f56846d221ecc55786bd3c2467ec32f11730`.
No file was staged or committed. An accepted-predecessor `rsync -anic
--delete` comparison exited 0 and reported only:

- `profiles/manager.md`
- `cli/curator.md`

No product code, schema, vector, generator, test, decision, protocol-core, or
release file was changed by this rework.

## Reviewer findings resolved

1. **Local config and files-ref admission**
   - Manager section 11.4 now defines record endings, token bytes, section and
     assignment productions, quoted subsection/value escapes, comments,
     boolean spellings, required `core.bare=false`, format states, ordinary
     versus security-relevant duplicate behavior, exact HEAD/loose-ref
     terminators, packed-ref traits/records, loose-over-packed precedence, and
     unconditional `refs/replace/*` rejection.
2. **Commit/tag grammar**
   - Manager section 11.5 now defines the header key/value/continuation byte
     grammar, opaque message boundary, structural-header ordering and reserved
     names, actual target-type checks, tag cycle/depth limits, and the
     outermost annotated-tag name equality requirement for an exact tagged
     declaration.
3. **Stable typed diagnostics**
   - Manager section 11.10 now defines 33 unique stable
     code/phase/state/severity mappings across schema, descriptor, identity,
     transport/credentials, source/object proof, audit, protected cache,
     receipts, markers, artifacts, currentness, compiler, transaction, and
     signer policy.
   - Five specific package-control codes cover argv, environment, output/PATH,
     credentials/host policy, and signing. Their precedence over generic
     schema/descriptor failures is normative.
4. **Disjoint offline semantics**
   - Only syntax-only `skill check` may return warning state
     `unverified-offline`.
   - Real or dry-run install/upgrade, update, repair, and coverage-claiming
     audit fail before mutation with state `blocked`, severity `error`, code
     `build_repository_source_unavailable`, and CLI exit code 1 when exact
     source is unavailable.
   - `unverified-offline` was removed from the install-family dry-run state
     list in both owned documents.
5. **SSH wrapper tuple**
   - Manager section 11.3 now requires byte-exact
     `argv[0]=<absolute-manager-binary-wrapper>`, host, remote command, and
     `argc=3`; relative, alternate, or aliased wrapper names reject.

## Validation evidence

All commands were run directly without `tee`.

- `make validate` under the pinned task venv/toolchain PATH: exit 0 after the
  rework and exit 0 again after the last ref-grammar correction. Each run
  validated 42 schemas, 400 vector files, 15 Python tests, and Go tool tests.
- `python3 tools/validate.py`: exit 0 on each direct run; 42 schemas and 400
  vector files validated, including local Markdown links.
- `python3 -B -m unittest discover -s tools -p 'test_*.py'`: exit 0; 15 tests
  passed.
- `go test ./tools/...`: exit 0.
- Focused five-finding presence/offline-state gate: exit 0.
- Diagnostic-matrix parser gate: exit 0 on both runs; 33 unique mappings and
  all five package-control rejection codes present.
- Git/offline/audit/cache/transaction/status/repair/GC lifecycle gate: exit 0.
- Exact Git 2.50.1 private-init and full-OID
  `cat-file --batch=%(objectname) %(objecttype) %(objectsize)` smoke under the
  documented clean environment: exit 0. It verified files refs, canonical
  batch framing, and the exact generated HEAD terminator.
- Two SSH fetch probes: exit 128 each, expected because the task-only observer
  wrapper deliberately exits 73 instead of contacting a remote. These are
  expected-red fetch attempts, not passing fetch gates.
- Initial SSH observer-log assertion: exit 1 because the first task-only probe
  logged Python's `-c` as argv[0] instead of the wrapper's shell `$0`. The
  probe was corrected; the repeated assertion exited 0 and proved the tuple:
  absolute wrapper argv[0], `git@example.test`, exact
  `git-upload-pack '/example/repo.git'`, and argc 3. This was a scratch-probe
  defect, not a profile or Git defect.
- `rsync -anic --delete` accepted-predecessor comparison: exit 0 on both
  final checks; only the two owned documentation files were reported.
- `git diff --check`: exit 0 on each run.
- `git diff --cached --exit-code`: exit 0 on each run.
- `git rev-parse HEAD`: exit 0 and returned the pinned HEAD above.

Final document hashes:

- `profiles/manager.md`:
  `4db786194415ac86df0b912b15201bcb24109a50ceb4d8dd0234740ce04b32ba`
- `cli/curator.md`:
  `6160f08ad4b433aacb772085949c8bb1f3361eddd6ca5cc52179bd5dbb7c1ba6`

## Handoff

The bounded documentation rework addresses all five changes requested by the
independent reviewer and is ready for a fresh independent review. This resource
does not self-accept the task.
