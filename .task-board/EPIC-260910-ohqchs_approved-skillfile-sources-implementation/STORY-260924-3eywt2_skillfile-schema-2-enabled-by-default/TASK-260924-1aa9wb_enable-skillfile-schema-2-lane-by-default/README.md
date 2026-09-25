# TASK-260924-1aa9wb: enable-skillfile-schema-2-lane-by-default

## Description
Make the skillfile-sources-v1 lane (Skillfile schema 2, sources, lock, marker v5, source policy) the default: internal/install DraftSourcesEnabled returns true unless the operator explicitly sets CURATOR_DRAFT_SOURCES_V1=0 (transitional opt-out, documented as deprecated); every call site (cmd/curator project_resolve, draft_status, draft_help, main help interception, internal/install install.go) follows; schema-1 projects keep their exact meaning and byte-identical outputs with the lane on (spec: Skillfile schema 1 retains its exact meaning; no implicit on-disk migration); README, docs/cli.md, docs/troubleshooting.md, help text and CHANGELOG updated; tests that asserted the refusal without the switch now assert the opt-out.

## Scope
(define task scope)

## Acceptance Criteria
1) default (unset) = lane on; =0 = off; any other value = on; 2) schema-1 byte identity with the lane on vs off proven on the existing v1 corpus and CLI outputs; 3) schema-2 project works with no env var through the real CLI entry; 4) opt-out row; 5) narrowing mutants (default back to off; =0 ignored) killed; 6) docs/help/CHANGELOG
