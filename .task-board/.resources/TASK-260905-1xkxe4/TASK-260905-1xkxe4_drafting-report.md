# TASK-260905-1xkxe4 drafting report: environments 1.1 batch 2 — schemas, cases, vectors, generator/validator

Worktree: `curator-spec/.temp/STORY-260905-1z93ju/worktree`, branch
`task-board/story/STORY-260905-1z93ju`, base `a68559b` (environments.md
revision 1.1). One signed commit, no push, no tag, no PR.

## Commit and signature

| Item | Value |
| --- | --- |
| Commit | `401b665ba964bc951092100d48c75be31a047788` (squash of the earlier `5da9b95` delivery and `28e442c` LOGBOOK commits; tree identical to `28e442c`) |
| Subject | Deliver the environments 1.1 section 13 schemas, cases, and vector families |
| Why re-cut | The board change request requires exactly one single-parent commit past checkpoint `a68559b`; run RUN-260905-901507 left two. Squashed with `git reset --soft a68559b` + `git commit -S`; `git rev-parse HEAD^{tree}` equals the old head's tree. |
| Author | Ivan Oparin <oparin@me.com> |
| Signature | `Good "git" signature for oparin@me.com with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM` — `%G?` = `G` against `maintainers.allowed_signers` |
| Diff | 255 files changed (254 + LOGBOOK.md) |
| Gates at 401b665 | `make validate` exit 0 (186 Python tests OK, go test ok) — `.temp/make-validate-02.log`; `make regenerate-check` exit 0 — `.temp/make-regenerate-check-03.log` |

The global `gpg.ssh.allowedSignersFile` on this machine points at a stale
`/private/tmp/curator-spec-rc8-verify.*` path, so a bare `git log
--show-signature` prints "No principal matched". Verification with
`-c gpg.ssh.allowedSignersFile=maintainers.allowed_signers` is the evidence
above.

## Section 13 surface → files produced

| §13 surface | Files |
| --- | --- |
| `agent-context-v1` (§2) | `schemas/v1/agent-context-v1.schema.json` (new); `conformance/v1/schema-cases/agent-context-v1/` (38 cases: 7 valid, 31 invalid) |
| `agent-mcp-v1` (§2.2) | `schemas/v1/agent-mcp-v1.schema.json` (new); `schema-cases/agent-mcp-v1/` (30: 5 valid, 25 invalid) |
| `context-lock-v1` (§1.3) | `schemas/v1/context-lock-v1.schema.json` (new); `schema-cases/context-lock-v1/` (30: 4 valid, 26 invalid); `valid.json` is the Decision 0012 §9 lock as the resolver produces it |
| rewritten `agent-environment-marker-v1` (§8.2) | `schemas/v1/agent-environment-marker-v1.schema.json` (rewritten in place); `schema-cases/agent-environment-marker-v1/` (61: 12 valid, 49 invalid) |
| rewritten `launch-env-fragment-v1` (§10.2) | `schemas/v1/launch-env-fragment-v1.schema.json` (rewritten in place: `argument` required on every `flag` descriptor, `name` exactly when `argument` is `name`, `precedence` object, `mcp` section, `path_prepend`); `schema-cases/launch-env-fragment-v1/` (59: 14 valid, 45 invalid) |
| withdrawn `profilefile-v1`, `context-manifest-v1` | schemas and case directories deleted; `schemas/v1/README.md` updated |
| version and range parsing (§1.4, coercion table, excluded forms) | `conformance/v1/vectors/context-versions.json`: `version_cases` (15), `ordering_cases` (1), `range_cases` (50), `satisfies_cases` (63) |
| resolution (conflict, downward re-selection, prerelease admission, exact-constraint unification, `\|\|` highest member, `latest` = `*`) | `context-versions.json`: `resolution_cases` (23, listed below) |
| lock canonicalization and hashing | `context-versions.json`: `lock_cases` (3: minimal, path root with skill, worked example) with `ccj1_bytes`, `byte_length`, `lock_sha256` |
| §5 materialization bytes under both `winner` and both `placement` primitives | `conformance/v1/vectors/environments.json` (4 header cases, 19 materialization cases) and `conformance/v1/expected/environments/*` (16 directories) |
| §5.6 hash binding | `surface_sha256` per materialization case, recomputed by the validator |
| MCP materialization bytes per adapter (§5.8) | `expected/environments/mcp-claude-code/.agent-context/mcp/claude_code.json`, `mcp-codex-cli/curator-mcp.config.toml`, `mcp-opencode/.agent-context/mcp/opencode.json`; `mcp-pi-none` writes nothing |
| detector classes (§9.1) | `conformance/v1/vectors/context-detectors.json` (12 cases) |
| §1.2 snapshot byte-exactness vector | untouched (`vectors/snapshot-acquisition.json`, `fixtures/byte-exact`, `expected/byte-exact-snapshot_sha256.txt`) |
| manifest and rc.9 pins | `conformance/v1/manifest.json` and `release/1.0.0-rc.9.json` regenerated; the rc.9 diff is the two `manifest_sha256` lines only |
| docs | `CHANGELOG.md` (Unreleased: Added, Removed), `conformance/README.md` bullet, `schemas/v1/README.md` paragraph |

