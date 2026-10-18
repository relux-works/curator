# BUG-261018-b2jnyg: repair refuses floats in a managed .claude.json (developer, P0)

## Problem

After Claude Code has run in a managed home, `env resolve --repair` (and every profile operation that reseeds the home) fails with `environment_repair_failed: CCJ-1 does not support float64`. Claude Code writes floating-point values into its own `.claude.json`, for example `cachedGrowthBookFeatures` probabilities and `projects[...].lastCost`, `lastFpsAverage`, `costUSD`. The home becomes unrepairable, task-board refuses its automatic repair, and a reboot does not restore the session. Reproduced on Linux with rc.7; the code path is platform-independent.

Root cause, read at current main:
- `internal/envprofile/managed.go`, `claudeSeed` (around line 804): when the file is present it decodes the whole tool-owned object with `encoding/json` into `map[string]any` (numbers become `float64`), adds the launch-directory entry, and re-encodes the **whole object** through `protocoljsonMarshal` → `protocoljson.MarshalCanonical` (CCJ-1, integers only; refusal at `internal/protocoljson/ccj.go:143`).
- The function's own contract says the opposite: "Later tool writes are its own state: repair only adds the launch directory's entry and never rewrites anything else."

## Required

1. **Read the governing spec first** (curator-spec managed homes / environments §7.4 seed rules, and wherever the `.claude.json` seed encoding is defined). If the spec requires the tool-owned file to be CCJ-1 after repair, stop the line: persist the conflict in the gaps file and hand off blocked. Otherwise CCJ-1 applies only to manager-authored state, and `.claude.json` after first run is tool-owned.
2. **Fix.** Repair of a present `.claude.json`:
   - when the launch-directory entry already carries the required flags, writes nothing (no-op; file bytes and mtime unchanged);
   - otherwise adds or updates only that entry and preserves every other value exactly: numbers keep their textual form (decode with `UseNumber` or an equivalent token-preserving approach; no float rounding, no exponent rewriting), unknown keys, nesting and value types are preserved. Key order and whitespace may change only if exact preservation is impractical; say which in the gaps file.
   - the provisioning seed for an **absent** file stays exactly as today (CCJ-1 bytes of the seed object).
   - the verification reader (`claudeProjects`, around line 889) must also accept floats anywhere in the file.
3. **Sweep.** List every other place where Curator re-encodes a tool-owned file of any harness (Claude, Codex, pi, Muse, opencode) through CCJ-1 or otherwise rewrites values it does not own. Fix those that share this failure class inside `internal/envprofile`; list the rest with file and line in the gaps file.
4. **Tests** (Go, temporary homes only, never the real HOME; one-line claim comment per test):
   - Regression witness: a `.claude.json` with the real-world float keys above (and a large integer, a negative number, an exponent form like `1e-7`, nested arrays of floats) repairs successfully; the launch entry is added; every other value is preserved byte-for-byte as a JSON value (compare decoded with `UseNumber`, and the number texts).
   - No-op: an already-satisfied file is not rewritten.
   - Neighbouring variants: float inside the launch-directory entry itself; `projects` missing; `projects` not an object (current behaviour stays, assert it).
   - Negative controls stay: non-object file, symlink, non-regular file, unreadable file refuse with the existing diagnostics and write nothing.
   - Absent file: seed bytes unchanged from today (golden).
   - Plant control: re-introducing the whole-object CCJ-1 re-encode fails the regression test for that reason.
5. **Validation.** Full local suite and the full hosted gate on the exact signed candidate. A runner infrastructure failure is a blocker to report, never a pass.

## Scope

- In scope: `internal/envprofile` (seed, repair, verification), helpers it needs, tests, a short note under `docs/` (not the root `README.md`).
- Not in scope: CCJ-1 itself, spec text, other harness files beyond the sweep list.

## Hand-off

1. Attach `BUG-261018-b2jnyg_gaps.md`: each required item with file and line, the tests that prove it, the sweep list, open spec questions.
2. Run `task-board handoff BUG-261018-b2jnyg --role developer`.
3. END YOUR TURN.
