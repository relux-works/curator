"""Explicit research pins authorize observation only, never release admission."""
import hashlib
import json
from pathlib import Path
import re

PROVIDERS = ("codex", "claude", "muse")


def load_candidate(path):
    value = json.loads(Path(path).read_text())
    if not isinstance(value, dict) or value.get("format") != "provider-release-candidate-v1" or value.get("platform") != "darwin/arm64" or value.get("admitted") is not False:
        raise ValueError("invalid candidate format, platform or admission")
    releases = value.get("releases")
    if not isinstance(releases, dict) or set(releases) != set(PROVIDERS):
        raise ValueError("candidate must pin all three providers")
    for name, row in releases.items():
        if not isinstance(row, dict) or not isinstance(row.get("version"), str) or not re.fullmatch(r"[0-9][0-9A-Za-z.+_-]*", row["version"]):
            raise ValueError("missing or invalid candidate version: " + name)
        if not isinstance(row.get("sha256"), str) or not re.fullmatch(r"[0-9a-f]{64}", row["sha256"]):
            raise ValueError("missing or invalid candidate digest: " + name)
    return {"format": value["format"], "platform": value["platform"], "admitted": False,
            "releases": {name: {k: releases[name][k] for k in ("version", "sha256")} for name in PROVIDERS}}


def check_binaries(directory, releases, names):
    binaries = {}
    for name in names:
        binary = (Path(directory) / name).resolve(strict=True)
        if hashlib.sha256(binary.read_bytes()).hexdigest() != releases[name]["sha256"]:
            raise ValueError("candidate digest differs: " + name + "; no execution or admission")
        binaries[name] = binary
    return binaries
