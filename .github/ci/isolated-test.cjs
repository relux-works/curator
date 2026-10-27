// Read the OS environment, including names outside Bash's variable table.
const { mkdtempSync, realpathSync } = require('node:fs');
const { tmpdir } = require('node:os');
const { join } = require('node:path');
const { spawnSync } = require('node:child_process');

let entries;
try {
  entries = Object.entries(process.env);
  if (entries.length === 0) throw new Error('empty environment');
} catch {
  console.error('isolated-test: raw environment enumeration failed');
  process.exit(1);
}
const families = ['GIT_', 'SSH_', 'XDG_'];
const selectors = new Set(['GOFLAGS', 'CODEX_HOME', 'CLAUDE_CONFIG_DIR', 'PI_CODING_AGENT_DIR']);
const childEnv = Object.fromEntries(entries.filter(([name]) =>
  !families.some(prefix => name.startsWith(prefix)) && !selectors.has(name)));
childEnv.HOME = realpathSync(mkdtempSync(join(tmpdir(), 'curator-test-home-')));
const [command, ...args] = process.argv.slice(2);
if (!command) {
  console.error('isolated-test: command required');
  process.exit(1);
}
const result = spawnSync(command, args, { env: childEnv, stdio: 'inherit' });
if (result.error) {
  console.error('isolated-test: child launch failed');
  process.exit(1);
}
if (result.signal) process.kill(process.pid, result.signal);
process.exit(result.status ?? 1);
