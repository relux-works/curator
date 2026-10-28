package envprofile

import (
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/envregistry"
)

// MachineFromConfig shares the already loaded effective machine policy with
// managed-home lock transitions; it never reloads a weaker configuration.
func MachineFromConfig(cfg *config.Config) envregistry.MachineConfig {
	machine := envregistry.DefaultMachineConfig()
	if cfg == nil {
		return machine
	}
	machine.Forms = cfg.Env.Forms
	machine.PassableEnvNames = cfg.Env.PassableEnvNames
	machine.PassableEnvNamesSet = cfg.Env.PassableEnvNamesSet
	machine.Isolation = cfg.Env.Isolation
	machine.CredentialStore = cfg.Env.CredentialStore
	machine.UserIsolation = cfg.UserIsolation
	machine.IsolationLocked = cfg.Locked["environments.isolation"]
	machine.IsolationLockSource = cfg.SystemConfigPath
	machine.Permissions = cfg.Env.Permissions
	machine.PermissionsLocked = cfg.Locked["environments.permissions"]
	return machine
}
