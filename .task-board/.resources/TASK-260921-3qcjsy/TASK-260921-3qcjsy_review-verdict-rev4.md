# TASK-260921-3qcjsy review verdict — revision 4 (CR-TASK-260921-3qcjsy-4)

Reviewer run RUN-260921-6d7016 (claude-opus-5/max), 2026-09-21.
Reviewed delta: base `8e65374c5dcba2ae3ac9e0f869a2cf51361db52c` → candidate
tree `72b7b3eaba406e1c9b4011ee3ea6b2da514a328a` (160 paths). The Story
worktree's working tree hashes to exactly that tree OID (temp-index
`git write-tree`; HEAD = 8e65374c, no commit past the checkpoint), the
rev4 patch resource (sha256 `277ddc8a…`) is byte-identical to the rev3
patch resource and reproduces the same tree OID when applied to
8e65374c in a disposable clone, and every rerun below ran on that clone.

## Verdict: ACCEPT → `accept_cr(TASK-260921-3qcjsy, revision=4)`

Revision 4 is exactly the revision-1 mechanics (accepted then) plus the
F1–F3/N1–N4 prose rework I asked for, re-applied unchanged on the fresh
trunk. The gate is green on the exact revision-4 tree, both on the
configured gate and on my own reruns. One non-blocking nit (N5) is
recorded for the next revision that touches 0017.

## rev4 − rev1 is the requested text delta only

Applied the rev1 patch to 802caee (tree `85565cd0…`, as reviewed) and the
rev4 patch to 8e65374c (tree `72b7b3ea…`) in the same clone and diffed the
two candidate trees, excluding the trunk move's own paths
(`conformance/draft-sources-v1/**`, BUG-260921-3cgij4, path-disjoint):

```
decisions/0017-environment-credential-modes.md     | 23 ++++++-----
decisions/0018-curator-run-permission-interface.md | 24 +++++++-----
profiles/manager.md                                |  8 ++--
protocol/environments.md                           | 44 ++++++++++++----------
4 files changed, 56 insertions(+), 43 deletions(-)
```

No schema, vector, generator, gate, CHANGELOG, COMPATIBILITY, index, or
cli change beyond rev1's accepted mechanics. Every hunk maps to one of
F1–F3/N1–N4 (checked hunk by hunk below).

## F1 — schema-1 marker prose vs the frozen v1 marker schema: FIXED

