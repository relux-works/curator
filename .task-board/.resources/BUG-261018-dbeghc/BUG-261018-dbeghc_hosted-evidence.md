# BUG-261018-dbeghc hosted gate evidence

[Workflow run](https://github.com/relux-works/curator/actions/runs/39088125072) concluded **success** on signed candidate `9e9d63976b5adfa61879183e913de8a1d7d61da8` (signature G). `gh run watch 39088125072 --exit-status` exited **0**. All 20 executed jobs succeeded. Two conditional jobs were skipped: dispatch-only Candidate suite and main-only Test (rose-air). No hosted failure is replaced by a local mirror.

| Job | Conclusion | Evidence |
| --- | --- | --- |
| Lint | success | [Job](https://github.com/relux-works/curator/actions/runs/39088125072/job/117558903205) |
| Go 1.25.5 driver (windows-latest) | success | [Job](https://github.com/relux-works/curator/actions/runs/39088125072/job/117558903310) |
| Interop conformance gate | success | [Job](https://github.com/relux-works/curator/actions/runs/39088125072/job/117558903339) |
| Gate self-test (ubuntu-latest) | success | [Job](https://github.com/relux-works/curator/actions/runs/39088125072/job/117558903340) |
| Gate self-test (macos-latest) | success | [Job](https://github.com/relux-works/curator/actions/runs/39088125072/job/117558903355) |
| Test (macos-latest) | success | [Job](https://github.com/relux-works/curator/actions/runs/39088125072/job/117558903356) |
| Gate self-test (windows-latest) | success | [Job](https://github.com/relux-works/curator/actions/runs/39088125072/job/117558903367) |
| Go 1.26.0 driver (macos-latest) | success | [Job](https://github.com/relux-works/curator/actions/runs/39088125072/job/117558903377) |
| Test (ubuntu-latest) | success | [Job](https://github.com/relux-works/curator/actions/runs/39088125072/job/117558903386) |
| Race (ubuntu-latest) | success | [Job](https://github.com/relux-works/curator/actions/runs/39088125072/job/117558903399) |
| Go 1.25.5 driver (macos-latest) | success | [Job](https://github.com/relux-works/curator/actions/runs/39088125072/job/117558903410) |
| Go 1.26.0 driver (ubuntu-latest) | success | [Job](https://github.com/relux-works/curator/actions/runs/39088125072/job/117558903427) |
| Go 1.27.1 driver (macos-latest) | success | [Job](https://github.com/relux-works/curator/actions/runs/39088125072/job/117558903448) |
| Test (windows-latest) | success | [Job](https://github.com/relux-works/curator/actions/runs/39088125072/job/117558903455) |
| Go 1.25.5 driver (ubuntu-latest) | success | [Job](https://github.com/relux-works/curator/actions/runs/39088125072/job/117558903471) |
| Race (macos-latest) | success | [Job](https://github.com/relux-works/curator/actions/runs/39088125072/job/117558903478) |
| Go 1.27.1 driver (ubuntu-latest) | success | [Job](https://github.com/relux-works/curator/actions/runs/39088125072/job/117558903483) |
| Go 1.26.0 driver (windows-latest) | success | [Job](https://github.com/relux-works/curator/actions/runs/39088125072/job/117558903492) |
| Naming gate | success | [Job](https://github.com/relux-works/curator/actions/runs/39088125072/job/117558903584) |
| Go 1.27.1 driver (windows-latest) | success | [Job](https://github.com/relux-works/curator/actions/runs/39088125072/job/117558903663) |
| Candidate suite (${{ matrix.os }}) | skipped | [Job](https://github.com/relux-works/curator/actions/runs/39088125072/job/117558903837) |
| Test (rose-air) | skipped | [Job](https://github.com/relux-works/curator/actions/runs/39088125072/job/117558903978) |

## hosted-linux-permission-tests.json

```json
{
  "run": 39088125072,
  "candidate": "9e9d63976b5adfa61879183e913de8a1d7d61da8",
  "artifact": "test-evidence-ubuntu-latest",
  "stream_sha256": "0474fd4f1c3f7b4aaf5d9afa67bcd1dbb8b3fd3349bd2d0d9fc56c04c39eeebe",
  "permission_tests": {
    "TestManagedSeedPreservesNativePermissions": "pass",
    "TestManagedHomeRepairTightensDirectories": "pass",
    "TestManagedAuthoredFilesArePrivate": "pass",
    "TestManagedSeedRefusalsPreserved": "pass",
    "TestManagedFallbackPermissions": "pass",
    "TestManagedModesNarrowingMutants": "pass"
  },
  "mutant_controls": {
    "TestManagedModesNarrowingMutants/single-fallback-0644": "pass",
    "TestManagedModesNarrowingMutants/executable-tree-fallback-0755": "pass",
    "TestManagedModesNarrowingMutants/private-seed-0644": "pass",
    "TestManagedModesNarrowingMutants/claude-trust-0644": "pass",
    "TestManagedModesNarrowingMutants/codex-directory-drift-admitted": "pass",
    "TestManagedModesNarrowingMutants/codex-directory-not-tightened": "pass"
  }
}
```

## hosted-macos-permission-tests.json

```json
{
  "run": 39088125072,
  "candidate": "9e9d63976b5adfa61879183e913de8a1d7d61da8",
  "artifact": "test-evidence-macos-latest",
  "stream_sha256": "8ac91028331678b325c65902fc1ac83f4946ee56dc8bef7479a4813ca9c62d29",
  "permission_tests": {
    "TestManagedSeedPreservesNativePermissions": "pass",
    "TestManagedHomeRepairTightensDirectories": "pass",
    "TestManagedAuthoredFilesArePrivate": "pass",
    "TestManagedSeedRefusalsPreserved": "pass",
    "TestManagedFallbackPermissions": "pass",
    "TestManagedModesNarrowingMutants": "pass"
  },
  "mutant_controls": {
    "TestManagedModesNarrowingMutants/single-fallback-0644": "pass",
    "TestManagedModesNarrowingMutants/executable-tree-fallback-0755": "pass",
    "TestManagedModesNarrowingMutants/private-seed-0644": "pass",
    "TestManagedModesNarrowingMutants/claude-trust-0644": "pass",
    "TestManagedModesNarrowingMutants/codex-directory-drift-admitted": "pass",
    "TestManagedModesNarrowingMutants/codex-directory-not-tightened": "pass"
  }
}
```

Downloaded test-event streams were parsed without failing events. All six new permission suites and six expected-red mutant controls passed on both Linux and macOS. Stream hashes above bind the original hosted artifacts. The control wrappers pass only when their child command exits 1 for the named permission failure. Windows backend properties are covered by its full successful Test/platform-case job.
