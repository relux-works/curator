// Read the OS environment, including names outside Bash's variable table.
const { chmodSync, lstatSync, mkdtempSync, readdirSync, realpathSync, rmSync } = require('node:fs');
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
const [command, ...args] = process.argv.slice(2);
if (!command) {
  console.error('isolated-test: command required');
  process.exit(1);
}
const temporaryHome = realpathSync(mkdtempSync(join(tmpdir(), 'curator-test-home-')));
childEnv.HOME = temporaryHome;

// Go module caches can leave read-only directories. Restore access only in
// this invocation's owned tree; never follow a child-created symlink.
function makeDirectoriesWritable(path) {
  const entry = lstatSync(path);
  if (!entry.isDirectory()) return;
  chmodSync(path, entry.mode | 0o700);
  for (const name of readdirSync(path)) makeDirectoriesWritable(join(path, name));
}

let result;
let cleanupFailed = false;
try {
  result = spawnSync(command, args, { env: childEnv, stdio: 'inherit' });
} finally {
  try {
    // The child may already have removed its own home.
    try {
      makeDirectoriesWritable(temporaryHome);
    } catch (error) {
      if (error.code !== 'ENOENT') throw error;
    }
    rmSync(temporaryHome, { recursive: true, force: true });
  } catch {
    console.error('isolated-test: temporary home cleanup failed');
    cleanupFailed = true;
  }
}
if (result.error) {
  console.error('isolated-test: child launch failed');
  process.exit(1);
}
if (result.signal) process.kill(process.pid, result.signal);
process.exit(result.status === 0 && cleanupFailed ? 1 : (result.status ?? 1));
