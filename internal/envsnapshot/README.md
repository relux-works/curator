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
Path and component collision keys use canonical caseless matching: NFD before
case folding and NFD again afterwards. Folding alone can introduce canonically
equivalent Greek spellings with different byte sequences.
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
Local `.gitattributes` disables EOL conversion for every embedded schema asset;
published README bytes are pinned just like JSON. The repository's existing
testdata attributes protect the separate corpus.

## Tools and verification

Run from the repository root. Temporary outputs belong under `.temp/`.

```sh
go test -count=1 -race -coverprofile=.temp/snapshot-coverage.out ./internal/envsnapshot
go test -count=1 -run '^TestPinnedSchemaCheckoutBytes$' -v ./internal/envsnapshot
go test -tags snapshotcontrols -count=1 -run '^TestNarrowingControls$' -v ./internal/envsnapshot
golangci-lint run ./internal/envsnapshot/...
go build -o .temp/curator-snapshot-check ./cmd/curator
sh scripts/remote-gate.sh
```

The `snapshotcontrols` build tag includes the opt-in narrowing controls. Normal
package tests do not register an unexecuted/skipped control test. The controls create isolated package copies in `.temp/`, run
named behavioral tests with `go test -count=1 -json`, require a named test failure
rather than a build/setup error, and remove only their own scratch allocations.
They never mutate the original package. Each child failure is expected exit 1;
the control harness succeeds only when every planted weakening is killed.

The retained `TestReviewerGreekCollision` regression exercises `Verify` with
iota and upsilon spellings as files and directory components, singleton positive
controls and unchanged-manifest assertions. M12 restores the former
normalization-before-fold-only algorithm while retaining collision checks; that
named test must fail. To reproduce successful admission of colliding entries by
the mutant, run with `TMPDIR` on a case-sensitive filesystem. On folding
filesystems the test still requires the precise `path-collision` refusal before
tree enumeration, but cannot establish that both physical names coexist.

Git is also required by `TestPinnedSchemaCheckoutBytes`. That test runs
`git hash-object`, `git update-index` and `git checkout-index` in isolated
temporary repositories with `core.autocrlf=true`, without commits or source
writes. An unprotected control must convert LF to CRLF; protected checkout
must retain the independently published digest. The M11 control enables text
conversion only for schema Markdown and must fail that named behavioral test.
