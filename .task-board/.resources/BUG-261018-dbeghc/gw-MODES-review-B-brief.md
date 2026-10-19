# Review — BUG-261018-dbeghc: managed-home seeds are world-readable

Review the exact candidate. Sweep these surfaces and give every row a result.

| Surface | What must hold |
| --- | --- |
| seed modes | every seed copy keeps the source bits and is never wider than the source; a 0600 source stays 0600 |
| directory modes | the environments root and each managed home are 0700 on creation |
| repair | an existing wider directory is tightened, contents preserved, under the existing lock and journal |
| other managed writes | each write site uses the documented table; trust-bearing or credential-adjacent files are 0600 |
| refusals | symlinked, unreadable and special-file seeds refuse exactly as before |
| platforms | the behavior is correct on macOS and Linux; Windows makes no false claim |
| tests | negative tests exist; a 0644 mutant fails for the stated reason |
| hosted gate | the full gate is green on the exact signed candidate |

Accept with `accept_cr` only when every row holds. Otherwise request changes and list the failing rows.
