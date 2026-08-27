# TASK-260728-3b8qym developer handoff evidence

## Scope and workspace

- Assigned worktree:
  `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-3b8qym/curator-spec-worktree`
- Accepted source baseline: detached
  `57c1f56846d221ecc55786bd3c2467ec32f11730`
- Seed: byte copy of accepted uncommitted state from
  `TASK-260728-wy3dsw`, excluding Git metadata, task temp data, caches,
  scratch probes, and build artifacts.
- The assigned worktree remains uncommitted and unstaged.
- A separate disposable `release-probe` repository was used only to exercise
  the real clean-checkout release gate. Its commit is not a protocol,
  implementation, or downstream pin.

## Delivered surfaces

- Generated external-repository fixtures:
  `raw-objects.json`, `lfs-pointers.json`, `local-config-and-refs.json`, and
  `pack-index.json`.
- Generated behavior vectors:
  `external-repository-acquisition.json`,
  `external-repository-lifecycle.json`, and
  `conformance-claim-v3-qualification.json`.
- Exact expected bytes:
  build receipt v2, mixed install marker v3, and mixed-build plan.
- Generator and validator support:
  rc.5 manifest generation, fixture size and exact-hash checks, required
  negative-case inventories, claim-v3 Linux exclusion, exact release-metadata
  verification, and frozen rc.4 byte guards.
- Tests:
  Go fixture/hash/ordering/legacy guards, Python semantic and release-pin
  tests, and release-gate candidate-pin regression coverage.
- Documentation and release metadata:
  author/operator guide, conformance/compatibility/release guidance,
  rc.4/rc.5 changelog entries, workflow/package coverage, and
  `release/1.0.0-rc.5.json`.

## Fixture and behavior matrix

| Area | Evidence |
| --- | --- |
| Acquisition | SHA-1/SHA-256; tagged/untagged; HTTPS/SSH; exact tag match, moved, missing, malformed; revision/tag/branch network substitutions; malformed ref rejected before Git |
| Local admission | valid SHA-1/SHA-256 files refs; gitfile, bare, linked worktree, include, alternate, replace, graft, promisor, partial clone, reftable, link/special-file rejection; filter/helper config proven inert |
| Object proof | exact commit/tag/tree/blob bytes and recomputed IDs; signed/extra headers; duplicate/misordered/missing separators; SHA-256; regular/executable files; symlink, gitlink/submodule, and special-mode rejection |
| Pack/index | exact empty pack v2/v3 SHA-1 and v2 SHA-256 bytes with index v2; pack v4, index v1, missing pair, checksum, and hash-family negatives |
| Git LFS | pinned `git-lfs-pointer-parser-v3.7.1`; canonical and legacy forms; duplicate keys/priorities; size zero; exact 1023/1024 cutoff; near misses and zero bytes |
| Ordering/cache | whole-snapshot validation and independent audit before cache/compiler; hit, miss, corrupt receipt/artifact/boundary, syntax-only offline, install offline |
| Compatibility | receipt v1/v2 and marker v2/v3 local/external/mixed planning; schemas 1-6, `go-v1`, receipt v1, marker v2, and claim v2 frozen guards |
| Lifecycle | private staging, consumer-last marker, rollback, uncertain recovery, read-only status, repair, GC roots, shim/PATH collision, and signing boundaries |

Largest task fixture: 13,256 bytes. The validator and Go tests enforce a
65,536-byte per-file fixture ceiling. The exact LFS boundary fixtures are 1,023
and 1,024 bytes.

## Exact rc.5 candidate and downstream pin

- Protocol version: `1.0.0-rc.5`
- Suite inventory: 42 schemas and 411 manifest-listed files
- Exact downstream candidate protocol pin:
  `sha256:30a64ed0da6e4e68abb5f46e8807f7bc57a4545c7c582e644c9d09c9406c9324`
  (SHA-256 of `conformance/v1/manifest.json`)
- Release metadata SHA-256:
  `ddd8fc11060e164d8e86192263ca6abaf5ec5881edc264e6579a2951f92a0fc3`
- Downstream candidate input: `CURATOR_CONFORMANCE_ROOT`, with mandatory
  comparison to the manifest SHA above.
- `committed_release_pin_advanced` is false. No manager implementation pin was
  advanced.

Claim-v3 candidate claims are empty. macOS and Windows
`go-repository-v1` qualification is pending downstream native evidence. Linux
is excluded until `TASK-260728-1skseh` passes. No platform evidence was
fabricated.

## Validation evidence

Green gates (real exit codes):

- `python tools/validate.py` in the task-local validation environment:
  exit 0, `validated 42 schemas and 411 vector files`.
- `python -B -m unittest discover -s tools -p 'test_*.py'`:
  exit 0, 17 tests.
- `go test ./tools/...`: exit 0.
- `PATH=<task-validation-venv>:$PATH make validate`: exit 0.
- `go vet ./tools/...`: exit 0.
- `test -z "$(gofmt -l tools)"`: exit 0.
- `PYTHONPYCACHEPREFIX=<task-temp> python -m compileall -q tools`: exit 0.
- `git diff --check`: exit 0.
- `git diff --cached --quiet` in the assigned worktree: exit 0.
- Clean release probe `make regenerate-check`, first run: exit 0.
- Clean release probe `make regenerate-check`, second consecutive run:
  exit 0 with zero generated diff.
- Clean release probe `make release-check VERSION=1.0.0-rc.5`: exit 0;
  validation, Python tests, Go tests, regeneration, exact metadata pin, clean
  checkout, version, and candidate release gate all passed.

Expected-red development gates, reported as failures:

- Initial `go test ./tools/generate-vectors`: exit 1 because the test still
  expected claim v2 to follow the upgraded suite constant; fixed by retaining
  the explicit rc.4 claim-v2 constant.
- Initial system `python3 tools/validate.py`: exit 1 because `jsonschema` was
  absent; fixed by installing `requirements-dev.txt` in a task-local venv.
- First dependency-equipped `python tools/validate.py`: exit 1 because the
  generated `invalid-rc4` claim case used rc.5 and became valid; fixed by
  binding it to the frozen rc.4 constant.

## Boundaries preserved

- No real remote was contacted.
- No Curator or csk manager implementation was added.
- No generic language driver or package-selected program/argv/environment,
  output, credential, helper, filter, signer, or fallback surface was added.
- Independent audit-before-cache/compiler, fail-closed offline behavior, and
  operator-owned signing remain intact.
- No source/prior worktree was modified by this task.
