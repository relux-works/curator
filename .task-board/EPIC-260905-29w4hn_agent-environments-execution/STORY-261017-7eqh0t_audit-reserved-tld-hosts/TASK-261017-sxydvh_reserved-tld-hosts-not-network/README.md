# TASK-261017-sxydvh: reserved-tld-hosts-not-network

## Description
Field request from an operator whose skill tests write reserved fixture hosts (for example a .test host) into config files and assert them, never fetching them; today audit.capability.network-undeclared flags them (only localhost and 127.0.0.1 are exempt in internal/audit). Requested controls: a .test host under a reserved TLD is not flagged; a lookalike such as x.test.evil.com and a bare 'test' are still flagged; example.com, example.net and example.org stay flagged so the static canary (exfiltrate.example.com) still fires; a mixed file with one reserved and one real undeclared host still reports the real one. Blocked on: curator-spec contract text and owner approval of the audit change.

## Scope
internal/audit network-contact classification and every audit path that uses it (member, migration, cache); tests; no spec change

## Acceptance Criteria
1. Loopback hosts and hosts whose normalized last label is test, example, invalid or localhost are not network contacts in member, migration and cached audits (curator-spec c1a2bb25, environments 9.1 and 9.4). 2. Lookalikes such as x.test.evil.com, a bare test label and example.com, example.net, example.org stay contacts; the static canary exfiltrate.example.com still blocks; a mixed file still reports its real undeclared host with the file. 3. All 40 curator-spec network-contact vectors are executable Go tests. 4. Named mutants go red for their reason; a pre-change cached finding cannot mask the new classification. 5. Full local suite and full hosted gate green on the exact signed candidate.
