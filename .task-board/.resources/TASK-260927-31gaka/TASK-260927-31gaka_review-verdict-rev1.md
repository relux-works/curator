# TASK-260927-31gaka review verdict — rev1: CHANGES REQUESTED

Reviewed CR-TASK-260927-31gaka-1 rev1, base 86552087, tree ed98b5b1 (`git diff 86552087 ed98b5b1`).

## Blocking finding F1 — stray compiled binary in the candidate
`curator` (repo root, mode 100755, blob 8034b01f, 20,249,680 bytes, `Mach-O 64-bit executable x86_64`) is part of the candidate tree.
It is a locally built CLI binary, not product/test/docs/.github/ci content. Violates the brief ("Write only inside your Story worktree"
+ campaign rule "git status --short contains only product, test, docs, CHANGELOG and .github/ci") and review note item 4 ("no stray files").
Landing it would put a 20 MB host binary on main.
Fix: `rm curator` in the Story worktree (keep it out of the tree; build to $TMPDIR), re-verify `git diff --name-only origin/main` lists the 7 leaf paths, republish.

## Non-blocking observations on the code (for rev2; no change required)
- internal/envregistry/envregistry.go: single switch `CodexSeedRevision = CodexSeedRevisionA`, B kept — matches brief item 2.
- internal/envprofile/managed.go gatherSeeds: A path parses names and keeps payload whole; warning `mcp_native_servers_ungoverned` with
  "next seed revision stops inheriting them" + migration hint; B path unchanged behind switch; test seam `codexSeedRevisionForTest` validated.
- status.go: shipped revision drives posture; mcp_seed_unstripped for A records only when shipped=B — consistent with §7.4.
- conformance-gaps.tsv: 9 B-shipped rows owned by TASK-260927-1e5qqm.
Mutant/test reruns not performed this round because F1 alone invalidates the candidate; the next review will rerun them on the rev2 tree.
