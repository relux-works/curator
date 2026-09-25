# TASK-260924-10d3l1 review verdict — CR rev3: CHANGES REQUESTED

## Finding F1 (blocking): candidate tree is stale; it reverts 60 trunk files
- The CR base is ab34556e. The candidate tree 430c6fea equals a48f584c plus the test change. a48f584c is an ANCESTOR of ab34556e (it is older).
- `git diff --name-only a48f584c 430c6fea` outside .task-board: internal/install/draftevidence_test.go only.
- `git diff ab34556e 430c6fea -- . ':!.task-board'`: 61 files, +1030/-2185. It reverts the work of five landed stories between a48f584c and ab34556e:
  - 2tyzhh (002273e9)
  - 3eywt2 (5dddbb57)
  - txgta4 (f03da5bc)
  - 2mla0q (07878da8)
  - 3qwrnl (16eec15e)
- This includes production files: internal/install/{install,global,draftsources}.go, internal/marker/marker.go, internal/manifest, internal/envprofile, internal/gitops, internal/scriptworker/exec.go, the CI scripts and ci.yml, and the docs.
- If integrated, it silently rolls back trunk. The rev3 validation log is green, but it validated the stale tree.

## What is fine
- Per-file `git patch-id --stable` for draftevidence_test.go is c7ce917d… in both rev2 and rev3. That is the content accepted in rev1 and rev2.
- CHANGELOG.md is absent. The patch sha256 matches the resource (ff1b9422…).

## Required rework
- Re-carry ONLY internal/install/draftevidence_test.go (patch-id c7ce917d…) onto the current base ab34556e (or current trunk).
- The CR delta against its base must be exactly that 1 path.
- Re-run the focused validation `go test ./internal/install -run '^TestDraftEvidenceExactMatch$'` on the carried tree.
