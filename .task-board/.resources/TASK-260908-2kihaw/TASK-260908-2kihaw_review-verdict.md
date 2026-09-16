# Review verdict — TASK-260908-2kihaw b2-skill-packages-delivery

Reviewer: independent (Claude claude-fable-5-1, RUN-260915-fad81d), 2026-09-15, host e11-1.
Shell: zsh, `set -o pipefail` where piped. Read-only clones under the Story worktree `.temp/b2-review/`.

## Verdict: ACCEPT (all three PRs)

| Repo | PR | Exact head reviewed | Base | Verdict |
|---|---|---|---|---|
| relux-works/skill-pdf | #1 (b2/manifest) | 6d2392a861cdf389c9aa0245c46e8207975a2bcd | 9e3e72a1509823998ab08586b6c981689f2245f2 | ACCEPT |
| relux-works/skill-creator | #1 (b2/manifest) | ea8fd665486233ccaccce2d8508ea8caa9378bcd | 0a49a505096831a07d6971d1a5120ea5ba8a432e | ACCEPT |
| relux-works/skill-agents-attachments | #1 (b2/manifest) | 240f0292a6484a743ede98fb9af0097f4a875480 | a8156f1d7df66ad4fc7e50e881a9a5d5bfacd911 | ACCEPT |

Heads confirmed with `gh pr view 1 -R relux-works/<repo> --json headRefOid` and by
`git rev-parse HEAD` in the checked-out clones (all three match the brief; a later push voids this review).
Landing (fast-forward of these exact heads to main) and the three signed `v0.1.0` tags remain the orchestrator's job.

## 1. Manifest validation (`curator main-04550e2`)

| Command | Exit |
|---|---|
| `curator skill check skill-pdf` | 0 (advisory `skill.command_resolution_contract_missing` only) |
| `curator skill check skill-creator` | 0 (same advisory) |
| `curator skill check skill-agents-attachments` | 0 (same advisory) |

Manifest surfaces match the evidence: pdf = 1 `script` command (`render-pdf` → `scripts/render-pdf.sh`), exec pandoc/weasyprint, 2 system dependencies;
skill-creator = 3 `script` commands (init-skill/package-skill/validate-skill), exec git/zip;
agents-attachments = 1 `build` command, driver `go-v1`, `source_dir` = `cmd/agents-attachments` under a single `build_roots` entry, no `modules`, exec convert/magick/sips, env_read the three documented vars, `dependencies: {}`. All schema_version 8.

## 2. Gates and tests rerun by the reviewer

| Command | Exit | Notes |
|---|---|---|
| `make test` in skill-creator | 0 | 10 GATE PASS lines incl. N1/N2/B1 killed; 5 unittest OK |
| `make test` in skill-agents-attachments | 0 | 9 GATE PASS lines incl. N1/N2/B1 killed; `go vet` clean; `go test ./...` ok |
| `go test -count=1 ./...` in cmd/agents-attachments | 0 | internal/attachments ok; `gofmt -l .` empty |
| `make test` in skill-pdf | 2 | `GATE FAIL: required toolchain binary missing: pandoc` — see bound B1 |
| pdf N1 (drop `capabilities`) via `curator skill check` on a temp copy | nonzero | killed: `schema v8 requires 'capabilities'` |
| pdf N2 (dangling `unix_path`) via `curator skill check` on a temp copy | nonzero | killed: `source file not found` |

## 3. Reviewer narrowing mutants (all via `curator skill check` on temp copies, gate kept in place)

Killed (gate bites): unknown top-level key (all 3 repos); pdf absolute `unix_path` and `../` traversal; attachments `source_dir` moved to an existing dir outside `build_roots`, `build_roots` emptied, `build_roots=["."]`, driver `go-v2`.

