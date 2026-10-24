package envprofile

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/mcpnative"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/transaction"
)

// Production publishes private native files, ownership records, skills and
// locks on both adapters. Status accepts those files and remains read-only;
// a later foreign-read ownership grant is refused without being repaired.
func TestProfileMCPPrivatePublishedFiles(t *testing.T) {
	for _, variant := range []string{"private", "foreign-read"} {
		t.Run(variant, func(t *testing.T) {
			fx, req := nativeWriterFixture(t)
			result, err := ProfileMCP(req)
			if err != nil || !result.Published {
				t.Fatalf("private install: %+v %v", result, err)
			}
			for _, id := range req.Environments {
				adapter, _ := envregistry.ByID(id)
				for _, path := range []string{nativeWriterPath(fx, id), nativeMCPLedger(fx.home, fx.profile, id), lockPath(fx.home, fx.profile), filepath.Join(ManagedHomeDir(fx.home, fx.profile, id), adapter.SkillsDir, "myskill", "SKILL.md")} {
					if err := privatedir.ValidateFile(path); err != nil {
						t.Fatalf("production file is not private: %v", err)
					}
				}
			}
			ledger := nativeMCPLedger(fx.home, fx.profile, "codex_cli")
			if variant == "foreign-read" {
				restore, err := makeWorldReadableFileForTest(ledger)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := restore(); err != nil {
						t.Error(err)
					}
				})
				if privatedir.ValidateFile(ledger) == nil {
					t.Fatal("foreign-read plant not applied")
				}
			}
			before := snapshotMCPRecovery(t, fx.home)
			req.Mode = "status"
			result, err = ProfileMCP(req)
			if variant == "foreign-read" {
				requireMCPError(t, err, mcpnative.Unreadable)
				if privatedir.ValidateFile(ledger) == nil {
					t.Fatal("status repaired foreign ACL")
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				for _, row := range result.Rows {
					if row.State != "unchanged" {
						t.Fatalf("private status row: %+v", row)
					}
				}
			}
			if result.Published || !reflect.DeepEqual(before, snapshotMCPRecovery(t, fx.home)) {
				t.Fatal("status changed payload bytes")
			}
		})
	}
}

// Identical private skill bytes admit with the platform's actual writable
// file mode. Read-only attributes and foreign-read access each refuse preview
// and install; neither refusal changes existing skill bytes or publishes MCP.
func TestProfileMCPPrivateSkillPreimage(t *testing.T) {
	for _, variant := range []string{"private", "read-only", "foreign-read"} {
		t.Run(variant, func(t *testing.T) {
			fx, req := nativeWriterFixture(t)
			req.Environments = []string{"claude_code"}
			adapter, _ := envregistry.ByID("claude_code")
			path := filepath.Join(ManagedHomeDir(fx.home, fx.profile, "claude_code"), adapter.SkillsDir, "myskill", "SKILL.md")
			body := []byte("# myskill\n")
			nativeWrite(t, path, body)
			// The normal empty mutation-lock footprint is not a payload write.
			op, err := beginOperation(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			if err := op.close(); err != nil {
				t.Fatal(err)
			}
			code := mcpnative.AuthorityUnavailable
			switch variant {
			case "read-only":
				if err := os.Chmod(path, 0o400); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := os.Chmod(path, 0o600); err != nil {
						t.Error(err)
					}
				})
			case "foreign-read":
				restore, err := makeWorldReadableFileForTest(path)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := restore(); err != nil {
						t.Error(err)
					}
				})
				if privatedir.ValidateFile(path) == nil {
					t.Fatal("foreign-read plant not applied")
				}
			}
			for _, preview := range []bool{true, false} {
				before := snapshotMCPRecovery(t, fx.home)
				req.DryRun = preview
				result, err := ProfileMCP(req)
				if variant == "private" {
					if err != nil || result.Published == preview {
						t.Fatalf("identical private preimage: %+v %v", result, err)
					}
				} else {
					requireMCPError(t, err, code)
					if result.Published {
						t.Fatal("untrusted skill preimage published")
					}
					if !reflect.DeepEqual(before, snapshotMCPRecovery(t, fx.home)) {
						t.Fatal("preimage refusal changed bytes")
					}
					nativeAbsent(t, nativeMCPLedger(fx.home, fx.profile, "claude_code"))
					nativeAbsent(t, nativeWriterPath(fx, "claude_code"))
				}
				if !bytes.Equal(body, nativeRead(t, path)) {
					t.Fatal("existing skill bytes changed")
				}
			}
		})
	}
}

// A protected DACL at creation is insufficient if a later grant appears on the
// staged ownership file. The persisted private contract refuses that grant
// before publication and restores prior payloads; its unchanged sibling admits.
func TestProfileMCPPrivateStagingBoundary(t *testing.T) {
	for _, broaden := range []bool{false, true} {
		t.Run(map[bool]string{false: "private", true: "foreign-read"}[broaden], func(t *testing.T) {
			fx, req := nativeWriterFixture(t)
			beforeLock := nativeRead(t, lockPath(fx.home, fx.profile))
			ledger := nativeMCPLedger(fx.home, fx.profile, "codex_cli")
			observed := false
			req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{Observe: func(event transaction.Event) {
				if event.Point != transaction.PointPrepared {
					return
				}
				path := filepath.Join(fx.home, "state", "transactions", "v1", event.TransactionID+".json")
				var journal transaction.Journal
				if err := json.Unmarshal(nativeRead(t, path), &journal); err != nil {
					t.Fatal(err)
				}
				for _, target := range journal.Targets {
					if target.LivePath != ledger {
						continue
					}
					observed = true
					if !broaden {
						return
					}
					restore, err := makeWorldReadableFileForTest(target.StagedPath)
					if err != nil {
						t.Fatal(err)
					}
					if privatedir.ValidateFile(target.StagedPath) == nil {
						t.Fatal("staged foreign-read plant missed")
					}
					t.Cleanup(func() {
						if _, err := os.Lstat(target.StagedPath); err == nil {
							if err := restore(); err != nil {
								t.Error(err)
							}
						} else if !os.IsNotExist(err) {
							t.Error(err)
						}
					})
				}
			}})}
			result, err := ProfileMCP(req)
			if !observed {
				t.Fatal("ownership staging boundary not observed")
			}
			if broaden {
				if err == nil || !strings.Contains(err.Error(), "private transaction replacement") || result.Published {
					t.Fatalf("private staged grant accepted: %+v %v", result, err)
				}
				if !bytes.Equal(beforeLock, nativeRead(t, lockPath(fx.home, fx.profile))) {
					t.Fatal("staged refusal changed profile lock")
				}
				for _, id := range req.Environments {
					nativeAbsent(t, nativeMCPLedger(fx.home, fx.profile, id))
					nativeAbsent(t, nativeWriterPath(fx, id))
				}
			} else if err != nil || !result.Published {
				t.Fatalf("private staged positive: %+v %v", result, err)
			}
		})
	}
}
