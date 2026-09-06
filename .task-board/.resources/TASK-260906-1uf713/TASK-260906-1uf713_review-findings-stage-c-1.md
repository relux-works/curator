# TASK-260906-1uf713 — review findings, stage (c) cycle 1

Verdict: **CHANGES REQUESTED**. Four blocking findings, three major, four minor.

- Subject: `~/Developer/ReluxWorks/.worktrees/curator-stage-c`, branch
  `feat/agent-environments-stage-c` at `833918d2ebd1f0fa60b117899844a36dec1688b6`
  (PR https://github.com/relux-works/curator/pull/61, head confirmed by
  `gh pr view 61 --json headRefOid`). The producer worktree was read-only for this
  review: `git status --short` empty and HEAD unchanged before and after.
- Authority: curator-spec `550579d` (this story worktree).
- Method: every finding below is driven through the production entry point —
  a `curator` binary built from `833918d2` (`go build ./cmd/curator`, exit 0),
  invoked as a standalone process against scratch homes under
  `.temp/review-c1/`, with `CURATOR_CONFIG`, `CURATOR_SYSTEM_CONFIG`,
  `CLAUDE_CONFIG_DIR`, `CODEX_HOME`, `XDG_CONFIG_HOME`, `PI_CODING_AGENT_DIR`,
  `HOME`/`USERPROFILE` and (where a git identity is needed)
  `GIT_CONFIG_GLOBAL` insteadOf rewrites pinned to the scratch tree. Mutants were
  applied to a throwaway copy (`.temp/review-c1/mut`), never to the worktree.

---

## Hosted lanes on the reviewed head

`gh pr checks 61` on `833918d2`:

| Lane | Result |
|---|---|
| Lint | pass 39s |
| Naming gate | pass 6s |
| Interop conformance gate | pass 22s |
| Gate self-test (ubuntu / macos / windows) | pass 8s / 10s / 28s |
| Test (ubuntu-latest) | pass 3m20s |
| Test (macos-latest) | pass 8m7s |
| Race (ubuntu-latest) | pass 10m0s |
| Race (macos-latest) | pass |
| Test (windows-latest) | **still pending when this verdict was recorded** |
| Candidate suite | skipping (expected: gated matrix) |

No hosted lane was red. **Test (windows-latest) had not reported**, so this review
does NOT claim every hosted lane was green on the accepted head — and nothing is
accepted here anyway. The pending Windows lane is the one that would catch the
class of fixture defect stages (a) and (b) paid for; the rework cycle must read it.

The two local `test-gate.sh` lane reproductions in the increment-2 report were not
re-run here: they are not what the blocking findings turn on, and re-running them
sequentially costs ~40 minutes of wall clock that produces no evidence about the
holes below. Stated plainly rather than implied.

---

## BLOCKING

### B1 — `profile install --use` bypasses the locked `require_current_profile` gate (§12.2)

`Config.CheckMachineUse` is called from `cmdProfileUse` and nowhere else
(`internal/config/environments.go:768`, its only production caller is
`cmd/curator/profile.go:212` inside `cmdProfileUse`).
`Install` performs the §9.2 machine-scope switch on `--use` and on first install
(its own doc comment says so, `internal/envprofile/envprofile.go:521`), and never
consults the gate.

Driven, system file locking `environments.require_current_profile` to `acme`:

```
$ curator profile use personal
curator: environments.require_current_profile: profile use of "personal" is refused:
         the system configuration requires current profile "acme"        # exit 1  — correct

$ curator profile install ./probe/rootgit --use
installed and activated profile p5 (...)                                  # exit 0
$ curator profile list
acme  |            default |            p5 | current            personal |
```

The machine current profile is now `p5` with the key locked to `acme`, exit 0, no
warning. Same on a fresh machine: the first `profile install` of any other profile
activates it. §12.2 makes `profile use` of any other profile in the machine scope a
configuration error; this is a bypass path around the check, and `overlays_allowed`
plus `require_current_profile` are the whole of revision 1's fleet-policy surface.

Negative shape: *a bypass path around the check*. Fix must drive the gate from the
activation seam (the §9.2 switch), not from one CLI row.

