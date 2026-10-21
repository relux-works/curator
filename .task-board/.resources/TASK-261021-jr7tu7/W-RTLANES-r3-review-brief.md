# Delta review: TASK-261021-jr7tu7 rework r3, covered lanes (curator-spec)

Your r1 finding R1 (frozen transport prose and the accepted manifest changed) is addressed at the new head in the request (branch `repository-transport-lanes`).

1. Verify, against `origin/main`, that `protocol/repository-transport.md`, the accepted source manifest, the candidate pin and the frozen inventory (`TASK-261021-jr7tu7_frozen-inventory-r3.json`) are byte-identical. Recompute the frozen set yourself; don't trust the inventory alone.
2. Rerun your R1 attack: the production `release_gate.validate_skillfile_sources_manifest` passes at the head, and the full normal-.git Python suite (879 tests, including the 40 release-gate tests) passes.
3. The seven lane-side back-links: accurate restatements of repository-transport §1–§3, nothing normative added, and none of them in a frozen file.
4. Validator, links, prose rules.
