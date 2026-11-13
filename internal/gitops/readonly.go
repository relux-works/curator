package gitops

import (
	"os"
	"strings"
)

// ReadOnlyEnv isolates local object inspection without creating temporary
// configuration files. It forbids every transport, implicit object hydration,
// optional lock writes and repository-configured executable helpers.
func ReadOnlyEnv() []string {
	env := []string{}
	for _, entry := range os.Environ() {
		name, _, _ := strings.Cut(entry, "=")
		if isDraftAllowedEnv(name) && !strings.EqualFold(name, "SSH_AUTH_SOCK") && !strings.EqualFold(name, "GIT_ASKPASS") {
			env = append(env, entry)
		}
	}
	return append(env,
		"LANG=C", "LC_ALL=C",
		"GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_SYSTEM="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1",
		"GIT_ALLOW_PROTOCOL=", "GIT_TERMINAL_PROMPT=0", "GIT_PROTOCOL_FROM_USER=0",
		"GIT_NO_LAZY_FETCH=1", "GIT_NO_REPLACE_OBJECTS=1", "GIT_OPTIONAL_LOCKS=0",
		"GIT_CONFIG_COUNT=5",
		"GIT_CONFIG_KEY_0=core.hooksPath", "GIT_CONFIG_VALUE_0="+os.DevNull,
		"GIT_CONFIG_KEY_1=core.fsmonitor", "GIT_CONFIG_VALUE_1=false",
		"GIT_CONFIG_KEY_2=credential.helper", "GIT_CONFIG_VALUE_2=",
		"GIT_CONFIG_KEY_3=gc.auto", "GIT_CONFIG_VALUE_3=0",
		"GIT_CONFIG_KEY_4=maintenance.auto", "GIT_CONFIG_VALUE_4=false",
	)
}
