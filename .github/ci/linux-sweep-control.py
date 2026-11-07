#!/usr/bin/env python3
"""Require compile-valid plants to fail at their own intended assertions."""
from pathlib import Path
import subprocess

# Readable proc evidence and inode write denial are independent guards. A UID
# omission no longer defeats the busy-file probe, so kill each guard separately.
plants = [
    ("same-user UID", Path("internal/buildcache/process_linux.go"),
     b"if int64(value) == int64(uid) {", b"if int64(value) == -1 {",
     "^TestLinuxProcessUIDControls$/^same$", "same-user UID misclassified"),
    ("busy executable", Path("internal/buildcache/probe_linux.go"),
     b"if errors.Is(probeErr, unix.ETXTBSY) {\n\t\t\t\treturn true, nil",
     b"if errors.Is(probeErr, unix.ETXTBSY) {\n\t\t\t\treturn false, nil",
     "^TestLinuxUnprivilegedSweep$/^unreadable-same-uid$", "running image removed"),
]
for name, source, anchor, replacement, test, message in plants:
    original = source.read_bytes()
    assert original.count(anchor) == 1, f"{name} anchor must match exactly once"
    try:
        source.write_bytes(original.replace(anchor, replacement, 1))
        result = subprocess.run(
            ["go", "test", "-count=1", "-timeout=590s", "./internal/buildcache",
             "-run", test],
            stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True, timeout=600,
            check=False,
        )
        assert result.returncode != 0, f"{name} plant survived"
        assert message in result.stdout, result.stdout[-5000:]
        print(result.stdout, end="", flush=True)
        print(f"{name} plant detected: {message}", flush=True)
    finally:
        source.write_bytes(original)
    assert source.read_bytes() == original, f"{name} plant was not removed"
