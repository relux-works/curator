# TASK-260827-2232c0 review verdict: CHANGES REQUESTED

Reviewer run against `CR-TASK-260827-2232c0-1` revision 1.
Delta reviewed: `git diff 903af23ad0d0fa21328c0a2100e17968bbac6f1e d2e497aa84f3a645141c762c021a365d8c22d4fe`
(5 paths, +280 / -345). `docs/prose-style.md` in that delta belongs to sibling
task `TASK-260827-2gmk4c` and is not judged here except where it contradicts
this repository.

## What passes

- README is 116 lines against a 220-line ceiling.
- `<details>` install blocks are present, one per platform, Homebrew `<details open>`.
- Both moved sections leave a one-paragraph summary plus a link to
  `docs/compiled-commands.md` at their former positions.
- `docs/implementation-plan.md` gains exactly the two-line historical header;
  its content is untouched.
- Style blacklist is clean in both files: no em-dashes or en-dashes, no
  guillemets, no antithesis constructions, no marketing register, no filler
  openers, no closing restatement paragraphs.
- **Every command and flag verifies against the tree binary.** Run from this
  worktree with `go run ./cmd/curator ... --help`: `status` (`--check`,
  `--json`), `global status` (`--check`, `--json`), `install`/`upgrade`
  (`--dry-run`), `global install --dry-run`, `bootstrap` (`--if-missing`,
  `--non-interactive`, `--skills-root`), `shell-init --install`, `gc`. No
  invented command or flag anywhere in the delta. The tree compiles (`go run`
  resolved and executed); the change is documentation-only and touches no Go
  source, so no suite was rerun.

## Blocking: the moved sections do not preserve every fact

AC: "compiled-commands.md preserves every fact and command from the moved
sections." The prose was restructured well, but ten load-bearing facts were
dropped in the rewrite, and one was distorted. Two are verifiable regressions
against the Go source, not just doc-versus-doc omissions.

### B1. The `cause` table is detached from `build-input-drift` (distortion)

`docs/compiled-commands.md:45` introduces the table as "A row may carry a
`cause` subcode to refine status details:" and then lists `build-root`,
`target`, `unattributed`. Those three are **`build-input-drift` causes only**.
The base README said so explicitly, and so does the source:

- `cmd/curator/builds.go:278-281`: "A build-input-drift row carries one of
  `inputCauses()`; an unusable-build-toolchain row carries the go-v1 boundary
  code that refused the operation. Every other state leaves it empty."
- `cmd/curator/builds.go:116-121`: `inputCauses()` returns exactly those three.

