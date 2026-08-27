# Mandatory composite rework

The first RUN-260729-9ead43 outcome is diagnostic only and must be superseded. It validated origin/main without the accepted compiled-build implementation.

1. Use /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-jrrgw9/worktree as the authoritative accepted product candidate. Verify it against TASK-260720-jrrgw9_production-integration-results.md, verifier4 evidence, and the final review verdict. This candidate already represents TASK-260729-2kaopg plus the accepted TASK-260720-2qqq0w implementation and accepted jrrgw9 integration.
2. Create a fresh TASK-260720-1pvfj5 rework composite; preserve every accepted product byte. Overlay only the CI/Makefile/narrow quality files that remain valid from the cancelled run. Do not use origin/main-only package inventory.
3. CI gates must exercise actual godriver, buildcache/buildsource, transaction/cache, install/atomicity, interop/conformance, DACL/reparse, readonly-source, resource-policy, executable and no-follow behavior as present in the accepted candidate. Do not downgrade an AC because origin/main lacks it.
4. Keep Go 1.25.5, the current qualified released pin, and candidate-only immutable revision/root semantics. No pin advancement or release/conformance claim.
5. Do not overlap heavy Go suites. Reuse accepted verifier4 full/race evidence where byte identity permits, and run only focused CI-script selftests/gates needed for the CI delta. Record exact composite provenance and hashes.
6. No stage, commit, publish, pin change, broad suppression, fixture weakening, timeout inflation, or product-code mutation. Attach a superseding task-scoped patch/results packet and hand off only when the exact accepted composite is proven.