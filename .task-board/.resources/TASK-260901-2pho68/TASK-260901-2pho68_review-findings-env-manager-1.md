# Review findings: manager-profile environments sections (round 1)

Reviewer run: RUN-260901-b19700 (attempt 4; RUN-260901-ef9251/9ccb6e/e31910 died on
provider API errors — `terminal_reason: api_error`, ENOTFOUND — with no verdict; nothing
carried over from them).

Subject: CR-TASK-260901-2pho68-1 rev 1 (curator repo, base `eb32105d`, candidate tree
`dac26b10`, delta `LOGBOOK.md` only) + the actual spec work in
`~/Developer/ReluxWorks/.worktrees/curator-spec-env-manager`, branch
`draft/environments-manager-profile`, head `6697c1e`.

## Verdict: CHANGES REQUESTED (one blocking finding) → to-dev

## Finding 1 (blocking): the CR delta corrupts a prior LOGBOOK entry

The only repository change in CR rev 1 deletes the heading line
`## 2026-08-28 — Authoring CLI commands documentation refresh (TASK-260827-21xw9d)`
instead of inserting the new 2026-09-01 section above it. In the candidate tree the
three bullets of that prior entry (`**Documentation created**`, `**README integration**`,
`**Validation & evidence**`) are left orphaned directly under the new
`## 2026-09-01 — ... (TASK-260901-2pho68)` heading, so TASK-260827-21xw9d's milestone is
misattributed to this task and its date/ID heading is destroyed. The logbook is the
institutional record ("one block per insight"); the CR's sole delta damages it.

Verified against both trees:
- base `eb32105d:LOGBOOK.md` has the 2026-08-28 heading with its bullets;
- candidate `dac26b10:LOGBOOK.md` lacks the heading, bullets merged into the new entry.

Fix (exact, one line): re-insert
`## 2026-08-28 — Authoring CLI commands documentation refresh (TASK-260827-21xw9d)`
followed by a blank line above the `- **Documentation created**:` bullet, keeping the
new 2026-09-01 entry (whose own content is accurate) on top. No other change requested.

## Everything else: verified clean (no further findings)

Checked myself, in the curator-spec worktree at `6697c1e`:

1. **Git facts** — head `6697c1e1a721`, parent = base = `origin/main` = `c3b29b1f`,
   working tree clean, delta exactly `profiles/manager.md` (+276) and `cli/curator.md`
   (+37/-1). Signature verifies `G` against `maintainers.allowed_signers`
   (principal `oparin@me.com`). `schemas/`, `conformance/`, `protocol/`, `CHANGELOG.md`
   untouched; not pushed.
2. **Consistency with environments.md** — every §12 claim traced to the normative text:
   §7.1 four adapters + opencode XDG seed-link maintenance; §7.2 form defaults and
   `environment_form_unsupported` as config error; §5.3 fall-back-to-monolithic with
   `environment_form_unavailable`; §7.5 shadowing warn; §7.6 exactly two
   `xcode-coding-assistant` rows, auto|off|explicit probing, embedded hosts' files
   unmanaged; §8.1 mode defaults; §8.2 marker fails closed, not-a-signature; §8.3
   backups never overwritten/hashed/GC'd; §8.4 drift and absence-vs-unreadable; §9.1
   install-every-declared-profile + activation-without-magic + always-strict audit +
   `context-secret-material`; §9.2 lock/journal/warning; §9.3 split-brain visibility;
   §9.4 `default` local-kind migration; §9.5 five-step onboarding subset, `path`-kind
   rejection (§1); §10.1 pure resolution + verify-and-repair + no-adoption; §10.3
   influence boundary; §12 status matrix/`--check`/GC live roots; §13 no-partial-claim.
3. **No byte-rule duplication** — spot-checked the riskiest: generation header, chapter
   parts, referenced-form layout, `opencode.json` CCJ-1 bytes, and the platform-path
   collision rule are all absent from §12 (referenced only). Marker/backup filenames and
   diagnostic code names are identifiers, not byte rules.
4. **Diagnostics tables** — all 16 cited codes exist in environments.md; each table's
   section attribution (§5.7/§7.7, §8.5, §1.1/§9.6, §10.4) is correct.
5. **Internal manager.md references** — §2.5 (home lock + journal), §2.6 and §11.9
   (no-adoption of candidate bytes), §5 (unknown-identifier warning,
   symlink-with-copy-fallback, gemini/cursor/windsurf rows), §7 pointer placement,
   §10 (read-only status, GC fail-safe) all check out; §12 follows the §11 top-level
   "manager profile" precedent and the house table style.
6. **cli/curator.md** — informative tone kept; command rows match the normative
   synopses (`profile install <git-url> [--use [name]]`, `use [--env|--target]`,
   `env resolve [--profile] [--format json|env|shell]`, `env status [--check] [--json]`,
   `global ... [--profile|--all-profiles]`); umbrella note matches §11 exactly
   (curator-run/curator-session as informative first providers, fail-with-exact-name,
   nothing implicit); the example block runs mentally against the synopses. Anchor
   `#12-agent-environments-manager-profile` derives correctly from the heading (same
   convention as the existing §11 link).
7. **Links validation** — reran `tools/validate.py` myself via the worktree venv:
   exit 0, "validated 53 schemas and 691 vector files" (this component includes
   `validate_local_links`; note it strips `#fragments`, so the anchor was verified
   manually). Unit tests / `go test ./tools/...` were not rerun — docs-only delta;
   producer evidence accepted for those.

## Judgment on the filed normative gap: REAL, keep it filed

environments.md §1 requires a direct `git` profile declaration to carry exactly one of
`tag`/`branch`/`revision`, and §12 (status) reports a "declared ref" per installed
profile — but §9.1's `profile install <git-url>` gives the operator no way to express
that ref and no default, and §2's `Profilefile.json` carries no source declaration
(names → in-repo paths only). No other section supplies the mechanism. The gap is real,
correctly filed in `TASK-260901-2pho68_env-manager-notes.md` rather than fixed (protocol
prose belongs to the sibling story), and the CLI row correctly stays generic instead of
inventing a flag. It remains a backlog item for the protocol story.

## Scope note

Ignored per brief: `tools/__pycache__`, `.temp/venv`. The spec worktree itself needs no
rework; only the curator-repo LOGBOOK.md line above does. After that one-line fix the
work is acceptance-ready.
