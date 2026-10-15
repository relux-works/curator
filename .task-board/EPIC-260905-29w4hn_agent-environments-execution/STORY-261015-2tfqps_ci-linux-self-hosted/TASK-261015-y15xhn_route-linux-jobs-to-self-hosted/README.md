# TASK-261015-y15xhn: route-linux-jobs-to-self-hosted

## Description
ci.yml: Linux entries of test, go-family-qualification, gate-selftest, lint, interop and naming-gate run on [self-hosted, relux-linux, curator] when vars.LINUX_SELF_HOSTED is true and the event is not a fork pull request; check names unchanged; race stays hosted until the runner has a C toolchain. docs/self-hosted-runner-setup.md gains a Linux section.

## Scope
(define task scope)

## Acceptance Criteria
Fork pull requests never reach the self-hosted label; check names unchanged; the first same-repository PR run executes the Linux jobs on relux-linux runners and is green or its failures are host prerequisites documented in the runner doc.
