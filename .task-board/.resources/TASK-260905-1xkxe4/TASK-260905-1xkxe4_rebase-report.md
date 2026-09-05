# TASK-260905-1xkxe4 rebase report: draft/environments-schemas-1-1 onto main f61ee9a

## Head

- Previous head: `794c7bd` (cycle-1-accepted `401b665` minus `LOGBOOK.md`).
- New head: `fd237ba` — exactly one signed commit past `origin/main` (`f61ee9a`).
- `git log --oneline origin/main..HEAD`: `fd237ba Deliver the environments 1.1 section 13 schemas, cases, and vector families`
- Signature: `git verify-commit HEAD` → `Good "git" signature with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM` (the allowed_signers principal lookup warning is the local verify config, not the commit).
- Pushed with `git push --force-with-lease origin HEAD:refs/heads/draft/environments-schemas-1-1` (`+ 794c7bd...fd237ba`). PR #42 head = `fd237ba`.

## Conflict resolutions

- `CHANGELOG.md`: kept batch 3's `### Added` bullet (manager-config-v2) and `### Changed` block from main; added this batch's Added bullet after it; added this batch's `### Removed` section after Changed. The two pre-existing snapshot byte-exactness bullets (which both sides had drifted under a wrong header) sit under `### Added` where they were at `a68559b`. One `## Unreleased` with Added / Changed / Removed.
- `conformance/README.md`: this batch's rewritten agent-environments bullet replaces the old revision-1 bullet; batch 3's `manager-config` schema-2 bullet kept after it.
- `conformance/v1/manifest.json`, `conformance/v1/schema-cases/index.json`, `release/1.0.0-rc.9.json`: took the branch side, then `make regenerate` rewrote all three (they are generated); the regeneration was amended into the commit.
- `gofmt -w tools/generate-vectors/*.go` reformatted `context_detectors.go`, `context_versions.go`, `environments.go` (local gofmt go1.25.5; go1.23.12 gofmt agrees: `gofmt -l tools` empty under both).

## Range-diff `git range-diff origin/main 794c7bd HEAD` (hunks per file)

