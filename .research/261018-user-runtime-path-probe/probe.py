#!/usr/bin/env python3
"""One-shot macOS runtime path probe. No prompt submission or model request."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile
import sys

OLD_ATIME = 946684800000000000


def write(path, content):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content)


def environment(root, home, config, runtime):
    # Deliberately do not inherit authentication, session, proxy or integration vars.
    result = {"PATH": "/usr/bin:/bin:/opt/homebrew/bin", "HOME": str(home),
              "TMPDIR": str(root / "tmp"), "TERM": "dumb", "LANG": "en_US.UTF-8",
              "XDG_CONFIG_HOME": str(root / "xdg/config"),
              "XDG_CACHE_HOME": str(root / "xdg/cache"),
              "XDG_DATA_HOME": str(root / "xdg/data"),
              "DISABLE_AUTOUPDATER": "1", "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
              "CLAUDE_CODE_TMPDIR": str(root / "tmp"), "PYTHONDONTWRITEBYTECODE": "1"}
    if config is not None:
        result["CODEX_HOME" if runtime == "codex" else "CLAUDE_CONFIG_DIR"] = str(config)
    return result


def policy(root):
    # Network is denied, and only this new scratch tree is writable.
    return ('(version 1) (deny default) (allow process*) (allow file-read*) '
            '(allow sysctl-read) (allow mach-lookup) '
            '(allow file-write* (subpath ' + json.dumps(str(root)) + ')) (deny network*)')


def run(root, args, env, cwd, timeout=25):
    p = subprocess.Popen(['/usr/bin/sandbox-exec', '-f', str(root / 'probe.sb'), *args],
                         cwd=cwd, env=env, stdin=subprocess.DEVNULL,
                         stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    with (root / 'pids.jsonl').open('a') as f:
        f.write(json.dumps({'pid': p.pid, 'entrypoint': Path(args[0]).name,
                            'args': args[1:]}) + '\n')
    timed_out = False
    try:
        stdout, stderr = p.communicate(timeout=timeout)
    except subprocess.TimeoutExpired:
        timed_out = True
        p.kill()  # Exact owned PID; no pattern or shared process kill.
        stdout, stderr = p.communicate()
    except BaseException:
        if p.poll() is None:
            p.kill()
        p.communicate()
        raise
    return {'exit': p.returncode, 'timed_out': timed_out,
            'stdout': stdout.decode(errors='replace'), 'stderr': stderr.decode(errors='replace')}


def plant(root, runtime):
    candidates = {}
    homes = [root / 'home', root / 'shifted-home']
    configs = [homes[0] / ('.codex' if runtime == 'codex' else '.claude'), root / 'managed']
    for d in [*homes, *configs, root / 'tmp', root / 'project']:
        d.mkdir(parents=True, exist_ok=True)
    (root / 'project/.git').mkdir()

    def put(path, content):
        write(path, content)
        candidates[str(path.relative_to(root))] = path

    for label, folder in [('home-codex', homes[0] / '.codex'),
                          ('home-agents', homes[0] / '.agents'),
                          ('home-claude', homes[0] / '.claude'),
                          ('managed', configs[1]),
                          ('managed-agents', configs[1] / '.agents'),
                          ('shifted-agents', homes[1] / '.agents')]:
        put(folder / ('skills/' + label + '/SKILL.md'),
            '---\nname: sentinel-' + label + '\ndescription: Sentinel ' + label + '\n---\n')
        put(folder / 'AGENTS.md', 'SENTINEL_AGENTS_' + label + '\n')
        put(folder / 'CLAUDE.md', 'SENTINEL_CLAUDE_' + label + '\n')
    put(root / 'project/AGENTS.md', 'SENTINEL_PROJECT_AGENTS\n')
    put(root / 'project/CLAUDE.md', 'SENTINEL_PROJECT_CLAUDE\n')
    for label, folder in [('default', configs[0]), ('managed', configs[1])]:
        if runtime == 'codex':
            put(folder / 'config.toml', 'developer_instructions = "SENTINEL_CONFIG_' + label + '"\n[analytics]\nenabled = false\n[feedback]\nenabled = false\n')
            put(folder / 'hooks.json', json.dumps({'hooks': {'SessionStart': [{'hooks': [
                {'type': 'command', 'command': 'echo SENTINEL_HOOK_' + label}]}]}}))
        else:
            put(folder / 'settings.json', json.dumps({'env': {'PROBE_SENTINEL': label},
                'enabledPlugins': {'sentinel-' + label + '@fixture': True},
                'hooks': {'SessionStart': [{'hooks': [{'type': 'command', 'command':
                    'printf SENTINEL_HOOK_' + label + ' >> "$HOME/hook-observed"'}]}]}}))
        if runtime == 'claude':
            plugin = folder / ('plugins/cache/fixture/sentinel-' + label + '/1.0.0')
            put(plugin / '.claude-plugin/plugin.json', json.dumps({'name': 'sentinel-' + label,
                'version': '1.0.0', 'description': 'Harmless scratch fixture'}))
            put(plugin / 'skills/sentinel/SKILL.md', '---\nname: plugin-sentinel-' + label + '\ndescription: Plugin positive control\n---\n')
            market = folder / 'plugins/marketplaces/fixture'
            put(market / '.claude-plugin/marketplace.json', json.dumps({'name': 'fixture',
                'owner': {'name': 'probe'}, 'plugins': [{'name': 'sentinel-' + label,
                    'source': './plugins/sentinel'}]}))
            put(market / 'plugins/sentinel/.claude-plugin/plugin.json', json.dumps(
                {'name': 'sentinel-' + label, 'version': '1.0.0'}))
            put(market / 'plugins/sentinel/skills/sentinel/SKILL.md',
                '---\nname: plugin-sentinel-' + label + '\ndescription: Plugin positive control\n---\n')
            put(folder / 'plugins/known_marketplaces.json', json.dumps({'fixture': {
                'source': {'source': 'directory', 'path': str(market)},
                'installLocation': str(market), 'lastUpdated': '2026-01-01T00:00:00Z'}}))
            put(folder / 'plugins/installed_plugins.json', json.dumps({'version': 2, 'plugins': {
                'sentinel-' + label + '@fixture': [{'scope': 'user', 'installPath': str(plugin),
                    'version': '1.0.0', 'installedAt': '2026-01-01T00:00:00Z',
                    'lastUpdated': '2026-01-01T00:00:00Z'}]}}))
        # A valid local marketplace with a real, harmless plugin.
        manifest = folder / '.agents/plugins/marketplace.json'
        put(manifest, json.dumps({'name': 'sentinel-' + label, 'plugins': [
            {'name': 'sentinel-plugin-' + label,
             'source': {'source': 'local', 'path': './plugins/sentinel'}}]}))
        put(folder / 'plugins/sentinel/.codex-plugin/plugin.json', json.dumps(
            {'name': 'sentinel-plugin-' + label, 'description': 'Local probe'}))
    # HOME-root marketplace: independent of the runtime configuration directory.
    put(homes[0] / '.agents/plugins/marketplace.json', json.dumps({'name': 'sentinel-home',
        'plugins': [{'name': 'sentinel-plugin-home', 'source': {'source': 'local', 'path': './plugins/sentinel'}}]}))
    put(homes[0] / 'plugins/sentinel/.codex-plugin/plugin.json', json.dumps(
        {'name': 'sentinel-plugin-home', 'description': 'Local probe'}))
    put(homes[1] / '.agents/plugins/marketplace.json', json.dumps({'name': 'sentinel-shifted',
        'plugins': [{'name': 'sentinel-plugin-shifted', 'source': {'source': 'local', 'path': './plugins/sentinel'}}]}))
    put(homes[1] / 'plugins/sentinel/.codex-plugin/plugin.json', json.dumps(
        {'name': 'sentinel-plugin-shifted', 'description': 'Local shifted control'}))
    put(root / 'unread-control.txt', 'No runtime references this control.\n')
    return homes, configs, candidates


def assess(row):
    """Check named positive controls and neighboring forbidden discoveries, per lane."""
    problems = []
    def require(condition, claim):
        if not condition:
            problems.append(claim)
    reads = row['file_reads']
    runtime, mode = row['runtime'], row['mode']
    prefix = 'home/.' + ('codex' if runtime == 'codex' else 'claude') if mode == 'default' else 'managed'
    require(row['atime_positive'], 'atime instrument did not detect its direct-read control')
    require(not reads['unread-control.txt'], 'unreferenced negative control was read')
    for name in ['version_readback', 'plugins', 'prompt' if runtime == 'codex' else 'init']:
        command = row.get(name, {})
        require(command.get('exit') == 0 and not command.get('timed_out'), name + ' did not finish green')
    require('error' not in row, row.get('error', 'lane exception'))
    config = 'config.toml' if runtime == 'codex' else 'settings.json'
    require(reads.get(prefix + '/' + config), 'selected configuration positive control unread')
    label = ('home-codex' if runtime == 'codex' else 'home-claude') if mode == 'default' else 'managed'
    require(reads.get(prefix + '/skills/' + label + '/SKILL.md'), 'selected skill positive control unread')
    try:
        plugins = json.loads(row['plugins']['stdout'])
    except (KeyError, ValueError):
        plugins = None
        require(False, 'plugin output is not JSON')
    if runtime == 'codex':
        require(row.get('version_readback', {}).get('stdout', '').strip() == 'codex-cli ' + row['version'],
                'Codex version readback differs from the requested exact version')
        prompt = row.get('prompt', {}).get('stdout', '')
        require('sentinel-' + label in prompt, 'selected skill absent from rendered prompt')
        home_label = 'shifted-agents' if mode == 'override-shifted-home' else 'home-agents'
        require('sentinel-' + home_label in prompt, 'HOME skill positive control absent')
        instruction = 'AGENTS.override.md' if mode == 'override-instructions-precedence' else 'AGENTS.md'
        require(reads.get(prefix + '/' + instruction), 'selected global instruction unread')
        require(reads.get(prefix + '/hooks.json'), 'selected hooks file unread')
        if mode == 'override-instructions-precedence':
            require('SENTINEL_AGENTS_OVERRIDE' in prompt, 'override instruction absent')
            require('SENTINEL_AGENTS_managed' not in prompt, 'lower-priority instruction leaked')
        else:
            require('SENTINEL_AGENTS_' + label in prompt, 'selected global instruction absent from prompt')
        home_prefix = 'shifted-home' if mode == 'override-shifted-home' else 'home'
        require(reads.get(home_prefix + '/.agents/plugins/marketplace.json'), 'HOME marketplace positive control unread')
        plugin_label = 'shifted' if mode == 'override-shifted-home' else 'home'
        require(isinstance(plugins, dict) and any(p.get('name') == 'sentinel-plugin-' + plugin_label for p in plugins.get('available', [])), 'HOME plugin positive control absent')
        if mode == 'override-shifted-home':
            require(not reads['home/.agents/skills/home-agents/SKILL.md'], 'old HOME skill read after HOME shift')
            require(not reads['home/.agents/plugins/marketplace.json'], 'old HOME marketplace read after HOME shift')
    else:
        label = 'default' if mode == 'default' else 'managed'
        require(reads.get(prefix + '/CLAUDE.md'), 'selected global instruction unread')
        require(row['hook_output'] == 'SENTINEL_HOOK_' + label, 'selected hook did not execute')
        for file in ['installed_plugins.json', 'known_marketplaces.json']:
            require(reads.get(prefix + '/plugins/' + file), 'selected plugin registry unread: ' + file)
        require(reads.get(prefix + '/plugins/marketplaces/fixture/plugins/sentinel/skills/sentinel/SKILL.md'), 'registered plugin skill positive control unread')
        require(isinstance(plugins, list) and any(p.get('id') == 'sentinel-' + label + '@fixture' and p.get('enabled') and not p.get('errors') for p in plugins), 'registered plugin did not load without errors')
        require(not reads['home/.agents/skills/home-agents/SKILL.md'], 'unexpected generic HOME skill discovery')
    if mode != 'default':
        old_prefix = 'home/.' + ('codex' if runtime == 'codex' else 'claude')
        require(not any(v for k, v in reads.items() if k.startswith(old_prefix + '/')), 'old runtime configuration directory was read')
    return problems


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--codex', action='append', default=[], metavar='VERSION=PATH')
    parser.add_argument('--claude')
    parser.add_argument('--output', required=True)
    parser.add_argument('--control-remove-first-skill', action='store_true',
                        help='Instrument control: remove the first selected skill; all other lanes must still run')
    a = parser.parse_args()
    scratch = Path(tempfile.mkdtemp(prefix='curator-user-paths-', dir='/private/tmp'))
    print('Scratch tree:', scratch, flush=True)
    results = {'schema': 1, 'platform': 'macOS arm64', 'model_requests_submitted': 0,
               'scratch': '$SCRATCH', 'lanes': []}
    targets = [('codex', *value.split('=', 1)) for value in a.codex]
    if a.claude:
        targets.append(('claude', 'installed', a.claude))
    if not targets:
        parser.error('at least one explicit runtime binary is required')
    if any(len(t) != 3 or not re.fullmatch(r'[A-Za-z0-9._-]+', t[1]) for t in targets):
        parser.error('Codex targets must use VERSION=PATH with a simple version label')
    guard = scratch / 'guard'
    guard.mkdir()
    for part in ['home', 'tmp']:
        (guard / part).mkdir()
    write(guard / 'probe.sb', policy(guard))
    env = environment(guard, guard / 'home', None, 'codex')
    outside = scratch / 'forbidden-write'
    checks = {
        'write_positive': run(guard, ['/usr/bin/touch', str(guard / 'allowed')], env, guard),
        'write_negative': run(guard, ['/usr/bin/touch', str(outside)], env, guard),
        'network_negative': run(guard, [sys.executable, '-c',
            'import socket,sys\ntry: socket.create_connection(("127.0.0.1",9),1)\n'
            'except OSError as e: print(e.errno); sys.exit(0 if e.errno in (1,13) else 2)\n'
            'sys.exit(3)'], env, guard)}
    results['guard_controls'] = checks
    guard_ok = (checks['write_positive']['exit'] == 0
                and (guard / 'allowed').exists()
                and checks['write_negative']['exit'] != 0
                and 'Operation not permitted' in checks['write_negative']['stderr']
                and not outside.exists()
                and checks['network_negative']['exit'] == 0)
    results['guard_passed'] = guard_ok
    if not guard_ok:
        Path(a.output).write_text(json.dumps(results, indent=2) + '\n')
        print('Scratch/network guard controls failed; runtime lanes not started.', flush=True)
        return 2
    for runtime, version, executable in targets:
        binary = Path(executable).resolve()
        modes = ['default', 'override-fixed-home', 'override-shifted-home']
        if runtime == 'codex':
            modes.append('override-instructions-precedence')
        for mode in modes:
            root = scratch / (runtime + '-' + version + '-' + mode)
            root.mkdir()
            homes, configs, candidates = plant(root, runtime)
            home = homes[1] if mode == 'override-shifted-home' else homes[0]
            config = None if mode == 'default' else configs[1]
            env = environment(root, home, config, runtime)
            write(root / 'probe.sb', policy(root))
            if mode == 'override-instructions-precedence':
                override = configs[1] / 'AGENTS.override.md'
                write(override, 'SENTINEL_AGENTS_OVERRIDE\n')
                candidates['managed/AGENTS.override.md'] = override
            if a.control_remove_first_skill and not results['lanes']:
                removed = root / 'home/.codex/skills/home-codex/SKILL.md'
                if runtime != 'codex':
                    parser.error('first-skill control requires a Codex target first')
                removed.unlink()
                candidates.pop('home/.codex/skills/home-codex/SKILL.md')
            for p in candidates.values():
                os.utime(p, ns=(OLD_ATIME, p.stat().st_mtime_ns))
            # A known direct read validates atime instrumentation on this filesystem.
            control = root / 'atime-positive.txt'
            control.write_text('read control')
            os.utime(control, ns=(OLD_ATIME, control.stat().st_mtime_ns))
            control.read_bytes()
            row = {'runtime': runtime, 'version': version, 'mode': mode,
                   'atime_positive': control.stat().st_atime_ns != OLD_ATIME}
            try:
                row['binary_sha256'] = hashlib.sha256(binary.read_bytes()).hexdigest()
                row['version_readback'] = run(root, [str(binary), '--version'], env, root / 'project')
                if runtime == 'codex':
                    row['prompt'] = run(root, [str(binary), 'debug', 'prompt-input'], env, root / 'project')
                    market = 'sentinel-shifted' if mode == 'override-shifted-home' else 'sentinel-home'
                    row['plugins'] = run(root, [str(binary), 'plugin', 'list', '--marketplace', market, '--available', '--json'], env, root / 'project')
                else:
                    row['init'] = run(root, [str(binary), '--init-only', '--debug-file', str(root / 'debug.log')], env, root / 'project')
                    row['plugins'] = run(root, [str(binary), 'plugin', 'list', '--json'], env, root / 'project')
            except Exception as e:
                row['error'] = str(e)
            row['file_reads'] = {name: p.stat().st_atime_ns != OLD_ATIME for name, p in candidates.items()}
            row['hook_output'] = (home / 'hook-observed').read_text() if (home / 'hook-observed').exists() else None
            if (root / 'debug.log').exists():
                row['debug'] = (root / 'debug.log').read_text()
            results['lanes'].append(row)
            row['problems'] = assess(row)
            if not row.get('version_readback') or row['version_readback']['exit'] in (126, 127):
                row['verdict'] = 'cannot_start'
            else:
                row['verdict'] = 'failed' if row['problems'] else 'passed'
            print(runtime, version, mode, row['verdict'], '; '.join(row['problems']), flush=True)
    results['control_remove_first_skill'] = a.control_remove_first_skill
    results['overall_exit'] = int(not guard_ok or any(r['problems'] for r in results['lanes']))
    # Replace task/scratch paths before persisting the public evidence.
    payload = json.dumps(results, indent=2).replace(str(scratch), '$SCRATCH')
    for _, _, executable in targets:
        payload = payload.replace(str(Path(executable).resolve()), '$RUNTIME_BINARY')
    Path(a.output).write_text(payload + '\n')
    return results['overall_exit']


if __name__ == '__main__':
    sys.exit(main())
