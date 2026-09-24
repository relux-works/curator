# TASK-260916-3dmjbc brief (orchestrator, binding) — R1 script manager/worker invocation

Story STORY-260822-2h0v9j (curator-go-script-worker), fresh workspace on current trunk
(f0a92b8b). The Story lands as ONE integrate after R5, so intermediate revisions never ship
alone. Read first: `.task-board/.resources/TASK-260916-3gcc00/TASK-260916-3gcc00_reconciliation.md`
(landing via PR #84 as `.research/TASK-260916-3gcc00_reconciliation.md`), the spec at the CI
pin (`SPEC_PIN` 87a0d0060bad64ab883d007dcdf35df7485368bf in .github/workflows/ci.yml; protocol
1.0.0-rc.9): `profiles/manager.md` §3.1 "Enforced script-worker-v1 command launch" (the
8-step per-invocation order, the fixed 3/4-node process graph, the 11 mandatory controls),
§2.2.1 (the go-v1 manager-worker-v1 boundary you mirror), Protocol Core §4.1.1 and §4.3; the
vector `conformance/v1/vectors/script-host-execution-policy.json` (`interpreters` 2,
`opt_in_cases` 6, `mandatory_controls` 11, `preflight_cases` 5 …). Existing code: schema-8
surface (`internal/skillspec` parseScriptExecution), admission `internal/scriptpolicy` (`Admit`
returns `script_execution_policy_unsupported` unconditionally), refusal rows
`internal/install/scriptpolicy_test.go`, the go-v1 worker hidden-mode dispatch in
`cmd/curator/main.go` and its identity primitives (`internal/godriver` executable identity /
launcher-link / hash rows, `internal/buildrepo` manager wrapper) — REUSE the identity
primitives, do not fork them.

## Rulings

R1 Vertical slice = the invocation path, production-grade: (1) interpreter identity
resolution for the two closed identifiers (`node-v1`, `python3-v1`) from operator-trusted
configuration ONLY (find/extend the machine-config seam the spec names; never repository,
runtime root, `.agents/bin`, user PATH or manifest); (2) manager self-identity (canonical
regular file, symlink/reparse/hard-link rejection, strong identity + hash) reused from the
go-v1 worker; (3) the fixed hidden-mode re-execution of the installed manager as the script
worker with a fresh session nonce, identity recheck at the launch boundary, explicit stream
binding, private runtime area, worker-domain termination and join; (4) installed enforced
shims route through this path; (5) the worker proves its identity/hash and the interpreter
file identity before running the interpreter against the commit-keyed runtime entry with
arguments forwarded verbatim and the child's exit status returned.
R2 Admission and honesty between R1 and R3: unsupported policies keep refusing
`script_execution_policy_unsupported` (unchanged rows). For `node-v1`/`python3-v1` the
manager now enters the spec's preflight: it enumerates the 11 mandatory portable controls
with an implementation table; controls R1 does not yet implement (declaration-derived
environment/PATH, offline configuration, inventory application, closed evidence record —
R2/R3) are reported as unavailable, so install and invoke refuse
`script_execution_control_unavailable` naming the missing controls, and NO enforced script
launches uncontained. The launch path is still real production code, proven at the
process boundary (R3 below). Do not add env-flag or config switches that open the path.
R3 Evidence at production boundaries: (a) install of an enforced node-v1/python3-v1 skill
through `install.Project`/CLI → preflight refusal `script_execution_control_unavailable`
with the missing-control names, shim not installed (extend the existing refusal rows);
(b) unsupported policy → `script_execution_policy_unsupported` unchanged; (c) the hidden
worker mode invoked at the real process boundary (spawn the built `curator` in worker mode
with the manager's fixed argv/nonce protocol): identity proof, interpreter identity check,
stream binding, runtime-entry execution with verbatim args and exit status, teardown of a
descendant that outlives the interpreter; negative rows: forged worker identity (a copied or
modified manager binary), symlink/launcher-link substitution, tampered interpreter file,
wrong/replayed nonce, interpreter resolved from PATH/manifest → every one refuses before the
interpreter runs; (d) the six `opt_in_cases` keep their current classification; (e) mutants:
drop the launch-boundary identity recheck, drop the nonce check, accept PATH interpreter,
skip teardown → each killed by a named row. POSIX + Windows: the worker/process rows must
run on windows-latest (the go-v1 worker rows do); declared skips only via the ledger
vocabulary and only where a go-v1 sibling row skips for the same reason.
R4 No behaviour change for declared-only script commands, build commands, or the go-v1
worker (goldens/rows unchanged); CHANGELOG `## Unreleased` → `### Added` entry stating the
R1 scope and that enforced launch remains refused until the control set is complete;
docs/troubleshooting.md section for `script_execution_control_unavailable`.
R5 This leaf is large: hand off when R1–R4 are complete and the gate is green; if you run
out of budget, publish what is complete with results.md stating exactly which items are
done/undone (the orchestrator continues you), never a partial that claims completeness.
results.md: design (process graph as implemented, identity seam reuse, nonce protocol),
row table (production boundary per row), mutant table, ratio line (spec vector keys
consumed by this leaf), Windows proof status, bounds.
