# TASK-261011-2xdx9u — R-CS2-contract-fix recovery and policy boundary

## Candidate and source repair

The predecessor's docs-only candidate covers 7/7 named R-CS2 A-rows and 8/8 in-scope C4 CIP rows. The successor read the pinned contract/audit and reviewed the inherited dispositions, owner/version coordination, managed PATH and first-consumer amendment. No product code, schemas or diagram changes were introduced by the successor. The four local donor PlantUML figures remain R-CS1 scope; R-CS2 names no existing local figure.

The predecessor validation outcome `TASK-261011-2xdx9u_change-request_rev1-validation.log` records exit 1 at command 1/1, caused by a broken local-link interpretation of a board-resource citation. The successor replaced the two such citations in the shared disposition/alignment with plain named references to the supplied owner-decision task precondition. The full wrapper was not rerun.

## Fresh direct evidence

| Command | Real exit code | Scope |
|---|---:|---|
| `python3 tools/validate.py` using the project validation environment | 0 | 73 schemas, 1294 vector files and repository local Markdown links; rerun after evidence edits |
| `git diff --check` | 0 | Tracked documentation whitespace, rerun after evidence edits |

The predecessor temporary document-packaging result is retained as historical evidence and was not rerun by this successor. No tests, builds, renderer, provider qualification or commits ran. Repository changes remain uncommitted.

## External policy constraint

The configured handoff command invokes `spec-gate.sh`. Its ordered steps are:

1. `python3 tools/validate.py`
2. `python3 -B -m unittest discover -s tools -p 'test_*.py'`
3. `go vet ./tools/...`
4. `go list ./tools/...`, then `go test -c -o /dev/null` for each package

The task's ONLY CURRENT INSTRUCTION explicitly says “No tests or builds on this host.” Calling developer handoff under this configuration would now progress from the repaired document check into the forbidden tests/builds. The handoff CLI exposes no per-task validation override. The shared wrapper and project validation configuration have not been modified.

No workaround or false green is used: fixing the document links resolved the prior validation failure, but does not resolve this policy conflict. The exact action needed from the coordinator is to configure documentation-only validation for this docs-only handoff, or provide a supported route to validation on a permitted external host. Documentation-only handoff validation is recommended because no executable, schema or conformance-vector bytes changed. Continuing with the current wrapper would violate the brief; bypassing validation or changing a global shared gate within this leaf would exceed the task's scope.

After policy correction, rerun the applicable document check, update the alignment/evidence resources if needed, then run `task-board handoff TASK-261011-2xdx9u --role developer`. Handoff has not been claimed by this successor.