### B2 — `--takeover` of a foreign-manager symlink writes THROUGH the link, outside the managed home

`materializeOne` writes the `claude_code` root context with
`os.WriteFile(target, document, 0o644)` (`internal/envprofile/switch.go:498`,
and the symlink-fallback copy at :509) without first removing an existing symlink.
`os.WriteFile` follows a symlink. Stage (c) is what first routes a foreign-manager
symlink into that write: before `--takeover` the entry was refused unconditionally.

Driven, `CLAUDE.md` a symlink to `<scratch>/foreign/CLAUDE.md` holding
`PRECIOUS FOREIGN CONTENT`:

```
$ curator profile use acme                      # exit 1, correct
claude_code: environment_foreign_manager_detected: CLAUDE.md is a symlink outside
             the manager store; abort, or take over with backup

$ curator profile use acme --takeover            # exit 0
notice: taking over unmanaged CLAUDE.md (foreign-manager symlink) for claude_code ...

$ ls -la <native>/claude/CLAUDE.md
CLAUDE.md -> <scratch>/foreign/CLAUDE.md          # STILL a symlink
$ cat <scratch>/foreign/CLAUDE.md
<!--
curator-root-context-v2                            # the generated document
...
```

After a "take over with backup", the surface is still owned by the other manager and
curator has overwritten a file it never inventoried, at a path outside every managed
home — in the field, the dotfile manager's own source of truth, which its next apply
propagates. §9.5 offers "abort, or take over with backup — never a silent
absorption"; this absorbs in the wrong direction.

The linked adapters are correct: `replaceLink` removes first, and driving the same
fixture on `codex_cli`/`AGENTS.md` leaves the foreign file untouched and repoints the
link into the store. Only the §8.1 always-copied `claude_code` surface (and the
symlink-fallback copy) is affected.

`TestForeignSymlinkStopsSwitch` (`internal/envprofile/takeover_test.go:161-169`)
asserts only `result.OK` after the takeover. It never asserts the surface stopped
being a symlink and never asserts the foreign bytes survived. Positive-path evidence
over exactly the hole.

### B3 — `profile import` swallows curator's own managed root-context files as "native" surfaces

`detectNative` reads the adapter root-context target with no marker or managed-state
check (`readRootSurface`, `internal/envprofile/import.go:233-248`, reached from
`detectNative` at :200). §9.5 step 1
inventories *unmanaged* root-context files, and §9.6's detected-surface list is that
inventory.

Driven on a machine curator already manages (profile `acme` current):

```
$ curator profile import                                        # exit 0
installed profile imported (root imported 1.0.0, lock sha256:b1f95e...)
$ head -8 <store>/context/imported/<hash>/context/claude_code.md
<!--
curator-root-context-v2
root: acme 1.0.0 state sha256:1b8ca9...
member: acme 1.0.0 state sha256:1b8ca9... weight 0
precedence: winner=higher-weight placement=winner-last
lock: sha256:d141a7...
generated: Curator Protocol environments revision 1 (...)
notice: generated file; direct edits are unsupported ...
-->
```

Four modules, one per adapter, each carrying a complete generation header, the
`notice:` line and the chapters of the *currently installed* profile. Activating that
profile nests one §5.1 generation header inside another inside a chapter — a shape
the closed §5.1 grammar does not admit, and which drift and marker verification then
read. `detectNative` already has an `isStoreEntry` guard for skills entries; the
root-context branch has no equivalent.

### B4 — divergent same-named skills across adapters collapse silently, with no loss-list entry

`reassembleImport` keys `requires.skills` by bare skill name
(`internal/envprofile/import.go:434`, `entries[skill.name] = ...`), so two adapters
whose global skills surfaces both carry `foo` at different commits produce one entry
— last writer wins by ascending environment identifier.

Driven, `claude_code/skills/foo` at `ae91c3ce…` and `codex_cli/skills/foo` at
`42d2fcb5…`, both with the same canonicalizing `origin`:

```
$ curator profile import                                        # exit 0
warning: environment_import_skill_foreign: foo was managed by other means ...
warning: environment_import_skill_foreign: foo was managed by other means ...
$ cat <store>/context/imported/<hash>/agent-context.json
  "requires": { "skills": { "foo": { "git": "example.com/foo",
                                     "revision": "42d2fcb5e2dcfd8e36c81bb5f0fb88ea1a632df6" } } }
```

