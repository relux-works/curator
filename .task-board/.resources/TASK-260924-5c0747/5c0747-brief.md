# TASK-260924-5c0747 — curator v0.15.0-rc.2 release prep (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md`. curator-spec v1.0.0-rc.13 is RELEASED: tag v1.0.0-rc.13 → commit 23435129ebc4c29e5b7f75ec72a0aa0cd3f16065
(release workflow green). Follow EXACTLY how v0.15.0-rc.1 was cut (release-prep commit d1bb0a4e "Keep a release candidate out of the install
channels" and its neighbours: CHANGELOG section, version references, release.yml / goreleaser / gates) — do not invent a new process.
1. SPEC_PIN: move every spec pin from the interim curator-spec dcc7f015 (TASK-260922-18ex37) to the rc.13 tag commit 23435129… (ci.yml SPEC_PIN,
   conformance tests, docs/ci-gates.md, conformance-case-counts, gap ledger). The rc.13 vectors are labelled protocol_version 1.0.0-rc.13
   (TASK-260926-4hd81z already accepts that label in internal/scriptpolicy) — find and fix any other place that assumes rc.9 labels.
2. Released skillfile-sources corpus: move the vendored conformance/skillfile-sources-v1 pin (TASK-260924-20o9dk vendored it at 5746367) to
   the rc.13 tag commit; bytes should be identical — prove it (or re-vendor and report every changed file).
3. CHANGELOG: create the v0.15.0-rc.2 section from EVERY "## CHANGELOG entry (for release prep)" section in the results resources
   (.task-board/.resources/*/*_results.md) of leaves that LANDED on main after v0.15.0-rc.1 (check each is done/landed; `git log
   v0.15.0-rc.1..origin/main` to confirm). Copy each entry text verbatim, grouped as the rc.1 section groups them; list the source task per
   entry in your results (not in the CHANGELOG). Anything already in CHANGELOG stays.
4. Bounded local runs of the conformance packages and gate scripts you touch (host memory is tight — never the whole module at once); the
   hosted gate is the arbiter. Results must give the exact signed-tag command for the orchestrator and the release workflow to watch.
No LOGBOOK.md. Attach results, check DoD, `task-board handoff TASK-260924-5c0747 --role developer`. A write-boundary `policy warn` block is a warning.
5. Trunk may move while you work (TASK-260923-2gt5f6 and others are landing): before handoff `git fetch origin main`; if it moved, combine
   the new diff (keep both sides) and add the newly landed leaves' CHANGELOG entries.
