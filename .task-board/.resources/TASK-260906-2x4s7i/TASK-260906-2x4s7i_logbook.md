# Flight Logbook

> Institutional memory. Concise, factual, high-signal.
> Newest entries first. One block per insight.

## 2026-09-06

### 1830 — PR 47: the F4 fix reintroduced the F1 class on a neighbouring Windows spelling
- FINDING: `schemas/v1/manager-config-v2.schema.json` `$defs/overlay` second `allOf` arm decides scheme-ness with `^[A-Za-z][A-Za-z0-9+.-]*://`. `*` admits a one-character scheme; a Windows drive letter is exactly one character.
- REGRESSION: `C://Users/operator/context` is INVALID in all 8 member columns (`then: false`). `origin/main` accepted the same source with a form. A change made to *create* a declaration removed the only shape that spelling had.
- ANOMALY: the two arms disagree about `C:` — arm 1 works hard to keep a drive letter out of the git classification, arm 2 readmits it as a scheme. `1://y` is VALID, `x://y` INVALID, because the scheme test starts with `[A-Za-z]`.
- FIX: `*` → `+` in that arm's scheme pattern restores the drive-letter class in both directions with `tools/validate.py` exit 0 over the whole corpus. Producer owns the fix.
- SCOPE: `schemas/v1/manager-config-v2.schema.json`, `conformance/v1/schema-cases/manager-config-v2/`.
- STATUS: TASK-260906-3x0w4y routed to `development`; PR 47 not safe to land.

### 1832 — A regex discriminator's unpinned halves survive an 18-case corpus
- FINDING: the core §6.1 host grammar `[A-Za-z0-9][A-Za-z0-9.-]*` — presented as half the F1 repair — is unpinned in **both** directions. Widening it to `[^/:\s]+` and narrowing it to ≥2 characters each leave `tools/validate.py` at exit 0. Only the backslash exclusion is load-bearing.
- FINDING: `directory` on a `path` source is refused by no case (narrowing mutant survives) though environments §1 and the AC both name it beside `range`/`tag`/`branch`/`revision`, which are all pinned.
- NOTE: 18 new cases raised distinct overlay `source` spellings from 4 to 16 and killed 17 of 21 mutants — breadth is not the same as coverage of each sub-behaviour of one regex.
- DECISION: `file:` URL admitted as a form-free `path` is wrong and now pinned by a published positive conformance case. §1 `path` is a directory operand; core §6.1 says only that a `file:` URL has no *network* identity, which does not make it a path.
- NOTE: entry kept in `.temp/review-2x4s7i/` rather than the repo root — the review is read-only outside scratch and a root `LOGBOOK.md` would be a stray file in a tracked tree. Durable record is the board artifact `TASK-260906-3x0w4y_review-findings-2.md`.
