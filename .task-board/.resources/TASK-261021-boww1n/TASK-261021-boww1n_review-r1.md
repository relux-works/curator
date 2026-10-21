# TASK-261021-boww1n review r1

Exact head: `7f97b0939ddb604ad7bbcccc5871e28cac6fe92f`.
Base: `cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7`.
Repository: curator; branch: research/project-source-exclusion-codex.

The orchestrator request, board task, named reviewer and all three preconditions were verified. The producer outcome and CIP-0002 were read; the local CIP bytes match the recorded spec-input digest. The single introduced commit has signature `G`. Review used a detached exact-head worktree; the full default suite used a separate normal-.git exact-head clone. Fetched main has seven later commits, including a separate profile-dependency change; this exact-head review does not test their combination. There are seven changed files, all research/qualification material and a README entry. No product code changes.

## Surface sweep

Held means the research claim is supported, not that a complete strict adapter exists.

| Surface | Held or broken | Attack and evidence |
| --- | --- | --- |
| Binary intake | Held | Independently hashed the exact standalone binary: 74c6a6d263f40d65f99c27a815275f5b4ed0bcc2baecc396b4cf3255498bdfc0. Replayed missing-opt-in, wrong-digest and invalid-root controls. Wrong bytes fail at the digest check; refused entries create no child fixture. Default explicit research invocation skips all vendor entries. |
| Offline/isolation boundary | Held within stated scope | Reviewed env-i construction, disposable HOME/XDG/CODEX_HOME and Seatbelt denies before execution. Reran ordinary synthetic read, Keychains-shaped specific denial with byte preservation, and unsandboxed reachable-loopback versus sandbox connect denial. The allow-loopback mutant fails with curl exit 0 versus required 7. Real native-directory and system-policy deny rules remain present. An independent metadata-only stat probe is denied for all eight existing native directories (one configured path is absent); no directory listing or native file bytes were read. A nonexistent-child probe returned ENOENT and is not counted as deny evidence. No real credential or Keychain item was read. |
| P1 root/ancestor instructions | Held | Real debug prompt-input yields complete JSON. Home and repository markers load in positive controls; document budget 0 removes project/root override markers and preserves the approved home marker. Outside-Git-root and CLAUDE markers are absent in the recorded default. |
| P2 nested/lazy instructions | Held as scoped | Nested startup includes root and nested markers; cap 0 excludes both while approved home remains. Lazy tool entry is expressly unmeasured: a separate startup is not a lazy-entry witness. |
| P3 repository skills | Held: negative result reproduced | Project distrust plus document cap retains repository skill markers. Exact-path disables exclude known skills while approved home remains. A newly created skill path bypasses that list at next startup. Both .agents/skills and .codex/skills markers appeared in this replay's positive prompt dump; the record conservatively bases its blocker on .agents/skills. |
| P4 hooks/plugins | Held as scoped | hooks/list yields approved and project hooks when trusted, only approved under distrust. Installed local plugin skill enablement follows the project trust control while approved plugin skills remain. This is discovery/enablement evidence; hook execution and other plugin entry points remain unmeasured. |
| P5 MCP empty/nonempty | Held | Replayed all four trust-by-empty-set combinations through real stdio app-server. Approved process starts and tool call returns its marker in every case. Project process/call occur only for trusted/nonempty; excluded cases have no project process receipt. No remote transport or inference. |
| P6 permissions/settings | Held as scoped | Compared config/read and actual thread/start policies: trusted project on-request/workspaceWrite/high; untrusted approved never/readOnly/low. Home remains loaded. Other settings, execution-policy rules and malformed inputs are not claimed. |
| P7 resume/children | Held as explicitly unmeasured | Replayed real start/inject/unsubscribe/resume at budgets 32768 and 0. Approved home instruction source remains. The project source is absent even from the uncapped instrument baseline, so this cannot prove exclusion. No model turn or child agent was launched; tainted model-history behavior remains unmeasured. |
| P8 late creation | Held: negative result reproduced | Within one running app-server, explicit skills/list forceReload discovers a new project skill despite distrust and retains approved skill. Reloaded config excludes late project MCP and preserves approved entry. Automatic in-turn reload and late instructions remain unmeasured. |
| Additional external sparse-worktree row | Held for measured startup/reload; late-write residual remains | Independent synthetic Git repository contains committed skill plants in both native skill roots. A sibling external worktree has real skip-worktree entries and included src/main.txt, with native roots unmaterialized. Original-checkout skill files remain present. Complete prompt dump and explicit app-server skill reload exclude both source markers while preserving approved home skill. Materializing both native roots restores their markers. Creating a new skill in the sparse view makes it appear on same-server reload. See additional row below. |
| Narrowing controls | Held | All four unchanged mutants exit 1 with their intended diagnostics: document cap, skill disable, project trust and loopback. Wrapper exit 0; no unrelated setup error is counted as a killed mutant. |
| Evidence integrity/free hunt | Held | Independently compared all three executable source hashes, exact prompt marker sets, config/thread-policy values, hook presence, MCP start/call pairs, installed plugin markers and clean API waits against committed observations. CIP digest matches local source. |
| Hygiene and scope | Held | Full added-byte scan and manual diff inspection found no personal paths, host/user identities, client/employer names or secrets. System executable locators are intentional fixture commands. Changed-document local links and git diff --check pass. |

