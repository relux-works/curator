# Review verdict: onboarding import + §9.1 ref selection — ACCEPT

Task TASK-260901-3gr1fe, CR `CR-TASK-260901-3gr1fe-1` rev 1.
Subject: curator-spec worktree `~/Developer/ReluxWorks/.worktrees/curator-spec-onboarding-import`,
branch `draft/environments-onboarding-import`, head `f8d7e7a`, base `62e592a` (verified
`62e592a` is an ancestor of head; one signed commit, ECDSA
`SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`, "Good git signature" against
`maintainers.allowed_signers` — verified by this reviewer, not taken from the notes).
Board CR delta in the curator repo is the LOGBOOK.md entry only; its claims were
checked against the spec worktree and reproduce.

## Priority check 1 — `release/1.0.0-rc.9.json` delta: BENIGN, verified mechanically forced

The 4-line delta changes exactly two fields, `candidate_protocol_pin.manifest_sha256`
and `downstream_consumption.required_manifest_sha256`, both `43e31b26… → 90ee8047…`.
Established facts:

- The immutable record of rc.9 is the `v1.0.0-rc.9` tag snapshot (target `0ed5c69`),
  which pins `803918bf…` — untouched. The in-tree file had **already** diverged from
  the tagged bytes on main: `cef93fb` (landed) advanced it `803918bf… → 43e31b26…`.
  The `historical_release` block (immutable rc.8 record) is byte-identical.
- The file is generator-owned: `tools/generate-vectors/main.go:2109` writes
  `release/1.0.0-rc.9.json`, and the Makefile `regenerate-check` target diffs it
  together with `conformance/v1`. Reverting the 4 lines would make `regenerate-check`
  fail by construction.
- I regenerated independently: `shasum -a 256 conformance/v1/manifest.json` after
  `go run ./tools/generate-vectors -root .` = `90ee8047…` = the committed pin value.
  The new pin is exactly the digest of the updated suite manifest.
