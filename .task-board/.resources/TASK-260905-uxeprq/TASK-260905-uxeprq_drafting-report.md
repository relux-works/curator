# Drafting report: Decision 0010 erratum (TASK-260905-uxeprq)

## Commit

- Worktree: `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-0010-erratum`
- Branch: `draft/decision-0010-erratum`, base `b4f29cd` (curator-spec main)
- Commit: `9198c64b5cdc42fe0ec2b6f2fca8955b3fe5796e`
- `git log --show-signature -1`:
  `Good "git" signature with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`
  (a following "No principal matched" line comes from a stale
  `gpg.ssh.allowedSignersFile` pointing at a deleted `/private/tmp/...` path in
  the local git config; the signature itself verifies.)
- Files changed: `decisions/0010-agent-environment-profiles.md` only
  (77 insertions, 9 deletions). Not pushed, no tag, no PR.

## Item -> line table (committed file, `9198c64`)

| Item | Where | Line(s) |
|---|---|---|
| Status sentence naming the erratum | `## Status` | 9-11 |
| `## Erratum (2026-09-05)` section | after Status, before Context | 13-76 |
| Item 1 entry (pi flags) | erratum list | 21-41 |
| Item 2 entry (sequencing) | erratum list | 43-55 |
| Item 3 entry (credentials) | erratum list | 57-76 |
| Item 1 annotation | Decision 2, "System prompt" | 255 |
| Item 2 annotation | Decision 1, `path` bullet | 156 |
| Item 2 annotation | Decision 11 phasing row, first cell | 721 |
| Item 3 annotation (Keychain / opencode sentence) | Decision 7 | 616 |
| Item 3 annotation (`isolated` sentence) | Decision 7 | 621 |

Context evidence list (lines 111-115 after the insert, formerly ~45) only names
binary versions and does not assert the pi flag verification; left unchanged
per the brief. Decision 6 loader-source claim (line 581) stands. The
credentials paragraph at 616-621 was re-wrapped to keep the house 80-column
style after inserting the marker; wording is unchanged.

## Binary re-verification (2026-09-05, this machine)

`pi --version` -> `0.84.2`. `pi --help` lines:

```
  --system-prompt <text>         System prompt (default: coding assistant prompt)
  --append-system-prompt <text>  Append text or file contents to the system prompt (can be used multiple times)
```

No `-file` spellings in the help. Direct probes (print mode):

```
$ pi --system-prompt-file /dev/null -p "say ok"
Error: Unknown option: --system-prompt-file
$ pi --append-system-prompt-file /dev/null -p "say ok"
Error: Unknown option: --append-system-prompt-file
```

(`pi --system-prompt-file ... --help` and `--version` short-circuit before
option validation and do not reject; the print-mode probe is the evidence.)

`claude --version` -> `2.1.261 (Claude Code)`. `claude --help` lists
`--system-prompt <prompt>` and `--append-system-prompt <prompt>` as option
rows and names the file variants in the `--bare` description:

```
                                        via: --system-prompt[-file],
                                        --append-system-prompt[-file], --add-dir
```

Direct probes confirm both file flags are parsed:

```
$ claude -p --system-prompt-file /nonexistent/x.md "hi"
Error: System prompt file not found: /nonexistent/x.md
$ claude -p --append-system-prompt-file /nonexistent/x.md "hi"
Error: Append system prompt file not found: /nonexistent/x.md
```

## Sources read

- `pre-implementation-review-v3.md` (STORY-260901-zddtn8): M4, M5, N13.
- `TASK-260902-2142et_lens-operator-ux.md`: F22.
- curator-spec `b4f29cd`: `protocol/environments.md` §1 (`path` is a
  revision-1 kind), Decision 0012 Context.

## Not done

- No LOGBOOK write (brief forbids writing into the control root).
- The story worktree `task-board/story/STORY-260905-1ytok2` is untouched; the
  commit lives on `draft/decision-0010-erratum` per the brief.
