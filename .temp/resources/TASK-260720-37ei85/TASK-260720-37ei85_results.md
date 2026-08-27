# TASK-260720-37ei85 implementation results

## Baseline and scope

- Isolated worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-37ei85/worktree`
- Frozen base: curator-spec `57c1f56846d221ecc55786bd3c2467ec32f11730`
- Accepted composite source: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-2zc6k1/worktree`
- The accepted predecessor product diff and untracked product files were compared byte-for-byte before implementation.
- Relative to the accepted composite, this task changes only `tools/generate-vectors/main_test.go` and the approved schema-v1 compatibility sentence in `protocol/core.md`.
- No commit or real-index staging was performed. A task-local alternate index was used only for deterministic regeneration checks over the intended uncommitted conformance baseline.

## Compatibility guards

- Both `agent-skill` and `csk-skill` schemas 1 through 5 explicitly retain the shared script/system-only command union and therefore reject `type: build`.
- Both schema-v1 manifests explicitly retain `additionalProperties: true`; `build_roots` remains undeclared and therefore has no protocol-defined build meaning when accepted incidentally.
- Both manifest names at schemas 2 through 5 explicitly keep `build_roots` undeclared with `additionalProperties: false`.
- Legacy schema-case index entries remain exactly `valid.json=true` and `invalid.json=false` for both manifest names at versions 1 through 5.
- The existing frozen digest for the 30 legacy manifest schema/evidence files is now identified with the exact `57c1f568` baseline.
- Claim v1 now has explicit schema-version 1, protocol rc.3, exact-property, and byte-hash guards.
- Install marker v1 now has explicit historical property/version bounds, no build-era fields, and byte-hash guards.
- Manager/system configuration schemas now have exact top-level surfaces and recursive declared-property rejection for driver, argv/environment, toolchain, output, hook, and build-policy override names.
- Audit and registry schemas now have exact top-level evidence surfaces and recursive declared-property rejection for local artifact attestation, receipt, provenance, build-source, and cache-key names. Audit required evidence remains source identity, commit, and `content_sha256`.
- The rc.4 prose now accurately states that schema-v1 `build_roots` is an extension without build semantics, schemas 2 through 5 reject it, and schemas 1 through 5 reject build commands.

## Frozen and intentional change evidence

- 48 legacy schema/case/config/registry/audit files compare byte-for-byte with `57c1f56846d221ecc55786bd3c2467ec32f11730`.
- Claim-v1 schema/valid/invalid hashes remain `c9f494…7b3f`, `799682…93c`, and `de9568…017`.
- Install-marker-v1 schema/valid/invalid hashes remain `aff807…855`, `80989f…28a`, and `c0ea74…0fa`.
- The intentional rc.4 conformance delta is 73 files relative to the frozen base: the manifest/index updates plus new v6, build-receipt-v1, claim-v2, and marker-v2 cases.
- Current generated manifest: protocol `1.0.0-rc.4`, 164 vector files across 35 schemas, SHA-256 `ce8a26cfe60c8ada0782328410c3e8e4cde78dd67c43dade066703a1d0703d36`.

## Tool readiness

- `rg 15.1.0`, `go1.25.5 darwin/arm64`, GNU Make 3.81, Git 2.50.1, task-board 0.20.1, and Python 3.14.4 were located and executed successfully.
- Host Python lacked `jsonschema`; validation used task-local `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-37ei85/venv`, installed from pinned `requirements-dev.txt` (`jsonschema==4.25.1`).

## Verification

- `gofmt -d tools/generate-vectors/main.go tools/generate-vectors/main_test.go` — pass.
- `go vet ./tools/...` — pass.
- `go test -count=1 ./tools/generate-vectors` — pass.
- `make regenerate` — pass twice as explicit passes, in addition to the generator executions inside both regeneration checks.
- `make validate` — pass: 35 schemas, 164 vector files, 8 Python tests, and all Go tool tests.
- `make regenerate-check` — pass twice with the task-local alternate index; no conformance diff.
- `git diff --check` — pass.
- Real Git index remains untouched.

## Unrelated board validation anomaly

- `task-board validate` reports 13 existing board-hygiene issues outside this task: 12 broken `blockedBy` references to July 12 epic IDs documented in `docs/implementation-plan.md`, plus one orphan resource file under `TASK-260713-7a9c1e`.
- The current task resolves successfully through the query API with all 10 checklist items checked and its task-scoped outcome resource linked. The unrelated board-hygiene findings do not affect curator-spec generation, validation, or this handoff.
