# TASK-261001-3fgu9f review verdict, rev2: ACCEPTED

CR-TASK-261001-3fgu9f-2, base d0920353, candidate tree 6cab337e1d0ad998e9c1413c8420a4a8bd941207
(worktree tree re-derived with `git add -A; git write-tree` = same OID, 16 paths).

## rev1 findings: all fixed
- LOGBOOK.md absent from worktree and from the 16 changed paths.
- `git diff d0920353 -- go.mod go.sum '*.golden'` is empty (0 bytes). go.mod pins skill-agents-management v0.5.22.
- Added-lines grep for `prompt_suggestion|prompt-suggestion|0.5.33` finds nothing. CHANGELOG, README (14, 51) and SPEC (386) name v0.5.22.
- Windows vet: baseline (a worktree at d0920353) and candidate both exit 1. Their outputs are identical modulo line order (`diff <(sort base) <(sort cand)` is empty). They are pre-existing syscall.Mkfifo/SYS_IOCTL failures, so the candidate adds no new Windows vet failure.
- Host `go vet ./...` exit 0.

## Scope
- Mapping: internal/mapping/mapping.go adds muse -> {System muse, Provider "muse"}. It is unit-tested and also asserted in TestMuseV3CompositionPreservesHOME.
- v3 reader: internal/fragment/fragment.go accepts IdentityV3. It requires permissions, and v1 still rejects them. It reads four XDG vars that share one parent (/config, /data, /state, /cache). It rejects HOME, a foreign XDG, a relative or ".." path, system_prompt for muse, and muse under v1/v2. TestV3OtherAdaptersRetainV2Rules covers claude/codex/opencode/pi under v3. The v1/v2 existing tests are unchanged and green.
- HOME is never set: the composition overlay carries only fragment XDG literals. The test asserts HOME stays the inherited native value and is not in EnvLiterals.
- Bounds, both tested against today's pin (v0.5.22):
  - `permission_mode_unsupported` is asserted by exact stderr suffix through run (muse_test.go:71).
  - The interactive refusal is asserted by exact error string through plan.Build (muse_test.go:125).
  - Both tests branch on the pinned plugin's real capability (PermissionMapping, SupportsMode(Interactive)), so they flip to admission assertions when a newer pin declares them. The flip branches are unexercised today, so they are not a proof.
- Side change worth noting: internal/defaults/lineup.go registers the muse system and lets systemCandidates use declaration-owned models for the vendor-unresolved muse runtime. This is needed so the muse model rows resolve, and it is covered by internal/defaults/muse_test.go. plan_test.go's TestUnresolvedVendorScope now expects ErrUnknownModel, because the muse system is now registered.

## Evidence run by me (real exit codes)
- `go test -p 1 ./...`: every package ok, including cmd/curator-run (188s), exit 0. `-count=1` reruns of ./internal/... and ./cmd/... both exit 0.
- Mutant A (v3 removed from enumMember in fragment.go): `go test ./internal/fragment` exit 1 (v3_test.go TestMuseV3Fragment and TestV3OtherAdaptersRetainV2Rules fail), and `-run Muse ./cmd/curator-run` exit 1 (resolve_fragment_invalid in TestMuseV3ThroughRunPermissionBound and TestMuseV3PermissionTransportThroughRun). KILLED.
- Mutant B (composition overlay sets HOME = managed parent for muse): `go test -run Muse ./cmd/curator-run` exit 1 (TestMuseV3CompositionPreservesHOME, "HOME replaced"). KILLED.
- Both mutants reverted (diff against saved originals is clean). The restored tree re-greens with exit 0 and the tree OID is still 6cab337e.

## Hygiene
No stray files (all changes are within the 16 paths). No employer name added. No commit made.

Bound to state: the v3 conformance schema check is exercised only through the canned fixture, not the curator-spec schema file.
