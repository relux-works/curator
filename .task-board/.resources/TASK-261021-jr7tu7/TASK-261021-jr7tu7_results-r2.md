# TASK-261021-jr7tu7 — results r2

## Delivery

Repository: `curator-spec`.
Branch: `repository-transport-lanes`.
Head: `d63a1ddf375ed5e40e7dde641a9beea83d67bf05`.
Signed commit: `git log -1 --format=%G?` returned `G`.
Base: fetched `origin/main`, `e41f96387287f5e76604627f9e63662f7d6d86d4`.

This supersedes r1's blocked draft. The orchestrator's answer in
`rt-covered-lanes` authorized reversed reference direction: plain lane names
in repository-transport, backlinks from the owning sections, no checker
exceptions. The delivered change follows that direction. The independence
checker and all its exceptions are unchanged. No board status was changed.

## Files changed

- `protocol/repository-transport.md`: §1 Covered lanes list; profile SSH
  requirement/no-policy example; explicit local/non-Git exclusions;
  revision 2 §4 points to the list. Existing grammar/ref/capability limits stay
  explicit.
- `protocol/skillfile-sources.md` §1: source-acquisition backlink and URL
  identity/connection-hint clarification.
- `protocol/core.md` §§4.4 and 5: the same backlink for skill dependencies,
  legacy/configured network skill sources and network substitutions.
- `profiles/manager.md` §§11 and 12.3: the same backlink for external build
  repositories/network substitutions and the full profile closure.
- `protocol/environments.md` §§2 and 2.2: the same backlink for network context
  packages and Git-sourced MCP declaration packages.
- `conformance/skillfile-sources-v1/manifest.json`: regenerated transport prose
  digest only.
- `conformance/draft-sources-v2/manifest.json`: regenerated live Skillfile
  sources prose digest only; accepted tagged prose pin is unchanged.
- `conformance/candidate.json`: regenerated enclosing source-suite digest only.

No schemas, runtime code, checker code, conformance case expectations or release
records changed. This clarifies existing endpoint scope; it creates no acquisition
lane, endpoint grammar, new version or permission.

## Acquisition inventory and contradiction review

| Lane | Owning sections | Existing limits retained |
| --- | --- | --- |
| Skillfile network sources; individual/collection selections | Skillfile sources §1 | New `git`/`repository` grammar remains narrower than legacy grammar |
| Legacy/configured network skill sources; project/global/hybrid | core §5; manager §§2, 2.6, 4 | Existing scopes and selected-closure/all-source fetch rules |
| Skill dependencies | core §4.4 | Exact tag/revision; package directory does not determine transport |
| Network development substitutions | core §5; manager §11 | Existing substitution forms, strict-audit refusal, effective-source identity |
| Profile installs, Git roots/overlays, direct/transitive closure requirements | manager §12.3; environments §§1, 2, 6, 9.1–9.2 | Profile range/tag/revision only; no branch admission |
| Git context packages (`agent-context.json`) | environments §2 | Network acquisition follows existing joint closure and source rules |
| Git MCP declaration packages (`agent-mcp.json`) | environments §2.2 | Git-only packages; package acquisition does not install/execute a server |
| External skill build repositories | core §§4.2.2, 6.3; manager §11 | Strict transport/executor/broker/raw-object rules; no ports/alias widening |
| Update/refresh, upgrade, repair, locked-snapshot reacquisition | Skillfile sources §§3–4; manager §§2.6, 11.9, 12.3; environments §10.1 | Covered only where the owning operation already calls for network Git I/O |

Local filesystem/file acquisition and committed-HEAD local substitutions have
no network endpoint selection. Registry HTTP requests and MCP server connections
are separate transports, not Git package acquisition. Draft remote-workplace
materialization is archive/import, not published worker Git fetch authority.

