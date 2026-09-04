# Review findings: protocol/environments.md draft 1 (TASK-260901-2tdoy5)

Reviewer run: RUN-260901-b7cddd. Subject: worktree
`~/Developer/ReluxWorks/.worktrees/curator-spec-environments-normative`, branch
`draft/environments-protocol`, head `eddd509` (signed, same ECDSA key as main's
`2a861e5`; base verified exactly `2a861e5` = merge-base with `origin/main`).
Delta: `protocol/environments.md` only (+947). Producer notes
(`environments-doc-draft-notes.md`) read in full.

**Verdict: CHANGES REQUESTED** — two major findings (both in the §5 byte-rule
surface the brief singled out), two minor, two nits. Everything else — coverage,
cross-references, deviations, style, diagnostics families — passes. The fixes
are localized; no structural rework is needed.

## Major

### M1. Managed `opencode.json` bytes are unspecified but hashed as a MUST-equal conformance surface

- Section: §5.3 (opencode referenced form) / §5.6 (content-hash binding) / §13.
- Quote: "for the referenced form the set contains the root file, every
  materialized module file, and — for `opencode` — the managed `opencode.json`.
  Identical (store entry, composition chain, environment, form) MUST yield an
  identical surface hash on every platform; this equality is a
  conformance-vector surface."
