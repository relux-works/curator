# TASK-260910-1xph2y producer review handoff

Status: ready for review. Independent reviewer acceptance is pending.

## Scope and boundary

Execution root: /Users/iv/Developer/ReluxWorks/curator-spec/.temp/STORY-260910-8fv3s5/worktree
Base/HEAD: d019f0e7179520b5c8dcde321c4fe51e04552f58; HEAD..main count was 0.
No commit, tag, push, PR, checkpoint, integration, branch switch, rebase or merge
was performed. Real index remains unchanged. All deliverable changes are
uncommitted documentation, schemas and declarative conformance artifacts.

## Contract

Unreleased source revision 1 adds disjoint Skillfile schema 2 selectors, frozen
local snapshots and membership, complete runtime/build/dependency behavior,
marker 5, receipt 3 and local audit bindings. Legacy v1 wire shapes are unchanged.
The separate transport amendment preserves the closed external-build executor
and broker, stable repository identity, bounded authorized fallback and fail-closed
security classes. Author/operator examples and canonical navigation are updated.

Draft schemas/cases use a separate draft-sources-v1 namespace so no rc.9 schema,
fixture, release digest, generated vector or implementation claim is rewritten.
The source contract explicitly distinguishes legacy configured Git without a
network identity from local dirty-byte snapshots and network Git identities.
Root project packages require machine-owned explicit root_inputs; broad aliases
selecting nested authored packages remain legal without that list.

## Validation evidence

All commands ran as standalone processes without tee or status-masking pipelines.
Python commands used .temp/source-contract/venv/bin/python after installing the
repository-pinned requirements-dev.txt in that task-scoped venv.

