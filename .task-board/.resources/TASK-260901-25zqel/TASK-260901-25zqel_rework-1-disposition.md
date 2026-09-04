# Rework 1 disposition — TASK-260901-25zqel (launcher SPEC 0.1.2-draft)

Producer run responding to review verdict `TASK-260901-25zqel_review-findings-launcher-pi-1.md` (RUN-260901-7354f1, CHANGES REQUESTED on CR-TASK-260901-25zqel-1 rev 1).

## Subject

`~/Developer/ReluxWorks/curator-agent-launcher`, branch `draft/spec-pi-file-channels`, new head `6de42d8` (signed, verifies G) atop reviewed `0029b36`, base `origin/main` @ `dae0c35`. Delta confined to SPEC.md (52+/23-). No code changes.

## Finding 1 (MAJOR) — FIXED, per the reviewer's preferred option

The 0.1.1-draft §5.1 probe was keyed on the fragment's `system_prompt.channels` list, so a fragment with no `system_prompt` section probed nothing while pi still applied a stray `SYSTEM.md` full-replace unwarned; the closing paragraph asserted Curator drift-and-repair covers the stray file, which environments.md does not provide (§10.1 repair leaves unmanaged files untouched; §8.4 drift covers only marker-recorded surfaces).

Rework applied:

- **Probe rekeyed on the environment adapter registry.** §5.1 now keys the probe on the env-id whose §7.3 registry declares `file`-kind channels (revision 1: exactly `pi` — `APPEND_SYSTEM.md`, `SYSTEM.md`), probing the closed registry filename set in the managed home on **every** launch, fragment `system_prompt` section or not. Stated: the registry set is closed and versioned with the environments protocol, and the launcher already owns the §4.2 env mapping — no new knowledge edge.
- **Present-and-readable with no corresponding fragment descriptor is an active channel** and triggers the full §5.2 warning set; the warning rationale covers both the machine-setting opt-in shape and the nobody-opted-in stray shape.
- **False drift-and-repair claim removed**, replaced by an honest ownership paragraph: the launcher never removes or edits a probed file; a file Curator materialized under `system_prompt_files` is a marker-recorded managed surface owned by Curator's materialization/drift/repair (§5.5, §8.4); any other file at a registry filename is unmanaged — Curator's repair explicitly leaves it untouched and drift does not flag it (§10.1, §8.4), so no automated contract removes it; removal is the operator's deliberate action, and every launcher-mediated launch warns until then. No Curator behavior is asserted that environments.md does not provide.
- **Residuals stated honestly in §9**: hand launches bypass the probe entirely (the launcher is never mandatory, §1), and the pre-exec probe is point-in-time, so a file written between the probe and the tool's own startup read goes undetected. Neither is closable from the launcher's seat.

## Minors / non-blocking notes from the report

- Note on CR rev 1 bundling the bootstrap logbook entry: reviewer marked no action needed; none taken (the story-branch provenance contract keeps one commit atop the authority, so the amended logbook commit still carries both entries).
- Note on §5.2 "before the exec" wording (carried from 0.1.0-draft, marked fine to leave): tightened anyway to "before the §4.5 handoff or exec" for consistency with §5.1.
- Consequential consistency edits: §5.1 Absent bullet now grounds absence-is-normal in the registry rather than the fragment; §5.2 plain-launch exception now says "where a file-kind channel's file is present in the home" (covers stray files); §6 diagnostic row wording changed from "fragment-named" to "registry-declared" for `sysprompt_file_unreadable`.

## Hygiene

- Version bumped `0.1.1-draft` → `0.1.2-draft` in the header, §8, and a new §8.1 revision-history row describing the rekeying, the removed false claim, and the §9 residuals.
- Signed commit `6de42d8ee96c9f69886057fac715f1287c58e45e`, `git log --show-signature` reports Good signature (G), author key ivan@relux.works ED25519.
- Story-branch logbook: rework entry added; tip amended to `b045c6db` (single signed commit, parent exactly the selected authority `979fa36e`, verifies G) to preserve the CR provenance contract that failed once before with `change_request_base_authority_mismatch`.

## Gates (real exit codes)

| Gate | Command | Exit |
|---|---|---:|
| build | `go build ./...` (curator-agent-launcher) | 0 |
| vet | `go vet ./...` (curator-agent-launcher) | 0 |

Docs-only change; no conformance vectors exist at draft stage, so none run (same expected-absence as rev 1). Not pushed — orchestrator lands via PR.