## Additional sparse-worktree result

The producer harness has separate Git metadata but its standard project fixture is fully materialized; it has no external sparse-worktree case. I copied the unchanged helper into a task-local, non-repository probe module and added one independent test. No reviewed repository file was edited. Supplemental outcome `TASK-261021-boww1n_sparse-probe-r1.go` contains the added test. To replay, place it beside the exact-head replay_test.go in a disposable Go module, set the same binary/run-root gates and run `go test -count=1 -v -run ^TestReviewerExternalSparse$`. It uses the unchanged helper and creates only synthetic repositories.

| Layout/control | Observation | Qualification limit |
| --- | --- | --- |
| Original synthetic checkout with .agents/skills and .codex/skills | Both project skill markers and approved home marker load | Positive source-discovery control |
| External sibling sparse worktree, only src/ materialized; document cap 0 and explicit distrust | Both unmaterialized project markers absent from valid startup prompt and explicit skill reload; approved home skill present | Codex does not recover these skill bytes from the Git common directory or original checkout in these measured entry points |
| Same sparse view, new untracked .agents/skills entry created during app-server lifetime | Late marker appears on forceReload; approved home skill remains | Sparse selection alone is not a write/discovery boundary for later-created native roots |
| Same view, tracked native skill roots materialized | Both original project markers return | Discriminating materialization positive control |

This result narrows the retained-skills finding to discovery-visible files. It does not establish a protected overlay, enforce writes, qualify automatic reload/lazy/resume/children, or certify CIP-0002 strict entry. The complete adapter remains unqualified, consistent with the record.

The first additional-probe attempt used an unsupported sparse-checkout add flag. The second exposed an empty index left by --no-checkout; its absence observation is not counted. The corrected run verifies included source bytes and skip-worktree index flags, then passes all controls. A subsequent final run also covers the second native skill root. These are reviewer fixture corrections, not product defects.

## Execution evidence

Private evidence is under `.temp/TASK-261021-boww1n/` in the reviewer workspace. Vendor commands use the harness's env-i sandbox. Go invocations remove GIT_DIR/GIT_WORK_TREE and use the short workspace-local `.t` TMPDIR/GOTMPDIR. Default-suite HOME and Go caches are isolated; PATH uses Go 1.25.5, system tools and the matching Rust 1.91.0 toolchain without installed optional Curator providers. Platform is macOS arm64.

| Check | Exit | Evidence |
| --- | ---: | --- |
| Default explicit research invocation | 0 | default-research-01.log: vendor entries skip before fixture creation |
| Exact-head gated research matrix | 0 | gated-research-01.log: all assertions pass, 41.665s |
| Four narrowing mutants | 0 wrapper; 1 each intended child | mutants-01.log and mutants-01/results.tsv: all KILLED |
| Independent final external sparse probe | 0 | sparse-04.log, 8.382s; sparse-04/TestReviewerExternalSparse/evidence/ |
| Observation/source-hash comparison | 0 | observation-compare-01.log |
| go build -p=2 ./... | 0 | build-01.log and adjacent exit file |
| go vet -p=2 ./... ./.research/project-exclusion-codex | 0 | vet-01.log and adjacent exit file |
| Native-directory metadata controls | 0 | native-boundary-02.json: all eight existing configured native paths denied; one path absent |
| Targeted MCP golden rerun with installed node/npx on isolated PATH | 0 | mcp-golden-02.log; exact unchanged TestProfileUpdateMCPDeltaConfirmationGolden |
| Full default suite, normal .git clone: go test -json -p=1 -count=1 -timeout=60m ./... | 1 | full-suite-01.jsonl: all 85 expected packages reached terminal results; 78 pass, 2 fail, 5 skip |
| Complete ciworkflow package rerun with installed node/npx on isolated PATH | 0 | ciworkflow-02.log, 59.460s |
| Full-run/replacement coverage audit | 0 | final-validation-audit.json: no missing package; both targeted replacement invocations exit 0 |

The initial full suite had exactly three failing top-level tests: TestProfileUpdateMCPDeltaConfirmationGolden (missing npx changes its golden diagnostic), TestLinuxRoutingKeepsForksHosted and TestLinuxRoutingRejectsRacePlants (node absent from PATH). The unchanged MCP test and the entire ciworkflow package pass after adding the installed Node tools to the isolated PATH. This is full-suite coverage plus successful targeted replacements, not an uninterrupted green full-suite invocation. Platform and optional-integration skips do not qualify those integrations. No product modification was needed.

Both validation checkouts remained clean and at the requested head. The vendor fixture process audit found none of the 90 recorded process IDs still alive. Raw fixture evidence is retained privately; the attached supplemental probe contains only synthetic material.

## Verdict

Accepted as research and accurately bounded qualification evidence. No blocking or nonblocking findings. This verdict does not certify a complete strict-context adapter; repository skill discovery and the explicitly unmeasured surfaces remain as documented.

```json
{"verdict":"accepted","head":"7f97b0939ddb604ad7bbcccc5871e28cac6fe92f","findings":[]}
```
