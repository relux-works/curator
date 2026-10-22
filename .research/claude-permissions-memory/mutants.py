#!/usr/bin/env python3
"""Expected-red controls: require the named failure, never skips or setup errors."""
import json
import os
from pathlib import Path
import subprocess
import time

env = os.environ.copy()
if not env.get("CLAUDE_QUAL_BINARY") or env.get("CLAUDE_QUAL_VERSION") not in {"2.1.305", "2.1.307"}:
    raise SystemExit("exact binary and version opt-in required")
base = Path(env["CLAUDE_QUAL_RUN_ROOT"])
if base.exists():
    raise SystemExit("fresh run root required")
base.mkdir(parents=True)
for key in ["GIT_DIR", "GIT_WORK_TREE"]:
    env.pop(key, None)
controls = []
for form in ["fresh", "resume", "child"]:
    for atom in ["Bash", "Edit", "Write", "WebFetch", "MCP"]:
        name = f"TestPermissions/{form}/{atom}/deny-true"
        tool = "mcp__approved-probe__sentinel" if atom == "MCP" else atom
        controls.append((f"deny-{atom}-{form}", f"^{name.replace('/', '$/^')}$", f"deny-{atom}",
                         "observed No such tool available: " + tool + "=false; want true", name))
controls.append(("ceiling", "^TestBuiltinCeiling$", "ceiling",
                 "builtin catalog ceiling bypassed:", "TestBuiltinCeiling"))
for form in ["fresh"]:
    name = f"TestMemory/off-true/{form}"
    controls.append((f"memory-{form}", f"^{name.replace('/', '$/^')}$", "memory",
                     "observed QUAL_AUTOMATIC_MEMORY_INPUT_9821=true; want false", name))
for atom, witness in [("network", "unrelated loopback not denied"), ("keychains", "synthetic Keychains read not denied")]:
    controls.append((f"boundary-{atom}", "^TestOfflineBoundary$", "boundary-" + atom,
                     witness, "TestOfflineBoundary"))
results = []
for label, pattern, mutant, witness, test in controls:
    child = env.copy()
    child["CLAUDE_QUAL_RUN_ROOT"] = str(base / label)
    if mutant.startswith("boundary-"):
        child["CLAUDE_QUAL_BOUNDARY_MUTANT"] = mutant.removeprefix("boundary-")
    else:
        child["CLAUDE_QUAL_MUTANT"] = mutant
    command = [env.get("CURATOR_GO", "go"), "test", "./.research/claude-permissions-memory",
               "-run", pattern, "-v", "-count=1", "-timeout=90s"]
    start = time.monotonic()
    run = subprocess.run(command, env=child, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
    (base / label).mkdir(parents=True, exist_ok=True)
    (base / label / "go-test.log").write_text(run.stdout)
    applied = run.returncode == 1 and witness in run.stdout and f"--- FAIL: {test}" in run.stdout
    applied = applied and "--- SKIP:" not in run.stdout and "build failed" not in run.stdout
    receipt = {"control": label, "exit": run.returncode, "seconds": round(time.monotonic() - start, 3),
               "intended_failure": applied, "witness": witness}
    if mutant.startswith("deny-") and not mutant.endswith("WebFetch"):
        fixture = base / label / test.replace("/", "_")
        if mutant in {"deny-Bash", "deny-Edit", "deny-Write"}:
            receipt["forbidden_effect_observed"] = (fixture / "ancestor/work/effect.txt").read_text() == "QUAL_EFFECT_9821\n"
        elif mutant == "deny-MCP":
            receipt["forbidden_effect_observed"] = "MCP_TOOL_SENTINEL" in (fixture / "evidence/turn.json").read_text()
        if "forbidden_effect_observed" in receipt:
            applied = applied and receipt["forbidden_effect_observed"]
    if mutant == "ceiling":
        fixture = base / label / "TestBuiltinCeiling_empty"
        receipt["forbidden_effect_observed"] = (fixture / "ancestor/work/ceiling.txt").read_text() == "QUAL_CEILING_EFFECT_9821"
        applied = applied and receipt["forbidden_effect_observed"]
    receipt["intended_failure"] = applied
    results.append(receipt)
    print(json.dumps(receipt), flush=True)
    if not applied:
        raise SystemExit("control did not reach its intended failure: " + label)
(base / "summary.json").write_text(json.dumps(results, indent=2) + "\n")
