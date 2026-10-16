# Review B hosted gate

Run: https://github.com/relux-works/curator/actions/runs/38943768729
Head: 6daf5436174f7abb20470c8ec973bfde6f416b1f
Candidate tree: 43b7ed0e9efd36ffcc321f6604d72aea221e2af7
Conclusion: success. 20 jobs succeeded; 2 optional jobs skipped by workflow conditions.

Independent execution of repository CI through the inspected signed snapshot variant of scripts/remote-gate.sh. Only snapshot signing/identity and an exact-tree assertion differ; workflow and job commands are unchanged. Story HEAD, index and candidate code were not changed.

| Job | Result |
| --- | --- |
| Lint | success |
| Test (macos-latest) | success |
| Gate self-test (windows-latest) | success |
| Go 1.25.5 driver (ubuntu-latest) | success |
| Race (macos-latest) | success |
| Gate self-test (ubuntu-latest) | success |
| Naming gate | success |
| Interop conformance gate | success |
| Test (ubuntu-latest) | success |
| Test (windows-latest) | success |
| Race (ubuntu-latest) | success |
| Go 1.26.0 driver (macos-latest) | success |
| Go 1.25.5 driver (macos-latest) | success |
| Gate self-test (macos-latest) | success |
| Go 1.26.0 driver (ubuntu-latest) | success |
| Go 1.25.5 driver (windows-latest) | success |
| Go 1.27.1 driver (windows-latest) | success |
| Go 1.27.1 driver (ubuntu-latest) | success |
| Go 1.27.1 driver (macos-latest) | success |
| Go 1.26.0 driver (windows-latest) | success |
| Optional candidate suite | skipped |
| Optional self-hosted test | skipped |
