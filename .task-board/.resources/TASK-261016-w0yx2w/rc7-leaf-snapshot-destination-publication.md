# SNAP-PUBLISH — Retained-handle destination publication

Owner repository: **curator**. Size: **5**. Required in rc.7; qualify supported macOS, Linux and Windows paths separately.

Normative clauses: snapshots §2 in full; §§4/5 (actual-tree identity and refusal invariants); environments §5 path collisions. CIP-0014 `MaterializeCLI_DestinationBoundary` and `MaterializeCLI_Idempotence`.

Attack surface: destination traversal, replacement races and destructive idempotence.

Acceptance criteria:

- Validate destination without writes; retain handles across every ancestor and target. Refuse links/reparse points instead of canonicalizing through them, unsafe owner/ACL evidence and directories writable by other principals.
- Use handle identity/ancestry to exclude targets equal to, inside, or containing manager roots, stores, managed/native homes, credential roots, repository administration or the active workspace. Session allocation therefore needs a permitted private directory outside the workspace.
- Accept absent, empty caller-owned, or intact same-version prior snapshots only. Publish captured bytes by a platform-proven no-replace operation; recheck an empty target by handle before replacing its entry. A competing replacement cannot be silently overwritten.
- Verify prior snapshots using SNAP-ID, including actual bytes and extras. Same exact verified generation is a no-op; a different generation refuses and preserves it. No force/takeover/recursive-cleanup flag.
- Private directory/file modes are 0700/0600, executable assets 0700 only from verified execute metadata; no imported links, xattrs, ACLs or set-id bits. Failure emits no success and cleans only this invocation's retained staging objects.

Mandatory tests with positive controls:

| Positive control | Negative case / required invariant |
| --- | --- |
| Valid absent/empty private destination outside protected roots succeeds | Linked ancestor/target, reparse, wrong/unknown owner, broad ACL or protected identity alias refuses without outside writes |
| Stable retained parent and empty target publishes once | Replace an ancestor/target between validation and publication; create a competing destination; require refusal or the verified intended object, never traversal or overwrite |
| Intact identical snapshot is unchanged | Same manifest with changed bytes/mode, extras, unreadable entry, unknown version or a new source generation refuses and preserves existing target |
| Failure during staging leaves source and prior target unchanged | Cleanup link/substitution attack cannot remove native/managed/user files; interruption cannot leave a success manifest naming partial output |

Controls: final-component-only path checks (M2) and saved-manifest-only idempotence (M6). Real syscall/ACL/race tests are mandatory; schema vectors cannot prove this boundary.

Dependencies/paths: parallel with capture/codec in new `internal/envsnapshot` destination/publication platform files; joins SNAP-ID for existing-target verification. Do not reuse `internal/envprofile/nofollow.go` as sufficient: its `managedPath` explicitly allows components at/above root to resolve normally and returns a pathname after lstat checks. `internal/snapshot` is source-package acquisition, uses different modes/identity and includes pathname-based publication; it is not CIP-0014. Audit narrowly reusable platform primitives (e.g. buildcache retained-handle patterns), preserving their existing consumers.


## Owner decision 2026-10-17 (supersedes any empty-target clause above)

The destination is either absent, under an existing caller-owned private parent, or an intact prior snapshot of this version (verified no-op). An existing empty directory refuses. Publication stages beside the absent target and uses a no-replace rename (macOS `renamex_np` with `RENAME_EXCL`, Linux `renameat2` with `RENAME_NOREPLACE`); unqualified platforms keep refusing. A different generation refuses and needs a fresh absent destination. Follow curator-spec main after TASK-261017-uyuj08 lands.
