# Windows gate diagnostic

Task: TASK-261016-ont5q1.

First hosted candidate: `726285604baae987a8c0fb6273639c29ae881807`.
Run: https://github.com/relux-works/curator/actions/runs/39026822677/job/117379337360

The hosted job reported `go test overall exit=0`, then platform-case gate exit 1. No failed test or package event exists in its uploaded `go-test.json`. The two classifier failures are:

- `TestMuseCredentialPolicyNarrowingMutants/denied-root-admitted`
- `TestMuseRootContextRefusals/denied`

Both printed `chmod cannot establish Windows ACL denial`, an unrecognized skip reason. Both fixtures really depend on POSIX chmod, and Windows ACL denial is explicitly outside their claim. The correction prefixes the messages with the existing `POSIX mode-bit unreadability` vocabulary. The skip condition and the classifier policy remain unchanged.

| Diagnostic command | Exit | Bound |
|---|---:|---|
| `CI_GATE_GOOS=windows bash .github/ci/platform-case-gate.sh .temp/TASK-261016-ont5q1/windows-evidence/test/go-test.json .temp/TASK-261016-ont5q1/windows-replay-red` | 1 | Original uploaded stream reproduces exactly the two unknown-class failures |
| `CI_GATE_GOOS=windows bash .github/ci/platform-case-gate.sh .temp/TASK-261016-ont5q1/windows-reclassified-fixture.json .temp/TASK-261016-ont5q1/windows-replay-green` | 0 | Fixture with exactly two skip-message replacements verifies existing-class compatibility only |

The second input is a deliberately transformed local fixture, not hosted execution evidence or a forged status. It does not qualify Windows denial or turn the first hosted gate green. The corrected candidate requires a separate real hosted run: https://github.com/relux-works/curator/actions/runs/39029607579 .
