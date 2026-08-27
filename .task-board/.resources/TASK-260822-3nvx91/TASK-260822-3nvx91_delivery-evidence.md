# TASK-260822-3nvx91 delivery evidence

## Reviewed scope delivery

- Story worktree: `/Users/iv/Developer/ReluxWorks/curator-spec/.temp/STORY-260822-1pm1c9/prose-worktree`
- Branch: `spec/module-roots-prose`
- Base: `ebfed8171cd49eec0c8c010801929a01d2352569` (`spec/sw-schema`)
- Commit: `61ab80154c6aa8a83a33f2f2bbd8ec6e3dc1df50`
- Commit subject: `Specify declared first-party module roots`
- Signature: good (`G`), signer `oparin@me.com`
- Push: `origin/spec/module-roots-prose`, exit 0, new upstream configured
- Remote verification: local HEAD, tracking ref, and `git ls-remote` all equal
  `61ab80154c6aa8a83a33f2f2bbd8ec6e3dc1df50`
- Worktree after regeneration and push: clean and up to date

No pull request was created, matching the sequencing instruction.

## Post-commit validation

Each gate was run directly as a standalone process.

| Gate | Exit | Evidence |
| --- | ---: | --- |
| `make validate` with task venv | 0 | 52 schemas, 686 vectors, 95 Python tests, Go tools pass |
| `test -z "$(gofmt -l tools)"` | 0 | no formatting drift |
| `git diff --check` | 0 | no whitespace errors |
| `lychee ... '**/*.md'` | 0 | 40 OK, 0 errors, 1 excluded |
| `make regenerate-check` with task venv | 0 | generated conformance/release bytes match HEAD |
| `git show --check --oneline HEAD` | 0 | committed patch clean |
| clean worktree/index check | 0 | no uncommitted or staged delta |

Logs are stored under `.temp/TASK-260822-3nvx91/` in the Curator checkout.

## Coordination

The branch extends shared manifest Schema 8 in place with `$defs.buildCommandV8`.
No Schema 9 was introduced, matching TASK-260822-1mwy10.
