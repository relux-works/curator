# THE ONLY CURRENT INSTRUCTION — TASK-261001-31dgus carry two research docs without LOGBOOK.md

Two research Change Requests were accepted but must not land as they are, because each also edits LOGBOOK.md. Producers never edit LOGBOOK.md.
- TASK-261001-3qugz9 rev1. Patch resource `TASK-261001-3qugz9_change-request_rev1.patch`, which adds:
  - .research/261001_second-operator-requirements-answers.md
  - .research/TASK-261001-3qugz9_capture-provider.py
  - .research/TASK-261001-3qugz9_evidence.json
  - .research/TASK-261001-3qugz9_probe.py
- TASK-261001-3s8csu rev1. Patch resource `TASK-261001-3s8csu_change-request_rev1.patch`, which adds:
  - .research/261001_mandates-launch-context-advice.md

Do:
1. In your workspace (current trunk), apply ONLY the `.research/` hunks of both patches, byte-identically. Do NOT touch LOGBOOK.md or anything else.
2. Prove byte identity: for each of the 5 files, the sha256 must equal the blob in the respective CR candidate tree (see .temp/changerequests/<TASK>/rev-000001.json candidate_tree_oid). Use `git rev-parse <tree>:<path>` against `git hash-object <file>`.
3. Record the five path/hash pairs in the results. Then run `task-board handoff TASK-261001-31dgus --role developer` and END YOUR TURN.

Never spell any employer name.
