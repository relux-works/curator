# Review brief: environments schemas and conformance vectors

## Subject
- Worktree `~/Developer/ReluxWorks/.worktrees/curator-spec-env-schemas`, branch `draft/environments-schemas`, head `cef93fb`, base exactly `c3b29b1` (= main).
- Producer notes: `TASK-260901-1u50cr_env-schemas-notes.md` on the task (read all resources). Producer was a recovery run that independently re-verified a predecessor's work — treat the deliverable as one unit at `cef93fb`.

## Dimensions
1. **Schema fidelity**: each of the four schemas (`profilefile-v1`, `context-manifest-v1`, `agent-environment-marker-v1`, `launch-env-fragment-v1`) against the normative prose of `protocol/environments.md` at main — field sets, closed enums, strictness, required/optional split. Prose wins; any divergence is a finding.
2. **Vector correctness**: spot-recompute at least three determinism vectors BY HAND from the §5 byte rules (header grammar, chapter join, opencode.json CCJ-1+LF, zero-modules) and compare to expected bytes/hashes. Negative schema-cases actually violate what they claim.
3. **Producer's judgment calls** (notes): prose-over-brief on the header grammar; env-name↔environment binding left semantic; `copy_fallback` required iff `mode=linked`. Verdict each.
4. **Wiring**: vectors registered the same way existing ones are (manifest, tools/validate.py, Makefile, CI); `make validate` re-run yourself; generator determinism re-run yourself (twice, byte-identical).
5. **Repo hygiene**: delta files only under schemas/ conformance/ tools/ (list them); signed commit verified; no protocol/profiles/cli/CHANGELOG changes.

## Verdict
`review-findings-env-schemas-1.md` on TASK-260901-1u50cr; blocking/major -> development; else ACCEPT explicit, leave to-review. Do not mark done. Ignore tools/__pycache__ and .temp/venv.
