# TASK-260916-2fu85y evidence — amend decision draft 0018 (config-driven permission mode, default yolo)

## Changed files (worktree, uncommitted)

- `decisions/0018-curator-run-permission-interface.md` — amended, status stays `proposed — not adopted`
- `UNRESOLVED_QUESTIONS.md` — "Filed proposals": 0018 bullet notes the amendment (still proposed — not adopted)
- `CHANGELOG.md` — Unreleased/Added bullet for the draft amendment (explicitly no normative change)

`git diff --stat`: 3 files, +194/−56. No other working-tree changes; no commits made.

## Amendment summary (operator decision 2026-09-16)

- Precedence per launch: CLI `--permissions` (or `--yolo`) > per-profile setting > launcher global default > built-in default `yolo` (untracked).
- Global default: `permissions` member on each env-id entry of launcher `defaults.json`; schema `curator-run-defaults-v2` (v1 + one optional member). v2 over v1.1 because SPEC §4.3 closes v1 ("readers MUST reject an unknown member"), so any new member breaks old readers regardless of label — the new token keeps the failure precise; v2 readers accept v1 files. Mirrors repo precedent (`manager-config`/`system-config` v1→v2).
- Per-profile setting: new environments §12.1 knob `permissions.<profile>` (`native|yolo`, default absent), resolved by Curator and delivered through the `env resolve` fragment — not a launcher-file entry, because §12.1 forbids knobs in implementation-private files, profiles are Curator's naming authority, and SPEC §4.7 forbids the launcher from duplicating §12.1 values (it reads every launch-shaping knob only through the fragment). No launcher section in machine configuration (the §4.7 open item stays open).
- Fleet-wide force-`native`: `permissions` joins the §12.2 lockable set, lockable only toward `native` (mirrors `isolation`→`shared`); system file names `environments.permissions` under manager §1 rules. Sits above the whole precedence: visible `yolo` (flag/`defaults.json`) under an engaged lock is a `usage` error naming the locked knob; total silence resolves `native` (the lock redefines the effective default on that machine).
- Provenance: `curator-run: permissions=<native|yolo> source=<flag|profile|global|default> mapped=<flag or none>`.
- Tracked mode: `yolo` from any level refused (`permission_mode_tracked_unsupported`); the built-in default does not cross into tracked mode (tracked silence resolves `native`, `source=default`). Untracked headless/CI has no detector in this proposal — explicit configuration or the lock; recorded as new open question 7 together with the fragment version-skew bound.
- Replaced the draft sentence "defaults.json and every config surface can never set yolo" (items 2–3); the mode is still never inferred from model/prompt/credentials/environment/`ax`. Refusal table and per-environment tool mapping unchanged. Bypass-by-default recorded as an explicit operator choice scoped to interactive untracked launches on operator-owned machines.

## Validation

`make validate` run with a project-local `.venv` (`python3 -m venv .venv`, `pip install -r requirements-dev.txt`, `PATH="$PWD/.venv/bin:$PATH" make validate`); the venv was removed afterwards so the Change Request snapshot contains only the three files above.

- `python3 tools/validate.py` → `validated 60 schemas and 1047 vector files`
- `python3 -B -m unittest discover -s tools -p 'test_*.py'` → `Ran 227 tests ... OK` (275.7s)
- `go test ./tools/...` → `ok .../tools/generate-vectors 0.581s`
- `make validate` exit code: **0**

No behavior-bearing code changed (decision prose + list/changelog entries only), so no new tests were added; the repo's own suite above is the verification gate.

## Acceptance-criteria mapping

- decisions/0018 updated (still proposed) — done, this change.
- `make validate` green — done, exit 0 quoted above.
- PR landed after independent review; issue #55 updated — orchestrator-owned (branch → PR → hosted checks → fast-forward → board closure per campaign rules); the developer handoff publishes the Change Request for independent review. Not performed in this role.

## Findings

- Version-skew bound: profile level and lock engagement reach the launcher only via the fragment; a pre-adoption Curator reports neither, resolving as profile-silent (bypass past an unenforced lock under default-yolo). Recorded in the decision as open question 7 with the operator obligation to upgrade both sides together until the adopting revision answers. No logbook entry beyond this artifact: the finding is captured normatively in the draft itself.

