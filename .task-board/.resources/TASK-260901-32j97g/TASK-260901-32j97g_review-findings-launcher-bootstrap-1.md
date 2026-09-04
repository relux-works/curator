# Review verdict: curator-agent-launcher bootstrap (TASK-260901-32j97g)

**Verdict: ACCEPT** — CR-TASK-260901-32j97g-1 revision 1.

## Subject

- Change Request delta (curator repo, story branch): `LOGBOOK.md` only — one
  logbook block recording the bootstrap milestone and its decisions. Base
  `979fa36e` verified ancestor of `origin/main`; candidate tree `cda3b620`
  matches commit `c50fc844` exactly.
- Real deliverable: external repo `~/Developer/ReluxWorks/curator-agent-launcher`,
  branch `main`, heads `d3400fb` (skeleton) + `dae0c35` (SPEC draft) — exactly
  the heads the review brief names. Nothing pushed; remote untouched, no
  remote configured in the local repo.

On the delta shape: the task's deliverable is a *fresh separate repository*,
which by design cannot appear inside the curator story branch. The LOGBOOK
entry is the correct in-repo trace, and I verified its every claim against the
external repo directly (commits, signatures, gate results, exit-code behavior,
decisions). The entry is accurate.

## What was verified (rerun by this reviewer, not accepted from notes)

- `go build ./...`, `go vet ./...`, `go test ./... -count=1`, `gofmt -l .`,
  and `make check` — all green in the launcher repo.
- Binary-level smoke at the production entry (compiled binary, not `go run`):
  `--version` → exit 0; `claude_code --profile companyA -- resume --last` →
  exit 2; no args → exit 2. The stub's gate (anything but the two
  informational single-flag shapes is refused) is negative-tested at the
  `run()` production dispatch site (`main` exits with its return value), with
  cases that fail if the gate admits a real env-id, a trailing-arg
  informational flag, a spec-declared-but-unimplemented flag, or bare `--`.
- Both commits signed, `git log --show-signature` reports Good signature for
  each. Working tree clean; `.gitignore` covers the local build binary and OS
  noise actually present in the directory.
- `LICENSE` is byte-equal to curator's Apache-2.0 (`cmp` clean). `NOTICE`
  mirrors curator's shape, naming "Curator Agent Launcher". `go.mod`: module
  `github.com/relux-works/curator-agent-launcher`, go 1.23. Stub imports
  stdlib only (`fmt`, `io`, `os`); agents-management deliberately not
  imported, recorded in SPEC §7 as the planned dependency.
- README status honest: specification draft, not implemented, stub behavior
  described accurately; PATH-discovery install note matches environments.md
  §11.

## SPEC.md vs the normative sources

Checked against Decision 0010 (Decisions 6 and 10), environments.md §5.5,
§7.3, §10, §11, and skill-agents-management SKILL.md (fetched via gh api):

- **Boundary honored**: no session state (fire vs manage), plans consumed as
  values and never rebuilt, Curator and ax as CLI contracts only,
  `github.com/relux-works/skill-agents-management` as the single planned
  module edge (path matches the SKILL's own go.mod line). SPEC §7's relied-on
  invariants (fail-open limit state, indeterminate-read-is-unknown, per-model
  effort with no injected default, argv parity, frozen digests) each verify
  against the SKILL's invariants section verbatim.
- **Composition algorithm**: five ordered steps, no partial launch, no
  degradation. Fragment obtained via `curator env resolve --format json`
  subprocess; closed-object parse per §10.2; resolve failure modes map
  Curator's §10.4 diagnostics 1:1 and malformed output is
  `resolve_fragment_invalid`, never treated as absence. Env merge process <
  plan < fragment with fragment winning on exactly its closed
  adapter-registry names — safe by the §10.3 profile-influence boundary,
  which the SPEC cites correctly.
- **ax handoff**: ALWAYS-when-configured, no `--no-ax` flag, bypass is a
  configuration change; `ax_handoff_failed` with no untracked fallback;
  recorded fragment data (profile name, effective commit or `state_sha256`
  for local, fragment digest) matches Decision 10's Session Record
  extensions recommendation, including the local-profile pin shape from
  §10.2.
- **System prompt**: opt-in only (`--system-prompt <append|replace>` with a
  required value — a justified tightening of the brief's bare flag: the
  launcher never defaults to the destructive replace channel), opting in to
  nothing is `sysprompt_channel_unavailable` rather than a silent no-op, and
  all three mandatory warning elements are present (customized naming
  profile+semantics; replace discards built-ins; cache/billing consequence
  with the open-question-7 pointer). Channel tables referenced to §7.3, not
  restated.
- **Diagnostics**: closed families, one code line to stderr, usage exit 2 /
  operational exit 1, and the two cross-family invariants
  (absence ≠ read failure; no diagnostic downgrades a launch) are exactly the
  negative-evidence discipline.
- **Versioning**: 0.1.0-draft, promotion to sibling
  `curator-agent-launcher-spec` at stabilization per Decision 6, `-draft`
  as the nothing-may-pin signal.
- **Flag surface**: minimal — each flag names the single plane it feeds;
  post-`--` argv verbatim and uninspected; unrecognized pre-`--` flags are
  usage errors, never forwarded. Nothing duplicates spawn-plane or Curator
  concerns. English throughout; prose matches the curator-spec house style.

## Findings

1. **Minor (non-blocking, for the next SPEC revision): `file`-kind channel
   descriptors are unaddressed and make the §5 selection rule ambiguous for
   `pi`.** environments.md §7.3 declares four channels for `pi` (flag/append,
   flag/replace, file/append `APPEND_SYSTEM.md`, file/replace `SYSTEM.md`),
   and fragment-v1 §10.2 names `file` as a known descriptor kind the
   launcher's closed parse must accept. SPEC §5 applies by kind for flag,
   config-key, and variable only, and selects "the fragment channel whose
   semantics equals the opt-in value" — for `pi` that matches two channels
   per semantics value. Per §5.5 the file channels are Curator
   machine-configuration territory applied unconditionally by the tool, so
   the resolution is one sentence: selection considers only descriptors of
   the kinds the launcher applies; `file`-kind descriptors are accepted as
   data and never selected or applied by the launcher. A `0.1.0-draft` spec
   explicitly permits incompatible revision and carries an Open items
   section; this belongs there or in §5 directly, and does not block
   acceptance of the bootstrap.
2. **Note**: the LOGBOOK observation that `go run` masks the child exit code
   as 1 is confirmed by design (go run reports a non-zero child as its own
   exit status 1); the SPEC's exit-2 contract was verified here against the
   compiled binary.

## Disposition

Accepted via `accept_cr(TASK-260901-32j97g, revision=1)`. Not pushed, not
marked done — the orchestrator lands the external repo and makes the `done`
transition with `commit_ack=scope_committed`. Finding 1 is handed to the next
SPEC revision, not returned for rework.
