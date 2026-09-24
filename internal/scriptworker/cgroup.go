package scriptworker

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Delegated cgroup v2 application for the Linux host-conditional controls
// `active-process-count-limit` (pids.max) and `aggregate-memory-limit`
// (memory.max).
//
// The file protocol is portable Go: the probe and the application read and
// write the cgroup filesystem only, so the same code runs against the real
// hierarchy on Linux and against a fixture hierarchy in tests. What the
// fixture cannot prove — kernel delegation on any particular host — the
// ubuntu-latest real-probe rows prove by branching on the true probe
// outcome.

const (
	// scriptCgroupPidsBound is the exact active-process bound one enforced
	// script invocation installs where the probe finds delegation. The
	// magnitude matches the go-v1 default.
	scriptCgroupPidsBound = 64
	// scriptCgroupMemoryBound is the exact aggregate memory bound in bytes
	// one enforced script invocation installs where the probe finds
	// delegation. The magnitude matches the go-v1 default.
	scriptCgroupMemoryBound = int64(2 * 1024 * 1024 * 1024)
)

// scriptCgroupFS names the cgroup filesystem roots one probe or
// application reads. Production always uses the real hierarchy;
// defaultScriptCgroupFS is that hierarchy.
type scriptCgroupFS struct {
	// Root is the cgroup v2 mount point.
	Root string
	// SelfPath names the file carrying this process's own cgroup
	// membership in /proc/self/cgroup format.
	SelfPath string
}

// defaultScriptCgroupFS is the production hierarchy.
func defaultScriptCgroupFS() scriptCgroupFS {
	return scriptCgroupFS{Root: "/sys/fs/cgroup", SelfPath: "/proc/self/cgroup"}
}

// testCgroupFS overrides the hierarchy when non-nil. Tests point it at a
// fixture directory to drive the delegation file protocol deterministically
// on any host; production leaves it nil. The override never crosses the
// process boundary directly: the parent resolves it and sends the effective
// roots to the worker in the request, so the worker confirms against the
// same hierarchy the parent applied.
var testCgroupFS *scriptCgroupFS

// OverrideCgroupFSForTest forces the cgroup hierarchy roots and returns a
// restore function. Production never sets it.
func OverrideCgroupFSForTest(filesystem scriptCgroupFS) (restore func()) {
	testCgroupFS = &filesystem
	return func() { testCgroupFS = nil }
}

func effectiveCgroupFS() scriptCgroupFS {
	if testCgroupFS != nil {
		return *testCgroupFS
	}
	return defaultScriptCgroupFS()
}

// scriptCgroupSelf reads this process's own cgroup v2 membership path,
// relative to the hierarchy root: the single `0::/path` line.
func scriptCgroupSelf(filesystem scriptCgroupFS) (string, error) {
	payload, err := os.ReadFile(filesystem.SelfPath) // #nosec G304 -- hierarchy root is fixed, membership is kernel-owned
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(payload), "\n") {
		rest, ok := strings.CutPrefix(strings.TrimSpace(line), "0::")
		if !ok {
			continue
		}
		relative := strings.TrimPrefix(rest, "/")
		if relative == "" {
			return "", fmt.Errorf("the process cgroup path is empty")
		}
		if !filepath.IsLocal(relative) {
			return "", fmt.Errorf("the process cgroup path escapes the hierarchy")
		}
		return filepath.Clean(relative), nil
	}
	return "", fmt.Errorf("no cgroup v2 membership line")
}

