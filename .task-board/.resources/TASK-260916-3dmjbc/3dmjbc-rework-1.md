# TASK-260916-3dmjbc rework 1 (orchestrator, binding)

Revision 1 gate FAILED only on `Test (windows-latest)` (run 35574902088); Linux/macOS lanes are
green. Three Windows-only test failures, all fixture/portability, not design:

1. `internal/config :: TestParseScriptInterpreters` — `script_interpreters.node-v1: must be an
   absolute executable path`: the fixture uses a POSIX absolute path; on Windows
   `filepath.IsAbs` needs a volume (`C:\...`). Build the fixture path per-OS (e.g. from
   `t.TempDir()` or `filepath.Join(volume, ...)`), and make sure the production validation
   itself accepts Windows absolute paths and rejects drive-relative/UNC-less forms the same
   way it rejects relative POSIX paths (add the Windows row).
2. `internal/scriptworker :: TestScriptWorkerRejectsForgedWorkerIdentity` and
   `TestScriptWorkerRejectsSubstitutedManager` — `exec: ".../curator-copy": executable file not
   found in %PATH%`: the copied/substituted manager binary is written without the `.exe`
   suffix on Windows, so it is not executable; name the fixture binaries with
   `curator-copy` + the platform executable suffix (mirror how the go-v1 worker rows copy the
   manager on Windows) and keep the negative semantics (forged/substituted identity still
   refused BEFORE the interpreter runs — the row must assert the refusal class, not the
   exec error).
Check every other new row for the same two assumptions (absolute paths, `.exe`) so the next
gate does not fail on the next one. Continue from the revision-1 tree in the Story workspace
(no checkout/clean/stash); append "Revision 2" to results.md (what changed, Windows proof
status); republish only when the gate is green. Rulings R1–R5 of 3dmjbc-brief.md unchanged.
