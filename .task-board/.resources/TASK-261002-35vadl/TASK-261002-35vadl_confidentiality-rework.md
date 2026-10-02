# TASK-261002-35vadl — oss-audit-wave-1 confidentiality rework

Ready for review. Earlier repository verdicts and per-check counts are unchanged: five FIX-HEAD, one NEEDS-NEW-REPO, one HOLD-pending-decision. Prior audit and independent-review evidence are reused; repository history was not re-audited during this confidentiality-only revision.

Local report: `~/oss-audit/wave1/REPORT.md` (0600; directory 0700). Its previous bytes are preserved; one sanitisation note was appended.

| Shared-artifact check | Count |
|---|---:|
| Board resources successfully re-read including this outcome | 10/10 |
| Preconditions / outcomes | 3/7 |
| Included spawn logs / verdicts / results / patches | 3/1/1/1 |
| Resources sanitised through supported resource updates | 3 |
| Restricted-affiliation / private-host literals redacted | 21/16 |
| Read errors | 0 |
| Remaining email / personal-path / employer / affiliation candidates | 0/0/0/0 |
| Remaining host / private-address / session-link candidates | 0/0/0 |
| Secret alerts | 0 |
| Task metadata candidates | 0 |
| Live transcript snapshots checked | 2/2 |
| Live transcript sensitive candidates / secret alerts | 0/0 |
| Repository files changed | 0 |

| Validation | Exit | Measured result |
|---|---:|---|
| Initial shared_artifact_gate.py scan | 1 | 9/9 read; 37 candidates; expected failure on residual literals |
| Sanitized local copies scan | 0 | 9/9 read; zero candidates |
| Re-fetched board resources scan | 0 | 10/10 read; zero candidates |
| gitleaks dir: three board snapshots and one live-transcript snapshot | 0, 0, 0, 0 | Zero alerts in each pass |
| test_shared_artifact_gate.py | 0 | 6/6 tests passed |
| Narrowing mutant: skip spawn logs while scanning other resources | 1 | 1/1 named regression failed as expected; mutant rejected |
| Initial live transcript sanitisation/recheck attempts | 1, 1, 1 | Failed during concurrent command annotations; no clean result claimed |
| Live transcript sanitisation/recheck from neutral invocation directory | 0 | 2/2 checked; zero remaining candidates |
| Service-state verification with pipefail | 0 | Running |
| Report preservation and permission check | 0 | Original bytes preserved; one note; 0600/0700 |

Named regression: `SharedArtifactRegression.test_rejects_sensitive_spawn_log_when_results_clean`. It launches the real `shared_artifact_gate.py scan` entry point with a clean result and a contaminated log. The narrowing mutant changes the resource scan call inside `scan`, preserving result scanning while bypassing log scanning. Additional cases cover redacted inputs, email-bearing logs, absent files, malformed UTF-8 and empty inventory. Failed reads return exit 2, never a clean result.

Tool versions: gitleaks 8.30.1; Python 3.14.6; task-board dev. All fresh secret scans used `gitleaks dir` with `--redact --no-banner --report-format json`; reports remained local.

Coverage bound: the stated counts cover the enumerated UTF-8 board resource bytes and task metadata at the recorded reads. Literal patterns come from the project naming gate and task scope. Undisclosed client names, arbitrary encoded payloads and bytes appended after the live-run checkpoint are not established by this check. Earlier audit access/provenance gaps remain as recorded in the existing results, including the inaccessible MIT-only target. The active run transcript is sanitised and checked separately before handoff; its post-session automatic attachment is for the reviewer to recheck.
