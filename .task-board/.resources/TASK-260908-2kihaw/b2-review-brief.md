# Review brief — TASK-260908-2kihaw b2-skill-packages-delivery (independent exact-head review of three PRs)

Candidates (private repos, org relux-works; bot SSH/gh access configured on this host). Review EXACTLY these heads; a later push voids this review:
- skill-pdf PR #1, branch b2/manifest, head `6d2392a861cdf389c9aa0245c46e8207975a2bcd` (base main 9e3e72a1)
- skill-creator PR #1, branch b2/manifest, head `ea8fd665486233ccaccce2d8508ea8caa9378bcd` (base main 0a49a505)
- skill-agents-attachments PR #1, branch b2/manifest, head `240f0292a6484a743ede98fb9af0097f4a875480` (base main a8156f1d)
Confirm with `GODEBUG=netdns=go gh pr view N -R relux-works/<repo> --json headRefOid`. Clone read-only under your worktree's `.temp/`.

Verify against `TASK-260908-2kihaw_producer-evidence.md`, `TASK-260908-2kihaw_validation.log`, `TASK-260908-2kihaw_umbrella-range-contract.py` and the B2 brief:
1. Each repo's `agent-skill.json` (schema 8) validates with the installed manager: `curator skill check <dir>` from the checkout; commands and capabilities match the evidence (pdf: 1 script command; skill-creator: 3 script commands; agents-attachments: 1 go-v1 build command with nested module at cmd/agents-attachments, no `modules` replacements).
2. Byte fidelity of imported bodies versus relux-agents-infra @ `dee5403` (`git -C /Users/administrator/Developer/ReluxWorks/relux-agents-infra show dee5403:<path>`): recompute the sha256 of SKILL.md, scripts and assets yourself; executable bits preserved.
3. The manifest gates (`tests/manifest_gate.sh`, `make check`/`make test`) pass in each checkout and bite: run at least two narrowing mutants per repo yourself.
4. Go build/test of agents-attachments (`go build ./... && go test ./...` inside cmd/agents-attachments), shellcheck cleanliness where the evidence claims it.
5. Umbrella range contract: `requires.skills` entries documented in the READMEs (`^0.1` admitting v0.1.0) are consistent with the manifests' versions; the range contract script's 24/24 claim reproduces.
6. Signatures: every PR commit signed by Ivan Oparin <ivan@relux.works> (key SHA256:Ng99…, public key /Users/administrator/.ssh/ivan-relux.pub), `[skip ci]` present; no unrelated files.

Record one verdict as a task-scoped outcome resource `TASK-260908-2kihaw_review-verdict.md` (ACCEPT / CHANGES_REQUESTED, per repo, with exact heads, commands, exit codes). Post the same verdict as a comment on each PR (`gh pr comment`). On ACCEPT set `set_status(TASK-260908-2kihaw, status=done, commit_ack=scope_committed)`: landing (fast-forward of the exact heads to main) and the three signed v0.1.0 tags are the orchestrator's job, not yours. On CHANGES_REQUESTED route to `to-dev` and list the required changes.
