# ax integration delta — authoring notes (TASK-260901-14atb7)

Proposal delta for `relux-works/agent-session-manager-spec` delivering the
curator-spec Decision 0010 §10 (D10) contract and the
`protocol/environments.md` §10/§11 promise as one minimal additive PR.

- Worktree: `~/Developer/ReluxWorks/.temp/ax-curator-integration/worktree`
- Branch: `draft/curator-environment-integration`, forked from `origin/main` = `28bf96d` (verified equal at creation)
- Signed commit: `d7075e1a30ff11d5241e6bddef971f3dc3aff5ca` (`git log --show-signature`: Good "git" signature, ECDSA SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM)
- Not pushed — the orchestrator opens the PR.

## Exact sections touched (SPEC.md, all prose-only, additive, no heading changes)

1. **§5.1 Session Record** — one paragraph after the Launch Plan sanitization
   paragraph: OPTIONAL Curator integration, SHOULD record launch-time
   environment provenance in Session Record `extensions` under
   `works.relux.curator.profile-name`, `works.relux.curator.profile-pin`
   (commit for a git profile / state hash for a local profile, never both —
   mirrors the fragment's pin shape), and
   `works.relux.curator.fragment-digest` (`sha256:`-prefixed digest of the
   exact `launch-env-fragment-v1` JSON bytes consumed at launch). Absent keys
   = not Curator-launched; §1.6 extensions rules explicitly restated as
   binding (no core-semantics influence; opaque preservation elsewhere).
2. **§7.5 Required operations** — one paragraph after the embedded-types
   table's trailing prose: fragment `env` variables merge into the Launch
   Plan / `SpawnPlan` `env_literals` map; fragment variable names are a
   closed set from the Curator environment adapter registry
   (`relux-works/curator-spec`, `protocol/environments.md`, §10, cited by
   name — no vendored copy); values are non-secret absolute managed-home
   paths; existing §5.1 name-grammar/count/byte/secret limits unchanged; no
   new `SpawnPlan` member.
3. **§13.10 Resume** — one paragraph at the section end (after the
   Antigravity note): on resume (and §13.8 fork, which records fresh values
   in the new record) SHOULD re-resolve via
   `curator env resolve ENVIRONMENT --profile PROFILE --format json`;
   resolved pin ≠ recorded `profile-pin` is environment drift; default
   warn-and-continue, MAY offer strict refusal; a failed resolution is a
   distinct fact and MUST NOT be treated as proof of currency; no lease/
   fencing/checkpoint/materialization semantics change.
4. **§14.1 Command surface** — informative closing note: the Curator
   umbrella may expose ax as `curator session` via `curator-NAME` PATH
   discovery (`protocol/environments.md` §11); packaging convention only,
   zero normative impact on ax.

Plus one mechanical companion outside SPEC.md:

5. **`scripts/validate_spec.py`** — `FROZEN_RELEASE_DOCUMENT_SHA256["SPEC.md"]`
   updated `562546d2…` → `6bbefa4d…`. The validator pins an LF-normalized
   SHA-256 of the reviewed SPEC.md prose and its own comment requires the map
   to be deliberately replaced for an intentional revision; without this the
   repo's single validation command is red on any SPEC.md edit. Maintainers
   may re-mint it on merge; flagged here so it is a visible, deliberate part
   of the proposal, not a smuggled change.

## Key grammar chosen and why

`works.relux.curator.profile-name` / `.profile-pin` / `.fragment-digest`.

- §1.6: extension keys are reverse-DNS, 3–253 lowercase chars, labels
  `[a-z][a-z0-9-]{0,62}` — underscores are not allowed, so hyphenated final
  labels are the only conforming spelling.
- The spec's own existing extension examples use exactly this style and
  prefix family: `works.relux.ax.launch-hint`, `works.relux.ax.goal-label`,
  `works.relux.ax.board-scope` (§9.3 bundles). `works.relux.curator.*`
  parallels that convention with the owning component as the third label.
- Values are plain strings (valid `ExtensionValue`), non-secret, immutable
  with the record.

## Deliberately NOT touched

- No version bump (VERSION, CHANGELOG.md, RELEASE_NOTES.md, README.md
  untouched — maintainers decide the release framing; CHANGELOG has no
  Unreleased section convention to extend).
- No normative invariants, no §8 matrices, no fixtures: the delta introduces
  no new schema, field, operation, error code, or fixture-list row —
  extension keys are already-valid extension data and every requirement is
  SHOULD/MAY on an OPTIONAL integration. No fixture list demanded a row.
- No new headings/anchors (validator checks heading anchors; delta is
  paragraph-level only), no diagrams, no CLI command-surface additions (the
  strict drift mode is left as "MAY offer a strict mode" rather than minting
  a normative flag).
- No JSON examples added (every JSON block in SPEC.md carries a recomputed
  canonical digest; adding one would grow the delta for no contract value).

## Validation evidence (all run in the worktree at d7075e1)

| Command | Exit | Result |
| --- | ---: | --- |
| `./scripts/validate_spec.py` (before hash update) | 1 | Expected red: single error — frozen SPEC.md baseline mismatch; all 279 semantic checks otherwise passing |
| `./scripts/validate_spec.py` (after hash update) | 0 | 279/279 semantic checks passed; links, anchors, ledgers, forbidden-claim scans green |
| `./run_validation.sh` | 0 | Full suite: contracts + Structurizr/C4 validation + diagram freshness/byte integrity — "Validation successful"; log `.temp/ax-curator-integration/logs/run_validation-01.log` |
| `./scripts/test_expected_red.sh` | 0 | 304/304 mutations correctly rejected with actionable diagnostics — the validator still refuses tampered prose/fixtures with the re-minted baseline (negative evidence for the frozen-hash gate); log `.temp/ax-curator-integration/logs/expected-red-01.log` |

## Anomalies / findings

- The frozen-baseline gate means **every** SPEC.md-touching PR must re-mint
  the hash to keep `run_validation.sh` green; this is by design (deliberate
  revision marker), recorded here so the orchestrator's PR description can
  state it.
- Local branch tracks `origin/main`; nothing was pushed and no stash was
  used. Worktree left in place for the orchestrator's PR step.
