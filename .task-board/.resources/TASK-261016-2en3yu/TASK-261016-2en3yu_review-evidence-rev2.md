# Revision 2 independent review evidence
Candidate tree: ddde39b56122ce8b78ecb49ab410914ab8693171.
Commands and scope are recorded in the verdict. Synthetic test paths and host-specific signing identity are redacted.

## identity.json

```text
{
  "base": "dccac3199aa5c543c589c129f5ef2bc437523a2a",
  "candidate_tree": "ddde39b56122ce8b78ecb49ab410914ab8693171",
  "verified_paths": [
    "README.md",
    "cmd/curator/env.go",
    "cmd/curator/env_credential_store_test.go",
    "cmd/curator/envalias.go",
    "cmd/curator/envconfig.go",
    "docs/environment-config.md",
    "internal/config/config.go",
    "internal/config/credential_store.go",
    "internal/config/credential_store_test.go",
    "internal/config/environments.go",
    "internal/config/environments_test.go",
    "internal/envprofile/credential_policy.go",
    "internal/envprofile/credential_policy_test.go",
    "internal/envprofile/managed.go",
    "internal/envregistry/credential_store.go",
    "internal/envregistry/envregistry.go"
  ],
  "head_unchanged": true
}
```

## toolchain.log

```text
go version go1.25.5 darwin/arm64
gh version 2.97.0 (2026-07-31)
https://github.com/cli/cli/releases/tag/v2.97.0
Python 3.14.7

```

## rework.log

```text
ok  	github.com/relux-works/curator/cmd/curator	0.791s

```

## grammar.log

```text
=== RUN   TestCredentialStoreConfigGrammar
=== RUN   TestCredentialStoreConfigGrammar/empty
=== RUN   TestCredentialStoreConfigGrammar/all_selectors
=== RUN   TestCredentialStoreConfigGrammar/null_map
=== RUN   TestCredentialStoreConfigGrammar/array_map
=== RUN   TestCredentialStoreConfigGrammar/null_profile
=== RUN   TestCredentialStoreConfigGrammar/invalid_profile
=== RUN   TestCredentialStoreConfigGrammar/unknown_id
=== RUN   TestCredentialStoreConfigGrammar/alias_id
=== RUN   TestCredentialStoreConfigGrammar/null_selector
=== RUN   TestCredentialStoreConfigGrammar/unknown_selector
=== RUN   TestCredentialStoreConfigGrammar/wrong_type
--- PASS: TestCredentialStoreConfigGrammar (0.01s)
    --- PASS: TestCredentialStoreConfigGrammar/empty (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/all_selectors (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/null_map (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/array_map (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/null_profile (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/invalid_profile (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/unknown_id (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/alias_id (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/null_selector (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/unknown_selector (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/wrong_type (0.00s)
=== RUN   TestCredentialStoreConfigVersionBoundary
--- PASS: TestCredentialStoreConfigVersionBoundary (0.00s)
=== RUN   TestCredentialStoreConfigMachineOwnership
--- PASS: TestCredentialStoreConfigMachineOwnership (0.00s)
PASS
ok  	github.com/relux-works/curator/internal/config	0.292s

```

## selectors.log

```text
=== RUN   TestCredentialStoreCLIConfigRoundTrip
--- PASS: TestCredentialStoreCLIConfigRoundTrip (1.06s)
=== RUN   TestCredentialStoreCLIConfigRefusesMalformed
--- PASS: TestCredentialStoreCLIConfigRefusesMalformed (0.01s)
=== RUN   TestCredentialStoreCLIPromotion
--- PASS: TestCredentialStoreCLIPromotion (0.00s)
=== RUN   TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery
--- PASS: TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery (0.54s)
=== RUN   TestCredentialStoreCLIDottedProfileRouting
=== RUN   TestCredentialStoreCLIDottedProfileRouting/team/codex
=== RUN   TestCredentialStoreCLIDottedProfileRouting/team/codex_cli
=== RUN   TestCredentialStoreCLIDottedProfileRouting/team/claude
=== RUN   TestCredentialStoreCLIDottedProfileRouting/team/claude_code
=== RUN   TestCredentialStoreCLIDottedProfileRouting/team.prod/codex
=== RUN   TestCredentialStoreCLIDottedProfileRouting/team.prod/codex_cli
=== RUN   TestCredentialStoreCLIDottedProfileRouting/team.prod/claude
=== RUN   TestCredentialStoreCLIDottedProfileRouting/team.prod/claude_code
=== RUN   TestCredentialStoreCLIDottedProfileRouting/Team..prod/codex
=== RUN   TestCredentialStoreCLIDottedProfileRouting/Team..prod/codex_cli
=== RUN   TestCredentialStoreCLIDottedProfileRouting/Team..prod/claude
=== RUN   TestCredentialStoreCLIDottedProfileRouting/Team..prod/claude_code
=== RUN   TestCredentialStoreCLIDottedProfileRouting/codex.prod/codex
=== RUN   TestCredentialStoreCLIDottedProfileRouting/codex.prod/codex_cli
=== RUN   TestCredentialStoreCLIDottedProfileRouting/codex.prod/claude
=== RUN   TestCredentialStoreCLIDottedProfileRouting/codex.prod/claude_code
--- PASS: TestCredentialStoreCLIDottedProfileRouting (0.03s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team/codex (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team/codex_cli (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team/claude (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team/claude_code (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team.prod/codex (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team.prod/codex_cli (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team.prod/claude (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team.prod/claude_code (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/Team..prod/codex (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/Team..prod/codex_cli (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/Team..prod/claude (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/Team..prod/claude_code (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/codex.prod/codex (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/codex.prod/codex_cli (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/codex.prod/claude (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/codex.prod/claude_code (0.00s)
=== RUN   TestCredentialStoreCLIInvalidProfileRouting
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/.team/set
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/.team/show
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/.team/unset
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/team./set
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/team./show
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/team./unset
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/team/prod/set
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/team/prod/show
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/team/prod/unset
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/CON.prod/set
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/CON.prod/show
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/CON.prod/unset
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/ppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/set
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/ppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/show
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/ppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/unset
--- PASS: TestCredentialStoreCLIInvalidProfileRouting (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/.team/set (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/.team/show (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/.team/unset (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/team./set (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/team./show (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/team./unset (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/team/prod/set (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/team/prod/show (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/team/prod/unset (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/CON.prod/set (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/CON.prod/show (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/CON.prod/unset (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/ppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/set (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/ppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/show (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/ppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/unset (0.00s)
PASS
ok  	github.com/relux-works/curator/cmd/curator	1.893s
=== RUN   TestCredentialStorePolicyActivation
=== RUN   TestCredentialStorePolicyActivation/claude_code/file
=== RUN   TestCredentialStorePolicyActivation/claude_code/keychain
=== RUN   TestCredentialStorePolicyActivation/codex_cli/file
=== RUN   TestCredentialStorePolicyActivation/codex_cli/keychain
=== RUN   TestCredentialStorePolicyActivation/muse/file
=== RUN   TestCredentialStorePolicyActivation/muse/keychain
=== RUN   TestCredentialStorePolicyActivation/pi/file
=== RUN   TestCredentialStorePolicyActivation/pi/keychain
=== RUN   TestCredentialStorePolicyActivation/opencode/file
=== RUN   TestCredentialStorePolicyActivation/opencode/keychain
--- PASS: TestCredentialStorePolicyActivation (0.94s)
    --- PASS: TestCredentialStorePolicyActivation/claude_code/file (0.00s)
    --- PASS: TestCredentialStorePolicyActivation/claude_code/keychain (0.00s)
    --- PASS: TestCredentialStorePolicyActivation/codex_cli/file (0.00s)
    --- PASS: TestCredentialStorePolicyActivation/codex_cli/keychain (0.00s)
    --- PASS: TestCredentialStorePolicyActivation/muse/file (0.00s)
    --- PASS: TestCredentialStorePolicyActivation/muse/keychain (0.00s)
    --- PASS: TestCredentialStorePolicyActivation/pi/file (0.00s)
    --- PASS: TestCredentialStorePolicyActivation/pi/keychain (0.00s)
    --- PASS: TestCredentialStorePolicyActivation/opencode/file (0.00s)
    --- PASS: TestCredentialStorePolicyActivation/opencode/keychain (0.00s)
PASS
ok  	github.com/relux-works/curator/internal/envprofile	1.264s

```

