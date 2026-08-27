# TASK-260720-3c0ss2 Windows CI rework

## Provenance and boundary

The candidate is in /Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-3c0ss2/worktree on branch task/TASK-260720-3c0ss2-build-source-identity at signed published base 2734beff1a0c93d725c00b1c66ef6ad22c3a780a. The published commit was not amended. The rework is unstaged and uncommitted for review. No SSH, Go command, branch push, PR, tag, release, or unrelated repository mutation was performed.

## Root cause and implementation

Windows os.DirEntry.stat reports zero st_dev and st_ino by documented design, while os.lstat and os.fstat return physical file identity. The fallback snapshot walk now retains only entry names from scandir and records every descendant through no-follow os.lstat, preserving the existing before-open, opened-handle, and after-read identity and size checks. NTFS interprets bad:name as a named data stream whose directory entry is only bad. New internal src/csk/builds/_windows.py enumerates DATA streams with FindFirstStreamW and FindNextStreamW using last-error preservation and extended paths. The snapshot rejects every named stream and unexpected enumeration failure for the root and each descendant while accepting only the unnamed default stream. POSIX descriptor traversal is unchanged.

## Tests

Added a platform-independent regression that makes cached DirEntry stat unusable, a Win32 API seam test for default-stream filtering and handle closure, and a native Windows named-stream rejection test. The pre-fix physical-stat regression ran directly and failed as expected with exit 1. After the fix, tests/test_build_source.py ran directly with exit 0: 19 passed and 1 Windows-only skip.

## Exact local gates

- Accepted conformance plus task-focused pytest: exit 0, 198 passed and 1 Windows-only skip.
- Full pytest with released conformance and accepted schema-v6 roots: exit 0, 707 passed and 2 platform skips.
- python -m mypy: exit 0, 58 source files clean.
- python -m build: exit 0, sdist and wheel built.
- python -m twine check dist/*: exit 0, all present distributions passed.
- Wheel and sdist inventory assertion for csk/builds/source.py and csk/builds/_windows.py: exit 0.
- python -m compileall -q src/csk tests: exit 0.
- git diff --check: exit 0.

The ambient python command remains unavailable and the first accidental invocation exited 127; all repository gates used PATH=/Users/iv/Developer/intranet/cocoaskills/.venv/bin:$PATH.

## Native Windows gate sequencing

Origin CI run 30500306639 is the red baseline at published commit 2734beff: all four windows-latest Python jobs fail the same eight source-boundary tests. The repository CI has only push to main and pull request to main triggers, so an unstaged candidate cannot execute on windows-latest. The operator directive requires the new fix commit only after re-review. Native origin Windows validation must therefore follow review-authorized commit publication; no temporary commit or PR workaround was introduced.