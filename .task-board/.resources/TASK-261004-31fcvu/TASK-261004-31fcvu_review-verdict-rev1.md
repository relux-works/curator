# TASK-261004-31fcvu — rc3-release-notes: review verdict, revision 1

Verdict: changes_requested. Route to to-dev; do not accept CR revision 1.

Reviewed base `876127f7c714e01092f43c8950dc879421c52461` and candidate tree `92a37bbd92b9686b6bd384ac95114d7f8d6e2a74`. Fresh `git ls-remote origin refs/heads/main` returned the same base (exit 0). No repository file was modified by the reviewer. This run is not goal-bound, as confirmed by `task-board spawn goal`.

## Findings

1. **P1 — Unproved deletion of release history.** At the rc.3/rc.2 boundary (candidate CHANGELOG.md:132), 27 former Unreleased entries have been dropped using the wrong predicate. None has an exact twin in a released section, either at the current base or at the rc.2 tag. All 27 are present under **Unreleased** in the rc.2 tag. The producer validator searches the entire tag file, so its passing result cannot establish the binding review condition. Examples absent from rc.3 include script-worker R1–R5, the GoReleaser channel guard (base lines 183–192), transaction namespace caching (234–240), undeclared vendor replacements (241–248), and raw-object snapshot extraction/collision handling (266–291). Some other entries have partial representation; this does not validate dropping all 27. Preserve or accurately consolidate the missing material without pretending it first landed after rc.2; correct the disposition ledger and make the validator require a released-section twin or explicit rc.3 coverage for every dropped entry. Leave all existing rc.2-and-older bytes untouched.
2. **P2 — Internal runner identifier in public notes.** Candidate CHANGELOG.md:84–86 includes a concrete internal runner name. Replace it with neutral wording such as “select the explicit self-hosted runner label.” The binding review note forbids internal hostnames. The identifier is intentionally not repeated in this public outcome.

## Swept surfaces and independently executed verification

| Surface | Result |
| --- | --- |
| Exact CR changed paths | PASS: only CHANGELOG.md; LOGBOOK.md unchanged |
| Earlier release preservation | PASS: byte equality from rc.2 heading through EOF |
| Structure | PASS: fresh empty Unreleased, normalized date/heading, five groups |
| Prior Unreleased dispositions | FAIL: 0/27 released-section twins; all 27 found in tag Unreleased, detailed below |
| History subject map | PASS for identity/accounting: 54 mapped non-administrative commits plus 258 commits independently checked to touch only board paths = 312/312 |
| History/content spot checks | 11 commit patches inspected, table below; no additional contradiction identified |
| Pin and hash policy | PASS: CI SPEC_PIN and manifest match notes; production EnableV2Writers=false and WriteVersion returns VersionV1 |
| Known issues | PASS: unresolved Windows flake, rc.4 v2 writer deferral, N1–N4 unfixed with #106, B3 excluded |
| Public wording | FAIL at candidate lines 84–86 |
| Whitespace | git diff --check BASE CANDIDATE exited 0 |

Read the full history subject list, readiness payload/pin/release-gates section, audit finding headings, producer result/map and validator. Independently recomputed the 54/258 history split and verified all mapped hash/subject pairs, rather than adopting its reported counts. The subject map itself does not prove prose completeness; manual review plus the deletion check above found the substantive failure.

Executed Python structural check exited 0; subject/path accounting exited 0; per-entry section-location check exited 0 (diagnostic command); this attached review checker intentionally exits 1 for the failed duplicate requirement. Read-only patch extraction for the 11 spot checks exited 0. No Go build, unit tests, platform stress tests or conformance gates were rerun or treated as reviewer evidence: this is documentation-only. The producer reports build exit 0, but it cannot detect these prose defects. This is ordinary implementation rework, not an external blocker. Findings are persisted here and in board notes; LOGBOOK.md remains untouched as explicitly required.

## Content spot checks

