# TASK-260924-291k0q — rework 2: keep rc.13 bytes, move the baseline statement (THE ONLY CURRENT INSTRUCTION)

Orchestrator decision: PRESERVE rc.13 identity. schemas/skillfile-sources-v1/README.md is pinned by the released suite manifest — revert it
to main's bytes (`git diff v1.0.0-rc.13 -- schemas/skillfile-sources-v1/README.md` empty). Put the rc.10-baseline statement and the one
conditional environments §9.4 exception in an UNPINNED document instead (COMPATIBILITY.md, or conformance/README.md if it is not in any
manifest — check the manifests and state which you chose). Everything else from revision 2 stays. Re-run: the gate on the candidate and on
main (exit 0), the gate tests, `python tools/validate.py` / the specification validation (exit 0). `task-board m 'set_status(TASK-260924-291k0q,
status=development)'` first; append "Revision 3 — rc.13 bytes preserved", `resource update`, `task-board handoff TASK-260924-291k0q --role
developer`. No LOGBOOK.md. A write-boundary `policy warn` block is a warning.
