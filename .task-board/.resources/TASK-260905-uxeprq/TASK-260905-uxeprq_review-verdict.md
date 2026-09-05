# Review findings, cycle 1: Decision 0010 erratum (TASK-260905-uxeprq)

Verdict: **ACCEPT** (CR-TASK-260905-uxeprq-1 revision 1).

Subject: branch `draft/decision-0010-erratum`, head `9198c64b5cdc42fe0ec2b6f2fca8955b3fe5796e`,
base `b4f29cd`, worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-0010-erratum`
(clean, HEAD = 9198c64). Reviewed 2026-09-05.

## Why the empty repository delta is the right outcome

The Change Request's candidate tree equals its base because the producer brief
directs the edit into a separate worktree on `draft/decision-0010-erratum` and
states the story worktree "stays untouched". The deliverable is the signed commit
`9198c64` on that branch plus the drafting report. Both exist and were verified
below, so no change to the story branch was expected from this leaf. Integration
of `draft/decision-0010-erratum` into trunk is a separate step outside this leaf.

## Dimension 1: amendment, not silent edit — PASS

- `git diff --stat b4f29cd..9198c64`: one file, `decisions/0010-agent-environment-profiles.md`,
  77 insertions / 9 deletions. No unrelated hunks.
- `## Erratum (2026-09-05)` sits at line 13, after `## Status` (3) and before `## Context` (78).
- Status paragraph carries one sentence naming the erratum (lines 9-11).
- Each item quotes the original verbatim; diffed the quotes against the base file text
  (item 1 line 252-255, item 2 line 155-156 and table row 721, item 3 lines 611-621):
  identical modulo line wrapping.
- Original passages remain, markers at 156 (item 2), 255 (item 1), 616 and 621 (item 3),
  721 (item 2, first cell of the phasing row). Table cells unchanged; `git diff -w --word-diff`
  of the credentials hunk shows only the two inserted markers, so the re-wrap changed no wording.
- Context evidence list (lines 111-115) names versions only; correctly left alone.

## Dimension 2: facts — PASS (re-verified on this machine, 2026-09-05)

```
$ pi --version
0.84.2
$ pi --help | grep system-prompt
  --system-prompt <text>         System prompt (default: coding assistant prompt)
  --append-system-prompt <text>  Append text or file contents to the system prompt (can be used multiple times)
$ pi --system-prompt-file /dev/null -p "say ok"
Error: Unknown option: --system-prompt-file
$ pi --append-system-prompt-file /dev/null -p "say ok"
Error: Unknown option: --append-system-prompt-file
$ claude --version
2.1.261 (Claude Code)
$ claude --help | grep -- '-file'
                                        via: --system-prompt[-file],
                                        --append-system-prompt[-file], --add-dir
$ claude -p --system-prompt-file /nonexistent/x.md "hi"
Error: System prompt file not found: /nonexistent/x.md
$ claude -p --append-system-prompt-file /nonexistent/x.md "hi"
Error: Append system prompt file not found: /nonexistent/x.md
```

The only `-file` string in `pi --help` is `--no-context-files`; no file-taking prompt flag.

- `path` promotion: `git show b4f29cd:protocol/environments.md` §1 lists `path` as a source
  kind with its own diagnostics table rows; `git log --oneline f8d7e7ab -1` = "Promote the path
  source kind and specify the onboarding import and install ref selection". Claim holds.
- Review M5 (pre-implementation-review-v3.md lines 140-160): erratum item 1 matches the
  verified facts and the resolution wording (append via `--append-system-prompt` with a
  launcher-verified readable path; `SYSTEM.md` the only replace path). Polymorphic-append
  claim matches M5 and lens B F22 (lines 597-609 of the lens resource).
- Review M4 (lines 111-137): erratum item 3 reproduces the verified facts (`Claude Code-credentials`
  Keychain item keyed by macOS account, fresh `CLAUDE_CONFIG_DIR` "Not logged in" because
  state lives in `.claude.json`, opencode `isolated` no-op with auth in `XDG_DATA_HOME`)
  and the resolution (`environment_isolated_unsupported`, lifted only on positive evidence;
  passthrough strategy and provisioning-seed class owned by environments 1.1).
  `oauth.claude.profile.<64-hex>` is labeled unverified with "nothing here relies on it".
- N13 (line 300) maps to O-M9 and B-F22; erratum item 2 cites "N13 (O-M9)" correctly.

## Dimension 3: no overreach — PASS

Each item ends by naming `protocol/environments.md` revision 1.1 as the owner of the
normative rewrite and states the erratum corrects only the claim. No normative text
was introduced into 0010; no other passage was touched.

## Dimension 4: signed commit — PASS

`git log --show-signature -1`: `Good "git" signature with ECDSA key
SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`, author `Ivan Oparin <oparin@me.com>`.
The trailing "No principal matched" is a stale `gpg.ssh.allowedSignersFile` pointing at a
deleted `/private/tmp/curator-spec-rc8-verify.*` path in the local git config; it is a
local verifier-config defect, not a signature defect. Exactly one commit above base.

## Findings

| # | Severity | Location | Issue | Fix |
|---|---|---|---|---|
| 1 | nit | drafting report, item->line table | Erratum entry ranges are off by one: entries start at lines 22, 44, 58 (report says 21, 43, 57). All annotation lines (156, 255, 616, 621, 721) and the section range (13-76) are exact. | Optional: correct the three start lines in the report. Does not affect the commit. |

No blocking or major findings.

## Not done by this review

- No LOGBOOK write, no edits, no commits, no pushes.
- Did not verify the `oauth.claude.profile.*` Keychain scheme itself; the erratum
  correctly labels it unverified, which is all this review required.
