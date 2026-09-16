# Review verdict — TASK-260916-2fu85y revision 1

Verdict: CHANGES_REQUESTED. Route to to-dev for a corrected draft and another independent review.

Reviewed CR-TASK-260916-2fu85y-1, base 871d11bcdfd240a6260d0722503bdd1642a8fce8, exact candidate tree 57de445309cc74a8f062b74af5ad5a5f55c56ec0. git diff candidate -- was empty before and after inspection. No repository edits made.

## Findings

1. High: decisions/0018-curator-run-permission-interface.md:220–225, 268–270, 292–296 relax the explicit brief requirement that headless/tracked/CI must not inherit yolo implicitly. Counterexample: an untracked CI or headless launch with no flag, profile permission, global permission, or lock takes item 2's built-in yolo default. Item 5 expressly supplies no headless/CI protection and defers whether it is needed. The final statement scoping bypass to interactive launches therefore is not supported by the specified resolution rules. Required correction: retain the interactive operator-owned default, but specify no implicit bypass for headless/CI, with a conservative native/refusal rule and explicit admission boundary. Detection mechanics may remain future implementation work; the required safety invariant cannot be optional.

2. High: decisions/0018-curator-run-permission-interface.md:262–268 explicitly permits a legacy fragment lacking policy/lock support to resolve as silence, yielding yolo past an unenforced native lock. This defeats the requested force-native control and conflates unsupported policy transport with legitimate absent configuration. Required correction: make verified support for policy and lock transport an adoption precondition, and specify refusal when that support cannot be established. The exact version token can remain open; fail-open mixed-version deployment must not be the proposed contract.

These are draft rework within the approved brief, not human-only decisions or external blockers.

## Verified and bounded evidence

- Independently ran in zsh: PATH=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH make validate. Exit 0.
- Validator: 60 schemas and 1047 vector files validated. Python: 227 tests passed in 176.937 seconds. Go generate-vectors package passed from cache; no fresh Go execution claimed.
- git diff --check against base: exit 0.
- Native mapping table comparison against base: all 6/6 table lines identical, including all four environment rows.
- Inspected launcher SPEC sections 4.3 and 4.7 and environments sections 12.1–12.2: config ownership and v2 closed-schema rationale fit those contracts.
- Proposed/not-adopted status retained; global permissions, per-profile surface, precedence, source enum, operator-choice statement, CHANGELOG and Filed proposals amendment are present.
- Existing raw-native passthrough and refusal selector lists are preserved. Force-native continues to mean the typed mode, not a security perimeter against raw native argv, as required by unchanged native semantics.
- Adversarial checks were prose counterexamples against two policy paths, not executable permission-gate tests. This draft changes no runtime. make validate does not prove these new permission invariants; no mutation or runtime coverage claim is made.
- No producer validation evidence was substituted for the independent validation above.
- PR landing and issue #55 comment remain delivery obligations; neither is attested by this review.
- spawn goal reported this run is not goal-bound; no directives were recorded.
