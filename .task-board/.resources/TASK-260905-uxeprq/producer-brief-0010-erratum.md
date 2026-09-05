# Producer brief: Decision 0010 erratum (review N13, M5, M4)

## Where and what

- Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-0010-erratum`,
  branch `draft/decision-0010-erratum`, base `b4f29cd` (= curator-spec main).
- Edit exactly one file: `decisions/0010-agent-environment-profiles.md`. Nothing
  else. The story worktree task-board provisions for this run stays untouched.
- Shape: a recorded amendment, never a silent edit. Add a `## Erratum (2026-09-05)`
  section immediately after `## Status` (before `## Context`) listing each
  corrected claim: the original sentence quoted verbatim, why it is wrong, the
  evidence (binary version / review item), and the corrected statement. Leave the
  original passages in place and annotate each with a bracketed marker of the form
  `[Erratum 2026-09-05, item N]` at the end of the affected sentence — mirror
  how a published RFC erratum reads. Update the Status paragraph with one
  sentence saying the erratum exists. Do not restructure or rewrite anything else.
- Deliverable: one signed commit (`git commit -S`; paste the
  `git log --show-signature -1` line into your report). Do not push, tag, open a
  PR, or mark the task done. Attach `TASK-260905-uxeprq_drafting-report.md` as an
  outcome resource with the commit hash and an item→line-number table. Then
  `task-board handoff TASK-260905-uxeprq --role developer`.

## Sources (read-only)

- Board resource `pre-implementation-review-v3.md` on STORY-260901-zddtn8
  (`~/Developer/ReluxWorks/curator/.task-board/.resources/STORY-260901-zddtn8/`):
  items M4 (credentials), M5 (pi flags), N13; and lens B finding F22 in
  `~/Developer/ReluxWorks/curator/.task-board/.resources/TASK-260902-2142et/TASK-260902-2142et_lens-operator-ux.md`
  (verified-on-binary evidence for pi 0.84.2).
- Installed binaries on this machine: re-verify `pi --help` (record the version
  from `pi --version`) shows `--system-prompt <text>` and
  `--append-system-prompt <text>` and no `-file` spellings; re-verify
  `claude --help` (record version) lists `--system-prompt-file` and
  `--append-system-prompt-file`. Paste the relevant help lines into the report.
- curator-spec main `b4f29cd`: `protocol/environments.md` §1 (the `path` kind is a
  revision-1 source kind, promoted by PR #34 `f8d7e7ab`), Decision 0012 Context
  (M1 Option A; the review's MUST items binding).

## The three items

1. **pi flag verification (Decision 2, "System prompt" paragraph, lines ~180–188).**
   Original: "`pi` takes the same flags and additionally reads agent-dir
   `APPEND_SYSTEM.md` (append) and `SYSTEM.md` (full replacement), both applied
   unconditionally when present (verified in 0.84.2)". Wrong part: pi 0.84.2 has
   no `--system-prompt-file`/`--append-system-prompt-file`; its flags are
   `--system-prompt <text>` and `--append-system-prompt <text>` (the latter
   polymorphic: text or file contents, so a dead path is sent as literal prompt
   text). The agent-dir file claim is separately verified in the loader source
   (Decision 6 says so) and stands. Corrected statement: pi exposes no
   file-taking replace flag; the only replace path is `SYSTEM.md`; the append
   channel is `--append-system-prompt` with a path the launcher verifies readable
   immediately before exec — the normative row is rewritten by environments.md
   revision 1.1 (review M5), this erratum only corrects the claim.
   Also correct the Context line ~45 evidence list if it asserts the flag
   verification for pi (read it; if it only names versions, leave it).
2. **Superseded sequencing sentence (Decision 1, `path` bullet, lines ~84–90).**
   Original: "it is delivered with the onboarding import story, not in
   revision 1." Wrong: the `path` source kind, onboarding import, and install
   ref selection were promoted into revision 1 by PR #34 (`f8d7e7ab`) and
   environments.md §1 lists `path` as a revision-1 kind. Also the Decision 11
   phasing row "Onboarding import (`path` source kind, lossless/lossy, skill
   migration) — ✓ own story between rev 1 and 2" — mark it with the same
   erratum item and state the corrected row (revision 1). Do not edit the
   table cells themselves; annotate the row's first cell.
3. **Credentials claim (Decision 7, lines ~542–555).** Original: "macOS Keychain
   entries are ambient and need nothing" and "A profile×environment pair MAY be
   configured `isolated` — no passthrough, the tool authenticates fresh inside
   the profile home — which is the supported shape for genuinely separate
   accounts", and for opencode "keeps auth in the XDG data directory, which the
   config swap never touches". Wrong per review M4 (verified): the Claude
   credential is the Keychain item `Claude Code-credentials` keyed by the macOS
   account, so a fresh login inside an `isolated` home rewrites the same item
   and clobbers every other home's credential; a fresh `CLAUDE_CONFIG_DIR`
   reports "Not logged in" because login state lives in `.claude.json`; opencode
   `isolated` is a no-op for auth. Corrected statement: `isolated` is
   unsupported in revision 1 for `claude_code` on macOS and for `opencode`
   (`environment_isolated_unsupported`), lifted only on positive evidence; the
   per-adapter passthrough *strategy* and the provisioning-seed class are
   specified by environments.md revision 1.1 (review M4); an
   `oauth.claude.profile.<64-hex>` Keychain account exists and is under
   investigation in the verification sprint — cite it as unverified.

## Constraints

- Quote originals exactly (copy from the file, do not retype). Verify every line
  number you cite in the report against the committed file.
- English; house style of the surrounding document. No other content changes.
- Never write LOGBOOK.md or anything into the control root.
