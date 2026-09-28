# Addendum (operator decision, 2026-09-24): lock replay on a fresh machine — binding

The accepted revision of `skillfile-sources.md` §3 changes one rule. Today installation with an existing lock fails with
`source_snapshot_unavailable` whenever the local snapshot store lacks the locked snapshot. That happens on every fresh machine, for
every source kind, so a committed lock never installs there.
In the accepted text, when the snapshot is missing, installation re-materializes it from the declared source:
- `git` and `repository`: fetch exactly the locked revision.
- `path`: read the current bytes.
It accepts the result only when the package identity and `content_sha256` equal the lock. A mismatch fails with
`source_snapshot_changed`. `source_snapshot_unavailable` remains for a source that cannot be reached at all. The lock stays
byte-identical, and no tag or branch is re-resolved.
The operator also decided that `Skillfile.lock.json` is committed with `Skillfile.json`, like `package-lock.json`. Curator implements
the same replay rule when it flips schema 2 to default-on.
