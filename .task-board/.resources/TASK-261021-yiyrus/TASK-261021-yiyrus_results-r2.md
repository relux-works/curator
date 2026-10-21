# TASK-261021-yiyrus — revision 2

## Scope and commit

Reapplied the previously reviewed CIP-0022 commit with a signed cherry-pick onto
current `origin/main`. No conflicts occurred and no manual edits were needed.
The CIP-0016 amendment remains in the base. The proposal remains Draft.

- Branch: `cip-mcp-in-project-skillfile-r2`
- Base: `3a3e041b3fbf76292e69b6abb4a0eca855524007`
- Original commit: `38aa3a0c777400a448f3eb7e7b2fed57a85654e2`
- New head: `c8b6245485c1a155fdf5148b14107712784a9021`
- Signature: `git log -1 --format='%G?'` returned `G`.

The original branch is preserved. No board status changes or pull requests.

## Files changed

Exactly the original five-file change: 489 insertions, no deletions.

- `cips/CIP-0022-mcp-server-declarations-in-project-skillfile.md`: Draft proposal.
- `cips/README.md`: Draft index entry.
- `README.md`: proposal discovery entry.
- `CHANGELOG.md`: Unreleased entry.
- `UNRESOLVED_QUESTIONS.md`: proposal decision pointer.

No normative specification, schema, vector, or executable tool changes.

## Diff of diffs

Command (exit 0):

```sh
git range-diff --no-color 38aa3a0c^..38aa3a0c HEAD^..HEAD
```

Output:

```text
1:  38aa3a0 = 1:  c8b6245 docs: draft CIP-0022 for project Skillfile MCP declarations
```

The patches are equivalent. There are no conflict-resolution hunks or additional
changes to review. The commands above and below use the explicit repository git
directory and work tree; the validation clone unsets inherited git bindings.

## Validation

Platform: macOS arm64; Python 3.14.7, jsonschema 4.25.1, Go 1.25.5,
GNU Make 3.81. The disposable validation clone had a normal `.git` and the
exact new head. Makefile stage executions unset `GIT_DIR` and `GIT_WORK_TREE`
and used `TMPDIR=/tmp/cw-dev1`. Go build cache was inside the assigned workspace.
The Draft checker also unset both git bindings; its fixture used the clone's
ignored `.temp/` directory.

`make validate` was invoked: its schema/vector stage passed (74 schemas,
1294 vector files), then the combined invocation was deliberately interrupted
at the shell runtime budget during Python tests (exit 2). **The combined
invocation is not reported as passing.** The same unittest discovery was then
run in bounded groups, followed by the unchanged Go recipe. The completed
recipe stages provide passing evidence: 879 Python tests in 27 groups, no
skipped-test summary, and the Go package passed in 16.933s.

The partition retained every discovered test ID exactly once, in original
order. All test bodies and their existing claim comments remain unchanged.
The coverage check accepted the actual partition and specifically rejected
omission and duplication control plants. This checks the test harness, not
runtime implementation of schema 3. No product behavior or gates changed;
additional product narrowing mutants are inapplicable to this reapplication.

The Draft checker passed: required template metadata/sections, unique Draft
index, parseable JSON, neutral text and actual repository local links. Its valid
link control passed and its broken-link plant was specifically rejected as a
broken local link.

## Commands and evidence

Abbreviations: `WORK` is the assigned Werkstatt root, `SPEC=$WORK/curator-spec`,
`VAL=$WORK/.val/TASK-261021-yiyrus-r2`, `LOG=$SPEC/.temp/TASK-261021-yiyrus/r2`,
and `PY=$SPEC/.temp/TASK-261021-yiyrus/venv/bin/python3`.
`g` expands to `git --git-dir="$WORK/curator-spec.git" --work-tree="$SPEC"`.
The prescribed agent-git-env helper was sourced before remote operations and
the signed cherry-pick.

