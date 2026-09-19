# TASK-260919-3ux95w: raise-windows-go-test-budget-to-120m

## Description
After STORY-260910-1s75e1 (TASK-260910-3eu4cy: transactional publication with a failure-at-every-target-class sweep) the Windows internal/install package runs at 97.7% of the 60m GO_TEST_TIMEOUT (gate run 35424565415: 3516 s; baseline-speed extrapolation ~79 min; rev2 run 35418260010 timed out at 60m). Every other push to main and every later Story gate would time out on Windows. Raise the Windows budget in .github/workflows/ci.yml (line 213 test job and line 702 candidate suite): runner.os == Windows -> 120m, others keep 30m; update the comment above line 213 with the measured numbers. A hang stays bounded and fatal. Reviewer F-W1 in TASK-260910-3eu4cy_review-verdict-rev3.md.

## Scope
.github/workflows/ci.yml GO_TEST_TIMEOUT expressions (test job and candidate suite) and the budget comment; gate-selftest rows that pin the expression if any; nothing else

## Acceptance Criteria
Windows GO_TEST_TIMEOUT is 120m in both the Test job and the Candidate suite job; non-Windows stays 30m; the comment states the measured internal/install cost (3516 s on run 35424565415, ~79 min baseline extrapolation) and that a hang is still bounded and fatal; gate self-test (if it checks the expression) updated and green; hosted gate green.
