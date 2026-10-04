# BUG-261004-2v9pbz: expanded-snapshot-budget-bypass-repeated-blobs

## Description
Report: docs/security-audit-2026-10-inline.md, finding N2 (priority: first). internal/buildrepo/admission.go objectReader.read returns a cached object before budget accounting, walkTree copies a blob per path, and frameSnapshot re-assembles the whole expanded snapshot in memory; there is no aggregate budget. With MaxExpandedBytes=16384, one 4096-byte blob referenced by 64 paths is ADMITTED (262161 content bytes, 263922 canonical bytes) via AdmitLocal, while 64 distinct blobs are refused. Local and network admission share proveRepository. Contract: manager profile §11.5 (rc.13+) bounds aggregate expanded bytes and memory.

## Scope
internal/buildrepo (admission.go, local.go), regressions through AdmitLocal and the network admission entry

## Acceptance Criteria
1. Aggregate accounting of emitted files, expanded bytes and canonical framing size, checked BEFORE each copy/append; unique-object limits are kept.
2. A bound on visited expanded tree entries; context cancellation is checked during the walk.
3. Red-first regressions through AdmitLocal: one blob at 64 paths over budget is refused with a stable incomplete-source code; the one-blob and 64-distinct-blob controls keep their current outcomes; a repeated-tree (tree DAG) fan-out case is refused.
4. Evidence that the network admission path enforces the same bound (shared proveRepository row or a network-entry regression).
5. No change to admitted canonical snapshot bytes for in-budget repositories (rc.14 conformance stays green).
