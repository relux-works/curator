package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func testOrdinaryGateUnion(t *testing.T, c testLane) {
	for _, control := range []string{"valid", "merge-refused", "unserved", "missing-required-subcase", "mixed-lanes", "wrong-timeout", "race-flags", "race-artifacts"} {
		t.Run(control, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.MkdirAll(filepath.Join(dir, ".github/ci"), 0755); err != nil {
				t.Fatal(err)
			}
			write := func(path, data string) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(dir, path), []byte(data), 0700); err != nil {
					t.Fatal(err)
				}
			}
			for _, name := range []string{"test-gate.sh", "platform-case-gate.sh", "skip-classes.tsv"} {
				data, err := os.ReadFile(filepath.Join("../../.github/ci", name))
				if err != nil {
					t.Fatal(err)
				}
				write(".github/ci/"+name, string(data))
			}
			write("go.mod", "module github.com/relux-works/curator\n")
			write("manifest.json", "{}\n")
			write("ledger.tsv", fmt.Sprintf("internal/envprofile\tTestShard/native-required\t%s\t-\t-\tNative required child\n", c.goos))
			var stream strings.Builder
			for _, p := range ordinaryPackages {
				fmt.Fprintf(&stream, "{\"Action\":\"pass\",\"Package\":%q,\"Test\":\"TestShard\"}\n{\"Action\":\"pass\",\"Package\":%q,\"Elapsed\":42}\n", p, p)
			}
			if control != "missing-required-subcase" {
				fmt.Fprintf(&stream, "{\"Action\":\"pass\",\"Package\":%q,\"Test\":\"TestShard/native-required\"}\n", ordinaryPackages[0])
			}
			write("union.json", stream.String())
			write("packages.txt", strings.Join(ordinaryPackages, "\n")+"\n")
			write(".github/ci/suite-plan.sh", `#!/usr/bin/env bash
set -eu
mkdir -p "$2"
cp packages.txt "$2/plan-served.txt"
if [ "$CONTROL" = unserved ]; then
  grep -v /internal/install "$2/plan-served.txt" >remaining.txt
  mv remaining.txt "$2/plan-served.txt"
fi
echo github.com/relux-works/curator/other >>"$2/plan-served.txt"
for plan in deferred excluded assert; do : >"$2/plan-$plan.txt"; done
`)
			write("fake-go", `#!/usr/bin/env bash
set -eu
echo "$*" >>commands.txt
case "$1" in
run)
  [ "$2" = ./tools/envprofilerace ]
  [ "$3" = -lane ] && [ "$4" = "$LANE" ]
  [ "$5" = -mode ] && [ "$6" = merge ]
  [ "$7" = -dir ] && [ "$8" = shards ] && [ "$9" = -out ]
  [ "$CONTROL" != merge-refused ] || exit 17
  cp union.json "${10}"
  cp packages.txt "${10}.packages"
  ;;
test)
  for p in $(cat packages.txt); do
    for arg in "$@"; do [ "$arg" != "$p" ] || exit 33; done
  done
  printf '%s\n' '{"Action":"pass","Package":"github.com/relux-works/curator/other","Test":"TestOther"}' '{"Action":"pass","Package":"github.com/relux-works/curator/other","Elapsed":1}'
  ;;
*) exit 42 ;;
esac
`)
			cmd := exec.Command("bash", ".github/ci/test-gate.sh", "evidence")
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "GO="+filepath.ToSlash(filepath.Join(dir, "fake-go")),
				"CURATOR_CONFORMANCE_ROOT="+filepath.ToSlash(dir), "CONTROL="+control,
				"CI_WINDOWS_TEST_SHARDS=", "CI_UNIX_TEST_SHARDS=", "CI_ENVPROFILE_SHARDS=", "GO_TEST_FLAGS=", "GO_TEST_TIMEOUT="+c.timeout, "LANE="+c.lane,
				"CI_GATE_GOOS="+c.goos, "CI_PLATFORM_CASES=ledger.tsv")
			key := "CI_UNIX_TEST_SHARDS"
			if c.goos == "windows" {
				key = "CI_WINDOWS_TEST_SHARDS"
			}
			cmd.Env = append(cmd.Env, key+"=shards")
			switch control {
			case "mixed-lanes":
				cmd.Env = append(cmd.Env, "CI_UNIX_TEST_SHARDS=shards", "CI_WINDOWS_TEST_SHARDS=shards")
			case "wrong-timeout":
				cmd.Env = append(cmd.Env, "GO_TEST_TIMEOUT=90m")
			case "race-flags":
				cmd.Env = append(cmd.Env, "GO_TEST_FLAGS=-race")
			case "race-artifacts":
				cmd.Env = append(cmd.Env, "CI_ENVPROFILE_SHARDS=shards")
			}
			out, err := cmd.CombinedOutput()
			if control == "valid" {
				if err != nil || !strings.Contains(string(out), "platform-case gate: ok") {
					t.Fatalf("valid native union: %v\n%s", err, out)
				}
				return
			}
			code := 1
			if control == "unserved" || control == "mixed-lanes" || control == "wrong-timeout" || control == "race-flags" || control == "race-artifacts" {
				code = 2
			}
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != code {
				t.Fatalf("want intended exit %d, got %v\n%s", code, err, out)
			}
			if control == "missing-required-subcase" {
				if !strings.Contains(string(out), "required case never ran on "+c.goos+": internal/envprofile :: TestShard/native-required") {
					t.Fatalf("wrong subcase rejection: %s", out)
				}
			} else {
				want := ""
				switch control {
				case "mixed-lanes":
					want = "ordinary shard lanes are mutually exclusive"
				case "wrong-timeout", "race-flags", "race-artifacts":
					want = "shard evidence requires the ordinary " + c.timeout + " lane"
				case "unserved":
					want = "package is not served: " + ordinaryPackages[1]
				}
				if want != "" && !strings.Contains(string(out), want) {
					t.Fatalf("wrong rejection: want %s: %s", want, out)
				}
				commands, err := os.ReadFile(filepath.Join(dir, "commands.txt"))
				if (err != nil && !os.IsNotExist(err)) || strings.Contains(string(commands), "test -json") {
					t.Fatalf("tests ran after rejected admission: %s, %v", commands, err)
				}
			}
		})
	}
}

// The real shell and unchanged ledger gate admit complete native unions and
// reject the intended missing-child/configuration controls on each platform.
// Go execution/planning are bounded stand-ins; native CI supplies runtime proof.
func TestOrdinaryGateUnion(t *testing.T) {
	for _, goos := range []string{"windows", "linux", "darwin"} {
		t.Run(goos, func(t *testing.T) {
			lane := "unix-test"
			if goos == "windows" {
				lane = "windows-test"
			}
			c, err := ordinaryConfig(lane, goos)
			if err != nil {
				t.Fatal(err)
			}
			testOrdinaryGateUnion(t, c)
		})
	}
}
