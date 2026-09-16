# TASK-260910-1wjst3 — review verdict, revision 1

Verdict: **changes_requested**. Route to **to-dev**, then a new independent review.
Reviewed CR-TASK-260910-1wjst3-1 revision 1. No candidate code or spec files modified.

## Candidate identity and empty curator delta

The runtime CR has an empty curator repository delta. That is appropriate for this leaf: its deliverable is in curator-spec, not curator. An empty curator delta is not itself a failure or an acceptance. Reviewed the actual curator-spec Story worktree `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-2awkzu/worktree` and attached `TASK-260910-1wjst3_spec-patch_rev1.patch` instead.

`git patch-id --stable` for both the attached patch and the worktree delta against origin/main is `9df274c051142d331ffb235d2f59893a0df2e11d`. To preserve the reviewer's read-only index, included the untracked new vector using `git diff --no-index /dev/null conformance/v1/vectors/shell-hook-trust.json`, combined with `git diff origin/main`, rather than `git add -N`.

Six changed paths: CHANGELOG.md, cli/curator.md, profiles/manager.md, conformance/v1/manifest.json, conformance/v1/vectors/shell-hook-trust.json, release/1.0.0-rc.9.json. The release descriptor changes only the two derived manifest pins. No implementation code, proposals, tags, or unrelated rule edits. Read both audits: spec S6 and manager S6/I1.

## Required corrections

### R1 — CLI guidance contradicts the first shipping rollout profile (medium)

`cli/curator.md:100` says “The hook sources a project `.agents/env.sh` or `.agents/env.ps1` only when its bytes are trusted”. It gives no Revision A exception. In contrast, `profiles/manager.md:1181` explicitly says Revision A “keeps the old sourcing behavior (it sources unapproved and changed files)” and ships first. The guide therefore promises refusal during a release that deliberately continues sourcing. Qualify the CLI paragraph by the two labelled revisions, explicitly retain sourcing plus warning/migration hint in A, and state skip plus warning in B.

### R2 — emitted-hook conformance surface and negative evidence are incomplete (high)

The new JSON describes 12 state/outcome combinations but never supplies candidate file bytes, actual manager-state approval inputs versus project-supplied records, or a procedure for exercising the emitted hook and observing sourcing/warnings. For example, `conformance/v1/vectors/shell-hook-trust.json:195` declares `approval: present-mismatch`, two hashes, and `sourced: false`; no consumer can reproduce the observed digest from fixture bytes in this case. The once-per-session flag similarly supplies no repeated activation sequence or observable warning count. No added normative text explicitly identifies emitted hook code as the conformance surface required by brief item 4. No reference to this vector exists in `tools/` (search for `shell-hook-trust`); the validator only picks it up through generic manifest integrity.

Measured negative check, isolated from the candidate: changed ONLY `sourced` to true in `changed-env-sh-B-enforcing-not-sourced` (narrowing enforcement by admitting changed POSIX files while retaining all other cases); regenerated the manifest and ran `python3 tools/validate.py`. Exit 0: `validated 60 schemas and 1048 vector files`. Thus the fast gate accepts this contradiction of MUST NOT source. This is not a claim that the manager implementation was tested: that implementation is explicitly out of scope.

Correction: declare the emitted-hook conformance surface and provide reproducible input/expected-output vectors (file bytes, manager-state records or absence, selected profile, activation sequence, source side effect and warning path/command/count), covering approved/unapproved/changed in both profiles for both env file types. Include a project-local forged approval that cannot authorize bytes. Tie the vector contract to the conformance checking mechanism, or explicitly document and assign the downstream execution binding if implementation remains separate. Report integrity validation separately from semantic/emitted-hook execution; do not present manifest inclusion as exercising the gate. Preserve the settled rollout and avoid manager implementation changes in this spec leaf.

Coverage: the declared state matrix is **12/12** (3 states × 2 profiles × 2 files). Emitted-hook executions in this spec review are **0/12**; the vector has no execution recipe/byte fixtures. The one targeted semantic mutant was detected **0/1** times by the fast validator. This does not establish other mutation or runtime coverage.

## Brief conformance, item by item

All paths/lines refer to the curator-spec candidate.

