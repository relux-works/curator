# TASK-260916-2bwfli results — wire resolved transport into CLI callers

Producer: developer. Shell for all commands below: bash in the Story
worktree (`task-board/story/STORY-260910-1bhj0g`). Quoted exit codes are
real process results. Headline first: wiring the production caller exposed
a latent main-branch bug — the default external acquisition passed the
repository NAME where the lane requires the URL, so the lane never fetched.
Repaired as the prerequisite of this leaf (details below); the legacy golden
pins the repaired lane.

## Scope reading (reviewer: read this)

No v2-source CLI install path exists yet (Expand/BuildExpanded have no
production caller), and the only cmd-reachable strict-lane acquisition is
the external-repository lane (`curator install` → `planExternalBuilds` →
default acquire → `buildrepo.AcquireNetwork`). Per repository-transport §3
("Such lanes can use their existing URL declaration and an admitted machine
policy") this leaf wires the resolved executor into that lane: URL
declarations only, behind the draft switch, with the policy's endpoint plan.
The logical `repository` spelling stays parser-only. The "existing
draft/opt-in switch" exists nowhere in code, so the switch is defined here
narrowly and operator-owned: env `CURATOR_DRAFT_TRANSPORT_RESOLUTION=1`,
read once at the CLI boundary into `install.ExternalDeps`; no flag, no
config knob, no package-visible surface ("do not widen").

## What was implemented

- `internal/install/drafttransport.go` (new): `acquireDraftNetwork`
  selection (switch off → `AcquireNetwork` untouched, never opens the policy
  file; switch on + absent policy → legacy; switch on + present policy →
  resolve via `config.ResolveRepositoryEndpoints`, convert
  `config.Resolution` → `buildrepo.TransportPlan` field-for-field,
  `AcquireNetworkResolved`); repo-scoped `draftTransportAuth` (named
  providers resolve to the repository's own HTTPS/SSH selections; unselected
  HTTPS means explicitly anonymous as in legacy); per-fetch SSH wrapper base
  discovery (platform ssh + fresh empty config files, removed after
  acquisition; missing ssh yields a zero base the executor refuses) and
  manager-binary admission as the wrapper copy source (unadmittable manager
  refuses SSH resolution before traffic, never copies a bypass).
- `internal/install/external.go`: `ExternalSource` carries the declared
  `GitURL` (set at both plan/stage sites); default closure fetches it and
  routes through `acquireDraftNetwork`; `ExternalDeps` gains
  `DraftTransportResolution` + `DraftPolicyPath`. THE REPAIR: the closure
  passed `selected.Declared.Repository` (the configured NAME, e.g. "tools")
  as the fetch URL, so admission always refused with
  `build_repository_identity_invalid: network source is not canonical parsed
  input` and zero fetches. Proven empirically pre-change; no test covered
  the default acquire with real inputs.
- `cmd/curator/main.go`: one-line SSH wrapper dispatch next to the HTTPS
  broker dispatch (`IsSSHWrapperInvocation` → `RunSSHWrapper`); 
...[truncated 6739 chars]