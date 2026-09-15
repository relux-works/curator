# Reviewer verdict — cycle 3 (RUN-260908-df630c, recovery attempt 3/3)

**VERDICT: ACCEPT.** Subject `chore/rc-stays-out-of-install-channels` @ `d1bb0a4e`,
base `04550e28`, PR #65. Agrees with cycles 1 and 2, re-derived independently —
every claim below was measured in this run, not read from a prior verdict.

`repeat-of: none`

## 0. What this run adds

Cycles 1 and 2 reached ACCEPT and attached the analysis. This run re-derived the
load-bearing claims from scratch and adds two things neither prior cycle had:

1. **A channel-coverage enumeration from the parsed head config** (§3), which is
   the question "do three fields actually cover every install channel?" answered
   structurally rather than by inspection.
2. **A live differential whose broken states are on the record** (§2). My first
   three fixture attempts reported "NO COMMIT" for every case including the
   controls — a fixture that fails before the publish leg produces the same
   observable as a working guard. That is the failed-read-as-absence shape, and
   it is why a control that *must* publish is not optional here.

## 1. Semantics — read from source at the version a tag would actually run

The DoD says "the pinned GoReleaser version". **There is no pin.**
`.github/workflows/release.yml:37-39` is `goreleaser/goreleaser-action@v6` with
`version: "~> v2"` — a range resolving to the newest stable v2 at run time.
Measured today: newest stable v2 is **v2.18.1** (2026-09-05). `v2.19.0-*` exist
upstream but are prereleases, which `~> v2` does not select.

Read at v2.18.1 (raw source, this run):

| Field | Source | Predicate |
|---|---|---|
| cask `skip_upload` | `internal/pipe/cask/cask.go:158` | `strings.TrimSpace(SkipUpload) == "auto" && ctx.Semver.Prerelease != ""` → `pipe.Skip` |
| scoop `skip_upload` | `internal/pipe/scoop/scoop.go:218` | same predicate → `pipe.Skip` |
| `release.prerelease` | `internal/pipe/release/release.go:75-79` | `case "auto":` → `ctx.PreRelease = true` when `ctx.Semver.Prerelease != ""` |
| tag → prerelease | `internal/pipe/semver/semver.go:24-39` | `ctx.Semver.Prerelease = sv.Prerelease()` |

So `auto` means skip-on-prerelease on all three, and is a no-op on a stable tag.
The three comments are accurate on this point.

**Third channel the comments do not name.** `internal/client/github.go:588-590`
sets `latest = "false"` whenever `ctx.PreRelease`, *unconditionally overriding*
the templated `make_latest`. `install.sh:29` resolves through
`api.github.com/.../releases/latest`, which excludes prereleases. The curl
installer is therefore protected too. Head config does not set `make_latest`.

**Fails closed on a malformed tag.** `semver.go:25-32` hard-errors unless
`--skip=validate`; the workflow passes `release --clean` with no skip. A tag like
`v0.15.0rc1` aborts the release rather than silently bypassing the guard.

## 2. Semantics — driven, not read

Real `goreleaser release` runs against two local bare git repos wired as the tap
and the bucket through `repository.git.url` + `private_key`, so the production
path `Pipe.Publish → publishAll → doPublish → NewGitUploadClient.CreateFile`
executes. **Observable = a landed commit**, not a log line. Binary: 2.17.0
(GitCommit `770a4fc7`); the four decision lines above are character-identical
between 2.17.0 and 2.18.1.

| Case | `skip_upload` | tag | tap | bucket | `goreleaser check` |
|---|---|---|---:|---:|---:|
| **A** control (the change) | `"auto"` | `v0.15.0-rc.1` | **NO COMMIT** | **NO COMMIT** | exit 0 |
| **B** control | `"auto"` | `v0.15.0` | COMMIT LANDED | COMMIT LANDED | exit 0 |
| **C** pre-change state | *absent* | `v0.15.0-rc.1` | COMMIT LANDED | COMMIT LANDED | exit 0 |
| **E** narrowing mutant | `"Auto"` | `v0.15.0-rc.1` | COMMIT LANDED | COMMIT LANDED | exit 0 |

- **B is the control that makes A mean anything.** The guard is conditional, not
  a kill switch — stable releases still publish. Without B, row A is
  indistinguishable from a fixture that never reached the publish leg (which is
  exactly what my first three attempts were).
- **C reproduces the August incident executably.** The commit that landed serves
  `version "0.15.0-rc.1"` under the message
  `Brew cask update for probe version v0.15.0-rc.1` — the same template as the
  real tap commit `3bbc3ada`.
- **E is a narrowing mutant, not a deletion.** Capitalising one letter defeats
  the gate and `goreleaser check` still exits 0.

Case A's log carries the exact upstream skip reasons:
`prerelease detected with 'auto' upload, skipping homebrew publish` and
`release is prerelease`.

## 3. Nothing else changes — and the three fields cover everything

Diff `04550e28..d1bb0a4e`: **1 file, +10/−0** (the brief says +8; the actual
count including comments is 10). Parsed-YAML leaf diff base→head: **3 keys
added, 0 removed, 0 changed**, every value the literal `auto`.

Top-level stanzas in the head config, enumerated from the parse:
`version, project_name, builds, archives, checksum, sboms, notarize, signs,
nfpms, homebrew_casks, scoops, release, changelog`.

