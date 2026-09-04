# Rework report 1: Decision 0010 draft (agent environment profiles)

Producer run for TASK-260831-1rjz6j, 2026-08-31. Input: board resource
`review-findings-1.md` (3 major, 7 minor, 3 nit). All 13 findings applied.

- Branch: `draft/agent-environment-profiles`
- Base head before rework: `3fd5617`
- Rework commit: `fe21fb0` (`fe21fb02b8008b2cbde83ffc9181f4f33657ba3e`),
  signed (`git log --show-signature` reports a good ECDSA signature by the
  configured author key), single commit, only
  `decisions/0010-agent-environment-profiles.md` changed
  (123 insertions, 51 deletions). Working tree clean; not pushed.

## Finding-by-finding disposition

| # | Severity | Disposition |
|---|---|---|
| 1 | major | Applied per brief. Decision 2 last paragraph now states what manager §7 actually delivers (raw-tree hashing, static canary as detector self-test that always blocks, deterministic detectors, revocation), says explicitly that today's detectors do not cover secrets in context modules, and adds both suggested rules: profile installation is always-strict (no advisory install for profiles) and the authorized normative work includes a secret-detection detector class over context modules and the profile manifest. "Secret canary" phrase removed everywhere. Security impact bullets 1 and 5 reworded coherently; Compatibility impact names the detector class as normative work joining the manager §7 pipeline. |
| 2 | major | Applied per brief. Decision 8 defines `default` with source kind `local`: no git identity/ref/commit, store key = §8 content hash of state, uniform switching/sync/status. Decision 4 gains the store-keying sentence (non-divergence guarantee holds); Decision 9 gains the `profile list` rendering (`local` source, `-` ref, state hash as effective pin). Deviation (recorded): also added one sentence to Decision 1 ("Decision 8 admits one additional source kind, `local`…") so the profile-model definition no longer flatly contradicts the migration, and a parenthetical to the Consequences store bullet — both purely for self-consistency, no new semantics beyond the brief's resolution. |
| 3 | major | Applied per brief: validation-and-reject. Decision 2 now rejects at snapshot validation any module that is not valid UTF-8, contains a non-LF line ending, or lacks exactly one trailing LF; a validated module stays opaque bytes and is never rewritten. Output = applicable modules in manifest order joined by exactly one empty line (single additional LF), no other transformation; LF encoding and single trailing LF hold by construction. Conformance-vector sentence unchanged and now true. |
| 4 | minor | Applied (weakened-claim option). Decision 6 boundary paragraph and Security impact bullet 3 now say: names from the closed registry, values are paths below the manager-owned environments root, the only profile-derived component is the profile-name path segment bounded by the §2 identifier grammar (no separators, no traversal). |
| 5 | minor | Applied. Decision 7 declares the claude_code passthrough per platform: macOS Keychain ambient, Linux `.credentials.json` passthrough, Windows deferred to open question 6; OQ6 extended to cover the credential shape. |
| 6 | minor | Applied (stronger option). Decision 3 adapter declaration now includes known shadowing paths (pi `AGENTS.override.md` named as the example) plus the rule that materialization and `env status` warn when a declared shadowing path exists; Decision 9 lists existing shadowing paths in the `env status` output. |
| 7 | minor | Applied. New Decision 3 paragraph discloses the `XDG_CONFIG_HOME` process-tree blast radius, relates it to the rejected `HOME` substitution, and narrows it: launcher-provisioned managed opencode parents are seeded with symlinks to the operator's other `~/.config` entries; a dedicated vendor variable supersedes the XDG mechanism if shipped. |
| 8 | minor | Applied per suggested sentence. Decision 4 "Mode defaults" paragraph: `linked` default for the four native-home in-place surfaces (manager §5 symlink-with-copy-fallback), `copied` for secondary fixed-home targets, managed homes always link from the store; per-adapter overrides live in the registry, never in profile data. This also makes Decision 3's forward reference true. |
| 9 | minor | Applied: "their manager §5 behavior". |
| 10 | minor | Applied: Decision 9 cites the manager §10 status discipline ("recompute and report, never mutate"); the §2.4 citation removed. |
| 11 | nit | Applied. OQ4 recommendation: keep the manager §5 native surface until a pinned opencode release proves `<home>/skills/`. OQ6 recommendation: hold the conformance-vector freeze on the platform evidence, draft against recorded shapes meanwhile. |
| 12 | nit | Applied. Context and Decision 3 mark the CodingAssistant parent path and Xcode 26.3 attribution as docs-confidence with verification folded into OQ6; OQ6 names both explicitly as implementation-time verification items. |
| 13 | nit | Applied. Environment marker named `.csk-environment.json` in Decision 4 and in the Compatibility impact identifier list. |

## Verification

- Grep confirms no residual "secret canary", "audit-blocked", bare "their
  §5", or "manager §2.4" citations; every remaining "Xcode 26.3" occurrence
  is hedged as docs-recorded.
- Spec wording for the rewritten claims was checked against
  `~/Developer/ReluxWorks/curator-spec`: manager §7 (canary/strict/advisory
  semantics), manager §10 (status discipline), manager §5
  (symlink-with-copy-fallback), core §2 (identifier grammar).
- Document self-consistency after edits: Decision 3 ↔ 4 mode-default
  forward reference now resolves; Decision 8's `local` kind is referenced
  from Decisions 1, 4, 9 and Consequences; no section renumbering.
