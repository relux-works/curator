# Evidence: CIP-0007 manager-provisioned CLI tools (2026-10-05)

Design research for curator-spec issue #108. Triage (2026-10-04)
recommended option B, a registry-only lane, noting it needs an owner and
service decision. Output: Draft
[`cips/CIP-0007-manager-provisioned-cli-tools.md`](../cips/CIP-0007-manager-provisioned-cli-tools.md)
plus one index row in `cips/README.md`.

Scope and limits: docs-only research against specification rc.14 in this
repository. No Curator implementation revision was inspected, no behavioral
probe was run, and no normative protocol or schema text was changed. All
citations below are file:line or file § in this repository unless noted.

## Sources read

- Issue #108 (full proposal text, `gh issue view 108`), including its
  relations to #100 (provider installs), #101 (newer Go families), #106
  (manager-provisioned Go toolchain), and #107 (shim resolution).
- `cips/README.md` (process, binding section list, index), `cips/TEMPLATE.md`,
  `cips/CIP-0001-curator-improvement-proposals.md` (Accepted process), and
  Draft CIPs 0002–0006 for tone and depth (CIP-0005's options table,
  normative sketch, and numbered open questions were the closest model).
- `protocol/core.md` §4 (skill manifest, system commands), §3 (no package
  code execution), §1 (frozen schemas); `protocol/registry.md` §§1–2, 4–5, 8–9
  (signatures, rotation, federation, rollback, offline, wire contract);
  `protocol/skillfile-sources.md` §§3–4 (lock identity, markers);
  `profiles/manager.md` §§2–4, 10 (install lifecycle, scopes, status).
- `schemas/v1/common.schema.json` (`systemCommand`), `schemas/v1/skillfile-*.schema.json`,
  `schemas/skillfile-sources-v1/skillfile-{v2,lock-v1}.schema.json`.

## Key findings (with citations)

1. System commands today are a bare name plus hint, presence-checked only:
   `protocol/core.md:210-213`; closed schema object
   `schemas/v1/common.schema.json` (`systemCommand`: `type`, `command`,
   `hint`). A missing command fails installation with the hint.
2. Readiness verification happens in read-only planning
   (`profiles/manager.md:163-166`, §2.1 step 7); nothing in the lifecycle
   provisions a tool. Install is plan → private staging → one serialized
   journaled transaction (`profiles/manager.md:134-599`, §§2.1–2.5).
3. No version-constraint or recorded-tool-identity surface exists: the lock
   records skill members and package identity
   (`protocol/skillfile-sources.md:161-170`), and `manifest_sha256` covers
   the whole declaring manifest, so a new tool declaration would naturally
   stale an old lock via `source_lock_stale`.
4. A reusable signed-registry machine exists: CCJ-1 canonical bytes
   (`protocol/registry.md` §1), Ed25519 envelope (§2), overlap rotation
   (§2.1), canonical URLs (§2.2), snapshot rollback with persisted
   high-water (§5), cache/offline grace with stated revocation residual
   (§8), deny-wins federation (§4). Option B reuses all of it without new
   key semantics.
5. The Go toolchain (§2.2) sets the trust bar any provisioning must meet:
   operator-trusted family, fixed probe vectors, fingerprinted tree, and
   rejection of package-selected executables.
6. CIP-0005 (audit backends) is the interaction point for audit evidence:
   resolved tool identity belongs in audit evidence, and provisioned tool
   paths must not widen analyzer child environments beyond the
   backend allowlist (`cips/CIP-0005-audit-backends-and-cli-secret-transport.md`
   §§"Design ¶5", "Security considerations").
7. Manifest-shape caution: core §1 freezes schemas 1–6 behavior and schema 1
   keeps deployed extension behavior, which is why the CIP recommends a new
   dependency block over extending `system` (open question 5).

## Fact-checks performed

- Verified `core.md:211-212` as cited by #108 (bare name, hint, install
  failure) — confirmed at `protocol/core.md:210-213` (line shift only).
- Verified related-issue titles via `gh issue view` for #100, #101, #106,
  #107 (all OPEN; titles match the CIP's Related line).
- Verified the CIP-0001 process requirements (template section list, index
  row, Draft status) and that CIP numbers 0002–0006 are taken, so 0007 is
  the next free number.
- Verified no `cips/` checks exist in `tools/validate.py` (grep for
  `cips|CIP` empty); validation run still recorded below for the repo gate.
- Checked the CIP and this note for the forbidden classes: no organisation,
  team, client, or person names; no personal paths or host names. The
  upstream catalogue is named only as the public `aqua-registry` project;
  the field-report source in #108 is described without attribution.

## Validation

- `python -B tools/validate.py` — exit code **0**
  (`validated 73 schemas and 1294 vector files`), run 2026-10-05 via an
  isolated venv with `requirements-dev.txt` (`jsonschema==4.25.1`); bare
  `python` is absent on this host so the venv interpreter was used. No
  normative files were touched, so no digest or vector regeneration was
  needed.
