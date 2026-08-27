# TASK-260728-2jaw7h reviewer verdict — cycle 2

## ACCEPTED

Route: `done`. No blocking or rework finding remains.

The cycle-1 defects are repaired without weakening the accepted toolchain
contract or advancing a release identity:

- `release/1.0.0-rc.5.json` is byte-identical to the accepted predecessor at
  SHA-256 `b32ee9d35fc2fce0539d0b1c4b15f9c5239115c22fe54a72401f5da6d6646441`.
- `conformance/v1/manifest.json` is byte-identical at
  `9ba9b8ecf6f06cafda1425aed3539dee8d12af43dabd36f803f4f420dbb1dacf`.
- `conformance/v1/schema-cases/index.json` is byte-identical at
  `2faa2baaadca30b3eebe3b9248260efcfac30e92cef4fa209bf37f3f23efd4f0`.
- `diff -rq` confirms the complete `conformance/v1` tree is byte-identical to
  the accepted `TASK-260728-2kp3tv` predecessor.
- `conformance/next/manifest.json` carries no `protocol_version`, records
  `released: false`, names `1.0.0-rc.5` only as `candidate_against`, and assigns
  pin ownership to `TASK-260728-251p01`. Its inventory is sorted, complete,
  digest-checked, and disjoint from the frozen suite.

The authored `release/frozen.json` record is enforced independently by the
generator, ordinary validator, and release gate. Targeted regression tests
prove that a rewritten suite manifest plus a matching release-document repin
still fails all three paths. The validator also rejects a version/release claim
on `conformance/next`, a candidate case in the released root, and orphan case
files in either suite. The reference-document drift tests bind
`source_ref.surface=registry` and the `matches=absence|value` classifier rules
to the schemas, shipped registry, `protocol/core.md`, and both classifier
tables.

The normative landing matches the accepted decisions:

- the requirement is a closed `{id, version}` object with the canonical
  three-part version grammar and exact/range/at-least union;
- identifiers are closed to `go`, `kotlin`, `rust`, and `swift`, with package
  requirements constrained to the driver's registry-owned primary toolchain;
- the toolchain member is required on both schema-8 build commands and optional
  on descriptor schema 2 through one shared definition;
- package-controlled paths, roots, URLs, mirrors, channels/tracks, version
  managers, install/package-manager commands, environment, credentials,
  keyrings, checksums, and trust roots are rejected by generated cases and the
  wire-surface gate;
- the registry, resolution evidence, resolved fingerprint identity, twelve-code
  diagnostic taxonomy, manager-owned guidance catalog, security boundary, and
  no-auto-install semantics are closed and cross-validated;
- preflight cases 61-64 and 107-113 prove Stage A before acquisition/cache/
  mutation/compiler work, Stage B ordering after external acquisition, and the
  no-persistent-mutation boundary for local and external sources.

Independent compatibility checks found zero changes across 41 legacy schema
files and 57 pre-existing `common.schema.json` definitions. The whole frozen
corpus remains identical, preserving schemas 1-7 and accepted Go behavior.

## Independent gates

- `tools/validate.py`: exit 0, 48 schemas and 592 vector files
  (422 released, 170 candidate).
- Python unit tests: exit 0, 169 tests.
- `go test ./tools/...`: exit 0.
- `go vet ./tools/...`: exit 0.
- `gofmt -l tools`: no output.
- Python `compileall`: exit 0 with an external bytecode cache.
- `git diff --check`: exit 0.
- Clean deterministic regeneration: two consecutive `make regenerate-check`
  runs, both exit 0 and leave the clean probe unchanged.
- Clean release gate: exit 0 for `1.0.0-rc.5` at
  `0b14f2028bd86b8934c2043ddaca8ee6a9b533a3`.
- Targeted anti-laundering/candidate/reference tests: 23 validator tests,
  5 release-gate tests, and 3 generator tests, all green.
- Boundary probe: exit 0 against Go 1.25.1 and Go 1.25.5, 16 `go` cases,
  13 `toolchain` cases, 331 closure checks, zero failures.
- All five boundary expected-red controls failed as required.

The review did not edit, stage, commit, publish, mint, or advance pins in the
candidate worktree.
