# Review note — TASK-260910-3i6vod sandbox posture docs (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Docs-only rev1 (base 3f60f7f0, tree 4a1579d6, 2 paths: README.md, SECURITY.md). Its gate run 36467710280 failed ONLY on macOS in
internal/envprofile TestPathInstallCapturesDirtyUntrackedInsideGit — an unrelated flake (git background maintenance removed
.git/objects/maintenance.lock during the E6 path-boundary walk); tracked and being fixed as BUG-260928-uyak0e. Confirm the docs diff cannot
affect that test. Review content: README and SECURITY state plainly that installed commands run with the user's privileges under portable
assurance and that script-worker-v1 enforced commands and verified mode are the enforcement paths; claims match curator-spec rc.13
(assurance.md, profiles/manager.md §3.1) and the code (no overclaiming — e.g. verified providers are not shipped yet; say so); links valid.
accept_cr or changes requested with file:line. No LOGBOOK.md.
