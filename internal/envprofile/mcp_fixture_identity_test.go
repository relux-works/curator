package envprofile

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// With no ambient Git identity, the signed-tag fixtures still reach their
// strict acceptance, changed-tag rejection, and cache-origin assertions.
// The preliminary Git refusal proves the subprocess environment cannot supply
// a missing fixture identity. TestMain replaces /dev/null with its own empty
// global config; individual transport fixtures may add only their URL rewrites.
func TestMCPSignedTagFixturesWithoutAmbientIdentity(t *testing.T) {
	home := t.TempDir()
	var env []string
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		key = strings.ToUpper(key)
		switch {
		case key == "HOME", key == "USERPROFILE", key == "XDG_CONFIG_HOME", key == "EMAIL",
			key == "GIT_DIR", key == "GIT_WORK_TREE", key == "GIT_CONFIG", key == "LC_ALL",
			strings.HasPrefix(key, "GIT_AUTHOR_"), strings.HasPrefix(key, "GIT_COMMITTER_"),
			strings.HasPrefix(key, "GIT_CONFIG_"):
			continue
		}
		env = append(env, entry)
	}
	env = append(env, "HOME="+home, "USERPROFILE="+home, "XDG_CONFIG_HOME="+home,
		"GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_SYSTEM="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1",
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	tests := []string{"TestProfileMCPSignedTagReplay", "TestProjectMCPLocalSignerCacheOrigin", "TestProjectMCPTagSignatureReplay"}
	command := exec.CommandContext(ctx, binary,
		"-test.run=^("+strings.Join(tests, "|")+")$",
		"-test.v", "-test.count=1", "-test.timeout=4m")
	command.Env = env
	output, err = command.CombinedOutput()
	if err != nil {
		t.Fatalf("signed-tag fixtures require their own Git identity: %v\n%s", err, output)
	}
	for _, name := range tests {
		if !strings.Contains(string(output), "--- PASS: "+name+" (") {
			t.Fatalf("signed-tag control did not pass %s:\n%s", name, output)
		}
	}
	t.Logf("signed-tag fixture controls:\n%s", output)
}
