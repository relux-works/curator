# TASK-260924-4mzun5 — record-dependency-directory-in-legacy-lane: revision 1 review

Verdict: changes_requested; route to to-dev. No acceptance, commit acknowledgement, or integration was performed. The run is not goal-bound (queried immediately before verdict).

Reviewed immutable base 3b6481c15d61b08d3f0fae7c3269329555336709 and candidate tree 31255c82ebc3f8f1fa172fabb359ff33292efab9 (44 paths). Fresh upstream main advertised that same base; fetched main agrees. The original Story workspace was not modified. Review probes and hosted workflow live only in a disposable clone.

## Specification decision

RECORD is the correct decision. Curator-spec 7eaeb73fcf22cb8ce8e21325a8237accdfaf9646, skillfile-sources §4 explicitly requires marker v6 for a core schema-9 installation even when the root uses Skillfile schema 1. Core §4.4 requires normalized directory in installation identity and per-package audit identity; manager §2.1 requires the directory in audit subjects/cache keys. I read core, skillfile-sources, manager profile, draft-sources-v2 README and cases from the pinned spec checkout.

The v6 carrier and directory plumbing are good, but marker migration was not carried through the associated runtime, lock, build and tag-policy contracts. This is implementation rework, not an external/human blocker.

## Required rework (complete swept review, not first-hit)

F1 (high) — package markers and script runtime publication disagree.
internal/install/install.go:1292 migrates legacy schema-9/subdirectory nodes to package markers. However runtimeKeys is computed only when draftLock is non-nil (install.go:797), and draftRuntimeKey falls back to node.Resolved.Commit (draftruntime.go:80). Thus a schema-1 installation publishes a commit-keyed script runtime while its package marker requires the source-v1 package-digest key under skillfile-sources §4. scopes/gc.go:219 marks only the package-derived key for a package marker, so the actual runtime is unreferenced and eligible for removal. This affects even a root schema-9 script skill, as well as selected dependency packages.
Hosted TestReviewLegacyRuntimeMatchesPackageMarker drives install.Project with a schema-1 project and runnable schema-9 script. Install returns ok, marker is readable, but the package-derived runtime script is absent. This is an observed candidate failure, not a mutant. The attached probe source reproduces it. Add production-entry script installation, launch, currentness and GC retention rows for the migrated lane; derive and publish the same package key that the marker/GC require.

F2 (high) — invented lock_sha256 preimage is not the specified lock.
internal/install/legacypackage.go:13 and :101 explicitly substitute a per-node struct (name, package parts, ref) for the absent lock document and hash json.Marshal's struct order. skillfile-sources §3 defines lock_sha256 as the CCJ-1 SHA-256 of Skillfile.lock.json with only lock_sha256 omitted; §4 requires a validated lock and matching manifest and leaves that meaning unchanged for v6. This code binds neither the parsed declaring manifest nor the full closure and is not CCJ-1 over the specified lock. A syntactically valid sha256 string is insufficient. Use the specified validated lock identity for this lane and test marker/manifest/closure binding; do not mint a different wire meaning in this field.

F3 (high) — valid build providers gain an unsupported refusal after compilation.
internal/install/install.go:1293 refuses every migrated node with builds, asserting it requires Skillfile schema 2. Core §4.4 explicitly permits build commands from schema-9 providers; skillfile-sources §4 specifies receipt-3 package build records for the v6 carrier and does not authorize this blanket schema-1 refusal. draftBuildPackages remains nil on the legacy lane, so the implementation avoids the required package receipt integration. The refusal is reached during target staging, after the private-build phase, rather than a source capability preflight. Its sole new pin calls buildMarker directly with a fabricated build record. Integrate the package receipt/build pipeline for these nodes and add real install-entry supported-platform build acceptance and negative receipt-binding rows. A new diagnostic cannot stand in for the specified behavior.

F4 (medium) — moved-tag policy is bypassed for newly migrated legacy installs.
Both project and global legacy installs call detectMovedTagsIn. Its existing recorded.Package != nil early continue (install.go:1423) was valid for the frozen-lock lane, but these new legacy package markers have no frozen lock and legacy resolution still resolves the live tag on each install. After the first schema-9 install, moving that tag no longer emits the warning that Audit.CheckMovedTags consumes, so strict moved-tag policy can be skipped. Manager §2.1 says moved-tag evaluation remains mandatory. Preserve the legacy tag-policy comparison using the package commit plus the actual declared tag binding; add a real second-install moved-tag refusal row. Do not infer absent prior evidence from package presence.

