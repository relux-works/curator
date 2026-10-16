# Revision 4 reviewer full gate

Command: `sh scripts/remote-gate.sh`, after loading the configured agent Git environment. Actual process exit: **0**.

Signed snapshot: `98452a23e13ca4521353cbe3e26207755dec2170`; signature verified `G`. Tree: `24299c092bc02d004545ff7103ffc192b541db37`.

[Hosted run](https://github.com/relux-works/curator/actions/runs/38972219571): success. 20 enabled jobs passed; 2 conditional jobs skipped, not claimed as runtime qualification.

```json
[
  {
    "name": "Race (ubuntu-latest)",
    "conclusion": "success"
  },
  {
    "name": "Race (macos-latest)",
    "conclusion": "success"
  },
  {
    "name": "Gate self-test (windows-latest)",
    "conclusion": "success"
  },
  {
    "name": "Gate self-test (ubuntu-latest)",
    "conclusion": "success"
  },
  {
    "name": "Gate self-test (macos-latest)",
    "conclusion": "success"
  },
  {
    "name": "Test (macos-latest)",
    "conclusion": "success"
  },
  {
    "name": "Naming gate",
    "conclusion": "success"
  },
  {
    "name": "Test (ubuntu-latest)",
    "conclusion": "success"
  },
  {
    "name": "Go 1.25.5 driver (ubuntu-latest)",
    "conclusion": "success"
  },
  {
    "name": "Lint",
    "conclusion": "success"
  },
  {
    "name": "Interop conformance gate",
    "conclusion": "success"
  },
  {
    "name": "Go 1.27.1 driver (macos-latest)",
    "conclusion": "success"
  },
  {
    "name": "Go 1.25.5 driver (windows-latest)",
    "conclusion": "success"
  },
  {
    "name": "Go 1.26.0 driver (macos-latest)",
    "conclusion": "success"
  },
  {
    "name": "Go 1.26.0 driver (windows-latest)",
    "conclusion": "success"
  },
  {
    "name": "Test (windows-latest)",
    "conclusion": "success"
  },
  {
    "name": "Go 1.27.1 driver (ubuntu-latest)",
    "conclusion": "success"
  },
  {
    "name": "Go 1.25.5 driver (macos-latest)",
    "conclusion": "success"
  },
  {
    "name": "Go 1.27.1 driver (windows-latest)",
    "conclusion": "success"
  },
  {
    "name": "Go 1.26.0 driver (ubuntu-latest)",
    "conclusion": "success"
  },
  {
    "name": "Optional self-hosted test",
    "conclusion": "skipped"
  },
  {
    "name": "Candidate suite (${{ matrix.os }})",
    "conclusion": "skipped"
  }
]
```

All 17 working candidate blobs still match after completion. Story HEAD equals the review base and the real index is empty. No product changes made by the reviewer.
