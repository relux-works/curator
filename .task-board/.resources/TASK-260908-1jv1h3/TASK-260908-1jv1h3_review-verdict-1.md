# TASK-260908-1jv1h3 — recorded review verdict (cycle 1)

**VERDICT: ACCEPT.** No blocking and no major finding.

Subject: `chore/rc-stays-out-of-install-channels` @ `d1bb0a4e`, base `04550e28`
(= current `origin/main`), PR https://github.com/relux-works/curator/pull/65.
One commit, signature `G`, author `Ivan Oparin <oparin@me.com>`. Diff is
`.goreleaser.yml` only, **+10/−0** (the brief says +8; the actual count is 10 —
3 fields plus 7 comment lines. Brief arithmetic, not a defect in the change).

This run is the recovery successor of `RUN-260908-326a7c`. That run produced a
complete cycle-1 analysis (`TASK-260908-1jv1h3_review-findings-1.md`) but left
the task at `to-review`, which is not one of the four verdict branches, so no
verdict was ever recorded. This document **records the branch** and reports the
independent re-verification this run performed. Findings F1–F5 of the cycle-1
document stand as written; §2 below is new evidence, not a restatement.

---

## 1. Version boundary — the workflow pins nothing

`.github/workflows/release.yml:37-39` uses `goreleaser/goreleaser-action@v6`
with `version: "~> v2"`. That is a **range**, not a pin: it resolves to the
newest stable v2.x at run time. Naming "the pinned GoReleaser version" is
therefore not possible; what is possible is bounding the range at both ends that
matter.

| Point of the range | How established | The four decision lines |
| --- | --- | --- |
| **2.17.1** | What actually ran the incident tag. Measured from the hosted log of run `31059325487`: `Downloading .../goreleaser/releases/download/v2.17.1/goreleaser_Darwin_all.tar.gz` | `git diff v2.17.0 v2.17.1` over `pipe/{release,cask,scoop}` + `client/github.go`: **only** `internal/client/github.go`, and only a `go-github` v88→v89 SDK type migration. Semantics identical. |
| **2.17.0** | Local binary, `GitCommit 770a4fc7` — byte-identical to the cloned tag, so the source read and the executed binary are one artifact | The lines quoted in §2 |
| **2.18.1** | Newest stable v2 today (`gh api repos/goreleaser/goreleaser/releases`) — what a tag cut now would run | `cask.go:158`, `scoop.go:218`, `release.go:75-83` — character-identical predicates; `config.go` still carries `skip_upload` on both structs |

Bound: a v2.19+ released between now and the rc1 tag is not covered by this
review.

## 2. `auto` semantics — driven end to end, not read

The cycle-1 document proved the semantics by source read plus Go probes into
`runAll`/`publishAll`. This run raised the bar: a **real `goreleaser release`
against real install channels**.

Fixture: a throwaway Go module tagged as curator, with `homebrew_casks[0]` and
`scoops[0]` pointed at two **local bare git repos** via `repository.git.url`.
That reaches the production publish path — `Pipe.Publish` → `publishAll` →
`doPublish` → `client.NewGitUploadClient(...).CreateFile` → real `git push` —
so the observable is not a log line or a return value but **whether a commit
lands in the install channel**.

| Case | `skip_upload` | tag | `goreleaser check` | commit landed in tap/bucket |
| --- | --- | --- | ---: | --- |
| **A** shipped config | `"auto"` | `v0.15.0-rc.1` | exit 0 | **NO** — `prerelease detected with 'auto' upload, skipping homebrew publish` / `release is prerelease` |
| **B** shipped config | `"auto"` | `v0.15.0` | exit 0 | **YES** — `Brew cask update for curator version v0.15.0`, `Scoop update for …` |
| **C** mutant: field removed (the pre-change state) | *absent* | `v0.15.0-rc.1` | exit 0 | **YES** — `Brew cask update for curator version v0.15.0-rc.1` |
| **D** mutant: gate **narrowed**, not deleted | `"sometimes"` | `v0.15.0-rc.1` | **exit 0** | **YES** |

Row A is the claim. Row B is the mutant that would fail if the guard were
unconditional — it proves the guard is prerelease-*conditional* and that stable
releases still reach the channels. Row C is the incident, reproduced
executably. Row D is the narrowing mutant the evidence contract asks for: it is
not a deletion, it keeps a plausible value in place, and it silently republishes
the candidate.

Raw log: `TASK-260908-1jv1h3_live-differential.log`.
Driver: `TASK-260908-1jv1h3_live-differential-driver.sh.txt`.

Code the runs exercise (2.17.0, identical at 2.18.1):

- `internal/pipe/cask/cask.go:147` — `TrimSpace(brew.SkipUpload)=="auto" && ctx.Semver.Prerelease!=""` → `pipe.Skip`
- `internal/pipe/scoop/scoop.go:216` — same predicate
- `internal/pipe/release/release.go:74-83` — `case "auto": if ctx.Semver.Prerelease != "" { ctx.PreRelease = true }`
- `internal/pipe/semver/semver.go:24-38` — `ctx.Semver.Prerelease` is the SemVer suffix from `Masterminds/semver.NewVersion(tag)`

Two consequences worth stating because the comments do not:

- **`prerelease: "auto"` also demotes "Latest".** `internal/client/github.go:584-592`:
  `if ctx.PreRelease { latest = "false" }`. `install.sh:29` resolves the
  version through `api.github.com/repos/$REPO/releases/latest`, which excludes
  prereleases — so the curl-installer is a **third** install channel this leaf
  protects, and the comments do not name it.
