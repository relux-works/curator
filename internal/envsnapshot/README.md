# Snapshot metadata and identity

`envsnapshot` is the read-only library boundary for managed-home-snapshot-v1.
It does not materialize, repair, authorize a launch, select a source closure,
transport archives, or persist a new baseline. Capture, publication and hosted
consumers must call it; those integrations are separate implementation leaves.

- `EncodeManifest` / `DecodeManifest` and `EncodeDescriptor` / `DecodeDescriptor`
  enforce the embedded, closed schemas and CCJ-1 plus exactly one LF. Decoding
  metadata alone never verifies an existing snapshot directory.
- `FragmentDigest` takes source CCJ-1 without presentation LF. `Digest` hashes
  exact bytes; metadata identities include their LF.
- `Verify(root, originalManifestSHA256)` validates metadata, channel semantics,
  the complete actual tree, independent regular files, bytes, sizes and private
  semantic modes. Pass an empty expected digest only to inspect an unadmitted
  candidate. Never replace an admitted expected digest after a mismatch.
- `PortableEqual(left, right)` verifies both trees anew and compares every
  manifest member except `fragment_sha256`, retaining producer, adapter,
  profile lock, descriptor, files and modes. It is not exact plan identity.

`Verified` returns copied metadata bytes. Its observation is not a continuing
lease against hostile concurrent writes. Callers retain their own immutable
baseline and any overlay identity, and reverify immediately before consumption.
Capture consistency, destination ownership/ACL policy and publication locking
remain the consuming leaves' responsibility.

Filesystem verification uses retained no-follow directory handles on Unix.
It rejects linked ancestors as supplied, including links preceding `..`,
symlinks, hard links, special files, missing/extra directories and file drift.
Other platforms currently refuse with the snapshot-mismatch code: no Windows
ACL or execute-mode equivalence is claimed. Codecs remain platform independent.
No runtime, archive, or materialize CLI qualification is claimed here.

The unchanged separate corpus is pinned under
`internal/crossconformance/testdata/managed-home-snapshot-v1/`. The two snapshot
schemas and their five transitive schema dependencies are embedded here,
byte-for-byte from curator-spec revision
`d2a942cf17e4eb017c45475442da1456d8dfbe6c`. Released fragment schemas are unchanged.
`TestCorpusPins` verifies 95 referenced corpus/schema files; the remaining
normative protocol document is a source reference, not an executable fixture.
The schema compiler uses the existing regexp2 dependency for the published
ECMAScript patterns and refuses unbundled resources without network access.

## Tools and verification

Run from the repository root. Temporary outputs belong under `.temp/`.

```sh
go test -count=1 -race -coverprofile=.temp/snapshot-coverage.out ./internal/envsnapshot
CURATOR_SNAPSHOT_CONTROLS=1 go test -count=1 -run '^TestNarrowingControls$' -v ./internal/envsnapshot
golangci-lint run ./internal/envsnapshot/...
go build -o .temp/curator-snapshot-check ./cmd/curator
sh scripts/remote-gate.sh
```

The opt-in narrowing controls create isolated package copies in `.temp/`, run
named behavioral tests with `go test -count=1 -json`, require a named test failure
rather than a build/setup error, and remove only their own scratch allocations.
They never mutate the original package. Each child failure is expected exit 1;
the control harness succeeds only when every planted weakening is killed.
