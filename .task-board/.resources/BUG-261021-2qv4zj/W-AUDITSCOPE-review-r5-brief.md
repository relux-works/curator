# Review brief r5: repair rebuilds a damaged projection (BUG-261021-2qv4zj, curator)

This is a delta re-review at the exact head in the request, `e32d101e`. Review it against:
- `W-AUDITSCOPE-rework-r5-brief.md`;
- `BUG-261021-2qv4zj_review-r4.md` and its probe `BUG-261021-2qv4zj_projection-repair-r4_test.go`;
- the producer's `BUG-261021-2qv4zj_results-r5.md`.

Review r4 accepted everything at `a2244ec1` except one narrow P1, F4. An existing projection with a missing, extra or changed file was diagnosed stale, but repair refused instead of rebuilding. Focus on `a2244ec1..e32d101e`. Also confirm the r4 surfaces still hold at the new head.

| Surface | Check |
| --- | --- |
| F4 closure | Run the r4 probe unchanged. All six cases repair and converge for Claude and Codex: a missing, an extra and a changed file. Afterwards the bytes equal the versioned identity. The intact and missing-directory controls still pass. |
| Rebuild safety | The new projection is staged and verified before it replaces the old one inside the recoverable transaction. No reader or link sees a half-written tree. Test a crash or failure injected between staging and replace: recovery leaves either the old or the new tree, never a mix. |
| Bounded refusal | If the store entry itself is damaged, so the staged rebuild does not match, repair returns the typed refusal and leaves the old directory untouched, with no loop or retry storm. |
| Read-only paths | Plain resolve and status stay read-only: they report stale and never rebuild. |
| Shared projection | Two profiles that share a member converge after one rebuild. |
| Regressions and suite | F1–F3 stay closed. Re-run your choice of 2–3 narrowing controls, including omitted-rebuild. Run the full default suite at the exact head with a private GOCACHE, and delete scratch afterwards. |

Give the verdict at the exact head. Attach `BUG-261021-2qv4zj_review-r5.md` as an outcome resource, with the JSON verdict block at the end. Use neutral QA wording.
