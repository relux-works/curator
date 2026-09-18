# TASK-260910-3u9t1e — review verdict revision 1

Verdict: **accepted**. No corrections required. Acceptance routes to integrating; producer integration remains outstanding.

Reviewed candidate tree `875c354cbfa677bc91181d7e51bbbe73ac8d45ae`, base `bd27d32a340855645965a6fc65285d3d3de03d77`, CR-TASK-260910-3u9t1e-1 revision 1. Published patch and `git diff HEAD` both SHA256 `f07fd724c9bd40e636ef4cdb63101a501bc8e3da11d8881b6f360c0c941d4036`; `git diff --exit-code <candidate> --` exit 0. Story history contains R5 d7f424c and R8 bd27d32. No candidate edits made.

Read campaign rules, producer brief/results, patch and validation log, audit R7, and curator-spec profile §6 at verified detached HEAD `47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe`. The scope is a local CLI warning, with no protocol schema or readiness-gate change.

| Review item | Result and evidence |
| --- | --- |
| Implicit-key warning | PASS: cli.py:293-304 emits prominent stderr warning containing home, --public-key and out-of-band guidance; cli.py:323-324 and :332-333 add warning to failure and success JSON respectively. |
| Explicit-key compatibility | PASS: cli.py:302,309 preserves key selection and omits warning; verification and 0/2 exits unchanged (:310-320,326,335). tests/test_registry.py:876 asserts exact prior envelopes on both verdicts. |
| Tests/negative behavior | PASS: tests/test_registry.py:799 covers valid backup and backup below newer checkpoint with stderr + JSON warnings; :876 covers explicit equivalents. Both drive main → parser dispatch → _cmd_verify_backup. |
| Compromised home | PASS: installed CLI probe replaces active private key and keyring after checkpoint. Implicit verification warns and refuses the old checkpoint; explicit original out-of-band pin stays quiet and accepts. 4/4 probe verification cases pass. |
| Docs | PASS: README.md:89,101 explains pinned key provenance, independent storage, refresh on rotation, fallback; SECURITY.md:30 documents expectation; CHANGELOG.md:36 names R7, warning field, unchanged exits and pinned spec. |
| Architecture/scope | PASS: localized CLI output change; no key-management, manager, spec, Docker/deployment, tags or release changes. |
| Independent gates | PASS: full pytest with pinned conformance root and strict mypy in disposable candidate copy; transcripts below. |
| Mutation evidence | PASS: 2/2 targeted narrowing mutants killed by shipped implicit-key test; details below. This is targeted warning-branch coverage, not an exhaustive mutation score. |
| Hygiene | PASS: status lists exactly 5 expected source/test/doc modifications; exact candidate tracked tree has no results/logbook/coverage/build artifacts; diff --check exit 0. Existing ignored caches and dist in the original worktree predate review and are not in the published candidate. Reviewer created no files there. |

## Validation provenance

Independent environment: macOS, CPython 3.14.6, `/tmp/R7-review/venv/bin/python`; editable install points to `/private/tmp/R7-review/candidate/src/csk_registry`. Candidate exported from exact tree using git archive. Tests run in `/tmp/R7-review/candidate`, shell `/bin/bash`, `set -o pipefail`. Dependency installation exit 0. No separate lint gate configured; mypy strict and diff --check pass.

Reused evidence: attached `TASK-260910-3u9t1e_change-request_rev1-validation.log`, hosted run 35346859531, all 9 jobs success (six OS/Python test jobs, strict mypy, distribution build, Docker build), wrapper exit 0. Hosted gate/build not independently rerun, per campaign rules. Local full pytest and mypy independently rerun as explicitly required by reviewer brief.

No newly discovered anomalies, regressions or spec gaps requiring a logbook entry. Known live-home trust limitation is the settled R7 design, now visible and documented. No goal bound to this run (`task-board spawn goal "$TASK_BOARD_RUN_ID"`: Active Goal: none).

## Full suite

