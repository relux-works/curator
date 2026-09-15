# TASK-260908-1jv1h3 — reviewer verdict, cycle 2 (independent re-verification)

**VERDICT: ACCEPT.** No blocking and no major finding. The branch already
carries this verdict (`done`, recorded by `RUN-260908-41a57e`); this document is
`RUN-260908-81cf84`'s own attestation, produced by re-running the load-bearing
evidence rather than reading the cycle-1 report, plus two new narrowing mutants
cycle 1 did not cover and one correction to a cycle-1 claim.

Subject: `chore/rc-stays-out-of-install-channels` @ `d1bb0a4e92304526c782f77607aa30abc2f3ae07`,
base `04550e28` (= current `origin/main`), PR
https://github.com/relux-works/curator/pull/65 (OPEN, MERGEABLE, head matches).
One commit, signature `G`, author `Ivan Oparin <oparin@me.com>`, 1 ahead / 0
behind `origin/main`. Diff: `.goreleaser.yml` only, **+10/−0**.

---

## 0. What this run re-measured itself

Everything below was executed or read in this run. Where I reached the same
conclusion as cycle 1, that is stated as agreement; where I did not, §3 says so.

| Claim | How I established it here |
| --- | --- |
| Diff shape | `git diff --stat 04550e28..d1bb0a4e` → 1 file, +10/−0 |
| Semantics of `auto` | source at **v2.17.0** and **v2.18.1** fetched fresh from `goreleaser/goreleaser` and diffed at the four decision sites |
| Semantics executed | my own **live differential** (§3) against real local install channels, control + 2 new mutants |
| Tag → `Prerelease` mapping | Go probe against **Masterminds/semver v3.5.0**, the exact version goreleaser v2.17.0 *and* v2.18.1 vendor (`go.mod` read at both tags) |
| `goreleaser check` discrimination | **6** config variants run through the local `goreleaser check` (§4) |
| "Nothing else changes" | parsed-YAML leaf diff, base vs head, computed here |
| Comment truth | tap/bucket commit history + `git show v0.14.0-rc.1:.goreleaser.yml` + `Default()` read |
| Hosted lanes | `gh api .../commits/d1bb0a4e.../check-runs` |

## 1. Version boundary — the DoD's "pinned GoReleaser version" does not exist

`.github/workflows/release.yml:37-39`:

```yaml
        uses: goreleaser/goreleaser-action@v6
        with:
          version: "~> v2"
          args: release --clean
```

`~> v2` is a **range**, resolved to the newest stable v2.x at run time. The DoD
row "verified against the pinned GoReleaser version" is not literally
satisfiable; what is satisfiable is bounding the range. Reported as a boundary,
not inferred away.

- **v2.17.0** — the local binary (`GitVersion 2.17.0`, `GitCommit 770a4fc7`),
  the one that actually **executed** every case in §3 and §4.
- **v2.18.1** — newest stable v2 today (`gh api …/releases/latest`), i.e. what a
  tag cut right now would run. Source read this cycle.
- Cycle 1 additionally pinned **v2.17.1** from the hosted log of the incident
  run `31059325487`. I did not re-read that log; accepted from cycle-1 evidence
  and labelled as such.

The four decision sites, **character-identical at v2.17.0 and v2.18.1**:

| Site | v2.17.0 | v2.18.1 | Predicate |
| --- | ---: | ---: | --- |
| `internal/pipe/cask/cask.go` | `:147` | `:158` | `strings.TrimSpace(brew.SkipUpload) == "auto" && ctx.Semver.Prerelease != ""` → `pipe.Skip` |
| `internal/pipe/scoop/scoop.go` | `:216` | `:218` | same predicate → `pipe.Skip` |
| `internal/pipe/release/release.go` | `:74-83` | `:75-84` | `switch Release.Prerelease { case "auto": if ctx.Semver.Prerelease != "" { ctx.PreRelease = true } }` |
| `internal/pipe/semver/semver.go` | `:24-38` | identical file | `ctx.Semver.Prerelease = sv.Prerelease()` from `Masterminds/semver.NewVersion(tag)` |

`diff` of `cask.go`/`scoop.go`/`release.go` across the two versions produces
**no** line touching `SkipUpload`, `Prerelease`, `Semver`, `doPublish` or
`pipe.Skip` — the single near-hit is `skips := pipe.SkipMemento{}` gaining a
second call site in `runAll`, which does not change what `publishAll` does with
a `pipe.Skip` return.

