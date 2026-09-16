# TASK-260910-1wjst3 — review verdict revision 2

Verdict: **accepted** for CR-TASK-260910-1wjst3-2, revision 2. Route through `accept_cr` to integrating; this is acceptance for producer integration, not a claim of landing.

## Candidate and review scope

The empty curator repository delta is the right outcome: this leaf changes curator-spec in its separate Story worktree, while manager implementation belongs to TASK-260910-1952mz and TASK-260910-3ungjy. Reviewed the actual curator-spec tree at `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-2awkzu/worktree` plus the attached spec patch, not the empty runtime patch as a proxy for delivery.

Attached rev2 patch and independently assembled worktree diff both have stable patch ID `d8a2632131680ddc196f67d810a88fd1468faabb`. The untracked vector was included with `git diff --no-index /dev/null conformance/v1/vectors/shell-hook-trust.json`; no candidate index writes were needed. Reviewed the producer evidence, rev1 verdict, rev1/rev2 patch comparison, and both audits (spec S6 and manager S6/I1).

Eight candidate paths: CHANGELOG.md, cli/curator.md, profiles/manager.md, conformance/v1/manifest.json, conformance/v1/vectors/shell-hook-trust.json, release/1.0.0-rc.9.json, tools/validate.py, tools/test_validate.py. The tools changes are spec validation explicitly authorized by R2, not manager implementation. No proposals, unrelated rules, schemas, tags, or release publication changed. The release descriptor changes only derived manifest pins.

## Brief conformance and regression review

Paths and line numbers below are in curator-spec.

| Item | Evidence and assessment |
|---|---|
| Normative trust gate, both files | Pass. profiles/manager.md:1091 says “exactly `.agents/env.sh`” and “`.agents/env.ps1`”; :1096 says “Under the enforcing rollout profile” and “MUST source ... ONLY when its exact bytes are trusted”. A exception remains explicit. |
| Closed trust sources | Pass. profiles/manager.md:1104 “manager itself generated or installed” and :1106 “operator approved the file once”. |
| Unknown/changed bytes and reapproval | Pass. profiles/manager.md:1111 “MUST NOT be treated as trusted”; :1112 “Re-approval is REQUIRED”; :1116 “MUST warn once per shell session”; :1118 “MUST continue without failing activation or the prompt”. |
| Record location and closed shape | Pass. profiles/manager.md:1124 “manager state below `<manager-home>`”; :1127 “MUST NOT read an approval record from package data, project data, or profile data”; :1132 “exactly these four members”: path, sha256, approved_by, approved_at. Location follows implementation-selected home in §1. |
| Byte digest, timestamp and approvers | Pass. profiles/manager.md:1135 “lowercase hex SHA-256 digest (64 characters, no prefix)” over exact bytes; :1137 “exactly `manager` or `operator`”; :1138 “RFC 3339 timestamp”. |
| Approval commands | Pass. profiles/manager.md:1150 and cli/curator.md:50 list `curator hook approve <path>`, `curator hook approvals`, `curator hook revoke <path>`. :1156 “MUST fail without recording” for absent/unreadable files; :1158 no-record revoke is unchanged state; :1160 “Listing MUST NOT mutate state”. |
| Closed diagnostics | Pass. profiles/manager.md:1164 “closed to exactly these two”; table :1169–1170 names `shell_hook_env_unapproved` and `shell_hook_env_changed`. Same spellings in CLI, changelog, vector, validator. |
| Rollout and R1 | Pass, R1 closed. profiles/manager.md:1181 “Revision A (`A-warning`, warning release)” retains sourcing and migration hint, ships first; :1186 “Revision B (`B-enforcing`, flip release)” skips. cli/curator.md:100 now says A “keeps sourcing” and :106 specifies B, including warning path and approval command. |
| Posture | Pass. profiles/manager.md:1197 “Read-only status MUST report”; :1199 lists each known file as approved, unapproved or changed, with path/approver and --check behavior. |
| Schema and knobs | No new knob or lock key requested; environments §12.1/§12.2 unchanged. Text-only record definition at profiles/manager.md:1140 retained as accepted in round 1; no shell-hook approval schema family added. |
| Conformance and R2 | Pass, R2 closed. profiles/manager.md:1207 “The emitted hook code is a conformance-vector surface”; :1217 “hook MUST ignore” forged project records; :1225 “twice in the same shell session”; :1231 “repository’s gates do NOT execute hooks” with downstream owners at :1239–1241. Vector contains bytes, exact records/absence, profiles, warning observables and execution recipe. |
| Manifest and changelog | Pass. conformance/v1/manifest.json:4180 registers `vectors/shell-hook-trust.json`. CHANGELOG.md:10 starts “S6: shell-hook project env trust gate”; :25 names Revision A then Revision B and migration behavior. |
| Conditional security-summary link | No standalone security/trust summary in this profile; conditional requirement remains inapplicable. CLI references manager section 8. |
| No round-1 regression | §§8.1–8.6 and command rows unchanged between patches. Changes are CLI rollout correction, §8.7, fixture-backed vectors, validator/tests, corresponding changelog sentence and regenerated digest pins. |
| Architecture and scope | Fits manager-local approval authority; project data cannot self-authorize. No runtime protection claimed from this spec review. Integration/merge remains the producer transaction after acceptance. |

