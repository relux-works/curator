### A1 fragment delivery (SPEC §4.1) — for the launcher control-root LOGBOOK (parent appends)
- MILESTONE: `internal/fragment` implements §4.1 end to end; `curator-run <env>` now resolves the fragment through the real `curator env resolve … --repair --format json` subprocess before the `not_implemented` refusal.
- FINDING: installed curator `v0.14.1-0.20260907213730-04550e282705` reproduces the three A0 digests byte-for-byte; canonical CCJ-1 of the parsed object equals Curator's printed line minus LF (environments.md §10.1 holds).
- FINDING: on this machine no profile is current — `curator env resolve pi` without `--profile` is `profile_unknown: no profile is current`; A0 evidence was recorded with `--profile default`.
- DECISION: fragment channel lists are validated for exact equality with the environments.md §7.3/§7.8 adapter registry; a registry revision in curator-spec requires a launcher release.
- DECISION: standard-library JSON is not used for the fragment (accepts duplicates, repairs bad UTF-8/surrogates, floats); a purpose-built CCJ-1 reader in `internal/fragment/json.go` enforces registry.md §1 pre-canonicalization rejections.
- FIX: SPEC §4.1 E6 clarification — code read from the first `curator: <code>: <detail>` stderr line, stderr forwarded verbatim incl. exit-0 warnings (STORY-260908-2utz8k).
- ANOMALY: curator-spec `conformance/v1/schema-cases/launch-env-fragment-v1/` holds 13 pre-D8 files absent from `index.json` (e.g. `valid-empty-channels.json` with `profile.commit`); not vendored.
- SCOPE: `internal/fragment/*`, `cmd/curator-run/main.go`, `main_test.go`, `SPEC.md` §4.1, `README.md`, `.scripts/fragment-mutants.sh`, `.scripts/cli-mutants.sh` M14.
- STATUS: make check exit 0; 40/40 narrowing mutants killed; handed off to review.
