# TASK-261016-i5gwaf — local default-suite diagnostic (exit 1)

`go test ./...` exited 1. Five status failures pass with a clean PATH;
three packages reached the local default 10m timeout and remain unverified
by this local command. No green default-suite claim derives from this log.
The mandatory hosted gate is separate evidence. This is the complete local
log with personal/scratch paths redacted, not a shortened passing log.

Original private log SHA-256: `af585e7754200baf1d229b581d01405bf150be5d0b2b58cace107d0f28db7c21`.

````text
--- FAIL: TestDraftStatusReportsUpToDateThroughCLI (5.32s)
    draft_status_test.go:140: status --check = 1, want 0
        stderr:
        warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults
--- FAIL: TestDraftStatusNeedsInstallAfterRefreshWithoutInstall (10.40s)
    draft_status_test.go:185: status --check after install = 1
        stdout:
        app: review up-to-date
        security_posture: permissive (profile)
        hook-trust: A-warning (shipped)
        registry-policy: advisory (profile)
        audit-mode: advisory (profile)
        source-allowlist: 0 (profile)
        shell-hook-trust: $TEST_PATH approved (approved_by=manager)
        shell-hook-trust: $TEST_PATH approved (approved_by=manager)
        provider engines: $PRIVATE_PATH (subcommand_provider_untrusted, non-current)
        provider run: $TEST_PATH (subcommand_provider_outside_trust_roots, current)
        provider session: $TEST_PATH (subcommand_provider_outside_trust_roots, current)
        
        stderr:
        warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults
--- FAIL: TestDraftStatusLiveMutationWithoutRefreshStaysCurrent (5.81s)
    draft_status_test.go:403: status --check = 1, want 0 without a refresh
        stdout:
        app: review up-to-date
        security_posture: permissive (profile)
        hook-trust: A-warning (shipped)
        registry-policy: advisory (profile)
        audit-mode: advisory (profile)
        source-allowlist: 0 (profile)
        shell-hook-trust: $TEST_PATH approved (approved_by=manager)
        shell-hook-trust: $TEST_PATH approved (approved_by=manager)
        provider engines: $PRIVATE_PATH (subcommand_provider_untrusted, non-current)
        provider run: $TEST_PATH (subcommand_provider_outside_trust_roots, current)
        provider session: $TEST_PATH (subcommand_provider_outside_trust_roots, current)
        
        stderr:
        warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults
--- FAIL: TestEnvStatusMatrix (15.17s)
    env_test.go:330: status --check after repair = 1
        security_posture: permissive (profile)
        hook-trust: A-warning (shipped)
        registry-policy:  (profile)
        audit-mode:  (profile)
        source-allowlist: 0 (profile)
        env-passthrough: s4-warn (shipped)
        transitive-system-modules:  (profile)
        provider-trust-roots: revision-A (shipped)
        source-signers: false (profile)
        update-confirmation: A-warning (shipped)
        codex-seed: B (shipped)
        store-boundary: enforced (shipped)
        write-discipline: enforced (shipped)
        mcp-package-allowlist: 0 (profile)
        require_source_signers: false
        update_confirmation: B-flip (a triggered update delta refuses profile_update_confirmation_required unless --confirm-system-delta is given)
        s4_profile: s4-warn, passable_env_names: unbounded
        codex-seed: revision B (shipped)
        warning: mcp_package_allowlist_empty: the MCP package allowlist is empty; every declaration package in the closure is admitted
        acme claude_code: current, provisioned, mode managed-home, form monolithic, lock sha256:e13ef5a9cd5f…
          surface root-context: current
          surface skills: current
          warning: environment_tool_version_unverified: claude_code detected 2.1.301, recorded 2.1.261
          passthrough: per-home-keychain
          seeds: .claude.json
          seeded-projects: $PRIVATE_PATH
        acme codex_cli: current, provisioned, mode managed-home, form monolithic, lock sha256:e13ef5a9cd5f…
          surface root-context: current
          surface skills: current
          warning: environment_tool_version_unverified: codex_cli detected 0.162.1, recorded 0.153.2
          passthrough: auth.json (keyring-preferred)
          codex-seed-record: revision B; native MCP servers none (none)
        acme muse: current, provisioned, mode managed-home, form monolithic, lock sha256:e13ef5a9cd5f…
          surface root-context: current
          surface skills: current
          warning: environment_tool_version_unverified: muse detected 1.4.1, recorded 1.4.7
          passthrough: config/muse/auth.json (file-link)
        acme opencode: current, provisioned, mode managed-home, form monolithic, lock sha256:e13ef5a9cd5f…
          surface root-context: current
        acme pi: current, provisioned, mode managed-home, form monolithic, lock sha256:e13ef5a9cd5f…
          surface root-context: current
          surface skills: current
          passthrough: auth.json (file-link)
        default claude_code: current, provisioned, mode managed-home, form , lock sha256:06fb1543fb69…
          surface skills: current
          warning: environment_tool_version_unverified: claude_code detected 2.1.301, recorded 2.1.261
          passthrough: per-home-keychain
          seeds: .claude.json
          seeded-projects: $PRIVATE_PATH
        default codex_cli: current, provisioned, mode managed-home, form , lock sha256:06fb1543fb69…
          surface skills: current
          warning: environment_tool_version_unverified: codex_cli detected 0.162.1, recorded 0.153.2
          passthrough: auth.json (keyring-preferred)
          codex-seed-record: revision B; native MCP servers none (none)
        default muse: current, provisioned, mode managed-home, form , lock sha256:06fb1543fb69…
          surface skills: current
          warning: environment_tool_version_unverified: muse detected 1.4.1, recorded 1.4.7
          passthrough: config/muse/auth.json (file-link)
        default opencode: current, provisioned, mode managed-home, form , lock sha256:06fb1543fb69…
        default pi: current, provisioned, mode managed-home, form , lock sha256:06fb1543fb69…
          surface skills: current
          passthrough: auth.json (file-link)
        scope machine: profile acme env claude_code native $TEST_PATH managed $TEST_PATH (provisioned)
        scope machine: profile acme env codex_cli native $TEST_PATH managed $TEST_PATH (provisioned)
        scope machine: profile acme env muse native $TEST_PATH managed $TEST_PATH (provisioned)
        scope machine: profile acme env opencode native $TEST_PATH managed $TEST_PATH (provisioned)
        scope machine: profile acme env pi native $TEST_PATH managed $TEST_PATH (provisioned)
        tool muse: recorded 1.4.7 detected 1.4.1
        tool claude_code: recorded 2.1.261 detected 2.1.301
        tool codex_cli: recorded 0.153.2 detected 0.162.1
        tool opencode: recorded unrecorded detected 1.18.30
        tool pi: recorded 0.84.2 detected 0.84.2
        provider engines: $PRIVATE_PATH (subcommand_provider_untrusted, non-current)
        provider run: $TEST_PATH (subcommand_provider_outside_trust_roots, current)
        provider session: $TEST_PATH (subcommand_provider_outside_trust_roots, current)
        target xcode-coding-assistant (claude_code): participating=false (auto: probe path absent, nothing materialized); the embedded host's MCP configuration and commands/ are present, unaudited, and outside this capability
        target xcode-coding-assistant (codex_cli): participating=false (auto: probe path absent, nothing materialized); the embedded host's MCP configuration and commands/ are present, unaudited, and outside this capability
        profile acme: lock sha256:e13ef5a9cd5f…, precedence winner=higher-weight placement=winner-last, transitive_system_modules=drop
          member context acme weight 0
        profile default: lock sha256:06fb1543fb69…, precedence winner=higher-weight placement=winner-last, transitive_system_modules=drop
          member context default weight 0
        note: Muse: monolithic root context delivered at config/muse/AGENTS.md (verified on 1.4.7); managed exec requires --no-foreign-personal-context when the target is materialized; serve cannot exclude foreign personal context (foreign-personal-context isolation gap); workspace AGENTS.md prevailed in observed conflicts; model-visible foreign removal, skills discovery, system-prompt discovery, refresh semantics, and state/cache internal layouts remain unverified
        note: opencode skills come from the machine-current profile, split-brain by construction (§7.1)