Two detected entries, two warnings, one entry — `ae91c3ce…` is gone. §9.6 requires
"one `requires.skills` entry per mapping skills entry", and the loss list must name
every loss; this is neither an entry nor a loss. A machine with one skill installed
at different commits in two agent homes is the ordinary onboarding case, and the
whole point of §9.5 is reaching managed state *without loss*.

Negative shape: *a capability claim that does not reproduce* — the report's
"one `requires.skills` entry per mapping entry" row is true only for a single
detecting adapter, which is the only shape `TestImportSkillForeignPinnedByRevision`
constructs.

---

## MAJOR

### M1 — `path` overlays are unreachable from every production surface, and the AC row claiming otherwise rests on a manufactured Policy

Three gates compose into a dead feature:

1. `config.parseOverlay` requires exactly one of `range`/`tag`/`revision` on every
   overlay (`internal/config/environments.go:426`);
2. `resolveOverlay` refuses any form on a path source with `profile_source_invalid`
   (`internal/envprofile/overlays.go:47-50`, refusal at :49);
3. `profile compose add` requires exactly one form (`cmd/curator/compose.go:53`).

Driven:

```
# no form
$ curator profile list
curator: environments.overlays.acme[0]: requires exactly one of range, tag, or revision   # exit 1
# with a revision (the shape the published valid.json uses for a path source)
$ curator profile install ./pkgs/root
curator: overlay 0 of profile "acme": profile_source_invalid: a path overlay carries no
         range, tag, revision, or directory                                                # exit 1
# the CLI row
$ curator profile compose acme add ./pkgs/ov
curator: profile compose add takes exactly one of --range, --tag, --revision               # exit 2
```

No operator can declare the §6 `path` overlay. `TestPathOverlayJoinsClosure`
(`internal/envprofile/overlays_test.go:204-208`) proves the behaviour by building
`Policy{Overlays: {"acme": {{Source: overlay}}}}` in Go — a state
`PolicyFromConfig` cannot produce. `cmdComposeList`'s `default: form = "path"` branch
is likewise dead.

The root cause is a genuine spec/schema tension — `manager-config-v2`'s `overlay`
`oneOf` requires a form, `environments.md` §1 makes a form on a `path` declaration
`profile_source_invalid`, and `valid.json` publishes a path source carrying a
revision — and the producer noted it in code comments. But the increment-2 report
lists this AC row as **driven**, not as a bound, and the finding here is that a
normative §6 sentence ships unusable with no declaration. Either escalate the spec
contradiction under stop-the-line, or carry it as an explicit stated bound; do not
report it driven.

Related: the ledger row
`internal/config TestPathOverlayDeclarationParses ... a path overlay carries no
requirement form and no directory` (`.github/ci/platform-cases.tsv`) states the exact
opposite of what the test asserts and the reader does.

### M2 — the unreadable-path diagnostic is shadowed; the declared M3 survivor's rationale is false

`stateForPath` wraps *every* `EnsureState` error with
`fmt.Errorf("%s: %v", DiagSourceInvalid, err)` (`internal/envprofile/envprofile.go:1030`).
Driven at the production entry point:

| Fixture | Emitted | §1.1 requires |
|---|---|---|
| path root `chmod 000` | `profile_source_invalid: read agent-context.json: ... permission denied` | `profile_source_path_unreadable` |
| unreadable dir below the root | `profile_source_invalid: profile_source_path_unreadable: ...` | `profile_source_path_unreadable` leading |
| unreadable file below the root | `profile_source_invalid: profile_source_path_unreadable: ...` | as above |
| missing operand | `profile_source_path_missing: ...` (clean) | correct |

The canonical §1.1 condition — "a `path` operand names a directory that cannot be
read" — reports `profile_source_invalid` and nothing else. §8.4: unreadable evidence
is reported as unreadable, never as something else.

`TestUnreadablePathIsUnreadable` (`internal/envprofile/pathkind_test.go:190`) asserts
`strings.Contains(err, DiagPathUnreadable)`, so the prefix is invisible to it, and the
unreadable-root case it does not construct has no test at all.

