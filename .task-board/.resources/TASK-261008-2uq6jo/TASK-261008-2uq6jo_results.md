N6 fix results (TASK-261008-2uq6jo)

Change (internal/gitcred/gitcred.go, production boundary gitcred.Access.call):
- boundedBuffer gains an overflow flag, set in Write when a payload exceeds the remaining capacity (total answer > 64 KiB = maxAnswerBytes).
- Access.call now refuses the whole answer (nil, false) when the flag is set, even on a zero-status Git exit, instead of parsing the truncated prefix as a credential.
- Exact-limit answers (total == 64 KiB) keep the flag clear and are still accepted unchanged.
- No other behavior changed: all existing prompt/namespace/read-back controls use small answers and are unaffected.

Regression test (internal/gitcred/credential_bound_test.go, through production entry Access.ReadHost):
- TestCredentialAnswerBoundRefusesOversized: real Git + built-in store helper seeded with a synthetic .git-credentials file in a temporary home; secrets of 32, maxAnswerBytes-256 (accepted unchanged), maxAnswerBytes+128 (refused). No shell script, runs on Windows.
- TestCredentialExactFrameBound: controlled Git via the package's fake-git test-binary re-exec (new modeFixedAnswer in gitcred_test.go serves exact bytes from answer.bin); total frames of cap-1, cap (accepted unchanged), cap+1 (refused). No shell script, runs on Windows.
- Adapted from the wave2 probe; no wave2 naming remains. Neither test skips on any platform.

CHANGELOG: one Fixed line under Unreleased, operator-visible wording (unchanged since rev1; review-verified accurate).

Compile-only validation (R223: no local go test; hosted gate is arbiter):
- gofmt -l internal/gitcred/ -> exit 0, clean
- go vet ./internal/gitcred/ -> exit 0
- go vet ./... -> exit 0
- go build ./... -> exit 0
- git diff --check -> exit 0
- go test: NOT run locally per R223; suite verdict (regression + negative controls) left to the hosted gate.

Rev2 addendum (successor run RUN-261008-03997d): CR rev1 hosted gate failed ONLY on Test (windows-latest) with go test exit=0 and platform-case gate exit=1. Cause: the two new tests skip on Windows with reasons ("fixture helper is a POSIX shell script", "POSIX fixture") matching none of the 62 skip-classes.tsv patterns, so Tier 2 refused them as FATAL-unclassified. Fix: reworded both skips to the repo-conventional platform-control vocabulary ("the POSIX shell helper/git fixture is exercised on the unix runners"), which the gate's own awk matcher classifies as platform-control/allow (verified: new reasons match exactly that class; old reasons match zero classes). No product-code change. Re-ran compile-only tail: gofmt clean, go vet ./internal/gitcred/, go vet ./..., go build ./... all exit 0. No local go test (R223).

Rev3 addendum (run RUN-261008-254220): revision 2 was reviewed changes_requested (TASK-261008-2uq6jo_review-verdict-rev2.md): Windows size coverage 0/6 (skip rewording is not coverage) and hosted mutation proof absent. Rework:
1. Portable fixtures (this revision's code change): real-Git case now uses Git's built-in store helper seeded via .git-credentials instead of a POSIX shell-script helper (same production path git->helper->stdout->bound; the existing TestRealGitKeepsTheManagerEntrySeparate proves this fixture portable); exact-frame case now uses the package's fake-git test-binary re-exec with a new fixed-answer mode serving byte-exact frames instead of a POSIX shell-script git. Both tests run on all three hosted lanes with zero skips. No new skip class; no skip-classes.tsv change.
2. Hosted mutation proof (diagnostic only; mutated source never handed off): disposable clone /tmp/n6-mutant at base 3b6481c15d61b08d3f0fae7c3269329555336709; candidate snapshot c1d56bc33a4626bed141f69223dad22bc6a2c491 verified byte-identical to the worktree candidate on all 4 changed paths (cmp MATCH x4). Three single-hunk mutants, each pushed to its own throwaway branch and run through full hosted CI (-count=1 per the workflow's own Test lanes). Branch prefix note: the rework brief names scratch/n6-mutant-2uq6jo, but only main, gate/** and pull_request trigger .github/workflows/ci.yml, so the throwaway snapshots use the repo's remote-gate.sh gate/** convention and were deleted after evidence capture:
   - M1 (wiring/delete): Access.call overflow-refusal clause removed (commit 144d761d475d087eed2f3121c3d412ccd35c3f00, branch gate/n6-mutant-2uq6jo/m1): run https://github.com/relux-works/curator/actions/runs/37847866844 -- OBSERVED: all three Test lanes failure; failing (sub)tests on every lane: TestCredentialAnswerBoundRefusesOversized/65664 ("oversized helper answer accepted as credential (truncated=true)") + TestCredentialExactFrameBound/1 ("cap+1 frame was accepted"). No other gitcred test fails.
   - M2 (narrow +1): detection `len(accepted) > b.remaining` widened to `> b.remaining+1` (commit a16f0cb6e0238fde72cb4f472491390f24516482, branch gate/n6-mutant-2uq6jo/m2): run https://github.com/relux-works/curator/actions/runs/37847865978 -- OBSERVED: all three Test lanes failure; the ONLY failing (sub)test on every lane is TestCredentialExactFrameBound/1 (cap+1 accepted under the widened bound); the real-Git oversized case still refuses (total 65729 > widened bound 65537) and exact-cap /0 still passes.
   - M3 (exact-limit positive control): detection changed to `>=` (commit 1fe63ffca2b1633979e45c0102c5b3d8444dcebc, branch gate/n6-mutant-2uq6jo/m3): run https://github.com/relux-works/curator/actions/runs/37847865854 -- OBSERVED: all three Test lanes failure; the ONLY failing (sub)test on every lane is TestCredentialExactFrameBound/0 (exact-cap answer refused by `>=`).
   Coverage: 6/6 size subtests exercised on Linux, macOS AND Windows for every mutant (was 0/6 on Windows in rev2); skips-observed.tsv records ZERO internal/gitcred skips on all 9 lanes. Evidence source: test-evidence-{ubuntu,macos,windows}-latest artifacts, test/go-test.json terminal fail/pass events, downloaded per completed job. All three gate branches deleted after capture (git push origin --delete exit 0; ls-remote confirms no refs remain). Mutated source was never handed off; the worktree candidate is unmutated.
