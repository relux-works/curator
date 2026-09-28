# TASK-260922-1nf6o6 results

## Delivered

- Added schema 2 at `schemas/v1/agent-environment-marker-v2.schema.json`, id `https://relux-works.github.io/curator-spec/schemas/v1/agent-environment-marker-v2.schema.json`.
- Schema 2 is structurally schema 1 outside the intended revision: only `$id`, `title`, `version.const`, and the `passthrough` definition differ. The schema-1 schema bytes are unchanged.
- Each schema-2 passthrough record requires `isolation`, `strategy`, `source_role`, `backend`, `backend_version`, and `provenance`, with strict enums and no additional members. `path` is omitted for `ambient` strategy/backend and isolated `per-home-keychain`; other records require it.
- Added generator-owned v2 cases: 7 valid and 19 invalid. Valid records cover every strategy, including both `keyring-preferred` backend/path branches. Negative cases cover each missing record member, unknown backend/provenance/isolation/source role/strategy, empty backend version, path present for linkless records, path absent for linkable records, and an unknown extra member.
- Kept the schema-1 schema and existing schema-1 cases unchanged, and added one schema-1 negative case proving that a schema-2 credential record is rejected. The v1 schema family is still valid without a credential record.
- Registered schema 2 in the generator, conformance index and manifest. Updated the generated RC9 candidate manifest pin as required by the generator; historical release pins and claims were not changed.
- Updated environments §§7.4, 8.2 and 8.4.1, manager §§12.4/12.5, and `CHANGELOG.md`. The docs require locked same-directory temporary publication plus atomic rename inside the journaled mutation, exact marker restoration on rollback, `lstat` no-follow discovery/backups, and no credential-byte archiving.
- Cached successful schema meta-validation by canonical schema content in `tools/validate.py`; changed schema content is rechecked. This avoids revalidating the same 64 schema documents for every end-to-end validation invocation in the Python suite.

## Curator follow-up leaf

Create the curator implementation leaf **Publish schema-2 credential records in the environment manager**. It must require the manager to publish schema id `https://relux-works.github.io/curator-spec/schemas/v1/agent-environment-marker-v2.schema.json` and the exact record members `isolation`, `strategy`, `source_role`, `backend`, `backend_version`, and `provenance`. It must omit `path` for ambient strategy/backend and isolated `per-home-keychain`, and require `path` for other strategies. A schema-1 marker must not be rewritten solely to add the record; when a successful manager-home mutation independently requires replacement marker publication, publish schema 2 with complete records. Rollback restores the exact prior marker bytes and version.

## Verification evidence

| Command | Exit | Result |
|---|---:|---|
| `python3 tools/validate.py` (with the task-local venv and `jsonschema==4.25.1`) | 0 | Validated 64 schemas and 1166 vector files, including all 26 schema-2 cases and the schema-1 compatibility case. |
| `python3 -B -m unittest test_validate.EnvironmentVectorTests.test_environment_schema_semantics_fail_closed` (from `tools/`, task-local venv active) | 0 | Focused marker semantic regression test passed. |
| `python3 -B -m unittest test_validate.SchemaRegistryCacheTests test_validate.EnvironmentVectorTests.test_environment_schema_semantics_fail_closed` (from `tools/`, task-local venv active) | 0 | Cache reuse and changed-schema invalidation passed alongside marker semantic validation. |
| `go test ./tools/generate-vectors` | 0 | Generator tests passed, including closed schema-case coverage. |
| `GOFLAGS=-p=1 GOMAXPROCS=2 go test ./tools/...` | 0 | Go tool tests passed with bounded build parallelism. |
| `go vet ./tools/...` | 0 | Go static analysis passed. |
| `make regenerate-check` | 0 | Generator output reproduced the staged conformance and RC9 generated artifacts. |
| `gofmt -w tools/generate-vectors/main.go tools/generate-vectors/environments.go tools/generate-vectors/environments_test.go` | 0 | Go files formatted. |
| `git diff --check` | 0 | No whitespace errors. |
| `PATH=.temp/validation-venv/bin:$PATH GOFLAGS=-p=1 GOMAXPROCS=2 make validate` | 0 | Validated 64 schemas and 1166 vector files; all 580 Python tests passed in 344.950 seconds; Go package tests passed in 0.761 seconds. |

Earlier `make validate` attempts reported these non-green outcomes: exit 2 under system Python because `jsonschema` was absent; exits 2 and 130 for task-venv runs interrupted during Python discovery; exit 130 after the cached run's Python suite passed but its default-parallel Go phase stalled. A separate uncapped `go test ./tools/...` attempt also exited 130 after 90 seconds; the bounded-parallel standalone Go test and final full target both exited 0. Direct schema meta-check timing was 0.061 seconds for both marker v1 and v2. A standalone repository lint target/config was not present in the worktree; `gofmt`, `go vet`, and whitespace validation passed.

## Scope and bounds

No frozen protocol schema was edited. No curator implementation, fleet-policy, or fragment work was included. The marker upgrade rule above is normative and consistent across environments and manager documentation. The final bounded-parallel `make validate` run exited 0; earlier interrupted attempts are listed above.
