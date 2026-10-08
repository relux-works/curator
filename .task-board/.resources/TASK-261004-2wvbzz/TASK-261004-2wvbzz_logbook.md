# Wave 2 research logbook

- Research-only pass at curator e5489b6ba22c9e9cf7a925e03f3653d115188999; isolated git-archive fixtures, no product edits.
- Confirmed N6: oversized synthetic helper output is truncated and accepted through Access.ReadHost with real Git.
- Confirmed N7: Global rejects a manager-owned legacy forwarding link, including a link emitted by Refresh, although ownership recognizes it. Foreign-file and late-drift controls preserve the foreign file.
- Confirmed N8: audit --allow accepts a non-digest path and writes trust.json outside its audit subtree in the temporary manager home.
- Confirmed N9: valid audit pins omit the creation time required by manager source-audit policy.
- Decision: revision-A umbrella outside-root dispatch is intentional warning-release behavior, not a newly reported defect.
- Validation corrections: the first install probe had the wrong OnStaged callback signature; the first conformance run used a relative root that Go resolved from package directories. Both errors are retained in evidence and the corrected subsets were rerun.
- Scope: Windows is read-only analysis. No real credentials, external package services, release installation, or destructive load tests were used. Repository LOGBOOK.md remains untouched as the current brief requires.
