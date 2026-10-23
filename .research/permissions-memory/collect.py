"""Project bounded synthetic witnesses, never native prompts, headers or credentials."""
import hashlib
import json
import os
from pathlib import Path
import re
import sys

MARKERS = ["QUAL_CODEX_MEMORY_CANARY", "QUAL_CODEX_PROFILE", "QUAL_PROJECT_DISCOVERY",
           "QUAL_CODEX_MEMORY_WRITE", "QUAL_MUSE_MEMORY_READ", "QUAL_MUSE_MEMORY_WRITE",
           "MCP_TOOL_SENTINEL"]
SIGNALS = ["sandbox_apply: Operation not permitted", "MCP tool call requires approval, but approval policy is never",
           "unsupported call: mcp__approvedsentinel", "failed to bind managed network proxy",
           "QUAL_MUSE_LOCAL_OK"]


def project(root):
    rows = []
    for path in sorted(root.glob("*/evidence/*.json")):
        value = json.loads(path.read_text())
        row = {"case": path.parent.parent.name, "record": path.name,
               "sha256_sanitized_record": hashlib.sha256(path.read_bytes()).hexdigest()}
        if path.name in ["verdict.json", "identity.json", "catalog-route.json", "candidate.json", "environment.json"]:
            row["observation"] = value
        elif path.name.endswith("requests.json"):
            row["requests"] = []
            for request in value:
                raw = json.dumps(request, sort_keys=True)
                if os.environ.get("QUAL_COLLECT_MUTANT")=="prompt":
                    row["unsafe_prompt_plant"] = request
                row["requests"].append({"sha256": hashlib.sha256(raw.encode()).hexdigest(),
                                        "markers": {marker: marker in raw for marker in MARKERS}})
        elif isinstance(value, dict) and "argv" in value:
            raw = json.dumps(value, sort_keys=True)
            row.update(exit=value["exit"], timeout=value.get("timeout", False),
                       signals={signal: signal in raw for signal in SIGNALS})

        else:
            continue
        rows.append(row)
    if not rows:
        raise ValueError("no native records; refusing empty evidence")
    return rows


def main():
    if len(sys.argv) < 3:
        raise SystemExit("usage: collect.py OUTPUT_JSON LABEL=RUN_ROOT [...]")
    records = {}
    for argument in sys.argv[2:]:
        label, root = argument.split("=", 1)
        if not re.fullmatch(r"[a-z0-9-]+", label) or label in records:
            raise SystemExit("invalid or duplicate run label")
        records[label] = project(Path(root))
    output = {"format": "permissions-memory-observations-v1", "runs": records,
              "interpretation": "Bounded exact-build facts only. Unknown and refused mandatory atoms prevent native-project certification. No whole prompt, API header or real credential is projected."}
    body = json.dumps(output, indent=2) + "\n"
    if re.search(r"/Users/|(?<![A-Za-z0-9_$])/home/[A-Za-z0-9_-]+/|-Users-[^-]+-|[A-Za-z0-9]+_curator_(?:dev|rev|orch)", body):
        raise SystemExit("private locator or seat remains; refusing publication")
    Path(sys.argv[1]).write_text(body)
    print("projected " + str(sum(map(len, records.values()))) + " bounded records")


if __name__ == "__main__":
    main()
