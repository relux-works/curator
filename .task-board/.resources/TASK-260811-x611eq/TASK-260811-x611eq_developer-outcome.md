# TASK-260811-x611eq — cross-language adapter conformance

Status: implemented, tested, gates green locally, handed to review.

Deliverable: one integration proof over the already-accepted adapter contracts,
in a new package `internal/crossconformance`, plus the supported/unsupported/
diagnostic/migration documentation and a committed protocol export for an
independent implementation.

No adapter was re-implemented, no accepted adapter production file was changed,
and no security decision was re-derived. The suite drives the delivered
adapters through their own exported seams.

## What landed

### `internal/crossconformance` (production, standard library only)

| File | Contents |
| --- | --- |
| `ccj.go` | An independent CCJ-1 scanner, canonicalizer, and domain-separated identity function. Hand-written tokenizer over raw bytes; it does not call `internal/protocoljson` or `internal/closuregraph`, which is the point of an oracle. |
| `corpus.go` | The embedded accepted 53-record CGP05/CGP10 corpus, its pinned SHA-256, and a parser that derives every identity itself and rejects a record whose published hash is not the one it derives. |
| `validate.go` | The structural claims: typed reference resolution per label, selection-neutral capture, binding-only target authority, CGP05 capture reuse with two distinct target branches, and the CGP10 stable/branch split. Emits the two accepted oracle summary lines from independently derived counters. |
| `suite.go` | The published contract: six delivered paths, seven normative obligations with their `Check*` functions, nineteen rejection vectors with their closed diagnostic sets, and a coverage matrix that refuses to be incomplete. |
| `export.go` | The whole contract rendered as exact CCJ-1 for an independent implementation. |
| `testdata/canonical-goldens.txt` | Byte-identical copy of the accepted corpus; a test proves it equals `internal/closuregraph/testdata/canonical-goldens.txt`. |
| `testdata/cross-adapter-protocol-export.json` | The committed export. |

A guard test proves the production files import no repository package at all,
so the oracle cannot end up checking itself, and that nothing in the package or
in `internal/closureexec` launches a child process outside the two shared
commit-before-start seams. The guard assembles its needles at run time so it
scans its own source as strictly as everything else; the cross-adapter
allowlist is not widened.

### Documentation

- `docs/source-closure-adapter-conformance.md` — supported profiles, explicit
  unsupported cases, the full stable diagnostic vocabulary with precedence, the
  conformance suite and its coverage tables, the delegation table, the
  environment requirement, and migration steps for an existing command.
- `README.md` — a new "Cross-adapter source-closure conformance" section after
  the per-manager profiles, and a row in the gates-and-tooling table.

## The 53-record corpus, independently

`internal/closuregraph` already round-trips the corpus through its own codecs.
This task adds a second, independent implementation and requires the two to
agree:

- every record is decoded by this package's own scanner, re-emitted, required
  to be byte-identical to the published payload, and hashed as
  `SHA256(label || 0x00 || CCJ(payload))`;
- the derived identity must equal the published one for all 53;
- the derived identity must also equal `closuregraph.IDFromCanonical`'s, so a
  change in either implementation surfaces;
- every typed reference is resolved by label and, for node references, by node
  kind: capture graphs, edges, selection contexts, bindings, active graphs,
  build plans, checkpoints, closures, expected cache inputs, observations, and
  execution/publication receipts. `125` typed references resolve.

The independently derived summary is byte-identical to the accepted Ruby
oracle's:

```text
canonical_goldens=pass labeled_records=53 cgp05_target_branches=2 cgp10_observation_branches=2
canonical_references=pass cgp05_capture_reused=true explicit_target_bindings=2 cgp10_all_refs_resolve=true
```

Four tamper cases prove the validator is load-bearing rather than decorative.
Each rewrites the corpus to a hash fixed point first, so the structural rule is
what rejects it, not the hashing:

| Tamper | Rejected because |
| --- | --- |
| capture absorbs the target platform node | `retains selection-specific node kind "target_platform"` |
| binding carries a command-product node | `binds forbidden node kind "command_product"` |
| `cgp10.closure` points at a non-record | `references unresolved fixture record` |
| C5 payload gains `active_graph_id` | `C5 may add no graph record` |

A fifth case proves a payload edit without a matching hash is rejected before
any structural rule runs.

### One finding about the corpus

