# CIP-0004 evidence: shell-hook no-source and PATH append

- Task: TASK-261004-1z2pgb — shell-hook-no-source-and-path-append-design
- Parent: STORY-261004-7fglii — shell-hook-path-and-backend-env-hardening
- Date: 2026-10-04
- Companion: [CIP draft](261004_CIP-0004-shell-hook-no-source-path-append.md)
- Scope: source inspection and scratch probes of current behavior; no candidate implementation, product-code changes, account operations, real credential/operator-manager-config reads, or `LOGBOOK.md` edit.

## Identity and limits

The initial researcher recorded a clean worktree at research start. `git rev-parse HEAD main` returned `ca1b776fb580ec0cee0173bf150daf063023aeaa` twice (exit 0). A fresh `git ls-remote https://github.com/relux-works/curator.git refs/heads/main refs/tags/v1.0.0-rc.14` returned that same main OID (exit 0; there was no matching curator tag row). This is a research source identity, not an integration-authority or landing attestation. The recovery run started with exactly these two untracked research artifacts; its fresh main-ref read returned the same OID (exit 0), and its standalone product/logbook diff against main exited 0.

For curator-spec, `git ls-remote https://github.com/relux-works/curator-spec.git refs/tags/v1.0.0-rc.14 'refs/tags/v1.0.0-rc.14^{}'` returned tag object `661bead088186db70c9f57ad0b115305ec2bef35` and peeled commit `43bf0a2506d5c354a73bbc3ea4623d4653db10c7` (exit 0). Text was read with `git show v1.0.0-rc.14:<path>` from the available spec checkout. Its HEAD was the same peeled commit. No checkout, fetch, commit or spec write was needed. Curator's `.github/workflows/ci.yml:35–43` pins this commit as its current corpus.

Host observations: Darwin x86_64; bash 3.2.57(1)-release; zsh 5.9; Go 1.26.0 darwin/amd64. Each version command exited 0. The module declares Go 1.25.5; these results use the installed 1.26.0 compiler, not a claim of qualification under 1.25.5. No `pwsh` executable was available. A combined `command -v go bash zsh python3 pwsh` discovery call exited 1 because of the missing last executable; it was not a validation gate. Personal executable paths are omitted here.

All `<scratch>` path spellings below replace one newly created temporary directory. `HOME`, `CURATOR_CONFIG`, PATH, locale and marker operands of every CLI/shell probe were explicit scratch values, not the operator's environment. Generated hooks came from the freshly built CLI. The reviewer matrix is an input requirement, not accepted prior execution evidence; the referenced cocoaskills revision and six review rounds were not independently reproduced. The GitHub browser fetch of that upstream test file failed, so no claim relies on its unseen contents.

## Source evidence index

Line numbers are at the frozen revisions above. Source-file links identify the first relevant line; ranges below describe what was inspected.