--- FAIL: TestEnvStatusRegistryBoundaryPostureAndCheck (11.25s)
    env_test.go:415: first-use registry status --check = 1
        stdout:
        {
          "homes": [
            {
              "profile": "acme",
              "environment": "claude_code",
              "mode": "managed-home",
              "form": "monolithic",
              "home": "$TEST_PATH",
              "provisioned": true,
              "current": true,
              "lock_hash": "sha256:e13ef5a9cd5f1afaedb4c29010a731b26caed36f798fb6396b14db3ea456ea9d",
              "marker_hash": "e13ef5a9cd5f1afaedb4c29010a731b26caed36f798fb6396b14db3ea456ea9d",
              "surfaces": [
                {
                  "key": "root-context",
                  "paths": [
                    "CLAUDE.md"
                  ],
                  "form": "monolithic",
                  "state": "current",
                  "detail": "current"
                },
                {
                  "key": "skills",
                  "paths": [],
                  "form": "",
                  "state": "current",
                  "detail": "current"
                }
              ],
              "passthrough": [
                "per-home-keychain"
              ],
              "seeds": [
                ".claude.json"
              ],
              "seed_links": [],
              "seeded_projects": [
                "$PRIVATE_PATH"
              ],
              "native_mcp_servers_disposition": "none",
              "backups_known": true,
              "backups": 0,
              "backups_oldest": "-",
              "backups_newest": "-",
              "findings": null,
              "warnings": [
                "environment_tool_version_unverified: claude_code detected 2.1.301, recorded 2.1.261"
              ]
            },
            {
              "profile": "acme",
              "environment": "codex_cli",
              "mode": "managed-home",
              "form": "monolithic",
              "home": "$TEST_PATH",
              "provisioned": true,
              "current": true,
              "lock_hash": "sha256:e13ef5a9cd5f1afaedb4c29010a731b26caed36f798fb6396b14db3ea456ea9d",
              "marker_hash": "e13ef5a9cd5f1afaedb4c29010a731b26caed36f798fb6396b14db3ea456ea9d",
              "surfaces": [
                {
                  "key": "root-context",
                  "paths": [
                    "AGENTS.md"
                  ],
                  "form": "monolithic",
                  "state": "current",
                  "detail": "current"
                },
                {
                  "key": "skills",
                  "paths": [],
                  "form": "",
                  "state": "current",
                  "detail": "current"
                }
              ],
              "passthrough": [
                "auth.json (keyring-preferred)"
              ],
              "seeds": [],
              "seed_links": [],
              "seeded_projects": [],
              "codex_seed_record": {
                "revision": "B",
                "native_mcp_servers": []
              },
              "native_mcp_servers_disposition": "none",
              "backups_known": true,
              "backups": 0,
              "backups_oldest": "-",
              "backups_newest": "-",
              "findings": null,
              "warnings": [
                "environment_tool_version_unverified: codex_cli detected 0.162.1, recorded 0.153.2"
              ]
            },
            {
              "profile": "acme",
              "environment": "muse",
              "mode": "managed-home",
              "form": "monolithic",
              "home": "$TEST_PATH",
              "provisioned": true,
              "current": true,
              "lock_hash": "sha256:e13ef5a9cd5f1afaedb4c29010a731b26caed36f798fb6396b14db3ea456ea9d",
              "marker_hash": "e13ef5a9cd5f1afaedb4c29010a731b26caed36f798fb6396b14db3ea456ea9d",
              "surfaces": [
                {
                  "key": "root-context",
                  "paths": [
                    "config/muse/AGENTS.md"
                  ],
                  "form": "monolithic",
                  "state": "current",
                  "detail": "current"
                },
                {
                  "key": "skills",
                  "paths": [],
                  "form": "",
                  "state": "current",
                  "detail": "current"
                }
              ],
              "passthrough": [
                "config/muse/auth.json (file-link)"
              ],
              "seeds": [],
              "seed_links": [],
              "seeded_projects": [],
              "native_mcp_servers_disposition": "none",
              "backups_known": true,
              "backups": 0,
              "backups_oldest": "-",
              "backups_newest": "-",
              "findings": null,
              "warnings": [
                "environment_tool_version_unverified: muse detected 1.4.1, recorded 1.4.7"
              ]
            },
            {
              "profile": "acme",
              "environment": "opencode",
              "mode": "managed-home",
              "form": "monolithic",
              "home": "$TEST_PATH",
              "provisioned": true,
              "current": true,
              "lock_hash": "sha256:e13ef5a9cd5f1afaedb4c29010a731b26caed36f798fb6396b14db3ea456ea9d",
              "marker_hash": "e13ef5a9cd5f1afaedb4c29010a731b26caed36f798fb6396b14db3ea456ea9d",
              "surfaces": [
                {
                  "key": "root-context",
                  "paths": [
                    "AGENTS.md"
                  ],
                  "form": "monolithic",
                  "state": "current",
                  "detail": "current"
                }
              ],
              "passthrough": null,
              "seeds": [],
              "seed_links": [],
              "seeded_projects": [],
              "native_mcp_servers_disposition": "none",
              "backups_known": true,
              "backups": 0,
              "backups_oldest": "-",
              "backups_newest": "-",
              "findings": null,
              "warnings": null
            },
            {
              "profile": "acme",
              "environment": "pi",
              "mode": "managed-home",
              "form": "monolithic",
              "home": "$TEST_PATH",
              "provisioned": true,
              "current": true,
              "lock_hash": "sha256:e13ef5a9cd5f1afaedb4c29010a731b26caed36f798fb6396b14db3ea456ea9d",
              "marker_hash": "e13ef5a9cd5f1afaedb4c29010a731b26caed36f798fb6396b14db3ea456ea9d",
              "surfaces": [
                {
                  "key": "root-context",
                  "paths": [
                    "AGENTS.md"
                  ],
                  "form": "monolithic",
                  "state": "current",
                  "detail": "current"
                },
                {
                  "key": "skills",
                  "paths": [],
                  "form": "",
                  "state": "current",
                  "detail": "current"
                }
              ],
              "passthrough": [
                "auth.json (file-link)"
              ],
              "seeds": [],
              "seed_links": [],
              "seeded_projects": [],
              "native_mcp_servers_disposition": "none",
              "backups_known": true,
              "backups": 0,
              "backups_oldest": "-",
              "backups_newest": "-",
              "findings": null,
              "warnings": null
            },
            {
              "profile": "default",
              "environment": "claude_code",
              "mode": "managed-home",
              "form": "",
              "home": "$TEST_PATH",
              "provisioned": true,
              "current": true,
              "lock_hash": "sha256:06fb1543fb69055047d8f34bbac2901f06b4da1374cdbbbc47a26d22126362a4",
              "marker_hash": "06fb1543fb69055047d8f34bbac2901f06b4da1374cdbbbc47a26d22126362a4",
              "surfaces": [
                {
                  "key": "skills",
                  "paths": [],
                  "form": "",
                  "state": "current",
                  "detail": "current"
                }
              ],
              "passthrough": [
                "per-home-keychain"
              ],
              "seeds": [
                ".claude.json"
              ],
              "seed_links": [],
              "seeded_projects": [
                "$PRIVATE_PATH"
              ],
              "native_mcp_servers_disposition": "none",
              "backups_known": true,
              "backups": 0,
              "backups_oldest": "-",
              "backups_newest": "-",
              "findings": null,
              "warnings": [
                "environment_tool_version_unverified: claude_code detected 2.1.301, recorded 2.1.261"
              ]
            },
            {
              "profile": "default",
              "environment": "codex_cli",
              "mode": "managed-home",
              "form": "",
              "home": "$TEST_PATH",
              "provisioned": true,
              "current": true,
              "lock_hash": "sha256:06fb1543fb69055047d8f34bbac2901f06b4da1374cdbbbc47a26d22126362a4",
              "marker_hash": "06fb1543fb69055047d8f34bbac2901f06b4da1374cdbbbc47a26d22126362a4",
              "surfaces": [
                {
                  "key": "skills",
                  "paths": [],
                  "form": "",
                  "state": "current",
                  "detail": "current"
                }
              ],
              "passthrough": [
                "auth.json (keyring-preferred)"
              ],
              "seeds": [],
              "seed_links": [],
              "seeded_projects": [],
              "codex_seed_record": {
                "revision": "B",
                "native_mcp_servers": []
              },
              "native_mcp_servers_disposition": "none",
              "backups_known": true,
              "backups": 0,
              "backups_oldest": "-",
              "backups_newest": "-",
              "findings": null,
              "warnings": [
                "environment_tool_version_unverified: codex_cli detected 0.162.1, recorded 0.153.2"
              ]
            },
            {
              "profile": "default",
              "environment": "muse",
              "mode": "managed-home",
              "form": "",
              "home": "$TEST_PATH",
              "provisioned": true,
              "current": true,
              "lock_hash": "sha256:06fb1543fb69055047d8f34bbac2901f06b4da1374cdbbbc47a26d22126362a4",
              "marker_hash": "06fb1543fb69055047d8f34bbac2901f06b4da1374cdbbbc47a26d22126362a4",
              "surfaces": [
                {
                  "key": "skills",
                  "paths": [],
                  "form": "",
                  "state": "current",
                  "detail": "current"
                }
              ],
              "passthrough": [
                "config/muse/auth.json (file-link)"
              ],
              "seeds": [],
              "seed_links": [],
              "seeded_projects": [],
              "native_mcp_servers_disposition": "none",
              "backups_known": true,
              "backups": 0,
              "backups_oldest": "-",
              "backups_newest": "-",
              "findings": null,
              "warnings": [
                "environment_tool_version_unverified: muse detected 1.4.1, recorded 1.4.7"
              ]
            },
            {
              "profile": "default",
              "environment": "opencode",
              "mode": "managed-home",
              "form": "",
              "home": "$TEST_PATH",
              "provisioned": true,
              "current": true,
              "lock_hash": "sha256:06fb1543fb69055047d8f34bbac2901f06b4da1374cdbbbc47a26d22126362a4",
              "marker_hash": "06fb1543fb69055047d8f34bbac2901f06b4da1374cdbbbc47a26d22126362a4",
              "surfaces": null,
              "passthrough": null,
              "seeds": [],
              "seed_links": [],
              "seeded_projects": [],
              "native_mcp_servers_disposition": "none",
              "backups_known": true,
              "backups": 0,
              "backups_oldest": "-",
              "backups_newest": "-",
              "findings": null,
              "warnings": null
            },
            {
              "profile": "default",
              "environment": "pi",
              "mode": "managed-home",
              "form": "",
              "home": "$TEST_PATH",
              "provisioned": true,
              "current": true,
              "lock_hash": "sha256:06fb1543fb69055047d8f34bbac2901f06b4da1374cdbbbc47a26d22126362a4",
              "marker_hash": "06fb1543fb69055047d8f34bbac2901f06b4da1374cdbbbc47a26d22126362a4",
              "surfaces": [
                {
                  "key": "skills",
                  "paths": [],
                  "form": "",
                  "state": "current",
                  "detail": "current"
                }
              ],
              "passthrough": [
                "auth.json (file-link)"
              ],
              "seeds": [],
              "seed_links": [],
              "seeded_projects": [],
              "native_mcp_servers_disposition": "none",
              "backups_known": true,
              "backups": 0,
              "backups_oldest": "-",
              "backups_newest": "-",
              "findings": null,
              "warnings": null
            }
          ],
          "scopes": [
            {
              "scope": "machine",
              "profile": "acme",
              "environment": "claude_code",
              "native": "$TEST_PATH",
              "managed": "$TEST_PATH",
              "provisioned": true,
              "provisioned_known": true
            },
            {
              "scope": "machine",
              "profile": "acme",
              "environment": "codex_cli",
              "native": "$TEST_PATH",
              "managed": "$TEST_PATH",
              "provisioned": true,
              "provisioned_known": true
            },
            {
              "scope": "machine",
              "profile": "acme",
              "environment": "muse",
              "native": "$TEST_PATH",
              "managed": "$TEST_PATH",
              "provisioned": true,
              "provisioned_known": true
            },
            {
              "scope": "machine",
              "profile": "acme",
              "environment": "opencode",
              "native": "$TEST_PATH",
              "managed": "$TEST_PATH",
              "provisioned": true,
              "provisioned_known": true
            },
            {
              "scope": "machine",
              "profile": "acme",
              "environment": "pi",
              "native": "$TEST_PATH",
              "managed": "$TEST_PATH",
              "provisioned": true,
              "provisioned_known": true
            }
          ],
          "adapters": [
            {
              "id": "muse",
              "recorded": "1.4.7",
              "detected": "1.4.1",
              "supported": [
                "monolithic"
              ]
            },
            {
              "id": "claude_code",
              "recorded": "2.1.261",
              "detected": "2.1.301",
              "supported": [
                "monolithic",
                "referenced"
              ]
            },
            {
              "id": "codex_cli",
              "recorded": "0.153.2",
              "detected": "0.162.1",
              "supported": [
                "monolithic"
              ]
            },
            {
              "id": "opencode",
              "recorded": "unrecorded",
              "detected": "1.18.30",
              "supported": [
                "monolithic",
                "referenced"
              ]
            },
            {
              "id": "pi",
              "recorded": "0.84.2",
              "detected": "0.84.2",
              "supported": [
                "monolithic"
              ]
            }
          ],
          "targets": [
            {
              "id": "xcode-coding-assistant",
              "adapter": "claude_code",
              "participating": false,
              "consented": false,
              "detail": "auto: probe path absent, nothing materialized",
              "ungoverned": "the embedded host's MCP configuration and commands/ are present, unaudited, and outside this capability"
            },
            {
              "id": "xcode-coding-assistant",
              "adapter": "codex_cli",
              "participating": false,
              "consented": false,
              "detail": "auto: probe path absent, nothing materialized",
              "ungoverned": "the embedded host's MCP configuration and commands/ are present, unaudited, and outside this capability"
            }
          ],
          "profiles": [
            {
              "profile": "acme",
              "lock_hash": "sha256:e13ef5a9cd5f1afaedb4c29010a731b26caed36f798fb6396b14db3ea456ea9d",
              "members": [
                {
                  "kind": "context",
                  "name": "acme",
                  "weight": 0
                }
              ],
              "precedence": {
                "winner": "higher-weight",
                "placement": "winner-last"
              },
              "transitive_system_modules": "drop",
              "dropped_system_modules": []
            },
            {
              "profile": "default",
              "lock_hash": "sha256:06fb1543fb69055047d8f34bbac2901f06b4da1374cdbbbc47a26d22126362a4",
              "members": [
                {
                  "kind": "context",
                  "name": "default",
                  "weight": 0
                }
              ],
              "precedence": {
                "winner": "higher-weight",
                "placement": "winner-last"
              },
              "transitive_system_modules": "drop",
              "dropped_system_modules": []
            }
          ],
          "providers": [
            {
              "name": "engines",
              "executable": "curator-engines",
              "resolved": null,
              "refused_path": "$PRIVATE_PATH",
              "verdict": "refused",
              "diagnostic": "subcommand_provider_untrusted",
              "current": false,
              "trust_roots_consulted": [
                "$TEST_PATH"
              ]
            },
            {
              "name": "run",
              "executable": "curator-run",
              "resolved": "$TEST_PATH",
              "verdict": "outside_trust_roots",
              "diagnostic": "subcommand_provider_outside_trust_roots",
              "current": true,
              "trust_roots_consulted": [
                "$TEST_PATH"
              ]
            },
            {
              "name": "session",
              "executable": "curator-session",
              "resolved": "$TEST_PATH",
              "verdict": "outside_trust_roots",
              "diagnostic": "subcommand_provider_outside_trust_roots",
              "current": true,
              "trust_roots_consulted": [
                "$TEST_PATH"
              ]
            }
          ],
          "unregistered_environments": null,
          "orphans": null,
          "notes": [
            "Muse: monolithic root context delivered at config/muse/AGENTS.md (verified on 1.4.7); managed exec requires --no-foreign-personal-context when the target is materialized; serve cannot exclude foreign personal context (foreign-personal-context isolation gap); workspace AGENTS.md prevailed in observed conflicts; model-visible foreign removal, skills discovery, system-prompt discovery, refresh semantics, and state/cache internal layouts remain unverified",
            "opencode skills come from the machine-current profile, split-brain by construction (§7.1)"
          ],
          "non_current": true,
          "shell_hook_trust": null,
          "s4_profile": "s4-warn",
          "passable_env_names": null,
          "warnings": [
            "mcp_package_allowlist_empty: the MCP package allowlist is empty; every declaration package in the closure is admitted"
          ],
          "codex_seed_rule": {
            "revision": "B",
            "provenance": "shipped"
          },
          "mcp_declarations": null,
          "registry_posture": [
            {
              "name": "trusted",
              "url": "https://registry.example.test",
              "high_water_version": null,
              "high_water_log_size": null,
              "last_boundary_verified": false
            }
          ],
          "require_source_signers": false,
          "source_signers": null,
          "update_confirmation_revision": "B-flip",
          "update_confirmation_behavior": "a triggered update delta refuses profile_update_confirmation_required unless --confirm-system-delta is given",
          "security_posture_rows": [
            {
              "gate": "security_posture",
              "value": "permissive",
              "source": "profile"
            },
            {
              "gate": "hook-trust",
              "value": "A-warning",
              "source": "shipped"
            },
            {
              "gate": "registry-policy",
              "value": "",
              "source": "profile"
            },
            {
              "gate": "audit-mode",
              "value": "",
              "source": "profile"
            },
            {
              "gate": "source-allowlist",
              "value": 0,
              "source": "profile"
            },
            {
              "gate": "env-passthrough",
              "value": "s4-warn",
              "source": "shipped"
            },
            {
              "gate": "transitive-system-modules",
              "value": "",
              "source": "profile"
            },
            {
              "gate": "provider-trust-roots",
              "value": "revision-A",
              "source": "shipped"
            },
            {
              "gate": "source-signers",
              "value": "false",
              "source": "profile"
            },
            {
              "gate": "update-confirmation",
              "value": "A-warning",
              "source": "shipped"
            },
            {
              "gate": "codex-seed",
              "value": "B",
              "source": "shipped"
            },
            {
              "gate": "store-boundary",
              "value": "enforced",
              "source": "shipped"
            },
            {
              "gate": "write-discipline",
              "value": "enforced",
              "source": "shipped"
            },
            {
              "gate": "mcp-package-allowlist",
              "value": 0,
              "source": "profile"
            }
          ]
        }
        
        stderr:
        warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults
