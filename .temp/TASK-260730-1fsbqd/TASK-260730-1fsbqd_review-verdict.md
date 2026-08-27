# TASK-260730-1fsbqd independent reviewer verdict

Verdict: **accepted**.

The exact unsigned commit `0aae5dff11ab90400fc6a0b003a4492767b38043`
is accepted for the subsequent fast-forward landing to
`relux-works/curator-spec` `main`. The superseded signed commit
`5c29c1a65bcf084c8ad27d91bcaf9d319f6146f3` was not reviewed as the landing
candidate and must not be landed.

## Exact commit and remote base

- The assigned branch and `HEAD` both resolve to
  `0aae5dff11ab90400fc6a0b003a4492767b38043`.
- The commit has exactly one parent,
  `57c1f56846d221ecc55786bd3c2467ec32f11730`, exactly one commit exists after
  that base, and the base is an ancestor (all gates exit 0).
- A final live `git ls-remote origin refs/heads/main` returned
  `57c1f56846d221ecc55786bd3c2467ec32f11730`; local `origin/main` matches it.
- The Git tree is exactly
  `78210085727ec33b79a050a807f51da253ffb0c8`.
- The subject is exactly `Release protocol suite v1.0.0-rc.5`.
- The raw commit object contains no `gpgsig` header; the explicit inverted
  header gate exits 0 and Git signature status is `N`.
- The assigned worktree is clean with zero porcelain output. No local tag
  points at the reviewed commit.

## Accepted committed bytes

A fresh `git archive` of the reviewed commit was extracted without `.git`
metadata and checked independently of working-copy bytes:

- `conformance/v1/manifest.json` contains 447 unique entries and declares
  protocol `1.0.0-rc.5`.
- The committed suite contains exactly those 447 members plus `manifest.json`,
  for 448 files total. Every listed path exists and all 447 recorded SHA-256
  values match its committed bytes.
- Manifest SHA-256 is exactly
  `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`.
- The accepted sorted whole-suite tree recipe over all 448 files returns
  exactly
  `e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae`.
- `release/1.0.0-rc.5.json` pins that exact manifest and the exact source base,
  records `committed_release_pin_advanced=false`, emits zero claim-v3 claims,
  and records `hardened_profile_claimed=false`.

The primary curator-spec worktree remains at the base with 20 dirty status
entries, while the assigned landing worktree is clean. The reviewed commit
reproduces both the exact whole-repository Git tree required by the review
brief and the independently accepted suite digests, excluding dirty-primary
contamination.

## Independent validation

In a separate `--no-local` clone detached at the exact reviewed commit:

- `make release-check VERSION=1.0.0-rc.5` exited 0.
- Validation reported 42 schemas and 447 vector files.
- Python unittest discovery passed all 41 tests.
- `go test ./tools/...` passed.
- Generation completed and the committed conformance/release diff gate exited
  0.
- The rc.5 release gate passed at the exact commit SHA.
- A second consecutive generation also left zero diff.
- `go vet ./tools/...`, gofmt cleanliness, `git diff --check`, and final clean
  status all exited 0.

## Publication boundary

A final remote check found no candidate branch and no `v1.0.0-rc.5` tag.
`gh release view` exited 1 with `release not found`, the expected absence
result. This reviewer did not modify code, push, tag, publish a release, sign a
commit, or advance downstream pins.

The next mover may fast-forward `main` only to the exact accepted commit above.
Any later tag or GitHub Release action is outside this reviewer verdict and
must follow the latest human authority recorded by the orchestrator; this
reviewer neither performs nor independently expands that publication authority.
