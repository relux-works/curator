package install

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/hookapproval"
	"github.com/relux-works/curator/internal/scopes"
	"github.com/relux-works/curator/internal/transaction"
)

// The actual install stager and profile publisher join env bytes and approvals
// in one journal through the default programmatic publisher (no injected
// profile callback). A fault after approval publication restores the old operator
// ledger and first-use profile absence. The healthy neighbor reaches the real
// maintenance collector, with only its process snapshot transport restricted
// to this fixture; this is not a native account or shell execution witness.
func TestProvisioningOriginGlobalApprovalPublication(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "healthy", true: "after-approval"}[fail], func(t *testing.T) {
			e := newEnv(t)
			e.skill("approval-skill")
			e.globalDeclare("approval-skill")
			user := t.TempDir()
			native := map[string]string{"codex_cli": filepath.Join(user, ".codex"), "claude_code": filepath.Join(user, ".claude"), "pi": filepath.Join(user, ".pi", "agent"), "opencode": filepath.Join(user, ".config", "opencode")}
			t.Setenv("HOME", user)
			t.Setenv("CODEX_HOME", native["codex_cli"])
			t.Setenv("CLAUDE_CONFIG_DIR", native["claude_code"])
			t.Setenv("PI_CODING_AGENT_DIR", native["pi"])
			t.Setenv("XDG_CONFIG_HOME", filepath.Join(user, ".config"))
			t.Setenv("XDG_DATA_HOME", filepath.Join(user, ".local", "share"))
			t.Setenv("XDG_STATE_HOME", filepath.Join(user, ".local", "state"))
			operator := filepath.Join(user, "operator-env.sh")
			if err := hookapproval.Upsert(e.home, hookapproval.Record{Path: operator, SHA256: strings.Repeat("a", 64), ApprovedBy: hookapproval.ApprovedByOperator, ApprovedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}); err != nil {
				t.Fatal(err)
			}
			prior, err := os.ReadFile(hookapproval.ApprovalsPath(e.home))
			if err != nil {
				t.Fatal(err)
			}
			hit, collected, joined := false, false, false
			planted := errors.New("paired-approval-fault")
			opts := Options{Commit: CommitDeps{Collect: func(req scopes.MaintenanceRequest) (scopes.MaintenanceResult, error) {
				collected = true
				store, err := buildcache.New(req.Home)
				if err != nil {
					return scopes.MaintenanceResult{}, err
				}
				store.ExecutablePaths = func() ([]string, error) { return nil, nil }
				req.Cache = store
				return scopes.Collect(req)
			}, Hooks: transaction.Hooks{Fault: func(event transaction.Event) error {
				if event.Point == transaction.PointPrepared {
					raw, err := os.ReadFile(filepath.Join(e.home, "state", "transactions", "v1", event.TransactionID+".json"))
					if err != nil {
						t.Fatal(err)
					}
					var journal transaction.Journal
					if err = json.Unmarshal(raw, &journal); err != nil {
						t.Fatal(err)
					}
					if journal.Unit != nil && journal.Unit.Legacy != nil {
						joined = len(journal.Unit.Profiles) > 0 && len(journal.Unit.Native) > 0
					}
				}
				if event.Point == transaction.PointAfterInstall && event.LivePath == hookapproval.ApprovalsPath(e.home) {
					hit = true
					if fail {
						return planted
					}
				}
				return nil
			}}}}

			result := Global(e.cfg, user, opts)
			if !hit || !joined {
				t.Fatalf("actual approval owner not reached: %+v", result)
			}
			if fail {
				if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), planted.Error()) || collected {
					t.Fatalf("wrong refusal or downstream continuation: %+v collected=%t", result, collected)
				}
				after, err := os.ReadFile(hookapproval.ApprovalsPath(e.home))
				if err != nil || string(after) != string(prior) {
					t.Fatalf("approval preimage changed: %v", err)
				}
				for _, path := range []string{filepath.Join(e.home, "profiles", "default", "source.json"), filepath.Join(e.home, "profiles", "default", "lock.json"), filepath.Join(e.home, "global", "env.sh"), filepath.Join(e.home, "global", "env.ps1")} {
					if _, err := os.Lstat(path); !os.IsNotExist(err) {
						t.Fatalf("paired refusal left mutable publication: %s %v", path, err)
					}
				}
			} else {
				if result.Status != "ok" || !collected {
					t.Fatalf("healthy actual continuation: %+v collected=%t", result, collected)
				}
				for _, name := range []string{"env.sh", "env.ps1"} {
					path := filepath.Join(e.home, "global", name)
					raw, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					record, found, err := hookapproval.Lookup(e.home, path)
					if err != nil || !found || record.SHA256 != hookapproval.Digest(raw) || record.ApprovedBy != hookapproval.ApprovedByManager {
						t.Fatalf("published approval differs: %+v %v", record, err)
					}
				}
				if record, found, err := hookapproval.Lookup(e.home, operator); err != nil || !found || record.ApprovedBy != hookapproval.ApprovedByOperator {
					t.Fatalf("unrelated approval lost: %+v %v", record, err)
				}
			}
		})
	}
}
