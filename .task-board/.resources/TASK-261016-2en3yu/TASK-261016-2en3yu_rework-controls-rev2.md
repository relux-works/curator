# Revision 2 test and narrowing-control evidence

This artifact records local and control results. The full hosted gate subsequently reached completed/success with exit 0; see TASK-261016-2en3yu_rework-rev2.md for terminal evidence. Candidate snapshot: f2a79188823008f44e7e0a45b59beb044b00f115.


The additional reserved-profile-show control also exited 1 in the intended forbidden-success assertion. Nine of nine controls are killed.

## source-identity.json

```json
{
  "README.md": "db73f1c85a1fe77a187afb85d7e5989e0ff28c060b84d0aeb5c48a403c377fad",
  "cmd/curator/env.go": "898a936590707f0a9679fd77841b1438cbee0857c80f06ec5cbd64d02379a7b4",
  "cmd/curator/env_credential_store_test.go": "52e434263280db47057086567a3a65967ed206415fd931bdf9656e2d282ed0bb",
  "cmd/curator/envalias.go": "4fb6a573f68a5e185214607438d77a359eaa9576cc5f088b23916ba2b7fa06ef",
  "cmd/curator/envconfig.go": "1d8b60df5743e463d7e762e9fecd081ba747a28fdc362f156927148d33f96cd6",
  "docs/environment-config.md": "e76a39440468022a62e74b14773fdcb8389a34d33b3318ac99654ebee5612830",
  "internal/config/config.go": "c5d2bad326a74ad0ef05d99b41135e69ea2c83b74acaf4ef460f5125b51f59fd",
  "internal/config/credential_store.go": "e2c00ee916e3bca4211349ee33c5dae0ce35461b7d7e7a174138a078b8f571a8",
  "internal/config/credential_store_test.go": "0d8ae95be9e2ed50888556b03c21199c10659aa231b4e9ce8801158698685d9c",
  "internal/config/environments.go": "20fbfa0e403497949e4fa9e23ed060279b2b7b6a7712a5769c68f49f7c82f7b8",
  "internal/config/environments_test.go": "a8b7c9cd8e7c37cafc3490daeea2b781f52aaddad06b40d186b98838a6c74fc1",
  "internal/envprofile/credential_policy.go": "d05872a8839f0a62c5348b921e5dd3350b283f428d2c01d6cc7054d083036a16",
  "internal/envprofile/credential_policy_test.go": "57e3af1dcd5a1f50d782128819b3fd04d6c649329410853e28da425ec1821f00",
  "internal/envprofile/managed.go": "fb509443f6a379e8d0b17bf83606adfd4746adaa51c732e6a3a83b8e35317da6",
  "internal/envregistry/credential_store.go": "7de3c4cfd57a99d5b71cc63a61eb6b96525b8aae9cd2c49c289989342b0bdc20",
  "internal/envregistry/envregistry.go": "081c77cd90499acf5d48513d9cf618deb3658c66d96a49c3fa51984524956c80"
}

```

## TASK-261016-2en3yu_rev2-mutant-definitions.json

