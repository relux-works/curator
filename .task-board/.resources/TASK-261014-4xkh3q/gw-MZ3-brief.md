# THE ONLY CURRENT INSTRUCTION — TASK-261014-4xkh3q: materialize the Muse root context in curator (W-MZ3, developer)

**Repository:** curator (Go). Spec: curator-spec main `eb1f08c` (PR #143) admits the Muse root-context target. Read `protocol/environments.md` §§7.1, 7.2, 7.9, 10.2, 12 and 13 in `/Users/iv/Developer/ReluxWorks/platform/curator-spec`.

**Do this.**
1. In `internal/envregistry`, give the `muse` adapter its root-context target, consistent with the managed XDG layout: the file `AGENTS.md` under `<home>/config/muse/`, which is `$XDG_CONFIG_HOME/muse/AGENTS.md` for the launched process. Use the forms row "monolithic only, default monolithic", and set the verified release / size-advisory fields to match the spec row (1.4.7, 32768). If the registry's `RootTarget` is a bare file name relative to the home, add the minimal field or path rule needed for a nested target; do not hard-code a muse special case in callers.
2. **Materialization:**
   - Muse managed homes now write the monolithic root-context surface, with marker and drift accounting like every other adapter.
   - `referenced` for muse is `environment_form_unsupported`, with no fallback.
   - A profile with no context writes no Muse root file; follow the spec text exactly.
3. **`env status`:** update the standing Muse note: the root context is now delivered; skills discovery stays open; the `serve` foreign-context residual and the workspace-precedence residual stay.
4. **The launch flag.** The spec requires managed Muse `exec` launches to pass `--no-foreign-personal-context`, and only when the target is materialized. If this repository composes Muse launch argv anywhere (search for muse exec argv composition, `env resolve` launch plans, or fragment consumers), add the flag there with that binding. If argv is composed only outside this repository (curator-agent-launcher, task-board), make no change here; record it in the gaps file as a hand-off and name the exact spec sentence.
5. **Tests.** Cover the monolithic output path and bytes for muse, the `referenced` refusal, drift on a changed or removed `AGENTS.md`, the status note, and (if point 4 applies) the argv flag present iff the target is materialized. Add the §13 vectors the spec lists, where this repository carries conformance vectors.

**Rules:**
- R223: compile only locally (`go build ./...`, `go vet ./...`); hosted CI is the arbiter. Use `sh scripts/remote-gate.sh` for the full suite.
- Never touch `~/.curator` or any live home. Tests use temp dirs.
- Signed commits; never pass `--no-gpg-sign`.
- Public repository: no private identifiers or local paths in code or docs.

**Hand-off:**
1. Attach `TASK-261014-4xkh3q_gaps.md`: the changes, the tests, the remote-gate result, and the point-4 hand-off.
2. Run `task-board handoff TASK-261014-4xkh3q --role developer`.
3. END YOUR TURN.
