# Snapshot metadata and identity

`envsnapshot` provides metadata identity, retained destination admission and
publication for managed-home-snapshot-v1. It does not repair, authorize a launch,
select a source closure, transport archives or persist a session baseline.
Source capture and materialize CLI integration are separate implementation leaves.

The first destination-admission unit adds `OpenDestination(path, policy)` on
macOS and Linux. It retains all supplied ancestor/target and protected-root
handles, checks owner/ACL evidence without writes, and refuses physical
equal/inside/containing protected roots. `DestinationPolicy.ProtectedRoots`
must be the coordinator's complete existing inventory (manager, stores,
managed/native homes, credential roots, repository administration, workspace).
This library does not discover that inventory. A missing or unreadable supplied
root refuses rather than becoming an empty inventory. `Revalidate` detects
changed retained entries and security evidence; `MatchPrior` rereads the actual
snapshot and compares exact manifest bytes, including the source binding.
Call `Close` to release every retained handle; it performs no cleanup writes.

Admission is an observation, not an atomic publication capability. No
materialize CLI calls it yet. Existing empty targets refuse with
`environment_materialize_destination_invalid`; only absent targets and verified
prior snapshots are admitted, following curator-spec revision
`15d2fc2be22c2b013faa8f637101947b1ff20830`.

`Destination.Publish(ctx, Publication)` freezes and validates captured bytes,
creates an unpredictable private sibling stage through the retained parent,
writes independent 0600 files (0700 only for captured execute metadata) and
0700 directories, and syncs files/directories before publication. macOS uses
`renameatx_np(RENAME_EXCL)`; Linux uses `renameat2(RENAME_NOREPLACE)` with no
fallback. An existing competing destination always refuses without overwrite.
The capture coordinator supplies authorized bytes, verified execute metadata
and the complete protected-root inventory; this library does not establish
source authorization from self-minted metadata.

Success returns `Verified` only after the published actual tree matches the
captured manifest (bytes, sizes, modes and exact entry set), and the retained
ancestor and target checks succeed. A post-publication mismatch or cancellation
leaves the published entry for the operator, with no destructive rollback.
Cleanup works backwards over recorded staging objects through retained handles,
never recursively enumerating a pathname or deleting the target. It refuses
observed identity substitutions and never follows links. As snapshot section 2
specifies, same-user mutation inside private staging is outside the guarantee:
an identity check cannot make a subsequent unlink conditional on an inode.
Context cancellation cleans unpublished staging; process death may retain a
private partial stage and never yields a success result. No crash recovery or
CLI success-output qualification is claimed.
The separate Windows backend is also outstanding: unsupported platforms refuse
with `environment_materialize_destination_invalid`, without filesystem writes.

Destination ancestors may belong to the caller or root and must not be writable
by group/other; the immediate parent and target must belong to the caller
and have mode 0700. Linux refuses extended access/default
POSIX ACLs, including unreadable or unsupported ACL evidence. macOS uses
`fgetattrlist` on retained descriptors and permits only absent/deny-only ACLs;
permit and unknown entries refuse conservatively. The direct macOS syscall has
no x/sys wrapper; its two narrowly named lint exceptions document the ABI and
pointer bounds. Real kernel ACL tests qualify the exercised runtime, not future
OS versions. Prior-snapshot admission supplements SNAP-ID with owner/ACL checks
on every actual member. It does not claim protection against subsequent writes
by another process with the same user's authority.

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
Capture consistency and source authorization remain the consuming leaves'
responsibility; destination policy requires their complete root inventory.

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
go test -count=1 -run '^Test(Destination|Publication)' -v ./internal/envsnapshot
go test -tags snapshotcontrols -count=1 -run '^TestDestinationNarrowingControls$' -v ./internal/envsnapshot
go build -o .temp/curator-snapshot-check ./cmd/curator
sh scripts/remote-gate.sh
```

The `snapshotcontrols` build tag includes the opt-in narrowing controls. Normal
package tests do not register an unexecuted/skipped control test. The controls create isolated package copies in `.temp/`, run
named behavioral tests with `go test -count=1 -json`, require a named test failure
rather than a build/setup error, and remove only their own scratch allocations.
They never mutate the original package. Each child failure is expected exit 1;
the control harness succeeds only when every planted weakening is killed.

The destination controls exercise the real `OpenDestination`, `Revalidate`,
`MatchPrior` and `Publish` library entry points in isolated package copies. They narrow linked
ancestor refusal (M2), saved-manifest byte trust (M6), owner/mode/ACL admission,
protected-root equality, ancestor/target identity and unrelated-target refusal.
The revised controls additionally cover empty-target refusal and private-parent
admission; the target-identity control replaces an intact prior snapshot with
an identical snapshot on a different inode. They do not claim the still-absent
`MaterializeCLI_*` production-entry tests.
Publication controls weaken no-replace to allow empty-directory replacement,
accept one post-verification byte mismatch, skip one recorded cleanup member,
or remove one foreign stage-looking sibling link. Kernel competition is
inserted after the last admission check, immediately before the real syscall.
Tests also exercise cancellation, a kernel permission failure, substituted
staging-root refusal and post-publication failure without rollback.
Destination syscall fixtures use private `.temp/destination-test-*` allocations
because shared temporary ancestors must refuse. The owner mutant requires an
unprivileged process. Windows tests qualify refusal only.

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
