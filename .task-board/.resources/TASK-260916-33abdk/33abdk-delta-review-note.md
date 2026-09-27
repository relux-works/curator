# Delta review note — TASK-260916-33abdk rev3 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Rev2 was ACCEPTED on content (your verdict). Rev3 re-applies it on trunk eca2bf27 (tree 84fdbc5b, gate green, same 11 paths). Orchestrator
checks: 6 paths byte-identical to rev2; cmd/curator/envstatus.go, internal/envprofile/status.go, conformance-case-counts.tsv equal the
`git merge-tree` of (55b94af2 → refs/campaign/1i1gfo-rev2-20260927) onto eca2bf27. Review ONLY:
1. Two test additions present in neither rev2 nor trunk: cmd/curator/env_credential_marker_test.go
   TestEnvResolvePreservesPreRuleCodexSeedAndReportsUnstrippedHome and cmd/curator/envstatus_test.go
   TestEnvResolveStripsAndReportsInlineNativeCodexMCPTable — are they correct per rc.13 §7.4/§8.2 and passing (real exit codes)? Tests only
   strengthen coverage; confirm no production code changed beyond rev2.
2. .github/ci/conformance-gaps.tsv: 6 lines differ from the merge-tree result — trunk's rows kept, only the 2 STORY-260916-1i1gfo rows removed,
   nothing re-added. Show the diff.
accept_cr or changes requested with file:line. No LOGBOOK.md.
