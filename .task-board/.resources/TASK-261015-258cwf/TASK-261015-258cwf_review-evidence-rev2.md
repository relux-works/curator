# TASK-261015-258cwf reviewer execution evidence

Candidate: a5c49ad3e15a43364c62dac70a642ce3fce77fc5.

## entry

```text
=== RUN   TestEnvConfigShowPermissionsAtProductionEntry
=== RUN   TestEnvConfigShowPermissionsAtProductionEntry/profile
=== RUN   TestEnvConfigShowPermissionsAtProductionEntry/global
--- PASS: TestEnvConfigShowPermissionsAtProductionEntry (1.29s)
    --- PASS: TestEnvConfigShowPermissionsAtProductionEntry/profile (0.65s)
    --- PASS: TestEnvConfigShowPermissionsAtProductionEntry/global (0.64s)
=== RUN   TestEnvConfigShowPermissionsRejectsUnreadableConfig
=== RUN   TestEnvConfigShowPermissionsRejectsUnreadableConfig/machine
=== RUN   TestEnvConfigShowPermissionsRejectsUnreadableConfig/system
--- PASS: TestEnvConfigShowPermissionsRejectsUnreadableConfig (0.47s)
    --- PASS: TestEnvConfigShowPermissionsRejectsUnreadableConfig/machine (0.00s)
    --- PASS: TestEnvConfigShowPermissionsRejectsUnreadableConfig/system (0.00s)
=== RUN   TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent
--- PASS: TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent (0.00s)
=== RUN   TestEnvConfigShowPermissionsRejectsUnreadableInventoryAncestor
--- PASS: TestEnvConfigShowPermissionsRejectsUnreadableInventoryAncestor (0.00s)
=== RUN   TestEnvConfigShowPermissionsWithoutProfiles
--- PASS: TestEnvConfigShowPermissionsWithoutProfiles (0.00s)
PASS
ok  	github.com/relux-works/curator/cmd/curator	2.354s

```

## inventory-metadata

```text
--- FAIL: TestEnvConfigShowPermissionsRejectsUnreadableInventoryAncestor (0.00s)
    envconfig_permissions_test.go:151: unreadable ancestor = 0, stdout="{\n  \"backup_retention\": 0,\n  \"current_profile\": null,\n  \"forms\": {},\n  \"in_place_mode\": {},\n  \"isolation\": {},\n  \"mcp_package_allowlist\": [],\n  \"overlay_default_weight\": 0,\n  \"overlays\": {},\n  \"overlays_allowed\": false,\n  \"passable_env_names\": [],\n  \"permissions\": {},\n  \"permissions_posture\": {},\n  \"precedence\": {\n    \"placement\": \"\",\n    \"winner\": \"\"\n  },\n  \"provider_directories\": [],\n  \"require_current_profile\": null,\n  \"require_source_signers\": false,\n  \"scoped_current\": {},\n  \"secret_material_waivers\": [],\n  \"shadow_acknowledged\": [],\n  \"source_signers\": {},\n  \"system_module_waivers\": [],\n  \"system_prompt_files\": {},\n  \"targets\": {},\n  \"transitive_system_modules\": \"\",\n  \"xdg_seed_allowlist\": []\n}\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults\n"
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	0.552s
FAIL

exit=1

```

## inventory-type

```text
--- FAIL: TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent (0.00s)
    envconfig_permissions_test.go:134: unreadable posture = 0, stdout="{\n  \"backup_retention\": 0,\n  \"current_profile\": null,\n  \"forms\": {},\n  \"in_place_mode\": {},\n  \"isolation\": {},\n  \"mcp_package_allowlist\": [],\n  \"overlay_default_weight\": 0,\n  \"overlays\": {},\n  \"overlays_allowed\": false,\n  \"passable_env_names\": [],\n  \"permissions\": {},\n  \"permissions_posture\": {},\n  \"precedence\": {\n    \"placement\": \"\",\n    \"winner\": \"\"\n  },\n  \"provider_directories\": [],\n  \"require_current_profile\": null,\n  \"require_source_signers\": false,\n  \"scoped_current\": {},\n  \"secret_material_waivers\": [],\n  \"shadow_acknowledged\": [],\n  \"source_signers\": {},\n  \"system_module_waivers\": [],\n  \"system_prompt_files\": {},\n  \"targets\": {},\n  \"transitive_system_modules\": \"\",\n  \"xdg_seed_allowlist\": []\n}\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults\n"
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	0.364s
FAIL

exit=1

```

## system-read

