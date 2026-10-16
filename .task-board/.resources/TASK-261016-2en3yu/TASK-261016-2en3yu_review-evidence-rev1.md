# Revision 1 reviewer evidence

Extract replay.py and attacks.go.txt into the same directory and run the script from the candidate repository. Remaining files are named command logs, identity records and mutant definitions. All source changes were Go overlays under .temp only.

## attacks.go.txt

```text
// A valid dotted profile must remain one profile in the public selector CLI.
func TestReviewDottedCredentialProfile(t *testing.T) {
 for _, profile := range []string{"team", "team.prod"} {
  t.Run(profile, func(t *testing.T) {
   source := writeMachineConfig(t, `{"schema_version":5,"skills_root":"x","projects":{}}`)
   code, stdout, stderr := runProfile(t, source, "env", "config", "set", "credential_store."+profile+".codex", "file")
   if code != exitOK { t.Fatalf("valid profile refused: code=%d stdout=%s stderr=%s", code, stdout, stderr) }
   source = reloadSource(t, source)
   code, stdout, stderr = runProfile(t, source, "env", "config", "show", "credential_store."+profile+".codex")
   if code != exitOK || strings.TrimSpace(stdout) != `"file"` { t.Fatalf("show=%d %s %s",code,stdout,stderr) }
   if code, _, stderr := runProfile(t, source, "env", "config", "unset", "credential_store."+profile+".codex"); code != exitOK { t.Fatal(stderr) }
  })
 }
}

// Read/unset must address a previously persisted valid dotted profile.
func TestReviewDottedCredentialProfileExisting(t *testing.T) {
 for _, op := range []string{"show", "unset"} {
  t.Run(op, func(t *testing.T) {
   source := writeMachineConfig(t, `{"schema_version":5,"skills_root":"x","projects":{},"environments":{"credential_store":{"team.prod":{"codex_cli":"file"}}}}`)
   code, stdout, stderr := runProfile(t, source, "env", "config", op, "credential_store.team.prod.codex_cli")
   if code != exitOK || (op == "show" && strings.TrimSpace(stdout) != `"file"`) { t.Fatalf("existing dotted profile %s: code=%d stdout=%s stderr=%s", op,code,stdout,stderr) }
  })
 }
}

```

## config-registry.log

```text
ok  	github.com/relux-works/curator/internal/config	0.340s
ok  	github.com/relux-works/curator/internal/envregistry	0.146s

```

## dotted-profile.log

```text
=== RUN   TestReviewDottedCredentialProfile
=== RUN   TestReviewDottedCredentialProfile/team
=== RUN   TestReviewDottedCredentialProfile/team.prod
    env_credential_store_test.go:196: valid profile refused: code=1 stdout= stderr=curator: environments.credential_store.team: unknown canonical environment "prod"
--- FAIL: TestReviewDottedCredentialProfile (0.00s)
    --- PASS: TestReviewDottedCredentialProfile/team (0.00s)
    --- FAIL: TestReviewDottedCredentialProfile/team.prod (0.00s)
=== RUN   TestReviewDottedCredentialProfileExisting
=== RUN   TestReviewDottedCredentialProfileExisting/show
    env_credential_store_test.go:211: existing dotted profile show: code=0 stdout=null
         stderr=
=== RUN   TestReviewDottedCredentialProfileExisting/unset
    env_credential_store_test.go:211: existing dotted profile unset: code=1 stdout= stderr=curator: environments knob "credential_store.team.prod.codex_cli" is not set
--- FAIL: TestReviewDottedCredentialProfileExisting (0.00s)
    --- FAIL: TestReviewDottedCredentialProfileExisting/show (0.00s)
    --- FAIL: TestReviewDottedCredentialProfileExisting/unset (0.00s)
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	0.689s
FAIL

```

## drop-codex-file.log

```text
=== RUN   TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery
    env_credential_store_test.go:168: unqualified file = 0 {"env":{"CODEX_HOME":"<temporary-fixture-path>"},"environment":"codex_cli","fragment":"launch-env-fragment-v2","permissions":{"locked":false,"mode":"native","source":"default"},"precedence":{"placement":"winner-last","winner":"higher-weight"},"profile":{"lock_sha256":"e13ef5a9cd5f1afaedb4c29010a731b26caed36f798fb6396b14db3ea456ea9d","name":"acme"}}
         warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults
        warning: environment_tool_version_unverified: codex_cli detected 0.162.1, recorded 0.153.2
--- FAIL: TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery (0.53s)
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	0.902s
FAIL

```

## error-surface.log

