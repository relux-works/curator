# Validation environment repair handoff

Task: TASK-260908-2zn8fu. Run: RUN-260908-db4532. Date: 2026-09-08.

The existing two-file documentation candidate was preserved without edits. HEAD remains 87a0d0060bad64ab883d007dcdf35df7485368bf; no commit or branch operation was performed.

Candidate Git blob identities:
- decisions/0013-execution-ownership-and-launch-plans.md: 42bfcd569a149f6f92bf0c4faa48acf968c72524
- protocol/environments.md: e2973a6867344ccbcfb864af830f5717b9c04d9c

Confirmed the supported configuration at curator-agent-launcher/.temp/launcher-migration/curator-spec.config.json, spawn.worktree_isolation.validation.commands, contains:

```sh
env PATH=/Users/iv/Developer/ReluxWorks/curator-agent-launcher/.temp/launcher-migration/spec-validation-venv/bin:$PATH make validate
```

Ran that exact command directly, with output redirected to the attached validation log: exit 0. It validated 60 schemas and 1047 vector files, passed 227 Python tests, and passed go test ./tools/... . Python 3.14.7 and declared jsonschema 4.25.1 were checked before validation. No installation or Makefile change was made. git diff --check exited 0. Before/after binary diffs compared identical with cmp, exit 0. Checklist item 4 was rechecked after the passing command. Earlier failed publication attempts remain failures; this run provides new passing local evidence under the repaired configuration. Runtime publication must independently bind its configured validation to the captured candidate.

Evidence mapping retained from the accepted brief: Decision 0013 sections 6.3/6.4 and environments 10.1 environment corrections map to A0 E4 plus accepted TASK-260908-1c0fwn sections 2–4 (agents-management v0.5.10 ChildEnv/Plan.Env contract). The complete argv suffix correction maps to reviewer F1 and A0 section 3.4. Pi changes in environments 5.5/7.3 map to A0 E5, section 4.3, installed Pi 0.84.2 dist/core/resource-loader.js source selection and discovery. Destination unset/PATH residual maps to accepted environment evidence section 4 and Decision 0013 closed grammar. No source probes were repeated in this validation-only retry; those accepted findings remain the source authority.

No protocol version, schema, vector, MCP channel, ax implementation, runtime home, LOGBOOK, control-root, or private board record edits. Prior candidate and evidence remain intact. This task-scoped resource records the repair anomaly in lieu of prohibited LOGBOOK writes. Parent owns independent review and signed delivery. Ready for review after the normal handoff succeeds.