panic: test timed out after 10m0s
	running tests:
		TestEnvResolveAcceptsAliases (0s)

goroutine 4233 [running]:
testing.(*M).startAlarm.func1()
	$PRIVATE_PATH +0x2b0
created by time.goFunc
	$PRIVATE_PATH +0x38

goroutine 1 [chan receive]:
testing.(*T).Run(0x14000504fc0, {0x1055a8094?, 0x140004b3a98?}, 0x1058b63f0)
	$PRIVATE_PATH +0x378
testing.runTests.func1(0x14000504fc0)
	$PRIVATE_PATH +0x38
testing.tRunner(0x14000504fc0, 0x140004b3bc8)
	$PRIVATE_PATH +0xc8
testing.runTests(0x1400000e228, {0x105db7240, 0x196, 0x196}, {0x140004b3c48?, 0x104e58a2c?, 0x105dbcae0?})
	$PRIVATE_PATH +0x3b8
testing.(*M).Run(0x140005a6fa0)
	$PRIVATE_PATH +0x530
github.com/relux-works/curator/cmd/curator.TestMain(0x140005a6fa0)
	$PRIVATE_PATH +0x184
main.main()
	_testmain.go:857 +0x88

goroutine 51 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000249a40)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestCLIAuditBindsDependencyDirectory(0x14000249a40)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000249a40, 0x1058b60c0)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 53 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000249dc0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestReviewCLIStatusDetectsChangedTagAtSameCommit(0x14000249dc0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000249dc0, 0x1058b6ac0)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 42 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000582e00)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestClassifySkillBuildsWithoutCompiledStateStaysSilent(0x14000582e00)
	$PRIVATE_PATH +0x20
