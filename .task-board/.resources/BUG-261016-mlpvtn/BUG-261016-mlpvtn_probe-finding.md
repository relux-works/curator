# Probe deadline finding

The production entry point internal/install/builddeps.go calls godriver.ConfigFromEnvironment and godriver.Probe. internal/godriver/session.go fixes defaultProbeTimeout to 15 seconds and clamps ProbeTimeout above that bound. ConfigFromEnvironment reads only CURATOR_GO and GOROOT; no existing documented CI deadline knob was found. No product flag or timeout change is introduced.

Workflow/ref concurrency cannot constrain other repositories or bound host-wide load. The requested three consecutive green self-hosted main runs are post-integration rollout evidence, not candidate gate evidence. LINUX_SELF_HOSTED was already true on this run; this producer did not toggle it.

Logbook: Recorded during the Linux CI isolation fix; loaded-host qualification and actual historical ABI linker reproduction remain stated bounds, not inferred successes.