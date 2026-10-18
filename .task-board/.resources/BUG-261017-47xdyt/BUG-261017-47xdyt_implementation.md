# BUG-261017-47xdyt — canonical ownership preservation

Developer candidate: **Refresh onto 402fdbc9 ready for review; current local suite and exact signed-candidate hosted gate green**.

Round 1 and Round 2 validation records below are historical. The unchanged
implementation and normative ownership rule are described in Round 2; the current
checkpoint, candidate and validation evidence are in "Refresh onto 402fdbc9" below.
Source: a field report. No operator product material is included.

## Candidate and governing source

Story checkpoint: `242daa3f69a6f8c55c21b649299f1de6c3d901c9`.
Signed detached CI snapshot (before the subsequent allowlist/logbook-only correction): `73a3874337293e7f70cfe6859aff0c6284cb60f3`.
The Story branch remains at its checkpoint; all candidate changes remain uncommitted there.
Normative source read from fetched curator-spec main `9dfc557d247243e0e8ba60fb877afe6d1facfd06`: skillfile-sources §2, core §10–§11, manager §2.5. The triage and producer brief were read from the assignment. No release pin or conformance pin changed.

## Implementation and caller analysis

`marker.CheckOwnership` checks the exact canonical entry with `stateread.Lstat`. A proven absent destination is permitted. A present destination must be a real directory with a regular, readable, schema-valid preceding `.csk-install.json` whose name matches the canonical basename. Missing, malformed, linked, unreadable or wrong-name evidence cannot authorize replacement. Ownership is separate from currentness: owned content drift and changed desired state remain repairable. The preceding marker is the canonical ownership record, not a cryptographic authorization; adapter ledgers authorize only their own native entries.

`stageNode` checks ownership before `marker.Current`, preventing invalid-marker repair semantics from becoming takeover authority. `stageStaleSkillRemovals` uses failure-aware inventory reads and checks each visible stale child, collects all named conflicts, and returns no partial plan on refusal. Existing dot-entry exclusion is retained; those entries are neither deleted nor claimed.

- Project: `Project -> runCommit -> stageProjectTargets -> stageNode / stageStaleSkillRemovals`. Selected project destinations and all proposed stale removals require ownership. The mixed fixture refuses without deleting the owned stale node, then removes that owned stale node after foreign conflicts are explicitly removed by the fixture.
- Global: `Global -> runCommit -> stageGlobalTargets -> stageNode / stageStaleSkillRemovals`. Same selected and stale protections; global/native bytes survive refusal and owned repair still works.
- Hybrid: `Project -> stageProjectTargets -> stageNode` with the hybrid store. Selected hybrid destinations receive the same protection. There is no project-driven hybrid sibling sweep; consumer-aware GC remains responsible for shared state. A foreign shared sibling survives a successful owned hybrid repair.
- Native adapter ledgers, authorized takeover backups, transaction ordering, protected-cache rollback and GC implementation are unchanged. Staging refusal returns before installation journaling; the existing publisher retains its protected-cache reversal path.

## Acceptance control map

Brief gap: no attack-surface-catalog surface table was supplied. No catalog-complete coverage is claimed. The six explicit AC rows are the denominator below.

| AC | Named behavioral tests | Production call site / scope |
| --- | --- | --- |
| 1 | `TestInstallPreservesUnownedStaleStoreEntries`, `TestInstallRefusesUnprovenStaleOwnership` | `Project -> stageProjectTargets -> stageStaleSkillRemovals`: unmanaged skill, notes folder and loose file named together, full installed-state/ledger digests unchanged, foreign native link resolves; invalid/read-failing/wrong-name/linked evidence refuses |
| 2 | `TestInstallRefusesUnownedSelectedCanonicalDestination`, `TestInstallOwnedCanonicalPositiveControls` | `Project -> stageProjectTargets -> stageNode`: manual skill and note, notes-only, malformed marker, directory marker, wrong marker name, marker link, canonical link and loose file; absent installation, owned repair and clean reinstall; unrelated instructions retained |
| 3 | Both refusal suites above; `TestCheckOwnershipDistinguishesAbsentFromUnaddressable` | Production ownership calls above and direct marker reader boundary: invalid and directory-marker read failure refuse; untraversable ancestor stays unreadable, dangling ownership link refuses. Permission-denied OS reads are not separately fault-injected; directory-marker and ancestor errors are actual filesystem failures. |
| 4 | `TestInstallPreservesUnownedStaleStoreEntries`, retained `TestRemovedSkillCleanedUp` | `Project -> stageStaleSkillRemovals`: mixed owned/unowned refusal, explicit fixture conflict removal, then successful owned stale removal and selected-node preservation |
| 5 | `TestGlobalCanonicalOwnershipPreservation`, `TestHybridCanonicalOwnershipPreservation`, retained `TestDraftUnmanagedAdapterDestinationRefuses`; existing transaction, staging, scopes and adapters suites; **retained invalid-marker repair reviewer test red** | `Global -> stageGlobalTargets`; `Project -> stageProjectTargets` hybrid branch; unchanged serialized publisher / consumer-aware GC |
| 6 | Five killed narrowing overlays below; full remote CI test lane red | Behavioral production suites under Go overlays; remote workflow runs the exact signed snapshot |

