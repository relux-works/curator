# CHANGES_REQUESTED — TASK-260916-11lwua revision 2

Reviewed candidate c8c697da2accf618f8bb5d694c21febf581ac6f2 against base f38ee3946110eeeee6bad22831118da117ad6292. All 10 changed worktree files were byte-compared with candidate blobs (10/10 match). No product code modified.

1. **Successful output and diagnostics leak aliases** — cmd/curator/envconfig.go:113-136 (also set lock diagnostic at :70). `env config set forms.claude referenced` exits 0; `env config unset forms.claude` exits 0 and prints `unset forms.claude`. `env config unset forms.codex` exits 1 and prints `curator: environments knob "forms.codex" is not set`. The normalized segments are used for lookup, but the original knob is printed. Render canonical knob paths in all success/refusal messages. Add production-entry output assertions; existing unset tests only check exit status.

Independent verification (zsh, set -o pipefail; observed exit codes):
- `go test ./cmd/curator ./internal/envregistry -run 'Test(EnvResolve|ProfileUse|EnvStatus|EnvConfig|RunDispatch|Normalize|UsageLists|UnknownEnvironment)' -count=1` — exit 0; curator 43.937s, registry 0.293s.
- `go build -o .temp/alias-review-rev2/curator ./cmd/curator` — exit 0.
- Adversarial commands above used that binary and CURATOR_CONFIG pointing to a disposable config under the assigned worktree. No host configuration changed.
- Output probes: canonical-output requirement fails 2/2 tested unset shapes. No mutant suite was run; independent attacks used actual production commands without code edits.

Attached evidence reused, not rerun: TASK-260916-11lwua_change-request_rev2-validation.log reports remote run 35098862453 success, `[exit 0]`, required command shards 1/1. Gate commit 58bc2fa08362995d03dd3ba0c560c4e4c5f3d209 resolves to the exact candidate tree. Linux/macOS/Windows test lanes and lint green per that log; rose-air skipped and therefore unverified. Full landing suite was not rerun.

Scope bounds: producer notes explicitly explain that status has no operand and unmanage is absent in this tree. Confirmed `env status claude` and `env status claude_code` both exit 2; `env unmanage --env claude` exits 2. Thus the literal task AC for status acceptance is not demonstrated; canonical matrix output is a narrower property. Resolve this scope discrepancy in the next handoff rather than claiming the literal AC passed. Spec manager CLI alias paragraph enumerates resolve/profile use/unmanage, while existing status grammar is matrix-only. Launcher refusal remains owned by the sibling task, and the fake dispatch provider does not establish real launch or refusal. Persisted marker/lock byte absence is not established by filename and stdout assertions alone. No wire/schema/default/spec edits were found.

Run goal queried: not goal-bound. Verdict: changes_requested; route to to-dev for implementation and another independent review. Findings recorded on the board; no LOGBOOK.md edit made because campaign rules forbid it.

Correction to initial attachment: the proposed system_prompt_files finding is WITHDRAWN. Canonical controls also exited 1 (0/2 successful); internal/config/environments.go:514 restricts this object to pi only. Do not broaden that schema or implement the withdrawn request. The earlier claim of canonical exit 0 was incorrect; this updated artifact supersedes it. Exactly one blocking finding remains: config output/diagnostic canonicalization.
