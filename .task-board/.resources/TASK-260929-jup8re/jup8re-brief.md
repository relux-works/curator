# TASK-260929-jup8re — naming gate: content-free reports + historical CI-echo exemption (THE ONLY CURRENT INSTRUCTION)

## Problem: a feedback loop that keeps main red

When `.github/ci/naming-gate.sh` fails, it prints every hit as `./path:LINE:<line content>`. The GitHub job log carries that content.
The task-board runner stores failure-log excerpts in board records: `progress.md` and the append-only
`.task-board/.activity/<ID>/events.ndjson`, which can never be edited or deleted. Those records are committed to main, so every
naming-gate failure plants new hits that the next gate run finds. Main 10821e67 is red for exactly this reason: the BUG-260928-uyak0e
`progress.md` and `events.ndjson` contain echoes of an earlier gate failure. Evidence: run 36533051091, job "Naming gate".

## Required change

1. The gate must NEVER print line content. On failure it prints only `path:line`, one per hit, plus the existing summary message. Do this
   for both the full-name and the short-name checks. Future failures then leave nothing name-bearing in any log or board record.
2. Add a narrow exemption for historical machine echoes of the old report format. Skip a candidate line only when it contains the GitHub
   Actions log-prefix signature of a naming-gate report: the step name "Employer name gate", then a TAB (a literal TAB character, or the
   two characters backslash-t as they appear inside JSON strings), then an RFC3339 UTC timestamp ending in `Z`, then a space and `./`.
   In regex terms: `Employer name gate(\t|\\t)[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9:.]+Z \./`. Lines without that signature are still scanned
   in full. Keep the existing binary-patch exemption unchanged.
3. Add rows to gate-selftest.sh. Assemble the names at runtime, as the existing rows do:
   - (f) a progress.md-style line with the signature and the short word → passes;
   - (g) a JSON line with `\t` escapes, the signature and the short word → passes;
   - (h) the short word on a line that has "Employer name gate" but no timestamp or `./` → fails;
   - (i) the gate's failure output contains no line content. Plant the short word in a temp tree, run the gate, and assert that
     stdout+stderr does not contain the planted word, only `path:line`.
4. Mutants, with real exit codes:
   - content printed again → row (i) fails;
   - signature relaxed to just "Employer name gate" → row (h) fails;
   - exemption removed → rows (f)/(g) fail, and so does the real tree.
5. Run `bash .github/ci/naming-gate.sh` against a `git archive origin/main` extraction (not the control root). It must exit 0 while the
   uyak0e records are present. Record the exit code.
6. In the results, state the residual: a deliberately forged line with that exact signature is exempt. Only machine-shaped CI-log echoes
   can be exempt; prose is never exempt.

Do not spell either name anywhere: not in code, tests, docs, results or the commit message. No CHANGELOG/LOGBOOK edit.

## Handoff

Update the results. Then run `task-board handoff TASK-260929-jup8re --role developer`, then END YOUR TURN. The runner publishes the CR
and runs the gate; do not wait for it.
