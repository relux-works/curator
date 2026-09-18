# Brief — TASK-260910-1tvf2t: signed bootstrap checkpoint interchange and the TOFU/equivocation residuals (S2)

Story `STORY-260910-6bo7ej` (tofu-and-equivocation-mitigations), wave 3 of the
2026-09 security-audit remediation; spec first, then the client task
`TASK-260910-2vnjej` (optional cross-registry Merkle-root comparison).
Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-6bo7ej/worktree`
(branch `task-board/story/STORY-260910-6bo7ej`, forked from curator-spec `main` `4a2fa3e`).
Rules: `remediation-spec-producer-rules.md` (attached). Role: doc-writer of
normative spec text; no implementation.

## Finding (read it first)
`docs/security-audit-2026-09.md` S2 (Medium): registry §5 persists high-water
state per registry URL, but the FIRST fixation is trust-on-first-use with no
authenticated checkpoint; recovery after loss requires out-of-band
rebootstrap; equivocation (divergent, each-monotonic views for different
clients) has no protocol-level detection. Text to revise: `protocol/registry.md`
§5 (rollback state / first use / rebootstrap), §2/§2.1 (keys) only for
cross-references, §8 offline behaviour, §10 verification wording;
`profiles/registry-service.md` §5/§6 (the operator checkpoint the service
already maintains for R3/P2 — reuse it, do not respecify) and §10 (threat
model, equivocation); `profiles/manager.md` §1 (registry configuration
knobs) and §10 (status); `SECURITY.md` (registry security state);
`schemas/v1/` if the registry configuration gains a member (frozen-schema
versioning rule); vectors `registry-client.json` (`rollback_state_cases`,
`snapshot_transitions` — extend, do not rewrite) and `registry-behavior.json`.

## Settled decisions (do not reopen)
- **Bootstrap checkpoint = a signed `registry-snapshot-v1` object supplied
  out of band** (the same object the R3/P2 startup comparison and the
  service's operator checkpoint use): a registry entry in machine
  configuration MAY carry `bootstrap_checkpoint` (a path to that JSON file
  or the inline object — choose one closed form and say why; a path is
  recommended, treated as manager-protected state under the S5 discipline)
  next to `public_keys`. When present: on first use the client verifies the
  checkpoint signature against the pinned keys and persists it as the
  initial high-water BEFORE any network response is accepted; the first
  network snapshot/page boundary must then satisfy §5 against it (below →
  tampered, equal-different → tampered) — no TOFU for that registry.
  Without it: TOFU stays, but the client MUST report it once as posture
  (`registry_bootstrap_tofu`, warning, naming the registry and the hint to
  pin a checkpoint) and `env status`/`curator status` list per registry
  whether the high-water was bootstrapped from a checkpoint or first use.
- **Rebootstrap after loss** uses the same object: an operator who knows the
  state was lost supplies a fresh checkpoint; the client MUST refuse a
  checkpoint whose version is below a still-present persisted high-water
  (`registry_checkpoint_regression`) — a checkpoint never lowers state.
- **Equivocation**: define the residual precisely (a registry can serve
  each client a monotonic but divergent view; the protocol detects it only
  when two views meet), and specify the OPTIONAL detection the client task
  implements: with two or more enabled registries that mirror one another
  (a closed `mirrors_of` relation in configuration, or an explicit
  federation group — choose the simpler closed form), the client compares
  `merkle_root` at the same `log_size` across them and reports divergence
  (`registry_view_divergence`, warning under advisory policy, error under
  strict) without changing resolution. Name it as optional (MAY) with the
  posture row; no quorum.
- Rollout direct (impact row "S2"). Closed diagnostics exactly the three
  above (or the tables' preferred spellings), identical everywhere.

## Deliverable
1. registry §5 (bootstrap rule, first-use posture, rebootstrap refusal),
   §10/§8 cross-references, the equivocation residual and the optional
   detection (new §5.1 or §10 subsection); registry-service §10 sentence;
   manager §1 knob rows (`bootstrap_checkpoint`, the mirror relation) and
   §10 posture rows; SECURITY.md paragraph.
2. Schema: the registry entry in `manager-config-v2` gains the closed
   members (follow the frozen-schema rule: v1 stays byte-identical; the
   `$ref` to v1's registry shape is replaced by a v2 definition that extends
   it); schema cases; generator if generated.
3. Vectors: `registry-client.json` gains `bootstrap_cases` (checkpoint
   accepted and persisted before network; first snapshot below/equal-
   different → tampered; no checkpoint → TOFU posture; rebootstrap
   regression refused; divergence detected / not detected / policy
   mapping); validator gate pins every scenario (rule 7); existing cases
   byte-identical; manifest/rc.9 regenerated.
4. `CHANGELOG.md` Unreleased entry "S2: …".

## Out of scope
Implementation (`TASK-260910-2vnjej`), R3/P2 (landed; reuse its object),
S1/S3 (`TASK-260910-2qtiho`, in review on the same base — do not restate
the hardened profile; if you need to mention policy mapping, reference
"registry policy" generically), P4 import high-water.

## Checklist and handoff
Tick the checklist items you satisfy; attach
`TASK-260910-1tvf2t_spec-patch_rev1.patch` (`git diff HEAD` of the worktree
with new files via `git add -N`; NOT `git diff origin/main`) and
`TASK-260910-1tvf2t_evidence.md` (with the `make validate` and regeneration
transcripts), then `task-board handoff TASK-260910-1tvf2t --role doc-writer`.
