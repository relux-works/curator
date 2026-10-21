# TASK-261021-yiyrus results, revision 1

## Delivery

- Repository: `relux-works/curator-spec`.
- Branch: `cip-mcp-in-project-skillfile`, created from fetched `origin/main`.
- Base: `e41f96387287f5e76604627f9e63662f7d6d86d4`.
- Signed head: `38aa3a0c777400a448f3eb7e7b2fed57a85654e2`; verification verdict `G`.
- Branch pushed and remote head read back at that SHA.
- No PR opened, default branch changed or board status changed.

## Summary

Filed CIP-0022, **Draft**, using the next unassigned number. It proposes
project-only schema 3 inheriting schema 2; URL, closed server-object and Git
package arms; neutral writer-local `install`; canonical environment ids;
package-first allowlist/audit resolution; and later CIP-0002 inline approval
bound to project identity, NAME and normalized server digest. It distinguishes
server, declaration, raw manifest and effective plan identity; known unsupported
environments report per-pair skips rather than refusals. It covers secret
blocking, no install-time execution, typed refusals, reporting and compatibility.

No normative schema, released artifact, executable implementation or conformance
corpus changed. Governance/CIP-0001 require normative artifacts at adoption,
not executable stubs for a filed Draft. Future positive/negative vectors and
narrowing mutants are enumerated in the CIP instead.

## Files changed

- `cips/CIP-0022-mcp-server-declarations-in-project-skillfile.md`: complete proposal and owner questions.
- `cips/README.md`: indexed CIP-0022 as Draft.
- `README.md`: proposal discovery entry, clearly non-normative.
- `CHANGELOG.md`: Unreleased documentation entry.
- `UNRESOLVED_QUESTIONS.md`: discovery entry linking the Draft's owner questions.

## Validation and command evidence

Commands ran in the specification checkout. Git operations used its explicit
separate git-directory/work-tree binding; those machine-local paths are omitted
from this public report. The local Python environment is task-scoped under
`.temp/TASK-261021-yiyrus/venv`, using repository-pinned dependencies. No tools
were installed outside the workspace.

| Command | Exit | Claim and limits |
|---|---:|---|
| `git fetch origin` | 0 | Updated the base before creating the named branch. |
| `git switch -c cip-mcp-in-project-skillfile origin/main` | 0 | The task branch starts at the named current base. |
| `python3 -B -m venv .temp/TASK-261021-yiyrus/venv` | 0 | Created a workspace-local validation environment. |
| `TMPDIR="$PWD/.temp/TASK-261021-yiyrus" PIP_DISABLE_PIP_VERSION_CHECK=1 .temp/TASK-261021-yiyrus/venv/bin/python -B -m pip install --no-cache-dir -r requirements-dev.txt` | 0 | Installed only the repository's pinned validation dependency and its dependencies into that environment. |
| `PYTHONDONTWRITEBYTECODE=1 .temp/TASK-261021-yiyrus/venv/bin/python -B tools/validate.py` | 0 | Existing schema, vector, immutable release and local-link checks pass: 74 schemas and 1294 vector files; this is specification-data validation, not runtime qualification. |
| `PYTHONDONTWRITEBYTECODE=1 .temp/TASK-261021-yiyrus/venv/bin/python -B .temp/TASK-261021-yiyrus/check-draft-01.py` | 0 | All template sections/metadata retained; Draft agrees with its unique index row; JSON example parses; new proposal has no implementation-specific product name or private path; actual local-link checker passes. |
| Same documentation check, positive link fixture and broken-link control | 0 | The real repository checker accepts an existing target, then specifically rejects `missing.md` as a broken local link; disposable fixture removed. No runtime invariant is claimed. |
| `git diff --cached --check` | 0 | Reviewed staged scope has no whitespace errors. |
| `git diff HEAD^ HEAD --exit-code -- protocol profiles schemas conformance release tools` | 0 | Delivered commit leaves normative/executable/corpus paths unchanged. |
| `git commit -S -m 'docs: draft CIP-0022 for project Skillfile MCP declarations'` | 0 | Created the signed task commit. |
| `git log -1 --format='%G? %H'` | 0 | Signature verdict `G` at the handed-over SHA. |
| `git push -u origin cip-mcp-in-project-skillfile` | 0 | Published only the named feature branch. |
| `git ls-remote origin refs/heads/cip-mcp-in-project-skillfile` | 0 | Published head equals the delivered signed head. |
| `git status --short` | 0 | Clean tracked/untracked task scope after commit; scratch is ignored. |

Local evidence: `.temp/TASK-261021-yiyrus/validate-01.log`, `validate-02.log`,
`draft-check-01.log`, `remote-head-01.log`, `deps-01.log` and tool-readiness logs.
The first bare-Python dependency probe failed with `ModuleNotFoundError:
jsonschema` (exit 1); the isolated environment above resolved it. The final
validator run followed the final documentation edits. Full Go and Python unit
suites were not run: this delta files a proposal and indexes without changing
runtime behavior, schemas or the validation instrument. Relevant validation
and the bounded link-checker negative control passed. No native adapter or
server process was launched.

## Open decisions for the owner

1. Stage-A native adapter set and the exact package-first adoption boundary.
2. Leave nested `install` semantics undefined initially, or standardize typed
   headers/auth with separate effective-configuration approval.
3. Adopt the proposed extra URL character guards and define a complete shared
   `npx`/`uvx` runtime pin grammar before claiming enforcement.
4. Keep inline audit waivers unavailable initially, or define a protected,
   digest-bound waiver identity without weakening package policy.
5. Freeze versioned carriers, native paths, ownership/transaction/drift/removal
   contracts and capability advertisement in the adoption leaf.

The recommendations are in the CIP. These questions intentionally keep its
status Draft; filing it makes no implementation or launch capability claim.