testing.tRunner(0x14000582e00, 0x1058b6170)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 52 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000249c00)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestCLIStatusReportsLegacyPackageMarkerCurrent(0x14000249c00)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000249c00, 0x1058b6110)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 62 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000603180)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestCLIAuditEmitsScriptLabels(0x14000603180)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000603180, 0x1058b60c8)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 63 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000603340)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestCLIAuditMixedSkillRecordsBothPolicies(0x14000603340)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000603340, 0x1058b60e0)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 81 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140006036c0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestCLIAuditJSONRecordsExplicitAbsence(0x140006036c0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140006036c0, 0x1058b60d8)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 64 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000603500)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestCLIAuditJSONCarriesScriptLabels(0x14000603500)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000603500, 0x1058b60d0)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 82 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000603880)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestClassifyBuildCommandMapsEveryCacheOutcomeToADistinctCode(0x14000603880)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000603880, 0x1058b6130)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 83 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000603c00)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestUnusableToolchainRowsCarryTheDriverBoundaryCode(0x14000603c00)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000603c00, 0x1058b6cd0)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 84 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000505500)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestClassifyBuildCommandDetectsRecordedIdentityDrift(0x14000505500)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000505500, 0x1058b6128)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 85 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000505a40)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestBuildInputDriftIsAttributedOnlyAsFarAsTheMarkerProves(0x14000505a40)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000505a40, 0x1058b60a8)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 86 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000505dc0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestInputCausesAreDistinctAndDocumented(0x14000505dc0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000505dc0, 0x1058b66e0)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 87 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000003dc0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestClassifySkillBuildsAcceptsOnlyAnExactlyCurrentInstallation(0x14000003dc0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000003dc0, 0x1058b6150)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 43 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000582fc0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestClassifySkillBuildsDetectsContextExposure(0x14000582fc0)
	$PRIVATE_PATH +0x20
testing.tRunner(0x14000582fc0, 0x1058b6160)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 44 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000583340)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestClassifySkillBuildsDetectsCommandDriftInBothDirections(0x14000583340)
	$PRIVATE_PATH +0x20
testing.tRunner(0x14000583340, 0x1058b6158)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 45 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140005836c0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestClassifySkillBuildsRefusesAMarkerThatCannotDescribeABuild(0x140005836c0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140005836c0, 0x1058b6168)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 46 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000583a40)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestMarkerDigestsDistinguishesAbsentAndUnreadable(0x14000583a40)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000583a40, 0x1058b6720)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 47 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000583dc0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestContextExposureDistinguishesAbsentAndUnreadableBuildRoots(0x14000583dc0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000583dc0, 0x1058b6210)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 48 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140001761c0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestMarkerRefusalSeparatesUnsupportedFromInvalid(0x140001761c0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140001761c0, 0x1058b6728)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 97 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000176540)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestCurrentnessCodesAreDistinctAndOnlyExactStatesPass(0x14000176540)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000176540, 0x1058b6228)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 98 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140001768c0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestEveryCurrentnessCodeIsDocumented(0x140001768c0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140001768c0, 0x1058b6548)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 99 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000176c40)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestCheckFailsForEveryNonCurrentCode(0x14000176c40)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000176c40, 0x1058b6120)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 100 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000176fc0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestBuildReportsNeverPublishAnAbsolutePath(0x14000176fc0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000176fc0, 0x1058b60b8)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 101 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000177340)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestBuildReportDetailIsBounded(0x14000177340)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000177340, 0x1058b60b0)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 102 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140001776c0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestGoToolchainGuidanceNamesTheAcceptedSelectionAndTestedFamilies(0x140001776c0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140001776c0, 0x1058b6640)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 103 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000177a40)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestRepairNoticesDistinguishRepairFromAPreservedInstallation(0x14000177a40)
	$PRIVATE_PATH +0x30
testing.tRunner(0x14000177a40, 0x1058b6ab0)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 104 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000177dc0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestMarkerDigestsDetectAConcurrentInstallMarkerChange(0x14000177dc0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000177dc0, 0x1058b6718)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 105 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140004001c0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestStatusReportMarksCompiledStateThatMovedDuringTheCheck(0x140004001c0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140004001c0, 0x1058b6c00)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 106 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000400540)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestStatusReportReportsCompiledCommandsOfAnUninstalledSkill(0x14000400540)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000400540, 0x1058b6c08)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 107 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140004008c0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestClassifySkillBuildsAcceptsEveryBuildBearingMarkerSchema(0x140004008c0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140004008c0, 0x1058b6148)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 108 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000400c40)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestStatusReportFindsASchema8InstallationCurrent(0x14000400c40)
	$PRIVATE_PATH +0x24