```json
[
  {
    "name": "drop-codex-file",
    "file": "cmd/curator/env.go",
    "replacements": [
      [
        "machine.CredentialStore = cfg.Env.CredentialStore",
        "machine.CredentialStore = map[string]map[string]envregistry.CredentialStoreSelection{}\n\tfor profile, entries := range cfg.Env.CredentialStore {\n\t\tmachine.CredentialStore[profile] = map[string]envregistry.CredentialStoreSelection{}\n\t\tfor id, selection := range entries {\n\t\t\tif id == envregistry.CodexCLI && selection == envregistry.CredentialStoreFile { continue }\n\t\t\tmachine.CredentialStore[profile][id] = selection\n\t\t}\n\t}"
      ]
    ],
    "package": "./cmd/curator",
    "test": "^TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery$"
  },
  {
    "name": "old-v3",
    "file": "internal/config/environments.go",
    "replacements": [
      [
        "if schema != SchemaVersion5 {",
        "if schema != SchemaVersion5 && schema != SchemaVersion3 {"
      ]
    ],
    "package": "./internal/config",
    "test": "^TestCredentialStoreConfigVersionBoundary$"
  },
  {
    "name": "permission-as-absent",
    "file": "internal/envprofile/managed.go",
    "replacements": [
      [
        "state, err := readRegularFile(path)\n\tif err != nil {",
        "state, err := readRegularFile(path)\n\tif errors.Is(err, os.ErrPermission) { state = stateread.File{Kind: stateread.KindAbsent}; err = nil }\n\tif err != nil {"
      ]
    ],
    "package": "./internal/envprofile",
    "test": "^TestCredentialStorePolicyNativeInspection$"
  },
  {
    "name": "persist-alias",
    "file": "internal/config/credential_store.go",
    "replacements": [
      [
        "envregistry.ByID(id)",
        "envregistry.ByID(envregistry.NormalizeEnvID(id))"
      ]
    ],
    "package": "./internal/config",
    "test": "^TestCredentialStoreConfigGrammar$"
  },
  {
    "name": "profile-normalization",
    "file": "cmd/curator/envalias.go",
    "replacements": [
      [
        "rewritten[2] = envregistry.NormalizeEnvID(rewritten[2])",
        "rewritten[1] = envregistry.NormalizeEnvID(rewritten[1])\n\t\trewritten[2] = envregistry.NormalizeEnvID(rewritten[2])"
      ]
    ],
    "package": "./cmd/curator",
    "test": "^TestCredentialStoreCLIConfigRoundTrip$"
  },
  {
    "name": "system-store",
    "file": "internal/config/environments.go",
    "replacements": [
      [
        "if !allowed[key] {",
        "if !allowed[key] && key != \"credential_store\" {"
      ],
      [
        "if _, err := parseEnvironments(raw); err != nil {",
        "if _, err := parseEnvironmentsVersion(raw, SchemaVersion5); err != nil {"
      ]
    ],
    "package": "./internal/config",
    "test": "^TestCredentialStoreConfigMachineOwnership$"
  },
  {
    "name": "unknown-auto",
    "file": "internal/envregistry/credential_store.go",
    "replacements": [
      [
        "s == CredentialStoreKeychain",
        "s == CredentialStoreKeychain || s == \"auto\""
      ]
    ],
    "package": "./internal/config",
    "test": "^TestCredentialStoreConfigGrammar$"
  },
  {
    "name": "dotted-profile-only",
    "file": "internal/config/environments.go",
    "replacements": [
      [
        "segments[0] == \"credential_store\" && len(segments) >= 3",
        "segments[0] == \"credential_store\" && len(segments) == 3"
      ]
    ],
    "package": "./cmd/curator",
    "test": "^TestCredentialStoreCLIDottedProfileRouting$"
  },
  {
    "name": "reserved-profile-show",
    "file": "internal/config/environments.go",
    "replacements": [
      [
        "if !identifiers.Valid(profile) {\n\t\t\t\treturn nil, verr.New(\"knob\",",
        "if !identifiers.Valid(profile) && profile != \"CON.prod\" {\n\t\t\t\treturn nil, verr.New(\"knob\","
      ]
    ],
    "package": "./cmd/curator",
    "test": "^TestCredentialStoreCLIInvalidProfileRouting$/^CON[.]prod$/^show$"
  }
]

```

## mutant-results.json

