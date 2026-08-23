# TASK-260720-2jfnz6 review verdict — RUN-260730-525645

## Verdict

CHANGES REQUESTED. Route to `to-dev`.

The current candidate closes the previously reported Linux mode-`0000`
capability and cleanup gap, and all declared test/type gates pass. It does not,
however, satisfy the explicit identical-concurrent-winner acceptance criterion
under an adversarial but valid publication schedule.

## Reviewed candidate

- Product worktree:
  `/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-2jfnz6/worktree`
- Branch: `task/TASK-260720-2jfnz6-protected-posix-build-cache`
- Recorded base: `495ad021847529ce5a544dba415ca2fe19949539`
- HEAD: `0d6ad16fce35c1bd8854511e13766cd236908e3b`
- `main`, `origin/main`, and the merge base resolve to the recorded base.
- Current candidate hashes:
  - `src/csk/builds/cache.py`:
    `d5936a40bc6628f37a7b5e485ff5651d11cd38ec939b27cb81e2cf4b9c6f0821`
  - `src/csk/builds/cache_posix.py`:
    `6068d2c772de0a2d9497bbc36def0f6ffe7d87ff26d21cf637208ec13d72a369`
  - `tests/test_build_cache_posix.py`:
    `aedbe7fd6a03c31f30cbc48609bf6124b2686a7d50c90b429af6b251800b8f13`
- Git state remained unchanged throughout review: only the producer-owned,
  unstaged modifications to `cache_posix.py` and
  `test_build_cache_posix.py`.
- This run is not goal-bound and had no operator directives at either
  checkpoint.

## Blocking finding R3-1 — a paused atomic winner is quarantined and then
misclassifies an identical replacement

Publication deliberately keeps a complete stage entry at mode `0700` until
after its atomic no-replace rename (`cache_posix.py:1168-1173`). After the
rename succeeds, it clears `stage_exists` and seals the winner by reopening the
live key (`cache_posix.py:362-380`, `1196-1206`).

A concurrent publisher treats a live mode-`0700` entry as untrusted. It retries
seven times and then quarantines that entry before attempting its own atomic
rename (`cache_posix.py:286-320`). If publisher A is descheduled after its
rename but before sealing for longer than this roughly 35–40 ms retry window,
publisher B therefore:

1. moves A's complete but still-`0700` winner to quarantine;
2. atomically publishes the same receipt and artifact bytes;
3. seals B's live entry to `0500`; and
4. returns `published`.

When A resumes, `_seal_published_entry` reopens the logical key expecting the
private pre-seal mode `0700`. It is now B's valid sealed `0500` winner, so A
raises `_UntrustedState`; `_publish_locked` converts that to
`cache_boundary_untrusted` (`cache_posix.py:424-428`). A never performs the
required identical-byte winner comparison and never returns `reused-winner`.

The deterministic reviewer probe paused only the first call to
`_seal_published_entry`, allowed the second publisher to pass the existing
retry window, then resumed the first. It produced:

```text
first_outcome=error:cache_boundary_untrusted:newly published cache entry mode 500 is not protected mode 700
second_outcome=result:published
final_inspection=hit
seal_calls=2
quarantine_count=1
```

Probe:
`concurrent-seal-window-probe.py`,
SHA-256
`44680607c85d02da52e7c8ce73fcda212c28dad40c1c57dd7532c9fb0ca31d2b`.
Output:
`concurrent-seal-window-probe.log`,
SHA-256
`8f17e90f22f822b20713ee8f3b319846887248d173daf4bc59da04c96f56c01f`.

The existing concurrency test at `tests/test_build_cache_posix.py:578-599`
starts publishers together but does not force this post-rename/pre-seal
window. Its normal scheduling pass therefore does not prove the required
interleaving.

This is not merely a cosmetic status mismatch. The implementation advertises
atomic identical-winner handling across independent backend instances, and the
task explicitly requires the identical loser to be discarded/reused rather
than surfaced as an untrusted-boundary failure. The caller-held guard does not
remove this requirement: the implementation and its focused test deliberately
exercise independent concurrent publishers, and atomic no-replace winner
selection is an explicit acceptance boundary.

## Required rework

1. Keep the winner identity stable across rename and sealing, or otherwise
   detect that the live key was replaced before operating on it. Do not
   blindly reopen and seal a different publisher's entry by logical name.
2. If a post-rename publisher loses the live name, inspect the selected live
   winner. Return `reused-winner` when receipt and artifact bytes are identical;
   return `CacheConflictError` (or the project's stable
   corruption/nondeterminism branch) when they differ.
3. Add a deterministic regression that pauses publisher A after atomic rename
   and before sealing beyond the current retry window. For identical bytes,
   prove one `published`, one `reused-winner`, and a protected final hit. Add
   the equivalent different-byte schedule and prove the conflict branch.
4. Rerun POSIX-focused pytest, strict mypy, full pytest, task-scoped lint/static
   checks, and the supported native-Linux matrix on the exact revised bytes.

## Passing evidence retained

- Fresh reviewer POSIX-focused pytest:
  `39 passed, 1 skipped in 0.35s`, exit 0. The skip is the Linux-only O_PATH
  case.
- Fresh reviewer strict mypy:
  `Success: no issues found in 63 source files`, exit 0.
- Fresh reviewer full pytest with the accepted conformance root:
  `888 passed, 7 skipped in 82.06s`, exit 0.
- `git diff --check` over the current task modifications: exit 0.
- Producer native Ubuntu matrix logs are consistent with the current candidate
  timestamps and hashes: Python 3.11, 3.12, 3.13, and 3.14 each report
  `40 passed`. This closes the prior capability finding but does not exercise
  R3-1.
- Static review otherwise confirmed the dedicated `builds/go-v1` namespace,
  boundary-first parsing, rooted no-follow traversal, effective-UID and exact
  mode checks, singly linked regular receipt/artifact files, canonical receipt
  and complete-input binding, derived artifact path, size/hash checks,
  read-only inspection, staging outside the live namespace, corrupt/untrusted
  quarantine, immutable final modes, and locked later-GC movement.

The reviewer changed no candidate code, index state, commit, branch, remote, or
PR. The probe and review logs exist only under the Curator task-scoped ignored
`.temp` evidence directory.