testing.tRunner(0x14000400c40, 0x1058b6bf8)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 1928 [chan receive, 4 minutes]:
testing.(*T).Parallel(0x1400038ac40)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/cmd/curator.TestProductionExternalDepsAssignsTransportProvenanceSink(0x1400038ac40)
	$PRIVATE_PATH +0x28
testing.tRunner(0x1400038ac40, 0x1058b67e0)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 4277 [runnable]:
syscall.syscall(0x140001c2920?, 0x4?, 0x140006ac528?, 0x104f00d0c?)
	$PRIVATE_PATH +0x40
syscall.Lstat({0x140001c2920?, 0x104f00864?}, 0x14000b00108)
	$PRIVATE_PATH +0x6c
os.lstatNolog.func1(...)
	$PRIVATE_PATH
os.ignoringEINTR(...)
	$PRIVATE_PATH
os.lstatNolog({0x140001c2920, 0x4})
	$PRIVATE_PATH +0x50
os.Lstat({0x140001c2920, 0x4})
	$PRIVATE_PATH +0x30
path/filepath.walkSymlinks({0x140003d9380?, 0x1a01000012?})
	$PRIVATE_PATH +0x14c
path/filepath.evalSymlinks(...)
	$PRIVATE_PATH
path/filepath.EvalSymlinks({0x140003d9380?, 0x76?})
	$PRIVATE_PATH +0x20
github.com/relux-works/curator/internal/transaction.canonicalNamespacePath({0x140003d9380, 0x76})
	$PRIVATE_PATH +0xb8
github.com/relux-works/curator/internal/transaction.(*Engine).resolveCanonical(0x105f1f570?, {0x140003d9380?, 0x105800e80?})
	$PRIVATE_PATH +0x4c
github.com/relux-works/curator/internal/transaction.(*Engine).resolveTargetPath(0xa80?, {0x140003d9380?, 0x140006ac8a8?}, 0x94?)
	$PRIVATE_PATH +0x1a0
github.com/relux-works/curator/internal/transaction.validateIndependentTargetNamespacesWithResolver({0x1400023e008, 0x2, 0x1?}, 0x76?, 0x1?, {0x0, 0x0, 0x76?})
	$PRIVATE_PATH +0x6f8
github.com/relux-works/curator/internal/transaction.validateJournalWithResolver(0x140003da210, 0x104e575f8?, 0x14000c62420?)
	$PRIVATE_PATH +0x12dc
github.com/relux-works/curator/internal/transaction.(*Engine).validateJournal(0x1400043e1e0, 0x140003da210)
	$PRIVATE_PATH +0xe8
github.com/relux-works/curator/internal/transaction.(*Engine).saveJournal(0x1400043e1e0, 0x140003da210)
	$PRIVATE_PATH +0x38
github.com/relux-works/curator/internal/transaction.(*Engine).stageTarget(0x1400043e1e0, 0x140003da210, 0x1)
	$PRIVATE_PATH +0xd4
github.com/relux-works/curator/internal/transaction.(*Engine).Prepare(0x1400043e1e0, {0x1058be240?, 0x14000516358?}, {{0x14000c62420, 0x28}, {0x10558dd58, 0x8}, {0x14000454000, 0x2, 0x2}, ...})
	$PRIVATE_PATH +0x324
github.com/relux-works/curator/internal/envprofile.(*operation).publish(0x140003385f0, 0x140000400f8, {0x0, 0x0, 0x0?})
	$PRIVATE_PATH +0x8d4
github.com/relux-works/curator/internal/envprofile.ensureDefault(_, {_, _}, {{0x0, 0x0, 0x0}, {0x0, 0x0, 0x0}, 0x140005a3590, ...})
	$PRIVATE_PATH +0x514
github.com/relux-works/curator/internal/envprofile.installLocked(_, {_, _}, {{0x140003e6600, 0x5b}, {0x0, 0x0}, {0x0, 0x0}, {0x0, ...}, ...})
	$PRIVATE_PATH +0x500
github.com/relux-works/curator/internal/envprofile.Install({_, _}, {{0x140003e6600, 0x5b}, {0x0, 0x0}, {0x0, 0x0}, {0x0, 0x0}, ...})
	$PRIVATE_PATH +0x150
github.com/relux-works/curator/cmd/curator.cli.cmdProfileInstall({{0x1058c06f0, 0x1400035e1c0}, {0x1058be060, 0x140005422a0}, {0x1058be060, 0x140005422c0}, 0x1058b7198, 0x0, 0x0, 0x0, ...}, ...)
	$PRIVATE_PATH +0x3dc
github.com/relux-works/curator/cmd/curator.cli.cmdProfile({{0x1058c06f0, 0x1400035e1c0}, {0x1058be060, 0x140005422a0}, {0x1058be060, 0x140005422c0}, 0x1058b7198, 0x0, 0x0, 0x0, ...}, ...)
	$PRIVATE_PATH +0x4a0
github.com/relux-works/curator/cmd/curator.cli.run({{0x1058c06f0, 0x1400035e1c0}, {0x1058be060, 0x140005422a0}, {0x1058be060, 0x140005422c0}, 0x1058b7198, 0x0, 0x0, 0x0, ...}, ...)
	$PRIVATE_PATH +0x6a8
github.com/relux-works/curator/cmd/curator.run(...)
	$PRIVATE_PATH
github.com/relux-works/curator/cmd/curator.runProfile(0x1400050e1c0?, {{0x140007c6150, 0x67}, 0x140001b0408, {0x0, 0x0}, {0x0, 0x0}}, {0x14000734ee0, 0x3, ...})
	$PRIVATE_PATH +0x120
github.com/relux-works/curator/cmd/curator.TestEnvResolveAcceptsAliases(0x1400050e1c0)
	$PRIVATE_PATH +0x130
testing.tRunner(0x1400050e1c0, 0x1058b63f0)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364
FAIL	github.com/relux-works/curator/cmd/curator	600.583s
?   	github.com/relux-works/curator/cmd/curator-spec-pin	[no test files]
ok  	github.com/relux-works/curator/internal/adapters	0.304s
ok  	github.com/relux-works/curator/internal/artifactpolicy	389.309s
?   	github.com/relux-works/curator/internal/artifactpolicy/conformance	[no test files]
ok  	github.com/relux-works/curator/internal/audit	6.262s
ok  	github.com/relux-works/curator/internal/buildcache	12.218s
ok  	github.com/relux-works/curator/internal/buildmeta	0.488s
ok  	github.com/relux-works/curator/internal/buildrepo	284.399s
ok  	github.com/relux-works/curator/internal/buildsource	0.802s
ok  	github.com/relux-works/curator/internal/capabilities	0.262s
ok  	github.com/relux-works/curator/internal/closure	97.367s
ok  	github.com/relux-works/curator/internal/closureexec	5.029s
ok  	github.com/relux-works/curator/internal/closuregraph	12.798s
ok  	github.com/relux-works/curator/internal/config	4.861s
ok  	github.com/relux-works/curator/internal/conformancecoverage	0.277s
ok  	github.com/relux-works/curator/internal/contextaudit	0.279s
ok  	github.com/relux-works/curator/internal/contextlock	0.283s
ok  	github.com/relux-works/curator/internal/contextmaterialize	0.266s
ok  	github.com/relux-works/curator/internal/contextpkg	0.301s
ok  	github.com/relux-works/curator/internal/contextresolve	0.274s
ok  	github.com/relux-works/curator/internal/contextstore	0.385s
ok  	github.com/relux-works/curator/internal/crossconformance	366.018s
ok  	github.com/relux-works/curator/internal/devsub	0.623s
ok  	github.com/relux-works/curator/internal/envfiles	0.506s
ok  	github.com/relux-works/curator/internal/envfragment	0.379s
ok  	github.com/relux-works/curator/internal/envmarker	0.370s
panic: test timed out after 10m0s
	running tests:
		TestRC14IdentityMigrationRefusesVersionAndPinMismatch (0s)
		TestRC14IdentityMigrationRefusesVersionAndPinMismatch/marker-version (0s)

