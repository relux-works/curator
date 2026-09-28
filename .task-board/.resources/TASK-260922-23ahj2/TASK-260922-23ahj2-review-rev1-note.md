# Review note for TASK-260922-23ahj2 revision 1 (orchestrator, binding) — corrections to the landed 0017/0018 adoption

Repository curator-spec. The adoption landed as 937e7952 (PR #77, reviewed and operator-confirmed).
This leaf lands the operator's three corrections (brief TASK-260922-23ahj2-brief.md, addendum
TASK-260922-23ahj2-addendum.md, and the operator's relay 0017-0018-operator-relay-260922.md,
which is the authority). Text/AC only — no schema, vector or generator change is allowed; the gate
(`make validate`) must be green on the exact revision-1 tree (verify).

Judge against the relay, item by item:
C1 (0018 flag spelling): every attribution of the yolo provider-flag spelling / argv grammar /
per-tool-release mapping to the launcher SPEC is gone; ownership = agents-management
(`LaunchRequest` member for `LaunchModeInteractive`, interactive goldens positive AND negative;
the module's `argvguard` forbids spelling argv in two places); launcher = mode resolution
(flag > profile > global > built-in default), headless detector, transport, stderr provenance line,
launch-record key; follow-up **F-M1 (skill-agents-management)** present with a one-line AC,
**F-L1** narrowed, F-A1 unchanged; the task-board note ("0017/0018 block nothing there; consumers
are `curator run` now and tracked sessions via ax later") present; statement that 0018 neither
violates 0013 D5 nor duplicates argv grammar.
C2 (0017 Pi): the exact observation recorded (managed link
`~/.curator/environments/default/pi/auth.json → ~/.pi/auth.json`, target missing, real credential
`~/.pi/agent/auth.json`, `env status` silent); F-C1 AC: `env status`/`resolve` report a dangling
or mis-targeted link as detached; F-C3 AC: the case "recorded link whose native target does not
exist" is covered with a narrowing mutant.
C3 (0017 Codex): absent `cli_auth_credentials_store` ⇒ platform default `file` ⇒ `isolated`
admitted; stated in §7.4, manager §12.4 and 0017 choice 4; the operator's profile design relies on
the platform default and nothing more.
Headless marker set unchanged ({CI, GITHUB_ACTIONS}). Run the greps the brief names. Do not write
into the curator control root's LOGBOOK.md; record findings only in your verdict resource.
Record exactly one verdict: accept_cr(TASK-260922-23ahj2, revision=1, evidence=<your outcome
resource>) on ACCEPT, or changes_requested with file:line and the exact missing sentence.
