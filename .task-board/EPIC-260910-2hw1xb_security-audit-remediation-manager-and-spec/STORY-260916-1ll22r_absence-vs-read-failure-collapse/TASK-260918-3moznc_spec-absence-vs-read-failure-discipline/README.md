# TASK-260918-3moznc: spec-absence-vs-read-failure-discipline

## Description
curator-spec: make the environments §8.4 absence-vs-read-failure discipline structural in the text — one general rule stated once (a failed read, stat, or parse of any state or surface file is never reported as absence; each read site names its unreadable diagnostic) referenced from every section that reads markers, seeds, locks, ledgers and passthrough entries, with a closed table of the unreadable outcomes per file class, and conformance vectors exercising unreadable-but-present for markers, seeds, locks and passthrough entries.

## Scope
(define task scope)

## Acceptance Criteria
§8.4 (or a new subsection it references) states the general rule once; every read site section (§1.3 lock, §7.4 seeds/passthrough, §8.2 marker, §8.3 ledger/backups, §9.5/§9.6 inventories, §10.1) references it and names its unreadable diagnostic from the closed table; vectors cover unreadable-but-present for markers, seeds, locks and passthrough entries with a rule-7 scenario-pinning validator gate; existing vectors byte-identical; make validate and regenerate-check exit 0; CHANGELOG entry.
