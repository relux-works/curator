#!/usr/bin/env bash
# The Linux runner's job-completed hook removes TMPDIR, not RUNNER_TEMP.
# Restore permissions only; leave removal to the runner owner.
set -u
root=${TMPDIR-}
if [[ -z "$root" || ! -d "$root" || -L "$root" ]]; then
  echo 'runner tmp restore: refusing unset, empty, non-directory or symlink TMPDIR'
  exit 0
fi
# Resolve aliases such as /tmp/.. before checking for the filesystem root.
root=$(cd -- "$root" && pwd -P) || { echo 'runner tmp restore: cannot resolve TMPDIR'; exit 0; }
if [[ "$root" == / ]]; then
  echo 'runner tmp restore: refusing filesystem root TMPDIR'
  exit 0
fi
# GNU chmod does not traverse symlinks encountered recursively. No -L and no
# trailing slash on the resolved root: targets outside TMPDIR stay untouched.
echo "runner tmp restore: runner=${RUNNER_NAME:-unknown}; restoring owner write permission"
chmod -R u+w -- "$root" || echo 'runner tmp restore: chmod could not restore every entry'
exit 0
