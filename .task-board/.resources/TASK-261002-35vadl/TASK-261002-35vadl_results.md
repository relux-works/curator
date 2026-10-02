# TASK-261002-35vadl — oss-audit-wave-1

| Repository | Verdict | S git/blob/meta/path | P | E confirmed/raw/context | I author/committer/email | L HEAD/history | R literal/dynamic |
|---|---|---|---|---|---|---|---|---|
| relux-works/skill-agents-attachments | FIX-HEAD | 0/0/0/0 | 0 | 0/0/0 | 1/1/6 | 3/5 | 0/0 |
| relux-works/skill-pdf | FIX-HEAD | 0/0/0/0 | 0 | 0/0/0 | 1/1/2 | 3/2 | 0/0 |
| relux-works/skill-youtube-scraper | FIX-HEAD | 0/0/0/0 | 0 | 0/0/3 | 3/3/0 | 2/4 | 0/0 |
| relux-works/skill-git-filter-repo | FIX-HEAD | 0/0/0/0 | 0 | 0/0/8 | 3/3/24 | 2/7 | 0/0 |
| relux-works/skill-youtrack | NEEDS-NEW-REPO | 0/0/0/0 | 0 | 37/37/30 | 1/1/0 | 3/20 | 0/0 |
| relux-works/relux-codes-landing | FIX-HEAD | 0/0/0/0 | 0 | 0/7/32 | 3/3/16 | 3/10 | 0/0 |
| ivanopcode/swift-realtime-openai-local | HOLD-pending-decision | not in scope | not in scope | not in scope | not in scope | UNKNOWN | not in scope |

S: secrets alerts from Git history / unique blobs / metadata / sensitive filenames. P: personal-data matching lines. E: confirmed naming matches / raw candidates / generic client-context candidates. I: distinct author identities / distinct committer identities / email-bearing blob, message, path and tag lines. L: current requested-policy issues / historical trees with policy observations. R: self-hosted workflow candidates / unresolved dynamic runs-on expressions. Content findings are deduplicated by object, line and subtype; context candidates are not confirmed client names. Licence-policy counts include missing NOTICE and SPDX markers; no blanket legal requirement is inferred. Identity consent remains unknown.

| Repository | Refs matched/advertised | Commits read/reachable | Trees read/distinct | Blobs read/distinct | Workflow blobs |
|---|---|---|---|---|---|
| relux-works/skill-agents-attachments | 7/7 | 5/5 | 5/5 | 18/18 | 0 |
| relux-works/skill-pdf | 4/4 | 2/2 | 2/2 | 11/11 | 0 |
| relux-works/skill-youtube-scraper | 3/3 | 4/4 | 4/4 | 73/73 | 0 |
| relux-works/skill-git-filter-repo | 3/3 | 7/7 | 7/7 | 9/9 | 0 |
| relux-works/skill-youtrack | 1/1 | 20/20 | 20/20 | 102/102 | 0 |
| relux-works/relux-codes-landing | 1/1 | 10/10 | 10/10 | 113/113 | 0 |
| ivanopcode/swift-realtime-openai-local | UNKNOWN | not inspected | not inspected | not inspected | not in scope |

Local report: `~/oss-audit/wave1/REPORT.md` (file 0600; directory 0700).

Tool versions: git version 2.50.1 (Apple Git-155); gh version 2.88.1 (2026-03-12); 8.30.1; Python 3.14.6.

What could not be checked:

- The MIT-only target is inaccessible: SSH clone exit 128, repository API exit 1 (404/inaccessible), HTTPS credential-helper clone exit 128. Upstream MIT attribution is UNKNOWN; 0/1 target accessible.
- Deleted or server-hidden refs, remote reflogs and unadvertised/local stashes. Advertised stash refs: 0 across the six accessible repositories.
- Undisclosed client names and publication consent: no authoritative client-name inventory or consent record was supplied.
- Image pixels/OCR/steganographic content. All 3 binary image blobs were byte-scanned, metadata-checked, and checked as UTF-16 in both byte orders; compressed-byte naming coincidences excluded: 7.
- Independent vendor provenance beyond explicit files/notices, deployed contact rendering, mailbox ownership and delivery. Literal public contact addresses at HEAD: 1; details only in the restricted report.
