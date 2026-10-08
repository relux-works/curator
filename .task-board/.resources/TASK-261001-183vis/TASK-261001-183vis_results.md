# TASK-261001-183vis — diagnose-validation-suite-drift

Research handoff, 2026-10-01 04:18Z. Read-only diagnosis; no landing, source/config edit, worktree mutation, binary installation, or daemon restart performed. Findings are ready for review.

## Root cause

The installed runner rejects an **ignored, untracked validation-policy source**, even though the recorded and currently configured validation identities match. `TASK_BOARD_CONFIG` and the landing manifests select `/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/orchestration/task-board.config.json`. Git ignores `.temp/` (`.gitignore:12`), so that file exists on disk but has no blob in either the accepted base/evidence trees or current trunk.

The exact failing path in the installed source is:

1. [cmd/worktree_integrate.go:305](/Users/administrator/.curator/cache/project-management/8f10a12bc16dcec542261b7ec64f76afafbcb8bc/snapshot/tools/board-cli/cmd/worktree_integrate.go:305) resolves the config path; lines 309–317 load commands and environment identity from disk; lines 343–360 pass them to `integration.Integrate`.
2. [internal/integration/integrate.go:543](/Users/administrator/.curator/cache/project-management/8f10a12bc16dcec542261b7ec64f76afafbcb8bc/snapshot/tools/board-cli/internal/integration/integrate.go:543) calls `selectReviewedValidationSuiteAt(record, state.HeadOID)`.
3. [reviewed_suite.go:111](/Users/administrator/.curator/cache/project-management/8f10a12bc16dcec542261b7ec64f76afafbcb8bc/snapshot/tools/board-cli/internal/integration/reviewed_suite.go:111) reads that same path from the authoritative tree. `validationPolicyAt`, lines 192–198, returns nil when a successful `git ls-tree` has no entry.
4. [reviewed_suite.go:25](/Users/administrator/.curator/cache/project-management/8f10a12bc16dcec542261b7ec64f76afafbcb8bc/snapshot/tools/board-cli/internal/integration/reviewed_suite.go:25) defines `matchesRequest`: nil fails immediately. Lines **115–116** therefore refuse with the message defined at line 98, before the evidence comparison at lines 120–127.

**Differing component:** the loaded disk policy is present, whereas the immutable-tree policy object at the configured relative path is **nil/absent**. There is no differing command hash or environment hash. Describing this as actual suite drift would be incorrect; the generic refusal conceals a source-presence regression.

## Identity definition and computed values

[spawn_worktree.go:322](/Users/administrator/.curator/cache/project-management/8f10a12bc16dcec542261b7ec64f76afafbcb8bc/snapshot/pkg/remoteconfig/spawn_worktree.go:322) computes suite SHA-256 over Go JSON encoding of the ordered command array. If `command_timeout_seconds != 0`, it instead hashes the object with fields `commands` and `command_timeout_seconds`, in that order. It does not hash config mtime, the entire config, executable bytes, daemon version, workflow bytes, or ambient shell variables.

`environment_sha256` is an explicitly configured string, not a runtime fingerprint of `.zshenv` or `os.Environ()`. The current config omits it, yielding the empty Go string. `suite_id` is the selected profile (`spawn` or `remote.spawn`) plus `.worktree_isolation.validation.commands`; see [spawn_worktree.go:519](/Users/administrator/.curator/cache/project-management/8f10a12bc16dcec542261b7ec64f76afafbcb8bc/snapshot/pkg/remoteconfig/spawn_worktree.go:519) and [reviewed_suite.go:18](/Users/administrator/.curator/cache/project-management/8f10a12bc16dcec542261b7ec64f76afafbcb8bc/snapshot/tools/board-cli/internal/integration/reviewed_suite.go:18).

For **TASK-261001-1klixs — carry-second-operator-guide**, revision 1:

| Component | Recorded CR validation | Current integrate request |
|---|---|---|
| Commands | Proven by matching digest below | `["sh scripts/remote-gate.sh"]` |
| Suite ID | `spawn.worktree_isolation.validation.commands` | Same effective profile; immutable-source gate refuses before setting its session ID |
| Suite SHA-256 | `b23106a78e4b40f6bbe2ee890ec88e5acb536c6d8bcd3321a7ca1345c182c952` | `b23106a78e4b40f6bbe2ee890ec88e5acb536c6d8bcd3321a7ca1345c182c952` |
| Environment identity | Field absent → `""` | Field absent → `""` |
| Config-source object in Git | Absent at configured path in evidence tree | Absent at configured path in trunk; disk policy is present |

