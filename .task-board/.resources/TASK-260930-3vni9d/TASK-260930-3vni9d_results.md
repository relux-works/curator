# TASK-260930-3vni9d results — content-hash v2 without the skillfile-sources half

Base: spec main 2b649c4. Source: refs/campaign/2vapkz-rev3-full-526a9aa (526a9aa0). The work is left uncommitted in the Story worktree. It is staged (`git add`) so that `make regenerate-check`'s index diff is meaningful. Nothing was committed.

## Steps
1. Applied `git diff 2b649c4 526a9aa -- . ':!.task-board'` with git apply.
2. Restored to 2b649c4, deleting the added files: conformance/skillfile-sources-v1/**, schemas/skillfile-sources-v1/**, protocol/skillfile-sources.md, and the skillfile-sources `manifest_sha256` in release/1.0.0-rc.13.json.
3. **Deviation, also restored: protocol/repository-transport.md.** The generator (tools/generate-vectors/main.go:2279) makes conformance/skillfile-sources-v1/manifest.json hash this file as a suite input. Keeping rev3's edit here therefore changes skillfile-sources-v1/manifest.json and its release pin, and step 5's byte-identity condition cannot then hold together with a green regenerate-check. The rev3 edit itself is Skillfile lock replay binding to `(hash_version, content_sha256)`, which is skillfile-lock-v2 semantics. It belongs to the skillfile-sources half and moves to TASK-260930-3ny11n.
4. Replaced the carrier-naming text with one sentence: "The skillfile-sources carriers `skillfile-lock-v2`, `install-marker-v6`, and `source-audit-v2` adopt `hash_version` in follow-up TASK-260930-3ny11n." This was done in CHANGELOG.md, protocol/core.md, and schemas/v1/README.md. The brief lists core/registry/CHANGELOG. registry.md named none of the carriers. schemas/v1/README.md did name them, so it got the same sentence (deviation, noted). In core.md, rev3's §10 "requires marker schema 6" (the v6 marker) was put back to 2b649c4's "marker schema 5".

## Gates (real exit codes)
- `go run ./tools/generate-vectors -root .` → 0
- `python tools/validate.py` → 0 ("validated 72 schemas and 1253 vector files"). This ran in a temporary venv with jsonschema, because the system python3 lacks jsonschema; the first run with system python3 exited 1 with ModuleNotFoundError.
- `make regenerate-check` → 0
- `go test ./tools/...` → 0
- `python -B -m unittest test_validate` (run from tools/) → 0, 551 tests OK
- `git diff --stat 2b649c4 -- conformance/skillfile-sources-v1 schemas/skillfile-sources-v1 protocol/skillfile-sources.md protocol/repository-transport.md` → empty

## Manifest digest
conformance/v1/manifest.json sha256: `950ee74ad148615c273fe95bbb93f1bc0f9bdf2ea2bd9395f1dc8e3601419e60`. It is identical to rev3-full, because the v1 suite covers none of the removed paths. In release/1.0.0-rc.13.json the candidate and required pins become 950ee74…. The skillfile-sources pin stays at 061ec05d… (2b649c4).

## Paths differing from rev3-full (22)
- CHANGELOG.md: the carrier list is replaced by the follow-up sentence.
- protocol/core.md: "marker schema 6" is reverted to 5, and the "source-extension marker v6 carries…" sentence is replaced by the follow-up sentence.
- schemas/v1/README.md: the "Source-extension carriers are …" sentence is replaced by the follow-up sentence.
- release/1.0.0-rc.13.json: the skillfile-sources manifest_sha256 is kept at 2b649c4's value.
- protocol/repository-transport.md: kept at 2b649c4 (see deviation 3).
- protocol/skillfile-sources.md, schemas/skillfile-sources-v1/README.md: kept at 2b649c4.
- schemas/skillfile-sources-v1/{install-marker-v6,skillfile-lock-v2,source-audit-v2}.schema.json: not added.
- conformance/skillfile-sources-v1/index.json, manifest.json: kept at 2b649c4.
- conformance/skillfile-sources-v1/schema-cases/{install-marker-v6 ×3, skillfile-lock-v2 ×4, source-audit-v2 ×3}: not added.
All other rev3 paths (126) are byte-identical to rev3-full.
