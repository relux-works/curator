# TASK-260728-17sclp review rework evidence

Worktree: `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-17sclp/curator-spec-worktree`

Pinned HEAD: `57c1f56846d221ecc55786bd3c2467ec32f11730`

## Finding-to-fix mapping

1. Transport, canonical identity, tag, and effective ref grammar:
   - Tightened the shared `repositoryGit` and `networkSourceIdentity` definitions.
   - Added deterministic semantic validation for exact HTTPS/SSH/SCP forms, Unicode HTTPS versus ASCII SSH paths, empty/dot components, forbidden bytes, canonical lowercase host/path identities, one removed lowercase `.git`, 1..255 UTF-8 byte refs, and effective SHA-1/SHA-256 revision width.
   - Added named generated cases for Unicode HTTPS, SSH URI/SCP, dot components, SSH metacharacters/non-ASCII, canonical host/path and `.git` spelling, valid 255-byte and invalid 256/300-byte refs, and SHA-1/SHA-256 structured revision boundaries.

2. Legacy schema 1 through 6 reserved surface:
   - Preserved every accepted legacy schema and existing case byte-for-byte.
   - Extended semantic version selection to reject top-level `build_repositories`, `repository`, `target`, and `driver: go-repository-v1`, plus command-level `repository`, `target`, and `driver: go-repository-v1`.
   - Generated all seven reserved-surface invalid cases for both manifest filenames at every version 1 through 6. Go inventory tests require all 84 cases.

3. Optional Skillfile.dev schema-2 extension:
   - Removed `build_repository_substitutions` from the required list while retaining the closed optional map.
   - Added ordinary-only, empty-map, populated-map, tag, branch, and revision valid cases.
   - Added package-ownership rejection cases for target, driver, command, output, credentials, and argv.

4. Missing semantic branch vectors:
   - Receipt v2 now has generated containment, declared/effective mismatch, canonical identity, effective commit width, effective structured-ref width, and substitution identity cases.
   - Marker v3 now has local-only, external-only, mixed build-source conditions, receipt-version branches, and declared/effective mismatch cases.
   - Claim v3 now has duplicate-driver and driver-platform-outside-top-level cases.
   - Go inventory tests require every named branch.

## Final green gates

Every command below ran directly without `tee`.

- `python -B tools/validate.py`: exit 0; 42 schemas and 389 vector files.
- Python unittest discovery: exit 0; 15 tests.
- `go test -count=1 ./tools/...`: exit 0.
- `PATH=<task-venv>/bin:$PATH make validate`: exit 0.
- `go vet ./tools/...`: exit 0.
- `gofmt -l tools/generate-vectors`: exit 0 with no output.
- `go build -o <task-temp>/generate-vectors-rework ./tools/generate-vectors`: exit 0.
- Two-pass regeneration plus `diff -rq` after each pass: exit 0; evidence directory `rework-final-idempotence.V8l2iS`.
- Accepted rc.4 legacy schema/case comparison: exit 0.
- Accepted schema-7 prose comparison for SECURITY, manager profile, Core, and Decisions 0004/0005: exit 0.
- `git diff --check`: exit 0.
- `git diff --cached --quiet`: exit 0.
- Pinned-HEAD assertion: exit 0.
- Repository build-artifact and Python-cache absence assertion: exit 0.

## Expected-red and harness evidence

- The first focused Go test after adding the new generated inventory exited 1 because two older exact-inventory assertions still expected no new legacy guards. The assertions were updated to require the guards; the rerun exited 0.
- The first schema validation after adding receipt containment coverage exited 1 because the invalid fixture used `build_root: "."`, which correctly contains every repository-relative source. The case was corrected to a non-root build root with an outside source; the rerun exited 0.
- The first accepted-baseline comparison exited 1 because it used a guessed, absent worktree path. The accepted TASK-260720-37ei85 outcome identified the authoritative path under `curator/.temp`; the corrected byte comparison exited 0.

No files were staged or committed. No manager implementation, exact Git execution profile, shared runtime vectors, or release promotion was added.
