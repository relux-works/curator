# TASK-260922-1t551d brief (orchestrator, binding) — F-C1 envprofile credential-link repairs (0017)

Story STORY-260922-1cenbr (EPIC-260905), fresh workspace on current curator trunk. Normative
source: curator-spec main 05053cd7 (adopted Decision 0017 + operator corrections):
`protocol/environments.md` §7.4 (~1372 "Credential passthrough, provisioning seeds, and isolation"),
§7.7 (~1699 diagnostics), §8.4.1 (conflict row; schema-1 marker records path+strategy only),
§10.1 (~2737 `env resolve`: repair conflict rule, never-silent migration), §10.4 (~3092);
`profiles/manager.md` §12.4/§12.5; `decisions/0017-environment-credential-modes.md` (choices 1–7,
incl. C2/C3 evidence). Curator code: `internal/envprofile/managed.go` — `effectivePassthrough`
(:497), `finalizeMarker` (:927, the unconditional `_ = os.Remove(full)` at :871/:936/:975),
`checkPassthrough` (:1330), `repairUnderLock` (:1609); `internal/envprofile/status.go` (`homeState`
:282, the existing "passthrough entry … is detached" reason :337). Attached: the operator's relay
(0017-0018-operator-relay-260922.md) — its Pi observation is the reproduction case.

## Rulings
R1 Repairs are fix-first and never destroy bytes: unlink only a RECORDED symlink whose current
target is the declared native store (or the recorded target); a regular file at a link path, a
symlink to an unexpected target, or a detached link that is not ours ⇒ refuse with
`environment_credential_conflict` naming the path (no removal); shared→isolated removes the stale
recorded link (the first 0017 hazard: `effectivePassthrough` returning empty for isolated must not
leave the old link behind — the removal set must include recorded links, not only marker surfaces).
R2 Dangling / mis-targeted links are reported: `env status` and `env resolve` classify a recorded
passthrough whose target does not exist (operator case: managed link
`~/.curator/environments/default/pi/auth.json → ~/.pi/auth.json`, real credential
`~/.pi/agent/auth.json`) as DETACHED with `environment_credential_conflict`-class wording — never
silence; the Pi native root becomes `~/.pi/agent` per 0017 choice 3 (registry strategy), with the
migration itself left to F-C2 (do not move files here; report and refuse).
R3 Codex `isolated` admission: effective store = `cli_auth_credentials_store` when present;
ABSENT ⇒ platform default `file` ⇒ `isolated` admitted; `keyring`/`auto` ⇒
`environment_isolated_unsupported`; any other selector ⇒ `environment_credential_unsupported`.
Sharing is defined by the native effective storage only (no config realignment).
R4 Frozen v1 marker schema untouched (path+strategy only); the extended credential record is F-S1
— do not add fields.
R5 Evidence: production-entry rows through `env resolve`/`env status` (and the Go API) on
temporary stores: (a) shared→isolated leaves no stale link (mutant: skip recorded-link removal →
fails); (b) regular file at a link path → conflict diagnostic naming the path, bytes untouched
(mutant: restore unconditional Remove → fails); (c) dangling recorded link → status/resolve report
detached with the conflict wording (mutant: silence → fails); (d) codex isolated: absent key ⇒
admitted; `keyring`/`auto` ⇒ `environment_isolated_unsupported`; unknown selector ⇒
`environment_credential_unsupported`; (e) Pi strategy targets `~/.pi/agent` for NEW provisioning
(existing wrong-target homes are reported, migrated by F-C2). Windows: symlink rows need the
privilege the existing envprofile rows already handle — follow their skip/ledger pattern.
CHANGELOG `### Fixed`/`### Changed`; docs/troubleshooting entries for the two diagnostics.
results.md: hazard analysis (the two 0017 hazards with file:line), row table, mutant table,
bounds (what F-C2/F-C3/F-S1 own). Publish the Change Request only when the gate is green.
