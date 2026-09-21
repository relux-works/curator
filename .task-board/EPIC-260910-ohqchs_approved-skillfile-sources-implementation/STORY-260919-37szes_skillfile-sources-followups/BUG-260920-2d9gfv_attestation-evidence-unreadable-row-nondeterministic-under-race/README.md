# BUG-260920-2d9gfv: attestation-evidence-unreadable-row-nondeterministic-under-race

## Description
Seen on gate run 35486280770 (BUG-260916-2f3xbf rev2, unrelated change): Race (ubuntu-latest) failed internal/crossconformance TestDraftSourcesSemanticCases/attestation-evidence-unreadable: fresh install errors = [every trusted audit registry served a tampered snapshot], want 'is not audited by any trusted registry'; the same row passed on the ubuntu/macos/windows test lanes of the same run and in the TASK-260910-1xya7x gates. Either the harness fixture for unreadable evidence is timing-sensitive under -race or the product classifies unreadable vs tampered evidence nondeterministically — determine which at the production entry and fix the root (a nondeterministic refusal class is a product defect).

## Scope
internal/crossconformance attestation evidence rows; internal/audit or registry evidence classification if the product is at fault

## Acceptance Criteria
attestation-evidence-unreadable is deterministic under -race across 20 consecutive local runs and on the hosted race lanes; if the product classified nondeterministically, the classification is fixed with a production-entry test and a narrowing mutant.
