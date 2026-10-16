# Hosted full-gate evidence

Command: `sh scripts/remote-gate.sh`. Supervising process exit: **0**.

Snapshot: c82d1cd07ba9d550b6c10c0fe9910ec27e28d592

Run: https://github.com/relux-works/curator/actions/runs/38904812207

Optional self-hosted and candidate-suite jobs were skipped by workflow conditions; they are not reported as executed. All executed jobs succeeded. The original optional self-hosted label is neutralized below.

## Script output

```text
remote gate: attempt 1/3: pushing c82d1cd07ba9d550b6c10c0fe9910ec27e28d592 as gate/STORY-261016-26uala/261016-130215-63750-1
remote: 
remote: Create a pull request for 'gate/STORY-261016-26uala/261016-130215-63750-1' on GitHub by visiting:        
remote:      https://github.com/relux-works/curator/pull/new/gate/STORY-261016-26uala/261016-130215-63750-1        
remote: 
remote gate: run 38904812207 (https://github.com/relux-works/curator/actions/runs/38904812207)
remote gate: run 38904812207 finished: success
  Race (macos-latest): success  
  Race (ubuntu-latest): success  
  Lint: success  
  Gate self-test (ubuntu-latest): success  
  Gate self-test (windows-latest): success  
  Gate self-test (macos-latest): success  
  Naming gate: success  
  Interop conformance gate: success  
  Test (ubuntu-latest): success  
  Test (macos-latest): success  
  Test (windows-latest): success  
  Go 1.26.0 driver (ubuntu-latest): success  
  Go 1.26.0 driver (macos-latest): success  
  Go 1.25.5 driver (ubuntu-latest): success  
  Go 1.27.1 driver (windows-latest): success  
  Go 1.25.5 driver (macos-latest): success  
  Go 1.26.0 driver (windows-latest): success  
  Go 1.27.1 driver (macos-latest): success  
  Go 1.27.1 driver (ubuntu-latest): success  
  Go 1.25.5 driver (windows-latest): success  
  Candidate suite (${{ matrix.os }}): skipped  
  Test (optional-self-hosted): skipped  
```

## Hosted verdict metadata

