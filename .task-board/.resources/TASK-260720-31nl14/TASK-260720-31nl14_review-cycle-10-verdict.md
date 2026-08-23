# TASK-260720-31nl14 review cycle 10 verdict

## Verdict

Changes requested. Route to to-dev, then require a fresh native Windows reviewer cycle.

## Native Windows evidence

The exact current detached transaction candidate at base 17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8 was cross-compiled with Go 1.25.5 as a Windows amd64 coverage test binary. Local and remote SHA-256 matched: 72115570f3ff5a9eeb5aabada64e58e456b18b46a5508972f71d37a46428e137; size 6635520 bytes. Native execution ran on Windows 10.0.19045 AMD64. The complete internal/transaction suite failed with 19 top-level failing tests, 101 Access-is-denied lines, final FAIL, and 69.9 percent partial coverage. The isolated TestPrepareCanonicalJournalReferencedKeysAndCommitCleanup reproduction exits 1 with the same durability failure. Raw log SHA-256: d447524147cbe5bae6e263299d1ddd249d02c0778d126ef47583226e92d8955c. Coverage SHA-256: 4455e14f02386e3e080bb069bd371d4124959b24e20ab66e23a3d0308c58152b. Local and remote hashes match and remote artifacts were removed with absence verified.

## P1 findings

1. Regular-file durability is not valid on Windows. internal/transaction/files.go lines 108-119 opens every existing regular target with os.Open and then calls File.Sync. On Windows that read-only handle reaches FlushFileBuffers and returns Access is denied. The engine invokes this path after backup, install, and restore renames at engine.go lines 374, 428, and 643, so ordinary commit, exact reverse rollback, cleanup, restart recovery, subprocess crash recovery, and directory targets all fail natively. The Windows implementation must establish real byte and namespace durability using handles and primitives with valid access; it must not merely swallow AccessDenied. Add a focused native Windows regression for regular-file/tree durability plus the existing complete boundary and subprocess suites.

2. Journal cleanup uses a non-portable synthetic mode digest. journal.go lines 796-809 calculates the expected journal digest with hard-coded mode 0600, while DigestPath hashes the actual Windows mode returned after Chmod. Native cleanup reports different actual and expected SHA-256 values, retains the journal, and causes the before-backup rollback and preparation-abort cleanup assertions to fail. Make journal removal verification use platform-stable persisted metadata semantics without weakening unknown-state refusal. Add a Windows regression proving a freshly saved canonical journal can be durably removed and that changed journal bytes remain preserved corruption.

## Rework gate

Keep scope within internal/transaction and preserve all Darwin/Linux behavior. After rework, rerun focused tests, race/vet/lint/build/compile gates, then independently cross-compile and execute the full internal/transaction suite natively on Windows, including boundary injection, subprocess crash recovery, rollback, cleanup, journal removal, and case handling. A passing cross-compile alone is not runtime evidence.