# Synthetic stdio server: records process admission and advertises one inert tool.
import json
import sys
from pathlib import Path

Path(sys.argv[1]).write_text("MCP_STARTED\n")
for line in sys.stdin:
    request = json.loads(line)
    if "id" not in request:
        continue
    method = request.get("method")
    result = {}
    if method == "initialize":
        result = {"protocolVersion": "2024-11-05", "capabilities": {"tools": {}},
                  "serverInfo": {"name": "exclusion-fixture", "version": "1.0.0"}}
    elif method == "tools/list":
        result = {"tools": [{"name": "sentinel", "description": "Synthetic admission witness",
                             "inputSchema": {"type": "object", "properties": {}},
                             "annotations": {"readOnlyHint": True, "destructiveHint": False, "idempotentHint": True, "openWorldHint": False}}]}
    elif method == "tools/call":
        Path(sys.argv[1]+".called").write_text("MCP_CALLED\n")
        result = {"content": [{"type": "text", "text": "MCP_TOOL_SENTINEL"}]}
    print(json.dumps({"jsonrpc": "2.0", "id": request["id"], "result": result}), flush=True)
