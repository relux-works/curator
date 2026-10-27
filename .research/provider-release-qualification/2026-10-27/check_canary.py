#!/usr/bin/env python3
"""Exercise evidence failures through the real CLI in isolated .tmp bundles."""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import sys

HERE = Path(__file__).resolve().parent
REPO = HERE.parents[2]
CELL = "claude/exec/project-mcp"
PROOF = "evidence/loader-observations.json"


def write_json(path, value):
    path.write_text(json.dumps(value, indent=2) + "\n")


def run(bundle, baseline, candidate, expected_exit, expected_error=None):
    command = [sys.executable, "-B", str(bundle / "canary.py"),
               "--baseline", str(baseline), "--candidate", str(candidate)]
    child = subprocess.run(command, capture_output=True, text=True, timeout=15)
    result = json.loads(child.stdout)
    # Claim: the actual CLI refuses the intended evidence defect, with no
    # admission, then stays quiet once the same bundle's proof is restored.
    assert child.returncode == expected_exit, (child.returncode, result, child.stderr)
    assert result["admitted"] is False, result
    assert result["qualification_required"] is (expected_exit != 0), result
    if expected_error is not None:
        assert result.get("error") == expected_error, result
    else:
        assert result["changes"] == [], result
    return {"exit": child.returncode, **result}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--scratch", required=True, type=Path)
    args = parser.parse_args()
    scratch = args.scratch.resolve()
    if not scratch.is_relative_to(REPO / ".tmp"):
        parser.error("fresh repository .tmp directory required")
    scratch.mkdir(parents=True)  # Refuse reuse of another run's fixture.
    fixture = json.loads((HERE / "canary-fixture.json").read_text())
    results = []
    for case in fixture["bundle_cases"]:
        bundle = scratch / case["name"]
        bundle.mkdir()
        for name in ("canary.py", "candidate.py", "canary-fixture.json", "baseline.json"):
            shutil.copyfile(HERE / name, bundle / name)
        shutil.copytree(HERE / "evidence", bundle / "evidence")
        candidate = bundle / "baseline.json"
        baseline = HERE / "baseline.json"
        if case.get("target") == "baseline":
            baseline, candidate = candidate, baseline
        positive = run(bundle, baseline, candidate, 0)
        changed = json.loads((bundle / "baseline.json").read_text())
        action = case["action"]
        if action in ("blank", "dangling"):
            changed["verdicts"][CELL]["evidence"] = " \n\t " if action == "blank" else "evidence/nonexistent-proof.json"
        elif action == "missing-directory":
            shutil.rmtree(bundle / "evidence")
        elif action == "missing-artifact":
            (bundle / PROOF).unlink()
        elif action == "altered-artifact":
            (bundle / PROOF).write_text("synthetic corruption\n")
        elif action == "missing-manifest":
            (bundle / "evidence/manifest.json").unlink()
        elif action == "altered-manifest":
            with (bundle / "evidence/manifest.json").open("a") as output:
                output.write("\n")
        elif action == "escaping-reference":
            manifest = json.loads((bundle / "evidence/manifest.json").read_text())
            manifest["evidence/../outside.json"] = "0" * 64
            write_json(bundle / "evidence/manifest.json", manifest)
            changed["evidence_manifest_sha256"] = hashlib.sha256((bundle / "evidence/manifest.json").read_bytes()).hexdigest()
        else:
            raise ValueError("unrecognized control action")
        write_json(bundle / "baseline.json", changed)
        negative = run(bundle, baseline, candidate, 3, case["error"])
        shutil.copytree(HERE / "evidence", bundle / "evidence", dirs_exist_ok=True)
        if action == "dangling":
            # The same formerly dangling reference now has actual bytes and
            # an entry in the candidate's newly pinned manifest.
            proof = bundle / "evidence/nonexistent-proof.json"
            proof.write_text("synthetic paired-control proof\n")
            manifest = json.loads((bundle / "evidence/manifest.json").read_text())
            manifest["evidence/nonexistent-proof.json"] = hashlib.sha256(proof.read_bytes()).hexdigest()
            write_json(bundle / "evidence/manifest.json", manifest)
            changed["evidence_manifest_sha256"] = hashlib.sha256((bundle / "evidence/manifest.json").read_bytes()).hexdigest()
            write_json(bundle / "baseline.json", changed)
        else:
            shutil.copyfile(HERE / "baseline.json", bundle / "baseline.json")
        repaired = run(bundle, baseline, candidate, 0)
        results.append({"case": case["name"], "intact": positive,
                        "negative": negative, "repaired": repaired})
    output = {"claim": "CLI evidence defects refuse for the intended reason; intact and repaired proofs pass without admission",
              "cases": results, "admitted": False}
    write_json(scratch / "results.json", output)
    print(json.dumps(output, indent=2))


if __name__ == "__main__":
    main()
