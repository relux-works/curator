# TASK-260916-1ihonr — launcher config family ownership check (THE ONLY CURRENT INSTRUCTION)

Repository: curator-agent-launcher (this control root). Security epic EPIC-260910-2hw1xb, Story STORY-260916-33vuzm. Read the launcher SPEC
§4.3/§4.6/§4.7 and cmd/curator-run configuration loading. Implement: defaults.json and ax.json (machine and operator files) are validated for
ownership and permissions before use — refuse a symlinked file, a file owned by another user, or a group/world-writable file (and, on
Windows, a DACL granting write to another identity) with a named diagnostic; absent files keep today's behaviour; an unreadable file is a
refusal, never "absent". Add the contract to SPEC §4.7. Rows for each refusal and the happy path; one mutant per rule killed (real exit
codes). Tests use fake-ax only; no interactive permission bypass. CHANGELOG per this repo's convention (entry text in results if the repo
forbids leaf CHANGELOG edits). Attach results, check DoD, `task-board handoff TASK-260916-1ihonr --role developer`, END YOUR TURN.