| ID | Source | Fact established |
|---|---|---|
| C1 | [`internal/envfiles/envfiles.go:24–63; 72–123`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/envfiles/envfiles.go#L24) | ProjectContent and GlobalContent prepend in both shells; direct writers record digests. Project POSIX bytes locate their source at runtime and contain no literal project root. |
| C2 | [`cmd/curator/main.go:2331–2372`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/cmd/curator/main.go#L2331) | Production shell-init entry calls Hook or InstallHook; --no-global is a generation flag; install output includes a raw cache path and SourceCommand. |
| C3 | [`internal/shell/shell.go:48–70, 90–152, 165–197, 204–248, 267–369, 373–434, 452–568, 588–821`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/shell/shell.go#L48) | Default A-warning; project check-then-source; global source bypass; raw warnings; env-file discovery; saved PATH restoration; root variable not cleared; shell integration and cache writer. |
| C4 | [`cmd/curator/hook.go:44–113`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/cmd/curator/hook.go#L44) | approve reads any selected current bytes and stores their digest; no canonical-template equality check; raw human paths in approval output. |
| C5 | [`internal/install/commit.go:695–718`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/install/commit.go#L695) | Post-commit live env-file digest recording; failures become warnings. This is not hook-cache refresh. |
| C6 | [`internal/shell/shell_hook_trust_test.go:521–594`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/shell/shell_hook_trust_test.go#L521) | Existing hostile-checkout test explicitly expects B refusal and A execution; PowerShell skips when unavailable. |
| C7 | [`internal/hookapproval/hookapproval.go:602–631`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/hookapproval/hookapproval.go#L602) | Go approval state writer stages, chmods 0600, syncs and atomically renames; this alone proves no hook read-side ownership/ACL policy and no execution snapshot binding. |
| C8 | [`internal/shell/shell_test.go:40–122, 141–248, 302–335, 413–466`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/shell/shell_test.go#L40) | Existing nested activation, saved-PATH, invalid-PWD, registration and cache tests. The leave assertion checks active state and PATH, not CSK_PROJECT_ROOT. |
| C9 | [`internal/envfiles/envfiles_test.go:34–115`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/envfiles/envfiles_test.go#L34) | Generated helper test demands a PATH prefix at line 79; tested helper behavior is currently prepend. |
| C10 | [`cmd/curator/security_posture.go:10–20`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/cmd/curator/security_posture.go#L10) | Status currently derives HookTrust from the binary DefaultTrustProfile, not proof of a cached or live hook revision. |
| S1 | [`profiles/manager.md:§3, 665–701`](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/profiles/manager.md#L665) | Distinguishes self-contained child launchers (prepend) from generated interactive env helpers (also currently prepend). Direct command locations remain shell independent. |
| S2 | [`profiles/manager.md:§§8–8.7, 1242–1427`](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/profiles/manager.md#L1242) | Optional activation, nearest env-file search, global precedence, closed digest approvals and commands, warning-only diagnostics, explicit staged A→B prohibition on an immediate flip, status and downstream real-shell binding. |
| S3 | [`conformance/v1/vectors/shell-hook-trust.json:43–105 and cases from 108`](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/conformance/v1/vectors/shell-hook-trust.json#L43) | Closed A/B rollout expectations. Parsed JSON contains 14 cases. Spec validation is structural, not hook execution. |
| S4 | [`protocol/environments.md:§11, 3440–3498`](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/environments.md#L3440) | Historical env-sourcing/PATH injection is motivation for independent provider trust roots; provider discovery has a separate staged rollout. |
| S5 | [`protocol/environments.md:§12, 3575–3592`](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/environments.md#L3575) | env status shares the manager posture inventory and must remain read-only; --check signals non-current state. |
| S6 | [`profiles/manager.md:2237–2244`](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/profiles/manager.md#L2237) | Compiled-repository environment exposure also says prepend; it needs an explicit interactive-versus-launcher distinction. |
| S7 | [`profiles/manager.md:§10, 1502–1514`](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/profiles/manager.md#L1502) | Closed hook-trust posture values are A-warning/B-enforcing; other gates have their own independent revision values. |

Additional inspected boundary: Environments §10.2, `protocol/environments.md:3234–3236`, reserves the child launch fragment's `path_prepend`. This proposal does not rename it or change its semantics. `internal/envfiles/stage.go:13–21` uses the same helper generators as direct writes. Scoped `rg` over `internal/install` and `cmd/curator/main.go` found InstallHook only at the shell-init install call; absence of automatic refresh is bounded to those production surfaces, not a claim about every external updater. [`cmd/curator/main.go:416–459`](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/cmd/curator/main.go#L416) passes `--flag=value` through the standard boolean flag parser: the proposed explicit `--no-global=false` uses existing syntax, while preserving an omitted flag on a cache refresh is new proposed behavior.

## Validation ledger

The following ledger and P1–P4 raw observations were inherited from the initial researcher's already-attached task-scoped evidence. They report standalone processes without tee or pipelines. Recovery rechecked the cited source/spec, reran the two focused Go tests and audited the revised artifacts; it did not rebuild the CLI or repeat the P1–P4 scripts. See the recovery ledger below for that distinction. No full repository gate, native Windows suite, proposed mutant or candidate implementation gate is claimed.

| Command (temporary prefix normalized) | Actual exit | Interpretation |
|---|---:|---|
| `go build -o <scratch>/curator ./cmd/curator` | 0 | Built this source revision for production-entry probes. |
| `go test ./internal/shell -run '^TestShellHookRefusesHostileCheckout$' -count=1 -v` | 0 | Current A/B contract passes on sh, dash, bash, zsh. PowerShell subtest skipped. A deliberately executes payloads. |
| `go test ./internal/envfiles -run '^TestWrite(ProjectShapesAndSourcing\|Global)$' -count=1 -v` | 0 | Current generated project/global helpers pass existing tests; the project test asserts prepend. |
| `python3 <scratch>/probe.py` | 0 | Observation harness ran; exit 0 means collection succeeded, not that the requested no-source properties passed. All child exits are in the JSON below. |
| `python3 <scratch>/controls.py` | 0 | Collected raw stderr bytes in two locales; both hooks emit unsafe raw controls. |
| `git diff --no-index --check /dev/null <CIP-path>` and the corresponding evidence-file command | 1 each | New-file comparisons report differences; no whitespace diagnostic was emitted. These are non-zero comparisons, not green validation gates. The independent audit below checks the new files directly. |
| `python3 <scratch>/check_artifacts.py` after first embedding of its own source | 1 | Failed on its own literal forbidden-prefix examples; corrected as described in the audit recipe. |
| `python3 <scratch>/check_artifacts.py` with equivalent split literals | 0 | Both artifacts have template sections, valid companion links, no trailing whitespace or raw control/personal paths; Git reports only the two new research files, with no product or logbook changes. |
| Pasted vulnerable approval hint, bash and zsh subprocesses | 127 each | Failing command in both shells, expected while demonstrating injection: the injected `touch PWNED` succeeds, then the synthetic trailing `y/...` command is absent. Never a passing safety gate. |

Focused existing-test stdout (durations are observed, not budgets):

```text
=== RUN   TestShellHookRefusesHostileCheckout
=== RUN   TestShellHookRefusesHostileCheckout/posix
=== RUN   TestShellHookRefusesHostileCheckout/posix/sh
=== RUN   TestShellHookRefusesHostileCheckout/posix/dash
=== RUN   TestShellHookRefusesHostileCheckout/posix/bash
=== RUN   TestShellHookRefusesHostileCheckout/posix/zsh
=== RUN   TestShellHookRefusesHostileCheckout/powershell
    shell_hook_trust_test.go:594: PowerShell is unavailable (no pwsh on this runner)
--- PASS: TestShellHookRefusesHostileCheckout (0.99s)
    --- PASS: TestShellHookRefusesHostileCheckout/posix (0.99s)
        --- PASS: TestShellHookRefusesHostileCheckout/posix/sh (0.25s)
        --- PASS: TestShellHookRefusesHostileCheckout/posix/dash (0.18s)
        --- PASS: TestShellHookRefusesHostileCheckout/posix/bash (0.23s)
        --- PASS: TestShellHookRefusesHostileCheckout/posix/zsh (0.26s)
    --- SKIP: TestShellHookRefusesHostileCheckout/powershell (0.00s)
PASS
ok  github.com/relux-works/curator/internal/shell 1.779s

=== RUN   TestWriteProjectShapesAndSourcing
--- PASS: TestWriteProjectShapesAndSourcing (0.23s)
=== RUN   TestWriteGlobal
--- PASS: TestWriteGlobal (0.26s)
PASS
ok  github.com/relux-works/curator/internal/envfiles 1.191s
```

## P1 — repository/global execution through the production generator

For each of bash and zsh, the built CLI emitted a fresh hook into a scratch file. A clean shell loaded that hook, entered a synthetic checkout and ran activation; zsh also exercised chpwd. Five project shapes were measured: direct root, nested entry, symlinked `.agents`, symlinked env file and symlinked project root. All **10/10** project runs printed `K3-SOURCE-MARKER` and created the scratch marker; both **2/2** global runs did too. Those observations contradict the requested no-source invariant. Each generator and activation child exited 0. The marker payload was only a printf and scratch-file creation.

## P2 — notice injection and terminal controls

Both **2/2** semicolon-name notices included an unquoted approval operand. Pasting the emitted command into a fresh shell (with a harmless `curator` function) created `PWNED` in that shell's scratch cwd, then exited 127 on the remainder. No real approval/account command was executed by the pasted text.

The original text-mode collector normalizes carriage return during decoding; therefore it is not evidence of the exact CR byte. `controls.py` reran the shells with binary capture. In **4/4** shell/locale combinations (`bash`, `zsh` × `C`, `en_US.UTF-8`), raw LF, ESC, CR and DEL and UTF-8 encodings of U+0085, U+202E and U+2066 were present in stderr. Each single notice had three LF bytes because the path occurred twice. This samples those characters only, not the whole requested C0/C1/bidi ranges. JSON below contains escaped representations; it does not reproduce terminal-active bytes.

## P3 — prepend and stale root

The probe reconstructs the literal POSIX ProjectContent concatenation from `internal/envfiles/envfiles.go:25–36` using JSON decoding of Go string literals. This is a fixture, not a second implementation of activation or a call to an invented generator API. Its SHA-256 is `ebe6c8ce75ef68bf4ff8eff7c549cccebf1c5c72b0671a3c82afb2a9b395b88c`; the actual Go generator is independently exercised by the existing envfiles test above. The fixture is approved through the real CLI in scratch state.

With PATH `/usr/bin:/bin`, both hooks inserted `<project>/.agents/bin` first. `command -v git` selected the project stub and calling `git` created the tool marker (**2/2** shells). Four stub names were materialized, but only git resolution/invocation was measured: **1/4** requested B1 command names per shell. Leaving restored `/usr/bin:/bin` but left `CSK_PROJECT_ROOT` pointing at the old project (**2/2**). B3 switching/reentry/duplicates and B2 empty PATH were not measured by this probe.

## P4 — current warm baseline

Each measurement starts a new real shell in the approved project, loads the actual generated hook, then invokes `_curator_auto_env` 1,000 times. Five batches ran per shell, **10,000/10,000** loop calls in total, all batch subprocess exits 0. Wall time uses Python `perf_counter`; startup and initial activation are included and amortized. Bash per-call batch averages were 2.2892, 1.8999, 2.5098, 1.8921 and 1.7840 ms (median 1.8999). Zsh averages were 2.0047, 2.1476, 2.2576, 1.8213 and 1.7923 ms (median 2.0047). No new-hook speedup, individual-sample p95 or cross-machine budget pass is inferred.

## Coverage bound

These baseline observations touched A1, A2, A3, A4, B1, B3, C1, C2 and D4: **9/18** reviewer rows, with the partial bounds above. A5, B2, B4, C3, C4, D1, D2, D3 and D5 were not probed here; targeted baseline unit tests are not relabelled as those new-vector results. No TOCTOU exploit, hardened approval store, candidate no-source implementation, candidate mutation test or native Windows behavior was qualified. Candidate conformance is **0/18**, as expected for a design-only task. The CIP defines all 18 rows and the alternative-only E obligations for future implementation.

## Reproduction scripts and raw observations

Create a fresh temporary directory, build the CLI there from the pinned curator checkout, save the following scripts beside it, and run from that checkout. Reproduce only in a disposable scratch home. The script uses subprocess argument arrays, a closed environment, bounded 15-second children and scratch-only markers. The Python harness follows the attached reviewer probe format; permanent downstream tests should use the repository's Go test stack and existing real-shell runners.

`probe.py`:

```python
from pathlib import Path
import hashlib, json, re, subprocess, sys, time, statistics
base=Path(__file__).resolve().parent
binary=base/'curator'
source=Path.cwd()
fixture=source.joinpath('internal/envfiles/envfiles.go').read_text().split('envSh :=',1)[1].split('envPs1 :=',1)[0]
canonical=''.join(json.loads(s) for s in re.findall(r'"(?:\\.|[^"\\])*"',fixture))
rows=[]
def env(root, **extra):
    home=root/'home'; home.mkdir(parents=True,exist_ok=True)
    return {'HOME':str(home),'CURATOR_CONFIG':str(home/'manager/config.json'),'PATH':'/usr/bin:/bin','LC_ALL':'C',**{k:str(v) for k,v in extra.items()}}
def call(argv, e, cwd):
    return subprocess.run(argv,env=e,cwd=cwd,capture_output=True,text=True,timeout=15)
def record(label,r,**values):
    rows.append({'case':label,'exit':r.returncode,**values})
def cli(args,e,root,label):
    r=call([str(binary),*args],e,root); record(label,r)
    if r.returncode: raise RuntimeError(label+': '+repr(r.stderr))
    return r
for name in ('bash','zsh'):
    exe='/bin/'+name
    prefix=[exe,'--noprofile','--norc','-c'] if name=='bash' else [exe,'-f','-c']
    for kind in ('root','nested','agents-link','env-link','root-link','global','notice-semicolon','notice-controls','path-order'):
        root=base/(name+'-'+kind); root.mkdir(exist_ok=True)
        proj=root/'repo'; (proj/'nested').mkdir(parents=True,exist_ok=True)
        agents=proj/'.agents'
        if kind=='agents-link':
            real=root/'agents-target'; real.mkdir(exist_ok=True); agents.symlink_to(real,target_is_directory=True)
        else: agents.mkdir(exist_ok=True)
        (agents/'bin').mkdir(exist_ok=True)
        if kind=='notice-semicolon':
            p=root/'x;touch PWNED;y'; proj.rename(p); proj=p; agents=proj/'.agents'
        if kind=='notice-controls':
            p=root/('x\n\x1b\r\x7f\u0085\u202e\u2066z'); proj.rename(p); proj=p; agents=proj/'.agents'
        e=env(root,HOOK=root/'hook',TARGET=proj/'nested' if kind=='nested' else proj,CANARY=root/'marker',OUTSIDE=root,TOOL_CANARY=root/'tool-marker')
        script="printf '%s\\n' K3-SOURCE-MARKER\n: > \"$CANARY\"\n"
        if kind=='path-order': script=canonical
        file=agents/'env.sh'
        if kind=='env-link':
            target=root/'payload'; target.write_text(script); file.symlink_to(target)
        else: file.write_text(script)
        if kind=='root-link':
            alias=root/'alias'; alias.symlink_to(proj,target_is_directory=True); e['TARGET']=str(alias/'nested')
        if kind=='global':
            global_file=root/'home/manager/global/env.sh'; global_file.parent.mkdir(parents=True,exist_ok=True); global_file.write_text(script)
            file.unlink(); e['TARGET']=str(root)
        args=['shell-init',name]+([] if kind=='global' else ['--no-global'])
        generated=cli(args,e,root,name+'/'+kind+'/generate'); (root/'hook').write_text(generated.stdout)
        body='. "$HOOK"; builtin cd "$TARGET"; _curator_auto_env; printf "END\\n"'
        if kind=='path-order':
            cli(['hook','approve',str(file)],e,root,name+'/path-order/approve')
            for tool in ('git','ssh','curl','python3'):
                stub=agents/'bin'/tool; stub.write_text('#!/bin/sh\n: > "$TOOL_CANARY"\n'); stub.chmod(0o755)
            body='. "$HOOK"; builtin cd "$TARGET"; _curator_auto_env; printf "PATH=%s\\n" "$PATH"; command -v git; git; builtin cd "$OUTSIDE"; _curator_auto_env; printf "LEAVE_ROOT=%s\\n" "${CSK_PROJECT_ROOT-unset}"; printf "LEAVE_PATH=%s\\n" "$PATH"'
        r=call(prefix+[body],e,root)
        values={'marker_stdout':'K3-SOURCE-MARKER' in r.stdout,'marker_file':(root/'marker').exists()}
        if kind.startswith('notice'):
            values['stderr_ascii']=ascii(r.stderr.replace(str(base),'<scratch>'))
        if kind=='notice-semicolon':
            notice=r.stderr.split('run curator hook approve ',1)[1].split(' to approve it',1)[0]
            paste=call(prefix+['curator() { :; }; curator hook approve '+notice],e,root)
            record(name+'/notice-semicolon/pasted-hint',paste,unexpected_file=(root/'PWNED').exists())
        if kind=='path-order':
            values.update(stdout=r.stdout.replace(str(base),'<scratch>'),stub_executed=(root/'tool-marker').exists())
        record(name+'/'+kind,r,**values)
    # Warm baseline: approved generated fixture, repeated actual prompt hook calls.
    e=env(base/(name+'-path-order'),HOOK=base/(name+'-path-order')/'hook',TARGET=base/(name+'-path-order')/'repo')
    samples=[]
    for n in range(5):
        body='builtin cd "$TARGET"; . "$HOOK"; i=0; while [ "$i" -lt 1000 ]; do _curator_auto_env; i=$((i+1)); done'
        started=time.perf_counter(); r=call(prefix+[body],e,base); elapsed=time.perf_counter()-started
        samples.append(elapsed*1000/1000); record(name+'/warm-batch-'+str(n+1),r,calls=1000,ms_per_call=round(samples[-1],4))
    rows.append({'case':name+'/warm-summary','median_batch_ms_per_call':round(statistics.median(samples),4),'min':round(min(samples),4),'max':round(max(samples),4)})
summary={'fixture_sha256':hashlib.sha256(canonical.encode()).hexdigest(),'source_revision':subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),'rows':rows}
(base/'probe-results.json').write_text(json.dumps(summary,ensure_ascii=True,indent=2)+'\n')
print(json.dumps(summary,ensure_ascii=True,indent=2))
```

`probe-results.json` (scratch prefix normalized by the collector):

```json
{
  "fixture_sha256": "ebe6c8ce75ef68bf4ff8eff7c549cccebf1c5c72b0671a3c82afb2a9b395b88c",
  "source_revision": "ca1b776fb580ec0cee0173bf150daf063023aeaa",
  "rows": [
    {
      "case": "bash/root/generate",
      "exit": 0
    },
    {
      "case": "bash/root",
      "exit": 0,
      "marker_stdout": true,
      "marker_file": true
    },
    {
      "case": "bash/nested/generate",
      "exit": 0
    },
    {
      "case": "bash/nested",
      "exit": 0,
      "marker_stdout": true,
      "marker_file": true
    },
    {
      "case": "bash/agents-link/generate",
      "exit": 0
    },
    {
      "case": "bash/agents-link",
      "exit": 0,
      "marker_stdout": true,
      "marker_file": true
    },
    {
      "case": "bash/env-link/generate",
      "exit": 0
    },
    {
      "case": "bash/env-link",
      "exit": 0,
      "marker_stdout": true,
      "marker_file": true
    },
    {
      "case": "bash/root-link/generate",
      "exit": 0
    },
    {
      "case": "bash/root-link",
      "exit": 0,
      "marker_stdout": true,
      "marker_file": true
    },
    {
      "case": "bash/global/generate",
      "exit": 0
    },
    {
      "case": "bash/global",
      "exit": 0,
      "marker_stdout": true,
      "marker_file": true
    },
    {
      "case": "bash/notice-semicolon/generate",
      "exit": 0
    },
    {
      "case": "bash/notice-semicolon/pasted-hint",
      "exit": 127,
      "unexpected_file": true
    },
    {
      "case": "bash/notice-semicolon",
      "exit": 0,
      "marker_stdout": true,
      "marker_file": true,
      "stderr_ascii": "'curator: shell_hook_env_unapproved: <scratch>/bash-notice-semicolon/x;touch PWNED;y/.agents/env.sh is not approved; run curator hook approve <scratch>/bash-notice-semicolon/x;touch PWNED;y/.agents/env.sh to approve it\\n'"
    },
    {
      "case": "bash/notice-controls/generate",
      "exit": 0
    },
    {
      "case": "bash/notice-controls",
      "exit": 0,
      "marker_stdout": true,
      "marker_file": true,
      "stderr_ascii": "'curator: shell_hook_env_unapproved: <scratch>/bash-notice-controls/x\\n\\x1b\\n\\x7f\\x85\\u202e\\u2066z/.agents/env.sh is not approved; run curator hook approve <scratch>/bash-notice-controls/x\\n\\x1b\\n\\x7f\\x85\\u202e\\u2066z/.agents/env.sh to approve it\\n'"
    },
    {
      "case": "bash/path-order/generate",
      "exit": 0
    },
    {
      "case": "bash/path-order/approve",
      "exit": 0
    },
    {
      "case": "bash/path-order",
      "exit": 0,
      "marker_stdout": false,
      "marker_file": false,
      "stdout": "PATH=<scratch>/bash-path-order/repo/.agents/bin:/usr/bin:/bin\n<scratch>/bash-path-order/repo/.agents/bin/git\nLEAVE_ROOT=<scratch>/bash-path-order/repo\nLEAVE_PATH=/usr/bin:/bin\n",
      "stub_executed": true
    },
    {
      "case": "bash/warm-batch-1",
      "exit": 0,
      "calls": 1000,
      "ms_per_call": 2.2892
    },
    {
      "case": "bash/warm-batch-2",
      "exit": 0,
      "calls": 1000,
      "ms_per_call": 1.8999
    },
    {
      "case": "bash/warm-batch-3",
      "exit": 0,
      "calls": 1000,
      "ms_per_call": 2.5098
    },
    {
      "case": "bash/warm-batch-4",
      "exit": 0,
      "calls": 1000,
      "ms_per_call": 1.8921
    },
    {
      "case": "bash/warm-batch-5",
      "exit": 0,
      "calls": 1000,
      "ms_per_call": 1.784
    },
    {
      "case": "bash/warm-summary",
      "median_batch_ms_per_call": 1.8999,
      "min": 1.784,
      "max": 2.5098
    },
    {
      "case": "zsh/root/generate",
      "exit": 0
    },
    {
      "case": "zsh/root",
      "exit": 0,
      "marker_stdout": true,
      "marker_file": true
    },
    {
      "case": "zsh/nested/generate",
      "exit": 0
    },
    {
      "case": "zsh/nested",
      "exit": 0,
      "marker_stdout": true,
      "marker_file": true
    },
    {
      "case": "zsh/agents-link/generate",
      "exit": 0
    },
    {
      "case": "zsh/agents-link",
      "exit": 0,
      "marker_stdout": true,
      "marker_file": true
    },
    {
      "case": "zsh/env-link/generate",
      "exit": 0
    },
    {
      "case": "zsh/env-link",
      "exit": 0,
      "marker_stdout": true,
      "marker_file": true
    },
    {
      "case": "zsh/root-link/generate",
      "exit": 0
    },
    {
      "case": "zsh/root-link",
      "exit": 0,
      "marker_stdout": true,
      "marker_file": true
    },
    {
      "case": "zsh/global/generate",
      "exit": 0
    },
    {
      "case": "zsh/global",
      "exit": 0,
      "marker_stdout": true,
      "marker_file": true
    },
    {
      "case": "zsh/notice-semicolon/generate",
      "exit": 0
    },
    {
      "case": "zsh/notice-semicolon/pasted-hint",
      "exit": 127,
      "unexpected_file": true
    },
    {
      "case": "zsh/notice-semicolon",
      "exit": 0,
      "marker_stdout": true,
      "marker_file": true,
      "stderr_ascii": "'curator: shell_hook_env_unapproved: <scratch>/zsh-notice-semicolon/x;touch PWNED;y/.agents/env.sh is not approved; run curator hook approve <scratch>/zsh-notice-semicolon/x;touch PWNED;y/.agents/env.sh to approve it\\n'"
    },
    {
      "case": "zsh/notice-controls/generate",
      "exit": 0
    },
    {
      "case": "zsh/notice-controls",
      "exit": 0,
      "marker_stdout": true,
      "marker_file": true,
      "stderr_ascii": "'curator: shell_hook_env_unapproved: <scratch>/zsh-notice-controls/x\\n\\x1b\\n\\x7f\\x85\\u202e\\u2066z/.agents/env.sh is not approved; run curator hook approve <scratch>/zsh-notice-controls/x\\n\\x1b\\n\\x7f\\x85\\u202e\\u2066z/.agents/env.sh to approve it\\n'"
    },
    {
      "case": "zsh/path-order/generate",
      "exit": 0
    },
    {
      "case": "zsh/path-order/approve",
      "exit": 0
    },
    {
      "case": "zsh/path-order",
      "exit": 0,
      "marker_stdout": false,
      "marker_file": false,
      "stdout": "PATH=<scratch>/zsh-path-order/repo/.agents/bin:/usr/bin:/bin\n<scratch>/zsh-path-order/repo/.agents/bin/git\nLEAVE_ROOT=<scratch>/zsh-path-order/repo\nLEAVE_PATH=/usr/bin:/bin\n",
      "stub_executed": true
    },
    {
      "case": "zsh/warm-batch-1",
      "exit": 0,
      "calls": 1000,
      "ms_per_call": 2.0047
    },
    {
      "case": "zsh/warm-batch-2",
      "exit": 0,
      "calls": 1000,
      "ms_per_call": 2.1476
    },
    {
      "case": "zsh/warm-batch-3",
      "exit": 0,
      "calls": 1000,
      "ms_per_call": 2.2576
    },
    {
      "case": "zsh/warm-batch-4",
      "exit": 0,
      "calls": 1000,
      "ms_per_call": 1.8213
    },
    {
      "case": "zsh/warm-batch-5",
      "exit": 0,
      "calls": 1000,
      "ms_per_call": 1.7923
    },
    {
      "case": "zsh/warm-summary",
      "median_batch_ms_per_call": 2.0047,
      "min": 1.7923,
      "max": 2.2576
    }
  ]
}
```

`controls.py`:

```python
from pathlib import Path
import json, subprocess
base=Path(__file__).resolve().parent
rows=[]
for name in ('bash','zsh'):
    root=base/(name+'-notice-controls'); project=next(p for p in root.iterdir() if p.name.startswith('x'))
    prefix=['/bin/bash','--noprofile','--norc','-c'] if name=='bash' else ['/bin/zsh','-f','-c']
    for locale in ('C','en_US.UTF-8'):
        e={'PATH':'/usr/bin:/bin','HOME':str(root/'home'),'CURATOR_CONFIG':str(root/'home/manager/config.json'),'HOOK':str(root/'hook'),'TARGET':str(project),'CANARY':str(root/'raw-canary'),'LC_ALL':locale}
        r=subprocess.run(prefix+['. "$HOOK"; builtin cd "$TARGET"; _curator_auto_env'],env=e,cwd=root,capture_output=True,timeout=15)
        rows.append({'shell':name,'locale':locale,'exit':r.returncode,'raw_LF_count':r.stderr.count(b'\n'),'raw_ESC':b'\x1b' in r.stderr,'raw_CR':b'\r' in r.stderr,'raw_DEL':b'\x7f' in r.stderr,'utf8_C1':b'\xc2\x85' in r.stderr,'utf8_bidi':b'\xe2\x80\xae' in r.stderr and b'\xe2\x81\xa6' in r.stderr,'stderr_repr':repr(r.stderr.replace(str(base).encode(),b'<scratch>'))})
(base/'controls-results.json').write_text(json.dumps(rows,indent=2)+'\n')
print(json.dumps(rows,indent=2))
```

`controls-results.json`:

```json
[
  {
    "shell": "bash",
    "locale": "C",
    "exit": 0,
    "raw_LF_count": 3,
    "raw_ESC": true,
    "raw_CR": true,
    "raw_DEL": true,
    "utf8_C1": true,
    "utf8_bidi": true,
    "stderr_repr": "b'curator: shell_hook_env_unapproved: <scratch>/bash-notice-controls/x\\n\\x1b\\r\\x7f\\xc2\\x85\\xe2\\x80\\xae\\xe2\\x81\\xa6z/.agents/env.sh is not approved; run curator hook approve <scratch>/bash-notice-controls/x\\n\\x1b\\r\\x7f\\xc2\\x85\\xe2\\x80\\xae\\xe2\\x81\\xa6z/.agents/env.sh to approve it\\n'"
  },
  {
    "shell": "bash",
    "locale": "en_US.UTF-8",
    "exit": 0,
    "raw_LF_count": 3,
    "raw_ESC": true,
    "raw_CR": true,
    "raw_DEL": true,
    "utf8_C1": true,
    "utf8_bidi": true,
    "stderr_repr": "b'curator: shell_hook_env_unapproved: <scratch>/bash-notice-controls/x\\n\\x1b\\r\\x7f\\xc2\\x85\\xe2\\x80\\xae\\xe2\\x81\\xa6z/.agents/env.sh is not approved; run curator hook approve <scratch>/bash-notice-controls/x\\n\\x1b\\r\\x7f\\xc2\\x85\\xe2\\x80\\xae\\xe2\\x81\\xa6z/.agents/env.sh to approve it\\n'"
  },
  {
    "shell": "zsh",
    "locale": "C",
    "exit": 0,
    "raw_LF_count": 3,
    "raw_ESC": true,
    "raw_CR": true,
    "raw_DEL": true,
    "utf8_C1": true,
    "utf8_bidi": true,
    "stderr_repr": "b'curator: shell_hook_env_unapproved: <scratch>/zsh-notice-controls/x\\n\\x1b\\r\\x7f\\xc2\\x85\\xe2\\x80\\xae\\xe2\\x81\\xa6z/.agents/env.sh is not approved; run curator hook approve <scratch>/zsh-notice-controls/x\\n\\x1b\\r\\x7f\\xc2\\x85\\xe2\\x80\\xae\\xe2\\x81\\xa6z/.agents/env.sh to approve it\\n'"
  },
  {
    "shell": "zsh",
    "locale": "en_US.UTF-8",
    "exit": 0,
    "raw_LF_count": 3,
    "raw_ESC": true,
    "raw_CR": true,
    "raw_DEL": true,
    "utf8_C1": true,
    "utf8_bidi": true,
    "stderr_repr": "b'curator: shell_hook_env_unapproved: <scratch>/zsh-notice-controls/x\\n\\x1b\\r\\x7f\\xc2\\x85\\xe2\\x80\\xae\\xe2\\x81\\xa6z/.agents/env.sh is not approved; run curator hook approve <scratch>/zsh-notice-controls/x\\n\\x1b\\r\\x7f\\xc2\\x85\\xe2\\x80\\xae\\xe2\\x81\\xa6z/.agents/env.sh to approve it\\n'"
  }
]
```

## External primary references

The following documentation was read for platform semantics, not as proof of curator behavior:

- [zsh Functions](https://zsh.sourceforge.io/Doc/Release/Functions.html): alias parsing and precmd/chpwd hook behavior.
- [Microsoft command precedence](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_command_precedence?view=powershell-7.5): alias/function/cmdlet/external-command distinctions.
- [Microsoft quoting rules](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_quoting_rules?view=powershell-7.5): single-quoted strings, embedded apostrophes and smart quotes.
- [Microsoft environment variables](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_environment_variables?view=powershell-7.5): platform PATH separators and PowerShell-version differences for empty environment values.

GNU Bash manual requests timed out and POSIX website requests returned 403; no unobserved page is cited as evidence. Bash behavior claims above come from source and measured real-shell results. Advisory comparisons of A/B versus C are design reasoning, not measurements of an unbuilt candidate.

## Recovery verification and provenance

The initial researcher attached the CIP and this evidence but did not publish a role handoff. Recovery resumed the same research task, retained its source pins and observation corpus, and revised the two artifacts. It introduced no implementation, additional research prerequisite or scheduled leaf. The P1–P4 measurements, build result, earlier audit history and original test stdout above are **inherited evidence**, not claims that recovery reran them.

The old evidence was fetched through `task-board resource get` before replacement. Its section from P1 through the scripts/raw observations matched this document byte-for-byte, SHA-256 `a8a9851e0db671cb1a12ce262af65963e1d752e8db4ed33d01253a86c0aec675`. The comparison subprocess exited 0. This preserves provenance for the original measurements; it does not turn them into a new execution. The cited product sources still match main `ca1b776fb580ec0cee0173bf150daf063023aeaa`; the clean available spec checkout and peeled rc.14 tag remain `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`. The four primary platform documentation links above were reopened during recovery.

Recovery ran the following commands itself. Each validation was a standalone process, without a tee or pipeline. The two Go invocations use the installed Go 1.26.0 and unset shell startup-file overrides; test fixtures select temporary manager state. They validate the **existing** contract, including its unsafe A-warning expectation, rather than the proposed design.

| Recovery command | Actual exit | Result and bound |
|---|---:|---|
| `git ls-remote https://github.com/relux-works/curator.git refs/heads/main` | 0 | Same main source OID as the initial evidence. |
| `git diff --exit-code main -- internal cmd go.mod go.sum .github LOGBOOK.md` | 0 | No changes on those tracked product/check/logbook surfaces. |
| `env -u BASH_ENV -u ENV -u ZDOTDIR GOENV=off GOTOOLCHAIN=local GOWORK=off go test ./internal/shell -run '^TestShellHookRefusesHostileCheckout$' -count=1 -v` | 0 | sh, dash, bash and zsh subtests passed; PowerShell skipped because pwsh was unavailable. Package time 1.747 s. |
| `env -u BASH_ENV -u ENV -u ZDOTDIR GOENV=off GOTOOLCHAIN=local GOWORK=off go test ./internal/envfiles -run '^TestWrite(ProjectShapesAndSourcing\|Global)$' -count=1 -v` | 0 | Both named tests passed, package time 0.857 s; assertions still require prepend. |
| `python3` inline rc.14 vector-count check | 0 | Parsed the pinned JSON and asserted 14 cases and the ordered A-warning/B-enforcing profile set. Structural observation only. |
| `python3` inline attached-evidence comparison | 0 | P1–P4, coverage, scripts and raw observations equal the already-attached evidence section; digest above. |
| `git diff --check` | 0 | No tracked whitespace errors; this alone does not inspect untracked research files. |
| `python3 <recovery-scratch>/check_artifacts.py` | 0 | Executed the documented audit below against both research files; rerun after editing this recovery ledger. Checks template sections, links, whitespace, unsafe characters, personal-path patterns, artifact budget, and exactly two untracked research files. No product/logbook changes. |

The complete proposed real-shell inventory remains unexecuted: candidate coverage is still 0/18. Recovery did not rerun the warm-path measurements, perform a TOCTOU exploit, or qualify Windows. The recommendation continues to be a draft for the operator's decision after review.

## Review handoff boundaries

This artifact and the CIP are the only intended repository changes. No commit, branch switch, merge, spec publication, implementation scheduling or producer spawn belongs to this task. No product material from another operator was copied into the public board. Only the supplied public matrix's case identifiers and requirements are summarized.

The generic checklist's conditional logbook item is inapplicable because the task's binding research rules explicitly require `LOGBOOK.md` untouched. Findings are recorded in these outcomes and task notes instead; no logbook write is asserted. The design remains pending operator acceptance after research review.

## Artifact audit recipe

The standalone audit below checks only the two documents and Git change names; it does not read a credential store. The initial audit exited 0. After its script was embedded in this evidence, the next run exited 1 because its own literal forbidden-prefix examples matched the disclosure check. The literals were split in the embedded script without changing the regular expression's meaning; the corrected audit was then rerun. This was an artifact-validation failure, not a product failure or a passing gate. The ledger's zero result refers to the corrected audit of the revised files.

```python
from pathlib import Path
import hashlib,json,re,subprocess
stem='261004_CIP-0004-shell-hook-no-source-path-append'
paths=[Path('.research')/(stem+'.md'),Path('.research')/(stem+'_evidence.md')]
expected={str(p) for p in paths}
r=subprocess.run(['git','status','--porcelain=v1','-z'],capture_output=True,check=True)
changes=[entry.decode() for entry in r.stdout.split(b'\0') if entry]
assert {s[3:] for s in changes}==expected, 'Unexpected repository change'
assert all(s.startswith('?? ') for s in changes), 'Expected only new research artifacts'
headings=['Summary','Motivation and user stories','Current state','Design','Security considerations','Compatibility and migration','Specification changes','Implementation plan','Test plan','Open questions for the operator']
cip=paths[0].read_text()
assert all('\n## '+h+'\n' in cip for h in headings), 'Missing CIP template section'
for p in paths:
    raw=p.read_bytes(); value=raw.decode('utf-8')
    assert raw.endswith(b'\n'), 'Missing final newline'
    assert not any(line.rstrip(' \t') != line for line in value.split('\n')), 'Trailing whitespace'
    assert not re.search(r'[\x00-\x08\x0b-\x1f\x7f-\x9f\u202a-\u202e\u2066-\u2069]',value), 'Raw unsafe display character'
    assert not re.search(r'/Us' + r'ers/|/ho' + r'me/administrator|curator-cip0004\.[A-Za-z0-9]+',value), 'Personal or unredacted scratch path'
    for target in re.findall(r'\]\(([^)]+)\)',value):
        if target.startswith(('https://','http://','#')): continue
        assert (p.parent/target.split('#',1)[0]).exists(), 'Broken companion link'
assert sum(p.stat().st_size for p in paths)<100000, 'Research artifact budget exceeded'
print(json.dumps({'result':'artifact-audit-pass','files':len(paths),'product_changes':0,'logbook_changes':0,'bytes':sum(p.stat().st_size for p in paths),'artifacts':[{'path':str(p),'sha256':hashlib.sha256(p.read_bytes()).hexdigest()} for p in paths]},indent=2))
```

Republished for review
