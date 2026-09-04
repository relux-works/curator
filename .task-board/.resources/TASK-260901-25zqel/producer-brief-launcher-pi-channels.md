# Producer brief: launcher SPEC — pi file-kind channel descriptors (carried minor)

## Setup
- Repo `~/Developer/ReluxWorks/curator-agent-launcher`, fetch origin, branch `draft/spec-pi-file-channels` from origin/main. Signed commits. Do not push (orchestrator lands via PR).

## Task
Close the carried review minor: SPEC.md §5 (system-prompt application) does not address file-kind channels, leaving the selection rule ambiguous for pi, which has BOTH flag-class (`--system-prompt`/`--append-system-prompt`) and file-kind channels (`APPEND_SYSTEM.md` append, `SYSTEM.md` full replacement — materialized by Curator into managed homes only under the per-profile×environment machine setting; applied by the tool unconditionally when present, protocol/environments.md §5.5/§7.3). Specify:
- how the launcher treats file-kind descriptors in the fragment: it does not place, remove, or edit the files (Curator's materialization owns them); on launch into a managed home where an active file-kind channel is present, the launcher MUST still detect it and print the same customized-system-prompt warning set even without the --system-prompt opt-in flag (the opt-in happened at the machine-setting level);
- selection when both kinds are available for one environment: opt-in flag engages the flag-class channel; file-kind presence is orthogonal and additive (pi applies files regardless); the warning enumerates every active channel;
- diagnostics: a code for a fragment that names a file-kind channel whose file is absent/unreadable at launch.
Bump SPEC version 0.1.0-draft -> 0.1.1-draft with a short changelog line in the doc header if the SPEC has one (add a minimal Revision history section if not).

## Deliverables
Signed commit; board resource `launcher-pi-channels-notes.md`; handoff to-review.