```json
{
  "conclusion": "success",
  "headSha": "c82d1cd07ba9d550b6c10c0fe9910ec27e28d592",
  "jobs": [
    {
      "completedAt": "2026-10-16T13:26:48Z",
      "conclusion": "success",
      "databaseId": 116985608045,
      "name": "Race (macos-latest)",
      "startedAt": "2026-10-16T13:02:26Z",
      "status": "completed",
      "steps": [
        {
          "completedAt": "2026-10-16T13:02:27Z",
          "conclusion": "success",
          "name": "Set up job",
          "number": 1,
          "startedAt": "2026-10-16T13:02:26Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:39Z",
          "conclusion": "success",
          "name": "Run actions/checkout@v4",
          "number": 2,
          "startedAt": "2026-10-16T13:02:27Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:40Z",
          "conclusion": "success",
          "name": "Checkout authoritative protocol suite",
          "number": 3,
          "startedAt": "2026-10-16T13:02:39Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:59Z",
          "conclusion": "success",
          "name": "Run actions/setup-go@v5",
          "number": 4,
          "startedAt": "2026-10-16T13:02:40Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:00Z",
          "conclusion": "success",
          "name": "Run actions/setup-node@v4",
          "number": 5,
          "startedAt": "2026-10-16T13:02:59Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:00Z",
          "conclusion": "success",
          "name": "Verify pnpm pin matches Go constant",
          "number": 6,
          "startedAt": "2026-10-16T13:03:00Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:01Z",
          "conclusion": "success",
          "name": "Install pinned pnpm via npm",
          "number": 7,
          "startedAt": "2026-10-16T13:03:00Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:01Z",
          "conclusion": "success",
          "name": "Verify rust pin matches Go constant",
          "number": 8,
          "startedAt": "2026-10-16T13:03:01Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:08Z",
          "conclusion": "success",
          "name": "Install pinned Rust toolchain via rustup",
          "number": 9,
          "startedAt": "2026-10-16T13:03:01Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:08Z",
          "conclusion": "success",
          "name": "Verify Go toolchain identity",
          "number": 10,
          "startedAt": "2026-10-16T13:03:08Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:26:39Z",
          "conclusion": "success",
          "name": "go test -race + platform-case gate",
          "number": 11,
          "startedAt": "2026-10-16T13:03:08Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:26:41Z",
          "conclusion": "success",
          "name": "Upload race evidence",
          "number": 12,
          "startedAt": "2026-10-16T13:26:39Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:26:41Z",
          "conclusion": "success",
          "name": "Post Run actions/setup-node@v4",
          "number": 21,
          "startedAt": "2026-10-16T13:26:41Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:26:41Z",
          "conclusion": "success",
          "name": "Post Run actions/setup-go@v5",
          "number": 22,
          "startedAt": "2026-10-16T13:26:41Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:26:42Z",
          "conclusion": "success",
          "name": "Post Checkout authoritative protocol suite",
          "number": 23,
          "startedAt": "2026-10-16T13:26:41Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:26:43Z",
          "conclusion": "success",
          "name": "Post Run actions/checkout@v4",
          "number": 24,
          "startedAt": "2026-10-16T13:26:42Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:26:45Z",
          "conclusion": "success",
          "name": "Complete job",
          "number": 25,
          "startedAt": "2026-10-16T13:26:43Z",
          "status": "completed"
        }
      ],
      "url": "https://github.com/relux-works/curator/actions/runs/38904812207/job/116985608045"
    },
    {
      "completedAt": "2026-10-16T13:31:23Z",
      "conclusion": "success",
      "databaseId": 116985608260,
      "name": "Race (ubuntu-latest)",
      "startedAt": "2026-10-16T13:02:22Z",
      "status": "completed",
      "steps": [
        {
          "completedAt": "2026-10-16T13:02:25Z",
          "conclusion": "success",
          "name": "Set up job",
          "number": 1,
          "startedAt": "2026-10-16T13:02:23Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:35Z",
          "conclusion": "success",
          "name": "Run actions/checkout@v4",
          "number": 2,
          "startedAt": "2026-10-16T13:02:25Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:37Z",
          "conclusion": "success",
          "name": "Checkout authoritative protocol suite",
          "number": 3,
          "startedAt": "2026-10-16T13:02:35Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:48Z",
          "conclusion": "success",
          "name": "Run actions/setup-go@v5",
          "number": 4,
          "startedAt": "2026-10-16T13:02:37Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:49Z",
          "conclusion": "success",
          "name": "Run actions/setup-node@v4",
          "number": 5,
          "startedAt": "2026-10-16T13:02:48Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:49Z",
          "conclusion": "success",
          "name": "Verify pnpm pin matches Go constant",
          "number": 6,
          "startedAt": "2026-10-16T13:02:49Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:50Z",
          "conclusion": "success",
          "name": "Install pinned pnpm via npm",
          "number": 7,
          "startedAt": "2026-10-16T13:02:49Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:50Z",
          "conclusion": "success",
          "name": "Verify rust pin matches Go constant",
          "number": 8,
          "startedAt": "2026-10-16T13:02:50Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:58Z",
          "conclusion": "success",
          "name": "Install pinned Rust toolchain via rustup",
          "number": 9,
          "startedAt": "2026-10-16T13:02:50Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:58Z",
          "conclusion": "success",
          "name": "Verify Go toolchain identity",
          "number": 10,
          "startedAt": "2026-10-16T13:02:58Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:31:16Z",
          "conclusion": "success",
          "name": "go test -race + platform-case gate",
          "number": 11,
          "startedAt": "2026-10-16T13:02:58Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:31:18Z",
          "conclusion": "success",
          "name": "Upload race evidence",
          "number": 12,
          "startedAt": "2026-10-16T13:31:16Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:31:18Z",
          "conclusion": "success",
          "name": "Post Run actions/setup-node@v4",
          "number": 21,
          "startedAt": "2026-10-16T13:31:18Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:31:18Z",
          "conclusion": "success",
          "name": "Post Run actions/setup-go@v5",
          "number": 22,
          "startedAt": "2026-10-16T13:31:18Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:31:19Z",
          "conclusion": "success",
          "name": "Post Checkout authoritative protocol suite",
          "number": 23,
          "startedAt": "2026-10-16T13:31:18Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:31:19Z",
          "conclusion": "success",
          "name": "Post Run actions/checkout@v4",
          "number": 24,
          "startedAt": "2026-10-16T13:31:19Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:31:19Z",
          "conclusion": "success",
          "name": "Complete job",
          "number": 25,
          "startedAt": "2026-10-16T13:31:19Z",
          "status": "completed"
        }
      ],
      "url": "https://github.com/relux-works/curator/actions/runs/38904812207/job/116985608260"
    },
    {
      "completedAt": "2026-10-16T13:03:35Z",
      "conclusion": "success",
      "databaseId": 116985608350,
      "name": "Lint",
      "startedAt": "2026-10-16T13:02:22Z",
      "status": "completed",
      "steps": [
        {
          "completedAt": "2026-10-16T13:02:25Z",
          "conclusion": "success",
          "name": "Set up job",
          "number": 1,
          "startedAt": "2026-10-16T13:02:23Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:37Z",
          "conclusion": "success",
          "name": "Run actions/checkout@v4",
          "number": 2,
          "startedAt": "2026-10-16T13:02:25Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:49Z",
          "conclusion": "success",
          "name": "Run actions/setup-go@v5",
          "number": 3,
          "startedAt": "2026-10-16T13:02:37Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:49Z",
          "conclusion": "success",
          "name": "Verify Go toolchain identity",
          "number": 4,
          "startedAt": "2026-10-16T13:02:49Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:32Z",
          "conclusion": "success",
          "name": "Run golangci/golangci-lint-action@v7",
          "number": 5,
          "startedAt": "2026-10-16T13:02:49Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:32Z",
          "conclusion": "success",
          "name": "No broad suppression",
          "number": 6,
          "startedAt": "2026-10-16T13:03:32Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:32Z",
          "conclusion": "success",
          "name": "Verify GoReleaser rc channel values",
          "number": 7,
          "startedAt": "2026-10-16T13:03:32Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:32Z",
          "conclusion": "success",
          "name": "Post Run golangci/golangci-lint-action@v7",
          "number": 12,
          "startedAt": "2026-10-16T13:03:32Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:32Z",
          "conclusion": "success",
          "name": "Post Run actions/setup-go@v5",
          "number": 13,
          "startedAt": "2026-10-16T13:03:32Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:33Z",
          "conclusion": "success",
          "name": "Post Run actions/checkout@v4",
          "number": 14,
          "startedAt": "2026-10-16T13:03:32Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:33Z",
          "conclusion": "success",
          "name": "Complete job",
          "number": 15,
          "startedAt": "2026-10-16T13:03:33Z",
          "status": "completed"
        }
      ],
      "url": "https://github.com/relux-works/curator/actions/runs/38904812207/job/116985608350"
    },
    {
      "completedAt": "2026-10-16T13:02:49Z",
      "conclusion": "success",
      "databaseId": 116985608360,
      "name": "Gate self-test (ubuntu-latest)",
      "startedAt": "2026-10-16T13:02:22Z",
      "status": "completed",
      "steps": [
        {
          "completedAt": "2026-10-16T13:02:24Z",
          "conclusion": "success",
          "name": "Set up job",
          "number": 1,
          "startedAt": "2026-10-16T13:02:23Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:31Z",
          "conclusion": "success",
          "name": "Run actions/checkout@v4",
          "number": 2,
          "startedAt": "2026-10-16T13:02:24Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:48Z",
          "conclusion": "success",
          "name": "Self-test the CI gates",
          "number": 3,
          "startedAt": "2026-10-16T13:02:31Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:48Z",
          "conclusion": "success",
          "name": "Post Run actions/checkout@v4",
          "number": 6,
          "startedAt": "2026-10-16T13:02:48Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:48Z",
          "conclusion": "success",
          "name": "Complete job",
          "number": 7,
          "startedAt": "2026-10-16T13:02:48Z",
          "status": "completed"
        }
      ],
      "url": "https://github.com/relux-works/curator/actions/runs/38904812207/job/116985608360"
    },
    {
      "completedAt": "2026-10-16T13:03:43Z",
      "conclusion": "success",
      "databaseId": 116985608398,
      "name": "Gate self-test (windows-latest)",
      "startedAt": "2026-10-16T13:02:23Z",
      "status": "completed",
      "steps": [
        {
          "completedAt": "2026-10-16T13:02:25Z",
          "conclusion": "success",
          "name": "Set up job",
          "number": 1,
          "startedAt": "2026-10-16T13:02:24Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:37Z",
          "conclusion": "success",
          "name": "Run actions/checkout@v4",
          "number": 2,
          "startedAt": "2026-10-16T13:02:25Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:40Z",
          "conclusion": "success",
          "name": "Self-test the CI gates",
          "number": 3,
          "startedAt": "2026-10-16T13:02:37Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:42Z",
          "conclusion": "success",
          "name": "Post Run actions/checkout@v4",
          "number": 6,
          "startedAt": "2026-10-16T13:03:40Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:42Z",
          "conclusion": "success",
          "name": "Complete job",
          "number": 7,
          "startedAt": "2026-10-16T13:03:42Z",
          "status": "completed"
        }
      ],
      "url": "https://github.com/relux-works/curator/actions/runs/38904812207/job/116985608398"
    },
    {
      "completedAt": "2026-10-16T13:02:53Z",
      "conclusion": "success",
      "databaseId": 116985608458,
      "name": "Gate self-test (macos-latest)",
      "startedAt": "2026-10-16T13:02:28Z",
      "status": "completed",
      "steps": [
        {
          "completedAt": "2026-10-16T13:02:30Z",
          "conclusion": "success",
          "name": "Set up job",
          "number": 1,
          "startedAt": "2026-10-16T13:02:29Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:40Z",
          "conclusion": "success",
          "name": "Run actions/checkout@v4",
          "number": 2,
          "startedAt": "2026-10-16T13:02:30Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:50Z",
          "conclusion": "success",
          "name": "Self-test the CI gates",
          "number": 3,
          "startedAt": "2026-10-16T13:02:40Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:51Z",
          "conclusion": "success",
          "name": "Post Run actions/checkout@v4",
          "number": 6,
          "startedAt": "2026-10-16T13:02:50Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:51Z",
          "conclusion": "success",
          "name": "Complete job",
          "number": 7,
          "startedAt": "2026-10-16T13:02:51Z",
          "status": "completed"
        }
      ],
      "url": "https://github.com/relux-works/curator/actions/runs/38904812207/job/116985608458"
    },
    {
      "completedAt": "2026-10-16T13:02:41Z",
      "conclusion": "success",
      "databaseId": 116985608515,
      "name": "Naming gate",
      "startedAt": "2026-10-16T13:02:28Z",
      "status": "completed",
      "steps": [
        {
          "completedAt": "2026-10-16T13:02:29Z",
          "conclusion": "success",
          "name": "Set up job",
          "number": 1,
          "startedAt": "2026-10-16T13:02:29Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:38Z",
          "conclusion": "success",
          "name": "Run actions/checkout@v4",
          "number": 2,
          "startedAt": "2026-10-16T13:02:29Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:39Z",
          "conclusion": "success",
          "name": "Employer name gate",
          "number": 3,
          "startedAt": "2026-10-16T13:02:38Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:39Z",
          "conclusion": "success",
          "name": "Post Run actions/checkout@v4",
          "number": 6,
          "startedAt": "2026-10-16T13:02:39Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:40Z",
          "conclusion": "success",
          "name": "Complete job",
          "number": 7,
          "startedAt": "2026-10-16T13:02:40Z",
          "status": "completed"
        }
      ],
      "url": "https://github.com/relux-works/curator/actions/runs/38904812207/job/116985608515"
    },
    {
      "completedAt": "2026-10-16T13:02:51Z",
      "conclusion": "success",
      "databaseId": 116985608551,
      "name": "Interop conformance gate",
      "startedAt": "2026-10-16T13:02:22Z",
      "status": "completed",
      "steps": [
        {
          "completedAt": "2026-10-16T13:02:24Z",
          "conclusion": "success",
          "name": "Set up job",
          "number": 1,
          "startedAt": "2026-10-16T13:02:23Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:37Z",
          "conclusion": "success",
          "name": "Run actions/checkout@v4",
          "number": 2,
          "startedAt": "2026-10-16T13:02:24Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:38Z",
          "conclusion": "success",
          "name": "Checkout authoritative protocol suite",
          "number": 3,
          "startedAt": "2026-10-16T13:02:37Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:46Z",
          "conclusion": "success",
          "name": "Run actions/setup-go@v5",
          "number": 4,
          "startedAt": "2026-10-16T13:02:38Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:46Z",
          "conclusion": "success",
          "name": "Verify Go toolchain identity",
          "number": 5,
          "startedAt": "2026-10-16T13:02:46Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:48Z",
          "conclusion": "success",
          "name": "Shared protocol suite",
          "number": 6,
          "startedAt": "2026-10-16T13:02:46Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:48Z",
          "conclusion": "success",
          "name": "Post Run actions/setup-go@v5",
          "number": 10,
          "startedAt": "2026-10-16T13:02:48Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:48Z",
          "conclusion": "success",
          "name": "Post Checkout authoritative protocol suite",
          "number": 11,
          "startedAt": "2026-10-16T13:02:48Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:48Z",
          "conclusion": "success",
          "name": "Post Run actions/checkout@v4",
          "number": 12,
          "startedAt": "2026-10-16T13:02:48Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:48Z",
          "conclusion": "success",
          "name": "Complete job",
          "number": 13,
          "startedAt": "2026-10-16T13:02:48Z",
          "status": "completed"
        }
      ],
      "url": "https://github.com/relux-works/curator/actions/runs/38904812207/job/116985608551"
    },
    {
      "completedAt": "2026-10-16T13:23:48Z",
      "conclusion": "success",
      "databaseId": 116985608554,
      "name": "Test (ubuntu-latest)",
      "startedAt": "2026-10-16T13:02:43Z",
      "status": "completed",
      "steps": [
        {
          "completedAt": "2026-10-16T13:02:48Z",
          "conclusion": "success",
          "name": "Set up job",
          "number": 1,
          "startedAt": "2026-10-16T13:02:44Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:54Z",
          "conclusion": "success",
          "name": "Run actions/checkout@v4",
          "number": 2,
          "startedAt": "2026-10-16T13:02:48Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:55Z",
          "conclusion": "success",
          "name": "Checkout authoritative protocol suite",
          "number": 3,
          "startedAt": "2026-10-16T13:02:54Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:06Z",
          "conclusion": "success",
          "name": "Run actions/setup-go@v5",
          "number": 4,
          "startedAt": "2026-10-16T13:02:55Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:06Z",
          "conclusion": "success",
          "name": "Run actions/setup-node@v4",
          "number": 5,
          "startedAt": "2026-10-16T13:03:06Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:06Z",
          "conclusion": "success",
          "name": "Verify pnpm pin matches Go constant",
          "number": 6,
          "startedAt": "2026-10-16T13:03:06Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:07Z",
          "conclusion": "success",
          "name": "Install pinned pnpm via npm",
          "number": 7,
          "startedAt": "2026-10-16T13:03:06Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:07Z",
          "conclusion": "success",
          "name": "Verify rust pin matches Go constant",
          "number": 8,
          "startedAt": "2026-10-16T13:03:07Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:15Z",
          "conclusion": "success",
          "name": "Install pinned Rust toolchain via rustup",
          "number": 9,
          "startedAt": "2026-10-16T13:03:07Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:15Z",
          "conclusion": "success",
          "name": "Verify Go toolchain identity",
          "number": 10,
          "startedAt": "2026-10-16T13:03:15Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:15Z",
          "conclusion": "success",
          "name": "Verify versioned Go install compatibility",
          "number": 11,
          "startedAt": "2026-10-16T13:03:15Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:16Z",
          "conclusion": "success",
          "name": "gofmt check",
          "number": 12,
          "startedAt": "2026-10-16T13:03:15Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:33Z",
          "conclusion": "success",
          "name": "go vet",
          "number": 13,
          "startedAt": "2026-10-16T13:03:16Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:37Z",
          "conclusion": "success",
          "name": "Ledger consistency",
          "number": 14,
          "startedAt": "2026-10-16T13:03:33Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:23:42Z",
          "conclusion": "success",
          "name": "go test + platform-case gate",
          "number": 15,
          "startedAt": "2026-10-16T13:03:37Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:23:44Z",
          "conclusion": "success",
          "name": "Upload gate evidence",
          "number": 16,
          "startedAt": "2026-10-16T13:23:42Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:23:44Z",
          "conclusion": "success",
          "name": "Post Run actions/setup-node@v4",
          "number": 29,
          "startedAt": "2026-10-16T13:23:44Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:23:44Z",
          "conclusion": "success",
          "name": "Post Run actions/setup-go@v5",
          "number": 30,
          "startedAt": "2026-10-16T13:23:44Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:23:44Z",
          "conclusion": "success",
          "name": "Post Checkout authoritative protocol suite",
          "number": 31,
          "startedAt": "2026-10-16T13:23:44Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:23:45Z",
          "conclusion": "success",
          "name": "Post Run actions/checkout@v4",
          "number": 32,
          "startedAt": "2026-10-16T13:23:45Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:23:45Z",
          "conclusion": "success",
          "name": "Complete job",
          "number": 33,
          "startedAt": "2026-10-16T13:23:45Z",
          "status": "completed"
        }
      ],
      "url": "https://github.com/relux-works/curator/actions/runs/38904812207/job/116985608554"
    },
    {
      "completedAt": "2026-10-16T13:28:45Z",
      "conclusion": "success",
      "databaseId": 116985608564,
      "name": "Test (macos-latest)",
      "startedAt": "2026-10-16T13:07:45Z",
      "status": "completed",
      "steps": [
        {
          "completedAt": "2026-10-16T13:07:48Z",
          "conclusion": "success",
          "name": "Set up job",
          "number": 1,
          "startedAt": "2026-10-16T13:07:45Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:07:59Z",
          "conclusion": "success",
          "name": "Run actions/checkout@v4",
          "number": 2,
          "startedAt": "2026-10-16T13:07:48Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:08:01Z",
          "conclusion": "success",
          "name": "Checkout authoritative protocol suite",
          "number": 3,
          "startedAt": "2026-10-16T13:07:59Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:08:30Z",
          "conclusion": "success",
          "name": "Run actions/setup-go@v5",
          "number": 4,
          "startedAt": "2026-10-16T13:08:01Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:08:31Z",
          "conclusion": "success",
          "name": "Run actions/setup-node@v4",
          "number": 5,
          "startedAt": "2026-10-16T13:08:30Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:08:31Z",
          "conclusion": "success",
          "name": "Verify pnpm pin matches Go constant",
          "number": 6,
          "startedAt": "2026-10-16T13:08:31Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:08:32Z",
          "conclusion": "success",
          "name": "Install pinned pnpm via npm",
          "number": 7,
          "startedAt": "2026-10-16T13:08:31Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:08:32Z",
          "conclusion": "success",
          "name": "Verify rust pin matches Go constant",
          "number": 8,
          "startedAt": "2026-10-16T13:08:32Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:08:40Z",
          "conclusion": "success",
          "name": "Install pinned Rust toolchain via rustup",
          "number": 9,
          "startedAt": "2026-10-16T13:08:32Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:08:40Z",
          "conclusion": "success",
          "name": "Verify Go toolchain identity",
          "number": 10,
          "startedAt": "2026-10-16T13:08:40Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:08:40Z",
          "conclusion": "success",
          "name": "Verify versioned Go install compatibility",
          "number": 11,
          "startedAt": "2026-10-16T13:08:40Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:08:41Z",
          "conclusion": "success",
          "name": "gofmt check",
          "number": 12,
          "startedAt": "2026-10-16T13:08:40Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:08:52Z",
          "conclusion": "success",
          "name": "go vet",
          "number": 13,
          "startedAt": "2026-10-16T13:08:41Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:09:01Z",
          "conclusion": "success",
          "name": "Ledger consistency",
          "number": 14,
          "startedAt": "2026-10-16T13:08:52Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:28:35Z",
          "conclusion": "success",
          "name": "go test + platform-case gate",
          "number": 15,
          "startedAt": "2026-10-16T13:09:01Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:28:37Z",
          "conclusion": "success",
          "name": "Upload gate evidence",
          "number": 16,
          "startedAt": "2026-10-16T13:28:35Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:28:37Z",
          "conclusion": "success",
          "name": "Post Run actions/setup-node@v4",
          "number": 29,
          "startedAt": "2026-10-16T13:28:37Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:28:37Z",
          "conclusion": "success",
          "name": "Post Run actions/setup-go@v5",
          "number": 30,
          "startedAt": "2026-10-16T13:28:37Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:28:38Z",
          "conclusion": "success",
          "name": "Post Checkout authoritative protocol suite",
          "number": 31,
          "startedAt": "2026-10-16T13:28:37Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:28:39Z",
          "conclusion": "success",
          "name": "Post Run actions/checkout@v4",
          "number": 32,
          "startedAt": "2026-10-16T13:28:38Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:28:42Z",
          "conclusion": "success",
          "name": "Complete job",
          "number": 33,
          "startedAt": "2026-10-16T13:28:39Z",
          "status": "completed"
        }
      ],
      "url": "https://github.com/relux-works/curator/actions/runs/38904812207/job/116985608564"
    },
    {
      "completedAt": "2026-10-16T13:51:02Z",
      "conclusion": "success",
      "databaseId": 116985608683,
      "name": "Test (windows-latest)",
      "startedAt": "2026-10-16T13:02:22Z",
      "status": "completed",
      "steps": [
        {
          "completedAt": "2026-10-16T13:02:25Z",
          "conclusion": "success",
          "name": "Set up job",
          "number": 1,
          "startedAt": "2026-10-16T13:02:23Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:42Z",
          "conclusion": "success",
          "name": "Run actions/checkout@v4",
          "number": 2,
          "startedAt": "2026-10-16T13:02:25Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:45Z",
          "conclusion": "success",
          "name": "Checkout authoritative protocol suite",
          "number": 3,
          "startedAt": "2026-10-16T13:02:42Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:22Z",
          "conclusion": "success",
          "name": "Run actions/setup-go@v5",
          "number": 4,
          "startedAt": "2026-10-16T13:02:45Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:27Z",
          "conclusion": "success",
          "name": "Run actions/setup-node@v4",
          "number": 5,
          "startedAt": "2026-10-16T13:03:22Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:27Z",
          "conclusion": "success",
          "name": "Verify pnpm pin matches Go constant",
          "number": 6,
          "startedAt": "2026-10-16T13:03:27Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:34Z",
          "conclusion": "success",
          "name": "Install pinned pnpm via npm",
          "number": 7,
          "startedAt": "2026-10-16T13:03:27Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:34Z",
          "conclusion": "success",
          "name": "Verify rust pin matches Go constant",
          "number": 8,
          "startedAt": "2026-10-16T13:03:34Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:45Z",
          "conclusion": "success",
          "name": "Install pinned Rust toolchain via rustup",
          "number": 9,
          "startedAt": "2026-10-16T13:03:34Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:45Z",
          "conclusion": "success",
          "name": "Verify Go toolchain identity",
          "number": 10,
          "startedAt": "2026-10-16T13:03:45Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:45Z",
          "conclusion": "success",
          "name": "Verify versioned Go install compatibility",
          "number": 11,
          "startedAt": "2026-10-16T13:03:45Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:45Z",
          "conclusion": "skipped",
          "name": "gofmt check",
          "number": 12,
          "startedAt": "2026-10-16T13:03:45Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:04:04Z",
          "conclusion": "success",
          "name": "go vet",
          "number": 13,
          "startedAt": "2026-10-16T13:03:45Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:04:51Z",
          "conclusion": "success",
          "name": "Ledger consistency",
          "number": 14,
          "startedAt": "2026-10-16T13:04:04Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:50:53Z",
          "conclusion": "success",
          "name": "go test + platform-case gate",
          "number": 15,
          "startedAt": "2026-10-16T13:04:51Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:50:55Z",
          "conclusion": "success",
          "name": "Upload gate evidence",
          "number": 16,
          "startedAt": "2026-10-16T13:50:53Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:50:55Z",
          "conclusion": "success",
          "name": "Post Run actions/setup-node@v4",
          "number": 29,
          "startedAt": "2026-10-16T13:50:55Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:50:55Z",
          "conclusion": "success",
          "name": "Post Run actions/setup-go@v5",
          "number": 30,
          "startedAt": "2026-10-16T13:50:55Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:50:58Z",
          "conclusion": "success",
          "name": "Post Checkout authoritative protocol suite",
          "number": 31,
          "startedAt": "2026-10-16T13:50:55Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:51:00Z",
          "conclusion": "success",
          "name": "Post Run actions/checkout@v4",
          "number": 32,
          "startedAt": "2026-10-16T13:50:58Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:51:00Z",
          "conclusion": "success",
          "name": "Complete job",
          "number": 33,
          "startedAt": "2026-10-16T13:51:00Z",
          "status": "completed"
        }
      ],
      "url": "https://github.com/relux-works/curator/actions/runs/38904812207/job/116985608683"
    },
    {
      "completedAt": "2026-10-16T13:03:10Z",
      "conclusion": "success",
      "databaseId": 116985608706,
      "name": "Go 1.26.0 driver (ubuntu-latest)",
      "startedAt": "2026-10-16T13:02:22Z",
      "status": "completed",
      "steps": [
        {
          "completedAt": "2026-10-16T13:02:24Z",
          "conclusion": "success",
          "name": "Set up job",
          "number": 1,
          "startedAt": "2026-10-16T13:02:23Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:31Z",
          "conclusion": "success",
          "name": "Run actions/checkout@v4",
          "number": 2,
          "startedAt": "2026-10-16T13:02:24Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:32Z",
          "conclusion": "success",
          "name": "Checkout authoritative protocol suite",
          "number": 3,
          "startedAt": "2026-10-16T13:02:31Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:47Z",
          "conclusion": "success",
          "name": "Run actions/setup-go@v5",
          "number": 4,
          "startedAt": "2026-10-16T13:02:32Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:48Z",
          "conclusion": "success",
          "name": "Verify family toolchain identity",
          "number": 5,
          "startedAt": "2026-10-16T13:02:47Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:52Z",
          "conclusion": "success",
          "name": "Driver qualification and reproducibility",
          "number": 6,
          "startedAt": "2026-10-16T13:02:48Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:54Z",
          "conclusion": "success",
          "name": "Driver vet",
          "number": 7,
          "startedAt": "2026-10-16T13:02:52Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:07Z",
          "conclusion": "success",
          "name": "Build manager",
          "number": 8,
          "startedAt": "2026-10-16T13:02:54Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:07Z",
          "conclusion": "success",
          "name": "Post Run actions/setup-go@v5",
          "number": 14,
          "startedAt": "2026-10-16T13:03:07Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:07Z",
          "conclusion": "success",
          "name": "Post Checkout authoritative protocol suite",
          "number": 15,
          "startedAt": "2026-10-16T13:03:07Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:07Z",
          "conclusion": "success",
          "name": "Post Run actions/checkout@v4",
          "number": 16,
          "startedAt": "2026-10-16T13:03:07Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:07Z",
          "conclusion": "success",
          "name": "Complete job",
          "number": 17,
          "startedAt": "2026-10-16T13:03:07Z",
          "status": "completed"
        }
      ],
      "url": "https://github.com/relux-works/curator/actions/runs/38904812207/job/116985608706"
    },
    {
      "completedAt": "2026-10-16T13:05:22Z",
      "conclusion": "success",
      "databaseId": 116985608754,
      "name": "Go 1.26.0 driver (macos-latest)",
      "startedAt": "2026-10-16T13:03:00Z",
      "status": "completed",
      "steps": [
        {
          "completedAt": "2026-10-16T13:03:02Z",
          "conclusion": "success",
          "name": "Set up job",
          "number": 1,
          "startedAt": "2026-10-16T13:03:01Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:13Z",
          "conclusion": "success",
          "name": "Run actions/checkout@v4",
          "number": 2,
          "startedAt": "2026-10-16T13:03:02Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:15Z",
          "conclusion": "success",
          "name": "Checkout authoritative protocol suite",
          "number": 3,
          "startedAt": "2026-10-16T13:03:13Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:39Z",
          "conclusion": "success",
          "name": "Run actions/setup-go@v5",
          "number": 4,
          "startedAt": "2026-10-16T13:03:15Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:40Z",
          "conclusion": "success",
          "name": "Verify family toolchain identity",
          "number": 5,
          "startedAt": "2026-10-16T13:03:39Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:05:00Z",
          "conclusion": "success",
          "name": "Driver qualification and reproducibility",
          "number": 6,
          "startedAt": "2026-10-16T13:03:40Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:05:02Z",
          "conclusion": "success",
          "name": "Driver vet",
          "number": 7,
          "startedAt": "2026-10-16T13:05:00Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:05:16Z",
          "conclusion": "success",
          "name": "Build manager",
          "number": 8,
          "startedAt": "2026-10-16T13:05:02Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:05:16Z",
          "conclusion": "success",
          "name": "Post Run actions/setup-go@v5",
          "number": 14,
          "startedAt": "2026-10-16T13:05:16Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:05:17Z",
          "conclusion": "success",
          "name": "Post Checkout authoritative protocol suite",
          "number": 15,
          "startedAt": "2026-10-16T13:05:16Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:05:17Z",
          "conclusion": "success",
          "name": "Post Run actions/checkout@v4",
          "number": 16,
          "startedAt": "2026-10-16T13:05:17Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:05:19Z",
          "conclusion": "success",
          "name": "Complete job",
          "number": 17,
          "startedAt": "2026-10-16T13:05:17Z",
          "status": "completed"
        }
      ],
      "url": "https://github.com/relux-works/curator/actions/runs/38904812207/job/116985608754"
    },
    {
      "completedAt": "2026-10-16T13:03:06Z",
      "conclusion": "success",
      "databaseId": 116985608769,
      "name": "Go 1.25.5 driver (ubuntu-latest)",
      "startedAt": "2026-10-16T13:02:22Z",
      "status": "completed",
      "steps": [
        {
          "completedAt": "2026-10-16T13:02:23Z",
          "conclusion": "success",
          "name": "Set up job",
          "number": 1,
          "startedAt": "2026-10-16T13:02:23Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:30Z",
          "conclusion": "success",
          "name": "Run actions/checkout@v4",
          "number": 2,
          "startedAt": "2026-10-16T13:02:23Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:31Z",
          "conclusion": "success",
          "name": "Checkout authoritative protocol suite",
          "number": 3,
          "startedAt": "2026-10-16T13:02:30Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:40Z",
          "conclusion": "success",
          "name": "Run actions/setup-go@v5",
          "number": 4,
          "startedAt": "2026-10-16T13:02:31Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:40Z",
          "conclusion": "success",
          "name": "Verify family toolchain identity",
          "number": 5,
          "startedAt": "2026-10-16T13:02:40Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:47Z",
          "conclusion": "success",
          "name": "Driver qualification and reproducibility",
          "number": 6,
          "startedAt": "2026-10-16T13:02:40Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:49Z",
          "conclusion": "success",
          "name": "Driver vet",
          "number": 7,
          "startedAt": "2026-10-16T13:02:47Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:03Z",
          "conclusion": "success",
          "name": "Build manager",
          "number": 8,
          "startedAt": "2026-10-16T13:02:49Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:03Z",
          "conclusion": "success",
          "name": "Post Run actions/setup-go@v5",
          "number": 14,
          "startedAt": "2026-10-16T13:03:03Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:04Z",
          "conclusion": "success",
          "name": "Post Checkout authoritative protocol suite",
          "number": 15,
          "startedAt": "2026-10-16T13:03:03Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:04Z",
          "conclusion": "success",
          "name": "Post Run actions/checkout@v4",
          "number": 16,
          "startedAt": "2026-10-16T13:03:04Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:04Z",
          "conclusion": "success",
          "name": "Complete job",
          "number": 17,
          "startedAt": "2026-10-16T13:03:04Z",
          "status": "completed"
        }
      ],
      "url": "https://github.com/relux-works/curator/actions/runs/38904812207/job/116985608769"
    },
    {
      "completedAt": "2026-10-16T13:05:12Z",
      "conclusion": "success",
      "databaseId": 116985608830,
      "name": "Go 1.27.1 driver (windows-latest)",
      "startedAt": "2026-10-16T13:02:23Z",
      "status": "completed",
      "steps": [
        {
          "completedAt": "2026-10-16T13:02:25Z",
          "conclusion": "success",
          "name": "Set up job",
          "number": 1,
          "startedAt": "2026-10-16T13:02:24Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:36Z",
          "conclusion": "success",
          "name": "Run actions/checkout@v4",
          "number": 2,
          "startedAt": "2026-10-16T13:02:25Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:02:39Z",
          "conclusion": "success",
          "name": "Checkout authoritative protocol suite",
          "number": 3,
          "startedAt": "2026-10-16T13:02:36Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:06Z",
          "conclusion": "success",
          "name": "Run actions/setup-go@v5",
          "number": 4,
          "startedAt": "2026-10-16T13:02:39Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:06Z",
          "conclusion": "success",
          "name": "Verify family toolchain identity",
          "number": 5,
          "startedAt": "2026-10-16T13:03:06Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:04:49Z",
          "conclusion": "success",
          "name": "Driver qualification and reproducibility",
          "number": 6,
          "startedAt": "2026-10-16T13:03:06Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:04:52Z",
          "conclusion": "success",
          "name": "Driver vet",
          "number": 7,
          "startedAt": "2026-10-16T13:04:49Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:05:06Z",
          "conclusion": "success",
          "name": "Build manager",
          "number": 8,
          "startedAt": "2026-10-16T13:04:52Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:05:06Z",
          "conclusion": "success",
          "name": "Post Run actions/setup-go@v5",
          "number": 14,
          "startedAt": "2026-10-16T13:05:06Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:05:08Z",
          "conclusion": "success",
          "name": "Post Checkout authoritative protocol suite",
          "number": 15,
          "startedAt": "2026-10-16T13:05:06Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:05:10Z",
          "conclusion": "success",
          "name": "Post Run actions/checkout@v4",
          "number": 16,
          "startedAt": "2026-10-16T13:05:08Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:05:10Z",
          "conclusion": "success",
          "name": "Complete job",
          "number": 17,
          "startedAt": "2026-10-16T13:05:10Z",
          "status": "completed"
        }
      ],
      "url": "https://github.com/relux-works/curator/actions/runs/38904812207/job/116985608830"
    },
    {
      "completedAt": "2026-10-16T13:07:38Z",
      "conclusion": "success",
      "databaseId": 116985608867,
      "name": "Go 1.25.5 driver (macos-latest)",
      "startedAt": "2026-10-16T13:05:28Z",
      "status": "completed",
      "steps": [
        {
          "completedAt": "2026-10-16T13:05:30Z",
          "conclusion": "success",
          "name": "Set up job",
          "number": 1,
          "startedAt": "2026-10-16T13:05:28Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:05:40Z",
          "conclusion": "success",
          "name": "Run actions/checkout@v4",
          "number": 2,
          "startedAt": "2026-10-16T13:05:30Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:05:42Z",
          "conclusion": "success",
          "name": "Checkout authoritative protocol suite",
          "number": 3,
          "startedAt": "2026-10-16T13:05:40Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:06:01Z",
          "conclusion": "success",
          "name": "Run actions/setup-go@v5",
          "number": 4,
          "startedAt": "2026-10-16T13:05:42Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:06:01Z",
          "conclusion": "success",
          "name": "Verify family toolchain identity",
          "number": 5,
          "startedAt": "2026-10-16T13:06:01Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:07:12Z",
          "conclusion": "success",
          "name": "Driver qualification and reproducibility",
          "number": 6,
          "startedAt": "2026-10-16T13:06:01Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:07:15Z",
          "conclusion": "success",
          "name": "Driver vet",
          "number": 7,
          "startedAt": "2026-10-16T13:07:12Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:07:31Z",
          "conclusion": "success",
          "name": "Build manager",
          "number": 8,
          "startedAt": "2026-10-16T13:07:15Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:07:32Z",
          "conclusion": "success",
          "name": "Post Run actions/setup-go@v5",
          "number": 14,
          "startedAt": "2026-10-16T13:07:31Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:07:32Z",
          "conclusion": "success",
          "name": "Post Checkout authoritative protocol suite",
          "number": 15,
          "startedAt": "2026-10-16T13:07:32Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:07:33Z",
          "conclusion": "success",
          "name": "Post Run actions/checkout@v4",
          "number": 16,
          "startedAt": "2026-10-16T13:07:32Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:07:35Z",
          "conclusion": "success",
          "name": "Complete job",
          "number": 17,
          "startedAt": "2026-10-16T13:07:33Z",
          "status": "completed"
        }
      ],
      "url": "https://github.com/relux-works/curator/actions/runs/38904812207/job/116985608867"
    },
    {
      "completedAt": "2026-10-16T13:06:36Z",
      "conclusion": "success",
      "databaseId": 116985608951,
      "name": "Go 1.26.0 driver (windows-latest)",
      "startedAt": "2026-10-16T13:02:51Z",
      "status": "completed",
      "steps": [
        {
          "completedAt": "2026-10-16T13:02:53Z",
          "conclusion": "success",
          "name": "Set up job",
          "number": 1,
          "startedAt": "2026-10-16T13:02:52Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:06Z",
          "conclusion": "success",
          "name": "Run actions/checkout@v4",
          "number": 2,
          "startedAt": "2026-10-16T13:02:53Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:10Z",
          "conclusion": "success",
          "name": "Checkout authoritative protocol suite",
          "number": 3,
          "startedAt": "2026-10-16T13:03:06Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:42Z",
          "conclusion": "success",
          "name": "Run actions/setup-go@v5",
          "number": 4,
          "startedAt": "2026-10-16T13:03:10Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:42Z",
          "conclusion": "success",
          "name": "Verify family toolchain identity",
          "number": 5,
          "startedAt": "2026-10-16T13:03:42Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:06:03Z",
          "conclusion": "success",
          "name": "Driver qualification and reproducibility",
          "number": 6,
          "startedAt": "2026-10-16T13:03:42Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:06:08Z",
          "conclusion": "success",
          "name": "Driver vet",
          "number": 7,
          "startedAt": "2026-10-16T13:06:03Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:06:29Z",
          "conclusion": "success",
          "name": "Build manager",
          "number": 8,
          "startedAt": "2026-10-16T13:06:08Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:06:29Z",
          "conclusion": "success",
          "name": "Post Run actions/setup-go@v5",
          "number": 14,
          "startedAt": "2026-10-16T13:06:29Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:06:32Z",
          "conclusion": "success",
          "name": "Post Checkout authoritative protocol suite",
          "number": 15,
          "startedAt": "2026-10-16T13:06:29Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:06:34Z",
          "conclusion": "success",
          "name": "Post Run actions/checkout@v4",
          "number": 16,
          "startedAt": "2026-10-16T13:06:32Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:06:34Z",
          "conclusion": "success",
          "name": "Complete job",
          "number": 17,
          "startedAt": "2026-10-16T13:06:34Z",
          "status": "completed"
        }
      ],
      "url": "https://github.com/relux-works/curator/actions/runs/38904812207/job/116985608951"
    },
    {
      "completedAt": "2026-10-16T13:11:15Z",
      "conclusion": "success",
      "databaseId": 116985609008,
      "name": "Go 1.27.1 driver (macos-latest)",
      "startedAt": "2026-10-16T13:08:21Z",
      "status": "completed",
      "steps": [
        {
          "completedAt": "2026-10-16T13:08:24Z",
          "conclusion": "success",
          "name": "Set up job",
          "number": 1,
          "startedAt": "2026-10-16T13:08:22Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:08:35Z",
          "conclusion": "success",
          "name": "Run actions/checkout@v4",
          "number": 2,
          "startedAt": "2026-10-16T13:08:24Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:08:37Z",
          "conclusion": "success",
          "name": "Checkout authoritative protocol suite",
          "number": 3,
          "startedAt": "2026-10-16T13:08:35Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:09:09Z",
          "conclusion": "success",
          "name": "Run actions/setup-go@v5",
          "number": 4,
          "startedAt": "2026-10-16T13:08:37Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:09:09Z",
          "conclusion": "success",
          "name": "Verify family toolchain identity",
          "number": 5,
          "startedAt": "2026-10-16T13:09:09Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:10:49Z",
          "conclusion": "success",
          "name": "Driver qualification and reproducibility",
          "number": 6,
          "startedAt": "2026-10-16T13:09:09Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:10:52Z",
          "conclusion": "success",
          "name": "Driver vet",
          "number": 7,
          "startedAt": "2026-10-16T13:10:49Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:11:08Z",
          "conclusion": "success",
          "name": "Build manager",
          "number": 8,
          "startedAt": "2026-10-16T13:10:52Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:11:08Z",
          "conclusion": "success",
          "name": "Post Run actions/setup-go@v5",
          "number": 14,
          "startedAt": "2026-10-16T13:11:08Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:11:09Z",
          "conclusion": "success",
          "name": "Post Checkout authoritative protocol suite",
          "number": 15,
          "startedAt": "2026-10-16T13:11:08Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:11:10Z",
          "conclusion": "success",
          "name": "Post Run actions/checkout@v4",
          "number": 16,
          "startedAt": "2026-10-16T13:11:09Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:11:11Z",
          "conclusion": "success",
          "name": "Complete job",
          "number": 17,
          "startedAt": "2026-10-16T13:11:10Z",
          "status": "completed"
        }
      ],
      "url": "https://github.com/relux-works/curator/actions/runs/38904812207/job/116985609008"
    },
    {
      "completedAt": "2026-10-16T13:03:45Z",
      "conclusion": "success",
      "databaseId": 116985609049,
      "name": "Go 1.27.1 driver (ubuntu-latest)",
      "startedAt": "2026-10-16T13:03:08Z",
      "status": "completed",
      "steps": [
        {
          "completedAt": "2026-10-16T13:03:10Z",
          "conclusion": "success",
          "name": "Set up job",
          "number": 1,
          "startedAt": "2026-10-16T13:03:09Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:13Z",
          "conclusion": "success",
          "name": "Run actions/checkout@v4",
          "number": 2,
          "startedAt": "2026-10-16T13:03:10Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:14Z",
          "conclusion": "success",
          "name": "Checkout authoritative protocol suite",
          "number": 3,
          "startedAt": "2026-10-16T13:03:13Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:24Z",
          "conclusion": "success",
          "name": "Run actions/setup-go@v5",
          "number": 4,
          "startedAt": "2026-10-16T13:03:14Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:24Z",
          "conclusion": "success",
          "name": "Verify family toolchain identity",
          "number": 5,
          "startedAt": "2026-10-16T13:03:24Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:28Z",
          "conclusion": "success",
          "name": "Driver qualification and reproducibility",
          "number": 6,
          "startedAt": "2026-10-16T13:03:24Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:29Z",
          "conclusion": "success",
          "name": "Driver vet",
          "number": 7,
          "startedAt": "2026-10-16T13:03:28Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:42Z",
          "conclusion": "success",
          "name": "Build manager",
          "number": 8,
          "startedAt": "2026-10-16T13:03:29Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:42Z",
          "conclusion": "success",
          "name": "Post Run actions/setup-go@v5",
          "number": 14,
          "startedAt": "2026-10-16T13:03:42Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:42Z",
          "conclusion": "success",
          "name": "Post Checkout authoritative protocol suite",
          "number": 15,
          "startedAt": "2026-10-16T13:03:42Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:43Z",
          "conclusion": "success",
          "name": "Post Run actions/checkout@v4",
          "number": 16,
          "startedAt": "2026-10-16T13:03:42Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:43Z",
          "conclusion": "success",
          "name": "Complete job",
          "number": 17,
          "startedAt": "2026-10-16T13:03:43Z",
          "status": "completed"
        }
      ],
      "url": "https://github.com/relux-works/curator/actions/runs/38904812207/job/116985609049"
    },
    {
      "completedAt": "2026-10-16T13:06:33Z",
      "conclusion": "success",
      "databaseId": 116985609130,
      "name": "Go 1.25.5 driver (windows-latest)",
      "startedAt": "2026-10-16T13:02:52Z",
      "status": "completed",
      "steps": [
        {
          "completedAt": "2026-10-16T13:02:53Z",
          "conclusion": "success",
          "name": "Set up job",
          "number": 1,
          "startedAt": "2026-10-16T13:02:53Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:05Z",
          "conclusion": "success",
          "name": "Run actions/checkout@v4",
          "number": 2,
          "startedAt": "2026-10-16T13:02:53Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:08Z",
          "conclusion": "success",
          "name": "Checkout authoritative protocol suite",
          "number": 3,
          "startedAt": "2026-10-16T13:03:05Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:52Z",
          "conclusion": "success",
          "name": "Run actions/setup-go@v5",
          "number": 4,
          "startedAt": "2026-10-16T13:03:08Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:03:52Z",
          "conclusion": "success",
          "name": "Verify family toolchain identity",
          "number": 5,
          "startedAt": "2026-10-16T13:03:52Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:06:06Z",
          "conclusion": "success",
          "name": "Driver qualification and reproducibility",
          "number": 6,
          "startedAt": "2026-10-16T13:03:52Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:06:09Z",
          "conclusion": "success",
          "name": "Driver vet",
          "number": 7,
          "startedAt": "2026-10-16T13:06:06Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:06:27Z",
          "conclusion": "success",
          "name": "Build manager",
          "number": 8,
          "startedAt": "2026-10-16T13:06:09Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:06:27Z",
          "conclusion": "success",
          "name": "Post Run actions/setup-go@v5",
          "number": 14,
          "startedAt": "2026-10-16T13:06:27Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:06:29Z",
          "conclusion": "success",
          "name": "Post Checkout authoritative protocol suite",
          "number": 15,
          "startedAt": "2026-10-16T13:06:27Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:06:31Z",
          "conclusion": "success",
          "name": "Post Run actions/checkout@v4",
          "number": 16,
          "startedAt": "2026-10-16T13:06:29Z",
          "status": "completed"
        },
        {
          "completedAt": "2026-10-16T13:06:31Z",
          "conclusion": "success",
          "name": "Complete job",
          "number": 17,
          "startedAt": "2026-10-16T13:06:31Z",
          "status": "completed"
        }
      ],
      "url": "https://github.com/relux-works/curator/actions/runs/38904812207/job/116985609130"
    },
    {
      "completedAt": "2026-10-16T13:02:20Z",
      "conclusion": "skipped",
      "databaseId": 116985610053,
      "name": "Candidate suite (${{ matrix.os }})",
      "startedAt": "2026-10-16T13:02:20Z",
      "status": "completed",
      "steps": [],
      "url": "https://github.com/relux-works/curator/actions/runs/38904812207/job/116985610053"
    },
    {
      "completedAt": "2026-10-16T13:02:20Z",
      "conclusion": "skipped",
      "databaseId": 116985610209,
      "name": "Test (optional-self-hosted)",
      "startedAt": "2026-10-16T13:02:20Z",
      "status": "completed",
      "steps": [],
      "url": "https://github.com/relux-works/curator/actions/runs/38904812207/job/116985610209"
    }
  ],
  "status": "completed",
  "url": "https://github.com/relux-works/curator/actions/runs/38904812207"
}
```