// probeScriptCgroup determines whether this host delegates a cgroup v2
// hierarchy the invocation can bound: it reads the controllers the parent
// cgroup offers, enables exactly the one under probe, creates a transient
// child, installs the exact bound this invocation would install, verifies
// the read-back, and removes the child again. Every expected-absence
// condition — no hierarchy, no v2 membership, no controller, no write
// permission — reports absent without an error; only the test-only fault
// injector reports an error, so a faulted probe refuses fail-closed.
func probeScriptCgroup(name string) (bool, error) {
	filesystem := effectiveCgroupFS()
	controller, bound, file := scriptCgroupSpec(name)
	if controller == "" {
		return false, diagnostic(CodeWorkerProtocolInvalid, "no cgroup mechanism for inventory control %q", name)
	}
	own, err := scriptCgroupSelf(filesystem)
	if err != nil {
		return false, nil
	}
	controllers, err := os.ReadFile(filepath.Join(filesystem.Root, own, "cgroup.controllers")) // #nosec G304 -- controller list below the validated hierarchy root
	if err != nil {
		return false, nil
	}
	if !cgroupControllerPresent(string(controllers), controller) {
		return false, nil
	}
	subtreePath := filepath.Join(filesystem.Root, own, "cgroup.subtree_control")
	original, err := os.ReadFile(subtreePath) // #nosec G304 -- controller line below the validated hierarchy root
	if err != nil {
		return false, nil
	}
	enabled := cgroupEnableLine(string(original), controller)
	if enabled != "" {
		if err := os.WriteFile(subtreePath, []byte(enabled), 0o644); err != nil {
			return false, nil
		}
		defer func() { _ = os.WriteFile(subtreePath, original, 0o644) }()
	}
	child, err := scriptCgroupTempChild(filesystem.Root, own)
	if err != nil {
		return false, nil
	}
	defer func() { _ = os.Remove(filepath.Join(filesystem.Root, child)) }()
	if err := os.WriteFile(filepath.Join(filesystem.Root, child, file), []byte(bound), 0o644); err != nil {
		return false, nil
	}
	readBack, err := os.ReadFile(filepath.Join(filesystem.Root, child, file)) // #nosec G304 -- read-back of the bound this invocation wrote
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(string(readBack)) == bound, nil
}

// scriptCgroupSpec maps one inventory control to its cgroup controller, the
// exact bound this invocation installs, and the limit file.
func scriptCgroupSpec(name string) (controller, bound, file string) {
	switch name {
	case ScriptControlActiveProcessCountLimit:
		return "pids", strconv.Itoa(scriptCgroupPidsBound), "pids.max"
	case ScriptControlAggregateMemoryLimit:
		return "memory", strconv.FormatInt(scriptCgroupMemoryBound, 10), "memory.max"
	default:
		return "", "", ""
	}
}

func cgroupControllerPresent(controllers, controller string) bool {
	for _, field := range strings.Fields(controllers) {
		if field == controller {
			return true
		}
	}
	return false
}

// cgroupEnableLine returns the subtree_control line that enables the
// controller, or "" when it is already enabled.
func cgroupEnableLine(current, controller string) string {
	for _, field := range strings.Fields(current) {
		if field == controller {
			return ""
		}
	}
	if current == "" {
		return "+" + controller
	}
	return "+" + controller + " " + strings.TrimSpace(current)
}

// scriptCgroupTempChild creates one transient child cgroup below the owner
// and returns its hierarchy-relative path.
func scriptCgroupTempChild(root, own string) (string, error) {
	var nonce [8]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return "", err
	}
	relative := filepath.Join(own, ".curator-script-probe-"+hex.EncodeToString(nonce[:]))
	if err := os.Mkdir(filepath.Join(root, relative), 0o755); err != nil {
		return "", err
	}
	return relative, nil
}

// scriptCgroupDomain is one prepared invocation child cgroup: the limits
// are installed and the worker is assigned to it before any session byte
// flows, so the interpreter it starts is born inside the bounded domain.
type scriptCgroupDomain struct {
	filesystem scriptCgroupFS
	relative   string
	restored   []byte
	subtree    string
}

