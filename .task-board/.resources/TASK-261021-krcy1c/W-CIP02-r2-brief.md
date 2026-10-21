# Rework round 2: TASK-261021-krcy1c, CIP-0002 candidate (curator-spec)

The review (`TASK-261021-krcy1c_review-r1.md`) requested changes: one P1.

**F1** (cips/CIP-0002-project-context-in-managed-launches.md:312): the closed slice binds a context snapshot inventory, but it defines no context entry or module manifest and no deterministic loading semantics inside an item. Two Markdown files under identical approval bytes therefore allow incompatible root selections. Only the skill arm inherits an existing package contract.

## Fix

- Prefer **inheritance over new grammar**. A project context item should be a snapshot of a **context package** under the existing context-package manifest contract. Look at environments §1 (context packages: root context, modules, weights and the like), the context manifest schema (schema 2, as in CIP-0021) and the lock rules. Then state that the item inherits its validation, root and module selection, ordering, weights and rendering, exactly as the skill arm inherits the skill package contract.
- If an inherited contract truly does not fit a bare repository instruction file (for example a single AGENTS.md with no manifest), define the **minimal subset** explicitly: a single-file item with one named entry and a fixed role, or an entry list with roles (entry or support), a deterministic order, byte validation and import closure.
- Bind entry selection into the approval digest, so that changing the root selection means re-approval.
- Add positive and negative vector families that discriminate: two Markdown files with a different root selection must produce different approval digests and different materialization. Also cover a missing entry, an entry outside the snapshot, an import escaping the closure, and duplicate roles.
- Keep everything else unchanged. Rerun `make validate` in a normal-.git clone.

## Hand-off

Push a **new commit** on the same branch `cip02-ready-for-acceptance` (no force-push), signed. Write results r2 with a delta summary, attach them, and hand over with the new head.
