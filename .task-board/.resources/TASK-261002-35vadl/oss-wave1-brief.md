# THE ONLY CURRENT INSTRUCTION — TASK-261002-35vadl: open-source pre-release audit, wave 1 (researcher, READ-ONLY)

## Confidentiality (binding, highest priority)
SENSITIVE findings must NEVER appear in this task's board results, in any board resource, in a Change Request, in any commit, or in any a2a room. Sensitive means:
- secret values;
- employer references;
- <restricted-affiliation> or client references;
- personal paths, hosts or addresses;
- email addresses.

Write them ONLY to the local report `~/oss-audit/wave1/REPORT.md` (create ~/oss-audit/wave1 with mode 0700; files 0600). Redact secret values even there (first 4 characters plus length).

The board results carry ONLY: per-repo verdicts, per-check COUNTS, and the local report path. Do not create any `.research/` file in the repository for this task.

## Scope (FULL history: every branch, tag and stash on the remote, every commit's tree, commit messages, author/committer identities)
- relux-works/skill-agents-attachments
- relux-works/skill-pdf
- relux-works/skill-youtube-scraper
- relux-works/skill-git-filter-repo
- relux-works/skill-youtrack (archived)
- relux-works/relux-codes-landing (also report on its public contact address)
- ivanopcode/swift-realtime-openai-local: ONLY its upstream MIT attribution check (Ivan has not decided whether to keep it). Mark it HOLD-pending-decision.

Mirror each repo into `~/oss-audit/mirrors/<name>.git` with `git clone --mirror` (read-only; never push), via gh/ssh with the existing credentials.

## Checks per repository
1. **Secrets.** Run `~/.local/bin/gitleaks git --log-opts=--all --redact` over the mirror (or the version-appropriate equivalent; record the command and version). Also look for sensitive file NAMES anywhere in history: .env, .p8, .pem, private keys, keystores, credentials.json.
2. **Personal data.**
   - /Users/ and /home/ paths;
   - host names: ivmbp, macbook-iv, rose-air, the mini's host name, *.ts.net, *.local;
   - LAN/tailnet addresses: 10/8, 172.16/12, 192.168/16, 100.64/10;
   - claude.ai/code/session links.

   Search all commits' trees and the commit messages: `git log --all -p` / `git grep` over `git rev-list --all`, deduplicated by blob.
3. **Employer, <restricted-affiliation> and client references.** Build the employer patterns the same way `.github/ci/naming-gate.sh` in the curator repo does, assembling the names at runtime. Never write either name into any file except the local report, and redact it there as "<employer>". Search for <restricted-affiliation> and client names likewise.
4. **Identities.** Every author and committer identity per repo with commit counts, for the consent record. Also any email address in content.
5. **Licences.** LICENSE, NOTICE and SPDX headers present and consistent: Apache-2.0 expected; vendor licence files kept; holders named.
6. **Runners.** Any `self-hosted` in `runs-on` on ANY ref (R120).

## Output
In REPORT.md, per repo: a verdict, the counts per check, and locations as repo:ref:path:line or SHA, with values redacted. Verdicts:
- CLEAN;
- FIX-HEAD (only HEAD content needs a fix);
- NEEDS-NEW-REPO (history contains sensitive data);
- HOLD.

In the board results, only:
- the verdicts table with counts;
- the REPORT path;
- the tool versions;
- what could not be checked.

Rules:
- Read-only. Never push, never modify remotes, no issues or PRs.
- Builds only through ~/.mini-build-lock. None should be needed.
- The host has syspolicyd exec stalls: check `launchctl print system/com.apple.security.syspolicy | grep -E "state|successive"` and wait while it is down.

## Handoff
Run `task-board handoff TASK-261002-35vadl --role researcher` and END TURN.
