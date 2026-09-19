# TASK-260919-3ux95w results — Windows go test budget 60m -> 120m

## Change (only file: .github/workflows/ci.yml)
- Test job GO_TEST_TIMEOUT (line 216): runner.os == Windows -> 120m, others 30m (was 60m/30m)
- Candidate suite GO_TEST_TIMEOUT (line 705): same 120m/30m
- Budget comment above Test job rewritten with measured numbers: internal/install 3516 s of 60m on gate run 35424565415 (97.7%), rev2 run 35418260010 timed out at 60m, baseline-speed extrapolation ~79 min; keeps "A hang is still bounded and still fatal."
- Candidate-lane comment: "the same 60m as the default lane" -> "the same 120m as the default lane" (one-word consistency fix; expression there also changed)
- Untouched, correctly: line 337 GO_TEST_TIMEOUT: 30m is the rose-air self-hosted macOS lane (fixed 30m, no Windows branch) — non-Windows stays 30m per AC

## gate-selftest grep (pins the expression?)
- grep -n 60m|30m|120m|GO_TEST_TIMEOUT .github/ci/gate-selftest.sh -> rows at 684,704,706,707,711,720-722
- Verdict: NO literal pin. The self-test asserts relationally: windows budget > unix budget (line 727) and unix budget == test-gate.sh default (line 734). No self-test edit needed; 120>30 holds and unix stays 30m.

## Evidence (all exit 0, run in Story worktree)
- bash .github/ci/gate-selftest.sh -> 187 passed, 0 failed, exit 0
- budget rows: ok every windows test-gate.sh lane declares GO_TEST_TIMEOUT; ok the windows per-package budget is larger than the unix one; ok the unix per-package budget still matches the gate default
- bash -n .github/ci/gate-selftest.sh -> OK
- ruby YAML.load_file(.github/workflows/ci.yml) -> YAML-OK (pyyaml and yamllint not installed on host)
- git diff --stat: 1 file changed, 9 insertions(+), 6 deletions(-), confined to .github/workflows/ci.yml

## Checklist
- [x] Test job Windows GO_TEST_TIMEOUT = 120m, others 30m
- [x] Candidate suite Windows GO_TEST_TIMEOUT = 120m, others 30m
- [x] Budget comment states 3516 s / run 35424565415 / ~79 min + bounded-and-fatal
- [x] gate-selftest grep reported; no pin rows; suite green (187/0)
- [x] YAML valid; shell syntax valid; diff confined to ci.yml
- [ ] Hosted gate green — reviewer-side (needs landing/candidate run showing GO_TEST_TIMEOUT: 120m in Windows Test job log)
