# TASK-260827-2232c0: readme-restructure

## Description
Restructure README.md per .spec/docs-refresh.md: definition, what Curator manages, install as collapsible details blocks per platform (Homebrew open by default, installer script, Scoop, Go toolchain), quick start, tightened protocol section, short development section linking CONTRIBUTING.md. MOVE OUT wholesale into a new docs/compiled-commands.md: the sections Compiled-command status diagnostics and repair, and Maintenance and the build-cache grace period; restructure their prose to the style guide but preserve every fact and command. Leave a one-paragraph summary plus link in README where each section was. Also add a two-line historical header to docs/implementation-plan.md (plan of record for v0.1 against protocol 1.0.0-rc.2; the task board is the live plan), content untouched. Verify every command you keep or move against go run ./cmd/curator ... --help from this tree.

## Scope
README.md, docs/compiled-commands.md (new), docs/implementation-plan.md header only.

## Acceptance Criteria
README under 220 lines with no reference dumps; compiled-commands.md preserves every fact and command from the moved sections, verified against the tree binary; details blocks render; historical header present; style guide holds.