Command: `CURATOR_CONFORMANCE_ROOT=/tmp/spec-47c3c8c/conformance/v1 /tmp/R7-review/venv/bin/python -m pytest -q`

```text
........................................................................ [ 40%]
........................................................................ [ 80%]
...................................                                      [100%]
=============================== warnings summary ===============================
../../../../tmp/R7-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/R7-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../tmp/R7-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/R7-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
179 passed, 2 warnings in 77.97s (0:01:17)

```
Exit: 0 (mutant subprocess exits 1 expected and asserted).

## Strict typing

Command: `/tmp/R7-review/venv/bin/python -m mypy`

```text
Success: no issues found in 15 source files

```
Exit: 0 (mutant subprocess exits 1 expected and asserted).

## Installed CLI probe

Command: `/tmp/R7-review/venv/bin/python /tmp/R7-review/probe.py`

```text
$ curator-skill-registry --home /tmp/R7-review/probe-home backup --out /tmp/R7-review/backup.db --checkpoint-out /tmp/R7-review/checkpoint.json 
exit: 0 
stdout: {
  "database": "/tmp/R7-review/backup.db",
  "checkpoint": "/tmp/R7-review/checkpoint.json",
  "log_size": 0,
  "head": "0000000000000000000000000000000000000000000000000000000000000000"
}
 stderr: 
$ curator-skill-registry --home /tmp/R7-review/probe-home verify-backup /tmp/R7-review/backup.db --checkpoint /tmp/R7-review/checkpoint.json 
exit: 0 
stdout: {
  "backup_valid": true,
  "log_size": 0,
  "head": "0000000000000000000000000000000000000000000000000000000000000000",
  "warning": "keys resolved from the live home /tmp/R7-review/probe-home; supply --public-key from an out-of-band copy for an independent check"
}
 stderr: WARNING: keys resolved from the live home /tmp/R7-review/probe-home; supply --public-key from an out-of-band copy for an independent check

$ curator-skill-registry --home /tmp/R7-review/probe-home verify-backup /tmp/R7-review/backup.db --checkpoint /tmp/R7-review/checkpoint.json --public-key ed25519:zuyW6kfcyrM2Q7Tk2D0Hj+6Vlxn5vbaTjiK7i4/kjIo= 
exit: 0 
stdout: {
  "backup_valid": true,
  "log_size": 0,
  "head": "0000000000000000000000000000000000000000000000000000000000000000"
}
 stderr: 
$ curator-skill-registry --home /tmp/R7-review/probe-home verify-backup /tmp/R7-review/backup.db --checkpoint /tmp/R7-review/checkpoint.json 
exit: 2 
stdout: {
  "backup_valid": false,
  "error": "checkpoint signature does not verify",
  "warning": "keys resolved from the live home /tmp/R7-review/probe-home; supply --public-key from an out-of-band copy for an independent check"
}
 stderr: WARNING: keys resolved from the live home /tmp/R7-review/probe-home; supply --public-key from an out-of-band copy for an independent check

$ curator-skill-registry --home /tmp/R7-review/probe-home verify-backup /tmp/R7-review/backup.db --checkpoint /tmp/R7-review/checkpoint.json --public-key ed25519:zuyW6kfcyrM2Q7Tk2D0Hj+6Vlxn5vbaTjiK7i4/kjIo= 
exit: 0 
stdout: {
  "backup_valid": true,
  "log_size": 0,
  "head": "0000000000000000000000000000000000000000000000000000000000000000"
}
 stderr: 
PASS: 4/4 installed CLI verification cases; replaced live keys refuse original checkpoint while independent pin still accepts.

```
Exit: 0 (mutant subprocess exits 1 expected and asserted).

## Narrowing mutants

Command: `/tmp/R7-review/venv/bin/python /tmp/R7-review/mutants.py`

