# Review verdict: CIP-0005 audit backends and CLI secret transport

- Task: TASK-261004-hy8zmn \u2014 audit-token-argv-and-backend-env-allowlist-design
- Parent: STORY-261004-2b8pnx \u2014 design-audit-backends-and-secret-transport
- Date: 2026-10-05
- Change Request: revision 1; candidate tree 78737c2cc008e8d227e3fb9a8e288c6422427713; base 9bc8e1a1eace93377e41c36465b26a58ee5ce05a.
- Verdict: accepted as a research deliverable. Operator design acceptance and prioritization remain pending. No implementation, producer scheduling, specification adoption or integration is authorized by this review.

## Swept review surfaces

| Surface | Finding |
| --- | --- |
| Brief and scope | All required subjects are addressed: argv refusal, env/file precedence, handle privacy/no-follow/64 KiB bounds, Codex/Claude/command selection, common child allowlist and mandatory denials, Git separation, canary, request cap, egress, strict/advisory failure and unsupported backend. |
| Decision readiness | Three real options, explicit recommendation B, descriptor/config defaults and precedence, eight decision-ready operator questions with recommended answers, normative sketch, versioned schema leaves and ordered implementation/test leaves. |
| Architecture | Native flag construction stays with agents-management under adopted Decision 0019; Curator owns execution/security policy. Profile strictness and its separate null-config integration seam are explicitly covered. Native compatibility is unestablished and remains a later qualification gate. |
| Evidence | Pinned production/spec citations checked below. P1-P5 are retained producer measurements with commands, synthetic fixture/environment description, expected exits and reproduction outline; they are not new reviewer executions. Infrastructure interruptions are recorded as failures, not green checks. |
| Test design | Real process observation, production-entry integration paths, negative vectors and narrowing mutants are specified. Native Windows/Linux/macOS qualifications remain future work. No unexecuted security gate is claimed green. |
| Safety | Documents contain public references and synthetic fixtures; manual inspection found no real credential value, internal hostname, personal filesystem path or private product material. No account/credential state was accessed by this reviewer. Historical no-secret/no-login claims are producer attestations, not independently observable proof of all past actions. |
| Diff hygiene | Exactly two .research paths in the candidate delta; each worktree file byte-matches the candidate blob. Tracked product/dependency/LOGBOOK diff is empty. No reviewer repository edits or commits. |

## Citation spot-checks: 16/16 supported

Read with `git show REV:PATH` and line numbering at curator ca1b776fb580ec0cee0173bf150daf063023aeaa and curator-spec 43bf0a2506d5c354a73bbc3ea4623d4653db10c7 (peeled rc.14 tag independently matched).

| # | Pinned citation | Verified claim |
| --- | --- | --- |
| 1 | cmd/curator/main.go:2120 | Literal token flag registered. |
| 2 | cmd/curator/main.go:2147-2160 | Argument/env precedence and record read precede Publish. |
| 3 | internal/config/config.go:129-145 | Raw backend configuration retained. |
| 4 | internal/config/config.go:963-1011 | Backend, cloud, request-cap and opaque descriptor parsing. |
| 5 | internal/config/config.go:42-43 | 1 MiB default / 10 MiB maximum. |
| 6 | internal/audit/audit.go:307-325 | Cache or deterministic detection; no analyzer dispatch in this path. |
| 7 | internal/audit/audit.go:413-429 | Backend/model cache filename; read/decode errors become misses. |
| 8 | internal/audit/audit.go:470-480 | Static findings receive configured backend label, without hash_version. |
| 9 | internal/envprofile/envprofile.go:1716-1732 | Separate strict config hardcodes null and calls Gate. |
| 10 | internal/registry/http.go:438-461 | Record validation precedes HTTP; token used in Bearer header. |
| 11 | internal/hashing/hashing.go:126-165 | Explicit versioned identity helper exists. |
| 12 | .github/workflows/ci.yml:37-43 | CI pins the stated rc.14 commit. |
| 13 | profiles/manager.md:1106-1120 | Static canary, configured analysis, public-only cloud/redaction and mode failures. |
| 14 | profiles/manager.md:1122-1129 | Version-2 state/pins and no revocation override. |
| 15 | profiles/manager.md:2977-3004 | Strict profile closure audit and unpinnable secret findings. |
| 16 | decisions/0019-fragment-consumers-and-one-construction-site.md:64-106 | Shared native harness construction ownership. |

The citation baseline is intentionally the producer's pinned research revision, not an assertion that today's public main is unchanged. No current-main freshness claim is made.

## Independent bounded checks

- `git diff --check BASE CANDIDATE`: exit 0.
- `git diff --exit-code HEAD -- LOGBOOK.md cmd internal go.mod go.sum`: exit 0.
- Candidate/working-copy equality: 2/2; documents total 79,228 bytes, below 96 KiB.
- Independent document assertions: 24/24, exit 0. Reproduction: assert the twelve template headings (Summary, Motivation and user stories, Current state, Design, Options considered, Recommendation, Security considerations, Compatibility and migration, Specification changes, Implementation plan, Test plan, Open questions for the operator); for each of two files assert terminal newline, no trailing whitespace, no literal private-path/host markers, titled task/story metadata and existing relative companion links; then assert the combined byte bound and exact two-path candidate delta. Marker checks were limited to personal home path prefixes, email domain marker and local-host-with-port suffix; manual inspection supplements this narrow check. This measures document packaging, not secret detector completeness or runtime security.
- Unauthenticated public PR metadata independently reports merged=true and head 0da153a54e7f8b446e63a6a177a9b5392e30da11. Independently retrieved environment.py and registry_token.py match 2/2 recorded byte counts/SHA-256 digests. Source confirms deny-after-merge, Windows name normalization and the token reader's Windows privacy limitation. No reference test suite rerun.
- [Claude CLI reference](https://code.claude.com/docs/en/cli-reference) and [Codex exec reference](https://learn.chatgpt.com/docs/developer-commands?surface=cli) were reopened. They support individual CLI building blocks; the proposal correctly requires separate qualification of the complete native profile.

Hosted-evidence mode: no local go build/go test, provider invocation, login/logout or native platform security probe. Prior scratch build/P1-P5 outcomes are accepted as explicitly attributed research observations, not new green tests. Future gate tests remain proposed.

## Checklist interpretation and disposition

Research acceptance criteria are satisfied; architecture fit is supported. Tests-green is satisfied only for applicable document checks; product suites are not applicable to this two-document research delta and were not rerun. Conditional logbook and changes-requested items are not applicable to this accepted research review; LOGBOOK.md remains untouched. There are no revision requests. Record acceptance via accept_cr with this outcome; route to integrating, with checkpoint/integration left to the authorized researcher/analyst producer run. The run-goal query reported no active goal (not goal-bound).
