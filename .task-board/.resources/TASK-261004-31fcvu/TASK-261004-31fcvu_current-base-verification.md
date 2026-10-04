# TASK-261004-31fcvu — rc3-release-notes: current-base developer outcome

Ready for review. CHANGELOG.md is the only modified repository file; work remains uncommitted.

Applied the attached `rc3-changelog-run3.md` verbatim on the requested current base `876127f7c714e01092f43c8950dc879421c52461`. HEAD and origin/main both equal that checkpoint. The history range is `v0.15.0-rc.2..origin/main` (rc.2 `5dc292671e66c105360e440472df7690644edee8`). Read the full subject list and the readiness plan's “Curator rc.3: payload, pin and release gates” section. Reconstructed the saved map from the prior results resource and verified every current hash/subject pair and administrative commit path boundary; no network fetch or release qualification is claimed in this run.

The release heading follows the latest application instruction and existing style: `## 0.15.0-rc.3 - 2026-10-04`, preceded by an empty `## Unreleased`. Every byte from `## 0.15.0-rc.2 - 2026-09-26` onward matches current origin/main. All 32 base Unreleased entries are accounted: five consolidated into rc.3, and 27 older entries proved verbatim present in the rc.2 tag and deliberately dropped as duplicate history. The ledger below names their original source lines; the newly attached JSON preserves the exact full text of every entry, so every omitted line is reviewable.

## Validation executed directly in this run

| Standalone command | Actual exit | Evidence and bounds |
| --- | --- | --- |
| `python3 /tmp/TASK-261004-31fcvu_validate.py` | 0 | 54/54 substantive commits mapped; 258/258 administrative commits verified board-only; 32/32 Unreleased entries accounted (5 represented, 27 pre-rc.2 duplicates). Checks empty Unreleased, release date/groups, current pin/digest, disabled v2 writer, required known issues, deduplication, exact source ledger, byte-identical earlier sections, and CHANGELOG-only scope. |
| `git diff --check` | 0 | Tracked diff whitespace check. |
| `env -u GOROOT GOMAXPROCS=2 go build -p 2 -o /tmp/TASK-261004-31fcvu-rc3-notes/curator ./cmd/curator` | 0 | Compiles the current assigned base plus CHANGELOG edit. Binary is outside the repository. |

The attached validator was adapted for the normalized heading, clean current-base scope (the readiness files are now tracked), current-origin earlier-section comparison, and reconstructed map input. Added full subject/path checks and per-entry duplicate/source-byte evidence. These updates address the changed workspace rather than weakening the task checks.

No code behavior changed; the task permits only CHANGELOG.md changes. No repository tests were added. The document consistency test ran directly and is attached. Go unit tests, native Windows/stress tests, full conformance, race and release-publication gates were not run in this documentation task. No prior test results are presented as rerun here. The approximately 21,000 hosted passes and two historical failures remain attributed to the readiness plan's cited incident evidence. The local build validates compilation, not release readiness.

## Current-base Unreleased disposition ledger

Line ranges refer to the unchanged origin/main CHANGELOG, before application. A dropped entry's reason covers every line in that range; the companion JSON stores the exact original lines, not merely the abbreviated title below. Empty lines and group headings are structurally replaced by the fresh Unreleased and five rc.3 groups.

