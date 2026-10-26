# TASK-261026-jk47fn observation logbook

- 2026-10-26: Fetched origin/main, fast-forwarded local main, verified both
  refs equal requested `c7b39b5474f68db1064256a91e59a9cdfb66b49b`, then created
  the qualification feature branch. Initial checkout was clean. Existing
  canonical task assigned and estimated before development. Scope/AC placeholders
  were replaced with six rows derived from the user's request.
- Copied the installed 1.4.13 executable to ignored task-local storage, mode
  0500, and verified source/copy SHA-256. The updating source and launcher
  were not executed or changed. No owner Muse home, auth or session read.
- Reused the 1.4.11 harness, fixtures and original reviewer test names.
  Byte-identical historical snapshot inputs retain `muse148-probe` and frozen
  digests. New intake names, sentinel and exact binary digest identify 1.4.13.
  Tmux explicitly uses `/dev/null` config to avoid owner configuration.
- First external-spec snapshot attempt inherited the session's `GIT_DIR` and
  archived the active repository instead; dependency installation failed,
  exit 1, on missing requirements. No reference test consumed that tree.
  Cleared `GIT_DIR`/`GIT_WORK_TREE` for external Git and test subprocesses;
  fetched actual spec main `88b0d8615a5df8728e0e9540a9edd708c0ca5660`, took a
  separate snapshot and installed its pinned requirements successfully.
  Both attempts remain private under task-local `.temp/`.
- Reread environments §§7.1, 7.9 and 10.2, CIP-0017, Muse v4 conformance
  README and CIP-0002's resume/source-exclusion boundary. Read prior 1.4.11
  P7 evidence from `project-source-exclusion-muse.md` and its harness.
- Full main replay exited 0, 45 loader/parser cases: 43 vendor exits 0,
  two intended parser exits 2; isolated version inspection exited 0.
  All 45 normalized case comparisons match 1.4.11. Exec excludes foreign
  sources; interactive still retains one skill despite the valid boolean.
- All ten patch mutants fail their intended assertion in disposable module
  copies. Exec omission fails `user_skills=1; want 0` beside its passing
  no-exclusion control. Interactive omission survives because exclusion is
  already ineffective; no interactive protection is established.
- Ported only P7 resume machinery, checker and offline controls from the
  prior project-exclusion harness. Initial resume-01 exited 1 after correctly
  observing revived catalogs: the port renamed the expected hook marker but
  retained unchanged historical fixture bytes. Restored expected literal
  `MUSE_EXCL_HOOK`; resume-02 passed. Added explicit pre-trust negative
  skills/hook/MCP assertions and home-hook positive; resume-03 passed.
- Initial lint exited 1 for an inherited unused request parameter and a
  helper left unused after narrowing the port to P7. Removed unused helper,
  named the parameter `_`; final lint, vet and resume-04 pass. Earlier
  passing resume runs are not the final-source evidence; resume-04 is.
- Resume-04 confirms trust/quit prompt, no continue-untrusted option, saved
  root listing with no project rules, and live project skill/hook/MCP revival
  after fixture-only trust. Project hook marker is absent before trust and
  contains exact synthetic bytes afterwards. Finding remains open and applies
  directly to 1.4.13. No complete strict resume exclusion is qualified.
- Resume network and synthetic-Keychains denial mutants both exit 1 at their
  intended assertions. Neither mutant runs Muse under a weakened sandbox.
  Ordinary/blocked positive and negative controls pass in the real policy.
- 1.4.12 is intentionally skipped as superseded on every known host per task
  intake. Task-board canary and additive-MSP reports do not override direct
  exclusion evidence. No runtime/registry/credential policy change is made.
- Complete sanitized command/test streams, source digests, case comparison
  and native resume excerpts are committed under this module. Absolute paths,
  raw synthetic sessions, tmux PIDs and installed-binary locators remain ignored.
  PR exact-head checks and review/handoff evidence are attached to the task.
- Final-source harness lint and ungated suite pass as harness-lint-03 and
  ungated-02. A full added-file `git diff --check` reports the mandatory
  space-plus-tab context encoding in unified mutant patches. The non-patch
  diff passes separately; all patch encodings were validated by actual
  application and intended failing behavioral tests, not stripped as prose.
