# Review note — TASK-260922-2u5jzw F-L1b, CR revision 1 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review against `2u5jzw-brief.md` + `2u5jzw-resume-3.md`, the task AC/checklist and the results. Read-only; disposable clone.
Source of truth: launcher SPEC 0.5.0-draft (on this Story branch, checkpointed F-L1a) §3, §4.1-§4.7, §6; curator-spec decision 0018
(choices 1-7) and environments §10.1/§10.2; skill-agents-management v0.5.22 public API (audit in TASK-260924-1nh93t_results.md).
1. Flags: `--permissions native|yolo`, `--yolo` alias (= and separate forms), conflicts refused, `-d/--danger` rejected.
2. Precedence flag > profile > global > default-interactive (yolo) > default-headless (native) with provenance `source`; headless
   detector = non-TTY ∪ {CI, GITHUB_ACTIONS} ∪ tracked ∪ `ClassifyNonInteractiveArgs`; lock above precedence (visible yolo → usage).
3. Composition ONLY through `vendorplugin.SpawnRequest.PermissionMode/ToolRelease/NativeArgs` (vendor admission); the §4.3 pre-admission
   line with `mapped=` from `Registry.PermissionMapping`; NO provider flag spelled anywhere in launcher sources (module test + your grep).
4. Refusals: tracked + yolo from every level (`permission_mode_tracked_unsupported`, no untracked fallback); transport not established
   (`permission_policy_unsupported`, fragment token `launch-env-fragment-v2`); module typed errors (conflict, unverified release,
   classifier indeterminate) mapped to SPEC §6 exits via errors.Is/As.
5. Choice 4: under `native`, `InspectStoredPolicy` → effective-native-policy stderr line (long allow-lists summarised); tracked → the
   launcher-SPEC key in the ax launch document; untracked → stderr only; never claims a not-inspected source.
6. Choice-5 row families driven through the REAL `curator run` entry with a fake tool (counts per family in results); narrowing mutants
   per refusal bound killed — re-apply two yourself. `make check` exit 0; hosted gate green (validation log).
7. SPEC follow-ups recorded (grammar v1→v2 wording in §4.4; stored-policy sources not enumerated) — do not require SPEC edits here.
Findings → changes requested with file:line; else accept_cr. No LOGBOOK.md.
