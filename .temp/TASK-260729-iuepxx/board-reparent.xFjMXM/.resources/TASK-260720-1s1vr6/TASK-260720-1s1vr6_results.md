# TASK-260720-1s1vr6 — build-driver conformance vector results

## Baseline and ownership

- Worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1s1vr6/worktree`
- Frozen base: `curator-spec` `57c1f56846d221ecc55786bd3c2467ec32f11730`
- Accepted predecessor: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-37ei85/worktree`
- Imported predecessor tracked-patch SHA-256: `7c80c7a7afd503867304aa518a5a4daf2240c3e2e4394462e295b1d32555b2b8`
- Imported predecessor untracked-inventory aggregate SHA-256: `1bc10b373b6176ea87dd65cbcc17035081062f50e9384156a25b3dde19c7e2e6`
- Task delta relative to the accepted predecessor is limited to the new build fixture, build expected artifacts, `build-drivers.json`, its manifest inventory, and generator code/tests. `manager-lifecycle.json` and shared transaction scenarios were not changed.
- No files were committed or staged in the real index.

## Produced conformance evidence

- Added `conformance/v1/fixtures/go-build-skill`, a schema-6 mixed script/build fixture whose nested build root contains one vendored `main` package and a transitive embedded empty/text input. `SKILL.md` and `assets/prompt.md` remain context-visible; the build root and runtime scripts are excluded.
- Added `conformance/v1/expected/build-driver` with exact context files/hash, `curator-build-source-v1` preimage/hash, accepted CCJ-1 build-input bytes/cache key, exact stored receipt bytes/hash, marker, and `curator-go-toolchain-v1` preimage/hash.
- Added generated `conformance/v1/vectors/build-drivers.json` with 7 positive cases, all 5 direct Go argv forms, the fixed environment, 75 named rejection outcomes, 10 build-source byte/edge cases, and 12 toolchain byte/edge cases.
- Exact accepted portable identities are pinned: cache key `sha256:3fcd714a40e8918eb67dbd35d435875dcce6c9047da811a1fa26626e5e57be48`, stored receipt hash `sha256:750f5f7557ffe1dac1ec06b7d47cb5976e390a7dc7d57ea616c9edda91013c11`, and toolchain digest `sha256:baf7c5f3b9c3f1fae3da4c356381bf74442aa7f8f0b6fb2304c9c10833d6032e`.
- Fixture build-source identity is `sha256:27cdcac0734aa3e069e95a10341e89b118a07c60002516e7b401e95477f01332` and includes the root `.csk-install.json`.
- The exact self-consistent forged receipt remains internally valid at `sha256:9a23f5b77e6173b0f10e7ed43cd2b21aa3b99f3a34945ec432fbb31338a6186d` but has the named `untrusted_provenance` rejection outcome.
- The marker-embed regression carries the accepted equal legacy hash and distinct `0017492c…` / `60fe9b76…` build-source identities, rejects before cache-key construction, and runs no Go command.
- The legacy NUL-stream vector proves one binary file and two records share the old structural stream while their length-framed build-source identities differ.
- Final generated suite aggregate SHA-256: `5ad10d7836a4dab1fea3fcee2367bfc718c46b02c0f5f7211e615fe8b39c019d`; `build-drivers.json` SHA-256: `fd613bbb4d506237fbabd222247478bc1a809266965f4f4b344ba8548661fe33`.

## Frozen evidence

- Generator tests pin every existing script-fixture file hash and every existing registry expected-file hash.
- Script-fixture aggregate SHA-256 remains `2b3279a2fe67f291989eec83eabe2c40d23b05df401cf5efb27fd2e3126005d8`.
- Registry-expected aggregate SHA-256 remains `dbf8be444917c954355332868166f102f33d41739fa225c6aefadeedcdd83389`.

## Verification

- `go test ./tools/generate-vectors -v` — pass, including new fixture, exact identity, coverage, determinism, and frozen-hash tests.
- `go test ./...` — pass.
- `go vet ./...` — pass.
- `gofmt` check and `git diff --check` — pass.
- Physical fixture `go list -mod=vendor -deps -json -buildvcs=false -compiler=gc -pgo=off .` — pass.
- Physical fixture fixed internal-link `go build` — pass; host artifact SHA-256 `b6b6f6f364b615c998e64e9ba5ba8c768bd5e174f81843f98b7a3b9a5d0fb012`; artifact was not executed.
- `make validate` under a task-local virtual environment populated from pinned `requirements-dev.txt` — pass: 35 schemas, 189 conformance files, 8 Python tests, and Go tool tests.
- Two consecutive `make regenerate` passes produced identical suite aggregate `5ad10d7836a4dab1fea3fcee2367bfc718c46b02c0f5f7211e615fe8b39c019d`.
- `make regenerate-check` — pass with a temporary alternate index seeded from the accepted composite, so the real index remained untouched.

The first system-Python `make validate` attempt stopped before validation because `jsonschema` was unavailable. The pinned task-local environment resolved that host-only dependency without changing project files. The first fixture build exposed an invalid multi-file string embed; the fixture was corrected to two explicit embed declarations, regenerated, and all gates above passed afterward.