| Commit | rc.3 entry | Inspected evidence |
| --- | --- | --- |
| e4a6a8d5 | Added 1 | Muse adapter/XDG variables and unverified context handling |
| c224404d | Fixed 2 | Live executable enumeration and skip-all-on-uncertainty sweep |
| 6bb3c626 | Changed 1, Known issues 2 | Full rc.14 pin/digest and owned snapshot writer gap |
| a2284441 | Security 4 | Platform owner SID and hard-link origin checks |
| ecc0ce7d | Fixed 3 | External marker cross-field validation and malformed revision cases |
| 7444178d | Changed 3 | Whole Codex config seed revision A and warning/status regression |
| 3f60f7f0 | Changed 4 | Posture revision A, permissive default, hardened policy |
| 0ef54f54 | Fixed 4 | Password-prompt secret request and refusal-path pipe handling |
| 1a57c71c | Added 2 | Fragment v2 and permission shape |
| 28b61779 | Added 3 | Restore-backups planning and restoration |
| 3265bc79 | Changed 2 | Versioned carriers/readers and retained v1 write switch |

## Every dropped entry: independent exact-text search

Line numbers refer to base CHANGELOG.md and the rc.2 tag CHANGELOG.md. “Released twin” searches only rc.2-and-older sections, excluding Unreleased, in both files. Entries 1–4 and 21 are represented in rc.3 and are not dropped.

| Entry | Base start line | Original entry begins | Exact match in rc.2 tag | Released twin |
| --- | --- | --- | --- | --- |
| 5 | 15 | R5 script-worker-v1 runtime conformance qualification. All 33 named | Unreleased:9 | NO |
| 6 | 33 | E4: umbrella provider lookup from trust roots (warning release, | Unreleased:27 | NO |
| 7 | 50 | R4 script audit warning classes for `script-worker-v1` (manager profile | Unreleased:44 | NO |
| 8 | 65 | R3 native probes, capability evidence, and preflight for | Unreleased:59 | NO |
| 9 | 101 | R2 declaration-derived enforcement for `script-worker-v1`. Every enforced | Unreleased:95 | NO |
| 10 | 120 | R1 script manager/worker invocation path (`script-worker-v1`). The manager | Unreleased:114 | NO |
| 11 | 137 | Scoped HTTPS credentials for external build repositories. A `build_https` | Unreleased:131 | NO |
| 12 | 146 | Operator documentation for scoped HTTPS build-repository token sources, | Unreleased:140 | NO |
| 13 | 152 | Schema-8 first-party module roots for the `go-v1` driver: a build root may | Unreleased:146 | NO |
| 14 | 156 | S6: shell-hook trust gate — warning release (`A-warning`). The POSIX and | Unreleased:150 | NO |
| 15 | 183 | CI guard for the GoReleaser rc channel values. | Unreleased:177 | NO |
| 16 | 193 | E2: direct-only `class: system` modules. Only the system modules of direct | Unreleased:187 | NO |
| 17 | 206 | S4 warning release (`s4-warn`, audit finding S4): `profile install`, | Unreleased:200 | NO |
| 18 | 224 | Conformance pin → v1.0.0-rc.12 (`dced9b8`): the hosted gate now runs | Unreleased:218 | NO |
| 19 | 234 | The install transaction engine now caches canonical namespace resolutions | Unreleased:228 | NO |
| 20 | 241 | A `go-v1` build root whose `vendor/modules.txt` carries a directory | Unreleased:235 | NO |
| 22 | 254 | E4: the user-bin shim directory counts as manager-published — and | Unreleased:246 | NO |
| 23 | 266 | Git snapshots are extracted from the object database (`git ls-tree -r -z` | Unreleased:258 | NO |
| 24 | 282 | Git snapshot extraction now folds directory components per component when | Unreleased:274 | NO |
| 25 | 293 | Status no longer reports a successfully installed schema-8 skill as | Unreleased:285 | NO |
| 26 | 301 | Garbage collection no longer drops the live build references of a marker v4. | Unreleased:293 | NO |
| 27 | 305 | A marker document at a readable schema that is nonetheless invalid is now | Unreleased:297 | NO |
| 28 | 309 | A snapshot-extraction spawn failure of `git cat-file --batch` is now | Unreleased:301 | NO |
| 29 | 314 | Draft Skillfile acquisition no longer consults user or system Git | Unreleased:306 | NO |
| 30 | 329 | Draft literal-URL acquisition now builds git's environment from an | Unreleased:321 | NO |
| 31 | 351 | Audit-registry snapshot verification no longer mistakes a snapshot | Unreleased:343 | NO |
| 32 | 364 | A `git check-ignore` spawn failure in the managed `.gitignore` gate is no | Unreleased:356 | NO |
