# Review verdict: accepted

Task: TASK-261004-2iewnz — research-105-design-and-implementation-plan.
Date: 2026-10-05. Role: reviewer. Change Request revision: 1.
Candidate tree: `1bc879cf3414fe2e17d1c5b6f5961f9a68390788`.
Base: `9bc8e1a1eace93377e41c36465b26a58ee5ce05a`.
Verdict branch: accepted; route through accept_cr to integrating, not done.

## R1 resolution

The prior verdict accepted the design and requested reproducible Python evidence. Companion evidence E8 now supplies three complete sanitized replacement bodies, fixture construction, configuration snapshots, pinned inputs, expected counts, standalone outputs, real exits and date. It explicitly states that the original bodies were unavailable and labels these as replacements; it does not invent recovered originals. E5/E7 remain historical, contextualized by E8.

- E8.1, lines 206–472: pinned export, synthetic settings/MCP/plain packages, refusal diagnostics, unchanged-config bytes, publication/selection paths and lock bookkeeping; recorded 18/18, exit 0 on 2026-10-05.
- E8.2, lines 474–616: pinned export, plain control install, unknown-knob/wrong-enum refusals and byte-identical post-install configuration; recorded 7/7, exit 0 on 2026-10-05.
- E8.3–E8.5, lines 618–859: complete artifact checker and bounded publication patterns, recorded 11/11, exit 0 on 2026-10-05, with explicit link/privacy/size bounds.

I extracted all three bodies to disposable scratch and compiled their Python syntax successfully. I independently reran the artifact checker as a standalone process: 11/11, exit 0 on 2026-10-05; 10/10 sections, 38/38 pinned links in range, 10/10 JSON examples, both files passing fence/newline/whitespace checks, exactly two research changes, diff check green, LOGBOOK.md untouched and zero publication-pattern matches. Final measured combined size: 96,520 bytes.

I accepted the attached 18/18 and 7/7 results after reading their full bodies and fixtures; I did not rerun those scripts because both invoke go build. Hosted-evidence mode prohibits local Go builds/tests. These script bodies are reproductions, not compliant Mac-host build-wrapper recipes; any future execution on this host must separately obey its -work/FIFO/crash-count rules. No local Go gate, provider process, login/logout or real credential read was performed in this review.

## Swept surfaces

| Surface | Result |
|---|---|
| Issue scope and decisions | Pass. Direct issue #105 read confirms nine provider paths plus per-server MCP intent; all are mapped. Decision 0014 remains proposed; Decision 0018 launch-mode authority remains separate. |
| Evidence and citations | Pass. Fifteen source/spec citation groups read at pinned revisions, listed below. Replacement checks are reproducible and measured claims are bounded. |
| Decision readiness | Pass. Four options, concrete machine record/API, presence/defaults/precedence, project profile/root guard, five decision-ready questions with recommendations. |
| Architecture and migration | Pass at research-draft level. Existing config/resolve/home/launcher surfaces, owned-key acquisition/drift/release, historical seeds, negative-only MCP and downgrade are covered. Native adapter assumptions are explicit implementation qualifications rather than claimed measurements. |
| Specification/leaves/tests | Pass at research-draft level. Narrow CIP adoption, versioned closed schemas/fragments, L1–L8 sizes/dependencies and real-entry negatives/narrowing mutants are specified. |
| Safety/publication | Pass within observable bounds. Reviewed artifacts contain no evident secrets, private hosts, personal paths or employer/client names. Bounded scan green. Prior no-secret assertions are not independently provable from artifact inspection alone. |
| Diff hygiene | Pass. Exact base-to-candidate delta contains only the two research files; both worktree files match candidate blobs; candidate diff check green. No code or LOGBOOK.md changes, and reviewer changed no repository files. |
| R1 and hosted validation | Pass. R1 checks addressed as above; hosted gate independently confirmed successful for the exact candidate tree. |

## Citation spot-checks and freshness

Curator baseline: `ca1b776fb580ec0cee0173bf150daf063023aeaa`.
Spec tag: `v1.0.0-rc.14`, cloned independently and verified peeled commit `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`.

