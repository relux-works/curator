# Rework brief — TASK-260910-s9jz1g, revision 2 (R6)

Revision 1 was rejected with two medium findings
(`TASK-260910-s9jz1g_review-verdict-rev1.md`); everything else passed. Keep
the rest byte-identical.

- **F1 — activation must honour the configured passphrase.**
  `keys.py:208-210` (`activate-key-rotation`) loads the staged key and
  `os.replace`s the file without going through the provider store, so a
  plain staged key stays plain after activation even though
  `CSK_REGISTRY_KEY_PASSPHRASE` is set (reviewer reproduction: unset →
  genkey → prepare-key-rotation → set → activate … → active PEM still
  `BEGIN PRIVATE KEY`). Route every write of the active/staged key through
  the provider seam so the configured encryption applies at activation
  (re-export under the passphrase, atomic replace, staged identity/public
  pins and rotation semantics preserved). Test: the reviewer's exact
  sequence ends with an encrypted active PEM; plus the inverse (encrypted
  staged key activated under the same passphrase stays loadable).
- **F2 — a configured empty passphrase is an error, not "plain".**
  `keys.py:98-107` treats `CSK_REGISTRY_KEY_PASSPHRASE=''` like unset and
  writes plain PEM. The settled rule is: set ⇒ encrypted, unset ⇒ plain.
  Since the backend cannot encrypt with an empty password, reject a
  present-but-empty value with ONE diagnostic naming the variable before
  any key write or mutation (genkey, rotation, serve start with a key to
  load); retain plain behaviour only for true absence. Replace the tests
  at `tests/test_key_passphrase.py:47-64` that currently enshrine the
  empty-means-plain behaviour with the refusal, and add the "empty at
  load time" case.

Validation as before (pytest with `CURATOR_CONFORMANCE_ROOT` at the
`47c3c8c` root per the corrected rules, mypy strict); results "Revision 2"
section; tick the checklist; `task-board handoff TASK-260910-s9jz1g --role developer`.