## Negative evidence and coverage

Independent in-memory adversarial run with the repository venv (candidate unchanged), exit 0:

```text
POSITIVE: 14/14 cases pass structural validation
DETECTED sourced shell-hook-trust case changed-env-sh-B-enforcing-not-sourced sourcing does not follow its trust state and profile
DETECTED digest shell-hook-trust fixture env-ps1-v1 sha256 does not match its bytes
```

Also reran the round-1 mutant through the real validator entry point in the disposable copy: changed only the named case's `sourced`, ran `make regenerate` (exit 0) to refresh integrity pins, then repository-venv `python3 -B tools/validate.py` (exit **1**):

```text
validation failed: shell-hook-trust case changed-env-sh-B-enforcing-not-sourced sourcing does not follow its trust state and profile
```

Thus rejection is semantic, not a stale manifest checksum. The first mutant changes only `sourced` to true for the changed POSIX B case, narrowing protection while keeping the other branches. Detection is 2/2 targeted mutants; fixture digests checked 4/4; structural case coverage 14/14 (12/12 core states × profiles × file types plus 2/2 forged POSIX profile cases). Runtime hook execution is 0/14 here, explicitly assigned downstream by §8.7. Forged PowerShell execution is not measured. The structural gate is called from tools/validate.py main; the unit suite includes positive control and seven negative tests. Structural evidence is not runtime authorization proof.

## Independent validation

Shell: zsh, `set -o pipefail`. From the actual candidate:

```sh
PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make validate
```

`make validate`: exit **0**, independently rerun on the candidate. Python gates reran; Go reported a cached result. No producer transcript substituted.

```text
python3 tools/validate.py
validated 60 schemas and 1048 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
...........................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 235 tests in 466.761s

OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	(cached)
```

Final candidate diff was byte-identical to the captured pre-validation diff (including the untracked vector).

`git diff --check`: exit 0, no output. `make regenerate-check` runs in a disposable byte copy with a Git index baseline of the candidate, because the target writes generated files and the real candidate is uncommitted/read-only. No commit created.

`make regenerate-check`: exit **0**.

```text
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
```

Review harness notes: first regeneration attempt started before the disposable copy finished and failed with a missing fixture directory; all candidate source files were then recopied, indexed, and the check rerun. An initial direct mutant invocation used system Python without jsonschema; rerun with the required repository venv passed and rejected both mutants. Neither harness failure is evidence of a candidate defect.

No unresolved findings or human-only decision. Campaign instructions prohibit LOGBOOK.md edits; review evidence is persisted on the task. `task-board spawn goal` reported this run is not goal-bound. All acceptance evidence is attached before lifecycle routing.
