# TASK-260924-19n6g2 review verdict — CR rev1 — ACCEPTED
Reviewer: claude-opus-5-5 (low). Candidate tree 720b7e2b verified in disposable clone (base 2c39c428 + read-tree of candidate; tree OID matched).

## Precedent (note 1)
rc.10/rc.11/rc.12 had NO release-prep commits: tags sit on feature commits (b8b03d5, 87a0d00, dced9b8) and the generator kept rewriting release/1.0.0-rc.9.json. Their release workflows all FAILED (gh runs 33012127813, 34155418640, 35338455580); rc.9 was the last green release. So "exactly like rc.10–12" cannot yield a green release. This candidate's changes to the gate, validator, generator, Makefile, ci.yml and release.yml (advancing PROTOCOL_VERSION, adding release/1.0.0-rc.13.json, freezing rc.9) follow the rc.8→rc.9 precedent. That is the justified deviation.

## rc.9.json restore (note 2)
Candidate rc.9.json == tag v1.0.0-rc.9 bytes (git diff --quiet: identical). #88's change to that file was only the generator re-pinning the moving core-manifest digest (cb7a98… from main). That rolling pin now lives in release/1.0.0-rc.13.json (be11bb…). Nothing from #88's normative erratum is lost. Released metadata is now frozen by digest in both validate.py and release_gate.py.

## skillfile-sources-v1 suite (note 3)
- The generator writes the manifest, and regeneration is clean (go run ./tools/generate-vectors, then git status empty).
- rc.10 core manifest sha256 = 803918bf… (checked with `git show v1.0.0-rc.10:conformance/v1/manifest.json | shasum`) and matches compatible_core.
- The manifest covers protocol/skillfile-sources.md, protocol/repository-transport.md, docs/skillfile-sources.md, schemas/skillfile-sources-v1 and conformance/skillfile-sources-v1.

## Vectors / CHANGELOG (note 4)
- The vector diff is only the protocol_version line, rc.9 → rc.13, in 18 lines.
- CHANGELOG renames Unreleased to "1.0.0-rc.13 - 2026-09-26" and keeps every entry. It adds text on the independent pin, the rc.10 consumability, no v9 `directory`, per-claim Implementations CI, and the #88 erratum pin.

## Gates (note 5), real exit codes, zsh + pipefail
- validate.py: exit 0 (64 schemas, 1169 vector files).
- release_gate.py --version 1.0.0-rc.13: exit 0. With --version 1.0.0-rc.12: exit 1, as expected.
- go test ./tools/generate-vectors: ok. test_release_gate: 35 tests OK. test_validate: not rerun (long suite; bound).
- Mutants, all rejected by the gate:
  - 1 byte appended to docs/skillfile-sources.md → digest mismatch
  - compatible_core tag set to rc.12 → "does not pin the accepted source suite"
  - rc.9.json set to rc.12-tag bytes → "published rc.9 release metadata changed"
  - an extra file in the suite → validate "inventory incomplete"
- release.yml diff only extends the regenerate diff path list. Signing and provenance steps are untouched.

## Tag command for orchestrator (after merge)
git tag -s v1.0.0-rc.13 -m "Curator Protocol v1.0.0-rc.13" <merged-main-sha> && git push origin v1.0.0-rc.13
Watch: the "Release specification" workflow (release.yml) on tag v1.0.0-rc.13.

## Non-blocking
- tools/test_validate.py was not rerun here.
- Hosted checks will cover it.
