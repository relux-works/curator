# Publication review \u2014 initial candidate

Task: TASK-260910-1xph2y; story: STORY-260910-8fv3s5.
Review scope: publish the accepted specification draft and truthful planning records only. No feature implementation, release qualification, source edits, staging, commits or remote mutations performed by this reviewer.

## Candidate identity

Base HEAD: `d019f0e7179520b5c8dcde321c4fe51e04552f58`.
Accepted CR2 normative tree: `4087f02f5459a96d1a03d78ddb343d82608df0e6`.
All 126 normative paths changed between base and CR2 compare byte-for-byte with the current worktree. Existing frozen schema JSON, conformance/v1, release and tools have zero candidate delta. No material normative findings. Prior accepted review was inspected, and its draft/adversarial checks independently rerun; publication claims remain explicitly unreleased and implementation-unverified.

## Independent validation (2026-09-15)

Platform: platform-neutral specification/tooling, macOS arm64 host; Python 3.14.7, Go 1.25.5, Make 3.81, Git 2.50.1. Tool readiness output: `.temp/publication-review/readiness-01.log`.

| Command | Exit | Evidence |
| --- | ---: | --- |
| `PATH="$PWD/.temp/source-contract/venv/bin:$PATH" make validate` | 0 | 60 schemas, 1047 vectors, 227 Python tests, Go tools tests pass; validate-01.log |
| `.temp/source-contract/venv/bin/python3 .temp/publication-review/draft.py` | 0 | Freshly extracted exact documented command: 102/102 cases, 82/82 negatives, 7 wire schemas, 3 snapshot vectors, 25 v4 fields, 2 build arms, 18 refusal mutations; draft-01.log |
| `.temp/source-contract/venv/bin/python3 .temp/review-rev2/adversarial.py` | 0 | 16 marker checks, endpoint cardinality and narrowed mutant, 11 selector attacks, 10 guide examples, 126 exact CR2 path comparisons; adversarial-01.log |
| `git diff --check d019f0e7179520b5c8dcde321c4fe51e04552f58 4087f02f5459a96d1a03d78ddb343d82608df0e6` | 0 | boundaries-01.log |
| `git diff --exit-code d019f0e7179520b5c8dcde321c4fe51e04552f58 4087f02f5459a96d1a03d78ddb343d82608df0e6 -- 'schemas/v1/*.json' conformance/v1 release tools` | 0 | frozen-01.log |

All named logs are below `.temp/publication-review/`. Manager semantic execution remains **0/73, unverified**; structural schema checks do not establish runtime behavior. No new implementation tests or production changes were introduced.

## CI configuration

`.github/workflows/ci.yml` configures Specification (ubuntu-latest), Specification (macos-latest), Specification (windows-latest), Formatting, and Links for PRs. Specification jobs run validation, Python tests, Go tests and deterministic regeneration, using Python 3.12 and go.mod version. Release target provenance is explicitly skipped on PRs and executes on main pushes/workflow dispatch.

`.github/workflows/implementations.yml` configures Implementations for Ubuntu/macOS/Windows, consuming the frozen conformance/v1 corpus using pinned existing manager/registry checkouts. This does not execute the new draft semantic corpus or claim new implementation support. `.github/workflows/release.yml` is tag-triggered only; publishing this draft branch/PR does not invoke a protocol release. Repository branch-protection required-check settings must be read remotely by the owning agent; workflow presence alone does not establish which checks protection requires. Hosted checks remain pending publication; local macOS evidence does not certify the remote platform matrix.

## Pending exact-head review

Parent owns final `.planning` corrections and commit/publication. Verify final committed head, complete `.planning` delta, normative preservation and native PR review after parent supplies exact head. No pre-approval of a future unknown head.

## Settled planning review

All four `.planning` documents reviewed after parent completion. No material findings remain: historical one-wave schedule and correction brief carry explicit superseded/resolved notices; the current plan states publication authorization and preserves the user pause on implementation. A structural consistency check passes with 7 unique stories, 15 unique tasks, exactly 20 listed dependency edges, and all 73 draft semantic IDs each assigned once to a declared task. Current canonical snapshot contains six phases and the same critical path as the implementation plan. Exact reviewed planning SHA-256 hashes are recorded in `.temp/publication-review/planning-01.json` for comparison with final committed objects. This review checks consistency of recorded recovery evidence; parent independently owns the authoritative board mutation/verification. Final committed-head identity remains pending.

## Final exact-head review and native verdict

Accepted publication head: `3535d63ea80f97bba2fcb6e1f06996cfc25cf7df`. Normative ancestor `a4fcaf024bf25aa94e22367aa111864802f1f8e5` has exactly CR2 tree `4087f02f5459a96d1a03d78ddb343d82608df0e6`. All four final planning files match the previously reviewed SHA-256 hashes. Complete remote PR patch bytes match the local base-to-head diff and all 130 changed paths are accounted for. Initial `gh pr view` exposed only 100 files; the assertion detected pagination truncation, then paginated REST files enumeration confirmed all 130. Worktree is clean; final exact-delta whitespace passes.

Actual GitHub review submitted with explicit commit_id and COMMENT event: https://github.com/relux-works/curator-spec/pull/50#pullrequestreview-5212030884 (review ID 5212030884, state COMMENTED). Verdict ACCEPT, no material findings. This author-account comment review is not a platform approval. Hosted checks and landing remain the parent owner's responsibility. Native response and complete remote evidence are under `.temp/publication-review/`.
