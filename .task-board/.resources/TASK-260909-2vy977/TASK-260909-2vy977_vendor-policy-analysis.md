# F4 vendor-policy analysis — TASK-260909-2vy977 (read-only, no source changes)

Role: solution-architect (analyst). No implementation, no CR, no suite runs.
Candidate bytes and checkpoint ae8676c preserved; `git status` inspected read-only only.
Verdict under analysis: `TASK-260909-2vy977_review-verdict-rev2.md` (CR rev2 NOT accepted; F4 blocking/new, F3 repeat).

## Question

Does any existing authoritative policy select one native-Pi runtime/vendor
(pi-anthropic / pi-openai / pi-google) before vendor-local `Lineup`, so that
SPEC §4.3 level 3 ("highest-ranked admitted model for the mapped system") is
determinate for `pi` → `pi-native`?

## Answer: no. A genuine operator product decision is required.

No authoritative source selects a Pi runtime before Lineup. The union ranking
in `internal/defaults/lineup.go` (`systemCandidates` + `Lineup(models)[0]`)
is therefore an invented product policy, not an implementation of SPEC §4.3.

## Evidence (exact sources)

1. Upstream v0.5.11 affirmatively forbids cross-vendor comparison:
   - `pkg/vendorplugin/vendor.go` (`CapabilityRank` doc): "Scores are per-vendor
     and never comparable across vendors: nothing in this module can honestly say
     a vendor's top model beats another vendor's, and a type that invited the
     comparison would get one made." Plus: "The scale's own units are the vendor's
     business — the ported rows run 10..120".
   - `pkg/vendorplugin/vendors/local-models/models.go` (`localCapabilityRank`):
     "A capability score is comparable only within one vendor's own lineup".
   - `pkg/vendorplugin/lineup.go`: "The DERIVED total order over a vendor's models."
     "A position is PRESENTATION. It is a fact about this list, not about the models."
     `LineupOf` = Lineup over "a registered vendor's own list";
     `RuntimeDeclaration.LineupOfDeclaration()` (`runtime.go:296`) = `Lineup(d.Models)`.
     Accepting an arbitrary `[]Model` is mechanical capability, not authority.
2. Upstream gives the launcher no default-runtime selector:
   - `pkg/vendorplugin/spawn.go:68-69`: "`Runtime` selects the declared (agentic
     system × vendor) pair." The caller supplies it; `BuildLaunch` resolves exactly
     the named runtime (`resolveLaunchBinding`, spawn.go:144,294).
   - `Registry.RuntimeDeclarations()` (`registry.go:607-619`) returns declarations in
     sorted id order — an iteration order, not a selection policy. Alphabetical order
     (pi-anthropic first) contradicts the score-union winner (pi-openai, 120 > 80 > 50),
     which proves the order carries no policy meaning.
