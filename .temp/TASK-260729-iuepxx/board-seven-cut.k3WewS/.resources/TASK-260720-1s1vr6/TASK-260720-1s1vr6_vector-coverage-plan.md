# Vector coverage plan — TASK-260720-1s1vr6

**Task:** `TASK-260720-1s1vr6` — Generate go-v1 build-driver conformance vectors  
**Accepted contract:** `TASK-260720-poa3ze_compile-only-build-drivers.md`  
**Contract SHA-256:** `6308d99d8bdad4445841bc1cbd230cadbed0020012d0e9d38d877b413348f681`  
**Repository baseline:** `curator-spec origin/main` at `57c1f56846d221ecc55786bd3c2467ec32f11730`

## Single deliverable

Generate one implementation-neutral Go fixture plus `conformance/v1/vectors/build-drivers.json`, generator code, and generator tests that cover the accepted go-v1 declaration, exact portable identities, and every build-driver rejection boundary. Shared install/rollback/concurrency scenarios stay in `TASK-260720-cw39jh`.

## Required positive evidence

| Evidence family | Required coverage |
|---|---|
| Manifest and context | Schema 6 with mixed script/build commands; nested `build_roots` subtree excluded while `SKILL.md` and unrelated eligible assets remain visible; cache-hit and dry-run paths execute no source-aware Go command. |
| Fixture | Separate module below `build/`, one `main` command, transitive embedded input, standard-library-only success, and a correctly vendored dependency variant. Existing script fixture and registry expected hashes remain unchanged. |
| Build source | Exact `curator-build-source-v1` bytes: domain prefix, unsigned UTF-8 path order, `uint64be` path/content framing, empty and binary files, root `.csk-install.json`, and mode/timestamp non-inputs. |
| Toolchain | Exact `curator-go-toolchain-v1` bytes: unsorted directories/files/internal links, normalized LF and CRLF `go version`, and mode/timestamp non-inputs. |
| Driver | All five manager-owned argv forms, fixed clean environment, native target/tuning, vendor-only dependency inspection, internal-link build, artifact path/hash/size, protected cache hit, and compiler-free dry-run miss. |
| Portable metadata | Exact CCJ-1 build-input bytes and cache key; exact stored canonical receipt bytes and receipt SHA-256; build-source, toolchain, artifact, and marker identities. |

## Required negative evidence

| Rejection cluster | Minimum cases |
|---|---|
| Structural manifest | Build command in schema 5; unknown driver; forbidden argv/environment/output/toolchain/hook fields; mixed command shapes. |
| Build-root/path | Missing, unused, overlapping, runtime-overlapping, or root build roots; outside/nested module; path escape, link, special file, non-directory, missing root `go.mod`, non-main, or multiple-package selection; build-root content appearing in context. |
| Build-source algorithm | Invalid Unicode, duplicate encoded paths, links/special files, mutation during use, root marker sensitivity, and a legacy NUL-stream structural-collision fixture whose length-framed identities differ. |
| Toolchain algorithm | Invalid Unicode, duplicate paths, escaping/dangling/absolute links, selected executable outside `GOROOT`, malformed/multiple terminal newlines, wrong executable/tool digest, unsupported release family, and tree mutation during use. |
| Module and source graph | Missing/inconsistent vendor data, workspace-only dependency, toolchain switching, cgo/native/SWIG inputs, root/transitive `.syso`, root/transitive assembly including escape attempts, escaped embedded inputs, `go:cgo_import_dynamic`, `default.pgo`, and attempted generators. |
| Process isolation | Poisoned `PATH`, fake repository Go, inherited Go/build variables, telemetry initialization or private-directory failure, external-link default, libgcc fallback, poisoned host tools, or any child outside fingerprinted `GOROOT/pkg/tool/<host>`. |
| Cache and receipt | Wrong key, target, toolchain, policy, receipt/artifact hash or size; noncanonical receipt; partial/link/special artifact; concurrent publisher mismatch; corrupt or untrusted boundary; exact-key self-consistent forgery outside protected state. |
| Context/cache regression | Cache-hit and dry-run context remain source-free without `go list`; root marker embed variants have equal legacy installed-tree hash, distinct build-source hash, semantic rejection before key construction, and zero Go commands. |

## Ownership and handoff

- Own the new Go fixture, expected identity/context files, `build-drivers.json`, and corresponding `tools/generate-vectors` code/tests.
- Do not add shared transaction lifecycle cases, normative prose, release metadata, or manager implementation behavior.
- Keep generated case names stable and unique so the validator can require them explicitly.
- Handoff evidence must include `go test ./tools/generate-vectors`, `make regenerate`, `make validate`, and `make regenerate-check`, plus proof that the existing script fixture and registry expected hashes did not change.

