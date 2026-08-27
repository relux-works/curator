# TASK-260720-3ag6pi review cycle 3 verdict

## Verdict

CHANGES REQUESTED → `to-dev`.

The cycle-2 regression is repaired. The exact rc.6 candidate restores every
accepted schema-6 compiled manager lifecycle case on the current portable
`manager-worker-v1` identity, restores fail-closed validator and release
guards, preserves the frozen legacy contracts, and passes the repository-local
gates independently.

Acceptance is still blocked by one ordinary implementation defect: both
GitHub Actions regeneration gates omit generated
`release/1.0.0-rc.6.json` from their diff scope. They can therefore pass while
the checked-in rc.6 metadata is stale relative to the generator. This is
recoverable rework, not an external blocker or human-only decision.

## Version and publication boundary

The board's reviewed resolution superseded the original literal rc.4 release
wording with rc.5. Rc.5 is now published and immutable, so the lifecycle repair
is correctly represented as rc.6 rather than as a rewrite of rc.5.

- Literal `make release-check VERSION=1.0.0-rc.4` exits 2 because the candidate
  identifies rc.6. This is truthful expected-red evidence, not a passing claim.
- `make release-check VERSION=1.0.0-rc.6` passes unwrapped.
- Published `release/1.0.0-rc.5.json` remains byte-identical at
  `sha256:75ae17fc029b4f51ca40ce768d04fd72991ec3db2602b8fe59213bee6ac34583`.
- Rc.6 defines no claim schema and emits no claim.

This verdict records the exact candidate bytes and passing local candidate-gate
evidence. It does not accept delivery or claim that rc.6 has been pushed,
tagged, signed, attested, or published.

## Exact reviewed candidate

- Product source:
  `.temp/TASK-260720-3ag6pi/rework-cycle-3`
- Clean task-local verification commit:
  `ddb181ca3b8e243f212e90ff26fcabe2234fb669`
- Sole parent:
  `f5d7673039226ab81de2f4f87e2155ae995c4df3`
- Git tree:
  `fa0eb87cacf6cd9ec8510f7d26a91ea76b6d74fd`
- Signature state: `N`
- Tags and remote refs at the verification commit: none
- Recursive product-source/verification-candidate comparison: byte-identical
- Gate Git executable: `/usr/bin/git`
- `GIT_INDEX_FILE`, `GIT_DIR`, and `GIT_WORK_TREE`: unset

The current remote `main` and the peeled signed annotated
`v1.0.0-rc.5` tag both resolve to the candidate parent
`f5d7673039226ab81de2f4f87e2155ae995c4df3`.

## Independent gates

| Gate | Result |
| --- | --- |
| `make validate` | PASS: 42 schemas, 447 conformance files, 59 Python tests, and Go tests |
| Skip audit | PASS: no Go/Python skip or xfail declarations and no skipped result |
| Fresh regeneration 1 | PASS: empty status and aggregate SHA-256 `8255898b37dd1f3b95423804bd0c35bd7ec48a16fbbe9b4d9e4cecc830900072` |
| Fresh regeneration 2 | PASS: empty status and the same aggregate SHA-256 |
| Recursive regenerated-tree comparison | PASS: byte-identical `conformance/v1` and rc.6 metadata |
| `make regenerate-check` | PASS |
| `make release-check VERSION=1.0.0-rc.6` | PASS at exact clean candidate commit |
| Literal rc.4 release check | Expected red, exit 2; not reported as passing |
| `gofmt` | PASS |
| `go vet ./...` | PASS |
| `go test ./... -count=1` | PASS |
| `git diff --check HEAD^ HEAD` | PASS |
| `.github/workflows/ci.yml` deterministic regeneration | **FAIL:** line 56 scopes diff only to `conformance/v1` and rc.5 metadata |
| `.github/workflows/release.yml` deterministic regeneration | **FAIL:** line 49 scopes diff only to `conformance/v1` and rc.5 metadata |

## Acceptance-blocking workflow drift gap

The Makefile was correctly updated:

```text
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json
```

Both GitHub Actions workflows still run the stale command:

```text
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json
```

An independent negative reproduction changed only the generator's rc.6
`created_at` output and regenerated:

- the exact workflow diff command exited 0;
- `release/1.0.0-rc.6.json` differed from the committed candidate;
- a diff scoped to rc.6 exited 1.

Thus the CI and tagged-release workflows do not enforce the same generated
inventory as `make regenerate-check`. A generator change can drift rc.6
metadata without failing either remote deterministic-regeneration gate. This
violates the task requirements that all release gates pass and that generated
release evidence remain exact.

Required rework:

1. Add `release/1.0.0-rc.6.json` to the deterministic-regeneration diff command
   in both `.github/workflows/ci.yml` and `.github/workflows/release.yml`, or
   invoke the canonical `make regenerate-check` target from both workflows.
2. Add an executable regression assertion that the CI/release workflow scope
   cannot diverge from the Makefile's generated-file inventory.
3. Rerun validate, two independent regenerations, regenerate-check, rc.6
   release-check, and the workflow-drift negative reproduction; attach the new
   evidence for another reviewer cycle.

## Cycle-2 regression closure

The current manager lifecycle vector has 32 unique named cases. All 22
previously missing compiled lifecycle cases are restored:

- planning and provider/command ordering;
- read-only compiled dry-run;
- private-build isolation;
- protected cache publication, race, corruption, and trust handling;
- cross-project isolation;
- deterministic lock/commit/rollback transactions;
- transaction-ID recovery and post-build recovery;
- currentness, repair, and locked garbage collection.

