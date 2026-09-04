# Review brief: launcher SPEC 0.1.1-draft (pi file-kind channels)

Subject: `~/Developer/ReluxWorks/curator-agent-launcher`, branch `draft/spec-pi-file-channels`, head `0029b36` (base = origin/main `dae0c35`). Producer notes on TASK-260901-25zqel. Read-only; do not push.

Verify against protocol/environments.md §5.5/§7.3 and Decision 0010 D6:
1. File-kind semantics correct: launcher never places/removes/edits the files; Curator materialization owns them under the machine setting; pi applies them unconditionally when present.
2. Warning rule: active file-kind channel in the target managed home triggers the full customized-system-prompt warning set even without the --system-prompt flag; both-kinds case enumerates every active channel; flag opt-in engages flag-class only, orthogonal to files.
3. Diagnostic for fragment-named file-kind channel absent/unreadable at launch: defined, coded, consistent with SPEC's diagnostics section.
4. Revision history present, 0.1.1-draft; delta confined to SPEC.md; commit signed; no code changes.
5. No contradiction introduced with the rest of SPEC (§ composition, ax handoff).

Verdict: `review-findings-launcher-pi-1.md` on the task; blocking/major -> development; else ACCEPT explicit + accept_cr on the current CR revision, leave to-review.
