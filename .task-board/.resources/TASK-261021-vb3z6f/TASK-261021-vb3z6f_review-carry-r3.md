# Review carry-forward: TASK-261021-vb3z6f, rebase onto main 9702382b

The accepted review is `TASK-261021-vb3z6f_review-r2.md` by ivmbp_curator_rev3, at head cf1241cbb828180e5308a8b02f2671ebedabcedd.

The republished head 76784377893ab578489f8ba098763b5937ed04fa (branch `exact-plan-final-form-r3`) is two signed cherry-picks onto main 9702382b. The orchestrator verified it independently:
- `cips/CIP-0023-exact-plan-hosted-launches.md` has the same blob as at the accepted head (byte-identical);
- the only other changes are index entries in CHANGELOG.md, README.md, UNRESOLVED_QUESTIONS.md and cips/README.md, with the existing CIP-0002 and CIP-0022 entries kept;
- the head fast-forwards main, and both commits are signature G. The producer reran the full validation at the head.

```json
{"verdict":"accepted","head":"76784377893ab578489f8ba098763b5937ed04fa","findings":[],"basis":"carry-forward of rev3 r2 at cf1241cb; CIP blob identical, index-only rebase delta verified by orchestrator"}
```