**So: `auto` means skip-on-prerelease, and on a stable tag the conjunct
`ctx.Semver.Prerelease != ""` is false and both channels publish normally.**
Confirmed by source at two points and by execution (§3 rows A2/B).

**Stated bound:** a v2.19+ stable released between now and the rc1 tag is not
covered. `v2.19.0-*-nightly` tags already exist upstream.

## 2. The third channel the comments do not name

`internal/client/github.go:589-590` (v2.17.0; `:588-589` at v2.18.1):

```go
	if ctx.PreRelease {
		latest = "false"
	}
```

and `install.sh:29` resolves the version through
`https://api.github.com/repos/$REPO/releases/latest`, which excludes
prereleases. So `release.prerelease: "auto"` protects the **curl installer** as
well as the releases page — a third install channel. Not a defect; worth
recording because the comment names only the releases page.

Also read: `CreateRelease` sends `Prerelease: &ctx.PreRelease` **unconditionally**
(`github.go:539`), so with the key unset the API receives an explicit `false`.
That backs cycle-1 §8's reading of the `v0.14.0-rc.1` release object.

## 3. `auto` attacked, not read — live differential, cycle-2 extension

Same fixture shape as cycle 1: a throwaway Go module whose `homebrew_casks[0]`
and `scoops[0]` point at two **real local bare git repos** via
`repository.git.url`. That reaches the production path
`Pipe.Publish → publishAll → doPublish → client.NewGitUploadClient(...).CreateFile`
→ real `git push`. The observable is **a commit landing in the install channel**,
not a log line. Each case resets both bare repos to their root commit first.

| Case | `skip_upload` | tag | `goreleaser check` | commit landed in tap+bucket | |
| --- | --- | --- | ---: | --- | --- |
| **A2** control, shipped config | `"auto"` | `v0.15.0-rc.1` | exit 0 | **NO** | harness discriminates |
| **E** mutant: literal **narrowed** by case | `"Auto"` | `v0.15.0-rc.1` | exit 0 | **YES** | gate is case-sensitive |
| **F** mutant: rc spelled as **build metadata** | `"auto"` | `v0.15.0+rc1` | exit 0 | **YES** | `Prerelease == ""`, gate never fires |

Cycle-1's rows A/B/C/D (shipped+rc → NO; shipped+stable → YES; field removed+rc
→ YES, the incident reproduced; `"sometimes"`+rc → YES) are accepted from
`TASK-260908-1jv1h3_live-differential.log`; I re-read that log, its skip reasons
are byte-matches of the source strings I read at v2.17.0
(`prerelease detected with 'auto' upload, skipping homebrew publish` /
`release is prerelease`), and its driver resets the channels per case. My row A2
reproduces its row A independently.

Neither E nor F is a deletion. Both keep a plausible value or a plausible tag in
place, both pass every check the repository can run, and both put a release
candidate into the tap.

Raw: `TASK-260908-1jv1h3_cycle2-live-differential.log`,
driver `TASK-260908-1jv1h3_cycle2-live-driver.sh.txt`.

### Tag → gate mapping, measured on the vendored semver

Go probe against `Masterminds/semver/v3 **v3.5.0**` (the version both v2.17.0
and v2.18.1 require in `go.mod`):

| tag | parses | `Prerelease()` | gate fires |
| --- | --- | --- | --- |
| `v0.15.0-rc.1` ← **the target tag** | yes | `rc.1` | **yes** |
| `v0.15.0-rc1` | yes | `rc1` | yes |
| `v0.15.0-rc.1+build5` | yes | `rc.1` | yes |
| `v0.15.0` | yes | `""` | no (correct — stable must publish) |
| `v0.15.0+rc1` | yes | `""` | **no ← bypass** |
| `v0.15.0rc1`, `v0.15.0.rc1`, `v0.15.0-` | **no** | – | release **aborts** (`semver.go:29-32`, no `--skip` in the workflow) |

## 4. `goreleaser check` discriminates nothing about the value

Six variants of the **head** config through the local `goreleaser check`:

| variant | `goreleaser check` |
| --- | ---: |
| shipped (`"auto"` ×2, `prerelease: "auto"`) | exit **0** |
| `skip_upload: "sometimes"` | exit **0** |
| `skip_upload: "Auto"` | exit **0** |
| `prerelease: "ato"` | exit **0** |
| all three fields removed (the pre-change state) | exit **0** |
| `skip_upload: "true"` | exit **0** |