## errors.log

```text
=== RUN   TestCredentialStorePolicyNativeInspection
=== RUN   TestCredentialStorePolicyNativeInspection/absent_key
=== RUN   TestCredentialStorePolicyNativeInspection/keyring
=== RUN   TestCredentialStorePolicyNativeInspection/auto
=== RUN   TestCredentialStorePolicyNativeInspection/unknown
=== RUN   TestCredentialStorePolicyNativeInspection/malformed
=== RUN   TestCredentialStorePolicyNativeInspection/permission
--- PASS: TestCredentialStorePolicyNativeInspection (0.19s)
    --- PASS: TestCredentialStorePolicyNativeInspection/absent_key (0.00s)
    --- PASS: TestCredentialStorePolicyNativeInspection/keyring (0.00s)
    --- PASS: TestCredentialStorePolicyNativeInspection/auto (0.00s)
    --- PASS: TestCredentialStorePolicyNativeInspection/unknown (0.00s)
    --- PASS: TestCredentialStorePolicyNativeInspection/malformed (0.00s)
    --- PASS: TestCredentialStorePolicyNativeInspection/permission (0.00s)
PASS
ok  	github.com/relux-works/curator/internal/envprofile	0.416s

```

## config-registry.log

```text
ok  	github.com/relux-works/curator/internal/config	0.497s
ok  	github.com/relux-works/curator/internal/envregistry	0.251s

```

## freehunt.log

```text
=== RUN   TestCredentialStoreCLIDottedProfileRouting
=== RUN   TestCredentialStoreCLIDottedProfileRouting/team/codex
=== RUN   TestCredentialStoreCLIDottedProfileRouting/team/codex_cli
=== RUN   TestCredentialStoreCLIDottedProfileRouting/team/claude
=== RUN   TestCredentialStoreCLIDottedProfileRouting/team/claude_code
=== RUN   TestCredentialStoreCLIDottedProfileRouting/team.prod/codex
=== RUN   TestCredentialStoreCLIDottedProfileRouting/team.prod/codex_cli
=== RUN   TestCredentialStoreCLIDottedProfileRouting/team.prod/claude
=== RUN   TestCredentialStoreCLIDottedProfileRouting/team.prod/claude_code
=== RUN   TestCredentialStoreCLIDottedProfileRouting/Team..prod/codex
=== RUN   TestCredentialStoreCLIDottedProfileRouting/Team..prod/codex_cli
=== RUN   TestCredentialStoreCLIDottedProfileRouting/Team..prod/claude
=== RUN   TestCredentialStoreCLIDottedProfileRouting/Team..prod/claude_code
=== RUN   TestCredentialStoreCLIDottedProfileRouting/codex.prod/codex
=== RUN   TestCredentialStoreCLIDottedProfileRouting/codex.prod/codex_cli
=== RUN   TestCredentialStoreCLIDottedProfileRouting/codex.prod/claude
=== RUN   TestCredentialStoreCLIDottedProfileRouting/codex.prod/claude_code
=== RUN   TestCredentialStoreCLIDottedProfileRouting/pppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/codex
=== RUN   TestCredentialStoreCLIDottedProfileRouting/pppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/codex_cli
=== RUN   TestCredentialStoreCLIDottedProfileRouting/pppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/claude
=== RUN   TestCredentialStoreCLIDottedProfileRouting/pppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/claude_code
=== RUN   TestCredentialStoreCLIDottedProfileRouting/a.bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/codex
=== RUN   TestCredentialStoreCLIDottedProfileRouting/a.bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/codex_cli
=== RUN   TestCredentialStoreCLIDottedProfileRouting/a.bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/claude
=== RUN   TestCredentialStoreCLIDottedProfileRouting/a.bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/claude_code
=== RUN   TestCredentialStoreCLIDottedProfileRouting/COM0.prod/codex
=== RUN   TestCredentialStoreCLIDottedProfileRouting/COM0.prod/codex_cli
=== RUN   TestCredentialStoreCLIDottedProfileRouting/COM0.prod/claude
=== RUN   TestCredentialStoreCLIDottedProfileRouting/COM0.prod/claude_code
=== RUN   TestCredentialStoreCLIDottedProfileRouting/a...b/codex
=== RUN   TestCredentialStoreCLIDottedProfileRouting/a...b/codex_cli
=== RUN   TestCredentialStoreCLIDottedProfileRouting/a...b/claude
=== RUN   TestCredentialStoreCLIDottedProfileRouting/a...b/claude_code
--- PASS: TestCredentialStoreCLIDottedProfileRouting (0.21s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team/codex (0.01s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team/codex_cli (0.04s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team/claude (0.04s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team/claude_code (0.01s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team.prod/codex (0.01s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team.prod/codex_cli (0.01s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team.prod/claude (0.01s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team.prod/claude_code (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/Team..prod/codex (0.01s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/Team..prod/codex_cli (0.01s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/Team..prod/claude (0.01s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/Team..prod/claude_code (0.01s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/codex.prod/codex (0.01s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/codex.prod/codex_cli (0.01s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/codex.prod/claude (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/codex.prod/claude_code (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/pppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/codex (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/pppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/codex_cli (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/pppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/claude (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/pppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/claude_code (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/a.bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/codex (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/a.bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/codex_cli (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/a.bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/claude (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/a.bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/claude_code (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/COM0.prod/codex (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/COM0.prod/codex_cli (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/COM0.prod/claude (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/COM0.prod/claude_code (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/a...b/codex (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/a...b/codex_cli (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/a...b/claude (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/a...b/claude_code (0.00s)
=== RUN   TestCredentialStoreCLIInvalidProfileRouting
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/.team/set
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/.team/show
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/.team/unset
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/team./set
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/team./show
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/team./unset
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/team/prod/set
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/team/prod/show
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/team/prod/unset
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/CON.prod/set
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/CON.prod/show
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/CON.prod/unset
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/ppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/set
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/ppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/show
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/ppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/unset
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/LPT1.prod/set
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/LPT1.prod/show
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/LPT1.prod/unset
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/com9.prod/set
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/com9.prod/show
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/com9.prod/unset
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/a:b/set
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/a:b/show
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/a:b/unset
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/a_b/set
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/a_b/show
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/a_b/unset
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/équipe/set
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/équipe/show
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/équipe/unset
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/a.bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/set
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/a.bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/show
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/a.bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/unset
--- PASS: TestCredentialStoreCLIInvalidProfileRouting (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/.team/set (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/.team/show (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/.team/unset (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/team./set (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/team./show (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/team./unset (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/team/prod/set (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/team/prod/show (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/team/prod/unset (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/CON.prod/set (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/CON.prod/show (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/CON.prod/unset (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/ppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/set (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/ppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/show (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/ppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/unset (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/LPT1.prod/set (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/LPT1.prod/show (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/LPT1.prod/unset (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/com9.prod/set (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/com9.prod/show (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/com9.prod/unset (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/a:b/set (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/a:b/show (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/a:b/unset (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/a_b/set (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/a_b/show (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/a_b/unset (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/équipe/set (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/équipe/show (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/équipe/unset (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/a.bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/set (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/a.bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/show (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/a.bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/unset (0.00s)
PASS
ok  	github.com/relux-works/curator/cmd/curator	0.578s

```

## mutant-results.json

