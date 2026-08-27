package scopes

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/staging"
)

// StageConsumer renders the machine-wide consumer registry that must exist
// after this checkout is installed, and returns it as the last commit target.
//
// The merge deliberately reads the live registry: the caller runs this while
// holding the manager-home mutation lock, so a checkout registered by another
// project that committed first is folded in rather than overwritten. Staging it
// before the lock would let the later of two concurrent installs erase the
// earlier one's consumer.
func StageConsumer(stageRoot, home, projectRoot string) (staging.Plan, error) {
	resolved, err := filepath.Abs(projectRoot)
	if err != nil {
		resolved = projectRoot
	}
	set := map[string]bool{resolved: true}
	existing, err := readConsumers(home)
	if err != nil {
		return staging.Plan{}, fmt.Errorf("read the consumer registry %s: %w", filepath.Join(home, ConsumersName), err)
	}
	for _, entry := range existing {
		set[entry] = true
	}
	payload, err := ConsumersPayload(set)
	if err != nil {
		return staging.Plan{}, fmt.Errorf("render consumer registry: %w", err)
	}
	staged := filepath.Join(stageRoot, "consumers")
	if err := os.MkdirAll(staged, 0o700); err != nil {
		return staging.Plan{}, fmt.Errorf("stage consumer registry: %w", err)
	}
	stagedPath := filepath.Join(staged, ConsumersName)
	if err := os.WriteFile(stagedPath, payload, 0o644); err != nil {
		return staging.Plan{}, fmt.Errorf("stage consumer registry: %w", err)
	}
	var plan staging.Plan
	plan.Replace(staging.ClassConsumer, ConsumersName, filepath.Join(home, ConsumersName), stagedPath)
	return plan, nil
}
