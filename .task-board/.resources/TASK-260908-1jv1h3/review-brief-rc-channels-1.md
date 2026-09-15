# Review brief: keep a release candidate out of the install channels (cycle 1)

## Subject

- Branch `chore/rc-stays-out-of-install-channels` at `d1bb0a4e` in
  `/Users/iv/Developer/ReluxWorks/.worktrees/curator-release-channels`, one signed
  commit past `origin/main` (`04550e28`). PR https://github.com/relux-works/curator/pull/66
  is already landed; this branch was rebased onto it with `-S` and range-diff
  proves the commit identical (`6a92b0f6 = d1bb0a4e`). Diff:
  `git diff 04550e28..HEAD` — one file, `.goreleaser.yml`, +8/−0.
- Authority (measured, not assumed): `v0.14.0-rc.1` reached
  `relux-works/homebrew-tap` as a commit on 2026-08-06, so brew served a
  candidate as the current release. That is the incident this leaf closes.
- Change Request: the Story workspace is a curator-spec checkout and this leaf
  delivers into curator, so the CR delta reads empty. That is structural in
  this epic and is not the finding (see TASK-260906-284db9 cycle-1 verdict §0).

## What this changes and why

Three fields move to `"auto"`:

1. `homebrew_casks[0].skip_upload: "auto"` — skip the cask upload on prerelease.
2. `scoops[0].skip_upload: "auto"` — same for the scoop manifest.
3. `release.prerelease: "auto"` — mark the GitHub release prerelease when the
   tag carries a SemVer prerelease suffix.

Each carries a comment stating the reason. Target tag: `v0.15.0-rc.1`.

## Review dimensions

1. **Semantics from the pinned toolchain, not from the comments.** Resolve which
   GoReleaser version the release workflow pins and verify from that version's
   docs/source that `skip_upload: "auto"` means skip-on-prerelease (and what it
   does on a stable tag), and that `release.prerelease: "auto"` keys off the
   SemVer prerelease suffix. If `auto` means anything else on the pinned
   version, the leaf is wrong no matter how honest the comments are.
2. **The comments must be true in every clause.** "v0.14.0-rc.1 reached the tap
   because this was unset" — confirm the tap commit exists and the unset field
   is the mechanism that allowed it, not a neighbouring default.
3. **Nothing else changes.** One file, three fields, three comments. Confirm no
   workflow, tap/bucket repo, or footer/template behaviour shifts.
4. **The stated boundary.** The publish phase cannot be reproduced locally
   without a real write, so this review is docs-confidence by construction.
   Say so explicitly in the verdict and name the first real proof: the rc1 tag
   itself, checked for prerelease marking and absence of a tap commit.
5. **Hosted lanes.** Read `gh pr checks 65` on head `d1bb0a4e`. All required
   lanes must be green on the reviewed head. The change is config-only; a red
   lane is either unrelated (prove it) or a rebase casualty (say so).

## Method

No gates to run locally beyond `goreleaser check`/`lint` if the toolchain
offers an offline validation for the touched stanzas. Do not attempt a publish
dry-run that writes anywhere.

## Constraints

Scratch under a throwaway copy or the worktree's `.temp/`; never write into
the producer's worktree or the control root. No LOGBOOK.md entry.

## Verdict contract

Attach `TASK-260908-1jv1h3_review-findings-1.md`. Blocking or major → set the
task to `development`. Otherwise an explicit ACCEPT at `to-review` with
`accept_cr` on the recorded revision, stating the pinned GoReleaser version,
what `auto` means on it per dimension 1, the hosted-lane state on `d1bb0a4e`,
and the docs-confidence boundary. Do not mark the task done. Then
`task-board handoff TASK-260908-1jv1h3 --role reviewer`.
