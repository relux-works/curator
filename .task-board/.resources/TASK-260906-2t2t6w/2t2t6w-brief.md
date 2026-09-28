# TASK-260906-2t2t6w — launcher SPEC minors from the 0.2.1 review (curator-agent-launcher)

Control root: /Users/administrator/Developer/ReluxWorks/curator/curator-agent-launcher; work only in your
assigned Story worktree `.temp/STORY-260905-3l0fav/worktree`. Read `campaign-producer-rules.md` first. The
board's landing gate is `sh scripts/remote-gate.sh` (hosted CI) — the runtime runs it once at handoff; run
`make check` (build, fmt-check, vet, test, race) yourself with real exit codes and cite them.

The SPEC is now `0.4.0-draft` (SPEC.md line 3, `specVersion` in cmd/curator-run/main.go:30, README). Apply the
four non-blocking minors recorded in
`/Users/administrator/Developer/ReluxWorks/curator/curator/.task-board/.resources/TASK-260905-2czqqy/TASK-260905-2czqqy_review-findings-launcher-0.2.1.md`
(section "Minor findings", items 1–4) as one three-file version bump to `0.4.1-draft`:

1. §4.1 resolve failure modes / §6 `resolve_invocation_failed` gloss: today §4.1 already says "Curator's
   stderr is forwarded to the operator verbatim, on failure and on success alike" — verify that this
   satisfies minor 1 (Curator's own diagnostic code and message reach the operator beneath the launcher's code
   line for the widened `--repair` failure surface: environment_marker_invalid,
   environment_surface_unmanaged_conflict, environment_backup_exists, environment_seed_unreadable). If §6's
   gloss still leads with "curator not startable" for that class, add the one pass-through clause the
   finding asks for (exactly as `ax_handoff_failed` does) and fix the gloss. Do not add mappings for those
   codes (the finding does not ask for them).
2. §4.6 three pre-launch checks (binary check, §4.5 codex layer stat, §5.1 file-kind probe): state the order
   ("run in this order, and the first failure is the one reported"), binary check first; make the
   implementation match if it does not already (find the check sequence in internal/execution and add a
   test that provokes all three failures at once and asserts the single reported code).
3. §9 residuals: add the item that whether `claude_code` (`--mcp-config <missing> --strict-mcp-config`) and
   `opencode` (`OPENCODE_CONFIG`) also proceed silently when the MCP configuration file is absent is
   unverified at the pinned releases, and that if either does, the §4.5 stat rule generalizes to it.
4. `cmd/curator-run/main_test.go` `TestSpecVersionPinned`: read SPEC.md and README.md and assert
   `specVersion` appears in both (the comment claims the three are one fact — make it true); prove with two
   mutants (SPEC.md and README.md set back to 0.4.0-draft) that the test now fails, then restore.

CHANGELOG entry (unreleased) naming the four minors; SPEC §(version history) row for 0.4.1-draft. No other
SPEC changes (the 0018 permission-interface revision is a separate leaf — do not touch §3/§4.3/§4.5 mode
text beyond minor 2). Attach `TASK-260906-2t2t6w_results.md` (per-minor before/after, mutant table for
minor 4, `make check` exit code) and hand off with `task-board handoff TASK-260906-2t2t6w --role developer`.