F5 (medium) — the new legacy audit record/cache binds only directory, not full package identity.
Core §4.4 requires the same package identity in the local audit record; skillfile-sources §4 states audit-cache equality includes the complete package identity. loadCachedFindings reads framing, findings and directory only (audit.go:550-589); it never compares canonical repository/configured source or locked commit, and persisted ordinary verdicts do not carry a typed package identity. Identical bytes and directory in another repository/commit can still hit this cache. Existing findings-versus-decision recomputation does not satisfy the explicit package-equality requirement. Bind the selected package identity and prove same-identity acceptance plus different-directory/repository/commit cache misses. This is a static contract finding; no repository/commit cache attack was executed in this review.

## Swept surfaces

| Surface | Evidence/result |
| --- | --- |
| Spec lane decision, schema gate, v6 reader/writer/currentness | Correct record decision; package dependency implications require F1-F3 |
| Project/global install and runtime/GC | Real install positive/negative exists; runnable probe fails (F1); global code has same marker/runtime split |
| Lock identity and manifest/ref binding | F2; invented per-node digest |
| Compiled providers/receipt records | F3; helper refusal only, valid behavior narrowed |
| Legacy moved tags and repeat installs | F4; package-marker skip bypasses legacy comparison |
| Audit install/CLI/sourceaudit/cache | Directory carried at production callers; CLI mutation killed; full identity remains incomplete (F5) |
| Glob grammar dependency path | Four manifest ?/[ vectors and three install-entry *,?,[ rows; star-only mutant killed by both levels |
| Vendored v1/v2 suite and rework | v1 directories have zero base-to-candidate delta; all 23 new skillspec JSON files and 1 marker schema byte-identical to spec |
| Changelog, changed tests, compile/lint/hosted evidence | Operator-visible changelog matches change; diff --check clean; evidence boundaries below |

## Validation and mutation evidence

Accepted existing evidence: hosted full gate 37827088930 completed success including lint, test and race lanes. Its head f9dd6d5f655300b0f9c59c19b82e1f5173369482 has tree 31255c82ebc3f8f1fa172fabb359ff33292efab9, verified through GitHub commit metadata; this is exactly the CR candidate. I did not rerun the full gate. Self-hosted rose-air and candidate-suite lanes were skipped in that run, not proven.

Reran myself: scoped go vet on install, CLI, skillspec and audit in the disposable clone, exit 0; go build ./... baseline and glob-star-only mutant, exit 0; mutant restored byte-for-byte from a saved copy. No local go test ran. Candidate diff --check clean. Vendored JSON byte comparison 24/24.

Hosted targeted run 37833422218, snapshot 2e94ab820cbb56dbe5ee524e3e06a6ee4102f481, differs from the candidate only by the review workflow and adversarial test (verified with git diff). Baseline named rows passed 5/5. Narrowing mutations detected 3/3:
- marker-directory-root: network-git package records root instead of selected directory; TestLegacyLaneRecordsDependencyDirectory fails.
- audit-cli-root: CLI supplies root instead of node.Directory; TestCLIAuditBindsDependencyDirectory fails.
- glob-star-only: ValidDirectory rejects only * instead of *?[]; all 4 manifest glob vectors fail and both ?/[ install-entry vectors fail their required directory diagnostic.
The mutant jobs are green because their named expected test failures were asserted. Runtime-probe is red because an unmutated candidate fails the added real install-entry runtime assertion. Coverage is 3/3 selected mutants, not exhaustive grammar/authorization coverage; package repository/commit cache behavior and compiled installation were not executed independently.

Review harness correction boundary: two preliminary snapshot attempts were excluded, one cancelled and one failing compilation due to omitted candidate additions. The counted snapshot includes every candidate path plus only the two review files. No preliminary infrastructure failure is counted as a killed mutant or candidate failure.

Evidence: TASK-260924-4mzun5_review-targeted-evidence.log and TASK-260924-4mzun5_review-runtime-probe.go. Full hosted logs: https://github.com/relux-works/curator/actions/runs/37833422218 . Spec reference: https://github.com/relux-works/curator-spec/blob/7eaeb73fcf22cb8ce8e21325a8237accdfaf9646/protocol/skillfile-sources.md .

Next producer: retain the correct v6 record decision, v1 preservation and dependency glob tests; address F1-F5, attach compile-only/hosted evidence, hand off a new candidate for another reviewer cycle.