```text
failure_warning_only_for_oserror exit 1 
 F                                                                        [100%]
=================================== FAILURES ===================================
_____ test_verify_backup_without_public_key_warns_on_stderr_and_in_output ______

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-175/test_verify_backup_without_pub0')
capsys = <_pytest.capture.CaptureFixture object at 0x11395d940>

    def test_verify_backup_without_public_key_warns_on_stderr_and_in_output(
        tmp_path: Path, capsys: pytest.CaptureFixture[str]
    ) -> None:
        home = tmp_path / "home"
        home.mkdir()
        protect_private_directory(home)
        key = signing.generate_key()
        key_path = home / "signing-key.pem"
        key_path.write_bytes(signing.export_key_pem(key))
        protect_private_file(key_path)
        store = Store(home / "registry.db")
        store.append(key.sign_record(_body()), created_at="2026-07-07T00:00:00Z")
        store.close()
    
        backup = tmp_path / "backup.db"
        checkpoint = tmp_path / "checkpoint.json"
        assert main(
            [
                "--home",
                str(home),
                "backup",
                "--out",
                str(backup),
                "--checkpoint-out",
                str(checkpoint),
            ]
        ) == 0
        capsys.readouterr()
    
        assert main(
            [
                "--home",
                str(home),
                "verify-backup",
                str(backup),
                "--checkpoint",
                str(checkpoint),
            ]
        ) == 0
        captured = capsys.readouterr()
        payload = json.loads(captured.out)
        assert payload["backup_valid"] is True
        assert str(home) in payload["warning"]
        assert "--public-key" in payload["warning"]
        assert "out-of-band" in payload["warning"]
        assert "WARNING" in captured.err
        assert str(home) in captured.err
        assert "--public-key" in captured.err
    
        live = Store(home / "registry.db")
        live.append(key.sign_record(_body("revoked")), created_at="2026-07-07T01:00:00Z")
        newer_checkpoint = tmp_path / "newer-checkpoint.json"
        newer_checkpoint.write_text(
            json.dumps(build_snapshot(live, key)),
            encoding="utf-8",
        )
        live.close()
        assert main(
            [
                "--home",
                str(home),
                "verify-backup",
                str(backup),
                "--checkpoint",
                str(newer_checkpoint),
            ]
        ) == 2
        captured = capsys.readouterr()
        payload = json.loads(captured.out)
        assert payload["backup_valid"] is False
        assert "error" in payload
>       assert str(home) in payload["warning"]
                            ^^^^^^^^^^^^^^^^^^
E       KeyError: 'warning'

tests/test_registry.py:870: KeyError
=============================== warnings summary ===============================
../../../../tmp/R7-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/R7-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../tmp/R7-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/R7-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_registry.py::test_verify_backup_without_public_key_warns_on_stderr_and_in_output
1 failed, 91 deselected, 2 warnings in 7.72s
 
success_warning_only_for_empty_backups exit 1 
 F                                                                        [100%]
=================================== FAILURES ===================================
_____ test_verify_backup_without_public_key_warns_on_stderr_and_in_output ______

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-176/test_verify_backup_without_pub0')
capsys = <_pytest.capture.CaptureFixture object at 0x108671940>

    def test_verify_backup_without_public_key_warns_on_stderr_and_in_output(
        tmp_path: Path, capsys: pytest.CaptureFixture[str]
    ) -> None:
        home = tmp_path / "home"
        home.mkdir()
        protect_private_directory(home)
        key = signing.generate_key()
        key_path = home / "signing-key.pem"
        key_path.write_bytes(signing.export_key_pem(key))
        protect_private_file(key_path)
        store = Store(home / "registry.db")
        store.append(key.sign_record(_body()), created_at="2026-07-07T00:00:00Z")
        store.close()
    
        backup = tmp_path / "backup.db"
        checkpoint = tmp_path / "checkpoint.json"
        assert main(
            [
                "--home",
                str(home),
                "backup",
                "--out",
                str(backup),
                "--checkpoint-out",
                str(checkpoint),
            ]
        ) == 0
        capsys.readouterr()
    
        assert main(
            [
                "--home",
                str(home),
                "verify-backup",
                str(backup),
                "--checkpoint",
                str(checkpoint),
            ]
        ) == 0
        captured = capsys.readouterr()
        payload = json.loads(captured.out)
        assert payload["backup_valid"] is True
>       assert str(home) in payload["warning"]
                            ^^^^^^^^^^^^^^^^^^
E       KeyError: 'warning'

tests/test_registry.py:841: KeyError
=============================== warnings summary ===============================
../../../../tmp/R7-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/R7-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../tmp/R7-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/R7-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_registry.py::test_verify_backup_without_public_key_warns_on_stderr_and_in_output
1 failed, 91 deselected, 2 warnings in 8.63s
 
PASS: 2/2 narrowing mutants killed; original restored byte-for-byte

```
Exit: 0 (mutant subprocess exits 1 expected and asserted).

