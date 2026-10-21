# Rework brief r5: repair rebuilds a damaged projection (BUG-261021-2qv4zj, curator)

Review r4 (`BUG-261021-2qv4zj_review-r4.md`, plus `BUG-261021-2qv4zj_projection-repair-r4_test.go` as outcome resources) confirms that F1, F2 and F3 stay closed, and requests changes for one narrow P1, F4. Run the attached probe first. Continue on branch `fix/audit-markdown-scope`.

## F4

An existing projection directory with a missing, extra or changed file is correctly diagnosed as stale. But `Resolve(Repair:true)` then refuses permanently (`environment_repair_failed: skill context projection differs from its identity`) instead of rebuilding it. Six cases across Claude and Codex fail. The intact and missing-directory controls pass.

## Required

Keep the change narrow. It is not a redesign.
- Under repair, a projection whose bytes differ from its versioned identity is rebuilt from the immutable store entry. The swap is atomic: stage the new projection, verify its identity, then replace the old one within the existing recoverable transaction. Readers and links never see a half-written tree.
- Plain resolve and status stay read-only and keep reporting stale.
- The rebuild is bounded: if the freshly staged projection still does not match its identity (for example, the store entry itself is damaged), refuse with a typed error and leave the old directory untouched. Never loop.
- A projection shared by two profiles is rebuilt once, and both profiles converge.

## Tests (each states its claim)

- The six r4 cases (missing, extra and changed file × Claude and Codex) repair and converge, and the bytes then equal the identity.
- A damaged store entry produces the typed refusal with no partial replacement.
- Mutant: dropping the rebuild branch fails the six cases.
- All prior regressions stay green.

## Run and hand-off

Run build, vet, focused tests, and the full default suite (private GOCACHE, deleted afterwards). Attach `BUG-261021-2qv4zj_results-r5.md`, push signed, hand over in `audit-md-scope`, and delete scratch. Use neutral QA wording.
