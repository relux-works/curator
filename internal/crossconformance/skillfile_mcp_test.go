package crossconformance

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/projectmcp"
	"github.com/relux-works/curator/internal/projectmcpstate"
)

const skillfileMCPPin = "dbea3ee0bb8716cb4fcef758609717c10557ceff"
const skillfileMCPData = "testdata/skillfile-mcp-v1"

type mcpVector struct {
	ID          string          `json:"id"`
	Claim       string          `json:"claim"`
	Kind        string          `json:"kind"`
	Raw         string          `json:"raw"`
	Schema      string          `json:"schema"`
	Instance    json.RawMessage `json:"instance"`
	Valid       bool            `json:"valid"`
	Declaration json.RawMessage `json:"declaration"`
	Normalized  json.RawMessage `json:"normalized"`
	Digest      string          `json:"digest"`
	Allowed     bool            `json:"allowed"`
	Facts       struct {
		Packages map[string]struct {
			Manifest json.RawMessage `json:"manifest"`
		} `json:"packages"`
	} `json:"facts"`
	Expected struct {
		Code *string `json:"code"`
	} `json:"expected"`
}

func loadMCPVectors(t *testing.T) []mcpVector {
	t.Helper()
	var cases []mcpVector
	if err := json.Unmarshal(readDraftFile(t, filepath.Join(skillfileMCPData, "corpus/conformance/skillfile-mcp-v1/cases.json")), &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 165 {
		t.Fatalf("corpus inventory changed: %d", len(cases))
	}
	return cases
}
func mcpJSON(t *testing.T, payload []byte) any {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		t.Fatal(err)
	}
	return value
}

// Every vendored byte belongs to the separate exact stage-A spec pin; released
// corpus directories and their pins are neither rewritten nor reinterpreted.
func TestSkillfileMCPPin(t *testing.T) {
	if got := strings.TrimSpace(string(readDraftFile(t, filepath.Join(skillfileMCPData, "SKILLFILE_MCP_PIN")))); got != skillfileMCPPin {
		t.Fatal(got)
	}
	expected := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(string(readDraftFile(t, filepath.Join(skillfileMCPData, "MANIFEST.sha256")))), "\n") {
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 {
			t.Fatal(line)
		}
		expected[parts[1]] = parts[0]
		sum := sha256.Sum256(readDraftFile(t, filepath.Join(skillfileMCPData, parts[1])))
		if hex.EncodeToString(sum[:]) != parts[0] {
			t.Fatalf("vendored bytes changed: %s", parts[1])
		}
	}
	err := filepath.WalkDir(filepath.Join(skillfileMCPData, "corpus"), func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(skillfileMCPData, path)
		if err != nil {
			return err
		}
		if _, ok := expected[filepath.ToSlash(relative)]; !ok {
			return fmt.Errorf("unmanifested corpus file %s", relative)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// This consumer slice exercises the published declaration/package parsers,
// normalization/digests, admission gate and all three passive carrier readers.
// Planning facts are not evidence of acquisition, audit, a native writer or
// protected action approval. Those effects have separate production fixtures.
func TestSkillfileMCPConsumerVectors(t *testing.T) {
	for _, vector := range loadMCPVectors(t) {
		if vector.Kind == "planning" && (mcpWriterVector(vector.ID) || (vector.Expected.Code != nil && *vector.Expected.Code == "upgrade_required")) {
			t.Logf("not consumed as a model verdict: %s (see consumer inventory)", vector.ID)
			continue
		}
		label := vector.ID
		if vector.Kind == "planning" {
			label = "parser-" + label
		}
		t.Run(label, func(t *testing.T) {
			t.Log(vector.Claim)
			switch vector.Kind {
			case "normalization":
				declaration, err := projectmcp.ParseDeclaration(mcpJSON(t, vector.Declaration))
				if err != nil {
					t.Fatal(err)
				}
				actual, err := projectmcpstate.Canonical(declaration.Model())
				if err != nil {
					t.Fatal(err)
				}
				want, err := projectmcpstate.Canonical(mcpJSON(t, vector.Normalized))
				if err != nil {
					t.Fatal(err)
				}
				digest, err := projectmcp.Digest(declaration.Model())
				if err != nil || !bytes.Equal(actual, want) || digest != vector.Digest {
					t.Fatalf("normalization %s %s: %v", actual, digest, err)
				}
			case "admission":
				declaration, err := projectmcp.ParseDeclaration(mcpJSON(t, vector.Declaration))
				if err != nil {
					t.Fatal(err)
				}
				if got := declaration.RequirePackage() == nil; got != vector.Allowed {
					t.Fatalf("admitted %t want %t", got, vector.Allowed)
				}
			case "carrier":
				payload, err := projectmcpstate.Canonical(mcpJSON(t, vector.Instance))
				if err != nil {
					t.Fatal(err)
				}
				switch vector.Schema {
				case "mcp-lock-v1":
					_, err = projectmcpstate.DecodeLock(payload)
				case "mcp-plan-v1":
					_, err = projectmcpstate.DecodePlan(payload)
				case "mcp-approval-v1":
					_, err = projectmcpstate.DecodeApproval(payload)
				default:
					t.Fatalf("unhandled carrier %s", vector.Schema)
				}
				if (err == nil) != vector.Valid {
					t.Fatalf("carrier admitted %t want %t: %v", err == nil, vector.Valid, err)
				}
			case "planning":
				parsed, err := manifest.ParseBytes([]byte(vector.Raw), "Skillfile.json")
				if err == nil {
					for _, name := range projectmcp.Names(parsed.MCPServers) {
						declaration := parsed.MCPServers[name]
						if declaration.Server != nil {
							err = declaration.RequirePackage()
							break
						}
						fixture := t.TempDir()
						fact, present := vector.Facts.Packages[name]
						if !present {
							// Absence is a replay precondition, not invented package bytes.
							continue
						}
						if len(fact.Manifest) > 0 && !bytes.Equal(fact.Manifest, []byte("null")) {
							if writeErr := os.WriteFile(filepath.Join(fixture, "agent-mcp.json"), fact.Manifest, 0600); writeErr != nil {
								t.Fatal(writeErr)
							}
						}
						_, err = projectmcp.LoadNamedPackage(fixture, name)
						if err != nil {
							break
						}
					}
				}
				expected := ""
				if vector.Expected.Code != nil {
					switch *vector.Expected.Code {
					case projectmcp.Invalid, projectmcp.InlineUnsupported, "context-secret-material":
						expected = *vector.Expected.Code
					}
				}
				if expected == "" {
					if err != nil {
						t.Fatal(err)
					}
				} else if err == nil || !strings.HasPrefix(err.Error(), expected+":") {
					t.Fatalf("want parser/admission %s: %v", expected, err)
				}
			default:
				t.Fatalf("unhandled vector kind %s", vector.Kind)
			}
		})
	}
}
func mcpWriterVector(id string) bool {
	return strings.HasPrefix(id, "action-") || id == "foreign-entry" || id == "unreadable-preimage"
}