Read these fifteen citation groups, checking their substantive support rather than only anchor existence:

1. internal/contextpkg/contextpkg.go:168–193 — closed manifest keys exclude settings/policy.
2. internal/config/environments.go:270–275 — native/yolo permission enum.
3. cmd/curator/envconfig.go:58–102 — ordinary JSON fallback and public mutation entry.
4. internal/envprofile/overlays.go:15–43 — per-profile machine composition.
5. internal/envprofile/managed.go:2389–2437 — named/env-current/machine-current selection.
6. internal/envregistry/envregistry.go:28–37 — shipped seed A.
7. internal/envprofile/managed.go:823–850 — upfront seeds and provisioning-only copying.
8. internal/contextmaterialize/mcp.go:214–238 — empty MCP set omits output.
9. internal/envmarker/envmarker.go:307–315 — closed four-surface ledger.
10. internal/envprofile/managed.go:204–223 — declaration names derive from MCP lock members.
11. decisions/0014-tool-configuration-surfaces.md:3–10 — proposed, no option selected, no implementation authorization.
12. decisions/0018-curator-run-permission-interface.md:324–330 — adopted choices and provider/launcher ownership separation.
13. protocol/environments.md:35–46 — revision-1 profiles exclude settings.
14. protocol/environments.md:1826–1857 — Claude strict channel and Codex inherited-base residual.
15. protocol/environments.md:3093–3134 — native is no override, launcher permission resolution and transport.

Fresh remote main is `9bc8e1a1eace93377e41c36465b26a58ee5ce05a`, equal to the CR base and local main. None of the eight inspected implementation files changed between the dated citation baseline and this base. The dated baseline therefore remains useful evidence; it is not presented here as today's HEAD.

The attached hosted validation log records remote-gate exit 0. Independently queried GitHub run [37260842224](https://github.com/relux-works/curator/actions/runs/37260842224): completed/success, head `b1ae0a57a43aefaf87ed693e3095dee85ed1c2e5`. Git resolves that commit's tree to the exact candidate tree above; their tree diff is empty. The attached log reports Linux/macOS/Windows tests, lint, race and conformance success, with two optional jobs skipped. No test-case coverage ratio is inferred from a workflow conclusion. The historical five selected baseline tests are accepted as historical only; no completed implementation of #105 is claimed.

## Non-blocking bounds and disposition

F-R1-size is transparently recorded in E8.5: the mandatory full script bodies increase the artifacts beyond the earlier 80 KiB budget. The explicit later R1 instruction requires those bodies, so this evidence correction does not warrant another research rework or human approval gate. Size is reported, not passed off as within budget.

Publication patterns are a bounded checker, not proof of universal secret/hostname absence. Native color-key support, rule equivalence, negative Codex layering, higher-precedence conflicts and PM selection remain assigned to L1/L2 and the decision-ready questions. Accepting this draft does not adopt CIP-0006, authorize its implementation or satisfy runtime #105 acceptance.

The attached evidence resource was downloaded through the board CLI and is byte-identical to the candidate companion. All repository files remain unchanged by the reviewer. The run goal query reports no active goal (not goal-bound). No blocking finding remains. Persist this task-scoped verdict before accept_cr; leave checkpoint/integration to the tracked researcher/analyst producer.

## Persisted transition and runtime anomaly

accept_cr returned exit 0 / ok=true: revision 1 accepted, element integrating, producer role researcher / archetype analyst, reviewer run stamped. The command took several minutes and remained active until completion; no duplicate acceptance mutation was issued.

The transaction also returned a non-blocking write-boundary report under policy warn. It observed changes on other tasks, some attributed to other runs and some unattributed. This reviewer made no such edits: repository artifacts were only read, scratch files were outside the repository, and board writes used the task-scoped CLI lifecycle/resource/checklist operations. The report is retained here as an attribution anomaly, not evidence that the reviewer edited those paths. Acceptance persisted successfully; integration remains the next tracked producer operation.