Two defects in one edit: the table now reads as the general `cause` vocabulary,
and the second sentence of the base ("`unusable-build-toolchain` carries the
`go-v1` boundary code that refused the operation") is gone entirely. A reader
modelling `--json` output from this document gets it wrong. This also fails the
DoD item "no discrepancies between code and description."

### B2. The trusted-toolchain allowlist is gone

Base: "It never searches `PATH` and never downloads a toolchain, **and it
accepts only release families it has tested against the `go-v1` vectors. A
missing, untrusted, or untested toolchain reports the failing boundary together
with those mechanisms and the tested families.**"

`docs/compiled-commands.md` keeps only the two negatives. The allowlist is real
behaviour: `internal/godriver/session.go:79` (`TestedFamilies`),
`session.go:256-257` (`unsupported_go_family`, "Go release family %s is not
allowlisted"). A user whose Go release family is refused now has no document
that explains why, and `unusable-build-toolchain` in the code table is the state
they will see.

### B3. `global status --check` fails closed twice over; one condition dropped

Base: "`--check` is the only surface that turns a verdict into a non-zero exit,
and it fails closed twice over: once for every non-current code, **and once when
the plan refused before it could describe every compiled command, because such a
run cannot prove the scope is current.**"

New: "The `--check` flag turns non-current verdicts into non-zero exit codes."
The second exit condition is a distinct, observable behaviour and is now
undocumented.

### B4. Consumer-registry rules dropped from the maintenance section

Base: "A consumer registry that exists but does not match the exact shape
Curator writes is reported and left untouched rather than rewritten; a
registered checkout is unregistered only once its scope is proven absent or
proven valid and empty."

Neither rule survives. The rewritten paragraph keeps only the ambiguous-registry
case.

### B5. "a receipt alone is never treated as proof of provenance or of a live consumer"

Dropped. The new sentence keeps "Entry content is never executed, adopted, or
permission-repaired" and stops there.

### B6. The parent-object retention rule and its security consequence

Base: "an entry whose parent is no longer that object is retained and reported.
Exchanging the cache-root path after validation can therefore neither redirect a
removal outside the Curator cache root nor let a planted replacement supply the
verdict for the entry that is actually being removed."

Both sentences are gone. What remains states the binding but not the rule it
enforces or the attack it defeats, which is the whole point of that paragraph.
The enumerated evidence ("its exact members, its receipt, its artifact bytes and
size") is also gone.

### B7. "never adopted by changing permissions or rewriting a marker"

Base states how an unusable entry is *not* recovered. New: "Unusable entries are
quarantined and replaced under the manager-home lock." The prohibition is gone.

### B8. The changed-cache warning contract

Base: "it says so per command instead of repeating the ordinary 'the live build
cache is unchanged' claim, **and the warning names which of the three it was. The
installation and its consumers are unchanged either way; nothing on these paths
is ever deleted, so the state left behind is always one a later `install` or
`upgrade` repairs.**"

Everything bolded is gone. The recovery instruction for the operator ("a later
install or upgrade repairs it") is the actionable half of that paragraph.

### B9. The `gc` lock guarantee and the pass ordering

Base: acquires the lock, recovers any incomplete transaction, "**and only then**
marks and sweeps, **so it cannot race an install, a rollback, or a recovery, and
cannot lose a consumer registry update.**" New drops the ordering and the
guarantee.

### B10. The publication-reversal sync chain

Base: "That return is synced too, because it is a move like any other; only if
its own sync also fails does the run report the cache as changed, with the entry
still live and readable. So a failed publication leaves the cache exactly as it
found it, and the reversal above is only ever needed for a publication that
fully succeeded."

Dropped. Without it the preceding paragraph no longer says what state a failed
publication leaves behind.

Also minor, same class: `install --dry-run` / `upgrade --dry-run` "a plan, never
a completed compiler check"; `--json` on a closure without compiled commands
producing "the historical document unchanged"; "and the run says so" after a
failed gate.

## Blocking: a 62-line CI reference was deleted with no destination

`README.md` "### Gates and tooling" (base lines 339-401) is gone. That is the
ten-row gate table (`test-gate.sh`, `suite-plan.sh`, `platform-case-gate.sh`,
`ledger-consistency.sh`, `excluded-packages.sh`, `candidate-suite.sh`,
`toolchain-identity.sh`, `no-broad-suppression.sh`, `gate-selftest.sh`, the
race variant) with what each gates, its `make` target, and its `EVIDENCE` path;
the `CURATOR_CONFORMANCE_ROOT` requirement for `make ci-test` / `make race` /
`make check-ci`; the `SPEC_PIN` and candidate-suite rules; and the whole
"#### The compiled-build platform carve-out" subsection.

The task authorised a short Development section linking `CONTRIBUTING.md`, but
`CONTRIBUTING.md` does not carry this material: `grep -n -i
'gate\|EVIDENCE\|make ci-test\|platform-cases\|\.github/ci' CONTRIBUTING.md`
returns nothing. `grep -rn "test-gate.sh\|platform-case-gate\|ledger-consistency\|no-broad-suppression" --include="*.md"`
across the repository finds it only in `LOGBOOK.md` narrative entries. So this
delta removes the only human-facing documentation of the gate contract without
moving it anywhere.

Decide the destination and state it: fold it into `CONTRIBUTING.md`, or open a
`docs/ci-gates.md` and leave the one-paragraph-plus-link pattern in README the
same way the compiled-command sections got. Deleting it is the one option that
should not stand.

## Blocking: a shim claim became inaccurate

`README.md`, "Commands without profile setup": "Global installation places shims
in user `PATH` directories."

Base: "Global installation publishes non-destructive forwarding shims to a safe
user directory already on `PATH` **when one is available; otherwise Curator
reports the canonical global bin location.**" The fallback is real:
`internal/globalbins/globalbins.go:113` ("Select chooses a safe writable user
directory that is already on PATH"), `globalbins.go:458` ("command shims were
installed in %s, but no safe PATH-visible user bin was found; set %s to a
writable PATH directory or use curator shell-init --install"). The flat claim
tells a user on such a host that their shims are on `PATH` when they are not.

## Blocking: the LOGBOOK entry records a false preservation claim

`LOGBOOK.md`, the new `TASK-260827-2232c0` block, states the two sections were
"moved out wholesale to `docs/compiled-commands.md` **with every fact and
command preserved**". B1-B10 above show that is not what happened. LOGBOOK is
institutional memory; correct the claim to what the delta actually did before
this lands.

## Non-blocking, for the producer and the style-sweep task

1. "Registry client guarantees" in README was cut from ~18 lines to 5 and lost:
   `state/registry` sitting outside the disposable `cache/registry`; upgrades
   migrating legacy state without lowering it; corruption and write failure
   being fail-closed; the protected catalog distinguishing first use from
   deletion; pagination rejecting repeated or oversized cursors, more than
   10,000 records per artifact query, and responses larger than 16 MiB; retry
   covering `429` and `503` as well as network failures; publication retrying
   only the exact body under its deterministic `Idempotency-Key`; other client
   errors and unsafe requests never being retried; the bearer-token reason for
   rejecting redirects. This section was not named in the task, and no planned
   document receives it. Raise it with the orchestrator rather than absorbing
   the loss silently.
2. The `Spec §N.M` citation convention is dropped from README. It is used
   across the repository and now goes unexplained.
3. The shell-init paragraph drops the Windows and PowerShell path entirely
   (base: zsh or bash from `SHELL`, Git Bash preserved on Windows, PowerShell
   otherwise) and drops "The cached hook does not start Curator during later
   shell launches." The remaining text names only `.zshrc` and `.bashrc`.
4. "What Curator manages" drops `agent-skill.json` schemas 1 through 5, the
   readable legacy `csk-skill.json` filename, activation modes, "deny-wins
   federation, snapshot verification", and "never selectable by a package" for
   operator credentials.
5. The operator-credential bullet now says "SSH and HTTPS credentials" but links
   only `docs/build-ssh.md`. `docs/build-https.md` exists in this worktree
   (sibling `TASK-260827-19aqkr`). Fine to leave to the style-sweep task, but it
   should not be forgotten.
6. The Status section drops lint and the naming gate from the CI description,
   and the Install section drops the link to the releases page.

## Cross-task, not this task's defect

`docs/prose-style.md:19` uses "`curator repair` restores broken links" as a
worked example. There is no `curator repair` command: it is absent from
`curator --help`, and the README this task rewrote states "There is no separate
repair command: `install` and `upgrade` are the reconciliation path." The style
guide teaching a nonexistent command is worth fixing while
`TASK-260827-2gmk4c` is already at `to-dev`.

## Verdict

Changes requested. Routing `TASK-260827-2232c0` to `to-dev`.

The restructuring itself is sound and the command surface is verified clean.
The rework is: restore B1-B10 into `docs/compiled-commands.md` (B1 and B2 first,
they are code-contradicting), decide a destination for the gates reference,
restore the shim fallback sentence, and correct the LOGBOOK claim.
