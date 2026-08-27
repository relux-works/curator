# External build repositories v3 — independent review verdict

Verdict: **changes requested**. Route `TASK-260720-1nvomm` to `analysis` for architecture rework and a fresh reviewer cycle.

Reviewed artifact: `TASK-260720-1nvomm_external-build-repositories-architecture-v3.md`, SHA-256 `eb2a219d61a7a53709e588c5abc026b5f0ef34e7f235944432795bcc8861b966`. The active run `RUN-260727-863275` is not goal-bound.

## Accepted portions

- Schema 7, `go-repository-v1`, descriptor target selection, receipt v2, marker v3, and rc.5/claim-v3 separation preserve schemas 1 through 6 and frozen `go-v1` semantics.
- Logical descriptor targets plus manifest command keys satisfy repository target selection without reopening executable-name or output-path control.
- Declared/effective identity, substitution binding, mixed receipt-v1/v2 marker-v3 behavior, independent whole-snapshot audit before artifact-cache lookup/compiler, protected-cache admission, status/repair/GC, rollback, signing separation, future-driver review, and board-impact mapping remain sound.
- The exact fresh-repository init and fetch vectors close tag, `FETCH_HEAD`, maintenance, commit-graph, submodule, refmap, filter, server-option, and upload-pack defaults. Full-object and exact-tag flows worked on Git 2.50.1 for SHA-1; SHA-256 protocol-v0 full-object fetch also worked.
- `GIT_SSH_VARIANT=ssh`, the three-argument wrapper contract, and the fixed OpenSSH option set close auto-detection, user/system configuration, prompts, TTY, forwarding, proxy/jump, local command, and connection sharing for the tested family.

## Required changes

### 1. Raw network object proof contradicts the exact child graph

Section 6.1 says the permitted child graph is exactly manager to trusted Git for `init` and `fetch`. Section 6.6 then permits `git cat-file --batch` and `git ls-tree` for network extraction, but neither adds them to that graph nor defines their exact argv, request stream, output grammar, no-filter/no-textconv/no-mailmap controls, limits, and `--no-replace-objects` / `--no-lazy-fetch` placement. A conforming implementation can therefore either violate the stated exact graph or invent a different extraction process.

Select one normative path. Either require the manager pack/object parser for network and local sources, or add the trusted read-only Git plumbing children to the graph and specify byte-exact direct argv/environment/input/output contracts. No audit or cache lookup may occur until that selected proof completes.

### 2. Local pack/index support is implementation-dependent

Section 6.5.3 accepts only documented pack/index formats that each implementation declares and lets an implementation reject an encountered documented version. That makes the same admitted local substitution succeed in one manager and fail in another, despite schema 7 being an interoperability contract.

Enumerate the exact v1 supported pack and index versions and checksum/hash-family rules, with one common stable rejection for all others. A simpler equally closed option is to byte-parse only config/refs and copy the admitted object-store bytes into a manager-private repository, then use the same fingerprinted, clean-environment, no-network, no-replace/no-lazy trusted Git plumbing contract from finding 1 while manager code recomputes every consumed object ID. This executes no source-selected Git behavior and avoids two independent custom pack parsers.

### 3. LFS rejection is asserted but not defined

Section 6.6 says a canonical Git LFS pointer fails, and the threat table says pointers are rejected, but no normative pointer grammar, scan scope, or error is defined. Raw extraction otherwise materializes the pointer as an ordinary regular blob; a Go `embed` input can successfully compile those pointer bytes. Git LFS documents that pointer files are the Git-resident bytes and actual content is external.

Either define an exact conservative all-blob Git LFS pointer detector from the official pointer grammar and a stable pre-audit rejection, or explicitly change the contract to treat pointer bytes as ordinary source while only hydration remains unsupported. The current MUST-level outcome cannot be implemented consistently.

## Validation evidence

- All nine fenced JSON blocks parsed.
- Git 2.50.1 accepted the exact init/common-fetch options; SHA-1 full-OID and exact-tag destinations were the only refs, with no `FETCH_HEAD` or commit graph. SHA-256 protocol-v0 full-OID fetch also passed.
- The fixed OpenSSH option vector parsed successfully on the host OpenSSH family.
- Host `python3 tools/validate.py` lacked `jsonschema`; the existing pinned task venv passed: `validated 30 schemas and 93 vector files`.
- `git diff --check` passed.
- Primary evidence: [git-fetch](https://git-scm.com/docs/git-fetch), [git-config SSH variants](https://git-scm.com/docs/git-config), [repository-version](https://git-scm.com/docs/repository-version), [repository layout](https://git-scm.com/docs/gitrepository-layout), [pack format](https://git-scm.com/docs/gitformat-pack), [OpenSSH ssh](https://man.openbsd.org/ssh.1), [OpenSSH client config](https://man.openbsd.org/ssh_config.5), and [Git LFS pointer specification](https://github.com/git-lfs/git-lfs/blob/main/docs/spec.md).

These are ordinary architecture rework items, not a stop-the-line decision. No product, spec, schema, code, test, release, prior producer resource, or prior verdict resource was modified by this review.