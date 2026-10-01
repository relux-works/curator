# Review note — TASK-260728-20ao7p native black-box + author guide + Windows external-build fix (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev2 (base bab2433b, tree 0d59fab2, 11 paths, gate green on every lane incl. Windows) against `20ao7p-brief.md` and
`20ao7p-gatefix-1.md`. The black-box found a real Windows defect: the external-build artifact `artifact` had no `.exe`, so the shim
failed. Verify:
1. Production fix: buildrepo.CacheArtifactName derives `artifact.exe` on Windows.
   - Every producer and consumer of that name agrees: the pipeline output, the cache path, the shim target, the receipt/marker fields.
   - The cache key and receipt bytes stay consistent with rc.13 (cite the clause). Existing Unix caches are unaffected.
   - Existing Windows caches without .exe: say what happens (rebuild, or a refusal with a clear message). No silent breakage.
2. The focused unit test covers the Windows name on all OSes. Mutant: drop the suffix → the test fails. Give the real exit code.
3. The black-box test drives install (build/cache/shim) → run the shim → reinstall (cache hit, no rebuild) → remove, through the real
   binary. The Windows HOME fixture is created via the private-dir helper, and nothing is skipped.
4. docs/external-build-repositories.md is accurate against the code and links the spec; the README link exists.
5. No CHANGELOG/LOGBOOK; no stray files; no Windows-reserved names.
accept_cr, or changes requested with file:line. Never spell any employer name.