```text
=== RUN   TestCredentialStorePolicyNativeInspection
=== RUN   TestCredentialStorePolicyNativeInspection/absent_key
=== RUN   TestCredentialStorePolicyNativeInspection/keyring
=== RUN   TestCredentialStorePolicyNativeInspection/auto
=== RUN   TestCredentialStorePolicyNativeInspection/unknown
=== RUN   TestCredentialStorePolicyNativeInspection/malformed
=== RUN   TestCredentialStorePolicyNativeInspection/permission
--- PASS: TestCredentialStorePolicyNativeInspection (0.18s)
    --- PASS: TestCredentialStorePolicyNativeInspection/absent_key (0.00s)
    --- PASS: TestCredentialStorePolicyNativeInspection/keyring (0.00s)
    --- PASS: TestCredentialStorePolicyNativeInspection/auto (0.00s)
    --- PASS: TestCredentialStorePolicyNativeInspection/unknown (0.00s)
    --- PASS: TestCredentialStorePolicyNativeInspection/malformed (0.00s)
    --- PASS: TestCredentialStorePolicyNativeInspection/permission (0.00s)
PASS
ok  	github.com/relux-works/curator/internal/envprofile	0.420s

```

## focused.log

```text
ok  	github.com/relux-works/curator/internal/config	0.310s
ok  	github.com/relux-works/curator/internal/envregistry	0.256s [no tests to run]
ok  	github.com/relux-works/curator/internal/envprofile	1.218s
ok  	github.com/relux-works/curator/cmd/curator	1.685s

```

## hosted-readback.json

```text
{"conclusion":"success","headSha":"8c55bff02936d64706bf97f34495aac758e11d2a","status":"completed","url":"https://github.com/relux-works/curator/actions/runs/38852843558"}

```

## identity.txt

```text
Candidate tree 7efa9fa6153d25c5cf051aaa77661ddf9a1803e7
16 of 16 changed files match exact candidate blobs.
Remote main equals base dccac3199aa5c543c589c129f5ef2bc437523a2a.
Remote gate snapshot 5f47cc8337d0f96f363ac41db4166dcb485970b1 has identical candidate tree.

```

## mutant-definitions.json

```text
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
  }
]

```

## mutant-results.json

```text
[
  {
    "name": "drop-codex-file",
    "exit": 1,
    "log": "drop-codex-file/result.log"
  },
  {
    "name": "old-v3",
    "exit": 1,
    "log": "old-v3/result.log"
  },
  {
    "name": "permission-as-absent",
    "exit": 1,
    "log": "permission-as-absent/result.log"
  },
  {
    "name": "persist-alias",
    "exit": 1,
    "log": "persist-alias/result.log"
  },
  {
    "name": "profile-normalization",
    "exit": 1,
    "log": "profile-normalization/result.log"
  },
  {
    "name": "system-store",
    "exit": 1,
    "log": "system-store/result.log"
  },
  {
    "name": "unknown-auto",
    "exit": 1,
    "log": "unknown-auto/result.log"
  }
]
```

## old-v3.log

```text
=== RUN   TestCredentialStoreConfigVersionBoundary
    credential_store_test.go:57: err = <nil>, want mention of "credential_store"
--- FAIL: TestCredentialStoreConfigVersionBoundary (0.00s)
FAIL
FAIL	github.com/relux-works/curator/internal/config	0.273s
FAIL

```

## permission-as-absent.log

```text
=== RUN   TestCredentialStorePolicyNativeInspection
=== RUN   TestCredentialStorePolicyNativeInspection/absent_key
=== RUN   TestCredentialStorePolicyNativeInspection/keyring
=== RUN   TestCredentialStorePolicyNativeInspection/auto
=== RUN   TestCredentialStorePolicyNativeInspection/unknown
=== RUN   TestCredentialStorePolicyNativeInspection/malformed
=== RUN   TestCredentialStorePolicyNativeInspection/permission
    credential_policy_test.go:81: Resolve = <nil>, codex_settings_unsupported: explicit Codex file storage requires qualified login/app-server/refresh behavior and strict successor fragment consumers; no native fallback; want environment_seed_unreadable
--- FAIL: TestCredentialStorePolicyNativeInspection (0.26s)
    --- PASS: TestCredentialStorePolicyNativeInspection/absent_key (0.00s)
    --- PASS: TestCredentialStorePolicyNativeInspection/keyring (0.00s)
    --- PASS: TestCredentialStorePolicyNativeInspection/auto (0.00s)
    --- PASS: TestCredentialStorePolicyNativeInspection/unknown (0.00s)
    --- PASS: TestCredentialStorePolicyNativeInspection/malformed (0.00s)
    --- FAIL: TestCredentialStorePolicyNativeInspection/permission (0.00s)
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	0.610s
FAIL

```

## persist-alias.log

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
    credential_store_test.go:40: Load = <nil>, want credential_store refusal
