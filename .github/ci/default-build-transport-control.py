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


try:
    source.write_bytes(original.replace(anchor, b"\t\t_ = policy // detection plant\n"))
    code, output = witness()
    # A crash, skipped case, setup failure or unrelated failure is not detection.
    if (code != 1 or "--- FAIL: TestDefaultTransportInstallScopes/project/selected" not in output
            or "SSH requires a bound manager wrapper policy" not in output):
        raise SystemExit("default-lane policy plant was not detected by its install witness")
finally:
    source.write_bytes(original)

code, output = witness()
if code != 0 or "--- PASS: TestDefaultTransportInstallScopes/project/selected" not in output:
    raise SystemExit("restored default-lane positive control did not pass")
print("default-lane policy plant detected; original source restored; positive control passed")
