# Shell hook security test matrix (from cocoaskills)

Source: cocoaskills main 0da153a (public: github.com/ivanopcode/cocoaskills), plus six review rounds on an
approval-gate candidate that is not landed. Every case runs REAL shells (bash and zsh as subprocesses with
`--noprofile --norc` / `-f`) on the generated hook, not string checks.

## A. Repository code never runs (the K3 invariant)
| # | Case | Expect |
|---|---|---|
| A1 | hostile `.agents/env.sh` (prints a marker, touches a file) in the project root; `cd` into it (zsh chpwd) / run the prompt function (bash) | marker absent, file absent, PATH/root still activated |
| A2 | same, entered via a nested subdirectory | same |
| A3 | `.agents` is a symlink, `env.sh` is a symlink, the project root is a symlink | same |
| A4 | hostile global env file (the manager home) | not sourced unless it is manager-owned and trusted |
| A5 | reintroduce sourcing in the generator (mutant) | A1 must fail; proves the test is not vacuous |
Reference: the probe in TASK note below; cocoaskills `tests/test_shell_init.py` (activation/restore, reentry).

## B. PATH order (no shadowing)
| # | Case | Expect |
|---|---|---|
| B1 | project bin contains `git`, `ssh`, `curl`, `python3` stubs that write a marker | after activation `command -v git` resolves to the system tool; marker absent |
| B2 | empty inherited PATH | project bin alone, no empty entry (no implicit `.`) |
| B3 | activate, leave, re-enter, switch projects | PATH entries owned by the hook are removed/restored exactly; no duplicates; project root variable unset on leave |
| B4 | `--no-global` / project-scan opt-out | respected, and preserved when a cached hook is refreshed by install |
Reference: `tests/test_env_files.py` (`test_env_ps1_appends_project_bin`), `tests/test_command_names.py` (reserved names),
`tests/test_tool_paths.py`, `tests/test_tool_paths_symlinks.py` (manager resolves git/ssh skipping shim dirs).

## C. Hostile paths and output (if the hook prints anything)
| # | Case | Expect |
|---|---|---|
| C1 | project dir named `x;touch PWNED;y` | any printed remedy command is single-quoted; pasting it runs nothing else |
| C2 | path with newline, ESC, CR, C1 (U+0080-U+009F), bidi controls (U+202A-U+202E, U+2066-U+2069), DEL | printed escaped in C and UTF-8 locales; no forged extra notice line |
| C3 | spaces, glob characters, very long path | handled literally |
| C4 | `PWD` empty, `.`, or relative | no hang, no activation |

## D. Hook robustness and performance
| # | Case | Expect |
|---|---|---|
| D1 | user aliases/functions named `ls`, `tr`, `cd`, `readlink` | hook decision unchanged (`command`/`builtin`, `command -p` for tools) |
| D2 | sourced twice / bash PROMPT_COMMAND already contains the hook | no duplicate registration |
| D3 | zsh: `cd` inside an activation step | no recursive re-entry |
| D4 | warm-path latency per prompt (bash PROMPT_COMMAND) | budget test (we target < 10 ms warm); measure before/after |
| D5 | PowerShell parity: same A/B/C cases on Windows | native pwsh lane in CI |

## E. If an approval/trust gate is kept instead (lessons from our six rounds)
- Trust = exact canonical bytes (root as the only variable) + realpath; store 0600 in a 0700 dir; refuse a store that is
  a symlink, group/other-writable, has ACL `+` (macOS `ls -le`), NUL/control bytes, or a foreign owner.
- A cached decision must never outlive a change: key on inode, size, mtime AND ctime at the finest resolution, and only
  cache when mtime/ctime are strictly older than the check (git's racy-clean rule); same-size same-second edit is a test.
- Install must never fail on trust-store health (lock contention, untrusted store): warn instead.
- `approve`-style output must escape everything (CR/ESC can hide the payload line from the human approving it).

## Probe (copy and adapt)
```python
from pathlib import Path; import shutil, subprocess
from csk import shell_init  # replace with the manager's hook generator
base=Path('probe'); base.mkdir(exist_ok=True)
for name in ('bash','zsh'):
    exe=shutil.which(name); root=base/name; proj=root/'repo'; nested=proj/'nested'; nested.mkdir(parents=True,exist_ok=True)
    (proj/'.agents').mkdir(exist_ok=True)
    (proj/'.agents'/'env.sh').write_text("printf '%s\\n' 'K3-SOURCE-MARKER'\n")
    hook=root/'hook.sh'; hook.write_text(shell_init.shell_init(name, include_global=False))
    body='. "$HOOK"; cd "$NESTED"; ' + ('_csk_auto_env; ' if name=='bash' else '') + "printf '%s\\n' END"
    argv=[exe,'--noprofile','--norc','-c',body] if name=='bash' else [exe,'-f','-c',body]
    r=subprocess.run(argv, env={'PATH':'/usr/bin:/bin','HOME':str(root/'home'),'HOOK':str(hook),'NESTED':str(nested)}, capture_output=True, text=True)
    assert 'K3-SOURCE-MARKER' not in r.stdout, f'{name}: repository code executed'
```
