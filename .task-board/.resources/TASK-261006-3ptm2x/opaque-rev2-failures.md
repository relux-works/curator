# rev2 hosted gate 37459410231 — failures (orchestrator extract)

## ubuntu-latest: 5 failing leaves
- cmd/curator TestStatusCheckRefusesV1NULAppearingAfterInstall
        nul_opaque_v1_rework_test.go:99: status --check stderr lacks the opaque finding:
        nul_opaque_v1_rework_test.go:99: status --check stderr lacks the opaque finding:
- internal/install TestDraftLaneStillBlocksNULBearingMember/audit-disabled
        nul_opaque_v2_test.go:198: source_member_invalid: frozen package context: audit.opaque.nul-byte: regular file contains a NUL byte and is treated as opaque (files: assets/a.bin)
        nul_opaque_v2_test.go:198: source_member_invalid: frozen package context: audit.opaque.nul-byte: regular file contains a NUL byte and is treated as opaque (files: assets/a.bin)
- internal/install TestDraftLaneStillBlocksNULBearingMember/audit-enabled
        nul_opaque_v2_test.go:189: source_member_invalid: frozen package context: audit.opaque.nul-byte: regular file contains a NUL byte and is treated as opaque (files: assets/a.bin)
        nul_opaque_v2_test.go:189: source_member_invalid: frozen package context: audit.opaque.nul-byte: regular file contains a NUL byte and is treated as opaque (files: assets/a.bin)
- internal/install TestV1ReinstallRefusesNULAppearingAfterInstall/collision-preserving-edit
        nul_opaque_v1_rework_test.go:67: status plan after v1-collision edit: status "ok": test: skill-a: warning: script-command-declared-only (commands.skill-a-tool.execution_policy): command 'skill-a-tool' is declared-only (no execution_policy); its capabilities bound nothing at run time; adopt execu
        nul_opaque_v1_rework_test.go:67: status plan after v1-collision edit: status "ok": test: skill-a: warning: script-command-declared-only (commands.skill-a-tool.execution_policy): command 'skill-a-tool' is declared-only (no execution_policy); its capabilities bound nothing at run time; adopt execu
- internal/install TestV1ReinstallRefusesNULAppearingAfterInstall/fresh-nul-file
        nul_opaque_v1_rework_test.go:74: status plan after NUL appears: status "ok": test: skill-a: warning: script-command-declared-only (commands.skill-a-tool.execution_policy): command 'skill-a-tool' is declared-only (no execution_policy); its capabilities bound nothing at run time; adopt execution_p
        nul_opaque_v1_rework_test.go:74: status plan after NUL appears: status "ok": test: skill-a: warning: script-command-declared-only (commands.skill-a-tool.execution_policy): command 'skill-a-tool' is declared-only (no execution_policy); its capabilities bound nothing at run time; adopt execution_p

## macos-latest: 5 failing leaves
- cmd/curator TestStatusCheckRefusesV1NULAppearingAfterInstall
        nul_opaque_v1_rework_test.go:99: status --check stderr lacks the opaque finding:
        nul_opaque_v1_rework_test.go:99: status --check stderr lacks the opaque finding:
- internal/install TestDraftLaneStillBlocksNULBearingMember/audit-disabled
        nul_opaque_v2_test.go:198: source_member_invalid: frozen package context: audit.opaque.nul-byte: regular file contains a NUL byte and is treated as opaque (files: assets/a.bin)
        nul_opaque_v2_test.go:198: source_member_invalid: frozen package context: audit.opaque.nul-byte: regular file contains a NUL byte and is treated as opaque (files: assets/a.bin)
- internal/install TestDraftLaneStillBlocksNULBearingMember/audit-enabled
        nul_opaque_v2_test.go:189: source_member_invalid: frozen package context: audit.opaque.nul-byte: regular file contains a NUL byte and is treated as opaque (files: assets/a.bin)
        nul_opaque_v2_test.go:189: source_member_invalid: frozen package context: audit.opaque.nul-byte: regular file contains a NUL byte and is treated as opaque (files: assets/a.bin)
