# Repaired publication validation recipe

Task TASK-260908-2zn8fu; run RUN-260908-722775.

Preserved the exact existing two-file documentation candidate without edits. Base HEAD: 87a0d0060bad64ab883d007dcdf35df7485368bf. Binary patch SHA256: a64beec6e97478d9ba06471a1672beff02dec475e72dd5de6edfbacb95740002. Comparison against the previous recovery/candidate-02.patch exited 0. The accepted source/version map remains TASK-260908-2zn8fu_results.md; installed A0 and environment probes were accepted, not rerun. Independent review remains parent-owned.

Confirmed TASK_BOARD_CONFIG resolves to /Users/iv/Developer/ReluxWorks/curator-agent-launcher/.temp/launcher-migration/curator-spec.config.json. Its spawn worktree validation command now explicitly selects the supplied venv:

```sh
env PATH=/Users/iv/Developer/ReluxWorks/curator-agent-launcher/.temp/launcher-migration/spec-validation-venv/bin:$PATH make validate
```

Ran that exact standalone command in the assigned worktree: exit 0. Full log attached as TASK-260908-2zn8fu_repaired-recipe-validation.log. Python version is 3.14.7 and installed jsonschema is 4.25.1 (direct probes exit 0). git diff --check exited 0. No dependency installation, source Makefile/configuration edit, hosted CI, ax invocation, commit, or protocol/schema/vector change by this worker. Prior failed publication results remain failures; this fresh pass does not rewrite them. Runtime handoff must independently bind its configured validation result to the candidate.

No directives were recorded at checkpoints. Initial optional skill-path inventory exited 2 because those directories were absent; the relevant assigned global skill was read. A projection using config_path was rejected (exit 1); corrected project_config() read succeeded (exit 0). These inspection failures did not serve as validation evidence.

This outcome is the permitted operational record instead of prohibited LOGBOOK/control-root writes. Ready for review after normal handoff; no signed landing is claimed.
