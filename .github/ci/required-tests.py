#!/usr/bin/env python3
"""Require exact named PASS events from a go test -json log."""
import json
import sys


def require_tests(path, required):
    with open(path, encoding="utf-8") as stream:
        passed = {row.get("Test") for line in stream if (row := json.loads(line)).get("Action") == "pass"}
    missing = set(required) - passed
    if missing:
        raise ValueError("required tests did not pass: " + ", ".join(sorted(missing)))


if __name__ == "__main__":
    if len(sys.argv) < 3:
        sys.exit("usage: required-tests.py LOG TEST [TEST ...]")
    require_tests(sys.argv[1], sys.argv[2:])
