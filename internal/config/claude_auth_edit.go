package config

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/relux-works/curator/internal/protocoljson"
)

// Fixed refusals never reproduce rejected configuration values or store errors.
var (
	ErrClaudeAuthEditRefused = errors.New("claude_auth_config_refused: invalid or unsafe metadata/source")
	ErrClaudeAuthStale       = errors.New("claude_auth_config_stale: source, candidate or system policy changed")
	ErrClaudeAuthEditIO      = errors.New("claude_auth_config_io: inspect the reported phase and retained backup")
)

const claudeAuthBackupTime = "20060102T150405.000000000Z"

// ClaudeAuthEditPlan contains only permitted metadata. Candidate bytes are
// deliberately excluded from JSON; private diff custody belongs to the caller.
type ClaudeAuthEditPlan struct {
	ConfigPath            string             `json:"config_path"`
	SourceSHA256          string             `json:"source_sha256"`
	CandidateSHA256       string             `json:"candidate_sha256"`
	PolicySHA256          string             `json:"policy_sha256"`
	SchemaFrom            int                `json:"schema_from"`
	SchemaTo              int                `json:"schema_to"`
	ClaudeAuth            LauncherClaudeAuth `json:"claude_auth"`
	PreservedTopLevelKeys []string           `json:"preserved_top_level_keys"`
	candidate             []byte
	source                *claudeAuthFile
}

// CandidateBytes returns a detached copy, never the source or credential bytes.
func (p *ClaudeAuthEditPlan) CandidateBytes() []byte { return bytes.Clone(p.candidate) }

// ClaudeAuthEditRequest binds a freshly recomputed candidate and system policy
// to the exact metadata the caller privately inspected in plan.
type ClaudeAuthEditRequest struct {
	Path                                                       string
	Selection                                                  LauncherClaudeAuth
	ExpectSourceSHA256, ExpectResultSHA256, ExpectPolicySHA256 string
}

// ClaudeAuthRestoreRequest binds a validated private backup and current file.
// An optional expected backup digest strengthens caller-held provenance; the
// captured backup bytes and identity are always rechecked before replacement.
type ClaudeAuthRestoreRequest struct {
	Path, BackupPath, ExpectCurrentSHA256, ExpectBackupSHA256 string
}

// ClaudeAuthEditResult distinguishes committed replacement from a retained
// recovery copy. A non-nil error never changes an applied status to refusal.
type ClaudeAuthEditResult struct {
	Status           string `json:"status"`
	ConfigPath       string `json:"config_path"`
	SourceSHA256     string `json:"source_sha256"`
	ResultSHA256     string `json:"result_sha256"`
	PolicySHA256     string `json:"policy_sha256"`
	BackupPath       string `json:"backup_path,omitempty"`
	BackupSHA256     string `json:"backup_sha256,omitempty"`
	Durable          bool   `json:"durable"`
	ReadbackVerified bool   `json:"readback_verified"`
}

type claudeAuthFile struct {
	path string
	raw  []byte
	info os.FileInfo
	sha  string
}

// Test seams inject real file interleavings/failures, not alternate acceptance
// predicates. Production always uses the same final recheck and atomic rename.
type claudeAuthEditOps struct {
	beforeRecheck, afterBackup, beforeFinalRecheck func() error
	rename                                         func(string, string) error
	syncDirectory                                  func(string) error
	syncBackupDirectory                            func(string) error
}

func claudeAuthSHA(raw []byte) string { sum := sha256.Sum256(raw); return hex.EncodeToString(sum[:]) }
func claudeAuthDigest(v string) bool {
	b, err := hex.DecodeString(v)
	return err == nil && len(b) == 32 && strings.ToLower(v) == v
}

func readClaudeAuthFile(path string) (*claudeAuthFile, error) {
	entry, err := os.Lstat(path)
	if err != nil || !privateClaudeAuthInfo(entry) {
		return nil, ErrClaudeAuthEditRefused
	}
	f, err := openClaudeAuthFile(path)
	if err != nil {
		return nil, ErrClaudeAuthEditRefused
	}
	defer func() { _ = f.Close() }()
	before, err := f.Stat()
	if err != nil || !privateClaudeAuthInfo(before) || !os.SameFile(entry, before) {
		return nil, ErrClaudeAuthEditRefused
	}
	raw, err := io.ReadAll(io.LimitReader(f, before.Size()+1))
	if err != nil {
		return nil, ErrClaudeAuthEditRefused
	}
	after, err := f.Stat()
	if err != nil || !privateClaudeAuthInfo(after) || claudeAuthFileIdentity(before) != claudeAuthFileIdentity(after) || int64(len(raw)) != after.Size() {
		return nil, ErrClaudeAuthEditRefused
	}
	current, err := os.Lstat(path)
	if err != nil || !privateClaudeAuthInfo(current) || !os.SameFile(current, after) {
		return nil, ErrClaudeAuthEditRefused
	}
	return &claudeAuthFile{path: path, raw: raw, info: after, sha: claudeAuthSHA(raw)}, nil
}

