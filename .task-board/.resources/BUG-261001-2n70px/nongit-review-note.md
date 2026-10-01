# Review note — BUG-261001-2n70px rev3 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review the newest CR (rev3, base f0119a8b, 10 paths, gate green) against `nongit-brief.md` and `nongit-gate-note.md`. It touches install safety and CI policy, so review both carefully.

1. **Behaviour, at the real `curator install` entry point** (throwaway HOME, temp non-git product folder holding a nested git repo):
   - Declared project skill bytes are present, exit 0, and a clear "not a git work tree, ignore check not applicable" notice is printed.
   - Inside a real git repo, the gitignore behaviour is unchanged: run an ignored-path conflict row and a missing-entry row.
   - A broken git (corrupt .git file, or any git exit other than the not-a-repository 128) refuses non-zero. Arbitrary git errors must never be treated as "not a repo". Check exactly how 128 is distinguished, e.g. stderr matching versus `rev-parse --is-inside-work-tree`.
2. **CI policy files** (.github/ci/platform-cases.tsv, skip-classes.tsv, gate-selftest.sh):
   - Did the producer add a NEW skip class or ledger rows?
   - Is any Windows skip justified, or could the row run on Windows with a shim? A new skip class must not weaken the gate for other tests; check that gate-selftest covers it.
   - Prefer running on Windows. A skip-only answer needs a concrete reason.
3. **Mutant:** restore the old skip; the positive row must fail. Use real exit codes.
4. **Docs and hygiene:** the cli.md paragraph is accurate; one CHANGELOG line; no LOGBOOK; never spell any employer name.

accept_cr, or changes requested with file:line.