`check` validates field name and type, never the value: **0 of 4** wrong-value
mutants discriminated.

And **no in-repo gate reads `.goreleaser.yml` at all.** The only goreleaser-aware
gate is `.github/ci/release-workflow-gate.sh:37`, which greps
`.github/workflows/release.yml` for `uses: goreleaser/goreleaser-action@`.
Measured coverage of the three fields under review: **0 of 3.**

The shipped values are nevertheless correct: the parsed-YAML leaf values are the
Go strings `"auto"`, `"auto"`, `"auto"` — exact matches for the case-sensitive
comparison — and row A2 proves them at runtime.

## 5. Nothing else in the release pipeline changes

Parsed-YAML leaf diff, base → head, computed this cycle:

```
ADDED  : {"/homebrew_casks[0]/skip_upload": "auto",
          "/scoops[0]/skip_upload": "auto",
          "/release/prerelease": "auto"}
REMOVED: {}   CHANGED: {}
```

One file. No workflow, tap/bucket repo, footer, template, signing, SBOM,
notarization or nfpm behaviour moves. The config declares exactly three
publishing surfaces — `homebrew_casks`, `scoops`, `release` — and carries no
`aur`, `nix`, `winget`, `chocolatey` or `dockers` stanza; `nfpms` builds
deb/rpm but has no publisher, so those ride the release. **Every install
channel present is covered.** `release:` is not disabled in the real config
(`draft: false`, no `disable:`), unlike the §3 fixture.

## 6. Comment truth

| Clause | Measured |
| --- | --- |
| "v0.14.0-rc.1 reached the tap" | `relux-works/homebrew-tap` `3bbc3ada` 2026-08-06T00:22:42Z *Brew cask update for curator version v0.14.0-rc.1*; it stayed the tip — the served cask — until `951cde1e` (v0.14.0) on 2026-09-06. **A month.** |
| "Same reason as the cask above" (scoop) | `relux-works/scoop-bucket` `973742e6` 2026-08-06T00:22:44Z |
| "…because this was unset" — mechanism, not a neighbouring default | `git show v0.14.0-rc.1:.goreleaser.yml` has **no** `skip_upload` and **no** `prerelease` key; no `Default()` in cask/scoop assigns `SkipUpload` (`grep 'SkipUpload *='` over both versions → **no assignment anywhere**), so unset ⇒ `""` ⇒ neither branch fires ⇒ publish. Cycle-1 row C reproduces it executably. |
| The commit really is GoReleaser's | observed message is a character match for the default `CommitMessageTemplate` `Brew cask update for {{ .ProjectName }} version {{ .Tag }}` (`cask.go` `Default()`). |
| "must not replace the stable **formula**" | **F1, minor:** the tap ships `Casks/curator.rb` — a **cask**. `gh api repos/relux-works/homebrew-tap/contents/Casks` → `curator.rb`. Wording only. |

## 7. Hosted lanes on `d1bb0a4e`

`gh api repos/relux-works/curator/commits/d1bb0a4e…/check-runs` → `total_count: 12`:
**11 success, 0 failure, 0 pending**, plus 1 `skipped`.

Success: Lint, Naming gate, Interop conformance gate, Gate self-test ×3
(ubuntu/macos/windows), Test ×3, Race ×2.
Skipped: `Candidate suite (${{ matrix.os }})`, correct **by design** —
`ci.yml:322-326` gates it on
`github.event_name == 'workflow_dispatch' && (inputs.candidate_ref != '' || inputs.candidate_root != '')`,
and this is a `pull_request` event. Not a rebase casualty.

(Cycle 1 said "11 lanes"; the API reports 12 check runs, 11 of them successful.
Same facts, different denominator.)

## 8. Confidence boundary and the first real proof

The **semantics** half is no longer docs-confidence. It is source-plus-execution
confidence at two points of the `~> v2` range, with the pre-change state
reproduced and three narrowing mutants that publish.

Reported as **unknown**, not inferred from a proxy:

1. Which v2.x the hosted runner resolves at tag time (`~> v2` is a range; v2.19
   nightlies exist).
2. The **GitHub API leg** of `prerelease: "auto"` — `ctx.PreRelease` → the
   release object's `prerelease` and `make_latest` fields. Everything upstream
   of that HTTP call is proven; the write itself is not, and the §3 fixture runs
   with `release: disable: true`, so it exercises the cask/scoop skip only, never
   the release pipe.
3. Interaction of the skip with real tap tokens, artifact upload, cosign signing
   and notarization in the hosted job.

