# TASK-260930-2mtgv7 — reconcile compiled-build board leaves with curator main (THE ONLY CURRENT INSTRUCTION; read-only research)

The board shows many curator leaves in EPIC-260720-21aq1i "declarative compiled builds" as backlog, yet main already carries a lot of
that work:
- schema 7 external build repositories were integrated on 2026-08-05 in commit 5709fc99 (internal/buildrepo, marker v3,
  install/external.go, skillspec schema 7);
- STORY-260720-3plyvy, the curator Go build driver, is done except BUG-260730-3eqseq;
- toolchain probe / dry-run work has landed;
- the Rust, Swift and Kotlin driver stories are done.

For EVERY open leaf in these stories, read its README, then decide against origin/main and curator-spec v1.0.0-rc.13 (SPEC_PIN
23435129; conformance root in the protocol-spec submodule or a disposable clone):
- STORY-260728-1ojb1p curator-external-build-repositories: 1ax4j0, 1uepyd, 20ao7p, pwbr32, rjxrgs;
- STORY-260728-2fsqtv compiled-build-toolchain-preflight: 2gbtb9 (curator), 3vtl57, ypbuav. 1j72zq is csk's: note it, do not assess it;
- STORY-260720-21bsr2 compiled-build-interop: every leaf. Mark which ones need csk;
- STORY-260728-19nx3g external-build-repository-interop: every leaf;
- STORY-260728-1eye8p linux-native-external-build-validation: 1e6811, 1skseh, 2sxx7k (blocked; say on what);
- STORY-260720-3plyvy: BUG-260730-3eqseq repair-curator-v0-13-main-ci. Is it obsolete?

Per leaf, give exactly one verdict:
- DONE-ON-MAIN, with file:line of the implementation, the test names that prove it, and the rc.13 vectors/cases driven (name the
  conformance family and counts);
- PARTIAL, with what exists and the precise gap;
- NOT-STARTED;
- OBSOLETE, with the reason (for example, superseded by a later schema or release).

Also state whether the leaf needs csk or human-only access, such as ssh to the lev Linux host.

Run the focused tests you cite, with real exit codes and bounded runs (for example ok  	github.com/relux-works/curator/internal/buildrepo	231.176s,
ok  	github.com/relux-works/curator/internal/install	19.387s). Write NO code and change no board status. Put the table in the results resource
`TASK-260930-2mtgv7_results.md`, followed by a proposed plan for the NOT-STARTED/PARTIAL leaves, in dependency order and with rough sizes. Never
spell any employer name.

Then run `task-board handoff TASK-260930-2mtgv7 --role researcher` and END YOUR TURN.
