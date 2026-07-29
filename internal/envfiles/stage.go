package envfiles

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/staging"
)

// StageProject writes the project helper files below stageRoot and returns the
// live targets they replace. No live path is created, replaced, or removed.
func StageProject(stageRoot, projectRoot string) (staging.Plan, error) {
	shell, powerShell := ProjectContent()
	return stageEnvFiles(stageRoot, "project", ProjectDir(projectRoot), shell, powerShell)
}

// StageGlobal writes the global helper files below stageRoot and returns the
// live targets they replace.
func StageGlobal(stageRoot, home string) (staging.Plan, error) {
	shell, powerShell := GlobalContent(home)
	return stageEnvFiles(stageRoot, "global", GlobalDir(home), shell, powerShell)
}

func stageEnvFiles(stageRoot, scope, liveDir string, shell, powerShell []byte) (staging.Plan, error) {
	staged := filepath.Join(stageRoot, "env", scope)
	if err := os.MkdirAll(staged, 0o700); err != nil {
		return staging.Plan{}, fmt.Errorf("stage %s env files: %w", scope, err)
	}
	var plan staging.Plan
	for _, file := range []struct {
		name    string
		content []byte
	}{
		{ShellName, shell},
		{PowerShellName, powerShell},
	} {
		stagedPath := filepath.Join(staged, file.name)
		if err := os.WriteFile(stagedPath, file.content, 0o644); err != nil {
			return staging.Plan{}, fmt.Errorf("stage %s %s: %w", scope, file.name, err)
		}
		plan.Replace(staging.ClassEnvFile, scope+"/"+file.name, filepath.Join(liveDir, file.name), stagedPath)
	}
	return plan, nil
}
