# BUG-260730-3eqseq: repair-curator-v0-13-main-ci

## Description
Investigate and repair the seven failing ordinary CI checks on push commit cfffd7cd7be33aba9ca44d26993ce5ab19b5fa4d after the successful v0.13.0 release workflow. Evidence run: https://github.com/relux-works/curator/actions/runs/30493882177. Failures shown include interop conformance, lint, macOS/Linux race, and macOS/Linux/Windows test jobs.

## Scope
Curator CI only. Preserve the already published v0.13.0 release and accepted compiled-build behavior. Start after the CocoaSkills Go parity RC gate as explicitly prioritized by the human.

## Acceptance Criteria
Root causes are independently reproduced, fixes are reviewed and landed, and the full ordinary CI workflow is green on macOS, Linux, and Windows without weakening gates or rewriting the v0.13.0 tag.
