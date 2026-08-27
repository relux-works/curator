# TASK-260822-f4qv7w developer handoff

## Scope

- Worktree: `/Users/iv/Developer/ReluxWorks/curator-spec/.temp/STORY-260822-3k3hbs/normative-worktree`
- Branch: `spec/script-worker-v1-normative`
- Prerequisite ancestry: `origin/main`, `spec/sw-core-prose`, `spec/sw-manager-security`, and `spec/sw-schema` are all ancestors of `HEAD` (four `git merge-base --is-ancestor` checks, exit 0).

## Delivered changes

- Generated `conformance/v1/vectors/script-host-execution-policy.json` with positive and negative cases for schema-8 opt-in parsing, legacy declared-only behavior, deny-by-default capability derivation, mandatory-control preflight, closed script capability evidence, and audit labels.
- Linux `active-process-count-limit` is `host-conditional` and backed only by delegated cgroup v2 `pids.max`; the vector explicitly proves probe unavailable -> evidence unavailable -> invocation succeeds. `RLIMIT_NPROC` is rejected by mutation tests.
- Generated `conformance/v1/expected/install-marker-v4.json` and added the schema-8 -> marker-v4 lifecycle case.
- Added Go contract tests and Python validator mutation tests; regenerated the conformance manifest and rc.8 suite digest.
- Updated `conformance/README.md` with the new vector family.

## Validation evidence

| Command | Exit | Result |
| --- | ---: | --- |
| `go run ./tools/generate-vectors -root .` (regeneration 3) | 0 | Generated authoritative vectors |
| `go run ./tools/generate-vectors -root .` (regeneration 4) | 0 | Second generation |
| `cmp -s generated-after-regenerate-03.patch generated-after-regenerate-04.patch` | 0 | Byte-identical second run |
| `PATH=<task-venv>/bin:$PATH make validate` | 0 | 52 schemas, 658 vector files, 95 Python tests, all Go tool tests |
| `gofmt -l tools` empty check | 0 | Formatting clean |
| `git diff --check` | 0 | Whitespace clean |
| CI-equivalent `lychee ... '**/*.md'` | 0 | 42 links checked; 0 errors on final retry |

The task-local venv installed exactly `requirements-dev.txt` because the system Python lacked `jsonschema==4.25.1`. Two earlier Python attempts exited 1 for invocation/import-environment reasons and are retained in task-local logs; they are not reported as passing. One intermediate link run exited 2 on a transient connection failure to the existing Curator GitHub URL; the final retry exited 0.

## Version-control boundary

Changes remain unstaged and uncommitted because the active repository instructions prohibit automatic staging or commits. Remote multi-OS CI therefore has not run on these bytes; the exact local CI commands are green and the branch is ready for reviewer inspection and the authorized commit/push step.
