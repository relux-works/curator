# TASK-260905-26o45p review findings, cycle 1 (system-config schema 2 at f39f4a9)

Subject: `git diff fcdb9ba..f39f4a9` on `task-board/story/STORY-260905-2z9pw4`; Change Request `CR-TASK-260905-26o45p-1` revision 1 (base OID = f39f4a9, repository delta empty).

## Blocking / major

None.

## Informational (no change requested)

1. **Inherited grammars are wider than a reader might expect** — `schemas/v1/system-config-v2.schema.json` `$defs/environments/properties/mcp_package_allowlist` → `manager-config-v2 ... mcp_package_allowlist` = `nonEmptyString` items. An entry such as `github.com/x/y` (not a canonical source identity) is admitted; likewise `passable_env_names` admits `foo-bar` and `1A` because the core §2 identifier grammar admits them. Both are the manager-config-v2 grammars the brief mandates by reference and match §12.1 as encoded at `fcdb9ba`; not a defect of this leaf. Fix, if ever wanted: tighten in manager-config-v2 under its own review.
2. **Empty repository delta on the Change Request** — the producer's single commit `f39f4a9` is the CR base OID itself, so the candidate tree equals the base. The reviewable work is `fcdb9ba..f39f4a9` (38 files), which this review covers in full. Nothing further was expected from the leaf after that commit.
3. **Drafting report scope note accepted** — no `vectors/system-config*.json` family exists and a system file pins no defaults; generator-emitted schema cases are the correct "vectors through the generator" for this surface.

## Attack evidence (this review, own instances, venv `.temp/venv`, jsonschema 4.25.1)

- 37 adversarial instances against the composed schema 2 (unknown/unlockable environments keys incl. `build_ssh`; `locked` unknown, unprefixed, dotted-path, uppercase, non-string, non-array; `schema_version` 3 / "2" / missing; `environments` non-object; each grammar: type, null, extra field, duplicate, empty, `isolated`, `SHARED`, flat map, bad identifiers): 34 rejected, 3 admitted (finding 1, legal). 6 expected-valid edge instances (null allowlists, empty maps, empty `locked`) all valid. Log: `.temp/review-sysconf-1/attack-01.log`.
- All 24 committed cases re-checked against the schema: flags agree.
- Validator gate mutants via `validate_system_config_v2_schema` (registered in `main()` list, confirmed): §12.2 drops `isolation` → killed; schema adds seventh key → killed; isolation widened to `isolated` → killed; `locked` enum + `environments.current_profile` → killed; environments object opened → killed; §12.2 names a key outside §12.1 → killed. Log: `.temp/review-sysconf-1/mutants-01.log`.
- Hosted lane: `grep -rn "system-config" ~/Developer/ReluxWorks/curator/internal --include='*_test.go'` → three comment lines only; no Go test reads a system-config case or vector. `git diff fcdb9ba..f39f4a9 --stat` over `system-config-v1.schema.json`, `schema-cases/system-config-v1`, `conformance/v1/vectors` is empty.
- `make validate` exit 0 (validate.py, 227 unittest, `go test ./tools/...`); `make regenerate-check` exit 0. Logs `.temp/review-sysconf-1/validate-02.log`, `regen-01.log`.
- Commit `f39f4a9` signed (Good "git" ECDSA signature); tree clean, no stray files.
- Text: manager §1 names the six `environments.<key>` locks, shared-only `isolation`, whole-knob replacement of the machine file's schema-2 knob, unlocked system knob as default ahead of §12.1, schema-1 reader rejection; consistent with environments §12.2 at `fcdb9ba`. COMPATIBILITY, CHANGELOG `## Unreleased`, conformance and schemas READMEs accurate; rc.9 pin advances only the manifest sha.

Verdict: ACCEPT. repeat-of: none.
