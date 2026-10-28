package envauthority

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/protocoljson"
)

// LegacyMigration is the read-only, pre-whole-unit migration format. New
// operations never emit it. It lacks parent/protection/origin preimages, so an
// unsealed instance cannot authorize automatic rollback or forward publication.
type LegacyMigration struct {
	Version     int                     `json:"version"`
	Plan        string                  `json:"plan"`
	Ops         []LegacyMigrationOp     `json:"ops"`
	Markers     []LegacyMigrationMarker `json:"markers"`
	MarkersDone bool                    `json:"markers_done"`
	Complete    bool                    `json:"complete"`
}
type LegacyMigrationOp struct {
	Profile    string `json:"profile"`
	EnvID      string `json:"env"`
	Kind       string `json:"kind"`
	Path       string `json:"path"`
	From       string `json:"from"`
	To         string `json:"to"`
	Done       bool   `json:"done"`
	Temp       string `json:"temp,omitempty"`
	TempTarget string `json:"temp_target,omitempty"`
}
type LegacyMigrationMarker struct {
	Live  string `json:"live"`
	Prior []byte `json:"prior"`
	New   []byte `json:"new"`
}

var legacyMigrationHash = regexp.MustCompile(`^[0-9a-f]{64}$`)

// ReadCompletedLegacyMigration is common admission, not a selected-home or
// any-journal exemption. A sealed journal permits only removal of that private
// journal after its fixed owned postimages and independent recorded authority
// verify. It never supplies native authority or reads credential contents.
func ReadCompletedLegacyMigration(manager string) (*LegacyMigration, error) {
	raw, present, err := envorigin.ReadPrivateAuthority(manager, filepath.Join(manager, "state", "env-migration.json"), 1<<20)
	if err != nil {
		return nil, err
	}
	if !present {
		return nil, nil
	}
	refuse := func(detail string) (*LegacyMigration, error) {
		return nil, fmt.Errorf("%s: legacy migration journal %s; preserve the journal and owned entries for explicit recovery; no continuation is admitted", envorigin.DiagInvalid, detail)
	}
	var j LegacyMigration
	if err := protocoljson.DecodeClosed(raw, &j); err != nil {
		return refuse("is unreadable or malformed")
	}
	if j.Version != 1 || !legacyMigrationHash.MatchString(j.Plan) {
		return refuse("has an invalid identity")
	}
	if !j.Complete {
		return refuse("lacks complete durable whole-unit metadata and origin preimages")
	}
	if len(j.Markers) > 0 && !j.MarkersDone {
		return refuse("has an incomplete marker phase")
	}
	seen := map[string]bool{}
	home := func(profile, adapter string) (string, *envmarker.Marker, error) {
		if !identifiers.Valid(profile) {
			return "", nil, fmt.Errorf("invalid profile")
		}
		switch adapter {
		case "codex_cli", "pi", "claude_code", "muse":
		default:
			return "", nil, fmt.Errorf("invalid credential adapter")
		}
		target := filepath.Join(manager, "environments", profile, adapter)
		raw, present, err := envorigin.ReadPrivateAuthority(manager, filepath.Join(target, envmarker.Name), 1<<20)
		if err != nil || !present {
			return "", nil, fmt.Errorf("missing or invalid marker")
		}
		marker, err := envmarker.Parse(raw)
		if err != nil || marker == nil {
			return "", nil, fmt.Errorf("missing or invalid marker")
		}
		if err := RequirePerHome(marker); err != nil {
			return "", nil, err
		}
		if err := validateRecorded(manager, profile, adapter, target, marker, nil); err != nil {
			return "", nil, err
		}
		_, source, err := envorigin.ReadOwnedRecord(manager, filepath.Join(manager, "profiles", profile, "source.json"), 16<<20)
		if err != nil || !source {
			return "", nil, fmt.Errorf("installed source missing or unreadable")
		}
		return target, marker, nil
	}
	for _, op := range j.Ops {
		target, marker, err := home(op.Profile, op.EnvID)
		if err != nil {
			return refuse("has no independently valid home binding")
		}
		auth := "auth.json"
		if op.EnvID == "claude_code" {
			auth = ".credentials.json"
		}
		if op.EnvID == "muse" {
			auth = "config/muse/auth.json"
		}
		if op.Path != auth || !op.Done || op.Temp != "" || op.TempTarget != "" {
			return refuse("has an unknown entry or incomplete phase")
		}
		path := filepath.Join(target, filepath.FromSlash(auth))
		if err := pathboundary.ValidateRouteWithOwner(manager, filepath.Dir(path), pathboundary.DefaultOwnerLookup()); err != nil {
			return refuse("credential entry boundary differs")
		}
		if seen[path] {
			return refuse("has duplicate entries")
		}
		seen[path] = true
		if op.Kind == "unlink" {
			if op.To != "" {
				return refuse("has an invalid unlink")
			}
			if _, err := os.Lstat(path); !os.IsNotExist(err) {
				return refuse("unlink postimage differs")
			}
			if marker.Passthrough != nil {
				for _, p := range *marker.Passthrough {
					if p.Path == auth {
						return refuse("unlink remains recorded as active")
					}
				}
			}
		} else if op.Kind == "relink" {
			originRaw, _, err := envorigin.ReadPrivateAuthority(manager, filepath.Join(target, envorigin.Name), envorigin.MaxRecordBytes)
			if err != nil {
				return refuse("origin cannot be verified")
			}
			record, err := envorigin.Parse(originRaw, envorigin.Binding{Profile: op.Profile, LockSHA256: "sha256:" + strings.TrimPrefix(marker.Profile.LockSHA256, "sha256:"), ManagedTarget: target}, op.EnvID, manager)
			if err != nil {
				return refuse("origin binding differs")
			}
			leaf := filepath.Base(filepath.FromSlash(auth))
			if op.To != filepath.Join(record.NativeRoot, leaf) {
				return refuse("intended target differs from recorded origin")
			}
			got, err := os.Readlink(path)
			if err != nil || got != op.To {
				return refuse("link postimage differs")
			}
		} else {
			return refuse("has an unknown operation")
		}
	}
	for _, m := range j.Markers {
		rel, err := filepath.Rel(filepath.Join(manager, "environments"), m.Live)
		parts := strings.Split(rel, string(filepath.Separator))
		if err != nil || len(parts) != 3 || parts[2] != envmarker.Name || m.Live != filepath.Join(manager, "environments", parts[0], parts[1], envmarker.Name) || seen[m.Live] {
			return refuse("has an unknown or duplicate marker")
		}
		seen[m.Live] = true
		target, _, err := home(parts[0], parts[1])
		if err != nil {
			return refuse("marker home binding differs")
		}
		got, present, err := envorigin.ReadPrivateAuthority(manager, filepath.Join(target, envmarker.Name), 1<<20)
		if err != nil || !present || !bytes.Equal(got, m.New) {
			return refuse("marker postimage differs")
		}
		if _, err := envmarker.Parse(m.Prior); err != nil {
			return refuse("prior marker is invalid")
		}
	}
	return &j, nil
}