`cgp05.platform.darwin` and `cgp10.platform` are the same record: same label,
same CCJ bytes, therefore one identity. The first draft of the parser treated a
repeated identity as a corpus defect and failed. That was wrong — a record
published under two fixture names is still one graph record. The parser now
rejects only a genuine collision: two different labels or payloads landing on
one identity, which would break domain separation.

## The normative semantic suite

Six delivered paths — Rust, npm, pnpm, Yarn Classic, modern Yarn, SwiftPM —
each project one fixed set of source inputs onto **two exact targets** through
their own production APIs, and each must satisfy the same seven obligations.

How each path is driven, and where its records come from:

| Path | Driven through | Selection-bound identity |
| --- | --- | --- |
| npm | `npmsource.Parse` → `npmsource.CaptureAndAdmit` → `nodesource.NewC0Checkpoint`/`Close` | `closuregraph.SelectionBinding` |
| pnpm | `pnpmsource.Parse` → `pnpmsource.CaptureAndAdmit` → same common Node contract | `closuregraph.SelectionBinding` |
| Yarn Classic | `yarnclassicsource.Parse` → `CaptureAndAdmit` → same | `closuregraph.SelectionBinding` |
| modern Yarn | `yarnmodernsource.Parse` → `CaptureAndAdmit` (real normalized cache ZIPs, real `CacheChecksum`) → same | `closuregraph.SelectionBinding` |
| SwiftPM | `swiftpmsource.CaptureAndClose` with fakes only at the four declared interface seams (evaluator, broker, mirror verifier, offline runner) | `closuregraph.SelectionBinding` |
| Rust | real `rustsource.NewManager` → `Capture` → `DeriveMetadata`, plus `ParseLock`/`ParseManifest`/`NewCaptureGraph`/`ParseMetadata`/`Reconcile` for the two-target projection | `rust-active-graph-v1` identity |

Rust is the one path whose selected overlay is not a `closuregraph`
`SelectionBinding` — `rust-source-v1` carries the exact target and toolchain in
its own active-graph identity. The obligation is stated over "the path's exact
selection-bound identity", so Rust proves the same requirement through its own
record family. The suite says so explicitly (`EmitsBindingRecords`) instead of
skipping the record-kind census.

Obligations proved, per path, all six:

| Obligation | What it required |
| --- | --- |
| `capture.selection_neutral` | no `target_platform` or `toolchain_component` node and no `targets`/`uses_tool` edge inside the capture, plus a text check that no bound tool fingerprint or target identity is spelled anywhere in the capture's own canonical records |
| `capture.stable_across_targets` | one exact capture identity across both targets |
| `binding.target_authority` | exactly one bound target platform node, only binding-legal node and edge kinds, at least one explicit `targets` edge reaching the bound platform, and at least one exact tool identity |
| `binding.diverges_per_target` | selection-bound, active, and plan identities all differ |
| `records.deterministic` | a second capture from freshly built inputs reproduces every identity |
| `evidence.causal_chain` | emitted checkpoints chain to their exact predecessor, C5 adds no graph record, and every pre-C5 evidence derivation answers with a valid, unique causal receipt |
| `artifact.shared_admission` | see below |

`coverage-is-complete` then refuses an incomplete matrix, and says so in its
failure text: a filtered `-run` that skips the proving subtests fails this gate
rather than reporting a green integration proof over an empty matrix.

### Shared artifact admission, stated correctly

The first version of this obligation required the whole shared dependency
corpus to produce identical results through all six profiles. It does not, and
should not: the accepted policy lets an adapter narrow its allowed source
grammars, so a Go or Python source fixture is legitimately `opaque.unknown`
under `rust-source-v1`. Forcing agreement there would have meant either
weakening a profile or asserting something false.

The obligation is now the accepted C12 claim plus its complement:

1. **deny classes are identical across paths** — every bare compiled leaf in
   the accepted corpus (`70` cases) is presented through each path's own
   declared adapter and profile identity, and all six must return the identical
   class, node decision, manifest decision, primary diagnostic, and leaf
   digest;
2. **each profile admits exactly its own source grammars** — every admitted
   dependency vector the corpus publishes for a path's profile must be admitted
   by that path with the accepted class and decision.

## The rejection matrix

Nineteen published vectors. Sixteen are driven here through the adapters' own
seams; each requires a stable diagnostic from a closed set, zero affected
process starts, and no publication.

