# THE ONLY CURRENT INSTRUCTION — TASK-261018-adogt7: rework round 8 of W-SKILLDIR (developer)

Review B on CR revision 3 requested changes. Your round-7 fix (the ruleset bump) is accepted. Read the attached verdict `TASK-261018-adogt7_review-verdict-rev3.md` and its probe `TASK-261018-adogt7_review-numeric-probe-rev3.py`.

## Finding `lock-numeric-discriminator-coercion` (bypass, AC 3)

The newly supported schema-4 lock must enforce the JSON types of the complete normative carrier (rc7-carriers §1): the numeric discriminators and member weights are JSON numbers. `internal/contextlock/contextlock.go:438-439` and `:452` decode those fields into `json.Number`. That type accepts a JSON string such as `"4"`, and the `Atoi`/`ParseInt` at `:461`, `:467` and `:491` erase the difference. `ReadWith` calls `Parse` at `:537` after canonical JSON validation, which does not establish carrier types. Result: `contextlock.Read` admits an on-disk schema-4 lock whose `schema_version`, `hash_version` or `weight` is a string.

## Do this

1. Make the schema-4 reader refuse a string (or any non-number) where the carrier requires a number: `schema_version`, `hash_version`, member `weight`, and every other numeric field of the schema-4 carrier. Use the typed refusal the reader already uses for malformed carriers. Fix the scoped reader without weakening the historical guarantees for earlier lock schemas: if older schemas share the decoder, state in the gaps file whether their behaviour changes and why that is correct under the spec.
2. Turn the reviewer's probe into named regression tests in the repository: each string-valued field must be refused through `contextlock.Read` on an on-disk file. Each one needs a passing numeric positive control. Add a narrowing mutant that re-admits a string in exactly one of these fields; it must fail the regression for that reason.
3. Full local suite on the exact candidate tree (the verdict requires this receipt before acceptance). A single shell call is limited to about 10 minutes, so run `go test ./...` in bounded sequential package groups, with every package covered exactly once. Record each group's command and exit code in the gaps file, plus `go build ./...` and `go vet ./...`. Do not reuse older receipts, and do not present the post-turn remote gate as local execution.
4. Stay inside the original brief `gw-SKILLDIR-brief.md`. rc.7 wave rule: `CHANGELOG.md` and the root `README.md` equal trunk. Do not run `refresh-candidate` unless `worktree status` reports a state that only it resolves. Do not add checklist items about the post-turn validation.

## Hand-off

1. Add a "Rework round 8" section to `TASK-261018-adogt7_gaps.md`: the fix with file and line, the tests, the mutant, and the full local suite table.
2. Run `task-board handoff TASK-261018-adogt7 --role developer`; it must publish a new CR revision.
3. END YOUR TURN.
