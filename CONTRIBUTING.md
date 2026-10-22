# Contributing to Curator

Curator follows a small set of working agreements. They are binding for every change, human or agent authored. The full engineering plan lives in [docs/implementation-plan.md](docs/implementation-plan.md).

## Workflow

- Work is tracked on the in-repo task board [.task-board/](.task-board/). Every change starts from a task; create the story or task card first, keep `progress.md` status current (todo, in-progress, review, done), and attach research artifacts under `.task-board/.resources/<ID>/`.
- Epics map one to one onto the plan phases; respect the dependency order in the plan (section 5).
- Spec first: the protocol specification (cited as `Spec §N.M`) is the source of truth. When implementation and spec disagree, stop and resolve the spec question before coding around it.

## Commits

- Discrete and meaningful: one logical step per commit, imperative subject, a body explaining what and why.
- Signed: SSH signature, verified on GitHub.
- Tests accompany the change in the same commit; a phase closes only with green CI on ubuntu, macos, and windows.

## Test environment

CLI tests discard the caller's PATH in `TestMain`. They retain a private fixture
directory and the system command directories (`/usr/bin` and `/bin` on Unix;
System32 and the standard Git for Windows command directory on Windows).
Individual provider tests can still supply an explicit fixture PATH, including
untrusted providers. Do not append developer tool directories to the default
test PATH.

Real Go launches in the CLI and script-worker fixtures use
`internal/testtoolchain.Go`: an absolute `CURATOR_GO`, then `GOROOT/bin/go`, then
the test binary's build GOROOT. There is no PATH fallback. Rust metadata probes
use the pinned native installation beside the Cargo selected by the product.
An absent toolchain skips only its dependent test with a reason; an invalid
selection fails. Run package tests with `go test ./cmd/curator/...` and the
affected internal packages. `TestCLIPathIsolation` plants ambient providers;
`TestCLIPathIsolationNarrowingMutant` proves that restoring ambient PATH breaks
the real status witness. Provider fixtures use test temp directories;
`TestMain` removes its private PATH directory when the package run exits.

## Code

- Go, stdlib first. New third-party dependencies only when the plan names them (section 2); anything else needs a plan update first.
- A new language source-closure adapter follows [docs/authoring-language-adapters.md](docs/authoring-language-adapters.md): the shared contract, the two allowed process seams, the reject-by-default source-analysis posture, and the conformance obligations a new path must publish.
- Windows is a first-class target: path handling, shims, and configuration surfaces are tested in the CI matrix.
- Error messages that the spec words normatively keep the same information content.

## Naming

- The name of the alternative protocol implementation appears exactly once in this repository: the README section about the open protocol. CI enforces this with a case-insensitive grep (zero matches outside README.md, exactly one inside). Protocol file names (`Skillfile.json`, `csk-skill.json`, `.csk-install.json`, `.csk-managed.json`) are part of the wire format and are used as-is.

## Prose

- Docs and board cards are plain technical English: no em dashes, no guillemets, human grammar.

## License

Apache License 2.0. By contributing you agree that your contributions are licensed under it.