```text
binding-duplicate-record=npm+pnpm+swiftpm+yarn-classic+yarn-modern
binding-dangling-reference=npm+pnpm+swiftpm+yarn-classic+yarn-modern
binding-wrong-kind=npm+pnpm+swiftpm+yarn-classic+yarn-modern
binding-replaces-capture=npm+pnpm+swiftpm+yarn-classic+yarn-modern
binding-missing-target=npm+pnpm+swiftpm+yarn-classic+yarn-modern
build-cycle=npm+pnpm+yarn-classic+yarn-modern
compiled-dependency-bytes=npm+pnpm+rust+swiftpm+yarn-classic+yarn-modern
opaque-dependency-bytes=npm+pnpm+rust+swiftpm+yarn-classic+yarn-modern
verified-binary-unavailable=npm+pnpm+rust+swiftpm+yarn-classic+yarn-modern
integrity-mismatch=npm+pnpm+yarn-classic
offline-input-missing=npm+swiftpm+yarn-modern
target-identity-drift=rust+swiftpm
toolchain-identity-drift=swiftpm
undeclared-process=swiftpm
undeclared-input=npm+pnpm+swiftpm+yarn-classic+yarn-modern
unreceipted-output=npm+pnpm+rust+swiftpm+yarn-classic+yarn-modern
```

The compiled-byte vector uses the pinned `GNUSharedObject()` fixture from the
accepted shared corpus, injected into each ecosystem's own package payload —
an npm/pnpm/Yarn Classic tarball member, a modern Yarn cache-ZIP member, a
SwiftPM package tree file, and a Rust path-dependency tree file — so the same
bytes cross six different capture front doors. Every one rejects with
`artifact_compiled_dependency_forbidden`, and SwiftPM's observer confirms zero
manifest process starts.

### Three vectors are delegated, and say so

`network-attempted`, `undeclared-write`, and `output-drift` cannot be
constructed through the delivered adapters' exported seams. Their failing
condition lives behind a live verified execution provider
(`closureexec.Executor` requires a provider-issued `Audit`) or a sealed
in-package authority (`artifactpolicy.LocalOutputAuthorization` has an
unexported method and is only constructible inside `internal/artifactpolicy`).
An integration package could reach them only by forging evidence, which would
prove nothing about the adapters.

They stay published in the matrix with their owning packages named, and
`TestDelegatedDiagnosticsAreDeclaredByTheirOwners` holds a compile-time
reference to each owner's own diagnostic constant, so a renamed or deleted code
breaks the build rather than quietly emptying the matrix.
`TestDelegatedVectorsNameRealOwningPackages` proves each named owner is a real
package directory.

| Delegated vector | Owning accepted suites |
| --- | --- |
| `network-attempted` | `internal/closureexec`, `internal/npmsource`, `internal/rustsource`, `internal/swiftpmbuild` |
| `undeclared-write` | `internal/closureexec`, `internal/rustsource`, `internal/swiftpmbuild` |
| `output-drift` | `internal/artifactpolicy`, `internal/swiftpmbuild`, `internal/nodesource` |

## The protocol export for an independent Python implementation

`testdata/cross-adapter-protocol-export.json` (`35399` bytes) is exact CCJ-1
containing the accepted corpus with independently derived identities, the
counters derived from it, the seven obligations with their normative sentences,
the six delivered paths, and all nineteen rejection vectors with their closed
diagnostic sets and delegation. Its tests prove the file is canonical, that
every exported record's payload still derives its exported identity, that every
obligation states a requirement, and that every published diagnostic is a
stable lowercase code. No Python code was added to this repository; the
committed export is the interface.

Regenerate with `CURATOR_WRITE_CROSS_EXPORT=1 go test -run
TestProtocolExportMatchesTheCommittedGolden ./internal/crossconformance` — the
test then fails on purpose so a regeneration cannot be mistaken for a pass.

## Gates

All commands run directly as standalone processes; exit codes are the real
ones. No pipes, no `tee`.

| Gate | Command | Exit |
| --- | --- | ---: |
| Cross-adapter suite | `go test -count=1 -timeout 20m -v ./internal/crossconformance` | 0 |
| Repository suite (excluding `cmd/curator`) | `go test -timeout 30m -count=1 $(go list ./... \| grep -v cmd/curator)` | 0, 53 packages ok |
| CLI suite | `go test -timeout 30m -count=1 ./cmd/curator` | 0, `316.568s` |
| Race (new package) | `go test -count=1 -race -timeout 25m ./internal/crossconformance` | 0, `25.842s` |
| Lint | `golangci-lint run` (v2.12.2) | 0, `0 issues.` |
| Format | `gofmt -l cmd internal` | 0, empty |
| Vet | `go vet ./...` | 0 |
| Whitespace | `git diff --check` | 0, empty |
| Broad suppression | `bash .github/ci/no-broad-suppression.sh` | 0, `no-broad-suppression: ok` |
| Accepted canonical oracle | `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb .research/260811_cross-language-closure-graph-and-checkpoints.md` | 0, both accepted lines |
| Board | `task-board --no-update-check validate` | 0, `Board is valid. No issues found.` |

