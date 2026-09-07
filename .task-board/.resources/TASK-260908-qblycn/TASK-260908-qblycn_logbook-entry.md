## 2026-09-08

### 0340 — A0 verification: pi cannot reach the managed home via agents-management v0.5.10
- FINDING: pi system plugin plans Binary=agents-infra, argv [pi --model M] (pkg/agentic/systems/pi/binary.go); agents-infra strips inherited PI_CODING_AGENT_DIR (tools/agents-infra/internal/infra/pi_launch_posix.go:137). curator run pi as specified never runs in the Curator pi home.
- FINDING: no frozen RuntimeDeclaration has System pi; pi only via local-models vendor + absent local-models.toml. SPEC 4.3 lineup and 4.4 plan impossible for pi at the tag.
- FINDING: agentic.BuildPlan needs Model.Effort (EffortSupport) from a vendor row and never reads provider limits; the real path is vendorplugin.BuildLaunch + providerlimits.Store.AvailabilityFor(VerdictQuery{Runtime,Model,Home}).
- FINDING: Plan.Env is a full filtered env (claude drops CLAUDECODE); SPEC 4.5 layer 1+2 overlay re-admits it.
- FINDING: pi 0.84.2 --append-system-prompt suppresses APPEND_SYSTEM.md discovery; trusted cwd .pi/SYSTEM.md wins over agent dir (dist/core/resource-loader.js:808-829).
- FINDING: codex 0.153.4 -p missing layer silently ignored (exit 0); -p twice = clap error; layer-is-directory = config load error. --strict-config not testable on non-runtime subcommands.
- FINDING: claude 2.1.263 --effort bogus only warns and exits 0; --mcp-config is variadic so --strict-mcp-config must follow the path before native args.
- DECISION: installed curator updated from clean remote main 04550e28 (isolated clone, GOBIN go install) per parent nudge; backup in task .temp. Old rc.3 binary lacked env/profile.
- STATUS: errata E1-E6 in TASK-260908-qblycn_a0-verification-findings.md; erratum stories precede A1 code.