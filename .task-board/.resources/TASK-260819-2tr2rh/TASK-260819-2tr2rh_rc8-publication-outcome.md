# TASK-260819-2tr2rh rc.8 publication outcome

## Published identity

- Version: `v1.0.0-rc.8`
- Normative candidate PR: https://github.com/relux-works/curator-spec/pull/20
- Reviewed release-target guard PR: https://github.com/relux-works/curator-spec/pull/21
- GitHub-verified squash commit: `f8c405aa3ad0a39d260c2ed93684e55c5a346359`
- Signed annotated tag object: `ad247840292487d5d88ac44331798b6b4182a79f`
- Tag target: `f8c405aa3ad0a39d260c2ed93684e55c5a346359`
- Release: https://github.com/relux-works/curator-spec/releases/tag/v1.0.0-rc.8
- Release workflow: https://github.com/relux-works/curator-spec/actions/runs/32202070660
- Post-merge Specification CI, including Release target provenance: https://github.com/relux-works/curator-spec/actions/runs/32201777861
- Post-merge Implementation conformance: https://github.com/relux-works/curator-spec/actions/runs/32201777858

GitHub reports both the squash commit and annotated tag signature as `verified=true`, `reason=valid`. The release is non-draft and prerelease.

## Immutable digests

- Conformance suite manifest `conformance/v1/manifest.json`: `d14e3a16bb4a01ff282791f08e3aefa269210234f41072beae6fe59b642595a1`
- Rc.8 metadata `release/1.0.0-rc.8.json`: `293f101d10665061aa049efa72141f9e3c5d608bbde300e882f6e3e095e31ede`
- `checksums.txt`: `5c91cb276cae1a12483e0fcc21d7c580697a1ce9af8d50d35a8ed3a9ffad5457`
- `curator-protocol-1.0.0-rc.8.tar.gz`: `b84f35703b8071a7db3d872ab7271eebf971ab72014f9a59b438c092b0927c83`
- `curator-protocol-1.0.0-rc.8.zip`: `489721fa24ee626e23d7432dcd667bc3ed39795fba40117d46aa4afeb8e42daf`

The downloaded archives passed `checksums.txt`, GitHub attestation verification for all three assets, tar/ZIP content equality, and comparison of every packaged path against the tagged checkout.

## Rc.7 immutability

- Rc.7 tag object remains `de704f2951e683d52ae8e475cb690b918a94d4c5`.
- Rc.7 target remains `99f70947d6f2447366d6c996127b73eca37a9159`.
- Rc.7 tag signature remains valid against `maintainers.allowed_signers`.
- Rc.7 metadata SHA-256 remains `e5872ee4dd207bf6b190d8c8be15a9366d9c1e3638047ea983620b97c9f84d5d`.

## Verified claim state

The published rc.8 metadata records:

- `assurance.verified_implementations = []`
- `assurance.verified_platform_claims = []`
- `claim_v4.claims_emitted = []`

No platform implementation conformance claim was published.

## Validation evidence

All required gates were run directly at detached exact commit `f8c405aa3ad0a39d260c2ed93684e55c5a346359` with generated Python bytecode disabled and exited 0:

- Repository merge-policy verifier.
- `tools/validate.py`: 49 schemas and 471 vector files.
- Python unit suite: 91 tests.
- `go test ./tools/...`.
- Deterministic vector regeneration and authoritative diff check.
- Go formatting and `git diff --check`.
- GitHub release-commit provenance verifier.
- `tools/release_gate.py --version 1.0.0-rc.8 --commit HEAD`.
- Clean-checkout test after all validation.
- Signed tag verification and remote tag/object resolution.
- Post-merge Specification CI, Release target provenance, Implementation conformance, and tag release workflow.
- Release checksum, package-content, and build-provenance attestation verification.

Two non-candidate environment invocations failed before their corrected reruns: `python` was absent on the host (exit 127; rerun with `python3` passed), and the first local release-commit verifier call omitted `GITHUB_REPOSITORY` (exit 1; rerun with the full workflow environment passed). System Python also lacked `jsonschema`; validation used an external temporary Python 3.12 virtual environment so the checkout remained clean.
