#!/usr/bin/env python3
"""Publish bounded native observations, never raw diagnostics or host locators."""
import hashlib
import json
from pathlib import Path
import re
import sys

root, output = map(Path, sys.argv[1:])
markers = ["EXCL_APPROVED_ROOT_7352", "QUAL_NATIVE_PROJECT_INPUT_9821",
           "QUAL_AUTOMATIC_MEMORY_INPUT_9821", "QUAL_CHILD_INPUT_9821",
           "QUAL_ORDINARY_READ_9821", "MCP_TOOL_SENTINEL"]
records = []
identity = None
for path in sorted(root.glob("*/evidence/*.json")):
    item = json.loads(path.read_text())
    if path.name == "identity.json":
        identity = item
        continue
    if "requests" not in item:
        continue
    requests = item["requests"]
    row = {"test": path.parent.parent.name, "turn": path.stem,
           "argv": item["argv"], "exit": item["exit"],
           "request_count": len(requests),
           "request_sha256": hashlib.sha256(json.dumps(requests, sort_keys=True).encode()).hexdigest(),
           "requests": []}
    for request in requests:
        serialized = json.dumps(request, sort_keys=True)
        errors = []
        for message in request.get("messages", []):
            content = message.get("content", [])
            if not isinstance(content, list):
                continue
            for block in content:
                if block.get("type") == "tool_result" and block.get("is_error"):
                    errors.append(str(block.get("content"))[:700])
        row["requests"].append({"model": request.get("model"),
                                "markers": {m: m in serialized for m in markers},
                                "tool_errors": errors})
    row["native_tools"] = item["init"].get("tools")
    row["permission_denials"] = item["result"].get("permission_denials")
    row["child_stats"] = item["result"].get("subagent_stats")
    row["terminal_reason"] = item["result"].get("terminal_reason")
    effects = {}
    fixture = path.parent.parent
    for candidate in [fixture / "ancestor/work/effect.txt", fixture / "ancestor/work/ceiling.txt",
                      fixture / "approved.witness", fixture / "ancestor/work/ordinary.txt",
                      fixture / "ancestor/work/shell.txt"]:
        if candidate.exists():
            effects[candidate.name] = candidate.read_text()
    for candidate in fixture.glob("home/.claude/projects/*/memory/*"):
        effects["memory/" + candidate.name] = candidate.read_text()
    row["final_fixture_effects"] = effects
    records.append(row)
if not identity or not records:
    raise SystemExit("missing identity or measured vendor turns")
result = {"identity": identity, "records": records}
body = json.dumps(result, indent=2, sort_keys=True) + "\n"
body = body.replace("$RUN/home/", "$SYNTHETIC_HOME/")
body = body.replace("$RUN_ID-home-", "$SYNTHETIC_HOME_ID-")
if re.search(r"/Users/|/home/[^/$ ]|[-]Users[-]|[-]home[-]", body):
    raise SystemExit("publication contains an ordinary or flattened home locator")
output.parent.mkdir(parents=True, exist_ok=True)
output.write_text(body)
print(f"records={len(records)} sha256={hashlib.sha256(body.encode()).hexdigest()}")
