#!/usr/bin/env python3
"""Replay existing offline witnesses on frozen bytes; never grant admission."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess

HERE = Path(__file__).resolve().parent
REPO = HERE.parents[2]
PINS = {
    "codex": "8813367285524ac798dd96d1b071aea30d58708c7226fd236cd275cf1cfdee31",
    "claude": "85d8c08707fcfd1e78ffa0204eefbc6ec9e4fe22f375fb2767ef85d957ed3161",
    "muse": "9fd40ae9c6dfb847e3ef58b61405c004d0ebbf19ceb0297be966f750b3b84aa0",
}
GROUPS = {
    "permissions": ("permissions-memory", {"QUAL_CODEX_BINARY": "codex", "QUAL_MUSE_BINARY": "muse"}, "QUAL_RUN_ROOT", "Test(NativeIntrospection|OfflineBoundary|NestedSandboxControl|Codex.*|Muse.*|FilesystemDenialAttribution|RequestAttribution|FunctionResults|VerdictGate|Intake)$"),
    "claude": ("claude-permissions-memory", {"CLAUDE_QUAL_BINARY": "claude"}, "CLAUDE_QUAL_RUN_ROOT", "Test(Identity|IntakeBoundary|OfflineBoundary|Permissions|ProjectPermissions|BuiltinCeiling|Memory|MemoryDirectedWrite|MemoryResumeHistory)$"),
    "muse-resume": ("muse1413/resume-trust", {"MUSE1413_RESUME_BINARY": "muse"}, "MUSE1413_RESUME_RUN_ROOT", "Test(Checker|OfflineBoundary|P7ResumeTrustGate)$"),
    "codex-loader": ("project-exclusion-codex", {"CODEX_EXCL_BINARY": "codex"}, "CODEX_EXCL_RUN_ROOT", "Test(ConfigLayers|OfflineBoundary|MCPRuntime|ResumeApprovedRoot)$"),
    "claude-loader": ("project-exclusion-claude", {"CLAUDE_EXCL_BINARY": "claude"}, "CLAUDE_EXCL_RUN_ROOT", "Test(Identity|OfflineBoundary|P4Hooks|P5MCP)$"),
}


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("group", choices=GROUPS)
    p.add_argument("--binaries", type=Path, required=True, help="directory of read-only copies: codex, claude, muse")
    p.add_argument("--scratch", type=Path, required=True, help="fresh directory under repository .tmp")
    p.add_argument("--controls", action="store_true", help="permissions group: run the existing 17 bounded narrowing controls")
    a = p.parse_args()
    scratch = a.scratch.resolve()
    if not scratch.is_relative_to(REPO / ".tmp") or scratch.exists():
        p.error("fresh repository .tmp scratch required")
    source, keys, root_key, pattern = GROUPS[a.group]
    if a.controls:
        if a.group != "permissions":
            p.error("--controls requires permissions")
        pattern = "TestNarrowing$"
    binaries = {}
    for key, name in keys.items():
        binary = (a.binaries / name).resolve(strict=True)
        if hashlib.sha256(binary.read_bytes()).hexdigest() != PINS[name]:
            p.error("candidate digest differs; no execution or admission")
        binaries[key] = str(binary)
    scratch.mkdir(parents=True)
    stage = scratch / "stage"
    stage.mkdir()
    origin = REPO / ".research" / source
    provenance = {}
    for file in sorted(origin.rglob("*")):
        if not file.is_file() or not (file.name.endswith("_test.go") or file.name == "collect.py" or "testdata" in file.parts):
            continue
        if a.group == "codex-loader" and file.name.endswith("_test.go") and file.name != "replay_test.go":
            continue
        data = file.read_bytes()
        provenance[str(file.relative_to(REPO))] = hashlib.sha256(data).hexdigest()
        target = stage / file.relative_to(origin)
        target.parent.mkdir(parents=True, exist_ok=True)
        if file.name.endswith("_test.go"):
            text = data.decode().replace('".temp"', '".tmp"').replace('/.temp/', '/.tmp/')
            if a.group == "permissions":
                text = text.replace("7118b9ca2227a6ec1b90cc506f8abdd4a35d14d8ab6683c9164b075002a55fff", PINS["muse"])
                text = text.replace('"1.4.12"', '"1.4.13"')
            if a.group == "claude":
                text = text.replace('"2.1.307": "5b113a7752119470f1da419cb0e608af3361cae25b614216b13022093958fb2a",', '"2.1.311": "' + PINS["claude"] + '",')
            if a.group == "codex-loader":
                text = text.replace("74c6a6d263f40d65f99c27a815275f5b4ed0bcc2baecc396b4cf3255498bdfc0", PINS["codex"]).replace("0.162.1", "0.163.0")
            if a.group == "claude-loader":
                text = text.replace("ea3ea3c08d2a9634a4651e687fcf900d410c489bce2b00f22bd7352b91ba704f", PINS["claude"]).replace("2.1.305", "2.1.311")
            target.write_text(text)
        else:
            target.write_bytes(data)
    (scratch / "source-sha256.json").write_text(json.dumps(provenance, indent=2) + "\n")
    expected = json.loads((HERE / "evidence" / (a.group + "-sources.json")).read_text())
    if provenance != expected:
        p.error("source witnesses changed; review and requalification required before execution")
    for d in ("home", "tmp", "cache", "mod", "gopath"):
        (scratch / d).mkdir()
    env = {"PATH": os.environ.get("PATH", "/usr/bin:/bin"), "HOME": str(scratch / "home"),
           "GOCACHE": str(REPO / ".tmp" / "qualification-go-cache"), "GOMODCACHE": str(scratch / "mod"),
           "GOPATH": str(scratch / "gopath"), "TMPDIR": str(scratch / "tmp"), "GOTOOLCHAIN": "local",
           "GOTELEMETRY": "off", "GOENV": "off", "GOWORK": "off", "GOPROXY": "off", "GOSUMDB": "off",
           "CLAUDE_QUAL_VERSION": "2.1.311", root_key: str(scratch / "runs"), **binaries}
    if a.controls:
        env["QUAL_RUN_MUTANTS"] = "1"
    # The harness must deny the actual native HOME, not merely the Go driver's
    # synthetic HOME. os.UserHomeDir uses HOME; pass the real locator to the
    # driver only. Every vendor still runs env -i with a separate scratch HOME.
    env["HOME"] = str(Path.home())
    cmd = [shutil.which("go"), "test", "./" + str(stage.relative_to(REPO)), "-count=1", "-v", "-timeout=8m", "-run", "^" + pattern]
    with (scratch / "replay.log").open("w") as log:
        child = subprocess.Popen(cmd, cwd=REPO, env=env, stdout=log, stderr=subprocess.STDOUT)
        (scratch / "pid.txt").write_text(str(child.pid) + "\n")
        code = child.wait(timeout=500)
    receipt = {"group": a.group, "exit": code, "pins": {name: PINS[name] for name in keys.values()}, "admitted": False}
    (scratch / "receipt.json").write_text(json.dumps(receipt, indent=2) + "\n")
    print(json.dumps(receipt), flush=True)
    raise SystemExit(code)


if __name__ == "__main__":
    main()
