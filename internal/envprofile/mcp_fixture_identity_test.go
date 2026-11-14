package envprofile

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// With no ambient Git identity, the signed-tag fixtures still reach their
// strict acceptance, changed-tag rejection, and cache-origin assertions.
// The preliminary Git refusal proves the subprocess environment cannot supply
// a missing fixture identity. TestMain replaces the input with its own empty
// global config; individual transport fixtures may add only their URL rewrites.
func TestMCPSignedTagFixturesWithoutAmbientIdentity(t *testing.T) {
	runFixturesWithoutAmbientIdentity(t, []string{"TestProfileMCPSignedTagReplay", "TestProjectMCPLocalSignerCacheOrigin", "TestProjectMCPTagSignatureReplay"})
}

func runFixturesWithoutAmbientIdentity(t *testing.T, tests []string) {
	t.Helper()
	home := t.TempDir()
	global := filepath.Join(home, "gitconfig")
	if err := os.WriteFile(global, nil, 0600); err != nil {
		t.Fatal(err)
	}
	var env []string
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		key = strings.ToUpper(key)
		switch {
		case key == "HOME", key == "USERPROFILE", key == "XDG_CONFIG_HOME", key == "EMAIL",
			key == "GIT_DIR", key == "GIT_WORK_TREE", key == "GIT_CONFIG", key == "LC_ALL", key == "GIT_CEILING_DIRECTORIES",
			strings.HasPrefix(key, "GIT_AUTHOR_"), strings.HasPrefix(key, "GIT_COMMITTER_"),
			strings.HasPrefix(key, "GIT_CONFIG_"):
			continue
		}
		env = append(env, entry)
	}
	env = append(env, "HOME="+home, "USERPROFILE="+home, "XDG_CONFIG_HOME="+home,
		"GIT_CEILING_DIRECTORIES="+home, "GIT_CONFIG_GLOBAL="+global, "GIT_CONFIG_SYSTEM="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_COUNT=1", "GIT_CONFIG_KEY_0=user.useConfigOnly", "GIT_CONFIG_VALUE_0=true", "LC_ALL=C")

	probe := exec.Command("git", "-C", home, "var", "GIT_COMMITTER_IDENT")
	probe.Env = env
	output, err := probe.CombinedOutput()
	if err == nil || !strings.Contains(string(output), "auto-detection is disabled") {
		t.Fatalf("identity-free control must refuse Git identity: %v\n%s", err, output)
	}

	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	// These process limits only bound a stuck fixture, not signature latency.
	// Success requires all named children to report their real assertions.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, binary,
		"-test.run=^("+strings.Join(tests, "|")+")$",
		"-test.v", "-test.count=1", "-test.timeout=4m")
	command.Env = env
	output, err = command.CombinedOutput()
	if err != nil {
		t.Fatalf("fixtures require their own Git identity: %v\n%s", err, output)
	}
	for _, name := range tests {
		if !strings.Contains(string(output), "--- PASS: "+name+" (") {
			t.Fatalf("fixture control did not pass %s:\n%s", name, output)
		}
	}
	t.Logf("identity-free fixture controls:\n%s", output)
}
