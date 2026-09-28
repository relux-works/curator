# TASK-260924-291k0q results — Revision 3 — rc.13 bytes preserved

## What changed in this revision (rework-2)

- Reverted `schemas/skillfile-sources-v1/README.md` to the `v1.0.0-rc.13` bytes:
  `git diff v1.0.0-rc.13 -- schemas/skillfile-sources-v1/README.md` is empty.
- Moved the rc.10-baseline statement and the single conditional environments
  §9.4 exception into `COMPATIBILITY.md` (new section
  "Partial-client baseline: rc.10 core plus skillfile-sources-v1").
- Everything else from revision 2 is unchanged: the gate
  (`tools/verify_skillfile_sources_independence.py`), its narrowing tests
  (`tools/test_skillfile_sources_independence.py`, 8 tests), the Specification
  CI job, and the Unreleased CHANGELOG entry.

## Unpinned-document choice (manifests checked, not assumed)

- `schemas/skillfile-sources-v1/README.md` IS digest-pinned in
  `conformance/skillfile-sources-v1/manifest.json` (also at the rc.13 tag) and
  covered by `validate_skillfile_sources_manifest()` — editing it breaks the
  rc.13 suite pin. Reverted.
- `conformance/skillfile-sources-v1/README.md` is pinned in the same manifest —
  not usable either.
- `COMPATIBILITY.md` is outside every manifest scope (skillfile manifest covers
  only `protocol/skillfile-sources.md`, `protocol/repository-transport.md`,
  `docs/skillfile-sources.md`, `schemas/skillfile-sources-v1`,
  `conformance/skillfile-sources-v1`; the core `conformance/v1` manifest covers
  neither it nor top-level `conformance/README.md`). Chose `COMPATIBILITY.md`
  because it already hosts the "Accepted Skillfile source extension" and
  "Rc.13 release boundary" notes.

## Final tree vs v1.0.0-rc.13 (only intended deltas)

- `M .github/workflows/ci.yml` (+3: gate step after `validate.py`)
- `M CHANGELOG.md` (+11: Unreleased entry, kept from rev 2 per rework-2)
- `M COMPATIBILITY.md` (+15: baseline section)
- `?? tools/verify_skillfile_sources_independence.py` (gate, rev 2)
- `?? tools/test_skillfile_sources_independence.py` (gate tests, rev 2)
- Pinned files verified byte-identical to rc.13: both skillfile READMEs, both
  manifests, `release/1.0.0-rc.13.json`, `conformance/v1/manifest.json`.

## Evidence (each gate run as a standalone process, real exit codes)

| Check | Command | Result |
|---|---|---|
| Gate on candidate | `python3 tools/verify_skillfile_sources_independence.py` | exit 0 — `89/89 byte-identical to v1.0.0-rc.10`; `28/28 clauses present at rc.10; 21 citations, 1 exact conditional exception allowlisted` |
| Gate on rc.13 tree | gate `--root <detached worktree at v1.0.0-rc.13>` (worktree removed afterwards) | exit 0, same counts |
| Gate tests | `python -B -m unittest tools.test_skillfile_sources_independence` | exit 0 — 8/8 OK (pass-through, rc.12-only `$ref` mutant, whitespace-changed-definition mutant, rc.12-only manager-clause mutant, exact conditional allowlist, unconditioned same-citation mutant, informative-only-marker mutant, wrong-file conditional mutant) |
| Spec validation | `python tools/validate.py` (venv with `requirements-dev.txt`) | exit 0 — `validated 64 schemas and 1169 vector files` |
| Full tools suite, per-module (from `tools/`) | `test_validate` in 5 class groups | OK: 4 + 55 + 162 + 93 + 221 (manifest-scope group incl. `SkillfileSourcesSuiteManifestTests` green) |
| `test_release_gate` full module | base tree: 35 OK; candidate rerun: 35 OK | exit 0 (see anomaly note) |
| `test_allowed_signers` / `test_verify_release_commit,merge_policy` / `test_implementation_coverage` | per-module | 7 OK / 10 OK / 39 OK, exit 0 |
| Lint (new files; no lint gate in CI) | `ruff check` on both new tools files | clean |
| `go vet ./tools/...` (no Go changes) | — | exit 0 |
| NOT run as one shot | `unittest discover -s tools` full run | exceeds the ~10 min single-call budget (killed uncompleted; covered instead by the per-module runs above, same test set) |
| NOT run | `go test ./tools/...` | no Go code touched; `go vet` run instead |

Note: the system `python` does not exist on this host; all Python gates were
run with `python3` (CI uses `python`) and an isolated `/tmp` venv for
`jsonschema==4.25.1` + `ruff`. No venv or artifact was left in the worktree.

## Anomalies (found, explained, resolved in-tree)

1. Terminating the over-budget full-discover run orphaned in-flight mutant
   files from `ReadFailureVectorTests.run_main_with_vector`
   (`conformance/v1/vectors/environments-read-failure.json`,
   `conformance/v1/manifest.json`, `release/1.0.0-rc.13.json` — that helper
   rewrites all three with recomputed pins and restores them in `finally`,
   which the kill skipped). Restored all three to HEAD bytes with
   `git checkout HEAD -- <paths>`; final tree verified to contain only the
   intended deltas. Lesson recorded: never terminate that suite mid-run;
   use per-class splits.
2. Two early full-module `test_release_gate` runs showed transient failures
   (1F+1E, then 1E on `test_accepts_complete_rc13_artifact_set`) while a heavy
   `test_validate` run was still consuming the machine; the module is isolated
   (patches `ROOT` to a temp dir) and passed 35/35 on the base tree, 35/35 on
   the candidate rerun, and the named test passes solo. Treated as
   load-induced flake, unrelated to this change.

## Mutant coverage statement

- `$ref` mutant to an rc.12-only v1 definition fails the gate through the real
  entry point (`cannot read ... at v1.0.0-rc.10`); a byte-changed but
  JSON-valid definition fails (`definition differs`, sha256 pair reported).
- Prose mutant citing an rc.12-only manager clause fails through the real
  entry point; the exact conditional environments §9.4 sentence passes only in
  `protocol/skillfile-sources.md` with its condition intact (unconditioned,
  informative-only-marked, and wrong-file variants all fail).
- Bound: absence-vs-read-failure distinction is the gate's own design (absent
  baseline file fails closed via `GateError`, never treated as satisfied).

## CHANGELOG entry (for release prep)

```markdown
- Gate `skillfile-sources-v1` schema references and protocol citations against
  the `v1.0.0-rc.10` core baseline, preserving partial-client independence in
  Specification CI. Its sole conditional-citation exception is the exact
  environments §9.4 profile-lock sentence in `protocol/skillfile-sources.md`;
  the same citation without its condition and every other post-rc.10 citation
  fail the gate.
```

(The same text is also present under `## Unreleased` in the worktree
`CHANGELOG.md`, kept from revision 2 per the rework-2 "everything else stays"
instruction; the release-prep leaf owns the final wording.)
