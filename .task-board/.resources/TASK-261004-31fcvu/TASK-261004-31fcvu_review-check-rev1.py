import subprocess,re
from pathlib import Path
base='876127f7c714e01092f43c8950dc879421c52461'
candidate='92a37bbd92b9686b6bd384ac95114d7f8d6e2a74'
def git(*a): return subprocess.check_output(['git',*a],text=True)
b=git('show',base+':CHANGELOG.md'); c=git('show',candidate+':CHANGELOG.md'); t=git('show','v0.15.0-rc.2:CHANGELOG.md'); h='## 0.15.0-rc.2 - 2026-09-26'
assert git('diff','--name-only',base,candidate).splitlines()==['CHANGELOG.md']
assert b[b.index(h):].encode()==c[c.index(h):].encode()
assert '## Unreleased\n\n## 0.15.0-rc.3 - 2026-10-04\n' in c
print('PASS: CHANGELOG-only; earlier release bytes unchanged; empty Unreleased and heading')
es=re.findall(r'^- .*?(?=\n- |\n### |\Z)',b.split(h)[0],re.M|re.S)
assert len(es)==32
rows=[]
for i,e in enumerate(es,1):
 if i in [1,2,3,4,21]:continue
 e=e.strip(); p=t.find(e); assert p>=0
 section=re.findall(r'^## .+$',t[:p],re.M)[-1]
 twin=e in b[b.index(h):] or e in t[t.index(h):]
 line=b[:b.index(e)].count('\n')+1
 rows.append(f'| {i} | {line} | {e.splitlines()[0][2:]} | {section.removeprefix("## ")}:{t[:p].count(chr(10))+1} | {"yes" if twin else "NO"} |')
assert len(rows)==27
report='''# TASK-261004-31fcvu — rc3-release-notes: review verdict, revision 1

Verdict: changes_requested. Route to to-dev; do not accept CR revision 1.

Reviewed base `'''+base+'''` and candidate tree `'''+candidate+'''`. Fresh `git ls-remote origin refs/heads/main` returned the same base (exit 0). No repository file was modified by the reviewer. This run is not goal-bound, as confirmed by `task-board spawn goal`.

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
'''+ '\n'.join(rows)+'\n'
Path('/tmp/TASK-261004-31fcvu_review-verdict-rev1.md').write_text(report)
print('FAIL: 27 dropped entries have no released-section twin; verdict artifact generated')
raise SystemExit(1)
