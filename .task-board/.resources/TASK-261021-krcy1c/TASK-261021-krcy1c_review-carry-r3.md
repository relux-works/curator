# Review carry-forward: TASK-261021-krcy1c, rebase onto main c8b62454

The accepted review is `TASK-261021-krcy1c_review-r2.md` by ivmbp_curator_rev3, at head f202f41ba4718b82cdfefac28c8dfaa9af3a7de5.

The republished head 9702382be175e2297e943f1ea97cc6e6bd8dda8f (branch `cip02-ready-for-acceptance-r3`) is two signed cherry-picks onto main c8b62454. The orchestrator diffed it independently against the accepted head:
- in `cips/CIP-0002-...md`, one hunk replaces the prose reference to the companion draft's branch with a link to the now-landed `CIP-0022-mcp-server-declarations-in-project-skillfile.md`. The orchestrator authorized this in the rebase request, and nothing else in the CIP changes;
- `UNRESOLVED_QUESTIONS.md`: a mechanical append conflict with CIP-0022's entry, resolved by keeping both entries;
- the head fast-forwards main, and both commits are signature G. The producer reran the full validation at the head: 879 Python tests including the 40 release-gate tests, 74 schemas, 1294 vectors, Go tools, independence gate.

Acceptance carries forward; the delta is mechanical and authorized.

```json
{"verdict":"accepted","head":"9702382be175e2297e943f1ea97cc6e6bd8dda8f","findings":[],"basis":"carry-forward of rev3 r2 at f202f41b; mechanical rebase delta verified by orchestrator"}
```