3. SPEC / Decisions predate the 1-system→3-runtimes shape and select system, not runtime:
   - SPEC §4.2 maps `pi` → system `pi-native` (+ `ax` provider `pi`); no runtime cell.
     SPEC §4.4 `Runtime` names only `claude`/`codex` and leaves "native-Pi availability
     subject to the upstream evidence and release boundary in §4.2".
   - SPEC §4.3 level 3 (="highest-ranked model of `vendorplugin.Lineup` … among the
     models the registry admits for the mapped system") restates Decision 0013 D6.2
     (`curator-spec/decisions/0013-execution-ownership-and-launch-plans.md:449-466`),
     written when each system had exactly one runtime. Neither text defines a
     cross-runtime selection step.
   - Decision 0012 (context packages/locks) contains no vendor/runtime preference.
4. Accepted native-Pi design (TASK-260908-ggxfte rev2 results §3) ships three frozen rows
   (`pi-anthropic`/`pi-openai`/`pi-google` = `pi-native` × vendor) and binds
   configured-model → runtime ("runtime = frozen pi-* row whose vendor `Models()`
   carry the id"). It never names an empty-default → runtime choice; §6 records
   "None unavoidable" with implied defaults limited to (three vendors, catalog-verified
   ids, provider-qualified argv). No vendor preference; review-verdict-rev2 confirms none.
5. `internal/mapping.Resolve` returns `Target{System, Provider}` only — no runtime.
6. Failed assumption (recorded, not re-litigated): "§4.3 already asks for the
   highest-ranked admitted model for a system" was read as complete. With three
   runtimes serving one system the sentence is indeterminate: the reviewer probe
   (real `NewRegistry` + tagged Lineup: anthropic top 80, google top 50, openai top
   120; google-units-×10 flips the winner) shows the union compares incommensurable
   scales. The candidate's own comment ("deterministic rather than meaningful across
   vendors", `lineup.go`) concedes the point. SPEC rev2's §4.2 paragraph ("resolving one
   of the frozen runtimes") and README's union paragraph ("Candidates union across
   every runtime… resolves to the union leader") are producer prose, not an accepted
   design decision or errata, and cannot supply the missing policy.

## Clean alternatives (no invented normalization, order, hard-code, or refusal-only)

- A. Reviewed runtime preference, then that vendor's Lineup (RECOMMENDED). SPEC (or a
  named launcher config knob) records an ordered Pi-runtime preference; level 3 runs
  `Lineup` over that runtime's driven rows only and binds its contributor. Preserves
  Lineup authority and row recommendation/no-effort semantics; deterministic;
  per-member flags/files overrides keep working. Cost: one explicit product decision.
- B. Refuse empty Pi defaults (`defaults_unresolvable` until configured). Honest but
  breaks the all-three-env positive-completion AC and fresh-install UX. Not recommended.
- C. Approved cross-vendor policy with its own cited basis (benchmark/cost/latency).
  Legitimate only with evidence and reviewed SPEC errata; no such basis exists today.
  Score renormalization on vendor-local units is fabrication. Not available now.
- D. First-runtime / declaration order. Arbitrary (alphabetical, rename-fragile),
  contradicts "position is presentation". Rejected.
- E. Refusal-only Pi. Narrows acceptance and fakes coverage. Rejected per brief.

## Minimal operator product decision (exact input needed)

1. Where does the Pi runtime preference live? (SPEC §4.2/§4.3 table vs launcher-owned
   config; recommended: SPEC table, single ordered preference, e.g. explicit ranking
   of the three frozen rows.)
2. What is the ordered value? (No recommendation invented here; candidates are any
   order of pi-anthropic/pi-openai/pi-google. Decision must cite a basis — installed
   Pi catalog, provider maturity, cost/latency — or declare itself a pure convention.)
3. Does operator/machine `defaults.json` override it per member as today (yes expected)?
4. Rework scope: replace union ranking with preference+vendor-local Lineup at
   `Files.Complete` (`internal/defaults/lineup.go:179` call site) + SPEC errata +
   production-entry tests; keep contributor binding and multi-contributor ambiguity
   refusal. This is analysis/rework routing, not the former missing-tag blocker
   (v0.5.11 tag object `0ea486e` / peeled `a2a6e9f` already verified in rev2).

## Bounded F3 attestation gate (outline only; no launcher tooling, no source edits)

Route under the CR-lifecycle skill as a reviewer checklist gate owned by the
orchestrator (repeat-of rev1-F3 class: full-suite accounting + coverage denominator):

1. Exactly ONE runtime-owned `make check` counts as validation
   (`change-request_revN-validation.log`, `$ make check` … `[exit N]`). Any manual full
   run is a second execution: it must be appended and counted, never overwritten, and
   the outcome's "exactly once" claim corrected append-only.
2. Fail-closed shell accounting: the log must name the executing shell and show
   `pipefail` (or equivalent) with `PIPESTATUS`/exit captured from `make` itself;
   `… | tail` + outer `echo` exit-0 proves nothing about the suite.
3. Honest production-entry denominator: a table whose rows are owned AC requirements
   and whose "driven" column names a committed test reaching the PRODUCTION call site
   (`cmd/curator-run.run → Files.Complete → EmitGroup`; `BuildLaunch` admission).
   Helper-direct tests, pin checks, and dependency inspection are bounds, not driven rows.
4. Mutant bar unchanged (narrowing per gate; token-preserving behavioral mutant for
   source-text gates with the behavioral suite executed), reported — not re-run for
   ceremony — when already attached and valid.

## State

No bytes changed, no CR published/accepted, no tests run, no tags/releases/ax/CI,
no control-root or LOGBOOK writes. Next: operator decision above → rework → new CR →
independent Astra-medium review.