**First real proof is the rc1 tag itself.** On `v0.15.0-rc.1`, check exactly:
the GitHub release reads `prerelease=true` and is **not** marked Latest; and
`relux-works/homebrew-tap` and `relux-works/scoop-bucket` gain **no** commit.

## 9. Findings

Cycle-1 findings F1–F5 stand as written. New or corrected here:

- **F1 (minor, wording, non-blocking).** "stable formula" → the tap ships a
  cask. Confirmed independently.
- **F2 / F4 (boundary, confirmed and widened).** `goreleaser check` discriminates
  **0 of 4** wrong-value mutants (§4) and no in-repo gate reads
  `.goreleaser.yml` at all — coverage **0 of 3** fields. A typo in any of the
  three literals silently restores the August incident with every hosted lane
  green. **Not requested as rework:** this leaf's AC forbids touching anything
  else in the release pipeline. Recommend a follow-up leaf for a config gate,
  with §3 rows E/F and §4 as its negative tests — and note that a gate which
  merely greps for the token `auto` would be defeated by mutant **E**, so the
  assertion must be on the parsed value, case included.
- **F6 (NEW — corrects a cycle-1 claim).** Cycle-1 §2 states *"`auto` cannot be
  defeated by tag spelling."* That is **too strong and does not reproduce.**
  `v0.15.0+rc1` is a legal git tag, parses as valid SemVer, yields
  `Prerelease() == ""`, and — measured, row **F** — publishes the cask and the
  scoop manifest for a release candidate. What *is* true is the narrower claim:
  a tag that is not valid SemVer at all aborts the release. This is a bound on
  the guard's reach, not a defect in the three fields: no value of
  `skip_upload` can catch `+rc1`, because the predicate lives inside GoReleaser.
  It does not change the verdict — the target tag is `v0.15.0-rc.1`, which the
  gate does catch — but the claim as written should not be carried forward.
- **F7 (boundary, informational).** The `~> v2` range means "the pinned
  GoReleaser version" in the DoD names something that does not exist. Bounded
  at both ends instead; see §1.

## 10. Routing — why this run could not record `accept_cr`

The runner queued this recovery attempt with:

> reviewer completion cannot infer acceptance from done for TASK-260908-1jv1h3;
> acceptance must be recorded by accept_cr and routed through integrating

`accept_cr` is **structurally unreachable for this element**, proven from both
sides rather than asserted (`TASK-260908-1jv1h3_accept-cr-refusals.log`):

```
accept_cr(…, revision=1, …)
  → change_request_acceptance_unauthorized: … this reviewer run was handed
    Change Request revision 0 and is accepting revision 1 …
accept_cr(…, revision=0, …)
  → revision must be a positive integer, got "0"
```

Revision 0 is the "no candidate was handed" sentinel and is refused as
not-positive; revision 1 is refused because this run was handed revision 0.
There is no argument this run can supply. Corroborating: no
`_change-request_rev*.patch` exists for this element, its `events.ndjson`
contains **zero** `change_request` entries, and
`worktree status STORY-260908-g7o5zw` lists no Change Request.

The cause is structural, not a missed step. Per
`references/tracked-background-spawn.md`, *"a reviewer-archetype run skips with
`not_a_producer` — a reviewer judges a candidate rather than publishing one"*,
so no reviewer run will ever create one. And the reviewed work lives in a
**different repository** (`relux-works/curator`, branch
`chore/rc-stays-out-of-install-channels`), produced outside this Story's managed
worktree — nothing ever entered the Story worktree, so no producer candidate was
ever snapshotted for a reviewer to accept. My prompt carries no
`Change Request Under Review` section, which is the same fact seen from the
other side.

The element is therefore already at the correct branch of the role contract —
**accepted → `done`**, recorded by `RUN-260908-41a57e` with evidence — and is
now terminal (`set_status` → `terminal_status: … is in terminal state done`).
My independent verdict **agrees** with it, so there is nothing to re-route. A
third attempt would hit the identical two-sided refusal.

**For the orchestrator, not for another reviewer cycle:** the open item is not
this review. It is (a) landing PR #65, and (b) the runner's reviewer-satisfaction
predicate, which demands a Change Request for a leaf whose subject repository is
not the Story workspace. That is an orchestrator/task-board decision, not
reviewable work.

---

**repeat-of:** none

**Branch:** ACCEPT. Recorded status `done` (already set by `RUN-260908-41a57e`);
`accept_cr` unreachable and evidenced above.