| Command/check | Exit | Result/evidence |
| --- | ---: | --- |
| Assignment query, four precondition gets/reads, ack | 0 each | Assignment confirmed before implementation. |
| `g fetch origin` | 0 | Base `3a3e041...` fetched. |
| `g switch -c cip-mcp-in-project-skillfile-r2 origin/main` | 0 | Fresh branch from requested base. |
| `g cherry-pick -S 38aa3a0c777400a448f3eb7e7b2fed57a85654e2` | 0 | No conflicts or manual edits. |
| `g log -1 --format="%G?"` | 0 | `G`; `head-scope-01.log`. |
| `g diff --check HEAD^ HEAD` | 0 | No whitespace errors; five original files only. |
| `g range-diff --no-color 38aa3a0c^..38aa3a0c HEAD^..HEAD` | 0 | Equivalent patches; `range-diff-01.log`. |
| `env -u GIT_DIR -u GIT_WORK_TREE git clone --no-hardlinks --no-checkout "$WORK/curator-spec.git" "$VAL"` | 0 | `clone-01.log`. |
| `env -u GIT_DIR -u GIT_WORK_TREE git -C "$VAL" checkout --detach c8b6245485c1a155fdf5148b14107712784a9021` | 0 | Exact head; normal `.git` verified. |
| `make validate` with virtualenv PATH and test environment above | 2 | Deliberate budget interruption; `make-validate-01.log`. |
| `python3 tools/validate.py` (first Makefile stage) | 0 | 74 schemas, 1294 vector files; completed before interruption. |
| `env -u GIT_DIR -u GIT_WORK_TREE TMPDIR=/tmp/cw-dev1 PYTHONDONTWRITEBYTECODE=1 GOCACHE="$WORK/.cache/go-build" "$PY" -B "$LOG/run-groups.py" N` | 0 each | All 27 groups below; exact original discovery. |
| `env -u GIT_DIR -u GIT_WORK_TREE TMPDIR=/tmp/cw-dev1 PYTHONDONTWRITEBYTECODE=1 "$PY" -B "$LOG/check-groups.py"` | 0 | Exact coverage and omission/duplication controls; `coverage-check-01.log`. |
| `env -u GIT_DIR -u GIT_WORK_TREE PYTHONDONTWRITEBYTECODE=1 "$PY" -B "$SPEC/.temp/TASK-261021-yiyrus/check-draft-01.py` | 0 | Template/JSON/neutrality/links and link control; `draft-check-01.log`. |
| `env -u GIT_DIR -u GIT_WORK_TREE TMPDIR=/tmp/cw-dev1 GOCACHE="$WORK/.cache/go-build" go test ./tools/...` | 0 | generate-vectors package passed; `go-test-01.log`. |
| Final clone head, normal `.git`, `git status --porcelain` | 0 | Exact head, clean; `clone-final-01.log`. |
| `g push -u origin cip-mcp-in-project-skillfile-r2` | 0 | New branch published. |
| `g ls-remote --heads origin refs/heads/cip-mcp-in-project-skillfile refs/heads/cip-mcp-in-project-skillfile-r2` and SHA assertions | 0 | New remote head verified; original branch unchanged at `38aa3a0...`; `remote-head-01.log`. |
| Guarded `rm -r "$VAL"` and absence check | 0 | Validation clone deleted; `cleanup-01.log`. |
| Final source checkout status | 0 | Clean. |

Procedural retries: the first pre-push guard returned 1 because zsh expanded an
unquoted `%G?`; the quoted guard passed before publication. A force-form clone
cleanup was rejected by the tool before execution (no shell exit), then the
guarded non-force removal passed. The first report-generator attempt returned 1
with a syntax error before writing; report generation was corrected. These
retries changed no repository content.

Named logs and grouping helper scripts are retained under `LOG`, outside the
deleted validation clone; the reused Draft checker remains at its command path
above. `groups.json` records the exact IDs per group;
`group-summary.json` records completed test counts and durations. The interrupted
combined-run log is not used as passing suite evidence.

## Python groups and claims

