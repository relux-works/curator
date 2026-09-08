# Protocol errata candidate — recovery handoff

Task: TASK-260908-2zn8fu. Run: RUN-260908-62fee1.
Base HEAD: 87a0d0060bad64ab883d007dcdf35df7485368bf.

The previous producer left the two documentation changes intact. No repository changes were necessary in this recovery. `cmp` of the current `git diff --binary` against TASK-260908-2zn8fu_change-request_rev1.patch exited 0: this candidate is byte-identical to revision 1. The existing TASK-260908-2zn8fu_results.md contains the accepted A0 E4/E5, reviewer F1, and TASK-260908-1c0fwn source/version map. I reread that map, the accepted A0 reviewer verdict, and every changed hunk; inherited-secret, removal, argv, Pi precedence, ownership, and residual boundaries remain as recorded there. Independent candidate review remains the parent's next step.

## Recovery cause and verification

The previous handoff's configured `make validate` failed, exit 2, because system Python lacked jsonschema. The prior producer's manual validation used an existing venv but did not give that environment to handoff. No validator or configuration change, installation, or bypass was needed: this recovery supplies the same existing venv PATH to both validation and handoff.

Directly rerun here:
- Readiness: existing venv imports jsonschema; git, rg, make, Go and task-board report versions. Log: .temp/TASK-260908-2zn8fu/recovery/readiness-01.log.
- `PATH=/Users/iv/Developer/ReluxWorks/curator-spec/.temp/STORY-260905-2z9pw4/worktree/.venv/bin:$PATH PYTHONDONTWRITEBYTECODE=1 make validate`: exit 0. Runs all three canonical Makefile commands (schema/vector validation, Python unit tests, Go tools tests). Full output is attached separately.
- `git diff --check`: exit 0.
- Candidate patch comparison with revision 1: exit 0.

Accepted prior evidence, not rerun: installed runtime probes and module source checks in the original evidence map. No hosted CI, ax calls, commits, runtime-home changes, schema/vector updates, protocol bump, or MCP edits. Only the original two documentation paths remain modified. Operator directives were polled and none were recorded. This outcome is the permitted operational record in place of forbidden LOGBOOK/control-root writes.

Ready for review; signed publication and independent review are parent-owned.
