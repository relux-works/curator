# TASK-260924-4mzun5 — record-dependency-directory-in-legacy-lane: revision 5 review

Verdict: **ACCEPTED**. No P0/P1 finding. Two P2 notes are filed below and do not hold the landing (tb-R226). This run is not goal-bound (`task-board spawn goal` returned none). No product code, commit, acknowledgement or integration change was made by this reviewer.

Candidate: base `3cb461b37d572bcbbb84182fab50fff323c7f9a8`, tree `5bc982a25b05c43a419f538357e76280f398579b`. I rebuilt the tree of the Story worktree with a temporary index (`read-tree HEAD` + `add -A` + `write-tree`) and it equals the candidate tree, so the working tree reviewed is the CR candidate (59 paths). Hosted commit `95f43e31` has that tree (GitHub API), and its full CI run 37900038171 is green (20/20 executed jobs; Candidate suite and rose-air skipped, not proven). The CR's own validation log shows the gate green with exit 0.

## Decision (unchanged, confirmed in earlier rounds)
RECORD. skillfile-sources §4 (curator-spec 7eaeb73f) requires marker v6 for a core schema-9 installation under a Skillfile schema 1; core §4.4 requires the normalized directory in install/lock identity and the same package identity in audit records. Revision-1 F1/F2/F3/F5 and revision-2 F4/F6 originals were verified fixed in rounds 2–4 and nothing in the rev-5 delta touches those mechanisms except `detectMovedTagsIn`, `previousLegacyRef` and the generation record (below).

## F4 (rev-4 P1: unrelated tag defeats strict same-tag refusal) — fixed
Mechanism check (static): `internal/install/install.go` `detectMovedTagsIn` no longer enumerates live tags. For a legacy package marker whose commit differs, whose source/repository/directory is unchanged, it recovers the prior declared ref with `previousLegacyRef` (legacygeneration.go) from the installed-generation record `.curator-generations/<name>.json`, published in the same staging plan as the marker (`stageLegacyGeneration`, project and global). The record is accepted only when: protocol JSON valid, closed shape, `lock.LockSHA256 == marker.lock_sha256`, `lock.CheckStale(manifest)` (lock self-integrity + manifest digest), the lock member's package equals the marker package, and the root selection index names the skill; a transitive dependency's ref is recovered by replaying `closure.buildQueue` order (BFS, sorted requirement names, first declaration wins) over manifests read from the locked Git objects with a content-hash check. Missing/unreadable/forged evidence fails closed with `DiagAbsent`/`DiagUnreadable`; it is never read as "declaration changed". The marker carriers (v5/v6) gain no member; schema-1 re-resolution is unchanged. Moved-tag evaluation also runs in dry runs (planning) because the gate sits before the dry-run split.