func claudeAuthObject(raw []byte) (map[string]any, error) {
	if protocoljson.Validate(raw) != nil {
		return nil, ErrClaudeAuthEditRefused
	}
	object, err := decodeConfigObject(raw, "")
	if err != nil {
		return nil, ErrClaudeAuthEditRefused
	}
	return object, nil
}

func claudeAuthPolicy() (string, error) {
	path, err := filepath.Abs(SystemCandidatePath())
	if err != nil {
		return "", ErrClaudeAuthEditRefused
	}
	return claudeAuthPolicyAt(path)
}

func claudeAuthPolicyAt(path string) (string, error) {
	entry, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return claudeAuthSHA([]byte("absent\x00" + path)), nil
	}
	if err != nil {
		return "", ErrClaudeAuthEditRefused
	}
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return "", ErrClaudeAuthEditRefused
	}
	raw, err := os.ReadFile(path) //nolint:gosec // G304: independently selected system-policy metadata, never the OAuth credential reference.
	if err != nil {
		return "", ErrClaudeAuthEditRefused
	}
	after, err := os.Stat(path)
	if err != nil || claudeAuthFileIdentity(info) != claudeAuthFileIdentity(after) {
		return "", ErrClaudeAuthStale
	}
	lastEntry, err := os.Lstat(path)
	if err != nil || claudeAuthFileIdentity(entry) != claudeAuthFileIdentity(lastEntry) {
		return "", ErrClaudeAuthStale
	}
	return claudeAuthSHA([]byte("present\x00" + path + "\x00" + claudeAuthFileIdentity(entry) + "\x00" + claudeAuthFileIdentity(info) + "\x00" + claudeAuthSHA(raw))), nil
}

func claudeAuthMachineSource(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", ErrClaudeAuthEditRefused
	}
	system, err := filepath.Abs(SystemCandidatePath())
	if err != nil || path == system {
		return "", ErrClaudeAuthEditRefused
	}
	if a, err := os.Stat(path); err == nil {
		if b, err := os.Stat(system); err == nil && os.SameFile(a, b) {
			return "", ErrClaudeAuthEditRefused
		}
	}
	return path, nil
}

// Both sides run the ordinary loader's exact overlay/parser path. Comparing
// complete parsed configs covers settings omitted by the narrow EffectiveJSON.
func claudeAuthEquivalent(path string, before, after map[string]any) bool {
	// Raw non-auth extensions must remain equal even if an enforced overlay
	// would hide the difference. Absent and empty extension sets are equivalent.
	remaining := func(object map[string]any) map[string]any {
		result := map[string]any{}
		if extensions, ok := object["extensions"].(map[string]any); ok {
			for key, value := range extensions {
				if key != "claude_auth" {
					result[key] = value
				}
			}
		}
		return result
	}
	if !reflect.DeepEqual(remaining(before), remaining(after)) {
		return false
	}
	a, err := loadObject(path, before, nil)
	if err != nil {
		return false
	}
	b, err := loadObject(path, after, nil)
	if err != nil {
		return false
	}
	a.Schema, b.Schema = 0, 0
	a.Extensions, b.Extensions = nil, nil
	if len(a.Env.CredentialStore) == 0 {
		a.Env.CredentialStore = nil
	}
	if len(b.Env.CredentialStore) == 0 {
		b.Env.CredentialStore = nil
	}
	return reflect.DeepEqual(a, b)
}

