# TASK-261021-jr7tu7 — blocked editorial draft, r1

## Delivery state

Branch `repository-transport-lanes` was created from fetched `origin/main`,
base/head `e41f96387287f5e76604627f9e63662f7d6d86d4`.
The draft is uncommitted and unpushed. No delivery head is claimed.
Board status was not changed. Draft patch: `TASK-261021-jr7tu7-r1.patch`,
attached separately as an outcome. No checker, schema or runtime was changed.

## Files changed in the draft

- `protocol/repository-transport.md`: four-item normative Covered lanes list
  in §1, acquisition-operation scope, explicit SSH-without-policy example,
  local/non-Git exclusions, and revision 2 §4 pointer.
- `conformance/skillfile-sources-v1/manifest.json`: regenerated digest for that
  prose input only.
- `conformance/candidate.json`: regenerated enclosing suite digest only.

## Acquisition inventory and own-section comparison

| Lane | Existing sections | Scope/limits |
| --- | --- | --- |
| Schema-2 Skillfile network sources, individual/collection selection | `protocol/skillfile-sources.md` §1 | `git`/`repository`; closed core §6.3 declaration grammar; local `path` is separate |
| Legacy Skillfile/configured network sources, project/global/hybrid | `protocol/core.md` §5; `profiles/manager.md` §§2, 2.6, 4; `cli/curator.md` upgrade/update description | Existing declared URL/configured source; selected-closure upgrade and all-source update retain their bounds |
| Transitive skill dependencies | core §4.4 | Exact tag/revision; selected package directory does not change endpoint selection |
| Skill network development substitutions | core §5 | Existing `git` plus exact ref; strict-audit rejection retained |
| Profile install root and Git overlays | manager §12.3; environments §§1, 6, 9.1–9.2 | Range/tag/revision; no profile branch admission; overlays seed the joint closure |
| Required context packages, skills and MCP packages, including transitive requirements | environments §2 `requires`; §2.2; manager §12.3 | `agent-context.json` context packages and Git-only `agent-mcp.json` packages share acquisition/closure rules |
| External build repositories and network build substitutions | core §§4.2.2, 5, 6.3; manager §§11, 11.2–11.3 | Strict endpoint/executor/broker/raw-object rules remain; no ports/alias admission |
| Refresh, upgrade, repair and missing locked-snapshot reacquisition | manager §2.6, §11.9, §12.3; Skillfile sources §3–4; environments §10.1 | Covered only when the owning operation already requires network Git I/O; no new fetch on status/store-only operations |
| Local filesystem/file sources and committed-HEAD local substitutions | core §§5–6; Skillfile sources §2; environments §1 | No network endpoint selection; retain local rules, including configured-Git sources lacking network identity |
| Registry HTTP and MCP server connections | registry protocol; environments §2.2; manager §6 | Not Git package acquisition; no manager installation/execution of MCP server software authorized |

No inspected lane-owning clause explicitly requires canonical-HTTPS fetch or
contradicts repository-transport §1/§2 endpoint selection. The environments §2
phrase "canonical git source" is an identity requirement referring to core
§6.1, not an HTTPS endpoint mandate. Existing no-port/strict-build limits are
intentional restrictions, not contradictions, and were preserved. No lane-owning
section was edited. Draft remote-workplace materialization is an archive/import
contract, not a published worker Git acquisition permission (remote-worker
R-TRANSFER/R-NET-TOOL); it does not add another Git lane.

## Exact blocker and required decision

The brief requires citations to manager §12.3 and context/MCP sections. The
existing CI command `python tools/verify_skillfile_sources_independence.py`
rejects these citations in `protocol/repository-transport.md` because they are
not present at `v1.0.0-rc.10`. It permits only explicitly allowlisted exact
conditional-capability clauses. The draft produces five intended diagnostic
lines: manager §12.3 and environments §§1, 2, 1, 2.2.

This cannot pass the current prose gate merely by regenerating metadata.
An honest fix would explicitly condition the profile references on implementing
the agent-environments capability, add narrowly scoped exact citation exceptions
to `CONDITIONAL_CITATION_ALLOWLIST`, and add positive/negative/narrowing controls
for those exceptions. That changes checker code/tests, beyond the brief's
editorial-only scope. No wording obfuscation or gate bypass was attempted.
Please authorize that bounded checker follow-up or provide another owner-approved
publication route for the required references. No product design decision remains.

## Commands, evidence and claims

Commands below run from the spec checkout. `PY` denotes the existing validation
venv's Python; `TASK_TMP` denotes the task-local `.temp/TASK-261021-jr7tu7/tmp`.
Git operations explicitly use this repository's separate git dir/work-tree.
Logs and scratch checks are under `.temp/TASK-261021-jr7tu7/`.

| Command/check | Exit | Claim and limit |
| --- | ---: | --- |
| `task-board q 'get(TASK-261021-jr7tu7) { status assignee }'`; resource enumeration/get | 0 each | Task exists, assigned to this lane; sole precondition brief materialized/read |
| `git fetch origin`; `git switch -c repository-transport-lanes origin/main` | 0 each | Fresh named branch from the named fetched base; clean starting tracked state |
| `"$CURATOR_GO" version` | 0 | Configured build tool runs; `go-readiness.log` |
| `TMPDIR=$TASK_TMP env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" run ./tools/generate-vectors -root .` | 0 | Canonical generator updates only the two required digest metadata files; `regenerate-01.log` |
| `GIT_DIR=<spec-git-dir> GIT_WORK_TREE=<spec-work-tree> TMPDIR=$TASK_TMP "$PY" -B tools/validate.py` before regeneration | 1 | Correctly rejects stale repository-transport prose digest; `validate-01.log` |
| Same validator after regeneration | 0 | 74 schemas/1294 vector files, manifests and local links pass; does not claim runtime behavior; `validate-02.log` |
| Same repository/TMPDIR setup, `"$PY" -B tools/verify_skillfile_sources_independence.py` | 1 | 89/89 schema references match rc.10; five new post-baseline citation refusals block publication; `prose-01.log` |
| `TMPDIR=$TASK_TMP env -u GIT_DIR -u GIT_WORK_TREE "$PY" -B -m unittest discover -s tools -p test_skillfile_sources_independence.py -v` | 0 | Eight existing checker controls pass; `prose-controls-01.log` |
| `TMPDIR=$TASK_TMP "$PY" -B .temp/TASK-261021-jr7tu7/check-links.py` | 0 | All 20 added local links/anchors resolve; real link checker accepts existing-file positive and rejects planted missing-file negative with intended diagnostic; `links-01.log` |
| `git diff --check` | 0 | No patch whitespace errors |
| `git diff --binary > <outcome-dir>/TASK-261021-jr7tu7-r1.patch` | 0 | Complete reviewable draft preserved outside repository |

The eight existing prose-gate tests prove these bounded claims: fixture candidate
accepted; rc.12-only schema dependency rejected; changed baseline definition
bytes rejected even when JSON is valid; rc.12-only manager citation rejected;
exact conditional clause accepted; unconditioned same citation rejected;
unrelated informative marker does not excuse a post-baseline citation; exact
clause in the wrong file remains rejected. Their passing results confirm the
blocking gate is behaving as designed, not that this draft passes it.

No behavior change or new gate was implemented, so no new product tests or
mutants were added. Full Python/Go/runtime suites and deterministic second
regeneration were not run: publication is blocked by the required prose gate;
this is not a green delivery. No validation clone was needed for these narrow
checks, and no background process remains.
