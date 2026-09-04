# Launcher SPEC — pi file-kind channel descriptors (TASK-260901-25zqel)

## What changed

Repo `curator-agent-launcher`, branch `draft/spec-pi-file-channels` (from
`origin/main` @ `dae0c35`), signed commit `0029b36892ee87a26e5231c01a1bf7b3e75e5064`
("Specify pi file-kind system-prompt channel semantics, 0.1.1-draft").
Not pushed — orchestrator lands via PR.

SPEC.md `0.1.0-draft` → `0.1.1-draft`:

- **§5 intro + application list**: added the **file-class** bullet
  (`pi`: `APPEND_SYSTEM.md`/append, `SYSTEM.md`/replace). The launcher MUST NOT
  place, remove, or edit these files — Curator's materialization owns them,
  gated by the per-profile×environment `system_prompt_files` machine setting
  (environments.md §5.5, default `off`). The tool applies them unconditionally
  when present (environments.md §7.3), so the launcher applies nothing; its
  duty is detection + warning.
- **New §5.1 — detection, orthogonality, warning**: a mandatory pre-exec probe
  of `<home>/<filename>` for each `file`-kind descriptor in the fragment, with
  three distinct outcomes:
  - *absent* → channel inactive (the legitimate `off` default; no warning, no
    diagnostic);
  - *present + readable regular file* → channel **active**; the full
    customized-system-prompt warning set MUST print even without the
    `--system-prompt` opt-in (the opt-in happened at the machine-setting level);
  - *anything else* (unreadable, not a regular file) →
    `sysprompt_file_unreadable`, terminal.
  File-kind presence is orthogonal and additive to the flag-class opt-in:
  the opt-in selects among non-`file` channels only, never suppresses or
  substitutes for an active file, both may engage on one launch, the warning
  enumerates every active channel, and the launcher does not arbitrate
  tool-side precedence between the two applications.
- **§5.2 (was the selection/refusal block)**: opt-in selection now explicitly
  considers only `flag`/`config-key`/`variable` descriptors; a `file`-kind
  descriptor with matching semantics does not avert
  `sysprompt_channel_unavailable`. The no-opt-in clause now carves out the
  file-kind warning case. Warning contract restated to cover both engagement
  paths identically, enumerating kind (+filename for file-kind), profile,
  semantics; replace clause fires if any active channel replaces; the missing
  command-line flag MUST NOT suppress the file-kind warning.
- **§6**: added `sysprompt_file_unreadable` to the system-prompt family;
  extended the absence-vs-read-failure invariant to name the file probe.
- **§8.1**: new Revision history section (0.1.1-draft, 0.1.0-draft).

## Design decision worth review attention

The producer brief asked for "a code for a fragment that names a file-kind
channel whose file is absent/unreadable at launch". Bare **absence is
deliberately NOT a diagnostic**: the fragment reproduces the adapter's §7.3
descriptor list as data (pi's fragment always names both file channels), and
`system_prompt_files` defaults to `off`, so an absent file is the normal state
of every plain pi launch — making absence an error would fail every unmodified
launch. The diagnostic `sysprompt_file_unreadable` therefore covers the
*present-but-unreadable / indeterminate probe* case, and the SPEC states
explicitly that absence is the channel's legitimate inactive state. This also
keeps the §6 invariant intact (absence and a failure to read are different
facts; a fallback defined for absence never fires on a read failure).

## Validation

- `go build ./...` — exit 0
- `go vet ./...` — exit 0
- `git log --show-signature -1` — Good ED25519 signature (ivan@relux.works),
  status `G`
- Docs-only change; no test suite applies to SPEC prose (no conformance
  vectors exist yet at draft stage — none were run because none exist).