```text
[
  {
    "name": "drop-codex-file",
    "exit_code": 1,
    "named_failures": [
      "--- FAIL: TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery (0.80s)"
    ]
  },
  {
    "name": "old-v3",
    "exit_code": 1,
    "named_failures": [
      "--- FAIL: TestCredentialStoreConfigVersionBoundary (0.00s)"
    ]
  },
  {
    "name": "permission-as-absent",
    "exit_code": 1,
    "named_failures": [
      "--- FAIL: TestCredentialStorePolicyNativeInspection (0.26s)",
      "    --- FAIL: TestCredentialStorePolicyNativeInspection/permission (0.00s)"
    ]
  },
  {
    "name": "persist-alias",
    "exit_code": 1,
    "named_failures": [
      "--- FAIL: TestCredentialStoreConfigGrammar (0.01s)",
      "    --- FAIL: TestCredentialStoreConfigGrammar/alias_id (0.00s)"
    ]
  },
  {
    "name": "profile-normalization",
    "exit_code": 1,
    "named_failures": [
      "--- FAIL: TestCredentialStoreCLIConfigRoundTrip (0.71s)"
    ]
  },
  {
    "name": "system-store",
    "exit_code": 1,
    "named_failures": [
      "--- FAIL: TestCredentialStoreConfigMachineOwnership (0.00s)"
    ]
  },
  {
    "name": "unknown-auto",
    "exit_code": 1,
    "named_failures": [
      "--- FAIL: TestCredentialStoreConfigGrammar (0.08s)",
      "    --- FAIL: TestCredentialStoreConfigGrammar/unknown_selector (0.00s)"
    ]
  },
  {
    "name": "dotted-profile-only",
    "exit_code": 1,
    "named_failures": [
      "--- FAIL: TestCredentialStoreCLIDottedProfileRouting (0.01s)",
      "    --- FAIL: TestCredentialStoreCLIDottedProfileRouting/team.prod/codex (0.00s)",
      "    --- FAIL: TestCredentialStoreCLIDottedProfileRouting/team.prod/codex_cli (0.00s)",
      "    --- FAIL: TestCredentialStoreCLIDottedProfileRouting/team.prod/claude (0.00s)",
      "    --- FAIL: TestCredentialStoreCLIDottedProfileRouting/team.prod/claude_code (0.00s)",
      "    --- FAIL: TestCredentialStoreCLIDottedProfileRouting/Team..prod/codex (0.00s)",
      "    --- FAIL: TestCredentialStoreCLIDottedProfileRouting/Team..prod/codex_cli (0.00s)",
      "    --- FAIL: TestCredentialStoreCLIDottedProfileRouting/Team..prod/claude (0.00s)",
      "    --- FAIL: TestCredentialStoreCLIDottedProfileRouting/Team..prod/claude_code (0.00s)",
      "    --- FAIL: TestCredentialStoreCLIDottedProfileRouting/codex.prod/codex (0.00s)",
      "    --- FAIL: TestCredentialStoreCLIDottedProfileRouting/codex.prod/codex_cli (0.00s)",
      "    --- FAIL: TestCredentialStoreCLIDottedProfileRouting/codex.prod/claude (0.00s)",
      "    --- FAIL: TestCredentialStoreCLIDottedProfileRouting/codex.prod/claude_code (0.00s)"
    ]
  },
  {
    "name": "reserved-profile-show",
    "exit_code": 1,
    "named_failures": [
      "--- FAIL: TestCredentialStoreCLIInvalidProfileRouting (0.00s)",
      "    --- FAIL: TestCredentialStoreCLIInvalidProfileRouting/CON.prod/show (0.00s)"
    ]
  }
]
```

## drop-codex-file (expected failure, exit 1)

```text
--- FAIL: TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery (0.80s)
    env_credential_store_test.go:168: unqualified file = 0 {"env":{"CODEX_HOME":"<test-temp>"},"environment":"codex_cli","fragment":"launch-env-fragment-v2","permissions":{"locked":false,"mode":"native","source":"default"},"precedence":{"placement":"winner-last","winner":"higher-weight"},"profile":{"lock_sha256":"e13ef5a9cd5f1afaedb4c29010a731b26caed36f798fb6396b14db3ea456ea9d","name":"acme"}}
         warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults
        warning: environment_tool_version_unverified: codex_cli detected 0.162.1, recorded 0.153.2
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	1.359s
FAIL

```

## old-v3 (expected failure, exit 1)

```text
--- FAIL: TestCredentialStoreConfigVersionBoundary (0.00s)
    credential_store_test.go:57: err = <nil>, want mention of "credential_store"
FAIL
FAIL	github.com/relux-works/curator/internal/config	0.284s
FAIL

```

## permission-as-absent (expected failure, exit 1)

```text
--- FAIL: TestCredentialStorePolicyNativeInspection (0.26s)
    --- FAIL: TestCredentialStorePolicyNativeInspection/permission (0.00s)
        credential_policy_test.go:81: Resolve = <nil>, codex_settings_unsupported: explicit Codex file storage requires qualified login/app-server/refresh behavior and strict successor fragment consumers; no native fallback; want environment_seed_unreadable
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	0.593s
FAIL

```

## persist-alias (expected failure, exit 1)

```text
--- FAIL: TestCredentialStoreConfigGrammar (0.01s)
    --- FAIL: TestCredentialStoreConfigGrammar/alias_id (0.00s)
        credential_store_test.go:40: Load = <nil>, want credential_store refusal
FAIL
FAIL	github.com/relux-works/curator/internal/config	0.294s
FAIL

```

## profile-normalization (expected failure, exit 1)

```text
--- FAIL: TestCredentialStoreCLIConfigRoundTrip (0.71s)
    env_credential_store_test.go:66: wrong profile/environment normalization: {
          "environments": {
            "credential_store": {
              "codex_cli": {
                "codex_cli": "file"
              }
            }
          },
          "projects": {},
          "schema_version": 5,
          "security_posture": "permissive",
          "skills_root": "<test-temp>"
        }
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	1.075s
FAIL

```

## system-store (expected failure, exit 1)

```text
--- FAIL: TestCredentialStoreConfigMachineOwnership (0.00s)
    credential_store_test.go:75: system store accepted: <nil>
FAIL
FAIL	github.com/relux-works/curator/internal/config	0.272s
FAIL

```

## unknown-auto (expected failure, exit 1)

```text
--- FAIL: TestCredentialStoreConfigGrammar (0.08s)
    --- FAIL: TestCredentialStoreConfigGrammar/unknown_selector (0.00s)
        credential_store_test.go:40: Load = <nil>, want credential_store refusal
FAIL
FAIL	github.com/relux-works/curator/internal/config	0.352s
FAIL

```

## dotted-profile-only (expected failure, exit 1)

```text
--- FAIL: TestCredentialStoreCLIDottedProfileRouting (0.01s)
    --- FAIL: TestCredentialStoreCLIDottedProfileRouting/team.prod/codex (0.00s)
        env_credential_store_test.go:237: set = 1  curator: environments.credential_store.team: unknown canonical environment "prod"
    --- FAIL: TestCredentialStoreCLIDottedProfileRouting/team.prod/codex_cli (0.00s)
        env_credential_store_test.go:237: set = 1  curator: environments.credential_store.team: unknown canonical environment "prod"
    --- FAIL: TestCredentialStoreCLIDottedProfileRouting/team.prod/claude (0.00s)
        env_credential_store_test.go:237: set = 1  curator: environments.credential_store.team: unknown canonical environment "prod"
    --- FAIL: TestCredentialStoreCLIDottedProfileRouting/team.prod/claude_code (0.00s)
        env_credential_store_test.go:237: set = 1  curator: environments.credential_store.team: unknown canonical environment "prod"
    --- FAIL: TestCredentialStoreCLIDottedProfileRouting/Team..prod/codex (0.00s)
        env_credential_store_test.go:237: set = 2  curator: knob: requires a section 12.1 knob name
    --- FAIL: TestCredentialStoreCLIDottedProfileRouting/Team..prod/codex_cli (0.00s)
        env_credential_store_test.go:237: set = 2  curator: knob: requires a section 12.1 knob name
    --- FAIL: TestCredentialStoreCLIDottedProfileRouting/Team..prod/claude (0.00s)
        env_credential_store_test.go:237: set = 2  curator: knob: requires a section 12.1 knob name
    --- FAIL: TestCredentialStoreCLIDottedProfileRouting/Team..prod/claude_code (0.00s)
        env_credential_store_test.go:237: set = 2  curator: knob: requires a section 12.1 knob name
    --- FAIL: TestCredentialStoreCLIDottedProfileRouting/codex.prod/codex (0.00s)
        env_credential_store_test.go:237: set = 1  curator: environments.credential_store.codex: unknown canonical environment "prod"
    --- FAIL: TestCredentialStoreCLIDottedProfileRouting/codex.prod/codex_cli (0.00s)
        env_credential_store_test.go:237: set = 1  curator: environments.credential_store.codex: unknown canonical environment "prod"
    --- FAIL: TestCredentialStoreCLIDottedProfileRouting/codex.prod/claude (0.00s)
        env_credential_store_test.go:237: set = 1  curator: environments.credential_store.codex: unknown canonical environment "prod"
    --- FAIL: TestCredentialStoreCLIDottedProfileRouting/codex.prod/claude_code (0.00s)
        env_credential_store_test.go:237: set = 1  curator: environments.credential_store.codex: unknown canonical environment "prod"
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	0.363s
FAIL

```

