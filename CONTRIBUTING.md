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

## Code

- Go, stdlib first. New third-party dependencies only when the plan names them (section 2); anything else needs a plan update first.
- Windows is a first-class target: path handling, shims, and configuration surfaces are tested in the CI matrix.
- Error messages that the spec words normatively keep the same information content.

## Naming

- The brand name of the reference implementation does not appear anywhere in this repository. CI enforces this with a case-insensitive grep. Protocol file names (`Skillfile.json`, `csk-skill.json`, `.csk-install.json`, `.csk-managed.json`) are part of the wire format and are used as-is.

## Prose

- Docs and board cards are plain technical English: no em dashes, no guillemets, human grammar.

## License

Apache License 2.0. By contributing you agree that your contributions are licensed under it.
