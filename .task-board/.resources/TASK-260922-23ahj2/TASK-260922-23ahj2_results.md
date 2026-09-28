# TASK-260922-23ahj2 results — operator corrections to the landed 0017/0018 adoption

Follow-up revision to the landed adoption (curator-spec `937e7952`, PR #77).
Text/AC only; no schema, vector, or generator change. Changed files (exactly 6):

- `decisions/0018-curator-run-permission-interface.md`
- `decisions/0017-environment-credential-modes.md`
- `protocol/environments.md` (§7.4, §10.1)
- `profiles/manager.md` (§12.4)
- `cli/curator.md`
- `CHANGELOG.md`

Revision 2 (2026-09-22) applies rework-1: three verbatim sentences from the
binding addendum inserted where the rev1 verdict names the lines, on the
revision-1 tree with no other change. See "Revision 2" at the end.

## C1 — provider flag spelling owned by agents-management (0018)

The mapping `yolo → --dangerously-skip-permissions` (Claude Code) /
`--dangerously-bypass-approvals-and-sandbox` (Codex) now lives in
agents-management (repository skill-agents-management) as a `LaunchRequest`
permission-mode member for `LaunchModeInteractive`, with goldens per tool
release. The launcher only resolves and passes the mode.

agents-management owns the LaunchRequest permission-mode member for
LaunchModeInteractive with positive and negative interactive goldens per tool
release; the module's argvguard forbids spelling argv grammar in two places,
which is why F-M1 owns the mapping and F-L1 carries no argv grammar.
0017/0018 block nothing in task-board: tracked children spell bypass in the
module's exec plugins themselves; 0018 serves curator run now and tracked
sessions later through ax (F-A1).

Fixed every attribution to the launcher SPEC:

- 0018 item 1: the "one reviewed provider-mapping owner" is named as
  agents-management (`LaunchRequest` member + goldens).
- 0018 choice 3: the parsing rule (prompt text vs flags) is owned by
  agents-management as part of its argv grammar (was: launcher SPEC).
- 0018 choice 4: unchanged ownership — the tracked `ax` launch document
  extension key stays launcher-SPEC-owned (genuine launcher ownership);
  clarified that the follow-up fixes the spelling of the launcher's own
  line and key.
- 0018 choice 5: launcher suite drives the resolved mode through the real
  `curator run` entry; per-tool-release mapping goldens live in
  agents-management per choice 6.
- 0018 choice 6: capability table keyed by (environment, tool release) +
  token/member encoding owned by agents-management (follow-up F-M1);
  drift ⇒ `yolo` refused first (unchanged rule).
- 0018 choice 7: unchanged — the `{CI, GITHUB_ACTIONS}` marker list stays
  versioned in launcher SPEC §4.6 (the headless detector is the
  launcher's).
- 0018 Compatibility: launcher follow-up narrowed to F-L1 (mode
  resolution, transport of the resolved mode, provenance only — no argv
  grammar; §4.2 mapping table and `internal/mapping` removed from the
  launcher list); spelling/mapping/grammar/capability-table owned by
  agents-management (F-M1) with positive and negative interactive goldens
  per tool release — the module's argvguard forbids spelling argv grammar
  in two places, which is why F-M1 owns the mapping and F-L1 carries no
  argv grammar; explicit statement that 0018 neither violates
  Decision 0013 D5 nor duplicates argv grammar; task-board note
  (0017/0018 block nothing there; consumers are `curator run` now and
  tracked sessions via ax later).
- environments §10.1: `yolo` requests the agents-management-declared
  native bypass; ownership paragraph rewritten (agents-management:
  spelling/mapping/goldens/grammar; launcher: mode resolution incl.
  conflict refusal, headless detector, transport, stderr provenance
  line, launch-record extension key). §10.2 needed no change (transport
  only, no spelling echo).
- `cli/curator.md`: launcher resolves and passes the mode;
  agents-management maps it; ownership sentence fixed.

AC grep `grep -n "launcher SPEC owns\|launcher-SPEC-owned"
decisions/0018*.md protocol/environments.md cli/curator.md` returns only
0018 choice 4 (launcher-SPEC-owned extension key — genuine).

## C2 — Pi dangling-link evidence (0017, §7.4, F-C1, F-C3)

Recorded as operator evidence 2026-09-21 in 0017 Context, 0017 choice 3,
and environments §7.4 (Migration): on the operator Mac the managed Pi
home links to `~/.pi/auth.json`, which does not exist — the real
credential is `~/.pi/agent/auth.json` — the link is dangling and `env
status` is silent. Exact observation: On the operator Mac,
~/.curator/environments/default/pi/auth.json points to ~/.pi/auth.json;
that target does not exist, the real credential is ~/.pi/agent/auth.json,
and env status does not show the dangling passthrough as detached.
F-C1 AC extended: `env status` (and `resolve`) report
a dangling or mis-targeted credential link with
`environment_credential_conflict`-class wording instead of silence; env
status and resolve report dangling or mis-targeted credential links as
detached, with environment_credential_conflict-class wording instead of
silence. F-C2 kept as is. F-C3 AC extended: The production-entry suite
covers a recorded link whose native target does not exist, with a
narrowing mutant proving this bound.

## C3 — absent `cli_auth_credentials_store` means effective `file` (0017, §7.4, §12.4, F-C1)

Stated in environments §7.4 (codex passthrough row + `isolated`
paragraph), manager §12.4, and 0017 choice 4: an absent key resolves to
effective `file` storage, so `isolated` is admitted for `codex_cli`;
`keyring`/`auto` ⇒ `environment_isolated_unsupported` (unchanged).
Operator evidence: The operator's machines have no
cli_auth_credentials_store key in config.toml, so the platform default
file applies and isolated is admitted; the operator's profile design
relies on that platform default and on nothing more from 0017.
F-C1 AC extended ("absent key resolves to `file`").

Headless marker set unchanged: {`CI`, `GITHUB_ACTIONS`} (operator-confirmed;
TTY primary).

## Amended follow-up table

curator:

- F-C1 envprofile credential-link repairs — implement fix-first repairs
  per §7.4/§10.1 (unlink only a recorded symlink still targeting the
  declared store; `environment_credential_conflict` on regular file /
  unexpected target; never remove bytes). AC: shared→isolated leaves no
  stale link and regular-file-at-link refuses with the conflict
  diagnostic naming the path; `env status` (and `resolve`) report a
  dangling or mis-targeted credential link with
  `environment_credential_conflict`-class wording instead of silence;
  env status and resolve report dangling or mis-targeted credential
  links as detached, with environment_credential_conflict-class wording
  instead of silence; an absent `cli_auth_credentials_store` key
  resolves to `file`.
- F-C2 explicit credential migration step — unchanged (inspect → plan →
  apply under the manager lock; Pi wrong-target home migrates to
  `~/.pi/agent` with bytes preserved and mode intact, printing the plan
  before applying).
- F-C3 production-entry tests on temporary stores — cover both
  0017 hazards, the migration, and every repair refusal, each with
  a narrowing mutant proving the bound. AC: `go test`
  ./internal/envprofile/... passes with a mutant-killed row per
  refusal (conflict admitted ⇒ test fails). The production-entry suite
  covers a recorded link whose native target does not exist, with a
  narrowing mutant proving this bound.

curator-agent-launcher:

- F-L1 permission interface, narrowed (SPEC + implementation + tests) —
  SPEC §§3/4.1/4.3–4.7/6 (no §4.2 mapping table, no argv grammar),
  `curator-run-defaults-v2`, internal/{cli,composition,execution},
  choice-5 negative rows with narrowing mutants, Q4 line + launch
  record; mode resolution (flag>profile>global>built-in default),
  transport of the resolved mode, and provenance only. AC: `curator
  run` resolves flag>profile>global>built-in default, refuses per the
  tables, and the fake-tool suite passes with every refusal
  mutation-killed.

skill-agents-management (new):

- F-M1 `LaunchRequest` permission-mode member for
  `LaunchModeInteractive` — provider flag mapping + goldens per tool
  release, drift ⇒ `yolo` refused first. AC: `yolo` maps to the pinned
  native bypass per (environment, tool release) with positive and
  negative interactive goldens per tool release, and an unverified
  release refuses `yolo` while `native` forwards verbatim; the module's
  argvguard forbids spelling argv grammar in two places, which is why
  F-M1 owns the mapping and F-L1 carries no argv grammar.

curator-spec (follow-up spec revisions): F-S1, F-S2, F-S3 unchanged.

agent-session-manager-spec: F-A1 unchanged (tracked yolo admitted only
through the versioned capability with D5/D3.6 satisfied). 0017/0018
block nothing in task-board: tracked children spell bypass in the
module's exec plugins themselves; 0018 serves curator run now and
tracked sessions later through ax (F-A1).

Operator-run experiments: E1, E2 unchanged.

## Validation evidence (own runs, this worktree)

Revision-2 tree, gate interpreter
`curator-spec/.temp/venv/bin/python3` (Python 3.14.6, jsonschema ok;
ambient `python3` lacks jsonschema), zsh:

- `tools/validate.py` → `validated 62 schemas and 1124 vector files`,
  exit 0 (~13 s).
- `go test ./tools/...` → ok, exit 0; `gofmt -l tools/` empty;
  `go vet ./tools/...` clean.
- unittest in bounded chunks, 579/579 OK:
  - chunk A (test_implementation_coverage, test_release_gate,
    test_verify_release_commit, test_verify_release_merge_policy):
    78 OK (~53 s).
  - chunk B1 (WireSemantic … WorkflowRegenerationScope): 56 OK (~22 s).
  - chunk B2a (EnvironmentVector … PathKindAdmissionVector): 115 OK
    (~43 s).
  - chunk B2b (SourceSignersVector … ContextDetectorVector): 103 OK
    (~94 s).
  - chunk B3a (SnapshotAcquisition, ShellHookTrust, SecurityPosture,
    UmbrellaProvider): 55 OK (~1 s).
  - chunk B3b (ManagerConfig, SystemConfigV2, RegistryPageBoundary,
    RegistryCheckpoint, RegistryBootstrap): 122 OK (~118 s).
  - chunk C (WriteNofollowVector, DotfileManagersVector,
    ReadFailureVector): 50 OK (~13.3 min wall under load 11; CPU-bound,
    not hung — verified via ps during the run).
  - Total 78+56+115+103+55+122+50 = 579/579, equal to rev1.
- Post-run `git status` path set is exactly the 6 prose files (transient
  per-case corpus rewrites restored byte-identical).
- No schema, vector, or generator file changed: `git status` shows only
  the 6 prose files.
- `git diff --check` clean.
- AC greps: launcher-ownership grep returns only 0018:319 (genuine
  launcher-owned launch-record extension key); argvguard sentence 1× in
  0018; task-board note 1× in 0018; exact Pi path 2× in 0017 + 1× in
  §7.4; detached wording 2× in 0017 + 1× in §7.4; C3 operator sentence
  1× each in 0017 choice 4, §7.4, manager §12.4.

## Notes and anomalies

- Revision-1 note (kept): the rev1 producer ran under severe memory
  pressure (~69 MB free), which SIGKILLed `uv`, the `~/.local`
  python3.14, and `pip install` attempts; rev1 gates ran via a
  project-local `.venv` (removed afterwards). No SIGKILL observed in
  the revision-2 session; the control-root gate venv
  (`curator-spec/.temp/venv`) served all python gates.
- The `uv run` SIGKILL symptom matches the brief's note about the
  configured gate retrying a SIGKILLed run.
- This leaf is text/AC only: no runtime behavior or new mutation
  coverage claimed. Missing-target and interactive-golden runtime
  proofs belong to the follow-up leaves (F-C3, F-M1).

## Revision 2 (2026-09-22) — rework-1 verbatim insertions

Continued from the revision-1 tree (no checkout/clean/stash); inserted
the three binding sentences verbatim; no other change. Line numbers on
the revision-2 tree:

1. C1 — `decisions/0018-curator-run-permission-interface.md:353`
   (Compatibility): "agents-management owns the LaunchRequest
   permission-mode member for LaunchModeInteractive with positive and
   negative interactive goldens per tool release; the module's argvguard
   forbids spelling argv grammar in two places, which is why F-M1 owns
   the mapping and F-L1 carries no argv grammar." Mirrored in the F-M1
   AC above. `decisions/0018-curator-run-permission-interface.md:359`
   (Compatibility): "0017/0018 block nothing in task-board: tracked
   children spell bypass in the module's exec plugins themselves; 0018
   serves curator run now and tracked sessions later through ax (F-A1)."
   Mirrored in the F-A1 row above.
2. C2 — `decisions/0017-environment-credential-modes.md:90` (Context
   rationale) and `:166` (choice 3),
   `protocol/environments.md:1493` (§7.4 Migration): "On the operator
   Mac, ~/.curator/environments/default/pi/auth.json points to
   ~/.pi/auth.json; that target does not exist, the real credential is
   ~/.pi/agent/auth.json, and env status does not show the dangling
   passthrough as detached." plus "env status and resolve report
   dangling or mis-targeted credential links as detached, with
   environment_credential_conflict-class wording instead of silence."
   F-C1 extended with the detached wording; F-C3 extended with "The
   production-entry suite covers a recorded link whose native target
   does not exist, with a narrowing mutant proving this bound." F-C2
   unchanged (bytes-preserved migration).
3. C3 — `decisions/0017-environment-credential-modes.md:167` (choice 4),
   `protocol/environments.md:1473` (§7.4), `profiles/manager.md:2748`
   (§12.4): "The operator's machines have no
   cli_auth_credentials_store key in config.toml, so the platform
   default file applies and isolated is admitted; the operator's
   profile design relies on that platform default and on nothing more
   from 0017." Keyring/auto refusal kept.
