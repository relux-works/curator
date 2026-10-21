"""Synthetic stdio MCP fixture. No network, imports of operator state, or auth."""
import json
import os
import sys

name, receipt = sys.argv[1:]
with open(receipt, "w", encoding="utf-8") as output:
    json.dump({"pid": os.getpid(), "name": name}, output)
for line in sys.stdin:
    request = json.loads(line)
    if "id" not in request:
        continue
    method = request["method"]
    if method == "initialize":
        result = {
            "protocolVersion": "2024-11-05",
            "capabilities": {"tools": {}},
            "serverInfo": {"name": name, "version": "1"},
        }
    elif method == "tools/list":
        result = {"tools": [{"name": name, "description": name, "inputSchema": {"type": "object"}}]}
    elif method == "tools/call":
        result = {"content": [{"type": "text", "text": name + "_CALLED"}]}
    elif method in ("resources/list", "resources/templates/list"):
        result = {"resources" if method == "resources/list" else "resourceTemplates": []}
    elif method == "ping":
        result = {}
    else:
        print(json.dumps({"jsonrpc": "2.0", "id": request["id"], "error": {"code": -32601, "message": "unsupported fixture method"}}), flush=True)
        continue
    print(json.dumps({"jsonrpc": "2.0", "id": request["id"], "result": result}), flush=True)
