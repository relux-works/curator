## Status
backlog

## Assigned To
(none)

## Created
2026-07-20T02:09:20Z

## Last Update
2026-08-01T07:19:57Z

## Blocked By
- TASK-260720-3pemm6
- TASK-260720-akf5kh

## Blocks
- TASK-260720-1utsx8
- TASK-260720-3pvihp
- TASK-260720-2g7avf
- TASK-260720-p7sdhg
- TASK-260728-1ph8rs
- TASK-260728-1j72zq
- TASK-260728-2yxdo7
- TASK-260728-3j60e3
- TASK-260728-3ar1qp
- TASK-260730-2gtlzn

## Checklist
- [ ] Record the exact integrated base and conformance SHAs in a clean task-scoped worktree.
- [ ] Run and attach full pytest, rc.4 vector, strict mypy, build, twine, diff-check, and cross-platform CI evidence.
- [ ] Attach an acceptance-criterion and negative-vector coverage matrix and route any substantive defect to its owning task.

## Notes
Cross-story verification boundary from STORY-260720-21bsr2: candidate-suite evidence is not release evidence and must record the exact supplied suite digest without advancing a committed pin. The released-suite pin and no-skip CI audit is TASK-260720-1utsx8 after protocol release qualification.
The updated rc.6 scope supersedes the original checklist reference to rc.4/rc.5. Verification must use exact curator-spec 432eb2ee and manifest sha256:12e58b82579645ba1ccafba49d3e2dd3216005ddf37ae63c68a9fafd46773071.

## Precondition Resources
(none)

## Outcome Resources
(none)

## Estimate
estimated(fibonacci(8))
