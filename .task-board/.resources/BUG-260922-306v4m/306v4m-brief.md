# BUG-260922-306v4m — make the rose-air rustup failure name its evidence (curator)

Control root: /Users/administrator/Developer/ReluxWorks/curator/curator; work only in your assigned
Story worktree (STORY-260915-3w11un). Read `campaign-producer-rules.md` first. Landing gate = hosted
CI (runtime runs it once at handoff).

## Situation
`.github/ci/install-rust-toolchain.sh` fails every main push since 2026-09-21T22:30Z with
`rust-pin: rustup is not installed on this runner; install it once per
docs/self-hosted-runner-setup.md, then re-run this lane` (runs 35663049586, 35725359745). The
OPERATOR HAS CHECKED the rose-air machine on 2026-09-22 and rustup IS installed there. So the
message is wrong about the remedy and, worse, carries no evidence. Three live hypotheses:
(a) a DIFFERENT self-hosted ARM64 mac took the job — the label set is `self-hosted, macOS, ARM64`
    and the runner is registered at organisation level (GitHub's API reports `runnerName: null` for
    that job), so the machine the operator inspected may not be the machine that ran;
(b) the runner service user differs from the user that owns rustup, so `$HOME/.cargo/bin` resolves
    to another home;
(c) rustup is somewhere the probe does not look (asdf/mise shim dir, `~/.local/bin`,
    `/usr/local/cargo/bin`, `/opt/rust/bin`, a non-default `CARGO_HOME`).

## Deliverable
1. On the failure path ONLY, print one diagnostic block to stderr before the existing remedy
   sentence (keep that sentence, the exit code and the success path exactly as they are):
   - `RUNNER_NAME` (GitHub sets it — this is what identifies the machine), `RUNNER_OS`,
     `hostname`, `whoami`, `HOME`, `CARGO_HOME` (and whether it was set or defaulted),
     `HOMEBREW_PREFIX`, and the `PATH` the script searched;
   - for EVERY probed candidate, one line: the exact path and `executable` / `exists-not-executable`
     / `absent`;
   - a bounded listing (at most ~20 entries) of `$CARGO_HOME/bin`, `/opt/homebrew/bin` and
     `/usr/local/bin` where they exist, filtered to names containing `rust` or `cargo`;
   - `command -v rustup` output and, if the shell has it, `type -a rustup`.
   No secrets: do not print the whole environment, only the named variables.
2. Keep the success path byte-identical in behaviour: same `GITHUB_PATH` writes, same
   `rust-pin: using rustup at <path>` line, same ordering.
3. `.github/ci/gate-selftest.sh` rows: (i) a fixture with rustup nowhere — the block is printed, the
   remedy sentence is still the last line, exit 1; (ii) a fixture with rustup only under a probed
   candidate — success, no block; (iii) a narrowing mutant (drop one probed candidate from the
   diagnostic enumeration) fails a named row. Fixtures must not depend on the host's real rustup.
4. Docs: `docs/self-hosted-runner-setup.md` gains one short paragraph saying the lane names the
   runner and the probed paths on failure, so the next red run is self-diagnosing.

## Boundaries
No change to which paths are probed in THIS revision unless the evidence already in hand proves a
missing one — the point is to get the evidence first. No product code. CHANGELOG entry.
Attach `BUG-260922-306v4m_results.md` (the block's exact shape, the three rows with exit codes) and
hand off with `task-board handoff BUG-260922-306v4m --role developer`.
