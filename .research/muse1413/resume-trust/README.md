# Muse 1.4.13 resume/trust replay

This is the bounded P7 witness from `.research/project-exclusion-muse`,
updated for the exact 1.4.13 digest. Only fixture/checker/offline machinery
and the resume test are retained. No broader P1–P8 project qualification
is claimed. The source fixtures retain their historical `MUSE_EXCL` markers.

```sh
# Use the same private read-only copy as MUSE1413_BINARY.
export MUSE1413_RESUME_BINARY="$MUSE1413_BINARY"
export MUSE1413_RESUME_RUN_ROOT="$PWD/.temp/TASK-261026-jk47fn/resume-new"
env -u GIT_DIR -u GIT_WORK_TREE go test ./.research/muse1413/resume-trust -count=1 -v -json
```

No opt-in means every test skips before fixture creation. Wrong bytes refuse
before a case exists. Each case requires a fresh absolute `.temp/` directory.
Git creates a synthetic checkout using external metadata; no user Git config
or real repository trust is used. All vendor calls use `env -i`, fixture
HOME/XDG, file credentials, echo, no auto-update and an outer sandbox denying
network, security lookups and real native/credential directories. No owner
Muse home, credentials or sessions are read. MCP `/usr/bin/true` entries
measure configuration admission, not a working server or remote connection.

The test first verifies no project rules/skill/MCP or hook execution under
fresh untrusted exec, beside working selected home sources. It then invokes
real `resume --last`, verifies trust/quit without continue-untrusted, grants
trust only to its synthetic checkout, sends echo and `/rules`, exits normally,
and checks revived project skill/hook/MCP catalogs plus actual marker-hook
execution. Saved root rules remain `project: none`. Native children, lazy
reads and permission-policy enforcement are unmeasured.

Tmux uses `/dev/null` config, a private relative socket, background terminal,
recorded server/pane PIDs and PID-checked cleanup. Raw process/log/session
files stay ignored beneath the run root. Public output consists only of the
sanitized test stream and bounded [native excerpts](../evidence/resume-observations.txt).

Sandbox controls remove only network or synthetic Keychains denial. Run
these **only** against `TestOfflineBoundary`, never vendor tests:

```sh
MUSE1413_RESUME_MUTANT=network go test ./.research/muse1413/resume-trust -run '^TestOfflineBoundary$' -count=1 -v
MUSE1413_RESUME_MUTANT=keychains go test ./.research/muse1413/resume-trust -run '^TestOfflineBoundary$' -count=1 -v
```

Use a fresh run root for each command. Both commands must exit 1 at the
specific denial assertion; normal paired controls must pass. The checker
accepts zero and one positive controls and rejects malformed, failed,
duplicate or contradictory startup records. This is a failure witness,
not a complete resume exclusion recipe or runtime admission authority.
