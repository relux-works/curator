# Final rc.8 publication review verdict

Verdict: accepted.

The published v1.0.0-rc.8 release satisfies the task acceptance criteria and Definition of Done. No code, architecture, release-process, metadata, signature, asset, immutability, or conformance-claim defect was found.

Independent identity and provenance checks:
- PR 20 is merged and its accepted candidate tree exactly matches the rebase result tree at 792c53c1887ce02b4b9c1d3954312c919ffb62ef.
- PR 21 is merged by GitHub squash as f8c405aa3ad0a39d260c2ed93684e55c5a346359; remote main points to that commit and GitHub reports verification valid.
- Signed annotated v1.0.0-rc.8 tag object ad247840292487d5d88ac44331798b6b4182a79f targets f8c405aa3ad0a39d260c2ed93684e55c5a346359. GitHub and local git verify-tag report a valid signature.
- Rc.7 remains immutable: tag object de704f2951e683d52ae8e475cb690b918a94d4c5 targets 99f70947d6f2447366d6c996127b73eca37a9159 and remains signature-valid.

Independent CI and release checks:
- Post-merge Specification CI 32201777861 succeeded at the exact squash commit, including Release target provenance, formatting, links, and all platform specification jobs.
- Post-merge Implementation conformance 32201777858 succeeded on Ubuntu, macOS, and Windows.
- Release workflow 32202070660 succeeded and published a non-draft GitHub prerelease.
- Exact tagged-tree clean-clone checks passed: 49 schemas and 471 vectors validated; 91 Python tests passed; Go tests passed; deterministic regeneration, gofmt, whitespace, merge-policy verification, GitHub release-commit verification, rc.8 release gate, and clean-tree check passed.

Independent artifact and metadata checks:
- conformance/v1/manifest.json SHA-256: d14e3a16bb4a01ff282791f08e3aefa269210234f41072beae6fe59b642595a1.
- release/1.0.0-rc.8.json SHA-256: 293f101d10665061aa049efa72141f9e3c5d608bbde300e882f6e3e095e31ede.
- Fresh release downloads match checksums: checksums.txt 5c91cb276cae1a12483e0fcc21d7c580697a1ce9af8d50d35a8ed3a9ffad5457; tar b84f35703b8071a7db3d872ab7271eebf971ab72014f9a59b438c092b0927c83; zip 489721fa24ee626e23d7432dcd667bc3ed39795fba40117d46aa4afeb8e42daf. GitHub build-provenance attestations verified for all three assets.
- Rc.7 metadata SHA-256 remains e5872ee4dd207bf6b190d8c8be15a9366d9c1e3638047ea983620b97c9f84d5d.
- assurance.verified_implementations, assurance.verified_platform_claims, and claim_v4.claims_emitted are all empty; no verified platform implementation claim was published.

Two isolated-clone verifier invocations initially stopped on reviewer environment setup only: GITHUB_TOKEN was absent, then origin/main pointed at the source clone stale ref. With the authenticated token and authoritative remote main fetched, both verifiers passed. These were not release defects and the checkout remained clean.