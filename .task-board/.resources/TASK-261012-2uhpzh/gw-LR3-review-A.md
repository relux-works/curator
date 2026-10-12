# TASK-261012-2uhpzh — apply-review-findings-lr — reviewer A
Record-only review of CR-TASK-261012-2uhpzh-1 revision 1.
Base: 9ee6e38deb4a5d4afefe59c637dd65105d29a941
Candidate tree: 8b07b30a521220b16bf854abe2a692b1bddccb95

No accept_cr/reject_cr or status mutation performed. Reviewer B owns the decision.
Recommendation: correct the public hygiene residue before acceptance.

## Findings
```yaml
findings:
  - id: LR3-A-1
    severity: robustness
    severity_reason: "P2: explicit public-hygiene acceptance criterion remains unmet; inherited text is still present in the candidate."
    location: "README.md:536-538; CHANGELOG.md:157"
    evidence: "README names the specific optional CI host in its job label and host-derived repository-variable name. CHANGELOG repeats the host-specific job label."
    impact: "Public documentation still exposes a host name despite the brief's explicit no-host-names rule."
    requested_change: "Describe the optional self-hosted macOS ARM64 lane generically in public prose, without the machine-specific job/variable spelling; remove the matching changelog reference. Keep this docs-only; do not rename workflow configuration as part of this fix."
  - id: LR3-A-2
    severity: note
    severity_reason: "P3 informational integration dependency, not a new launcher implementation defect."
    location: "SPEC.md §§10.1, 10.2, 10.4, 10.5; neighbouring module specs"
    evidence: "The inspected local user-manager §7 still closes every descriptor above 2 and stages a pipe before exec; it also retains the ticket path. Local dispatcher retains command-mode transport and local session-host retains the ticket publication question."
    impact: "Cross-repository publication is not yet established by these local snapshots. The candidate correctly keeps activation gated and must not be treated as an activated joint contract."
    requested_change: "Carry this dependency to integration review and verify the corresponding module revisions before activation; no duplicated descriptor lifecycle should be added to the launcher."
```

## Swept surfaces
| Surface | Evidence / result |
| --- | --- |
| Brief item 1, F-6/R-FD3 | §10.1 references the single user-manager lifecycle; permits only caller-owned 0–2/read-only FD3; excludes internal handles. Bounded immutable staging preference and qualified pipe capacity/feeder/deadline/cleanup are present. 48/80 KiB bounds and intake/digest/framing/read-deadline gates remain; joint mechanism publication and failure/concurrency qualification are explicit. |
| Brief item 2, F-4/R-START | §§10.2/10.5, README and four figures remove ticket/redemption/host-role authority. host_start, dispatch-op/1 and the four permit-binding names match the resolution. Cancellation, ambiguous same-effect recovery, fresh host channel, epoch readiness and dispatcher-only starts agree. |
| Brief item 3, F-5 | §§10.1/10.2/10.4 and all four figures replace command/deferred-daemon text with an initial Unix-socket service. Existing 128 KiB service limit is retained; obsolete 96 KiB stdin limit removed. No command fallback; recovery/readiness gate admission. |
| F-1 | §10.4, README and figures specify home deletion, restricted metadata export and explicit retention without confidentiality assurance. No arbitrary-home classification alternative remains in the launcher target. |
| Shared names / wire restraint | R-FD3 and R-START are unchanged. R-OBS/R-VERDICT/R-DELEG/R-MAINT are not redefined by this scoped change. No added refusal code or invented wire field found in the exact delta. Existing receipt members and refusal table are unchanged. |
| Figures | 4/4 changed sources inspected: components, launch-paths, launch-states, artifact-layout. All linked from §10 and diagrams/README. Current interactive FD3/FD4 stays separate. Source matches target text; rendered syntax/layout remains unverified. |
| Revision / scope | Top-of-spec dated revision present. Exact delta has seven documentation files, no code or LOGBOOK changes. Working files for those paths matched the candidate tree at inspection. |
| Gap report | Four touched findings are mapped to sections/figures and open gates accurately. Its hygiene evidence is narrower than the whole public-hygiene AC: the inherited host references above remain. |
| Public hygiene | Full README/SPEC/figure scan plus broader Markdown search found LR3-A-1. No added personal paths, session links or board-resource links found. This is source inspection, not a universal personal-name classifier. |

## Evidence and limits
Reviewed the exact base-to-candidate diff, surrounding §10, README, figure index and all changed figure sources; inspected the attached TASK-261012-2uhpzh_gaps.md and CR validation log.
No tests, builds, renderer or validation commands were run by this reviewer.
Accepted only the attached log's report: configured remote gate exit 0, required shards 1/1 green; Ubuntu/macOS test and race jobs success, optional self-hosted lane skipped. Test-case coverage remains unknown. Hosted run/source binding was not independently re-attested.

Neighbour snapshots inspected (not assumed to represent parallel unpublished work):
- user-manager 889e141f25f086f1068d576aefa1f3593d6fc433: §§4.4.1, 7 and descriptor/launch references.
- dispatcher 5d0476dac90802b3e8c1d31d4b8c8ca5546ad6c8: §§3.2, 5.3–5.4, 7.2, 8.3 and frame bounds.
- session-host fb8c6ed6174432a2c0fa0ee65118e3488a906838: launch-channel/descriptor constraints and I4 publication question.

Run goal query returned no active goal. LOGBOOK and repository files were not modified.
