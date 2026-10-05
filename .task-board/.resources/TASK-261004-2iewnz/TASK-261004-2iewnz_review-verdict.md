# Review verdict: changes requested

Task: TASK-261004-2iewnz — research-105-design-and-implementation-plan.
Review date: 2026-10-05. Role: reviewer. Route: analysis (research evidence rework).

## Blocking finding R1 — reproduce the claimed Python checks

The binding research-review-note.md requires every verified or measured claim to have a reproducible probe in the evidence file. The companion `.research/261004_CIP-0006-legacy-provider-settings-and-mcp-optouts_evidence.md` does not supply the programs behind these successful-exit claims:

- Line 142: “A separate Python assertion process exited **0**” for no profile/config/surface publication and the positive-control selection. Neither the assertion body nor exact discovery paths are included.
- Line 176: “A separate Python assertion exited **0**” for unchanged config after both config refusals. The assertion body is missing.
- Lines 196–198: “a standalone `python3 -` inline artifact checker, exit **0**” and the 10/10 sections, 38/38 source-link ranges, 5/5 JSON examples, 2/2 changed files and publication-pattern checks. `python3 -` without its input is not a reproducible command. The source/range enumeration and privacy patterns are unavailable.

These are useful bounded claims, but a subsequent reviewer cannot reproduce the reported checks as written. Supply the exact sanitized script bodies in the companion evidence, including fixture construction and configuration snapshots used by the mutation assertions. Specify the pinned source/spec inputs and expected counts. Preserve historical results as historical; do not invent a recovered original script or claim a replacement was the script previously run. If the original code is unavailable, label a replacement reproduction accurately or downgrade/remove the unsupported assertion. This is an evidence-documentation correction, not a request to run local Go gates in hosted-evidence mode. No change to the recommended design is required by this finding.

## Swept surfaces

| Surface | Result |
|---|---|
| Issue scope / operator decisions | Held: all nine provider paths plus MCP are mapped; Decision 0014 remains proposed, Decision 0018 remains independent. |
| Current implementation / spec citations | Held: 14 citation groups checked as listed below. |
| Options / recommendation / control surface | Held: four alternatives, concrete typed record, defaults, presence, precedence and scope semantics. |
| Migration / ownership / launch | Held at draft level: acquisition, drift, conservative release, old seeds, negative-only MCP, compatibility and downgrade described. Native enforcement is explicitly unproved and assigned to L2. |
| Spec changes / implementation leaves / tests | Held at draft level: L1–L8 with sizes/dependencies, entry-point negatives and narrowing mutants. No claim of completed feature implementation. |
| Reproducibility | Not held: R1 above. |
| Safety / publication | Held within observable bounds: reviewed documents contain no evident secrets, personal paths or internal hosts; bounded pattern scan returned zero matches. No credentials or native account state were accessed by this reviewer. Prior researcher no-secret assertions are not independently attested by document inspection. |
| Diff hygiene | Held: exactly two untracked research documents; tracked and staged diffs empty; LOGBOOK.md untouched. Reviewer changed no repository file. |

No additional blocking finding emerged from the remaining surface sweep. This is one consolidated review round.

## Independent checks and source identity

Fresh `git ls-remote` reports main `54bed271b7609bf206a04369202473c430d0d96a`, equal to worktree HEAD/main. The author's `ca1b776fb580ec0cee0173bf150daf063023aeaa` is a dated research baseline, not today's head. A Git diff of the seven central cited implementation files below from that baseline to current HEAD is empty. No refresh is requested merely because unrelated main changes landed.

Read issue #105 directly with `gh issue view 105 -R relux-works/curator --json title,body`. Read pinned source using `git show ca1b776fb580ec0cee0173bf150daf063023aeaa:PATH` and pinned spec using `git show 43bf0a2506d5c354a73bbc3ea4623d4653db10c7:PATH` in the respective repositories. Spot-checks:

1. contextpkg.go:168–255 — manifest keys exclude settings and policy.
2. config/environments.go:270–275 — permission enum native/yolo.
3. cmd/curator/envconfig.go:58–102 — JSON decoding fallback and mutation entry.
4. envprofile/overlays.go:15–78 — per-profile machine overlays.
5. envprofile/managed.go:2389–2437 — named/env-current/machine-current selection.
6. contextmaterialize/mcp.go:95–127 — environment filtering and name validation.
7. contextmaterialize/mcp.go:214–238 — empty set omits file.
8. envregistry/envregistry.go:28–37 — shipped seed selector A.
9. envmarker/envmarker.go:307–315 — closed four-surface key set.
10. config/config.go:25–35 — machine schema versions 1–3.
11. decisions/0014-tool-configuration-surfaces.md:1–89 — proposed status and unselected alternatives.
12. decisions/0018-curator-run-permission-interface.md:175–240 — adopted launch-mode surfaces and native lock.
13. protocol/environments.md:1826–1857 — Claude strict MCP versus Codex base-layer residual.
14. protocol/environments.md:3672–3710 — closed machine knobs.

Implementation paths above are relative to internal/ unless cmd/ is given. The seven unchanged central paths checked were contextpkg/contextpkg.go, config/environments.go, cmd/curator/envconfig.go, envprofile/managed.go, envregistry/envregistry.go, envmarker/envmarker.go and contextmaterialize/mcp.go.

The reviewer also opened the official Claude [permissions](https://code.claude.com/docs/en/permissions) and [settings](https://code.claude.com/docs/en/settings) pages as documentation corroboration only; this does not qualify pinned native execution.

Artifact inspection measured 67,120 combined bytes, zero trailing-whitespace lines, and zero matches for personal home path prefixes, PEM private-key headers, GitHub-token and sk-token patterns. This bounded scan is not proof of all possible secret absence. Both tracked and staged Git diffs were empty; `git status --short` lists only the two submitted research files.

No build, Go test, native provider or scratch runtime probe was rerun by this reviewer under hosted-evidence mode. The author's five passing baseline tests and explicit CLI probe results were reviewed as attached historical evidence only; the unavailable Python assertion bodies are the reason acceptance is withheld. Two reviewer source-reading attempts failed (range overflow, then a syntax error); the corrected read succeeded and supplied the spec checks above. Failed reads count as no evidence.

`task-board spawn goal` reports this run is not goal-bound. Record this task-scoped outcome before routing to analysis. The next researcher should address R1, hand off a new revision, and obtain another reviewer cycle. No human-only blocker or approval request is needed.
