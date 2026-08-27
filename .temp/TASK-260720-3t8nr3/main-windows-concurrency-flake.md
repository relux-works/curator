# Main Windows concurrency failure after planner landing

- Exact main SHA: `b3a5031ed551b27a298eef486a068b5175beaacc`
- Main CI run: https://github.com/ivanopcode/cocoaskills/actions/runs/30556125542
- Failed job: Windows Python 3.14, job `90916913692`
- The identical SHA passed all Windows versions in PR run `30554363746`.
- Windows 3.11, 3.12, and 3.13 passed in the main push run.

The only failure was:

`tests/test_transactions.py::test_concurrent_project_transactions_preserve_both_consumers`

Both worker threads had an empty `errors` list, but one remained alive after the
test's fixed `thread.join(timeout=5)` budget. The assertion at
`tests/test_transactions.py:2170` failed. This is outside the planner rework and
inside the current transactional project/hybrid scope.

Treat a successful one-time failed-job rerun as flaky timing evidence, not proof
that the concurrency test is robust. While implementing the project transaction
integration, preserve the vector and assess whether the test needs deterministic
coordination or a platform-appropriate bounded wait. Do not merely skip/xfail it
or remove its liveness assertion.
