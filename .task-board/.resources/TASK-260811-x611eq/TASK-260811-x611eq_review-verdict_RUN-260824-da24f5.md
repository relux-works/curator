# Reviewer verdict for TASK-260811-x611eq

Verdict: **accepted -> done**

## Run and scope evidence

- Reviewer run: `RUN-260824-da24f5` (not goal-bound; `task-board spawn goal` reports
  `Active Goal: none`)
- Reviewed task: `TASK-260811-x611eq` — integrate-cross-language-adapter-conformance
- Producer run: `RUN-260824-a03e56`
- Reviewed deliverable: new package `internal/crossconformance`,
  `docs/source-closure-adapter-conformance.md`, a `README.md` section and gates row,
  and the committed protocol export
- Review policy: `required`
- Directives at the verdict checkpoint: none

This artifact records only the `accepted` branch. As a reviewer-archetype run it
supplies no `commit_ack` and modified no file under `cmd/`, `internal/`, `docs/`,
or `README.md`.

## Acceptance findings, per mandated scope item

### 1. Independent oracle for the 53 records — accepted

The no-repo-import guard is real, not decorative. `go list -f '{{.Imports}}'
./internal/crossconformance` returns only `crypto/sha256 embed encoding/hex errors
fmt regexp sort strconv strings unicode/utf16 unicode/utf8`. The production package
imports no repository package at all, so the oracle cannot validate itself;
`TestIntegrationProductionSourceImportsNoRepositoryPackage` enforces it and requires
at least four production files so an emptied glob cannot pass.

The corpus is the accepted corpus, verified independently rather than by trusting
the pin. I extracted the 53 four-line records directly from
`.research/260811_cross-language-closure-graph-and-checkpoints.md`
(SHA-256 `874c3a40b9ff9fcf130f37f22e9f5aa2bdd6d9f246403e059c98e84736e27bbc`, the
accepted digest) and they are byte-identical to both
`internal/crossconformance/testdata/canonical-goldens.txt` and
`internal/closuregraph/testdata/canonical-goldens.txt`, at
`fed9657b33297650e64178d605c7ee7f47dde640fefd12a99dec3651bb9cadcb`.

I then re-derived every identity with a third, independently written CCJ-1
canonicalizer: `records=53 rehash_ok=53`, every payload already canonical. The only
shared identity is `cgp05.platform.darwin` == `cgp10.platform` — same label, same
bytes, therefore one graph record. The producer's parser finding is correct, and
`corpus.go` rejects only a genuine collision (two different labels or payloads on
one identity), which is the right rule.

`125` typed references resolve. The two derived summary lines are byte-identical to
the accepted Ruby oracle's, which I ran myself against the accepted verifier
(SHA-256 `2254776d4780e4c32ee37ecbf1b22ad092f029ae3ca3be1749ef373c8162d075`, exit 0):

```text
canonical_goldens=pass labeled_records=53 cgp05_target_branches=2 cgp10_observation_branches=2
canonical_references=pass cgp05_capture_reused=true explicit_target_bindings=2 cgp10_all_refs_resolve=true
```

The validator is load-bearing. Each of the four tamper cases rewrites the corpus to
a hash fixed point *first*, so what rejects it is the structural rule and not the
hashing, and each asserts the specific reason text. A fifth case proves a payload
edit without a matching hash is refused before any structural rule runs. All pass.

### 2. One normative semantic across all six adapter paths — accepted

Each path is driven through its own production API, not a re-implementation:
npm/pnpm/Yarn Classic/modern Yarn through `<manager>source.Parse` ->
`CaptureAndAdmit` -> `nodesource.NewC0Checkpoint`/`nodesource.Close`; SwiftPM through
`swiftpmsource.CaptureAndClose` with fakes only at the four declared interface seams
(evaluator, broker, mirror verifier, offline runner); Rust through the real
`rustsource.NewManager` -> `Capture` -> `BuildToolchain` -> `DeriveMetadata`, plus
`ParseLock`/`ParseManifest`/`NewCaptureGraph`/`ParseMetadata`/`Reconcile` for the
two-target projection. Each projects onto two exact targets and discharges the seven
obligations.

The coverage matrix genuinely fails when a path stops proving an obligation. I
verified the exact adversarial case the brief calls out:

```text
go test -count=1 -run 'TestCrossAdapterConformance/coverage-is-complete' ./internal/crossconformance
-> FAIL, "obligations never proved (a filtered -run cannot satisfy this gate)" plus
   every rejection vector listed as uncovered
```

