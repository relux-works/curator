# External build repositories v6 — independent review verdict

Verdict: **ACCEPTED**. Route `TASK-260720-1nvomm` to `done`.

Reviewed artifact: `TASK-260720-1nvomm_external-build-repositories-architecture-v6.md`, SHA-256 `2abae77d80eba6789f9911db7e9722595b4f21ba47391ca9eafd0064af03d67e`. Active run `RUN-260727-825e19` is not goal-bound and had no pending directives.

## Findings

No blocking or rework finding remains. Revision 6 closes the sole v5 defect without regressing the previously accepted architecture.

- The v5-to-v6 diff is limited to optional-tag acquisition/verification, related ordering and error semantics, receipt/marker/repair wording, six new conformance cases, sources, and evidence. Previously accepted schema-7, source identity, Git object boundary, audit, receipt-v2/marker-v3, cache, lifecycle, signing, compatibility, and future-driver text is otherwise byte-identical.
- Every unsubstituted declaration with `tag` uses exactly `refs/tags/<tag>:refs/curator/tag` as its sole acquisition attempt in fresh private state. It never attempts the locked OID, branch, all-tags, or another ref as fallback. Untagged declarations alone use the full locked OID.
- Manager parsing recomputes lightweight/annotated tag objects with the common network/local byte grammar, enforces target types and chain bounds, and requires the terminal commit to equal `locked_commit.hex`.
- Matching, moved, missing, malformed/incomplete, and transport outcomes are normatively distinguished. The tag check gates audit success, artifact-cache lookup, compiler execution, receipt/marker publication, and repair. Syntax-only offline checks only warn; install/update/repair/coverage audit fail before mutation when the exact tag cannot be obtained.
- Receipt v2 and marker v3 record the declared tag only after the producing unsubstituted operation has completed the tag/lock check. They do not add a forgeable verification boolean. Read-only status validates protected snapshot/receipt/marker relationships and does not contact the remote merely to retest later tag movement; missing or unprovable protected evidence is not current. Repair always reacquires through the exact-tag-only path.
- Existing schema 1-6 and `go-v1` semantics remain frozen; schema 7, `go-repository-v1`, receipt v2, marker v3, and rc.5 remain a separate future delivery subtree.

## Independent evidence

- All 13 fenced JSON blocks parsed.
- Git 2.50.1 six-case smoke passed with `uploadpack.allowAnySHA1InWant` false and true. Matching annotated tags resolved to the lock; moved lightweight tags resolved to a different commit and mapped to `build_repository_ref_moved`; missing tags made fetch exit 128 and mapped to `build_repository_source_unavailable`. Each success exposed only `refs/curator/tag`; all six cases left `FETCH_HEAD` absent; no direct-OID refspec was attempted.
- Official Git fetch documentation confirms that explicit tag refspecs fetch the named tag despite `--no-tags`, `--no-write-fetch-head` suppresses `FETCH_HEAD`, and empty `--refmap` ignores configured refspecs. Official Git configuration documentation confirms the direct-object want controls are server policy and therefore irrelevant to this exact-tag-only path.
- Pinned validation environment: `tools/validate.py` passed 30 schemas and 93 vector files; `make validate` passed those checks, 8 Python tests, and `go test ./tools/...`; `git diff --check` passed.
- The host `python3 tools/validate.py` lacks `jsonschema`; the already pinned task-local venv contains jsonschema 4.25.1 and passed the required validator. This is an environment issue, not an artifact defect.
- The first temporary smoke inherited commit-signing configuration and failed before creating its fixture commit; rerunning in fresh temporary state with signing explicitly disabled produced the passing evidence above. No project/spec/schema/code/test/release or producer/prior-review resource was modified.
