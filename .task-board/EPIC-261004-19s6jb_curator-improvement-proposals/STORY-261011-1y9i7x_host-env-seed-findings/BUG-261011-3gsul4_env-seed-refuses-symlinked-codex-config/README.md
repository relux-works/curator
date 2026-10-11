# BUG-261011-3gsul4: env-seed-refuses-symlinked-codex-config

## Description
curator env resolve claude_code fails with environment_seed_unreadable because ~/.codex/config.toml is a symlink to the central agents-management config under ~/.agents/.configs. Curator refuses non-regular seed files; agents-management installs symlinks. Decide one contract: Curator admits a same-owner symlink resolving inside the user home (opened without following further links, target identity recorded), or agents-management installs regular files. Reported by tb-keeper from run 3arj7b, 2026-10-11. Do not change the host file in place.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
(define fix acceptance criteria)