```text
--- FAIL: TestEnvConfigShowPermissionsRejectsUnreadableConfig (0.47s)
    --- FAIL: TestEnvConfigShowPermissionsRejectsUnreadableConfig/system (0.00s)
        envconfig_permissions_test.go:118: unreadable config = 0, stdout="{\n  \"backup_retention\": 5,\n  \"current_profile\": null,\n  \"forms\": {},\n  \"in_place_mode\": {},\n  \"isolation\": {},\n  \"mcp_package_allowlist\": [],\n  \"overlay_default_weight\": 1000,\n  \"overlays\": {},\n  \"overlays_allowed\": true,\n  \"passable_env_names\": [],\n  \"permissions\": {},\n  \"permissions_posture\": {\n    \"acme\": {\n      \"mode\": \"native\",\n      \"locked\": false,\n      \"source\": \"default\"\n    },\n    \"default\": {\n      \"mode\": \"native\",\n      \"locked\": false,\n      \"source\": \"default\"\n    }\n  },\n  \"precedence\": {\n    \"placement\": \"winner-last\",\n    \"winner\": \"higher-weight\"\n  },\n  \"provider_directories\": [],\n  \"require_current_profile\": null,\n  \"require_source_signers\": false,\n  \"scoped_current\": {},\n  \"secret_material_waivers\": [],\n  \"shadow_acknowledged\": [],\n  \"source_signers\": {},\n  \"system_module_waivers\": [],\n  \"system_prompt_files\": {},\n  \"targets\": {},\n  \"transitive_system_modules\": \"drop\",\n  \"xdg_seed_allowlist\": [\n    \"git\",\n    \"gh\",\n    \"ssh\"\n  ]\n}\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults\n"
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	0.859s
FAIL

exit=1

```

## current-read

```text
--- FAIL: TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent (0.00s)
    envconfig_permissions_test.go:134: unreadable posture = 0, stdout="{\n  \"mode\": \"native\",\n  \"locked\": false,\n  \"source\": \"default\"\n}\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults\n"
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	0.374s
FAIL

exit=1

```

## global-lock

```text
--- FAIL: TestEnvConfigShowPermissionsAtProductionEntry (7.71s)
    --- FAIL: TestEnvConfigShowPermissionsAtProductionEntry/global (4.05s)
        envconfig_permissions_test.go:65: acme posture = map[locked:false mode:native source:default], want map[locked:true mode:native source:global]
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	8.092s
FAIL

exit=1

```

## freehunt

```text
=== RUN   TestReviewLockReadFaults
=== RUN   TestReviewLockReadFaults/inventory
=== RUN   TestReviewLockReadFaults/current
=== RUN   TestReviewLockReadFaults/malformed
--- PASS: TestReviewLockReadFaults (0.02s)
    --- PASS: TestReviewLockReadFaults/inventory (0.01s)
    --- PASS: TestReviewLockReadFaults/current (0.00s)
    --- PASS: TestReviewLockReadFaults/malformed (0.01s)
PASS
ok  	github.com/relux-works/curator/cmd/curator	0.599s

```

## fragment

```text
ok  	github.com/relux-works/curator/cmd/curator	3.397s

```

## Commands

Entry: `go test ./cmd/curator -run "^TestEnvConfigShowPermissions" -count=1 -v -timeout=2m` (0). Fragment: `go test ./cmd/curator -run "^TestEnvResolveEmitsPermissionsV2AtProductionEntry$" -count=1 -timeout=2m` (0). Mutants use the successor overlays and exact commands in TASK-261015-258cwf_successor-handoff.md, rerun unchanged here; all five exited 1 with named assertion failures. Free hunt: `go test -overlay .temp/review-lock-r2/attack.json ./cmd/curator -run "^TestReviewLockReadFaults$" -count=1 -v -timeout=2m` (0).

## Additional test (overlay append)

```go
// Proves cyclic inventory/current links and malformed config cannot report unlocked posture.
func TestReviewLockReadFaults(t *testing.T) {
 for _, target := range []string{"inventory", "current"} {
  t.Run(target, func(t *testing.T) {
   source, home := profileHome(t)
   path := filepath.Join(home, "profiles")
   if target == "current" {
    if err := os.MkdirAll(path, 0700); err != nil { t.Fatal(err) }
    path = filepath.Join(path, "current")
   }
   if err := os.Symlink(path, path); err != nil { t.Fatal(err) }
   args := []string{"env", "config", "show"}
   if target == "current" { args = append(args, "permissions") }
   code, stdout, stderr := runProfile(t, source, args...)
   if code != exitFail || stdout != "" || !strings.Contains(stderr, "unreadable") { t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout, stderr) }
  })
 }
 t.Run("malformed", func(t *testing.T) {
  source, _ := profileHome(t)
  t.Setenv("CURATOR_SYSTEM_CONFIG", "")
  if err := os.WriteFile(source.path, []byte("{"), 0600); err != nil { t.Fatal(err) }
  for _, args := range [][]string{{"env", "config", "show"}, {"env", "config", "show", "permissions"}} {
   code, stdout, stderr := runFileConfigProfile(t, source.path, args...)
   if code != exitFail || stdout != "" || stderr == "" { t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout, stderr) }
  }
 })
}

```
