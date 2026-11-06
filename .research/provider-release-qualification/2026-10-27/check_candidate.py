#!/usr/bin/env python3
"""Paired candidate intake controls, including one real pinned loader replay."""
import argparse
import copy
import json
from pathlib import Path
import shlex
import subprocess
import sys

from candidate import load_candidate
from replay import REPO

HERE = Path(__file__).resolve().parent


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--candidate", type=Path, required=True)
    p.add_argument("--binaries", type=Path, required=True)
    p.add_argument("--scratch", type=Path, required=True)
    a = p.parse_args()
    root = a.scratch.resolve()
    if not root.is_relative_to(REPO / ".tmp") or root.exists():
        p.error("fresh repository .tmp scratch required")
    candidate = load_candidate(a.candidate)
    root.mkdir(parents=True)
    plant = root / "executed"
    fake = root / "binaries"
    fake.mkdir()
    (fake / "codex").write_text("#!/bin/sh\nprintf forbidden > " + shlex.quote(str(plant)) + "\n")
    (fake / "codex").chmod(0o700)
    cases = []

    def run(name, args, expected, witness=None):
        with (root / (name + ".log")).open("w") as output:
            child = subprocess.Popen([sys.executable, "-B", *map(str, args)], stdout=output, stderr=subprocess.STDOUT)
            (root / (name + ".pid")).write_text(str(child.pid) + "\n")
            code = child.wait(timeout=500)
        text = (root / (name + ".log")).read_text()
        # Claim: reject the intended pin error before fixture creation or
        # execution, rather than counting a later runtime/setup failure.
        assert code == expected, (name, code, text)
        if witness:
            assert witness in text, (name, text)
        if Path(args[0]).name == "canary.py":
            observed = json.loads(text)
            assert observed["admitted"] is False and observed["qualification_required"] is True
            if witness:
                assert observed["error"] == witness
        assert not plant.exists(), "foreign candidate executable ran"
        cases.append({"case": name, "exit": code, "expected_reason": witness,
                      "plant_executed": False, "admitted": False})

    for kind in ("mismatch", "missing-provider", "missing-digest", "missing-version", "admission-injection"):
        value = copy.deepcopy(candidate)
        if kind == "missing-provider":
            del value["releases"]["codex"]
            witness = "candidate must pin all three providers"
        elif kind == "missing-digest":
            del value["releases"]["codex"]["sha256"]
            witness = "missing or invalid candidate digest: codex"
        elif kind == "missing-version":
            del value["releases"]["codex"]["version"]
            witness = "missing or invalid candidate version: codex"
        elif kind == "admission-injection":
            value["admitted"] = True
            witness = "invalid candidate format, platform or admission"
        else:
            witness = "candidate digest differs: codex; no execution or admission"
        path = root / (kind + ".json")
        path.write_text(json.dumps(value))
        for script, group in (("replay.py", ["codex-loader"]), ("startup.py", ["--provider", "codex"])):
            name = kind + "-" + script[:-3]
            scratch = root / (name + "-fixture")
            run(name, [HERE / script, *group, "--candidate", path, "--binaries", fake, "--scratch", scratch], 2, witness)
            assert not scratch.exists(), "rejected intake created fixture"
            cases[-1]["fixture_created"] = False
        run(kind + "-canary", [HERE / "canary.py", "--candidate", path, "--binaries", fake], 3, witness)
    # Claim: omitting the candidate option retains the baseline byte gate;
    # foreign bytes cannot execute by falling back to an implicit new pin.
    for script, group in (("replay.py", ["codex-loader"]), ("startup.py", ["--provider", "codex"])):
        name = "absent-candidate-" + script[:-3]
        scratch = root / (name + "-fixture")
        run(name, [HERE / script, *group, "--binaries", fake, "--scratch", scratch], 2,
            "candidate digest differs: codex; no execution or admission")
        assert not scratch.exists(), "absent candidate created fixture"
        cases[-1]["fixture_created"] = False
    # Claim: a matching version/digest pin reaches real native loader probes
    # and records the exact candidate, without altering any admission policy.
    positive = root / "matching-loader"
    run("matching-replay", [HERE / "replay.py", "codex-loader", "--candidate", a.candidate.resolve(),
                            "--binaries", a.binaries.resolve(), "--scratch", positive], 0)
    receipt = json.loads((positive / "receipt.json").read_text())
    assert receipt["candidate"] == candidate and receipt["admitted"] is False
    assert "\nPASS\n" in (positive / "replay.log").read_text()
    assert "--- SKIP:" not in (positive / "replay.log").read_text()
    for test in ("TestConfigLayers", "TestMCPRuntime", "TestResumeApprovedRoot"):
        assert "--- PASS: " + test in (positive / "replay.log").read_text()
    cases[-1]["candidate_recorded"] = True
    run("matching-canary", [HERE / "canary.py", "--candidate", a.candidate.resolve(), "--binaries", a.binaries.resolve()], 2)
    result = {"controls": cases, "matching_replay_exit": receipt["exit"], "candidate": candidate, "admitted": False}
    (root / "results.json").write_text(json.dumps(result, indent=2) + "\n")
    print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()