This also refutes the increment-2 mutant table's only declared survivor. M3's stated
bound is "`pathManifestDiag` precedes the store, so the store branch fires only on a
manifest/snapshot race (TOCTOU)". It does not: `pathManifestDiag` decides the
unreadable-root case and returns the wrong diagnostic, and the store branch is reached
by any unreadable entry below the root — an ordinary addressing mode, not a race.
Stage (a)'s F1 shape, as the review brief anticipated.

### M3 — `env status` does not report the locked `require_current_profile`, and increment 2 dropped the bound

§12.2: a locked `require_current_profile` "makes `profile use` of any other profile in
the machine scope a configuration error under the manager §1 locked-key rules, and
`env status` reports the requirement." Driven with the key locked to `acme`:

```
$ curator env status --json | ...
has require_current_profile: False
top-level keys: ['homes','scopes','adapters','targets','profiles',
                 'unregistered_environments','orphans','notes','non_current']
$ curator env status | grep -i requir      # no output
```

Increment 1 declared this a stated bound. Increment 2 claims "AC coverage: 15 of 15
stage (c) rows driven" and does not carry it forward. An unimplemented normative
sentence presented as complete is a report defect on top of the gap.

---

## MINOR

### m1 — surviving narrowing mutant on the lockable-subset gate (mine, not the producer's)

Applied to the throwaway copy, `internal/config/environments.go:795`:

```go
- if !allowed[key] {
+ if !allowed[key] && key != "forms" {   // gate stays, admits exactly one member
```

`go test -count=1 ./internal/config/` → **ok**, and
`CURATOR_CONFORMANCE_ROOT=<spec-main>/conformance/v1 go test -count=1 ./internal/config/`
→ **ok**, with `TestSystemConfigV2SchemaCases` confirmed running all 24 cases. The
published family covers `current_profile` and `overlay_default_weight` but not
`forms`, so the gate's coverage is per-knob-name rather than class-wide, and neither
report's mutant table attacks `parseSystemEnvironments` at all.

For contrast, two other narrowing mutants I applied were killed, so this is a coverage
gap and not a general one:

| My mutant | Narrowed to admit | Result |
|---|---|---|
| `resolveOverlay` path-form check drops only `Revision` | exactly a `revision` on a path source | **killed** by `TestPathOverlayFormIsSourceInvalid` (`err = <nil>`) |
| `materializeOne` unmanaged-conflict gate `&& path != "AGENTS.md"` | exactly `AGENTS.md` | **killed** by `TestUsePartialRecordsNothing` |
| `parseSystemEnvironments` allowed-knob gate `&& key != "forms"` | exactly `forms` | **SURVIVED** |

### m2 — a ledger read failure is treated as "nothing ledgered"

`readSkillsLedger` (`internal/envprofile/import.go:290`) returns an empty recorded
set when `.csk-managed.json` cannot be read or does not decode. §9.6 excludes ledgered
entries because they belong to the machine-global scope and reach managed state through
§9.4. On a failed read the manager cannot prove an entry is unledgered, and importing
it anyway is a fallback defined for absence firing on a read failure (§8.4).

### m3 — doubled diagnostic prefixes

`profile_source_invalid: profile_source_invalid: context/link.md is not a directory or
a regular file`, and likewise for hard links, fifos and nested `.git`. Cosmetic, same
root cause as M2's wrapper.

### m4 — `profile compose list` prints declarations a locked `overlays_allowed: false` has emptied

With the key locked false, resolution correctly drops every overlay (verified: the
materialized header loses the member) and the manager §1 warning fires — but
`compose list` still prints the declared row with no indication it is inert. The row is
documented as informative file editing, so this is a judgement call, not a spec break.

---

## What was verified clean

Driven at the production entry point and correct:

- Every §12.1 knob present with its exact name, value grammar and default (18
  top-level names / 20 table rows); unknown field, bad `precedence.winner`,
  `xdg_seed_allowlist` `opencode`, and a `sha256:`-prefixed waiver pin all refused.
- Schema-1 file stays valid; a schema-1 file carrying `environments` refused;
  `schema_version` 0/3/99 refused explicitly; explicit `environments: null` refused;
  `environments: {}` accepted.
