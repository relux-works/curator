# TASK-260910-3i6vod — developer results

## Changes

- Added an installed-command security section to README.md stating that installed commands run as the invoking user with that user's operating-system privileges under portable assurance.
- Added SECURITY.md with the same posture and the distinction between declared-only script commands, opt-in `script-worker-v1` enforcement, and explicitly selected provider-backed `verified` mode.
- Both sections link to curator-spec v1.0.0-rc.13 at `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`: Protocol Core §§4.1.1 and 4.2.1, Assurance Protocol §§1, 2, and 5. The exact pinned commit and linked heading anchors were checked from the local spec repository.
- The pinned spec defines the portable script and build enforcement paths and verified-provider contract. It has no clause specifying the installed process OS-user identity, so that S7 disclosure is stated directly in both requested documents. The task remains docs-only.

## Verification evidence

The repository has no documentation build or Markdown link-check target in its Makefile, CI workflows, or `.github/ci` scripts. I attached the task-local checker `TASK-260910-3i6vod_docs_link_check.py`; it checks both required sections, binds each enforcement claim to its wording, and resolves the README-to-SECURITY link plus spec files and anchors against the pinned commit.

- Task-local documentation/spec checker against the changed README.md and SECURITY.md: exit 0 (`PASS: installed-command security wording and pinned links resolve`).
- The same checker against the pre-change `HEAD` README and an empty SECURITY.md fixture: exit 1 as expected because both requested sections were absent.
- Narrow user-privilege mutant (removed `operating-system privileges` from SECURITY.md): exit 1; the checker reported the missing privilege claim and its portable-assurance relationship.
- Narrow script-worker mutant (changed the enforcement claim to declared-only in both docs): exit 1; the checker reported the missing `script-worker-v1` enforcement path in both docs.
- Narrow verified-mode mutant (changed `provider-backed enforcement path` to `provider-backed evidence path` in both docs): exit 1; the checker reported the missing verified enforcement path in both docs.
- Spec-anchor mutant (changed the Core §4.1.1 link anchor in both docs): exit 1; both stale anchors were rejected against the pinned heading.
- `git diff --check`: exit 0.
- Markdown trailing-whitespace/final-newline check for README.md and SECURITY.md: exit 0.
- No Go tests or build were run: the change is documentation-only and the repository has no docs build target.
- Worktree changes are limited to README.md and SECURITY.md. No CHANGELOG.md or LOGBOOK.md edits were made; findings and release-prep wording are recorded here.

## CHANGELOG entry (for release prep)

Clarify in the README and SECURITY policy that installed commands run with the invoking user's privileges under portable assurance, and identify script-worker-v1 and explicitly selected verified mode as the enforcement paths.
