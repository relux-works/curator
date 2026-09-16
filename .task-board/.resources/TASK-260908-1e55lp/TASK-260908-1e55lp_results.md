# B7 producer evidence — TASK-260908-1e55lp

## Candidate and scope
Base HEAD: 3535d63ea80f97bba2fcb6e1f06996cfc25cf7df. Story worktree: curator-spec/.temp/STORY-260908-2haegq/worktree. Candidate remains uncommitted. Exactly four changed files: decisions/0014-tool-configuration-surfaces.md, decisions/0015-per-project-policy.md, decisions/0016-managed-home-command-roots.md, UNRESOLVED_QUESTIONS.md.

Preserved drafts and index links already present at run entry, checked their contents, and added numbering provenance to 0014. Latest tracked decision is 0013; 0011 remains reserved. All three documents explicitly say Status: proposed — not adopted; no option selected. Filing authorizes no normative change, workaround, adoption, or landing. Draft content is also attached as three task-scoped outcome resources.

## Inputs and inspection bounds
Read epic resources agents-infra-to-curator-mapping.md (2026-09-07 inventory) and goal-launcher-and-infra-migration.md, exact workstream B7, through task-board resource get. Cross-checked environments §§1, 1.4, 7.4, 7.8, 9.2–9.4, 10.2–10.3, 11, 12.1–12.2, Decision 0012 OQ6 and Decision 0013 launch ownership/defaults. Current code inspection: curator internal/envfragment/envfragment.go Fragment has no path_prepend; internal/envprofile/managed.go claudeSeed and buildFragment match the limited seed and fragment descriptions. Launcher SPEC defaults remain launcher-owned. Skillfile source extensions remain skill-scoped. These are inspection bounds, not live tool/platform or production behavior proof. Historical mapping is explicitly not a current-runtime claim. No implementation gate changed; no runtime mutation or cross-platform coverage claimed.

## Validation run directly in this producer session
Shell: /bin/zsh. Commands were standalone, without tee or pipelines.
- PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make validate: exit 0. Validated 60 schemas and 1047 vector files; Python unittest ran 227 tests in 611.851s, OK; go test ./tools/... passed (generate-vectors, 12.941s). Process remained attached through bounded sequential polling until exit; nothing backgrounded across turn end.
- git diff --check: exit 0.
- Inline python3 document/scope assertion: exit 0; 3/3 proposals contain required status/context/gap/options/open-question/source citations, 3/3 index links present, exact four-file changed set. This is a structural check, not a semantic proof or newly shipped test.
No prior-run validation accepted as passing evidence. The runtime-owned landing suite was not run manually; make validate was run as specifically required by the B7 brief.

## Candidate SHA-256
- decisions/0014-tool-configuration-surfaces.md: 72774f8dc1da009984eabe06e1d5c606f819aa57526e0f94d24e12f81a5d9bca
- decisions/0015-per-project-policy.md: 38d2e8266b67f3d20302e87fe29c282961e5a1e57672d335f3c8d3420a031302
- decisions/0016-managed-home-command-roots.md: eda5d4c6c37c2fcfabd4cf7e72ae840a85b05c54436ac858e011f31523a69dbc
- UNRESOLVED_QUESTIONS.md: d70242b86d4d61fa91edba1c53231da86dec9e24de268fd5beb4da904e0a5767

## Handoff bounds
Independent reviewer acceptance remains pending and its checklist item is unchecked. The orchestrator owns review routing and closure. No commit, PR, merge, tag, deployment, or installation performed. B7 specifically excludes draft landing. LOGBOOK.md edits are prohibited by the campaign; the continuation finding is recorded in board notes and this outcome instead, and the literal logbook checkbox remains unchecked.
