# TASK-261021-krcy1c results r3: signed replay onto current main

## Delivery

- Repository: curator-spec; new branch: `cip02-ready-for-acceptance-r3`.
- Fetched and verified base: `c8b6245485c1a155fdf5148b14107712784a9021`, including the CIP-0016 delegated amendment and CIP-0022 Draft.
- New signed commits: `bd601b94686b5777ff69cf08f95b01e70aef9d74` and head `9702382be175e2297e943f1ea97cc6e6bd8dda8f`. Both signature verdicts are `G`.
- Ordinary push of the new branch exited 0; remote-tracking readback equals the tested head. The previously reviewed branch and commits were not rewritten. No force push, PR, main push or board-status mutation.

## Mechanical changes only

The two previously reviewed commits, `752307c6a825c07256521e818fef25ba5ae78fc7` and `f202f41ba4718b82cdfefac28c8dfaa9af3a7de5`, were reapplied with signed cherry-picks.

The first cherry-pick stopped on two independent EOF additions in the unresolved-question index. Both complete entries were retained, without editing either entry's prose. The old companion branch reference in CIP-0002 was replaced with the authorized landed link to `CIP-0022-mcp-server-declarations-in-project-skillfile.md`. The second commit applied unchanged.

Files changed relative to the new base:

- `cips/CIP-0002-project-context-in-managed-launches.md`: original reviewed candidate plus the mechanical CIP-0022 reference update.
- `UNRESOLVED_QUESTIONS.md`: the original reviewed CIP-0002 entry appended after main's intact CIP-0022 entry.

An exact byte comparison proves that the whole CIP differs from reviewed r2 only in the authorized reference paragraph. The index equals the complete new-main bytes followed by the complete old task addition. README, CHANGELOG, index entries for other CIPs, schemas, validators and conformance inputs are untouched. Draft status, owner decisions, F1 manifest/lock selection contract, remaining owner questions and qualification limits are preserved.

## Validation and commands

Evidence directory E: `~/Werkstatt/curator-dev4/curator-spec/.temp/TASK-261021-krcy1c/r3/`. PY is the existing task-local validation venv. Go uses the already configured CURATOR_GO (1.25.5). No installation was needed. Source Git commands use the explicit spec git-dir/work-tree; validation clone commands remove inherited GIT_DIR/GIT_WORK_TREE.

| Command/check | Exit | Claim and evidence |
| --- | ---:| --- |
| `a2a inbox`; task/assignee/preconditions query; five resource gets; ack | 0 each | Latest addressed orchestrator rebase request and board assignment confirmed; all five preconditions read. Authority/scope recorded in `rebase-request.md`. |
| `git fetch origin`; local main fast-forward and HEAD/origin comparison; new branch creation | 0 each | The new branch starts at the fetched main SHA above, rather than the old task branch. |
| `git cherry-pick -S 752307c6 f202f41b` | 1, expected conflict | Independent index additions conflict. Original conflict retained in `unresolved-conflict.txt`; no semantic conflict resolved. |
| `git cherry-pick -S --continue` after scoped resolution | 0 | Both replay commits created with signatures G; second patch applied unchanged. |
| `PY -B E/check-rebase.py` | 0 | Two-file scope, exactly two signed commits, main ancestry, exact CIP byte equivalence except the authorized reference, complete preservation of both index entries, valid landed CIP target and unchanged second patch. `rebase-check.log`. |
| `PY -B tools/verify_skillfile_sources_independence.py` with explicit spec Git environment | 0 | Existing prose gate passes: 89/89 schema references, 34/34 cited clauses, existing conditional exception unchanged. `prose.log`. |
| `PY -B tools/validate.py` in sanitized clone environment | 0 | 74 schemas and 1294 vector files validate at the delivered head, including local Markdown links. `validator.log`, `validation-exits.json`. |
| `env -u GIT_DIR -u GIT_WORK_TREE PY -B E/suite-groups.py N`, N=0..124 | 0 each | All 879 uniquely discovered Python tests run exactly once in 125 groups, including all 40 release-gate tests. Retained positive/negative controls remain green. Every group records IDs, command, count, duration, exit and matching unittest OK summary. `suite-groups.json`, `python-001..125.{json,log}`. |
| `"$CURATOR_GO" test ./tools/...` in sanitized clone environment | 0 | Full existing Go tools test suite passes at the delivered head. `go.log`, `validation-exits.json`. |
| `PY -B E/final-check.py` | 0 | All group records/counts/OK summaries agree; both normal-.git clones are clean at the exact delivered head; source is clean, both commits G, main ancestry and whitespace checks pass. `validation-summary.json`. |
| `git range-diff --no-color 752307c6^..f202f41b c8b6245..9702382` | 0 | Full output below and in `range-diff.txt`; only first-patch mechanical context/reference differences, second patch identical. |
| `git push -u origin cip02-ready-for-acceptance-r3`; tracking-head readback | 0 each | New branch published; readback equals the signed, tested head. |

