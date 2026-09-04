# Onboarding import + §9.1 ref selection — implementation notes

Task: TASK-260901-3gr1fe. Branch `draft/environments-onboarding-import` in
`~/Developer/ReluxWorks/.worktrees/curator-spec-onboarding-import`, based on
fresh `origin/main` = 62e592a. One signed commit: f8d7e7a (ECDSA
SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM, verifies against
`maintainers.allowed_signers`). Not pushed, per brief.

## Scope item 1 — `path` source kind + onboarding import

### Decisions

1. **`path` folded into revision 1, not a revision 2.** Decision 0010 D5
   targets the import "between revisions 1 and 2" — that is delivery
   sequencing of authoring work, not a wire-revision mandate. The
   environments capability has never been claimed by an implementation and
   is pinned by no tagged release; a revision bump for `path` alone would
   fork the frozen generation-header bytes ("environments revision 1") and
   the vector capability identity for zero compatibility benefit. Revision
   1's closed source-kind set is now `git`/`local`/`path`, and the §9.5
   deferral text plus the manager §12.3 MUST-reject rule are removed.

2. **`path` operand shape.** The operand names a repository-shaped
   directory (root `Profilefile.json`, §2/§3 unchanged), not a bare profile
   directory — the ordinary pipeline stays uniform and the
   authoring-before-repo use case (D1) later converts to `git` install with
   no reshaping. Recognition is syntactic (`/`, `./`, `../`, platform
   absolute), never probed from the filesystem, so an scp-form git operand
   can never be shadowed by a same-named local directory.

3. **Snapshot semantics.** Copy-on-install, never re-read, never network.
   Tree discipline reuses core §6.2 (no links/special files/platform
   collisions). Root `.git` excluded from the snapshot (authoring case);
   `.git` below the root rejected (submodule-shaped, unsupported). Pin =
   core §8 content hash = state hash, exactly the `local` pin shape.

4. **Audit identity for a `path` profile.** Always-strict manager §7 audit
   applies unchanged, `context-secret-material` included. No network
   identity exists: local revocation keys off the state hash, the core §6.1
   allowlist does not apply (local sources bypass it), and a `path` snapshot
   never produces a shared `audit-record-v1` (its shape requires network
   identity + commit). Stated normatively in §9.1 and echoed in manager
   §12.6.

5. **Detected-surface list (closed, revision 1).** Per registered adapter,
   over its native default home: the §7.1 root-context file, plus each
   global skills entry the adapter ledger does not record (ledgered entries
   reach managed state via the §9.4 migration, never import). Secondary
   fixed-home targets join inventory/backup but contribute no surface; a
   divergent unmanaged secondary-target root file is a named lossy finding
   (its distinct bytes would not carry over; backup preserves them).

6. **Loss-list definition.** Lossless iff every detected surface maps:
   root-context file maps when readable and valid UTF-8; skills entry maps
   when a complete exact declaration is recoverable from the entry's own
   records (valid core §10 install marker, or a git checkout whose `origin`
   canonicalizes under core §6.1 with a clean committed HEAD). Losses:
   unreadable file, non-UTF-8 root context, undeclarable skill, divergent
   secondary-target root file. Loss entries name adapter + platform path +
   reason. Absence ≠ failed read (§8.4 discipline): absent surfaces never
   enter the loss list; a failed read is always a loss, never absence.

7. **Reassembly determinism.** One module `context/<env-id>.md` per adapter
   with a detected root file, selector `environments: ["<env-id>"]`, class
   `root`, manifest in ascending environment-identifier order. Line-ending
   normalization (CRLF/CR→LF, exactly one trailing LF) happens only at
   reassembly — §3's no-normalization rule for snapshot modules is
   untouched — and is defined as content-preserving, so it does not make an
   import lossy; original bytes are already in the §9.5 backup. Default
   profile name `imported`, operator-overridable under core §2; a taken
   name stops before any write.