Survived (recorded blind spots of the installed static checker, not defects of the candidate manifests): `schema_version: 7` (all 3); `capabilities.exec: []` with dependencies kept (all 3); pdf `unix_path` retargeted to existing non-script `README.md`; skill-creator two commands pointing at the same script and `unix_path` → non-executable `references/workflows.md`; attachments `env_read` removed. These belong to the manager's check semantics; the repos' own gates correctly place behavioral verdicts in the behavioral layer (B1), which I confirmed for skill-creator and agents-attachments by execution.

## 4. Byte fidelity (sha256 recomputed by the reviewer)

Bound B2: `dee5403` does not exist in the local relux-agents-infra clone after `git fetch origin` (all refs) and GitHub returns `No commit found for SHA: dee5403`. It was evidently an unpublished commit on the producer's original machine. Comparison was therefore done against `relux-agents-infra` `origin/main` = 459742ea67e3c6b84169520b92d74fe7f73e3002, whose file hashes equal the prefixes recorded in the producer evidence for dee5403.

- pdf: SKILL.md 02be9010…, assets/template.html5 9ae6d627…, assets/themes/prose-classic.css e7a8f7f3…, assets/themes/report-clean.css 193149e0…, scripts/render-pdf.sh 29a7b5ae… — all SAME; modes 100644/100755 preserved (render-pdf.sh 100755).
- skill-creator: SKILL.md 93b5880f24b5ffd05995a13413d37dedf83330b2f0d130b0ba68fa4031722811 SAME as `.skills/skill-creator/SKILL.md`; the standalone tree carries more files than the agents-infra copy (successor reuse per STORY-260903-2pxzvb), scripts 100755.
- agents-attachments: internal/attachments/attachments.go 85abb9fa… SAME, attachments_test.go f1cf5946… SAME versus `tools/agents-infra/internal/attachments/`. `.instructions/INSTRUCTIONS_ATTACHMENTS.md` 73ddf388… equals the SKILL.md body except for one blank separator line directly after the YAML frontmatter (cosmetic; noted, not blocking). LICENSE bytes identical to agents-infra LICENSE (31d2bdde…).

## 5. Umbrella range contract

`TASK-260908-2kihaw_umbrella-range-contract.py` hardcodes `/Users/iv/Developer/ReluxWorks/curator-spec`; run with the path substituted to `/Users/administrator/Developer/ReluxWorks/curator/curator-spec` (checkout at 3535d63). Result: 24/24 PASS, exit 0. README `requires.skills` entries (`^0.1`, git identities) are consistent across the three repos and admit v0.1.0.

## 6. Signatures and scope

All six commits (2 per repo) verify `Good "git" signature for ivan@relux.works with ED25519 key SHA256:Ng99XGF2pboYgFVfWJhYI2JRi0PyYsV9UwsJ70NBYd0` (allowed_signers built from /Users/administrator/.ssh/ivan-relux.pub); author Ivan Oparin <ivan@relux.works>; every subject ends `[skip ci]`. PR diffs contain only manifest, gate, README/Makefile wiring and gate fixtures (pdf 4 files; skill-creator 6; agents-attachments 6). All three repos private, default branch main, no reviews yet, one producer comment each recording the amended head.

## 7. Bounds (unverified here, stated, not inferred as passing)

- B1: pdf behavioral gate layer (render/reject/B1) not rerun: pandoc, weasyprint, pdftotext absent on this host and installs are barred. Static layer and N1/N2 reproduced manually. The producer's transcript claim for that layer is accepted as producer evidence only.
- B2: fidelity reference `dee5403` unreachable; compared against agents-infra origin/main (hashes equal the evidence).
- B3: shellcheck not installed here; shellcheck cleanliness not independently verified. The gate scripts were read in full and use POSIX sh with `set -eu`, temp dirs and traps.
- B4: live range resolution against real tags happens at B1 umbrella install time (as the evidence states).

## 8. Required follow-ups for the orchestrator (not blocking acceptance)

- Fast-forward the exact heads above to main; create signed `v0.1.0` tags on the landed commits.
- Optional hygiene for a later revision: parameterize the curator-spec path in the range contract script; drop the extra blank line in the agents-attachments SKILL.md body if strict byte-equality of the body is wanted.
