# TASK-260922-18ex37: move-spec-pin-to-current-main-with-gaps

## Description
Move SPEC_PIN from dced9b8 (v1.0.0-rc.12) to the chosen current curator-spec commit and fill the gap ledger from the REAL failures the new root produces, each row attributed to the board element that owes it (E1 -> STORY-260916-ioemse, E3 -> STORY-260916-1i1gfo, E5 -> STORY-260916-73a5zg, E6/path-kind -> STORY-260916-wgt8vz, S1/S3 -> STORY-260910-2qmrb8 or 234vmx, S2 -> STORY-260910-6bo7ej, S5 -> STORY-260910-148pj1, R3/P2 -> STORY-260910-25yc0h, 8.4.1 -> STORY-260916-1ll22r, 0017 manager cases -> STORY-260922-1cenbr, dotfile table -> TASK-260918-bi6ouz; verify each mapping against the spec commit that published the case rather than trusting this list). Depends on the ledger leaf.

## Scope
curator: SPEC_PIN in .github/workflows/ci.yml, the gap ledger rows, any test that pins the pin value, docs/CHANGELOG. No spec change; no implementation of the gapped behaviour in this leaf.

## Acceptance Criteria
1) SPEC_PIN equals the chosen spec commit and the hosted matrix is green with every new failure either implemented or an attributed gap row; 2) no gap row without an owner; 3) the cases that curator ALREADY implements at the new root are driven, not gapped — prove the counts per family before and after; 4) the dotfile-manager, marker-v5 and 0017/0018 vectors are now published and their status is stated explicitly; 5) CHANGELOG entry naming the new pin and the gap count