| Command | Exit | Evidence |
|---|---:|---|
| Python extraction of the documented draft validation block, then venv/bin/python .temp/source-contract/validate-draft.py | 0 | draft-validation-02.log: 67/67 schema cases, 53/53 negatives, 7/7 wire schemas, 3/3 snapshot vectors |
| venv/bin/python tools/validate.py (latest docs) | 0 | validate-02.log: 60 schemas, 1047 vectors |
| venv/bin/python -B -m unittest discover -s tools -p 'test_*.py' | 0 | unittest-01.log: 227 tests |
| go test ./tools/... | 0 | go-test-01.log |
| git diff --check | 0 | diff-check-02.log |
| git diff --exit-code -- schemas/v1/*.json conformance/v1 release tools | 0 | frozen-check-01.log |
| git diff --cached --exit-code | 0 | index-check-01.log |

The existing make validate components ran individually; make validate itself and
make regenerate-check were not run. No generator was invoked and no historical
expected bytes were edited. Earlier draft check passed 59/59; the expanded
67-case check supersedes it. Existing validator was rerun after navigation edits.

## Evidence bounds

Manager semantic execution: 0/33 semantic cases. These are explicit downstream
requirements, not passing integration tests. Schema/digest checks do not prove
physical path containment, capture-race rejection, write-boundary safety,
authentication fallback, native execution or local v2 installation. No such
implementation or platform claim is made. The independent reviewer should
inspect the actual CR diff and run the documented draft validation command.

## Logbook

- Initial required set_status command exited 1 because an estimate was missing.
  Set estimate 8, then set_status development exited 0. No gate was bypassed.
- System Python lacked jsonschema (ModuleNotFoundError, exit 1). Installed only
  requirements-dev.txt into the task venv; pip installation exited 0.
- Existing corpus couples schema discovery, generator output and release hashes.
  Kept new draft schemas and cases isolated instead of rewriting frozen evidence.
- No logbook CLI/connector was available; scoped board schema lookup confirmed
  unknown operation logbook (exit 1). This attached task-scoped Logbook section
  and the separate logbook resource persist the findings through resource CRUD.
- Routine syntax chosen: explicit repository field for logical identity;
  source-policy.json ordered endpoints with optional pin; root_inputs for root
  package admission. Advanced ports/mirrors/aliases remain unsupported and are
  recorded in UNRESOLVED_QUESTIONS.md.
- Source-audit is a machine-local evidence binding, never a registry attestation.
  Required network evidence for a local source fails closed. Verified script
  operations are not invented; existing assurance build input fields bind receipt 3.

## Exact changed paths and SHA-256

- `CHANGELOG.md` — `c98b76cc8b1d6d5dca44eca2df0d659d7eda67530feba9db504aa1874276d4dc`
- `COMPATIBILITY.md` — `2d38d6ccce08b6326619409fa590cfeb5e4e60303846c41a9a16a5ab2691c90f`
- `README.md` — `7d6426bce7e02d2b41ab2680a6af1a3e2605daccfa1fb0f67abd0db76a6e2b31`
- `UNRESOLVED_QUESTIONS.md` — `e4a7145e6c3bc5737c94c9c8fd64afac172f328042ea04a2994e4a562df25641`
- `conformance/README.md` — `f1ec73eb2846a78daef8718a7f3c2caa538923291cd8484cff0f18e655a73b02`
- `conformance/draft-sources-v1/README.md` — `f058aa0e6206288ad39a9b9a70d0e6b2af1080af04bd341a5017b80aababe06a`
- `conformance/draft-sources-v1/index.json` — `f4cf029af9d395c4ae038a2ca6b77fef658b4a74b333d0b87443f5b228ebc493`
- `conformance/draft-sources-v1/schema-cases/build-receipt-v3/invalid-kind.json` — `4934f9d47028c095e8ffa488c8bda3c43bd9d18415ca1281c83a6686bb6f4aae`
- `conformance/draft-sources-v1/schema-cases/build-receipt-v3/invalid-policy.json` — `cdf4abf438fc77afc21f743b4ac50b3d2ef785ac94669f6a1a7becdcfabec512`
- `conformance/draft-sources-v1/schema-cases/build-receipt-v3/invalid-unknown-top-level.json` — `af8d8fffa8456032d9f56fe213a234308898561f234fcda1a6457df1ece21738`
- `conformance/draft-sources-v1/schema-cases/build-receipt-v3/valid.json` — `4249b5735724a5166e526089077ea62728a6f3c96b263ad73c5a70f48c50e411`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-fake-commit.json` — `c7f2e6cb7c32e47d5521a0b9e556969db69628ce4bd2b0622e01f1d5e6c94e98`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-unknown-top-level.json` — `dcf0ce677bff6a961d14d97a4263b3fceb3a13faed5840b654f41a2390d814d5`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/valid.json` — `5982c6a579e2b49f6cc9c897161c6e332a1c47cf2728ff6f5056a5bdde644f7f`
- `conformance/draft-sources-v1/schema-cases/local-snapshot-v1/invalid-algorithm.json` — `4e927ab70268d9fc0a4db30529f6339814a8d411df1cb2fb41c8b65a6ff6b549`
- `conformance/draft-sources-v1/schema-cases/local-snapshot-v1/invalid-unknown-top-level.json` — `d4a5921a44821f739f9e0ffc950614683355eac706fe4b541c89ee19f70442e2`
- `conformance/draft-sources-v1/schema-cases/local-snapshot-v1/valid.json` — `ab530341bc5d50c4069981603ad2a5cc975abda957a9c30eece2c64f6e4821fa`
- `conformance/draft-sources-v1/schema-cases/skillfile-lock-v1/invalid-commit-format.json` — `54d2e4ebca07335f088f757433782287ca463d4e1e45b41bf0df15da6ba881d8`
- `conformance/draft-sources-v1/schema-cases/skillfile-lock-v1/invalid-configured-escape.json` — `5f0af4a654afc5979a3924fa45f341542de654ea4db7dd088ea43103df89379c`
- `conformance/draft-sources-v1/schema-cases/skillfile-lock-v1/invalid-fake-commit.json` — `c037ccdd11cf84aa2492aed07be8d4dc725e445bd34be34ae9ea760037810085`
- `conformance/draft-sources-v1/schema-cases/skillfile-lock-v1/invalid-unknown-top-level.json` — `79730aa0bb621a2c05f54625bfb5fcd0994ce80f7700048d994d838e2f179ad6`
- `conformance/draft-sources-v1/schema-cases/skillfile-lock-v1/valid-configured-git.json` — `fe69fa386641d0f70eda164f87be0e04be77273d0074375703d0a1411399723d`
- `conformance/draft-sources-v1/schema-cases/skillfile-lock-v1/valid-git.json` — `d5c5d87db858bb5ccdfacff88392b6dd08d0bd63945f80c639aaae19a8fcc596`
- `conformance/draft-sources-v1/schema-cases/skillfile-lock-v1/valid.json` — `2807faa0a7c7ca6d6a7bcd46342e5920e744ac5756329bba1c8e0dbecb8330e1`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-directory-0.json` — `c734a90c50619b2b13b389030cabd7d0d5d259430a0f14d4b1f3ab50121a7246`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-directory-1.json` — `a928f977644c19dd2c6f7f7223bd370d2d6df65964182731c9e00c3d5483bf7f`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-directory-10.json` — `c5e89127383dc4af6ca8040556452c4f8f5ee31189da63764c58b29682596577`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-directory-11.json` — `9cdcc14555513a4e1511c06b62482cc684fbfdf16a8a924ee28b673b3de46dd6`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-directory-12.json` — `f8ecfe293e4c233e569fb64fed36078e5402faccc8aea3e33e556f84b7d42333`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-directory-13.json` — `cf8465b94f9a4f69053faac77ef2d68eef424ce3e9a60a39da3a049e76a30697`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-directory-2.json` — `4e9c79de7b17e53a62fd74528095743fcc536bef967730c1e9a868d7f217e9b2`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-directory-3.json` — `1f566db688c3000ca9119e67c17707631341ca562928062cb5d51e1bafe30f85`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-directory-4.json` — `d8d177d5ef26b3e5aa8df01af7087462d774917b14a83d122bc9ab10c1942762`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-directory-5.json` — `7e4922bd837c4e0bedf46e3f4f7a9d11ab151b59cf3598b96255dbd41328412c`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-directory-6.json` — `01901d2e5329a8d661aafad2106f2737a8c74920fa7aa6a4b69c0a1264c66bce`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-directory-7.json` — `b932d5592dd46c0d9738bcb2211f74508a3b88b83aef59902027a6e7f578f99b`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-directory-8.json` — `1e9b658d8233c89674f4ba6021fbcdd92161020bced8379007aa64f1b9c5b7fb`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-directory-9.json` — `1591df0dd03d5d900f90817aa7e5b7867b18da377aef381d3385518e62d4afa1`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-duplicate.json` — `9bf1dd34ea279f7a5e7b02363fa88354161012119d464dceebec0c0984751dd0`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-empty.json` — `9ca8630bf91460f70a6a070d426e58a55481e633756356e41a091670e3fe357e`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-mixed-collection.json` — `e3052cd31e0a09fa85d3f40e846b058162f6cfb8256b0f2131c98a1664dc54bd`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-mixed-git.json` — `664daa1da82237c667140cfd61582a3427255d40fec2bff406070d0f85dba086`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-mixed-ref.json` — `d3fbd6d23833bf489a67f64b620d4d720ec9c1a1d4e1290d7eaa38941bd11498`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-mixed-source.json` — `85d404db77256b547a799f9642911e3a3bd77ba7d7dde35d5ae5c14e110212c1`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-partial.json` — `dcf8ed12f8339611af7bf93d1e5fef3928d98a066f9179d03467abbc8464bc90`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-path.json` — `b65545a59b94c94e76f2ad464b001f66ba6543bef2777dd3998c82e5e2b31dc6`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-recursive.json` — `7894bf46e8c121e09b07009ded27ccb9d4f4b0abfbbe6dd171f3555b51e74486`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-source-0.json` — `d174a4e749e2e8831a66186b3b3a2b7df0e08289ebb430897cb10dbdd1e7d471`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-source-1.json` — `59892d40eb8584a4ab5c8d18998c32dbfe6ecceb49237f5e4503db39228fe094`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-source-2.json` — `f0b5cd5ab20ed76f976d174a7fda78c33135f377bd25448916e62bb5ea43d7f6`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-source-3.json` — `484023e4262e62d75e4fe609bcb0b4cf4af1640134ff50212fcaee9f805c233b`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-source-4.json` — `fb285c34eb9cabddae4a480589e4d581b31c5470fa5c772c55db33ffa1280098`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-source-5.json` — `47bdc9ebd29a7bc5331f69ec73992fb57609b145278aeeaf0e23181924793e45`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-source-6.json` — `3d334eca8ab07e47ddb41754b54090177d7dbcae91e27e2d9e378ac29f9d7787`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-source-7.json` — `17d26094e124aea32875d59e7b49b49775b722ecaddce99a16a14c2e56148835`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-source-8.json` — `8993f2bd092ee9365088fd44da065401c63721ef2486cdf2687e351b2a12e2d4`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-source-9.json` — `093c43252f0778795856620df46d0ade2117a9e6845422b099f054099ccb306a`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-unknown.json` — `3c97eae28fc38ab958d91e65f4afc83ae5839ed75f1259f9a56a221e28d6830d`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/invalid-wildcard-exclude.json` — `147d03c915994071e895b43c3b02a173e35071050c1ef87709fb4fcc07f3644d`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/valid-absolute.json` — `e2ebdbcafd213362cab2f2e8b5eb7bba0abb7e70d5f461c898e5f3daaca815c7`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/valid-https.json` — `78de47cab44545a585933435528aa7da01f269dda572f650162af17868358723`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/valid-legacy-only.json` — `ee5d21856c9825d17df185fdd565e4eeef9e14439842dc4fece0f3461d4daa7b`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/valid-mixed.json` — `15194d5ebadc4b1602df08f7e026404d34f0ef04a7c02fbfc5851bca0faa1c84`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/valid-relative.json` — `a3bb55be1b152ce5e41062a42ffbf7328b40099cd8ebc9454d234ff3cee3801b`
- `conformance/draft-sources-v1/schema-cases/skillfile-v2/valid-ssh.json` — `b27bbb269e404c53d3946062a2e69d166e5a764d6aee576574aa0a2601189776`
- `conformance/draft-sources-v1/schema-cases/source-audit-v1/invalid-decision.json` — `0d0d5bb25e4084187a500d7d487d8664d616eaa84fabf1936a922d4a963af75e`
- `conformance/draft-sources-v1/schema-cases/source-audit-v1/invalid-no-evidence.json` — `86da032b288461e5cd22dcdf0e8b57e27294817d3f47e2a401f4010c608841d4`
- `conformance/draft-sources-v1/schema-cases/source-audit-v1/invalid-unknown-top-level.json` — `35caed1cdd9927d5c369c167abf95be5c900c17f152020ac4823d7ea503dd90f`
- `conformance/draft-sources-v1/schema-cases/source-audit-v1/valid.json` — `4877ffd7c5a34b6e601d44fff2b79dd0d89127603ddd9158d0a9c83ddc03748a`
- `conformance/draft-sources-v1/schema-cases/source-policy-v1/invalid-fallback.json` — `085076969730dc5d7e5c00056e290f8da078bcc40ffa76af2667373288222cba`
- `conformance/draft-sources-v1/schema-cases/source-policy-v1/invalid-helper.json` — `c5aaf8699194f3cddc0a11e56c872ce3ce79e97193e34940a9e206ea2e463c8d`
- `conformance/draft-sources-v1/schema-cases/source-policy-v1/invalid-too-many.json` — `00d13f7c6479d9f4dc30a99ecd2cabbbd482565fef68c29b5708d086327b5fae`
- `conformance/draft-sources-v1/schema-cases/source-policy-v1/invalid-unknown-top-level.json` — `08891f094b92e8ab28f8a34c4ad1204f85f7f544b5a8208665a8ec9fbb5033b7`
- `conformance/draft-sources-v1/schema-cases/source-policy-v1/valid.json` — `1234c2333e64935e6a962a99118994418cf9cad6f6458e676de45dc03a6dd10d`
- `conformance/draft-sources-v1/semantic-cases.json` — `4eb185011b9540d69554c35350f8e4f43a9169ef4ea4f1cf2cd3be2691e52a56`
- `conformance/draft-sources-v1/snapshot-cases.json` — `1922256efe21f934b667ec913f34a2af3b16004ecb00c63ee8712eacaf347999`
- `docs/skillfile-sources.md` — `182662d3ec11d8678df0da3015d75287bb7964fab19901e88839a93b14502db0`
- `profiles/manager.md` — `9cc6eb7242c46d44888a66393b72eccecc1b829a77ce5577ee130086a0af436a`
- `protocol/core.md` — `a88d9645474239f54f7c7c1860e91c6a1caa3f46ae6f4f53d5e49c8085223fc1`
- `protocol/repository-transport.md` — `a89f8c691d42c5e98244be955806520cda1b2a9269b5cf1925e391416a0ca5cd`
- `protocol/skillfile-sources.md` — `d6940b83d05f12bd6081cedb77c2446e02cb5c98f450cc5b2fd65fde4db7092f`
- `schemas/draft-sources-v1/README.md` — `584edc85d340d611b40373a36f65bc1b4dc5a800aa5f04f0b954151da24165ff`
- `schemas/draft-sources-v1/build-receipt-v3.schema.json` — `5e76dce74334f342cc3dc6fdefc78f3a66a13ac68a7fe03c5c9aa80b8a45002f`
- `schemas/draft-sources-v1/install-marker-v5.schema.json` — `227213e493e2c9c78ab1090ced50679e66b73d2abab1f7299555743c1946c663`
- `schemas/draft-sources-v1/local-snapshot-v1.schema.json` — `c91d14cc9300c7bf1799dfb8f332a50516752933452abcdf60c161ac91a68a5b`
- `schemas/draft-sources-v1/skillfile-lock-v1.schema.json` — `9215b64982015226e86bf46d1c8734e098709223eee4bd2d268020a7d728f97f`
- `schemas/draft-sources-v1/skillfile-v2.schema.json` — `4574302117961f85260136b932e924061a1ad917833548dea99412083a0e36b8`
- `schemas/draft-sources-v1/source-audit-v1.schema.json` — `265e2ed6a010f9df95a3c95c89d3528092b8c1bd612659c6caa1ef5c01256b26`
- `schemas/draft-sources-v1/source-policy-v1.schema.json` — `841254f63c39faa70739355273b121ef2769492a77114bf835b53068ad42ce60`
- `schemas/draft-sources-v1/source-types-v1.schema.json` — `fbb389bbfe5e2d05293e669e196d150f4fbaef1601ae239f8e7054c200621c15`
- `schemas/v1/README.md` — `c4760b95beacce13e4a687cc298fb6678a46367050e6e4e9ba0f953dfaa5c5ae`
