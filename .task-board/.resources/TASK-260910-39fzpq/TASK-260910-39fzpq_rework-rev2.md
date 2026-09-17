# Rework brief — TASK-260910-39fzpq, revision 2 (S5 protected boundary)

Revision 1 was rejected with four corrections
(`TASK-260910-39fzpq_review-verdict-rev1.md`, F1–F4). The orchestrator settles
the design points below; everything the verdict table marks as passing stays
byte-identical unless a correction touches it.

## Settled design (implement exactly this)
- **Store integrity is verified against the pin, not the marker.** The
  integrity baseline of a store entry is its pin identity: for a `git` member
  the entry's tree MUST hash to the resolved commit's tree (the manager
  recomputes the tree hash of the entry from its bytes — the git tree object
  identity of the pinned commit, or the recorded snapshot tree hash the lock
  carries), for a `path` member the `state_sha256`. This check needs no home
  marker and runs at every `env resolve` and before provisioning/repair of
  any home (F2: an unprovisioned home is verified the same way; missing
  hashes never count as passed; absent vs unreadable/malformed marker are
  distinct facts — §8.4 — with the unreadable/malformed one being
  `environment_marker_unreadable`-class non-current, never "absent").
- **Home currency is a separate check.** The marker's recorded surface
  hashes are compared with the surfaces the CURRENT lock would generate
  (§5.1/§5.6); a mismatch because the marker belongs to another lock (stale
  after `profile update`, §9.2) is the ordinary stale-home condition
  (`environment_home_stale`) repaired from the verified store — never
  `environment_store_untrusted` (F1). Add the two cases the reviewer names:
  intact updated store + old marker → stale, repair succeeds; swapped
  updated store + old marker → untrusted, never adopted.
- **Two failure classes with an ordering (F3).** (a) An enclosing boundary
  that cannot be proven — the environments root or the store root (wrong
  owner, world/group-writable, symlinked, not a directory, not contained) —
  refuses every mutating operation before its first write and every resolve
  (no fragment); nothing is rebuilt, because there is no protected place to
  rebuild into; posture row names the boundary; the operator repairs the
  boundary out of band. (b) An individual entry (store entry, lock file,
  marker) that fails its own checks or its pin hash inside a proven enclosing
  boundary is untrusted; a real operation rebuilds it from the revalidated
  snapshot into newly established protected state (operation-private
  staging, atomic publication under the mutation lock), dry-run reports
  `would-rebuild-untrusted-store`, resolve fails closed until rebuilt. State
  the order: enclosing boundary → entries → pin hashes → home currency.
  Reconcile the dry-run and GC wording with this split.
- **Validator pins scenarios (F4).** The gate binds each named case to its
  discriminating inputs (which object — environments root, store root,
  entry, lock, marker — and which check — ownership, permissions,
  containment, regular type, link safety, pin hash, home currency — fails)
  and refuses a corpus where a named branch no longer exercises its check
  (the reviewer's five-to-one narrowing); add that narrowing as a negative
  test in `tools/test_validate.py`; add environments-root and store-root
  cases (the two enclosing-boundary classes) and the entry-rebuild case.

## Validation and handoff
`make validate` and the regeneration proof (exit codes quoted); evidence
"Revision 2" section; `TASK-260910-39fzpq_spec-patch_rev2.patch` =
`git diff HEAD` of the worktree (base `684c9f1`) with new files via
`git add -N`; `task-board handoff TASK-260910-39fzpq --role doc-writer`.
Worktree and rules unchanged (the rules file gained rule 7 — read it).
