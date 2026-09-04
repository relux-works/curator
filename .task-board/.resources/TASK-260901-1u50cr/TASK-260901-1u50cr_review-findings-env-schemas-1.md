# Review findings — environments schemas and conformance vectors (round 1)

Task: TASK-260901-1u50cr · CR-TASK-260901-1u50cr-1 rev 1
Reviewer run: RUN-260901-7c4db1 (reviewer archetype)
Subject: curator-spec worktree `~/Developer/ReluxWorks/.worktrees/curator-spec-env-schemas`,
branch `draft/environments-schemas`, head `cef93fbd95b43a13932bb0d8b397e177c5301045`,
base `c3b29b1` (= main = origin/main, verified). Curator-repo CR delta: `LOGBOOK.md` only.

## Verdict: ACCEPT

No blocking or major findings. All review dimensions pass; minor observations below.

## 1. Schema fidelity (prose of protocol/environments.md wins) — PASS

- `profilefile-v1` vs §2: version const 1, non-empty `profiles`, identifier member
  names, portable-path values, `additionalProperties: false`. Aliased/nested
  profile roots are cross-field and enforced in `tools/validate.py`
  `validate_wire_semantics` (lines 722–731), the repo's established convention.
- `context-manifest-v1` vs §3: required possibly-empty `modules`; entries with
  required portable `path`, optional non-empty unique `environments`, optional
  `class` root|system (default root left to semantics, correct — a schema default
  would not be strict). Duplicate paths in wire semantics (735–741).
- `agent-environment-marker-v1` vs §8.2: git/local pin branches closed via oneOf
  with `additionalProperties: false` each (commit XOR state_sha256 enforced —
  `invalid-git-with-state-pin`, `invalid-local-with-ref` prove it);
  composition⇔precedence `dependentRequired` in both directions; mode enum
  exact; per-surface `paths`/`form`/`content_sha256`; `copy_fallback` required
  iff `mode=linked` and rejected otherwise; `seed_links` rejected off
  managed-home. Sorted surface keys in wire semantics (742–747).
- `launch-env-fragment-v1` vs §10.2: fragment const, closed environment enum,
  pinned profile/composition oneOf, closed `env` variable-name enum
  (`CLAUDE_CONFIG_DIR`/`CODEX_HOME`/`XDG_CONFIG_HOME`/`PI_CODING_AGENT_DIR`
  matching §7.1), all four channel-descriptor kinds with their exact carrier
  field, closed semantics enum, empty `channels` admitted (opencode declares
  none in rev 1). Unknown fields/kinds/semantics rejected as §10.2 requires,
  with negative cases for each.
- All referenced `common.schema.json` `$defs` exist and are apt (`pathSet`
  allows the empty array — "required arrays are present even when empty").

## 2. Vector correctness — PASS (recomputed, all cases, independently)

I wrote an independent Python implementation from the prose alone (§5, §5.1–5.6,
core §8, registry §1 CCJ-1) without importing `tools/validate.py`
(scratch `recompute.py`) and recomputed **every** vector, exceeding the
three-by-hand bar: all 4 header cases (bytes, sha256, line counts), all 11
materialization cases (per-file expected bytes, per-file sha256, §5.6 surface
hashes). 0 mismatches. Flagship confirmations:

- header grammar: 6 lines uncomposed, 6+N+1 composed, exact `generated:` and
  `notice:` byte strings, `state sha256:` pin form for local;
- chapter part exactly `---` LF LF `## Profile: <name>` LF, including the empty
  chapter (emptyoverlay) and part-joining with exactly one blank line;
- opencode CCJ-1: `{"instructions":[...]}` sorted-key, no-whitespace bytes plus
  exactly one trailing LF; zero-modules variant `{"instructions":[]}` + LF;
  opencode referenced root file is the header part alone (its sha equals the
  single-profile header sha — internally consistent);
- zero-modules monolithic output is the header alone; no-context writes nothing
  and binds no surface hash; system-prompt output is header-free and only the
  applicable system modules in chain order; none-applicable writes nothing.
- Sampled negative schema-cases (8 across the four schemas, including
  `invalid-surfaces-unsorted`, `invalid-managed-home-copy-fallback`,
  `invalid-seed-links-on-linked-home`, `invalid-flag-channel-with-filename`,
  `invalid-nested-root`, `invalid-duplicate-path`) each violate exactly what
  their name claims; the case runner asserts both directions, so a mislabeled
  case fails `make validate` (proven below).

## 3. Gate attacks (licence to defeat) — gate held, fail-closed

Run against a scratch copy of the tree, driving the production entry point
`python tools/validate.py` (wired into `make validate` and CI):

- byte flip in an expected file → rejected;
- **full self-consistent forgery** — tampered expected bytes with per-file
  sha256, §5.6 surface hash, `conformance/v1/manifest.json` entries, and the
  `release/1.0.0-rc.9.json` manifest pin all recomputed over the forged bytes —
  still rejected: `environment case monolithic-composed-empty-chapter: digest
  for CLAUDE.md is stale`. The validator recomputes ground truth from the case
  inputs per the §5 rules; no downstream artifact is trusted;