No inspected lane-owning section explicitly mandates canonical-HTTPS fetch or
contradicts repository-transport §1/§2. Environments §2's "canonical git source"
refers to core §6.1 identity, not an HTTPS endpoint mandate. The profile-closure
SSH example now explicitly says the no-policy path attempts the declared SSH
endpoint once and cannot synthesize HTTPS from canonical identity. No conflicting
lane requirement was silently edited.

## Verification, commands and exit codes

All final checks below use the exact signed head named above. `PY` denotes the
existing validation venv Python; `TASK_TMP` is the task-local
`.temp/TASK-261021-jr7tu7/tmp`. Commands run from the spec checkout. Git commands
use the explicit separate spec git-dir/work-tree; validator/prose commands set
`GIT_DIR=<spec-git-dir> GIT_WORK_TREE=<spec-work-tree>` for their internal Git
reads. Logs are in `.temp/TASK-261021-jr7tu7/`.

| Command | Exit | Claim/evidence |
| --- | ---: | --- |
| `a2a inbox`; board task/assignee/precondition query; brief materialization | 0 each | Authenticated orchestrator answer read; assignment and sole brief reconfirmed |
| `git fetch origin` | 0 | Remote refs refreshed before continuing existing task branch |
| `TMPDIR=$TASK_TMP "$PY" -B tools/validate.py` with correct repository Git environment | 0 | 74 schemas/1294 vector files, manifest integrity and local file links pass; `validate-04.log` |
| Same environment, `"$PY" -B tools/verify_skillfile_sources_independence.py` | 0 | 89/89 referenced schema definitions byte-identical to rc.10; 34/34 cited clauses present; only the one existing conditional exception; `prose-03.log` |
| `TMPDIR=$TASK_TMP "$PY" -B .temp/TASK-261021-jr7tu7/check-final-links.py` | 0 | All eight added links/anchors resolve: seven lane backlinks and revision-2 pointer; `links-03.log` |
| Same link check, isolated real-checker controls | 0 | Existing-file positive accepted; planted missing file rejected with intended `broken local link` diagnostic; no product fixture changes |
| `TMPDIR=$TASK_TMP env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" run ./tools/generate-vectors -root .` | 0 | Canonical generator maintains the three required digest-only metadata changes; `regenerate-03.log` |
| Same generator at signed head, then `git diff --exit-code` | 0 each | Regeneration at the delivered head is deterministic and leaves a clean tracked tree; `regenerate-04.log` |
| `git diff --cached --check`; `git commit -S -m 'docs: enumerate covered repository transport lanes'` | 0 each | Reviewed eight-file scope committed with explicit signing |
| `git log -1 --format='%H %G?'` | 0 | Exact signed head above; signature verdict `G` |

Earlier r1 evidence remains attached and is not misrepresented as a final-head
run: `prose-controls-01.log` records eight passing existing independence-gate
unit controls. They prove candidate acceptance, altered-definition and
post-baseline-reference rejection, and exact conditional-exception boundaries
including wrong-file and unconditioned negatives. Checker code/configuration
never changed, so that instrument-control evidence remains applicable.

The r1 rejected-citation gate run and stale-digest validator run remain in the
history and have been resolved by reference inversion and canonical regeneration,
respectively. No bypass was added. No new behavior/gate warrants product tests
or narrowing mutants for this editorial delta; existing checker controls and
real link positive/negative controls support the checks used here. Full
Python/Go/runtime suites were not run, and runtime qualification is not claimed.
Narrow checks did not require a full-suite validation clone.

## Open decisions

None for this task. No PR was opened and no default branch was pushed.

## Publication

`git push -u origin repository-transport-lanes` exited 0.
`git rev-parse origin/repository-transport-lanes` exited 0 and returned
`d63a1ddf375ed5e40e7dde641a9beea83d67bf05`, equal to the tested signed head.
Final `git status --porcelain` exited 0 with no output. No background process
remains. This results file is attached as `TASK-261021-jr7tu7_results-r2.md`.