- `protocol/environments.md:1893-1894` (§8.2 passthrough bullet) is
  byte-identical to the base text again ("with their section 7.4
  strategy, and the recorded provisioning `seeds`"); §8.2 as a whole has
  no hunk against 8e65374c.
- `protocol/environments.md:1486-1500` (§7.4 "Credential record") now
  opens "From the marker revision Decision 0017 choice 5 defines
  (follow-up), every passthrough record carries …" and states "A schema-1
  marker records `path` and `strategy` only and is never rewritten to add
  the record." The linkless-strategy "record without a path" sentence sits
  inside that scope, so nothing claims a schema-1 entry without `path`
  (`agent-environment-marker-v1.schema.json` requires `path` + `strategy`,
  closed items — untouched here).
- `decisions/0017-environment-credential-modes.md:155` (choice 5) carries
  the same schema-1 sentence; `:161-170` (Compatibility) now lists "the
  credential-record content the follow-up marker revision defines" and
  says "§8.2 and §12.1/§12.2 unchanged — a schema-1 marker records `path`
  and `strategy` only". results.md row 5 and the amendments list say the
  same.

## F2 — `isolated` for `codex_cli` under `auto` storage: FIXED, fail-closed

- `protocol/environments.md:1383` (passthrough table row): "under
  `keyring` or `auto` storage `isolated` is
  `environment_isolated_unsupported`".
- `:1440-1448` (rule paragraph): names `keyring` or `auto`, gives the
  reason (under `auto` the file-link is inert on a keyring host, so an
  `isolated` home would authenticate through the operator-global keyring),
  keeps "never a silently shared home", and defers any `auto` admission to
  a future revision "only where the manager proves the effective store is
  `file`, probe-gated".
- `:1466-1467`: "`isolated` remains available for `codex_cli` under `file`
  storage only, for `pi`, and for `claude_code` on Linux."
- `:1604` (isolation matrix row): "`isolated` available under `file`
  storage only (`isolated` refused under `keyring` or `auto` storage)".
- `profiles/manager.md:2741-2745` (paragraph, with the rationale and a
  §7.4 pointer) and `:2766` (table row) name `keyring` or `auto`.
- `decisions/0017-environment-credential-modes.md:154` (choice 4) and
  results.md row 4 state the same rule and the probe-gated refinement.
- The `shared` + `auto` file-link in the strategy column (`:1383`) is
  unchanged and correct: under `shared` the keyring is the shared store by
  definition, the inert link is harmless. No conformance vector or schema
  case names `cli_auth_credentials_store`, `keyring`,
  `environment_isolated_unsupported`, `environment_credential_conflict`
  or `environment_credential_unsupported`, so no vector contradicts the
  new rule (checked over `conformance/v1/**`).
- Privilege direction: the change only narrows admission (one more
  configuration becomes a refusal); nothing widens.

## F3 — proposal-era sentences: FIXED

`git grep -n "not an adoption\|adopting revision amends\|stays open question" -- decisions/0017-environment-credential-modes.md decisions/0018-curator-run-permission-interface.md`
→ no matches (exit 1). Each rewording checked:

- `0017:54-56` / `0018:60-63`: "recorded below as adopted: options … adopted
  as recommended / as amended, open questions 1–7 resolved to the recorded
  choices."
- `0018:186-187`: "(the fragment schema is untouched by this revision; the
  member names are choice 7, follow-up F-S2)" — true (`launch-env-fragment-v1`
  has no hunk); `0018:190` "this decision creates no launcher section".
- `0018:242-243`: "(the closed marker set choice 7 fixes — {`CI`,
  `GITHUB_ACTIONS`})" — item 5 now points at choice 7, which fixes the set;
  environments §10.1 (`:2919-2920`) carries the same closed set.
- `0018:275-277`: "The minimum fragment/Curator version token is choice 7
  (follow-up); the fail-closed rule itself is specified here, not deferred."