Evidence: [CR record:28](/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/changerequests/TASK-261001-1klixs/rev-000001.json:28), suite fields at lines 27495–27496. Validation tree: `891d805d4bca6262d870c02e46cff265db5a0955`; source tree: `9985f174fa989d081464dd2f4ab5faad838d96c1`; base commit: `bab2433ba115a7eafb2298dba6ef16154e2cc62e`; base tree: `7fdb0fdd76d444c882ace12ae4888ab257943424`.

The broader `inputs_sha256` also binds the suite ID, suite hash, environment string, `dependencies_recorded`, and path-sorted dependency entries ([source_tree.go:279](/Users/administrator/.curator/cache/project-management/8f10a12bc16dcec542261b7ec64f76afafbcb8bc/snapshot/tools/board-cli/internal/changerequest/source_tree.go:279)). Recomputing it with the current policy fields and recorded dependency set gives the exact recorded value:
`sha256:b64fa477b02dd13127f6bbd06eaacdcb4436add31f06661abf5a74ed23932f4d`.
This comparison does not claim the dependency set was freshly re-enumerated from today's trunk. That later binding gate is not reached by this refusal.

The CR's actual validation timestamp is **02:03:25Z**, with exit status **0**, and acceptance is **02:47:48Z**. The brief's 03:04Z timestamp does not describe this CR's validation. These are existing recorded results, not a suite rerun by this researcher.

## Read-only reproduction and checks

The standalone `python3 /tmp/TASK-261001-183vis/verify.py` exited **0**. It independently computed the command and input digests, asserted their equality to the CR, and checked successful Git reads. Its source is embedded below for reproduction.

For each of these four Git objects, `git ls-tree <object> -- .temp/orchestration/task-board.config.json` exited **0** with empty output (proven absence, not a read failure):

- Current trunk / failed landing trunk: `c803afd7e9ed9fe016f85be10fc366cd6dccc0e1`.
- 1klixs base: `bab2433ba115a7eafb2298dba6ef16154e2cc62e`.
- 1klixs evidence tree: `891d805d4bca6262d870c02e46cff265db5a0955`.
- Successful 1kylpg landing's prior trunk: `5432c85f61aff7776ae69b7215093d3c47562e1b`.

`git show <object>:task-board.config.json` exited **0** for all four objects; that tracked alternative has the same validation suite and empty environment identity. `git ls-remote --symref origin HEAD` exited **0**, advertising `main` at `c803afd7…`; local `HEAD` and `refs/heads/main` equal that OID. No fetch or ref mutation was needed.

Current orchestration config and `.current-hold` are byte-equal, SHA-256 `69e197d0eb17ebd29552c49eddd6164138180af33fc5d517228de096183185bf`. `git check-ignore -v` exited **0** and identified `.gitignore:12`.

No product validation suite, integration, or regression test was executed in this read-only task. Existing regression tests and historical suite results are cited as existing evidence, not claimed as newly green. Navigation misses/unsupported CLI probes were not gates and are not presented as successful validation.

## Why success stopped; candidate causes

The [installed receipt](/Users/administrator/.curator/global/skills/project-management/.csk-install.json) maps executable cache `f32104b46f09699c79ba79adf893d86453ab1e15cb5e86df6bf70d9820510a51` to source commit `8f10a12bc16dcec542261b7ec64f76afafbcb8bc`, artifact SHA-256 `88bdf62c937bb87c000a8d9bfa3eb86ae4cd6d601bab9d0d445a493b9debd93c`, installed **2026-09-30 20:22:15Z**. Its build receipt has the matching source digest `610f629c1afbd1cacc6aaa4a1b2443a907ec81c309976b66aa80e16a78b44d1d`.

The prior implementation returned early on equal suite hashes (`reviewed_suite.go:17–18` in cached older snapshots, or commit `6c063c7a^`). Commit `6c063c7a` introduced the tree-aware gate; it is an ancestor of installed `8f10a12b` (`git merge-base --is-ancestor` exited **0**). Therefore the old implementation permits this ignored-config/no-drift case while the installed implementation rejects it deterministically.

