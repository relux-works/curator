# TASK-261015-1yj9x7: buildrepo-tests-portable-false-stub

## Description
buildrepo tests hard-code /usr/bin/false as a regular credential broker and SSH wrapper; on uutils-coreutils hosts it is a symlink and 28 tests fail. Use a temp regular stub; keep the symlink refusal proven.

## Scope
(define task scope)

## Acceptance Criteria
No test relies on /usr/bin/false or /usr/bin/true being a regular file; buildrepo tests pass where they are symlinks and on macOS; a test proves a symlinked broker is still refused; production code unchanged
