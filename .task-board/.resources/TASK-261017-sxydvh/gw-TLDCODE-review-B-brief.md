# Review — TASK-261017-sxydvh: reserved special-use hosts are not network contacts

Review the exact candidate against curator-spec main c1a2bb25, environments §9.1 and §9.4. Where the producer's notes and the spec disagree, the spec wins.

Sweep these surfaces and give every row a result:

| Surface | What must hold |
| --- | --- |
| host normalization | case and a trailing dot are normalized before the last-label check, the same way in every audit path |
| exemption set | exactly loopback plus the last labels `test`, `example`, `invalid` and `localhost`; nothing else |
| lookalikes | `x.test.evil.com`, a bare `test` and `test.com` stay contacts |
| canary | example.com, example.net and example.org stay contacts; `exfiltrate.example.com` still blocks |
| mixed file | a real undeclared host beside a reserved one is still reported, with its file |
| migration audit | §9.4 migration uses the same classification |
| audit cache | a finding cached before the change cannot mask the new classification, and the reverse also holds |
| vectors | all 40 network-contact vectors are mapped to executable tests |
| mutants | each named mutant goes red for the stated reason |
| hosted gate | the full gate is green on this exact signed candidate |

Accept with `accept_cr` only when every row holds; otherwise request changes with the failing rows.