After removing only the intended `manager-worker-v1` execution-policy field and
substituting its derived cache-key and receipt identities with the accepted
pre-policy identities, the current lifecycle JSON is byte-identical to the
accepted `TASK-260720-cw39jh` vector. Both normalized documents hash to
`676e617a0e0a6d575310f38e1de740eab583d709e2351be9eaa818c9882d78d4`.

Focused negative tests prove that each of the 22 cases, every lifecycle group,
the schema version, and the current portable identity fail closed when removed
or changed. Release tests also reject missing required artifacts, renamed v6
schemas, refreshed-hash lifecycle deletion, stale compiled identity, altered
claim history, modified published rc.5 metadata, duplicate/stale rc.6 pins,
and fabricated rc.6 claims.

## Inventory and legacy compatibility

- Manifest: 447 sorted unique entries; inventory is exact and every SHA-256
  matches.
- Schema-case index: 376 unique entries; index paths exactly equal case files.
- Required groups: agent-skill-v6 24, csk-skill-v6 24,
  build-receipt-v1 18, install-marker-v2 14, claim-v2 7, claim-v3 13.
- Go build fixture: 13 files.
- Expected build-driver output: 11 files.
- Build-driver vectors: 8 positive, 77 rejection, 10 build-source identity,
  and 12 toolchain identity cases.
- Manager lifecycle: 32 unique named cases.
- Both rc.6 release pins equal the exact manifest identity.

The 12 required legacy schema files and their 24 baseline valid/invalid cases
are byte-identical to both the pre-v6 baseline
`57c1f56846d221ecc55786bd3c2467ec32f11730` and published rc.5 `main`.
Their 24 schema-index rows and 24 manifest entries retain the same semantics
and hashes. The 70 later schema-7 guard cases for manifest schemas 1 through 5
remain indexed invalid.

This covers:

- agent-skill schemas 1 through 5;
- csk-skill schemas 1 through 5;
- install-marker-v1;
- conformance-claim-v1.

## Story acceptance matrix

| STORY-260720-35dck7 criterion | Result | Executable evidence |
| --- | --- | --- |
| New schema version validates agreed build declarations | PASS | V6 canonical/legacy schema cases, receipt-v1 and marker-v2 cases, and build-driver structural positives/rejections all execute under `make validate` |
| Go driver semantics and install ordering are normative | PASS | Decision 0004 and manager profile remain unchanged; lifecycle tests enforce audit-first, provider-first, deterministic lock/target, consumer-last, and reverse rollback ordering |
| Build sources are excluded from agent context | PASS | Positive context-exclusion vector plus build-root-content and marker-embedding rejection vectors |
| Dry-run and audit-before-build are explicit | PASS | Compiler-free and compiled-cache-miss dry-run cases plus all-source-and-trust-gates-before-build |
| Compatibility and security impact are recorded | PASS | Updated rc.6 docs plus byte/index/manifest legacy comparison |
| Valid builds and key rejection cases are covered | PASS | 8 positive, 77 rejection, 10 build-source, 12 toolchain, and 32 lifecycle cases |
| Validation and deterministic regeneration pass | **FAIL integrated** | Local commands and two clean clones pass, but CI and tagged-release workflow diff scopes permit rc.6 metadata drift |

## Minimum rejection-cluster matrix

| Cluster | Passing evidence |
| --- | --- |
| Structural manifest | V6 invalid cases; forbidden args/env/output/toolchain/hooks and mixed-shape rejections |
| Build-root/source/context paths | Missing, unused, overlapping, root, escaped, outside, link, special, non-directory, context-leak, and marker-embedding cases |
| Build-source identity | All 10 framing, path, metadata, mutation, collision, and root-marker cases |
| Toolchain identity/release boundary | All 12 toolchain cases plus switch, family, executable, and digest rejections |
| Module/dependency/compiler graph | Module, package, vendor, workspace, cgo/native/SWIG/syso/assembly/embed/generate/PGO rejections |
| Process/host isolation | PATH, GOFLAGS/GOENV/GOWORK, VCS, fake-Go, telemetry, external-link, libgcc, and child-tool cases |
| Cache/receipt/protected trust | Cache identity, receipt/artifact, canonical form, link/special, race, forged receipt, and protected-state cases |
| Cache-hit/dry-run/marker-context | Protected hit, compiler-free miss, compiled read-only miss, marker embedding, and context leakage |
| Claim transition/stale suite | Frozen claim v1/v2 identities, claim-v3 rc.5 transition, no rc.6 claim, stale/duplicate pin rejection |
| Private build/cache publication | Private staging, second-build isolation, protected atomic publication, race, corruption, and trust cases |
| Commit/swap/rollback | Deterministic locks, target order/consumer-last, and reverse locked rollback |
| Concurrent projects/recovery | Two-project success, isolated rollback, transaction-ID recovery, and post-build recovery |
| Currentness/repair/GC | Current/failure matrix, rebuild-only repair, locked mark/sweep, and post-commit warning |
| Fail-closed inventory/release | **FAIL integrated:** local missing/renamed/lifecycle/fixture/history/pin/claim negatives pass, but both remote workflow regeneration commands ignore rc.6 metadata drift |

## Safety and reviewer boundary

Every build-driver rejection requires `result=reject`,
`artifact_executed=false`, and `reuse=false`. Compiled dry-run forbids Go
`list` and `build` and records no artifact execution. The generator and
validator have no process-execution API; release-gate subprocess use is
Git-only.

The reviewer changed no product code and performed no stage, commit, push, tag,
release, claim, pin, signature, checksum, or attestation operation. No
`commit_ack` was supplied.
