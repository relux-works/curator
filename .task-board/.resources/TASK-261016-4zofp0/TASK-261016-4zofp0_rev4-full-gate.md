# Revision 4 full-gate evidence

Command: `sh scripts/remote-gate.sh` (after sourcing the configured agent Git environment).
Real process exit: **0**. No pipeline or tee. No local mirror or forged status.

Hosted run: [https://github.com/relux-works/curator/actions/runs/38958207647](https://github.com/relux-works/curator/actions/runs/38958207647). Run conclusion: **success**.
Snapshot: `3bdf4c4f77a89afc12c06c242f69cd3d16deda9e`; tree: `24299c092bc02d004545ff7103ffc192b541db37`; signature: **G**.
Story base and unchanged HEAD: `c29663024abfe6dc0353749018fd65b2bf1833e1`. Final candidate tree equals the hosted tree; the Story candidate remains uncommitted.

**20 jobs succeeded; 2 conditional jobs skipped.** Standard test suites on macOS, Linux and Windows, race suites on macOS and Linux, all nine Go-family driver jobs (1.25.5/1.26.0/1.27.1 on all three platforms), lint, self-tests, naming and interoperability gates succeeded. The optional self-hosted test and non-default candidate-spec suite were skipped by workflow conditions; no execution claim is made for either. Windows POSIX sentinel/mode/unreadability bounds remain as documented in the coverage map.

| Hosted job | Conclusion |
|---|---|
| Lint | success |
| Test (macos-latest) | success |
| Gate self-test (ubuntu-latest) | success |
| Gate self-test (macos-latest) | success |
| Go 1.27.1 driver (macos-latest) | success |
| Test (ubuntu-latest) | success |
| Naming gate | success |
| Race (macos-latest) | success |
| Go 1.25.5 driver (macos-latest) | success |
| Go 1.25.5 driver (ubuntu-latest) | success |
| Gate self-test (windows-latest) | success |
| Go 1.26.0 driver (macos-latest) | success |
| Interop conformance gate | success |
| Race (ubuntu-latest) | success |
| Go 1.25.5 driver (windows-latest) | success |
| Go 1.26.0 driver (ubuntu-latest) | success |
| Test (windows-latest) | success |
| Go 1.27.1 driver (ubuntu-latest) | success |
| Go 1.26.0 driver (windows-latest) | success |
| Go 1.27.1 driver (windows-latest) | success |
| Optional self-hosted test (conditional) | skipped |
| Candidate-spec suite (non-default conditional) | skipped |

## Standalone process output

```text
remote gate: attempt 1/3: pushing 3bdf4c4f77a89afc12c06c242f69cd3d16deda9e as gate/STORY-261016-1wj1v9/261016-195313-4397-1
remote: 
remote: Create a pull request for 'gate/STORY-261016-1wj1v9/261016-195313-4397-1' on GitHub by visiting:        
remote:      https://github.com/relux-works/curator/pull/new/gate/STORY-261016-1wj1v9/261016-195313-4397-1        
remote: 
remote gate: run 38958207647 (https://github.com/relux-works/curator/actions/runs/38958207647)
remote gate: run 38958207647 finished: success
  Lint: success  
  Test (macos-latest): success  
  Gate self-test (ubuntu-latest): success  
  Gate self-test (macos-latest): success  
  Go 1.27.1 driver (macos-latest): success  
  Test (ubuntu-latest): success  
  Naming gate: success  
  Race (macos-latest): success  
  Go 1.25.5 driver (macos-latest): success  
  Go 1.25.5 driver (ubuntu-latest): success  
  Gate self-test (windows-latest): success  
  Go 1.26.0 driver (macos-latest): success  
  Interop conformance gate: success  
  Race (ubuntu-latest): success  
  Go 1.25.5 driver (windows-latest): success  
  Go 1.26.0 driver (ubuntu-latest): success  
  Test (windows-latest): success  
  Go 1.27.1 driver (ubuntu-latest): success  
  Go 1.26.0 driver (windows-latest): success  
  Go 1.27.1 driver (windows-latest): success  
  Optional self-hosted test: skipped  
  Candidate suite (${{ matrix.os }}): skipped  
```
