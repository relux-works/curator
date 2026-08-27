# TASK-260822-f4qv7w reviewer verdict

Verdict: **changes requested -> `to-dev`**.

## Accepted implementation evidence

- Reviewed the unstaged delta in `/Users/iv/Developer/ReluxWorks/curator-spec/.temp/STORY-260822-3k3hbs/normative-worktree` on `spec/script-worker-v1-normative` at `a690d63`.
- `origin/main` at `be7861c` and all three prerequisite heads are ancestors of `HEAD`: `spec/sw-core-prose` `78d544d`, `spec/sw-manager-security` `e5df43d`, and `spec/sw-schema` `ebfed81`.
- The generated vector pins Linux `active-process-count-limit` to `host-conditional` / delegated cgroup v2 `pids.max`, contains no `RLIMIT_NPROC` substitution, and includes probe unavailable -> evidence unavailable -> invocation succeeds.
- Positive and negative cases cover schema-8 opt-in, legacy declared-only behavior, deny-by-default derivation, mandatory-control preflight rejection, evidence-record cardinality/field/version/foreign-policy closure, and audit labels.
- Independent isolated-copy validation passed: `make validate` validated 52 schemas and 658 vector files, ran 95 Python tests successfully, and passed `go test ./tools/...`.
- Two independent regeneration passes both exited 0; SHA-256 inventories of all files below `conformance/v1` and `release` were byte-identical. `gofmt -d tools/generate-vectors` was empty.
- The implementation matches the reviewed prose and project generator/validator architecture. No content defect was found in the task-owned vector and validation surfaces.

## Required delivery rework

The task acceptance criteria are not yet satisfied on the delivered bytes:

1. The vector delta is unstaged and uncommitted; `spec/script-worker-v1-normative` is only at merge commit `a690d63`, while all task changes remain in the worktree.
2. `git ls-remote --heads origin spec/script-worker-v1-normative` returns no branch. Therefore no remote branch CI has run on the reviewed bytes, and the checked board item claiming green spec CI is unsupported.
3. A commit-owning mover must commit only the reviewed scope (excluding `.temp/` and `tools/__pycache__/`), push the story branch, and record the branch CI run and green required gates. Then return through another reviewer cycle.
4. Do not land the generated `release/1.0.0-rc.8.json` rewrite as historical rc.8 evidence. `TASK-260822-c0rxj7` already owns the re-scoped shared Schema 8 candidate and rc.9 release migration; the committing/landing path must reconcile this generated intermediate delta there.

This is ordinary delivery rework, not a stop-the-line external blocker.
