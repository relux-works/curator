# rc.8 publication blocker: unsigned GitHub rebase result

## Constraint

PR #20 was independently accepted at signed candidate commit `4d8fda59edfc89367ea050732a414c265be288ed`. The repository rejects merge commits and permits only squash or rebase merges. The first attempted `gh pr merge --merge` was rejected with exit 1. The subsequent policy-allowed `gh pr merge --rebase` succeeded without a protection bypass flag, but GitHub rewrote the commit as `792c53c1887ce02b4b9c1d3954312c919ffb62ef` and removed its signature.

GitHub reports the merged commit verification as `verified=false`, `reason=unsigned`. The repository's own `tools/verify_release_commit.py --commit HEAD` exits 1 with `release target is not a GitHub-created merge commit`. GitHub documents that server-side rebase-and-merge recreates commits without commit signature verification: https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/about-merge-methods-on-github

Tagging this commit would violate the rc.8 acceptance criteria and the release workflow's provenance gate. No rc.8 tag or GitHub release was created. The signed rc.7 tag object `de704f2951e683d52ae8e475cb690b918a94d4c5` remains unchanged.

## Evidence

- PR #20: https://github.com/relux-works/curator-spec/pull/20, state MERGED at `2026-08-19T00:13:40Z`.
- Merged main commit: `792c53c1887ce02b4b9c1d3954312c919ffb62ef`.
- Post-merge Specification CI: https://github.com/relux-works/curator-spec/actions/runs/32200415955 — success.
- Post-merge Implementation conformance: https://github.com/relux-works/curator-spec/actions/runs/32200415997 — success.
- Clean-clone validation: exit 0, 49 schemas and 471 vector files.
- Python test suite: exit 0, 86 tests.
- `go test ./...`: exit 0.
- deterministic vector regeneration: exit 0; authoritative diff check: exit 0.
- clean-tree check: exit 0.
- rc.8 release gate at `792c53c...`: exit 0.
- release-commit verification at `792c53c...`: exit 1, expected blocker because the merged commit is unsigned and not a GitHub-signed squash/merge commit.
- Remote `v1.0.0-rc.8` tag and GitHub release remain absent.

## Failed assumptions and attempts

1. The requested merge-commit method was attempted, but repository settings reject merge commits.
2. Because the signed candidate was exactly one commit ahead of main, rebase merge was expected to preserve acceptable provenance. GitHub instead always rewrites the SHA and does not sign server-side rebase results.
3. The configured required-signature branch protection did not stop this operation because the authenticated admin account is in the protection bypass allowance and administrator enforcement is disabled. No explicit bypass option was used.

## Viable recovery options

1. **Recommended:** authorize a minimal signed no-content release-anchor commit directly on `main`, using the maintainer key. This preserves all history and release bytes, makes the release target pass the repository's trusted-maintainer signature path, and avoids inventing a normative change solely to obtain a PR diff. Tradeoff: the anchor is not itself a PR merge commit, although the release content was independently reviewed and merged through PR #20.
2. Create a meaningful corrective change (for example, release-process enforcement that prevents server-side rebase for release PRs), obtain a new independent review, and squash-merge it so GitHub creates a signed release target. Tradeoff: more scope and another review cycle; a cosmetic-only change would be a forced fit and is not acceptable.
3. Rewrite or force-update `main` to replace `792c53c...` with a signed equivalent. This is not recommended and conflicts with the no-history-rewrite constraint.

## Required human input

Choose and authorize option 1, or authorize the scope and independent review cycle for option 2. Until that decision is supplied, the signed tag, prerelease, and artifacts must remain unpublished.

