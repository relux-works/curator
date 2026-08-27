# TASK-260720-3ag6pi — review cycle 3 verification outcome

## Outcome

The rc.6 protocol-v6 candidate satisfies the integrated conformance contract.
The review-cycle-2 defect is repaired: all 22 accepted compiled lifecycle cases
and their fail-closed validator/release guards are present on the current
`manager-worker-v1` identity. No additional product correction was needed
during this verification pass.

The original literal rc.4 release wording has been superseded by the reviewed
rc.5 resolution and the current rc.6 rework directive. The current rc.6 gate
passes. The literal rc.4 command is retained and reported truthfully as an
expected rejection, not as release evidence.

## Provenance and clean-candidate boundary

- Product source:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3ag6pi/rework-cycle-3`.
- Remote `main`:
  `f5d7673039226ab81de2f4f87e2155ae995c4df3`.
- Annotated `v1.0.0-rc.5` tag object:
  `3780f50847ce7f513436c950ec1656a8ab298185`; peeled commit:
  `f5d7673039226ab81de2f4f87e2155ae995c4df3`.
- Disposable verification commit:
  `ddb181ca3b8e243f212e90ff26fcabe2234fb669`.
- Verification parent:
  `f5d7673039226ab81de2f4f87e2155ae995c4df3`.
- Verification tree:
  `fa0eb87cacf6cd9ec8510f7d26a91ea76b6d74fd`.
- Verification signature state: `N`; tags at verification SHA: zero.
- Recursive product-source/candidate comparison: exit 0.

The verification commit is task-local and unsigned. It exists only because the
repository release gate correctly requires `HEAD` to identify a clean candidate
checkout. No product source index was staged, and no commit, push, tag,
implementation pin, claim, release, checksum, signature, or attestation was
added to an authoritative repository.

## Standalone gate results

Every command below ran as a standalone process without `tee`. Redirected logs
preserve raw stdout/stderr, and the exit codes below are the actual process
results.

| Command | Exit | Result |
| --- | ---: | --- |
| Recursive source/candidate byte comparison | 0 | Exact candidate parity |
| `/usr/bin/git ls-remote origin refs/heads/main ...` | 0 | Remote main/tag provenance above |
| `make validate` | 0 | 42 schemas, 447 vector files, 59 Python tests, Go tests; no skips |
| Fresh clone 1: `make regenerate` | 0 | No generated diff; empty porcelain status |
| Fresh clone 1: generated-path `git diff --exit-code` | 0 | Stable bytes |
| Fresh clone 2: `make regenerate` | 0 | No generated diff; empty porcelain status |
| Fresh clone 2: generated-path `git diff --exit-code` | 0 | Stable bytes |
| Recursive comparison of regenerated `conformance/v1` trees | 0 | Byte-identical |
| Comparison of independent aggregate digest logs | 0 | Both `8255898b37dd1f3b95423804bd0c35bd7ec48a16fbbe9b4d9e4cecc830900072` |
| `make regenerate-check` | 0 | Generator and tracked-output diff gate pass |
| `make release-check VERSION=1.0.0-rc.6` | 0 | Release gate passes at `ddb181ca…` |
| `make release-check VERSION=1.0.0-rc.4` | **2** | Expected red after supersession: README is rc.6 |
| `gofmt -l` on changed Go files | 0 | Empty output; follow-up empty-file assertion exits 0 |
| `go vet ./...` | 0 | Clean |
| `go test ./... -count=1` | 0 | Uncached Go tests pass |
| `git diff --check HEAD^ HEAD` | 0 | Clean |
| Manifest/index/inventory audit | 0 | Exact inventory and hashes |
| Legacy compatibility audit | 0 | Frozen bytes and semantics |
| Accepted/current lifecycle semantic parity audit | 0 | Exact after intentional identity normalization |
| No-execution/no-fabrication audit | 0 | Safety invariants pass |

The host Homebrew Python does not provide `jsonschema`; its readiness import
failed with exit 1. All Python gates therefore used the existing task-local
environment populated from the repository's pinned development requirements.
`PATH` included that venv plus standard tool locations; `git` resolved to
`/usr/bin/git`, not the historical task wrapper.

## Integrated evidence

- Manifest protocol: `1.0.0-rc.6`.
- Manifest entries: 447, sorted and unique.
- Manifest SHA-256:
  `72c5d717027ca096b14bc32f5d60bb740676974e9429f3d09b730897e5fba89b`.
- Both rc.6 release pins equal the manifest identity.
- Schema files: 42.
- Indexed schema cases: 376, exactly equal to case files on disk.
- Required case groups: agent-skill-v6 24, csk-skill-v6 24,
  build-receipt-v1 18, install-marker-v2 14, claim-v2 7, claim-v3 13.
- Go build fixture files: 13.
- Expected build-driver files: 11.
- Build-driver vectors: 8 positive, 77 rejection, 10 build-source identity,
  and 12 toolchain identity cases.
- Manager lifecycle: 32 named cases, including all 22 accepted compiled cases.
- Current manager lifecycle SHA-256:
  `25f573a40652dfaf6bb52818b2209c2f8c1e6aec2a7515023e0ac531c2d91289`.
- Accepted rc.4 lifecycle SHA-256:
  `676e617a0e0a6d575310f38e1de740eab583d709e2351be9eaa818c9882d78d4`.
- After normalizing only the intended execution-policy, cache-key, and receipt
  identity revision, accepted and current lifecycle semantics are
  byte-identical at normalized SHA-256
  `11fe66182e719e8e5c067d3ff2ab646dd889de7c97130881fcacdbf9694a7b6c`.

The current portable cache key is
`sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b`;
the current receipt identity is
`sha256:919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd`.
The superseded rc.4 identities remain negative/historical evidence and do not
alias the current candidate.

## Legacy compatibility

The following are byte-identical to both published rc.5 `origin/main` and the
pre-v6 baseline `57c1f56846d221ecc55786bd3c2467ec32f11730`:

- agent-skill schemas 1 through 5;
- csk-skill schemas 1 through 5;
- install-marker-v1;
- conformance-claim-v1;
- the 24 baseline valid/invalid schema cases for those contracts.

Their existing schema-case index and conformance-manifest entries are
preserved. The 70 added schema-7 guards for legacy manifest schemas are all
indexed invalid. Install marker v1 remains schema 1. Claim v1 remains schema 1
and protocol `1.0.0-rc.3`.

Published rc.5 metadata is byte-identical to `origin/main`, SHA-256
`75ae17fc029b4f51ca40ce768d04fd72991ec3db2602b8fe59213bee6ac34583`.

## Safety and publication honesty

- All 77 build-driver rejection cases require `result=reject`,
  `artifact_executed=false`, and `reuse=false`.
- Compiled dry-run forbids Go `list` and `build` and records
  `artifact_executed=false`.
- Generator and validator code contain no process-execution API.
- The release gate's only subprocess scope is Git.
- `GIT_INDEX_FILE`, `GIT_DIR`, and `GIT_WORK_TREE` are unset.
- rc.6 emits zero claims, advances no committed downstream pin, and claims no
  hardened execution profile.
- Published rc.5 bytes and remote refs remain unchanged.

The unchecked cross-platform, downstream-manager, signature, archive,
attestation, and publication boxes in `RELEASE.md` remain unchecked. This local
conformance verification does not manufacture those external artifacts.

## Code and test scope under review

The candidate carries the narrowly scoped cycle-3 implementation:

- restored generated lifecycle vector and current portable identity;
- generator logic and deterministic Go regression tests;
- fail-closed validator/release guards and Python negative tests;
- rc.6 manifest/release metadata plus compatible documentation updates;
- byte-frozen rc.5 metadata.

The detailed story-criterion and rejection-cluster mapping is attached as
`TASK-260720-3ag6pi_cycle-3-coverage-matrix.md`.
