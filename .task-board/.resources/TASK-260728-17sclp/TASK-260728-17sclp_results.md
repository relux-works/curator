# TASK-260728-17sclp wire-schema outcome

Worktree: `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-17sclp/curator-spec-worktree`

Pinned HEAD: `57c1f56846d221ecc55786bd3c2467ec32f11730`

## Baselines

- Accepted composite rc.4 schemas/cases: TASK-260720-37ei85 worktree.
- Accepted schema-7 prose: TASK-260728-1a52au worktree.
- SECURITY.md, profiles/manager.md, protocol/core.md, Decision 0004, and Decision 0005 remain byte-identical to the accepted schema-7 prose worktree.
- Manifest schemas/cases 1-6, receipt v1, markers v1/v2, claims v1/v2, and Skillfile.dev v1 remain byte-identical to the accepted rc.4 composite.

## Changed-file map

New schemas: agent-skill-v7, csk-skill-v7, curator-build-v1, Skillfile.dev-v2, build-receipt-v2, install-marker-v3, and conformance-claim-v3. Common schema definitions carry the separate v6/v7 wire types. `schemas/v1/README.md`, `tools/generate-vectors/main.go`, `tools/generate-vectors/main_test.go`, `tools/validate.py`, and `tools/test_validate.py` provide registry selection, deterministic generation, semantic cross-field validation, and compatibility guards. The conformance manifest, schema-case index, and all owned generated cases were regenerated.

The review rework tightened transport/canonical/ref grammar, made external Skillfile.dev substitutions optional, generated all reserved schema-7 rejection cases across both manifest filenames and versions 1 through 6, and added the previously missing receipt/marker/claim semantic branch cases. Detailed mapping is attached as `TASK-260728-17sclp_rework-1.md`.

Review rework 2 closes the remaining marker-v3 inventory gap. The marker corpus now contains explicit valid cases for empty builds, safe network tag and branch substitutions, SHA-1 and SHA-256 revision substitutions, an unsubstituted SHA-256 external record, and an untagged external record. It also contains marker-native invalid cases for local/network substitution identity-kind mismatches and both structured revision-width mismatches. The Go inventory test requires every new case. Exact names and gate evidence are attached as `TASK-260728-17sclp_rework-2.md`.

The schemas remain closed around immutable SHA-1/SHA-256 locks, strict repository commands and descriptor targets, local/network substitutions, declared/effective receipt branches, mixed receipt-v1/v2 marker entries, and claim-v3 Go driver/platform assertions. Package-controlled argv, environment, output/name, credentials, signing, hooks, plugins, generators, fallbacks, and generic drivers are rejected.

No manager implementation, exact Git CLI profile text, shared runtime vector corpus, or release promotion was added; the conformance manifest remains rc.4 while claim schema 3 defines the future rc.5 wire claim.

## Validation evidence

All listed gates ran directly without `tee`.

- Task-local pinned Python environment: jsonschema 4.25.1.
- `python -B tools/validate.py`: exit 0; 42 schemas and 400 vector files.
- Python unittest discovery: exit 0; 15 tests.
- `go test -count=1 ./tools/...`: exit 0.
- `PATH=<task-venv>/bin:$PATH make validate`: exit 0.
- `go vet ./tools/...`: exit 0.
- `gofmt -l tools/generate-vectors`: exit 0 with no output.
- Go generator build to a task-temp output: exit 0.
- Clean-from-empty generation and checked-corpus comparisons: exit 0.
- Second generation pass retained digest `631d1555ecfa7d4e7223b1e1789dd60c26e8a8e320b1e6db5e51c7b758a7296f`; all second-pass directory comparisons exited 0.
- Accepted rc.4 legacy schema/case byte comparison: exit 0.
- Accepted schema-7 prose byte comparison: exit 0.
- `git diff --check`: exit 0.
- `git diff --cached --quiet`: exit 0.
- Pinned-HEAD assertion: exit 0.
- Repository build-artifact and Python-cache absence assertion: exit 0.

Expected-red evidence is recorded in the two rework artifacts. In rework 2, the focused Go inventory test exited 1 before corpus regeneration because the newly required marker cases were not yet present; after regeneration, the same test exited 0.
