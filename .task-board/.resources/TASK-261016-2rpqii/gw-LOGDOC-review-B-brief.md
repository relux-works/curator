# Review brief LOGDOC (deciding review B): public agent-logins reference (TASK-261016-2rpqii)

Read gw-LOGDOC-brief.md and logins-facts.md first. Attack surfaces:

1. **Truth against the code.** Every capability marked "works today" exists on `main`: run the CLI help or read the code. Anything from rc.7 or planned is marked as such. A command documented as available but absent is a P1.
2. **Truth against the facts.** Every measured statement matches logins-facts.md and carries its tool version. No invented behaviour. In particular, check the Claude desktop/ssh split, Muse's Keychain-first default and the Codex `file` default.
3. **Credential store.**
   - CLI spelling `curator env config set credential_store.<profile>.<env-id> native|file|keychain`, default `native`.
   - Machine-owned, with its reason.
   - The per-tool mapping, including the Claude refusals.
   - The statement that changing the store never migrates a login.
4. **Visibility.** There is a README pointer with the command and the default, and environment-config.md links to the new doc.
5. **Diagrams.** They exist, render, are embedded, and match the text.
6. **Hygiene, public repository.** No host, person, account, seat, lane or incident names, and no local paths.
7. **Checks.** The link check and `git diff --check` are green when you re-run them.

Verdict ACCEPT or REJECT with findings P1/P2/P3. Follow the board's review contract.
