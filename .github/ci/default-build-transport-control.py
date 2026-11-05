#!/usr/bin/env python3
"""Prove the default-lane install witness detects a dropped SSH policy."""
import os
from pathlib import Path
import signal
import subprocess

source = Path("internal/buildrepo/admission.go")
original = source.read_bytes()
anchor = b"\t\trequest.sshPolicy = &policy\n"
if original.count(anchor) != 1:
    raise SystemExit("default-lane policy mutation anchor is not unique")
command = ["go", "test", "./internal/install", "-run",
           "^TestDefaultTransportInstallScopes/project/selected$", "-v",
           "-count=1", "-timeout=4m"]


def witness():
    child = subprocess.Popen(command, stdout=subprocess.PIPE,
                             stderr=subprocess.STDOUT, start_new_session=True)
    try:
        output, _ = child.communicate(timeout=280)
    except subprocess.TimeoutExpired:
        os.killpg(child.pid, signal.SIGKILL)
        child.communicate()
        raise SystemExit("default-lane detection control timed out")
    text = output.decode("utf-8", errors="replace")
    print(text, end="", flush=True)
    return child.returncode, text


plants = [
    ("binding", anchor, b"\t\t_ = policy // detection plant\n", "SSH requires a bound manager wrapper policy"),
    ("ambient-agent", b"\treturn env\n}\n\nfunc runGit(",
     b'\tenv = append(env, "SSH_AUTH_SOCK="+os.Getenv("SSH_AUTH_SOCK")) // detection plant\n\treturn env\n}\n\nfunc runGit(', "ambient agent reached fixture"),
    ("ambient-config", b'"HOME="+p.home', b'"HOME="+os.Getenv("HOME")', "ambient config reached fixture"),
]
try:
    for name, old, new, reason in plants:
        if original.count(old) != 1:
            raise SystemExit(name + " mutation anchor is not unique")
        source.write_bytes(original.replace(old, new))
        code, output = witness()
        if (code != 1 or "--- FAIL: TestDefaultTransportInstallScopes/project/selected" not in output
                or reason not in output):
            raise SystemExit(name + " plant was not detected for the intended reason")
        print(name + " plant detected for: " + reason, flush=True)
finally:
    source.write_bytes(original)

code, output = witness()
if code != 0 or "--- PASS: TestDefaultTransportInstallScopes/project/selected" not in output:
    raise SystemExit("restored default-lane positive control did not pass")
print("all default-lane plants detected; original source restored; positive control passed")
