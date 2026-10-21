# THE ONLY CURRENT INSTRUCTION — TASK-261021-mw4laf: Muse 1.4.11 foreign-context exclusion qualification (developer)

## Goal

Produce a vendor observation record for **Muse Code 1.4.11 (1.4.11-R6667.1)**: the foreign-context exclusion replay P1–P4 (environments §7.1), the loader rows, and the §7.9 release-gate status, **per form**: `exec` and `interactive`. Each form ends as either qualified, or refused with evidence. Consumer: the agents-management lane adds per-form qualified rows from this record, and task-board main (0fa34b3c9) stays fail-closed on Muse versions other than 1.4.7 until it exists. This is research qualification: no runtime constants, registry pins, qualification gates or credential policy change in this leaf.

## Template: follow it, do not reinvent

Mirror the 1.4.8 record and harness exactly:
- `docs/qualification/muse-1.4.8-R6256.1.md`, the record format and section order;
- `.research/muse148/`, the opt-in harness, fixtures and README.

Create `.research/muse1411/` and `docs/qualification/muse-1.4.11-R6667.1.md` in the same shape. Reuse the 1.4.8 fixtures where they are unchanged; regenerate any digest that depends on the binary or on the current main. Re-read the normative sources yourself on current curator-spec main: environments §§7.1 P1–P4, 7.9, 10.2; CIP-0017; `conformance/environments-muse-v4/README.md`. Record the spec main OID you read.

## Hard rules (unchanged from 1.4.8)

- **Binary:** the exact vendor executable from the task-board Muse pin store (see the round note at the end), never the copy in `~/.local/bin`, which the auto-updater replaces, passed through an opt-in environment gate `MUSE1411_BINARY`. Never the updating launcher `~/.local/bin/muse`. Record its SHA-256; the harness refuses a different digest. Without the gate, every vendor test skips before creating a fixture. Record the absolute path only in an ignored task-local `identity.json`: the board and the repo are public, so the record uses `$MUSE1411_BINARY`.
- **Isolation:** every invocation runs with `MUSE_NO_AUTO_UPDATE=1`, `env -i`, disposable HOME and XDG directories under task-local `.temp/`, `TBH_CREDENTIAL_BACKEND=file`, and `--provider echo`. It runs under the Apple sandbox-exec profile that denies all network, security-service lookups, Keychain file access, and the real native config/library directories. Keep the offline boundary controls: a loopback-allowed control against a denied connection, and readable ordinary bytes against a denied synthetic Keychains file. No login, logout or auth command. No real Keychain item. No remote inference: the shared Muse quota must not be touched.
- **Interactive form:** a private background tmux terminal, never the active UI.
- **Production entry points:** install the fixture with the real `curator profile install`, then `env resolve muse --profile <probe> --repair --format json`, and check the snapshot with the published spec checker, exactly as for 1.4.8.

## What the record must answer

1. P1–P4 per form, each with the actual observation and its exit code.
2. Has the interactive P1–P4 failure seen on 1.4.8 and 1.4.10 changed in 1.4.11? State it plainly either way. Do not promote an earlier claim to 1.4.11.
3. Parser refusals: the exec-only flag on interactive, and the interactive `-c` key on exec, as in 1.4.8.
4. The resulting §7.9 gate status per form, as one line at the top of the record, like 1.4.8's bold line.

## Tests and hygiene

Each harness test states the claim it proves. Every negative has a positive control. Add at least one narrowing control showing that the harness detects a planted violation (for example, a fixture that lets the foreign skill through must make the P-row fail). Run `go build ./...`, `go vet ./...`, the default `go test` (vendor tests skip) and the gated vendor run, and record exit codes. rc.7 wave rule: no edits to `CHANGELOG.md` or the root `README.md`; put a one-line changelog entry in the gaps file. Public-repo hygiene: no personal paths, hostnames, client or employer names, secrets.

## Hand-off

1. Write `TASK-261021-mw4laf_gaps.md`: the per-form verdict table, commands with exit codes, the binary digest, the changelog line, and any row left unmeasured, with the reason.
2. Run `task-board handoff TASK-261021-mw4laf --role developer`.
3. END YOUR TURN.

## Round note specific to 1.4.11

- The binary is the read-only pin `$HOME/Library/Application Support/tb-muse/muse-pins/muse-bin-1.4.11-R6667.1`, SHA-256 `6c04ce1719844685153ef6f04b98113cb61b942d97f0a8b0bfb3932c343984d9`. Never write to the `tb-muse` directory: copy the binary into task-local `.temp/` if the sandbox rules need it, and verify the copy's digest.
- The 1.4.10 record and harness are on main (`docs/qualification/muse-1.4.10-R6427.1.md`, `.research/muse1410/`, PR #150). Reuse their fixtures and harness code where unchanged, and make the 1.4.11 record state the per-row comparison with 1.4.10 (same, fixed or regressed).
- MSP schema changes in 1.4.11 are additive (session/leave, monitor inbox notifications, recoveryMode). Note any loader-relevant flag or config change you observe, especially for interactive `context.foreign_personal_skills`.
