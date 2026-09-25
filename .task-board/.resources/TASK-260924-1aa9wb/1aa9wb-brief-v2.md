# TASK-260924-1aa9wb — Skillfile schema 2 default-on (THE ONLY CURRENT INSTRUCTION; supersedes 1aa9wb-brief.md)

Follow `skillfile-default-on-handoff-20260924.md` (attached) exactly — it CHANGES the earlier brief: the switch is REMOVED, there is
NO `=0` opt-out. Control root: curator; your Story worktree only (if a previous run left uncommitted edits, review them against this
brief and keep only what fits). Read `campaign-producer-rules.md`. Landing gate = hosted CI at handoff.
- Remove `CURATOR_DRAFT_SOURCES_V1` / `EnvDraftSourcesV1` / `DraftSourcesEnabled` and every `DraftSourcesV1` option/policy field
  (install Options, manifest ParseOptions, envprofile policy in envprofile.go/overlays.go); schema 2 is the default reader path at every
  call site listed in the handoff. `grep -rn "DRAFT_SOURCES\|DraftSourcesV1\|DraftSourcesEnabled"` over cmd/ internal/ docs/ README must
  end empty (tests included, except a row proving the variable is now ignored).
- PROVE schema-1 byte identity: a `schema_version: 1` Skillfile never enters the schema 2 path — run the v1 corpus / golden CLI outputs
  on the base binary vs the candidate binary and diff (reuse the stbg4d "switch on vs off, base vs candidate" recipe); zero diffs.
- A schema-2 project resolves/installs/updates end to end through the real CLI with no environment variable (path source, git tag
  source, repository source — local fixtures/bare repos, no network in tests).
- Global scope unchanged: `internal/install/global.go:395` behaviour ("no draft lane") kept; a row proving schema 2 is still not a global
  input.
- Diagnostics/help/docs: remove draft wording (docs/draft-source-expansion.md, docs/draft-transport-resolution.md — rename/retitle if they
  describe the now-default path — docs/cli.md, README, docs/troubleshooting.md), CHANGELOG entry. Rename internal "draft" identifiers only
  where it is mechanical and safe; test names/case IDs that name spec draft vectors stay until the pin moves.
- Tests that asserted "schema 2 refused without the switch" become "schema 2 accepted by default" rows; never delete assertions.
- Narrowing mutants (disposable copy): schema-1 file routed into the schema-2 path → killed; schema 2 refused again → killed.
Bounded local runs (≤10 min/call, per package). Attach results, `resource update`, check DoD, `task-board handoff TASK-260924-1aa9wb
--role developer`. A `run_wrote_outside_worktree … policy warn` block is a warning — verify status `to-review`.
