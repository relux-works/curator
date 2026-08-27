# TASK-260720-2g21eg review rework, cycle 2

## Reviewed source state

- CocoaSkills worktree:
  `/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-2g21eg/worktree`
- Branch: `task/TASK-260720-2g21eg-go-v1-compile-driver`
- Base and current `HEAD`:
  `97a0ed870782b48eebc5a9c25a9cfa8fea5ff245`
- Accepted rc.5 vectors:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1`
- Host used for native validation: Darwin 25.5.0 arm64

## Review findings addressed

1. The installed manager identity now binds the launcher, its interpreter
   invocation/link chain and executable, and the complete installed `csk`
   package tree. The full proof is carried in the worker request, checked
   against the package and interpreter actually loaded by the worker, and
   reverified around each Go process and again before result acceptance. A
   fixed empty private bytecode cache prevents local `__pycache__` state from
   becoming an unproved execution input.
2. The session now carries a frozen identity for both
   `<GOROOT>/bin/go` and the complete executable tree below the native
   `<GOROOT>/pkg/tool/<GOHOSTOS>_<GOHOSTARCH>`. The worker's direct process
   boundary accepts only that Go identity and rechecks the Go/tool set before
   and after actual execution. macOS vnode guards detect replacement/restore;
   Windows retained handles deny replacement while the operation is live.
   Named negatives cover Go and tool replacement/restore and an outside
   executable attempt through the worker and real executor boundary.
3. Worker-domain teardown is fail closed. Windows Job Object termination is
   followed by an active-process-zero query and checked handle close; a failed
   close retains the handle for retry/diagnosis. macOS kills and joins the full
   process group and proves it absent. Teardown errors suppress the result and
   remove staging. Success and failure seams exist for both native paths, and
   the installed-wheel fixture exercises result suppression after a teardown
   failure.

The installed-wheel negatives start a real worker, replace either installed
package code or the installed launcher before the worker identity proof,
return `build_execution_worker_identity_invalid`, reach no compiler state,
leave no staging result, and terminate the worker.

## Gate ledger

Every command below ran directly as a standalone process without a pipe or
`tee`.

| Gate | Result |
| --- | --- |
| Focused accepted-root source/toolchain/driver/fixture pytest | exit 0; 170 passed, 8 skipped |
| Full accepted-root pytest | exit 0; 857 passed, 10 skipped |
| `python -m mypy src/csk` | exit 0; no issues in 60 source files |
| `python -m compileall -q src tests` | exit 0 |
| `python -m tabnanny` on the four task files | exit 0 |
| `git diff --check` | exit 0 |
| Explicit trailing-whitespace check on the four task files | exit 0 |
| `python -m build --wheel` | exit 0 |
| `python -m twine check` on the fresh wheel | exit 0; passed |
| Source vs fresh installed-wheel `go_v1.py` comparison | exit 0; exact |
| Source vs fresh installed-wheel `cli.py` comparison | exit 0; exact |
| Fresh wheel-installed native macOS Go fixture | exit 0; 4 passed in 19.98s |

The native fixture covers one accepted real vendored Go build, installed
package replacement, installed launcher replacement, and teardown-result
suppression. The verified artifact was never executed.

Native Windows execution was not available on this Darwin host, and no Windows
runtime was installed. The Windows Job Object success/failure and handle
retention logic passed the platform-neutral seam tests included in both pytest
gates. A native `windows-latest` run remains an explicit reviewer/CI follow-up;
this evidence makes no native Windows success claim.

## Candidate hashes

- `src/csk/cli.py`:
  `9e0724b53e6fbcd86f967f611c19336130620d9a7fb7500b9dc4d879dc35a92c`
- `src/csk/builds/go_v1.py`:
  `1a395b52b8bc7c0caf35565f06f1e9a7fe71faa97865955309be657ff9f6766f`
- `tests/test_builds_go_v1.py`:
  `5acbcac78d935de7f7a796ae1e553a89ea2c96057fbeaa14ffbf37b9c8a3a015`
- `tests/test_builds_go_v1_fixture.py`:
  `ee47bf895deea53920a99ff8ee4d4d33ad897c71d0fcc2729c4fb1c47780e4d2`
- Fresh wheel:
  `b4d50ba66a7015c92a9797f1004af003673788e40a7d20f5acfb98bf4943cbc7`

No commit or index mutation was made. The task worktree contains only the four
task-owned source/test paths.
