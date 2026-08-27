# TASK-260728-17sclp review rework 2 evidence

Worktree: `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-17sclp/curator-spec-worktree`

Pinned HEAD: `57c1f56846d221ecc55786bd3c2467ec32f11730`

## Finding closure

The cycle-2 reviewer found that marker-v3 generation did not exercise every live `buildRecordV2` and `validate_effective_source` integration branch. Rework 2 changes only the generator, its Go inventory assertion, and the generated conformance corpus/index/manifest.

New valid marker-v3 cases:

- `valid-empty-builds.json`: empty `builds`, empty `commands` and `build_roots`, and no top-level `build_source`.
- `valid-network-substitution-tag.json`: network-git substitution with safe tag `v1.4.0`.
- `valid-network-substitution-branch.json`: network-git substitution with safe branch `release/v2`.
- `valid-network-sha1-revision.json`: SHA-1 effective source with an exact 40-hex revision.
- `valid-network-sha256-revision.json`: SHA-256 effective source with an exact 64-hex revision.
- `valid-sha256-external.json`: unsubstituted SHA-256 external repository record.
- `valid-untagged-external.json`: unsubstituted external repository record without `declared_tag`.

New invalid marker-v3 cases:

- `invalid-marker-local-identity-kind-mismatch.json`: local-path substitution paired with a network-git effective identity.
- `invalid-marker-network-identity-kind-mismatch.json`: network-git substitution paired with an operator-local-git effective identity.
- `invalid-marker-sha1-effective-revision-width.json`: SHA-1 effective source paired with a 64-hex structured revision.
- `invalid-marker-sha256-effective-revision-width.json`: SHA-256 effective source paired with a 40-hex structured revision.

`TestGeneratedSchemaV7CasesCoverEveryWireBranch` now requires all eleven case names and verifies their valid/invalid classification through the generated index.

## Validation evidence

Every gate below ran directly without `tee`.

- Expected red: `go test ./tools/generate-vectors` before regeneration exited 1 because the new inventory requirement correctly reported missing `valid-empty-builds.json`.
- `go run ./tools/generate-vectors -root .`: exit 0.
- `go test -count=1 ./tools/generate-vectors`: exit 0.
- Pinned-venv `python -B tools/validate.py`: exit 0; 42 schemas and 400 vector files.
- Pinned-venv Python unittest discovery: exit 0; 15 tests.
- `go test -count=1 ./tools/...`: exit 0.
- Pinned-venv `make validate`: exit 0.
- `go vet ./tools/...`: exit 0.
- `gofmt -l tools/generate-vectors` emptiness assertion: exit 0.
- `go build -o <task-temp>/generate-vectors-rework2 ./tools/generate-vectors`: exit 0.
- A built generator populated an empty output root, and `diff -rq`/`cmp` comparisons against checked schema cases, expected files, vectors, and manifest all exited 0.
- A second pass retained byte digest `631d1555ecfa7d4e7223b1e1789dd60c26e8a8e320b1e6db5e51c7b758a7296f`; all second-pass comparisons exited 0.
- Accepted rc.4 legacy schemas/cases and accepted schema-7 prose comparisons: exit 0.
- `git diff --check`, `git diff --cached --quiet`, pinned-HEAD, and repository-artifact-cleanliness assertions: exit 0.

No schemas, semantic validators, accepted prose, legacy bytes, manager implementation, shared runtime vectors, or release-promotion files were changed in rework 2. No files were staged or committed.
