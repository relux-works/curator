#!/usr/bin/env python3
"""Prove the real Linux process control detects omission of the owner's UID."""
from pathlib import Path
import subprocess

source = Path("internal/buildcache/process_linux.go")
original = source.read_bytes()
anchor = b"if int64(value) == int64(uid) {"
assert original.count(anchor) == 1, "same-user UID anchor must match exactly once"
try:
    source.write_bytes(original.replace(anchor, b"if false {", 1))
    result = subprocess.run(
        ["go", "test", "-count=1", "-timeout=590s", "./internal/buildcache",
         "-run", "^TestLinuxUnprivilegedSweep/running-build$"],
        stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True, timeout=600,
        check=False,
    )
    assert result.returncode != 0, "same-user omission plant survived"
    assert "running/unknown image removed" in result.stdout, result.stdout[-5000:]
    print("same-user omission plant detected: running/unknown image removed")
finally:
    source.write_bytes(original)
assert source.read_bytes() == original, "control plant was not removed"