| Requirement | Evidence and assessment |
| --- | --- |
| Normative trust rule, both files | `profiles/manager.md:1091`: “exactly `.agents/env.sh` … and `.agents/env.ps1`”; `:1096`: “Under the enforcing rollout profile … MUST source … ONLY when its exact bytes are trusted”. Pass; explicit A exception at :1101. |
| Manager-recorded or operator-approved only | `profiles/manager.md:1104`: “manager itself generated or installed … recorded its digest”; :1106 “operator approved … once”. Pass. |
| Changed/unknown, reapproval, warning and continue | `profiles/manager.md:1109`: “MUST NOT be treated as trusted”; :1112 “Re-approval is REQUIRED”; :1116 “MUST warn once per shell session”; :1118 “MUST continue without failing activation or the prompt”. Pass. |
| Closed approval record, storage and byte digest | `profiles/manager.md:1124`: “manager state below `<manager-home>`”; :1127 “MUST NOT read an approval record from package data, project data, or profile data”; :1132 “exactly these four members”; :1134–1138 path, sha256, approved_by, approved_at; :1135 “64 characters, no prefix”; :1138 “RFC 3339”. Pass. The location is implementation-selected within manager home as §1 permits. |
| CLI approve/list/revoke | `profiles/manager.md:1150`, :1153, :1154 and `cli/curator.md:50` contain the identical three command spellings. :1156 requires failure without recording on absent/unreadable files; :1158 specifies unknown revoke; :1160 forbids listing mutation. Command rows pass; developer-shell guidance needs R1. |
| Diagnostics closed set | `profiles/manager.md:1169` / :1170 list `shell_hook_env_unapproved` / `shell_hook_env_changed`; :1164 closes the set. Exact spellings match CLI, changelog, vector header and cases. No new knob or lock key; environments §12.1/§12.2 additions are not applicable. |
| Two explicit rollout revisions | `profiles/manager.md:1181` “Revision A (`A-warning`, warning release)”; :1185 “ships this revision first”; :1186 “Revision B (`B-enforcing`, flip release)” and “MUST NOT source”. Pass in normative profile; R1 remains in CLI. |
| Status posture | `profiles/manager.md:1197` “Read-only status MUST report”; :1199–1203 enumerate approved/unapproved/changed, path, approver, and --check behavior. Pass. |
| Vectors and manifest | `conformance/v1/vectors/shell-hook-trust.json:58` begins 12 cases; `conformance/v1/manifest.json:4180` registers the vector with its digest. Matrix is present, but emitted-hook surface and reproducible negative evidence need R2. |
| Approval schema conditional | `profiles/manager.md:1140` explicitly states text-only normative shape. Existing schema families inspected; no existing shell-hook manager-home approval record family. Accept the text-only option for this revision rather than adding an unsolicited schema. |
| Changelog | `CHANGELOG.md:10` starts “S6: shell-hook project env trust gate”; :25 onward names A warning then B enforcing. Pass. |
| Security/trust summary cross-link | No standalone security/trust summary exists in this profile; conditional requirement is not applicable. Shell and CLI references are present. |
| Scope and architecture | Fits manager-local state versus untrusted project data. Only six spec/derived paths changed. No new config key, lock key or existing schema member; existing closed sets unchanged. |
| Evidence and validation | Producer patch/evidence resources read. Patch equality independently established. Independent commands and limits below. No acceptance inferred from producer results. |

## Independent validation

Shell: zsh. Candidate command:

```sh
set -o pipefail
PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make validate
```

`make validate` exit **0**, independently rerun on the candidate. All three gates completed:

```text
python3 tools/validate.py
validated 60 schemas and 1048 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
[progress dots omitted]
----------------------------------------------------------------------
Ran 227 tests in 188.324s

OK
go test ./tools/...
ok  github.com/relux-works/curator-spec/tools/generate-vectors (cached)
```

Python gates reran independently; Go reported its existing cache result. No producer transcript was substituted for these commands. These green gates do not resolve R1/R2.

`git diff --check`: exit 0, no output.

`make regenerate-check`: exit 0 in a disposable byte copy of the candidate with a temporary Git baseline of that candidate (needed because this target regenerates files and compares against Git, while the real candidate is intentionally uncommitted and review-only). Output:

```
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
```

The subsequent adversarial check changed only the disposable copy:

```
make regenerate                       # exit 0
go run ./tools/generate-vectors -root .
python3 tools/validate.py              # exit 0, despite invalid sourcing outcome
validated 60 schemas and 1048 vector files
```

No `LOGBOOK.md` edit: campaign instructions explicitly prohibit it. Findings are persisted in this task-scoped verdict and board notes. No human decision or external blocker is needed; these are ordinary producer rework findings.
