package main

import (
	"context"
	"fmt"
	"os"

	"github.com/relux-works/curator-agent-launcher/internal/plan"
	"github.com/relux-works/skill-agents-management/pkg/agentic"
	"github.com/relux-works/skill-agents-management/pkg/agentic/systems/muse"
	"github.com/relux-works/skill-agents-management/pkg/providerlimits"
	"github.com/relux-works/skill-agents-management/pkg/vendorplugin"
)

func main() {
	systems := agentic.NewRegistry()
	if err := systems.Register(muse.New()); err != nil { panic(err) }
	registry := vendorplugin.NewRegistry(systems)
	if err := vendorplugin.SeedFrozenRuntimes(registry); err != nil { panic(err) }
	dir, err := os.MkdirTemp("", "muse-plan-probe-")
	if err != nil { panic(err) }
	defer os.RemoveAll(dir)
	availabilityCalls := 0
	_, err = plan.Build(context.Background(), plan.Deps{
		Registry: registry,
		BuildLaunch: vendorplugin.BuildLaunchWithEnvironment,
		Availability: func(providerlimits.VerdictQuery) (vendorplugin.Availability, error) {
			availabilityCalls++
			panic("unexpected limits call after mode refusal")
		},
	}, plan.Request{
		Runtime: "muse", Model: "muse-spark-1.3-contributor", Effort: "high",
		PermissionMode: agentic.PermissionModeNative,
		Home: dir, WorkDir: dir, Env: []string{"HOME=" + dir, "PATH=/usr/bin:/bin"},
	})
	fmt.Printf("declared_modes=%v\navailability_calls=%d\nerror=%v\n", muse.New().Capabilities().LaunchModes, availabilityCalls, err)
	if err != nil { os.Exit(1) }
}
