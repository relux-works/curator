# Skillfile MCP stage-A consumer inventory

Spec pin: `dbea3ee0bb8716cb4fcef758609717c10557ceff`. The vendored bytes are checked against `MANIFEST.sha256`; released corpus pins remain unchanged.

`TestSkillfileMCPConsumerVectors` consumes production parser, normalizer, digest, admission and passive carrier readers. Planning vectors exercise only raw Skillfile parsing and supplied package-manifest parsing. They do not establish acquisition, machine-policy, audit, transaction, native-writer or ownership-action behavior. Real Git and CLI fixtures in `internal/envprofile/project_mcp_test.go` and `cmd/curator/project_mcp_test.go` exercise the production effects.

| Slice | Cases | Evidence |
| --- | ---: | --- |
| normalization | 2 | Complete vector verdict through production consumer |
| admission | 2 | Complete vector verdict through production consumer |
| carrier | 28 | Complete vector verdict through production consumer |
| Planning parser projection | 115 | Parser/admission verdict only; other model facts are outside this projection |
| Deferred native/action model | 13 | Subsequent native-writer task |
| Separate component upgrade model | 5 | Manager schema dispatch is covered at production entrypoints |

## Exact model cases not consumed

The orchestrator assigned native writer and ownership-action model vectors to the subsequent writer task. No synthetic native planner or Muse writer is introduced here.

- `foreign-entry`: native preimage, native writer, or ownership-action authorization model is outside stage A.
- `unreadable-preimage`: native preimage, native writer, or ownership-action authorization model is outside stage A.
- `action-no-authority`: native preimage, native writer, or ownership-action authorization model is outside stage A.
- `action-policy-denied`: native preimage, native writer, or ownership-action authorization model is outside stage A.
- `action-authorized`: native preimage, native writer, or ownership-action authorization model is outside stage A.
- `action-replay-project_id`: native preimage, native writer, or ownership-action authorization model is outside stage A.
- `action-replay-name`: native preimage, native writer, or ownership-action authorization model is outside stage A.
- `action-replay-server_sha256`: native preimage, native writer, or ownership-action authorization model is outside stage A.
- `action-replay-content_sha256`: native preimage, native writer, or ownership-action authorization model is outside stage A.
- `action-replay-policy_sha256`: native preimage, native writer, or ownership-action authorization model is outside stage A.
- `action-replay-operation`: native preimage, native writer, or ownership-action authorization model is outside stage A.
- `action-replay-state`: native preimage, native writer, or ownership-action authorization model is outside stage A.
- `action-repository-record`: native preimage, native writer, or ownership-action authorization model is outside stage A.

The component upgrade model refuses schema versions that the full manager already supports. The manager retains schema 1/2 behavior, accepts project schema 3, refuses global schema 3, and refuses future versions. Those production entrypoints have separate tests.

- `global-upgrade`: component upgrade verdict is not substituted for manager version dispatch.
- `old-reader-upgrade`: component upgrade verdict is not substituted for manager version dispatch.
- `version-1`: component upgrade verdict is not substituted for manager version dispatch.
- `version-2`: component upgrade verdict is not substituted for manager version dispatch.
- `version-4`: component upgrade verdict is not substituted for manager version dispatch.

`snapshot-absent` has no supplied package bytes. Its raw Skillfile is parsed without inventing a package; missing protected snapshots are rejected by the real replay fixture. `manifest-missing` supplies a package observation without its manifest and therefore exercises the actual package parser rejection.

Existing declaration-level source errors absent from the published closed row-code enum keep their existing diagnostic spellings. They are not converted to invented pair refusal codes; known pairs remain unsupported and never become successful. This enum omission was reported to the spec owner.
