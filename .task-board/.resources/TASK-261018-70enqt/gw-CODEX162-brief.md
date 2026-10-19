# TASK-261018-70enqt: qualify Codex 0.162.1 for the codex_cli adapter (developer)

## Why

Hosts now run Codex 0.162.1. `internal/envregistry/envregistry.go:285` still records `VerifiedRelease: "0.153.2"`, and every resolve warns that 0.162.1 is detected while 0.153.2 is recorded. The owner asked to refresh the qualification.

## Do this

Follow the pattern of the Muse 1.4.8 qualification (commit 007033f1, `.research/muse148/`): a research record with probes, raw outputs and an opt-in replay.

1. **Re-verify, on Codex 0.162.1 in temporary homes only, every fact the codex_cli registry row and the managed-home code rely on:**
   - `CODEX_HOME` selects the home;
   - the root context is loaded from `<home>/AGENTS.md`, and the size advisory still applies;
   - skills are discovered under `<home>/skills/`;
   - the `config.toml` seed is accepted;
   - auth is inherited through the linked `auth.json` (credential scope `home`, write in place);
   - exec still needs `--skip-git-repo-check` outside git;
   - the `-p` profile flag behaves as recorded.
2. **Record two behaviours found in the 2026-10-18 Werkstatt pilot**, as facts with evidence, without changing product behaviour:
   - the per-directory trust prompt for project hooks ("Trust all and continue");
   - the workspace-write sandbox protecting every `.git` under writable roots, and how `--separate-git-dir` with `GIT_DIR`/`GIT_WORK_TREE` avoids it.

   Reference these in the record for CIP-0020 and the Werkstatt layout.
3. **If every fact holds,** bump `VerifiedRelease` to `0.162.1`, with a test that pins it. If a fact changed, stop the line with the evidence and the exact decision needed. Do not adapt behaviour silently.
4. **Opt-in replay:** an env-gated test that re-runs the probes on a host with Codex installed. It is skipped by default and never touches the real HOME.

## Scope

- In scope: `.research/codex162/`, `internal/envregistry` (the verified release and its test), and a short CHANGELOG line under Unreleased.
- Not in scope: the root `README.md`, spec text (note in the gaps file anything the spec should record), and the real HOME.

## Hand-off

1. Attach `TASK-261018-70enqt_gaps.md`: each fact with its evidence file, and the two pilot behaviours.
2. Run `task-board handoff TASK-261018-70enqt --role developer`.
3. END YOUR TURN.
