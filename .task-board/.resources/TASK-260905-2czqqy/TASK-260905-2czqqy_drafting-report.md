# Drafting report: launcher SPEC `0.2.1-draft` follow-ups (TASK-260905-2czqqy)

Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-agent-launcher-spec-0.2.1`, branch
`draft/spec-0.2.1`, base `e19eb9f` (SPEC 0.2.0-draft). Authority: curator-spec main `fcdb9ba`
(`protocol/environments.md` 1.1 §5.8, §7.8, §9.2 step 5, §10.1, §10.4, §12.1; `profiles/manager.md`
§12.5) and the cycle-2 review verdict `TASK-260905-3ewdq0_review-verdict-cycle2.md`.

## Status: BLOCKED on the signing passphrase — change is STAGED, NOT COMMITTED (see "Blocker")

Files changed (4): `SPEC.md` (+183/−45 across the four files; 220 changed SPEC lines), `README.md`,
`cmd/curator-run/main.go`, `cmd/curator-run/main_test.go`. `git diff --check` clean. The managed
curator-spec story worktree is untouched (empty delta by design). Nothing pushed, tagged, or PR'd.
No LOGBOOK.md or control-root write.

Patch: `TASK-260905-2czqqy_spec-0.2.1.patch` (`git diff --cached --binary` of the staged tree).

## The four brief items

| # | Item | Where | What was written |
|---|---|---|---|
| 1 | `--repair` | §4.1, §1 non-goal, §6 resolve row, README | Invocation is now `curator env resolve <env-id> [--profile <name>] --repair --format json`, always. Read-only default and fail-closed stale semantics of §10.1 stated (no fragment on `environment_home_stale`); under `--repair` the lock-free verification runs first, a current home emits without any lock, only a stale home takes the mutation lock and repairs. No read-only launch mode is offered. Diagnostic mapping: `resolve_repair_failed` kept for `environment_repair_failed`; **`resolve_lock_unavailable` added** for `environment_lock_unavailable` (environments 1.1 makes the two distinct; folding both into one code would lose that); `environment_home_stale` declared unreachable from a `--repair` call and handled as `resolve_invocation_failed` if it ever appears. |
| 2 | cycle-2 residual minors | §4.6, §6, §9 | (1) `ax.json` `enabled: false` = not configured; (2) machine-over-operator precedence explained (tracking is machine policy) against `defaults.json`'s operator-over-machine-unless-locked; (3) §6 `defaults` row now cites §4.6 and `ax.json`; (4) §9 docs-confidence item covers both files; (5) the note: the configuration read precedes argument validation, so `defaults_config_invalid` fires before a usage error. |
| 3 | codex layer stat, single `-p` | §4.5, §4.6, §6 `mcp` family | Whenever the argv carries `-p curator-mcp`, the launcher MUST stat `mcp.path` (`<home>/curator-mcp.config.toml`) immediately before handoff/exec in both modes; three outcomes: regular readable file → launch; absent → `mcp_layer_missing`; anything else → `mcp_layer_unreadable`. Neither degrades to a launch without `-p`; the launcher never writes/repairs the file. `-p` takes exactly one value (verified on 0.153.2), so an operator `-p` after `--` fails the launch at the tool; the launcher does not inspect native args to prevent it; Decision 0012 OQ3 recorded closed. §9 codex `-p` open item closed. §4.6 lists the three pre-launch checks together. |
| 4 | file family vs environments §12.1 | new §4.7, §4.3 pointer, §9 | `defaults.json` (`curator-run-defaults-v1`: model/effort defaults + lock) and `ax.json` (`curator-run-ax-v1`: integration switch) are launcher-owned knobs; no §12.1 knob is read by the launcher and none of the launcher's knobs is a §12.1 knob. Every §12.1 knob that shapes a launch reaches the launcher only through the fragment (`passable_env_names`, `system_prompt_files`, `current_profile`/`scoped_current`, `isolation`/`forms`/`in_place_mode`). Move-by-revision clause kept in §9. |

Also: normative-references bullet updated — environments 1.1 has landed at `fcdb9ba`, so the "until
that rewrite lands" hedge is gone and §7.8/§5.8/§9.2/§12.1 and manager §12.5 are cited; the §4.1
`composition` sentence no longer says "once the rewrite lands". §8 version and §8.1 row added.

## Deviations from the brief (flag for the reviewer)

- **New diagnostic codes** beyond the brief's list: `resolve_lock_unavailable`, `mcp_layer_missing`,
  `mcp_layer_unreadable`. Rationale: the brief kept `resolve_repair_failed`; environments 1.1 §10.4
  makes lock timeout a distinct condition with a different remedy (retry vs. store repair), and the
  stat rule needs a code for each of its two failure facts (absence vs. read failure are different
  facts under §6 invariant 1). If the reviewer prefers fewer codes, folding is a one-line change.

## Gates (run directly, real exit codes)

| Command | Exit | Evidence |
|---|---|---|
| `git diff --check` | 0 | inline |
| `make check` (build, vet, `go test ./... -count=1`) at the staged tree | 0 | `make-check-01.log` |
| Mutant: `specVersion = "0.2.0-draft"` in `main.go`, `go test ./... -count=1` | 1 | `mutant-specversion-01.log` — `TestSpecVersionPinned` fails: `specVersion = "0.2.0-draft", want "0.2.1-draft"`; stub restored afterwards, `make check` green again |

Mutant table:

| Mutant | Narrows the gate to | Named failing test | Survivor bound |
|---|---|---|---|
| `specVersion` reverted to `0.2.0-draft` | the stub reports the spec version the test pins | `TestSpecVersionPinned` (exit 1) | — |

Bound stated: this repository's only executable gate is the version pin. No test covers SPEC prose,
README, or the §8/§8.1 rows; those were self-reviewed by re-reading the diff. Cross-file
consistency (SPEC §3/§8/§8.1, README line 24, `main.go`, `main_test.go` all `0.2.1-draft`)
verified by grep.

## Blocker: signed commit not produced

`git commit -S` stalls: `gpg.format = ssh`, `user.signingkey = ~/.ssh/ivan_relux_signing`, the
private key is passphrase-protected, is **not** loaded in `ssh-agent` (`ssh-add -L` lists three
other keys) and `ssh-add --apple-use-keychain` prompts (no keychain entry). A headless session has no
TTY for the passphrase; `ssh-keygen -Y sign` hung until killed. The cycle-1/2 commits were
orchestrator-applied interactively for the same reason.

I did not create an unsigned commit (the brief says one signed commit; an unsigned WIP commit
invites landing unsigned). The change is left **staged** in the worktree. To produce the commit:

```bash
cd /Users/iv/Developer/ReluxWorks/.worktrees/curator-agent-launcher-spec-0.2.1
git commit -S -m "Apply the launcher 0.2.1-draft follow-ups: --repair, codex layer stat, ax.json minors, file family" \
  -m "SPEC 0.2.1-draft: §4.1 resolves with --repair and states the read-only/fail-closed semantics of environments.md 1.1 §10.1 (resolve_lock_unavailable added, resolve_repair_failed kept); §4.5 stat-before-launch rule for the codex layer file with mcp_layer_missing/unreadable and the single-value -p consequence; §4.6 ax.json enabled:false, precedence rationale, read order; new §4.7 names the defaults.json/ax.json family against environments §12.1; §6, §9, §8.1, README, and the stub's specVersion bumped."
