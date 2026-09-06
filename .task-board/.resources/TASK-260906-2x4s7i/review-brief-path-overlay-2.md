# Review brief: path overlay declarability, cycle 2 (rework 1)

## Subject

- **Not the managed story workspace this time.** The story branch had drifted — it carried a commit the
  orchestrator had already landed under a different hash, so the CR machinery refused to capture a
  candidate. The producer's work was moved verbatim onto a branch cut from current main and committed
  by the orchestrator. Review branch `feat/path-overlay-declarable` at `bd39adb` in
  `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-path-overlay`, one signed commit past
  curator-spec main. PR https://github.com/relux-works/curator-spec/pull/47. Diff:
  `git diff origin/main..bd39adb`.
- Read `TASK-260906-3x0w4y_review-findings-1.md` (your predecessor's cycle-1 findings),
  `producer-brief-path-overlay-rework-1.md`, and the producer's rework report if one is attached.
- Authority: `protocol/environments.md` §1, §6, §12.1; `protocol/core.md` §6.1;
  `profiles/manager.md`; `schemas/v1/manager-config-v2.schema.json`.

## Standing

Cycle 1 found F1 (the discriminator classified a Windows absolute path as `git`, failing in both
directions), F2 (the discriminator had no killing evidence — the corpus published four `source`
spellings), F3 (`profiles/manager.md` still stated the universal form; the orchestrator ruled it in
scope), F4 (`svn://host/x` classified as `path` and was admitted form-free). All four were addressed.
Verify each, and treat the fixes as claims, not as facts.

## Review dimensions

1. **F1, driven.** Build the full matrix against the *committed* schema with a JSON Schema validator,
   both directions: `C:\…` and `C:/…` bare (must be valid) and each carrying `range`, `tag`,
   `revision`, `directory`, `branch` (must be invalid); the same for a POSIX absolute path and a
   project-relative spelling; a `git` source bare (must be invalid) and with each form (valid). The
   orchestrator ran the pattern alone over 16 spellings with zero mismatches, but a pattern test is
   not a schema test — drive the schema.

2. **The residual edges the orchestrator did not resolve.** Two are known and neither is settled:
   - `C:foo` — a drive letter with no separator — still classifies as `git`. Is that a real spelling
     of anything? Decide whether it matters and say why.
     - `file:///x` is classified as a legal **path** source. §1 says a `path` source is "an
     operator-local package directory named by an absolute path, or by a project-relative path". A
     `file:` URL is a URL, not a path spelling. Decide from §1 whether admitting it is right; if it is
     not, the second `allOf` arm's scheme allowlist is where it enters.
   Look for further edges of your own: an empty-ish source, a source with a colon inside a relative
   first segment, a UNC path (`\\server\share\…`), a scheme-relative `//host/x`. For each, say which
   direction it fails in — misclassified toward `git` makes a real path overlay undeclarable;
   misclassified toward `path` lets a git overlay skip its form requirement.

3. **F2 — is the evidence actually killing now?** Eighteen cases were added. Test them as a gate:
   mutate the discriminator (swap an arm, drop the backslash exclusion, drop the host grammar, widen
   the scheme allowlist) and confirm a *named case* flips. Report any mutant that survives the new
   corpus. Your predecessor's own repair pattern was green against the old corpus; check whether it
   would now be distinguishable from the committed one.

4. **F4 and the new second `allOf` arm.** A URI scheme outside `{ssh, git, http, https, file}` is now
   refused outright with `then: false`. Verify that is faithful to core §6.1 and §1 rather than a
   convenient catch-all, that it cannot refuse something §1 permits, and that its interaction with the
   first arm is what the cases claim.

5. **F3.** The manager sentence must agree with the amended §12.1 row and stay in the manager's voice —
   manager-side obligations, citing the environments section, not restating the rule normatively.

6. **Mechanics and gates.** One signed commit past `origin/main`, human identity, no stray file —
   including no `tools/__pycache__`, which is ignored on this base but was not on the previous one.
   Re-run `make validate` and `make regenerate-check` yourself; do not trust the PR body. Confirm from
   the pinned manager's source, not from a claim, that no file it consumes gained a case it cannot
   pass. Read `gh pr checks 47` and treat any red lane as blocking.

## Constraints

Read-only outside your scratch. Never write into the control root, and do not push or modify the branch.

## Verdict contract

Attach `TASK-260906-3x0w4y_review-findings-2.md` with a `repeat-of:` line naming any cycle-1 class that
recurs. Blocking or major → set the task to `development`. Otherwise an explicit ACCEPT at `to-review`
with `accept_cr` on the recorded revision, noting that the reviewed artefact is PR #47 rather than a
story-branch candidate. Do not mark the task done. Then
`task-board handoff TASK-260906-3x0w4y --role reviewer`.
