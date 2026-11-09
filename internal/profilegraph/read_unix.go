//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package profilegraph

import (
	"errors"
	"golang.org/x/sys/unix"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Descriptor-relative no-follow opens hold every parent directory stable. Links
// are expanded explicitly, rechecked against the deny-list and depth budget,
// and restarted at the retained admitted root. No check/open gap follows links.
func (r *Reader) open(root, rel string, directory bool, links ...bool) (*os.File, error) {
	original, _ := filepath.Abs(root)
	allowedPrefix := r.boundaries[original].prefix
	root, rel, err := r.boundary(root, rel)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(root)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, refusal("graph_symlink_escape")
	}
	base, err := unix.Open(root, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		if errors.Is(err, unix.ELOOP) || (info.IsDir() && errors.Is(err, unix.ENOTDIR)) {
			return nil, refusal("graph_symlink_escape")
		}
		return nil, err
	}
	defer func() { _ = unix.Close(base) }()
	seen := map[string]bool{}
	for depth := 0; depth <= r.Limits.Depth; depth++ {
		if seen[rel] {
			return nil, refusal("graph_cycle")
		}
		seen[rel] = true
		fd, err := unix.Dup(base)
		if err != nil {
			return nil, err
		}
		parts := strings.Split(rel, "/")
		restart := false
		for i, part := range parts {
			if err := r.inspect(); err != nil {
				_ = unix.Close(fd)
				return nil, err
			}
			var st unix.Stat_t
			if err := unix.Fstatat(fd, part, &st, unix.AT_SYMLINK_NOFOLLOW); err != nil {
				_ = unix.Close(fd)
				return nil, err
			}
			if st.Mode&unix.S_IFMT == unix.S_IFLNK {
				if len(links) > 0 && !links[0] {
					_ = unix.Close(fd)
					return nil, refusal("graph_symlink_escape")
				}
				buf := make([]byte, 4097)
				n, err := unix.Readlinkat(fd, part, buf)
				_ = unix.Close(fd)
				if err != nil || n >= len(buf) {
					return nil, refusal("graph_invalid_input")
				}
				target := string(buf[:n])
				next := ""
				if path.IsAbs(target) {
					relative, err := filepath.Rel(root, target)
					if err != nil {
						return nil, refusal("graph_symlink_escape")
					}
					next = path.Clean(path.Join(filepath.ToSlash(relative), path.Join(parts[i+1:]...)))
				} else {
					next = path.Clean(path.Join(path.Join(parts[:i]...), target, path.Join(parts[i+1:]...)))
				}
				if next == ".." || strings.HasPrefix(next, "../") {
					return nil, refusal("graph_symlink_escape")
				}
				if allowedPrefix != "" && next != allowedPrefix && !strings.HasPrefix(next, allowedPrefix+"/") {
					return nil, refusal("graph_symlink_escape")
				}
				if denied(next) {
					return nil, refusal("graph_path_denied")
				}
				if next == rel {
					return nil, refusal("graph_cycle")
				}
				rel = next
				restart = true
				break
			}
			flags := unix.O_RDONLY | unix.O_CLOEXEC | unix.O_NOFOLLOW | unix.O_NONBLOCK
			if i < len(parts)-1 || directory {
				flags |= unix.O_DIRECTORY
			}
			next, err := unix.Openat(fd, part, flags, 0)
			_ = unix.Close(fd)
			if err != nil {
				if errors.Is(err, unix.ELOOP) {
					return nil, refusal("graph_symlink_escape")
				}
				return nil, err
			}
			fd = next
		}
		if !restart {
			return os.NewFile(uintptr(fd), rel), nil
		}
	}
	return nil, refusal("graph_limit_exceeded")
}