- GOVERNANCE ("Release tags are immutable") and COMPATIBILITY ("without changing one
  byte of rc.8") govern tags and superseded metadata; neither is touched. The live
  candidate pin advancing with the suite is the file's designed behavior.

Not blocking. It is a fair process observation that a *candidate* pin living inside a
file named like a release record invites exactly this review question every time —
but that is pre-existing repo design (cef93fb precedent), not this delta's defect.

## Priority check 2 — path-in-revision-1 promotion: JUSTIFIED, consistently executed

Decision 0010 D1/D5 say `path` "is delivered with the onboarding import story, not in
revision 1" / "own story between rev 1 and 2". This run *is* that story — the named
delivery vehicle executing. The "not in revision 1" clause was authoring sequencing:
I verified no tagged release contains the marker schema at all (`v1.0.0-rc.10:schemas/
v1/agent-environment-marker-v1.schema.json` does not exist; `cef93fb` postdates the
rc.10 tag target), no implementation claim exists (`verified_implementations: []`,
`claims_emitted: []`), so widening the revision-1 closed source-kind set breaks no
deployed reader and forks no frozen bytes; a revision bump would change the frozen
generation-header identity for zero compatibility benefit. The promotion was also
explicitly mandated by the producer brief. Consistency verified: the §1 reserved
paragraph, the §9.5 deferral paragraph, and the manager §12.3 MUST-reject rule +
diagnostic row are all removed; `git grep` at `f8d7e7a` over protocol/profiles/cli
finds no stale "not in revision 1" / "revision-1 onboarding subset" / deferral text;
`profile_source_kind_unsupported` survives only for genuinely unknown kinds. The
preamble now also forbids partial source-kind support (closes the git-only subset
loophole). See minor finding M3 about the decision document itself.

## Priority check 3 — schema-1 in-place evolution: SOUND; gate attacked and held

Rationale verified independently (above): no tag, release manifest, or claim pins
`agent-environment-marker-v1` bytes; COMPATIBILITY.md's in-place prohibition protects
published series and deployed readers, of which there are none. A v2 fork would be
ceremony. In-place evolution is purely additive (a third `oneOf` branch keyed by
`source_kind: const "path"`): every previously valid instance stays valid, and no
previously invalid git/local case can leak into the new branch (const discriminator).
All 773 vectors including the pre-existing marker cases pass.

Gate attacked, not read — three mutants run by me against `tools/validate.py`
(producer ran two of these; the widening mutant is mine):

- loosen `imported_from_native` `const: true` → `{"type": "boolean"}`: exit 1 on
  `invalid-path-imported-from-native-false` (narrowing-class mutant caught);
- widen the path branch `additionalProperties: false → true`: exit 1 on
  `invalid-path-with-commit` (closure is load-bearing, not decorative);
- drop `source_path` from `required`: exit 1 on `invalid-path-missing-source-path`.

Schema restored, final `validate.py` exit 0. The new branch is closed and strict:
commit on a path profile, ref on a path profile, source_path on a git profile,
missing source_path, and `imported_from_native: false` are all rejected by named
cases wired through `schema-cases/index.json`, generated by
`tools/generate-vectors/environments.go` so they survive regeneration.

## Priority check 4 — normative quality: GOOD

- **Path recognition** is syntactic and unshadowable: only `/`, `./`, `../` (or a
  platform absolute spelling) select `path`; an scp-form git operand
  (`git@host:path`) starts with none of these and can never be shadowed by a
  same-named local directory; no filesystem probe participates in classification.
- **Import flow** concretizes D5 without contradiction: closed detected-surface list
  (root-context file + unledgered global skills entries; ledgered entries routed to
  the §9.4 migration), lossless/lossy with a named loss list, per-operation consent
  gate with MUST-NOT-pre-record, deterministic reassembly (ascending env-id order,
  content-preserving CRLF/CR→LF + single trailing LF, applied only at reassembly so
  §3's no-normalization rule is intact, originals in the §9.5 backup), skill
  migration via recovered exact declarations with `environment_import_skill_foreign`,
  authentication untouched end to end, and the assembled directory going through the
  ordinary §9.1 path pipeline including the always-strict audit.
- **Absence vs failed read** (§8.4 discipline) is carried into both the §1.1
  missing/unreadable split and the loss-list rules explicitly.
- **Diagnostics** are closed and owned: new codes live in environments §1.1 and
  §9.7; the manager §12.3 table rows are pointer-style repeats referencing
  "(environments §1.1, §9.7)", the pre-existing pattern.
- **§9.1 ref selection** matches the twice-filed gap's asked shape: exactly one of
  `--tag|--branch|--revision`, 1:1 onto the §1 declaration forms (core §6.3 tag
  grammar, full lowercase commit, core §6.2 branch resolution), default = the branch
  the remote's `HEAD` symref names (a remote advertising none is
  `profile_source_invalid` — closed, no guessed branch name), resolved commit = the
  effective pin, `--strict-tags` unchanged, whole-snapshot scope with the
  by-construction Profilefile argument, mixed refs rejected via
  `profile_install_ref_conflict`, two-commit coexistence explicitly unsupported.
  Aligned with core §6 and with the Skillfile direct-declaration branch allowance.
- **Marker vs fragment**: marker records provenance (`source_path` informative,
  never identity; `imported_from_native` exactly for §9.6 profiles); fragment stays
  pin-only with the rationale stated in §10.2. Determinism vectors unchanged is
  right: the §5.1 pin grammar is closed at `commit`/`state` spellings and a "path"
  header case would be byte-identical in class to `local-state-pin`.

## Priority check 5 — validation: REPRODUCED INDEPENDENTLY

On a clean export of tree `f8d7e7a` (scratchpad, python 3.13 venv + jsonschema
4.26.0, go 1.25.5): `make validate` exit 0 — "validated 57 schemas and 773 vector
files", 147 unittests OK, `go test ./tools/...` ok. Generator run twice; sha256
inventory of `conformance/v1` + `release` byte-identical between runs
(GENERATOR-DETERMINISTIC), and regeneration reproduces the committed tree exactly.

## Priority check 6 — scope hygiene: CLEAN

15 changed paths, all in scope. manager.md changes are the noted minimal
consistency edits (§12.3 ref-selection sentence + import paragraph + table rows,
§12.6 path-audit sentence, stale subset wording); cli/curator.md gains three example
lines + comments; no CHANGELOG, no push, reviews/ untouched, no drift elsewhere.

## Minor findings (non-blocking, next-editor notes)

- **M1** §1 defines a `path` operand as absolute, "or project-relative when the
  operation runs inside a project". A `./x` operand *outside* a project is
  syntactically recognized as `path` by §9.1 but no rule names its outcome; the
  nearest fit is `profile_source_invalid`. One sentence would close it.
- **M2** §9.6 makes a secondary-target root-context file lossy "exactly when its
  bytes differ from the same adapter's detected native root-context file". When the
  primary native root-context file is absent, the comparison is undefined while the
  secondary's bytes still would not carry over; by intent that should be lossy, but
  the letter presupposes a detected primary file.
- **M3** Decision 0010 still carries the now-superseded sequencing sentence ("not in
  revision 1", "own story between rev 1 and 2"). As a decision record that is
  arguably correct-as-written history, and the normative surface is environments.md;
  but a one-line delivery note in 0010 (or the story's closing record) would spare
  the next reader the same reconciliation this review had to do.

## Verdict

ACCEPT. All six priority checks pass; the presumptively-blocking release-manifest
concern is proven benign and mechanically forced; the gate held under mutants the
producer did not run. Minor findings M1–M3 are recorded for a future editing pass
and do not warrant a rework cycle. Recording via `accept_cr(TASK-260901-3gr1fe,
revision=1)`; `done` transition and `commit_ack=scope_committed` belong to the
orchestrator.
