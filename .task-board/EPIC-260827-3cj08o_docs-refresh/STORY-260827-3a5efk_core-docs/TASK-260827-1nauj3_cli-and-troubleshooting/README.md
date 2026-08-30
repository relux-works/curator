# TASK-260827-1nauj3: cli-and-troubleshooting

## Description
Two new documents. (1) docs/cli.md: full command reference for curator, one section per command group, every synopsis and flag verified verbatim against go run ./cmd/curator ... --help (or make build then ./bin or ./dist binary from this tree; state which in the outcome); shared flags described once consistently. (2) docs/troubleshooting.md: symptom-cause-remedy entries for the highest-value failures: compiled-command diagnostics from the README material, toolchain preflight mismatches, external repository fetch and credential failures (SSH and HTTPS), verify every error identifier against internal/ sources with grep evidence. Then add a Commands section to README.md before the protocol section: collapsible details groups with one-line command # what it does entries and a link to docs/cli.md.

## Scope
docs/cli.md (new), docs/troubleshooting.md (new), README.md Commands section.

## Acceptance Criteria
Every synopsis and flag matches the tree binary help verbatim; every error string greps in internal/; README Commands groups cover all top-level groups; links resolve; style guide holds.