The only publishing surfaces are `homebrew_casks`, `scoops`, `release`. There is
no `brews`, `aurs`, `nix`, `winget`, `chocolateys`, `dockers`, `krews`,
`publishers`, `blobs`, `uploads` or `announce` stanza. `nfpms` builds `.deb`/
`.rpm` artifacts but publishes them only as GitHub release assets, which the
`release.prerelease` + forced `make_latest=false` pair already covers. **Three
fields is complete coverage of the channels this config declares** — that is a
structural result, not an inspection.

## 4. Comment truth

- Tap `3bbc3ada` (2026-08-06) `Brew cask update for curator version v0.14.0-rc.1`
  stood as the served cask for a month until `951cde1e` (v0.14.0, 2026-09-06).
  Bucket `973742e6` the same minute. The incident is real and measured.
- `git show v0.14.0-rc.1:.goreleaser.yml` has **neither** key.
- No `SkipUpload =` assignment exists in cask or scoop defaults, so unset stays
  `""`, matching neither the `"true"` nor the `"auto"` branch → publish proceeds.
  **Unset is the mechanism, not a neighbouring default.** Row C reproduces it.

**F1 (minor, wording).** The cask comment says "must not replace the stable
**formula**", but `directory: Casks` and the tap ships `Casks/curator.rb` — it is
a cask, not a formula. Cosmetic; does not affect behaviour.

## 5. Coverage of the change itself — measured as a ratio

- `goreleaser check` discriminates **0 of 6** wrong-value mutants run here:
  `auto`, `Auto`, `sometimes`, `true`, field-removed, and `prerelease: "ato"` all
  exit 0. The struct tag is `jsonschema:"oneof_type=string;boolean"` — it
  validates field *name* and *type*, never the *value*. (It does catch an unknown
  field name; that is what it is for.)
- **No in-repo gate reads `.goreleaser.yml` at all.** `git grep` for the filename
  outside the file itself returns only board resources. The one goreleaser-aware
  gate, `release-workflow-gate.sh`, asserts *ordering inside `release.yml`* and
  never opens the config. **Coverage of the three fields: 0 of 3.**

**Consequence, recorded not requested:** a future typo silently restores the
August incident with every lane green. Not requested as rework — this leaf's AC
forbids touching anything else. Recommend a follow-up leaf for a config gate,
with two constraints learned here: assert on the **parsed value including case**
(row E defeats a gate that merely greps for the token `auto`), and use row C as
its negative test.

## 6. Hosted lanes on `d1bb0a4e`

`total_count 12` — **11 success, 0 failure, 0 pending, 1 skipped**. The skip is
`Candidate suite`, correct by design: `ci.yml:324-326` gates it on
`github.event_name == 'workflow_dispatch'` with candidate inputs, and this is a
`pull_request`. Not a rebase casualty. PR #65 is OPEN / MERGEABLE, head matches
the reviewed commit, 1 ahead / 0 behind `origin/main`, signed `G` by Ivan Oparin.

## 7. Confidence boundary — what is proven and what is not

The semantics half is **no longer docs-confidence**. It is source-plus-execution
confidence: the predicate read at v2.18.1, and the behaviour driven end to end at
2.17.0 with the pre-change state reproduced and two controls that publish.

Reported as **unknown**, not inferred:

1. **Which v2.x the runner resolves at tag time.** `~> v2` is a range; bounded at
   2.17.0/2.17.1/2.18.1, uncovered for any v2.19+ released before the rc1 tag.
2. **The GitHub API leg of `prerelease: "auto"`.** The fixture runs with
   `release: disable: true`, so it exercises the cask/scoop skip only. Everything
   upstream of the API write is proven (`ctx.PreRelease` is set, and
   `github.go:540` always sends `Prerelease: &ctx.PreRelease`); the write itself
   is not reproducible without a real GitHub release.
3. **Interaction with real tokens, artifact upload, cosign and notarization.**

**FIRST REAL PROOF IS THE rc1 TAG ITSELF.** After `v0.15.0-rc.1` is cut, the
GitHub release must read `prerelease=true` and not-latest, and
`relux-works/homebrew-tap` and `relux-works/scoop-bucket` must gain no commit.
Until then item 2 stays unproven by construction, which is why this leaf is
accepted on the evidence it *can* carry, not on a claim it cannot.

## 8. Routing — `accept_cr` is unreachable, proven from both sides

My prompt carries **no `Change Request Under Review` section**, so the role
contract's default branch governs: accepted → `done`. Probed anyway, because
authorization is run-bound and cannot be inherited from cycle 2:

- `accept_cr(revision=1)` → `change_request_acceptance_unauthorized: run
  RUN-260908-df630c ... was handed Change Request revision 0`
- `accept_cr(revision=0)` → `revision must be a positive integer`

No reachable revision exists. This is structural, not a missed step: the reviewed
work lives in a **different repository** (`relux-works/curator`) and was produced
outside this Story's managed worktree, so no producer candidate was ever
snapshotted and no Change Request record exists for the element.
`set_status(done)` succeeds (idempotent, `done`→`done`), so the accepted verdict
branch **is** recorded on the board.

Open items belong to the orchestrator, not to another reviewer cycle: land
PR #65, and fix the runner's reviewer-satisfaction predicate, which demands a
Change Request for a leaf whose subject repository is not the Story workspace.
A fourth reviewer attempt hits the identical refusal.
