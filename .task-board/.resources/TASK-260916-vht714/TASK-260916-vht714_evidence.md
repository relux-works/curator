# TASK-260916-vht714 evidence — file credential and permission spec-gap proposals

Shell: zsh (via bash tool). Worktree:
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-1on1d2/worktree`
(branch `task-board/story/STORY-260916-1on1d2`). No commits made; work left
uncommitted for handoff snapshot.

## Lifecycle note

`task-board m 'set_status(TASK-260916-vht714, status=development)'` exit 1:
`cannot set TASK-260916-vht714 to development — blocked by:
TASK-260916-2timlf (status: integrating): blocked by unfinished
dependencies`. Work proceeded in the worktree regardless; handoff will
record the true end state.

## Files changed (git status --short, exit 0)

- `M UNRESOLVED_QUESTIONS.md` (+8 lines, Filed proposals only)
- `?? decisions/0017-environment-credential-modes.md` (new)
- `?? decisions/0018-curator-run-permission-interface.md` (new)

No normative file touched: no `protocol/`, `profiles/`, `schemas/`,
`conformance/`, `cli/`, `docs/` edits; no CHANGELOG entry — per repo
convention proposed drafts add none (B7 `a68854d` touched only
UNRESOLVED_QUESTIONS.md + decisions/; 0013 `48babce` touched only its
decision file). No `decisions/` index file exists in the repo; the Filed
proposals list is the index. B7 block in UNRESOLVED_QUESTIONS.md kept
byte-identical; new paragraph + two bullets appended.

## Verdict §4 coverage

Credential draft (0017) — every §4 (credential draft) item present:
host observation with dates/sizes (B5 outcomes; 819 B managed
`.credentials.json` mtime 2026-09-16 04:30; Keychain item created
2026-07-08 / modified 2026-08-17, no suffixed item; 509 B native JSON
mtime 02:17; residual answered NEGATIVELY at 2.1.273); Pi root
correction (`~/.pi/agent`, `switch.go:78`); hazard 1
(`managed.go:500-502,934-942,1670-1677,1342-1345`) and hazard 2
(`managed.go:936` vs manager §12.4); no-copy/no-Keychain-export
boundary; darwin disposable-account experiment protocol (read-first,
refresh rewrite, suffixed-item conditions) gating
`environment_shared_unsupported`; marker/config fields (OQ5);
touchpoint list (environments §7.4/§7.7/§10.1, §12.1/§12.2 unchanged,
manager §12.4, frozen v1 untouched, manager-config schema 2).

Permission draft (0018) — every §4 (permission draft) item present:
`--permissions <native|yolo>` grammar + `--yolo` alias, default native,
before `--`; mapping table with installed-`--help` quotes and versions
(Claude 2.1.273, Codex 0.153.4 top-level + exec, Pi 0.84.2, embedded
line refs to `TASK-260916-2timlf_native-help.txt`); full refusal list
incl. `-c`/`--config` keys, `=`/separate forms, aliases, exec
placement, usage exit 2; tracked-mode refusal
(`permission_mode_tracked_unsupported`) + ax admission precondition
(SPEC §4.6, Decision 0013 D5/D3.6); stderr provenance line
`curator-run: permissions=… source=… mapped=…`; defaults.json closed to
`{model, effort}`, no config surface can ever set yolo; Pi
(`permission_mode_unsupported`) and opencode (`env_unsupported`)
refusals.

Both drafts: status `proposed — not adopted`; cite board task
TASK-260916-2timlf outcome resources `TASK-260916-2timlf_report.md`,
`TASK-260916-2timlf_review-verdict.md`, `TASK-260916-2timlf_native-help.txt`
by board id; state that filing authorizes no implementation.

## Validation

Venv created at worktree `.temp/venv` (ignored by `.gitignore`);
`pip install -r requirements-dev.txt` exit 0.
`make validate` (PATH prefixed with `.temp/venv/bin`) exit 0:
`tools/validate.py` → "validated 60 schemas and 1047 vector files";
`unittest discover` → 227 tests OK; `go test ./tools/...` → ok.

## Out of developer scope (orchestrator-owned per campaign rules)

Signed branch → PR → hosted checks → fast-forward, and the two GitHub
issues referencing the drafts, are parent (orchestrator) steps, not
developer steps. No PR was opened and no issues were created from this
run; the AC items "PR landed; two GitHub issues opened" therefore await
integration, with this Change Request as their input.