The nine retired `expected/environments/*` sets keep their names and are
regenerated under `curator-root-context-v2`; `monolithic-composed-empty-chapter`
is re-cut as `monolithic-composed-no-chapter` (the `emptyoverlay` member
appears as a `member:` line and contributes no chapter). New sets:
`weights-winner-{higher,lower}-placement-{last,first}` (six-member closure
with a weight tie between `figma` and `ios` that no primitive inverts) and
`mcp-{claude-code,codex-cli,opencode}`.

`.github/ci/implementation-coverage.tsv` was not edited: no row names an
environments artifact, and `tools/implementation_coverage.py families` only
requires that declared artifacts stay published.

### Resolution cases

`worked-example-default-policy` (Decision 0012 §9, lock identical to the
decision's listing), `range-conflict-empty-intersection`,
`downward-reselection`, `selection-never-increases`, `prerelease-admission`,
`prerelease-excluded-by-latest`, `exact-constraint-unification`,
`exact-constraints-disagree`, `exact-outside-range`, `or-highest-member`,
`latest-is-star`, `no-version-tags`, `non-version-tag-exact`,
`version-mismatch`, `skill-exact-dependency`, `weight-conflict`,
`weight-conflict-root-map-wins` (warning; root map value wins),
`weights-not-root`, `weights-duplicate`, `weight-unknown`,
`overlay-joint-resolution-conflict`, `overlay-git-explicit-weight`,
`overlay-duplicate-name`.

## Generator and validator functions added

`tools/generate-vectors/context_versions.go` (new): `parseVersion`,
`parseTagVersion`, `compareVersions`, `parsePartial`, `desugarPrimitive`,
`parseRange`, `rangeSatisfies`, `resolveClosure` (the §1.4 four-step
algorithm with downward re-selection, ceilings, member departure, the §6
weight rules and every diagnostic), `lockHash`, `workedExampleInput`,
`resolutionCases`, `versionParseCases`, `versionOrderingCase`,
`rangeParseCases`, `satisfiesCases`, `lockCanonicalizationCases`,
`writeContextVersionVectors`.

`tools/generate-vectors/context_detectors.go` (new): `detectorPatterns`
(closed classes `aws-access-key-id`, `private-key-block`, `bearer-token`),
`detectSecretMaterial`, `systemModuleWarnings`, `detectorCases`,
`writeContextDetectorVectors`.

`tools/generate-vectors/environments.go` (rewritten): `environmentClosure`,
`environmentLock`, `environmentEmittedOrder` (core §7 Kahn order stably
sorted by weight under the four primitive pairs), `environmentHeader` (v2),
`environmentRootContextFiles`, `environmentSystemPromptFiles`,
`environmentMCPSet`, `environmentMCPFiles`, `tomlBasicString`,
`environmentEnvNames`, `validAgentContextV1`, `agentContextSchemaExamples`,
`validAgentMCPV1`, `agentMCPSchemaExamples`, `validContextLockV1`,
`contextLockSchemaExamples`, rewritten `validEnvironmentMarkerV1`,
`environmentMarkerSchemaExamples`, `adapterSystemPromptChannels`,
`adapterMCPChannel`, `launchFragmentFor`, `launchEnvFragmentSchemaExamples`.
`main.go`: the two new writers and the five schema-case registrations
(profilefile/context-manifest registrations removed).

Go tests: `environments_test.go` (rewritten), `context_versions_test.go`,
`context_detectors_test.go` — header literal, emitted order under all four
pairs with the tie check, no-chapter bytes, MCP byte literals per adapter,
lock ordering and hash, worked-example lock equality, every resolution case
outcome, the widened-range narrowing case, detector spans, waiver exactness
(a one-byte shift matches nothing), pin-does-not-clear.

`tools/validate.py` (independent Python implementation): `semver_parse`,
`semver_parse_tag`, `semver_compare`, `range_parse`, `comparator_text`,
`range_satisfies`, `resolve_closure`, `validate_context_version_vectors`,
`environment_emitted_order`, `environment_header_bytes`,
`environment_mcp_set`, `toml_basic_string`, `environment_mcp_files`,
`environment_case_files` (v2), `validate_environment_vectors` (v2: also
validates every case lock against `context-lock-v1`, recomputes
`lock_sha256`, `emitted_order`, byte lengths, `mcp_set`, `env_names`, the
LF discipline of every expected file), `detector_findings`,
`detector_system_module_warnings`, `validate_context_detector_vectors`;
`validate_wire_semantics` gains `agent-context-v1` (unique module paths),
`context-lock-v1` (sorted unique members, root is a context member with no
requirer and not an overlay, `required_by` sorted/unique/known),
`agent-environment-marker-v1` (sorted surfaces, unique members including the
root, sorted `seeded_projects`), `launch-env-fragment-v1` (channels equal the
closed registry descriptors, sorted `env_names`, `path_prepend` below the
environments root); the profilefile and context-manifest branches are gone.
`tools/test_validate.py`: `EnvironmentVectorTests` rewritten for v2 plus
`ContextVersionVectorTests` and `ContextDetectorVectorTests` (each mutates
one expectation and requires rejection; a hand-edited codex TOML with a
blank separator line fails).

## Gate tails

`make validate` (`.temp/make-validate-01.log`, exit 0):

```text
validated 58 schemas and 943 vector files
Ran 186 tests in 34.071s
OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	0.795s
```

`make regenerate-check` after the commit (`.temp/make-regenerate-check-02.log`, exit 0):

```text
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
```

The first `make regenerate-check` run (`.temp/make-regenerate-check-01.log`,
exit 2) ran before the commit and failed on `git diff --exit-code` because
the regenerated files were unstaged; it is not a regeneration difference.
Baseline before any change: `make validate` exit 0 (`.temp/validate-baseline-01.log`).
Python: venv `.temp/venv`, Python 3.14.7, jsonschema 4.25.1. Go: `go test
./tools/...` green.

## node-semver evidence

`node v25.6.1` with `semver@7.7.4` at
`/opt/homebrew/lib/node_modules/npm/node_modules/semver` (the version §1.4
records). Log: `.temp/node-semver-evidence-01.log`.

- All 63 `satisfies_cases` pairs agree with `semver.satisfies(version, range)`
  (`latest` passed as `*`): 0 mismatches.
- All 38 valid `range_cases` desugar to the same comparator sets as
  `new semver.Range(r).set`, with four spelling-only differences: node prints
  an exact primitive as `1.2.3`, the vector as `=1.2.3`; node prints `^0` as
  `<1.0.0-0` alone, the vector as `>=0.0.0 <1.0.0-0` exactly as §1.4 spells
  it. The any comparator is `""` in node and `*` in the vector.
- Of the 12 excluded forms, node itself rejects `1.2.3-` and `latest || ^1`;
  the rest (`1.2.3 - 2.3.4`, `v1.2.3`, `^v1`, `>=v1.0.0`, `""`, `||`,
  `^1 ||`, `1.2.3.4`, `^01.2`, `>>1`) node accepts or coerces and §1.4
  excludes; the vector records them as `profile_source_invalid`.

## Text readings recorded (no protocol edit made)

1. §5.8 `codex_cli` row: "one key per line, LF line endings, exactly one
   trailing LF, and no other bytes". Read literally: no blank line between
   `[mcp_servers.<name>]` tables. The vector bytes are
   `[mcp_servers.a]\ncommand = "…"\nargs = […]\n[mcp_servers.b]\n…`. If a
   blank separator was intended, the row needs an erratum; the byte rule as
   written admits none.
2. §7.8 `opencode` channel is `variable OPENCODE_CONFIG`; §7.3 defines
   `argument` for `flag` descriptors only, and §7.8 says "descriptor grammar
   with `argument` and `with`". Read: `argument`/`with` belong to `flag`
   descriptors; the opencode MCP descriptor is
   `{ "kind": "variable", "variable": "OPENCODE_CONFIG" }` with no `argument`.
   The fragment schema and the validator's closed registry table encode
   that reading.
3. §5 part sequence: a chapter part is emitted "for each `context` member in
   emitted order that has at least one applicable root module", with no
   single-member exception. Read: a lone root with modules also gets its
   `## Context:` chapter (the v1 vectors omitted chapters for a single
   profile). `monolithic-claude-code` now carries one chapter.
4. §1.4 "Two exact constraints on one name MUST peel to one commit" names
   no diagnostic of its own; the vector reports `context_range_conflict`
   ("a final constraint is unsatisfied", §1.1) for `exact-constraints-disagree`.
5. Core §7 "Cycles fail and name the cycle" names no diagnostic code, so no
   cycle case is vectored; the resolver panics on a context cycle in the
   materialization fixtures rather than inventing a code.
6. §8.2 marker `surfaces` "whether any entry is a copy rather than a link …
   each recorded with its reason": shaped as `copies: [{path, reason}]` with
   `reason` ∈ {`symlink-fallback`, `claude-code-root-context`}, required
   under `linked` and `managed-home`, forbidden under `copied`. §7.4
   passthrough strategies shaped as the enum {`per-home-keychain`,
   `file-link`, `keyring-preferred`, `ambient`, `in-place`}. Both are schema
   choices the text leaves open; review if another spelling is preferred.
7. §9.1 pattern classes are "closed and vectored" but unnamed in the text;
   the vector fixes three: `aws-access-key-id` (`AKIA[0-9A-Z]{16}`),
   `private-key-block` (PEM `BEGIN … PRIVATE KEY` header), `bearer-token`
   (`Bearer` + ≥20 token characters), with the placeholder rule "body after
   the class prefix is one repeated character or ends in EXAMPLE". Spans are
   byte offsets, start inclusive, end exclusive. A waiver that matches is
   reported as `context_secret_waiver_applied` (a spelling the text does not
   fix; it says "reported as a warning naming the waiver").
8. The fragment validator bounds `path_prepend` and `env` values to the
   `/manager/environments/` root used by every case; a real manager's root
   differs, and only the schema's absolute-path rule is normative.
9. `^0` per §1.4 is `>=0.0.0 <1.0.0-0`; node spells the same set as
   `<1.0.0-0`. Semantics identical (63/63 satisfies agree).

## Not delivered / out of scope

- No context-cycle resolution case (no diagnostic code to assert, item 5).
- No `mcp_package_not_allowed` resolution case: the allowlist is machine
  configuration, and the brief's resolution list does not name it.
- `protocol/environments.md` untouched, as required.
