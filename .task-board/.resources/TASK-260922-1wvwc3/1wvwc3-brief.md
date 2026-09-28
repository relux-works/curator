# TASK-260922-1wvwc3 — F-M1b: versioned capability table, drift refusal, release (skill-agents-management)

Control root: /Users/administrator/Developer/ReluxWorks/skill-agents-management (shared with the
operator's own session — never write outside your assigned Story worktree
`.temp/STORY-260922-h3epwn/worktree`). Read `campaign-producer-rules.md` first. Board validation
command runs once at handoff.

## Depends on
F-M1a (TASK-260922-1s0zja), landed on this Story before you: `LaunchRequest.PermissionMode`
(`native`/`yolo`), the per-plugin bypass-flag constants (claude `--dangerously-skip-permissions`,
codex `--dangerously-bypass-approvals-and-sandbox`, pi/pinative refuse with
`ErrPermissionModeUnsupported`), the module-wide two-site ownership guard, and the interactive
goldens. Read its `TASK-260922-1s0zja_results.md` and the accepted verdict first — do not
re-implement or re-shape any of it.

## Source of truth
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/decisions/0018-curator-run-permission-interface.md`
adoption choices 3 and 6 (and the Compatibility section):
- choice 6: "Each mapping is re-verified per tool release; the capability table keys (environment,
  tool release) to a grammar version and is owned by agents-management as part of the
  `LaunchRequest` permission-mode member for `LaunchModeInteractive`, with goldens per tool release.
  On version drift the `yolo` mapping fails closed first (refuse `yolo` for unverified releases;
  `native` still forwards verbatim with no claims)."
- choice 3: "The provider grammar is closed per pinned tool release; unknown future native policy
  forms (new Codex `-c` keys, new Claude modes) are refused (`usage`, exit 2), never resolved into a
  policy claim. The grammar distinguishes prompt text from flags (agents-management owns the parsing
  rule as part of its argv grammar)."

## Deliverable
1. **Capability table**: a declared, versioned table keyed by (environment/system id, tool release)
   → permission-grammar version, with the verified releases F-M1a recorded (claude 2.1.261, codex
   0.153.2, pi 0.84.2 = unsupported). The table is data, spelled once, reachable from the plugin
   that owns each mapping; the grammar version token must be quotable by the launcher (F-L1 cites
   it), so state it in the CHANGELOG and README.
2. **Release detection + drift rule**: the launch path establishes the running tool release (the
   repository already resolves binaries per plugin — reuse that path; do not invent a network
   call), looks it up, and on an UNVERIFIED or newer release refuses `yolo` with a named diagnostic
   while `native` still forwards verbatim and makes no policy claim. Detection failure is itself a
   refusal for `yolo` and a pass for `native` (fail closed, never a claim).
3. **Unknown native policy forms** (choice 3): a new codex `-c` key or a new claude mode in the
   caller's own arguments is refused as `usage` (exit 2 semantics at this layer = a named error the
   caller maps to exit 2), never resolved into a policy claim. Rows for at least one new `-c` key
   and one new claude mode.
4. **Parsing rule**: prompt text that looks like a flag is never parsed as one (this is the rule
   0018 says agents-management owns). Rows proving both directions.
5. **Goldens per tool release**: positive (verified release → exact flag) and negative (unverified
   release → refusal; detection failure → refusal for yolo, pass for native) for claude_code and
   codex_cli; pi stays unsupported.
6. One narrowing mutant per refusal bound, executed and killed, in a table.
7. README + CHANGELOG, then **cut the release** the launcher pins: next patch of `v0.5.x` — a signed
   annotated tag is the ORCHESTRATOR's step, so do NOT tag; state the exact version you expect and
   make sure `CHANGELOG.md` is release-ready.

## Boundaries
No launcher, curator, ax or task-board change. No real provider launches in tests (fake binaries /
argv capture only). Hand off with `task-board handoff TASK-260922-1wvwc3 --role developer` after
attaching `TASK-260922-1wvwc3_results.md` (table, drift rows, mutants, narrow test exit codes,
proposed release version).