Each row is one command with `N` as its group number. Full output is retained in
`python-group-NN.log`. These are the existing fixture assertions, not native
runtime qualification of this Draft.

| Group | Module | Tests | Time | Exit | Claim exercised |
| ---: | --- | ---: | ---: | ---: | --- |
| 1 | `test_allowed_signers` | 7 | 0.019s | 0 | Valid signer lines and malformed/unknown key rejection. |
| 2 | `test_implementation_coverage` | 39 | 1.390s | 0 | Inventory mappings and accepted/rejected stream parsing. |
| 3 | `test_local_hosted_launches` | 10 | 0.027s | 0 | Launch selection, conflicting flags and forbidden receiver probes. |
| 4 | `test_local_setup_token` | 17 | 17.233s | 0 | Local-source refusals, preservation and secret/branch control plants. |
| 5 | `test_login_readiness` | 9 | 0.536s | 0 | Metadata controls and malformed/unindexed-case rejection. |
| 6 | `test_managed_home_snapshots` | 16 | 1.399s | 0 | Snapshot schema, adapter/profile binding and dispatch. |
| 7 | `test_muse` | 11 | 0.864s | 0 | Fragment/vector controls and forbidden channels. |
| 8 | `test_muse_launch_argv` | 6 | 0.023s | 0 | Argv exclusions and typed overrides. |
| 9 | `test_muse_v4` | 7 | 2.333s | 0 | Accepted v4 cases and backend/schema rejection controls. |
| 10 | `test_network_contacts` | 8 | 0.040s | 0 | Blocking canary and narrowing-mutant controls. |
| 11 | `test_profile_codex_settings` | 8 | 2.705s | 0 | Closed settings enums and qualification assertions. |
| 12 | `test_rc7_carriers` | 18 | 11.890s | 0 | Invalid carrier plants and ambient/status-store invariants. |
| 13 | `test_release_gate` | 10 | 541.909s | 0 | Exact ID subset of artifact, pin, review and malformed-release assertions. |
| 14 | `test_release_gate` | 10 | 248.783s | 0 | Exact ID subset of artifact, pin, review and malformed-release assertions. |
| 15 | `test_release_gate` | 10 | 333.004s | 0 | Exact ID subset of artifact, pin, review and malformed-release assertions. |
| 16 | `test_release_gate` | 10 | 133.270s | 0 | Exact ID subset of artifact, pin, review and malformed-release assertions. |
| 17 | `test_skillfile_sources_independence` | 8 | 12.262s | 0 | Accepted candidate and citation violation controls. |
| 18 | `test_source_marker` | 14 | 23.551s | 0 | Ownership/refusal positive and negative fixtures. |
| 19 | `test_validate` | 100 | 123.519s | 0 | Exact ID subset of schema, vector, semantic and text-contract assertions. |
| 20 | `test_validate` | 100 | 8.176s | 0 | Exact ID subset of schema, vector, semantic and text-contract assertions. |
| 21 | `test_validate` | 100 | 200.390s | 0 | Exact ID subset of schema, vector, semantic and text-contract assertions. |
| 22 | `test_validate` | 100 | 18.301s | 0 | Exact ID subset of schema, vector, semantic and text-contract assertions. |
| 23 | `test_validate` | 100 | 1.671s | 0 | Exact ID subset of schema, vector, semantic and text-contract assertions. |
| 24 | `test_validate` | 74 | 61.028s | 0 | Exact ID subset of schema, vector, semantic and text-contract assertions. |
| 25 | `test_verify_release_commit` | 5 | 0.001s | 0 | Accepted verification and rejected signatures. |
| 26 | `test_verify_release_merge_policy` | 5 | 0.005s | 0 | Accepted and forbidden merge policies. |
| 27 | `test_workspace_trust` | 77 | 0.454s | 0 | Carrier and trust-transition valid/invalid fixtures. |

## Open decisions

No new decisions. The proposal's existing owner questions are unchanged.
Revision 1 was accepted as a Draft filing; this report does not claim a review
verdict for the new head.
