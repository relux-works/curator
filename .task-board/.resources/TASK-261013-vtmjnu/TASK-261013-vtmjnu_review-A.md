# TASK-261013-vtmjnu — apply-wave5-transfer-untrusted-cips: reviewer A

Record-only review of CR-TASK-261013-vtmjnu-1 revision 1. Reviewer B owns the decision. No acceptance/rejection or task-status mutation was performed; the explicit reviewer-A hard rule supersedes the generic lifecycle instructions. No repository code, documentation or LOGBOOK.md was changed.

## Review identity and result

Base: `6a3353df3e376e553de45e49d1f4818d85dd9148`.
Candidate tree: `33a5c6cdc99cb272c105b4ec7e5497a18def4c4e`.
Reviewed the exact 16-path delta, surrounding contract text, all changed figure sources, and TASK-261013-vtmjnu_gaps.md. Independently compared local bytes with candidate Git objects: 16/16 match, including the untracked transfer figure (blob `ad52e33df79a1a9df71e75462b72dc51f7cddc9f`). Ordinary git diff against a tree omits that untracked file, so byte comparison was used rather than interpreting its apparent deletion as a defect.

No actionable defect found in this bounded review. This is evidence for reviewer B, not an accepted verdict.

## Findings

```yaml
findings: []
```

No severity or P-level assigned because there is no defect finding. Evidence limits below are not claims of qualification.

## Swept surfaces

| Surface | Checked requirement and result |
|---|---|
| R-TRANSFER | CIP-0009 D5/D5a/D7, protocol transfer section: all eight stages, caller-owned B/archive/manifest/plan coverage, child-UID approved materialize/export, ignored worker Git history, account-wide absence and workplace fence, bounded all-or-nothing export, requester-only result, cleanup obligation/explicit waiver, deterministic import, specified hard/flagged paths, signed recomputable C, independent trusted exact-C review, no secret-bearing CI before acceptance, signed rework from C, intermediate export, signed fast-forward and moved-base re-review are present. |
| R-NET-TOOL | CIP-0008 R3 and inserted network paragraphs/question 6; CIP-0009 D0/D4/D5b: zero-egress workplaces with denied loopback/no proxy; no default search/fetch; plan-only provider-search control and injection rationale; signed read-only digested resources and letter requests; both qualified trusted-circle opt-ins, separate B identity/C sandbox, donor key custody/consent and search API-only egress. Host boundaries stay unchanged by circle admission. |
| R-EXEC / R-VERIFY | CIP-0009 D3/D5c and protocol: OS allowlist, non-executable writable areas, all excluded categories, qualified light checks and protected D-PATH wrapper, read-only caches, withdrawn agent-headless-build, all worker-code execution on secret-free ephemeral runners on another machine, controlled dependency egress, board testing-state data and tester orchestration only. |
| R-UNTRUSTED / R-STANDING | CIP-0009 D4/D5d; CIP-0010 boundary/C5; protocol: recorded admission class, default untrusted workplaces, slice-0/v0 refusal, no subscription binding/lease even for trusted-circle workplaces, donor-only standing workplace loop, trusted-only local standing loop, launch/restart/resume surface checks and explicit drift repair. |
| R-CEILINGS / R-RANGE-DENY / R-SOCKETS / R-SHARED | CIP-0009 D5d/D9 and protocol: qualified per-release matrix and class minimums, no OS-name exclusion, Linux selected containment/kernel baseline, macOS soft/absent gaps and open soft-row decision, static UID-range deny with active precedence probes, declared socket reachability, separate public ingress/no colocated shared services. |
| R-WATCH / R-ALLSTOP | CIP-0009 D7 and protocol: least-privileged per-UID sampler, one-account automatic journaled narrowing, phone alerts; manual operator node stop with admission hold, range deny, confirmed stops, broker freeze/vendor-token caveat, delivery closure/snapshots, journaled partial/unknown closure and explicit recovery. Existing suspended/channel-down behavior retained. |
| Publication and earlier names | CIP-0009 questions 13–17, CIP-0008 question 6, CIP-0010 question 6 and protocol publication register distinguish decided behavior from unpublished schemas, caps, records, commands, identity/profile and runner qualification. No new wire fields/refusal codes/numeric cap values found in the delta. R-WORKPLACE, R-OBS, R-OPERATOR and related existing controls retained. |
| Figures | Read all 11 changed/new figure sources: credential-source-lifecycle, donor-first-deployment, donor-layout, donor-lifecycle, donor-side DSL and PlantUML, donor-variants, join-handshake, remote-worker-retire, remote-worker-transfer, remote-worker-turn. They agree with the changed contract at their depicted level; the README links every figure, including the new transfer sequence. RB in the turn sequence is the project-side service, not the donor bridge; its export-request arrow is not a donor control path. |
| Gaps outcome / revisions / hygiene | Gaps outcome covers 12/12 wave-5 shared names with sections/figures and open items, plus retained earlier controls. Four changed specs and figure index/sources have revision notes. Reviewed public changes and scoped searches found no personal first names, private host names, local user absolute paths, session links or board-resource links. Generic product installation paths are not private machine paths. |

## Neighbor contracts and evidence limits

Compared the local neighboring dispatcher spec §5.2 workspace transfer, §5.4 cleanup and §16 question 22, and user-manager helper spec §7 workspace transfer, §8 networking/watcher/all-stop and §14 question 15. Dispatcher-owned materialization/export, child-UID execution, exact candidate review, original-caller result and export-before-deletion agree with this candidate. These neighboring local files are supporting context, not part of the reviewed candidate tree.

RWP v2.1 §8/§10 source was not found in the available local repositories. Gate coverage was checked against the explicit authoritative R-TRANSFER list supplied in gw5-decisions.md and the neighboring dispatcher/helper adoption. Direct fidelity to the original RWP document remains unverified; no remote source or runtime conformance is claimed.

Independently ran only read-only inspection, byte/link comparison and `git diff --check` on the exact base/tree (exit 0). Relative Markdown file targets in changed Markdown resolve; anchor semantics and rendered layout are not mechanically proven. No tests, builds, candidate code or renderer were run by this reviewer. Figure syntax/semantics were read as source; render success remains unknown.

Read the existing revision-1 validation resource: it reports 73 schemas, 1294 vector files, 689 tests, a compile-only Go step and exit 0; command-shard coverage 1/1, test-case coverage unknown. This is attached handoff evidence only, not independently rerun evidence or proof of wave-5 runtime enforcement. Draft contracts explicitly reserve future negative qualification.

`task-board spawn goal` reports no active run goal; directives reports none. No goal verdict was recorded. Findings are persisted here rather than in LOGBOOK.md, as the current brief forbids writing that file.
