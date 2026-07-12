# Golden interoperability fixtures

The conformance gate of the implementation plan (P10). The `expected/`
values were produced once by an independent conforming implementation of the
protocol:

- `snapshot_sha256.txt`: the content hash of the raw `skill-fixture/` tree
  (Spec 8.5 byte layout).
- `context_files.json`, `context_sha256.txt`: the whitelisted context copy of
  the fixture (runtime roots excluded) and its hash.
- `marker.json`: a marker object with a normalized timestamp and commit,
  including the required empty-array and nullable-locale wire values.
- `registry/`: Ed25519-signed audit records and a snapshot produced with a
  fixed test key (seed bytes 0..31; the pinned public key sits next to them),
  plus a forged record that must fail verification.

`internal/interop` asserts byte equality and verification outcomes. CI runs
it as a dedicated job. Updating anything under `expected/` is a deliberate
protocol decision, never a test fix.

Regenerate from the repository root with:

```text
go run ./testdata/golden/regenerate -root .
```

The generator is a standalone protocol implementation and deliberately
imports no Curator packages. Always review the complete fixture diff.