My own probes, hosted on a scratch branch of a disposable clone (run https://github.com/relux-works/curator/actions/runs/37912102819, branch deleted afterwards; compile-only locally: `go vet ./internal/install` exit 0, no local `go test`, per R223):
- `TestReviewR5RefreshedEvidenceThenMoveAgain` (project + global): strict install; move v1 with alias `archive1`; dry-run strict refuses; non-strict rebind; move v1 again with alias `archive2`; dry-run strict and real strict both refuse with `moved tag for consumer: v1 <second> -> <third>`. This proves the evidence is refreshed by the rebind (not frozen at the first install). PASS on both entries.
- `TestReviewR5SameCommitRedeclareThenMove` (project + global): v1→v2 declaration at the same commit installs; moving the abandoned v1 does not refuse; moving declared v2 refuses as `v2`. PASS on both entries.
- Producer's permanent rows ran again in my baseline job and pass: `TestLegacyLaneSameTagMoveWithAlias` (project/global × none/lightweight/annotated, 6/6), `TestLegacyLaneChangedTagAfterOldTagDeleted` (2/2, the rev-4 P2 BUG-261009-2s6t3y probe also passes now), `TestLegacyLaneRejectsUnboundPreviousGeneration` (12/12), `TestLegacyLaneDependencyTagMovement` (4/4), `TestLegacyPackageStatusBindingAndRepair` (2/2), F6 status regressions (2/2).

Narrowing mutants I ran (each applied to the candidate in the disposable clone, `go build ./...` ok, then the install suites):
| mutant | result |
|---|---|
| `any-prior-tag` (drop `previous.Value == node.Resolved.Ref`, i.e. the rev-2 false-refusal shape) | killed: `TestLegacyLaneChangedTagAfterOldTagDeleted/{project,global}`, `TestLegacyLaneDependencyTagMovement/*/changed-declaration`, `TestReviewChangedDeclaredTagIsNotMovedTag/*` |
| `never-prior` (waive refusal always, the alias-waiver family) | killed: `TestLegacyLaneSameTagMoveWithAlias` (all 6, including none), `TestLegacyLaneDependencyTagMovement/*/same-tag` |
| `no-lock-digest-check` (accept a record whose lock digest does not match the marker) | killed: `TestLegacyLaneRejectsUnboundPreviousGeneration/{project,global}/rehashed-lock` |
| `no-stale-check` (skip manifest↔lock binding) | killed: `.../{project,global}/manifest` |
| `skip-source-check` (drop `sameLegacyPackageSource`) | **survived** (all suites ok) — see P2-1 |
Mutants killed: 4 of 5. The producer's own red/narrowing runs (37900037908, 37900038056) are consistent with this and were not re-run by me.

## P2 notes (separate BUG notes, do not hold acceptance)
- **P2-1 (test gap, robustness).** `sameLegacyPackageSource` guard in `detectMovedTagsIn` has no negative row: when a skill's repository/directory changes while the same tag name is declared and commits differ, correct behaviour is "no moved-tag diagnostic"; no test would fail if the guard were removed (mutant `skip-source-check` survived). Severity P2: the legacy non-package branch (`recorded.RefKind == "tag" && recorded.Ref == …`) has the same parity gap, and removing the guard yields a recoverable false refusal at worst. Suggested row: install at v1 from repo A, redeclare the same name/tag from repo B (different commit), strict reinstall must succeed.
- **P2-2 (robustness, not reproduced; stated bound).** A prior-declaration lookup failure is a hard install error even without StrictTags (`cannot establish prior declared ref for …`). Reachable cases I reasoned about but did not run: a draft-lane package marker (no `.curator-generations` record) re-installed under a schema-1 Skillfile after the commit differs, and a transitive dependency whose old consumer commit was pruned from the cache clone. Fail-closed is the specified safe direction, the message names the cause, and a fresh install repairs it, so this is P2; I did not attempt hosted reproductions.

## Coverage statement
Targeted: F4 at project and global production entries, evidence refresh, same-commit redeclaration, tampering of the stored record (existing rows), five mutants. Not attacked in this round: hybrid-store layouts, concurrent installs against the generation record, external build arms, registry/hybrid selection. The full configured gate (green on the exact tree) is the broader arbiter.

Attached: `TASK-260924-4mzun5_review-probes-rev5.go`, `TASK-260924-4mzun5_review-harness-rev5.txt`.

```verdict-findings
{
  "findings": [],
  "bug_notes": [
    {
      "id": "P2-BUG-1",
      "row": "Legacy moved tags and repeat installs",
      "severity": "P2",
      "severity_reason": "Missing negative test for the source-identity guard; mutant skip-source-check survives; worst case is a recoverable false refusal.",
      "mechanism": "internal/install/install.go sameLegacyPackageSource guard has no row where repository/directory changes while the tag name stays.",
      "test_file": "TASK-260924-4mzun5_review-harness-rev5.txt",
      "command": "Hosted run 37912102819, job mutants (skip-source-check): install suites ok",
      "repeat-of": "none"
    },
    {
      "id": "P2-BUG-2",
      "row": "Legacy moved tags and repeat installs",
      "severity": "P2",
      "severity_reason": "Fail-closed hard error without StrictTags when prior-declaration evidence cannot be established; not reproduced, repairable by reinstall.",
      "mechanism": "detectMovedTagsIn returns an error for a missing/unusable installed-generation record or pruned consumer commit regardless of strict policy.",
      "test_file": "none",
      "command": "static analysis only",
      "repeat-of": "none"
    }
  ],
  "notes": [
    "Candidate tree 5bc982a2 verified equal to the reviewed worktree; hosted full CI green on commit 95f43e31 (run 37900038171).",
    "No independent hybrid/registry/external-build/concurrency coverage is claimed."
  ],
  "surface_results": [
    {"row": "Spec decision and v6 carrier", "result": "held"},
    {"row": "Project/global runtime and GC", "result": "held"},
    {"row": "Effective lock and manifest binding", "result": "held"},
    {"row": "Compiled providers and receipt-3 records", "result": "held"},
    {"row": "Legacy moved tags and repeat installs", "result": "held"},
    {"row": "CLI status/currentness", "result": "held"},
    {"row": "Audit install/CLI/sourceaudit/cache", "result": "held"},
    {"row": "Dependency directory glob grammar", "result": "held"},
    {"row": "Vendored v1/v2 corpus preservation", "result": "held"},
    {"row": "Generation reads and atomic publication", "result": "not-attacked", "detail": "Source review and exact-tree full gate; no separate concurrency adversary."}
  ]
}
```
