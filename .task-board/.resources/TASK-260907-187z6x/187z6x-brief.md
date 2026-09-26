# TASK-260907-187z6x — git same-source reinstall drops --use and --takeover (curator)

Control root: /Users/administrator/Developer/ReluxWorks/curator/curator; work only in your assigned Story
worktree `.temp/STORY-260906-1a2i5a/worktree`. Read `campaign-producer-rules.md` first. Landing gate =
hosted CI (runtime runs it once at handoff); run the narrow package tests yourself with real exit codes.

Defect (stage (c) review cycle 6, C6-m1 = repeat of C5-M1; a trunk defect): `curator profile install
<git-url> --use --takeover` on an ALREADY-installed git root (the retry the §9.5 onboarding stop invites)
accepts both flags, does nothing, and exits 0 saying "updated profile". Route: internal/envprofile/
envprofile.go install path (≈ lines 640–670): "A git reinstall is an update, so it delegates to
updateLocked" (≈ :909) — `updateLocked` has no --use/--takeover activation handling, whereas a path
reinstall goes through `reinstallPathLocked` (≈ :731) whose `reinstallActivation` (≈ :836) honours the
flags. The flags DO work on a git FIRST install.

Fix: a same-source reinstall of a git root honours `--use` and `--takeover` exactly as a first install
does (activate when --use is given and the root is not current; `--takeover` without `--use` behaves as
the path reinstall rule already states), or refuses with a diagnostic that tells the operator what to run
instead — pick the path-reinstall parity unless you find a normative reason (environments §9.1/§9.5,
cli/curator.md install row in the pinned curator-spec) that forbids it; cite it either way. No path-root
behaviour change.

Tests: drive `run()` (cmd/curator production entry) for all four flag combinations (none, --use,
--takeover, --use --takeover) on a git root that is current and one that is not — eight rows — using the
repository's hermetic fixture pattern: a local repo served under a fake canonical identity through a
`GIT_CONFIG_GLOBAL` insteadOf rewrite (see internal/envprofile/network_fixture_test.go and
cmd/curator/profile_test.go; `file://` is refused by canonicalGit — do not use it). Assert the activation
outcome (current profile, materialized surfaces, backup for takeover) and the exact operator line. One
narrowing mutant (drop the activation call on the git reinstall branch) must kill a named test. Register
platform-case rows if the ledger demands it.

CHANGELOG (unreleased, Fixed). Attach `TASK-260907-187z6x_results.md` (defect, fix, eight-row table,
mutant, narrow `go test ./internal/envprofile ./cmd/curator -run … -count=1` exit codes) and hand off with
`task-board handoff TASK-260907-187z6x --role developer`.
