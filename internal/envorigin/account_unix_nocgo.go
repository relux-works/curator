//go:build (darwin || linux) && !cgo

package envorigin

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// boundedAccountOutput stops an oversized account record without buffering it.
type boundedAccountOutput struct{ bytes.Buffer }

func (b *boundedAccountOutput) Write(p []byte) (int, error) {
	if len(p) > 1<<20-b.Len() {
		return 0, fmt.Errorf("account record exceeds bound")
	}
	return b.Buffer.Write(p)
}

// LookupAccount uses the platform's absolute native account-database utility.
// The child has no inherited selectors, shell, HOME or loader environment. The
// numeric realUID/effective identity and returned UID are independently bound.
func LookupAccount() (Account, error) {
	realUID, effective, err := processAccountIdentity()
	if err != nil || realUID != effective {
		return Account{}, fmt.Errorf("realUID/effective UID mismatch: %v", err)
	}
	path, args, fields, home := "/usr/bin/getent", []string{"passwd", strconv.FormatUint(effective, 10)}, 7, 5
	if runtime.GOOS == "darwin" {
		path, args, fields, home = "/usr/bin/id", []string{"-P", strconv.FormatUint(effective, 10)}, 10, 8
	}
	// Both executables belong to the trusted OS installation, never PATH.
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0022 != 0 {
		return Account{}, fmt.Errorf("native account utility unavailable or writable")
	}
	if err := trustedAccountExecutable(info); err != nil {
		return Account{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, args...) // #nosec G204 -- fixed absolute OS utility, root-owned/nonwritable check above; no shell or PATH lookup.
	cmd.Env = []string{"LANG=C", "LC_ALL=C"}
	var out boundedAccountOutput
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return Account{}, fmt.Errorf("native account lookup failed: %w", err)
	}
	return parseAccountEntry(out.String(), realUID, effective, fields, home)
}

func parseAccountEntry(raw string, realUID, effective uint64, fields, home int) (Account, error) {
	raw = strings.TrimSuffix(raw, "\n")
	values := strings.Split(raw, ":")
	if len(values) != fields || strings.ContainsAny(raw, "\r\n\x00") || values[0] == "" {
		return Account{}, invalid("ambiguous native account record")
	}
	uid, err := strconv.ParseUint(values[2], 10, 64)
	if err != nil || values[2] != strconv.FormatUint(uid, 10) || uid != effective || realUID != effective || !cleanAbsolute(values[home]) {
		return Account{}, invalid("native account identity/home mismatch")
	}
	return Account{UID: uid, RealUID: realUID, EffectiveUID: effective, Home: values[home]}, nil
}
