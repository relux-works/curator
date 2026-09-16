# TASK-260915-4f44cs — landing evidence

- PR: https://github.com/relux-works/curator/pull/70
- Reviewed heads: 4d240bac6a30aea2585f566068b55a2b325cbef7 (rev1, ACCEPT: TASK-260915-4f44cs_review-verdict-pr70.md) and 559447efe4a9d6f0c5c0f2a9e254cd3cedea883d (rev2 amendment gating the rose-air lane on `vars.ROSE_AIR_RUNNER`, ACCEPT: review-pr70-rev2-and-rootctx-pr1.md, posted as PR comment).
- Hosted checks on 559447ef: 11 pass (Test/Race/Lint/Naming/Interop/Gate self-test across ubuntu/macos/windows); Test (rose-air) skipped because the repository variable was set after the run started; Candidate suite skipped by design.
- Landing: plain fast-forward push `559447ef:refs/heads/main` by reluxbot on 2026-09-15; curator main now 559447ef. Both commits signed by Ivan Oparin <oparin@me.com> (ECDSA SHA256:V6Ji…).
- Post-landing verification: `task-board q 'project_config(view=spawn-preflight, role=reviewer, agent=claude)'` on the tracked config admits exactly claude-fable-5-1:low; developer/codex admits exactly gpt-6-astra:low.
- Deviation recorded: the two exact-head reviews were performed by an independent Claude Fable reviewer launched outside task-board (Agent tool), because the pre-change ceilings could not admit the policy reviewer; from this landing on, reviewers are tracked task-board runs.
