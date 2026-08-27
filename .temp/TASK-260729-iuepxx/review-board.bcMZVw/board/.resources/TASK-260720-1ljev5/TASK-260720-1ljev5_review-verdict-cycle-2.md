# TASK-260720-1ljev5 review verdict — cycle 2

## Verdict: CHANGES REQUESTED

Route to to-dev. Rework cycle 1 closes the previously reported ordinary two-pass uncertainty cases, moves rename and recursive deletion through an os.Root bound to the validated cache-root object, and resolves the Windows evidence contradiction. Two P1 safety gaps remain.

## P1 — duplicate known fields bypass the strict consumer-registry gate

internal/scopes/consumers.go:116-146 decodes consumers.json into a struct with json.Decoder.DisallowUnknownFields. Go struct decoding still accepts duplicate known object fields and keeps the later value. Therefore a registry such as {schema_version:1, consumers:[live], consumers:[]} is accepted as trusted, Collect rewrites it empty, and a later pass no longer visits the live checkout whose marker protects a build. RecordConsumer and StageConsumer also accept and normalize the ambiguous registry. This recreates the exact pass-one-forgets/pass-two-sweeps failure that cycle 1 set out to close.

Required rework: reject duplicate schema_version and consumers members, preferably by exact canonical-byte validation or duplicate-aware token parsing; reject any other noncanonical ambiguity before any writer normalizes it. Add a two-consecutive-pass regression with a duplicate consumers field proving Collect never calls the build sweep and preserves the original bytes on both passes. Add matching RecordConsumer and StageConsumer fail-closed assertions and an expected-red negative control against the current parser.

## P1 — decisive entry classification reopens the mutable Unix pathname

internal/buildcache/collect.go:192-217 opens and parent-binds the candidate, stats it, and reads its receipt through the proven entry handle. It then calls store.inspectEntry(entryPath, ...). internal/buildcache/cache.go:125-170 calls openProtectedEntry with that pathname, and internal/buildcache/protection_unix.go:18-61 starts a fresh traversal from the manager-home pathname. On Unix, the already-open original handles do not prevent a cache-root exchange. A swap after the parent assertion and receipt read but before the second traversal can make the exact-content and artifact-hash checks validate a replacement entry; retireEntry then removes the different entry in the original proven root through root.mutator. A corrupt or structurally unexpected original candidate can therefore be deleted without ever passing the decisive classification.

The current TestSweepRemovalSurvivesACacheRootExchangedMidPass hook runs only after inspectUnexpected returns, so it proves mutation binding but cannot exercise classification binding. Windows prevents the exchange with handles opened without FILE_SHARE_DELETE; the unresolved Unix window still violates the Unix and race acceptance criteria.

Required rework: perform receipt decode, exact entry-content validation, canonical receipt validation, artifact open/hash/size validation, and publication-time lookup from the already-proven candidate/root object. Do not reopen the cache entry through store.home or root.path during sweep classification. Add a deterministic exchange seam before the decisive full classification and a negative control proving a replacement-valid/original-invalid pair cannot cause retirement of the original.

## Independent validation

PASS: go test ./internal/scopes ./internal/buildcache ./cmd/curator -count=1.
PASS: go test -race ./internal/scopes ./internal/buildcache ./cmd/curator.
PASS: gofmt -l on scoped packages, git diff --check, and go vet ./....
PASS: go test ./... -count=1, including internal/install in 276.866s and internal/install/atomicity in 399.048s.
Reviewed cycle-2 evidence also shows native Windows scopes and curator GC green, with buildcache reds isolated to inherited publication/DACL cases.

No product code was modified during review. The two findings are ordinary implementation rework, not a stop-the-line boundary.