// PlanClaudeAuth validates references and the complete selected source before
// producing candidate bytes. It never opens the selected credential store.
func PlanClaudeAuth(path string, selection LauncherClaudeAuth) (*ClaudeAuthEditPlan, error) {
	path, err := claudeAuthMachineSource(path)
	if err != nil {
		return nil, err
	}
	source, err := readClaudeAuthFile(path)
	if err != nil {
		return nil, err
	}
	policy, err := claudeAuthPolicy()
	if err != nil {
		return nil, err
	}
	before, err := claudeAuthObject(source.raw)
	if err != nil {
		return nil, err
	}
	version, ok := integerValue(before["schema_version"])
	if !ok || (version != 2 && version != SchemaVersion7 && version != SchemaVersion8) {
		return nil, ErrClaudeAuthEditRefused
	}
	if _, err := loadObject(path, before, nil); err != nil {
		return nil, ErrClaudeAuthEditRefused
	}
	selectedBytes, err := json.Marshal(selection)
	if err != nil {
		return nil, ErrClaudeAuthEditRefused
	}
	var selected map[string]any
	if json.Unmarshal(selectedBytes, &selected) != nil {
		return nil, ErrClaudeAuthEditRefused
	}
	if _, err := parseLauncherExtensions(map[string]any{"claude_auth": selected}, SchemaVersion7); err != nil {
		return nil, ErrClaudeAuthEditRefused
	}
	after, err := claudeAuthObject(source.raw)
	if err != nil {
		return nil, err
	}
	extensions, _ := after["extensions"].(map[string]any)
	if extensions == nil {
		extensions = map[string]any{}
	}
	extensions["claude_auth"] = selected
	after["extensions"] = extensions
	targetSchema := version
	if targetSchema == 2 {
		targetSchema = SchemaVersion7
	}
	after["schema_version"] = float64(targetSchema)
	// Defense in depth: for currently valid schema-2/7/8 sources this constructor
	// changes only schema/auth, so no valid input currently reaches the refusal.
	// Restore independently needs this check for later unrelated edits.
	if !claudeAuthEquivalent(path, before, after) {
		return nil, ErrClaudeAuthEditRefused
	}
	raw, err := json.MarshalIndent(after, "", "  ")
	if err != nil {
		return nil, ErrClaudeAuthEditRefused
	}
	raw = append(raw, '\n')
	currentPolicy, err := claudeAuthPolicy()
	if err != nil || currentPolicy != policy {
		return nil, ErrClaudeAuthStale
	}
	keys := []string{}
	for key := range before {
		if key != "schema_version" && key != "extensions" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return &ClaudeAuthEditPlan{ConfigPath: path, SourceSHA256: source.sha, CandidateSHA256: claudeAuthSHA(raw), PolicySHA256: policy, SchemaFrom: version, SchemaTo: targetSchema, ClaudeAuth: selection, PreservedTopLevelKeys: keys, candidate: raw, source: source}, nil
}

func syncClaudeAuthDirectory(path string) error {
	f, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return f.Sync()
}

func writeClaudeAuthExclusive(path string, raw []byte, syncDirectory func(string) error) (published bool, err error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600) //nolint:gosec // G304: explicitly selected output uses O_EXCL/0600; an existing path cannot be overwritten.
	if err != nil {
		return false, ErrClaudeAuthEditIO
	}
	// Only an incomplete file created by this call is removed. A verified,
	// durable recovery backup is never removed on later transaction failure.
	info, statErr := f.Stat()
	complete := false
	defer func() {
		_ = f.Close()
		if !complete && statErr == nil {
			if now, e := os.Lstat(path); e == nil && os.SameFile(info, now) {
				_ = os.Remove(path)
			}
		}
	}()
	if statErr != nil {
		return false, ErrClaudeAuthEditIO
	}
	if _, err = f.Write(raw); err != nil {
		return false, ErrClaudeAuthEditIO
	}
	if err = f.Sync(); err != nil {
		return false, ErrClaudeAuthEditIO
	}
	if err = f.Close(); err != nil {
		return false, ErrClaudeAuthEditIO
	}
	verified, err := readClaudeAuthFile(path)
	if err != nil || !bytes.Equal(verified.raw, raw) {
		return false, ErrClaudeAuthEditIO
	}
	complete = true
	if syncDirectory == nil {
		syncDirectory = syncClaudeAuthDirectory
	}
	if syncDirectory(path) != nil {
		return true, ErrClaudeAuthEditIO
	}
	return true, nil
}

// WriteClaudeAuthCandidate creates only the explicitly named private output.
func WriteClaudeAuthCandidate(plan *ClaudeAuthEditPlan, path string) error {
	if plan == nil || claudeAuthSHA(plan.candidate) != plan.CandidateSHA256 {
		return ErrClaudeAuthEditRefused
	}
	_, err := writeClaudeAuthExclusive(path, plan.candidate, nil)
	return err
}

func recheckClaudeAuth(source *claudeAuthFile, policy string) error {
	current, err := readClaudeAuthFile(source.path)
	if err != nil || current.sha != source.sha || !os.SameFile(current.info, source.info) {
		return ErrClaudeAuthStale
	}
	actual, err := claudeAuthPolicy()
	if err != nil || actual != policy {
		return ErrClaudeAuthStale
	}
	return nil
}

