# TASK-260720-8nxlgx review verdict — RUN-260730-e46794

## Verdict

**CHANGES REQUESTED → `to-dev`.**

The implementation is broadly aligned with the protected-cache contract and
its existing gates are green, but two independently reproduced Windows
security boundaries do not yet fail closed. Both are acceptance-criteria
failures, not environment failures or human-only blockers.

The reviewer changed no CocoaSkills product source or repository test file.

## Exact candidate

- Base/HEAD:
  `15860e3f309888845b9271a257fb95f7c2825b56`
- Branch:
  `task/TASK-260720-8nxlgx-protected-windows-build-cache`
- `src/csk/builds/cache.py`:
  `28b977d4bba4e2c75ab646344a95a0cb338cc2420aa332bd3195dffb1482e6f6`
- `src/csk/builds/cache_windows.py`:
  `43ff8994d860dd16109ee385fd77c054cf1e4e984da27c280b2028679b75bbf5`
- `tests/test_build_cache_windows.py`:
  `9ad5540fb3287161f042c3b807ad494406a14e0f7a52e120edd1217f68cd0ec0`
- The exact three hashes matched the native Windows snapshot.
- Signed dependency commits `0d6ad16f` and `540af8ef` are represented in this
  rebased base by patch-equivalent commits `09f2298` and `138ab82`.

## Finding R1 — late hard links are admitted

Severity: acceptance-blocking security defect.

`_open_protected_child_file` checks `number_of_links == 1` only on initial
open. Its final `_revalidate_child` path checks identity and the security
profile, but never checks the retained handle or the selected path's current
link count. `_validate_source_unchanged` has the same omission for a
publication artifact source.

The attached native probe deterministically inserts each hard link after the
relevant byte-verification helper has completed, leaving the file multiply
linked at the final trust boundary. Exact result:

```text
late artifact hard link: inspect=hit
late receipt hard link: inspect=hit
late source hard link: publish=published
VIOLATION: lookup admitted a multiply linked artifact
VIOLATION: lookup admitted a multiply linked receipt
VIOLATION: publication admitted a multiply linked artifact source
```

This violates the requirement that hard-linked receipt/artifact state and
containment races force a miss or fail closed and never admit candidate bytes.
Windows exposes the live link count through
[`FILE_STANDARD_INFO.NumberOfLinks`](https://learn.microsoft.com/en-us/windows/win32/api/winbase/ns-winbase-file_standard_info);
an additional hard link is another directory entry for the same file
([`CreateHardLinkW`](https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-createhardlinkw)).

Required rework:

1. Revalidate `number_of_links == 1` on both the retained receipt/artifact
   handles and the reopened selected paths at the final lookup boundary.
2. Revalidate the publication source link count on the retained handle and the
   selected path at every final source-stability boundary.
3. Add deterministic native Windows regressions that insert the links after
   byte verification and assert no receipt/path/bytes are returned and no
   publication is accepted.

## Finding R2 — inheritable untrusted mutation ACEs are accepted

Severity: acceptance-blocking boundary defect.

`_validate_manager_home_security` ignores an untrusted mutating allow ACE when
`INHERIT_ONLY_ACE` is set. The reviewer added Everyone full control as
`(OI)(CI)(IO)F`; the backend accepted the manager home and returned a normal
miss:

```text
inherit-only Everyone full-control ACE: inspect=miss, flags=(11,)
VIOLATION: manager home accepted an inheritable untrusted mutation ACE
```

Flags `11` are `OBJECT_INHERIT_ACE | CONTAINER_INHERIT_ACE |
INHERIT_ONLY_ACE`. Although the ACE is not effective on the manager home, it
becomes effective on children. Microsoft documents that new objects created
without an explicit security descriptor derive their DACL from inheritable
ACEs ([DACL for a New Object](https://learn.microsoft.com/en-us/windows/win32/secauthz/dacl-for-a-new-object))
and that inherit-only ACEs become effective on child objects
([Access Control Inheritance](https://learn.microsoft.com/en-us/windows/win32/ad/access-control-inheritance)).
The current `os.mkdir` path creates each cache root before applying its exact
protected profile, exposing an untrusted-mutation window.

Required rework:

1. Reject untrusted allow ACEs with mutating rights when they are effective on
   the manager home **or inheritable by files/directories created beneath it**.
2. Do not count inherit-only manager ACEs as effective manager-home grants.
3. Add native negative cases for `(OI)(CI)(IO)` and object-only/container-only
   untrusted mutation grants, proving inspection is untrusted and publication
   creates no cache roots or candidate state.

## Independent green baseline

- Local focused:
  `python -m pytest -q tests/test_build_cache_posix.py tests/test_build_cache_windows.py`
  → `42 passed, 16 skipped`
- Local full pytest → `853 passed, 74 skipped`
- Local strict mypy → `Success: no issues found in 65 source files`
- Task Ruff → `All checks passed!`
- Exact native Windows focused suite → `16 passed`
- Exact native Windows module mypy
  (`--platform win32 --follow-imports skip`) → one source file clean
- `git diff --check` → exit 0

These green gates establish that the requested changes are narrow; the new
native negative probe is the failing security gate.

## Re-review gate

Return to review with:

- the two deterministic native regressions above;
- native Windows focused and full-suite results;
- cross-platform focused/full pytest;
- strict `python -m mypy`;
- task-wide Ruff and diff hygiene;
- a fresh task-scoped implementation evidence artifact.