goroutine 11274 [running]:
testing.(*M).startAlarm.func1()
	$PRIVATE_PATH +0x2b0
created by time.goFunc
	$PRIVATE_PATH +0x38

goroutine 1 [chan receive]:
testing.(*T).Run(0x14000181500, {0x10114122f?, 0x140002a5ac8?}, 0x10137c510)
	$PRIVATE_PATH +0x378
testing.runTests.func1(0x14000181500)
	$PRIVATE_PATH +0x38
testing.tRunner(0x14000181500, 0x140002a5bf8)
	$PRIVATE_PATH +0xc8
testing.runTests(0x140001b41f8, {0x101789460, 0x14d, 0x14d}, {0x1f?, 0x140002a5c88?, 0x10178e3e0?})
	$PRIVATE_PATH +0x3b8
testing.(*M).Run(0x1400046d400)
	$PRIVATE_PATH +0x530
github.com/relux-works/curator/internal/envprofile.TestMain(0x1400046d400)
	$PRIVATE_PATH +0xa4
main.main()
	_testmain.go:711 +0x88

goroutine 11781 [runnable]:
syscall.readlen(0x9?, 0x1400003fa20?, 0x8?)
	$PRIVATE_PATH +0xd4
syscall.forkExec({0x1400035e3e0?, 0xa8?}, {0x14000440630, 0x3, 0x3}, 0x2?)
	$PRIVATE_PATH +0x394
syscall.StartProcess(...)
	$PRIVATE_PATH
os.startProcess({0x1400035e3e0, 0xc}, {0x14000440630, 0x3, 0x3}, 0x1400003fca0)
	$PRIVATE_PATH +0x1e8
os.StartProcess({0x1400035e3e0, 0xc}, {0x14000440630, 0x3, 0x3}, 0x1400003fca0)
	$PRIVATE_PATH +0x50
os/exec.(*Cmd).Start(0x140004a4180)
	$PRIVATE_PATH +0x3e8
os/exec.(*Cmd).Run(0x140004a4180)
	$PRIVATE_PATH +0x20
github.com/relux-works/curator/internal/gitops.runWithEnv({0x140004ec160, 0xa8}, {0x1400053c008, 0xe, 0x82}, {0x1400003ff00, 0x2, 0x2})
	$PRIVATE_PATH +0x12c
github.com/relux-works/curator/internal/gitops.PinnedCommitTree({0x140004ec160, 0xa8}, {0x1400002bf80, 0x28})
	$PRIVATE_PATH +0x120
github.com/relux-works/curator/internal/envprofile.pinnedCommitTree({0x1400055dd00?, 0xc0?}, {{0x14000351a8b, 0x5}, {0x14000351a90, 0x7}, {0x140004602a0, 0x1b}, {0x0, 0x0}, ...})
	$PRIVATE_PATH +0xcc
github.com/relux-works/curator/internal/envprofile.expectedStoreHash({0x1400055dd00?, 0x1400051e000?}, {{0x14000351a8b, 0x5}, {0x14000351a90, 0x7}, {0x140004602a0, 0x1b}, {0x0, 0x0}, ...})
	$PRIVATE_PATH +0xf8
github.com/relux-works/curator/internal/envprofile.storeEntryPinHashes({0x1400055dd00?, 0x140000403a8?}, {{0x14000351a8b, 0x5}, {0x14000351a90, 0x7}, {0x140004602a0, 0x1b}, {0x0, 0x0}, ...}, ...)
	$PRIVATE_PATH +0x58
github.com/relux-works/curator/internal/envprofile.validateNamedStorePins({0x1400055dd00, 0x7e}, 0x140000200c0)
	$PRIVATE_PATH +0x164
github.com/relux-works/curator/internal/envprofile.Resolve({{0x1400055dd00, 0x7e}, {0x1011078fc, 0x6}, {0x10110679a, 0x2}, {0x140001df080, 0x7e}, {0x1400056e3c0, 0x1400056e660, ...}, ...})
	$PRIVATE_PATH +0x714
github.com/relux-works/curator/internal/envprofile.provision(0x1400037ba40, 0x14000636900, {0x10110679a, 0x2}, {0x1400056e3c0, 0x1400056e660, 0x0, 0x0, {0x0, 0x0}, ...})
	$PRIVATE_PATH +0xa8
github.com/relux-works/curator/internal/envprofile.legacyIdentityFixture(0x1400037ba40, {0x14000041f08, 0x1, 0x172?})
	$PRIVATE_PATH +0x114
github.com/relux-works/curator/internal/envprofile.TestRC14IdentityMigrationRefusesVersionAndPinMismatch.func1(0x1400037ba40)
	$PRIVATE_PATH +0x54
testing.tRunner(0x1400037ba40, 0x14000216648)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 11780
	$PRIVATE_PATH +0x364

goroutine 11780 [chan receive]:
testing.(*T).Run(0x1400037b880, {0x10110d1fd?, 0x1400039cf38?}, 0x14000216648)
	$PRIVATE_PATH +0x378
github.com/relux-works/curator/internal/envprofile.TestRC14IdentityMigrationRefusesVersionAndPinMismatch(0x1400037b880)
	$PRIVATE_PATH +0x78
testing.tRunner(0x1400037b880, 0x10137c510)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364
FAIL	github.com/relux-works/curator/internal/envprofile	600.360s
ok  	github.com/relux-works/curator/internal/envregistry	0.265s
ok  	github.com/relux-works/curator/internal/gitcred	5.981s
ok  	github.com/relux-works/curator/internal/gitignore	2.678s
ok  	github.com/relux-works/curator/internal/gitops	39.242s
ok  	github.com/relux-works/curator/internal/globalbins	2.050s
ok  	github.com/relux-works/curator/internal/godriver	107.377s
ok  	github.com/relux-works/curator/internal/hashing	0.343s
ok  	github.com/relux-works/curator/internal/hookapproval	1.074s
ok  	github.com/relux-works/curator/internal/identifiers	0.262s
ok  	github.com/relux-works/curator/internal/identity	0.320s
panic: test timed out after 10m0s
	running tests:
		TestGlobalReconcilesOwnedLegacyForwardingLink (3s)
		TestGlobalReconcilesOwnedLegacyForwardingLink/foreign-regular (1s)

goroutine 4067 [running]:
testing.(*M).startAlarm.func1()
	$PRIVATE_PATH +0x2b0
created by time.goFunc
	$PRIVATE_PATH +0x38

goroutine 1 [chan receive]:
testing.(*T).Run(0x14000183500, {0x1055e1626?, 0x35?}, 0x1058585c0)
	$PRIVATE_PATH +0x378
testing.runTests.func1(0x14000183500)
	$PRIVATE_PATH +0x38
testing.tRunner(0x14000183500, 0x14000257c28)
	$PRIVATE_PATH +0xc8
testing.runTests(0x140001aa1c8, {0x105c7e0a0, 0x16b, 0x16b}, {0xf?, 0x14000257cb8?, 0x105c83340?})
	$PRIVATE_PATH +0x3b8
testing.(*M).Run(0x1400046c820)
	$PRIVATE_PATH +0x530
github.com/relux-works/curator/internal/install.TestMain(0x1400046c820)
	$PRIVATE_PATH +0x2c
main.main()
	_testmain.go:771 +0x88

goroutine 232 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000583a40)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestScriptAuditLabelsAtInstallEntry(0x14000583a40)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000583a40, 0x105858978)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 4057 [runnable]:
syscall.syscall(0x140003a4270?, 0x92?, 0x14000a5a438?, 0x14001000000?)
	$PRIVATE_PATH +0x40
golang.org/x/sys/unix.Statfs({0x140003a4270?, 0xcd?}, 0x14000a5a490)
	$PRIVATE_PATH +0x6c
github.com/relux-works/curator/internal/transaction.namespaceNormalizationInsensitive({0x140003a4270?, 0x8?})
	$PRIVATE_PATH +0x4c
github.com/relux-works/curator/internal/transaction.validateIndependentTargetNamespacesWithResolver({0x14000400008, 0x5, 0x1?}, 0x17?, 0x4?, {0x0, 0x0, 0x9c?})
	$PRIVATE_PATH +0x71c
