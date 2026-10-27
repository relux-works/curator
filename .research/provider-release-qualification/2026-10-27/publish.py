#!/usr/bin/env python3
"""Project task-local evidence without raw prompts, native stores or locators."""
import argparse
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

from replay import HERE, REPO


def digest(data):
    return hashlib.sha256(data).hexdigest()


def write(path, value):
    text = value if isinstance(value, str) else json.dumps(value, indent=2, sort_keys=True) + "\n"
    text = re.sub(r"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}", "$SYNTHETIC_ID", text)
    if re.search(r"/Users/|/home/[^/$ ]|[-]Users[-]|[-]home[-]|Bearer |synthetic-not-a-real-key", text):
        raise ValueError("private locator, prompt fragment or credential material in export")
    path.write_text(text)


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--scratch", type=Path, required=True, help="parent containing the completed *-01 groups")
    a = p.parse_args()
    root = a.scratch.resolve()
    if not root.is_relative_to(REPO / ".tmp"):
        p.error("repository .tmp input required")
    out = HERE / "evidence"
    out.mkdir(exist_ok=True)
    for group in ("permissions", "claude", "codex-loader", "claude-loader", "muse-resume", "controls"):
        src = root / (group + "-01")
        receipt = json.loads((src / "receipt.json").read_text())
        if receipt["exit"] != 0:
            raise ValueError("incomplete replay: " + group)
        log = (src / "replay.log").read_text()
        if "--- FAIL:" in log or "--- SKIP:" in log or "\nPASS\n" not in log:
            raise ValueError("failed, skipped or missing completed replay: " + group)
        log = log.replace(str(REPO), "$REPO")
        write(out / (group + "-replay.log"), log)
        write(out / (group + "-receipt.json"), receipt)
        write(out / (group + "-sources.json"), json.loads((src / "source-sha256.json").read_text()))
    # Existing collectors export request digests and synthetic effects only.
    subprocess.run([sys.executable, str(REPO / ".research/permissions-memory/collect.py"), str(out / "permissions-memory.json"), "measured=" + str(root / "permissions-01/runs")], check=True)
    subprocess.run([sys.executable, str(REPO / ".research/claude-permissions-memory/collect.py"), str(root / "claude-01/runs"), str(out / "claude.json")], check=True)
    for name in ("permissions-memory.json", "claude.json"):
        write(out / name, json.loads((out / name).read_text()))
    loader = {}
    for group in ("codex-loader", "claude-loader"):
        rows = []
        for path in sorted((root / (group + "-01/runs")).glob("*/evidence/*.json")):
            value = json.loads(path.read_text())
            raw = path.read_bytes()
            row = {"case": path.parent.parent.name, "record": path.name, "sha256_sanitized_record": digest(raw)}
            if group == "codex-loader":
                if path.name not in ("config-result.json", "hooks-result.json", "permission-thread.json", "mcp-status.json"):
                    continue
                row["facts"] = {marker: marker.encode() in raw for marker in ("PROJECT_HOOK", "APPROVED_HOOK", "project_tool", "approved_tool", '"approval_policy":"never"', '"type":"readOnly"')}
                if path.name == "mcp-status.json":
                    row["server_names"] = [s["name"] for s in value["data"]]
            else:
                if "requests" not in value:
                    continue
                row["exit"] = value["exit"]
                row["request_count"] = len(value["requests"])
                row["native_mcp"] = value["init"].get("mcp_servers")
                row["effects"] = {name: (path.parent.parent / (name + ".witness")).exists() for name in ("approved-hook", "foreign-hook", "approved-mcp", "foreign-mcp")}
            rows.append(row)
        if not rows:
            raise ValueError("empty loader evidence")
        loader[group] = rows
    write(out / "loader-observations.json", loader)
    startup = json.loads((root / "startup-02/observations.json").read_text())
    write(out / "startup-observations.json", startup)
    muse = root / "muse-resume-01/runs/TestP7ResumeTrustGate"
    write(out / "muse-resume-observations.json", {
        "approved_hook_effect": (muse / "approved-hook").read_text() == "MUSE_EXCL_HOOK\n",
        "project_hook_effect_after_trust": (muse / "project-hook").read_text() == "MUSE_EXCL_HOOK\n",
        "resume_trust_prompt": "Do you trust this workspace?" in (muse / "resume-trust.txt").read_text(),
        "resume_offers_continue_untrusted": "Continue without trusting" in (muse / "resume-trust.txt").read_text(),
        "source": "muse-resume-replay.log: TestP7ResumeTrustGate; startup catalog counts plus real marker-hook effects",
        "admitted": False,
    })
    manifest = {str(path.relative_to(HERE)): digest(path.read_bytes()) for path in sorted(out.iterdir()) if path.name != "manifest.json" and path.is_file()}
    write(out / "manifest.json", manifest)
    print("published bounded evidence; admission=false")


if __name__ == "__main__":
    main()
