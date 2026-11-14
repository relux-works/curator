package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/startupverification"
)

func startupChanged() error {
	return &startupverification.Diagnostic{Code: "startup_verification_config_changed", Detail: "configuration changed during edit"}
}

// readStartupEdit opens an owned regular file without following a final link.
// It admits existing non-writable-to-others modes, preserving the source mode.
func readStartupEdit(path string) ([]byte, os.FileInfo, error) {
	entry, err := os.Lstat(path)
	if err != nil || !startupEditInfo(entry) {
		return nil, nil, startupverification.Invalid("configuration edit requires an owned regular file")
	}
	f, err := openClaudeAuthFile(path) // shared O_NOFOLLOW/O_NONBLOCK descriptor primitive
	if err != nil {
		return nil, nil, startupverification.Invalid("configuration cannot be opened safely")
	}
	defer func() { _ = f.Close() }()
	before, err := f.Stat()
	if err != nil || !startupEditInfo(before) || !os.SameFile(entry, before) {
		return nil, nil, startupChanged()
	}
	raw, err := io.ReadAll(io.LimitReader(f, before.Size()+1))
	if err != nil {
		return nil, nil, startupChanged()
	}
	after, err := f.Stat()
	current, statErr := os.Lstat(path)
	if err != nil || statErr != nil || !startupEditInfo(after) || !os.SameFile(before, current) || before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) || len(raw) != int(after.Size()) {
		return nil, nil, startupChanged()
	}
	return raw, after, nil
}

// SetStartupVerification changes only the machine level and its carrier.
// Publication is a single rename after source/policy readback. As with the
// existing config editor this is not an atomic CAS against hostile writers.
func SetStartupVerification(ctx context.Context, path string, level int) error {
	return setStartupVerification(ctx, path, level, nil)
}

func setStartupVerification(ctx context.Context, path string, level int, beforePublish func()) error {
	if _, err := startupverification.EffectiveLevel(level); err != nil {
		return err
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return startupverification.Invalid("configuration path is invalid")
	}
	system, err := filepath.Abs(SystemCandidatePath())
	if err != nil || path == system {
		return startupverification.Invalid("machine edit cannot target system policy")
	}
	raw, info, err := readStartupEdit(path)
	if err != nil {
		return err
	}
	if other, err := os.Stat(system); err == nil && os.SameFile(info, other) {
		return startupverification.Invalid("machine edit cannot target system policy")
	}
	if protocoljson.Validate(raw) != nil || validateStartupJSON(raw) != nil {
		return startupverification.Invalid("malformed source configuration")
	}
	var object map[string]any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if decoder.Decode(&object) != nil || object == nil {
		return startupverification.Invalid("source configuration is not an object")
	}
	cfg, err := loadObject(path, object, nil)
	if err != nil {
		return err
	}
	if cfg.Locked["startup_verification"] {
		return startupverification.Invalid("startup_verification is locked by system policy")
	}
	policy, err := claudeAuthPolicy() // binds policy absence or complete file identity
	if err != nil {
		return startupverification.Invalid("system policy cannot be observed")
	}
	// Schema1's implicit permissive posture must not silently become the
	// modern carrier's hardened default merely because the level is edited.
	if version, _ := integerValue(object["schema_version"]); version == SchemaVersion {
		object["security_posture"] = cfg.SecurityPosture
	}
	object["schema_version"] = StartupVerificationSchemaVersion
	object["startup_verification"] = map[string]any{"level": level}
	if _, err := loadObject(path, object, nil); err != nil {
		return err
	}
	candidate, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		return startupverification.Invalid("candidate cannot be encoded")
	}
	candidate = append(candidate, '\n')
	f, err := os.CreateTemp(filepath.Dir(path), ".startup-verification-*.tmp")
	if err != nil {
		return fmt.Errorf("startup_verification_invalid: cannot stage configuration")
	}
	tmp := f.Name()
	defer func() { _ = f.Close(); _ = os.Remove(tmp) }()
	if err := f.Chmod(info.Mode().Perm()); err != nil {
		return err
	}
	if _, err := f.Write(candidate); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if beforePublish != nil {
		beforePublish()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	observed, observedInfo, err := readStartupEdit(path)
	if err != nil || !os.SameFile(info, observedInfo) || info.Mode() != observedInfo.Mode() || !bytes.Equal(raw, observed) {
		return startupChanged()
	}
	nowPolicy, err := claudeAuthPolicy()
	if err != nil || policy != nowPolicy {
		return startupChanged()
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("startup_verification_invalid: configuration publication failed")
	}
	// Publication has occurred. A sync/readback failure is reported as applied
	// but unverified; never remove the destination to retry a failed rename.
	if err := syncClaudeAuthDirectory(path); err != nil {
		return fmt.Errorf("startup_verification_unavailable: edit applied; durability unverified")
	}
	readback, _, err := readStartupEdit(path)
	if err != nil || !bytes.Equal(readback, candidate) {
		return fmt.Errorf("startup_verification_unavailable: edit applied; readback unverified")
	}
	return nil
}
