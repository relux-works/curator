# Review findings cycle 1: environments 1.1 batch 2 — schemas, cases, vector families

Subject: `task-board/story/STORY-260905-1z93ju` at `401b665` (CR-TASK-260905-1xkxe4-1 rev 1, tree `1a4b0c3`), base `a68559b`. Reviewer run RUN-260905-de46bb. Read-only; scratch under `.temp/review/`.

## Verdict: ACCEPT (no blocking or major findings)

Gates re-run by the reviewer at `401b665` (not accepted from the producer's logs):

| Gate | Result | Evidence |
| --- | --- | --- |
| `make validate` | exit 0, `validated 58 schemas and 943 vector files`, 186 Python tests OK, `go test ./tools/...` ok | `.temp/review/make-validate-review.log` |
| `make regenerate-check` | exit 0 (generator output byte-identical to the commit) | `.temp/review/make-regenerate-check-review.log` |
| commit signature | `Good "git" signature with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM` | `git verify-commit 401b665` (principal lookup fails only because the machine's global allowed-signers path is stale, as the producer recorded) |
| M3 / hosted-lane files | `git diff a68559b --stat` over `fixtures/byte-exact`, `snapshot-acquisition.json`, `expected/byte-exact-snapshot_sha256.txt`, `manager-config.json`, `canonical-*.json`, `source-identities.json`, `identifiers.json`, `locale-selectors.json`, `expected/snapshot_sha256.txt` is empty | untouched |
| rc.9 pin | diff is exactly the two `manifest_sha256` lines | regenerated only |
| `implementation-coverage.tsv` | not edited; `implementation_coverage.py families` → 18 claims upheld | no new family claims pinned consumption |

## Dimension 1 — schemas (attacked with 76 reviewer-authored instances)

Script `.temp/review/schema-attack.py` (jsonschema 4.25.1, local `$id` registry). 61 of 76 behaved as expected; the 15 divergences were judged against the text:

Not findings (text admits them): nested module path `sub/00.md` (§3: "portable relative path … below `context/`"); SCP-form `git` (core §6.1 admits `[user@]host:path`; `skillfile-v1` types `git` the same way); lowercase `env_names` (§2.2: core §2 identifier grammar); unsorted package `env_names` (only the fragment union is sorted); uppercase ASCII host; uppercase package name (identifier grammar); `env` trailing slash.

Findings (all minor, none blocks acceptance):

| ID | Sev | File | What | Fix |
| --- | --- | --- | --- | --- |
| F1 | minor | `schemas/v1/context-lock-v1.schema.json`, `tools/validate.py` wire semantics | a member listing itself in `required_by` passes schema and the suite's semantic check; §1.4 says a cycle fails and names it | add `name not in required_by` to `validate_wire_semantics` (schema cannot express it); one negative case |
| F2 | minor | `schemas/v1/agent-environment-marker-v1.schema.json` | `copies[].path` is not required to be one of the entry's `paths`; `form` is admitted on the `mcp`/`skills`/`system-prompt` surfaces though §8.2 says "its form where the surface has one"; an unknown surface key (`zzz`) is admitted (text closes only `mcp` explicitly) | per-key `properties` for the four surfaces with `form` only under `root-context`; copies ⊆ paths in wire semantics |
| F3 | low | `schemas/v1/launch-env-fragment-v1.schema.json` | `path_prepend` and `env` values admit `..` segments and the environments root itself; §10.2 "MUST reject any value outside the environments root" is enforced only by the suite's fixed `/manager/environments/` bound (producer reading 8) | reject `..` segments by pattern; containment stays a reader rule |

Existing case dirs already cover every closed object's unknown-member rejection and the `argument`/`name` pairing (`invalid-flag-missing-argument`, `invalid-flag-name-argument-without-name`, `invalid-flag-name-without-name-argument`).

## Dimension 2 — version/range vectors vs node-semver 7.7.4

`node v25.6.1`, `semver@7.7.4` (`/opt/homebrew/lib/node_modules/npm/node_modules/semver`), script `.temp/review/semver-check.js`, log `.temp/review/node-semver-review.log`. Every valid `range_cases` comparator set equals `new semver.Range(r).set` (spelling normalised: exact `1.2.3` ↔ `=1.2.3`, any `""` ↔ `*`, `^0` ↔ `<1.0.0-0`); all 63 `satisfies_cases` agree; `latest` is rejected by node (Curator spelling, expected).

| Range | Vector | node 7.7.4 | Text §1.4 |
| --- | --- | --- | --- |
| `^0.2.3` | `>=0.2.3 <0.3.0-0` | same | same |
| `^0.0.3` | `>=0.0.3 <0.0.4-0` | same | same |
| `~1` | `>=1.0.0 <2.0.0-0` | same | same |
| `1.x` | `>=1.0.0 <2.0.0-0` | same | same |
| `>=2.1` | `>=2.1.0` | same | same |
| `<3` | `<3.0.0-0` | same | same; `3.0.0-rc.1` ∉ `<3` (node false) |
| `^2.0.0-rc.0` ∋ `2.0.0-rc.1` | true | true | true |
| `>=2.0.0-rc.0` ∋ `2.1.0-rc.1` | false | false | false |
| `^1 \|\| ^3` over {1.5.0, 3.2.0, 3.9.0-beta} | 3.2.0 | `maxSatisfying` 3.2.0 | highest member |
| `latest` | `*` | n/a (rejected) | equivalent to `*` |
| `1.2.3 - 2.3.4` | `profile_source_invalid` | accepted | excluded |
| `>=v1.0.0`, `^v1`, `v1.2.3` | `profile_source_invalid` | coerced | excluded |

Deviations from the text: none. F4 (low, text follow-up): the vector excludes `latest || ^1`; §1.4 says `latest` is "equivalent to `*`" without saying it is only admissible as the whole range. Reading is defensible; an erratum should state it.

## Dimension 3 — resolution replayed by hand (Decision 0012 D2 / §1.4 steps)

`range-conflict-empty-intersection`: seed root `*` → root@1.0.0 (only candidate); expand: core `^3`, lib `^1`; pending {core, lib}; core (smallest): `^3` over {2.5.0, 3.1.0} → 3.1.0; lib: `^1` → 1.0.0, expands core `^2` → core pending; core effective `^3 ∩ ^2` = ∅ → `context_range_conflict`, requirers `root@1.0.0 range ^3`, `lib@1.0.0 range ^2`, candidates [2.5.0, 3.1.0]. Matches the vector.

`downward-reselection`: root@1.0.0 → lib `*`, plugin `^1`; lib first: `*` → 2.0.0; plugin: 1.0.0, adds lib `<2`; lib re-selected to the highest candidate ≤ 2.0.0 satisfying `* ∩ <2` → 1.5.0; lock lib 1.5.0 `required_by [plugin, root]`, members sorted (kind, name) lib, plugin, root. Matches.

`selection-never-increases`: app `^1` → 1.1.0 (adds helper `^1`); lib → 2.0.0; helper 1.0.0 adds lib `<2`, app `<1.1`; app (smallest pending) re-selects to 1.0.0, drops helper `^1` attributed to app@1.1.0; helper has no requirer and leaves with lib `<2`; lib re-selected under `*` not above 2.0.0 → 2.0.0; app stays 1.0.0. Matches.

Diagnostics present as cases: `context_range_conflict` (3 cases), `context_version_mismatch`, `context_weight_conflict` (+ root-map-wins warning), `context_weights_not_root`, `weights-duplicate`, `weight-unknown`, overlay conflict/duplicate. Effective weights and pins follow §6 in the worked example (identical to Decision 0012 §9 listing). Lock hashes recompute (CCJ-1 + SHA-256) for every resolution and materialization case.

## Dimension 4 — materialization bytes

`.temp/review/recompute.py` rebuilt from the vector inputs, independently of both generator and validator: all 4 header cases byte-equal (type line, `root:`, `member:` with `weight <n>` and ` overlay`, `precedence:`, `lock:`, `generated:`, `notice:`); every monolithic document (header + `---`/`## Context: <name> <version>` chapters + module bytes, parts joined by one LF) byte-equal to the expected file (the one replay miss was my script ignoring the `environments` selector in `monolithic-codex-selector-excluded`, confirmed by inspection); every file length and sha256; every core §8 surface hash incl. the MCP file; every lock hash. No-chapter set: `emptyoverlay` appears as a `member: … weight 1000 overlay` line and contributes no chapter. Weight sets: `[core, org, figma, ios, umbrella, personal]` under higher/last and lower/first, fully reversed under the other two, `figma, ios` tie never inverted. MCP bytes: `claude_code` and `opencode` CCJ-1 + one LF; `codex_cli` TOML `[mcp_servers.<name>]` sorted, `command`, `args = ["-y", "figma-developer-mcp", "--stdio"]` / `args = []`, one trailing LF, no blank separator (the literal reading of "no other bytes"); `pi` writes nothing. Note (not a finding): `mcp-pi-none` records `env_names` for pi although pi has no channel — informational field only.

## Dimension 5 — detectors

Case names equal §9.1 exactly: `secret-aws-access-key`, `secret-private-key-block`, `secret-bearer-token`, `secret-in-mcp-args`, `secret-in-mcp-url`, `placeholder-example-key`, `content-hash-not-secret`, `waived-span-clears-only-itself`, `pin-does-not-clear-finding`, plus `waiver-at-other-pin-does-not-apply` (`context_secret_waiver_unmatched`), `file-outside-scope-ignored`, `system-module-present` (`context-system-module-present` warning, installs). Findings are blocking with file and byte span; a pinned hash does not clear.

## Dimension 6 — generator/validator attacked

Seven mutants on scratch copies (`.temp/review/mut/*`), each with the manifest and rc.9 pin re-cut so the semantic recomputation had to catch it: byte flip in an expected CLAUDE.md → "expected bytes for CLAUDE.md differ"; blank line in the codex TOML → "expected bytes … differ"; forged `lock_sha256` → "lock_sha256 is stale"; unsorted lock members → "expected lock is stale"; flipped satisfies flag → "satisfies case … is stale"; emptied detector findings → "expected findings are stale"; swapped emitted order → "emitted_order is not the section 5 order". All seven exit 1. Without the re-cut the manifest digest catches them first.

## Dimension 7 — hosted lane: no file the pinned Go manager's golden test reads changed (table above).

## Dimension 8 — producer's text readings

1 (codex TOML, no blank line): literal byte rule is unambiguous; keep, no erratum needed unless the author intended a separator. 2 (opencode MCP descriptor without `argument`): correct per §7.3/§7.8. 3 (single root gets a chapter): the text is clear; correct. 4 (`exact-constraints-disagree` → `context_range_conflict`): acceptable; erratum candidate to name it. 5 (no cycle diagnostic code): agree, follow-up. 6/7 (`copies[].reason` enum, passthrough strategy enum, pattern class names, `context_secret_waiver_applied`): schema choices the text leaves open — follow-up erratum to fix the spellings in `environments.md` so the schema is not the only authority. 8 (`path_prepend` bound): see F3. 9 (`^0`): semantics identical.

## Stray

`LOGBOOK.md` added by the commit does not belong in curator-spec; the orchestrator removes it on the draft branch before landing. Noted, not blocking. The reviewer did not write a repository logbook entry for that reason.

## Follow-ups (not blocking): F1–F4 above; erratum items 4, 5, 6/7 from the producer's readings.

repeat-of: none