```json
[
  {
    "name": "drop-codex-file",
    "command": [
      "go",
      "test",
      "-overlay",
      ".temp/TASK-261016-2en3yu/rev2/mutants/drop-codex-file/overlay.json",
      "./cmd/curator",
      "-run",
      "^TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery$",
      "-count=1"
    ],
    "exit_code": 1,
    "named_failures": [
      "--- FAIL: TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery (0.59s)"
    ]
  },
  {
    "name": "old-v3",
    "command": [
      "go",
      "test",
      "-overlay",
      ".temp/TASK-261016-2en3yu/rev2/mutants/old-v3/overlay.json",
      "./internal/config",
      "-run",
      "^TestCredentialStoreConfigVersionBoundary$",
      "-count=1"
    ],
    "exit_code": 1,
    "named_failures": [
      "--- FAIL: TestCredentialStoreConfigVersionBoundary (0.00s)"
    ]
  },
  {
    "name": "permission-as-absent",
    "command": [
      "go",
      "test",
      "-overlay",
      ".temp/TASK-261016-2en3yu/rev2/mutants/permission-as-absent/overlay.json",
      "./internal/envprofile",
      "-run",
      "^TestCredentialStorePolicyNativeInspection$",
      "-count=1"
    ],
    "exit_code": 1,
    "named_failures": [
      "--- FAIL: TestCredentialStorePolicyNativeInspection (0.22s)",
      "--- FAIL: TestCredentialStorePolicyNativeInspection/permission (0.00s)"
    ]
  },
  {
    "name": "persist-alias",
    "command": [
      "go",
      "test",
      "-overlay",
      ".temp/TASK-261016-2en3yu/rev2/mutants/persist-alias/overlay.json",
      "./internal/config",
      "-run",
      "^TestCredentialStoreConfigGrammar$",
      "-count=1"
    ],
    "exit_code": 1,
    "named_failures": [
      "--- FAIL: TestCredentialStoreConfigGrammar (0.01s)",
      "--- FAIL: TestCredentialStoreConfigGrammar/alias_id (0.00s)"
    ]
  },
  {
    "name": "profile-normalization",
    "command": [
      "go",
      "test",
      "-overlay",
      ".temp/TASK-261016-2en3yu/rev2/mutants/profile-normalization/overlay.json",
      "./cmd/curator",
      "-run",
      "^TestCredentialStoreCLIConfigRoundTrip$",
      "-count=1"
    ],
    "exit_code": 1,
    "named_failures": [
      "--- FAIL: TestCredentialStoreCLIConfigRoundTrip (0.56s)"
    ]
  },
  {
    "name": "system-store",
    "command": [
      "go",
      "test",
      "-overlay",
      ".temp/TASK-261016-2en3yu/rev2/mutants/system-store/overlay.json",
      "./internal/config",
      "-run",
      "^TestCredentialStoreConfigMachineOwnership$",
      "-count=1"
    ],
    "exit_code": 1,
    "named_failures": [
      "--- FAIL: TestCredentialStoreConfigMachineOwnership (0.00s)"
    ]
  },
  {
    "name": "unknown-auto",
    "command": [
      "go",
      "test",
      "-overlay",
      ".temp/TASK-261016-2en3yu/rev2/mutants/unknown-auto/overlay.json",
      "./internal/config",
      "-run",
      "^TestCredentialStoreConfigGrammar$",
      "-count=1"
    ],
    "exit_code": 1,
    "named_failures": [
      "--- FAIL: TestCredentialStoreConfigGrammar (0.01s)",
      "--- FAIL: TestCredentialStoreConfigGrammar/unknown_selector (0.00s)"
    ]
  },
  {
    "name": "dotted-profile-only",
    "command": [
      "go",
      "test",
      "-overlay",
      ".temp/TASK-261016-2en3yu/rev2/mutants/dotted-profile-only/overlay.json",
      "./cmd/curator",
      "-run",
      "^TestCredentialStoreCLIDottedProfileRouting$",
      "-count=1"
    ],
    "exit_code": 1,
    "named_failures": [
      "--- FAIL: TestCredentialStoreCLIDottedProfileRouting (0.02s)",
      "--- FAIL: TestCredentialStoreCLIDottedProfileRouting/team.prod/codex (0.00s)",
      "--- FAIL: TestCredentialStoreCLIDottedProfileRouting/team.prod/codex_cli (0.00s)",
      "--- FAIL: TestCredentialStoreCLIDottedProfileRouting/team.prod/claude (0.00s)",
      "--- FAIL: TestCredentialStoreCLIDottedProfileRouting/team.prod/claude_code (0.00s)",
      "--- FAIL: TestCredentialStoreCLIDottedProfileRouting/Team..prod/codex (0.00s)",
      "--- FAIL: TestCredentialStoreCLIDottedProfileRouting/Team..prod/codex_cli (0.00s)",
      "--- FAIL: TestCredentialStoreCLIDottedProfileRouting/Team..prod/claude (0.00s)",
      "--- FAIL: TestCredentialStoreCLIDottedProfileRouting/Team..prod/claude_code (0.00s)",
      "--- FAIL: TestCredentialStoreCLIDottedProfileRouting/codex.prod/codex (0.00s)",
      "--- FAIL: TestCredentialStoreCLIDottedProfileRouting/codex.prod/codex_cli (0.00s)",
      "--- FAIL: TestCredentialStoreCLIDottedProfileRouting/codex.prod/claude (0.00s)",
      "--- FAIL: TestCredentialStoreCLIDottedProfileRouting/codex.prod/claude_code (0.00s)"
    ]
  },
  {
    "name": "reserved-profile-show",
    "command": [
      "go",
      "test",
      "-overlay",
      ".temp/TASK-261016-2en3yu/rev2/mutants/reserved-profile-show/overlay.json",
      "./cmd/curator",
      "-run",
      "^TestCredentialStoreCLIInvalidProfileRouting$/^CON[.]prod$/^show$",
      "-count=1"
    ],
    "exit_code": 1,
    "named_failures": [
      "--- FAIL: TestCredentialStoreCLIInvalidProfileRouting (0.00s)",
      "--- FAIL: TestCredentialStoreCLIInvalidProfileRouting/CON.prod/show (0.00s)"
    ]
  }
]

```

## regression-red-01.log

```text
--- FAIL: TestReviewDottedCredentialProfile (0.00s)
    --- FAIL: TestReviewDottedCredentialProfile/team.prod (0.00s)
        env_credential_store_test.go:197: valid profile refused: code=1 stdout= stderr=curator: environments.credential_store.team: unknown canonical environment "prod"
--- FAIL: TestReviewDottedCredentialProfileExisting (0.00s)
    --- FAIL: TestReviewDottedCredentialProfileExisting/show (0.00s)
        env_credential_store_test.go:218: existing dotted profile show: code=0 stdout=null
             stderr=
    --- FAIL: TestReviewDottedCredentialProfileExisting/unset (0.00s)
        env_credential_store_test.go:218: existing dotted profile unset: code=1 stdout= stderr=curator: environments knob "credential_store.team.prod.codex_cli" is not set
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
FAIL	github.com/relux-works/curator/cmd/curator	0.378s
FAIL

```

