# TASK-260924-mcmova — spec erratum: hard-link substitution vs Windows system executables (THE ONLY CURRENT INSTRUCTION)

Control root: curator-spec; your Story worktree only. Read `campaign-producer-rules.md`. Gate = the board validation
command at handoff; locally run the three `validate:` recipe lines in bounded parts (unittest per file) + regenerate-check.

Context: core.md (enforced script commands) says exec names resolve "under the interpreter resolution rules above", which
require "Symlink, reparse-point, and hard-link substitution MUST be rejected". On Windows, `%SystemRoot%\System32\cmd.exe`
is natively hard-linked from the component store (WinSxS), so a literal nlink>1 rejection makes `cmd.exe` undeclarable.
Curator R5 (TASK-260916-2ok97n) bounds an allowance to: default Windows search list only, canonical System32 derived from
the MANAGER's captured SYSTEMROOT, target physically below it.
Deliverable: an erratum that defines "substitution" (a link that makes the resolved identity differ from the platform-owned
file the manager intended) and states that bounded System32 case explicitly, keeping every other multiply-linked target
rejected and the interpreter rule for python/node unchanged. Update profiles/manager.md where it restates the rule, the
conformance vector/case if one encodes it, CHANGELOG (unreleased). No other normative change.
Attach `TASK-260924-mcmova_results.md` (before/after text, validation table), check the DoD item, `task-board handoff TASK-260924-mcmova --role developer`.
A `run_wrote_outside_worktree … policy warn` block is a warning — verify status `to-review`.
