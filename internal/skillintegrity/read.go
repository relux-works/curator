package skillintegrity

import (
	"io"
	"os"
)

func readCapturedFile(path string, expected os.FileInfo, remaining int64) ([]byte, os.FileInfo, error) {
	f, err := openCaptureFile(path)
	if err != nil {
		if current, observeErr := os.Lstat(path); observeErr == nil && (!current.Mode().IsRegular() || !singleLink(path, current)) {
			return nil, nil, failure("skill_integrity_boundary_invalid", "file changed to a forbidden kind")
		}
		return nil, nil, failure("skill_integrity_unavailable", "skill file cannot be opened")
	}
	defer func() { _ = f.Close() }()
	before, err := f.Stat()
	if err != nil {
		return nil, nil, failure("skill_integrity_unavailable", "skill file cannot be observed")
	}
	if !before.Mode().IsRegular() || !singleLink(path, before) {
		return nil, nil, failure("skill_integrity_boundary_invalid", "file changed to a forbidden kind")
	}
	if !os.SameFile(expected, before) || expected.Size() != before.Size() || expected.Mode() != before.Mode() || !expected.ModTime().Equal(before.ModTime()) {
		return nil, nil, failure("skill_integrity_unavailable", "skill file changed during capture")
	}
	data, err := io.ReadAll(io.LimitReader(f, remaining+1))
	if err != nil {
		return nil, nil, failure("skill_integrity_unavailable", "skill file cannot be read")
	}
	if int64(len(data)) > remaining {
		return nil, nil, failure("skill_integrity_manifest_invalid", "capture byte budget exceeded")
	}
	after, err := f.Stat()
	if err != nil {
		return nil, nil, failure("skill_integrity_unavailable", "skill file cannot be rechecked")
	}
	current, err := os.Lstat(path)
	if err != nil {
		return nil, nil, failure("skill_integrity_unavailable", "skill file cannot be rechecked")
	}
	if !after.Mode().IsRegular() || !current.Mode().IsRegular() || !singleLink(path, after) || !singleLink(path, current) {
		return nil, nil, failure("skill_integrity_boundary_invalid", "file changed to a forbidden kind")
	}
	if !os.SameFile(after, current) || before.Mode() != after.Mode() || before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) || int64(len(data)) != after.Size() {
		return nil, nil, failure("skill_integrity_unavailable", "skill file changed during capture")
	}
	return data, after, nil
}