The successful run ended at 18:05Z, over two hours before the recorded installation. Its manifest selects the same ignored config path and its CR has the same suite hash and empty environment string. This supports **binary upgrade exposing the untracked-source regression** as the timeline explanation. However, historical manifests do not record the task-board executable hash: the exact historical binary of that terminated successful process is **unknown**. The brief's claim that all historical runs, including the success, used the current cache entry is not established by current process paths and conflicts with the install timeline. Do not promote that claim to fact.

- **Daemon/runner version skew:** a restart alone cannot repair this deterministic installed-code branch. [spawn_integration_landing.go:17](/Users/administrator/.curator/cache/project-management/8f10a12bc16dcec542261b7ec64f76afafbcb8bc/snapshot/tools/board-cli/cmd/spawn_integration_landing.go:17), lines 40–52, invokes the runner's `os.Executable()` as the foreground public integrate command. No daemon version participates in the identity comparison. Daemon left untouched.
- **Environment changes:** `TASK_BOARD_CONFIG` matters because it selects the ignored source. No evidence of a changed environment hash; ambient `.zshenv` contents are not hashed by this gate.
- **Workflow change:** `.github/workflows/ci.yml` may matter to source binding/revalidation downstream, but it does not participate in this policy-source lookup and cannot explain the refusal at line 115. This conclusion does not certify downstream gates will pass after repair.
- **mtime/full-config changes:** neither is part of suite identity, and byte equality was independently checked.

All outcomes below were read from `.temp/spawn-runs/<RUN>/state.json`; no landing was retried:

| Run | Recorded terminal UTC | Exit | Result |
|---|---|---:|---|
| RUN-260930-a223cf | Sep 30 18:05:13 | 0 | integration succeeded |
| RUN-260930-dec728 | Sep 30 20:41:00 | 1 | `integration_runner_failed`; empty detail, exact cause unknown |
| RUN-260930-96f5cd | Sep 30 21:04:31 | 1 | `validation_suite_changed` |
| RUN-260930-bf7687 | Sep 30 21:31:46 | 1 | `validation_suite_changed` |
| RUN-260930-b10079 | Sep 30 22:01:06 | 1 | `validation_suite_changed` |
| RUN-261001-61ec44 | Oct 1 02:12:14 | 1 | `validation_suite_changed` |
| RUN-261001-7fed72 | Oct 1 03:46:16 | 1 | `validation_suite_changed` |
| RUN-261001-e841c2 | Oct 1 04:12:20 | 1 | `validation_suite_changed`; trunk unchanged |

The first failure started at 20:22:50Z; it did not terminate at 20:22Z. Its empty failure detail cannot honestly be relabelled as proven suite drift. The additional run's provider attempt exited 0, but the landing/run exited **1** with completion rejected. Both latest refusals left trunk at `c803afd7…`.

## Smallest legitimate fix and handoff recommendation

A fix already exists in locally available source history: **`c5ff3d193e432f53ab69c7698841733503c8b71b`**, PR #474, **BUG-260930-1dh28f — integrate classifies an untracked validation policy like the no-path branch**. The commit records accepted review; this researcher inspected its exact source diff, not its full historical test execution. It is absent from installed `8f10a12b`.

Its 25-line production change distinguishes an untracked config absent in **both** base and current tree from deletion of a previously tracked policy. For the former it compares recorded/current suite and environment identities; actual drift still refuses. Git read errors still propagate, tracked-source behavior remains strict, and the remaining tree/evidence/revalidation gates still execute. See [fixed reviewed_suite.go:115](/Users/administrator/Developer/ReluxWorks/skill-project-management-validation-fixes/tools/board-cli/internal/integration/reviewed_suite.go:115), [negative drift tests:89](/Users/administrator/Developer/ReluxWorks/skill-project-management-validation-fixes/tools/board-cli/internal/integration/untracked_validation_policy_test.go:89), [tracked deletion test:167](/Users/administrator/Developer/ReluxWorks/skill-project-management-validation-fixes/tools/board-cli/internal/integration/untracked_validation_policy_test.go:167), and [public Cobra entry test:49](/Users/administrator/Developer/ReluxWorks/skill-project-management-validation-fixes/tools/board-cli/cmd/worktree_integrate_untracked_policy_test.go:49).