git verify-commit HEAD
```

Or apply `TASK-260905-2czqqy_spec-0.2.1.patch` with `git apply --index` on a clean `e19eb9f`
checkout and commit the same way.

## Checklist against the DoD

- [x] SPEC 0.2.1-draft with the four items (staged)
- [x] `specVersion`, README, §8/§8.1 bumped; `make check` green (exit 0)
- [ ] one signed commit — blocked on the interactive passphrase; command above
- [x] no push
- [x] report attached

## Second run (RUN-260905-8bd5f5, 2026-09-05): re-verification and second signing attempt

The staged tree was found intact (`git status`: four files staged, no unstaged delta). Nothing in
the draft was changed by this run. Re-verified against curator-spec `fcdb9ba`:
environments.md §7.8 row 1272 (`-p` exactly one value, missing layer silently ignored under
`--strict-config`, launcher MUST stat), §10.4 rows for `environment_repair_failed` and
`environment_lock_unavailable` (line 1975 makes them distinct), §12.1 knob rows named in §4.7
(`current_profile`, `scoped_current`, `forms`, `system_prompt_files.<profile>.pi`, `isolation`,
`passable_env_names`, `in_place_mode`), manager.md §12.5 (`env resolve --repair` provisions and
re-links). All claims in the draft hold.

| Command | Exit | Evidence |
|---|---|---|
| `make check` at the staged tree | 0 | inline: build, vet, `ok cmd/curator-run 0.337s` |
| Mutant `specVersion = "0.2.0-draft"`, `go test -run TestSpecVersionPinned` | non-zero (`FAIL`) | `main_test.go:54: specVersion = "0.2.0-draft", want "0.2.1-draft"`; file restored, staged state unchanged |

Signing, second attempt, all non-interactive with a 20 s bound:

| Step | Result |
|---|---|
| `ssh-add -L` contains `SHA256:Ng99XGF2pboYgFVfWJhYI2JRi0PyYsV9UwsJ70NBYd0` (the signing key) | no |
| `ssh-keygen -Y sign -f ~/.ssh/ivan_relux_signing` with askpass forced to fail | exit 255, `incorrect passphrase supplied to decrypt private key` |
| `ssh-add --apple-load-keychain` | exit 0, loaded five other identities; the signing key has no keychain passphrase entry |
| `ssh-add --apple-use-keychain ~/.ssh/ivan_relux_signing` non-interactive | exit 1 |

Conclusion: the key can only be unlocked by a human typing the passphrase. Exact unblock, either:

```bash
ssh-add --apple-use-keychain ~/.ssh/ivan_relux_signing   # once; stores the passphrase in the keychain for later headless runs
```

then re-run this task, or run the `git commit -S` command from the Blocker section above by hand
and `git verify-commit HEAD`. After the signed commit exists, checklist item 2 can be checked and
the task handed off with `task-board handoff TASK-260905-2czqqy --role developer`.