- self-consistent header forgery (expected_bytes+sha+line_count) → rejected
  (`header case single-profile bytes are stale`);
- dropped materialization case → rejected (exact case-inventory check);
- expected `opencode.json` with trailing LF stripped, chain forged → rejected;
- negative schema case relabeled `valid: true` → rejected with the real
  violation named.
- Restored copy validates green after each attack (no false positives).

Negative test suites narrow, not just delete: `EnvironmentVectorTests`
(13 fail-closed tests: precedence/pin mutation, CRLF bytes, missing trailing
LF, selector widening, surface-hash mutation, absence↔written contradictions,
stale inventory) and 6 Go tests including determinism and grammar exactness.
Production call site: `validate_environment_vectors` in `main()` of
`tools/validate.py` — executed by `make validate` and by CI directly.

## 4. Producer judgment calls — verdicts

1. **Header grammar, prose over brief ("7-line")** — CORRECT. §5.1 is normative
   and yields 6 lines uncomposed / 6+N+1 composed; all four header vectors
   byte-match my independent recompute of §5.1. The brief's "7-line" is
   informal display counting.
2. **`environment` ↔ `env` variable-name binding left semantic** — ACCEPTED.
   Matches the repo convention that cross-field rules live in
   `validate_wire_semantics`, the value space is closed to the four registry
   names, and the divergence was flagged, not silent. Optional follow-up: a
   one-line wire-semantics rule (or per-environment `if/then`) could bind the
   pair structurally; not required for acceptance.
3. **`copy_fallback` required iff `mode=linked`** — CORRECT. Faithful strict
   reading of §8.2 ("for a `linked` home — whether any entry fell back")
   under the required-even-when-empty discipline; both directions carry
   negative cases and both were proven live (attack 5 + schema case sampling).

## 5. Wiring and regeneration — PASS (re-run by me)

- `make validate` (worktree `.venv`): exit 0 — 57 schemas, 766 vector files,
  147 unittests OK, `go test ./tools/...` ok.
- Generator determinism: `go run ./tools/generate-vectors -root .` run twice;
  tree hash over `conformance/v1` + `release/` identical both runs:
  `a84f1c14e9eaa515870745f96341b12d6c09510d1e6dd987b1183110cb7728fc` (matches
  the producer's reported hash) and `git diff --exit-code` clean vs the
  committed bytes — regeneration reproduces commit `cef93fb` byte-for-byte.
- Registration follows the existing pattern exactly: 57 cases in
  `schema-cases/index.json` (9+11+19+18), `vectors/environments.json` +
  `expected/environments/` in `manifest.json`, validator in `main()`. CI
  (`.github/workflows/ci.yml`) already runs validate, unittest discovery,
  `go test`, and the regenerate + `git diff --exit-code` proof including
  `release/1.0.0-rc.9.json` — no workflow edit needed, as the notes state.

## 6. Repo hygiene — PASS

- Delta (89 files, +4410/−3) confined to `schemas/v1/`, `conformance/`,
  `tools/`, plus `conformance/README.md`, `schemas/v1/README.md`, and
  `release/1.0.0-rc.9.json`. The release-pin change is mechanically forced by
  the generator and gated by `regenerate-check` (Makefile and CI both diff that
  file), so it is convention-mandated, disclosed in the notes, and correct —
  the review brief's "only schemas/ conformance/ tools/" list was incomplete,
  not the delta wrong.
- No `protocol/`, `profiles/`, `cli/`, `CHANGELOG`, or `decisions/` changes.
- Commit `cef93fb` carries a Good ECDSA signature by the repo's configured
  signing key (`~/.ssh/ivanopcode`, fingerprint
  `SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`); principal matching is
  unavailable locally only because the repo-local `allowedSignersFile` points
  at a stale temp path — pre-existing config, not part of the delta.
- Worktree clean at head; my transient `tools/__pycache__/` removed.
- Curator-repo CR delta (`LOGBOOK.md` only): the entry's claims all reproduced
  under this review (venv requirement, validate numbers, tree hash, signed
  commit). `repository_delta=present` and accurate.

## Minor observations (non-blocking, no rework requested)

- Latent tooling edge: `environment_case_files` returns no files whenever the
  activated profile lacks `context/`, even under composition with
  context-bearing overlays; the prose is arguably silent on that combination
  and no vector exercises it. Validator-internal only — no normative artifact
  is affected. Worth a vector when the prose settles the case.
- Fragment `env` values ("absolute managed-home path") are structurally
  `nonEmptyString`; absoluteness is semantic, same class as judgment call 2.

## Definition of Done check

Four schemas prose-exact and strict — yes. Determinism vectors byte-exact,
generator twice byte-identical — yes, re-proven. `make validate` green — yes,
re-run. Signed commit — verified. Notes resource — present and accurate.
Negative tests fail when the gate admits what it must reject, production call
site named — proven by live attacks. Architecture fit — follows the existing
schema/vector/wire-semantics conventions exactly.

ACCEPT. Handoff parked at `to-review` via `accept_cr`; `done` +
`commit_ack=scope_committed` belongs to the orchestrator.
