# Review verdict: launcher SPEC 0.1.2-draft (cycle 2, stray-file probe rework) — ACCEPT

Reviewer run: RUN-260901-a8b624 on TASK-260901-25zqel, CR-TASK-260901-25zqel-2 rev 2.
Subject: curator-agent-launcher `draft/spec-pi-file-channels`, head `6de42d8` (signed, verifies G, key ivan@relux.works ED25519), atop reviewed `0029b36`, base `origin/main` @ `dae0c35`. Rework delta confined to SPEC.md (52+/23-), no code changes. Curator-side CR rev 2 delta: LOGBOOK.md only; entries match the SPEC facts, including the honest residuals. Story tip `b045c6db` is a single signed commit whose sole parent equals the CR base authority `979fa36e` — provenance contract holds.

## Primary: MAJOR from cycle 1 genuinely resolved

Re-verified in SPEC text against environments.md (curator-spec `protocol/environments.md`) and Decision 0010 D6, not against the disposition's claims:

1. **Probe rekeyed on the adapter registry.** §5.1 now opens with "keyed on the environment, not the fragment": for an env-id whose §7.3 registry declares file-kind channels (revision 1 exactly `pi`: `APPEND_SYSTEM.md`/append, `SYSTEM.md`/replace — matches the §7.3 table verbatim; claude_code/codex_cli/opencode declare none, so the pi-only claim is exact), the probe runs on EVERY launch, before the §4.5 handoff or exec, whether or not the fragment carries a `system_prompt` section. The cycle-1 bypass (no-system-modules fragment + stray `SYSTEM.md` → unwarned full-replace run) is closed: present-and-readable with no corresponding fragment descriptor is an active channel and triggers the full §5.2 warning set. The "no new knowledge edge" claim is sound — the registry is closed and the launcher already owns the §4.2 mapping.
2. **False drift-and-repair claim gone.** The replacement ownership paragraph asserts only what environments.md provides: a Curator-materialized file is marker-recorded and covered by §5.5/§8.4; any other file at a registry filename is unmanaged — §10.1 repair "leav[es] … unmanaged files … untouched" and MUST NOT adopt candidate bytes, §8.4 drift covers only marker-recorded surfaces. I additionally attacked via §7.5 (shadowing paths) and §8.5 (`environment_surface_unmanaged_conflict`): neither covers a stray `SYSTEM.md` under `system_prompt_files=off` (pi's only declared shadowing path is `AGENTS.override.md`; the conflict diagnostic fires only when materialization would write the path, which `off` never does) — so "no automated contract in the cited protocol removes it" is honest and correct, not merely plausible.
3. **Residuals stated honestly in §9**: hand launches bypass the probe (§1: launcher never mandatory), and the probe is point-in-time (probe-to-exec write race undetected). Both are real, both are outside the launcher's seat, neither is overclaimed as covered.

## Secondary checks

- **Warning contract**: §5.2 triple matches Decision 0010 D6 (customized + per-channel enumeration incl. file-kind filename, replace-discards-built-in, cache/billing note); identical set with or without the opt-in flag; non-suppressible; both-kinds launch enumerates both channels with no launcher-side deduplication/arbitration.
- **Orthogonality**: opt-in selects non-`file` channels only; a matching-semantics file does not avert `sysprompt_channel_unavailable`; file presence never suppressed by, nor satisfying, the flag. Refusal path attacked: opt-in + no non-file channel + present file fails the launch rather than silently running customized — consistent.
- **Diagnostics**: §6 row reworded "registry-declared" (minor applied); absence-vs-read-failure invariant intact — absence is the channel's legitimate inactive state, `sysprompt_file_unreadable` only on present-but-unreadable/non-regular; dangling symlink routed to unreadable.
- **Minors from cycle 1**: "before the exec" tightened to "before the §4.5 handoff or exec" (§5.2); §5.1 Absent bullet grounded in the registry; §5.2 plain-launch exception now covers stray files. No stale fragment-keyed probe language remains (grepped).
- **Hygiene**: 0.1.2-draft in header, §8, and a new §8.1 revision-history row; commit `6de42d8` signature G; delta SPEC.md only; producer reports build/vet exit 0 (docs-only change, consistent with the validation log; no conformance vectors exist at draft stage).
- **No new contradictions** with §4 composition, ax handoff (§4.5 wording now referenced consistently), or §6 diagnostics ownership.

## Verdict

ACCEPT. Recording `accept_cr(TASK-260901-25zqel, revision=2)`; element parks at `to-review` for the orchestrator to checkpoint/integrate and make the `done` transition with `commit_ack=scope_committed`. Landing of the launcher branch (`draft/spec-pi-file-channels`, not pushed) remains the orchestrator's PR step.
