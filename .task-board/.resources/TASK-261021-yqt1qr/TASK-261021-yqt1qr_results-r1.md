# TASK-261021-yqt1qr results, revision 1

## Delivery

- Repository: curator-spec
- Branch: `cip16-delegated-launches`
- Base: `origin/main` at `e41f96387287f5e76604627f9e63662f7d6d86d4`
- Head: `3a3e041b3fbf76292e69b6abb4a0eca855524007`
- Signed commit verification: `G`.
- Board status and assignment were read, not changed. No PR was opened and no default branch was pushed.

## Summary

Recorded the delegated local launch amendment in a dated CIP-0016 revision, accepted by the owner on 2026-10-21. The revision preserves caller-side composition through the shared agents-management construction site, fragment-digest refusal, unchanged resolved inputs, no Remote Control, explicit live-home/provenance limits, typed resume, no native fallback, per-runtime activation, retirement and registry backend selection. Home leases and one profile per Werkstatt are explicitly proposed and not yet accepted. A minimal protocol pointer scopes the existing numbered contract to the exact-plan route; README and CHANGELOG now distinguish the accepted direction from implementation and runtime qualification. No wire schema, conformance corpus or runtime implementation changed.

## Files changed

- `cips/CIP-0016-local-hosted-launches.md`: dated accepted revision and separate pending conditions.
- `protocol/local-hosted-launches.md`: amendment cross-reference and exact-plan scope clarification; section numbers unchanged.
- `README.md`: update the existing accepted local hosted direction entry.
- `CHANGELOG.md`: record the amendment under Unreleased / Changed.

## Commands and exit codes

`SPEC_ROOT` and `SPEC_GIT_DIR` below stand for the local checkout and its separate Git directory. Private machine paths are omitted. Commands ran from the specification checkout. `PY` stands for `.temp/TASK-261021-yqt1qr/venv/bin/python`.

| Command | Exit | Evidence / claim |
|---|---:|---|
| `a2a inbox` | 0 | Received the addressed orchestrator request in the required room/topic. |
| `task-board q 'get(TASK-261021-yqt1qr) { status assignee }'` | 0 | Assignment matches the worker; status is development. |
| `task-board q 'get(TASK-261021-yqt1qr) { full }'` | 0 | Exactly two precondition resources are attached. |
| `task-board resource get TASK-261021-yqt1qr W-CIP16-brief.md` | 0 | Brief materialized and read. |
| `task-board resource get TASK-261021-yqt1qr front-door-appendix-A.md` | 0 | Amendment source materialized and read. |
| `a2a post ... --topic cip16-delegated --kind ack ...` | 0 | Task acknowledgement posted. |
| `git --git-dir="$SPEC_GIT_DIR" --work-tree="$SPEC_ROOT" fetch origin` | 0 | Fetched before branching. Agent Git environment was sourced first. |
| `git --git-dir="$SPEC_GIT_DIR" --work-tree="$SPEC_ROOT" switch -c cip16-delegated-launches origin/main` | 0 | New branch starts at the requested current remote base. |
| `python3 -c 'import jsonschema'` | 1 | Readiness probe found missing validation dependency; no validation success was claimed. |
| `python3 -m venv .temp/TASK-261021-yqt1qr/venv` | 0 | Isolated project-local validation environment. |
| `$PY -m pip install --no-cache-dir --disable-pip-version-check -r requirements-dev.txt` | 0 | Installed the repository-pinned validator dependency inside that venv only. |
| `GIT_DIR="$SPEC_GIT_DIR" GIT_WORK_TREE="$SPEC_ROOT" $PY -B tools/validate.py` | 0 | Validated 74 schemas, 1294 vector files, released-record/schema immutability and local link targets; specification checks, no runtime qualification. |
| `$PY -B -m unittest discover -s tools -p 'test_local_hosted_launches.py' -v` | 0 | All 10 existing exact-plan semantic corpus checks passed; claims below. |
| `$PY -B .temp/TASK-261021-yqt1qr/check-docs.py` | 0 | All 5 newly added local links/anchors resolve; added text has no em dash, guillemets or checked private identifier patterns. Existing link checker accepts a valid target and specifically rejects a planted missing target. |
| `git --git-dir="$SPEC_GIT_DIR" --work-tree="$SPEC_ROOT" diff --check` | 0 | No whitespace errors. |
| `git --git-dir="$SPEC_GIT_DIR" --work-tree="$SPEC_ROOT" commit -S -m 'cips: record accepted delegated local launch amendment'` | 0 | Created the four-file reviewed scope, signed. Agent Git environment was sourced first. |
| `git --git-dir="$SPEC_GIT_DIR" --work-tree="$SPEC_ROOT" log -1 --format='%H %G?'` | 0 | Head as above; signature status G. |

