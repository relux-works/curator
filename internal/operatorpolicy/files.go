package operatorpolicy

import (
	"io"
	"os"
	"path/filepath"
)

// readBounded anchors reads below an already checked private directory, rejects
// links/special files and detects replacement around the bounded open/read.
func readBounded(dir, name string, limit int64) ([]byte, error) {
	root, e := os.OpenRoot(dir)
	if e != nil {
		return nil, refusal("workspace_trust_ledger_unreadable")
	}
	defer func() { _ = root.Close() }()
	before, e := root.Lstat(name)
	if e != nil || !before.Mode().IsRegular() || before.Mode().Perm()&0400 == 0 || before.Size() > limit {
		return nil, refusal("workspace_trust_ledger_unreadable")
	}
	f, e := root.Open(name)
	if e != nil {
		return nil, refusal("workspace_trust_ledger_unreadable")
	}
	defer func() { _ = f.Close() }()
	opened, e := f.Stat()
	if e != nil || !sameFile(before, opened) {
		return nil, refusal("workspace_trust_ledger_unreadable")
	}
	b, e := io.ReadAll(io.LimitReader(f, limit+1))
	if e != nil || int64(len(b)) != before.Size() {
		return nil, refusal("workspace_trust_ledger_unreadable")
	}
	after, e := root.Lstat(name)
	if e != nil || !sameFile(before, after) {
		return nil, refusal("workspace_trust_ledger_unreadable")
	}
	opened, e = f.Stat()
	if e != nil || !sameFile(before, opened) {
		return nil, refusal("workspace_trust_ledger_unreadable")
	}
	return b, nil
}
func sameFile(a, b os.FileInfo) bool {
	return b != nil && a.Mode() == b.Mode() && os.SameFile(a, b) && a.Size() == b.Size() && a.ModTime().Equal(b.ModTime())
}

// ReadSignedCommand reads only a bounded regular command file. Its bytes are
// inert until both signatures authenticate against the independent key pin.
func ReadSignedCommand(p string) (SignedCommand, error) {
	b, e := readBounded(filepath.Dir(p), filepath.Base(p), 1<<20)
	if e != nil {
		return SignedCommand{}, refusal("workspace_trust_policy_authority")
	}
	return ParseSignedCommand(b)
}
