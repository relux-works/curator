# Rework 1 — TASK-260916-11lwua (curator CLI aliases)

Verdict rev2: CHANGES_REQUESTED, one blocking finding (resource TASK-260916-11lwua_review-verdict-rev2.md): `env config set/unset` and the set-lock diagnostic (cmd/curator/envconfig.go ~:70, :113-136) print the ORIGINAL knob path (forms.claude) in success and refusal messages. Render the canonical knob path (forms.claude_code) in every success/refusal message; add production-entry output assertions for set/unset/refusal shapes (both aliases and canonical controls).

Scope clarification to record in results.md (orchestrator ruling): `curator env status` takes no environment operand in this tree (matrix output only) and `env unmanage` does not exist, so the literal "env status accepts both spellings" item is satisfied by the canonical matrix output alone — say so explicitly under Bounds instead of claiming the literal AC; do not add an operand to status or invent unmanage. Do NOT implement the withdrawn system_prompt_files finding.

Narrow tests only (`go test ./cmd/curator -run 'EnvConfig|Normalize|RunDispatch|Alias' -count=1`, `./internal/envregistry`); evidence with exit codes; tick checklist; handoff rev3 from the Story worktree.
