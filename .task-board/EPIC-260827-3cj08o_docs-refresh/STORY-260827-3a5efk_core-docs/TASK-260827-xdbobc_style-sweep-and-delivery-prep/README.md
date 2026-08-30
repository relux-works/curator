# TASK-260827-xdbobc: style-sweep-and-delivery-prep

## Description
After all content tasks are accepted: sweep README.md and every docs/*.md against docs/prose-style.md and its blacklist; fix violations in place with file:line evidence in the outcome; verify all internal links resolve; confirm no .task-board or unrelated files are staged in the final diff list you produce for the orchestrator (list the exact docs-scope files).

## Scope
All shipped markdown; fixes in place; no git operations.

## Acceptance Criteria
Zero blacklist hits outside labeled negative examples; links resolve; exact delivery file list in the outcome.
