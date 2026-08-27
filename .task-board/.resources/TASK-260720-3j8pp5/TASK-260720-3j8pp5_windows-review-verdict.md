# TASK-260720-3j8pp5 independent Windows re-review verdict

Date: 2026-07-30
Reviewer run: `RUN-260730-df2b40`

## Verdict

**Accepted locally → `done`.**

No implementation rework is requested. The narrow fresh-`os.lstat` change
addresses the known Windows false-mutation failure without removing the
file/directory identity comparisons or the full-tree close-time verification.

This is a local acceptance of the current uncommitted bytes. A commit-owning
mover must still commit and publish these exact candidate hashes and obtain a
green GitHub Actions Windows matrix for that exact commit. The reviewer did not
stage, commit, push, bulk-format, or edit product files.

## Provenance and exact candidate

- Product repository: `/Users/iv/Developer/intranet/cocoaskills`
- Review worktree:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3j8pp5/worktree`
- Branch: `task/TASK-260720-3j8pp5-toolchain-identity`
- Clean canonical `main`, worktree `HEAD`, local `main`, and `origin/main`:
  `d5d16bfcaa2fe43dc994b819c2659512c4fd8f0a`
- Canonical origin: `git@github.com:ivanopcode/cocoaskills.git`
- Scope: exactly two modified, unstaged files; 50 insertions and 6 deletions:
  - `src/csk/builds/toolchain.py`: 5 insertions, 5 deletions
  - `tests/test_builds_toolchain.py`: 45 insertions, 1 deletion
- Candidate SHA-256:
  - `c7b5bd70d2784d2c57a8dc336035df010b40befe388dd8ed026b3b1d4d882edd`
    — `src/csk/builds/toolchain.py`
  - `201ba9f2abe42eaa26a49f8d2786d5ce194b79bcf0cf51c7a0a2e877a5224360`
    — `tests/test_builds_toolchain.py`
  - `71faf1fbd73c224f95f8ff26513a4595889f1bee9b9d7cc75924517baeb4e187`
    — `git diff --binary` for the two-file candidate

## Mapping to GitHub Windows failure

[CI run 30503926948](https://github.com/ivanopcode/cocoaskills/actions/runs/30503926948)
tested PR #8 head
`51d8713ad14a26bdc0bafc5216fbed173ba6009b` merged onto this task's base.
All Ubuntu and macOS jobs and strict mypy passed; all four Windows test jobs
failed.

[Windows Python 3.14 job 90749459882](https://github.com/ivanopcode/cocoaskills/actions/runs/30503926948/job/90749459882)
reported `8 failed, 736 passed, 39 skipped`:

- Seven failures were false `toolchain_mutated` results. File cases failed
  when the `VERSION` stat captured from `DirEntry.stat()` disagreed with the
  subsequently opened descriptor's `os.fstat()` identity. The directory case
  failed when cached `DirEntry.stat()` identity disagreed with the later
  path `lstat()`.
- One failure was the focused runner assertion expecting `b"True\n"` while
  native Windows text-mode Python emitted `b"True\r\n"`.

The production fix now uses `os.scandir()` only to collect entry names, closes
the iterator, and obtains each initial physical identity with fresh
`os.lstat(directory / component)`. This removes the Windows `DirEntry` metadata
snapshot from the trust comparison and aligns the initial observation with the
later path-based checks.

The security checks remain in place:

- directory records are rechecked with `Path.lstat()` and `_same_stat`;
- file records are compared with descriptor `os.fstat()` before reading and
  with a final `Path.lstat()` afterward;
- the read byte count must equal the opened size;
- session close re-fingerprints the complete tree and compares both digest and
  frozen tree state.

The new deterministic fake-`DirEntry` regression makes every fake entry's
`.stat()` raise and still fingerprints a nested file successfully, proving the
scanner no longer consumes cached `DirEntry` physical identity. Existing
file-byte, directory-addition, and close-time mutation controls remain green.
The runner assertion now expects the platform-native CRLF only when
`os.name == "nt"`.

## Independent validation

Authoritative interpreter:
`/Users/iv/Developer/intranet/cocoaskills/.venv/bin/python`
(Python 3.14.4, pytest 9.0.3, mypy 2.1.0, build 1.5.0, Twine 6.2.0).
Other tools: git 2.50.1, task-board 0.23.0, gh 2.83.2, uv 0.11.3,
Ruff 0.16.0.

| Gate | Exit | Result |
|---|---:|---|
| Fake-`DirEntry` regression | 0 | `1 passed` |
| File/directory/close mutation controls | 0 | `3 passed` |
| Focused `tests/test_builds_toolchain.py` | 0 | `63 passed in 0.27s` |
| Strict `python -m mypy` | 0 | `Success: no issues found in 58 source files` |
| Full pytest with both fixture roots | 0 | `768 passed, 1 skipped in 84.62s` |
| `python -m compileall -q src/csk` | 0 | Clean |
| `git diff --check` | 0 | Clean |
| Ruff safety lint `E4,E7,E9,F` | 0 | All checks passed |
| Forbidden `go list` / `go build` diff search | 0 | Absent |
| `python -m build` | 0 | Wheel and sdist built |
| `python -m twine check` | 0 | Wheel and sdist passed |
| Packaged source comparison | 0 | Wheel `toolchain.py` SHA-256 equals reviewed source |

Full-suite fixture provenance:

- `CURATOR_CONFORMANCE_ROOT`:
  `/Users/iv/Developer/ReluxWorks/curator-spec/conformance/v1`;
  manifest SHA-256
  `7951cda1711d34d2a9dd9a873cf9d537c41ca4e9527e94f138f38743610a379e`.
  Its bytes match the CI-pinned
  `00b1688a9b2457ca397a0bb550acf47cad8ee967` conformance tree.
- `CURATOR_SCHEMA_V6_ROOT`:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/release-probe/conformance/v1`;
  accepted 48-case manifest SHA-256
  `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`.

Build artifacts were emitted outside the product worktree under
`.temp/TASK-260720-3j8pp5/review-dist-RUN-260730-df2b40`:

- wheel:
  `19cbbbeaef213c3ba8d4e40678859ab32bf9f09c938e5f2d41369428d851a7fb`
- sdist:
  `8d1a0397d3df218520fc7c38dd9c6ed56fda53a8299f781daff03b4ab958bf18`

After all gates, `git status --short` still listed exactly the two reviewed
unstaged product files, and their hashes plus the binary-diff hash were
unchanged.

## Required post-acceptance delivery gate

The current PR #8 checks are evidence for the original failure, not for this
uncommitted fix. The commit-owning mover must:

1. verify the two file hashes and binary-diff hash above;
2. commit and publish that exact scope;
3. run the complete GitHub Actions matrix on that exact commit, with all
   Windows Python 3.11–3.14 jobs green; and
4. request another reviewer cycle if any candidate byte changes or Windows CI
   exposes a new failure.

