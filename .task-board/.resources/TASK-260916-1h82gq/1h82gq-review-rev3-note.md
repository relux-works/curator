# Review note for TASK-260916-1h82gq revision 3 (orchestrator, binding) — R2 script worker

Brief 1h82gq-brief.md (scope: manager-built environment incl. the reserved-name set, manager-built
PATH, offline network configuration + proxy/resolver scrub, operation-private runtime area APPLIED,
declared-secrets-as-identifiers, deny-by-default derivation; plus R-A native launcher for enforced
shims, R-C permit frame, R-G guard; R1 honesty: inventory-controls-applied and the closed evidence
record stay "unavailable" until R3 so production still refuses `script_execution_control_unavailable`;
table-injection seam production-unsettable). The producer hit the 150-min timeout once and was
continued; revisions 1–2 failed only on Windows (NTFS link-count false positive, mode assert, `.cmd`
duplicate staged target, `SYSTEMROOT` injected by os/exec); revision 3 is green on all lanes: run
35638584352 — verify the gate commit resolves to the exact revision-3 tree. Read results.md's
done/undone list first: an honest partial is allowed by the brief, but every "done" claim must be
proven and every "undone" item must be listed.

Judge with your own reruns (disposable clone; build the binary; bounded commands; retry once on
host stalls; do NOT write into the curator control root's LOGBOOK.md — record findings only in
your verdict resource):
1. Spec §3.1 reserved names: the manager-built environment withholds EVERY reserved name (all
   platforms; macOS `DYLD_*`; Windows set incl. `SYSTEMROOT`/`WINDIR` which must be manager-set,
   case-insensitive; `python3-v1` `PYTHON*`/`__PYVENV_LAUNCHER__`; `node-v1` `NODE_*`/`NPM_CONFIG_*`)
   even when `env_read` names it; an interpreter identifier without a reserved set is refused;
   mutants per control (pass one reserved name; keep inherited PATH; skip the proxy scrub with
   network=none; skip runtime-area binding; …) — rerun at least four.
2. Four derivation cases at the production entry; PATH = interpreter + resolved `exec` only,
   `exec` manager-resolved (never from package data); secrets identifiers only; declared hosts
   reporting-only (no host-filtering claim).
3. Launcher (R-A): installed enforced shims route through a NATIVE launcher (no shell/`.cmd`/symlink
   between manager and interpreter; Windows naming vs `.cmd` shims — the rev1 duplicate-target
   bug); permit frame (R-C): the parent validates the worker's proof BEFORE the interpreter runs
   (probe: withhold permit → interpreter never starts); R-G guard replaced.
4. Admission honesty: production entries still refuse `script_execution_control_unavailable`; the
   table-injection seam cannot be set from production (row + your own attempt via CLI/config/env).
5. Windows rows really ran on windows-latest (extract from the gate's Windows evidence); NTFS
   link-count fix is a real cause fix, not a relaxed check; `internal/snapshot` flake handling as
   results.md states.
6. R-B stream model, R-F, and the undone list: state precisely what R3 must pick up.
Record exactly one verdict: accept_cr(TASK-260916-1h82gq, revision=3, evidence=<your outcome
resource>) on ACCEPT, or changes_requested with file:line and reproduction.
