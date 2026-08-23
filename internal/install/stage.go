package install

import (
	"context"
	"fmt"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/closureexec"
)

// StagedBuild is one privately staged build output. Path stays inside the
// operation-private session root: nothing here is published to the protected
// cache, an install marker, a runtime target, or a shim.
type StagedBuild struct {
	skill            string
	command          string
	key              buildmeta.CacheKey
	path             string
	receipt          buildmeta.Receipt
	executionReceipt closureexec.BuildSessionReceipt
	artifact         buildmeta.Artifact
}

// Skill is the closure node that declared the command.
func (staged StagedBuild) Skill() string { return staged.skill }

// Command is the exported command name.
func (staged StagedBuild) Command() string { return staged.command }

// CacheKey is the protected build-cache key the output would be published to.
func (staged StagedBuild) CacheKey() buildmeta.CacheKey { return staged.key }

// Path is the operation-private staged artifact path.
func (staged StagedBuild) Path() string { return staged.path }

// Receipt is the complete CCJ-1 receipt for the staged output.
func (staged StagedBuild) Receipt() buildmeta.Receipt { return staged.receipt }

// ExecutionReceipt is the exact portable or verified build-session evidence
// required before protected publication.
func (staged StagedBuild) ExecutionReceipt() closureexec.BuildSessionReceipt {
	return staged.executionReceipt
}

// Artifact is the verified artifact metadata of the staged output.
func (staged StagedBuild) Artifact() buildmeta.Artifact { return staged.artifact }

// Describe renders one staged output for the installation report.
func (staged StagedBuild) Describe() string {
	return fmt.Sprintf("%s.%s staged key=%s sha256=%s", staged.skill, staged.command, staged.key, staged.artifact.SHA256)
}

// Staged is the operation-private result of one staging pass. It describes
// what a later atomic commit may publish; it changes no live target itself and
// its lifetime is owned by the plan that produced it.
type Staged struct {
	scope  string
	builds []StagedBuild
}

// Builds returns the staged outputs in provider-first, command-lexical order.
func (staged Staged) Builds() []StagedBuild {
	return append([]StagedBuild(nil), staged.builds...)
}

// Empty reports whether the pass staged nothing, either because the closure
// declares no build command or because every planned command was a cache hit.
func (staged Staged) Empty() bool { return len(staged.builds) == 0 }

// Lines renders one scope-prefixed report line per staged output.
func (staged Staged) Lines() []string {
	lines := make([]string, 0, len(staged.builds))
	for _, build := range staged.builds {
		lines = append(lines, staged.scope+": "+build.Describe())
	}
	return lines
}

// stageBuilds builds every planned miss in operation-private staging, in the
// plan's provider-first, command-lexical order.
//
// A cache hit performs no source-aware Go command. A failure at any point
// returns an error without publishing anything, so closing the plan removes
// the whole operation-private root, including outputs staged earlier in the
// same pass, and leaves the installation and the live cache untouched.
//
// The pass ends by finalizing trust: the toolchain is re-fingerprinted through
// the last build child and every frozen source is rechecked. A staged result is
// therefore returned only when the identities the plan committed to still hold,
// and that happens before the caller hands the result anywhere.
func stageBuilds(ctx context.Context, plan BuildPlan, deps BuildDeps) (Staged, error) {
	staged := Staged{scope: plan.scope}
	if plan.Empty() {
		return staged, nil
	}
	if plan.session == nil {
		return staged, fmt.Errorf("staging requires a trusted Go session")
	}
	for _, build := range plan.builds {
		if build.outcome.blocking() {
			return Staged{}, fmt.Errorf("%s.%s is %s: %s", build.skill, build.command, build.outcome, build.reason)
		}
		if !build.outcome.buildable() {
			continue
		}
		source := plan.sources[build.skill]
		if source == nil {
			return Staged{}, fmt.Errorf("%s.%s: build source was not validated", build.skill, build.command)
		}
		artifact, err := deps.Builder.Stage(ctx, StageRequest{
			Session:       plan.session,
			Source:        source,
			CommandObject: build.commandObject,
			BuildRoot:     build.buildRoot,
			SourceDir:     build.sourceDir,
			Command:       build.command,
		})
		if err != nil {
			return Staged{}, fmt.Errorf("%s.%s: %w", build.skill, build.command, err)
		}
		receipt, err := buildmeta.NewReceipt(build.input, artifact.Metadata)
		if err != nil {
			return Staged{}, fmt.Errorf("%s.%s: %w", build.skill, build.command, err)
		}
		if receipt.CacheKey != build.logicalKey {
			return Staged{}, fmt.Errorf("%s.%s: staged receipt input key %s does not match the planned input key %s",
				build.skill, build.command, receipt.CacheKey, build.logicalKey)
		}
		staged.builds = append(staged.builds, StagedBuild{
			skill:            build.skill,
			command:          build.command,
			key:              build.key,
			path:             artifact.Path,
			receipt:          receipt,
			executionReceipt: artifact.ExecutionReceipt,
			artifact:         artifact.Metadata,
		})
	}
	if err := plan.Verify(ctx); err != nil {
		return Staged{}, err
	}
	return staged, nil
}