Measured: **6 of 6 AC rows driven; 4 of 6 green**. AC5 cannot claim retained repair-contract compatibility, and AC6 full remote gate is red. The AC2–AC3 invalid-marker refusal controls pass, while the retained external repair test demands the opposite outcome for a schema-invalid marker. Every previously committed reviewer test remains under its original name; the named reviewer regression body is byte-identical, and the only existing test-file change is the obsolete absence-reader allowlist data row described below. Full CI found the retained invalid-marker repair test red; no all-test green claim is made. The checklist for a newly shipped source-text gate is inapplicable: ownership is decided from filesystem and parsed marker state, and the existing static checker only loses an obsolete allowlist data row; its checking logic is unchanged.

## Narrowing mutants

All gates remain present. Go `-overlay` changes behavior without modifying the candidate. No source-text checker is used. Every command below returns real exit **1**, an expected-red behavioral failure, not a green gate. There are no survivors among executed mutants.

| Mutant | Gate narrowed to / admitted member | Named failing test | Result / survival bound |
| --- | --- | --- | --- |
| `stale-loose-file` | Keep all ownership checks except permit the stale entry named `loose.txt` | `TestInstallPreservesUnownedStaleStoreEntries` (missing required named conflict) | Killed, exit 1; this mixed fixture witnesses diagnostic completeness, not deletion after single-conflict admission |
| `selected-notes-only` | Keep selected ownership checks except permit a destination with no SKILL.md | `TestInstallRefusesUnownedSelectedCanonicalDestination/notes-only` (install incorrectly succeeds) | Killed, exit 1 |
| `selected-invalid-marker` | Keep selected ownership gate except admit typed invalid-marker errors | `TestInstallRefusesUnownedSelectedCanonicalDestination/invalid` (install incorrectly succeeds) | Killed, exit 1 |
| `wrong-name-marker` | Keep marker validation except permit the valid marker name `other` | `TestInstallRefusesUnownedSelectedCanonicalDestination/wrong-name`, `TestInstallRefusesUnprovenStaleOwnership/wrong-name` | Killed, exit 1 |
| `stale-unreadable-marker` | Keep stale gate except permit an ownership-marker directory | `TestInstallRefusesUnprovenStaleOwnership/unreadable` | Killed, exit 1 (install incorrectly succeeds) |

Mutant bounds: no exhaustive mutation claim, no cryptographic marker-origin proof, and no malicious concurrent filesystem-race proof. Link neighbours are exercised; arbitrary device/FIFO/socket types are not claimed. Existing transaction boundary/race tests and remote race lanes supplement the ordinary filesystem tests.

## Commands and evidence

Commands are standalone processes with redirected logs, no `tee`. Host: Go 1.25.5, darwin/arm64. Tests use existing private temporary fixtures and Go, with no new product/test library.

