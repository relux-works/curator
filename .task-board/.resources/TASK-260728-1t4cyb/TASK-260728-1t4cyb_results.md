# TASK-260728-1t4cyb results
Change (uncommitted in story worktree): docs/external-build-repositories.md gains informative section
"Threat review for a new build driver" (12 checklist items: network, source/toolchain read-only, write roots,
process tree/resource bounds, executable allowlist/argv provenance, env scrubbing, cache identity, receipt fields,
symlink/hard-link, substitution/dry-run, fail-closed ordering admission->audit->cache->compiler, signing/exec policy),
each citing protocol/core.md §4.2/§4.2.1/§4.2.2/§6.4/§6.5/§8.1/§8.2/§9.2/§9.3/§10/§12.1-12.3 and decision 0005.
States Rust/Swift/Kotlin-JVM/C-C++/.NET are not generic Go equivalents; each needs its own versioned closed driver.
No new MUST, no vector or schema change. CHANGELOG Unreleased/Added entry.
Evidence (real exit codes; system python lacks jsonschema, so used venv /tmp/1t4cyb-venv):
- python tools/validate.py -> EXIT 0 ("validated 72 schemas and 1253 vector files"; includes local Markdown link check)
- doc JSON examples vs csk-skill-v7, agent-skill-v7, skill-build-v1, skillfile-dev-v2 schemas -> 0 errors each; new anchors resolve to existing headings
- python -B -m unittest discover -s tools -p test_validate.py -> EXIT 0 (551 tests, 805s)
- go test ./tools/... -> EXIT 0
Not run: other tools/test_*.py unit modules (release/signer gates unrelated to docs); conformance/draft-sources README gate (no vectors touched).
