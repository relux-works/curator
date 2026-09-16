# TASK-260916-rkrphg rev2 — launcher aliases: persisted-byte regression (rework 1)

Story worktree: `curator-agent-launcher/.temp/STORY-260916-prdjid/worktree`
(branch `task-board/story/STORY-260916-prdjid`, changes uncommitted for handoff snapshot)

## Scope of this rework

Verdict rev1 (CHANGES_REQUESTED, one finding): the persisted-byte regression
required by review-brief item 2 was missing — `TestProductionAliasEquivalence`
compared child payload/stderr but never read launcher-persisted state, and
`TestLoadRejectsAliasKeys` proved only reader rejection.

This rework adds exactly that, per `rkrphg-rework-1.md`, and keeps everything
else unchanged (no production-code change, no golden change, no SPEC/README
change beyond rev1).

## Changes (2 files)

1. `cmd/curator-run/pipeline_test.go` (+258):
   - `TestProductionAliasPersistedBytesEqual` — through `run` (the production
     entry point), both aliases (`claude`, `codex`) with the canonical
     spellings (`claude_code`, `codex_cli`) as controls, tracked and untracked
     (4 subtests x 2 launches). Each launch snapshots the whole fixture tree
     before/after and asserts:
     - the launch happened and the snapshot is non-vacuous (`started` fake
       child/ax side effect present; managed-home `mcp.toml`/`prompt.md` and
       tracked `ax.json` present);
     - pre-existing managed-home/config files are byte-identical after the
       alias launch (launcher/child must not rewrite managed state);
     - alias-run persisted bytes equal canonical-run persisted bytes
       (relpath sets equal, per-file bytes equal, fixture root normalized to
       `<ROOT>` as in `golden()`);
     - no persisted file content carries the alias as an environment id
       (`assertNoAliasEnvID`: JSON `"environment":"<alias>"`,
       `environment=<alias>`, and alias-shaped default session names
       `<alias>-YYYYMMDDTHHMMSSZ`; file names are never scanned, so provider
       executable names are legitimate by construction).
   - `TestAliasEnvIDScanFlagsCanary` — proves the scan is not vacuous: every
     hunted pattern fires on a planted canary, and a provider-name mention
     (`--provider <alias>`) is not flagged.
   - Enumeration (in the test comment, from code paths): the launcher writes
     no files — `defaults.Load`/`axconfig.Load` are reads-only,
     `diagnostics.Emit`/`defaults.EmitGroup` write to the passed stderr
     writer, no `os.Write*`-family call exists in non-test `internal/` code
     (only a `/dev/tty` open in `internal/execution/process.go`), `plan.Build`
     documents `AvailabilityFor` as a read, and
     `vendorplugin.BuildLaunchWithEnvironment` is pure planning (no
     claim/report write path). Bounds stated in-test: Curator-owned repair
     writes are out of reach (scripted resolver runs no subprocess — the fake
     evidence does not prove Curator persistence; Curator side is
     TASK-260916-11lwua); real-ax session records are ax-owned (fake ax
     persists only constant `started`; handoff bytes proven by child-argv
     capture + goldens in `TestProductionAliasEquivalence`).
   - Bounded per subtest: 2-minute context for the resolve/plan stages, no
     sleeps, no polling.
2. `internal/defaults/defaults_test.go` (comment only): fixed the overstated
   `TestLoadRejectsAliasKeys` comment — it is now documented as the
   reader-side bound (load/refuse), with launch-time persistence deferred to
   `TestProductionAliasPersistedBytesEqual`.

## Evidence (real exit codes, shell: bash via tool runtime)

Narrow gates only, per brief (host stalls on big suites; remote gate runs at
handoff; the full landing suite is left to the handoff runtime):

- `gofmt -l cmd/curator-run internal/defaults internal/cli` → exit 0, no files
- `go build ./...` → exit 0
- `go vet ./cmd/curator-run ./internal/cli ./internal/defaults` → exit 0
- `go test ./cmd/curator-run -run 'Alias|Persist' -count=1` → exit 0
  (`ok ... 4.690s` post-revert re-verify; an earlier cold-cache run also
  exit 0 in 161.274s). Covers `TestRunAliasesBehaveAsCanonical`,
  `TestRunAliasFragmentNeverAccepted`, `TestProductionAliasEquivalence`,
  `TestAliasEnvIDScanFlagsCanary`, `TestProductionAliasPersistedBytesEqual`
  (all PASS, incl. all 4 alias/tracked persisted-bytes subtests).
- `go test ./internal/cli ./internal/defaults -count=1` → exit 0
  (`ok internal/cli 0.448s`, `ok internal/defaults 1.480s`).

Mutant attack (narrowing, then reverted): `NormalizeEnvID` stubbed to the
identity map → `go test ./cmd/curator-run -run 'Alias|Persist' -count=1` →
exit 1 (expected-red): every alias subtest fails closed with
`resolve_fragment_invalid: fragment names environment "<canonical>", resolve
was asked for "<alias>"`. The fragment layer refuses the mismatch, so a
broken normalizer cannot launch. Mutant reverted; `git diff` confirms the
rev1 `internal/cli/cli.go` hunk intact, and the suites above were re-run
green after the revert.

Piped runs in this session quoted `${PIPESTATUS[0]}` (gate status preserved);
the final re-verify runs above are unpiped with direct exit codes.

## Checklist status

Producer items (1–8) remain satisfied by this rework: entry-point tests for
both spellings incl. the new persisted-byte regression, help/README from
rev1 untouched, unknown ids still refused (`TestRunUnknownSpellingsStillRefused`
unchanged), narrow tests/goldens green with quoted exit codes, build green,
new task-scoped outcome artifact (this file). Items 9–12
(`Implementation matches AC`, `Solution fits project architecture`,
`Tests green`, verdict routing) are reviewer branches — left unchecked for
review.

## Full diff vs Story base (19 files, +636/−31)

rev1 (unchanged): `README.md`, `SPEC.md`, `cmd/curator-run/main.go`,
`cmd/curator-run/main_test.go`, `internal/cli/cli.go`,
`internal/cli/cli_test.go`, `internal/cli/testdata/cases.golden`,
`cmd/curator-run/testdata/{help,forbidden-*}.golden`.
rev2 (this rework): `cmd/curator-run/pipeline_test.go` (+258 new test code),
`internal/defaults/defaults_test.go` (comment fix, +3/−2).