| Command | Real exit | Evidence |
| --- | ---: | --- |
| Initial lifecycle mutation | 1 | Estimate was required; estimate 5 then development mutation both returned 0 |
| `go test ./internal/install -run '^TestInstall(PreservesUnownedStaleStoreEntries\|RefusesUnownedSelectedCanonicalDestination\|RefusesUnprovenStaleOwnership\|OwnedCanonicalPositiveControls)$' -count=1` before fix | 1 | `regression-red-01.log`; expected regression witness |
| Same targeted suite with `-v`, candidate | 0 | `regression-green-02.log`; repeat justified by changed inventory reader identity |
| `go test ./internal/install -run '^Test(GlobalCanonicalOwnershipPreservation\|HybridCanonicalOwnershipPreservation\|DraftUnmanagedAdapterDestinationRefuses\|RemovedSkillCleanedUp)$' -count=1 -v` | 0 | `callers-02.log` (four matched tests; rerun after the inventory-reader change) |
| `go test ./internal/marker ./internal/staging ./internal/transaction -count=1` | 0 | `support-01.log` |
| `go test ./internal/marker -count=1` after marker test added | 0 | `marker-03.log` |
| `go test ./internal/scopes ./internal/adapters -count=1` | 0 | `cross-scope-01.log` |
| `go build -o .temp/BUG-261017-47xdyt/curator ./cmd/curator` | 0 | `build-02.log` |
| `golangci-lint run ./internal/install/... ./internal/marker/...` | 0 | `lint-03.log` |
| Four initial overlay `go test -overlay ... ./internal/install -run <named test> -count=1 -v` commands | 1 each | `mutant-{stale,selected,invalid,name}-01.log`; expected-red, named failures above |
| Fifth overlay | 1 | `mutant-unreadable-01.log` |
| Signed detached snapshot push, attempt 1 | 1 | Shell refspec expansion error; no publication; corrected explicit variable boundaries |
| Signed detached snapshot push, attempt 2 | 0 | `gate-push-02.log`; signature verified `G`; Story HEAD unchanged |
| Full remote CI test lane | 1 | [run 39058113827](https://github.com/relux-works/curator/actions/runs/39058113827), head `73a3874337293e7f70cfe6859aff0c6284cb60f3` |

The failed full remote gate follows the existing `scripts/remote-gate.sh` snapshot/push/CI protocol in bounded start/poll/readback calls, rather than holding a shell longer than the headless command budget. No local suite substitutes for a required hosted status. The Linux job artifact reports `go test exit=1, platform-case gate exit=0`. Its six failure events are the crossconformance package, the named invalid-marker repair test, the semantic batch and its execution-policy child, and the envprofile package/static guard. The static guard failure was ordinary stale-allowlist maintenance and was corrected locally; the unchanged named reviewer test remains red. Remaining remote work was cancelled after the definite failure to release this run's runners. This is a failed gate, not a passing or mirrored gate; only the targeted and supporting commands above were rerun locally.

Additional direct validation history: `regression-green-01.log` exit 0 preceded the inventory-reader change; `regression-green-02.log` exit 0 supersedes it for that changed identity. `callers-01.log` exit 0 preceded the same change and `callers-02.log` exit 0 verifies the current identity. `build-01.log` exit 0 compiled to a task-owned root binary, which was moved into task scratch; `build-02.log` exit 0 used the explicit scratch output. `lint-01.log`, `lint-02.log` and `lint-03.log` each exit 0; the last includes the new marker tests. An intermediate shell ran the marker suite then build without preserving the first command's individual status; that marker execution is not credited, and standalone `marker-03.log` exit 0 supplies the required evidence. A compact board query with an unsupported resources field printed a query error despite process exit 0; its result was discarded and the supported checklist-only projection was read successfully. No failed/unknown read is treated as absence.

## Scope, declared bounds and handoff

The product delta is limited to `internal/install/targets.go` and `internal/marker/ownership.go`; new tests live in those same modules. README and LOGBOOK document the preservation behavior and finding; UNRESOLVED_QUESTIONS records the exact owner decision. `internal/envprofile/state_read_guard_test.go` removes only the obsolete `stageStaleSkillRemovals` allowlist row, because that reader now goes through `stateread`; the static scanner and all original test names are unchanged. These are the declared out-of-module changes. No adapter, GC, transaction, source, permission or takeover policy is changed.

Out of contract: triage B, D and E and new generic takeover behavior (brief explicitly assigns only A/C2 and forbids generic takeover); project-driven pruning of hybrid siblings (AC5 requires existing caller contracts); exhaustive exotic filesystem objects or adversarial concurrent mutation (bounded marker/install staging scope; reported as unproven, not covered). Dot-prefixed siblings retain existing exclusion and are left untouched. Missing attack-surface table remains an explicit brief gap.

No direct control-root edits, branch commits, switches, rebases or merges were made. This note is the task-scoped logbook/evidence handoff; repository LOGBOOK changes travel in the candidate Change Request.


## Stop-the-line: exact blocker and decision

The brief requires invalid ownership evidence to refuse replacement and specifically asks for invalid-marker canonical refusal. The current gate implements that faithfully. However, the retained reviewer test `TestDraftExternalExecutionPolicyInvalidMarkerReDerives` and published semantic row `external-evidence-mismatch-execution_policy` explicitly mutate `builds.*.execution_policy` into a schema-invalid marker and demand a successful `install.Project` re-derivation over that destination. `driveExternalMismatch` asserts `marker.ReadState` returns `KindPresent` with `marker.InvalidError`, then requires `result.Status == "ok"`. Both its standalone test and full semantic batch fail with the new ownership refusal. This is a genuine behavior conflict, not unavailable CI or a setup error.

A direct local rerun of the unchanged named reviewer test also exits **1** (`reviewer-contract-red-01.log`). Current-main skillfile-sources §4 requires repair to revalidate exact locked source/evidence; core §10 requires revalidation/rebuild rather than adopting candidate bytes. The pinned semantic corpus requires that repair succeed in this invalid-marker scenario. No conformance row/test was deleted, renamed, weakened, skipped or added to a gap ledger.

Failed assumption: full schema validity of the sole preceding canonical marker could serve as the ownership predicate while preserving the existing invalid-marker repair contract. It cannot satisfy both outcomes for the same schema-invalid preceding marker. Adapter ledgers own native entries; they do not establish canonical ownership. Consumer state records project consumers rather than an independent per-canonical-node ownership record.

Clean options:

1. **Recommended for the preservation-first brief:** make invalid canonical markers refuse, and have the format/contract owner explicitly align the published repair case and retained reviewer expectation. This preserves a simple fail-closed ownership boundary, but requires authorization to revise the conflicting conformance/reviewer contract. This leaf does not silently edit that contract.
2. Preserve successful invalid-marker repair by specifying independently verifiable preceding canonical ownership (for example a separately versioned canonical ownership record). That keeps repair possible without treating the invalid marker as authority, but changes the ownership state model, migration and rollback surface beyond this leaf.
3. Explicitly define a structurally validated ownership subset that remains valid despite invalid build evidence, and narrow the brief's invalid-marker requirement accordingly. The format owner must define that boundary and its negative tests; special-casing `execution_policy` locally would be an unsupported exception.

Exact input needed: the owner must select the canonical ownership/invalid-marker repair contract and authorize the corresponding scope. Until then, full gate green and the unconditional "reviewer tests kept and green" precondition cannot be met together with the current brief. Product-code changes stopped at this evidence; no compensating field-specific bypass was added.

Post-diagnostic changes: stale absence-reader allowlist row removed; `go test ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$' -count=1 -v` exits **0** (`read-guard-01.log`). The full gate was not replayed after this test-only correction because the independently reproduced contract failure remains; no green result is attributed to the changed tree. The current candidate is still uncommitted at the recorded Story checkpoint.


Evidence attachments: `BUG-261017-47xdyt_evidence.txt` contains sanitized logs, real command exits, five narrowing mutant sources and the failed hosted test events. A ZIP attachment attempt was refused by the board's binary-resource contract (exit 1, `resource_binary_refused`); the same evidence was attached as plain text (exit 0). No binary resource is claimed. Final expanded lint over install, marker and envprofile exits 0 (`lint-04.log`). `git diff --check` exits 0. Checkpoint remains unchanged. AC5/AC6 and the code-per-all-AC checklist remain unchecked; no developer handoff is attempted while the unconditional retained-reviewer-green precondition fails.

Hosted run readback: **completed / cancelled**, head `73a3874337293e7f70cfe6859aff0c6284cb60f3`. Cancellation was requested only after the Linux test job had reported an actual failure and uploaded its artifact; the retained failure is not reclassified as a cancellation-only unknown.


## Round 2 against curator-spec 3ca4b84f

The normative owner selected the ownership-valid subset (PR #160), resolving
round 1's conflict without a field-specific exception. Source marker ownership
now requires a regular readable marker opened without following a final link,
link-free canonical components below the trusted project/manager root, a parsed
JSON object, recognized integer schema 5 or 6, and the node's exact basename.
Invalid build/evidence fields remain non-current and repairable. Historical
legacy/global markers retain full validation. A missing/unreadable/unparseable,
linked, special, wrong-name or unknown-version ownership record refuses with
`source_ownership_unproven`. Project and global stagers collect selected and
stale conflicts together before runtime/target staging under the manager lock.
The same checks run at node replacement and stale-removal planning. Existing
transaction preimages and physical boundary checks retain their protections.

Both Skillfile lanes now write the specified manifestless pure-context form:
v6 with `skill_schema_version: null`, never integer zero or a fabricated
manifest. Existing markers are not rewritten on a clean reinstall. The writer
rejects command/runtime/build declarations in that form; normal package loading
still rejects invalid or unreadable present manifests.

Signed candidate: `1c85e638c92edfe1bd87a593b40bb524766c3824`, tree
`54bffdc4a8460c329fd5d68cd4dd412458720c4d`, signature verified `G`.
Story checkpoint stays `1724b4a52b4c2d163c0a3e6ab7c2ad2bbe54ac0c`; no Story
branch commit, switch, rebase or merge. Hosted gate:
https://github.com/relux-works/curator/actions/runs/39078199935 (success).
Earlier round-2 snapshot run 39076712760 was cancelled as superseded after
adding explicit byte comparisons for foreign symlink targets; it is not credited.

### Published ownership control map

The vendored ownership and writer inputs are byte-identical to public
curator-spec `3ca4b84f` under `internal/install/testdata/ownership-r2/`.
Every row below has an executable subtest with the same corpus ID under
`TestInstallDraftOwnershipCases`. Production call site: `install.Project` ->
serialized `stageProjectTargets` -> `checkCanonicalStore`, `stageNode` and
`stageStaleSkillRemovals`. Each refusal checks the named diagnostic and the
prior canonical tree, lock, adapters, runtime and shims, plus foreign symlink
target trees. Every fixture first runs a valid selected control. A separate
control skill remains selected so an empty selection does not shortcut install.

| Published row | Executable subtest suffix | Bound |
| --- | --- | --- |
| `unowned-sibling-preserved-and-named` | `unowned-sibling-preserved-and-named` | Exact filesystem/install verdict on Unix. |
| `unowned-selected-destination-no-partial-change` | `unowned-selected-destination-no-partial-change` | Exact filesystem/install verdict on Unix. |
| `all-unowned-paths-named` | `all-unowned-paths-named` | Exact filesystem/install verdict on Unix. |
| `unreadable-marker` | `unreadable-marker` | Unix regular-file permission failure; Windows uses a directory marker because chmod does not deny reads there. |
| `unparseable-marker` | `unparseable-marker` | Exact filesystem/install verdict on Unix. |
| `nonobject-marker` | `nonobject-marker` | Exact filesystem/install verdict on Unix. |
| `wrong-name-marker` | `wrong-name-marker` | Exact filesystem/install verdict on Unix. |
| `missing-schema-identity` | `missing-schema-identity` | Exact filesystem/install verdict on Unix. |
| `unknown-schema-identity` | `unknown-schema-identity` | Exact filesystem/install verdict on Unix. |
| `string-schema-version` | `string-schema-version` | Exact filesystem/install verdict on Unix. |
| `symlink-marker` | `symlink-marker` | Exact filesystem/install verdict on Unix. |
| `special-marker` | `special-marker` | Real FIFO on Unix; Windows uses a non-regular/unowned directory control, not a FIFO claim. |
| `symlink-canonical-node` | `symlink-canonical-node` | Exact filesystem/install verdict on Unix. |
| `symlink-canonical-parent` | `symlink-canonical-parent` | Exact filesystem/install verdict on Unix. |
| `special-canonical-entry` | `special-canonical-entry` | Real FIFO on Unix; Windows uses a non-regular/unowned directory control, not a FIFO claim. |
| `v6-reader-unaware` | `v6-reader-unaware` | Conditional case: this reader recognizes v6. The executable unknown-carrier control uses 77; it does not claim this implementation is v6-unaware. |
| `ownership-changed-before-publication` | `ownership-changed-before-publication` | Exact filesystem/install verdict on Unix. |
| `owned-stale-cleanup` | `owned-stale-cleanup` | Exact filesystem/install verdict on Unix. |
| `new-destination-without-preceding-marker` | `new-destination-without-preceding-marker` | Exact filesystem/install verdict on Unix. |
| `owned-v5-invalid-execution-policy-rederived` | `owned-v5-invalid-execution-policy-rederived` | Project install re-derivation; published status/repair semantics additionally covered by the retained external mismatch suite for v5. No separate v6 CLI status/repair claim. |
| `owned-v6-invalid-execution-policy-rederived` | `owned-v6-invalid-execution-policy-rederived` | Project install re-derivation; published status/repair semantics additionally covered by the retained external mismatch suite for v5. No separate v6 CLI status/repair claim. |

Mapping: **21 of 21 ownership rows mapped**; **20 of 21 use the published
reader identity on Unix**. The conditional v6-unaware row is an explicitly
stated bound. **17 of 17 rejection classifications have named executable
negative controls**, with that conditional identity substitution declared.
This is not a claim to drive all status/repair suboperations in the two
multi-operation rows. The retained
`TestDraftExternalExecutionPolicyInvalidMarkerReDerives` passed directly;
its original name and body are unchanged.

### Published writer control map

Each published writer row has an executable same-ID subtest under
`TestInstallDraftMarkerWriterCases`.

| Published row | Production entry | Claim / bound |
| --- | --- | --- |
| `manifestless-schema1` | `install.Project` | Writes expected carrier/version on the stated Skillfile lane; clean reinstall leaves marker bytes identical. |
| `manifestless-schema2` | `install.Project` | Writes expected carrier/version on the stated Skillfile lane; clean reinstall leaves marker bytes identical. |
| `schema9-schema1` | `install.Project` | Writes expected carrier/version on the stated Skillfile lane; clean reinstall leaves marker bytes identical. |
| `manifest8-schema2` | `install.Project` | Writes expected carrier/version on the stated Skillfile lane; clean reinstall leaves marker bytes identical. |
| `manifestless-declares-command` | `marker.Write` | Refuses a candidate manifestless marker carrying this declaration; existing marker bytes preserved. Package-admission and legacy-fallback semantics are not claimed by this marker-level control. |
| `manifestless-declares-runtime` | `marker.Write` | Refuses a candidate manifestless marker carrying this declaration; existing marker bytes preserved. Package-admission and legacy-fallback semantics are not claimed by this marker-level control. |
| `manifestless-declares-build` | `marker.Write` | Refuses a candidate manifestless marker carrying this declaration; existing marker bytes preserved. Package-admission and legacy-fallback semantics are not claimed by this marker-level control. |
| `manifestless-declares-legacy-agents-runtime` | `marker.Write` | Refuses a candidate manifestless marker carrying this declaration; existing marker bytes preserved. Package-admission and legacy-fallback semantics are not claimed by this marker-level control. |
| `invalid-manifest-is-not-absence` | `closure.RefreshDraft` | Named invalid/unreadable present manifest refusal at real source admission; no lock or installation publication. |
| `unreadable-manifest-is-not-absence` | `closure.RefreshDraft` | Named invalid/unreadable present manifest refusal at real source admission; no lock or installation publication. |

Mapping: **10 of 10 writer rows driven at the named production boundaries**:
4 Project acceptance/reinstall controls, 2 source-admission refusals, 4
marker-writer refusals. These boundaries are declared rather than counted as
10 end-to-end CLI installs. No candidate-conformance result is used to claim
execution of the new draft corpus; hosted candidate comparison remains rc.14.

### Narrowing mutant results

All mutants preserve the gate and admit one bounded class. Go overlays execute
behavioral suites; no source-token checker substitutes for behavior. Real exit
1 is an expected-red failure, never a passing gate. No executed mutant survives.

| Mutant | Gate narrowed to / admitted member | Named failing test | Result / survival bound |
| --- | --- | --- | --- |
| stale-loose-file | Admit only `loose.txt` despite absent ownership | `TestInstallPreservesUnownedStaleStoreEntries` | Killed, exit 1; diagnostic completeness witness in a mixed fixture |
| selected-notes-only | Admit markerless notes-only directories, keep other checks | `TestInstallRefusesUnownedSelectedCanonicalDestination/notes-only` | Killed, exit 1; unauthorized install succeeds |
| selected-invalid-marker | Admit exactly an empty parsed object as ownership | `TestInstallRefusesUnownedSelectedCanonicalDestination/invalid` | Killed, exit 1; invalid identity admitted |
| wrong-name-marker | Admit name `other` only when the full marker otherwise validates | `TestInstallRefusesUnownedSelectedCanonicalDestination/wrong-name` | Killed, exit 1 |
| stale-unreadable-marker | Admit a directory in place of the ownership marker | `TestInstallRefusesUnprovenStaleOwnership/unreadable` | Killed, exit 1; unauthorized stale removal succeeds |
| parseable-wrong-name | Admit the two-member parseable schema/name object naming `other` | `TestInstallRefusesUnownedSelectedCanonicalDestination/wrong-name-partial` | Killed, exit 1; ownership subset still requires exact name |
| manifestless-runtime | Permit a runtime root in the null-form marker, keep other validation | `TestInstallDraftMarkerWriterCases/manifestless-declares-runtime` | Killed, exit 1; pure-context guard is necessary |

### Caller contracts and scope

Project and global sweep only ownership-proven stale nodes and refuse every
unproven selected/stale path together. Hybrid checks selected shared nodes;
project install does not acquire authority to prune unrelated hybrid siblings.
The retained positive controls cover absent install, owned update, clean
reinstall, actual owned stale removal after resolving mixed conflicts, and
unrelated authored instructions. Native adapter refusal and authorized-backup
semantics, GC and transaction implementation are unchanged.

The rework is inside install staging and marker modules. Declared supporting
changes: the existing envprofile source-read guard loses only the obsolete
stale-reader exception; README and LOGBOOK document the contract. The previous
unresolved ownership/repair paragraph has been removed because spec PR #160
settled that decision. Writer production changes extend to `draftruntime.go`,
`install.go` and `legacypackage.go` in the named install module, required by the
explicit round-2 writer corpus. No release/conformance pin change or generic
takeover flag is introduced.

Brief gap: no attack-surface-catalog surface table was supplied. The explicit
AC/control maps are used; no catalog-complete assertion is made. Out of scope:
triage D/E operator documentation, new native takeover behavior (the original
brief assigns A/C2 and forbids generic takeover), project pruning of unrelated
hybrid siblings (AC5 retains caller contracts), arbitrary malicious filesystem
races beyond manager serialization/transaction guards, and unsupported-platform
FIFO behavior. The v6-unaware conditional and per-row execution boundaries above
are stated bounds, not silent coverage. No new source-text gate is introduced.

### Validation status

Full local coverage is **85 of 85 packages green**, assembled from bounded
standalone package/test-name subsets. All **420 of 420 CLI top-level tests**
pass on the current corrected test candidate, with no CLI skip. Non-CLI inputs
remain exact and unchanged across the CLI-fixture-only snapshot correction.
`r2-local-suite-coverage.json` records the exact green command set; the evidence
attachment contains command exits and full available green logs. Install
coverage includes all 371 top-level tests, envprofile 369, cross-conformance 54.
No timed-out or failed parent process is called a pass: unaffected green package
results are retained, and affected packages have replacement green subsets.

Build `r2-build-01`, final module lint `r2-lint-final`, CLI fixture lint and
`git diff --check` exit 0. Seven narrowing overlays exit 1 with their intended
named failure, not a generic crash. Every prior named test is retained: 3666
names across 669 files, checked independently of execution evidence.

Local toolchain is Go 1.25.5 darwin/arm64 with the immutable rc.14 corpus root.
The existing godriver platform/optional-real-driver skips retain their stated
conditions; real compiler/probe qualification is separately run by hosted jobs.
This does not claim local Windows behavior or optional driver flags were run.

Earlier red command records are retained as diagnostics: ambient provider PATH,
inherited Git signing injection, an archive export-subst byte-exact fixture,
a missing sibling corpus and bounded package timeouts. Each affected test/input
has a specified corrected green replacement; none is silently credited as green.
The published corpus fixtures themselves remain byte-identical to the owner rev.

The full hosted gate is **success on the exact signed candidate**: 20 executed
jobs green, including Linux/macOS test and race, Windows test, all nine qualified
Go-family jobs, lint, interoperability and gate selftests. The workflow's
candidate-comparison and self-hosted macOS jobs are conditionally skipped, not
executed successes. There is no local-CI fallback or synthetic check. All five
completed test/race artifacts were downloaded and checked for final verdicts,
complete served-package results, no fail actions and no deferred packages.
The artifact summary records hashes, platform-case reports and the named
ownership/writer/reviewer verdicts. Windows runs the declared platform
substitutions in the row map, not a claimed Unix FIFO/permission model.

Exact remote head, signature G, worktree tree and unchanged Story checkpoint
were verified. The three task-owned temporary gate branches were then removed
with exact expected-head leases (exit 0 each); signed objects and hosted evidence
are retained. The Story branch remains at its checkpoint and all producer
changes remain uncommitted for the Change Request snapshot.

### Hosted diagnostic and bounded fixture correction

The preceding signed snapshot `2749f82b` ran the full hosted test lane. Its
Linux Go JSON reports exactly three failing named tests (not the expected-red
output printed inside existing mutant harnesses):
`TestStatusCheckRefusesV1NULAppearingAfterInstall`,
`TestStatusCheckAdmitsV2NULBearingInstall`, and
`TestStatusAcceptsAnUnchangedLegacyMarkerSchema`. The lane exit is 1.
The remaining jobs were cancelled after preserving the failure artifact,
not credited as green. This was a repository test-fixture mismatch, not an
infrastructure failure or a justification for a local-CI fallback.

All three tests used manifest absence to select the historical marker lane.
The new normative manifestless writer now selects source v6 with a frozen v1
package identity, so those inputs no longer exercise their claimed legacy
carrier/hash contract. The three tests keep their original names and every
assertion. Their individual package fixtures now declare a valid schema-4
manifest before tagging the source. No shared fixture or product code changed
for this correction. The three named tests rerun green together (exit 0), and
CLI fixture lint exits 0. New manifestless behavior remains covered by both
published Skillfile writer rows and their no-rewrite controls. This supporting
CLI-test-only change is outside the named install/marker module and is explicitly
declared here; the requirement to preserve legacy behavior motivates it.

All previously green non-CLI package inputs, ownership tests and mutant sources
remain byte-identical. Their exact command results are retained. The complete
CLI test set is being rerun in bounded subsets on the corrected test candidate;
no earlier hosted failure or cancelled lane is converted into a success.

### Local fixture-environment diagnostic

The first local 60-test CLI batch covering hostile Git configuration exited 1:
`TestSignerRowsSurviveHostileGlobalGitConfig` failed its signing precondition.
Inherited agent Git configuration injection overrode the fixture's generated
signing key. The repeat runs in a child environment without external
`GIT_CONFIG_COUNT`/key/value injection and SSH agent/transport variables, while
retaining the fixture's hostile global configuration and every assertion. The
same 60-test batch exits 0, including the real signed probe and named signer
test (`r2-cli-large-11-14-isolated`, 268.228s). This is a test-environment
correction, not a production change or a request for another signing identity.
Only detached candidate publication uses the configured agent signing lane.

Non-CLI local coverage is now 84 of 84 packages green, including all 371 install,
369 envprofile and 54 cross-conformance top-level tests in bounded subsets.
Earlier package-process timeouts and setup failures remain red records; their
previously passing unaffected packages are reused, and the failed/timed-out
packages have explicit replacement green subsets. At that checkpoint CLI
coverage was still running; the later validation section records its green replacement.

### Round 2 acceptance control map

The current denominator is the six supplied AC rows. The Round 1 map above is a
historical failed candidate and does not describe this candidate.

| AC | Executable witnesses | Production entry / call site |
| --- | --- | --- |
| 1 | `TestInstallPreservesUnownedStaleStoreEntries`, `TestInstallRefusesUnprovenStaleOwnership` | `Project -> stageProjectTargets -> checkCanonicalStore / stageStaleSkillRemovals`: all unowned paths named, exact canonical/ledger/adapter bytes preserved, foreign link still resolves |
| 2 | `TestInstallRefusesUnownedSelectedCanonicalDestination`, `TestInstallOwnedCanonicalPositiveControls` | `Project -> stageProjectTargets -> stageNode`: unproven selected nodes refuse; absent install, owned update and clean reinstall succeed; unrelated instructions survive |
| 3 | `TestInstallDraftOwnershipCases` (21 mapped rows), `TestCheckOwnershipDistinguishesAbsentFromUnaddressable` | `Project` serialized stagers and `marker.CheckOwnership`: real Unix permission denial, invalid JSON/identity, wrong name, canonical/marker links and special files; owned invalid build fields re-derive |
| 4 | `TestInstallPreservesUnownedStaleStoreEntries`, retained `TestRemovedSkillCleanedUp` | `Project -> stageStaleSkillRemovals`: mixed conflict refuses atomically; after removing fixture conflicts, owned stale entries are removed and selected nodes survive |
| 5 | `TestGlobalCanonicalOwnershipPreservation`, `TestHybridCanonicalOwnershipPreservation`, retained `TestDraftUnmanagedAdapterDestinationRefuses`, retained `TestDraftExternalExecutionPolicyInvalidMarkerReDerives`; transaction/staging/GC suites | `Global -> stageGlobalTargets`; project/hybrid `stageProjectTargets`; unchanged serialized publisher and consumer-aware GC |
| 6 | Seven narrowing overlays and the exact signed candidate's full hosted gate | Same live ownership/writer gates through behavioral suites; hosted workflow on `1c85e638` |

No previously committed reviewer test was removed or renamed. The repair
regression body remains byte-identical. The three declared legacy CLI fixtures
retain names and assertions and now declare their actual intended manifest lane.
The existing static read scanner still detects its planted collapse in
`TestManagerReadScannerFindsAliasedNotExistCollapse`; its logic is unchanged.
Ownership mutants preserve the searched `stateread.ReadDir` token and execute
the behavioral suite, including the narrowed loose-file refusal control.

Measured: **6 of 6 AC rows driven; 6 of 6 green** within the explicit per-row bounds. Seven of seven narrowing mutants are killed; no survivor. Full local and hosted gate requirements are satisfied. The developer candidate is ready for review; acceptance and integration remain the review lifecycle.


## Refresh onto 402fdbc9

### Commands and actual verdicts

| Command / lane | Exit / verdict | Bound |
| --- | --- | --- |
| README byte comparison against `HEAD:README.md` | 0 | Removing the single ownership paragraph reproduces trunk bytes exactly. |
| `go build -o .temp/BUG-261017-47xdyt/r3-refresh/curator ./cmd/curator` | 0 | Fresh darwin/arm64 build of this tree. |
| `golangci-lint run ./internal/install/... ./internal/marker/... ./cmd/curator/...` | 0 | Changed modules and supporting CLI fixtures. |
| `go test -count=1 -timeout=5m -json -p 4 <packages> -run <mask>`; remaining consolidated CLI calls use `-timeout=7m` | 0 for all 98 selected commands | **85/85 packages; 3504/3504 top-level tests driven.** Every exact argv, PID, exit and complete log is in the evidence archive and `local-suite-coverage.json`. |
| Three `go test` source-shape masks of `TestProfileInstallActivationMatrix` | 0 each | **36/36 matrix combinations pass**; included in the full-suite denominator. |
| Seven `go test -overlay ...` behavioral controls | 1 each, expected red | **7/7 killed**, named assertions recorded; no survivor. |
| `go test -tags snapshotcontrols -count=1 -timeout=5m -run TestDestinationNarrowingControls ./internal/envsnapshot` | 0 | Trunk README's retained snapshot-publication controls. |
| `git diff --check` | 0 | Whitespace. |
| `python3 verify-coverage-checker.py` | 0; inner intact check 0 / truncated-log check 1 | Isolated missing-package-verdict plant is rejected for the intended reason; expected-red inner exit is not green product evidence. |
| Full hosted CI on `db030e1faca9dfc0506876e14e724414ac563030` | success | [https://github.com/relux-works/curator/actions/runs/39090264021](https://github.com/relux-works/curator/actions/runs/39090264021): 20 executed jobs green, 2 conditional skips. All five test/race artifacts verified; no fail action or deferred package. |

Python uses only the standard library for task-specific command orchestration, byte inventories and evidence reporting, as in the triage; no product test or dependency is added.

Local environment: Go 1.25.5 darwin/arm64, exact rc.14 manifest digest verified;
provider PATH and external Git injection are isolated as in Round 2. Optional
real-driver flags are unset locally; hosted qualification supplies them. Upstream
platform/optional skips are explicitly listed in the coverage JSON, not counted
as executed successes. The prior cancelled scratch CI on `5d8e6b07` supplies no
current-tree evidence.

Diagnostic `suite-048` has real exit **1**: its process reached the 360s kill
boundary after shared host-GOROOT lock contention, while the activation matrix
was still running; there is no named failing assertion. It is not green evidence.
All four neighbouring tests are rerun in a successful consolidated mask, and all
36 matrix cases are rerun in the three successful source-shape masks. Eighteen
unrun masks were consolidated into eight bounded CLI calls without omitting any
test. An initial lint wrapper exited 1 before launching the binary because PATH
isolation excluded its directory; the standalone replacement above exits 0.
No product or test was modified to address either command-environment issue.

**6 of 6 AC rows driven and green**, at the production entries and within the
explicit bounds of the retained implementation control map. All 3676 prior test
names are retained; the twelve named preservation, repair and scanner controls
pass directly in this run. No reviewer test is deleted, renamed or weakened.

Brief gap retained: no attack-surface-catalog surface table was supplied. No
catalog-complete claim. No new out-of-contract row is introduced: the Round 2
v6-unaware conditional substitution, marker-writer-only declaration controls,
Windows permission/FIFO substitutions, hostile concurrent filesystem-race bound,
and triage D/E exclusion remain explicit in the implementation note. Project,
global and hybrid caller contracts and native takeover/backup behavior remain as
reviewed. No `refresh-candidate` was needed: this run already starts at the
requested converged checkpoint. Story HEAD and its real index remain at that
checkpoint; the candidate stays uncommitted for handoff.

### Task-scoped logbook

The README combination is coherent and requires no edit. This refresh changes
no repository source, tests, configuration or documentation bytes. Existing
producer changes are preserved. The full local suite, seven narrowing mutants
and full signed-candidate hosted gate were rerun; no local mirror or synthetic
hosted status is used. The candidate is ready for review; acceptance and
integration remain the orchestrator's lifecycle.

### Replayed narrowing controls

| Mutant | Gate narrowed to | Named failing test | Verdict / survivor bound |
| --- | --- | --- | --- |
| manifestless-runtime | Permit a runtime root in the null-form marker | `TestInstallDraftMarkerWriterCases/manifestless-declares-runtime` | Killed, exit 1; no survivor |
| parseable-wrong-name | Admit a two-member schema/name object naming other | `TestInstallRefusesUnownedSelectedCanonicalDestination/wrong-name-partial` | Killed, exit 1; no survivor |
| selected-invalid-marker | Admit exactly an empty parsed marker object | `TestInstallRefusesUnownedSelectedCanonicalDestination/invalid` | Killed, exit 1; no survivor |
| selected-notes-only | Admit markerless notes-only directories | `TestInstallRefusesUnownedSelectedCanonicalDestination/notes-only` | Killed, exit 1; no survivor |
| stale-loose-file | Admit only loose.txt without ownership; other stale checks remain | `TestInstallPreservesUnownedStaleStoreEntries` | Killed, exit 1; no survivor; mixed-fixture diagnostic witness |
| stale-unreadable-marker | Admit a directory as the ownership marker | `TestInstallRefusesUnprovenStaleOwnership/unreadable` | Killed, exit 1; no survivor |
| wrong-name-marker | Admit other only for an otherwise fully valid marker | `TestInstallRefusesUnownedSelectedCanonicalDestination/wrong-name` | Killed, exit 1; no survivor |
