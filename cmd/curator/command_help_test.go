package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func buildHelpCLI(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "curator"+productionBinarySuffix())
	command := exec.Command("go", "build", "-o", binary, ".")
	if out, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, out)
	}
	return binary
}

type helpSandbox struct {
	root string
	cwd  string
	env  []string
}

func newHelpSandbox(t *testing.T) helpSandbox {
	t.Helper()
	root := t.TempDir()
	sandbox := helpSandbox{root: root, cwd: filepath.Join(root, "cwd")}
	for _, name := range []string{"cwd", "home", "curator", "tmp", "xdg"} {
		dir := filepath.Join(root, name)
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "sentinel"), []byte("preserve existing bytes\n"), 0o640); err != nil {
			t.Fatal(err)
		}
	}
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(key, "CURATOR_") || strings.HasPrefix(key, "XDG_") {
			continue
		}
		switch key {
		case "HOME", "USERPROFILE", "TMPDIR", "TMP", "TEMP", "GIT_CONFIG_GLOBAL", "GIT_CONFIG_NOSYSTEM":
			continue
		}
		sandbox.env = append(sandbox.env, entry)
	}
	sandbox.env = append(sandbox.env,
		"HOME="+filepath.Join(root, "home"), "USERPROFILE="+filepath.Join(root, "home"),
		"CURATOR_HOME="+filepath.Join(root, "curator"), "CURATOR_CONFIG="+filepath.Join(root, "curator", "config.json"),
		"XDG_CONFIG_HOME="+filepath.Join(root, "xdg"), "TMPDIR="+filepath.Join(root, "tmp"),
		"TMP="+filepath.Join(root, "tmp"), "TEMP="+filepath.Join(root, "tmp"),
		"GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1")
	command := exec.Command("git", "init", "--quiet")
	command.Dir, command.Env = sandbox.cwd, sandbox.env
	if out, err := command.CombinedOutput(); err != nil {
		t.Fatalf("empty git repo: %v\n%s", err, out)
	}
	return sandbox
}

type helpTreeEntry struct {
	Size int64
	Mode fs.FileMode
	Hash [sha256.Size]byte
}

func snapshotHelpTree(t *testing.T, root string) map[string]helpTreeEntry {
	t.Helper()
	entries := make(map[string]helpTreeEntry)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		value := helpTreeEntry{Size: info.Size(), Mode: info.Mode()}
		if info.Mode().IsRegular() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			value.Hash = sha256.Sum256(data)
		}
		entries[rel] = value
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return entries
}

func (sandbox helpSandbox) invoke(t *testing.T, binary string, args ...string) (int, string, string) {
	t.Helper()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	command := exec.CommandContext(ctx, binary, args...)
	command.Dir, command.Env = sandbox.cwd, sandbox.env
	var stdout, stderr strings.Builder
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	if ctx.Err() != nil {
		t.Fatalf("CLI %v timed out", args)
	}
	code := 0
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			code = exit.ExitCode()
		} else {
			t.Fatal(err)
		}
	}
	return code, stdout.String(), stderr.String()
}

// Every registered built-in and child topic answers both help flags at each
// command-path position and after operands, plus help <command>, via main().
// Persistent tree paths, sizes, modes and file bytes must remain unchanged;
// snapshots cannot observe a transient write that is removed before exit.
func TestRegisteredSubcommandHelpNoWrites(t *testing.T) {
	binary := buildHelpCLI(t)
	seen := make(map[string]bool)
	for _, command := range commandRegistry {
		t.Run(strings.ReplaceAll(command.name, " ", "/"), func(t *testing.T) {
			if seen[command.name] {
				t.Fatalf("duplicate registry entry %q", command.name)
			}
			seen[command.name] = true
			path := strings.Fields(command.name)
			cases := [][]string{append([]string{"help"}, path...)}
			for _, flag := range []string{"-h", "--help"} {
				for position := 1; position <= len(path); position++ {
					args := append([]string{}, path[:position]...)
					args = append(args, flag)
					args = append(args, path[position:]...)
					cases = append(cases, args)
				}
				cases = append(cases, append(append([]string{}, path...), "operand", "--invalid-option", flag, "--", "--help"))
			}
			sandbox := newHelpSandbox(t)
			var dedicated string
			for index, args := range cases {
				t.Run(fmt.Sprint(index), func(t *testing.T) {
					before := snapshotHelpTree(t, sandbox.root)
					code, stdout, stderr := sandbox.invoke(t, binary, args...)
					after := snapshotHelpTree(t, sandbox.root)
					if !reflect.DeepEqual(before, after) {
						t.Errorf("%v wrote to the isolated tree: before=%v after=%v", args, before, after)
					}
					if code != 0 || stderr != "" || !strings.Contains(stdout, "curator "+command.name) {
						t.Fatalf("%v: exit=%d stdout=%q stderr=%q; want dedicated help and exit 0", args, code, stdout, stderr)
					}
					if index == 0 {
						dedicated = stdout
					} else if stdout != dedicated {
						t.Fatalf("%v differs from help %s", args, command.name)
					}
				})
			}
		})
	}
}

