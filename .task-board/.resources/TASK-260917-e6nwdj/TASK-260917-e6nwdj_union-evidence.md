# TASK-260917-e6nwdj — review evidence

## Four composed protocol hunks

### Hunk 1

Accepted S5 conflict side:
```text
entry of the immutable profile store, link-target identity is necessary but
no longer sufficient currency: the store entry's integrity is verified, not
assumed (section 4). A copied surface —
```

Landed E5 conflict side:
```text
entry of the immutable profile store, link-target identity is sufficient
currency (the store entry's integrity is the store's own invariant, section
4; the link target is read with `lstat`-class semantics, section 8.3.1),
so a launch does not re-hash a large skills tree. A copied surface —
```

Delivery protocol/environments.md:2346–2350:
```text
2346: entry of the immutable profile store, link-target identity is necessary but
2347: no longer sufficient currency: the store entry's integrity is verified, not
2348: assumed (section 4); the link target is read with `lstat`-class semantics
2349: (section 8.3.1). A copied surface —
2350: the `claude_code` root-context file in every mode, or a manager §5
```

### Hunk 2

Accepted S5 conflict side:
```text
found in the home. A store entry is re-applied only after that entry passed
the section 4 contract and its pin hash: a non-trusted entry is never
re-applied, so `env resolve --repair` is not persistence for a tampered
store (audit note E7). Failure classes split (section 4): an enclosing
boundary that cannot be proven refuses every mutating operation before its
first write — nothing is rebuilt, because there is no protected place to
rebuild into, and the operator repairs the boundary out of band; an
individual entry (store entry, lock file, marker file) that fails inside a
proven enclosing boundary is rebuilt by a real operation from the
revalidated snapshot into newly established protected state
(operation-private staging, atomic publication under the mutation lock):
for a `git` member the manager re-acquires the pinned commit's exact
snapshot bytes (section 1.2) and republishes the entry with a fresh
boundary; for a `path` or `local` member there is no second copy of the
snapshot — the source directory is never read again (section 1) — so the
entry cannot be rebuilt and repair fails with
`environment_repair_failed` while the operator reinstalls. Dry-run
evaluation of an entry-class failure reports
`would-rebuild-untrusted-store` and mutates nothing; dry-run evaluation of
an enclosing-boundary failure reports `environment_store_untrusted` with
no rebuild planned and mutates nothing.
Lock acquisition that times out is
```

Landed E5 conflict side:
```text
found in the home. Repair writes are section 8.3.1 writes: repair
replaces directory entries and refuses with
`environment_write_would_follow_link` rather than writing through a link
the manager does not own. Lock acquisition that times out is
```

Delivery protocol/environments.md:2373–2404:
```text
2373: found in the home. Repair writes are section 8.3.1 writes: repair
2374: replaces directory entries and refuses with
2375: `environment_write_would_follow_link` rather than writing through a link
2376: the manager does not own.
2377: A store entry is re-applied only after that entry passed
2378: the section 4 contract and its pin hash: a non-trusted entry is never
2379: re-applied, so `env resolve --repair` is not persistence for a tampered
2380: store (audit note E7). Failure classes split (section 4): an enclosing
2381: boundary that cannot be proven refuses every mutating operation before its
2382: first write — nothing is rebuilt, because there is no protected place to
2383: rebuild into, and the operator repairs the boundary out of band; an
2384: individual entry (store entry, lock file, marker file) that fails inside a
2385: proven enclosing boundary is rebuilt by a real operation from the
2386: revalidated snapshot into newly established protected state
2387: (operation-private staging, atomic publication under the mutation lock):
2388: for a `git` member the manager re-acquires the pinned commit's exact
2389: snapshot bytes (section 1.2) and republishes the entry with a fresh
2390: boundary; for a `path` or `local` member there is no second copy of the
2391: snapshot — the source directory is never read again (section 1) — so the
2392: entry cannot be rebuilt and repair fails with
2393: `environment_repair_failed` while the operator reinstalls. Dry-run
2394: evaluation of an entry-class failure reports
2395: `would-rebuild-untrusted-store` and mutates nothing; dry-run evaluation of
2396: an enclosing-boundary failure reports `environment_store_untrusted` with
2397: no rebuild planned and mutates nothing.
2398: Lock acquisition that times out is
2399: `environment_lock_unavailable`, distinct from `environment_repair_failed`,
2400: which keeps meaning that the store cannot restore this home — an entry is
2401: missing or fails validation, including the section 4 contract. Neither emits a fragment.
2402: 
2403: **Store-trust verification.** On every resolve, as part of the lock-free
2404: verification, the manager MUST verify in order: enclosing boundary →
```

### Hunk 3

Accepted S5 conflict side:
```text
switched, stale, store-untrusted (`environment_store_untrusted`),
refused-provider (section 11), or
unreadable state is non-current; unreadable evidence is
```

Landed E5 conflict side:
```text
switched, stale, refused-provider (section 11), link-blocked
(section 8.3.1), or unreadable state is non-current; unreadable evidence is
```

Delivery protocol/environments.md:2814–2817:
```text
2814: switched, stale, store-untrusted (`environment_store_untrusted`),
2815: refused-provider (section 11), link-blocked (section 8.3.1), or
2816: unreadable state is non-current; unreadable evidence is
2817: reported as unreadable, never as absence (section 8.4). A section 11
```

### Hunk 4

