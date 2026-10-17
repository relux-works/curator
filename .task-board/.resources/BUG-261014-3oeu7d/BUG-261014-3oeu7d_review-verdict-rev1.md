# BUG-261014-3oeu7d review, CR revision 1

Candidate: 88c4200db2dcfce899bde15444031a6e4ea28d13; base 26d8ddbf1278ac025a3333aee1053adbdb37f1bf.
All five changed worktree files byte-match their candidate blobs. Fresh origin/main equals the base. No product files modified by review.
Spec main: 1059b94b4135089d842bf017070b8ab47d79a536, environments 9.1 and 9.4 read directly. AC coverage: 5 of 5 rows mapped to named driving tests in the producer handoff; independently executed controls below. Review budget was not supplied; bounded targeted review used instead. The brief omitted a catalog table; its six explicit review rows are retained below, with catalog associations where applicable.

## Swept surface results

| Row | Result | Executed attack/evidence and limit |
| --- | --- | --- |
| 1 Spec fidelity (identity/error surface) | held | NetworkCapabilities absent/none/other-host/pre-capability cases and NetworkUndeclaredFollowUp drove EnsureDefaultWithPolicy -> migrateGlobalSkills -> strictAuditMember. Required skill/host/file/remedy and no publication asserted; declared exact/glob/alias positive controls passed. behavior.log, exit 0. |
| 2 No widened admission (identity/read boundary/error surface) | held | ManifestErrorsBlock malformed JSON/network and directory-as-manifest; DeclarationKeepsOtherGates neighbour/revocation/canary; old-allow cache forgery and old-refusal controls; Install context negative and positive controls passed. behavior.log, exit 0. Read failure witness is a directory, not OS permission denial. All three strictAuditMember call sites inspected: auditMember, auditNewUpdateMember, migrateGlobalSkills. |
| 3 Mutation controls (identity/error surface) | held | Independently reran TestGlobalSkillCapabilitiesNarrowingMutants: six overlays killed with child exit 1 and specific behavioral assertions. Harness exit 0. No delete-only or static-only substitution. mutants.log. |
| 4 Exact remote gate | held | Live gh run view confirms run 39033322875 success, head 0a2cdb08929afa6b5130c90d9354818077299ff7. Local object signature G; tree equals candidate exactly. Reused producer's completed full gate; did not rerun full remote CI. 20 successful jobs, 2 explicitly skipped optional jobs. |
| 5 Public text hygiene | held | Reviewed all five candidate blobs and searched changed content for private path/identity patterns. Only synthetic example domains and repository-relative source paths occur in added text. No operator material found; this is a bounded content review, not a secret-exhaustiveness claim. |
| 6 Schema floor and generic pipeline (identity) | held | Missing/old schemas produce network finding, not require-pin; modern schema-nine admission passes. Floor-removal mutant killed. Exact modern-schema forwarding is source-inspected, not observably distinguished by schema-nine admission. internal/audit unchanged in candidate. |

## Free hunt and independent execution ledger

No blocking finding reproduced. Added two reviewer-only table cases through a Go overlay over the exact candidate: suffix-confused declaration and wildcard limited to children of the contacted host. Both refused through EnsureDefaultWithPolicy with full diagnostic and no-publication assertions; original valid controls still admitted. Product code was unchanged. The portable reproduction patch and logs are attached in BUG-261014-3oeu7d_review-evidence-rev1.log.

Commands personally rerun (all exit 0, Go 1.25.5 darwin/arm64):
- `go test ./internal/envprofile -count=1 -timeout=5m -v -run '^(TestMigratedGlobalSkill|TestInstalledMemberCapabilitiesStayKindScoped)'`
- `go test ./internal/envprofile -count=1 -timeout=5m -v -run '^TestGlobalSkillCapabilitiesNarrowingMutants$'`
- `go test ./internal/envprofile -count=1 -timeout=3m -v -run '^Test(Revoked(Skill|MCP|Source)MemberIsRefused|RevokedSourceIsRefused|StrictAuditCanary(Passes|FailureBlocksInstall)|UpdateBlocks(Secret|Revoked|V1Revoked)OverlayMember.*|ReinstallBlocksRevokedOverlayMember)$'`
- `go test ./internal/audit ./internal/skillspec -count=1 -timeout=3m`
- `go test -overlay <review-overlay.json> ./internal/envprofile -count=1 -timeout=2m -v -run '^TestMigratedGlobalSkillNetworkCapabilities$'`
- `git diff --check`

Fresh main has no delta from the CR base. No earlier revision findings exist for this first revision. Full remote gate evidence reused from BUG-261014-3oeu7d_remote-gate-02.log and verified live at https://github.com/relux-works/curator/actions/runs/39033322875. No unrelated failures waived. Board run goal queried: not goal-bound. Review made no code or LOGBOOK edits; this task-scoped outcome records the review.

## Verdict

```json
{
  "verdict": "accepted",
  "revision": 1,
  "findings": [],
  "notes": [
    "The supplied brief omitted catalog surface rows and review/free-hunt minutes; all six explicit review requirements were swept and mapped above.",
    "Unreadable-manifest coverage uses a directory-as-file witness, not permission-denied.",
    "Modern schema exact forwarding is source-inspected; the schema-nine test alone proves admission, not the internal numeric value.",
    "Held results report the named attacks only, not proof of absence. Acceptance is not landing."
  ]
}
```