- internal/install TestV1ReinstallRefusesNULAppearingAfterInstall/collision-preserving-edit
        nul_opaque_v1_rework_test.go:67: status plan after v1-collision edit: status "ok": test: skill-a: warning: script-command-declared-only (commands.skill-a-tool.execution_policy): command 'skill-a-tool' is declared-only (no execution_policy); its capabilities bound nothing at run time; adopt execu
        nul_opaque_v1_rework_test.go:67: status plan after v1-collision edit: status "ok": test: skill-a: warning: script-command-declared-only (commands.skill-a-tool.execution_policy): command 'skill-a-tool' is declared-only (no execution_policy); its capabilities bound nothing at run time; adopt execu
- internal/install TestV1ReinstallRefusesNULAppearingAfterInstall/fresh-nul-file
        nul_opaque_v1_rework_test.go:74: status plan after NUL appears: status "ok": test: skill-a: warning: script-command-declared-only (commands.skill-a-tool.execution_policy): command 'skill-a-tool' is declared-only (no execution_policy); its capabilities bound nothing at run time; adopt execution_p
        nul_opaque_v1_rework_test.go:74: status plan after NUL appears: status "ok": test: skill-a: warning: script-command-declared-only (commands.skill-a-tool.execution_policy): command 'skill-a-tool' is declared-only (no execution_policy); its capabilities bound nothing at run time; adopt execution_p

## windows-latest: 5 failing leaves
- cmd/curator TestStatusCheckRefusesV1NULAppearingAfterInstall
        nul_opaque_v1_rework_test.go:99: status --check stderr lacks the opaque finding:
        nul_opaque_v1_rework_test.go:99: status --check stderr lacks the opaque finding:
- internal/install TestDraftLaneStillBlocksNULBearingMember/audit-disabled
        nul_opaque_v2_test.go:198: source_member_invalid: frozen package context: audit.opaque.nul-byte: regular file contains a NUL byte and is treated as opaque (files: assets/a.bin)
        nul_opaque_v2_test.go:198: source_member_invalid: frozen package context: audit.opaque.nul-byte: regular file contains a NUL byte and is treated as opaque (files: assets/a.bin)
- internal/install TestDraftLaneStillBlocksNULBearingMember/audit-enabled
        nul_opaque_v2_test.go:189: source_member_invalid: frozen package context: audit.opaque.nul-byte: regular file contains a NUL byte and is treated as opaque (files: assets/a.bin)
        nul_opaque_v2_test.go:189: source_member_invalid: frozen package context: audit.opaque.nul-byte: regular file contains a NUL byte and is treated as opaque (files: assets/a.bin)
- internal/install TestV1ReinstallRefusesNULAppearingAfterInstall/collision-preserving-edit
        nul_opaque_v1_rework_test.go:67: status plan after v1-collision edit: status "ok": test: skill-a: warning: script-command-declared-only (commands.skill-a-tool.execution_policy): command 'skill-a-tool' is declared-only (no execution_policy); its capabilities bound nothing at run time; adopt execu
        nul_opaque_v1_rework_test.go:67: status plan after v1-collision edit: status "ok": test: skill-a: warning: script-command-declared-only (commands.skill-a-tool.execution_policy): command 'skill-a-tool' is declared-only (no execution_policy); its capabilities bound nothing at run time; adopt execu
- internal/install TestV1ReinstallRefusesNULAppearingAfterInstall/fresh-nul-file
        nul_opaque_v1_rework_test.go:74: status plan after NUL appears: status "ok": test: skill-a: warning: script-command-declared-only (commands.skill-a-tool.execution_policy): command 'skill-a-tool' is declared-only (no execution_policy); its capabilities bound nothing at run time; adopt execution_p
        nul_opaque_v1_rework_test.go:74: status plan after NUL appears: status "ok": test: skill-a: warning: script-command-declared-only (commands.skill-a-tool.execution_policy): command 'skill-a-tool' is declared-only (no execution_policy); its capabilities bound nothing at run time; adopt execution_p
