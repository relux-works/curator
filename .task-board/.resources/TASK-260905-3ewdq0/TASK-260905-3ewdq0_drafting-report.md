# TASK-260905-3ewdq0 drafting report: launcher SPEC `0.2.0-draft`

## Commit

- Repository: `curator-agent-launcher`, worktree
  `/Users/iv/Developer/ReluxWorks/.worktrees/curator-agent-launcher-spec-0.2`,
  branch `draft/spec-0.2`, base `6de42d8`.
- Head: `ffe9b68f3579d05a7b17bce32e0f0126f141852f` — one signed commit, not pushed, no tag, no PR.
- `git log --show-signature -1`:

```text
commit ffe9b68f3579d05a7b17bce32e0f0126f141852f
Good "git" signature for ivan@relux.works with ED25519 key SHA256:Ng99XGF2pboYgFVfWJhYI2JRi0PyYsV9UwsJ70NBYd0
Author: Ivan Oparin <ivan@relux.works>
Date:   Sat Sep 5 11:37:00 2026 +0400
```

- Files: `SPEC.md`, `README.md` (line 24), `cmd/curator-run/main.go`
  (`specVersion`, usage text), `cmd/curator-run/main_test.go`
  (`TestSpecVersionPinned`). Nothing else touched; LOGBOOK.md and the
  control root untouched.

## Decision 0013 item → SPEC section

| Decision 0013 item | Brief item | SPEC `0.2.0-draft` section |
|---|---|---|
| D6.1 fragment-first ordering, `LaunchRequest.Home` = fragment home variable, `WorkDir` = cwd, limit state keyed by (provider, home) | 1 | §4 preamble (six steps, order as contract), §4.1 (home variable per environments §7.1), §4.4 `Home`/`WorkDir` bullets |
| D5 / D6.5 `LaunchModeInteractive` by name, no provider flag, argv = model + effort transport, empty `Composition`, `ErrCompositionNotInteractive` | 2 | §4.4 request members; §1 "No plan rebuilding" restated (composition ≠ rebuilding) |
| D6.2 default precedence flags → machine config → lineup, per-member, resolved pair + level on stderr, `--effort` completion | 3 | §4.3 (with `defaults.json` location + closed `curator-run-defaults-v1` schema + lock rule); §3 `--model`/`--effort` rows |
| D6.3 composition rule: argv order, MCP flags per adapter, four-layer environment, `env_names` = `mcp.env_names`, literal-vs-lookup collision rule (F5), stdin under D4 | 4 | §4.5 (argv / environment / env_names / stdin); order-is-contract sentence there; per-tool boundary item in §9 |
| D6.4 tracked mode: exact `ax start` invocation, document members, `argv_suffix` minus element 0, `env_literals` composer-only, four extension keys with derivations, provider-id column, `--ax-profile`, session name `<env-id>-<utc-stamp>`, `--name` + ax §2.1 grammar + >64 = `usage`, `ax_handoff_failed` with Structured Error pass-through, untracked exec | 5 | §4.6 (invocation, operand list, document table, extension-key table, failure, untracked shape); §4.2 provider-id column; §3 `--name`, `--ax-profile` rows and parsing rules |
| D6 "keep §5 and §6 unless named" | 6 | §5 unchanged except cross-references §4.5→§4.6; §6 gains `defaults` family (`defaults_config_invalid`, `defaults_unresolvable`), usage row extended, both invariants kept and the first extended to the defaults file |
| D5 dependency on the interactive-mode module release | 7 | §7: release carrying `LaunchModeInteractive` required; exact `go.mod` pin restated when it exists |
| Versioning, §8.1 row, §9 open items (ax shape replaced by D6.4 + PR #1 revision open; per-tool boundary; implementability per Consequences; open questions 2/5) | 8 | §8 version, §8.1 `0.2.0-draft` row, §9 rewritten |
| D1 ownership / single composer | — | §4 preamble, §4.6 |
| Decision 0012 D6 + D8 fragment members (`lock_sha256`, `mcp`, precedence primitives, `composition` withdrawn; 1.1 rewrite pending) | — | Normative references; §4.1 member list |
| D7 item 6 fragment-digest over CCJ-1 of the parsed object | 5 | §4.1 last paragraph; §4.6 key table |

## Gates

`make check`, run directly, exit code 0 (worktree head `ffe9b68`):

```text
go build ./...
go vet ./...
go test ./... -count=1
ok  	github.com/relux-works/curator-agent-launcher/cmd/curator-run	0.471s
```

`git diff --check` exit 0. `gofmt -l cmd/` empty.

Negative evidence for the one new test: with `specVersion` mutated back
to `0.1.0-draft`, `go test -run TestSpecVersionPinned` exited 1 (FAIL);
restored before the commit.

## Docs-confidence items (not verified against a running tool or release)

- `defaults.json` location (`$XDG_CONFIG_HOME/curator-run/`,
  `/etc/curator-run/`), the `curator-run-defaults-v1` schema, and the
  `locked` rule are this draft's own design; Decision 0013 D6.2 says only
  "closed launcher machine-configuration mapping … lockable". Flagged in
  §9 as movable into Curator machine configuration by revision.
- `--ax-profile` on an untracked machine is specified as a `usage` error;
  `--name` on an untracked machine is accepted and ignored. D6.4 fixes
  neither; both are drafting choices for review.
- Environments §7.1 "home variable" for `opencode` is `XDG_CONFIG_HOME`
  (parent, not home); §4.1 passes the parent path as `Home`. Moot in
  this revision (`opencode` is `env_unsupported`), noted for the row's
  future fill.
- All `ax` section numbers (§2.1, §5.1, §7.1, §14.1, §15.1) and
  `agents-management` identifiers are cited from Decision 0013, not
  re-read from those repositories.
- Per-tool argv boundary and codex `-p` layering remain unverified by
  design (§9).