**Recommended human host step:** between managed sessions, deploy a task-board build containing that already-reviewed fix through the normal installer workflow, then start fresh runner sessions on the fixed build. Coordinate daemon replacement only outside active managed sessions. Do not merely restart the existing buggy build. No install/restart was performed here, and no validation bypass is proposed.

An alternative with no binary change is to reconcile the desired orchestration settings into the already tracked `task-board.config.json` through review and start a newly bound session using that tracked source. Its validation policy already matches all four inspected trees. Do not silently switch existing frozen run manifests or discard their other orchestration settings. Merely retrying handoff, invalidating identical validation evidence, touching the config, or rerunning the suite does not fix the missing-source branch.

The next implementation/operations slice is deployment of the existing reviewed fix, then one normal tracked landing with fresh authority and all ordinary checks. There is no need for another research prerequisite or a new bypass flag.

## Reproduction source

```python
import json, hashlib, subprocess
from pathlib import Path
root=Path('/Users/administrator/Developer/ReluxWorks/curator/curator')
source=Path('/Users/administrator/.curator/cache/project-management/8f10a12bc16dcec542261b7ec64f76afafbcb8bc/snapshot')
rel='.temp/orchestration/task-board.config.json'
v=json.loads((root/'.temp/changerequests/TASK-261001-1klixs/rev-000001.json').read_text())['validation']
c=json.loads((root/rel).read_text())['spawn']['worktree_isolation']['validation']
def digest(value):
 return hashlib.sha256(json.dumps(value,separators=(',',':'),ensure_ascii=False).replace('<','\\u003c').replace('>','\\u003e').replace('&','\\u0026').encode()).hexdigest()
identity=c['commands'] if not c.get('command_timeout_seconds') else {'commands':c['commands'],'command_timeout_seconds':c['command_timeout_seconds']}
suite=digest(identity)
print('commands:',json.dumps(c['commands']))
print('recorded suite:',v['suite_sha256'])
print('current suite: ',suite)
print('recorded/current environment:',repr(v.get('environment_sha256','')),repr(c.get('environment_sha256','')))
assert suite==v['suite_sha256'] and c.get('environment_sha256','')==v.get('environment_sha256','')
inputs={'suite_id':v['suite_id'],'suite_sha256':suite,'environment_sha256':c.get('environment_sha256',''),'dependencies_recorded':v['dependencies_recorded'],'dependencies':sorted(v['dependencies'],key=lambda x:x['path'])}
recomputed='sha256:'+digest(inputs)
print('recorded inputs:',v['inputs_sha256']); print('recomputed inputs using recorded dependencies:',recomputed)
assert recomputed==v['inputs_sha256']
refs=['c803afd7e9ed9fe016f85be10fc366cd6dccc0e1','bab2433ba115a7eafb2298dba6ef16154e2cc62e',v['tree_oid'],'5432c85f61aff7776ae69b7215093d3c47562e1b']
for ref in refs:
 result=subprocess.run(['git','-C',str(root),'ls-tree',ref,'--',rel],text=True,capture_output=True)
 print('ls-tree',ref,'exit',result.returncode,'source',repr(result.stdout))
 assert result.returncode==0 and result.stdout==''
 tracked=subprocess.run(['git','-C',str(root),'show',ref+':task-board.config.json'],text=True,capture_output=True)
 print('tracked config read exit',tracked.returncode)
 assert tracked.returncode==0
 vc=json.loads(tracked.stdout)['spawn']['worktree_isolation']['validation']
 assert digest(vc['commands'])==suite and vc.get('environment_sha256','')==''
 print('tracked config suite:',digest(vc['commands']))
a=(root/rel).read_bytes(); b=(root/(rel+'.current-hold')).read_bytes()
assert a==b
print('current/held config byte equality: true; sha256',hashlib.sha256(a).hexdigest())
code=(source/'tools/board-cli/internal/integration/reviewed_suite.go').read_text()
assert 'if !current.matchesRequest(s.req)' in code and 'return p != nil && p.suite != nil' in code
print('Installed source: absent git entry returns nil; matchesRequest(nil) is false; refusal at reviewed_suite.go:115-116')
print('Read-only verification PASS; no product gate or landing invoked')
```
