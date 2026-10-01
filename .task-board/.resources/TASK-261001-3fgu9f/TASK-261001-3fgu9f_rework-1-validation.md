# TASK-261001-3fgu9f — rework 1 validation output

## Full suite: go test -p 1 ./... — exit 0
```
ok  	github.com/relux-works/curator-agent-launcher/cmd/curator-run	102.600s
ok  	github.com/relux-works/curator-agent-launcher/internal/axconfig	(cached)
ok  	github.com/relux-works/curator-agent-launcher/internal/cli	(cached)
ok  	github.com/relux-works/curator-agent-launcher/internal/composition	0.948s
ok  	github.com/relux-works/curator-agent-launcher/internal/configfile	(cached)
ok  	github.com/relux-works/curator-agent-launcher/internal/defaults	1.682s
ok  	github.com/relux-works/curator-agent-launcher/internal/diagnostics	1.658s
ok  	github.com/relux-works/curator-agent-launcher/internal/execution	15.724s
ok  	github.com/relux-works/curator-agent-launcher/internal/fragment	8.172s
ok  	github.com/relux-works/curator-agent-launcher/internal/mapping	(cached)
ok  	github.com/relux-works/curator-agent-launcher/internal/plan	1.722s
ok  	github.com/relux-works/curator-agent-launcher/internal/systemprompt	(cached)
```

## Narrowing pin mutant — exit 1
```
--- FAIL: TestProductionPipelineGoldens (1.70s)
    --- FAIL: TestProductionPipelineGoldens/claude_code/tracked=false (1.70s)
        pipeline_test.go:297: pipeline-claude_code-false mismatch
            got {
              "Child": {
                "Argv": [
                  "--model",
                  "claude-opus-5",
                  "--effort",
                  "medium",
                  "--append-system-prompt-file",
                  "<ROOT>/managed/prompt.md",
                  "--mcp-config",
                  "<ROOT>/managed/mcp.toml",
                  "--strict-mcp-config",
                  "",
                  "--model=native",
                  "a\nb"
                ],
                "Env": [
                  "CLAUDE_CODE_ENABLE_PROMPT_SUGGESTION=false",
                  "CLAUDE_CONFIG_DIR=<ROOT>/managed",
                  "CODEX_THREAD_ID=nested",
                  "FIGMA_API_KEY=source-secret",
                  "HOME=<ROOT>",
                  "PARENT=direct-only",
                  "PATH=<ROOT>"
                ],
                "Stdin": "cGFyZW50IHN0ZGluCgD/",
                "WorkDir": "<ROOT>"
              },
              "Stderr": "curator warning\n\u0000\ufffdcurator-run: provider: path=<ROOT>/curator-run\ncurator-run: defaults: model=claude-opus-5 (flag) effort=medium (flag)\ncurator-run: permissions=native source=default-headless mapped=none\ncurator-run: warning: profile \"default\": applied flag/append system-prompt channel \"--append-system-prompt-file\". A custom system prefix can change request caching and billing: the default may use shared prompt caching; a custom prefix forms its own cache prefix.\nhelper stderr\n"
            }
            
            want {
              "Child": {
                "Argv": [
                  "--model",
                  "claude-opus-5",
                  "--effort",
                  "medium",
                  "--append-system-prompt-file",
                  "<ROOT>/managed/prompt.md",
                  "--mcp-config",
                  "<ROOT>/managed/mcp.toml",
                  "--strict-mcp-config",
                  "",
                  "--model=native",
                  "a\nb"
                ],
                "Env": [
                  "CLAUDE_CONFIG_DIR=<ROOT>/managed",
                  "CODEX_THREAD_ID=nested",
                  "FIGMA_API_KEY=source-secret",
                  "HOME=<ROOT>",
                  "PARENT=direct-only",
                  "PATH=<ROOT>"
                ],
                "Stdin": "cGFyZW50IHN0ZGluCgD/",
                "WorkDir": "<ROOT>"
              },
              "Stderr": "curator warning\n\u0000\ufffdcurator-run: provider: path=<ROOT>/curator-run\ncurator-run: defaults: model=claude-opus-5 (flag) effort=medium (flag)\ncurator-run: permissions=native source=default-headless mapped=none\ncurator-run: warning: profile \"default\": applied flag/append system-prompt channel \"--append-system-prompt-file\". A custom system prefix can change request caching and billing: the default may use shared prompt caching; a custom prefix forms its own cache prefix.\nhelper stderr\n"
            }
FAIL
FAIL	github.com/relux-works/curator-agent-launcher/cmd/curator-run	3.707s
FAIL
```

## HOME mutant — exit 1
```
--- FAIL: TestMuseV3CompositionPreservesHOME (0.11s)
    muse_test.go:166: HOME replaced: "/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TestMuseV3CompositionPreservesHOME3872672618/001/managed"
FAIL
FAIL	github.com/relux-works/curator-agent-launcher/cmd/curator-run	1.696s
FAIL
```