---

## Revision 2 (rework after independent review, same task)

Reviewer verdict on revision 1: CHANGES_REQUESTED with two High
findings (`TASK-260916-2fu85y_review-verdict.md`): (1) untracked
headless/CI silence inherited the built-in `yolo` default; (2) a
legacy fragment lacking policy/lock support resolved as silence,
yielding `yolo` past an unenforced force-`native` lock. Both are
corrected in `decisions/0018-curator-run-permission-interface.md`
(status stays `proposed — not adopted`):

- Item 5 rewritten: the built-in `yolo` default applies only to
  interactive untracked launches on operator-owned machines. A launch
  is headless when stdin/stdout is not a TTY, the native args select a
  non-interactive form (`-p`/`exec`-style), a non-interactive marker
  (`CI`, `GITHUB_ACTIONS`, or an equivalent the adopting revision
  enumerates) is present, or the launch is tracked. Headless/CI/tracked
  silence resolves `native`; untracked headless/CI may still select
  `yolo` explicitly (flag/profile/global, subject to lock, mapping,
  and refusals); tracked `yolo` stays refused
  (`permission_mode_tracked_unsupported`) until the `ax` capability
  admits it.
- Item 5 legacy transport: verified policy/lock transport support is a
  precondition for admitting `yolo`. A fragment predating the adopting
  revision is not silence: would-be `yolo` from any level — including
  the flag, because the item-3 lock sits above it and an invisible
  lock cannot be enforced except by refusal — is refused
  (`permission_policy_unsupported`); `native` resolutions proceed.
  Legacy transport yields `native` or a refusal, never `yolo`. An
  absent knob in a current-version config is still a silent level;
  only an unproven transport fails closed.
- Provenance enum extended to
  `source=flag|profile|global|default-interactive|default-headless`
  (item 6); precedence, header, gap statement, and item 1 updated for
  the split default; item 3 cross-references the item-5 refusal when
  lock engagement is unestablished; open question 5 gains the
  headless/legacy negative-test rows; open question 7 now asks only
  for the version token, fragment member names, and the full
  non-interactive marker enumeration (the fail-closed rule is
  specified, not deferred); compatibility touch-list and the security
  statement updated accordingly.
- Kept from revision 1: config surfaces (`defaults.json` v2 member,
  §12.1 `permissions.<profile>` knob via fragment), precedence order,
  lockable force-`native`, per-environment mapping table (byte-unchanged),
  refusal table, tracked-mode outcome.

`CHANGELOG.md` and `UNRESOLVED_QUESTIONS.md` bullets updated to the
corrected rules. `git diff --stat` at revision 2: 3 files,
+245/−55; no commits made.

### Revision 2 validation

Same gate as revision 1, project-local `.venv` (created, used,
removed; the Change Request snapshot contains only the three files
above). Each command run directly as a standalone process (shell:
bash), real exit codes:

- `python3 tools/validate.py` → `validated 60 schemas and 1047 vector files`, exit 0
- `python3 -B -m unittest discover -s tools -p 'test_*.py'` → `Ran 227 tests ... OK` (198.9s), exit 0
- `go test ./tools/...` → `ok .../tools/generate-vectors 0.643s`, exit 0
- `git diff --check` → exit 0

These three commands are exactly the `make validate` recipe (read
from the Makefile), so the gate is green by composition; `make`
itself was not re-invoked separately — it would only re-run the same
~200s suite.

No behavior-bearing code changed (decision prose + list/changelog
entries only), so no new tests were added; the repo suite above is the
verification gate. It does not prove the new permission invariants —
they are draft prose with no runtime — and no mutation or coverage
claim is made, matching the reviewer's bound.

### Acceptance-criteria mapping (revision 2)

- decisions/0018 updated (still proposed) — done, this change.
- `make validate` green — done, exits 0 quoted above.
- PR landed after independent review; issue #55 updated —
  orchestrator-owned; the developer handoff publishes Change Request
  revision 2 for independent review. Not performed in this role.

### Findings (revision 2)

- The revision-1 version-skew finding is now closed normatively in the
  draft itself (item 5 fail-closed rule + item 3 cross-reference +
  open question 7 reduced to token/enumeration details). No logbook
  entry beyond this artifact: the correction is captured in the draft.
