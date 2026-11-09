package profilegraph

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Reader owns the input budget and ledger. All payload opens pass through it.
// OpenHook is an observation seam called only after path admission.
type Reader struct {
	Limits         Limits
	OpenHook       func(root, path string)
	total, entries int
	inputs         map[string]Input
	cache          map[string][]byte
	boundaries     map[string]rootBinding
}
type rootBinding struct{ root, prefix string }

// bind keeps derived cache paths inside the admitted managed root. Treating
// each cache leaf as a new trust root would miss a substituted parent link.
func (r *Reader) bind(root, boundary string) error {
	full, err := filepath.Abs(root)
	if err != nil {
		return refusal("graph_invalid_input")
	}
	base, err := filepath.Abs(boundary)
	if err != nil {
		return refusal("graph_invalid_input")
	}
	prefix, err := filepath.Rel(base, full)
	if err != nil || !cleanRelative(filepath.ToSlash(prefix)) {
		return refusal("graph_symlink_escape")
	}
	if r.boundaries == nil {
		r.boundaries = map[string]rootBinding{}
	}
	r.boundaries[full] = rootBinding{base, filepath.ToSlash(prefix)}
	return nil
}
func (r *Reader) boundary(root, rel string) (string, string, error) {
	full, err := filepath.Abs(root)
	if err != nil {
		return "", "", refusal("graph_invalid_input")
	}
	root = full
	if binding, ok := r.boundaries[full]; ok {
		root = binding.root
		rel = path.Join(binding.prefix, rel)
	}
	if denied(root) || denied(rel) {
		return "", "", refusal("graph_path_denied")
	}
	return root, rel, nil
}

// NewReader creates an isolated input ledger with explicit work budgets.
func NewReader(limits Limits) *Reader {
	return &Reader{Limits: limits, inputs: map[string]Input{}, cache: map[string][]byte{}}
}
func denied(p string) bool {
	for _, c := range strings.Split(strings.ToLower(filepath.ToSlash(p)), "/") {
		if c == ".ssh" || c == ".gnupg" || c == "logs" || c == ".env" || c == "shell-snapshots" || c == "shell_snapshots" || strings.HasSuffix(c, ".log") || strings.HasSuffix(c, ".db") || strings.Contains(c, ".sqlite") || strings.HasSuffix(c, "-wal") || strings.HasSuffix(c, "-shm") || strings.HasSuffix(c, "_history") || strings.HasSuffix(c, ".pem") || strings.HasSuffix(c, ".key") {
			return true
		}
		for _, prefix := range []string{"auth", "credential", "session", "history", "id_"} {
			if strings.HasPrefix(c, prefix) {
				return true
			}
		}
	}
	return false
}
func cleanRelative(p string) bool {
	return p != "" && !strings.ContainsAny(p, "\\\x00\r\n") && !path.IsAbs(p) && p == path.Clean(p) && p != ".." && !strings.HasPrefix(p, "../")
}
func (r *Reader) inspect() error {
	r.entries++
	if r.entries > r.Limits.Entries {
		return refusal("graph_limit_exceeded")
	}
	return nil
}
func readError(err error) error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	if errors.Is(err, fs.ErrNotExist) {
		return refusal("graph_missing_input")
	}
	return refusal("graph_invalid_input")
}
func (r *Reader) Read(root, label, rel string) ([]byte, error) {
	if !cleanRelative(rel) {
		return nil, refusal("graph_invalid_input")
	}
	if denied(root) || denied(rel) {
		return nil, refusal("graph_path_denied")
	}
	key := label + "\x00" + rel
	if b, ok := r.cache[key]; ok {
		return b, nil
	}
	f, err := r.open(root, rel, false)
	if err != nil {
		return nil, readError(err)
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return nil, refusal("graph_invalid_input")
	}
	if info.Size() > int64(r.Limits.File) || info.Size() > int64(r.Limits.Input-r.total) {
		return nil, refusal("graph_limit_exceeded")
	}
	if r.OpenHook != nil {
		r.OpenHook(label, rel)
	}
	raw, err := io.ReadAll(io.LimitReader(f, int64(min(r.Limits.File, r.Limits.Input-r.total))+1))
	if err != nil {
		return nil, readError(err)
	}
	if len(raw) > r.Limits.File || len(raw) > r.Limits.Input-r.total {
		return nil, refusal("graph_limit_exceeded")
	}
	after, err := f.Stat()
	if err != nil || info.Size() != int64(len(raw)) || after.Size() != info.Size() || after.ModTime() != info.ModTime() {
		return nil, refusal("graph_invalid_input")
	}
	r.total += len(raw)
	r.cache[key] = raw
	r.inputs[key] = Input{label, rel, hash(raw), len(raw)}
	return raw, nil
}
func (r *Reader) directory(root, rel string) ([]os.DirEntry, error) {
	if !cleanRelative(rel) {
		return nil, refusal("graph_invalid_input")
	}
	if denied(root) || denied(rel) {
		return nil, refusal("graph_path_denied")
	}
	f, err := r.open(root, rel, true)
	if err != nil {
		return nil, readError(err)
	}
	defer func() { _ = f.Close() }()
	entries, err := f.ReadDir(r.Limits.Entries - r.entries + 1)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, readError(err)
	}
	r.entries += len(entries)
	if r.entries > r.Limits.Entries {
		return nil, refusal("graph_limit_exceeded")
	}
	return entries, nil
}
