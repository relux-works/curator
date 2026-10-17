# L15-1 — Resolve the machine-owned credential-store policy once

- Owner repository: **curator**.
- Normative clauses: login-readiness §§2, 2.1, 5; environments §§7.4, 8.2,
  10.1–10.3, 12.1–12.2; profile-codex-settings §§1, 3–6; D9, D10, D14.
- Attack surface: source/store/isolation confusion, native credential
  resurrection, and implicit migration while resolving or changing policy.
- Size: 5. Baseline: curator `0d00a109`, spec `5509499`.

## CLI/config and acceptance

```text
curator env config set credential_store.<profile>.<env-id> native|file|keychain
curator env config show credential_store.<profile>.<env-id>
curator env config unset credential_store.<profile>.<env-id>
```

Use the allocated successor manager-config version. The persisted member is
`environments.credential_store`, a profile map of canonical environment IDs.
Absent means native; null, unknown IDs/values and malformed maps are invalid.
Reuse established CLI aliases (`claude`/`codex`) in environment positions;
persist canonical IDs and preserve existing alias/canonical collision rules.
Packages, overlays, repositories and system locks never select this policy.
An existing isolation lock continues to bind isolation.

One resolved typed policy object feeds home verification, readiness, login and
the fragment producer. Order: **source → store → isolation admission → links**.
Do not have the login command or consumer rediscover policy independently.

| Adapter / selection | Required result |
|---|---|
| Codex native | Inspect native config with absence/failure distinction; absent key means file; no launch override |
| Codex explicit file | Typed `file` override; isolated file allowed even with verified native keyring/auto; shared requires compatible native file |
| Codex explicit keychain | Typed `keyring` only on qualified macOS mapping; isolated always refuses pending scope qualification |
| Managed Muse absent/native/file | Effective file; four managed XDG paths plus literal `TBH_CREDENTIAL_BACKEND=file`; report mapping, preserve stored knob |
| Managed Muse shared | Native file login under captured pre-override native XDG config; missing file gives guidance and no shared fragment |
| Claude native | Session-qualified native choice; macOS shared still refused |
| Claude forced file/keychain; Muse forced keychain; Pi/OpenCode non-native | Refuse with the reconciled diagnostic; never silent native fallback |
| Adopted node/helper/broker source with non-native home-store override | Refuse as inapplicable, without source replacement or local login |

Changing the knob only changes machine policy. It never migrates/deletes a
credential or rewrites vendor `config.toml`. Recorded backend/link conflicts
require existing inspect → plan → apply migration; `--repair` cannot erase a
detached credential, synthesize a placeholder, export Keychain material or
resurrect an old file. Native malformed/unknown Codex selectors refuse even
when an explicit isolated/file override is requested.

## Negative tests and positive controls

| Negative / forbidden effects | Positive control |
|---|---|
| Old config envelope, system lock, package/overlay value or malformed selector accepted | Valid machine successor entry; unset restores effective default without touching homes |
| Codex shared explicit file uses a stale `auth.json` while native config selects keyring/auto | Shared native file uses only the verified expected link |
| Codex native auto/keyring admitted as isolated or explicit keyring treated as per-home | Isolated explicit file with a valid native selector, no native link |
| Native config permission/parse failure becomes absent key/file | Readable valid config without key defaults to file |
| Muse native/unset omits file literal or uses default global Keychain readiness | Effective file at both login and launch; shared uses compatible native file |
| Selector switch followed by repair removes a vendor-owned file or silently changes ownership | Explicit migration with non-conflicting fixture follows existing transaction contract |
| Alias persists as an environment ID or normalization changes a profile name | Canonical and accepted CLI alias forms resolve to the same canonical target |
| Foreign source admitted beside forced home store | Explicit adopted external source with native home-store selector retains its source contract |

Drive `env config`, `env resolve` and the actual fragment producer. Preserve
before/after config, marker, link and synthetic credential state. Add a bounded
control that deliberately drops only the hosted override and require the real
composition test to detect it; schema agreement alone is insufficient.

## Dependencies, paths, release cut

L15-0 before config emission; the typed policy feeds the v4/Muse and successor
Codex producer/consumer leaves. Codex explicit-store enablement also needs
supported-release [login/app-server/refresh qualification](codex-storage-qualification.md) and strict consumers
(profile-codex-settings §5); accepting a schema does not satisfy it.

Likely paths: `internal/config/{config,environments}.go`,
`internal/envregistry/envregistry.go`, `internal/envprofile/managed.go`,
`cmd/curator/{env,envconfig,envalias}.go`, corresponding focused tests,
`docs/environment-config.md`. Serialize shared `managed.go`/registry/config
edits with Muse producer, Codex identity and materialize work. The new policy
module can be implemented independently, then integrated once by its owner.
Required in rc.7; unsupported mappings remain explicit refusals.