The gate is not satisfiable by a filtered run, and obligations are recorded only
behind `if t.Failed() { return }` / `if !t.Failed()`, so a failing check cannot
credit coverage.

Two honest narrowings, both declared in code, are recorded here because the summary
tables in the outcome document and `docs/` read marginally stronger than what those
two paths exercise. Neither weakens a security claim and neither is rework:

- Rust reports `EmitsBindingRecords: false` and emits no build-plan identity, so
  `binding.diverges_per_target` proves divergence through its own
  `rust-active-graph-v1` identity but does not prove plan divergence, and
  `CheckBindingOwnsTargetAuthority` skips the node/edge-kind census for it. The
  obligation is stated over "the path's exact selection-bound identity", which is
  accurate; the Rust build plan belongs to `TASK-260811-3kbf3l`'s accepted suite.
- The Node paths supply one checkpoint (C0) at this seam, so the
  chains-to-exact-predecessor and C5-adds-no-graph-record rules materially fire only
  on SwiftPM. C1-C7 depth is owned by each adapter's accepted suite, which is the
  correct boundary for an integration package.

### 3. CGP05/CGP10 exact contract — accepted

`validateCGP05` requires both bindings to name the exact capture identity, both
platform nodes to be `target_platform` and distinct, each `targets` edge to reach its
own platform *and* be present in that binding's `binding_edge_ids`, and all eight of
platform/selection/targets/binding/active/plan/C4/C5 to differ across branches.

`validateCGP10` requires all 22 stable records present, each of the read/write/tool
slots bound exactly once and with exactly one slot name, both branches reusing the
stable action, produces edge, expected output, closure, and expected-cache-input,
each execution receipt carrying exactly its own observation, each publication
chaining its execution and the expected cache input, all three of
observation/execution/publication differing across branches, and the two branches
observing different output bytes. This is stricter than the accepted Ruby oracle
(which does not check `producer_action_id`, `published_observation_ids`, or distinct
observed bytes), and it agrees with it.

### 4. Rejection matrix across adapters — accepted

Sixteen of nineteen vectors are driven here. The live coverage output from my own run
matches the published table in `docs/source-closure-adapter-conformance.md` and the
outcome document exactly, vector for vector and path for path.

The compiled-byte vector genuinely crosses six different front doors: the pinned
`policyconformance.GNUSharedObject()` fixture is injected into an npm, pnpm and Yarn
Classic tarball member, a modern Yarn normalized cache-ZIP member, a SwiftPM package
tree file, and a Cargo path-dependency tree file driven through the real
`rustsource.Manager`. Every one returns `artifact_compiled_dependency_forbidden`.

`CheckRejection` refuses a nil error, refuses a code outside the vector's closed set,
refuses any process start, and refuses any publication. The graph family tampers with
each path's real closed overlay — duplicate, dangling, wrong-kind, capture-replacing,
and missing-target bindings, with `insertSorted` keeping canonical order so the rule
under test fires rather than the ordering rule.