Publication: `git --git-dir="$SPEC_GIT_DIR" --work-tree="$SPEC_ROOT" push -u origin cip16-delegated-launches` exited 0. The agent Git environment was sourced first. The remote-tracking feature ref matches the signed head.

## Each hosted test and its claim

These tests read the existing exact-plan selection corpus. They do not execute a manager, provider, delegated request or runtime gate.

| Test | Claim proved by its assertions |
|---|---|
| `test_closed_case_and_runtime_matrix` | Existing corpus retains all 18 scenario names and the three runtime identities. |
| `test_inputs_bind_each_scenario_without_hidden_failure` | Each scenario has the exact declared inputs relative to the valid baseline. |
| `test_native_and_ax_never_probe_receiver` | Native and legacy ax corpus rows forbid receiver lookup, including native with backend selection. |
| `test_explicit_hosted_absence_is_stably_local` | Exact-plan explicit-hosted absent-backend and explicit-local rows retain local selection. This does not qualify delegated registry behavior. |
| `test_missing_and_unreachable_refuse_without_native_fallback` | Missing and unreachable receiver corpus rows keep distinct refusals; the valid local row remains admitted. |
| `test_defaults_require_migration_and_machine_gate` | Migrated operator default is the positive control; unmigrated and machine defaults retain refusal. |
| `test_ax_conflict_precedes_default_gate` | Hosted ax conflicts take precedence while native ax remains valid. |
| `test_dispatcher_selection_never_becomes_local` | Unimplemented dispatcher row refuses while explicit local selection remains valid. |
| `test_ingress_gates_refuse_before_contact` | Network, terminal and stdin corpus rows retain intended refusal/lookup outcomes; attached stdin has no preparation-contact claim. |
| `test_forbidden_observations_narrow_one_outcome` | Each corpus counterexample differs in exactly the named forbidden route or lookup outcome. |

## Self-review and limits

Read GOVERNANCE, the CIP process notes, CIP-0016, its protocol text, CIP-0014 and the CIP-0002/CIP-0013 revision examples. The change records acceptance of a design direction and a cross-reference, not normative wire adoption. The original exact-plan contract, section numbering and frozen schemas/vectors are preserved. Compatibility and security limits are explicit: delegated launches use live mutable content and do not offer CIP-0014 isolation or the exact-plan provenance guarantee.

No new executable gate or behavior was implemented, so new runtime tests and runtime narrowing mutants are inapplicable to this documentation change. Reused positive/negative corpus checks and a bounded missing-link control provide the relevant checks. Full Python/Go suites and live runtime probes were not run; a full-suite validation clone was therefore not needed. Validation ran on the same tracked content committed at the delivered head; the commit added no later content edits.

Local evidence: `.temp/TASK-261021-yqt1qr/validate-01.log`, `hosted-01.log`, `docs-01.log`, and the uncommitted `check-docs.py`; dependency setup log is `.temp/venv-setup-01.log`. These logs remain local and contain no product source changes.

## Open decisions

- Operator confirmation is still required for the proposed home lease.
- Operator confirmation is still required for the proposed exclusive profile home per Werkstatt.
- Request/record schemas and runtime qualification remain owned by curator-run and its consumers; this amendment claims neither.