- Remaining "adopting revision" at `0018:262-263` ("a fragment that predates
  the adopting revision — one that cannot carry the profile level or the
  item-3 lock engagement") is coherent as adopted text: the em-dash clause
  is the operative condition and §10.2 (`:2999-3007`) states the same rule
  ("until they are, every fragment predates the transport").

## N1–N4: done as requested

- N1 `protocol/environments.md:1416-1420`: "a non-empty one fails the
  repair with `environment_credential_conflict` instead of destroying its
  contents." Consistent with the vector
  `passthrough-directory-at-entry-detached` (status-side diagnostic stays
  `environment_passthrough_detached`, `repair_relinks: true` for the empty
  directory the prose relinks); the §7.7/§10.4/manager rows still describe
  the regular-file/unexpected-target shape only — acceptable, the
  non-empty-directory sentence is the one normative site and names the
  same never-remove diagnostic.
- N2 results.md F-S1 now includes "the first spec vector rows pinning
  `environment_credential_conflict` / `environment_credential_unsupported`
  (no spec vector pins them today)" with an AC; re-verified: nothing under
  `conformance/`, `release/`, `schemas/` names either diagnostic.
- N3 results.md 0018 row 7 carries the operator-confirmation item (thin
  {`CI`, `GITHUB_ACTIONS`} set; TTY prong primary; silence does not fail
  open by default).
- N4 `0018:333-337`: "the versioned provider-capability encoding (choice 6)
  is owned by the launcher SPEC revision (launcher follow-up F-L1)";
  results.md F-L1 mirrors it; choice 6 ("fixed by the implementing
  revision") agrees.

## Verified (own reruns, disposable clone at tree 72b7b3ea…)

Full detail in `TASK-260921-3qcjsy_review-rev4-gate.log`.

- `python3 tools/validate.py` → `validated 62 schemas and 1124 vector
  files`, exit 0.
- `python3 -B -m unittest discover -s tools -p 'test_*.py'` split into
  bounded chunks (single-call limit): 36 + 32 + 5 + 5 (other files) + 193 +
  258 + 32 + 18 (`test_validate` by class) = 579 tests, 0 failures, 0
  errors; clone `git status` clean afterwards.
- `go test -count=1 ./tools/...` → ok (0.728 s) on the second attempt; the
  first attempt was `signal: killed` at 74.8 s — the host exec-stall window
  (a freshly built hello-world binary stalled the same way at 0 CPU;
  ad-hoc codesign / xattr changes make no difference; `/usr/bin/grep`
  stalled once too). Environmental, not the change: identical symptom to
  the rev2/rev3 logs, and the configured gate's rev4 log shows the same
  step green in 0.951 s.
- `go run ./tools/generate-vectors -root .` then `git diff --exit-code --
  conformance/v1 release/1.0.0-rc.{5..9}.json` → exit 0: cases, index,
  manifest, `vectors/manager-config-v2.json` and the rc.9 pin are
  byte-identical generator output; rc.5–rc.8 untouched. `gofmt -l tools/`
  empty; `go vet ./tools/...` clean.
- Configured gate on the same candidate
  (`TASK-260921-3qcjsy_change-request_rev4-validation.log`,
  `spec-gate.sh` = the three `make validate` steps with a retry only on
  `signal: killed`): 62/1124, 579 OK, go test ok, exit 0.
- Mechanics unchanged since rev1 (schemas: `manager-config-v2`
  `environments.permissions` identifier-keyed `native|yolo` map;
  `system-config-v2` `permissions` `native`-only plus `locked` entry; no
  frozen v1 schema touched; rc.9 repin allowed by the repo's live-pin
  convention; narrowing negatives in `tools/test_validate.py` and the Go
  generator tests) — re-verified only by the byte-identical rev4 − rev1
  diff and the green reruns, not re-reviewed.
- Decision tables vs results.md: 0017 rows 1–7 and 0018 rows 1–7 match in
  substance (results.md condenses; row 4/5 and row 7 carry the rework
  text). Follow-up leaves still map 1:1 to the deferrals (F-C1/F-C2/F-C3,
  F-L1, F-S1/F-S2/F-S3, F-A1, E1/E2) with one-line ACs.
- Index/links: UNRESOLVED_QUESTIONS.md "Adopted decisions" lists both;
  "Filed proposals" no longer does; `validate_local_links` (in validate.py,
  green) covers every `*.md`.

## Non-blocking (record for the next revision touching 0017)

- N5 `decisions/0017-environment-credential-modes.md:14`: the Status
  adoption note still lists "§8.2" among the sections amended in this
  revision, while the Compatibility section (`:168-170`) correctly says
  §8.2 is unchanged and §8.2 has no hunk against the base. A one-token
  removal; it fell out of the F1 fix (the list was accurate at rev1). Fold
  it into F-S1, which must edit 0017 choice 5 / Compatibility anyway.
- The follow-up labels F-S2 / F-L1 cited inside the decision documents
  resolve only through results.md today; replace them with the leaf IDs
  once the orchestrator creates the leaves (cosmetic).

## Checklist disposition

- Tests green: yes (own reruns above; configured gate green on the same tree).
- Implementation matches AC: yes — Status adopted with dated notes, every
  option and open question resolved to a recorded choice, amended
  normative text consistent (F1–F3 resolved), `make validate` green,
  follow-up leaves named per repository with one-line ACs.
- Solution fits project architecture: yes — prose stays behind the frozen
  marker schema, the codex rule is fail-closed, generator-owned corpus.
- Logbook: per the binding review note the control-root LOGBOOK.md is
  operator-owned; findings are recorded only in this verdict resource.
