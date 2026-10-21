"""Project this harness's sanitized evidence, keeping actual input and tool witnesses."""
import hashlib
import json
from pathlib import Path
import re
import sys


def texts(value):
    if isinstance(value, dict):
        if value.get("type") == "text" and isinstance(value.get("text"), str):
            yield value["text"]
        if value.get("type") == "tool_result":
            yield json.dumps(value, sort_keys=True)
        for child in value.values():
            yield from texts(child)
    elif isinstance(value, list):
        for child in value:
            yield from texts(child)


def collect(root):
    observations = []
    markers = set()
    files = sorted(root.glob("*/evidence/*.json"))
    for path in files:
        markers.update(re.findall(r"EXCL_[A-Za-z0-9_-]+", path.read_text()))
    for path in files:
        data = json.loads(path.read_text())
        if "argv" not in data:
            observations.append({"case": path.parent.parent.name, "identity": data})
            continue
        requests = []
        for index, request in enumerate(data["requests"]):
            body = json.dumps(request, sort_keys=True)
            excerpts = []
            for text in texts(request.get("messages", [])):
                if "EXCL_" not in text and "tool_result" not in text:
                    continue
                if len(text) > 3500:
                    text = "\n".join(line for line in text.splitlines()
                                     if "EXCL_" in line or "Unknown skill:" in line)
                if text not in excerpts:
                    excerpts.append(text)
            requests.append({"index": index, "model": request["model"],
                             "sha256_sanitized_request": hashlib.sha256(body.encode()).hexdigest(),
                             "markers": {marker: marker in body for marker in sorted(markers)},
                             "input_excerpts": excerpts})
        init_keys = ["claude_code_version", "model", "permissionMode", "session_id",
                     "skills", "plugins", "agents", "mcp_servers"]
        debug = [line for line in data["debug"].splitlines()
                 if ("project memory is off" in line or
                     "Successfully connected" in line or
                     "STDIO connection closed" in line or
                     "Found " in line and "plugins (" in line)]
        witnesses = {p.name: p.read_text().strip()
                     for p in sorted(path.parent.parent.glob("*.witness"))}
        observations.append({"case": path.parent.parent.name, "invocation": path.stem,
                             "argv": data["argv"], "vendor_exit": data["exit"],
                             "init": {key: data["init"].get(key) for key in init_keys},
                             "terminal": {key: data["result"].get(key) for key in
                                          ["is_error", "subtype", "terminal_reason", "permission_denials", "subagent_stats"]},
                             "requests": requests, "effect_witnesses": witnesses,
                             "loader_excerpts": debug})
    return observations


def main():
    if len(sys.argv) != 3:
        raise SystemExit("usage: collect.py RUN_ROOT OUTPUT_JSON")
    root = Path(sys.argv[1])
    observations = collect(root)
    if not observations:
        raise SystemExit("no completed vendor records; refusing empty evidence")
    body = json.dumps({"format": "claude-exclusion-observations-v1",
                       "projection": "Actual request markers, bounded input/tool excerpts, init and file effects. Full sanitized records stay in the task-local run root.",
                       "observations": observations}, indent=2) + "\n"
    # Reject both ordinary and native path-flattened private locators.
    if re.search(r"/Users/|(?<![A-Za-z0-9_$])/home/[A-Za-z0-9_-]+/|-Users-[^-]+-", body):
        raise SystemExit("private locator remains; refusing publication")
    target = Path(sys.argv[2])
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_text(body)
    print(f"collected {len(observations)} vendor records")


if __name__ == "__main__":
    main()