// The no-help positive control reaches cmdInit through main() and creates both
// project files; top-level help/version controls still preserve the tree.
func TestCommandHelpInitPositiveControl(t *testing.T) {
	binary := buildHelpCLI(t)
	sandbox := newHelpSandbox(t)
	for _, args := range [][]string{{"-h"}, {"--help"}, {"help"}, {"--version"}} {
		before := snapshotHelpTree(t, sandbox.root)
		if code, _, stderr := sandbox.invoke(t, binary, args...); code != 0 || stderr != "" {
			t.Fatalf("%v: exit=%d stderr=%q", args, code, stderr)
		}
		if !reflect.DeepEqual(before, snapshotHelpTree(t, sandbox.root)) {
			t.Fatalf("%v changed the tree", args)
		}
	}
	if code, stdout, stderr := sandbox.invoke(t, binary, "init"); code != 0 {
		t.Fatalf("init: exit=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	for _, name := range []string{"Skillfile.json", ".gitignore"} {
		if info, err := os.Stat(filepath.Join(sandbox.cwd, name)); err != nil || info.Size() == 0 {
			t.Fatalf("init did not create nonempty %s: %v", name, err)
		}
	}
}

// The real run provider receives -- and help-looking operands byte-for-byte;
// a help flag before -- instead prevents provider execution (no JSON output).
func TestCommandHelpRunPassthrough(t *testing.T) {
	binary := buildHelpCLI(t)
	sandbox := newHelpSandbox(t)
	providerDir := t.TempDir()
	provider := filepath.Join(providerDir, "curator-run"+productionBinarySuffix())
	source := filepath.Join(providerDir, "main.go")
	if err := os.WriteFile(source, []byte(`package main
import ("encoding/json"; "os")
func main() { _ = json.NewEncoder(os.Stdout).Encode(os.Args[1:]) }
`), 0o600); err != nil {
		t.Fatal(err)
	}
	build := exec.Command("go", "build", "-o", provider, source)
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build provider: %v\n%s", err, out)
	}
	// Revision A accepts this PATH-only fixture with a warning. A future
	// enforcing revision must configure it in provider_directories instead.
	for i, entry := range sandbox.env {
		if strings.HasPrefix(entry, "PATH=") {
			sandbox.env[i] = "PATH=" + providerDir + string(os.PathListSeparator) + os.Getenv("PATH")
		}
	}
	for _, operands := range [][]string{{"example", "--", "-h"}, {"example", "--", "--help", "-h"}, {"--", "-h"}} {
		code, stdout, stderr := sandbox.invoke(t, binary, append([]string{"run"}, operands...)...)
		var got []string
		if code != 0 || json.Unmarshal([]byte(stdout), &got) != nil || !reflect.DeepEqual(got, operands) {
			t.Fatalf("run %v: exit=%d stdout=%q stderr=%q", operands, code, stdout, stderr)
		}
	}
	code, stdout, stderr := sandbox.invoke(t, binary, "run", "example", "-h", "--", "--help")
	if code != 0 || stderr != "" || !strings.HasPrefix(stdout, "usage: curator run ") {
		t.Fatalf("run help: exit=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}

// Unknown explicit help topics are usage errors and never execute a command.
func TestCommandHelpUnknownTopic(t *testing.T) {
	var stdout, stderr strings.Builder
	code := run([]string{"help", "unknown-topic"}, stubConfigSource{}, &stdout, &stderr)
	if code != exitUsage || stdout.Len() != 0 || !strings.Contains(stderr.String(), "unknown help topic") {
		t.Fatalf("unknown topic: exit=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}
