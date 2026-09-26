# BUG-260920-3vfwch: hosted-macos-git-permission-denied-flake

## Description
Recurring hosted macos-latest flake: a test that shells out to git fails with fork/exec /opt/homebrew/bin/git: permission denied (seen on gate runs 35340496757 TestDryRunEffectBindingsSeeWhatARealOperationWrites, 35481906193 TestEndToEndInstall, 35510984798 TestGlobalDryRunPlansBuildsWithoutSessionOrPersistentState). Each occurrence costs a republish + ~60 min gate. Investigate: the Homebrew git on the hosted image is intermittently non-executable at test time (image race with brew upgrade? concurrent test rewriting PATH?), or the tests resolve git through a PATH that changes mid-run. Fix: resolve git once per test binary (exec.LookPath at init) from a stable location, or pin the lane to /usr/bin/git for tests, with evidence from a diagnostic step in the gate that records `ls -l $(command -v git)` when the failure recurs.

## Scope
test helpers that locate git (internal/testutil or the per-package helpers), .github/ci test-gate diagnostics

## Acceptance Criteria
the macOS lane no longer fails with git permission denied across 10 consecutive gate runs, or a diagnostic step proves the image-level cause and the tests bind git to a stable path; no test semantics weakened.
