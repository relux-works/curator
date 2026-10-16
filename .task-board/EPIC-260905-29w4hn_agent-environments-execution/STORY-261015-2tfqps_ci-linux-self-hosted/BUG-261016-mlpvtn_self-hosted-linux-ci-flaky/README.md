# BUG-261016-mlpvtn: self-hosted-linux-ci-flaky

## Description
After Linux CI moved to the self-hosted Linux runners, main went red on Test (ubuntu-latest) with different tests each time: (1) Go link failures "relocation target golang.org/x/sys/unix.Syscall not defined for ABI0 (but is defined for ABIInternal)" in TestInstallGitignoreEntryPoint, the signature of a build cache shared by concurrently running runners and several Go toolchains (1.25.5 / 1.26.0 / 1.27.1 matrix); (2) "go-v1 process_timeout: Go probe exceeded its deadline" in TestReviewProbeLegacyLaneIdentityInvalid while the host load average reached about 50 on 16 cores with other repositories also scheduled there. Mitigation applied 2026-10-16: repository variable LINUX_SELF_HOSTED=false, so Linux jobs run on hosted runners again.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
1. Each self-hosted runner uses its own GOCACHE and GOMODCACHE (or a toolchain-keyed cache) and the job cleans or isolates them; a known-bad shared-cache plant reproduces the ABI0 failure and the fix prevents it. 2. Runner concurrency on the host is bounded so curator jobs do not compete with other repositories beyond the test deadlines; or deadline-sensitive tests are qualified for a loaded host. 3. Three consecutive green main runs on the self-hosted pool before LINUX_SELF_HOSTED is set back to true.
