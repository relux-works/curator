# Accepted plan contract clarification

The original BuildPlan checklist wording is superseded by already accepted SPEC0.3.0 section4.4 (launcher PR7), not a new product decision. BuildLaunch owns runtime/vendor/model/effort admission; never bypass it through bare BuildPlan or reconstruct EffortSupport. Request Runtime matches mapping, Model/Effort are explicitly resolved, Home is managed fragment Home, WorkDir current, Env os.Environ in BOTH modes, Composition empty, Run zero; goal/budget/service-tier/assignment unset. Named LaunchModeInteractive, no permission bypass/yolo. Plan building starts no child.

Provider limits are a separate explicit Store.AvailabilityFor read keyed by exact Runtime/Model/managed Home. Only Serviceable healthy admits; determinate no-record follows the module contract, unknown/failed read never implies healthy. Surface state/Until/Checked/Observed/Failures and module errors; no retry, fallback or model downgrade. Add launcher --effort guidance to required-effort refusals.

Read accepted A0 and current SPEC4.2-4.4, reverify only changed actual tag/runtime. Native Pi v0.5.11 is operator-only and absent at last remote check. No agent tags/pseudo-version/local replace/committed workspace override. Actual three-adapter tagged behavior is required before final delivery. Full main TASK-260908-1o7i8y must call these APIs; isolated tests are not main wiring proof.

Muse Spark xhigh producer, Astra medium reviewer; normal tracked public CR and signed exact-head delivery. Necessary local checks, no hosted CI; do not manually duplicate runtime make check. No real ax, runtime-home edits or control-root source writes. Preserve accepted defaults checkpoint and landed diagnostics/execution/composition/systemprompt APIs.
