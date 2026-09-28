# TASK-260921-3qcjsy: adopt-decisions-0017-and-0018-in-curator-spec

## Description
Turn decisions 0017 and 0018 from proposed into adopted in curator-spec. For every option set and open question pick the option the draft recommends; where the draft records no recommendation choose the fail-closed / least-privilege option and list the choice explicitly in results.md for operator confirmation. Apply the amendments to the normative text the decisions target (environments credential modes; curator run permission interface incl. the 2026-09-16 amendment: config-driven mode, flag override, default yolo) or, where an amendment needs implementation-side work, record it as a named follow-up leaf per repository. Keep the documents' compatibility/security-impact sections truthful.

## Scope
curator-spec decisions/0017-environment-credential-modes.md, decisions/0018-curator-run-permission-interface.md, amended normative sections, decision index

## Acceptance Criteria
Status: adopted in both documents with a dated adoption note; every open question resolved to a recorded choice; amended normative text consistent and make validate green; follow-up implementation leaves named