- **Fails closed on a non-SemVer tag.** `semver.go:29-32` returns a hard error
  unless `--skip=validate`; the workflow passes `release --clean` with no skip.
  A tag like `v0.15.0rc1` aborts the release rather than silently bypassing the
  guard. `auto` cannot be defeated by tag spelling.

## 3. Comment truth — every clause measured

| Clause | Evidence |
| --- | --- |
| "v0.14.0-rc.1 reached the tap because this was unset" | `relux-works/homebrew-tap` commit `3bbc3ada` 2026-08-06T00:22:42Z `Brew cask update for curator version v0.14.0-rc.1`; it stood as the served cask until `951cde1e` (v0.14.0) on 2026-09-06 — **a month** |
| …and the mechanism is the unset field, not a neighbouring default | `git show v0.14.0-rc.1:.goreleaser.yml` has no `skip_upload` and no `prerelease`; no `Default()` in the cask/scoop pipes assigns `SkipUpload`; case **C** above reproduces it |
| Scoop was hit too ("same reason as the cask above") | `relux-works/scoop-bucket` commit `973742e6` 2026-08-06T00:22:44Z |
| The hosted run really did push both | Hosted log `31059325487`: `pushing repository=relux-works/homebrew-tap` and `pushing repository=relux-works/scoop-bucket` |

**Correction to one clause (F1, minor, non-blocking):** the cask comment says
"must not replace the stable **formula**". The tap ships `Casks/curator.rb`, a
cask, not a formula. Wording only.

## 4. Nothing else in the release pipeline changes

One file. Parsed-YAML leaf diff base→head, run in this cycle: added
`/homebrew_casks[0]/skip_upload=auto`, `/scoops[0]/skip_upload=auto`,
`/release/prerelease=auto`; **0 removed, 0 changed**. No
workflow, tap/bucket, footer, template, signing, SBOM, notarization or nfpm
behaviour moves. `goreleaser check` exit 0 on both base and head.

Completeness of the fix: the config declares exactly three publishing surfaces
— `homebrew_casks`, `scoops`, and the GitHub release itself. `nfpms` builds
deb/rpm but no apt/yum publisher exists, so those ride the release. There is no
`aur`, `nix`, `winget`, `chocolatey` or `dockers` stanza. **All install channels
present are covered by the three fields.**

## 5. Hosted lanes on `d1bb0a4e`

`gh api repos/relux-works/curator/commits/d1bb0a4e…/check-runs` — 11 lanes,
**11 success, 0 red**: Lint, Naming gate, Interop conformance gate, Gate
self-test ×3, Test ×3, Race ×2. `Candidate suite (${{ matrix.os }})` reads
`skipped`, which is correct by design, not a casualty: `ci.yml:322-326` gates it
on `workflow_dispatch` with a `candidate_ref`/`candidate_root` input, and this
is a `pull_request` event.

## 6. Confidence boundary and first real proof

The semantics half is **no longer docs-confidence**. It is source-and-execution
confidence at two points of the `~> v2` range, with the pre-change state
reproduced and a narrowing mutant that publishes.

What remains genuinely unprovable without a real write, and is therefore
reported as unknown rather than inferred:

1. Which v2.x the hosted runner resolves at tag time.
2. The GitHub API leg of `prerelease: "auto"` — `ctx.PreRelease` → the release
   object's `prerelease` / `make_latest` fields. Everything upstream of that
   call is proven (`ctx.Semver.Prerelease` is the very field cases A–D
   discriminate on); only the API write is untested.
3. Interaction of the skip with real tap tokens, artifact upload, cosign
   signing and notarization in the hosted job.

**First real proof is the rc1 tag itself.** On `v0.15.0-rc.1`, check exactly:
the GitHub release reads `prerelease=true` and is **not** marked Latest; and
`relux-works/homebrew-tap` and `relux-works/scoop-bucket` gain **no** commit.

## 7. Residual risk, recorded not requested

`goreleaser check` returns **exit 0** on `skip_upload: "sometimes"` (case D) —
it validates field name and type, never the value. No in-repo gate reads
`.goreleaser.yml` at all: measured coverage **0 of 3** fields. So a future typo
in any of these three literals silently restores the August incident with every
hosted lane green.

This is **not** requested as rework: this leaf's AC forbids touching anything
else in the release pipeline, and the shipped literals are correct by grep and
by cases A/B. Recommend a follow-up leaf for a config gate asserting the three
literals, using case D as its negative test.

## 8. One unresolved observation, reported as inference

`v0.14.0-rc.1` currently reads `prerelease=true` on GitHub, which looks like it
contradicts the third comment's premise. It does not: `release.go:74-83` has no
default branch, `github.go:517-541` sends `Prerelease: &ctx.PreRelease` = false
when the key is unset, and only `release.go:77/81` ever write that field. So
GoReleaser did not set it. The release's `author` is `github-actions[bot]`, but
that records the creator, not the last editor, and GitHub exposes no release
edit history. **Most likely hand-edited after publication** — stated as a
deduction with its bound, not as a read. It does not weaken the change: the
field makes the marking deterministic from config instead of dependent on a
human noticing.

---

**repeat-of:** none

**Branch:** ACCEPT → `done`.

*Routing note:* the brief's verdict contract asked for "an explicit ACCEPT at
`to-review` … Do not mark the task done", with `accept_cr` on a recorded
revision. Neither is available here. No `Change Request Under Review` section
was handed to this run and the element carries no CR revision (`get(…){full}`
shows no CR field, `integrationCheckpointed: false`), so `accept_cr` would be
refused as unauthorized. And `to-review` is not a verdict branch — following
that clause literally is precisely why `RUN-260908-326a7c` was recorded
unsatisfied and this successor was queued. The role contract's four branches
govern: accepted → `done`.