| file | hunks | cause |
| --- | ---: | --- |
| commit message | 1 | context only (CHANGELOG lines) |
| CHANGELOG.md | 1 | conflict resolution |
| conformance/README.md | 1 | conflict resolution |
| conformance/v1/manifest.json | 3 | regeneration (batch 3 entries + index hash) |
| conformance/v1/schema-cases/index.json | 10 | regeneration (manager-config-v2 cases interleaved) |
| release/1.0.0-rc.9.json | 2 | regeneration (manifest pin) |
| tools/generate-vectors/context_detectors.go | 1 | gofmt |
| tools/generate-vectors/context_versions.go | 2 | gofmt |
| tools/generate-vectors/environments.go | 1 | gofmt |
| tools/validate.py | 1 | context lines only (batch 3's `validate_manager_config_vectors` import neighbour); no content change |

Full range-diff (1017 lines) below.

## Consumed-by-pin files

`git diff origin/main --stat -- conformance/v1/vectors/manager-config.json conformance/v1/vectors/manager-config-v2.json conformance/v1/vectors/canonical-valid.json conformance/v1/vectors/canonical-invalid.json conformance/v1/vectors/source-identities.json conformance/v1/vectors/identifiers.json conformance/v1/vectors/locale-selectors.json conformance/v1/expected/snapshot_sha256.txt` → empty output, exit 0.

## Gate tails (run directly, venv `.temp/venv`, at fd237ba)

`make validate` → exit 0:
```
----------------------------------------------------------------------
Ran 204 tests in 36.303s

OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	0.439s
```

`make regenerate-check` → exit 0:
```
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
```

`go test ./tools/generate-vectors/...` → ok, exit 0. `gofmt -l tools` → empty, exit 0.

## PR #42 checks at fd237ba

| check | result |
| --- | --- |
| Formatting | pass |
| Links | pass |
| Specification (ubuntu/macos/windows) | pass / pass / pass |
| Implementations (ubuntu/macos/windows) | pass / pass / pass |
| Release target provenance | skipping (not a release target) |

Note: the Formatting failure shown first by `gh pr checks --watch` belonged to the superseded run for `794c7bd` (run 33970049430); the runs for `fd237ba` (33970489083, 33970489086) are all green.

## Full range-diff

```
1:  794c7bd ! 1:  fd237ba Deliver the environments 1.1 section 13 schemas, cases, and vector families
    @@ Commit message
     
      ## CHANGELOG.md ##
     @@ CHANGELOG.md: Versioning for the complete specification set.
    - 
    - ### Added
    +   the schema each case's `schema_version` selects and cross-checks the schema-2 knob names and
    +   literal defaults against the section 12.1 table.
      
     +- Delivered the environments.md revision 1.1 conformance surfaces of section
     +  13. Schemas: new `agent-context-v1`, `agent-mcp-v1`, and `context-lock-v1`;
    @@ CHANGELOG.md: Versioning for the complete specification set.
     +  recomputes every hash, byte length, order, lock, finding, and materialized
     +  byte independently of the generator and fails on a hand-edited expected
     +  file.
    -+
    ++- Added the environments section 1.2 snapshot byte-exactness rule: a snapshot
    ++  produced from a commit carries exactly the committed blob bytes, and neither
    ++  working-tree conversion (`core.autocrlf`, `text`/`eol`, filters, `ident`) nor
    ++  attribute-driven archive processing (`export-subst`, `export-ignore`) may
    ++  alter, add, or omit an entry. Content hashes, state hashes, effective pins,
    ++  and every hash-bound identity are therefore platform- and
    ++  configuration-independent, which the section 5.6 cross-platform equality
    ++  claim had assumed without stating. `git archive` violates the rule under
    ++  `core.autocrlf=true` and for `export-subst` entries; object-database
    ++  extraction satisfies it.
    ++- Added the `snapshot-acquisition.json` vector over the new
    ++  `fixtures/byte-exact` tree (`* text=auto`, an `export-subst` entry, LF, CRLF,
    ++  and mixed-ending files) with its expected content hash, fixture blobs
    ++  committed through plumbing so they survive every checkout unconverted under
    ++  the repository's `eol=lf` policy (a root `.gitattributes` note explains why
    ++  no attribute rule can protect them against the fixture's own nested
    ++  `* text=auto`), and validator cross-checks that fail on a normalized
    ++  checkout, an expanded placeholder, or a hash that omits `.gitattributes`.
    ++
    + ### Changed
    + 
    + - Rewrote manager profile section 12 on the environments 1.1 model per the
    +@@ CHANGELOG.md: Versioning for the complete specification set.
    +   `env config` rows, and the `curator run` provider pointer to Decision 0013
    +   Decision 6.4.
    + 
    +-- Added the environments section 1.2 snapshot byte-exactness rule: a snapshot
    +-  produced from a commit carries exactly the committed blob bytes, and neither
    +-  working-tree conversion (`core.autocrlf`, `text`/`eol`, filters, `ident`) nor
    +-  attribute-driven archive processing (`export-subst`, `export-ignore`) may
    +-  alter, add, or omit an entry. Content hashes, state hashes, effective pins,
    +-  and every hash-bound identity are therefore platform- and
    +-  configuration-independent, which the section 5.6 cross-platform equality
    +-  claim had assumed without stating. `git archive` violates the rule under
    +-  `core.autocrlf=true` and for `export-subst` entries; object-database
    +-  extraction satisfies it.
    +-- Added the `snapshot-acquisition.json` vector over the new
    +-  `fixtures/byte-exact` tree (`* text=auto`, an `export-subst` entry, LF, CRLF,
    +-  and mixed-ending files) with its expected content hash, fixture blobs
    +-  committed through plumbing so they survive every checkout unconverted under
    +-  the repository's `eol=lf` policy (a root `.gitattributes` note explains why
    +-  no attribute rule can protect them against the fixture's own nested
    +-  `* text=auto`), and validator cross-checks that fail on a normalized
    +-  checkout, an expanded placeholder, or a hash that omits `.gitattributes`.
     +### Removed
     +
     +- Withdrew `profilefile-v1` and `context-manifest-v1` with their schema cases
     +  and the `monolithic-composed-empty-chapter` expected set, replaced under
     +  section 13 by the lock model.
    -+
    - - Added the environments section 1.2 snapshot byte-exactness rule: a snapshot
    -   produced from a commit carries exactly the committed blob bytes, and neither
    -   working-tree conversion (`core.autocrlf`, `text`/`eol`, filters, `ident`) nor
    + 
    + ## 1.0.0-rc.9 - 2026-08-23
    + 
     
      ## conformance/README.md ##
     @@ conformance/README.md: The suite contains:
    @@ conformance/README.md: The suite contains:
     +  pattern class, MCP `args` and `url` in scope, the scoped waiver that clears
     +  only its own span at its pin, the unpinnable case, and the
     +  `context-system-module-present` warning);
    - - the environments section 1.2 snapshot byte-exactness vector
    -   (`vectors/snapshot-acquisition.json` over `fixtures/byte-exact`): a
    -   committed tree with `* text=auto`, an `export-subst` entry, and LF, CRLF,
    + - the `manager-config` schema-2 surface: `schema-cases/manager-config-v2`
    +   with every environments section 12.1 knob present, one negative per
    +   closed-object rule and per value grammar, and a schema-1 rejection, plus
     
      ## conformance/v1/expected/environments/mcp-claude-code/.agent-context/mcp/claude_code.json (new) ##
     @@
    @@ conformance/v1/manifest.json
     -      "sha256": "sha256:a4c9a7b2af1a57d0ca6504fe2d62e93b8dc779153bab51cc93ec1e9aacd581cd"
     +      "path": "schema-cases/context-lock-v1/invalid-context-without-version.json",
     +      "sha256": "sha256:bfcd9f752f05adbc84421abad469fb77d2d5639e6f3c18d7bd58ae0771ac9539"
    ++    },
    ++    {
    ++      "path": "schema-cases/context-lock-v1/invalid-empty-members.json",
    ++      "sha256": "sha256:773c14847042284006f52baa8814d2894cf074657f544c2e15f8327accb507a8"
          },
          {
     -      "path": "schema-cases/context-manifest-v1/invalid-duplicate-path.json",
     -      "sha256": "sha256:96eb36509e28650fffab4116abd0b788e406e8c23e0d7dd6839802a981028952"
    -+      "path": "schema-cases/context-lock-v1/invalid-empty-members.json",
    -+      "sha256": "sha256:773c14847042284006f52baa8814d2894cf074657f544c2e15f8327accb507a8"
    ++      "path": "schema-cases/context-lock-v1/invalid-member-both-pins.json",
    ++      "sha256": "sha256:4f2fbbedfa1f2fa054ce79313b3f1f7bfe0fb728ac1c072153e0ea605170fcbb"
          },
          {
     -      "path": "schema-cases/context-manifest-v1/invalid-empty-environments.json",
     -      "sha256": "sha256:0dfc02d7566008e8664fa930b2cf883f308e1d073434b142072ae4b4547be488"
    -+      "path": "schema-cases/context-lock-v1/invalid-member-both-pins.json",
    -+      "sha256": "sha256:4f2fbbedfa1f2fa054ce79313b3f1f7bfe0fb728ac1c072153e0ea605170fcbb"
    ++      "path": "schema-cases/context-lock-v1/invalid-member-no-pin.json",
    ++      "sha256": "sha256:0bbcd004f9a8d2ecbee6fbc61b44ee4abbb9c0ab60493f98d598921490aa1570"
          },
          {
     -      "path": "schema-cases/context-manifest-v1/invalid-parent-path.json",
     -      "sha256": "sha256:9def476c862009cb9c63b847f7977f292444d4f2006c8993007be5842ea8153c"
    -+      "path": "schema-cases/context-lock-v1/invalid-member-no-pin.json",
    -+      "sha256": "sha256:0bbcd004f9a8d2ecbee6fbc61b44ee4abbb9c0ab60493f98d598921490aa1570"
    ++      "path": "schema-cases/context-lock-v1/invalid-member-unknown-field.json",
    ++      "sha256": "sha256:0cb6292168eea32026614fca0e336c68d40954ab34f2b40c47e6192937f62e1f"
          },
          {
     -      "path": "schema-cases/context-manifest-v1/invalid-unknown-class.json",
     -      "sha256": "sha256:bdc1114173ad1c67871698e2da9901bbfc817871c3354ffcbfa34f1b802d9fc3"
    -+      "path": "schema-cases/context-lock-v1/invalid-member-unknown-field.json",
    -+      "sha256": "sha256:0cb6292168eea32026614fca0e336c68d40954ab34f2b40c47e6192937f62e1f"
    ++      "path": "schema-cases/context-lock-v1/invalid-member-unknown-kind.json",
    ++      "sha256": "sha256:599c3ec318711c431dcd53571cfcc14be0a64033ed69b39b57090712c49592fc"
          },
          {
     -      "path": "schema-cases/context-manifest-v1/invalid-unknown-entry-field.json",
     -      "sha256": "sha256:ce8c0be1da0403e00906cba2e6485ec069a59a5ce3f8c52750cc444fe0b591f2"
    -+      "path": "schema-cases/context-lock-v1/invalid-member-unknown-kind.json",
    -+      "sha256": "sha256:599c3ec318711c431dcd53571cfcc14be0a64033ed69b39b57090712c49592fc"
    ++      "path": "schema-cases/context-lock-v1/invalid-members-duplicate.json",
    ++      "sha256": "sha256:82e8f0c000d31833cdc5f81f95cefb222afc730c08ebc8dacc1f75d3ae702640"
          },
          {
     -      "path": "schema-cases/context-manifest-v1/invalid-version.json",
     -      "sha256": "sha256:09b67ed318ed536f1261e5d4a3c75c73f1ccee6b01b94f8eb7e7dc85ccd43c6d"
    -+      "path": "schema-cases/context-lock-v1/invalid-members-duplicate.json",
    -+      "sha256": "sha256:82e8f0c000d31833cdc5f81f95cefb222afc730c08ebc8dacc1f75d3ae702640"
    ++      "path": "schema-cases/context-lock-v1/invalid-members-unsorted-by-kind.json",
    ++      "sha256": "sha256:356dea443108220046ae5f91b4909027375b7beb05e07255a2786eb781fad319"
          },
          {
     -      "path": "schema-cases/context-manifest-v1/invalid.json",
     -      "sha256": "sha256:4b8894d57dfa621e534ef4eb25263e8f00254cbcb4327f1f98796314ac279dde"
    -+      "path": "schema-cases/context-lock-v1/invalid-members-unsorted-by-kind.json",
    -+      "sha256": "sha256:356dea443108220046ae5f91b4909027375b7beb05e07255a2786eb781fad319"
    ++      "path": "schema-cases/context-lock-v1/invalid-members-unsorted-by-name.json",
    ++      "sha256": "sha256:24fd43e527dd7c54516635b4f8175e72c8a4aadf877f29db449edfe9b232e3cb"
          },
          {
     -      "path": "schema-cases/context-manifest-v1/valid-empty-modules.json",
     -      "sha256": "sha256:577cbca9d1059fc55f6f930a02163a95b06096fa10b3620c6ca536d1145209dc"
    -+      "path": "schema-cases/context-lock-v1/invalid-members-unsorted-by-name.json",
    -+      "sha256": "sha256:24fd43e527dd7c54516635b4f8175e72c8a4aadf877f29db449edfe9b232e3cb"
    ++      "path": "schema-cases/context-lock-v1/invalid-path-member-skill-kind.json",
    ++      "sha256": "sha256:b8a5acab92342bada34c41ab22aead80fe15460e4dffe1ce5fea9290ad0a53bc"
          },
          {
     -      "path": "schema-cases/context-manifest-v1/valid-system-selector.json",
     -      "sha256": "sha256:a788b2903122f0294464a9d03f3a3f7b9d9158b5c77cc31950a0c20451e64b9c"
    -+      "path": "schema-cases/context-lock-v1/invalid-path-member-skill-kind.json",
    -+      "sha256": "sha256:b8a5acab92342bada34c41ab22aead80fe15460e4dffe1ce5fea9290ad0a53bc"
    ++      "path": "schema-cases/context-lock-v1/invalid-path-member-with-directory.json",
    ++      "sha256": "sha256:8a7e788d165c72edc6656472f5d983a58053515c5c442a818a6001781052abd6"
          },
          {
     -      "path": "schema-cases/context-manifest-v1/valid.json",
     -      "sha256": "sha256:1b578799caa9b1b23dcaf3e73c1d0e5d1b25c84eac1c0685d08cbe5492015364"
    -+      "path": "schema-cases/context-lock-v1/invalid-path-member-with-directory.json",
    -+      "sha256": "sha256:8a7e788d165c72edc6656472f5d983a58053515c5c442a818a6001781052abd6"
    -+    },
    -+    {
     +      "path": "schema-cases/context-lock-v1/invalid-path-member-with-source.json",
     +      "sha256": "sha256:819d2da39349b578d9ae6d067c81aeebd67653be2af4174d0617175bf7b25077"
     +    },
    @@ conformance/v1/manifest.json
          },
          {
            "path": "schema-cases/index.json",
    --      "sha256": "sha256:9390843815dadbdf6a794eb533557b4161f8a30d398c62f8c653b7d54c664aa1"
    -+      "sha256": "sha256:a2ecc2fc36a88a9028fda9e4d812d5bb4ad4b9197650a629248c6acf6158941f"
    +-      "sha256": "sha256:461f3b6e1768cb199962ea5c7376800baf9568ead4b804cb2d2ec56e9c679eb4"
    ++      "sha256": "sha256:c0b99fca3de20a6dd3b4171b3b6362f4adbd8eb94f7905ccd7906ecdef452073"
          },
          {
            "path": "schema-cases/install-marker-v1/invalid.json",
    @@ conformance/v1/manifest.json
          {
            "path": "schema-cases/launch-env-fragment-v1/valid-config-key-channel.json",
     @@
    +       "path": "schema-cases/launch-env-fragment-v1/valid-local-state-pin.json",
            "sha256": "sha256:3b73c4351398b08ccc2ecf945e50d17e62fd90df6c3fe2ee59eafd026ad22aae"
          },
    -     {
    --      "path": "schema-cases/launch-env-fragment-v1/valid-no-composition.json",
    --      "sha256": "sha256:6f087d0e1482af29f478ff1e7405273c2238cc8169153417445e4b3d6daac6f4"
    --    },
    --    {
    --      "path": "schema-cases/launch-env-fragment-v1/valid.json",
    --      "sha256": "sha256:2c5432d4997341802096c1bb386a0c69254d9b9711c7741c10a41b16fb3d2e4c"
    --    },
    --    {
    --      "path": "schema-cases/log-response-v1/invalid.json",
    --      "sha256": "sha256:de6ced3c0077dde9b01ed257a9c7e96e575d59e4918d4afa9711a1b0d070a3f2"
    ++    {
     +      "path": "schema-cases/launch-env-fragment-v1/valid-mcp-only.json",
     +      "sha256": "sha256:2ed8e3abe51114487b24fb18bfb94a8ecd67c5583406e320707c079be02fe0d0"
    -     },
    -     {
    --      "path": "schema-cases/log-response-v1/valid.json",
    --      "sha256": "sha256:52b0224bfd696c95d62932517a1004f332bed7e9fec6b12917b44d5a23134494"
    ++    },
    ++    {
     +      "path": "schema-cases/launch-env-fragment-v1/valid-minimal.json",
     +      "sha256": "sha256:ef6f1bf37f0509a89d02e5b0b52817de13ca539a7f46077e7f436e552a5b9d0c"
    -     },
    ++    },
          {
    --      "path": "schema-cases/manager-config-v1/invalid.json",
    --      "sha256": "sha256:a4490b4aeefe0600f599d34b64ec6ed5b6824d7847482a8cae2001a7633a0363"
    -+      "path": "schema-cases/launch-env-fragment-v1/valid-no-composition.json",
    -+      "sha256": "sha256:6f087d0e1482af29f478ff1e7405273c2238cc8169153417445e4b3d6daac6f4"
    +       "path": "schema-cases/launch-env-fragment-v1/valid-no-composition.json",
    +       "sha256": "sha256:6f087d0e1482af29f478ff1e7405273c2238cc8169153417445e4b3d6daac6f4"
          },
    -     {
    --      "path": "schema-cases/manager-config-v1/valid.json",
    --      "sha256": "sha256:1490a267b8b5764b31270549c0f2466add662434d561c890c1bb5bbfb1619292"
    ++    {
     +      "path": "schema-cases/launch-env-fragment-v1/valid-opencode-empty-channels-and-variable.json",
     +      "sha256": "sha256:7ce8b5b3c5b14b24e8e21d3dcf6c899259b91ced18ff2944ad0537b01cdd05d6"
    -     },
    -     {
    --      "path": "schema-cases/profilefile-v1/invalid-duplicate-directory.json",
    --      "sha256": "sha256:ad79aa06b9c53d92605b2c10a7e7848bb34ba5a0fba0b220d0109d705790536b"
    ++    },
    ++    {
     +      "path": "schema-cases/launch-env-fragment-v1/valid-path-prepend.json",
     +      "sha256": "sha256:29e1a29c850581c82fc65448a93fa00a28fc16bf12ce1dab1037e9ebccf9709d"
    -     },
    -     {
    --      "path": "schema-cases/profilefile-v1/invalid-nested-root.json",
    --      "sha256": "sha256:6e144e26f07ad34259a6ec3e926d4f65661ebb028f90fa0e3d279309b6340747"
    ++    },
    ++    {
     +      "path": "schema-cases/launch-env-fragment-v1/valid-pi-file-channels.json",
     +      "sha256": "sha256:fdca4a07627804ceaa8f6af3beeebddcb50442ec0e358764dabef11106643d7c"
    -     },
    -     {
    --      "path": "schema-cases/profilefile-v1/invalid-profile-name.json",
    --      "sha256": "sha256:a58a9f8c5636cc0446a249d63fbd93ad963aab8ad1187cc340f05229f40308ef"
    ++    },
    ++    {
     +      "path": "schema-cases/launch-env-fragment-v1/valid-system-prompt-only.json",
     +      "sha256": "sha256:2b361b8abe5a354abcc3a3200f337b852f07533a0eef6a61a7c6b23293a47683"
    -     },
    -     {
    --      "path": "schema-cases/profilefile-v1/invalid-traversal-path.json",
    --      "sha256": "sha256:24aad00d5c68530bd750d29ac61f686c92b2550dc690df2323e7ae702bf18c0c"
    ++    },
    ++    {
     +      "path": "schema-cases/launch-env-fragment-v1/valid-winner-first.json",
     +      "sha256": "sha256:5fade7c4080bebbb1ee0ff9f3f5978a99b4de72fdf6290904ad866e9d7dc31d9"
    -     },
    ++    },
          {
    --      "path": "schema-cases/profilefile-v1/invalid-unknown-field.json",
    --      "sha256": "sha256:6e97dee2fd547044d33ab7b826e3c40e26553aa58043180441801ee7826aa8ad"
    -+      "path": "schema-cases/launch-env-fragment-v1/valid.json",
    +       "path": "schema-cases/launch-env-fragment-v1/valid.json",
    +-      "sha256": "sha256:2c5432d4997341802096c1bb386a0c69254d9b9711c7741c10a41b16fb3d2e4c"
     +      "sha256": "sha256:8ec5a08a84fe81a939b505f182834d9939e545fd385c1b28957d70c39bea2c48"
          },
          {
    +       "path": "schema-cases/log-response-v1/invalid.json",
    +@@
    +       "path": "schema-cases/manager-config-v2/valid.json",
    +       "sha256": "sha256:e95b4b2fb64c3f15d519cce7d9ab38bec488b2a88f7ba08dd1606ba0034e7fbe"
    +     },
    +-    {
    +-      "path": "schema-cases/profilefile-v1/invalid-duplicate-directory.json",
    +-      "sha256": "sha256:ad79aa06b9c53d92605b2c10a7e7848bb34ba5a0fba0b220d0109d705790536b"
    +-    },
    +-    {
    +-      "path": "schema-cases/profilefile-v1/invalid-nested-root.json",
    +-      "sha256": "sha256:6e144e26f07ad34259a6ec3e926d4f65661ebb028f90fa0e3d279309b6340747"
    +-    },
    +-    {
    +-      "path": "schema-cases/profilefile-v1/invalid-profile-name.json",
    +-      "sha256": "sha256:a58a9f8c5636cc0446a249d63fbd93ad963aab8ad1187cc340f05229f40308ef"
    +-    },
    +-    {
    +-      "path": "schema-cases/profilefile-v1/invalid-traversal-path.json",
    +-      "sha256": "sha256:24aad00d5c68530bd750d29ac61f686c92b2550dc690df2323e7ae702bf18c0c"
    +-    },
    +-    {
    +-      "path": "schema-cases/profilefile-v1/invalid-unknown-field.json",
    +-      "sha256": "sha256:6e97dee2fd547044d33ab7b826e3c40e26553aa58043180441801ee7826aa8ad"
    +-    },
    +-    {
     -      "path": "schema-cases/profilefile-v1/invalid-version.json",
     -      "sha256": "sha256:b4fec4d54436ce8371b56e419c72ada57f9020e066180d4a48fb81da632ebb27"
    -+      "path": "schema-cases/log-response-v1/invalid.json",
    -+      "sha256": "sha256:de6ced3c0077dde9b01ed257a9c7e96e575d59e4918d4afa9711a1b0d070a3f2"
    -     },
    -     {
    +-    },
    +-    {
     -      "path": "schema-cases/profilefile-v1/invalid.json",
     -      "sha256": "sha256:9c9b0dd225a06b8dbf4d956d07a892e661413d3cd0be5d85223a065f23ebd863"
    -+      "path": "schema-cases/log-response-v1/valid.json",
    -+      "sha256": "sha256:52b0224bfd696c95d62932517a1004f332bed7e9fec6b12917b44d5a23134494"
    -     },
    -     {
    +-    },
    +-    {
     -      "path": "schema-cases/profilefile-v1/valid-single-profile.json",
     -      "sha256": "sha256:681fa8a01bac3e58ca9511c96b08578e03626d39f46573f54f8ea4ac7cb0f2a7"
    -+      "path": "schema-cases/manager-config-v1/invalid.json",
    -+      "sha256": "sha256:a4490b4aeefe0600f599d34b64ec6ed5b6824d7847482a8cae2001a7633a0363"
    -     },
    -     {
    +-    },
    +-    {
     -      "path": "schema-cases/profilefile-v1/valid.json",
     -      "sha256": "sha256:e68b4a6028ba82bb621d1bda76b8841552c7733c38259dcb1d5852dacf47d2fb"
    -+      "path": "schema-cases/manager-config-v1/valid.json",
    -+      "sha256": "sha256:1490a267b8b5764b31270549c0f2466add662434d561c890c1bb5bbfb1619292"
    -     },
    +-    },
          {
            "path": "schema-cases/provider-capability-receipt-v1/invalid-observed-not-before-expires.json",
    +       "sha256": "sha256:946cbb14c0be40adb8c1a9ef7892988a3bdf163bc0071f63045ec1abf147cead"
     @@
            "path": "vectors/conformance-claim-v3-qualification.json",
            "sha256": "sha256:aefa6ee21c70cb46e5552f3fd608849d0f69b0ea034428f81eacd00d9d1fffe1"
    @@ conformance/v1/schema-cases/index.json
     +  },
     +  {
     +    "instance": "agent-environment-marker-v1/invalid-git-without-requirement.json",
    ++    "schema": "agent-environment-marker-v1.schema.json",
    ++    "valid": false
    ++  },
    ++  {
    ++    "instance": "agent-environment-marker-v1/invalid-git-requirement-branch.json",
          "schema": "agent-environment-marker-v1.schema.json",
          "valid": false
        },
        {
     -    "instance": "agent-environment-marker-v1/invalid-local-with-ref.json",
    -+    "instance": "agent-environment-marker-v1/invalid-git-requirement-branch.json",
    ++    "instance": "agent-environment-marker-v1/invalid-git-requirement-two-forms.json",
          "schema": "agent-environment-marker-v1.schema.json",
          "valid": false
        },
        {
     -    "instance": "agent-environment-marker-v1/invalid-path-with-commit.json",
    -+    "instance": "agent-environment-marker-v1/invalid-git-requirement-two-forms.json",
    ++    "instance": "agent-environment-marker-v1/invalid-git-with-source-path.json",
          "schema": "agent-environment-marker-v1.schema.json",
          "valid": false
        },
        {
     -    "instance": "agent-environment-marker-v1/invalid-path-with-ref.json",
    -+    "instance": "agent-environment-marker-v1/invalid-git-with-source-path.json",
    -+    "schema": "agent-environment-marker-v1.schema.json",
    -+    "valid": false
    -+  },
    -+  {
     +    "instance": "agent-environment-marker-v1/invalid-local-with-source.json",
          "schema": "agent-environment-marker-v1.schema.json",
          "valid": false
    @@ conformance/v1/schema-cases/index.json
     +  },
     +  {
     +    "instance": "agent-environment-marker-v1/invalid-seed-links-on-linked-home.json",
    -+    "schema": "agent-environment-marker-v1.schema.json",
    -+    "valid": false
    -+  },
    +     "schema": "agent-environment-marker-v1.schema.json",
    +     "valid": false
    +   },
     +  {
     +    "instance": "agent-environment-marker-v1/invalid-seeded-projects-on-copied.json",
     +    "schema": "agent-environment-marker-v1.schema.json",
    @@ conformance/v1/schema-cases/index.json
     +  },
     +  {
     +    "instance": "agent-environment-marker-v1/invalid-seeded-projects-unsorted.json",
    -     "schema": "agent-environment-marker-v1.schema.json",
    -     "valid": false
    -   },
    ++    "schema": "agent-environment-marker-v1.schema.json",
    ++    "valid": false
    ++  },
     +  {
     +    "instance": "agent-mcp-v1/valid.json",
     +    "schema": "agent-mcp-v1.schema.json",
    @@ conformance/v1/schema-cases/index.json
     -    "instance": "context-manifest-v1/valid-empty-modules.json",
     -    "schema": "context-manifest-v1.schema.json",
     +    "instance": "context-lock-v1/valid-single-root.json",
    ++    "schema": "context-lock-v1.schema.json",
    ++    "valid": true
    ++  },
    ++  {
    ++    "instance": "context-lock-v1/valid-path-root.json",
     +    "schema": "context-lock-v1.schema.json",
          "valid": true
        },
        {
     -    "instance": "context-manifest-v1/valid-system-selector.json",
     -    "schema": "context-manifest-v1.schema.json",
    -+    "instance": "context-lock-v1/valid-path-root.json",
    ++    "instance": "context-lock-v1/valid-unversioned-skill.json",
     +    "schema": "context-lock-v1.schema.json",
          "valid": true
        },
        {
     -    "instance": "context-manifest-v1/invalid-version.json",
     -    "schema": "context-manifest-v1.schema.json",
    -+    "instance": "context-lock-v1/valid-unversioned-skill.json",
    -+    "schema": "context-lock-v1.schema.json",
    -+    "valid": true
    -+  },
    -+  {
     +    "instance": "context-lock-v1/invalid-schema-version.json",
     +    "schema": "context-lock-v1.schema.json",
     +    "valid": false
    @@ conformance/v1/schema-cases/index.json
     +  {
     +    "instance": "context-lock-v1/invalid-member-no-pin.json",
     +    "schema": "context-lock-v1.schema.json",
    -     "valid": false
    -   },
    -   {
    --    "instance": "context-manifest-v1/invalid-unknown-entry-field.json",
    --    "schema": "context-manifest-v1.schema.json",
    ++    "valid": false
    ++  },
    ++  {
     +    "instance": "context-lock-v1/invalid-path-member-with-source.json",
     +    "schema": "context-lock-v1.schema.json",
    -     "valid": false
    -   },
    -   {
    --    "instance": "context-manifest-v1/invalid-empty-environments.json",
    --    "schema": "context-manifest-v1.schema.json",
    ++    "valid": false
    ++  },
    ++  {
     +    "instance": "context-lock-v1/invalid-path-member-with-directory.json",
     +    "schema": "context-lock-v1.schema.json",
    -     "valid": false
    -   },
    -   {
    --    "instance": "context-manifest-v1/invalid-duplicate-environments.json",
    --    "schema": "context-manifest-v1.schema.json",
    ++    "valid": false
    ++  },
    ++  {
     +    "instance": "context-lock-v1/invalid-path-member-skill-kind.json",
     +    "schema": "context-lock-v1.schema.json",
    -     "valid": false
    -   },
    -   {
    --    "instance": "context-manifest-v1/invalid-unknown-class.json",
    --    "schema": "context-manifest-v1.schema.json",
    ++    "valid": false
    ++  },
    ++  {
     +    "instance": "context-lock-v1/invalid-context-without-version.json",
     +    "schema": "context-lock-v1.schema.json",
    -     "valid": false
    -   },
    -   {
    --    "instance": "context-manifest-v1/invalid-parent-path.json",
    --    "schema": "context-manifest-v1.schema.json",
    ++    "valid": false
    ++  },
    ++  {
     +    "instance": "context-lock-v1/invalid-skill-with-directory.json",
     +    "schema": "context-lock-v1.schema.json",
    -     "valid": false
    -   },
    -   {
    --    "instance": "context-manifest-v1/invalid-duplicate-path.json",
    --    "schema": "context-manifest-v1.schema.json",
    ++    "valid": false
    ++  },
    ++  {
     +    "instance": "context-lock-v1/invalid-skill-overlay.json",
     +    "schema": "context-lock-v1.schema.json",
     +    "valid": false
    @@ conformance/v1/schema-cases/index.json
     +  {
     +    "instance": "context-lock-v1/invalid-required-by-unknown-member.json",
     +    "schema": "context-lock-v1.schema.json",
    -+    "valid": false
    -+  },
    -+  {
    +     "valid": false
    +   },
    +   {
    +-    "instance": "context-manifest-v1/invalid-unknown-entry-field.json",
    +-    "schema": "context-manifest-v1.schema.json",
     +    "instance": "context-lock-v1/invalid-root-not-a-member.json",
     +    "schema": "context-lock-v1.schema.json",
    -+    "valid": false
    -+  },
    -+  {
    +     "valid": false
    +   },
    +   {
    +-    "instance": "context-manifest-v1/invalid-empty-environments.json",
    +-    "schema": "context-manifest-v1.schema.json",
     +    "instance": "context-lock-v1/invalid-root-with-requirers.json",
     +    "schema": "context-lock-v1.schema.json",
    -+    "valid": false
    -+  },
    -+  {
    +     "valid": false
    +   },
    +   {
    +-    "instance": "context-manifest-v1/invalid-duplicate-environments.json",
    +-    "schema": "context-manifest-v1.schema.json",
     +    "instance": "context-lock-v1/invalid-root-flagged-overlay.json",
     +    "schema": "context-lock-v1.schema.json",
    -+    "valid": false
    -+  },
    -+  {
    +     "valid": false
    +   },
    +   {
    +-    "instance": "context-manifest-v1/invalid-unknown-class.json",
    +-    "schema": "context-manifest-v1.schema.json",
     +    "instance": "context-lock-v1/invalid-members-unsorted-by-kind.json",
     +    "schema": "context-lock-v1.schema.json",
    -+    "valid": false
    -+  },
    -+  {
    +     "valid": false
    +   },
    +   {
    +-    "instance": "context-manifest-v1/invalid-parent-path.json",
    +-    "schema": "context-manifest-v1.schema.json",
     +    "instance": "context-lock-v1/invalid-members-unsorted-by-name.json",
     +    "schema": "context-lock-v1.schema.json",
    -+    "valid": false
    -+  },
    -+  {
    +     "valid": false
    +   },
    +   {
    +-    "instance": "context-manifest-v1/invalid-duplicate-path.json",
    +-    "schema": "context-manifest-v1.schema.json",
     +    "instance": "context-lock-v1/invalid-members-duplicate.json",
     +    "schema": "context-lock-v1.schema.json",
          "valid": false
    @@ conformance/v1/schema-cases/index.json
        {
     -    "instance": "launch-env-fragment-v1/valid-local-state-pin.json",
     +    "instance": "launch-env-fragment-v1/valid-minimal.json",
    -+    "schema": "launch-env-fragment-v1.schema.json",
    -+    "valid": true
    -+  },
    -+  {
    -+    "instance": "launch-env-fragment-v1/valid-system-prompt-only.json",
    -+    "schema": "launch-env-fragment-v1.schema.json",
    -+    "valid": true
    -+  },
    -+  {
    -+    "instance": "launch-env-fragment-v1/valid-mcp-only.json",
          "schema": "launch-env-fragment-v1.schema.json",
          "valid": true
        },
        {
     -    "instance": "launch-env-fragment-v1/valid-no-composition.json",
    -+    "instance": "launch-env-fragment-v1/valid-codex-config-key-and-name-channel.json",
    ++    "instance": "launch-env-fragment-v1/valid-system-prompt-only.json",
          "schema": "launch-env-fragment-v1.schema.json",
          "valid": true
        },
        {
     -    "instance": "launch-env-fragment-v1/valid-config-key-channel.json",
    -+    "instance": "launch-env-fragment-v1/valid-opencode-empty-channels-and-variable.json",
    ++    "instance": "launch-env-fragment-v1/valid-mcp-only.json",
          "schema": "launch-env-fragment-v1.schema.json",
          "valid": true
        },
        {
     -    "instance": "launch-env-fragment-v1/valid-empty-channels.json",
    -+    "instance": "launch-env-fragment-v1/valid-pi-file-channels.json",
    ++    "instance": "launch-env-fragment-v1/valid-codex-config-key-and-name-channel.json",
          "schema": "launch-env-fragment-v1.schema.json",
          "valid": true
        },
        {
     -    "instance": "launch-env-fragment-v1/valid-file-channels.json",
    ++    "instance": "launch-env-fragment-v1/valid-opencode-empty-channels-and-variable.json",
    ++    "schema": "launch-env-fragment-v1.schema.json",
    ++    "valid": true
    ++  },
    ++  {
    ++    "instance": "launch-env-fragment-v1/valid-pi-file-channels.json",
    ++    "schema": "launch-env-fragment-v1.schema.json",
    ++    "valid": true
    ++  },
    ++  {
     +    "instance": "launch-env-fragment-v1/valid-winner-first.json",
     +    "schema": "launch-env-fragment-v1.schema.json",
     +    "valid": true
    @@ conformance/v1/schema-cases/index.json
        {
     -    "instance": "launch-env-fragment-v1/invalid-profile-both-pins.json",
     +    "instance": "launch-env-fragment-v1/invalid-missing-precedence.json",
    -     "schema": "launch-env-fragment-v1.schema.json",
    -     "valid": false
    -   },
    -   {
    --    "instance": "launch-env-fragment-v1/invalid-composition-without-precedence.json",
    ++    "schema": "launch-env-fragment-v1.schema.json",
    ++    "valid": false
    ++  },
    ++  {
     +    "instance": "launch-env-fragment-v1/invalid-precedence-string.json",
    -     "schema": "launch-env-fragment-v1.schema.json",
    -     "valid": false
    -   },
    -   {
    --    "instance": "launch-env-fragment-v1/invalid-precedence-without-composition.json",
    ++    "schema": "launch-env-fragment-v1.schema.json",
    ++    "valid": false
    ++  },
    ++  {
     +    "instance": "launch-env-fragment-v1/invalid-profile-commit-pin.json",
    -     "schema": "launch-env-fragment-v1.schema.json",
    -     "valid": false
    -   },
    -   {
    --    "instance": "launch-env-fragment-v1/invalid-unknown-channel-kind.json",
    ++    "schema": "launch-env-fragment-v1.schema.json",
    ++    "valid": false
    ++  },
    ++  {
     +    "instance": "launch-env-fragment-v1/invalid-profile-lock-prefixed.json",
    -     "schema": "launch-env-fragment-v1.schema.json",
    -     "valid": false
    -   },
    -   {
    --    "instance": "launch-env-fragment-v1/invalid-unknown-semantics.json",
    ++    "schema": "launch-env-fragment-v1.schema.json",
    ++    "valid": false
    ++  },
    ++  {
     +    "instance": "launch-env-fragment-v1/invalid-profile-unknown-field.json",
    -     "schema": "launch-env-fragment-v1.schema.json",
    -     "valid": false
    -   },
    -   {
    --    "instance": "launch-env-fragment-v1/invalid-flag-channel-with-filename.json",
    ++    "schema": "launch-env-fragment-v1.schema.json",
    ++    "valid": false
    ++  },
    ++  {
     +    "instance": "launch-env-fragment-v1/invalid-env-empty.json",
    -     "schema": "launch-env-fragment-v1.schema.json",
    -     "valid": false
    -   },
    -   {
    --    "instance": "launch-env-fragment-v1/invalid-lowercase-variable-name.json",
    ++    "schema": "launch-env-fragment-v1.schema.json",
    ++    "valid": false
    ++  },
    ++  {
     +    "instance": "launch-env-fragment-v1/invalid-env-lowercase-variable.json",
    -     "schema": "launch-env-fragment-v1.schema.json",
    -     "valid": false
    -   },
    -   {
    --    "instance": "launch-env-fragment-v1/invalid-empty-env.json",
    ++    "schema": "launch-env-fragment-v1.schema.json",
    ++    "valid": false
    ++  },
    ++  {
     +    "instance": "launch-env-fragment-v1/invalid-env-wrong-adapter-variable.json",
    -     "schema": "launch-env-fragment-v1.schema.json",
    -     "valid": false
    -   },
    -   {
    --    "instance": "log-response-v1/valid.json",
    --    "schema": "log-response-v1.schema.json",
    --    "valid": true
    ++    "schema": "launch-env-fragment-v1.schema.json",
    ++    "valid": false
    ++  },
    ++  {
     +    "instance": "launch-env-fragment-v1/invalid-env-two-variables.json",
     +    "schema": "launch-env-fragment-v1.schema.json",
     +    "valid": false
    -   },
    -   {
    --    "instance": "log-response-v1/invalid.json",
    --    "schema": "log-response-v1.schema.json",
    ++  },
    ++  {
     +    "instance": "launch-env-fragment-v1/invalid-env-relative-path.json",
     +    "schema": "launch-env-fragment-v1.schema.json",
    -     "valid": false
    -   },
    -   {
    --    "instance": "manager-config-v1/valid.json",
    --    "schema": "manager-config-v1.schema.json",
    --    "valid": true
    ++    "valid": false
    ++  },
    ++  {
     +    "instance": "launch-env-fragment-v1/invalid-system-prompt-unknown-field.json",
     +    "schema": "launch-env-fragment-v1.schema.json",
     +    "valid": false
    -   },
    -   {
    --    "instance": "manager-config-v1/invalid.json",
    --    "schema": "manager-config-v1.schema.json",
    ++  },
    ++  {
     +    "instance": "launch-env-fragment-v1/invalid-system-prompt-missing-channels.json",
     +    "schema": "launch-env-fragment-v1.schema.json",
    -     "valid": false
    -   },
    -   {
    --    "instance": "profilefile-v1/valid.json",
    --    "schema": "profilefile-v1.schema.json",
    --    "valid": true
    ++    "valid": false
    ++  },
    ++  {
     +    "instance": "launch-env-fragment-v1/invalid-system-prompt-channels-not-registry.json",
     +    "schema": "launch-env-fragment-v1.schema.json",
     +    "valid": false
    -   },
    -   {
    --    "instance": "profilefile-v1/invalid.json",
    --    "schema": "profilefile-v1.schema.json",
    ++  },
    ++  {
     +    "instance": "launch-env-fragment-v1/invalid-channel-unknown-kind.json",
     +    "schema": "launch-env-fragment-v1.schema.json",
    -     "valid": false
    -   },
    -   {
    --    "instance": "profilefile-v1/valid-single-profile.json",
    --    "schema": "profilefile-v1.schema.json",
    --    "valid": true
    ++    "valid": false
    ++  },
    ++  {
     +    "instance": "launch-env-fragment-v1/invalid-channel-unknown-semantics.json",
     +    "schema": "launch-env-fragment-v1.schema.json",
     +    "valid": false
    @@ conformance/v1/schema-cases/index.json
     +    "instance": "launch-env-fragment-v1/invalid-flag-name-without-name-argument.json",
     +    "schema": "launch-env-fragment-v1.schema.json",
     +    "valid": false
    -   },
    -   {
    --    "instance": "profilefile-v1/invalid-version.json",
    --    "schema": "profilefile-v1.schema.json",
    ++  },
    ++  {
     +    "instance": "launch-env-fragment-v1/invalid-flag-name-argument-without-name.json",
     +    "schema": "launch-env-fragment-v1.schema.json",
     +    "valid": false
    @@ conformance/v1/schema-cases/index.json
     +  },
     +  {
     +    "instance": "launch-env-fragment-v1/invalid-mcp-channel-not-registry.json",
    -+    "schema": "launch-env-fragment-v1.schema.json",
    -+    "valid": false
    -+  },
    -+  {
    +     "schema": "launch-env-fragment-v1.schema.json",
    +     "valid": false
    +   },
    +   {
    +-    "instance": "launch-env-fragment-v1/invalid-composition-without-precedence.json",
     +    "instance": "launch-env-fragment-v1/invalid-mcp-unknown-field.json",
    -+    "schema": "launch-env-fragment-v1.schema.json",
    -+    "valid": false
    -+  },
    -+  {
    +     "schema": "launch-env-fragment-v1.schema.json",
    +     "valid": false
    +   },
    +   {
    +-    "instance": "launch-env-fragment-v1/invalid-precedence-without-composition.json",
     +    "instance": "launch-env-fragment-v1/invalid-mcp-missing-env-names.json",
    -+    "schema": "launch-env-fragment-v1.schema.json",
    +     "schema": "launch-env-fragment-v1.schema.json",
          "valid": false
        },
        {
    --    "instance": "profilefile-v1/invalid-unknown-field.json",
    --    "schema": "profilefile-v1.schema.json",
    +-    "instance": "launch-env-fragment-v1/invalid-unknown-channel-kind.json",
     +    "instance": "launch-env-fragment-v1/invalid-mcp-env-names-unsorted.json",
    -+    "schema": "launch-env-fragment-v1.schema.json",
    -+    "valid": false
    -+  },
    -+  {
    +     "schema": "launch-env-fragment-v1.schema.json",
    +     "valid": false
    +   },
    +   {
    +-    "instance": "launch-env-fragment-v1/invalid-unknown-semantics.json",
     +    "instance": "launch-env-fragment-v1/invalid-mcp-env-names-reserved.json",
    -+    "schema": "launch-env-fragment-v1.schema.json",
    +     "schema": "launch-env-fragment-v1.schema.json",
          "valid": false
        },
        {
    --    "instance": "profilefile-v1/invalid-profile-name.json",
    --    "schema": "profilefile-v1.schema.json",
    +-    "instance": "launch-env-fragment-v1/invalid-flag-channel-with-filename.json",
     +    "instance": "launch-env-fragment-v1/invalid-mcp-two-channels.json",
    -+    "schema": "launch-env-fragment-v1.schema.json",
    +     "schema": "launch-env-fragment-v1.schema.json",
          "valid": false
        },
        {
    --    "instance": "profilefile-v1/invalid-traversal-path.json",
    --    "schema": "profilefile-v1.schema.json",
    +-    "instance": "launch-env-fragment-v1/invalid-lowercase-variable-name.json",
     +    "instance": "launch-env-fragment-v1/invalid-mcp-on-pi.json",
    -+    "schema": "launch-env-fragment-v1.schema.json",
    +     "schema": "launch-env-fragment-v1.schema.json",
          "valid": false
        },
        {
    --    "instance": "profilefile-v1/invalid-duplicate-directory.json",
    --    "schema": "profilefile-v1.schema.json",
    +-    "instance": "launch-env-fragment-v1/invalid-empty-env.json",
     +    "instance": "launch-env-fragment-v1/invalid-path-prepend-relative.json",
     +    "schema": "launch-env-fragment-v1.schema.json",
    -     "valid": false
    -   },
    -   {
    --    "instance": "profilefile-v1/invalid-nested-root.json",
    --    "schema": "profilefile-v1.schema.json",
    -+    "instance": "launch-env-fragment-v1/invalid-path-prepend-outside-root.json",
    -+    "schema": "launch-env-fragment-v1.schema.json",
    -+    "valid": false
    -+  },
    -+  {
    -+    "instance": "log-response-v1/valid.json",
    -+    "schema": "log-response-v1.schema.json",
    -+    "valid": true
    -+  },
    -+  {
    -+    "instance": "log-response-v1/invalid.json",
    -+    "schema": "log-response-v1.schema.json",
     +    "valid": false
     +  },
     +  {
    -+    "instance": "manager-config-v1/valid.json",
    -+    "schema": "manager-config-v1.schema.json",
    -+    "valid": true
    -+  },
    -+  {
    -+    "instance": "manager-config-v1/invalid.json",
    -+    "schema": "manager-config-v1.schema.json",
    ++    "instance": "launch-env-fragment-v1/invalid-path-prepend-outside-root.json",
    +     "schema": "launch-env-fragment-v1.schema.json",
    +     "valid": false
    +   },
    +@@
    +     "schema": "manager-config-v2.schema.json",
          "valid": false
        },
    +-  {
    +-    "instance": "profilefile-v1/valid.json",
    +-    "schema": "profilefile-v1.schema.json",
    +-    "valid": true
    +-  },
    +-  {
    +-    "instance": "profilefile-v1/invalid.json",
    +-    "schema": "profilefile-v1.schema.json",
    +-    "valid": false
    +-  },
    +-  {
    +-    "instance": "profilefile-v1/valid-single-profile.json",
    +-    "schema": "profilefile-v1.schema.json",
    +-    "valid": true
    +-  },
    +-  {
    +-    "instance": "profilefile-v1/invalid-version.json",
    +-    "schema": "profilefile-v1.schema.json",
    +-    "valid": false
    +-  },
    +-  {
    +-    "instance": "profilefile-v1/invalid-unknown-field.json",
    +-    "schema": "profilefile-v1.schema.json",
    +-    "valid": false
    +-  },
    +-  {
    +-    "instance": "profilefile-v1/invalid-profile-name.json",
    +-    "schema": "profilefile-v1.schema.json",
    +-    "valid": false
    +-  },
    +-  {
    +-    "instance": "profilefile-v1/invalid-traversal-path.json",
    +-    "schema": "profilefile-v1.schema.json",
    +-    "valid": false
    +-  },
    +-  {
    +-    "instance": "profilefile-v1/invalid-duplicate-directory.json",
    +-    "schema": "profilefile-v1.schema.json",
    +-    "valid": false
    +-  },
    +-  {
    +-    "instance": "profilefile-v1/invalid-nested-root.json",
    +-    "schema": "profilefile-v1.schema.json",
    +-    "valid": false
    +-  },
        {
    +     "instance": "provider-capability-receipt-v1/valid.json",
    +     "schema": "provider-capability-receipt-v1.schema.json",
     
      ## conformance/v1/schema-cases/launch-env-fragment-v1/invalid-channel-missing-semantics.json (new) ##
     @@
    @@ release/1.0.0-rc.9.json
          "verified_provider_contract": "host-execution-provider-v1"
        },
        "candidate_protocol_pin": {
    --    "manifest_sha256": "sha256:ab25038b10612d77ca36429daaeb857f27d6eb449e3acd41379da4381ae8ba91",
    -+    "manifest_sha256": "sha256:0aa2899dca27e596eab47f4f5f64706f8d8200c0e943bf084dbadba0aa2e8d47",
    +-    "manifest_sha256": "sha256:695d2ecf0935a132b326143d2964f2cd0d1efe2f84bbe76f1eb5636fec34cc81",
    ++    "manifest_sha256": "sha256:dd25de1553f049866bc80d52edd280070d00a955e90e945d1051e2b3c94bf656",
          "suite_root": "conformance/v1"
        },
        "claim_v5": {
    @@ release/1.0.0-rc.9.json
        "downstream_consumption": {
          "committed_release_pin_advanced": false,
          "environment": "CURATOR_CONFORMANCE_ROOT",
    --    "required_manifest_sha256": "sha256:ab25038b10612d77ca36429daaeb857f27d6eb449e3acd41379da4381ae8ba91"
    -+    "required_manifest_sha256": "sha256:0aa2899dca27e596eab47f4f5f64706f8d8200c0e943bf084dbadba0aa2e8d47"
    +-    "required_manifest_sha256": "sha256:695d2ecf0935a132b326143d2964f2cd0d1efe2f84bbe76f1eb5636fec34cc81"
    ++    "required_manifest_sha256": "sha256:dd25de1553f049866bc80d52edd280070d00a955e90e945d1051e2b3c94bf656"
        },
        "historical_release": {
          "immutable": true,
    @@ tools/generate-vectors/context_detectors.go (new)
     +			"context/00-base.md": "# Base\n\nLegacy: " + awsKey + "\n",
     +		}, nil},
     +		{"system-module-present", "every class: system module is reported with its package, path, and selector as a warning that never blocks", "context", pin, false, map[string]string{
    -+			"agent-context.json": contextManifestText("companyA", map[string]any{"path": "00-base.md"}, map[string]any{"path": "90-system.md", "class": "system", "environments": []any{"claude_code"}}, map[string]any{"path": "95-review.md", "class": "system"}),
    -+			"context/00-base.md": "# Base\n",
    ++			"agent-context.json":   contextManifestText("companyA", map[string]any{"path": "00-base.md"}, map[string]any{"path": "90-system.md", "class": "system", "environments": []any{"claude_code"}}, map[string]any{"path": "95-review.md", "class": "system"}),
    ++			"context/00-base.md":   "# Base\n",
     +			"context/90-system.md": "You are the companyA reviewer.\n",
     +			"context/95-review.md": "Prefer short answers.\n",
     +		}, nil},
    @@ tools/generate-vectors/context_versions.go (new)
     +
     +type semver struct {
     +	major, minor, patch int
    -+	prerelease         []string
    ++	prerelease          []string
     +}
     +
     +func (v semver) String() string {
    @@ tools/generate-vectors/context_versions.go (new)
     +		"fonts":  {Kind: "skill", Source: "github.com/example/skill-fonts", Tags: map[string]string{"v1.0.0": commitFor("fonts", "1.0.0"), "v1.1.0": commitFor("fonts", "1.1.0"), "release-1.1": commitFor("fonts", "1.1.0"), "v1.2.0": commitFor("fonts", "1.2.0")}, Commits: map[string]*resolutionManifest{}},
     +	})})
     +	cases = append(cases, resolutionCase{"weight-conflict", "two direct requirers disagree on a member's edge weight and the root's weights map does not name it: context_weight_conflict", simpleInput("root", "*", map[string]*resolutionPackage{
    -+		"root": ctxPackage("root", map[string]*resolutionManifest{"1.0.0": plainManifest(0, contextRequirement("a", "range", "^1", nil), contextRequirement("b", "range", "^1", nil))}),
    -+		"a":    ctxPackage("a", map[string]*resolutionManifest{"1.0.0": plainManifest(0, contextRequirement("shared", "range", "^1", intPointer(30)))}),
    -+		"b":    ctxPackage("b", map[string]*resolutionManifest{"1.0.0": plainManifest(0, contextRequirement("shared", "range", "^1", intPointer(50)))}),
    ++		"root":   ctxPackage("root", map[string]*resolutionManifest{"1.0.0": plainManifest(0, contextRequirement("a", "range", "^1", nil), contextRequirement("b", "range", "^1", nil))}),
    ++		"a":      ctxPackage("a", map[string]*resolutionManifest{"1.0.0": plainManifest(0, contextRequirement("shared", "range", "^1", intPointer(30)))}),
    ++		"b":      ctxPackage("b", map[string]*resolutionManifest{"1.0.0": plainManifest(0, contextRequirement("shared", "range", "^1", intPointer(50)))}),
     +		"shared": ctxPackage("shared", map[string]*resolutionManifest{"1.0.0": {Weight: 5}}),
     +	})})
     +	cases = append(cases, resolutionCase{"weight-conflict-root-map-wins", "the same disagreement is a warning when the root's weights map names the member; the root has the final word", simpleInput("root", "*", map[string]*resolutionPackage{
    -+		"root": ctxPackage("root", map[string]*resolutionManifest{"1.0.0": {Weights: map[string]int{"shared": 70}, Requires: []resolutionRequirement{contextRequirement("a", "range", "^1", nil), contextRequirement("b", "range", "^1", nil)}}}),
    -+		"a":    ctxPackage("a", map[string]*resolutionManifest{"1.0.0": plainManifest(0, contextRequirement("shared", "range", "^1", intPointer(30)))}),
    -+		"b":    ctxPackage("b", map[string]*resolutionManifest{"1.0.0": plainManifest(0, contextRequirement("shared", "range", "^1", intPointer(50)))}),
    ++		"root":   ctxPackage("root", map[string]*resolutionManifest{"1.0.0": {Weights: map[string]int{"shared": 70}, Requires: []resolutionRequirement{contextRequirement("a", "range", "^1", nil), contextRequirement("b", "range", "^1", nil)}}}),
    ++		"a":      ctxPackage("a", map[string]*resolutionManifest{"1.0.0": plainManifest(0, contextRequirement("shared", "range", "^1", intPointer(30)))}),
    ++		"b":      ctxPackage("b", map[string]*resolutionManifest{"1.0.0": plainManifest(0, contextRequirement("shared", "range", "^1", intPointer(50)))}),
     +		"shared": ctxPackage("shared", map[string]*resolutionManifest{"1.0.0": {Weight: 5}}),
     +	})})
     +	cases = append(cases, resolutionCase{"weights-not-root", "a non-root member carrying a non-empty weights map fails context_weights_not_root at resolution", simpleInput("root", "*", map[string]*resolutionPackage{
    @@ tools/generate-vectors/environments.go: func writeEnvironmentVectors(dir, expect
     +		{name: "invalid-source-with-git-suffix", valid: false, instance: withLockMember(valid, 4, func(m map[string]any) { m["source"] = "github.com/companyA/root-context-ios-developer-umbrella.git" })},
     +		{name: "invalid-source-uppercase-host", valid: false, instance: withLockMember(valid, 4, func(m map[string]any) { m["source"] = "GitHub.com/companyA/root-context-ios-developer-umbrella" })},
     +		{name: "invalid-weight-negative", valid: false, instance: withLockMember(valid, 4, func(m map[string]any) { m["weight"] = -1 })},
    -+		{name: "invalid-required-by-duplicate", valid: false, instance: withLockMember(valid, 0, func(m map[string]any) { m["required_by"] = []any{"companyA-root-context-core", "companyA-root-context-core"} })},
    ++		{name: "invalid-required-by-duplicate", valid: false, instance: withLockMember(valid, 0, func(m map[string]any) {
    ++			m["required_by"] = []any{"companyA-root-context-core", "companyA-root-context-core"}
    ++		})},
     +		{name: "invalid-required-by-unsorted", valid: false, instance: withLockMember(valid, 0, func(m map[string]any) {
     +			m["required_by"] = []any{"companyA-root-context-ios-developer-umbrella", "companyA-root-context-developers-core"}
     +		})},
    @@ tools/validate.py: def main() -> int:
     +        validate_context_version_vectors,
     +        validate_context_detector_vectors,
              validate_snapshot_acquisition_vectors,
    +         validate_manager_config_vectors,
              validate_local_links,
    -     ]
```
