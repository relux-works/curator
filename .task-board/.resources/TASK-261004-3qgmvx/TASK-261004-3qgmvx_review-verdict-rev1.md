# Accepted research review — revision 1

Task: TASK-261004-3qgmvx — csk-gap-analysis-follow-ups-design.
Parent: STORY-261004-3e03l7 — design-csk-gap-follow-ups.
Change Request: CR-TASK-261004-3qgmvx-1, revision 1.
Verdict: **accepted** as a research/CIP draft; no implementation authorization or release qualification is implied.

## Review identities and method

Reviewed the exact delta from base `9bc8e1a1eace93377e41c36465b26a58ee5ce05a` to candidate tree `cad398242453a42af3675bc844ef1a09352b8843`. It adds only `.research/261004_csk-gap-followups-design.md` (171 lines). The workspace document equals the candidate blob. The pinned research source is curator `ca1b776fb580ec0cee0173bf150daf063023aeaa`; spec rc.14 peels to `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`. A fresh unauthenticated public remote read confirmed the spec tag object and peeled commit. Both independently downloaded raw spec files match the evidence packet's SHA-256 digests.

Read the draft completely, all four scratch probe sources, all 13 recorded command results, production/test fixture call sites, and the exact candidate diff. Rechecked whitespace, artifact identity, evidence digests, privacy patterns and upstream changes. No local Go build or test was run, as required by hosted-evidence mode. The scratch-process results below are accepted from the attached evidence, not independently rerun by this reviewer. The run reports no active goal binding.

## Swept surfaces

| Surface | Result and evidence bound |
|---|---|
| Manager tool resolution | Git production call sites execute bare Git; SSH helper uses LookPath/final EvalSymlinks. Recorded direct and intermediate-hop Git scenarios reach Resolve and execute the synthetic payload. SSH measurement is explicitly helper-only. Safe resolver proposal includes link ancestry, refusal on unreadable checks, launch rechecks and a stated residual same-user replacement risk. |
| Reserved exports | Exact-name collision ownership has no system reservation. Four production project dry-run cases are reproducible from the included source. Case-folding and extension rules are explicitly proposed security rules, separate from general identifier equality. |
| Ordinary launcher PATH | Production staging supplies dependency directories to both Unix and Windows prefixes. rc.14 explicitly requires this. Draft correctly refutes a current spec violation and presents the suffix order as a compatibility-changing amendment; the recorded measurement executes an installed launcher. |
| Audit attribution | Gate runs deterministic detection and stores the selected backend label without dispatch. Included strict-mode probe records zero blocking errors and a falsely labelled verdict. Unsupported refusal, cache provenance, child environment, canary, cap, egress and secret transport are addressed without claiming measured leakage. Detailed backend work stays with its existing owner. |
| Global selector | CLI flags lack --only; corrected CLI process reaches flag parsing with synthetic configs and exits 2. Whole desired-state derivation/removal supports the retained-state concern. Proposed selector freezes profile state, retains unrelated installs and refuses shared conflicts atomically. |
| Go families | Frozen source permits only 1.25; recorded real 1.26 probe refuses. Actual 1.27 and platform qualification are explicitly unmeasured. rc.14's mandatory tested 1.23 sentence is correctly identified for reconciliation. See the upstream freshness note below. |
| Legacy extraction totals | Extract reaches unbounded buffered listing and per-file-only planning. Five metadata entries exceed proposed 2 GiB and are accepted by the helper; no payload/disk-exhaustion claim is made. Per-path accounting, listing-stream bounds, overflow, cleanup and creation/reauthentication are covered by the proposal. |

## Citation spot-checks

These independently read anchors exceed the required eight checks. Curator anchors refer to the frozen source commit above; spec anchors refer to the independently verified rc.14 files.

