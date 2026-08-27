# BUG-260731-3a5q1p: curator-install-dryrun-rc6-multiproject-binding

## Description
internal/install TestAuthoritativeDryRunCasesMutateNothingPersistent fails against an rc.6 conformance root because the new authoritative dry-run case scope multi-project has no executable binding. Verified identically on Curator main cfffd7cd AND on BUG-260731-3gm8kc PR 9 head bd6ba08, so it is pre-existing and not introduced by the lifecycle-vector gate. It is currently invisible because Curator CI pins SPEC_PIN=00b1688 (rc.3) and internal/install is not in the curator-spec Implementations job set. It becomes a live Curator CI failure the moment SPEC_PIN advances to rc.6. Raised on the explicit recommendation of both independent reviewers of BUG-260731-3gm8kc (RUN-260731-4afbab follow-up b, RUN-260731-109b9b follow-up 7b), which also record that weakening the default branch to make it pass was considered and rejected.

## Scope
Curator internal/install authoritative dry-run case bindings against an rc.6 conformance root. Give scope multi-project a real executable binding or state a defensible reason it has none; do not weaken the default branch of TestAuthoritativeDryRunCasesMutateNothingPersistent and do not special-case the assertion to skip the new case.

## Acceptance Criteria
go test ./internal/install passes with CURATOR_CONFORMANCE_ROOT pointing at an rc.6 root, without weakening TestAuthoritativeDryRunCasesMutateNothingPersistent; advancing SPEC_PIN to rc.6 does not turn Curator CI red on this test.
