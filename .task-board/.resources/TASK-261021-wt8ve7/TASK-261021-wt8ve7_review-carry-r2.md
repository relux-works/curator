# Review carry-forward: TASK-261021-wt8ve7 r2

The accepted review is TASK-261021-wt8ve7_review-r1.md by ivmbp_curator_rev2, at 9f66334b48b4765f3011d11f0e1170d678cc4b40, with one non-blocking P2 (R1). The republished head 0c5c02c8be22878e37f1565318024aa94d261395 (branch research/project-source-exclusion-claude-r2) is two signed commits onto main 2faaa544. The orchestrator verified both independently:
- 64e93191, the research commit: all 12 non-README paths are byte-identical to the accepted 9f66334b. README.md keeps both the Codex and the Claude entries (a mechanical merge);
- 0c5c02c8 changes one line of the record: the child mode is async/background, and foreground children are explicitly unmeasured. That is exactly the change R1 asked for.
Both commits are signature G and fast-forward main. The producer reran build, vet and the default suite at the head.

```json
{"verdict":"accepted","head":"0c5c02c8be22878e37f1565318024aa94d261395","findings":[],"basis":"carry-forward of rev2 r1; mechanical README merge plus the requested P2 label fix, verified by orchestrator"}
```
