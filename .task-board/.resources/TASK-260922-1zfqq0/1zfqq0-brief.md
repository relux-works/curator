# TASK-260922-1zfqq0 — F-L1a: launcher SPEC revision for the 0018 permission interface

Control root: /Users/administrator/Developer/ReluxWorks/curator/curator-agent-launcher; work only in
your assigned Story worktree `.temp/STORY-260922-39hxog/worktree`. Read `campaign-producer-rules.md`
first. Landing gate = hosted CI (runtime runs it once at handoff). The SPEC is at `0.4.1-draft`
(SPEC.md line 3, `specVersion` in cmd/curator-run/main.go, README) — this leaf produces
`0.5.0-draft` (a new interface, not an erratum).

## Source of truth (read before writing)
- `curator-spec/decisions/0018-curator-run-permission-interface.md`: the Compatibility section lists
  exactly which launcher sections change, and choices 1, 4, 5, 7.
- `curator-spec/protocol/environments.md` §10.1 "Permission mode" (the normative resolution,
  headless detector, lock precedence and transport rule) and §10.2 (the fragment carries the
  profile level and the lock engagement).
- F-S2 (TASK-260922-1hla8q, landed before you on curator-spec main): the EXACT fragment member names
  and the minimum transport version token. Cite them; do not invent names.
- **Decision 0013 D5: the launcher never spells a provider flag.** F-M1 (skill-agents-management)
  owns the spelling, the mapping, the argv grammar and the capability table. This SPEC must contain
  no provider flag string at all.

## Deliverable (SPEC.md + README.md + CHANGELOG only — no Go code, that is F-L1b/TASK-2u5jzw)
1. §3 flag table: `--permissions native|yolo`; `--yolo` as an exact alias of the yolo mode (same
   increment, choice 1); `-d`/`--danger` rejected.
2. §4.1: the fragment members from F-S2 and the transport version precondition.
3. §4.3: defaults v2 and the item-2 precedence — flag > per-profile knob > launcher-global default >
   built-in default (`yolo` for interactive untracked, `native` for headless/CI/tracked).
4. §4.5: composition placement — the resolved mode is passed to the agents-management
   `LaunchRequest` permission-mode member, never spelled.
5. §4.6: tracked refusal from every level (`permission_mode_tracked_unsupported`, no fallback to
   untracked); the item-5 headless detector with the closed marker set {`CI`, `GITHUB_ACTIONS`}
   (mirror of environments §10.1, stated once as closed and versioned); headless/CI/tracked silence
   ⇒ `native` with `source=default-headless`; the §12.2 lock above the whole precedence (visible
   `yolo` ⇒ `usage`, silence ⇒ `native`).
6. §4.7: the file-family note for the new knob.
7. §6: diagnostics — `permission_policy_unsupported` (transport not established),
   `permission_mode_tracked_unsupported`, and the `usage` rows with their exit codes.
8. Choice 4: the effective-native-policy stderr line (exact spelling) printed when native stored
   settings relax the posture beneath a `native` request, and the launch-record extension key that
   records it (tracked = the ax launch document's launcher-SPEC-owned key; untracked = stderr
   provenance only, no persistent record).
9. README options table; CHANGELOG (unreleased) + the SPEC version-history row for `0.5.0-draft`;
   bump `specVersion` and the help golden so `TestSpecVersionPinned` (which now reads both
   documents) stays green.

## Acceptance evidence to produce
A doc-level grep row proving no provider bypass flag string appears anywhere in the launcher docs,
the diagnostics table complete with exit codes, and `make check` exit 0. Attach
`TASK-260922-1zfqq0_results.md` (section-by-section before/after, the F-S2 names you cited, the grep
row, gate exit code) and hand off with `task-board handoff TASK-260922-1zfqq0 --role developer`.

## Addendum 2026-09-23 (orchestrator) — dependencies are now landed
- F-S2 is on curator-spec main (`ec8dc656`): token = fragment revision `launch-env-fragment-v2`;
  required closed member `permissions { mode: native|yolo, locked: bool, source: profile|global|default }`;
  lattice: `locked` iff `source == global`; `mode` is `native` whenever `source` is `global` or `default`
  (`source:default` = the profile level is SILENT, fall through to the launcher-global default).
  Headless markers {CI, GITHUB_ACTIONS} are stated in environments §10.1 — mirror them in §4.6.
- F-M1 is released as skill-agents-management **v0.5.18** (`149569d`): `LaunchRequest.PermissionMode`
  (native|yolo), `LaunchRequest.ToolRelease`, `LaunchRequest.NativeArgs`, grammar token
  `permission-grammar-v1`, drift refusal `ErrPermissionModeUnverifiedRelease`. The SPEC cites the member
  and the token by name; it still spells NO provider flag.
- Every spawn runs with `--context-profile full` (skill-project-management#359).