Slowest packages in the repository run: `internal/rustsource 145.340s`,
`internal/artifactpolicy 133.478s`, `internal/install 117.523s`,
`internal/install/atomicity 115.203s`, `internal/swiftpminterop 106.475s`,
`internal/npmsource 102.969s`, `internal/godriver 94.448s`.

Attached log digests:

- `crossconformance-verbose.log` `efe3a0501f774e6fc93a3fa14759bc129d5d40c04c44b8bce53cfc571aed7811`
- `full-suite-noncmd.log` `59f31ade83407226badcfe7ced1f1989a57d340cbc72d42a303db6922cc33716`
- `full-suite-cmd.log` `378e4ae59b3a18f7d2f55ca2e254fd16129588e9f62363572f815fd6d457b7aa`
- `race-crossconformance.log` `c5a413900e12829026c6183b589d27a48ca1528ab45d6bc3452828c892275a7a`
- `canonical-verifier.log` `1847364de63c9d8e706d54739f97d05e8406cca5f9cd7aec4bb12c028998b75a`
- `lint.log` `e92606b0bf483111dff0a120c315ea165821348f31365020e2468a0059095c47`
- `no-broad-suppression.log` `45f2812ab39fa44040a0da7ad9717c50d7b0adfe17f58fcc12d7fb55784628bc`
- `gofmt.txt`, `vet.log`, `diff-check.log` are empty
  (`e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`).

No compiled artifact was added: every new file is `text/plain` or
`application/json`.

## Two things a reviewer should decide on

### 1. The Rust path needs the pinned Cargo, and CI installs no Rust

The cross-adapter suite drives the real `rustsource.Manager` so that Rust's
causal-evidence obligation consumes the same C0 Cargo registration, pinned
vendor transform, and metadata receipts the accepted Rust suite does.
`rust-source-v1` admits exactly one approved Cargo descriptor per native target
(`1.91.0`, commit `ea2d97820c16195b0ca3fadb4319fe512c199a43`, currently only
`aarch64-apple-darwin`), so those cases need that toolchain on the host.

This is a **pre-existing delivery condition, not a new one**:
`internal/rustsource`'s own accepted conformance tests call
`rustsource.NewManager` and `t.Fatal` on failure, with no skip class in
`.github/ci/skip-classes.tsv` for an absent Cargo. `.github/workflows/ci.yml`
installs Go only — no Rust, Node, pnpm, Yarn, or Swift. So `internal/rustsource`
is already red on every CI runner, and the cross-adapter suite inherits that
condition on the Rust cases.

I deliberately did not paper over it. The alternatives were both worse: a
host-conditional skip is exactly what `skip-classes.tsv` exists to police, and
dropping the manager would leave Rust's causal-evidence obligation with no
receipts to prove. The environment requirement is documented in
`docs/source-closure-adapter-conformance.md`. Whether CI grows a Rust
toolchain step, or `internal/rustsource` and the Rust cases here gain a declared
skip class, is a delivery decision above this task.

### 2. No platform-case ledger row was added

No adapter package (`rustsource`, `npmsource`, `pnpmsource`,
`yarnclassicsource`, `yarnmodernsource`, `swiftpmsource`, `swiftpminterop`,
`swiftpmbuild`, `nodesource`, `closuregraph`, `artifactpolicy`) has a row in
`.github/ci/platform-cases.tsv`; that ledger is scoped to compiled-build
platform behaviour. This task follows the same precedent. Adding rows for the
portable corpus and export cases would be defensible, but adding one for the
whole cross-adapter case would encode a Rust-toolchain requirement into the
per-runner ledger, which is the wrong place to decide item 1.

## Scope notes

- No accepted adapter production file was modified. `git status` shows the new
  package, the new document, and the README edit as this task's only changes;
  everything else in the tree is the previous tasks' committed and uncommitted
  delivery, preserved as instructed.
- Nothing was staged or committed.
- The suite adds no `os/exec` use anywhere and does not widen the cross-adapter
  process-guard allowlist.