Accepted S5 conflict side:
```text
`--all` confirming every profile of the run; and the section 4
protected-boundary cases
(`vectors/environments-store-boundary.json`) — the intact resolve that
emits a fragment; the swapped system-prompt bytes, swapped root-context
bytes, symlinked entry root, wrong ownership, wrong permissions,
containment escape, non-regular component, and pin-hash mismatch cases
that refuse with `environment_store_untrusted` and emit no fragment; the
environments-root and store-root enclosing-boundary cases that refuse with
no rebuild; the intact-updated-store with old marker case that reports
`environment_home_stale` and repairs, and the swapped-updated-store with
old marker case that refuses with `environment_store_untrusted` and is
never adopted; the unprovisioned-home intact and swapped cases; the
unreadable-marker case that reports
`environment_marker_unreadable`; the dry-run `would-rebuild-untrusted-store`
entry-class case and the enclosing no-rebuild case; the repair rebuild,
entry-rebuild, enclosing-refusal, stale-repair, and unprovisioned cases;
the `env status` non-current posture rows naming the failing check and the
boundary; and the negative cases whose fragment-emitting,
current-reporting, or re-applying observation is non-conforming. The nine
```

Landed E5 conflict side:
```text
`--all` confirming every profile of the run; and the section 8.3.1
write-discipline vectors
(`vectors/environments-write-nofollow.json`) — the symlinked-target
takeover (replaced with backup under authorization, stopped with
`environment_foreign_manager_detected` without), the symlinked-parent
refusal with `environment_write_would_follow_link` under both
authorization states, the post-provisioning planted-link repair, the
manager-owned-link replace, the symlinked-backup-destination refusals (a traversed parent link and a directly symlinked target),
the inside-pointing-link ledger refusal, and the clean-path and
recorded-file positives — every foreign-link case asserting the link's
former target is byte-identical afterwards. The nine
```

Delivery protocol/environments.md:3038–3067:
```text
3038: `--all` confirming every profile of the run; the section 8.3.1
3039: write-discipline vectors
3040: (`vectors/environments-write-nofollow.json`) — the symlinked-target
3041: takeover (replaced with backup under authorization, stopped with
3042: `environment_foreign_manager_detected` without), the symlinked-parent
3043: refusal with `environment_write_would_follow_link` under both
3044: authorization states, the post-provisioning planted-link repair, the
3045: manager-owned-link replace, the symlinked-backup-destination refusals (a traversed parent link and a directly symlinked target),
3046: the inside-pointing-link ledger refusal, and the clean-path and
3047: recorded-file positives — every foreign-link case asserting the link's
3048: former target is byte-identical afterwards. ;
3049: and the section 4
3050: protected-boundary cases
3051: (`vectors/environments-store-boundary.json`) — the intact resolve that
3052: emits a fragment; the swapped system-prompt bytes, swapped root-context
3053: bytes, symlinked entry root, wrong ownership, wrong permissions,
3054: containment escape, non-regular component, and pin-hash mismatch cases
3055: that refuse with `environment_store_untrusted` and emit no fragment; the
3056: environments-root and store-root enclosing-boundary cases that refuse with
3057: no rebuild; the intact-updated-store with old marker case that reports
3058: `environment_home_stale` and repairs, and the swapped-updated-store with
3059: old marker case that refuses with `environment_store_untrusted` and is
3060: never adopted; the unprovisioned-home intact and swapped cases; the
3061: unreadable-marker case that reports
3062: `environment_marker_unreadable`; the dry-run `would-rebuild-untrusted-store`
3063: entry-class case and the enclosing no-rebuild case; the repair rebuild,
3064: entry-rebuild, enclosing-refusal, stale-repair, and unprovisioned cases;
3065: the `env status` non-current posture rows naming the failing check and the
3066: boundary; and the negative cases whose fragment-emitting,
3067: current-reporting, or re-applying observation is non-conforming.  The nine
```

## Mechanical fidelity transcript
```text
TASK-260910-39fzpq_spec-patch_rev2.patch: sha256 88840932af637e7901f9b235fec8f540244959ec397d4ec9eff2049c5eda6783
s5-union-vs-9912db7.patch: sha256 be740eb4e01c89c98b9367d0c513f5440710f63dda004c7e373ee14193935257
attached union equals git diff: True
CHANGELOG.md: merge-file exit 1, equals delivery False
conformance/v1/manifest.json: merge-file exit 1, equals delivery False
conformance/v1/vectors/environments-store-boundary.json: candidate byte-identical
profiles/manager.md: merge-file exit 0, equals delivery True
protocol/environments.md: merge-file exit 4, equals delivery False
release/1.0.0-rc.9.json: merge-file exit 2, equals delivery False
tools/test_validate.py: merge-file exit 0, equals delivery True
tools/validate.py: merge-file exit 0, equals delivery True
 CHANGELOG.md                                       |  52 ++
 conformance/v1/manifest.json                       |   4 +
 .../v1/vectors/environments-store-boundary.json    | 634 +++++++++++++++++++++
 profiles/manager.md                                |  43 +-
 protocol/environments.md                           | 234 +++++++-
 release/1.0.0-rc.9.json                            |   4 +-
 tools/test_validate.py                             | 229 ++++++++
 tools/validate.py                                  | 409 +++++++++++++
 8 files changed, 1578 insertions(+), 31 deletions(-)

Full protocol union equals delivery with the three documented punctuation/whitespace joins: True
CHANGELOG conflict is literal landed + candidate concatenation: True
Candidate file inventory equals delivery diff: True (8/8)

```