| Anchor | Verified meaning |
|---|---|
| internal/gitops/gitops.go:225–240,596–605,813–817 | Bare Git execution and buffered listing. |
| internal/install/buildsshcandidates.go:125–145 | SSH tool discovery, final symlink resolution and restricted child environment. |
| internal/closure/closure.go:233–245,361 | Exact command ownership and real acquisition resolver route. |
| internal/identifiers/identifiers.go:27–55 | Portable filename/device rules rather than tool reservations. |
| internal/install/install.go:1022–1066; internal/install/targets.go:191–199 | Dependency-directory resolution and production staging. |
| internal/runtimestore/runtimestore.go:161–172,190–205 | Unix and Windows inherited-PATH prefix composition. |
| internal/audit/audit.go:195–205,307–325,413–415,465–478 | Static canary/detection and backend-labelled cache. |
| internal/config/config.go:963–1011 | Backend configuration parsing. Static mode uses the existing backend name string `"null"`; JSON null is not admitted by this parser. |
| cmd/curator/main.go:714–741,1810–1817,1975–1985,2118–2120 | Global parser, Git selection and token argv surface. |
| internal/install/global.go:128–150,393–428,445–449 | Whole closure/desired-state/stale removal. |
| internal/godriver/session.go:42,79–89,245–257 | Frozen tested-family set and refusal. |
| internal/gitops/gitops.go:675–699; internal/snapshot/snapshot.go:91,131 | Per-file bound and creation/reauthentication extraction routes. |
| internal/buildrepo/admission.go:68–85,738,1058 | Separate external-repository defaults and accounting. |
| .github/workflows/ci.yml:38–43 | Exact rc.14 pin. |
| protocol/core.md:51–68,203–213,1221–1225,1540–1576 | Case-sensitive identifiers, ownership/failure, implementation limits and toolchain identity. |
| profiles/manager.md:193–201,662–676,1037–1043,1106–1119,2723–2738 | Go baseline, ordinary/enforced launchers, global scope, optional backend and profile-lock obligations. |

The public GitHub issue #87 remains open and describes the family-policy problem. The official [Go release history](https://go.dev/doc/devel/release) supports the stated two-newer-releases policy and listed 1.26/1.27 releases.

## Evidence and safety

All four probe-source hashes and all 13 output hashes match (4/4 and 13/13). Included probe code plus pinned production fixtures makes each measured finding reproducible. Recorded coverage is 7/7 inspected, 6/7 with a production API/CLI observation, one aggregate helper-only finding, and an additional helper-only SSH subcase. Eleven distinct expected-red Go assertion scenarios establish research gaps; seven existing top-level controls passed. The discarded initial Git/CLI attempts are explicitly excluded. No full-suite, cross-platform, backend execution or Go qualification result is inferred.

The candidate contains no product changes. LOGBOOK.md is byte-identical between base and candidate (SHA-256 `fc621a2abe71bb73935772ffade97d7db7ec4d7dc478d28677b34e4f70476789`). Tracked workspace and index diffs are empty; the sole untracked document is the candidate artifact. Candidate whitespace check passed. Document and evidence inspection found no internal hostnames, personal paths, employer/client identifiers or credential material. Reproduction uses empty environments, private synthetic roots, synthetic SSH tooling/socket and fixture Git identities. No operator login/logout, credential/Keychain access or private comparison report was performed by this reviewer. Inspection supports the producer's safety account; it is not independent historical process attestation.

## Decision readiness and freshness

The draft follows the CIP structure, states design pending, offers three real options with tradeoffs, recommends staged option B, defines mandatory defaults and operator control, gives S1–S7 normative sketches, compatibility guidance, eight recommended operator decisions, ordered leaves and production-entry negative tests/mutants. Every required topic is answered or exposed as an explicit decision with a recommendation. Fixed budget values are labelled design choices rather than measured corpus limits. Existing sibling ownership is preserved.

Fresh public curator main is `a9a7ae683bf8668a45fd1b16aa5fece1ec8f761e`. Upstream has no overlap with the candidate's added research path. It now admits Go 1.26/1.27 in internal/godriver/session.go:42 (landed qualification commit `9bc8e1a1`), so the draft's item 10 describes its explicitly frozen earlier observation, not today's admission set. Draft lines 23–26 expressly exclude claims about later main. Before prioritizing the implied Go leaves, reconcile them with this landed work; this review does not attest its qualification matrix. The upstream audit change removes pin-based finding bypass and does not implement backend dispatch. Neither change invalidates the draft's pinned measurements.

No blocking findings or changes requested. Research acceptance checks are satisfied. The generic tests-green checklist means recorded existing controls and reviewer artifact checks passed here; intentional research reds are not failing product acceptance tests. The nonacceptance-routing checklist branch is inapplicable because this verdict accepts the work. Acceptance must use accept_cr and route the element to integrating; landing and done remain with the authorized producer lifecycle.

## Persisted acceptance and runtime anomaly

`accept_cr` exited 0 and persisted revision 1 as accepted, with TASK-261004-3qgmvx — csk-gap-analysis-follow-ups-design routed to integrating. Its result requires a new integration run bound to researcher/analyst. No reviewer commit acknowledgement or done transition was supplied.

The acceptance operation emitted a non-blocking write-boundary report under policy `warn`. The report includes concurrent foreign-worktree changes and unattributed activity/progress changes for a different task; this reviewer did not edit those files. It separately attributes this task's verdict/resource, checklist and status writes to the authorized board mutations. This is a runtime attribution anomaly, not an approval rejection or a research finding. No foreign files were repaired or changed. The large boundary scan accounts for the delayed board response.
