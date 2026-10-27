#!/usr/bin/env python3
"""Compare release observations; report drift, never admit or edit product policy."""
import argparse
import copy
import hashlib
import json
from pathlib import Path, PurePosixPath
import re

PROVIDERS = ("codex", "claude", "muse")
FORMS = ("interactive", "exec", "resume")
SURFACES = ("permissions", "memory-off", "project-mcp", "project-trust", "hooks")
KEYS = {f"{p}/{f}/{s}" for p in PROVIDERS for f in FORMS for s in SURFACES}


def proof_path(root, name):
    """Admit only canonical evidence paths confined to this observation bundle."""
    if not isinstance(name, str) or "\\" in name:
        raise ValueError("invalid evidence path")
    path = PurePosixPath(name)
    if len(path.parts) < 2 or path.parts[0] != "evidence" or ".." in path.parts or path.as_posix() != name:
        raise ValueError("invalid evidence path")
    target = root / name
    if not target.resolve().is_relative_to((root / "evidence").resolve()):
        raise ValueError("evidence path escapes bundle")
    return target


def validate(value, root):
    if not isinstance(value, dict) or value.get("format") != "provider-release-observations-v1" or value.get("platform") != "darwin/arm64":
        raise ValueError("unknown observation format or platform")
    if value.get("admitted") is not False:
        raise ValueError("observations must explicitly refuse admission")
    if not isinstance(value.get("releases"), dict) or set(value["releases"]) != set(PROVIDERS):
        raise ValueError("missing or unknown provider")
    for row in value["releases"].values():
        if not isinstance(row, dict) or not isinstance(row.get("version"), str) or not row["version"] or not isinstance(row.get("sha256"), str) or not re.fullmatch(r"[0-9a-f]{64}", row["sha256"]):
            raise ValueError("missing version or exact binary digest")
    if not isinstance(value.get("verdicts"), dict) or set(value["verdicts"]) != KEYS:
        raise ValueError("matrix must cover all 45 known cells")
    for row in value["verdicts"].values():
        if not isinstance(row, dict) or row.get("verdict") not in ("pass", "fail", "unknown") or not isinstance(row.get("evidence"), str):
            raise ValueError("invalid verdict or missing evidence")
        if not row["evidence"].strip():
            raise ValueError("blank evidence reference")
    evidence = root / "evidence"
    if not evidence.is_dir():
        raise ValueError("missing evidence directory")
    if not evidence.resolve().is_relative_to(root.resolve()):
        raise ValueError("evidence directory escapes bundle")
    manifest_path = proof_path(root, "evidence/manifest.json")
    if not manifest_path.is_file():
        raise ValueError("missing evidence manifest")
    raw = manifest_path.read_bytes()
    if hashlib.sha256(raw).hexdigest() != value.get("evidence_manifest_sha256"):
        raise ValueError("evidence manifest hash mismatch")
    manifest = json.loads(raw)
    if not isinstance(manifest, dict) or not manifest:
        raise ValueError("invalid evidence manifest")
    for row in value["verdicts"].values():
        # A reference is an exact manifest key, not prose or a report anchor.
        if row["evidence"] not in manifest:
            raise ValueError("unbound evidence reference: " + row["evidence"])
    for name, expected in manifest.items():
        target = proof_path(root, name)
        if name == "evidence/manifest.json" or not isinstance(expected, str) or not re.fullmatch(r"[0-9a-f]{64}", expected):
            raise ValueError("invalid evidence manifest entry")
        if not target.is_file():
            raise ValueError("missing evidence artifact: " + name)
        if hashlib.sha256(target.read_bytes()).hexdigest() != expected:
            raise ValueError("evidence artifact hash mismatch: " + name)


