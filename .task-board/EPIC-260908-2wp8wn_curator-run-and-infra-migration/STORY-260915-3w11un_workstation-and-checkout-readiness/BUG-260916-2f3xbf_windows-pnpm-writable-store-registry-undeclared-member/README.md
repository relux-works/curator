# BUG-260916-2f3xbf: windows-pnpm-writable-store-registry-undeclared-member

## Description
Discovered by TASK-260916-5aqozl once real pnpm ran on windows-latest for the first time (gate run 35098955988): internal/pnpmsource real-pnpm cases TestRealPinnedPNPMLockSupersetSnapshotDependencies and TestRealPinnedPNPMPrivateStoreAndOfflineMaterialization fail with closure_input_undeclared: pnpm writable store registry contains an undeclared member (entry:<hash>). On Windows pnpm 10.33.0 writes a store member the closure registry does not declare (store layout / hard-link fallback differs from POSIX). Not a CI problem: the writable-store closure model needs a Windows-correct declaration (or an explicit, tested refusal of pnpm sources on Windows). Until fixed, the two cases are declared Windows-deferred in the platform-case ledger by the pin task.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
Real-pnpm conformance cases run and pass on windows-latest without a ledger deferral; the writable store registry declares every member pnpm 10.33.0 writes on Windows, or pnpm sources are refused on Windows with a typed diagnostic and a test; negative test for an undeclared member remains.