All three recipes shown by `make -n validate` were completed at the exact delivered head in disposable normal-.git clones. Python discovery was split per the standing long-suite instruction; this is full recipe-equivalent validation, not a claim of a monolithic `make validate` invocation. Release-gate tests ran individually; other groups contained at most ten tests. The slowest group took 188.383 seconds. Every suite invocation used `TMPDIR=/tmp/cw-dev4` and removed inherited GIT_DIR/GIT_WORK_TREE; Go caches stayed inside this Werkstatt. Both clones were deleted after final clean HEAD/status verification. Logs remain in E.

No behavior or checker changed, so no new product tests/gates were introduced. Existing executable regression and negative controls were retained and run. This is a mechanical replay of the reviewed Draft, with no new implementation or native runtime qualification claim.

## Open decisions

None for this rebase. The candidate's existing later owner questions about root/system replacement and private-memory retention remain unchanged. The orchestrator owns review, acceptance and landing.

## Full range-diff

```diff
1:  752307c ! 1:  bd601b9 cips: prepare project context candidate for acceptance
    @@ Commit message
         cips: prepare project context candidate for acceptance
     
      ## UNRESOLVED_QUESTIONS.md ##
    -@@ UNRESOLVED_QUESTIONS.md: interactive hosted rows remain distinct. `tb-sessiond` lookup/restart is
    - refused under existing hosted §§4/6; enabling it requires an explicit later
    - amendment, not a callback or hidden wrapper. Unqualified vendor child-leak or
    - settings-reload behavior keeps capability unavailable.
    +@@ UNRESOLVED_QUESTIONS.md: keeps the stage-A adapter set, nested `install` contract, runtime pin grammar,
    + inline audit waivers and versioned carrier/writer details open for operator
    + review. Its package-first and digest-approved inline design is proposed only;
    + filing the Draft enables no native writer or managed-launch admission.
     +
     +## Project context acceptance candidate (CIP-0002)
     +
    @@ cips/CIP-0002-project-context-in-managed-launches.md: The operator confirmed the
     +
     +On 2026-10-21 the owner also accepted inline MCP in the project Skillfile as
     +an individually approved project contribution, bound to the normalized
    -+`server` object digest. The companion draft, **MCP server declarations in the
    -+project Skillfile**, is on branch `cip-mcp-in-project-skillfile`; this base has
    -+no numbered main-branch entry for it, so this proposal refers to it by title.
    -+It owns Skillfile parsing and server normalization, not launch authority.
    ++`server` object digest. The companion draft,
    ++[CIP-0022: MCP server declarations in the project Skillfile](CIP-0022-mcp-server-declarations-in-project-skillfile.md),
    ++owns Skillfile parsing and server normalization, not launch authority.
     +
     +The same date's accepted CIP-0016 amendment admits `launch-delegated/1` as an
     +interim local hosted route. [Launch route binding](#launch-route-binding)
2:  f202f41 = 2:  9702382 docs: bind project context selection to package manifests and locks
```
