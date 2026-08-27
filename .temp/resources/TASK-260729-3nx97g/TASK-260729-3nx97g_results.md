# TASK-260729-3nx97g developer handoff evidence

Carry the independently accepted schema-6 build-driver golden suite forward
into the exact accepted rc.5 `TASK-260728-2kp3tv` candidate under
`execution_policy=manager-worker-v1`. Candidate-only. Nothing staged,
committed, tagged, pinned, published, or claimed.

## 1. Workspace and provenance

- Assigned worktree:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree`
- Base: fresh detached `git worktree add --detach` from the pinned
  `curator-spec` base `57c1f56846d221ecc55786bd3c2467ec32f11730`.
- Candidate import: byte copy of the accepted `TASK-260728-2kp3tv` rc.5 state
  (`rsync -a --delete --exclude=.git`). Immediately after import,
  `diff -r --exclude=.git` between source and destination exited 0, and
  `shasum -a 256 conformance/v1/manifest.json` was
  `9ba9b8ecf6f06cafda1425aed3539dee8d12af43dabd36f803f4f420dbb1dacf` on both
  sides — exactly the accepted rc.5 pin.
- The board's recorded provenance digest was independently reproduced with the
  same normalized-tree recipe recorded by `TASK-260729-1kq1rd` (path, entry
  kind, mode and content digest per row; `.git`, `.temp`, `__pycache__` and
  `.pytest_cache` excluded). The source candidate hashes to
  `3e4fd26acd9cafd1a76b2b5312da49ee35d234738263beb17a42be971d9dc582` over 514
  files, matching the recorded value exactly. This candidate is 539 files
  (514 + the 25 added golden files) at
  `38a75faad32e50db11824864e1d92b1fd7503218a59fd71b414feac6568de0a2`. Script
  kept at `.temp/TASK-260729-3nx97g/normalized-tree.js`.
- Golden-suite import: `conformance/v1/fixtures/go-build-skill` copied byte for
  byte from the accepted `TASK-260720-1s1vr6` worktree (`diff -r` exit 0, 13
  files). Generator/expected-artifact semantics were taken from the exact
  `TASK-260720-37ei85 -> TASK-260720-1s1vr6` delta, which is confined to the
  fixture, `expected/build-driver/`, `vectors/build-drivers.json`, the manifest
  inventory, and `tools/generate-vectors/{main.go,main_test.go}`.
- Cross-check: the rc.4 canonical snapshot `TASK-260720-q5oy3o` publishes a
  byte-identical fixture, `expected/build-driver/` tree and
  `build-drivers.json` (`fd613bbb…`), so `1s1vr6` is the accepted semantic
  source for this carry-forward.
- Source worktrees were read only. Re-checked after all work: `2kp3tv` still at
  HEAD `57c1f568…` with 127 uncommitted paths and manifest pin `9ba9b8ec…`;
  `TASK-260720-1ljev5` (Curator source) still at `17804cea…` with its own 89
  uncommitted paths.
- Assigned worktree after all work: HEAD still `57c1f568…`,
  `git rev-list --count <pin>..HEAD` = 0, `git diff --cached --quiet` exit 0,
  zero staged paths.
- Disposable scratch: `.temp/TASK-260729-3nx97g/release-probe` (clean git repo
  used only to exercise the real clean-checkout gates; synthetic commit
  `a4ac2d64d790ec560c4d8028974573fa5f3c7585` is not a protocol, implementation
  or downstream pin) and `.temp/TASK-260729-3nx97g/curator-probe` (read-only
  copy of the accepted Curator source used only to run the candidate tests; no
  Curator file was edited).

## 2. Exact rc.5 candidate delta

Six files modified, three trees added, none removed:

| Path | Change |
| --- | --- |
| `conformance/v1/fixtures/go-build-skill/` | added (13 files, imported byte-exact) |
| `conformance/v1/expected/build-driver/` | added (11 generated artifacts) |
| `conformance/v1/vectors/build-drivers.json` | added (generated) |
| `conformance/v1/manifest.json` | regenerated inventory |
| `release/1.0.0-rc.5.json` | candidate pin only (both fields) |
| `tools/generate-vectors/main.go` | build-driver emitter + build-root context exclusion |
| `tools/generate-vectors/main_test.go` | 9 build-driver tests |
| `tools/validate.py` | `validate_build_driver_vectors` and two helpers, wired into `validate_vector_semantics` |
| `tools/test_validate.py` | `BuildDriverGoldenSuiteTests`, 12 tests |

Manifest inventory: **422 → 447** files. 25 added, **0 removed, 0 changed** —
every one of the 422 accepted rc.5 entries keeps its exact digest.

## 3. The identity carry-forward

The rc.4 golden input and the rc.5 golden input differ by exactly one required
field. `validGoBuildInputV1()` in the rc.5 candidate already carries
`policy.execution_policy = "manager-worker-v1"`, so regenerating the accepted
rc.4 expected artifacts against the rc.5 generator produces the required
identities without any hand-written constant:

| Identity | rc.4 (accepted) | rc.5 (this candidate) |
| --- | --- | --- |
| portable cache key | `sha256:3fcd714a…` | `sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b` |
| stored receipt hash | `sha256:750f5f75…` | `sha256:919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd` |
| fixture build-source | `sha256:27cdcac0…` | `sha256:27cdcac0…` (unchanged) |
| toolchain digest | `sha256:baf7c5f3…` | `sha256:baf7c5f3…` (unchanged) |
| fixture context hash | `sha256:82c0a18e…` | `sha256:82c0a18e…` (unchanged) |

Both new values were recomputed independently in a separate Python process
directly from the published candidate bytes, before the generator was written
and again after every gate:

- `SHA-256(expected/build-driver/build-input.ccj.json)` (869 bytes) =
  `sha256:529370…`, equal to `cache-key.txt`.
- `SHA-256(expected/build-driver/receipt.ccj.json)` (1120 bytes) =
  `sha256:919fbbad…`, equal to `receipt-sha256.txt`.
- Both files are exact CCJ-1 bytes with no terminal newline, the receipt binds
  its own key and its own input, and `policy.execution_policy` is
  `manager-worker-v1`.

The build-source, toolchain and context identities do not move because none of
them takes the execution policy as input — which is what proves the revision is
scoped to the build input.

## 4. Legacy and reserved-hardened as explicit non-alias negatives

`build-drivers.json` gains a `cache_identity` block that mirrors the accepted
`go-host-execution-policy.json` shape. Each key is the CCJ-1 digest of its own
stored input, all three are distinct, and `aliases` is `false`:

| Entry | `execution_policy` | Cache key | `schema_valid` |
| --- | --- | --- | --- |
| `portable` | `manager-worker-v1` | `sha256:529370…` | `true` |
| `reserved_hardened` | `hardened-worker-v1` | `sha256:13736230d33ce59de7f7323dcd4cffd510655ad8dabd5ee9e8b6cb182ec70037` | `false` |
| `legacy_rc4_without_execution_policy` | `null` | `sha256:3fcd714a40e8918eb67dbd35d435875dcce6c9047da811a1fa26626e5e57be48` | `false` |

The two negatives also appear as named rejection cases —
`legacy-rc4-input-without-execution-policy` (`build_input_invalid`) and
`reserved-hardened-execution-policy`
(`build_execution_policy_unsupported`) — each carrying its own
`derived_cache_key` and an explicit
`schema_valid=false, aliases_portable_cache_key=false,
cache_lookup_performed=false` outcome. `validate.py` proves the verdicts
against the **real compiled** `build-receipt-v1` schema rather than asserting
them: the portable input validates, the hardened input fails the
`goExecutionPolicyV1` const, and the legacy input fails the `execution_policy`
requirement.

## 5. Cluster coverage: preserved or explicitly superseded

| Cluster | rc.4 | rc.5 | Dropped |
| --- | --- | --- | --- |
| positive cases | 7 | 8 | 0 |
| rejection cases | 75 | 77 | 0 |
| build-source byte edges | 10 | 10 | 0 |
| toolchain byte edges | 12 | 12 | 0 |

Added: positive `portable-execution-policy-is-required-input` and the two
execution-policy rejections above. Every rc.4 case name survives verbatim.

One identity is **explicitly superseded**, and only one:
`self-consistent-forged-receipt-outside-protected-state`. Its whole point is a
receipt whose internal checks all pass, so it embeds the portable input and its
digest necessarily moves with the revision:
`sha256:9a23f5b7…` → `sha256:e15a8b198ddc4b9892747af3fc070713e72b72ad121512f7bdd5919d3581bd6d`.
The vector stays internally self-consistent (its `cache_key` is the CCJ-1
digest of its own input; its `receipt_sha256` is the digest of its own canonical
bytes), and both the Go test and the Python validator assert that, plus assert
the superseded rc.4 digest is no longer reproduced. Pinning the old digest
would have required a receipt that lies about its own key.

## 6. Generator change, and why the context selector moved

`selectedContextFiles` now also excludes declared `build_roots`, matching the
accepted rc.4 semantics. That is a shared helper, so the scope was proved
rather than assumed: `TestBuildDriverContextSelectionExcludesOnlyDeclaredBuildRoots`
runs the selector over both fixtures and asserts the build fixture yields
exactly `["SKILL.md", "assets/prompt.md"]` while the script-only fixture's
selection and `expected/context_sha256.txt` are unchanged. The 422 pre-existing
manifest digests being byte-identical is the independent confirmation.

## 7. Tests added

Go (`tools/generate-vectors/main_test.go`), 9 tests:

- `TestGeneratedGoBuildFixtureContextAndIdentity` — fixture root, schema-6
  build root, mixed build/script command shapes, transitive embed and vendored
  dependency, domain-prefixed preimage, root marker inside the preimage.
- `TestBuildDriverPortableIdentityIsByteExact` — execution policy on the
  identity and inside the input, CCJ-1 round trip, digest of the input, exact
  bytes of all four `expected/build-driver` identity files, no terminal
  newline, artifact fields, marker binding.
- `TestBuildDriverCacheIdentityMissesInsteadOfAliasing` — three distinct
  self-derived keys, per-entry `schema_valid`, deferred hardened owner, and the
  two matching named rejections.
- `TestBuildDriverPositiveProcessCacheAndDryRunCoverage` — exact positive set,
  five argv forms, fixed environment, forbidden inherited variables, and no
  source-aware Go command on cache hit or dry run.
- `TestBuildDriverRejectionCoverageAndNamedOutcomes` — exact 77-name set, every
  case non-executing with a named error, forged-receipt supersession.
- `TestBuildSourceAndToolchainByteVectors` — exact byte-edge sets and pinned
  digests including the legacy NUL-stream collision.
- `TestBuildDriverGenerationIsDeterministic` — two independent generations into
  temp roots are byte-identical.
- `TestBuildDriverGenerationPreservesScriptFixtureAndRegistryGoldens` — 15
  frozen script-fixture and registry digests, all still the accepted rc.4
  values.
- `TestBuildDriverContextSelectionExcludesOnlyDeclaredBuildRoots` — see §6.

Python (`tools/test_validate.py`, class `BuildDriverGoldenSuiteTests`), 12
tests: the happy path with the two AC identities; the compiled-schema proof for
all three inputs; and eight guard-fires-when tests covering a negative claiming
validity, a negative aliasing the portable key, declared aliasing, each of the
three explicitness fields, an executing rejection, a dropped positive /
build-source / toolchain cluster, a forged receipt losing self-consistency, and
a drifted byte-edge digest. Also two byte-level checks: the build-source
preimage frames the fixture on disk, and no declared build root reaches the
agent context.

Test totals: Python **29 → 41**, Go generator package all pass.

## 8. Guards proved to fire (negative probes)

Mutating `build-drivers.json` alone trips the manifest digest check first, so
each probe also re-stamped the manifest and release pin to reach the semantic
guard. All four were expected red and were red for the intended reason, exit 1:

| Probe | Message |
| --- | --- |
| portable cache key silently reverted to the rc.4 value | `build-driver cache key is not SHA-256(CCJ-1(input))` |
| reserved-hardened negative claims schema validity | `build-driver cache identity misreports reserved_hardened` |
| legacy negative aliases the portable key | `legacy_rc4_without_execution_policy aliases portable` |
| a prior positive cluster silently dropped | `build-driver positive coverage changed` |

The three earlier bare mutations (no re-stamp) each failed with
`vector digest mismatch for vectors/build-drivers.json`, exit 1 — also correct,
just a different guard. After every probe the baseline was restored and
`python tools/validate.py` returned exit 0 with the manifest back at
`b6f56aac…`.

## 9. Preserved invariants

All nine rc.4 byte-frozen artifacts re-verified after every edit and after the
gates, digests unchanged: `agent-skill-v6` `982832e4…`, `csk-skill-v6`
`2148eafc…`, `build-receipt-v1` `f673a881…`, `install-marker-v2` `6d7b65db…`,
`conformance-claim-v2` `4c05a97a…`, and their four `valid.json` cases
`cf029927…`, `cf029927…`, `538d12bb…`, `f7e7cc86…`.

`diff -r schemas/` against the accepted rc.5 candidate exits 0: schema 1
through schema 7 declaration bytes are untouched. The frozen
`curator-build-source-v1` algorithm namespace is preserved and still present;
`TestRepositoryDescriptorIsManagerNeutral` and the `validate.py` absence guard
both still pass over the enlarged tree, so no new surface names the retired
descriptor.

No manifest, descriptor, receipt, marker or claim field was added, removed or
widened. No generic language driver, fallback, package-selected program, argv,
environment, output, credential, helper, filter, signer, hook or plugin. No
alias and no dual-read path.

## 10. Exact rc.5 candidate identity after the carry-forward

- Protocol version: `1.0.0-rc.5` (unchanged)
- New candidate protocol pin (SHA-256 of `conformance/v1/manifest.json`):
  `sha256:b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`
  (predecessor was `sha256:9ba9b8ec…`)
- `release/1.0.0-rc.5.json` differs from the predecessor in exactly two lines:
  `candidate_protocol_pin.manifest_sha256` and
  `downstream_consumption.required_manifest_sha256`. Recompute with
  `shasum -a 256 conformance/v1/manifest.json`.
- `conformance/v1/vectors/build-drivers.json` SHA-256:
  `f412c107091cf82f980523afe5361212a3b89a3425f5d885373191f8acb12aea`
- Aggregate over every file under `conformance/v1` and `release/`:
  `6c989066e21cc304890e76df877b388b51fadabf861ddd5e4fed3cac7d201ea8`
- Suite inventory: 42 schemas, 447 manifest-listed files, 25 of them
  build-driver.
- `committed_release_pin_advanced` remains `false`; `claim_v3.claims_emitted`
  remains `[]`; macOS and Windows remain pending downstream native evidence;
  Linux remains excluded until `TASK-260728-1skseh`. The execution-policy block
  is untouched.

## 11. Gates and exact exit codes

Every command was run directly as a standalone process, not through `tee` or a
pipe chain.

Assigned worktree:

| Command | Exit |
| --- | --- |
| `go run ./tools/generate-vectors -root .` | 0 |
| `python tools/validate.py` (task venv) — `validated 42 schemas and 447 vector files` | 0 |
| `python -B -m unittest discover -s tools -p 'test_*.py'` — 41 tests, OK | 0 |
| `go test ./tools/...` | 0 |
| `go vet ./tools/...` | 0 |
| `test -z "$(gofmt -l tools)"` | 0 |
| `python -m compileall -q tools` | 0 |
| `git diff --check` | 0 |
| `git diff --cached --quiet` | 0 |
| `git rev-list --count 57c1f568..HEAD` = 0 | 0 |

Deterministic regeneration in the assigned worktree: two further consecutive
`go run ./tools/generate-vectors -root .` runs left the aggregate over
`conformance/v1` and `release/` at
`6c989066e21cc304890e76df877b388b51fadabf861ddd5e4fed3cac7d201ea8`, identical
to the pre-run value.

Clean disposable release probe at commit `a4ac2d64…`, byte-identical to the
assigned worktree before and after the gates (`diff -r --exclude=.git` exit 0
both times):

| Command | Exit |
| --- | --- |
| `make regenerate-check`, first run | 0 |
| `make regenerate-check`, second consecutive run | 0 |
| `make release-check VERSION=1.0.0-rc.5` — validation, Python, Go, regeneration, exact metadata pin, execution-policy honesty, descriptor identity, clean checkout, version, candidate gate | 0 |
| `git status --porcelain` after the gates — zero lines | 0 |

Curator candidate metadata artifacts, run from a read-only copy of the accepted
Curator source with an explicit conformance root:

| Root | `TestCandidateBuildMetadataArtifacts` | `TestCandidateBuildReceiptSchemaCase` | Exit |
| --- | --- | --- | --- |
| accepted rc.5 `TASK-260728-2kp3tv` (baseline) | **SKIP** — `publishes no expected/build-driver artifacts` | PASS | 0 |
| this candidate | **PASS** | PASS | 0 |

Full package against this candidate root:
`CURATOR_CONFORMANCE_ROOT=<candidate>/conformance/v1 go test ./internal/buildmeta/...`
exit 0, **12 top-level PASS and 0 SKIP**. The silent coverage hole recorded in
the 2026-07-29 0455 logbook entry is closed: a candidate root now exercises
byte-exact build-driver goldens under rc.5 semantics.

Expected-red gates, reported truthfully as failures: the seven negative probes
in §8, each exit 1 and each red for its intended reason.

Not run, and why: no native macOS or Windows manager execution was exercised.
This task is specification-only and adds no manager implementation, so there is
no worker binary to run — the candidate still emits zero claim-v3 tuples. The
physical `go build` of the fixture recorded by `TASK-260720-1s1vr6` was not
repeated; the fixture bytes were imported unchanged and its build-source
identity `sha256:27cdcac0…` was recomputed here from those bytes and matched.

## 12. Boundaries preserved

- No Curator or CocoaSkills product edit. `internal/buildmeta` was executed
  from an unmodified read-only copy; the accepted Curator worktree it was
  copied from is unchanged at `17804cea…`.
- Nothing staged, committed, tagged, pushed or published. No pin advancement,
  no release claim, no platform claim.
- No source repository, predecessor worktree, release, ref, tag or downstream
  pin was modified.
- Evidence is candidate-only and authorizes no landing or publication.