github.com/relux-works/curator/internal/transaction.validateJournalWithResolver(0x140001d4210, 0x104ff8920?, 0x140006dd7d0?)
	$PRIVATE_PATH +0x12dc
github.com/relux-works/curator/internal/transaction.(*Engine).validateJournal(0x140001fe320, 0x140001d4210)
	$PRIVATE_PATH +0xe8
github.com/relux-works/curator/internal/transaction.(*Engine).saveJournal(0x140001fe320, 0x140001d4210)
	$PRIVATE_PATH +0x38
github.com/relux-works/curator/internal/transaction.(*Engine).stageTarget(0x140001fe320, 0x140001d4210, 0x2)
	$PRIVATE_PATH +0x1fc
github.com/relux-works/curator/internal/transaction.(*Engine).Prepare(0x140001fe320, {0x10585f4c0?, 0x14000138158?}, {{0x140006dd7d0, 0x28}, {0x14000340b40, 0x82}, {0x140001ecd88, 0x5, 0x8}, ...})
	$PRIVATE_PATH +0x324
github.com/relux-works/curator/internal/install.runCommit({_, _}, {{0x1055acff5, 0x6}, {0x1400090c480, 0x7b}, {0x0, 0x0}, {0x14000046280, {0x1058617b8, ...}, ...}, ...})
	$PRIVATE_PATH +0xc98
github.com/relux-works/curator/internal/install.globalAttempt(_, {_, _}, {{0x0, 0x0}, 0x0, 0x0, 0x0, 0x0, 0x0, ...}, ...)
	$PRIVATE_PATH +0x3b74
github.com/relux-works/curator/internal/install.Global.func2()
	$PRIVATE_PATH +0xa0
github.com/relux-works/curator/internal/install.runWithRestarts({0x1055acff5, _}, {{_, _}, {_, _}, {_, _}, _, {_, ...}, ...}, ...)
	$PRIVATE_PATH +0xe0
github.com/relux-works/curator/internal/install.Global(_, {_, _}, {{0x0, 0x0}, 0x0, 0x0, 0x0, 0x0, 0x0, ...})
	$PRIVATE_PATH +0x548
github.com/relux-works/curator/internal/install.TestGlobalReconcilesOwnedLegacyForwardingLink.func1(0x140005b6380)
	$PRIVATE_PATH +0x5b4
testing.tRunner(0x140005b6380, 0x140001aa078)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 4008
	$PRIVATE_PATH +0x364

goroutine 311 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140005ce000)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestRunCommitRecheckRefusesSwappedParent(0x140005ce000)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140005ce000, 0x105858910)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 312 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140005ce1c0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestRunCommitRecheckPassesUnchangedBoundaries(0x140005ce1c0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140005ce1c0, 0x105858900)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 310 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140000f9dc0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestRunCommitRecheckRefusesDestinationInsideAdmitted(0x140000f9dc0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140000f9dc0, 0x105858908)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 233 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000583c00)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestScriptAuditLabelsAtGlobalInstallEntry(0x14000583c00)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000583c00, 0x105858970)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 234 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000583dc0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestDeclaredOnlyWithHostsGlobalInstallsWithOnlyDeclaredOnly(0x14000583dc0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000583dc0, 0x105858178)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 3247 [chan receive, 1 minutes]:
testing.(*T).Parallel(0x14000500540)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestDraftTransportEnabled(0x14000500540)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000500540, 0x105858450)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 235 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000202000)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestDeclaredOnlyWithHostsInstallsWithOnlyDeclaredOnly(0x14000202000)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000202000, 0x105858180)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 309 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140000f9c00)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestCacheWrongTargetRebuildCompletesWithinDeadline(0x140000f9c00)
	$PRIVATE_PATH +0x20
testing.tRunner(0x140000f9c00, 0x105858110)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 308 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140000f9a40)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestAuthoritativeCacheRejectionsAreRebuiltNeverAdopted(0x140000f9a40)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140000f9a40, 0x105858058)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 307 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140000f9880)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestAuthoritativeCacheOutcomesDriveInstallation(0x140000f9880)
	$PRIVATE_PATH +0x24
testing.tRunner(0x140000f9880, 0x105858050)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 313 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140005ce380)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestRunCommitWithoutBoundariesSkipsRecheck(0x140005ce380)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140005ce380, 0x105858928)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 314 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140005ce700)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestCommitOrdersTargetClassesWithConsumerLast(0x140005ce700)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140005ce700, 0x105858128)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 315 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140005cea80)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestConsumerLedgerIsAbsentAfterAFailedFirstInstallAndCommitsLastOnSuccess(0x140005cea80)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140005cea80, 0x105858150)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 316 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140005cee00)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestCachePublicationFailureLeavesInstallationAndProtectedCacheUntouched(0x140005cee00)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140005cee00, 0x105858108)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 317 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140005cf180)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestAFailedCommitRestoresTheBuildCacheItReplaced(0x140005cf180)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140005cf180, 0x105857f68)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 318 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140005cf500)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestARolledBackTargetCommitRestoresTheBuildCacheItReplaced(0x140005cf500)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140005cf500, 0x105857fa0)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 319 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140005cf880)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestAnInFlightTransactionKeepsThePublishedCacheEntry(0x140005cf880)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140005cf880, 0x105858030)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 320 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140005cfc00)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestAPublicationThatChangedTheCacheIsReportedAsRetained(0x140005cfc00)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140005cfc00, 0x105857f88)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 321 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140005de000)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestAReversalThatDidNotCompleteIsReportedAsRetained(0x140005de000)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140005de000, 0x105857f98)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 322 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140005de380)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestBuildPublicationFailuresAreRedactedInTheResult(0x140005de380)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140005de380, 0x1058580d8)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 323 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140005de700)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestConcurrentProjectInstallsPreserveBothConsumers(0x140005de700)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140005de700, 0x105858130)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 325 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140005dee00)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestRollbackCannotRestoreOverAnotherProjectsCommittedSharedTargets(0x140005dee00)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140005dee00, 0x1058588f8)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 326 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140005defc0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestRecoveryCompletesBeforeAnyNewMutation(0x140005defc0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140005defc0, 0x105858820)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 327 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140005df340)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestStaleInstalledGenerationRestartsInsteadOfApplyingTheOldPlan(0x140005df340)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140005df340, 0x105858a00)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 328 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140005df6c0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestMaintenanceFailureAfterCommitIsAWarning(0x140005df6c0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140005df6c0, 0x1058586d8)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 295 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000496fc0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestNoSharedTargetChangesBeforeTheManagerHomeLock(0x14000496fc0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000496fc0, 0x105858708)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 296 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000497340)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestGlobalCommitCarriesNoConsumerLedger(0x14000497340)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000497340, 0x105858538)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 297 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140004976c0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestStagedPlanRejectsTwoProducersClaimingOneLivePath(0x140004976c0)
	$PRIVATE_PATH +0x24
testing.tRunner(0x140004976c0, 0x1058589f0)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 298 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000497a40)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestAdapterLedgerCommitsAfterTheMirrorsItClaims(0x14000497a40)
	$PRIVATE_PATH +0x30
testing.tRunner(0x14000497a40, 0x105858000)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 299 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x14000497dc0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestRedactDiagnosticReplacesEveryAbsoluteLocation(0x14000497dc0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000497dc0, 0x105858840)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 300 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140004b61c0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestRedactDiagnosticCannotDriveATerminal(0x140004b61c0)
	$PRIVATE_PATH +0x24
testing.tRunner(0x140004b61c0, 0x105858828)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 301 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140004b6540)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestRedactDiagnosticIsBounded(0x140004b6540)
	$PRIVATE_PATH +0x24
testing.tRunner(0x140004b6540, 0x105858830)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 302 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140004b68c0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestRedactDiagnosticKeepsInvalidUTF8Out(0x140004b68c0)
	$PRIVATE_PATH +0x20