- The §12.2 lockable subset: locking `environments.forms`, `environments.backup_retention`,
  bare `environments`, an env key under system schema 1, and a locked-but-unset key are
  each refused with a distinct message; a system file carrying a non-lockable knob is
  refused; `isolation` toward `isolated` refused, toward `shared` accepted; system
  `schema_version` 3 refused.
- Locked `overlays_allowed: false` empties the overlay list at resolution with the
  manager §1 warning; the header loses the member.
- Precedence flows config → materialized bytes (`precedence: winner=lower-weight
  placement=winner-first` in the emitted header), and
  `ascending := (Winner==higher) == (Placement==last)` is the correct truth table for
  all four pairs; `sort.SliceStable` keeps ties in topological order.
- The four effective-weight rules in order, with **all four disagreeing** in one
  closure (manifest 5, non-root edge 60, root map 10, overlay 7000): rule 4 wins with
  7000, and with the overlay removed the root map's 10 wins. `context_weights_duplicate`,
  `context_weight_unknown`, `context_weights_not_root` and the
  `context_weight_conflict` error/warning split read correctly in
  `internal/contextresolve/contextresolve.go:709-812`.
- Overlays join the closure flagged `overlay` at the machine default weight (git
  overlay driven end to end through the CLI); a repeat of another overlay's name and a
  repeat of the root's name are both `environment_composition_invalid`;
  `profile remove` of an overlay member is `profile_in_use`.
- `path` kind: symlink, hard link, fifo, nested `.git` and non-directory are all
  `profile_source_invalid`; a root-level `.git` is excluded from the snapshot (verified
  in the store); a missing operand is `profile_source_path_missing`; the snapshot is
  genuinely immutable (mutating the source then `profile sync` changes nothing);
  a requirement flag or `--directory` with a path operand is
  `profile_install_ref_conflict` per §9.7.
- Onboarding: the foreign-manager stop is a real stop with the explicit choice;
  `profile list`, `env status` and `env resolve` without `--repair` never onboard,
  never write a backup and never prompt (native home byte-identical afterwards); the
  takeover backup lands before the first write and preserves the replaced bytes.
- Import: absent surfaces are never losses; an unreadable file and invalid UTF-8 are
  both losses and stop with `environment_import_lossy` and the loss list;
  `--allow-lossy` proceeds and re-reports as warnings; no configuration knob feeds
  `AllowLossy` or `Policy.Takeover`; reassembly emits `name` `imported`,
  `version` `1.0.0`, `weight` `0`, no `weights`, modules in ascending environment
  order with `class: root` and the single-element `environments` selector; the import
  writes nothing into any native home by itself (verified with a current profile
  already set and no `--use`); `profile list` records `imported`.
- `env config show|set|unset` round-trips one knob by its §12.1 table name, validates
  before writing (an invalid `precedence.winner`, and a formless path overlay under
  `overlays`, are both refused with the machine file left byte-identical), refuses a
  locked knob with the system-file warning, and `unset` restores the table default.
- The Windows-fixture lesson is honoured: `gitFileURL`/`gitFileURLFor` is the single
  insteadOf seam and carries its own narrowing mutant test
  (`TestGitFileURLForKeepsVolumesOutOfTheHost`), and `pinOperatorHome` sets both
  `HOME` and `USERPROFILE`.
- `internal/config` is registered in `.github/ci/root-artifacts.tsv` with all three new
  artifacts, and 57 ledger rows were added to `.github/ci/platform-cases.tsv`.

## Rework guidance

`repeat-of: none` — this is the first review cycle for this element.

Order the fixes by blast radius: B2 and B3 corrupt operator data, B1 defeats the only
fleet-policy surface revision 1 has, B4 loses an operator declaration silently. M1 and
M3 are as much report-accuracy defects as code defects — the next report must not
list an unreachable surface or an unimplemented §12 sentence as driven. Each new gate
needs a narrowing mutant that keeps the gate present, and each fix needs a test that
asserts the *leading* diagnostic and the *post-state* (is the surface still a symlink?
did the foreign bytes survive?), not `strings.Contains` and `result.OK`.
