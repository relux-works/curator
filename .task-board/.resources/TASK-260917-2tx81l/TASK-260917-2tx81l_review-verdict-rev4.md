# TASK-260917-2tx81l — rev4 refresh identity review: ACCEPTED

Rev4 (base bd126a9a, tree 4001bbe4) is a base refresh of rev3 (base 30b3d678, tree 92a072c4, ACCEPTED by the prior reviewer; substance is cited from `TASK-260917-2tx81l_review-verdict-rev3.md`, findings: none). This review proves identity only. `findings: []`.

## Identity proof
- Path sets: `git diff --name-only` rev3 vs rev4 — both 45 paths, identical lists.
- Per-path blobs: 41/45 blobs byte-identical rev3→rev4 (all tests, all five added files, conformance-gaps.tsv). The 4 differing paths: `.github/ci/conformance-case-counts.tsv`, `internal/config/config.go`, `internal/envprofile/managed.go`, `internal/install/targets.go`.
- Patch +/- lines, two-way: `git diff 30b3d678 <rev3>` vs `git diff bd126a9a <rev4>` — 2793 vs 2793 +/- lines, sorted `diff` empty. The candidate delta is the same change.
- The 4 differing blobs: each rev4 blob equals `git merge-file(ours=rev3 blob, base=30b3d678 blob, theirs=bd126a9a blob)` byte-for-byte (cmp exit 0 on all four); the trunk-only change in each is the incoming delta (counts +4, config.go +5/-1, managed.go +4/-1, targets.go +1/-1). So any difference is solely trunk context.
- Worktree files for all 45 paths hash-equal the rev4 tree blobs (5 untracked additions included); no extra tracked non-board changes.

## Counts file
- Both sides present: rev3's `skillfile-sources-v1/schema-cases 131→121` under candidate digest 950ee74a… and trunk's four `external-repository/acquisition/*` rows under rc.13 digest be11bb1e….
- Independent re-derivation: rc.13 vectors/external-repository-acquisition.json → cases 12, clean_environment 17, common_fetch_argv 55, forbidden_fetch_features 11 (matches). Candidate b1a2efb conformance/skillfile-sources-v1/index.json has 121 entries and 121 schema-case files (matches). Disposable clones verified: manifest sha256 of rc.13 clone = be11bb1e…, b1a2efb clone = 950ee74a…. Totals by digest in the file: rc.13 92 pins / 1731 cases; candidate 97 pins / 1722 cases (match producer's recomputation).
- `go test ./internal/conformancecoverage -count=1`: exit 0 under rc.13 root; exit 0 under b1a2efb root.

## Other reruns (real exits)
- `go test ./internal/hashing ./internal/registry ./internal/marker ./internal/contextlock -count=1`: exit 0 under b1a2efb root; same plus config and envmarker: exit 0 under rc.13 root.
- Mutant sites (hashing/registry/marker/contextlock sources and all tests) are byte-identical to rev3, so rev3's three-mutant kill evidence (exit 1 each) carries; not re-mutated here. Producer's rev4 mutant evidence is attached in the refresh4 artifacts.

## Hygiene
No CHANGELOG/LOGBOOK path, no Windows-reserved filenames, 45 non-board worktree entries = the 45 owned paths (5 untracked are the declared additions), no stray files. No added line names an employer.

## Bounds
Not re-run here: full hosted gate (runner's for rev4), the rev3 base-test overlay replay (carried; test files byte-identical), the pre-existing N1 snapshot-acquisition driver mismatch under b1a2efb noted in the rev3 verdict (not part of this task, reproduces on base).
