# TASK-260728-wy3dsw manager-profile review

Verdict: CHANGES REQUESTED. Route to documentation rework, then require a fresh reviewer cycle.

## Scope and independent evidence

- Reviewed the task worktree read-only at pinned HEAD 57c1f56846d221ecc55786bd3c2467ec32f11730.
- `rsync -anic --delete` against accepted predecessor TASK-260728-17sclp reported only `profiles/manager.md` and `cli/curator.md`; the index is clean.
- Document hashes match the producer handoff: manager 44db1200a5b22f63785acf2ea304c20b0c26ec5a981b65818dbd097eb56b5839; CLI 6f2525612ba3b7e62cf94e8b859798d72e7e332b7d0f4a2406fa1329b69c059d.
- Independent `make validate` passed: 42 schemas, 400 vector files, 15 Python tests, and `go test ./tools/...`. `git diff --check` and `git diff --cached --exit-code` passed.
- Git 2.50.1 exact-init and full-OID `cat-file --batch=%(objectname) %(objecttype) %(objectsize)` smoke passed under the documented empty config/home/PATH environment.

## Required changes

1. Exact local config/ref admission is not actually pinned in manager section 11.4. The text delegates to undocumented Git quoted-value escapes and omits the architecture-v6 byte grammar for section/key tokens, accepted escapes, comment handling, boolean spellings, required `core.bare=false`, inert ordinary keys, exact security-key duplicate handling, `refs/replace` rejection, HEAD terminator rules, and loose-versus-packed precedence. Different conforming readers can therefore admit different source-controlled configuration/ref bytes. Replace the summary with the exact accepted subset or normatively reference an in-repository section containing that complete grammar.

2. The common commit/tag grammar in manager section 11.5 omits an accepted tag-integrity condition: for an exact tagged declaration, the outermost annotated tag object must name the requested tag. This requirement exists in `protocol/core.md` and architecture-v6 section 6.6.2, but the profile that claims to pin the raw-object process does not carry or reference it. Also make the extra-header/continuation grammar self-contained rather than leaving `ascii-key`, `bounded-value`, and ignored-extra continuation acceptance implementation-defined.

3. Stable diagnostics are incomplete for this task scope. Section 11.10 names the twelve architecture-level codes, then says schema/driver/descriptor/identity/credential/audit/protected-boundary/receipt/marker/artifact/currentness/compiler/transaction/signer failures must merely be typed. Decision 0005 explicitly delegates their CLI rendering to the manager profile, and the CLI supplies only one JSON example. Define stable codes and phase/state/severity rendering for those fail-closed classes, including package-controlled argv/env/output/credential/signing rejections, or normatively map them to already defined stable project codes.

4. Separate syntax-only offline reporting from install dry-run reporting. Manager 11.7 and the CLI dry-run block list `unverified-offline` as a dry-run state, while the same section says that state is legal only for syntax-only validation and that install/update/repair/coverage audit fail before mutation without exact source. Publish disjoint command/state rules so an implementation cannot treat `curator install --dry-run` as syntax-only success.

5. Complete the exact SSH wrapper invocation tuple by specifying and checking `argv[0]` as the absolute manager wrapper, matching architecture-v6 section 6.4; section 11.3 currently lists only `argv[1]`, `argv[2]`, and `argc`.

These are bounded documentation defects, not a stop-the-line boundary. No code changes were made during review.