=== RUN   TestCredentialStoreConfigGrammar/null_selector
=== RUN   TestCredentialStoreConfigGrammar/unknown_selector
=== RUN   TestCredentialStoreConfigGrammar/wrong_type
--- FAIL: TestCredentialStoreConfigGrammar (0.01s)
    --- PASS: TestCredentialStoreConfigGrammar/empty (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/all_selectors (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/null_map (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/array_map (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/null_profile (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/invalid_profile (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/unknown_id (0.00s)
    --- FAIL: TestCredentialStoreConfigGrammar/alias_id (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/null_selector (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/unknown_selector (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/wrong_type (0.00s)
FAIL
FAIL	github.com/relux-works/curator/internal/config	0.275s
FAIL

```

## profile-normalization.log

```text
=== RUN   TestCredentialStoreCLIConfigRoundTrip
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
          "skills_root": "<temporary-fixture-path>"
        }
--- FAIL: TestCredentialStoreCLIConfigRoundTrip (0.58s)
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	0.949s
FAIL

```

## replay.py

```text
"""Run from the candidate repository; all generated files stay under .temp."""
from pathlib import Path
import json
import subprocess
import sys
root = Path.cwd()
bundle = Path(__file__).resolve().parent
out = root / '.temp' / 'l151-review-replay'
out.mkdir(parents=True, exist_ok=True)
source = root / 'cmd/curator/env_credential_store_test.go'
replacement = out / 'attack_test.go'
replacement.write_text(source.read_text() + '\n' + (bundle / 'attacks.go.txt').read_text())
overlay = out / 'overlay.json'
overlay.write_text(json.dumps({'Replace': {str(source): str(replacement)}}))
command = ['go', 'test', '-overlay', str(overlay), './cmd/curator', '-run', '^TestReviewDottedCredentialProfile', '-count=1', '-v']
sys.exit(subprocess.run(command).returncode)

```

## rerun-final.json

```text
{"conclusion":"","headSha":"5f47cc8337d0f96f363ac41db4166dcb485970b1","status":"in_progress","url":"https://github.com/relux-works/curator/actions/runs/38864359459"}

```

## system-store.log

```text
=== RUN   TestCredentialStoreConfigMachineOwnership
    credential_store_test.go:75: system store accepted: <nil>
--- FAIL: TestCredentialStoreConfigMachineOwnership (0.00s)
FAIL
FAIL	github.com/relux-works/curator/internal/config	0.280s
FAIL

```

## unknown-auto.log

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
    credential_store_test.go:40: Load = <nil>, want credential_store refusal
=== RUN   TestCredentialStoreConfigGrammar/wrong_type
--- FAIL: TestCredentialStoreConfigGrammar (0.01s)
    --- PASS: TestCredentialStoreConfigGrammar/empty (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/all_selectors (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/null_map (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/array_map (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/null_profile (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/invalid_profile (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/unknown_id (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/alias_id (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/null_selector (0.00s)
    --- FAIL: TestCredentialStoreConfigGrammar/unknown_selector (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/wrong_type (0.00s)
FAIL
FAIL	github.com/relux-works/curator/internal/config	0.282s
FAIL

```

## verdict.json

```text
{
  "verdict": "changes_requested",
  "revision": 1,
  "candidate_tree": "7efa9fa6153d25c5cf051aaa77661ddf9a1803e7",
  "findings": [
    {
      "id": "dotted-profile-selector-path",
      "row": "selector paths",
      "invariant": "Every valid profile identifier, including team.prod, remains one profile operand in credential_store.<profile>.<env-id> set/show/unset; show reports the selected value.",
      "mechanism": "internal/config/environments.go:1360 splits every dot into a path segment; cmd/curator/envalias.go:35 normalizes segment 2 and cmd/curator/envconfig.go:70 handles effective selection only for exactly three segments. The new selector CLI therefore cannot address valid dotted profile names.",
      "reproductions": [
        {
          "test_file": "attacks.go.txt (overlaid on cmd/curator/env_credential_store_test.go)",
          "tests": [
            "TestReviewDottedCredentialProfile",
            "TestReviewDottedCredentialProfileExisting"
          ],
          "command": "python3 replay.py (from candidate repository; replay.py and attacks.go.txt from the evidence bundle)",
          "expected_failure": "Exit 1: team positive control passes; team.prod set refuses unknown canonical environment prod; existing team.prod show returns null instead of file; unset refuses not set.",
          "log": "dotted-profile.log",
          "candidate_tree": "7efa9fa6153d25c5cf051aaa77661ddf9a1803e7"
        }
      ],
      "severity": "bypass",
      "priority": "P2",
      "repeat-of": "none"
    }
  ],
  "notes": [
    "The same path mechanism also causes valid set/unset regressions. One finding covers all three symptoms.",
    "Slice 1 only: effective backend application, runtime qualification, login/readiness and hosted override transport remain explicitly deferred.",
    "No review/free-hunt time budget was supplied. Supplemental producer surface table was used in its given order.",
    "Native permission failure attack uses the existing read seam; this is not a physical permission-mode qualification."
  ]
}
```

## Final hosted readback (supersedes earlier in-progress snapshot)

```json
{"conclusion":"cancelled","headSha":"5f47cc8337d0f96f363ac41db4166dcb485970b1","status":"completed","url":"https://github.com/relux-works/curator/actions/runs/38864359459"}

```