testing.tRunner(0x140004b68c0, 0x105858838)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 303 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140004b6c40)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestPlanLinesRedactAnUntrustedReason(0x140004b6c40)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140004b6c40, 0x105858730)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 304 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140004b6fc0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestBuildFailuresAreRedactedInTheResult(0x140004b6fc0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140004b6fc0, 0x105858090)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 337 [chan receive, 9 minutes]:
testing.(*T).Parallel(0x140004b7340)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestGoToolchainRemedyReachesTheOperatorIntact(0x140004b7340)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140004b7340, 0x1058585f8)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 3424 [chan receive, 1 minutes]:
testing.(*T).Parallel(0x140006276c0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestDraftPlanNeedsSSH(0x140006276c0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140006276c0, 0x1058583d8)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 3892 [chan receive]:
testing.(*T).Parallel(0x14000500fc0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestDraftTransportPlanV2Conversion(0x14000500fc0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000500fc0, 0x105858460)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 3248 [chan receive, 1 minutes]:
testing.(*T).Parallel(0x14000500700)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestDraftTransportPlanConversion(0x14000500700)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000500700, 0x105858458)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 3965 [chan receive]:
testing.(*T).Parallel(0x14000501a40)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestGitFixtureRetryPredicate(0x14000501a40)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000501a40, 0x105858520)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 3961 [chan receive]:
testing.(*T).Parallel(0x14000501340)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestGitFixtureRetriesOnceOnSpawnEACCES(0x14000501340)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000501340, 0x105858510)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 3960 [chan receive]:
testing.(*T).Parallel(0x14000500e00)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestObservationsRecheckDocumentsAgainstTheParsedBytes(0x14000500e00)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000500e00, 0x105858710)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 3958 [chan receive]:
testing.(*T).Parallel(0x140005b7dc0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestDeclarationGenerationIsContentAddressed(0x140005b7dc0)
	$PRIVATE_PATH +0x24
testing.tRunner(0x140005b7dc0, 0x105858160)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 3966 [chan receive]:
testing.(*T).Parallel(0x14000501c00)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestGitFixturePersistentSpawnFailureCarriesDiagnostic(0x14000501c00)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000501c00, 0x105858500)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 3964 [chan receive]:
testing.(*T).Parallel(0x14000501880)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestGitFixtureNeverRetriesWrappedExitError(0x14000501880)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000501880, 0x1058584e0)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 3441 [chan receive, 1 minutes]:
testing.(*T).Parallel(0x140005b61c0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestDraftPlanNamesProvider(0x140005b61c0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140005b61c0, 0x1058583d0)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 3963 [chan receive]:
testing.(*T).Parallel(0x140005016c0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestGitFixtureNeverRetriesNonZeroExit(0x140005016c0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140005016c0, 0x1058584d8)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 3893 [chan receive]:
testing.(*T).Parallel(0x14000501180)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestDraftPlanNeedsSSHRevision2(0x14000501180)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000501180, 0x1058583e0)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 3959 [chan receive]:
testing.(*T).Parallel(0x14000500c40)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestDocumentGenerationIsStableWhenUnreadable(0x14000500c40)
	$PRIVATE_PATH +0x24
testing.tRunner(0x14000500c40, 0x1058581d8)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 3444 [chan receive, 1 minutes]:
testing.(*T).Parallel(0x140005b6700)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestDraftManagerExecutable(0x140005b6700)
	$PRIVATE_PATH +0x20
testing.tRunner(0x140005b6700, 0x1058583b8)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 3962 [chan receive]:
testing.(*T).Parallel(0x14000501500)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestGitFixtureRetriesOnceOnSpawnEAGAIN(0x14000501500)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000501500, 0x105858518)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 3967 [chan receive]:
testing.(*T).Parallel(0x14000501dc0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestGitFixtureNonRetryableSpawnErrorFailsFastWithDiagnostic(0x14000501dc0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000501dc0, 0x1058584e8)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 3968 [chan receive]:
testing.(*T).Parallel(0x14000312000)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestGitFixtureRealSpawnEACCESRetriesOnce(0x14000312000)
	$PRIVATE_PATH +0x28
testing.tRunner(0x14000312000, 0x105858508)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 3970 [chan receive]:
testing.(*T).Parallel(0x14000312700)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestGitFixtureOuterEntryRunsRealGit(0x14000312700)
	$PRIVATE_PATH +0x24
testing.tRunner(0x14000312700, 0x1058584f8)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 3971 [chan receive]:
testing.(*T).Parallel(0x140003128c0)
	$PRIVATE_PATH +0x190
github.com/relux-works/curator/internal/install.TestGitFixtureOuterEntryLogsRetryOnTransientEACCES(0x140003128c0)
	$PRIVATE_PATH +0x28
testing.tRunner(0x140003128c0, 0x1058584f0)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364

goroutine 4008 [chan receive]:
testing.(*T).Run(0x14000312380, {0x1055b43ce?, 0x37?}, 0x140001aa078)
	$PRIVATE_PATH +0x378
github.com/relux-works/curator/internal/install.TestGlobalReconcilesOwnedLegacyForwardingLink(0x14000312380)
	$PRIVATE_PATH +0x90
testing.tRunner(0x14000312380, 0x1058585c0)
	$PRIVATE_PATH +0xc8
created by testing.(*T).Run in goroutine 1
	$PRIVATE_PATH +0x364
FAIL	github.com/relux-works/curator/internal/install	600.384s
ok  	github.com/relux-works/curator/internal/install/atomicity	338.881s
ok  	github.com/relux-works/curator/internal/interop	0.301s
ok  	github.com/relux-works/curator/internal/interop/environments	0.295s
ok  	github.com/relux-works/curator/internal/locale	0.372s
ok  	github.com/relux-works/curator/internal/managerlock	2.137s
ok  	github.com/relux-works/curator/internal/manifest	3.432s
ok  	github.com/relux-works/curator/internal/marker	17.380s
ok  	github.com/relux-works/curator/internal/mcp	0.434s
ok  	github.com/relux-works/curator/internal/moduleroots	1.324s
ok  	github.com/relux-works/curator/internal/nodesource	1.436s
ok  	github.com/relux-works/curator/internal/npmsource	113.524s
ok  	github.com/relux-works/curator/internal/opaquescan	0.345s
ok  	github.com/relux-works/curator/internal/pathboundary	0.515s
ok  	github.com/relux-works/curator/internal/pkgversion	0.258s
ok  	github.com/relux-works/curator/internal/pnpmsource	104.496s
ok  	github.com/relux-works/curator/internal/privatedir	0.728s
ok  	github.com/relux-works/curator/internal/protocoljson	0.288s
ok  	github.com/relux-works/curator/internal/registry	17.893s
ok  	github.com/relux-works/curator/internal/runtimestore	5.886s
ok  	github.com/relux-works/curator/internal/rustsource	190.528s
ok  	github.com/relux-works/curator/internal/scopes	5.977s
ok  	github.com/relux-works/curator/internal/scriptpolicy	0.271s
ok  	github.com/relux-works/curator/internal/scriptworker	24.127s
ok  	github.com/relux-works/curator/internal/shell	118.893s
ok  	github.com/relux-works/curator/internal/skillcheck	1.271s
ok  	github.com/relux-works/curator/internal/skillspec	12.706s
ok  	github.com/relux-works/curator/internal/snapshot	6.611s
ok  	github.com/relux-works/curator/internal/sourcelock	0.522s
ok  	github.com/relux-works/curator/internal/staging	0.862s
ok  	github.com/relux-works/curator/internal/stateread	0.331s
ok  	github.com/relux-works/curator/internal/swiftpmbuild	52.895s
ok  	github.com/relux-works/curator/internal/swiftpminterop	155.792s
ok  	github.com/relux-works/curator/internal/swiftpmsource	45.288s
?   	github.com/relux-works/curator/internal/testcli	[no test files]
?   	github.com/relux-works/curator/internal/testgitenv	[no test files]
?   	github.com/relux-works/curator/internal/testtoolchain	[no test files]
ok  	github.com/relux-works/curator/internal/transaction	110.826s
ok  	github.com/relux-works/curator/internal/ui	0.336s
ok  	github.com/relux-works/curator/internal/verr	0.257s
ok  	github.com/relux-works/curator/internal/version	0.260s
ok  	github.com/relux-works/curator/internal/whitelist	0.779s
ok  	github.com/relux-works/curator/internal/yarnclassicsource	11.777s
ok  	github.com/relux-works/curator/internal/yarnmodernsource	14.765s
ok  	github.com/relux-works/curator/tools/goreleaserconfig	0.263s
FAIL

````