## focused-04.log

```text
ok  	github.com/relux-works/curator/internal/config	0.187s
ok  	github.com/relux-works/curator/internal/envregistry	0.157s [no tests to run]
ok  	github.com/relux-works/curator/internal/envprofile	0.888s
ok  	github.com/relux-works/curator/cmd/curator	1.645s

```

## config-registry-07.log

```text
ok  	github.com/relux-works/curator/internal/config	1.192s
ok  	github.com/relux-works/curator/internal/envregistry	0.435s

```

## build-06.log

```text
(empty output)

```

## vet-08.log

```text
(empty output)

```

## lint-10.log

```text
0 issues.

```

## diff-check-09.log

```text
(empty output)

```

## Mutant dotted-profile-only: actual expected-red output (exit 1)

```text
--- FAIL: TestCredentialStoreCLIDottedProfileRouting (0.02s)
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
FAIL	github.com/relux-works/curator/cmd/curator	0.573s
FAIL

```

## Mutant drop-codex-file: actual expected-red output (exit 1)

```text
--- FAIL: TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery (0.59s)
    env_credential_store_test.go:168: unqualified file = 0 {"env":{"CODEX_HOME":"/var/folders/cz/jqbthtks55zbkpcdkfyk4bl80000gp/T/TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery4124389845/001/environments/acme/codex_cli"},"environment":"codex_cli","fragment":"launch-env-fragment-v2","permissions":{"locked":false,"mode":"native","source":"default"},"precedence":{"placement":"winner-last","winner":"higher-weight"},"profile":{"lock_sha256":"e13ef5a9cd5f1afaedb4c29010a731b26caed36f798fb6396b14db3ea456ea9d","name":"acme"}}
         warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults
        warning: environment_tool_version_unverified: codex_cli detected 0.162.1, recorded 0.153.2
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	1.296s
FAIL

```

## Mutant old-v3: actual expected-red output (exit 1)

```text
--- FAIL: TestCredentialStoreConfigVersionBoundary (0.00s)
    credential_store_test.go:57: err = <nil>, want mention of "credential_store"
FAIL
FAIL	github.com/relux-works/curator/internal/config	0.278s
FAIL

```

## Mutant permission-as-absent: actual expected-red output (exit 1)

```text
--- FAIL: TestCredentialStorePolicyNativeInspection (0.22s)
    --- FAIL: TestCredentialStorePolicyNativeInspection/permission (0.00s)
        credential_policy_test.go:81: Resolve = <nil>, codex_settings_unsupported: explicit Codex file storage requires qualified login/app-server/refresh behavior and strict successor fragment consumers; no native fallback; want environment_seed_unreadable
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	0.560s
FAIL

```

## Mutant persist-alias: actual expected-red output (exit 1)

```text
--- FAIL: TestCredentialStoreConfigGrammar (0.01s)
    --- FAIL: TestCredentialStoreConfigGrammar/alias_id (0.00s)
        credential_store_test.go:40: Load = <nil>, want credential_store refusal
FAIL
FAIL	github.com/relux-works/curator/internal/config	0.279s
FAIL

```

## Mutant profile-normalization: actual expected-red output (exit 1)

```text
--- FAIL: TestCredentialStoreCLIConfigRoundTrip (0.56s)
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
          "skills_root": "/var/folders/cz/jqbthtks55zbkpcdkfyk4bl80000gp/T/TestCredentialStoreCLIConfigRoundTrip2984612345/001/skills"
        }
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	0.920s
FAIL

```

## Mutant system-store: actual expected-red output (exit 1)

```text
--- FAIL: TestCredentialStoreConfigMachineOwnership (0.00s)
    credential_store_test.go:75: system store accepted: <nil>
FAIL
FAIL	github.com/relux-works/curator/internal/config	0.274s
FAIL

```

## Mutant unknown-auto: actual expected-red output (exit 1)

```text
--- FAIL: TestCredentialStoreConfigGrammar (0.01s)
    --- FAIL: TestCredentialStoreConfigGrammar/unknown_selector (0.00s)
        credential_store_test.go:40: Load = <nil>, want credential_store refusal
FAIL
FAIL	github.com/relux-works/curator/internal/config	0.284s
FAIL

```

## Mutant reserved-profile-show: actual expected-red output (exit 1)

```text
--- FAIL: TestCredentialStoreCLIInvalidProfileRouting (0.00s)
    --- FAIL: TestCredentialStoreCLIInvalidProfileRouting/CON.prod/show (0.00s)
        env_credential_store_test.go:296: invalid profile = 0 "native"
             
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	0.363s
FAIL

```
