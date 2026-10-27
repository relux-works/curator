package install

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/stateread"
)

// PrepareProfileSkillPlan is the explicit PR1 preparation entry. It runs the
// selected closure through ordinary requirement, registry, driver, assurance,
// private compilation and receipt verification before the profile manager
// publishes lock-v5. It publishes no cache artifacts, commands or runtime.
// The private build session survives until the record transaction finishes.
func PrepareProfileSkillPlan(cfg *config.Config, profile, expectedHash string, policy envprofile.Policy, opts Options) (envprofile.SkillPlanResult, error) {
	private := &privateRoot{prefix: operationPrivatePrefix}
	defer func() { _ = private.remove() }()
	var plan BuildPlan
	defer func() { _ = plan.Release() }()
	return envprofile.PrepareSkillPlan(cfg.Home(), profile, expectedHash, policy, func(nodes []*closure.Node) error {
		validation := Result{}
		if !validateNodes(nodes, "", "profile:"+profile, &validation) {
			return fmt.Errorf("%s", strings.Join(validation.Errors, "; "))
		}
		if err := checkSystemCommands(nodes); err != nil {
			return err
		}
		if err := checkLegacySkillDependencies(nodes); err != nil {
			return err
		}
		// Profile source admission/strict audit already ran on every extraction.
		// Preserve the ordinary registry availability and attestation gates.
		registryResult, err := resolveRegistryEvidence(cfg, nodes, "profile:"+profile, true, false, opts)
		if err != nil {
			return err
		}
		result := Result{}
		if addRegistryDiagnostic(&result, cfg, registryResult) {
			return fmt.Errorf("%s", strings.Join(result.Errors, "; "))
		}
		deps, err := opts.Build.resolve(cfg.Home(), private, []string{GlobalRoot(cfg.Home()), filepath.Join(cfg.Home(), "runtime"), cfg.SkillsRoot})
		if err != nil {
			return err
		}
		plan, err = planBuilds(opts.context(), buildPlanRequest{scope: "profile:" + profile, nodes: nodes, deps: deps})
		if err != nil {
			return err
		}
		external, err := planExternalBuilds(opts.context(), "profile:"+profile, "profile:"+profile, cfg.Home(), nodes, nil, nil, deps.Toolchain, opts.External, deps.Assurance, false)
		if err != nil {
			return err
		}
		stagedExternal, err := stageExternalBuilds(opts.context(), external, deps.Toolchain, deps.Builder, private)
		if err != nil {
			return fmt.Errorf("profile_skill_compile_failed: %w", err)
		}
		staged, err := stageBuilds(opts.context(), plan, deps)
		if err != nil {
			return fmt.Errorf("profile_skill_compile_failed: %w", err)
		}
		if opts.OnStaged != nil {
			if err := opts.OnStaged(staged); err != nil {
				return err
			}
		}
		for _, build := range staged.Builds() {
			if err := verifyProfileArtifact(build.Path(), build.Artifact().SHA256, build.Artifact().Size); err != nil {
				return err
			}
		}
		for _, commands := range stagedExternal.entries {
			for _, entry := range commands {
				if err := verifyProfileArtifact(entry.artifactPath, entry.record.ArtifactSHA256, -1); err != nil {
					return err
				}
				receipt, err := stateread.ReadRegularFile(entry.receiptPath)
				if err != nil || receipt.Kind != stateread.KindPresent || fmt.Sprintf("sha256:%x", sha256.Sum256(receipt.Bytes)) != string(entry.record.ReceiptSHA256) {
					return fmt.Errorf("profile_skill_plan_binding_invalid: external receipt changed")
				}
			}
		}
		return plan.Verify(opts.context())
	})
}

// Final staged-byte observation catches replacement after compiler/receipt checks.
func verifyProfileArtifact(path, digest string, size int64) error {
	file, err := stateread.ReadRegularFile(path)
	if err != nil || file.Kind != stateread.KindPresent || fmt.Sprintf("sha256:%x", sha256.Sum256(file.Bytes)) != digest || (size >= 0 && int64(len(file.Bytes)) != size) {
		return fmt.Errorf("profile_skill_plan_binding_invalid: staged artifact changed")
	}
	info, err := os.Lstat(path)
	if err != nil || (runtime.GOOS != "windows" && info.Mode().Perm()&0111 == 0) {
		return fmt.Errorf("profile_skill_plan_binding_invalid: staged artifact is not executable")
	}
	return nil
}
