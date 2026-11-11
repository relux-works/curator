#!/usr/bin/env python3
"""Bounded offline entry smoke probes; absence during startup proves no policy."""
import argparse
import errno
import fcntl
import hashlib
import json
import os
from pathlib import Path
import pty
import select
import shlex
import signal
import subprocess
import struct
import termios
import time

from replay import PINS, REPO
from candidate import check_binaries, load_candidate


def run(binary, args, root, real_home, terminal=False, seconds=6):
    home, work = root / "home", root / "project"
    for d in (home, work, root / "tmp", root / "config", root / "data", root / "state", root / "cache"):
        d.mkdir(parents=True, exist_ok=True)
    policy = '(version 1)(allow default)(deny network*)(deny file-write*)\n'
    policy += '(allow file-write* (subpath ' + json.dumps(str(root)) + ')(literal "/dev/null")(literal "/dev/tty"))\n'
    policy += '(deny file-read* file-write* (regex #"/Keychains(/|$)"))\n'
    policy += '(deny mach-lookup (global-name-regex #"^com\\.apple\\.(security|Security)"))\n'
    for d in (".curator", ".codex", ".claude", ".muse", ".config", ".local", ".agents", ".ssh", ".gnupg", "Library"):
        policy += '(deny file-read* file-write* (subpath ' + json.dumps(str(real_home / d)) + '))\n'
    for d in ("/etc/codex", "/etc/muse", "/Library/Application Support/ClaudeCode", "/Library/Application Support/Codex", "/Library/Application Support/Muse", "/Library/Managed Preferences"):
        policy += '(deny file-read* file-write* (subpath ' + json.dumps(d) + '))\n'
    for parent in (root, *root.parents):
        for name in ("CLAUDE.md", "CLAUDE.local.md", "AGENTS.md", ".claude"):
            policy += '(deny file-read* (subpath ' + json.dumps(str(parent / name)) + '))\n'
    (root / "offline.sb").write_text(policy)
    env = dict(HOME=str(home), CODEX_HOME=str(home / ".codex"), CLAUDE_CONFIG_DIR=str(home / ".claude"),
               XDG_CONFIG_HOME=str(root / "config"), XDG_DATA_HOME=str(root / "data"),
               XDG_CACHE_HOME=str(root / "cache"), XDG_STATE_HOME=str(root / "state"),
               TMPDIR=str(root / "tmp"), PATH="/usr/bin:/bin:/usr/sbin:/sbin", TERM="xterm-256color",
               LANG="en_US.UTF-8", GIT_CONFIG_NOSYSTEM="1", GIT_CONFIG_GLOBAL="/dev/null",
               DISABLE_AUTOUPDATER="1", MUSE_NO_AUTO_UPDATE="1", TBH_CREDENTIAL_BACKEND="file",
               CLAUDE_CODE_DISABLE_AUTO_MEMORY="1", CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC="1",
               ANTHROPIC_API_KEY="synthetic-not-a-real-key", ANTHROPIC_BASE_URL="http://127.0.0.1:1",
               DISABLE_TELEMETRY="1", DISABLE_ERROR_REPORTING="1")
    subprocess.run(["/usr/bin/git", "init", "-q", "--separate-git-dir=" + str(root / "metadata"), str(work)], env=env, check=True, capture_output=True)
    hook = "printf 'QUAL_STARTUP_HOOK\\n' >> " + shlex.quote(str(root / "hook-effect"))
    # Synthetic project policy challenges. A marker is an effect witness;
    # a missing marker before a usable turn is explicitly inconclusive.
    for d in (".codex", ".claude", ".muse"):
        (work / d).mkdir(exist_ok=True)
    (work / "AGENTS.md").write_text("QUAL_STARTUP_PROJECT\n")
    (work / "CLAUDE.md").write_text("QUAL_STARTUP_PROJECT\n")
    hooks = {"SessionStart": [{"hooks": [{"type": "command", "command": hook}]}]}
    (work / ".claude/settings.json").write_text(json.dumps({"hooks": hooks, "permissions": {"allow": ["Bash", "Write", "Read"]}}))
    (work / ".codex/hooks.json").write_text(json.dumps({"hooks": hooks}))
    (work / ".codex/config.toml").write_text("sandbox_mode='danger-full-access'\n[features]\nmemories=true\n[mcp_servers.project_probe]\ncommand='/usr/bin/false'\n")
    (work / ".mcp.json").write_text('{"mcpServers":{"project_probe":{"command":"/usr/bin/false"}}}')
    for d in (home / ".codex", home / ".claude"):
        d.mkdir(exist_ok=True)
    (home / ".codex/config.toml").write_text('model="offline-probe"\nmodel_provider="fixture"\ncheck_for_update_on_startup=false\ncli_auth_credentials_store="file"\n[model_providers.fixture]\nname="offline"\nbase_url="http://127.0.0.1:1"\nwire_api="responses"\nrequires_openai_auth=false\n')
    (home / ".claude/.claude.json").write_text('{"hasCompletedOnboarding":true}')
    cmd = ["/usr/bin/sandbox-exec", "-f", str(root / "offline.sb"), "/usr/bin/env", "-i", *[k + "=" + v for k, v in env.items()], str(binary), *args]
    master = slave = None
    if terminal:
        master, slave = pty.openpty()
        fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack("HHHH", 50, 160, 0, 0))
        child = subprocess.Popen(cmd, cwd=work, env={"PATH": "/usr/bin:/bin"}, stdin=slave, stdout=slave, stderr=slave, start_new_session=True)
        os.close(slave)
        os.set_blocking(master, False)
    else:
        child = subprocess.Popen(cmd, cwd=work, env={"PATH": "/usr/bin:/bin"}, stdin=subprocess.DEVNULL, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, start_new_session=True)
        os.set_blocking(child.stdout.fileno(), False)
    (root / "pid.txt").write_text(str(child.pid) + "\n")
    output = bytearray()
    deadline = time.monotonic() + seconds
    fd = master if terminal else child.stdout.fileno()
    while time.monotonic() < deadline:
        if select.select([fd], [], [], 0.1)[0]:
            try:
                block = os.read(fd, 65536)
                output.extend(block)
            except OSError as e:
                if e.errno != errno.EIO:
                    raise
                break
            if not block:
                break
            # Answer terminal cursor-position requests only, never a trust or
            # login question. PTY is private and never focuses a user terminal.
            if terminal and b"\x1b[6n" in block:
                os.write(master, b"\x1b[1;1R")
        if child.poll() is not None:
            break
    bounded_stop = child.poll() is None
    if bounded_stop:
        # Stop only the foreground PID created here. A process may finish
        # between poll and signal; its normal wait still supplies the receipt.
        try:
            child.terminate()
        except (ProcessLookupError, PermissionError):
            if child.poll() is None:
                raise
        try:
            child.wait(timeout=2)
        except subprocess.TimeoutExpired:
            child.kill()
    code = child.wait(timeout=2)
    if master is not None:
        os.close(master)
    text = output.decode(errors="replace").replace(str(root), "$RUN").replace(str(binary), "$BINARY").replace(str(real_home), "$NATIVE_HOME")
    (root / "output.txt").write_text(text)
    return {"argv": args, "exit": code, "bounded_stop": bounded_stop,
            "sha256_output": hashlib.sha256(text.encode()).hexdigest(), "pid": child.pid,
            "hook_effect": (root / "hook-effect").exists(),
            "signals": {s: s.lower() in text.lower() for s in ["trust", "login", "sign in", "not found", "sandbox_apply", "Operation not permitted", "codex-cli 0.163.0", "2.1.311 (Claude Code)", "1.4.13"]}}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binaries", type=Path, required=True)
    parser.add_argument("--scratch", type=Path, required=True)
    parser.add_argument("--candidate", type=Path, help="explicit version/SHA research pins")
    parser.add_argument("--provider", choices=PINS, help="probe only this provider")
    a = parser.parse_args()
    scratch = a.scratch.resolve()
    if not scratch.is_relative_to(REPO / ".tmp") or scratch.exists():
        parser.error("fresh repository .tmp required")
    candidate = None
    releases = {name: {"sha256": digest} for name, digest in PINS.items()}
    try:
        if a.candidate:
            candidate = load_candidate(a.candidate)
            releases = candidate["releases"]
        binaries = check_binaries(a.binaries, releases, [a.provider] if a.provider else PINS)
    except (OSError, ValueError) as error:
        parser.error(str(error))
    scratch.mkdir(parents=True)
    (scratch / "candidate-receipt.json").write_text(json.dumps({"candidate": candidate, "admitted": False}, indent=2) + "\n")
    rows = {}
    manager = ["-c", 'approval_policy="never"', "-c", 'sandbox_mode="read-only"']
    for key in ("features.memories", "features.memory_tool", "memories.generate_memories", "memories.use_memories", "memories.dedicated_tools"):
        manager += ["-c", key + "=false"]
    forms = {
        "codex": {"interactive": manager, "exec": ["exec", *manager, "QUAL_OFFLINE_STARTUP"], "resume": ["exec", "resume", *manager, "--last", "QUAL_OFFLINE_STARTUP"]},
        "claude": {"interactive": ["--tools", "Read"], "exec": ["-p", "QUAL_OFFLINE_STARTUP", "--tools", "Read"], "resume": ["-p", "QUAL_OFFLINE_STARTUP", "--tools", "Read", "--continue"]},
        "muse": {"interactive": ["--provider", "echo"], "exec": ["exec", "--json", "--provider", "echo", "QUAL_OFFLINE_STARTUP"], "resume": ["resume", "--last", "--provider", "echo"]},
    }
    for name, binary in binaries.items():
        for form, args in {"version": ["--version"], "help": ["--help"], **forms[name]}.items():
            row = run(binary, args, scratch / (name + "-" + form), Path.home(), terminal=form == "interactive", seconds=15 if form in ("help", "version") else 6)
            rows[name + "-" + form] = row
            print(name + "-" + form + ": " + str(row["exit"]) + " bounded=" + str(row["bounded_stop"]), flush=True)
    (scratch / "observations.json").write_text(json.dumps(rows, indent=2) + "\n")


if __name__ == "__main__":
    main()
