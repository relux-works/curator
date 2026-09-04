# Review findings 2: Decision 0010 draft (agent environment profiles)

Reviewer run for TASK-260831-1rjz6j, review cycle 2, 2026-08-31.
Subject: `decisions/0010-agent-environment-profiles.md` at `fe21fb0`
(`fe21fb02b8008b2cbde83ffc9181f4f33657ba3e`) on
`draft/agent-environment-profiles`, on top of `3fd5617`, `b6b4ef9`.
Inputs: `review-findings-1.md` (13 findings), `rework-report-1.md`.

**Verdict: ACCEPT.** All 13 cycle-1 findings are genuinely resolved in the
document text — verified against the file, not the rework report. The
regression sweep found no new blocking, major, or minor issues; one new nit
is recorded below for the normative-authoring phase, not as rework.

## Finding-by-finding verification (all 13 resolved)

Each item was checked in the document at `fe21fb0`; spec claims re-verified
against `~/Developer/ReluxWorks/curator-spec` (read-only main checkout).

1. **major — secret-canary overclaim: RESOLVED.** Decision 2's last paragraph
   now describes manager §7 exactly as the spec states it (raw-tree content
   hash; static canary whose failure always blocks; deterministic detectors;
   revocation — verified verbatim against manager §7), states explicitly that
   today's detectors do not cover secrets in context modules, and adds both
   rules from the brief: always-strict profile installation ("an advisory
   profile install does not exist") and a secret-detection detector class
   over context modules and the profile manifest as authorized normative
   work. Security impact bullets 1 and 5 and the Compatibility impact
   paragraph are coherent with both rules. Grep confirms zero occurrences of
   "secret canary" and "audit-blocked" anywhere in the document.
2. **major — builtin `default` profile: RESOLVED.** Decision 8 defines the
   `local` source kind (no git identity, no ref, no effective commit; store
   key = §8 content hash of state; switching/sync/status uniform). Decision 4
   carries the store-keying sentence with the non-divergence guarantee;
   Decision 9 defines the `profile list` rendering (`local` source, `-` ref,
   state hash in the effective-pin column). Both recorded deviations verified
   present and correct: the Decision 1 sentence admitting the `local` kind
   (removing the flat contradiction with "installed from a git source") and
   the Consequences parenthetical ("state-hash-keyed for a `local`
   profile"). Both are pure self-consistency, no new semantics — deviations
   accepted.
3. **major — IR determinism: RESOLVED** with the brief's
   validation-and-reject shape: snapshot validation rejects a module that is
   not valid UTF-8, contains a non-LF line ending, or lacks exactly one
   trailing LF; a validated module stays opaque and is never rewritten;
   output = applicable modules in manifest order joined by one empty line (a
   single additional LF), no other transformation. LF encoding and single
   trailing LF hold by construction for one or more modules. The
   conformance-vector sentence is retained and is now true.
4. **minor — fragment boundary: RESOLVED** (weakened-claim option). Decision
   6 and Security impact bullet 3 both now bound the claim correctly: names
   from the closed registry, values below the manager-owned environments
   root, the profile-name path segment as the only profile-derived component,
   bounded by the §2 identifier grammar.
5. **minor — claude_code passthrough per platform: RESOLVED.** Decision 7
   splits by platform (macOS Keychain ambient, Linux `.credentials.json`
   passthrough, Windows deferred to OQ6); OQ6 names the credential shape.
6. **minor — shadowing paths: RESOLVED** (stronger option). Decision 3's
   adapter declaration includes known shadowing paths with pi's
   `AGENTS.override.md` as the named example, plus the warn rule for
   materialization and `env status`; Decision 9 lists existing shadowing
   paths in status output.
7. **minor — XDG_CONFIG_HOME blast radius: RESOLVED.** New Decision 3
   paragraph discloses the process-tree scope, relates it to the rejected
   `HOME` substitution, and narrows it (symlink seeding of other `~/.config`
   entries; vendor variable supersedes).
8. **minor — mode defaults: RESOLVED.** Decision 4 "Mode defaults" paragraph:
   `linked` for the four native-home in-place surfaces (manager §5
   symlink-with-copy-fallback — wording verified against manager §5),
   `copied` for secondary targets, managed homes always link from the store,
   overrides live in the registry. Decision 3's forward reference now
   resolves.
9. **minor — bare §5: RESOLVED.** Now "manager §5 behavior"; all eight §5
   citations in the document carry the `manager` prefix (grep-verified).
10. **minor — wrong status anchor: RESOLVED.** Decision 9 cites the manager
    §10 status discipline; §2.4 citation removed (grep confirms zero `§2.4`
    occurrences). Manager §10 verified as the read-only
    recompute-and-report / never-mutate section.
11. **nit — OQ recommendations: RESOLVED.** OQ4 and OQ6 both carry explicit
    recommendations matching the suggested working assumptions.
12. **nit — Xcode docs-confidence: RESOLVED.** Context and Decision 3 mark
    the CodingAssistant parent path and Xcode 26.3 attribution as
    docs-confidence/docs-recorded with verification folded into OQ6, which
    names both explicitly.
13. **nit — environment marker name: RESOLVED.** Named
    `.csk-environment.json` (schema 1) in Decision 4 and in the Compatibility
    impact identifier list.

## Regression sweep (reworked and unchanged sections)

- Cross-references: Decision 1 ↔ 8 (`local` kind), Decision 3 ↔ 4 (mode
  defaults), Decision 8 ↔ 4/9 (store key, list rendering), Security ↔
  Decision 2 (always-strict + detector class), Compatibility ↔ Decision 2
  (detector class as normative work) — all consistent. Open questions still
  number 1–7 with no dangling references; phasing table unchanged and
  coherent.
- Spec citations introduced by the rework re-verified against
  `profiles/manager.md`: §7 audit pipeline wording, §10 status discipline,
  §5 fallback discipline and unknown-identifier behavior — all accurate.
- No residual overclaim language (grep: "secret canary", "audit-blocked",
  "§2.4" all absent).
- **New nit (not rework; record for the normative phase):** Decision 2's
  determinism paragraph does not state the zero-applicable-modules case — a
  profile whose modules all carry selectors excluding some environment.
  Empty output cannot "end with exactly one trailing LF", and whether the
  root-context surface materializes as an empty file or is absent is
  unstated. `protocol/environments.md` should pin this; no document change
  requested now.

## Validation evidence

- Worktree verified: branch `draft/agent-environment-profiles`, head
  `fe21fb0`, signed (good ECDSA signature; the local principal-mapping
  warning is stale verifier config, not a signature failure), single commit,
  only `decisions/0010-agent-environment-profiles.md` changed (123
  insertions, 51 deletions), tree clean except the `tools/__pycache__/`
  residue the cycle-2 brief excludes.
- Re-ran myself: `go test ./tools/...` in the draft worktree — ok.
- Accepted from producer evidence: `tools/validate.py` (venv) and unittest
  gates. Basis: stock python3 lacks `jsonschema` here exactly as the
  producer's LOGBOOK finding records, `validate.py` contains zero references
  to `decisions/` (grep-verified), and the delta is markdown-only under
  `decisions/` — it cannot affect schema/vector validation.
- Change Request `CR-TASK-260831-1rjz6j-1` rev 1 delta in the curator story
  worktree reviewed: a LOGBOOK.md-only addition (two entries: audit-pipeline
  secret-detector finding, rework-1 disposition). Content is accurate
  against everything verified above; no other paths touched.

## Verdict

ACCEPT. Recording acceptance via `accept_cr(TASK-260831-1rjz6j, revision=1)`
with evidence `TASK-260831-1rjz6j_review-verdict.md`. The one new nit above
is guidance for the normative-authoring work this decision authorizes, not a
condition on acceptance.
