# BUG-260922-3v8k23: ubuntu-install-package-timeout-budget

## Description
Hosted ubuntu-latest lanes leave too little headroom for internal/install under the 30m per-package GO_TEST_TIMEOUT (.github/workflows/ci.yml lines 216/337/705; .github/ci/test-gate.sh default 30m). Baselines from gate run 35709048693 (TASK-260916-1xib1x rev3): Test(ubuntu) internal/install 926 s (51 % of budget), Race(ubuntu) internal/install 1115 s (62 %), Windows 1929 s of its 120m budget (27 %, raised by 3ux95w for the same reason). Gate run 35713206841 (rev4, byte-identical to rev3) failed Test(ubuntu) with FAIL internal/install 1800.165s: the evidence timeline shows every parallel test in the package taking 2.1x the rev3 time (TestPlanLinesRedactAnUntrustedReason 761 s -> 1600 s; TestScriptAuditLabelsAtInstallEntry 835 s -> 1778 s) and the global-install rows still passing one by one (1732 s, 1799.7 s) when the budget expired: a slow runner, not a hang. Cost: one full republish cycle per occurrence. Fix: give the ubuntu/macOS lanes the same budget class as Windows (e.g. 60m) or reduce the internal/install wall time (shared CLI build per package, fewer full-pipeline rows), and pin the chosen expression in the gate self-test so it cannot silently regress; keep the per-package timeout semantics (a genuine hang must still fail).

## Scope
CI budget only: .github/workflows/ci.yml GO_TEST_TIMEOUT expressions, .github/ci/test-gate.sh default, gate self-test pins, docs/ci notes; no product code, no test-semantics changes beyond wall-time reduction.

## Acceptance Criteria
1) Test(ubuntu) and Race(ubuntu) keep at least 2x headroom over the measured internal/install baseline (evidence: two consecutive green gate runs with elapsed/budget ratio <= 50 %); 2) gate self-test pins the new timeout expressions for every lane; 3) a synthetic hanging package still fails the gate within the budget (negative row); 4) CHANGELOG entry.