8. **Skill migration.** Each mapping entry reproduces its recovered
   declaration (install marker's declared ref preferred, else `revision`
   pin to the checkout's committed HEAD) and is reported with
   `environment_import_skill_foreign` — managed by other means, SHOULD
   re-declare from upstream (tag/branch) for updates. Matches D5 wording.

9. **Consent gate.** Lossless proceeds; lossy stops with
   `environment_import_lossy` + loss list; proceeds only under an explicit
   per-operation consent flag (re-reported as warnings, same code). Machine
   configuration MUST NOT pre-record consent. Authentication untouched
   end to end.

### Diagnostics added

- `profile_source_path_missing` / `profile_source_path_unreadable` (§1.1;
  absence and read-failure kept distinct by explicit rule)
- `profile_source_invalid` widened: ref on a `path` declaration,
  non-directory operand, snapshot-tree violations
- `profile_install_ref_conflict` (§9.7; >1 ref flag, or ref flag with path
  operand)
- `environment_import_lossy` (stop without consent / warning with consent)
- `environment_import_skill_foreign` (warning)
- `profile_import_name_taken`

## Scope item 2 — §9.1 install ref selection

Took the proposed shape unchanged; nothing strictly better found.
Install-level `--tag | --branch | --revision`, exactly one, mapping 1:1
onto the §1 declaration forms (core §6.3 tag grammar, full lowercase
commit, core §6.2 branch resolution). Default with no flag: track the
remote default branch = the branch the remote's `HEAD` symref names at
resolution time; a remote advertising none is `profile_source_invalid`
(closed rule instead of a guessed branch name). Resolved commit recorded
as the effective pin; `--strict-tags` semantics unchanged. Whole-snapshot
scope stated with the by-construction argument (Profilefile names sibling
dirs of one snapshot); revision 1 rejects mixed refs via
`profile_install_ref_conflict`; two-commit coexistence of one repo's
profiles explicitly unsupported (supported shape: separate repositories).
Aligned with core §6 grammar throughout.

## Marker schema / vector delta + evolution rationale

**Decision: in-place evolution of `agent-environment-marker-v1` schema 1 is
admissible; no v2.** Rationale: COMPATIBILITY.md's in-place prohibition
("a release never redefines old schema bytes in place"; new
`schema_version` for incompatible wire changes) governs published schema
series. The marker schema landed in cef93fb, after the live rc.9 release
pin; no tag, release manifest, or implementation claim pins its bytes
(`release/1.0.0-rc.9.json` predates and does not reference it), so there
is no deployed reader to protect and a v2 fork would be pure ceremony.
The same reasoning covers extending revision 1's closed source-kind set.

Schema delta: third `profile` oneOf branch — `source_kind: "path"`,
required `source_path` (nonEmptyString; recorded verbatim, informative
provenance, never enters identity), required `state_sha256` (hex256, same
pin def as `local`), optional `imported_from_native` (`const: true`,
present exactly for §9.6-created profiles; `false` is invalid rather than
ambiguous).

Fragment schema **unchanged deliberately**: `pinnedProfile`'s
`state_sha256` branch already carries a `path` profile's pin, and §10.2 now
states the fragment is pin-only (no source-kind/source-path record — a
consumer needs the pin, not the provenance).

Determinism vectors **unchanged deliberately**: the §5.1 pin grammar stays
closed at `commit <hex>` / `state sha256:<hex>`; a `path` profile uses the
state spelling, and header/materialization vectors carry pin strings, not
source kinds — a "path" header case would be byte-identical in class to
the existing `local-state-pin` case and prove nothing new. Stated in §5.1.

Schema-cases added (generated via `tools/generate-vectors/environments.go`,
so they survive regeneration): `valid-path-profile`,
`valid-imported-path-profile`, `invalid-git-with-source-path`,
`invalid-path-with-commit`, `invalid-path-with-ref`,
`invalid-path-missing-source-path`,
`invalid-path-imported-from-native-false`.

`release/1.0.0-rc.9.json` candidate manifest pin advanced with the suite
manifest — generator-owned behavior with precedent (cef93fb did the same).

## Consistency-only edits outside environments.md (per brief)

- `profiles/manager.md` §12.3: install sentence gains the ref-selection
  clause; the deferral paragraph and the `path`-MUST-reject diagnostic row
  replaced with the §9.6 pointer and the two new diagnostics; §12.6 gains
  the path-audit-identity sentence; stale "revision-1 onboarding subset"
  wording fixed.
- `cli/curator.md`: three example lines (install-level `--tag`, path
  install) with two comment lines.
- No CHANGELOG (per brief; surface unreleased). No push.

## Validation evidence (real exit codes)

- `make validate` (python validate.py + 147 unittests + `go test
  ./tools/...`): exit 0, "validated 57 schemas and 773 vector files".
  Python 3.14.7 with jsonschema 4.25.1 from a scratchpad venv (system
  python lacks jsonschema; venv PATH-prefixed for make).
- Generator determinism: `go run ./tools/generate-vectors -root .` twice;
  sha256 inventory of `conformance/v1` + `release` diff-clean between runs
  (exit 0, `GENERATOR-DETERMINISTIC`).
- Negative-evidence mutations (both restored afterwards, final validate
  exit 0):
  - loosening `imported_from_native` to `{"type": "boolean"}` →
    validate.py exit 1 on `invalid-path-imported-from-native-false`
    (narrowing mutant caught, not just delete);
  - deleting the path oneOf branch → validate.py exit 1 on
    `valid-path-profile` (gate reachable).
