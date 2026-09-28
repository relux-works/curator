# BUG-260922-k6eypp — gate fix 2 (THE ONLY CURRENT INSTRUCTION, with k6eypp-fix-1.md)

Revision 1 (tree a95804da) is green except `Gate self-test (windows-latest)` (run 36359040455): the new Homebrew rows in
.github/ci/gate-selftest.sh fail on Windows Git Bash ("the Homebrew fixture links rustup from bin into its versioned keg" … "rustc is not on
PATH"): `ln -s` there does not create a real symlink (MSYS copies or needs developer mode), so the keg resolution the fixture relies on
cannot happen. Homebrew exists only on macOS/Linux runners. Make the Homebrew fixture rows run on non-Windows hosts and print an explicit
`skip (Homebrew layout is POSIX-only)` line on Windows (detect via the same OS check the selftest already uses for other POSIX-only rows),
keeping every other Windows row unchanged; the mutant (keg step removed) must still fail on linux/macos. Do not change the installer logic
unless the Windows lane of the installer itself needs it (it should not — rustup on the Windows runner lives in CARGO_HOME).
Set status development; update the results ("Revision 3 — Homebrew selftest rows POSIX-only"); handoff; END YOUR TURN. No CHANGELOG/LOGBOOK.

The previous run republished the IDENTICAL tree (rev2 == rev1 == a95804da, run 36364906427 failed the same way) — it changed nothing. You
MUST edit .github/ci/gate-selftest.sh this time: wrap the Homebrew-fixture block (the rows "the Homebrew fixture links rustup from bin into
its versioned keg" … "the Homebrew GITHUB_PATH writes are CARGO_HOME, prefix, then keg proxies") in a host check such as
`case "$(uname -s)" in MINGW*|MSYS*|CYGWIN*) echo "skip  Homebrew rustup layout (POSIX-only; Git Bash ln -s is not a symlink)";; *) … ;; esac`
(or the selftest's existing Windows-detection idiom if one exists — `RUNNER_OS`/`OSTYPE`), keep the rows unchanged on linux/macOS, and show
`git diff a95804da` non-empty in the results. Then update results, handoff, END YOUR TURN.
