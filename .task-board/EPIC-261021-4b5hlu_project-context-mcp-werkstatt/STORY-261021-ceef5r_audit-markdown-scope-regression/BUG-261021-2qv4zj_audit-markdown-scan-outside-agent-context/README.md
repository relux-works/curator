# BUG-261021-2qv4zj: audit-markdown-scan-outside-agent-context

## Description
The audit network-declaration detector scans every Markdown file in the snapshot (since PR 151), so citations in .research, board resources and build-root vendor READMEs block task-board install. Scope it to core section 3 agent-facing context.

## Scope
curator internal/audit detector scope and tests; no spec change.

## Acceptance Criteria
1. Markdown is scanned only in core section 3 agent-facing context, minus core exclusion patterns and runtime/build root subtrees, reusing the existing eligibility code.
2. Regression: .research, .task-board resources and build-root vendor READMEs give no finding.
3. Positive controls: SKILL.md, references and assets URLs still flagged; declared host passes.
4. CIP-0021 and network-contact vectors stay green; narrowing mutants killed.
5. Full suite green.
