# TASK-260910-2vnjej — Windows gate fix (THE ONLY CURRENT INSTRUCTION, with 2vnjej-sec-brief.md)

Rev1 (tree ac31ab6d, base 213a53e5) is green on linux/macOS but Test (windows-latest) fails 8 tests (run 36446388699, go-test.json of
test-evidence-windows-latest): internal/registry TestReconcileBootstrapCheckpointPreservesAndAdvancesHighWater/{lower_refused,
equal_conflicting_refused, higher_advances} ("reconcile = state {…BootstrapSource:first-use BootstrapCheckpointID: …}, applied=false
regression=false error=<nil>") and internal/install TestRegistryCheckpointPersistsBeforeTamperedNetworkSnapshot,
TestRegistryInvalidFirstUseCheckpointFailsClosed, TestRegistryRebootstrapRegressionKeepsExistingHighWater, TestRegistryBootstrapVectorsDriveInstall.
On Windows the bootstrap checkpoint is never recognised (no checkpoint id, applied=false, no error). Find the real cause with evidence —
likely candidates: CRLF in a signed checkpoint/testdata file on the Windows checkout (signature or note parsing mismatch → silently
treated as absent; add a .gitattributes `-text`/`eol=lf` rule for that data or parse byte-exactly), a path built with "/" vs filepath, or a
read that maps a Windows error to absence. Two fixes are required: (1) the root cause; (2) a silent `applied=false, error=nil` for a
present-but-unparseable checkpoint must become a fail-closed error (a present checkpoint that cannot be verified is never "absent" —
§8.4.1 discipline, stateread). Verify with `GOOS=windows go vet ./internal/registry ./internal/install` and a unit row that feeds CRLF
content. Set status development; update results with `git diff ac31ab6d` non-empty; handoff; END YOUR TURN. No CHANGELOG/LOGBOOK edit.