## reserved-profile-show (expected failure, exit 1)

```text
--- FAIL: TestCredentialStoreCLIInvalidProfileRouting (0.00s)
    --- FAIL: TestCredentialStoreCLIInvalidProfileRouting/CON.prod/show (0.00s)
        env_credential_store_test.go:296: invalid profile = 0 "native"
             
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	0.355s
FAIL

```

## Free-hunt overlay delta

Apply this delta as a Go overlay of cmd/curator/env_credential_store_test.go on the exact candidate.

```diff
--- cmd/curator/env_credential_store_test.go	2026-10-16 11:50:06
+++ .temp/review-2/freehunt.go	2026-10-16 13:44:52
@@ -224,7 +224,7 @@
 // Full selector paths preserve profile punctuation and case for both aliases and
 // canonical IDs. Failed writes and missing unsets leave the existing map intact.
 func TestCredentialStoreCLIDottedProfileRouting(t *testing.T) {
-	for _, profile := range []string{"team", "team.prod", "Team..prod", "codex.prod"} {
+	for _, profile := range []string{"team", "team.prod", "Team..prod", "codex.prod", strings.Repeat("p",128), "a."+strings.Repeat("b",126), "COM0.prod", "a...b"} {
 		for _, environment := range []string{"codex", "codex_cli", "claude", "claude_code"} {
 			t.Run(profile+"/"+environment, func(t *testing.T) {
 				source := writeMachineConfig(t, `{"schema_version":5,"skills_root":"x","projects":{},"environments":{"credential_store":{"neighbor":{"codex_cli":"keychain"}}}}`)
@@ -285,7 +285,7 @@
 	if err != nil {
 		t.Fatal(err)
 	}
-	for _, profile := range []string{".team", "team.", "team/prod", "CON.prod", strings.Repeat("p", 129)} {
+	for _, profile := range []string{".team", "team.", "team/prod", "CON.prod", strings.Repeat("p", 129), "LPT1.prod", "com9.prod", "a:b", "a\nb", "équipe", "a."+strings.Repeat("b",127)} {
 		for _, op := range []string{"set", "show", "unset"} {
 			t.Run(profile+"/"+op, func(t *testing.T) {
 				args := []string{"env", "config", op, "credential_store." + profile + ".codex"}

```

## remote-gate.log

```text
Detached snapshot signature verified.
remote gate: attempt 1/3: pushing f86c636bf7720ea0d5f3ad85e68df33e66c5e1a5 as gate/STORY-261016-26uala/261016-094332-93083-1
remote: 
remote: Create a pull request for 'gate/STORY-261016-26uala/261016-094332-93083-1' on GitHub by visiting:        
remote:      https://github.com/relux-works/curator/pull/new/gate/STORY-261016-26uala/261016-094332-93083-1        
remote: 
remote gate: run 38882423088 (https://github.com/relux-works/curator/actions/runs/38882423088)
remote gate: run 38882423088 finished: success
  Test (macos-latest): success  
  Lint: success  
  Test (ubuntu-latest): success  
  Test (windows-latest): success  
  Go 1.27.1 driver (macos-latest): success  
  Naming gate: success  
  Gate self-test (windows-latest): success  
  Gate self-test (macos-latest): success  
  Race (macos-latest): success  
  Go 1.25.5 driver (windows-latest): success  
  Race (ubuntu-latest): success  
  Go 1.27.1 driver (ubuntu-latest): success  
  Gate self-test (ubuntu-latest): success  
  Go 1.25.5 driver (ubuntu-latest): success  
  Go 1.25.5 driver (macos-latest): success  
  Interop conformance gate: success  
  Go 1.26.0 driver (ubuntu-latest): success  
  Go 1.26.0 driver (macos-latest): success  
  Go 1.27.1 driver (windows-latest): success  
  Go 1.26.0 driver (windows-latest): success  
  Optional self-hosted test job: skipped  
  Candidate suite (${{ matrix.os }}): skipped  

```

## hosted-final.json

