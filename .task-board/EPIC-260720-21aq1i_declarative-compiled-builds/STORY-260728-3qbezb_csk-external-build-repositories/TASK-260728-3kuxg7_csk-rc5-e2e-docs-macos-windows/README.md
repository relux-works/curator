# TASK-260728-3kuxg7: csk-rc5-e2e-docs-macos-windows

## Description
Complete cocoaskills/csk documentation and rc.5 end-to-end qualification for external build repositories on macOS and Windows after the independent implementation is accepted. Exercise shared fixtures plus real repository access, protected offline reinstall, project/global activation, and lifecycle failures.

## Scope
csk docs/examples, shared-suite consumer, clean Python tests, native runs using SSH aliases relux and win, exact spec/Curator compatibility evidence, and release qualification. Linux remains excluded for the later dedicated story.

## Acceptance Criteria
Documentation is complete and does not recommend unsafe script wrappers or PATH hacks; csk passes shared rc.5 vectors independently of Curator internals; macOS and Windows evidence covers exact tag match/move/missing, inaccessible source, audit ordering, build/install, project/global PATH activation, cache corruption, repair, crash/rollback, uninstall, and schema-1-6 regression; exact OS, Python, Git, Go, csk, Curator and spec revisions are recorded; no Linux support is claimed.
