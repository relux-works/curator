// Command stubinterp is a test-only stand-in for a script interpreter. It
// records how the worker started it — argument vector, executable path,
// working directory, standard input — so tests prove verbatim forwarding,
// stream binding, and exit-status return without depending on a host
// node or python installation. Behaviour is driven by STUB_* environment
// variables the test sets in the session environment.
package main

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type report struct {
	Executable string            `json:"executable"`
	Argv       []string          `json:"argv"`
	Cwd        string            `json:"cwd"`
	Stdin      string            `json:"stdin"`
	Probe      string            `json:"probe"`
	Env        map[string]string `json:"env"`
	LookPath   map[string]string `json:"lookpath"`
	// Write records one attempted file write per STUB_WRITE_TARGETS
	// entry: the path maps to "" on success and to the error otherwise.
	Write map[string]string `json:"write"`
	// Truncate records one os.Truncate(path, 0) attempt per
	// STUB_TRUNCATE_TARGETS entry; OTrunc records one open with
	// O_RDONLY|O_TRUNC over the same entries. The path maps to "" on
	// success and to the error otherwise.
	Truncate map[string]string `json:"truncate"`
	OTrunc   map[string]string `json:"otrunc"`
	// Unlink, Rmdir, Mkdir, and Mkfifo record one os.Remove, os.Mkdir,
	// or FIFO creation per STUB_UNLINK_TARGETS, STUB_RMDIR_TARGETS,
	// STUB_MKDIR_TARGETS, or STUB_MKFIFO_TARGETS entry. Rename records
	// one os.Rename per STUB_RENAME_SRC entry to the parallel
	// STUB_RENAME_DST entry, keyed by the source. Each maps to "" on
	// success and to the error otherwise.
	Unlink map[string]string `json:"unlink"`
	Rmdir  map[string]string `json:"rmdir"`
	Mkdir  map[string]string `json:"mkdir"`
	Mkfifo map[string]string `json:"mkfifo"`
	Rename map[string]string `json:"rename"`
	// ExecTry records the STUB_EXEC_TRY spawn attempt: "" when the
	// program ran, the error otherwise.
	ExecTry string `json:"exec_try"`
	// SpawnPid records the STUB_SPAWN_SLEEP descendant: its process
	// identifier when the spawn succeeded, 0 otherwise. SpawnErr
	// records the spawn failure, "" on success. The descendant
	// publishes its identity through this report — never through a
	// side file, which the write confinement may not grant.
	SpawnPid int    `json:"spawn_pid"`
	SpawnErr string `json:"spawn_err"`
}

func main() {
	os.Exit(run())
}

