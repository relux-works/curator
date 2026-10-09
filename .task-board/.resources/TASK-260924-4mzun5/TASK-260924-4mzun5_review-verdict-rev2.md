# TASK-260924-4mzun5 — record-dependency-directory-in-legacy-lane: revision 2 review

Verdict: **changes_requested**. Reject revision 2 and route to `to-dev` once for the complete finding set. No code, commit, acknowledgement, or integration changes were made to the Story workspace. This run is not goal-bound.

Reviewed base `3b6481c15d61b08d3f0fae7c3269329555336709` and candidate tree `9e2ff0961f1cdd424003f94fbb308d3e16b8ab49` (52 changed paths). Fresh origin/main advertised and fetched that base. Review modifications and verification occurred only in a disposable clone.

## Spec decision and prior findings

RECORD remains correct. Curator-spec `7eaeb73fcf22cb8ce8e21325a8237accdfaf9646`, [skillfile-sources §4](https://github.com/relux-works/curator-spec/blob/7eaeb73fcf22cb8ce8e21325a8237accdfaf9646/protocol/skillfile-sources.md#4-runtime-builds-audit-and-persistent-records) explicitly requires marker v6 for schema-9 installations under a schema-1 Skillfile. [Core §4.4](https://github.com/relux-works/curator-spec/blob/7eaeb73fcf22cb8ce8e21325a8237accdfaf9646/protocol/core.md#44-dependencies) defines normalized directory/package identity; the source contract defines complete package audit-cache equality and receipt-3 records. Core §10 requires changed declarations to make an installation non-current and limits strict-tag refusal to a moved tag. The v6 carrier retains those currentness rules. Skillfile-sources §4 further requires status to compare package and lock against the effective plan and explicitly states that a changed declared ref is non-current.

- Revision-1 F1: fixed within checked scope. Project/global source-v1 runtime tests passed; project shim launch, repeat-install currentness and GC retention passed. The project-only runtime mutant fails the global production row.
- Revision-1 F2: the invented per-node preimage is removed. The legacy plan now uses a validated full-closure sourcelock.New digest with a CCJ-1 declaring-manifest digest. Production tests compare all migrated markers against a reconstructed lock; my manifest-only edit probe changes their digest. This review does not infer that schema-1 re-resolution becomes a schema-2 frozen-lock lane, nor require that migration.
- Revision-1 F3: fixed within checked scope. Real CLI schema-9 compilation, launch, reinstall and checking status passed on hosted macOS. Permanent project production tests assert receipt-3 input.package equality and repair a tampered marker key; my global install probe records receipt 3. External build arms were not independently attacked.
- Revision-1 F4: actual movement of the same tag now refuses under strict policy, but the added comparator falsely refuses an explicit change to another unchanged tag. This is a new mechanism in the same moved-tag class, so the finding's repeat-of is none rather than rev1/F4 (which was the blanket package-marker skip). The next rework must carry a named regression and narrowing mutant for this repeated class, inside this leaf.
- Revision-1 F5: fixed within checked scope. Install/CLI records bind the typed full package; same identity hits and repository/directory/commit changes miss. My corrected real CLI repository-cache probe kills the kind/commit-only comparison mutant.

## Blocking findings

**F4 — valid explicit tag changes fail under strict-tag policy (regression).** `internal/install/install.go:1519` compares the package commit to any currently declared tag without proving that it is the same tag previously installed. Install schema-9 consumer at v1; create v2 at a second commit without moving v1; edit the Skillfile to v2. Project and global install both fail with `moved tag for consumer: v2 old -> new`. Core §10's moved-tag policy and the existing legacy comparison distinguish changing the declaration from moving the same tag. Recover/bind the previous declaration/ref using validated evidence, keep actual same-tag movement detection, and retain schema-1 semantics. Add `TestReviewChangedDeclaredTagIsNotMovedTag` as a permanent project/global production regression with a mutant narrowing detection to commit-only evidence. This is ordinary implementation rework, not a human decision blocker.

**F6-declared-ref-currentness — status admits a stale declared-ref binding (bypass).** `cmd/curator/main.go:1350` reports package markers current on commit equality alone. Install at v1; create an unchanged v2 tag at the same commit; change the Skillfile to v2. Real `status --json` says consumer is up-to-date and `status --check` exits 0. The marker's manifest/lock digest would change on reinstall; status currently ignores it. Preserve read-only status and compare the complete applicable package/declaration binding (including ref identity). Add `TestReviewCLIStatusDetectsChangedTagAtSameCommit` permanently with a commit-only narrowing mutant. This new mechanism has repeat-of none.

## Swept surfaces

| Surface | Result/evidence |
| --- | --- |
| Spec decision and v6 carrier | Correct RECORD decision; exact-tree full gate and marker source review |
| Project/global runtime and GC | Production rows green; runtime-project-only mutant killed |
| Effective lock and manifest binding | Shared canonical full-closure digest; own manifest-only production probe green |
| Compiled providers and receipt-3 records | Real hosted macOS CLI compile/launch/currentness green; project tamper repair and global receipt row green |
| Legacy moved tags and repeat installs | Actual moved-tag control green; changed-declaration project/global probes fail (F4) |
| CLI status/currentness | Current installation control green; changed-ref same-commit probe fails (F6) |
| Audit install/CLI/sourceaudit/cache | Full package rows green; corrected CLI mismatch probe kills narrowed comparison |
| Dependency directory glob grammar | ?/[ production diagnostics pinned; star-only mutant killed by both dependency rows |
| Vendored v1/v2 corpus preservation | v1 zero delta; 24/24 v2 JSON bytes match pinned spec |
| Generation reads and atomic publication | Exact-tree full gate reused; inspected single-read payload plumbing; no new concurrency attack |

## Verification identities and bounds

Accepted already-attached evidence: [full hosted gate 37842134668](https://github.com/relux-works/curator/actions/runs/37842134668), success. Its head `0cdc94a5c1a1e31bd2ee99848f3d099cf78aeb5d` has the exact CR tree, verified from GitHub commit metadata. Lint/test/race matrix passed. Candidate-suite and self-hosted rose-air lanes were skipped, not proven. I did not replay the full gate.

Reran myself: baseline `go build ./...` and scoped `go vet` for install, CLI, audit, skillspec and marker, exit 0; vet of every added reviewer probe, exit 0; `glob-star-only` mutant `go build ./...`, exit 0, restored byte-for-byte. No local `go test`. Candidate diff --check clean. All Story code remains unchanged by this review.

[Targeted run 37852472269](https://github.com/relux-works/curator/actions/runs/37852472269), immutable snapshot `8b19472fce2779df4d5bd87ab93e8ad6d6af3de2`, differs only by four review files. Valid counted baseline: 16/16 top-level tests across install/CLI/grammar/audit (the original cache probe is excluded from proof). Directory grammar/schema vectors: 31/31 selected subcases; production glob rows: 3/3. Both project/global changed-declaration subcases fail on the unmutated candidate. Runtime-project-only and glob-star-only mutants are killed by named real install rows.

[Status/native run 37852624160](https://github.com/relux-works/curator/actions/runs/37852624160), snapshot `5d4323c792f6756e56ac2a55d201499e1bddc045`, also differs only by the four review files. Real native build probe passes 1/1 on macOS; changed-ref CLI status probe fails on unmutated production with both JSON and exit-code assertions.

[Corrected audit run 37852982027](https://github.com/relux-works/curator/actions/runs/37852982027), snapshot `decc8973ecaddd9783bd289d965c8344e5cfb6b1`: corrected CLI cache baseline passes 1/1 and kind/commit-only mutant is killed. Correction boundary: the initial CLI cache probe could pass a wrong cache hit because script-policy backfill rewrote a null policies record. It is excluded from evidence. The corrected fixture supplies an empty policies array, preventing unrelated backfill from satisfying the oracle; baseline and mutant were both rerun.

Total counted positive targeted checks: 18/18 top-level tests. Selected narrowing mutants detected: 3/3 (`runtime-project-only`, `glob-star-only`, corrected `audit-commit-only`). This is targeted coverage, not an exhaustive conformance or mutation claim. Hybrid lock selection indexes, mixed hash-framing/registry behavior and external build arms were not independently attacked.

Evidence attached as plain text: `TASK-260924-4mzun5_review-evidence-rev2.log`, the install/CLI/corrected-audit probe sources and `TASK-260924-4mzun5_review-harness-rev2.txt`. Log fixture paths are neutralized. Temporary hosted branches are deleted after all runs reached terminal state.

Next producer: retain the fixed F1/F2/F3/F5 behavior, v1 preservation and dependency glob rows. Fix both declaration-binding defects, keep both changed-declaration admission and actual moved-tag refusal controls, add permanent regressions plus narrowing mutants, attach compile-only/hosted evidence and publish another reviewer revision. No new research/harness task is required.

```verdict-findings
{
  "findings": [
    {
      "id": "F4",
      "row": "Legacy moved tags and repeat installs",
      "invariant": "Strict-tag policy rejects movement of the same declared tag, while an explicit declaration change between two unchanged tags remains installable.",
      "mechanism": "internal/install/install.go:1519 compares only the old package commit with the newly declared tag commit; the package marker has lost the old ref kind/name. A v1-to-v2 declaration change is therefore falsely reported as movement of v2, on both project and global entries.",
      "reproductions": [
        {
          "test_file": "TASK-260924-4mzun5_review-install-probes-rev2.go",
          "command": "go test -count=1 -timeout=3m -v ./internal/install -run '^TestReviewChangedDeclaredTagIsNotMovedTag$' (hosted run 37852472269; snapshot 8b19472fce2779df4d5bd87ab93e8ad6d6af3de2)",
          "expected_failure": "Both project and global subtests return failed with moved tag for consumer: v2, although v1 and v2 never moved.",
          "pinned_blobs": [
            "sha256:06b29a163466aa77a11e95a6276f44add9a184b850c9480cd292863f42646812",
            "sha256:0be096a79094819e8d467493752978884a8827868bf6e0def77933eaf92ed07a"
          ]
        }
      ],
      "severity": "regression",
      "repeat-of": "none"
    },
    {
      "id": "F6-declared-ref-currentness",
      "row": "CLI status/currentness",
      "invariant": "A changed declared ref makes a v6 package installation non-current even when it resolves to the same commit; status --check must fail.",
      "mechanism": "cmd/curator/main.go:1350 marks a legacy package installation up-to-date using commit equality alone, bypassing the prior ref-kind/name comparison and the v6 lock/manifest binding. Changing v1 to an unchanged v2 tag at the same commit returns up-to-date and exit 0.",
      "reproductions": [
        {
          "test_file": "TASK-260924-4mzun5_review-cli-probes-rev2.go",
          "command": "go test -count=1 -timeout=3m -v ./cmd/curator -run '^TestReviewCLIStatusDetectsChangedTagAtSameCommit$' (hosted run 37852624160; snapshot 5d4323c792f6756e56ac2a55d201499e1bddc045)",
          "expected_failure": "JSON skills.consumer is up-to-date and status --check exits 0 after the Skillfile tag changes from v1 to v2 at the same commit.",
          "pinned_blobs": [
            "sha256:7ff9e7681fca59cf0d453479fdfffc91500c1b4b91eea77397c1d7e4bb16ecc8",
            "sha256:ee6b2483bdc6e024585bf6cd0eb8e5d0bc6bd2b554dd3db588a77b1c4b7ed7dd"
          ]
        }
      ],
      "severity": "bypass",
      "repeat-of": "none"
    }
  ],
  "notes": [
    "Legacy hybrid selection indexes and mixed hash-framing/registry paths were not independently attacked; no conformance claim is made for them."
  ],
  "surface_results": [
    {
      "row": "Spec decision and v6 carrier",
      "result": "held"
    },
    {
      "row": "Project/global runtime and GC",
      "result": "held"
    },
    {
      "row": "Effective lock and manifest binding",
      "result": "held"
    },
    {
      "row": "Compiled providers and receipt-3 records",
      "result": "held"
    },
    {
      "row": "Legacy moved tags and repeat installs",
      "result": "broken"
    },
    {
      "row": "CLI status/currentness",
      "result": "broken"
    },
    {
      "row": "Audit install/CLI/sourceaudit/cache",
      "result": "held"
    },
    {
      "row": "Dependency directory glob grammar",
      "result": "held"
    },
    {
      "row": "Vendored v1/v2 corpus preservation",
      "result": "held"
    },
    {
      "row": "Generation reads and atomic publication",
      "result": "not-attacked",
      "detail": "Reused exact-tree full gate plus source inspection; no separate concurrency adversary in this review."
    }
  ],
  "free_hunt": [
    "Changed-ref status bypass is newly introduced by revision 2."
  ]
}
```
