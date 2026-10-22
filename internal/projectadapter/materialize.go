package projectadapter

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/privatedir"
)

// Materialize creates a fresh manager allocation atomically, or verifies an
// identical existing one without replacing its baseline. The caller must hold
// the manager lock and enforce I-LEASE's idle-home rule before calling; this
// adapter never owns, changes or releases a lease. It launches nothing.
func (p *Plan) Materialize(ctx context.Context, held managerlock.Held, managerHome string) error {
	if err := held.Validate(managerHome); err != nil {
		return err
	}
	rel, err := filepath.Rel(managerHome, p.home)
	if err != nil {
		return err
	}
	parts := strings.Split(filepath.ToSlash(rel), "/")
	if len(parts) != 5 || parts[0] != "environments" || parts[1] != "project-views" || !identifiers.Valid(parts[2]) || !identifiers.Valid(parts[3]) || parts[4] != "codex_cli" {
		return fmt.Errorf("project_context_invalid: home is not a manager project-view allocation")
	}
	if inside(managerHome, p.worktree) || inside(p.worktree, p.home) {
		return fmt.Errorf("project_context_invalid: worktree overlaps manager authority")
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	if err = pathboundary.ValidateRoot(managerHome); err != nil {
		return err
	}
	root, err := os.OpenRoot(managerHome)
	if err != nil {
		return err
	}
	defer root.Close()
	current := ""
	for _, part := range parts[:4] {
		current = filepath.Join(current, part)
		if err = root.Mkdir(current, 0700); err != nil && !os.IsExist(err) {
			return err
		}
		if err = pathboundary.ValidateRouteWithOwner(managerHome, filepath.Join(managerHome, current), pathboundary.DefaultOwnerLookup()); err != nil {
			return err
		}
	}
	if _, err = root.Lstat(rel); err == nil {
		return p.Verify(ctx)
	} else if !os.IsNotExist(err) {
		return err
	}
	parent := filepath.Dir(p.home)
	stage, err := os.MkdirTemp(parent, ".codex-composition-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	for name, b := range p.files {
		if err = ctx.Err(); err != nil {
			return err
		}
		target := filepath.Join(stage, filepath.FromSlash(name))
		if err = privatedir.MakeAll(filepath.Dir(target)); err != nil {
			return err
		}
		if err = os.WriteFile(target, b, os.FileMode(p.generated[name].Mode)); err != nil {
			return err
		}
	}
	if err = held.Validate(managerHome); err != nil {
		return err
	}
	if err = pathboundary.ValidateWithin(managerHome, stage); err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	// A manager lock serializes publication; no existing allocation is replaced.
	if _, err = root.Lstat(rel); err == nil {
		return fmt.Errorf("project_context_drift: home appeared before publication")
	} else if !os.IsNotExist(err) {
		return err
	}
	if err = root.Rename(filepath.Join(filepath.Dir(rel), filepath.Base(stage)), rel); err != nil {
		return err
	}
	return p.Verify(ctx)
}
func inside(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// Verify compares bytes, size, mode and generated skill inventory to the
// immutable receipt. Native mutable state and credentials are never read.
// Failure never repairs, rewrites or recaptures a baseline.
func (p *Plan) Verify(ctx context.Context) error {
	if err := privatedir.Validate(p.home); err != nil {
		return fmt.Errorf("project_context_drift: native home boundary: %w", err)
	}
	root, err := os.OpenRoot(p.home)
	if err != nil {
		return err
	}
	defer root.Close()
	for name, expected := range p.generated {
		if err = ctx.Err(); err != nil {
			return err
		}
		if err = pathboundary.ValidateRouteWithOwner(p.home, filepath.Join(p.home, filepath.FromSlash(name)), pathboundary.DefaultOwnerLookup()); err != nil {
			return fmt.Errorf("project_context_drift: generated path boundary: %w", err)
		}
		st, e := root.Lstat(name)
		if e != nil || !st.Mode().IsRegular() || uint32(st.Mode().Perm()) != expected.Mode || st.Size() != expected.Size {
			return fmt.Errorf("project_context_drift: generated metadata differs")
		}
		f, e := openGenerated(root, name)
		if e != nil {
			return e
		}
		opened, e := f.Stat()
		if e != nil || !os.SameFile(st, opened) || st.Mode() != opened.Mode() || st.Size() != opened.Size() {
			_ = f.Close()
			return fmt.Errorf("project_context_drift: generated file changed before read")
		}
		b, e := io.ReadAll(io.LimitReader(f, expected.Size+1))
		after, afterErr := f.Stat()
		_ = f.Close()
		if afterErr != nil || !os.SameFile(st, after) || after.Size() != st.Size() || after.Mode() != st.Mode() || !after.ModTime().Equal(st.ModTime()) {
			return fmt.Errorf("project_context_drift: generated file changed during read")
		}
		live, liveErr := root.Lstat(name)
		if liveErr != nil || !os.SameFile(st, live) || live.Mode() != st.Mode() || live.Size() != st.Size() || !live.ModTime().Equal(st.ModTime()) {
			return fmt.Errorf("project_context_drift: generated path changed during read")
		}
		if e != nil || surface(b, expected.Mode) != expected {
			return fmt.Errorf("project_context_drift: generated bytes differ")
		}
	}
	for _, name := range []string{"AGENTS.override.md"} {
		if _, e := root.Lstat(name); e == nil || !os.IsNotExist(e) {
			return fmt.Errorf("project_context_drift: unexpected native instruction override")
		}
	}
	for _, dir := range []string{"skills", ".agent-context"} {
		err = filepath.WalkDir(filepath.Join(p.home, dir), func(full string, entry os.DirEntry, e error) error {
			if e != nil {
				if full == filepath.Join(p.home, dir) && os.IsNotExist(e) {
					return nil
				}
				return e
			}
			if entry.IsDir() {
				return nil
			}
			rel, e := filepath.Rel(p.home, full)
			if e != nil {
				return e
			}
			if _, ok := p.generated[filepath.ToSlash(rel)]; !ok {
				return fmt.Errorf("project_context_drift: unreceipted generated file")
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return ctx.Err()
}
