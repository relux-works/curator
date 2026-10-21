# TASK-261021-krcy1c results, revision 1

## Delivery

- Repository: curator-spec.
- Branch: `cip02-ready-for-acceptance`.
- Base: `origin/main` at `e41f96387287f5e76604627f9e63662f7d6d86d4`.
- Signed head: `752307c6a825c07256521e818fef25ba5ae78fc7`.
- Signature verification: `G`.
- Candidate status: **Draft**, with **Proposed for acceptance** disposition. Owner acceptance is still a separate review.
- Board status was not changed. No PR was opened and no default branch was pushed.

## Decisions integrated

The 2026-10-10 Operator input section is preserved byte-for-byte. Added the owner answers of 2026-10-18: option-A persistent checkout × profile × environment home named Werkstatt; approved repository overlay and external sparse worktree with trusted-code delivery first; Claude setup-token through its admitted credential channel; default-deny permissions programmed by the manager, agent changes only through grants, and no project permission import; memory off with explicit opt-in required; busy repair/switch/sync wait or refuse; checked destination D-PATH, with skills and CLIs scoped to the Werkstatt; inert foreign-candidate review. Incorporated the 2026-10-21 correction to Claude/Codex/Muse as the first wave, keeping OpenCode/Pi later in the design. Added individually approved inline MCP bound to the normalized SERVER digest and the accepted delegated hosted direction alongside exact-plan composition.

The three former disposition gaps are addressed: selected stable home/lease semantics; separately versioned optional per-environment certified delivery with typed refusal for unsupported required items; and a closed leaf-1 candidate for registration, approved context/whole-skill/inline-MCP items, approval/generation/composition, lease state, fragment members, destination transform, source-exclusion capability, diagnostics and bounded inventories. Nine named vector families pair positive controls, negative witnesses and future narrowing controls. Explicit deferrals distinguish the accepted broader operator model from the initial executable subset.

## Important compatibility correction

The old project proposal called its future carrier fragment v4. Published v4 now belongs to Muse, and v5 is allocated to the login/store and Codex publication. The candidate therefore freezes the separate project grammar and retains the historical fragment-v4 vector label, while requiring coordinated allocation of an unused outer carrier. It explicitly forbids adding project fields to published v4/v5 or inventing v6 allocation in this Draft. This closes the member grammar without falsely repurposing an existing schema. Actual tests confirmed that both published versions still reject the two proposed project members.

The MCP companion was read as inert text from `origin/cip-mcp-in-project-skillfile` at `38aa3a0c777400a448f3eb7e7b2fed57a85654e2`. Because this task's main-branch base has no numbered entry, the CIP refers to **MCP server declarations in the project Skillfile** by title and branch, as requested. Normalization preserves argument/scalar bytes, sorts the defined sets, expands URL shorthand and uses the normalized SERVER digest separately from manifest/declaration identity.

Delegated requests/records remain owned by curator-run. The candidate does not add private wire fields or claim that an existing consumer can launch a Werkstatt. Route understanding, home indirection and certified source exclusions must qualify first. Delegated profile-pin/fragment-digest provenance is distinguished from exact-plan snapshot provenance. The project-specific busy-Werkstatt decision does not promote CIP-0016's still-proposed general home conditions to accepted status.

## Files changed

- `cips/CIP-0002-project-context-in-managed-launches.md`: dated disposition/decisions, closed first-slice candidate, route binding, updated implementation/test plans and two remaining owner questions.
- `UNRESOLVED_QUESTIONS.md`: link the candidate and its remaining questions/publication gates.

No schema, vector, runtime, README or changelog file changed. The existing README and CIP index still correctly identify this as a Draft proposal; this candidate publishes no normative protocol change or release record.

## Remaining owner questions