// ApplyClaudeAuth uses observed hashes and a final recheck, not filesystem CAS.
// Actual adoption requires exclusive custody against noncooperating writers.
func ApplyClaudeAuth(ctx context.Context, req ClaudeAuthEditRequest) (*ClaudeAuthEditResult, error) {
	return applyClaudeAuth(ctx, req, claudeAuthEditOps{})
}

func applyClaudeAuth(ctx context.Context, req ClaudeAuthEditRequest, ops claudeAuthEditOps) (*ClaudeAuthEditResult, error) {
	if !claudeAuthDigest(req.ExpectSourceSHA256) || !claudeAuthDigest(req.ExpectResultSHA256) || !claudeAuthDigest(req.ExpectPolicySHA256) {
		return nil, ErrClaudeAuthEditRefused
	}
	policy, err := claudeAuthPolicy()
	if err != nil || policy != req.ExpectPolicySHA256 {
		return nil, ErrClaudeAuthStale
	}
	plan, err := PlanClaudeAuth(req.Path, req.Selection)
	if err != nil {
		return nil, err
	}
	if plan.SourceSHA256 != req.ExpectSourceSHA256 || plan.CandidateSHA256 != req.ExpectResultSHA256 || plan.PolicySHA256 != req.ExpectPolicySHA256 {
		return nil, ErrClaudeAuthStale
	}
	if ops.beforeRecheck != nil {
		if err := ops.beforeRecheck(); err != nil {
			return nil, ErrClaudeAuthEditIO
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := recheckClaudeAuth(plan.source, plan.PolicySHA256); err != nil {
		return nil, err
	}
	result := &ClaudeAuthEditResult{Status: "not_applied_backup_retained", ConfigPath: plan.ConfigPath, SourceSHA256: plan.SourceSHA256, ResultSHA256: plan.CandidateSHA256, PolicySHA256: plan.PolicySHA256, BackupPath: plan.ConfigPath + ".bak-" + time.Now().UTC().Format(claudeAuthBackupTime), BackupSHA256: plan.SourceSHA256}
	if published, err := writeClaudeAuthExclusive(result.BackupPath, plan.source.raw, ops.syncBackupDirectory); err != nil {
		// Preserve the observed complete publication even if a later re-read
		// is unavailable. This reports its path, not fresh verification/durability.
		if published {
			return result, err
		}
		return nil, err
	}
	if ops.afterBackup != nil {
		if err := ops.afterBackup(); err != nil {
			return result, ErrClaudeAuthEditIO
		}
	}
	backup, err := readClaudeAuthFile(result.BackupPath)
	if err != nil || backup.sha != plan.SourceSHA256 {
		return result, ErrClaudeAuthEditIO
	}
	return publishClaudeAuth(ctx, plan.source, backup, plan.PolicySHA256, plan.candidate, result, ops, "applied")
}

func publishClaudeAuth(ctx context.Context, source, backup *claudeAuthFile, policy string, raw []byte, result *ClaudeAuthEditResult, ops claudeAuthEditOps, status string) (*ClaudeAuthEditResult, error) {
	if err := ctx.Err(); err != nil {
		return result, err
	}
	f, err := os.CreateTemp(filepath.Dir(source.path), ".claude-auth-*.tmp")
	if err != nil {
		return result, ErrClaudeAuthEditIO
	}
	tmp := f.Name()
	info, statErr := f.Stat()
	defer func() {
		_ = f.Close()
		if statErr == nil {
			if current, e := os.Lstat(tmp); e == nil && os.SameFile(current, info) {
				_ = os.Remove(tmp)
			}
		}
	}()
	if statErr != nil {
		return result, ErrClaudeAuthEditIO
	}
	if _, err := f.Write(raw); err != nil {
		return result, ErrClaudeAuthEditIO
	}
	if f.Sync() != nil {
		return result, ErrClaudeAuthEditIO
	}
	if f.Close() != nil {
		return result, ErrClaudeAuthEditIO
	}
	if ops.beforeFinalRecheck != nil {
		if ops.beforeFinalRecheck() != nil {
			return result, ErrClaudeAuthEditIO
		}
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if err := recheckClaudeAuth(source, policy); err != nil {
		return result, err
	}
	verified, err := readClaudeAuthFile(backup.path)
	if err != nil || verified.sha != backup.sha || !os.SameFile(verified.info, backup.info) {
		return result, ErrClaudeAuthStale
	}
	rename := ops.rename
	if rename == nil {
		rename = os.Rename
	}
	if rename(tmp, source.path) != nil {
		return result, ErrClaudeAuthEditIO
	}
	// This is the commit point. No subsequent failure may claim the source
	// is unchanged or remove the retained recovery copy.
	result.Status = status
	readback, readErr := readClaudeAuthFile(source.path)
	result.ReadbackVerified = readErr == nil && readback.sha == claudeAuthSHA(raw)
	syncDir := ops.syncDirectory
	if syncDir == nil {
		syncDir = syncClaudeAuthDirectory
	}
	syncErr := syncDir(source.path)
	result.Durable = syncErr == nil
	if readErr != nil || !result.ReadbackVerified || syncErr != nil {
		return result, ErrClaudeAuthEditIO
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	return result, nil
}

// RestoreClaudeAuth never consumes a recovery copy or silently overwrites a
// changed current file. It has the same documented final-recheck race limit.
func RestoreClaudeAuth(ctx context.Context, req ClaudeAuthRestoreRequest) (*ClaudeAuthEditResult, error) {
	return restoreClaudeAuth(ctx, req, claudeAuthEditOps{})
}

func restoreClaudeAuth(ctx context.Context, req ClaudeAuthRestoreRequest, ops claudeAuthEditOps) (*ClaudeAuthEditResult, error) {
	if !claudeAuthDigest(req.ExpectCurrentSHA256) || (req.ExpectBackupSHA256 != "" && !claudeAuthDigest(req.ExpectBackupSHA256)) {
		return nil, ErrClaudeAuthEditRefused
	}
	path, err := claudeAuthMachineSource(req.Path)
	if err != nil {
		return nil, err
	}
	backup, err := filepath.Abs(req.BackupPath)
	if err != nil {
		return nil, ErrClaudeAuthEditRefused
	}
	prefix := filepath.Base(path) + ".bak-"
	if filepath.Dir(backup) != filepath.Dir(path) || !strings.HasPrefix(filepath.Base(backup), prefix) {
		return nil, ErrClaudeAuthEditRefused
	}
	if _, err := time.Parse(claudeAuthBackupTime, strings.TrimPrefix(filepath.Base(backup), prefix)); err != nil {
		return nil, ErrClaudeAuthEditRefused
	}
	current, err := readClaudeAuthFile(path)
	if err != nil {
		return nil, err
	}
	saved, err := readClaudeAuthFile(backup)
	if err != nil {
		return nil, err
	}
	if current.sha != req.ExpectCurrentSHA256 || (req.ExpectBackupSHA256 != "" && saved.sha != req.ExpectBackupSHA256) {
		return nil, ErrClaudeAuthStale
	}
	policy, err := claudeAuthPolicy()
	if err != nil {
		return nil, err
	}
	before, err := claudeAuthObject(current.raw)
	if err != nil {
		return nil, err
	}
	after, err := claudeAuthObject(saved.raw)
	if err != nil {
		return nil, err
	}
	for _, obj := range []map[string]any{before, after} {
		v, ok := integerValue(obj["schema_version"])
		if !ok || (v != 2 && v != SchemaVersion7 && v != SchemaVersion8) {
			return nil, ErrClaudeAuthEditRefused
		}
	}
	currentSchema, _ := integerValue(before["schema_version"])
	savedSchema, _ := integerValue(after["schema_version"])
	if currentSchema >= SchemaVersion8 && savedSchema != currentSchema {
		return nil, ErrClaudeAuthEditRefused
	}
	if !claudeAuthEquivalent(path, before, after) {
		return nil, ErrClaudeAuthEditRefused
	}
	if ops.beforeRecheck != nil {
		if ops.beforeRecheck() != nil {
			return nil, ErrClaudeAuthEditIO
		}
	}
	if err := recheckClaudeAuth(current, policy); err != nil {
		return nil, err
	}
	verified, err := readClaudeAuthFile(backup)
	if err != nil || verified.sha != saved.sha || !os.SameFile(verified.info, saved.info) {
		return nil, ErrClaudeAuthStale
	}
	result := &ClaudeAuthEditResult{Status: "not_restored", ConfigPath: path, SourceSHA256: current.sha, ResultSHA256: saved.sha, PolicySHA256: policy, BackupPath: backup, BackupSHA256: saved.sha}
	return publishClaudeAuth(ctx, current, saved, policy, saved.raw, result, ops, "restored")
}
