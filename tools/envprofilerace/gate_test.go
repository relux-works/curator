package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The real shell entry point must consume the verified union, avoid rerunning
// envprofile, and apply the real platform gate to required subcases. The Go
// launcher/planner are bounded stand-ins; merge validation is tested separately.
func TestGateUsesVerifiedShardUnion(t *testing.T) {
	for _, control := range []string{"valid", "merge-refused", "missing-required-subcase"} {
		t.Run(control, func(t *testing.T) {
			dir := t.TempDir()
			ci := filepath.Join(dir, ".github", "ci")
			if err := os.MkdirAll(ci, 0755); err != nil {
				t.Fatal(err)
			}
			write := func(path, body string) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(dir, path), []byte(body), 0700); err != nil {
					t.Fatal(err)
				}
			}
			for _, file := range []string{"test-gate.sh", "platform-case-gate.sh"} {
				data, err := os.ReadFile(filepath.Join("../../.github/ci", file))
				if err != nil {
					t.Fatal(err)
				}
				write(filepath.Join(".github/ci", file), string(data))
			}
			write("go.mod", "module github.com/relux-works/curator\n")
			write("manifest.json", "{}\n")
			write("ledger.tsv", "internal/envprofile\tTestShard/required\tlinux,darwin,windows\t-\t-\tRequired shard subcase\n")
			classes, err := os.ReadFile("../../.github/ci/skip-classes.tsv")
			if err != nil {
				t.Fatal(err)
			}
			write("classes.tsv", string(classes))
			write(".github/ci/suite-plan.sh", `#!/usr/bin/env bash
set -eu
mkdir -p "$2"
printf '%s\n' github.com/relux-works/curator/internal/envprofile github.com/relux-works/curator/other >"$2/plan-served.txt"
for plan in deferred excluded assert; do : >"$2/plan-$plan.txt"; done
`)
			var stream strings.Builder
			if control != "missing-required-subcase" {
				fmt.Fprintf(&stream, "{\"Action\":\"pass\",\"Package\":%q,\"Test\":\"TestShard/required\"}\n", pkg)
			}
			stream.Write(fixtureStream([]string{"TestShard"}))
			write("union.json", stream.String())
			write("fake-go", `#!/usr/bin/env bash
set -eu
printf '%s\n' "$*" >>"$FIXTURE_ROOT/commands.txt"
case "$1" in
run)
  [ "$2" = ./tools/envprofilerace ]
  [ "$3" = -mode ] && [ "$4" = merge ]
  [ "$5" = -dir ] && [ "$6" = shards ] && [ "$7" = -out ]
  [ "$MERGE_REFUSED" = 0 ] || exit 17
  cp "$FIXTURE_ROOT/union.json" "$8"
  ;;
test)
  for arg in "$@"; do
    if [ "$arg" = github.com/relux-works/curator/internal/envprofile ]; then
      echo 'envprofile must not run again after a verified union' >&2
      exit 33
    fi
  done
  printf '%s\n' '{"Action":"pass","Package":"github.com/relux-works/curator/other","Test":"TestOther"}' '{"Action":"pass","Package":"github.com/relux-works/curator/other","Elapsed":1}'
  ;;
*) exit 42 ;;
esac
`)
			cmd := exec.Command("bash", ".github/ci/test-gate.sh", "evidence")
			cmd.Dir = dir
			// This fixture intentionally replaces its own gate inputs and Go launcher.
			cmd.Env = append(os.Environ(),
				"GO="+filepath.ToSlash(filepath.Join(dir, "fake-go")),
				"CURATOR_CONFORMANCE_ROOT="+filepath.ToSlash(dir),
				"CI_ENVPROFILE_SHARDS=shards", "GO_TEST_FLAGS=-race",
				"CI_PLATFORM_CASES=ledger.tsv", "CI_SKIP_CLASSES=classes.tsv",
				"CI_GATE_GOOS=linux", "FIXTURE_ROOT="+filepath.ToSlash(dir), "MERGE_REFUSED=0")
			if control == "merge-refused" {
				cmd.Env = append(cmd.Env, "MERGE_REFUSED=1")
			}
			out, err := cmd.CombinedOutput()
			if control == "valid" {
				if err != nil || !strings.Contains(string(out), "platform-case gate: ok") {
					t.Fatalf("valid union: %v\n%s", err, out)
				}
				commands, err := os.ReadFile(filepath.Join(dir, "commands.txt"))
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(commands), "-mode merge -dir shards") || !strings.Contains(string(commands), "test -json") {
					t.Fatalf("missing merge or remaining suite: %s", commands)
				}
				return
			}
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 {
				t.Fatalf("want intended gate exit 1, got %v\n%s", err, out)
			}
			if control == "merge-refused" {
				commands, err := os.ReadFile(filepath.Join(dir, "commands.txt"))
				if err != nil {
					t.Fatal(err)
				}
				if strings.Contains(string(commands), "test -json") {
					t.Fatal("tests ran after refused merge")
				}
			} else if !strings.Contains(string(out), "required case never ran on linux: internal/envprofile :: TestShard/required") {
				t.Fatalf("missing intended required-subcase verdict:\n%s", out)
			}
		})
	}
}