1. Later replacement of profile root/system context versus additive-only context. Recommendation: additive-only in the first release; require a separately specified, approved replacement item under immutable ceilings before later opt-in.
2. Private-memory retention after explicit opt-in. Recommendation: private per-Werkstatt retention until explicit deletion; no publication, cross-Werkstatt reuse or automatic admission as knowledge. Default remains off, and leaf 1 accepts off only.

Outer-carrier/diagnostic publication, consumer support and real adapter/route qualification remain engineering gates, not reopened owner choices about homes, PATH, permissions or the first wave. Deferred grammar/surfaces are listed explicitly in the CIP.

## Commands and exit codes

Private machine paths and identity details are omitted. `SPEC_ROOT`/`SPEC_GIT_DIR` denote the local checkout and its separate Git directory. `PY` denotes the existing task-local validation venv interpreter reused from the preceding task. `TMPDIR` for the final checks was the task's `.temp/TASK-261021-krcy1c/tmp` directory.

| Command | Exit | What it establishes |
|---|---:|---|
| `a2a inbox` | 0 | Read the newest addressed orchestrator request in the required room/topic. |
| `task-board q 'get(TASK-261021-krcy1c) { status assignee }; get(TASK-261021-krcy1c) { full }'` | 0 | Verified assignment and sole precondition resource. |
| `task-board resource get TASK-261021-krcy1c W-CIP02-brief.md` | 0 | Materialized and read the complete brief. |
| `a2a post ... --topic cip02-acceptance --kind ack ...` | 0 | Task acknowledged after board verification. |
| `git --git-dir="$SPEC_GIT_DIR" --work-tree="$SPEC_ROOT" fetch origin` | 0 | Fresh remote base; provisioned agent Git environment sourced first. |
| `git --git-dir="$SPEC_GIT_DIR" --work-tree="$SPEC_ROOT" switch -c cip02-ready-for-acceptance origin/main` | 0 | Created the named branch from the fetched named base, with clean tracked state. |
| `GIT_DIR="$SPEC_GIT_DIR" GIT_WORK_TREE="$SPEC_ROOT" $PY -B tools/validate.py` (initial) | 1 | Local-link checker detected a scratch Markdown fragment's relative link under .temp. Renamed inert scratch reference fragments to .txt; no product checker change. |
| Same validator (after scratch cleanup and on final signed head) | 0 | Validated 74 schemas and 1294 vector files, released-record/schema bytes, corpus properties and local link targets. |
| `$PY -B .temp/TASK-261021-krcy1c/check-docs.py` (final signed head) | 0 | Nine new local links/anchors resolve; original owner input unchanged; Draft retained; exactly two documentation files changed; punctuation/private-identifier checks; v4/v5 positive/negative/narrowing controls; actual local-link positive/negative control. |
| `TMPDIR=<task-local tmp> GIT_DIR="$SPEC_GIT_DIR" GIT_WORK_TREE="$SPEC_ROOT" $PY -B -m unittest discover -s tools -p 'test_rc7_carriers.py' -v` | 0 | All 18 existing publication/property tests passed on the signed head. Individual claims below. |
| `git --git-dir="$SPEC_GIT_DIR" --work-tree="$SPEC_ROOT" diff --check` | 0 | No whitespace errors. |
| `git --git-dir="$SPEC_GIT_DIR" --work-tree="$SPEC_ROOT" commit -S -m 'cips: prepare project context candidate for acceptance'` | 0 | Created the reviewed two-file signed commit; provisioned agent Git environment sourced first. |
| `git --git-dir="$SPEC_GIT_DIR" --work-tree="$SPEC_ROOT" log -1 --format='%H %G?'` | 0 | Exact head above with signature status G. |

Publication: `git --git-dir="$SPEC_GIT_DIR" --work-tree="$SPEC_ROOT" push -u origin cip02-ready-for-acceptance` exited 0. The provisioned agent Git environment was sourced first. The remote-tracking feature ref matches the signed head.

## Each existing carrier test and the claim it proves

