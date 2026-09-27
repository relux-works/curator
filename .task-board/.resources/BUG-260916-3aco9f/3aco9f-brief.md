# BUG-260916-3aco9f — profile install/reinstall drops --use/--takeover: residual shapes (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md`, the task description and issue https://github.com/relux-works/curator/issues/73 (read-only). TASK-260907-187z6x
(landed on main in STORY-260906-1a2i5a) fixed the git same-source reinstall shape with eight run()-driven rows; path reinstall was fixed
earlier. Cover the RESIDUAL: enumerate every install shape (git, path, repository/Skillfile schema-2 sources incl. collections, local,
global/project scopes) × {first install, same-source reinstall, changed-source reinstall} × {--use, --takeover, neither} and prove through
the CLI production entry that each honours --use/--takeover exactly as a first install does; fix any shape that does not. Table in results
(shape → row → status). Mutant: drop the activation step for one shape → killed. No CHANGELOG/LOGBOOK edit (entry text in results).
Focused bounded runs (host memory). Attach results, check DoD, `task-board handoff BUG-260916-3aco9f --role developer`. Do not post to the issue.
A write-boundary `policy warn` block is a warning.