```text
{"conclusion":"success","headSha":"f86c636bf7720ea0d5f3ad85e68df33e66c5e1a5","jobs":[{"completedAt":"2026-10-16T10:05:35Z","conclusion":"success","databaseId":116912077803,"name":"Test (macos-latest)","startedAt":"2026-10-16T09:43:50Z","status":"completed","steps":[{"completedAt":"2026-10-16T09:43:53Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T09:43:51Z","status":"completed"},{"completedAt":"2026-10-16T09:44:02Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T09:43:53Z","status":"completed"},{"completedAt":"2026-10-16T09:44:04Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T09:44:02Z","status":"completed"},{"completedAt":"2026-10-16T09:44:26Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T09:44:04Z","status":"completed"},{"completedAt":"2026-10-16T09:44:27Z","conclusion":"success","name":"Run actions/setup-node@v4","number":5,"startedAt":"2026-10-16T09:44:26Z","status":"completed"},{"completedAt":"2026-10-16T09:44:27Z","conclusion":"success","name":"Verify pnpm pin matches Go constant","number":6,"startedAt":"2026-10-16T09:44:27Z","status":"completed"},{"completedAt":"2026-10-16T09:44:29Z","conclusion":"success","name":"Install pinned pnpm via npm","number":7,"startedAt":"2026-10-16T09:44:27Z","status":"completed"},{"completedAt":"2026-10-16T09:44:29Z","conclusion":"success","name":"Verify rust pin matches Go constant","number":8,"startedAt":"2026-10-16T09:44:29Z","status":"completed"},{"completedAt":"2026-10-16T09:44:37Z","conclusion":"success","name":"Install pinned Rust toolchain via rustup","number":9,"startedAt":"2026-10-16T09:44:29Z","status":"completed"},{"completedAt":"2026-10-16T09:44:37Z","conclusion":"success","name":"Verify Go toolchain identity","number":10,"startedAt":"2026-10-16T09:44:37Z","status":"completed"},{"completedAt":"2026-10-16T09:44:38Z","conclusion":"success","name":"Verify versioned Go install compatibility","number":11,"startedAt":"2026-10-16T09:44:37Z","status":"completed"},{"completedAt":"2026-10-16T09:44:38Z","conclusion":"success","name":"gofmt check","number":12,"startedAt":"2026-10-16T09:44:38Z","status":"completed"},{"completedAt":"2026-10-16T09:44:51Z","conclusion":"success","name":"go vet","number":13,"startedAt":"2026-10-16T09:44:38Z","status":"completed"},{"completedAt":"2026-10-16T09:45:08Z","conclusion":"success","name":"Ledger consistency","number":14,"startedAt":"2026-10-16T09:44:51Z","status":"completed"},{"completedAt":"2026-10-16T10:05:28Z","conclusion":"success","name":"go test + platform-case gate","number":15,"startedAt":"2026-10-16T09:45:08Z","status":"completed"},{"completedAt":"2026-10-16T10:05:30Z","conclusion":"success","name":"Upload gate evidence","number":16,"startedAt":"2026-10-16T10:05:28Z","status":"completed"},{"completedAt":"2026-10-16T10:05:30Z","conclusion":"success","name":"Post Run actions/setup-node@v4","number":29,"startedAt":"2026-10-16T10:05:30Z","status":"completed"},{"completedAt":"2026-10-16T10:05:30Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":30,"startedAt":"2026-10-16T10:05:30Z","status":"completed"},{"completedAt":"2026-10-16T10:05:31Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":31,"startedAt":"2026-10-16T10:05:30Z","status":"completed"},{"completedAt":"2026-10-16T10:05:32Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":32,"startedAt":"2026-10-16T10:05:31Z","status":"completed"},{"completedAt":"2026-10-16T10:05:33Z","conclusion":"success","name":"Complete job","number":33,"startedAt":"2026-10-16T10:05:32Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38882423088/job/116912077803"},{"completedAt":"2026-10-16T09:44:48Z","conclusion":"success","databaseId":116912078054,"name":"Lint","startedAt":"2026-10-16T09:43:39Z","status":"completed","steps":[{"completedAt":"2026-10-16T09:43:41Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T09:43:40Z","status":"completed"},{"completedAt":"2026-10-16T09:43:50Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T09:43:41Z","status":"completed"},{"completedAt":"2026-10-16T09:44:00Z","conclusion":"success","name":"Run actions/setup-go@v5","number":3,"startedAt":"2026-10-16T09:43:50Z","status":"completed"},{"completedAt":"2026-10-16T09:44:00Z","conclusion":"success","name":"Verify Go toolchain identity","number":4,"startedAt":"2026-10-16T09:44:00Z","status":"completed"},{"completedAt":"2026-10-16T09:44:42Z","conclusion":"success","name":"Run golangci/golangci-lint-action@v7","number":5,"startedAt":"2026-10-16T09:44:00Z","status":"completed"},{"completedAt":"2026-10-16T09:44:42Z","conclusion":"success","name":"No broad suppression","number":6,"startedAt":"2026-10-16T09:44:42Z","status":"completed"},{"completedAt":"2026-10-16T09:44:43Z","conclusion":"success","name":"Verify GoReleaser rc channel values","number":7,"startedAt":"2026-10-16T09:44:43Z","status":"completed"},{"completedAt":"2026-10-16T09:44:43Z","conclusion":"success","name":"Post Run golangci/golangci-lint-action@v7","number":12,"startedAt":"2026-10-16T09:44:43Z","status":"completed"},{"completedAt":"2026-10-16T09:44:43Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":13,"startedAt":"2026-10-16T09:44:43Z","status":"completed"},{"completedAt":"2026-10-16T09:44:43Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":14,"startedAt":"2026-10-16T09:44:43Z","status":"completed"},{"completedAt":"2026-10-16T09:44:43Z","conclusion":"success","name":"Complete job","number":15,"startedAt":"2026-10-16T09:44:43Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38882423088/job/116912078054"},{"completedAt":"2026-10-16T10:05:52Z","conclusion":"success","databaseId":116912078104,"name":"Test (ubuntu-latest)","startedAt":"2026-10-16T09:43:39Z","status":"completed","steps":[{"completedAt":"2026-10-16T09:43:41Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T09:43:39Z","status":"completed"},{"completedAt":"2026-10-16T09:43:52Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T09:43:41Z","status":"completed"},{"completedAt":"2026-10-16T09:43:53Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T09:43:52Z","status":"completed"},{"completedAt":"2026-10-16T09:44:04Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T09:43:53Z","status":"completed"},{"completedAt":"2026-10-16T09:44:05Z","conclusion":"success","name":"Run actions/setup-node@v4","number":5,"startedAt":"2026-10-16T09:44:04Z","status":"completed"},{"completedAt":"2026-10-16T09:44:05Z","conclusion":"success","name":"Verify pnpm pin matches Go constant","number":6,"startedAt":"2026-10-16T09:44:05Z","status":"completed"},{"completedAt":"2026-10-16T09:44:09Z","conclusion":"success","name":"Install pinned pnpm via npm","number":7,"startedAt":"2026-10-16T09:44:05Z","status":"completed"},{"completedAt":"2026-10-16T09:44:09Z","conclusion":"success","name":"Verify rust pin matches Go constant","number":8,"startedAt":"2026-10-16T09:44:09Z","status":"completed"},{"completedAt":"2026-10-16T09:44:16Z","conclusion":"success","name":"Install pinned Rust toolchain via rustup","number":9,"startedAt":"2026-10-16T09:44:09Z","status":"completed"},{"completedAt":"2026-10-16T09:44:16Z","conclusion":"success","name":"Verify Go toolchain identity","number":10,"startedAt":"2026-10-16T09:44:16Z","status":"completed"},{"completedAt":"2026-10-16T09:44:16Z","conclusion":"success","name":"Verify versioned Go install compatibility","number":11,"startedAt":"2026-10-16T09:44:16Z","status":"completed"},{"completedAt":"2026-10-16T09:44:16Z","conclusion":"success","name":"gofmt check","number":12,"startedAt":"2026-10-16T09:44:16Z","status":"completed"},{"completedAt":"2026-10-16T09:44:32Z","conclusion":"success","name":"go vet","number":13,"startedAt":"2026-10-16T09:44:16Z","status":"completed"},{"completedAt":"2026-10-16T09:44:37Z","conclusion":"success","name":"Ledger consistency","number":14,"startedAt":"2026-10-16T09:44:32Z","status":"completed"},{"completedAt":"2026-10-16T10:05:46Z","conclusion":"success","name":"go test + platform-case gate","number":15,"startedAt":"2026-10-16T09:44:37Z","status":"completed"},{"completedAt":"2026-10-16T10:05:47Z","conclusion":"success","name":"Upload gate evidence","number":16,"startedAt":"2026-10-16T10:05:46Z","status":"completed"},{"completedAt":"2026-10-16T10:05:48Z","conclusion":"success","name":"Post Run actions/setup-node@v4","number":29,"startedAt":"2026-10-16T10:05:47Z","status":"completed"},{"completedAt":"2026-10-16T10:05:48Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":30,"startedAt":"2026-10-16T10:05:48Z","status":"completed"},{"completedAt":"2026-10-16T10:05:48Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":31,"startedAt":"2026-10-16T10:05:48Z","status":"completed"},{"completedAt":"2026-10-16T10:05:48Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":32,"startedAt":"2026-10-16T10:05:48Z","status":"completed"},{"completedAt":"2026-10-16T10:05:48Z","conclusion":"success","name":"Complete job","number":33,"startedAt":"2026-10-16T10:05:48Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38882423088/job/116912078104"},{"completedAt":"2026-10-16T10:31:01Z","conclusion":"success","databaseId":116912078125,"name":"Test (windows-latest)","startedAt":"2026-10-16T09:43:38Z","status":"completed","steps":[{"completedAt":"2026-10-16T09:43:41Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T09:43:39Z","status":"completed"},{"completedAt":"2026-10-16T09:44:02Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T09:43:41Z","status":"completed"},{"completedAt":"2026-10-16T09:44:06Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T09:44:02Z","status":"completed"},{"completedAt":"2026-10-16T09:44:41Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T09:44:06Z","status":"completed"},{"completedAt":"2026-10-16T09:44:46Z","conclusion":"success","name":"Run actions/setup-node@v4","number":5,"startedAt":"2026-10-16T09:44:41Z","status":"completed"},{"completedAt":"2026-10-16T09:44:46Z","conclusion":"success","name":"Verify pnpm pin matches Go constant","number":6,"startedAt":"2026-10-16T09:44:46Z","status":"completed"},{"completedAt":"2026-10-16T09:44:54Z","conclusion":"success","name":"Install pinned pnpm via npm","number":7,"startedAt":"2026-10-16T09:44:46Z","status":"completed"},{"completedAt":"2026-10-16T09:44:54Z","conclusion":"success","name":"Verify rust pin matches Go constant","number":8,"startedAt":"2026-10-16T09:44:54Z","status":"completed"},{"completedAt":"2026-10-16T09:45:06Z","conclusion":"success","name":"Install pinned Rust toolchain via rustup","number":9,"startedAt":"2026-10-16T09:44:54Z","status":"completed"},{"completedAt":"2026-10-16T09:45:06Z","conclusion":"success","name":"Verify Go toolchain identity","number":10,"startedAt":"2026-10-16T09:45:06Z","status":"completed"},{"completedAt":"2026-10-16T09:45:06Z","conclusion":"success","name":"Verify versioned Go install compatibility","number":11,"startedAt":"2026-10-16T09:45:06Z","status":"completed"},{"completedAt":"2026-10-16T09:45:06Z","conclusion":"skipped","name":"gofmt check","number":12,"startedAt":"2026-10-16T09:45:06Z","status":"completed"},{"completedAt":"2026-10-16T09:45:23Z","conclusion":"success","name":"go vet","number":13,"startedAt":"2026-10-16T09:45:06Z","status":"completed"},{"completedAt":"2026-10-16T09:46:05Z","conclusion":"success","name":"Ledger consistency","number":14,"startedAt":"2026-10-16T09:45:23Z","status":"completed"},{"completedAt":"2026-10-16T10:30:52Z","conclusion":"success","name":"go test + platform-case gate","number":15,"startedAt":"2026-10-16T09:46:05Z","status":"completed"},{"completedAt":"2026-10-16T10:30:54Z","conclusion":"success","name":"Upload gate evidence","number":16,"startedAt":"2026-10-16T10:30:52Z","status":"completed"},{"completedAt":"2026-10-16T10:30:54Z","conclusion":"success","name":"Post Run actions/setup-node@v4","number":29,"startedAt":"2026-10-16T10:30:54Z","status":"completed"},{"completedAt":"2026-10-16T10:30:54Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":30,"startedAt":"2026-10-16T10:30:54Z","status":"completed"},{"completedAt":"2026-10-16T10:30:56Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":31,"startedAt":"2026-10-16T10:30:54Z","status":"completed"},{"completedAt":"2026-10-16T10:30:58Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":32,"startedAt":"2026-10-16T10:30:56Z","status":"completed"},{"completedAt":"2026-10-16T10:30:58Z","conclusion":"success","name":"Complete job","number":33,"startedAt":"2026-10-16T10:30:58Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38882423088/job/116912078125"},{"completedAt":"2026-10-16T09:46:47Z","conclusion":"success","databaseId":116912078131,"name":"Go 1.27.1 driver (macos-latest)","startedAt":"2026-10-16T09:44:00Z","status":"completed","steps":[{"completedAt":"2026-10-16T09:44:02Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T09:44:00Z","status":"completed"},{"completedAt":"2026-10-16T09:44:12Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T09:44:02Z","status":"completed"},{"completedAt":"2026-10-16T09:44:14Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T09:44:12Z","status":"completed"},{"completedAt":"2026-10-16T09:44:38Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T09:44:14Z","status":"completed"},{"completedAt":"2026-10-16T09:44:38Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T09:44:38Z","status":"completed"},{"completedAt":"2026-10-16T09:46:22Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T09:44:38Z","status":"completed"},{"completedAt":"2026-10-16T09:46:25Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T09:46:22Z","status":"completed"},{"completedAt":"2026-10-16T09:46:40Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T09:46:25Z","status":"completed"},{"completedAt":"2026-10-16T09:46:41Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T09:46:40Z","status":"completed"},{"completedAt":"2026-10-16T09:46:41Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T09:46:41Z","status":"completed"},{"completedAt":"2026-10-16T09:46:42Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T09:46:41Z","status":"completed"},{"completedAt":"2026-10-16T09:46:44Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T09:46:42Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38882423088/job/116912078131"},{"completedAt":"2026-10-16T09:43:51Z","conclusion":"success","databaseId":116912078138,"name":"Naming gate","startedAt":"2026-10-16T09:43:39Z","status":"completed","steps":[{"completedAt":"2026-10-16T09:43:40Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T09:43:39Z","status":"completed"},{"completedAt":"2026-10-16T09:43:48Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T09:43:40Z","status":"completed"},{"completedAt":"2026-10-16T09:43:49Z","conclusion":"success","name":"Employer name gate","number":3,"startedAt":"2026-10-16T09:43:48Z","status":"completed"},{"completedAt":"2026-10-16T09:43:50Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":6,"startedAt":"2026-10-16T09:43:49Z","status":"completed"},{"completedAt":"2026-10-16T09:43:50Z","conclusion":"success","name":"Complete job","number":7,"startedAt":"2026-10-16T09:43:50Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38882423088/job/116912078138"},{"completedAt":"2026-10-16T09:45:15Z","conclusion":"success","databaseId":116912078191,"name":"Gate self-test (windows-latest)","startedAt":"2026-10-16T09:43:39Z","status":"completed","steps":[{"completedAt":"2026-10-16T09:43:40Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T09:43:39Z","status":"completed"},{"completedAt":"2026-10-16T09:43:59Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T09:43:40Z","status":"completed"},{"completedAt":"2026-10-16T09:45:11Z","conclusion":"success","name":"Self-test the CI gates","number":3,"startedAt":"2026-10-16T09:43:59Z","status":"completed"},{"completedAt":"2026-10-16T09:45:13Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":6,"startedAt":"2026-10-16T09:45:11Z","status":"completed"},{"completedAt":"2026-10-16T09:45:13Z","conclusion":"success","name":"Complete job","number":7,"startedAt":"2026-10-16T09:45:13Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38882423088/job/116912078191"},{"completedAt":"2026-10-16T09:44:13Z","conclusion":"success","databaseId":116912078199,"name":"Gate self-test (macos-latest)","startedAt":"2026-10-16T09:43:56Z","status":"completed","steps":[{"completedAt":"2026-10-16T09:43:57Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T09:43:57Z","status":"completed"},{"completedAt":"2026-10-16T09:44:03Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T09:43:57Z","status":"completed"},{"completedAt":"2026-10-16T09:44:10Z","conclusion":"success","name":"Self-test the CI gates","number":3,"startedAt":"2026-10-16T09:44:03Z","status":"completed"},{"completedAt":"2026-10-16T09:44:11Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":6,"startedAt":"2026-10-16T09:44:10Z","status":"completed"},{"completedAt":"2026-10-16T09:44:11Z","conclusion":"success","name":"Complete job","number":7,"startedAt":"2026-10-16T09:44:11Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38882423088/job/116912078199"},{"completedAt":"2026-10-16T10:17:28Z","conclusion":"success","databaseId":116912078216,"name":"Race (macos-latest)","startedAt":"2026-10-16T09:44:20Z","status":"completed","steps":[{"completedAt":"2026-10-16T09:44:22Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T09:44:20Z","status":"completed"},{"completedAt":"2026-10-16T09:44:34Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T09:44:22Z","status":"completed"},{"completedAt":"2026-10-16T09:44:36Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T09:44:34Z","status":"completed"},{"completedAt":"2026-10-16T09:45:01Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T09:44:36Z","status":"completed"},{"completedAt":"2026-10-16T09:45:02Z","conclusion":"success","name":"Run actions/setup-node@v4","number":5,"startedAt":"2026-10-16T09:45:01Z","status":"completed"},{"completedAt":"2026-10-16T09:45:02Z","conclusion":"success","name":"Verify pnpm pin matches Go constant","number":6,"startedAt":"2026-10-16T09:45:02Z","status":"completed"},{"completedAt":"2026-10-16T09:45:03Z","conclusion":"success","name":"Install pinned pnpm via npm","number":7,"startedAt":"2026-10-16T09:45:02Z","status":"completed"},{"completedAt":"2026-10-16T09:45:04Z","conclusion":"success","name":"Verify rust pin matches Go constant","number":8,"startedAt":"2026-10-16T09:45:03Z","status":"completed"},{"completedAt":"2026-10-16T09:45:11Z","conclusion":"success","name":"Install pinned Rust toolchain via rustup","number":9,"startedAt":"2026-10-16T09:45:04Z","status":"completed"},{"completedAt":"2026-10-16T09:45:11Z","conclusion":"success","name":"Verify Go toolchain identity","number":10,"startedAt":"2026-10-16T09:45:11Z","status":"completed"},{"completedAt":"2026-10-16T10:17:20Z","conclusion":"success","name":"go test -race + platform-case gate","number":11,"startedAt":"2026-10-16T09:45:11Z","status":"completed"},{"completedAt":"2026-10-16T10:17:22Z","conclusion":"success","name":"Upload race evidence","number":12,"startedAt":"2026-10-16T10:17:20Z","status":"completed"},{"completedAt":"2026-10-16T10:17:23Z","conclusion":"success","name":"Post Run actions/setup-node@v4","number":21,"startedAt":"2026-10-16T10:17:22Z","status":"completed"},{"completedAt":"2026-10-16T10:17:23Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":22,"startedAt":"2026-10-16T10:17:23Z","status":"completed"},{"completedAt":"2026-10-16T10:17:24Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":23,"startedAt":"2026-10-16T10:17:23Z","status":"completed"},{"completedAt":"2026-10-16T10:17:25Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":24,"startedAt":"2026-10-16T10:17:24Z","status":"completed"},{"completedAt":"2026-10-16T10:17:26Z","conclusion":"success","name":"Complete job","number":25,"startedAt":"2026-10-16T10:17:25Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38882423088/job/116912078216"},{"completedAt":"2026-10-16T09:47:12Z","conclusion":"success","databaseId":116912078289,"name":"Go 1.25.5 driver (windows-latest)","startedAt":"2026-10-16T09:43:39Z","status":"completed","steps":[{"completedAt":"2026-10-16T09:43:42Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T09:43:40Z","status":"completed"},{"completedAt":"2026-10-16T09:43:57Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T09:43:42Z","status":"completed"},{"completedAt":"2026-10-16T09:44:00Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T09:43:57Z","status":"completed"},{"completedAt":"2026-10-16T09:44:40Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T09:44:00Z","status":"completed"},{"completedAt":"2026-10-16T09:44:40Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T09:44:40Z","status":"completed"},{"completedAt":"2026-10-16T09:46:46Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T09:44:40Z","status":"completed"},{"completedAt":"2026-10-16T09:46:49Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T09:46:46Z","status":"completed"},{"completedAt":"2026-10-16T09:47:05Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T09:46:49Z","status":"completed"},{"completedAt":"2026-10-16T09:47:06Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T09:47:05Z","status":"completed"},{"completedAt":"2026-10-16T09:47:08Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T09:47:06Z","status":"completed"},{"completedAt":"2026-10-16T09:47:10Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T09:47:08Z","status":"completed"},{"completedAt":"2026-10-16T09:47:10Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T09:47:10Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38882423088/job/116912078289"},{"completedAt":"2026-10-16T10:13:30Z","conclusion":"success","databaseId":116912078292,"name":"Race (ubuntu-latest)","startedAt":"2026-10-16T09:43:39Z","status":"completed","steps":[{"completedAt":"2026-10-16T09:43:41Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T09:43:40Z","status":"completed"},{"completedAt":"2026-10-16T09:43:50Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T09:43:41Z","status":"completed"},{"completedAt":"2026-10-16T09:43:51Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T09:43:50Z","status":"completed"},{"completedAt":"2026-10-16T09:44:01Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T09:43:51Z","status":"completed"},{"completedAt":"2026-10-16T09:44:02Z","conclusion":"success","name":"Run actions/setup-node@v4","number":5,"startedAt":"2026-10-16T09:44:01Z","status":"completed"},{"completedAt":"2026-10-16T09:44:02Z","conclusion":"success","name":"Verify pnpm pin matches Go constant","number":6,"startedAt":"2026-10-16T09:44:02Z","status":"completed"},{"completedAt":"2026-10-16T09:44:04Z","conclusion":"success","name":"Install pinned pnpm via npm","number":7,"startedAt":"2026-10-16T09:44:02Z","status":"completed"},{"completedAt":"2026-10-16T09:44:04Z","conclusion":"success","name":"Verify rust pin matches Go constant","number":8,"startedAt":"2026-10-16T09:44:04Z","status":"completed"},{"completedAt":"2026-10-16T09:44:12Z","conclusion":"success","name":"Install pinned Rust toolchain via rustup","number":9,"startedAt":"2026-10-16T09:44:04Z","status":"completed"},{"completedAt":"2026-10-16T09:44:12Z","conclusion":"success","name":"Verify Go toolchain identity","number":10,"startedAt":"2026-10-16T09:44:12Z","status":"completed"},{"completedAt":"2026-10-16T10:13:23Z","conclusion":"success","name":"go test -race + platform-case gate","number":11,"startedAt":"2026-10-16T09:44:12Z","status":"completed"},{"completedAt":"2026-10-16T10:13:26Z","conclusion":"success","name":"Upload race evidence","number":12,"startedAt":"2026-10-16T10:13:23Z","status":"completed"},{"completedAt":"2026-10-16T10:13:26Z","conclusion":"success","name":"Post Run actions/setup-node@v4","number":21,"startedAt":"2026-10-16T10:13:26Z","status":"completed"},{"completedAt":"2026-10-16T10:13:26Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":22,"startedAt":"2026-10-16T10:13:26Z","status":"completed"},{"completedAt":"2026-10-16T10:13:26Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":23,"startedAt":"2026-10-16T10:13:26Z","status":"completed"},{"completedAt":"2026-10-16T10:13:26Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":24,"startedAt":"2026-10-16T10:13:26Z","status":"completed"},{"completedAt":"2026-10-16T10:13:26Z","conclusion":"success","name":"Complete job","number":25,"startedAt":"2026-10-16T10:13:26Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38882423088/job/116912078292"},{"completedAt":"2026-10-16T09:44:21Z","conclusion":"success","databaseId":116912078340,"name":"Go 1.27.1 driver (ubuntu-latest)","startedAt":"2026-10-16T09:43:38Z","status":"completed","steps":[{"completedAt":"2026-10-16T09:43:40Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T09:43:39Z","status":"completed"},{"completedAt":"2026-10-16T09:43:47Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T09:43:40Z","status":"completed"},{"completedAt":"2026-10-16T09:43:47Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T09:43:47Z","status":"completed"},{"completedAt":"2026-10-16T09:43:58Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T09:43:47Z","status":"completed"},{"completedAt":"2026-10-16T09:43:58Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T09:43:58Z","status":"completed"},{"completedAt":"2026-10-16T09:44:02Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T09:43:58Z","status":"completed"},{"completedAt":"2026-10-16T09:44:04Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T09:44:02Z","status":"completed"},{"completedAt":"2026-10-16T09:44:18Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T09:44:04Z","status":"completed"},{"completedAt":"2026-10-16T09:44:18Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T09:44:18Z","status":"completed"},{"completedAt":"2026-10-16T09:44:18Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T09:44:18Z","status":"completed"},{"completedAt":"2026-10-16T09:44:18Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T09:44:18Z","status":"completed"},{"completedAt":"2026-10-16T09:44:18Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T09:44:18Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38882423088/job/116912078340"},{"completedAt":"2026-10-16T09:44:09Z","conclusion":"success","databaseId":116912078366,"name":"Gate self-test (ubuntu-latest)","startedAt":"2026-10-16T09:43:39Z","status":"completed","steps":[{"completedAt":"2026-10-16T09:43:40Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T09:43:40Z","status":"completed"},{"completedAt":"2026-10-16T09:43:47Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T09:43:40Z","status":"completed"},{"completedAt":"2026-10-16T09:44:06Z","conclusion":"success","name":"Self-test the CI gates","number":3,"startedAt":"2026-10-16T09:43:47Z","status":"completed"},{"completedAt":"2026-10-16T09:44:06Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":6,"startedAt":"2026-10-16T09:44:06Z","status":"completed"},{"completedAt":"2026-10-16T09:44:06Z","conclusion":"success","name":"Complete job","number":7,"startedAt":"2026-10-16T09:44:06Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38882423088/job/116912078366"},{"completedAt":"2026-10-16T09:45:05Z","conclusion":"success","databaseId":116912078414,"name":"Go 1.25.5 driver (ubuntu-latest)","startedAt":"2026-10-16T09:44:22Z","status":"completed","steps":[{"completedAt":"2026-10-16T09:44:24Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T09:44:23Z","status":"completed"},{"completedAt":"2026-10-16T09:44:30Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T09:44:24Z","status":"completed"},{"completedAt":"2026-10-16T09:44:31Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T09:44:30Z","status":"completed"},{"completedAt":"2026-10-16T09:44:40Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T09:44:31Z","status":"completed"},{"completedAt":"2026-10-16T09:44:40Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T09:44:40Z","status":"completed"},{"completedAt":"2026-10-16T09:44:46Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T09:44:40Z","status":"completed"},{"completedAt":"2026-10-16T09:44:48Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T09:44:46Z","status":"completed"},{"completedAt":"2026-10-16T09:45:02Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T09:44:48Z","status":"completed"},{"completedAt":"2026-10-16T09:45:02Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T09:45:02Z","status":"completed"},{"completedAt":"2026-10-16T09:45:02Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T09:45:02Z","status":"completed"},{"completedAt":"2026-10-16T09:45:03Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T09:45:02Z","status":"completed"},{"completedAt":"2026-10-16T09:45:03Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T09:45:03Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38882423088/job/116912078414"},{"completedAt":"2026-10-16T09:49:25Z","conclusion":"success","databaseId":116912078431,"name":"Go 1.25.5 driver (macos-latest)","startedAt":"2026-10-16T09:47:18Z","status":"completed","steps":[{"completedAt":"2026-10-16T09:47:20Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T09:47:19Z","status":"completed"},{"completedAt":"2026-10-16T09:47:29Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T09:47:20Z","status":"completed"},{"completedAt":"2026-10-16T09:47:31Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T09:47:29Z","status":"completed"},{"completedAt":"2026-10-16T09:47:51Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T09:47:31Z","status":"completed"},{"completedAt":"2026-10-16T09:47:51Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T09:47:51Z","status":"completed"},{"completedAt":"2026-10-16T09:49:07Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T09:47:51Z","status":"completed"},{"completedAt":"2026-10-16T09:49:09Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T09:49:07Z","status":"completed"},{"completedAt":"2026-10-16T09:49:20Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T09:49:09Z","status":"completed"},{"completedAt":"2026-10-16T09:49:20Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T09:49:20Z","status":"completed"},{"completedAt":"2026-10-16T09:49:21Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T09:49:20Z","status":"completed"},{"completedAt":"2026-10-16T09:49:21Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T09:49:21Z","status":"completed"},{"completedAt":"2026-10-16T09:49:22Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T09:49:21Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38882423088/job/116912078431"},{"completedAt":"2026-10-16T09:45:13Z","conclusion":"success","databaseId":116912078450,"name":"Interop conformance gate","startedAt":"2026-10-16T09:44:47Z","status":"completed","steps":[{"completedAt":"2026-10-16T09:44:49Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T09:44:48Z","status":"completed"},{"completedAt":"2026-10-16T09:44:56Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T09:44:49Z","status":"completed"},{"completedAt":"2026-10-16T09:44:57Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T09:44:56Z","status":"completed"},{"completedAt":"2026-10-16T09:45:07Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T09:44:57Z","status":"completed"},{"completedAt":"2026-10-16T09:45:08Z","conclusion":"success","name":"Verify Go toolchain identity","number":5,"startedAt":"2026-10-16T09:45:07Z","status":"completed"},{"completedAt":"2026-10-16T09:45:10Z","conclusion":"success","name":"Shared protocol suite","number":6,"startedAt":"2026-10-16T09:45:08Z","status":"completed"},{"completedAt":"2026-10-16T09:45:10Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":10,"startedAt":"2026-10-16T09:45:10Z","status":"completed"},{"completedAt":"2026-10-16T09:45:11Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":11,"startedAt":"2026-10-16T09:45:10Z","status":"completed"},{"completedAt":"2026-10-16T09:45:11Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":12,"startedAt":"2026-10-16T09:45:11Z","status":"completed"},{"completedAt":"2026-10-16T09:45:11Z","conclusion":"success","name":"Complete job","number":13,"startedAt":"2026-10-16T09:45:11Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38882423088/job/116912078450"},{"completedAt":"2026-10-16T09:44:55Z","conclusion":"success","databaseId":116912078474,"name":"Go 1.26.0 driver (ubuntu-latest)","startedAt":"2026-10-16T09:44:10Z","status":"completed","steps":[{"completedAt":"2026-10-16T09:44:12Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T09:44:11Z","status":"completed"},{"completedAt":"2026-10-16T09:44:19Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T09:44:12Z","status":"completed"},{"completedAt":"2026-10-16T09:44:20Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T09:44:19Z","status":"completed"},{"completedAt":"2026-10-16T09:44:31Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T09:44:20Z","status":"completed"},{"completedAt":"2026-10-16T09:44:31Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T09:44:31Z","status":"completed"},{"completedAt":"2026-10-16T09:44:36Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T09:44:31Z","status":"completed"},{"completedAt":"2026-10-16T09:44:37Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T09:44:36Z","status":"completed"},{"completedAt":"2026-10-16T09:44:52Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T09:44:37Z","status":"completed"},{"completedAt":"2026-10-16T09:44:52Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T09:44:52Z","status":"completed"},{"completedAt":"2026-10-16T09:44:53Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T09:44:52Z","status":"completed"},{"completedAt":"2026-10-16T09:44:53Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T09:44:53Z","status":"completed"},{"completedAt":"2026-10-16T09:44:53Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T09:44:53Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38882423088/job/116912078474"},{"completedAt":"2026-10-16T10:08:20Z","conclusion":"success","databaseId":116912078475,"name":"Go 1.26.0 driver (macos-latest)","startedAt":"2026-10-16T10:05:43Z","status":"completed","steps":[{"completedAt":"2026-10-16T10:05:45Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T10:05:43Z","status":"completed"},{"completedAt":"2026-10-16T10:06:02Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T10:05:45Z","status":"completed"},{"completedAt":"2026-10-16T10:06:04Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T10:06:02Z","status":"completed"},{"completedAt":"2026-10-16T10:06:32Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T10:06:04Z","status":"completed"},{"completedAt":"2026-10-16T10:06:33Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T10:06:32Z","status":"completed"},{"completedAt":"2026-10-16T10:07:56Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T10:06:33Z","status":"completed"},{"completedAt":"2026-10-16T10:07:59Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T10:07:56Z","status":"completed"},{"completedAt":"2026-10-16T10:08:14Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T10:07:59Z","status":"completed"},{"completedAt":"2026-10-16T10:08:14Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T10:08:14Z","status":"completed"},{"completedAt":"2026-10-16T10:08:15Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T10:08:14Z","status":"completed"},{"completedAt":"2026-10-16T10:08:15Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T10:08:15Z","status":"completed"},{"completedAt":"2026-10-16T10:08:17Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T10:08:15Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38882423088/job/116912078475"},{"completedAt":"2026-10-16T09:47:45Z","conclusion":"success","databaseId":116912078534,"name":"Go 1.27.1 driver (windows-latest)","startedAt":"2026-10-16T09:43:53Z","status":"completed","steps":[{"completedAt":"2026-10-16T09:43:56Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T09:43:55Z","status":"completed"},{"completedAt":"2026-10-16T09:44:11Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T09:43:56Z","status":"completed"},{"completedAt":"2026-10-16T09:44:14Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T09:44:11Z","status":"completed"},{"completedAt":"2026-10-16T09:44:47Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T09:44:14Z","status":"completed"},{"completedAt":"2026-10-16T09:44:48Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T09:44:47Z","status":"completed"},{"completedAt":"2026-10-16T09:47:13Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T09:44:48Z","status":"completed"},{"completedAt":"2026-10-16T09:47:17Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T09:47:13Z","status":"completed"},{"completedAt":"2026-10-16T09:47:36Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T09:47:17Z","status":"completed"},{"completedAt":"2026-10-16T09:47:37Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T09:47:36Z","status":"completed"},{"completedAt":"2026-10-16T09:47:39Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T09:47:37Z","status":"completed"},{"completedAt":"2026-10-16T09:47:42Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T09:47:39Z","status":"completed"},{"completedAt":"2026-10-16T09:47:42Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T09:47:42Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38882423088/job/116912078534"},{"completedAt":"2026-10-16T09:48:57Z","conclusion":"success","databaseId":116912078915,"name":"Go 1.26.0 driver (windows-latest)","startedAt":"2026-10-16T09:44:49Z","status":"completed","steps":[{"completedAt":"2026-10-16T09:44:50Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T09:44:49Z","status":"completed"},{"completedAt":"2026-10-16T09:45:01Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T09:44:50Z","status":"completed"},{"completedAt":"2026-10-16T09:45:04Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T09:45:01Z","status":"completed"},{"completedAt":"2026-10-16T09:45:49Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T09:45:04Z","status":"completed"},{"completedAt":"2026-10-16T09:45:50Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T09:45:49Z","status":"completed"},{"completedAt":"2026-10-16T09:48:18Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T09:45:50Z","status":"completed"},{"completedAt":"2026-10-16T09:48:24Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T09:48:18Z","status":"completed"},{"completedAt":"2026-10-16T09:48:50Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T09:48:24Z","status":"completed"},{"completedAt":"2026-10-16T09:48:50Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T09:48:50Z","status":"completed"},{"completedAt":"2026-10-16T09:48:53Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T09:48:50Z","status":"completed"},{"completedAt":"2026-10-16T09:48:55Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T09:48:53Z","status":"completed"},{"completedAt":"2026-10-16T09:48:55Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T09:48:55Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38882423088/job/116912078915"},{"completedAt":"2026-10-16T09:43:36Z","conclusion":"skipped","databaseId":116912079313,"name":"Optional self-hosted test job","startedAt":"2026-10-16T09:43:37Z","status":"completed","steps":[],"url":"https://github.com/relux-works/curator/actions/runs/38882423088/job/116912079313"},{"completedAt":"2026-10-16T09:43:36Z","conclusion":"skipped","databaseId":116912079971,"name":"Candidate suite (${{ matrix.os }})","startedAt":"2026-10-16T09:43:37Z","status":"completed","steps":[],"url":"https://github.com/relux-works/curator/actions/runs/38882423088/job/116912079971"}],"status":"completed","url":"https://github.com/relux-works/curator/actions/runs/38882423088"}

```

Full gate process: exit 0. Terminal readback: completed/success, exact candidate tree; 20 successes, 2 conditional skips.
