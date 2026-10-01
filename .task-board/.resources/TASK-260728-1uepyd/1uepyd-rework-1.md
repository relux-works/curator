# TASK-260728-1uepyd — rework 1 (THE ONLY CURRENT INSTRUCTION, with 1uepyd-brief.md)

The rev2 review (`TASK-260728-1uepyd_review-verdict-rev2.md`) accepted the builders, the reorder and M1–M3. One blocking finding: the
consumer proves the BUILDERS but not what the production call site hands to Git. Both call-site mutants survive:
- M4: an extra `-c fetch.prune=true` at admission.go:410;
- M5: a `GIT_SSL_NO_VERIFY=1` env leak at admission.go:356.

The fix is test and shim only. No production change is expected.
1. acquisition_conformance_test.go:512-519: compare the FULL logged fetch argv against the vector's common_fetch_argv, resolved per
   case. Resolve `<operation-private>` from the logged --git-dir / core.hooksPath values after asserting they sit under one private
   root, and resolve `<selected>` from the case transport.
2. testdata/acquisitiongitshim: also log the environment Git received, and compare it with the vector's clean_environment plus the
   transport additions and the platform allowlist (the same `want` map testExternalRepositoryCleanEnvironment builds). A single
   unexpected or missing variable must fail.
3. Re-run M4 and M5 and show both killed, with real exit codes. Keep M1–M3 killed.
4. Residuals, fix or state as bounds:
   - R-a: the helper-selected-transport row checks a constant; make it real or remove it;
   - R-b: success cases assert vector data only; drive them through RunPipeline with an ordering hook, or state a bound;
   - R-c: no action.
Set status development, update the results, run `task-board handoff TASK-260728-1uepyd --role developer`, then END YOUR TURN. No
CHANGELOG/LOGBOOK edit. No Windows-reserved names.