| Entry | Source lines | Original entry begins | Disposition/reason |
| --- | --- | --- | --- |
| 1 | 7–7 | Marker v3/v4 readers reject inconsistent external repository identities, substitution kinds, and effective revision widths. | represented: Fixed 3 |
| 2 | 8–8 | Accept the rc.14 candidate conformance digest with exact counts and gap accounting, including the snapshot v2-write gap owned by the pin cut-over; keep rc.13 pinned and v2 writers off. | represented: Changed 1 and Known issues 2 (stale rc.13 wording superseded by released rc.14 pin; v1 writes retained) |
| 3 | 9–9 | Windows executable resolution now proves platform ownership and every component-store hard-link origin before granting the captured System32 exception. | represented: Security 4 |
| 4 | 11–11 | Unix HTTPS askpass requests the secret only after accepting the password prompt, preventing broken-pipe transport errors on refusal paths. | represented: Fixed 4 |
| 5 | 15–32 | R5 script-worker-v1 runtime conformance qualification. All 33 named | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 6 | 33–49 | E4: umbrella provider lookup from trust roots (warning release, | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 7 | 50–64 | R4 script audit warning classes for `script-worker-v1` (manager profile | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 8 | 65–100 | R3 native probes, capability evidence, and preflight for | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 9 | 101–119 | R2 declaration-derived enforcement for `script-worker-v1`. Every enforced | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 10 | 120–136 | R1 script manager/worker invocation path (`script-worker-v1`). The manager | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 11 | 137–145 | Scoped HTTPS credentials for external build repositories. A `build_https` | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 12 | 146–151 | Operator documentation for scoped HTTPS build-repository token sources, | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 13 | 152–155 | Schema-8 first-party module roots for the `go-v1` driver: a build root may | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 14 | 156–182 | S6: shell-hook trust gate — warning release (`A-warning`). The POSIX and | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 15 | 183–192 | CI guard for the GoReleaser rc channel values. | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 16 | 193–205 | E2: direct-only `class: system` modules. Only the system modules of direct | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 17 | 206–223 | S4 warning release (`s4-warn`, audit finding S4): `profile install`, | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 18 | 224–230 | Conformance pin → v1.0.0-rc.12 (`dced9b8`): the hosted gate now runs | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 19 | 234–240 | The install transaction engine now caches canonical namespace resolutions | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 20 | 241–248 | A `go-v1` build root whose `vendor/modules.txt` carries a directory | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 21 | 252–252 | Project install now materializes skills at non-git product roots with a hygiene notice, while unexpected Git failures refuse installation. | represented: Fixed 1 |
| 22 | 254–265 | E4: the user-bin shim directory counts as manager-published — and | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 23 | 266–281 | Git snapshots are extracted from the object database (`git ls-tree -r -z` | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 24 | 282–291 | Git snapshot extraction now folds directory components per component when | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 25 | 293–300 | Status no longer reports a successfully installed schema-8 skill as | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 26 | 301–304 | Garbage collection no longer drops the live build references of a marker v4. | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 27 | 305–308 | A marker document at a readable schema that is nonetheless invalid is now | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 28 | 309–313 | A snapshot-extraction spawn failure of `git cat-file --batch` is now | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 29 | 314–328 | Draft Skillfile acquisition no longer consults user or system Git | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 30 | 329–350 | Draft literal-URL acquisition now builds git's environment from an | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 31 | 351–363 | Audit-registry snapshot verification no longer mistakes a snapshot | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |
| 32 | 364–371 | A `git check-ignore` spawn failure in the managed `.gitignore` gate is no | dropped: Verbatim pre-rc.2 carryover already present in the rc.2 tag; omitted as duplicate release history, per readiness plan. |

## Task logbook (LOGBOOK.md untouched)

- Finding: all 27 dropped carryover entries are present verbatim at rc.2. The readiness plan explicitly calls for post-rc.2 reconciliation instead of promoting old Unreleased prose indiscriminately.
- Finding: the rc.14 pin leaf leaves `EnableV2Writers=false`. rc.3 states v1 production writes and defers atomic migration/v2 writes to rc.4; it does not repeat the stale “keep rc.13 pinned” line.
- Finding: the subject “curator-go-module-roots” carries marker cross-field validation. Its entry follows the actual diff.
- Decision: credential migration is described as existing rc.2 behavior, alongside new Decision 0018/v2 fragments.
- Decision: live-process build-cache protection is distinct from unfixed audit N1's incomplete runtime reference set.
- Decision: fp8vx7 stays a documented risk; audit N1–N4 remain unfixed with #106/remediation links; B3 is excluded.
- Verification: previous workspace limitations are resolved on the requested base. The audit link and rc.14 pin exist here, readiness files are tracked, the earlier sections match current main, and only CHANGELOG.md differs. No version constants or LOGBOOK.md were edited.

