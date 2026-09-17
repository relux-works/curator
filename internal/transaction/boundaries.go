package transaction

// BoundaryCheck re-validates destination separation and physical identity
// for one journal target immediately before a publication write.
//
// Skillfile sources revision 1 section 2 requires this recheck before EACH
// publication write, under the serialized transaction — a single
// pre-journal check is not sufficient, because the engine commits targets
// one at a time and a boundary can change between writes. The staging
// layer owns the check itself (internal/staging); this package owns the
// enforcement point: a plan carries an in-memory BoundaryCheck, the
// engine retains it across Prepare and Commit, and commitTarget runs it
// immediately before every live mutation (the backup rename and the
// install rename of each target). A refusal fails the commit exactly like
// any other commit error: the journal moves to rolling back,
// already-published targets are restored from their backups, and
// unpublished targets are never touched.
//
// Arguments are the journal target index and that record's live path, so
// the check can verify it is rechecking the target the engine is about to
// write and fail closed on any ordering drift. An error refuses the write;
// its text is preserved verbatim (the staging check reports the
// skillfile-sources section 5 diagnostic class, or
// boundary_identity_unreadable for inspection failures).
//
// The in-memory check covers the same-process commit. Restart recovery
// re-verifies the durable proof instead (Plan.BoundaryProof, persisted as
// Journal.BoundaryProof): the same canonical spellings and filesystem
// identities in journal-stable form, checked immediately before each
// remaining write. A nil proof is a legacy unguarded journal and skips;
// a non-nil proof that cannot be restored fails closed with safe
// rollback, never legacy. A nil check with a non-nil proof (durable-only)
// still guards both paths.
type BoundaryCheck func(targetIndex int, livePath string) error
