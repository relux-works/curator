# TASK-260730-1fsbqd: land-accepted-curator-spec-candidate

## Description
Create one reviewable commit from the independently accepted TASK-260729-3nx97g curator-spec rc.5 candidate on exact origin/main base 57c1f56846d221ecc55786bd3c2467ec32f11730, verify the committed tree against the accepted 447-file manifest and candidate digests, then land that exact reviewed commit to relux-works/curator-spec origin/main. This authorizes main synchronization only, not tag, release publication, signing, or downstream released-pin advancement.

## Scope
Accepted curator-spec rc.5 candidate commit, independent verification, and fast-forward main landing

## Acceptance Criteria
The exact independently accepted 447-file curator-spec rc.5 candidate is committed without byte drift or dirty-primary contamination, independently reviewed to done, then that exact commit is fast-forward pushed to relux-works/curator-spec main without changing downstream implementation pins. Tag and GitHub Release are explicitly deferred until a new human command.
