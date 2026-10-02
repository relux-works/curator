# THE ONLY CURRENT INSTRUCTION — BUG-261002-ot3ea1: the build-cache sweep removes builds that live processes still execute (curator)

## Report (tb-keeper, tb-R145; ivmbp swap 2026-10-02 04:13–04:15Z)
`curator global upgrade` published the new task-board shim and then swept the previous go-v1 build (cache key 1361260be120…). At that moment 24 live processes were still executing that binary: spawn runners and `spawn wait`.

They kept running, because macOS keeps the inode. But any RE-EXEC of the absolute path failed:
- task-board's `startSpawnRunnerProcess` uses `os.Executable` for successor runners;
- the printed observation commands carry the full path.

Older tb-sessiond builds used by live daemons WERE retained, so the sweep already has some liveness check, probably for daemons only.

## Do
1. Find the sweep (curator global upgrade → the build-cache sweep / gc in internal/buildrepo or the cache module) and its current liveness rule.
2. Retain any cache build whose binary path (or any file under its build dir) is the executable of a live process.
   - Enumerate processes portably: macOS `proc_pidpath` via sysctl/libproc or `ps -axo comm=` with full paths; Linux `/proc/*/exe`; Windows the process image path.
   - An enumeration failure FAILS SAFE: keep the build, warn, never delete.
   - Optional, if simple: also keep the immediately previous build of each command for a grace period. Report which design you chose and why.
3. Tests:
   - a fake process-table seam with an in-use build → retained;
   - not in use → swept;
   - enumeration error → retained with a warning;
   - the daemon rule unchanged.

   Run an entry-point test through `curator global upgrade` / gc if one exists. Mutant: drop the liveness check → the in-use row fails.
4. Run `go test` for the touched packages and `GOOS=windows go vet` for them. Record real exit codes. One CHANGELOG line. No LOGBOOK. Never spell any employer name.

The host has syspolicyd exec stalls: check `launchctl print system/com.apple.security.syspolicy | grep -E "state|successive"` and wait while it is down. Builds go through ~/.mini-build-lock if present (R136).

## Handoff
Update the results, then run `task-board handoff BUG-261002-ot3ea1 --role developer`, then END YOUR TURN.