def compare(baseline, candidate, baseline_root, candidate_root):
    validate(baseline, baseline_root)
    validate(candidate, candidate_root)
    changes = []
    for provider in PROVIDERS:
        for field in ("version", "sha256"):
            old, new = baseline["releases"][provider][field], candidate["releases"][provider][field]
            if old != new:
                changes.append({"kind": "release", "provider": provider, "field": field, "before": old, "after": new})
    for key in sorted(KEYS):
        old, new = baseline["verdicts"][key]["verdict"], candidate["verdicts"][key]["verdict"]
        if old != new:
            changes.append({"kind": "verdict", "cell": key, "before": old, "after": new, "evidence": candidate["verdicts"][key]["evidence"]})
    return {"admitted": False, "qualification_required": bool(changes), "changes": changes,
            "disposition": "qualification review required" if changes else "no observed drift; admission still belongs to product policy"}


def controls(baseline, fixture, root):
    # Claim: independent expected edits catch release/byte/verdict drift, while
    # unchanged observations stay quiet. Invalid/missing cells fail closed.
    results = []
    for case in fixture["cases"]:
        control_base = copy.deepcopy(baseline)
        if "baseline_value" in case:
            target = control_base
            for name in case["path"][:-1]:
                target = target[name]
            target[case["path"][-1]] = case["baseline_value"]
        candidate = copy.deepcopy(control_base)
        target = candidate
        for name in case["path"][:-1]:
            target = target[name]
        if case.get("delete"):
            del target[case["path"][-1]]
        else:
            target[case["path"][-1]] = case["value"]
        try:
            result = compare(control_base, candidate, root, root)
            actual = sorted({c["kind"] for c in result["changes"]})
            assert not result["admitted"]
        except ValueError as error:
            if "error" in case and str(error) != case["error"]:
                raise ValueError("control failed for unintended reason: " + case["name"]) from error
            actual = ["invalid"]
        if actual != case["expected"]:
            raise ValueError("control missed intended change: " + case["name"])
        results.append({"case": case["name"], "caught": actual, "admitted": False})
    # Claim: even an all-pass observation cannot authorize a release.
    candidate = copy.deepcopy(baseline)
    for row in candidate["verdicts"].values():
        row["verdict"] = "pass"
    assert compare(baseline, candidate, root, root)["admitted"] is False
    results.append({"case": "all-pass-never-admits", "admitted": False})
    return results


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--baseline", type=Path, default=Path(__file__).with_name("baseline.json"))
    mode = p.add_mutually_exclusive_group(required=True)
    mode.add_argument("--candidate", type=Path, help="a separately reviewed observation matrix")
    mode.add_argument("--binaries", type=Path, help="hash frozen copies only; never execute foreign bytes")
    mode.add_argument("--self-test", action="store_true")
    a = p.parse_args()
    try:
        baseline = json.loads(a.baseline.read_text())
        validate(baseline, a.baseline.resolve().parent)
        if a.self_test:
            result = {"controls": controls(baseline, json.loads(Path(__file__).with_name("canary-fixture.json").read_text()), a.baseline.resolve().parent), "admitted": False}
            code = 0
        elif a.binaries:
            changes = []
            for name in PROVIDERS:
                digest = hashlib.sha256((a.binaries / name).read_bytes()).hexdigest()
                old = baseline["releases"][name]["sha256"]
                if digest != old:
                    changes.append({"kind": "release", "provider": name, "field": "sha256", "before": old, "after": digest})
            result = {"admitted": False, "qualification_required": bool(changes), "changes": changes,
                      "disposition": "freeze candidate; open qualification task; collect paired offline observations; independent review before any policy change" if changes else "same bytes; no admission implied"}
            code = 2 if changes else 0
        else:
            result = compare(baseline, json.loads(a.candidate.read_text()), a.baseline.resolve().parent, a.candidate.resolve().parent)
            code = 2 if result["qualification_required"] else 0
    except (OSError, ValueError, TypeError, KeyError) as e:
        result, code = {"admitted": False, "qualification_required": True, "error": str(e)}, 3
    print(json.dumps(result, indent=2))
    raise SystemExit(code)


if __name__ == "__main__":
    main()
