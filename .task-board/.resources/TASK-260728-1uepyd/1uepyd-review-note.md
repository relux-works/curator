# Review note — TASK-260728-1uepyd rc.13 external-repository-acquisition consumer (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev2 (base 5432c85f, tree f7416750, 5 paths, gate green on every lane; rev1 failed only on a Windows trusted-git fixture)
against `1uepyd-brief.md` and `1uepyd-gatefix-1.md`. Verify:
1. Every row of conformance/v1/vectors/external-repository-acquisition.json is driven through the production entry
   (AcquireNetwork / AdmitLocal, or the narrowest production seam): all 12 cases, the 55 fetch-argv rows, the 17 clean-environment rows
   and the 11 forbidden-feature rows. The exact argv/env are compared, not substrings. Coverage is reported through conformancecoverage,
   and the exact count row is in conformance-case-counts.tsv.
2. Any production change that fixes a real mismatch is correct. Any known-gap or bound row is owned, with a precise reason; nothing
   skips silently. Review the 3 older bounds (local-config-and-refs 2, pack-index 1): driven or re-justified.
3. The Windows fix changed the fixture, not the probe, and adds no Windows skip.
4. Re-run the three mutants yourself with real exit codes: an extra fetch flag allowed; a clean-env variable leaked; a forbidden
   feature accepted. Each must be killed.
5. No CHANGELOG/LOGBOOK; no stray files; no Windows-reserved names.
accept_cr, or changes requested with file:line. Never spell any employer name.