// prepareScriptCgroup creates the invocation child cgroup below this
// process's own membership, enables exactly the installable controllers,
// and installs the exact bounds. It creates no process and refuses before
// the worker exists.
func prepareScriptCgroup(probes []ScriptControlProbe) (*scriptCgroupDomain, error) {
	filesystem := effectiveCgroupFS()
	var wanted []string
	for _, probe := range probes {
		if !scriptControlInstallable(probe) {
			continue
		}
		switch probe.Name {
		case ScriptControlActiveProcessCountLimit, ScriptControlAggregateMemoryLimit:
			wanted = append(wanted, probe.Name)
		}
	}
	if len(wanted) == 0 {
		return nil, nil
	}
	own, err := scriptCgroupSelf(filesystem)
	if err != nil {
		return nil, diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot read the process cgroup membership")
	}
	subtree := filepath.Join(filesystem.Root, own, "cgroup.subtree_control")
	original, err := os.ReadFile(subtree) // #nosec G304 -- controller line below the validated hierarchy root
	if err != nil {
		return nil, diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot read the delegated controllers")
	}
	line := string(original)
	for _, name := range wanted {
		controller, _, _ := scriptCgroupSpec(name)
		if addition := cgroupEnableLine(line, controller); addition != "" {
			line = addition
		}
	}
	if line != string(original) {
		if err := os.WriteFile(subtree, []byte(line), 0o644); err != nil {
			return nil, diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot enable the delegated controllers")
		}
	}
	var nonce [8]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot derive an invocation cgroup name")
	}
	relative := filepath.Join(own, "curator-script-"+strconv.Itoa(os.Getpid())+"-"+hex.EncodeToString(nonce[:]))
	if err := os.Mkdir(filepath.Join(rootOf(filesystem), relative), 0o755); err != nil {
		return nil, diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot create the invocation cgroup")
	}
	domain := &scriptCgroupDomain{filesystem: filesystem, relative: relative, restored: original, subtree: subtree}
	for _, name := range wanted {
		_, bound, file := scriptCgroupSpec(name)
		if err := os.WriteFile(filepath.Join(rootOf(filesystem), relative, file), []byte(bound), 0o644); err != nil {
			domain.close()
			return nil, diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot install the invocation %s bound", file)
		}
	}
	return domain, nil
}

func rootOf(filesystem scriptCgroupFS) string { return filesystem.Root }

// assign moves one started worker into the invocation cgroup and verifies
// the membership from the hierarchy itself. It runs after the fork and
// before any session byte flows, while the worker is still blocked reading
// the request, so no interpreter can exist outside the bounded domain.
func (domain *scriptCgroupDomain) assign(pid int) error {
	if domain == nil {
		return nil
	}
	procs := filepath.Join(domain.filesystem.Root, domain.relative, "cgroup.procs")
	if err := os.WriteFile(procs, []byte(strconv.Itoa(pid)), 0o644); err != nil {
		return diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot assign the worker to the invocation cgroup")
	}
	members, err := os.ReadFile(procs) // #nosec G304 -- membership of the cgroup this invocation created
	if err != nil {
		return diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot confirm the invocation cgroup membership")
	}
	for _, field := range strings.Fields(string(members)) {
		if field == strconv.Itoa(pid) {
			return nil
		}
	}
	return diagnostic(CodeWorkerProtocolInvalid, "the worker is not a member of the invocation cgroup")
}

// close removes the invocation child cgroup and restores the parent's
// controller line best-effort.
func (domain *scriptCgroupDomain) close() {
	if domain == nil {
		return
	}
	_ = os.Remove(filepath.Join(domain.filesystem.Root, domain.relative))
	_ = os.WriteFile(domain.subtree, domain.restored, 0o644)
}

// confirmScriptCgroup runs inside the worker. It proves the invocation
// cgroup the parent prepared names this worker: the membership file is
// kernel-maintained on the real hierarchy, so the read-back is kernel
// truth, and the same file protocol runs against a fixture hierarchy in
// tests.
func confirmScriptCgroup(filesystem scriptCgroupFS, relative string) error {
	members, err := os.ReadFile(filepath.Join(filesystem.Root, filepath.Clean(relative), "cgroup.procs")) // #nosec G304 -- membership of the parent-prepared invocation cgroup
	if err != nil {
		return diagnosticErr(CodeCapabilityEvidenceInvalid, err, "cannot read the invocation cgroup membership")
	}
	own := strconv.Itoa(os.Getpid())
	for _, field := range strings.Fields(string(members)) {
		if field == own {
			return nil
		}
	}
	return diagnostic(CodeCapabilityEvidenceInvalid, "the worker is not a member of the invocation cgroup")
}
