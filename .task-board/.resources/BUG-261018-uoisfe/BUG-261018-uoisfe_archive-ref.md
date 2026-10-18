# Local raw validation archive

The public board refuses binary resources (`resource_binary_refused`; upload exit 1). No binary archive was attached. Following its supported text-reference path, the complete sanitized archive is retained outside Git at `.temp/BUG-261018-uoisfe/BUG-261018-uoisfe_validation-logs.tar.gz`, relative to the managed Story worktree. Preserve this ignored artifact until review/integration consumes it.

- Candidate: `b32823ac920d8b0485c4a7a7111b48fff8e3ca2b`.
- SHA-256: `ca67a410e3e171a6180fe6d03f2d7ef54c1042c77004fb19295f52fb8f0c2209`.
- Size: 2,086,483 bytes; 191 text files inside.
- Per-file hashes: `BUG-261018-uoisfe_archive-manifest.json`.
- Parse/privacy audit: `BUG-261018-uoisfe_archive-audit.json` (exit 0).

The full accepted Go event stream is separately attached as plaintext `BUG-261018-uoisfe_validation-stream.jsonl`, byte-identical to the sanitized stream checked by the unchanged platform gate (exit 0). The text metadata bundle contains command exits, masks, fixture and tool identities, hosted job verdicts, current Windows controls, local gate reports and scripts. Historical/raw command streams remain in the hash-bound local archive; no missing or truncated stream is used as passing evidence.
