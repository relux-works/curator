# TASK-260916-16ys92: launcher-print-resolved-provider-path

## Description
curator-agent-launcher: print the resolved provider path (its own executable path as resolved by the umbrella) in the launch stderr line-group and record it in the SPEC §4.3 line-group description.

## Scope
curator-agent-launcher SPEC.md §2/§4.3/§8.1 + changelog, cmd/curator-run and internal/defaults line-group emission of the resolved provider path, goldens and unit tests, CHANGELOG

## Acceptance Criteria
SPEC revision and implementation merged; golden shows the provider path line