One measurement note, recorded because it is a narrower fact than "zero process
starts" reads: `ProcessStarts` is actually observed only on the SwiftPM path, which
wires `swiftpmsource.Config.ProcessStartObserver`. On the other five it is the zero
default rather than an instrumented count. The outcome document credits only
SwiftPM's observer, so nothing false is claimed, and the per-adapter zero-spawn
vectors (including Rust's pre-vendor zero-Cargo-spawn case) are owned by the accepted
adapter suites.

`opaque-dependency-bytes`, `verified-binary-unavailable` and `unreceipted-output` are
driven through the shared `artifactpolicy` service with each path's own declared
adapter and profile identity rather than through each adapter's capture pipeline.
That is the correct shape for the accepted C12 claim — the shared classifier owns
byte admission for every adapter — and the deny corpus is drawn from
`policyconformance.Cases()` with a `len(denyCases) < 20` floor so it cannot silently
shrink.

### 5. Honest delegation of the three non-runnable vectors — accepted; the honesty gate holds

I verified the sealing claims independently rather than accepting them:

- `artifactpolicy.LocalOutputAuthorization` is an interface whose only method is the
  unexported `artifactPolicyLocalOutputAuthorization()`, and `types.go` states the
  package "intentionally exposes no issuer". No package outside `internal/artifactpolicy`
  can construct one.
- `closureexec` binds `Audit.Network` to `"not-observed"` in portable mode and permits
  `"none"` only under `AssuranceVerified` with a provider-issued audit
  (`models.go:471`). A network-attempt observation therefore genuinely requires a
  live verified provider, exactly as the accepted assurance spec and `2qfnai`'s
  observed-read provider semantics require.

The named owners really prove the delegated codes:
`internal/closureexec/portable_runner_other_test.go:105` runs a script that actually
writes outside the declared set and requires `closure_write_undeclared`;
`internal/closureexec/closureexec_test.go:688,733` require both delegated execution
codes with zero receipts; `internal/swiftpmbuild/conformance_test.go:301,316`
reconcile them; `internal/npmsource/conformance_test.go:399` requires the network
code; `internal/artifactpolicy/policy_test.go` requires `CodeLocalOutputDrift` on
four paths. The compile-time constant references plus
`TestDelegatedVectorsNameRealOwningPackages` make a rename or deletion a build break
rather than a silently emptied matrix. The vectors stay published in the matrix and
in the export with their owners named. This is delegation, not a silent claim.

### 6. Integration only — no re-implementation, no widening — accepted

No accepted adapter production file was modified by this run. The producer run began
around 00:09 local and ended at 01:06; the only files whose mtimes fall inside that
window are `internal/crossconformance/*` (00:14-01:03), `README.md` and
`docs/source-closure-adapter-conformance.md` (00:50), and `LOGBOOK.md` (01:04). The
`internal/swiftpmsource/*` and `internal/closuregraph/*` deltas in the worktree are
22:16 and earlier — the previous tasks' uncommitted delivery, preserved as instructed.
Nothing was staged, committed, reset, or cleaned by this review either.

The package adds no `os/exec` use: it is absent from the resolved import set, and the
guard test forbids both `exec.Command` and the `os/exec` import while assembling its
needles at run time so it scans its own source as strictly. The cross-adapter process
guard allowlist is byte-identical to the one `internal/swiftpmbuild` and
`internal/swiftpminterop` already use — `{acquisition.go, portable_runner.go}` — and
is not widened. The canonical verifier is not forked: it is run as-is and its output
compared. `tkurtl` reject-by-default and `2qfnai` observed-read semantics are intact,
and both packages' suites are green in my own rerun.

### 7. Scope hygiene — accepted

Kotlin/Gradle/Maven appear exactly once, in the unsupported-ecosystems list, with
Dart, .NET, non-SwiftPM native systems and a new Python adapter. No Python file was
added anywhere in the delivered scope. Every new file is `text/plain` or
`application/json`; no vendored binary of any kind.

### 8. Evidence — accepted

Every attached log digest matches its board description byte for byte:

| Artifact | SHA-256 | Verified content |
| --- | --- | --- |
| `full-go-01.log` | `0b6f5468…e574b1d` | `EXIT:0`, 54 `ok` packages, 0 `FAIL`, `internal/crossconformance 35.632s` present |
| `full-suite-noncmd.log` | `59f31ade…22cc33716` | matches description |
| `full-suite-cmd.log` | `378e4ae5…d6f457b7aa` | `ok ./cmd/curator 316.568s` |
| `race-crossconformance.log` | `c5a41390…892275a7a` | `ok … 25.842s` |
| `crossconformance-verbose.log` | `efe3a050…f571aed7811` | matches description |
| `lint.log` | `e92606b0…59095c47` | `0 issues.` |
| `canonical-verifier.log` | `1847364d…28998b75a` | both accepted oracle lines |

The committed protocol export and its board copy are the same bytes
(`f4ff9e7703bc3a3cbfb3b96c3d23a8cac474d901e04707253795b2d934d6d298`), it is exact
CCJ-1, it names the accepted corpus digest, it carries 53 records that each still
derive their exported identity, and regeneration under `CURATOR_WRITE_CROSS_EXPORT=1`
deliberately fails so it cannot be mistaken for a pass.

## Documentation check

`docs/source-closure-adapter-conformance.md` publishes 80 stable diagnostic codes.
That set is **exactly** the accepted normative active code set from
`TASK-260810-1dgdos` — zero additions, zero omissions — and every one of the 80 has a
definition site outside `internal/crossconformance`. Supported profiles, unsupported
cases, precedence, the coverage and delegation tables, the environment requirement,
and the six migration steps all match tested behaviour. The `README.md` section and
gates row are consistent with it.

## Gates rerun by this reviewer

All commands were run as standalone processes in bounded foreground slices; nothing
was backgrounded and no exit code is inferred.

| Gate | Command | Result |
| --- | --- | --- |
| New package | `go test -count=1 -timeout 9m ./internal/crossconformance` | ok, 10.935s |
| Adapters and closure core, 13 packages | `go test -timeout 25m -count=1 ./internal/{crossconformance,closuregraph,closureexec,artifactpolicy/...,rustsource,npmsource,pnpmsource,yarnclassicsource,yarnmodernsource,swiftpmsource,swiftpminterop,swiftpmbuild,nodesource}` | exit 0, 1:43 wall |
| Remaining 41 non-`cmd/curator` packages | `go test -timeout 25m -count=1 $(rest)` | exit 0, 1:41 wall |
| CLI suite | `go test -timeout 9m -count=1 ./cmd/curator` | ok, 327.879s |
| Race, new package | `go test -count=1 -race -timeout 9m ./internal/crossconformance` | ok, 27.018s |
| Coverage gate, adversarial | `go test -run '…/coverage-is-complete'` | FAIL as required |
| Lint | `golangci-lint run` (v2.12.2) | `0 issues.` |
| Format | `gofmt -l cmd internal` | empty |
| Vet | `go vet ./...` | exit 0 |
| Whitespace | `git diff --check` | empty |
| Broad suppression | `bash .github/ci/no-broad-suppression.sh` | `no-broad-suppression: ok` |
| Accepted oracle | `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb …` | exit 0, both lines |
| Independent third re-hash | 53 records via an independent Python CCJ-1 implementation | `records=53 rehash_ok=53` |
| Board | `task-board --no-update-check validate` | `Board is valid. No issues found.` |

Combined, the two slices cover all 54 packages of the repository suite, so the whole
suite was re-run here rather than accepted from the producer's logs. The monolithic
`full-go-01.log` is additionally hash-verified above.

## The producer's two reviewer decisions

**1. The Rust path needs the pinned Cargo, and CI installs no Rust — accepted as a
pre-existing delivery condition, not rework.**

Verified independently. `.github/workflows/ci.yml` has no Rust, Node, pnpm, Yarn, or
Swift setup step. `internal/rustsource/authority_external_test.go:82-84` already calls
`rustsource.NewManager` and `t.Fatal`s on failure, and `.github/ci/skip-classes.tsv`
publishes no class that would tolerate an absent Cargo. The condition was therefore
already carried by the accepted, `done` Rust delivery; the cross-adapter suite
inherits it rather than introducing it. The producer's two rejected alternatives are
right: a host-conditional skip is exactly what `skip-classes.tsv` exists to catch, and
dropping the manager would leave Rust's causal-evidence obligation with no receipts at
all. This host satisfies the descriptor (`cargo 1.91.0 (ea2d97820 2025-10-10)`), which
is why the Rust cases are green here.

This is a delivery-level environment decision above this task — grow a CI Rust
toolchain step, or give `internal/rustsource` and the Rust cross cases a declared skip
class — and is recorded for the orchestrator. It does not block acceptance of this
task's scope.

**2. No `platform-cases.tsv` row — accepted.** No adapter package has one; that ledger
is scoped to compiled-build platform behaviour. Adding a row for the whole
cross-adapter case would encode the Rust-toolchain requirement into the per-runner
ledger, which is the wrong place to settle decision 1.

## Acceptance criteria mapping

| AC clause | Evidence |
| --- | --- |
| Every adapter passes the same normative semantic suite plus accepted ecosystem vectors | Six paths x seven obligations, completeness-gated; deny corpus and per-profile source vectors drawn from `policyconformance.Cases()` |
| All 53 records independently derive their exact hashes and all references resolve | Independent CCJ-1 oracle, third-party re-hash 53/53, 125 typed references, summary byte-identical to the accepted Ruby oracle |
| Darwin and Linux share one exact capture while selection, binding, active, plan, C4, C5 differ through explicit `targets` edges | `validateCGP05`; capture reuse of `sha256:1bcd31f3…c04f5f2` with 8 divergent kinds and 2 explicit target bindings |
| CGP10 branches keep every stable identity while only observation, execution, publication differ | `validateCGP10`; 22 stable records, exactly-once slots, distinct branch identities and observed bytes |
| Hidden, duplicate, dangling, wrong-kind, missing, capture-replacing bindings, cycles, drift, network, undeclared processes/inputs, compiled or opaque bytes, unreceipted outputs fail at the specified checkpoint with no later execution or publication | 16-vector matrix through the adapters' own seams plus 3 honestly delegated vectors; `CheckRejection` enforces code, zero process starts, no publication |
| Supported and unsupported documentation matches tested behaviour and full repository validation passes | `docs/source-closure-adapter-conformance.md` diagnostic set is exactly the accepted normative 80; live coverage matches its tables; all 54 packages green in this reviewer's own rerun |
| Export protocol goldens for independent Python without adding Python code | Committed canonical CCJ-1 export, zero Python files |
