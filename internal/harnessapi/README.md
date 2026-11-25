# Harness contract v1

The harness seam migrates native operations one adapter at a time. Unported
operations continue through their legacy callers. The eight registered IDs
have real identity/version metadata. The five baseline release-probe methods
return real value-only specs; other probe requests return
`harness_capability_unsupported`.

The Gemini adapter also declares skill projection and read-only MCP
inspection. Other unported staged methods return
`harness_adapter_operation_not_ported`; that refusal does not claim that the
vendor lacks the capability. Capabilities describe only ported operations.
ParseReleaseProbe and vendor qualification remain unported, independently of
the adapter API version.

Context, Snapshot and plans have private fields and bounded constructors.
Snapshot currently accepts only closed nonsecret observations. A migration
requiring more observations must extend and test the collector/value contract;
raw native configuration, credential bytes and effect handles are not admitted.
Paths are lexical relative values. Their constructors do not establish
filesystem ownership or symlink safety; the shared transaction owner does.

The closed registry rejects unsupported registration versions before calling
their constructor. It verifies the adapter's version, ID and capabilities before
returning an adapter. Unknown IDs have no default fallback. CLI aliases and the
default canonical ID retain their existing meanings.

Go checks:

```sh
go test ./internal/harnessapi ./internal/harnessvalue
go test -count=1 ./internal/harnessboundary ./internal/harnessprobe
```

Use isolated HOME/TMP and Go caches. Store local evidence under ignored `.temp/`.
These checks qualify the bounded internal contract and source boundaries, not
live vendor releases or native operations still awaiting migration.