## Final candidate hygiene

Command: `git status --short --untracked-files=all`

```text
 M CHANGELOG.md
 M README.md
 M SECURITY.md
 M src/csk_registry/cli.py
 M tests/test_registry.py

```
Exit: 0 (mutant subprocess exits 1 expected and asserted).

## Reproduction scripts

### probe.py

```python
import json, subprocess
from pathlib import Path
from csk_registry.keys import initialize_key
from csk_registry.store import Store
from csk_registry.permissions import protect_private_directory
home=Path('/tmp/R7-review/probe-home'); home.mkdir(exist_ok=True); protect_private_directory(home)
key=initialize_key(home)
store=Store(home/'registry.db'); store.close()
cli='/tmp/R7-review/venv/bin/curator-skill-registry'
def run(*args):
    p=subprocess.run([cli,'--home',str(home),*args],capture_output=True,text=True)
    print('$ curator-skill-registry --home',home,*args, '\nexit:',p.returncode,'\nstdout:',p.stdout,'stderr:',p.stderr)
    return p,json.loads(p.stdout)
backup='/tmp/R7-review/backup.db'; checkpoint='/tmp/R7-review/checkpoint.json'
p,_=run('backup','--out',backup,'--checkpoint-out',checkpoint); assert p.returncode==0
args=['verify-backup',backup,'--checkpoint',checkpoint]
for explicit in [False,True]:
    p,v=run(*args,*(['--public-key',key.public_pinned] if explicit else []))
    assert p.returncode==0 and v['backup_valid']
    assert ('warning' in v)==(not explicit)
    assert bool(p.stderr)==(not explicit)
initialize_key(home,replace=True) # simulate compromised home replacing both active key and keyring
for explicit in [False,True]:
    p,v=run(*args,*(['--public-key',key.public_pinned] if explicit else []))
    assert p.returncode==(0 if explicit else 2)
    assert v['backup_valid']==explicit
    assert ('warning' in v)==(not explicit)
    assert bool(p.stderr)==(not explicit)
print('PASS: 4/4 installed CLI verification cases; replaced live keys refuse original checkpoint while independent pin still accepts.')

```

### mutants.py

```python
from pathlib import Path
import subprocess
p=Path('/tmp/R7-review/candidate/src/csk_registry/cli.py')
original=p.read_text()
mutants={
 'failure_warning_only_for_oserror': ('if warning is not None:\n            payload["warning"]', 'if warning is not None and isinstance(exc, OSError):\n            payload["warning"]'),
 'success_warning_only_for_empty_backups': ('if warning is not None:\n        success["warning"]', 'if warning is not None and boundary.log_size == 0:\n        success["warning"]'),
}
try:
 for name,(old,new) in mutants.items():
    assert original.count(old)==1
    p.write_text(original.replace(old,new))
    result=subprocess.run(['/tmp/R7-review/venv/bin/python','-B','-m','pytest','-q','tests/test_registry.py','-k','test_verify_backup_without_public_key_warns_on_stderr_and_in_output'],cwd='/tmp/R7-review/candidate',capture_output=True,text=True)
    print(name, 'exit',result.returncode,'\n',result.stdout,result.stderr)
    assert result.returncode==1 and "KeyError: 'warning'" in result.stdout
finally:
 p.write_text(original)
assert p.read_text()==original
print('PASS: 2/2 narrowing mutants killed; original restored byte-for-byte')

```