func run() int {
	if os.Getenv("STUB_SLEEP_CHILD") == "1" {
		seconds, _ := strconv.Atoi(os.Getenv("STUB_SPAWN_SECONDS"))
		if seconds <= 0 {
			seconds = 120
		}
		time.Sleep(time.Duration(seconds) * time.Second)
		return 0
	}
	if marker := os.Getenv("STUB_MARKER"); marker != "" {
		file, err := os.OpenFile(marker, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err == nil {
			_, _ = file.WriteString("ran\n")
			_ = file.Close()
		}
	}
	spawnPid := 0
	spawnErr := ""
	if os.Getenv("STUB_SPAWN_SLEEP") == "1" {
		child := exec.Command(os.Args[0], "sleep-child")
		child.Env = append(os.Environ(), "STUB_SLEEP_CHILD=1")
		if err := child.Start(); err != nil {
			spawnErr = err.Error()
		} else if child.Process != nil {
			spawnPid = child.Process.Pid
			_ = child.Process.Release()
		}
		// Deliberately not waited: the descendant must outlive this
		// interpreter so the test proves the worker-domain teardown reaps
		// it. A successful start is proof of life: the child sleeps far
		// longer than the session, so no wait is needed before the
		// interpreter exits.
	}
	stdin, _ := io.ReadAll(os.Stdin)
	cwd, _ := os.Getwd()
	environment := map[string]string{}
	for _, item := range os.Environ() {
		if key, value, ok := strings.Cut(item, "="); ok {
			environment[key] = value
		}
	}
	resolved := map[string]string{}
	// STUB_LOOKPATH names bare executables to resolve through the
	// process PATH, separated by the platform list separator, so tests
	// prove what bare-name resolution reaches inside the invocation.
	// STUB_LOOKPATH_MISSING names executables that must not resolve; a
	// resolution is recorded with a "resolved:" prefix so the test can
	// tell an unexpected hit from an expected miss.
	if names := os.Getenv("STUB_LOOKPATH"); names != "" {
		for _, name := range filepath.SplitList(names) {
			if name == "" {
				continue
			}
			if path, err := exec.LookPath(name); err == nil {
				resolved[name] = path
			} else {
				resolved[name] = "unresolved:" + err.Error()
			}
		}
	}
	if names := os.Getenv("STUB_LOOKPATH_MISSING"); names != "" {
		for _, name := range filepath.SplitList(names) {
			if name == "" {
				continue
			}
			if path, err := exec.LookPath(name); err == nil {
				resolved[name] = "resolved:" + path
			} else {
				resolved[name] = "missing"
			}
		}
	}
	// STUB_TRUNCATE_TARGETS names absolute existing files to truncate,
	// so Linux tests prove which truncations the write confinement
	// allows: one truncate(2) per entry plus one open with
	// O_RDONLY|O_TRUNC, which needs the truncate right without write
	// access. Truncations run before writes, so a file named in both
	// lists ends with the write payload.
	truncates := map[string]string{}
	otrunces := map[string]string{}
	if targets := os.Getenv("STUB_TRUNCATE_TARGETS"); targets != "" {
		for _, target := range filepath.SplitList(targets) {
			if target == "" {
				continue
			}
			if err := os.Truncate(target, 0); err == nil {
				truncates[target] = ""
			} else {
				truncates[target] = err.Error()
			}
			if file, err := os.OpenFile(target, os.O_RDONLY|os.O_TRUNC, 0); err == nil {
				otrunces[target] = ""
				_ = file.Close()
			} else {
				otrunces[target] = err.Error()
			}
		}
	}
	// STUB_WRITE_TARGETS names absolute files to create, separated by the
	// platform list separator, so Linux tests prove which writes the
	// Landlock write confinement allows. STUB_EXEC_TRY names one
	// absolute program to spawn, so Linux tests prove which descendant
	// executions the exec denial allows.
	writes := map[string]string{}
	if targets := os.Getenv("STUB_WRITE_TARGETS"); targets != "" {
		for _, target := range filepath.SplitList(targets) {
			if target == "" {
				continue
			}
			if err := os.WriteFile(target, []byte("stub-write\n"), 0o644); err == nil {
				writes[target] = ""
			} else {
				writes[target] = err.Error()
			}
		}
	}
	// STUB_UNLINK_TARGETS, STUB_RMDIR_TARGETS, STUB_MKDIR_TARGETS, and
	// STUB_MKFIFO_TARGETS name absolute paths to unlink, remove,
	// create, or FIFO-create, so Linux tests prove which directory
	// mutations the write confinement allows. STUB_RENAME_SRC and
	// STUB_RENAME_DST are parallel lists of rename sources and
	// destinations.
	unlinks := map[string]string{}
	for _, target := range filepath.SplitList(os.Getenv("STUB_UNLINK_TARGETS")) {
		if target == "" {
			continue
		}
		if err := os.Remove(target); err == nil {
			unlinks[target] = ""
		} else {
			unlinks[target] = err.Error()
		}
	}
	rmdirs := map[string]string{}
	for _, target := range filepath.SplitList(os.Getenv("STUB_RMDIR_TARGETS")) {
		if target == "" {
			continue
		}
		if err := os.Remove(target); err == nil {
			rmdirs[target] = ""
		} else {
			rmdirs[target] = err.Error()
		}
	}
	mkdirs := map[string]string{}
	for _, target := range filepath.SplitList(os.Getenv("STUB_MKDIR_TARGETS")) {
		if target == "" {
			continue
		}
		if err := os.Mkdir(target, 0o755); err == nil {
			mkdirs[target] = ""
		} else {
			mkdirs[target] = err.Error()
		}
	}
	mkfifos := map[string]string{}
	for _, target := range filepath.SplitList(os.Getenv("STUB_MKFIFO_TARGETS")) {
		if target == "" {
			continue
		}
		if err := makeFifo(target); err == nil {
			mkfifos[target] = ""
		} else {
			mkfifos[target] = err.Error()
		}
	}
	renames := map[string]string{}
	renameSources := filepath.SplitList(os.Getenv("STUB_RENAME_SRC"))
	renameDests := filepath.SplitList(os.Getenv("STUB_RENAME_DST"))
	for index, source := range renameSources {
		if source == "" {
			continue
		}
		if index >= len(renameDests) || renameDests[index] == "" {
			renames[source] = "missing parallel STUB_RENAME_DST entry"
			continue
		}
		if err := os.Rename(source, renameDests[index]); err == nil {
			renames[source] = ""
		} else {
			renames[source] = err.Error()
		}
	}
	execTry := ""
	if program := os.Getenv("STUB_EXEC_TRY"); program != "" {
		child := exec.Command(program)
		child.Stdout, child.Stderr = nil, nil
		if err := child.Run(); err != nil {
			execTry = err.Error()
		}
	}
	payload, _ := json.Marshal(report{
		Executable: os.Args[0],
		Argv:       append([]string(nil), os.Args...),
		Cwd:        cwd,
		Stdin:      string(stdin),
		Probe:      os.Getenv("STUB_INHERIT_PROBE"),
		Env:        environment,
		LookPath:   resolved,
		Write:      writes,
		Truncate:   truncates,
		OTrunc:     otrunces,
		Unlink:     unlinks,
		Rmdir:      rmdirs,
		Mkdir:      mkdirs,
		Mkfifo:     mkfifos,
		Rename:     renames,
		ExecTry:    execTry,
		SpawnPid:   spawnPid,
		SpawnErr:   spawnErr,
	})
	_, _ = os.Stdout.Write(append(payload, '\n'))
	if text := os.Getenv("STUB_STDERR"); text != "" {
		_, _ = os.Stderr.WriteString(text)
	}
	if raw := os.Getenv("STUB_EXIT"); raw != "" {
		if code, err := strconv.Atoi(raw); err == nil {
			return code
		}
		return 1
	}
	return 0
}
