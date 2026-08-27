# TASK-260730-1fsbqd developer commit evidence

## Handoff result

The rejected auto-signed commit was replaced by one unsigned reviewable commit without changing the accepted Git tree or any candidate bytes. This evidence supersedes the prior commit identity. No push, tag, GitHub Release, signing, downstream-pin advancement, or accepted-byte edit occurred.

## Exact Git identity

- Worktree: /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree
- Branch: release/curator-spec-v1.0.0-rc.5-candidate
- Commit: 0aae5dff11ab90400fc6a0b003a4492767b38043
- Parent: 57c1f56846d221ecc55786bd3c2467ec32f11730
- Git tree: 78210085727ec33b79a050a807f51da253ffb0c8
- Subject: Release protocol suite v1.0.0-rc.5
- Signature status: N; raw commit object contains no gpgsig header
- Commit count after base: 1

The prior commit 5c29c1a65bcf084c8ad27d91bcaf9d319f6146f3 is superseded. The old-tree versus new-tree diff exited 0, proving the repair changed only commit-object metadata. git fetch origin main exited 0 and origin/main remains exactly the parent base.

## Accepted committed-suite identity

A fresh git archive of HEAD was extracted and verified independently of working-copy bytes. The archive contains protocol 1.0.0-rc.5, 447 unique manifest entries, and 448 files under conformance/v1 including manifest.json. Every manifest member exists and matches its recorded SHA-256.

- Manifest SHA-256: b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c
- Sorted whole-suite tree SHA-256: e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae

## Validation exits

Every gate ran as a standalone process; the digest pipeline explicitly enabled pipefail.

- git fetch origin main: exit 0
- git commit amend with commit.gpgsign=false and --no-gpg-sign: exit 0
- exact parent, tree, subject, one-commit, origin-base, fast-forward, clean-status, and diff-check gates: exit 0 each
- raw no-gpgsig-header gate: exit 0
- corrected Git signature-status N gate: exit 0
- old accepted tree versus new unsigned tree diff: exit 0
- ../venv/bin/python tools/validate.py: exit 0; 42 schemas and 447 vectors
- Python unittest discovery: exit 0; 41 tests passed
- go test ./tools/...: exit 0
- go vet ./tools/...: exit 0
- gofmt cleanliness assertion: exit 0
- go run ./tools/generate-vectors -root .: exit 0
- regeneration diff assertion: exit 0
- rc.5 release gate at exact HEAD: exit 0
- git archive creation and extraction: exit 0 each
- committed manifest membership, count, version, and SHA assertion: exit 0
- committed sorted suite-tree digest command and exact digest assertion: exit 0 each
- explicit rc.5 metadata assertions: exit 0; exact manifest/base pins, downstream advancement false, zero claims, hardened profile false
- remote candidate-branch absence assertion: exit 0
- remote rc.5-tag absence assertion: exit 0
- final git status porcelain: exit 0 with zero output

Non-green diagnostics are reported truthfully:

- The first signature-status assertion left %G? unquoted; zsh rejected the glob before Git ran. It exited 1. The corrected quoted assertion then exited 0 and returned N.
- GitHub Release lookup exited 1 with the expected HTTP 404, proving no v1.0.0-rc.5 release currently exists. This is not reported as a passing command.

## Isolation and scope

All Git mutation and verification ran in the assigned accepted worktree. The unsigned commit has the exact same tree as the independently accepted candidate and reproduces both accepted content digests, so no bytes were sourced from the dirty primary worktree. The worktree is clean. No local tag points at HEAD, and no remote candidate branch or rc.5 tag exists. Tag and GitHub Release publication remain deferred until a new human command. The standalone logbook facility remains unavailable; the signing regression and repair are persisted in task notes and outcome resources instead.