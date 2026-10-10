# Review verdict — revision 1

Task: TASK-261010-1992si — research-project-surfaces-coverage.
Verdict: **changes_requested**; route to **analysis** for research candidate rework.
Base: `aec9e800db12669871e0815f9952228202fe3a4f`.
Candidate tree: `2793eb6e0b393d8c6495a6e694f5d47f5cf7281b`.

## Swept surfaces

| Review surface | Result |
|---|---|
| Brief question 1 | Five-environment inventory present, with pinned/latest distinctions and explicit unavailable-source bounds. Claude/Muse older-pin exhaustiveness remains unknown; OpenCode has no pin. |
| Brief question 2 | Sections 3–4 map each inventory class to discovery, closed normalization, admission, precedence, protected outputs and suppression. Proposals are distinguished from implementation. |
| Brief question 3 | Section 5 defines Skillfile/lock ownership and migration of gitignored repository installs. |
| Brief question 4 / architecture | Sections 6–7 compare alternatives and recommend a CIP-0002 amendment with a bounded context/knowledge preview slice. Strict launch remains refused pending separate qualification, consistent with the cited operator input. |
| Citation sample | 12 sampled claims supported by 11 distinct cited sources; eight sample rows concern baseline inventory/behavior. See below. No sampled citation contradicted its associated claim. |
| Exact candidate scope | FAIL: two added research files, including an unrelated sibling-task study. LOGBOOK.md is absent from the delta. |
| Hygiene / execution | No secrets, personal absolute paths or host names observed in reviewed delta. Only source/document reads and review lifecycle operations performed; no harness, tests or builds. Prior editorial audit counts are producer-reported, not independently rerun. |

## Findings

```json
{
  "findings": [
    {
      "id": "candidate-scope-contamination",
      "severity": "robustness",
      "severity_reason": "P1: violates the explicit single-study candidate boundary; acceptance would also attest unrelated task material. This is scope rework, not minor citation drift.",
      "blocking": true,
      "repeat-of": "none",
      "mechanism": ".research/261010_modular-instructions-design.md:1 is an additional 242-line study for TASK-261010-2uqd3t \u2014 research-modular-agents-md, included in the exact base-to-candidate delta.",
      "evidence": "git diff --name-status aec9e800db12669871e0815f9952228202fe3a4f 2793eb6e0b393d8c6495a6e694f5d47f5cf7281b lists two added files, rather than only the required project-surfaces study.",
      "requested_change": "Rebuild the task candidate against the appropriate checkpoint so its delta contains only .research/261010_project-surfaces-coverage.md. Preserve the sibling study through its own task workflow; do not delete shared work merely to narrow this candidate. Publish a new revision for review."
    }
  ]
}
```

## Independent citation checks

Each source below was fetched read-only with `gh api repos/OWNER/REPO/contents/PATH?ref=COMMIT`, requesting raw content. All source fetches succeeded. These checks sample claims; they do not certify exhaustive inventories or runtime isolation.

| Citation | Claim checked | Supporting location | Result |
|---|---|---|---|
| [c-reg](https://github.com/relux-works/curator/blob/1d7eb18c24c4f156730f5c14aa4790a8c86ed891/internal/envregistry/envregistry.go) | Recorded releases are Claude 2.1.261, Codex 0.153.2, Pi 0.84.2, Muse 1.4.1-R4503.1; OpenCode is empty. | VerifiedRelease entries, lines 238–335 | Supported |
| [cx-doc-pin](https://github.com/openai/codex/blob/657a993cbee87acf52d14b758ce49dbd46d1b8eb/codex-rs/core/src/agents_md.rs) | Instruction selection orders override, AGENTS, fallback names; project chain is root to CWD. | candidate_filenames and agents_md_paths, lines 185–279 | Supported |
| [cx-skills-pin](https://github.com/openai/codex/blob/657a993cbee87acf52d14b758ce49dbd46d1b8eb/codex-rs/ext/skills/src/host_roots.rs) | Skill roots enumerate all config layers and separately scan repository .agents/skills. | all_layers_high_to_low at line 80; repo_agents_skill_roots | Supported |
| [pi-loader](https://github.com/earendil-works/pi/blob/914cf1472e715297caa30db4b9535d534a9eb718/packages/coding-agent/src/core/resource-loader.ts) | Context filename precedence includes override and uppercase variants. | loadContextFileFromDir, line 72 | Supported |
| [pi-loader](https://github.com/earendil-works/pi/blob/914cf1472e715297caa30db4b9535d534a9eb718/packages/coding-agent/src/core/resource-loader.ts) | Disabling context removes home/project context together; system files choose trusted project before global. | loadProjectContextFiles; lines 515–521 and 1022–1047 | Supported |
| [pi-main](https://github.com/earendil-works/pi/blob/914cf1472e715297caa30db4b9535d534a9eb718/packages/coding-agent/src/main.ts) | Startup session lookup constructs a project-reading settings manager before later runtime trust configuration. | startupSettingsManager line 658 versus runtimeSettingsManager line 737 | Supported |
| [pi-usage](https://github.com/earendil-works/pi/blob/914cf1472e715297caa30db4b9535d534a9eb718/packages/coding-agent/docs/usage.md) | Baseline has no built-in MCP or permission popups. | line 304 | Supported |
| [pi-skills](https://github.com/earendil-works/pi/blob/914cf1472e715297caa30db4b9535d534a9eb718/packages/coding-agent/docs/skills.md) | Explicit skills survive --no-skills; Pi roots allow standalone Markdown. | lines 25–41 | Supported |
| [oc-instructions](https://github.com/anomalyco/opencode/blob/53d1eabb61e21162157817bf677da0a4ad3332e3/packages/opencode/src/session/instruction.ts) | V1 startup prefers AGENTS, CLAUDE compatibility, CONTEXT; nested resolver is separate. | instructionFiles lines 64–68, systemPaths and resolve | Supported |
| [cl-new](https://github.com/anthropics/claude-code/blob/2301018b1f61073c501a8e7a4813ef48c239163b/CHANGELOG.md) | AGENTS support starts at 2.1.277; 2.1.288 fixes Write/Edit activation of scoped rules. | 2.1.277 entry and line 672 under 2.1.288 | Supported |
| [pi-new-mcp](https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/coding-agent/docs/mcp.md) | Latest supports project MCP after trust, partial server overrides and command-valued environment/header fields. | lines 32–36 and 81 | Supported |
| [cip](https://github.com/relux-works/curator-spec/blob/2f0531b4edcc99c6118c277deb00e8736392c04d/cips/CIP-0002-project-context-in-managed-launches.md) | Operator input retains skill declaration in Skillfile and external admitted project layer. | Operator input, lines 9–19 | Supported |

No failing sampled citations. The Codex execution-policy source was also inspected, but disabled-layer filtering depends on stack behavior beyond that file; that subclaim is not counted as independently verified here. Vendor rolling documentation and remaining citations were not exhaustively revalidated.

The exact diff itself is the scope reproduction; no executable regression test is appropriate for this research-only correction. The next reviewer should repeat the exact candidate path comparison and assess any changed study content.

Goal check: `task-board spawn goal` reported no active goal for this run. LOGBOOK.md remains unchanged under the campaign rule. This artifact and the task note carry the findings.