## v3-rejected mutant — exit 1
```
--- FAIL: TestMuseV3Fragment (0.00s)
    v3_test.go:13: fragment: /fragment: "launch-env-fragment-v3" is not one of launch-env-fragment-v1, launch-env-fragment-v2
--- FAIL: TestV3OtherAdaptersRetainV2Rules (0.00s)
    --- FAIL: TestV3OtherAdaptersRetainV2Rules/claude_code (0.00s)
        v3_test.go:78: fragment: /fragment: "launch-env-fragment-v3" is not one of launch-env-fragment-v1, launch-env-fragment-v2
    --- FAIL: TestV3OtherAdaptersRetainV2Rules/codex_cli (0.00s)
        v3_test.go:78: fragment: /fragment: "launch-env-fragment-v3" is not one of launch-env-fragment-v1, launch-env-fragment-v2
    --- FAIL: TestV3OtherAdaptersRetainV2Rules/opencode (0.00s)
        v3_test.go:78: fragment: /fragment: "launch-env-fragment-v3" is not one of launch-env-fragment-v1, launch-env-fragment-v2
    --- FAIL: TestV3OtherAdaptersRetainV2Rules/pi (0.00s)
        v3_test.go:78: fragment: /fragment: "launch-env-fragment-v3" is not one of launch-env-fragment-v1, launch-env-fragment-v2
FAIL
FAIL	github.com/relux-works/curator-agent-launcher/internal/fragment	1.496s
--- FAIL: TestMuseV3ThroughRunPermissionBound (1.90s)
    muse_test.go:73: permission bound changed: exit=1 stderr=curator-run: resolve_fragment_invalid: curator env resolve exited 0 but its output is not a valid launch-env-fragment-v1, v2, or v3: fragment: /fragment: "launch-env-fragment-v3" is not one of launch-env-fragment-v1, launch-env-fragment-v2
--- FAIL: TestMuseV3PermissionTransportThroughRun (1.98s)
    muse_test.go:186: v3 permission transport: exit=1 builds=0 stderr=curator-run: resolve_fragment_invalid: curator env resolve exited 0 but its output is not a valid launch-env-fragment-v1, v2, or v3: fragment: /fragment: "launch-env-fragment-v3" is not one of launch-env-fragment-v1, launch-env-fragment-v2
FAIL
FAIL	github.com/relux-works/curator-agent-launcher/cmd/curator-run	5.915s
FAIL
```

## Candidate Windows vet — exit 1
```
# github.com/relux-works/curator-agent-launcher/internal/axconfig_test
# [github.com/relux-works/curator-agent-launcher/internal/axconfig_test]
vet: internal/axconfig/fifo_test.go:10:20: undefined: syscall.Mkfifo
# github.com/relux-works/curator-agent-launcher/internal/systemprompt_test
# [github.com/relux-works/curator-agent-launcher/internal/systemprompt_test]
vet: internal/systemprompt/systemprompt_test.go:230:24: undefined: syscall.Mkfifo
# github.com/relux-works/curator-agent-launcher/internal/execution
internal/execution/process.go:18:41: undefined: syscall.SYS_IOCTL
internal/execution/process.go:18:65: not enough arguments in call to syscall.Syscall
	have (unknown type, uintptr, uintptr, uintptr)
	want (uintptr, uintptr, uintptr, uintptr, uintptr)
internal/execution/process.go:29:97: undefined: syscall.SIGCONT
internal/execution/process.go:35:38: undefined: syscall.TIOCGPGRP
internal/execution/process.go:35:96: undefined: syscall.Getpgrp
internal/execution/process.go:39:41: unknown field Setpgid in struct literal of type syscall.SysProcAttr
internal/execution/process.go:41:19: cmd.SysProcAttr.Foreground undefined (type *syscall.SysProcAttr has no field or method Foreground)
internal/execution/process.go:42:19: cmd.SysProcAttr.Ctty undefined (type *syscall.SysProcAttr has no field or method Ctty)
internal/execution/process.go:45:25: undefined: syscall.SIGTTOU
internal/execution/process.go:46:30: undefined: syscall.SIGTTOU
internal/execution/process.go:46:30: too many errors
```

## Baseline Windows vet — exit 1
```
# github.com/relux-works/curator-agent-launcher/internal/axconfig_test
# [github.com/relux-works/curator-agent-launcher/internal/axconfig_test]
vet: internal/axconfig/fifo_test.go:10:20: undefined: syscall.Mkfifo
# github.com/relux-works/curator-agent-launcher/internal/systemprompt_test
# [github.com/relux-works/curator-agent-launcher/internal/systemprompt_test]
vet: internal/systemprompt/systemprompt_test.go:230:24: undefined: syscall.Mkfifo
# github.com/relux-works/curator-agent-launcher/internal/execution
internal/execution/process.go:18:41: undefined: syscall.SYS_IOCTL
internal/execution/process.go:18:65: not enough arguments in call to syscall.Syscall
	have (unknown type, uintptr, uintptr, uintptr)
	want (uintptr, uintptr, uintptr, uintptr, uintptr)
internal/execution/process.go:29:97: undefined: syscall.SIGCONT
internal/execution/process.go:35:38: undefined: syscall.TIOCGPGRP
internal/execution/process.go:35:96: undefined: syscall.Getpgrp
internal/execution/process.go:39:41: unknown field Setpgid in struct literal of type syscall.SysProcAttr
internal/execution/process.go:41:19: cmd.SysProcAttr.Foreground undefined (type *syscall.SysProcAttr has no field or method Foreground)
internal/execution/process.go:42:19: cmd.SysProcAttr.Ctty undefined (type *syscall.SysProcAttr has no field or method Ctty)
internal/execution/process.go:45:25: undefined: syscall.SIGTTOU
internal/execution/process.go:46:30: undefined: syscall.SIGTTOU
internal/execution/process.go:46:30: too many errors
```

Normalized sorted outputs equal: true. Windows vet is a failing baseline gate.

Native build: exit 0. Native vet: exit 0. git diff --check: exit 0. Base dependency/golden diff: exit 0, no output. LOGBOOK.md absent.