- What is wrong: §5.3 defines only that the manager "maintains the
  `instructions` member as the ordered list" — it never defines the complete
  member set of the manager-created file nor its serialization (member order,
  whitespace, escaping). Core §1 states JSON member order and whitespace "have
  no meaning unless a section explicitly defines canonical bytes"; no section
  does, yet §5.6 hashes those bytes into a cross-platform MUST-equal claim.
  Failure scenario: implementation A writes compact JSON, implementation B (or
  A's next release) writes indented or key-reordered JSON — same inputs,
  different surface hash, and every opencode referenced home reports false
  `environment_surface_drift`. This is the shape of a determinism claim two
  implementations can legitimately read differently — exactly what §5's opening
  paragraph forbids.
- Suggested fix: state the managed file's exact content normatively — e.g. the
  file is exactly the CCJ-1 bytes (registry.md §1) of
  `{"instructions": [ ... ]}` with the listed values and nothing else, plus one
  trailing LF or none (say which) — or exclude `opencode.json` from the
  byte-equality claim and bind it by a separately defined canonical form. Either
  way the §13 vector surface needs producible exact bytes.

### M2. "Collision-free by construction" fails on case-insensitive filesystems; the core §2 platform-collision guard is never invoked on the materialization write path

- Section: §5.3 (referenced layout), also §8.1 (managed-home layout).
- Quote: "Grouping by profile name makes two composed profiles carrying the
  same module filename collision-free by construction."
- What is wrong: profile-name comparison is case-sensitive (§1), so `Base` and
  `base` are distinct installable profiles and a legal composition chain
  (chain-repeat only forbids the *same* member). On default macOS/Windows
  filesystems `.agent-context/modules/Base/10-style.md` and
  `.agent-context/modules/base/10-style.md` are one platform path: the second
  write silently clobbers the first, the root file's two reference lines
  resolve to one file's bytes, and the recorded per-protocol-path surface
  hashes cannot both verify — silent corruption at write time surfacing later
  as false drift. The same collision hits §8.1 managed homes: profiles `Base`
  and `base` map to one `<environments-root>/<profile>/<env>` directory, two
  markers overwriting each other. Core §2 already contains the exact needed
  rule ("Filesystem extraction MUST detect two protocol paths that map to one
  platform path and fail before writing"), but it is stated for extraction and
  nothing in this document extends it to materialization writes — a guard that
  exists but is uncalled on this path. The ledger cannot catch it either:
  §8.3's unmanaged-conflict check is protocol-path-based, so the colliding
  write targets a "different" recorded path and no refusal fires. Note the
  `modules/` literal guard against the `system-prompt.md` sibling is correct;
  it is the per-profile-grouping uniqueness argument that holds only in
  protocol-path space.
- Suggested fix: one normative sentence in §5 (covering §5.3 and §8.1): any
  materialization or provisioning step that would write two protocol paths
  mapping to one platform path MUST detect the collision and fail before
  writing (core §2 rule extended to materialization), with a stable diagnostic
  (existing `environment_composition_invalid` for chain-induced collisions, or
  a dedicated code). Then drop or qualify "by construction".

## Minor

### m1. §7.2 failure path has no stable diagnostic code

- Quote: "Requesting `referenced` for an adapter that does not support it is a
  configuration error, not a fallback case."
- What is wrong: this is the one failure path in the document with no code —
  §7.7 does not cover it and `environment_form_unavailable` is explicitly the
  *other* case (warning + fallback). Violates the document's own
  every-failure-path-has-a-code discipline.
- Suggested fix: add a code (e.g. `environment_form_unsupported`) to the §7.7
  table.

### m2. `--append-system-prompt-file` / `--system-prompt-file` spellings presented as closed fact without the verification caveat the document gives its other docs-confidence data

- Section: §7.3 table, §10.2 example.
- What is wrong: Decision 0010 records `--system-prompt`/`--append-system-prompt`(-file)
  as verified in claude_code 2.1.251; the document narrows to the `-file`
  variants only and states them as closed normative declarations. §7.6 shows
  the right pattern for exactly this situation — "recorded from vendor
  documentation… verify against a pinned Xcode release before the conformance
  vectors freeze" — but §7.3 carries no such note. As written, a property
  inferred from a proxy (the fragment hands over a path, so `-file` flags
  "must" exist) is reported as established instead of as pending verification.
  The producer's board notes flag it, but board notes are not the normative
  document; the caveat must ride where the claim rides.
- Suggested fix: add a §7.6-style pre-freeze verification sentence to §7.3 (or
  verify the spellings against the pinned releases now and cite them).

## Nits

- n1. §6 and §11 carry inline unnumbered diagnostics tables while every other
  section uses a numbered `### N.M Diagnostics` subsection — align.
- n2. §8.4 "a read failure is `environment_marker_invalid` or an I/O error" —
  using the marker-invalid code for a *surface-file* read failure is loose;
  consider naming the intended code (or `unknown`) for surface reads.

## Deliberate-deviation verdicts (producer notes "Additional normative resolutions")

| Deviation | Verdict |
| --- | --- |
| Header `generated:` URL = spec repo | Sound. URL matches the actual origin remote. Stays an open item for the orchestrator before vectors freeze, as the producer flagged. |
| System-prompt output without header/chapters | Sound. The decision's header requirement is scoped to root-context files; injecting generated bytes into a system prompt would change model behavior; provenance via §5.6 marker hash is coherent. |
| Flat non-recursive composition | Sound. Narrows the decision, contradicts nothing, and removes cycles by construction. |
| Marker-as-ledger split from core §11 adapter ledger | Sound. Skills keep the frozen `.csk-managed.json`; environment surfaces get the marker; the two never merge — consistent with Decision 4's "alongside the generalized ledger". |
| `system_prompt_files: off\|append\|replace`, default `off` | Sound. Concretizes the setting Decision 6 already names; `off` default preserves "plain launch carries no active system prompt". |
| opencode skills kept at manager §5 native surface | Sound. Follows the decision's own open-question-3 recommendation; the profile-isolation consequence is explicitly surfaced. |
| §12 shadow-inert = non-current | Sound but keep visible: stricter than decision §3's warning framing; defensible (the tool objectively does not read the managed surface). Orchestrator may flip for `--check` ergonomics. |

## What was verified (evidence)

- **Base/signature**: head `eddd509` carries a good git signature by the same
  ECDSA key (`SHA256:V6JiKG…81cM`) that signed main's `2a861e5`; base is
  exactly `2a861e5` (= merge-base with `origin/main`); delta is
  `protocol/environments.md` only; CHANGELOG and everything else untouched.
- **Coverage vs Decision 0010**: every revision-1 row of the decision's §11
  phasing table has normative prose (profiles/sources §1, repo shape §2,
  manifest §3, store §4, materialization §5, composition §6, registry §7
  incl. channels/passthrough/shadowing/secondary targets, modes+marker §8,
  lifecycle incl. onboarding subset §9, resolve+fragment §10, umbrella §11,
  status/GC §12, conformance §13). No out-of-scope leakage: launcher
  internals, MCP write, settings, path-kind mechanics, schema files all
  explicitly excluded.
- **Cross-references**: all cited sections read in the worktree and match —
  core §1 (strict parsing), §1.1 (frozen identifiers), §2 (identifier grammar
  `^[A-Za-z0-9][A-Za-z0-9._-]*$`, portable paths, platform-collision rule),
  §5 (one-of tag/branch/revision, direct-declaration branch allowance),
  §6.1/6.2/6.3 (canonical identity, git safety and branch resolution, tag
  grammar), §7 (closure), §8 (content hash — §5.6's use is well-formed),
  §9.4 (GC live roots, fail-safe), §10 (moved-tag/strict-tag; marker not an
  authorization token; repair never adopts candidate bytes), §11
  (fail-rather-than-overwrite), §12.3 (closed-set admission); manager §1, §2,
  §2.5 (mutation lock + journal), §5 (adapter table, unknown-identifier
  warning, symlink-with-copy fallback, opencode native `~/.agents/skills`),
  §6 (MCP read-only), §7 (always-strict audit machinery), §10 (recompute-and-
  report, `--check` non-zero).
- **Byte rules attacked**: assembled monolithic, composed, referenced, and
  zero-module outputs by hand from the part rules — joins, trailing-LF, and
  chapter bytes are unambiguous; the `modules/` literal guard against the
  `system-prompt.md` sibling is correct; the two survivors are M1 and M2
  above. Header grammar is closed and pin shapes cover both source kinds.
- **Diagnostics**: code families complete and non-overlapping against every
  failure path in the prose except the §7.2 gap (m1); no code without a
  defined condition.
- **Repo consistency**: no pre-existing `schemas/v1/` schema collides with the
  new surfaces; `tools/validate.py` rerun by this reviewer in a fresh venv at
  head `eddd509`: exit 0, "validated 53 schemas and 691 vector files"
  (spot-check of the producer's `make validate` claim; docs-only delta).
- **Gate attacks on the normative model**: profile-influence boundary (names
  from registry only, values inside environments root, name-segment bounded);
  marker fail-closed on unreadable; backup never overwritten; foreign-manager
  stop; repair never adopts candidate bytes; credential-file no-touch rule;
  umbrella dispatch input = argv+PATH only. No bypass found beyond M2.

## Routing

Blocking per the review brief's rule (major present): task goes back to the
producer. Fix M1 and M2 (both are localized additions to §5/§7.7), address
m1/m2, take or decline the nits, then hand back to review. The CR revision was
NOT accepted; `accept_cr` was not called.