These are specification-data/publication assertions. None executes a provider or proves project-context implementation, native source exclusion, permission enforcement or runtime qualification.

| Test | Claim |
|---|---|
| `test_bounded_invalid_containing_carrier_plant` | The real corpus checker rejects a planted wrong enclosing discriminator beside a valid control, rather than failing setup. |
| `test_compatibility_dispatch_and_v4_ownership` | Old/new complete controls use their own readers; unknown additions/tokens cannot upgrade old envelopes. |
| `test_config_enclosing_fields_and_old_readers` | Old config readers reject the new selector; the successor validates its full envelope. |
| `test_cross_document_status_boundary_regression` | A widened consuming-document execution clause is detected even with the central prohibition retained. |
| `test_delivery_vectors_refuse_partial_or_dropped_transport` | Publication vectors require supported delivery or exact refusal, with no partial-success substitute. |
| `test_draft_reserved_versions_are_not_repurposed` | Other draft schema numbers are not silently reused by the independent publication. |
| `test_duplicate_keys_and_invalid_utf8` | Duplicate keys and invalid encoding fail before schema admission. |
| `test_full_carrier_required_members` | Every required successor enclosing field is enforced against a complete valid control. |
| `test_muse_and_other_adapter_boundaries` | V5 retains Muse's parent/value constraints and forbids Codex payloads on other adapters. |
| `test_muse_status_source_backend_regression` | Known home/shared Muse candidates require file metadata; external/unknown/conflict cases stay distinct. |
| `test_package_authority_and_empty_provenance` | Empty package settings remain valid, while package-derived store/account authority is rejected. |
| `test_production_corpus_and_canonical_bytes` | Canonical bytes, full carrier verdicts and checker dispatch agree on the published corpus. |
| `test_published_muse_decision_table` | The published Muse seam decision and CLI text retain their specified distinctions. |
| `test_renderer_transition_and_cli_published` | Native projections and nonempty refusals are explicit in published text; no unsupported downgrade is permitted there. |
| `test_single_typed_payload_required_and_agrees` | Native omits typed settings; explicit file/keyring requires one matching payload. |
| `test_status_enclosing_fields_and_secret_closure` | A valid imported row cannot validate a malformed wrapper; secret closure is enforced without mutating inputs. |
| `test_status_no_agent_execution_published` | Published status clauses prohibit implicit version/auth execution; this is a textual contract assertion. |
| `test_status_store_and_ambient_invariants` | The published matrix distinguishes configured/native/effective store and preserves unknown ambient OpenCode. |

## Bounded compatibility and checker controls

The task-local check loaded known-valid published v4 and v5 fixtures through the real JSON Schema implementation. Adding `project_context` or `command_environment` was rejected specifically by root `additionalProperties` in each schema, with no unrelated failure. Four in-memory narrowing controls changed that existing guard to allow extra properties; the same planted additions then validated. This verifies discrimination of the published boundary, not a new project schema implementation.

The actual repository local-link checker accepted an existing fixture target and rejected a planted missing target with the intended broken-local-link failure. No fixture or mutant changed tracked product files. The nine future project vector families and runtime narrowing controls are design requirements only; none is claimed executed or killed here.

## Evidence and limits

Final checks ran at signed head `752307c6a825c07256521e818fef25ba5ae78fc7`, with no later tracked edits. Evidence stays local under `.temp/TASK-261021-krcy1c/`: `validate-04.log`, `carriers-02.log`, `docs-03.log`, `check-docs.py`, `self-review.txt` and the inert MCP source extract. Earlier logs retain the initial scratch-link failure and preliminary runs.

Full Python/Go suites and runtime probes were not run: this task changes a Draft CIP and unresolved-question pointer, with no executable/schema/corpus delta. Therefore no full-suite disposable clone was needed. The tests above establish document/link and unchanged-carrier properties only. No credential, native runtime or hostile-code containment probe was performed. No worker or server outlives this turn.
