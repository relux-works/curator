# Review findings cycle 3: environments 1.1 batch 2 — rebased landing head fd237ba / CR rev 3 (9454cd3)

## Verdict: ACCEPT (no blocking or major findings; cycle-1 minors F1–F4 stand as non-blocking follow-ups)

repeat-of: none

## Landing-head confirmation (draft worktree `curator-spec-env-schemas`, read-only)

| Check | Result |
| --- | --- |
| `git log --oneline origin/main..HEAD` | exactly one commit: `fd237ba` on top of `f61ee9a` |
| `git log --show-signature -1 fd237ba` | `Good "git" signature with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM` (principal lookup fails only because the machine's allowed-signers path is stale, as in cycles 1–2) |
| `git range-diff origin/main 794c7bd HEAD` | non-`=` hunks only in: `CHANGELOG.md` (2), `conformance/README.md` (2), `conformance/v1/manifest.json` (3), `schema-cases/index.json` (10), `release/1.0.0-rc.9.json` (2), `context_detectors.go` (1, gofmt map alignment), `context_versions.go` (2, gofmt struct/map alignment), `environments.go` (1, gofmt closure reflow), `tools/validate.py` (1, pure context: batch 3's `validate_manager_config_vectors` neighbour line). Nothing else. |
| `CHANGELOG.md` `## Unreleased` | Added: manager-config-v2 (batch 3), this batch's section-13 bullet, snapshot byte-exactness rule + vector; Changed: manager §12 rewrite, cli rows (batch 3); Removed: profilefile/context-manifest withdrawal. `uniq -d` over bullets: none. |
| `conformance/README.md` | context-versions, context-detectors, manager-config-v2, snapshot-acquisition bullets each once; `uniq -d`: none |
| `gofmt -l tools/` | empty |
| `make validate` | green: `validate.py` ok, 204 Python tests OK (28.9s), `go test ./tools/...` ok |
| `make regenerate-check` | green: `git diff --exit-code` over `conformance/v1` and the rc.5–rc.9 pins produced no output |
| Pinned-lane files `git diff origin/main --stat` (manager-config.json, manager-config-v2.json, canonical-*.json, source-identities, identifiers, locale-selectors, snapshot_sha256.txt) | empty |
| `git ls-files \| grep LOGBOOK` | empty |
| PR #42 | head `fd237ba`, OPEN, MERGEABLE; checks: Formatting, Links, Specification ×3, Implementations ×3 all pass; Release target provenance skipped (expected, not a release) |
| CR rev 3 candidate | story worktree HEAD `9454cd3`, tree `08f50f351ae7579e95ec13a2d96d926ed6e7847a` = the CR's candidate tree OID; `git diff fd237ba 9454cd3 --stat` empty; `git status --short` clean |

## Spot-checks on this head (all hold byte for byte)

### node-semver 7.7.4 (`/opt/homebrew/lib/node_modules/npm/node_modules/semver`, node v25.6.1)

| Range | Vector `comparator_sets` | node `Range.set` | Text §1.4 |
| --- | --- | --- | --- |
| `^0.0.3` | `[[">=0.0.3","<0.0.4-0"]]` | same | same |
| `^0.2.3` | `[[">=0.2.3","<0.3.0-0"]]` | same | same |
| `<3` | `[["<3.0.0-0"]]` | same; `3.0.0-rc.1` ∉ (vector `satisfies:false`, node false) | same |
| `~1` | `[[">=1.0.0","<2.0.0-0"]]` | same | same |
| `1.x` | `[[">=1.0.0","<2.0.0-0"]]` | same | same |
| `>=2.1` | `[[">=2.1.0"]]` | same | same |
| `^1 \|\| ^3` over {1.5.0, 3.2.0, 3.9.0-beta} | satisfies 3.1.0/1.9.0 true, 2.0.0 false | `maxSatisfying` 3.2.0 | highest member |
| `latest` | `[["*"]]` | n/a | `*` |
| `1.2.3 - 2.3.4`, `^v1` | `profile_source_invalid` | node accepts | excluded by text — vector follows text |

### Hand-recomputed v2 header: `weights-winner-higher-placement-first/CLAUDE.md`

Rebuilt from the vector's `lock`, `precedence`, `emitted_order` with an independent script (`.temp/review3`, not the generator or validator):
- lock hash: SHA-256 over CCJ-1 bytes (sorted keys, no whitespace, no trailing LF) of `lock` = `sha256:b8448aa1…de06` = the vector's `lock_sha256` and the `lock:` line (text §1.3 line 177: CCJ-1 bytes hashed).
- All 11 leading lines (`<!--`, type line, `root: umbrella 2.3.0 commit 6666…`, six `member:` lines with `commit`/`state sha256:` pins, `weight <n>`, ` overlay` on `personal`, `precedence: winner=higher-weight placement=winner-first`, `lock:`) equal; `generated:`/`notice:`/`-->` equal to §5 grammar (lines 605–613).
- Emitted order `personal, umbrella, figma, ios, org, core` = descending weight with the `figma, ios` tie kept in topological order, as `emitted_order_rule` and §6 require; chapters `## Context: <name> <version>` follow in that order.
- File: 1334 bytes, sha256 `f5f3…342b` — both match the vector's `files[0]`.

### Extra schema probes (jsonschema with a local registry of `schemas/v1/*.schema.json`)
- Unknown member `zzz` on: lock root object, lock member, `agent-mcp` `server`, fragment `profile`, marker `precedence` → all REJECTED.
- `agent-context` `weights` value 2147483648 → REJECTED (max 2147483647); 2147483647 → admitted (boundary correct).

## Not findings / notes
- The draft worktree carries an untracked `tools/__pycache__/` left by the gate runs; not in the commit, not in `.gitignore`. Housekeeping for the orchestrator, not a review finding.
- Cycle-1 minor F1–F4 (lock self-`required_by`, marker `copies ⊆ paths`/`form` per surface/unknown surface key, fragment `..` segments, marker `mcp-pi-none` `env_names` note) remain unchanged and non-blocking; they belong to a follow-up item, not to this revision.
