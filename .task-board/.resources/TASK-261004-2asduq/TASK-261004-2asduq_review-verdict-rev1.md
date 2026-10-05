# Review verdict: accepted

Task: TASK-261004-2asduq — launch-command-environment-fragment-design.
Change Request: CR-TASK-261004-2asduq-1, revision 1.
Reviewed base: 54bed271b7609bf206a04369202473c430d0d96a.
Reviewed candidate: a7d393fa053120a9084eed201c984a283919ae22.
Reviewer role: reviewer. Reviewer goal query returned no active goal (not goal-bound).

## Verdict and scope

Accepted as a research CIP draft and companion evidence. No blocking findings or requested changes. This is not operator acceptance of the design, implementation authorization, scheduling, production conformance, or landing. Route through accept_cr to integrating; the researcher/analyst producer owns subsequent integration.

## Swept surfaces

| Surface | Result and evidence |
|---|---|
| Binding scope and format | Pass. CIP lines 1–19 preserve draft/design-pending status, requested template and project context requirements; lines 195–210 defer every implementation leaf until acceptance. |
| Five native environments | Pass. CIP native inventory and managed-behavior tables cover claude_code, codex_cli, opencode, pi and muse, including context/rules/knowledge, MCP, permissions, skills, commands and native settings. Documentation, measured probes and unknown enforcement are explicitly separated. |
| Decision readiness | Pass. CIP lines 57–65 compare three real home strategies; lines 67–120 give defaults, per-item approval, composition/precedence, leases and login cost. Lines 230–238 supply seven decision-ready questions with recommended answers. |
| Hostile repository and command transport | Pass. CIP lines 122–171 specify one protected dispatcher, destination append, tracked refusal, immutable admission, TOCTOU/replay controls, secret references and the same-UID protection limitation. Append is expressly not integrity. |
| Spec and implementation/test leaves | Pass. CIP lines 181–228 give normative sketches, version compatibility, ordered leaves, production entry points, negatives, narrowing mutants and platform coverage. First slice and adapter qualifications remain proposed. |
| Source evidence | Pass. Independently read 11 pinned source ranges listed below; every read exited 0 and matched the cited claim. rc.14 peeled tag independently equals 43bf0a2506d5c354a73bbc3ea4623d4653db10c7. |
| Probe bounds | Pass within reported scope. Evidence P1–P6 includes versions, synthetic fixtures, environment isolation, native command lines, child exit results and explicit limits. P2/P3 distinguish instructions from trusted configuration; P4 explicitly declines MCP enforcement proof; P5 has positive/negative extension trust results. Original native probe results are accepted from attached evidence, not claimed as independently rerun. P6 independently rerun: 2/2 cases, exit 0, BASE then DISPATCH as reported. |
| Safety/publication | Pass for inspected artifacts and diff. No credential or account file accessed by this reviewer; no login/logout, inference or external messaging. Research reports synthetic homes and no account access. No public personal paths, internal hostnames or employer/client identifiers found on inspection. CLAUDE.local.md/settings.local.json are native filenames, not internal hosts. Safety of the earlier process is supported by its bounded evidence statement, not retroactively independently observed. |
| Diff hygiene | Pass. Exact candidate changes only the two requested .research Markdown files (238 + 211 lines). Candidate bytes equal workspace bytes. LOGBOOK.md has no delta; no product change. Git diff --check exits 0; JSON example parses; both documents have zero trailing-whitespace faults. Reviewer changes no repository file. |

## Independently spot-checked pinned citations (11/11)

- Curator ca1b776: internal/envprofile/managed.go:63–85 — profile/environment home and separate LaunchDir.
- Curator ca1b776: internal/envfragment/envfragment.go:23–70 — v2 default, Muse v3, emitted members.
- curator-spec rc.14 43bf0a2: protocol/environments.md:2583–2606 — machine-current shims and managed command limitation.
- Curator ca1b776: internal/envprofile/managed.go:2290–2312 — conditional MCP member.
- launcher d092035: internal/composition/composition.go:74–139 — conditional MCP flags, owned literals/name lookups.
- launcher d092035: internal/mapping/mapping.go:18–32 — Claude/Codex/Pi mappings only.
- launcher d092035: internal/fragment/fragment.go:324–333 — v1/v2 parser admission.
- launcher d092035: internal/fragment/fragment.go:493–506 — old two-level prepend-root assumption.
- curator-spec rc.14 43bf0a2: protocol/environments.md:3226–3233 — nonempty MCP set condition.
- Curator ca1b776: internal/envfragment/envfragment.go:141–154 — canonical fragment digest bytes.
- curator-spec rc.14 43bf0a2: decisions/0017-environment-credential-modes.md:113–136 — credential-sharing scope and unsupported Claude/macOS shared mode.

## Documentation spot checks

Primary pages retrieved and checked on 2026-10-05: [Claude memory](https://code.claude.com/docs/en/memory) supports instruction discovery and AGENTS minimum 2.1.277; [Codex skills](https://developers.openai.com/codex/skills/) supports repository .agents/skills scanning; [OpenCode config](https://opencode.ai/docs/config/) supports custom-before-project merging; [Muse configuration](https://dev.meta.ai/docs/muse-code/configuration) supports first-file-per-level instruction order, trust and project memory. These checks establish documented behavior only, consistent with the draft; they do not certify native isolation.

## Verification limits and checklist interpretation

No go build/go test run under hosted-evidence instructions. No product suite claimed green. For this research-only task, implementation matches AC means the two documents satisfy the research acceptance criteria; tests green means the applicable document/source checks and independently rerun synthetic PATH probe passed. Native P1–P5 and earlier document audits are recorded evidence accepted within their stated limits, not new reviewer execution. Full adapter enforcement, live MCP, login reuse, tracked transform and lease races remain explicitly unproven, with qualification assigned to future implementation leaves. No test absence is treated as production proof.
